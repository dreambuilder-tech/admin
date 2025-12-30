package review

import (
	"admin/internal/common/auth"
	"admin/internal/service/review_service"
	"wallet/common-lib/app"
	"wallet/common-lib/dto/req_dto"
	"wallet/common-lib/zapx"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ListReq struct {
	req_dto.PageArgs
}

func List(c *gin.Context) {
	req := new(ListReq)
	if err := c.ShouldBindJSON(req); err != nil {
		app.InvalidParams(c, err.Error())
		return
	}
	req.Init()
	resp, cnt, err := review_service.List(c.Request.Context(), req.Page, req.Size)
	if err != nil {
		zapx.ErrorCtx(c.Request.Context(), "query review list error", zap.Error(err))
		app.InternalError(c, err.Error())
		return
	}
	app.ResultPage(c, resp, cnt)
}

type ApproveReq struct {
	ID int64 `json:"id"`
}

func Approve(c *gin.Context) {
	req := new(ApproveReq)
	if err := c.ShouldBindJSON(req); err != nil {
		app.InvalidParams(c, err.Error())
		return
	}
	if req.ID <= 0 {
		app.InvalidParams(c, "empty ID")
		return
	}
	err := review_service.Approve(c.Request.Context(), auth.AdminID(c), req.ID)
	if err != nil {
		zapx.ErrorCtx(c.Request.Context(), "review approve error", zap.Error(err))
		app.InternalError(c, err.Error())
		return
	}
	app.Success(c)
}

type RejectReq struct {
	ID     int64  `json:"id"`
	Reason string `json:"reason"`
}

func Reject(c *gin.Context) {
	req := new(RejectReq)
	if err := c.ShouldBindJSON(req); err != nil {
		app.InvalidParams(c, err.Error())
		return
	}
	if req.ID <= 0 || req.Reason == "" {
		app.InvalidParams(c, "empty ID or reason")
		return
	}
	err := review_service.Reject(c.Request.Context(), auth.AdminID(c), req.ID, req.Reason)
	if err != nil {
		app.InternalError(c, err.Error())
		return
	}
	app.Success(c)
}
