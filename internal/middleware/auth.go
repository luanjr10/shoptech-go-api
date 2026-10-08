package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/luanjr10/shoptech-go-api/internal/response"
	"github.com/luanjr10/shoptech-go-api/internal/token"
)

const (
	CtxUserID = "userID"
	CtxRole   = "role"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || raw == "" {
			response.Error(c, http.StatusUnauthorized, "Bạn chưa đăng nhập", nil)
			return
		}
		claims, err := token.Parse(secret, raw)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Token không hợp lệ hoặc đã hết hạn", nil)
			return
		}
		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(roles, c.GetString(CtxRole)) {
			response.Error(c, http.StatusForbidden, "Bạn không có quyền thực hiện thao tác này", nil)
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) uint { return c.GetUint(CtxUserID) }
func Role(c *gin.Context) string { return c.GetString(CtxRole) }
