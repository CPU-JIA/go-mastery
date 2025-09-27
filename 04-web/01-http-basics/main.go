package main

import (
	"bufio"
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// =============================================================================
// 1. HTTP 基础概念
// =============================================================================

/*
HTTP（HyperText Transfer Protocol）是Web开发的基础协议：

核心概念：
1. 请求-响应模型：客户端发送请求，服务器返回响应
2. 无状态协议：每个请求都是独立的
3. 明文协议：数据以文本形式传输（除非使用HTTPS）
4. 基于TCP：运行在TCP协议之上

Go 标准库 net/http：
1. http.Server：HTTP服务器
2. http.Client：HTTP客户端
3. http.Handler：处理器接口
4. http.HandlerFunc：函数适配器
5. http.ServeMux：请求路由器

HTTP 方法：
- GET：获取资源
- POST：创建资源
- PUT：更新资源
- DELETE：删除资源
- HEAD：获取头部信息
- OPTIONS：获取允许的方法
- PATCH：部分更新

状态码：
- 1xx：信息性响应
- 2xx：成功响应
- 3xx：重定向
- 4xx：客户端错误
- 5xx：服务器错误
*/

// =============================================================================
// 2. 基础 HTTP 服务器
// =============================================================================

// 简单的处理器函数
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Custom-Header", "Go-HTTP-Server")

	// 写入响应
	fmt.Fprintf(w, "Hello, World!\n")
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "URL Path: %s\n", r.URL.Path)
	fmt.Fprintf(w, "Remote Address: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "User-Agent: %s\n", r.Header.Get("User-Agent"))
}

// 处理请求信息的函数
func requestInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	fmt.Fprintf(w, "=== 请求信息 ===\n")
	fmt.Fprintf(w, "方法: %s\n", r.Method)
	fmt.Fprintf(w, "URL: %s\n", r.URL.String())
	fmt.Fprintf(w, "协议: %s\n", r.Proto)
	fmt.Fprintf(w, "主机: %s\n", r.Host)
	fmt.Fprintf(w, "远程地址: %s\n", r.RemoteAddr)

	fmt.Fprintf(w, "\n=== 请求头 ===\n")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", name, value)
		}
	}

	// 如果有查询参数
	if len(r.URL.Query()) > 0 {
		fmt.Fprintf(w, "\n=== 查询参数 ===\n")
		for name, values := range r.URL.Query() {
			for _, value := range values {
				fmt.Fprintf(w, "%s: %s\n", name, value)
			}
		}
	}

	// 如果是POST请求，尝试读取请求体
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err == nil && len(body) > 0 {
			fmt.Fprintf(w, "\n=== 请求体 ===\n")
			fmt.Fprintf(w, "%s\n", string(body))
		}
	}
}

// ============================================================================
// 安全工具函数
// ============================================================================

// InputValidator 输入验证器
type InputValidator struct{}

// ValidateFormInput 验证表单输入
func (v *InputValidator) ValidateFormInput(name, email, message string) []string {
	var errors []string

	// 验证姓名
	if len(strings.TrimSpace(name)) == 0 {
		errors = append(errors, "姓名不能为空")
	}
	if len(name) > 100 {
		errors = append(errors, "姓名长度不能超过100个字符")
	}

	// 验证邮箱
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if len(strings.TrimSpace(email)) == 0 {
		errors = append(errors, "邮箱不能为空")
	} else if !emailRegex.MatchString(email) {
		errors = append(errors, "邮箱格式不正确")
	}
	if len(email) > 200 {
		errors = append(errors, "邮箱长度不能超过200个字符")
	}

	// 验证消息
	if len(message) > 1000 {
		errors = append(errors, "消息长度不能超过1000个字符")
	}

	return errors
}

// SafeHTMLEscape 安全的HTML转义
func SafeHTMLEscape(text string) string {
	return html.EscapeString(text)
}

// 表单处理函数
func formHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// 显示表单
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>表单示例</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>表单提交示例</h1>
    <form method="POST" action="/form">
        <div>
            <label for="name">姓名:</label>
            <input type="text" id="name" name="name" required>
        </div>
        <div>
            <label for="email">邮箱:</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div>
            <label for="message">消息:</label>
            <textarea id="message" name="message" rows="4" cols="50"></textarea>
        </div>
        <div>
            <input type="submit" value="提交">
        </div>
    </form>
