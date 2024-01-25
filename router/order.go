package router

import (
	"math"
	"math/big"
	"strconv"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/everVision/everpay-kits/utils"
	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/router/schema"
)

func (r *Router) orderStatusProc(order *Order) {
	// clean: order destruction
	defer func(orderHash string) {
		delete(r.orders, orderHash)
	}(order.Bundle.HashHex())

	// move order lps back to router core after order finished
	unsigners := map[string]string{}
	for _, lp := range order.Lps {

		// No ws session means this address is offline during the time this lp in order (see func lpUnregisterProc), so no need move lp back.
		if _, ok := r.lpAddrToID[lp.AccID]; !ok {
			log.Info("lp is offline, no need move back", "lp address", lp.AccID)
			continue
		}

		if err := r.core.AddLiquidityByLp(lp); err != nil {
			log.Error("move order core to router core failed(Add LiquidityByLp), need recover core", "err", err)
			// TODO: handle core move error
		}

		sid := r.lpAddrToID[lp.AccID]
		if _, ok := order.isLpSigned[sid]; !ok {
			log.Error("lp have not sign order.", "lpID", lp.ID(), "accid", lp.AccID)
			unsigners[sid] = lp.AccID
		}
	}
	order.Lps = nil

	//close the ws con fo unsigned lp
	for sid, accid := range unsigners {
		r.lpHub.CloseSession(sid)
		r.penalty.AddFailRecord(accid, time.Now().Unix(), order.EverHash, schema.LpPenaltyForNoSign)
	}

	// close the ws con of reject lp
	for _, rejectMsg := range order.rejectMsgs {
		r.lpHub.CloseSession(rejectMsg.ID)
	}

	if order.Status != schema.OrderStatusSuccess {
		// close the ws con of the insufficient balance lp
		if order.InternalErr != nil && order.InternalErr.Msg == "err_insufficient_balance" {
			path := order.UserMsg.Paths[order.InternalErr.Index]
			if path.From != order.UserMsg.Address {
				sid, ok := r.lpAddrToID[path.From]
				if ok {
					r.lpHub.CloseSession(sid)
					log.Error("lp err_insufficient_balance. close", "accid", path.From)
					r.penalty.AddFailRecord(path.From, time.Now().Unix(), order.EverHash, schema.LpPenaltyForNoEnoughBalance)
				} else {
					log.Error("failed to find ws session id (err_insufficient_balance)", "accid", path.From)
				}
			} else {
				log.Error("user err_insufficient_balance", "accid", path.From)
				r.penalty.AddFailRecord(path.From, time.Now().Unix(), order.EverHash, schema.UserPenaltyForNoEnoughBalance)
			}
		}
		return
	}

	paths := order.UserMsg.Paths
	// update core & push new order to target user
	if err := r.core.Update(order.UserMsg.Address, paths); err != nil {
		log.Error("can not update core", "err", err)
		return
	}
	log.Info("core updated")
	// push newOrder to user
	// get tokenTags from paths
	tokenTags := map[string]bool{}
	for _, path := range paths {
		tokenTags[path.TokenTag] = true
	}
	// get query msg & push
	for tag, _ := range tokenTags {
		qrys := r.userQueryTag[tag]
		if qrys == nil {
			continue
		}

		for _, qry := range qrys {
			r.userQueryProc(qry)
		}
	}

	go r.saveOrder(order)
}

