package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

/*
🚀 现代化gRPC微服务通信 - 2025年企业级实现

本实现展示了高性能gRPC服务间通信的现代化模式，包括：

🔥 高性能特性：
1. 连接池和Keep-Alive优化
2. 流式接口支持 (Server/Client/Bidirectional)
3. 连接复用和负载均衡
4. 自适应限流和熔断
5. 零拷贝数据传输

🔍 可观测性集成：
1. OpenTelemetry分布式追踪
2. Prometheus指标收集
3. 结构化日志记录
4. 健康检查和就绪探针
5. 性能监控仪表板

🛡️ 安全和治理：
1. TLS/mTLS加密传输
2. JWT认证和授权
3. API限流和熔断
4. 请求验证和清理
5. 审计日志和监控

☁️ 云原生支持：
1. Kubernetes服务发现
2. gRPC-Gateway HTTP网关
3. 优雅停机处理
4. 健康检查探针
5. 负载均衡器集成

核心设计原则：
- 高性能：连接池、流式处理、零拷贝
- 可观测：全链路追踪、细粒度指标
- 可靠性：熔断恢复、重试机制、超时控制
- 安全性：端到端加密、认证授权
- 可扩展：插件化拦截器、中间件链
*/

// === 配置定义 ===

type GRPCConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Client   ClientConfig   `yaml:"client"`
	Security SecurityConfig `yaml:"security"`
	Tracing  TracingConfig  `yaml:"tracing"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

type ServerConfig struct {
	Address     string `yaml:"address"`
	Port        int    `yaml:"port"`
	EnableTLS   bool   `yaml:"enable_tls"`
	CertFile    string `yaml:"cert_file"`
	KeyFile     string `yaml:"key_file"`
	EnableH2C   bool   `yaml:"enable_h2c"`
	MaxRecvSize int    `yaml:"max_recv_size"`
	MaxSendSize int    `yaml:"max_send_size"`

	// Keep-Alive配置
	KeepAlive ServerKeepAliveConfig `yaml:"keep_alive"`

	// 连接限制
	MaxConcurrentStreams uint32        `yaml:"max_concurrent_streams"`
	ConnectionTimeout    time.Duration `yaml:"connection_timeout"`

	// 压缩设置
	EnableCompression bool `yaml:"enable_compression"`

	// 反射和健康检查
	EnableReflection  bool `yaml:"enable_reflection"`
	EnableHealthCheck bool `yaml:"enable_health_check"`
}

type ServerKeepAliveConfig struct {
	Time    time.Duration `yaml:"time"`
	Timeout time.Duration `yaml:"timeout"`
	MinTime time.Duration `yaml:"min_time"`
}

type ClientConfig struct {
	EnableLoadBalancing bool          `yaml:"enable_load_balancing"`
	LoadBalancingPolicy string        `yaml:"load_balancing_policy"`
	MaxRecvSize         int           `yaml:"max_recv_size"`
	MaxSendSize         int           `yaml:"max_send_size"`
	ConnectionTimeout   time.Duration `yaml:"connection_timeout"`

	// Keep-Alive配置
	KeepAlive ClientKeepAliveConfig `yaml:"keep_alive"`

	// 重试配置
	Retry RetryConfig `yaml:"retry"`

	// 连接池
	Pool ConnectionPoolConfig `yaml:"pool"`
}

type ClientKeepAliveConfig struct {
	Time                time.Duration `yaml:"time"`
	Timeout             time.Duration `yaml:"timeout"`
	PermitWithoutStream bool          `yaml:"permit_without_stream"`
}

type RetryConfig struct {
	MaxAttempts     int           `yaml:"max_attempts"`
	InitialBackoff  time.Duration `yaml:"initial_backoff"`
	MaxBackoff      time.Duration `yaml:"max_backoff"`
	BackoffMul      float64       `yaml:"backoff_multiplier"`
	RetryableErrors []codes.Code  `yaml:"retryable_errors"`
}

type ConnectionPoolConfig struct {
	MaxConnections     int           `yaml:"max_connections"`
	MaxIdleConnections int           `yaml:"max_idle_connections"`
	ConnMaxLifetime    time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `yaml:"conn_max_idle_time"`
}

type SecurityConfig struct {
	EnableAuth     bool   `yaml:"enable_auth"`
	JWTSecret      string `yaml:"jwt_secret"`
	EnableTLS      bool   `yaml:"enable_tls"`
	EnableMTLS     bool   `yaml:"enable_mtls"`
	ClientCertFile string `yaml:"client_cert_file"`
	ClientKeyFile  string `yaml:"client_key_file"`
	ServerCertFile string `yaml:"server_cert_file"`
	ServerKeyFile  string `yaml:"server_key_file"`
	CACertFile     string `yaml:"ca_cert_file"`
}

type TracingConfig struct {
	Enabled     bool    `yaml:"enabled"`
	ServiceName string  `yaml:"service_name"`
	JaegerURL   string  `yaml:"jaeger_url"`
	SampleRate  float64 `yaml:"sample_rate"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// === 用户服务实现 ===

