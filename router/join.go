package router

import "github.com/permadao/permaswap/halo/hvm/schema"

func (r *Router) Join() {
	httpEndpoint := ""
	wsEndpoint := ""
	if r.domain != "" {
		httpEndpoint = "https://" + r.domain
		wsEndpoint = "wss://" + r.domain
	}

	pools := map[string]*schema.Pool{}
	for _, pool := range r.core.Pools {
		pools[pool.ID()] = &schema.Pool{
			TokenXTag: pool.TokenXTag,
			TokenYTag: pool.TokenYTag,
			FeeRatio:  pool.FeeRatio.String(),
		}
	}

	routerState := schema.RouterState{
		Router:           r.sdk.AccId,
		Name:             r.name,
		HTTPEndpoint:     httpEndpoint,
		WSEndpoint:       wsEndpoint,
		SwapFeeRecipient: r.core.FeeRecepient,
		SwapFeeRatio:     r.core.FeeRatio.String(),
		Pools:            pools,
	}
	tx, err := r.haloSDK.Join(routerState)
	if err != nil {
		log.Error("AutoJoin tx submit failed: %v", err)
		return
	}
	log.Info("AutoJoin tx submit success", "tx", tx.EverHash)
}