</body>
</html>`
		fmt.Fprint(w, html)

	case "POST":
		// 处理表单提交
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "解析表单失败", http.StatusBadRequest)
			return
		}

		// 获取表单数据
		name := r.FormValue("name")
		email := r.FormValue("email")
		message := r.FormValue("message")

		// 验证输入
		validator := &InputValidator{}
		validationErrors := validator.ValidateFormInput(name, email, message)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if len(validationErrors) > 0 {
			// 显示验证错误
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "<h1>表单验证失败</h1>\n")
			fmt.Fprintf(w, "<ul>\n")
			for _, errMsg := range validationErrors {
				fmt.Fprintf(w, "<li>%s</li>\n", SafeHTMLEscape(errMsg))
			}
			fmt.Fprintf(w, "</ul>\n")
			fmt.Fprintf(w, "<a href=\"/form\">返回表单</a>\n")
			return
		}

		// 显示成功页面（安全转义所有用户输入）
		fmt.Fprintf(w, "<h1>表单提交成功</h1>\n")
		fmt.Fprintf(w, "<p><strong>姓名:</strong> %s</p>\n", SafeHTMLEscape(name))
		fmt.Fprintf(w, "<p><strong>邮箱:</strong> %s</p>\n", SafeHTMLEscape(email))
		fmt.Fprintf(w, "<p><strong>消息:</strong> %s</p>\n", SafeHTMLEscape(message))
		fmt.Fprintf(w, "<a href=\"/form\">返回表单</a>\n")

	default:
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
	}
}

func demonstrateBasicHTTPServer() {
	fmt.Println("=== 1. 基础 HTTP 服务器 ===")

	// 创建多路复用器
	mux := http.NewServeMux()

	// 注册处理器
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/info", requestInfoHandler)
	mux.HandleFunc("/form", formHandler)

	// 创建服务器
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	fmt.Println("HTTP 服务器启动在 :8080")
	fmt.Println("访问 http://localhost:8080/ 查看Hello页面")
	fmt.Println("访问 http://localhost:8080/info 查看请求信息")
	fmt.Println("访问 http://localhost:8080/form 查看表单示例")
	fmt.Println("按 Ctrl+C 停止服务器")

	// 启动服务器（这会阻塞）
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// =============================================================================
// 3. HTTP 客户端
// =============================================================================

func demonstrateHTTPClient() {
	fmt.Println("=== 2. HTTP 客户端 ===")

	// 创建自定义客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// GET 请求示例
	fmt.Println("发送 GET 请求:")
	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		fmt.Printf("GET 请求失败: %v\n", err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("状态码: %s\n", resp.Status)
		fmt.Printf("响应头数量: %d\n", len(resp.Header))

		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Printf("响应体长度: %d 字节\n", len(body))
		}
	}

	// POST 请求示例
	fmt.Println("\n发送 POST 请求:")
	data := url.Values{}
	data.Set("name", "张三")
	data.Set("email", "zhangsan@example.com")

	resp, err = client.PostForm("https://httpbin.org/post", data)
	if err != nil {
		fmt.Printf("POST 请求失败: %v\n", err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("状态码: %s\n", resp.Status)

		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Printf("响应体长度: %d 字节\n", len(body))
		}
	}

	// 自定义请求示例
	fmt.Println("\n发送自定义请求:")
	req, err := http.NewRequest("GET", "https://httpbin.org/headers", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	// 设置自定义头部
	req.Header.Set("User-Agent", "Go-HTTP-Client/1.0")
	req.Header.Set("X-Custom-Header", "自定义头部值")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("状态码: %s\n", resp.Status)
	}

	fmt.Println()
}

// =============================================================================
// 4. 上下文和超时控制
// =============================================================================

func demonstrateContextAndTimeout() {
	fmt.Println("=== 3. 上下文和超时控制 ===")

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 创建带上下文的请求
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/delay/3", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}

	client := &http.Client{}

	fmt.Println("发送带超时控制的请求...")
	start := time.Now()

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("请求失败 (耗时: %v): %v\n", duration, err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("请求成功 (耗时: %v): %s\n", duration, resp.Status)
	}

	// 演示取消请求
	fmt.Println("\n演示手动取消请求:")
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2() // 确保context被取消，防止泄漏

	req2, err := http.NewRequestWithContext(ctx2, "GET", "https://httpbin.org/delay/10", nil)
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return // cancel2在defer中会被调用
	}

	// 2秒后取消请求
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("取消请求...")
		cancel2()
	}()

	start = time.Now()
	resp, err = client.Do(req2)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("请求被取消 (耗时: %v): %v\n", duration, err)
	} else {
		defer resp.Body.Close()
		fmt.Printf("请求完成 (耗时: %v): %s\n", duration, resp.Status)
	}

	fmt.Println()
}

// =============================================================================
// 5. 中间件模式
// =============================================================================

// Middleware 中间件类型
type Middleware func(http.Handler) http.Handler

// loggingMiddleware 日志记录中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 记录请求
		fmt.Printf("[%s] %s %s", start.Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)

		// 调用下一个处理器
		next.ServeHTTP(w, r)

		// 记录响应时间
		duration := time.Since(start)
		fmt.Printf(" - %v\n", duration)
	})
}

// authMiddleware 认证中间件
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查Authorization头部
		auth := r.Header.Get("Authorization")
		if auth != "Bearer secret-token" {
			http.Error(w, "未授权", http.StatusUnauthorized)
			return
		}

		// 认证通过，继续处理
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware 跨域中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// chainMiddlewares 链式中间件
func chainMiddlewares(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// 受保护的处理器
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"message": "这是受保护的资源", "user": "authenticated"}`)
}

