package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// 1. 路由和REST API基础概念
// =============================================================================

/*
REST (Representational State Transfer) 是一种Web API设计风格：

核心原则：
1. 无状态：每个请求包含所有必要信息
2. 统一接口：使用标准HTTP方法
3. 分层系统：客户端无需了解服务器内部结构
4. 缓存：支持缓存以提高性能
5. 按需代码：服务器可以发送可执行代码（可选）

HTTP方法映射：
- GET /users       - 获取用户列表
- GET /users/123   - 获取特定用户
- POST /users      - 创建新用户
- PUT /users/123   - 更新特定用户
- DELETE /users/123 - 删除特定用户
- PATCH /users/123 - 部分更新用户

状态码规范：
- 200 OK：请求成功
- 201 Created：资源创建成功
- 204 No Content：删除成功
- 400 Bad Request：请求格式错误
- 401 Unauthorized：未认证
- 403 Forbidden：无权限
- 404 Not Found：资源不存在
- 409 Conflict：资源冲突
- 500 Internal Server Error：服务器错误

Go语言路由：
1. 标准库 http.ServeMux：基础路由功能
2. 第三方路由器：gorilla/mux, chi, gin等
3. 路径参数：/users/{id}
4. 查询参数：/users?page=1&limit=10
*/

// =============================================================================
// 2. 数据模型
// =============================================================================

// User 用户模型
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Name     *string `json:"name,omitempty"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta"`
}

// =============================================================================
// 3. 数据存储层（内存模拟）
// =============================================================================

// UserStore 用户存储接口
type UserStore interface {
	Create(user *User) error
	GetByID(id int) (*User, error)
	GetAll(page, limit int) ([]*User, int, error)
	Update(id int, updates *UpdateUserRequest) (*User, error)
	Delete(id int) error
	GetByUsername(username string) (*User, error)
}

// InMemoryUserStore 内存用户存储实现
type InMemoryUserStore struct {
	users  map[int]*User
	nextID int
	mu     sync.RWMutex
}

// NewInMemoryUserStore 创建内存用户存储
func NewInMemoryUserStore() *InMemoryUserStore {
	store := &InMemoryUserStore{
		users:  make(map[int]*User),
		nextID: 1,
	}

	// 添加一些示例数据
	store.seedData()
	return store
}

// seedData 添加示例数据
func (s *InMemoryUserStore) seedData() {
	users := []*CreateUserRequest{
		{Username: "alice", Email: "alice@example.com", Name: "Alice Smith"},
		{Username: "bob", Email: "bob@example.com", Name: "Bob Johnson"},
		{Username: "charlie", Email: "charlie@example.com", Name: "Charlie Brown"},
	}

	for _, userReq := range users {
		user := &User{
			ID:        s.nextID,
			Username:  userReq.Username,
			Email:     userReq.Email,
			Name:      userReq.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		s.users[s.nextID] = user
		s.nextID++
	}
}

// Create 创建用户
func (s *InMemoryUserStore) Create(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查用户名是否已存在
	for _, existingUser := range s.users {
		if existingUser.Username == user.Username {
			return fmt.Errorf("用户名已存在")
		}
	}

	user.ID = s.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	s.users[s.nextID] = user
	s.nextID++

	return nil
}

// GetByID 根据ID获取用户
func (s *InMemoryUserStore) GetByID(id int) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	// 返回副本以避免外部修改
	userCopy := *user
	return &userCopy, nil
}

// GetAll 获取所有用户（分页）
func (s *InMemoryUserStore) GetAll(page, limit int) ([]*User, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 转换为切片
	allUsers := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		userCopy := *user
		allUsers = append(allUsers, &userCopy)
	}

	total := len(allUsers)

	// 计算分页
	start := (page - 1) * limit
	if start >= total {
		return []*User{}, total, nil
	}

	end := start + limit
	if end > total {
		end = total
	}

	return allUsers[start:end], total, nil
}

// Update 更新用户
func (s *InMemoryUserStore) Update(id int, updates *UpdateUserRequest) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("用户不存在")
	}

	// 应用更新
	if updates.Username != nil {
		// 检查用户名冲突
		for _, existingUser := range s.users {
			if existingUser.ID != id && existingUser.Username == *updates.Username {
				return nil, fmt.Errorf("用户名已存在")
			}
		}
		user.Username = *updates.Username
	}

	if updates.Email != nil {
		user.Email = *updates.Email
	}

	if updates.Name != nil {
		user.Name = *updates.Name
	}

	user.UpdatedAt = time.Now()

	// 返回副本
	userCopy := *user
	return &userCopy, nil
}

// Delete 删除用户
func (s *InMemoryUserStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("用户不存在")
	}

	delete(s.users, id)
	return nil
}

