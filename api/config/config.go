package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig
	Security SecurityConfig
	Redis    RedisConfig
	MongoDB  MongoDBConfig
	Image    ImageConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	AllowedOrigins  []string
	AllowedReferers []string
	RateLimit       int
	RateWindow      int
}

// RedisConfig Redis配置
type RedisConfig struct {
	Address  string
	Password string
	DB       int
	Enabled  bool
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	URI            string
	Database       string
	Collection     string
	Enabled        bool
}

// ImageConfig 图片处理配置
type ImageConfig struct {
	MaxSize      int64 // 最大图片大小（字节）
	MaxColors    int
	DefaultQuality int
	DefaultIterations int
	DownloadTimeout int // 下载超时（秒）
}

var (
	instance *Config
	once     sync.Once
)

// Load 加载配置
func Load() (*Config, error) {
	var loadErr error
	once.Do(func() {
		instance = &Config{}
		loadErr = instance.loadFromEnv()
	})
	if loadErr != nil {
		return nil, loadErr
	}
	return instance, nil
}

// GetConfig 获取配置单例
func GetConfig() *Config {
	if instance == nil {
		Load()
	}
	return instance
}

// loadFromEnv 从环境变量加载配置
func (c *Config) loadFromEnv() error {
	// 加载.env文件
	if err := c.loadEnvFile(); err != nil {
		// .env文件不存在不是错误，仅打印提示
		fmt.Printf("提示：未加载.env文件（%v），将使用环境变量\n", err)
	}

	// 加载服务器配置
	c.Server.Port = getEnvWithDefault("PORT", "3000")

	// 加载安全配置
	c.Security.AllowedOrigins = parseStringSlice(getEnvWithDefault("ALLOWED_ORIGINS", "*"))
	c.Security.AllowedReferers = parseStringSlice(getEnvWithDefault("ALLOWED_REFERERS", ""))
	c.Security.RateLimit = getEnvIntWithDefault("RATE_LIMIT", 100)
	c.Security.RateWindow = getEnvIntWithDefault("RATE_WINDOW", 60)

	// 加载Redis配置
	c.Redis.Address = getEnvWithDefault("REDIS_ADDRESS", "")
	c.Redis.Password = getEnvWithDefault("REDIS_PASSWORD", "")
	c.Redis.DB = getEnvIntWithDefault("REDIS_DB", 0)
	c.Redis.Enabled = getEnvBool("USE_REDIS_CACHE", false)

	// 加载MongoDB配置
	c.MongoDB.URI = getEnvWithDefault("MONGO_URI", "")
	c.MongoDB.Database = getEnvWithDefault("MONGO_DB", "img2color")
	c.MongoDB.Collection = getEnvWithDefault("MONGO_COLLECTION", "colors")
	c.MongoDB.Enabled = getEnvBool("USE_MONGODB", false)

	// 加载图片配置
	c.Image.MaxSize = getEnvInt64WithDefault("MAX_IMAGE_SIZE", 10*1024*1024) // 默认10MB
	c.Image.MaxColors = getEnvIntWithDefault("MAX_COLORS", 5)
	c.Image.DefaultQuality = getEnvIntWithDefault("DEFAULT_QUALITY", 3)
	c.Image.DefaultIterations = getEnvIntWithDefault("DEFAULT_ITERATIONS", 10)
	c.Image.DownloadTimeout = getEnvIntWithDefault("DOWNLOAD_TIMEOUT", 10)

	return c.validate()
}

// loadEnvFile 加载.env文件
func (c *Config) loadEnvFile() error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}

	envFile := filepath.Join(currentDir, ".env")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return fmt.Errorf(".env文件不存在")
	}

	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("加载.env文件失败: %w", err)
	}

	return nil
}

// validate 验证配置
func (c *Config) validate() error {
	// 验证端口
	if c.Server.Port == "" {
		return fmt.Errorf("PORT不能为空")
	}

	// 验证Redis配置
	if c.Redis.Enabled && c.Redis.Address == "" {
		return fmt.Errorf("启用Redis缓存时REDIS_ADDRESS不能为空")
	}

	// 验证MongoDB配置
	if c.MongoDB.Enabled && c.MongoDB.URI == "" {
		return fmt.Errorf("启用MongoDB时MONGO_URI不能为空")
	}

	// 验证图片大小限制
	if c.Image.MaxSize <= 0 {
		return fmt.Errorf("MAX_IMAGE_SIZE必须大于0")
	}

	return nil
}

// getEnvWithDefault 获取环境变量，带默认值
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvIntWithDefault 获取整数环境变量，带默认值
func getEnvIntWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// getEnvInt64WithDefault 获取int64环境变量，带默认值
func getEnvInt64WithDefault(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// getEnvBool 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true"
}

// parseStringSlice 解析逗号分隔的字符串为切片
func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
