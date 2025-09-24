package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// =============================================================================
// 1. 中间件和认证基础概念
// =============================================================================

/*
中间件（Middleware）是Web开发中的核心模式：

中间件的作用：
1. 请求预处理：日志记录、认证、授权
2. 响应后处理：压缩、缓存头设置
3. 错误处理：统一错误格式、错误日志
4. 横切关注点：CORS、安全头部、限流

中间件模式：
1. 链式模式：middleware1 -> middleware2 -> handler
2. 洋葱模式：请求向内，响应向外
3. 条件中间件：基于条件应用不同中间件

认证（Authentication）vs 授权（Authorization）：
- 认证：验证用户身份（你是谁？）
- 授权：验证用户权限（你能做什么？）

常见认证方式：
1. Session-Cookie：传统Web应用
2. JWT Token：RESTful API
3. Basic Auth：简单场景
4. OAuth 2.0：第三方登录
5. API Key：服务间调用

安全考虑：
1. 密码哈希：使用bcrypt等强哈希算法
2. 会话管理：安全的会话ID生成和存储
3. CSRF保护：防止跨站请求伪造
4. XSS保护：防止跨站脚本攻击
5. SQL注入：参数化查询
*/

// =============================================================================
// 2. 用户模型和存储
// =============================================================================

// User 用户模型
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // 不在JSON中显示
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	LoginAt   time.Time `json:"login_at,omitempty"`
}

// Session 会话模型
type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// AuthStore 认证存储接口
type AuthStore interface {
	// 用户管理
	CreateUser(user *User) error
	GetUserByID(id int) (*User, error)
	GetUserByUsername(username string) (*User, error)
	UpdateUser(user *User) error

	// 会话管理
	CreateSession(session *Session) error
	GetSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	CleanupExpiredSessions() int
}

// InMemoryAuthStore 内存认证存储实现
type InMemoryAuthStore struct {
	users    map[int]*User
	sessions map[string]*Session
	nextID   int
	mu       sync.RWMutex
}

// NewInMemoryAuthStore 创建内存认证存储
func NewInMemoryAuthStore() *InMemoryAuthStore {
	store := &InMemoryAuthStore{
		users:    make(map[int]*User),
		sessions: make(map[string]*Session),
		nextID:   1,
	}

	// 创建管理员用户
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := &User{
		ID:        store.nextID,
		Username:  "admin",
		Email:     "admin@example.com",
		Password:  string(adminPassword),
		Role:      "admin",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	store.users[store.nextID] = admin
	store.nextID++

	// 创建普通用户
	userPassword, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user := &User{
		ID:        store.nextID,
		Username:  "user",
		Email:     "user@example.com",
		Password:  string(userPassword),
		Role:      "user",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	store.users[store.nextID] = user
	store.nextID++

	return store
}

// CreateUser 创建用户
func (s *InMemoryAuthStore) CreateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查用户名是否存在
	for _, existingUser := range s.users {
		if existingUser.Username == user.Username {
			return fmt.Errorf("用户名已存在")
		}
	}

	user.ID = s.nextID
	user.CreatedAt = time.Now()
	s.users[s.nextID] = user
	s.nextID++

	return nil
}

// GetUserByID 根据ID获取用户
func (s *InMemoryAuthStore) GetUserByID(id int) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	userCopy := *user
	return &userCopy, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *InMemoryAuthStore) GetUserByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, fmt.Errorf("用户不存在")
}

// UpdateUser 更新用户
func (s *InMemoryAuthStore) UpdateUser(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return fmt.Errorf("用户不存在")
	}

	s.users[user.ID] = user
	return nil
}

// CreateSession 创建会话
func (s *InMemoryAuthStore) CreateSession(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

// GetSession 获取会话
func (s *InMemoryAuthStore) GetSession(sessionID string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	// 检查会话是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("会话已过期")
	}

	sessionCopy := *session
	return &sessionCopy, nil
}

// DeleteSession 删除会话
func (s *InMemoryAuthStore) DeleteSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// CleanupExpiredSessions 清理过期会话
func (s *InMemoryAuthStore) CleanupExpiredSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	count := 0

	for sessionID, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, sessionID)
			count++
		}
	}

	return count
}

// =============================================================================
// 3. 认证服务
// =============================================================================

// AuthService 认证服务
type AuthService struct {
	store         AuthStore
	sessionExpiry time.Duration
}

// NewAuthService 创建认证服务
func NewAuthService(store AuthStore) *AuthService {
	service := &AuthService{
		store:         store,
		sessionExpiry: 24 * time.Hour, // 24小时有效期
	}

	// 启动清理goroutine
	go service.cleanupWorker()

	return service
}

