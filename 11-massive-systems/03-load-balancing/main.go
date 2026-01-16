// Package main 演示负载均衡策略
// 本模块涵盖常见的负载均衡算法：
// - 轮询（Round Robin）
// - 加权轮询（Weighted Round Robin）
// - 最少连接（Least Connections）
// - 一致性哈希（Consistent Hashing）
// - 自适应负载均衡
package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// 后端服务器定义
// ============================================================================

// Backend 后端服务器
type Backend struct {
	ID              string
	Address         string
	Port            int
	Weight          int
	CurrentWeight   int
	ActiveConns     int64
	TotalRequests   int64
	FailedRequests  int64
	ResponseTime    time.Duration
	IsHealthy       bool
	LastHealthCheck time.Time
	mu              sync.RWMutex
}

// NewBackend 创建后端服务器
func NewBackend(id, address string, port, weight int) *Backend {
	return &Backend{
		ID:        id,
		Address:   address,
		Port:      port,
		Weight:    weight,
		IsHealthy: true,
	}
}

// IncrementConns 增加连接数
func (b *Backend) IncrementConns() {
	atomic.AddInt64(&b.ActiveConns, 1)
	atomic.AddInt64(&b.TotalRequests, 1)
}

// DecrementConns 减少连接数
func (b *Backend) DecrementConns() {
	atomic.AddInt64(&b.ActiveConns, -1)
}

// GetActiveConns 获取活跃连接数
func (b *Backend) GetActiveConns() int64 {
	return atomic.LoadInt64(&b.ActiveConns)
}

// String 字符串表示
func (b *Backend) String() string {
	return fmt.Sprintf("%s (%s:%d, weight=%d, conns=%d)",
		b.ID, b.Address, b.Port, b.Weight, b.GetActiveConns())
}

// ============================================================================
// 负载均衡器接口
// ============================================================================

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Name() string
	Select(backends []*Backend, key string) *Backend
	UpdateStats(backend *Backend, responseTime time.Duration, success bool)
}

// ============================================================================
// 轮询负载均衡
// ============================================================================

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	current uint64
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{}
}

func (r *RoundRobinBalancer) Name() string { return "Round Robin" }

func (r *RoundRobinBalancer) Select(backends []*Backend, _ string) *Backend {
	if len(backends) == 0 {
		return nil
	}

	// 过滤健康的后端
	healthy := filterHealthy(backends)
	if len(healthy) == 0 {
		return nil
	}

	// 原子递增并取模
	idx := atomic.AddUint64(&r.current, 1) % uint64(len(healthy))
	return healthy[idx]
}

func (r *RoundRobinBalancer) UpdateStats(_ *Backend, _ time.Duration, _ bool) {}

// ============================================================================
// 加权轮询负载均衡
// ============================================================================

// WeightedRoundRobinBalancer 加权轮询负载均衡器
type WeightedRoundRobinBalancer struct {
	mu sync.Mutex
}

// NewWeightedRoundRobinBalancer 创建加权轮询负载均衡器
func NewWeightedRoundRobinBalancer() *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{}
}

func (w *WeightedRoundRobinBalancer) Name() string { return "Weighted Round Robin" }

func (w *WeightedRoundRobinBalancer) Select(backends []*Backend, _ string) *Backend {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(backends) == 0 {
		return nil
	}

	healthy := filterHealthy(backends)
	if len(healthy) == 0 {
		return nil
	}

	// Nginx 平滑加权轮询算法
	var best *Backend
	totalWeight := 0

	for _, b := range healthy {
		b.mu.Lock()
		b.CurrentWeight += b.Weight
		totalWeight += b.Weight

		if best == nil || b.CurrentWeight > best.CurrentWeight {
			best = b
		}
		b.mu.Unlock()
	}

	if best != nil {
		best.mu.Lock()
		best.CurrentWeight -= totalWeight
		best.mu.Unlock()
	}

	return best
}

func (w *WeightedRoundRobinBalancer) UpdateStats(_ *Backend, _ time.Duration, _ bool) {}

// ============================================================================
// 最少连接负载均衡
// ============================================================================

// LeastConnectionsBalancer 最少连接负载均衡器
type LeastConnectionsBalancer struct{}

