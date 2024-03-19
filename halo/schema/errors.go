package schema

import "errors"

var (
	ErrInvalidGenesisTx          = errors.New("err_invalid_genesis_tx")
	ErrInvalidGenesisTotalSupply = errors.New("err_invalid_genesis_total_supply")
	ErrInvalidGenesisBalance     = errors.New("err_invalid_genesis_balance")
	ErrInvalidGenesisStake       = errors.New("err_invalid_genesis_stake")
	ErrInvalidSubmitTxNonce      = errors.New("err_invalid_submit_tx_nonce")
	ErrMissParams                = errors.New("err_miss_params")
)
