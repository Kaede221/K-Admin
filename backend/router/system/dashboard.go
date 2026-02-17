package system

import (
	"k-admin-system/api/v1/system"
	"k-admin-system/middleware"

	"github.com/gin-gonic/gin"
)

// InitDashboardRouter 初始化仪表盘路由
func InitDashboardRouter(router *gin.RouterGroup) {
	dashboardApi := system.DashboardApi{}

	// 受保护的路由（需要JWT认证）
	protectedGroup := router.Group("/dashboard")
	protectedGroup.Use(middleware.JWTAuth())
	{
		protectedGroup.GET("/stats", dashboardApi.GetDashboardStats)
	}
}
