package schema

import (
	"encoding/json"

	coreSchema "github.com/permadao/permaswap/core/schema"
	everSchema "github.com/everVision/everpay-kits/schema"
)

const (
	// msg from user
	UserMsgEventQuery  = "query"
	UserMsgEventSubmit = "submit"

	// msg to user
	UserMsgEventResponse = "response"
	UserMsgEventOrder    = "order"

	// notic order status msg in order.go
)

type UserMsg struct {
	Event string `json:"event"`
}

// {"event":"query","address":"123","tokenIn":"ethereum-eth-0x0000000000000000000000000000000000000000","tokenOut":"ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede", "amountIn":"4000000000000000"}
type UserMsgQuery struct {
	ID       string `json:"id"`
	Event    string `json:"event"`
	Address  string `json:"address"`
	TokenIn  string `json:"tokenIn"`
	TokenOut string `json:"tokenOut"`
	AmountIn string `json:"amountIn"`
}

func (l UserMsgQuery) Marshal() []byte {
	l.Event = UserMsgEventQuery
	by, _ := json.Marshal(l)
	return by
}

type UserMsgSubmit struct {
	ID       string                    `json:"id"`
	Event    string                    `json:"event"`
	Address  string                    `json:"address"`
	TokenIn  string                    `json:"tokenIn"`
	TokenOut string                    `json:"tokenOut"`
	Bundle   everSchema.BundleWithSigs `json:"bundle"`
	Paths    []coreSchema.Path         `json:"paths"`
}

func (l UserMsgSubmit) Marshal() []byte {
	l.Event = UserMsgEventSubmit
	by, _ := json.Marshal(l)
	return by
}

type UserMsgResponse struct {
	Event string `json:"event"`
	Msg   string `json:"msg"`
}

func (l UserMsgResponse) Marshal() []byte {
	l.Event = UserMsgEventResponse
	by, _ := json.Marshal(l)
	return by
}

type UserMsgOrder struct {
	Event       string            `json:"event" default:"order"`
	UserAddr    string            `json:"userAddr"`
	TokenIn     string            `json:"tokenIn"`
	TokenOut    string            `json:"tokenOut"`
	Price       string            `json:"price"`
	PriceImpact string            `json:"priceImpact"`
	Bundle      everSchema.Bundle `json:"bundle"`
	Paths       []coreSchema.Path `json:"paths"`
}

func (u UserMsgOrder) Marshal() []byte {
	u.Event = UserMsgEventOrder
	by, _ := json.Marshal(u)
	return by
}
