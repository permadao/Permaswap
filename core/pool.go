package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/permadao/permaswap/core/schema"
)

func NewPool(tokenXTag, tokenYTag, feeRatio string) (*schema.Pool, error) {
	feeRatio_ := new(apd.Decimal)
	_, _, err := feeRatio_.SetString(feeRatio)
	if err != nil {
		return nil, ERR_INVALID_FEE
	}

	if tokenXTag >= tokenYTag {
		return nil, ERR_INVALID_TOKEN
	}

	return &schema.Pool{
		TokenXTag: tokenXTag,
		TokenYTag: tokenYTag,
		FeeRatio:  feeRatio_,
		Lps:       make(map[string]*schema.Lp),
	}, nil
}

func PoolAddLiquidity(pool *schema.Pool, lp *schema.Lp) error {
	if pool.ID() != lp.PoolID {
		return ERR_INVALID_POOL
	}
	pool.Lps[lp.ID()] = lp
	return nil
}

func PoolRemoveLiquidity(pool *schema.Pool, lpID string) {
	delete(pool.Lps, lpID)
}

func GetPoolLps(pool *schema.Pool, excludedLpIDs []string) []*schema.Lp {

	lps := []*schema.Lp{}
	for _, lp := range pool.Lps {
		isExcluded := false
		for _, lpID := range excludedLpIDs {
			if lp.ID() == lpID {
				isExcluded = true
				break
			}
		}
		if !isExcluded {
			lps = append(lps, lp)
		}
	}

	return lps
}

func GetPoolLps2(pool *schema.Pool) []schema.Lp {
	lps := []schema.Lp{}
	for _, lp := range pool.Lps {
		lps = append(lps, *lp)
	}
	return lps
}

func GetPoolLp(pool *schema.Pool, lpID string) (*schema.Lp, bool) {
	lp, ok := pool.Lps[lpID]
	return lp, ok
}

func GetPoolTicks(pool *schema.Pool, priceDirection string, excludedLpIDs []string) ([]Tick, error) {
	lps := GetPoolLps(pool, excludedLpIDs)
	return ticksFromLps(lps, priceDirection)
}

