package main

import (
	"blog-system/internal/config"
	"blog-system/internal/handler"
	"blog-system/internal/middleware"
	"blog-system/internal/repository"
	"blog-system/internal/service"
	"blog-system/internal/validation"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("没有找到.env文件，使用默认配置")
	}

	// 初始化自定义验证器
	validation.InitCustomValidator()

	// 加载配置
	configPath := filepath.Join("configs", "config.yaml")
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	db, err := repository.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 自动迁移数据库结构
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 创建索引
	if err := db.CreateIndexes(); err != nil {
		log.Printf("创建数据库索引警告: %v", err)
	}

	// 初始化数据
	if err := db.SeedData(); err != nil {
		log.Fatalf("初始化数据失败: %v", err)
	}

	// 初始化仓储层
	repos := repository.NewRepositories(db.DB)

	// 初始化服务层
	services := service.NewServices(repos, cfg)

	// 初始化处理器
	h := handler.NewHandler(services)

	// 创建路由
	router := setupRouter(h, services, cfg)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("🚀 服务器启动在 %s", cfg.GetServerAddr())
		log.Printf("📖 API文档: http://%s/swagger/index.html", cfg.GetServerAddr())
		log.Printf("🌐 环境: %s", cfg.Server.Mode)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务器...")

	// 设置5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
	} else {
		log.Println("✅ 服务器已优雅关闭")
	}
}

// setupRouter 设置路由
func setupRouter(h *handler.Handler, services *service.Services, cfg *config.Config) *gin.Engine {
	router := gin.New()

	// 安全中间件（必须放在最前面）
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.InputSanitizer())
	router.Use(middleware.PathTraversalProtection())
	router.Use(middleware.ValidateContentType("application/json", "application/x-www-form-urlencoded", "multipart/form-data"))

	// 请求日志中间件（使用安全版本）
	router.Use(middleware.RequestLogger())

	// 全局中间件
	router.Use(middleware.RequestLoggerMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// 增强的限流中间件
	if cfg.IsProduction() {
		router.Use(middleware.RateLimiter(50, 10)) // 每秒50个请求，突发10个
	} else {
		router.Use(middleware.RateLimiter(100, 20)) // 开发环境更宽松
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API v1 路由组
	v1 := router.Group("/api/v1")
	v1.Use(middleware.PaginationMiddleware())

	// 公开路由（不需要认证）
	{
		// 认证相关 - 应用更严格的限流
		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimiter(10, 3)) // 认证接口限流更严格
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
			auth.POST("/refresh", h.RefreshToken)
		}

		// 文章公开接口
		articles := v1.Group("/articles")
		articles.Use(middleware.OptionalAuthMiddleware(services.Auth))
		{
			articles.GET("", h.ListArticles)
			articles.GET("/popular", h.GetPopularArticles)
			articles.GET("/recent", h.GetRecentArticles)
			articles.GET("/search", h.SearchArticles)
			articles.GET("/:id", h.GetArticle)
			articles.GET("/slug/:slug", h.GetArticleBySlug)
		}
	}

	// 需要认证的路由
	authenticated := v1.Group("")
	authenticated.Use(middleware.AuthMiddleware(services.Auth))
	{
		// 用户相关
		user := authenticated.Group("/user")
		{
			user.GET("/me", h.Me)
			user.PUT("/password", h.ChangePassword)
		}

		// 文章互动
		articleActions := authenticated.Group("/articles")
		{
			articleActions.POST("/:id/like", h.LikeArticle)
		}
	}

	// 作者权限路由
	authorRoutes := v1.Group("")
	authorRoutes.Use(middleware.AuthMiddleware(services.Auth))
	authorRoutes.Use(middleware.AuthorMiddleware())
	{
		articles := authorRoutes.Group("/articles")
		{
			articles.POST("", h.CreateArticle)
			articles.PUT("/:id", h.UpdateArticle)
			articles.DELETE("/:id", h.DeleteArticle)
			articles.POST("/:id/publish", h.PublishArticle)
		}
	}

	// 管理员路由
	adminRoutes := v1.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware(services.Auth))
	adminRoutes.Use(middleware.AdminMiddleware())
	{
		// 管理员专用的文章管理、用户管理等
		adminRoutes.GET("/articles", h.ListArticles) // 可以查看所有文章，包括草稿
	}

	// 静态文件服务（增加安全限制）
	router.Static("/uploads", "./uploads")
	router.StaticFile("/favicon.ico", "./assets/favicon.ico")

	// 404处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "接口不存在",
		})
	})

	return router
}
