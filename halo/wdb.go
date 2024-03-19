package halo

import (
	"github.com/permadao/permaswap/halo/schema"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type WDB struct {
	db *gorm.DB
}

func NewWDB(dsn string) *WDB {
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}

	return &WDB{db}
}

func (w *WDB) Migrate() {
	w.db.AutoMigrate(&schema.HaloTransaction{})
}

func (w *WDB) CreateHaloTx(haloTx *schema.HaloTransaction, tx *gorm.DB) error {
	if tx == nil {
		tx = w.db
	}
	return tx.Create(&haloTx).Error
}

func (w *WDB) GetTx(hash string) (haloTx *schema.HaloTransaction, err error) {
	err = w.db.Where("ever_hash = ?", hash).Or("halo_hash = ?", hash).First(&haloTx).Error
	return
}
