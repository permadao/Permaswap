package router

import (
	"crypto/sha256"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/everVision/everpay-kits/utils"

	coreSchema "github.com/permadao/permaswap/core/schema"
	everSchema "github.com/everVision/everpay-kits/schema"
)

func ConvertPathsToBundle(paths []coreSchema.Path, tokens map[string]*everSchema.Token, expireAt int64, salt string) (bundle everSchema.Bundle, err error) {
	items := []everSchema.BundleItem{}
	for _, p := range paths {
		if _, ok := tokens[p.TokenTag]; !ok {
			log.Warn("invalid token", "tag", p.TokenTag)
			err = WsErrInvalidToken
			return
		}

		items = append(items, everSchema.BundleItem{
			Tag:     p.TokenTag,
			ChainID: tokens[p.TokenTag].ChainID,
			From:    p.From,
			To:      p.To,
			Amount:  p.Amount,
		})
	}
	bundle = everSchema.Bundle{
		Items:      items,
		Expiration: expireAt,
		Salt:       salt,
		Version:    everSchema.BundleTxVersionV1,
	}
	return
}

func VerifyBundleAndPaths(bundle everSchema.Bundle, paths []coreSchema.Path, tokens map[string]*everSchema.Token) (err error) {
	// todo: recheck see if anything miss
	b, err := ConvertPathsToBundle(paths, tokens, bundle.Expiration, bundle.Salt)
	if err != nil {
		return
	}

	if b.HashHex() != bundle.HashHex() {
		return WsErrInvalidPathsOrBundle
	}

	return nil
}

func CalPrice(
	paths []coreSchema.Path, tokens map[string]*everSchema.Token,
	tokenInTag, tokenOutTag, userAddr string,
) (price, tokenInAmount, tokenOutAmount *big.Float) {
	tokenIn, ok := tokens[tokenInTag]
	if !ok {
		return
	}

	tokenOut, ok := tokens[tokenOutTag]
	if !ok {
		return
	}

	tokenInSum := big.NewFloat(0)
	tokenOutSum := big.NewFloat(0)
	_, addr, err := utils.IDCheck(userAddr)
	if err != nil {
		return
	}
	for _, path := range paths {

		amount := new(big.Int)
		amount, ok := amount.SetString(path.Amount, 10)
		if !ok {
			return
		}

		_, from, err := utils.IDCheck(path.From)
		if err != nil {
			return
		}
		if from == addr && path.TokenTag == tokenIn.Tag() {
			tokenInSum.Add(tokenInSum, new(big.Float).SetInt(amount))
			continue
		}
		_, to, err := utils.IDCheck(path.To)
		if err != nil {
			return
		}
		if to == addr && path.TokenTag == tokenOut.Tag() {
			tokenOutSum.Add(tokenOutSum, new(big.Float).SetInt(amount))
		}
	}

	if tokenIn.Decimals == 0 || tokenOut.Decimals == 0 || tokenOutSum.Cmp(big.NewFloat(0)) == 0 {
		log.Error("invalid price calculation")
		return big.NewFloat(0), big.NewFloat(0), big.NewFloat(0)
	}

	// cal price tokenOut/tokenIn
	tokenInDecFactor := new(big.Float).SetFloat64(math.Pow(10, float64(tokenIn.Decimals)))
	tokenOutDecFactor := new(big.Float).SetFloat64(math.Pow(10, float64(tokenOut.Decimals)))
	tokenInAmount = new(big.Float).Quo(tokenInSum, tokenInDecFactor)
	tokenOutAmount = new(big.Float).Quo(tokenOutSum, tokenOutDecFactor)
	price = new(big.Float).Quo(tokenInAmount, tokenOutAmount)
	return
}

func VerifySig(accType, accID string, msg, sig string, chainID int) (err error) {
	switch accType {
	case everSchema.AccountTypeEVM:
		hash := accounts.TextHash([]byte(msg))
		_, err = utils.Verify(accType, accID, sig, hash, chainID)
		return
	case everSchema.AccountTypeAR:
		hash := sha256.Sum256([]byte(msg))
		_, err = utils.Verify(accType, accID, sig, hash[:], chainID)
		return
	default:
		return everSchema.ERR_INVALID_ACCOUNT_TYPE
	}
}
