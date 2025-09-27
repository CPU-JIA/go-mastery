# ================================
# Build Stage
# ================================
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o blog-server cmd/server/main.go

# ================================
# Final Stage
# ================================
FROM alpine:3.18

# 安装必要的包
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser

# 设置工作目录
WORKDIR /app

# 创建必要的目录
RUN mkdir -p logs uploads && \
    chown -R appuser:appuser /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/blog-server .
COPY --from=builder /app/configs ./configs

# 复制静态文件（如果有）
# COPY --from=builder /app/assets ./assets

# 设置文件权限
RUN chmod +x blog-server

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./blog-server"]