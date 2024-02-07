package schema

import (
	"math/big"

	"github.com/permadao/permaswap/halo/account"
	"github.com/permadao/permaswap/halo/token"
	tokSchema "github.com/permadao/permaswap/halo/token/schema"
)

type RouterState struct {
	Router           string           `json:"router"` // router address
	Name             string           `json:"name"`
	Logo             string           `json:"logo"` // logo url
	HTTPEndpoint     string           `json:"httpEndpoint"`
	WSEndpoint       string           `json:"wsEndpoint"`
	SwapFeeRatio     string           `json:"swapFeeRatio"`
	SwapFeeRecipient string           `json:"swapFeeRecipient"`
	Pools            map[string]*Pool `json:"pools"`      // pool id -> pool
	LpMinStake       string           `json:"lpMinStake"` // minum amount lp stake
	LpPenalty        string           `json:"lpPenalty"`  // penalty for lp evil
	Info             string           `json:"info"`       // router info
}

type State struct {
	Dapp         string `json:"dapp"`
	ChainID      string `json:"chainID"`
	Govern       string `json:"govern"` // govern is an temporary solution and will use voting in the future
	FeeRecipient string `json:"feeRecipient"`

	RouterMinStake string                  `json:"routerMinStake"` // minum amount router stake
	Routers        []string                `json:"routers"`        // [] router address
	RouterStates   map[string]*RouterState `json:"routerState"`    // router address -> router state

	Token    *token.Token                `json:"-"`       // halo token
	Accounts map[string]*account.Account `json:"account"` // account ID -> account

	Proposals []*Proposal `json:"proposals"` // proposal hash -> proposal

	StakePools []string `json:"stakePools"` // stake pool
	//stakes pool only accept unstake
	OnlyUnStakePools []string `json:"onlyUnStakePools"`

	Executed []string        `json:"executed"` // executed tx everhash hash
	Validity map[string]bool `json:"validity"` // executed tx everhash hash -> bool

	LatestTxHash     string `json:"latestTxHash"`
	LatestTxEverHash string `json:"latestTxEverHash"`

	StateHash string `json:"stateHash"`
}

type StateForProposal struct {
	Dapp         string `json:"dapp"`
	ChainID      string `json:"chainID"`
	Govern       string `json:"govern"` // govern is an temporary solution and will use voting in the future
	FeeRecipient string `json:"feeRecipient"`

	RouterMinStake string                  `json:"minRouterStake"` // minum amount router stake
	Routers        []string                `json:"routers"`
	RouterStates   map[string]*RouterState `json:"routerState"` // router id -> router state

	StakePools       []string `json:"stakePools"` // stake pool
	OnlyUnStakePools []string `json:"onlyUnStakePools"`

	Token *token.Token `json:"token"` // halo token
}

func (s *State) Hash() string {
	// todo: implement
	return ""
}

func CopyRouterState(dst, src *RouterState) {
	dst.Router = src.Router
	dst.Name = src.Name
	dst.Logo = src.Logo
	dst.HTTPEndpoint = src.HTTPEndpoint
	dst.WSEndpoint = src.WSEndpoint
	dst.SwapFeeRatio = src.SwapFeeRatio
	dst.SwapFeeRecipient = src.SwapFeeRecipient
	dst.LpMinStake = src.LpMinStake
	dst.LpPenalty = src.LpPenalty
	dst.Info = src.Info

	dst.Pools = make(map[string]*Pool)
	for _, p := range src.Pools {
		np := &Pool{
			TokenXTag: p.TokenXTag,
			TokenYTag: p.TokenYTag,
			FeeRatio:  p.FeeRatio,
		}
		dst.Pools[np.ID()] = np
	}
}

func CopyToken(dst, src *token.Token) {
	dst.Symbol = src.Symbol
	dst.Decimals = src.Decimals
	dst.TotalSupply = src.TotalSupply

	dst.Balances = make(map[string]*big.Int)
	for a, b := range src.Balances {
		nb := new(big.Int)
		nb.Add(b, big.NewInt(0))
		dst.Balances[a] = nb
	}

	dst.Stakes = make(map[string]map[string][]tokSchema.Stake)
	for a, b := range src.Stakes {
		dst.Stakes[a] = make(map[string][]tokSchema.Stake)
		for c, d := range b {
			dst.Stakes[a][c] = make([]tokSchema.Stake, len(d))
			copy(dst.Stakes[a][c], d)
		}
	}
}

func (s *State) GetStateForProposal() *StateForProposal {
	routers := make([]string, len(s.Routers))
	copy(routers, s.Routers)
	routerStates := make(map[string]*RouterState)
	for _, rs := range s.RouterStates {
		nrs := &RouterState{}
		CopyRouterState(nrs, rs)
		routerStates[nrs.Router] = nrs
	}
	token := &token.Token{}
	CopyToken(token, s.Token)

	stakePools := make([]string, len(s.StakePools))
	copy(stakePools, s.StakePools)
	onlyUnStakePools := make([]string, len(s.OnlyUnStakePools))
	copy(onlyUnStakePools, s.OnlyUnStakePools)

	return &StateForProposal{
		Dapp:             s.Dapp,
		ChainID:          s.ChainID,
		Govern:           s.Govern,
		FeeRecipient:     s.FeeRecipient,
		Routers:          routers,
		RouterStates:     routerStates,
		RouterMinStake:   s.RouterMinStake,
		Token:            token,
		StakePools:       stakePools,
		OnlyUnStakePools: onlyUnStakePools,
	}
}

func (s *State) UpdateState(ns *StateForProposal) {
	s.Dapp = ns.Dapp
	s.ChainID = ns.ChainID
	s.Govern = ns.Govern
	s.FeeRecipient = ns.FeeRecipient
	s.Routers = ns.Routers
	s.RouterStates = ns.RouterStates
	s.RouterMinStake = ns.RouterMinStake
	s.Token = ns.Token
	s.StakePools = ns.StakePools
	s.OnlyUnStakePools = ns.OnlyUnStakePools
}
