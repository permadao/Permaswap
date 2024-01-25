package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestPoolSwapWithOneLp(t *testing.T) {
	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"0.003")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	liquidity := "273861278752583"
	lowPrice := "1E-9"
	currentPrice := "3E-9"
	highPrice := "5E-9"
	lp1, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice),
		liquidity, schema.PriceDirectionBoth)

	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	t.Log("lps:", GetPoolLps(pool, []string{}), "\n")

	priceUp, err := GetPoolCurrentPrice2(pool, schema.PriceDirectionUp)
	assert.NoError(t, err)
	t.Log("priceUp", priceUp)

	priceDown, err := GetPoolCurrentPrice2(pool, schema.PriceDirectionDown)
	assert.NoError(t, err)
	t.Log("priceDown", priceDown)

	amountIn := big.NewInt(1000000)
	swapOut, err := PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	t.Log(swapOut, "\n")

	amountIn = big.NewInt(1000 * 1000000)
	swapOut, err = PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	t.Log(swapOut, "\n")

	amountIn = big.NewInt(1000000000000000000 / 100)
	swapOut, err = PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	t.Log(swapOut, "\n")

	amountIn2 := new(big.Int)
	amountIn2, _ = amountIn2.SetString("100000000000000000000", 10)
	swapOut, err = PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn2,
		nil,
		false,
	)
	assert.EqualError(t, err, "err_out_of_range")
	t.Log(swapOut, "\n")

}

