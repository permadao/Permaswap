package schema

// token on everpay
type EverToken struct {
	ID        string `json:"id"`  // On Native-Chain tokenId
	Tag       string `json:"tag"` // On everPay tokenTag
	Symbol    string `json:"symbol"`
	Decimals  int    `json:"decimals"`  //On everPay decimals
	ChainType string `json:"chainType"` // On everPay chainType; tns102 type is everpay
	ChainID   string `json:"chainID"`   // On everPay chainId
}
