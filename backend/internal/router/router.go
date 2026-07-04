package router

import (
	"github.com/gin-gonic/gin"
	"gkube/pkg/middleware"
)

func Engine() *gin.Engine {
	router := gin.Default()

	// 全局CORS中间件
	router.Use(middleware.CORSMiddleware())

	v1 := router.Group("v1")

	// 公开路由（无需认证）
	registerPublicAuthRoutes(v1)

	// 需要JWT认证的路由
	authorized := v1.Group("", middleware.JWTAuth())
	{
		registerAuthRoutes(authorized)
		registerClusterRoutes(authorized)
		registerDashboardRoutes(authorized)
		registerSettingsRoutes(authorized)
		registerK8sRoutes(authorized)
	}

	return router
}
