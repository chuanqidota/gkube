package router

import (
	"github.com/gin-gonic/gin"
	dashboardApi "gkube/internal/dashboard/api"
)

// registerDashboardRoutes 注册仪表盘路由
func registerDashboardRoutes(rg *gin.RouterGroup) {
	dashboard := rg.Group("dashboard")
	{
		dashboard.GET("overview", dashboardApi.Dashboard.Overview)
		dashboard.GET("resources", dashboardApi.Dashboard.Resources)
		dashboard.GET("workloads", dashboardApi.Dashboard.Workloads)
		dashboard.GET("events", dashboardApi.Dashboard.Events)
	}
}
