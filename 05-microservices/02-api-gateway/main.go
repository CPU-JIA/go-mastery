package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

/*
微服务架构 - API网关和路由练习

本练习涵盖微服务架构中的API网关模式，包括：
1. 请求路由和转发
2. 负载均衡和故障转移
3. 认证和授权
4. 速率限制和熔断
5. 请求/响应转换
6. 协议转换
7. 监控和日志
8. 缓存和性能优化

主要概念：
- API网关模式
- 反向代理
- 服务聚合
- 横切关注点
- 微服务治理
*/

// === 路由规则定义 ===

type RouteRule struct {
	ID          string            `json:"id"`
	Path        string            `json:"path"`         // 匹配路径
	Method      string            `json:"method"`       // HTTP方法
	ServiceName string            `json:"service_name"` // 目标服务名
	TargetPath  string            `json:"target_path"`  // 目标路径
	Rewrite     bool              `json:"rewrite"`      // 是否重写路径
	Middleware  []string          `json:"middleware"`   // 中间件列表
	Metadata    map[string]string `json:"metadata"`     // 元数据
	Weight      int               `json:"weight"`       // 权重（用于A/B测试）
	Enabled     bool              `json:"enabled"`      // 是否启用
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type ServiceEndpoint struct {
	ServiceName string   `json:"service_name"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	Protocol    string   `json:"protocol"` // http, https, grpc
	Health      string   `json:"health"`   // healthy, unhealthy
	Weight      int      `json:"weight"`   // 负载均衡权重
	Tags        []string `json:"tags"`
}

// === 路由管理器 ===

type RouteManager struct {
	routes    map[string]*RouteRule
	endpoints map[string][]*ServiceEndpoint
	mutex     sync.RWMutex
}

func NewRouteManager() *RouteManager {
	rm := &RouteManager{
		routes:    make(map[string]*RouteRule),
		endpoints: make(map[string][]*ServiceEndpoint),
	}

	// 初始化默认路由
	rm.initDefaultRoutes()

	return rm
}

func (rm *RouteManager) initDefaultRoutes() {
	defaultRoutes := []*RouteRule{
		{
			ID:          "user-service-route",
			Path:        "/api/users",
			Method:      "*",
			ServiceName: "user-service",
			TargetPath:  "/users",
			Rewrite:     true,
			Middleware:  []string{"auth", "ratelimit", "logging"},
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "order-service-route",
			Path:        "/api/orders",
			Method:      "*",
			ServiceName: "order-service",
			TargetPath:  "/orders",
			Rewrite:     true,
			Middleware:  []string{"auth", "ratelimit", "logging"},
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "payment-service-route",
			Path:        "/api/payments",
			Method:      "*",
			ServiceName: "payment-service",
			TargetPath:  "/payments",
			Rewrite:     true,
			Middleware:  []string{"auth", "ratelimit", "logging"},
			Enabled:     true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, route := range defaultRoutes {
		rm.routes[route.ID] = route
	}

	// 初始化默认服务端点
	defaultEndpoints := map[string][]*ServiceEndpoint{
		"user-service": {
			{ServiceName: "user-service", Host: "localhost", Port: 8081, Protocol: "http", Health: "healthy", Weight: 100},
			{ServiceName: "user-service", Host: "localhost", Port: 8082, Protocol: "http", Health: "healthy", Weight: 100},
		},
		"order-service": {
			{ServiceName: "order-service", Host: "localhost", Port: 8083, Protocol: "http", Health: "healthy", Weight: 100},
		},
		"payment-service": {
			{ServiceName: "payment-service", Host: "localhost", Port: 8084, Protocol: "http", Health: "healthy", Weight: 100},
		},
	}

	rm.endpoints = defaultEndpoints
}

func (rm *RouteManager) AddRoute(route *RouteRule) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()
	rm.routes[route.ID] = route
}

func (rm *RouteManager) GetRoute(path, method string) *RouteRule {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	for _, route := range rm.routes {
		if !route.Enabled {
			continue
		}

		// 方法匹配
		if route.Method != "*" && route.Method != method {
			continue
		}

		// 路径匹配（支持简单的通配符）
		if rm.matchPath(route.Path, path) {
			return route
		}
	}

	return nil
}

func (rm *RouteManager) matchPath(pattern, path string) bool {
	// 简单的路径匹配实现
	if pattern == path {
		return true
	}

	// 支持前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	// 支持正则匹配
	matched, _ := regexp.MatchString(pattern, path)
	return matched
}

func (rm *RouteManager) GetEndpoints(serviceName string) []*ServiceEndpoint {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	endpoints := rm.endpoints[serviceName]
	var healthyEndpoints []*ServiceEndpoint

	for _, endpoint := range endpoints {
		if endpoint.Health == "healthy" {
			healthyEndpoints = append(healthyEndpoints, endpoint)
		}
	}

	return healthyEndpoints
}

func (rm *RouteManager) UpdateEndpointHealth(serviceName, host string, port int, health string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	endpoints := rm.endpoints[serviceName]
	for _, endpoint := range endpoints {
		if endpoint.Host == host && endpoint.Port == port {
			endpoint.Health = health
			break
		}
	}
}

// === 负载均衡器 ===

type LoadBalancer interface {
	Select(endpoints []*ServiceEndpoint) *ServiceEndpoint
}

// 加权轮询负载均衡
type WeightedRoundRobinBalancer struct {
	counters map[string]int
	mutex    sync.Mutex
}

func NewWeightedRoundRobinBalancer() *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{
		counters: make(map[string]int),
	}
}

func (b *WeightedRoundRobinBalancer) Select(endpoints []*ServiceEndpoint) *ServiceEndpoint {
	if len(endpoints) == 0 {
		return nil
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 计算总权重
	totalWeight := 0
	for _, endpoint := range endpoints {
		totalWeight += endpoint.Weight
	}

	// 获取服务名（假设所有端点都是同一服务）
	serviceName := endpoints[0].ServiceName
	counter := b.counters[serviceName]

	// 根据权重选择端点
	currentWeight := 0
	for _, endpoint := range endpoints {
		currentWeight += endpoint.Weight
		if counter%totalWeight < currentWeight {
			b.counters[serviceName] = (counter + 1) % totalWeight
			return endpoint
		}
	}

	// 兜底返回第一个
	b.counters[serviceName] = (counter + 1) % totalWeight
	return endpoints[0]
}

// === 中间件系统 ===

type Middleware interface {
	Name() string
	Process(ctx *GatewayContext) error
}

type GatewayContext struct {
	Request   *http.Request
	Response  http.ResponseWriter
	Route     *RouteRule
	Endpoint  *ServiceEndpoint
	Metadata  map[string]interface{}
	StartTime time.Time
	RequestID string
	UserID    string
	ClientIP  string
}

// 认证中间件
type AuthMiddleware struct {
	validTokens map[string]string
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		validTokens: map[string]string{
			"token123": "user1",
			"token456": "user2",
			"admin789": "admin",
		},
	}
}

func (m *AuthMiddleware) Name() string {
	return "auth"
}

func (m *AuthMiddleware) Process(ctx *GatewayContext) error {
	// 从Authorization头获取token
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("缺少Authorization头")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	userID, exists := m.validTokens[token]
	if !exists {
		return fmt.Errorf("无效的token")
	}

	ctx.UserID = userID
	ctx.Metadata["user_id"] = userID

	return nil
}

// 速率限制中间件
type RateLimitMiddleware struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimitMiddleware(r rate.Limit, b int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

func (m *RateLimitMiddleware) Name() string {
	return "ratelimit"
}

func (m *RateLimitMiddleware) Process(ctx *GatewayContext) error {
	key := ctx.ClientIP
	if ctx.UserID != "" {
		key = ctx.UserID
	}

	limiter := m.getLimiter(key)
	if !limiter.Allow() {
		return fmt.Errorf("请求频率超限")
	}

	return nil
}

func (m *RateLimitMiddleware) getLimiter(key string) *rate.Limiter {
	m.mutex.RLock()
	limiter, exists := m.limiters[key]
	m.mutex.RUnlock()

	if exists {
		return limiter
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	limiter = rate.NewLimiter(m.rate, m.burst)
	m.limiters[key] = limiter

	return limiter
}

// 日志中间件
type LoggingMiddleware struct {
	logger *log.Logger
}

func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) Name() string {
	return "logging"
}

func (m *LoggingMiddleware) Process(ctx *GatewayContext) error {
	m.logger.Printf("[%s] %s %s -> %s:%d %s",
		ctx.RequestID,
		ctx.Request.Method,
		ctx.Request.URL.Path,
		ctx.Endpoint.Host,
		ctx.Endpoint.Port,
		ctx.UserID,
	)
	return nil
}

// 响应转换中间件
type ResponseTransformMiddleware struct{}

func NewResponseTransformMiddleware() *ResponseTransformMiddleware {
	return &ResponseTransformMiddleware{}
}

func (m *ResponseTransformMiddleware) Name() string {
	return "transform"
}

func (m *ResponseTransformMiddleware) Process(ctx *GatewayContext) error {
	// 添加通用响应头
	ctx.Response.Header().Set("X-Gateway", "go-api-gateway")
	ctx.Response.Header().Set("X-Request-ID", ctx.RequestID)
	ctx.Response.Header().Set("X-Processing-Time", time.Since(ctx.StartTime).String())

	return nil
}

// === 熔断器 ===

type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailTime time.Time
	state        string // closed, open, half-open
	mutex        sync.Mutex
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// 检查是否可以执行
	if cb.state == "open" {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = "half-open"
		} else {
			return fmt.Errorf("熔断器开启，拒绝请求")
		}
	}

	// 执行函数
	err := fn()
	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	cb.state = "closed"
}

func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = "open"
	}
}

// === API网关核心 ===

type APIGateway struct {
	routeManager   *RouteManager
	loadBalancer   LoadBalancer
	middlewares    map[string]Middleware
	circuitBreaker *CircuitBreaker
	httpClient     *http.Client
}

func NewAPIGateway() *APIGateway {
	gateway := &APIGateway{
		routeManager:   NewRouteManager(),
		loadBalancer:   NewWeightedRoundRobinBalancer(),
		middlewares:    make(map[string]Middleware),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}

	// 注册中间件
	gateway.middlewares["auth"] = NewAuthMiddleware()
	gateway.middlewares["ratelimit"] = NewRateLimitMiddleware(rate.Limit(10), 20)
	gateway.middlewares["logging"] = NewLoggingMiddleware(log.Default())
	gateway.middlewares["transform"] = NewResponseTransformMiddleware()

	return gateway
}

func (g *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := generateRequestID()

	// 创建网关上下文
	ctx := &GatewayContext{
		Request:   r,
		Response:  w,
		Metadata:  make(map[string]interface{}),
		StartTime: startTime,
		RequestID: requestID,
		ClientIP:  getClientIP(r),
	}

	// 查找路由
	route := g.routeManager.GetRoute(r.URL.Path, r.Method)
	if route == nil {
		http.Error(w, "路由不存在", http.StatusNotFound)
		return
	}

	ctx.Route = route

	// 获取服务端点
	endpoints := g.routeManager.GetEndpoints(route.ServiceName)
	if len(endpoints) == 0 {
		http.Error(w, "服务不可用", http.StatusServiceUnavailable)
		return
	}

	// 负载均衡选择端点
	endpoint := g.loadBalancer.Select(endpoints)
	if endpoint == nil {
		http.Error(w, "负载均衡失败", http.StatusServiceUnavailable)
		return
	}

	ctx.Endpoint = endpoint

	// 执行中间件
	for _, middlewareName := range route.Middleware {
		middleware, exists := g.middlewares[middlewareName]
		if !exists {
			continue
		}

		if err := middleware.Process(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}

	// 代理请求
	err := g.circuitBreaker.Call(func() error {
		return g.proxyRequest(ctx)
	})

	if err != nil {
		log.Printf("代理请求失败: %v", err)
		http.Error(w, "服务错误", http.StatusBadGateway)
	}
}

func (g *APIGateway) proxyRequest(ctx *GatewayContext) error {
	// 构建目标URL
	targetURL := fmt.Sprintf("%s://%s:%d%s",
		ctx.Endpoint.Protocol,
		ctx.Endpoint.Host,
		ctx.Endpoint.Port,
		ctx.Route.TargetPath)

	if !ctx.Route.Rewrite {
		targetURL = fmt.Sprintf("%s://%s:%d%s",
			ctx.Endpoint.Protocol,
			ctx.Endpoint.Host,
			ctx.Endpoint.Port,
			ctx.Request.URL.Path)
	}

	// 创建代理请求
	proxyURL, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("解析目标URL失败: %w", err)
	}

	// 使用httputil.ReverseProxy进行代理
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL = proxyURL
			req.Host = proxyURL.Host

			// 添加追踪头
			req.Header.Set("X-Request-ID", ctx.RequestID)
			req.Header.Set("X-Forwarded-For", ctx.ClientIP)
			req.Header.Set("X-Gateway-User", ctx.UserID)
		},
		ModifyResponse: func(resp *http.Response) error {
			// 修改响应头
			resp.Header.Set("X-Gateway", "go-api-gateway")
			resp.Header.Set("X-Request-ID", ctx.RequestID)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("代理错误: %v", err)
			// 标记端点为不健康
			g.routeManager.UpdateEndpointHealth(
				ctx.Route.ServiceName,
				ctx.Endpoint.Host,
				ctx.Endpoint.Port,
				"unhealthy",
			)
			http.Error(w, "服务不可用", http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(ctx.Response, ctx.Request)
	return nil
}

// === 管理API ===

type GatewayAdmin struct {
	gateway *APIGateway
}

func NewGatewayAdmin(gateway *APIGateway) *GatewayAdmin {
	return &GatewayAdmin{gateway: gateway}
}

func (a *GatewayAdmin) GetRoutes(w http.ResponseWriter, r *http.Request) {
	a.gateway.routeManager.mutex.RLock()
	routes := make([]*RouteRule, 0, len(a.gateway.routeManager.routes))
	for _, route := range a.gateway.routeManager.routes {
		routes = append(routes, route)
	}
	a.gateway.routeManager.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

func (a *GatewayAdmin) CreateRoute(w http.ResponseWriter, r *http.Request) {
	var route RouteRule
	if err := json.NewDecoder(r.Body).Decode(&route); err != nil {
		http.Error(w, "无效的路由配置", http.StatusBadRequest)
		return
	}

	// 验证路由配置
	if route.ID == "" || route.Path == "" || route.ServiceName == "" {
		http.Error(w, "路由配置不完整", http.StatusBadRequest)
		return
	}

	a.gateway.routeManager.AddRoute(&route)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(route)
}

func (a *GatewayAdmin) GetEndpoints(w http.ResponseWriter, r *http.Request) {
	a.gateway.routeManager.mutex.RLock()
	endpoints := make(map[string][]*ServiceEndpoint)
	for serviceName, serviceEndpoints := range a.gateway.routeManager.endpoints {
		endpoints[serviceName] = serviceEndpoints
	}
	a.gateway.routeManager.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(endpoints)
}

func (a *GatewayAdmin) UpdateEndpointHealth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	var req struct {
		Host   string `json:"host"`
		Port   int    `json:"port"`
		Health string `json:"health"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	a.gateway.routeManager.UpdateEndpointHealth(serviceName, req.Host, req.Port, req.Health)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "端点状态已更新"})
}

