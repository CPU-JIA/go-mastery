// Package main 演示高可用架构设计
// 本模块涵盖高可用系统的核心组件：
// - 健康检查机制
// - 故障检测与自动恢复
// - 主从切换（Failover）
// - 多活架构
package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// ============================================================================
// 健康检查机制
// ============================================================================

// HealthStatus 健康状态
type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusDegraded
	HealthStatusUnhealthy
	HealthStatusUnknown
)

func (s HealthStatus) String() string {
	switch s {
	case HealthStatusHealthy:
		return "Healthy"
	case HealthStatusDegraded:
		return "Degraded"
	case HealthStatusUnhealthy:
		return "Unhealthy"
	default:
		return "Unknown"
	}
}

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check(ctx context.Context) HealthCheckResult
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Status    HealthStatus
	Message   string
	Latency   time.Duration
	Timestamp time.Time
	Details   map[string]interface{}
}

// CompositeHealthChecker 组合健康检查器
type CompositeHealthChecker struct {
	checks  []HealthCheck
	timeout time.Duration
	mu      sync.RWMutex
	results map[string]HealthCheckResult
	history []HealthCheckResult
	maxHist int
}

// NewCompositeHealthChecker 创建组合健康检查器
func NewCompositeHealthChecker(timeout time.Duration) *CompositeHealthChecker {
	return &CompositeHealthChecker{
		checks:  make([]HealthCheck, 0),
		timeout: timeout,
		results: make(map[string]HealthCheckResult),
		history: make([]HealthCheckResult, 0),
		maxHist: 100,
	}
}

// AddCheck 添加健康检查
func (c *CompositeHealthChecker) AddCheck(check HealthCheck) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks = append(c.checks, check)
}

// RunAll 运行所有健康检查
func (c *CompositeHealthChecker) RunAll(ctx context.Context) map[string]HealthCheckResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	results := make(map[string]HealthCheckResult)
	var wg sync.WaitGroup

	resultChan := make(chan struct {
		name   string
		result HealthCheckResult
	}, len(c.checks))

	for _, check := range c.checks {
		wg.Add(1)
		go func(hc HealthCheck) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, c.timeout)
			defer cancel()

			result := hc.Check(checkCtx)
			resultChan <- struct {
				name   string
				result HealthCheckResult
			}{hc.Name(), result}
		}(check)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		results[r.name] = r.result
		c.results[r.name] = r.result
	}

	return results
}

// GetOverallStatus 获取整体健康状态
func (c *CompositeHealthChecker) GetOverallStatus() HealthStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.results) == 0 {
		return HealthStatusUnknown
	}

	unhealthyCount := 0
	degradedCount := 0

	for _, result := range c.results {
		switch result.Status {
		case HealthStatusUnhealthy:
			unhealthyCount++
		case HealthStatusDegraded:
			degradedCount++
		}
	}

	if unhealthyCount > 0 {
		return HealthStatusUnhealthy
	}
	if degradedCount > 0 {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// DatabaseHealthCheck 数据库健康检查
type DatabaseHealthCheck struct {
	name     string
	simulate bool
}

func NewDatabaseHealthCheck(name string) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{name: name, simulate: true}
}

func (d *DatabaseHealthCheck) Name() string { return d.name }

func (d *DatabaseHealthCheck) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()

	// 模拟数据库连接检查
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	// 随机模拟不同状态
	status := HealthStatusHealthy
	message := "数据库连接正常"

	if rand.Float32() < 0.1 {
		status = HealthStatusDegraded
		message = "数据库响应较慢"
	}

	return HealthCheckResult{
		Status:    status,
		Message:   message,
		Latency:   time.Since(start),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"connections_active": rand.Intn(100),
			"connections_idle":   rand.Intn(50),
			"query_latency_ms":   rand.Intn(100),
		},
	}
}

// CacheHealthCheck 缓存健康检查
type CacheHealthCheck struct {
	name string
}

