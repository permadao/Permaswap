package schema

type Volume struct {
	//Duration  int64 `json:"duration"` // in minutes
	Timestamp int64   `json:"timestamp"`
	PoolID    string  `json:"poolID"`
	AccID     string  `json:"accID"`
	LpID      string  `json:"lpID"`
	TokenX    string  `json:"tokenX"`
	TokenY    string  `json:"tokenY"`
	X         float64 `json:"volumeX"`
	Y         float64 `json:"volumeY"`
	USD       float64 `json:"volumeInUSD"`
	RewardX   float64 `json:"rewardX"`
	RewardY   float64 `json:"rewardY"`
	RewardUSD float64 `json:"rewardInUSD"`
	SwapCount int64   `json:"swapCount"`
}

type TVL struct {
	Timestamp int64   `json:"timestamp"`
	PoolID    string  `json:"poolID"`
	LpID      string  `json:"lpID"`
	AccID     string  `json:"accID"`
	X         float64 `json:"tokenXTVL"`
	Y         float64 `json:"tokenYTVL"`
	USD       float64 `json:"tvlInUSD"`
}
