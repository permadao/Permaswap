package router

import (
	"math"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/everVision/everpay-kits/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/router/schema"
)

func ClosedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if PermaswapClosed {
			c.AbortWithStatusJSON(http.StatusOK, map[string]string{"status": "closed"})
		}
		c.Next()
	}
}

func (r *Router) runAPI(port, haloAPIURLPrefix string) {
	e := r.engine
	e.Use(cors.Default())
	e.Use(ClosedMiddleware())
	// websocket
	e.GET("/wsuser", r.wsUser)
	e.GET("/wslp", r.wsLp)

	// api
	e.GET("/info", r.getInfo)
	e.GET("/orders/*accid", r.getOrders)
	e.GET("/pool/:poolid", r.getPool)
	e.GET("/lps", r.getLps)
	e.GET("/nft", r.getNFT)
	e.GET("/stats", r.getStats)
	e.GET("/lpreward", r.getLpReward)
	e.GET("/penalty", r.getPenalty)

	if haloAPIURLPrefix != "" {
		r.haloServer.RegisterRouter(e, haloAPIURLPrefix)
	}

	r.server = &http.Server{
		Addr:    port,
		Handler: r.engine,
	}
	log.Info("listening", "port", port)
	if err := r.server.ListenAndServe(); err != nil {
		log.Warn("server closed", "err", err)
	}
}

func (r *Router) wsUser(c *gin.Context) {
	r.userHub.RegisterSession(c.Writer, c.Request)
}

func (r *Router) wsLp(c *gin.Context) {
	r.lpHub.RegisterSession(c.Writer, c.Request)
}

func (r *Router) getInfo(c *gin.Context) {
	tokenList := []string{}
	r.apiTokenTagsLock.RLock()
	for tag, ok := range r.apiTokenTags {
		if ok {
			tokenList = append(tokenList, tag)
		}
	}
	r.apiTokenTagsLock.RUnlock()
	sort.Strings(tokenList)

	routerAddress := ""
	if r.sdk != nil {
		routerAddress = r.sdk.AccId
	}

	c.JSON(http.StatusOK, schema.InfoRes{
		ChainID:       r.chainID,
		RouterAddress: routerAddress,
		NFTWhiteList:  SetNFTWhiteList(r.chainID),
		TokenList:     tokenList,
		PoolList:      r.core.Pools,
		LpClientInfo:  r.LpClientInfo,
	})
}

func (r *Router) getOrders(c *gin.Context) {
	accid := c.Param("accid")
	accid = strings.TrimPrefix(accid, "/")

	countStr := c.DefaultQuery("count", "10")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewWsErr(err.Error()))
		return
	}
	if count > 200 || count < 1 {
		c.JSON(http.StatusBadRequest, NewWsErr("err_invalid_param"))
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewWsErr(err.Error()))
		return
	}

	start, _ := time.ParseInLocation("2006-01-02", "2022-05-01", time.Local)
	startStr := c.DefaultQuery("start", "")
	if startStr != "" {
		start, _ = time.ParseInLocation("2006-01-02", startStr, time.Local)
	}

	end, _ := time.ParseInLocation("2006-01-02", time.Now().Add(24*time.Hour).Format("2006-01-02"), time.Local)
	endStr := c.DefaultQuery("end", "")
	if endStr != "" {
		end, _ = time.ParseInLocation("2006-01-02", endStr, time.Local)
		end = end.Add(24 * time.Hour)
	}

	orders := []*schema.PermaOrder{}
	if accid == "" {
		orders, _ = r.wdb.GetOrders(int(page), count, start, end)
	} else {
		_, accid, err = utils.IDCheck(accid)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewWsErr(err.Error()))
			return
		}
		//no count total number of a user's orders. because of performance issue.
		//num, _ := r.wdb.TotalOrdersNumByUser(accid)
		orders, _ = r.wdb.GetOrdersByUser(accid, int(page), count, start, end)
	}

	c.JSON(http.StatusOK, schema.OrdersRes{
		//Total:  num,
		Orders: orders,
	})
}

func (r *Router) getLpReward(c *gin.Context) {
	rewards := []*schema.PermaLpReward{}
	accid := ""
	lpID := ""
	if accid = c.Query("accid"); accid != "" {
		_, accid, err := utils.IDCheck(accid)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewWsErr(err.Error()))
			return
		}
		rewards, _ = r.wdb.GetPermaRewards(accid, nil)
	} else {
		lpID = c.Query("lpid")
		if lpID == "" {
			c.JSON(http.StatusBadRequest, NewWsErr("err_no_param"))
			return
		}
		reward, _ := r.wdb.GetPermaReward(lpID, nil)
		accid = reward.AccID
		rewards = append(rewards, reward)
	}

	c.JSON(http.StatusOK, schema.LpRewardsRes{
		Address: accid,
		LpID:    lpID,
		Rewards: rewards,
	})
}

