package schema

import "errors"

var (
	ErrNilAmount           = errors.New("err_nil_amount")
	ErrNegativeAmount      = errors.New("err_negative_amount")
	ErrZeroAmount          = errors.New("err_zero_amount")
	ErrNegativeFee         = errors.New("err_negative_fee")
	ErrTooLargeAmount      = errors.New("err_too_large_amount")
	ErrInsufficientBalance = errors.New("err_insufficient_balance")
	ErrInsufficientStake   = errors.New("err_insufficient_stake")
)
