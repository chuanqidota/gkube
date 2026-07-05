package router

import (
	"github.com/gin-gonic/gin"
	clusterApi "gkube/internal/cluster/api"
	"gkube/pkg/middleware"
)

// registerClusterRoutes 注册集群管理路由
func registerClusterRoutes(rg *gin.RouterGroup) {
	clusters := rg.Group("clusters", middleware.RBAC("cluster", "*"))
	{
		clusters.GET("", clusterApi.Cluster.List)
		clusters.POST("", clusterApi.Cluster.Create)
		clusters.GET("/:id", clusterApi.Cluster.Detail)
		clusters.PUT("/:id", clusterApi.Cluster.Update)
		clusters.DELETE("/:id", clusterApi.Cluster.Delete)
		clusters.GET("/:id/check", clusterApi.Cluster.Check)
	}
}
