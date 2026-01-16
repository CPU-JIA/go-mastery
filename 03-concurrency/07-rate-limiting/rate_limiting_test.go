package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =============================================================================
// TokenBucket 测试
// =============================================================================

// TestNewTokenBucket 测试创建令牌桶
func TestNewTokenBucket(t *testing.T) {
	tests := []struct {
		name       string
		capacity   int64
		refillRate int64
	}{
		{"标准配置", 10, 2},
		{"大容量", 1000, 100},
		{"小容量", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := NewTokenBucket(tt.capacity, tt.refillRate)

			if bucket == nil {
				t.Fatal("NewTokenBucket 返回 nil")
			}

			if bucket.capacity != tt.capacity {
				t.Errorf("期望容量 %d, 实际 %d", tt.capacity, bucket.capacity)
			}

			if bucket.refillRate != tt.refillRate {
				t.Errorf("期望补充速率 %d, 实际 %d", tt.refillRate, bucket.refillRate)
			}

			// 初始时桶应该是满的
			if bucket.GetTokens() != tt.capacity {
				t.Errorf("初始令牌数应该等于容量 %d, 实际 %d", tt.capacity, bucket.GetTokens())
			}
		})
	}
}

// TestTokenBucketAllow 测试令牌桶允许请求
func TestTokenBucketAllow(t *testing.T) {
	bucket := NewTokenBucket(5, 1)

	// 前5个请求应该被允许
	for i := 0; i < 5; i++ {
		if !bucket.Allow() {
			t.Errorf("第 %d 个请求应该被允许", i+1)
		}
	}

	// 第6个请求应该被拒绝
	if bucket.Allow() {
		t.Error("令牌耗尽后请求应该被拒绝")
	}
}

// TestTokenBucketAllowN 测试批量令牌请求
func TestTokenBucketAllowN(t *testing.T) {
	bucket := NewTokenBucket(10, 2)

	// 请求5个令牌
	if !bucket.AllowN(5) {
		t.Error("应该允许请求5个令牌")
	}

	if bucket.GetTokens() != 5 {
		t.Errorf("剩余令牌应该是5, 实际 %d", bucket.GetTokens())
	}

	// 请求6个令牌应该失败
	if bucket.AllowN(6) {
		t.Error("不应该允许请求6个令牌（只剩5个）")
	}

	// 令牌数不应该改变
	if bucket.GetTokens() != 5 {
		t.Errorf("失败的请求不应该消耗令牌, 期望5, 实际 %d", bucket.GetTokens())
	}
}

// TestTokenBucketRefill 测试令牌补充
func TestTokenBucketRefill(t *testing.T) {
	bucket := NewTokenBucket(10, 5) // 每秒补充5个

	// 消耗所有令牌
	bucket.AllowN(10)

	if bucket.GetTokens() != 0 {
		t.Errorf("消耗后令牌应该为0, 实际 %d", bucket.GetTokens())
	}

	// 等待1秒让令牌补充
	time.Sleep(1100 * time.Millisecond)

	tokens := bucket.GetTokens()
	if tokens < 5 {
		t.Errorf("等待1秒后应该至少有5个令牌, 实际 %d", tokens)
	}
}

// TestTokenBucketReserve 测试令牌预约
func TestTokenBucketReserve(t *testing.T) {
	bucket := NewTokenBucket(5, 10) // 每秒补充10个

	// 消耗所有令牌
	bucket.AllowN(5)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	err := bucket.Reserve(ctx)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("预约应该成功, 但返回错误: %v", err)
	}

	// 应该等待了一段时间
	if duration < 50*time.Millisecond {
		t.Log("预约几乎立即完成（可能有剩余令牌）")
	}
}

// TestTokenBucketReserveTimeout 测试令牌预约超时
func TestTokenBucketReserveTimeout(t *testing.T) {
	bucket := NewTokenBucket(5, 1) // 每秒只补充1个

	// 消耗所有令牌
	bucket.AllowN(5)

	// 使用非常短的超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := bucket.ReserveN(ctx, 10) // 请求10个令牌
	if err == nil {
		t.Error("预约应该因超时而失败")
	}
}

