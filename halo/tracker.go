package halo

import (
	everSchema "github.com/everVision/everpay-kits/schema"
)

func (h *Halo) track(startCusor int64) {
	h.tracker = h.everSDK.Cli.SubscribeTxs(everSchema.FilterQuery{
		StartCursor: startCusor,
		Address:     h.HaloAddr,
	})
	log.Info("start tracking", "startCusor", startCusor, "addr", h.HaloAddr)
}
