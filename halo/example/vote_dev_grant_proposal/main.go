package proposal

import (
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/permadao/permaswap/halo/hvm/schema"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
)

var (
	ErrPropsalInvalidInitData     = errors.New("err_proposal_invalid_init_data")
	ErrPropsalInvalidTxAction     = errors.New("err_proposal_invalid_tx_action")
	ErrPropsalInvalidTxCallParams = errors.New("err_proposal_invalid_tx_call_params")
	ErrPropsalInvalidVoteTime     = errors.New("err_proposal_invalid_vote_time")
	ErrPropsalInvalidVoteParams   = errors.New("err_proposal_invalid_vote_params")
	ErrPropsalNoStakes            = errors.New("err_proposal_no_stakes")
	ErrPropsalInvalidExectueTime  = errors.New("err_proposal_invalid_execute_time")
	ErrPropsalExectued            = errors.New("err_proposal_executed")
	ErrPropsalInvalidGrantAmount  = errors.New("err_proposal_invalid_grant_amount")
	ErrPropsalInvalidFunction     = errors.New("err_proposal_invalid_function")
)

type InitData struct {
	VoteDeadLine int64  `json:"voteDeadline"`
	GrantFrom    string `json:"grantFrom"`
	GrantTo      string `json:"grantTo"`
	GrantAmount  string `json:"grantAmount"`
}

type Voted struct {
	Infavor bool  `json:"infavor"`
	Weight  int64 `json:"weight"`
}

type LocalState struct {
	Voteds   map[string]Voted `json:"voteds"`
	Infavor  int64            `json:"infavor"`
	Against  int64            `json:"against"`
	Approved bool             `json:"approved"`
	Executed bool             `json:"executed"`
}

func (l *LocalState) String() string {
	keys := make([]string, 0, len(l.Voteds))
	for k := range l.Voteds {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vote := "vote:"
	for _, k := range keys {
		vote += k + ":" + strconv.FormatBool(l.Voteds[k].Infavor) + ":" + strconv.FormatInt(l.Voteds[k].Weight, 10) + ";"
	}

	return "voteds:" + vote + "\n" +
		"infavor:" + strconv.FormatInt(l.Infavor, 10) + "\n" +
		"against:" + strconv.FormatInt(l.Against, 10) + "\n" +
		"approved:" + strconv.FormatBool(l.Approved) + "\n" +
		"executed:" + strconv.FormatBool(l.Executed) + "\n"
}

func (l *LocalState) Hash() string {
	return hexutil.Encode(accounts.TextHash([]byte(l.String())))
}

// function vote's params
const (
	FunctionVote    = "Vote"
	FunctionExecute = "Execute"
)

type VoteParams struct {
	Infavor bool `json:"infavor"`
}

// todo: improve voteWeight
func voteWeight(stakes map[string][]tokSchema.Stake) int64 {
	staked := big.NewInt(0)

	for _, stakePools := range stakes {
		for _, stake := range stakePools {
			staked = new(big.Int).Add(staked, stake.Amount)
		}
	}
	//staked = new(big.Int).Div(staked, big.NewInt(1e10))
	return staked.Int64()
}

func Execute(tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle, localState, initData string) (*schema.StateForProposal, string, string, error) {
	var init InitData
	if initData == "" {
		return state, localState, "", ErrPropsalInvalidInitData
	}
	if err := json.Unmarshal([]byte(initData), &init); err != nil {
		return state, localState, "", ErrPropsalInvalidInitData
	}

	var local LocalState
	if localState != "" {
		if err := json.Unmarshal([]byte(localState), &local); err != nil {
			return state, localState, "", err
		}
	}

	if local.Voteds == nil {
		local.Voteds = make(map[string]Voted)
	}

	if tx.Action != schema.TxActionCall {
		return state, localState, "", ErrPropsalInvalidTxAction
	}

	var txCallParams schema.TxCallParams
	if err := json.Unmarshal([]byte(tx.Params), &txCallParams); err != nil {
		return state, localState, "", ErrPropsalInvalidTxCallParams
	}

	now, _ := strconv.ParseInt(tx.Nonce, 10, 64)
	now = now / 1000

	switch txCallParams.Function {
	case FunctionVote:

		if now > init.VoteDeadLine {
			return state, localState, "", ErrPropsalInvalidVoteTime
		}

		var voteParams VoteParams
		if err := json.Unmarshal([]byte(txCallParams.Params), &voteParams); err != nil {
			return state, localState, "", ErrPropsalInvalidVoteParams
		}
		stakes, ok := state.Token.Stakes[tx.From]
		if !ok {
			return state, localState, "", ErrPropsalNoStakes
		}

		voteWeight := voteWeight(stakes)
		if preVoted, ok := local.Voteds[tx.From]; ok {
			if preVoted.Infavor {
				local.Infavor -= preVoted.Weight
			} else {
				local.Against -= preVoted.Weight
			}
		}
		local.Voteds[tx.From] = Voted{
			Infavor: voteParams.Infavor,
			Weight:  voteWeight,
		}

		if voteParams.Infavor {
			local.Infavor += voteWeight
		} else {
			local.Against += voteWeight
		}

		localStateNew, err := json.Marshal(local)
		if err != nil {
			return state, localState, "", err
		} else {
			localStateHashNew := local.Hash()
			return state, string(localStateNew), localStateHashNew, nil
		}
	case FunctionExecute:
		if now <= init.VoteDeadLine {
			return state, localState, "", ErrPropsalInvalidExectueTime
		}
		if local.Executed {
			return state, localState, "", ErrPropsalExectued
		}

		if local.Infavor <= local.Against {
			local.Approved = false
			localStateNew, err := json.Marshal(local)
			if err != nil {
				return state, localState, "", err
			}
			localStateHashNew := local.Hash()
			return state, string(localStateNew), localStateHashNew, nil
		}

		local.Approved = true
		local.Executed = true
		localStateNew, err := json.Marshal(local)
		if err != nil {
			return state, localState, "", err
		}
		localStateHashNew := local.Hash()

		amount, ok := new(big.Int).SetString(init.GrantAmount, 10)
		if !ok {
			return state, localState, "", ErrPropsalInvalidGrantAmount
		}
		// todo check grantTo is a valid address
		err = state.Token.Transfer(init.GrantFrom, init.GrantTo, amount, state.FeeRecipient, big.NewInt(0), false)
		return state, string(localStateNew), localStateHashNew, err
	default:
		return state, localState, "", ErrPropsalInvalidFunction
	}
}
