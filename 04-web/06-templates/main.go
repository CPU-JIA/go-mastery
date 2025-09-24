package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

/*
模板渲染和前端集成练习

本练习涵盖Go语言中的模板渲染和前端集成，包括：
1. HTML模板系统（html/template）
2. 模板继承和布局
3. 数据绑定和模板函数
4. 静态文件服务
5. 表单处理和验证
6. 前端资源管理
7. 模板安全（XSS防护）
8. 模板缓存和性能优化

主要概念：
- 模板解析和执行
- 模板继承（layout）
- 自定义模板函数
- 前后端数据交互
- 静态资源服务
*/

// === 数据模型 ===

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
	Role     string    `json:"role"`
	JoinDate time.Time `json:"join_date"`
}

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    User      `json:"author"`
	Tags      []string  `json:"tags"`
	Views     int       `json:"views"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"created_at"`
}

type PageData struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	User        *User       `json:"user,omitempty"`
	Posts       []Post      `json:"posts,omitempty"`
	Flash       FlashData   `json:"flash,omitempty"`
	CSRFToken   string      `json:"csrf_token,omitempty"`
	Meta        interface{} `json:"meta,omitempty"`
}

type FlashData struct {
	Type    string `json:"type"` // success, error, warning, info
	Message string `json:"message"`
}

// === 模板管理器 ===

type TemplateManager struct {
	templates map[string]*template.Template
	funcMap   template.FuncMap
	basePath  string
	devMode   bool
}

func NewTemplateManager(basePath string, devMode bool) *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		basePath:  basePath,
		devMode:   devMode,
		funcMap:   make(template.FuncMap),
	}

	// 注册自定义模板函数
	tm.registerTemplateFunctions()

	// 加载模板
	if err := tm.loadTemplates(); err != nil {
		log.Printf("加载模板失败: %v", err)
	}

	return tm
}

// 注册自定义模板函数
func (tm *TemplateManager) registerTemplateFunctions() {
	tm.funcMap = template.FuncMap{
		// 时间格式化
		"formatTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"timeAgo": func(t time.Time) string {
			duration := time.Since(t)
			if duration < time.Minute {
				return "刚刚"
			} else if duration < time.Hour {
				return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
			} else if duration < 24*time.Hour {
				return fmt.Sprintf("%d小时前", int(duration.Hours()))
			} else {
				return fmt.Sprintf("%d天前", int(duration.Hours()/24))
			}
		},

		// 字符串处理
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,

		// 数字格式化
		"comma": func(n int) string {
			str := strconv.Itoa(n)
			if len(str) <= 3 {
				return str
			}
			// 简单的千位分隔符
			result := ""
			for i, char := range str {
				if i > 0 && (len(str)-i)%3 == 0 {
					result += ","
				}
				result += string(char)
			}
			return result
		},

		// 条件判断
		"eq":  func(a, b interface{}) bool { return a == b },
		"ne":  func(a, b interface{}) bool { return a != b },
		"gt":  func(a, b int) bool { return a > b },
		"gte": func(a, b int) bool { return a >= b },
		"lt":  func(a, b int) bool { return a < b },
		"lte": func(a, b int) bool { return a <= b },

		// 数组/切片操作
		"len": func(v interface{}) int {
			switch val := v.(type) {
			case []string:
				return len(val)
			case []Post:
				return len(val)
			case string:
				return len(val)
			default:
				return 0
			}
		},
		"join": func(items []string, sep string) string {
			return strings.Join(items, sep)
		},

		// HTML辅助函数
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"safeAttr": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s)
		},
		"safeURL": func(s string) template.URL {
			return template.URL(s)
		},

		// CSS/JavaScript辅助
		"asset": func(path string) string {
			return "/static/" + path
		},
		"cssClass": func(condition bool, class string) string {
			if condition {
				return class
			}
			return ""
		},
	}
}

