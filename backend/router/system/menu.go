package system

import (
	"k-admin-system/api/v1/system"
	"k-admin-system/middleware"

	"github.com/gin-gonic/gin"
)

// InitMenuRouter 初始化菜单路由
func InitMenuRouter(router *gin.RouterGroup) {
	menuApi := system.MenuApi{}

	// 受保护的路由（需要JWT认证）
	protectedGroup := router.Group("/menu")
	protectedGroup.Use(middleware.JWTAuth())
	{
		// 菜单CRUD操作
		protectedGroup.POST("", menuApi.CreateMenu)
		protectedGroup.PUT("", menuApi.UpdateMenu)
		protectedGroup.DELETE("/:id", menuApi.DeleteMenu)
		protectedGroup.GET("/:id", menuApi.GetMenu)
		protectedGroup.GET("/all", menuApi.GetAllMenus)
		protectedGroup.GET("/tree", menuApi.GetMenuTree)
	}
}
