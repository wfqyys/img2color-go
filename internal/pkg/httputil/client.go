package httputil

import (
	"net/http"
	"sync"
	"time"
)

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout            time.Duration
	MaxIdleConns       int
	IdleConnTimeout    time.Duration
	DisableCompression bool
}

// DefaultHTTPClientConfig 默认HTTP客户端配置
var DefaultHTTPClientConfig = HTTPClientConfig{
	Timeout:            10 * time.Second,
	MaxIdleConns:       100,
	IdleConnTimeout:    90 * time.Second,
	DisableCompression: false,
}

var (
	httpClient     *http.Client
	httpClientOnce sync.Once
)

// GetHTTPClient 获取HTTP客户端单例
func GetHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		httpClient = createHTTPClient(DefaultHTTPClientConfig)
	})
	return httpClient
}

// GetHTTPClientWithConfig 获取自定义配置的HTTP客户端
func GetHTTPClientWithConfig(config HTTPClientConfig) *http.Client {
	return createHTTPClient(config)
}

// createHTTPClient 创建HTTP客户端
func createHTTPClient(config HTTPClientConfig) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableCompression:  config.DisableCompression,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     20,
	}

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}
}

// SetTimeout 设置超时时间
func SetTimeout(timeout time.Duration) {
	DefaultHTTPClientConfig.Timeout = timeout
	// 重新创建客户端
	httpClientOnce = sync.Once{}
}

// CreateUserAgent 创建User-Agent头
func CreateUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"
}

// CreateRequest 创建HTTP请求
func CreateRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", CreateUserAgent())
	return req, nil
}
