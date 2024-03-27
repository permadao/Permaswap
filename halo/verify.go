package halo

import (
	"encoding/json"
	"math/big"
	"strconv"
	"time"

	everSchema "github.com/everVision/everpay-kits/schema"
	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"

	"github.com/permadao/permaswap/halo/schema"
	"github.com/permadao/permaswap/halo/token"
)

func GenesisTxVerify(txResp everSchema.TxResponse) (state *hvmSchema.State, err error) {
	genesisTxData := schema.GenesisTxData{}

	if err := json.Unmarshal([]byte(txResp.Data), &genesisTxData); err != nil {
		return nil, err
	}
	if genesisTxData.Dapp == "" || genesisTxData.ChainID == "" || genesisTxData.Govern == "" || genesisTxData.FeeRecipient == "" {
		log.Error("Invalid genesis tx data")
		return nil, schema.ErrInvalidGenesisTx
	}

	totalSupply, err := totalSupplyVerify(genesisTxData.TokenTotalSupply)
	if err != nil {
		return nil, err
	}
	balance, err := balanceVerify(genesisTxData.TokenBalance, totalSupply)
	if err != nil {
		log.Error("balanceVerify failed", "err", err)
		return nil, schema.ErrInvalidGenesisBalance
	}
	stake, err := stakeVerify(genesisTxData.TokenStake, txResp.Nonce/1000)
	if err != nil {
		log.Error("stakeVerify failed", "err", err)
		return nil, schema.ErrInvalidGenesisStake
	}
	if genesisTxData.TokenSymbol == "" || genesisTxData.TokenDecimals <= 0 {
		log.Error("Invalid genesis tx data for token", "symbol", genesisTxData.TokenSymbol, "decimals", genesisTxData.TokenDecimals)
		return nil, schema.ErrInvalidGenesisTx
	}

	// todo: add stake info in genesis tx

	token := token.New(genesisTxData.TokenSymbol, genesisTxData.TokenDecimals, totalSupply, balance, stake)
	state = &hvmSchema.State{
		Dapp:             genesisTxData.Dapp,
		ChainID:          genesisTxData.ChainID,
		Govern:           genesisTxData.Govern,
		FeeRecipient:     genesisTxData.FeeRecipient,
		RouterMinStake:   genesisTxData.RouterMinStake,
		Routers:          genesisTxData.Routers,
		RouterStates:     genesisTxData.RouterStates,
		StakePools:       genesisTxData.StakePools,
		OnlyUnStakePools: genesisTxData.OnlyUnStakePools,
		Token:            token,
	}
	return state, nil
}

func totalSupplyVerify(totalSupply string) (*big.Int, error) {
	ts, ok := new(big.Int).SetString(totalSupply, 10)
	if !ok {
		return nil, schema.ErrInvalidGenesisTotalSupply
	}
	if ts.Cmp(big.NewInt(0)) != 1 {
		return nil, schema.ErrInvalidGenesisTotalSupply
	}
	return ts, nil
}

func balanceVerify(balances map[string]string, totalSupply *big.Int) (map[string]*big.Int, error) {
	total := big.NewInt(0)
	nbs := make(map[string]*big.Int)
	for address, balance := range balances {
		// _, _, err := account.IDCheck(address)
		// if err != nil {
		// 	return nil, schema.ErrInvalidGenesisBalance
		// }

		nb, ok := new(big.Int).SetString(balance, 10)
		if !ok {
			return nil, schema.ErrInvalidGenesisBalance
		}
		if nb.Cmp(big.NewInt(0)) < 0 {
			return nil, schema.ErrInvalidGenesisBalance
		}
		nbs[address] = nb
		total.Add(total, nb)
	}
	if total.Cmp(totalSupply) == 1 {
		log.Error("total supply is less than sum of balances", "totalSupply", totalSupply, "sumOfBalances", total)
		return nil, schema.ErrInvalidGenesisBalance
	}
	return nbs, nil
}

// todo: 1. verify stake  + token balances <= total supply 2. stake pool must be in stakePools
func stakeVerify(stakes map[string]map[string]string, timestamp int64) (map[string]map[string][]tokSchema.Stake, error) {
	nss := make(map[string]map[string][]tokSchema.Stake)
	for address, pools := range stakes {
		for pool, amount := range pools {
			na, ok := new(big.Int).SetString(amount, 10)
			if !ok {
				return nil, schema.ErrInvalidGenesisStake
			}
			if na.Cmp(big.NewInt(0)) < 0 {
				return nil, schema.ErrInvalidGenesisStake
			}

			stake := tokSchema.Stake{
				StakedAt: timestamp,
				Amount:   na,
			}

			if _, ok := nss[address]; !ok {
				nss[address] = make(map[string][]tokSchema.Stake)
				nss[address][pool] = []tokSchema.Stake{stake}
			} else {
				if _, ok := nss[address][pool]; !ok {
					nss[address][pool] = []tokSchema.Stake{stake}
				} else {
					nss[address][pool] = append(nss[address][pool], stake)
				}
			}
		}
	}
	return nss, nil
}

func verifyNonce(nonce string) error {
	n, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		log.Error("invalid nonce", "nonce", nonce, "err", err)
		return schema.ErrInvalidSubmitTxNonce
	}
	curTs := time.Now().UnixNano() / 1000000
	if n < curTs-200000 || n > curTs+200000 {
		log.Error("invalid nonce", "nonce", nonce, "curTimestamp", curTs)
		return schema.ErrInvalidSubmitTxNonce
	}
	return nil
}

func BundleTxVerify(txResp everSchema.TxResponse) (tx *hvmSchema.Transaction, err error) {
	tx = &hvmSchema.Transaction{}
	if txResp.Action != schema.EverTxActionBundle {
		return nil, schema.ErrInvalidBundleTxAction
	}
	tx.EverHash = txResp.EverHash
	tx.Router = txResp.From
	tx.Nonce = strconv.FormatInt(txResp.Nonce, 10)
	tx.Action = hvmSchema.TxActionSwap

	bundleData := everSchema.BundleData{}
	if err := json.Unmarshal([]byte(txResp.Data), &bundleData); err != nil {
		log.Error("unmarshal bundle tx data failed", "err", err)
		return nil, err
	}

	params := hvmSchema.TxSwapParams{
		InternalStatus: txResp.InternalStatus,
		TxData:         txResp.Data,
	}
	params_, err := json.Marshal(params)
	if err != nil {
		log.Error("marshal swap tx params failed", "err", err)
	}
	tx.Params = string(params_)

	return tx, nil
}
