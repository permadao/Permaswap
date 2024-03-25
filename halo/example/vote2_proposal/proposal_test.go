package proposal

import (
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/token"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
)

func testGenToken() *token.Token {
	totalSupply, _ := new(big.Int).SetString("1000000000000000000000000000", 10)
	stakes := map[string]map[string][]tokSchema.Stake{
		"0x1": {
			"basic": {
				tokSchema.Stake{
					StakedAt: 1,
					Amount:   big.NewInt(1),
				},
				tokSchema.Stake{
					StakedAt: 2,
					Amount:   big.NewInt(2),
				},
			},
			"dev": {
				tokSchema.Stake{
					StakedAt: 3,
					Amount:   big.NewInt(3),
				},
				tokSchema.Stake{
					StakedAt: 4,
					Amount:   big.NewInt(4),
				},
			},
		},
		"0x2": {
			"basic": {
				tokSchema.Stake{
					StakedAt: 1,
					Amount:   big.NewInt(4),
				},
			},
			"dev": {
				tokSchema.Stake{
					StakedAt: 3,
					Amount:   big.NewInt(5),
				},
				tokSchema.Stake{
					StakedAt: 4,
					Amount:   big.NewInt(6),
				},
			},
		},
		"0x3": {
			"basic": {
				tokSchema.Stake{
					StakedAt: 1,
					Amount:   big.NewInt(10),
				},
			},
		},
	}
	return &token.Token{
		Symbol:      "halo",
		Decimals:    18,
		TotalSupply: totalSupply,
		Balances:    map[string]*big.Int{},
		Stakes:      stakes,
	}
}

func testGenState() *schema.StateForProposal {
	t := testGenToken()
	return &schema.StateForProposal{
		Dapp:           "halo_test",
		ChainID:        "5",
		Govern:         "",
		FeeRecipient:   "",
		RouterMinStake: "80000000000000000000000",
		Routers:        []string{},
		RouterStates:   map[string]*schema.RouterState{},
		Token:          t,
		StakePools:     []string{"basic", "dev"},
	}
}

func testGenTx() *schema.Transaction {
	nonce := time.Now().UnixNano() / 1000000
	return &schema.Transaction{
		From:   "0x1",
		Action: "call",
		Nonce:  strconv.FormatInt(nonce, 10),
		Params: `{"function": "Vote", "params": "{\"infavor\": false}"}`,
	}
}

func testGenInitData() string {
	return `{"stakePool": "dev", 
			"voteStartAt": 1711357200, 
			"threshold": "10000000000000000000", 
			"minVoteDuration": 900, 
			"confirmDuration": 900}`
}

func TestExecute(t *testing.T) {
	tx := testGenTx()
	state := testGenState()
	initData := testGenInitData()
	stateNew, localState, _, err := Execute(tx, state, nil, "", initData)
	t.Log(stateNew)
	t.Log(localState)
	t.Log(err)
}