// Register 用户注册
func (as *AuthService) Register(req *RegisterRequest) (*User, error) {
	// 验证输入
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return nil, fmt.Errorf("用户名、密码和邮箱不能为空")
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("密码长度不能少于6位")
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码处理失败")
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "user", // 默认角色
		IsActive: true,
	}

	if err := as.store.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (as *AuthService) Login(req *LoginRequest, ipAddress, userAgent string) (*Session, *User, error) {
	// 获取用户
	user, err := as.store.GetUserByUsername(req.Username)
	if err != nil {
		return nil, nil, fmt.Errorf("用户名或密码错误")
	}

	// 检查用户是否活跃
	if !user.IsActive {
		return nil, nil, fmt.Errorf("用户账号已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, nil, fmt.Errorf("用户名或密码错误")
	}

	// 创建会话
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, nil, fmt.Errorf("创建会话失败")
	}

	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(as.sessionExpiry),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := as.store.CreateSession(session); err != nil {
		return nil, nil, fmt.Errorf("创建会话失败")
	}

	// 更新登录时间
	user.LoginAt = time.Now()
	as.store.UpdateUser(user)

	return session, user, nil
}

// Logout 用户登出
func (as *AuthService) Logout(sessionID string) error {
	return as.store.DeleteSession(sessionID)
}

// GetUserBySession 根据会话获取用户
func (as *AuthService) GetUserBySession(sessionID string) (*User, error) {
	session, err := as.store.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return as.store.GetUserByID(session.UserID)
}

// ChangePassword 修改密码
func (as *AuthService) ChangePassword(userID int, req *ChangePasswordRequest) error {
	user, err := as.store.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("原密码错误")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码处理失败")
	}

	user.Password = string(hashedPassword)
	return as.store.UpdateUser(user)
}

// cleanupWorker 清理过期会话的工作goroutine
func (as *AuthService) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		count := as.store.CleanupExpiredSessions()
		if count > 0 {
			log.Printf("清理了 %d 个过期会话", count)
		}
	}
}

// generateSessionID 生成会话ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// =============================================================================
// 4. 认证中间件
// =============================================================================

// AuthMiddleware 认证中间件
func AuthMiddleware(authService *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从Cookie获取会话ID
			cookie, err := r.Cookie("session_id")
			if err != nil {
				http.Error(w, "未认证", http.StatusUnauthorized)
				return
			}

			// 验证会话
			user, err := authService.GetUserBySession(cookie.Value)
			if err != nil {
				http.Error(w, "会话无效", http.StatusUnauthorized)
				return
			}

			// 将用户信息添加到请求头中（简化版本，实际应使用context）
			r.Header.Set("X-User-ID", strconv.Itoa(user.ID))
			r.Header.Set("X-User-Role", user.Role)

			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddleware 角色中间件
func RoleMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Header.Get("X-User-Role")

			// 检查用户角色
			hasRole := false
			for _, role := range requiredRoles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "权限不足", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnlyMiddleware 仅管理员中间件
func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return RoleMiddleware("admin")(next)
}

// =============================================================================
// 5. 安全中间件
// =============================================================================

