package core

import (
	"math/big"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/everVision/everpay-kits/utils"
	"github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/logger"
	routerSchema "github.com/permadao/permaswap/router/schema"
)

var log = logger.New("core")

type Core struct {
	FeeRecepient      string                  // router fee recepient address
	FeeRatio          *apd.Decimal            // router fee ratio
	Pools             map[string]*schema.Pool // pool id -> Pool
	Lps               map[string]*schema.Lp   // lp id -> Lp
	AddressToLpIDs    map[string][]string     // address -> []lpID
	MaxPoolPathLength int
	TokenTagToPoolIDs map[string][]string // tokentag -> []lpID
}

func New(pools map[string]*schema.Pool, routerFeeRecepient, routerFeeRatio string) *Core {

	tokenTagToPoolIDs := map[string][]string{}
	for _, pool := range pools {

		poolIDs := tokenTagToPoolIDs[pool.TokenXTag]
		tokenTagToPoolIDs[pool.TokenXTag] = append(poolIDs, pool.ID())

		poolIDs = tokenTagToPoolIDs[pool.TokenYTag]
		tokenTagToPoolIDs[pool.TokenYTag] = append(poolIDs, pool.ID())
	}

	var err error
	recepient := routerFeeRecepient
	if routerFeeRecepient != "" {
		_, recepient, err = utils.IDCheck(routerFeeRecepient)
		if err != nil {
			panic(err)
		}
	}

	feeRatio := new(apd.Decimal).SetInt64(0)
	if routerFeeRatio != "" {
		feeRatio, err = StringToDecimal(routerFeeRatio)
		if err != nil {
			panic(err)
		}
	}

	return &Core{
		Pools:             pools,
		Lps:               make(map[string]*schema.Lp),
		AddressToLpIDs:    make(map[string][]string),
		MaxPoolPathLength: MaxPoolPathLength,
		TokenTagToPoolIDs: tokenTagToPoolIDs,
		FeeRatio:          feeRatio,
		FeeRecepient:      recepient,
	}
}

func (c *Core) AddLiquidity(address string, msg routerSchema.LpMsgAdd) error {
	pool, err := c.FindPool(msg.TokenX, msg.TokenY, msg.FeeRatio)
	if err != nil {
		return err
	}

	lp, err := NewLp(
		pool.ID(), msg.TokenX, msg.TokenY, address,
		msg.FeeRatio, msg.LowSqrtPrice, msg.CurrentSqrtPrice, msg.HighSqrtPrice,
		msg.Liquidity, msg.PriceDirection,
	)
	if err != nil {
		return err
	}

	err = PoolAddLiquidity(pool, lp)
	if err != nil {
		return err
	}
	lpID := lp.ID()

	// only new lp will be added to AddressToLpIDs, and lp id is unique.
	if _, ok := c.Lps[lpID]; !ok {
		lpIDs := c.AddressToLpIDs[lp.AccID]
		c.AddressToLpIDs[lp.AccID] = append(lpIDs, lp.ID())
	}

	c.Lps[lpID] = lp

	return nil
}

func (c *Core) RemoveLiquidity(address string, msg routerSchema.LpMsgRemove) error {
	_, accid, err := utils.IDCheck(address)
	if err != nil {
		return err
	}

	pool, err := c.FindPool(msg.TokenX, msg.TokenY, msg.FeeRatio)
	if err != nil {
		return err
	}

	lpID := GetLpID(pool.ID(), address, msg.LowSqrtPrice, msg.HighSqrtPrice, msg.PriceDirection)

	PoolRemoveLiquidity(pool, lpID)
	delete(c.Lps, lpID)
	if lpIDs, ok := c.AddressToLpIDs[accid]; ok {
		for i, other := range lpIDs {
			if other == lpID {
				lpIDs = append(lpIDs[:i], lpIDs[i+1:]...)
				break
			}
		}
		c.AddressToLpIDs[accid] = lpIDs
	}
	return nil
}

func (c *Core) AddLiquidityByLp(lp *schema.Lp) error {
	if lp == nil {
		return nil
	}
	pool, err := c.FindPool(lp.TokenXTag, lp.TokenYTag, lp.FeeRatio)
	if err != nil {
		return err
	}

	err = PoolAddLiquidity(pool, lp)
	if err != nil {
		return err
	}
	lpID := lp.ID()

	// only new lp will be added to AddressToLpIDs, and lp id is unique.
	if _, ok := c.Lps[lpID]; !ok {
		lpIDs := c.AddressToLpIDs[lp.AccID]
		c.AddressToLpIDs[lp.AccID] = append(lpIDs, lp.ID())
	}

	c.Lps[lpID] = lp
	return nil
}

