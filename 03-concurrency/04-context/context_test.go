package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// Context 值获取函数测试
// =============================================================================

// TestGetUserID 测试获取用户ID
func TestGetUserID(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantValue string
		wantOK    bool
	}{
		{
			name:      "存在用户ID",
			ctx:       context.WithValue(context.Background(), userIDKey, "user-123"),
			wantValue: "user-123",
			wantOK:    true,
		},
		{
			name:      "不存在用户ID",
			ctx:       context.Background(),
			wantValue: "",
			wantOK:    false,
		},
		{
			name:      "错误类型的值",
			ctx:       context.WithValue(context.Background(), userIDKey, 12345),
			wantValue: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := getUserID(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("getUserID() ok = %v, want %v", ok, tt.wantOK)
			}
			if value != tt.wantValue {
				t.Errorf("getUserID() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

// TestGetRequestID 测试获取请求ID
func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantValue string
		wantOK    bool
	}{
		{
			name:      "存在请求ID",
			ctx:       context.WithValue(context.Background(), requestIDKey, "req-abc"),
			wantValue: "req-abc",
			wantOK:    true,
		},
		{
			name:      "不存在请求ID",
			ctx:       context.Background(),
			wantValue: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := getRequestID(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("getRequestID() ok = %v, want %v", ok, tt.wantOK)
			}
			if value != tt.wantValue {
				t.Errorf("getRequestID() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

// TestGetTraceID 测试获取追踪ID
func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		wantValue string
		wantOK    bool
	}{
		{
			name:      "存在追踪ID",
			ctx:       context.WithValue(context.Background(), traceIDKey, "trace-xyz"),
			wantValue: "trace-xyz",
			wantOK:    true,
		},
		{
			name:      "不存在追踪ID",
			ctx:       context.Background(),
			wantValue: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := getTraceID(tt.ctx)
			if ok != tt.wantOK {
				t.Errorf("getTraceID() ok = %v, want %v", ok, tt.wantOK)
			}
			if value != tt.wantValue {
				t.Errorf("getTraceID() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

// TestContextValueInheritance 测试 context 值继承
func TestContextValueInheritance(t *testing.T) {
	// 创建带值的父 context
	parentCtx := context.Background()
	parentCtx = context.WithValue(parentCtx, userIDKey, "user-parent")
	parentCtx = context.WithValue(parentCtx, requestIDKey, "req-parent")

	// 创建子 context（带超时）
	childCtx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	// 子 context 应该继承父 context 的值
	userID, ok := getUserID(childCtx)
	if !ok || userID != "user-parent" {
		t.Errorf("子 context 应该继承父 context 的 userID")
	}

	requestID, ok := getRequestID(childCtx)
	if !ok || requestID != "req-parent" {
		t.Errorf("子 context 应该继承父 context 的 requestID")
	}

	// 子 context 可以添加新值
	childCtx = context.WithValue(childCtx, traceIDKey, "trace-child")
	traceID, ok := getTraceID(childCtx)
	if !ok || traceID != "trace-child" {
		t.Errorf("子 context 应该能添加新值")
	}

	// 父 context 不应该有子 context 添加的值
	_, ok = getTraceID(parentCtx)
	if ok {
		t.Errorf("父 context 不应该有子 context 添加的值")
	}
}

// =============================================================================
// slowTask 测试
// =============================================================================

// TestSlowTaskCompletion 测试慢任务正常完成
func TestSlowTaskCompletion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := slowTask(ctx, 1, 100*time.Millisecond)
	if err != nil {
		t.Errorf("任务应该正常完成, 但返回错误: %v", err)
	}
}

// TestSlowTaskTimeout 测试慢任务超时
func TestSlowTaskTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := slowTask(ctx, 1, 500*time.Millisecond)
	if err == nil {
		t.Error("任务应该因超时而返回错误")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("期望 DeadlineExceeded 错误, 实际: %v", err)
	}
}

// TestSlowTaskCancellation 测试慢任务取消
func TestSlowTaskCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// 在另一个 goroutine 中取消
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := slowTask(ctx, 1, 500*time.Millisecond)
	if err == nil {
		t.Error("任务应该因取消而返回错误")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("期望 Canceled 错误, 实际: %v", err)
	}
}

// =============================================================================
// httpRequest 测试
// =============================================================================

// TestHttpRequestTimeout 测试 HTTP 请求超时
func TestHttpRequestTimeout(t *testing.T) {
	// 使用非常短的超时来确保超时
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := httpRequest(ctx, "https://example.com/test")
	if err == nil {
		t.Error("请求应该因超时而返回错误")
	}
}

// =============================================================================
// batchJob 测试
// =============================================================================

// TestBatchJobCompletion 测试批处理任务完成
func TestBatchJobCompletion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := batchJob(ctx, 1, 3) // 3个项目，每个200ms，总共600ms
	if err != nil {
		t.Errorf("批处理任务应该完成, 但返回错误: %v", err)
	}
}

// TestBatchJobCancellation 测试批处理任务取消
func TestBatchJobCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err := batchJob(ctx, 1, 10) // 10个项目，每个200ms，总共2秒，会超时
	if err == nil {
		t.Error("批处理任务应该因超时而返回错误")
	}
}

// =============================================================================
// Pipeline 测试
// =============================================================================

// TestPipelineStage 测试管道阶段
func TestPipelineStage(t *testing.T) {
	pipeline := NewPipeline("TestPipeline")

	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "test-req")

	// 使用足够长的超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := pipeline.Stage(ctx, 1, "测试数据")
	if err != nil {
		t.Errorf("管道阶段应该成功, 但返回错误: %v", err)
	}

	if result == "" {
		t.Error("管道阶段应该返回非空结果")
	}
}

// TestPipelineProcess 测试完整管道处理
func TestPipelineProcess(t *testing.T) {
	pipeline := NewPipeline("TestPipeline")

	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "test-req")

	// 使用足够长的超时（3个阶段，每个最多700ms）
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := pipeline.Process(ctx, "初始数据")
	if err != nil {
		t.Errorf("管道处理应该成功, 但返回错误: %v", err)
	}

	if result == "" {
		t.Error("管道处理应该返回非空结果")
	}
}

// TestPipelineTimeout 测试管道超时
func TestPipelineTimeout(t *testing.T) {
	pipeline := NewPipeline("TestPipeline")

	// 使用非常短的超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := pipeline.Process(ctx, "测试数据")
	if err == nil {
		t.Error("管道处理应该因超时而返回错误")
	}
}

// =============================================================================
// HTTPServer 测试
// =============================================================================

// TestNewHTTPServer 测试创建 HTTP 服务器
func TestNewHTTPServer(t *testing.T) {
	server := NewHTTPServer(5 * time.Second)

	if server == nil {
		t.Fatal("NewHTTPServer 返回 nil")
	}

	if server.timeout != 5*time.Second {
		t.Errorf("期望超时 5s, 实际 %v", server.timeout)
	}
}

// TestHTTPServerProcessRequest 测试处理请求
func TestHTTPServerProcessRequest(t *testing.T) {
	server := NewHTTPServer(10 * time.Second)

	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "test-req")
	ctx = context.WithValue(ctx, "user", &User{
		ID:   "user-1",
		Name: "测试用户",
		Role: "admin",
	})

	req := &Request{
		ID:   "req-1",
		Type: "test",
		Data: map[string]interface{}{
			"query": "测试查询",
		},
	}

	result, err := server.processRequest(ctx, req)
	if err != nil {
		t.Errorf("处理请求失败: %v", err)
	}

	if result == "" {
		t.Error("处理请求应该返回非空结果")
	}
}

// TestHTTPServerQueryDatabase 测试数据库查询
func TestHTTPServerQueryDatabase(t *testing.T) {
	server := NewHTTPServer(5 * time.Second)

	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "test-req")

	result, err := server.queryDatabase(ctx, "SELECT * FROM users")
	if err != nil {
		t.Errorf("数据库查询失败: %v", err)
	}

	if result == "" {
		t.Error("数据库查询应该返回非空结果")
	}
}

// TestHTTPServerQueryDatabaseTimeout 测试数据库查询超时
func TestHTTPServerQueryDatabaseTimeout(t *testing.T) {
	server := NewHTTPServer(5 * time.Second)

	// 使用非常短的超时
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	ctx = context.WithValue(ctx, requestIDKey, "test-req")

	_, err := server.queryDatabase(ctx, "SELECT * FROM users")
	if err == nil {
		t.Error("数据库查询应该因超时而返回错误")
	}
}

// =============================================================================
// 并发测试
// =============================================================================

// TestContextValuesConcurrent 测试并发访问 context 值
func TestContextValuesConcurrent(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-concurrent")
	ctx = context.WithValue(ctx, requestIDKey, "req-concurrent")
	ctx = context.WithValue(ctx, traceIDKey, "trace-concurrent")

	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 并发读取 context 值
			userID, ok := getUserID(ctx)
			if !ok || userID != "user-concurrent" {
				t.Errorf("并发读取 userID 失败")
			}

			requestID, ok := getRequestID(ctx)
			if !ok || requestID != "req-concurrent" {
				t.Errorf("并发读取 requestID 失败")
			}

			traceID, ok := getTraceID(ctx)
			if !ok || traceID != "trace-concurrent" {
				t.Errorf("并发读取 traceID 失败")
			}
		}()
	}

	wg.Wait()
}

