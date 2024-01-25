package router

import (
	"time"

	"github.com/permadao/permaswap/router/schema"
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
	w.db.AutoMigrate(&schema.PermaOrder{})
	w.db.AutoMigrate(&schema.PermaVolume{})
	w.db.AutoMigrate(&schema.PermaLpReward{})
	w.db.AutoMigrate(&schema.PermaLpsSnapshot{})
	w.db.AutoMigrate(&schema.NFTWhiteList{})
}

func (w *WDB) CreatePermaOrder(order *schema.PermaOrder, tx *gorm.DB) error {
	if tx == nil {
		tx = w.db
	}
	return tx.Create(&order).Error
}

func (w *WDB) CreatePermaVolume(volume *schema.PermaVolume, tx *gorm.DB) error {
	if tx == nil {
		tx = w.db
	}
	return tx.Create(&volume).Error
}

func (w *WDB) TotalOrdersNumByUser(accid string) (num int64, err error) {
	err = w.db.Model(&schema.PermaOrder{}).Where("user_addr = ?", accid).Count(&num).Error
	return
}

func (w *WDB) GetOrders(page, count int, start, end time.Time) (orders []*schema.PermaOrder, err error) {
	dbPage := page - 1
	if dbPage < 0 {
		dbPage = 0
	}
	err = w.db.Model(&schema.PermaOrder{}).Order("id desc").Where("created_at BETWEEN ? AND ?", start, end).Offset(dbPage * count).Limit(count).Find(&orders).Error
	return
}

func (w *WDB) GetOrdersByUser(accid string, page, count int, start, end time.Time) (orders []*schema.PermaOrder, err error) {
	dbPage := page - 1
	if dbPage < 0 {
		dbPage = 0
	}
	err = w.db.Model(&schema.PermaOrder{}).Where("user_addr = ?", accid).Order("id desc").Where("created_at BETWEEN ? AND ?", start, end).Offset(dbPage * count).Limit(count).Find(&orders).Error
	return
}

func (w *WDB) GetVolumesByTime(start, end time.Time) (volumes []*schema.PermaVolume, err error) {
	err = w.db.Model(&schema.PermaVolume{}).Where("created_at BETWEEN ? AND ?", start, end).Find(&volumes).Error
	return
}

func (w *WDB) SumVolumesByTime(start, end time.Time) (res []*schema.SumPermaVolumeRes, err error) {
	err = w.db.Model(&schema.PermaVolume{}).Select("pool_id, lp_id, acc_id, sum(amount_x) as amount_x, sum(amount_y) as amount_y, sum(reward_x) as reward_x, sum(reward_y) as reward_y").
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("pool_id").Group("lp_id").Group("acc_id").
		Scan(&res).Error
	return
}

func (w *WDB) SumPoolVolumesByTime(start, end time.Time) (res []*schema.SumPermaVolumeRes, err error) {
	err = w.db.Model(&schema.PermaVolume{}).Select("pool_id, sum(amount_x) as amount_x, sum(amount_y) as amount_y, sum(reward_x) as reward_x, sum(reward_y) as reward_y").
		Where("created_at BETWEEN ? AND ?", start, end).
		Group("pool_id").Scan(&res).Error
	return
}

func (w *WDB) SavePermaLpsSnapshot(lpsSnapshot *schema.PermaLpsSnapshot, tx *gorm.DB) (err error) {
	if tx == nil {
		tx = w.db
	}

	ls := &schema.PermaLpsSnapshot{}
	err = tx.First(ls).Error
	if err == nil {
		return tx.First(ls).Update("lps", lpsSnapshot.Lps).Error
	} else if err == gorm.ErrRecordNotFound {
		return tx.Create(&lpsSnapshot).Error
	} else {
		return
	}
}

func (w *WDB) UpdatePermaLpReward(lpReward *schema.PermaLpReward, tx *gorm.DB) (err error) {
	if tx == nil {
		tx = w.db
	}

	r := &schema.PermaLpReward{}
	err = tx.Where("lp_id = ?", lpReward.LpID).First(r).Error
	if err == nil {
		if lpReward.RewardX > 0 {
			return tx.Model(&schema.PermaLpReward{}).Where("lp_id = ?", lpReward.LpID).Update("reward_x", gorm.Expr("reward_x + ?", lpReward.RewardX)).Error
		} else {
			return tx.Model(&schema.PermaLpReward{}).Where("lp_id = ?", lpReward.LpID).Update("reward_y", gorm.Expr("reward_y + ?", lpReward.RewardY)).Error
		}
	} else if err == gorm.ErrRecordNotFound {
		return tx.Create(&lpReward).Error
	} else {
		return
	}
}

func (w *WDB) GetPermaRewards(accid string, tx *gorm.DB) (rewards []*schema.PermaLpReward, err error) {
	if tx == nil {
		tx = w.db
	}
	err = tx.Where("acc_id = ?", accid).Find(&rewards).Error
	return
}

func (w *WDB) GetPermaReward(lpid string, tx *gorm.DB) (reward *schema.PermaLpReward, err error) {
	if tx == nil {
		tx = w.db
	}
	err = tx.Where("lp_id = ?", lpid).First(&reward).Error
	return
}

func (w *WDB) LoadPermaLpsSnapshot() (lpsSnapshot *schema.PermaLpsSnapshot, err error) {
	err = w.db.First(&lpsSnapshot).Error
	return
}

func (w *WDB) LoadNFTWhiteList() (list []*schema.NFTWhiteList, err error) {
	err = w.db.Find(&list).Error
	return
}
