package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"os"

	apd "github.com/cockroachdb/apd/v3"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/permadao/permaswap/core"
	coreSchema "github.com/permadao/permaswap/core/schema"
	routerSchema "github.com/permadao/permaswap/router/schema"

	"github.com/urfave/cli/v2"
	"gopkg.in/h2non/gentleman.v2"
)

var (
	ErrMissParam    = errors.New("err_miss_param")
	ErrInvalidParam = errors.New("err_invalid_param")
)

func main() {
	app := &cli.App{
		Name: "lpconfig",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "info", Aliases: []string{"i"}, Value: false, Usage: "swap info"},
			&cli.StringFlag{Name: "router", Aliases: []string{"r"}, Value: "", Usage: "perma router http url"},
			&cli.StringFlag{Name: "network", Aliases: []string{"n"}, Value: "mainnet", Usage: "nework: testnet or mainnet"},
			&cli.StringFlag{Name: "pool_id", Aliases: []string{"p"}, Usage: "pool id"},
			&cli.BoolFlag{Name: "full_range", Aliases: []string{"f"}, Value: false, Usage: "use full range price"},
			&cli.StringFlag{Name: "low_price", Aliases: []string{"low"}, Usage: "lowest price"},
			&cli.StringFlag{Name: "current_price", Aliases: []string{"current"}, Usage: "current price"},
			&cli.StringFlag{Name: "high_price", Aliases: []string{"high"}, Usage: "highest price"},
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Usage: "the config file"},

			&cli.StringFlag{Name: "amount_x", Aliases: []string{"x"}, Usage: "the amount of token x"},
			&cli.StringFlag{Name: "amount_y", Aliases: []string{"y"}, Usage: "the amount of token y"},

			&cli.StringFlag{Name: "price_direction", Value: "both", Aliases: []string{"d"}, Usage: "price direction: both or up or down"},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getSwapInfo(url string) (info routerSchema.InfoRes, err error) {
	cli := gentleman.New().URL(url)
	req := cli.Request()
	req.Path("/info")

	res, err := req.Send()
	if err != nil {
		return
	}
	defer res.Close()

	err = json.Unmarshal(res.Bytes(), &info)
	return
}

func getRevisedSqrtPrice(price string, decimalsX, decimalsY int) (*apd.Decimal, error) {
	price_, ok := new(big.Float).SetString(price)
	if !ok {
		return nil, ErrInvalidParam
	}
	factor := new(big.Float).SetFloat64(math.Pow(10, float64((decimalsY - decimalsX))))
	return core.SqrtPrice(new(big.Float).Mul(price_, factor).String())
}

func getRevisedSqrtPrice2(price string, decimalsX, decimalsY int) (*apd.Decimal, error) {
	price_, _, err := new(apd.Decimal).SetString(price)
	if err != nil {
		return nil, ErrInvalidParam
	}

	content := apd.BaseContext.WithPrecision(core.PRECISION)
	factor := new(apd.Decimal)
	ten := new(apd.Decimal).SetInt64(10)

	exponent := new(apd.Decimal).SetInt64(int64(decimalsY - decimalsX))
	_, err = content.Pow(factor, ten, exponent)
	if err != nil {
		return nil, ErrInvalidParam
	}
	_, err = content.Mul(price_, price_, factor)
	if err != nil {
		return nil, ErrInvalidParam
	}
	return core.SqrtPrice(price_.String())
}

func getRevisedAmount(amount string, decimals int, isMul bool) (string, error) {
	amount_, ok := new(big.Float).SetString(amount)
	if !ok {
		return "", ErrInvalidParam
	}
	factor := new(big.Float).SetFloat64(math.Pow(10, float64(decimals)))
	if isMul {
		return new(big.Float).Mul(amount_, factor).String(), nil
	}
	return new(big.Float).Quo(amount_, factor).String(), nil
}

func getRevisedAmount2(amount string, decimals int, isMul bool) (string, error) {
	amount_, _, err := new(apd.Decimal).SetString(amount)
	if err != nil {
		return "", ErrInvalidParam
	}

	content := apd.BaseContext.WithPrecision(core.PRECISION)
	factor := new(apd.Decimal)
	ten := new(apd.Decimal).SetInt64(10)

	exponent := new(apd.Decimal).SetInt64(int64(decimals))
	_, err = content.Pow(factor, ten, exponent)
	if err != nil {
		return "", ErrInvalidParam
	}

	if isMul {
		_, err = content.Mul(amount_, amount_, factor)
		if err != nil {
			return "", ErrInvalidParam
		}
	} else {
		_, err = content.Quo(amount_, amount_, factor)
		if err != nil {
			return "", ErrInvalidParam
		}
	}
	return amount_.String(), nil
}

func loadConfig(configPath string) (msgs []routerSchema.LpMsgAdd, err error) {
	file, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}

	msgs = []routerSchema.LpMsgAdd{}
	if err = json.Unmarshal(data, &msgs); err != nil {
		return
	}

	return
}

