package system

import (
	"k-admin-system/api/v1/system"
	"k-admin-system/middleware"

	"github.com/gin-gonic/gin"
)

// InitMenuRouter 初始化菜单路由
func InitMenuRouter(router *gin.RouterGroup) {
	menuApi := system.MenuApi{}

	// 受保护的路由（需要JWT认证和Casbin授权）
	protectedGroup := router.Group("/menu")
	protectedGroup.Use(middleware.JWTAuth())
	protectedGroup.Use(middleware.CasbinAuth())
	{
		// 菜单CRUD操作
		protectedGroup.POST("", menuApi.CreateMenu)
		protectedGroup.PUT("", menuApi.UpdateMenu)
		protectedGroup.DELETE("/:id", menuApi.DeleteMenu)
		protectedGroup.GET("/:id", menuApi.GetMenu)
		protectedGroup.GET("/all", menuApi.GetAllMenus)
	}

	// 菜单树查询（仅需要JWT认证，不需要Casbin授权）
	// 因为该接口根据roleId过滤菜单，已经实现了权限控制
	menuTreeGroup := router.Group("/menu")
	menuTreeGroup.Use(middleware.JWTAuth())
	{
		menuTreeGroup.GET("/tree", menuApi.GetMenuTree)
	}
}
