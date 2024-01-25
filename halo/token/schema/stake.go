package schema

import "math/big"

type Stake struct {
	StakedAt int64    `json:"stakedAt"`
	Amount   *big.Int `json:"amount"`
}

type StakeInfo struct {
	StakedAt int64  `json:"stakedAt"`
	Amount   string `json:"amount"`
}
