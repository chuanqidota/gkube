package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gkube/pkg/auth"
	"gkube/pkg/response"
)

// RequireAdmin 校验当前登录用户是否在管理员白名单(config.Conf.Security.AdminUsers)内,否则 403。
// 依赖 JWTAuth 已注入 username 到 context。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
			return
		}
		name, ok := username.(string)
		if !ok || !auth.IsAdmin(name) {
			response.FailWithStatus(c, http.StatusForbidden, "权限不足")
			return
		}
		c.Next()
	}
}
