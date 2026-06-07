package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"ai_system_oncall/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Cache Redis 缓存客户端
type Cache struct {
	client *redis.Client
	logger *zap.Logger
}

// cacheKey 缓存键结构
type cacheKey struct {
	prefix    string
	ttl       time.Duration
	skipCache bool // 是否跳过缓存回填（用于不适合缓存的数据）
}

// 预定义的缓存键
var (
	// 项目相关
	KeyProjectDetail = cacheKey{prefix: "project:detail", ttl: 5 * time.Minute}
	KeyProjectList   = cacheKey{prefix: "project:list", ttl: 5 * time.Minute}
	KeyUserProjects  = cacheKey{prefix: "project:user", ttl: 5 * time.Minute}

	// 服务相关
	KeyServiceDetail      = cacheKey{prefix: "service:detail", ttl: 5 * time.Minute}
	KeyServiceList        = cacheKey{prefix: "service:list", ttl: 5 * time.Minute}
	KeyProjectServices    = cacheKey{prefix: "service:project", ttl: 5 * time.Minute}
	KeyServiceByCode      = cacheKey{prefix: "service:code", ttl: 5 * time.Minute}

	// 知识文档相关
	KeyKnowledgeDocDetail = cacheKey{prefix: "knowledge:detail", ttl: 10 * time.Minute}
	KeyKnowledgeDocList   = cacheKey{prefix: "knowledge:list", ttl: 5 * time.Minute}
	KeyKnowledgeDocSearch = cacheKey{prefix: "knowledge:search", ttl: 3 * time.Minute}

	// 报告相关
	KeyDailyReport   = cacheKey{prefix: "report:daily", ttl: 30 * time.Minute}
	KeyWeeklyReport  = cacheKey{prefix: "report:weekly", ttl: 1 * time.Hour}
	KeyReportDetail  = cacheKey{prefix: "report:detail", ttl: 30 * time.Minute}
	KeyReportList    = cacheKey{prefix: "report:list", ttl: 5 * time.Minute}

	// Dashboard 相关
	KeyDashboardStats = cacheKey{prefix: "dashboard:stats", ttl: 1 * time.Minute}
	KeyDashboardTrend = cacheKey{prefix: "dashboard:trend", ttl: 5 * time.Minute}
)

var globalCache *Cache

// Init 初始化 Redis 缓存
func Init(logger *zap.Logger) error {
	cfg := config.GetConfig()
	if cfg == nil || cfg.Redis.Host == "" {
		logger.Warn("Redis config not found, cache disabled")
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	globalCache = &Cache{
		client: client,
		logger: logger,
	}

	logger.Info("Redis cache initialized successfully",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
		zap.Int("db", cfg.Redis.DB))

	return nil
}

// GetCache 获取全局缓存实例
func GetCache() *Cache {
	return globalCache
}

// IsEnabled 检查缓存是否启用
func (c *Cache) IsEnabled() bool {
	return c != nil && c.client != nil
}

// buildKey 构建完整的缓存键，包含前缀和 TTL 抖动
func (c *Cache) buildKey(key cacheKey, parts ...interface{}) string {
	keyStr := key.prefix
	for _, part := range parts {
		keyStr += fmt.Sprintf(":%v", part)
	}
	return keyStr
}

// getTTLWithJitter 获取带抖动的 TTL
func (c *Cache) getTTLWithJitter(baseTTL time.Duration) time.Duration {
	// TTL 抖动：±10%~20%
	jitterPercent := 0.1 + rand.Float64()*0.1 // 10%~20%
	jitter := time.Duration(float64(baseTTL) * jitterPercent)
	if rand.Intn(2) == 0 {
		return baseTTL - jitter
	}
	return baseTTL + jitter
}

// Get 获取缓存
func (c *Cache) Get(ctx context.Context, key cacheKey, dest interface{}, keyParts ...interface{}) error {
	if !c.IsEnabled() {
		return redis.Nil
	}

	fullKey := c.buildKey(key, keyParts...)
	c.logger.Debug("Getting cache", zap.String("key", fullKey))

	val, err := c.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err != redis.Nil {
			c.logger.Warn("Failed to get cache",
				zap.String("key", fullKey),
				zap.Error(err))
		}
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Warn("Failed to unmarshal cache value",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cache hit", zap.String("key", fullKey))
	return nil
}

// Set 设置缓存
func (c *Cache) Set(ctx context.Context, key cacheKey, value interface{}, keyParts ...interface{}) error {
	if !c.IsEnabled() {
		return nil
	}

	fullKey := c.buildKey(key, keyParts...)
	ttl := c.getTTLWithJitter(key.ttl)

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Warn("Failed to marshal cache value",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		c.logger.Warn("Failed to set cache",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cache set",
		zap.String("key", fullKey),
		zap.Duration("ttl", ttl))
	return nil
}

// SetWithTTL 设置缓存（自定义 TTL）
func (c *Cache) SetWithTTL(ctx context.Context, key cacheKey, value interface{}, ttl time.Duration, keyParts ...interface{}) error {
	if !c.IsEnabled() {
		return nil
	}

	fullKey := c.buildKey(key, keyParts...)
	actualTTL := c.getTTLWithJitter(ttl)

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Warn("Failed to marshal cache value",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	if err := c.client.Set(ctx, fullKey, data, actualTTL).Err(); err != nil {
		c.logger.Warn("Failed to set cache",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cache set with custom TTL",
		zap.String("key", fullKey),
		zap.Duration("ttl", actualTTL))
	return nil
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, key cacheKey, keyParts ...interface{}) error {
	if !c.IsEnabled() {
		return nil
	}

	fullKey := c.buildKey(key, keyParts...)
	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		c.logger.Warn("Failed to delete cache",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Cache deleted", zap.String("key", fullKey))
	return nil
}

// DeleteByPattern 按模式删除缓存
func (c *Cache) DeleteByPattern(ctx context.Context, pattern string) error {
	if !c.IsEnabled() {
		return nil
	}

	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Warn("Failed to find keys by pattern",
			zap.String("pattern", pattern),
			zap.Error(err))
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Warn("Failed to delete keys by pattern",
			zap.String("pattern", pattern),
			zap.Int("count", len(keys)),
			zap.Error(err))
		return err
	}

	c.logger.Info("Cache keys deleted by pattern",
		zap.String("pattern", pattern),
		zap.Int("count", len(keys)))
	return nil
}

// SetNull 设置空值（防穿透）
func (c *Cache) SetNull(ctx context.Context, key cacheKey, keyParts ...interface{}) error {
	if !c.IsEnabled() {
		return nil
	}

	fullKey := c.buildKey(key, keyParts...)
	ttl := c.getTTLWithJitter(time.Minute) // 空值缓存较短

	if err := c.client.Set(ctx, fullKey, "null", ttl).Err(); err != nil {
		c.logger.Warn("Failed to set null cache",
			zap.String("key", fullKey),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Null cache set", zap.String("key", fullKey))
	return nil
}

// IsNull 检查是否是空值
func (c *Cache) IsNull(data []byte) bool {
	return string(data) == "null"
}

// Close 关闭连接
func (c *Cache) Close() error {
	if c != nil && c.client != nil {
		return c.client.Close()
	}
	return nil
}