func NewCacheHealthCheck(name string) *CacheHealthCheck {
	return &CacheHealthCheck{name: name}
}

func (c *CacheHealthCheck) Name() string { return c.name }

func (c *CacheHealthCheck) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()

	// 模拟缓存检查
	time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)

	return HealthCheckResult{
		Status:    HealthStatusHealthy,
		Message:   "缓存服务正常",
		Latency:   time.Since(start),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"hit_rate":    0.95,
			"memory_used": "256MB",
			"keys_count":  rand.Intn(10000),
		},
	}
}

// ============================================================================
// 故障检测与自动恢复
// ============================================================================

// FailureDetector 故障检测器
type FailureDetector struct {
	nodes           map[string]*NodeState
	heartbeatPeriod time.Duration
	failureTimeout  time.Duration
	mu              sync.RWMutex
	onFailure       func(nodeID string)
	onRecovery      func(nodeID string)
}

// NodeState 节点状态
type NodeState struct {
	ID            string
	LastHeartbeat time.Time
	Status        NodeStatus
	FailureCount  int
	RecoveryCount int
	Metadata      map[string]string
}

// NodeStatus 节点状态枚举
type NodeStatus int

const (
	NodeStatusAlive NodeStatus = iota
	NodeStatusSuspected
	NodeStatusFailed
	NodeStatusRecovering
)

func (s NodeStatus) String() string {
	switch s {
	case NodeStatusAlive:
		return "Alive"
	case NodeStatusSuspected:
		return "Suspected"
	case NodeStatusFailed:
		return "Failed"
	case NodeStatusRecovering:
		return "Recovering"
	default:
		return "Unknown"
	}
}

// NewFailureDetector 创建故障检测器
func NewFailureDetector(heartbeatPeriod, failureTimeout time.Duration) *FailureDetector {
	return &FailureDetector{
		nodes:           make(map[string]*NodeState),
		heartbeatPeriod: heartbeatPeriod,
		failureTimeout:  failureTimeout,
	}
}

// RegisterNode 注册节点
func (fd *FailureDetector) RegisterNode(nodeID string, metadata map[string]string) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	fd.nodes[nodeID] = &NodeState{
		ID:            nodeID,
		LastHeartbeat: time.Now(),
		Status:        NodeStatusAlive,
		Metadata:      metadata,
	}
}

// Heartbeat 接收心跳
func (fd *FailureDetector) Heartbeat(nodeID string) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	if node, exists := fd.nodes[nodeID]; exists {
		node.LastHeartbeat = time.Now()

		if node.Status == NodeStatusFailed || node.Status == NodeStatusSuspected {
			node.Status = NodeStatusRecovering
			node.RecoveryCount++
			fmt.Printf("  [FailureDetector] 节点 %s 正在恢复\n", nodeID)

			if fd.onRecovery != nil {
				go fd.onRecovery(nodeID)
			}
		}

		if node.Status == NodeStatusRecovering {
			node.Status = NodeStatusAlive
			fmt.Printf("  [FailureDetector] 节点 %s 已恢复\n", nodeID)
		}
	}
}

// CheckNodes 检查所有节点状态
func (fd *FailureDetector) CheckNodes() {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	now := time.Now()

	for nodeID, node := range fd.nodes {
		elapsed := now.Sub(node.LastHeartbeat)

		switch node.Status {
		case NodeStatusAlive:
			if elapsed > fd.failureTimeout {
				node.Status = NodeStatusSuspected
				fmt.Printf("  [FailureDetector] 节点 %s 疑似故障 (超时: %v)\n", nodeID, elapsed)
			}

		case NodeStatusSuspected:
			if elapsed > fd.failureTimeout*2 {
				node.Status = NodeStatusFailed
				node.FailureCount++
				fmt.Printf("  [FailureDetector] 节点 %s 确认故障\n", nodeID)

				if fd.onFailure != nil {
					go fd.onFailure(nodeID)
				}
			}
		}
	}
}