// TestTokenBucketConcurrent 测试令牌桶并发安全
func TestTokenBucketConcurrent(t *testing.T) {
	bucket := NewTokenBucket(1000, 100)

	const numGoroutines = 100
	var wg sync.WaitGroup
	var allowedCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if bucket.Allow() {
					atomic.AddInt64(&allowedCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	// 允许的请求数不应该超过初始容量
	if allowedCount > 1000 {
		t.Errorf("允许的请求数 %d 超过了容量 1000", allowedCount)
	}

	t.Logf("并发测试: 允许了 %d 个请求", allowedCount)
}

// =============================================================================
// LeakyBucket 测试
// =============================================================================

// TestNewLeakyBucket 测试创建漏桶
func TestNewLeakyBucket(t *testing.T) {
	tests := []struct {
		name     string
		capacity int64
		leakRate int64
	}{
		{"标准配置", 10, 2},
		{"大容量", 1000, 100},
		{"小容量", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := NewLeakyBucket(tt.capacity, tt.leakRate)

			if bucket == nil {
				t.Fatal("NewLeakyBucket 返回 nil")
			}

			if bucket.capacity != tt.capacity {
				t.Errorf("期望容量 %d, 实际 %d", tt.capacity, bucket.capacity)
			}

			if bucket.leakRate != tt.leakRate {
				t.Errorf("期望漏出速率 %d, 实际 %d", tt.leakRate, bucket.leakRate)
			}

			// 初始时桶应该是空的
			if bucket.GetLevel() != 0 {
				t.Errorf("初始水位应该为0, 实际 %d", bucket.GetLevel())
			}
		})
	}
}

// TestLeakyBucketAllow 测试漏桶允许请求
func TestLeakyBucketAllow(t *testing.T) {
	bucket := NewLeakyBucket(5, 1)

	// 前5个请求应该被允许
	for i := 0; i < 5; i++ {
		if !bucket.Allow() {
			t.Errorf("第 %d 个请求应该被允许", i+1)
		}
	}

	// 第6个请求应该被拒绝（桶满了）
	if bucket.Allow() {
		t.Error("桶满后请求应该被拒绝")
	}
}

// TestLeakyBucketAllowN 测试批量请求
func TestLeakyBucketAllowN(t *testing.T) {
	bucket := NewLeakyBucket(10, 2)

	// 请求5个
	if !bucket.AllowN(5) {
		t.Error("应该允许请求5个")
	}

	if bucket.GetLevel() != 5 {
		t.Errorf("水位应该是5, 实际 %d", bucket.GetLevel())
	}

	// 再请求6个应该失败
	if bucket.AllowN(6) {
		t.Error("不应该允许请求6个（会超过容量）")
	}

	// 水位不应该改变
	if bucket.GetLevel() != 5 {
		t.Errorf("失败的请求不应该改变水位, 期望5, 实际 %d", bucket.GetLevel())
	}
}

// TestLeakyBucketLeak 测试漏桶漏水
func TestLeakyBucketLeak(t *testing.T) {
	bucket := NewLeakyBucket(10, 5) // 每秒漏出5个

	// 填满桶
	bucket.AllowN(10)

	if bucket.GetLevel() != 10 {
		t.Errorf("填满后水位应该是10, 实际 %d", bucket.GetLevel())
	}

	// 等待1秒让水漏出
	time.Sleep(1100 * time.Millisecond)

	level := bucket.GetLevel()
	if level > 5 {
		t.Errorf("等待1秒后水位应该不超过5, 实际 %d", level)
	}
}

// TestLeakyBucketConcurrent 测试漏桶并发安全
func TestLeakyBucketConcurrent(t *testing.T) {
	bucket := NewLeakyBucket(1000, 100)

	const numGoroutines = 100
	var wg sync.WaitGroup
	var allowedCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if bucket.Allow() {
					atomic.AddInt64(&allowedCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	// 水位不应该超过容量
	if bucket.GetLevel() > 1000 {
		t.Errorf("水位 %d 超过了容量 1000", bucket.GetLevel())
	}

	t.Logf("并发测试: 允许了 %d 个请求", allowedCount)
}

// =============================================================================
// SlidingWindow 测试
// =============================================================================

// TestNewSlidingWindow 测试创建滑动窗口
func TestNewSlidingWindow(t *testing.T) {
	tests := []struct {
		name       string
		windowSize time.Duration
		limit      int64
	}{
		{"1秒窗口", 1 * time.Second, 10},
		{"5秒窗口", 5 * time.Second, 100},
		{"100毫秒窗口", 100 * time.Millisecond, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window := NewSlidingWindow(tt.windowSize, tt.limit)

			if window == nil {
				t.Fatal("NewSlidingWindow 返回 nil")
			}

			if window.windowSize != tt.windowSize {
				t.Errorf("期望窗口大小 %v, 实际 %v", tt.windowSize, window.windowSize)
			}

			if window.limit != tt.limit {
				t.Errorf("期望限制 %d, 实际 %d", tt.limit, window.limit)
			}

			// 初始时请求数应该为0
			if window.GetRequestCount() != 0 {
				t.Errorf("初始请求数应该为0, 实际 %d", window.GetRequestCount())
			}
		})
	}
}

// TestSlidingWindowAllow 测试滑动窗口允许请求
func TestSlidingWindowAllow(t *testing.T) {
	window := NewSlidingWindow(1*time.Second, 5)

	// 前5个请求应该被允许
	for i := 0; i < 5; i++ {
		if !window.Allow() {
			t.Errorf("第 %d 个请求应该被允许", i+1)
		}
	}

	// 第6个请求应该被拒绝
	if window.Allow() {
		t.Error("超过限制后请求应该被拒绝")
	}

	if window.GetRequestCount() != 5 {
		t.Errorf("请求数应该是5, 实际 %d", window.GetRequestCount())
	}
}

// TestSlidingWindowSlide 测试窗口滑动
func TestSlidingWindowSlide(t *testing.T) {
	window := NewSlidingWindow(500*time.Millisecond, 5)

	// 填满窗口
	for i := 0; i < 5; i++ {
		window.Allow()
	}

	if window.GetRequestCount() != 5 {
		t.Errorf("填满后请求数应该是5, 实际 %d", window.GetRequestCount())
	}

	// 等待窗口滑动
	time.Sleep(600 * time.Millisecond)

	// 旧请求应该过期
	count := window.GetRequestCount()
	if count != 0 {
		t.Errorf("窗口滑动后请求数应该是0, 实际 %d", count)
	}

	// 应该能再次允许请求
	if !window.Allow() {
		t.Error("窗口滑动后应该允许新请求")
	}
}

// TestSlidingWindowConcurrent 测试滑动窗口并发安全
func TestSlidingWindowConcurrent(t *testing.T) {
	window := NewSlidingWindow(1*time.Second, 100)

	const numGoroutines = 50
	var wg sync.WaitGroup
	var allowedCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				if window.Allow() {
					atomic.AddInt64(&allowedCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	// 允许的请求数不应该超过限制
	if allowedCount > 100 {
		t.Errorf("允许的请求数 %d 超过了限制 100", allowedCount)
	}

	t.Logf("并发测试: 允许了 %d 个请求", allowedCount)
}

// =============================================================================
// AdaptiveRateLimiter 测试
// =============================================================================

// TestNewAdaptiveRateLimiter 测试创建自适应限流器
func TestNewAdaptiveRateLimiter(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(5, 50)

	if limiter == nil {
		t.Fatal("NewAdaptiveRateLimiter 返回 nil")
	}

	if limiter.minRate != 5 {
		t.Errorf("期望最小速率 5, 实际 %d", limiter.minRate)
	}

	if limiter.maxRate != 50 {
		t.Errorf("期望最大速率 50, 实际 %d", limiter.maxRate)
	}

	// 初始速率应该是最小速率
	if limiter.GetCurrentRate() != 5 {
		t.Errorf("初始速率应该是 5, 实际 %d", limiter.GetCurrentRate())
	}
}

// TestAdaptiveRateLimiterAllow 测试自适应限流器允许请求
func TestAdaptiveRateLimiterAllow(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(10, 100)

	// 应该能允许一些请求
	allowedCount := 0
	for i := 0; i < 30; i++ {
		if limiter.Allow() {
			allowedCount++
		}
	}

	if allowedCount == 0 {
		t.Error("应该至少允许一些请求")
	}

	t.Logf("允许了 %d 个请求", allowedCount)
}

// TestAdaptiveRateLimiterRecordSuccess 测试记录成功
func TestAdaptiveRateLimiterRecordSuccess(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(5, 50)

	// 记录成功不应该崩溃
	for i := 0; i < 100; i++ {
		limiter.RecordSuccess()
	}
}

// TestAdaptiveRateLimiterRecordError 测试记录错误
func TestAdaptiveRateLimiterRecordError(t *testing.T) {
	limiter := NewAdaptiveRateLimiter(5, 50)

	// 记录错误不应该崩溃
	for i := 0; i < 100; i++ {
		limiter.RecordError()
	}
}

// =============================================================================
// min/max 辅助函数测试
// =============================================================================

// TestMin 测试 min 函数
func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int64
	}{
		{1, 2, 1},
		{2, 1, 1},
		{5, 5, 5},
		{0, 10, 0},
		{-5, 5, -5},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// TestMax 测试 max 函数
func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int64
	}{
		{1, 2, 2},
		{2, 1, 2},
		{5, 5, 5},
		{0, 10, 10},
		{-5, 5, 5},
	}

	for _, tt := range tests {
		result := max(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("max(%d, %d) = %d, 期望 %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkTokenBucketAllow 基准测试令牌桶
func BenchmarkTokenBucketAllow(b *testing.B) {
	bucket := NewTokenBucket(1000000, 1000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}

// BenchmarkLeakyBucketAllow 基准测试漏桶
func BenchmarkLeakyBucketAllow(b *testing.B) {
	bucket := NewLeakyBucket(1000000, 1000000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Allow()
	}
}

// BenchmarkSlidingWindowAllow 基准测试滑动窗口
func BenchmarkSlidingWindowAllow(b *testing.B) {
	window := NewSlidingWindow(time.Hour, 1000000) // 大窗口避免清理

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		window.Allow()
	}
}

// BenchmarkTokenBucketConcurrent 基准测试令牌桶并发
func BenchmarkTokenBucketConcurrent(b *testing.B) {
	bucket := NewTokenBucket(1000000, 1000000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.Allow()
		}
	})
}

// BenchmarkLeakyBucketConcurrent 基准测试漏桶并发
func BenchmarkLeakyBucketConcurrent(b *testing.B) {
	bucket := NewLeakyBucket(1000000, 1000000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.Allow()
		}
	})
}

// BenchmarkSlidingWindowConcurrent 基准测试滑动窗口并发
func BenchmarkSlidingWindowConcurrent(b *testing.B) {
	window := NewSlidingWindow(time.Hour, 10000000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			window.Allow()
		}
	})
}

