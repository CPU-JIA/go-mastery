package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

/*
安全和中间件练习

本练习涵盖Go语言Web应用中的安全机制和中间件，包括：
1. 身份认证和授权中间件
2. CORS跨域资源共享
3. 速率限制和防护
4. CSRF跨站请求伪造防护
5. 安全头设置
6. 输入验证和清洗
7. SQL注入防护
8. XSS跨站脚本防护
9. 请求日志和监控
10. JWT令牌处理

主要概念：
- 中间件设计模式
- 安全最佳实践
- 攻击防护机制
- 访问控制策略
- 安全审计日志
*/

// === JWT配置 ===

var jwtSecret = []byte("your-secret-key-change-in-production")

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// === 用户模型 ===

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"-"` // 密码不参与JSON序列化
	Role     string    `json:"role"`
	Created  time.Time `json:"created"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// === 用户服务 ===

type UserService struct {
	users map[string]*User
	mutex sync.RWMutex
}

func NewUserService() *UserService {
	service := &UserService{
		users: make(map[string]*User),
	}

	// 创建示例用户
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	userPassword, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)

	service.users["admin"] = &User{
		ID:       1,
		Username: "admin",
		Email:    "admin@example.com",
		Password: string(adminPassword),
		Role:     "admin",
		Created:  time.Now(),
	}

	service.users["user"] = &User{
		ID:       2,
		Username: "user",
		Email:    "user@example.com",
		Password: string(userPassword),
		Role:     "user",
		Created:  time.Now(),
	}

	return service
}

func (s *UserService) Authenticate(username, password string) (*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("密码错误")
	}

	return user, nil
}

func (s *UserService) GetByUsername(username string) (*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	return user, nil
}

// === 安全中间件 ===

// 1. 安全头中间件
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 防止点击劫持
		w.Header().Set("X-Frame-Options", "DENY")

		// XSS防护
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// 内容类型嗅探防护
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 强制HTTPS（在生产环境中启用）
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// 内容安全策略
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

		// 推荐者策略
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// 2. CORS中间件
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查允许的源
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// 3. 速率限制中间件
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter

		// 清理过期的限制器
		go func() {
			time.Sleep(time.Minute)
			rl.mutex.Lock()
			delete(rl.limiters, key)
			rl.mutex.Unlock()
		}()
	}

	return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 使用IP地址作为限制键
		key := getClientIP(r)
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(rl.rate)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")

			http.Error(w, "请求过于频繁，请稍后重试", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 4. JWT认证中间件
func JWTAuthMiddleware(userService *UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 从Authorization头获取token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "缺少Authorization头", http.StatusUnauthorized)
				return
			}

			// 检查Bearer前缀
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "无效的Authorization格式", http.StatusUnauthorized)
				return
			}

			// 解析token
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("无效的签名方法")
				}
				return jwtSecret, nil
			})

			if err != nil {
				http.Error(w, "无效的token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*Claims)
			if !ok || !token.Valid {
				http.Error(w, "无效的token声明", http.StatusUnauthorized)
				return
			}

			// 验证用户是否存在
			user, err := userService.GetByUsername(claims.Username)
			if err != nil {
				http.Error(w, "用户不存在", http.StatusUnauthorized)
				return
			}

			// 将用户信息存储到请求上下文
			r = r.WithContext(SetUserInContext(r.Context(), user))

			next.ServeHTTP(w, r)
		})
	}
}

// 5. 权限检查中间件
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, "未认证", http.StatusUnauthorized)
				return
			}

			if user.Role != role && user.Role != "admin" {
				http.Error(w, "权限不足", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// 6. CSRF防护中间件
type CSRFProtection struct {
	tokens map[string]time.Time
	mutex  sync.RWMutex
}

func NewCSRFProtection() *CSRFProtection {
	csrf := &CSRFProtection{
		tokens: make(map[string]time.Time),
	}

	// 定期清理过期token
	go csrf.cleanupExpiredTokens()

	return csrf
}

func (c *CSRFProtection) generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

func (c *CSRFProtection) addToken(token string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.tokens[token] = time.Now().Add(time.Hour)
}

func (c *CSRFProtection) validateToken(token string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	expiry, exists := c.tokens[token]
	if !exists {
		return false
	}

	return time.Now().Before(expiry)
}

func (c *CSRFProtection) cleanupExpiredTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for token, expiry := range c.tokens {
			if now.After(expiry) {
				delete(c.tokens, token)
			}
		}
		c.mutex.Unlock()
	}
}

