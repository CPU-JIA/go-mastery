/*
=== Go性能掌控：分布式性能分析 ===

本模块专注于分布式系统中的性能分析技术，探索：
1. 跨服务性能追踪和链路分析
2. 微服务架构性能瓶颈识别
3. 分布式系统延迟分析和优化
4. 网络性能监控和调优
5. 数据库连接池和查询优化
6. 负载均衡算法和性能影响
7. 容器化环境性能监控
8. 云原生架构性能调优
9. 分布式缓存性能优化
10. 服务网格性能分析

学习目标：
- 掌握分布式系统性能分析方法
- 理解微服务架构性能挑战
- 学会分布式链路追踪技术
- 掌握大规模系统性能优化策略
*/

package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"net"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// 1. 分布式追踪系统
// ==================

// DistributedTracer 分布式追踪器
type DistributedTracer struct {
	services     map[string]*ServiceNode
	traces       map[string]*DistributedTrace
	spans        map[string]*DistributedSpan
	dependencies map[string][]string
	metrics      TracingMetrics
	sampler      TraceSampler
	exporter     TraceExporter
	mutex        sync.RWMutex
	running      bool
	stopCh       chan struct{}
}

// ServiceNode 服务节点
type ServiceNode struct {
	Name         string
	Version      string
	Instances    []*ServiceInstance
	Dependencies []string
	Metrics      ServiceMetrics
	Health       HealthStatus
}

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID       string
	Address  string
	Port     int
	Region   string
	Zone     string
	Metadata map[string]string
	Metrics  InstanceMetrics
	Status   InstanceStatus
}

// DistributedTrace 分布式追踪
type DistributedTrace struct {
	TraceID        string
	RootSpan       *DistributedSpan
	Spans          []*DistributedSpan
	Services       map[string]*ServiceInvolvement
	StartTime      time.Time
	EndTime        time.Time
	Duration       time.Duration
	Status         TraceStatus
	ErrorRate      float64
	CriticalPath   []*DistributedSpan
	BottleneckSpan *DistributedSpan
}

// DistributedSpan 分布式Span
type DistributedSpan struct {
	SpanID        string
	TraceID       string
	ParentSpanID  string
	ServiceName   string
	OperationName string
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Tags          map[string]interface{}
	Logs          []SpanLog
	Events        []SpanEvent
	Links         []SpanLink
	Status        SpanStatus
	Error         error
	Resources     ResourceUsage
	Network       NetworkMetrics
}

// ServiceInvolvement 服务参与度
type ServiceInvolvement struct {
	ServiceName string
	SpanCount   int
	TotalTime   time.Duration
	ErrorCount  int
	CPUTime     time.Duration
	MemoryUsage int64
}

// SpanEvent Span事件
type SpanEvent struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]interface{}
}

// SpanLink Span链接
type SpanLink struct {
	TraceID string
	SpanID  string
	Type    LinkType
}

// LinkType 链接类型
type LinkType int

const (
	ChildOf LinkType = iota
	FollowsFrom
	CausedBy
)

// ServiceMetrics 服务指标
type ServiceMetrics struct {
	RequestCount    int64
	ErrorCount      int64
	AvgResponseTime time.Duration
	Throughput      float64
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string
	Message   string
	Timestamp time.Time
}

// InstanceStatus 实例状态
type InstanceStatus struct {
	State     string
	Health    string
	LastCheck time.Time
}

// TraceStatus 追踪状态
type TraceStatus struct {
	Code    int
	Message string
	Success bool
}

// SpanLog Span日志
type SpanLog struct {
	Timestamp time.Time
	Message   string
	Level     string
	Fields    map[string]interface{}
}

// SpanStatus Span状态
type SpanStatus struct {
	Code        int
	Message     string
	IsError     bool
	Description string
}

// Span状态常量
var (
	SpanStatusOK = SpanStatus{
		Code:        0,
		Message:     "OK",
		IsError:     false,
		Description: "Success",
	}
	SpanStatusError = SpanStatus{
		Code:        1,
		Message:     "Error",
		IsError:     true,
		Description: "Operation failed",
	}
)

// Trace状态常量
var (
	TraceStatusOK = TraceStatus{
		Code:    0,
		Message: "OK",
		Success: true,
	}
	TraceStatusError = TraceStatus{
		Code:    1,
		Message: "Error",
		Success: false,
	}
)

// Instance状态常量
var (
	InstanceStatusHealthy = InstanceStatus{
		State:     "running",
		Health:    "healthy",
		LastCheck: time.Now(),
	}
)

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	ID        string
	Service   string
	Type      string
	Message   string
	Severity  string
	Value     float64
	Threshold float64
	Timestamp time.Time
}

// 告警严重程度常量
const (
	AlertInfo     = "info"
	AlertWarning  = "warning"
	AlertCritical = "critical"
)

// EndpointMetrics 端点指标
type EndpointMetrics struct {
	Path          string
	Method        string
	RequestCount  int64
	ErrorCount    int64
	ResponseTime  time.Duration
	ThroughputRPS float64
}

// ConnectionPool 连接池
type ConnectionPool struct {
	pool              chan net.Conn
	maxConn           int
	factory           func() (net.Conn, error)
	MaxConnections    int
	ActiveConnections int
	IdleConnections   int
	ConnectionTimeout time.Duration
	IdleTimeout       time.Duration
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUUsage    float64
	MemoryUsage int64
	DiskIO      int64
	NetworkIO   int64
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	Latency    time.Duration
	Bandwidth  int64
	PacketLoss float64
	Retries    int
	Timeouts   int
}

// TracingMetrics 追踪指标
type TracingMetrics struct {
	TracesGenerated int64
	SpansGenerated  int64
	TracesDropped   int64
	SpansDropped    int64
	AvgTraceSize    float64
	AvgSpanCount    float64
	SamplingRate    float64
}

// TraceSampler 追踪采样器
type TraceSampler interface {
	ShouldSample(traceID string, operationName string) bool
	GetSamplingRate() float64
}

// TraceExporter 追踪导出器
type TraceExporter interface {
	ExportSpans(spans []*DistributedSpan) error
	ExportTraces(traces []*DistributedTrace) error
}

func NewDistributedTracer() *DistributedTracer {
	return &DistributedTracer{
		services:     make(map[string]*ServiceNode),
		traces:       make(map[string]*DistributedTrace),
		spans:        make(map[string]*DistributedSpan),
		dependencies: make(map[string][]string),
		sampler:      &ProbabilisticSampler{rate: 0.1}, // 10%采样率
		exporter:     &ConsoleExporter{},
		stopCh:       make(chan struct{}),
	}
}

func (dt *DistributedTracer) RegisterService(service *ServiceNode) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	dt.services[service.Name] = service
	fmt.Printf("注册服务: %s (版本: %s)\n", service.Name, service.Version)
}