// =============================================================================
// 边界条件测试
// =============================================================================

// TestTokenBucketZeroCapacity 测试零容量令牌桶
func TestTokenBucketZeroCapacity(t *testing.T) {
	bucket := NewTokenBucket(0, 1)

	// 零容量桶不应该允许任何请求
	if bucket.Allow() {
		t.Error("零容量桶不应该允许请求")
	}
}

// TestLeakyBucketZeroCapacity 测试零容量漏桶
func TestLeakyBucketZeroCapacity(t *testing.T) {
	bucket := NewLeakyBucket(0, 1)

	// 零容量桶不应该允许任何请求
	if bucket.Allow() {
		t.Error("零容量桶不应该允许请求")
	}
}

// TestSlidingWindowZeroLimit 测试零限制滑动窗口
func TestSlidingWindowZeroLimit(t *testing.T) {
	window := NewSlidingWindow(time.Second, 0)

	// 零限制窗口不应该允许任何请求
	if window.Allow() {
		t.Error("零限制窗口不应该允许请求")
	}
}

// TestTokenBucketLargeRequest 测试大量令牌请求
func TestTokenBucketLargeRequest(t *testing.T) {
	bucket := NewTokenBucket(10, 1)

	// 请求超过容量的令牌应该失败
	if bucket.AllowN(100) {
		t.Error("请求超过容量的令牌应该失败")
	}

	// 令牌数不应该变成负数
	if bucket.GetTokens() < 0 {
		t.Error("令牌数不应该变成负数")
	}
}

