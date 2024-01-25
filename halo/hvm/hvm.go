package hvm

import (
	"github.com/permadao/permaswap/halo/account"
	"github.com/permadao/permaswap/halo/hvm/schema"
	"github.com/permadao/permaswap/halo/logger"
)

var log = logger.New("hvm")

type HVM struct {
	schema.State
}

func New(initState schema.State) (h *HVM) {
	if initState.Routers == nil {
		initState.Routers = []string{}
	}
	if initState.RouterStates == nil {
		initState.RouterStates = map[string]*schema.RouterState{}
	}
	if initState.Accounts == nil {
		initState.Accounts = map[string]*account.Account{}
	}
	if initState.Proposals == nil {
		initState.Proposals = []*schema.Proposal{}
	}
	if initState.Executed == nil {
		initState.Executed = []string{}
	}
	if initState.Validity == nil {
		initState.Validity = map[string]bool{}
	}
	if initState.StakePools == nil {
		initState.StakePools = []string{}
	}
	if initState.OnlyUnStakePools == nil {
		initState.OnlyUnStakePools = []string{}
	}
	if initState.RouterMinStake == "" {
		initState.RouterMinStake = "0"
	}
	return &HVM{
		State: initState,
	}
}

func (h *HVM) getOrCreateAccount(addr string, dryRun bool) (acc *account.Account, err error) {
	acc = h.Accounts[addr]
	if acc == nil {
		if acc, err = account.New(addr); err != nil {
			return
		}
		if !dryRun {
			h.Accounts[acc.ID] = acc
		}
	}
	return
}
