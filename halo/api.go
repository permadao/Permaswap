package halo

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/schema"

	"github.com/gin-gonic/gin"
)

func (h *Halo) runAPI(port string) {
	h.engine.Use(cors.Default())
	h.RegisterRouter(h.engine, "")
	if !strings.Contains(port, ":") {
		port = ":" + port
	}
	if err := h.engine.Run(port); err != nil {
		panic(err)
	}
}

func (h *Halo) RegisterRouter(router *gin.Engine, group string) {
	if group == "" {
		group = "/"
	}
	g := router.Group(group)
	g.GET("/info", h.info)
	g.GET("/token", h.tokenInfo)
	g.POST("/submit", h.submit)
}

func (h *Halo) info(c *gin.Context) {
	h.stateChan <- struct{}{}
	stateRes := <-h.stateResChan

	state := hvmSchema.State{}
	if stateRes != "" {
		if err := json.Unmarshal([]byte(stateRes), &state); err != nil {
			log.Error("unmarshal state failed", "err", err)
		}
	}

	res := &schema.InfoRes{
		State:             state,
		GenesisTxEverHash: h.GenesisTxEverHash,
		HaloAddr:          h.HaloAddr,
	}

	c.JSON(http.StatusOK, res)
}

func (h *Halo) tokenInfo(c *gin.Context) {
	h.tokenChan <- struct{}{}
	tokenRes := <-h.tokenResChan
	c.JSON(http.StatusOK, tokenRes)
}

func (h *Halo) submit(c *gin.Context) {
	tx := hvmSchema.Transaction{}
	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, hvmSchema.ErrInvalidTx.Error())
		return
	}
	if err := verifyNonce(tx.Nonce); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// add router address
	tx.Router = h.everSDK.AccId

	// submit to hvm
	h.txApplyChan <- &schema.TxApply{
		Tx:     tx,
		DryRun: true,
	}
	if err := <-h.txApplyResChan; err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	// submit to everpay
	by, _ := json.Marshal(tx)
	eth := "ethereum-eth-0x0000000000000000000000000000000000000000"
	everTx, err := h.everSDK.Transfer(eth, big.NewInt(0), h.HaloAddr, string(by))
	if err != nil {
		log.Error("submit to everpay tx failed", "err", err)
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, schema.SubmitRes{EverHash: everTx.HexHash()})
}
