package middleware

import (
	"github.com/gin-gonic/gin"
	"gkube/app/auth/model"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

func RBAC(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isSuperAdmin, exists := c.Get("isSuperAdmin")
		if exists && isSuperAdmin.(bool) {
			c.Next()
			return
		}
		userID, exists := c.Get("userID")
		if !exists {
			response.Fail(c, "未获取到用户信息")
			c.Abort()
			return
		}
		var user model.User
		if err := database.DB.Preload("Roles.Permissions").First(&user, userID).Error; err != nil {
			response.Fail(c, "用户不存在")
			c.Abort()
			return
		}
		authorized := false
		for _, role := range user.Roles {
			for _, perm := range role.Permissions {
				if perm.Resource == resource && perm.Action == action {
					authorized = true
					break
				}
				if perm.Resource == "*" && perm.Action == "*" {
					authorized = true
					break
				}
			}
			if authorized {
				break
			}
		}
		if !authorized {
			response.Fail(c, "权限不足")
			c.Abort()
			return
		}
		c.Next()
	}
}