// GetByUsername 根据用户名获取用户
func (s *InMemoryUserStore) GetByUsername(username string) (*User, error) {
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

// =============================================================================
// 4. 路由器实现
// =============================================================================

// Route 路由定义
type Route struct {
	Method  string
	Pattern string
	Handler http.HandlerFunc
}

// Router 简单路由器
type Router struct {
	routes []Route
}

// NewRouter 创建新路由器
func NewRouter() *Router {
	return &Router{
		routes: make([]Route, 0),
	}
}

// AddRoute 添加路由
func (r *Router) AddRoute(method, pattern string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		Method:  method,
		Pattern: pattern,
		Handler: handler,
	})
}

// GET 添加GET路由
func (r *Router) GET(pattern string, handler http.HandlerFunc) {
	r.AddRoute("GET", pattern, handler)
}

// POST 添加POST路由
func (r *Router) POST(pattern string, handler http.HandlerFunc) {
	r.AddRoute("POST", pattern, handler)
}

// PUT 添加PUT路由
func (r *Router) PUT(pattern string, handler http.HandlerFunc) {
	r.AddRoute("PUT", pattern, handler)
}

// DELETE 添加DELETE路由
func (r *Router) DELETE(pattern string, handler http.HandlerFunc) {
	r.AddRoute("DELETE", pattern, handler)
}

// PATCH 添加PATCH路由
func (r *Router) PATCH(pattern string, handler http.HandlerFunc) {
	r.AddRoute("PATCH", pattern, handler)
}

// ServeHTTP 实现http.Handler接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if route.Method == req.Method && r.matchPattern(route.Pattern, req.URL.Path) {
			// 提取路径参数
			params := r.extractParams(route.Pattern, req.URL.Path)

			// 将参数添加到请求上下文中（简化版本，实际应使用context）
			if len(params) > 0 {
				req.Header.Set("X-Path-Params", encodeParams(params))
			}

			route.Handler(w, req)
			return
		}
	}

	// 没有找到匹配的路由
	http.NotFound(w, req)
}

// matchPattern 匹配路径模式
func (r *Router) matchPattern(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// 路径参数，跳过匹配
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}

	return true
}

// extractParams 提取路径参数
func (r *Router) extractParams(pattern, path string) map[string]string {
	params := make(map[string]string)
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := part[1 : len(part)-1]
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// encodeParams 编码参数（简化版本）
func encodeParams(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "&")
}

// getPathParam 从请求中获取路径参数
func getPathParam(r *http.Request, name string) string {
	paramStr := r.Header.Get("X-Path-Params")
	if paramStr == "" {
		return ""
	}

	parts := strings.Split(paramStr, "&")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 && kv[0] == name {
			return kv[1]
		}
	}
	return ""
}

// =============================================================================
// 5. API控制器
// =============================================================================

// UserController 用户控制器
type UserController struct {
	store UserStore
}

// NewUserController 创建用户控制器
func NewUserController(store UserStore) *UserController {
	return &UserController{store: store}
}

// GetUsers 获取用户列表
func (uc *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	users, total, err := uc.store.GetAll(page, limit)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	totalPages := (total + limit - 1) / limit

	response := PaginatedResponse{
		Success: true,
		Data:    users,
		Meta: PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// GetUser 获取单个用户
func (uc *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := getPathParam(r, "id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := uc.store.GetByID(id)
	if err != nil {
		writeErrorResponse(w, http.StatusNotFound, "用户不存在")
		return
	}

	writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    user,
	})
}

// CreateUser 创建用户
func (uc *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 验证请求
	if req.Username == "" || req.Email == "" || req.Name == "" {
		writeErrorResponse(w, http.StatusBadRequest, "用户名、邮箱和姓名不能为空")
		return
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Name:     req.Name,
	}

	if err := uc.store.Create(user); err != nil {
		if strings.Contains(err.Error(), "已存在") {
			writeErrorResponse(w, http.StatusConflict, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "创建用户失败")
		}
		return
	}

	writeJSONResponse(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    user,
		Message: "用户创建成功",
	})
}

// UpdateUser 更新用户
func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := getPathParam(r, "id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	user, err := uc.store.Update(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "已存在") {
			writeErrorResponse(w, http.StatusConflict, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "更新用户失败")
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    user,
		Message: "用户更新成功",
	})
}

// DeleteUser 删除用户
func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := getPathParam(r, "id")
	if idStr == "" {
		writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	if err := uc.store.Delete(id); err != nil {
		if strings.Contains(err.Error(), "不存在") {
			writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "删除用户失败")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// =============================================================================
// 6. 工具函数
// =============================================================================

// writeJSONResponse 写入JSON响应
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse 写入错误响应
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}
	writeJSONResponse(w, statusCode, response)
}

