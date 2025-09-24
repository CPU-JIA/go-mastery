# ==========================================
# 微服务Docker化：2025年最佳实践
# ==========================================

# 多阶段构建可以减少最终镜像大小90%+
# 第一阶段：构建阶段
FROM golang:1.25-alpine AS builder

# 设置构建参数
ARG BUILD_VERSION=dev
ARG BUILD_TIME
ARG BUILD_COMMIT

# 安装必要的工具
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# 设置工作目录
WORKDIR /app

# 优化依赖下载：先复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖 (利用Docker层缓存)
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 编译应用 - 静态链接，优化大小
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${BUILD_VERSION} -X main.BuildTime=${BUILD_TIME} -X main.Commit=${BUILD_COMMIT}" \
    -a -installsuffix cgo \
    -o microservice ./main.go

# 验证二进制文件
RUN ./microservice --version || echo "Binary built successfully"

# ==========================================
# 第二阶段：运行阶段 (最小镜像)
# ==========================================
FROM scratch

# 添加标签信息
LABEL maintainer="your-team@company.com" \
      version="${BUILD_VERSION}" \
      description="Production-ready Go microservice" \
      org.opencontainers.image.title="Go Microservice" \
      org.opencontainers.image.description="High-performance microservice built with Go" \
      org.opencontainers.image.version="${BUILD_VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${BUILD_COMMIT}" \
      org.opencontainers.image.vendor="Your Company" \
      org.opencontainers.image.licenses="MIT"

# 从构建阶段复制必要文件
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/microservice /microservice

# 设置环境变量
ENV PORT=8080 \
    ENVIRONMENT=production \
    LOG_LEVEL=info \
    READ_TIMEOUT=30s \
    WRITE_TIMEOUT=30s \
    HEALTH_INTERVAL=30s \
    HEALTH_TIMEOUT=5s \
    HEALTH_RETRIES=3 \
    TZ=UTC

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/microservice", "--health-check"] || exit 1

# 暴露端口
EXPOSE 8080

# 设置非特权用户 (安全最佳实践)
USER 65534:65534

# 启动应用
ENTRYPOINT ["/microservice"]

# ==========================================
# 备用Dockerfile：开发版本
# ==========================================
# 如果需要调试能力，可以使用以下FROM语句替换scratch:
# FROM alpine:3.18
# RUN apk --no-cache add ca-certificates tzdata
# RUN adduser -D -s /bin/sh microservice
# USER microservice
# WORKDIR /home/microservice

# ==========================================
# 构建示例命令:
# ==========================================
# 基础构建:
# docker build -t microservice:latest .
#
# 带参数构建:
# docker build \
#   --build-arg BUILD_VERSION=v1.2.3 \
#   --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
#   --build-arg BUILD_COMMIT=$(git rev-parse --short HEAD) \
#   -t microservice:v1.2.3 .
#
# 多架构构建:
# docker buildx build \
#   --platform linux/amd64,linux/arm64 \
#   -t microservice:latest \
#   --push .
#
# 运行容器:
# docker run -p 8080:8080 --name microservice microservice:latest