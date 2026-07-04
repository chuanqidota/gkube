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
	rg.GET("auth/oidc/login", authApi.OIDC.GetLoginURL)
	rg.GET("auth/oidc/callback", authApi.OIDC.HandleCallback)
}

// registerAuthRoutes 注册需要认证的用户/角色/RBAC路由
func registerAuthRoutes(rg *gin.RouterGroup) {
	rg.GET("auth/me", authApi.Auth.GetMe)

	// Users
	users := rg.Group("users", middleware.RBAC("user", "*"))
	{
		users.GET("", authApi.User.List)
		users.POST("", authApi.User.Create)
		users.PUT("", authApi.User.Update)
		users.DELETE("", authApi.User.Delete)
		users.PUT("change-password", authApi.User.ChangePassword)
	}

	// Roles
	roles := rg.Group("roles", middleware.RBAC("role", "*"))
	{
		roles.GET("", authApi.Role.List)
		roles.POST("", authApi.Role.Create)
		roles.PUT("", authApi.Role.Update)
		roles.DELETE("", authApi.Role.Delete)
	}

	// RBAC Matrix
	rg.GET("rbac/matrix", authApi.RBACMatrix.GetMatrix)
	rg.GET("rbac/matrix/role", authApi.RBACMatrix.GetRoleMatrix)
	rg.PUT("rbac/matrix", authApi.RBACMatrix.UpdateMatrix)
	rg.POST("rbac/matrix/toggle", authApi.RBACMatrix.TogglePermission)
	rg.POST("rbac/matrix/init", authApi.RBACMatrix.InitializeMatrix)
}