// SecurityHeaders 安全头部中间件
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 防止XSS攻击
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// HTTPS相关
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// 内容安全策略
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// 引用者策略
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(requests int, duration time.Duration) func(http.Handler) http.Handler {
	type client struct {
		requests int
		window   time.Time
	}

	clients := make(map[string]*client)
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			now := time.Now()

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				clients[ip] = &client{
					requests: 1,
					window:   now,
				}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			// 检查时间窗口
			if now.Sub(c.window) > duration {
				c.requests = 1
				c.window = now
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			// 检查请求数量
			if c.requests >= requests {
				mu.Unlock()
				w.Header().Set("Retry-After", strconv.Itoa(int(duration.Seconds())))
				http.Error(w, "请求过于频繁", http.StatusTooManyRequests)
				return
			}

			c.requests++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// CSRFMiddleware CSRF保护中间件
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 对于状态改变的请求，验证CSRF token
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH" {
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("csrf_token")
			}

			// 这里应该验证CSRF token的有效性
			// 简化版本：检查token是否存在
			if token == "" {
				http.Error(w, "CSRF token 缺失", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头部
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头部
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}

	return ip
}

// =============================================================================
// 6. 控制器
// =============================================================================

// AuthController 认证控制器
type AuthController struct {
	authService *AuthService
}

// NewAuthController 创建认证控制器
func NewAuthController(authService *AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register 注册处理器
func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}

	user, err := ac.authService.Register(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "注册成功",
		"user":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login 登录处理器
func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}

	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	session, user, err := ac.authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// 设置会话Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为true（HTTPS）
		SameSite: http.SameSiteStrictMode,
	})

	response := map[string]interface{}{
		"success": true,
		"message": "登录成功",
		"user":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout 登出处理器
func (ac *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		ac.authService.Logout(cookie.Value)
	}

	// 清除Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	response := map[string]interface{}{
		"success": true,
		"message": "登出成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Profile 个人信息处理器
func (ac *AuthController) Profile(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	user, err := ac.authService.store.GetUserByID(userID)
	if err != nil {
		http.Error(w, "用户不存在", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"user":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangePassword 修改密码处理器
func (ac *AuthController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "请求格式错误", http.StatusBadRequest)
		return
	}

	if err := ac.authService.ChangePassword(userID, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "密码修改成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdminDashboard 管理员面板
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"message": "欢迎访问管理员面板",
		"data": map[string]interface{}{
			"stats": map[string]int{
				"users":    100,
				"sessions": 25,
				"requests": 1500,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PublicResource 公共资源
func PublicResource(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"success": true,
		"message": "这是公共资源，无需认证",
		"data":    "公开数据",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ProtectedResource 受保护资源
func ProtectedResource(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	userRole := r.Header.Get("X-User-Role")

	response := map[string]interface{}{
		"success": true,
		"message": "这是受保护的资源",
		"user": map[string]string{
			"id":   userID,
			"role": userRole,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// 7. 路由设置
// =============================================================================

func setupRoutes(authController *AuthController, authService *AuthService) http.Handler {
	mux := http.NewServeMux()

	// 公共路由
	mux.HandleFunc("/api/v1/public", PublicResource)
	mux.HandleFunc("/api/v1/auth/register", authController.Register)
	mux.HandleFunc("/api/v1/auth/login", authController.Login)
	mux.HandleFunc("/api/v1/auth/logout", authController.Logout)

	// 需要认证的路由
	authMux := http.NewServeMux()
	authMux.HandleFunc("/api/v1/auth/profile", authController.Profile)
	authMux.HandleFunc("/api/v1/auth/change-password", authController.ChangePassword)
	authMux.HandleFunc("/api/v1/protected", ProtectedResource)

	// 仅管理员路由
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("/api/v1/admin/dashboard", AdminDashboard)

	// 应用中间件
	mux.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth",
		AuthMiddleware(authService)(authMux)))

	mux.Handle("/api/v1/protected",
		AuthMiddleware(authService)(http.HandlerFunc(ProtectedResource)))

	mux.Handle("/api/v1/admin/", http.StripPrefix("/api/v1/admin",
		AuthMiddleware(authService)(AdminOnlyMiddleware(adminMux))))

	// 应用全局中间件
	handler := SecurityHeaders(
		RateLimitMiddleware(100, time.Minute)( // 每分钟100个请求
			loggingMiddleware(mux),
		),
	)

	return handler
}

// loggingMiddleware 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("%s %s %d %v %s",
			r.Method, r.URL.Path, wrapped.statusCode,
			time.Since(start), getClientIP(r))
	})
}

// responseWriter 包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// =============================================================================
// 8. 主应用程序
// =============================================================================

func main() {
	fmt.Println("Go Web 开发 - 中间件和认证")
	fmt.Println("============================")

	// 初始化存储和服务
	authStore := NewInMemoryAuthStore()
	authService := NewAuthService(authStore)
	authController := NewAuthController(authService)

	// 设置路由
	handler := setupRoutes(authController, authService)

	// 创建服务器
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	fmt.Println("认证服务器启动在 :8080")
	fmt.Println("\n预置账号:")
	fmt.Println("管理员 - 用户名: admin, 密码: admin123")
	fmt.Println("普通用户 - 用户名: user, 密码: user123")

	fmt.Println("\nAPI端点:")
	fmt.Println("POST /api/v1/auth/register     - 用户注册")
	fmt.Println("POST /api/v1/auth/login        - 用户登录")
	fmt.Println("POST /api/v1/auth/logout       - 用户登出")
	fmt.Println("GET  /api/v1/auth/profile      - 个人信息 (需要认证)")
	fmt.Println("POST /api/v1/auth/change-password - 修改密码 (需要认证)")
	fmt.Println("GET  /api/v1/protected         - 受保护资源 (需要认证)")
	fmt.Println("GET  /api/v1/admin/dashboard   - 管理员面板 (需要管理员权限)")
	fmt.Println("GET  /api/v1/public            - 公共资源")

	fmt.Println("\n测试命令:")
	fmt.Println("# 用户登录")
	fmt.Println(`curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"username":"user","password":"user123"}' -c cookies.txt`)
	fmt.Println("# 访问受保护资源")
	fmt.Println(`curl http://localhost:8080/api/v1/protected -b cookies.txt`)
	fmt.Println("# 管理员登录")
	fmt.Println(`curl -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}' -c admin-cookies.txt`)
	fmt.Println("# 访问管理员面板")
	fmt.Println(`curl http://localhost:8080/api/v1/admin/dashboard -b admin-cookies.txt`)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

/*
练习任务：
1. 实现JWT Token认证
2. 添加密码重置功能（邮件验证）
3. 实现OAuth 2.0第三方登录
4. 添加双因素认证（2FA）
5. 实现API密钥认证
6. 添加IP白名单中间件
7. 实现请求签名验证
8. 添加审计日志中间件
9. 实现分布式会话存储（Redis）
10. 添加暴力破解防护
*/
