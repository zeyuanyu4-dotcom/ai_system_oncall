package handler

import (
	"strconv"

	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

// AIAnalysisTaskHandler AI分析任务处理器
type AIAnalysisTaskHandler struct {
	taskService *service.AIAnalysisTaskService
}

// NewAIAnalysisTaskHandler 创建处理器
func NewAIAnalysisTaskHandler(taskService *service.AIAnalysisTaskService) *AIAnalysisTaskHandler {
	return &AIAnalysisTaskHandler{taskService: taskService}
}

// CreateTask 创建分析任务
func (h *AIAnalysisTaskHandler) CreateTask(c *gin.Context) {
	// 检查 AI 服务是否启用
	if config.GetConfig() != nil && !config.GetConfig().AI.Enabled {
		response.Fail(c, 10001, "AI 服务未启用")
		return
	}

	issueID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	task, err := h.taskService.CreateTask(issueID)
	if err != nil {
		response.Fail(c, 10001, err.Error())
		return
	}

	// 异步执行任务
	go h.taskService.ExecuteTask(task.ID)

	response.Success(c, task)
}

// GetTask 获取任务状态
func (h *AIAnalysisTaskHandler) GetTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.taskService.GetTask(taskID)
	if err != nil {
		response.Fail(c, 10001, "任务不存在")
		return
	}

	response.Success(c, task)
}

// GetIssueTasks 获取问题的所有分析任务
func (h *AIAnalysisTaskHandler) GetIssueTasks(c *gin.Context) {
	issueID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	tasks, err := h.taskService.GetTasksByIssueID(issueID)
	if err != nil {
		response.Fail(c, 10001, "获取任务列表失败")
		return
	}

	response.Success(c, tasks)
}

// CancelTask 取消任务
func (h *AIAnalysisTaskHandler) CancelTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.taskService.CancelTask(taskID); err != nil {
		response.Fail(c, 10001, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "已取消"})
}

// UpdateTaskProgress 更新任务进度（供 Python Agent 回调）
func (h *AIAnalysisTaskHandler) UpdateTaskProgress(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Progress    string `json:"progress"`
		CurrentStep string `json:"current_step"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	if err := h.taskService.UpdateTaskProgress(taskID, req.Progress, req.CurrentStep); err != nil {
		response.Fail(c, 10001, err.Error())
		return
	}

	response.Success(c, nil)
}