func (dt *DistributedTracer) StartTrace(serviceName, operationName string, parentTraceID, parentSpanID string) *DistributedTrace {
	traceID := dt.generateTraceID()
	if parentTraceID != "" {
		traceID = parentTraceID
	}

	rootSpan := &DistributedSpan{
		SpanID:        dt.generateSpanID(),
		TraceID:       traceID,
		ParentSpanID:  parentSpanID,
		ServiceName:   serviceName,
		OperationName: operationName,
		StartTime:     time.Now(),
		Tags:          make(map[string]interface{}),
		Logs:          make([]SpanLog, 0),
		Events:        make([]SpanEvent, 0),
		Links:         make([]SpanLink, 0),
		Status:        SpanStatusOK,
	}

	trace := &DistributedTrace{
		TraceID:   traceID,
		RootSpan:  rootSpan,
		Spans:     []*DistributedSpan{rootSpan},
		Services:  make(map[string]*ServiceInvolvement),
		StartTime: time.Now(),
		Status:    TraceStatusOK,
	}

	dt.mutex.Lock()
	dt.traces[traceID] = trace
	dt.spans[rootSpan.SpanID] = rootSpan
	atomic.AddInt64(&dt.metrics.TracesGenerated, 1)
	atomic.AddInt64(&dt.metrics.SpansGenerated, 1)
	dt.mutex.Unlock()

	return trace
}

func (dt *DistributedTracer) StartSpan(traceID, parentSpanID, serviceName, operationName string) *DistributedSpan {
	span := &DistributedSpan{
		SpanID:        dt.generateSpanID(),
		TraceID:       traceID,
		ParentSpanID:  parentSpanID,
		ServiceName:   serviceName,
		OperationName: operationName,
		StartTime:     time.Now(),
		Tags:          make(map[string]interface{}),
		Logs:          make([]SpanLog, 0),
		Events:        make([]SpanEvent, 0),
		Links:         make([]SpanLink, 0),
		Status:        SpanStatusOK,
	}

	dt.mutex.Lock()
	dt.spans[span.SpanID] = span

	if trace, exists := dt.traces[traceID]; exists {
		trace.Spans = append(trace.Spans, span)

		// 更新服务参与度
		if involvement, exists := trace.Services[serviceName]; exists {
			involvement.SpanCount++
		} else {
			trace.Services[serviceName] = &ServiceInvolvement{
				ServiceName: serviceName,
				SpanCount:   1,
			}
		}
	}

	atomic.AddInt64(&dt.metrics.SpansGenerated, 1)
	dt.mutex.Unlock()

	return span
}

func (dt *DistributedTracer) FinishSpan(spanID string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	span, exists := dt.spans[spanID]
	if !exists {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	// 更新追踪
	if trace, exists := dt.traces[span.TraceID]; exists {
		if involvement, exists := trace.Services[span.ServiceName]; exists {
			involvement.TotalTime += span.Duration
			if span.Status == SpanStatusError {
				involvement.ErrorCount++
			}
		}

		// 如果是根span，更新trace
		if trace.RootSpan.SpanID == spanID {
			trace.EndTime = span.EndTime
			trace.Duration = span.Duration
			dt.analyzeTrace(trace)
		}
	}
}

func (dt *DistributedTracer) analyzeTrace(trace *DistributedTrace) {
	// 分析关键路径
	trace.CriticalPath = dt.findCriticalPath(trace)

	// 找出瓶颈span
	trace.BottleneckSpan = dt.findBottleneckSpan(trace)

	// 计算错误率
	errorCount := 0
	for _, span := range trace.Spans {
		if span.Status == SpanStatusError {
			errorCount++
		}
	}
	if len(trace.Spans) > 0 {
		trace.ErrorRate = float64(errorCount) / float64(len(trace.Spans)) * 100
	}
}

func (dt *DistributedTracer) findCriticalPath(trace *DistributedTrace) []*DistributedSpan {
	// 构建span依赖图
	spanMap := make(map[string]*DistributedSpan)
	children := make(map[string][]*DistributedSpan)

	for _, span := range trace.Spans {
		spanMap[span.SpanID] = span
		if span.ParentSpanID != "" {
			children[span.ParentSpanID] = append(children[span.ParentSpanID], span)
		}
	}

	// 递归计算最长路径
	var findLongestPath func(spanID string) ([]*DistributedSpan, time.Duration)
	findLongestPath = func(spanID string) ([]*DistributedSpan, time.Duration) {
		span := spanMap[spanID]
		if span == nil {
			return nil, 0
		}

		maxPath := []*DistributedSpan{span}
		maxDuration := span.Duration

		for _, child := range children[spanID] {
			childPath, childDuration := findLongestPath(child.SpanID)
			if childDuration > maxDuration {
				maxPath = append([]*DistributedSpan{span}, childPath...)
				maxDuration = childDuration
			}
		}

		return maxPath, maxDuration
	}

	if trace.RootSpan != nil {
		path, _ := findLongestPath(trace.RootSpan.SpanID)
		return path
	}

	return nil
}

func (dt *DistributedTracer) findBottleneckSpan(trace *DistributedTrace) *DistributedSpan {
	var bottleneck *DistributedSpan
	maxDuration := time.Duration(0)

	for _, span := range trace.Spans {
		if span.Duration > maxDuration {
			maxDuration = span.Duration
			bottleneck = span
		}
	}

	return bottleneck
}

func (dt *DistributedTracer) generateTraceID() string {
	return fmt.Sprintf("trace_%d_%d", time.Now().UnixNano(), rand.Int63())
}

func (dt *DistributedTracer) generateSpanID() string {
	return fmt.Sprintf("span_%d_%d", time.Now().UnixNano(), rand.Int63())
}

// ==================
// 2. 微服务性能分析器
// ==================

// MicroserviceAnalyzer 微服务性能分析器
type MicroserviceAnalyzer struct {
	services        map[string]*MicroserviceMetrics
	dependencies    *DependencyGraph
	sla             map[string]*SLA
	alerts          []PerformanceAlert
	recommendations []OptimizationRecommendation
	mutex           sync.RWMutex
}

// MicroserviceMetrics 微服务指标
type MicroserviceMetrics struct {
	ServiceName     string
	RequestRate     float64
	ErrorRate       float64
	ResponseTime    ResponseTimeMetrics
	Throughput      float64
	Availability    float64
	ResourceUsage   ResourceMetrics
	Dependencies    []DependencyMetrics
	Endpoints       map[string]*EndpointMetrics
	InstanceMetrics map[string]*InstanceMetrics
}

// ResponseTimeMetrics 响应时间指标
type ResponseTimeMetrics struct {
	P50  time.Duration
	P90  time.Duration
	P95  time.Duration
	P99  time.Duration
	P999 time.Duration
	Mean time.Duration
	Max  time.Duration
	Min  time.Duration
}

// ResourceMetrics 资源指标
type ResourceMetrics struct {
	CPUUsage       float64
	MemoryUsage    float64
	DiskUsage      float64
	NetworkIO      float64
	FileHandles    int
	Connections    int
	ThreadCount    int
	GoroutineCount int
}

// DependencyMetrics 依赖指标
type DependencyMetrics struct {
	ServiceName  string
	CallRate     float64
	ErrorRate    float64
	ResponseTime time.Duration
	CircuitState CircuitBreakerState
	RetryCount   int
	TimeoutCount int
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// InstanceMetrics 实例指标
type InstanceMetrics struct {
	InstanceID   string
	RequestRate  float64
	ErrorRate    float64
	ResponseTime time.Duration
	CPUUsage     float64
	MemoryUsage  float64
	Health       HealthStatus
	LoadBalance  LoadBalanceMetrics
}

// LoadBalanceMetrics 负载均衡指标
type LoadBalanceMetrics struct {
	Weight          int
	ActiveRequests  int
	TotalRequests   int64
	FailureCount    int64
	ResponseTime    time.Duration
	LastHealthCheck time.Time
}

// DependencyGraph 依赖图
type DependencyGraph struct {
	nodes map[string]*ServiceNode
	edges map[string][]*DependencyEdge
	mutex sync.RWMutex
}

// DependencyEdge 依赖边
type DependencyEdge struct {
	From         string
	To           string
	CallRate     float64
	ErrorRate    float64
	ResponseTime time.Duration
	Protocol     string
	Weight       float64
}

// SLA 服务等级协议
type SLA struct {
	ServiceName     string
	ResponseTimeP95 time.Duration
	ResponseTimeP99 time.Duration
	Availability    float64
	ErrorRate       float64
	Throughput      float64
	TimeWindow      time.Duration
}

// OptimizationRecommendation 优化建议
type OptimizationRecommendation struct {
	Type        RecommendationType
	Service     string
	Priority    Priority
	Description string
	Impact      string
	Effort      string
	Details     map[string]interface{}
}

// RecommendationType 建议类型
type RecommendationType int

const (
	ScaleUp RecommendationType = iota
	ScaleDown
	OptimizeQuery
	AddCache
	OptimizeAlgorithm
	UpgradeHardware
	RebalanceLoad
	AddCircuitBreaker
)

// Priority 优先级
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

func NewMicroserviceAnalyzer() *MicroserviceAnalyzer {
	return &MicroserviceAnalyzer{
		services:        make(map[string]*MicroserviceMetrics),
		dependencies:    NewDependencyGraph(),
		sla:             make(map[string]*SLA),
		alerts:          make([]PerformanceAlert, 0),
		recommendations: make([]OptimizationRecommendation, 0),
	}
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*ServiceNode),
		edges: make(map[string][]*DependencyEdge),
	}
}

