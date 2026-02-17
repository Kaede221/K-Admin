package middleware

import (
	"k-admin-system/global"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 请求日志中间件
// 记录所有HTTP请求的详细信息，包括时间戳、方法、路径、状态码、延迟和客户端IP
//
// 使用示例:
//
//	router.Use(middleware.Logger())
//
// 日志格式:
//
//	{
//	  "timestamp": "2024-01-01T12:00:00Z",
//	  "level": "info",
//	  "msg": "HTTP Request",
//	  "method": "GET",
//	  "path": "/api/v1/users",
//	  "status": 200,
//	  "latency": "15.234ms",
//	  "client_ip": "192.168.1.1"
//	}
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 获取客户端IP
		clientIP := c.ClientIP()

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 处理请求
		c.Next()

		// 计算请求延迟
		latency := time.Since(startTime)

		// 获取响应状态码
		statusCode := c.Writer.Status()

		// 记录日志
		if global.Logger != nil {
			global.Logger.Info("HTTP Request",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", statusCode),
				zap.Duration("latency", latency),
				zap.String("client_ip", clientIP),
			)
		}
	}
}
