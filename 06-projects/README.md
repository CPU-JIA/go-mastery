# 🚀 第六阶段：实战项目 (06-projects)

> **学习目标**: 通过6个实战项目掌握Go语言的实际应用
>
> **学习时长**: 4-6周
>
> **难度级别**: ⭐️⭐️⭐️⭐️ 高级

---

## 🎯 项目概览

本阶段包含**6个完整的实战项目**，每个项目都是可独立运行的应用程序，涵盖了从Web开发到分布式系统的各个领域。

```mermaid
graph LR
    A[01-blog-system<br/>博客系统] --> B[02-ecommerce-backend<br/>电商后端]
    B --> C[03-chat-system<br/>聊天系统]
    C --> D[04-task-scheduler<br/>任务调度器]
    D --> E[05-monitoring-system<br/>监控系统]
    E --> F[06-file-storage<br/>文件存储服务]
```

---

## 📖 项目详解

### 1️⃣ [01-blog-system](./01-blog-system/) - 博客系统 📝

**技术栈**:
- **Web框架**: Gin / Echo
- **数据库**: PostgreSQL + GORM
- **缓存**: Redis
- **认证**: JWT
- **前端**: Vue.js (可选)

**核心功能**:
- ✅ 用户注册/登录（JWT认证）
- ✅ 文章CRUD操作
- ✅ 评论系统
- ✅ 标签分类
- ✅ Markdown编辑器
- ✅ 全文搜索

**学习要点**:
- RESTful API设计
- 数据库关系建模
- 中间件使用（认证、日志、CORS）
- 缓存策略
- 分页和排序

**运行方式**:
```bash
cd 01-blog-system
go build
./blog-system
# 访问 http://localhost:8080
```

---

### 2️⃣ [02-ecommerce-backend](./02-ecommerce-backend/) - 电商后端 🛒

**技术栈**:
- **微服务架构**: gRPC
- **数据库**: PostgreSQL + MongoDB
- **消息队列**: RabbitMQ / Kafka
- **缓存**: Redis
- **搜索**: Elasticsearch (可选)

**核心功能**:
- ✅ 商品管理（增删改查）
- ✅ 购物车系统
- ✅ 订单处理流程
- ✅ 库存管理
- ✅ 支付集成（模拟）
- ✅ 优惠券系统

**微服务拆分**:
- `user-service`: 用户服务
- `product-service`: 商品服务
- `order-service`: 订单服务
- `payment-service`: 支付服务

**学习要点**:
- 微服务架构设计
- gRPC服务间通信
- 分布式事务处理（SAGA模式）
- 消息队列异步处理
- 库存并发扣减

---

### 3️⃣ [03-chat-system](./03-chat-system/) - 聊天系统 💬

**技术栈**:
- **实时通信**: WebSocket
- **数据库**: MongoDB (消息存储)
- **缓存**: Redis (在线用户、未读消息)
- **消息推送**: 自研推送服务

**核心功能**:
- ✅ 实时消息收发
- ✅ 单聊和群聊
- ✅ 在线状态显示
- ✅ 消息历史记录
- ✅ 文件传输
- ✅ 消息已读回执

**技术亮点**:
- WebSocket长连接管理
- 消息广播机制
- 离线消息推送
- 并发连接处理

**学习要点**:
- WebSocket协议实现
- 长连接维护和心跳机制
- 消息分发和路由
- 高并发连接处理

---

### 4️⃣ [04-task-scheduler](./04-task-scheduler/) - 任务调度器 ⏰

**技术栈**:
- **调度引擎**: Cron表达式解析
- **任务队列**: Redis + RabbitMQ
- **数据库**: PostgreSQL
- **分布式锁**: Redis分布式锁

**核心功能**:
- ✅ 定时任务调度（Cron）
- ✅ 延迟任务执行
- ✅ 任务依赖管理
- ✅ 失败重试机制
- ✅ 任务执行日志
- ✅ 分布式任务调度

