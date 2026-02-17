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

	// 创建默认菜单
	menus := []system.SysMenu{
		// 仪表盘
		{
			ParentID:  0,
			Path:      "/dashboard",
			Name:      "Dashboard",
			Component: "dashboard",
			Sort:      1,
			Meta:      `{"icon":"HomeIcon","title":"仪表盘","hidden":false,"keep_alive":true}`,
			BtnPerms:  `[]`,
		},
		// 系统管理
		{
			ParentID:  0,
			Path:      "/system",
			Name:      "System",
			Component: "Layout",
			Sort:      2,
			Meta:      `{"icon":"CogIcon","title":"系统管理","hidden":false,"keep_alive":true}`,
			BtnPerms:  `[]`,
		},
		// 工具箱
		{
			ParentID:  0,
			Path:      "/tools",
			Name:      "Tools",
			Component: "Layout",
			Sort:      3,
			Meta:      `{"icon":"WrenchIcon","title":"工具箱","hidden":false,"keep_alive":true}`,
			BtnPerms:  `[]`,
		},
	}

	// 批量创建菜单
	if err := global.DB.Create(&menus).Error; err != nil {
		global.Logger.Error("Failed to create menus", zap.Error(err))
		return err
	}
	global.Logger.Info("Default menus created", zap.Int("count", len(menus)))

	// 获取父菜单ID
	var systemMenu, toolsMenu system.SysMenu
	global.DB.Where("name = ?", "System").First(&systemMenu)
	global.DB.Where("name = ?", "Tools").First(&toolsMenu)

	// 创建子菜单
	subMenus := []system.SysMenu{
		// 系统管理子菜单
		{
			ParentID:  systemMenu.ID,
			Path:      "/system/user",
			Name:      "User",
			Component: "system/user",
			Sort:      1,
			Meta:      `{"icon":"UserIcon","title":"用户管理","hidden":false,"keep_alive":true}`,
			BtnPerms:  `["user:create","user:update","user:delete"]`,
		},
		{
			ParentID:  systemMenu.ID,
			Path:      "/system/role",
			Name:      "Role",
			Component: "system/role",
			Sort:      2,
			Meta:      `{"icon":"ShieldCheckIcon","title":"角色管理","hidden":false,"keep_alive":true}`,
			BtnPerms:  `["role:create","role:update","role:delete"]`,
		},
		{
			ParentID:  systemMenu.ID,
			Path:      "/system/menu",
			Name:      "Menu",
			Component: "system/menu",
			Sort:      3,
			Meta:      `{"icon":"Bars3Icon","title":"菜单管理","hidden":false,"keep_alive":true}`,
			BtnPerms:  `["menu:create","menu:update","menu:delete"]`,
		},
		// 工具箱子菜单
		{
			ParentID:  toolsMenu.ID,
			Path:      "/tools/code-generator",
			Name:      "CodeGenerator",
			Component: "tools/code-generator",
			Sort:      1,
			Meta:      `{"icon":"CodeBracketIcon","title":"代码生成器","hidden":false,"keep_alive":true}`,
			BtnPerms:  `["code:generate"]`,
		},
		{
			ParentID:  toolsMenu.ID,
			Path:      "/tools/db-inspector",
			Name:      "DbInspector",
			Component: "tools/db-inspector",
			Sort:      2,
			Meta:      `{"icon":"CircleStackIcon","title":"数据库检查器","hidden":false,"keep_alive":true}`,
			BtnPerms:  `["db:inspect"]`,
		},
	}

	// 批量创建子菜单
	if err := global.DB.Create(&subMenus).Error; err != nil {
		global.Logger.Error("Failed to create sub menus", zap.Error(err))
		return err
	}
	global.Logger.Info("Default sub menus created", zap.Int("count", len(subMenus)))

	// 将所有菜单关联到管理员角色
	allMenus := append(menus, subMenus...)
	if err := global.DB.Model(adminRole).Association("Menus").Append(allMenus); err != nil {
		global.Logger.Error("Failed to associate menus with admin role", zap.Error(err))
		return err
	}
	global.Logger.Info("Menus associated with admin role", zap.Int("menuCount", len(allMenus)))

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
