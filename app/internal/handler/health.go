package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"img2color-go/app/internal/pkg/logger"
	"img2color-go/app/internal/storage"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	cache        storage.Cache
	mongoStorage *storage.MongoDBStorage
	version      string
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(cache storage.Cache, mongoStorage *storage.MongoDBStorage) *HealthHandler {
	return &HealthHandler{
		cache:        cache,
		mongoStorage: mongoStorage,
		version:      "2.0.0",
	}
}

// ServeHTTP 处理健康检查请求
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":       "ok",
		"version":      h.version,
		"timestamp":    time.Now().Format(time.RFC3339),
		"dependencies": h.checkDependencies(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(health); err != nil {
		logger.Error("健康检查响应编码失败: %v", err)
	}
}

// checkDependencies 检查依赖服务状态
func (h *HealthHandler) checkDependencies() map[string]interface{} {
	deps := make(map[string]interface{})

	// 检查Redis
	deps["redis"] = h.checkRedis()

	// 检查MongoDB
	deps["mongodb"] = h.checkMongoDB()

	return deps
}

// checkRedis 检查Redis状态
func (h *HealthHandler) checkRedis() map[string]interface{} {
	status := map[string]interface{}{
		"status": "unknown",
	}

	// 检查是否为空缓存
	if _, ok := h.cache.(*storage.NopCache); ok {
		status["status"] = "disabled"
		return status
	}

	// 尝试Ping
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 尝试设置和获取一个测试值
	testKey := "health:check"
	testValue := "ok"

	if err := h.cache.Set(ctx, testKey, testValue); err != nil {
		status["status"] = "error"
		status["error"] = err.Error()
		return status
	}

	value, err := h.cache.Get(ctx, testKey)
	if err != nil || value != testValue {
		status["status"] = "error"
		if err != nil {
			status["error"] = err.Error()
		}
		return status
	}

	status["status"] = "ok"
	return status
}

// checkMongoDB 检查MongoDB状态
func (h *HealthHandler) checkMongoDB() map[string]interface{} {
	status := map[string]interface{}{
		"status": "unknown",
	}

	if h.mongoStorage == nil || !h.mongoStorage.IsEnabled() {
		status["status"] = "disabled"
		return status
	}

	// MongoDB连接正常
	status["status"] = "ok"
	return status
}



