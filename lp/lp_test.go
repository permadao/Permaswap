package lp

import (
	"testing"
	"time"

	"github.com/everFinance/goether"
	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/permadao/permaswap/router"
	"github.com/stretchr/testify/assert"
)

const (
	testPort          = ":8080"
	testRouterURL     = "ws://localhost" + testPort + "/wslp"
	testRouterHttpURL = "http://localhost" + testPort
)

func TestLpAdd(t *testing.T) {
	testRouter := router.New(5, nil, "", "", true, "")
	testRouter.Run(testPort, "")
	defer testRouter.Close()

	signer, err := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	assert.NoError(t, err)
	everSDK, err := sdk.New(signer, "https://api-dev.everpay.io")
	assert.NoError(t, err)

	testSDK := NewRSDK(testRouterURL, testRouterHttpURL, everSDK)

	testLp := New(5, false, testSDK)
	testLp.Run("./test.json")
	time.Sleep(100 * time.Millisecond)
	// test lp core
	lps := testLp.core.GetLps(testLp.rsdk.AccID)
	assert.Equal(t, "0x7bd8bbec75143287a3ac339d7f3235f130dd8e779663cde432558852d6d33d80", lps[0].PoolID)
	// test router core
	lps, err = testLp.rsdk.GetLps()
	assert.NoError(t, err)
	assert.Equal(t, "ethereum-eth-0x0000000000000000000000000000000000000000", lps[0].TokenXTag)
}

func TestTxs(t *testing.T) {
	client := sdk.NewClient("https://api.everpay.io")
	txs, err := client.Txs(0, "desc", 1, everSchema.TxOpts{
		Address: "0xd110107adb30bce6c0646eaf77cc1c815012331d",
	})
	assert.NoError(t, err)
	t.Log("latestTxRawId:", txs.Txs[0].RawId)
}
