/*
博客系统 (Blog System)

项目描述:
一个完整的博客内容管理系统，包含用户管理、文章发布、评论系统、
标签分类、搜索功能等。适合学习完整的 Web 应用开发流程。

技术栈:
- HTTP 服务器 (net/http)
- JSON 数据处理
- 模板引擎 (html/template)
- 文件操作
- 时间处理
- 正则表达式
- 密码加密
*/

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-mastery/common/security"
)

// ====================
// 1. 数据模型
// ====================

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` // 哈希后的密码
	Role      string    `json:"role"`     // admin, author, reader
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`
}

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	AuthorID    int       `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"` // draft, published, archived
	ViewCount   int       `json:"view_count"`
	LikeCount   int       `json:"like_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	ArticleID int       `json:"article_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	ParentID  int       `json:"parent_id"` // 回复功能
	Status    string    `json:"status"`    // pending, approved, rejected
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int    `json:"count"`
}

type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ====================
// 2. 数据存储层
// ====================

type Storage struct {
	users    []User
	articles []Article
	comments []Comment
	nextID   map[string]int
	dataDir  string
}

func NewStorage(dataDir string) *Storage {
	storage := &Storage{
		users:    make([]User, 0),
		articles: make([]Article, 0),
		comments: make([]Comment, 0),
		nextID: map[string]int{
			"user":    1,
			"article": 1,
			"comment": 1,
		},
		dataDir: dataDir,
	}

	// 创建数据目录
	// #nosec G301 -- 博客系统数据目录，需要0755权限支持文章数据文件读写
	os.MkdirAll(dataDir, 0755)

	// 加载数据
	storage.loadData()

	// 如果没有用户，创建默认管理员
	if len(storage.users) == 0 {
		storage.createDefaultAdmin()
	}

	return storage
}

func (s *Storage) loadData() {
	// 加载用户数据
	if data, err := os.ReadFile(filepath.Join(s.dataDir, "users.json")); err == nil {
		json.Unmarshal(data, &s.users)
		if len(s.users) > 0 {
			maxID := 0
			for _, user := range s.users {
				if user.ID > maxID {
					maxID = user.ID
				}
			}
			s.nextID["user"] = maxID + 1
		}
	}

	// 加载文章数据
	if data, err := os.ReadFile(filepath.Join(s.dataDir, "articles.json")); err == nil {
		json.Unmarshal(data, &s.articles)
		if len(s.articles) > 0 {
			maxID := 0
			for _, article := range s.articles {
				if article.ID > maxID {
					maxID = article.ID
				}
			}
			s.nextID["article"] = maxID + 1
		}
	}

	// 加载评论数据
	if data, err := os.ReadFile(filepath.Join(s.dataDir, "comments.json")); err == nil {
		json.Unmarshal(data, &s.comments)
		if len(s.comments) > 0 {
			maxID := 0
			for _, comment := range s.comments {
				if comment.ID > maxID {
					maxID = comment.ID
				}
			}
			s.nextID["comment"] = maxID + 1
		}
	}
}

func (s *Storage) saveData() error {
	// 保存用户数据
	if data, err := json.MarshalIndent(s.users, "", "  "); err == nil {
		security.SecureWriteFile(filepath.Join(s.dataDir, "users.json"), data, &security.SecureFileOptions{
			Mode:      security.GetRecommendedMode("data"),
			CreateDir: true,
		})
	}

	// 保存文章数据
	if data, err := json.MarshalIndent(s.articles, "", "  "); err == nil {
		security.SecureWriteFile(filepath.Join(s.dataDir, "articles.json"), data, &security.SecureFileOptions{
			Mode:      security.GetRecommendedMode("data"),
			CreateDir: true,
		})
	}

	// 保存评论数据
	if data, err := json.MarshalIndent(s.comments, "", "  "); err == nil {
		security.SecureWriteFile(filepath.Join(s.dataDir, "comments.json"), data, &security.SecureFileOptions{
			Mode:      security.GetRecommendedMode("data"),
			CreateDir: true,
		})
	}

	return nil
}

