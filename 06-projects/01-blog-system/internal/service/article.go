package service

import (
	"blog-system/internal/model"
	"blog-system/internal/repository"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ArticleService 文章服务接口
type ArticleService interface {
	Create(req *CreateArticleRequest, authorID uint) (*model.Article, error)
	GetByID(id uint) (*model.Article, error)
	GetBySlug(slug string) (*model.Article, error)
	Update(id uint, req *UpdateArticleRequest, userID uint) (*model.Article, error)
	Delete(id uint, userID uint) error
	List(params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error)
	GetPublished(params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error)
	GetByAuthor(authorID uint, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error)
	GetByCategory(categorySlug string, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error)
	GetByTag(tagSlug string, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error)
	Search(keyword string, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error)
	Publish(id uint, userID uint) error
	Unpublish(id uint, userID uint) error
	IncrementView(id uint) error
	Like(id uint, userID uint) error
	GetPopular(limit int) ([]model.Article, error)
	GetRecent(limit int) ([]model.Article, error)
	GetRelated(articleID uint, limit int) ([]model.Article, error)
}

// CreateArticleRequest 创建文章请求
type CreateArticleRequest struct {
	Title       string   `json:"title" binding:"required,safe_string,max=200"`
	Content     string   `json:"content" binding:"required,safe_html"`
	Summary     string   `json:"summary" binding:"omitempty,safe_string,max=500"`
	FeaturedImg string   `json:"featured_img" binding:"omitempty,url"`
	CategoryID  uint     `json:"category_id" binding:"omitempty,min=1"`
	Tags        []string `json:"tags" binding:"omitempty,dive,safe_string,max=50"`
	Status      string   `json:"status" binding:"omitempty,oneof=draft published"` // draft, published
}

// UpdateArticleRequest 更新文章请求
type UpdateArticleRequest struct {
	Title       string   `json:"title" binding:"max=200"`
	Content     string   `json:"content"`
	Summary     string   `json:"summary"`
	FeaturedImg string   `json:"featured_img"`
	CategoryID  uint     `json:"category_id"`
	Tags        []string `json:"tags"`
	Status      string   `json:"status"`
}

type articleService struct {
	articleRepo  repository.ArticleRepository
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
}

// NewArticleService 创建文章服务
func NewArticleService(
	articleRepo repository.ArticleRepository,
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
) ArticleService {
	return &articleService{
		articleRepo:  articleRepo,
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
}

func (s *articleService) Create(req *CreateArticleRequest, authorID uint) (*model.Article, error) {
	// 验证作者权限
	author, err := s.userRepo.GetByID(authorID)
	if err != nil {
		return nil, errors.New("作者不存在")
	}

	if !author.CanEdit() {
		return nil, errors.New("没有创建文章的权限")
	}

	// 生成URL友好的slug
	slug := s.generateSlug(req.Title)

	// 检查slug是否已存在
	if _, err := s.articleRepo.GetBySlug(slug); err == nil {
		// 如果存在，添加时间戳
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	// 生成摘要（如果未提供）
	summary := req.Summary
	if summary == "" {
		summary = s.generateSummary(req.Content, 200)
	}

	// 确定发布时间
	var publishedAt *time.Time
	status := model.ArticleStatus(req.Status)
	if status == "" {
		status = model.StatusDraft
	}

	if status == model.StatusPublished {
		now := time.Now()
		publishedAt = &now
	}

	// 创建文章
	article := &model.Article{
		Title:       req.Title,
		Slug:        slug,
		Content:     req.Content,
		Summary:     summary,
		FeaturedImg: req.FeaturedImg,
		AuthorID:    authorID,
		CategoryID:  req.CategoryID,
		Status:      status,
		PublishedAt: publishedAt,
	}

	// 处理标签
	if len(req.Tags) > 0 {
		tags, err := s.tagRepo.GetOrCreateByNames(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("处理标签失败: %w", err)
		}
		article.Tags = tags
	}

	if err := s.articleRepo.Create(article); err != nil {
		return nil, fmt.Errorf("创建文章失败: %w", err)
	}

	// 重新获取完整信息
	return s.articleRepo.GetByID(article.ID)
}

func (s *articleService) GetByID(id uint) (*model.Article, error) {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}
	return article, nil
}

func (s *articleService) GetBySlug(slug string) (*model.Article, error) {
	article, err := s.articleRepo.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文章不存在")
		}
		return nil, err
	}
	return article, nil
}

