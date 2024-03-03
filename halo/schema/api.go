package schema

import (
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
)

type InfoRes struct {
	hvmSchema.State
	GenesisTxEverHash string `json:"genesisTxEverHash"`
	HaloAddr          string `json:"haloAddr"`
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
