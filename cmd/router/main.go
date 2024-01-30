package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/everFinance/goether"
	"github.com/everVision/everpay-kits/sdk"
	halosdk "github.com/permadao/permaswap/halo/sdk"

	"github.com/permadao/permaswap/router"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "router",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "port", Value: ":8080", EnvVars: []string{"PORT"}},
			&cli.StringFlag{Name: "pay", Value: "https://api-dev.everpay.io", Usage: "pay url", EnvVars: []string{"PAY"}},
			&cli.StringFlag{Name: "nft", Value: "", Usage: "nft api url", EnvVars: []string{"NFT"}},
			&cli.Int64Flag{Name: "eth_chain_id", Value: 5, Usage: "eth chainId", EnvVars: []string{"ETH_CHAIN_ID"}},
			&cli.StringFlag{Name: "mysql", Value: "root@tcp(127.0.0.1:3306)/perma?charset=utf8mb4&parseTime=True&loc=Local", Usage: "mysql dsn", EnvVars: []string{"MYSQL"}},
			&cli.StringFlag{Name: "ecc_private", Value: "", Usage: "ecc custodian private", EnvVars: []string{"ECC_PRIVATE"}},
			&cli.StringFlag{Name: "name", Value: "perma", Usage: "router name", EnvVars: []string{"ROUTER_NAME"}},
			&cli.StringFlag{Name: "domain", Value: "", Usage: "router domain", EnvVars: []string{"ROUTER_DOMAIN"}},

			// halo
			&cli.StringFlag{Name: "halo_genesis_tx", Value: "", Usage: "halo genesis tx everhash", EnvVars: []string{"HALO_GENESIS_TX"}},
			&cli.StringFlag{Name: "halo_api_url_prefix", Value: "", Usage: "halo api url prefix", EnvVars: []string{"HALO_API_URL_PREFIX"}},
			&cli.StringFlag{Name: "halo_url", Value: "", Usage: "default halo tx submit url", EnvVars: []string{"HALO_URL"}},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	signer, err := goether.NewSigner(c.String("ecc_private"))
	if err != nil {
		panic(err)
	}
	everSDK, err := sdk.New(signer, c.String("pay"))
	if err != nil {
		panic(err)
	}

	var haloSDK *halosdk.SDK
	if c.String("halo_url") != "" {
		haloSDK, err = halosdk.New(signer, c.String("halo_url"))
		if err != nil {
			panic(err)
		}
	}

	r := router.New(c.String("name"), c.String("domain"), c.Int64("eth_chain_id"), everSDK, c.String("nft"), c.String("mysql"), c.String("halo_genesis_tx"), haloSDK, false)
	if c.String("halo_genesis_tx") != "" {
		r.Run(c.String("port"), c.String("halo_api_url_prefix"))
	} else {
		r.Run(c.String("port"), "")
	}

	<-signals
	r.Close()

	return nil
}
