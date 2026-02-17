package middleware

import (
	"k-admin-system/global"
	"k-admin-system/model/common"
	"k-admin-system/model/system"

	"github.com/gin-gonic/gin"
)

// CasbinAuth Casbin授权中间件
// 从JWT claims中提取角色信息，使用Casbin enforcer检查API访问权限
func CasbinAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取roleId（由JWT中间件设置）
		roleIdInterface, exists := c.Get("roleId")
		if !exists {
			common.FailWithCode(c, 401, "未找到角色信息")
			c.Abort()
			return
		}

		roleId, ok := roleIdInterface.(uint)
		if !ok {
			common.FailWithCode(c, 500, "角色信息格式错误")
			c.Abort()
			return
		}

		// 从数据库查询角色的role_key
		var role system.SysRole
		if err := global.DB.First(&role, roleId).Error; err != nil {
			global.Logger.Error("Failed to query role: " + err.Error())
			common.FailWithCode(c, 403, "角色不存在")
			c.Abort()
			return
		}

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 使用Casbin enforcer检查权限
		allowed, err := global.CasbinEnforcer.Enforce(role.RoleKey, path, method)
		if err != nil {
			global.Logger.Error("Casbin enforce error: " + err.Error())
			common.FailWithCode(c, 500, "权限检查失败")
			c.Abort()
			return
		}

		if !allowed {
			global.Logger.Warn("Access denied for role: " + role.RoleKey + " path: " + path + " method: " + method)
			common.FailWithCode(c, 403, "无权访问")
			c.Abort()
			return
		}

		// 权限检查通过，继续处理请求
		c.Next()
	}
}
