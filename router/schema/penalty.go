package schema

const (
	LpPenaltyForNoSign            = "lp_no_sign"
	LpPenaltyForNoEnoughBalance   = "lp_no_enough_balance"
	UserPenaltyForNoEnoughBalance = "user_no_enough_balance"
)

type FailureRecord struct {
	Accid     string `json:"accid"`
	Timestamp int64  `json:"timestamp"`
	EverHash  string `json:"everHash"`
	Reason    string `json:"reason"`
}