// === 监控和指标 ===

type GatewayMetrics struct {
	RequestCount    int64                      `json:"request_count"`
	ErrorCount      int64                      `json:"error_count"`
	ResponseTimes   []time.Duration            `json:"-"`
	AvgResponseTime time.Duration              `json:"avg_response_time"`
	ServiceMetrics  map[string]*ServiceMetrics `json:"service_metrics"`
	mutex           sync.RWMutex
}

type ServiceMetrics struct {
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
}

func NewGatewayMetrics() *GatewayMetrics {
	return &GatewayMetrics{
		ServiceMetrics: make(map[string]*ServiceMetrics),
	}
}

func (m *GatewayMetrics) RecordRequest(serviceName string, duration time.Duration, isError bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.RequestCount++
	m.ResponseTimes = append(m.ResponseTimes, duration)

	if isError {
		m.ErrorCount++
	}

	// 记录服务指标
	if m.ServiceMetrics[serviceName] == nil {
		m.ServiceMetrics[serviceName] = &ServiceMetrics{}
	}

	serviceMetrics := m.ServiceMetrics[serviceName]
	serviceMetrics.RequestCount++
	if isError {
		serviceMetrics.ErrorCount++
	}

	// 计算平均响应时间
	m.calculateAverageResponseTime()
}

