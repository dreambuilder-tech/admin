package member

import (
	"admin/internal/service/member_service"
	"wallet/common-lib/app"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	req := new(member_service.ListReq)
	if err := c.ShouldBindJSON(req); err != nil {
		app.InvalidParams(c, err.Error())
		return
	}

	resp, err := member_service.List(c.Request.Context(), req)
	if err != nil {
		app.InternalError(c, err.Error())
		return
	}

	app.ResultPage(c, resp.List, resp.Total)
}
