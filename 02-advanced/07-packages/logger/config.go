package logger

import (
	"fmt"
	"os"
)

// Config 日志配置结构体
type Config struct {
	Level      string `json:"level"`
	Output     string `json:"output"`
	Format     string `json:"format"`
	TimeFormat string `json:"time_format"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:      "INFO",
		Output:     "stdout",
		Format:     "text",
		TimeFormat: "2006-01-02 15:04:05",
	}
}

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv() *Config {
	config := DefaultConfig()

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = level
	}

	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = output
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = format
	}

	if timeFormat := os.Getenv("LOG_TIME_FORMAT"); timeFormat != "" {
		config.TimeFormat = timeFormat
	}

	return config
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	validLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}

	if !validLevels[c.Level] {
		return fmt.Errorf("无效的日志级别: %s", c.Level)
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
	}

	if !validOutputs[c.Output] {
		return fmt.Errorf("无效的输出类型: %s", c.Output)
	}

	return nil
}

// Package initialization complete
