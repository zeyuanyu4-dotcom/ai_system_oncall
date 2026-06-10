package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai_system_oncall/internal/client"
	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/database"
	"ai_system_oncall/internal/grpcclient"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/mq"
	"ai_system_oncall/internal/repository"
	"ai_system_oncall/internal/task"
	"ai_system_oncall/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "config file path")
}

func main() {
	flag.Parse()

	// 加载配置
	if err := config.Init(configPath); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	cfg := config.GetConfig()

	// 初始化日志
	if err := logger.Init(&logger.LogConfig{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化数据库
	if err := database.Init(&cfg.Database); err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 初始化 Redis 客户端（用于取消订阅）
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 测试 Redis 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	cancel()

	// 初始化 RabbitMQ
	if err := mq.Init(zap.L()); err != nil {
		logger.Warnf("RabbitMQ initialization failed (running in HTTP-only mode): %v", err)
	}
	mqClient := mq.GetClient()
	defer func() {
		if mqClient != nil {
			mqClient.Close()
		}
	}()

	// 初始化取消管理器
	cancelMgr := task.NewCancelManager(redisClient, zap.L())

	// 启动取消订阅（后台运行）
	cancelCtx, cancelCancel := context.WithCancel(context.Background())
	go func() {
		if err := cancelMgr.Subscribe(cancelCtx); err != nil && err != context.Canceled {
			logger.Warnf("Cancel subscription error: %v", err)
		}
	}()

	// 初始化 Repository
	taskRepo := repository.NewAIAnalysisTaskRepository(database.GetDB())
	issueRepo := repository.NewIssueRepository(database.GetDB())

	// 初始化 AI 客户端（HTTP 兜底）
	aiClient := client.NewAIClient(&cfg.AI)

	// 初始化 gRPC Agent 客户端（优先走 gRPC streaming）
	var grpcAgent *grpcclient.AgentClient
	if cfg.AI.GRPCAddr != "" {
		var err error
		grpcAgent, err = grpcclient.NewAgentClient(cfg.AI.GRPCAddr, cfg.AI.GRPCTimeout)
		if err != nil {
			logger.Warnf("Failed to connect to Agent gRPC, fallback to HTTP: %v", err)
		} else {
			logger.Infof("Agent gRPC client connected to %s", cfg.AI.GRPCAddr)
		}
	}

	// 创建任务处理器
	handler := task.NewAIAnalysisHandler(
		taskRepo,
		issueRepo,
		aiClient,
		grpcAgent,
		mqClient,
		cancelMgr,
		zap.L(),
		cfg.Asynq.Timeout,
	)

	// 启动 MQ 结果消费者（如果 MQ 启用）
	if mqClient != nil && mqClient.IsEnabled() {
		resultConsumer := mq.NewResultConsumer(taskRepo, issueRepo, zap.L())
		resultConsumer.Start(context.Background(), mqClient)
		logger.Info("MQ result consumer started")
	}

	// 创建 Asynq Server
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Asynq.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Asynq.Concurrency,
			Queues: map[string]int{
				"default":  6,
				"critical": 10,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logger.Errorf("Task failed: %v, task type: %s", err, task.Type())
			}),
			Logger: &asynqLogger{logger: zap.L()},
		},
	)

	// 注册处理器
	mux := asynq.NewServeMux()
	mux.HandleFunc(task.TypeAIAnalysis, handler.ProcessTask)

	// 扫描并重新入队未完成的任务
	logger.Info("Scanning pending tasks...")
	if err := recoverPendingTasks(taskRepo); err != nil {
		logger.Warnf("Failed to recover pending tasks: %v", err)
	}

	// 启动 Worker
	logger.Infof("Starting Asynq worker on %s with concurrency %d",
		cfg.Asynq.RedisAddr, cfg.Asynq.Concurrency)

	if err := srv.Start(mux); err != nil {
		logger.Fatalf("Failed to start worker: %v", err)
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")

	// 优雅关闭
	srv.Shutdown()
	cancelCancel()
	if grpcAgent != nil {
		grpcAgent.Close()
	}

	logger.Info("Worker exited")
}

// recoverPendingTasks 恢复未完成的任务
func recoverPendingTasks(taskRepo *repository.AIAnalysisTaskRepository) error {
	// 查找所有 pending 和 running 状态的任务
	// running 状态的任务可能是服务重启前未完成的
	var tasks []model.AIAnalysisTask
	if err := database.GetDB().Where("status IN ?", []string{
		model.TaskStatusPending,
		model.TaskStatusRunning,
	}).Find(&tasks).Error; err != nil {
		return err
	}

	if len(tasks) == 0 {
		logger.Info("No pending tasks to recover")
		return nil
	}

	// 创建 Producer 用于重新入队
	cfg := config.GetConfig()
	producer := task.NewTaskProducer(cfg.Asynq.RedisAddr)
	defer producer.Close()

	for _, t := range tasks {
		// running 状态的任务重置为 pending
		if t.Status == model.TaskStatusRunning {
			t.Status = model.TaskStatusPending
			t.StartedAt = nil
			database.GetDB().Save(&t)
			logger.Infof("Reset running task %d to pending", t.ID)
		}

		// 重新入队
		// 注意：原 user JWT 不会落库（也不应该落），恢复时无法还原。
		// 传空串让任务继续跑——LLM 主体分析能完成，但内部工具调用会 401。
		// 如需严格"无 token 不跑"，把 userToken 改成一个 sentinel 并在 worker 端拒绝。
		_, err := producer.EnqueueAIAnalysis(t.ID, t.IssueID, "", cfg.Asynq.RetryLimit)
		if err != nil {
			logger.Errorf("Failed to re-enqueue task %d: %v", t.ID, err)
			continue
		}

		logger.Infof("Re-enqueued task %d (issue: %d)", t.ID, t.IssueID)
	}

	logger.Infof("Recovered %d tasks", len(tasks))
	return nil
}

// asynqLogger 适配 zap 到 asynq.Logger 接口
type asynqLogger struct {
	logger *zap.Logger
}

func (l *asynqLogger) Debug(args ...interface{}) {
	l.logger.Sugar().Debug(args...)
}

func (l *asynqLogger) Info(args ...interface{}) {
	l.logger.Sugar().Info(args...)
}

func (l *asynqLogger) Warn(args ...interface{}) {
	l.logger.Sugar().Warn(args...)
}

func (l *asynqLogger) Error(args ...interface{}) {
	l.logger.Sugar().Error(args...)
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	l.logger.Sugar().Fatal(args...)
}
