package schema

import (
	"github.com/permadao/permaswap/halo/hvm/schema"
)

const (
	EverTxActionTransfer = "transfer"
	EverTxActionBundle   = "bundle"
)

type GenesisTxData struct {
	Dapp             string                         `json:"dapp"`
	ChainID          string                         `json:"chainID"`
	Govern           string                         `json:"govern"` // govern is an temporary solution and will use voting in the future
	FeeRecipient     string                         `json:"feeRecipient"`
	RouterMinStake   string                         `json:"routerMinStake"`
	Routers          []string                       `json:"routers"`
	RouterStates     map[string]*schema.RouterState `json:"routerStates"`
	StakePools       []string                       `json:"stakePools"`
	OnlyUnStakePools []string                       `json:"onlyUnStakePools"`
	TokenSymbol      string                         `json:"tokenSymbol"`
	TokenDecimals    int64                          `json:"tokenDecimals"`
	TokenTotalSupply string                         `json:"tokenTotalSupply"`
	TokenBalance     map[string]string              `json:"tokenBalance"`
	TokenStake       map[string]map[string]string   `json:"tokenStake"`
}

type TxApply struct {
	Tx     schema.Transaction `json:"tx"`
	DryRun bool               `json:"dryRun"`
}
