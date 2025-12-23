package routerx

import (
	"admin/internal/common/auth"
	"admin/internal/middleware"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get(r *gin.RouterGroup, path string, h gin.HandlerFunc) {
	route(r, http.MethodGet, path, "", h)
}

func Post(r *gin.RouterGroup, path string, h gin.HandlerFunc) {
	route(r, http.MethodPost, path, "", h)
}

func GetPerm(r *gin.RouterGroup, path string, code auth.PermCode, h gin.HandlerFunc) {
	route(r, http.MethodGet, path, code, h)
}

func PostPerm(r *gin.RouterGroup, path string, code auth.PermCode, h gin.HandlerFunc) {
	route(r, http.MethodPost, path, code, h)
}

func route(r *gin.RouterGroup, method, path string, code auth.PermCode, h gin.HandlerFunc) {
	if code != "" {
		key := fmt.Sprintf("%s:%s%s", method, r.BasePath(), path)
		auth.AllRouterPerms[key] = code
	}
	r.Handle(method, path, middleware.CheckPerm(), h)
}