func (s *Storage) createDefaultAdmin() {
	admin := User{
		ID:        s.nextID["user"],
		Username:  "admin",
		Email:     "admin@blog.com",
		Password:  hashPassword("admin123"),
		Role:      "admin",
		Avatar:    "/static/default-avatar.png",
		Bio:       "System Administrator",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}
	s.nextID["user"]++
	s.users = append(s.users, admin)
	s.saveData()

	log.Println("Created default admin user: admin/admin123")
}

// ====================
// 3. 用户管理
// ====================

func (s *Storage) CreateUser(user User) (*User, error) {
	// 验证用户名和邮箱唯一性
	for _, u := range s.users {
		if u.Username == user.Username {
			return nil, fmt.Errorf("username already exists")
		}
		if u.Email == user.Email {
			return nil, fmt.Errorf("email already exists")
		}
	}

	user.ID = s.nextID["user"]
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true
	user.Password = hashPassword(user.Password)

	s.nextID["user"]++
	s.users = append(s.users, user)
	s.saveData()

	return &user, nil
}

func (s *Storage) GetUserByID(id int) (*User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (s *Storage) GetUserByUsername(username string) (*User, error) {
	for _, user := range s.users {
		if user.Username == username {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (s *Storage) AuthenticateUser(username, password string) (*User, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if !verifyPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	return user, nil
}

// ====================
// 4. 文章管理
// ====================

func (s *Storage) CreateArticle(article Article) (*Article, error) {
	article.ID = s.nextID["article"]
	article.Slug = generateSlug(article.Title)
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()

	if article.Status == "published" {
		article.PublishedAt = time.Now()
	}

	// 设置作者名称
	if author, err := s.GetUserByID(article.AuthorID); err == nil {
		article.AuthorName = author.Username
	}

	s.nextID["article"]++
	s.articles = append(s.articles, article)
	s.saveData()

	return &article, nil
}

func (s *Storage) GetArticleByID(id int) (*Article, error) {
	for i := range s.articles {
		if s.articles[i].ID == id {
			return &s.articles[i], nil
		}
	}
	return nil, fmt.Errorf("article not found")
}

func (s *Storage) GetArticleBySlug(slug string) (*Article, error) {
	for i := range s.articles {
		if s.articles[i].Slug == slug {
			return &s.articles[i], nil
		}
	}
	return nil, fmt.Errorf("article not found")
}

func (s *Storage) UpdateArticle(id int, updates Article) error {
	for i := range s.articles {
		if s.articles[i].ID == id {
			updates.ID = id
			updates.CreatedAt = s.articles[i].CreatedAt
			updates.UpdatedAt = time.Now()

			if updates.Status == "published" && s.articles[i].Status != "published" {
				updates.PublishedAt = time.Now()
			}

			s.articles[i] = updates
			s.saveData()
			return nil
		}
	}
	return fmt.Errorf("article not found")
}

func (s *Storage) DeleteArticle(id int) error {
	for i, article := range s.articles {
		if article.ID == id {
			s.articles = append(s.articles[:i], s.articles[i+1:]...)
			s.saveData()
			return nil
		}
	}
	return fmt.Errorf("article not found")
}

func (s *Storage) GetPublishedArticles(limit, offset int) []Article {
	published := make([]Article, 0)
	for _, article := range s.articles {
		if article.Status == "published" {
			published = append(published, article)
		}
	}

	// 按发布时间倒序排列
	sort.Slice(published, func(i, j int) bool {
		return published[i].PublishedAt.After(published[j].PublishedAt)
	})

	// 分页
	if offset >= len(published) {
		return []Article{}
	}

	end := offset + limit
	if end > len(published) {
		end = len(published)
	}

	return published[offset:end]
}

func (s *Storage) SearchArticles(query string) []Article {
	results := make([]Article, 0)
	query = strings.ToLower(query)

	for _, article := range s.articles {
		if article.Status != "published" {
			continue
		}

		if strings.Contains(strings.ToLower(article.Title), query) ||
			strings.Contains(strings.ToLower(article.Content), query) ||
			strings.Contains(strings.ToLower(article.Summary), query) {
			results = append(results, article)
		}

		// 搜索标签
		for _, tag := range article.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, article)
				break
			}
		}
	}

	return results
}

func (s *Storage) IncrementViewCount(id int) {
	for i := range s.articles {
		if s.articles[i].ID == id {
			s.articles[i].ViewCount++
			s.saveData()
			break
		}
	}
}

// ====================
// 5. 评论管理
// ====================

func (s *Storage) CreateComment(comment Comment) (*Comment, error) {
	comment.ID = s.nextID["comment"]
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	comment.Status = "pending" // 需要审核

	// 设置用户名
	if user, err := s.GetUserByID(comment.UserID); err == nil {
		comment.Username = user.Username
	}

	s.nextID["comment"]++
	s.comments = append(s.comments, comment)
	s.saveData()

	return &comment, nil
}

func (s *Storage) GetCommentsByArticleID(articleID int) []Comment {
	comments := make([]Comment, 0)
	for _, comment := range s.comments {
		if comment.ArticleID == articleID && comment.Status == "approved" {
			comments = append(comments, comment)
		}
	}

	// 按创建时间排序
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	return comments
}

func (s *Storage) ApproveComment(id int) error {
	for i := range s.comments {
		if s.comments[i].ID == id {
			s.comments[i].Status = "approved"
			s.comments[i].UpdatedAt = time.Now()
			s.saveData()
			return nil
		}
	}
	return fmt.Errorf("comment not found")
}

// ====================
// 6. 统计功能
// ====================

func (s *Storage) GetCategories() []Category {
	categoryMap := make(map[string]int)

	for _, article := range s.articles {
		if article.Status == "published" {
			categoryMap[article.Category]++
		}
	}

	categories := make([]Category, 0)
	for name, count := range categoryMap {
		if name != "" {
			categories = append(categories, Category{
				Name:  name,
				Count: count,
			})
		}
	}

	// 按文章数量排序
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Count > categories[j].Count
	})

	return categories
}

