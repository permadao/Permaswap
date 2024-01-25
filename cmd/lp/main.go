package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/everFinance/goar"
	"github.com/everFinance/goether"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/permadao/permaswap/lp"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "lp",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "pay", Value: "https://api-dev.everpay.io", Usage: "pay url", EnvVars: []string{"PAY"}},
			&cli.StringFlag{Name: "perma_ws", Value: "wss://swap-dev.everpay.io/wslp", Usage: "perma router ws url", EnvVars: []string{"PERMA_WS"}},
			&cli.StringFlag{Name: "perma_http", Value: "https://swap-dev.everpay.io", Usage: "perma router http url", EnvVars: []string{"PERMA_HTTP"}},
			&cli.StringFlag{Name: "lp_config", Value: "./lp/test.json", Usage: "perma lp config", EnvVars: []string{"LP_CONFIG"}},
			&cli.Int64Flag{Name: "eth_chain_id", Value: 5, Usage: "eth chainId", EnvVars: []string{"ETH_CHAIN_ID"}},
			&cli.StringFlag{Name: "ecc_private", Value: "", Usage: "ecc custodian private", EnvVars: []string{"ECC_PRIVATE"}},
			&cli.StringFlag{Name: "ar_wallet", Usage: "arweave wallet json file path", EnvVars: []string{"AR_WALLET"}},
			&cli.BoolFlag{Name: "lp_api", Value: false, Usage: "enable lp api", EnvVars: []string{"LP_API"}},
		},
		Action: run,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) (err error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	var signer interface{}

	if c.String("ar_wallet") != "" {
		signer, err = goar.NewSignerFromPath(c.String("ar_wallet"))
	} else {
		signer, err = goether.NewSigner(c.String("ecc_private"))
	}

	if err != nil {
		panic(err)
	}
	everSDK, err := sdk.New(signer, c.String("pay"))
	if err != nil {
		panic(err)
	}

	fmt.Println("LP address:", everSDK.AccId)

	l := lp.New(c.Int64("eth_chain_id"), c.Bool("lp_api"),
		lp.NewRSDK(c.String("perma_ws"), c.String("perma_http"), everSDK),
	)
	l.Run(c.String("lp_config"))

	<-signals
	l.Close()

	return nil
}