func (m *GatewayMetrics) calculateAverageResponseTime() {
	if len(m.ResponseTimes) == 0 {
		return
	}

	var total time.Duration
	for _, duration := range m.ResponseTimes {
		total += duration
	}

	m.AvgResponseTime = total / time.Duration(len(m.ResponseTimes))

	// 保持最近1000个记录
	if len(m.ResponseTimes) > 1000 {
		m.ResponseTimes = m.ResponseTimes[len(m.ResponseTimes)-1000:]
	}
}

func (m *GatewayMetrics) GetMetrics() *GatewayMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回副本
	metrics := &GatewayMetrics{
		RequestCount:    m.RequestCount,
		ErrorCount:      m.ErrorCount,
		AvgResponseTime: m.AvgResponseTime,
		ServiceMetrics:  make(map[string]*ServiceMetrics),
	}

	for serviceName, serviceMetrics := range m.ServiceMetrics {
		metrics.ServiceMetrics[serviceName] = &ServiceMetrics{
			RequestCount:    serviceMetrics.RequestCount,
			ErrorCount:      serviceMetrics.ErrorCount,
			AvgResponseTime: serviceMetrics.AvgResponseTime,
		}
	}

	return metrics
}

// === 辅助函数 ===

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// === 示例后端服务 ===

