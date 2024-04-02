package schema

import (
	apd "github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	Fee01   = "0.01"
	Fee003  = "0.003"
	Fee001  = "0.001"
	Fee0005 = "0.0005"
)

var MinAmountInsForPriceQuery = map[string]string{
	// chainID:1
	"ethereum-eth-0x0000000000000000000000000000000000000000":                                                    "100000000000000",     // 0.0001 eth
	"ethereum-usdc-0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48":                                                   "100000",              // 0.1 usdc
	"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0x4fadc7a98f2dc96510e42dd1a74141eeae0c1543": "10000000000",         // 0.01 ar;
	"ethereum-usdt-0xdac17f958d2ee523a2206206994597c13d831ec7":                                                   "100000",              // 0.1 usdt
	"arweave-ardrive--8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ":                                                "100000000000000000",  // 0.1 ardrive
	"everpay-acnh-0x72247989079da354c9f0a6886b965bcc86550f8a":                                                    "10000000",            // 0.1 achn
	"ethereum-ans-0x937efa4a5ff9d65785691b70a1136aaf8ada7e62":                                                    "100000000000000000",  // 0.1 ans
	"arweave-u-KTzTXT_ANmF84fWEKHzWURD1LWd9QaFR9yfYUwH2Lxw":                                                      "100000",              // 0.1 u
	"arweave-stamp-TlqASNDLA1Uh8yFiH-BzR_1FDag4s735F3PoUFEv2Mo":                                                  "100000000000",        // 0.1 stamp
	"everpay-acnh-0xac4cbc2009cf9ad96c2e1a4b34b1c8cb312fbce4":                                                    "10000000",            // 0.1 achn
	"ethereum-map-0x9e976f211daea0d652912ab99b0dc21a7fd728e4":                                                    "1000000000000000000", // 1 map
	"bsc-usdt-0x55d398326f99059ff775485246999027b3197955":                                                        "100000000000000000",  // 0.1 busdt
	"aostest-aocred-Sa0iBLPNyJQrwpTTG-tWLQU-1QeUAJA73DdxGGiKoJc":                                                 "5000",                // 5 aocred
	"aostest-trunk-OT9qTE2467gcozb2g8R6D6N3nQS94ENcaAIJfUzHCww":                                                  "3000",                // 3 aocred
	"aostest-0rbt-BUhZLMwQ6yZHguLtJYA5lLUa9LQzLXMXRfaq9FVcPJc":                                                   "1000000000000",       // 1 0rbt
	"psntest-halo-0x0000000000000000000000000000000000000000":                                                    "1000000000000000000", // 1 halo
	// chainID:5
	"ethereum-usdc-0xb7a4f3e9097c08da09517b5ab877f7a917224ede":                                                   "100000",      // 0.1 usdc
	"ethereum-usdt-0xd85476c906b5301e8e9eb58d174a6f96b9dfc5ee":                                                   "100000",      // 0.1 usdt
	"arweave,ethereum-ar-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA,0xcc9141efa8c20c7df0778748255b1487957811be": "10000000000", // 0.01 ar;
	//"bsc-psn-0xeb999d021591649f3cfc402e59439c58e8ac8bf4":                                                         "1000000000000000000", // 1 psn
	//"bsc-war-0xac213e987d24dfce13d90c1c724892758fb92af1":                                                         "10000000000000000",   // 0.01 war
	//"arweave-ardrive--8A6RexFkpfWwuyVO98wzSFZh0d6VJuI-buTJvlwOJQ":                                                "100000000000000000", // 0.1 ardrive
	"bsc-tusdc-0xf17a50ecc5fe5f476de2da5481cdd0f0ffef7712":    "100000",             // 0.1 tusdc
	"bsc-tar-0xf1458ee7e9a2096bce7a21c160840a3a291bcb55":      "10000000000",        // 0.01 ar;
	"bsc-tardrive-0xf4233b165f1b8da4f9aa94abc35c9ad2a7612979": "100000000000000000", // 0.1 ardrive
}

type Pool struct {
	TokenXTag string         `json:"tokenXTag" toml:"x"`
	TokenYTag string         `json:"tokenYTag" toml:"y"`
	FeeRatio  *apd.Decimal   `json:"feeRatio" toml:"fee_ratio"`
	Lps       map[string]*Lp `json:"-"`
}

func (pool *Pool) String() string {
	return "TokenXTag:" + pool.TokenXTag + "\n" +
		"TokenYTag:" + pool.TokenYTag + "\n" +
		"FeeRatio:" + pool.FeeRatio.Text('f')
}

func (pool *Pool) ID() string {
	h := accounts.TextHash([]byte(pool.String()))
	return hexutil.Encode(h)
}
