package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// =============================================================================
// 1. 限流和节流基础概念
// =============================================================================

/*
限流（Rate Limiting）和节流（Throttling）是并发系统中重要的流量控制机制：

核心概念：
1. 限流：限制单位时间内的请求数量
2. 节流：平滑处理请求，避免突发流量
3. 背压：当系统负载过高时的反压机制
4. 熔断：在系统故障时快速失败

常见算法：
1. 令牌桶（Token Bucket）：允许突发流量，平均速率控制
2. 漏桶（Leaky Bucket）：平滑输出，严格速率控制
3. 固定窗口（Fixed Window）：简单但有临界问题
4. 滑动窗口（Sliding Window）：更平滑的统计
5. 滑动日志（Sliding Log）：精确但内存消耗大

应用场景：
- API 接口限流
- 数据库连接限制
- 消息队列背压
- 网络请求控制
- 资源访问保护
*/

// =============================================================================
// 2. 令牌桶算法
// =============================================================================

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	capacity   int64      // 桶容量
	tokens     int64      // 当前令牌数
	refillRate int64      // 每秒补充令牌数
	lastRefill time.Time  // 上次补充时间
	mu         sync.Mutex // 互斥锁
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity, // 初始时桶是满的
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// refill 补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// 计算应该补充的令牌数
	tokensToAdd := int64(elapsed.Seconds()) * tb.refillRate
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (tb *TokenBucket) AllowN(n int64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// Reserve 预约令牌（如果没有足够令牌则等待）
func (tb *TokenBucket) Reserve(ctx context.Context) error {
	return tb.ReserveN(ctx, 1)
}

// ReserveN 预约N个令牌
func (tb *TokenBucket) ReserveN(ctx context.Context, n int64) error {
	for {
		if tb.AllowN(n) {
			return nil
		}

		// 计算需要等待的时间
		waitTime := time.Duration(n) * time.Second / time.Duration(tb.refillRate)
		if waitTime > time.Second {
			waitTime = time.Second // 最多等待1秒
		}

		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// GetTokens 获取当前令牌数
func (tb *TokenBucket) GetTokens() int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func demonstrateTokenBucket() {
	fmt.Println("=== 1. 令牌桶算法 ===")

	// 创建令牌桶：容量10，每秒补充2个令牌
	bucket := NewTokenBucket(10, 2)

	fmt.Printf("初始令牌数: %d\n", bucket.GetTokens())

	// 模拟突发请求
	fmt.Println("模拟突发请求:")
	for i := 1; i <= 15; i++ {
		if bucket.Allow() {
			fmt.Printf("请求 %d: ✓ 允许 (剩余令牌: %d)\n", i, bucket.GetTokens())
		} else {
			fmt.Printf("请求 %d: ✗ 拒绝 (剩余令牌: %d)\n", i, bucket.GetTokens())
		}
	}

	// 等待令牌补充
	fmt.Println("\n等待 3 秒补充令牌...")
	time.Sleep(3 * time.Second)
	fmt.Printf("等待后令牌数: %d\n", bucket.GetTokens())

	// 测试预约机制
	fmt.Println("\n测试预约机制:")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	if err := bucket.ReserveN(ctx, 3); err != nil {
		fmt.Printf("预约失败: %v\n", err)
	} else {
		fmt.Printf("预约成功，等待时间: %v\n", time.Since(start))
	}

	fmt.Println()
}

// =============================================================================
// 3. 漏桶算法
// =============================================================================

// LeakyBucket 漏桶限流器
type LeakyBucket struct {
	capacity int64      // 桶容量
	level    int64      // 当前水位
	leakRate int64      // 每秒漏出速率
	lastLeak time.Time  // 上次漏水时间
	mu       sync.Mutex // 互斥锁
}

// NewLeakyBucket 创建漏桶
func NewLeakyBucket(capacity, leakRate int64) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		level:    0,
		leakRate: leakRate,
		lastLeak: time.Now(),
	}
}

// leak 漏水
func (lb *LeakyBucket) leak() {
	now := time.Now()
	elapsed := now.Sub(lb.lastLeak)

	// 计算应该漏出的水量
	waterToLeak := int64(elapsed.Seconds()) * lb.leakRate
	if waterToLeak > 0 {
		lb.level = max(0, lb.level-waterToLeak)
		lb.lastLeak = now
	}
}

// Allow 检查是否允许请求
func (lb *LeakyBucket) Allow() bool {
	return lb.AllowN(1)
}

// AllowN 检查是否允许N个请求
func (lb *LeakyBucket) AllowN(n int64) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()

	if lb.level+n <= lb.capacity {
		lb.level += n
		return true
	}
	return false
}

