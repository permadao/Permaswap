package router

import (
	"encoding/json"

	"github.com/permadao/permaswap/router/schema"
)

func (r *Router) runJobs() {
	r.scheduler.Every(2).Minute().SingletonMode().Do(r.saveLpsSnapshot)
	r.scheduler.Every(5).Minute().SingletonMode().Do(r.loadNFTWhiteList)
	r.scheduler.Every(5).Minute().SingletonMode().Do(r.cleanUpExpiredPenalty)
	r.scheduler.StartAsync()
}

func (r *Router) saveLpsSnapshot() {
	r.getAllLpsReq <- struct{}{}
	lps := <-r.apiGetLpsRes
	lps_, err := json.Marshal(lps)

	if err != nil {
		log.Error("saveLpsSnapshot: failed to marshal lps")
		return
	}

	lpsSnapshot := &schema.PermaLpsSnapshot{Lps: string(lps_)}
	r.wdb.SavePermaLpsSnapshot(lpsSnapshot, nil)
}

func (r *Router) loadNFTWhiteList() {
	wls, err := r.wdb.LoadNFTWhiteList()
	if err == nil && len(wls) > 0 {
		addrs := []string{}
		for _, w := range wls {
			addrs = append(addrs, w.UserAddr)
		}
		r.NFTInfo.SetWhitelist(addrs)
	}
}

func (r *Router) cleanUpExpiredPenalty() {
	r.penalty.ClearUpExpired()
}
