package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/everFinance/goether"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/permadao/permaswap/halo"
	halosdk "github.com/permadao/permaswap/halo/sdk"

	"github.com/permadao/permaswap/router"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Router router.Config
	Halo   halo.Config
}

func main() {
	app := &cli.App{
		Name: "router",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Value: "", Usage: "router node toml config file", EnvVars: []string{"CONFIG"}},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	var config Config
	if _, err := toml.DecodeFile(c.String("config"), &config); err != nil {
		panic(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	signer, err := goether.NewSigner(config.Router.Pk)
	if err != nil {
		panic(err)
	}
	everSDK, err := sdk.New(signer, config.Router.EverpayApi)
	if err != nil {
		panic(err)
	}

	var haloSDK *halosdk.SDK
	if c.String("halo_url") != "" {
		haloSDK, err = halosdk.New(signer, config.Halo.DefaulHaloNodeUrl)
		if err != nil {
			panic(err)
		}
	}

	r := router.New(&config.Router, &config.Halo, everSDK, haloSDK, false)
	if c.String(config.Halo.Genesis) != "" {
		r.Run(config.Router.Port, config.Halo.UrlPrefix)
	} else {
		r.Run(config.Router.Port, "")
	}

	<-signals
	r.Close()

	return nil
}
