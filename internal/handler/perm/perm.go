package perm

import (
	"admin/internal/common/auth"
	"admin/internal/service/perm_service"
	"errors"
	"strconv"
	"wallet/common-lib/app"
	"wallet/common-lib/zapx"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UpdatePermissionsReq struct {
	Permissions []string `json:"permissions"`
}

func GetAllPermissions(c *gin.Context) {
	perms := auth.GetAllPerms()
	app.Result(c, gin.H{
		"permissions": perms,
	})
}

func GetCurrentUserPermissions(c *gin.Context) {
	currentUID := auth.AdminID(c)

	perms, err := perm_service.UserPerms(c.Request.Context(), currentUID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		zapx.ErrorCtx(c.Request.Context(), "failed to get current user permissions", zap.Error(err))
		app.InternalError(c, "failed to get current user permissions")
		return
	}

	if perms == nil {
		perms = []string{}
	}

	app.Result(c, gin.H{
		"permissions": perms,
	})
}

func GetUserPermissions(c *gin.Context) {
	uidStr := c.Param("uid")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		app.InvalidParams(c, "invalid uid format")
		return
	}

	perms, err := perm_service.UserPerms(c.Request.Context(), uid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		zapx.ErrorCtx(c.Request.Context(), "failed to get user permissions", zap.Error(err))
		app.InternalError(c, "failed to get user permissions")
		return
	}

	if perms == nil {
		perms = []string{}
	}

	app.Result(c, gin.H{
		"uid":         uid,
		"permissions": perms,
	})
}

func UpdateUserPermissions(c *gin.Context) {
	uidStr := c.Param("uid")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		app.InvalidParams(c, "invalid uid format")
		return
	}

	var req UpdatePermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.InvalidParams(c, "%s", err.Error())
		return
	}

	if err := perm_service.UpdateUserPermissions(c.Request.Context(), auth.AdminID(c), uid, req.Permissions); err != nil {
		zapx.ErrorCtx(c.Request.Context(), "failed to update user permissions", zap.Error(err))
		app.InternalError(c, "%s", err.Error())
		return
	}

	app.Success(c)
}
