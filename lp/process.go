package lp

import (
	"encoding/json"
	"errors"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/everVision/everpay-kits/utils"

	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/lp/schema"

	"github.com/permadao/permaswap/router"

	//"github.com/permadao/permaswap/lp/schema"
	routerSchema "github.com/permadao/permaswap/router/schema"
)

func (l *Lp) runProcess() {
	for {
		select {

		case msg := <-l.rsdk.SubscribeOrder():
			l.processOrder(msg)

		case msg := <-l.rsdk.SubscribeOrderStatus():
			l.processOrderStatus(*msg)

		case <-l.rsdk.SubscribeReconnect():
			l.reconnect()

		case tx := <-l.sub.Subscribe():
			l.processRouterOrder(tx)

		// api
		case <-l.apiInfoReq:
			l.apiInfoRes <- l.getInfo()

		case lpID := <-l.apiRemoveLpReq:
			l.apiRemoveLpRes <- l.removeLpProc(lpID)

		case lpMsgAdd := <-l.apiAddLpReq:
			l.apiAddLpRes <- l.addLpProc(lpMsgAdd)

		case <-l.close:
			log.Info("process closed")
			close(l.closed)
			return
		}
	}
}

func (l *Lp) processOrder(msg *routerSchema.LpMsgOrder) {
	// proce order one by one
	if l.order != nil {
		// notice router processing
		log.Warn("order is processing", "orderHash", l.order.Bundle.HashHex())
		return
	}
	l.order = msg

	paths := msg.Paths
	if err := router.VerifyBundleAndPaths(msg.Bundle, paths, l.tokens); err != nil {
		log.Error("invalid order", "err", err)
		return
	}

	log.Debug("pathFilter", "originPaths", paths)
	paths = pathsFilter(l.rsdk.AccID, paths)
	log.Debug("pathFilter", "resPaths", paths)
	if err := l.core.Verify(msg.UserAddr, paths); err != nil {
		log.Error("order verify failed", "err", err)
		if err := l.rsdk.RejectOrder(*msg); err != nil {
			log.Error("order reject failed", "err", err)
		}
		return
	}

	l.savePendingOrder(msg)

	if err := l.rsdk.SignOrder(*msg); err != nil {
		log.Error("order sign failed", "err", err)
	}
}

func (l *Lp) processOrderStatus(msg routerSchema.OrderMsgStatus) {
	// clean cache
	defer func() { l.order = nil }()

	if l.order == nil {
		return
	}

	if l.order.Bundle.HashHex() != msg.OrderHash {
		log.Warn("invalid orderHash", "curOrderHash", l.order.Bundle.HashHex(), "orderHash", msg.OrderHash)
		return
	}

	if msg.EverHash == "" {
		log.Warn("invalid everHash", "everHash", msg.EverHash)
		return
	}

	// verify order on everPay with retry
	tx, bundle, status, err := l.rsdk.EverSDK.Cli.BundleByHash(msg.EverHash)
	if err != nil {
		log.Error("can not get submited bundle tx in first time", "everHash", msg.EverHash, "err", err)

		for i := 0; i < 1000; i++ {
			time.Sleep(200 * time.Millisecond)
			tx, bundle, status, err = l.rsdk.EverSDK.Cli.BundleByHash(msg.EverHash)
			if err == nil {
				break
			}
			log.Error("can not get submited bundle tx after retry", "times", i+2, "everHash", msg.EverHash, "err", err)
		}
		log.Warn("bundle tx after retry", "everHash", msg.EverHash, "err", err)
	}

	if l.order.Bundle.HashHex() != bundle.HashHex() {
		log.Warn("invalid orderHash", "curOrderHash", l.order.Bundle.HashHex(), "bundleHash", bundle.HashHex)
		return
	}
	if status.Status != everSchema.InternalStatusSuccess {
		log.Warn("order failed", "curOrderHash", l.order.Bundle.HashHex(), "status", status.Status)
		return
	}

	if err = l.removePendingOrder(); err != nil {
		log.Error("failed remove pending order", "err", err)
	}

	if err = l.saveLatestOrder(tx); err != nil {
		log.Error("failed to save latest order", "err", err)
	}

	go l.saveOrder(msg, l.order)

	// if order is success update
	paths := pathsFilter(l.rsdk.AccID, l.order.Paths)
	if err := l.core.Update(l.order.UserAddr, paths); err != nil {
		log.Error("core update failed", "err", err)
	}
	log.Info("core update success")

	if err = l.UpdateLiquidity(); err != nil {
		log.Error("can not update liquidity config", "err", err)
	}
}

func (l *Lp) processRouterOrder(tx everSchema.TxResponse) {
	if l.order == nil {
		return
	}
	bundleData := everSchema.BundleData{}
	err := json.Unmarshal([]byte(tx.Data), &bundleData)
	if err != nil {
		log.Warn("invalid router order data, ignore.", "order data", tx.Data)
		return
	}

	orderHash := bundleData.Bundle.Bundle.HashHex()
	if orderHash == l.order.Bundle.HashHex() {
		msg := routerSchema.OrderMsgStatus{
			Event:     "status",
			OrderHash: orderHash,
			EverHash:  tx.EverHash,
		}
		l.processOrderStatus(msg)
	}

}

