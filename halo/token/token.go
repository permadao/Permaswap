package token

import (
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/permadao/permaswap/halo/logger"
	"github.com/permadao/permaswap/halo/token/schema"
)

var log = logger.New("token")

type Token struct {
	Symbol   string `json:"symbol"`
	Decimals int64  `json:"decimals"`

	TotalSupply *big.Int                             `json:"totalSupply"`
	Balances    map[string]*big.Int                  `json:"balances"` // account id -> balance
	Stakes      map[string]map[string][]schema.Stake `json:"stakes"`   // account id -> stake pool -> stakes
	lock        sync.RWMutex
}

func New(symbol string, decimals int64, totalSupply *big.Int, balances map[string]*big.Int, stakes map[string]map[string][]schema.Stake) *Token {
	if balances == nil {
		balances = map[string]*big.Int{}
	}
	if stakes == nil {
		stakes = map[string]map[string][]schema.Stake{}
	}
	return &Token{
		Symbol:      symbol,
		TotalSupply: totalSupply,
		Decimals:    decimals,
		Balances:    balances,
		Stakes:      stakes,
	}
}

func (t *Token) Transfer(from, to string, amount *big.Int, feeRecipient string, fee *big.Int, dryRun bool) (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if amount == nil {
		return schema.ErrNilAmount
	}

	if fee == nil {
		fee = big.NewInt(0)
	}

	if amount.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeAmount
	}

	if fee.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeFee
	}

	sum := new(big.Int).Add(amount, fee)
	if sum.Cmp(abi.MaxUint256) > 0 {
		return schema.ErrTooLargeAmount
	}

	if err = t.sub(from, new(big.Int).Add(amount, fee), dryRun); err != nil {
		return
	}
	if err = t.add(to, amount, dryRun); err != nil {
		return
	}
	if err = t.add(feeRecipient, fee, dryRun); err != nil {
		return
	}

	return nil
}

func (t *Token) Stake(from, stakePool string, amount *big.Int, stakedAt int64, feeRecipient string, fee *big.Int, dryRun bool) (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if amount == nil {
		return schema.ErrNilAmount
	}

	if fee == nil {
		fee = big.NewInt(0)
	}

	// no stake 0 amount
	if amount.Cmp(big.NewInt(0)) == 0 {
		return schema.ErrZeroAmount
	}

	if amount.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeAmount
	}

	if fee.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeFee
	}

	sum := new(big.Int).Add(amount, fee)
	if sum.Cmp(abi.MaxUint256) > 0 {
		return schema.ErrTooLargeAmount
	}

	if err = t.sub(from, new(big.Int).Add(amount, fee), dryRun); err != nil {
		return err
	}
	if err = t.add(feeRecipient, fee, dryRun); err != nil {
		return err
	}

	if !dryRun {
		stake := schema.Stake{
			StakedAt: stakedAt,
			Amount:   amount,
		}
		if _, ok := t.Stakes[from]; !ok {
			t.Stakes[from] = map[string][]schema.Stake{}
			t.Stakes[from][stakePool] = []schema.Stake{stake}
			return
		}
		if _, ok := t.Stakes[from][stakePool]; !ok {
			t.Stakes[from][stakePool] = []schema.Stake{stake}
			return
		}
		t.Stakes[from][stakePool] = append(t.Stakes[from][stakePool], stake)
	}
	return
}

