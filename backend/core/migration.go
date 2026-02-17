package core

import (
	"k-admin-system/global"
	"k-admin-system/model/system"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterTables 注册需要自动迁移的表
func RegisterTables(db *gorm.DB) error {
	// 注册系统模型 - 注意顺序：先创建被引用的表，再创建引用它们的表
	err := db.AutoMigrate(
		&system.SysRole{},       // 先创建角色表
		&system.SysMenu{},       // 再创建菜单表
		&system.SysUser{},       // 最后创建用户表（依赖角色表）
		&system.SysCasbinRule{}, // Casbin 规则表
	)
	if err != nil {
		global.Logger.Error("Failed to migrate tables", zap.Error(err))
		return err
	}

	global.Logger.Info("Database tables registered for migration")
	return nil
}

// AutoMigrate 执行数据库自动迁移
func AutoMigrate() error {
	if global.DB == nil {
		global.Logger.Error("Database connection is nil, cannot perform migration")
		return gorm.ErrInvalidDB
	}

	global.Logger.Info("Starting database migration...")

	err := RegisterTables(global.DB)
	if err != nil {
		global.Logger.Error("Database migration failed", zap.Error(err))
		return err
	}

	global.Logger.Info("Database migration completed successfully")
	return nil
}
