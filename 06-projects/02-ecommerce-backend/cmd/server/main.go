package main

import (
	"context"
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/handler"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 连接数据库
	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	if err := migrateDatabase(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 初始化仓储层
	repos := repository.NewRepositories(db)

	// 初始化服务层
	services := service.NewServices(repos, cfg)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(services.Auth)
	productHandler := handler.NewProductHandler(services.Product)

	// 创建Gin实例
	router := gin.New()

	// 添加中间件
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	// 根路由
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":    "E-commerce Backend API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"database":  "connected",
		})
	})

	// API 路由组
	api := router.Group("/api/v1")
	{
		// 认证路由（不需要身份验证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 商品路由（公共访问）
		products := api.Group("/products")
		{
			products.GET("", productHandler.GetProducts)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/featured", productHandler.GetFeaturedProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.GET("/:id/related", productHandler.GetRelatedProducts)
			products.GET("/slug/:slug", productHandler.GetProductBySlug)
		}

		// 需要认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.Auth(services.Auth))
		{
			// 用户相关
			user := authenticated.Group("/user")
			{
				user.PUT("/password", authHandler.ChangePassword)
			}

			// 购物车相关
			// cart := authenticated.Group("/cart")
			// {
			// 	cart.GET("", cartHandler.GetCart)
			// 	cart.POST("", cartHandler.AddToCart)
			// 	cart.PUT("/items/:product_id", cartHandler.UpdateQuantity)
			// 	cart.DELETE("/items/:product_id", cartHandler.RemoveFromCart)
			// 	cart.DELETE("", cartHandler.ClearCart)
			// }

			// 订单相关
			// orders := authenticated.Group("/orders")
			// {
			// 	orders.GET("", orderHandler.GetOrders)
			// 	orders.POST("", orderHandler.CreateOrder)
			// 	orders.GET("/:id", orderHandler.GetOrder)
			// 	orders.PUT("/:id/cancel", orderHandler.CancelOrder)
			// }

			// 支付相关
			// payments := authenticated.Group("/payments")
			// {
			// 	payments.POST("/orders/:order_id", paymentHandler.ProcessPayment)
			// 	payments.GET("/:id", paymentHandler.GetPayment)
			// }
		}

		// 管理员路由
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(services.Auth))
		admin.Use(middleware.RequireRole("admin"))
		{
			// 商品管理
			// adminProducts := admin.Group("/products")
			// {
			// 	adminProducts.POST("", productHandler.CreateProduct)
			// 	adminProducts.PUT("/:id", productHandler.UpdateProduct)
			// 	adminProducts.DELETE("/:id", productHandler.DeleteProduct)
			// 	adminProducts.PUT("/:id/stock", productHandler.UpdateStock)
			// }

			// 订单管理
			// adminOrders := admin.Group("/orders")
			// {
			// 	adminOrders.GET("", orderHandler.GetAllOrders)
			// 	adminOrders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
			// 	adminOrders.GET("/stats", orderHandler.GetOrderStats)
			// }

			// 用户管理
			// adminUsers := admin.Group("/users")
			// {
			// 	adminUsers.GET("", userHandler.GetUsers)
			// 	adminUsers.GET("/:id", userHandler.GetUser)
			// 	adminUsers.PUT("/:id/status", userHandler.UpdateUserStatus)
			// }
		}
	}

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 优雅启动和关闭
	go func() {
		log.Printf("🚀 E-commerce Backend API server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Printf("📝 API Documentation: http://localhost:%d/", cfg.Server.Port)
		log.Printf("🔍 Health Check: http://localhost:%d/health", cfg.Server.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// 5秒超时关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited gracefully")
}

func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 根据配置选择数据库驱动
	switch cfg.Database.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
			cfg.Database.Postgres.Host,
			cfg.Database.Postgres.Port,
			cfg.Database.Postgres.User,
			cfg.Database.Postgres.Password,
			cfg.Database.Postgres.DBName,
			cfg.Database.Postgres.SSLMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.Database.SQLite.Path), gormConfig)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", cfg.Database.Driver, err)
	}

	// 配置连接池（仅对PostgreSQL有效）
	if cfg.Database.Driver == "postgres" {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, err
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	log.Printf("✅ Connected to %s database successfully", cfg.Database.Driver)
	return db, nil
}

func migrateDatabase(db *gorm.DB) error {
	log.Println("🔄 Running database migrations...")

	err := db.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
		&model.Address{},
		&model.Category{},
		&model.Product{},
		&model.ProductImage{},
		&model.ProductTag{},
		&model.CartItem{},
		&model.Order{},
		&model.OrderItem{},
		&model.Payment{},
		&model.Coupon{},
		&model.Review{},
		&model.Wishlist{},
		&model.InventoryLog{},
		&model.Setting{},
	)

	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("✅ Database migration completed successfully")

	// 创建示例数据
	if err := createSampleData(db); err != nil {
		log.Printf("⚠️  Warning: Failed to create sample data: %v", err)
	}

	return nil
}

func createSampleData(db *gorm.DB) error {
	// 检查是否已有数据
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil // 已有数据，跳过
	}

	log.Println("🌱 Creating sample data...")

	// 创建示例用户
	users := []model.User{
		{
			Username: "admin",
			Email:    "admin@ecommerce.com",
			Password: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj.N5/oQi8j2", // admin123
			Role:     model.RoleAdmin,
			Status:   model.StatusActive,
			Profile: &model.UserProfile{
				FirstName: "Admin",
				LastName:  "User",
			},
		},
		{
			Username: "customer",
			Email:    "customer@example.com",
			Password: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj.N5/oQi8j2", // customer123
			Role:     model.RoleCustomer,
			Status:   model.StatusActive,
			Profile: &model.UserProfile{
				FirstName: "Test",
				LastName:  "Customer",
			},
		},
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			return err
		}
	}

	// 创建示例分类
	categories := []model.Category{
		{Name: "Electronics", Slug: "electronics", Description: "Electronic devices and gadgets", IsActive: true},
		{Name: "Clothing", Slug: "clothing", Description: "Fashion and apparel", IsActive: true},
		{Name: "Home & Garden", Slug: "home-garden", Description: "Home improvement and gardening", IsActive: true},
		{Name: "Sports", Slug: "sports", Description: "Sports and outdoor equipment", IsActive: true},
	}

	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			return err
		}
	}

	log.Println("✅ Sample data created successfully")
	log.Println("🔑 Default accounts:")
	log.Println("   Admin: admin / admin123")
	log.Println("   Customer: customer / customer123")

	return nil
}
