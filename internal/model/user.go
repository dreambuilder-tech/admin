package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          int64      `gorm:"column:id" json:"id"`
	Account     string     `gorm:"column:account" json:"account"`
	Password    string     `gorm:"column:password" json:"-"`
	RoleID      int        `gorm:"column:role_id" json:"role_id"`
	Status      int        `gorm:"column:status" json:"status"`
	LastLoginAt *time.Time `gorm:"column:last_login_at" json:"last_login_at"`
	LastLoginIP string     `gorm:"column:last_login_ip" json:"last_login_ip"`
	MfaSecret   []byte     `gorm:"column:mfa_secret" json:"-"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

func (*User) TableName() string {
	return "users"
}

// Create 创建用户
func (u *User) Create(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Create(u).Error
}

// GetByID 根据ID获取用户
func (u *User) GetByID(ctx context.Context, db *gorm.DB, userID int64) error {
	return db.WithContext(ctx).Where("`id` = ?", userID).Take(u).Error
}

// GetByAccount 根据账号获取用户
func (u *User) GetByAccount(ctx context.Context, db *gorm.DB, account string) error {
	return db.WithContext(ctx).Where("`account` = ?", account).Take(u).Error
}

// UpdateLoginInfo 更新登录信息
func (u *User) UpdateLoginInfo(ctx context.Context, db *gorm.DB, userID int64, loginIP string) error {
	now := time.Now()
	dst := map[string]any{
		"last_login_at": &now,
		"last_login_ip": loginIP,
	}
	return db.WithContext(ctx).Table(u.TableName()).Where("`id` = ?", userID).Updates(dst).Error
}

// UpdateMfaSecret 更新MFA密钥
func (u *User) UpdateMfaSecret(ctx context.Context, db *gorm.DB, userID int64, secret []byte) error {
	dst := map[string]any{
		"mfa_secret": secret,
	}
	return db.WithContext(ctx).Table(u.TableName()).Where("`id` = ?", userID).Updates(dst).Error
}

// ClearMfaSecret 清空MFA密钥（解绑）
func (u *User) ClearMfaSecret(ctx context.Context, db *gorm.DB, userID int64) error {
	dst := map[string]any{
		"mfa_secret": nil,
	}
	return db.WithContext(ctx).Table(u.TableName()).Where("`id` = ?", userID).Updates(dst).Error
}

// Update 更新用户信息
func (u *User) Update(ctx context.Context, db *gorm.DB, userID int64, dst map[string]any) error {
	return db.WithContext(ctx).Table(u.TableName()).Where("`id` = ?", userID).Updates(dst).Error
}

// AccountExists 检查账号是否存在
func (u *User) AccountExists(ctx context.Context, db *gorm.DB, account string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Table(u.TableName()).Where("`account` = ?", account).Count(&count).Error
	return count > 0, err
}
