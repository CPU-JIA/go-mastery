package service

import (
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/shopspring/decimal"
)

// productService 商品服务实现
type productService struct {
	productRepo repository.ProductRepository
	config      *config.Config
}

// NewProductService 创建商品服务
func NewProductService(repos *repository.Repositories, cfg *config.Config) ProductService {
	return &productService{
		productRepo: repos.Product,
		config:      cfg,
	}
}

func (s *productService) Create(req *CreateProductRequest) (*model.Product, error) {
	// 生成slug
	slug := generateSlug(req.Name)

	// 将规格转为JSON字符串
	specifications := ""
	if req.Specifications != nil {
		if data, err := json.Marshal(req.Specifications); err == nil {
			specifications = string(data)
		}
	}

	product := &model.Product{
		Name:            req.Name,
		Slug:            slug,
		Description:     req.Description,
		Price:           req.Price,
		OriginalPrice:   req.OriginalPrice,
		CategoryID:      req.CategoryID,
		Brand:           req.Brand,
		SKU:             req.SKU,
		Stock:           req.Stock,
		MinStock:        req.MinStock,
		Status:          model.ProductStatusActive,
		Specifications:  specifications,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
	}

	return product, s.productRepo.Create(product)
}

func (s *productService) GetByID(id uint) (*model.Product, error) {
	return s.productRepo.GetByIDWithDetails(id)
}

func (s *productService) GetBySlug(slug string) (*model.Product, error) {
	return s.productRepo.GetBySlug(slug)
}

func (s *productService) Update(id uint, req *UpdateProductRequest) (*model.Product, error) {
	product, err := s.productRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != nil {
		product.Name = *req.Name
		product.Slug = generateSlug(*req.Name)
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.OriginalPrice != nil {
		product.OriginalPrice = *req.OriginalPrice
	}
	if req.Brand != nil {
		product.Brand = *req.Brand
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Status != nil {
		product.Status = *req.Status
	}

	return product, s.productRepo.Update(product)
}

func (s *productService) Delete(id uint) error {
	return s.productRepo.Delete(id)
}

func (s *productService) List(params *ProductListRequest) (*ProductListResponse, error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = s.config.Pagination.DefaultPageSize
	}
	if params.Limit > s.config.Pagination.MaxPageSize {
		params.Limit = s.config.Pagination.MaxPageSize
	}

	offset := (params.Page - 1) * params.Limit

	repoParams := repository.ProductListParams{
		CategoryID: params.CategoryID,
		Brand:      params.Brand,
		MinPrice:   params.MinPrice,
		MaxPrice:   params.MaxPrice,
		Status:     string(params.Status),
		SortBy:     params.SortBy,
		Order:      params.Order,
		Limit:      params.Limit,
		Offset:     offset,
	}

	products, total, err := s.productRepo.List(repoParams)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &ProductListResponse{
		Products: products,
		Total:    total,
		Page:     params.Page,
		Limit:    params.Limit,
		Pages:    totalPages,
	}, nil
}

func (s *productService) Search(query string, page, limit int) (*ProductListResponse, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = s.config.Pagination.DefaultPageSize
	}
	if limit > s.config.Pagination.MaxPageSize {
		limit = s.config.Pagination.MaxPageSize
	}

	offset := (page - 1) * limit
	products, total, err := s.productRepo.Search(query, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ProductListResponse{
		Products: products,
		Total:    total,
		Page:     page,
		Limit:    limit,
		Pages:    totalPages,
	}, nil
}

func (s *productService) GetFeatured(limit int) ([]model.Product, error) {
	return s.productRepo.GetFeaturedProducts(limit)
}

func (s *productService) GetRelated(productID uint, limit int) ([]model.Product, error) {
	return s.productRepo.GetRelatedProducts(productID, limit)
}

func (s *productService) UpdateStock(id uint, quantity int) error {
	return s.productRepo.UpdateStock(id, quantity)
}

func (s *productService) IncrementViewCount(id uint) error {
	return s.productRepo.UpdateViewCount(id)
}

// cartService 购物车服务实现 - 存根
type cartService struct {
	cartRepo repository.CartRepository
	config   *config.Config
}

func NewCartService(repos *repository.Repositories, cfg *config.Config) CartService {
	return &cartService{
		cartRepo: repos.Cart,
		config:   cfg,
	}
}

func (s *cartService) AddItem(userID uint, req *AddCartItemRequest) error {
	return errors.New("not implemented")
}

func (s *cartService) GetCartItems(userID uint) ([]model.CartItem, error) {
	return nil, errors.New("not implemented")
}

func (s *cartService) UpdateQuantity(userID, productID uint, quantity int) error {
	return errors.New("not implemented")
}

func (s *cartService) RemoveItem(userID, productID uint) error {
	return errors.New("not implemented")
}

func (s *cartService) ClearCart(userID uint) error {
	return errors.New("not implemented")
}

func (s *cartService) GetCartSummary(userID uint) (*CartSummary, error) {
	return nil, errors.New("not implemented")
}

// orderService 订单服务实现 - 存根
type orderService struct {
	orderRepo repository.OrderRepository
	config    *config.Config
}

func NewOrderService(repos *repository.Repositories, cfg *config.Config) OrderService {
	return &orderService{
		orderRepo: repos.Order,
		config:    cfg,
	}
}

func (s *orderService) Create(userID uint, req *CreateOrderRequest) (*model.Order, error) {
	return nil, errors.New("not implemented")
}

func (s *orderService) GetByID(userID, orderID uint) (*model.Order, error) {
	return nil, errors.New("not implemented")
}

func (s *orderService) GetUserOrders(userID uint, page, limit int) (*OrderListResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *orderService) UpdateStatus(orderID uint, status model.OrderStatus) error {
	return errors.New("not implemented")
}

func (s *orderService) CancelOrder(userID, orderID uint) error {
	return errors.New("not implemented")
}

func (s *orderService) GetOrderStats() (map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}

// paymentService 支付服务实现 - 存根
type paymentService struct {
	paymentRepo repository.PaymentRepository
	config      *config.Config
}

func NewPaymentService(repos *repository.Repositories, cfg *config.Config) PaymentService {
	return &paymentService{
		paymentRepo: repos.Payment,
		config:      cfg,
	}
}

func (s *paymentService) ProcessPayment(orderID uint, req *PaymentRequest) (*model.Payment, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetPaymentByID(id uint) (*model.Payment, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) GetOrderPayments(orderID uint) ([]model.Payment, error) {
	return nil, errors.New("not implemented")
}

func (s *paymentService) RefundPayment(paymentID uint, amount decimal.Decimal) error {
	return errors.New("not implemented")
}

// 工具函数
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// 简化的slug生成，实际项目中应该使用更复杂的逻辑
	return fmt.Sprintf("%s-%d", slug, 1)
}
