package system

import (
	"k-admin-system/model/common"
	systemService "k-admin-system/service/system"

	"github.com/gin-gonic/gin"
)

type DashboardApi struct{}

// DashboardStatsResponse 仪表盘统计数据响应
type DashboardStatsResponse struct {
	UserCount   int64 `json:"userCount"`
	RoleCount   int64 `json:"roleCount"`
	MenuCount   int64 `json:"menuCount"`
	ConfigCount int64 `json:"configCount"`
}

// GetDashboardStats godoc
// @Summary 获取仪表盘统计数据
// @Description 获取系统各模块的统计数据
// @Tags 仪表盘
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response{data=DashboardStatsResponse} "获取成功"
// @Failure 200 {object} common.Response "获取失败"
// @Router /api/v1/dashboard/stats [get]
func (a *DashboardApi) GetDashboardStats(c *gin.Context) {
	dashboardService := systemService.DashboardService{}
	stats, err := dashboardService.GetDashboardStats()
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.OkWithData(c, stats)
}
