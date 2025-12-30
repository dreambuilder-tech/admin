package middleware

import (
	"admin/internal/common/auth"
	"admin/internal/service/perm_service"
	"fmt"
	"wallet/common-lib/app"

	"github.com/gin-gonic/gin"
)

func CheckPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("%s:%s", c.Request.Method, c.FullPath())
		if perm, ok := auth.AllRouterPerms[key]; ok {
			if !check(c, perm) {
				app.PermissionDenied(c)
				return
			}
		}
		c.Next()
	}
}

func check(c *gin.Context, perm auth.PermCode) bool {
	userID := auth.AdminID(c)
	if userID <= 0 {
		return false
	}
	return perm_service.CheckPerms(c, userID, perm)
}