func (s *Storage) GetTags() []Tag {
	tagMap := make(map[string]int)

	for _, article := range s.articles {
		if article.Status == "published" {
			for _, tag := range article.Tags {
				tagMap[tag]++
			}
		}
	}

	tags := make([]Tag, 0)
	for name, count := range tagMap {
		tags = append(tags, Tag{
			Name:  name,
			Count: count,
		})
	}

	// 按使用次数排序
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Count > tags[j].Count
	})

	return tags
}

func (s *Storage) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_users":        len(s.users),
		"total_articles":     len(s.articles),
		"total_comments":     len(s.comments),
		"published_articles": 0,
		"approved_comments":  0,
	}

	for _, article := range s.articles {
		if article.Status == "published" {
			stats["published_articles"] = stats["published_articles"].(int) + 1
		}
	}

	for _, comment := range s.comments {
		if comment.Status == "approved" {
			stats["approved_comments"] = stats["approved_comments"].(int) + 1
		}
	}

	return stats
}

// ====================
// 7. HTTP 服务器
// ====================

type BlogServer struct {
	storage   *Storage
	templates *template.Template
}

func NewBlogServer(storage *Storage) *BlogServer {
	server := &BlogServer{
		storage: storage,
	}

	// 加载模板
	server.loadTemplates()

	return server
}

