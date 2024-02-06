package hvm

import (
	"math/big"
	"testing"

	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/token"
	"github.com/stretchr/testify/assert"
)

const source = `package proposal

import (
	"math/big"

	"github.com/permadao/permaswap/halo/hvm/schema"
)

func Execute(tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle, localState string, initData string) (*schema.StateForProposal, string, string, error) {
	amount, _ := new(big.Int).SetString("10000000000000000000", 10)
	to := "0x7759cb78EaF06c470165F0B57af7Ffd737407D56"
	dryRun := false
	err := state.Token.Transfer("ecosystem", to, amount, state.FeeRecipient, big.NewInt(0), dryRun)
	localState = "1234"
	return state, localState, "", err
}
`

func TestProposal(t *testing.T) {
	executor, err := NewExecutor(source)
	assert.NoError(t, err)
	proposal := NewProposal("test", 0, 0, 1, source, "", nil, executor)
	amount, _ := new(big.Int).SetString("100000000000000000000", 10)
	token := token.Token{
		Balances: map[string]*big.Int{
			"ecosystem": amount,
			"0x7759cb78EaF06c470165F0B57af7Ffd737407D56": amount,
			"0x36da5367c7fC6f446ec9faC87Af73581cD3ADAe7": amount,
		},
	}
	state := schema.State{
		FeeRecipient: "0x36da5367c7fC6f446ec9faC87Af73581cD3ADAe7",
		Token:        &token,
	}
	t.Log("Original state:", state.Token.Balances)

	tx := schema.Transaction{}
	state2, err := ProposalExecute(proposal, &tx, state.GetStateForProposal(), nil)
	assert.NoError(t, err)

	t.Log("Original state after execute:", state.Token.Balances)
	t.Log("New state:", state2.Token.Balances)
}
