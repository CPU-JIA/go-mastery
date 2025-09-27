package repository

import (
	"context"
	"file-storage-service/internal/model"
	"io"
	"time"
)

// FileRepository 文件元数据仓储接口
type FileRepository interface {
	Create(ctx context.Context, file *model.File) error
	GetByID(ctx context.Context, id uint) (*model.File, error)
	GetByUUID(ctx context.Context, uuid string) (*model.File, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.File, int64, error)
	Update(ctx context.Context, file *model.File) error
	Delete(ctx context.Context, id uint) error
	GetWithDetails(ctx context.Context, id uint) (*model.File, error)
	Search(ctx context.Context, userID uint, query string, limit, offset int) ([]*model.File, int64, error)
	GetByStatus(ctx context.Context, status model.FileStatus, limit, offset int) ([]*model.File, error)
	UpdateStatus(ctx context.Context, id uint, status model.FileStatus) error
	GetUserStorageStats(ctx context.Context, userID uint) (*model.StorageStats, error)
}

// StorageRepository 物理存储仓储接口
type StorageRepository interface {
	Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
	GetObjectInfo(ctx context.Context, key string) (*ObjectInfo, error)
	GeneratePresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
	UploadMultipart(ctx context.Context, key string, parts []io.Reader, contentType string) error
	ListObjects(ctx context.Context, prefix string, limit int) ([]string, error)
}

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByUUID(ctx context.Context, uuid string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	UpdateStorageUsed(ctx context.Context, userID uint, sizeDelta int64) error
	CheckStorageQuota(ctx context.Context, userID uint, requiredSize int64) (bool, error)
	List(ctx context.Context, limit, offset int) ([]*model.User, int64, error)
	UpdateLastLoginAt(ctx context.Context, userID uint) error
}

// AccessLogRepository 访问日志仓储接口
type AccessLogRepository interface {
	Create(ctx context.Context, log *model.AccessLog) error
	GetByFileID(ctx context.Context, fileID uint, limit, offset int) ([]*model.AccessLog, int64, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.AccessLog, int64, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*model.AccessLog, int64, error)
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*model.AccessLog, int64, error)
	DeleteOldLogs(ctx context.Context, before time.Time) error
}

// UploadTokenRepository 上传令牌仓储接口
type UploadTokenRepository interface {
	Create(ctx context.Context, token *model.UploadToken) error
	GetByToken(ctx context.Context, token string) (*model.UploadToken, error)
	Update(ctx context.Context, uploadToken *model.UploadToken) error
	Delete(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	UpdateUploadedChunks(ctx context.Context, token string, chunkIndex int) error
}

// FileShareRepository 文件分享仓储接口
type FileShareRepository interface {
	Create(ctx context.Context, share *model.FileShare) error
	GetByShareToken(ctx context.Context, token string) (*model.FileShare, error)
	GetByFileID(ctx context.Context, fileID uint) ([]*model.FileShare, error)
	Update(ctx context.Context, share *model.FileShare) error
	Delete(ctx context.Context, id uint) error
	IncrementDownloadCount(ctx context.Context, shareToken string) error
	GetActiveShares(ctx context.Context, limit, offset int) ([]*model.FileShare, int64, error)
}

// FolderRepository 文件夹仓储接口
type FolderRepository interface {
	Create(ctx context.Context, folder *model.Folder) error
	GetByID(ctx context.Context, id uint) (*model.Folder, error)
	GetByUserID(ctx context.Context, userID uint) ([]*model.Folder, error)
	GetByParentID(ctx context.Context, parentID *uint) ([]*model.Folder, error)
	Update(ctx context.Context, folder *model.Folder) error
	Delete(ctx context.Context, id uint) error
	GetWithChildren(ctx context.Context, id uint) (*model.Folder, error)
	Move(ctx context.Context, folderID uint, newParentID *uint) error
}

// ObjectInfo 对象存储信息
type ObjectInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}

// Repositories 仓储聚合器
type Repositories struct {
	File        FileRepository
	Storage     StorageRepository
	User        UserRepository
	AccessLog   AccessLogRepository
	UploadToken UploadTokenRepository
	FileShare   FileShareRepository
	Folder      FolderRepository
}

// ListParams 列表查询参数
type ListParams struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
	Filter map[string]interface{}
}

// SearchParams 搜索查询参数
type SearchParams struct {
	Query     string
	UserID    uint
	Limit     int
	Offset    int
	FileType  string
	DateRange *DateRange
}

// DateRange 日期范围
type DateRange struct {
	StartDate time.Time
	EndDate   time.Time
}
