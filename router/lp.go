package router

import (
	"encoding/json"

	"github.com/permadao/permaswap/core"
	"github.com/permadao/permaswap/router/schema"
	"github.com/everVision/everpay-kits/utils"
	"github.com/google/uuid"
	"golang.org/x/mod/semver"
)

func (r *Router) runLpMsgUnmarshal() {
	r.lpHub.Run(
		func(id string) {
			r.lpInit <- id
		},
		func(id string) {
			r.lpUnregister <- id
		})

	for {
		src := <-r.lpHub.Subscribe()

		msg := &schema.LpMsg{}
		if err := json.Unmarshal(src.Data, msg); err != nil {
			r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
			log.Error("invalid message from lp", "err", err, "msg", string(src.Data))
			continue
		}

		// authorization verification
		if msg.Event != schema.LpMsgEventRegister && !r.isLpByID(src.ID) {
			r.lpHub.Publish(src.ID, []byte(WsErrNoAuthorization.Error()))
			log.Error("no auth from lp", "id", src.ID)
			continue
		}

		switch msg.Event {
		case schema.LpMsgEventRegister:
			regMsg := &schema.LpMsgRegister{}
			if err := json.Unmarshal(src.Data, regMsg); err != nil {
				r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from lp", "err", err, "msg", string(src.Data))
				continue
			}

			// append src session id to internal msg
			regMsg.ID = src.ID
			r.lpRegister <- regMsg

		case schema.LpMsgEventAdd:
			addMsg := &schema.LpMsgAdd{}
			if err := json.Unmarshal(src.Data, addMsg); err != nil {
				r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from lp", "err", err, "msg", string(src.Data))
				continue
			}

			addMsg.ID = src.ID
			r.lpAdd <- addMsg

		case schema.LpMsgEventRemove:
			removeMsg := &schema.LpMsgRemove{}
			if err := json.Unmarshal(src.Data, removeMsg); err != nil {
				r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from lp", "err", err, "msg", string(src.Data))
				continue
			}

			removeMsg.ID = src.ID
			r.lpRemove <- removeMsg

		case schema.LpMsgEventSign:
			signMsg := &schema.LpMsgSign{}
			if err := json.Unmarshal(src.Data, signMsg); err != nil {
				r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from lp", "err", err, "msg", string(src.Data))
				continue
			}

			signMsg.ID = src.ID
			r.lpSign <- signMsg

		case schema.LpMsgEventReject:
			rejectMsg := &schema.LpMsgReject{}
			if err := json.Unmarshal(src.Data, rejectMsg); err != nil {
				r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
				log.Error("invalid message from lp", "err", err, "msg", string(src.Data))
				continue
			}

			rejectMsg.ID = src.ID
			r.lpReject <- rejectMsg

		default:
			r.lpHub.Publish(src.ID, []byte(WsErrInvalidMsg.Error()))
			log.Error("invalid message action", "msg", string(src.Data))
		}
	}
}

func (r *Router) isLpByID(id string) bool {
	_, ok := r.lpIDtoAddr[id]
	return ok
}

func (r *Router) isLpByAddr(addr string) bool {
	_, ok := r.lpAddrToID[addr]
	return ok
}

func (r *Router) lpInitProc(id string) {
	salt := uuid.NewString()
	r.lpSalt[id] = salt

	r.lpHub.Publish(id, schema.LpMsgSalt{Salt: salt}.Marshal())
}

func (r *Router) lpRegisterProc(msg *schema.LpMsgRegister) {

	// check lpclient info
	lpClienInfo, ok := r.LpClientInfo[msg.LpClientName]
	if !ok {
		log.Error("lp client name is invalid")
		r.lpHub.Publish(msg.ID, []byte(WsErrInvalidLpClient.Error()))
		return
	}
	if semver.Compare(lpClienInfo.Version, msg.LpClientVersion) == 1 {
		log.Error("lp client version is invalid", "lp name", msg.LpClientName, "lp version", msg.LpClientVersion, "user", msg.Address)
		r.lpHub.Publish(msg.ID, []byte(WsErrInvalidLpClient.Error()))
		return
	}

	// check black list
	if ok := r.penalty.IsBlackListed(msg.Address); ok {
		log.Error("lp account is in black list")
		r.lpHub.Publish(msg.ID, []byte(WsErrBlackListed.Error()))
		return
	}

	// get salt
	salt, ok := r.lpSalt[msg.ID]
	if !ok {
		log.Error("salt not found")
		r.lpHub.Publish(msg.ID, []byte(WsErrNotFoundSalt.Error()))
		return
	}

	// get account address
	accType, accid, err := utils.IDCheck(msg.Address)
	if err != nil {
		log.Warn("invalid account", "err", err)
		r.lpHub.Publish(msg.ID, []byte(WsErrInvalidAddress.Error()))
		return
	}

	//check is nft holder
	if r.CheckNFTOrNot() && !r.NFTInfo.Passed(accid) {
		log.Warn("not a nft owner", "acc.ID", accid)
		r.lpHub.Publish(msg.ID, []byte(WsErrNotNFTOwner.Error()))
		return
	}

	// check duplicate registration
	if r.isLpByAddr(accid) {
		r.lpHub.Publish(msg.ID, []byte(WsErrDuplicateRegistration.Error()))
		return
	}

	// verify sig
	err = VerifySig(accType, accid, salt, msg.Sig, int(r.chainID))
	if err != nil {
		log.Warn("invalid account", "err", err)
		r.lpHub.Publish(msg.ID, []byte(WsErrInvalidSignature.Error()))
		return
	}

	// register
	r.lpAddrToID[accid] = msg.ID
	r.lpIDtoAddr[msg.ID] = accid
	r.lpHub.Publish(msg.ID, schema.LpMsgOk.Marshal())
}

