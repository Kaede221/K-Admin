package middleware

import (
	"k-admin-system/model/common"
	"k-admin-system/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			common.FailWithCode(c, 401, "未提供认证令牌")
			c.Abort()
			return
		}

		// 验证Bearer格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			common.FailWithCode(c, 401, "认证令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			switch err {
			case utils.ErrTokenExpired:
				common.FailWithCode(c, 401, "令牌已过期")
			case utils.ErrTokenBlacklisted:
				common.FailWithCode(c, 401, "令牌已失效")
			default:
				common.FailWithCode(c, 401, "令牌无效")
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleId", claims.RoleID)

		c.Next()
	}
}
