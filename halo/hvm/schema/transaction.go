package schema

import (
	"crypto/sha256"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	PoolInc  = "incentive"
	PoolEco  = "ecosystem"
	PoolInv  = "investor"
	PoolTeam = "team"
)

const (
	TxVersionV1 = "v1"
)

const (
	TxActionTransfer = "transfer"

	TxActionStake   = "stake"
	TxActionUnstake = "unstake"

	// router join or leave
	TxActionJoin  = "join"
	TxActionLeave = "leave"

	// proposal
	TxActionPropose = "propose"
	TxActionCall    = "call"

	TxActionSwap = "swap"
)

var TxActionsSupported = []string{
	TxActionTransfer,
	TxActionStake,
	TxActionUnstake,
	TxActionJoin,
	TxActionLeave,
	TxActionPropose,
	TxActionCall,
	TxActionSwap,
}

type Transaction struct {
	Dapp         string `json:"dapp"`
	ChainID      string `json:"chainID"`
	EverHash     string `json:"everHash"` // tx hash from everpay
	Router       string `json:"router"`   // tx from which router
	Action       string `json:"action"`
	From         string `json:"from"`
	Fee          string `json:"fee"`
	FeeRecipient string `json:"feeRecipient"`
	Nonce        string `json:"nonce"`
	Version      string `json:"version"`
	Params       string `json:"params"`
	Sig          string `json:"sig"`
}

func (t *Transaction) String() string {
	return "dapp:" + t.Dapp + "\n" +
		"chainID:" + t.ChainID + "\n" +
		"action:" + t.Action + "\n" +
		"from:" + t.From + "\n" +
		"fee:" + t.Fee + "\n" +
		"feeRecipient:" + t.FeeRecipient + "\n" +
		"nonce:" + t.Nonce + "\n" +
		"version:" + t.Version + "\n" +
		"params:" + t.Params + "\n"
}

func (t *Transaction) Hash() []byte {
	return accounts.TextHash([]byte(t.String()))
}

func (t *Transaction) HexHash() string {
	return hexutil.Encode(t.Hash())
}

func (t *Transaction) ArHash() []byte {
	msg := sha256.Sum256([]byte(t.String()))
	return msg[:]
}

type TxTransferParams struct {
	To     string `json:"To"`
	Amount string `json:"Amount"`
}

type TxStakeParams struct {
	StakePool string `json:"StakePool"`
	Amount    string `json:"Amount"`
}

type TxUnstakeParams struct {
	StakePool string `json:"StakePool"`
	Amount    string `json:"Amount"`
}

type TxProposeParams struct {
	Name                  string   `json:"name"`
	Category              string   `json:"category"`
	Description           string   `json:"description"`
	Start                 int64    `json:"start"`
	End                   int64    `json:"end"`
	RunTimes              int64    `json:"runTimes"`
	Source                string   `json:"source"`
	InitData              string   `json:"initData"`
	OnlyAcceptedTxActions []string `json:"onlyAcceptedTxActions"`
}

type TxCallParams struct {
	ProposalID string `json:"proposalID"` // proposal id is hexhash
	Function   string `json:"function"`
	Params     string `json:"params"`
}

type SwapOrderItem struct {
	PoolID    string   `json:"poolID"`
	User      string   `json:"user"`
	Lp        string   `json:"lp"`
	TokenIn   string   `json:"tokenIn"`
	AmountIn  *big.Int `json:"amountIn"`
	TokenOut  string   `json:"tokenOut"`
	AmountOut *big.Int `json:"amountOut"`
}

type SwapOrder struct {
	User         string           `json:"user"`
	TimeStamp    int64            `json:"timeStamp"`
	Items        []*SwapOrderItem `json:"items"`
	FeeRecipient string           `json:"feeRecipient"`
	Fee          string           `json:"fee"`
	Index        int64            `json:"index"` // index of the error orderItem
	Err          string           `json:"err"`
}

func (o *SwapOrder) Success() bool {
	return o.Err == ""
}

type TxSwapParams struct {
	InternalStatus string `json:"internalStatus"`
	TxData         string `json:"txData"`
}

type TxApply struct {
	Tx     Transaction `json:"tx"`
	DryRun bool        `json:"dryRun"`
}
