package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// =============================================================================
// 安全随机数生成辅助函数
// =============================================================================

// secureRandomInt 生成安全的随机整数 [0, max)
func secureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// 如果加密随机数生成失败，使用时间作为fallback（虽然不够安全，但不会panic）
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano() % int64(max)
		// 检查是否在int范围内
		if fallback > int64(^uint(0)>>1) {
			fallback = fallback % int64(^uint(0)>>1)
		}
		return int(fallback)
	}
	// G115安全修复：检查int64到int的安全转换
	result := n.Int64()
	if result > int64(^uint(0)>>1) {
		result = result % int64(max)
	}
	return int(result)
}

// secureRandomFloat32 生成安全的随机浮点数 [0.0, 1.0)
func secureRandomFloat32() float32 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		// 如果加密随机数生成失败，使用时间作为fallback
		return float32(time.Now().UnixNano()%1000) / 1000.0
	}
	return float32(n.Int64()) / float32(1<<24)
}

// =============================================================================
// 1. Context 基础概念
// =============================================================================

/*
Context 是 Go 语言并发编程中用于控制 goroutine 生命周期的重要机制：

核心功能：
1. 取消信号传播：向下游 goroutine 传递取消信号
2. 超时控制：设置操作的最大执行时间
3. 截止时间：设置操作必须在指定时间前完成
4. 值传递：在调用链中传递请求范围的值

主要类型：
1. context.Background()：根 context，永不取消
2. context.TODO()：当不确定使用哪个 context 时的占位符
3. context.WithCancel()：可手动取消的 context
4. context.WithTimeout()：带超时的 context
5. context.WithDeadline()：带截止时间的 context
6. context.WithValue()：携带值的 context

使用原则：
1. Context 应该作为函数的第一个参数传递
2. 不要将 Context 存储在结构体中
3. 不要传递 nil context，使用 context.TODO()
4. Context 是并发安全的，可以被多个 goroutine 同时使用
5. Context 的取消是级联的，父 context 取消时子 context 也会取消
*/

// =============================================================================
// 2. 基础 Context 操作
// =============================================================================

// worker 基础工作函数
func worker(ctx context.Context, id int) {
	fmt.Printf("Worker %d 开始工作\n", id)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d 收到取消信号: %v\n", id, ctx.Err())
			return
		default:
			fmt.Printf("Worker %d 正在工作...\n", id)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func demonstrateBasicContext() {
	fmt.Println("=== 1. 基础 Context 操作 ===")

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())

	// 启动多个工作者
	var wg sync.WaitGroup
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(ctx, id)
		}(i)
	}

	// 让工作者运行一段时间
	fmt.Println("让工作者运行 2 秒...")
	time.Sleep(2 * time.Second)

	// 取消所有工作者
	fmt.Println("取消所有工作者...")
	cancel()

	// 等待所有工作者完成
	wg.Wait()
	fmt.Println("所有工作者已停止")
	fmt.Println()
}

// =============================================================================
// 3. 超时控制
// =============================================================================

// slowTask 模拟慢任务
func slowTask(ctx context.Context, taskID int, duration time.Duration) error {
	fmt.Printf("任务 %d 开始执行，预计耗时: %v\n", taskID, duration)

	select {
	case <-time.After(duration):
		fmt.Printf("任务 %d 完成\n", taskID)
		return nil
	case <-ctx.Done():
		fmt.Printf("任务 %d 被取消: %v\n", taskID, ctx.Err())
		return ctx.Err()
	}
}

// httpRequest 模拟HTTP请求
func httpRequest(ctx context.Context, url string) (string, error) {
	fmt.Printf("开始请求: %s\n", url)

	// 模拟网络延迟
	delay := time.Duration(secureRandomInt(3000)) * time.Millisecond

	select {
	case <-time.After(delay):
		response := fmt.Sprintf("Response from %s (延迟: %v)", url, delay)
		fmt.Printf("请求完成: %s\n", response)
		return response, nil
	case <-ctx.Done():
		fmt.Printf("请求被取消: %s, 原因: %v\n", url, ctx.Err())
		return "", ctx.Err()
	}
}

