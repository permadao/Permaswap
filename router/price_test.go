package router

import "testing"

func TestGetTokenPriceByRedstone(t *testing.T) {
	price, err := GetTokenPriceByRedstone("AR", "USDC", "")
	t.Log("AR", price, err)
	price, err = GetTokenPriceByRedstone("U", "USDC", "")
	t.Log("U", price, err)
	price, err = GetTokenPriceByRedstone("ARDRIVE", "USDC", "")
	t.Log("ARDRIVE", price, err)
	price, err = GetTokenPriceByRedstone("ACNH", "USDC", "")
	t.Log("ACNH", price, err)
}
