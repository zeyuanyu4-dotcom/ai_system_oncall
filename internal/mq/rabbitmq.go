package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"ai_system_oncall/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// 消息类型
const (
	MessageTypeCommand  = "analysis.command"
	MessageTypeResult   = "analysis.result"
	MessageTypeProgress = "analysis.progress"
)

// AnalysisCommand 分析命令消息
type AnalysisCommand struct {
	TaskID  uint64                 `json:"task_id"`
	IssueID uint64                 `json:"issue_id"`
	Payload map[string]interface{} `json:"payload"`
}

// AnalysisResult 分析结果消息
type AnalysisResult struct {
	TaskID  uint64                 `json:"task_id"`
	IssueID uint64                 `json:"issue_id"`
	Success bool                   `json:"success"`
	Summary string                 `json:"summary"`
	Result  map[string]interface{} `json:"result"`
	Error   string                 `json:"error,omitempty"`
}

// AnalysisProgress 分析进度消息
type AnalysisProgress struct {
	TaskID      uint64 `json:"task_id"`
	Progress    string `json:"progress"`
	CurrentStep string `json:"current_step"`
}

// RabbitMQClient RabbitMQ 客户端
type RabbitMQClient struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	config   *config.RabbitMQConfig
	logger   *zap.Logger
	mu       sync.Mutex
}

var globalClient *RabbitMQClient

// Init 初始化 RabbitMQ 连接
func Init(logger *zap.Logger) error {
	cfg := config.GetConfig()
	if cfg == nil || cfg.RabbitMQ.URL == "" {
		logger.Warn("RabbitMQ config not found, MQ disabled")
		return nil
	}

	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error("Failed to open channel", zap.Error(err))
		return fmt.Errorf("failed to open channel: %w", err)
	}

	client := &RabbitMQClient{
		conn:    conn,
		channel: ch,
		config:  &cfg.RabbitMQ,
		logger:  logger,
	}

	// 声明交换机
	if err := client.declareExchange(); err != nil {
		client.Close()
		return err
	}

	// 声明队列
	if err := client.declareQueues(); err != nil {
		client.Close()
		return err
	}

	globalClient = client
	logger.Info("RabbitMQ initialized successfully",
		zap.String("exchange", cfg.RabbitMQ.Exchange))

	return nil
}

// GetClient 获取全局客户端
func GetClient() *RabbitMQClient {
	return globalClient
}

// IsEnabled 检查是否启用
func (c *RabbitMQClient) IsEnabled() bool {
	return c != nil && c.conn != nil && !c.conn.IsClosed()
}

// declareExchange 声明交换机
func (c *RabbitMQClient) declareExchange() error {
	return c.channel.ExchangeDeclare(
		c.config.Exchange, // name
		"topic",           // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // no-wait
		nil,               // arguments
	)
}

// declareQueues 声明队列并绑定
func (c *RabbitMQClient) declareQueues() error {
	queues := []struct {
		name       string
		routingKey string
	}{
		{c.config.CommandQueue, MessageTypeCommand},
		{c.config.ResultQueue, MessageTypeResult},
		{c.config.ProgressQueue, MessageTypeProgress},
	}

	for _, q := range queues {
		_, err := c.channel.QueueDeclare(
			q.name,
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		// 绑定到交换机
		err = c.channel.QueueBind(
			q.name,
			q.routingKey,
			c.config.Exchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q.name, err)
		}

		c.logger.Debug("Queue declared and bound",
			zap.String("queue", q.name),
			zap.String("routing_key", q.routingKey))
	}

	return nil
}

// PublishCommand 发布分析命令
func (c *RabbitMQClient) PublishCommand(ctx context.Context, cmd *AnalysisCommand) error {
	if !c.IsEnabled() {
		return fmt.Errorf("RabbitMQ not enabled")
	}

	body, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("marshal command failed: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		c.config.Exchange,
		MessageTypeCommand,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)

	if err != nil {
		c.logger.Error("Failed to publish command",
			zap.Uint64("task_id", cmd.TaskID),
			zap.Error(err))
		return err
	}

	c.logger.Info("Command published",
		zap.Uint64("task_id", cmd.TaskID),
		zap.Uint64("issue_id", cmd.IssueID))
	return nil
}

// PublishResult 发布分析结果
func (c *RabbitMQClient) PublishResult(ctx context.Context, result *AnalysisResult) error {
	if !c.IsEnabled() {
		return fmt.Errorf("RabbitMQ not enabled")
	}

	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result failed: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		c.config.Exchange,
		MessageTypeResult,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)

	if err != nil {
		c.logger.Error("Failed to publish result",
			zap.Uint64("task_id", result.TaskID),
			zap.Error(err))
		return err
	}

	c.logger.Info("Result published",
		zap.Uint64("task_id", result.TaskID),
		zap.Bool("success", result.Success))
	return nil
}

// PublishProgress 发布进度更新
func (c *RabbitMQClient) PublishProgress(ctx context.Context, progress *AnalysisProgress) error {
	if !c.IsEnabled() {
		return fmt.Errorf("RabbitMQ not enabled")
	}

	body, err := json.Marshal(progress)
	if err != nil {
		return fmt.Errorf("marshal progress failed: %w", err)
	}

	// 进度消息非持久化
	err = c.channel.PublishWithContext(
		ctx,
		c.config.Exchange,
		MessageTypeProgress,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Transient, // 非持久化
			ContentType:  "application/json",
			Body:         body,
		},
	)

	if err != nil {
		c.logger.Warn("Failed to publish progress",
			zap.Uint64("task_id", progress.TaskID),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Progress published",
		zap.Uint64("task_id", progress.TaskID),
		zap.String("step", progress.CurrentStep))
	return nil
}

// Consume 消费消息
func (c *RabbitMQClient) Consume(ctx context.Context, queue string, handler func([]byte) error) error {
	if !c.IsEnabled() {
		return fmt.Errorf("RabbitMQ not enabled")
	}

	msgs, err := c.channel.Consume(
		queue,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume queue %s: %w", queue, err)
	}

	c.logger.Info("Started consuming queue", zap.String("queue", queue))

	go func() {
		for {
			select {
			case <-ctx.Done():
				c.logger.Info("Consumer stopped", zap.String("queue", queue))
				return
			case msg, ok := <-msgs:
				if !ok {
					c.logger.Warn("Message channel closed", zap.String("queue", queue))
					return
				}

				if err := handler(msg.Body); err != nil {
					c.logger.Error("Failed to handle message",
						zap.String("queue", queue),
						zap.Error(err))
					// Nack 消息，重新入队
					msg.Nack(false, true)
				} else {
					// Ack 消息
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

// Close 关闭连接
func (c *RabbitMQClient) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.logger.Info("RabbitMQ connection closed")
	return nil
}