func PoolSwap(pool *schema.Pool, tokenIn, tokenOut string, amountIn *big.Int, excludedLpIDs []string, isDryRun bool) ([]schema.SwapOutput, error) {

	zero := big.NewInt(0)
	if tokenIn == tokenOut {
		return nil, ERR_INVALID_TOKEN
	}
	if tokenIn != pool.TokenXTag && tokenIn != pool.TokenYTag {
		return nil, ERR_INVALID_TOKEN
	}
	if tokenOut != pool.TokenXTag && tokenOut != pool.TokenYTag {
		return nil, ERR_INVALID_TOKEN
	}

	if amountIn.Cmp(zero) != 1 {
		return nil, ERR_INVALID_AMOUNT
	}

	fee, err := getAndCheckFee(amountIn, pool.FeeRatio)
	if err != nil {
		return nil, err
	}

	amountIn2 := new(big.Int).Sub(amountIn, fee)

	priceDirection := schema.PriceDirectionDown
	tokenInIsX := true
	if tokenIn == pool.TokenYTag {
		priceDirection = schema.PriceDirectionUp
		tokenInIsX = false
	}
	log.Debug("func PoolSwap:", "fee:", fee, "amountIn2:", amountIn2,
		"priceDirection:", priceDirection, "tokenInIsX:", tokenInIsX)

	ticks, err := GetPoolTicks(pool, priceDirection, excludedLpIDs)
	if err != nil {
		return nil, err
	}
	log.Debug("func PoolSwap:", "ticks:", len(ticks), "ticks:", ticks)

	// Get the endSqrtPrice of pool
	amountOut := big.NewInt(0)
	endSqrtPrice := new(apd.Decimal)
	amountRemain := new(big.Int).Set(amountIn2)
	cl := big.NewInt(0)
	for i, tick := range ticks {
		if (i + 1) >= len(ticks) {
			return nil, ERR_OUT_OF_RANGE
		}
		ct := tick
		cp := ct.SqrtPrice

		cl.Add(cl, ct.Liquidity)

		if cl.Cmp(zero) == 0 {
			// This interval of price is blank, no lp.
			log.Debug("func PoolSwap: skip blank tick", "tick i:", i, "cl", cl)
			continue
		}

		nt := ticks[i+1]
		np := nt.SqrtPrice

		log.Debug("func PoolSwap:", "tick i:", i, "cp:", cp, "np:", np, "cl:", cl)

		ai, ao, _, _, err := SwapAmount(cp, np, cl)
		if err != nil {
			return nil, err
		}

		log.Debug("func PoolSwap:", "ai:", ai, "ao:", ao, "amountRemain:", amountRemain)

		if amountRemain.Cmp(ai) != 1 {

			o, p, err := SwapOut(cp, cl, amountRemain, tokenInIsX)
			if err != nil {
				return nil, err
			}

			log.Debug("func PoolSwap(the last tick):", "cp:", cp, "cl:", cl,
				"amountRemain:", amountRemain, "o:", o, "p:", p)

			endSqrtPrice.Set(p)
			amountOut.Add(amountOut, o)
			break
		} else {
			amountOut.Add(amountOut, ao)
			amountRemain.Sub(amountRemain, ai)
		}
	}

	log.Debug("func PoolSwap:", "endSqrtPrice:", endSqrtPrice)

	if amountOut.Cmp(zero) != 1 {
		return nil, ERR_INVALID_AMOUNT
	}

	// Get the amountIn for every related lp of this pool
	lpIDToAmountIn := make(map[string]*big.Int)
	totalAmountIn := big.NewInt(0)
	for i, lp := range GetPoolLps(pool, excludedLpIDs) {
		lpID := lp.ID()
		log.Debug("func PoolSwap (amountIn for every lp):", "lp i", i, "lpID", lpID)

		if lp.PriceDirection != schema.PriceDirectionBoth && priceDirection != lp.PriceDirection {
			continue
		}

		lpEndSqrtPrice := new(apd.Decimal).Set(endSqrtPrice)
		if priceDirection == schema.PriceDirectionUp {
			if lp.CurrentSqrtPrice.Cmp(lp.HighSqrtPrice) != -1 {
				continue
			}
			if endSqrtPrice.Cmp(lp.CurrentSqrtPrice) != 1 {
				continue
			}
			if endSqrtPrice.Cmp(lp.HighSqrtPrice) == 1 {
				lpEndSqrtPrice.Set(lp.HighSqrtPrice)
			}
		} else {
			if lp.CurrentSqrtPrice.Cmp(lp.LowSqrtPrice) != 1 {
				continue
			}
			if endSqrtPrice.Cmp(lp.CurrentSqrtPrice) != -1 {
				continue
			}
			if endSqrtPrice.Cmp(lp.LowSqrtPrice) == -1 {
				lpEndSqrtPrice.Set(lp.LowSqrtPrice)
			}
		}

		_, _, amountInDecimal, _, err := SwapAmount(lp.CurrentSqrtPrice, lpEndSqrtPrice, lp.Liquidity)
		if err != nil {
			return nil, err
		}

		one := apd.New(1, 0)
		divisor := new(apd.Decimal)
		_, err = roundDownContext.Sub(divisor, one, pool.FeeRatio)
		if err != nil {
			return nil, err
		}
		_, err = roundUpContext.Quo(amountInDecimal, amountInDecimal, divisor)
		if err != nil {
			return nil, err
		}

		// amountIn is round down. otherwise it will out of range when end_price is high_price/low_price
		amountIn_, err := DecimalToBigDotInt(amountInDecimal, false)
		if err != nil {
			return nil, err
		}
		if amountIn_.Cmp(zero) != 1 {
			continue
		}
		//make sure fee > 1 for ever lp
		_, err = getAndCheckFee(amountIn_, pool.FeeRatio)
		if err != nil {
			continue
		}

		log.Debug("func PoolSwap (amountIn for every lp):", "lpID:", lpID, "amountIn_:", amountIn_)

		lpIDToAmountIn[lpID] = amountIn_
		totalAmountIn.Add(totalAmountIn, amountIn_)
	}

	if len(lpIDToAmountIn) == 0 {
		return nil, ERR_NO_PATH
	}

	//Fix when totalAmountIn != amountIn
	difference := new(big.Int).Sub(totalAmountIn, amountIn)
	log.Debug("func PoolSwap:", "totalAmountIn:", totalAmountIn, "amountIn:", amountIn, "difference:", difference)
	if difference.Cmp(zero) != 0 {
		for lpID, ai_ := range lpIDToAmountIn {
			newAmountIn := new(big.Int).Sub(ai_, difference)
			lp, _ := GetPoolLp(pool, lpID)
			_, err := LpSwap(lp, tokenIn, tokenOut, newAmountIn, true)
			if err != nil {
				log.Debug("func PoolSwap (fix difference) err, try next lp.", "lpID", lpID, "err", err)
			} else {
				lpIDToAmountIn[lpID] = newAmountIn
				log.Debug("func PoolSwap (fixed difference):", "lpID:", lpID, "newAmountIn", newAmountIn)
				break
			}
		}
	}

	totalAmountIn = big.NewInt(0)
	swapOutputs := []schema.SwapOutput{}
	for i, ai := range lpIDToAmountIn {
		lp, _ := GetPoolLp(pool, i)
		so, err := LpSwap(lp, tokenIn, tokenOut, ai, isDryRun)
		if err != nil {
			return nil, err
		}
		swapOutputs = append(swapOutputs, *so)
		totalAmountIn.Add(totalAmountIn, ai)
	}

	//check again
	difference = new(big.Int).Sub(totalAmountIn, amountIn)
	log.Debug("func PoolSwap:", "totalAmountIn again:", totalAmountIn, "amountIn:", amountIn, "difference:", difference)
	if difference.Cmp(zero) != 0 {
		log.Warn("func PoolSwap:Failed to fix difference. Make an offer anyway.")
	}

	return swapOutputs, nil
}

