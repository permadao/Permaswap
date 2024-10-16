package proposal

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/permadao/permaswap/halo/hvm/schema"
)

const (
	POOL = "incentive"
)

var (
	ErrPropsalInvalidInitData    = errors.New("err_proposal_invalid_init_data")
	ErrPropsalInvalidLocalState  = errors.New("err_proposal_invalid_local_state")
	ErrPropsalInvalidTxAction    = errors.New("err_proposal_invalid_tx_action")
	ErrPropsalInvalidSwapOrder   = errors.New("err_proposal_invalid_swap_order")
	ErrPropsalMiningNotStart     = errors.New("err_proposal_mining_not_start")
	ErrPropsalMiningEnd          = errors.New("err_proposal_mining_end")
	ErrorInvalidTimeElapsed      = errors.New("err_invalid_time_elapsed")
	ErrPropsalInvalidTotalSupply = errors.New("err_proposal_invalid_total_supply")
)

type Liquidity struct {
	Name        string `json:"name"`
	Router      string `json:"router"` // router address
	Pool        string `json:"pool"`   // pool id
	BaseToken   string `json:"baseToken"`
	TotalSupply string `json:"totalSupply"` // total supply for this liquidity mining
	Start       int64  `json:"start"`
	End         int64  `json:"end"`

	LastMining int64             `json:"last_mining"`
	Mined      map[string]string `json:"mined"`
	TotalMined *big.Int          `json:"totalMined"`

	// debug
	TxMined map[string]map[string]*big.Int `json:"txMined"` // tx hash -> lp -> amount
}

func Execute(tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle, localState, initData string) (*schema.StateForProposal, string, string, error) {
	if tx.Action != schema.TxActionSwap {
		return state, localState, "", ErrPropsalInvalidTxAction
	}

	if tx.SwapOrder == nil || tx.SwapOrder.Err != "" {
		return state, localState, "", ErrPropsalInvalidSwapOrder
	}

	var liquidity Liquidity
	if localState == "" {
		if err := json.Unmarshal([]byte(initData), &liquidity); err != nil {
			return state, localState, "", ErrPropsalInvalidInitData
		}
		liquidity.LastMining = liquidity.Start
		liquidity.Mined = make(map[string]string)
		liquidity.TotalMined = big.NewInt(0)

		liquidity.TxMined = make(map[string]map[string]*big.Int)
	} else {
		if err := json.Unmarshal([]byte(localState), &liquidity); err != nil {
			return state, localState, "", ErrPropsalInvalidLocalState
		}
	}

	if liquidity.LastMining >= liquidity.End {
		return state, localState, "", ErrPropsalMiningEnd
	}
	now := tx.SwapOrder.TimeStamp
	if now < liquidity.Start {
		return state, localState, "", ErrPropsalMiningNotStart
	}
	if now > liquidity.End {
		now = liquidity.End
	}
	timeElapsed := now - liquidity.LastMining
	if timeElapsed <= 0 {
		return state, localState, "", ErrorInvalidTimeElapsed
	}
	totalSupply, ok := new(big.Int).SetString(liquidity.TotalSupply, 10)
	if !ok {
		return state, localState, "", ErrPropsalInvalidTotalSupply
	}

	lpToVolume := make(map[string]*big.Int)
	totalVolume := big.NewInt(0)
	for _, item := range tx.SwapOrder.Items {
		if item.PoolID != liquidity.Pool {
			continue
		}
		if _, ok := lpToVolume[item.Lp]; !ok {
			lpToVolume[item.Lp] = big.NewInt(0)
		}
		volume := item.AmountOut
		if item.TokenIn == liquidity.BaseToken {
			volume = item.AmountIn
		}
		lpToVolume[item.Lp].Add(lpToVolume[item.Lp], volume)
		totalVolume.Add(totalVolume, volume)
	}

	if totalVolume.Cmp(big.NewInt(0)) == 0 {
		return state, localState, "", nil
	}

	liquidity.LastMining = now
	lpToAmount := make(map[string]*big.Int)
	timeElapsed_ := big.NewInt(timeElapsed)
	totalTime := big.NewInt(liquidity.End - liquidity.Start)
	for lp, volume := range lpToVolume {
		// amount = timeElapsed * totalSupply * volume / totalVolume / totalTime
		lpToAmount[lp] = new(big.Int).Div(new(big.Int).Div(new(big.Int).Mul(volume, new(big.Int).Mul(timeElapsed_, totalSupply)), totalVolume), totalTime)
	}

	for _, amount := range lpToAmount {
		liquidity.TotalMined = new(big.Int).Add(liquidity.TotalMined, amount)
	}

	liquidity.TxMined[tx.EverHash] = lpToAmount

	for lp, amount := range lpToAmount {
		err := state.Token.Transfer(POOL, lp, amount, state.FeeRecipient, big.NewInt(0), false)
		if err != nil {
			continue
		}
		if _, ok := liquidity.Mined[lp]; !ok {
			liquidity.Mined[lp] = "0"
		}
		mined, _ := new(big.Int).SetString(liquidity.Mined[lp], 10)
		liquidity.Mined[lp] = new(big.Int).Add(mined, amount).String()
	}

	localStateNew, err := json.Marshal(liquidity)
	if err != nil {
		return state, localState, "", err
	}

	return state, string(localStateNew), "", nil
}