// NewLeastConnectionsBalancer 创建最少连接负载均衡器
func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{}
}

func (l *LeastConnectionsBalancer) Name() string { return "Least Connections" }

func (l *LeastConnectionsBalancer) Select(backends []*Backend, _ string) *Backend {
	if len(backends) == 0 {
		return nil
	}

	healthy := filterHealthy(backends)
	if len(healthy) == 0 {
		return nil
	}

	var best *Backend
	minConns := int64(-1)

	for _, b := range healthy {
		conns := b.GetActiveConns()
		// 考虑权重：有效连接数 = 实际连接数 / 权重
		effectiveConns := conns * 100 / int64(b.Weight)

		if minConns < 0 || effectiveConns < minConns {
			minConns = effectiveConns
			best = b
		}
	}

	return best
}

func (l *LeastConnectionsBalancer) UpdateStats(_ *Backend, _ time.Duration, _ bool) {}

// ============================================================================
// 一致性哈希负载均衡
// ============================================================================

// ConsistentHashBalancer 一致性哈希负载均衡器
type ConsistentHashBalancer struct {
	ring         []uint32
	nodes        map[uint32]*Backend
	virtualNodes int
	mu           sync.RWMutex
}

// NewConsistentHashBalancer 创建一致性哈希负载均衡器
func NewConsistentHashBalancer(virtualNodes int) *ConsistentHashBalancer {
	return &ConsistentHashBalancer{
		ring:         make([]uint32, 0),
		nodes:        make(map[uint32]*Backend),
		virtualNodes: virtualNodes,
	}
}

func (c *ConsistentHashBalancer) Name() string { return "Consistent Hash" }

// hash 计算哈希值
func (c *ConsistentHashBalancer) hash(key string) uint32 {
	h := md5.Sum([]byte(key))
	return binary.BigEndian.Uint32(h[:4])
}

// AddNode 添加节点到哈希环
func (c *ConsistentHashBalancer) AddNode(backend *Backend) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.virtualNodes; i++ {
		virtualKey := fmt.Sprintf("%s#%d", backend.ID, i)
		hash := c.hash(virtualKey)
		c.ring = append(c.ring, hash)
		c.nodes[hash] = backend
	}

	sort.Slice(c.ring, func(i, j int) bool {
		return c.ring[i] < c.ring[j]
	})
}

// RemoveNode 从哈希环移除节点
func (c *ConsistentHashBalancer) RemoveNode(backend *Backend) {
	c.mu.Lock()
	defer c.mu.Unlock()

	newRing := make([]uint32, 0)
	for _, hash := range c.ring {
		if c.nodes[hash].ID != backend.ID {
			newRing = append(newRing, hash)
		} else {
			delete(c.nodes, hash)
		}
	}
	c.ring = newRing
}

func (c *ConsistentHashBalancer) Select(backends []*Backend, key string) *Backend {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.ring) == 0 {
		return nil
	}

	hash := c.hash(key)

	// 二分查找第一个大于等于 hash 的节点
	idx := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i] >= hash
	})

	// 如果没找到，回到环的开头
	if idx >= len(c.ring) {
		idx = 0
	}

	backend := c.nodes[c.ring[idx]]

	// 如果节点不健康，顺时针查找下一个健康节点
	if !backend.IsHealthy {
		for i := 1; i < len(c.ring); i++ {
			nextIdx := (idx + i) % len(c.ring)
			nextBackend := c.nodes[c.ring[nextIdx]]
			if nextBackend.IsHealthy {
				return nextBackend
			}
		}
		return nil
	}

	return backend
}

func (c *ConsistentHashBalancer) UpdateStats(_ *Backend, _ time.Duration, _ bool) {}

// ============================================================================
// 自适应负载均衡
// ============================================================================

// AdaptiveBalancer 自适应负载均衡器
type AdaptiveBalancer struct {
	stats map[string]*BackendStats
	mu    sync.RWMutex
}

// BackendStats 后端统计信息
type BackendStats struct {
	AvgResponseTime time.Duration
	SuccessRate     float64
	RequestCount    int64
	SuccessCount    int64
	Score           float64
}

// NewAdaptiveBalancer 创建自适应负载均衡器
func NewAdaptiveBalancer() *AdaptiveBalancer {
	return &AdaptiveBalancer{
		stats: make(map[string]*BackendStats),
	}
}

