# 📚 Go Mastery 文档导航

欢迎来到Go语言从入门到通天的完整学习路径文档中心！

## 🎯 核心文档

### 📖 主要指南
| 文档 | 描述 | 适用人群 |
|------|------|----------|
| [📋 README.md](README.md) | 项目总览和快速开始 | 所有用户 |
| [🚀 LEARNING_GUIDE.md](LEARNING_GUIDE.md) | 详细学习路径和进度规划 | 学习者 |
| [⚡ QUICK_REFERENCE.md](QUICK_REFERENCE.md) | 命令和工作流速查手册 | 开发者 |
| [🤝 CONTRIBUTING.md](CONTRIBUTING.md) | 开发和贡献指南 | 贡献者 |

### 🔧 配置文件文档
| 文件 | 用途 | 说明 |
|------|------|------|
| [Makefile](Makefile) | 构建自动化 | 包含所有构建、测试、质量检查命令 |
| [docker-compose.yml](docker-compose.yml) | 环境配置 | 多环境Docker配置 |
| [Dockerfile](Dockerfile) | 容器镜像 | 多阶段构建配置 |
| [.golangci.yml](.golangci.yml) | 代码质量 | Linting规则配置 |
| [.github/workflows/ci-cd.yml](.github/workflows/ci-cd.yml) | CI/CD | GitHub Actions配置 |

## 🎯 学习路径文档

### 📚 分阶段学习指南

#### 📊 评估系统
- [00-assessment-system/README.md](00-assessment-system/README.md) - 学习进度评估和测试系统

#### 🌱 基础阶段 (1-4周)
- [01-basics/README.md](01-basics/README.md) - Go语言基础语法
- [02-advanced/README.md](02-advanced/README.md) - 进阶特性和概念

#### 🚀 应用阶段 (5-8周)
- [03-concurrency/README.md](03-concurrency/README.md) - 并发编程精通
- [04-web/README.md](04-web/README.md) - Web开发技能

#### 🏗️ 架构阶段 (9-12周)
- [05-microservices/README.md](05-microservices/README.md) - 微服务架构
- [06-projects/README.md](06-projects/README.md) - 实战项目开发
- [06.5-performance-fundamentals/README.md](06.5-performance-fundamentals/README.md) - 性能基础原理

#### 🔬 专家阶段 (13-18个月)
- [07-runtime-internals/README.md](07-runtime-internals/README.md) - Go运行时内核
- [08-performance-mastery/README.md](08-performance-mastery/README.md) - 性能优化专精

#### ⚙️ 大师阶段 (6-15个月)
- [09-system-programming/README.md](09-system-programming/README.md) - 系统编程
- [10-compiler-toolchain/README.md](10-compiler-toolchain/README.md) - 编译器工具链

#### 🏛️ 通天阶段 (15-24个月)
- [11-massive-systems/README.md](11-massive-systems/README.md) - 大规模系统
- [12-ecosystem-contribution/README.md](12-ecosystem-contribution/README.md) - 生态贡献
- [13-language-design/README.md](13-language-design/README.md) - 语言设计
- [14-tech-leadership/README.md](14-tech-leadership/README.md) - 技术领导力
- [15-opensource-contribution/README.md](15-opensource-contribution/README.md) - 开源项目贡献实践

## 🛠️ 开发环境文档

### 🐳 Docker环境
- **开发环境**: 热重载 + 调试端口 + 数据库
- **生产环境**: 优化镜像 + 安全配置 + 监控
- **测试环境**: 专用测试数据库 + 集成测试
- **性能环境**: 基准测试 + 性能分析工具

详细配置请参考：[Docker开发环境章节](README.md#-docker开发环境)

### 📊 监控和观测
- **Prometheus**: 指标收集 (端口: 9091)
- **Grafana**: 监控面板 (端口: 3000, admin/admin123)
- **Jaeger**: 分布式链路追踪 (端口: 16686)
- **pprof**: Go应用性能分析 (端口: 6060)

### 🗄️ 数据存储
- **PostgreSQL**: 关系型数据库 (开发: 5432, 生产: 5433, 测试: 5434)
- **Redis**: 缓存服务 (开发: 6379, 生产: 6380, 测试: 6381)
- **MinIO**: 对象存储 (API: 9000, 控制台: 9001)

## 🚀 快速导航

### 新手入门
1. 阅读 [README.md](README.md) 了解项目概况
2. 参考 [LEARNING_GUIDE.md](LEARNING_GUIDE.md) 制定学习计划
3. 使用 [QUICK_REFERENCE.md](QUICK_REFERENCE.md) 快速上手

### 开发者
1. 查阅 [CONTRIBUTING.md](CONTRIBUTING.md) 了解开发规范
2. 使用 [QUICK_REFERENCE.md](QUICK_REFERENCE.md) 查找常用命令
3. 参考各模块README了解技术细节

### 学习者
1. 按照 [LEARNING_GUIDE.md](LEARNING_GUIDE.md) 的时间规划学习
2. 使用 Docker 环境进行实践
3. 完成每个阶段的验收标准

## 📋 文档维护状态

### ✅ 已完成文档
- [x] 项目主README
- [x] 学习路径指南
- [x] 快速参考手册
- [x] Docker环境配置
- [x] CI/CD管道配置
- [x] 代码质量标准

### 🔄 定期更新
- 学习进度跟踪
- 新特性文档
- 最佳实践更新
- 社区反馈整合

## 🤝 文档贡献

### 如何贡献文档
1. Fork项目并创建分支
2. 更新或创建文档
3. 确保文档格式一致
4. 提交Pull Request

### 文档规范
- 使用Markdown格式
- 包含目录结构
- 添加代码示例
- 保持简洁明了

## 🔗 外部资源

### 官方资源
- [Go官方文档](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go语言规范](https://golang.org/ref/spec)

### 学习资源
- [Go by Example](https://gobyexample.com)
- [Go语言圣经](https://books.studygolang.com/gopl-zh/)
- [Go语言中文网](https://studygolang.com)

### 社区资源
- [Awesome Go](https://github.com/avelino/awesome-go)
- [Go Forum](https://forum.golangbridge.org)
- [Reddit r/golang](https://reddit.com/r/golang)

---

**🎯 文档导航提示**:
- 📚 初学者从 README 开始
- 🚀 开发者查看 QUICK_REFERENCE
- 🎯 学习者遵循 LEARNING_GUIDE
- 🤝 贡献者参考 CONTRIBUTING

**Happy Learning & Coding! 🚀**