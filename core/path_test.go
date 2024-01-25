package core

import (
	"testing"

	"github.com/permadao/permaswap/core/schema"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	a := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	b := "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"
	c := "0x41fCE022647de219EBd6dc361016Ff0D63aB3f5D"
	invalidPaths := []schema.Path{
		{"1", a, b, "eth", "1"},
		{"2", b, a, "usdt", "3000"},
		{"3", b, a, "usdt", "6000"},
		{"4", a, "d", "usdc", "3"},
		{"5", "c", a, "usdt", "9000"},
	}
	_, err := PathsToSwapInputs(a, invalidPaths)
	assert.EqualError(t, err, "err_invalid_path")

	paths1 := []schema.Path{
		{"0x1", a, b, "eth", "1"},
		{"0x1", b, a, "usdt", "3000"},
		{"0x2", b, a, "usdt", "6000"},
		{"0x2", a, b, "usdc", "3"},
		{"", a, c, "usdc", "3"},
	}
	sis, err := PathsToSwapInputs(b, paths1)
	assert.NoError(t, err)
	for id, si := range sis {
		t.Log(id, si)
	}
	assert.Equal(t, len(paths1), 5)
}
