package router

import (
	"context"
	"net/http"
	"sync"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"

	"github.com/everVision/everpay-kits/sdk"
	halosdk "github.com/permadao/permaswap/halo/sdk"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/halo"
	"github.com/permadao/permaswap/logger"
	"github.com/permadao/permaswap/router/schema"
	"github.com/permadao/permaswap/wshub"
)

var log = logger.New("router")

type Router struct {
	name    string
	domain  string
	engine  *gin.Engine
	server  *http.Server
	chainID int64
	tokens  map[string]*everSchema.Token
	core    *core.Core
	sdk     *sdk.SDK // everPay sdk
	wdb     *WDB

	// api instruction sets
	apiGetLpsByAccidReq  chan string
	apiGetLpsByPoolidReq chan string
	apiGetLpsRes         chan []coreSchema.Lp

	apiGetPoolReq chan string
	apiGetPoolRes chan *schema.PoolRes

	// api cache
	apiTokenTags     map[string]bool
	apiTokenTagsLock sync.RWMutex

	//stats
	getAllLpsReq chan struct{}

	lpHub *wshub.Hub
	// lp instruction sets
	lpInit       chan string
	lpRegister   chan *schema.LpMsgRegister
	lpUnregister chan string
	lpAdd        chan *schema.LpMsgAdd
	lpRemove     chan *schema.LpMsgRemove
	lpSign       chan *schema.LpMsgSign
	lpReject     chan *schema.LpMsgReject

	// lp cache
	lpAddrToID map[string]string // lp addr -> lp session id
	lpIDtoAddr map[string]string // lp session id -> lp addr
	lpSalt     map[string]string // lp session id -> salt

	userHub *wshub.Hub
	// user instruction sets
	userQuery  chan *schema.UserMsgQuery
	userSubmit chan *schema.UserMsgSubmit
	// if price change, use userQueryTag broadcastring new order
	userUnregister chan string
	// user cache
	userQueryTag map[string]map[string]*schema.UserMsgQuery // tag -> sessionid -> qryMsg

	// order instruction set
	orderStatus chan *Order
	// submit order cache
	orders map[string]*Order // orderHash -> order

	close  chan struct{}
	closed chan struct{}
	dryRun bool

	NFTInfo        *NFTInfo
	NFTOwnerChange chan *NFTOwnerChangeMsg

	Stats        *Stats
	scheduler    *gocron.Scheduler
	LpClientInfo map[string]*schema.LpClientInfo

	penalty *Penalty

	// halo
	haloServer *halo.Halo
	haloSDK    *halosdk.SDK
}

func New(name, domain string, chainID int64, everSDK *sdk.SDK, nftApiURL, dsn string,
	halo_genesis_tx string, haloSDK *halosdk.SDK, dryRun bool) *Router {
	w := &WDB{}
	if !dryRun {
		w = NewWDB(dsn)
	}
	tokens, err := everSDK.Cli.GetTokens()
	if err != nil {
		log.Error("failed to get tokens info", "err", err)
	}

	feeRecepient, feeRatio := GetFeeConf(chainID)
	c := core.New(core.InitPools(chainID), feeRecepient, feeRatio)

	var nftInfo *NFTInfo
	var nftOwnerChange chan *NFTOwnerChangeMsg
	if nftApiURL != "" {
		nftOwnerChange = make(chan *NFTOwnerChangeMsg, 200)
		nftInfo = NewNFTInfo(nftApiURL, []string{}, nftOwnerChange)
	}

	stats := NewStats(tokens, core.InitPools(chainID), w)

	var haloServer *halo.Halo
	if halo_genesis_tx != "" {
		haloServer = halo.New(halo_genesis_tx, dsn, everSDK)
	}

	return &Router{
		name:    name,
		domain:  domain,
		engine:  gin.Default(),
		chainID: chainID,
		tokens:  tokens,
		core:    c,
		sdk:     everSDK,
		wdb:     w,

		apiGetLpsByAccidReq:  make(chan string),
		apiGetLpsByPoolidReq: make(chan string),
		apiGetLpsRes:         make(chan []coreSchema.Lp),
		apiGetPoolReq:        make(chan string),
		apiGetPoolRes:        make(chan *schema.PoolRes),
		apiTokenTags:         make(map[string]bool),

		getAllLpsReq: make(chan struct{}),

		lpHub:      wshub.New(),
		lpInit:     make(chan string),
		lpRegister: make(chan *schema.LpMsgRegister),

		// 200 is the max total number of peramswap nft
		lpUnregister: make(chan string, 200),

		lpAdd:      make(chan *schema.LpMsgAdd),
		lpRemove:   make(chan *schema.LpMsgRemove),
		lpSign:     make(chan *schema.LpMsgSign),
		lpReject:   make(chan *schema.LpMsgReject),
		lpAddrToID: make(map[string]string),
		lpIDtoAddr: make(map[string]string),
		lpSalt:     make(map[string]string),

		userHub:        wshub.New(),
		userQuery:      make(chan *schema.UserMsgQuery),
		userSubmit:     make(chan *schema.UserMsgSubmit),
		userUnregister: make(chan string),
		userQueryTag:   make(map[string]map[string]*schema.UserMsgQuery),

		orders:      make(map[string]*Order),
		orderStatus: make(chan *Order),

		close:  make(chan struct{}),
		closed: make(chan struct{}),
		dryRun: dryRun,

		NFTInfo:        nftInfo,
		NFTOwnerChange: nftOwnerChange,

		Stats:        stats,
		scheduler:    gocron.NewScheduler(time.UTC),
		LpClientInfo: GetLpClientInfoConf(chainID),

		penalty: NewPenalty(),

		haloServer: haloServer,
		haloSDK:    haloSDK,
	}
}

func (r *Router) Run(port, haloAPIURLPrefix string) {
	if !r.dryRun {
		r.wdb.Migrate()
	}

	if r.NFTInfo != nil {
		// set nft whitelist
		wls, err := r.wdb.LoadNFTWhiteList()
		if err == nil && len(wls) > 0 {
			addrs := []string{}
			for _, w := range wls {
				addrs = append(addrs, w.UserAddr)
			}
			r.NFTInfo.SetWhitelist(addrs)
		}

		r.NFTInfo.Run()
	}

	if !r.dryRun {
		r.Stats.Run()
	}

	go r.runAPI(port, haloAPIURLPrefix)
	go r.runUserMsgUnmarshal()
	go r.runLpMsgUnmarshal()
	go r.runProcess()

	if !r.dryRun {
		go r.runJobs()
	}
	if r.haloServer != nil {
		r.haloServer.Run("")
	}
	if r.haloSDK != nil {
		r.Join()
	}
}

func (r *Router) Close() {
	close(r.close)
	<-r.closed

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	r.server.Shutdown(ctx)
	if r.haloSDK != nil {
		tx, err := r.haloSDK.Leave()
		log.Info("Leave tx submitted:", "tx", tx.EverHash, "err", err)
	}
	log.Info("router closed")
}