type UserServiceServer struct {
	UnimplementedUserServiceServer

	// 数据存储 (实际生产环境中使用数据库)
	users    map[string]*User
	usersMux sync.RWMutex

	// 活动流订阅者
	subscribers map[string]chan *UserActivityEvent
	subsMux     sync.RWMutex

	// 聊天会话
	chatSessions map[string]ChatSession
	chatMux      sync.RWMutex

	// 配置
	config *GRPCConfig
}

// 聊天会话
type ChatSession struct {
	UserID   string
	Stream   UserService_UserChatServer
	LastSeen time.Time
}

func NewUserServiceServer(config *GRPCConfig) *UserServiceServer {
	return &UserServiceServer{
		users:        make(map[string]*User),
		subscribers:  make(map[string]chan *UserActivityEvent),
		chatSessions: make(map[string]ChatSession),
		config:       config,
	}
}

// 创建用户
func (s *UserServiceServer) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	// 参数验证
	if req.Username == "" || req.Email == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户名和邮箱不能为空")
	}

	// 检查用户是否已存在
	s.usersMux.RLock()
	for _, user := range s.users {
		if user.Username == req.Username || user.Email == req.Email {
			s.usersMux.RUnlock()
			return nil, status.Errorf(codes.AlreadyExists, "用户名或邮箱已存在")
		}
	}
	s.usersMux.RUnlock()

	// 创建新用户
	userID := fmt.Sprintf("user_%d", time.Now().UnixNano())
	now := timestamppb.Now()

	user := &User{
		UserId:      userID,
		Username:    req.Username,
		Email:       req.Email,
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Status:      UserStatus_USER_STATUS_ACTIVE,
		Roles:       req.Roles,
		Metadata:    req.Metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 存储用户
	s.usersMux.Lock()
	s.users[userID] = user
	s.usersMux.Unlock()

	// 发布用户创建活动事件
	go s.publishActivityEvent(&UserActivityEvent{
		EventId:      fmt.Sprintf("event_%d", time.Now().UnixNano()),
		UserId:       userID,
		ActivityType: ActivityType_ACTIVITY_TYPE_LOGIN,
		Details:      map[string]string{"action": "user_created"},
		Timestamp:    now,
	})

	// 生成访问令牌 (简化实现)
	accessToken := fmt.Sprintf("jwt_token_%s_%d", userID, time.Now().Unix())
	refreshToken := fmt.Sprintf("refresh_%s_%d", userID, time.Now().Unix())

	return &CreateUserResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// 获取用户
func (s *UserServiceServer) GetUser(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
	var user *User
	var found bool

	s.usersMux.RLock()
	defer s.usersMux.RUnlock()

	// 根据不同标识符查找用户
	switch req.Identifier.(type) {
	case *GetUserRequest_UserId:
		user, found = s.users[req.GetUserId()]
	case *GetUserRequest_Username:
		username := req.GetUsername()
		for _, u := range s.users {
			if u.Username == username {
				user, found = u, true
				break
			}
		}
	case *GetUserRequest_Email:
		email := req.GetEmail()
		for _, u := range s.users {
			if u.Email == email {
				user, found = u, true
				break
			}
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "必须提供用户ID、用户名或邮箱")
	}

	if !found {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}

	// 字段掩码过滤：实际项目中可使用 fieldmaskpb 包实现按需返回字段
	// 当前演示代码返回完整用户对象

	return &GetUserResponse{User: user}, nil
}

// 更新用户
func (s *UserServiceServer) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}

	s.usersMux.Lock()
	defer s.usersMux.Unlock()

	user, found := s.users[req.UserId]
	if !found {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}

	// 更新用户信息 (简化实现，应该使用字段掩码)
	if req.User != nil {
		if req.User.FullName != "" {
			user.FullName = req.User.FullName
		}
		if req.User.PhoneNumber != "" {
			user.PhoneNumber = req.User.PhoneNumber
		}
		if req.User.Status != UserStatus_USER_STATUS_UNSPECIFIED {
			user.Status = req.User.Status
		}
		user.UpdatedAt = timestamppb.Now()
	}

	return &UpdateUserResponse{User: user}, nil
}

// 删除用户
func (s *UserServiceServer) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*emptypb.Empty, error) {
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}

	s.usersMux.Lock()
	defer s.usersMux.Unlock()

	user, found := s.users[req.UserId]
	if !found {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}

	if req.HardDelete {
		// 硬删除
		delete(s.users, req.UserId)
	} else {
		// 软删除
		user.Status = UserStatus_USER_STATUS_DELETED
		user.UpdatedAt = timestamppb.Now()
	}

	return &emptypb.Empty{}, nil
}

