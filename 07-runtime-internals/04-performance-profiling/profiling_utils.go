/*
=== 性能分析工具函数库 ===

提供可复用的性能分析、监控和诊断工具函数。
这些函数可以在生产环境中使用，帮助诊断和优化应用性能。

主要功能：
1. 性能指标收集器
2. 延迟直方图
3. 吞吐量计算器
4. 性能基线比较
5. 自动性能报告生成
*/

package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// 1. 性能指标收集器
// ==================

// MetricsCollector 性能指标收集器
// 用于收集和聚合各种性能指标
type MetricsCollector struct {
	// 计数器指标
	counters map[string]*atomic.Int64
	// 计量器指标
	gauges map[string]*atomic.Int64
	// 直方图指标
	histograms map[string]*LatencyHistogram
	// 互斥锁
	mu sync.RWMutex
	// 开始时间
	startTime time.Time
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		counters:   make(map[string]*atomic.Int64),
		gauges:     make(map[string]*atomic.Int64),
		histograms: make(map[string]*LatencyHistogram),
		startTime:  time.Now(),
	}
}

// IncrementCounter 增加计数器
func (c *MetricsCollector) IncrementCounter(name string) {
	c.mu.RLock()
	counter, exists := c.counters[name]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		if _, exists := c.counters[name]; !exists {
			c.counters[name] = &atomic.Int64{}
		}
		counter = c.counters[name]
		c.mu.Unlock()
	}

	counter.Add(1)
}

// AddCounter 增加计数器指定值
func (c *MetricsCollector) AddCounter(name string, delta int64) {
	c.mu.RLock()
	counter, exists := c.counters[name]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		if _, exists := c.counters[name]; !exists {
			c.counters[name] = &atomic.Int64{}
		}
		counter = c.counters[name]
		c.mu.Unlock()
	}

	counter.Add(delta)
}

// SetGauge 设置计量器值
func (c *MetricsCollector) SetGauge(name string, value int64) {
	c.mu.RLock()
	gauge, exists := c.gauges[name]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		if _, exists := c.gauges[name]; !exists {
			c.gauges[name] = &atomic.Int64{}
		}
		gauge = c.gauges[name]
		c.mu.Unlock()
	}

	gauge.Store(value)
}

// RecordLatency 记录延迟
func (c *MetricsCollector) RecordLatency(name string, latency time.Duration) {
	c.mu.RLock()
	histogram, exists := c.histograms[name]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		if _, exists := c.histograms[name]; !exists {
			c.histograms[name] = NewLatencyHistogram()
		}
		histogram = c.histograms[name]
		c.mu.Unlock()
	}

	histogram.Record(latency)
}

// GetCounter 获取计数器值
func (c *MetricsCollector) GetCounter(name string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if counter, exists := c.counters[name]; exists {
		return counter.Load()
	}
	return 0
}

// GetGauge 获取计量器值
func (c *MetricsCollector) GetGauge(name string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if gauge, exists := c.gauges[name]; exists {
		return gauge.Load()
	}
	return 0
}

// GetHistogramStats 获取直方图统计
func (c *MetricsCollector) GetHistogramStats(name string) HistogramStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if histogram, exists := c.histograms[name]; exists {
		return histogram.GetStats()
	}
	return HistogramStats{}
}

// GetAllMetrics 获取所有指标
func (c *MetricsCollector) GetAllMetrics() MetricsSummary {
	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := MetricsSummary{
		Timestamp: time.Now(),
		Uptime:    time.Since(c.startTime),
		Counters:  make(map[string]int64),
		Gauges:    make(map[string]int64),
		Latencies: make(map[string]HistogramStats),
	}

	for name, counter := range c.counters {
		summary.Counters[name] = counter.Load()
	}

	for name, gauge := range c.gauges {
		summary.Gauges[name] = gauge.Load()
	}

	for name, histogram := range c.histograms {
		summary.Latencies[name] = histogram.GetStats()
	}

	return summary
}