// GetNodeStatus 获取节点状态
func (fd *FailureDetector) GetNodeStatus(nodeID string) (NodeStatus, bool) {
	fd.mu.RLock()
	defer fd.mu.RUnlock()

	if node, exists := fd.nodes[nodeID]; exists {
		return node.Status, true
	}
	return NodeStatusFailed, false
}

// SetCallbacks 设置回调函数
func (fd *FailureDetector) SetCallbacks(onFailure, onRecovery func(nodeID string)) {
	fd.onFailure = onFailure
	fd.onRecovery = onRecovery
}

// ============================================================================
// 主从切换（Failover）
// ============================================================================

// FailoverManager 故障转移管理器
type FailoverManager struct {
	primary    string
	replicas   []string
	status     map[string]bool
	mu         sync.RWMutex
	onSwitch   func(oldPrimary, newPrimary string)
	autoSwitch bool
}

// NewFailoverManager 创建故障转移管理器
func NewFailoverManager(primary string, replicas []string) *FailoverManager {
	status := make(map[string]bool)
	status[primary] = true
	for _, r := range replicas {
		status[r] = true
	}

	return &FailoverManager{
		primary:    primary,
		replicas:   replicas,
		status:     status,
		autoSwitch: true,
	}
}

// SetSwitchCallback 设置切换回调
func (fm *FailoverManager) SetSwitchCallback(callback func(oldPrimary, newPrimary string)) {
	fm.onSwitch = callback
}

// MarkNodeDown 标记节点下线
func (fm *FailoverManager) MarkNodeDown(nodeID string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.status[nodeID] = false
	fmt.Printf("  [Failover] 节点 %s 标记为下线\n", nodeID)

	// 如果是主节点下线，触发故障转移
	if nodeID == fm.primary && fm.autoSwitch {
		return fm.doFailover()
	}

	return nil
}

// MarkNodeUp 标记节点上线
func (fm *FailoverManager) MarkNodeUp(nodeID string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.status[nodeID] = true
	fmt.Printf("  [Failover] 节点 %s 标记为上线\n", nodeID)
}

// doFailover 执行故障转移
func (fm *FailoverManager) doFailover() error {
	oldPrimary := fm.primary

	// 选择新的主节点
	for _, replica := range fm.replicas {
		if fm.status[replica] {
			fm.primary = replica

			// 从副本列表中移除新主节点
			newReplicas := make([]string, 0)
			for _, r := range fm.replicas {
				if r != replica {
					newReplicas = append(newReplicas, r)
				}
			}
			fm.replicas = newReplicas

			fmt.Printf("  [Failover] 主节点切换: %s -> %s\n", oldPrimary, fm.primary)

			if fm.onSwitch != nil {
				go fm.onSwitch(oldPrimary, fm.primary)
			}

			return nil
		}
	}

	return fmt.Errorf("没有可用的副本节点进行故障转移")
}

// GetPrimary 获取当前主节点
func (fm *FailoverManager) GetPrimary() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.primary
}

// GetReplicas 获取副本列表
func (fm *FailoverManager) GetReplicas() []string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return append([]string{}, fm.replicas...)
}

// GetStatus 获取所有节点状态
func (fm *FailoverManager) GetStatus() map[string]bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	result := make(map[string]bool)
	for k, v := range fm.status {
		result[k] = v
	}
	return result
}

// ============================================================================
// 多活架构
// ============================================================================

// MultiActiveCluster 多活集群
type MultiActiveCluster struct {
	regions     map[string]*Region
	routingRule RoutingRule
	syncManager *SyncManager
	mu          sync.RWMutex
}

// Region 区域
type Region struct {
	ID       string
	Name     string
	Endpoint string
	Weight   int
	IsActive bool
	Latency  time.Duration
	Capacity int
	Load     int
}

// RoutingRule 路由规则
type RoutingRule int

