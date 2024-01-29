package hvm

import (
	"encoding/json"
	"math/big"

	"github.com/permadao/permaswap/halo/account"
	"github.com/permadao/permaswap/halo/hvm/schema"
)

// todo: stake and vote
func (h *HVM) verifyProposer(proposer string) error {
	if proposer != h.Govern {
		return schema.ErrInvalidProposer
	}
	return nil
}

func (h *HVM) verifyFromRouter(router string) bool {
	for _, r := range h.Routers {
		if r == router {
			return true
		}
	}
	return false
}

func (h *HVM) TxVerify(tx schema.Transaction, dryRun bool) (acc *account.Account, nonce int64, fee *big.Int, err error) {
	if !h.verifyFromRouter(tx.Router) {
		return nil, 0, nil, schema.ErrInvalidFromRouter
	}

	if tx.Dapp != h.Dapp || tx.ChainID != h.ChainID || tx.Version != schema.TxVersionV1 {
		return nil, 0, nil, schema.ErrInvalidTxField
	}

	if tx.Action == "" || tx.From == "" || tx.Nonce == "" || tx.Sig == "" {
		return nil, 0, nil, schema.ErrInvalidTxField
	}

	if !dryRun && tx.EverHash == "" {
		return nil, 0, nil, schema.ErrInvalidTxField
	}

	// check tx fee && feeRecipient
	fee, ok := new(big.Int).SetString(tx.Fee, 10)
	if !ok {
		return nil, 0, nil, schema.ErrInvalidFee
	}
	if tx.FeeRecipient != h.FeeRecipient {
		return nil, 0, nil, schema.ErrInvalidFeeRecipient
	}

	// account
	acc, err = h.getOrCreateAccount(tx.From, dryRun)
	if err != nil {
		return nil, 0, nil, err
	}

	// verify tx nonce and signature
	nonce, err = acc.Verify(account.Transaction{
		Nonce: tx.Nonce,
		Hash:  tx.Hash(),
		Sig:   tx.Sig,
	})
	if err != nil {
		return nil, 0, nil, err
	}
	return acc, nonce, fee, nil
}

func TxTransferParamsVerify(txParams string) (to string, amount *big.Int, err error) {
	params := schema.TxTransferParams{}
	if err := json.Unmarshal([]byte(txParams), &params); err != nil {
		log.Error("invalid params of transfer tx to unmarshal", "params", txParams, "err", err)
		return "", nil, schema.ErrInvalidTxParams
	}

	_, to, err = account.IDCheck(params.To)
	if err != nil {
		log.Error("invalid to of transfer tx ", "to", params.To, "err", err)
		return "", nil, err
	}
	amount, ok := new(big.Int).SetString(params.Amount, 10)
	if !ok {
		log.Error("invalid amount of transfer tx ", "amount", params.Amount)
		return "", nil, schema.ErrInvalidAmount
	}
	return
}

func TxStakeParamsVerify(txParams string) (stakePool string, amount *big.Int, err error) {
	params := schema.TxStakeParams{}
	if err := json.Unmarshal([]byte(txParams), &params); err != nil {
		log.Error("invalid params of stake tx to unmarshal", "params", txParams, "err", err)
		return "", nil, schema.ErrInvalidTxParams
	}
	stakePool = params.StakePool
	if stakePool == "" {
		log.Error("invalid stakePool of stake tx ", "stakePool", params.StakePool)
		return "", nil, schema.ErrInvalidStakePool
	}
	amount, ok := new(big.Int).SetString(params.Amount, 10)
	if !ok {
		log.Error("invalid amount of stake tx ", "amount", params.Amount)
		return "", nil, schema.ErrInvalidAmount
	}
	return
}

func TxUnstakeParamsVerify(txParams string) (stakePool string, amount *big.Int, err error) {
	params := schema.TxUnstakeParams{}
	if err := json.Unmarshal([]byte(txParams), &params); err != nil {
		log.Error("invalid params of unstake tx to unmarshal", "params", txParams, "err", err)
		return "", nil, schema.ErrInvalidTxParams
	}
	stakePool = params.StakePool
	if stakePool == "" {
		log.Error("invalid stakePool of unstake tx ", "stakePool", params.StakePool)
		return "", nil, schema.ErrInvalidStakePool
	}
	amount, ok := new(big.Int).SetString(params.Amount, 10)
	if !ok {
		log.Error("invalid amount of unstake tx ", "amount", params.Amount)
		return "", nil, schema.ErrInvalidAmount
	}
	return
}

