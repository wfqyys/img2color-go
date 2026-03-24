package handler

import (
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"img2color-go/api/pkg/errorx"
	"img2color-go/api/pkg/logger"
)

// Middleware 中间件类型
type Middleware func(http.Handler) http.Handler

// Chain 中间件链
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// CORSMiddleware CORS中间件
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置CORS头
			origin := r.Header.Get("Origin")
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) > 0 && allowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Referer")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// 处理预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed 检查Origin是否允许
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// RefererMiddleware Referer验证中间件
func RefererMiddleware(allowedReferers []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果没有配置允许的Referer，则允许所有
			if len(allowedReferers) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			referer := r.Header.Get("Referer")
			if !isRefererAllowed(referer, allowedReferers) {
				logger.Warn("非法Referer: %s", referer)
				errorx.WriteError(w, errorx.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isRefererAllowed 检查Referer是否允许
func isRefererAllowed(referer string, allowedReferers []string) bool {
	if len(allowedReferers) == 0 {
		return true
	}

	for _, allowed := range allowedReferers {
		// 支持通配符匹配
		pattern := strings.ReplaceAll(allowed, ".", "\\.")
		pattern = strings.ReplaceAll(pattern, "*", ".*")

		match, err := regexp.MatchString(pattern, referer)
		if err == nil && match {
			return true
		}
	}

	return false
}

// RateLimiter 速率限制器
type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

// ClientInfo 客户端信息
type ClientInfo struct {
	timestamps []time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(limit int, windowSeconds int) *RateLimiter {
	limiter := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		limit:   limit,
		window:  time.Duration(windowSeconds) * time.Second,
	}

	// 启动清理goroutine
	go limiter.cleanup()

	return limiter
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]
	if !exists {
		client = &ClientInfo{
			timestamps: make([]time.Time, 0),
		}
		rl.clients[clientIP] = client
	}

	// 移除过期的请求记录
	validTimestamps := make([]time.Time, 0)
	for _, ts := range client.timestamps {
		if now.Sub(ts) <= rl.window {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	client.timestamps = validTimestamps

	// 检查是否超过限制
	if len(client.timestamps) >= rl.limit {
		return false
	}

	// 记录本次请求
	client.timestamps = append(client.timestamps, now)
	return true
}

// cleanup 定期清理过期数据
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			validTimestamps := make([]time.Time, 0)
			for _, ts := range client.timestamps {
				if now.Sub(ts) <= rl.window {
					validTimestamps = append(validTimestamps, ts)
				}
			}

			if len(validTimestamps) == 0 {
				delete(rl.clients, ip)
			} else {
				client.timestamps = validTimestamps
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware 速率限制中间件
func RateLimitMiddleware(limiter *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !limiter.Allow(clientIP) {
				logger.Warn("速率限制触发: %s", clientIP)
				errorx.WriteError(w, errorx.ErrRateLimitExceeded)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 尝试从X-Forwarded-For获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 使用RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应包装器
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger.Info("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// responseWriter 响应包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader 写入状态码
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
