package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	limiter_redis "github.com/ulule/limiter/v3/drivers/store/redis"
	limiter_memory "github.com/ulule/limiter/v3/drivers/store/memory"
	"go.uber.org/zap"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Requests int
	Duration time.Duration
	KeyFunc  func(*gin.Context) string
}

// RateLimiter 限流器
type RateLimiter struct {
	redisLimiter *limiter.Limiter
	localLimiter *limiter.Limiter
	logger       *zap.Logger
}

var globalLimiter *RateLimiter

// InitRateLimiter 初始化限流器
func InitRateLimiter(logger *zap.Logger) error {
	// 创建本地限流器（降级使用）
	localStore := limiter_memory.NewStore()
	localLimiter := limiter.New(localStore, limiter.Rate{
		Period: time.Hour,
		Limit:  int64(20), // 默认本地限流阈值
	})

	var redisLimiter *limiter.Limiter

	// 尝试创建 Redis 限流器
	cfg := config.GetConfig()
	if cfg != nil && cfg.Redis.Host != "" {
		redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
		redisClient := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := redisClient.Ping(ctx).Err(); err != nil {
			logger.Warn("Redis not available for rate limiter, using local limiter", zap.Error(err))
		} else {
			store, err := limiter_redis.NewStoreWithOptions(redisClient, limiter.StoreOptions{
				Prefix: "rate_limit",
			})
			if err != nil {
				logger.Warn("Failed to create Redis rate limiter store, using local limiter", zap.Error(err))
			} else {
				redisLimiter = limiter.New(store, limiter.Rate{
					Period: time.Hour,
					Limit:  int64(20),
				})
				logger.Info("Redis rate limiter initialized", zap.String("addr", redisAddr))
			}
		}
		cancel()
	}

	globalLimiter = &RateLimiter{
		redisLimiter: redisLimiter,
		localLimiter: localLimiter,
		logger:       logger,
	}

	logger.Info("Rate limiter initialized")
	return nil
}

// GetRateLimiter 获取全局限流器
func GetRateLimiter() *RateLimiter {
	return globalLimiter
}

// Limit 限流中间件
func (r *RateLimiter) Limit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取限流 key
		key := config.KeyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		// 尝试 Redis 限流
		if r.redisLimiter != nil {
			r.redisLimiter.Rate = limiter.Rate{
				Period: config.Duration,
				Limit:  int64(config.Requests),
			}

			limiterCtx, err := r.redisLimiter.Get(c, key)
			if err != nil {
				r.logger.Warn("Redis rate limit failed, using local limiter", zap.Error(err))
				// 降级到本地限流
				limiterCtx, _ = r.localLimiter.Get(c, key)
			}

			r.setRateLimitHeaders(c, limiterCtx)
			if limiterCtx.Reached {
				r.logger.Warn("Rate limit exceeded",
					zap.String("key", key),
					zap.Int64("limit", limiterCtx.Limit),
					zap.Int64("remaining", limiterCtx.Remaining))

				r.rateLimitExceeded(c, limiterCtx)
				return
			}
		} else {
			// 本地限流
			r.localLimiter.Rate = limiter.Rate{
				Period: config.Duration,
				Limit:  int64(config.Requests),
			}

			limiterCtx, _ := r.localLimiter.Get(c, key)
			r.setRateLimitHeaders(c, limiterCtx)

			if limiterCtx.Reached {
				r.logger.Warn("Local rate limit exceeded",
					zap.String("key", key))

				r.rateLimitExceeded(c, limiterCtx)
				return
			}
		}

		c.Next()
	}
}

// setRateLimitHeaders 设置限流响应头
func (r *RateLimiter) setRateLimitHeaders(c *gin.Context, limiterCtx limiter.Context) {
	c.Header("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))
}

// rateLimitExceeded 限流响应
func (r *RateLimiter) rateLimitExceeded(c *gin.Context, limiterCtx limiter.Context) {
	retryAfter := limiterCtx.Reset - time.Now().Unix()
	if retryAfter < 0 {
		retryAfter = 0
	}

	c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))

	response.Fail(c, response.CodeTooManyRequests, 
		fmt.Sprintf("请求过于频繁，请%d秒后重试", retryAfter))
	c.Abort()
}

// IPKey 按 IP 限流
func IPKey(c *gin.Context) string {
	return fmt.Sprintf("ip:%s", c.ClientIP())
}

// UserKey 按用户限流
func UserKey(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return "" // 未登录用户不限流
	}
	return fmt.Sprintf("user:%v", userID)
}

// LoginRateLimit 登录限流：5次/分钟/IP
func LoginRateLimit() gin.HandlerFunc {
	if globalLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return globalLimiter.Limit(RateLimitConfig{
		Requests: 5,
		Duration: time.Minute,
		KeyFunc:  IPKey,
	})
}

// AIAnalysisRateLimit AI 分析限流：20次/小时/用户
func AIAnalysisRateLimit() gin.HandlerFunc {
	if globalLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return globalLimiter.Limit(RateLimitConfig{
		Requests: 20,
		Duration: time.Hour,
		KeyFunc:  UserKey,
	})
}

// VectorizeRateLimit 向量化限流：50次/小时/用户
func VectorizeRateLimit() gin.HandlerFunc {
	if globalLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return globalLimiter.Limit(RateLimitConfig{
		Requests: 50,
		Duration: time.Hour,
		KeyFunc:  UserKey,
	})
}

// UploadRateLimit 上传限流：20次/小时/用户
func UploadRateLimit() gin.HandlerFunc {
	if globalLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return globalLimiter.Limit(RateLimitConfig{
		Requests: 20,
		Duration: time.Hour,
		KeyFunc:  UserKey,
	})
}
