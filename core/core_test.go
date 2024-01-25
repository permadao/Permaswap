package core

import (
	"encoding/json"
	"testing"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/permadao/permaswap/core/schema"
	routerSchema "github.com/permadao/permaswap/router/schema"
	"github.com/stretchr/testify/assert"
)

func testStringToDecimal(s string) *apd.Decimal {
	d := new(apd.Decimal)
	d.SetString(s)
	return d
}

func TestNew(t *testing.T) {
	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	assert.NoError(t, err)

	poolID := pool.ID()
	t.Log("PoolId:", poolID)

	core := New(map[string]*schema.Pool{poolID: pool}, "", "")
	t.Log("core:", core, "\n")

	address := "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"

	addMsg := routerSchema.LpMsgAdd{
		TokenX:           "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.003"),
		LowSqrtPrice:     testStringToDecimal("0.000044721359549995793928183473374626"),
		CurrentSqrtPrice: testStringToDecimal("0.000054792195750516611345696978280080"),
		HighSqrtPrice:    testStringToDecimal("0.000063245553203367586639977870888654"),
		Liquidity:        "50000000000000000",
		PriceDirection:   "both",
	}

	t.Log(addMsg)

	err = core.AddLiquidity(address, addMsg)
	assert.NoError(t, err)
	t.Log("core:", core)
	t.Log("core.TokenTagToPoolIDs:", core.TokenTagToPoolIDs, "\n")

	pools, err := core.FindPoolPaths("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee")
	assert.NoError(t, err)
	for i, path := range pools {
		for j, pool := range path {
			t.Log(i, j, pool)
		}
	}

	pools, err = core.FindPoolPaths("ethereum-eth-0x0000000000000000000000000000000000000000",
		"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xcc9141efa8c20c7df0778748255b1487957811be")
	assert.Equal(t, ERR_NO_POOL, err)

}

func TestQueryAndUpdate(t *testing.T) {
	user := "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"
	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	poolID := pool.ID()

	core := New(map[string]*schema.Pool{poolID: pool}, "", "")

	lpAddress := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	addMsg := routerSchema.LpMsgAdd{
		TokenX:           "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.003"),
		LowSqrtPrice:     testStringToDecimal("0.000044721359549995793928183473374626"),
		CurrentSqrtPrice: testStringToDecimal("0.000054792195750516611345696978280080"),
		HighSqrtPrice:    testStringToDecimal("0.000063245553203367586639977870888654"),
		Liquidity:        "50000000000000000",
		PriceDirection:   "both",
	}

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	t.Log("core:", core, "\n", "AddressToLpIDs:", core.AddressToLpIDs, "\n", "lps:", core.Lps)

	paths, err := core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "1000000",
	})
	assert.NoError(t, err)
	t.Log("Paths:", len(paths), paths, "\n")

	paths, err = core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "1",
	})
	assert.EqualError(t, err, "err_invalid_amount")

	paths, err = core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "1000000000",
	})
	assert.NoError(t, err)
	t.Log("Paths:", len(paths), paths, "\n")

	err = core.Verify(user, paths)
	assert.NoError(t, err)
	t.Log("Verify Done", "\n")

	err = core.Update(user, paths)
	assert.NoError(t, err)
	t.Log("Lp after update:")
	for _, lp := range core.Lps {
		t.Log(lp.ID(), lp.CurrentSqrtPrice)
	}
}

func TestQuery(t *testing.T) {
	user := "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"
	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	poolID := pool.ID()

	core := New(map[string]*schema.Pool{poolID: pool}, "", "")

	lpAddress := user
	addMsg := routerSchema.LpMsgAdd{
		TokenX:           "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.003"),
		LowSqrtPrice:     testStringToDecimal("0.000044721359549995793928183473374626"),
		CurrentSqrtPrice: testStringToDecimal("0.000054792195750516611345696978280080"),
		HighSqrtPrice:    testStringToDecimal("0.000063245553203367586639977870888654"),
		Liquidity:        "50000000000000000",
		PriceDirection:   "both",
	}

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	t.Log("core:", core, "\n", "AddressToLpIDs:", core.AddressToLpIDs, "\n", "lps:", core.Lps)

	paths, err := core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "1000000",
	})
	//assert.Equal(t, err, ERR_NO_PATH)
	assert.Equal(t, err, ERR_INVALID_SWAP_USER)
	assert.Equal(t, len(paths), 0)
}

