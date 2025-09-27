package repository

import (
	"ecommerce-backend/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	GetByUsernameOrEmail(identifier string) (*model.User, error)
	Update(user *model.User) error
	UpdateProfile(userID uint, profile *model.UserProfile) error
	Delete(id uint) error
	List(offset, limit int) ([]model.User, int64, error)
	UpdateLastLogin(userID uint) error
}

// AddressRepository 地址仓储接口
type AddressRepository interface {
	Create(address *model.Address) error
	GetByID(id uint) (*model.Address, error)
	GetByUserID(userID uint) ([]model.Address, error)
	GetDefaultByUserID(userID uint) (*model.Address, error)
	Update(address *model.Address) error
	Delete(id uint) error
	SetDefault(userID, addressID uint) error
}

// CategoryRepository 分类仓储接口
type CategoryRepository interface {
	Create(category *model.Category) error
	GetByID(id uint) (*model.Category, error)
	GetBySlug(slug string) (*model.Category, error)
	GetRootCategories() ([]model.Category, error)
	GetByParentID(parentID uint) ([]model.Category, error)
	GetWithChildren(id uint) (*model.Category, error)
	Update(category *model.Category) error
	Delete(id uint) error
	List(offset, limit int) ([]model.Category, int64, error)
	GetActiveCategories() ([]model.Category, error)
}

// ProductRepository 商品仓储接口
type ProductRepository interface {
	Create(product *model.Product) error
	GetByID(id uint) (*model.Product, error)
	GetByIDWithDetails(id uint) (*model.Product, error)
	GetBySlug(slug string) (*model.Product, error)
	GetBySKU(sku string) (*model.Product, error)
	Update(product *model.Product) error
	Delete(id uint) error
	List(params ProductListParams) ([]model.Product, int64, error)
	Search(query string, limit, offset int) ([]model.Product, int64, error)
	GetByCategoryID(categoryID uint, limit, offset int) ([]model.Product, int64, error)
	GetFeaturedProducts(limit int) ([]model.Product, error)
	GetRelatedProducts(productID uint, limit int) ([]model.Product, error)
	UpdateStock(id uint, quantity int) error
	UpdateViewCount(id uint) error
	UpdateRating(productID uint) error
	GetLowStockProducts(threshold int) ([]model.Product, error)
	BulkUpdateStatus(ids []uint, status model.ProductStatus) error
}

// ProductListParams 商品列表查询参数
type ProductListParams struct {
	CategoryID *uint
	Brand      string
	MinPrice   *decimal.Decimal
	MaxPrice   *decimal.Decimal
	Status     string
	SortBy     string // price, name, created_at, rating, sales
	Order      string // asc, desc
	Limit      int
	Offset     int
}

// ProductImageRepository 商品图片仓储接口
type ProductImageRepository interface {
	Create(image *model.ProductImage) error
	GetByProductID(productID uint) ([]model.ProductImage, error)
	Update(image *model.ProductImage) error
	Delete(id uint) error
	SetPrimary(productID, imageID uint) error
	DeleteByProductID(productID uint) error
}

// ProductTagRepository 商品标签仓储接口
type ProductTagRepository interface {
	Create(tag *model.ProductTag) error
	GetByID(id uint) (*model.ProductTag, error)
	GetByName(name string) (*model.ProductTag, error)
	GetBySlug(slug string) (*model.ProductTag, error)
	Update(tag *model.ProductTag) error
	Delete(id uint) error
	List() ([]model.ProductTag, error)
	GetByProductID(productID uint) ([]model.ProductTag, error)
	AddToProduct(productID uint, tagIDs []uint) error
	RemoveFromProduct(productID uint, tagIDs []uint) error
}

// CartRepository 购物车仓储接口
type CartRepository interface {
	AddItem(item *model.CartItem) error
	GetByUserID(userID uint) ([]model.CartItem, error)
	GetItem(userID, productID uint) (*model.CartItem, error)
	UpdateQuantity(userID, productID uint, quantity int) error
	RemoveItem(userID, productID uint) error
	ClearCart(userID uint) error
	GetTotalAmount(userID uint) (decimal.Decimal, error)
	GetItemCount(userID uint) (int64, error)
}

// OrderRepository 订单仓储接口
type OrderRepository interface {
	Create(order *model.Order) error
	GetByID(id uint) (*model.Order, error)
	GetByOrderNumber(orderNumber string) (*model.Order, error)
	GetByUserID(userID uint, limit, offset int) ([]model.Order, int64, error)
	Update(order *model.Order) error
	UpdateStatus(id uint, status model.OrderStatus) error
	UpdatePaymentStatus(id uint, status model.PaymentStatus) error
	Delete(id uint) error
	List(params OrderListParams) ([]model.Order, int64, error)
	GetOrderStats() (map[string]interface{}, error)
	GetSalesReport(startDate, endDate string) ([]map[string]interface{}, error)
}

