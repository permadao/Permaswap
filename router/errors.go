package router

import "encoding/json"

type WsErr struct {
	Event string `json:"event"`
	Msg   string `json:"msg"`
}

func NewWsErr(msg string) WsErr {
	return WsErr{Event: "error", Msg: msg}
}

func (w WsErr) Error() string {
	by, _ := json.Marshal(w)
	return string(by)
}

var (
	WsErrNotFoundPath          = NewWsErr("err_not_found_path")
	WsErrNotFoundSalt          = NewWsErr("err_not_found_salt")
	WsErrNotFoundLp            = NewWsErr("err_not_found_lp")
	WsErrNoAuthorization       = NewWsErr("err_no_authorization")
	WsErrCanNotUpdateLp        = NewWsErr("err_can_not_update_lp")
	WsErrDuplicateRegistration = NewWsErr("err_duplicate_registration")
	WsErrInvalidMsg            = NewWsErr("err_invalid_msg")
	WsErrInvalidToken          = NewWsErr("err_invalid_token")
	WsErrInvalidOrder          = NewWsErr("err_invalid_order")
	WsErrInvalidAddress        = NewWsErr("err_invalid_address")
	WsErrInvalidSignature      = NewWsErr("err_invalid_signature")
	WsErrInvalidPathsOrBundle  = NewWsErr("err_invalid_paths_or_bundle")
	WsErrNotNFTOwner           = NewWsErr("err_not_nft_owner")
	WsErrInvalidNFTData        = NewWsErr("err_invalid_nft_data")
	WsErrInvalidLpClient       = NewWsErr("err_invalid_lp_client")
	WsErrBlackListed           = NewWsErr("err_blacklisted")
)
