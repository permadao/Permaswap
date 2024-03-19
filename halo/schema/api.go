package schema

import (
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/token/schema"
)

type InfoRes struct {
	hvmSchema.State
	GenesisTxEverHash string `json:"genesisTxEverHash"`
	HaloAddr          string `json:"haloAddr"`
}

type BalanceRes struct {
	Balance string                        `json:"balance"`
	Stakes  map[string][]schema.StakeInfo `json:"stakes"`
}

type TxWithValidity struct {
	Tx       string `json:"tx"`
	Validity bool   `json:"validity"`
}
type TxRes struct {
	Executed []TxWithValidity `json:"executed"` // executed tx
}

type SubmitRes struct {
	EverHash string `json:"everHash"`
}