func (a *AdaptiveBalancer) Name() string { return "Adaptive" }

func (a *AdaptiveBalancer) Select(backends []*Backend, _ string) *Backend {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(backends) == 0 {
		return nil
	}

	healthy := filterHealthy(backends)
	if len(healthy) == 0 {
		return nil
	}

	// 计算每个后端的得分
	var best *Backend
	bestScore := float64(-1)

	for _, b := range healthy {
		score := a.calculateScore(b)
		if bestScore < 0 || score > bestScore {
			bestScore = score
			best = b
		}
	}

	return best
}

func (a *AdaptiveBalancer) calculateScore(backend *Backend) float64 {
	stats, exists := a.stats[backend.ID]
	if !exists {
		// 新节点给予较高初始分数
		return float64(backend.Weight) * 100
	}

	// 综合考虑：权重、成功率、响应时间、当前连接数
	weightScore := float64(backend.Weight)
	successScore := stats.SuccessRate * 100
	latencyScore := 100.0 / (1.0 + float64(stats.AvgResponseTime.Milliseconds())/100.0)
	connScore := 100.0 / (1.0 + float64(backend.GetActiveConns())/10.0)

	// 加权平均
	return weightScore*0.2 + successScore*0.3 + latencyScore*0.3 + connScore*0.2
}

func (a *AdaptiveBalancer) UpdateStats(backend *Backend, responseTime time.Duration, success bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	stats, exists := a.stats[backend.ID]
	if !exists {
		stats = &BackendStats{}
		a.stats[backend.ID] = stats
	}

	stats.RequestCount++
	if success {
		stats.SuccessCount++
	}

	// 指数移动平均更新响应时间
	alpha := 0.3
	if stats.AvgResponseTime == 0 {
		stats.AvgResponseTime = responseTime
	} else {
		stats.AvgResponseTime = time.Duration(
			float64(stats.AvgResponseTime)*(1-alpha) + float64(responseTime)*alpha,
		)
	}

	stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.RequestCount)
	stats.Score = a.calculateScore(backend)
}

// GetStats 获取统计信息
func (a *AdaptiveBalancer) GetStats() map[string]*BackendStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make(map[string]*BackendStats)
	for k, v := range a.stats {
		result[k] = &BackendStats{
			AvgResponseTime: v.AvgResponseTime,
			SuccessRate:     v.SuccessRate,
			RequestCount:    v.RequestCount,
			SuccessCount:    v.SuccessCount,
			Score:           v.Score,
		}
	}
	return result
}

// ============================================================================
// 负载均衡器管理器
// ============================================================================

// LoadBalancerManager 负载均衡器管理器
type LoadBalancerManager struct {
	backends []*Backend
	balancer LoadBalancer
	mu       sync.RWMutex
}

// NewLoadBalancerManager 创建负载均衡器管理器
func NewLoadBalancerManager(balancer LoadBalancer) *LoadBalancerManager {
	return &LoadBalancerManager{
		backends: make([]*Backend, 0),
		balancer: balancer,
	}
}

// AddBackend 添加后端
func (m *LoadBalancerManager) AddBackend(backend *Backend) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.backends = append(m.backends, backend)

	// 如果是一致性哈希，需要添加到哈希环
	if ch, ok := m.balancer.(*ConsistentHashBalancer); ok {
		ch.AddNode(backend)
	}
}

// RemoveBackend 移除后端
func (m *LoadBalancerManager) RemoveBackend(backendID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, b := range m.backends {
		if b.ID == backendID {
			// 如果是一致性哈希，需要从哈希环移除
			if ch, ok := m.balancer.(*ConsistentHashBalancer); ok {
				ch.RemoveNode(b)
			}

			m.backends = append(m.backends[:i], m.backends[i+1:]...)
			break
		}
	}
}

// Select 选择后端
func (m *LoadBalancerManager) Select(key string) *Backend {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.balancer.Select(m.backends, key)
}

