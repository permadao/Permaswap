package lp

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/permadao/permaswap/router/schema"
)

func (l *Lp) processPendingOrder() {
	pendingOrder, err := l.loadPendingOrder()

	if err != nil {
		return
	}

	log.Warn("found pending order")
	latestOrder, err := l.loadLatestOrder()
	if err != nil {
		panic("failed to load latest order")
	}

	// when everpay restart, rawid maybe changed. so use everhash to find the new rawid
	tx, err := l.rsdk.EverSDK.Cli.TxByHash(latestOrder.EverHash)
	if err != nil {
		log.Error("failed to load latest order from everpay", "err", err)
		panic(err)
	}
	lastRawID := tx.Tx.RawId

	txs := getTxsByCursor(l.rsdk.EverSDK.Cli, l.routerAddress, lastRawID)
	if len(txs) == 0 {
		log.Info("pending order was not submitted to everpay. ignore it.")
		l.removePendingOrder()
		return
	}

	l.order = pendingOrder
	for _, tx := range txs {
		l.processRouterOrder(tx)
	}
	if l.order != nil {
		panic("process pending order failed")
	}
	log.Info("finish pending order")

}

func (l *Lp) getFilePath(fileName string) (string, error) {
	absPath, err := filepath.Abs(l.configPath)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(filepath.Dir(absPath), fileName)
	return filePath, nil
}

func (l *Lp) savePendingOrder(msg *schema.LpMsgOrder) error {
	orderFilePath, err := l.getFilePath("pending_order.json")
	if err != nil {
		return err
	}

	by, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(orderFilePath, by, 0644)
}

func (l *Lp) loadPendingOrder() (msg *schema.LpMsgOrder, err error) {
	orderFilePath, err := l.getFilePath("pending_order.json")
	if err != nil {
		return
	}

	file, err := os.Open(orderFilePath)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &msg); err != nil {
		return
	}

	return
}

func (l *Lp) removePendingOrder() error {
	orderFilePath, err := l.getFilePath("pending_order.json")
	if err != nil {
		return err
	}

	return os.Remove(orderFilePath)
}

func (l *Lp) saveLatestOrder(tx everSchema.TxResponse) error {
	orderFilePath, err := l.getFilePath("latest_order.json")
	if err != nil {
		return err
	}

	by, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(orderFilePath, by, 0644)
}

func (l *Lp) loadLatestOrder() (tx everSchema.TxResponse, err error) {
	orderFilePath, err := l.getFilePath("latest_order.json")
	if err != nil {
		return
	}

	file, err := os.Open(orderFilePath)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	if err = json.Unmarshal(data, &tx); err != nil {
		return
	}

	return
}
