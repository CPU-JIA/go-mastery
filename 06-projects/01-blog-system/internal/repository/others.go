package repository

import (
	"blog-system/internal/model"

	"gorm.io/gorm"
)

// ============ Category Repository ============
type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *model.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) GetByID(id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.Preload("Parent").Preload("Children").First(&category, id).Error
	return &category, err
}

func (r *categoryRepository) GetBySlug(slug string) (*model.Category, error) {
	var category model.Category
	err := r.db.Preload("Parent").Preload("Children").Where("slug = ?", slug).First(&category).Error
	return &category, err
}

func (r *categoryRepository) Update(category *model.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.Category{}, id).Error
}

func (r *categoryRepository) List() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Preload("Parent").Preload("Children").Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetWithArticleCount() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Preload("Parent").
		Select("categories.*, COUNT(articles.id) as article_count").
		Joins("LEFT JOIN articles ON categories.id = articles.category_id AND articles.status = ? AND articles.deleted_at IS NULL", model.StatusPublished).
		Group("categories.id").
		Order("categories.name ASC").
		Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetParentCategories() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("parent_id IS NULL").Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) GetChildCategories(parentID uint) ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Where("parent_id = ?", parentID).Order("name ASC").Find(&categories).Error
	return categories, err
}

// ============ Tag Repository ============
type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepository) GetByID(id uint) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.First(&tag, id).Error
	return &tag, err
}

func (r *tagRepository) GetBySlug(slug string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("slug = ?", slug).First(&tag).Error
	return &tag, err
}

func (r *tagRepository) GetByName(name string) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("name = ?", name).First(&tag).Error
	return &tag, err
}

func (r *tagRepository) Update(tag *model.Tag) error {
	return r.db.Save(tag).Error
}

func (r *tagRepository) Delete(id uint) error {
	return r.db.Delete(&model.Tag{}, id).Error
}

func (r *tagRepository) List() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Order("name ASC").Find(&tags).Error
	return tags, err
}

func (r *tagRepository) GetPopularTags(limit int) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Select("tags.*, COUNT(article_tags.article_id) as usage_count").
		Joins("LEFT JOIN article_tags ON tags.id = article_tags.tag_id").
		Joins("LEFT JOIN articles ON article_tags.article_id = articles.id AND articles.status = ? AND articles.deleted_at IS NULL", model.StatusPublished).
		Group("tags.id").
		Order("usage_count DESC, tags.name ASC").
		Limit(limit).
		Find(&tags).Error
	return tags, err
}

func (r *tagRepository) GetTagsWithArticleCount() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Select("tags.*, COUNT(article_tags.article_id) as article_count").
		Joins("LEFT JOIN article_tags ON tags.id = article_tags.tag_id").
		Joins("LEFT JOIN articles ON article_tags.article_id = articles.id AND articles.status = ? AND articles.deleted_at IS NULL", model.StatusPublished).
		Group("tags.id").
		Order("article_count DESC, tags.name ASC").
		Find(&tags).Error
	return tags, err
}

func (r *tagRepository) GetOrCreateByNames(names []string) ([]model.Tag, error) {
	var tags []model.Tag

	for _, name := range names {
		var tag model.Tag
		err := r.db.Where("name = ?", name).First(&tag).Error
		if err == gorm.ErrRecordNotFound {
			// 创建新标签
			tag = model.Tag{
				Name:  name,
				Slug:  name,      // 简化处理，实际应该生成URL友好的slug
				Color: "#6c757d", // 默认颜色
			}
			if err := r.db.Create(&tag).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// ============ Comment Repository ============
type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *commentRepository) GetByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Preload("User").Preload("Article").Preload("Parent").First(&comment, id).Error
	return &comment, err
}

func (r *commentRepository) Update(comment *model.Comment) error {
	return r.db.Save(comment).Error
}

func (r *commentRepository) Delete(id uint) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

func (r *commentRepository) GetByArticle(articleID uint, params PaginationParams) ([]model.Comment, *PaginationResult, error) {
	var comments []model.Comment
	var total int64

	query := r.db.Model(&model.Comment{}).
		Where("article_id = ? AND status = ? AND parent_id IS NULL", articleID, model.CommentApproved)

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.Preload("User").
		Preload("Children", "status = ?", model.CommentApproved).
		Preload("Children.User").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&comments).Error

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

	return comments, result, err
}

