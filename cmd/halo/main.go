package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/everFinance/goether"
	"github.com/everVision/everpay-kits/sdk"
	"github.com/permadao/permaswap/halo"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "halo",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "port", Value: ":8080", EnvVars: []string{"PORT"}},
			&cli.StringFlag{Name: "mysql", Value: "root@tcp(127.0.0.1:3306)/halo?charset=utf8mb4&parseTime=True&loc=Local", Usage: "mysql dsn", EnvVars: []string{"MYSQL"}},
			&cli.StringFlag{Name: "pay", Value: "https://api-dev.everpay.io", Usage: "pay url", EnvVars: []string{"PAY"}},
			&cli.StringFlag{Name: "ecc_private", Value: "", Usage: "ecc custodian private", EnvVars: []string{"ECC_PRIVATE"}},
			&cli.StringFlag{Name: "genesis_tx", Value: "", Usage: "genesis tx everhash", EnvVars: []string{"GENESIS_TX"}},
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

	h := halo.New(c.String("genesis_tx"), c.String("mysql"), everSDK)
	h.Run(c.String("port"))

	<-signals
	//fmt.Println("halo is closing")
	h.Close()

	return nil
}