func (ma *MicroserviceAnalyzer) RegisterService(serviceName string, sla *SLA) {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	ma.services[serviceName] = &MicroserviceMetrics{
		ServiceName:     serviceName,
		Endpoints:       make(map[string]*EndpointMetrics),
		InstanceMetrics: make(map[string]*InstanceMetrics),
		Dependencies:    make([]DependencyMetrics, 0),
	}

	if sla != nil {
		ma.sla[serviceName] = sla
	}

	fmt.Printf("注册微服务: %s\n", serviceName)
}

func (ma *MicroserviceAnalyzer) UpdateMetrics(serviceName string, metrics *MicroserviceMetrics) {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	if existing, exists := ma.services[serviceName]; exists {
		// 更新指标
		existing.RequestRate = metrics.RequestRate
		existing.ErrorRate = metrics.ErrorRate
		existing.ResponseTime = metrics.ResponseTime
		existing.Throughput = metrics.Throughput
		existing.Availability = metrics.Availability
		existing.ResourceUsage = metrics.ResourceUsage

		// 检查SLA违反
		ma.checkSLAViolations(serviceName, existing)

		// 生成优化建议
		ma.generateRecommendations(serviceName, existing)
	}
}

func (ma *MicroserviceAnalyzer) checkSLAViolations(serviceName string, metrics *MicroserviceMetrics) {
	sla, exists := ma.sla[serviceName]
	if !exists {
		return
	}

	// 检查响应时间
	if metrics.ResponseTime.P95 > sla.ResponseTimeP95 {
		alert := PerformanceAlert{
			ID:        fmt.Sprintf("sla_violation_%s_%d", serviceName, time.Now().Unix()),
			Type:      "SLA_VIOLATION",
			Message:   fmt.Sprintf("服务 %s P95响应时间超过SLA: %v > %v", serviceName, metrics.ResponseTime.P95, sla.ResponseTimeP95),
			Severity:  AlertWarning,
			Timestamp: time.Now(),
			Value:     float64(metrics.ResponseTime.P95.Milliseconds()),
			Threshold: float64(sla.ResponseTimeP95.Milliseconds()),
		}
		ma.alerts = append(ma.alerts, alert)
	}

	// 检查可用性
	if metrics.Availability < sla.Availability {
		alert := PerformanceAlert{
			ID:        fmt.Sprintf("availability_violation_%s_%d", serviceName, time.Now().Unix()),
			Type:      "AVAILABILITY_VIOLATION",
			Message:   fmt.Sprintf("服务 %s 可用性低于SLA: %.2f%% < %.2f%%", serviceName, metrics.Availability, sla.Availability),
			Severity:  AlertCritical,
			Timestamp: time.Now(),
			Value:     metrics.Availability,
			Threshold: sla.Availability,
		}
		ma.alerts = append(ma.alerts, alert)
	}

	// 检查错误率
	if metrics.ErrorRate > sla.ErrorRate {
		alert := PerformanceAlert{
			ID:        fmt.Sprintf("error_rate_violation_%s_%d", serviceName, time.Now().Unix()),
			Type:      "ERROR_RATE_VIOLATION",
			Message:   fmt.Sprintf("服务 %s 错误率超过SLA: %.2f%% > %.2f%%", serviceName, metrics.ErrorRate, sla.ErrorRate),
			Severity:  AlertCritical,
			Timestamp: time.Now(),
			Value:     metrics.ErrorRate,
			Threshold: sla.ErrorRate,
		}
		ma.alerts = append(ma.alerts, alert)
	}
}

func (ma *MicroserviceAnalyzer) generateRecommendations(serviceName string, metrics *MicroserviceMetrics) {
	// CPU使用率高的情况
	if metrics.ResourceUsage.CPUUsage > 80 {
		recommendation := OptimizationRecommendation{
			Type:        ScaleUp,
			Service:     serviceName,
			Priority:    PriorityHigh,
			Description: "CPU使用率过高，建议扩容实例",
			Impact:      "改善响应时间和吞吐量",
			Effort:      "低",
			Details: map[string]interface{}{
				"current_cpu":      metrics.ResourceUsage.CPUUsage,
				"threshold":        80.0,
				"suggested_action": "增加实例数量或升级CPU规格",
			},
		}
		ma.recommendations = append(ma.recommendations, recommendation)
	}

	// 内存使用率高的情况
	if metrics.ResourceUsage.MemoryUsage > 85 {
		recommendation := OptimizationRecommendation{
			Type:        UpgradeHardware,
			Service:     serviceName,
			Priority:    PriorityMedium,
			Description: "内存使用率过高，建议优化内存使用或扩容",
			Impact:      "避免OOM和性能下降",
			Effort:      "中",
			Details: map[string]interface{}{
				"current_memory":   metrics.ResourceUsage.MemoryUsage,
				"threshold":        85.0,
				"suggested_action": "优化内存使用或增加内存容量",
			},
		}
		ma.recommendations = append(ma.recommendations, recommendation)
	}

	// 响应时间高的情况
	if metrics.ResponseTime.P95 > 500*time.Millisecond {
		recommendation := OptimizationRecommendation{
			Type:        AddCache,
			Service:     serviceName,
			Priority:    PriorityHigh,
			Description: "响应时间过长，建议添加缓存或优化查询",
			Impact:      "显著改善用户体验",
			Effort:      "中",
			Details: map[string]interface{}{
				"current_p95":      metrics.ResponseTime.P95.Milliseconds(),
				"threshold":        500,
				"suggested_action": "添加Redis缓存或优化数据库查询",
			},
		}
		ma.recommendations = append(ma.recommendations, recommendation)
	}

	// 错误率高的情况
	if metrics.ErrorRate > 1.0 {
		recommendation := OptimizationRecommendation{
			Type:        AddCircuitBreaker,
			Service:     serviceName,
			Priority:    PriorityCritical,
			Description: "错误率过高，建议添加熔断器和重试机制",
			Impact:      "提高系统稳定性",
			Effort:      "中",
			Details: map[string]interface{}{
				"current_error_rate": metrics.ErrorRate,
				"threshold":          1.0,
				"suggested_action":   "实现熔断器模式和指数退避重试",
			},
		}
		ma.recommendations = append(ma.recommendations, recommendation)
	}
}

