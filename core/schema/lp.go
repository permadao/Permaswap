package schema

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// https://github.com/Uniswap/v3-core/blob/main/contracts/libraries/TickMath.sol#L9
const (
	FullRangeLowSqrtPrice  = "5.4210108624275221700372640043498E-20" //2**-64
	FullRangeHighSqrtPrice = "18446744073709551616"                  //2**64
)

const (
	MinSqrtPriceFactor = "1.00005"
)

const (
	PriceDirectionUp   = "up"
	PriceDirectionDown = "down"
	PriceDirectionBoth = "both"
)

type Lp struct {
	PoolID    string       `json:"poolID"`
	TokenXTag string       `json:"tokenX"`
	TokenYTag string       `json:"tokenY"`
	FeeRatio  *apd.Decimal `json:"feeRatio"`
	AccID     string       `json:"accID"`

	LowSqrtPrice     *apd.Decimal `json:"lowSqrtPrice"`
	CurrentSqrtPrice *apd.Decimal `json:"currentSqrtPrice"`
	HighSqrtPrice    *apd.Decimal `json:"highSqrtPrice"`

	Liquidity      *big.Int `json:"liquidity"`
	PriceDirection string   `json:"priceDirection"`
}

func (lp *Lp) String() string {
	return "PoolID:" + lp.PoolID + "\n" +
		"Address:" + lp.AccID + "\n" +
		"LowSqrtPrice:" + lp.LowSqrtPrice.Text('f') + "\n" +
		"HighSqrtPrice:" + lp.HighSqrtPrice.Text('f') + "\n" +
		"PriceDirection:" + lp.PriceDirection
}

func (lp *Lp) ID() string {
	h := accounts.TextHash([]byte(lp.String()))
	return hexutil.Encode(h)
}
