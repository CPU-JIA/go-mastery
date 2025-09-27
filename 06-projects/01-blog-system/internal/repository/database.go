package repository

import (
	"fmt"
	"log"
	"time"

	"blog-system/internal/config"
	"blog-system/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建新的数据库连接
func NewDatabase(cfg *config.Config) (*Database, error) {
	var dialector gorm.Dialector

	switch cfg.Database.Driver {
	case "postgres":
		dialector = postgres.Open(cfg.GetDSN())
	case "sqlite":
		dialector = sqlite.Open(cfg.GetDSN())
	default:
		dialector = sqlite.Open("blog.db") // 默认SQLite
	}

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel(cfg.Logging.Level)),
	}

	// 如果是生产环境，禁用默认事务以提高性能
	if cfg.IsProduction() {
		gormConfig.SkipDefaultTransaction = true
	}

	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &Database{DB: db}, nil
}

// AutoMigrate 自动迁移数据库结构
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.Comment{},
		&model.Setting{},
		&model.Media{},
	)
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateIndexes 创建额外的索引（如果需要）
func (d *Database) CreateIndexes() error {
	// 为文章标题创建全文索引（仅PostgreSQL）
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_articles_title_fulltext ON articles USING gin(to_tsvector('english', title))").Error; err != nil {
		// SQLite不支持此语法，忽略错误
		log.Printf("创建全文索引失败（可能是SQLite）: %v", err)
	}

	// 为文章内容创建全文索引（仅PostgreSQL）
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_articles_content_fulltext ON articles USING gin(to_tsvector('english', content))").Error; err != nil {
		log.Printf("创建内容全文索引失败（可能是SQLite）: %v", err)
	}

	return nil
}

// SeedData 初始化数据
func (d *Database) SeedData() error {
	// 检查是否已有管理员用户
	var adminCount int64
	d.DB.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount)

	if adminCount == 0 {
		// 创建默认管理员用户
		admin := &model.User{
			Username: "admin",
			Email:    "admin@blog.local",
			Password: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/ZmYPL.uKhOIhfMwWa", // password: admin123
			FullName: "System Administrator",
			Role:     model.RoleAdmin,
			Status:   model.StatusActive,
		}

		if err := d.DB.Create(admin).Error; err != nil {
			return fmt.Errorf("创建管理员用户失败: %w", err)
		}

		log.Println("✓ 创建默认管理员用户: admin / admin123")
	}

	// 创建默认分类
	var categoryCount int64
	d.DB.Model(&model.Category{}).Count(&categoryCount)

	if categoryCount == 0 {
		categories := []model.Category{
			{Name: "技术", Slug: "tech", Description: "技术相关文章", Color: "#3498db"},
			{Name: "生活", Slug: "life", Description: "生活感悟和分享", Color: "#e74c3c"},
			{Name: "随笔", Slug: "notes", Description: "随笔和想法", Color: "#f39c12"},
		}

		for _, category := range categories {
			if err := d.DB.Create(&category).Error; err != nil {
				return fmt.Errorf("创建默认分类失败: %w", err)
			}
		}

		log.Println("✓ 创建默认分类")
	}

	// 创建默认标签
	var tagCount int64
	d.DB.Model(&model.Tag{}).Count(&tagCount)

	if tagCount == 0 {
		tags := []model.Tag{
			{Name: "Go", Slug: "go", Color: "#00ADD8"},
			{Name: "Web开发", Slug: "web-dev", Color: "#61DAFB"},
			{Name: "数据库", Slug: "database", Color: "#336791"},
			{Name: "前端", Slug: "frontend", Color: "#f7df1e"},
			{Name: "后端", Slug: "backend", Color: "#68217a"},
		}

		for _, tag := range tags {
			if err := d.DB.Create(&tag).Error; err != nil {
				return fmt.Errorf("创建默认标签失败: %w", err)
			}
		}

		log.Println("✓ 创建默认标签")
	}

	// 创建系统设置
	var settingCount int64
	d.DB.Model(&model.Setting{}).Count(&settingCount)

	if settingCount == 0 {
		settings := []model.Setting{
			{Key: "site_name", Value: "我的博客", Description: "网站名称", Category: "general", IsPublic: true},
			{Key: "site_description", Value: "一个基于Go语言开发的现代化博客系统", Description: "网站描述", Category: "general", IsPublic: true},
			{Key: "posts_per_page", Value: "10", Description: "每页文章数量", Category: "display", IsPublic: true},
			{Key: "allow_comments", Value: "true", Description: "是否允许评论", Category: "comments", IsPublic: true},
			{Key: "comment_moderation", Value: "false", Description: "评论是否需要审核", Category: "comments", IsPublic: false},
		}

		for _, setting := range settings {
			if err := d.DB.Create(&setting).Error; err != nil {
				return fmt.Errorf("创建系统设置失败: %w", err)
			}
		}

		log.Println("✓ 创建系统设置")
	}

	return nil
}

// getLogLevel 将字符串日志级别转换为GORM日志级别
func getLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}
