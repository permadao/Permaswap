package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/everVision/everpay-kits/common"
	"github.com/permadao/ffp/utils"
	"math/big"
	"time"

	vmmSchema "github.com/hymatrix/hymx/vmm/schema"
	ffpSchema "github.com/permadao/ffp/schema"
	goarSchema "github.com/permadao/goar/schema"
	"github.com/permadao/permaswap/agent/schema"
)

var log = common.NewLog("agent")

type UserAgent struct {
	pid             string
	owner           string
	settlePid       string
	limitOrders     map[string]*schema.LimitOrder // orderID -> LimitOrder
	noteIDToOrderID map[string]string             // poolNoteID -> orderID

	balances          map[string]*big.Int // tokenId -> balance // todo
	predictedBalances map[string]*big.Int // tokenId -> predicted balance // todo
}

func newUserAgent(pid, owner, settlePid string) *UserAgent {
	return &UserAgent{
		pid:             pid,
		owner:           owner,
		settlePid:       settlePid,
		limitOrders:     make(map[string]*schema.LimitOrder),
		noteIDToOrderID: make(map[string]string),
	}
}

func (a *UserAgent) Checkpoint() (data string, err error) {
	return "", nil
}

func (a *UserAgent) Restore(data string) error {
	return nil
}

func (a *UserAgent) Close() error {
	return nil
}

func (a *UserAgent) Apply(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
	switch meta.Action {
	case "LimitOrder": // user add limit order
		res = a.handleLimitOrder(from, meta)

	case "CancelLimitOrder": // user cancel limit order
		res = a.handleCancelLimitOrder(from, meta)

	case schema.AgentPriceNotice: // pool-agent sent price notice
		res = a.handlePoolPriceNotice(from, meta)

	case "Credit-Notice":
		switch meta.Params[ffpSchema.XFFPFor] {
		case ffpSchema.FFPSuccess, ffpSchema.FFPRefund:
			res = a.handleSettleNotice(from, meta)
		default:
			// deposit balances
			log.Debug("deposit balances success...")
		}

	case "Withdraw": // withdrawal agent balances
		res = a.handleWithdraw(from, meta)
	}
	return
}

func (a *UserAgent) handleLimitOrder(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
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
					{Name: "Action", Value: "LimitOrder-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	if from != a.owner {
		err = fmt.Errorf("only owner can create limit order")
		return
	}

	orderReq, err := ParseOrderReqParams(meta.Params)
	if err != nil {
		return
	}

	minAmountOut, err := calcSlippageAmountOut(orderReq.MaxAmountOut, orderReq.Slippage)
	if err != nil {
		return
	}
	limitOrder := &schema.LimitOrder{
		OrderID:      meta.ItemId,
		User:         from,
		TokenIn:      orderReq.TokenIn,
		TokenOut:     orderReq.TokenOut,
		AmountIn:     orderReq.AmountIn,
		MinAmountOut: minAmountOut,
		MaxAmountOut: orderReq.MaxAmountOut,
		Slippage:     orderReq.Slippage,
		PoolAgent:    orderReq.PoolAgent,
		Status:       schema.LimitOrderStatusPending,
		CreatedAt:    meta.Timestamp,
	}
	a.limitOrders[limitOrder.OrderID] = limitOrder

	// request pool price
	msgs = append(msgs, requestPoolPrice(limitOrder))

	orderJSON, _ := json.Marshal(limitOrder)
	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "LimitOrder-Notice"},
			{Name: "OrderID", Value: limitOrder.OrderID},
		},
		Data: string(orderJSON),
	})
	return
}

func (a *UserAgent) handlePoolPriceNotice(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
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
					{Name: "Action", Value: "PoolPriceNotice-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	order, poolAmountOutNum, err := a.parsePoolPriceNoticeParams(from, meta.Params)
	if err != nil {
		return
	}

	swapAmountOut, err := getSwapAmountOut(poolAmountOutNum, order)
	if err != nil {
		order.Status = schema.LimitOrderStatusFailedPrice
		return
	}

	note, err := a.assembleSwapNote(order, swapAmountOut)
	if err != nil {
		return
	}

	// send token transfer to settle vm for submit note settle
	msgs = append(msgs, requestSettleSubmitNote(a.settlePid, note))
	// set order status executing
	order.Status = schema.LimitOrderStatusExecuting
	return
}

func (a *UserAgent) handleCancelLimitOrder(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
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
					{Name: "Action", Value: "CancelLimitOrder-Error"},
					{Name: "Error", Value: err.Error()},
				},
			})
			res.Error = err
		}
	}()

	if from != a.owner {
		err = errors.New("only owner can cancel limit order")
		return
	}

	orderID, ok := meta.Params["OrderID"]
	if !ok {
		err = errors.New("OrderID is required")
		return
	}

	order, exists := a.limitOrders[orderID]
	if !exists {
		err = errors.New("order not found")
		return
	}

	// only handle pending order
	if order.Status != schema.LimitOrderStatusPending {
		err = errors.New("only pending orders can be cancelled")
		return
	}
	order.Status = schema.LimitOrderStatusCancelled

	orderJSON, _ := json.Marshal(order)
	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: from,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "CancelLimitOrder-Notice"},
			{Name: "OrderID", Value: orderID},
		},
		Data: string(orderJSON),
	})

	return
}