// GetLevel 获取当前水位
func (lb *LeakyBucket) GetLevel() int64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.leak()
	return lb.level
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func demonstrateLeakyBucket() {
	fmt.Println("=== 2. 漏桶算法 ===")

	// 创建漏桶：容量10，每秒漏出1个
	bucket := NewLeakyBucket(10, 1)

	fmt.Printf("初始水位: %d\n", bucket.GetLevel())

	// 模拟持续请求
	fmt.Println("模拟持续请求:")
	for i := 1; i <= 15; i++ {
		if bucket.Allow() {
			fmt.Printf("请求 %d: ✓ 允许 (当前水位: %d)\n", i, bucket.GetLevel())
		} else {
			fmt.Printf("请求 %d: ✗ 拒绝 (当前水位: %d)\n", i, bucket.GetLevel())
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 等待桶排空
	fmt.Println("\n等待桶排空...")
	for bucket.GetLevel() > 0 {
		fmt.Printf("当前水位: %d\n", bucket.GetLevel())
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("最终水位: %d\n", bucket.GetLevel())
	fmt.Println()
}

// =============================================================================
// 4. 滑动窗口算法
// =============================================================================

// SlidingWindow 滑动窗口限流器
type SlidingWindow struct {
	windowSize time.Duration // 窗口大小
	limit      int64         // 窗口内限制
	requests   []time.Time   // 请求时间记录
	mu         sync.Mutex    // 互斥锁
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(windowSize time.Duration, limit int64) *SlidingWindow {
	return &SlidingWindow{
		windowSize: windowSize,
		limit:      limit,
		requests:   make([]time.Time, 0),
	}
}

// cleanOldRequests 清理过期请求
func (sw *SlidingWindow) cleanOldRequests() {
	now := time.Now()
	cutoff := now.Add(-sw.windowSize)

	// 找到第一个未过期的请求
	i := 0
	for i < len(sw.requests) && sw.requests[i].Before(cutoff) {
		i++
	}

	// 移除过期请求
	if i > 0 {
		sw.requests = sw.requests[i:]
	}
}

// Allow 检查是否允许请求
func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.cleanOldRequests()

	if int64(len(sw.requests)) < sw.limit {
		sw.requests = append(sw.requests, time.Now())
		return true
	}
	return false
}

// GetRequestCount 获取当前窗口内请求数
func (sw *SlidingWindow) GetRequestCount() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.cleanOldRequests()
	return len(sw.requests)
}

func demonstrateSlidingWindow() {
	fmt.Println("=== 3. 滑动窗口算法 ===")

	// 创建滑动窗口：3秒窗口，最多5个请求
	window := NewSlidingWindow(3*time.Second, 5)

	fmt.Println("模拟在3秒窗口内的请求:")

	// 快速发送请求
	for i := 1; i <= 8; i++ {
		if window.Allow() {
			fmt.Printf("请求 %d: ✓ 允许 (窗口内请求数: %d)\n", i, window.GetRequestCount())
		} else {
			fmt.Printf("请求 %d: ✗ 拒绝 (窗口内请求数: %d)\n", i, window.GetRequestCount())
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 等待窗口滑动
	fmt.Println("\n等待窗口滑动...")
	time.Sleep(4 * time.Second)

	// 再次发送请求
	fmt.Println("窗口滑动后的请求:")
	for i := 9; i <= 12; i++ {
		if window.Allow() {
			fmt.Printf("请求 %d: ✓ 允许 (窗口内请求数: %d)\n", i, window.GetRequestCount())
		} else {
			fmt.Printf("请求 %d: ✗ 拒绝 (窗口内请求数: %d)\n", i, window.GetRequestCount())
		}
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println()
}

// =============================================================================
// 5. 自适应限流器
// =============================================================================

// AdaptiveRateLimiter 自适应限流器
type AdaptiveRateLimiter struct {
	minRate        int64         // 最小速率
	maxRate        int64         // 最大速率
	currentRate    int64         // 当前速率
	successCount   int64         // 成功计数
	errorCount     int64         // 错误计数
	windowSize     time.Duration // 统计窗口
	adjustInterval time.Duration // 调整间隔
	bucket         *TokenBucket  // 底层令牌桶
	lastAdjust     time.Time     // 上次调整时间
	mu             sync.Mutex    // 互斥锁
}

// NewAdaptiveRateLimiter 创建自适应限流器
func NewAdaptiveRateLimiter(minRate, maxRate int64) *AdaptiveRateLimiter {
	initialRate := minRate
	return &AdaptiveRateLimiter{
		minRate:        minRate,
		maxRate:        maxRate,
		currentRate:    initialRate,
		windowSize:     10 * time.Second,
		adjustInterval: 5 * time.Second,
		bucket:         NewTokenBucket(initialRate*2, initialRate),
		lastAdjust:     time.Now(),
	}
}

// Allow 检查是否允许请求
func (arl *AdaptiveRateLimiter) Allow() bool {
	arl.adjust()
	return arl.bucket.Allow()
}

// RecordSuccess 记录成功
func (arl *AdaptiveRateLimiter) RecordSuccess() {
	atomic.AddInt64(&arl.successCount, 1)
}

// RecordError 记录错误
func (arl *AdaptiveRateLimiter) RecordError() {
	atomic.AddInt64(&arl.errorCount, 1)
}

// adjust 调整速率
func (arl *AdaptiveRateLimiter) adjust() {
	arl.mu.Lock()
	defer arl.mu.Unlock()

	now := time.Now()
	if now.Sub(arl.lastAdjust) < arl.adjustInterval {
		return
	}

	successCount := atomic.SwapInt64(&arl.successCount, 0)
	errorCount := atomic.SwapInt64(&arl.errorCount, 0)
	totalRequests := successCount + errorCount

	if totalRequests == 0 {
		return
	}

	errorRate := float64(errorCount) / float64(totalRequests)

	// 调整策略
	var newRate int64
	if errorRate > 0.1 { // 错误率超过10%，降低速率
		newRate = int64(float64(arl.currentRate) * 0.8)
		fmt.Printf("自适应限流: 错误率 %.2f%%，降低速率到 %d\n", errorRate*100, newRate)
	} else if errorRate < 0.01 { // 错误率低于1%，提高速率
		newRate = int64(float64(arl.currentRate) * 1.2)
		fmt.Printf("自适应限流: 错误率 %.2f%%，提高速率到 %d\n", errorRate*100, newRate)
	} else {
		newRate = arl.currentRate
	}

	// 限制在最小值和最大值之间
	if newRate < arl.minRate {
		newRate = arl.minRate
	}
	if newRate > arl.maxRate {
		newRate = arl.maxRate
	}

	if newRate != arl.currentRate {
		arl.currentRate = newRate
		arl.bucket = NewTokenBucket(newRate*2, newRate)
		fmt.Printf("自适应限流: 速率调整为 %d/秒\n", newRate)
	}

	arl.lastAdjust = now
}

// GetCurrentRate 获取当前速率
func (arl *AdaptiveRateLimiter) GetCurrentRate() int64 {
	arl.mu.Lock()
	defer arl.mu.Unlock()
	return arl.currentRate
}

func demonstrateAdaptiveRateLimiter() {
	fmt.Println("=== 4. 自适应限流器 ===")

	limiter := NewAdaptiveRateLimiter(5, 50)

	// 模拟不同阶段的请求处理
	phases := []struct {
		name      string
		duration  time.Duration
		errorRate float64
	}{
		{"低错误率阶段", 8 * time.Second, 0.05},
		{"高错误率阶段", 8 * time.Second, 0.20},
		{"恢复阶段", 8 * time.Second, 0.02},
	}

	for _, phase := range phases {
		fmt.Printf("\n--- %s (错误率: %.1f%%) ---\n", phase.name, phase.errorRate*100)

		start := time.Now()
		requestCount := 0

		for time.Since(start) < phase.duration {
			if limiter.Allow() {
				requestCount++

				// 模拟请求处理
				if rand.Float64() < phase.errorRate {
					limiter.RecordError()
				} else {
					limiter.RecordSuccess()
				}

				if requestCount%20 == 0 {
					fmt.Printf("已处理 %d 个请求，当前速率: %d/秒\n",
						requestCount, limiter.GetCurrentRate())
				}
			}

			time.Sleep(50 * time.Millisecond)
		}

		fmt.Printf("%s完成，总请求数: %d\n", phase.name, requestCount)
	}

	fmt.Printf("最终速率: %d/秒\n", limiter.GetCurrentRate())
	fmt.Println()
}

// =============================================================================
// 6. 分布式限流器
// =============================================================================

// DistributedRateLimiter 分布式限流器（模拟）
type DistributedRateLimiter struct {
	nodeID      string
	totalLimit  int64
	nodeLimit   int64
	nodeCount   int64
	syncCh      chan NodeStatus
	localBucket *TokenBucket
	globalStats map[string]NodeStatus
	mu          sync.RWMutex
}

// NodeStatus 节点状态
type NodeStatus struct {
	NodeID    string
	Rate      int64
	Timestamp time.Time
}

// NewDistributedRateLimiter 创建分布式限流器
func NewDistributedRateLimiter(nodeID string, totalLimit, nodeCount int64) *DistributedRateLimiter {
	nodeLimit := totalLimit / nodeCount

	return &DistributedRateLimiter{
		nodeID:      nodeID,
		totalLimit:  totalLimit,
		nodeLimit:   nodeLimit,
		nodeCount:   nodeCount,
		syncCh:      make(chan NodeStatus, 10),
		localBucket: NewTokenBucket(nodeLimit*2, nodeLimit),
		globalStats: make(map[string]NodeStatus),
	}
}

// Start 启动分布式限流器
func (drl *DistributedRateLimiter) Start() {
	go drl.syncWorker()
	go drl.reportWorker()
}

// Allow 检查是否允许请求
func (drl *DistributedRateLimiter) Allow() bool {
	return drl.localBucket.Allow()
}

// syncWorker 同步工作者
func (drl *DistributedRateLimiter) syncWorker() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case status := <-drl.syncCh:
			drl.mu.Lock()
			drl.globalStats[status.NodeID] = status
			drl.mu.Unlock()

			fmt.Printf("节点 %s 收到来自节点 %s 的状态更新: 速率 %d\n",
				drl.nodeID, status.NodeID, status.Rate)

		case <-ticker.C:
			drl.adjustNodeLimit()
		}
	}
}

// reportWorker 报告工作者
func (drl *DistributedRateLimiter) reportWorker() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// 向其他节点报告自己的状态
		status := NodeStatus{
			NodeID:    drl.nodeID,
			Rate:      drl.nodeLimit,
			Timestamp: time.Now(),
		}

		// 模拟向其他节点发送状态（这里只是打印）
		fmt.Printf("节点 %s 报告状态: 速率 %d\n", drl.nodeID, status.Rate)
	}
}

