package system

import (
	"errors"
	"fmt"

	"k-admin-system/global"
	"k-admin-system/model/system"
	"k-admin-system/utils"

	"gorm.io/gorm"
)

// UserService 用户服务
type UserService struct{}

// Login 用户登录
// 验证用户凭据并生成访问令牌和刷新令牌
func (s *UserService) Login(username, password string) (accessToken, refreshToken string, user *system.SysUser, err error) {
	// 查询用户
	var dbUser system.SysUser
	if err := global.DB.Where("username = ?", username).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", nil, errors.New("invalid username or password")
		}
		return "", "", nil, fmt.Errorf("failed to query user: %w", err)
	}

	// 检查用户是否激活
	if !dbUser.Active {
		return "", "", nil, errors.New("user account is disabled")
	}

	// 验证密码
	if !utils.CheckPassword(dbUser.Password, password) {
		return "", "", nil, errors.New("invalid username or password")
	}

	// 生成令牌
	accessToken, refreshToken, err = utils.GenerateToken(dbUser.ID, dbUser.Username, dbUser.RoleID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return accessToken, refreshToken, &dbUser, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *system.SysUser) error {
	// 检查用户名是否已存在
	var count int64
	if err := global.DB.Model(&system.SysUser{}).Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check username uniqueness: %w", err)
	}
	if count > 0 {
		return errors.New("username already exists")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	// 创建用户
	if err := global.DB.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *system.SysUser) error {
	// 检查用户是否存在
	var existingUser system.SysUser
	if err := global.DB.First(&existingUser, user.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// 如果更新用户名，检查新用户名是否已被其他用户使用
	if user.Username != existingUser.Username {
		var count int64
		if err := global.DB.Model(&system.SysUser{}).
			Where("username = ? AND id != ?", user.Username, user.ID).
			Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check username uniqueness: %w", err)
		}
		if count > 0 {
			return errors.New("username already exists")
		}
	}

	// 如果提供了新密码，加密密码
	if user.Password != "" {
		hashedPassword, err := utils.HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	} else {
		// 如果没有提供新密码，保留原密码
		user.Password = existingUser.Password
	}

	// 更新用户
	if err := global.DB.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *UserService) DeleteUser(id uint) error {
	// 检查用户是否存在
	var user system.SysUser
	if err := global.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// 软删除用户
	if err := global.DB.Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*system.SysUser, error) {
	var user system.SysUser
	if err := global.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUserList 获取用户列表（支持分页和过滤）
func (s *UserService) GetUserList(page, pageSize int, filters map[string]interface{}) ([]system.SysUser, int64, error) {
	var users []system.SysUser
	var total int64

	// 构建查询
	query := global.DB.Model(&system.SysUser{})

	// 应用过滤条件
	if username, ok := filters["username"].(string); ok && username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if nickname, ok := filters["nickname"].(string); ok && nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if phone, ok := filters["phone"].(string); ok && phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if email, ok := filters["email"].(string); ok && email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if roleID, ok := filters["role_id"].(uint); ok && roleID > 0 {
		query = query.Where("role_id = ?", roleID)
	}
	if active, ok := filters["active"].(bool); ok {
		query = query.Where("active = ?", active)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}

	return users, total, nil
}

// ChangePassword 修改密码（需要验证旧密码）
func (s *UserService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	// 查询用户
	var user system.SysUser
	if err := global.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// 验证旧密码
	if !utils.CheckPassword(user.Password, oldPassword) {
		return errors.New("old password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := global.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ResetPassword 重置密码（管理员操作，不需要验证旧密码）
func (s *UserService) ResetPassword(userID uint, newPassword string) error {
	// 查询用户
	var user system.SysUser
	if err := global.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := global.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ToggleUserStatus 切换用户状态（启用/禁用）
func (s *UserService) ToggleUserStatus(userID uint, active bool) error {
	// 查询用户
	var user system.SysUser
	if err := global.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// 更新状态
	if err := global.DB.Model(&user).Update("active", active).Error; err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}