func TestPoolSwapWithFullRangeLp(t *testing.T) {
	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"0.003")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	liquidity := "273861278752583"
	lowSqrtPrice, _, _ := new(apd.Decimal).SetString(schema.FullRangeLowSqrtPrice)
	currentPrice := "3E-9"
	highSqrtPrice, _, _ := new(apd.Decimal).SetString(schema.FullRangeHighSqrtPrice)
	lp1, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		lowSqrtPrice, testSqrtPrice(currentPrice), highSqrtPrice,
		liquidity, schema.PriceDirectionBoth)

	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	t.Log("lps:", GetPoolLps(pool, []string{}), "\n")

	priceUp, err := GetPoolCurrentPrice2(pool, schema.PriceDirectionUp)
	assert.NoError(t, err)
	t.Log("priceUp", priceUp)

	priceDown, err := GetPoolCurrentPrice2(pool, schema.PriceDirectionDown)
	assert.NoError(t, err)
	t.Log("priceDown", priceDown)

	amountIn := big.NewInt(1000000)
	swapOut, err := PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	t.Log(swapOut, "\n")

	amountIn2 := new(big.Int)
	amountIn2, _ = amountIn2.SetString("100000000000000000000000000", 10) // 1,0000,0000 eth
	swapOut, err = PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn2,
		nil,
		false,
	)
	t.Log(swapOut, "\n")
}
func TestPoolSwapWithFullRangeLp2(t *testing.T) {
	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool, err := NewPool("arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		"0.003")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	liquidity := "198000000000000000" // (2200000000000000 * 17820000000000000000) ** 0.5

	amountX, _ := new(big.Int).SetString("2200000000000000", 10)
	amountY, _ := new(big.Int).SetString("17820000000000000000", 10)
	k := new(big.Int).Mul(amountX, amountY)
	t.Log("x:", amountX, "y:", amountY, "k:", k, "\n")

	lowSqrtPrice, _, _ := new(apd.Decimal).SetString(schema.FullRangeLowSqrtPrice)
	currentPrice := "8100"
	highSqrtPrice, _, _ := new(apd.Decimal).SetString(schema.FullRangeHighSqrtPrice)
	lp1, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		lowSqrtPrice, testSqrtPrice(currentPrice), highSqrtPrice,
		liquidity, schema.PriceDirectionBoth)

	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	//t.Log("lps:", GetPoolLps(pool, []string{}), "\n")

	t.Log("CurrentSqrtPrice:", lp1.CurrentSqrtPrice)
	amountIn := big.NewInt(1000000000000000)
	swapOut, err := PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	//t.Log(swapOut, "\n")
	amountX = amountX.Sub(amountX, swapOut[0].AmountOut)
	amountY = amountY.Add(amountY, swapOut[0].AmountIn)
	k = new(big.Int).Mul(amountX, amountY)
	t.Log("x:", amountX, "y:", amountY, "k:", k, "\n")

	t.Log("CurrentSqrtPrice:", lp1.CurrentSqrtPrice)
	amountIn = big.NewInt(100000000000000)
	swapOut, err = PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	//t.Log(swapOut, "\n")
	amountX = amountX.Sub(amountX, swapOut[0].AmountOut)
	amountY = amountY.Add(amountY, swapOut[0].AmountIn)
	k = new(big.Int).Mul(amountX, amountY)
	t.Log("x:", amountX, "y:", amountY, "k:", k, "\n")

	t.Log("CurrentSqrtPrice:", lp1.CurrentSqrtPrice)
	amountIn = big.NewInt(1000000000000)
	swapOut, err = PoolSwap(pool, "arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	//t.Log(swapOut, "\n")
	amountX = amountX.Add(amountX, swapOut[0].AmountIn)
	amountY = amountY.Sub(amountY, swapOut[0].AmountOut)
	k = new(big.Int).Mul(amountX, amountY)
	t.Log("x:", amountX, "y:", amountY, "k:", k, "\n")
}
func TestPoolSwapWithMultiLps(t *testing.T) {
	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"0.003")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	liquidity := "273861278752583"
	lowPrice := "1E-9"
	currentPrice := "3E-9"
	highPrice := "5E-9"
	lp1, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice),
		liquidity, schema.PriceDirectionBoth)
	amountX, amountY, err := LiquidityToAmount(liquidity, testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	t.Log("lp1", lp1.ID(), "low", lowPrice, "current", currentPrice, "highPrice", highPrice, "amountX", amountX, "amountY", amountY)
	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	liquidity = "5000000000000"
	lowPrice = "2E-9"
	currentPrice = "3E-9"
	highPrice = "4E-9"
	lp2, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice),
		liquidity, schema.PriceDirectionBoth)
	amountX, amountY, err = LiquidityToAmount(liquidity, testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	t.Log("lp2", lp2.ID(), "low", lowPrice, "current", currentPrice, "highPrice", highPrice, "amountX", amountX, "amountY", amountY)
	err = PoolAddLiquidity(pool, lp2)
	assert.NoError(t, err)

	liquidity = "547722557505166"
	lowPrice = "1E-9"
	currentPrice = "2E-9"
	highPrice = "25E-10"
	lp3, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice),
		liquidity, schema.PriceDirectionBoth)
	amountX, amountY, err = LiquidityToAmount(liquidity, testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	t.Log("lp3", lp3.ID(), "low", lowPrice, "current", currentPrice, "highPrice", highPrice, "amountX", amountX, "amountY", amountY)
	err = PoolAddLiquidity(pool, lp3)
	assert.NoError(t, err)

	lps := GetPoolLps(pool, []string{})
	t.Log("lps:", len(lps), lps, "\n")

	// 1e-06 usdt buy eth
	amountIn := big.NewInt(1)
	swapOut, err := PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		true,
	)
	assert.EqualError(t, err, "err_invalid_amount")
	t.Log(len(swapOut), swapOut, "\n")

	// 1 usdt buy eth
	amountIn = big.NewInt(1000000)
	swapOut, err = PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		true,
	)
	assert.NoError(t, err)
	t.Log("lp3.CurrentSqrtPrice", lp3.CurrentSqrtPrice)
	t.Log(len(swapOut), swapOut, "\n")

	// 3000 usdt buy eth
	amountIn = big.NewInt(3000 * 1000000)
	swapOut, err = PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		true,
	)
	assert.NoError(t, err)
	t.Log(len(swapOut), swapOut, "\n")

	// 10000 usdt buy eth
	amountIn = big.NewInt(10000 * 1000000)
	swapOut, err = PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		true,
	)
	assert.EqualError(t, err, "err_out_of_range")
	t.Log(len(swapOut), swapOut, "\n")

	// sell 0.01 eth
	amountIn = big.NewInt(1000000000000000000 / 100)
	swapOut, err = PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn,
		nil,
		true,
	)
	assert.NoError(t, err)
	t.Log(len(swapOut), swapOut, "\n")

	so, err := summarySwapOuts(swapOut)
	assert.NoError(t, err)
	t.Log("summary swapOut:", so, "\n")

	// sell 100 eth
	amountIn2 := new(big.Int)
	amountIn2, _ = amountIn2.SetString("100000000000000000000", 10)
	swapOut, err = PoolSwap(pool, "ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		amountIn2,
		nil,
		true,
	)
	assert.EqualError(t, err, "err_out_of_range")
	t.Log(len(swapOut), swapOut, "\n")

	assert.Equal(t, lp3.CurrentSqrtPrice.String(), testSqrtPrice(currentPrice).String())

	priceUp, err := GetPoolCurrentPrice2(pool, schema.PriceDirectionUp)
	assert.NoError(t, err)
	t.Log("priceUp", priceUp)

	priceDown, err := GetPoolCurrentPrice2(pool, schema.PriceDirectionDown)
	assert.NoError(t, err)
	t.Log("priceDown", priceDown)
}

