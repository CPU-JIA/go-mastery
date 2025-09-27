package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Logging    LoggingConfig    `mapstructure:"logging"`
	Pagination PaginationConfig `mapstructure:"pagination"`
	Upload     UploadConfig     `mapstructure:"upload"`
	Features   FeatureConfig    `mapstructure:"features"`
	Security   SecurityConfig   `mapstructure:"security"`
	Payment    PaymentConfig    `mapstructure:"payment"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver   string         `mapstructure:"driver"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type SQLiteConfig struct {
	Path string `mapstructure:"path"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	ExpiresIn        time.Duration `mapstructure:"expires_in"`
	RefreshExpiresIn time.Duration `mapstructure:"refresh_expires_in"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type PaginationConfig struct {
	DefaultPageSize int `mapstructure:"default_page_size"`
	MaxPageSize     int `mapstructure:"max_page_size"`
}

type UploadConfig struct {
	MaxSize int64  `mapstructure:"max_size"`
	Path    string `mapstructure:"path"`
}

type FeatureConfig struct {
	RegistrationEnabled bool `mapstructure:"registration_enabled"`
	EmailVerification   bool `mapstructure:"email_verification"`
	OrderNotification   bool `mapstructure:"order_notification"`
	InventoryTracking   bool `mapstructure:"inventory_tracking"`
}

type SecurityConfig struct {
	BcryptCost     int           `mapstructure:"bcrypt_cost"`
	SessionTimeout time.Duration `mapstructure:"session_timeout"`
	RateLimit      struct {
		Requests int           `mapstructure:"requests"`
		Window   time.Duration `mapstructure:"window"`
	} `mapstructure:"rate_limit"`
}

type PaymentConfig struct {
	Currency        string                 `mapstructure:"currency"`
	TaxRate         float64                `mapstructure:"tax_rate"`
	ShippingFee     float64                `mapstructure:"shipping_fee"`
	FreeShippingMin float64                `mapstructure:"free_shipping_min"`
	Gateways        map[string]interface{} `mapstructure:"gateways"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("ECOMMERCE")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
		log.Println("Using default configuration and environment variables")
	} else {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults() {
	// 服务器配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "60s")
	viper.SetDefault("server.write_timeout", "60s")

	// 数据库配置
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.sqlite.path", "ecommerce.db")
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.sslmode", "disable")

	// Redis配置
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// JWT配置
	viper.SetDefault("jwt.secret", "your-super-secret-jwt-key-change-in-production")
	viper.SetDefault("jwt.expires_in", "24h")
	viper.SetDefault("jwt.refresh_expires_in", "168h")

	// 日志配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.file", "logs/app.log")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 30)

	// 分页配置
	viper.SetDefault("pagination.default_page_size", 20)
	viper.SetDefault("pagination.max_page_size", 100)

	// 上传配置
	viper.SetDefault("upload.max_size", 10485760) // 10MB
	viper.SetDefault("upload.path", "uploads/")

	// 功能开关
	viper.SetDefault("features.registration_enabled", true)
	viper.SetDefault("features.email_verification", false)
	viper.SetDefault("features.order_notification", true)
	viper.SetDefault("features.inventory_tracking", true)

	// 安全配置
	viper.SetDefault("security.bcrypt_cost", 12)
	viper.SetDefault("security.session_timeout", "3600s")
	viper.SetDefault("security.rate_limit.requests", 100)
	viper.SetDefault("security.rate_limit.window", "60s")

	// 支付配置
	viper.SetDefault("payment.currency", "CNY")
	viper.SetDefault("payment.tax_rate", 0.13)
	viper.SetDefault("payment.shipping_fee", 15.0)
	viper.SetDefault("payment.free_shipping_min", 99.0)
}