// MetricsSummary 指标摘要
type MetricsSummary struct {
	Timestamp time.Time
	Uptime    time.Duration
	Counters  map[string]int64
	Gauges    map[string]int64
	Latencies map[string]HistogramStats
}

// String 格式化输出
func (s MetricsSummary) String() string {
	result := fmt.Sprintf("=== 性能指标摘要 ===\n")
	result += fmt.Sprintf("时间: %s\n", s.Timestamp.Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("运行时间: %v\n", s.Uptime)

	if len(s.Counters) > 0 {
		result += "\n计数器:\n"
		for name, value := range s.Counters {
			result += fmt.Sprintf("  %s: %d\n", name, value)
		}
	}

	if len(s.Gauges) > 0 {
		result += "\n计量器:\n"
		for name, value := range s.Gauges {
			result += fmt.Sprintf("  %s: %d\n", name, value)
		}
	}

	if len(s.Latencies) > 0 {
		result += "\n延迟统计:\n"
		for name, stats := range s.Latencies {
			result += fmt.Sprintf("  %s:\n", name)
			result += fmt.Sprintf("    Count: %d, Avg: %v, P99: %v\n",
				stats.Count, stats.Avg, stats.P99)
		}
	}

	return result
}

// ==================
// 2. 延迟直方图
// ==================

// LatencyHistogram 延迟直方图
// 用于记录和分析延迟分布
type LatencyHistogram struct {
	values []time.Duration
	mu     sync.Mutex
	max    int // 最大记录数
}

// NewLatencyHistogram 创建新的延迟直方图
func NewLatencyHistogram() *LatencyHistogram {
	return &LatencyHistogram{
		values: make([]time.Duration, 0, 10000),
		max:    100000,
	}
}

// Record 记录延迟值
func (h *LatencyHistogram) Record(latency time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.values = append(h.values, latency)

	// 限制记录数量
	if len(h.values) > h.max {
		// 保留后半部分
		h.values = h.values[len(h.values)/2:]
	}
}

// GetStats 获取统计信息
func (h *LatencyHistogram) GetStats() HistogramStats {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.values) == 0 {
		return HistogramStats{}
	}

	// 复制并排序
	sorted := make([]time.Duration, len(h.values))
	copy(sorted, h.values)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	var total time.Duration
	for _, v := range sorted {
		total += v
	}

	n := len(sorted)
	return HistogramStats{
		Count: n,
		Min:   sorted[0],
		Max:   sorted[n-1],
		Avg:   total / time.Duration(n),
		P50:   sorted[n*50/100],
		P90:   sorted[n*90/100],
		P95:   sorted[n*95/100],
		P99:   sorted[n*99/100],
	}
}

// Reset 重置直方图
func (h *LatencyHistogram) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.values = h.values[:0]
}

// HistogramStats 直方图统计
type HistogramStats struct {
	Count int
	Min   time.Duration
	Max   time.Duration
	Avg   time.Duration
	P50   time.Duration
	P90   time.Duration
	P95   time.Duration
	P99   time.Duration
}

// ==================
// 3. 吞吐量计算器
// ==================

// ThroughputCalculator 吞吐量计算器
// 用于计算操作的吞吐量
type ThroughputCalculator struct {
	// 操作计数
	count atomic.Int64
	// 开始时间
	startTime time.Time
	// 窗口大小
	windowSize time.Duration
	// 窗口计数
	windowCounts []windowCount
	// 互斥锁
	mu sync.Mutex
}

type windowCount struct {
	timestamp time.Time
	count     int64
}

// NewThroughputCalculator 创建新的吞吐量计算器
func NewThroughputCalculator(windowSize time.Duration) *ThroughputCalculator {
	return &ThroughputCalculator{
		startTime:    time.Now(),
		windowSize:   windowSize,
		windowCounts: make([]windowCount, 0),
	}
}

