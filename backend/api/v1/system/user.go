package system

import (
	"strconv"

	"k-admin-system/model/common"
	"k-admin-system/model/system"
	systemService "k-admin-system/service/system"

	"github.com/gin-gonic/gin"
)

type UserApi struct{}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string          `json:"accessToken"`
	RefreshToken string          `json:"refreshToken"`
	User         *system.SysUser `json:"user"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Nickname  string `json:"nickname"`
	HeaderImg string `json:"headerImg"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	RoleID    uint   `json:"roleId" binding:"required"`
	Active    bool   `json:"active"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID        uint   `json:"id" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password"` // 可选，如果提供则更新密码
	Nickname  string `json:"nickname"`
	HeaderImg string `json:"headerImg"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	RoleID    uint   `json:"roleId" binding:"required"`
	Active    bool   `json:"active"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	UserID      uint   `json:"userId" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ToggleStatusRequest 切换状态请求
type ToggleStatusRequest struct {
	UserID uint `json:"userId" binding:"required"`
	Active bool `json:"active"`
}

// GetUserListRequest 获取用户列表请求
type GetUserListRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=100"`
	Username string `form:"username"`
	Nickname string `form:"nickname"`
	Phone    string `form:"phone"`
	Email    string `form:"email"`
	RoleID   uint   `form:"roleId"`
	Active   *bool  `form:"active"` // 使用指针以区分未设置和false
}

// GetUserListResponse 获取用户列表响应
type GetUserListResponse struct {
	List  []system.SysUser `json:"list"`
	Total int64            `json:"total"`
}

// Login godoc
// @Summary 用户登录
// @Description 验证用户凭据并返回访问令牌和刷新令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} common.Response{data=LoginResponse} "登录成功"
// @Failure 200 {object} common.Response "登录失败"
// @Router /api/v1/user/login [post]
func (a *UserApi) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	userService := systemService.UserService{}
	accessToken, refreshToken, user, err := userService.Login(req.Username, req.Password)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// CreateUser godoc
// @Summary 创建用户
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateUserRequest true "创建用户请求"
// @Success 200 {object} common.Response{data=system.SysUser} "创建成功"
// @Failure 200 {object} common.Response "创建失败"
// @Router /api/v1/user [post]
func (a *UserApi) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	user := &system.SysUser{
		Username:  req.Username,
		Password:  req.Password,
		Nickname:  req.Nickname,
		HeaderImg: req.HeaderImg,
		Phone:     req.Phone,
		Email:     req.Email,
		RoleID:    req.RoleID,
		Active:    req.Active,
	}

	userService := systemService.UserService{}
	if err := userService.CreateUser(user); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, user)
}

// UpdateUser godoc
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateUserRequest true "更新用户请求"
// @Success 200 {object} common.Response{data=system.SysUser} "更新成功"
// @Failure 200 {object} common.Response "更新失败"
// @Router /api/v1/user [put]
func (a *UserApi) UpdateUser(c *gin.Context) {
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	user := &system.SysUser{
		Username:  req.Username,
		Password:  req.Password,
		Nickname:  req.Nickname,
		HeaderImg: req.HeaderImg,
		Phone:     req.Phone,
		Email:     req.Email,
		RoleID:    req.RoleID,
		Active:    req.Active,
	}
	user.ID = req.ID

	userService := systemService.UserService{}
	if err := userService.UpdateUser(user); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, user)
}

// DeleteUser godoc
// @Summary 删除用户
// @Description 删除用户（软删除）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response "删除成功"
// @Failure 200 {object} common.Response "删除失败"
// @Router /api/v1/user/{id} [delete]
func (a *UserApi) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid user ID")
		return
	}

	userService := systemService.UserService{}
	if err := userService.DeleteUser(uint(id)); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "user deleted successfully")
}

// GetUser godoc
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response{data=system.SysUser} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/user/{id} [get]
func (a *UserApi) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid user ID")
		return
	}

	userService := systemService.UserService{}
	user, err := userService.GetUserByID(uint(id))
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, user)
}

// GetUserList godoc
// @Summary 获取用户列表
// @Description 获取用户列表，支持分页和过滤
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码" minimum(1)
// @Param pageSize query int true "每页数量" minimum(1) maximum(100)
// @Param username query string false "用户名（模糊搜索）"
// @Param nickname query string false "昵称（模糊搜索）"
// @Param phone query string false "手机号（模糊搜索）"
// @Param email query string false "邮箱（模糊搜索）"
// @Param roleId query int false "角色ID"
// @Param active query bool false "是否激活"
// @Success 200 {object} common.Response{data=GetUserListResponse} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/user/list [get]
func (a *UserApi) GetUserList(c *gin.Context) {
	var req GetUserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	// 构建过滤条件
	filters := make(map[string]interface{})
	if req.Username != "" {
		filters["username"] = req.Username
	}
	if req.Nickname != "" {
		filters["nickname"] = req.Nickname
	}
	if req.Phone != "" {
		filters["phone"] = req.Phone
	}
	if req.Email != "" {
		filters["email"] = req.Email
	}
	if req.RoleID > 0 {
		filters["role_id"] = req.RoleID
	}
	if req.Active != nil {
		filters["active"] = *req.Active
	}

	userService := systemService.UserService{}
	users, total, err := userService.GetUserList(req.Page, req.PageSize, filters)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, GetUserListResponse{
		List:  users,
		Total: total,
	})
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 用户修改自己的密码（需要验证旧密码）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} common.Response "修改成功"
// @Failure 200 {object} common.Response "修改失败"
// @Router /api/v1/user/change-password [post]
func (a *UserApi) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	// 从JWT中获取当前用户ID（这里假设JWT中间件会设置userID）
	userID, exists := c.Get("userID")
	if !exists {
		common.Fail(c, "user not authenticated")
		return
	}

	userService := systemService.UserService{}
	if err := userService.ChangePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "password changed successfully")
}

// ResetPassword godoc
// @Summary 重置密码
// @Description 管理员重置用户密码（不需要验证旧密码）
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} common.Response "重置成功"
// @Failure 200 {object} common.Response "重置失败"
// @Router /api/v1/user/reset-password [post]
func (a *UserApi) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	userService := systemService.UserService{}
	if err := userService.ResetPassword(req.UserID, req.NewPassword); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "password reset successfully")
}

// ToggleStatus godoc
// @Summary 切换用户状态
// @Description 启用或禁用用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body ToggleStatusRequest true "切换状态请求"
// @Success 200 {object} common.Response "操作成功"
// @Failure 200 {object} common.Response "操作失败"
// @Router /api/v1/user/toggle-status [post]
func (a *UserApi) ToggleStatus(c *gin.Context) {
	var req ToggleStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	userService := systemService.UserService{}
	if err := userService.ToggleUserStatus(req.UserID, req.Active); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "user status updated successfully")
}