// =============================================================================
// 7. 中间件
// =============================================================================

// jsonMiddleware JSON中间件
func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware CORS中间件
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 日志中间件
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装ResponseWriter以捕获状态码
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

// responseWriter 包装http.ResponseWriter以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// =============================================================================
// 8. API文档和健康检查
// =============================================================================

// healthCheckHandler 健康检查处理器
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
	}

	writeJSONResponse(w, http.StatusOK, response)
}

var startTime = time.Now()

// apiDocsHandler API文档处理器
func apiDocsHandler(w http.ResponseWriter, r *http.Request) {
	docs := map[string]interface{}{
		"title":       "User Management API",
		"version":     "1.0.0",
		"description": "RESTful API for user management",
		"endpoints": map[string]interface{}{
			"GET /api/v1/users": map[string]string{
				"description": "获取用户列表",
				"parameters":  "page (int), limit (int)",
			},
			"GET /api/v1/users/{id}": map[string]string{
				"description": "获取特定用户",
				"parameters":  "id (int)",
			},
			"POST /api/v1/users": map[string]string{
				"description": "创建新用户",
				"body":        "CreateUserRequest",
			},
			"PUT /api/v1/users/{id}": map[string]string{
				"description": "更新用户",
				"parameters":  "id (int)",
				"body":        "UpdateUserRequest",
			},
			"DELETE /api/v1/users/{id}": map[string]string{
				"description": "删除用户",
				"parameters":  "id (int)",
			},
			"GET /health": map[string]string{
				"description": "健康检查",
			},
		},
		"schemas": map[string]interface{}{
			"User": map[string]string{
				"id":         "int",
				"username":   "string",
				"email":      "string",
				"name":       "string",
				"created_at": "string (ISO 8601)",
				"updated_at": "string (ISO 8601)",
			},
			"CreateUserRequest": map[string]string{
				"username": "string (required)",
				"email":    "string (required)",
				"name":     "string (required)",
			},
			"UpdateUserRequest": map[string]string{
				"username": "string (optional)",
				"email":    "string (optional)",
				"name":     "string (optional)",
			},
		},
	}

	writeJSONResponse(w, http.StatusOK, docs)
}

// =============================================================================
// 9. 主应用程序
// =============================================================================

func main() {
	fmt.Println("Go Web 开发 - 路由和REST API")
	fmt.Println("============================")

	// 创建用户存储
	userStore := NewInMemoryUserStore()

	// 创建用户控制器
	userController := NewUserController(userStore)

	// 创建路由器
	router := NewRouter()

	// 注册API路由
	router.GET("/api/v1/users", userController.GetUsers)
	router.GET("/api/v1/users/{id}", userController.GetUser)
	router.POST("/api/v1/users", userController.CreateUser)
	router.PUT("/api/v1/users/{id}", userController.UpdateUser)
	router.DELETE("/api/v1/users/{id}", userController.DeleteUser)

	// 注册其他路由
	router.GET("/health", healthCheckHandler)
	router.GET("/docs", apiDocsHandler)

	// 应用中间件
	handler := loggingMiddleware(
		corsMiddleware(
			jsonMiddleware(router),
		),
	)

	// 创建服务器
	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Println("REST API 服务器启动在 :8080")
	fmt.Println("API文档: http://localhost:8080/docs")
	fmt.Println("健康检查: http://localhost:8080/health")
	fmt.Println("\nAPI端点:")
	fmt.Println("GET    /api/v1/users       - 获取用户列表")
	fmt.Println("GET    /api/v1/users/{id}  - 获取特定用户")
	fmt.Println("POST   /api/v1/users       - 创建新用户")
	fmt.Println("PUT    /api/v1/users/{id}  - 更新用户")
	fmt.Println("DELETE /api/v1/users/{id}  - 删除用户")
	fmt.Println("\n示例:")
	fmt.Println("curl http://localhost:8080/api/v1/users")
	fmt.Println("curl -X POST http://localhost:8080/api/v1/users -H 'Content-Type: application/json' -d '{\"username\":\"john\",\"email\":\"john@example.com\",\"name\":\"John Doe\"}'")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

/*
练习任务：
1. 添加用户搜索功能 (GET /api/v1/users/search?q=keyword)
2. 实现用户密码管理 (POST /api/v1/users/{id}/password)
3. 添加批量操作 (POST /api/v1/users/batch)
4. 实现数据验证中间件
5. 添加API版本控制
6. 实现请求限流
7. 添加缓存层
8. 实现软删除功能
9. 添加字段过滤 (GET /api/v1/users?fields=id,name)
10. 实现排序功能 (GET /api/v1/users?sort=name,-created_at)
*/
