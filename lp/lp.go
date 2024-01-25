package lp

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/gin-gonic/gin"
	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/logger"
	"github.com/permadao/permaswap/lp/schema"
	routerSchema "github.com/permadao/permaswap/router/schema"
)

var LpName = "lp-golang"
var LpVersion = "v0.5.1"

var log = logger.New("lp")

type Lp struct {
	chainID int64
	tokens  map[string]*everSchema.Token
	pools   map[string]*coreSchema.Pool

	core *core.Core
	rsdk *RSDK

	configPath string
	order      *routerSchema.LpMsgOrder

	close  chan struct{}
	closed chan struct{}

	routerAddress string
	sub           *sdk.SubscribeTx

	engine *gin.Engine

	apiEnabled     bool
	apiInfoReq     chan struct{}
	apiInfoRes     chan *schema.InfoRes
	apiRemoveLpReq chan string
	apiRemoveLpRes chan *schema.RemoveLpRes
	apiAddLpReq    chan *routerSchema.LpMsgAdd
	apiAddLpRes    chan *schema.AddLpRes

	wdb *WDB
}

func New(chainID int64, apiEnabled bool, rsdk *RSDK) *Lp {

	tokens, err := rsdk.EverSDK.Cli.GetTokens()
	if err != nil {
		log.Error("failed to get tokens info", "err", err)
	}
	return &Lp{
		chainID: chainID,
		tokens:  tokens,
		pools:   make(map[string]*coreSchema.Pool),
		rsdk:    rsdk,

		close:  make(chan struct{}),
		closed: make(chan struct{}),

		engine: gin.Default(),

		apiEnabled:     apiEnabled,
		apiInfoReq:     make(chan struct{}),
		apiInfoRes:     make(chan *schema.InfoRes),
		apiRemoveLpReq: make(chan string),
		apiRemoveLpRes: make(chan *schema.RemoveLpRes),
		apiAddLpReq:    make(chan *routerSchema.LpMsgAdd),
		apiAddLpRes:    make(chan *schema.AddLpRes),
	}
}

func (l *Lp) Run(configPath string) {
	log.Info("lp start", "name", LpName, "version", LpVersion)
	dir, err := filepath.Abs(configPath)
	if err != nil {
		log.Error("Invalid config path.", "err", err)
		panic(err)
	}
	l.wdb = NewWDB(filepath.Dir(dir))
	l.wdb.Migrate()

	l.getRouterInfo()
	l.getCore()

	l.configPath = configPath
	if err := l.loadConfig(); err != nil {
		log.Error("failed to load config.", "err", err)
		panic(err)
	}

	l.processPendingOrder()
	l.subscribeRouterOrder()

	go l.runProcess()

	if err := l.cleanLiquidity(); err != nil {
		panic(err)
	}
	if err := l.registerLiquidity(); err != nil {
		panic(err)
	}

	if l.apiEnabled {
		go l.runAPI("8081")
	}
}

func (l *Lp) Close() {
	l.cleanLiquidity()
	l.rsdk.Close()
	close(l.close)
	l.sub.Unsubscribe()
	<-l.closed
	log.Info("lp closed")
}

