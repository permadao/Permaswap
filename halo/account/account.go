package account

import (
	"fmt"
	"strconv"

	"github.com/everFinance/goether"
	"github.com/permadao/permaswap/halo/logger"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/everFinance/goar/utils"
)

var log = logger.New("account")

const (
	AccountTypeEVM = "ethereum"
	AccountTypeAR  = "arweave"
)

type Transaction struct {
	Nonce string
	Hash  []byte
	Sig   string
}

type Account struct {
	ID    string // ID is eth address, notice: Case Sensitive
	Type  string
	Nonce int64
}

func New(id string) (*Account, error) {
	accType, accid, err := IDCheck(id)
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:    accid,
		Type:  accType,
		Nonce: 0,
	}, nil
}

func (a *Account) UpdateNonce(nonce int64) {
	a.Nonce = nonce
}

// Verify & return transaction's nonce
func (a *Account) Verify(tx Transaction) (nonce int64, err error) {
	nonce, err = strconv.ParseInt(tx.Nonce, 10, 64)
	if err != nil {
		log.Error("invalid nonce", "nonce", tx.Nonce, "err", err)
		return 0, ERR_INVALID_NONCE
	}

	if nonce <= a.Nonce {
		log.Warn("nonce too low", "accountNonce", a.Nonce, "nonce", nonce)
		return 0, ERR_NONCE_TOO_LOW
	}

	if err = a.VerifySig(tx); err != nil {
		log.Error("invalid signature", "err", err)
		return 0, ERR_INVALID_SIGNATURE
	}

	return
}

// only verify signature
func (a *Account) VerifySig(tx Transaction) (err error) {
	switch a.Type {

	case AccountTypeEVM:
		sig := DecodeEthSig(tx.Sig)

		_, addr, err := goether.Ecrecover(tx.Hash, sig)
		if err != nil {
			return fmt.Errorf("ecrecover failed, hash:%s, sig:%s, err:%v", hexutil.Encode(tx.Hash), tx.Sig, err)
		}

		if addr.String() != a.ID {
			return fmt.Errorf("address not equal, ecAddr:%s, accid:%s", addr.String(), a.ID)
		}
		return nil

	case AccountTypeAR:
		sig, pubKey, addr, err := DecodeArSig(tx.Sig)
		if err != nil {
			return fmt.Errorf("decode sig failed, sig:%s, err:%v", tx.Sig, err)
		}

		if addr != a.ID {
			return fmt.Errorf("address not equal, ownerAddr:%s, accid:%s", addr, a.ID)
		}

		return utils.Verify(tx.Hash, pubKey, sig)

	default:
		return fmt.Errorf("not support account type: %v", a.Type)

	}
}
