package schema

import (
	"encoding/json"
	"time"
)

const (
	OrderStatusPending = "pending"
	OrderStatusSuccess = "success"
	OrderStatusFailed  = "failed"
	OrderStatusExpired = "expired"

	OrderMsgEventStatus = "status"

	OrderExpire = 10 * time.Second // order life cycle
)

type OrderMsgStatus struct {
	Event     string `json:"event" default:"status"`
	OrderHash string `json:"orderHash"`
	EverHash  string `json:"everHash"`
	Status    string `json:"status"`
}

func (o OrderMsgStatus) Marshal() []byte {
	o.Event = OrderMsgEventStatus
	by, _ := json.Marshal(o)
	return by
}