func startMockService(name string, port int) {
	router := mux.NewRouter()

	// 模拟业务端点
	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		users := []map[string]interface{}{
			{"id": 1, "name": "Alice", "service": name},
			{"id": 2, "name": "Bob", "service": name},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	}).Methods("GET")

	router.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		orders := []map[string]interface{}{
			{"id": 1, "product": "Laptop", "service": name},
			{"id": 2, "product": "Phone", "service": name},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(orders)
	}).Methods("GET")

	router.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		payments := []map[string]interface{}{
			{"id": 1, "amount": 1000, "service": name},
			{"id": 2, "amount": 500, "service": name},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payments)
	}).Methods("GET")

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":  "healthy",
			"service": name,
			"port":    port,
			"time":    time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}).Methods("GET")

	log.Printf("模拟服务 %s 启动在端口 %d", name, port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}

func main() {
	// 启动模拟后端服务
	go startMockService("user-service-1", 8081)
	go startMockService("user-service-2", 8082)
	go startMockService("order-service", 8083)
	go startMockService("payment-service", 8084)

	// 等待服务启动
	time.Sleep(2 * time.Second)

	// 创建API网关
	gateway := NewAPIGateway()
	admin := NewGatewayAdmin(gateway)
	metrics := NewGatewayMetrics()

	// 创建路由器
	router := mux.NewRouter()

	// 管理API
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/routes", admin.GetRoutes).Methods("GET")
	adminRouter.HandleFunc("/routes", admin.CreateRoute).Methods("POST")
	adminRouter.HandleFunc("/endpoints", admin.GetEndpoints).Methods("GET")
	adminRouter.HandleFunc("/endpoints/{service}/health", admin.UpdateEndpointHealth).Methods("PUT")

	// 监控API
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics.GetMetrics())
	}).Methods("GET")

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]interface{}{
			"status":    "healthy",
			"service":   "api-gateway",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(health)
	}).Methods("GET")

	// 添加指标中间件
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapper := &responseWrapper{ResponseWriter: w}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)
			isError := wrapper.statusCode >= 400

			// 记录指标
			metrics.RecordRequest("gateway", duration, isError)
		})
	})

	// API网关路由（必须放在最后）
	router.PathPrefix("/api").Handler(gateway)

	fmt.Println("=== API网关启动 ===")
	fmt.Println("网关端点:")
	fmt.Println("  主网关:     http://localhost:8080")
	fmt.Println("  管理API:    http://localhost:8080/admin")
	fmt.Println("  监控指标:   http://localhost:8080/metrics")
	fmt.Println("  健康检查:   http://localhost:8080/health")
	fmt.Println()
	fmt.Println("业务API:")
	fmt.Println("  用户服务:   GET /api/users")
	fmt.Println("  订单服务:   GET /api/orders")
	fmt.Println("  支付服务:   GET /api/payments")
	fmt.Println()
	fmt.Println("管理API:")
	fmt.Println("  GET  /admin/routes         - 获取路由配置")
	fmt.Println("  POST /admin/routes         - 创建路由")
	fmt.Println("  GET  /admin/endpoints      - 获取服务端点")
	fmt.Println("  PUT  /admin/endpoints/{service}/health - 更新端点健康状态")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  # 需要认证的请求")
	fmt.Println(`  curl -H "Authorization: Bearer token123" http://localhost:8080/api/users`)
	fmt.Println("  # 查看路由配置")
	fmt.Println("  curl http://localhost:8080/admin/routes")
	fmt.Println("  # 查看监控指标")
	fmt.Println("  curl http://localhost:8080/metrics")

	log.Fatal(http.ListenAndServe(":8080", router))
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWrapper) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(data)
}