// adjustNodeLimit 调整节点限制
func (drl *DistributedRateLimiter) adjustNodeLimit() {
	drl.mu.RLock()
	activeNodes := int64(len(drl.globalStats) + 1) // +1 包括自己
	drl.mu.RUnlock()

	// 根据活跃节点数调整每个节点的限制
	newNodeLimit := drl.totalLimit / activeNodes

	if newNodeLimit != drl.nodeLimit {
		drl.nodeLimit = newNodeLimit
		drl.localBucket = NewTokenBucket(newNodeLimit*2, newNodeLimit)

		fmt.Printf("节点 %s 调整限制: %d/秒 (活跃节点数: %d)\n",
			drl.nodeID, newNodeLimit, activeNodes)
	}
}

// ReceiveStatus 接收其他节点状态（模拟网络通信）
func (drl *DistributedRateLimiter) ReceiveStatus(status NodeStatus) {
	select {
	case drl.syncCh <- status:
	default:
		// 通道满了，丢弃
	}
}

func demonstrateDistributedRateLimiter() {
	fmt.Println("=== 5. 分布式限流器 ===")

	// 创建3个节点的分布式限流器，总限制100/秒
	nodes := []*DistributedRateLimiter{
		NewDistributedRateLimiter("node-1", 100, 3),
		NewDistributedRateLimiter("node-2", 100, 3),
		NewDistributedRateLimiter("node-3", 100, 3),
	}

	// 启动所有节点
	for _, node := range nodes {
		node.Start()
	}

	// 模拟节点间状态同步
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C

			// 模拟节点间状态交换
			for i, node := range nodes {
				for j, otherNode := range nodes {
					if i != j {
						status := NodeStatus{
							NodeID:    otherNode.nodeID,
							Rate:      otherNode.nodeLimit,
							Timestamp: time.Now(),
						}
						node.ReceiveStatus(status)
					}
				}
			}
		}
	}()

	// 模拟各节点的请求处理
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(n *DistributedRateLimiter) {
			defer wg.Done()

			requestCount := 0
			for i := 0; i < 100; i++ {
				if n.Allow() {
					requestCount++
				}
				time.Sleep(50 * time.Millisecond)
			}

			fmt.Printf("节点 %s 处理了 %d 个请求\n", n.nodeID, requestCount)
		}(node)
	}

	wg.Wait()
	fmt.Println()
}