func (l *Lp) removeLpProc(lpID string) *schema.RemoveLpRes {
	if l.order != nil {
		log.Error("order is processing, so can not update lp.", "orderHash", l.order.Bundle.HashHex())
		return &schema.RemoveLpRes{
			LpID:   lpID,
			Result: "failed",
			Error:  "err_can_not_update_lp_with_order",
		}
	}

	if _, ok := l.core.Lps[lpID]; !ok {
		return &schema.RemoveLpRes{
			LpID:   lpID,
			Result: "failed",
			Error:  "err_invalid_lpid",
		}
	}
	lp := *l.core.Lps[lpID]
	l.rsdk.RemoveLiquidity(LpToRemoveMsg(lp))

	// wait and process lp remove response
	resp := <-l.rsdk.SubscribeLpRemoveResponseOnce()
	if resp.LpID != "" && resp.LpID != lpID {
		log.Error("Lp remove resopnse wiht invalid lpid", "LpRemoveResponse", resp)
		return &schema.RemoveLpRes{
			LpID:   lpID,
			Result: "failed",
			Error:  "err_invalid_response",
		}
	}

	if resp.Msg != "ok" {
		return &schema.RemoveLpRes{
			LpID:   lpID,
			Result: resp.Msg,
			Error:  resp.Error,
		}
	}

	if _, err := l.core.RemoveLiquidityByID(lpID); err != nil {
		log.Error("Lp remove in local core failed", "err", err)
		// core is not consistent with router's. so panic to restart it.
		panic(errors.New("err_remove_lp_in_local_core"))
	}

	if err := l.UpdateLiquidity(); err != nil {
		log.Error("can not update liquidity config file", "err", err)
	}

	return &schema.RemoveLpRes{
		LpID:   lpID,
		Result: "ok",
		Error:  "",
	}

}

func (l *Lp) addLpProc(msg *routerSchema.LpMsgAdd) *schema.AddLpRes {
	if l.order != nil {
		log.Error("order is processing, so can not update lp.", "orderHash", l.order.Bundle.HashHex())
		return &schema.AddLpRes{
			Result: "failed",
			Error:  "err_can_not_update_lp_with_order",
		}
	}

	pool, err := l.core.FindPool(msg.TokenX, msg.TokenY, msg.FeeRatio)
	if err != nil {
		return &schema.AddLpRes{
			Result: "failed",
			Error:  "err_no_pool",
		}
	}

	lpID := core.GetLpID(pool.ID(), l.rsdk.AccID, msg.LowSqrtPrice, msg.HighSqrtPrice, msg.PriceDirection)
	if _, ok := l.core.Lps[lpID]; ok {
		return &schema.AddLpRes{
			Result: "failed",
			Error:  "err_close_lp_first",
		}
	}
	lpToAdd, err := core.NewLp(
		pool.ID(), msg.TokenX, msg.TokenY, l.rsdk.AccID,
		msg.FeeRatio, msg.LowSqrtPrice, msg.CurrentSqrtPrice, msg.HighSqrtPrice,
		msg.Liquidity, msg.PriceDirection,
	)
	if err != nil {
		return &schema.AddLpRes{
			Result: "failed",
			Error:  "err_invalid_lp",
		}
	}

	lps := l.core.GetLps(l.rsdk.AccID)
	lps = append(lps, *lpToAdd)
	if err := l.checkBalance(lps, false); err != nil {
		return &schema.AddLpRes{
			Result: "failed",
			Error:  "err_balance_not_enough",
		}
	}

	l.rsdk.AddLiquidity(*msg)

	// wait and process lp remove response
	resp := <-l.rsdk.SubscribeLpAddResponseOnce()
	if resp.Msg != "ok" {
		return &schema.AddLpRes{
			LpID:   resp.LpID,
			Result: resp.Msg,
			Error:  resp.Error,
		}
	}

	if err := l.core.AddLiquidity(l.rsdk.AccID, *msg); err != nil {
		log.Error("Lp add in local core failed", "err", err)
		// core is not consistent with router's. so panic to restart it.
		panic(errors.New("err_add_lp_in_local_core"))
	}

	if err := l.UpdateLiquidity(); err != nil {
		log.Error("can not update liquidity config file", "err", err)
	}

	return &schema.AddLpRes{
		LpID:   resp.LpID,
		Result: "ok",
		Error:  "",
	}
}

func pathsFilter(lpAccID string, paths []coreSchema.Path) (resPaths []coreSchema.Path) {
	resPaths = []coreSchema.Path{}
	for _, path := range paths {
		_, from, _ := utils.IDCheck(path.From)
		_, to, _ := utils.IDCheck(path.To)

		if from == lpAccID || to == lpAccID {
			resPaths = append(resPaths, path)
		}
	}

	return
}

func (l *Lp) saveOrder(msg routerSchema.OrderMsgStatus, order *routerSchema.LpMsgOrder) {
	orderToSave := schema.Order{
		UserAddr:    l.rsdk.AccID,
		EverHash:    msg.EverHash,
		OrderStatus: msg.Status,
		LpMsgOrder:  string(order.Marshal()),
	}
	if err := l.wdb.CreateOrder(&orderToSave, nil); err != nil {
		log.Error("save order to db failed", "err", err)
	}
}
