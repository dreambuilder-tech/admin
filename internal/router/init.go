package router

import (
	"admin/internal/app/routerx"
	"admin/internal/common/auth"
	adminHandler "admin/internal/handler/admin"
	"admin/internal/handler/agent/review"
	"admin/internal/handler/member"
	"admin/internal/handler/perm"
	"admin/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Init(engine *gin.Engine) {
	engine.GET("ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r := engine.Group("/api/v1")
	adminRouter(r)
	memberRouter(r.Group("/member"))
	permRouter(r.Group("/perm"))
	agentRouter(r.Group("/agent"))
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

func permRouter(r *gin.RouterGroup) {
	r.GET("/", perm.GetAllPermissions)
	r.GET("/current", perm.GetCurrentUserPermissions)
	r.GET("/:uid", perm.GetUserPermissions)
	r.POST("/:uid", perm.UpdateUserPermissions)
}

func memberRouter(r *gin.RouterGroup) {
	routerx.PostPerm(r, "/list", auth.MemberList, member.List)
}

func agentRouter(r *gin.RouterGroup) {
	re := r.Group("/review", middleware.Auth())
	{
		re.POST("/list", review.List)
		re.POST("/approve", review.Approve)
		re.POST("/reject", review.Reject)
	}
}
