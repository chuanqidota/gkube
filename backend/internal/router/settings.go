package router

import (
	"github.com/gin-gonic/gin"
	settingsApi "gkube/internal/settings/api"
)

// registerSettingsRoutes 注册系统设置路由
func registerSettingsRoutes(rg *gin.RouterGroup) {
	settings := rg.Group("settings")
	{
		settings.GET("auth", settingsApi.Settings.GetAuthSettings)
		settings.PUT("auth", settingsApi.Settings.UpdateAuthSettings)
	}
}
