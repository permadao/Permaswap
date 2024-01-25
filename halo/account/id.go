package account

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/everFinance/goar/utils"
)

func IDCheck(id string) (accountType, accid string, err error) {
	if common.IsHexAddress(id) {
		return AccountTypeEVM, common.HexToAddress(id).String(), nil
	}

	if IsArAddress(id) {
		return AccountTypeAR, id, nil
	}

	return "", "", ERR_INVALID_ID
}

func IsArAddress(addr string) bool {
	if len(addr) != 43 {
		return false
	}
	_, err := utils.Base64Decode(addr)
	if err != nil {
		return false
	}

	return true
}
