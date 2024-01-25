package router

import (
	"encoding/json"
	"math/big"
	"time"

	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/router/schema"
	"github.com/everVision/everpay-kits/utils"
	"github.com/google/uuid"
)

func (r *Router) runUserMsgUnmarshal() {
	r.userHub.Run(nil, func(id string) {
		r.userUnregister <- id
	})

	for {
		src := <-r.userHub.Subscribe()

		msg := &schema.UserMsg{}
		if err := json.Unmarshal(src.Data, msg); err != nil {
			r.userHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
			log.Error("invalid message from user", "err", err, "msg", string(src.Data))
			continue
		}

		switch msg.Event {
		case schema.UserMsgEventQuery:
			qryMsg := &schema.UserMsgQuery{}
			if err := json.Unmarshal(src.Data, qryMsg); err != nil {
				r.userHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from user", "err", err, "msg", string(src.Data))
				continue
			}

			// append src session di to internal msg
			qryMsg.ID = src.ID
			r.userQuery <- qryMsg

		case schema.UserMsgEventSubmit:
			submitMsg := &schema.UserMsgSubmit{}
			if err := json.Unmarshal(src.Data, submitMsg); err != nil {
				r.userHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from user", "err", err, "msg", string(src.Data))
				continue
			}

			submitMsg.ID = src.ID
			r.userSubmit <- submitMsg

		default:
			r.userHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
			log.Error("invalid message action", "msg", string(src.Data))
		}

	}
}

func (r *Router) userQueryProc(msg *schema.UserMsgQuery) {
	orderMsg, err := r.queryOrder(msg)
	if err != nil {
		r.userHub.Publish(msg.ID, []byte(NewWsErr(err.Error()).Error()))
		return
	}

	// update cache userQueryTag
	r.cleanUserQueryTag(msg.ID)
	for _, item := range orderMsg.Bundle.Items {
		if sessions, ok := r.userQueryTag[item.Tag]; ok {
			sessions[msg.ID] = msg
		} else {
			sessions := map[string]*schema.UserMsgQuery{msg.ID: msg}
			r.userQueryTag[item.Tag] = sessions
		}
	}

	r.userHub.Publish(msg.ID, orderMsg.Marshal())
}

func (r *Router) userSubmitProc(msg *schema.UserMsgSubmit) {
	log.Info("order submited", "user", msg.Address, "tokenIn", msg.TokenIn, "tokenOut", msg.TokenOut,
		"paths length", len(msg.Paths))

	// check black list
	if r.penalty.IsBlackListed(msg.Address) {
		log.Error("user is blacklisted", "user", msg.Address)
		r.userHub.Publish(msg.ID, []byte(WsErrBlackListed.Error()))
		return
	}

	bundle := msg.Bundle
	// verify bundle tx
	userAddr, _, err := VerifyBundleByAddr(bundle, msg.Address, r.chainID)
	if err != nil {
		r.userHub.Publish(msg.ID, []byte(err.Error()))
		return
	}

	//verify paths
	if err := VerifyBundleAndPaths(bundle.Bundle, msg.Paths, r.tokens); err != nil {
		r.userHub.Publish(msg.ID, []byte(err.Error()))
		return
	}

	// verify price in core
	if err = r.core.Verify(userAddr, msg.Paths); err != nil {
		r.userHub.Publish(msg.ID, []byte(NewWsErr(err.Error()).Error()))
		return
	}

	// filter & get(copy) lp session id
	lpSessions := make(map[string]string) // lp addr -> lp session id
	for _, item := range bundle.Items {
		_, lpAcc, err := utils.IDCheck(item.From)
		if err != nil {
			r.userHub.Publish(msg.ID, []byte(WsErrInvalidAddress.Error()))
			return
		}

		if lpAcc == userAddr {
			continue
		}

		lpID, ok := r.lpAddrToID[lpAcc]
		if !ok {
			r.userHub.Publish(msg.ID, []byte(WsErrNotFoundLp.Error()))
			return
		}

		lpSessions[lpAcc] = lpID
	}

	// lock core: move router core to order core
	lps := map[string]*coreSchema.Lp{}
	isRemoved := map[string]bool{}
	for _, path := range msg.Paths {

		// path fee
		if path.LpID == "" {
			continue
		}

		// don not remove lp twice
		_, ok := lps[path.LpID]
		if ok {
			continue
		}

		lp, err := r.core.RemoveLiquidityByID(path.LpID)

		// make sure only remove one liquidity once
		if isRemoved[path.LpID] {
			continue
		}

		if err != nil {
			log.Warn("move router core to order core failed(RemoveLiquidityByID)", "err", err)
			continue
		}

		lps[lp.ID()] = lp
	}

	order := NewOrder(r.chainID, msg, bundle, lps, r, lpSessions, r.dryRun)
	r.orders[order.Bundle.HashHex()] = order
	order.Run()
}

