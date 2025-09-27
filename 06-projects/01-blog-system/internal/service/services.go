package service

import (
	"blog-system/internal/config"
	"blog-system/internal/model"
	"blog-system/internal/repository"
	"errors"
)

// Services 服务集合
type Services struct {
	Auth    AuthService
	Article ArticleService
	User    UserService
	Comment CommentService
}

// UserService 用户服务接口（简化）
type UserService interface {
	GetByID(id uint) (*model.User, error)
	Update(user *model.User) error
	List(params repository.PaginationParams) ([]model.User, *repository.PaginationResult, error)
	Search(keyword string, params repository.PaginationParams) ([]model.User, *repository.PaginationResult, error)
}

// CommentService 评论服务接口（简化）
type CommentService interface {
	Create(req *CreateCommentRequest, userID uint) (*model.Comment, error)
	GetByArticle(articleID uint, params repository.PaginationParams) ([]model.Comment, *repository.PaginationResult, error)
	Approve(id uint, userID uint) error
	Reject(id uint, userID uint) error
	Delete(id uint, userID uint) error
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content   string `json:"content" binding:"required"`
	ArticleID uint   `json:"article_id" binding:"required"`
	ParentID  *uint  `json:"parent_id"`
}

// userService 用户服务实现（简化）
type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) Update(user *model.User) error {
	return s.userRepo.Update(user)
}

func (s *userService) List(params repository.PaginationParams) ([]model.User, *repository.PaginationResult, error) {
	return s.userRepo.List(params)
}

func (s *userService) Search(keyword string, params repository.PaginationParams) ([]model.User, *repository.PaginationResult, error) {
	return s.userRepo.SearchByKeyword(keyword, params)
}

// commentService 评论服务实现（简化）
type commentService struct {
	commentRepo repository.CommentRepository
	articleRepo repository.ArticleRepository
	userRepo    repository.UserRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	articleRepo repository.ArticleRepository,
	userRepo repository.UserRepository,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		articleRepo: articleRepo,
		userRepo:    userRepo,
	}
}

func (s *commentService) Create(req *CreateCommentRequest, userID uint) (*model.Comment, error) {
	// 验证文章存在
	article, err := s.articleRepo.GetByID(req.ArticleID)
	if err != nil {
		return nil, errors.New("文章不存在")
	}

	if !article.CanBeCommented() {
		return nil, errors.New("文章不允许评论")
	}

	// 创建评论
	comment := &model.Comment{
		Content:   req.Content,
		ArticleID: req.ArticleID,
		UserID:    userID,
		ParentID:  req.ParentID,
		Status:    model.CommentApproved, // 简化实现，直接批准
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	return s.commentRepo.GetByID(comment.ID)
}

func (s *commentService) GetByArticle(articleID uint, params repository.PaginationParams) ([]model.Comment, *repository.PaginationResult, error) {
	return s.commentRepo.GetByArticle(articleID, params)
}

func (s *commentService) Approve(id uint, userID uint) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !user.IsAdmin() {
		return errors.New("没有权限")
	}

	return s.commentRepo.ApproveComment(id)
}

func (s *commentService) Reject(id uint, userID uint) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !user.IsAdmin() {
		return errors.New("没有权限")
	}

	return s.commentRepo.RejectComment(id)
}

func (s *commentService) Delete(id uint, userID uint) error {
	comment, err := s.commentRepo.GetByID(id)
	if err != nil {
		return errors.New("评论不存在")
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if comment.UserID != userID && !user.IsAdmin() {
		return errors.New("没有删除权限")
	}

	return s.commentRepo.Delete(id)
}

// NewServices 创建服务集合
func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth: NewAuthService(repos.User, cfg),
		Article: NewArticleService(
			repos.Article,
			repos.User,
			repos.Category,
			repos.Tag,
		),
		User:    NewUserService(repos.User),
		Comment: NewCommentService(repos.Comment, repos.Article, repos.User),
	}
}
