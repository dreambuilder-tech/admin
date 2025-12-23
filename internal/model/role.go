package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID        int       `gorm:"column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (*Role) TableName() string {
	return "roles"
}

// GetAll 获取所有角色列表
func (r *Role) GetAll(ctx context.Context, db *gorm.DB) ([]*Role, error) {
	var list []*Role
	err := db.WithContext(ctx).Table(r.TableName()).Order("`id` ASC").Find(&list).Error
	return list, err
}

// GetByID 根据ID获取角色
func (r *Role) GetByID(ctx context.Context, db *gorm.DB, roleID int) error {
	return db.WithContext(ctx).Where("`id` = ?", roleID).Take(r).Error
}

// GetByName 根据名称获取角色
func (r *Role) GetByName(ctx context.Context, db *gorm.DB, name string) error {
	return db.WithContext(ctx).Where("`name` = ?", name).Take(r).Error
}

// Exists 检查角色是否存在
func (r *Role) Exists(ctx context.Context, db *gorm.DB, roleID int) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Table(r.TableName()).Where("`id` = ?", roleID).Count(&count).Error
	return count > 0, err
}
