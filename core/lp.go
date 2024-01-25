package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/everVision/everpay-kits/utils"
	"github.com/permadao/permaswap/core/schema"
)

func NewLp(poolID, tokenXTag, tokenYTag, address string,
	feeRatio, lowSqrtPrice, currentSqrtPrice, highSqrtPrice *apd.Decimal,
	liquidity, priceDirection string) (*schema.Lp, error) {

	if GetPoolID(tokenXTag, tokenYTag, feeRatio) != poolID {
		return nil, ERR_INVALID_POOL
	}

	_, acct, err := utils.IDCheck(address)
	if err != nil {
		return nil, err
	}

	liquidity_ := new(big.Int)
	liquidity_, ok := liquidity_.SetString(liquidity, 10)
	if !ok {
		return nil, ERR_INVALID_LIQUIDITY
	}
	if liquidity_.Cmp(big.NewInt(0)) != 1 {
		return nil, ERR_INVALID_LIQUIDITY
	}

	//TODO: did lowSqrtPrice should +1 and highSqrtPrice -1?

	if (currentSqrtPrice.Cmp(lowSqrtPrice) == -1) || (currentSqrtPrice.Cmp(highSqrtPrice) == 1) ||
		(lowSqrtPrice.Cmp(highSqrtPrice) == 0) || (lowSqrtPrice.Cmp(highSqrtPrice) == 1) {
		return nil, ERR_INVALID_PRICE
	}

	min, _, _ := new(apd.Decimal).SetString(schema.FullRangeLowSqrtPrice)
	max, _, _ := new(apd.Decimal).SetString(schema.FullRangeHighSqrtPrice)
	if lowSqrtPrice.Cmp(min) == -1 || highSqrtPrice.Cmp(max) == 1 {
		return nil, ERR_INVALID_PRICE
	}

	if !QuotientGreaterThan(highSqrtPrice, lowSqrtPrice, schema.MinSqrtPriceFactor) {
		return nil, ERR_INVALID_PRICE
	}

	if (priceDirection != schema.PriceDirectionUp) && (priceDirection != schema.PriceDirectionDown) &&
		(priceDirection != schema.PriceDirectionBoth) {
		return nil, ERR_INVALID_PRICE_DIRECTION
	}

	return &schema.Lp{
		PoolID:           poolID,
		TokenXTag:        tokenXTag,
		TokenYTag:        tokenYTag,
		FeeRatio:         feeRatio,
		AccID:            acct,
		Liquidity:        liquidity_,
		LowSqrtPrice:     lowSqrtPrice,
		CurrentSqrtPrice: currentSqrtPrice,
		HighSqrtPrice:    highSqrtPrice,
		PriceDirection:   priceDirection,
	}, nil
}

func LpLowPrice(lp schema.Lp) string {
	lowPrice, _ := SqrtPriceToPrice(*lp.LowSqrtPrice)
	return lowPrice
}

func LpCurrentPrice(lp schema.Lp) string {
	currentPrice, _ := SqrtPriceToPrice(*lp.CurrentSqrtPrice)
	return currentPrice
}

func LpHighPrice(lp schema.Lp) string {
	highPrice, _ := SqrtPriceToPrice(*lp.HighSqrtPrice)
	return highPrice
}

func LpSwap(lp *schema.Lp, tokenIn, tokenOut string, amountIn *big.Int, isDryRun bool) (*schema.SwapOutput, error) {
	if tokenIn == tokenOut {
		return nil, ERR_INVALID_TOKEN
	}
	if tokenIn != lp.TokenXTag && tokenIn != lp.TokenYTag {
		return nil, ERR_INVALID_TOKEN
	}
	if tokenOut != lp.TokenXTag && tokenOut != lp.TokenYTag {
		return nil, ERR_INVALID_TOKEN
	}

	zero := big.NewInt(0)
	if amountIn.Cmp(zero) != 1 {
		return nil, ERR_INVALID_AMOUNT
	}

	fee, err := getAndCheckFee(amountIn, lp.FeeRatio)
	if err != nil {
		return nil, err
	}

	amount := new(big.Int)
	amount.Sub(amountIn, fee)

	var amountOut *big.Int
	var endSqrtPrice *apd.Decimal
	if tokenIn == lp.TokenXTag {
		//price will down
		if lp.PriceDirection == schema.PriceDirectionUp {
			return nil, ERR_INVALID_TOKEN
		}
		amountOut, endSqrtPrice, err = SwapOut(lp.CurrentSqrtPrice, lp.Liquidity, amount, true)
		if err != nil {
			return nil, err
		}
		if endSqrtPrice.Cmp(lp.LowSqrtPrice) == -1 {
			return nil, ERR_OUT_OF_RANGE
		}
	} else {
		//price will up
		if lp.PriceDirection == schema.PriceDirectionDown {
			return nil, ERR_INVALID_TOKEN
		}

		amountOut, endSqrtPrice, err = SwapOut(lp.CurrentSqrtPrice, lp.Liquidity, amount, false)
		if err != nil {
			return nil, err
		}
		if endSqrtPrice.Cmp(lp.HighSqrtPrice) == 1 {
			return nil, ERR_OUT_OF_RANGE
		}
	}

	result := schema.SwapOutput{
		LpID:           lp.ID(),
		TokenIn:        tokenIn,
		AmountIn:       amountIn,
		TokenOut:       tokenOut,
		AmountOut:      amountOut,
		Fee:            fee,
		StartSqrtPrice: lp.CurrentSqrtPrice,
		EndSqrtPrice:   endSqrtPrice,
		IsDryRun:       isDryRun,
	}

	if !isDryRun {
		lp.CurrentSqrtPrice = endSqrtPrice
	}

	return &result, nil
}

func LpVerifySwap(lp *schema.Lp, tokenIn, tokenOut string, amountIn, amountOut *big.Int) (bool, error) {
	r, err := LpSwap(lp, tokenIn, tokenOut, amountIn, false)
	if err != nil {
		return false, err
	}
	log.Debug("VerifySwap:", "amountOut_by_swap", r.AmountOut, "amountOut_in_params", amountOut)
	if r.AmountOut.Cmp(amountOut) != -1 {
		return true, nil
	}
	return false, nil
}

func GetLpID(poolID, address string, lowSqrtPrice, highSqrtPrice *apd.Decimal, priceDirection string) string {
	s := "PoolID:" + poolID + "\n" +
		"Address:" + address + "\n" +
		"LowSqrtPrice:" + lowSqrtPrice.Text('f') + "\n" + //https://pkg.go.dev/github.com/cockroachdb/apd#Decimal.Text
		"HighSqrtPrice:" + highSqrtPrice.Text('f') + "\n" +
		"PriceDirection:" + priceDirection

	h := accounts.TextHash([]byte(s))
	return hexutil.Encode(h)
}
