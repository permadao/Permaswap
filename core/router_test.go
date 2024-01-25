package core

import (
	"testing"

	"github.com/permadao/permaswap/core/schema"
	"github.com/stretchr/testify/assert"
)

func TestFindPoolPaths(t *testing.T) {
	idToPool := map[string]*schema.Pool{
		"1": &schema.Pool{"eth", "usdt", nil, nil},
		"2": &schema.Pool{"usdc", "usdt", nil, nil},
		"3": &schema.Pool{"eth", "wbtc", nil, nil},
		"4": &schema.Pool{"usdt", "wbtc", nil, nil},
	}
	tokenTagToPoolIDs := map[string][]string{
		"eth":  []string{"1", "3"},
		"usdt": []string{"1", "2", "4"},
		"usdc": []string{"2"},
		"wbtc": []string{"3", "4"},
	}

	paths, err := findPoolPaths(idToPool, tokenTagToPoolIDs, "usdc", "eth", nil)
	assert.NoError(t, err)
	t.Log(len(paths))
	for _, path := range paths {
		t.Log(path)
	}

	paths, err = findPoolPaths(idToPool, tokenTagToPoolIDs, "usdc", "ar", nil)
	assert.NoError(t, err)
	t.Log(len(paths))
	for _, path := range paths {
		t.Log(path)
	}

	paths, err = findPoolPaths(idToPool, tokenTagToPoolIDs, "eth", "wbtc", nil)
	assert.NoError(t, err)
	t.Log(len(paths))
	for _, path := range paths {
		t.Log(path)
	}

	paths, err = findPoolPaths(idToPool, tokenTagToPoolIDs, "eth", "eth", nil)
	assert.EqualError(t, err, "err_invalid_token")
}

func TestFindPoolPaths2(t *testing.T) {
	idToPool := map[string]*schema.Pool{
		"1": &schema.Pool{"eth", "usdc", nil, nil},
		"2": &schema.Pool{"usdc", "usdt", nil, nil},
		"3": &schema.Pool{"ar", "eth", nil, nil},
		"4": &schema.Pool{"ar", "usdc", nil, nil},
	}

	tokenTagToPoolIDs := map[string][]string{
		"eth":  []string{"1", "3"},
		"usdc": []string{"1", "2", "4"},
		"usdt": []string{"2"},
		"ar":   []string{"3", "4"},
	}

	tradeList := [][]string{
		[]string{"usdc", "eth"},
		[]string{"usdc", "ar"},
		[]string{"ar", "usdt"},
		[]string{"eth", "wbtc"},
		[]string{"usdc", "usdt"},
		[]string{"eth", "ar"},
		[]string{"usdt", "eth"},
		[]string{"usdt", "ar"},
	}
	for _, trade := range tradeList {
		tokenIn := trade[0]
		tokenOut := trade[1]
		paths, err := findPoolPaths(idToPool, tokenTagToPoolIDs, tokenIn, tokenOut, nil)
		assert.NoError(t, err)
		t.Log("tokenIn:", tokenIn, "tokenOut:", tokenOut)
		t.Log(len(paths))
		for _, path := range paths {
			t.Log("path:", path)
		}
		t.Log("\n")
	}
}
