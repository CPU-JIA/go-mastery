package repository

import (
	"blog-system/internal/model"

	"gorm.io/gorm"
)

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// PaginationResult 分页结果
type PaginationResult struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
	List(params PaginationParams) ([]model.User, *PaginationResult, error)
	SearchByKeyword(keyword string, params PaginationParams) ([]model.User, *PaginationResult, error)
	GetActiveUsers() ([]model.User, error)
	UpdateLastLogin(id uint) error
}

// ArticleRepository 文章仓储接口
type ArticleRepository interface {
	Create(article *model.Article) error
	GetByID(id uint) (*model.Article, error)
	GetBySlug(slug string) (*model.Article, error)
	Update(article *model.Article) error
	Delete(id uint) error
	List(params PaginationParams) ([]model.Article, *PaginationResult, error)
	GetPublished(params PaginationParams) ([]model.Article, *PaginationResult, error)
	GetByAuthor(authorID uint, params PaginationParams) ([]model.Article, *PaginationResult, error)
	GetByCategory(categoryID uint, params PaginationParams) ([]model.Article, *PaginationResult, error)
	GetByTag(tagID uint, params PaginationParams) ([]model.Article, *PaginationResult, error)
	SearchByKeyword(keyword string, params PaginationParams) ([]model.Article, *PaginationResult, error)
	IncrementViewCount(id uint) error
	IncrementLikeCount(id uint) error
	GetPopular(limit int) ([]model.Article, error)
	GetRecent(limit int) ([]model.Article, error)
}

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	Create(category *model.Category) error
	GetByID(id uint) (*model.Category, error)
	GetBySlug(slug string) (*model.Category, error)
	Update(category *model.Category) error
	Delete(id uint) error
	List() ([]model.Category, error)
	GetWithArticleCount() ([]model.Category, error)
	GetParentCategories() ([]model.Category, error)
	GetChildCategories(parentID uint) ([]model.Category, error)
}

// TagRepository 标签仓储接口
type TagRepository interface {
	Create(tag *model.Tag) error
	GetByID(id uint) (*model.Tag, error)
	GetBySlug(slug string) (*model.Tag, error)
	GetByName(name string) (*model.Tag, error)
	Update(tag *model.Tag) error
	Delete(id uint) error
	List() ([]model.Tag, error)
	GetPopularTags(limit int) ([]model.Tag, error)
	GetTagsWithArticleCount() ([]model.Tag, error)
	GetOrCreateByNames(names []string) ([]model.Tag, error)
}

// CommentRepository 评论仓储接口
type CommentRepository interface {
	Create(comment *model.Comment) error
	GetByID(id uint) (*model.Comment, error)
	Update(comment *model.Comment) error
	Delete(id uint) error
	GetByArticle(articleID uint, params PaginationParams) ([]model.Comment, *PaginationResult, error)
	GetByUser(userID uint, params PaginationParams) ([]model.Comment, *PaginationResult, error)
	GetPending(params PaginationParams) ([]model.Comment, *PaginationResult, error)
	ApproveComment(id uint) error
	RejectComment(id uint) error
	GetReplies(parentID uint) ([]model.Comment, error)
	GetCommentTree(articleID uint) ([]model.Comment, error)
}

// MediaRepository 媒体仓储接口
type MediaRepository interface {
	Create(media *model.Media) error
	GetByID(id uint) (*model.Media, error)
	Update(media *model.Media) error
	Delete(id uint) error
	List(params PaginationParams) ([]model.Media, *PaginationResult, error)
	GetByUser(userID uint, params PaginationParams) ([]model.Media, *PaginationResult, error)
	GetByType(mimeType string, params PaginationParams) ([]model.Media, *PaginationResult, error)
}

// SettingRepository 设置仓储接口
type SettingRepository interface {
	GetByKey(key string) (*model.Setting, error)
	GetByCategory(category string) ([]model.Setting, error)
	GetPublicSettings() ([]model.Setting, error)
	Update(setting *model.Setting) error
	CreateOrUpdate(setting *model.Setting) error
	GetAllAsMap() (map[string]string, error)
}

// Repositories 仓储集合
type Repositories struct {
	User     UserRepository
	Article  ArticleRepository
	Category CategoryRepository
	Tag      TagRepository
	Comment  CommentRepository
	Media    MediaRepository
	Setting  SettingRepository
	DB       *gorm.DB
}

// NewRepositories 创建仓储集合
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:     NewUserRepository(db),
		Article:  NewArticleRepository(db),
		Category: NewCategoryRepository(db),
		Tag:      NewTagRepository(db),
		Comment:  NewCommentRepository(db),
		Media:    NewMediaRepository(db),
		Setting:  NewSettingRepository(db),
		DB:       db,
	}
}

// WithTx 使用事务
func (r *Repositories) WithTx(tx *gorm.DB) *Repositories {
	return &Repositories{
		User:     NewUserRepository(tx),
		Article:  NewArticleRepository(tx),
		Category: NewCategoryRepository(tx),
		Tag:      NewTagRepository(tx),
		Comment:  NewCommentRepository(tx),
		Media:    NewMediaRepository(tx),
		Setting:  NewSettingRepository(tx),
		DB:       tx,
	}
}
