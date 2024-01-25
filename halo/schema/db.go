package schema

import (
	"time"

	hvmSchema "github.com/permadao/permaswap/halo/hvm/schema"
)

type HaloTransaction struct {
	ID        int64      `gorm:"primary_key;auto_increment" json:"-"`
	UpdatedAt *time.Time `gorm:"ASSOCIATION_AUTOUPDATE" json:"-"`
	CreatedAt *time.Time `gorm:"ASSOCIATION_AUTOCREATE" json:"createdAt"`
	EverHash  string     `gorm:"type:varchar(66);uniqueIndex" json:"everHash"`
	HaloHash  string     `gorm:"type:varchar(66);uniqueIndex" json:"haloHash"`
	hvmSchema.Transaction
	Error string `json:"error"`
}
