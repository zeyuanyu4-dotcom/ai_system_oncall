package task

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// 任务类型定义
const (
	TypeAIAnalysis = "ai:analysis"
)

// 取消信号频道
const CancelChannel = "task:cancel"

// AIAnalysisPayload AI 分析任务载荷
type AIAnalysisPayload struct {
	TaskID  uint64 `json:"task_id"`
	IssueID uint64 `json:"issue_id"`
}

// NewAIAnalysisTask 创建 AI 分析任务
func NewAIAnalysisTask(taskID, issueID uint64) (*asynq.Task, error) {
	payload, err := json.Marshal(AIAnalysisPayload{
		TaskID:  taskID,
		IssueID: issueID,
	})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeAIAnalysis, payload), nil
}

// TaskProducer 任务生产者
type TaskProducer struct {
	client *asynq.Client
}

// NewTaskProducer 创建任务生产者
func NewTaskProducer(redisAddr string) *TaskProducer {
	return &TaskProducer{
		client: asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
	}
}

// EnqueueAIAnalysis 入队 AI 分析任务
func (p *TaskProducer) EnqueueAIAnalysis(taskID, issueID uint64, retryLimit int) (*asynq.TaskInfo, error) {
	task, err := NewAIAnalysisTask(taskID, issueID)
	if err != nil {
		return nil, fmt.Errorf("create task failed: %w", err)
	}

	info, err := p.client.Enqueue(task,
		asynq.MaxRetry(retryLimit),
		asynq.Timeout(0), // 超时由 Worker 控制
		asynq.Retention(0),
	)
	if err != nil {
		return nil, fmt.Errorf("enqueue task failed: %w", err)
	}

	return info, nil
}

// Close 关闭生产者
func (p *TaskProducer) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}
