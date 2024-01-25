package hvm

import (
	"testing"

	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	h := New(schema.State{})
	t.Log("h:", h, "h.RouterStates:", h.RouterStates)
	assert.Equal(t, h.Dapp, "")
}
