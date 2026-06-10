package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	agentv1 "ai_system_oncall/api/proto/agent/v1"
	"ai_system_oncall/internal/client"
	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/grpcclient"
	"ai_system_oncall/internal/mq"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"
	"ai_system_oncall/pkg/jwt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// AIAnalysisHandler AI 分析任务处理器
type AIAnalysisHandler struct {
	taskRepo   *repository.AIAnalysisTaskRepository
	issueRepo  *repository.IssueRepository
	aiClient   *client.AIClient
	grpcAgent  *grpcclient.AgentClient // 【新增】gRPC 直连 Agent
	mqClient   *mq.RabbitMQClient
	cancelMgr  *CancelManager
	logger     *zap.Logger
	timeout    time.Duration
}

// NewAIAnalysisHandler 创建处理器
func NewAIAnalysisHandler(
	taskRepo *repository.AIAnalysisTaskRepository,
	issueRepo *repository.IssueRepository,
	aiClient *client.AIClient,
	grpcAgent *grpcclient.AgentClient,
	mqClient *mq.RabbitMQClient,
	cancelMgr *CancelManager,
	logger *zap.Logger,
	timeoutSeconds int,
) *AIAnalysisHandler {
	return &AIAnalysisHandler{
		taskRepo:  taskRepo,
		issueRepo: issueRepo,
		aiClient:  aiClient,
		grpcAgent: grpcAgent,
		mqClient:  mqClient,
		cancelMgr: cancelMgr,
		logger:    logger,
		timeout:   time.Duration(timeoutSeconds) * time.Second,
	}
}

// shouldUseMQ 判断是否使用 MQ（灰度逻辑）
func (h *AIAnalysisHandler) shouldUseMQ(taskID uint64) bool {
	cfg := config.GetConfig()
	if cfg == nil || !cfg.RabbitMQ.Enabled {
		return false
	}

	// 灰度百分比
	grayPercent := cfg.RabbitMQ.GrayPercent
	if grayPercent <= 0 {
		return false
	}
	if grayPercent >= 100 {
		return true
	}

	// 根据 taskID 取模判断
	return int(taskID%100) < grayPercent
}

// ProcessTask 处理任务
func (h *AIAnalysisHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	// 解析载荷
	var payload AIAnalysisPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		h.logger.Error("Failed to unmarshal payload", zap.Error(err))
		return fmt.Errorf("unmarshal payload failed: %w", err)
	}

	taskID := payload.TaskID
	issueID := payload.IssueID
	userToken := payload.UserToken

	h.logger.Info("Processing AI analysis task",
		zap.Uint64("task_id", taskID),
		zap.Uint64("issue_id", issueID),
		zap.Bool("has_user_token", userToken != ""))

	// 检查是否已取消
	if h.cancelMgr.IsCancelled(taskID) {
		h.logger.Info("Task already cancelled", zap.Uint64("task_id", taskID))
		h.markTaskCancelled(taskID)
		return nil
	}

	// 获取任务
	task, err := h.taskRepo.FindByID(taskID)
	if err != nil {
		h.logger.Error("Failed to find task", zap.Uint64("task_id", taskID), zap.Error(err))
		return err
	}

	// 检查状态
	if task.Status != model.TaskStatusPending {
		h.logger.Warn("Task status is not pending",
			zap.Uint64("task_id", taskID),
			zap.String("status", task.Status))
		return nil
	}

	// 更新状态为运行中
	if err := h.startTask(task); err != nil {
		h.logger.Error("Failed to start task", zap.Uint64("task_id", taskID), zap.Error(err))
		return err
	}

	// 判断使用 MQ 还是 HTTP
	useMQ := h.shouldUseMQ(taskID)
	h.logger.Info("Executing analysis",
		zap.Uint64("task_id", taskID),
		zap.Bool("use_mq", useMQ))

	var result *mq.AnalysisResult
	var execErr error

	if useMQ && h.mqClient != nil && h.mqClient.IsEnabled() {
		// MQ 方式：发布命令，等待结果（由 Consumer 处理）
		execErr = h.executeViaMQ(ctx, task, issueID)
		// MQ 方式结果由 Consumer 更新，这里直接返回
		if execErr != nil {
			h.markTaskFailed(task, execErr.Error())
		}
		return execErr
	}

	// 判断走 gRPC 还是 HTTP 兜底
	if h.grpcAgent != nil {
		result, execErr = h.executeAnalysisGRPC(ctx, task, issueID, userToken)
	} else {
		// HTTP 方式：直接调用 Agent（兜底/兼容）
		timeoutCtx, cancel := context.WithTimeout(ctx, h.timeout)
		defer cancel()

		result, execErr = h.executeAnalysisHTTP(timeoutCtx, task, issueID, userToken)
	}

	// 再次检查取消状态
	if h.cancelMgr.IsCancelled(taskID) {
		h.logger.Info("Task cancelled during execution", zap.Uint64("task_id", taskID))
		h.markTaskCancelled(taskID)
		return nil
	}

	if execErr != nil {
		h.logger.Error("Analysis failed",
			zap.Uint64("task_id", taskID),
			zap.Error(execErr))
		h.markTaskFailed(task, execErr.Error())
		// 返回错误触发重试
		return fmt.Errorf("analysis failed: %w", execErr)
	}

	// 标记完成（仅 HTTP 方式）
	if result != nil {
		if err := h.completeTask(task, result); err != nil {
			h.logger.Error("Failed to complete task", zap.Uint64("task_id", taskID), zap.Error(err))
			return err
		}
	}

	h.logger.Info("Task completed successfully", zap.Uint64("task_id", taskID))
	return nil
}