// 加载模板文件
func (tm *TemplateManager) loadTemplates() error {
	// 基础布局模板
	layoutFiles := []string{
		filepath.Join(tm.basePath, "layouts", "base.html"),
		filepath.Join(tm.basePath, "layouts", "header.html"),
		filepath.Join(tm.basePath, "layouts", "footer.html"),
	}

	// 页面模板
	pageFiles := map[string][]string{
		"home": {
			filepath.Join(tm.basePath, "pages", "home.html"),
		},
		"post-list": {
			filepath.Join(tm.basePath, "pages", "post-list.html"),
		},
		"post-detail": {
			filepath.Join(tm.basePath, "pages", "post-detail.html"),
		},
		"user-profile": {
			filepath.Join(tm.basePath, "pages", "user-profile.html"),
		},
		"create-post": {
			filepath.Join(tm.basePath, "pages", "create-post.html"),
		},
		"error": {
			filepath.Join(tm.basePath, "pages", "error.html"),
		},
	}

	// 为每个页面创建模板
	for name, files := range pageFiles {
		allFiles := append(layoutFiles, files...)
		tmpl, err := template.New("base.html").Funcs(tm.funcMap).ParseFiles(allFiles...)
		if err != nil {
			return fmt.Errorf("解析模板 %s 失败: %w", name, err)
		}
		tm.templates[name] = tmpl
	}

	return nil
}

// 渲染模板
func (tm *TemplateManager) Render(w http.ResponseWriter, name string, data interface{}) error {
	// 开发模式下每次重新加载模板
	if tm.devMode {
		if err := tm.loadTemplates(); err != nil {
			return err
		}
	}

	tmpl, exists := tm.templates[name]
	if !exists {
		return fmt.Errorf("模板 %s 不存在", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

// === 静态文件服务器 ===

type StaticFileServer struct {
	basePath string
	maxAge   time.Duration
}

func NewStaticFileServer(basePath string) *StaticFileServer {
	return &StaticFileServer{
		basePath: basePath,
		maxAge:   24 * time.Hour, // 缓存24小时
	}
}

func (sfs *StaticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 安全检查：防止目录遍历攻击
	requestedPath := r.URL.Path
	if strings.Contains(requestedPath, "..") {
		http.Error(w, "禁止访问", http.StatusForbidden)
		return
	}

	// 构建文件路径
	filePath := filepath.Join(sfs.basePath, strings.TrimPrefix(requestedPath, "/static/"))

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// 设置缓存头
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(sfs.maxAge.Seconds())))
	w.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))

	// 检查If-Modified-Since头
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil {
		if fileInfo.ModTime().Before(t.Add(1 * time.Second)) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// 设置内容类型
	contentType := getContentType(filePath)
	w.Header().Set("Content-Type", contentType)

	// 提供文件
	http.ServeFile(w, r, filePath)
}

// 获取文件内容类型
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	default:
		return "application/octet-stream"
	}
}

// === Web应用程序 ===

type WebApp struct {
	templateManager *TemplateManager
	staticServer    *StaticFileServer
	users           []User
	posts           []Post
}

func NewWebApp() *WebApp {
	app := &WebApp{
		templateManager: NewTemplateManager("templates", true),
		staticServer:    NewStaticFileServer("static"),
	}

	// 创建示例数据
	app.createSampleData()

	// 确保模板和静态文件目录存在
	app.ensureDirectories()

	return app
}

// 创建示例数据
func (app *WebApp) createSampleData() {
	app.users = []User{
		{
			ID:       1,
			Username: "alice",
			Email:    "alice@example.com",
			Avatar:   "/static/images/avatar1.png",
			Role:     "admin",
			JoinDate: time.Now().AddDate(0, -6, 0),
		},
		{
			ID:       2,
			Username: "bob",
			Email:    "bob@example.com",
			Avatar:   "/static/images/avatar2.png",
			Role:     "user",
			JoinDate: time.Now().AddDate(0, -3, 0),
		},
	}

	app.posts = []Post{
		{
			ID:        1,
			Title:     "Go语言模板系统入门",
			Content:   "Go语言的html/template包提供了强大的模板功能...",
			Author:    app.users[0],
			Tags:      []string{"Go", "Web", "模板"},
			Views:     156,
			Published: true,
			CreatedAt: time.Now().AddDate(0, 0, -2),
		},
		{
			ID:        2,
			Title:     "前端集成最佳实践",
			Content:   "在Go web应用中集成前端资源的最佳方法...",
			Author:    app.users[1],
			Tags:      []string{"前端", "集成", "最佳实践"},
			Views:     89,
			Published: true,
			CreatedAt: time.Now().AddDate(0, 0, -1),
		},
	}
}

// 确保目录存在并创建示例文件
func (app *WebApp) ensureDirectories() {
	// 创建目录
	dirs := []string{
		"templates/layouts",
		"templates/pages",
		"static/css",
		"static/js",
		"static/images",
	}

	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// 创建基础模板文件
	app.createTemplateFiles()
	app.createStaticFiles()
}

