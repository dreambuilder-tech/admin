package model

import (
	"context"
	"time"
	"wallet/common-lib/consts/agent_apply"
	"wallet/common-lib/consts/currency"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type AgentApplication struct {
	ID           int64                 `gorm:"column:id"           json:"id"`
	MemberID     int64                 `gorm:"column:member_id"    json:"member_id"`
	Direction    agent_apply.Direction `gorm:"column:direction"    json:"direction"`
	Status       agent_apply.Status    `gorm:"column:status"       json:"status"`
	Currency     currency.Code         `gorm:"column:currency"     json:"currency"`
	Deposit      decimal.Decimal       `gorm:"column:deposit"      json:"deposit"`
	FreezeID     int64                 `gorm:"column:freeze_id"    json:"-"`
	ReviewedBy   int64                 `gorm:"column:reviewed_by"  json:"reviewed_by"`
	RejectReason string                `gorm:"column:reject_reason" json:"reject_reason"`
	AppliedAt    time.Time             `gorm:"column:applied_at"   json:"applied_at"`
	ReviewedAt   *time.Time            `gorm:"column:reviewed_at"  json:"reviewed_at"`
	ReleasedAt   *time.Time            `gorm:"column:released_at"  json:"released_at"`
	CreatedAt    time.Time             `gorm:"column:created_at"   json:"created_at"`
	UpdatedAt    time.Time             `gorm:"column:updated_at"   json:"updated_at"`
}

func (*AgentApplication) TableName() string {
	return "agent_applications"
}

func (a *AgentApplication) GetOne(ctx context.Context, db *gorm.DB, id int64) error {
	return db.WithContext(ctx).Where("`id` = ?", id).Take(a).Error
}

func (a *AgentApplication) GetList(ctx context.Context, db *gorm.DB, page, size int) ([]*AgentApplication, int64, error) {
	var r []*AgentApplication
	var cnt int64
	err := db.WithContext(ctx).Table(a.TableName()).Count(&cnt).Offset((page - 1) * size).Limit(size).Order("id DESC").Find(&r).Error
	return r, cnt, err
}

func (a *AgentApplication) Reviewed(ctx context.Context, db *gorm.DB) error {
	dst := map[string]any{
		"status":        a.Status,
		"reviewed_by":   a.ReviewedBy,
		"reviewed_at":   a.ReviewedAt,
		"reject_reason": a.RejectReason,
	}
	return db.WithContext(ctx).Table(a.TableName()).Where("`id` = ? AND `status` <> ?", a.ID, a.Status).Updates(dst).Error
}

func (a *AgentApplication) Release(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Table(a.TableName()).Where("`id` = ? AND `released_at` IS NULL", a.ID).UpdateColumn("released_at", time.Now()).Error
}
