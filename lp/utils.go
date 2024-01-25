package lp

import (
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/everVision/everpay-kits/sdk"
)

func getTxsByCursor(client *sdk.Client, accid string, startCursor int64) (txs []everSchema.TxResponse) {
	cursorId := startCursor
	for {
		data, err := client.Txs(cursorId, "ASC", 50, everSchema.TxOpts{
			Address: accid,
		})
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		num := len(data.Txs)
		if num == 0 {
			return
		}

		cursorId = data.Txs[num-1].RawId
		txs = append(txs, data.Txs...)
		time.Sleep(100 * time.Millisecond)
	}
}