// =============================================================================
// 7. 限流器性能比较
// =============================================================================

func demonstratePerformanceComparison() {
	fmt.Println("=== 6. 限流器性能比较 ===")

	const requestCount = 10000

	// 测试令牌桶性能
	tokenBucket := NewTokenBucket(1000, 500)
	start := time.Now()
	for i := 0; i < requestCount; i++ {
		tokenBucket.Allow()
	}
	tokenBucketTime := time.Since(start)

	// 测试漏桶性能
	leakyBucket := NewLeakyBucket(1000, 500)
	start = time.Now()
	for i := 0; i < requestCount; i++ {
		leakyBucket.Allow()
	}
	leakyBucketTime := time.Since(start)

	// 测试滑动窗口性能
	slidingWindow := NewSlidingWindow(time.Second, 1000)
	start = time.Now()
	for i := 0; i < requestCount; i++ {
		slidingWindow.Allow()
	}
	slidingWindowTime := time.Since(start)

	fmt.Printf("性能对比 (%d 次请求):\n", requestCount)
	fmt.Printf("令牌桶:     %v\n", tokenBucketTime)
	fmt.Printf("漏桶:       %v\n", leakyBucketTime)
	fmt.Printf("滑动窗口:   %v\n", slidingWindowTime)

	fmt.Println("\n算法特点对比:")
	fmt.Println("令牌桶:")
	fmt.Println("  ✓ 允许突发流量")
	fmt.Println("  ✓ 性能好")
	fmt.Println("  ✓ 实现简单")

	fmt.Println("漏桶:")
	fmt.Println("  ✓ 输出平滑")
	fmt.Println("  ✓ 严格速率控制")
	fmt.Println("  ✗ 不允许突发")

	fmt.Println("滑动窗口:")
	fmt.Println("  ✓ 精确统计")
	fmt.Println("  ✓ 灵活配置")
	fmt.Println("  ✗ 内存消耗大")

	fmt.Println()
}

