package router

import (
	"admin/internal/app/routerx"
	"admin/internal/common/auth"
	adminHandler "admin/internal/handler/admin"
	"admin/internal/handler/member"
	"admin/internal/handler/perm"
	"admin/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine) {
	engine.GET("ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r := engine.Group("/api/admin/v1")
	adminRouter(r)
	memberRouter(r)
}

// adminRouter 管理员相关路由
func adminRouter(r *gin.RouterGroup) {
	a := r.Group("/admin")

	// 不需要认证的接口
	routerx.Post(a, "/login", adminHandler.Login)

	// 需要认证的接口
	authGroup := a.Group("")
	authGroup.Use(middleware.Auth())
	{
		routerx.Get(authGroup, "/roles", adminHandler.GetRoles)
		routerx.Post(authGroup, "/create", adminHandler.CreateAdmin)
		routerx.Post(authGroup, "/mfa/generate", adminHandler.GenerateMFASecret)
		routerx.Post(authGroup, "/mfa/bind", adminHandler.BindMFA)
		routerx.Post(authGroup, "/mfa/unbind", adminHandler.UnbindMFA)
	}
}

func memberRouter(r *gin.RouterGroup) {
	m := r.Group("/member")
	routerx.PostPerm(m, "/list", auth.MemberList, member.List)

	p := r.Group("/permissions")
	p.GET("", perm.GetAllPermissions)
	p.GET("/current", perm.GetCurrentUserPermissions)
	u := r.Group("/user")
	u.GET("/:uid/permissions", perm.GetUserPermissions)
	u.POST("/:uid/permissions", perm.UpdateUserPermissions)
}
