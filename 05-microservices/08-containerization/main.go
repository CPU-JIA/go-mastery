/*
容器化与部署 (Containerization & Deployment)

学习目标:
1. Docker 容器化
2. Kubernetes 部署
3. CI/CD 管道
4. 健康检查
5. 滚动更新
6. 服务网格
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ====================
// 1. 应用配置管理
// ====================

type Config struct {
	AppName     string `json:"app_name"`
	Port        string `json:"port"`
	Environment string `json:"environment"`
	DBHost      string `json:"db_host"`
	DBPort      string `json:"db_port"`
	RedisHost   string `json:"redis_host"`
	RedisPort   string `json:"redis_port"`
	LogLevel    string `json:"log_level"`
	Version     string `json:"version"`
}

func LoadConfig() *Config {
	config := &Config{
		AppName:     getEnv("APP_NAME", "microservice-app"),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		RedisHost:   getEnv("REDIS_HOST", "localhost"),
		RedisPort:   getEnv("REDIS_PORT", "6379"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Version:     getEnv("APP_VERSION", "1.0.0"),
	}
	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ====================
// 2. 健康检查系统
// ====================

type HealthCheck struct {
	Status      string           `json:"status"`
	Version     string           `json:"version"`
	Timestamp   time.Time        `json:"timestamp"`
	Checks      map[string]Check `json:"checks"`
	Uptime      time.Duration    `json:"uptime"`
	Environment string           `json:"environment"`
}

type Check struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Latency time.Duration `json:"latency"`
}

type HealthChecker struct {
	startTime time.Time
	config    *Config
	checks    []CheckFunc
}

type CheckFunc func() Check

func NewHealthChecker(config *Config) *HealthChecker {
	return &HealthChecker{
		startTime: time.Now(),
		config:    config,
		checks:    make([]CheckFunc, 0),
	}
}

func (hc *HealthChecker) AddCheck(name string, checkFunc CheckFunc) {
	hc.checks = append(hc.checks, checkFunc)
}

func (hc *HealthChecker) GetHealth() HealthCheck {
	checks := make(map[string]Check)
	overallStatus := "healthy"

	// 数据库连接检查
	checks["database"] = hc.checkDatabase()
	if checks["database"].Status == "unhealthy" {
		overallStatus = "unhealthy"
	}

	// Redis 连接检查
	checks["redis"] = hc.checkRedis()
	if checks["redis"].Status == "unhealthy" {
		overallStatus = "degraded"
	}

	// 磁盘空间检查
	checks["disk_space"] = hc.checkDiskSpace()
	if checks["disk_space"].Status == "unhealthy" {
		overallStatus = "unhealthy"
	}

	// 内存使用检查
	checks["memory"] = hc.checkMemory()
	if checks["memory"].Status == "warning" && overallStatus == "healthy" {
		overallStatus = "warning"
	}

	return HealthCheck{
		Status:      overallStatus,
		Version:     hc.config.Version,
		Timestamp:   time.Now(),
		Checks:      checks,
		Uptime:      time.Since(hc.startTime),
		Environment: hc.config.Environment,
	}
}

func (hc *HealthChecker) checkDatabase() Check {
	start := time.Now()
	// 模拟数据库连接检查
	time.Sleep(time.Millisecond * 10)

	return Check{
		Status:  "healthy",
		Message: "Database connection successful",
		Latency: time.Since(start),
	}
}

func (hc *HealthChecker) checkRedis() Check {
	start := time.Now()
	// 模拟 Redis 连接检查
	time.Sleep(time.Millisecond * 5)

	return Check{
		Status:  "healthy",
		Message: "Redis connection successful",
		Latency: time.Since(start),
	}
}

func (hc *HealthChecker) checkDiskSpace() Check {
	start := time.Now()
	// 模拟磁盘空间检查
	time.Sleep(time.Millisecond * 2)

	return Check{
		Status:  "healthy",
		Message: "Disk space sufficient (85% used)",
		Latency: time.Since(start),
	}
}

func (hc *HealthChecker) checkMemory() Check {
	start := time.Now()
	// 模拟内存使用检查
	time.Sleep(time.Millisecond * 1)

	return Check{
		Status:  "healthy",
		Message: "Memory usage normal (65% used)",
		Latency: time.Since(start),
	}
}

// ====================
// 3. 监控指标收集
// ====================

type MetricsCollector struct {
	requestCount    int64
	errorCount      int64
	responseTimeSum time.Duration
	startTime       time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		startTime: time.Now(),
	}
}

func (mc *MetricsCollector) RecordRequest(responseTime time.Duration, isError bool) {
	mc.requestCount++
	mc.responseTimeSum += responseTime
	if isError {
		mc.errorCount++
	}
}

func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	uptime := time.Since(mc.startTime)
	avgResponseTime := time.Duration(0)
	if mc.requestCount > 0 {
		avgResponseTime = mc.responseTimeSum / time.Duration(mc.requestCount)
	}

	errorRate := float64(0)
	if mc.requestCount > 0 {
		errorRate = float64(mc.errorCount) / float64(mc.requestCount) * 100
	}

	return map[string]interface{}{
		"uptime_seconds":       uptime.Seconds(),
		"requests_total":       mc.requestCount,
		"errors_total":         mc.errorCount,
		"error_rate_percent":   errorRate,
		"avg_response_time_ms": avgResponseTime.Milliseconds(),
		"requests_per_second":  float64(mc.requestCount) / uptime.Seconds(),
	}
}

// ====================
// 4. 应用服务器
// ====================

type Server struct {
	config           *Config
	healthChecker    *HealthChecker
	metricsCollector *MetricsCollector
	server           *http.Server
}

func NewServer(config *Config) *Server {
	healthChecker := NewHealthChecker(config)
	metricsCollector := NewMetricsCollector()

	s := &Server{
		config:           config,
		healthChecker:    healthChecker,
		metricsCollector: metricsCollector,
	}

	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/api/hello", s.handleHello)
	mux.HandleFunc("/api/version", s.handleVersion)

	// 健康检查路由
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/health/live", s.handleLiveness)
	mux.HandleFunc("/health/ready", s.handleReadiness)

	// 监控路由
	mux.HandleFunc("/metrics", s.handleMetrics)

	s.server = &http.Server{
		Addr:         ":" + config.Port,
		Handler:      s.loggingMiddleware(s.metricsMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("Starting %s on port %s (env: %s, version: %s)",
		s.config.AppName, s.config.Port, s.config.Environment, s.config.Version)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.server.Shutdown(ctx)
}

// ====================
// 5. 中间件
// ====================

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("%s %s %v", r.Method, r.URL.Path, duration)
	})
}

func (s *Server) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 记录响应状态
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		isError := rw.statusCode >= 400

		s.metricsCollector.RecordRequest(duration, isError)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ====================
// 6. 处理器函数
// ====================

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Microservice App</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .card { background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .status-healthy { color: #4CAF50; }
        .status-warning { color: #FF9800; }
        .status-error { color: #F44336; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🐳 Containerized Microservice</h1>
        <div class="card">
            <h2>Application Info</h2>
            <p><strong>Name:</strong> %s</p>
            <p><strong>Version:</strong> %s</p>
            <p><strong>Environment:</strong> %s</p>
            <p><strong>Port:</strong> %s</p>
        </div>
        <div class="card">
            <h2>Available Endpoints</h2>
            <ul>
                <li><a href="/api/hello">GET /api/hello</a> - Hello API</li>
                <li><a href="/api/version">GET /api/version</a> - Version Info</li>
                <li><a href="/health">GET /health</a> - Health Check</li>
                <li><a href="/health/live">GET /health/live</a> - Liveness Probe</li>
                <li><a href="/health/ready">GET /health/ready</a> - Readiness Probe</li>
                <li><a href="/metrics">GET /metrics</a> - Metrics</li>
            </ul>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, html, s.config.AppName, s.config.Version, s.config.Environment, s.config.Port)
}

func (s *Server) handleHello(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":   "Hello from containerized microservice!",
		"timestamp": time.Now().UTC(),
		"version":   s.config.Version,
		"hostname":  getHostname(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"version":     s.config.Version,
		"app_name":    s.config.AppName,
		"environment": s.config.Environment,
		"build_time":  "2024-01-01T00:00:00Z", // 实际应用中应该注入构建时间
		"git_commit":  "abc123def456",         // 实际应用中应该注入 Git 提交哈希
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.healthChecker.GetHealth()

	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health.Status == "degraded" || health.Status == "warning" {
		statusCode = http.StatusOK // 仍然认为服务可用，但有警告
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Liveness probe - 检查应用是否还活着
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Readiness probe - 检查应用是否准备好接收流量
	health := s.healthChecker.GetHealth()

	if health.Status == "healthy" || health.Status == "warning" {
		response := map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		response := map[string]interface{}{
			"status":    "not ready",
			"reason":    "service degraded or unhealthy",
			"timestamp": time.Now().UTC(),
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.metricsCollector.GetMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// ====================
// 7. 辅助函数
// ====================

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// ====================
// 8. 优雅关闭
// ====================

func gracefulShutdown(server *Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %s", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	} else {
		log.Println("Server shutdown complete")
	}
}

// ====================
// 主函数
// ====================

func main() {
	config := LoadConfig()
	server := NewServer(config)

	// 启动优雅关闭处理
	go gracefulShutdown(server)

	// 启动服务器
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server start error: %v", err)
	}
}

/*
=== Dockerfile 示例 ===

# 多阶段构建
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 最终镜像
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# 启动应用
CMD ["./main"]

=== Kubernetes 部署清单 ===

# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: microservice-app
  labels:
    app: microservice-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: microservice-app
  template:
    metadata:
      labels:
        app: microservice-app
    spec:
      containers:
      - name: microservice-app
        image: microservice-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: PORT
          value: "8080"
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        volumeMounts:
        - name: config
          mountPath: /etc/config
      volumes:
      - name: config
        configMap:
          name: app-config

---
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: microservice-app-service
spec:
  selector:
    app: microservice-app
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: ClusterIP

---
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  app.yaml: |
    app_name: "microservice-app"
    environment: "production"
    log_level: "info"

=== CI/CD 管道示例 (GitHub Actions) ===

# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21

    - name: Run tests
      run: go test -v ./...

    - name: Run linting
      uses: golangci/golangci-lint-action@v3

  build-and-deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v3

    - name: Build Docker image
      run: |
        docker build -t microservice-app:${{ github.sha }} .
        docker tag microservice-app:${{ github.sha }} microservice-app:latest

    - name: Push to registry
      run: |
        echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
        docker push microservice-app:${{ github.sha }}
        docker push microservice-app:latest

    - name: Deploy to Kubernetes
      run: |
        # 更新部署配置中的镜像标签
        sed -i 's|microservice-app:latest|microservice-app:${{ github.sha }}|' k8s/deployment.yaml

        # 应用到 Kubernetes 集群
        kubectl apply -f k8s/

=== 练习任务 ===

1. 基础容器化:
   - 创建 Dockerfile
   - 构建镜像
   - 运行容器
   - 配置环境变量

2. Kubernetes 部署:
   - 创建 Deployment
   - 配置 Service
   - 设置 ConfigMap
   - 配置健康检查

3. 监控与日志:
   - 集成 Prometheus 指标
   - 配置日志收集
   - 设置告警规则

4. CI/CD 管道:
   - 自动化测试
   - 镜像构建
   - 自动部署
   - 滚动更新

5. 高级功能:
   - 服务网格 (Istio)
   - 蓝绿部署
   - 金丝雀发布
   - 灾难恢复
*/
