package agent

import (
	"encoding/json"
	"errors"
	"github.com/cockroachdb/apd/v3"
	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	ffpSchema "github.com/permadao/ffp/schema"
	"github.com/permadao/ffp/utils"
	goarSchema "github.com/permadao/goar/schema"
	"github.com/permadao/permaswap/agent/schema"
	"github.com/permadao/permaswap/core"
	psSchema "github.com/permadao/permaswap/core/schema"
	routerSchema "github.com/permadao/permaswap/router/schema"
	"math/big"
)

type PoolAgent struct {
	pid            string
	owner          string
	settlePid      string
	psCore         *core.Core
	notes          map[string]*schema.AgentNote // noteID -> note
	reservedAmount map[string]*big.Int          // swap reserved amount
}

func newPoolAgent(pid, owner, settlePid string) *PoolAgent {
	return &PoolAgent{
		pid:       pid,
		owner:     owner,
		settlePid: settlePid,
		psCore: &core.Core{
			FeeRecepient:      "",
			FeeRatio:          new(apd.Decimal).SetInt64(0),
			Pools:             make(map[string]*psSchema.Pool),
			Lps:               make(map[string]*psSchema.Lp),
			AddressToLpIDs:    make(map[string][]string),
			MaxPoolPathLength: 3,
			TokenTagToPoolIDs: make(map[string][]string),
		},
		notes:          make(map[string]*schema.AgentNote),
		reservedAmount: map[string]*big.Int{},
	}
}

func (a *PoolAgent) Apply(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	switch meta.Action {
	case schema.AgentRequestPrice: // handle user agent submit action
		res = a.handleRequestPrice(from, meta)

	case ffpSchema.FFPExecute: // handle settle execute request
		res = a.handleSettleExecute(from, meta)

	case "Credit-Notice":
		switch meta.Params[ffpSchema.XFFPFor] {
		case ffpSchema.FFPSuccess, ffpSchema.FFPRefund:
			res = a.handleSettleNotice(from, meta)
		default:
			// deposit balances
			log.Debug("deposit balances success...")
		}
	case "AddPool":
		res = a.handleAddPool(from, meta)

	case "AddPoolLiquidity":
		res = a.handleAddPoolLiquidity(from, meta)

	case "RemovePoolLiquidity":
		res = a.handleRemovePoolLiquidity(from, meta)

	case "GetLpsBalances": // todo changed api
		res = a.handleGetLpsBalances(from)

	case "SwapQuery": // todo changed api
		res = a.handleSwapQuery(from, meta)
	}
	return
}

func (a *PoolAgent) handleRequestPrice(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		err  error
		msgs = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "RequestPrice-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	tokenIn := meta.Params["TokenIn"]
	tokenOut := meta.Params["TokenOut"]
	amountIn := meta.Params["AmountIn"]

	amountInNum, ok := new(big.Int).SetString(amountIn, 10)
	if !ok {
		err = errors.New("amountIn incorrect")
		return
	}
	_, poolAmountOut, err := a.swapOutputs(from, tokenIn, tokenOut, amountInNum)
	if err != nil {
		err = errors.New("not found amm paths")
		return
	}
	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: schema.AgentPriceNotice},
			{Name: "X-OrderId", Value: meta.Params["X-OrderId"]},
			{Name: "AmountOut", Value: poolAmountOut.String()},
		},
	})
	return
}

func (a *PoolAgent) handleSettleExecute(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		err    error
		msgs   = make([]*vmmSchema.ResMessage, 0)
		noteId = ""
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Cancel"}, // send Cancel to settle vm
					{Name: "NoteID", Value: noteId},   // send Cancel to settle noteId
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	if from != a.settlePid {
		err = errors.New("from must be settle")
		return
	}

	input, note, side, err := a.parseExecuteParams(meta.Params)
	if err != nil {
		return
	}
	noteId = note.NoteID

	// try swap
	swapAmountOut, err := a.poolSwap(input)
	if err != nil {
		return
	}

	// send holdAmount to settle
	msgs = append(msgs, requestSwapSettle(input, from, noteId))

	// store note
	a.storeNote(swapAmountOut, input, side, note)
	return
}