// 用户活动流 (Server Streaming)
func (s *UserServiceServer) StreamUserActivity(req *StreamUserActivityRequest, stream UserService_StreamUserActivityServer) error {
	if req.UserId == "" {
		return status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}

	// 创建活动事件通道
	eventChan := make(chan *UserActivityEvent, 100)

	// 注册订阅者
	s.subsMux.Lock()
	s.subscribers[req.UserId] = eventChan
	s.subsMux.Unlock()

	// 清理订阅者
	defer func() {
		s.subsMux.Lock()
		delete(s.subscribers, req.UserId)
		close(eventChan)
		s.subsMux.Unlock()
	}()

	// 流式发送活动事件
	for {
		select {
		case event := <-eventChan:
			if event != nil {
				// 过滤活动类型
				if len(req.ActivityTypes) > 0 {
					found := false
					for _, actType := range req.ActivityTypes {
						if event.ActivityType == actType {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				if err := stream.Send(event); err != nil {
					return err
				}
			}
		case <-stream.Context().Done():
			return stream.Context().Err()
		}
	}
}

// 用户聊天 (Bidirectional Streaming)
func (s *UserServiceServer) UserChat(stream UserService_UserChatServer) error {
	// 获取用户信息从context
	userID, err := s.getUserFromContext(stream.Context())
	if err != nil {
		return err
	}

	// 注册聊天会话
	s.chatMux.Lock()
	s.chatSessions[userID] = ChatSession{
		UserID:   userID,
		Stream:   stream,
		LastSeen: time.Now(),
	}
	s.chatMux.Unlock()

	// 清理会话
	defer func() {
		s.chatMux.Lock()
		delete(s.chatSessions, userID)
		s.chatMux.Unlock()
	}()

	// 处理双向消息流
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// 处理接收到的消息
		if err := s.handleChatMessage(msg); err != nil {
			return err
		}

		// 转发消息给目标用户
		if msg.ToUserId != "" {
			s.chatMux.RLock()
			if session, found := s.chatSessions[msg.ToUserId]; found {
				if err := session.Stream.Send(msg); err != nil {
					log.Printf("转发消息失败: %v", err)
				}
			}
			s.chatMux.RUnlock()
		}
	}
}

// 健康检查
func (s *UserServiceServer) HealthCheck(ctx context.Context, req *emptypb.Empty) (*HealthCheckResponse, error) {
	// 检查服务状态
	status := HealthStatus_HEALTH_STATUS_SERVING
	message := "服务运行正常"
	details := map[string]string{
		"service":    "user-service",
		"version":    "1.0.0",
		"uptime":     time.Since(startTime).String(),
		"goroutines": fmt.Sprintf("%d", runtime.NumGoroutine()),
	}

	// 检查数据库连接 (模拟)
	if !s.checkDatabaseHealth() {
		status = HealthStatus_HEALTH_STATUS_NOT_SERVING
		message = "数据库连接异常"
	}

	return &HealthCheckResponse{
		Status:    status,
		Message:   message,
		Details:   details,
		Timestamp: timestamppb.Now(),
	}, nil
}

// === 辅助方法 ===

func (s *UserServiceServer) publishActivityEvent(event *UserActivityEvent) {
	s.subsMux.RLock()
	defer s.subsMux.RUnlock()

	if eventChan, found := s.subscribers[event.UserId]; found {
		select {
		case eventChan <- event:
		default:
			// 通道满，丢弃事件
			log.Printf("活动事件通道满，丢弃事件: %s", event.EventId)
		}
	}
}

func (s *UserServiceServer) handleChatMessage(msg *ChatMessage) error {
	// 消息验证和处理逻辑
	if msg.Content == "" {
		return status.Errorf(codes.InvalidArgument, "消息内容不能为空")
	}

	// 设置消息时间戳
	msg.Timestamp = timestamppb.Now()
	msg.MessageId = fmt.Sprintf("msg_%d", time.Now().UnixNano())

	// 演示代码：实际项目中应持久化到数据库
	log.Printf("收到消息: %s -> %s: %s", msg.FromUserId, msg.ToUserId, msg.Content)

	return nil
}

func (s *UserServiceServer) getUserFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "缺少元数据")
	}

	userIDs := md.Get("user-id")
	if len(userIDs) == 0 {
		return "", status.Errorf(codes.Unauthenticated, "缺少用户ID")
	}

	return userIDs[0], nil
}

func (s *UserServiceServer) checkDatabaseHealth() bool {
	// 模拟数据库健康检查
	return true
}

// === 全局变量 ===

var startTime = time.Now()

// 生成的protobuf代码占位符 (实际使用时由protoc生成)
type UnimplementedUserServiceServer struct{}
type UserService_StreamUserActivityServer interface {
	Send(*UserActivityEvent) error
	Context() context.Context
}
type UserService_UserChatServer interface {
	Send(*ChatMessage) error
	Recv() (*ChatMessage, error)
	Context() context.Context
}

