package model

import (
	"time"

	"gorm.io/gorm"
)

// File 文件信息模型
type File struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UUID           string         `gorm:"uniqueIndex;not null" json:"uuid"`
	OriginalName   string         `gorm:"not null" json:"original_name"`
	StorageName    string         `gorm:"not null" json:"storage_name"`
	Size           int64          `gorm:"not null" json:"size"`
	MimeType       string         `gorm:"not null" json:"mime_type"`
	IsEncrypted    bool           `gorm:"default:false" json:"is_encrypted"`
	EncryptKey     string         `json:"-"` // 不在JSON中暴露
	ChecksumMD5    string         `json:"checksum_md5"`
	ChecksumSHA256 string         `json:"checksum_sha256"`
	UserID         uint           `gorm:"not null;index" json:"user_id"`
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status         FileStatus     `gorm:"default:active" json:"status"`
	UploadedAt     time.Time      `json:"uploaded_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联的文件元数据
	ImageInfo     *ImageInfo     `gorm:"foreignKey:FileID" json:"image_info,omitempty"`
	ThumbnailInfo *ThumbnailInfo `gorm:"foreignKey:FileID" json:"thumbnail_info,omitempty"`
	AccessLogs    []AccessLog    `gorm:"foreignKey:FileID" json:"-"`
}

// ImageInfo 图像文件信息
type ImageInfo struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	FileID    uint      `gorm:"uniqueIndex" json:"file_id"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Format    string    `json:"format"`
	ColorMode string    `json:"color_mode"`
	HasAlpha  bool      `json:"has_alpha"`
	DPI       int       `json:"dpi"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ThumbnailInfo 缩略图信息
type ThumbnailInfo struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FileID      uint      `gorm:"uniqueIndex" json:"file_id"`
	StorageName string    `gorm:"not null" json:"storage_name"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Size        int64     `json:"size"`
	Format      string    `json:"format"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AccessLog 文件访问日志
type AccessLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FileID     uint      `gorm:"index" json:"file_id"`
	File       File      `gorm:"foreignKey:FileID" json:"-"`
	UserID     uint      `gorm:"index" json:"user_id"`
	User       User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Action     string    `json:"action"` // upload, download, view, delete
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	AccessedAt time.Time `json:"accessed_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// FileStatus 文件状态枚举
type FileStatus string

const (
	FileStatusActive     FileStatus = "active"
	FileStatusProcessing FileStatus = "processing"
	FileStatusError      FileStatus = "error"
	FileStatusDeleted    FileStatus = "deleted"
	FileStatusArchived   FileStatus = "archived"
)

// UploadToken 上传令牌（用于大文件分片上传）
type UploadToken struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Token          string    `gorm:"uniqueIndex;not null" json:"token"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"-"`
	OriginalName   string    `json:"original_name"`
	TotalSize      int64     `json:"total_size"`
	ChunkSize      int       `json:"chunk_size"`
	TotalChunks    int       `json:"total_chunks"`
	UploadedChunks []int     `gorm:"serializer:json" json:"uploaded_chunks"`
	ExpiresAt      time.Time `json:"expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FileShare 文件分享链接
type FileShare struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	FileID        uint       `gorm:"index" json:"file_id"`
	File          File       `gorm:"foreignKey:FileID" json:"file,omitempty"`
	ShareToken    string     `gorm:"uniqueIndex;not null" json:"share_token"`
	Password      string     `json:"-"`                              // 加密存储
	MaxDownloads  int        `gorm:"default:0" json:"max_downloads"` // 0表示无限制
	DownloadCount int        `gorm:"default:0" json:"download_count"`
	ExpiresAt     *time.Time `json:"expires_at"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (File) TableName() string {
	return "files"
}

func (ImageInfo) TableName() string {
	return "file_image_info"
}

func (ThumbnailInfo) TableName() string {
	return "file_thumbnails"
}

func (AccessLog) TableName() string {
	return "file_access_logs"
}

func (UploadToken) TableName() string {
	return "upload_tokens"
}

func (FileShare) TableName() string {
	return "file_shares"
}
