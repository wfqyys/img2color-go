package main

import (
	"log"
	"net/http"

	"img2color-go/api/config"
	"img2color-go/api/handler"
	"img2color-go/api/pkg/logger"
	"img2color-go/api/service"
	"img2color-go/api/storage"
)

func main() {
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
	apiHandler := handler.NewHandler(extractor, cfg.Image.DownloadTimeout)
	healthHandler := handler.NewHealthHandler(cache, mongoStorage)

	// 创建速率限制器
	limiter := handler.NewRateLimiter(cfg.Security.RateLimit, cfg.Security.RateWindow)

	// 创建API处理器中间件链
	apiChain := handler.Chain(
		apiHandler,
		handler.LoggingMiddleware(),
		handler.RateLimitMiddleware(limiter),
		handler.RefererMiddleware(cfg.Security.AllowedReferers),
		handler.CORSMiddleware(cfg.Security.AllowedOrigins),
	)

	// 创建健康检查处理器中间件链
	healthChain := handler.Chain(
		healthHandler,
		handler.LoggingMiddleware(),
		handler.CORSMiddleware(cfg.Security.AllowedOrigins),
	)

	// 注册路由
	http.Handle("/api", apiChain)
	http.Handle("/health", healthChain)

	// 获取端口
	port := cfg.Server.Port
	if port == "" {
		port = "3000"
	}

	logger.Info("服务器启动，监听端口: %s", port)
	log.Printf("服务器监听在 :%s...\n", port)

	// 启动服务器
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
