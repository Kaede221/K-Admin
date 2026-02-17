package system

import (
	"fmt"

	"k-admin-system/global"
	"k-admin-system/model/system"
)

// DashboardService 仪表盘服务
type DashboardService struct{}

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	UserCount   int64 `json:"userCount"`
	RoleCount   int64 `json:"roleCount"`
	MenuCount   int64 `json:"menuCount"`
	ConfigCount int64 `json:"configCount"`
}

// GetDashboardStats 获取仪表盘统计数据
func (s *DashboardService) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}

	// 统计用户数量
	if err := global.DB.Model(&system.SysUser{}).Count(&stats.UserCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// 统计角色数量
	if err := global.DB.Model(&system.SysRole{}).Count(&stats.RoleCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count roles: %w", err)
	}

	// 统计菜单数量
	if err := global.DB.Model(&system.SysMenu{}).Count(&stats.MenuCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count menus: %w", err)
	}

	// 系统配置数量（这里暂时使用固定值，后续可以根据实际配置表统计）
	stats.ConfigCount = 15

	return stats, nil
}