func (c *Core) RemoveLiquidityByID(lpID string) (lp *schema.Lp, err error) {
	ok := false
	lp, ok = c.Lps[lpID]
	if !ok {
		err = ERR_NO_LP
		return
	}

	pool, err := c.FindPool(lp.TokenXTag, lp.TokenYTag, lp.FeeRatio)
	if err != nil {
		return
	}

	PoolRemoveLiquidity(pool, lpID)
	delete(c.Lps, lpID)
	if lpIDs, ok := c.AddressToLpIDs[lp.AccID]; ok {
		for i, other := range lpIDs {
			if other == lpID {
				lpIDs = append(lpIDs[:i], lpIDs[i+1:]...)
				break
			}
		}
		c.AddressToLpIDs[lp.AccID] = lpIDs
	}
	return
}

func (c *Core) RemoveLiquidityByAddress(lpAddr string) (err error) {
	_, accid, err := utils.IDCheck(lpAddr)
	if err != nil {
		return err
	}

	if lpIDs, ok := c.AddressToLpIDs[accid]; ok {
		lpIDsCopy := make([]string, len(lpIDs))
		copy(lpIDsCopy, lpIDs)

		for _, lpID := range lpIDsCopy {
			_, err = c.RemoveLiquidityByID(lpID)
		}
		return err
	}

	return nil
}

func (c *Core) Query(msg routerSchema.UserMsgQuery) ([]schema.Path, error) {
	_, addr, err := utils.IDCheck(msg.Address)
	if err != nil {
		return nil, err
	}

	if lpIDs, ok := c.AddressToLpIDs[addr]; ok {
		if len(lpIDs) > 0 {
			return nil, ERR_INVALID_SWAP_USER
		}
	}

	amountIn, ok := new(big.Int).SetString(msg.AmountIn, 10)
	if !ok {
		return nil, ERR_INVALID_AMOUNT
	}

	zero := big.NewInt(0)
	if amountIn.Cmp(zero) != 1 {
		return nil, ERR_INVALID_AMOUNT
	}

	poolPaths, err := c.FindPoolPaths(msg.TokenIn, msg.TokenOut)
	if err != nil {
		log.Warn("Failed to find pool paths.", "err", err)
		return nil, err
	}

	routerFee := big.NewInt(0)
	if c.FeeRecepient != "" && c.FeeRatio.Cmp(new(apd.Decimal).SetInt64(0)) == 1 {
		//routerFee, err = getAndCheckFee(amountIn, c.FeeRatio)
		routerFee, err = getFee(amountIn, c.FeeRatio, true)
		if err != nil {
			return nil, err
		}
		amountIn.Sub(amountIn, routerFee)
	}

	amountOut := big.NewInt(0)
	sos := []schema.SwapOutput{}
	errs := []error{}
	for i, poolPath := range poolPaths {
		sos_, amountOut_, err := PoolsSwap(poolPath, msg.TokenIn, msg.TokenOut, amountIn, c.AddressToLpIDs[addr])
		if err != nil {
			log.Debug("Failed to swap in one poolPaths", "path_index", i, "len(poolPaths)", len(poolPaths), "pooPath", poolPath,
				"tokenIn", msg.TokenIn, "tokenOut", msg.TokenOut, "amountIn", amountIn, "err", err)
			errs = append(errs, err)
			continue
		}
		if amountOut_.Cmp(amountOut) == 1 {
			amountOut = amountOut_
			sos = sos_
		}
	}

	if len(sos) == 0 {
		log.Error("Order query: Failed to find a swap path.", "lps in core", len(c.Lps))
		isInvalidAmount := true
		for _, e := range errs {
			if e.Error() != ERR_INVALID_AMOUNT.Error() {
				isInvalidAmount = false
				break
			}
		}
		if isInvalidAmount {
			return nil, ERR_INVALID_AMOUNT
		}

		return nil, ERR_NO_PATH
	}

	paths, err := SwapOutputsToPaths(addr, c, sos)
	if err != nil {
		return nil, err
	}

	// fee -> fee recepient
	if routerFee.Cmp(zero) == 1 {
		pathFee := schema.Path{
			LpID:     "",
			From:     msg.Address,
			To:       c.FeeRecepient,
			TokenTag: msg.TokenIn,
			Amount:   routerFee.String(),
		}
		paths = append(paths, pathFee)
	}

	return paths, nil
}

