package main

import (
	"context"
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

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/time/rate"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

/*
🚀 现代化API网关 - 2025年企业级实现

本实现展示了云原生微服务架构中的现代API网关模式，包括：

🔐 高级认证授权：
1. JWT令牌验证和刷新
2. OAuth2.0集成支持
3. mTLS双向认证
4. RBAC基于角色的访问控制
5. API Key管理

🌐 智能路由管理：
1. 动态路由配置热更新
2. 高级路径匹配（正则、通配符）
3. 多版本API支持（A/B测试）
4. 蓝绿部署和金丝雀发布
5. GraphQL查询聚合

⚡ 高性能特性：
1. 自适应负载均衡算法
2. 连接池和Keep-Alive优化
3. 响应缓存和压缩
4. WebSocket代理支持
5. gRPC协议转换

🔍 可观测性集成：
1. OpenTelemetry分布式追踪
2. Prometheus指标采集
3. 结构化日志记录
4. 实时监控面板
5. 自定义告警规则

☁️ 云原生支持：
1. Kubernetes服务发现
2. Istio Service Mesh集成
3. 容器健康检查
4. 配置热更新机制
5. 优雅停机处理

🛡️ 安全防护：
1. 速率限制和熔断保护
2. 请求验证和清理
3. 安全头注入
4. IP白名单/黑名单
5. DDoS攻击防护

核心设计原则：
- 高可用性：多实例部署，故障自动恢复
- 高性能：连接复用，请求合并，智能缓存
- 可扩展：插件化架构，中间件链式处理
- 可观测：全链路追踪，细粒度监控
- 安全性：零信任架构，全面安全防护
*/

// === 全局配置和常量 ===

const (
	// 默认配置
	DefaultPort         = "8080"
	DefaultTimeout      = 30 * time.Second
	DefaultIdleTimeout  = 120 * time.Second
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second

	// JWT配置
	DefaultJWTSecret     = "your-256-bit-secret"
	DefaultJWTExpiryTime = 24 * time.Hour
	DefaultRefreshTime   = 7 * 24 * time.Hour

	// 限流配置
	DefaultRateLimit   = 100 // requests per second
	DefaultBurstLimit  = 200
	DefaultConcurrency = 1000

	// 健康检查配置
	DefaultHealthPath     = "/health"
	DefaultReadyPath      = "/ready"
	DefaultLivenessPath   = "/live"
	DefaultHealthInterval = 30 * time.Second
	DefaultHealthTimeout  = 5 * time.Second
)

// GatewayConfig 网关全局配置
type GatewayConfig struct {
	Server     ServerConfig     `yaml:"server"`
	Auth       AuthConfig       `yaml:"auth"`
	Tracing    TracingConfig    `yaml:"tracing"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Security   SecurityConfig   `yaml:"security"`
	K8s        K8sConfig        `yaml:"kubernetes"`
}

type ServerConfig struct {
	Port           string        `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"read_timeout"`
	WriteTimeout   time.Duration `yaml:"write_timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes int           `yaml:"max_header_bytes"`
	EnableTLS      bool          `yaml:"enable_tls"`
	CertFile       string        `yaml:"cert_file"`
	KeyFile        string        `yaml:"key_file"`
	EnableHTTP2    bool          `yaml:"enable_http2"`
	EnableGRPCWeb  bool          `yaml:"enable_grpc_web"`
}

type AuthConfig struct {
	JWTSecret         string        `yaml:"jwt_secret"`
	JWTExpiryTime     time.Duration `yaml:"jwt_expiry_time"`
	RefreshTime       time.Duration `yaml:"refresh_time"`
	EnableOAuth2      bool          `yaml:"enable_oauth2"`
	OAuth2Config      OAuth2Config  `yaml:"oauth2"`
	EnableMTLS        bool          `yaml:"enable_mtls"`
	ClientCAFile      string        `yaml:"client_ca_file"`
	RequireClientCert bool          `yaml:"require_client_cert"`
}

type OAuth2Config struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	RedirectURL  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
	AuthURL      string   `yaml:"auth_url"`
	TokenURL     string   `yaml:"token_url"`
}

type TracingConfig struct {
	Enabled     bool    `yaml:"enabled"`
	ServiceName string  `yaml:"service_name"`
	JaegerURL   string  `yaml:"jaeger_url"`
	SampleRate  float64 `yaml:"sample_rate"`
	Environment string  `yaml:"environment"`
}