func GetPoolID(tokenXTag, tokenYTag string, feeRatio *apd.Decimal) string {
	s := "TokenXTag:" + tokenXTag + "\n" +
		"TokenYTag:" + tokenYTag + "\n" +
		"FeeRatio:" + feeRatio.Text('f')
	h := accounts.TextHash([]byte(s))
	return hexutil.Encode(h)
}

// PoolsSwap is not for update, use it for query
func PoolsSwap(poolPaths []*schema.Pool, tokenIn, tokenOut string, amountIn *big.Int, excludedLpIDs []string) ([]schema.SwapOutput, *big.Int, error) {
	//log.Info("PoolsSwap", "excludedLpIDs", excludedLpIDs)
	if poolPaths == nil || len(poolPaths) == 0 {
		return nil, nil, ERR_INVALID_POOL_PATHS
	}

	swapOutputs := []schema.SwapOutput{}

	tokenInTmp := tokenIn
	amountInTmp := amountIn

	sso := &schema.SwapOutput{}

	for _, pool := range poolPaths {

		tokenOutTmp := pool.TokenYTag
		if tokenInTmp == pool.TokenYTag {
			tokenOutTmp = pool.TokenXTag
		}

		sos, err := PoolSwap(pool, tokenInTmp, tokenOutTmp, amountInTmp, excludedLpIDs, true)
		if err != nil {
			return nil, nil, err
		}

		sso, err = summarySwapOuts(sos)
		if err != nil {
			return nil, nil, err
		}

		for _, so := range sos {
			swapOutputs = append(swapOutputs, so)
		}

		tokenInTmp = sso.TokenOut
		amountInTmp = sso.AmountOut
	}

	if sso.TokenOut != tokenOut {
		return nil, nil, ERR_INVALID_POOL_PATHS
	}

	return swapOutputs, sso.AmountOut, nil
}

func GetPoolCurrentPrice(pool *schema.Pool, priceDirection string) (string, error) {
	ticks, err := GetPoolTicks(pool, priceDirection, []string{})
	if err != nil {
		return "", err
	}
	return SqrtPriceToPriceWithFee(*ticks[0].SqrtPrice, *pool.FeeRatio, priceDirection)
}

func GetPoolCurrentPrice2(pool *schema.Pool, priceDirection string) (string, error) {

	tokenIn := pool.TokenXTag
	tokenOut := pool.TokenYTag
	if priceDirection == schema.PriceDirectionUp {
		tokenIn = pool.TokenYTag
		tokenOut = pool.TokenXTag
	}
	amountIn_, ok := schema.MinAmountInsForPriceQuery[tokenIn]
	if !ok {
		return "", ERR_INVALID_TOKEN
	}
	amountIn, _ := new(big.Int).SetString(amountIn_, 10)

	sos, err := PoolSwap(pool, tokenIn, tokenOut, amountIn, nil, true)
	if err != nil {
		return "", err
	}

	amountOut := big.NewInt(0)
	for _, so := range sos {
		amountOut.Add(amountOut, so.AmountOut)
	}
	//fmt.Println(priceDirection, sos, amountIn, amountOut)
	price := new(big.Float).Quo(new(big.Float).SetInt(amountOut), new(big.Float).SetInt(amountIn))
	return price.Text('f', 32), nil
}