func (a *UserAgent) handleSettleNotice(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
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

	order, ffpFor, quantity, err := a.parseSettleNoticeParams(meta.Params)
	if err != nil {
		return
	}
	noteId = order.NoteId

	msgs, err = a.processSettleNotice(order, ffpFor, quantity, from)
	return
}

func (a *UserAgent) handleWithdraw(from string, meta vmmSchema.Meta) (res vmmSchema.Result) {
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
					{Name: "Action", Value: "User-Agent-Withdraw-Failed"},
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
	tokenId := meta.Params["TokenId"]
	recipient := meta.Params["Recipient"]
	quantity := meta.Params["Quantity"]

	msgs = append(msgs, &vmmSchema.ResMessage{
		Target: tokenId,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Transfer"},
			{Name: "Quantity", Value: quantity},
			{Name: "Recipient", Value: recipient},
			{Name: "X-FFP-User-Agent-For", Value: "Withdraw"},
		},
	})
	return
}

func (a *UserAgent) parseSettleNoticeParams(params map[string]string) (order *schema.LimitOrder, ffpFor string, quantity *big.Int, err error) {
	sender := params["Sender"]
	if sender != a.settlePid {
		err = errors.New("sender must be settle pid")
		return
	}

	quantity, ok := utils.ParseAmount(params["Quantity"])
	if !ok {
		err = errors.New("quantity incorrect")
		return
	}

	ffpFor, ok = params[ffpSchema.XFFPFor]
	if !ok {
		err = errors.New("not found ffpFor")
		return
	}

	noteId := params[ffpSchema.XFFPNoteID]
	orderId, ok := a.noteIDToOrderID[noteId]
	if !ok {
		err = errors.New("not found noteId")
		return
	}
	order, ok = a.limitOrders[orderId]
	if !ok {
		err = errors.New("not found order")
		return
	}
	// verify order
	if order.Status != schema.LimitOrderStatusExecuting {
		err = errors.New("order status must be executing")
		return
	}

	return
}

func (a *UserAgent) parsePoolPriceNoticeParams(from string, params map[string]string) (order *schema.LimitOrder, poolAmountOutNum *big.Int, err error) {
	orderID, ok := params["X-OrderId"]
	if !ok {
		err = errors.New("orderId is required")
		return
	}

	order, ok = a.limitOrders[orderID]
	if !ok {
		err = errors.New("order not found")
		return
	}
	if order.PoolAgent != from {
		err = errors.New("order not belong to this pool-agent")
		return
	}

	if order.Status != schema.LimitOrderStatusPending {
		err = errors.New("order not pending")
		return
	}

	poolAmountOut, ok := params["AmountOut"]
	if !ok {
		err = errors.New("AmountOut is required")
		return
	}
	poolAmountOutNum, ok = utils.ParseAmount(poolAmountOut)
	if !ok {
		err = errors.New("AmountOut incorrect")
		return
	}
	return
}

