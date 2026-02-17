package tools

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"k-admin-system/global"
	"k-admin-system/model/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDBInspectorTest(t *testing.T) *gin.Engine {
	// 初始化测试数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	global.DB = db

	// 创建测试表
	global.DB.Exec(`CREATE TABLE IF NOT EXISTS test_table (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		age INTEGER,
		email TEXT
	)`)

	// 插入测试数据
	global.DB.Exec(`INSERT INTO test_table (name, age, email) VALUES ('Alice', 25, 'alice@example.com')`)
	global.DB.Exec(`INSERT INTO test_table (name, age, email) VALUES ('Bob', 30, 'bob@example.com')`)

	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := &DBInspectorAPI{}

	// 注册路由
	router.GET("/tools/db/tables", api.GetTables)
	router.GET("/tools/db/tables/:tableName/schema", api.GetTableSchema)
	router.GET("/tools/db/tables/:tableName/data", api.GetTableData)
	router.POST("/tools/db/execute", api.ExecuteSQL)
	router.POST("/tools/db/tables/:tableName/records", api.CreateRecord)
	router.PUT("/tools/db/tables/:tableName/records/:id", api.UpdateRecord)
	router.DELETE("/tools/db/tables/:tableName/records/:id", api.DeleteRecord)

	return router
}

func teardownDBInspectorTest() {
	global.DB.Exec(`DROP TABLE IF EXISTS test_table`)
}

func TestGetTables(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tools/db/tables", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证返回的表列表包含test_table
	tables, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, tables)

	// 检查是否包含test_table
	found := false
	for _, table := range tables {
		if table.(string) == "test_table" {
			found = true
			break
		}
	}
	assert.True(t, found, "test_table should be in the table list")
}

func TestGetTableSchema(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tools/db/tables/test_table/schema", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证返回的列信息
	columns, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, columns)

	// 检查列名
	columnNames := make([]string, 0)
	for _, col := range columns {
		colMap := col.(map[string]interface{})
		columnNames = append(columnNames, colMap["name"].(string))
	}

	assert.Contains(t, columnNames, "id")
	assert.Contains(t, columnNames, "name")
	assert.Contains(t, columnNames, "age")
	assert.Contains(t, columnNames, "email")
}

func TestGetTableData(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tools/db/tables/test_table/data?page=1&pageSize=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证返回的数据
	dataMap, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	list, ok := dataMap["list"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(list))

	total, ok := dataMap["total"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(2), total)
}

func TestExecuteSQL_Select(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	reqBody := map[string]interface{}{
		"sql":      "SELECT * FROM test_table WHERE name = 'Alice'",
		"readOnly": false,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tools/db/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证查询结果
	results, ok := response.Data.([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(results))
}

func TestExecuteSQL_DangerousOperation(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	reqBody := map[string]interface{}{
		"sql":      "DROP TABLE test_table",
		"readOnly": false,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tools/db/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, response.Code, "Dangerous operation should be rejected")
	assert.Contains(t, response.Msg, "dangerous operation")
}

func TestExecuteSQL_ReadOnlyMode(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	reqBody := map[string]interface{}{
		"sql":      "INSERT INTO test_table (name, age) VALUES ('Charlie', 35)",
		"readOnly": true,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tools/db/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, response.Code, "Write operation should be rejected in read-only mode")
	assert.Contains(t, response.Msg, "read-only mode")
}

func TestCreateRecord(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	reqBody := map[string]interface{}{
		"name":  "Charlie",
		"age":   35,
		"email": "charlie@example.com",
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tools/db/tables/test_table/records", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证记录已创建
	var count int64
	global.DB.Raw("SELECT COUNT(*) FROM test_table WHERE name = 'Charlie'").Scan(&count)
	assert.Equal(t, int64(1), count)
}

func TestUpdateRecord(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	reqBody := map[string]interface{}{
		"age": 26,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tools/db/tables/test_table/records/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证记录已更新
	var age int
	global.DB.Raw("SELECT age FROM test_table WHERE id = 1").Scan(&age)
	assert.Equal(t, 26, age)
}

func TestDeleteRecord(t *testing.T) {
	router := setupDBInspectorTest(t)
	defer teardownDBInspectorTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tools/db/tables/test_table/records/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response common.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 0, response.Code)

	// 验证记录已删除
	var count int64
	global.DB.Raw("SELECT COUNT(*) FROM test_table WHERE id = 1").Scan(&count)
	assert.Equal(t, int64(0), count)
}
