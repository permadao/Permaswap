package token

import (
	"math/big"
	"testing"
	"time"

	"github.com/permadao/permaswap/halo/token/schema"
	"github.com/stretchr/testify/assert"
)

func TestStake(t *testing.T) {
}

func TestUnstakeAlgo(t *testing.T) {
	stakes := []*big.Int{big.NewInt(10), big.NewInt(2), big.NewInt(5)}
	amount := big.NewInt(6)
	for i := len(stakes) - 1; i >= 0; i-- {
		stake := stakes[i]
		if stake.Cmp(amount) == -1 {
			amount = new(big.Int).Sub(amount, stake)
			stakes = stakes[:i]
		} else {
			stakes[i] = new(big.Int).Sub(stake, amount)
			break
		}
	}
	t.Log("stakes:", stakes)
	assert.Equal(t, len(stakes), 2)
}

func TestTotalStaked(t *testing.T) {
	stakes := map[string]map[string][]schema.Stake{
		"0x1": {
			"basic": {
				schema.Stake{
					StakedAt: 1,
					Amount:   big.NewInt(1),
				},
				schema.Stake{
					StakedAt: 2,
					Amount:   big.NewInt(2),
				},
			},
			"dev": {
				schema.Stake{
					StakedAt: 3,
					Amount:   big.NewInt(3),
				},
				schema.Stake{
					StakedAt: 4,
					Amount:   big.NewInt(4),
				},
			},
		},
	}
	testToken := New("test", 18, big.NewInt(1000), nil, stakes)
	totalStaked := testToken.TotalStaked("0x1", "")
	assert.Equal(t, totalStaked, "10")

	totalStaked = testToken.TotalStaked("0x1", "basic")
	assert.Equal(t, totalStaked, "3")

	totalStaked = testToken.TotalStaked("0x1", "dev")
	assert.Equal(t, totalStaked, "7")
}

func TestTransferToStake(t *testing.T) {
	balances := map[string]*big.Int{
		"eco": big.NewInt(100),
	}
	testToken := New("test", 18, big.NewInt(1000), balances, nil)
	testToken.TransferToStake("eco", "0x61EbF673c200646236B2c53465bcA0699455d5FA",
		big.NewInt(10), "basic", time.Now().UnixNano()/1000000, "0x61EbF673c200646236B2c53465bcA0699455d5FA", big.NewInt(0), false)
	t.Log("balances:", testToken.Balances)
	t.Log("stakes:", testToken.Stakes)
}
