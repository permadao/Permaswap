package core

import (
	"math/big"
	"testing"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
	"github.com/stretchr/testify/assert"
)

func testSqrtPrice(price string) *apd.Decimal {
	r, _ := SqrtPrice(price)
	return r
}
func TestConvert(t *testing.T) {
	oneInt := big.NewInt(1)
	oneDecimal, _, err := apd.NewFromString("1")
	assert.NoError(t, err)
	oneDecimalFromInt, err := BigDotIntToDecimal(oneInt)
	assert.NoError(t, err)
	assert.Equal(t, oneDecimal, oneDecimalFromInt, "oneDecimal equals oneDecimalFromOneInt")

	oneAndHalfDecimal, _, err := apd.NewFromString("1.5")
	assert.NoError(t, err)
	oneIntFromDecimal, err := DecimalToBigDotInt(oneAndHalfDecimal, false)
	assert.NoError(t, err)
	assert.Equal(t, oneIntFromDecimal, oneInt, "oneIntFromDecimal equals oneInt")
}

func TestSqrtPrice(t *testing.T) {
	price := "0.04"
	priceDecimal := testSqrtPrice(price)
	priceDecimal2, _, err := apd.NewFromString("0.2")
	assert.NoError(t, err)

	//will fail
	assert.NotEqual(t, priceDecimal, priceDecimal2, "the sqrt of price is 0.2")
}
func TestFullRange(t *testing.T) {
	fullRangeLowSqrtPrice := new(apd.Decimal)
	x, _, _ := new(apd.Decimal).SetString("2")
	y, _, _ := new(apd.Decimal).SetString("-64")
	roundUpContext.Pow(fullRangeLowSqrtPrice, x, y)
	t.Log("fullRangeLowSqrtPrice:", fullRangeLowSqrtPrice)
}

func TestSwap(t *testing.T) {
	startSqrtPrice := testSqrtPrice("3E-09")
	t.Log("startSqrtPrice:", startSqrtPrice)
	liquidity := big.NewInt(54772255750516)      // int((10**18 * 3000 * 10 **6) ** 0.5)+1
	amountX := big.NewInt(5 * 10000000000000000) // 0.05 eth
	newSqrtPrice, err := getNewSqrtPriceFromAmountXRoundUp(startSqrtPrice, liquidity, amountX, true)
	assert.NoError(t, err)
	newPrice, err := SqrtPriceToPrice(*newSqrtPrice)
	t.Log("newPrice_x:", newPrice)

	amountY := big.NewInt(100 * 1000000) // 100 usdt
	newSqrtPrice, err = getNewSqrtPriceFromAmountYRoundingDown(startSqrtPrice, liquidity, amountY, true)
	newPrice, err = SqrtPriceToPrice(*newSqrtPrice)
	t.Log("newPrice_y:", newPrice)

	amountOut, newSqrtPrice, err := SwapOut(startSqrtPrice, liquidity, amountX, true)
	amountIn, amountOut, _, _, err := SwapAmount(startSqrtPrice, newSqrtPrice, liquidity)
	t.Log("TestswapAmountDown: amountIn:", amountIn, "amountOut:", amountOut)

	newPrice, err = SqrtPriceToPrice(*newSqrtPrice)
	t.Log("newPrice:", newPrice, "amountOut:", amountOut)

	amountOut, newSqrtPrice, err = SwapOut(startSqrtPrice, liquidity, amountY, false)
	newPrice, err = SqrtPriceToPrice(*newSqrtPrice)
	t.Log("TestswapOut newPrice2:", newPrice, "amountOut2:", amountOut)

	amountIn, amountOut, _, _, err = SwapAmount(startSqrtPrice, newSqrtPrice, liquidity)
	t.Log("TestswapAmountUp: amountIn:", amountIn, "amountOut:", amountOut)
}

func TestSwap2(t *testing.T) {
	startSqrtPrice := testSqrtPrice("3E-09")
	endSqrtPrice := testSqrtPrice("1E-09")
	liquidity := big.NewInt(54772255750516) // int((10**18 * 3000 * 10 **6) ** 0.5)+1

	amountX, _, _, _, _ := swapAmountDown(startSqrtPrice, endSqrtPrice, liquidity)
	_, amountX2, _, _, _ := swapAmountUp(endSqrtPrice, startSqrtPrice, liquidity)

	t.Log("amountX:", amountX, "amountX2:", amountX2)
}

func TestSwap3(t *testing.T) {
	highSqrtPrice := testSqrtPrice("5E-09")
	currentSqrtPrice := testSqrtPrice("3E-09")
	lowSqrtPrice := testSqrtPrice("1E-09")
	liquidity := big.NewInt(242996656554273) // 242996656554273 is the result of TestAmountXY 1

	amountX, amountY, _, _, _ := SwapAmount(currentSqrtPrice, lowSqrtPrice, liquidity)
	t.Log("amountX:", amountX, "amountY:", amountY)

	amountY2, amountX2, _, _, _ := SwapAmount(currentSqrtPrice, highSqrtPrice, liquidity)
	t.Log("amountX2:", amountX2, "amountY2:", amountY2)

	liquidity2 := big.NewInt(242996656634653) // 242996656634653 is the result of TestAmountXY 2
	amountX3, amountY3, _, _, _ := SwapAmount(currentSqrtPrice, lowSqrtPrice, liquidity2)
	t.Log("amountX3:", amountX3, "amountY3:", amountY3)
	amountY4, amountX4, _, _, _ := SwapAmount(currentSqrtPrice, highSqrtPrice, liquidity2)
	t.Log("amountX4:", amountX4, "amountY4:", amountY4)
}
func TestAmountXY(t *testing.T) {
	lowPrice := "1E-09"
	currentPrice := "3E-09"
	highPrice := "5E-09"

	amountX := "1000000000000000000"
	liquidity, err := LiquidityFromAmountX(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountX)
	amountY, err := AmountYFromAmountX(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountX)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY, "liquidity:", liquidity)

	//amountY = "5625246034"
	liquidity, err = LiquidityFromAmountY(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountY)
	amountX, err = AmountXFromAmountY(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountY)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY, "liquidity:", liquidity)
}