func (t *Token) Unstake(from, stakePool string, amount *big.Int, feeRecipient string, fee *big.Int, dryRun bool) (err error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	if amount == nil {
		return schema.ErrNilAmount
	}

	if fee == nil {
		fee = big.NewInt(0)
	}

	// no unstake 0 amount
	if amount.Cmp(big.NewInt(0)) == 0 {
		return schema.ErrZeroAmount
	}

	if amount.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeAmount
	}

	if fee.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeFee
	}

	sum := new(big.Int).Add(amount, fee)
	if sum.Cmp(abi.MaxUint256) > 0 {
		return schema.ErrTooLargeAmount
	}

	stakes, ok := t.Stakes[from]
	if !ok {
		return schema.ErrInsufficientStake
	}
	stakesByPool, ok := stakes[stakePool]
	if !ok {
		return schema.ErrInsufficientStake
	}

	totalStaked := big.NewInt(0)
	for _, stake := range stakesByPool {
		totalStaked = new(big.Int).Add(totalStaked, stake.Amount)
	}
	if totalStaked.Cmp(amount) == -1 {
		return schema.ErrInsufficientStake
	}

	if err = t.add(feeRecipient, fee, dryRun); err != nil {
		return err
	}

	if err = t.add(from, amount, dryRun); err != nil {
		return err
	}

	if !dryRun {
		for i := len(stakesByPool) - 1; i >= 0; i-- {
			stake := stakesByPool[i]
			if stake.Amount.Cmp(amount) == -1 {
				amount = new(big.Int).Sub(amount, stake.Amount)
				stakesByPool = stakesByPool[:i]
			} else if stake.Amount.Cmp(amount) == 0 {
				stakesByPool = stakesByPool[:i]
				break
			} else {
				stakesByPool[i].Amount = new(big.Int).Sub(stake.Amount, amount)
				break
			}
		}
		t.Stakes[from][stakePool] = stakesByPool
	}

	return
}

func (t *Token) BalanceOf(addr string) string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.balanceOf(addr).String()
}

func (t *Token) TotalStaked(addr, stakePool string) string {
	t.lock.RLock()
	defer t.lock.RUnlock()

	staked := big.NewInt(0)
	if stakePool != "" {
		for _, stake := range t.Stakes[addr][stakePool] {
			staked = new(big.Int).Add(staked, stake.Amount)
		}
		return staked.String()
	}

	for _, stakePools := range t.Stakes[addr] {
		for _, stake := range stakePools {
			staked = new(big.Int).Add(staked, stake.Amount)
		}
	}
	return staked.String()

}

// to avoid big.int json marshal loss accuracy
func (t *Token) Info() *schema.TokenInfo {
	t.lock.RLock()
	defer t.lock.RUnlock()

	balances := map[string]string{}
	for addr, balance := range t.Balances {
		balances[addr] = balance.String()
	}

	stakes := map[string]map[string][]schema.StakeInfo{}
	for addr, stakePools := range t.Stakes {
		stakes[addr] = map[string][]schema.StakeInfo{}
		for stakePool, stakes_ := range stakePools {
			stakesInfos := []schema.StakeInfo{}
			for _, stake := range stakes_ {
				stakesInfos = append(stakesInfos, schema.StakeInfo{
					StakedAt: stake.StakedAt,
					Amount:   stake.Amount.String(),
				})
			}
			stakes[addr][stakePool] = stakesInfos
		}
	}

	return &schema.TokenInfo{
		Symbol:      t.Symbol,
		Decimals:    t.Decimals,
		TotalSupply: t.TotalSupply.String(),
		Balances:    balances,
		Stakes:      stakes,
	}
}

func (t *Token) TransferToStake(from, to string, amount *big.Int, stakePool string, stakedAt int64,
	feeRecipient string, fee *big.Int, dryRun bool) (err error) {

	t.lock.Lock()
	defer t.lock.Unlock()

	if amount == nil {
		return schema.ErrNilAmount
	}

	if fee == nil {
		fee = big.NewInt(0)
	}

	// no stake 0 amount
	if amount.Cmp(big.NewInt(0)) == 0 {
		return schema.ErrZeroAmount
	}

	if amount.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeAmount
	}

	if fee.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeFee
	}

	sum := new(big.Int).Add(amount, fee)
	if sum.Cmp(abi.MaxUint256) > 0 {
		return schema.ErrTooLargeAmount
	}

	if err = t.sub(from, new(big.Int).Add(amount, fee), dryRun); err != nil {
		return
	}
	if err = t.add(feeRecipient, fee, dryRun); err != nil {
		return
	}

	if !dryRun {
		stake := schema.Stake{
			StakedAt: stakedAt,
			Amount:   amount,
		}
		if _, ok := t.Stakes[to]; !ok {
			t.Stakes[to] = map[string][]schema.Stake{}
			t.Stakes[to][stakePool] = []schema.Stake{stake}
			return
		}
		if _, ok := t.Stakes[to][stakePool]; !ok {
			t.Stakes[to][stakePool] = []schema.Stake{stake}
			return
		}
		t.Stakes[to][stakePool] = append(t.Stakes[to][stakePool], stake)
	}

	return nil
}
