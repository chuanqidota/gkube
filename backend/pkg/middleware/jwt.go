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
		// WebSocket 路由优先消费一次性 ticket(?ticket=),消费即失效。
		// ticket 不与 Authorization header 同时使用,且仅限 WS 场景。
		if ticket := c.Query("ticket"); ticket != "" {
			userID, username, ok := auth.ConsumeTicket(ticket)
			if !ok {
				response.FailWithStatus(c, http.StatusUnauthorized, "Ticket 无效或已过期")
				return
			}
			c.Set("userID", userID)
			c.Set("username", username)
			c.Next()
			return
		}

		var tokenStr string

		// 从 Authorization header 读取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenStr = parts[1]
			}
		}

		// WebSocket 连接无法设置自定义 header,兼容从 query 参数读取 token
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
		c.Next()
	}
}
