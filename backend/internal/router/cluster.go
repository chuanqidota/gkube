package router

import (
	"github.com/gin-gonic/gin"
	clusterApi "gkube/internal/cluster/api"
	"gkube/pkg/middleware"
)

// registerClusterRoutes 注册集群管理路由
func registerClusterRoutes(rg *gin.RouterGroup) {
	clusters := rg.Group("clusters")
	{
		clusters.GET("", clusterApi.Cluster.List)
		clusters.GET("/:id", clusterApi.Cluster.Detail)
		clusters.GET("/:id/check", clusterApi.Cluster.Check)
	}
	// 集群创建/更新/删除属高危操作,需管理员
	adminClusters := clusters.Group("", middleware.RequireAdmin())
	{
		adminClusters.POST("", clusterApi.Cluster.Create)
		adminClusters.PUT("/:id", clusterApi.Cluster.Update)
		adminClusters.DELETE("/:id", clusterApi.Cluster.Delete)
	}
}
