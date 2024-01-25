package lp

import (
	"encoding/json"
	"errors"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"

	"github.com/everVision/everpay-kits/sdk"
	"github.com/gorilla/websocket"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/router/schema"
	"gopkg.in/h2non/gentleman.v2"
)

// RSDK is RouterSDK
type RSDK struct {
	AccID   string
	wsURL   string
	wsConn  *websocket.Conn
	httpCli *gentleman.Client
	EverSDK *sdk.SDK

	order                        chan *schema.LpMsgOrder
	orderStatus                  chan *schema.OrderMsgStatus
	addResponse                  chan *schema.LpMsgAddResponse
	addResponseOnceSubscribed    bool
	removeResponse               chan *schema.LpMsgRemoveResponse
	removeResponseOnceSubscribed bool

	reconnect chan struct{}
}

func NewRSDK(wsURL, httpURL string, everSDK *sdk.SDK) *RSDK {
	r := &RSDK{
		AccID:   everSDK.AccId,
		wsURL:   wsURL,
		httpCli: gentleman.New().URL(httpURL),
		EverSDK: everSDK,

		order:       make(chan *schema.LpMsgOrder),
		orderStatus: make(chan *schema.OrderMsgStatus),

		addResponse:                  make(chan *schema.LpMsgAddResponse),
		addResponseOnceSubscribed:    false,
		removeResponse:               make(chan *schema.LpMsgRemoveResponse),
		removeResponseOnceSubscribed: false,

		reconnect: make(chan struct{}),
	}

	err := r.connectRouter()
	if err != nil {
		panic(err)
	}

	go r.runMsgUnmarshal()

	return r
}

func (r *RSDK) Close() {
	r.wsConn.Close()
	log.Info("rsdk websocket closed")
}

func (r *RSDK) SubscribeOrder() <-chan *schema.LpMsgOrder {
	return r.order
}

func (r *RSDK) SubscribeOrderStatus() <-chan *schema.OrderMsgStatus {
	return r.orderStatus
}

func (r *RSDK) SubscribeReconnect() <-chan struct{} {
	return r.reconnect
}

func (r *RSDK) SubscribeLpAddResponseOnce() <-chan *schema.LpMsgAddResponse {
	r.addResponseOnceSubscribed = true
	return r.addResponse
}

func (r *RSDK) SubscribeLpRemoveResponseOnce() <-chan *schema.LpMsgRemoveResponse {
	r.removeResponseOnceSubscribed = true
	return r.removeResponse
}

func (r *RSDK) AddLiquidity(msg schema.LpMsgAdd) error {
	return r.wsConn.WriteMessage(websocket.TextMessage, msg.Marshal())
}

func (r *RSDK) RemoveLiquidity(msg schema.LpMsgRemove) error {
	return r.wsConn.WriteMessage(websocket.TextMessage, msg.Marshal())
}

func (r *RSDK) GetInfo() (info schema.InfoRes, err error) {
	req := r.httpCli.Request()
	req.Path("/info")

	res, err := req.Send()
	if err != nil {
		return
	}
	defer res.Close()

	err = json.Unmarshal(res.Bytes(), &info)
	return
}

func (r *RSDK) GetLps() (lps []coreSchema.Lp, err error) {
	req := r.httpCli.Request()
	req.Path("/lps")
	req.AddQuery("accid", r.AccID)

	res, err := req.Send()
	if err != nil {
		return
	}
	defer res.Close()

	lpsRes := schema.LpsRes{}
	if err = json.Unmarshal(res.Bytes(), &lpsRes); err != nil {
		return
	}
	lps = lpsRes.Lps
	return
}

func (r *RSDK) SignOrder(msg schema.LpMsgOrder) error {
	sig, err := r.EverSDK.Sign(msg.Bundle.String())
	if err != nil {
		return err
	}

	return r.wsConn.WriteMessage(websocket.TextMessage, schema.LpMsgSign{
		Event:   schema.LpMsgEventSign,
		Address: r.EverSDK.AccId,
		Bundle: everSchema.BundleWithSigs{
			Bundle: msg.Bundle,
			Sigs: map[string]string{
				r.EverSDK.AccId: sig,
			},
		},
	}.Marshal())
}