func (r *Router) lpUnregisterProc(id string) {
	if addr, ok := r.lpIDtoAddr[id]; ok {
		err := r.core.RemoveLiquidityByAddress(addr)
		if err != nil {
			log.Warn("failed to remove lps in core", "address", addr, "error", err)
		}
		delete(r.lpAddrToID, addr)
		delete(r.lpIDtoAddr, id)
		delete(r.lpSalt, id)
		log.Info("lp removed", "address", addr)
	}
}

func (r *Router) lpAddProc(msg *schema.LpMsgAdd) {
	addr := r.lpIDtoAddr[msg.ID]

	pool, err := r.core.FindPool(msg.TokenX, msg.TokenY, msg.FeeRatio)
	if err != nil {
		r.lpHub.Publish(msg.ID, schema.LpMsgAddResponse{
			Msg:   "failed",
			Error: err.Error(),
		}.Marshal())
		return
	}

	lpID := core.GetLpID(pool.ID(), addr, msg.LowSqrtPrice, msg.HighSqrtPrice, msg.PriceDirection)
	for _, order := range r.orders {
		if _, ok := order.Lps[lpID]; ok {
			r.lpHub.Publish(msg.ID, schema.LpMsgAddResponse{
				LpID:  lpID,
				Msg:   "failed",
				Error: WsErrCanNotUpdateLp.Error(),
			}.Marshal())
			return
		}
	}

	err = r.core.AddLiquidity(addr, *msg)
	if err != nil {
		r.lpHub.Publish(msg.ID, schema.LpMsgAddResponse{
			LpID:  lpID,
			Msg:   "failed",
			Error: err.Error(),
		}.Marshal())
		return
	}

	r.lpHub.Publish(msg.ID, schema.LpMsgAddResponse{
		LpID: lpID,
		Msg:  "ok",
	}.Marshal())

	// add token tag to cache
	go func(tagA, tagB string) {
		r.apiTokenTagsLock.Lock()
		defer r.apiTokenTagsLock.Unlock()

		r.apiTokenTags[tagA] = true
		r.apiTokenTags[tagB] = true

	}(msg.TokenX, msg.TokenY)

	r.pushNewOrder(msg.TokenX, msg.TokenY)

	log.Info("lp added", "address", addr, "msg", msg)
}

func (r *Router) lpRemoveProc(msg *schema.LpMsgRemove) {
	addr := r.lpIDtoAddr[msg.ID]

	pool, err := r.core.FindPool(msg.TokenX, msg.TokenY, msg.FeeRatio)
	if err != nil {
		r.lpHub.Publish(msg.ID, schema.LpMsgRemoveResponse{
			Msg:   "failed",
			Error: err.Error(),
		}.Marshal())
		return
	}

	lpID := core.GetLpID(pool.ID(), addr, msg.LowSqrtPrice, msg.HighSqrtPrice, msg.PriceDirection)
	for _, order := range r.orders {
		if _, ok := order.Lps[lpID]; ok {
			r.lpHub.Publish(msg.ID, schema.LpMsgRemoveResponse{
				LpID:  lpID,
				Msg:   "failed",
				Error: WsErrCanNotUpdateLp.Error(),
			}.Marshal())
			return
		}
	}

	err = r.core.RemoveLiquidity(addr, *msg)
	if err != nil {
		r.lpHub.Publish(msg.ID, schema.LpMsgRemoveResponse{
			LpID:  lpID,
			Msg:   "failed",
			Error: err.Error(),
		}.Marshal())
		return
	}

	r.lpHub.Publish(msg.ID, schema.LpMsgRemoveResponse{
		LpID: lpID,
		Msg:  "ok",
	}.Marshal())

	r.pushNewOrder(msg.TokenX, msg.TokenY)
}

func (r *Router) lpSignProc(msg *schema.LpMsgSign) {
	// get order from cache
	order, ok := r.orders[msg.Bundle.HashHex()]
	if !ok {
		log.Warn("not found order")
		r.lpHub.Publish(msg.ID, []byte(WsErrInvalidOrder.Error()))
		return
	}
	order.lpSig <- msg
}

func (r *Router) lpRejectProc(msg *schema.LpMsgReject) {
	// get order from cache
	order, ok := r.orders[msg.OrderHash]
	if !ok {
		log.Warn("not found order")
		r.lpHub.Publish(msg.ID, []byte(WsErrInvalidOrder.Error()))
		return
	}

	order.lpReject <- msg
}
