package lp

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/permadao/permaswap/lp/schema"
	routerSchema "github.com/permadao/permaswap/router/schema"
)

func (l *Lp) runAPI(port string) {

	l.engine.GET("/info", l.info)
	l.engine.GET("/orders", l.getOrders)
	l.engine.POST("/remove_lp", l.removeLp)
	l.engine.POST("/add_lp", l.addLp)

	if !strings.Contains(port, ":") {
		port = ":" + port
	}
	if err := l.engine.Run(port); err != nil {
		panic(err)
	}
}

func (l *Lp) info(c *gin.Context) {
	l.apiInfoReq <- struct{}{}
	c.JSON(http.StatusOK, <-l.apiInfoRes)
}

func (l *Lp) removeLp(c *gin.Context) {
	req := schema.RemoveLpReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if req.LpID == "" {
		c.JSON(http.StatusBadRequest, errors.New("err_no_lpid"))
		return
	}
	l.apiRemoveLpReq <- req.LpID
	c.JSON(http.StatusOK, <-l.apiRemoveLpRes)
}

func (l *Lp) addLp(c *gin.Context) {
	msg := routerSchema.LpMsgAdd{}
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	l.apiAddLpReq <- &msg
	c.JSON(http.StatusOK, <-l.apiAddLpRes)
}

func (l *Lp) getOrders(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	orders, _ := l.wdb.GetOrders(int(page), 10)
	c.JSON(http.StatusOK, orders)
}
