package router

import (
	"fmt"

	"github.com/permadao/permaswap/router/schema"
)

type Pool struct {
	X   string
	Y   string
	Fee string
}

type Config struct {
	Name         string
	Logo         string
	Desc         string
	Domain       string
	Ip           string
	AccountType  string
	KeyFile      string
	Port         string
	Mysql        string
	ChainId      int64  `toml:"chain_id"`
	EverpayApi   string `toml:"everpay_api"`
	NftWhitelist bool   `toml:"nft_whitelist"`
	NftApi       string `toml:"nft_api"`

	FeeRatio     string `toml:"fee_ratio"`
	FeeRecipient string `toml:"fee_recipient"`
	Pools        []Pool
}

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
