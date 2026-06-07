package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CancelManager 取消管理器
// 通过 Redis Pub/Sub 实现跨进程取消通知
type CancelManager struct {
	client *redis.Client
	logger *zap.Logger

	// 本地取消状态缓存
	mu       sync.RWMutex
	cancelled map[uint64]bool
}

// NewCancelManager 创建取消管理器
func NewCancelManager(client *redis.Client, logger *zap.Logger) *CancelManager {
	return &CancelManager{
		client:    client,
		logger:    logger,
		cancelled: make(map[uint64]bool),
	}
}

// PublishCancel 发布取消信号
func (m *CancelManager) PublishCancel(ctx context.Context, taskID uint64) error {
	// 更新本地状态
	m.mu.Lock()
	m.cancelled[taskID] = true
	m.mu.Unlock()

	// 发布到 Redis
	err := m.client.Publish(ctx, CancelChannel, taskID).Err()
	if err != nil {
		m.logger.Error("Failed to publish cancel signal",
			zap.Uint64("task_id", taskID),
			zap.Error(err))
		return err
	}

	m.logger.Info("Cancel signal published", zap.Uint64("task_id", taskID))
	return nil
}

// Subscribe 取消订阅
func (m *CancelManager) Subscribe(ctx context.Context) error {
	pubsub := m.client.Subscribe(ctx, CancelChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	m.logger.Info("Cancel manager subscribed", zap.String("channel", CancelChannel))

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Cancel manager subscription stopped")
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return nil
			}

			var taskID uint64
			if _, err := fmt.Sscanf(msg.Payload, "%d", &taskID); err != nil {
				m.logger.Warn("Failed to parse cancel message",
					zap.String("payload", msg.Payload),
					zap.Error(err))
				continue
			}

			m.mu.Lock()
			m.cancelled[taskID] = true
			m.mu.Unlock()

			m.logger.Debug("Received cancel signal", zap.Uint64("task_id", taskID))
		}
	}
}

// IsCancelled 检查任务是否被取消
func (m *CancelManager) IsCancelled(taskID uint64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cancelled[taskID]
}

// ClearCancelled 清除取消状态
func (m *CancelManager) ClearCancelled(taskID uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cancelled, taskID)
}