func demonstrateTimeout() {
	fmt.Println("=== 2. 超时控制 ===")

	// 测试任务超时
	fmt.Println("测试任务超时:")

	// 设置 1.5 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup

	// 启动多个任务
	tasks := []struct {
		id       int
		duration time.Duration
	}{
		{1, 1000 * time.Millisecond}, // 应该完成
		{2, 2000 * time.Millisecond}, // 应该超时
		{3, 800 * time.Millisecond},  // 应该完成
	}

	for _, task := range tasks {
		wg.Add(1)
		go func(id int, duration time.Duration) {
			defer wg.Done()
			slowTask(ctx, id, duration)
		}(task.id, task.duration)
	}

	wg.Wait()

	// 测试HTTP请求超时
	fmt.Println("\n测试HTTP请求超时:")

	// 设置 2 秒超时
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer httpCancel()

	urls := []string{
		"https://api.example.com/users",
		"https://api.example.com/orders",
		"https://api.example.com/products",
	}

	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			httpRequest(httpCtx, u)
		}(url)
	}

	wg.Wait()
	fmt.Println()
}

// =============================================================================
// 4. 截止时间控制
// =============================================================================

// batchJob 批处理任务
func batchJob(ctx context.Context, jobID int, itemCount int) error {
	fmt.Printf("批处理任务 %d 开始，处理 %d 个项目\n", jobID, itemCount)

	for i := 1; i <= itemCount; i++ {
		select {
		case <-ctx.Done():
			fmt.Printf("批处理任务 %d 在第 %d 个项目时被取消: %v\n", jobID, i, ctx.Err())
			return ctx.Err()
		default:
			fmt.Printf("批处理任务 %d 处理项目 %d/%d\n", jobID, i, itemCount)
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Printf("批处理任务 %d 完成\n", jobID)
	return nil
}

func demonstrateDeadline() {
	fmt.Println("=== 3. 截止时间控制 ===")

	// 设置截止时间为 3 秒后
	deadline := time.Now().Add(3 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	fmt.Printf("设置截止时间: %v\n", deadline.Format("15:04:05"))

	var wg sync.WaitGroup

	// 启动多个批处理任务
	jobs := []struct {
		id        int
		itemCount int
	}{
		{1, 8},  // 应该完成 (8 * 200ms = 1.6s)
		{2, 20}, // 应该被截止时间中断
		{3, 5},  // 应该完成 (5 * 200ms = 1s)
	}

	for _, job := range jobs {
		wg.Add(1)
		go func(id, count int) {
			defer wg.Done()
			batchJob(ctx, id, count)
		}(job.id, job.itemCount)
	}

	wg.Wait()

	// 检查 context 状态
	if deadline, ok := ctx.Deadline(); ok {
		fmt.Printf("截止时间: %v\n", deadline.Format("15:04:05"))
		fmt.Printf("当前时间: %v\n", time.Now().Format("15:04:05"))
	}

	fmt.Printf("Context 错误: %v\n", ctx.Err())
	fmt.Println()
}

// =============================================================================
// 5. 值传递
// =============================================================================

// 定义 context key 类型以避免冲突
type contextKey string

const (
	userIDKey    contextKey = "userID"
	requestIDKey contextKey = "requestID"
	traceIDKey   contextKey = "traceID"
)

// getUserID 从 context 中获取用户ID
func getUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

// getRequestID 从 context 中获取请求ID
func getRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok
}

// getTraceID 从 context 中获取追踪ID
func getTraceID(ctx context.Context) (string, bool) {
	traceID, ok := ctx.Value(traceIDKey).(string)
	return traceID, ok
}

// serviceA 服务A
func serviceA(ctx context.Context, data string) string {
	userID, _ := getUserID(ctx)
	requestID, _ := getRequestID(ctx)
	traceID, _ := getTraceID(ctx)

	fmt.Printf("服务A [用户:%s, 请求:%s, 追踪:%s] 处理数据: %s\n",
		userID, requestID, traceID, data)

	// 调用服务B
	result := serviceB(ctx, "从服务A传递的数据")

	return fmt.Sprintf("服务A结果: %s", result)
}

// serviceB 服务B
func serviceB(ctx context.Context, data string) string {
	userID, _ := getUserID(ctx)
	requestID, _ := getRequestID(ctx)
	traceID, _ := getTraceID(ctx)

	fmt.Printf("服务B [用户:%s, 请求:%s, 追踪:%s] 处理数据: %s\n",
		userID, requestID, traceID, data)

	// 调用服务C
	result := serviceC(ctx, "从服务B传递的数据")

	return fmt.Sprintf("服务B结果: %s", result)
}

// serviceC 服务C
func serviceC(ctx context.Context, data string) string {
	userID, _ := getUserID(ctx)
	requestID, _ := getRequestID(ctx)
	traceID, _ := getTraceID(ctx)

	fmt.Printf("服务C [用户:%s, 请求:%s, 追踪:%s] 处理数据: %s\n",
		userID, requestID, traceID, data)

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	return "服务C处理完成"
}

func demonstrateContextValues() {
	fmt.Println("=== 4. Context 值传递 ===")

	// 创建带值的 context
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, "user-12345")
	ctx = context.WithValue(ctx, requestIDKey, "req-abcde")
	ctx = context.WithValue(ctx, traceIDKey, "trace-xyz789")

	fmt.Println("在调用链中传递 context 值:")

	// 模拟HTTP请求处理
	result := serviceA(ctx, "原始请求数据")
	fmt.Printf("最终结果: %s\n", result)

	// 演示值的继承
	fmt.Println("\n演示 context 值的继承:")

	// 子 context 继承父 context 的值
	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if userID, ok := getUserID(childCtx); ok {
		fmt.Printf("子 context 继承的用户ID: %s\n", userID)
	}

	// 子 context 可以添加新的值
	childCtx = context.WithValue(childCtx, contextKey("sessionID"), "session-999")

	if sessionID, ok := childCtx.Value(contextKey("sessionID")).(string); ok {
		fmt.Printf("子 context 的新值: %s\n", sessionID)
	}

	fmt.Println()
}

// =============================================================================
// 6. 实际应用：HTTP 服务器
// =============================================================================

// User 用户结构体
type User struct {
	ID   string
	Name string
	Role string
}

// Request 请求结构体
type Request struct {
	ID   string
	Type string
	Data map[string]interface{}
}

// HTTPServer 模拟HTTP服务器
type HTTPServer struct {
	timeout time.Duration
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(timeout time.Duration) *HTTPServer {
	return &HTTPServer{timeout: timeout}
}

// handleRequest 处理HTTP请求
func (s *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// 为请求创建带超时的 context
	ctx, cancel := context.WithTimeout(r.Context(), s.timeout)
	defer cancel()

	// 生成请求ID
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	ctx = context.WithValue(ctx, requestIDKey, requestID)

	// 模拟用户认证
	user := &User{
		ID:   "user-123",
		Name: "张三",
		Role: "admin",
	}
	ctx = context.WithValue(ctx, "user", user)

	fmt.Printf("处理请求 %s，用户: %s\n", requestID, user.Name)

	// 处理请求
	result, err := s.processRequest(ctx, &Request{
		ID:   requestID,
		Type: "user_query",
		Data: map[string]interface{}{
			"query": "获取用户信息",
		},
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "请求处理成功: %s", result)
}

// processRequest 处理业务逻辑
func (s *HTTPServer) processRequest(ctx context.Context, req *Request) (string, error) {
	requestID, _ := getRequestID(ctx)
	user := ctx.Value("user").(*User)

	fmt.Printf("开始处理请求 %s，类型: %s，用户: %s\n", requestID, req.Type, user.Name)

	// 调用数据库服务
	dbResult, err := s.queryDatabase(ctx, req.Data["query"].(string))
	if err != nil {
		return "", fmt.Errorf("数据库查询失败: %w", err)
	}

	// 调用缓存服务
	cacheResult, err := s.queryCache(ctx, "user:"+user.ID)
	if err != nil {
		fmt.Printf("缓存查询失败: %v，使用数据库结果\n", err)
		cacheResult = "缓存未命中"
	}

	result := fmt.Sprintf("数据库: %s, 缓存: %s", dbResult, cacheResult)
	fmt.Printf("请求 %s 处理完成\n", requestID)

	return result, nil
}

// queryDatabase 模拟数据库查询
func (s *HTTPServer) queryDatabase(ctx context.Context, query string) (string, error) {
	requestID, _ := getRequestID(ctx)
	fmt.Printf("数据库查询开始 [请求: %s]: %s\n", requestID, query)

	// 模拟数据库查询时间
	select {
	case <-time.After(800 * time.Millisecond):
		result := "数据库查询结果"
		fmt.Printf("数据库查询完成 [请求: %s]: %s\n", requestID, result)
		return result, nil
	case <-ctx.Done():
		fmt.Printf("数据库查询被取消 [请求: %s]: %v\n", requestID, ctx.Err())
		return "", ctx.Err()
	}
}

// queryCache 模拟缓存查询
func (s *HTTPServer) queryCache(ctx context.Context, key string) (string, error) {
	requestID, _ := getRequestID(ctx)
	fmt.Printf("缓存查询开始 [请求: %s]: %s\n", requestID, key)

	// 模拟缓存查询时间
	select {
	case <-time.After(100 * time.Millisecond):
		// 模拟缓存命中率
		if secureRandomFloat32() < 0.7 { // 70% 命中率
			result := "缓存命中数据"
			fmt.Printf("缓存查询完成 [请求: %s]: %s\n", requestID, result)
			return result, nil
		} else {
			fmt.Printf("缓存未命中 [请求: %s]\n", requestID)
			return "", fmt.Errorf("缓存未命中")
		}
	case <-ctx.Done():
		fmt.Printf("缓存查询被取消 [请求: %s]: %v\n", requestID, ctx.Err())
		return "", ctx.Err()
	}
}

func demonstrateHTTPServer() {
	fmt.Println("=== 5. 实际应用：HTTP服务器 ===")

	server := NewHTTPServer(2 * time.Second)

	// 模拟多个并发请求
	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(reqNum int) {
			defer wg.Done()

			fmt.Printf("\n--- 模拟请求 %d ---\n", reqNum)

			// 创建模拟的 HTTP 请求
			ctx := context.Background()

			// 模拟不同的请求处理时间
			if reqNum == 2 {
				// 第二个请求设置较短的超时，可能会超时
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
				defer cancel()
			}

			result, err := server.processRequest(ctx, &Request{
				ID:   fmt.Sprintf("req-%d", reqNum),
				Type: "user_query",
				Data: map[string]interface{}{
					"query": fmt.Sprintf("查询用户%d的信息", reqNum),
				},
			})

			if err != nil {
				fmt.Printf("请求 %d 失败: %v\n", reqNum, err)
			} else {
				fmt.Printf("请求 %d 成功: %s\n", reqNum, result)
			}
		}(i)

		// 错开请求启动时间
		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println()
}

// =============================================================================
// 7. Context 链式操作
// =============================================================================

// Pipeline 处理管道
type Pipeline struct {
	name string
}

// NewPipeline 创建处理管道
func NewPipeline(name string) *Pipeline {
	return &Pipeline{name: name}
}

// Stage 处理阶段
func (p *Pipeline) Stage(ctx context.Context, stageNum int, data string) (string, error) {
	stageName := fmt.Sprintf("%s-Stage%d", p.name, stageNum)

	// 为每个阶段创建子 context
	stageCtx := context.WithValue(ctx, contextKey("stage"), stageName)

	if requestID, ok := getRequestID(stageCtx); ok {
		fmt.Printf("[%s] 开始处理 [请求: %s]: %s\n", stageName, requestID, data)
	}

	// 模拟处理时间
	processingTime := time.Duration(secureRandomInt(500)+200) * time.Millisecond

	select {
	case <-time.After(processingTime):
		result := fmt.Sprintf("%s处理后的%s", stageName, data)
		if requestID, ok := getRequestID(stageCtx); ok {
			fmt.Printf("[%s] 处理完成 [请求: %s]: %s (耗时: %v)\n",
				stageName, requestID, result, processingTime)
		}
		return result, nil
	case <-stageCtx.Done():
		if requestID, ok := getRequestID(stageCtx); ok {
			fmt.Printf("[%s] 处理被取消 [请求: %s]: %v\n",
				stageName, requestID, stageCtx.Err())
		}
		return "", stageCtx.Err()
	}
}

// Process 处理整个管道
func (p *Pipeline) Process(ctx context.Context, data string) (string, error) {
	// 阶段1
	result1, err := p.Stage(ctx, 1, data)
	if err != nil {
		return "", err
	}

	// 阶段2
	result2, err := p.Stage(ctx, 2, result1)
	if err != nil {
		return "", err
	}

	// 阶段3
	result3, err := p.Stage(ctx, 3, result2)
	if err != nil {
		return "", err
	}

	return result3, nil
}

func demonstrateContextChaining() {
	fmt.Println("=== 6. Context 链式操作 ===")

	var wg sync.WaitGroup

	// 创建不同超时时间的请求
	requests := []struct {
		id      string
		timeout time.Duration
		data    string
	}{
		{"fast-req", 2 * time.Second, "快速数据"},
		{"slow-req", 800 * time.Millisecond, "慢速数据"},
		{"normal-req", 1500 * time.Millisecond, "正常数据"},
	}

	for _, req := range requests {
		wg.Add(1)
		go func(requestID string, timeout time.Duration, data string) {
			defer wg.Done()

			fmt.Printf("\n--- 处理请求: %s (超时: %v) ---\n", requestID, timeout)

			// 创建带超时的 context
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// 添加请求ID
			ctx = context.WithValue(ctx, requestIDKey, requestID)

			// 创建处理管道
			pipeline := NewPipeline("DataPipeline")

			// 处理数据
			result, err := pipeline.Process(ctx, data)

			if err != nil {
				fmt.Printf("请求 %s 失败: %v\n", requestID, err)
			} else {
				fmt.Printf("请求 %s 成功: %s\n", requestID, result)
			}
		}(req.id, req.timeout, req.data)

		// 错开请求时间
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
	fmt.Println()
}

// =============================================================================
// 8. Context 最佳实践
// =============================================================================

func demonstrateContextBestPractices() {
	fmt.Println("=== 7. Context 最佳实践 ===")

	fmt.Println("1. Context 传递规则:")
	fmt.Println("   ✓ Context 应该作为函数的第一个参数")
	fmt.Println("   ✓ 参数名通常使用 ctx")
	fmt.Println("   ✗ 不要将 Context 存储在结构体中")
	fmt.Println("   ✗ 不要传递 nil context")

	fmt.Println("\n2. 取消和超时:")
	fmt.Println("   ✓ 始终调用 cancel 函数（使用 defer）")
	fmt.Println("   ✓ 检查 ctx.Done() 以响应取消信号")
	fmt.Println("   ✓ 使用 select 语句处理取消和正常操作")
	fmt.Println("   ✓ 设置合理的超时时间")

	fmt.Println("\n3. 值传递:")
	fmt.Println("   ✓ 只传递请求范围的数据")
	fmt.Println("   ✓ 使用自定义类型作为 key 避免冲突")
	fmt.Println("   ✗ 不要传递可选参数")
	fmt.Println("   ✗ 不要传递大量数据")

	fmt.Println("\n4. 错误处理:")
	fmt.Println("   ✓ 检查 ctx.Err() 获取取消原因")
	fmt.Println("   ✓ 区分用户取消和超时")
	fmt.Println("   ✓ 适当记录取消事件")

	// 演示正确的 context 使用模式
	fmt.Println("\n正确的使用示例:")

	// 好的做法：函数签名
	goodFunction := func(ctx context.Context, userID string) error {
		select {
		case <-time.After(100 * time.Millisecond):
			fmt.Printf("处理用户 %s 完成\n", userID)
			return nil
		case <-ctx.Done():
			fmt.Printf("处理用户 %s 被取消: %v\n", userID, ctx.Err())
			return ctx.Err()
		}
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel() // 好的做法：使用 defer 确保调用 cancel

	err := goodFunction(ctx, "user123")
	if err != nil {
		fmt.Printf("函数执行错误: %v\n", err)
	}

	fmt.Println("\n常见陷阱和避免方法:")
	fmt.Println("1. 忘记调用 cancel 函数 → 使用 defer cancel()")
	fmt.Println("2. Context 值过度使用 → 只传递请求范围的必要数据")
	fmt.Println("3. 没有检查取消信号 → 在长时间运行的操作中定期检查 ctx.Done()")
	fmt.Println("4. 超时时间设置不当 → 根据实际业务需要设置合理的超时")

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - Context 上下文管理")
	fmt.Println("=================================")

	// crypto/rand 不需要种子设置，已经是加密安全的

	demonstrateBasicContext()
	demonstrateTimeout()
	demonstrateDeadline()
	demonstrateContextValues()
	demonstrateHTTPServer()
	demonstrateContextChaining()
	demonstrateContextBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个支持优雅关闭的 HTTP 服务器")
	fmt.Println("2. 创建一个分布式任务调度系统")
	fmt.Println("3. 实现一个支持链路追踪的微服务框架")
	fmt.Println("4. 编写一个并发文件下载器")
	fmt.Println("5. 创建一个支持熔断的服务调用客户端")
	fmt.Println("6. 实现一个请求限流中间件")
	fmt.Println("\n请在此基础上练习更多 Context 的使用场景！")
}
