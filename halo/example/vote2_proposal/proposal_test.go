package proposal

import (
	"math/big"
	"testing"

	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/token"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
	"github.com/stretchr/testify/assert"
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
					Amount:   big.NewInt(50),
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

func testGenInitData() string {
	return `{"stakePool": "dev", 
			"voteStartAt": 1711357200, 
			"threshold": "100", 
			"minVoteDuration": 900, 
			"confirmDuration": 900}`
}

func TestExecute(t *testing.T) {
	state1 := testGenState()
	initData := testGenInitData()

	tx1 := &schema.Transaction{
		//nonce := time.Now().UnixNano() / 1000000
		From:   "0x1",
		Action: "call",
		Nonce:  "1711357210000",
		Params: `{"function": "Vote", "params": "{\"infavor\": false}"}`,
	}
	state2, localState2, _, err := Execute(tx1, state1, nil, "", initData)
	t.Log("state2", state2)
	t.Log("localState2", localState2, "\n")
	assert.NoError(t, err)

	tx2 := &schema.Transaction{
		From:   "0x2",
		Action: "call",
		Nonce:  "1711367310000",
		Params: `{"function": "Vote", "params": "{\"infavor\": true}"}`,
	}
	state3, localState3, _, err := Execute(tx2, state2, nil, localState2, initData)
	t.Log("state3", state3)
	t.Log("localState3", localState3, "\n")
	assert.NoError(t, err)

	tx3 := &schema.Transaction{
		From:   "0x3",
		Action: "call",
		Nonce:  "1711358200000",
		Params: `{"function": "Vote", "params": "{\"infavor\": true}"}`,
	}
	state4, localState4, _, err := Execute(tx3, state3, nil, localState3, initData)
	t.Log("state4", state4)
	t.Log("localState4", localState4, "\n")
	assert.NoError(t, err)

	tx4 := &schema.Transaction{
		From:   "0x3",
		Action: "call",
		Nonce:  "1711359300000",
		Params: `{"function": "Vote", "params": "{\"infavor\": true}"}`,
	}
	state5, localState5, _, err := Execute(tx4, state4, nil, localState4, initData)
	t.Log("state5", state5)
	t.Log("localState5", localState5)
	t.Log("err", err, "\n")

	tx5 := &schema.Transaction{
		From:   "0x3",
		Action: "call",
		Nonce:  "1711359300000",
		Params: `{"function": "Execute"}`,
	}
	state6, localState6, _, err := Execute(tx5, state5, nil, localState5, initData)
	t.Log("state6", state6)
	t.Log("localState6", localState6, "\n")
	t.Log("err", err)

}
