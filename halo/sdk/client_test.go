package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testClient *Client

func init() {
	testClient = NewClient("https://router-dev.permaswap.network/halo")
}

func TestGetInfo(t *testing.T) {
	info, err := testClient.GetInfo()
	assert.NoError(t, err)
	assert.Equal(t, "1", info.ChainID)
	assert.Equal(t, "halo", info.Dapp)
}

func TestClient_GetTx(t *testing.T) {
	haloHash := "0xce5bfe2732bd58f401b2e98041591b4be76123621fc1e84b4795cc41162dbfe5"
	_, err := testClient.GetTx(haloHash)
	assert.NoError(t, err)
}