func TestAmountXY2(t *testing.T) {
	lowPrice := "1E-09"
	currentPrice := "3E-09"
	highPrice := "5E-09"

	amountY := "1000000000000000000"
	liquidity, err := LiquidityFromAmountY(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountY)
	amountX, err := AmountXFromAmountY(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountY)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY, "liquidity:", liquidity)

	liquidity, err = LiquidityFromAmountX(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountX)
	amountY, err = AmountYFromAmountX(testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), amountX)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY, "liquidity:", liquidity)
}

func TestLiquidityToAmount(t *testing.T) {
	lowPrice := "1E-09"
	currentPrice := "3E-09"
	highPrice := "5E-09"

	liquidity := "54772255750516"

	amountX, amountY, err := LiquidityToAmount(liquidity, testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

	currentPrice = lowPrice
	amountX, amountY, err = LiquidityToAmount(liquidity, testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

	currentPrice = highPrice
	amountX, amountY, err = LiquidityToAmount(liquidity, testSqrtPrice(lowPrice), testSqrtPrice(currentPrice), testSqrtPrice(highPrice), schema.PriceDirectionBoth)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

}

func TestLiquidityToAmount2(t *testing.T) {
	lowSqrtPrice, _ := StringToDecimal("0.000022360679774997896964091736687313")
	currentSqrtPrice, _ := StringToDecimal("0.000029255380953307190093117837024707")
	highSqrtPrice, _ := StringToDecimal("0.000070710678118654752440084436210485")
	liquidity := "44721359549996"

	amountX, amountY, err := LiquidityToAmount(liquidity, lowSqrtPrice, currentSqrtPrice, highSqrtPrice, schema.PriceDirectionBoth)
	assert.NoError(t, err)

	assert.Equal(t, liquidity, "44721359549996")
	assert.Equal(t, lowSqrtPrice.String(), "0.000022360679774997896964091736687313")
	assert.Equal(t, currentSqrtPrice.String(), "0.000029255380953307190093117837024707")
	assert.Equal(t, highSqrtPrice.String(), "0.000070710678118654752440084436210485")
	t.Log(liquidity, lowSqrtPrice, currentSqrtPrice, highSqrtPrice)

	t.Log("amountX:", amountX, "amountY:", amountY)

	amountX, amountY, err = LiquidityToAmount2(liquidity, lowSqrtPrice, currentSqrtPrice, highSqrtPrice)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

	amountX, amountY, err = LiquidityToAmount(liquidity, lowSqrtPrice, lowSqrtPrice, highSqrtPrice, schema.PriceDirectionBoth)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

	amountX, amountY, err = LiquidityToAmount2(liquidity, lowSqrtPrice, lowSqrtPrice, highSqrtPrice)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

	amountX, amountY, err = LiquidityToAmount(liquidity, lowSqrtPrice, highSqrtPrice, highSqrtPrice, schema.PriceDirectionBoth)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)

	amountX, amountY, err = LiquidityToAmount2(liquidity, lowSqrtPrice, highSqrtPrice, highSqrtPrice)
	assert.NoError(t, err)
	t.Log("amountX:", amountX, "amountY:", amountY)
}

func TestLiquidityFromAmount(t *testing.T) {
	x := "10000000000000000000" // 10 eth
	y := "20000000000"          // 20000 usdt
	liquidity, _ := LiquidityFromAmount(x, y)
	assert.Equal(t, liquidity, "447213595499958")
}

func TestQuotientGreaterThan(t *testing.T) {
	lowSqrtPrice, _ := StringToDecimal("1.0001000025000002")
	highSqrtPrice, _ := StringToDecimal("1.00015")
	r := QuotientGreaterThan(highSqrtPrice, lowSqrtPrice, schema.MinSqrtPriceFactor)
	assert.Equal(t, r, false)

	lowSqrtPrice, _ = StringToDecimal("1.0005001125150024")
	highSqrtPrice, _ = StringToDecimal("1.0005501375206283")
	r = QuotientGreaterThan(highSqrtPrice, lowSqrtPrice, schema.MinSqrtPriceFactor)
	assert.Equal(t, r, true)

	lowSqrtPrice, _ = StringToDecimal("2")
	highSqrtPrice, _ = StringToDecimal("1")
	r = QuotientGreaterThan(highSqrtPrice, lowSqrtPrice, schema.MinSqrtPriceFactor)
	assert.Equal(t, r, false)
}

func TestSqrtPriceToPriceWithFee(t *testing.T) {
	sqrtPrice := testSqrtPrice("0.04")

	fee := new(apd.Decimal)
	fee.SetString("0.003")

	price, err := SqrtPriceToPriceWithFee(*sqrtPrice, *fee, schema.PriceDirectionUp)
	assert.NoError(t, err)
	t.Log("price:", price, "sqrtPrice", sqrtPrice.String(), "fee", fee.String())

	price, err = SqrtPriceToPriceWithFee(*sqrtPrice, *fee, schema.PriceDirectionDown)
	assert.NoError(t, err)
	t.Log("price:", price, "sqrtPrice", sqrtPrice.String(), "fee", fee.String())
}