func (c *Core) GetLps(address string) (lps []schema.Lp) {
	_, accid, _ := utils.IDCheck(address)

	for _, lpID := range c.AddressToLpIDs[accid] {
		lp := c.Lps[lpID]
		lps = append(lps, *lp)
	}
	return lps
}

func (c *Core) GetAllLps() (lps []schema.Lp) {
	for _, lp := range c.Lps {
		lps = append(lps, *lp)
	}
	return lps
}

// Verify requires strict verification and should return an error if it is not the LP's own transaction
func (c *Core) Verify(userAddr string, paths []schema.Path) error {
	return c.update(userAddr, paths, true)
}

// Update is only required for all transactions of paths (if they exist)
func (c *Core) Update(userAddr string, paths []schema.Path) error {
	err := c.Verify(userAddr, paths)
	if err != nil {
		return err
	}
	return c.update(userAddr, paths, false)
}

func (c *Core) verifyPathFee(user string, paths []schema.Path) error {
	tokenIn := paths[0].TokenTag
	pathFee := paths[len(paths)-1]
	amount, ok := new(big.Int).SetString(pathFee.Amount, 10)
	if !ok {
		return ERR_INVALID_PATH_FEE
	}
	if pathFee.From != user {
		return ERR_INVALID_PATH_FEE
	}
	if pathFee.To != c.FeeRecepient {
		return ERR_INVALID_PATH_FEE
	}
	if pathFee.TokenTag != tokenIn {
		return ERR_INVALID_PATH_FEE
	}

	amountIn := big.NewInt(0)
	for _, path := range paths {
		if path.From == user && path.TokenTag == tokenIn {
			a, _ := new(big.Int).SetString(path.Amount, 10)
			amountIn.Add(amountIn, a)
		}
	}
	fee, err := getAndCheckFee(amountIn, c.FeeRatio)
	if err != nil {
		return ERR_INVALID_PATH_FEE
	}
	if amount.Cmp(fee) == -1 {
		return ERR_INVALID_PATH_FEE
	}
	return nil
}

func (c *Core) update(userAddr string, paths []schema.Path, isDryRun bool) error {
	_, userAddrID, err := utils.IDCheck(userAddr)
	if err != nil {
		return err
	}

	if lpIDs, ok := c.AddressToLpIDs[userAddr]; ok {
		if len(lpIDs) > 0 {
			return ERR_INVALID_SWAP_USER
		}
	}

	if c.FeeRecepient != "" && c.FeeRatio.Cmp(new(apd.Decimal).SetInt64(0)) == 1 {
		err = c.verifyPathFee(userAddr, paths)
		if err != nil {
			return err
		}
	}

	swapInputs, err := PathsToSwapInputs(userAddrID, paths)
	if err != nil {
		return err
	}

	for lpID, si := range swapInputs {
		//log.Info("func update", "lpID", lpID, "swapInput", si)
		lp, ok := c.Lps[lpID]
		if !ok {
			return ERR_NO_LP
		}
		so, err := LpSwap(lp, si.TokenIn, si.TokenOut, si.AmountIn, isDryRun)
		if err != nil {
			return err
		}

		if isDryRun {
			if so.TokenOut != si.TokenOut {
				log.Error("Invalid tokenOut", "lp", lp, "tokenOut in params", si.TokenOut, "tokenOut actual", so.TokenOut)
				return ERR_INVALID_PATH
			}

			if so.AmountOut.Cmp(si.AmountOut) == -1 {
				log.Error("amountOut is too small", "lp", lp)
				return ERR_INVALID_PATH
			}
		}

	}
	return nil
}

func (c *Core) FindPool(tokenX, tokenY string, feeRatio *apd.Decimal) (*schema.Pool, error) {
	for _, pool := range c.Pools {
		if pool.TokenXTag == tokenX && pool.TokenYTag == tokenY && pool.FeeRatio.String() == feeRatio.String() {
			return pool, nil
		}
	}
	return nil, ERR_NO_POOL
}

func (c *Core) GetPoolCurrentPrice(poolID, priceDirection string) (string, error) {
	if pool, ok := c.Pools[poolID]; ok {
		return GetPoolCurrentPrice(pool, priceDirection)
	}
	return "", nil
}

func (c *Core) GetPoolCurrentPrice2(poolID, priceDirection string) (string, error) {
	if pool, ok := c.Pools[poolID]; ok {
		return GetPoolCurrentPrice2(pool, priceDirection)
	}
	return "", nil
}
