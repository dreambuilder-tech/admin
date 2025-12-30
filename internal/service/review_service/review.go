package review_service

import (
	"admin/internal/model"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
	"wallet/common-lib/consts/agent_apply"
	"wallet/common-lib/consts/ledger"
	"wallet/common-lib/consts/member_role"
	"wallet/common-lib/consts/rds_keys"
	"wallet/common-lib/dbs"
	"wallet/common-lib/errs"
	"wallet/common-lib/rdb"
	"wallet/common-lib/rpcx/contracts"
	"wallet/common-lib/rpcx/wallet_rpcx"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func Approve(ctx context.Context, adminID, applyID int64) error {
	m := new(model.AgentApplication)
	if err := m.GetOne(ctx, dbs.Member, applyID); err != nil {
		return err
	}
	if m.MemberID <= 0 {
		return errors.New("empty member ID")
	}
	// redis lock.
	l := rdb.NewLock(rdb.Client, rds_keys.AgentApply(m.MemberID))
	if _, err := l.Lock(ctx); err != nil {
		return err
	}
	defer l.Unlock()

	// double check.
	if err := m.GetOne(ctx, dbs.Member, applyID); err != nil {
		return err
	}
	if m.FreezeID <= 0 {
		return errors.New("freeze ID lost")
	}
	if m.Status == agent_apply.Approved {
		return nil
	}
	if m.Status != agent_apply.Reviewing {
		return fmt.Errorf("wrong status %s", m.Status)
	}

	// 先改状态，后解冻
	errX := dbs.Member.Transaction(func(tx *gorm.DB) error {
		role := member_role.Member
		if m.Direction == agent_apply.ToAgent {
			role = member_role.Agent
		}
		member := new(model.Member)
		if err := member.UpdateRole(ctx, tx, m.MemberID, role); err != nil {
			return err
		}
		now := time.Now()
		m.Status = agent_apply.Approved
		m.ReviewedBy = adminID
		m.ReviewedAt = &now
		if err := m.Reviewed(ctx, dbs.Member); err != nil {
			return err
		}
		return nil
	})
	if errX != nil {
		return errX
	}
	// 代理->用户，通过: 退还保证金
	// 用户->代理，通过: 不用操作保证金
	if m.Direction == agent_apply.ToMember {
		_, err := wallet_rpcx.Release(ctx, &contracts.ReleaseReq{
			FreezeID: m.FreezeID,
			Delta:    decimal.Zero, // means all
			Reason:   ledger.ApplyToMember_Refund,
			TxID:     strconv.FormatInt(m.ID, 10),
			Desc:     "商户申请退还保证金-审核通过-解冻",
		})
		if err != nil && !errs.IsErrIdempotentHit(err) {
			return err
		}
		if err = m.Release(ctx, dbs.Member); err != nil {
			return fmt.Errorf("update release time error: %s", err.Error())
		}
	}
	return nil
}

func Reject(ctx context.Context, adminID, applyID int64, reason string) error {
	if adminID <= 0 {
		return errors.New("empty admin ID")
	}
	if applyID <= 0 {
		return errors.New("empty apply ID")
	}
	m := new(model.AgentApplication)
	if err := m.GetOne(ctx, dbs.Member, applyID); err != nil {
		return err
	}

	// redis lock.
	l := rdb.NewLock(rdb.Client, rds_keys.AgentApply(m.MemberID))
	if _, err := l.Lock(ctx); err != nil {
		return err
	}
	defer l.Unlock()

	// double check.
	if err := m.GetOne(ctx, dbs.Member, applyID); err != nil {
		return err
	}
	if m.Status == agent_apply.Rejected {
		return nil
	}
	if m.Status != agent_apply.Reviewing {
		return fmt.Errorf("wrong status %s", m.Status)
	}
	if m.FreezeID <= 0 {
		return errors.New("freeze ID lost")
	}

	// 先改状态，后解冻
	now := time.Now()
	m.Status = agent_apply.Rejected
	m.RejectReason = reason
	m.ReviewedBy = adminID
	m.ReviewedAt = &now
	if err := m.Reviewed(ctx, dbs.Member); err != nil {
		return err
	}

	// 用户->代理，拒绝: 解冻保证金
	// 代理->用户，拒绝: 不用操作保证金
	if m.Direction == agent_apply.ToAgent {
		_, err := wallet_rpcx.Release(ctx, &contracts.ReleaseReq{
			FreezeID: m.FreezeID,
			Delta:    decimal.Zero, // means all
			Reason:   ledger.ApplyToAgent_Reject,
			TxID:     strconv.FormatInt(m.ID, 10),
			Desc:     "用户申请成为商户-审核拒绝-解冻",
		})
		if err != nil && !errs.IsErrIdempotentHit(err) {
			return err
		}
		if err = m.Release(ctx, dbs.Member); err != nil {
			return fmt.Errorf("update release time error: %s", err.Error())
		}
	}
	return nil
}