// =============================================================================
// 8. 限流最佳实践
// =============================================================================

func demonstrateRateLimitingBestPractices() {
	fmt.Println("=== 7. 限流最佳实践 ===")

	fmt.Println("1. 算法选择指南:")
	fmt.Println("   • 令牌桶：适合允许突发流量的场景")
	fmt.Println("   • 漏桶：适合需要平滑输出的场景")
	fmt.Println("   • 滑动窗口：适合需要精确统计的场景")
	fmt.Println("   • 固定窗口：适合简单场景但注意临界问题")

	fmt.Println("\n2. 实施策略:")
	fmt.Println("   ✓ 多层限流：应用层、网关层、数据库层")
	fmt.Println("   ✓ 分级限流：不同用户不同限制")
	fmt.Println("   ✓ 动态调整：根据系统负载自适应")
	fmt.Println("   ✓ 优雅降级：超限时的处理策略")

	fmt.Println("\n3. 监控和告警:")
	fmt.Println("   ✓ 限流触发率监控")
	fmt.Println("   ✓ 系统负载监控")
	fmt.Println("   ✓ 错误率监控")
	fmt.Println("   ✓ 响应时间监控")

	fmt.Println("\n4. 常见陷阱:")
	fmt.Println("   ✗ 限流参数设置过于严格")
	fmt.Println("   ✗ 没有考虑突发流量")
	fmt.Println("   ✗ 限流器性能成为瓶颈")
	fmt.Println("   ✗ 缺乏限流效果监控")

	fmt.Println("\n5. 与其他机制的配合:")
	fmt.Println("   • 熔断器：快速失败保护")
	fmt.Println("   • 重试机制：处理临时失败")
	fmt.Println("   • 背压：系统过载保护")
	fmt.Println("   • 负载均衡：流量分散")

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - 限流和节流")
	fmt.Println("=========================")

	demonstrateTokenBucket()
	demonstrateLeakyBucket()
	demonstrateSlidingWindow()
	demonstrateAdaptiveRateLimiter()
	demonstrateDistributedRateLimiter()
	demonstratePerformanceComparison()
	demonstrateRateLimitingBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个支持多种算法的统一限流器")
	fmt.Println("2. 创建一个基于Redis的分布式限流器")
	fmt.Println("3. 实现一个HTTP中间件形式的限流器")
	fmt.Println("4. 编写一个支持配额管理的限流系统")
	fmt.Println("5. 创建一个智能限流器，支持机器学习调优")
	fmt.Println("6. 实现一个多维度限流器（IP、用户、API等）")
	fmt.Println("\n请在此基础上练习更多限流技术的应用！")
}
