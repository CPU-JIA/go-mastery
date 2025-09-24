package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
)

/*
微服务架构 - 服务发现和注册练习

本练习涵盖微服务架构中的服务发现和注册机制，包括：
1. 服务注册中心（Service Registry）
2. 服务发现（Service Discovery）
3. 健康检查（Health Checks）
4. 负载均衡（Load Balancing）
5. 服务治理（Service Governance）
6. 配置中心集成
7. 动态服务路由
8. 故障转移机制

主要概念：
- 微服务架构模式
- 服务注册和发现
- 心跳机制和健康检查
- 分布式系统一致性
- 服务网格基础
*/

// === 服务模型定义 ===

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID           string            `json:"id"`
	ServiceName  string            `json:"service_name"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Tags         []string          `json:"tags"`
	Metadata     map[string]string `json:"metadata"`
	Health       HealthStatus      `json:"health"`
	RegisterTime time.Time         `json:"register_time"`
	LastSeen     time.Time         `json:"last_seen"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string    `json:"status"` // healthy, unhealthy, unknown
	Message   string    `json:"message"`
	CheckedAt time.Time `json:"checked_at"`
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	ServiceName   string            `json:"service_name"`
	Port          int               `json:"port"`
	HealthPath    string            `json:"health_path"`
	CheckInterval time.Duration     `json:"check_interval"`
	Tags          []string          `json:"tags"`
	Metadata      map[string]string `json:"metadata"`
}

// === 服务注册接口 ===

type ServiceRegistry interface {
	Register(instance *ServiceInstance) error
	Deregister(serviceID string) error
	Discover(serviceName string) ([]*ServiceInstance, error)
	GetService(serviceID string) (*ServiceInstance, error)
	UpdateHealth(serviceID string, status HealthStatus) error
	WatchService(serviceName string) (<-chan []*ServiceInstance, error)
}

// === 内存服务注册中心实现 ===

type MemoryServiceRegistry struct {
	services map[string]*ServiceInstance          // serviceID -> instance
	watchers map[string][]chan []*ServiceInstance // serviceName -> watchers
	mutex    sync.RWMutex
}

func NewMemoryServiceRegistry() *MemoryServiceRegistry {
	registry := &MemoryServiceRegistry{
		services: make(map[string]*ServiceInstance),
		watchers: make(map[string][]chan []*ServiceInstance),
	}

	// 启动健康检查协程
	go registry.healthChecker()

	return registry
}

func (r *MemoryServiceRegistry) Register(instance *ServiceInstance) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance.RegisterTime = time.Now()
	instance.LastSeen = time.Now()
	instance.Health = HealthStatus{
		Status:    "unknown",
		Message:   "刚注册，等待健康检查",
		CheckedAt: time.Now(),
	}

	r.services[instance.ID] = instance

	log.Printf("服务注册成功: %s (%s:%d)", instance.ServiceName, instance.Host, instance.Port)

	// 通知观察者
	r.notifyWatchers(instance.ServiceName)

	return nil
}

func (r *MemoryServiceRegistry) Deregister(serviceID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("服务不存在: %s", serviceID)
	}

	delete(r.services, serviceID)

	log.Printf("服务注销成功: %s", serviceID)

	// 通知观察者
	r.notifyWatchers(instance.ServiceName)

	return nil
}

