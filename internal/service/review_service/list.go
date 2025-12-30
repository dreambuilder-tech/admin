package review_service

import (
	"admin/internal/model"
	"context"
	"wallet/common-lib/dbs"
)

func List(ctx context.Context, page, size int) ([]*model.AgentApplication, int64, error) {
	m := new(model.AgentApplication)
	return m.GetList(ctx, dbs.Member, page, size)
}
