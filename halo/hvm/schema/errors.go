package schema

import "errors"

var (
	ErrInvalidFromRouter    = errors.New("err_invalid_from_router")
	ErrInvalidNonce         = errors.New("err_invalid_nonce")
	ErrInvalidAmount        = errors.New("err_invalid_amount")
	ErrInvalidFee           = errors.New("err_invalid_fee")
	ErrInvalidFeeRecipient  = errors.New("err_invalid_fee_recipient")
	ErrInvalidTx            = errors.New("err_invalid_tx")
	ErrInvalidTxField       = errors.New("err_invalid_tx_field")
	ErrInvalidTxAction      = errors.New("err_invalid_tx_action")
	ErrInvalidTxParams      = errors.New("err_invalid_tx_params")
	ErrInvalidProposer      = errors.New("err_invalid_proposer")
	ErrInvalidProposal      = errors.New("err_invalid_proposal")
	ErrTxExecuted           = errors.New("err_tx_executed")
	ErrInvalidStakePool     = errors.New("err_invalid_stake_pool")
	ErrInvalidRouterAddress = errors.New("err_invalid_router_address")
	ErrInsufficientStake    = errors.New("err_insufficient_stake")
	ErrRouterAlreadyJoined  = errors.New("err_router_already_joined")
	ErrNotARouter           = errors.New("err_not_a_router")
	ErrNoProposalFound      = errors.New("err_no_proposal_found")
)