// Record 记录一次操作
func (c *ThroughputCalculator) Record() {
	c.count.Add(1)

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.windowCounts = append(c.windowCounts, windowCount{
		timestamp: now,
		count:     1,
	})

	// 清理过期的窗口计数
	cutoff := now.Add(-c.windowSize)
	for len(c.windowCounts) > 0 && c.windowCounts[0].timestamp.Before(cutoff) {
		c.windowCounts = c.windowCounts[1:]
	}
}

// RecordN 记录多次操作
func (c *ThroughputCalculator) RecordN(n int64) {
	c.count.Add(n)

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.windowCounts = append(c.windowCounts, windowCount{
		timestamp: now,
		count:     n,
	})

	// 清理过期的窗口计数
	cutoff := now.Add(-c.windowSize)
	for len(c.windowCounts) > 0 && c.windowCounts[0].timestamp.Before(cutoff) {
		c.windowCounts = c.windowCounts[1:]
	}
}

// GetTotalThroughput 获取总吞吐量（操作/秒）
func (c *ThroughputCalculator) GetTotalThroughput() float64 {
	elapsed := time.Since(c.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(c.count.Load()) / elapsed
}

// GetWindowThroughput 获取窗口吞吐量（操作/秒）
func (c *ThroughputCalculator) GetWindowThroughput() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.windowCounts) == 0 {
		return 0
	}

	var total int64
	for _, wc := range c.windowCounts {
		total += wc.count
	}

	return float64(total) / c.windowSize.Seconds()
}

// GetStats 获取吞吐量统计
func (c *ThroughputCalculator) GetStats() ThroughputStats {
	return ThroughputStats{
		TotalCount:       c.count.Load(),
		TotalThroughput:  c.GetTotalThroughput(),
		WindowThroughput: c.GetWindowThroughput(),
		Uptime:           time.Since(c.startTime),
	}
}

// ThroughputStats 吞吐量统计
type ThroughputStats struct {
	TotalCount       int64
	TotalThroughput  float64
	WindowThroughput float64
	Uptime           time.Duration
}

// ==================
// 4. 性能基线比较器
// ==================

// PerformanceBaseline 性能基线
type PerformanceBaseline struct {
	Name      string
	Timestamp time.Time
	Metrics   map[string]float64
}

// BaselineComparator 基线比较器
type BaselineComparator struct {
	baselines map[string]PerformanceBaseline
	mu        sync.RWMutex
}

// NewBaselineComparator 创建新的基线比较器
func NewBaselineComparator() *BaselineComparator {
	return &BaselineComparator{
		baselines: make(map[string]PerformanceBaseline),
	}
}

// SetBaseline 设置基线
func (c *BaselineComparator) SetBaseline(name string, metrics map[string]float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.baselines[name] = PerformanceBaseline{
		Name:      name,
		Timestamp: time.Now(),
		Metrics:   metrics,
	}
}

// Compare 比较当前指标与基线
func (c *BaselineComparator) Compare(baselineName string, current map[string]float64) ComparisonResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	baseline, exists := c.baselines[baselineName]
	if !exists {
		return ComparisonResult{
			BaselineName: baselineName,
			Error:        "基线不存在",
		}
	}

	result := ComparisonResult{
		BaselineName:      baselineName,
		BaselineTimestamp: baseline.Timestamp,
		Comparisons:       make(map[string]MetricComparison),
	}

	for name, currentValue := range current {
		baselineValue, exists := baseline.Metrics[name]
		if !exists {
			continue
		}

		var changePercent float64
		if baselineValue != 0 {
			changePercent = (currentValue - baselineValue) / baselineValue * 100
		}

		result.Comparisons[name] = MetricComparison{
			Baseline:      baselineValue,
			Current:       currentValue,
			ChangePercent: changePercent,
			Improved:      currentValue < baselineValue, // 假设越小越好
		}
	}

	return result
}

