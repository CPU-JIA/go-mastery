package service

import (
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(req *RegisterRequest) (*model.User, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	RefreshToken(token string) (*AuthResponse, error)
	VerifyToken(tokenString string) (*TokenClaims, error)
	ChangePassword(userID uint, req *ChangePasswordRequest) error
}

// ProductService 商品服务接口
type ProductService interface {
	Create(req *CreateProductRequest) (*model.Product, error)
	GetByID(id uint) (*model.Product, error)
	GetBySlug(slug string) (*model.Product, error)
	Update(id uint, req *UpdateProductRequest) (*model.Product, error)
	Delete(id uint) error
	List(params *ProductListRequest) (*ProductListResponse, error)
	Search(query string, page, limit int) (*ProductListResponse, error)
	GetFeatured(limit int) ([]model.Product, error)
	GetRelated(productID uint, limit int) ([]model.Product, error)
	UpdateStock(id uint, quantity int) error
	IncrementViewCount(id uint) error
}

// CartService 购物车服务接口
type CartService interface {
	AddItem(userID uint, req *AddCartItemRequest) error
	GetCartItems(userID uint) ([]model.CartItem, error)
	UpdateQuantity(userID, productID uint, quantity int) error
	RemoveItem(userID, productID uint) error
	ClearCart(userID uint) error
	GetCartSummary(userID uint) (*CartSummary, error)
}

// OrderService 订单服务接口
type OrderService interface {
	Create(userID uint, req *CreateOrderRequest) (*model.Order, error)
	GetByID(userID, orderID uint) (*model.Order, error)
	GetUserOrders(userID uint, page, limit int) (*OrderListResponse, error)
	UpdateStatus(orderID uint, status model.OrderStatus) error
	CancelOrder(userID, orderID uint) error
	GetOrderStats() (map[string]interface{}, error)
}

// PaymentService 支付服务接口
type PaymentService interface {
	ProcessPayment(orderID uint, req *PaymentRequest) (*model.Payment, error)
	GetPaymentByID(id uint) (*model.Payment, error)
	GetOrderPayments(orderID uint) ([]model.Payment, error)
	RefundPayment(paymentID uint, amount decimal.Decimal) error
}

// DTO 结构定义
type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Phone     string `json:"phone"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	LoginID  string `json:"login_id" binding:"required"` // 用户名或邮箱
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         *model.User `json:"user"`
}

type TokenClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type CreateProductRequest struct {
	Name            string                 `json:"name" binding:"required"`
	Description     string                 `json:"description"`
	Price           decimal.Decimal        `json:"price" binding:"required"`
	OriginalPrice   decimal.Decimal        `json:"original_price"`
	CategoryID      *uint                  `json:"category_id"`
	Brand           string                 `json:"brand"`
	SKU             string                 `json:"sku" binding:"required"`
	Stock           int                    `json:"stock"`
	MinStock        int                    `json:"min_stock"`
	Specifications  map[string]interface{} `json:"specifications"`
	Images          []string               `json:"images"`
	Tags            []string               `json:"tags"`
	MetaTitle       string                 `json:"meta_title"`
	MetaDescription string                 `json:"meta_description"`
}

type UpdateProductRequest struct {
	Name            *string                 `json:"name,omitempty"`
	Description     *string                 `json:"description,omitempty"`
	Price           *decimal.Decimal        `json:"price,omitempty"`
	OriginalPrice   *decimal.Decimal        `json:"original_price,omitempty"`
	CategoryID      *uint                   `json:"category_id,omitempty"`
	Brand           *string                 `json:"brand,omitempty"`
	SKU             *string                 `json:"sku,omitempty"`
	Stock           *int                    `json:"stock,omitempty"`
	MinStock        *int                    `json:"min_stock,omitempty"`
	Status          *model.ProductStatus    `json:"status,omitempty"`
	Specifications  *map[string]interface{} `json:"specifications,omitempty"`
	Images          []string                `json:"images,omitempty"`
	Tags            []string                `json:"tags,omitempty"`
	MetaTitle       *string                 `json:"meta_title,omitempty"`
	MetaDescription *string                 `json:"meta_description,omitempty"`
}

type ProductListRequest struct {
	CategoryID *uint               `json:"category_id,omitempty"`
	Brand      string              `json:"brand,omitempty"`
	MinPrice   *decimal.Decimal    `json:"min_price,omitempty"`
	MaxPrice   *decimal.Decimal    `json:"max_price,omitempty"`
	Status     model.ProductStatus `json:"status,omitempty"`
	SortBy     string              `json:"sort_by,omitempty"`
	Order      string              `json:"order,omitempty"`
	Page       int                 `json:"page,omitempty"`
	Limit      int                 `json:"limit,omitempty"`
}

type ProductListResponse struct {
	Products []model.Product `json:"products"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	Limit    int             `json:"limit"`
	Pages    int             `json:"pages"`
}

type AddCartItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type CartSummary struct {
	Items      []model.CartItem `json:"items"`
	TotalItems int              `json:"total_items"`
	SubTotal   decimal.Decimal  `json:"subtotal"`
	Tax        decimal.Decimal  `json:"tax"`
	Total      decimal.Decimal  `json:"total"`
}

type CreateOrderRequest struct {
	ShippingAddress AddressRequest  `json:"shipping_address" binding:"required"`
	BillingAddress  *AddressRequest `json:"billing_address,omitempty"`
	PaymentMethod   string          `json:"payment_method" binding:"required"`
	CouponCode      string          `json:"coupon_code,omitempty"`
	Notes           string          `json:"notes,omitempty"`
}

type AddressRequest struct {
	Name       string `json:"name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Province   string `json:"province" binding:"required"`
	City       string `json:"city" binding:"required"`
	District   string `json:"district" binding:"required"`
	Street     string `json:"street" binding:"required"`
	PostalCode string `json:"postal_code"`
}

type OrderListResponse struct {
	Orders []model.Order `json:"orders"`
	Total  int64         `json:"total"`
	Page   int           `json:"page"`
	Limit  int           `json:"limit"`
	Pages  int           `json:"pages"`
}

type PaymentRequest struct {
	Method string          `json:"method" binding:"required"`
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

// Services 服务集合
type Services struct {
	Auth    AuthService
	Product ProductService
	Cart    CartService
	Order   OrderService
	Payment PaymentService
}

// NewServices 创建服务集合
func NewServices(repos *repository.Repositories, cfg *config.Config) *Services {
	return &Services{
		Auth:    NewAuthService(repos, cfg),
		Product: NewProductService(repos, cfg),
		Cart:    NewCartService(repos, cfg),
		Order:   NewOrderService(repos, cfg),
		Payment: NewPaymentService(repos, cfg),
	}
}

// authService 认证服务实现
type authService struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService 创建认证服务
func NewAuthService(repos *repository.Repositories, cfg *config.Config) AuthService {
	return &authService{
		userRepo: repos.User,
		config:   cfg,
	}
}

func (s *authService) Register(req *RegisterRequest) (*model.User, error) {
	// 检查用户名和邮箱是否已存在
	if existingUser, _ := s.userRepo.GetByUsername(req.Username); existingUser != nil {
		return nil, errors.New("username already exists")
	}

	if existingUser, _ := s.userRepo.GetByEmail(req.Email); existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.Security.BcryptCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Phone:    req.Phone,
		Role:     model.RoleCustomer,
		Status:   model.StatusActive,
		Profile: &model.UserProfile{
			FirstName: req.FirstName,
			LastName:  req.LastName,
		},
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// 清空密码字段
	user.Password = ""
	return user, nil
}

func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	// 查找用户
	user, err := s.userRepo.GetByUsernameOrEmail(req.LoginID)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 检查用户状态
	if user.Status != model.StatusActive {
		return nil, errors.New("account is inactive")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 更新登录记录
	s.userRepo.UpdateLastLogin(user.ID)

	// 生成 Token
	accessToken, err := s.generateToken(user, s.config.JWT.ExpiresIn)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, s.config.JWT.RefreshExpiresIn)
	if err != nil {
		return nil, err
	}

	// 清空密码字段
	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWT.ExpiresIn.Seconds()),
		User:         user,
	}, nil
}

func (s *authService) RefreshToken(token string) (*AuthResponse, error) {
	claims, err := s.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Status != model.StatusActive {
		return nil, errors.New("account is inactive")
	}

	// 生成新 Token
	accessToken, err := s.generateToken(user, s.config.JWT.ExpiresIn)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, s.config.JWT.RefreshExpiresIn)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWT.ExpiresIn.Seconds()),
		User:         user,
	}, nil
}

func (s *authService) VerifyToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *authService) ChangePassword(userID uint, req *ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// 生成新密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.Security.BcryptCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

func (s *authService) generateToken(user *model.User, expiration time.Duration) (string, error) {
	claims := &TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ecommerce-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Secret))
}
