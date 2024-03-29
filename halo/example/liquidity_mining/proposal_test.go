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

	tx2 := &schema.Transaction{
		Nonce:  "1105000",
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
					Lp:        "0xL3",
					TokenIn:   "usdt",
					AmountIn:  big.NewInt(2000000000),
					TokenOut:  "eth",
					AmountOut: big.NewInt(200),
				},
			},
		},
	}

	state3, localState3, _, err := Execute(tx2, state2, oracle, localState2, initData)
	t.Log("state3 token balance", state3.Token.Balances, "\n")
	t.Log("localState3", localState3, "\n")
	assert.NoError(t, err)

	tx3 := &schema.Transaction{
		Nonce:  "1115000",
		Action: schema.TxActionSwap,
		SwapOrder: &schema.SwapOrder{
			Items: []*schema.SwapOrderItem{
				{
					PoolID:    "0xP1",
					User:      "0xa",
					Lp:        "0xL2",
					TokenIn:   "usdt",
					AmountIn:  big.NewInt(1000000000),
					TokenOut:  "ar",
					AmountOut: big.NewInt(1000),
				},
			},
		},
	}
	state4, localState4, _, err := Execute(tx3, state3, oracle, localState3, initData)
	t.Log("state3 token balance", state4.Token.Balances, "\n")
	t.Log("localState3", localState4, "\n")
	assert.NoError(t, err)
}
