package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
)

type DecimalSlice []*apd.Decimal

func (p DecimalSlice) Len() int           { return len(p) }
func (p DecimalSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p DecimalSlice) Less(i, j int) bool { return p[i].Cmp(p[j]) < 0 }

func StringToDecimal(s string) (*apd.Decimal, error) {
	d := new(apd.Decimal)
	_, c, err := d.SetString(s)
	if err != nil {
		return nil, err
	}
	if c.String() != "" {
		return nil, ERR_INVALID_NUMBER
	}
	return d, nil
}

func BigDotIntToDecimal(b *big.Int) (*apd.Decimal, error) {
	d := new(apd.Decimal)
	_, c, err := d.SetString(b.String())
	if err != nil {
		return nil, err
	}
	if c.String() != "" {
		return nil, ERR_INVALID_NUMBER
	}
	return d, nil
}

func DecimalToBigDotInt(d *apd.Decimal, isRoundUp bool) (*big.Int, error) {
	var integ, frac apd.Decimal
	d.Modf(&integ, &frac)

	v := integ.Coeff.MathBigInt()

	var ed apd.ErrDecimal
	if err := ed.Err(); err != nil {
		return nil, err
	}

	ten := big.NewInt(10)
	for i := int32(0); i < integ.Exponent; i++ {
		v.Mul(v, ten)
	}

	if d.Negative {
		log.Error("DecimalToBigDotInt no support negative.", "d", d)
		return nil, ERR_NO_IMPLEMENT
	}

	if !frac.IsZero() && isRoundUp {
		one := big.NewInt(1)
		v.Add(v, one)
	}

	return v, nil
}

func TokenXYForPool(tokenA, tokenB string) (string, string) {
	if tokenA < tokenB {
		return tokenA, tokenB
	}
	return tokenB, tokenA
}

func getFee(amountIn *big.Int, feeRatio *apd.Decimal, isRoundUp bool) (*big.Int, error) {

	amountIn_, err := BigDotIntToDecimal(amountIn)
	if err != nil {
		return nil, err
	}

	fee_ := new(apd.Decimal)
	_, err = roundUpContext.Mul(fee_, amountIn_, feeRatio)
	if err != nil {
		return nil, err
	}

	fee, err := DecimalToBigDotInt(fee_, isRoundUp)
	if err != nil {
		return nil, err
	}
	return fee, nil
}

// make sure fee > 1 , otherwise amountIn is too small.
func getAndCheckFee(amountIn *big.Int, feeRatio *apd.Decimal) (*big.Int, error) {
	fee, err := getFee(amountIn, feeRatio, true)
	if err != nil {
		return nil, ERR_INVALID_FEE
	}
	if fee.Cmp(big.NewInt(1)) != 1 {
		log.Error("fee is too small", "amoutIn", amountIn.String(), "fee", fee)
		return nil, ERR_INVALID_AMOUNT
	}
	return fee, nil
}

func summarySwapOuts(swapOuts []schema.SwapOutput) (*schema.SwapOutput, error) {

	if len(swapOuts) == 0 {
		return nil, ERR_INVALID_SWAPOUTS
	}

	so := &schema.SwapOutput{}
	for i, so_ := range swapOuts {
		if i == 0 {
			so.TokenIn = so_.TokenIn
			so.TokenOut = so_.TokenOut

			// make sure do not change swapOuts
			so.AmountIn = new(big.Int).Set(so_.AmountIn)
			so.AmountOut = new(big.Int).Set(so_.AmountOut)

			continue
		}

		if (so.TokenIn != so_.TokenIn) || (so.TokenOut != so_.TokenOut) {
			return nil, ERR_INVALID_SWAPOUTS
		}
		so.AmountIn.Add(so.AmountIn, so_.AmountIn)
		so.AmountOut.Add(so.AmountOut, so_.AmountOut)
	}
	return so, nil
}
