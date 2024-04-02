package hvm

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/permadao/permaswap/halo/account"
	"github.com/permadao/permaswap/halo/hvm/schema"
)

func (h *HVM) VerifyTx(tx schema.Transaction, oracle *schema.Oracle) (err error) {
	_, nonce, fee, err := h.TxVerify(tx, true)
	if err != nil {
		return err
	}

	switch tx.Action {
	case schema.TxActionTransfer:
		to, amount, err := TxTransferParamsVerify(tx.Params)
		if err != nil {
			return err
		}
		err = h.Token.Transfer(tx.From, to, amount, tx.FeeRecipient, fee, true)
		if err != nil {
			return err
		}
	case schema.TxActionPropose:
		_, err := h.ProposalVerify(tx, nonce)
		if err != nil {
			return err
		}

	case schema.TxActionStake:
		stakePool, amount, err := TxStakeParamsVerify(tx.Params)
		if err != nil {
			return err
		}

		// in stakepools, and not in depreated stakepools
		if !InSlice(h.StakePools, stakePool) {
			return schema.ErrInvalidStakePool
		}
		if InSlice(h.OnlyUnStakePools, stakePool) {
			return schema.ErrInvalidStakePool
		}

		stakeAt := nonce / 1000
		err = h.Token.Stake(tx.From, stakePool, amount, stakeAt, tx.FeeRecipient, fee, true)
		if err != nil {
			return err
		}

	case schema.TxActionUnstake:
		stakePool, amount, err := TxUnstakeParamsVerify(tx.Params)
		if err != nil {
			return err
		}

		if !InSlice(h.StakePools, stakePool) {
			return schema.ErrInvalidStakePool
		}
		err = h.Token.Unstake(tx.From, stakePool, amount, tx.FeeRecipient, fee, true)
		if err != nil {
			return err
		}
	case schema.TxActionJoin:
		if InSlice(h.Routers, tx.From) {
			return schema.ErrRouterAlreadyJoined
		}
		routerState, err := TxJoinParamsVerify(tx.Params)
		if err != nil {
			return err
		}
		if routerState.Router != tx.From {
			return schema.ErrInvalidRouterAddress
		}
		for _, r := range h.RouterStates {
			if strings.EqualFold(r.Name, routerState.Name) {
				log.Error("router name already exists", "name", routerState.Name)
				return schema.ErrInvalidRouterName
			}
		}
		staked := h.Token.TotalStaked(tx.From, "")
		staked_, _ := new(big.Int).SetString(staked, 10)
		routerMinStake, _ := new(big.Int).SetString(h.RouterMinStake, 10)
		if staked_.Cmp(routerMinStake) == -1 {
			return schema.ErrInsufficientStake
		}

	case schema.TxActionLeave:
		if !InSlice(h.Routers, tx.From) {
			return schema.ErrNotARouter
		}

	case schema.TxActionCall:
		proposalID, _, _, err := TxCallParamsVerify(tx.Params)
		if err != nil {
			return err
		}
		proposal := FindProposal(h.Proposals, proposalID)
		if proposal == nil {
			return schema.ErrNoProposalFound
		}

	default:
		return schema.ErrInvalidTxAction
	}
	return nil
}

