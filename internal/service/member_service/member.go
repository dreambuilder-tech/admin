package member_service

import (
	"admin/internal/model"
	"context"
	"wallet/common-lib/dbs"
	"wallet/common-lib/dto/req_dto"
	"wallet/common-lib/zapx"

	"go.uber.org/zap"
)

type ListReq struct {
	req_dto.PageArgs
	Account string `json:"account"` // 账号筛选
	Phone   string `json:"phone"`   // 手机号筛选
}

type ListResp struct {
	List  []*model.Member `json:"list"`
	Total int64           `json:"total"`
}

func List(ctx context.Context, req *ListReq) (*ListResp, error) {
	// 初始化分页参数（使用 dto 的 Init 方法）
	req.PageArgs.Init()

	member := new(model.Member)
	list, total, err := member.GetList(ctx, dbs.Member, req.Page, req.Size, req.Account, req.Phone)
	if err != nil {
		zapx.ErrorCtx(ctx, "member.GetList failed", zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("size", req.Size),
			zap.String("account", req.Account),
			zap.String("phone", req.Phone))
		return nil, err
	}

	return &ListResp{
		List:  list,
		Total: total,
	}, nil
}
