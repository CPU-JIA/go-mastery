package handler

import (
	"blog-system/internal/model"
	"blog-system/internal/repository"
	"blog-system/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Response 通用API响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Code       int                          `json:"code"`
	Message    string                       `json:"message"`
	Data       interface{}                  `json:"data,omitempty"`
	Pagination *repository.PaginationResult `json:"pagination,omitempty"`
}

// Handler HTTP处理器集合
type Handler struct {
	services *service.Services
}

// NewHandler 创建处理器
func NewHandler(services *service.Services) *Handler {
	return &Handler{services: services}
}

// ============ 认证相关处理器 ============

// Register 用户注册
func (h *Handler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	user, err := h.services.Auth.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    201,
		Message: "注册成功",
		Data:    user,
	})
}

// Login 用户登录
func (h *Handler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	authResp, err := h.services.Auth.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "登录成功",
		Data:    authResp,
	})
}

// RefreshToken 刷新令牌
func (h *Handler) RefreshToken(c *gin.Context) {
	type RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	authResp, err := h.services.Auth.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "刷新成功",
		Data:    authResp,
	})
}

// Me 获取当前用户信息
func (h *Handler) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "未认证用户",
		})
		return
	}

	user, err := h.services.User.GetByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    user,
	})
}

// ChangePassword 修改密码
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	err := h.services.Auth.ChangePassword(userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "密码修改成功",
	})
}

// ============ 文章相关处理器 ============

// CreateArticle 创建文章
func (h *Handler) CreateArticle(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req service.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	article, err := h.services.Article.Create(&req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    201,
		Message: "创建成功",
		Data:    article,
	})
}

// GetArticle 获取文章详情
func (h *Handler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的文章ID",
		})
		return
	}

	article, err := h.services.Article.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "文章不存在",
		})
		return
	}

	// 增加浏览量
	h.services.Article.IncrementView(uint(id))

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    article,
	})
}

// GetArticleBySlug 通过slug获取文章
func (h *Handler) GetArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.services.Article.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "文章不存在",
		})
		return
	}

	// 增加浏览量
	h.services.Article.IncrementView(article.ID)

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    article,
	})
}

// ListArticles 获取文章列表
func (h *Handler) ListArticles(c *gin.Context) {
	page, _ := c.Get("page")
	pageSize, _ := c.Get("page_size")

	params := repository.PaginationParams{
		Page:     page.(int),
		PageSize: pageSize.(int),
	}

	// 判断是否只获取已发布的文章
	onlyPublished := c.Query("published") != "false"

	var articles []model.Article
	var pagination *repository.PaginationResult
	var err error

	if onlyPublished {
		articles, pagination, err = h.services.Article.GetPublished(params)
	} else {
		articles, pagination, err = h.services.Article.List(params)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取文章列表失败",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Code:       200,
		Message:    "获取成功",
		Data:       articles,
		Pagination: pagination,
	})
}

// UpdateArticle 更新文章
func (h *Handler) UpdateArticle(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的文章ID",
		})
		return
	}

	var req service.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "参数错误",
			Data:    err.Error(),
		})
		return
	}

	article, err := h.services.Article.Update(uint(id), &req, userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "更新成功",
		Data:    article,
	})
}

// DeleteArticle 删除文章
func (h *Handler) DeleteArticle(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的文章ID",
		})
		return
	}

	err = h.services.Article.Delete(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "删除成功",
	})
}

// SearchArticles 搜索文章
func (h *Handler) SearchArticles(c *gin.Context) {
	page, _ := c.Get("page")
	pageSize, _ := c.Get("page_size")
	keyword := c.Query("q")

	params := repository.PaginationParams{
		Page:     page.(int),
		PageSize: pageSize.(int),
	}

	articles, pagination, err := h.services.Article.Search(keyword, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "搜索失败",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Code:       200,
		Message:    "搜索成功",
		Data:       articles,
		Pagination: pagination,
	})
}

// PublishArticle 发布文章
func (h *Handler) PublishArticle(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的文章ID",
		})
		return
	}

	err = h.services.Article.Publish(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "发布成功",
	})
}

// LikeArticle 点赞文章
func (h *Handler) LikeArticle(c *gin.Context) {
	userID, _ := c.Get("user_id")

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的文章ID",
		})
		return
	}

	err = h.services.Article.Like(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "点赞成功",
	})
}

// GetPopularArticles 获取热门文章
func (h *Handler) GetPopularArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	articles, err := h.services.Article.GetPopular(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取热门文章失败",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    articles,
	})
}

// GetRecentArticles 获取最新文章
func (h *Handler) GetRecentArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	articles, err := h.services.Article.GetRecent(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "获取最新文章失败",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "获取成功",
		Data:    articles,
	})
}