// OrderListParams 订单列表查询参数
type OrderListParams struct {
	UserID        *uint
	Status        string
	PaymentStatus string
	StartDate     string
	EndDate       string
	SortBy        string // created_at, total_amount, status
	Order         string // asc, desc
	Limit         int
	Offset        int
}

// OrderItemRepository 订单项仓储接口
type OrderItemRepository interface {
	CreateBatch(items []model.OrderItem) error
	GetByOrderID(orderID uint) ([]model.OrderItem, error)
	Update(item *model.OrderItem) error
	Delete(id uint) error
}

// PaymentRepository 支付仓储接口
type PaymentRepository interface {
	Create(payment *model.Payment) error
	GetByID(id uint) (*model.Payment, error)
	GetByOrderID(orderID uint) ([]model.Payment, error)
	GetByTransactionID(transactionID string) (*model.Payment, error)
	Update(payment *model.Payment) error
	UpdateStatus(id uint, status model.PaymentStatus) error
	List(limit, offset int) ([]model.Payment, int64, error)
}

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	Create(coupon *model.Coupon) error
	GetByID(id uint) (*model.Coupon, error)
	GetByCode(code string) (*model.Coupon, error)
	Update(coupon *model.Coupon) error
	Delete(id uint) error
	List(offset, limit int) ([]model.Coupon, int64, error)
	GetActiveCoupons() ([]model.Coupon, error)
	IncrementUsage(id uint) error
	ValidateCoupon(code string, orderAmount decimal.Decimal) (*model.Coupon, error)
}

// ReviewRepository 评价仓储接口
type ReviewRepository interface {
	Create(review *model.Review) error
	GetByID(id uint) (*model.Review, error)
	GetByProductID(productID uint, limit, offset int) ([]model.Review, int64, error)
	GetByUserID(userID uint, limit, offset int) ([]model.Review, int64, error)
	Update(review *model.Review) error
	Delete(id uint) error
	GetProductRatingStats(productID uint) (map[string]interface{}, error)
	CanUserReview(userID, productID uint) (bool, error)
}

// WishlistRepository 愿望单仓储接口
type WishlistRepository interface {
	Add(userID, productID uint) error
	Remove(userID, productID uint) error
	GetByUserID(userID uint, limit, offset int) ([]model.Wishlist, int64, error)
	Exists(userID, productID uint) (bool, error)
	GetCount(userID uint) (int64, error)
}

// InventoryLogRepository 库存记录仓储接口
type InventoryLogRepository interface {
	Create(log *model.InventoryLog) error
	GetByProductID(productID uint, limit, offset int) ([]model.InventoryLog, int64, error)
	List(params InventoryLogParams) ([]model.InventoryLog, int64, error)
}

// InventoryLogParams 库存记录查询参数
type InventoryLogParams struct {
	ProductID  *uint
	Type       string
	StartDate  string
	EndDate    string
	OperatorID *uint
	Limit      int
	Offset     int
}

// SettingRepository 设置仓储接口
type SettingRepository interface {
	Get(key string) (*model.Setting, error)
	Set(key, value, settingType, description string, isPublic bool) error
	Update(setting *model.Setting) error
	Delete(key string) error
	GetAll() ([]model.Setting, error)
	GetPublic() ([]model.Setting, error)
	GetByType(settingType string) ([]model.Setting, error)
}

// Repositories 仓储集合
type Repositories struct {
	User         UserRepository
	Address      AddressRepository
	Category     CategoryRepository
	Product      ProductRepository
	ProductImage ProductImageRepository
	ProductTag   ProductTagRepository
	Cart         CartRepository
	Order        OrderRepository
	OrderItem    OrderItemRepository
	Payment      PaymentRepository
	Coupon       CouponRepository
	Review       ReviewRepository
	Wishlist     WishlistRepository
	InventoryLog InventoryLogRepository
	Setting      SettingRepository
}

// NewRepositories 创建仓储集合
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:         NewUserRepository(db),
		Address:      NewAddressRepository(db),
		Category:     NewCategoryRepository(db),
		Product:      NewProductRepository(db),
		ProductImage: NewProductImageRepository(db),
		ProductTag:   NewProductTagRepository(db),
		Cart:         NewCartRepository(db),
		Order:        NewOrderRepository(db),
		OrderItem:    NewOrderItemRepository(db),
		Payment:      NewPaymentRepository(db),
		Coupon:       NewCouponRepository(db),
		Review:       NewReviewRepository(db),
		Wishlist:     NewWishlistRepository(db),
		InventoryLog: NewInventoryLogRepository(db),
		Setting:      NewSettingRepository(db),
	}
}
