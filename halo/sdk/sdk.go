package sdk

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/logger"
)

var log = logger.New("sdk")

type SDK struct {
	dapp         string
	chainID      string
	feeRecipient string
	fee          string
	signerType   string // ecc, rsa
	signer       interface{}

	AccId string
	Cli   *Client

	lastNonce    int64 // last everTx used nonce
	sendTxLocker sync.Mutex
}

func New(signer interface{}, haloUrl string) (*SDK, error) {
	signerType, signerAddr, err := reflectSigner(signer)
	if err != nil {
		return nil, err
	}

	sdk := &SDK{
		signer:       signer,
		signerType:   signerType,
		AccId:        signerAddr,
		Cli:          NewClient(haloUrl),
		lastNonce:    time.Now().UnixNano() / 1000000,
		sendTxLocker: sync.Mutex{},
	}
	err = sdk.updateInfo()
	if err != nil {
		return nil, err
	}

	return sdk, nil
}

func (s *SDK) updateInfo() error {
	info, err := s.Cli.GetInfo()
	if err != nil {
		return err
	}
	s.dapp = info.Dapp
	s.chainID = info.ChainID
	s.feeRecipient = info.FeeRecipient

	// todo
	s.fee = "0"

	return nil
}

func (s *SDK) Transfer(to, amount string) (*schema.Transaction, error) {
	params := schema.TxTransferParams{
		To:     to,
		Amount: amount,
	}
	by, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return s.sendTx(schema.TxActionTransfer, string(by))
}

func (s *SDK) Unstake(stakePool, amount string) (*schema.Transaction, error) {
	params := schema.TxUnstakeParams{
		StakePool: stakePool,
		Amount:    amount,
	}
	by, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return s.sendTx(schema.TxActionUnstake, string(by))
}

func (s *SDK) Stake(stakePool, amount string) (*schema.Transaction, error) {
	params := schema.TxStakeParams{
		StakePool: stakePool,
		Amount:    amount,
	}
	by, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return s.sendTx(schema.TxActionStake, string(by))
}

func (s *SDK) Join(routerState schema.RouterState) (*schema.Transaction, error) {
	by, err := json.Marshal(routerState)
	if err != nil {
		return nil, err
	}

	return s.sendTx(schema.TxActionJoin, string(by))
}
func (s *SDK) Leave() (*schema.Transaction, error) {
	return s.sendTx(schema.TxActionLeave, "")
}

func (s *SDK) getNonce() int64 {
	for {
		newNonce := time.Now().UnixNano() / 1000000
		if newNonce > s.lastNonce {
			s.lastNonce = newNonce
			return newNonce
		}
	}
}

func (s *SDK) sendTx(action, params string) (*schema.Transaction, error) {
	s.sendTxLocker.Lock()
	defer s.sendTxLocker.Unlock()

	// assemble tx
	hTx := schema.Transaction{
		Dapp:         s.dapp,
		ChainID:      s.chainID,
		Action:       action,
		From:         s.AccId,
		Fee:          s.fee,
		FeeRecipient: s.feeRecipient,
		Nonce:        fmt.Sprintf("%d", s.getNonce()),
		Params:       params,
		Version:      schema.TxVersionV1,
		Sig:          "",
	}

	sign, err := s.Sign(hTx.String())
	if err != nil {
		log.Error("Sign failed", "error", err)
		return &hTx, err
	}
	hTx.Sig = sign

	// submit to everpay server
	everhash, err := s.Cli.SubmitTx(hTx)
	if err != nil {
		log.Error("submit hTx", "error", err)
		return &hTx, err
	}
	hTx.EverHash = everhash
	return &hTx, nil
}
