package service

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"time"

	"img2color-go/internal/pkg/logger"
	"img2color-go/internal/storage"
)

// ColorExtractor 颜色提取器
type ColorExtractor struct {
	validator    *Validator
	imageService *ImageService
	colorService *ColorService
	cache        storage.Cache
	mongoStorage *storage.MongoDBStorage
}

// NewColorExtractor 创建颜色提取器
func NewColorExtractor(
	maxImageSize int64,
	cache storage.Cache,
	mongoStorage *storage.MongoDBStorage,
) *ColorExtractor {
	return &ColorExtractor{
		validator:    NewValidator(),
		imageService: NewImageService(maxImageSize),
		colorService: NewColorService(),
		cache:        cache,
		mongoStorage: mongoStorage,
	}
}

// Extract 提取图片主色调
func (e *ColorExtractor) Extract(ctx context.Context, imgURL string) (string, error) {
	// 1. 验证URL
	if err := e.validator.ValidateURL(imgURL); err != nil {
		return "", err
	}

	// 2. 计算缓存键
	cacheKey := e.calculateCacheKey(imgURL)

	// 3. 尝试从缓存获取
	cachedColor, err := e.cache.Get(ctx, cacheKey)
	if err == nil && cachedColor != "" {
		logger.Info("缓存命中: %s -> %s", imgURL, cachedColor)
		return cachedColor, nil
	}

	// 4. 下载图片
	imageData, err := e.imageService.Download(ctx, imgURL)
	if err != nil {
		return "", err
	}

	// 5. 提取颜色
	color, err := e.colorService.Extract(imageData)
	if err != nil {
		return "", err
	}

	// 6. 存入缓存
	if err := e.cache.Set(ctx, cacheKey, color); err != nil {
		logger.Warn("缓存存储失败: %v", err)
	}

	// 7. 存入MongoDB
	if e.mongoStorage != nil && e.mongoStorage.IsEnabled() {
		if err := e.mongoStorage.Save(ctx, imgURL, color); err != nil {
			logger.Warn("MongoDB存储失败: %v", err)
		}
	}

	return color, nil
}

// calculateCacheKey 计算缓存键
func (e *ColorExtractor) calculateCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// ExtractWithTimeout 带超时的颜色提取
func (e *ColorExtractor) ExtractWithTimeout(imgURL string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.Extract(ctx, imgURL)
}