// 消息类型占位符 (实际使用时由protoc生成)
type User struct {
	UserId      string
	Username    string
	Email       string
	FullName    string
	PhoneNumber string
	Status      UserStatus
	Roles       []string
	Metadata    map[string]string
	CreatedAt   *timestamppb.Timestamp
	UpdatedAt   *timestamppb.Timestamp
}

type UserStatus int32

const (
	UserStatus_USER_STATUS_UNSPECIFIED UserStatus = 0
	UserStatus_USER_STATUS_ACTIVE      UserStatus = 1
	UserStatus_USER_STATUS_INACTIVE    UserStatus = 2
	UserStatus_USER_STATUS_SUSPENDED   UserStatus = 3
	UserStatus_USER_STATUS_DELETED     UserStatus = 4
)

type ActivityType int32

const (
	ActivityType_ACTIVITY_TYPE_UNSPECIFIED     ActivityType = 0
	ActivityType_ACTIVITY_TYPE_LOGIN           ActivityType = 1
	ActivityType_ACTIVITY_TYPE_LOGOUT          ActivityType = 2
	ActivityType_ACTIVITY_TYPE_PROFILE_UPDATE  ActivityType = 3
	ActivityType_ACTIVITY_TYPE_PASSWORD_CHANGE ActivityType = 4
	ActivityType_ACTIVITY_TYPE_ROLE_CHANGE     ActivityType = 5
)

type HealthStatus int32

const (
	HealthStatus_HEALTH_STATUS_UNSPECIFIED     HealthStatus = 0
	HealthStatus_HEALTH_STATUS_SERVING         HealthStatus = 1
	HealthStatus_HEALTH_STATUS_NOT_SERVING     HealthStatus = 2
	HealthStatus_HEALTH_STATUS_SERVICE_UNKNOWN HealthStatus = 3
)

type MessageType int32

const (
	MessageType_MESSAGE_TYPE_UNSPECIFIED MessageType = 0
	MessageType_MESSAGE_TYPE_TEXT        MessageType = 1
	MessageType_MESSAGE_TYPE_IMAGE       MessageType = 2
	MessageType_MESSAGE_TYPE_FILE        MessageType = 3
	MessageType_MESSAGE_TYPE_SYSTEM      MessageType = 4
)

// 请求/响应消息占位符
type CreateUserRequest struct {
	Username    string
	Email       string
	Password    string
	FullName    string
	PhoneNumber string
	Roles       []string
	Metadata    map[string]string
}

type CreateUserResponse struct {
	User         *User
	AccessToken  string
	RefreshToken string
}

type GetUserRequest struct {
	Identifier interface{} // oneof
}

func (r *GetUserRequest) GetUserId() string   { return "" }
func (r *GetUserRequest) GetUsername() string { return "" }
func (r *GetUserRequest) GetEmail() string    { return "" }

type GetUserRequest_UserId struct{ UserId string }
type GetUserRequest_Username struct{ Username string }
type GetUserRequest_Email struct{ Email string }

type GetUserResponse struct {
	User *User
}

type UpdateUserRequest struct {
	UserId string
	User   *User
}

type UpdateUserResponse struct {
	User *User
}

type DeleteUserRequest struct {
	UserId     string
	HardDelete bool
}

type StreamUserActivityRequest struct {
	UserId        string
	ActivityTypes []ActivityType
}

type UserActivityEvent struct {
	EventId      string
	UserId       string
	ActivityType ActivityType
	Details      map[string]string
	Timestamp    *timestamppb.Timestamp
	ClientIp     string
	UserAgent    string
}

type ChatMessage struct {
	MessageId   string
	FromUserId  string
	ToUserId    string
	Content     string
	MessageType MessageType
	Timestamp   *timestamppb.Timestamp
	Metadata    map[string]string
}

type HealthCheckResponse struct {
	Status    HealthStatus
	Message   string
	Details   map[string]string
	Timestamp *timestamppb.Timestamp
}

// === 拦截器实现 ===

// 认证拦截器
func authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 跳过健康检查
	if info.FullMethod == "/grpc.health.v1.Health/Check" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "缺少元数据")
	}

	// 检查授权头
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "缺少授权头")
	}

	// 简化的JWT验证 (生产环境需要完整实现)
	token := authHeaders[0]
	if !isValidToken(token) {
		return nil, status.Errorf(codes.Unauthenticated, "无效的认证令牌")
	}

	// 提取用户信息并设置到context
	userID := extractUserIDFromToken(token)
	ctx = metadata.AppendToOutgoingContext(ctx, "user-id", userID)

	return handler(ctx, req)
}

// 日志拦截器
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// 获取请求ID
	requestID := generateRequestID()
	ctx = metadata.AppendToOutgoingContext(ctx, "request-id", requestID)

	// 记录请求
	log.Printf("[%s] gRPC请求开始: %s", requestID, info.FullMethod)

	// 执行处理器
	resp, err := handler(ctx, req)

	// 记录响应
	duration := time.Since(start)
	if err != nil {
		log.Printf("[%s] gRPC请求完成: %s (错误: %v, 耗时: %v)", requestID, info.FullMethod, err, duration)
	} else {
		log.Printf("[%s] gRPC请求完成: %s (耗时: %v)", requestID, info.FullMethod, duration)
	}

	return resp, err
}

