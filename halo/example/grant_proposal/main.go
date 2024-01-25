package proposal

import (
	"math/big"

	"github.com/permadao/permaswap/halo/hvm/schema"
)

func Execute(tx *schema.Transaction, state *schema.StateForProposal, localState string) (*schema.StateForProposal, string, string, error) {
	amount, _ := new(big.Int).SetString("1000000000000000000000000", 10)
	feeRecipient := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	to := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	dryRun := false
	err := state.Token.Transfer("ecosystem", to, amount, feeRecipient, big.NewInt(0), dryRun)
	return state, localState, "", err
}
