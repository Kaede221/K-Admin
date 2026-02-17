package system

import (
	"errors"
	"fmt"

	"k-admin-system/global"
	"k-admin-system/model/system"

	"gorm.io/gorm"
)

// RoleService 角色服务
type RoleService struct{}

// CreateRole 创建角色
func (s *RoleService) CreateRole(role *system.SysRole) error {
	// 检查角色键是否已存在
	var count int64
	if err := global.DB.Model(&system.SysRole{}).Where("role_key = ?", role.RoleKey).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check role key uniqueness: %w", err)
	}
	if count > 0 {
		return errors.New("role key already exists")
	}

	// 创建角色
	if err := global.DB.Create(role).Error; err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// UpdateRole 更新角色信息
func (s *RoleService) UpdateRole(role *system.SysRole) error {
	// 检查角色是否存在
	var existingRole system.SysRole
	if err := global.DB.First(&existingRole, role.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to query role: %w", err)
	}

	// 如果更新角色键，检查新角色键是否已被其他角色使用
	if role.RoleKey != existingRole.RoleKey {
		var count int64
		if err := global.DB.Model(&system.SysRole{}).
			Where("role_key = ? AND id != ?", role.RoleKey, role.ID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check role key uniqueness: %w", err)
		}
		if count > 0 {
			return errors.New("role key already exists")
		}
	}

	// 更新角色
	if err := global.DB.Save(role).Error; err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	return nil
}

// DeleteRole 删除角色
func (s *RoleService) DeleteRole(id uint) error {
	// 检查角色是否存在
	var role system.SysRole
	if err := global.DB.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to query role: %w", err)
	}

	// 检查是否有用户关联此角色
	var userCount int64
	if err := global.DB.Model(&system.SysUser{}).Where("role_id = ?", id).Count(&userCount).Error; err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}
	if userCount > 0 {
		return errors.New("cannot delete role with associated users")
	}

	// 删除角色
	if err := global.DB.Delete(&role).Error; err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// GetRoleByID 根据ID获取角色
func (s *RoleService) GetRoleByID(id uint) (*system.SysRole, error) {
	var role system.SysRole
	if err := global.DB.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	return &role, nil
}

// GetRoleList 获取角色列表（支持分页）
func (s *RoleService) GetRoleList(page, pageSize int) ([]system.SysRole, int64, error) {
	var roles []system.SysRole
	var total int64

	// 获取总数
	if err := global.DB.Model(&system.SysRole{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := global.DB.Offset(offset).Limit(pageSize).Order("sort ASC, id DESC").Find(&roles).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query roles: %w", err)
	}

	return roles, total, nil
}

// AssignMenus 为角色分配菜单权限
func (s *RoleService) AssignMenus(roleID uint, menuIDs []uint) error {
	// 检查角色是否存在
	var role system.SysRole
	if err := global.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to query role: %w", err)
	}

	// 查询菜单
	var menus []system.SysMenu
	if len(menuIDs) > 0 {
		if err := global.DB.Where("id IN ?", menuIDs).Find(&menus).Error; err != nil {
			return fmt.Errorf("failed to query menus: %w", err)
		}
	}

	// 使用事务更新角色菜单关联
	err := global.DB.Transaction(func(tx *gorm.DB) error {
		// 清除现有关联
		if err := tx.Model(&role).Association("Menus").Clear(); err != nil {
			return fmt.Errorf("failed to clear existing menu associations: %w", err)
		}

		// 添加新关联
		if len(menus) > 0 {
			if err := tx.Model(&role).Association("Menus").Append(&menus); err != nil {
				return fmt.Errorf("failed to assign menus: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetRoleMenus 获取角色的菜单权限
func (s *RoleService) GetRoleMenus(roleID uint) ([]uint, error) {
	// 检查角色是否存在
	var role system.SysRole
	if err := global.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	// 查询角色关联的菜单
	var menus []system.SysMenu
	if err := global.DB.Model(&role).Association("Menus").Find(&menus); err != nil {
		return nil, fmt.Errorf("failed to query role menus: %w", err)
	}

	// 提取菜单ID
	menuIDs := make([]uint, len(menus))
	for i, menu := range menus {
		menuIDs[i] = menu.ID
	}

	return menuIDs, nil
}

// AssignAPIs 为角色分配API权限（通过Casbin策略）
// policies 格式: [][]string{{"path", "method"}, ...}
func (s *RoleService) AssignAPIs(roleID uint, policies [][]string) error {
	// 检查角色是否存在
	var role system.SysRole
	if err := global.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role not found")
		}
		return fmt.Errorf("failed to query role: %w", err)
	}

	// TODO: 实现Casbin策略更新
	// 这将在Task 8中实现Casbin manager后完成
	// 目前返回未实现错误
	return errors.New("API permission assignment not yet implemented - requires Casbin manager")
}

// GetRoleAPIs 获取角色的API权限
func (s *RoleService) GetRoleAPIs(roleID uint) ([][]string, error) {
	// 检查角色是否存在
	var role system.SysRole
	if err := global.DB.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, fmt.Errorf("failed to query role: %w", err)
	}

	// TODO: 实现Casbin策略查询
	// 这将在Task 8中实现Casbin manager后完成
	// 目前返回空列表
	return [][]string{}, nil
}