const (
	RoutingRuleLatency    RoutingRule = iota // 最低延迟
	RoutingRuleWeight                        // 权重
	RoutingRuleGeo                           // 地理位置
	RoutingRuleRoundRobin                    // 轮询
)

// SyncManager 同步管理器
type SyncManager struct {
	conflictResolver ConflictResolver
	syncInterval     time.Duration
	lastSync         time.Time
}

// ConflictResolver 冲突解决策略
type ConflictResolver int

const (
	ConflictResolverLastWrite ConflictResolver = iota // 最后写入胜出
	ConflictResolverMerge                             // 合并
	ConflictResolverCustom                            // 自定义
)

// NewMultiActiveCluster 创建多活集群
func NewMultiActiveCluster() *MultiActiveCluster {
	return &MultiActiveCluster{
		regions:     make(map[string]*Region),
		routingRule: RoutingRuleLatency,
		syncManager: &SyncManager{
			conflictResolver: ConflictResolverLastWrite,
			syncInterval:     time.Second,
		},
	}
}

// AddRegion 添加区域
func (c *MultiActiveCluster) AddRegion(region *Region) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.regions[region.ID] = region
}

// SetRoutingRule 设置路由规则
func (c *MultiActiveCluster) SetRoutingRule(rule RoutingRule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.routingRule = rule
}

// Route 路由请求到最佳区域
func (c *MultiActiveCluster) Route(userRegion string) (*Region, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	activeRegions := make([]*Region, 0)
	for _, region := range c.regions {
		if region.IsActive && region.Load < region.Capacity {
			activeRegions = append(activeRegions, region)
		}
	}

	if len(activeRegions) == 0 {
		return nil, fmt.Errorf("没有可用的活跃区域")
	}

	switch c.routingRule {
	case RoutingRuleLatency:
		return c.routeByLatency(activeRegions), nil
	case RoutingRuleWeight:
		return c.routeByWeight(activeRegions), nil
	case RoutingRuleGeo:
		return c.routeByGeo(activeRegions, userRegion), nil
	default:
		return activeRegions[0], nil
	}
}

func (c *MultiActiveCluster) routeByLatency(regions []*Region) *Region {
	var best *Region
	for _, r := range regions {
		if best == nil || r.Latency < best.Latency {
			best = r
		}
	}
	return best
}

func (c *MultiActiveCluster) routeByWeight(regions []*Region) *Region {
	totalWeight := 0
	for _, r := range regions {
		totalWeight += r.Weight
	}

	if totalWeight == 0 {
		return regions[0]
	}

	random := rand.Intn(totalWeight)
	for _, r := range regions {
		random -= r.Weight
		if random < 0 {
			return r
		}
	}
	return regions[0]
}

func (c *MultiActiveCluster) routeByGeo(regions []*Region, userRegion string) *Region {
	// 优先选择同区域
	for _, r := range regions {
		if r.ID == userRegion {
			return r
		}
	}
	// 否则选择延迟最低的
	return c.routeByLatency(regions)
}

// SimulateRequest 模拟请求处理
func (c *MultiActiveCluster) SimulateRequest(userRegion string) {
	region, err := c.Route(userRegion)
	if err != nil {
		fmt.Printf("  路由失败: %v\n", err)
		return
	}

	c.mu.Lock()
	region.Load++
	c.mu.Unlock()

	fmt.Printf("  请求路由到区域: %s (延迟: %v, 负载: %d/%d)\n",
		region.Name, region.Latency, region.Load, region.Capacity)
}