func demonstrateMiddleware() {
	fmt.Println("=== 4. 中间件模式 ===")

	mux := http.NewServeMux()

	// 普通处理器
	mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "这是公开资源")
	})

	// 使用中间件的处理器
	protectedEndpoint := chainMiddlewares(
		loggingMiddleware,
		corsMiddleware,
		authMiddleware,
	)(http.HandlerFunc(protectedHandler))

	mux.Handle("/protected", protectedEndpoint)

	// 只使用日志中间件的处理器
	loggedEndpoint := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "这是有日志记录的资源")
	}))

	mux.Handle("/logged", loggedEndpoint)

	server := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("中间件演示服务器启动在 :8081")
	fmt.Println("访问 /public - 公开资源")
	fmt.Println("访问 /logged - 带日志的资源")
	fmt.Println("访问 /protected - 需要认证 (Authorization: Bearer secret-token)")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("中间件服务器启动失败: %v", err)
	}
}

// =============================================================================
// 6. 文件服务器
// =============================================================================

func demonstrateFileServer() {
	fmt.Println("=== 5. 文件服务器 ===")

	// 创建临时目录和文件用于演示
	err := os.MkdirAll("./static", 0755)
	if err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	// 创建示例HTML文件
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>静态文件示例</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>欢迎访问静态文件服务器</h1>
    <p>这是一个静态HTML文件。</p>
    <a href="/static/">查看文件列表</a>
</body>
</html>`

	err = os.WriteFile("./static/index.html", []byte(htmlContent), 0644)
	if err != nil {
		fmt.Printf("创建HTML文件失败: %v\n", err)
		return
	}

	// 创建示例文本文件
	textContent := "这是一个文本文件示例。\n可以通过HTTP访问。"
	err = os.WriteFile("./static/example.txt", []byte(textContent), 0644)
	if err != nil {
		fmt.Printf("创建文本文件失败: %v\n", err)
		return
	}

	mux := http.NewServeMux()

	// 文件服务器
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// 自定义文件下载处理器
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		filename := strings.TrimPrefix(r.URL.Path, "/download/")
		if filename == "" {
			http.Error(w, "文件名不能为空", http.StatusBadRequest)
			return
		}

		filepath := "./static/" + filename

		// 检查文件是否存在
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			http.Error(w, "文件不存在", http.StatusNotFound)
			return
		}

		// 设置下载头部
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Type", "application/octet-stream")

		// 提供文件
		http.ServeFile(w, r, filepath)
	})

	server := &http.Server{
		Addr:              ":8082",
		Handler:           mux,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("文件服务器启动在 :8082")
	fmt.Println("访问 http://localhost:8082/static/ 浏览文件")
	fmt.Println("访问 http://localhost:8082/static/index.html 查看HTML文件")
	fmt.Println("访问 http://localhost:8082/download/example.txt 下载文件")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("文件服务器启动失败: %v", err)
	}
}

// =============================================================================
// 7. 优雅关闭
// =============================================================================

func demonstrateGracefulShutdown() {
	fmt.Println("=== 6. 优雅关闭 ===")

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 模拟长时间处理
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, "处理完成")
	})

	server := &http.Server{
		Addr:              ":8083",
		Handler:           mux,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 在goroutine中启动服务器
	go func() {
		fmt.Println("优雅关闭演示服务器启动在 :8083")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待一段时间
	time.Sleep(5 * time.Second)

	// 优雅关闭服务器
	fmt.Println("开始优雅关闭服务器...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("服务器关闭失败: %v\n", err)
	} else {
		fmt.Println("服务器已优雅关闭")
	}
}

// =============================================================================
// 8. HTTP 最佳实践
// =============================================================================

func demonstrateHTTPBestPractices() {
	fmt.Println("=== 7. HTTP 最佳实践 ===")

	fmt.Println("1. 服务器配置:")
	fmt.Println("   ✓ 设置合理的超时时间")
	fmt.Println("   ✓ 限制请求体大小")
	fmt.Println("   ✓ 使用HTTPS (生产环境)")
	fmt.Println("   ✓ 实现优雅关闭")

	fmt.Println("\n2. 安全考虑:")
	fmt.Println("   ✓ 输入验证和清理")
	fmt.Println("   ✓ 防止SQL注入")
	fmt.Println("   ✓ 使用CSRF保护")
	fmt.Println("   ✓ 设置安全头部")

	fmt.Println("\n3. 性能优化:")
	fmt.Println("   ✓ 使用连接池")
	fmt.Println("   ✓ 启用压缩")
	fmt.Println("   ✓ 缓存静态资源")
	fmt.Println("   ✓ 使用CDN")

	fmt.Println("\n4. 错误处理:")
	fmt.Println("   ✓ 返回适当的状态码")
	fmt.Println("   ✓ 提供有意义的错误信息")
	fmt.Println("   ✓ 记录错误日志")
	fmt.Println("   ✓ 实现错误恢复")

	fmt.Println("\n5. 监控和调试:")
	fmt.Println("   ✓ 记录访问日志")
	fmt.Println("   ✓ 监控响应时间")
	fmt.Println("   ✓ 健康检查端点")
	fmt.Println("   ✓ 指标收集")

	// 演示安全头部设置
	fmt.Println("\n安全头部示例:")
	secureHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Content-Security-Policy":   "default-src 'self'",
	}

	for header, value := range secureHeaders {
		fmt.Printf("  %s: %s\n", header, value)
	}

	fmt.Println()
}

// =============================================================================
// 主函数和菜单
// =============================================================================

func showMenu() {
	fmt.Println("\n=== Go HTTP 编程演示菜单 ===")
	fmt.Println("1. 基础 HTTP 服务器")
	fmt.Println("2. HTTP 客户端")
	fmt.Println("3. 上下文和超时控制")
	fmt.Println("4. 中间件模式")
	fmt.Println("5. 文件服务器")
	fmt.Println("6. 优雅关闭")
	fmt.Println("7. HTTP 最佳实践")
	fmt.Println("0. 退出")
	fmt.Print("请选择演示项目 (0-7): ")
}

func main() {
	fmt.Println("Go Web 开发 - HTTP 基础")
	fmt.Println("========================")

	reader := bufio.NewReader(os.Stdin)

	for {
		showMenu()

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("读取输入失败: %v\n", err)
			continue
		}
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			demonstrateBasicHTTPServer()
		case "2":
			demonstrateHTTPClient()
		case "3":
			demonstrateContextAndTimeout()
		case "4":
			demonstrateMiddleware()
		case "5":
			demonstrateFileServer()
		case "6":
			demonstrateGracefulShutdown()
		case "7":
			demonstrateHTTPBestPractices()
		case "0":
			fmt.Println("退出演示程序")
			return
		default:
			fmt.Println("无效选择，请重新输入")
		}

		fmt.Println("\n按 Enter 键继续...")
		if _, err := reader.ReadString('\n'); err != nil {
			fmt.Printf("读取输入失败: %v\n", err)
		}
	}
}

/*
练习任务：
1. 创建一个支持多种响应格式的API服务器（JSON、XML、HTML）
2. 实现一个文件上传服务器，支持进度显示
3. 编写一个反向代理服务器
4. 创建一个支持WebSocket升级的HTTP服务器
5. 实现一个HTTP缓存服务器
6. 编写一个支持负载均衡的HTTP客户端
7. 创建一个HTTP监控和指标收集系统
*/
