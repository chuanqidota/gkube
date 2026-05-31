package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gkube/pkg/auth"
	"gkube/pkg/response"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Fail(c, "未提供认证 Token")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, "Token 格式错误")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.Fail(c, "Token 无效或已过期")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isSuperAdmin", claims.IsSuperAdmin)
		c.Next()
	}
}
