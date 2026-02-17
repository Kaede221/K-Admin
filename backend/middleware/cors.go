package middleware

import (
	"k-admin-system/config"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS CORS中间件
// 处理跨域请求，设置适当的Access-Control响应头
//
// 使用示例:
//
//	router.Use(middleware.CORS(global.Config.CORS))
//
// 配置示例 (config.yaml):
//
//	cors:
//	  allow_origins:
//	    - "http://localhost:3000"
//	    - "https://yourdomain.com"
//	  allow_methods:
//	    - "GET"
//	    - "POST"
//	    - "PUT"
//	    - "DELETE"
//	  allow_headers:
//	    - "Content-Type"
//	    - "Authorization"
//	  allow_credentials: true
//	  max_age: 86400
func CORS(corsConfig config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查origin是否在允许列表中
		if origin != "" && isOriginAllowed(origin, corsConfig.AllowOrigins) {
			// 设置允许的源
			c.Header("Access-Control-Allow-Origin", origin)

			// 设置允许的方法
			if len(corsConfig.AllowMethods) > 0 {
				c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowMethods, ", "))
			}

			// 设置允许的请求头
			if len(corsConfig.AllowHeaders) > 0 {
				c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowHeaders, ", "))
			}

			// 设置暴露的响应头
			if len(corsConfig.ExposeHeaders) > 0 {
				c.Header("Access-Control-Expose-Headers", strings.Join(corsConfig.ExposeHeaders, ", "))
			}

			// 设置是否允许携带凭证
			if corsConfig.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			// 设置预检请求的缓存时间
			if corsConfig.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))
			}
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed 检查origin是否在允许列表中
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		// 支持通配符 *
		if allowed == "*" {
			return true
		}
		// 精确匹配
		if allowed == origin {
			return true
		}
		// 支持通配符子域名匹配，例如 *.example.com
		if strings.HasPrefix(allowed, "*.") {
			domain := allowed[2:] // 去掉 *.
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}
