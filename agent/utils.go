package agent

import (
	"errors"
	"github.com/permadao/ffp/utils"
	"github.com/permadao/permaswap/agent/schema"
	"math/big"
)

func ParseOrderReqParams(params map[string]string) (order schema.OrderReq, err error) {
	tokenIn, ok := params["TokenIn"]
	if !ok {
		err = errors.New("TokenIn is required")
		return
	}
	tokenOut, ok := params["TokenOut"]
	if !ok {
		err = errors.New("TokenOut is required")
		return
	}
	amountIn, ok := params["AmountIn"]
	if !ok {
		err = errors.New("AmountIn is required")
		return
	}
	amountInNum, ok := utils.ParseAmount(amountIn)
	if !ok {
		err = errors.New("amountIn incorrect")
		return
	}

	maxAmountOut, ok := params["AmountOut"]
	if !ok {
		err = errors.New("AmountOut is required")
		return
	}

	maxAmountOutNum, ok := utils.ParseAmount(maxAmountOut)
	if !ok {
		err = errors.New("amountOut incorrect")
		return
	}

	slippage, ok := params["Slippage"] // 5 = 0.5%, 10 = 1%
	if !ok {
		err = errors.New("Slippage is required")
		return
	}
	slippageNum, ok := utils.ParseAmount(slippage)
	if !ok {
		err = errors.New("slippage incorrect")
		return
	}

	poolAgent, ok := params["PoolAgent"]
	if !ok {
		err = errors.New("PoolAgent is required")
		return
	}

	return schema.OrderReq{
		TokenIn:      tokenIn,
		TokenOut:     tokenOut,
		AmountIn:     amountInNum,
		MaxAmountOut: maxAmountOutNum,
		Slippage:     slippageNum,
		PoolAgent:    poolAgent,
	}, nil

}

func calcSlippageAmountOut(maxAmountOut, slippage *big.Int) (*big.Int, error) {
	// minAmountOut = maxAmountOut * (1-slippage/1000) = maxAmountOut- maxAmountOut*slippage/1000
	slipNum := new(big.Int).Div(new(big.Int).Mul(maxAmountOut, slippage), big.NewInt(1000))
	minAmountOut := new(big.Int).Sub(maxAmountOut, slipNum)
	if minAmountOut.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("minAmountOut must be greater than 0")
	}
	return minAmountOut, nil
}

func BigIntSub(a, b string) (string, bool) {
	ai, ok := new(big.Int).SetString(a, 10)
	if !ok {
		return "", false
	}
	bi, ok := new(big.Int).SetString(b, 10)
	if !ok {
		return "", false
	}
	return ai.Sub(ai, bi).String(), true
}

func BigIntMul(a, b string) string {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return ai.Mul(ai, bi).String()
}

func BigIntDiv(a, b string) string {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	if bi.Sign() == 0 {
		return "0"
	}
	return ai.Div(ai, bi).String()
}

func BigIntLt(a, b string) bool {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return ai.Cmp(bi) < 0
}

func BigIntLte(a, b string) bool {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return ai.Cmp(bi) <= 0
}

func BigIntGt(a, b string) bool {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return ai.Cmp(bi) > 0
}

func BigIntGte(a, b string) bool {
	ai, _ := new(big.Int).SetString(a, 10)
	bi, _ := new(big.Int).SetString(b, 10)
	return ai.Cmp(bi) >= 0
}
