package mq

import (
	"context"
	"encoding/json"

	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"

	"go.uber.org/zap"
)

// ResultConsumer 结果消费者
type ResultConsumer struct {
	taskRepo *repository.AIAnalysisTaskRepository
	issueRepo *repository.IssueRepository
	logger   *zap.Logger
}

// NewResultConsumer 创建结果消费者
func NewResultConsumer(
	taskRepo *repository.AIAnalysisTaskRepository,
	issueRepo *repository.IssueRepository,
	logger *zap.Logger,
) *ResultConsumer {
	return &ResultConsumer{
		taskRepo: taskRepo,
		issueRepo: issueRepo,
		logger:   logger,
	}
}

// HandleResult 处理结果消息
func (c *ResultConsumer) HandleResult(body []byte) error {
	var result AnalysisResult
	if err := json.Unmarshal(body, &result); err != nil {
		c.logger.Error("Failed to unmarshal result", zap.Error(err))
		return err
	}

	c.logger.Info("Processing result message",
		zap.Uint64("task_id", result.TaskID),
		zap.Bool("success", result.Success))

	// 获取任务
	task, err := c.taskRepo.FindByID(result.TaskID)
	if err != nil {
		c.logger.Warn("Task not found for result", zap.Uint64("task_id", result.TaskID))
		return nil // 任务不存在，不再重试
	}

	// 更新任务状态
	if result.Success {
		task.Status = model.TaskStatusCompleted
		task.Progress = "8/8"
		task.CurrentStep = "分析完成"

		resultJSON, _ := json.Marshal(result.Result)
		task.Result = string(resultJSON)

		// 更新问题的 AI 分析结果
		c.issueRepo.UpdateAIAnalysis(task.IssueID, result.Summary, string(resultJSON))
	} else {
		task.Status = model.TaskStatusFailed
		task.ErrorMessage = result.Error
	}

	if err := c.taskRepo.Update(task); err != nil {
		c.logger.Error("Failed to update task",
			zap.Uint64("task_id", result.TaskID),
			zap.Error(err))
		return err
	}

	c.logger.Info("Task updated from result",
		zap.Uint64("task_id", result.TaskID),
		zap.String("status", task.Status))

	return nil
}

// HandleProgress 处理进度消息
func (c *ResultConsumer) HandleProgress(body []byte) error {
	var progress AnalysisProgress
	if err := json.Unmarshal(body, &progress); err != nil {
		c.logger.Error("Failed to unmarshal progress", zap.Error(err))
		return err
	}

	c.logger.Debug("Processing progress message",
		zap.Uint64("task_id", progress.TaskID),
		zap.String("step", progress.CurrentStep))

	// 获取任务
	task, err := c.taskRepo.FindByID(progress.TaskID)
	if err != nil {
		c.logger.Warn("Task not found for progress", zap.Uint64("task_id", progress.TaskID))
		return nil
	}

	// 只更新运行中的任务
	if task.Status != model.TaskStatusRunning {
		return nil
	}

	// 更新进度
	task.Progress = progress.Progress
	task.CurrentStep = progress.CurrentStep

	if err := c.taskRepo.Update(task); err != nil {
		c.logger.Warn("Failed to update task progress",
			zap.Uint64("task_id", progress.TaskID),
			zap.Error(err))
		return err
	}

	return nil
}

// Start 启动消费者
func (c *ResultConsumer) Start(ctx context.Context, client *RabbitMQClient) error {
	cfg := client.config

	// 消费结果队列
	go client.Consume(ctx, cfg.ResultQueue, c.HandleResult)

	// 消费进度队列
	go client.Consume(ctx, cfg.ProgressQueue, c.HandleProgress)

	c.logger.Info("Result consumer started")
	return nil
}
