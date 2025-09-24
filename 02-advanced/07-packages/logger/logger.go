// Package logger provides structured logging functionality with context support
// and enterprise-grade features including request tracking and JSON output.
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Fields represents key-value pairs for structured logging
type Fields map[string]interface{}

// StructuredLogger interface for enterprise-grade logging with context support
type StructuredLogger interface {
	Debug(msg string, fields Fields)
	Info(msg string, fields Fields)
	Warn(msg string, fields Fields)
	Error(msg string, fields Fields)
	DebugContext(ctx context.Context, msg string, fields Fields)
	InfoContext(ctx context.Context, msg string, fields Fields)
	WarnContext(ctx context.Context, msg string, fields Fields)
	ErrorContext(ctx context.Context, msg string, fields Fields)
	SetLevel(level LogLevel)
	// Migration methods for fmt.Print compatibility
	PrintMigration(args ...interface{})
	PrintlnMigration(args ...interface{})
	PrintfMigration(format string, args ...interface{})
	PrintfMigrationContext(ctx context.Context, format string, args ...interface{})
}

// Logger interface (legacy interface for backward compatibility)
type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	SetLevel(level LogLevel)
}

// structuredLogger enterprise-grade structured logger implementation
type structuredLogger struct {
	level      LogLevel
	output     io.Writer
	format     string
	timeFormat string
}

// simpleLogger legacy simple logger implementation (for backward compatibility)
type simpleLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewStructured creates a new structured logger instance with enhanced capabilities
func NewStructured(config *Config, customOutput ...io.Writer) StructuredLogger {
	var level LogLevel
	switch config.Level {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	default:
		level = INFO
	}

	var output io.Writer = os.Stdout
	if len(customOutput) > 0 {
		output = customOutput[0]
	} else if config.Output == "stderr" {
		output = os.Stderr
	}

	format := config.Format
	if format == "" {
		format = "json"
	}

	timeFormat := config.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	return &structuredLogger{
		level:      level,
		output:     output,
		format:     format,
		timeFormat: timeFormat,
	}
}

// New creates a legacy logger instance for backward compatibility
func New(levelStr string) Logger {
	var level LogLevel
	switch levelStr {
	case "DEBUG":
		level = DEBUG
	case "INFO":
		level = INFO
	case "WARN":
		level = WARN
	case "ERROR":
		level = ERROR
	default:
		level = INFO
	}

	return &simpleLogger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Structured logger method implementations

// Debug logs debug level message with fields
func (l *structuredLogger) Debug(msg string, fields Fields) {
	if l.level <= DEBUG {
		l.logStructured(DEBUG, msg, fields, context.Background())
	}
}

// Info logs info level message with fields
func (l *structuredLogger) Info(msg string, fields Fields) {
	if l.level <= INFO {
		l.logStructured(INFO, msg, fields, context.Background())
	}
}

// Warn logs warning level message with fields
func (l *structuredLogger) Warn(msg string, fields Fields) {
	if l.level <= WARN {
		l.logStructured(WARN, msg, fields, context.Background())
	}
}

// Error logs error level message with fields
func (l *structuredLogger) Error(msg string, fields Fields) {
	if l.level <= ERROR {
		l.logStructured(ERROR, msg, fields, context.Background())
	}
}

// DebugContext logs debug level message with context and fields
func (l *structuredLogger) DebugContext(ctx context.Context, msg string, fields Fields) {
	if l.level <= DEBUG {
		l.logStructured(DEBUG, msg, fields, ctx)
	}
}

// InfoContext logs info level message with context and fields
func (l *structuredLogger) InfoContext(ctx context.Context, msg string, fields Fields) {
	if l.level <= INFO {
		l.logStructured(INFO, msg, fields, ctx)
	}
}

// WarnContext logs warning level message with context and fields
func (l *structuredLogger) WarnContext(ctx context.Context, msg string, fields Fields) {
	if l.level <= WARN {
		l.logStructured(WARN, msg, fields, ctx)
	}
}

// ErrorContext logs error level message with context and fields
func (l *structuredLogger) ErrorContext(ctx context.Context, msg string, fields Fields) {
	if l.level <= ERROR {
		l.logStructured(ERROR, msg, fields, ctx)
	}
}

// SetLevel sets the logging level
func (l *structuredLogger) SetLevel(level LogLevel) {
	l.level = level
}

// logStructured internal method for structured logging
func (l *structuredLogger) logStructured(level LogLevel, msg string, fields Fields, ctx context.Context) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(l.timeFormat),
		"level":     level.String(),
		"message":   msg,
	}

	// Add context information
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		logEntry["request_id"] = requestID
	}

	// Add custom fields
	if fields != nil {
		for k, v := range fields {
			logEntry[k] = v
		}
	}

	var output []byte
	var err error

	if l.format == "json" {
		output, err = json.Marshal(logEntry)
		if err != nil {
			// Fallback to simple format if JSON marshaling fails
			output = []byte(fmt.Sprintf("[%s] [%s] %s\n",
				time.Now().Format(l.timeFormat), level.String(), msg))
		} else {
			output = append(output, '\n')
		}
	} else {
		// Text format
		output = []byte(fmt.Sprintf("[%s] [%s] %s\n",
			time.Now().Format(l.timeFormat), level.String(), msg))
	}

	l.output.Write(output)
}

