package core

import (
	"k-admin-system/global"
	"k-admin-system/model/system"
	"k-admin-system/utils"

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

// InitializeData 初始化默认数据
func InitializeData() error {
	if global.DB == nil {
		global.Logger.Error("Database connection is nil, cannot initialize data")
		return gorm.ErrInvalidDB
	}

	global.Logger.Info("Checking if initial data needs to be created...")

	// 检查是否已有管理员角色
	var roleCount int64
	if err := global.DB.Model(&system.SysRole{}).Count(&roleCount).Error; err != nil {
		global.Logger.Error("Failed to count roles", zap.Error(err))
		return err
	}

	if roleCount > 0 {
		global.Logger.Info("Initial data already exists, skipping initialization")
		return nil
	}

	global.Logger.Info("Creating initial data...")

	// 创建默认管理员角色
	adminRole := &system.SysRole{
		RoleName:  "超级管理员",
		RoleKey:   "admin",
		DataScope: "all",
		Sort:      1,
		Status:    true,
		Remark:    "系统默认超级管理员角色",
	}
	if err := global.DB.Create(adminRole).Error; err != nil {
		global.Logger.Error("Failed to create admin role", zap.Error(err))
		return err
	}
	global.Logger.Info("Admin role created", zap.Uint("roleId", adminRole.ID))

	// 创建默认管理员用户
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		global.Logger.Error("Failed to hash password", zap.Error(err))
		return err
	}

	adminUser := &system.SysUser{
		Username: "admin",
		Password: hashedPassword,
		Nickname: "系统管理员",
		RoleID:   adminRole.ID,
		Active:   true,
	}
	if err := global.DB.Create(adminUser).Error; err != nil {
		global.Logger.Error("Failed to create admin user", zap.Error(err))
		return err
	}
	global.Logger.Info("Admin user created", zap.Uint("userId", adminUser.ID))

	global.Logger.Info("Initial data created successfully")
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

	// 初始化默认数据
	if err := InitializeData(); err != nil {
		global.Logger.Error("Failed to initialize data", zap.Error(err))
		return err
	}

	return nil
}
