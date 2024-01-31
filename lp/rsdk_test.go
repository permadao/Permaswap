package lp

import (
	"testing"

	"github.com/everFinance/goether"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/gorilla/websocket"
	"github.com/permadao/permaswap/router"
	"github.com/stretchr/testify/assert"
)

func TestSDKConnect(t *testing.T) {
	testRouter := router.New("", "", 5, nil, "", "", "", nil, false)
	testRouter.Run(testPort, "")
	defer testRouter.Close()

	signer, err := goether.NewSigner("1a7ffbdae668acf43251ed8913596f7db0ce0f90bcd27d4aa85b2bd8a3d0c550")
	assert.NoError(t, err)
	everSDK, err := sdk.New(signer, "https://api-dev.everpay.io")
	assert.NoError(t, err)

	testSDK := NewRSDK(testRouterURL, testRouterHttpURL, everSDK)
	err = testSDK.wsConn.WriteMessage(websocket.TextMessage, []byte{})
	assert.NoError(t, err)
}
