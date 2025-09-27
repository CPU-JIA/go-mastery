# 🚀 现代化博客系统 - Enterprise Blog System

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.24)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10+-00D4AA?style=for-the-badge)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-v1.25+-FF6B6B?style=for-the-badge)](https://gorm.io/)

一个基于Go语言开发的现代化企业级博客系统，采用Clean Architecture设计，集成了2025年的最佳实践和企业级特性。

## ✨ 核心特性

### 🏗️ 架构设计
- **Clean Architecture** - 清晰的分层架构设计
- **Repository模式** - 数据访问层抽象
- **依赖注入** - 松耦合的组件设计
- **RESTful API** - 标准的REST接口设计

### 🔒 安全特性
- **JWT认证** - 基于Token的身份验证
- **密码加密** - bcrypt安全哈希
- **权限控制** - 基于角色的访问控制(RBAC)
- **输入验证** - 完整的数据验证和清理
- **CORS支持** - 跨域资源共享配置

### 📊 数据管理
- **GORM集成** - 现代化的Go ORM
- **自动迁移** - 数据库结构自动同步
- **软删除** - 数据安全删除机制
- **关联关系** - 完整的数据关联支持
- **事务支持** - 数据一致性保证

### 🌐 Web框架
- **Gin Framework** - 高性能HTTP框架
- **中间件链** - 灵活的请求处理管道
- **路由分组** - 清晰的API版本管理
- **优雅停机** - 服务平稳关闭支持

### 🚀 企业级特性
- **配置管理** - 基于Viper的配置系统
- **结构化日志** - 完整的请求日志记录
- **健康检查** - 服务状态监控端点
- **限流保护** - API调用频率控制
- **分页支持** - 标准化分页实现

## 📋 系统要求

- **Go 1.24+**
- **SQLite 3.x** (开发环境)
- **PostgreSQL 12+** (生产环境推荐)
- **Redis 6.0+** (缓存，可选)

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd 01-blog-system
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置环境

复制并编辑配置文件：

```bash
cp configs/config.yaml.example configs/config.yaml
```

创建环境变量文件：

```bash
cp .env.example .env
```

### 4. 启动服务

```bash
# 开发环境
go run cmd/server/main.go

# 或使用 air 热重载
air

# 生产环境构建
go build -o blog-server cmd/server/main.go
./blog-server
```

### 5. 验证安装

访问以下端点验证服务：

- **健康检查**: http://localhost:8080/health
- **API文档**: http://localhost:8080/api/v1 (开发中)

## 📖 API 文档

### 认证接口

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 用户注册 | 否 |
| POST | `/api/v1/auth/login` | 用户登录 | 否 |
| POST | `/api/v1/auth/refresh` | 刷新令牌 | 否 |

### 文章接口

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/articles` | 获取文章列表 | 可选 |
| GET | `/api/v1/articles/:id` | 获取文章详情 | 可选 |
| GET | `/api/v1/articles/slug/:slug` | 通过slug获取文章 | 可选 |
| GET | `/api/v1/articles/search` | 搜索文章 | 可选 |
| POST | `/api/v1/articles` | 创建文章 | 作者+ |
| PUT | `/api/v1/articles/:id` | 更新文章 | 作者+ |
| DELETE | `/api/v1/articles/:id` | 删除文章 | 作者+ |

### 用户接口

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/user/me` | 获取当前用户信息 | 是 |
| PUT | `/api/v1/user/password` | 修改密码 | 是 |

## 🏗️ 项目结构

```
01-blog-system/
├── cmd/server/           # 应用入口
│   └── main.go
├── internal/             # 私有代码
│   ├── config/          # 配置管理
│   ├── handler/         # HTTP处理器
│   ├── middleware/      # 中间件
│   ├── model/          # 数据模型
│   ├── repository/     # 数据仓储
│   └── service/        # 业务逻辑
├── configs/            # 配置文件
│   └── config.yaml
├── docs/              # 项目文档
├── scripts/           # 脚本文件
├── test/             # 测试文件
├── go.mod            # Go模块定义
└── README.md         # 项目说明
```

## 🔧 配置说明

### 服务器配置

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"          # debug, release, test
  read_timeout: 60s
  write_timeout: 60s
```

### 数据库配置

```yaml
database:
  driver: "sqlite"       # sqlite, postgres
  sqlite:
    path: "blog.db"
  postgres:
    host: "localhost"
    port: 5432
    user: "blog_user"
    password: "blog_password"
    dbname: "blog_system"
```

### JWT配置

```yaml
jwt:
  secret: "your-super-secret-jwt-key"
  expires_in: 24h
  refresh_expires_in: 168h
```

## 🐳 Docker 部署

### 构建镜像

```bash
docker build -t blog-system:latest .
```

### 运行容器

```bash
docker run -d -p 8080:8080 \
  --name blog-system \
  -e DATABASE_DRIVER=sqlite \
  blog-system:latest
```

### Docker Compose

```bash
docker-compose up -d
```

## 📝 开发指南

### 添加新功能

1. **数据模型** - 在 `internal/model/` 中定义
2. **数据仓储** - 在 `internal/repository/` 中实现
3. **业务逻辑** - 在 `internal/service/` 中实现
4. **HTTP处理** - 在 `internal/handler/` 中实现
5. **路由注册** - 在 `cmd/server/main.go` 中添加

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/service/

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 代码质量

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 使用 golangci-lint
golangci-lint run
```

## 🔍 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库配置
   - 确认数据库服务运行状态

2. **JWT验证失败**
   - 检查JWT secret配置
   - 确认token格式正确

3. **权限被拒绝**
   - 检查用户角色设置
   - 确认路由权限配置

### 日志查看

```bash
# 实时查看日志
tail -f logs/app.log

# 搜索错误日志
grep "ERROR" logs/app.log
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Gin](https://gin-gonic.com/) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [Viper](https://github.com/spf13/viper) - Configuration management
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT implementation

---

**🎉 恭喜！** 您已成功部署现代化企业级博客系统！