// ComparisonResult 比较结果
type ComparisonResult struct {
	BaselineName      string
	BaselineTimestamp time.Time
	Comparisons       map[string]MetricComparison
	Error             string
}

// MetricComparison 指标比较
type MetricComparison struct {
	Baseline      float64
	Current       float64
	ChangePercent float64
	Improved      bool
}

// String 格式化输出
func (r ComparisonResult) String() string {
	if r.Error != "" {
		return fmt.Sprintf("比较错误: %s", r.Error)
	}

	result := fmt.Sprintf("=== 性能基线比较 ===\n")
	result += fmt.Sprintf("基线: %s (%s)\n\n",
		r.BaselineName, r.BaselineTimestamp.Format("2006-01-02 15:04:05"))

	for name, comp := range r.Comparisons {
		status := "退化"
		if comp.Improved {
			status = "改善"
		}
		result += fmt.Sprintf("%s:\n", name)
		result += fmt.Sprintf("  基线: %.2f, 当前: %.2f\n", comp.Baseline, comp.Current)
		result += fmt.Sprintf("  变化: %.2f%% (%s)\n\n", comp.ChangePercent, status)
	}

	return result
}

// ==================
// 5. 性能报告生成器
// ==================

// PerformanceReportGenerator 性能报告生成器
type PerformanceReportGenerator struct {
	collector *MetricsCollector
}

// NewPerformanceReportGenerator 创建新的报告生成器
func NewPerformanceReportGenerator(collector *MetricsCollector) *PerformanceReportGenerator {
	return &PerformanceReportGenerator{
		collector: collector,
	}
}

// GenerateReport 生成性能报告
func (g *PerformanceReportGenerator) GenerateReport() PerformanceReport {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	metrics := g.collector.GetAllMetrics()

	return PerformanceReport{
		Timestamp: time.Now(),
		Uptime:    metrics.Uptime,

		// 运行时指标
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),

		// 内存指标
		HeapAlloc:   ms.HeapAlloc,
		HeapSys:     ms.HeapSys,
		HeapObjects: ms.HeapObjects,
		StackInuse:  ms.StackInuse,

		// GC 指标
		NumGC:        ms.NumGC,
		GCPauseTotal: time.Duration(ms.PauseTotalNs),

		// 自定义指标
		Counters:  metrics.Counters,
		Gauges:    metrics.Gauges,
		Latencies: metrics.Latencies,
	}
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	Timestamp time.Time
	Uptime    time.Duration

	// 运行时指标
	NumGoroutine int
	NumCPU       int
	GOMAXPROCS   int

	// 内存指标
	HeapAlloc   uint64
	HeapSys     uint64
	HeapObjects uint64
	StackInuse  uint64

	// GC 指标
	NumGC        uint32
	GCPauseTotal time.Duration

	// 自定义指标
	Counters  map[string]int64
	Gauges    map[string]int64
	Latencies map[string]HistogramStats
}

// String 格式化输出
func (r PerformanceReport) String() string {
	return fmt.Sprintf(`
=== 性能报告 ===
时间: %s
运行时间: %v

运行时:
  Goroutine 数: %d
  CPU 核心数: %d
  GOMAXPROCS: %d

内存:
  堆分配: %d MB
  堆系统: %d MB
  堆对象: %d
  栈使用: %d KB

GC:
  GC 次数: %d
  GC 总暂停: %v

计数器: %d 个
计量器: %d 个
延迟指标: %d 个
`,
		r.Timestamp.Format("2006-01-02 15:04:05"),
		r.Uptime,
		r.NumGoroutine,
		r.NumCPU,
		r.GOMAXPROCS,
		r.HeapAlloc/(1024*1024),
		r.HeapSys/(1024*1024),
		r.HeapObjects,
		r.StackInuse/1024,
		r.NumGC,
		r.GCPauseTotal,
		len(r.Counters),
		len(r.Gauges),
		len(r.Latencies),
	)
}

