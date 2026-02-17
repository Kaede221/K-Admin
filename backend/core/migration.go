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
		global.Logger.Info("Roles already exist, checking menu associations and Casbin policies...")

		// 检查管理员角色的菜单关联
		var adminRole system.SysRole
		if err := global.DB.Where("role_key = ?", "admin").First(&adminRole).Error; err != nil {
			global.Logger.Error("Failed to find admin role", zap.Error(err))
			return err
		}

		// 检查并修复菜单关联
		var totalMenuCount int64
		if err := global.DB.Model(&system.SysMenu{}).Count(&totalMenuCount).Error; err != nil {
			global.Logger.Error("Failed to count total menus", zap.Error(err))
			return err
		}
		global.Logger.Info("Total menus in database", zap.Int64("count", totalMenuCount))

		if totalMenuCount == 0 {
			global.Logger.Warn("No menus in database, creating default menus...")
			if err := createDefaultMenus(&adminRole); err != nil {
				return err
			}
		} else {
			menuCount := global.DB.Model(&adminRole).Association("Menus").Count()
			if menuCount == 0 {
				global.Logger.Warn("Admin role has no menu associations, fixing...")
				var allMenus []system.SysMenu
				if err := global.DB.Find(&allMenus).Error; err != nil {
					global.Logger.Error("Failed to find menus", zap.Error(err))
					return err
				}
				if err := global.DB.Model(&adminRole).Association("Menus").Append(allMenus); err != nil {
					global.Logger.Error("Failed to associate menus with admin role", zap.Error(err))
					return err
				}
				global.Logger.Info("Fixed menu associations for admin role", zap.Int("menuCount", len(allMenus)))
			} else {
				global.Logger.Info("Admin role already has menu associations", zap.Int64("menuCount", menuCount))
			}
		}

		// 检查并添加 Casbin 策略
		if err := ensureAdminCasbinPolicies(); err != nil {
			return err
		}

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
	if err := createDefaultMenus(adminRole); err != nil {
		return err
	}

	// 添加 admin 角色的 Casbin 策略
	if err := ensureAdminCasbinPolicies(); err != nil {
		return err
	}

	global.Logger.Info("Initial data created successfully")
	return nil
}

