package router

import (
	"encoding/json"
	"math"
	"math/big"
	"sync"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	"github.com/permadao/permaswap/router/schema"
)

type Stats struct {
	wdb   *WDB
	Price *Price

	lock sync.RWMutex

	AccIDToVolume24hs map[string][]*schema.Volume
	AccIDToRewards    map[string][]*schema.PermaLpReward
	AccIDToTVLs       map[string][]*schema.TVL

	PoolIDToVolume24h map[string]*schema.Volume
	PoolIDToTVL       map[string]*schema.TVL

	tokens map[string]*everSchema.Token
	pools  map[string]*coreSchema.Pool
}

func NewStats(tokens map[string]*everSchema.Token, pools map[string]*coreSchema.Pool, w *WDB) *Stats {
	return &Stats{
		Price: NewPrice(tokens),
		wdb:   w,

		AccIDToVolume24hs: make(map[string][]*schema.Volume),
		AccIDToRewards:    make(map[string][]*schema.PermaLpReward),
		AccIDToTVLs:       make(map[string][]*schema.TVL),
		PoolIDToVolume24h: make(map[string]*schema.Volume),
		PoolIDToTVL:       make(map[string]*schema.TVL),

		tokens: tokens,
		pools:  pools,
	}
}

func (s *Stats) Run() {
	log.Info("Stats running")
	s.Price.Run()

	// update volume
	go func() {
		for {
			err := s.updateVolume()
			if err != nil {
				log.Warn("Failed to update volume", "err", err)
				time.Sleep(1 * time.Minute)
			}
			time.Sleep(10 * time.Minute)
		}
	}()

	// update tvl
	go func() {
		for {
			err := s.updateTVL()
			if err != nil {
				log.Warn("Failed to update TVL", "err", err)
				time.Sleep(1 * time.Minute)
			}
			time.Sleep(2 * time.Minute)
		}
	}()

}

func (s *Stats) sumPermaVolumeResToVolume(res schema.SumPermaVolumeRes, swapCount int64) schema.Volume {
	var volumeInUSD, rewardInUSD float64

	tokenX := s.pools[res.PoolID].TokenXTag
	tokenY := s.pools[res.PoolID].TokenYTag

	if price, ok := s.Price.GetPrice(tokenX); ok {
		volumeInUSD = price * res.AmountX
	} else {
		if price, ok := s.Price.GetPrice(tokenY); ok {
			volumeInUSD = price * res.AmountY
		}
	}

	if price, ok := s.Price.GetPrice(tokenX); ok {
		rewardInUSD += price * res.RewardX
	}
	if price, ok := s.Price.GetPrice(tokenY); ok {
		rewardInUSD += price * res.RewardY
	}

	v := schema.Volume{
		Timestamp: time.Now().Unix(),
		PoolID:    res.PoolID,
		AccID:     res.AccID,
		LpID:      res.LpID,
		TokenX:    tokenX,
		TokenY:    tokenY,
		X:         res.AmountX,
		Y:         res.AmountY,
		USD:       volumeInUSD,
		RewardX:   res.RewardX,
		RewardY:   res.RewardY,
		RewardUSD: rewardInUSD,
		SwapCount: swapCount,
	}

	return v
}

func (s *Stats) updateVolume() (err error) {
	err = s.updateAccVolume()
	if err != nil {
		log.Warn("updateAccVolume failed", "error", err)
	}

	err = s.updatePoolVolume()
	if err != nil {
		log.Warn("updatePoolVolume failed", "error", err)
	}

	return
}

