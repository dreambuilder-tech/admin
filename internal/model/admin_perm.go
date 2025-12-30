package model

import (
	"context"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AdminPerm struct {
	ID        int64          `gorm:"column:id;primaryKey" json:"id"`
	UID       int64          `gorm:"column:uid;uniqueIndex" json:"uid"`
	Perms     datatypes.JSON `gorm:"column:perms;type:json" json:"perms"`
	CreatedAt int64          `gorm:"column:created_at" json:"created_at"`
	UpdatedAt int64          `gorm:"column:updated_at" json:"updated_at"`
}

func (*AdminPerm) TableName() string {
	return "admin_perms"
}

func (up *AdminPerm) GetByUID(ctx context.Context, db *gorm.DB, uid int64) error {
	return db.WithContext(ctx).Where("uid = ?", uid).Take(up).Error
}

func (up *AdminPerm) CreateOrUpdate(ctx context.Context, db *gorm.DB, uid int64, perms datatypes.JSON) error {
	now := time.Now().Unix()
	result := db.WithContext(ctx).Model(&AdminPerm{}).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"perms":      perms,
			"updated_at": now,
		})

	if result.RowsAffected == 0 {
		return db.WithContext(ctx).Create(&AdminPerm{
			UID:       uid,
			Perms:     perms,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
	}

	return result.Error
}

func (up *AdminPerm) DeleteByUID(ctx context.Context, db *gorm.DB, uid int64) error {
	return db.WithContext(ctx).Where("uid = ?", uid).Delete(up).Error
}