// TestLeakyBucketLargeRequest 测试大量漏桶请求
func TestLeakyBucketLargeRequest(t *testing.T) {
	bucket := NewLeakyBucket(10, 1)

	// 请求超过容量应该失败
	if bucket.AllowN(100) {
		t.Error("请求超过容量应该失败")
	}

	// 水位不应该超过容量
	if bucket.GetLevel() > 10 {
		t.Error("水位不应该超过容量")
	}
}

// =============================================================================
// 压力测试
// =============================================================================

// TestTokenBucketStress 压力测试令牌桶
func TestTokenBucketStress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}

	bucket := NewTokenBucket(10000, 1000)

	const numGoroutines = 100
	const numRequests = 1000
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				bucket.Allow()
			}
		}()
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(30 * time.Second):
		t.Fatal("压力测试超时")
	}
}

// TestSlidingWindowStress 压力测试滑动窗口
func TestSlidingWindowStress(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过压力测试")
	}

	window := NewSlidingWindow(10*time.Second, 100000)

	const numGoroutines = 50
	const numRequests = 500
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				window.Allow()
			}
		}()
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(30 * time.Second):
		t.Fatal("压力测试超时")
	}
}

// =============================================================================
// 算法对比测试
// =============================================================================

// TestRateLimiterComparison 对比不同限流算法
func TestRateLimiterComparison(t *testing.T) {
	const limit int64 = 100
	const requests = 150

	// 令牌桶
	tokenBucket := NewTokenBucket(limit, limit)
	tokenAllowed := 0
	for i := 0; i < requests; i++ {
		if tokenBucket.Allow() {
			tokenAllowed++
		}
	}

	// 漏桶
	leakyBucket := NewLeakyBucket(limit, limit)
	leakyAllowed := 0
	for i := 0; i < requests; i++ {
		if leakyBucket.Allow() {
			leakyAllowed++
		}
	}

	// 滑动窗口
	slidingWindow := NewSlidingWindow(time.Hour, limit)
	windowAllowed := 0
	for i := 0; i < requests; i++ {
		if slidingWindow.Allow() {
			windowAllowed++
		}
	}

	t.Logf("令牌桶允许: %d/%d", tokenAllowed, requests)
	t.Logf("漏桶允许: %d/%d", leakyAllowed, requests)
	t.Logf("滑动窗口允许: %d/%d", windowAllowed, requests)

	// 所有算法允许的请求数都不应该超过限制
	if int64(tokenAllowed) > limit {
		t.Errorf("令牌桶允许数 %d 超过限制 %d", tokenAllowed, limit)
	}
	if int64(leakyAllowed) > limit {
		t.Errorf("漏桶允许数 %d 超过限制 %d", leakyAllowed, limit)
	}
	if int64(windowAllowed) > limit {
		t.Errorf("滑动窗口允许数 %d 超过限制 %d", windowAllowed, limit)
	}
}
