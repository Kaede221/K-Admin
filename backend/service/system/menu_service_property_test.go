package system

import (
	"encoding/json"
	"testing"

	"k-admin-system/global"
	"k-admin-system/model/system"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupMenuTestDB 设置测试数据库
func setupMenuTestDB(t *testing.T) {
	var err error
	global.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// 自动迁移
	err = global.DB.AutoMigrate(&system.SysMenu{}, &system.SysRole{}, &system.SysUser{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
}

// TestProperty5_MenuTreeAuthorizationFiltering 测试菜单树授权过滤
// **Property 5: Menu Tree Authorization Filtering**
// **Validates: Requirements 3.2**
//
// GIVEN a set of menus and roles with different permissions
// WHEN GetMenuTree is called with a specific role ID
// THEN only menus assigned to that role should be returned
func TestProperty5_MenuTreeAuthorizationFiltering(t *testing.T) {
	setupMenuTestDB(t)

	service := MenuService{}

	// 创建测试菜单
	menu1 := &system.SysMenu{Name: "Dashboard", Path: "/dashboard", Sort: 1}
	menu2 := &system.SysMenu{Name: "Users", Path: "/users", Sort: 2}
	menu3 := &system.SysMenu{Name: "Settings", Path: "/settings", Sort: 3}
	global.DB.Create(menu1)
	global.DB.Create(menu2)
	global.DB.Create(menu3)

	// 创建角色并分配不同的菜单权限
	role1 := &system.SysRole{RoleName: "Admin", RoleKey: "admin"}
	role2 := &system.SysRole{RoleName: "User", RoleKey: "user"}
	global.DB.Create(role1)
	global.DB.Create(role2)

	// Admin 角色有所有菜单权限
	global.DB.Model(role1).Association("Menus").Append([]system.SysMenu{*menu1, *menu2, *menu3})

	// User 角色只有 Dashboard 和 Users 权限
	global.DB.Model(role2).Association("Menus").Append([]system.SysMenu{*menu1, *menu2})

	// 测试 Admin 角色
	adminMenus, err := service.GetMenuTree(role1.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(adminMenus), "Admin should have access to all 3 menus")

	// 测试 User 角色
	userMenus, err := service.GetMenuTree(role2.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(userMenus), "User should have access to only 2 menus")

	// 验证 User 角色只能看到 Dashboard 和 Users
	menuNames := make(map[string]bool)
	for _, menu := range userMenus {
		menuNames[menu.Name] = true
	}
	assert.True(t, menuNames["Dashboard"], "User should have Dashboard menu")
	assert.True(t, menuNames["Users"], "User should have Users menu")
	assert.False(t, menuNames["Settings"], "User should not have Settings menu")
}

// TestProperty15_MenuHierarchyPreservation 测试菜单层级保持
// **Property 15: Menu Hierarchy Preservation**
// **Validates: Requirements 6.2**
//
// GIVEN a hierarchical menu structure with parent-child relationships
// WHEN BuildMenuTree is called
// THEN the returned tree should preserve the exact parent-child relationships and nesting levels
func TestProperty15_MenuHierarchyPreservation(t *testing.T) {
	setupMenuTestDB(t)

	service := MenuService{}

	// 创建层级菜单结构
	// System (parent)
	//   - Users (child)
	//     - User List (grandchild)
	//     - User Roles (grandchild)
	//   - Settings (child)
	// Dashboard (parent)

	menuSystem := &system.SysMenu{Name: "System", Path: "/system", Sort: 1, ParentID: 0}
	global.DB.Create(menuSystem)

	menuUsers := &system.SysMenu{Name: "Users", Path: "/system/users", Sort: 1, ParentID: menuSystem.ID}
	menuSettings := &system.SysMenu{Name: "Settings", Path: "/system/settings", Sort: 2, ParentID: menuSystem.ID}
	global.DB.Create(menuUsers)
	global.DB.Create(menuSettings)

	menuUserList := &system.SysMenu{Name: "User List", Path: "/system/users/list", Sort: 1, ParentID: menuUsers.ID}
	menuUserRoles := &system.SysMenu{Name: "User Roles", Path: "/system/users/roles", Sort: 2, ParentID: menuUsers.ID}
	global.DB.Create(menuUserList)
	global.DB.Create(menuUserRoles)

	menuDashboard := &system.SysMenu{Name: "Dashboard", Path: "/dashboard", Sort: 2, ParentID: 0}
	global.DB.Create(menuDashboard)

	// 获取所有菜单并构建树
	allMenus, err := service.GetAllMenus()
	assert.NoError(t, err)

	tree := service.BuildMenuTree(allMenus, 0)

	// 验证根节点数量
	assert.Equal(t, 2, len(tree), "Should have 2 root menus")

	// 验证 System 菜单的子菜单
	var systemMenu *system.SysMenu
	for i := range tree {
		if tree[i].Name == "System" {
			systemMenu = &tree[i]
			break
		}
	}
	assert.NotNil(t, systemMenu, "System menu should exist")
	assert.Equal(t, 2, len(systemMenu.Children), "System should have 2 children")

	// 验证 Users 菜单的子菜单
	var usersMenu *system.SysMenu
	for i := range systemMenu.Children {
		if systemMenu.Children[i].Name == "Users" {
			usersMenu = &systemMenu.Children[i]
			break
		}
	}
	assert.NotNil(t, usersMenu, "Users menu should exist")
	assert.Equal(t, 2, len(usersMenu.Children), "Users should have 2 children")

	// 验证孙子菜单
	grandchildNames := make(map[string]bool)
	for _, child := range usersMenu.Children {
		grandchildNames[child.Name] = true
	}
	assert.True(t, grandchildNames["User List"], "User List should be a grandchild")
	assert.True(t, grandchildNames["User Roles"], "User Roles should be a grandchild")
}

// TestProperty16_MenuMetadataSerializationRoundTrip 测试菜单元数据序列化往返
// **Property 16: Menu Metadata Serialization Round-Trip**
// **Validates: Requirements 6.3**
//
// GIVEN a menu with metadata (icon, title, hidden, keepAlive) stored as JSON
// WHEN the menu is saved to database and then retrieved
// THEN the metadata should be identical to the original (round-trip consistency)
func TestProperty16_MenuMetadataSerializationRoundTrip(t *testing.T) {
	setupMenuTestDB(t)

	service := MenuService{}

	// 创建包含元数据的菜单
	metadata := map[string]interface{}{
		"icon":      "user-icon",
		"title":     "User Management",
		"hidden":    false,
		"keepAlive": true,
	}
	metaJSON, err := json.Marshal(metadata)
	assert.NoError(t, err)

	menu := &system.SysMenu{
		Name:      "Users",
		Path:      "/users",
		Component: "views/users/index",
		Sort:      1,
		Meta:      string(metaJSON),
	}

	// 保存菜单
	err = service.CreateMenu(menu)
	assert.NoError(t, err)

	// 检索菜单
	retrievedMenu, err := service.GetMenuByID(menu.ID)
	assert.NoError(t, err)

	// 验证元数据一致性
	assert.Equal(t, menu.Meta, retrievedMenu.Meta, "Metadata should be identical after round-trip")

	// 解析并验证 JSON 内容
	var retrievedMetadata map[string]interface{}
	err = json.Unmarshal([]byte(retrievedMenu.Meta), &retrievedMetadata)
	assert.NoError(t, err)

	assert.Equal(t, "user-icon", retrievedMetadata["icon"])
	assert.Equal(t, "User Management", retrievedMetadata["title"])
	assert.Equal(t, false, retrievedMetadata["hidden"])
	assert.Equal(t, true, retrievedMetadata["keepAlive"])
}

// TestProperty17_HiddenMenuRouteAccessibility 测试隐藏菜单路由可访问性
// **Property 17: Hidden Menu Route Accessibility**
// **Validates: Requirements 6.8**
//
// GIVEN a menu with hidden=true in metadata
// WHEN the menu tree is retrieved
// THEN the hidden menu should still be included in the tree (hidden only affects UI display, not route accessibility)
func TestProperty17_HiddenMenuRouteAccessibility(t *testing.T) {
	setupMenuTestDB(t)

	service := MenuService{}

	// 创建可见菜单
	visibleMetadata := map[string]interface{}{
		"title":  "Visible Menu",
		"hidden": false,
	}
	visibleMetaJSON, _ := json.Marshal(visibleMetadata)
	visibleMenu := &system.SysMenu{
		Name: "Visible",
		Path: "/visible",
		Meta: string(visibleMetaJSON),
		Sort: 1,
	}
	global.DB.Create(visibleMenu)

	// 创建隐藏菜单
	hiddenMetadata := map[string]interface{}{
		"title":  "Hidden Menu",
		"hidden": true,
	}
	hiddenMetaJSON, _ := json.Marshal(hiddenMetadata)
	hiddenMenu := &system.SysMenu{
		Name: "Hidden",
		Path: "/hidden",
		Meta: string(hiddenMetaJSON),
		Sort: 2,
	}
	global.DB.Create(hiddenMenu)

	// 创建角色并分配所有菜单
	role := &system.SysRole{RoleName: "Admin", RoleKey: "admin"}
	global.DB.Create(role)
	global.DB.Model(role).Association("Menus").Append([]system.SysMenu{*visibleMenu, *hiddenMenu})

	// 获取菜单树
	menus, err := service.GetMenuTree(role.ID)
	assert.NoError(t, err)

	// 验证隐藏菜单仍然在树中
	assert.Equal(t, 2, len(menus), "Both visible and hidden menus should be in the tree")

	// 验证隐藏菜单的元数据
	var foundHidden bool
	for _, menu := range menus {
		if menu.Name == "Hidden" {
			foundHidden = true
			var metadata map[string]interface{}
			err := json.Unmarshal([]byte(menu.Meta), &metadata)
			assert.NoError(t, err)
			assert.Equal(t, true, metadata["hidden"], "Hidden flag should be preserved")
		}
	}
	assert.True(t, foundHidden, "Hidden menu should be accessible in the tree")
}

// TestMenuServiceCRUD 测试菜单 CRUD 基本操作
func TestMenuServiceCRUD(t *testing.T) {
	setupMenuTestDB(t)

	service := MenuService{}

	// 测试创建菜单
	menu := &system.SysMenu{
		Name:      "Test Menu",
		Path:      "/test",
		Component: "views/test/index",
		Sort:      1,
	}
	err := service.CreateMenu(menu)
	assert.NoError(t, err)
	assert.NotZero(t, menu.ID)

	// 测试获取菜单
	retrievedMenu, err := service.GetMenuByID(menu.ID)
	assert.NoError(t, err)
	assert.Equal(t, menu.Name, retrievedMenu.Name)

	// 测试更新菜单
	menu.Name = "Updated Menu"
	err = service.UpdateMenu(menu)
	assert.NoError(t, err)

	updatedMenu, err := service.GetMenuByID(menu.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Menu", updatedMenu.Name)

	// 测试删除菜单
	err = service.DeleteMenu(menu.ID)
	assert.NoError(t, err)

	_, err = service.GetMenuByID(menu.ID)
	assert.Error(t, err)
}

// TestMenuServiceParentChildValidation 测试父子菜单验证
func TestMenuServiceParentChildValidation(t *testing.T) {
	setupMenuTestDB(t)

	service := MenuService{}

	// 创建父菜单
	parentMenu := &system.SysMenu{
		Name: "Parent",
		Path: "/parent",
		Sort: 1,
	}
	err := service.CreateMenu(parentMenu)
	assert.NoError(t, err)

	// 创建子菜单
	childMenu := &system.SysMenu{
		Name:     "Child",
		Path:     "/parent/child",
		ParentID: parentMenu.ID,
		Sort:     1,
	}
	err = service.CreateMenu(childMenu)
	assert.NoError(t, err)

	// 测试不能删除有子菜单的父菜单
	err = service.DeleteMenu(parentMenu.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete menu with child menus")

	// 测试不能设置自己为父菜单
	childMenu.ParentID = childMenu.ID
	err = service.UpdateMenu(childMenu)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set self as parent menu")

	// 测试不能设置不存在的父菜单
	childMenu.ParentID = 9999
	err = service.UpdateMenu(childMenu)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent menu not found")
}
