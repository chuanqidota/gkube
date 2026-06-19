package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gkube/pkg/auth"
	"gkube/pkg/response"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.FailWithStatus(c, http.StatusUnauthorized, "未提供认证 Token")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.FailWithStatus(c, http.StatusUnauthorized, "Token 格式错误")
			return
		}
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.FailWithStatus(c, http.StatusUnauthorized, "Token 无效或已过期")
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isSuperAdmin", claims.IsSuperAdmin)
		c.Next()
	}
}
