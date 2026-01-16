package storage

import (
	"file-storage-service/internal/config"
	"fmt"
)

// NewFileStorage 根据配置创建文件存储实例
func NewFileStorage(config config.StorageConfig) (FileStorage, error) {
	switch config.Provider {
	case "local":
		return NewLocalStorage(config)
	case "minio":
		return NewMinIOStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage provider: %s", config.Provider)
	}
}

// NewDatabaseStorage 根据配置创建数据库存储实例
func NewDatabaseStorage(config config.DatabaseConfig) (DatabaseStorage, error) {
	switch config.Driver {
	case "sqlite":
		return NewGormStorage(config)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}