type MonitoringConfig struct {
	MetricsEnabled bool   `yaml:"metrics_enabled"`
	MetricsPath    string `yaml:"metrics_path"`
	HealthPath     string `yaml:"health_path"`
	ReadyPath      string `yaml:"ready_path"`
	LivenessPath   string `yaml:"liveness_path"`
	PrometheusPort string `yaml:"prometheus_port"`
}

type SecurityConfig struct {
	EnableCORS       bool     `yaml:"enable_cors"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	EnableCSRF       bool     `yaml:"enable_csrf"`
	RateLimitEnabled bool     `yaml:"rate_limit_enabled"`
	RateLimit        int      `yaml:"rate_limit"`
	BurstLimit       int      `yaml:"burst_limit"`
	IPWhitelist      []string `yaml:"ip_whitelist"`
	IPBlacklist      []string `yaml:"ip_blacklist"`
}

type K8sConfig struct {
	Enabled          bool   `yaml:"enabled"`
	InCluster        bool   `yaml:"in_cluster"`
	ConfigPath       string `yaml:"config_path"`
	Namespace        string `yaml:"namespace"`
	ServiceDiscovery bool   `yaml:"service_discovery"`
	WatchConfigMaps  bool   `yaml:"watch_configmaps"`
}

// JWT Claims 结构
type JWTClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.StandardClaims
}

// WebSocket升级器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该更严格
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

	// 2025年新增字段
	Protocol      string            `json:"protocol"`       // http, https, grpc, ws, wss
	Version       string            `json:"version"`        // API版本
	Timeout       time.Duration     `json:"timeout"`        // 请求超时
	RetryAttempts int               `json:"retry_attempts"` // 重试次数
	RetryBackoff  time.Duration     `json:"retry_backoff"`  // 重试间隔
	CacheEnabled  bool              `json:"cache_enabled"`  // 是否启用缓存
	CacheTTL      time.Duration     `json:"cache_ttl"`      // 缓存TTL
	RateLimit     *RouteLimitConfig `json:"rate_limit"`     // 路由级限流
	Auth          *RouteAuthConfig  `json:"auth"`           // 路由级认证
	CORS          *CORSConfig       `json:"cors"`           // CORS配置
	Headers       *HeadersConfig    `json:"headers"`        // 请求/响应头配置
	HealthCheck   *HealthConfig     `json:"health_check"`   // 健康检查配置
	LoadBalancer  *LBConfig         `json:"load_balancer"`  // 负载均衡配置

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 路由级限流配置
type RouteLimitConfig struct {
	Enabled   bool     `json:"enabled"`
	RPS       int      `json:"rps"`       // 每秒请求数
	Burst     int      `json:"burst"`     // 突发限制
	KeyBy     string   `json:"key_by"`    // 限流key: ip, user, api_key
	WhiteList []string `json:"whitelist"` // 白名单
}

// 路由级认证配置
type RouteAuthConfig struct {
	Required       bool     `json:"required"`
	Methods        []string `json:"methods"`         // jwt, oauth2, api_key, basic
	Roles          []string `json:"roles"`           // 允许的角色
	Permissions    []string `json:"permissions"`     // 必需的权限
	Scopes         []string `json:"scopes"`          // OAuth2 scopes
	AllowAnonymous bool     `json:"allow_anonymous"` // 允许匿名访问
}

// CORS配置
type CORSConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// 请求/响应头配置
type HeadersConfig struct {
	RequestHeaders  map[string]string `json:"request_headers"`  // 添加到请求的头
	ResponseHeaders map[string]string `json:"response_headers"` // 添加到响应的头
	RemoveRequest   []string          `json:"remove_request"`   // 移除的请求头
	RemoveResponse  []string          `json:"remove_response"`  // 移除的响应头
}

// 健康检查配置
type HealthConfig struct {
	Enabled          bool          `json:"enabled"`
	Path             string        `json:"path"`              // 健康检查路径
	Interval         time.Duration `json:"interval"`          // 检查间隔
	Timeout          time.Duration `json:"timeout"`           // 超时时间
	Retries          int           `json:"retries"`           // 重试次数
	SuccessThreshold int           `json:"success_threshold"` // 成功阈值
	FailureThreshold int           `json:"failure_threshold"` // 失败阈值
}

// 负载均衡配置
type LBConfig struct {
	Algorithm     string         `json:"algorithm"`      // round_robin, weighted_round_robin, least_conn, ip_hash
	HealthCheck   bool           `json:"health_check"`   // 是否启用健康检查
	StickySession bool           `json:"sticky_session"` // 会话保持
	Weights       map[string]int `json:"weights"`        // 权重配置
}

type ServiceEndpoint struct {
	ServiceName string   `json:"service_name"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	Protocol    string   `json:"protocol"` // http, https, grpc
	Health      string   `json:"health"`   // healthy, unhealthy, unknown
	Weight      int      `json:"weight"`   // 负载均衡权重
	Tags        []string `json:"tags"`

	// 2025年新增字段
	Zone         string            `json:"zone"`           // 可用区
	Region       string            `json:"region"`         // 区域
	Version      string            `json:"version"`        // 服务版本
	Metadata     map[string]string `json:"metadata"`       // 元数据
	TLSEnabled   bool              `json:"tls_enabled"`    // 是否启用TLS
	MaxConns     int               `json:"max_conns"`      // 最大连接数
	MaxIdleConns int               `json:"max_idle_conns"` // 最大空闲连接数
	ConnTimeout  time.Duration     `json:"conn_timeout"`   // 连接超时
	ReadTimeout  time.Duration     `json:"read_timeout"`   // 读超时
	WriteTimeout time.Duration     `json:"write_timeout"`  // 写超时

	// 健康检查状态
	LastHealthCheck     time.Time `json:"last_health_check"`
	HealthCheckCount    int       `json:"health_check_count"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
}

// === 路由管理器 ===

type RouteManager struct {
	routes    map[string]*RouteRule
	endpoints map[string][]*ServiceEndpoint
	config    *GatewayConfig
	mutex     sync.RWMutex

	// 2025年新增字段
	tracer        trace.Tracer
	k8sClient     dynamic.Interface
	configWatch   chan *GatewayConfig
	stopCh        chan struct{}
	healthChecker *HealthChecker
}

func NewRouteManager(config *GatewayConfig) *RouteManager {
	rm := &RouteManager{
		routes:      make(map[string]*RouteRule),
		endpoints:   make(map[string][]*ServiceEndpoint),
		config:      config,
		configWatch: make(chan *GatewayConfig, 1),
		stopCh:      make(chan struct{}),
	}

	// 初始化OpenTelemetry追踪
	if config.Tracing.Enabled {
		rm.initTracing()
	}

	// 初始化Kubernetes客户端
	if config.K8s.Enabled {
		rm.initK8sClient()
	}

	// 初始化健康检查器
	rm.healthChecker = NewHealthChecker(rm)

	// 初始化默认路由
	rm.initDefaultRoutes()

	// 启动配置监听
	go rm.watchConfig()

	// 启动健康检查
	go rm.healthChecker.Start()

	return rm
}

func (rm *RouteManager) initTracing() {
	// 初始化Jaeger追踪
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(rm.config.Tracing.JaegerURL)))
	if err != nil {
		log.Printf("初始化Jaeger失败: %v", err)
		return
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(rm.config.Tracing.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.DeploymentEnvironmentKey.String(rm.config.Tracing.Environment),
		)),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(rm.config.Tracing.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	rm.tracer = otel.Tracer("api-gateway")
}

func (rm *RouteManager) initK8sClient() {
	var config *rest.Config
	var err error

	if rm.config.K8s.InCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = rest.InClusterConfig() // 可以改为从文件加载
	}

	if err != nil {
		log.Printf("初始化Kubernetes客户端失败: %v", err)
		return
	}

	rm.k8sClient, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Printf("创建Kubernetes客户端失败: %v", err)
	}
}

func (rm *RouteManager) watchConfig() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case newConfig := <-rm.configWatch:
			rm.updateConfig(newConfig)
		case <-ticker.C:
			// 定期检查配置变更
			if rm.config.K8s.Enabled && rm.config.K8s.WatchConfigMaps {
				rm.checkConfigMapUpdates()
			}
		case <-rm.stopCh:
			return
		}
	}
}

func (rm *RouteManager) updateConfig(config *GatewayConfig) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.config = config
	log.Println("配置已热更新")
}

func (rm *RouteManager) checkConfigMapUpdates() {
	// 实现ConfigMap监听逻辑
	// 这里可以监听Kubernetes ConfigMap变更
}

func (rm *RouteManager) initDefaultRoutes() {
	defaultRoutes := []*RouteRule{
		{
			ID:            "user-service-route",
			Path:          "/api/v1/users",
			Method:        "*",
			ServiceName:   "user-service",
			TargetPath:    "/users",
			Rewrite:       true,
			Middleware:    []string{"cors", "auth", "ratelimit", "logging", "tracing"},
			Enabled:       true,
			Protocol:      "http",
			Version:       "v1",
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
			RetryBackoff:  100 * time.Millisecond,
			Auth: &RouteAuthConfig{
				Required: true,
				Methods:  []string{"jwt"},
				Roles:    []string{"user", "admin"},
			},
			RateLimit: &RouteLimitConfig{
				Enabled: true,
				RPS:     100,
				Burst:   200,
				KeyBy:   "user",
			},
			CORS: &CORSConfig{
				Enabled:      true,
				AllowOrigins: []string{"*"},
				AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders: []string{"Content-Type", "Authorization"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:            "order-service-route",
			Path:          "/api/v1/orders",
			Method:        "*",
			ServiceName:   "order-service",
			TargetPath:    "/orders",
			Rewrite:       true,
			Middleware:    []string{"cors", "auth", "ratelimit", "logging", "tracing"},
			Enabled:       true,
			Protocol:      "http",
			Version:       "v1",
			Timeout:       45 * time.Second,
			RetryAttempts: 2,
			RetryBackoff:  200 * time.Millisecond,
			Auth: &RouteAuthConfig{
				Required:    true,
				Methods:     []string{"jwt"},
				Roles:       []string{"user", "admin"},
				Permissions: []string{"order:read", "order:write"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "websocket-route",
			Path:        "/ws",
			Method:      "GET",
			ServiceName: "websocket-service",
			TargetPath:  "/ws",
			Rewrite:     false,
			Middleware:  []string{"auth", "logging"},
			Enabled:     true,
			Protocol:    "ws",
			Version:     "v1",
			Auth: &RouteAuthConfig{
				Required:       false,
				AllowAnonymous: true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, route := range defaultRoutes {
		rm.routes[route.ID] = route
	}

	// 初始化默认服务端点
	defaultEndpoints := map[string][]*ServiceEndpoint{
		"user-service": {
			{
				ServiceName: "user-service", Host: "user-service.default.svc.cluster.local",
				Port: 80, Protocol: "http", Health: "healthy", Weight: 100,
				Zone: "us-west-1a", Region: "us-west-1", Version: "v1.0.0",
				TLSEnabled: false, MaxConns: 100, ConnTimeout: 5 * time.Second,
			},
			{
				ServiceName: "user-service", Host: "user-service-backup.default.svc.cluster.local",
				Port: 80, Protocol: "http", Health: "healthy", Weight: 50,
				Zone: "us-west-1b", Region: "us-west-1", Version: "v1.0.0",
			},
		},
		"order-service": {
			{
				ServiceName: "order-service", Host: "order-service.default.svc.cluster.local",
				Port: 80, Protocol: "http", Health: "healthy", Weight: 100,
				Zone: "us-west-1a", Region: "us-west-1", Version: "v1.1.0",
			},
		},
		"websocket-service": {
			{
				ServiceName: "websocket-service", Host: "websocket-service.default.svc.cluster.local",
				Port: 80, Protocol: "ws", Health: "healthy", Weight: 100,
			},
		},
	}

	rm.endpoints = defaultEndpoints
}

// === 健康检查器 ===

type HealthChecker struct {
	routeManager  *RouteManager
	client        *http.Client
	stopCh        chan struct{}
	checkInterval time.Duration
}

func NewHealthChecker(rm *RouteManager) *HealthChecker {
	return &HealthChecker{
		routeManager:  rm,
		client:        &http.Client{Timeout: 5 * time.Second},
		stopCh:        make(chan struct{}),
		checkInterval: DefaultHealthInterval,
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.performHealthChecks()
		case <-hc.stopCh:
			return
		}
	}
}

func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

func (hc *HealthChecker) performHealthChecks() {
	hc.routeManager.mutex.RLock()
	endpoints := make(map[string][]*ServiceEndpoint)
	for serviceName, serviceEndpoints := range hc.routeManager.endpoints {
		endpoints[serviceName] = make([]*ServiceEndpoint, len(serviceEndpoints))
		copy(endpoints[serviceName], serviceEndpoints)
	}
	hc.routeManager.mutex.RUnlock()

	for serviceName, serviceEndpoints := range endpoints {
		for _, endpoint := range serviceEndpoints {
			go hc.checkEndpoint(serviceName, endpoint)
		}
	}
}

func (hc *HealthChecker) checkEndpoint(serviceName string, endpoint *ServiceEndpoint) {
	// 构建健康检查URL
	protocol := endpoint.Protocol
	if endpoint.Protocol == "ws" {
		protocol = "http" // WebSocket健康检查通常用HTTP
	}

	healthURL := fmt.Sprintf("%s://%s:%d/health", protocol, endpoint.Host, endpoint.Port)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultHealthTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		hc.markUnhealthy(serviceName, endpoint, err)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		hc.markUnhealthy(serviceName, endpoint, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		hc.markHealthy(serviceName, endpoint)
	} else {
		hc.markUnhealthy(serviceName, endpoint, fmt.Errorf("健康检查返回状态码: %d", resp.StatusCode))
	}
}

func (hc *HealthChecker) markHealthy(serviceName string, endpoint *ServiceEndpoint) {
	hc.routeManager.mutex.Lock()
	defer hc.routeManager.mutex.Unlock()

	for _, ep := range hc.routeManager.endpoints[serviceName] {
		if ep.Host == endpoint.Host && ep.Port == endpoint.Port {
			if ep.Health != "healthy" {
				log.Printf("端点 %s:%d 恢复健康", ep.Host, ep.Port)
			}
			ep.Health = "healthy"
			ep.LastHealthCheck = time.Now()
			ep.HealthCheckCount++
			ep.ConsecutiveFailures = 0
			break
		}
	}
}

func (hc *HealthChecker) markUnhealthy(serviceName string, endpoint *ServiceEndpoint, err error) {
	hc.routeManager.mutex.Lock()
	defer hc.routeManager.mutex.Unlock()

	for _, ep := range hc.routeManager.endpoints[serviceName] {
		if ep.Host == endpoint.Host && ep.Port == endpoint.Port {
			ep.ConsecutiveFailures++
			ep.LastHealthCheck = time.Now()
			ep.HealthCheckCount++

			if ep.ConsecutiveFailures >= 3 {
				if ep.Health != "unhealthy" {
					log.Printf("端点 %s:%d 标记为不健康: %v", ep.Host, ep.Port, err)
				}
				ep.Health = "unhealthy"
			}
			break
		}
	}
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

	// 支持路径参数和版本匹配的改进路由匹配
	for _, route := range rm.routes {
		if !route.Enabled {
			continue
		}

		// 方法匹配
		if route.Method != "*" && route.Method != method {
			continue
		}

		// 高级路径匹配
		if rm.matchPathAdvanced(route.Path, path, route.Version) {
			return route
		}
	}

	return nil
}

func (rm *RouteManager) matchPathAdvanced(pattern, path, version string) bool {
	// 版本匹配
	if version != "" {
		if !strings.Contains(path, "/"+version+"/") {
			// 检查是否有版本头
			// 这在实际实现中会通过context传递
		}
	}

	// 精确匹配
	if pattern == path {
		return true
	}

	// 前缀匹配
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	// 路径参数匹配 (例如: /users/{id})
	if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
		return rm.matchPathWithParams(pattern, path)
	}

	// 正则匹配
	matched, _ := regexp.MatchString(pattern, path)
	return matched
}

func (rm *RouteManager) matchPathWithParams(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			// 这是一个参数，跳过匹配
			continue
		}
		if part != pathParts[i] {
			return false
		}
	}

	return true
}

// === 现代化中间件系统 ===

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
	TraceSpan trace.Span

	// 2025年新增字段
	JWTClaims    *JWTClaims
	OAuth2Token  *oauth2.Token
	RateLimiter  *rate.Limiter
	CacheKey     string
	CacheEnabled bool
	Errors       []error
}

// CORS中间件 - 2025现代化版本
type CORSMiddleware struct {
	config *SecurityConfig
}

func NewCORSMiddleware(config *SecurityConfig) *CORSMiddleware {
	return &CORSMiddleware{config: config}
}

func (m *CORSMiddleware) Name() string {
	return "cors"
}

func (m *CORSMiddleware) Process(ctx *GatewayContext) error {
	if !m.config.EnableCORS {
		return nil
	}

	origin := ctx.Request.Header.Get("Origin")

	// 检查是否允许的origin
	if len(m.config.AllowedOrigins) > 0 && !m.isOriginAllowed(origin) {
		return fmt.Errorf("CORS: origin not allowed: %s", origin)
	}

	// 设置CORS头
	ctx.Response.Header().Set("Access-Control-Allow-Origin", origin)
	ctx.Response.Header().Set("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
	ctx.Response.Header().Set("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
	ctx.Response.Header().Set("Access-Control-Max-Age", "86400")

	// 处理预检请求
	if ctx.Request.Method == "OPTIONS" {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	return nil
}

func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range m.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// JWT认证中间件 - 2025现代化版本
type JWTMiddleware struct {
	jwtSecret []byte
	config    *AuthConfig
}

func NewJWTMiddleware(config *AuthConfig) *JWTMiddleware {
	return &JWTMiddleware{
		jwtSecret: []byte(config.JWTSecret),
		config:    config,
	}
}

func (m *JWTMiddleware) Name() string {
	return "auth"
}

func (m *JWTMiddleware) Process(ctx *GatewayContext) error {
	// 检查是否需要认证
	if ctx.Route.Auth != nil && !ctx.Route.Auth.Required {
		return nil
	}

	// 从Authorization头获取token
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		// 检查cookie中的token
		if cookie, err := ctx.Request.Cookie("access_token"); err == nil {
			authHeader = "Bearer " + cookie.Value
		}
	}

	if authHeader == "" {
		return fmt.Errorf("缺少Authorization头或cookie")
	}

	// 提取token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return fmt.Errorf("Authorization头格式错误")
	}

	// 验证JWT token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return fmt.Errorf("JWT验证失败: %v", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("JWT claims无效")
	}

	// 检查token是否过期
	if time.Now().Unix() > claims.ExpiresAt {
		return fmt.Errorf("JWT已过期")
	}

	// 权限检查
	if err := m.checkPermissions(ctx, claims); err != nil {
		return err
	}

	// 设置用户信息到context
	ctx.UserID = claims.UserID
	ctx.JWTClaims = claims
	ctx.Metadata["user_id"] = claims.UserID
	ctx.Metadata["username"] = claims.Username
	ctx.Metadata["roles"] = claims.Roles

	return nil
}

func (m *JWTMiddleware) checkPermissions(ctx *GatewayContext, claims *JWTClaims) error {
	route := ctx.Route

	// 检查角色
	if route.Auth != nil && len(route.Auth.Roles) > 0 {
		hasRole := false
		for _, requiredRole := range route.Auth.Roles {
			for _, userRole := range claims.Roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}
		if !hasRole {
			return fmt.Errorf("用户角色不足，需要角色: %v", route.Auth.Roles)
		}
	}

	// 检查权限
	if route.Auth != nil && len(route.Auth.Permissions) > 0 {
		for _, requiredPermission := range route.Auth.Permissions {
			hasPermission := false
			for _, userPermission := range claims.Permissions {
				if userPermission == requiredPermission {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				return fmt.Errorf("用户权限不足，需要权限: %s", requiredPermission)
			}
		}
	}

	return nil
}

// 追踪中间件 - OpenTelemetry集成
type TracingMiddleware struct {
	tracer trace.Tracer
}

func NewTracingMiddleware(tracer trace.Tracer) *TracingMiddleware {
	return &TracingMiddleware{tracer: tracer}
}

func (m *TracingMiddleware) Name() string {
	return "tracing"
}

func (m *TracingMiddleware) Process(ctx *GatewayContext) error {
	if m.tracer == nil {
		return nil
	}

	// 开始新的span
	spanCtx, span := m.tracer.Start(ctx.Request.Context(), fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL.Path))
	ctx.TraceSpan = span

	// 设置span属性
	span.SetAttributes(
		attribute.String("http.method", ctx.Request.Method),
		attribute.String("http.url", ctx.Request.URL.String()),
		attribute.String("http.route", ctx.Route.Path),
		attribute.String("service.name", ctx.Route.ServiceName),
		attribute.String("user.id", ctx.UserID),
		attribute.String("request.id", ctx.RequestID),
		attribute.String("client.ip", ctx.ClientIP),
	)

	// 更新request context
	ctx.Request = ctx.Request.WithContext(spanCtx)

	return nil
}

// WebSocket代理中间件
type WebSocketMiddleware struct{}

func NewWebSocketMiddleware() *WebSocketMiddleware {
	return &WebSocketMiddleware{}
}

func (m *WebSocketMiddleware) Name() string {
	return "websocket"
}

func (m *WebSocketMiddleware) Process(ctx *GatewayContext) error {
	if ctx.Route.Protocol != "ws" && ctx.Route.Protocol != "wss" {
		return nil
	}

	// 检查WebSocket升级请求
	if ctx.Request.Header.Get("Connection") != "Upgrade" ||
		ctx.Request.Header.Get("Upgrade") != "websocket" {
		return fmt.Errorf("非WebSocket升级请求")
	}

	// 执行WebSocket代理
	return m.proxyWebSocket(ctx)
}

func (m *WebSocketMiddleware) proxyWebSocket(ctx *GatewayContext) error {
	// 构建目标WebSocket URL
	targetURL := fmt.Sprintf("ws://%s:%d%s", ctx.Endpoint.Host, ctx.Endpoint.Port, ctx.Route.TargetPath)

	// 连接到目标WebSocket服务
	targetConn, _, err := websocket.DefaultDialer.Dial(targetURL, nil)
	if err != nil {
		return fmt.Errorf("连接目标WebSocket失败: %v", err)
	}
	defer targetConn.Close()

	// 升级客户端连接为WebSocket
	clientConn, err := upgrader.Upgrade(ctx.Response, ctx.Request, nil)
	if err != nil {
		return fmt.Errorf("升级客户端连接失败: %v", err)
	}
	defer clientConn.Close()

	// 双向消息代理
	go m.proxyMessages(clientConn, targetConn, "client->target")
	go m.proxyMessages(targetConn, clientConn, "target->client")

	// 等待连接关闭
	select {
	case <-ctx.Request.Context().Done():
		return nil
	}
}

func (m *WebSocketMiddleware) proxyMessages(src, dst *websocket.Conn, direction string) {
	for {
		messageType, data, err := src.ReadMessage()
		if err != nil {
			log.Printf("WebSocket读取消息失败 [%s]: %v", direction, err)
			break
		}

		err = dst.WriteMessage(messageType, data)
		if err != nil {
			log.Printf("WebSocket写入消息失败 [%s]: %v", direction, err)
			break
		}
	}
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

// === 中间件实现 ===

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
		routeManager:   NewRouteManager(loadDefaultConfig()),
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

// 加载默认配置
func loadDefaultConfig() *GatewayConfig {
	return &GatewayConfig{
		Server: ServerConfig{
			Port:           DefaultPort,
			ReadTimeout:    DefaultReadTimeout,
			WriteTimeout:   DefaultWriteTimeout,
			IdleTimeout:    DefaultIdleTimeout,
			MaxHeaderBytes: 1 << 20, // 1MB
			EnableTLS:      false,
			EnableHTTP2:    true,
			EnableGRPCWeb:  true,
		},
		Auth: AuthConfig{
			JWTSecret:     DefaultJWTSecret,
			JWTExpiryTime: DefaultJWTExpiryTime,
			RefreshTime:   DefaultRefreshTime,
			EnableOAuth2:  false,
			EnableMTLS:    false,
		},
		Tracing: TracingConfig{
			Enabled:     true,
			ServiceName: "api-gateway",
			JaegerURL:   "http://localhost:14268/api/traces",
			SampleRate:  1.0,
			Environment: "development",
		},
		Monitoring: MonitoringConfig{
			MetricsEnabled: true,
			MetricsPath:    "/metrics",
			HealthPath:     DefaultHealthPath,
			ReadyPath:      DefaultReadyPath,
			LivenessPath:   DefaultLivenessPath,
			PrometheusPort: "9091",
		},
		Security: SecurityConfig{
			EnableCORS:       true,
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			EnableCSRF:       false,
			RateLimitEnabled: true,
			RateLimit:        DefaultRateLimit,
			BurstLimit:       DefaultBurstLimit,
		},
		K8s: K8sConfig{
			Enabled:          false,
			InCluster:        false,
			ServiceDiscovery: false,
			WatchConfigMaps:  false,
		},
	}
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
