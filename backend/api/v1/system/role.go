package system

import (
	"strconv"

	"k-admin-system/model/common"
	"k-admin-system/model/system"
	systemService "k-admin-system/service/system"

	"github.com/gin-gonic/gin"
)

type RoleApi struct{}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleName  string `json:"roleName" binding:"required"`
	RoleKey   string `json:"roleKey" binding:"required"`
	DataScope string `json:"dataScope"`
	Sort      int    `json:"sort"`
	Status    bool   `json:"status"`
	Remark    string `json:"remark"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID        uint   `json:"id" binding:"required"`
	RoleName  string `json:"roleName" binding:"required"`
	RoleKey   string `json:"roleKey" binding:"required"`
	DataScope string `json:"dataScope"`
	Sort      int    `json:"sort"`
	Status    bool   `json:"status"`
	Remark    string `json:"remark"`
}

// GetRoleListRequest 获取角色列表请求
type GetRoleListRequest struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=100"`
}

// GetRoleListResponse 获取角色列表响应
type GetRoleListResponse struct {
	List  []system.SysRole `json:"list"`
	Total int64            `json:"total"`
}

// AssignMenusRequest 分配菜单权限请求
type AssignMenusRequest struct {
	RoleID  uint   `json:"roleId" binding:"required"`
	MenuIDs []uint `json:"menuIds"`
}

// AssignAPIsRequest 分配API权限请求
type AssignAPIsRequest struct {
	RoleID   uint       `json:"roleId" binding:"required"`
	Policies [][]string `json:"policies"`
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateRoleRequest true "创建角色请求"
// @Success 200 {object} common.Response{data=system.SysRole} "创建成功"
// @Failure 200 {object} common.Response "创建失败"
// @Router /api/v1/role [post]
func (a *RoleApi) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	role := &system.SysRole{
		RoleName:  req.RoleName,
		RoleKey:   req.RoleKey,
		DataScope: req.DataScope,
		Sort:      req.Sort,
		Status:    req.Status,
		Remark:    req.Remark,
	}

	roleService := systemService.RoleService{}
	if err := roleService.CreateRole(role); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, role)
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateRoleRequest true "更新角色请求"
// @Success 200 {object} common.Response{data=system.SysRole} "更新成功"
// @Failure 200 {object} common.Response "更新失败"
// @Router /api/v1/role [put]
func (a *RoleApi) UpdateRole(c *gin.Context) {
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	role := &system.SysRole{
		RoleName:  req.RoleName,
		RoleKey:   req.RoleKey,
		DataScope: req.DataScope,
		Sort:      req.Sort,
		Status:    req.Status,
		Remark:    req.Remark,
	}
	role.ID = req.ID

	roleService := systemService.RoleService{}
	if err := roleService.UpdateRole(role); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, role)
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 删除角色（不能删除有关联用户的角色）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "角色ID"
// @Success 200 {object} common.Response "删除成功"
// @Failure 200 {object} common.Response "删除失败"
// @Router /api/v1/role/{id} [delete]
func (a *RoleApi) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid role ID")
		return
	}

	roleService := systemService.RoleService{}
	if err := roleService.DeleteRole(uint(id)); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "role deleted successfully")
}

// GetRole godoc
// @Summary 获取角色详情
// @Description 根据ID获取角色详细信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "角色ID"
// @Success 200 {object} common.Response{data=system.SysRole} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/role/{id} [get]
func (a *RoleApi) GetRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid role ID")
		return
	}

	roleService := systemService.RoleService{}
	role, err := roleService.GetRoleByID(uint(id))
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, role)
}

// GetRoleList godoc
// @Summary 获取角色列表
// @Description 获取角色列表，支持分页
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int true "页码" minimum(1)
// @Param pageSize query int true "每页数量" minimum(1) maximum(100)
// @Success 200 {object} common.Response{data=GetRoleListResponse} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/role/list [get]
func (a *RoleApi) GetRoleList(c *gin.Context) {
	var req GetRoleListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	roleService := systemService.RoleService{}
	roles, total, err := roleService.GetRoleList(req.Page, req.PageSize)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, GetRoleListResponse{
		List:  roles,
		Total: total,
	})
}

// AssignMenus godoc
// @Summary 分配菜单权限
// @Description 为角色分配菜单权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AssignMenusRequest true "分配菜单权限请求"
// @Success 200 {object} common.Response "分配成功"
// @Failure 200 {object} common.Response "分配失败"
// @Router /api/v1/role/assign-menus [post]
func (a *RoleApi) AssignMenus(c *gin.Context) {
	var req AssignMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	roleService := systemService.RoleService{}
	if err := roleService.AssignMenus(req.RoleID, req.MenuIDs); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "menus assigned successfully")
}

// GetRoleMenus godoc
// @Summary 获取角色菜单权限
// @Description 获取角色已分配的菜单ID列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "角色ID"
// @Success 200 {object} common.Response{data=[]uint} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/role/{id}/menus [get]
func (a *RoleApi) GetRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid role ID")
		return
	}

	roleService := systemService.RoleService{}
	menuIDs, err := roleService.GetRoleMenus(uint(id))
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, menuIDs)
}

// AssignAPIs godoc
// @Summary 分配API权限
// @Description 为角色分配API权限（通过Casbin策略）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body AssignAPIsRequest true "分配API权限请求"
// @Success 200 {object} common.Response "分配成功"
// @Failure 200 {object} common.Response "分配失败"
// @Router /api/v1/role/assign-apis [post]
func (a *RoleApi) AssignAPIs(c *gin.Context) {
	var req AssignAPIsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	roleService := systemService.RoleService{}
	if err := roleService.AssignAPIs(req.RoleID, req.Policies); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "API permissions assigned successfully")
}

// GetRoleAPIs godoc
// @Summary 获取角色API权限
// @Description 获取角色已分配的API权限列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "角色ID"
// @Success 200 {object} common.Response{data=[][]string} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/role/{id}/apis [get]
func (a *RoleApi) GetRoleAPIs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid role ID")
		return
	}

	roleService := systemService.RoleService{}
	policies, err := roleService.GetRoleAPIs(uint(id))
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, policies)
}