// createDefaultMenus 创建默认菜单并关联到角色
func createDefaultMenus(adminRole *system.SysRole) error {
	// 创建默认菜单
	menus := []system.SysMenu{
		// 仪表盘
		{
			ParentID:  0,
			Path:      "/dashboard",
			Name:      "Dashboard",
			Component: "dashboard",
			Sort:      1,
			Meta: system.MenuMeta{
				Icon:      "HomeOutlined",
				Title:     "仪表盘",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{},
		},
		// 系统管理
		{
			ParentID:  0,
			Path:      "/system",
			Name:      "System",
			Component: "Layout",
			Sort:      2,
			Meta: system.MenuMeta{
				Icon:      "SettingOutlined",
				Title:     "系统管理",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{},
		},
		// 工具箱
		{
			ParentID:  0,
			Path:      "/tools",
			Name:      "Tools",
			Component: "Layout",
			Sort:      3,
			Meta: system.MenuMeta{
				Icon:      "ToolOutlined",
				Title:     "工具箱",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{},
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
			Meta: system.MenuMeta{
				Icon:      "UserOutlined",
				Title:     "用户管理",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{"user:create", "user:update", "user:delete"},
		},
		{
			ParentID:  systemMenu.ID,
			Path:      "/system/role",
			Name:      "Role",
			Component: "system/role",
			Sort:      2,
			Meta: system.MenuMeta{
				Icon:      "SafetyOutlined",
				Title:     "角色管理",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{"role:create", "role:update", "role:delete"},
		},
		{
			ParentID:  systemMenu.ID,
			Path:      "/system/menu",
			Name:      "Menu",
			Component: "system/menu",
			Sort:      3,
			Meta: system.MenuMeta{
				Icon:      "MenuOutlined",
				Title:     "菜单管理",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{"menu:create", "menu:update", "menu:delete"},
		},
		// 工具箱子菜单
		{
			ParentID:  toolsMenu.ID,
			Path:      "/tools/code-generator",
			Name:      "CodeGenerator",
			Component: "tools/code-generator",
			Sort:      1,
			Meta: system.MenuMeta{
				Icon:      "CodeOutlined",
				Title:     "代码生成器",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{"code:generate"},
		},
		{
			ParentID:  toolsMenu.ID,
			Path:      "/tools/db-inspector",
			Name:      "DbInspector",
			Component: "tools/db-inspector",
			Sort:      2,
			Meta: system.MenuMeta{
				Icon:      "DatabaseOutlined",
				Title:     "数据库检查器",
				Hidden:    false,
				KeepAlive: true,
			},
			BtnPerms: []string{"db:inspect"},
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

	return nil
}

// ensureAdminCasbinPolicies 确保 admin 角色拥有所有 API 访问权限
func ensureAdminCasbinPolicies() error {
	if global.CasbinEnforcer == nil {
		global.Logger.Warn("Casbin enforcer is nil, skipping policy initialization")
		return nil
	}

	// 检查 admin 角色是否已有策略
	policies, err := global.CasbinEnforcer.GetFilteredPolicy(0, "admin")
	if err != nil {
		global.Logger.Error("Failed to get filtered policies", zap.Error(err))
		return err
	}

	if len(policies) > 0 {
		global.Logger.Info("Admin role already has Casbin policies", zap.Int("count", len(policies)))
		return nil
	}

	global.Logger.Info("Adding Casbin policies for admin role...")

	// 为 admin 角色添加所有 API 访问权限
	// 使用通配符 * 表示允许访问所有路径和方法
	adminPolicies := [][]string{
		// 用户管理
		{"admin", "/api/v1/user/list", "GET"},
		{"admin", "/api/v1/user/:id", "GET"},
		{"admin", "/api/v1/user", "POST"},
		{"admin", "/api/v1/user/:id", "PUT"},
		{"admin", "/api/v1/user/:id", "DELETE"},
		{"admin", "/api/v1/user/:id/status", "PUT"},
		{"admin", "/api/v1/user/reset-password", "POST"},

		// 角色管理
		{"admin", "/api/v1/role/list", "GET"},
		{"admin", "/api/v1/role/:id", "GET"},
		{"admin", "/api/v1/role", "POST"},
		{"admin", "/api/v1/role/:id", "PUT"},
		{"admin", "/api/v1/role/:id", "DELETE"},
		{"admin", "/api/v1/role/assign-menus", "POST"},
		{"admin", "/api/v1/role/:id/menus", "GET"},
		{"admin", "/api/v1/role/assign-apis", "POST"},
		{"admin", "/api/v1/role/:id/apis", "GET"},

		// 菜单管理
		{"admin", "/api/v1/menu/tree", "GET"},
		{"admin", "/api/v1/menu/list", "GET"},
		{"admin", "/api/v1/menu/:id", "GET"},
		{"admin", "/api/v1/menu", "POST"},
		{"admin", "/api/v1/menu/:id", "PUT"},
		{"admin", "/api/v1/menu/:id", "DELETE"},

		// 仪表盘
		{"admin", "/api/v1/dashboard/stats", "GET"},

		// 工具箱
		{"admin", "/api/v1/tools/code-generator/tables", "GET"},
		{"admin", "/api/v1/tools/code-generator/generate", "POST"},
		{"admin", "/api/v1/tools/db-inspector/tables", "GET"},
		{"admin", "/api/v1/tools/db-inspector/table/:tableName", "GET"},
	}

	// 批量添加策略
	_, err = global.CasbinEnforcer.AddPolicies(adminPolicies)
	if err != nil {
		global.Logger.Error("Failed to add Casbin policies for admin", zap.Error(err))
		return err
	}

	global.Logger.Info("Casbin policies added for admin role", zap.Int("count", len(adminPolicies)))
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
