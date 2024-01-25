package halo

import (
	"encoding/json"

	everSchema "github.com/everVision/everpay-kits/schema"
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/schema"
)

func (h *Halo) runProcess() {
	for {
		select {
		case txToAppy := <-h.txApplyChan:
			h.txApplyResChan <- h.txApplyProc(txToAppy)

		case txRsep := <-h.tracker.Subscribe():
			h.processHaloTx(txRsep)

		// fetch hvm state and hvm token info
		case <-h.stateChan:
			state, err := json.Marshal(h.hvm.State)
			if err != nil {
				log.Error("marshal state failed", "err", err)
				state = []byte{}
			}
			h.stateResChan <- string(state)
		case <-h.tokenChan:
			h.tokenResChan <- h.hvm.State.Token.Info()

		// close when other process finished
		case <-h.close:
			log.Info("process closed")
			close(h.txSave)
			close(h.closed)
			return
		}
	}
}

func (h *Halo) txSaveProcess() {
	for halotx := range h.txSave {
		err := h.wdb.CreateHaloTx(halotx, nil)
		if err != nil {
			log.Error("create halo tx failed", "everHash", halotx.EverHash, "err", err)
		}
	}
}

func (h *Halo) txApplyProc(txToAppy *schema.TxApply) (err error) {
	if txToAppy.DryRun {
		err = h.hvm.VerifyTx(txToAppy.Tx)
	} else {
		err = h.hvm.ExecuteTx(txToAppy.Tx)
	}
	return
}

func (h *Halo) processHaloTx(txResp everSchema.TxResponse) {
	log.Info("got new onchain halo tx", "everHash", txResp.EverHash, "action", txResp.Action)

	tx := hvmSchema.Transaction{}
	switch txResp.Action {
	case schema.EverTxActionTransfer:
		if err := json.Unmarshal([]byte(txResp.Data), &tx); err != nil {
			log.Error("invalid halo tx", "err", err)
			return
		}
	case schema.EverTxActionBundle:
		tx.Action = hvmSchema.TxActionSwap
		tx.Params = txResp.Data
	}

	tx.EverHash = txResp.EverHash
	tx.Router = txResp.From
	// submit to hvm
	var err error
	error := ""
	if err = h.hvm.ExecuteTx(tx); err != nil {
		error = err.Error()
	}
	log.Info("execute tx return", "everHash", txResp.EverHash, "err", err)

	if err == hvmSchema.ErrTxExecuted {
		return
	}

	haloTx := &schema.HaloTransaction{
		EverHash:    txResp.EverHash,
		HaloHash:    tx.HexHash(),
		Transaction: tx,
		Error:       error,
	}
	h.txSave <- haloTx

}