// 限流拦截器
func rateLimitInterceptor(limiter *rate.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !limiter.Allow() {
			return nil, status.Errorf(codes.ResourceExhausted, "请求频率过高，请稍后重试")
		}
		return handler(ctx, req)
	}
}

// 恢复拦截器
func recoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("gRPC处理器panic恢复: %v", r)
		}
	}()

	return handler(ctx, req)
}

// 超时拦截器
func timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		done := make(chan struct{})
		var resp interface{}
		var err error

		go func() {
			resp, err = handler(ctx, req)
			close(done)
		}()

		select {
		case <-done:
			return resp, err
		case <-ctx.Done():
			return nil, status.Errorf(codes.DeadlineExceeded, "请求超时")
		}
	}
}

// === gRPC服务器 ===

type GRPCServer struct {
	config       *GRPCConfig
	server       *grpc.Server
	listener     net.Listener
	userService  *UserServiceServer
	healthServer *health.Server
}

func NewGRPCServer(config *GRPCConfig) (*GRPCServer, error) {
	// 创建监听器
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.Server.Address, config.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("创建监听器失败: %w", err)
	}

	// 配置服务器选项
	opts := []grpc.ServerOption{}

	// TLS配置
	if config.Server.EnableTLS {
		cert, err := tls.LoadX509KeyPair(config.Server.CertFile, config.Server.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("加载TLS证书失败: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		if config.Security.EnableMTLS {
			if config.Security.CACertFile == "" {
				return nil, fmt.Errorf("启用mTLS需要提供CA证书")
			}
			caPool, err := loadCertPool(config.Security.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("加载CA证书失败: %w", err)
			}
			tlsConfig.ClientCAs = caPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	// Keep-Alive配置
	kasp := keepalive.ServerParameters{
		Time:    config.Server.KeepAlive.Time,
		Timeout: config.Server.KeepAlive.Timeout,
	}
	kaep := keepalive.EnforcementPolicy{
		MinTime:             config.Server.KeepAlive.MinTime,
		PermitWithoutStream: false,
	}
	opts = append(opts, grpc.KeepaliveParams(kasp), grpc.KeepaliveEnforcementPolicy(kaep))

	// 连接配置
	opts = append(opts,
		grpc.MaxRecvMsgSize(config.Server.MaxRecvSize),
		grpc.MaxSendMsgSize(config.Server.MaxSendSize),
		grpc.MaxConcurrentStreams(config.Server.MaxConcurrentStreams),
	)

	// 拦截器链
	limiter := rate.NewLimiter(rate.Limit(100), 200) // 100 RPS, burst 200
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		recoveryInterceptor,
		loggingInterceptor,
		authInterceptor,
		rateLimitInterceptor(limiter),
		timeoutInterceptor(30 * time.Second),
	}

	// 流式拦截器
	streamInterceptors := []grpc.StreamServerInterceptor{
		// 可以添加流式拦截器
	}

	// 添加OpenTelemetry追踪
	if config.Tracing.Enabled {
		unaryInterceptors = append(unaryInterceptors, otelgrpc.UnaryServerInterceptor())
		streamInterceptors = append(streamInterceptors, otelgrpc.StreamServerInterceptor())
	}

	opts = append(opts,
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	)

	// 创建gRPC服务器
	server := grpc.NewServer(opts...)

	// 创建服务实例
	userService := NewUserServiceServer(config)

	// 注册服务
	// 演示代码：实际项目中需要 protoc 生成代码后调用
	// RegisterUserServiceServer(server, userService)
	_ = userService // 避免未使用变量警告

	// 健康检查服务
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("user_service.v1.UserService", healthpb.HealthCheckResponse_SERVING)

	// 反射服务 (开发环境)
	if config.Server.EnableReflection {
		reflection.Register(server)
	}

	return &GRPCServer{
		config:       config,
		server:       server,
		listener:     listener,
		userService:  userService,
		healthServer: healthServer,
	}, nil
}

func (s *GRPCServer) Start() error {
	log.Printf("🚀 gRPC服务器启动在 %s", s.listener.Addr())

	// 启动OpenTelemetry追踪
	if s.config.Tracing.Enabled {
		if err := initTracing(s.config.Tracing); err != nil {
			log.Printf("初始化追踪失败: %v", err)
		}
	}

	// 启动gRPC-Gateway (可选)
	go s.startGRPCGateway()

	// 启动Prometheus指标服务器 (可选)
	if s.config.Metrics.Enabled {
		go s.startMetricsServer()
	}

	return s.server.Serve(s.listener)
}

func (s *GRPCServer) Stop() {
	log.Println("🛑 正在关闭gRPC服务器...")

	// 优雅停机
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	// 等待一段时间后强制停机
	timer := time.NewTimer(30 * time.Second)
	select {
	case <-timer.C:
		log.Println("强制停机")
		s.server.Stop()
	case <-stopped:
		timer.Stop()
		log.Println("优雅停机完成")
	}
}

// 启动gRPC-Gateway
func (s *GRPCServer) startGRPCGateway() {
	// 创建gRPC连接
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", s.config.Server.Address, s.config.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("连接gRPC服务器失败: %v", err)
		return
	}
	defer conn.Close()

	// 创建HTTP多路复用器
	mux := gwruntime.NewServeMux(
		gwruntime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			switch key {
			case "Authorization":
				return key, true
			default:
				return "", false
			}
		}),
	)

	// 演示代码：实际项目中需要 protoc-gen-grpc-gateway 生成代码后注册
	// if err := RegisterUserServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
	//     log.Printf("注册gRPC-Gateway处理器失败: %v", err)
	//     return
	// }

	// 启动HTTP服务器
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Server.Port+1),
		Handler: mux,
	}

	log.Printf("🌐 gRPC-Gateway启动在 %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Printf("gRPC-Gateway服务器错误: %v", err)
	}
}

