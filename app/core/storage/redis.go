package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"img2color-go/app/core/pkg/logger"

	"github.com/go-redis/redis/v8"
)

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
}

var (
	redisInstance *RedisCache
	redisOnce     sync.Once
)

// NewRedisCache 创建Redis缓存
func NewRedisCache(address, password string, db int) (Cache, error) {
	var initErr error
	redisOnce.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:     address,
			Password: password,
			DB:       db,
		})

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			initErr = fmt.Errorf("Redis连接失败: %w", err)
			logger.Error("Redis连接失败: %v", err)
			return
		}

		redisInstance = &RedisCache{
			client: client,
		}
		logger.Info("Redis连接成功: %s", address)
	})

	if initErr != nil {
		// 返回空缓存实现作为降级
		logger.Warn("Redis不可用，使用空缓存")
		return NewNopCache(), nil
	}

	if redisInstance == nil {
		return NewNopCache(), nil
	}

	return redisInstance, nil
}

// Get 获取缓存值
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", nil
	}

	result, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 缓存未命中
		return "", nil
	}
	if err != nil {
		logger.Error("Redis获取缓存失败: %v", err)
		return "", err
	}

	return result, nil
}

// Set 设置缓存值
func (c *RedisCache) Set(ctx context.Context, key string, value string) error {
	if c.client == nil {
		return nil
	}

	err := c.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		logger.Error("Redis设置缓存失败: %v", err)
		return err
	}

	return nil
}

// Delete 删除缓存值
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if c.client == nil {
		return nil
	}

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		logger.Error("Redis删除缓存失败: %v", err)
		return err
	}

	return nil
}

// Close 关闭连接
func (c *RedisCache) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}