// 创建模板文件
func (app *WebApp) createTemplateFiles() {
	// 基础布局模板
	baseTemplate := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Go Web应用</title>
    <meta name="description" content="{{.Description}}">
    <link rel="stylesheet" href="{{asset "css/bootstrap.min.css"}}">
    <link rel="stylesheet" href="{{asset "css/style.css"}}">
</head>
<body>
    {{template "header" .}}

    <main class="container my-4">
        {{if .Flash.Message}}
        <div class="alert alert-{{.Flash.Type}} alert-dismissible fade show" role="alert">
            {{.Flash.Message}}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
        {{end}}

        {{template "content" .}}
    </main>

    {{template "footer" .}}

    <script src="{{asset "js/bootstrap.bundle.min.js"}}"></script>
    <script src="{{asset "js/app.js"}}"></script>
</body>
</html>`

	// 头部模板
	headerTemplate := `{{define "header"}}
<nav class="navbar navbar-expand-lg navbar-dark bg-primary">
    <div class="container">
        <a class="navbar-brand" href="/">Go Web应用</a>

        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav">
            <span class="navbar-toggler-icon"></span>
        </button>

        <div class="collapse navbar-collapse" id="navbarNav">
            <ul class="navbar-nav me-auto">
                <li class="nav-item">
                    <a class="nav-link" href="/">首页</a>
                </li>
                <li class="nav-item">
                    <a class="nav-link" href="/posts">文章</a>
                </li>
            </ul>

            <ul class="navbar-nav">
                {{if .User}}
                <li class="nav-item dropdown">
                    <a class="nav-link dropdown-toggle" href="#" role="button" data-bs-toggle="dropdown">
                        <img src="{{.User.Avatar}}" alt="头像" class="rounded-circle me-1" width="24" height="24">
                        {{.User.Username}}
                    </a>
                    <ul class="dropdown-menu">
                        <li><a class="dropdown-item" href="/profile">个人资料</a></li>
                        <li><a class="dropdown-item" href="/posts/new">写文章</a></li>
                        <li><hr class="dropdown-divider"></li>
                        <li><a class="dropdown-item" href="/logout">退出登录</a></li>
                    </ul>
                </li>
                {{else}}
                <li class="nav-item">
                    <a class="nav-link" href="/login">登录</a>
                </li>
                {{end}}
            </ul>
        </div>
    </div>
</nav>
{{end}}`

	// 页脚模板
	footerTemplate := `{{define "footer"}}
<footer class="bg-light py-4 mt-5">
    <div class="container">
        <div class="row">
            <div class="col-md-6">
                <p>&copy; {{formatDate .Meta.CurrentTime}} Go Web应用. 保留所有权利.</p>
            </div>
            <div class="col-md-6 text-end">
                <p>基于Go语言和Bootstrap构建</p>
            </div>
        </div>
    </div>
</footer>
{{end}}`

	// 首页模板
	homeTemplate := `{{define "content"}}
<div class="row">
    <div class="col-lg-8">
        <h1>欢迎来到Go Web应用</h1>
        <p class="lead">这是一个使用Go语言html/template包构建的示例应用。</p>

        <h2>最新文章</h2>
        {{range .Posts}}
        <div class="card mb-3">
            <div class="card-body">
                <h5 class="card-title">
                    <a href="/posts/{{.ID}}" class="text-decoration-none">{{.Title}}</a>
                </h5>
                <p class="card-text">{{truncate .Content 150}}</p>
                <div class="d-flex justify-content-between align-items-center">
                    <small class="text-muted">
                        by {{.Author.Username}} • {{timeAgo .CreatedAt}} • {{comma .Views}} 次浏览
                    </small>
                    <div>
                        {{range .Tags}}
                        <span class="badge bg-secondary">{{.}}</span>
                        {{end}}
                    </div>
                </div>
            </div>
        </div>
        {{end}}
    </div>

    <div class="col-lg-4">
        <div class="card">
            <div class="card-header">
                <h5>统计信息</h5>
            </div>
            <div class="card-body">
                <p>总用户数: {{len .Meta.Users}}</p>
                <p>总文章数: {{len .Posts}}</p>
                <p>今日访问: {{.Meta.TodayViews}}</p>
            </div>
        </div>
    </div>
</div>
{{end}}`

	// 写入模板文件
	templates := map[string]string{
		"templates/layouts/base.html":   baseTemplate,
		"templates/layouts/header.html": headerTemplate,
		"templates/layouts/footer.html": footerTemplate,
		"templates/pages/home.html":     homeTemplate,
	}

	for path, content := range templates {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				log.Printf("创建模板文件 %s 失败: %v", path, err)
			}
		}
	}
}

// 创建静态文件
func (app *WebApp) createStaticFiles() {
	// 基础CSS
	css := `
/* 自定义样式 */
body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.navbar-brand {
    font-weight: bold;
}

.card {
    box-shadow: 0 0.125rem 0.25rem rgba(0, 0, 0, 0.075);
    border: 1px solid rgba(0, 0, 0, 0.125);
}

.card:hover {
    box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
    transition: box-shadow 0.15s ease-in-out;
}

.badge {
    margin-right: 0.25rem;
}

footer {
    border-top: 1px solid #dee2e6;
}
`

	// 基础JavaScript
	js := `
// 应用初始化
document.addEventListener('DOMContentLoaded', function() {
    console.log('Go Web应用已加载');

    // 自动隐藏警告消息
    setTimeout(function() {
        const alerts = document.querySelectorAll('.alert');
        alerts.forEach(function(alert) {
            if (alert.classList.contains('show')) {
                const bsAlert = new bootstrap.Alert(alert);
                bsAlert.close();
            }
        });
    }, 5000);
});

// 工具函数
function formatNumber(num) {
    return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}

function timeAgo(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;

    const minute = 60 * 1000;
    const hour = minute * 60;
    const day = hour * 24;

    if (diff < minute) {
        return '刚刚';
    } else if (diff < hour) {
        return Math.floor(diff / minute) + '分钟前';
    } else if (diff < day) {
        return Math.floor(diff / hour) + '小时前';
    } else {
        return Math.floor(diff / day) + '天前';
    }
}
`

	// 写入静态文件
	staticFiles := map[string]string{
		"static/css/style.css": css,
		"static/js/app.js":     js,
	}

	for path, content := range staticFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				log.Printf("创建静态文件 %s 失败: %v", path, err)
			}
		}
	}
}

// === HTTP处理器 ===

// 首页处理器
func (app *WebApp) HandleHome(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "首页",
		Description: "Go Web应用首页",
		User:        &app.users[0], // 模拟登录用户
		Posts:       app.posts,
		Meta: map[string]interface{}{
			"Users":       app.users,
			"CurrentTime": time.Now(),
			"TodayViews":  1234,
		},
	}

	if err := app.templateManager.Render(w, "home", data); err != nil {
		http.Error(w, "渲染模板失败: "+err.Error(), http.StatusInternalServerError)
	}
}

// 文章列表处理器
func (app *WebApp) HandlePostList(w http.ResponseWriter, r *http.Request) {
	// 实现分页逻辑
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	data := PageData{
		Title:       "文章列表",
		Description: "浏览所有文章",
		User:        &app.users[0],
		Posts:       app.posts, // 实际应用中需要分页
		Meta: map[string]interface{}{
			"CurrentPage": page,
			"TotalPages":  1,
		},
	}

	if err := app.templateManager.Render(w, "post-list", data); err != nil {
		http.Error(w, "渲染模板失败: "+err.Error(), http.StatusInternalServerError)
	}
}

// API端点：获取文章JSON数据
func (app *WebApp) HandlePostsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts": app.posts,
		"total": len(app.posts),
	})
}

// 文件上传处理器
func (app *WebApp) HandleFileUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 解析multipart表单
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "获取文件失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 创建上传目录
	uploadDir := "static/uploads"
	os.MkdirAll(uploadDir, 0755)

	// 保存文件
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filepath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "创建文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "保存文件失败", http.StatusInternalServerError)
		return
	}

	// 返回文件URL
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"url":      "/static/uploads/" + filename,
		"filename": header.Filename,
		"size":     strconv.FormatInt(header.Size, 10),
	})
}

// === 表单处理示例 ===

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (cf ContactForm) Validate() map[string]string {
	errors := make(map[string]string)

	if strings.TrimSpace(cf.Name) == "" {
		errors["name"] = "姓名不能为空"
	}

	if strings.TrimSpace(cf.Email) == "" {
		errors["email"] = "邮箱不能为空"
	} else if !strings.Contains(cf.Email, "@") {
		errors["email"] = "邮箱格式不正确"
	}

	if strings.TrimSpace(cf.Subject) == "" {
		errors["subject"] = "主题不能为空"
	}

	if len(strings.TrimSpace(cf.Message)) < 10 {
		errors["message"] = "消息内容至少10个字符"
	}

	return errors
}

func (app *WebApp) HandleContact(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var form ContactForm
		if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
			http.Error(w, "解析表单数据失败", http.StatusBadRequest)
			return
		}

		// 验证表单
		if errors := form.Validate(); len(errors) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"errors":  errors,
			})
			return
		}

		// 处理表单数据（发送邮件、保存到数据库等）
		log.Printf("收到联系表单: %+v", form)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "感谢您的留言，我们会尽快回复！",
		})
	} else {
		// 显示联系表单页面
		data := PageData{
			Title:       "联系我们",
			Description: "通过表单联系我们",
		}

		if err := app.templateManager.Render(w, "contact", data); err != nil {
			http.Error(w, "渲染模板失败: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// === 安全最佳实践演示 ===

func demonstrateTemplateSecurity() {
	fmt.Println("=== 模板安全最佳实践 ===")

	fmt.Println("1. XSS防护:")
	fmt.Println("   ✓ 使用html/template自动转义")
	fmt.Println("   ✓ 谨慎使用safeHTML函数")
	fmt.Println("   ✓ 验证和清洗用户输入")
	fmt.Println("   ✓ 设置Content-Security-Policy头")

	fmt.Println("2. CSRF防护:")
	fmt.Println("   ✓ 在表单中包含CSRF令牌")
	fmt.Println("   ✓ 验证请求来源")
	fmt.Println("   ✓ 使用SameSite cookie属性")

	fmt.Println("3. 文件安全:")
	fmt.Println("   ✓ 验证文件类型和大小")
	fmt.Println("   ✓ 防止路径遍历攻击")
	fmt.Println("   ✓ 限制上传文件权限")
	fmt.Println("   ✓ 扫描恶意文件")
}

func main() {
	// 创建Web应用
	app := NewWebApp()

	// 创建路由器
	router := mux.NewRouter()

	// 静态文件服务
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", app.staticServer))

	// 页面路由
	router.HandleFunc("/", app.HandleHome).Methods("GET")
	router.HandleFunc("/posts", app.HandlePostList).Methods("GET")
	router.HandleFunc("/contact", app.HandleContact).Methods("GET", "POST")

	// API路由
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/posts", app.HandlePostsAPI).Methods("GET")
	api.HandleFunc("/upload", app.HandleFileUpload).Methods("POST")

	// 演示安全最佳实践
	demonstrateTemplateSecurity()

	fmt.Println("=== 模板渲染服务器启动 ===")
	fmt.Println("页面端点:")
	fmt.Println("  GET  /           - 首页")
	fmt.Println("  GET  /posts      - 文章列表")
	fmt.Println("  GET  /contact    - 联系表单")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  GET  /api/posts  - 获取文章JSON数据")
	fmt.Println("  POST /api/upload - 文件上传")
	fmt.Println()
	fmt.Println("静态资源:")
	fmt.Println("  /static/css/     - CSS文件")
	fmt.Println("  /static/js/      - JavaScript文件")
	fmt.Println("  /static/images/  - 图片文件")
	fmt.Println()
	fmt.Println("服务器运行在 http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
练习任务：

1. 基础练习：
   - 创建更多页面模板（用户资料、文章详情）
   - 实现模板缓存机制
   - 添加更多自定义模板函数
   - 实现多语言支持（i18n）

2. 中级练习：
   - 实现模板继承系统
   - 添加表单验证和错误显示
   - 实现文件上传和管理
   - 集成前端框架（Bootstrap、Tailwind CSS）

3. 高级练习：
   - 实现服务端渲染（SSR）
   - 添加模板性能监控
   - 实现模板热重载
   - 集成前端构建工具（Webpack、Vite）

4. 安全练习：
   - 实现CSRF令牌验证
   - 添加内容安全策略（CSP）
   - 实现文件上传安全检查
   - 添加模板输出过滤

5. 性能优化：
   - 实现模板预编译
   - 添加Gzip压缩
   - 实现资源版本控制
   - 优化静态资源缓存

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux

2. 创建目录结构：
   mkdir -p templates/layouts templates/pages static/css static/js static/images

3. 运行程序：go run main.go

4. 访问应用：http://localhost:8080

目录结构：
project/
├── main.go
├── templates/
│   ├── layouts/
│   │   ├── base.html
│   │   ├── header.html
│   │   └── footer.html
│   └── pages/
│       ├── home.html
│       ├── post-list.html
│       └── contact.html
└── static/
    ├── css/
    │   └── style.css
    ├── js/
    │   └── app.js
    └── images/
        └── (图片文件)

扩展建议：
- 集成模板引擎（如Pongo2）
- 实现组件化模板系统
- 添加实时预览功能
- 集成CDN资源管理
*/
