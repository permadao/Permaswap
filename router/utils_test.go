package router

import (
	"math/big"
	"testing"
	"time"

	"github.com/permadao/permaswap/core/schema"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var testClient *sdk.Client

func init() {
	testClient = sdk.NewClient("https://api-dev.everpay.io")
}

func TestConvertPathsToBundle(t *testing.T) {
	tokens, err := testClient.GetTokens()
	assert.NoError(t, err)
	t.Log(len(tokens), tokens)

	user := "0x45a9f23ac5af2dcdcb11f1386977d2bfea7dad5b"
	lp := "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	lp2 := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	eth := "ethereum-eth-0x0000000000000000000000000000000000000000"
	usdc := "ethereum-usdc-0xf044320bcc3cd1f6100cd197754c71941469e79c"
	usdt := "ethereum-usdt-0x923fcb255da521037385457fb549a51f78ef0af4"

	paths := []schema.Path{
		{"1", user, lp, eth, big.NewInt(1 * 1000000000000000000).String()},
		{"2", lp, user, usdt, big.NewInt(3000 * 1000000).String()},
		{"3", user, lp, eth, big.NewInt(1000000000000000000 / 10).String()},
		{"4", lp, user, usdt, big.NewInt(300 * 1000000).String()},
		{"5", lp, user, usdc, big.NewInt(300 * 1000000).String()},
		{"6", user, lp, usdc, big.NewInt(300 * 1000000).String()},
		{"7", user, lp2, usdc, big.NewInt(300 * 1000000).String()},
		{"8", lp2, user, usdt, big.NewInt(300 * 1000000).String()},
	}

	expireAt := time.Now().Unix() + 120
	salt := uuid.NewString()
	b, err := ConvertPathsToBundle(paths, tokens, expireAt, salt)
	assert.NoError(t, err)

	err = VerifyBundleAndPaths(b, paths, tokens)
	assert.NoError(t, err)
}