func (l *Lp) loadConfig() (err error) {
	file, err := os.Open(l.configPath)
	if err != nil {
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	msgs := []routerSchema.LpMsgAdd{}
	if err = json.Unmarshal(data, &msgs); err != nil {
		return
	}

	for _, msg := range msgs {
		err = l.core.AddLiquidity(l.rsdk.AccID, msg)
		if err != nil {
			log.Error("load liquidity failed", "err", err)
			return
		}
	}
	return
}

func (l *Lp) checkBalance(lps []coreSchema.Lp, retry bool) (err error) {
	balances := map[string]*big.Int{}

	for _, lp := range lps {
		amountX, amountY, err := core.LiquidityToAmount(lp.Liquidity.String(), lp.LowSqrtPrice, lp.CurrentSqrtPrice, lp.HighSqrtPrice, lp.PriceDirection)
		if err != nil {
			return err
		}
		x, ok := new(big.Int).SetString(amountX, 10)
		if !ok {
			return ERR_INVALID_AMOUNT
		}
		y, ok := new(big.Int).SetString(amountY, 10)
		if !ok {
			return ERR_INVALID_AMOUNT
		}

		if b, ok := balances[lp.TokenXTag]; ok {
			balances[lp.TokenXTag] = new(big.Int).Add(b, x)
		} else {
			balances[lp.TokenXTag] = x
		}

		if b, ok := balances[lp.TokenYTag]; ok {
			balances[lp.TokenYTag] = new(big.Int).Add(b, y)
		} else {
			balances[lp.TokenYTag] = y
		}
	}

	log.Info("balance needed:")
	for t, v := range balances {
		log.Info("", t, v)
	}

	for token, amount := range balances {
		a := ""

		n := 1
		for {
			a, err = l.rsdk.GetBalance(token)
			if err == nil {
				break
			}
			if !retry {
				return err
			}
			if n >= 3 {
				return err
			}
			n += 1
			time.Sleep(200 * time.Millisecond)
		}

		a_, ok := new(big.Int).SetString(a, 10)
		if !ok {
			return ERR_INVALID_AMOUNT
		}

		if amount.Cmp(a_) == 1 {
			log.Error("lp balance is not enough", "token", token, "needed", amount, "have", a_)
			return ERR_NO_ENOUGH_BALANCE
		}
	}

	return
}

func (l *Lp) getRouterInfo() {
	for {
		info, err := l.rsdk.GetInfo()
		if err == nil {
			l.routerAddress = info.RouterAddress
			for _, p := range info.PoolList {
				pool, err := core.NewPool(p.TokenXTag, p.TokenYTag, p.FeeRatio.String())
				if err != nil {
					log.Error("failed to get pool from router api", "err", err)
					panic(err)
				}
				l.pools[pool.ID()] = pool
			}
			break
		}
		log.Warn("failed to get router address", "err", err)
		time.Sleep(100 * time.Millisecond)
	}
	log.Info("Router address:", "router", l.routerAddress)
}

func (l *Lp) getCore() {
	l.core = core.New(l.pools, "", "")
}

func (l *Lp) subscribeRouterOrder() {
	latestTxRawId := int64(0)
	txs, err := l.rsdk.EverSDK.Cli.Txs(0, "desc", 1, everSchema.TxOpts{
		Address: l.routerAddress,
	})
	if err != nil {
		log.Error("failed to get the latest tx of router")
		panic(err)
	}
	if txs.Txs != nil {
		if len(txs.Txs) > 0 {
			latestTxRawId = txs.Txs[0].RawId
			log.Info("latest tx of router", "rawId", latestTxRawId)
		}
	}

	l.sub = l.rsdk.EverSDK.Cli.SubscribeTxs(everSchema.FilterQuery{
		StartCursor: latestTxRawId,
		Address:     l.routerAddress,
	})
	log.Info("Start to subscribe router's order", "routerAddress", l.routerAddress, "latestTxRawId", latestTxRawId)
}

func (l *Lp) reconnect() (err error) {
	if tokens, err := l.rsdk.EverSDK.Cli.GetTokens(); err != nil {
		return err
	} else {
		l.tokens = tokens
	}

	l.getRouterInfo()
	l.getCore()
	if err := l.loadConfig(); err != nil {
		log.Error("failed to load config.", "err", err)
		panic(err)
	}
	l.processPendingOrder()
	return l.registerLiquidity()
}

func (l *Lp) registerLiquidity() (err error) {
	lps := l.core.GetLps(l.rsdk.AccID)
	if err := l.checkBalance(lps, true); err != nil {
		log.Error("check lp balance: failed.", "err", err)
		panic(err)
	}

	for _, lp := range lps {
		err = l.rsdk.AddLiquidity(LpToAddMsg(lp))
		if err != nil {
			return
		}

		amountX, amountY, _ := core.LiquidityToAmount(lp.Liquidity.String(), lp.LowSqrtPrice, lp.CurrentSqrtPrice, lp.HighSqrtPrice, lp.PriceDirection)
		log.Info("register liquidity", "x", lp.TokenXTag, "y", lp.TokenYTag, "amountX", amountX, "amountY", amountY, "currentSqrtPrice", lp.CurrentSqrtPrice)
		log.Info("lp price range", "low", core.LpLowPrice(lp), "current", core.LpCurrentPrice(lp), "high", core.LpHighPrice(lp), "liquidity", lp.Liquidity.String())
	}

	return
}

func (l *Lp) cleanLiquidity() (err error) {
	lps, err := l.rsdk.GetLps()
	if err != nil {
		return
	}

	for {
		for _, lp := range lps {
			l.rsdk.RemoveLiquidity(LpToRemoveMsg(lp))
		}

		lps, err = l.rsdk.GetLps()
		if err != nil {
			return
		}
		if len(lps) == 0 {
			log.Info("liqudity is clean")
			return
		}

		log.Warn("liquidity not clean", "accID", l.rsdk.AccID, "len(lps)", len(lps))
		time.Sleep(200 * time.Millisecond)
	}
}

func (l *Lp) UpdateLiquidity() error {
	lps := l.core.GetLps(l.rsdk.AccID)

	msgs := make([]routerSchema.LpMsgAdd, len(lps))
	for i, lp := range lps {
		msgs[i] = LpToAddMsg(lp)
	}

	by, err := json.MarshalIndent(msgs, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(l.configPath, by, 0644)
}

func LpToAddMsg(lp coreSchema.Lp) routerSchema.LpMsgAdd {
	return routerSchema.LpMsgAdd{
		TokenX:           lp.TokenXTag,
		TokenY:           lp.TokenYTag,
		FeeRatio:         lp.FeeRatio,
		LowSqrtPrice:     lp.LowSqrtPrice,
		CurrentSqrtPrice: lp.CurrentSqrtPrice,
		HighSqrtPrice:    lp.HighSqrtPrice,
		Liquidity:        lp.Liquidity.String(),
		PriceDirection:   lp.PriceDirection,
	}
}

func LpToRemoveMsg(lp coreSchema.Lp) routerSchema.LpMsgRemove {
	return routerSchema.LpMsgRemove{
		TokenX:         lp.TokenXTag,
		TokenY:         lp.TokenYTag,
		FeeRatio:       lp.FeeRatio,
		LowSqrtPrice:   lp.LowSqrtPrice,
		HighSqrtPrice:  lp.HighSqrtPrice,
		PriceDirection: lp.PriceDirection,
	}
}

func (l *Lp) getInfo() *schema.InfoRes {
	lps := map[string]coreSchema.Lp{}
	for i, lp := range l.core.Lps {
		lps[i] = *lp
	}
	info := &schema.InfoRes{
		Address:       l.rsdk.AccID,
		ChainID:       l.chainID,
		Tokens:        l.tokens,
		Pools:         l.core.Pools,
		RouterAddress: l.routerAddress,
		Lps:           lps,
	}
	return info
}
