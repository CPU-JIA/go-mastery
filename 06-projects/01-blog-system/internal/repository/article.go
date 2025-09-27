package repository

import (
	"blog-system/internal/model"
	"fmt"

	"gorm.io/gorm"
)

type articleRepository struct {
	db *gorm.DB
}

// NewArticleRepository 创建文章仓储
func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(article *model.Article) error {
	return r.db.Create(article).Error
}

func (r *articleRepository) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("Comments", "status = ?", model.CommentApproved).
		Preload("Comments.User").
		First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) GetBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("Comments", "status = ? AND parent_id IS NULL", model.CommentApproved).
		Preload("Comments.User").
		Preload("Comments.Children", "status = ?", model.CommentApproved).
		Preload("Comments.Children.User").
		Where("slug = ?", slug).
		First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) Update(article *model.Article) error {
	return r.db.Save(article).Error
}

func (r *articleRepository) Delete(id uint) error {
	// 软删除，GORM会自动处理
	return r.db.Delete(&model.Article{}, id).Error
}

func (r *articleRepository) List(params PaginationParams) ([]model.Article, *PaginationResult, error) {
	var articles []model.Article
	var total int64

	// 计算总数
	if err := r.db.Model(&model.Article{}).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询文章列表
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&articles).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return articles, result, nil
}

func (r *articleRepository) GetPublished(params PaginationParams) ([]model.Article, *PaginationResult, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).
		Where("status = ? AND published_at IS NOT NULL", model.StatusPublished)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询文章列表
	err := query.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Order("published_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&articles).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return articles, result, nil
}

func (r *articleRepository) GetByAuthor(authorID uint, params PaginationParams) ([]model.Article, *PaginationResult, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).Where("author_id = ?", authorID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询文章列表
	err := query.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&articles).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return articles, result, nil
}

func (r *articleRepository) GetByCategory(categoryID uint, params PaginationParams) ([]model.Article, *PaginationResult, error) {
	var articles []model.Article
	var total int64

	query := r.db.Model(&model.Article{}).
		Where("category_id = ? AND status = ? AND published_at IS NOT NULL", categoryID, model.StatusPublished)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询文章列表
	err := query.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Order("published_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&articles).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return articles, result, nil
}

func (r *articleRepository) GetByTag(tagID uint, params PaginationParams) ([]model.Article, *PaginationResult, error) {
	var articles []model.Article
	var total int64

	// 使用子查询来处理多对多关系
	subQuery := r.db.Table("article_tags").
		Select("article_id").
		Where("tag_id = ?", tagID)

	query := r.db.Model(&model.Article{}).
		Where("id IN (?) AND status = ? AND published_at IS NOT NULL", subQuery, model.StatusPublished)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询文章列表
	err := query.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Order("published_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&articles).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return articles, result, nil
}

func (r *articleRepository) SearchByKeyword(keyword string, params PaginationParams) ([]model.Article, *PaginationResult, error) {
	var articles []model.Article
	var total int64

	searchTerm := fmt.Sprintf("%%%s%%", keyword)
	query := r.db.Model(&model.Article{}).
		Where("(title LIKE ? OR content LIKE ? OR summary LIKE ?) AND status = ? AND published_at IS NOT NULL",
			searchTerm, searchTerm, searchTerm, model.StatusPublished)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询文章列表，按相关性排序（简单的标题匹配优先）
	orderClause := fmt.Sprintf("CASE WHEN title LIKE '%s' THEN 1 ELSE 2 END, published_at DESC", searchTerm)
	err := query.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Order(orderClause).
		Limit(params.PageSize).
		Offset(offset).
		Find(&articles).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return articles, result, nil
}

func (r *articleRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *articleRepository) IncrementLikeCount(id uint) error {
	return r.db.Model(&model.Article{}).
		Where("id = ?", id).
		Update("like_count", gorm.Expr("like_count + 1")).Error
}

func (r *articleRepository) GetPopular(limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Where("status = ? AND published_at IS NOT NULL", model.StatusPublished).
		Order("view_count DESC, like_count DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

func (r *articleRepository) GetRecent(limit int) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Select("id, title, slug, summary, featured_img, author_id, category_id, status, view_count, like_count, created_at, updated_at, published_at").
		Where("status = ? AND published_at IS NOT NULL", model.StatusPublished).
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}