func (r *Router) saveOrder(order *Order) {
	if r.dryRun {
		return
	}

	msg := order.UserMsg
	price, tokenInAmount, tokenOutAmount := CalPrice(msg.Paths, r.tokens, msg.TokenIn, msg.TokenOut, msg.Address)
	_, userAddr, _ := utils.IDCheck(msg.Address)
	po := &schema.PermaOrder{
		UserAddr:       userAddr,
		EverHash:       order.EverHash,
		TokenInTag:     msg.TokenIn,
		TokenOutTag:    msg.TokenOut,
		TokenInAmount:  tokenInAmount.Text('f', 10),
		TokenOutAmount: tokenOutAmount.Text('f', 10),
		Price:          price.Text('f', 10),
		OrderStatus:    order.Status,
		OrderTimestamp: order.Timestamp,
	}
	if err := r.wdb.CreatePermaOrder(po, nil); err != nil {
		log.Error("perma order save to db falied", "err", err)
		return
	}
	swapInputs, err := core.PathsToSwapInputs(order.UserMsg.Address, order.UserMsg.Paths)
	if err != nil {
		log.Error("paths to swapInputs failed when save perma order volume", "err", err)
		return
	}

	for lpID, si := range swapInputs {
		lp, ok := order.router.core.Lps[lpID]
		if !ok {
			log.Warn("failed to find lp when save perma order volume")
			continue
		}

		pool, ok := order.router.core.Pools[lp.PoolID]
		if !ok {
			log.Warn("failed to find pool when save perma order volume")
			continue
		}
		feeRatio, _ := pool.FeeRatio.Float64()

		tokens := order.router.tokens
		tokenIn := tokens[si.TokenIn]
		tokenOut := tokens[si.TokenOut]
		tokenInDecFactor := new(big.Float).SetFloat64(math.Pow(10, float64(tokenIn.Decimals)))
		tokenOutDecFactor := new(big.Float).SetFloat64(math.Pow(10, float64(tokenOut.Decimals)))
		tokenInAmount := new(big.Float).Quo(new(big.Float).SetInt(si.AmountIn), tokenInDecFactor)
		tokenOutAmount := new(big.Float).Quo(new(big.Float).SetInt(si.AmountOut), tokenOutDecFactor)
		amountIn, _ := tokenInAmount.Float64()
		amountOut, _ := tokenOutAmount.Float64()

		tokenXIsTokenIn := true
		amountX := amountIn
		amountY := amountOut
		rewardX := amountIn * feeRatio
		rewardY := float64(0)
		if si.TokenIn == pool.TokenYTag {
			tokenXIsTokenIn = false
			amountX = amountOut
			amountY = amountIn
			rewardX = float64(0)
			rewardY = amountIn * feeRatio
		}
		pv := &schema.PermaVolume{
			OrderID:         po.ID,
			EverHash:        po.EverHash,
			PoolID:          lp.PoolID,
			AccID:           lp.AccID,
			LpID:            lp.ID(),
			TokenXIsTokenIN: tokenXIsTokenIn,
			AmountX:         amountX,
			AmountY:         amountY,
			RewardX:         rewardX,
			RewardY:         rewardY,
		}
		if err := r.wdb.CreatePermaVolume(pv, nil); err != nil {
			log.Error("perma volume save to db failed", "err", err)
			continue
		}

		lpReward := &schema.PermaLpReward{
			LpID:    lpID,
			PoolID:  lp.PoolID,
			AccID:   lp.AccID,
			RewardX: rewardX,
			RewardY: rewardY,
		}
		if err := r.wdb.UpdatePermaLpReward(lpReward, nil); err != nil {
			log.Error("perma update lp reward failed", "err", err)
			continue
		}
	}
}

// Order doing in order goroutines
// 1. make LP signature
// 2. submit order to everPay
// 3. update order status & notice update core
type Order struct {
	ChainID     int64
	UserMsg     *schema.UserMsgSubmit
	Status      string
	InternalErr *everSchema.InternalErr
	EverHash    string
	Timestamp   int64
	Bundle      *everSchema.BundleWithSigs
	Lps         map[string]*coreSchema.Lp // lpID -> lp
	lpAddrToID  map[string]string         // addr -> session id
	isLpSigned  map[string]bool           // sessionid -> bool

	lpSig    chan *schema.LpMsgSign
	lpReject chan *schema.LpMsgReject
	router   *Router

	rejectMsgs []*schema.LpMsgReject

	dryRun bool
}

// NewOrder lpSessions: lp addr -> lp session id
func NewOrder(ChainID int64,
	userMsg *schema.UserMsgSubmit, bundle everSchema.BundleWithSigs,
	lps map[string]*coreSchema.Lp, router *Router, lpSessions map[string]string, dryRun bool,
) *Order {

	return &Order{
		ChainID:    ChainID,
		UserMsg:    userMsg,
		Status:     schema.OrderStatusPending,
		Bundle:     &bundle,
		Lps:        lps,
		lpAddrToID: lpSessions,
		isLpSigned: make(map[string]bool),

		lpSig:    make(chan *schema.LpMsgSign),
		lpReject: make(chan *schema.LpMsgReject),

		router: router,

		dryRun: dryRun,
	}
}

func (o *Order) Run() {
	go o.process()
}

func (o *Order) process() {
	// ask first time
	o.askSig()

	// ticker for timeout & retry
	timeoutTicker := time.NewTicker(schema.OrderExpire)
	retryTicker := time.NewTicker(2 * time.Second)
	defer func() {
		timeoutTicker.Stop()
		retryTicker.Stop()

		o.submitToEver()
		o.notice()
		log.Info("order sataus noticed")
		o.router.orderStatus <- o
		log.Info("-------------- order done --------------")
	}()

	for {
		select {
		case msg := <-o.lpSig:
			_, sig, err := VerifyBundleByAddr(msg.Bundle, msg.Address, o.ChainID)
			if err != nil {
				log.Error("invalid bundle sig", "err", err)
				continue
			}

			for k, v := range sig {
				o.Bundle.Sigs[k] = v
			}
			o.isLpSigned[msg.ID] = true

			if _, _, err := utils.VerifyBundleSigs(*o.Bundle, time.Now().UnixNano(), int(o.ChainID)); err == nil {
				return
			}

		// TODO: case msg := <- o.lpReject
		// when lp reject, router need verify it's Legitimate Rejection
		case msg := <-o.lpReject:
			o.rejectMsgs = append(o.rejectMsgs, msg)

		case <-retryTicker.C:
			o.askSig()

		case <-timeoutTicker.C:
			return
		}
	}
}