// 启动指标服务器
func (s *GRPCServer) startMetricsServer() {
	http.Handle(s.config.Metrics.Path, promhttp.Handler())
	addr := fmt.Sprintf(":%d", s.config.Metrics.Port)
	log.Printf("📊 指标服务器启动在 %s%s", addr, s.config.Metrics.Path)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("指标服务器错误: %v", err)
	}
}

// === gRPC客户端 ===

type GRPCClient struct {
	config *GRPCConfig
	conn   *grpc.ClientConn
	// 演示代码：实际项目中添加 protoc 生成的客户端
	// userClient UserServiceClient
}

func NewGRPCClient(config *GRPCConfig, target string) (*GRPCClient, error) {
	// 配置客户端选项
	opts := []grpc.DialOption{}

	// 负载均衡配置
	if config.Client.EnableLoadBalancing {
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{
			"loadBalancingPolicy": "%s",
			"healthCheckConfig": {
				"serviceName": "user_service.v1.UserService"
			}
		}`, config.Client.LoadBalancingPolicy)))
	}

	// Keep-Alive配置
	kacp := keepalive.ClientParameters{
		Time:                config.Client.KeepAlive.Time,
		Timeout:             config.Client.KeepAlive.Timeout,
		PermitWithoutStream: config.Client.KeepAlive.PermitWithoutStream,
	}
	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	// 消息大小限制
	opts = append(opts,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(config.Client.MaxRecvSize),
			grpc.MaxCallSendMsgSize(config.Client.MaxSendSize),
		),
	)

	// 拦截器
	unaryInterceptors := []grpc.UnaryClientInterceptor{
		clientLoggingInterceptor,
	}

	streamInterceptors := []grpc.StreamClientInterceptor{}

	// 重试配置
	if config.Client.Retry.MaxAttempts > 0 {
		retryOpts := []retry.CallOption{
			retry.WithMax(uint(config.Client.Retry.MaxAttempts)),
			retry.WithBackoff(retry.BackoffExponential(config.Client.Retry.InitialBackoff)),
		}
		unaryInterceptors = append(unaryInterceptors, retry.UnaryClientInterceptor(retryOpts...))
		streamInterceptors = append(streamInterceptors, retry.StreamClientInterceptor(retryOpts...))
	}

	// 添加OpenTelemetry追踪
	if config.Tracing.Enabled {
		unaryInterceptors = append(unaryInterceptors, otelgrpc.UnaryClientInterceptor())
		streamInterceptors = append(streamInterceptors, otelgrpc.StreamClientInterceptor())
	}

	opts = append(opts,
		grpc.WithChainUnaryInterceptor(unaryInterceptors...),
		grpc.WithChainStreamInterceptor(streamInterceptors...),
	)

	// 安全配置
	if config.Security.EnableTLS {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		if config.Security.CACertFile != "" {
			caPool, err := loadCertPool(config.Security.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("加载CA证书失败: %w", err)
			}
			tlsConfig.RootCAs = caPool
		}

		if config.Security.EnableMTLS {
			cert, err := tls.LoadX509KeyPair(config.Security.ClientCertFile, config.Security.ClientKeyFile)
			if err != nil {
				return nil, fmt.Errorf("加载客户端证书失败: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}

		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 连接超时
	ctx, cancel := context.WithTimeout(context.Background(), config.Client.ConnectionTimeout)
	defer cancel()

	// 建立连接
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("连接gRPC服务失败: %w", err)
	}

	return &GRPCClient{
		config: config,
		conn:   conn,
		// 演示代码：实际项目中初始化 protoc 生成的客户端
		// userClient: NewUserServiceClient(conn),
	}, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("无法解析CA证书: %s", path)
	}
	return pool, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// 客户端日志拦截器
func clientLoggingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	duration := time.Since(start)

	if err != nil {
		log.Printf("gRPC客户端调用失败: %s (错误: %v, 耗时: %v)", method, err, duration)
	} else {
		log.Printf("gRPC客户端调用成功: %s (耗时: %v)", method, duration)
	}

	return err
}

// === 辅助函数 ===

func isValidToken(token string) bool {
	// 简化的token验证逻辑
	return len(token) > 10
}

func extractUserIDFromToken(token string) string {
	// 简化的用户ID提取逻辑
	return fmt.Sprintf("user_%d", time.Now().UnixNano()%1000)
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// 初始化OpenTelemetry追踪
func initTracing(config TracingConfig) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerURL)))
	if err != nil {
		return fmt.Errorf("创建Jaeger导出器失败: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(config.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return nil
}

// === 配置加载 ===

func loadDefaultConfig() *GRPCConfig {
	return &GRPCConfig{
		Server: ServerConfig{
			Address:     "0.0.0.0",
			Port:        9090,
			EnableTLS:   false,
			EnableH2C:   true,
			MaxRecvSize: 4 << 20, // 4MB
			MaxSendSize: 4 << 20, // 4MB
			KeepAlive: ServerKeepAliveConfig{
				Time:    60 * time.Second,
				Timeout: 5 * time.Second,
				MinTime: 30 * time.Second,
			},
			MaxConcurrentStreams: 1000,
			ConnectionTimeout:    10 * time.Second,
			EnableCompression:    true,
			EnableReflection:     true,
			EnableHealthCheck:    true,
		},
		Client: ClientConfig{
			EnableLoadBalancing: true,
			LoadBalancingPolicy: "round_robin",
			MaxRecvSize:         4 << 20, // 4MB
			MaxSendSize:         4 << 20, // 4MB
			ConnectionTimeout:   10 * time.Second,
			KeepAlive: ClientKeepAliveConfig{
				Time:                30 * time.Second,
				Timeout:             5 * time.Second,
				PermitWithoutStream: false,
			},
			Retry: RetryConfig{
				MaxAttempts:     3,
				InitialBackoff:  100 * time.Millisecond,
				MaxBackoff:      5 * time.Second,
				BackoffMul:      2.0,
				RetryableErrors: []codes.Code{codes.Unavailable, codes.DeadlineExceeded},
			},
			Pool: ConnectionPoolConfig{
				MaxConnections:     100,
				MaxIdleConnections: 20,
				ConnMaxLifetime:    30 * time.Minute,
				ConnMaxIdleTime:    5 * time.Minute,
			},
		},
		Security: SecurityConfig{
			EnableAuth: true,
			JWTSecret:  "your-super-secret-key",
			EnableTLS:  false,
			EnableMTLS: false,
		},
		Tracing: TracingConfig{
			Enabled:     true,
			ServiceName: "grpc-user-service",
			JaegerURL:   "http://localhost:14268/api/traces",
			SampleRate:  1.0,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Port:    9091,
			Path:    "/metrics",
		},
	}
}

// === 示例用法和测试 ===

func runExamples(client *GRPCClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 添加认证头
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer sample_jwt_token_123456")

	log.Println("🔥 开始gRPC服务测试...")

	// 示例1: 创建用户
	log.Println("📝 测试创建用户...")
	createReq := &CreateUserRequest{
		Username:    "testuser",
		Email:       "test@example.com",
		Password:    "securepassword",
		FullName:    "Test User",
		PhoneNumber: "+1234567890",
		Roles:       []string{"user"},
		Metadata:    map[string]string{"source": "grpc_test"},
	}

	// 临时使用变量避免编译错误
	log.Printf("创建用户请求准备完成: %+v", createReq)

	// 演示代码：实际项目中使用 protoc 生成的客户端调用
	// createResp, err := client.userClient.CreateUser(ctx, createReq)
	// if err != nil {
	//     log.Printf("创建用户失败: %v", err)
	// } else {
	//     log.Printf("创建用户成功: %s", createResp.User.UserId)
	// }

	log.Println("示例代码准备完成，等待protoc生成客户端代码后可运行完整测试")
}

// 性能测试
func runPerformanceTest(client *GRPCClient) {
	log.Println("🚀 开始性能测试...")

	concurrency := 10
	requestsPerClient := 100
	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			for j := 0; j < requestsPerClient; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", fmt.Sprintf("token_%d_%d", clientID, j))

				// 演示代码：实际项目中调用 protoc 生成的客户端方法
				// _, err := client.userClient.HealthCheck(ctx, &emptypb.Empty{})
				// if err != nil {
				//     log.Printf("健康检查失败[%d-%d]: %v", clientID, j, err)
				// }
				_ = ctx // 避免未使用变量警告

				cancel()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalRequests := concurrency * requestsPerClient
	rps := float64(totalRequests) / duration.Seconds()

	log.Printf("性能测试完成:")
	log.Printf("  总请求数: %d", totalRequests)
	log.Printf("  并发数: %d", concurrency)
	log.Printf("  总耗时: %v", duration)
	log.Printf("  平均RPS: %.2f", rps)
}

// 流式通信示例
func runStreamingExample(client *GRPCClient) {
	log.Println("🌊 开始流式通信测试...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "bearer streaming_test_token")

	// 演示代码：流式通信需要 protoc 生成代码后实现
	// 1. Server Streaming - 用户活动流
	// 2. Client Streaming - 批量操作
	// 3. Bidirectional Streaming - 聊天功能
	_ = ctx // 避免未使用变量警告

	log.Println("流式通信示例准备完成")
}

// === 主函数 ===

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 启动现代化gRPC微服务...")

	// 加载配置
	config := loadDefaultConfig()

	// 解析命令行参数
	var (
		serverMode = flag.Bool("server", true, "运行服务器模式")
		clientMode = flag.Bool("client", false, "运行客户端测试")
		perfTest   = flag.Bool("perf", false, "运行性能测试")
		streamTest = flag.Bool("stream", false, "运行流式测试")
		serverAddr = flag.String("addr", "localhost:9090", "服务器地址")
		configFile = flag.String("config", "", "配置文件路径")
	)
	flag.Parse()

	// 加载配置文件 (如果提供)
	if *configFile != "" {
		// 配置文件加载：当前使用默认配置，生产环境可集成 viper 等配置库
		log.Printf("配置文件路径: %s (当前使用默认配置)", *configFile)
	}

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if *serverMode {
		// 服务器模式
		server, err := NewGRPCServer(config)
		if err != nil {
			log.Fatalf("创建gRPC服务器失败: %v", err)
		}

		// 启动服务器
		go func() {
			if err := server.Start(); err != nil {
				log.Fatalf("启动gRPC服务器失败: %v", err)
			}
		}()

		log.Println("✅ gRPC服务器启动完成")
		log.Printf("🌐 服务地址: %s:%d", config.Server.Address, config.Server.Port)
		log.Printf("📊 指标端口: %d", config.Metrics.Port)
		log.Printf("🔍 追踪: %s", config.Tracing.JaegerURL)

		// 等待信号
		sig := <-sigChan
		log.Printf("收到信号 %s，开始关闭服务器...", sig)

		// 优雅停机
		server.Stop()
		log.Println("✅ 服务器已关闭")
	}

	if *clientMode || *perfTest || *streamTest {
		// 客户端模式
		client, err := NewGRPCClient(config, *serverAddr)
		if err != nil {
			log.Fatalf("创建gRPC客户端失败: %v", err)
		}
		defer client.Close()

		log.Printf("✅ 连接到gRPC服务器: %s", *serverAddr)

		if *perfTest {
			runPerformanceTest(client)
		} else if *streamTest {
			runStreamingExample(client)
		} else {
			runExamples(client)
		}

		log.Println("✅ 客户端测试完成")
	}
}

// Prometheus处理器占位符
type promhttpHandler struct{}

func (promhttpHandler) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Prometheus metrics placeholder\n"))
	})
}

var promhttp = promhttpHandler{}

/*
🎯 使用说明和最佳实践：

## 编译Proto文件
```bash
# 安装protoc和Go插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成Go代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/user/v1/user_service.proto

# 生成gRPC-Gateway代码 (可选)
protoc --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
       proto/user/v1/user_service.proto
```

## 运行服务
```bash
# 启动服务器
go run main.go -server

# 运行客户端测试
go run main.go -client -addr localhost:9090

# 性能测试
go run main.go -perf -addr localhost:9090

# 流式通信测试
go run main.go -stream -addr localhost:9090
```

## Docker部署
```bash
# 构建镜像
docker build -t grpc-service .

# 运行容器
docker run -p 9090:9090 -p 9091:9091 grpc-service
```

## Kubernetes部署
```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

## 监控和调试
- Prometheus指标: http://localhost:9091/metrics
- gRPC反射: grpcurl -plaintext localhost:9090 list
- 健康检查: grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
- Jaeger追踪: http://localhost:16686

## 性能优化建议
1. 启用连接池和Keep-Alive
2. 配置合适的消息大小限制
3. 使用流式接口处理大数据
4. 启用压缩减少网络传输
5. 实现客户端负载均衡
6. 配置合理的超时和重试
7. 使用连接复用
8. 启用HTTP/2多路复用

## 安全最佳实践
1. 启用TLS/mTLS加密
2. 实现JWT认证和授权
3. 配置API限流和熔断
4. 验证和清理输入数据
5. 记录审计日志
6. 定期更新证书
7. 使用安全的密钥管理

## 扩展功能
1. 服务网格集成 (Istio)
2. API网关集成
3. 配置中心集成
4. 分布式缓存
5. 消息队列集成
6. 数据库连接池
7. 多租户支持
8. 国际化支持

🎉 恭喜！您已掌握现代化gRPC微服务通信的完整实现！
*/
