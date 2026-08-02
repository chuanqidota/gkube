package router

import (
	"github.com/gin-gonic/gin"
	dashboard "gkube/internal/dashboard"
)

// registerDashboardRoutes 注册仪表盘路由
func registerDashboardRoutes(rg *gin.RouterGroup) {
	dash := rg.Group("dashboard")
	{
		dash.GET("overview", dashboard.Dashboard.Overview)
		dash.GET("resources", dashboard.Dashboard.Resources)
		dash.GET("workloads", dashboard.Dashboard.Workloads)
		dash.GET("namespaces", dashboard.Dashboard.Namespaces)
		dash.GET("health", dashboard.Dashboard.Health)
		dash.GET("events", dashboard.Dashboard.Events)
	}
}
