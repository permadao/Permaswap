package proposal

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/permadao/permaswap/halo/hvm/schema"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
)

var (
	ErrPropsalInvalidInitData     = errors.New("err_proposal_invalid_init_data")
	ErrPropsalInvalidTxAction     = errors.New("err_proposal_invalid_tx_action")
	ErrPropsalInvalidTxCallParams = errors.New("err_proposal_invalid_tx_call_params")
	ErrPropsalInvalidVoteTime     = errors.New("err_proposal_invalid_vote_time")
	ErrPropsalInvalidVoteParams   = errors.New("err_proposal_invalid_vote_params")
	ErrPropsalUserNoStakes        = errors.New("err_proposal_user_no_stakes")
	ErrPropsalInvalidExectueTime  = errors.New("err_proposal_invalid_execute_time")
	ErrPropsalExectued            = errors.New("err_proposal_executed")
	ErrPropsalInvalidGrantAmount  = errors.New("err_proposal_invalid_grant_amount")
	ErrPropsalInvalidFunction     = errors.New("err_proposal_invalid_function")
	ErrPropsalVoteEnd             = errors.New("err_proposal_vote_end")
)

const (
	THRESHOLD       = "300000000000000000000000"
	MAJORITY        = "0.6666"
	MINVOTEDURATION = 3 * 24 * 60 * 60
	MAXVOTEDURATION = 7 * 24 * 60 * 60
	CONFIRMDURATION = 24 * 60 * 60
	VOTEWEIGHT      = 10
)

type InitData struct {
	VoteStartAt int64  `json:"voteStartAt"`
	StakePool   string `json:"stakePool"`

	//a vote prams
	Threshold       string `json:"threshold"`
	Majority        string `json:"majority"`
	MinVoteDuration int64  `json:"minVoteDuration"`
	MaxVoteDuration int64  `json:"maxVoteDuration"`
	ConfirmDuration int64  `json:"confirmDuration"`
	VoteWeight      int64  `json:"voteWeight"`
}

var (
	PeriodVote    = "vote"
	PeriodConfirm = "confirm"
	PeriodExecute = "execute"
	PeriodEnd     = "end"
)

type Voted struct {
	Infavor bool     `json:"infavor"`
	Amount  *big.Int `json:"amount"`
	Weight  *big.Int `json:"weight"`
}

// local status
type Voting struct {
	Period                string `json:"period"`
	CurrentVoteStartAt    int64  `json:"currentVoteStartAt"`
	CurrentConfirmStartAt int64  `json:"currentConfirmStartAt"`

	Voted   map[string]Voted `json:"voteds"` // address -> voted
	Infavor *big.Int         `json:"infavor"`
	Against *big.Int         `json:"against"`

	Approved bool `json:"approved"`
	Executed bool `json:"executed"`

	//a vote prams
	Threshold       string `json:"threshold"`
	Majority        string `json:"majority"`
	MinVoteDuration int64  `json:"minVoteDuration"`
	MaxVoteDuration int64  `json:"maxVoteDuration"`
	ConfirmDuration int64  `json:"confirmDuration"`
	VoteWeight      int64  `json:"voteWeight"`
}

func NewVoting(init InitData) Voting {
	voting := Voting{
		Period:                PeriodVote,
		CurrentVoteStartAt:    init.VoteStartAt,
		CurrentConfirmStartAt: 0,
		Voted:                 make(map[string]Voted),
		Infavor:               big.NewInt(0),
		Against:               big.NewInt(0),
		Executed:              false,
		Approved:              false,

		Threshold:       THRESHOLD,
		Majority:        MAJORITY,
		MinVoteDuration: MINVOTEDURATION,
		MaxVoteDuration: MAXVOTEDURATION,
		ConfirmDuration: CONFIRMDURATION,
		VoteWeight:      VOTEWEIGHT,
	}

	if init.Threshold != "" {
		voting.Threshold = init.Threshold
	}
	if init.Majority != "" {
		voting.Majority = init.Majority
	}
	if init.MinVoteDuration != 0 {
		voting.MinVoteDuration = init.MinVoteDuration
	}
	if init.MaxVoteDuration != 0 {
		voting.MaxVoteDuration = init.MaxVoteDuration
	}
	if init.ConfirmDuration != 0 {
		voting.ConfirmDuration = init.ConfirmDuration
	}
	if init.VoteWeight != 0 {
		voting.VoteWeight = init.VoteWeight
	}

	return voting
}

func (v *Voting) TotalVoted() (amount, weight *big.Int) {
	for _, voted := range v.Voted {
		amount = new(big.Int).Add(amount, voted.Amount)
		weight = new(big.Int).Add(weight, voted.Weight)
	}
	return amount, weight
}

func (v *Voting) Infavored() bool {
	_, total := v.TotalVoted()
	totalf := new(big.Float).SetInt(total)
	m, _ := new(big.Float).SetString(v.Majority)
	infavorf := new(big.Float).SetInt(v.Infavor)
	return infavorf.Cmp(totalf.Mul(totalf, m)) == 1
}

