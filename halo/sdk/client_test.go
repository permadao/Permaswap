package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testClient *Client

func init() {
	testClient = NewClient("https://node.halonode.com/")
}

func TestGetInfo(t *testing.T) {
	info, err := testClient.GetInfo()
	assert.NoError(t, err)
	assert.Equal(t, "1", info.ChainID)
	assert.Equal(t, "halo", info.Dapp)
}
