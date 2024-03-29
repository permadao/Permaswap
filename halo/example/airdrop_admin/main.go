package proposal

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/logger"
)

var log = logger.New("proposal")

func InSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var (
	ErrProposalInvalidInitData = errors.New("err_proposal_invalid_init_data")
	ErrProposalInvalidAmount   = errors.New("err_proposal_invalid_amount")
)

type Pay struct {
	To        string `json:"to"`
	Amount    string `json:"amount"`
	StakePool string `json:"stakePool"`
}

type InitData struct {
	From string `json:"from"`
	Pays []Pay  `json:"pays"`
}

func Execute(tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle, localState, initData string) (*schema.StateForProposal, string, string, error) {
	var toPay InitData

	if err := json.Unmarshal([]byte(initData), &toPay); err != nil {
		return state, localState, "", ErrProposalInvalidInitData
	}

	// add new admin stake pool
	if !InSlice(state.StakePools, "admin") {
		state.StakePools = append(state.StakePools, "admin")
	}
	if !InSlice(state.OnlyUnStakePools, "admin") {
		state.OnlyUnStakePools = append(state.OnlyUnStakePools, "admin")
	}

	from := toPay.From
	now, _ := strconv.ParseInt(tx.Nonce, 10, 64)
	now = now / 1000
	for _, pay := range toPay.Pays {
		amount, ok := new(big.Int).SetString(pay.Amount, 10)
		if !ok {
			return state, localState, "", ErrProposalInvalidAmount
		}
		feeRecipient := state.FeeRecipient
		dryRun := false

		//todo: check if pay.to is valid address
		err := state.Token.TransferToStake(from, pay.To, amount, pay.StakePool, now, feeRecipient, big.NewInt(0), dryRun)
		if err != nil {
			return state, localState, "", err
		}
	}
	return state, localState, "", nil
}
