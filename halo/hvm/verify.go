package hvm

import (
	"encoding/json"
	"strconv"

	"math/big"

	everSchema "github.com/everVision/everpay-kits/schema"
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
	hash := []byte{}
	switch acc.Type {
	case everSchema.AccountTypeEVM:
		hash = tx.Hash()
	case everSchema.AccountTypeAR:
		hash = tx.ArHash()
	default:
		return nil, 0, nil, schema.ErrInvalidAccountType
	}

	nonce, err = acc.Verify(account.Transaction{
		Nonce: tx.Nonce,
		Hash:  hash,
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
	if len(routerState.Desc) > 250 {
		log.Error("info of join tx is too long", "description", routerState.Desc)
		return nil, schema.ErrInvalidTxParams
	}
	if routerState.Name == "" || len(routerState.Name) > 20 {
		log.Error("invalid router name in join tx", "name", routerState.Name)
		return nil, schema.ErrInvalidRouterName
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

func findPool(x, y string, pools map[string]*schema.Pool) (pool *schema.Pool) {
	for _, pool := range pools {
		if pool.TokenXTag == x && pool.TokenYTag == y {
			return pool
		}
		if pool.TokenYTag == x && pool.TokenXTag == y {
			return pool
		}
	}
	return
}

func TxSwapParamsVerify(txParams string, nonce string, pools map[string]*schema.Pool, tokens map[string]everSchema.TokenInfo) (order *schema.SwapOrder, Err error) {
	params := schema.TxSwapParams{}
	if err := json.Unmarshal([]byte(txParams), &params); err != nil {
		log.Error("invalid params of swap tx to unmarshal", "params", txParams, "err", err)
		return nil, schema.ErrInvalidTxParams
	}
	internalStatus := everSchema.InternalStatus{}
	if err := json.Unmarshal([]byte(params.InternalStatus), &internalStatus); err != nil {
		log.Error("failed to unmarshal swap tx internalStatus", "internalStatus", params.InternalStatus, "err", err)
		return nil, schema.ErrInvalidTxParams
	}
	bundleData := everSchema.BundleData{}
	if err := json.Unmarshal([]byte(params.TxData), &bundleData); err != nil {
		return nil, schema.ErrInvalidTxParams
	}
	nonce_, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		return nil, schema.ErrInvalidNonce
	}
	bundle := bundleData.Bundle.Bundle
	user := bundle.Items[0].From
	order = &schema.SwapOrder{
		User:      user,
		TimeStamp: nonce_ / 1000,
		Index:     -1,
	}
	if internalStatus.Status != everSchema.InternalStatusSuccess {
		order.Index = int64(internalStatus.Index / 2)
		order.Err = internalStatus.InternalErr.Msg
	}

	first := everSchema.BundleItem{}
	second := everSchema.BundleItem{}
	for i, item := range bundle.Items {
		if i%2 == 0 {
			first = item
			continue
		}
		second = item

		pool := findPool(first.Tag, second.Tag, pools)
		if pool == nil {
			log.Error("failed to find pool", "x", first.Tag, "y", second.Tag)
			return nil, schema.ErrNoPoolFound
		}
		if _, ok := tokens[first.Tag]; !ok {
			log.Error("failed to find first token", "tokenTag", first.Tag)
			return nil, schema.ErrNoTokenFound
		}
		if _, ok := tokens[second.Tag]; !ok {
			log.Error("failed to find second token", "tokenTag", second.Tag)
			return nil, schema.ErrNoTokenFound
		}

		lp := first.To
		tokenIn := first.Tag
		amountIn, _ := new(big.Int).SetString(first.Amount, 10)
		tokenOut := second.Tag
		amountOut, _ := new(big.Int).SetString(second.Amount, 10)
		if lp == user {
			lp = second.To
			tokenIn = second.Tag
			amountIn, _ = new(big.Int).SetString(second.Amount, 10)
			tokenOut = first.Tag
			amountOut, _ = new(big.Int).SetString(first.Amount, 10)
		}
		order.Items = append(order.Items, &schema.SwapOrderItem{
			PoolID:    pool.ID(),
			User:      user,
			Lp:        lp,
			TokenIn:   tokenIn,
			AmountIn:  amountIn,
			TokenOut:  tokenOut,
			AmountOut: amountOut,
		})
	}
	feePath := bundle.Items[len(bundle.Items)-1]
	order.FeeRecipient = feePath.To
	order.Fee = feePath.Amount
	return order, nil
}
