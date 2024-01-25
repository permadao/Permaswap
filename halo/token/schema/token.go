package schema

type TokenInfo struct {
	Symbol   string `json:"symbol"`
	Decimals int64  `json:"decimals"`

	TotalSupply string                            `json:"totalSupply"`
	Balances    map[string]string                 `json:"balances"` // account id -> balance
	Stakes      map[string]map[string][]StakeInfo `json:"stakes"`   // account id -> stake pool -> stakes
}
