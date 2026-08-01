package router

import (
	"github.com/gin-gonic/gin"
	authApi "gkube/internal/auth/api"
	"gkube/pkg/middleware"
)

// registerPublicAuthRoutes 注册公开的认证路由（无需JWT）
func registerPublicAuthRoutes(rg *gin.RouterGroup) {
	rg.POST("auth/login", authApi.Auth.Login)
	rg.POST("auth/refresh", authApi.Auth.Refresh)
}

// registerAuthRoutes 注册需要认证的用户路由
func registerAuthRoutes(rg *gin.RouterGroup) {
	// WebSocket ticket 签发(需 JWT)
	rg.POST("auth/ws-ticket", authApi.Auth.WsTicket)

	// Users:仅查询/改密对登录用户开放;增删改需管理员
	users := rg.Group("users")
	{
		users.GET("", authApi.User.List)
		users.PUT("change-password", authApi.User.ChangePassword)
	}
	// 用户管理(创建/更新/删除)需管理员
	adminUsers := users.Group("", middleware.RequireAdmin())
	{
		adminUsers.POST("", authApi.User.Create)
		adminUsers.PUT("", authApi.User.Update)
		adminUsers.DELETE("", authApi.User.Delete)
	}
}
