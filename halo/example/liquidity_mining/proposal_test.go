package proposal

import (
	"math/big"
	"testing"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/token"
	"github.com/stretchr/testify/assert"
)

func testGenToken() *token.Token {
	totalSupply, _ := new(big.Int).SetString("1000000000000000000000000000", 10)
	return &token.Token{
		Symbol:      "halo",
		Decimals:    18,
		TotalSupply: totalSupply,
		Balances:    map[string]*big.Int{"incentive": totalSupply},
	}
}

func testGenRouterState() *schema.RouterState {
	return &schema.RouterState{
		Router: "0x1",
		Pools: map[string]*schema.Pool{
			"0xP1": {
				TokenXTag: "ar",
				TokenYTag: "usdt",
				FeeRatio:  "0.003",
			},
			"0xP2": {
				TokenXTag: "eth",
				TokenYTag: "usdt",
				FeeRatio:  "0.003",
			},
		},
	}
}
func testGenOracle() *schema.Oracle {
	return &schema.Oracle{
		EverTokens: map[string]everSchema.TokenInfo{
			"ar": {
				Decimals: 12,
			},
			"usdt": {
				Decimals: 6,
			},
			"eth": {
				Decimals: 18,
			},
		},
	}

}
func testGenState() *schema.StateForProposal {
	t := testGenToken()
	r := testGenRouterState()
	return &schema.StateForProposal{
		Dapp:           "halo_test",
		ChainID:        "5",
		Govern:         "",
		FeeRecipient:   "",
		RouterMinStake: "80000000000000000000000",
		Routers:        []string{r.Router},
		RouterStates: map[string]*schema.RouterState{
			r.Router: r,
		},
		Token: t,
	}
}

func testGenInitData() string {
	return `{
		"name": "",
		"router": "0x1",
		"pool": "0xP2",
		"baseToken": "usdt",
		"totalSupply": "10000000",
		"start": 1000,
		"end": 2000
	}`
}

func TestExecute(t *testing.T) {
	state1 := testGenState()
	initData := testGenInitData()
	oracle := testGenOracle()
	tx1 := &schema.Transaction{
		Nonce:  "1005000",
		Action: schema.TxActionSwap,
		SwapOrder: &schema.SwapOrder{
			Items: []*schema.SwapOrderItem{
				{
					PoolID:    "0xP2",
					User:      "0xa",
					Lp:        "0xL1",
					TokenIn:   "usdt",
					AmountIn:  big.NewInt(1000000000),
					TokenOut:  "eth",
					AmountOut: big.NewInt(100),
				},
				{
					PoolID:    "0xP2",
					User:      "0xa",
					Lp:        "0xL2",
					TokenIn:   "usdt",
					AmountIn:  big.NewInt(2000000000),
					TokenOut:  "eth",
					AmountOut: big.NewInt(200),
				},
			},
		},
	}
	state2, localState2, _, err := Execute(tx1, state1, oracle, "", initData)
	t.Log("state2 token balance", state2.Token.Balances, "\n")
	t.Log("localState2", localState2, "\n")
	assert.NoError(t, err)
}
