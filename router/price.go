package router

import (
	"fmt"
	"strings"
	"sync"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	"github.com/tidwall/gjson"
	"gopkg.in/h2non/gentleman.v2"
)

type Price struct {
	lock sync.RWMutex

	tokens          map[string]*everSchema.Token
	tokenTagToPrice map[string]float64
}

func NewPrice(tokens map[string]*everSchema.Token) *Price {
	return &Price{
		tokens:          tokens,
		tokenTagToPrice: make(map[string]float64),
	}
}

func (p *Price) Run() {
	go func() {
		for {
			p.updatePrice()
			time.Sleep(2 * time.Minute)
		}
	}()
}

func (p *Price) updatePrice() (err error) {
	tokenTagToPrice := map[string]float64{}
	for _, token := range p.tokens {

		price := float64(0)
		symbol := token.Symbol
		if symbol == "tUSDC" {
			price, err = GetTokenPriceByRedstone("USDC", "USDC", "")
		} else if symbol == "tAR" {
			price, err = GetTokenPriceByRedstone("AR", "USDC", "")
		} else if symbol == "tARDRIVE" {
			price, err = 3.5, nil
		} else {
			price, err = GetTokenPriceByRedstone(symbol, "USDC", "")
		}

		if err == nil {
			tokenTagToPrice[token.Tag()] = price
		} else {
			log.Error("Failed to get price", "token", token.Symbol, "err", err)
		}
		time.Sleep(1 * time.Second)
	}
	log.Debug("Updated price", "price", tokenTagToPrice)
	p.lock.Lock()
	defer p.lock.Unlock()
	p.tokenTagToPrice = tokenTagToPrice

	return
}

func (p *Price) GetPrice(tokenTag string) (price float64, ok bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	price, ok = p.tokenTagToPrice[tokenTag]
	return
}

func GetTokenPriceByRedstone(tokenSymbol string, currency string, timestamp string) (float64, error) {
	cli := gentleman.New()
	cli.URL("https://api.redstone.finance")
	req := cli.Request()
	req.AddPath("/prices")
	req.AddQuery("symbols", fmt.Sprintf("%s,%s", strings.ToUpper(tokenSymbol), strings.ToUpper(currency)))
	req.AddQuery("provider", "redstone")
	if timestamp != "" {
		req.AddQuery("toTimestamp", timestamp)
	}

	resp, err := req.Send()
	if err != nil {
		return 0.0, err
	}

	if !resp.Ok {
		return 0.0, fmt.Errorf("get token: %s currency: %s prices from redstone failed", tokenSymbol, currency)
	}
	defer resp.Close()
	tokenJsonPath := fmt.Sprintf("%s.value", strings.ToUpper(tokenSymbol))
	currencyJsonPath := fmt.Sprintf("%s.value", strings.ToUpper(currency))
	prices := gjson.GetManyBytes(resp.Bytes(), tokenJsonPath, currencyJsonPath)
	if len(prices) != 2 {
		return 0.0, fmt.Errorf("get token: %s currency: %s prices from redstone failed, response price number incorrect", tokenSymbol, currency)
	}
	tokenPrice := prices[0].Float()
	currencyPrice := prices[1].Float()
	if currencyPrice <= 0.0 {
		return 0.0, fmt.Errorf("get currency: %s price from redstone less than 0.0; currencyPrice: %f", currency, currencyPrice)
	}
	price := tokenPrice / currencyPrice
	return price, nil
}
