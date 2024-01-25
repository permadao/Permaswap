package router

import (
	"fmt"

	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/router/schema"
)

const (
	PermaswapClosed = false

	// penalty config
	CumulativeFailures = 3
	ExpirationDuration = 3600
)

func GetLpClientInfoConf(chainID int64) (lpClients map[string]*schema.LpClientInfo) {
	switch chainID {

	case 1:
		lpGo := schema.LpClientInfo{
			Name:    "lp-golang",
			Version: "v0.4.0",
		}
		lpJs := schema.LpClientInfo{
			Name:    "lp-javascript",
			Version: "v0.1.6",
		}
		lpClients = map[string]*schema.LpClientInfo{
			lpGo.Name: &lpGo,
			lpJs.Name: &lpJs,
		}
	case 5:
		lpGo := schema.LpClientInfo{
			Name:    "lp-golang",
			Version: "v0.4.0",
		}
		lpJs := schema.LpClientInfo{
			Name:    "lp-javascript",
			Version: "v0.1.6",
		}
		lpClients = map[string]*schema.LpClientInfo{
			lpGo.Name: &lpGo,
			lpJs.Name: &lpJs,
		}
	default:
		panic(fmt.Sprintf("can not get lp clients conf, invalid chainID: %d\n", chainID))
	}

	return
}

func GetFeeConf(chainID int64) (feeRecepient, feeRatio string) {
	switch chainID {
	case 1:
		feeRecepient = "0xc6B2FcadaEC9FdC6dA8e416B682d4915F85986f6"
		feeRatio = coreSchema.Fee0005
	case 5:
		feeRecepient = "0x41fCE022647de219EBd6dc361016Ff0D63aB3f5D"
		feeRatio = coreSchema.Fee0005
	default:
		panic(fmt.Sprintf("can not get fee conf, invalid chainID: %d\n", chainID))
	}
	return
}

func SetNFTWhiteList(chainID int64) bool {
	switch chainID {
	case 1:
		return false
	case 5:
		return false
	default:
		panic(fmt.Sprintf("SetNFTWhiteList, invalid chainID: %d\n", chainID))
	}
}
