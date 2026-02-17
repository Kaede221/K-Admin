package tools

import (
	"k-admin-system/api/v1/tools"
	"k-admin-system/middleware"

	"github.com/gin-gonic/gin"
)

// InitDBInspectorRouter 初始化数据库检查器路由
func InitDBInspectorRouter(router *gin.RouterGroup) {
	dbInspectorApi := &tools.DBInspectorAPI{}

	// 所有DB Inspector路由都需要JWT认证和管理员权限
	dbGroup := router.Group("/db")
	dbGroup.Use(middleware.JWTAuth())
	// TODO: 添加Casbin中间件检查管理员权限
	// dbGroup.Use(middleware.CasbinAuth())
	{
		// 表管理
		dbGroup.GET("/tables", dbInspectorApi.GetTables)
		dbGroup.GET("/tables/:tableName/schema", dbInspectorApi.GetTableSchema)
		dbGroup.GET("/tables/:tableName/data", dbInspectorApi.GetTableData)

		// 记录CRUD操作
		dbGroup.POST("/tables/:tableName/records", dbInspectorApi.CreateRecord)
		dbGroup.PUT("/tables/:tableName/records/:id", dbInspectorApi.UpdateRecord)
		dbGroup.DELETE("/tables/:tableName/records/:id", dbInspectorApi.DeleteRecord)

		// SQL执行（需要超级管理员权限）
		dbGroup.POST("/execute", dbInspectorApi.ExecuteSQL)
	}
}