// SimulateRequest 模拟请求
func (m *LoadBalancerManager) SimulateRequest(key string) {
	backend := m.Select(key)
	if backend == nil {
		fmt.Println("  没有可用的后端服务器")
		return
	}

	backend.IncrementConns()

	// 模拟请求处理
	responseTime := time.Duration(rand.Intn(100)+10) * time.Millisecond
	time.Sleep(10 * time.Millisecond) // 简化模拟

	success := rand.Float32() > 0.05 // 95% 成功率

	backend.DecrementConns()

	m.balancer.UpdateStats(backend, responseTime, success)

	status := "成功"
	if !success {
		status = "失败"
	}

	fmt.Printf("  请求 -> %s [%s] (响应: %v)\n", backend.ID, status, responseTime)
}

// GetBackends 获取所有后端
func (m *LoadBalancerManager) GetBackends() []*Backend {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]*Backend{}, m.backends...)
}

// ============================================================================
// 辅助函数
// ============================================================================

func filterHealthy(backends []*Backend) []*Backend {
	healthy := make([]*Backend, 0)
	for _, b := range backends {
		if b.IsHealthy {
			healthy = append(healthy, b)
		}
	}
	return healthy
}

// ============================================================================
// 演示函数
// ============================================================================

func createBackends() []*Backend {
	return []*Backend{
		NewBackend("server-1", "10.0.1.1", 8080, 5),
		NewBackend("server-2", "10.0.1.2", 8080, 3),
		NewBackend("server-3", "10.0.1.3", 8080, 2),
	}
}

func demonstrateRoundRobin() {
	fmt.Println("\n=== 轮询负载均衡演示 ===")
	fmt.Println("特点: 简单公平，依次分配请求")

	manager := NewLoadBalancerManager(NewRoundRobinBalancer())
	for _, b := range createBackends() {
		manager.AddBackend(b)
	}

	fmt.Println("\n发送 9 个请求:")
	for i := 0; i < 9; i++ {
		manager.SimulateRequest("")
	}

	fmt.Println("\n请求分布:")
	for _, b := range manager.GetBackends() {
		fmt.Printf("  %s: %d 个请求\n", b.ID, b.TotalRequests)
	}
}

func demonstrateWeightedRoundRobin() {
	fmt.Println("\n=== 加权轮询负载均衡演示 ===")
	fmt.Println("特点: 根据权重分配，高性能服务器处理更多请求")

	manager := NewLoadBalancerManager(NewWeightedRoundRobinBalancer())
	for _, b := range createBackends() {
		manager.AddBackend(b)
	}

	fmt.Println("\n服务器权重:")
	for _, b := range manager.GetBackends() {
		fmt.Printf("  %s: 权重 %d\n", b.ID, b.Weight)
	}

	fmt.Println("\n发送 10 个请求:")
	for i := 0; i < 10; i++ {
		manager.SimulateRequest("")
	}

	fmt.Println("\n请求分布 (应接近权重比例 5:3:2):")
	for _, b := range manager.GetBackends() {
		fmt.Printf("  %s: %d 个请求\n", b.ID, b.TotalRequests)
	}
}

func demonstrateLeastConnections() {
	fmt.Println("\n=== 最少连接负载均衡演示 ===")
	fmt.Println("特点: 优先选择连接数最少的服务器")

	manager := NewLoadBalancerManager(NewLeastConnectionsBalancer())
	backends := createBackends()
	for _, b := range backends {
		manager.AddBackend(b)
	}

	// 模拟不同的初始连接数
	backends[0].ActiveConns = 10
	backends[1].ActiveConns = 5
	backends[2].ActiveConns = 2

	fmt.Println("\n初始连接状态:")
	for _, b := range manager.GetBackends() {
		fmt.Printf("  %s: %d 个活跃连接\n", b.ID, b.GetActiveConns())
	}

	fmt.Println("\n发送 6 个请求 (应优先选择连接少的服务器):")
	for i := 0; i < 6; i++ {
		manager.SimulateRequest("")
	}
}

