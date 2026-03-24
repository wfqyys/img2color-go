package storage

import "context"

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存值
	Get(ctx context.Context, key string) (string, error)
	// Set 设置缓存值
	Set(ctx context.Context, key string, value string) error
	// Delete 删除缓存值
	Delete(ctx context.Context, key string) error
	// Close 关闭连接
	Close() error
}

// NopCache 空缓存实现（用于禁用缓存时）
type NopCache struct{}

// NewNopCache 创建空缓存
func NewNopCache() *NopCache {
	return &NopCache{}
}

// Get 获取缓存值
func (c *NopCache) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}

// Set 设置缓存值
func (c *NopCache) Set(ctx context.Context, key string, value string) error {
	return nil
}

// Delete 删除缓存值
func (c *NopCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Close 关闭连接
func (c *NopCache) Close() error {
	return nil
}


