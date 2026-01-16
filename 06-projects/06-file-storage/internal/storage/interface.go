package storage

import (
	"context"
	"file-storage-service/internal/models"
	"io"
)

// FileStorage 文件存储接口
type FileStorage interface {
	// 基本文件操作
	Save(ctx context.Context, path string, content io.Reader, size int64) error
	Get(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	Exists(ctx context.Context, path string) (bool, error)
	Size(ctx context.Context, path string) (int64, error)

	// 批量操作
	SaveMultiple(ctx context.Context, files map[string]io.Reader) error
	DeleteMultiple(ctx context.Context, paths []string) error

	// 元数据操作
	GetMetadata(ctx context.Context, path string) (map[string]string, error)
	SetMetadata(ctx context.Context, path string, metadata map[string]string) error

	// 预签名URL（用于直接上传/下载）
	GetPresignedUploadURL(ctx context.Context, path string, expiry int64) (string, error)
	GetPresignedDownloadURL(ctx context.Context, path string, expiry int64) (string, error)

	// 列出文件
	List(ctx context.Context, prefix string, maxKeys int) ([]FileInfo, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// FileInfo 文件信息
type FileInfo struct {
	Path         string
	Size         int64
	LastModified int64
	ETag         string
	Metadata     map[string]string
}

// DatabaseStorage 数据库存储接口
type DatabaseStorage interface {
	// 文件记录管理
	CreateFile(ctx context.Context, file *models.File) error
	GetFileByID(ctx context.Context, id string) (*models.File, error)
	GetFilesByOwner(ctx context.Context, ownerID string, page, perPage int) ([]*models.File, int64, error)
	UpdateFile(ctx context.Context, file *models.File) error
	DeleteFile(ctx context.Context, id string) error
	SearchFiles(ctx context.Context, query string, filters map[string]interface{}, page, perPage int) ([]*models.File, int64, error)

	// 版本管理
	CreateFileVersion(ctx context.Context, version *models.FileVersion) error
	GetFileVersions(ctx context.Context, fileID string) ([]*models.FileVersion, error)

	// 图片信息管理
	CreateImageInfo(ctx context.Context, imageInfo *models.ImageInfo) error
	UpdateImageInfo(ctx context.Context, imageInfo *models.ImageInfo) error
	CreateThumbnail(ctx context.Context, thumbnail *models.Thumbnail) error

	// 上传令牌管理
	CreateUploadToken(ctx context.Context, token *models.UploadToken) error
	GetUploadToken(ctx context.Context, token string) (*models.UploadToken, error)
	UpdateUploadToken(ctx context.Context, token *models.UploadToken) error
	DeleteUploadToken(ctx context.Context, token string) error
	CleanupExpiredTokens(ctx context.Context) error

	// 访问日志
	CreateAccessLog(ctx context.Context, log *models.AccessLog) error
	GetAccessLogs(ctx context.Context, fileID string, limit int) ([]*models.AccessLog, error)

	// 统计信息
	GetFileStats(ctx context.Context) (*models.StatsResponse, error)
	GetStorageUsage(ctx context.Context) (int64, error)

	// 健康检查
	HealthCheck(ctx context.Context) error
}