func (a *PoolAgent) handleSettleNotice(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		noteId string
		err    error
		msgs   = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "Finish-Settle-Error"},
					{Name: "Error", Value: err.Error()},
					{Name: "NoteID", Value: noteId},
				},
			})
			res.Error = err
		}
	}()
	note, ffpFor, quantity, err := a.parseSettleNoticeParams(meta.Params)
	if err != nil {
		return
	}
	noteId = note.NoteID
	err = executeSettleNotice(from, ffpFor, quantity, note)
	return
}

func (a *PoolAgent) handleAddPool(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		err  error
		msgs = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "FFP-Agent-Add-Pool-Failed"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	if from != a.owner {
		err = errors.New("from_incorrect")
		return
	}

	tokenX := meta.Params["TokenX"]
	tokenY := meta.Params["TokenY"]
	feeRatio := meta.Params["FeeRatio"]

	if tokenX >= tokenY {
		tmp := tokenX
		tokenX = tokenY
		tokenY = tmp
	}

	pool, err := core.NewPool(tokenX, tokenY, feeRatio)
	if err != nil {
		log.Error("core.NewPool(tokenX, tokenY, feeRatio) failed", "err", err)
		return
	}

	if _, ok := a.psCore.Pools[pool.ID()]; ok {
		err = errors.New("exist_pool")
		return
	}
	a.psCore.Pools[pool.ID()] = pool

	a.psCore.TokenTagToPoolIDs[pool.TokenXTag] = append(a.psCore.TokenTagToPoolIDs[pool.TokenXTag], pool.ID())
	a.psCore.TokenTagToPoolIDs[pool.TokenYTag] = append(a.psCore.TokenTagToPoolIDs[pool.TokenYTag], pool.ID())

	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: a.pid,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "FFP-Agent-Add-Pool-Success"},
			{Name: "PoolID", Value: pool.ID()},
			{Name: "PoolName", Value: pool.String()},
		},
	})
	return
}

func (a *PoolAgent) handleAddPoolLiquidity(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		err  error
		msgs = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "FFP-Agent-Add-Pool-Liquidity-Failed"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	if from != a.owner {
		err = errors.New("from_incorrect")
		return
	}

	tokenX := meta.Params["TokenX"]
	tokenY := meta.Params["TokenY"]
	feeRatio := meta.Params["FeeRatio"]

	lowSqrtPrice := meta.Params["LowSqrtPrice"]
	currentSqrtPrice := meta.Params["CurrentSqrtPrice"]
	highSqrtPrice := meta.Params["HighSqrtPrice"]
	liquidity := meta.Params["Liquidity"]
	priceDirection := meta.Params["PriceDirection"]

	if tokenX >= tokenY {
		tmp := tokenX
		tokenX = tokenY
		tokenY = tmp
	}

	feeRatioNum, _, err := new(apd.Decimal).SetString(feeRatio)
	if err != nil {
		return
	}
	lowSqrtPriceNum, _, err := new(apd.Decimal).SetString(lowSqrtPrice)
	if err != nil {
		return
	}

	currentSqrtPriceNum, _, err := new(apd.Decimal).SetString(currentSqrtPrice)
	if err != nil {
		return
	}
	highSqrtPriceNum, _, err := new(apd.Decimal).SetString(highSqrtPrice)
	if err != nil {
		return
	}

	// calc liquidity to amountx amounty
	amountX, amountY, err := core.LiquidityToAmount(liquidity, lowSqrtPriceNum, currentSqrtPriceNum, highSqrtPriceNum, priceDirection)
	if err != nil {
		return
	}

	// add liquidity
	err = a.psCore.AddLiquidity(a.pid, routerSchema.LpMsgAdd{
		ID:               "",
		Event:            "",
		TokenX:           tokenX,
		TokenY:           tokenY,
		FeeRatio:         feeRatioNum,
		CurrentSqrtPrice: currentSqrtPriceNum,
		LowSqrtPrice:     lowSqrtPriceNum,
		HighSqrtPrice:    highSqrtPriceNum,
		Liquidity:        liquidity,
		PriceDirection:   priceDirection,
	})
	if err != nil {
		return
	}

	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: a.pid,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "FFP-Agent-Add-Pool-Liquidity-Success"},
			{Name: "AmountX", Value: amountX},
			{Name: "AmountY", Value: amountY},
		},
	})
	return
}

