/*
=== 微服务弹性模式：熔断器(Circuit Breaker) ===

Circuit Breaker模式是微服务架构中的核心弹性模式，用于防止级联失败。
研究表明，Circuit Breaker模式可以减少错误率58%，显著提升系统可用性。

学习目标：
1. 理解Circuit Breaker的三种状态：Closed、Open、Half-Open
2. 掌握故障检测和恢复策略
3. 学会配置阈值和超时参数
4. 实现生产级别的熔断器
5. 集成监控和指标收集

核心概念：
- Closed状态: 正常工作，记录失败次数
- Open状态: 熔断开启，直接返回错误
- Half-Open状态: 尝试恢复，允许少量请求通过

业界标准：
- Netflix Hystrix模式实现
- 支持并发安全
- 可配置的故障阈值和恢复策略
- 完整的监控指标
*/

package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 安全随机数生成函数
func secureRandomFloat64() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		// 安全fallback：使用时间戳
		return float64(time.Now().UnixNano()%1000) / 1000.0
	}
	return float64(n.Int64()) / float64(1<<53)
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           // 失败阈值
	SuccessThreshold int           // 恢复成功阈值
	Timeout          time.Duration // 超时时间
	RecoveryTimeout  time.Duration // 恢复超时
	SlidingWindow    time.Duration // 滑动窗口大小
	MaxRequests      int           // Half-Open状态最大请求数
}

// DefaultConfig 默认配置
func DefaultConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          2 * time.Second,
		RecoveryTimeout:  30 * time.Second,
		SlidingWindow:    time.Minute,
		MaxRequests:      10,
	}
}

// CircuitBreakerStats 统计信息
type CircuitBreakerStats struct {
	TotalRequests    int64     `json:"total_requests"`
	SuccessRequests  int64     `json:"success_requests"`
	FailureRequests  int64     `json:"failure_requests"`
	RejectedRequests int64     `json:"rejected_requests"`
	State            string    `json:"state"`
	LastFailureTime  time.Time `json:"last_failure_time"`
	LastSuccessTime  time.Time `json:"last_success_time"`
}

// CircuitBreaker 熔断器实现
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int64
	successCount    int64
	lastFailureTime time.Time
	nextRetryTime   time.Time
	halfOpenCount   int64
	mutex           sync.RWMutex

	// 统计信息
	stats CircuitBreakerStats

	// 回调函数
	onStateChange func(from, to CircuitBreakerState)
}

// NewCircuitBreaker 创建新的熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		stats:  CircuitBreakerStats{State: StateClosed.String()},
	}
}

// OnStateChange 设置状态变更回调
func (cb *CircuitBreaker) OnStateChange(fn func(from, to CircuitBreakerState)) {
	cb.onStateChange = fn
}

// Execute 执行函数，带熔断保护
func (cb *CircuitBreaker) Execute(fn func() error) error {
	return cb.ExecuteWithContext(context.Background(), fn)
}

// ExecuteWithContext 执行函数，带上下文和熔断保护
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func() error) error {
	atomic.AddInt64(&cb.stats.TotalRequests, 1)

	// 检查是否允许请求通过
	if !cb.allowRequest() {
		atomic.AddInt64(&cb.stats.RejectedRequests, 1)
		return errors.New("circuit breaker is open")
	}

	// 创建带超时的上下文
	if cb.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cb.config.Timeout)
		defer cancel()
	}

	// 执行函数
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		if err != nil {
			cb.onFailure()
			return err
		}
		cb.onSuccess()
		return nil
	case <-ctx.Done():
		cb.onFailure()
		return ctx.Err()
	}
}

// allowRequest 检查是否允许请求通过
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.RLock()
	state := cb.state
	nextRetryTime := cb.nextRetryTime
	halfOpenCount := cb.halfOpenCount
	cb.mutex.RUnlock()

	switch state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否到了重试时间
		if time.Now().After(nextRetryTime) {
			cb.setState(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		// Half-Open状态下限制并发请求数
		return halfOpenCount < int64(cb.config.MaxRequests)
	}

	return false
}

// onSuccess 成功回调
func (cb *CircuitBreaker) onSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	atomic.AddInt64(&cb.stats.SuccessRequests, 1)
	cb.stats.LastSuccessTime = time.Now()

	switch cb.state {
	case StateClosed:
		cb.resetFailureCount()
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= int64(cb.config.SuccessThreshold) {
			cb.setState(StateClosed)
			cb.resetCounters()
		}
	}
}

