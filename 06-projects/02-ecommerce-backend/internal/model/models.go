package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 用户角色枚举
type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
	RoleSeller   UserRole = "seller"
)

// 用户状态枚举
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusInactive  UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// 订单状态枚举
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

// 支付状态枚举
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

// 商品状态枚举
type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusOutOfStock   ProductStatus = "out_of_stock"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

// 优惠券类型枚举
type CouponType string

const (
	CouponTypePercentage CouponType = "percentage"
	CouponTypeFixed      CouponType = "fixed"
)

// User 用户模型
type User struct {
	gorm.Model
	Username string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email    string     `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password string     `gorm:"size:255;not null" json:"-"` // 不在JSON中显示
	Phone    string     `gorm:"size:20" json:"phone"`
	Role     UserRole   `gorm:"type:varchar(20);default:'customer'" json:"role"`
	Status   UserStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	Avatar   string     `gorm:"size:500" json:"avatar"`

	// 用户资料
	Profile *UserProfile `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"profile,omitempty"`

	// 关联数据
	Addresses []Address  `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"addresses,omitempty"`
	CartItems []CartItem `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"cart_items,omitempty"`
	Orders    []Order    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"orders,omitempty"`
	Reviews   []Review   `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"reviews,omitempty"`
	Wishlists []Wishlist `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"wishlists,omitempty"`

	// 审计字段
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LoginCount  int        `gorm:"default:0" json:"login_count"`
}

// UserProfile 用户详细资料
type UserProfile struct {
	gorm.Model
	UserID    uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	FirstName string     `gorm:"size:50" json:"first_name"`
	LastName  string     `gorm:"size:50" json:"last_name"`
	Gender    string     `gorm:"size:10" json:"gender"`
	Birthday  *time.Time `json:"birthday,omitempty"`
	Bio       string     `gorm:"type:text" json:"bio"`
}

// Address 用户地址
type Address struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Name       string `gorm:"size:100;not null" json:"name"`
	Phone      string `gorm:"size:20;not null" json:"phone"`
	Province   string `gorm:"size:50;not null" json:"province"`
	City       string `gorm:"size:50;not null" json:"city"`
	District   string `gorm:"size:50;not null" json:"district"`
	Street     string `gorm:"size:200;not null" json:"street"`
	PostalCode string `gorm:"size:20" json:"postal_code"`
	IsDefault  bool   `gorm:"default:false" json:"is_default"`

	// 关联
	User *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// Category 商品分类
type Category struct {
	gorm.Model
	Name        string `gorm:"size:100;not null" json:"name"`
	Slug        string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `gorm:"index" json:"parent_id,omitempty"`
	Level       int    `gorm:"default:1" json:"level"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	ImageURL    string `gorm:"size:500" json:"image_url"`

	// 自关联
	Parent   *Category  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"children,omitempty"`

	// 商品关联
	Products []Product `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"products,omitempty"`
}

