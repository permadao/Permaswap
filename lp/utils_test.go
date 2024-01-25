package lp

import (
	"testing"

	"github.com/everVision/everpay-kits/sdk"
)

func TestGetTxsByCursord(t *testing.T) {
	client := sdk.NewClient("https://api.everpay.io")
	txs := getTxsByCursor(client, "0xd110107adb30bce6c0646eaf77cc1c815012331d", 12907668)
	t.Log(len(txs), txs[0].EverHash, txs[1].EverHash, txs[len(txs)-1].EverHash)
}
