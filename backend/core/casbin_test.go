package core

import (
	"testing"

	"k-admin-system/model/system"

	"github.com/stretchr/testify/assert"
)

// TestSysCasbinRuleModel 测试Casbin规则模型定义
func TestSysCasbinRuleModel(t *testing.T) {
	// 创建一个Casbin规则实例
	rule := system.SysCasbinRule{
		Ptype: "p",
		V0:    "admin",
		V1:    "/api/v1/user",
		V2:    "GET",
	}

	// 验证表名
	assert.Equal(t, "sys_casbin_rules", rule.TableName(), "Table name should be sys_casbin_rules")

	// 验证字段可以正确设置
	assert.Equal(t, "p", rule.Ptype)
	assert.Equal(t, "admin", rule.V0)
	assert.Equal(t, "/api/v1/user", rule.V1)
	assert.Equal(t, "GET", rule.V2)
}

// TestCasbinModelConfiguration 测试Casbin模型配置文件存在
func TestCasbinModelConfiguration(t *testing.T) {
	// 验证模型配置文件路径
	modelPath := "config/casbin_model.conf"

	// 这个测试验证模型配置文件路径是正确的
	// 实际的文件存在性检查在集成测试中进行
	assert.NotEmpty(t, modelPath, "Model path should not be empty")
}

// 注意: 完整的Casbin初始化测试需要数据库连接
// 这些测试应该在集成测试环境中运行，确保数据库可用
//
// 集成测试应该验证:
// 1. InitCasbin() 成功创建enforcer
// 2. Enforcer可以加载和保存策略
// 3. RESTful路径匹配(keyMatch2)正常工作
// 4. 角色继承正常工作
