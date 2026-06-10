package handler

import (
	"strconv"
	"strings"

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
		response.Fail(c, response.CodeInvalidParam, "无效的问题ID")
		return
	}

	task, err := h.taskService.CreateTask(issueID)
	if err != nil {
		response.Fail(c, 10001, err.Error())
		return
	}

	// 提取用户 JWT，用于在异步链路中透传给 Python Agent
	userToken := ""
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		userToken = strings.TrimPrefix(auth, "Bearer ")
	}

	// 使用 Asynq 入队（替代 go ExecuteTask）
	cfg := config.GetConfig()
	if cfg != nil && cfg.Asynq.Enabled {
		// 新逻辑：入队任务
		if err := h.taskService.EnqueueTask(task.ID, issueID, userToken); err != nil {
			// 入队失败，记录日志但仍返回成功
			// Worker 可以通过启动时扫描恢复
			c.Error(err)
		}
	} else {
		// 兼容逻辑：直接 goroutine 执行（Asynq 未启用时）
		go h.taskService.ExecuteTask(task.ID, userToken)
	}

	response.Success(c, task)
}

// GetTask 获取任务状态
func (h *AIAnalysisTaskHandler) GetTask(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的任务ID")
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
		response.Fail(c, response.CodeInvalidParam, "无效的问题ID")
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
		response.Fail(c, response.CodeInvalidParam, "无效的任务ID")
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
		response.Fail(c, response.CodeInvalidParam, "无效的任务ID")
		return
	}

	var req struct {
		Progress    string `json:"progress"`
		CurrentStep string `json:"current_step"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误")
		return
	}

	if err := h.taskService.UpdateTaskProgress(taskID, req.Progress, req.CurrentStep); err != nil {
		response.Fail(c, 10001, err.Error())
		return
	}

	response.Success(c, nil)
}