func (s *articleService) Update(id uint, req *UpdateArticleRequest, userID uint) (*model.Article, error) {
	// 获取文章
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("文章不存在")
	}

	// 验证权限
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	if article.AuthorID != userID && !user.IsAdmin() {
		return nil, errors.New("没有编辑权限")
	}

	// 更新字段
	if req.Title != "" && req.Title != article.Title {
		article.Title = req.Title
		article.Slug = s.generateSlug(req.Title)
	}

	if req.Content != "" {
		article.Content = req.Content
		// 重新生成摘要
		if req.Summary == "" {
			article.Summary = s.generateSummary(req.Content, 200)
		}
	}

	if req.Summary != "" {
		article.Summary = req.Summary
	}

	if req.FeaturedImg != "" {
		article.FeaturedImg = req.FeaturedImg
	}

	if req.CategoryID > 0 {
		article.CategoryID = req.CategoryID
	}

	// 处理状态变更
	if req.Status != "" {
		newStatus := model.ArticleStatus(req.Status)
		if newStatus != article.Status {
			article.Status = newStatus
			if newStatus == model.StatusPublished && article.PublishedAt == nil {
				now := time.Now()
				article.PublishedAt = &now
			}
		}
	}

	// 处理标签
	if len(req.Tags) > 0 {
		tags, err := s.tagRepo.GetOrCreateByNames(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("处理标签失败: %w", err)
		}
		article.Tags = tags
	}

	if err := s.articleRepo.Update(article); err != nil {
		return nil, fmt.Errorf("更新文章失败: %w", err)
	}

	return s.articleRepo.GetByID(id)
}

func (s *articleService) Delete(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if article.AuthorID != userID && !user.IsAdmin() {
		return errors.New("没有删除权限")
	}

	return s.articleRepo.Delete(id)
}

func (s *articleService) List(params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error) {
	return s.articleRepo.List(params)
}

func (s *articleService) GetPublished(params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error) {
	return s.articleRepo.GetPublished(params)
}

func (s *articleService) GetByAuthor(authorID uint, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error) {
	return s.articleRepo.GetByAuthor(authorID, params)
}

func (s *articleService) GetByCategory(categorySlug string, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error) {
	category, err := s.categoryRepo.GetBySlug(categorySlug)
	if err != nil {
		return nil, nil, errors.New("分类不存在")
	}
	return s.articleRepo.GetByCategory(category.ID, params)
}

func (s *articleService) GetByTag(tagSlug string, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error) {
	tag, err := s.tagRepo.GetBySlug(tagSlug)
	if err != nil {
		return nil, nil, errors.New("标签不存在")
	}
	return s.articleRepo.GetByTag(tag.ID, params)
}

func (s *articleService) Search(keyword string, params repository.PaginationParams) ([]model.Article, *repository.PaginationResult, error) {
	if strings.TrimSpace(keyword) == "" {
		return s.GetPublished(params)
	}
	return s.articleRepo.SearchByKeyword(keyword, params)
}

func (s *articleService) Publish(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if article.AuthorID != userID && !user.IsAdmin() {
		return errors.New("没有发布权限")
	}

	article.Status = model.StatusPublished
	if article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
	}

	return s.articleRepo.Update(article)
}