// ==================
// 6. 操作计时器
// ==================

// OperationTimer 操作计时器
// 用于测量操作执行时间
type OperationTimer struct {
	name      string
	startTime time.Time
	collector *MetricsCollector
}

// StartTimer 开始计时
func StartTimer(name string, collector *MetricsCollector) *OperationTimer {
	return &OperationTimer{
		name:      name,
		startTime: time.Now(),
		collector: collector,
	}
}

// Stop 停止计时并记录
func (t *OperationTimer) Stop() time.Duration {
	elapsed := time.Since(t.startTime)
	if t.collector != nil {
		t.collector.RecordLatency(t.name, elapsed)
		t.collector.IncrementCounter(t.name + "_count")
	}
	return elapsed
}

// ==================
// 7. 便捷函数
// ==================

// MeasureOperation 测量操作执行时间
func MeasureOperation(name string, op func()) time.Duration {
	start := time.Now()
	op()
	return time.Since(start)
}

// MeasureOperationWithResult 测量操作执行时间（带返回值）
func MeasureOperationWithResult[T any](name string, op func() T) (T, time.Duration) {
	start := time.Now()
	result := op()
	return result, time.Since(start)
}

// BenchmarkOperation 基准测试操作
func BenchmarkOperation(name string, iterations int, op func()) BenchmarkResult {
	// 预热
	for i := 0; i < iterations/10; i++ {
		op()
	}

	// 正式测试
	var totalDuration time.Duration
	durations := make([]time.Duration, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		op()
		durations[i] = time.Since(start)
		totalDuration += durations[i]
	}

	// 排序计算百分位数
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	return BenchmarkResult{
		Name:       name,
		Iterations: iterations,
		TotalTime:  totalDuration,
		AvgTime:    totalDuration / time.Duration(iterations),
		MinTime:    durations[0],
		MaxTime:    durations[iterations-1],
		P50:        durations[iterations*50/100],
		P90:        durations[iterations*90/100],
		P99:        durations[iterations*99/100],
		OpsPerSec:  float64(iterations) / totalDuration.Seconds(),
	}
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Name       string
	Iterations int
	TotalTime  time.Duration
	AvgTime    time.Duration
	MinTime    time.Duration
	MaxTime    time.Duration
	P50        time.Duration
	P90        time.Duration
	P99        time.Duration
	OpsPerSec  float64
}

// String 格式化输出
func (r BenchmarkResult) String() string {
	return fmt.Sprintf(`
基准测试: %s
  迭代次数: %d
  总时间: %v
  平均时间: %v
  最小时间: %v
  最大时间: %v
  P50: %v
  P90: %v
  P99: %v
  吞吐量: %.2f ops/sec
`,
		r.Name,
		r.Iterations,
		r.TotalTime,
		r.AvgTime,
		r.MinTime,
		r.MaxTime,
		r.P50,
		r.P90,
		r.P99,
		r.OpsPerSec,
	)
}

// CalculateStdDev 计算标准差
func CalculateStdDev(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}

	// 计算平均值
	var sum time.Duration
	for _, v := range values {
		sum += v
	}
	mean := sum / time.Duration(len(values))

	// 计算方差
	var variance float64
	for _, v := range values {
		diff := float64(v - mean)
		variance += diff * diff
	}
	variance /= float64(len(values))

	// 返回标准差
	return time.Duration(math.Sqrt(variance))
}

// PrintRuntimeStats 打印运行时统计
func PrintRuntimeStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	fmt.Println("\n=== 运行时统计 ===")
	fmt.Printf("Goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Printf("CPU 核心数: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("堆分配: %d MB\n", ms.HeapAlloc/(1024*1024))
	fmt.Printf("堆对象: %d\n", ms.HeapObjects)
	fmt.Printf("GC 次数: %d\n", ms.NumGC)
	fmt.Printf("GC 总暂停: %v\n", time.Duration(ms.PauseTotalNs))
}