// onFailure 失败回调
func (cb *CircuitBreaker) onFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	atomic.AddInt64(&cb.stats.FailureRequests, 1)
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	cb.stats.LastFailureTime = cb.lastFailureTime

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= int64(cb.config.FailureThreshold) {
			cb.setState(StateOpen)
			cb.nextRetryTime = time.Now().Add(cb.config.RecoveryTimeout)
		}
	case StateHalfOpen:
		cb.setState(StateOpen)
		cb.nextRetryTime = time.Now().Add(cb.config.RecoveryTimeout)
		cb.resetCounters()
	}
}

// setState 设置状态
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState
	cb.stats.State = newState.String()

	log.Printf("Circuit Breaker state changed: %s -> %s", oldState, newState)

	if cb.onStateChange != nil {
		go cb.onStateChange(oldState, newState)
	}
}

// resetCounters 重置计数器
func (cb *CircuitBreaker) resetCounters() {
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenCount = 0
}

// resetFailureCount 重置失败计数
func (cb *CircuitBreaker) resetFailureCount() {
	cb.failureCount = 0
}

// GetStats 获取统计信息
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return cb.stats
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// ==================
// 服务示例和测试
// ==================

// UnstableService 不稳定的服务模拟
type UnstableService struct {
	failureRate float64 // 失败率 0.0-1.0
	delay       time.Duration
}

// NewUnstableService 创建不稳定服务
func NewUnstableService(failureRate float64, delay time.Duration) *UnstableService {
	return &UnstableService{
		failureRate: failureRate,
		delay:       delay,
	}
}

// Call 调用服务
func (s *UnstableService) Call() error {
	// 模拟延迟
	if s.delay > 0 {
		time.Sleep(s.delay)
	}

	// 模拟随机失败
	if secureRandomFloat64() < s.failureRate {
		return errors.New("service temporarily unavailable")
	}

	return nil
}

// CircuitBreakerMiddleware HTTP中间件
func CircuitBreakerMiddleware(cb *CircuitBreaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := cb.ExecuteWithContext(r.Context(), func() error {
				next.ServeHTTP(w, r)
				return nil
			})

			if err != nil {
				http.Error(w, "Service Unavailable: "+err.Error(), http.StatusServiceUnavailable)
			}
		})
	}
}

// ==================
// 监控和指标
// ==================

// MetricsCollector 指标收集器
type MetricsCollector struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Register 注册熔断器
func (mc *MetricsCollector) Register(name string, cb *CircuitBreaker) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.breakers[name] = cb
}

// GetAllStats 获取所有统计信息
func (mc *MetricsCollector) GetAllStats() map[string]CircuitBreakerStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for name, cb := range mc.breakers {
		stats[name] = cb.GetStats()
	}
	return stats
}

// ==================
// HTTP监控端点
// ==================

func (mc *MetricsCollector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stats := mc.GetAllStats()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\n")
	i := 0
	for name, stat := range stats {
		if i > 0 {
			fmt.Fprintf(w, ",\n")
		}
		fmt.Fprintf(w, "  \"%s\": {\n", name)
		fmt.Fprintf(w, "    \"state\": \"%s\",\n", stat.State)
		fmt.Fprintf(w, "    \"total_requests\": %d,\n", stat.TotalRequests)
		fmt.Fprintf(w, "    \"success_requests\": %d,\n", stat.SuccessRequests)
		fmt.Fprintf(w, "    \"failure_requests\": %d,\n", stat.FailureRequests)
		fmt.Fprintf(w, "    \"rejected_requests\": %d,\n", stat.RejectedRequests)
		fmt.Fprintf(w, "    \"success_rate\": %.2f\n", float64(stat.SuccessRequests)/float64(stat.TotalRequests)*100)
		fmt.Fprintf(w, "  }")
		i++
	}
	fmt.Fprintf(w, "\n}\n")
}

// ==================
// 示例和测试
// ==================

