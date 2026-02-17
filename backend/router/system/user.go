package system

import (
	"k-admin-system/api/v1/system"
	"k-admin-system/middleware"

	"github.com/gin-gonic/gin"
)

// InitUserRouter 初始化用户路由
func InitUserRouter(router *gin.RouterGroup) {
	userApi := system.UserApi{}

	// 公共路由（不需要JWT认证）
	publicGroup := router.Group("/user")
	{
		publicGroup.POST("/login", userApi.Login)
	}

	// 受保护的路由（需要JWT认证）
	protectedGroup := router.Group("/user")
	protectedGroup.Use(middleware.JWTAuth())
	{
		// 用户CRUD操作
		protectedGroup.POST("", userApi.CreateUser)
		protectedGroup.PUT("", userApi.UpdateUser)
		protectedGroup.DELETE("/:id", userApi.DeleteUser)
		protectedGroup.GET("/:id", userApi.GetUser)
		protectedGroup.GET("/list", userApi.GetUserList)

		// 密码管理
		protectedGroup.POST("/change-password", userApi.ChangePassword)
		protectedGroup.POST("/reset-password", userApi.ResetPassword)

		// 状态管理
		protectedGroup.POST("/toggle-status", userApi.ToggleStatus)
	}
}
