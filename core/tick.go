package core

import (
	"math/big"
	"sort"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
)

type Tick struct {
	SqrtPrice *apd.Decimal
	Liquidity *big.Int // net liquidity
}

func ticksFromLps(lps []*schema.Lp, priceDirection string) ([]Tick, error) {

	if (priceDirection != schema.PriceDirectionUp) && (priceDirection != schema.PriceDirectionDown) {
		return nil, ERR_INVALID_PRICE_DIRECTION
	}

	priceToLiquidity := make(map[string]*big.Int)
	for i, lp := range lps {
		if priceDirection == schema.PriceDirectionUp && lp.CurrentSqrtPrice.Cmp(lp.HighSqrtPrice) != -1 {
			continue
		}
		if priceDirection == schema.PriceDirectionDown && lp.CurrentSqrtPrice.Cmp(lp.LowSqrtPrice) != 1 {
			continue
		}
		if priceDirection == schema.PriceDirectionUp && lp.PriceDirection == schema.PriceDirectionDown {
			continue
		}
		if priceDirection == schema.PriceDirectionDown && lp.PriceDirection == schema.PriceDirectionUp {
			continue
		}

		pl := lp.LowSqrtPrice.String()
		pc := lp.CurrentSqrtPrice.String()
		ph := lp.HighSqrtPrice.String()

		if l, ok := priceToLiquidity[pc]; ok {
			priceToLiquidity[pc].Add(l, lp.Liquidity)
		} else {
			priceToLiquidity[pc] = new(big.Int).Set(lp.Liquidity)
		}
		if priceDirection == schema.PriceDirectionUp {
			if l, ok := priceToLiquidity[ph]; ok {
				priceToLiquidity[ph].Sub(l, lp.Liquidity)
			} else {
				priceToLiquidity[ph] = big.NewInt(0)
				priceToLiquidity[ph].Neg(lp.Liquidity)
			}
		} else {
			if l, ok := priceToLiquidity[pl]; ok {
				priceToLiquidity[pl].Sub(l, lp.Liquidity)
			} else {
				priceToLiquidity[pl] = big.NewInt(0)
				priceToLiquidity[pl].Neg(lp.Liquidity)
			}
		}
		log.Debug("func ticksFromLps", "i:", i, "lpID:", lp.ID(), "liquidity:", lp.Liquidity, "priceToLiquidity:", priceToLiquidity)
	}

	if len(priceToLiquidity) == 0 {
		return nil, ERR_NO_LP
	}

	keys := DecimalSlice{}
	for k := range priceToLiquidity {
		p := new(apd.Decimal)
		p.SetString(k)
		keys = append(keys, p)
	}
	sort.Sort(keys)

	var ticks []Tick
	for _, k := range keys {
		t := Tick{k, priceToLiquidity[k.String()]}
		ticks = append(ticks, t)
	}

	// Reverse ticks if priceDirection is down
	if priceDirection == schema.PriceDirectionDown {
		for i, j := 0, len(ticks)-1; i < j; i, j = i+1, j-1 {
			ticks[i], ticks[j] = ticks[j], ticks[i]
		}
	}

	return ticks, nil
}