func (ma *MicroserviceAnalyzer) AnalyzeDependencies() {
	ma.dependencies.mutex.RLock()
	defer ma.dependencies.mutex.RUnlock()

	fmt.Println("=== 依赖关系分析 ===")

	// 分析服务依赖深度
	depths := ma.calculateServiceDepths()
	fmt.Println("服务依赖深度:")
	for service, depth := range depths {
		fmt.Printf("  %s: %d层\n", service, depth)
	}

	// 识别关键路径
	criticalServices := ma.findCriticalServices()
	fmt.Println("\n关键服务 (影响多个下游服务):")
	for _, service := range criticalServices {
		fmt.Printf("  %s\n", service)
	}

	// 分析潜在的单点故障
	spof := ma.findSinglePointsOfFailure()
	if len(spof) > 0 {
		fmt.Println("\n潜在单点故障:")
		for _, service := range spof {
			fmt.Printf("  ⚠️  %s\n", service)
		}
	}
}

func (ma *MicroserviceAnalyzer) calculateServiceDepths() map[string]int {
	depths := make(map[string]int)
	visited := make(map[string]bool)

	var dfs func(service string, depth int) int
	dfs = func(service string, depth int) int {
		if visited[service] {
			return depths[service]
		}

		visited[service] = true
		maxDepth := depth

		for _, edge := range ma.dependencies.edges[service] {
			childDepth := dfs(edge.To, depth+1)
			if childDepth > maxDepth {
				maxDepth = childDepth
			}
		}

		depths[service] = maxDepth
		return maxDepth
	}

	for service := range ma.dependencies.nodes {
		if !visited[service] {
			dfs(service, 0)
		}
	}

	return depths
}

func (ma *MicroserviceAnalyzer) findCriticalServices() []string {
	downstreamCount := make(map[string]int)

	// 计算每个服务的下游服务数量
	for _, edges := range ma.dependencies.edges {
		for _, edge := range edges {
			downstreamCount[edge.From]++
		}
	}

	// 找出下游服务数量较多的服务
	var critical []string
	for service, count := range downstreamCount {
		if count >= 3 { // 影响3个或更多下游服务
			critical = append(critical, service)
		}
	}

	return critical
}

func (ma *MicroserviceAnalyzer) findSinglePointsOfFailure() []string {
	// 简化版本：没有备用实例的服务
	var spof []string

	for serviceName, service := range ma.dependencies.nodes {
		if len(service.Instances) == 1 {
			spof = append(spof, serviceName)
		}
	}

	return spof
}

func (ma *MicroserviceAnalyzer) GetRecommendations() []OptimizationRecommendation {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	// 按优先级排序
	recommendations := make([]OptimizationRecommendation, len(ma.recommendations))
	copy(recommendations, ma.recommendations)

	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority > recommendations[j].Priority
	})

	return recommendations
}

// ==================
// 3. 数据库性能优化器
// ==================

// DatabaseOptimizer 数据库性能优化器
type DatabaseOptimizer struct {
	connections   map[string]*ConnectionPool
	queryAnalyzer *QueryAnalyzer
	cacheManager  *CacheManager
	metrics       DatabaseMetrics
	mutex         sync.RWMutex
}

// QueryAnalyzer 查询分析器
type QueryAnalyzer struct {
	slowQueries      []SlowQuery
	queryPatterns    map[string]*QueryPattern
	indexSuggestions []IndexSuggestion
	mutex            sync.RWMutex
}

// SlowQuery 慢查询
type SlowQuery struct {
	Query         string
	ExecutionTime time.Duration
	Timestamp     time.Time
	Database      string
	Table         string
	RowsExamined  int64
	RowsReturned  int64
	LockTime      time.Duration
}

// QueryPattern 查询模式
type QueryPattern struct {
	Pattern     string
	Count       int64
	TotalTime   time.Duration
	AverageTime time.Duration
	Tables      []string
	Type        QueryType
}

// QueryType 查询类型
type QueryType int

const (
	QuerySelect QueryType = iota
	QueryInsert
	QueryUpdate
	QueryDelete
)

// IndexSuggestion 索引建议
type IndexSuggestion struct {
	Table   string
	Columns []string
	Type    IndexType
	Benefit float64
	Cost    float64
	Reason  string
}

// IndexType 索引类型
type IndexType int

const (
	IndexBTree IndexType = iota
	IndexHash
	IndexComposite
)

// CacheManager 缓存管理器
type CacheManager struct {
	caches   map[string]*Cache
	hitRates map[string]float64
	policies map[string]CachePolicy
	mutex    sync.RWMutex
}

// Cache 缓存
type Cache struct {
	Name      string
	Size      int64
	MaxSize   int64
	HitCount  int64
	MissCount int64
	Items     map[string]*CacheItem
	Policy    CachePolicy
	mutex     sync.RWMutex
}

// CacheItem 缓存项
type CacheItem struct {
	Key         string
	Value       interface{}
	Expiry      time.Time
	AccessTime  time.Time
	AccessCount int64
	Size        int64
}

// CachePolicy 缓存策略
type CachePolicy int

const (
	LRU CachePolicy = iota
	LFU
	FIFO
	TTL
)

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	ConnectionsActive int
	ConnectionsIdle   int
	ConnectionsMax    int
	QueriesPerSecond  float64
	SlowQueries       int64
	CacheHitRate      float64
	AvgResponseTime   time.Duration
	DeadlockCount     int64
	LockWaitTime      time.Duration
}

func NewDatabaseOptimizer() *DatabaseOptimizer {
	return &DatabaseOptimizer{
		connections:   make(map[string]*ConnectionPool),
		queryAnalyzer: NewQueryAnalyzer(),
		cacheManager:  NewCacheManager(),
	}
}

func NewQueryAnalyzer() *QueryAnalyzer {
	return &QueryAnalyzer{
		slowQueries:      make([]SlowQuery, 0),
		queryPatterns:    make(map[string]*QueryPattern),
		indexSuggestions: make([]IndexSuggestion, 0),
	}
}

func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches:   make(map[string]*Cache),
		hitRates: make(map[string]float64),
		policies: make(map[string]CachePolicy),
	}
}