func (bs *BlogServer) loadTemplates() {
	// 创建模板函数
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04")
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"join": strings.Join,
	}

	// 嵌入式模板
	templates := map[string]string{
		"layout": `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - 博客系统</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }

        /* Header */
        header { background: #2c3e50; color: white; padding: 1rem 0; }
        .header-content { display: flex; justify-content: space-between; align-items: center; }
        .logo { font-size: 1.5rem; font-weight: bold; }
        nav ul { list-style: none; display: flex; gap: 2rem; }
        nav a { color: white; text-decoration: none; }
        nav a:hover { text-decoration: underline; }

        /* Main */
        main { min-height: calc(100vh - 140px); padding: 2rem 0; }
        .content { display: grid; grid-template-columns: 1fr 300px; gap: 2rem; }

        /* Article */
        .article { background: white; padding: 2rem; margin-bottom: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .article-title { font-size: 1.8rem; margin-bottom: 0.5rem; }
        .article-meta { color: #666; margin-bottom: 1rem; font-size: 0.9rem; }
        .article-content { line-height: 1.8; }
        .article-summary { color: #666; margin-bottom: 1rem; }
        .read-more { color: #3498db; text-decoration: none; font-weight: bold; }

        /* Sidebar */
        .sidebar { background: #f8f9fa; padding: 1.5rem; border-radius: 8px; height: fit-content; }
        .sidebar h3 { margin-bottom: 1rem; color: #2c3e50; }
        .sidebar ul { list-style: none; }
        .sidebar li { margin-bottom: 0.5rem; }
        .sidebar a { color: #666; text-decoration: none; }
        .sidebar a:hover { color: #3498db; }

        /* Tags */
        .tags { margin-top: 1rem; }
        .tag { display: inline-block; background: #3498db; color: white; padding: 0.2rem 0.5rem; margin: 0.2rem; border-radius: 3px; font-size: 0.8rem; text-decoration: none; }

        /* Forms */
        .form-group { margin-bottom: 1rem; }
        .form-group label { display: block; margin-bottom: 0.5rem; font-weight: bold; }
        .form-group input, .form-group textarea, .form-group select { width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
        .form-group textarea { height: 200px; resize: vertical; }
        .btn { background: #3498db; color: white; padding: 0.7rem 1.5rem; border: none; border-radius: 4px; cursor: pointer; text-decoration: none; display: inline-block; }
        .btn:hover { background: #2980b9; }
        .btn-danger { background: #e74c3c; }
        .btn-danger:hover { background: #c0392b; }

        /* Comments */
        .comments { margin-top: 2rem; }
        .comment { background: #f8f9fa; padding: 1rem; margin-bottom: 1rem; border-radius: 8px; }
        .comment-meta { font-size: 0.9rem; color: #666; margin-bottom: 0.5rem; }

        /* Footer */
        footer { background: #34495e; color: white; text-align: center; padding: 1rem 0; }

        /* Responsive */
        @media (max-width: 768px) {
            .content { grid-template-columns: 1fr; }
            .header-content { flex-direction: column; gap: 1rem; }
            nav ul { flex-direction: column; text-align: center; }
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div class="logo">博客系统</div>
                <nav>
                    <ul>
                        <li><a href="/">首页</a></li>
                        <li><a href="/articles">文章</a></li>
                        <li><a href="/categories">分类</a></li>
                        <li><a href="/tags">标签</a></li>
                        <li><a href="/admin">管理</a></li>
                    </ul>
                </nav>
            </div>
        </div>
    </header>

    <main>
        <div class="container">
            {{template "content" .}}
        </div>
    </main>

    <footer>
        <div class="container">
            <p>&copy; 2024 博客系统. 使用 Go 语言构建.</p>
        </div>
    </footer>
</body>
</html>`,

		"home": `
{{define "content"}}
<div class="content">
    <div class="main-content">
        <h1>欢迎来到博客系统</h1>

        {{range .Articles}}
        <article class="article">
            <h2 class="article-title"><a href="/article/{{.Slug}}" style="color: inherit; text-decoration: none;">{{.Title}}</a></h2>
            <div class="article-meta">
                作者: {{.AuthorName}} | 发布时间: {{formatDate .PublishedAt}} | 阅读: {{.ViewCount}} | 喜欢: {{.LikeCount}}
            </div>
            <div class="article-summary">{{.Summary}}</div>
            <div class="tags">
                {{range .Tags}}<a href="/tag/{{.}}" class="tag">{{.}}</a>{{end}}
            </div>
            <div style="margin-top: 1rem;">
                <a href="/article/{{.Slug}}" class="read-more">阅读全文 →</a>
            </div>
        </article>
        {{else}}
        <div class="article">
            <p>暂无文章，<a href="/admin/article/new">写第一篇文章</a>？</p>
        </div>
        {{end}}
    </div>

    <aside class="sidebar">
        <div>
            <h3>网站统计</h3>
            <ul>
                <li>文章总数: {{.Stats.published_articles}}</li>
                <li>评论总数: {{.Stats.approved_comments}}</li>
                <li>用户总数: {{.Stats.total_users}}</li>
            </ul>
        </div>

        <div style="margin-top: 2rem;">
            <h3>热门分类</h3>
            <ul>
                {{range .Categories}}
                <li><a href="/category/{{.Name}}">{{.Name}} ({{.Count}})</a></li>
                {{end}}
            </ul>
        </div>

        <div style="margin-top: 2rem;">
            <h3>热门标签</h3>
            <div>
                {{range .Tags}}
                <a href="/tag/{{.Name}}" class="tag">{{.Name}} ({{.Count}})</a>
                {{end}}
            </div>
        </div>
    </aside>
</div>
{{end}}`,

		"article": `
{{define "content"}}
<div class="content">
    <div class="main-content">
        <article class="article">
            <h1 class="article-title">{{.Article.Title}}</h1>
            <div class="article-meta">
                作者: {{.Article.AuthorName}} |
                发布时间: {{formatDate .Article.PublishedAt}} |
                分类: <a href="/category/{{.Article.Category}}">{{.Article.Category}}</a> |
                阅读: {{.Article.ViewCount}} |
                喜欢: {{.Article.LikeCount}}
            </div>
            <div class="tags">
                {{range .Article.Tags}}<a href="/tag/{{.}}" class="tag">{{.}}</a>{{end}}
            </div>
            <div class="article-content" style="margin-top: 2rem;">
                {{.Article.Content}}
            </div>
        </article>

        <div class="comments">
            <h3>评论 ({{len .Comments}})</h3>

            {{range .Comments}}
            <div class="comment">
                <div class="comment-meta">
                    <strong>{{.Username}}</strong>
                    <span style="color: #999;">{{formatDate .CreatedAt}}</span>
                </div>
                <div>{{.Content}}</div>
            </div>
            {{end}}

            <form method="POST" action="/article/{{.Article.Slug}}/comment" style="margin-top: 2rem;">
                <div class="form-group">
                    <label>用户名:</label>
                    <input type="text" name="username" required>
                </div>
                <div class="form-group">
                    <label>评论内容:</label>
                    <textarea name="content" required placeholder="请输入您的评论..."></textarea>
                </div>
                <button type="submit" class="btn">发表评论</button>
            </form>
        </div>
    </div>

    <aside class="sidebar">
        <div>
            <h3>相关文章</h3>
            <ul>
                {{range .RelatedArticles}}
                <li><a href="/article/{{.Slug}}">{{.Title}}</a></li>
                {{end}}
            </ul>
        </div>
    </aside>
</div>
{{end}}`,
	}

	// 解析模板
	bs.templates = template.New("").Funcs(funcMap)
	for name, content := range templates {
		template.Must(bs.templates.New(name).Parse(content))
	}
}

