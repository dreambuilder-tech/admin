package model

import (
	"context"

	"gorm.io/gorm"
)

type SystemConf struct {
	ID   int    `gorm:"column:id" json:"id"`
	Key  string `gorm:"column:key" json:"key"`
	Val  string `gorm:"column:val" json:"val"`
	Desc string `gorm:"column:desc" json:"desc"`
	Show int    `gorm:"column:show" json:"show"`
	Edit int    `gorm:"column:edit" json:"edit"`
}

func (*SystemConf) TableName() string {
	return "system_conf"
}

func (s *SystemConf) GetByKey(ctx context.Context, db *gorm.DB, key string) error {
	return db.WithContext(ctx).Where("`key` = ?", key).Take(s).Error
}
