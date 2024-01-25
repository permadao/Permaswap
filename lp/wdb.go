package lp

import (
	"path"

	"github.com/permadao/permaswap/lp/schema"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	sqliteName = "lp.sqlite"
)

type WDB struct {
	db *gorm.DB
}

func NewWDB(dbDir string) *WDB {
	db, err := gorm.Open(sqlite.Open(path.Join(dbDir, sqliteName)), &gorm.Config{
		Logger:          logger.Default.LogMode(logger.Silent),
		CreateBatchSize: 200,
	})
	if err != nil {
		panic(err)
	}
	log.Info("connect sqlite db success")
	return &WDB{db: db}
}

func (w *WDB) Migrate() {
	w.db.AutoMigrate(&schema.Order{})
}

func (w *WDB) CreateOrder(order *schema.Order, tx *gorm.DB) error {
	if tx == nil {
		tx = w.db
	}
	return tx.Create(&order).Error
}

func (w *WDB) GetOrders(page, count int) (orders []*schema.Order, err error) {
	dbPage := page - 1
	if dbPage < 0 {
		dbPage = 0
	}
	err = w.db.Model(&schema.Order{}).Order("id desc").Offset(dbPage * count).Limit(count).Find(&orders).Error
	return
}
