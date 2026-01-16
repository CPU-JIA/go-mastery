package config

import (
	"os"
	"strconv"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	Upload   UploadConfig   `yaml:"upload"`
	Security SecurityConfig `yaml:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Provider    string `yaml:"provider"`
	LocalPath   string `yaml:"local_path"`
	MinIOConfig `yaml:"minio"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	BucketName string `yaml:"bucket_name"`
	UseSSL     bool   `yaml:"use_ssl"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxSize      int64    `yaml:"max_size"`
	AllowedTypes []string `yaml:"allowed_types"`
	TempDir      string   `yaml:"temp_dir"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EncryptionKey string `yaml:"encryption_key"`
	JWTSecret     string `yaml:"jwt_secret"`
}

// Load 加载配置
func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("HOST", "localhost"),
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("MODE", "development"),
		},
		Database: DatabaseConfig{
			Driver: getEnv("DB_DRIVER", "sqlite"),
			DSN:    getEnv("DB_DSN", "file_storage.db"),
		},
		Storage: StorageConfig{
			Provider:  getEnv("STORAGE_PROVIDER", "local"),
			LocalPath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
			MinIOConfig: MinIOConfig{
				Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
				AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
				SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
				BucketName: getEnv("MINIO_BUCKET", "files"),
				UseSSL:     getEnvBool("MINIO_USE_SSL", false),
			},
		},
		Upload: UploadConfig{
			MaxSize:      getEnvInt64("UPLOAD_MAX_SIZE", 100*1024*1024), // 100MB
			AllowedTypes: getEnvSlice("UPLOAD_ALLOWED_TYPES", []string{"image/*", "application/pdf", "text/*"}),
			TempDir:      getEnv("UPLOAD_TEMP_DIR", "./temp"),
		},
		Security: SecurityConfig{
			EncryptionKey: getEnv("ENCRYPTION_KEY", "myverystrongpasswordo32bitlength"),
			JWTSecret:     getEnv("JWT_SECRET", "jwt-secret-key"),
		},
	}, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool 获取布尔类型环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

// getEnvInt64 获取整数类型环境变量
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

// getEnvSlice 获取字符串切片类型环境变量
func getEnvSlice(key string, defaultValue []string) []string {
	// 简化实现，生产环境可以用更复杂的解析
	return defaultValue
}