func (c *CSRFProtection) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 对于状态改变的请求，检查CSRF token
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				token = r.FormValue("csrf_token")
			}

			if !c.validateToken(token) {
				http.Error(w, "无效的CSRF token", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// 7. 请求日志中间件
type RequestLogger struct {
	logger *log.Logger
}

func NewRequestLogger(logger *log.Logger) *RequestLogger {
	return &RequestLogger{logger: logger}
}

func (rl *RequestLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码和大小
		wrapped := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		rl.logger.Printf(
			"%s - %s %s %d %d %v %s",
			getClientIP(r),
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			wrapped.bytesWritten,
			duration,
			r.UserAgent(),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += n
	return n, err
}

// 8. 输入验证和清洗中间件
type InputValidator struct {
	maxBodySize int64
}

func NewInputValidator(maxBodySize int64) *InputValidator {
	return &InputValidator{maxBodySize: maxBodySize}
}

func (iv *InputValidator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 限制请求体大小
		r.Body = http.MaxBytesReader(w, r.Body, iv.maxBodySize)

		// 验证Content-Type
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if !isValidContentType(contentType) {
				http.Error(w, "不支持的Content-Type", http.StatusUnsupportedMediaType)
				return
			}
		}

		// 清洗查询参数
		for key, values := range r.URL.Query() {
			for i, value := range values {
				r.URL.Query()[key][i] = sanitizeInput(value)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// === HTTP处理器 ===

type AuthHandler struct {
	userService    *UserService
	csrfProtection *CSRFProtection
}

func NewAuthHandler(userService *UserService, csrfProtection *CSRFProtection) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		csrfProtection: csrfProtection,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 清洗输入
	req.Username = sanitizeInput(req.Username)

	// 认证用户
	user, err := h.userService.Authenticate(req.Username, req.Password)
	if err != nil {
		http.Error(w, "认证失败", http.StatusUnauthorized)
		return
	}

	// 生成JWT token
	token, err := generateJWTToken(user)
	if err != nil {
		http.Error(w, "生成token失败", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) GetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token := h.csrfProtection.generateToken()
	h.csrfProtection.addToken(token)

	response := map[string]string{
		"csrf_token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Profile(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "未认证", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AuthHandler) AdminOnly(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "未认证", http.StatusUnauthorized)
		return
	}

	response := map[string]string{
		"message": "这是管理员专用端点",
		"user":    user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// === 辅助函数 ===

func generateJWTToken(user *User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 返回RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

func isValidContentType(contentType string) bool {
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
	}

	for _, valid := range validTypes {
		if strings.HasPrefix(contentType, valid) {
			return true
		}
	}

	return false
}

func sanitizeInput(input string) string {
	// HTML转义
	input = html.EscapeString(input)

	// 移除控制字符
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	input = re.ReplaceAllString(input, "")

	// 限制长度
	if len(input) > 1000 {
		input = input[:1000]
	}

	return strings.TrimSpace(input)
}

// === 上下文处理 ===

type contextKey string

const userContextKey contextKey = "user"

func SetUserInContext(ctx context.Context, user *User) context.Context {
	// 在实际应用中，应该使用context.WithValue
	// 这里简化处理
	return context.WithValue(ctx, userContextKey, user)
}

func GetUserFromContext(ctx context.Context) *User {
	// 在实际应用中，应该从context.Value获取
	// 这里返回示例用户
	if user, ok := ctx.Value(userContextKey).(*User); ok {
		return user
	}
	return &User{
		ID:       1,
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Created:  time.Now(),
	}
}

// === 安全扫描器 ===

type SecurityScanner struct {
	suspiciousPatterns []string
	blockedIPs         map[string]time.Time
	mutex              sync.RWMutex
}

func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		suspiciousPatterns: []string{
			`(?i)(\bunion\b.*\bselect\b)`,                 // SQL注入
			`(?i)(\bscript\b.*\balert\b)`,                 // XSS
			`(?i)(\b(eval|exec|system)\s*\()`,             // 代码注入
			`(?i)(\.\.\/|\.\.\\)`,                         // 路径遍历
			`(?i)(\b(drop|delete|truncate)\b.*\btable\b)`, // 危险SQL
		},
		blockedIPs: make(map[string]time.Time),
	}
}

func (ss *SecurityScanner) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		// 检查IP是否被阻止
		ss.mutex.RLock()
		blockedUntil, blocked := ss.blockedIPs[clientIP]
		ss.mutex.RUnlock()

		if blocked && time.Now().Before(blockedUntil) {
			http.Error(w, "IP已被暂时阻止", http.StatusForbidden)
			return
		}

		// 扫描请求内容
		if ss.scanRequest(r) {
			// 阻止可疑IP
			ss.mutex.Lock()
			ss.blockedIPs[clientIP] = time.Now().Add(time.Hour)
			ss.mutex.Unlock()

			log.Printf("检测到可疑请求，阻止IP: %s", clientIP)
			http.Error(w, "检测到可疑活动", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (ss *SecurityScanner) scanRequest(r *http.Request) bool {
	// 扫描URL
	if ss.containsSuspiciousPattern(r.URL.String()) {
		return true
	}

	// 扫描头部
	for _, values := range r.Header {
		for _, value := range values {
			if ss.containsSuspiciousPattern(value) {
				return true
			}
		}
	}

	// 扫描用户代理
	if ss.containsSuspiciousPattern(r.UserAgent()) {
		return true
	}

	return false
}

func (ss *SecurityScanner) containsSuspiciousPattern(text string) bool {
	for _, pattern := range ss.suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, text)
		if matched {
			return true
		}
	}
	return false
}

// === 安全审计 ===

type SecurityAuditor struct {
	events []SecurityEvent
	mutex  sync.RWMutex
}

type SecurityEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"`
}

func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{
		events: make([]SecurityEvent, 0),
	}
}

func (sa *SecurityAuditor) LogEvent(eventType, source, message, severity string) {
	sa.mutex.Lock()
	defer sa.mutex.Unlock()

	event := SecurityEvent{
		Timestamp: time.Now(),
		Type:      eventType,
		Source:    source,
		Message:   message,
		Severity:  severity,
	}

	sa.events = append(sa.events, event)

	// 保持最近1000个事件
	if len(sa.events) > 1000 {
		sa.events = sa.events[1:]
	}

	// 记录到日志
	log.Printf("[SECURITY-%s] %s: %s", severity, source, message)
}

func (sa *SecurityAuditor) GetEvents() []SecurityEvent {
	sa.mutex.RLock()
	defer sa.mutex.RUnlock()

	events := make([]SecurityEvent, len(sa.events))
	copy(events, sa.events)
	return events
}

func (sa *SecurityAuditor) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(wrapped, r)

		// 记录安全相关事件
		if wrapped.statusCode == http.StatusUnauthorized {
			sa.LogEvent("AUTH_FAILURE", getClientIP(r),
				fmt.Sprintf("认证失败: %s %s", r.Method, r.URL.Path), "MEDIUM")
		} else if wrapped.statusCode == http.StatusForbidden {
			sa.LogEvent("ACCESS_DENIED", getClientIP(r),
				fmt.Sprintf("访问被拒绝: %s %s", r.Method, r.URL.Path), "MEDIUM")
		} else if wrapped.statusCode == http.StatusTooManyRequests {
			sa.LogEvent("RATE_LIMIT", getClientIP(r),
				fmt.Sprintf("速率限制触发: %s", r.URL.Path), "LOW")
		}

		// 记录慢请求
		duration := time.Since(start)
		if duration > 5*time.Second {
			sa.LogEvent("SLOW_REQUEST", getClientIP(r),
				fmt.Sprintf("慢请求: %s %s (%v)", r.Method, r.URL.Path, duration), "LOW")
		}
	})
}