// startTask 开始任务
func (h *AIAnalysisHandler) startTask(task *model.AIAnalysisTask) error {
	now := time.Now()
	task.Status = model.TaskStatusRunning
	task.StartedAt = &now
	task.Progress = "0/8"
	task.CurrentStep = "正在启动分析..."
	return h.taskRepo.Update(task)
}

// executeViaMQ 通过 MQ 执行分析
func (h *AIAnalysisHandler) executeViaMQ(ctx context.Context, task *model.AIAnalysisTask, issueID uint64) error {
	// 发布命令到 MQ
	cmd := &mq.AnalysisCommand{
		TaskID:  task.ID,
		IssueID: issueID,
		Payload: map[string]interface{}{
			"task_id":  task.ID,
			"issue_id": issueID,
		},
	}

	if err := h.mqClient.PublishCommand(ctx, cmd); err != nil {
		return fmt.Errorf("publish command failed: %w", err)
	}

	h.logger.Info("Analysis command published to MQ",
		zap.Uint64("task_id", task.ID))

	// MQ 方式：结果由 Python Agent 通过 MQ 返回，Go Consumer 更新数据库
	// 这里任务已经发布，返回 nil 表示 Asynq 任务完成
	// 实际结果由 MQ Consumer 处理
	return nil
}

// executeAnalysisHTTP 通过 HTTP 执行分析（原有逻辑）
func (h *AIAnalysisHandler) executeAnalysisHTTP(ctx context.Context, task *model.AIAnalysisTask, issueID uint64, userToken string) (*mq.AnalysisResult, error) {
	// 获取问题信息
	issue, err := h.issueRepo.FindByID(issueID)
	if err != nil {
		return nil, fmt.Errorf("find issue failed: %w", err)
	}

	// 检查取消
	if h.cancelMgr.IsCancelled(task.ID) {
		return nil, errors.New("task cancelled")
	}

	// 校验用户 JWT：撤销或过期后，已入队任务不应再继续（防止撤销不及时）
	if userToken != "" {
		if _, err := jwt.ParseToken(userToken); err != nil {
			h.logger.Warn("User token invalid or expired, abort task",
				zap.Uint64("task_id", task.ID),
				zap.Error(err))
			return nil, fmt.Errorf("user token invalid/expired: %w", err)
		}
	}

	// 构建请求
	req := &client.AgentAnalysisRequest{
		TaskID:       int(task.ID),
		IssueID:      int(issue.ID),
		IssueNo:      issue.IssueNo,
		Title:        issue.Title,
		Description:  issue.Description,
		ErrorMessage: issue.ErrorMessage,
		LogExcerpt:   issue.LogExcerpt,
		Environment:  issue.Environment,
		ProjectID:    issue.ProjectID,
		ProjectName:  "",
		ServiceName:  "",
		ImpactScope:  issue.ImpactScope,
	}

	if issue.Project != nil {
		req.ProjectName = issue.Project.Name
	}
	if issue.Service != nil {
		req.ServiceName = issue.Service.Name
	}

	// 调用 Agent（带超时）
	resultChan := make(chan *client.AgentResult, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := h.aiClient.AgentAnalyze(req, userToken)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case agentResult := <-resultChan:
		// 转换为 MQ 结果格式
		return &mq.AnalysisResult{
			TaskID:  task.ID,
			IssueID: issueID,
			Success: true,
			Summary: agentResult.Summary,
			Result: map[string]interface{}{
				"summary":     agentResult.Summary,
				"root_cause":  agentResult.SuspectedCause,
				"solutions":   agentResult.Suggestions,
				"tool_calls":  agentResult.ToolCalls,
			},
		}, nil
	case err := <-errChan:
		return nil, err
	}
}

