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
		var tokenStr string

		// 优先从 Authorization header 读取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// WebSocket 连接无法设置自定义 header，从 query 参数读取 token
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			response.FailWithStatus(c, http.StatusUnauthorized, "未提供认证 Token")
			return
		}

		claims, err := auth.ParseToken(tokenStr)
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
