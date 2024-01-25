package lp

import "errors"

var (
	ERR_NO_ENOUGH_BALANCE = errors.New("err_no_enough_balance")
	ERR_INVALID_AMOUNT    = errors.New("err_invalid_amount")
)
