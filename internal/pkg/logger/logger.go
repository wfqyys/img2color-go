package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var (
	levelNames = map[LogLevel]string{
		DebugLevel: "DEBUG",
		InfoLevel:  "INFO",
		WarnLevel:  "WARN",
		ErrorLevel: "ERROR",
	}
)

// Logger 日志记录器
type Logger struct {
	logger *log.Logger
	level  LogLevel
	mu     sync.Mutex
}

var (
	instance *Logger
	once     sync.Once
)

// GetLogger 获取日志记录器单例
func GetLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			logger: log.New(os.Stdout, "", 0),
			level:  InfoLevel,
		}
	})
	return instance
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// log 写入日志
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Printf("[%s] %s: %s", timestamp, levelName, message)
}

// Debug 输出Debug级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info 输出Info级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn 输出Warn级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error 输出Error级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// 全局便捷函数

// Debug 输出Debug级别日志
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

// Info 输出Info级别日志
func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

// Warn 输出Warn级别日志
func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

// Error 输出Error级别日志
func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	GetLogger().SetLevel(level)
}