func (h *HVM) ExecuteTx(tx schema.Transaction, oracle *schema.Oracle) (err error) {
	defer func() {
		if err != schema.ErrTxExecuted {
			if err != nil {
				h.Validity[tx.EverHash] = false
			} else {
				h.Validity[tx.EverHash] = true
			}
			h.Executed = append(h.Executed, tx.EverHash)
		}
	}()

	if _, ok := h.Validity[tx.EverHash]; ok {
		log.Error("tx have been executed", "everHash", tx.EverHash)
		return schema.ErrTxExecuted
	}

	var acc *account.Account
	var nonce int64
	var fee *big.Int
	var ProposalCalled *schema.Proposal
	if tx.Action != schema.TxActionSwap {
		acc, nonce, fee, err = h.TxVerify(tx, false)
		if err != nil {
			return err
		}
	} else {
		// verify swap tx's router
		if !h.verifyFromRouter(tx.Router) {
			return schema.ErrInvalidFromRouter
		}
		nonce, _ = strconv.ParseInt(tx.Nonce, 10, 64)
	}

	switch tx.Action {
	case schema.TxActionTransfer:
		to, amount, err := TxTransferParamsVerify(tx.Params)
		if err != nil {
			return err
		}
		err = h.Token.Transfer(tx.From, to, amount, tx.FeeRecipient, fee, false)
		if err != nil {
			return err
		}
	case schema.TxActionPropose:
		proposal, err := h.ProposalVerify(tx, nonce)
		if err != nil {
			return err
		}

		h.Proposals = append(h.Proposals, proposal)

	case schema.TxActionStake:
		stakePool, amount, err := TxStakeParamsVerify(tx.Params)
		if err != nil {
			return err
		}

		// in stakepools, and not in depreated stakepools
		if !InSlice(h.StakePools, stakePool) {
			return schema.ErrInvalidStakePool
		}
		if InSlice(h.OnlyUnStakePools, stakePool) {
			return schema.ErrInvalidStakePool
		}

		stakeAt := nonce / 1000
		err = h.Token.Stake(tx.From, stakePool, amount, stakeAt, tx.FeeRecipient, fee, false)
		if err != nil {
			return err
		}

	case schema.TxActionUnstake:
		stakePool, amount, err := TxUnstakeParamsVerify(tx.Params)
		if err != nil {
			return err
		}

		if !InSlice(h.StakePools, stakePool) {
			return schema.ErrInvalidStakePool
		}

		err = h.Token.Unstake(tx.From, stakePool, amount, tx.FeeRecipient, fee, false)
		if err != nil {
			return err
		}

		// check if router need to be removed after unstake
		if InSlice(h.Routers, tx.From) {
			staked := h.Token.TotalStaked(tx.From, "")
			staked_, _ := new(big.Int).SetString(staked, 10)
			routerMinStake, _ := new(big.Int).SetString(h.RouterMinStake, 10)
			if staked_.Cmp(routerMinStake) == -1 {
				log.Info("Remove router after unstake", "router", tx.From)
				h.Routers = RemoveFromSlice(h.Routers, tx.From)
			}
		}

	case schema.TxActionJoin:
		if InSlice(h.Routers, tx.From) {
			return schema.ErrRouterAlreadyJoined
		}
		routerState, err := TxJoinParamsVerify(tx.Params)
		if err != nil {
			return err
		}
		if routerState.Router != tx.From {
			return schema.ErrInvalidRouterAddress
		}
		for _, r := range h.RouterStates {
			if strings.EqualFold(r.Name, routerState.Name) {
				log.Error("router name already exists", "name", routerState.Name)
				return schema.ErrInvalidRouterName
			}
		}
		staked := h.Token.TotalStaked(tx.From, "")
		staked_, _ := new(big.Int).SetString(staked, 10)
		routerMinStake, _ := new(big.Int).SetString(h.RouterMinStake, 10)
		if staked_.Cmp(routerMinStake) == -1 {
			return schema.ErrInsufficientStake
		}

		h.Routers = append(h.Routers, tx.From)
		h.RouterStates[tx.From] = routerState

	case schema.TxActionLeave:
		if !InSlice(h.Routers, tx.From) {
			return schema.ErrNotARouter
		}
		log.Info("Remove router from routers", "router", tx.From)
		h.Routers = RemoveFromSlice(h.Routers, tx.From)
		delete(h.RouterStates, tx.From)

	case schema.TxActionCall:
		proposalID, _, _, err := TxCallParamsVerify(tx.Params)
		if err != nil {
			return err
		}
		ProposalCalled = FindProposal(h.Proposals, proposalID)
		if ProposalCalled == nil {
			return schema.ErrNoProposalFound
		}

	case schema.TxActionSwap:
		routerState, ok := h.RouterStates[tx.Router]
		if !ok {
			log.Error("router not found", "router", tx.Router)
			return err
		}
		pools := routerState.Pools
		if len(pools) == 0 {
			log.Error("router have no pools", "router", tx.Router)
			return err
		}

		routerFee := false
		if routerState.SwapFeeRecipient != "" && routerState.SwapFeeRatio != "0" {
			routerFee = true
		}
		order, err := TxSwapParamsVerify(tx.Params, tx.Nonce, routerFee, pools, oracle.EverTokens)
		if err != nil {
			log.Error("swap params verify failed", "err", err)
			return err
		}

		tx.SwapOrder = order

		//log.Info("swap tx", "everhash", tx.EverHash, "order error", order.Err, "order items", len(order.Items))

	default:
		return schema.ErrInvalidTxAction
	}

	// swap tx no need to update nonce
	if acc != nil && nonce > 0 && tx.Action != schema.TxActionSwap {
		acc.UpdateNonce(nonce)
	}
	h.LatestTxHash = tx.HexHash()
	h.LatestTxEverHash = tx.EverHash

	if ProposalCalled != nil {
		// run proposal executor
		log.Debug("proposal called", "ID", ProposalCalled.ID, "name", ProposalCalled.Name, "tx", tx.HexHash())

		ns, err := ProposalExecute(ProposalCalled, &tx, h.GetStateForProposal(), oracle)
		if err != nil {
			log.Error("execute proposal failed", "ID", ProposalCalled.ID, "name", ProposalCalled.Name, "tx", tx.HexHash(), "err", err)
			ProposalCalled.ExecutedTxs[tx.EverHash] = err.Error()
			return err
		}
		ProposalCalled.ExecutedTxs[tx.EverHash] = ""
		h.UpdateState(ns)
	} else {
		// run every proposal executor
		// todo check if proposal is unstarted
		for _, proposal := range h.Proposals {

			// skip proposal if not in onlyAcceptedTxActions
			if len(proposal.OnlyAcceptedTxActions) > 0 && !InSlice(proposal.OnlyAcceptedTxActions, tx.Action) {
				continue
			}

			log.Debug("execute proposal", "ID", proposal.ID, "name", proposal.Name, "tx", tx.HexHash())
			ns, err := ProposalExecute(proposal, &tx, h.GetStateForProposal(), oracle)
			if err != nil {
				log.Error("execute proposal failed", "ID", proposal.ID, "name", proposal.Name, "tx", tx.HexHash(), "err", err)
				proposal.ExecutedTxs[tx.EverHash] = err.Error()
				continue
			}
			proposal.ExecutedTxs[tx.EverHash] = ""
			h.UpdateState(ns)
		}
	}

	// remove expired/finished proposal
	proposals := []*schema.Proposal{}
	for _, proposal := range h.Proposals {
		// finished proposal
		log.Debug("try to remove proposal", "ID", proposal.ID, "name", proposal.Name, "runnedTimes", proposal.Executor.RunnedTimes, "runTimes", proposal.RunTimes)
		if proposal.RunTimes > 0 && proposal.Executor.RunnedTimes >= proposal.RunTimes {
			continue
		}
		// expired proposal
		if proposal.End > 0 && nonce/1000 >= proposal.End {
			continue
		}
		proposals = append(proposals, proposal)
	}
	h.Proposals = proposals

	h.StateHash = h.Hash()
	return nil
}