func (s *articleService) Unpublish(id uint, userID uint) error {
	article, err := s.articleRepo.GetByID(id)
	if err != nil {
		return errors.New("文章不存在")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if article.AuthorID != userID && !user.IsAdmin() {
		return errors.New("没有权限")
	}

	article.Status = model.StatusDraft
	return s.articleRepo.Update(article)
}

func (s *articleService) IncrementView(id uint) error {
	return s.articleRepo.IncrementViewCount(id)
}

func (s *articleService) Like(id uint, userID uint) error {
	// 点赞功能：当前为简化实现，直接增加计数
	// 生产环境应添加点赞记录表防止重复点赞
	_ = userID // 预留参数，用于后续实现点赞记录
	return s.articleRepo.IncrementLikeCount(id)
}

func (s *articleService) GetPopular(limit int) ([]model.Article, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.articleRepo.GetPopular(limit)
}

func (s *articleService) GetRecent(limit int) ([]model.Article, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.articleRepo.GetRecent(limit)
}

func (s *articleService) GetRelated(articleID uint, limit int) ([]model.Article, error) {
	// 获取当前文章
	article, err := s.articleRepo.GetByID(articleID)
	if err != nil {
		return nil, err
	}

	// 简单实现：获取同分类的其他文章
	// 实际项目中可以基于标签相似度、内容相似度等算法
	params := repository.PaginationParams{Page: 1, PageSize: limit + 1}
	related, _, err := s.articleRepo.GetByCategory(article.CategoryID, params)
	if err != nil {
		return nil, err
	}

	// 排除当前文章
	var result []model.Article
	for _, a := range related {
		if a.ID != articleID {
			result = append(result, a)
		}
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// generateSlug 生成URL友好的slug
func (s *articleService) generateSlug(title string) string {
	// 转换为小写
	slug := strings.ToLower(title)

	// 简单的中文字符到拼音的映射（基本实现）
	// 在生产环境中，建议使用专业的拼音转换库
	chineseMap := map[string]string{
		"测试": "ceshi",
		"文章": "wenzhang",
		"博客": "blog",
		"系统": "xitong",
		"用户": "yonghu",
		"管理": "guanli",
		"分类": "fenlei",
		"标签": "biaoqian",
		"评论": "pinglun",
		"内容": "neirong",
		"发布": "fabu",
		"编辑": "bianji",
	}

	// 替换常见中文词汇
	for chinese, pinyin := range chineseMap {
		slug = strings.ReplaceAll(slug, chinese, pinyin)
	}

	// 去除剩余的非ASCII字符，但保留已转换的拼音
	reg := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = reg.ReplaceAllString(slug, "")

	// 如果slug为空（全是未映射的中文），使用默认值
	if strings.TrimSpace(slug) == "" {
		slug = fmt.Sprintf("article-%d", time.Now().Unix())
	}

	// 将空格替换为连字符
	reg = regexp.MustCompile(`\s+`)
	slug = reg.ReplaceAllString(slug, "-")

	// 去除首尾连字符
	slug = strings.Trim(slug, "-")

	// 限制长度
	if len(slug) > 100 {
		slug = slug[:100]
	}

	return slug
}

// generateSummary 生成文章摘要
func (s *articleService) generateSummary(content string, maxLength int) string {
	// 去除HTML标签（简单实现）
	reg := regexp.MustCompile(`<[^>]*>`)
	text := reg.ReplaceAllString(content, "")

	// 去除多余空白
	reg = regexp.MustCompile(`\s+`)
	text = reg.ReplaceAllString(strings.TrimSpace(text), " ")

	// 截取指定长度
	if len(text) <= maxLength {
		return text
	}

	// 尝试在句号处截断
	if idx := strings.LastIndex(text[:maxLength], "。"); idx > 0 && idx > maxLength/2 {
		return text[:idx+3] // 包含句号
	}

	// 否则在空格处截断
	if idx := strings.LastIndex(text[:maxLength], " "); idx > 0 && idx > maxLength/2 {
		return text[:idx] + "..."
	}

	// 直接截断
	return text[:maxLength] + "..."
}