func (do *DatabaseOptimizer) CreateConnectionPool(name string, maxConn int, connString string) *ConnectionPool {
	pool := &ConnectionPool{
		pool:    make(chan net.Conn, maxConn),
		maxConn: maxConn,
		factory: func() (net.Conn, error) {
			// 模拟数据库连接创建
			return nil, nil
		},
	}

	do.mutex.Lock()
	do.connections[name] = pool
	do.mutex.Unlock()

	fmt.Printf("创建数据库连接池: %s (最大连接: %d)\n", name, maxConn)
	return pool
}

func (qa *QueryAnalyzer) AnalyzeQuery(query string, executionTime time.Duration, database string) {
	qa.mutex.Lock()
	defer qa.mutex.Unlock()

	// 记录慢查询
	if executionTime > 100*time.Millisecond { // 慢查询阈值
		slowQuery := SlowQuery{
			Query:         query,
			ExecutionTime: executionTime,
			Timestamp:     time.Now(),
			Database:      database,
		}
		qa.slowQueries = append(qa.slowQueries, slowQuery)

		// 保持最近1000个慢查询
		if len(qa.slowQueries) > 1000 {
			qa.slowQueries = qa.slowQueries[100:]
		}
	}

	// 分析查询模式
	pattern := qa.extractQueryPattern(query)
	if existing, exists := qa.queryPatterns[pattern]; exists {
		existing.Count++
		existing.TotalTime += executionTime
		existing.AverageTime = existing.TotalTime / time.Duration(existing.Count)
	} else {
		qa.queryPatterns[pattern] = &QueryPattern{
			Pattern:     pattern,
			Count:       1,
			TotalTime:   executionTime,
			AverageTime: executionTime,
			Type:        qa.getQueryType(query),
		}
	}

	// 生成索引建议
	qa.generateIndexSuggestions(query, executionTime)
}

func (qa *QueryAnalyzer) extractQueryPattern(query string) string {
	// 简化的查询模式提取
	query = strings.ToLower(strings.TrimSpace(query))

	// 替换具体值为占位符
	patterns := []struct {
		regex       string
		replacement string
	}{
		{`\d+`, "?"},     // 数字
		{`'[^']*'`, "?"}, // 字符串
		{`"[^"]*"`, "?"}, // 字符串
		{`\s+`, " "},     // 多个空格
	}

	for _, _ = range patterns {
		// 这里应该用正则表达式，简化处理
		if strings.Contains(query, "'") {
			parts := strings.Split(query, "'")
			for i := 1; i < len(parts); i += 2 {
				parts[i] = "?"
			}
			query = strings.Join(parts, "")
		}
	}

	return query
}

func (qa *QueryAnalyzer) getQueryType(query string) QueryType {
	query = strings.ToLower(strings.TrimSpace(query))
	if strings.HasPrefix(query, "select") {
		return QuerySelect
	} else if strings.HasPrefix(query, "insert") {
		return QueryInsert
	} else if strings.HasPrefix(query, "update") {
		return QueryUpdate
	} else if strings.HasPrefix(query, "delete") {
		return QueryDelete
	}
	return QuerySelect
}

func (qa *QueryAnalyzer) generateIndexSuggestions(query string, executionTime time.Duration) {
	// 简化的索引建议生成
	if executionTime > 500*time.Millisecond && strings.Contains(strings.ToLower(query), "where") {
		// 分析WHERE条件中的列
		suggestion := IndexSuggestion{
			Table:   "users", // 简化
			Columns: []string{"id", "status"},
			Type:    IndexBTree,
			Benefit: float64(executionTime.Milliseconds()),
			Cost:    100.0, // 估算成本
			Reason:  "WHERE子句中的条件列缺少索引",
		}

		qa.indexSuggestions = append(qa.indexSuggestions, suggestion)
	}
}

func (cm *CacheManager) CreateCache(name string, maxSize int64, policy CachePolicy) *Cache {
	cache := &Cache{
		Name:    name,
		MaxSize: maxSize,
		Items:   make(map[string]*CacheItem),
		Policy:  policy,
	}

	cm.mutex.Lock()
	cm.caches[name] = cache
	cm.policies[name] = policy
	cm.mutex.Unlock()

	fmt.Printf("创建缓存: %s (最大大小: %d, 策略: %v)\n", name, maxSize, policy)
	return cache
}

func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	item, exists := cache.Items[key]
	if !exists {
		atomic.AddInt64(&cache.MissCount, 1)
		return nil, false
	}

	// 检查过期
	if !item.Expiry.IsZero() && time.Now().After(item.Expiry) {
		delete(cache.Items, key)
		atomic.AddInt64(&cache.MissCount, 1)
		return nil, false
	}

	// 更新访问统计
	item.AccessTime = time.Now()
	item.AccessCount++
	atomic.AddInt64(&cache.HitCount, 1)

	return item.Value, true
}

func (cache *Cache) Set(key string, value interface{}, ttl time.Duration) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	item := &CacheItem{
		Key:         key,
		Value:       value,
		AccessTime:  time.Now(),
		AccessCount: 1,
		Size:        100, // 简化的大小计算
	}

	if ttl > 0 {
		item.Expiry = time.Now().Add(ttl)
	}

	// 检查是否需要驱逐
	if cache.Size+item.Size > cache.MaxSize {
		cache.evict()
	}

	cache.Items[key] = item
	cache.Size += item.Size
}

func (cache *Cache) evict() {
	switch cache.Policy {
	case LRU:
		cache.evictLRU()
	case LFU:
		cache.evictLFU()
	case FIFO:
		cache.evictFIFO()
	}
}

func (cache *Cache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time = time.Now()

	for key, item := range cache.Items {
		if item.AccessTime.Before(oldestTime) {
			oldestTime = item.AccessTime
			oldestKey = key
		}
	}

	if oldestKey != "" {
		cache.Size -= cache.Items[oldestKey].Size
		delete(cache.Items, oldestKey)
	}
}

func (cache *Cache) evictLFU() {
	var leastKey string
	var leastCount int64 = math.MaxInt64

	for key, item := range cache.Items {
		if item.AccessCount < leastCount {
			leastCount = item.AccessCount
			leastKey = key
		}
	}

	if leastKey != "" {
		cache.Size -= cache.Items[leastKey].Size
		delete(cache.Items, leastKey)
	}
}

func (cache *Cache) evictFIFO() {
	// FIFO实现需要维护插入顺序，这里简化处理
	for key, item := range cache.Items {
		cache.Size -= item.Size
		delete(cache.Items, key)
		break
	}
}

// ==================
// 4. 负载均衡优化器
// ==================

// LoadBalancerOptimizer 负载均衡优化器
type LoadBalancerOptimizer struct {
	algorithms map[string]LoadBalanceAlgorithm
	backends   map[string][]*Backend
	metrics    LoadBalancerMetrics
	mutex      sync.RWMutex
}

// LoadBalanceAlgorithm 负载均衡算法
type LoadBalanceAlgorithm interface {
	SelectBackend(backends []*Backend, request *Request) *Backend
	Name() string
}

// Backend 后端服务器
type Backend struct {
	ID          string
	Address     string
	Port        int
	Weight      int
	Health      HealthStatus
	Metrics     BackendMetrics
	CurrentLoad int32
	MaxLoad     int32
}