// Product 商品模型
type Product struct {
	gorm.Model
	Name          string          `gorm:"size:200;not null" json:"name"`
	Slug          string          `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Description   string          `gorm:"type:text" json:"description"`
	Price         decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	OriginalPrice decimal.Decimal `gorm:"type:decimal(10,2)" json:"original_price"`
	CategoryID    *uint           `gorm:"index" json:"category_id,omitempty"`
	Brand         string          `gorm:"size:100" json:"brand"`
	SKU           string          `gorm:"uniqueIndex;size:100;not null" json:"sku"`
	Barcode       string          `gorm:"size:100" json:"barcode"`
	Status        ProductStatus   `gorm:"type:varchar(20);default:'active'" json:"status"`

	// 库存信息
	Stock      int  `gorm:"default:0" json:"stock"`
	MinStock   int  `gorm:"default:0" json:"min_stock"`
	MaxStock   int  `gorm:"default:999999" json:"max_stock"`
	TrackStock bool `gorm:"default:true" json:"track_stock"`

	// 规格和属性
	Specifications string `gorm:"type:json" json:"specifications"` // JSON 格式存储规格
	Weight         int    `gorm:"default:0" json:"weight"`         // 克
	Dimensions     string `gorm:"size:100" json:"dimensions"`      // 长x宽x高

	// SEO 信息
	MetaTitle       string `gorm:"size:200" json:"meta_title"`
	MetaDescription string `gorm:"size:500" json:"meta_description"`
	MetaKeywords    string `gorm:"size:200" json:"meta_keywords"`

	// 统计数据
	ViewCount   int             `gorm:"default:0" json:"view_count"`
	SalesCount  int             `gorm:"default:0" json:"sales_count"`
	Rating      decimal.Decimal `gorm:"type:decimal(3,2);default:0" json:"rating"`
	ReviewCount int             `gorm:"default:0" json:"review_count"`

	// 时间字段
	PublishedAt *time.Time `json:"published_at,omitempty"`

	// 关联数据
	Category   *Category      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"category,omitempty"`
	Images     []ProductImage `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"images,omitempty"`
	Tags       []ProductTag   `gorm:"many2many:product_tag_relations;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"tags,omitempty"`
	CartItems  []CartItem     `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"cart_items,omitempty"`
	OrderItems []OrderItem    `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order_items,omitempty"`
	Reviews    []Review       `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"reviews,omitempty"`
	Wishlists  []Wishlist     `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"wishlists,omitempty"`
}

// ProductImage 商品图片
type ProductImage struct {
	gorm.Model
	ProductID uint   `gorm:"not null;index" json:"product_id"`
	ImageURL  string `gorm:"size:500;not null" json:"image_url"`
	AltText   string `gorm:"size:200" json:"alt_text"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	IsPrimary bool   `gorm:"default:false" json:"is_primary"`

	// 关联
	Product *Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product,omitempty"`
}

// ProductTag 商品标签
type ProductTag struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Slug  string `gorm:"uniqueIndex;size:50;not null" json:"slug"`
	Color string `gorm:"size:20" json:"color"`

	// 多对多关联通过中间表
	Products []Product `gorm:"many2many:product_tag_relations;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"products,omitempty"`
}

// CartItem 购物车项
type CartItem struct {
	gorm.Model
	UserID    uint            `gorm:"not null;index" json:"user_id"`
	ProductID uint            `gorm:"not null;index" json:"product_id"`
	Quantity  int             `gorm:"not null;default:1" json:"quantity"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`

	// 关联
	User    *User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Product *Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product,omitempty"`
}

// Order 订单模型
type Order struct {
	gorm.Model
	OrderNumber   string        `gorm:"uniqueIndex;size:50;not null" json:"order_number"`
	UserID        *uint         `gorm:"index" json:"user_id,omitempty"`
	Status        OrderStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	PaymentStatus PaymentStatus `gorm:"type:varchar(20);default:'pending'" json:"payment_status"`
	PaymentMethod string        `gorm:"size:50" json:"payment_method"`

	// 金额信息
	SubTotal       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	ShippingFee    decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"shipping_fee"`
	DiscountAmount decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	TaxAmount      decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	TotalAmount    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`

	// 地址信息 - 使用JSON存储
	ShippingAddress string `gorm:"type:json" json:"shipping_address"`
	BillingAddress  string `gorm:"type:json" json:"billing_address"`

	// 物流信息
	TrackingNumber string     `gorm:"size:100" json:"tracking_number"`
	Carrier        string     `gorm:"size:100" json:"carrier"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`

	// 备注和元数据
	Notes    string `gorm:"type:text" json:"notes"`
	Metadata string `gorm:"type:json" json:"metadata"`

	// 关联数据
	User     *User       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"user,omitempty"`
	Items    []OrderItem `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"items,omitempty"`
	Payments []Payment   `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"payments,omitempty"`

	// 使用的优惠券
	CouponID *uint   `gorm:"index" json:"coupon_id,omitempty"`
	Coupon   *Coupon `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"coupon,omitempty"`
}

// OrderItem 订单项
type OrderItem struct {
	gorm.Model
	OrderID   uint            `gorm:"not null;index" json:"order_id"`
	ProductID uint            `gorm:"not null;index" json:"product_id"`
	Quantity  int             `gorm:"not null;default:1" json:"quantity"`
	Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	Total     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total"`

	// 快照数据
	ProductSnapshot string `gorm:"type:json" json:"product_snapshot"` // 商品信息快照

	// 关联
	Order   *Order   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order,omitempty"`
	Product *Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product,omitempty"`
}

