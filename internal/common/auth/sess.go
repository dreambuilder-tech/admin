package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID      int64  `json:"id"`
	Account string `json:"account"`
	Role    int    `json:"role"`
	LoginIP string `json:"login_ip,omitempty"`
	LoginAt int64  `json:"login_at"`
}

const (
	SessionHeader     = "X-Session-Id"
	ReqUserID         = "userID"
	ReqUserAccount    = "userAccount"
	ReqRoleID         = "roleID"
	SessionExpireTime = 3 * 24 * time.Hour

	SuperAdminRoleID = 1 // super_admin角色的ID（对应roles表中的第一条记录）
)

func NewSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GetSessionID(c *gin.Context) string {
	return c.GetHeader(SessionHeader)
}

func SetSessionID(c *gin.Context, sid string) {
	c.Request.Header.Set(SessionHeader, sid)
}

func UserID(c *gin.Context) int64 {
	return c.GetInt64(ReqUserID)
}

func UserAccount(c *gin.Context) string {
	return c.GetString(ReqUserAccount)
}

func RoleID(c *gin.Context) int {
	return c.GetInt(ReqRoleID)
}

func IsSuperAdmin(c *gin.Context) bool {
	role := c.GetInt(ReqRoleID)
	return role == SuperAdminRoleID
}
