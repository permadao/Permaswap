package schema

import "time"

type PermaOrder struct {
	ID             int64      `gorm:"primary_key;auto_increment" json:"id"`
	UpdatedAt      *time.Time `gorm:"ASSOCIATION_AUTOUPDATE" json:"-"`
	CreatedAt      *time.Time `gorm:"ASSOCIATION_AUTOCREATE" json:"-"`
	UserAddr       string     `gorm:"index:poindex1" json:"address"`
	EverHash       string     `gorm:"index:poindex2" json:"everHash"`
	TokenInTag     string     `json:"tokenInTag"`
	TokenOutTag    string     `json:"tokenOutTag"`
	TokenInAmount  string     `json:"tokenInAmount"`
	TokenOutAmount string     `json:"tokenOutAmount"`
	Price          string     `json:"price"`
	OrderStatus    string     `json:"status"`
	OrderTimestamp int64      `json:"timestamp"` // nonce
}

type PermaVolume struct {
	ID              int64      `gorm:"primary_key;auto_increment" json:"id"`
	CreatedAt       *time.Time `gorm:"ASSOCIATION_AUTOCREATE" json:"-"`
	OrderID         int64      `json:"orderId"`
	EverHash        string     `json:"everHash"`
	PoolID          string     `json:"poolID"`
	AccID           string     `json:"accID"`
	LpID            string     `json:"lpID"`
	TokenXIsTokenIN bool       `json:"tokenXIsTokenIn"`
	AmountX         float64    `json:"amountX"`
	AmountY         float64    `json:"amountY"`
	RewardX         float64    `json:"rewardX"`
	RewardY         float64    `json:"rewardY"`
}

type PermaLpsSnapshot struct {
	ID        int64      `gorm:"primary_key;auto_increment" json:"id"`
	CreatedAt *time.Time `gorm:"ASSOCIATION_AUTOCREATE" json:"-"`
	UpdatedAt *time.Time `gorm:"ASSOCIATION_AUTOUPDATE" json:"-"`
	Lps       string     `gorm:"type:longtext"` //json text of all lps snapshot
}

type PermaLpReward struct {
	ID        int64      `gorm:"primary_key;auto_increment" json:"id"`
	CreatedAt *time.Time `gorm:"ASSOCIATION_AUTOCREATE" json:"-"`
	LpID      string     `gorm:"index:plrindex1,unique" json:"lpID"`
	PoolID    string     `json:"poolID"`
	AccID     string     `json:"accID"`
	RewardX   float64    `json:"rewardX"`
	RewardY   float64    `json:"rewardY"`
}

type SumPermaVolumeRes struct {
	PoolID    string  `json:"poolID"`
	AccID     string  `json:"accID"`
	LpID      string  `json:"lpID"`
	AmountX   float64 `json:"amountX"`
	AmountY   float64 `json:"amountY"`
	RewardX   float64 `json:"rewardX"`
	RewardY   float64 `json:"rewardY"`
	SwapCount int64   `json:"swapCount"`
}

type NFTWhiteList struct {
	ID        int64      `gorm:"primary_key;auto_increment"`
	UpdatedAt *time.Time `gorm:"ASSOCIATION_AUTOUPDATE"`
	CreatedAt *time.Time `gorm:"ASSOCIATION_AUTOCREATE"`
	UserAddr  string
	Remark    string
}