func (r *RSDK) RejectOrder(msg schema.LpMsgOrder) error {
	return r.wsConn.WriteMessage(websocket.TextMessage, schema.LpMsgReject{
		Event:     schema.LpMsgEventReject,
		Address:   r.EverSDK.AccId,
		OrderHash: msg.Bundle.HashHex(),
	}.Marshal())
}

func (r *RSDK) GetBalance(tokenTag string) (amount string, err error) {
	b, err := r.EverSDK.Cli.Balance(tokenTag, r.EverSDK.AccId)
	if err != nil {
		return
	}
	return b.Balance.Amount, nil
}

func (r *RSDK) runMsgUnmarshal() {
	for {
		_, data, err := r.wsConn.ReadMessage()
		if err != nil {
			log.Warn("connection disconnected", "err", err)
			time.Sleep(2 * time.Second)

			log.Info("reconnect...")
			if err = r.connectRouter(); err != nil {
				log.Error("reconnect failed", "err", err)
				r.wsConn.Close()
			} else {
				log.Info("reconnect success")
				r.reconnect <- struct{}{}
			}
			continue
		}

		msg := &schema.LpMsg{}
		if err = json.Unmarshal(data, msg); err != nil {
			log.Error("invalid msg from router", "err", err, "msg", string(data))
			continue
		}

		switch msg.Event {
		case schema.LpMsgEventOrder:
			orderMsg := &schema.LpMsgOrder{}
			if err = json.Unmarshal(data, orderMsg); err != nil {
				log.Error("invalid order from router", "err", err, "msg", string(data))
				continue
			}
			r.order <- orderMsg

		case schema.OrderMsgEventStatus:
			statusMsg := &schema.OrderMsgStatus{}
			if err = json.Unmarshal(data, statusMsg); err != nil {
				log.Error("invalid status from router", "err", err, "msg", string(data))
				continue
			}
			r.orderStatus <- statusMsg

		case schema.LpMsgEventAddResponse:
			addResponseMsg := &schema.LpMsgAddResponse{}
			if err = json.Unmarshal(data, addResponseMsg); err != nil {
				log.Error("invalid lp add response from router", "err", err, "msg", string(data))
				addResponseMsg = &schema.LpMsgAddResponse{
					Event: schema.LpMsgEventAddResponse,
					Msg:   "failed",
					Error: "error_invalid_response",
				}
			}
			if r.addResponseOnceSubscribed {
				r.addResponseOnceSubscribed = false
				r.addResponse <- addResponseMsg
			}

		case schema.LpMsgEventRemoveResponse:
			removeResponseMsg := &schema.LpMsgRemoveResponse{}
			if err = json.Unmarshal(data, removeResponseMsg); err != nil {
				log.Error("invalid lp add response from router", "err", err, "msg", string(data))
				removeResponseMsg = &schema.LpMsgRemoveResponse{
					Event: schema.LpMsgEventRemoveResponse,
					Msg:   "failed",
					Error: "error_invalid_response",
				}
			}
			if r.removeResponseOnceSubscribed {
				r.removeResponseOnceSubscribed = false
				r.removeResponse <- removeResponseMsg
			}

		default:
			log.Error("invalid message event", "msg", string(data))
		}
	}
}

func (r *RSDK) connectRouter() (err error) {
	wsConn, _, err := websocket.DefaultDialer.Dial(r.wsURL, nil)
	if err != nil {
		return
	}
	r.wsConn = wsConn

	// auto register
	_, msg, err := r.wsConn.ReadMessage()
	if err != nil {
		return
	}
	saltMsg := schema.LpMsgSalt{}
	if err = json.Unmarshal(msg, &saltMsg); err != nil {
		return
	}
	salt := saltMsg.Salt
	sig, err := r.EverSDK.Sign(salt)
	if err != nil {
		return
	}
	err = r.wsConn.WriteMessage(websocket.TextMessage, schema.LpMsgRegister{
		Address:         r.EverSDK.AccId,
		Sig:             sig,
		LpClientName:    LpName,
		LpClientVersion: LpVersion,
	}.Marshal())
	if err != nil {
		return
	}
	_, msg, err = r.wsConn.ReadMessage()
	if err != nil {
		return
	}
	if string(msg) != `{"event":"response","msg":"ok"}` {
		err = errors.New(string(msg))
	}

	return
}
