package core

import (
	"github.com/permadao/permaswap/core/schema"
)

func (c *Core) FindPoolPaths(tokenIn, tokenOut string) ([][]*schema.Pool, error) {
	poolPaths, _ := findPoolPaths(c.Pools, c.TokenTagToPoolIDs, tokenIn, tokenOut, []string{})
	if poolPaths == nil || len(poolPaths) == 0 {
		return nil, ERR_NO_POOL
	}

	result := [][]*schema.Pool{}

	for _, path := range poolPaths {
		if len(path) > c.MaxPoolPathLength {
			continue
		}
		pools := []*schema.Pool{}
		for _, poolID := range path {
			pool, _ := c.Pools[poolID]
			pools = append(pools, pool)
		}
		result = append(result, pools)
	}

	return result, nil
}

func findPoolPaths(idToPool map[string]*schema.Pool, tokenTagToPoolIDs map[string][]string,
	tokenIn, tokenOut string, prevPath []string) ([][]string, error) {

	if tokenIn == tokenOut {
		return nil, ERR_INVALID_TOKEN
	}

	paths := [][]string{}

	poolIDs, ok := tokenTagToPoolIDs[tokenIn]
	if !ok {
		return nil, ERR_NO_POOL
	}

	poolIDs2 := []string{}
	for _, id := range poolIDs {
		isInPrevPools := false

		for _, id2 := range prevPath {
			if id == id2 {
				isInPrevPools = true
				break
			}
		}

		if !isInPrevPools {
			poolIDs2 = append(poolIDs2, id)
		}
	}

	if len(poolIDs2) == 0 {
		return nil, ERR_NO_POOL
	}

	for _, id := range poolIDs2 {
		path := append(prevPath, id)
		pool, ok := idToPool[id]
		if !ok {
			continue
		}

		if (pool.TokenXTag == tokenIn && pool.TokenYTag == tokenOut) ||
			(pool.TokenYTag == tokenIn && pool.TokenXTag == tokenOut) {
			paths = append(paths, path)
			continue
		} else {
			tokenIn2 := pool.TokenYTag
			if pool.TokenYTag == tokenIn {
				tokenIn2 = pool.TokenXTag
			}
			paths2, _ := findPoolPaths(idToPool, tokenTagToPoolIDs, tokenIn2, tokenOut, path)
			if len(paths2) != 0 {
				for _, p := range paths2 {
					if len(p) != 0 {
						paths = append(paths, p)
					}
				}
			}
		}
	}
	return paths, nil
}
