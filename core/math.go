package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
)

const (
	PRECISION uint32 = 32
)

var roundDownContext = &apd.Context{
	Precision:   PRECISION,
	Rounding:    apd.RoundDown,
	MaxExponent: apd.MaxExponent,
	MinExponent: apd.MinExponent,
	Traps:       apd.DefaultTraps,
}

var roundUpContext = &apd.Context{
	Precision:   PRECISION,
	Rounding:    apd.RoundUp,
	MaxExponent: apd.MaxExponent,
	MinExponent: apd.MinExponent,
	Traps:       apd.DefaultTraps,
}

func SqrtPrice(price string) (*apd.Decimal, error) {
	p, err := StringToDecimal(price)
	if err != nil {
		return nil, err
	}

	c := apd.BaseContext.WithPrecision(PRECISION)
	s := new(apd.Decimal)
	_, err = c.Sqrt(s, p)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func SqrtPriceToPrice(sqrtPrice apd.Decimal) (string, error) {
	price := new(apd.Decimal)
	c := apd.BaseContext.WithPrecision(PRECISION)
	_, err := c.Mul(price, &sqrtPrice, &sqrtPrice)
	if err != nil {
		return "", err
	}
	return price.String(), nil
}

func getNewSqrtPriceFromAmountXRoundUp(startSqrtPrice *apd.Decimal, liquidity, amount *big.Int, add bool) (*apd.Decimal, error) {
	// getNextSqrtPriceFromAmount0RoundingUp: l*s/(a*s+l)
	l, err := BigDotIntToDecimal(liquidity)
	if err != nil {
		return nil, err
	}

	a, err := BigDotIntToDecimal(amount)
	if err != nil {
		return nil, err
	}
	if add {
		product := new(apd.Decimal)
		_, err = roundUpContext.Mul(product, startSqrtPrice, l)
		if err != nil {
			return nil, err
		}

		product2 := new(apd.Decimal)
		_, err = roundDownContext.Mul(product2, startSqrtPrice, a)
		if err != nil {
			return nil, err
		}

		denominator := new(apd.Decimal)
		_, err = roundDownContext.Add(denominator, product2, l)
		if err != nil {
			return nil, err
		}

		quotient := new(apd.Decimal)
		_, err = roundUpContext.Quo(quotient, product, denominator)
		if err != nil {
			return nil, err
		}
		return quotient, nil
	} else {
		return nil, ERR_NO_IMPLEMENT
	}
}

func getNewSqrtPriceFromAmountYRoundingDown(startSqrtPrice *apd.Decimal, liquidity, amount *big.Int, add bool) (*apd.Decimal, error) {
	// getNextSqrtPriceFromAmount1RoundingDown: a/l + s
	l, err := BigDotIntToDecimal(liquidity)
	if err != nil {
		return nil, err
	}

	a, err := BigDotIntToDecimal(amount)
	if err != nil {
		return nil, err
	}

	if add {
		quotient := new(apd.Decimal)
		_, err := roundDownContext.Quo(quotient, a, l)
		if err != nil {
			return nil, err
		}
		sum := new(apd.Decimal)
		_, err = roundDownContext.Add(sum, quotient, startSqrtPrice)
		if err != nil {
			return nil, err
		}
		return sum, nil
	} else {
		return nil, ERR_NO_IMPLEMENT
	}
}

func SwapOut(startSqrtPrice *apd.Decimal, liquidity, amountIn *big.Int, tokenInIsX bool) (amountOut *big.Int, endSqrtPrice *apd.Decimal, err error) {
	// x: l * (1/s - 1/e), y: l*(s - e);

	l, err := BigDotIntToDecimal(liquidity)
	if err != nil {
		return nil, nil, err
	}

	if tokenInIsX {
		endSqrtPrice, err := getNewSqrtPriceFromAmountXRoundUp(startSqrtPrice, liquidity, amountIn, true)
		if err != nil {
			return nil, nil, err
		}

		difference := new(apd.Decimal)
		_, err = roundDownContext.Sub(difference, startSqrtPrice, endSqrtPrice)
		if err != nil {
			return nil, nil, err
		}

		product := new(apd.Decimal)
		_, err = roundDownContext.Mul(product, l, difference)
		if err != nil {
			return nil, nil, err
		}

		amountOut, err := DecimalToBigDotInt(product, false)
		if err != nil {
			return nil, nil, err
		}
		return amountOut, endSqrtPrice, nil

	} else {
		one := apd.New(1, 0)
		endSqrtPrice, err := getNewSqrtPriceFromAmountYRoundingDown(startSqrtPrice, liquidity, amountIn, true)
		if err != nil {
			return nil, nil, err
		}
		reciprocal := new(apd.Decimal)
		_, err = roundDownContext.Quo(reciprocal, one, startSqrtPrice)
		if err != nil {
			return nil, nil, err
		}

		reciprocal2 := new(apd.Decimal)
		_, err = roundUpContext.Quo(reciprocal2, one, endSqrtPrice)
		if err != nil {
			return nil, nil, err
		}

		difference := new(apd.Decimal)
		_, err = roundDownContext.Sub(difference, reciprocal, reciprocal2)
		if err != nil {
			return nil, nil, err
		}

		product := new(apd.Decimal)
		_, err = roundDownContext.Mul(product, l, difference)
		if err != nil {
			return nil, nil, err
		}
		amountOut, err := DecimalToBigDotInt(product, false)
		if err != nil {
			return nil, nil, err
		}
		return amountOut, endSqrtPrice, nil
	}
}

func SwapAmount(startSqrtPrice, endSqrtPrice *apd.Decimal, liquidity *big.Int) (amountIn, amountOut *big.Int, amountInDecimal, amountOutDecimal *apd.Decimal, err error) {
	if startSqrtPrice.Cmp(endSqrtPrice) == 0 {
		return nil, nil, nil, nil, ERR_INVALID_PRICE
	}

	if startSqrtPrice.Cmp(endSqrtPrice) == -1 {
		return swapAmountUp(startSqrtPrice, endSqrtPrice, liquidity)
	} else {
		return swapAmountDown(startSqrtPrice, endSqrtPrice, liquidity)
	}
}

func swapAmountUp(startSqrtPrice, endSqrtPrice *apd.Decimal, liquidity *big.Int) (amountIn, amountOut *big.Int, amountInDecimal, amountOutDecimal *apd.Decimal, err error) {
	// end_price > start_price
	// token_in is y; y = l * (e - s); round_up;
	// token_out is x; x = l * (1/s - 1/e); round_down.

	l, err := BigDotIntToDecimal(liquidity)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	difference := new(apd.Decimal)
	_, err = roundUpContext.Sub(difference, endSqrtPrice, startSqrtPrice)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	product := new(apd.Decimal)
	_, err = roundUpContext.Mul(product, l, difference)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	amountIn, err = DecimalToBigDotInt(product, true)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	one := apd.New(1, 0)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	reciprocal := new(apd.Decimal)
	_, err = roundDownContext.Quo(reciprocal, one, startSqrtPrice)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	reciprocal2 := new(apd.Decimal)
	_, err = roundUpContext.Quo(reciprocal2, one, endSqrtPrice)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	difference2 := new(apd.Decimal)
	_, err = roundDownContext.Sub(difference2, reciprocal, reciprocal2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	product2 := new(apd.Decimal)
	_, err = roundDownContext.Mul(product2, l, difference2)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	amountOut, err = DecimalToBigDotInt(product2, false)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return amountIn, amountOut, product, product2, nil
}

func swapAmountDown(startSqrtPrice, endSqrtPrice *apd.Decimal, liquidity *big.Int) (amountIn, amountOut *big.Int, amountInDecimal, amountOutDecimal *apd.Decimal, err error) {
	// end_price < start_price
	// token_in is x; x = l * (1/e - 1/s); round_up;
	// token_out is y; y = l * (s - e); round_down.

	l, err := BigDotIntToDecimal(liquidity)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	one := apd.New(1, 0)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	reciprocal := new(apd.Decimal)
	_, err = roundUpContext.Quo(reciprocal, one, endSqrtPrice)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	reciprocal2 := new(apd.Decimal)
	_, err = roundDownContext.Quo(reciprocal2, one, startSqrtPrice)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	difference := new(apd.Decimal)
	_, err = roundUpContext.Sub(difference, reciprocal, reciprocal2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	product := new(apd.Decimal)
	_, err = roundUpContext.Mul(product, l, difference)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	amountIn, err = DecimalToBigDotInt(product, true)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	difference2 := new(apd.Decimal)
	_, err = roundDownContext.Sub(difference2, startSqrtPrice, endSqrtPrice)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	product2 := new(apd.Decimal)
	_, err = roundDownContext.Mul(product2, l, difference2)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	amountOut, err = DecimalToBigDotInt(product2, false)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return amountIn, amountOut, product, product2, nil
}

func getLiquidity(liquidity string) (*big.Int, error) {
	liquidity_ := new(big.Int)
	liquidity_, ok := liquidity_.SetString(liquidity, 10)
	if !ok {
		return nil, ERR_INVALID_LIQUIDITY
	}
	if liquidity_.Cmp(big.NewInt(0)) != 1 {
		return nil, ERR_INVALID_LIQUIDITY
	}
	return liquidity_, nil
}

func getAmount(amount string) (*big.Int, error) {
	amount_ := new(big.Int)
	amount_, ok := amount_.SetString(amount, 10)
	if !ok {
		return nil, ERR_INVALID_AMOUNT
	}
	if amount_.Cmp(big.NewInt(0)) != 1 {
		return nil, ERR_INVALID_AMOUNT
	}
	return amount_, nil
}

func getAmount2(amount string) (*apd.Decimal, error) {
	amount_, err := StringToDecimal(amount)
	if err != nil {
		return nil, ERR_INVALID_AMOUNT
	}
	if amount_.Cmp(apd.New(0, 0)) != 1 {
		return nil, ERR_INVALID_AMOUNT
	}
	return amount_, nil
}

func LiquidityToAmount(liquidity string, lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal, priceDirection string) (amountX, amountY string, err error) {
	liquidity_, err := getLiquidity(liquidity)
	if err != nil {
		return "", "", err
	}

	one := big.NewInt(1)
	amountX = "0"
	amountY = "0"
	if currentSqrtPrice.Cmp(lowSqrtPrice) == 0 {
		amountX_, _, _, _, err := swapAmountDown(highSqrtPrice, currentSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountX_ = amountX_.Add(amountX_, one)
		amountX = amountX_.String()
	} else if currentSqrtPrice.Cmp(highSqrtPrice) == 0 {
		amountY_, _, _, _, err := swapAmountUp(lowSqrtPrice, currentSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountY_ = amountY_.Add(amountY_, one)
		amountY = amountY_.String()
	} else {
		amountX_, _, _, _, err := swapAmountDown(highSqrtPrice, currentSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountX_ = amountX_.Add(amountX_, one)
		amountX = amountX_.String()

		amountY_, _, _, _, err := swapAmountUp(lowSqrtPrice, currentSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountY_ = amountY_.Add(amountY_, one)
		amountY = amountY_.String()
	}

	if priceDirection == "up" {
		amountX_, _, _, _, err := swapAmountDown(highSqrtPrice, currentSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountX_ = amountX_.Add(amountX_, one)
		amountX = amountX_.String()
		amountY = "0"
	}

	if priceDirection == "down" {
		amountY_, _, _, _, err := swapAmountUp(lowSqrtPrice, currentSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountY_ = amountY_.Add(amountY_, one)
		amountY = amountY_.String()
		amountX = "0"
	}

	return amountX, amountY, nil
}

func LiquidityFromAmountY(lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal, amountY string) (liquidity string, err error) {
	// l = y /(s - e); round_down.
	amount, err := getAmount2(amountY)
	if err != nil {
		return "", err
	}

	liquidity_ := new(apd.Decimal)
	if currentSqrtPrice == lowSqrtPrice {
		return "", ERR_INVALID_PRICE
	} else {
		difference := new(apd.Decimal)
		_, err = roundUpContext.Sub(difference, currentSqrtPrice, lowSqrtPrice)
		if err != nil {
			return "", err
		}
		roundDownContext.Quo(liquidity_, amount, difference)
	}

	l, _ := DecimalToBigDotInt(liquidity_, false)
	liquidity = l.String()
	return liquidity, nil
}

func AmountXFromAmountY(lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal, amountY string) (amountX string, err error) {
	liquidity, err := LiquidityFromAmountY(lowSqrtPrice, currentSqrtPrice, highSqrtPrice, amountY)
	if err != nil {
		return "", err
	}

	amountX, _, err = LiquidityToAmount(liquidity, lowSqrtPrice, currentSqrtPrice, highSqrtPrice, schema.PriceDirectionBoth)
	if err != nil {
		return "", err
	}

	return amountX, nil
}

func LiquidityFromAmountX(lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal, amountX string) (liquidity string, err error) {
	// l = x /(1/s - 1/e) = x * s*e/(e-s) ; round_down.
	amount, err := getAmount2(amountX)
	if err != nil {
		return "", err
	}

	liquidity_ := new(apd.Decimal)
	if currentSqrtPrice == highSqrtPrice {
		return "", ERR_INVALID_PRICE
	} else {

		product := new(apd.Decimal)
		_, err = roundDownContext.Mul(product, currentSqrtPrice, highSqrtPrice)
		if err != nil {
			return "", err
		}

		difference := new(apd.Decimal)
		_, err = roundUpContext.Sub(difference, highSqrtPrice, currentSqrtPrice)
		if err != nil {
			return "", err
		}

		quotient := new(apd.Decimal)
		roundDownContext.Quo(quotient, product, difference)

		_, err = roundDownContext.Mul(liquidity_, amount, quotient)
		if err != nil {
			return "", err
		}
	}

	l, _ := DecimalToBigDotInt(liquidity_, false)
	liquidity = l.String()
	return liquidity, nil
}

func AmountYFromAmountX(lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal, amountX string) (amountY string, err error) {
	liquidity, err := LiquidityFromAmountX(lowSqrtPrice, currentSqrtPrice, highSqrtPrice, amountX)
	if err != nil {
		return "", err
	}

	_, amountY, err = LiquidityToAmount(liquidity, lowSqrtPrice, currentSqrtPrice, highSqrtPrice, schema.PriceDirectionBoth)
	if err != nil {
		return "", err
	}

	return amountY, nil
}

func LiquidityFromAmount(amountX, amountY string) (string, error) {
	// l = (x * y) ** 0.5

	c := apd.BaseContext.WithPrecision(PRECISION)

	amountX2, err := getAmount2(amountX)
	if err != nil {
		return "", err
	}

	amountY2, err := getAmount2(amountY)
	if err != nil {
		return "", err
	}

	product := new(apd.Decimal)
	_, err = c.Mul(product, amountX2, amountY2)
	if err != nil {
		return "", err
	}

	s := new(apd.Decimal)
	_, err = c.Sqrt(s, product)
	if err != nil {
		return "", err
	}

	l, err := DecimalToBigDotInt(s, true)
	if err != nil {
		return "", err
	}
	return l.String(), nil
}

// deprecated
func LiquidityToAmount2(liquidity string, lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal) (amountX, amountY string, err error) {
	liquidity_, err := getLiquidity(liquidity)
	if err != nil {
		return "", "", err
	}

	one := big.NewInt(1)
	amountX = "0"
	amountY = "0"
	if currentSqrtPrice.Cmp(lowSqrtPrice) == 0 {
		_, amountX_, _, _, err := swapAmountUp(currentSqrtPrice, highSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountX_ = amountX_.Add(amountX_, one)
		amountX = amountX_.String()
	} else if currentSqrtPrice.Cmp(highSqrtPrice) == 0 {
		_, amountY_, _, _, err := swapAmountDown(currentSqrtPrice, lowSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountY_ = amountY_.Add(amountY_, one)
		amountY = amountY_.String()
	} else {
		_, amountX_, _, _, err := swapAmountUp(currentSqrtPrice, highSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountX_ = amountX_.Add(amountX_, one)
		amountX = amountX_.String()

		_, amountY_, _, _, err := swapAmountDown(currentSqrtPrice, lowSqrtPrice, liquidity_)
		if err != nil {
			return "", "", ERR_INVALID_PRICE
		}
		amountY_ = amountY_.Add(amountY_, one)
		amountY = amountY_.String()
	}

	return amountX, amountY, nil
}

func QuotientGreaterThan(dividend, divisor *apd.Decimal, minimum string) bool {
	quotient := new(apd.Decimal)
	_, err := roundDownContext.Quo(quotient, dividend, divisor)
	if err != nil {
		return false
	}
	min, _ := StringToDecimal(minimum)
	if quotient.Cmp(min) != 1 {
		return false
	}

	return true
}

func SqrtPriceToPriceWithFee(sqrtPrice, feeRatio apd.Decimal, priceDirection string) (string, error) {
	price := new(apd.Decimal)

	c := apd.BaseContext.WithPrecision(PRECISION)
	_, err := c.Mul(price, &sqrtPrice, &sqrtPrice)
	if err != nil {
		return "", err
	}
	difference := new(apd.Decimal)
	_, err = c.Sub(difference, apd.New(1, 0), &feeRatio)
	if err != nil {
		return "", err
	}

	if priceDirection == schema.PriceDirectionDown {
		_, err = c.Mul(price, price, difference)
		if err != nil {
			return "", err
		}
	} else if priceDirection == schema.PriceDirectionUp {
		reciprocal := new(apd.Decimal)
		_, err = c.Quo(reciprocal, apd.New(1, 0), price)
		if err != nil {
			return "", err
		}

		_, err = c.Mul(price, reciprocal, difference)
		if err != nil {
			return "", err
		}

	} else {
		return "", ERR_INVALID_PRICE_DIRECTION
	}

	return price.Text('f'), nil
}