func (a *UserAgent) processSettleNotice(order *schema.LimitOrder, ffpFor string, quantity *big.Int, tokenId string) (msgs []*vmmSchema.ResMessage, err error) {
	switch ffpFor {
	case ffpSchema.FFPSuccess: // settle success
		if order.TokenOut != tokenId {
			err = errors.New("received token not equal order token out")
			return
		}

		// must order.MinAmountOut <= quantity <= order.MaxAmountOut
		if quantity.Cmp(order.MinAmountOut) < 0 || quantity.Cmp(order.MaxAmountOut) > 0 {
			err = errors.New("received token quantity incorrect")
			return
		}

		order.Status = schema.LimitOrderStatusCompleted // settled
		order.SettledAmountOut = quantity.String()
		// todo notice to user
	case ffpSchema.FFPRefund: // refund
		if order.TokenIn != tokenId {
			err = errors.New("refund token not equal order token in")
			return
		}
		if quantity.Cmp(order.AmountIn) != 0 {
			err = errors.New("refund token quantity incorrect")
			return
		}

		if order.RetryCount < 3 { //  retry swap
			order.RetryCount = order.RetryCount + 1
			// request Pool Price
			msgs = append(msgs, requestPoolPrice(order))
		} else {
			// order failed
			order.Status = schema.LimitOrderStatusFailedSettle // failed settle
			// todo notice to user
		}
	default:
		err = errors.New("not support X-FFP-For")
	}
	return
}

func (a *UserAgent) assembleSwapNote(order *schema.LimitOrder, swapAmountOut *big.Int) (note ffpSchema.Note, err error) {
	note = ffpSchema.Note{
		Issuer:        a.pid,
		AssetID:       order.TokenIn,
		Amount:        order.AmountIn,
		Holder:        order.PoolAgent,
		HolderAssetID: order.TokenOut,
		HolderAmount:  swapAmountOut,
		IssueDate:     time.Now().UnixMilli(),
		DueDate:       time.Now().UnixMilli() + schema.TimeoutPeriod,
	}
	noteId, err := utils.GenerateNoteID(note)
	if err != nil {
		err = errors.New("generate noteID failed")
		return
	}
	order.NoteId = noteId

	a.noteIDToOrderID[noteId] = order.OrderID
	return
}

func requestPoolPrice(order *schema.LimitOrder) *vmmSchema.ResMessage {
	return &vmmSchema.ResMessage{
		Target: order.PoolAgent,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: schema.AgentRequestPrice},
			{Name: "X-OrderId", Value: order.OrderID},
			{Name: "TokenIn", Value: order.TokenIn},
			{Name: "TokenOut", Value: order.TokenOut},
			{Name: "AmountIn", Value: order.AmountIn.String()},
			{Name: "RetryCount", Value: fmt.Sprintf("%d", order.RetryCount)},
		},
	}
}

func requestSettleSubmitNote(settlePid string, note ffpSchema.Note) *vmmSchema.ResMessage {
	noteStr, _ := utils.EncodeNote(note)
	return &vmmSchema.ResMessage{
		Target: note.AssetID,
		Tags: []goarSchema.Tag{
			{Name: "Action", Value: "Transfer"},
			{Name: "Recipient", Value: settlePid},
			{Name: "Quantity", Value: note.Amount.String()},

			{Name: ffpSchema.XFFPFor, Value: ffpSchema.FFPSubmit},
			{Name: ffpSchema.XFFPNote, Value: noteStr},
		},
	}
}

func getSwapAmountOut(poolAmountOutNum *big.Int, order *schema.LimitOrder) (swapAmountOut *big.Int, err error) {
	// poolAmountOut < minAmountOut
	if poolAmountOutNum.Cmp(order.MinAmountOut) < 0 {
		err = errors.New("price insufficient")
		return
	}

	switch order.RetryCount {
	case 0: // first handle
		// minAmountOut <= poolAmountOut <= maxAmountOut
		if poolAmountOutNum.Cmp(order.MinAmountOut) >= 0 && poolAmountOutNum.Cmp(order.MaxAmountOut) <= 0 {
			swapAmountOut = poolAmountOutNum
			return
		}

		// poolAmountOut > maxAmountOut
		if poolAmountOutNum.Cmp(order.MaxAmountOut) > 0 {
			swapAmountOut = order.MaxAmountOut
			return
		}

	case 1: // second handle
		// poolAmountOut >= minAmountOut
		halfSlippage := new(big.Int).Div(order.Slippage, big.NewInt(2))
		swapAmountOut, err = calcSlippageAmountOut(poolAmountOutNum, halfSlippage)
		if err != nil {
			err = errors.New("price insufficient")
			return
		}

		if swapAmountOut.Cmp(order.MinAmountOut) < 0 {
			swapAmountOut = order.MinAmountOut
		}
		if swapAmountOut.Cmp(order.MaxAmountOut) > 0 {
			swapAmountOut = order.MaxAmountOut
		}
		return

	case 2: // three handle
		swapAmountOut = order.MinAmountOut
		return
	default:
		err = errors.New("order retry count incorrect")
	}
	return
}
