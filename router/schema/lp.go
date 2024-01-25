package schema

import (
	"encoding/json"

	apd "github.com/cockroachdb/apd/v3"
	coreSchema "github.com/permadao/permaswap/core/schema"
	everSchema "github.com/everVision/everpay-kits/schema"
)

type LpClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

const (
	// msg from lp
	LpMsgEventRegister = "register"
	LpMsgEventAdd      = "add"
	LpMsgEventRemove   = "remove"
	LpMsgEventSign     = "sign"
	LpMsgEventReject   = "reject"

	// msg to lp
	LpMsgEventResponse       = "response"
	LpMsgEventSalt           = "salt" // salt for register
	LpMsgEventOrder          = "order"
	LpMsgEventAddResponse    = "addResponse"    // response to add lp
	LpMsgEventRemoveResponse = "removeResponse" // response to remove lp
	// notice ordre status msg in order.go
)

var (
	// response msg
	LpMsgOk = LpMsgResponse{Event: LpMsgEventResponse, Msg: "ok"}
)

type LpMsg struct {
	Event string `json:"event"`
}

type LpMsgRegister struct {
	ID              string `json:"id"`
	Event           string `json:"event"`
	Address         string `json:"address"`
	Sig             string `json:"sig"`
	LpClientName    string `json:"lpClientName"`
	LpClientVersion string `json:"lpClientVerison"`
}

func (l LpMsgRegister) Marshal() []byte {
	l.Event = LpMsgEventRegister
	by, _ := json.Marshal(l)
	return by
}

type LpMsgAdd struct {
	ID    string `json:"id"`
	Event string `json:"event"`

	TokenX   string       `json:"tokenX"`
	TokenY   string       `json:"tokenY"`
	FeeRatio *apd.Decimal `json:"feeRatio"`

	CurrentSqrtPrice *apd.Decimal `json:"currentSqrtPrice"`
	LowSqrtPrice     *apd.Decimal `json:"lowSqrtPrice"`
	HighSqrtPrice    *apd.Decimal `json:"highSqrtPrice"`
	Liquidity        string       `json:"liquidity"`
	PriceDirection   string       `json:"priceDirection"`
}

func (l LpMsgAdd) Marshal() []byte {
	l.Event = LpMsgEventAdd
	by, _ := json.Marshal(l)
	return by
}

type LpMsgRemove struct {
	ID    string `json:"id"`
	Event string `json:"event"`

	TokenX   string       `json:"tokenX"`
	TokenY   string       `json:"tokenY"`
	FeeRatio *apd.Decimal `json:"feeRatio"`

	LowSqrtPrice   *apd.Decimal `json:"lowSqrtPrice"`
	HighSqrtPrice  *apd.Decimal `json:"highSqrtPrice"`
	PriceDirection string       `json:"priceDirection"`
}

func (l LpMsgRemove) Marshal() []byte {
	l.Event = LpMsgEventRemove
	by, _ := json.Marshal(l)
	return by
}

type LpMsgSign struct {
	ID      string                    `json:"id"`
	Event   string                    `json:"event"`
	Address string                    `json:"address"`
	Bundle  everSchema.BundleWithSigs `json:"bundle"`
}

func (l LpMsgSign) Marshal() []byte {
	l.Event = LpMsgEventSign
	by, _ := json.Marshal(l)
	return by
}

type LpMsgReject struct {
	ID        string `json:"id"`
	Event     string `json:"event"`
	Address   string `json:"address"`
	OrderHash string `json:"orderHash"`
}

func (l LpMsgReject) Marshal() []byte {
	l.Event = LpMsgEventReject
	by, _ := json.Marshal(l)
	return by
}

type LpMsgResponse struct {
	Event string `json:"event"`
	Msg   string `json:"msg"`
}

func (l LpMsgResponse) Marshal() []byte {
	l.Event = LpMsgEventResponse
	by, _ := json.Marshal(l)
	return by
}

type LpMsgSalt struct {
	Event string `json:"event"`
	Salt  string `json:"salt"`
}

func (l LpMsgSalt) Marshal() []byte {
	l.Event = LpMsgEventSalt
	by, _ := json.Marshal(l)
	return by
}

type LpMsgOrder struct {
	Event    string            `json:"event"`
	UserAddr string            `json:"userAddr"`
	Bundle   everSchema.Bundle `json:"bundle"`
	Paths    []coreSchema.Path `json:"paths"`
}

func (l LpMsgOrder) Marshal() []byte {
	l.Event = LpMsgEventOrder
	by, _ := json.Marshal(l)
	return by
}

type LpMsgAddResponse struct {
	Event string `json:"event"`
	LpID  string `json:"lpID"`
	Msg   string `json:"msg"` // "ok" or "failed"
	Error string `json:"error"`
}

func (l LpMsgAddResponse) Marshal() []byte {
	l.Event = LpMsgEventAddResponse
	by, _ := json.Marshal(l)
	return by
}

type LpMsgRemoveResponse struct {
	Event string `json:"event"`
	LpID  string `json:"lpID"`
	Msg   string `json:"msg"` // "ok" or "failed"
	Error string `json:"error"`
}

func (l LpMsgRemoveResponse) Marshal() []byte {
	l.Event = LpMsgEventRemoveResponse
	by, _ := json.Marshal(l)
	return by
}
