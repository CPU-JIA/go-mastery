# 🚀 Go Mastery 快速参考手册

## ⚡ 快速启动

### 🐳 Docker方式 (推荐)
```bash
# 开发环境 (热重载)
docker-compose up go-mastery-dev

# 访问应用
http://localhost:8080

# 查看日志
docker-compose logs -f go-mastery-dev
```

### 🔧 本地开发
```bash
# 环境设置
make setup

# 运行程序
cd 01-basics/01-hello
go run main.go

# 运行测试
make test
```

## 🛠️ 开发命令速查

### 构建和测试
```bash
make build              # 构建应用
make test               # 运行测试
make test-race          # 竞态检测测试
make coverage           # 覆盖率报告
make bench              # 基准测试
```

### 代码质量
```bash
make fmt                # 格式化代码
make lint               # 代码检查
make vet                # Go vet检查
make security           # 安全扫描
make quality-check      # 完整质量检查
```

### Docker操作
```bash
# 启动服务
docker-compose up go-mastery-dev          # 开发环境
docker-compose up go-mastery-prod         # 生产环境
docker-compose up go-mastery-test         # 测试环境
docker-compose --profile monitoring up    # 监控套件

# 管理服务
docker-compose ps                          # 查看状态
docker-compose logs go-mastery-dev         # 查看日志
docker-compose exec go-mastery-dev sh      # 进入容器
docker-compose restart go-mastery-dev      # 重启服务
docker-compose down                        # 停止所有服务
```

## 📊 服务端口速查

| 服务 | 端口 | 用途 |
|------|------|------|
| 应用服务 | 8080 | 主应用 |
| 调试端口 | 8081 | 调试器 |
| pprof | 6060 | 性能分析 |
| 指标 | 9090 | 应用指标 |
| PostgreSQL | 5432 | 开发数据库 |
| Redis | 6379 | 开发缓存 |
| Prometheus | 9091 | 指标收集 |
| Grafana | 3000 | 监控面板 |
| Jaeger | 16686 | 链路追踪 |

## 🎯 学习路径速查

### 📚 模块概览
```
01-basics          → 基础语法 (1-2周)
02-advanced        → 进阶特性 (2-3周)
03-concurrency     → 并发编程 (3-4周)
04-web             → Web开发 (3-4周)
05-microservices   → 微服务 (4-5周)
06-projects        → 实战项目 (4-6周)
07-runtime-internals → 运行时内核 (3-6个月)
08-performance-mastery → 性能优化 (3-6个月)
09-system-programming → 系统编程 (6-9个月)
10-compiler-toolchain → 编译器工具链 (6-9个月)
11-massive-systems → 大规模系统 (9-15个月)
12-ecosystem-contribution → 生态贡献 (9-15个月)
13-language-design → 语言设计 (15-24个月)
14-tech-leadership → 技术领导力 (15-24个月)
```

### 🎯 学习检查点
```bash
# 基础阶段验收
cd 01-basics && make test      # 基础语法测试
cd 02-advanced && make test    # 进阶特性测试

# 应用阶段验收
cd 03-concurrency && make test # 并发编程测试
cd 04-web && make test         # Web开发测试

# 专家阶段验收
cd 07-runtime-internals && make test  # 运行时测试
cd 08-performance-mastery && make bench # 性能基准测试
```

## 🔍 调试和故障排除

### 应用调试
```bash
# 查看应用日志
docker-compose logs go-mastery-dev

# 进入容器调试
docker-compose exec go-mastery-dev sh

# 性能分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap
```

### 常见问题解决
```bash
# 端口被占用
netstat -tulpn | grep :8080
# 或
lsof -i :8080

# 权限问题 (Linux/macOS)
sudo chown -R $USER:$USER .

# Docker容器无法启动
docker-compose down
docker-compose up --build

# 数据库连接问题
docker-compose exec postgres-dev pg_isready -U dev_user
```

### 清理和重置
```bash
# 清理构建产物
make clean

# 清理Docker环境
docker-compose down -v
docker system prune -a

# 重置数据库
docker volume rm go-mastery_postgres-dev-data
docker-compose up postgres-dev
```

## 📈 监控和观测

### 应用监控
```bash
# 启动监控套件
docker-compose --profile monitoring up

# 访问监控面板
http://localhost:3000  # Grafana (admin/admin123)
http://localhost:9091  # Prometheus
http://localhost:16686 # Jaeger
```

### 性能监控
```bash
# CPU使用率
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存使用
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine状态
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 阻塞分析
go tool pprof http://localhost:6060/debug/pprof/block
```

## 🧪 测试策略

### 测试类型
```bash
# 单元测试
go test ./...

# 集成测试
go test -tags=integration ./...

# 竞态检测
go test -race ./...

# 基准测试
go test -bench=. ./...

# 覆盖率测试
go test -cover ./...
```

### 测试最佳实践
```bash
# 并行测试
go test -parallel 4 ./...

# 详细输出
go test -v ./...

# 测试特定函数
go test -run TestFunctionName

# 基准测试比较
go test -bench=. -benchmem ./...
```

## 🚀 CI/CD工作流

### GitHub Actions触发
```bash
# 推送代码触发CI
git push origin main

# 创建PR触发检查
git checkout -b feature/new-feature
git push origin feature/new-feature
# 创建Pull Request
```

### 本地CI模拟
```bash
# 运行完整CI流程
make ci

# 预提交检查
make pre-commit

# 质量门控检查
make quality-check
```

## 🔧 开发环境定制

### IDE配置 (VS Code)
```json
{
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v", "-race"],
  "go.buildTags": "integration",
  "go.testTimeout": "60s"
}
```

### Git钩子设置
```bash
# 设置pre-commit钩子
make dev-setup

# 手动运行pre-commit
make pre-commit
```

## 📖 重要文档链接

### 项目文档
- [学习指南](LEARNING_GUIDE.md) - 详细学习路径
- [贡献指南](CONTRIBUTING.md) - 开发和贡献指南
- [主README](README.md) - 项目总览

### 配置文件
- [Makefile](Makefile) - 构建自动化
- [Docker Compose](docker-compose.yml) - 环境配置
- [GitHub Actions](.github/workflows/ci-cd.yml) - CI/CD配置
- [Linting配置](.golangci.yml) - 代码质量标准

### 外部资源
- [Go官方文档](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go语言圣经](https://books.studygolang.com/gopl-zh/)

## 🆘 紧急联系

### 问题报告
- 创建GitHub Issue描述问题
- 提供错误日志和环境信息
- 包含重现步骤

### 学习支持
- 查看模块内README文档
- 参考代码注释和示例
- 在Go社区寻求帮助

---

**💡 提示**: 将此页面加入书签，开发过程中随时查阅！