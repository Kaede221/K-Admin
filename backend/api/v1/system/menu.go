package system

import (
	"strconv"

	"k-admin-system/global"
	"k-admin-system/model/common"
	"k-admin-system/model/system"
	systemService "k-admin-system/service/system"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MenuApi struct{}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	ParentID  uint            `json:"parentId"`
	Path      string          `json:"path" binding:"required"`
	Name      string          `json:"name" binding:"required"`
	Component string          `json:"component"`
	Sort      int             `json:"sort"`
	Meta      system.MenuMeta `json:"meta"`
	BtnPerms  []string        `json:"btnPerms"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	ID        uint            `json:"id" binding:"required"`
	ParentID  uint            `json:"parentId"`
	Path      string          `json:"path" binding:"required"`
	Name      string          `json:"name" binding:"required"`
	Component string          `json:"component"`
	Sort      int             `json:"sort"`
	Meta      system.MenuMeta `json:"meta"`
	BtnPerms  []string        `json:"btnPerms"`
}

// GetMenuTreeRequest 获取菜单树请求
type GetMenuTreeRequest struct {
	RoleID uint `form:"roleId"`
}

// CreateMenu godoc
// @Summary 创建菜单
// @Description 创建新菜单
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body CreateMenuRequest true "创建菜单请求"
// @Success 200 {object} common.Response{data=system.SysMenu} "创建成功"
// @Failure 200 {object} common.Response "创建失败"
// @Router /api/v1/menu [post]
func (a *MenuApi) CreateMenu(c *gin.Context) {
	var req CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	menu := &system.SysMenu{
		ParentID:  req.ParentID,
		Path:      req.Path,
		Name:      req.Name,
		Component: req.Component,
		Sort:      req.Sort,
		Meta:      req.Meta,
		BtnPerms:  req.BtnPerms,
	}

	menuService := systemService.MenuService{}
	if err := menuService.CreateMenu(menu); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, menu)
}

// UpdateMenu godoc
// @Summary 更新菜单
// @Description 更新菜单信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body UpdateMenuRequest true "更新菜单请求"
// @Success 200 {object} common.Response{data=system.SysMenu} "更新成功"
// @Failure 200 {object} common.Response "更新失败"
// @Router /api/v1/menu [put]
func (a *MenuApi) UpdateMenu(c *gin.Context) {
	var req UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	menu := &system.SysMenu{
		ParentID:  req.ParentID,
		Path:      req.Path,
		Name:      req.Name,
		Component: req.Component,
		Sort:      req.Sort,
		Meta:      req.Meta,
		BtnPerms:  req.BtnPerms,
	}
	menu.ID = req.ID

	menuService := systemService.MenuService{}
	if err := menuService.UpdateMenu(menu); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, menu)
}

// DeleteMenu godoc
// @Summary 删除菜单
// @Description 删除菜单（不能删除有子菜单的菜单）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "菜单ID"
// @Success 200 {object} common.Response "删除成功"
// @Failure 200 {object} common.Response "删除失败"
// @Router /api/v1/menu/{id} [delete]
func (a *MenuApi) DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid menu ID")
		return
	}

	menuService := systemService.MenuService{}
	if err := menuService.DeleteMenu(uint(id)); err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithDetailed(c, nil, "menu deleted successfully")
}

// GetMenu godoc
// @Summary 获取菜单详情
// @Description 根据ID获取菜单详细信息
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "菜单ID"
// @Success 200 {object} common.Response{data=system.SysMenu} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/menu/{id} [get]
func (a *MenuApi) GetMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, "invalid menu ID")
		return
	}

	menuService := systemService.MenuService{}
	menu, err := menuService.GetMenuByID(uint(id))
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, menu)
}

// GetAllMenus godoc
// @Summary 获取所有菜单
// @Description 获取所有菜单列表（不构建树结构）
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response{data=[]system.SysMenu} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/menu/all [get]
func (a *MenuApi) GetAllMenus(c *gin.Context) {
	menuService := systemService.MenuService{}
	menus, err := menuService.GetAllMenus()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, menus)
}

// GetMenuTree godoc
// @Summary 获取菜单树
// @Description 获取菜单树结构，可根据角色过滤
// @Tags 菜单管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param roleId query int false "角色ID（0表示获取所有菜单）"
// @Success 200 {object} common.Response{data=[]system.SysMenu} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/menu/tree [get]
func (a *MenuApi) GetMenuTree(c *gin.Context) {
	var req GetMenuTreeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, "invalid request parameters: "+err.Error())
		return
	}

	// 记录请求参数
	global.Logger.Info("GetMenuTree API called",
		zap.Uint("roleID", req.RoleID),
		zap.String("queryString", c.Request.URL.RawQuery))

	menuService := systemService.MenuService{}
	tree, err := menuService.GetMenuTree(req.RoleID)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, tree)
}
