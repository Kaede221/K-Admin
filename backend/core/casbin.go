package core

import (
	"k-admin-system/global"
	"k-admin-system/model/system"

	"github.com/casbin/casbin/v3"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"go.uber.org/zap"
)

// InitCasbin 初始化Casbin enforcer
// 使用Gorm adapter连接数据库，加载RBAC模型配置
func InitCasbin() (*casbin.Enforcer, error) {
	// 创建Gorm adapter，使用sys_casbin_rules表
	adapter, err := gormadapter.NewAdapterByDBWithCustomTable(
		global.DB,
		&system.SysCasbinRule{},
		"sys_casbin_rules",
	)
	if err != nil {
		global.Logger.Error("Failed to create Casbin adapter", zap.Error(err))
		return nil, err
	}

	// 加载Casbin模型配置文件
	enforcer, err := casbin.NewEnforcer("config/casbin_model.conf", adapter)
	if err != nil {
		global.Logger.Error("Failed to create Casbin enforcer", zap.Error(err))
		return nil, err
	}

	// 从数据库加载策略
	err = enforcer.LoadPolicy()
	if err != nil {
		global.Logger.Error("Failed to load Casbin policies", zap.Error(err))
		return nil, err
	}

	global.Logger.Info("Casbin enforcer initialized successfully")
	return enforcer, nil
}
