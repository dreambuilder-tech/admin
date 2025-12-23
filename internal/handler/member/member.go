package member

import (
	"admin/internal/app"
	"admin/internal/service/member_service"

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

	app.SuccessPage(c, resp.List, resp.Total)
}