func (r *commentRepository) GetByUser(userID uint, params PaginationParams) ([]model.Comment, *PaginationResult, error) {
	var comments []model.Comment
	var total int64

	query := r.db.Model(&model.Comment{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.Preload("User").Preload("Article").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&comments).Error

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

	return comments, result, err
}

func (r *commentRepository) GetPending(params PaginationParams) ([]model.Comment, *PaginationResult, error) {
	var comments []model.Comment
	var total int64

	query := r.db.Model(&model.Comment{}).Where("status = ?", model.CommentPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.Preload("User").Preload("Article").
		Order("created_at ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&comments).Error

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

	return comments, result, err
}

func (r *commentRepository) ApproveComment(id uint) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).Update("status", model.CommentApproved).Error
}

func (r *commentRepository) RejectComment(id uint) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", id).Update("status", model.CommentRejected).Error
}

func (r *commentRepository) GetReplies(parentID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Preload("User").
		Where("parent_id = ? AND status = ?", parentID, model.CommentApproved).
		Order("created_at ASC").
		Find(&comments).Error
	return comments, err
}

func (r *commentRepository) GetCommentTree(articleID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Preload("User").
		Preload("Children", "status = ?", model.CommentApproved).
		Preload("Children.User").
		Where("article_id = ? AND status = ? AND parent_id IS NULL", articleID, model.CommentApproved).
		Order("created_at DESC").
		Find(&comments).Error
	return comments, err
}

// ============ Media Repository ============
type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(media *model.Media) error {
	return r.db.Create(media).Error
}

func (r *mediaRepository) GetByID(id uint) (*model.Media, error) {
	var media model.Media
	err := r.db.Preload("User").First(&media, id).Error
	return &media, err
}

func (r *mediaRepository) Update(media *model.Media) error {
	return r.db.Save(media).Error
}

func (r *mediaRepository) Delete(id uint) error {
	return r.db.Delete(&model.Media{}, id).Error
}

func (r *mediaRepository) List(params PaginationParams) ([]model.Media, *PaginationResult, error) {
	var media []model.Media
	var total int64

	if err := r.db.Model(&model.Media{}).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := r.db.Preload("User").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&media).Error

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

	return media, result, err
}

func (r *mediaRepository) GetByUser(userID uint, params PaginationParams) ([]model.Media, *PaginationResult, error) {
	var media []model.Media
	var total int64

	query := r.db.Model(&model.Media{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.Preload("User").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&media).Error

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

	return media, result, err
}

func (r *mediaRepository) GetByType(mimeType string, params PaginationParams) ([]model.Media, *PaginationResult, error) {
	var media []model.Media
	var total int64

	query := r.db.Model(&model.Media{}).Where("mime_type = ?", mimeType)

	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	offset := (params.Page - 1) * params.PageSize
	err := query.Preload("User").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&media).Error

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

	return media, result, err
}

// ============ Setting Repository ============
type settingRepository struct {
	db *gorm.DB
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return &settingRepository{db: db}
}

func (r *settingRepository) GetByKey(key string) (*model.Setting, error) {
	var setting model.Setting
	err := r.db.Where("key = ?", key).First(&setting).Error
	return &setting, err
}

func (r *settingRepository) GetByCategory(category string) ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Where("category = ?", category).Order("key ASC").Find(&settings).Error
	return settings, err
}

func (r *settingRepository) GetPublicSettings() ([]model.Setting, error) {
	var settings []model.Setting
	err := r.db.Where("is_public = ?", true).Order("category ASC, key ASC").Find(&settings).Error
	return settings, err
}

func (r *settingRepository) Update(setting *model.Setting) error {
	return r.db.Save(setting).Error
}

func (r *settingRepository) CreateOrUpdate(setting *model.Setting) error {
	var existing model.Setting
	err := r.db.Where("key = ?", setting.Key).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.Create(setting).Error
	} else if err != nil {
		return err
	}

	existing.Value = setting.Value
	existing.Description = setting.Description
	existing.Category = setting.Category
	existing.IsPublic = setting.IsPublic
	return r.db.Save(&existing).Error
}

func (r *settingRepository) GetAllAsMap() (map[string]string, error) {
	var settings []model.Setting
	err := r.db.Find(&settings).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}
	return result, nil
}