// GetClusterStatus 获取集群状态
func (c *MultiActiveCluster) GetClusterStatus() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fmt.Println("\n  多活集群状态:")
	for _, region := range c.regions {
		status := "活跃"
		if !region.IsActive {
			status = "非活跃"
		}
		fmt.Printf("    %s (%s): %s, 延迟=%v, 负载=%d/%d, 权重=%d\n",
			region.ID, region.Name, status, region.Latency, region.Load, region.Capacity, region.Weight)
	}
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateHealthCheck() {
	fmt.Println("\n=== 健康检查机制演示 ===")
	fmt.Println("场景: 微服务健康状态监控")

	checker := NewCompositeHealthChecker(5 * time.Second)

	// 添加各种健康检查
	checker.AddCheck(NewDatabaseHealthCheck("primary-db"))
	checker.AddCheck(NewDatabaseHealthCheck("replica-db"))
	checker.AddCheck(NewCacheHealthCheck("redis-cache"))

	ctx := context.Background()

	fmt.Println("\n执行健康检查:")
	results := checker.RunAll(ctx)

	for name, result := range results {
		fmt.Printf("  %s: %s (延迟: %v)\n", name, result.Status, result.Latency)
		fmt.Printf("    消息: %s\n", result.Message)
		if len(result.Details) > 0 {
			fmt.Printf("    详情: %v\n", result.Details)
		}
	}

	fmt.Printf("\n整体健康状态: %s\n", checker.GetOverallStatus())
}

func demonstrateFailureDetection() {
	fmt.Println("\n=== 故障检测与自动恢复演示 ===")
	fmt.Println("场景: 分布式系统节点故障检测")

	detector := NewFailureDetector(
		500*time.Millisecond, // 心跳周期
		1*time.Second,        // 故障超时
	)

	// 设置回调
	detector.SetCallbacks(
		func(nodeID string) {
			fmt.Printf("  [回调] 节点 %s 故障处理已触发\n", nodeID)
		},
		func(nodeID string) {
			fmt.Printf("  [回调] 节点 %s 恢复处理已触发\n", nodeID)
		},
	)

	// 注册节点
	nodes := []string{"node-1", "node-2", "node-3"}
	for _, nodeID := range nodes {
		detector.RegisterNode(nodeID, map[string]string{
			"role": "worker",
			"zone": "us-west-1",
		})
		fmt.Printf("  注册节点: %s\n", nodeID)
	}

	// 模拟心跳
	fmt.Println("\n模拟正常心跳:")
	for i := 0; i < 3; i++ {
		for _, nodeID := range nodes {
			detector.Heartbeat(nodeID)
		}
		detector.CheckNodes()
		time.Sleep(300 * time.Millisecond)
	}

	// 模拟 node-2 故障
	fmt.Println("\n模拟 node-2 停止心跳:")
	for i := 0; i < 5; i++ {
		detector.Heartbeat("node-1")
		detector.Heartbeat("node-3")
		// node-2 不发送心跳
		detector.CheckNodes()
		time.Sleep(500 * time.Millisecond)
	}

	// 模拟 node-2 恢复
	fmt.Println("\n模拟 node-2 恢复:")
	detector.Heartbeat("node-2")
	detector.CheckNodes()

	// 显示最终状态
	fmt.Println("\n最终节点状态:")
	for _, nodeID := range nodes {
		status, _ := detector.GetNodeStatus(nodeID)
		fmt.Printf("  %s: %s\n", nodeID, status)
	}
}

func demonstrateFailover() {
	fmt.Println("\n=== 主从切换演示 ===")
	fmt.Println("场景: 数据库主从故障转移")

	fm := NewFailoverManager(
		"db-primary",
		[]string{"db-replica-1", "db-replica-2"},
	)

	fm.SetSwitchCallback(func(oldPrimary, newPrimary string) {
		fmt.Printf("  [回调] 执行切换后处理: %s -> %s\n", oldPrimary, newPrimary)
		fmt.Println("    - 更新连接池配置")
		fmt.Println("    - 通知应用层")
		fmt.Println("    - 记录审计日志")
	})

	fmt.Printf("\n初始状态:\n")
	fmt.Printf("  主节点: %s\n", fm.GetPrimary())
	fmt.Printf("  副本: %v\n", fm.GetReplicas())
	fmt.Printf("  状态: %v\n", fm.GetStatus())

	// 模拟主节点故障
	fmt.Println("\n模拟主节点故障:")
	if err := fm.MarkNodeDown("db-primary"); err != nil {
		fmt.Printf("  故障转移失败: %v\n", err)
	}

	time.Sleep(100 * time.Millisecond) // 等待回调执行

	fmt.Printf("\n故障转移后状态:\n")
	fmt.Printf("  主节点: %s\n", fm.GetPrimary())
	fmt.Printf("  副本: %v\n", fm.GetReplicas())
	fmt.Printf("  状态: %v\n", fm.GetStatus())

	// 模拟原主节点恢复
	fmt.Println("\n模拟原主节点恢复 (作为副本加入):")
	fm.MarkNodeUp("db-primary")
}

