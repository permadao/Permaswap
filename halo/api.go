package halo

import (
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"

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
	g.GET("/txs", h.txs)
	g.GET("/tx/:hash", h.getTx)
	g.GET("/proposal/:id", h.getProposal)
	g.GET("/balance/:accid", h.getBalance)
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
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
	state.Executed = []string{}
	state.Validity = map[string]bool{}

	pids := []string{}
	for _, p := range state.Proposals {
		pids = append(pids, p.ID)
	}
	state.Proposals = []*hvmSchema.Proposal{}

	res := &schema.InfoRes{
		State:             state,
		ProposalIds:       pids,
		GenesisTxEverHash: h.GenesisTxEverHash,
		HaloAddr:          h.HaloAddr,
	}

	c.JSON(http.StatusOK, res)
}

func (h *Halo) getProposal(c *gin.Context) {
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, schema.ErrMissParams.Error())
		return
	}

	detail_ := c.DefaultQuery("detail", "false")
	detail, err := strconv.ParseBool(detail_)
	if err != nil {
		detail = false
	}

	h.stateChan <- struct{}{}
	stateRes := <-h.stateResChan

	state := hvmSchema.State{}
	if stateRes != "" {
		if err := json.Unmarshal([]byte(stateRes), &state); err != nil {
			log.Error("unmarshal state failed", "err", err)
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
	for _, p := range state.Proposals {
		if p.ID == c.Param("id") {
			if !detail {
				p.ExecutedTxs = map[string]string{}
			}
			c.JSON(http.StatusOK, p)
			return
		}
	}
}

func (h *Halo) getTx(c *gin.Context) {
	if c.Param("hash") == "" {
		c.JSON(http.StatusBadRequest, schema.ErrMissParams.Error())
		return
	}

	tx, err := h.wdb.GetTx(c.Param("hash"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *Halo) txs(c *gin.Context) {
	h.stateChan <- struct{}{}
	stateRes := <-h.stateResChan

	state := hvmSchema.State{}
	if stateRes != "" {
		if err := json.Unmarshal([]byte(stateRes), &state); err != nil {
			log.Error("unmarshal txs state failed", "err", err)
		}
	}
	executed := []schema.TxWithValidity{}
	for _, tx := range state.Executed {
		executed = append(executed, schema.TxWithValidity{
			Tx:       tx,
			Validity: state.Validity[tx],
		})

	}
	res := &schema.TxRes{
		Executed: executed,
	}
	c.JSON(http.StatusOK, res)
}

func (h *Halo) tokenInfo(c *gin.Context) {
	h.tokenChan <- struct{}{}
	tokenRes := <-h.tokenResChan
	c.JSON(http.StatusOK, tokenRes)
}

func (h *Halo) getBalance(c *gin.Context) {
	if c.Param("accid") == "" {
		c.JSON(http.StatusBadRequest, schema.ErrMissParams.Error())
		return
	}

	h.tokenChan <- struct{}{}
	tokenRes := <-h.tokenResChan

	balance := "0"
	if b, ok := tokenRes.Balances[c.Param("accid")]; ok {
		balance = b
	}
	stakes := map[string][]tokSchema.StakeInfo{}
	if s, ok := tokenRes.Stakes[c.Param("accid")]; ok {
		stakes = s
	}

	c.JSON(http.StatusOK, schema.BalanceRes{
		Balance: balance,
		Stakes:  stakes,
	})
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
	c.JSON(http.StatusOK, schema.SubmitRes{
		EverHash: everTx.HexHash(),
		HaloHash: tx.HexHash()})
}
