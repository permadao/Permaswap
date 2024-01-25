package core

import (
	"fmt"

	"github.com/permadao/permaswap/core/schema"
)

const (
	MaxPoolPathLength = 3
)

func InitPools(chainID int64) (pools map[string]*schema.Pool) {
	switch chainID {

	case 1: // everPay mainnet
		// eth_usdt, _ := NewPool(
		// 	"ethereum-eth-0x0000000000000000000000000000000000000000",
		// 	"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		// 	schema.Fee003,
		// )
		// ar_usdt, _ := NewPool(
		// 	"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
		// 	"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
		// 	schema.Fee003,
		// )
		// ar_cfx, _ := NewPool(
		//	"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
		//	"conflux-cfx-0x0000000000000000000000000000000000000000",
		//	schema.Fee003,
		// )

		ar_ardrive, _ := NewPool(
			"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
			"arweave-ardrive--8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ",
			schema.Fee003,
		)

		eth_usdc, _ := NewPool(
			"ethereum-eth-0x0000000000000000000000000000000000000000",
			"ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			schema.Fee003,
		)

		ar_usdc, _ := NewPool(
			"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
			"ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			schema.Fee003,
		)

		ar_eth, _ := NewPool(
			"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
			"ethereum-eth-0x0000000000000000000000000000000000000000",
			schema.Fee003,
		)

		usdc_usdt, _ := NewPool(
			"ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7",
			schema.Fee0005,
		)

		usdc_acnh, _ := NewPool(
			"ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			"everpay-acnh-0x72247989079da354c9f0a6886b965bcc86550f8a",
			schema.Fee0005,
		)

		ar_ans, _ := NewPool(
			"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
			"ethereum-ans-0x937efa4a5ff9d65785691b70a1136aaf8ada7e62",
			schema.Fee003,
		)

		ar_u, _ := NewPool(
			"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
			"arweave-u-KTzTXT_ANmF84fWEKHzWURD1LWd9QaFR9yfYUwH2Lxw",
			schema.Fee003,
		)

		ar_stamp, _ := NewPool(
			"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543",
			"arweave-stamp-TlqASNDLA1Uh8yFiH-BzR_1FDag4s735F3PoUFEv2Mo",
			schema.Fee003,
		)

		eth_map, _ := NewPool(
			"ethereum-eth-0x0000000000000000000000000000000000000000",
			"ethereum-map-0x9e976f211daea0d652912ab99b0dc21a7fd728e4",
			schema.Fee003,
		)

		busdt_usdc, _ := NewPool(
			"bsc-usdt-0x55d398326f99059ff775485246999027b3197955",
			"ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			schema.Fee0005,
		)

		pools = map[string]*schema.Pool{
			// eth_usdt.ID():   eth_usdt,
			// ar_usdt.ID():    ar_usdt,
			//ar_cfx.ID():    ar_cfx,

			ar_ardrive.ID(): ar_ardrive,
			eth_usdc.ID():   eth_usdc,
			ar_eth.ID():     ar_eth,
			usdc_usdt.ID():  usdc_usdt,
			ar_usdc.ID():    ar_usdc,
			usdc_acnh.ID():  usdc_acnh,
			ar_ans.ID():     ar_ans,
			ar_u.ID():       ar_u,
			ar_stamp.ID():   ar_stamp,
			eth_map.ID():    eth_map,
			busdt_usdc.ID(): busdt_usdc,
		}

	case 5: // everPay testnet
		//eth_usdt, _ := NewPool(
		//	"ethereum-eth-0x0000000000000000000000000000000000000000",
		//	"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		//	schema.Fee003,
		//)

		//eth_usdc, _ := NewPool(
		//	"ethereum-eth-0x0000000000000000000000000000000000000000",
		//	"ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		//	schema.Fee003,
		//)

		//usdc_usdt, _ := NewPool(
		//	"ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede",
		//	"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee",
		//	schema.Fee0005,
		//)

		tar_tusdc, _ := NewPool(
			"bsc-tar-0xf1458ee7e9a2096bce7a21c160840a3a291bcb55",
			"bsc-tusdc-0xf17a50ecc5fe5f476de2da5481cdd0f0ffef7712",
			schema.Fee003,
		)

		tar_tardrive, _ := NewPool(
			"bsc-tar-0xf1458ee7e9a2096bce7a21c160840a3a291bcb55",
			"bsc-tardrive-0xf4233b165f1b8da4f9aa94abc35c9ad2a7612979",
			schema.Fee003,
		)

		tusdc_acnh, _ := NewPool(
			"bsc-tusdc-0xf17a50ecc5fe5f476de2da5481cdd0f0ffef7712",
			"everpay-acnh-0xac4cbc2009cf9ad96c2e1a4b34b1c8cb312fbce4",
			schema.Fee003,
		)

		// tardrive_tusdc, _ := NewPool(
		// 	"bsc-tardrive-0xf4233b165f1b8da4f9aa94abc35c9ad2a7612979",
		// 	"bsc-tusdc-0xf17a50ecc5fe5f476de2da5481cdd0f0ffef7712",
		// 	schema.Fee003,
		// )

		pools = map[string]*schema.Pool{
			//eth_usdt.ID():  eth_usdt,
			//usdc_usdt.ID(): usdc_usdt,
			//eth_usdc.ID():  eth_usdc,
			tusdc_acnh.ID():   tusdc_acnh,
			tar_tardrive.ID(): tar_tardrive,
			//tardrive_tusdc.ID(): tardrive_tusdc,
			tar_tusdc.ID(): tar_tusdc,
		}

	default:
		panic(fmt.Sprintf("can not init pools, invalid chainID: %d\n", chainID))
	}

	return
}
