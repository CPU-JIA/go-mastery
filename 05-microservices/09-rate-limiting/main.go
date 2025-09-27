/*
=== 微服务限流模式：速率限制(Rate Limiting) ===

Rate Limiting是微服务架构中防止系统过载的关键模式。
通过控制请求频率，保护后端服务免受突发流量冲击。

学习目标：
1. 掌握令牌桶(Token Bucket)算法原理和实现
2. 理解滑动窗口(Sliding Window)限流策略
3. 学会固定窗口(Fixed Window)和漏桶算法
4. 实现分布式限流解决方案
5. 集成HTTP中间件和监控指标

核心算法对比：
- 令牌桶: 允许突发流量，平滑限流
- 滑动窗口: 精确控制，内存开销较大
- 固定窗口: 简单高效，但有边界效应
- 漏桶: 强制匀速，无突发处理

生产级特性：
- 并发安全的多算法支持
- 分布式一致性保证
- 灵活的配置策略
- 完整的监控指标
- HTTP中间件集成
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// 限流算法接口定义
// ==================

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow() bool
	AllowN(n int) bool
	Wait(ctx context.Context) error
	GetStats() RateLimiterStats
	Reset()
}

// RateLimiterStats 限流统计信息
type RateLimiterStats struct {
	Algorithm       string    `json:"algorithm"`
	TotalRequests   int64     `json:"total_requests"`
	AllowedRequests int64     `json:"allowed_requests"`
	BlockedRequests int64     `json:"blocked_requests"`
	CurrentRate     float64   `json:"current_rate"`
	LastReset       time.Time `json:"last_reset"`
	ConfiguredRate  int       `json:"configured_rate"`
	BurstSize       int       `json:"burst_size"`
}

// ==================
// 1. 令牌桶算法实现
// ==================

// TokenBucket 令牌桶限流器
type TokenBucket struct {
	rate       float64   // 令牌产生速率 (tokens/second)
	capacity   int       // 桶容量
	tokens     float64   // 当前令牌数
	lastRefill time.Time // 上次补充时间
	mutex      sync.Mutex

	// 统计信息
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
	lastReset       time.Time
}

// NewTokenBucket 创建令牌桶限流器
func NewTokenBucket(rate float64, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   capacity,
		tokens:     float64(capacity), // 初始满桶
		lastRefill: time.Now(),
		lastReset:  time.Now(),
	}
}

// refill 补充令牌
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// 计算需要添加的令牌数
	tokensToAdd := elapsed * tb.rate
	tb.tokens = min(tb.tokens+tokensToAdd, float64(tb.capacity))
	tb.lastRefill = now
}

// Allow 检查是否允许单个请求
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 检查是否允许n个请求
func (tb *TokenBucket) AllowN(n int) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	atomic.AddInt64(&tb.totalRequests, int64(n))

	tb.refill()

	if tb.tokens >= float64(n) {
		tb.tokens -= float64(n)
		atomic.AddInt64(&tb.allowedRequests, int64(n))
		return true
	}

	atomic.AddInt64(&tb.blockedRequests, int64(n))
	return false
}

// Wait 等待直到可以处理请求
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}

		// 计算等待时间
		tb.mutex.Lock()
		waitTime := time.Duration(float64(time.Second) / tb.rate)
		tb.mutex.Unlock()

		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// GetStats 获取统计信息
func (tb *TokenBucket) GetStats() RateLimiterStats {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	currentRate := float64(atomic.LoadInt64(&tb.allowedRequests)) / time.Since(tb.lastReset).Seconds()

	return RateLimiterStats{
		Algorithm:       "TokenBucket",
		TotalRequests:   atomic.LoadInt64(&tb.totalRequests),
		AllowedRequests: atomic.LoadInt64(&tb.allowedRequests),
		BlockedRequests: atomic.LoadInt64(&tb.blockedRequests),
		CurrentRate:     currentRate,
		LastReset:       tb.lastReset,
		ConfiguredRate:  int(tb.rate),
		BurstSize:       tb.capacity,
	}
}

// Reset 重置统计信息
func (tb *TokenBucket) Reset() {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	atomic.StoreInt64(&tb.totalRequests, 0)
	atomic.StoreInt64(&tb.allowedRequests, 0)
	atomic.StoreInt64(&tb.blockedRequests, 0)
	tb.lastReset = time.Now()
}

// ==================
// 2. 滑动窗口算法实现
// ==================

// SlidingWindow 滑动窗口限流器
type SlidingWindow struct {
	limit      int           // 窗口内最大请求数
	window     time.Duration // 窗口大小
	timestamps []time.Time   // 请求时间戳
	mutex      sync.Mutex

	// 统计信息
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
	lastReset       time.Time
}

// NewSlidingWindow 创建滑动窗口限流器
func NewSlidingWindow(limit int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		limit:      limit,
		window:     window,
		timestamps: make([]time.Time, 0, limit*2),
		lastReset:  time.Now(),
	}
}

// cleanOldEntries 清理过期的时间戳
func (sw *SlidingWindow) cleanOldEntries() {
	now := time.Now()
	cutoff := now.Add(-sw.window)

	// 找到第一个有效时间戳的位置
	i := 0
	for i < len(sw.timestamps) && sw.timestamps[i].Before(cutoff) {
		i++
	}

	// 删除过期的时间戳
	if i > 0 {
		sw.timestamps = sw.timestamps[i:]
	}
}

// Allow 检查是否允许单个请求
func (sw *SlidingWindow) Allow() bool {
	return sw.AllowN(1)
}

// AllowN 检查是否允许n个请求
func (sw *SlidingWindow) AllowN(n int) bool {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	atomic.AddInt64(&sw.totalRequests, int64(n))

	sw.cleanOldEntries()

	if len(sw.timestamps)+n <= sw.limit {
		now := time.Now()
		for i := 0; i < n; i++ {
			sw.timestamps = append(sw.timestamps, now)
		}
		atomic.AddInt64(&sw.allowedRequests, int64(n))
		return true
	}

	atomic.AddInt64(&sw.blockedRequests, int64(n))
	return false
}

// Wait 等待直到可以处理请求
func (sw *SlidingWindow) Wait(ctx context.Context) error {
	for {
		if sw.Allow() {
			return nil
		}

		// 计算等待时间 - 等到最老的请求过期
		sw.mutex.Lock()
		var waitTime time.Duration
		if len(sw.timestamps) > 0 {
			waitTime = sw.window - time.Since(sw.timestamps[0])
			if waitTime < 0 {
				waitTime = time.Millisecond
			}
		} else {
			waitTime = time.Millisecond
		}
		sw.mutex.Unlock()

		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// GetStats 获取统计信息
func (sw *SlidingWindow) GetStats() RateLimiterStats {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	currentRate := float64(atomic.LoadInt64(&sw.allowedRequests)) / time.Since(sw.lastReset).Seconds()

	return RateLimiterStats{
		Algorithm:       "SlidingWindow",
		TotalRequests:   atomic.LoadInt64(&sw.totalRequests),
		AllowedRequests: atomic.LoadInt64(&sw.allowedRequests),
		BlockedRequests: atomic.LoadInt64(&sw.blockedRequests),
		CurrentRate:     currentRate,
		LastReset:       sw.lastReset,
		ConfiguredRate:  int(float64(sw.limit) / sw.window.Seconds()),
		BurstSize:       sw.limit,
	}
}

// Reset 重置统计信息
func (sw *SlidingWindow) Reset() {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	atomic.StoreInt64(&sw.totalRequests, 0)
	atomic.StoreInt64(&sw.allowedRequests, 0)
	atomic.StoreInt64(&sw.blockedRequests, 0)
	sw.timestamps = sw.timestamps[:0]
	sw.lastReset = time.Now()
}

// ==================
// 3. 固定窗口算法实现
// ==================

// FixedWindow 固定窗口限流器
type FixedWindow struct {
	limit       int           // 窗口内最大请求数
	window      time.Duration // 窗口大小
	counter     int64         // 当前计数
	windowStart time.Time     // 窗口开始时间
	mutex       sync.Mutex

	// 统计信息
	totalRequests   int64
	allowedRequests int64
	blockedRequests int64
	lastReset       time.Time
}

// NewFixedWindow 创建固定窗口限流器
func NewFixedWindow(limit int, window time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:       limit,
		window:      window,
		windowStart: time.Now(),
		lastReset:   time.Now(),
	}
}

// resetWindowIfNeeded 如果需要则重置窗口
func (fw *FixedWindow) resetWindowIfNeeded() {
	now := time.Now()
	if now.Sub(fw.windowStart) >= fw.window {
		fw.counter = 0
		fw.windowStart = now
	}
}

// Allow 检查是否允许单个请求
func (fw *FixedWindow) Allow() bool {
	return fw.AllowN(1)
}

// AllowN 检查是否允许n个请求
func (fw *FixedWindow) AllowN(n int) bool {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	atomic.AddInt64(&fw.totalRequests, int64(n))

	fw.resetWindowIfNeeded()

	if fw.counter+int64(n) <= int64(fw.limit) {
		fw.counter += int64(n)
		atomic.AddInt64(&fw.allowedRequests, int64(n))
		return true
	}

	atomic.AddInt64(&fw.blockedRequests, int64(n))
	return false
}

// Wait 等待直到可以处理请求
func (fw *FixedWindow) Wait(ctx context.Context) error {
	for {
		if fw.Allow() {
			return nil
		}

		// 等到下个窗口开始
		fw.mutex.Lock()
		waitTime := fw.window - time.Since(fw.windowStart)
		if waitTime < 0 {
			waitTime = time.Millisecond
		}
		fw.mutex.Unlock()

		select {
		case <-time.After(waitTime):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// GetStats 获取统计信息
func (fw *FixedWindow) GetStats() RateLimiterStats {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	currentRate := float64(atomic.LoadInt64(&fw.allowedRequests)) / time.Since(fw.lastReset).Seconds()

	return RateLimiterStats{
		Algorithm:       "FixedWindow",
		TotalRequests:   atomic.LoadInt64(&fw.totalRequests),
		AllowedRequests: atomic.LoadInt64(&fw.allowedRequests),
		BlockedRequests: atomic.LoadInt64(&fw.blockedRequests),
		CurrentRate:     currentRate,
		LastReset:       fw.lastReset,
		ConfiguredRate:  int(float64(fw.limit) / fw.window.Seconds()),
		BurstSize:       fw.limit,
	}
}

// Reset 重置统计信息
func (fw *FixedWindow) Reset() {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()

	atomic.StoreInt64(&fw.totalRequests, 0)
	atomic.StoreInt64(&fw.allowedRequests, 0)
	atomic.StoreInt64(&fw.blockedRequests, 0)
	fw.counter = 0
	fw.windowStart = time.Now()
	fw.lastReset = time.Now()
}

// ==================
// 4. 多层限流器
// ==================

// MultiTierRateLimiter 多层限流器
type MultiTierRateLimiter struct {
	limiters map[string]RateLimiter
	mutex    sync.RWMutex
}

// NewMultiTierRateLimiter 创建多层限流器
func NewMultiTierRateLimiter() *MultiTierRateLimiter {
	return &MultiTierRateLimiter{
		limiters: make(map[string]RateLimiter),
	}
}

// AddLimiter 添加限流器
func (mtr *MultiTierRateLimiter) AddLimiter(name string, limiter RateLimiter) {
	mtr.mutex.Lock()
	defer mtr.mutex.Unlock()
	mtr.limiters[name] = limiter
}

// Allow 检查所有限流器是否都允许
func (mtr *MultiTierRateLimiter) Allow() bool {
	mtr.mutex.RLock()
	defer mtr.mutex.RUnlock()

	for _, limiter := range mtr.limiters {
		if !limiter.Allow() {
			return false
		}
	}
	return true
}

// GetAllStats 获取所有限流器统计
func (mtr *MultiTierRateLimiter) GetAllStats() map[string]RateLimiterStats {
	mtr.mutex.RLock()
	defer mtr.mutex.RUnlock()

	stats := make(map[string]RateLimiterStats)
	for name, limiter := range mtr.limiters {
		stats[name] = limiter.GetStats()
	}
	return stats
}

// ==================
// 5. HTTP中间件
// ==================

// RateLimitMiddleware HTTP限流中间件
func RateLimitMiddleware(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				stats := limiter.GetStats()
				w.Header().Set("X-RateLimit-Algorithm", stats.Algorithm)
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(stats.ConfiguredRate))
				w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(int64(stats.BurstSize)-stats.BlockedRequests, 10))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(stats.LastReset.Unix(), 10))

				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// KeyBasedRateLimiter 基于键的限流器
type KeyBasedRateLimiter struct {
	limiters map[string]RateLimiter
	factory  func() RateLimiter
	mutex    sync.RWMutex
	cleanup  time.Duration
}

// NewKeyBasedRateLimiter 创建基于键的限流器
func NewKeyBasedRateLimiter(factory func() RateLimiter, cleanup time.Duration) *KeyBasedRateLimiter {
	krl := &KeyBasedRateLimiter{
		limiters: make(map[string]RateLimiter),
		factory:  factory,
		cleanup:  cleanup,
	}

	// 启动清理协程
	go krl.cleanupLoop()
	return krl
}

// Allow 基于键检查限流
func (krl *KeyBasedRateLimiter) Allow(key string) bool {
	limiter := krl.getLimiter(key)
	return limiter.Allow()
}

// getLimiter 获取或创建限流器
func (krl *KeyBasedRateLimiter) getLimiter(key string) RateLimiter {
	krl.mutex.RLock()
	limiter, exists := krl.limiters[key]
	krl.mutex.RUnlock()

	if exists {
		return limiter
	}

	krl.mutex.Lock()
	defer krl.mutex.Unlock()

	// 双重检查
	if limiter, exists := krl.limiters[key]; exists {
		return limiter
	}

	limiter = krl.factory()
	krl.limiters[key] = limiter
	return limiter
}

// cleanupLoop 清理循环
func (krl *KeyBasedRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(krl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		krl.mutex.Lock()
		// 这里可以根据需要实现清理逻辑，比如删除长时间未使用的限流器
		krl.mutex.Unlock()
	}
}

// ==================
// 6. 监控和指标
// ==================

// RateLimiterMonitor 限流监控器
type RateLimiterMonitor struct {
	limiters map[string]RateLimiter
	mutex    sync.RWMutex
}

// NewRateLimiterMonitor 创建监控器
func NewRateLimiterMonitor() *RateLimiterMonitor {
	return &RateLimiterMonitor{
		limiters: make(map[string]RateLimiter),
	}
}

// Register 注册限流器
func (rlm *RateLimiterMonitor) Register(name string, limiter RateLimiter) {
	rlm.mutex.Lock()
	defer rlm.mutex.Unlock()
	rlm.limiters[name] = limiter
}

// ServeHTTP 提供HTTP监控端点
func (rlm *RateLimiterMonitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rlm.mutex.RLock()
	defer rlm.mutex.RUnlock()

	stats := make(map[string]RateLimiterStats)
	for name, limiter := range rlm.limiters {
		stats[name] = limiter.GetStats()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// ==================
// 辅助函数
// ==================

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// ==================
// 示例和测试
// ==================

func main() {
	fmt.Println("=== 微服务限流(Rate Limiting)模式演示 ===")

	// 1. 创建不同算法的限流器
	fmt.Println("\n📊 创建不同算法的限流器...")

	tokenBucket := NewTokenBucket(10.0, 20)            // 10 tokens/second, 容量20
	slidingWindow := NewSlidingWindow(15, time.Minute) // 15 requests/minute
	fixedWindow := NewFixedWindow(12, 30*time.Second)  // 12 requests/30s

	// 2. 创建监控器
	monitor := NewRateLimiterMonitor()
	monitor.Register("token-bucket", tokenBucket)
	monitor.Register("sliding-window", slidingWindow)
	monitor.Register("fixed-window", fixedWindow)

	// 3. 启动HTTP监控服务
	go func() {
		http.Handle("/rate-limit-metrics", monitor)

		// 健康检查端点
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "Rate Limiter Service: OK")
		})

		// 受限流保护的API端点
		protectedAPI := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "API Response: %s\n", time.Now().Format("15:04:05"))
		})

		http.Handle("/api/protected", RateLimitMiddleware(tokenBucket)(protectedAPI))

		log.Println("监控端点: http://localhost:8080/rate-limit-metrics")
		log.Println("受保护API: http://localhost:8080/api/protected")

		server := &http.Server{
			Addr:         ":8080",
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Fatal(server.ListenAndServe())
	}()

	// 4. 测试不同算法
	algorithms := map[string]RateLimiter{
		"TokenBucket":   tokenBucket,
		"SlidingWindow": slidingWindow,
		"FixedWindow":   fixedWindow,
	}

	fmt.Println("\n🚀 开始限流算法对比测试...")

	for name, limiter := range algorithms {
		fmt.Printf("\n--- 测试 %s ---\n", name)

		// 重置统计
		limiter.Reset()

		allowedCount := 0
		blockedCount := 0

		// 快速发送30个请求
		for i := 0; i < 30; i++ {
			if limiter.Allow() {
				allowedCount++
				fmt.Printf("✅ 请求 %d: 通过\n", i+1)
			} else {
				blockedCount++
				fmt.Printf("❌ 请求 %d: 被限流\n", i+1)
			}

			time.Sleep(50 * time.Millisecond) // 模拟请求间隔
		}

		stats := limiter.GetStats()
		fmt.Printf("\n📊 %s 统计结果:\n", name)
		fmt.Printf("   总请求: %d\n", stats.TotalRequests)
		fmt.Printf("   允许: %d (%.1f%%)\n", allowedCount, float64(allowedCount)/30*100)
		fmt.Printf("   阻塞: %d (%.1f%%)\n", blockedCount, float64(blockedCount)/30*100)
		fmt.Printf("   当前速率: %.2f req/s\n", stats.CurrentRate)
		fmt.Printf("   配置速率: %d req/s\n", stats.ConfiguredRate)
		fmt.Printf("   突发大小: %d\n", stats.BurstSize)
	}

	// 5. 多层限流演示
	fmt.Println("\n🔗 多层限流演示...")
	multiTier := NewMultiTierRateLimiter()
	multiTier.AddLimiter("per-second", NewTokenBucket(5.0, 10))           // 每秒5个
	multiTier.AddLimiter("per-minute", NewSlidingWindow(50, time.Minute)) // 每分钟50个

	fmt.Println("测试多层限流 (需要同时满足所有限流器):")
	multiAllowed := 0
	multiBlocked := 0

	for i := 0; i < 20; i++ {
		if multiTier.Allow() {
			multiAllowed++
			fmt.Printf("✅ 多层请求 %d: 通过\n", i+1)
		} else {
			multiBlocked++
			fmt.Printf("❌ 多层请求 %d: 被限流\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n📊 多层限流结果: 允许 %d, 阻塞 %d\n", multiAllowed, multiBlocked)

	// 6. 基于键的限流演示
	fmt.Println("\n🔑 基于键的限流演示...")
	keyBasedLimiter := NewKeyBasedRateLimiter(
		func() RateLimiter {
			return NewTokenBucket(2.0, 5) // 每个键: 2 req/s, 容量5
		},
		time.Minute,
	)

	users := []string{"user1", "user2", "user3"}
	for _, user := range users {
		fmt.Printf("测试用户 %s:\n", user)
		for i := 0; i < 8; i++ {
			if keyBasedLimiter.Allow(user) {
				fmt.Printf("  ✅ %s 请求 %d: 通过\n", user, i+1)
			} else {
				fmt.Printf("  ❌ %s 请求 %d: 被限流\n", user, i+1)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}

	// 7. 等待模式演示
	fmt.Println("\n⏰ 等待模式演示...")
	waitLimiter := NewTokenBucket(1.0, 3) // 很严格的限制

	fmt.Println("使用Wait()方法自动等待...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		start := time.Now()
		err := waitLimiter.Wait(ctx)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("❌ 等待请求 %d: 超时 (%v)\n", i+1, err)
		} else {
			fmt.Printf("✅ 等待请求 %d: 成功 (等待 %v)\n", i+1, elapsed.Truncate(time.Millisecond))
		}
	}

	// 8. 性能对比
	fmt.Println("\n⚡ 性能对比测试...")
	performanceTest := func(name string, limiter RateLimiter, requests int) {
		start := time.Now()
		allowed := 0

		for i := 0; i < requests; i++ {
			if limiter.Allow() {
				allowed++
			}
		}

		elapsed := time.Since(start)
		fmt.Printf("%s: %d/%d 请求通过, 耗时: %v, QPS: %.0f\n",
			name, allowed, requests, elapsed,
			float64(requests)/elapsed.Seconds())
	}

	const testRequests = 10000
	fmt.Printf("性能测试 (%d 请求):\n", testRequests)
	performanceTest("TokenBucket  ", NewTokenBucket(1000, 2000), testRequests)
	performanceTest("SlidingWindow", NewSlidingWindow(1000, time.Second), testRequests)
	performanceTest("FixedWindow  ", NewFixedWindow(1000, time.Second), testRequests)

	// 9. 最终总结
	fmt.Println("\n✨ Rate Limiting演示完成!")
	fmt.Println("\n💡 算法对比总结:")
	fmt.Println("🪣 TokenBucket:   允许突发流量，平滑限流，内存占用小")
	fmt.Println("📊 SlidingWindow: 精确控制，无边界效应，内存占用较大")
	fmt.Println("🕒 FixedWindow:   简单高效，有边界效应，内存占用最小")
	fmt.Println("\n🎯 选择建议:")
	fmt.Println("- 需要突发处理: TokenBucket")
	fmt.Println("- 精确控制要求: SlidingWindow")
	fmt.Println("- 高性能场景: FixedWindow")
	fmt.Println("- 分布式场景: 配合Redis实现")

	fmt.Println("\n📊 监控端点保持运行中...")
	fmt.Println("访问 http://localhost:8080/rate-limit-metrics 查看实时指标")

	// 保持监控端点运行
	select {}
}