func (r *MemoryServiceRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var instances []*ServiceInstance
	for _, instance := range r.services {
		if instance.ServiceName == serviceName && instance.Health.Status == "healthy" {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}

func (r *MemoryServiceRegistry) GetService(serviceID string) (*ServiceInstance, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.services[serviceID]
	if !exists {
		return nil, fmt.Errorf("服务不存在: %s", serviceID)
	}

	return instance, nil
}

func (r *MemoryServiceRegistry) UpdateHealth(serviceID string, status HealthStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.services[serviceID]
	if !exists {
		return fmt.Errorf("服务不存在: %s", serviceID)
	}

	oldStatus := instance.Health.Status
	instance.Health = status
	instance.LastSeen = time.Now()

	// 如果状态发生变化，通知观察者
	if oldStatus != status.Status {
		log.Printf("服务健康状态变更: %s %s -> %s", serviceID, oldStatus, status.Status)
		r.notifyWatchers(instance.ServiceName)
	}

	return nil
}

func (r *MemoryServiceRegistry) WatchService(serviceName string) (<-chan []*ServiceInstance, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	watcher := make(chan []*ServiceInstance, 10)
	r.watchers[serviceName] = append(r.watchers[serviceName], watcher)

	// 发送当前状态
	go func() {
		instances, _ := r.Discover(serviceName)
		select {
		case watcher <- instances:
		default:
		}
	}()

	return watcher, nil
}

func (r *MemoryServiceRegistry) notifyWatchers(serviceName string) {
	watchers, exists := r.watchers[serviceName]
	if !exists {
		return
	}

	instances, _ := r.Discover(serviceName)
	for _, watcher := range watchers {
		select {
		case watcher <- instances:
		default:
			// 如果channel满了，跳过这个watcher
		}
	}
}

func (r *MemoryServiceRegistry) healthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.mutex.Lock()
		for serviceID, instance := range r.services {
			// 检查服务是否超时
			if time.Since(instance.LastSeen) > time.Minute*2 {
				log.Printf("服务超时，标记为不健康: %s", serviceID)
				instance.Health = HealthStatus{
					Status:    "unhealthy",
					Message:   "服务超时",
					CheckedAt: time.Now(),
				}
			} else {
				// 执行HTTP健康检查
				go r.checkServiceHealth(instance)
			}
		}
		r.mutex.Unlock()
	}
}

func (r *MemoryServiceRegistry) checkServiceHealth(instance *ServiceInstance) {
	healthURL := fmt.Sprintf("http://%s:%d/health", instance.Host, instance.Port)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)

	status := HealthStatus{CheckedAt: time.Now()}

	if err != nil {
		status.Status = "unhealthy"
		status.Message = fmt.Sprintf("健康检查失败: %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			status.Status = "healthy"
			status.Message = "健康检查通过"
		} else {
			status.Status = "unhealthy"
			status.Message = fmt.Sprintf("健康检查返回状态码: %d", resp.StatusCode)
		}
	}

	r.UpdateHealth(instance.ID, status)
}

// === Consul服务注册中心实现 ===

type ConsulServiceRegistry struct {
	client *api.Client
	config *api.Config
}

func NewConsulServiceRegistry(address string) (*ConsulServiceRegistry, error) {
	config := api.DefaultConfig()
	if address != "" {
		config.Address = address
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &ConsulServiceRegistry{
		client: client,
		config: config,
	}, nil
}

func (c *ConsulServiceRegistry) Register(instance *ServiceInstance) error {
	service := &api.AgentServiceRegistration{
		ID:      instance.ID,
		Name:    instance.ServiceName,
		Tags:    instance.Tags,
		Port:    instance.Port,
		Address: instance.Host,
		Meta:    instance.Metadata,
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d/health", instance.Host, instance.Port),
			Interval:                       "30s",
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: "2m",
		},
	}

	return c.client.Agent().ServiceRegister(service)
}

func (c *ConsulServiceRegistry) Deregister(serviceID string) error {
	return c.client.Agent().ServiceDeregister(serviceID)
}

