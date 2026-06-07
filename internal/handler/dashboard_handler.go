package handler

import (
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetDashboardStats 获取看板统计数据
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	result, err := h.dashboardService.GetDashboardStats(userRole, userID)
	if err != nil {
		response.Fail(c, 30001, err.Error())
		return
	}

	response.Success(c, result)
}

// GetTrendData 获取趋势图数据
func (h *DashboardHandler) GetTrendData(c *gin.Context) {
	var req dto.TrendQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)

	// 获取项目过滤条件
	var projectID *uint64
	if userRole == "project_admin" {
		projectIDs, err := h.dashboardService.GetUserProjectFilter(userRole, userID)
		if err != nil {
			response.Fail(c, 30001, err.Error())
			return
		}
		projectID = projectIDs
	} else if userRole != "system_admin" {
		response.Fail(c, 30001, "无权限")
		return
	}

	result, err := h.dashboardService.GetTrendData(req.StartDate, req.EndDate, projectID)
	if err != nil {
		response.Fail(c, 30002, err.Error())
		return
	}

	response.Success(c, result)
}

// GenerateDailyStat 生成每日统计（定时任务调用）
func (h *DashboardHandler) GenerateDailyStat(c *gin.Context) {
	var req dto.GenerateDailyStatRequest
	c.ShouldBindJSON(&req)

	err := h.dashboardService.GenerateDailyStat(req.Date)
	if err != nil {
		response.Fail(c, 30003, err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "统计生成成功",
	})
}
