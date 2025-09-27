package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UUID        string     `gorm:"uniqueIndex;not null" json:"uuid"`
	Username    string     `gorm:"uniqueIndex;not null" json:"username"`
	Email       string     `gorm:"uniqueIndex;not null" json:"email"`
	Password    string     `gorm:"not null" json:"-"` // 不在JSON中暴露
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Avatar      string     `json:"avatar"`
	Role        UserRole   `gorm:"default:customer" json:"role"`
	Status      UserStatus `gorm:"default:active" json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`

	// 存储配额
	StorageQuota int64 `gorm:"default:1073741824" json:"storage_quota"` // 默认1GB
	StorageUsed  int64 `gorm:"default:0" json:"storage_used"`
	MaxFileSize  int64 `gorm:"default:104857600" json:"max_file_size"` // 默认100MB

	// 权限设置
	CanUpload       bool `gorm:"default:true" json:"can_upload"`
	CanShare        bool `gorm:"default:true" json:"can_share"`
	CanCreateFolder bool `gorm:"default:true" json:"can_create_folder"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	Files        []File        `gorm:"foreignKey:UserID" json:"-"`
	AccessLogs   []AccessLog   `gorm:"foreignKey:UserID" json:"-"`
	UploadTokens []UploadToken `gorm:"foreignKey:UserID" json:"-"`
}

// UserRole 用户角色枚举
type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleModerator UserRole = "moderator"
	UserRoleCustomer  UserRole = "customer"
	UserRoleGuest     UserRole = "guest"
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// Folder 文件夹模型
type Folder struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UUID        string         `gorm:"uniqueIndex;not null" json:"uuid"`
	Name        string         `gorm:"not null" json:"name"`
	Path        string         `gorm:"not null" json:"path"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ParentID    *uint          `gorm:"index" json:"parent_id"`
	Parent      *Folder        `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []Folder       `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// FileFolder 文件和文件夹的关联表
type FileFolder struct {
	FileID    uint   `gorm:"primaryKey"`
	FolderID  uint   `gorm:"primaryKey"`
	File      File   `gorm:"foreignKey:FileID"`
	Folder    Folder `gorm:"foreignKey:FolderID"`
	CreatedAt time.Time
}

// UserSettings 用户设置
type UserSettings struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"uniqueIndex;not null" json:"user_id"`
	User     User   `gorm:"foreignKey:UserID" json:"-"`
	Language string `gorm:"default:en" json:"language"`
	Timezone string `gorm:"default:UTC" json:"timezone"`
	Theme    string `gorm:"default:light" json:"theme"`

	// 通知设置
	EmailNotifications  bool `gorm:"default:true" json:"email_notifications"`
	UploadNotifications bool `gorm:"default:true" json:"upload_notifications"`
	ShareNotifications  bool `gorm:"default:true" json:"share_notifications"`

	// 默认设置
	DefaultFolderID       *uint `json:"default_folder_id"`
	AutoGenerateThumbnail bool  `gorm:"default:true" json:"auto_generate_thumbnail"`
	DefaultShareExpiry    int   `gorm:"default:24" json:"default_share_expiry"` // 小时

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StorageStats 存储统计信息
type StorageStats struct {
	UserID              uint             `json:"user_id"`
	TotalFiles          int64            `json:"total_files"`
	TotalSize           int64            `json:"total_size"`
	UsedQuotaPercentage float64          `json:"used_quota_percentage"`
	FilesByType         map[string]int64 `json:"files_by_type"`
	FilesByMonth        map[string]int64 `json:"files_by_month"`
	TopFiles            []File           `json:"top_files"`
	RecentFiles         []File           `json:"recent_files"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Folder) TableName() string {
	return "folders"
}

func (FileFolder) TableName() string {
	return "file_folders"
}

func (UserSettings) TableName() string {
	return "user_settings"
}
