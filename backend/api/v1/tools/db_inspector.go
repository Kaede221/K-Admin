package tools

import (
	"k-admin-system/model/common"
	"k-admin-system/service/tools"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DBInspectorAPI struct {
	service tools.DBInspectorService
}

// GetTables 获取所有表
// @Summary 获取数据库所有表
// @Description 获取当前数据库中的所有表名列表
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=[]string} "成功"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/tables [get]
func (api *DBInspectorAPI) GetTables(c *gin.Context) {
	tables, err := api.service.GetTables()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OkWithData(c, tables)
}

// GetTableSchema 获取表结构
// @Summary 获取表结构信息
// @Description 获取指定表的列信息，包括列名、类型、是否可空等
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Success 200 {object} common.Response{data=[]tools.ColumnInfo} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/tables/{tableName}/schema [get]
func (api *DBInspectorAPI) GetTableSchema(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		common.Fail(c, "table name is required")
		return
	}

	schema, err := api.service.GetTableSchema(tableName)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}
	common.OkWithData(c, schema)
}

// GetTableData 获取表数据
// @Summary 获取表数据
// @Description 分页获取指定表的数据记录
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} common.Response{data=map[string]interface{}} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/tables/{tableName}/data [get]
func (api *DBInspectorAPI) GetTableData(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		common.Fail(c, "table name is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	data, total, err := api.service.GetTableData(tableName, page, pageSize)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, map[string]interface{}{
		"list":     data,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// ExecuteSQL 执行SQL语句
// @Summary 执行SQL语句
// @Description 执行自定义SQL语句，支持查询和修改操作
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "SQL请求" example({"sql":"SELECT * FROM users","readOnly":false})
// @Success 200 {object} common.Response{data=interface{}} "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/execute [post]
func (api *DBInspectorAPI) ExecuteSQL(c *gin.Context) {
	var req struct {
		SQL      string `json:"sql" binding:"required"`
		ReadOnly bool   `json:"readOnly"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request: "+err.Error())
		return
	}

	// TODO: 检查超级管理员权限（危险操作）
	// 这里应该从JWT claims中获取用户角色，检查是否为超级管理员
	// 如果不是超级管理员且SQL包含危险操作，应该拒绝

	result, err := api.service.ExecuteSQL(req.SQL, req.ReadOnly)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, result)
}

// CreateRecord 创建记录
// @Summary 创建表记录
// @Description 在指定表中创建新记录
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param data body map[string]interface{} true "记录数据"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/tables/{tableName}/records [post]
func (api *DBInspectorAPI) CreateRecord(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		common.Fail(c, "table name is required")
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		common.Fail(c, "invalid request: "+err.Error())
		return
	}

	if err := api.service.CreateRecord(tableName, data); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "record created successfully")
}

// UpdateRecord 更新记录
// @Summary 更新表记录
// @Description 更新指定表中的记录
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path string true "记录ID"
// @Param data body map[string]interface{} true "更新数据"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 404 {object} common.Response "记录不存在"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/tables/{tableName}/records/{id} [put]
func (api *DBInspectorAPI) UpdateRecord(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		common.Fail(c, "table name is required")
		return
	}

	id := c.Param("id")
	if id == "" {
		common.Fail(c, "record id is required")
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		common.Fail(c, "invalid request: "+err.Error())
		return
	}

	if err := api.service.UpdateRecord(tableName, id, data); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "record updated successfully")
}

// DeleteRecord 删除记录
// @Summary 删除表记录
// @Description 删除指定表中的记录
// @Tags DB Inspector
// @Accept json
// @Produce json
// @Param tableName path string true "表名"
// @Param id path string true "记录ID"
// @Success 200 {object} common.Response "成功"
// @Failure 400 {object} common.Response "参数错误"
// @Failure 404 {object} common.Response "记录不存在"
// @Failure 500 {object} common.Response "失败"
// @Security ApiKeyAuth
// @Router /tools/db/tables/{tableName}/records/{id} [delete]
func (api *DBInspectorAPI) DeleteRecord(c *gin.Context) {
	tableName := c.Param("tableName")
	if tableName == "" {
		common.Fail(c, "table name is required")
		return
	}

	id := c.Param("id")
	if id == "" {
		common.Fail(c, "record id is required")
		return
	}

	if err := api.service.DeleteRecord(tableName, id); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "record deleted successfully")
}
