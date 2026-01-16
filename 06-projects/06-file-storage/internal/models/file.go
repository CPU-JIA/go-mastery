package models

import (
	"gorm.io/gorm"
	"time"
)

// File 文件模型
type File struct {
	ID           string `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"not null"`
	OriginalName string `json:"original_name" gorm:"not null"`
	Path         string `json:"path" gorm:"not null"`
	URL          string `json:"url"`
	Size         int64  `json:"size" gorm:"not null"`
	MimeType     string `json:"mime_type"`
	Extension    string `json:"extension"`
	Checksum     string `json:"checksum" gorm:"index"`

	// 所有者和权限
	OwnerID    string `json:"owner_id" gorm:"index"`
	Visibility string `json:"visibility" gorm:"default:private"` // public, private, shared

	// 元数据
	Metadata map[string]interface{} `json:"metadata" gorm:"serializer:json"`
	Tags     []string               `json:"tags" gorm:"serializer:json"`

	// 版本控制
	Version  int           `json:"version" gorm:"default:1"`
	Versions []FileVersion `json:"versions" gorm:"foreignKey:FileID"`

	// 处理状态
	Status         string                 `json:"status" gorm:"default:processing"` // uploading, processing, ready, error
	ProcessingInfo map[string]interface{} `json:"processing_info" gorm:"serializer:json"`

	// 时间戳
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	AccessedAt *time.Time     `json:"accessed_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// 图片信息（如果是图片）
	ImageInfo *ImageInfo `json:"image_info,omitempty" gorm:"foreignKey:FileID"`

	// 加密信息
	Encrypted     bool   `json:"encrypted" gorm:"default:false"`
	EncryptionKey string `json:"-" gorm:"column:encryption_key"` // 不在JSON中显示
}

// FileVersion 文件版本
type FileVersion struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FileID    string    `json:"file_id" gorm:"not null;index"`
	Version   int       `json:"version" gorm:"not null"`
	Path      string    `json:"path" gorm:"not null"`
	Size      int64     `json:"size" gorm:"not null"`
	Checksum  string    `json:"checksum" gorm:"not null"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

// ImageInfo 图片信息
type ImageInfo struct {
	ID         uint                   `json:"id" gorm:"primaryKey"`
	FileID     string                 `json:"file_id" gorm:"not null;uniqueIndex"`
	Width      int                    `json:"width"`
	Height     int                    `json:"height"`
	Format     string                 `json:"format"`
	ColorModel string                 `json:"color_model"`
	HasAlpha   bool                   `json:"has_alpha"`
	Thumbnails []Thumbnail            `json:"thumbnails" gorm:"foreignKey:ImageInfoID"`
	ExifData   map[string]interface{} `json:"exif_data" gorm:"serializer:json"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Thumbnail 缩略图
type Thumbnail struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ImageInfoID uint      `json:"image_info_id" gorm:"not null;index"`
	Size        string    `json:"size"` // small, medium, large, custom
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Path        string    `json:"path" gorm:"not null"`
	URL         string    `json:"url"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
}

// UploadToken 上传令牌
type UploadToken struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Token        string    `json:"token" gorm:"uniqueIndex;not null"`
	OwnerID      string    `json:"owner_id" gorm:"not null"`
	MaxSize      int64     `json:"max_size"`
	AllowedTypes []string  `json:"allowed_types" gorm:"serializer:json"`
	UsageCount   int       `json:"usage_count" gorm:"default:0"`
	MaxUsage     int       `json:"max_usage" gorm:"default:1"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
}

// AccessLog 访问日志
type AccessLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	FileID    string    `json:"file_id" gorm:"not null;index"`
	UserID    string    `json:"user_id" gorm:"index"`
	Action    string    `json:"action" gorm:"not null"` // upload, download, view, delete, share
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
	Success   bool      `json:"success" gorm:"default:true"`
	ErrorMsg  string    `json:"error_msg"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	Message string  `json:"message"`
	Files   []*File `json:"files"`
	Count   int     `json:"count"`
}

// ListFilesResponse 文件列表响应
type ListFilesResponse struct {
	Files      []*File `json:"files"`
	TotalCount int64   `json:"total_count"`
	Page       int     `json:"page"`
	PerPage    int     `json:"per_page"`
	TotalPages int     `json:"total_pages"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	Token        string   `json:"token"`
	ExpiresIn    int      `json:"expires_in"`
	MaxSize      int64    `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
	MaxUsage     int      `json:"max_usage"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// StatsResponse 统计响应
type StatsResponse struct {
	TotalFiles    int64                  `json:"total_files"`
	TotalSize     int64                  `json:"total_size"`
	FilesByType   map[string]int64       `json:"files_by_type"`
	FilesByOwner  map[string]int64       `json:"files_by_owner"`
	RecentUploads int64                  `json:"recent_uploads"`
	StorageUsage  map[string]interface{} `json:"storage_usage"`
}

// IsImage 检查是否为图片文件
func (f *File) IsImage() bool {
	return f.MimeType != "" && (f.MimeType[:6] == "image/")
}

// GetURL 获取文件访问URL
func (f *File) GetURL(baseURL string) string {
	if f.URL != "" {
		return f.URL
	}
	return baseURL + "/api/v1/files/" + f.ID + "/download"
}

// CanAccess 检查用户是否可以访问文件
func (f *File) CanAccess(userID string) bool {
	if f.Visibility == "public" {
		return true
	}
	if f.OwnerID == userID {
		return true
	}
	// 可以扩展更复杂的权限逻辑
	return false
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

func (FileVersion) TableName() string {
	return "file_versions"
}

func (ImageInfo) TableName() string {
	return "image_info"
}

func (Thumbnail) TableName() string {
	return "thumbnails"
}

func (UploadToken) TableName() string {
	return "upload_tokens"
}

func (AccessLog) TableName() string {
	return "access_logs"
}