func (c *ConsulServiceRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	services, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	var instances []*ServiceInstance
	for _, service := range services {
		instance := &ServiceInstance{
			ID:          service.Service.ID,
			ServiceName: service.Service.Service,
			Host:        service.Service.Address,
			Port:        service.Service.Port,
			Tags:        service.Service.Tags,
			Metadata:    service.Service.Meta,
			Health: HealthStatus{
				Status:    "healthy", // Consul已过滤健康的服务
				CheckedAt: time.Now(),
			},
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

func (c *ConsulServiceRegistry) GetService(serviceID string) (*ServiceInstance, error) {
	service, _, err := c.client.Agent().Service(serviceID, nil)
	if err != nil {
		return nil, err
	}

	if service == nil {
		return nil, fmt.Errorf("服务不存在: %s", serviceID)
	}

	return &ServiceInstance{
		ID:          service.ID,
		ServiceName: service.Service,
		Host:        service.Address,
		Port:        service.Port,
		Tags:        service.Tags,
		Metadata:    service.Meta,
	}, nil
}

func (c *ConsulServiceRegistry) UpdateHealth(serviceID string, status HealthStatus) error {
	// Consul使用代理的健康检查，这里可以自定义实现
	return nil
}

func (c *ConsulServiceRegistry) WatchService(serviceName string) (<-chan []*ServiceInstance, error) {
	// 实现Consul watch功能
	watcher := make(chan []*ServiceInstance)
	// 这里应该使用Consul的Watch API
	return watcher, nil
}

// === 服务代理和负载均衡 ===

type LoadBalancer interface {
	Select(instances []*ServiceInstance) *ServiceInstance
}

// 轮询负载均衡
type RoundRobinLoadBalancer struct {
	counter uint64
	mutex   sync.Mutex
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

func (lb *RoundRobinLoadBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	index := lb.counter % uint64(len(instances))
	lb.counter++

	return instances[index]
}

// 随机负载均衡
type RandomLoadBalancer struct{}

func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{}
}

func (lb *RandomLoadBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}

	index := time.Now().UnixNano() % int64(len(instances))
	return instances[index]
}

// === 服务客户端 ===

type ServiceClient struct {
	registry     ServiceRegistry
	loadBalancer LoadBalancer
	httpClient   *http.Client
}

func NewServiceClient(registry ServiceRegistry, loadBalancer LoadBalancer) *ServiceClient {
	return &ServiceClient{
		registry:     registry,
		loadBalancer: loadBalancer,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *ServiceClient) Call(serviceName, method, path string, body interface{}) (*http.Response, error) {
	// 服务发现
	instances, err := c.registry.Discover(serviceName)
	if err != nil {
		return nil, fmt.Errorf("服务发现失败: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("没有可用的服务实例: %s", serviceName)
	}

	// 负载均衡选择实例
	instance := c.loadBalancer.Select(instances)
	if instance == nil {
		return nil, fmt.Errorf("负载均衡器无法选择实例")
	}

	// 构建请求
	url := fmt.Sprintf("http://%s:%d%s", instance.Host, instance.Port, path)

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// 请求失败，可能需要标记实例为不健康
		log.Printf("请求失败，服务实例: %s:%d, 错误: %v", instance.Host, instance.Port, err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	return resp, nil
}

// === 示例微服务 ===

type UserService struct {
	config   ServiceConfig
	registry ServiceRegistry
	server   *http.Server
}

func NewUserService(config ServiceConfig, registry ServiceRegistry) *UserService {
	return &UserService{
		config:   config,
		registry: registry,
	}
}

func (s *UserService) Start() error {
	// 创建路由
	router := mux.NewRouter()

	// 业务端点
	router.HandleFunc("/users", s.getUsers).Methods("GET")
	router.HandleFunc("/users/{id}", s.getUser).Methods("GET")

	// 健康检查端点
	router.HandleFunc("/health", s.healthCheck).Methods("GET")

	// 创建服务器
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: router,
	}

	// 注册服务
	instance := &ServiceInstance{
		ID:          fmt.Sprintf("%s-%d", s.config.ServiceName, s.config.Port),
		ServiceName: s.config.ServiceName,
		Host:        "localhost",
		Port:        s.config.Port,
		Tags:        s.config.Tags,
		Metadata:    s.config.Metadata,
	}

	if err := s.registry.Register(instance); err != nil {
		return fmt.Errorf("注册服务失败: %w", err)
	}

	// 启动心跳
	go s.heartbeat(instance.ID)

	// 启动HTTP服务器
	log.Printf("用户服务启动在端口 %d", s.config.Port)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("启动HTTP服务器失败: %w", err)
	}

	return nil
}

func (s *UserService) Stop() error {
	// 注销服务
	instance := &ServiceInstance{
		ID: fmt.Sprintf("%s-%d", s.config.ServiceName, s.config.Port),
	}
	s.registry.Deregister(instance.ID)

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *UserService) getUsers(w http.ResponseWriter, r *http.Request) {
	users := []map[string]interface{}{
		{"id": 1, "name": "Alice", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "email": "bob@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *UserService) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	user := map[string]interface{}{
		"id":    id,
		"name":  "User " + id,
		"email": "user" + id + "@example.com",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *UserService) healthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   s.config.ServiceName,
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(time.Now()).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *UserService) heartbeat(serviceID string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := HealthStatus{
			Status:    "healthy",
			Message:   "心跳正常",
			CheckedAt: time.Now(),
		}
		s.registry.UpdateHealth(serviceID, status)
	}
}

// === API网关 ===

type APIGateway struct {
	registry ServiceRegistry
	client   *ServiceClient
	server   *http.Server
}

func NewAPIGateway(registry ServiceRegistry) *APIGateway {
	return &APIGateway{
		registry: registry,
		client:   NewServiceClient(registry, NewRoundRobinLoadBalancer()),
	}
}

func (gw *APIGateway) Start(port int) error {
	router := mux.NewRouter()

	// 代理规则
	router.PathPrefix("/api/users").HandlerFunc(gw.proxyToUserService)
	router.HandleFunc("/health", gw.healthCheck).Methods("GET")

	// 服务发现端点
	router.HandleFunc("/services", gw.listServices).Methods("GET")
	router.HandleFunc("/services/{name}", gw.getServiceInstances).Methods("GET")

	gw.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.Printf("API网关启动在端口 %d", port)
	if err := gw.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("启动API网关失败: %w", err)
	}

	return nil
}

func (gw *APIGateway) proxyToUserService(w http.ResponseWriter, r *http.Request) {
	// 移除/api前缀
	path := strings.TrimPrefix(r.URL.Path, "/api")

	resp, err := gw.client.Call("user-service", r.Method, path, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	io.Copy(w, resp.Body)
}

func (gw *APIGateway) listServices(w http.ResponseWriter, r *http.Request) {
	// 这里应该从注册中心获取所有服务
	services := []string{"user-service", "order-service", "payment-service"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func (gw *APIGateway) getServiceInstances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	instances, err := gw.registry.Discover(serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

func (gw *APIGateway) healthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "api-gateway",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// === 示例：微服务架构最佳实践 ===

func demonstrateMicroservicesBestPractices() {
	fmt.Println("=== 微服务架构最佳实践 ===")

	fmt.Println("1. 服务设计原则:")
	fmt.Println("   ✓ 单一职责原则")
	fmt.Println("   ✓ 数据隔离")
	fmt.Println("   ✓ API优先设计")
	fmt.Println("   ✓ 无状态服务")

	fmt.Println("2. 服务发现:")
	fmt.Println("   ✓ 客户端发现 vs 服务端发现")
	fmt.Println("   ✓ 健康检查和自动故障转移")
	fmt.Println("   ✓ 服务注册和注销")
	fmt.Println("   ✓ 负载均衡策略")

	fmt.Println("3. 通信模式:")
	fmt.Println("   ✓ 同步通信（HTTP/REST）")
	fmt.Println("   ✓ 异步通信（消息队列）")
	fmt.Println("   ✓ 事件驱动架构")
	fmt.Println("   ✓ 服务网格")

	fmt.Println("4. 数据管理:")
	fmt.Println("   ✓ 数据库per服务")
	fmt.Println("   ✓ 分布式事务处理")
	fmt.Println("   ✓ 事件溯源")
	fmt.Println("   ✓ CQRS模式")

	fmt.Println("5. 监控和运维:")
	fmt.Println("   ✓ 分布式追踪")
	fmt.Println("   ✓ 集中式日志")
	fmt.Println("   ✓ 指标监控")
	fmt.Println("   ✓ 健康检查")
}

func main() {
	// 创建服务注册中心
	registry := NewMemoryServiceRegistry()

	// 演示微服务架构最佳实践
	demonstrateMicroservicesBestPractices()

	// 创建用户服务配置
	userServiceConfig := ServiceConfig{
		ServiceName:   "user-service",
		Port:          8081,
		HealthPath:    "/health",
		CheckInterval: 30 * time.Second,
		Tags:          []string{"user", "v1.0"},
		Metadata: map[string]string{
			"version": "1.0.0",
			"region":  "us-west-1",
		},
	}

	// 创建API网关
	gateway := NewAPIGateway(registry)

	// 启动服务
	go func() {
		userService := NewUserService(userServiceConfig, registry)
		if err := userService.Start(); err != nil {
			log.Printf("用户服务启动失败: %v", err)
		}
	}()

	// 启动API网关
	go func() {
		if err := gateway.Start(8080); err != nil {
			log.Printf("API网关启动失败: %v", err)
		}
	}()

	// 演示服务发现和调用
	go func() {
		time.Sleep(2 * time.Second) // 等待服务启动

		client := NewServiceClient(registry, NewRoundRobinLoadBalancer())

		for i := 0; i < 5; i++ {
			resp, err := client.Call("user-service", "GET", "/users", nil)
			if err != nil {
				log.Printf("调用用户服务失败: %v", err)
			} else {
				log.Printf("调用用户服务成功: %d", resp.StatusCode)
				resp.Body.Close()
			}
			time.Sleep(2 * time.Second)
		}
	}()

	fmt.Println("\n=== 微服务系统启动 ===")
	fmt.Println("服务端点:")
	fmt.Println("  API网关:    http://localhost:8080")
	fmt.Println("  用户服务:   http://localhost:8081")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  GET  /api/users        - 获取用户列表（通过网关）")
	fmt.Println("  GET  /api/users/{id}   - 获取用户详情（通过网关）")
	fmt.Println("  GET  /services         - 获取服务列表")
	fmt.Println("  GET  /services/{name}  - 获取服务实例")
	fmt.Println("  GET  /health           - 健康检查")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  curl http://localhost:8080/api/users")
	fmt.Println("  curl http://localhost:8080/services")
	fmt.Println("  curl http://localhost:8080/services/user-service")
	fmt.Println()
	fmt.Println("按 Ctrl+C 退出...")

	// 优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("收到关闭信号，正在优雅关闭...")

	// 这里应该关闭所有服务
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if gateway.server != nil {
		gateway.server.Shutdown(ctx)
	}

	log.Println("系统已关闭")
}

/*
练习任务：

1. 基础练习：
   - 实现更多负载均衡算法（加权轮询、一致性哈希）
   - 添加服务健康检查策略
   - 实现服务版本管理
   - 添加服务标签过滤

2. 中级练习：
   - 集成Consul或Etcd作为服务注册中心
   - 实现服务配置动态更新
   - 添加服务熔断机制
   - 实现请求路由策略

3. 高级练习：
   - 实现服务网格集成
   - 添加分布式追踪
   - 实现多数据中心支持
   - 集成Kubernetes服务发现

4. 监控和运维：
   - 实现服务监控指标
   - 添加服务依赖图
   - 实现自动扩缩容
   - 添加故障自愈机制

5. 安全和治理：
   - 实现服务间认证
   - 添加API限流和配额
   - 实现服务访问控制
   - 添加API版本管理

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/hashicorp/consul/api

2. 可选：启动Consul
   consul agent -dev

3. 运行程序：go run main.go

架构组件：
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   客户端     │────│  API网关     │────│ 服务注册中心 │
└─────────────┘    └─────────────┘    └─────────────┘
                         │                    │
                    ┌────────────┐            │
                    │ 负载均衡器  │            │
                    └────────────┘            │
                         │                    │
        ┌────────────────┼────────────────────┼──────┐
        │                │                    │      │
   ┌─────────┐    ┌─────────┐    ┌─────────┐  │      │
   │ 用户服务 │    │ 订单服务 │    │ 支付服务 │  │      │
   │ (8081)  │    │ (8082)  │    │ (8083)  │  │      │
   └─────────┘    └─────────┘    └─────────┘  │      │
        │                │                │   │      │
        └────────────────┼────────────────┘   │      │
                         │                    │      │
                    ┌────────────┐            │      │
                    │ 健康检查器  │────────────┘      │
                    └────────────┘                   │
                         │                           │
                    ┌────────────┐                   │
                    │ 服务监控器  │───────────────────┘
                    └────────────┘

扩展建议：
- 实现服务治理平台
- 集成APM工具（如Jaeger、Zipkin）
- 添加服务依赖分析
- 实现蓝绿部署支持
*/