// TestPipelineConcurrent 测试并发管道处理
func TestPipelineConcurrent(t *testing.T) {
	const numGoroutines = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			pipeline := NewPipeline("ConcurrentPipeline")

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			ctx = context.WithValue(ctx, requestIDKey, "concurrent-req")

			_, err := pipeline.Process(ctx, "并发数据")
			if err != nil {
				t.Errorf("goroutine %d 管道处理失败: %v", id, err)
			}
		}(i)
	}

	wg.Wait()
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkGetUserID 基准测试获取用户ID
func BenchmarkGetUserID(b *testing.B) {
	ctx := context.WithValue(context.Background(), userIDKey, "user-bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getUserID(ctx)
	}
}

// BenchmarkContextWithValue 基准测试创建带值的 context
func BenchmarkContextWithValue(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = context.WithValue(ctx, userIDKey, "user-bench")
	}
}

// BenchmarkContextWithCancel 基准测试创建可取消的 context
func BenchmarkContextWithCancel(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, cancel := context.WithCancel(ctx)
		cancel()
	}
}

// BenchmarkContextWithTimeout 基准测试创建带超时的 context
func BenchmarkContextWithTimeout(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, cancel := context.WithTimeout(ctx, time.Second)
		cancel()
	}
}

// BenchmarkPipelineStage 基准测试管道阶段
func BenchmarkPipelineStage(b *testing.B) {
	pipeline := NewPipeline("BenchPipeline")
	ctx := context.Background()
	ctx = context.WithValue(ctx, requestIDKey, "bench-req")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pipeline.Stage(ctx, 1, "bench-data")
		cancel()
	}
}