// BackendMetrics 后端指标
type BackendMetrics struct {
	ActiveConnections int32
	TotalRequests     int64
	SuccessRequests   int64
	FailedRequests    int64
	ResponseTime      time.Duration
	LastHealthCheck   time.Time
	Throughput        float64
}

// Request 请求
type Request struct {
	ID       string
	ClientIP string
	Path     string
	Method   string
	Headers  map[string]string
	Body     []byte
	Weight   int
}

// LoadBalancerMetrics 负载均衡器指标
type LoadBalancerMetrics struct {
	TotalRequests    int64
	BalancedRequests int64
	FailedRequests   int64
	AverageLatency   time.Duration
	AlgorithmMetrics map[string]AlgorithmMetrics
}

// AlgorithmMetrics 算法指标
type AlgorithmMetrics struct {
	Name            string
	Requests        int64
	AverageLatency  time.Duration
	Distribution    map[string]int64 // 后端ID -> 请求数
	EfficiencyScore float64
}

func NewLoadBalancerOptimizer() *LoadBalancerOptimizer {
	optimizer := &LoadBalancerOptimizer{
		algorithms: make(map[string]LoadBalanceAlgorithm),
		backends:   make(map[string][]*Backend),
		metrics: LoadBalancerMetrics{
			AlgorithmMetrics: make(map[string]AlgorithmMetrics),
		},
	}

	// 注册负载均衡算法
	optimizer.RegisterAlgorithm(&RoundRobinAlgorithm{})
	optimizer.RegisterAlgorithm(&WeightedRoundRobinAlgorithm{})
	optimizer.RegisterAlgorithm(&LeastConnectionsAlgorithm{})
	optimizer.RegisterAlgorithm(&ConsistentHashAlgorithm{})

	return optimizer
}

func (lbo *LoadBalancerOptimizer) RegisterAlgorithm(algorithm LoadBalanceAlgorithm) {
	lbo.mutex.Lock()
	defer lbo.mutex.Unlock()

	lbo.algorithms[algorithm.Name()] = algorithm
	lbo.metrics.AlgorithmMetrics[algorithm.Name()] = AlgorithmMetrics{
		Name:         algorithm.Name(),
		Distribution: make(map[string]int64),
	}

	fmt.Printf("注册负载均衡算法: %s\n", algorithm.Name())
}

func (lbo *LoadBalancerOptimizer) AddBackend(pool string, backend *Backend) {
	lbo.mutex.Lock()
	defer lbo.mutex.Unlock()

	lbo.backends[pool] = append(lbo.backends[pool], backend)
	fmt.Printf("添加后端服务器: %s -> %s:%d\n", pool, backend.Address, backend.Port)
}

func (lbo *LoadBalancerOptimizer) BenchmarkAlgorithms(pool string, requests []*Request) {
	lbo.mutex.RLock()
	backends := lbo.backends[pool]
	algorithms := make([]LoadBalanceAlgorithm, 0, len(lbo.algorithms))
	for _, algo := range lbo.algorithms {
		algorithms = append(algorithms, algo)
	}
	lbo.mutex.RUnlock()

	fmt.Printf("=== 负载均衡算法性能测试 (池: %s, 请求数: %d) ===\n", pool, len(requests))

	for _, algorithm := range algorithms {
		start := time.Now()
		distribution := make(map[string]int64)

		// 重置后端状态
		for _, backend := range backends {
			atomic.StoreInt32(&backend.CurrentLoad, 0)
		}

		// 执行负载均衡
		for _, request := range requests {
			backend := algorithm.SelectBackend(backends, request)
			if backend != nil {
				distribution[backend.ID]++
				atomic.AddInt32(&backend.CurrentLoad, 1)
			}
		}

		elapsed := time.Since(start)

		// 计算分布均匀性
		uniformity := lbo.calculateUniformity(distribution, len(backends))

		fmt.Printf("算法: %s\n", algorithm.Name())
		fmt.Printf("  执行时间: %v\n", elapsed)
		fmt.Printf("  分布均匀性: %.2f%%\n", uniformity*100)
		fmt.Printf("  请求分布:\n")
		for backendID, count := range distribution {
			percentage := float64(count) / float64(len(requests)) * 100
			fmt.Printf("    %s: %d (%.1f%%)\n", backendID, count, percentage)
		}
		fmt.Println()

		// 更新指标
		lbo.mutex.Lock()
		metrics := lbo.metrics.AlgorithmMetrics[algorithm.Name()]
		metrics.Requests += int64(len(requests))
		metrics.AverageLatency = elapsed / time.Duration(len(requests))
		metrics.Distribution = distribution
		metrics.EfficiencyScore = uniformity
		lbo.metrics.AlgorithmMetrics[algorithm.Name()] = metrics
		lbo.mutex.Unlock()
	}
}

func (lbo *LoadBalancerOptimizer) calculateUniformity(distribution map[string]int64, backendCount int) float64 {
	if len(distribution) == 0 {
		return 0
	}

	var total int64
	for _, count := range distribution {
		total += count
	}

	expectedPerBackend := float64(total) / float64(backendCount)
	variance := 0.0

	for _, count := range distribution {
		diff := float64(count) - expectedPerBackend
		variance += diff * diff
	}

	variance /= float64(len(distribution))
	stdDev := math.Sqrt(variance)

	// 归一化到0-1范围，均匀性越高值越大
	uniformity := 1.0 - (stdDev / expectedPerBackend)
	if uniformity < 0 {
		uniformity = 0
	}

	return uniformity
}

// ==================
// 4.1 负载均衡算法实现
// ==================

// RoundRobinAlgorithm 轮询算法
type RoundRobinAlgorithm struct {
	current int32
}

func (rr *RoundRobinAlgorithm) Name() string {
	return "round_robin"
}

func (rr *RoundRobinAlgorithm) SelectBackend(backends []*Backend, request *Request) *Backend {
	if len(backends) == 0 {
		return nil
	}

	// 过滤健康的后端
	healthy := make([]*Backend, 0)
	for _, backend := range backends {
		if backend.Health.Status == "healthy" {
			healthy = append(healthy, backend)
		}
	}

	if len(healthy) == 0 {
		return nil
	}

	index := atomic.AddInt32(&rr.current, 1) % int32(len(healthy))
	return healthy[index]
}

// WeightedRoundRobinAlgorithm 加权轮询算法
type WeightedRoundRobinAlgorithm struct {
	weights []int
	current int
	mutex   sync.Mutex
}

func (wrr *WeightedRoundRobinAlgorithm) Name() string {
	return "weighted_round_robin"
}

func (wrr *WeightedRoundRobinAlgorithm) SelectBackend(backends []*Backend, request *Request) *Backend {
	if len(backends) == 0 {
		return nil
	}

	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	// 初始化权重
	if len(wrr.weights) != len(backends) {
		wrr.weights = make([]int, len(backends))
		for i, backend := range backends {
			wrr.weights[i] = backend.Weight
		}
	}

	// 找到最大权重的后端
	maxWeight := -1
	selectedIndex := -1

	for i, backend := range backends {
		if backend.Health.Status == "healthy" && wrr.weights[i] > maxWeight {
			maxWeight = wrr.weights[i]
			selectedIndex = i
		}
	}

	if selectedIndex == -1 {
		return nil
	}

	// 减少选中后端的权重
	wrr.weights[selectedIndex]--

	// 如果所有权重都为0，重置
	allZero := true
	for _, weight := range wrr.weights {
		if weight > 0 {
			allZero = false
			break
		}
	}

	if allZero {
		for i, backend := range backends {
			wrr.weights[i] = backend.Weight
		}
	}

	return backends[selectedIndex]
}