**调度策略**:
- **时间触发**: Cron表达式
- **事件触发**: 消息队列
- **依赖触发**: DAG任务图

**学习要点**:
- Cron表达式解析和调度
- 分布式锁避免重复执行
- 任务队列设计
- 失败重试策略
- DAG任务依赖

---

### 5️⃣ [05-monitoring-system](./05-monitoring-system/) - 监控系统 📊

**技术栈**:
- **指标采集**: Prometheus格式
- **时序数据库**: InfluxDB / Prometheus
- **可视化**: Grafana
- **告警**: 自研告警引擎

**核心功能**:
- ✅ 指标采集和上报
- ✅ 时序数据存储
- ✅ 实时指标查询
- ✅ 告警规则配置
- ✅ 告警通知（邮件/短信/钉钉）
- ✅ 可视化面板

**监控指标**:
- **系统指标**: CPU、内存、磁盘
- **应用指标**: QPS、延迟、错误率
- **业务指标**: 订单量、交易额

**学习要点**:
- Prometheus指标格式
- 时序数据存储和查询
- 告警规则引擎
- 通知渠道集成

---

### 6️⃣ [06-file-storage](./06-file-storage/) - 文件存储服务 📁

**技术栈**:
- **对象存储**: MinIO / OSS
- **CDN**: 云厂商CDN (可选)
- **数据库**: PostgreSQL (元数据)
- **缓存**: Redis

**核心功能**:
- ✅ 文件上传/下载
- ✅ 断点续传
- ✅ 秒传（文件去重）
- ✅ 图片缩略图生成
- ✅ 文件分享（临时链接）
- ✅ 访问权限控制

**技术亮点**:
- 分片上传和合并
- 文件哈希去重
- 图片处理（缩放、裁剪）
- 预签名URL生成

**学习要点**:
- 大文件上传策略
- 文件哈希和去重
- 对象存储集成
- 图片处理库使用

---

## 🏗️ 项目架构对比

| 项目 | 架构模式 | 数据库 | 缓存 | 消息队列 | 难度 |
|------|---------|-------|------|---------|------|
| 01-blog-system | 单体应用 | PostgreSQL | Redis | - | ⭐️⭐️ |
| 02-ecommerce-backend | 微服务 | PG+Mongo | Redis | RabbitMQ | ⭐️⭐️⭐️⭐️ |
| 03-chat-system | WebSocket | MongoDB | Redis | - | ⭐️⭐️⭐️ |
| 04-task-scheduler | 分布式 | PostgreSQL | Redis | RabbitMQ | ⭐️⭐️⭐️⭐️ |
| 05-monitoring-system | 时序数据 | InfluxDB | - | - | ⭐️⭐️⭐️ |
| 06-file-storage | 对象存储 | PostgreSQL | Redis | - | ⭐️⭐️⭐️ |

---

## 🚀 快速启动所有项目

### 方式一：Docker Compose (推荐)
```bash
# 启动所有项目及其依赖服务
cd 06-projects
docker-compose up -d

# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs -f blog-system
```

### 方式二：本地运行
```bash
# 启动PostgreSQL和Redis (Docker)
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:15
docker run -d -p 6379:6379 redis:7

# 编译并运行各个项目
cd 01-blog-system && go build && ./blog-system &
cd 02-ecommerce-backend && go build && ./ecommerce-backend &
cd 03-chat-system && go build && ./chat-system &
cd 04-task-scheduler && go build && ./task-scheduler &
cd 05-monitoring-system && go build && ./monitoring-system &
cd 06-file-storage && go build && ./file-storage &
```

---

## 📊 质量统计

- **总项目数**: 6个完整应用
- **代码行数**: 约25000行（包含注释）
- **编译状态**: ✅ 全部编译成功
- **测试覆盖**: 部分项目有单元测试
- **文档完整性**: 每个项目有独立README

---

## 🎯 学习建议