// vote <-> confirm -> execute -> end
// vote -> end
func (v *Voting) UpdatePeriod(now int64) string {
	total, _ := v.TotalVoted()
	if v.Period == PeriodVote {
		threshold, _ := new(big.Int).SetString(v.Threshold, 10)
		if total.Cmp(threshold) == 1 && v.Infavored() && now > v.CurrentVoteStartAt+v.MinVoteDuration {
			v.CurrentConfirmStartAt = now
			v.Period = PeriodConfirm
		}
		// proposal failed
		if now > v.CurrentVoteStartAt+v.MaxVoteDuration {
			v.Period = PeriodEnd
			v.Approved = false
			v.Executed = false
		}
		return v.Period
	}

	if v.Period == PeriodConfirm {
		if !v.Infavored() {
			v.Period = PeriodVote
			v.CurrentVoteStartAt = now
		}
		if v.Infavored() && now > v.CurrentConfirmStartAt+v.ConfirmDuration {
			v.Period = PeriodExecute
			v.Approved = true
		}
		return v.Period
	}

	if v.Period == PeriodExecute && v.Executed {
		v.Period = PeriodEnd
		return v.Period
	}

	return v.Period
}

// function vote's params
const (
	FunctionVote    = "Vote"
	FunctionExecute = "Execute"
)

type VoteParams struct {
	Infavor bool `json:"infavor"`
}

func voteWeight(stakePool string, voteWeight int64, stakes map[string][]tokSchema.Stake) (amount, weight *big.Int) {
	staked := big.NewInt(0)
	w := big.NewInt(0)
	for pool, stakePools := range stakes {
		for _, stake := range stakePools {
			if pool == stakePool {
				nw := new(big.Int).Mul(stake.Amount, big.NewInt(voteWeight))
				w = new(big.Int).Add(w, nw)
			} else {
				w = new(big.Int).Add(w, stake.Amount)
			}
			staked = new(big.Int).Add(staked, stake.Amount)
		}
	}
	return staked, w
}

// Specific proposal details
func _execute(tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle, localState, initData string) (*schema.StateForProposal, string, string, error) {
	return state, localState, "", nil
}

func Execute(tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle, localState, initData string) (*schema.StateForProposal, string, string, error) {

	if tx.Action != schema.TxActionCall {
		return state, localState, "", ErrPropsalInvalidTxAction
	}

	var init InitData
	if initData == "" {
		return state, localState, "", ErrPropsalInvalidInitData
	}
	if err := json.Unmarshal([]byte(initData), &init); err != nil {
		return state, localState, "", ErrPropsalInvalidInitData
	}

	now, _ := strconv.ParseInt(tx.Nonce, 10, 64)
	now = now / 1000
	if now < init.VoteStartAt {
		return state, localState, "", ErrPropsalInvalidVoteTime
	}

	var voting Voting
	if localState == "" {
		voting = NewVoting(init)
	} else {
		if err := json.Unmarshal([]byte(localState), &voting); err != nil {
			return state, localState, "", err
		}
	}

	// Update period first
	voting.UpdatePeriod(now)
	localStateNew, err := json.Marshal(voting)
	if err != nil {
		return state, localState, "", err
	}
	localState = string(localStateNew)

	if voting.Period == PeriodEnd {
		return state, localState, "", ErrPropsalVoteEnd
	}

	var txCallParams schema.TxCallParams
	if err := json.Unmarshal([]byte(tx.Params), &txCallParams); err != nil {
		return state, localState, "", ErrPropsalInvalidTxCallParams
	}

	switch txCallParams.Function {
	case FunctionVote:

		if voting.Period == PeriodExecute {
			return state, localState, "", ErrPropsalInvalidVoteTime
		}

		var voteParams VoteParams
		if err := json.Unmarshal([]byte(txCallParams.Params), &voteParams); err != nil {
			return state, localState, "", ErrPropsalInvalidVoteParams
		}
		stakes, ok := state.Token.Stakes[tx.From]
		if !ok {
			return state, localState, "", ErrPropsalUserNoStakes
		}

		voteAmount, voteWeight := voteWeight(init.StakePool, voting.VoteWeight, stakes)
		if preVoted, ok := voting.Voted[tx.From]; ok {
			if preVoted.Infavor {
				voting.Infavor = new(big.Int).Sub(voting.Infavor, preVoted.Weight)
			} else {
				voting.Against = new(big.Int).Sub(voting.Against, preVoted.Weight)
			}
		}
		voting.Voted[tx.From] = Voted{
			Infavor: voteParams.Infavor,
			Amount:  voteAmount,
			Weight:  voteWeight,
		}
		if voteParams.Infavor {
			voting.Infavor = new(big.Int).Add(voting.Infavor, voteWeight)
		} else {
			voting.Against = new(big.Int).Add(voting.Against, voteWeight)
		}

		localStateNew, err := json.Marshal(voting)
		if err != nil {
			return state, localState, "", err
		} else {
			return state, string(localStateNew), "", nil
		}
	case FunctionExecute:
		if voting.Period != PeriodExecute {
			return state, localState, "", ErrPropsalInvalidExectueTime
		}

		if voting.Executed {
			return state, localState, "", ErrPropsalExectued
		}

		voting.Executed = true
		localStateNew, err := json.Marshal(voting)
		if err != nil {
			return state, localState, "", err
		}

		// execute
		state, localStateNew2, _, err := _execute(tx, state, oracle, string(localStateNew), initData)
		return state, localStateNew2, "", err

	default:
		return state, localState, "", ErrPropsalInvalidFunction
	}
}
