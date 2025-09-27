package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	gorm.Model
	Username  string     `gorm:"uniqueIndex;size:50;not null" json:"username" binding:"required,min=3,max=50"`
	Email     string     `gorm:"uniqueIndex;size:100;not null" json:"email" binding:"required,email"`
	Password  string     `gorm:"size:255;not null" json:"-"` // 从JSON响应中排除密码
	FullName  string     `gorm:"size:100" json:"full_name"`
	Avatar    string     `gorm:"size:500" json:"avatar"`
	Bio       string     `gorm:"type:text" json:"bio"`
	Role      UserRole   `gorm:"type:varchar(20);default:'reader'" json:"role"`
	Status    UserStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	LastLogin *time.Time `json:"last_login"`

	// 关联关系
	Articles []Article `gorm:"foreignKey:AuthorID" json:"articles,omitempty"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"comments,omitempty"`
	Tags     []Tag     `gorm:"many2many:user_tags;" json:"tags,omitempty"`
}

// UserRole 用户角色枚举
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleAuthor UserRole = "author"
	RoleReader UserRole = "reader"
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusBanned   UserStatus = "banned"
)

// Article 文章模型
type Article struct {
	gorm.Model
	Title       string        `gorm:"size:200;not null;index" json:"title" binding:"required,max=200"`
	Slug        string        `gorm:"uniqueIndex;size:250;not null" json:"slug"`
	Content     string        `gorm:"type:longtext;not null" json:"content" binding:"required"`
	Summary     string        `gorm:"type:text" json:"summary"`
	FeaturedImg string        `gorm:"size:500" json:"featured_img"`
	AuthorID    uint          `gorm:"not null;index" json:"author_id"`
	CategoryID  uint          `gorm:"index" json:"category_id"`
	Status      ArticleStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	ViewCount   int           `gorm:"default:0" json:"view_count"`
	LikeCount   int           `gorm:"default:0" json:"like_count"`
	PublishedAt *time.Time    `gorm:"index" json:"published_at"`

	// 关联关系
	Author   User      `gorm:"foreignKey:AuthorID" json:"author"`
	Category Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Comments []Comment `gorm:"foreignKey:ArticleID" json:"comments,omitempty"`
	Tags     []Tag     `gorm:"many2many:article_tags;" json:"tags,omitempty"`
}

// ArticleStatus 文章状态枚举
type ArticleStatus string

const (
	StatusDraft     ArticleStatus = "draft"
	StatusPublished ArticleStatus = "published"
	StatusArchived  ArticleStatus = "archived"
)

// Category 分类模型
type Category struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;size:50;not null" json:"name" binding:"required,max=50"`
	Slug        string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	Color       string `gorm:"size:7" json:"color"` // 十六进制颜色值
	ParentID    *uint  `gorm:"index" json:"parent_id"`

	// 自关联：父子分类
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Articles []Article  `gorm:"foreignKey:CategoryID" json:"articles,omitempty"`
}

// Tag 标签模型
type Tag struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;size:50;not null" json:"name" binding:"required,max=50"`
	Slug  string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Color string `gorm:"size:7" json:"color"` // 十六进制颜色值

	// 多对多关联
	Articles []Article `gorm:"many2many:article_tags;" json:"articles,omitempty"`
	Users    []User    `gorm:"many2many:user_tags;" json:"users,omitempty"`
}

// Comment 评论模型
type Comment struct {
	gorm.Model
	Content   string        `gorm:"type:text;not null" json:"content" binding:"required"`
	ArticleID uint          `gorm:"not null;index" json:"article_id" binding:"required"`
	UserID    uint          `gorm:"not null;index" json:"user_id"`
	ParentID  *uint         `gorm:"index" json:"parent_id"` // 回复功能
	Status    CommentStatus `gorm:"type:varchar(20);default:'approved';index" json:"status"`
	LikeCount int           `gorm:"default:0" json:"like_count"`
	IPAddress string        `gorm:"size:45" json:"-"` // 不在API中显示IP
	UserAgent string        `gorm:"size:500" json:"-"`

	// 关联关系
	Article  Article   `gorm:"foreignKey:ArticleID" json:"article"`
	User     User      `gorm:"foreignKey:UserID" json:"user"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Comment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// CommentStatus 评论状态枚举
type CommentStatus string

const (
	CommentPending  CommentStatus = "pending"
	CommentApproved CommentStatus = "approved"
	CommentRejected CommentStatus = "rejected"
	CommentSpam     CommentStatus = "spam"
)

// Setting 系统设置模型
type Setting struct {
	gorm.Model
	Key         string `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Description string `gorm:"type:text" json:"description"`
	Category    string `gorm:"size:50;index" json:"category"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"`
}

// Media 媒体文件模型
type Media struct {
	gorm.Model
	FileName     string `gorm:"size:255;not null" json:"file_name"`
	OriginalName string `gorm:"size:255;not null" json:"original_name"`
	FilePath     string `gorm:"size:500;not null" json:"file_path"`
	FileSize     int64  `gorm:"not null" json:"file_size"`
	MimeType     string `gorm:"size:100;not null" json:"mime_type"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	UserID       uint   `gorm:"not null;index" json:"user_id"`

	// 关联关系
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// TableName 定义表名（可选，GORM会自动复数化）
func (User) TableName() string     { return "users" }
func (Article) TableName() string  { return "articles" }
func (Category) TableName() string { return "categories" }
func (Tag) TableName() string      { return "tags" }
func (Comment) TableName() string  { return "comments" }
func (Setting) TableName() string  { return "settings" }
func (Media) TableName() string    { return "media" }

// BeforeCreate 创建前的钩子函数示例
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 可以在这里添加创建前的逻辑
	// 例如：生成用户名，设置默认头像等
	if u.Role == "" {
		u.Role = RoleReader
	}
	if u.Status == "" {
		u.Status = StatusActive
	}
	return nil
}

func (a *Article) BeforeCreate(tx *gorm.DB) error {
	if a.Status == "" {
		a.Status = StatusDraft
	}
	return nil
}

// IsAdmin 检查用户是否为管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanEdit 检查用户是否能编辑
func (u *User) CanEdit() bool {
	return u.Role == RoleAdmin || u.Role == RoleAuthor
}

// IsPublished 检查文章是否已发布
func (a *Article) IsPublished() bool {
	return a.Status == StatusPublished && a.PublishedAt != nil
}

// CanBeCommented 检查文章是否可以被评论
func (a *Article) CanBeCommented() bool {
	return a.IsPublished()
}

// IsApproved 检查评论是否已批准
func (c *Comment) IsApproved() bool {
	return c.Status == CommentApproved
}
