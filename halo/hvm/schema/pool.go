package schema

import (
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	Fee01   = "0.01"
	Fee003  = "0.003"
	Fee001  = "0.001"
	Fee0005 = "0.0005"
)

// pool on permaswap
type Pool struct {
	TokenXTag string `json:"tokenXTag"`
	TokenYTag string `json:"tokenYTag"`
	FeeRatio  string `json:"feeRatio"`
}

func (pool *Pool) String() string {
	return "TokenXTag:" + pool.TokenXTag + "\n" +
		"TokenYTag:" + pool.TokenYTag + "\n" +
		"FeeRatio:" + pool.FeeRatio
}

func (pool *Pool) ID() string {
	h := accounts.TextHash([]byte(pool.String()))
	return hexutil.Encode(h)
}
