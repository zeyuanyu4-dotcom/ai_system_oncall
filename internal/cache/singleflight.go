package cache

import (
	"context"
	"encoding/json"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// SingleflightCache 单飞缓存，防止缓存击穿
type SingleflightCache struct {
	cache *Cache
	group singleflight.Group
	logger *zap.Logger
}

// NewSingleflightCache 创建单飞缓存
func NewSingleflightCache(cache *Cache, logger *zap.Logger) *SingleflightCache {
	return &SingleflightCache{
		cache:  cache,
		logger: logger,
	}
}

// GetWithLoad 获取缓存，如果不存在则通过 loader 加载
// loader 函数负责从数据库加载数据
func (s *SingleflightCache) GetWithLoad(
	ctx context.Context,
	key cacheKey,
	dest interface{},
	keyParts []interface{},
	loader func() (interface{}, error),
) error {
	// 如果缓存未启用，直接从数据库加载
	if s.cache == nil || !s.cache.IsEnabled() {
		return s.loadDirectly(loader, dest)
	}

	fullKey := s.cache.buildKey(key, keyParts...)

	// 使用 singleflight 防止击穿
	result, err, shared := s.group.Do(fullKey, func() (interface{}, error) {
		// 双重检查：在 singleflight 内部再次检查缓存
		// 防止其他 goroutine 已经设置了缓存
		var tempData json.RawMessage
		if err := s.cache.client.Get(ctx, fullKey).Scan(&tempData); err == nil {
			s.logger.Debug("Cache hit in singleflight", zap.String("key", fullKey))
			return tempData, nil
		}

		// 从数据库加载
		s.logger.Debug("Loading from database in singleflight", zap.String("key", fullKey))
		data, err := loader()
		if err != nil {
			return nil, err
		}

		// 处理空值（防穿透）
		if data == nil {
			// 如果 skipCache 为 true，不缓存空值
			if !key.skipCache {
				s.cache.SetNull(ctx, key, keyParts...)
			}
			return []byte("null"), nil
		}

		// 如果 skipCache 为 true，不回填缓存
		if !key.skipCache {
			// 设置缓存
			s.cache.Set(ctx, key, data, keyParts...)
		}

		// 序列化返回
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		return jsonData, nil
	})

	if err != nil {
		return err
	}

	// 反序列化到目标
	if jsonData, ok := result.([]byte); ok {
		if string(jsonData) == "null" {
			return nil // 空值
		}
		return json.Unmarshal(jsonData, dest)
	}

	// 直接返回加载的数据
	if data, ok := result.(interface{}); ok && data != nil {
		// 复制数据到 dest
		return s.copyData(data, dest)
	}

	return nil
}

// loadDirectly 直接从数据库加载
func (s *SingleflightCache) loadDirectly(loader func() (interface{}, error), dest interface{}) error {
	data, err := loader()
	if err != nil {
		return err
	}

	return s.copyData(data, dest)
}

// copyData 复制数据
func (s *SingleflightCache) copyData(src, dest interface{}) error {
	if src == nil {
		return nil
	}

	// 通过 JSON 序列化/反序列化复制
	jsonData, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, dest)
}

// Invalidate 使缓存失效
func (s *SingleflightCache) Invalidate(ctx context.Context, key cacheKey, keyParts ...interface{}) error {
	if s.cache == nil || !s.cache.IsEnabled() {
		return nil
	}

	// 删除缓存
	err := s.cache.Delete(ctx, key, keyParts...)
	if err != nil {
		s.logger.Warn("Failed to invalidate cache",
			zap.String("key", s.cache.buildKey(key, keyParts...)),
			zap.Error(err))
		return err
	}

	s.logger.Debug("Cache invalidated", zap.String("key", s.cache.buildKey(key, keyParts...)))
	return nil
}

// InvalidateByPattern 按模式使缓存失效
func (s *SingleflightCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	if s.cache == nil || !s.cache.IsEnabled() {
		return nil
	}

	return s.cache.DeleteByPattern(ctx, pattern)
}

// 全局 Singleflight 缓存实例
var globalSF *SingleflightCache
var sfOnce sync.Once

// GetSingleflightCache 获取全局 Singleflight 缓存实例
func GetSingleflightCache() *SingleflightCache {
	sfOnce.Do(func() {
		globalSF = NewSingleflightCache(globalCache, globalCache.logger)
	})
	return globalSF
}
