package core

import (
	"math/big"

	"github.com/everVision/everpay-kits/utils"
	"github.com/permadao/permaswap/core/schema"
)

func SwapOutputsToPaths(user string, core *Core, swapOutputs []schema.SwapOutput) ([]schema.Path, error) {
	paths := []schema.Path{}
	for _, so := range swapOutputs {

		lp, ok := core.Lps[so.LpID]
		if !ok {
			return nil, ERR_NO_LP
		}

		pathIn := schema.Path{
			LpID:     so.LpID,
			From:     user,
			To:       lp.AccID,
			TokenTag: so.TokenIn,
			Amount:   so.AmountIn.String(),
		}
		paths = append(paths, pathIn)

		pathOut := schema.Path{
			LpID:     so.LpID,
			From:     lp.AccID,
			To:       user,
			TokenTag: so.TokenOut,
			Amount:   so.AmountOut.String(),
		}
		paths = append(paths, pathOut)
	}

	return paths, nil
}

func PathsToSwapInputs(user string, paths []schema.Path) (map[string]*schema.SwapInput, error) {
	_, accid, err := utils.IDCheck(user)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return nil, ERR_INVALID_PATH
	}
	// remove path fee
	if len(paths)%2 != 0 && paths[len(paths)-1].LpID == "" {
		paths = paths[:len(paths)-1]
	}
	if len(paths)%2 != 0 {
		return nil, ERR_INVALID_PATH
	}

	lpID2SwapInput := make(map[string]*schema.SwapInput)
	lpID2Counter := make(map[string]int)
	for _, path := range paths {

		token := path.TokenTag

		amount := new(big.Int)
		amount, ok := amount.SetString(path.Amount, 10)
		if !ok {
			return nil, ERR_INVALID_PATH
		}
		if amount.Cmp(big.NewInt(0)) != 1 {
			return nil, ERR_INVALID_PATH
		}

		lpID := path.LpID
		lpID2Counter[lpID] = lpID2Counter[lpID] + 1
		_, from, err := utils.IDCheck(path.From)
		if err != nil {
			return nil, err
		}
		_, to, err := utils.IDCheck(path.To)
		if err != nil {
			return nil, err
		}
		if from == accid {
			if si, ok := lpID2SwapInput[lpID]; ok {
				if si.TokenIn != "" {
					return nil, ERR_INVALID_PATH
				}
				si.TokenIn = token
				si.AmountIn = amount

			} else {
				lpID2SwapInput[lpID] = &schema.SwapInput{
					LpID:     lpID,
					TokenIn:  token,
					AmountIn: amount,
				}
			}
		} else {

			if to != accid {
				return nil, ERR_INVALID_PATH
			}

			if si, ok := lpID2SwapInput[lpID]; ok {
				if si.TokenOut != "" {
					return nil, ERR_INVALID_PATH
				}
				si.TokenOut = token
				si.AmountOut = amount
			} else {
				lpID2SwapInput[lpID] = &schema.SwapInput{
					LpID:      lpID,
					TokenOut:  token,
					AmountOut: amount,
				}
			}
		}
	}

	for _, n := range lpID2Counter {
		if n != 2 {
			return nil, ERR_INVALID_PATH
		}
	}

	for _, si := range lpID2SwapInput {
		if si.TokenIn == "" || si.TokenOut == "" {
			return nil, ERR_INVALID_PATH
		}
	}

	return lpID2SwapInput, nil
}