func run(c *cli.Context) error {

	pay := ""
	perma := ""
	if c.String("network") == "testnet" {
		pay = "https://api-dev.everpay.io"
		perma = "https://router-dev.permaswap.network"
	} else if c.String("network") == "mainnet" {
		pay = "https://api.everpay.io"
		perma = "https://router.permaswap.network"
	} else {
		return ErrInvalidParam
	}

	if c.String("router") != "" {
		perma = c.String("router")
	}

	info, err := getSwapInfo(perma)
	if err != nil {
		return err
	}

	if c.Bool("info") {
		fmt.Print("Pool List:", "\n\n")
		for i, pool := range info.PoolList {
			fmt.Println("PoolID:", i)
			fmt.Print(pool, "\n\n")
		}
		return nil
	}

	if c.String("pool_id") == "" || c.String("current_price") == "" || c.String("config") == "" {
		return ErrMissParam
	}
	pool, ok := info.PoolList[c.String("pool_id")]
	if !ok {
		return ErrInvalidParam
	}

	if c.String("amount_x") == "" && c.String("amount_y") == "" {
		return ErrMissParam
	}

	client := sdk.NewClient(pay)
	tokens, err := client.GetTokens()
	if err != nil {
		fmt.Println("failed to get tokens info", "err", err)
	}
	decimalX := tokens[pool.TokenXTag].Decimals
	decimalY := tokens[pool.TokenYTag].Decimals

	currentSqrtPrice, err := getRevisedSqrtPrice2(c.String("current_price"), decimalX, decimalY)
	if err != nil {
		return err
	}

	var lowSqrtPrice, highSqrtPrice *apd.Decimal
	if c.Bool("full_range") {
		lowSqrtPrice, _, _ = new(apd.Decimal).SetString(coreSchema.FullRangeLowSqrtPrice)
		highSqrtPrice, _, _ = new(apd.Decimal).SetString(coreSchema.FullRangeHighSqrtPrice)
	} else {
		if c.String("low_price") == "" || c.String("high_price") == "" {
			return ErrMissParam
		}

		if c.String("low_price") == "zero" {
			lowSqrtPrice, _, _ = new(apd.Decimal).SetString(coreSchema.FullRangeLowSqrtPrice)
		} else {
			lowSqrtPrice, err = getRevisedSqrtPrice2(c.String("low_price"), decimalX, decimalY)
			if err != nil {
				return err
			}
		}

		if c.String("high_price") == "inf" {
			highSqrtPrice, _, _ = new(apd.Decimal).SetString(coreSchema.FullRangeHighSqrtPrice)
		} else {
			highSqrtPrice, err = getRevisedSqrtPrice2(c.String("high_price"), decimalX, decimalY)
			if err != nil {
				return err
			}
		}
	}

	priceDirection := c.String("price_direction")
	if priceDirection != coreSchema.PriceDirectionBoth && priceDirection != coreSchema.PriceDirectionUp && priceDirection != coreSchema.PriceDirectionDown {
		return ErrInvalidParam
	}

	//fmt.Println("Your lp price range:", "lowSqrtPrice", lowSqrtPrice, "currentSqrtPrice:", currentSqrtPrice, "highSqrtPrice", highSqrtPrice, "priceDirection", priceDirection)

	liquidity := ""
	amountX := c.String("amount_x")
	amountY := c.String("amount_y")
	if amountY == "" {
		amountX, err := getRevisedAmount2(amountX, decimalX, true)
		if err != nil {
			return err
		}
		liquidity, _ = core.LiquidityFromAmountX(lowSqrtPrice, currentSqrtPrice, highSqrtPrice, amountX)
	} else {
		amountY, err := getRevisedAmount2(amountY, decimalY, true)
		if err != nil {
			return err
		}
		liquidity, _ = core.LiquidityFromAmountY(lowSqrtPrice, currentSqrtPrice, highSqrtPrice, amountY)
	}

	amountX, amountY, err = core.LiquidityToAmount(liquidity, lowSqrtPrice, currentSqrtPrice, highSqrtPrice, priceDirection)
	if err != nil {
		return err
	}

	fmt.Println("Your LP config file:", c.String("config"))
	fmt.Println("Your LP liquidity:", liquidity)

	amountX, _ = getRevisedAmount2(amountX, decimalX, false)
	amountY, _ = getRevisedAmount2(amountY, decimalY, false)

	// use amount in cmd option
	if c.String("amount_x") != "" {
		x, _ := new(big.Float).SetString(c.String("amount_x"))
		x_, _ := new(big.Float).SetString(amountX)
		if x.Cmp(x_) == 1 {
			amountX = c.String("amount_x")
		}
	}

	if c.String("amount_y") != "" {
		y, _ := new(big.Float).SetString(c.String("amount_y"))
		y_, _ := new(big.Float).SetString(amountY)
		if y.Cmp(y_) == 1 {
			amountY = c.String("amount_y")
		}
	}

	fmt.Println("Your LP account balance must be >= amountX:", amountX, tokens[pool.TokenXTag].Symbol, "amountY:", amountY, tokens[pool.TokenYTag].Symbol)

	// output to config file
	msg := routerSchema.LpMsgAdd{
		TokenX:           pool.TokenXTag,
		TokenY:           pool.TokenYTag,
		FeeRatio:         pool.FeeRatio,
		CurrentSqrtPrice: currentSqrtPrice,
		LowSqrtPrice:     lowSqrtPrice,
		HighSqrtPrice:    highSqrtPrice,
		Liquidity:        liquidity,
		PriceDirection:   priceDirection,
	}
	msgs, err := loadConfig(c.String("config"))
	if err != nil {
		msgs = []routerSchema.LpMsgAdd{msg}
	} else {
		msgs = append(msgs, msg)
	}

	by, _ := json.MarshalIndent(msgs, "", "  ")
	ioutil.WriteFile(c.String("config"), by, 0644)

	return nil
}