func (a *PoolAgent) handleRemovePoolLiquidity(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		err  error
		msgs = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "FFP-Agent-Add-Pool-Liquidity-Failed"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	if from != a.owner {
		err = errors.New("from_incorrect")
		return
	}

	tokenX := meta.Params["TokenX"]
	tokenY := meta.Params["TokenY"]
	feeRatio := meta.Params["FeeRatio"]

	lowSqrtPrice := meta.Params["LowSqrtPrice"]
	highSqrtPrice := meta.Params["HighSqrtPrice"]
	priceDirection := meta.Params["PriceDirection"]

	if tokenX >= tokenY {
		tmp := tokenX
		tokenX = tokenY
		tokenY = tmp
	}

	feeRatioNum, _, err := new(apd.Decimal).SetString(feeRatio)
	if err != nil {
		return
	}
	lowSqrtPriceNum, _, err := new(apd.Decimal).SetString(lowSqrtPrice)
	if err != nil {
		return
	}

	highSqrtPriceNum, _, err := new(apd.Decimal).SetString(highSqrtPrice)
	if err != nil {
		return
	}

	err = a.psCore.RemoveLiquidity(a.pid, routerSchema.LpMsgRemove{
		ID:             "",
		Event:          "",
		TokenX:         tokenX,
		TokenY:         tokenY,
		FeeRatio:       feeRatioNum,
		LowSqrtPrice:   lowSqrtPriceNum,
		HighSqrtPrice:  highSqrtPriceNum,
		PriceDirection: priceDirection,
	})
	if err != nil {
		return
	}

	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: a.pid,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "FFP-Agent-Remove-Pool-Liquidity-Success"},
		},
	})
	return
}

func (a *PoolAgent) handleGetLpsBalances(from string) (res vmmSchema.Result) {
	var (
		err  error
		msgs = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "FFP-Agent-GetLpsBalances-Failed"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	balances := map[string]*big.Int{}
	for _, lp := range a.psCore.Lps {
		var amountX, amountY string
		amountX, amountY, err = core.LiquidityToAmount(lp.Liquidity.String(), lp.LowSqrtPrice, lp.CurrentSqrtPrice, lp.HighSqrtPrice, lp.PriceDirection)
		if err != nil {
			return
		}
		x, ok := new(big.Int).SetString(amountX, 10)
		if !ok {
			err = errors.New("err_invalid_amount")
			return
		}
		y, ok := new(big.Int).SetString(amountY, 10)
		if !ok {
			err = errors.New("err_invalid_amount")
			return
		}
		if b, ok := balances[lp.TokenXTag]; ok {
			balances[lp.TokenXTag] = new(big.Int).Add(b, x)
		} else {
			balances[lp.TokenXTag] = x
		}

		if b, ok := balances[lp.TokenYTag]; ok {
			balances[lp.TokenYTag] = new(big.Int).Add(b, y)
		} else {
			balances[lp.TokenYTag] = y
		}
	}

	balancesJs, _ := json.Marshal(balances)
	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: a.pid,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "FFP-Agent-GetLpsBalances-Success"},
		},
		Data: string(balancesJs),
	})
	return
}

