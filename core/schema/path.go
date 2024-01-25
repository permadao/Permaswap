package schema

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
)

type Path struct {
	LpID     string `json:"lpId"`
	From     string `json:"from"`
	To       string `json:"to"`
	TokenTag string `json:"tokenTag"`
	Amount   string `json:"amount"`
}

type SwapOutput struct {
	LpID           string
	TokenIn        string
	AmountIn       *big.Int
	TokenOut       string
	AmountOut      *big.Int
	Fee            *big.Int
	StartSqrtPrice *apd.Decimal
	EndSqrtPrice   *apd.Decimal
	IsDryRun       bool
}

type SwapInput struct {
	LpID      string
	TokenIn   string
	AmountIn  *big.Int
	TokenOut  string
	AmountOut *big.Int
}
