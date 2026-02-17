package tools

import (
	"k-admin-system/model/common"
	"k-admin-system/service/tools"

	"github.com/gin-gonic/gin"
)

type CodeGeneratorAPI struct {
	Service *tools.CodeGeneratorService
}

// GetTableMetadata 获取表元数据
// @Summary 获取表元数据
// @Description 获取指定表的元数据信息，包括列名、类型、约束等
// @Tags Code Generator
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Success 200 {object} common.Response{data=tools.TableMetadata} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/gen/metadata/{tableName} [get]
func (api *CodeGeneratorAPI) GetTableMetadata(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		common.Fail(c, "table name is required")
		return
	}

	metadata, err := api.Service.GetTableMetadata(tableName)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	// Convert columns to field configs
	fields := make([]tools.FieldConfig, 0, len(metadata.Columns))
	for _, col := range metadata.Columns {
		fields = append(fields, tools.ConvertColumnToField(col))
	}

	result := map[string]interface{}{
		"table_name":    metadata.TableName,
		"table_comment": metadata.TableComment,
		"fields":        fields,
	}

	common.OkWithData(c, result)
}

// GenerateCode 生成代码
// @Summary 生成代码
// @Description 根据配置生成后端和前端代码，并写入文件
// @Tags Code Generator
// @Accept json
// @Produce json
// @Param config body tools.GenerateConfig true "生成配置"
// @Success 200 {object} common.Response{data=map[string]string} "成功，返回生成的文件路径列表"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/gen/generate [post]
func (api *CodeGeneratorAPI) GenerateCode(c *gin.Context) {
	var config tools.GenerateConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		common.Fail(c, "invalid request: "+err.Error())
		return
	}

	// Validate required fields
	if config.TableName == "" {
		common.Fail(c, "table_name is required")
		return
	}
	if config.StructName == "" {
		common.Fail(c, "struct_name is required")
		return
	}
	if config.PackageName == "" {
		common.Fail(c, "package_name is required")
		return
	}

	// Generate code
	files, err := api.Service.GenerateCode(config)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	// Write files to disk
	if err := api.Service.WriteGeneratedCode(files); err != nil {
		common.Fail(c, "failed to write files: "+err.Error())
		return
	}

	// Return list of generated file paths
	filePaths := make([]string, 0, len(files))
	for path := range files {
		filePaths = append(filePaths, path)
	}

	common.OkWithData(c, map[string]interface{}{
		"files": filePaths,
		"count": len(filePaths),
	})
}

// PreviewCode 预览代码
// @Summary 预览生成的代码
// @Description 根据配置生成代码预览，不写入文件
// @Tags Code Generator
// @Accept json
// @Produce json
// @Param config body tools.GenerateConfig true "生成配置"
// @Success 200 {object} common.Response{data=map[string]string} "成功，返回文件路径和内容的映射"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/gen/preview [post]
func (api *CodeGeneratorAPI) PreviewCode(c *gin.Context) {
	var config tools.GenerateConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		common.Fail(c, "invalid request: "+err.Error())
		return
	}

	// Validate required fields
	if config.TableName == "" {
		common.Fail(c, "table_name is required")
		return
	}
	if config.StructName == "" {
		common.Fail(c, "struct_name is required")
		return
	}
	if config.PackageName == "" {
		common.Fail(c, "package_name is required")
		return
	}

	// Preview code (no file writing)
	files, err := api.Service.PreviewCode(config)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, files)
}

// CreateTable 创建表
// @Summary 创建数据库表
// @Description 根据字段定义创建新的数据库表
// @Tags Code Generator
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "创建表请求" example({"table_name":"test_table","fields":[{"column_name":"name","field_type":"varchar(100)","nullable":false,"comment":"名称"}]})
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/gen/table [post]
func (api *CodeGeneratorAPI) CreateTable(c *gin.Context) {
	var req struct {
		TableName string              `json:"table_name" binding:"required"`
		Fields    []tools.FieldConfig `json:"fields" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request: "+err.Error())
		return
	}

	if len(req.Fields) == 0 {
		common.Fail(c, "at least one field is required")
		return
	}

	if err := api.Service.CreateTable(req.TableName, req.Fields); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "table created successfully")
}
