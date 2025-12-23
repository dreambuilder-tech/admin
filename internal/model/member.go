package model

import (
	"context"

	"gorm.io/gorm"
)

type Member struct {
	ID          int64  `gorm:"column:id" json:"id"`
	Account     string `gorm:"column:account" json:"account"`
	AreaCode    string `gorm:"column:area_code" json:"area_code"`
	Phone       string `gorm:"column:phone" json:"phone"`
	NickName    string `gorm:"column:nickname" json:"nickname"`
	Avatar      string `gorm:"column:avatar" json:"avatar"`
	Email       string `gorm:"column:email" json:"email"`
	Lang        string `gorm:"column:lang" json:"lang"`
	Status      int    `gorm:"column:status" json:"status"`
	LastLoginIP string `gorm:"column:last_login_ip" json:"last_login_ip"`
	LoginTimes  int    `gorm:"column:login_times" json:"login_times"`
	LastLoginAt int64  `gorm:"column:last_login_at" json:"last_login_at"`
	RegisterIP  string `gorm:"column:register_ip" json:"register_ip"`
	CreatedAt   int64  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   int64  `gorm:"column:updated_at" json:"updated_at"`
}

func (*Member) TableName() string {
	return "members"
}

func (m *Member) GetList(ctx context.Context, db *gorm.DB, page, pageSize int, account, phone string) ([]*Member, int64, error) {
	var list []*Member
	var total int64

	offset := (page - 1) * pageSize

	// 构建查询条件
	query := db.WithContext(ctx).Table(m.TableName())

	// 账号筛选（模糊查询）
	if account != "" {
		query = query.Where("`account` LIKE ?", "%"+account+"%")
	}

	// 手机号筛选（模糊查询）
	if phone != "" {
		query = query.Where("`phone` LIKE ?", "%"+phone+"%")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	if err := query.Order("`id` DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