func (s *Stats) updateAccVolume() error {
	accIDToVolume24h := make(map[string][]*schema.Volume)
	accIDToRewards := make(map[string][]*schema.PermaLpReward)
	start := time.Now().Add(-24 * time.Hour)
	res, err := s.wdb.SumVolumesByTime(start, time.Now())
	if err != nil {
		return err
	}
	for _, sv := range res {
		poolID := sv.PoolID
		if _, ok := s.pools[poolID]; !ok {
			continue
		}
		v := s.sumPermaVolumeResToVolume(*sv, 0)
		accIDToVolume24h[sv.AccID] = append(accIDToVolume24h[sv.AccID], &v)
	}
	for accid := range accIDToVolume24h {
		if rewards, err := s.wdb.GetPermaRewards(accid, nil); err == nil {
			accIDToRewards[accid] = rewards
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.AccIDToVolume24hs = accIDToVolume24h
	s.AccIDToRewards = accIDToRewards

	return nil
}

func (s *Stats) updatePoolVolume() error {
	poolIDToVolume24h := make(map[string]*schema.Volume)

	start := time.Now().Add(-24 * time.Hour)
	res, err := s.wdb.SumPoolVolumesByTime(start, time.Now())
	if err != nil {
		return err
	}

	scRes, err := s.wdb.SumPoolSwapCountByTime(start, time.Now())
	if err != nil {
		return err
	}
	poolIDToSwapCount := make(map[string]int64)
	for _, sc := range scRes {
		poolIDToSwapCount[sc.PoolID] = sc.SwapCount
	}

	for _, sv := range res {
		poolID := sv.PoolID
		if _, ok := s.pools[poolID]; !ok {
			continue
		}
		swapCount, ok := poolIDToSwapCount[poolID]
		if !ok {
			swapCount = 0
		}
		v := s.sumPermaVolumeResToVolume(*sv, swapCount)
		poolIDToVolume24h[sv.PoolID] = &v
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.PoolIDToVolume24h = poolIDToVolume24h

	return nil
}

func (s *Stats) getLpTVL(lp coreSchema.Lp) (tvl *schema.TVL, err error) {
	x, y, err := core.LiquidityToAmount(lp.Liquidity.String(), lp.LowSqrtPrice, lp.CurrentSqrtPrice, lp.HighSqrtPrice, lp.PriceDirection)
	if err != nil {
		return
	}
	x_, _ := new(big.Float).SetString(x)
	y_, _ := new(big.Float).SetString(y)
	tokenX, ok := s.tokens[lp.TokenXTag]
	if !ok {
		return
	}
	tokenY, ok := s.tokens[lp.TokenYTag]
	if !ok {
		return nil, WsErrInvalidToken
	}
	if err != nil {
		return nil, WsErrInvalidToken
	}

	decFactorX := new(big.Float).SetFloat64(math.Pow(10, float64(tokenX.Decimals)))
	decFactorY := new(big.Float).SetFloat64(math.Pow(10, float64(tokenY.Decimals)))
	amountX := new(big.Float).Quo(x_, decFactorX)
	amountY := new(big.Float).Quo(y_, decFactorY)
	fx, _ := amountX.Float64()
	fy, _ := amountY.Float64()

	var usdX, usdY float64
	if price, ok := s.Price.GetPrice(lp.TokenXTag); ok {
		usdX = price * fx
	}
	if price, ok := s.Price.GetPrice(lp.TokenYTag); ok {
		usdY = price * fy
	}

	tvl = &schema.TVL{
		Timestamp: time.Now().Unix(),
		PoolID:    lp.PoolID,
		LpID:      lp.ID(),
		AccID:     lp.AccID,
		X:         fx,
		Y:         fy,
		USD:       usdX + usdY,
	}

	return
}

func (s *Stats) updateTVL() (err error) {
	lpSnapshot, err := s.wdb.LoadPermaLpsSnapshot()
	if err != nil {
		return err
	}
	var lps []coreSchema.Lp
	if err := json.Unmarshal([]byte(lpSnapshot.Lps), &lps); err != nil {
		log.Warn("failed to load lpsSnapshot")
		return err
	}

	accidToTVLs := map[string][]*schema.TVL{}
	for _, lp := range lps {
		tvl, err := s.getLpTVL(lp)
		if err != nil {
			continue
		}
		accidToTVLs[lp.AccID] = append(accidToTVLs[lp.AccID], tvl)
	}

	poolidToTVL := map[string]*schema.TVL{}
	for _, lp := range lps {
		tvl, err := s.getLpTVL(lp)
		if err != nil {
			continue
		}
		if t, ok := poolidToTVL[lp.PoolID]; ok {
			t.X += tvl.X
			t.Y += tvl.Y
			t.USD += tvl.USD
		} else {
			poolidToTVL[lp.PoolID] = &schema.TVL{
				Timestamp: tvl.Timestamp,
				PoolID:    tvl.PoolID,
				X:         tvl.X,
				Y:         tvl.Y,
				USD:       tvl.USD,
			}
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.AccIDToTVLs = accidToTVLs
	s.PoolIDToTVL = poolidToTVL

	return
}

func (s *Stats) GetTVLsByAccid(accid string) (tvls []schema.TVL) {
	tvls = []schema.TVL{}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if ts, ok := s.AccIDToTVLs[accid]; ok {
		for _, t := range ts {
			tvls = append(tvls, *t)
		}
	}

	return
}

func (s *Stats) GetTVLByPoolID(poolid string) (tvl schema.TVL) {
	tvl = schema.TVL{}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if t, ok := s.PoolIDToTVL[poolid]; ok {
		tvl = *t
	}
	return
}

func (s *Stats) GetVolumesByAccid(accid string) (volumes []schema.Volume) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	volumes = []schema.Volume{}
	if vs, ok := s.AccIDToVolume24hs[accid]; ok {
		for _, v := range vs {
			volumes = append(volumes, *v)
		}
	}

	return
}

func (s *Stats) GetRewardsByAccid(accid string) (rewards []schema.PermaLpReward) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	rewards = []schema.PermaLpReward{}
	if rs, ok := s.AccIDToRewards[accid]; ok {
		for _, r := range rs {
			rewards = append(rewards, *r)
		}
	}

	return
}

func (s *Stats) GetVolumeByPoolID(poolid string) (volume schema.Volume) {
	volume = schema.Volume{}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if v, ok := s.PoolIDToVolume24h[poolid]; ok {
		volume = *v
	}
	return
}

func (s *Stats) GetTotalVolume() (volume schema.Volume) {
	volume = schema.Volume{}
	for _, v := range s.PoolIDToVolume24h {
		volume.USD += v.USD
	}
	return
}

func (s *Stats) GetTotalTVL() (tvl schema.TVL) {
	tvl = schema.TVL{}
	for _, t := range s.PoolIDToTVL {
		tvl.USD += t.USD
	}
	return
}