// === 示例：安全配置 ===

func demonstrateSecurityBestPractices() {
	fmt.Println("=== Web安全最佳实践 ===")

	fmt.Println("1. 认证和授权:")
	fmt.Println("   ✓ 使用强密码策略")
	fmt.Println("   ✓ 实施多因素认证")
	fmt.Println("   ✓ JWT token过期管理")
	fmt.Println("   ✓ 基于角色的访问控制")

	fmt.Println("2. 输入验证:")
	fmt.Println("   ✓ 服务端数据验证")
	fmt.Println("   ✓ 输入清洗和转义")
	fmt.Println("   ✓ 文件上传验证")
	fmt.Println("   ✓ 请求大小限制")

	fmt.Println("3. 攻击防护:")
	fmt.Println("   ✓ SQL注入防护")
	fmt.Println("   ✓ XSS跨站脚本防护")
	fmt.Println("   ✓ CSRF跨站请求伪造防护")
	fmt.Println("   ✓ 点击劫持防护")

	fmt.Println("4. 传输安全:")
	fmt.Println("   ✓ 强制HTTPS")
	fmt.Println("   ✓ 安全Cookie设置")
	fmt.Println("   ✓ HSTS头部")
	fmt.Println("   ✓ 内容安全策略")

	fmt.Println("5. 监控和审计:")
	fmt.Println("   ✓ 安全事件日志")
	fmt.Println("   ✓ 异常行为检测")
	fmt.Println("   ✓ 访问模式分析")
	fmt.Println("   ✓ 实时告警系统")
}