// Payment 支付记录
type Payment struct {
	gorm.Model
	OrderID         uint            `gorm:"not null;index" json:"order_id"`
	Method          string          `gorm:"size:50;not null" json:"method"` // alipay, wechat, credit_card, paypal
	Amount          decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status          PaymentStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TransactionID   string          `gorm:"size:100" json:"transaction_id"`
	GatewayResponse string          `gorm:"type:json" json:"gateway_response"`
	FailureReason   string          `gorm:"type:text" json:"failure_reason"`

	// 关联
	Order *Order `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order,omitempty"`
}

// Coupon 优惠券
type Coupon struct {
	gorm.Model
	Code              string          `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name              string          `gorm:"size:100;not null" json:"name"`
	Description       string          `gorm:"type:text" json:"description"`
	Type              CouponType      `gorm:"type:varchar(20);not null" json:"type"`
	Value             decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"value"`
	MinOrderAmount    decimal.Decimal `gorm:"type:decimal(10,2);default:0" json:"min_order_amount"`
	MaxDiscountAmount decimal.Decimal `gorm:"type:decimal(10,2)" json:"max_discount_amount"`
	UsageLimit        int             `gorm:"default:0" json:"usage_limit"` // 0表示无限制
	UsageCount        int             `gorm:"default:0" json:"usage_count"`
	StartDate         time.Time       `gorm:"not null" json:"start_date"`
	EndDate           time.Time       `gorm:"not null" json:"end_date"`
	IsActive          bool            `gorm:"default:true" json:"is_active"`

	// 关联
	Orders []Order `gorm:"foreignKey:CouponID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"orders,omitempty"`
}

// Review 商品评价
type Review struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	ProductID  uint   `gorm:"not null;index" json:"product_id"`
	Rating     int    `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Title      string `gorm:"size:200" json:"title"`
	Content    string `gorm:"type:text" json:"content"`
	Images     string `gorm:"type:json" json:"images"`          // JSON数组存储图片URL
	IsVerified bool   `gorm:"default:false" json:"is_verified"` // 是否已购买验证

	// 关联
	User    *User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Product *Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product,omitempty"`
}

// Wishlist 愿望单
type Wishlist struct {
	gorm.Model
	UserID    uint `gorm:"not null;index" json:"user_id"`
	ProductID uint `gorm:"not null;index" json:"product_id"`

	// 关联
	User    *User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
	Product *Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product,omitempty"`
}

// InventoryLog 库存记录
type InventoryLog struct {
	gorm.Model
	ProductID   uint   `gorm:"not null;index" json:"product_id"`
	Type        string `gorm:"size:20;not null" json:"type"` // in, out, adjust
	Quantity    int    `gorm:"not null" json:"quantity"`
	Reason      string `gorm:"size:200" json:"reason"`
	Reference   string `gorm:"size:100" json:"reference"` // 关联的订单号或其他参考
	OperatorID  *uint  `json:"operator_id,omitempty"`
	PreviousQty int    `gorm:"not null" json:"previous_qty"`
	CurrentQty  int    `gorm:"not null" json:"current_qty"`

	// 关联
	Product  *Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"product,omitempty"`
	Operator *User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"operator,omitempty"`
}

// Setting 系统设置
type Setting struct {
	gorm.Model
	Key         string `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value       string `gorm:"type:text" json:"value"`
	Type        string `gorm:"size:20;default:'string'" json:"type"` // string, number, boolean, json
	Description string `gorm:"size:500" json:"description"`
	IsPublic    bool   `gorm:"default:false" json:"is_public"` // 是否可以通过API访问
}
