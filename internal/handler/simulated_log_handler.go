package handler

import (
	"strconv"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

type SimulatedLogHandler struct {
	logService *service.SimulatedLogService
}

func NewSimulatedLogHandler(logService *service.SimulatedLogService) *SimulatedLogHandler {
	return &SimulatedLogHandler{logService: logService}
}

// CreateLog creates a new simulated log
func (h *SimulatedLogHandler) CreateLog(c *gin.Context) {
	var req dto.CreateSimulatedLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// Check project access
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)
	if globalRole != "system_admin" {
		hasAccess, err := h.logService.CheckUserProjectAccess(userID, req.ProjectID)
		if err != nil || !hasAccess {
			response.Fail(c, 10001, "无权限操作该项目")
			return
		}
	}

	logInfo, err := h.logService.CreateLog(&req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, logInfo)
}

// BatchCreateLogs creates multiple simulated logs
func (h *SimulatedLogHandler) BatchCreateLogs(c *gin.Context) {
	var req dto.BatchCreateSimulatedLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// Validate access for all logs (check first log's project)
	if len(req.Logs) > 0 {
		userID := middleware.GetUserID(c)
		globalRole := middleware.GetUserRole(c)
		if globalRole != "system_admin" {
			hasAccess, err := h.logService.CheckUserProjectAccess(userID, req.Logs[0].ProjectID)
			if err != nil || !hasAccess {
				response.Fail(c, 10001, "无权限操作该项目")
				return
			}
		}
	}

	count, err := h.logService.BatchCreateLogs(&req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, gin.H{"created_count": count})
}

// GetLog gets a log by ID
func (h *SimulatedLogHandler) GetLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的日志ID")
		return
	}

	logInfo, err := h.logService.GetLog(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	// Check project access
	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)
	if globalRole != "system_admin" {
		hasAccess, err := h.logService.CheckUserProjectAccess(userID, logInfo.ProjectID)
		if err != nil || !hasAccess {
			response.Fail(c, 10001, "无权限查看该日志")
			return
		}
	}

	response.Success(c, logInfo)
}

// ListLogs lists logs with filters
func (h *SimulatedLogHandler) ListLogs(c *gin.Context) {
	var req dto.SimulatedLogListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// Check project access if project_id is specified
	if req.ProjectID > 0 {
		userID := middleware.GetUserID(c)
		globalRole := middleware.GetUserRole(c)
		if globalRole != "system_admin" {
			hasAccess, err := h.logService.CheckUserProjectAccess(userID, req.ProjectID)
			if err != nil || !hasAccess {
				response.Fail(c, 10001, "无权限查看该项目日志")
				return
			}
		}
	}

	result, err := h.logService.ListLogs(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetLogsByTraceID gets logs by trace ID
func (h *SimulatedLogHandler) GetLogsByTraceID(c *gin.Context) {
	traceID := c.Param("trace_id")
	if traceID == "" {
		response.BadRequest(c, "Trace ID不能为空")
		return
	}

	logs, err := h.logService.GetLogsByTraceID(traceID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, logs)
}

// GetLogsByService gets logs by service ID
func (h *SimulatedLogHandler) GetLogsByService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的服务ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.logService.GetLogsByService(serviceID, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// LinkIssue links a log to an issue
func (h *SimulatedLogHandler) LinkIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的日志ID")
		return
	}

	var req dto.LinkIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// Get log to check project access
	logInfo, err := h.logService.GetLog(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)
	if globalRole != "system_admin" {
		hasAccess, err := h.logService.CheckUserProjectAccess(userID, logInfo.ProjectID)
		if err != nil || !hasAccess {
			response.Fail(c, 10001, "无权限操作该日志")
			return
		}
	}

	if err := h.logService.LinkIssue(id, req.IssueID); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteLog deletes a log
func (h *SimulatedLogHandler) DeleteLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的日志ID")
		return
	}

	// Get log to check project access
	logInfo, err := h.logService.GetLog(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)
	if globalRole != "system_admin" {
		hasAccess, err := h.logService.CheckUserProjectAccess(userID, logInfo.ProjectID)
		if err != nil || !hasAccess {
			response.Fail(c, 10001, "无权限删除该日志")
			return
		}
	}

	if err := h.logService.DeleteLog(id); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetLogsByIssue gets logs linked to an issue
func (h *SimulatedLogHandler) GetLogsByIssue(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	logs, err := h.logService.GetLogsByIssue(issueID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, logs)
}
