package middleware

import (
	"fmt"
	"k-admin-system/global"
	"k-admin-system/model/common"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 恢复中间件
// 捕获panic，记录堆栈跟踪，返回500错误，并保持服务器运行
//
// 使用示例:
//
//	router.Use(middleware.Recovery())
//
// 当发生panic时，日志格式:
//
//	{
//	  "timestamp": "2024-01-01T12:00:00Z",
//	  "level": "error",
//	  "msg": "Panic recovered",
//	  "error": "runtime error: invalid memory address or nil pointer dereference",
//	  "path": "/api/v1/users",
//	  "method": "GET",
//	  "stack": "goroutine 1 [running]:\n..."
//	}
//
// 响应格式:
//
//	{
//	  "code": 500,
//	  "data": null,
//	  "msg": "Internal server error"
//	}
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈跟踪
				stack := string(debug.Stack())

				// 获取请求信息
				path := c.Request.URL.Path
				method := c.Request.Method

				// 记录panic日志
				if global.Logger != nil {
					global.Logger.Error("Panic recovered",
						zap.Any("error", err),
						zap.String("path", path),
						zap.String("method", method),
						zap.String("stack", stack),
					)
				}

				// 返回500错误响应
				c.JSON(http.StatusOK, common.Response{
					Code: 500,
					Data: nil,
					Msg:  fmt.Sprintf("Internal server error: %v", err),
				})

				// 中止请求处理
				c.Abort()
			}
		}()

		// 处理请求
		c.Next()
	}
}
