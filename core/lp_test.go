package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestLp(t *testing.T) {

	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool := &schema.Pool{
		TokenXTag: "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenYTag: "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		FeeRatio:  fee,
	}

	liquidity := "273861278752583"
	lowPrice := "1E-9"
	currentPrice := "3.005309963620149008176244919E-9"
	highPrice := "5E-9"
	lp, err := NewLp(pool.ID(), pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), liquidity, schema.PriceDirectionBoth)
	assert.NoError(t, err)
	t.Log("New lp:", lp)
	t.Log("New lp ID:", lp.ID())

	amountIn := big.NewInt(10 * 1000000)
	r, err := LpSwap(lp,
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		false)
	t.Log(r)

	amountIn = big.NewInt(1000000000000000000 / 10)
	r, err = LpSwap(lp,
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn,
		false)
	t.Log(r)

	amountOut := big.NewInt(293766517)
	r2, err := LpVerifySwap(lp,
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn,
		amountOut)
	t.Log(r2)

	amountOut = big.NewInt(293766518)
	r2, err = LpVerifySwap(lp,
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn,
		amountOut)
	t.Log(r2)
}

func TestLpID(t *testing.T) {

	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool := &schema.Pool{
		TokenXTag: "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenYTag: "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		FeeRatio:  fee,
	}

	liquidity := "273861278752583"
	lowPrice := "1E-9"
	currentPrice := "3.005309963620149008176244919E-9"
	highPrice := "5E-9"
	lp, err := NewLp(pool.ID(), pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), liquidity, schema.PriceDirectionBoth)
	assert.NoError(t, err)

	id := GetLpID(pool.ID(), "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6",
		testSqrtPrice(lowPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	assert.Equal(t, lp.ID(), id)
}
