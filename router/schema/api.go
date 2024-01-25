package schema

import (
	"github.com/permadao/permaswap/core/schema"
)

type InfoRes struct {
	ChainID       int64                    `json:"chainID"`
	RouterAddress string                   `json:"routerAddress"`
	NFTWhiteList  bool                     `json:"nftWhiteList"`
	TokenList     []string                 `json:"tokenList"`
	PoolList      map[string]*schema.Pool  `json:"poolList"`
	LpClientInfo  map[string]*LpClientInfo `json:"lpClientInfo"`
}

type OrdersRes struct {
	//Total  int64         `json:"total"`
	Orders []*PermaOrder `json:"orders"`
}

type LpsRes struct {
	Lps []schema.Lp `json:"lps"`
}

type NFTRes struct {
	NFTToHolder  map[string]string   `json:"nftToHolder"`
	HolderToNFTs map[string][]string `json:"holderToNFTs"`
	WhiteList    []string            `json:"whitelist"`
}

type PoolRes struct {
	schema.Pool
	CurrentPriceUP   string      `json:"currentPriceUp"`
	CurrentPriceDown string      `json:"currentPriceDown"`
	Lps              []schema.Lp `json:"lps"`
}

type PoolStatsRes struct {
	PoolID string `json:"poolID"`
	Volume Volume `json:"volume"`
	TVL    TVL    `json:"tvl"`
}

type AccountStatsRes struct {
	Address string          `json:"address"`
	Volumes []Volume        `json:"volumes"`
	Rewards []PermaLpReward `json:"rewards"`
	TVLs    []TVL           `json:"tvls"`
}

type LpRewardsRes struct {
	Address string           `json:"address"`
	LpID    string           `json:"lpID"`
	Rewards []*PermaLpReward `json:"rewards"`
}

type PenaltyRes struct {
	ExpirationDuration int64                      `json:"expirationDuration"`
	CumulativeFailures int64                      `json:"cumulativeFailures"`
	FailureRecords     map[string][]FailureRecord `json:"failureRecords"`
	BlackList          map[string]int64           `json:"blackList"`
}