func demonstrateMultiActive() {
	fmt.Println("\n=== 多活架构演示 ===")
	fmt.Println("场景: 全球多活数据中心")

	cluster := NewMultiActiveCluster()

	// 添加区域
	cluster.AddRegion(&Region{
		ID:       "us-west",
		Name:     "美西",
		Endpoint: "us-west.example.com",
		Weight:   30,
		IsActive: true,
		Latency:  50 * time.Millisecond,
		Capacity: 1000,
		Load:     0,
	})

	cluster.AddRegion(&Region{
		ID:       "us-east",
		Name:     "美东",
		Endpoint: "us-east.example.com",
		Weight:   30,
		IsActive: true,
		Latency:  80 * time.Millisecond,
		Capacity: 1000,
		Load:     0,
	})

	cluster.AddRegion(&Region{
		ID:       "eu-west",
		Name:     "欧洲",
		Endpoint: "eu-west.example.com",
		Weight:   20,
		IsActive: true,
		Latency:  120 * time.Millisecond,
		Capacity: 800,
		Load:     0,
	})

	cluster.AddRegion(&Region{
		ID:       "ap-east",
		Name:     "亚太",
		Endpoint: "ap-east.example.com",
		Weight:   20,
		IsActive: true,
		Latency:  150 * time.Millisecond,
		Capacity: 600,
		Load:     0,
	})

	cluster.GetClusterStatus()

	// 测试不同路由策略
	fmt.Println("\n测试延迟优先路由:")
	cluster.SetRoutingRule(RoutingRuleLatency)
	for i := 0; i < 3; i++ {
		cluster.SimulateRequest("unknown")
	}

	fmt.Println("\n测试权重路由:")
	cluster.SetRoutingRule(RoutingRuleWeight)
	for i := 0; i < 5; i++ {
		cluster.SimulateRequest("unknown")
	}

	fmt.Println("\n测试地理位置路由:")
	cluster.SetRoutingRule(RoutingRuleGeo)
	cluster.SimulateRequest("us-west")
	cluster.SimulateRequest("eu-west")
	cluster.SimulateRequest("ap-east")

	cluster.GetClusterStatus()
}

func main() {
	fmt.Println("=== 高可用架构设计 ===")
	fmt.Println()
	fmt.Println("本模块演示高可用系统的核心组件:")
	fmt.Println("1. 健康检查机制 - 服务状态监控")
	fmt.Println("2. 故障检测与自动恢复 - 节点故障发现")
	fmt.Println("3. 主从切换 - 故障转移机制")
	fmt.Println("4. 多活架构 - 全球分布式部署")

	demonstrateHealthCheck()
	demonstrateFailureDetection()
	demonstrateFailover()
	demonstrateMultiActive()

	fmt.Println("\n=== 高可用架构演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 健康检查: 多维度监控，组合检查，状态聚合")
	fmt.Println("- 故障检测: 心跳机制，超时判定，状态机转换")
	fmt.Println("- 主从切换: 自动故障转移，回调通知，状态同步")
	fmt.Println("- 多活架构: 多区域部署，智能路由，负载均衡")
}
