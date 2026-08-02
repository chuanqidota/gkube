package router

import (
	"github.com/gin-gonic/gin"
	cluster "gkube/internal/cluster"
	"gkube/pkg/middleware"
)

// registerClusterRoutes 注册集群管理路由
func registerClusterRoutes(rg *gin.RouterGroup) {
	clusters := rg.Group("clusters")
	{
		clusters.GET("", cluster.Cluster.List)
		clusters.GET("/:id", cluster.Cluster.Detail)
		clusters.GET("/:id/check", cluster.Cluster.Check)
	}
	// 集群创建/更新/删除属高危操作,需管理员
	adminClusters := clusters.Group("", middleware.RequireAdmin())
	{
		adminClusters.POST("", cluster.Cluster.Create)
		adminClusters.PUT("/:id", cluster.Cluster.Update)
		adminClusters.DELETE("/:id", cluster.Cluster.Delete)
	}
}