func (r *Router) getLps(c *gin.Context) {
	if accid := c.Query("accid"); accid != "" {
		_, accid, err := utils.IDCheck(accid)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewWsErr(err.Error()))
			return
		}
		r.apiGetLpsByAccidReq <- accid
	} else {
		poolID := c.Query("poolid")
		if poolID == "" {
			c.JSON(http.StatusBadRequest, NewWsErr("err_no_param"))
			return
		}
		r.apiGetLpsByPoolidReq <- poolID
	}

	c.JSON(http.StatusOK, schema.LpsRes{Lps: <-r.apiGetLpsRes})
}

func (r *Router) getLpsByAccidProc(accid string) []coreSchema.Lp {
	lps := r.core.GetLps(accid)

	// append lp from pending order
	for _, o := range r.orders {
		for _, lp := range o.Lps {
			lps = append(lps, *lp)
		}
	}

	if lps == nil {
		return []coreSchema.Lp{}
	}
	return lps
}

func (r *Router) getLpsByPoolidProc(poolID string) []coreSchema.Lp {
	if pool, ok := r.core.Pools[poolID]; ok {
		return core.GetPoolLps2(pool)
	} else {
		return []coreSchema.Lp{}
	}
}

func (r *Router) getAllLpsProc() []coreSchema.Lp {
	lps := r.core.GetAllLps()

	// append lp from pending order
	for _, o := range r.orders {
		for _, lp := range o.Lps {
			lps = append(lps, *lp)
		}
	}

	if lps == nil {
		return []coreSchema.Lp{}
	}
	return lps
}

func (r *Router) getNFT(c *gin.Context) {
	nftRes := schema.NFTRes{
		NFTToHolder:  map[string]string{},
		HolderToNFTs: map[string][]string{},
		WhiteList:    []string{},
	}
	if r.NFTInfo != nil {
		n, h, w := r.NFTInfo.GetNFTInfo()
		nftRes = schema.NFTRes{
			NFTToHolder:  n,
			HolderToNFTs: h,
			WhiteList:    w}
	}
	c.JSON(http.StatusOK, nftRes)
}

func (r *Router) getPool(c *gin.Context) {
	poolID := c.Param("poolid")
	if poolID == "" {
		c.JSON(http.StatusBadRequest, NewWsErr("err_no_param"))
		return
	}
	r.apiGetPoolReq <- poolID
	c.JSON(http.StatusOK, <-r.apiGetPoolRes)
}

func (r *Router) getPoolProc(poolID string) *schema.PoolRes {
	if pool, ok := r.core.Pools[poolID]; ok {
		priceUp, err1 := r.core.GetPoolCurrentPrice2(poolID, coreSchema.PriceDirectionUp)
		priceDown, err2 := r.core.GetPoolCurrentPrice2(poolID, coreSchema.PriceDirectionDown)

		priceDown2 := ""
		priceUp2 := ""
		if err1 == nil && err2 == nil {
			decimalsX := r.tokens[pool.TokenXTag].Decimals
			decimalsY := r.tokens[pool.TokenYTag].Decimals
			factor := new(big.Float).SetFloat64(math.Pow(10, float64((decimalsX - decimalsY))))

			priceDown_, _ := new(big.Float).SetString(priceDown)
			priceUp_, _ := new(big.Float).SetString(priceUp)

			priceDown2 = new(big.Float).Mul(priceDown_, factor).Text('f', 32)
			priceUp2 = new(big.Float).Quo(priceUp_, factor).Text('f', 32)
		}

		lps := []coreSchema.Lp{}
		if pool, ok := r.core.Pools[poolID]; ok {
			lps = core.GetPoolLps2(pool)
		}
		return &schema.PoolRes{
			Pool:             *pool,
			CurrentPriceUP:   priceUp2,
			CurrentPriceDown: priceDown2,
			Lps:              lps,
		}
	}

	return nil
}

func (r *Router) getStats(c *gin.Context) {
	if accid := c.Query("accid"); accid != "" {
		_, accid, err := utils.IDCheck(accid)
		if err != nil {
			c.JSON(http.StatusBadRequest, NewWsErr(err.Error()))
			return
		}
		volumes := r.Stats.GetVolumesByAccid(accid)
		rewards := r.Stats.GetRewardsByAccid(accid)
		tvls := r.Stats.GetTVLsByAccid(accid)
		c.JSON(http.StatusOK, schema.AccountStatsRes{
			Address: accid,
			Volumes: volumes,
			Rewards: rewards,
			TVLs:    tvls,
		})
	} else {
		poolID := c.Query("poolid")
		if poolID == "" {
			c.JSON(http.StatusBadRequest, NewWsErr("err_no_param"))
			return
		}
		volume := r.Stats.GetVolumeByPoolID(poolID)
		tvl := r.Stats.GetTVLByPoolID(poolID)
		c.JSON(http.StatusOK, schema.PoolStatsRes{
			PoolID: poolID,
			Volume: volume,
			TVL:    tvl,
		})
	}
}

func (r *Router) getPenalty(c *gin.Context) {
	blacklist, failure := r.penalty.GetPenalty()
	c.JSON(http.StatusOK, schema.PenaltyRes{
		CumulativeFailures: CumulativeFailures,
		ExpirationDuration: ExpirationDuration,
		FailureRecords:     failure,
		BlackList:          blacklist,
	})
}
