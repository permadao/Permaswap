package halo

import (
	"github.com/everVision/everpay-kits/sdk"
	"github.com/gin-gonic/gin"
	"github.com/permadao/permaswap/halo/hvm"
	"github.com/permadao/permaswap/halo/logger"
	"github.com/permadao/permaswap/halo/schema"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
)

var log = logger.New("halo")

type Halo struct {
	hvm *hvm.HVM

	everSDK           *sdk.SDK
	GenesisTxEverHash string
	GenesisTxRawID    int64
	HaloAddr          string

	tracker *sdk.SubscribeTx
	engine  *gin.Engine
	wdb     *WDB

	// channels
	close          chan struct{}
	closed         chan struct{}
	txApplyChan    chan *schema.TxApply
	txApplyResChan chan error
	stateChan      chan struct{}
	stateResChan   chan string
	tokenChan      chan struct{}
	tokenResChan   chan *tokSchema.TokenInfo
	txSave         chan *schema.HaloTransaction
}

func New(genesisTx, dsn string, everSDK *sdk.SDK) (h *Halo) {
	tx, err := everSDK.Cli.TxByHash(genesisTx)
	if err != nil {
		log.Error("get genesis tx failed", "everHash", genesisTx, "err", err)
		panic(err)
	}
	txResp := *tx.Tx

	state, err := GenesisTxVerify(txResp)
	if err != nil {
		log.Error("verify genesis tx failed", "err", err)
		panic(err)
	}
	return &Halo{
		hvm:               hvm.New(*state),
		everSDK:           everSDK,
		GenesisTxEverHash: genesisTx,
		GenesisTxRawID:    txResp.RawId,
		HaloAddr:          txResp.To,
		engine:            gin.Default(),
		wdb:               NewWDB(dsn),
		close:             make(chan struct{}),
		closed:            make(chan struct{}),
		txApplyChan:       make(chan *schema.TxApply),
		txApplyResChan:    make(chan error),
		stateChan:         make(chan struct{}),
		stateResChan:      make(chan string),
		tokenChan:         make(chan struct{}),
		tokenResChan:      make(chan *tokSchema.TokenInfo),
		txSave:            make(chan *schema.HaloTransaction),
	}
}

func (h *Halo) Run(port string) {
	h.wdb.Migrate()
	h.track(h.GenesisTxRawID)
	go h.runProcess()
	go h.txSaveProcess()
	if port != "" {
		go h.runAPI(port)
	}
}

func (h *Halo) Close() {
	h.tracker.Unsubscribe()

	close(h.close)
	<-h.closed
	log.Info("halo closed")
}