// LeastConnectionsAlgorithm 最少连接算法
type LeastConnectionsAlgorithm struct{}

func (lc *LeastConnectionsAlgorithm) Name() string {
	return "least_connections"
}

func (lc *LeastConnectionsAlgorithm) SelectBackend(backends []*Backend, request *Request) *Backend {
	if len(backends) == 0 {
		return nil
	}

	var selected *Backend
	minConnections := int32(math.MaxInt32)

	for _, backend := range backends {
		if backend.Health.Status == "healthy" {
			connections := atomic.LoadInt32(&backend.CurrentLoad)
			if connections < minConnections {
				minConnections = connections
				selected = backend
			}
		}
	}

	return selected
}

// ConsistentHashAlgorithm 一致性哈希算法
type ConsistentHashAlgorithm struct {
	hashRing     map[uint32]*Backend
	sortedHashes []uint32
	mutex        sync.RWMutex
}

func (ch *ConsistentHashAlgorithm) Name() string {
	return "consistent_hash"
}

func (ch *ConsistentHashAlgorithm) SelectBackend(backends []*Backend, request *Request) *Backend {
	if len(backends) == 0 {
		return nil
	}

	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	// 构建哈希环
	if ch.hashRing == nil {
		ch.buildHashRing(backends)
	}

	// 计算请求的哈希值
	key := request.ClientIP + request.Path
	hash := ch.hash(key)

	// 在哈希环上找到对应的后端
	return ch.findBackend(hash)
}

func (ch *ConsistentHashAlgorithm) buildHashRing(backends []*Backend) {
	ch.hashRing = make(map[uint32]*Backend)
	ch.sortedHashes = make([]uint32, 0)

	// 为每个后端创建多个虚拟节点
	virtualNodes := 150
	for _, backend := range backends {
		if backend.Health.Status == "healthy" {
			for i := 0; i < virtualNodes; i++ {
				key := fmt.Sprintf("%s:%d#%d", backend.Address, backend.Port, i)
				hash := ch.hash(key)
				ch.hashRing[hash] = backend
				ch.sortedHashes = append(ch.sortedHashes, hash)
			}
		}
	}

	// 排序哈希值
	sort.Slice(ch.sortedHashes, func(i, j int) bool {
		return ch.sortedHashes[i] < ch.sortedHashes[j]
	})
}

func (ch *ConsistentHashAlgorithm) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (ch *ConsistentHashAlgorithm) findBackend(hash uint32) *Backend {
	if len(ch.sortedHashes) == 0 {
		return nil
	}

	// 二分查找第一个大于等于hash的节点
	idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
		return ch.sortedHashes[i] >= hash
	})

	// 如果没找到，使用第一个节点（环形）
	if idx == len(ch.sortedHashes) {
		idx = 0
	}

	return ch.hashRing[ch.sortedHashes[idx]]
}

// ==================
// 5. 采样器和导出器实现
// ==================

// ProbabilisticSampler 概率采样器
type ProbabilisticSampler struct {
	rate float64
}

func (ps *ProbabilisticSampler) ShouldSample(traceID string, operationName string) bool {
	return rand.Float64() < ps.rate
}

func (ps *ProbabilisticSampler) GetSamplingRate() float64 {
	return ps.rate
}

// ConsoleExporter 控制台导出器
type ConsoleExporter struct{}

func (ce *ConsoleExporter) ExportSpans(spans []*DistributedSpan) error {
	fmt.Printf("导出 %d 个spans到控制台\n", len(spans))
	return nil
}

func (ce *ConsoleExporter) ExportTraces(traces []*DistributedTrace) error {
	fmt.Printf("导出 %d 个traces到控制台\n", len(traces))
	return nil
}

// ==================
// 6. 主演示函数
// ==================