/*
练习任务：

1. 基础练习：
   - 实现更多路由匹配策略（正则、通配符）
   - 添加请求/响应转换功能
   - 实现API版本管理
   - 添加缓存中间件

2. 中级练习：
   - 实现WebSocket代理
   - 添加A/B测试支持
   - 实现蓝绿部署
   - 添加服务降级机制

3. 高级练习：
   - 实现GraphQL聚合
   - 添加gRPC代理支持
   - 实现分布式追踪
   - 集成服务网格

4. 监控和运维：
   - 实现实时监控仪表板
   - 添加告警机制
   - 实现日志聚合
   - 添加性能分析

5. 安全和治理：
   - 实现OAuth2集成
   - 添加API密钥管理
   - 实现IP白名单
   - 添加防护机制

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get golang.org/x/time/rate

2. 运行程序：go run main.go

网关架构：
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    客户端    │────│  API网关     │────│ 服务注册中心 │
└─────────────┘    └─────────────┘    └─────────────┘
                         │
               ┌─────────┼─────────┐
               │         │         │
        ┌─────────┐ ┌─────────┐ ┌─────────┐
        │路由管理器│ │负载均衡器│ │ 中间件  │
        └─────────┘ └─────────┘ └─────────┘
               │         │         │
               └─────────┼─────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
   ┌─────────┐    ┌─────────┐    ┌─────────┐
   │用户服务1 │    │ 订单服务 │    │ 支付服务 │
   │ (8081)  │    │ (8083)  │    │ (8084)  │
   └─────────┘    └─────────┘    └─────────┘
   ┌─────────┐
   │用户服务2 │
   │ (8082)  │
   └─────────┘

扩展建议：
- 集成Kubernetes Ingress Controller
- 实现动态配置热更新
- 添加流量复制功能
- 实现多协议支持（HTTP/2、gRPC）
*/
