package middleware

import (
	"admin/internal/common/auth"
	"encoding/json"
	"errors"
	"wallet/common-lib/app"
	"wallet/common-lib/rdb"
	"wallet/common-lib/zapx"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sid := auth.GetSessionID(c)
		if sid == "" {
			app.Unauthorized(c, "empty session")
			return
		}
		var (
			rds = rdb.Client
			ctx = c.Request.Context()
		)
		data, err := rds.Get(ctx, sid).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				app.Unauthorized(c, "session expired")
			} else {
				app.Unauthorized(c, "auth error")
			}
			zapx.ErrorCtx(ctx, "read session cache error", zap.Error(err))
			return
		}
		user := new(auth.User)
		if err = json.Unmarshal(data, user); err != nil {
			zapx.ErrorCtx(ctx, "unmarshal session cache error", zap.Error(err))
			app.Unauthorized(c, "invalid session data")
			return
		}
		_ = rds.Expire(ctx, sid, auth.SessionExpireTime)

		c.Set(auth.ReqAdminID, user.ID)
		c.Set(auth.ReqRoleID, user.Role)
		c.Set(auth.ReqAdminAccount, user.Account)

		c.Next()
	}
}