func demonstrateDistributedProfiling() {
	fmt.Println("=== Go分布式性能分析演示 ===")

	// 1. 分布式追踪演示
	fmt.Println("\n1. 分布式追踪系统演示")
	tracer := NewDistributedTracer()

	// 注册服务
	userService := &ServiceNode{
		Name:    "user-service",
		Version: "v1.0.0",
		Instances: []*ServiceInstance{
			{ID: "user-1", Address: "10.0.1.1", Port: 8080, Status: InstanceStatusHealthy},
			{ID: "user-2", Address: "10.0.1.2", Port: 8080, Status: InstanceStatusHealthy},
		},
	}

	orderService := &ServiceNode{
		Name:    "order-service",
		Version: "v1.2.0",
		Instances: []*ServiceInstance{
			{ID: "order-1", Address: "10.0.2.1", Port: 8080, Status: InstanceStatusHealthy},
		},
	}

	tracer.RegisterService(userService)
	tracer.RegisterService(orderService)

	// 模拟分布式调用链
	trace := tracer.StartTrace("api-gateway", "process_order", "", "")

	// API Gateway -> User Service
	userSpan := tracer.StartSpan(trace.TraceID, trace.RootSpan.SpanID, "user-service", "get_user")
	time.Sleep(50 * time.Millisecond) // 模拟处理时间
	tracer.FinishSpan(userSpan.SpanID)

	// API Gateway -> Order Service
	orderSpan := tracer.StartSpan(trace.TraceID, trace.RootSpan.SpanID, "order-service", "create_order")

	// Order Service -> User Service (验证用户)
	verifySpan := tracer.StartSpan(trace.TraceID, orderSpan.SpanID, "user-service", "verify_user")
	time.Sleep(30 * time.Millisecond)
	tracer.FinishSpan(verifySpan.SpanID)

	time.Sleep(80 * time.Millisecond) // Order处理时间
	tracer.FinishSpan(orderSpan.SpanID)

	tracer.FinishSpan(trace.RootSpan.SpanID)

	// 分析追踪结果
	fmt.Printf("追踪分析结果:\n")
	fmt.Printf("  追踪ID: %s\n", trace.TraceID)
	fmt.Printf("  总耗时: %v\n", trace.Duration)
	fmt.Printf("  Span数量: %d\n", len(trace.Spans))
	fmt.Printf("  服务数量: %d\n", len(trace.Services))

	if trace.BottleneckSpan != nil {
		fmt.Printf("  瓶颈Span: %s.%s (%v)\n",
			trace.BottleneckSpan.ServiceName,
			trace.BottleneckSpan.OperationName,
			trace.BottleneckSpan.Duration)
	}

	// 2. 微服务性能分析
	fmt.Println("\n2. 微服务性能分析")
	analyzer := NewMicroserviceAnalyzer()

	// 注册服务和SLA
	userSLA := &SLA{
		ServiceName:     "user-service",
		ResponseTimeP95: 100 * time.Millisecond,
		ResponseTimeP99: 200 * time.Millisecond,
		Availability:    99.9,
		ErrorRate:       1.0,
		TimeWindow:      24 * time.Hour,
	}

	analyzer.RegisterService("user-service", userSLA)
	analyzer.RegisterService("order-service", nil)

	// 模拟服务指标
	userMetrics := &MicroserviceMetrics{
		ServiceName: "user-service",
		RequestRate: 1500.0,
		ErrorRate:   0.5,
		ResponseTime: ResponseTimeMetrics{
			P50:  45 * time.Millisecond,
			P90:  85 * time.Millisecond,
			P95:  120 * time.Millisecond, // 超过SLA
			P99:  180 * time.Millisecond,
			Mean: 55 * time.Millisecond,
		},
		Throughput:   1499.0,
		Availability: 99.95,
		ResourceUsage: ResourceMetrics{
			CPUUsage:       85.5, // 高CPU使用率
			MemoryUsage:    72.3,
			DiskUsage:      45.2,
			NetworkIO:      125.6,
			GoroutineCount: 250,
		},
	}

	analyzer.UpdateMetrics("user-service", userMetrics)

	// 显示优化建议
	recommendations := analyzer.GetRecommendations()
	fmt.Printf("性能优化建议 (%d条):\n", len(recommendations))
	for i, rec := range recommendations {
		if i >= 3 { // 只显示前3条
			break
		}
		fmt.Printf("  %d. [%s] %s - %s\n",
			i+1, rec.Service, rec.Description, rec.Impact)
	}

	// 3. 数据库性能优化
	fmt.Println("\n3. 数据库性能优化")
	dbOptimizer := NewDatabaseOptimizer()

	// 创建连接池
	dbOptimizer.CreateConnectionPool("primary", 20, "postgres://localhost/app")

	// 分析查询性能
	queries := []struct {
		sql  string
		time time.Duration
	}{
		{"SELECT * FROM users WHERE id = 1", 45 * time.Millisecond},
		{"SELECT * FROM orders WHERE user_id = 1 AND status = 'pending'", 250 * time.Millisecond},
		{"SELECT COUNT(*) FROM products WHERE category = 'electronics'", 1200 * time.Millisecond}, // 慢查询
	}

	for _, q := range queries {
		dbOptimizer.queryAnalyzer.AnalyzeQuery(q.sql, q.time, "app")
	}

	fmt.Printf("查询分析结果:\n")
	fmt.Printf("  慢查询数量: %d\n", len(dbOptimizer.queryAnalyzer.slowQueries))
	fmt.Printf("  查询模式数量: %d\n", len(dbOptimizer.queryAnalyzer.queryPatterns))
	fmt.Printf("  索引建议数量: %d\n", len(dbOptimizer.queryAnalyzer.indexSuggestions))

	// 4. 负载均衡优化
	fmt.Println("\n4. 负载均衡算法性能对比")
	lbOptimizer := NewLoadBalancerOptimizer()

	// 添加后端服务器
	backends := []*Backend{
		{ID: "backend-1", Address: "10.0.1.1", Port: 8080, Weight: 5, Health: HealthStatus{Status: "healthy"}},
		{ID: "backend-2", Address: "10.0.1.2", Port: 8080, Weight: 3, Health: HealthStatus{Status: "healthy"}},
		{ID: "backend-3", Address: "10.0.1.3", Port: 8080, Weight: 2, Health: HealthStatus{Status: "healthy"}},
	}

	for _, backend := range backends {
		lbOptimizer.AddBackend("web-pool", backend)
	}

	// 生成测试请求
	requests := make([]*Request, 1000)
	for i := range requests {
		requests[i] = &Request{
			ID:       fmt.Sprintf("req-%d", i),
			ClientIP: fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			Path:     fmt.Sprintf("/api/v1/users/%d", rand.Intn(1000)),
			Method:   "GET",
		}
	}

	// 基准测试各种算法
	lbOptimizer.BenchmarkAlgorithms("web-pool", requests)

	// 5. 性能指标汇总
	fmt.Println("5. 分布式系统性能指标汇总")
	fmt.Printf("分布式追踪指标:\n")
	fmt.Printf("  生成的追踪数: %d\n", tracer.metrics.TracesGenerated)
	fmt.Printf("  生成的Span数: %d\n", tracer.metrics.SpansGenerated)
	fmt.Printf("  采样率: %.1f%%\n", tracer.sampler.GetSamplingRate()*100)

	fmt.Printf("\n微服务分析指标:\n")
	fmt.Printf("  监控的服务数: %d\n", len(analyzer.services))
	fmt.Printf("  活跃告警数: %d\n", len(analyzer.alerts))
	fmt.Printf("  优化建议数: %d\n", len(analyzer.recommendations))

	fmt.Printf("\n负载均衡指标:\n")
	for name, metrics := range lbOptimizer.metrics.AlgorithmMetrics {
		fmt.Printf("  %s: 效率分数 %.2f\n", name, metrics.EfficiencyScore)
	}
}

func main() {
	demonstrateDistributedProfiling()

	fmt.Println("\n=== Go分布式性能分析演示完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 分布式追踪：跨服务的完整请求链路监控")
	fmt.Println("2. 微服务分析：服务级别的性能瓶颈识别")
	fmt.Println("3. 数据库优化：查询分析和索引优化建议")
	fmt.Println("4. 负载均衡：算法性能对比和选择策略")
	fmt.Println("5. 系统监控：分布式系统健康状态管理")
	fmt.Println("6. SLA管理：服务等级协议违反检测")
	fmt.Println("7. 容量规划：基于指标的扩缩容决策")

	fmt.Println("\n分布式系统特性:")
	fmt.Println("- 链路追踪提供端到端的性能可观测性")
	fmt.Println("- 依赖图分析帮助识别关键服务和单点故障")
	fmt.Println("- 智能采样降低性能监控的开销")
	fmt.Println("- 多维度指标支持复杂的性能分析")
	fmt.Println("- 自动化的性能异常检测和告警")
	fmt.Println("- 基于机器学习的性能优化建议")
}

/*
=== 练习题 ===

1. 分布式追踪进阶：
   - 实现自适应采样算法
   - 添加业务指标追踪
   - 实现追踪数据压缩
   - 创建分布式调用图可视化

2. 微服务监控优化：
   - 实现服务网格性能监控
   - 添加业务SLI定义
   - 实现多租户性能隔离
   - 创建性能异常根因分析

3. 数据库性能调优：
   - 实现查询执行计划分析
   - 添加分布式事务监控
   - 实现连接池智能调优
   - 创建数据库瓶颈预测

4. 负载均衡高级特性：
   - 实现动态权重调整
   - 添加地理位置感知路由
   - 实现故障转移和恢复
   - 创建负载预测模型

工具集成：
- Jaeger/Zipkin 分布式追踪
- Prometheus/Grafana 指标监控
- Elasticsearch/Kibana 日志分析
- Istio 服务网格监控

重要概念：
- Distributed Tracing: 分布式请求追踪
- Service Mesh: 服务网格架构
- Circuit Breaker: 熔断器模式
- Bulkhead: 舱壁隔离模式
- Saga Pattern: 分布式事务模式
- CQRS: 命令查询职责分离
*/
