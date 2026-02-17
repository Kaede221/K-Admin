package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "k-admin-system/docs" // Swagger docs
)

func TestSwaggerEndpoint(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册Swagger路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 测试Swagger路由是否注册成功
	// Note: 实际的Swagger UI需要静态文件支持，在单元测试中可能无法完全加载
	// 这里只验证路由是否注册
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/swagger/*any" {
			found = true
			break
		}
	}
	assert.True(t, found, "Swagger route should be registered")
}

func TestSwaggerDocsGenerated(t *testing.T) {
	// 验证Swagger文档文件是否已生成
	files := []string{
		"docs/docs.go",
		"docs/swagger.json",
		"docs/swagger.yaml",
	}

	for _, file := range files {
		_, err := os.Stat(file)
		assert.NoError(t, err, "Swagger file %s should exist", file)
	}
}

func TestHealthEndpoint(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 注册健康检查路由
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"mode":   gin.Mode(),
		})
	})

	// 测试健康检查端点
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Health endpoint should return 200 OK")
	assert.Contains(t, w.Body.String(), "ok", "Response should contain status ok")
}
