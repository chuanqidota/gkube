package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gkube/config"
)

// IsOriginAllowed 判断 Origin 是否在配置的白名单内。空配置视为允许所有来源
// (开发友好;生产环境应通过 cors_origins 显式收紧白名单)。
func IsOriginAllowed(origin string) bool {
	origins := config.Conf.Security.CORSOrigins
	if len(origins) == 0 {
		// 未配置白名单:允许所有来源
		return true
	}
	for _, o := range origins {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

// CORSMiddleware 跨域中间件:按请求 Origin 反射白名单,Allow-Credentials 仅与具体 Origin 配套。
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" && IsOriginAllowed(origin) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Vary", "Origin")
		}
		// origin 不在白名单时不设置 ACAO,浏览器将拒绝跨域请求

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if c.Request.Method == http.MethodOptions {
			if origin != "" && !IsOriginAllowed(origin) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