func (o *Order) askSig() {
	// get all session id
	ids := []string{}
	isAsked := map[string]bool{}
	for _, item := range o.Bundle.Items {
		_, from, _ := utils.IDCheck(item.From)
		id, ok := o.lpAddrToID[from]
		if !ok {
			continue
		}
		if o.isLpSigned[id] {
			continue
		}
		if isAsked[id] {
			continue
		} else {
			isAsked[id] = true
		}
		ids = append(ids, id)
	}

	// ask lp nodes
	for _, id := range ids {
		o.router.lpHub.Publish(id, schema.LpMsgOrder{
			UserAddr: o.UserMsg.Address,
			Bundle:   o.Bundle.Bundle,
			Paths:    o.UserMsg.Paths,
		}.Marshal())
	}
}

func (o *Order) submitToEver() {
	if o.dryRun || o.router.sdk == nil {
		o.EverHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
		o.Status = schema.OrderStatusSuccess
		return
	}

	everSDK := o.router.sdk
	var err error
	var everTx *everSchema.Transaction
	for {
		if o.Bundle.Expiration < time.Now().Unix() {
			o.Status = schema.OrderStatusExpired
			return
		}
		// need retry
		if everTx, err = everSDK.Bundle("ethereum-eth-0x0000000000000000000000000000000000000000", everSDK.AccId, big.NewInt(0), *o.Bundle); err == nil {
			break
		} else {
			log.Error("sumbit to everPay failed", "err", err)
			time.Sleep(1 * time.Second)
		}
	}
	// update status
	o.EverHash = everTx.HexHash()
	o.Timestamp, _ = strconv.ParseInt(everTx.Nonce, 10, 64)

	// get bundle tx status
	_, _, status, err := everSDK.Cli.BundleByHash(o.EverHash)
	if err != nil {
		log.Error("can not get submited bundle tx in first time", "everHash", o.EverHash, "err", err)

		for i := 0; i < 1000; i++ {
			time.Sleep(200 * time.Millisecond)
			_, _, status, err = everSDK.Cli.BundleByHash(o.EverHash)
			if err == nil {
				break
			}
			log.Error("can not get submited bundle tx after retry", "times", i+2, "everHash", o.EverHash, "err", err)
		}
		log.Warn("bundle tx after retry", "everHash", o.EverHash, "err", err)
	}

	switch status.Status {
	case everSchema.InternalStatusSuccess:
		o.Status = schema.OrderStatusSuccess
	case everSchema.InternalStatusFailed:
		log.Error("Order submitToEver failed", "tx InternalStatus", status.Status, "InternalErr", status.InternalErr, "err", err)
		o.Status = schema.OrderStatusFailed
		o.InternalErr = status.InternalErr
	default:
		log.Error("invalid InternalStatus", "status", status.Status)
	}
}

func (o *Order) notice() {
	statusMsg := schema.OrderMsgStatus{
		OrderHash: o.Bundle.HashHex(),
		EverHash:  o.EverHash,
		Status:    o.Status,
	}

	// notice user
	o.router.userHub.Publish(o.UserMsg.ID, statusMsg.Marshal())

	// notice lps
	for _, id := range o.lpAddrToID {
		o.router.lpHub.Publish(id, statusMsg.Marshal())
	}
}

func VerifyBundleByAddr(bundle everSchema.BundleWithSigs, address string, chainID int64) (accid string, accsig map[string]string, err error) {
	// VerifyBundleByAddr check bundle sigs by only one address
	// sdk.VerifyBundleSigs check every sigs in bundle

	if len(bundle.Items) == 0 || len(bundle.Sigs) == 0 {
		err = WsErrInvalidOrder
		return
	}

	_, accid, err = utils.IDCheck(address)
	if err != nil {
		err = WsErrInvalidAddress
		return
	}

	signature := ""
	for addr, sig := range bundle.Sigs {
		_, id, err := utils.IDCheck(addr)
		if err != nil {
			return "", nil, WsErrInvalidAddress
		}
		if id == accid {
			accsig = make(map[string]string)
			accsig[addr] = sig
			signature = sig
			break
		}
	}

	if signature == "" {
		err = WsErrInvalidSignature
		return
	}

	// err = account.VerifyTxSig(int(chainID), accid, bundle.String(), signature, time.Now().UnixNano())
	// if err != nil {
	// 	err = WsErrInvalidSignature
	// 	return
	// }

	return
}
