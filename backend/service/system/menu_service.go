package system

import (
	"errors"
	"fmt"

	"k-admin-system/global"
	"k-admin-system/model/system"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MenuService 菜单服务
type MenuService struct{}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(menu *system.SysMenu) error {
	// 如果有父菜单，检查父菜单是否存在
	if menu.ParentID > 0 {
		var parent system.SysMenu
		if err := global.DB.First(&parent, menu.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("parent menu not found")
			}
			return fmt.Errorf("failed to query parent menu: %w", err)
		}
	}

	// 创建菜单
	if err := global.DB.Create(menu).Error; err != nil {
		return fmt.Errorf("failed to create menu: %w", err)
	}

	return nil
}

// UpdateMenu 更新菜单信息
func (s *MenuService) UpdateMenu(menu *system.SysMenu) error {
	// 检查菜单是否存在
	var existingMenu system.SysMenu
	if err := global.DB.First(&existingMenu, menu.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("menu not found")
		}
		return fmt.Errorf("failed to query menu: %w", err)
	}

	// 如果更新父菜单，检查父菜单是否存在，且不能设置自己为父菜单
	if menu.ParentID > 0 {
		if menu.ParentID == menu.ID {
			return errors.New("cannot set self as parent menu")
		}
		var parent system.SysMenu
		if err := global.DB.First(&parent, menu.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("parent menu not found")
			}
			return fmt.Errorf("failed to query parent menu: %w", err)
		}
	}

	// 更新菜单
	if err := global.DB.Save(menu).Error; err != nil {
		return fmt.Errorf("failed to update menu: %w", err)
	}

	return nil
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(id uint) error {
	// 检查菜单是否存在
	var menu system.SysMenu
	if err := global.DB.First(&menu, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("menu not found")
		}
		return fmt.Errorf("failed to query menu: %w", err)
	}

	// 检查是否有子菜单
	var childCount int64
	if err := global.DB.Model(&system.SysMenu{}).Where("parent_id = ?", id).Count(&childCount).Error; err != nil {
		return fmt.Errorf("failed to check child menus: %w", err)
	}
	if childCount > 0 {
		return errors.New("cannot delete menu with child menus")
	}

	// 删除菜单
	if err := global.DB.Delete(&menu).Error; err != nil {
		return fmt.Errorf("failed to delete menu: %w", err)
	}

	return nil
}

// GetMenuByID 根据ID获取菜单
func (s *MenuService) GetMenuByID(id uint) (*system.SysMenu, error) {
	var menu system.SysMenu
	if err := global.DB.First(&menu, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("menu not found")
		}
		return nil, fmt.Errorf("failed to query menu: %w", err)
	}

	return &menu, nil
}

// GetAllMenus 获取所有菜单（不构建树结构）
func (s *MenuService) GetAllMenus() ([]system.SysMenu, error) {
	var menus []system.SysMenu
	if err := global.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("failed to query menus: %w", err)
	}

	return menus, nil
}

// GetMenuTree 获取菜单树（根据角色过滤）
// 如果 roleID 为 0，返回所有菜单
func (s *MenuService) GetMenuTree(roleID uint) ([]system.SysMenu, error) {
	var menus []system.SysMenu

	global.Logger.Info("GetMenuTree called",
		zap.Uint("roleID", roleID))

	if roleID == 0 {
		// 获取所有菜单
		if err := global.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
			return nil, fmt.Errorf("failed to query menus: %w", err)
		}
		global.Logger.Info("Fetched all menus",
			zap.Int("count", len(menus)))
	} else {
		// 根据角色获取菜单
		var role system.SysRole
		if err := global.DB.Preload("Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort ASC, id ASC")
		}).First(&role, roleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				global.Logger.Error("Role not found", zap.Uint("roleID", roleID))
				return nil, errors.New("role not found")
			}
			global.Logger.Error("Failed to query role",
				zap.Uint("roleID", roleID),
				zap.Error(err))
			return nil, fmt.Errorf("failed to query role menus: %w", err)
		}
		menus = role.Menus
		global.Logger.Info("Fetched role menus",
			zap.Uint("roleID", roleID),
			zap.String("roleName", role.RoleName),
			zap.Int("menuCount", len(menus)))
	}

	// 构建树结构
	tree := s.BuildMenuTree(menus, 0)
	global.Logger.Info("Built menu tree",
		zap.Int("treeNodeCount", len(tree)))
	return tree, nil
}

// BuildMenuTree 构建菜单树（递归）
// parentID 为 0 表示根节点
func (s *MenuService) BuildMenuTree(menus []system.SysMenu, parentID uint) []system.SysMenu {
	tree := make([]system.SysMenu, 0) // 初始化为空数组而不是 nil

	for _, menu := range menus {
		if menu.ParentID == parentID {
			// 递归查找子菜单
			children := s.BuildMenuTree(menus, menu.ID)
			if len(children) > 0 {
				menu.Children = children
			}
			tree = append(tree, menu)
		}
	}

	return tree
}

// GetMenusByRoleIDs 根据多个角色ID获取菜单树（用于用户有多个角色的情况）
func (s *MenuService) GetMenusByRoleIDs(roleIDs []uint) ([]system.SysMenu, error) {
	if len(roleIDs) == 0 {
		return make([]system.SysMenu, 0), nil // 返回空数组而不是 nil
	}

	// 查询所有角色的菜单（去重）
	var menus []system.SysMenu
	if err := global.DB.
		Distinct().
		Joins("JOIN sys_role_menus ON sys_role_menus.sys_menu_id = sys_menus.id").
		Where("sys_role_menus.sys_role_id IN ?", roleIDs).
		Order("sort ASC, id ASC").
		Find(&menus).Error; err != nil {
		return nil, fmt.Errorf("failed to query menus by role IDs: %w", err)
	}

	// 构建树结构
	tree := s.BuildMenuTree(menus, 0)
	return tree, nil
}