// Context helper functions

type contextKey string

const requestIDKey contextKey = "request_id"

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestIDFromContext retrieves request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	// Also check for generic request_id key
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// Legacy logger implementations remain unchanged below
func (l *simpleLogger) Debug(msg string) {
	if l.level <= DEBUG {
		l.log(DEBUG, msg)
	}
}

// Info 记录信息级别日志
func (l *simpleLogger) Info(msg string) {
	if l.level <= INFO {
		l.log(INFO, msg)
	}
}

// Warn 记录警告级别日志
func (l *simpleLogger) Warn(msg string) {
	if l.level <= WARN {
		l.log(WARN, msg)
	}
}

// Error 记录错误级别日志
func (l *simpleLogger) Error(msg string) {
	if l.level <= ERROR {
		l.log(ERROR, msg)
	}
}

// SetLevel 设置日志级别
func (l *simpleLogger) SetLevel(level LogLevel) {
	l.level = level
}

// log 内部日志记录方法
func (l *simpleLogger) log(level LogLevel, msg string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMsg := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), msg)
	l.logger.Println(logMsg)
}

// 包级别的便利函数
var defaultLogger = New("INFO")

// Debug 包级别的调试日志函数
func Debug(msg string) {
	defaultLogger.Debug(msg)
}

// Info 包级别的信息日志函数
func Info(msg string) {
	defaultLogger.Info(msg)
}

// Warn 包级别的警告日志函数
func Warn(msg string) {
	defaultLogger.Warn(msg)
}

// Error 包级别的错误日志函数
func Error(msg string) {
	defaultLogger.Error(msg)
}

// SetLevel 设置默认日志器的日志级别
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// Package initialization complete

// Migration methods for fmt.Print compatibility

// PrintMigration migrates fmt.Print calls to structured logging
func (l *structuredLogger) PrintMigration(args ...interface{}) {
	if l.level <= INFO {
		message := fmt.Sprint(args...)
		l.logStructured(INFO, message, Fields{"migration": "fmt.Print"}, context.Background())
	}
}

// PrintlnMigration migrates fmt.Println calls to structured logging
func (l *structuredLogger) PrintlnMigration(args ...interface{}) {
	if l.level <= INFO {
		message := fmt.Sprint(args...)
		l.logStructured(INFO, message, Fields{"migration": "fmt.Println"}, context.Background())
	}
}

// PrintfMigration migrates fmt.Printf calls to structured logging
func (l *structuredLogger) PrintfMigration(format string, args ...interface{}) {
	if l.level <= INFO {
		message := fmt.Sprintf(format, args...)
		l.logStructured(INFO, message, Fields{"migration": "fmt.Printf"}, context.Background())
	}
}

// PrintfMigrationContext migrates fmt.Printf calls with context support
func (l *structuredLogger) PrintfMigrationContext(ctx context.Context, format string, args ...interface{}) {
	if l.level <= INFO {
		message := fmt.Sprintf(format, args...)
		l.logStructured(INFO, message, Fields{"migration": "fmt.Printf"}, ctx)
	}
}
