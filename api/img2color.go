package handler

import (
	"log"
	"net/http"

	"img2color-go/internal/config"
	"img2color-go/internal/handler"
	"img2color-go/internal/pkg/logger"
	"img2color-go/internal/service"
	"img2color-go/internal/storage"
)

var (
	apiHandler    http.Handler
	healthHandler http.Handler
)

func init() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	logger.Info("配置加载成功")

	// 初始化Redis缓存
	var cache storage.Cache
	if cfg.Redis.Enabled {
		cache, err = storage.NewRedisCache(cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			logger.Warn("Redis初始化失败，使用空缓存: %v", err)
			cache = storage.NewNopCache()
		}
	} else {
		cache = storage.NewNopCache()
	}

	// 初始化MongoDB存储
	var mongoStorage *storage.MongoDBStorage
	if cfg.MongoDB.Enabled {
		mongoStorage, err = storage.NewMongoStorage(cfg.MongoDB.URI, cfg.MongoDB.Database, cfg.MongoDB.Collection)
		if err != nil {
			logger.Warn("MongoDB初始化失败: %v", err)
		}
	}

	// 创建颜色提取器
	extractor := service.NewColorExtractor(cfg.Image.MaxSize, cache, mongoStorage)

	// 创建处理器
	h := handler.NewHandler(extractor, cfg.Image.DownloadTimeout)
	hh := handler.NewHealthHandler(cache, mongoStorage)

	// 创建速率限制器
	limiter := handler.NewRateLimiter(cfg.Security.RateLimit, cfg.Security.RateWindow)

	// 创建API处理器中间件链（在init中创建，避免每次请求都创建）
	apiHandler = handler.Chain(
		h,
		handler.LoggingMiddleware(),
		handler.RateLimitMiddleware(limiter),
		handler.RefererMiddleware(cfg.Security.AllowedReferers),
		handler.CORSMiddleware(cfg.Security.AllowedOrigins),
	)

	// 创建健康检查处理器中间件链
	healthHandler = handler.Chain(
		hh,
		handler.LoggingMiddleware(),
		handler.CORSMiddleware(cfg.Security.AllowedOrigins),
	)

	logger.Info("服务初始化完成")
}

// Handler Vercel入口函数 - API接口
func Handler(w http.ResponseWriter, r *http.Request) {
	// 根据路径判断是API还是健康检查
	if r.URL.Path == "/health" {
		healthHandler.ServeHTTP(w, r)
		return
	}
	apiHandler.ServeHTTP(w, r)
}
