package system

import (
	"k-admin-system/api/v1/system"
	"k-admin-system/middleware"

	"github.com/gin-gonic/gin"
)

// InitRoleRouter 初始化角色路由
func InitRoleRouter(router *gin.RouterGroup) {
	roleApi := system.RoleApi{}

	// 受保护的路由（需要JWT认证和管理员权限）
	protectedGroup := router.Group("/role")
	protectedGroup.Use(middleware.JWTAuth())
	// TODO: 在Task 8实现Casbin后，添加管理员权限检查中间件
	// protectedGroup.Use(middleware.CasbinAuth())
	{
		// 角色CRUD操作
		protectedGroup.POST("", roleApi.CreateRole)
		protectedGroup.PUT("", roleApi.UpdateRole)
		protectedGroup.DELETE("/:id", roleApi.DeleteRole)
		protectedGroup.GET("/:id", roleApi.GetRole)
		protectedGroup.GET("/list", roleApi.GetRoleList)

		// 权限分配
		protectedGroup.POST("/assign-menus", roleApi.AssignMenus)
		protectedGroup.GET("/:id/menus", roleApi.GetRoleMenus)
		protectedGroup.POST("/assign-apis", roleApi.AssignAPIs)
		protectedGroup.GET("/:id/apis", roleApi.GetRoleAPIs)
	}
}