// =============================================================================
// 边界条件测试
// =============================================================================

// TestSlowTaskZeroDuration 测试零持续时间任务
func TestSlowTaskZeroDuration(t *testing.T) {
	ctx := context.Background()

	err := slowTask(ctx, 1, 0)
	if err != nil {
		t.Errorf("零持续时间任务应该立即完成, 但返回错误: %v", err)
	}
}

// TestBatchJobZeroItems 测试零项目批处理
func TestBatchJobZeroItems(t *testing.T) {
	ctx := context.Background()

	err := batchJob(ctx, 1, 0)
	if err != nil {
		t.Errorf("零项目批处理应该立即完成, 但返回错误: %v", err)
	}
}

// TestContextKeyType 测试 context key 类型安全
func TestContextKeyType(t *testing.T) {
	ctx := context.Background()

	// 使用字符串 key（不推荐）
	ctx = context.WithValue(ctx, "userID", "string-key-user")

	// 使用自定义类型 key
	ctx = context.WithValue(ctx, userIDKey, "typed-key-user")

	// 字符串 key 不应该与自定义类型 key 冲突
	userID, ok := getUserID(ctx)
	if !ok || userID != "typed-key-user" {
		t.Error("自定义类型 key 应该正确工作")
	}

	// 直接获取字符串 key 的值
	stringKeyValue := ctx.Value("userID")
	if stringKeyValue != "string-key-user" {
		t.Error("字符串 key 应该独立存储")
	}
}

// TestContextCancellationPropagation 测试取消信号传播
func TestContextCancellationPropagation(t *testing.T) {
	parentCtx, parentCancel := context.WithCancel(context.Background())

	childCtx1, childCancel1 := context.WithCancel(parentCtx)
	defer childCancel1()

	childCtx2, childCancel2 := context.WithCancel(parentCtx)
	defer childCancel2()

	// 取消父 context
	parentCancel()

	// 子 context 应该也被取消
	select {
	case <-childCtx1.Done():
		// 正确：子 context 被取消
	case <-time.After(100 * time.Millisecond):
		t.Error("子 context 1 应该被取消")
	}

	select {
	case <-childCtx2.Done():
		// 正确：子 context 被取消
	case <-time.After(100 * time.Millisecond):
		t.Error("子 context 2 应该被取消")
	}
}

// TestContextDeadline 测试 context 截止时间
func TestContextDeadline(t *testing.T) {
	deadline := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	gotDeadline, ok := ctx.Deadline()
	if !ok {
		t.Error("context 应该有截止时间")
	}

	// 允许1毫秒的误差
	if gotDeadline.Sub(deadline) > time.Millisecond {
		t.Errorf("截止时间不匹配: 期望 %v, 实际 %v", deadline, gotDeadline)
	}
}
