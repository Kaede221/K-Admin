package tools

import (
	"k-admin-system/api/v1/tools"
	"k-admin-system/global"
	"k-admin-system/middleware"
	toolsService "k-admin-system/service/tools"

	"github.com/gin-gonic/gin"
)

// InitCodeGeneratorRouter 初始化代码生成器路由
func InitCodeGeneratorRouter(router *gin.RouterGroup) {
	service := toolsService.NewCodeGeneratorService(global.DB)
	codeGenApi := &tools.CodeGeneratorAPI{
		Service: service,
	}

	// 所有Code Generator路由都需要JWT认证和管理员权限
	genGroup := router.Group("/gen")
	genGroup.Use(middleware.JWTAuth())
	// TODO: 添加Casbin中间件检查管理员权限
	// genGroup.Use(middleware.CasbinAuth())
	{
		// 获取表元数据
		genGroup.GET("/metadata/:tableName", codeGenApi.GetTableMetadata)

		// 代码生成
		genGroup.POST("/preview", codeGenApi.PreviewCode)
		genGroup.POST("/generate", codeGenApi.GenerateCode)

		// 表创建
		genGroup.POST("/table", codeGenApi.CreateTable)
	}
}
