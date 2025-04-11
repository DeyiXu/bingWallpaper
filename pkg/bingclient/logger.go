package bingclient

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel 表示日志级别
type LogLevel int

const (
	// LogLevelDebug 调试信息
	LogLevelDebug LogLevel = iota
	// LogLevelInfo 一般信息
	LogLevelInfo
	// LogLevelWarning 警告信息
	LogLevelWarning
	// LogLevelError 错误信息
	LogLevelError
)

// Logger 接口定义了日志记录器应具备的方法
type Logger interface {
	// Debug 记录调试级别的日志
	Debug(format string, args ...interface{})
	// Info 记录信息级别的日志
	Info(format string, args ...interface{})
	// Warning 记录警告级别的日志
	Warning(format string, args ...interface{})
	// Error 记录错误级别的日志
	Error(format string, args ...interface{})
	// SetLevel 设置日志级别
	SetLevel(level LogLevel)
	// GetLevel 获取当前日志级别
	GetLevel() LogLevel
}

// DefaultLogger 是默认的日志记录器实现
type DefaultLogger struct {
	writer    io.Writer
	level     LogLevel
	showTime  bool
	showLevel bool
}

// LoggerOption 定义日志记录器选项
type LoggerOption func(*DefaultLogger)

// WithWriter 设置日志输出的写入器
func WithWriter(writer io.Writer) LoggerOption {
	return func(l *DefaultLogger) {
		l.writer = writer
	}
}

// WithLevel 设置日志级别
func WithLevel(level LogLevel) LoggerOption {
	return func(l *DefaultLogger) {
		l.level = level
	}
}

// WithTimeDisplay 设置是否显示时间
func WithTimeDisplay(show bool) LoggerOption {
	return func(l *DefaultLogger) {
		l.showTime = show
	}
}

// WithLevelDisplay 设置是否显示级别标签
func WithLevelDisplay(show bool) LoggerOption {
	return func(l *DefaultLogger) {
		l.showLevel = show
	}
}

// NewLogger 创建一个新的默认日志记录器
func NewLogger(options ...LoggerOption) *DefaultLogger {
	// 默认选项
	logger := &DefaultLogger{
		writer:    os.Stdout,
		level:     LogLevelInfo,
		showTime:  true,
		showLevel: true,
	}

	// 应用选项
	for _, option := range options {
		option(logger)
	}

	return logger
}

// formatMessage 格式化日志消息
func (l *DefaultLogger) formatMessage(level LogLevel, format string, args ...interface{}) string {
	var result string

	// 添加时间戳
	if l.showTime {
		now := time.Now().Format("2006-01-02 15:04:05")
		result += fmt.Sprintf("[%s] ", now)
	}

	// 添加日志级别标签
	if l.showLevel {
		var levelTag string
		switch level {
		case LogLevelDebug:
			levelTag = "DEBUG"
		case LogLevelInfo:
			levelTag = "INFO"
		case LogLevelWarning:
			levelTag = "WARN"
		case LogLevelError:
			levelTag = "ERROR"
		}
		result += fmt.Sprintf("[%s] ", levelTag)
	}

	// 添加实际消息
	message := fmt.Sprintf(format, args...)
	result += message

	return result
}

// log 记录日志的内部方法
func (l *DefaultLogger) log(level LogLevel, format string, args ...interface{}) {
	// 检查日志级别是否需要记录
	if level < l.level {
		return
	}

	// 格式化消息
	message := l.formatMessage(level, format, args...)

	// 输出日志
	if l.writer != nil {
		fmt.Fprintln(l.writer, message)

		// 强制刷新（如果是标准输出或标准错误）
		if l.writer == os.Stdout || l.writer == os.Stderr {
			// 在一些系统中可能需要额外的刷新操作，但fmt.Fprintln通常会自动刷新
		}
	}
}

// Debug 实现 Logger 接口
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

// Info 实现 Logger 接口
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

// Warning 实现 Logger 接口
func (l *DefaultLogger) Warning(format string, args ...interface{}) {
	l.log(LogLevelWarning, format, args...)
}

// Error 实现 Logger 接口
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

// SetLevel 实现 Logger 接口
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// GetLevel 实现 Logger 接口
func (l *DefaultLogger) GetLevel() LogLevel {
	return l.level
}

// NullLogger 不输出任何日志的记录器
type NullLogger struct{}

// Debug 实现 Logger 接口
func (l *NullLogger) Debug(format string, args ...interface{}) {}

// Info 实现 Logger 接口
func (l *NullLogger) Info(format string, args ...interface{}) {}

// Warning 实现 Logger 接口
func (l *NullLogger) Warning(format string, args ...interface{}) {}

// Error 实现 Logger 接口
func (l *NullLogger) Error(format string, args ...interface{}) {}

// SetLevel 实现 Logger 接口
func (l *NullLogger) SetLevel(level LogLevel) {}

// GetLevel 实现 Logger 接口
func (l *NullLogger) GetLevel() LogLevel {
	return LogLevelError
}
