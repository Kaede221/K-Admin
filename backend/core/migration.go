package core

import (
	"k-admin-system/global"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RegisterTables 注册需要自动迁移的表
func RegisterTables(db *gorm.DB) error {
	// 在这里注册所有需要自动迁移的模型
	// 例如: db.AutoMigrate(&model.SysUser{}, &model.SysRole{}, &model.SysMenu{})

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