func (r *Router) userUnregisterProc(id string) {
	r.cleanUserQueryTag(id)
}

// return 0 when price impact is very little;
// return "" when failed to get price impact
func (r *Router) calPriceImpact(msg *schema.UserMsgQuery, price *big.Float) (priceImpact string) {
	amountIn, ok := coreSchema.MinAmountInsForPriceQuery[msg.TokenIn]
	if !ok {
		log.Error("Failed to get min amountIn for price impact", "tokenIn", msg.TokenIn)
		return
	}

	x, ok := new(big.Int).SetString(amountIn, 10)
	if !ok {
		return
	}
	y, ok := new(big.Int).SetString(msg.AmountIn, 10)
	if !ok {
		return
	}

	if y.Cmp(x) != 1 {
		return "0"
	}

	umq := schema.UserMsgQuery{
		Event:    msg.Event,
		Address:  msg.Address,
		TokenIn:  msg.TokenIn,
		TokenOut: msg.TokenOut,
		AmountIn: amountIn,
	}

	paths, err := r.core.Query(umq)
	if err != nil || len(paths) == 0 {
		log.Error("Failed to get path", "tokenIn", msg.TokenIn, "tokenOut",
			msg.TokenOut, "err", err, "len(paths)", len(paths))
		return
	}

	price2, _, _ := CalPrice(paths, r.tokens, msg.TokenIn, msg.TokenOut, msg.Address)
	if price2 == nil {
		log.Error("Failed to calcaulte indicative price", "tokenIn", msg.TokenIn, "tokenOut",
			msg.TokenOut)
		return
	}

	// price_impact = abs(current_price - executed_price)/current_price
	//diff := new(big.Float)
	//if price2.Cmp(price) == 1 {
	//	diff.Sub(price2, price)
	//} else {
	//	diff.Sub(price, price2)
	//}
	//priceImpact = new(big.Float).Quo(diff, price2).Text('f', 10)

	// uniswap v3 way: price_impact = (current_price - executed_price)/executed_price
	diff := new(big.Float).Sub(price2, price)
	priceImpact = new(big.Float).Quo(diff, price).Text('f', 10)

	return
}

func (r *Router) queryOrder(msg *schema.UserMsgQuery) (*schema.UserMsgOrder, error) {
	paths, err := r.core.Query(*msg)
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, WsErrNotFoundPath
	}

	// Convert Paths To Bundle
	bundle, err := ConvertPathsToBundle(paths, r.tokens, time.Now().Unix()+120, uuid.NewString())
	if err != nil {
		return nil, err
	}

	price, _, _ := CalPrice(paths, r.tokens, msg.TokenIn, msg.TokenOut, msg.Address)

	priceImpact := r.calPriceImpact(msg, price)
	return &schema.UserMsgOrder{
		Event:       schema.UserMsgEventOrder,
		UserAddr:    msg.Address,
		TokenIn:     msg.TokenIn,
		TokenOut:    msg.TokenOut,
		Price:       price.Text('f', 10),
		PriceImpact: priceImpact,
		Bundle:      bundle,
		Paths:       paths,
	}, nil
}

// if price change, push new order to user
func (r *Router) pushNewOrder(tokenTags ...string) {
	for _, tag := range tokenTags {
		msgs, ok := r.userQueryTag[tag]
		if !ok {
			continue
		}

		for _, msg := range msgs {
			orderMsg, err := r.queryOrder(msg)
			if err != nil {
				continue
			}

			r.userHub.Publish(msg.ID, orderMsg.Marshal())
		}
	}
}

func (r *Router) cleanUserQueryTag(id string) {
	for _, tagSessions := range r.userQueryTag {
		delete(tagSessions, id)
	}
}
