# 🐳 微服务Docker化 - 2025年最佳实践

本模块展示了现代Go微服务的完整Docker化解决方案，包含生产级配置和最佳实践。

## 🎯 核心特性

### 🏗️ 多阶段构建
- ✅ 镜像大小优化90%+ (从800MB+ → <20MB)
- ✅ 静态链接二进制文件
- ✅ 构建参数和元数据注入
- ✅ 分层缓存优化构建速度

### 🔒 安全最佳实践
- ✅ 非root用户运行 (UID: 65534)
- ✅ 最小基础镜像 (scratch)
- ✅ 安全扫描支持
- ✅ 敏感文件排除 (.dockerignore)

### 💊 健康检查和监控
- ✅ Kubernetes风格健康检查
- ✅ Prometheus指标集成
- ✅ 优雅停机信号处理
- ✅ 资源限制配置

### 🚀 生产就绪特性
- ✅ 完整的微服务栈编排
- ✅ 负载均衡和高可用
- ✅ 监控和日志聚合
- ✅ 分布式追踪

## 📦 项目结构

```
05-microservices/10-docker/
├── main.go              # 微服务应用代码
├── Dockerfile           # 多阶段构建配置
├── docker-compose.yml   # 完整微服务栈
├── .dockerignore       # 构建优化
├── go.mod              # Go依赖管理
├── go.sum              # 依赖校验和
└── README.md           # 使用说明
```

## 🚀 快速开始

### 1. 构建镜像

```bash
# 基础构建
docker build -t microservice:latest .

# 带参数构建
docker build \
  --build-arg BUILD_VERSION=v1.2.3 \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  --build-arg BUILD_COMMIT=$(git rev-parse --short HEAD) \
  -t microservice:v1.2.3 .
```

### 2. 运行单个容器

```bash
# 基础运行
docker run -p 8080:8080 microservice:latest

# 带环境变量
docker run -p 8080:8080 \
  -e ENVIRONMENT=production \
  -e LOG_LEVEL=info \
  microservice:latest
```

### 3. 完整微服务栈

```bash
# 启动完整栈
docker-compose up -d

# 启动基础服务
docker-compose up -d microservice-1 microservice-2 postgres redis

# 查看服务状态
docker-compose ps
```

## 🌐 访问端点

