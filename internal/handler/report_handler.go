package handler

import (
	"strconv"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GenerateDailyReport generates a daily report manually
func (h *ReportHandler) GenerateDailyReport(c *gin.Context) {
	var req dto.GenerateDailyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no date provided, use today
		req.Date = ""
	}

	userID := middleware.GetUserID(c)
	result, err := h.reportService.GenerateDailyReport(userID, req.Date, false)
	if err != nil {
		response.Fail(c, 20001, err.Error())
		return
	}

	response.Success(c, result)
}

// GetDailyReport gets a daily report by date
func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		response.BadRequest(c, "请指定日期")
		return
	}

	result, err := h.reportService.GetDailyReport(date)
	if err != nil {
		response.Fail(c, 20002, "未找到该日期的报告")
		return
	}

	response.Success(c, result)
}

// GenerateDailyReportAuto generates a daily report automatically (for cron)
func (h *ReportHandler) GenerateDailyReportAuto(c *gin.Context) {
	// Use system user ID for auto reports
	result, err := h.reportService.GenerateDailyReport(1, "", true)
	if err != nil {
		response.Fail(c, 20001, err.Error())
		return
	}

	response.Success(c, result)
}

// GenerateWeeklyReport generates a weekly report
func (h *ReportHandler) GenerateWeeklyReport(c *gin.Context) {
	var req dto.GenerateWeeklyReportRequest
	// 允许不传参数，后端自动计算本周一
	c.ShouldBindJSON(&req)

	userID := middleware.GetUserID(c)
	result, err := h.reportService.GenerateWeeklyReport(userID, req.WeekStart, false)
	if err != nil {
		response.Fail(c, 20001, err.Error())
		return
	}

	response.Success(c, result)
}

// GetWeeklyReport gets a weekly report by week start date
func (h *ReportHandler) GetWeeklyReport(c *gin.Context) {
	weekStart := c.Param("week")
	if weekStart == "" {
		response.BadRequest(c, "请指定周开始日期")
		return
	}

	result, err := h.reportService.GetWeeklyReport(weekStart)
	if err != nil {
		response.Fail(c, 20002, "未找到该周的报告")
		return
	}

	response.Success(c, result)
}

// GenerateIncidentReview generates an incident review report
func (h *ReportHandler) GenerateIncidentReview(c *gin.Context) {
	var req dto.GenerateIncidentReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请选择需要复盘的问题单")
		return
	}

	userID := middleware.GetUserID(c)
	result, err := h.reportService.GenerateIncidentReview(userID, req.IssueIDs)
	if err != nil {
		response.Fail(c, 20001, err.Error())
		return
	}

	response.Success(c, result)
}

// GetIncidentReview gets an incident review by report ID
func (h *ReportHandler) GetIncidentReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报告ID")
		return
	}

	result, err := h.reportService.GetReport(id)
	if err != nil {
		response.Fail(c, 20002, "未找到该报告")
		return
	}

	response.Success(c, result)
}

// ListReports lists all reports with pagination
func (h *ReportHandler) ListReports(c *gin.Context) {
	var req dto.ReportListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	result, err := h.reportService.ListReports(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetReport gets a report by ID
func (h *ReportHandler) GetReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的报告ID")
		return
	}

	result, err := h.reportService.GetReport(id)
	if err != nil {
		response.Fail(c, 20002, "未找到该报告")
		return
	}

	response.Success(c, result)
}