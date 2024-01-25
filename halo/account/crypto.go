package account

import (
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/everFinance/goar/utils"
)

func DecodeEthSig(sig string) []byte {
	return common.FromHex(sig)
}

func DecodeArSig(sig string) (s []byte, pub *rsa.PublicKey, addr string, err error) {
	ss := strings.Split(sig, ",")
	if len(ss) != 2 {
		err = fmt.Errorf("invalid length of sig:%s", sig)
		return
	}

	addr, err = utils.OwnerToAddress(ss[1])
	if err != nil {
		return
	}

	pub, err = utils.OwnerToPubKey(ss[1])
	if err != nil {
		return
	}

	s, err = utils.Base64Decode(ss[0])
	return
}
