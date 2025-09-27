package service_test

import (
	"blog-system/internal/config"
	"blog-system/internal/model"
	"blog-system/internal/repository"
	"blog-system/internal/service"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestingT represents interface for both *testing.T and *testing.B
type TestingT interface {
	Fatalf(format string, args ...interface{})
}

// setupTestDB 设置测试数据库
func setupTestDB(t TestingT) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("连接测试数据库失败: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.Comment{},
		&model.Setting{},
		&model.Media{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	return db
}

// setupTestServices 设置测试服务
func setupTestServices(t TestingT) (*service.Services, *gorm.DB) {
	db := setupTestDB(t)
	repos := repository.NewRepositories(db)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:           "test-secret",
			ExpiresIn:        time.Hour,
			RefreshExpiresIn: time.Hour * 24,
		},
	}

	services := service.NewServices(repos, cfg)
	return services, db
}

func TestAuthService_Register(t *testing.T) {
	services, _ := setupTestServices(t)

	tests := []struct {
		name    string
		req     *service.RegisterRequest
		wantErr bool
	}{
		{
			name: "成功注册",
			req: &service.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
				FullName: "Test User",
			},
			wantErr: false,
		},
		{
			name: "用户名重复",
			req: &service.RegisterRequest{
				Username: "testuser", // 重复用户名
				Email:    "test2@example.com",
				Password: "password123",
				FullName: "Test User 2",
			},
			wantErr: true,
		},
		{
			name: "邮箱重复",
			req: &service.RegisterRequest{
				Username: "testuser2",
				Email:    "test@example.com", // 重复邮箱
				Password: "password123",
				FullName: "Test User 2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := services.Auth.Register(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if user.Username != tt.req.Username {
					t.Errorf("Register() username = %v, want %v", user.Username, tt.req.Username)
				}
				if user.Email != tt.req.Email {
					t.Errorf("Register() email = %v, want %v", user.Email, tt.req.Email)
				}
				if user.Password != "" {
					t.Errorf("Register() password should be empty in response")
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	services, _ := setupTestServices(t)

	// 先注册一个用户
	registerReq := &service.RegisterRequest{
		Username: "logintest",
		Email:    "login@example.com",
		Password: "password123",
		FullName: "Login Test",
	}
	_, err := services.Auth.Register(registerReq)
	if err != nil {
		t.Fatalf("注册用户失败: %v", err)
	}

	tests := []struct {
		name    string
		req     *service.LoginRequest
		wantErr bool
	}{
		{
			name: "用户名登录成功",
			req: &service.LoginRequest{
				LoginID:  "logintest",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "邮箱登录成功",
			req: &service.LoginRequest{
				LoginID:  "login@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "密码错误",
			req: &service.LoginRequest{
				LoginID:  "logintest",
				Password: "wrongpassword",
			},
			wantErr: true,
		},
		{
			name: "用户不存在",
			req: &service.LoginRequest{
				LoginID:  "nonexistent",
				Password: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResp, err := services.Auth.Login(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if authResp.AccessToken == "" {
					t.Errorf("Login() access token should not be empty")
				}
				if authResp.RefreshToken == "" {
					t.Errorf("Login() refresh token should not be empty")
				}
				if authResp.User == nil {
					t.Errorf("Login() user should not be nil")
				}
			}
		})
	}
}

func TestArticleService_Create(t *testing.T) {
	services, db := setupTestServices(t)

	// 创建测试用户
	user := &model.User{
		Username: "author",
		Email:    "author@example.com",
		Password: "hashedpassword",
		Role:     model.RoleAuthor,
		Status:   model.StatusActive,
	}
	db.Create(user)

	// 创建测试分类
	category := &model.Category{
		Name: "测试分类",
		Slug: "test-category",
	}
	db.Create(category)

	tests := []struct {
		name    string
		req     *service.CreateArticleRequest
		userID  uint
		wantErr bool
	}{
		{
			name: "成功创建文章",
			req: &service.CreateArticleRequest{
				Title:      "测试文章",
				Content:    "这是一篇测试文章的内容。",
				Summary:    "测试摘要",
				CategoryID: category.ID,
				Tags:       []string{"Go", "测试"},
				Status:     "draft",
			},
			userID:  user.ID,
			wantErr: false,
		},
		{
			name: "无权限创建",
			req: &service.CreateArticleRequest{
				Title:   "测试文章2",
				Content: "内容",
			},
			userID:  999, // 不存在的用户ID
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			article, err := services.Article.Create(tt.req, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if article.Title != tt.req.Title {
					t.Errorf("Create() title = %v, want %v", article.Title, tt.req.Title)
				}
				if article.AuthorID != tt.userID {
					t.Errorf("Create() authorID = %v, want %v", article.AuthorID, tt.userID)
				}
				if article.Slug == "" {
					t.Errorf("Create() slug should not be empty")
				}
			}
		})
	}
}

func TestArticleService_GetBySlug(t *testing.T) {
	services, db := setupTestServices(t)

	// 创建测试数据
	user := &model.User{
		Username: "author",
		Email:    "author@example.com",
		Password: "hashedpassword",
		Role:     model.RoleAuthor,
		Status:   model.StatusActive,
	}
	db.Create(user)

	article := &model.Article{
		Title:    "测试文章",
		Slug:     "test-article",
		Content:  "测试内容",
		Summary:  "测试摘要",
		AuthorID: user.ID,
		Status:   model.StatusPublished,
	}
	now := time.Now()
	article.PublishedAt = &now
	db.Create(article)

	// 测试获取文章
	foundArticle, err := services.Article.GetBySlug("test-article")
	if err != nil {
		t.Errorf("GetBySlug() error = %v", err)
		return
	}

	if foundArticle.Title != article.Title {
		t.Errorf("GetBySlug() title = %v, want %v", foundArticle.Title, article.Title)
	}

	if foundArticle.Slug != article.Slug {
		t.Errorf("GetBySlug() slug = %v, want %v", foundArticle.Slug, article.Slug)
	}

	// 测试不存在的文章
	_, err = services.Article.GetBySlug("nonexistent")
	if err == nil {
		t.Errorf("GetBySlug() should return error for nonexistent article")
	}
}

// BenchmarkAuthService_Register 性能基准测试
func BenchmarkAuthService_Register(b *testing.B) {
	services, _ := setupTestServices(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &service.RegisterRequest{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Password: "password123",
			FullName: "Test User",
		}
		services.Auth.Register(req)
	}
}