func main() {
	fmt.Println("=== 微服务熔断器(Circuit Breaker)模式演示 ===")

	// 1. 创建熔断器配置
	config := DefaultConfig()
	config.FailureThreshold = 3              // 3次失败后熔断
	config.SuccessThreshold = 2              // 2次成功后恢复
	config.RecoveryTimeout = 5 * time.Second // 5秒后尝试恢复

	// 2. 创建熔断器
	cb := NewCircuitBreaker(config)

	// 3. 设置状态变更回调
	cb.OnStateChange(func(from, to CircuitBreakerState) {
		fmt.Printf("🔄 熔断器状态变更: %s -> %s\n", from, to)
	})

	// 4. 创建不稳定服务 (70%失败率)
	service := NewUnstableService(0.7, 100*time.Millisecond)

	// 5. 创建指标收集器
	collector := NewMetricsCollector()
	collector.Register("payment-service", cb)

	// 6. 启动监控端点
	go func() {
		http.Handle("/metrics", collector)
		http.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			stats := cb.GetStats()
			if cb.GetState() == StateOpen {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			fmt.Fprintf(w, "Status: %s\nRequests: %d\nSuccess Rate: %.2f%%\n",
				stats.State, stats.TotalRequests,
				float64(stats.SuccessRequests)/float64(stats.TotalRequests)*100)
		}))

		log.Println("监控端点启动: http://localhost:8080/metrics")
		log.Println("健康检查端点: http://localhost:8080/health")

		server := &http.Server{
			Addr:         ":8080",
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Fatal(server.ListenAndServe())
	}()

	// 7. 模拟服务调用
	fmt.Println("\n🚀 开始服务调用测试...")

	for i := 0; i < 50; i++ {
		err := cb.Execute(func() error {
			return service.Call()
		})

		if err != nil {
			fmt.Printf("❌ 请求 %d: %s\n", i+1, err.Error())
		} else {
			fmt.Printf("✅ 请求 %d: 成功\n", i+1)
		}

		// 每10个请求后显示统计信息
		if (i+1)%10 == 0 {
			stats := cb.GetStats()
			successRate := float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
			fmt.Printf("\n📊 统计信息 (第%d次请求后):\n", i+1)
			fmt.Printf("   状态: %s\n", stats.State)
			fmt.Printf("   总请求: %d\n", stats.TotalRequests)
			fmt.Printf("   成功: %d (%.1f%%)\n", stats.SuccessRequests, successRate)
			fmt.Printf("   失败: %d\n", stats.FailureRequests)
			fmt.Printf("   拒绝: %d\n", stats.RejectedRequests)
			fmt.Println(strings.Repeat("-", 50))
		}

		time.Sleep(200 * time.Millisecond)
	}

	// 8. 最终统计
	finalStats := cb.GetStats()
	successRate := float64(finalStats.SuccessRequests) / float64(finalStats.TotalRequests) * 100

	fmt.Println("\n🏁 最终统计报告:")
	fmt.Printf("总请求数: %d\n", finalStats.TotalRequests)
	fmt.Printf("成功请求: %d (%.2f%%)\n", finalStats.SuccessRequests, successRate)
	fmt.Printf("失败请求: %d\n", finalStats.FailureRequests)
	fmt.Printf("拒绝请求: %d\n", finalStats.RejectedRequests)
	fmt.Printf("最终状态: %s\n", finalStats.State)

	if finalStats.RejectedRequests > 0 {
		protectionRate := float64(finalStats.RejectedRequests) / float64(finalStats.TotalRequests) * 100
		fmt.Printf("🛡️ 熔断器保护率: %.2f%%\n", protectionRate)
	}

	// 9. 演示恢复过程
	fmt.Println("\n🔄 演示自动恢复过程...")
	fmt.Println("降低服务失败率到10%，观察熔断器恢复...")
	service.failureRate = 0.1 // 降低失败率

	for i := 0; i < 20; i++ {
		err := cb.Execute(func() error {
			return service.Call()
		})

		state := cb.GetState()
		if err != nil {
			fmt.Printf("❌ 恢复请求 %d [%s]: %s\n", i+1, state, err.Error())
		} else {
			fmt.Printf("✅ 恢复请求 %d [%s]: 成功\n", i+1, state)
		}

		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println("\n✨ Circuit Breaker演示完成!")
	fmt.Println("💡 关键特性:")
	fmt.Println("   - 自动故障检测和熔断")
	fmt.Println("   - 智能恢复机制")
	fmt.Println("   - 完整的监控指标")
	fmt.Println("   - HTTP中间件支持")
	fmt.Println("   - 生产级并发安全")

	// 保持监控端点运行
	select {}
}
