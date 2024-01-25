package token

import (
	"math/big"

	"github.com/permadao/permaswap/halo/token/schema"
)

func (t *Token) balanceOf(addr string) *big.Int {
	if bal, ok := t.Balances[addr]; ok {
		return bal
	}

	return big.NewInt(0)
}

func (t *Token) add(addr string, amount *big.Int, dryRun bool) error {
	if amount == nil {
		return schema.ErrNilAmount
	}

	// amount must >= 0
	if amount.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeAmount
	}

	// if amount == 0, then return nil
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	if !dryRun {
		bal := t.balanceOf(addr)
		t.Balances[addr] = new(big.Int).Add(bal, amount)
	}

	return nil
}

func (t *Token) sub(addr string, amount *big.Int, dryRun bool) error {
	if amount == nil {
		return schema.ErrNilAmount
	}

	// amount must >= 0
	if amount.Cmp(big.NewInt(0)) == -1 {
		return schema.ErrNegativeAmount
	}

	// if amount == 0, then return nil
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	bal := t.balanceOf(addr)
	if bal.Cmp(amount) < 0 {
		return schema.ErrInsufficientBalance
	}

	if !dryRun {
		if bal.Cmp(amount) == 0 {
			delete(t.Balances, addr)
			return nil
		}

		t.Balances[addr] = new(big.Int).Sub(bal, amount)
	}

	return nil
}