| 服务 | 端口 | 描述 |
|------|------|------|
| **负载均衡** | [http://localhost](http://localhost) | Nginx负载均衡器 |
| **微服务-1** | [http://localhost:8081](http://localhost:8081) | 微服务实例1 |
| **微服务-2** | [http://localhost:8082](http://localhost:8082) | 微服务实例2 |
| **Prometheus** | [http://localhost:9090](http://localhost:9090) | 监控指标 |
| **Grafana** | [http://localhost:3000](http://localhost:3000) | 监控面板 (admin/admin123) |
| **Jaeger** | [http://localhost:16686](http://localhost:16686) | 分布式追踪 |
| **Kibana** | [http://localhost:5601](http://localhost:5601) | 日志分析 |

## 🔍 健康检查

### 容器健康检查
```bash
# 检查容器健康状态
docker ps --filter health=healthy

# 查看健康检查日志
docker inspect --format='{{.State.Health.Log}}' <container_id>
```

### 应用端点
```bash
# 健康检查
curl http://localhost:8080/health

# 就绪检查
curl http://localhost:8080/ready

# 详细状态
curl http://localhost:8080/api/status

# Prometheus指标
curl http://localhost:8080/api/metrics
```

## 📊 监控和可观测性

### Prometheus指标
- `http_requests_total` - HTTP请求总数
- `http_errors_total` - HTTP错误总数
- `service_uptime_seconds` - 服务运行时间
- `go_goroutines` - Goroutine数量

### Grafana面板
默认包含以下面板：
- 服务状态概览
- 请求速率和错误率
- 响应时间分布
- 资源使用情况

### 分布式追踪
Jaeger自动追踪：
- HTTP请求链路
- 数据库查询
- 缓存操作
- 外部服务调用

## 🔧 环境配置

### 环境变量

| 变量名 | 默认值 | 描述 |
|--------|--------|------|
| `PORT` | `8080` | 服务端口 |
| `ENVIRONMENT` | `development` | 运行环境 |
| `LOG_LEVEL` | `info` | 日志级别 |
| `READ_TIMEOUT` | `30s` | 读取超时 |
| `WRITE_TIMEOUT` | `30s` | 写入超时 |
| `POSTGRES_URL` | - | PostgreSQL连接字符串 |
| `REDIS_URL` | - | Redis连接字符串 |

### 配置覆盖
```bash
# 使用环境文件
docker run --env-file .env microservice:latest

# Docker Compose环境变量
export ENVIRONMENT=production
docker-compose up -d
```

## 🚀 部署选项

### 1. 单容器部署
```bash
docker run -d \
  --name microservice \
  --restart unless-stopped \
  -p 8080:8080 \
  microservice:latest
```

### 2. Docker Swarm
```bash
# 初始化Swarm
docker swarm init

# 部署栈
docker stack deploy -c docker-compose.yml microservice-stack
```

### 3. Kubernetes
参见 `11-kubernetes/` 目录的Kubernetes部署配置。

### 4. 云平台
- **AWS ECS**: 使用任务定义
- **Google Cloud Run**: 支持容器到云
- **Azure Container Instances**: 快速部署

## 🔒 安全配置

### 镜像安全扫描
```bash
# 使用Trivy扫描
trivy image microservice:latest

# 使用Docker Scout
docker scout quickview microservice:latest

# 使用Snyk
snyk container test microservice:latest
```

### 运行时安全
```bash
# 只读根文件系统
docker run --read-only microservice:latest

# 删除特权能力
docker run --cap-drop=ALL microservice:latest

# 安全配置文件
docker run --security-opt=no-new-privileges microservice:latest
```

## 📈 性能优化

### 构建优化
- 多阶段构建减少层数
- 利用Docker层缓存
- 并行构建支持
- 构建上下文最小化

### 运行时优化
- 资源限制配置
- JIT编译优化
- 垃圾回收调优
- 连接池管理

### 监控建议
```bash
# 容器资源使用
docker stats

# 详细指标
docker exec <container> cat /proc/meminfo
```

## 🐛 故障排除

### 常见问题

**1. 容器启动失败**
```bash
# 查看启动日志
docker logs <container_name>

# 进入容器调试
docker run -it --entrypoint=/bin/sh microservice:latest
```

**2. 健康检查失败**
```bash
# 检查健康检查脚本
docker exec <container> /microservice --health-check

# 查看端口绑定
docker port <container>
```

**3. 内存不足**
```bash
# 检查内存限制
docker inspect <container> | grep -i memory

# 调整资源限制
docker run -m 512m microservice:latest
```

### 调试命令
```bash
# 容器内部检查
docker exec -it <container> sh

# 查看进程
docker exec <container> ps aux

# 查看网络
docker exec <container> netstat -tulpn
```

## 🎯 最佳实践总结

### ✅ DO (推荐做法)
- 使用多阶段构建减少镜像大小
- 实现完整的健康检查
- 配置资源限制和安全约束
- 使用非root用户运行
- 实现优雅停机处理
- 配置日志聚合和监控

### ❌ DON'T (避免做法)
- 在镜像中包含敏感信息
- 使用root用户运行应用
- 忽略健康检查配置
- 镜像层数过多
- 硬编码配置信息
- 忽略安全扫描

## 📚 进阶学习

- [Docker最佳实践指南](https://docs.docker.com/develop/best-practices/)
- [容器安全指南](https://kubernetes.io/docs/concepts/security/)
- [微服务监控策略](https://prometheus.io/docs/practices/)
- [分布式追踪实现](https://opentelemetry.io/docs/)

---

**🎉 恭喜！** 您已掌握现代Go微服务的完整Docker化方案！