// executeAnalysisGRPC 通过 gRPC streaming 执行分析
func (h *AIAnalysisHandler) executeAnalysisGRPC(ctx context.Context, task *model.AIAnalysisTask, issueID uint64, userToken string) (*mq.AnalysisResult, error) {
	issue, err := h.issueRepo.FindByID(issueID)
	if err != nil {
		return nil, fmt.Errorf("find issue failed: %w", err)
	}

	if h.cancelMgr.IsCancelled(task.ID) {
		return nil, errors.New("task cancelled")
	}

	// JWT 前置校验
	if userToken != "" {
		if _, err := jwt.ParseToken(userToken); err != nil {
			h.logger.Warn("User token invalid or expired, abort task", zap.Uint64("task_id", task.ID), zap.Error(err))
			return nil, fmt.Errorf("user token invalid/expired: %w", err)
		}
	}

	req := &agentv1.RunAgentRequest{
		TaskId:       int64(task.ID),
		IssueId:      int64(issue.ID),
		IssueNo:      issue.IssueNo,
		Title:        issue.Title,
		Description:  issue.Description,
		ErrorMessage: issue.ErrorMessage,
		LogExcerpt:   issue.LogExcerpt,
		Environment:  issue.Environment,
		ProjectId:    issue.ProjectID,
		ProjectName:  "",
		ServiceName:  "",
		ImpactScope:  issue.ImpactScope,
	}
	if issue.Project != nil {
		req.ProjectName = issue.Project.Name
	}
	if issue.Service != nil {
		req.ServiceName = issue.Service.Name
	}

	stream, err := h.grpcAgent.RunAgent(ctx, req, userToken)
	if err != nil {
		return nil, fmt.Errorf("start gRPC RunAgent failed: %w", err)
	}

	var finalResult *mq.AnalysisResult
	for {
		resp, err := stream.Recv()
		if err != nil {
			return nil, fmt.Errorf("gRPC stream recv error: %w", err)
		}

		switch ev := resp.Event.(type) {
		case *agentv1.RunAgentResponse_Progress:
			// 进度事件：写 DB + 打日志
			h.logger.Info("Agent progress",
				zap.Uint64("task_id", task.ID),
				zap.Int32("step", ev.Progress.Step),
				zap.String("message", ev.Progress.Message))
			// 简化为：更新当前 step
			p := fmt.Sprintf("%d/8", ev.Progress.Step)
			h.taskRepo.UpdateProgress(task.ID, p, ev.Progress.Message)

		case *agentv1.RunAgentResponse_Result:
			// 最终结果
			r := ev.Result
			finalResult = &mq.AnalysisResult{
				TaskID:  task.ID,
				IssueID: issueID,
				Success: true,
				Summary: r.Summary,
				Result: map[string]interface{}{
					"summary":     r.Summary,
					"issue_type":  r.IssueType,
					"root_cause":  r.SuspectedCause,
					"solutions":   r.Suggestions,
					"confidence":  r.Confidence,
					"tool_calls":  agentToolCallsToInterface(r.ToolCalls),
				},
			}
			return finalResult, nil

		case *agentv1.RunAgentResponse_Error:
			return nil, fmt.Errorf("agent analysis error: %s - %s", ev.Error.Code, ev.Error.Message)
		}
	}
}

func agentToolCallsToInterface(tcs []*agentv1.ToolCallRecord) []interface{} {
	out := make([]interface{}, len(tcs))
	for i, tc := range tcs {
		out[i] = map[string]interface{}{
			"step":       tc.Step,
			"tool_name":  tc.ToolName,
			"input":      tc.Input,
			"output":     tc.Output,
			"thought":    tc.Thought,
			"executed_at": tc.ExecutedAt,
			"duration_ms": tc.DurationMs,
		}
	}
	return out
}

// completeTask 完成任务（HTTP 方式）
func (h *AIAnalysisHandler) completeTask(task *model.AIAnalysisTask, result *mq.AnalysisResult) error {
	task.Status = model.TaskStatusCompleted
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.Progress = "8/8"
	task.CurrentStep = "分析完成"

	resultJSON, _ := json.Marshal(result.Result)
	task.Result = string(resultJSON)

	if err := h.taskRepo.Update(task); err != nil {
		return err
	}

	// 更新问题的 AI 分析结果
	h.issueRepo.UpdateAIAnalysis(task.IssueID, result.Summary, string(resultJSON))

	return nil
}

// markTaskFailed 标记任务失败
func (h *AIAnalysisHandler) markTaskFailed(task *model.AIAnalysisTask, errMsg string) {
	task.Status = model.TaskStatusFailed
	task.ErrorMessage = errMsg
	now := time.Now()
	task.CompletedAt = &now
	h.taskRepo.Update(task)
}

// markTaskCancelled 标记任务取消
func (h *AIAnalysisHandler) markTaskCancelled(taskID uint64) {
	task, err := h.taskRepo.FindByID(taskID)
	if err != nil {
		return
	}
	task.Status = model.TaskStatusCancelled
	now := time.Now()
	task.CompletedAt = &now
	task.CurrentStep = "任务已取消"
	h.taskRepo.Update(task)

	// 清除取消状态
	h.cancelMgr.ClearCancelled(taskID)
}
