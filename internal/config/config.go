// Package config provides centralized configuration management for the Go mastery project.
// This package implements enterprise-grade configuration standards with environment variable
// support, validation, and structured logging integration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Environment represents the deployment environment
type Environment string

const (
	// Development environment for local development
	Development Environment = "development"
	// Testing environment for automated testing
	Testing Environment = "testing"
	// Production environment for production deployment
	Production Environment = "production"
)

// Config holds all configuration values for the application
type Config struct {
	// Environment settings
	Environment Environment `json:"environment"`
	LogLevel    string      `json:"log_level"`

	// Server configuration
	ServerAddress string        `json:"server_address"`
	ServerPort    int           `json:"server_port"`
	ReadTimeout   time.Duration `json:"read_timeout"`
	WriteTimeout  time.Duration `json:"write_timeout"`

	// Database configuration
	DatabaseURL         string `json:"database_url"`
	DatabaseMaxOpenConn int    `json:"database_max_open_conn"`
	DatabaseMaxIdleConn int    `json:"database_max_idle_conn"`

	// Cache configuration
	RedisURL     string        `json:"redis_url"`
	CacheTimeout time.Duration `json:"cache_timeout"`

	// Security configuration
	JWTSecret     string        `json:"jwt_secret"`
	JWTExpiration time.Duration `json:"jwt_expiration"`

	// Performance settings
	WorkerPoolSize int `json:"worker_pool_size"`
	QueueSize      int `json:"queue_size"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Environment:         Development,
		LogLevel:            "info",
		ServerAddress:       "localhost",
		ServerPort:          8080,
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        30 * time.Second,
		DatabaseMaxOpenConn: 25,
		DatabaseMaxIdleConn: 5,
		CacheTimeout:        15 * time.Minute,
		JWTExpiration:       24 * time.Hour,
		WorkerPoolSize:      10,
		QueueSize:           1000,
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := DefaultConfig()

	// Environment settings
	if env := os.Getenv("GO_ENV"); env != "" {
		config.Environment = Environment(env)
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	// Server configuration
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		config.ServerAddress = addr
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.ServerPort = p
		}
	}

	// Database configuration
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}

	// Cache configuration
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		config.RedisURL = redisURL
	}

	// Security configuration
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		config.JWTSecret = jwtSecret
	}

	return config
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ServerPort <= 0 || c.ServerPort > 65535 {
		return fmt.Errorf("invalid server port: %d", c.ServerPort)
	}

	if c.DatabaseMaxOpenConn <= 0 {
		return fmt.Errorf("database max open connections must be positive")
	}

	if c.WorkerPoolSize <= 0 {
		return fmt.Errorf("worker pool size must be positive")
	}

	if c.Environment == Production && c.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required in production environment")
	}

	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == Development
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == Production
}

// GetServerAddr returns the full server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.ServerAddress, c.ServerPort)
}
