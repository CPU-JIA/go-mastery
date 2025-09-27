package handler

import (
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler 处理器集合
type Handler struct {
	services *service.Services
}

// NewHandler 创建处理器
func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

// AuthHandler 认证相关处理器
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		var status int
		switch err.Error() {
		case "username already exists", "email already exists":
			status = http.StatusConflict
		default:
			status = http.StatusInternalServerError
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
			"code":  "REGISTRATION_FAILED",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	authResp, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"code":  "LOGIN_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    authResp,
	})
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"code":  "INVALID_REQUEST",
		})
		return
	}

	authResp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
			"code":  "REFRESH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"data":    authResp,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
			"code":  "AUTH_REQUIRED",
		})
		return
	}

	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	if err := h.authService.ChangePassword(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "PASSWORD_CHANGE_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// ProductHandler 商品相关处理器
type ProductHandler struct {
	productService service.ProductService
}

// NewProductHandler 创建商品处理器
func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// GetProducts 获取商品列表
func (h *ProductHandler) GetProducts(c *gin.Context) {
	var req service.ProductListRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid query parameters",
			"code":    "INVALID_QUERY",
			"details": err.Error(),
		})
		return
	}

	// 设置默认分页参数
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	resp, err := h.productService.List(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "PRODUCT_LIST_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products retrieved successfully",
		"data":    resp,
	})
}

// GetProduct 获取单个商品
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
			"code":  "INVALID_ID",
		})
		return
	}

	product, err := h.productService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
			"code":  "PRODUCT_NOT_FOUND",
		})
		return
	}

	// 增加浏览次数
	go h.productService.IncrementViewCount(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"message": "Product retrieved successfully",
		"data":    product,
	})
}

// GetProductBySlug 通过slug获取商品
func (h *ProductHandler) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")

	product, err := h.productService.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Product not found",
			"code":  "PRODUCT_NOT_FOUND",
		})
		return
	}

	// 增加浏览次数
	go h.productService.IncrementViewCount(product.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Product retrieved successfully",
		"data":    product,
	})
}

// SearchProducts 搜索商品
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
			"code":  "QUERY_REQUIRED",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resp, err := h.productService.Search(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "SEARCH_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Search completed successfully",
		"data":    resp,
	})
}

// GetFeaturedProducts 获取推荐商品
func (h *ProductHandler) GetFeaturedProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := h.productService.GetFeatured(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "FEATURED_PRODUCTS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Featured products retrieved successfully",
		"data":    products,
	})
}

// GetRelatedProducts 获取相关商品
func (h *ProductHandler) GetRelatedProducts(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid product ID",
			"code":  "INVALID_ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))

	products, err := h.productService.GetRelated(uint(id), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "RELATED_PRODUCTS_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Related products retrieved successfully",
		"data":    products,
	})
}

// 工具函数：构建分页信息
func buildPagination(total int64, page, limit int) map[string]interface{} {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return map[string]interface{}{
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"has_prev":    page > 1,
		"has_next":    page < totalPages,
	}
}