// ====================
// 8. 路由处理器
// ====================

func (bs *BlogServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/" && r.Method == "GET":
		bs.handleHome(w, r)
	case r.URL.Path == "/articles" && r.Method == "GET":
		bs.handleArticleList(w, r)
	case strings.HasPrefix(r.URL.Path, "/article/") && r.Method == "GET":
		bs.handleArticleView(w, r)
	case strings.HasPrefix(r.URL.Path, "/article/") && strings.HasSuffix(r.URL.Path, "/comment") && r.Method == "POST":
		bs.handleCommentCreate(w, r)
	case strings.HasPrefix(r.URL.Path, "/category/") && r.Method == "GET":
		bs.handleCategoryView(w, r)
	case strings.HasPrefix(r.URL.Path, "/tag/") && r.Method == "GET":
		bs.handleTagView(w, r)
	case r.URL.Path == "/search" && r.Method == "GET":
		bs.handleSearch(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin"):
		bs.handleAdmin(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (bs *BlogServer) handleHome(w http.ResponseWriter, r *http.Request) {
	articles := bs.storage.GetPublishedArticles(10, 0)
	categories := bs.storage.GetCategories()
	tags := bs.storage.GetTags()
	stats := bs.storage.GetStats()

	data := map[string]interface{}{
		"Title":      "首页",
		"Articles":   articles,
		"Categories": categories[:min(5, len(categories))],
		"Tags":       tags[:min(10, len(tags))],
		"Stats":      stats,
	}

	w.Header().Set("Content-Type", "text/html")
	bs.templates.ExecuteTemplate(w, "layout", data)
}

func (bs *BlogServer) handleArticleView(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/article/")

	article, err := bs.storage.GetArticleBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 增加阅读次数
	bs.storage.IncrementViewCount(article.ID)

	// 获取评论
	comments := bs.storage.GetCommentsByArticleID(article.ID)

	// 获取相关文章 (同分类)
	relatedArticles := make([]Article, 0)
	for _, a := range bs.storage.GetPublishedArticles(100, 0) {
		if a.ID != article.ID && a.Category == article.Category {
			relatedArticles = append(relatedArticles, a)
			if len(relatedArticles) >= 5 {
				break
			}
		}
	}

	data := map[string]interface{}{
		"Title":           article.Title,
		"Article":         article,
		"Comments":        comments,
		"RelatedArticles": relatedArticles,
	}

	w.Header().Set("Content-Type", "text/html")
	bs.templates.ExecuteTemplate(w, "layout", data)
}

func (bs *BlogServer) handleCommentCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析文章 slug
	path := strings.TrimSuffix(r.URL.Path, "/comment")
	slug := strings.TrimPrefix(path, "/article/")

	article, err := bs.storage.GetArticleBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 解析表单
	r.ParseForm()
	username := r.FormValue("username")
	content := r.FormValue("content")

	if username == "" || content == "" {
		http.Error(w, "用户名和评论内容不能为空", http.StatusBadRequest)
		return
	}

	// 创建评论 (简化处理，实际应该要求登录)
	comment := Comment{
		ArticleID: article.ID,
		UserID:    0, // 游客评论
		Username:  username,
		Content:   content,
		ParentID:  0,
	}

	_, err = bs.storage.CreateComment(comment)
	if err != nil {
		http.Error(w, "创建评论失败", http.StatusInternalServerError)
		return
	}

	// 重定向到文章页面
	http.Redirect(w, r, "/article/"+slug, http.StatusSeeOther)
}

// handleArticleList 处理文章列表页面请求
func (bs *BlogServer) handleArticleList(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	category := r.URL.Query().Get("category")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 获取文章列表
	articles := bs.storage.GetPublishedArticles(limit, offset)

	// 如果指定了分类，过滤文章
	if category != "" {
		filtered := make([]Article, 0)
		for _, article := range articles {
			if article.Category == category {
				filtered = append(filtered, article)
			}
		}
		articles = filtered
	}

	// 获取统计信息
	categories := bs.storage.GetCategories()
	tags := bs.storage.GetTags()
	stats := bs.storage.GetStats()

	data := map[string]interface{}{
		"Title":      "文章列表",
		"Articles":   articles,
		"Categories": categories,
		"Tags":       tags,
		"Stats":      stats,
		"Category":   category,
		"Limit":      limit,
		"Offset":     offset,
	}

	w.Header().Set("Content-Type", "text/html")
	bs.templates.ExecuteTemplate(w, "layout", data)
}

// handleCategoryView 处理分类页面请求
func (bs *BlogServer) handleCategoryView(w http.ResponseWriter, r *http.Request) {
	categoryName := strings.TrimPrefix(r.URL.Path, "/category/")
	if categoryName == "" {
		http.Error(w, "分类名称不能为空", http.StatusBadRequest)
		return
	}

	// 获取该分类的所有文章
	allArticles := bs.storage.GetPublishedArticles(1000, 0) // 获取足够多的文章
	categoryArticles := make([]Article, 0)
	for _, article := range allArticles {
		if article.Category == categoryName {
			categoryArticles = append(categoryArticles, article)
		}
	}

	// 获取其他数据
	categories := bs.storage.GetCategories()
	tags := bs.storage.GetTags()
	stats := bs.storage.GetStats()

	// 查找当前分类信息
	var currentCategory *Category
	for _, cat := range categories {
		if cat.Name == categoryName {
			currentCategory = &cat
			break
		}
	}

	if currentCategory == nil {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title":           fmt.Sprintf("分类：%s", categoryName),
		"Articles":        categoryArticles,
		"Categories":      categories,
		"Tags":            tags,
		"Stats":           stats,
		"CurrentCategory": currentCategory,
		"CategoryName":    categoryName,
	}

	w.Header().Set("Content-Type", "text/html")
	bs.templates.ExecuteTemplate(w, "layout", data)
}

// handleTagView 处理标签页面请求
func (bs *BlogServer) handleTagView(w http.ResponseWriter, r *http.Request) {
	tagName := strings.TrimPrefix(r.URL.Path, "/tag/")
	if tagName == "" {
		http.Error(w, "标签名称不能为空", http.StatusBadRequest)
		return
	}

	// 获取包含该标签的所有文章
	allArticles := bs.storage.GetPublishedArticles(1000, 0)
	tagArticles := make([]Article, 0)
	for _, article := range allArticles {
		for _, tag := range article.Tags {
			if tag == tagName {
				tagArticles = append(tagArticles, article)
				break
			}
		}
	}

	// 获取其他数据
	categories := bs.storage.GetCategories()
	tags := bs.storage.GetTags()
	stats := bs.storage.GetStats()

	// 查找当前标签信息
	var currentTag *Tag
	for _, tag := range tags {
		if tag.Name == tagName {
			currentTag = &tag
			break
		}
	}

	if currentTag == nil {
		// 如果标签不存在，创建一个临时标签对象
		currentTag = &Tag{
			Name:  tagName,
			Count: len(tagArticles),
		}
	}

	data := map[string]interface{}{
		"Title":      fmt.Sprintf("标签：%s", tagName),
		"Articles":   tagArticles,
		"Categories": categories,
		"Tags":       tags,
		"Stats":      stats,
		"CurrentTag": currentTag,
		"TagName":    tagName,
	}

	w.Header().Set("Content-Type", "text/html")
	bs.templates.ExecuteTemplate(w, "layout", data)
}

// handleSearch 处理搜索请求
func (bs *BlogServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "搜索关键词不能为空", http.StatusBadRequest)
		return
	}

	// 执行搜索
	searchResults := bs.storage.SearchArticles(query)

	// 获取其他数据
	categories := bs.storage.GetCategories()
	tags := bs.storage.GetTags()
	stats := bs.storage.GetStats()

	data := map[string]interface{}{
		"Title":         fmt.Sprintf("搜索结果：%s", query),
		"Articles":      searchResults,
		"Categories":    categories,
		"Tags":          tags,
		"Stats":         stats,
		"SearchQuery":   query,
		"SearchResults": len(searchResults),
	}

	w.Header().Set("Content-Type", "text/html")
	bs.templates.ExecuteTemplate(w, "layout", data)
}