### 推荐学习顺序
1. **01-blog-system** - 从简单的单体应用开始，理解Web开发基础
2. **03-chat-system** - 学习WebSocket实时通信
3. **06-file-storage** - 学习文件处理和对象存储
4. **05-monitoring-system** - 学习时序数据和监控
5. **04-task-scheduler** - 学习分布式调度
6. **02-ecommerce-backend** - 最后挑战微服务架构

### 学习策略
- **先运行**: 每个项目先运行起来，看看效果
- **读代码**: 从main.go入口开始，理解项目结构
- **改代码**: 尝试添加新功能或修改现有功能
- **写测试**: 为关键功能补充单元测试
- **重构**: 优化代码结构，应用设计模式

---

## 🔑 核心知识点

### Web开发 (01-blog-system, 03-chat-system)
- ✅ HTTP路由和中间件
- ✅ RESTful API设计
- ✅ JWT认证
- ✅ WebSocket通信

### 微服务 (02-ecommerce-backend)
- ✅ gRPC服务拆分
- ✅ 服务注册与发现
- ✅ API网关
- ✅ 分布式事务

### 分布式系统 (04-task-scheduler)
- ✅ 分布式锁
- ✅ 任务调度
- ✅ 消息队列
- ✅ 失败重试

### 数据处理 (05-monitoring-system, 06-file-storage)
- ✅ 时序数据存储
- ✅ 文件上传下载
- ✅ 对象存储集成
- ✅ 图片处理

---

## 🛠️ 技术栈总览

### 数据库
- **PostgreSQL**: 关系型数据库，适合结构化数据
- **MongoDB**: 文档数据库，适合消息存储
- **InfluxDB**: 时序数据库，适合监控指标

### 缓存
- **Redis**: 通用缓存，分布式锁，消息队列

### 消息队列
- **RabbitMQ**: 传统消息队列，支持多种模式
- **Kafka**: 高吞吐消息队列（可选）

### Web框架
- **Gin**: 高性能Web框架
- **Echo**: 简洁优雅的Web框架

### RPC框架
- **gRPC**: 高性能RPC框架

---

## 🔗 相关资源

### 前置学习
- ⬅️ [04-web](../04-web/) - Web开发基础
- ⬅️ [05-microservices](../05-microservices/) - 微服务架构

### 下一步学习
- ➡️ [07-runtime-internals](../07-runtime-internals/) - 运行时内核
- ➡️ [08-performance-mastery](../08-performance-mastery/) - 性能优化

---

## 💡 实战提升建议

1. **代码重构**: 优化项目代码结构，应用设计模式
2. **性能优化**: 使用pprof分析性能瓶颈
3. **添加测试**: 提高测试覆盖率到80%+
4. **容器化**: 为每个项目编写Dockerfile和docker-compose.yml
5. **CI/CD**: 配置GitHub Actions自动测试和部署
6. **监控**: 集成Prometheus和Grafana
7. **文档**: 补充API文档和架构设计文档

---

## 🎓 项目扩展方向

### 01-blog-system
- 添加Markdown实时预览
- 集成第三方登录（GitHub OAuth）
- 添加全文搜索（Elasticsearch）

### 02-ecommerce-backend
- 实现SAGA分布式事务
- 添加秒杀功能
- 集成支付宝/微信支付

### 03-chat-system
- 添加语音/视频通话
- 实现端到端加密
- 添加聊天机器人

### 04-task-scheduler
- 实现DAG任务依赖
- 添加任务可视化界面
- 支持动态添加任务

### 05-monitoring-system
- 添加分布式链路追踪
- 实现智能告警（机器学习）
- 集成日志分析

### 06-file-storage
- 添加视频转码
- 实现CDN加速
- 支持WebDAV协议

---

**作者**: JIA
**最后更新**: 2025-10-04
**质量标准**: 0错误0警告，最高标准，详尽中文注释
**编译状态**: ✅ 全部6个项目编译成功