func TestPoolSwapWithBlankTicks(t *testing.T) {
	fee := new(apd.Decimal)
	_, _, err := fee.SetString("0.003")
	assert.NoError(t, err)

	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"0.003")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	liquidity := "273861278752583"
	lowPrice := "28E-10"
	currentPrice := "3E-9"
	highPrice := "4E-9"
	lp1, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice),
		liquidity, schema.PriceDirectionBoth)

	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	liquidity = "5000000000000"
	lowPrice = "2E-9"
	currentPrice = "25E-10"
	highPrice = "26E-10"
	lp1, err = NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice),
		liquidity, schema.PriceDirectionBoth)

	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	t.Log("lps:", GetPoolLps(pool, []string{}), "\n")

	amountIn := big.NewInt(1000 * 1000000)
	swapOut, err := PoolSwap(pool, "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		"ethereum-eth-0x0000000000000000000000000000000000000000",
		amountIn,
		nil,
		false,
	)
	assert.NoError(t, err)
	assert.Equal(t, len(swapOut), 2)
	t.Log(len(swapOut), swapOut, "\n")
}

func TestPoolSwapWith2Lps(t *testing.T) {
	tokenX := "ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	tokenY := "ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7"
	pool, err := NewPool(tokenX, tokenY, "0.0005")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	fee := testStringToDecimal("0.0005")

	liquidity := "1250000000000"
	lowSqrtPrice := "0.99498743710661995473447982100121"
	currentSqrtPrice := "0.99498743710747199342086543911350"
	highSqrtPrice := "1.0049875621120890270219264912760"
	lp1, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testStringToDecimal(lowSqrtPrice), testStringToDecimal(currentSqrtPrice), testStringToDecimal(highSqrtPrice),
		liquidity, schema.PriceDirectionBoth)
	t.Log("lpId:", lp1.ID())
	err = PoolAddLiquidity(pool, lp1)
	assert.NoError(t, err)

	liquidity = "2500000000000"
	lowSqrtPrice = "0.98994949366116653416118210694679"
	highSqrtPrice = "1.0099504938362077953363385917070"
	lp2, err := NewLp(poolID, pool.TokenXTag, pool.TokenYTag, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", fee,
		testStringToDecimal(lowSqrtPrice), testStringToDecimal(currentSqrtPrice), testStringToDecimal(highSqrtPrice),
		liquidity, schema.PriceDirectionBoth)
	t.Log("lpId:", lp2.ID())
	err = PoolAddLiquidity(pool, lp2)
	assert.NoError(t, err)

	lps := GetPoolLps(pool, []string{})
	t.Log("lps:", len(lps), lps, "\n")

	// sell 100 usdc
	amountIn := big.NewInt(100000000)
	swapOut, err := PoolSwap(pool, tokenX,
		tokenY,
		amountIn,
		nil,
		true,
	)
	t.Log(err)
	t.Log(len(swapOut), swapOut, "\n")
}
