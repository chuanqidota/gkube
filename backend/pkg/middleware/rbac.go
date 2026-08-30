package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gkube/pkg/auth"
	"gkube/pkg/response"
)

// RequireAdmin 校验当前登录用户是否在管理员白名单(config.Conf.Security.AdminUsers)内
// 或在数据库中标记为 is_super_admin，否则 403。
// 依赖 JWTAuth 已注入 username 和 userID 到 context。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, exists := c.Get("username")
		if !exists {
			response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
			return
		}
		name, ok := username.(string)
		if !ok {
			response.FailWithStatus(c, http.StatusForbidden, "权限不足")
			return
		}

		// config 白名单检查（内存，无 DB 查询）
		if auth.IsAdmin(name) {
			c.Next()
			return
		}

		// DB is_super_admin 检查（走缓存，5min TTL）
		if userIDVal, uidExists := c.Get("userID"); uidExists {
			if uid, ok := userIDVal.(uint); ok {
				cached := auth.GetUserPermissions(uid)
				if cached.IsSuperAdmin {
					c.Next()
					return
				}
			}
		}

		response.FailWithStatus(c, http.StatusForbidden, "权限不足")
	}
}
