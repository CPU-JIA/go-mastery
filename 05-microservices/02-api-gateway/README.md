# 🚀 现代化API网关 - 2025年企业级实现

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.24)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](LICENSE)
[![Docker](https://img.shields.io/badge/docker-ready-blue?style=for-the-badge&logo=docker)](Dockerfile)
[![Kubernetes](https://img.shields.io/badge/kubernetes-native-326ce5?style=for-the-badge&logo=kubernetes)](k8s/)

## 📋 概述

这是一个基于Go 1.24+实现的现代化API网关，集成了2025年的最佳实践和云原生特性。该网关专为高并发、高可用的微服务架构而设计。

## ✨ 核心特性

### 🔐 高级认证授权
- **JWT令牌验证** - 支持HS256/RS256算法
- **OAuth2.0集成** - Google、GitHub、自定义提供商
- **mTLS双向认证** - 证书验证和客户端认证
- **RBAC权限控制** - 基于角色和权限的访问控制
- **API Key管理** - 多层级密钥管理系统

### 🌐 智能路由管理
- **动态路由配置** - 支持热更新，零停机配置变更
- **高级路径匹配** - 正则表达式、路径参数、通配符
- **API版本管理** - 多版本并行支持，A/B测试
- **金丝雀部署** - 流量渐进式切换
- **GraphQL网关** - 查询聚合和数据合成

### ⚡ 高性能特性
- **自适应负载均衡** - 加权轮询、最少连接、IP哈希
- **连接池优化** - Keep-Alive连接复用
- **响应缓存** - Redis集成，智能缓存策略
- **WebSocket代理** - 实时通信支持
- **gRPC协议转换** - HTTP到gRPC的无缝转换

### 🔍 可观测性集成
- **OpenTelemetry追踪** - 分布式链路追踪
- **Prometheus指标** - 详细的性能监控
- **结构化日志** - JSON格式，可搜索日志
- **健康检查** - Kubernetes风格的探针
- **实时监控面板** - Grafana仪表板

### ☁️ 云原生支持
- **Kubernetes原生** - 自动服务发现和配置热更新
- **Istio Service Mesh** - 流量管理和安全策略
- **容器优化** - 多阶段构建，最小化镜像
- **优雅停机** - 零丢失请求的关闭机制
- **水平扩展** - HPA自动扩缩容支持

### 🛡️ 安全防护
- **速率限制** - 多维度限流：IP、用户、API
- **熔断保护** - 自动故障检测和恢复
- **安全头注入** - OWASP推荐的安全头
- **CORS处理** - 跨域资源共享控制
- **DDoS防护** - 分布式拒绝服务攻击防护

## 🚀 快速开始

### 1. 环境要求

- Go 1.24+
- Docker 20.10+
- Kubernetes 1.24+ (可选)
- Redis (可选，用于缓存)

### 2. 安装依赖

```bash
# 克隆仓库
git clone https://github.com/your-org/modern-api-gateway.git
cd modern-api-gateway

# 安装Go依赖
go mod tidy

# 启动依赖服务 (可选)
docker-compose up -d redis jaeger prometheus
```

### 3. 配置

编辑 `config.yaml` 文件，配置你的环境：

```yaml
server:
  port: "8080"
  enable_tls: false

auth:
  jwt_secret: "your-super-secret-key"

tracing:
  enabled: true
  jaeger_url: "http://localhost:14268/api/traces"

monitoring:
  metrics_enabled: true
```

### 4. 运行网关

```bash
# 开发环境
go run main.go

# 生产环境
go build -o gateway main.go
./gateway
```

### 5. 验证部署

```bash
# 健康检查
curl http://localhost:8080/health

# 查看路由配置
curl http://localhost:8080/admin/routes

# 查看监控指标
curl http://localhost:8080/metrics
```

## 🔧 配置指南

### JWT认证配置

```yaml
auth:
  jwt_secret: "your-256-bit-secret"
  jwt_expiry_time: 24h
  enable_oauth2: true
  oauth2:
    client_id: "your-oauth2-client-id"
    client_secret: "your-oauth2-client-secret"
```

### 路由配置

```json
{
  "id": "user-api-v1",
  "path": "/api/v1/users",
  "method": "*",
  "service_name": "user-service",
  "target_path": "/users",
  "auth": {
    "required": true,
    "roles": ["user", "admin"]
  },
  "rate_limit": {
    "enabled": true,
    "rps": 100,
    "burst": 200
  }
}
```

### Kubernetes部署

```bash
# 应用Kubernetes配置
kubectl apply -f k8s/

# 检查部署状态
kubectl get pods -l app=api-gateway

# 查看服务
kubectl get svc api-gateway
```

## 📊 监控和可观测性

### Prometheus指标

- `http_requests_total` - HTTP请求总数
- `http_request_duration_seconds` - 请求响应时间
- `gateway_route_hits` - 路由命中统计
- `gateway_auth_failures` - 认证失败统计
- `gateway_circuit_breaker_state` - 熔断器状态

### 分布式追踪

网关集成了OpenTelemetry，支持：
- Jaeger追踪
- Zipkin追踪
- 自定义采样策略
- 跨服务链路追踪

### 日志聚合

结构化JSON日志输出，包含：
- 请求ID追踪
- 用户身份信息
- 响应时间统计
- 错误堆栈信息

## 🔄 中间件系统

### 内置中间件

1. **认证中间件** - JWT/OAuth2验证
2. **授权中间件** - RBAC权限检查
3. **限流中间件** - 多维度速率限制
4. **CORS中间件** - 跨域请求处理
5. **追踪中间件** - 分布式追踪
6. **缓存中间件** - 响应缓存
7. **WebSocket中间件** - WebSocket代理

### 自定义中间件

```go
type CustomMiddleware struct {
    config *Config
}

func (m *CustomMiddleware) Name() string {
    return "custom"
}

func (m *CustomMiddleware) Process(ctx *GatewayContext) error {
    // 自定义逻辑
    return nil
}
```

## 🐳 Docker部署

### 构建镜像

```bash
# 多阶段构建
docker build -t api-gateway:latest .

# 运行容器
docker run -d -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  api-gateway:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  api-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_PATH=/app/config.yaml
    volumes:
      - ./config.yaml:/app/config.yaml
    depends_on:
      - redis
      - jaeger
```

## 🔒 安全最佳实践

### 认证安全
- 使用强密码的JWT secret (至少256位)
- 实现token刷新机制
- 启用mTLS双向认证
- 定期轮换API密钥

### 网络安全
- 配置适当的CORS策略
- 启用安全响应头
- 实施IP白名单/黑名单
- 使用TLS 1.3加密

### 运维安全
- 定期更新依赖
- 启用安全扫描
- 监控异常访问
- 实施日志审计

## 📈 性能调优

### 连接池配置
```yaml
http_client:
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  idle_conn_timeout: 90s
  dial_timeout: 30s
```

### 缓存优化
```yaml
cache:
  enabled: true
  ttl: 300s
  max_size: 10000
  compression: true
```

### 负载均衡
```yaml
load_balancer:
  algorithm: "weighted_round_robin"
  health_check_enabled: true
  health_check_interval: 30s
```

## 🎯 最佳实践

### ✅ 推荐做法
- 使用版本化API路由
- 实现完整的健康检查
- 配置合理的超时时间
- 启用分布式追踪
- 实施多层次监控
- 使用配置热更新

### ❌ 避免做法
- 在代码中硬编码密钥
- 忽略错误处理
- 过度复杂的路由规则
- 禁用安全头
- 忽略性能监控
- 单点故障设计

## 🔗 相关资源

- [Go 1.24新特性](https://golang.org/doc/go1.24)
- [OpenTelemetry Go](https://opentelemetry.io/docs/go/)
- [Kubernetes网关API](https://gateway-api.sigs.k8s.io/)
- [Prometheus监控](https://prometheus.io/docs/guides/go-application/)
- [JWT最佳实践](https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/)

## 🤝 贡献指南

欢迎贡献代码和改进建议！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细信息。

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

**🎉 恭喜！** 您已掌握现代化API网关的完整实现方案！