func demonstrateConsistentHash() {
	fmt.Println("\n=== 一致性哈希负载均衡演示 ===")
	fmt.Println("特点: 相同 key 总是路由到相同服务器，适合缓存场景")

	balancer := NewConsistentHashBalancer(100) // 100 个虚拟节点
	manager := NewLoadBalancerManager(balancer)

	for _, b := range createBackends() {
		manager.AddBackend(b)
	}

	// 测试相同 key 的路由一致性
	testKeys := []string{"user:1001", "user:1002", "user:1003", "order:5001", "order:5002"}

	fmt.Println("\n第一轮请求:")
	firstRound := make(map[string]string)
	for _, key := range testKeys {
		backend := manager.Select(key)
		firstRound[key] = backend.ID
		fmt.Printf("  %s -> %s\n", key, backend.ID)
	}

	fmt.Println("\n第二轮请求 (相同 key 应路由到相同服务器):")
	for _, key := range testKeys {
		backend := manager.Select(key)
		match := "匹配"
		if firstRound[key] != backend.ID {
			match = "不匹配"
		}
		fmt.Printf("  %s -> %s [%s]\n", key, backend.ID, match)
	}

	// 模拟节点下线
	fmt.Println("\n模拟 server-2 下线:")
	for _, b := range manager.GetBackends() {
		if b.ID == "server-2" {
			b.IsHealthy = false
			break
		}
	}

	fmt.Println("节点下线后的路由 (只有原本路由到 server-2 的 key 会变化):")
	for _, key := range testKeys {
		backend := manager.Select(key)
		change := ""
		if firstRound[key] == "server-2" && backend.ID != "server-2" {
			change = " [已迁移]"
		}
		fmt.Printf("  %s -> %s%s\n", key, backend.ID, change)
	}
}

func demonstrateAdaptive() {
	fmt.Println("\n=== 自适应负载均衡演示 ===")
	fmt.Println("特点: 根据实时性能指标动态调整路由")

	balancer := NewAdaptiveBalancer()
	manager := NewLoadBalancerManager(balancer)

	backends := createBackends()
	for _, b := range backends {
		manager.AddBackend(b)
	}

	fmt.Println("\n发送 20 个请求进行学习:")
	for i := 0; i < 20; i++ {
		manager.SimulateRequest("")
	}

	fmt.Println("\n自适应统计信息:")
	stats := balancer.GetStats()
	for id, s := range stats {
		fmt.Printf("  %s:\n", id)
		fmt.Printf("    请求数: %d\n", s.RequestCount)
		fmt.Printf("    成功率: %.2f%%\n", s.SuccessRate*100)
		fmt.Printf("    平均响应: %v\n", s.AvgResponseTime)
		fmt.Printf("    综合得分: %.2f\n", s.Score)
	}

	// 模拟一个服务器性能下降
	fmt.Println("\n模拟 server-1 性能下降 (响应变慢):")
	for i := 0; i < 10; i++ {
		backend := manager.Select("")
		if backend.ID == "server-1" {
			// 模拟慢响应
			balancer.UpdateStats(backend, 500*time.Millisecond, true)
		} else {
			balancer.UpdateStats(backend, 50*time.Millisecond, true)
		}
	}

	fmt.Println("\n更新后的统计信息:")
	stats = balancer.GetStats()
	for id, s := range stats {
		fmt.Printf("  %s: 响应=%v, 得分=%.2f\n", id, s.AvgResponseTime, s.Score)
	}
}

func main() {
	fmt.Println("=== 负载均衡策略 ===")
	fmt.Println()
	fmt.Println("本模块演示五种常见的负载均衡算法:")
	fmt.Println("1. 轮询 - 简单公平的请求分配")
	fmt.Println("2. 加权轮询 - 根据服务器能力分配")
	fmt.Println("3. 最少连接 - 优先选择空闲服务器")
	fmt.Println("4. 一致性哈希 - 保证路由一致性")
	fmt.Println("5. 自适应 - 根据实时性能动态调整")

	demonstrateRoundRobin()
	demonstrateWeightedRoundRobin()
	demonstrateLeastConnections()
	demonstrateConsistentHash()
	demonstrateAdaptive()

	fmt.Println("\n=== 负载均衡演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 轮询: 实现简单，适合同质化服务器")
	fmt.Println("- 加权轮询: Nginx 平滑加权算法，避免请求突发")
	fmt.Println("- 最少连接: 适合长连接场景，考虑权重")
	fmt.Println("- 一致性哈希: 虚拟节点解决数据倾斜，适合缓存")
	fmt.Println("- 自适应: 综合多维指标，动态优化路由决策")
}
