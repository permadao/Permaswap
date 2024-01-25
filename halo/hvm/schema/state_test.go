package schema

import (
	"math/big"
	"testing"

	"github.com/permadao/permaswap/halo/token"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
)

func TestCopyToken(t *testing.T) {
	balances := map[string]*big.Int{
		"0x1": big.NewInt(1),
		"0x2": big.NewInt(2),
	}
	stakes := map[string]map[string][]tokSchema.Stake{
		"0x1": {"basic": {{StakedAt: 1, Amount: big.NewInt(1)}, {StakedAt: 2, Amount: big.NewInt(2)}}},
		"0x2": {"dev": {{StakedAt: 3, Amount: big.NewInt(3)}, {StakedAt: 4, Amount: big.NewInt(4)}}},
	}
	halo := token.New("HALO", 18, big.NewInt(1000000000000000000), balances, stakes)
	t.Log("halo:", halo)
	t.Log("halo.stakes:", halo.Stakes)

	tokenCopied := &token.Token{}
	CopyToken(tokenCopied, halo)
	t.Log("tokenCopied:", tokenCopied)
	t.Log("stakes:", tokenCopied.Stakes)
}