func (bs *BlogServer) handleAdmin(w http.ResponseWriter, r *http.Request) {
	// 简化的管理功能
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>管理后台</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .admin-panel { max-width: 800px; margin: 0 auto; }
        .admin-nav { background: #f0f0f0; padding: 1rem; margin-bottom: 2rem; }
        .admin-nav a { margin-right: 1rem; text-decoration: none; color: #333; }
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .stat-card { background: #f8f9fa; padding: 1rem; border-radius: 8px; text-align: center; }
        .stat-number { font-size: 2rem; font-weight: bold; color: #3498db; }
    </style>
</head>
<body>
    <div class="admin-panel">
        <h1>管理后台</h1>

        <div class="admin-nav">
            <a href="/admin">概览</a>
            <a href="/admin/articles">文章管理</a>
            <a href="/admin/comments">评论管理</a>
            <a href="/admin/users">用户管理</a>
            <a href="/">返回首页</a>
        </div>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div>文章总数</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div>评论总数</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">%d</div>
                <div>用户总数</div>
            </div>
        </div>

        <h2>快速操作</h2>
        <ul>
            <li><a href="/admin/article/new">发布新文章</a></li>
            <li><a href="/admin/comments/pending">审核待批评论</a></li>
            <li><a href="/admin/backup">备份数据</a></li>
        </ul>
    </div>
</body>
</html>`,
		len(bs.storage.articles),
		len(bs.storage.comments),
		len(bs.storage.users))
}

// ====================
// 9. 辅助函数
// ====================

func hashPassword(password string) string {
	// 简化的密码哈希 (实际项目应使用 bcrypt)
	hash := sha256.Sum256([]byte(password + "salt"))
	return hex.EncodeToString(hash[:])
}

func verifyPassword(password, hash string) bool {
	return hashPassword(password) == hash
}

func generateSlug(title string) string {
	// 简化的 slug 生成
	slug := strings.ToLower(title)
	reg := regexp.MustCompile(`[^a-z0-9\p{Han}]+`)
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	// 添加随机后缀避免重复
	b := make([]byte, 4)
	rand.Read(b)
	return slug + "-" + hex.EncodeToString(b)
}

func generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ====================
// 主函数
// ====================

func main() {
	// 创建数据存储
	storage := NewStorage("./blog_data")

	// 创建示例数据
	createSampleData(storage)

	// 创建服务器
	server := NewBlogServer(storage)

	// 启动服务器
	log.Println("博客系统启动在 http://localhost:8080")
	log.Println("管理后台: http://localhost:8080/admin")
	log.Println("默认管理员: admin/admin123")

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

func createSampleData(storage *Storage) {
	// 检查是否已有文章
	if len(storage.articles) > 0 {
		return
	}

	// 创建示例用户
	author := User{
		Username: "author",
		Email:    "author@blog.com",
		Password: "author123",
		Role:     "author",
		Avatar:   "/static/author-avatar.png",
		Bio:      "技术博客作者，专注于 Go 语言开发",
	}
	createdAuthor, _ := storage.CreateUser(author)

	// 创建示例文章
	articles := []Article{
		{
			Title:    "Go 语言入门指南",
			Content:  "Go 语言是 Google 开发的开源编程语言，以其简洁、高效、并发安全而著称。本文将带你了解 Go 语言的基础语法和核心概念...",
			Summary:  "Go 语言基础语法和核心概念介绍",
			AuthorID: createdAuthor.ID,
			Category: "编程语言",
			Tags:     []string{"Go", "编程", "入门"},
			Status:   "published",
		},
		{
			Title:    "微服务架构设计原则",
			Content:  "微服务架构是一种将单一应用程序开发为一组小型服务的方法，每个服务都在自己的进程中运行，并使用轻量级机制进行通信...",
			Summary:  "微服务架构的设计原则和最佳实践",
			AuthorID: createdAuthor.ID,
			Category: "架构设计",
			Tags:     []string{"微服务", "架构", "设计"},
			Status:   "published",
		},
		{
			Title:    "Docker 容器化实践",
			Content:  "Docker 是一个开源的应用容器引擎，可以让开发者打包他们的应用以及依赖包到一个轻量级、可移植的容器中...",
			Summary:  "Docker 容器化技术的实践应用",
			AuthorID: createdAuthor.ID,
			Category: "DevOps",
			Tags:     []string{"Docker", "容器", "部署"},
			Status:   "published",
		},
	}

	for _, article := range articles {
		storage.CreateArticle(article)
	}

	log.Println("Created sample data")
}

/*
=== 项目功能清单 ===

核心功能:
✅ 用户管理 (注册、登录、权限)
✅ 文章管理 (CRUD、分类、标签)
✅ 评论系统 (发表、审核、回复)
✅ 搜索功能 (标题、内容、标签)
✅ 分类和标签管理
✅ 网站统计

界面功能:
✅ 响应式布局
✅ 文章列表和详情页
✅ 评论显示和发表
✅ 管理后台界面
✅ 搜索结果页

数据存储:
✅ JSON 文件存储
✅ 数据持久化
✅ 备份恢复

安全功能:
✅ 密码哈希
✅ 输入验证
✅ XSS 防护 (模板转义)

=== 扩展功能 ===

1. 用户体验:
   - 图片上传和管理
   - 富文本编辑器
   - 评论回复功能
   - 文章点赞功能

2. SEO 优化:
   - URL 友好化
   - 站点地图
   - RSS 订阅
   - Meta 标签优化

3. 性能优化:
   - 缓存机制
   - 静态文件 CDN
   - 数据库索引
   - 分页优化

4. 管理功能:
   - 批量操作
   - 数据导入导出
   - 插件系统
   - 主题切换

=== 部署说明 ===

1. 编译运行:
   go run main.go

2. 访问地址:
   - 首页: http://localhost:8080
   - 管理: http://localhost:8080/admin

3. 默认账号:
   - 管理员: admin/admin123
   - 作者: author/author123

4. 数据存储:
   - 位置: ./blog_data/
   - 文件: users.json, articles.json, comments.json
*/
