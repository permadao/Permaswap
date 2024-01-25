package core

import "errors"

var (
	ERR_INVALID_POOL       = errors.New("err_invalid_pool")
	ERR_NO_POOL            = errors.New("err_no_pool")
	ERR_NO_LP              = errors.New("err_no_lp")
	ERR_NO_PATH            = errors.New("err_no_path")
	ERR_INVALID_AMOUNT     = errors.New("err_invalid_amount")
	ERR_OUT_OF_RANGE       = errors.New("err_out_of_range")
	ERR_INVALID_TOKEN      = errors.New("err_invalid_token")
	ERR_INVALID_NUMBER     = errors.New("err_invalid_number")
	ERR_NO_IMPLEMENT       = errors.New("err_no_implement")
	ERR_INVALID_PATH       = errors.New("err_invalid_path")
	ERR_INVALID_SWAPOUTS   = errors.New("err_invalid_swapouts")
	ERR_INVALID_POOL_PATHS = errors.New("err_invalid_pool_paths")
	ERR_INVALID_SWAP_USER  = errors.New("err_invalid_swap_user")
	ERR_INVALID_PATH_FEE   = errors.New("err_invalid_path_fee")
	ERR_FEE_TOO_SMALL      = errors.New("err_fee_too_small")
	// error for lp
	ERR_INVALID_PRICE           = errors.New("err_invalid_price")
	ERR_INVALID_LIQUIDITY       = errors.New("err_invalid_liquidity")
	ERR_INVALID_PRICE_DIRECTION = errors.New("err_invalid_price_direction")
	ERR_INVALID_FEE             = errors.New("err_invalid_fee")
)
