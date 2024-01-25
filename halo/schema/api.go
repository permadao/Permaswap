package schema

import (
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
)

type InfoRes struct {
	hvmSchema.State
	GenesisTxEverHash string `json:"genesisTxEverHash"`
	HaloAddr          string `json:"haloAddr"`
}
