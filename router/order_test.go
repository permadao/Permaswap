package router

import (
	"encoding/json"
	"testing"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/stretchr/testify/assert"
)

func TestVerifyBundleByAddr(t *testing.T) {
	data := `
	{
		"items": [
			{
				"tag": "ethereum-eth-0x0000000000000000000000000000000000000000",
				"chainID": "5",
				"from": "0x911f42b0229c15bbb38d648b7aa7ca480ed977d6",
				"to": "0x61ebf673c200646236b2c53465bca0699455d5fa",
				"amount": "100000000000000000"
			},
			{
				"tag": "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
				"chainID": "5",
				"from": "0x61ebf673c200646236b2c53465bca0699455d5fa",
				"to": "0x911f42b0229c15bbb38d648b7aa7ca480ed977d6",
				"amount": "296147410"
			}
		],
		"expiration": 1645336839,
		"salt": "af2b2d0a-d979-4d15-90d2-de7d7fc0bbd9",
		"version": "v1",
		"sigs": {
			"0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6": "0xb231b845a843c6e2f813739e00519dc069ae8f2240d0707ab2ae82a5ed46a1242267759f4f4cc82d302584895de0249405fe3f877d596b39fc370aad2c65a59a1c",
			"0x61ebf673c200646236b2c53465bca0699455d5fa": "0x112fe6e8981b2f2592a19b18fa99e4ebefbe6d8eee9cc27a1743600439962f1e18946f178717235173e146f037a9ea8efd3b5953d092fb63331d4024086898e41c"
		}
	}`

	var bundle everSchema.BundleWithSigs
	err := json.Unmarshal([]byte(data), &bundle)
	assert.NoError(t, err)

	accid, sig, err := VerifyBundleByAddr(bundle, "0x61ebf673c200646236b2c53465bca0699455d5fa", 5)
	assert.NoError(t, err)
	assert.Equal(t, "0x61EbF673c200646236B2c53465bcA0699455d5FA", accid)
	assert.Equal(t, "0x112fe6e8981b2f2592a19b18fa99e4ebefbe6d8eee9cc27a1743600439962f1e18946f178717235173e146f037a9ea8efd3b5953d092fb63331d4024086898e41c", sig["0x61ebf673c200646236b2c53465bca0699455d5fa"])

	accid, sig, err = VerifyBundleByAddr(bundle, "0x911f42b0229c15bbb38d648b7aa7ca480ed977d6", 5)
	assert.NoError(t, err)
	assert.Equal(t, "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6", accid)
	assert.Equal(t, "0xb231b845a843c6e2f813739e00519dc069ae8f2240d0707ab2ae82a5ed46a1242267759f4f4cc82d302584895de0249405fe3f877d596b39fc370aad2c65a59a1c", sig["0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"])
}
