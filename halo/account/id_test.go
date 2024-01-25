package account

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDCheck(t *testing.T) {
	accType, arID, err := IDCheck("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	assert.NoError(t, err)
	assert.Equal(t, "arweave", accType)
	t.Log(arID)
}
