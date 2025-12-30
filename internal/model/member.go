package model

import (
	"context"
	"time"
	"wallet/common-lib/consts/member_role"
	"wallet/common-lib/consts/member_status"

	"gorm.io/gorm"
)

type Member struct {
	ID             int64              `gorm:"column:id"`
	Account        *string            `gorm:"column:account"`
	AreaCode       *string            `gorm:"column:area_code"`
	Phone          []byte             `gorm:"column:phone"`
	PhoneDigest    *string            `json:"phone_digest"`
	Password       string             `gorm:"column:password"`
	PIN            string             `gorm:"column:pin"`
	PayeeCode      string             `gorm:"column:payee_code"`
	RealName       []byte             `gorm:"column:real_name"`
	RealNameDigest string             `gorm:"column:real_name_digest"`
	Nickname       string             `gorm:"column:nickname"`
	Avatar         string             `gorm:"column:avatar"`
	Email          string             `gorm:"column:email"`
	Lang           string             `gorm:"column:lang"`
	Status         member_status.Code `gorm:"column:status"`
	Role           member_role.Code   `gorm:"column:role"`
	LastLoginIP    string             `gorm:"column:last_login_ip"`
	LoginTimes     int                `gorm:"column:login_times"`
	LastLoginAt    int64              `gorm:"column:last_login_at"`
	RegisterIP     string             `gorm:"column:register_ip"`
	CreatedAt      time.Time          `gorm:"column:created_at"`
	UpdatedAt      time.Time          `gorm:"column:updated_at"`
}

func (*Member) TableName() string {
	return "members"
}

func (m *Member) GetList(ctx context.Context, db *gorm.DB, page, pageSize int, account, phone string) ([]*Member, int64, error) {
	var list []*Member
	var total int64

	offset := (page - 1) * pageSize
	query := db.WithContext(ctx).Table(m.TableName())
	if account != "" {
		query = query.Where("`account` LIKE ?", "%"+account+"%")
	}
	if phone != "" {
		query = query.Where("`phone` LIKE ?", "%"+phone+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("`id` DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (m *Member) UpdateRole(ctx context.Context, db *gorm.DB, memberID int64, role member_role.Code) error {
	return db.WithContext(ctx).Table(m.TableName()).Where("`id` = ? AND `role` <> ?", memberID, role).UpdateColumn("role", role).Error
}