func TestQuery2(t *testing.T) {
	user := "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"
	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	poolID := pool.ID()

	core := New(map[string]*schema.Pool{poolID: pool}, "", "")

	lpAddress := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	addMsg := routerSchema.LpMsgAdd{
		TokenX:           "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.003"),
		LowSqrtPrice:     testStringToDecimal("0.000044721359549995793928183473374626"),
		CurrentSqrtPrice: testStringToDecimal("0.000054792195750516611345696978280080"),
		HighSqrtPrice:    testStringToDecimal("0.000063245553203367586639977870888654"),
		Liquidity:        "50000000000000000",
		PriceDirection:   "both",
	}

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)

	addMsg = routerSchema.LpMsgAdd{
		TokenX:           "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.003"),
		LowSqrtPrice:     testStringToDecimal("0.000044721359549995793928183473374626"),
		CurrentSqrtPrice: testStringToDecimal("0.000054792195750516611345696978280080"),
		HighSqrtPrice:    testStringToDecimal("0.000073245553203367586639977870888654"),
		Liquidity:        "10000000000000000",
		PriceDirection:   "both",
	}

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	t.Log("core:", core, "\n", "AddressToLpIDs:", core.AddressToLpIDs, "\n")
	for lpID, lp := range core.Lps {
		t.Log("lpID:", lpID)
		t.Log(lp)
	}
	t.Log("\n")

	paths, err := core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "10000000",
	})
	assert.NoError(t, err)
	for _, path := range paths {
		t.Log(path)
	}
	t.Log("\n")

	err = core.Verify(user, paths)
	assert.NoError(t, err)
	t.Log("Verify Done", "\n")
}

func TestQueryWithMultiPools(t *testing.T) {
	user := "0x911F42b0229c15bBB38D648B7Aa7CA480eD977d6"

	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	pool2, err := NewPool("ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.001")
	pools := map[string]*schema.Pool{
		pool.ID():  pool,
		pool2.ID(): pool2,
	}
	core := New(pools, "", "")

	t.Log("core:", "TokenTagToPoolIDs:", len(core.TokenTagToPoolIDs), core.TokenTagToPoolIDs, "\n")

	lpAddress := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	addMsg := routerSchema.LpMsgAdd{
		TokenX:           "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.003"),
		LowSqrtPrice:     testStringToDecimal("0.000044721359549995793928183473374626"),
		CurrentSqrtPrice: testStringToDecimal("0.000054792195750516611345696978280080"),
		HighSqrtPrice:    testStringToDecimal("0.000063245553203367586639977870888654"),
		Liquidity:        "50000000000000000",
		PriceDirection:   "both",
	}
	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)

	lpAddress2 := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	addMsg2 := routerSchema.LpMsgAdd{
		TokenX:           "ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		TokenY:           "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:         testStringToDecimal("0.001"),
		LowSqrtPrice:     testStringToDecimal("0.9899494936611666"),
		CurrentSqrtPrice: testStringToDecimal("1"),
		HighSqrtPrice:    testStringToDecimal("1.0099504938362078"),
		Liquidity:        "40000000000000000",
		PriceDirection:   "both",
	}
	err = core.AddLiquidity(lpAddress2, addMsg2)
	assert.NoError(t, err)

	t.Log("core:", core, "\n\n", "AddressToLpIDs:", core.AddressToLpIDs, "\n\n", "lps:", core.Lps, "\n")

	paths, err := core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "1000000",
	})
	assert.NoError(t, err)
	t.Log("Paths:", len(paths), paths, "\n")

	paths, err = core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		TokenOut: "ethereum-eth-0x0000000000000000000000000000000000000000",
		AmountIn: "1000000000",
	})
	assert.NoError(t, err)
	t.Log("Paths:", len(paths), paths, "\n")

	err = core.Verify(user, paths)
	assert.NoError(t, err)
	t.Log("Verify Done", "\n")

	err = core.Update(user, paths)
	assert.NoError(t, err)

	t.Log("Lp after update:")
	for _, lp := range core.Lps {
		t.Log(lp.ID(), lp.CurrentSqrtPrice)
	}

	paths, err = core.Query(routerSchema.UserMsgQuery{
		Address:  user,
		TokenIn:  "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenOut: "ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		AmountIn: "100000000000000000",
	})
	assert.NoError(t, err)
	t.Log("Paths:", len(paths), paths, "\n")

}