func TxJoinParamsVerify(txParams string) (routerState *schema.RouterState, err error) {
	if err := json.Unmarshal([]byte(txParams), &routerState); err != nil {
		log.Error("invalid params of join tx to unmarshal", "params", txParams, "err", err)
		return nil, schema.ErrInvalidTxParams
	}
	if len(routerState.Info) > 250 {
		log.Error("info of join tx is too long", "info", routerState.Info)
		return nil, schema.ErrInvalidTxParams
	}
	return
}

func TxProposeParamsVerify(params schema.TxProposeParams) (
	start, end, runTimes int64, source, initData string,
	onlyAcceptedTxActions []string, err error) {
	if params.Source == "" || params.Name == "" {
		log.Error("empty source or name of propose tx ", "source", params.Source, "name", params.Name)
		return 0, 0, 0, "", "", nil, schema.ErrInvalidTxParams
	}

	onlyAcceptedTxActions = []string{}
	if params.OnlyAcceptedTxActions != nil {
		for _, action := range params.OnlyAcceptedTxActions {
			if !InSlice(schema.TxActionsSupported, action) {
				log.Error("invalid tx action to only accepted of propose tx ", "action", action)
				return 0, 0, 0, "", "", nil, schema.ErrInvalidTxParams
			}
		}
		onlyAcceptedTxActions = params.OnlyAcceptedTxActions
	}

	if params.RunTimes == 0 {
		if params.Start == 0 || params.End == 0 || params.Start >= params.End {
			log.Error("invalid start or end of propose tx ", "start", params.Start, "end", params.End)
			return 0, 0, 0, "", "", nil, schema.ErrInvalidTxParams
		}
		return params.Start, params.End, 0, params.Source, params.InitData, onlyAcceptedTxActions, nil
	}
	if params.RunTimes < 1 {
		log.Error("invalid runTimes of propose tx ", "runTimes", params.RunTimes)
		return 0, 0, 0, "", "", nil, schema.ErrInvalidTxParams
	}

	return 0, 0, params.RunTimes, params.Source, params.InitData, onlyAcceptedTxActions, nil
}

func (h *HVM) ProposalVerify(tx schema.Transaction, nonce int64) (*schema.Proposal, error) {
	if err := h.verifyProposer(tx.From); err != nil {
		return nil, err
	}

	params := schema.TxProposeParams{}
	if err := json.Unmarshal([]byte(tx.Params), &params); err != nil {
		log.Error("invalid params of propose tx ", "params", tx.Params, "err", err)
		return nil, schema.ErrInvalidTxParams
	}

	start, end, runTimes, source, initData, onlyAcceptedTxActions, err := TxProposeParamsVerify(params)
	if err != nil {
		return nil, err
	}
	if start > 0 && nonce/1000 >= start {
		log.Error("invalid start of propose tx ", "start", start, "nonce", nonce)
		return nil, schema.ErrInvalidTxParams
	}
	executor, err := NewExecutor(source)
	if err != nil {
		return nil, err
	}
	proposal := NewProposal(params.Name, start, end, runTimes, source, initData, onlyAcceptedTxActions, executor)
	return proposal, nil
}

func TxCallParamsVerify(txParams string) (proposalID, function, callParams string, err error) {
	params := schema.TxCallParams{}
	if err := json.Unmarshal([]byte(txParams), &params); err != nil {
		log.Error("invalid params of call tx to unmarshal", "params", txParams, "err", err)
		return "", "", "", schema.ErrInvalidTxParams
	}
	proposalID = params.ProposalID
	if proposalID == "" {
		log.Error("no proposal id of call tx ")
		return "", "", "", schema.ErrInvalidTxParams
	}
	function = params.Function
	if function == "" {
		log.Error("no function of call tx ")
		return "", "", "", schema.ErrInvalidTxParams
	}
	return proposalID, function, params.Params, nil
}