func (a *PoolAgent) handleSwapQuery(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	var (
		err  error
		msgs = make([]*vmmSchema.ResMessage, 0)
	)
	defer func() {
		res = vmmSchema.Result{Messages: msgs}
		if err != nil {
			res.Messages = append(res.Messages, &vmmSchema.ResMessage{
				Target: from,
				Tags: []goarSchema.Tag{
					{Name: "Action", Value: "FFP-Agent-Swap-Query-Failed"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()
	address := from
	tokenIn := meta.Params["TokenIn"]
	tokenOut := meta.Params["TokenOut"]
	amountIn := meta.Params["AmountIn"]

	paths, err := a.psCore.Query(routerSchema.UserMsgQuery{
		Address:  address,
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		AmountIn: amountIn,
	})
	if err != nil {
		return
	}

	pathsJs, err := json.Marshal(paths)
	if err != nil {
		return
	}
	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "FFP-Agent-Swap-Query-Success"},
			{Name: "Paths", Value: string(pathsJs)},
			{Name: "TokenIn", Value: tokenIn},
			{Name: "TokenOut", Value: tokenOut},
			{Name: "AmountIn", Value: amountIn},
			{Name: "Address", Value: address},
		},
	})
	return
}

func (a *PoolAgent) swapOutputs(sender, tokenIn, tokenOut string, amountIn *big.Int) ([]psSchema.SwapOutput, *big.Int, error) {
	poolPaths, err := a.psCore.FindPoolPaths(tokenIn, tokenOut)
	if err != nil {
		return nil, nil, err
	}
	var (
		amountOut = big.NewInt(0)
		sos       = make([]psSchema.SwapOutput, 0)
		errs      = make([]error, 0)
	)
	for _, poolPath := range poolPaths {
		sos_, amountOut_, err := core.PoolsSwap(poolPath, tokenIn, tokenOut, amountIn, a.psCore.AddressToLpIDs[sender])
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if amountOut_.Cmp(amountOut) == 1 { // get great price
			amountOut = amountOut_
			sos = sos_
		}
	}
	// error
	if len(sos) == 0 {
		isInvalidAmount := true
		for _, e := range errs {
			if e.Error() != core.ERR_INVALID_AMOUNT.Error() {
				isInvalidAmount = false
				break
			}
		}
		if isInvalidAmount {
			return nil, nil, core.ERR_INVALID_AMOUNT
		}

		return nil, nil, core.ERR_NO_PATH
	}

	return sos, amountOut, nil
}

func (a *PoolAgent) parseSettleNoticeParams(params map[string]string) (note *schema.AgentNote, ffpFor string, quantity *big.Int, err error) {
	noteId := params[ffpSchema.XFFPNoteID]
	note, ok := a.notes[noteId]
	if !ok {
		err = errors.New("not found note")
		return
	}
	sender := params["Sender"]
	if sender != a.settlePid {
		err = errors.New("sender must be settle pid")
		return
	}
	quantity, ok = utils.ParseAmount(params["Quantity"])
	if !ok {
		err = errors.New("quantity incorrect")
		return
	}
	ffpFor, ok = params[ffpSchema.XFFPFor]
	if !ok {
		err = errors.New("not found ffpFor")
		return
	}
	return
}

func executeSettleNotice(from, ffpFor string, quantity *big.Int, note *schema.AgentNote) (err error) {
	var (
		tokenId    string
		amount     *big.Int
		noteStatus string
	)
	switch ffpFor {
	case ffpSchema.FFPSuccess:
		switch note.Side {
		case ffpSchema.SideHolder:
			tokenId = note.AssetID
			amount = note.Amount
		case ffpSchema.SideIssuer:
			tokenId = note.HolderAssetID
			amount = note.HolderAmount
		default:
			err = errors.New("not support side")
			return
		}
		noteStatus = schema.PoolAgentNoteStatusSettled
	case ffpSchema.FFPRefund:
		switch note.Side {
		case ffpSchema.SideHolder:
			tokenId = note.HolderAssetID
			amount = note.HolderAmount
		case ffpSchema.SideIssuer:
			tokenId = note.AssetID
			amount = note.Amount
		default:
			err = errors.New("not support side")
			return
		}
		noteStatus = schema.PoolAgentNoteStatusRefund
	default:
		err = errors.New("not support X-FFP-For")
		return
	}

	if tokenId != from {
		err = errors.New("from must be note assetID")
		return
	}

	if amount.Cmp(quantity) != 0 {
		err = errors.New("quantity not equal note amount")
		return
	}
	note.Status = noteStatus
	return
}

func (a *PoolAgent) parseExecuteParams(params map[string]string) (input schema.SwapInputReq, note ffpSchema.Note, side string, err error) {
	noteStr := params[ffpSchema.XFFPNote]
	note, err = utils.DecodeNote(noteStr)
	if err != nil {
		err = errors.New("decode note failed")
		return
	}
	noteId, err := utils.GenerateNoteID(note)
	if err != nil {
		err = errors.New("generate noteId failed")
		return
	}
	note.NoteID = noteId
	// check note exist
	if _, ok := a.notes[noteId]; ok {
		err = errors.New("repeat execute note")
		return
	}

	executedSide := params[ffpSchema.FFPExecuted] // normal issuer
	side = params[ffpSchema.FFPSide]              // normal holder
	switch side {
	case ffpSchema.SideHolder:
		// executedSide must be issuer
		if executedSide != ffpSchema.SideIssuer {
			err = errors.New("executedSide incorrect")
			return
		}

		input = schema.SwapInputReq{
			Address:   note.Issuer,
			TokenIn:   note.AssetID,
			TokenOut:  note.HolderAssetID,
			AmountIn:  note.Amount,
			AmountOut: note.HolderAmount,
		}
	case ffpSchema.SideIssuer:
		// executedSide must be holder
		if executedSide != ffpSchema.SideHolder {
			err = errors.New("executedSide incorrect")
			return
		}
		input = schema.SwapInputReq{
			Address:   note.Holder,
			TokenIn:   note.HolderAssetID,
			TokenOut:  note.AssetID,
			AmountIn:  note.HolderAmount,
			AmountOut: note.Amount,
		}
	default:
		err = errors.New("FFPSide incorrect")
		return
	}
	return

}

func (a *PoolAgent) poolSwap(input schema.SwapInputReq) (swapAmountOut *big.Int, err error) {
	// verify note settle price
	sos, swapAmountOut, err := a.swapOutputs(input.Address, input.TokenIn, input.TokenOut, input.AmountIn)
	if err != nil {
		err = errors.New("not found amm paths")
		return
	}
	if swapAmountOut.Cmp(input.AmountOut) < 0 {
		err = errors.New("pool price insufficient")
		return
	}
	// pool swap
	paths, err := core.SwapOutputsToPaths(input.Address, a.psCore, sos)
	if err != nil {
		log.Error("core.SwapOutputsToPaths", "err", err)
		err = errors.New("get swap paths failed")
		return
	}

	if err = a.psCore.Update(input.Address, paths); err != nil {
		log.Error("a.psCore.Update", "err", err)
		err = errors.New("pool swap failed")
		return
	}
	return
}

func (a *PoolAgent) storeNote(swapAmountOut *big.Int, input schema.SwapInputReq, side string, note ffpSchema.Note) {
	// store reserved amount
	reservedAmountOut := new(big.Int).Sub(swapAmountOut, input.AmountOut)
	amt, ok := a.reservedAmount[input.TokenOut]
	if !ok {
		a.reservedAmount[input.TokenOut] = reservedAmountOut
	} else {
		a.reservedAmount[input.TokenOut] = new(big.Int).Add(amt, reservedAmountOut)
	}

	// store note
	a.notes[note.NoteID] = &schema.AgentNote{
		Note:   note,
		Side:   side,
		Status: schema.PoolAgentNoteStatusExecuting,
	}
}

func requestSwapSettle(input schema.SwapInputReq, settlePid, noteId string) *vmmSchema.ResMessage {
	return &vmmSchema.ResMessage{
		Target: input.TokenOut,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Transfer"},
			{Name: "Recipient", Value: settlePid},
			{Name: "Quantity", Value: input.AmountOut.String()},

			{Name: ffpSchema.XFFPFor, Value: ffpSchema.FFPExecute},
			{Name: ffpSchema.XFFPNoteID, Value: noteId},
		},
	}
}
