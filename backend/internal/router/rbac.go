package router

import (
	"github.com/gin-gonic/gin"
	"gkube/internal/rbac"
	"gkube/pkg/middleware"
)

// registerRbacRoutes 注册 RBAC 权限管理路由
func registerRbacRoutes(rg *gin.RouterGroup) {
	// 所有已认证用户可访问（普通用户查自己权限、查看角色列表、查看集群成员）
	rg.GET("rbac/roles", rbac.RbacHandler.ListRoles)
	rg.GET("rbac/my-permissions", rbac.RbacHandler.MyPermissions)
	rg.GET("rbac/cluster-members", rbac.RbacHandler.ClusterMembers)

	// 仅管理员可访问（写操作 + 管理查询）
	// AuditLog 放在 RequireAdmin 之前，确保 403 失败尝试也被审计
	admin := rg.Group("rbac", middleware.AuditLog(), middleware.RequireAdmin())
	{
		admin.GET("bindings", rbac.RbacHandler.ListBindings)
		admin.POST("bindings", rbac.RbacHandler.CreateBinding)
		admin.PUT("bindings/:id", rbac.RbacHandler.UpdateBinding)
		admin.DELETE("bindings/:id", rbac.RbacHandler.DeleteBinding)
	}
}
