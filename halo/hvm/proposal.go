package hvm

import (
	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/hvm/symbol"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func NewExecutor(source string) (executor *schema.Executor, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("NewExecutor failed", "err", r)
			err = schema.ErrInvalidProposal
		}
	}()

	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	i.Use(symbol.Symbols)

	if _, err := i.Eval(source); err != nil {
		return nil, err
	}

	v, err := i.Eval("proposal.Execute")
	if err != nil {
		return nil, err
	}
	execute := v.Interface().(func(*schema.Transaction, *schema.StateForProposal, *schema.Oracle, string, string) (*schema.StateForProposal, string, string, error))
	return &schema.Executor{Execute: execute, RunnedTimes: 0, LocalState: ""}, nil
}

func NewProposal(name string, start, end, runTimes int64, source, initData string, onlyAcceptedTxActions []string, executor *schema.Executor) *schema.Proposal {
	proposal := &schema.Proposal{
		Name:                  name,
		Start:                 start,
		End:                   end,
		RunTimes:              runTimes,
		Source:                source,
		InitData:              initData,
		OnlyAcceptedTxActions: onlyAcceptedTxActions,
		Executor:              executor,
	}
	proposal.ID = proposal.HexHash()
	return proposal
}

func ProposalExecute(proposal *schema.Proposal, tx *schema.Transaction, state *schema.StateForProposal, oracle *schema.Oracle) (*schema.StateForProposal, error) {

	txCopied := &schema.Transaction{}
	if err := DeepCopyTx(tx, txCopied); err != nil {
		log.Error("deep copy tx failed", "err", err)
		return state, err
	}

	proposal.Executor.RunnedTimes++

	// todo: if panic need recover
	stateNew, localStateNew, localStateHashNew, err := proposal.Executor.Execute(txCopied, state, oracle, proposal.Executor.LocalState, proposal.InitData)
	if err != nil {
		return state, err
	}
	proposal.Executor.LocalState = localStateNew
	proposal.Executor.LocalStateHash = localStateHashNew
	return stateNew, nil
}

func FindProposal(proposals []*schema.Proposal, proposalID string) *schema.Proposal {
	for _, proposal := range proposals {
		if proposal.ID == proposalID {
			return proposal
		}
	}
	return nil
}