func TestLiquidity(t *testing.T) {
	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	poolID := pool.ID()

	core := New(map[string]*schema.Pool{poolID: pool}, "", "")

	lpAddress := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	config := `{
		"tokenX": "ethereum-eth-0x0000000000000000000000000000000000000000",
		"tokenY": "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"feeRatio": "0.003",
		"lowSqrtPrice": "0.000044721359549995793928183473374626",
		"currentSqrtPrice": "0.000054792195750516611345696978280080",
		"highSqrtPrice": "0.000063245553203367586639977870888654",
		"liquidity": "50000000000000000",
		"priceDirection": "both"
	}`
	addMsg := routerSchema.LpMsgAdd{}
	json.Unmarshal([]byte(config), &addMsg)

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	t.Log("core:", core, "\n", "Pool lps count:", len(GetPoolLps(pool, []string{})), "AddressToLpIDs:", core.AddressToLpIDs, "\n")

	addMsg2 := routerSchema.LpMsgAdd{}
	json.Unmarshal([]byte(config), &addMsg2)

	err = core.AddLiquidity(lpAddress, addMsg2)
	assert.NoError(t, err)
	t.Log("core after add 2 lps:", core, "\n", "Pool lps count:", len(GetPoolLps(pool, []string{})), "AddressToLpIDs:", core.AddressToLpIDs, "\n")
	rmMsg := routerSchema.LpMsgRemove{
		TokenX:         "ethereum-eth-0x0000000000000000000000000000000000000000",
		TokenY:         "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		FeeRatio:       testStringToDecimal("0.003"),
		LowSqrtPrice:   testSqrtPrice("1E-9"),
		HighSqrtPrice:  testSqrtPrice("5E-9"),
		PriceDirection: schema.PriceDirectionBoth,
	}
	err = core.RemoveLiquidity(lpAddress, rmMsg)
	assert.NoError(t, err)
	t.Log("core:", core, "\n", "Pool lps count:", len(GetPoolLps(pool, []string{})), "AddressToLpIDs:", core.AddressToLpIDs, "\n")

}

func TestRemoveLiquidityByAddress(t *testing.T) {
	pool, err := NewPool("ethereum-eth-0x0000000000000000000000000000000000000000",
		"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"0.003")
	poolID := pool.ID()

	core := New(map[string]*schema.Pool{poolID: pool}, "", "")

	lpAddress := "0x61EbF673c200646236B2c53465bcA0699455d5FA"
	config := `{
		"tokenX": "ethereum-eth-0x0000000000000000000000000000000000000000",
		"tokenY": "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"feeRatio": "0.003",
		"lowSqrtPrice": "0.000044721359549995793928183473374626",
		"currentSqrtPrice": "0.000054792195750516611345696978280080",
		"highSqrtPrice": "0.000063245553203367586639977870888654",
		"liquidity": "50000000000000000",
		"priceDirection": "both"
	}`
	addMsg := routerSchema.LpMsgAdd{}
	json.Unmarshal([]byte(config), &addMsg)

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	assert.Equal(t, len(core.Lps), 1)

	config = `{
		"tokenX": "ethereum-eth-0x0000000000000000000000000000000000000000",
		"tokenY": "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"feeRatio": "0.003",
		"lowSqrtPrice": "0.000044721359549995793928183473374626",
		"currentSqrtPrice": "0.000054792195750516611345696978280080",
		"highSqrtPrice": "0.000093245553203367586639977870888654",
		"liquidity": "10000000000000000",
		"priceDirection": "both"
	}`

	addMsg = routerSchema.LpMsgAdd{}
	json.Unmarshal([]byte(config), &addMsg)

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	assert.Equal(t, len(core.Lps), 2)

	config = `{
		"tokenX": "ethereum-eth-0x0000000000000000000000000000000000000000",
		"tokenY": "ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		"feeRatio": "0.003",
		"lowSqrtPrice": "0.000044721359549995793928183473374626",
		"currentSqrtPrice": "0.000054792195750516611345696978280080",
		"highSqrtPrice": "0.000091245553203367586639977870888654",
		"liquidity": "10000000000000000",
		"priceDirection": "both"
	}`

	addMsg = routerSchema.LpMsgAdd{}
	json.Unmarshal([]byte(config), &addMsg)

	err = core.AddLiquidity(lpAddress, addMsg)
	assert.NoError(t, err)
	assert.Equal(t, len(core.Lps), 3)

	err = core.RemoveLiquidityByAddress(lpAddress)
	assert.NoError(t, err)
	assert.Equal(t, len(core.Lps), 0)
}