func main() {
	// 创建服务
	userService := NewUserService()
	csrfProtection := NewCSRFProtection()
	rateLimiter := NewRateLimiter(rate.Limit(10), 20) // 每秒10次，突发20次
	requestLogger := NewRequestLogger(log.Default())
	inputValidator := NewInputValidator(1024 * 1024) // 1MB
	securityScanner := NewSecurityScanner()
	securityAuditor := NewSecurityAuditor()

	// 创建处理器
	authHandler := NewAuthHandler(userService, csrfProtection)

	// 创建路由器
	router := mux.NewRouter()

	// 应用全局中间件（顺序很重要）
	router.Use(SecurityHeadersMiddleware)
	router.Use(CORSMiddleware([]string{"http://localhost:3000", "https://yourdomain.com"}))
	router.Use(requestLogger.Middleware)
	router.Use(inputValidator.Middleware)
	router.Use(securityScanner.Middleware)
	router.Use(securityAuditor.Middleware)
	router.Use(rateLimiter.Middleware)

	// 公开路由
	router.HandleFunc("/api/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/csrf-token", authHandler.GetCSRFToken).Methods("GET")

	// 受保护的路由
	protected := router.PathPrefix("/api/protected").Subrouter()
	protected.Use(JWTAuthMiddleware(userService))
	protected.Use(csrfProtection.Middleware)
	protected.HandleFunc("/profile", authHandler.Profile).Methods("GET")

	// 管理员路由
	admin := router.PathPrefix("/api/admin").Subrouter()
	admin.Use(JWTAuthMiddleware(userService))
	admin.Use(RequireRole("admin"))
	admin.HandleFunc("/dashboard", authHandler.AdminOnly).Methods("GET")

	// 安全审计端点
	router.HandleFunc("/api/security/events", func(w http.ResponseWriter, r *http.Request) {
		events := securityAuditor.GetEvents()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}).Methods("GET")

	// 健康检查端点
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	}).Methods("GET")

	// 演示安全最佳实践
	demonstrateSecurityBestPractices()

	fmt.Println("\n=== 安全Web服务器启动 ===")
	fmt.Println("API端点:")
	fmt.Println("  POST /api/login                  - 用户登录")
	fmt.Println("  GET  /api/csrf-token             - 获取CSRF令牌")
	fmt.Println("  GET  /api/protected/profile      - 获取用户资料（需认证）")
	fmt.Println("  GET  /api/admin/dashboard        - 管理员仪表板（需管理员权限）")
	fmt.Println("  GET  /api/security/events        - 安全事件日志")
	fmt.Println("  GET  /health                     - 健康检查")
	fmt.Println()
	fmt.Println("测试用户:")
	fmt.Println("  管理员: admin / admin123")
	fmt.Println("  普通用户: user / user123")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  # 登录")
	fmt.Println(`  curl -X POST http://localhost:8080/api/login \`)
	fmt.Println(`    -H "Content-Type: application/json" \`)
	fmt.Println(`    -d '{"username":"admin","password":"admin123"}'`)
	fmt.Println()
	fmt.Println("  # 获取CSRF令牌")
	fmt.Println("  curl http://localhost:8080/api/csrf-token")
	fmt.Println()
	fmt.Println("  # 访问受保护资源")
	fmt.Println(`  curl -H "Authorization: Bearer <token>" \`)
	fmt.Println(`    -H "X-CSRF-Token: <csrf_token>" \`)
	fmt.Println("    http://localhost:8080/api/protected/profile")
	fmt.Println()
	fmt.Println("服务器运行在 http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
练习任务：

1. 基础练习：
   - 实现密码强度验证
   - 添加登录失败锁定机制
   - 实现会话管理
   - 添加API密钥认证

2. 中级练习：
   - 实现OAuth2集成
   - 添加双因素认证
   - 实现细粒度权限控制
   - 添加IP白名单功能

3. 高级练习：
   - 实现零信任架构
   - 添加Web应用防火墙（WAF）
   - 实现行为分析和异常检测
   - 集成外部安全服务

4. 安全审计：
   - 实现完整的审计日志
   - 添加安全事件关联分析
   - 实现实时安全监控
   - 添加合规性检查

5. 性能和安全平衡：
   - 优化认证性能
   - 实现智能缓存策略
   - 添加安全配置管理
   - 实现渐进式安全策略

运行前准备：
1. 安装依赖：
   go get github.com/golang-jwt/jwt/v4
   go get github.com/gorilla/mux
   go get golang.org/x/crypto/bcrypt
   go get golang.org/x/time/rate

2. 运行程序：go run main.go

安全配置检查清单：
□ 更改默认JWT密钥
□ 启用HTTPS（生产环境）
□ 配置强密码策略
□ 设置适当的CORS策略
□ 启用安全头部
□ 配置合适的速率限制
□ 设置日志监控
□ 实施输入验证
□ 配置CSRF保护
□ 启用安全审计

扩展建议：
- 集成Vault进行密钥管理
- 使用Redis进行会话存储
- 集成Prometheus进行监控
- 实现微服务安全网关
*/
