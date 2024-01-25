package schema

import (
	everSchema "github.com/everVision/everpay-kits/schema"
	coreSchema "github.com/permadao/permaswap/core/schema"
)

type InfoRes struct {
	ChainID       int64                        `json:"chainID"`
	RouterAddress string                       `json:"routerAddress"`
	Address       string                       `json:"address"`
	Tokens        map[string]*everSchema.Token `json:"tokens"`
	Pools         map[string]*coreSchema.Pool  `json:"pools"`
	Lps           map[string]coreSchema.Lp     `json:"lps"`
}

type RemoveLpReq struct {
	LpID string `json:"lpID"`
}

type RemoveLpRes struct {
	LpID   string `json:"lpID"`
	Result string `json:"result"` // "ok" or "failed"
	Error  string `json:"error"`
}

type AddLpRes struct {
	LpID   string `json:"lpID"`
	Result string `json:"result"` // "ok" or "failed"
	Error  string `json:"error"`
}
