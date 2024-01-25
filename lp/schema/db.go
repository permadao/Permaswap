package schema

import "time"

type Order struct {
	ID          int64      `gorm:"primary_key;auto_increment" json:"id"`
	UpdatedAt   *time.Time `gorm:"ASSOCIATION_AUTOUPDATE" json:"-"`
	CreatedAt   *time.Time `gorm:"ASSOCIATION_AUTOCREATE" json:"-"`
	UserAddr    string     `gorm:"index:poindex1" json:"address"`
	EverHash    string     `gorm:"index:poindex2" json:"everHash"`
	OrderStatus string     `json:"status"`
	LpMsgOrder  string     `json:"lpMsgOrde"`
}
