/*
=== 调度器工具函数库 ===

提供可复用的调度器监控、分析和调优工具函数。
这些函数可以在生产环境中使用，帮助诊断和优化调度器性能。

主要功能：
1. 调度器统计信息收集
2. Goroutine 泄漏检测
3. 调度延迟测量
4. GOMAXPROCS 自动调优
5. 工作负载分析
*/

package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// 1. 调度器统计工具
// ==================

// SchedulerStatsCollector 调度器统计收集器
// 用于收集和分析调度器运行时统计信息
type SchedulerStatsCollector struct {
	// Goroutine 数量历史
	goroutineHistory []int
	// 调度延迟历史
	latencyHistory []time.Duration
	// 收集间隔
	interval time.Duration
	// 最大历史记录数
	maxHistory int
	// 互斥锁
	mu sync.RWMutex
	// 运行状态
	running atomic.Bool
	// 停止信号
	stopCh chan struct{}
}

// NewSchedulerStatsCollector 创建新的调度器统计收集器
func NewSchedulerStatsCollector(interval time.Duration, maxHistory int) *SchedulerStatsCollector {
	return &SchedulerStatsCollector{
		goroutineHistory: make([]int, 0, maxHistory),
		latencyHistory:   make([]time.Duration, 0, maxHistory),
		interval:         interval,
		maxHistory:       maxHistory,
		stopCh:           make(chan struct{}),
	}
}

// Start 启动统计收集
func (c *SchedulerStatsCollector) Start() {
	if !c.running.CompareAndSwap(false, true) {
		return
	}

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.collect()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop 停止统计收集
func (c *SchedulerStatsCollector) Stop() {
	if c.running.CompareAndSwap(true, false) {
		close(c.stopCh)
	}
}

func (c *SchedulerStatsCollector) collect() {
	// 收集 Goroutine 数量
	numGoroutine := runtime.NumGoroutine()

	// 测量调度延迟
	latency := measureSchedulingLatency()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.goroutineHistory = append(c.goroutineHistory, numGoroutine)
	c.latencyHistory = append(c.latencyHistory, latency)

	// 限制历史记录数量
	if len(c.goroutineHistory) > c.maxHistory {
		c.goroutineHistory = c.goroutineHistory[1:]
	}
	if len(c.latencyHistory) > c.maxHistory {
		c.latencyHistory = c.latencyHistory[1:]
	}
}

// measureSchedulingLatency 测量单次调度延迟
func measureSchedulingLatency() time.Duration {
	start := time.Now()
	done := make(chan struct{})

	go func() {
		close(done)
	}()

	<-done
	return time.Since(start)
}

// GetGoroutineStats 获取 Goroutine 统计信息
func (c *SchedulerStatsCollector) GetGoroutineStats() (current, min, max, avg int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.goroutineHistory) == 0 {
		return runtime.NumGoroutine(), 0, 0, 0
	}

	current = c.goroutineHistory[len(c.goroutineHistory)-1]
	min = c.goroutineHistory[0]
	max = c.goroutineHistory[0]
	sum := 0

	for _, n := range c.goroutineHistory {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
		sum += n
	}

	avg = sum / len(c.goroutineHistory)
	return
}

// GetLatencyPercentiles 获取调度延迟百分位数
func (c *SchedulerStatsCollector) GetLatencyPercentiles() (p50, p90, p95, p99 time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.latencyHistory) == 0 {
		return 0, 0, 0, 0
	}

	// 复制并排序
	sorted := make([]time.Duration, len(c.latencyHistory))
	copy(sorted, c.latencyHistory)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	n := len(sorted)
	p50 = sorted[n*50/100]
	p90 = sorted[n*90/100]
	p95 = sorted[n*95/100]
	p99 = sorted[n*99/100]

	return
}

// ==================
// 2. Goroutine 泄漏检测器
// ==================

// GoroutineLeakDetector Goroutine 泄漏检测器
// 用于检测 Goroutine 泄漏问题
type GoroutineLeakDetector struct {
	// 基线 Goroutine 数量
	baseline int
	// 检测阈值（超过基线的百分比）
	threshold float64
	// 检测间隔
	interval time.Duration
	// 泄漏回调
	onLeak func(current, baseline int)
	// 运行状态
	running atomic.Bool
	// 停止信号
	stopCh chan struct{}
}

// NewGoroutineLeakDetector 创建新的 Goroutine 泄漏检测器
// threshold: 泄漏阈值，如 0.5 表示超过基线 50% 时报警
func NewGoroutineLeakDetector(threshold float64, interval time.Duration) *GoroutineLeakDetector {
	return &GoroutineLeakDetector{
		baseline:  runtime.NumGoroutine(),
		threshold: threshold,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// SetLeakCallback 设置泄漏回调函数
func (d *GoroutineLeakDetector) SetLeakCallback(callback func(current, baseline int)) {
	d.onLeak = callback
}

// ResetBaseline 重置基线
func (d *GoroutineLeakDetector) ResetBaseline() {
	d.baseline = runtime.NumGoroutine()
}

// Start 启动泄漏检测
func (d *GoroutineLeakDetector) Start() {
	if !d.running.CompareAndSwap(false, true) {
		return
	}

	go func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				d.check()
			case <-d.stopCh:
				return
			}
		}
	}()
}

// Stop 停止泄漏检测
func (d *GoroutineLeakDetector) Stop() {
	if d.running.CompareAndSwap(true, false) {
		close(d.stopCh)
	}
}

func (d *GoroutineLeakDetector) check() {
	current := runtime.NumGoroutine()
	increase := float64(current-d.baseline) / float64(d.baseline)

	if increase > d.threshold && d.onLeak != nil {
		d.onLeak(current, d.baseline)
	}
}

// CheckNow 立即检测
func (d *GoroutineLeakDetector) CheckNow() (leaked bool, current, baseline int) {
	current = runtime.NumGoroutine()
	baseline = d.baseline
	increase := float64(current-baseline) / float64(baseline)
	leaked = increase > d.threshold
	return
}

// ==================
// 3. 调度延迟测量器
// ==================

// SchedulingLatencyMeasurer 调度延迟测量器
// 用于精确测量 Goroutine 调度延迟
type SchedulingLatencyMeasurer struct {
	// 测量结果
	results []time.Duration
	// 互斥锁
	mu sync.Mutex
}

// NewSchedulingLatencyMeasurer 创建新的调度延迟测量器
func NewSchedulingLatencyMeasurer() *SchedulingLatencyMeasurer {
	return &SchedulingLatencyMeasurer{
		results: make([]time.Duration, 0),
	}
}

// Measure 执行多次测量
func (m *SchedulingLatencyMeasurer) Measure(iterations int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.results = make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		done := make(chan struct{})

		go func() {
			close(done)
		}()

		<-done
		latency := time.Since(start)
		m.results = append(m.results, latency)
	}
}

// MeasureUnderLoad 在负载下测量
func (m *SchedulingLatencyMeasurer) MeasureUnderLoad(iterations, concurrency int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.results = make([]time.Duration, 0, iterations)

	// 创建负载
	stopLoad := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopLoad:
					return
				default:
					// CPU 密集型工作
					sum := 0
					for j := 0; j < 1000; j++ {
						sum += j
					}
					runtime.Gosched()
				}
			}
		}()
	}

	// 测量延迟
	for i := 0; i < iterations; i++ {
		start := time.Now()
		done := make(chan struct{})

		go func() {
			close(done)
		}()

		<-done
		latency := time.Since(start)
		m.results = append(m.results, latency)
	}

	// 停止负载
	close(stopLoad)
	wg.Wait()
}

// GetResults 获取测量结果
func (m *SchedulingLatencyMeasurer) GetResults() LatencyResults {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.results) == 0 {
		return LatencyResults{}
	}

	// 复制并排序
	sorted := make([]time.Duration, len(m.results))
	copy(sorted, m.results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	var total time.Duration
	for _, d := range sorted {
		total += d
	}

	n := len(sorted)
	return LatencyResults{
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

// LatencyResults 延迟测量结果
type LatencyResults struct {
	Count int
	Min   time.Duration
	Max   time.Duration
	Avg   time.Duration
	P50   time.Duration
	P90   time.Duration
	P95   time.Duration
	P99   time.Duration
}

// String 格式化输出
func (r LatencyResults) String() string {
	return fmt.Sprintf(
		"调度延迟统计 (n=%d):\n"+
			"  最小: %v\n"+
			"  最大: %v\n"+
			"  平均: %v\n"+
			"  P50:  %v\n"+
			"  P90:  %v\n"+
			"  P95:  %v\n"+
			"  P99:  %v",
		r.Count, r.Min, r.Max, r.Avg, r.P50, r.P90, r.P95, r.P99,
	)
}

// ==================
// 4. GOMAXPROCS 自动调优器
// ==================

// GOMAXPROCSAutoTuner GOMAXPROCS 自动调优器
// 根据工作负载自动调整 GOMAXPROCS
type GOMAXPROCSAutoTuner struct {
	// 原始 GOMAXPROCS
	original int
	// 最小值
	minProcs int
	// 最大值
	maxProcs int
	// 调整间隔
	interval time.Duration
	// CPU 使用率阈值
	cpuThreshold float64
	// 运行状态
	running atomic.Bool
	// 停止信号
	stopCh chan struct{}
	// 当前值
	current atomic.Int32
}

// NewGOMAXPROCSAutoTuner 创建新的 GOMAXPROCS 自动调优器
func NewGOMAXPROCSAutoTuner(minProcs, maxProcs int, interval time.Duration) *GOMAXPROCSAutoTuner {
	if minProcs < 1 {
		minProcs = 1
	}
	if maxProcs < minProcs {
		maxProcs = runtime.NumCPU()
	}

	tuner := &GOMAXPROCSAutoTuner{
		original:     runtime.GOMAXPROCS(0),
		minProcs:     minProcs,
		maxProcs:     maxProcs,
		interval:     interval,
		cpuThreshold: 0.8, // 80% CPU 使用率阈值
		stopCh:       make(chan struct{}),
	}
	tuner.current.Store(int32(tuner.original))

	return tuner
}

// Start 启动自动调优
func (t *GOMAXPROCSAutoTuner) Start() {
	if !t.running.CompareAndSwap(false, true) {
		return
	}

	go func() {
		ticker := time.NewTicker(t.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.tune()
			case <-t.stopCh:
				return
			}
		}
	}()
}

// Stop 停止自动调优并恢复原始设置
func (t *GOMAXPROCSAutoTuner) Stop() {
	if t.running.CompareAndSwap(true, false) {
		close(t.stopCh)
		runtime.GOMAXPROCS(t.original)
	}
}

func (t *GOMAXPROCSAutoTuner) tune() {
	// 获取当前 Goroutine 数量和 GOMAXPROCS
	numGoroutine := runtime.NumGoroutine()
	currentProcs := int(t.current.Load())

	// 简单的调优策略：
	// - 如果 Goroutine 数量远大于 GOMAXPROCS，增加 GOMAXPROCS
	// - 如果 Goroutine 数量接近 GOMAXPROCS，保持不变
	// - 如果 Goroutine 数量远小于 GOMAXPROCS，减少 GOMAXPROCS

	ratio := float64(numGoroutine) / float64(currentProcs)

	var newProcs int
	switch {
	case ratio > 10 && currentProcs < t.maxProcs:
		// Goroutine 数量远大于 GOMAXPROCS，增加
		newProcs = min(currentProcs+1, t.maxProcs)
	case ratio < 2 && currentProcs > t.minProcs:
		// Goroutine 数量接近 GOMAXPROCS，可以减少
		newProcs = max(currentProcs-1, t.minProcs)
	default:
		// 保持不变
		return
	}

	if newProcs != currentProcs {
		runtime.GOMAXPROCS(newProcs)
		t.current.Store(int32(newProcs))
		fmt.Printf("GOMAXPROCS 自动调整: %d -> %d (Goroutines: %d)\n",
			currentProcs, newProcs, numGoroutine)
	}
}

// GetCurrent 获取当前 GOMAXPROCS
func (t *GOMAXPROCSAutoTuner) GetCurrent() int {
	return int(t.current.Load())
}

// ==================
// 5. 工作负载分析器
// ==================

// WorkloadAnalyzer 工作负载分析器
// 分析当前工作负载特征
type WorkloadAnalyzer struct {
	// 采样间隔
	interval time.Duration
	// 采样数量
	samples int
	// 结果
	results WorkloadAnalysis
	// 互斥锁
	mu sync.Mutex
}

// WorkloadAnalysis 工作负载分析结果
type WorkloadAnalysis struct {
	// Goroutine 统计
	GoroutineCount  int
	GoroutineGrowth float64 // 每秒增长率
	GoroutineChurn  float64 // 创建/销毁频率

	// 调度统计
	SchedulingLatency time.Duration
	ContextSwitches   int64

	// 负载类型判断
	IsCPUBound    bool
	IsIOBound     bool
	IsConcurrency bool

	// 建议
	Recommendations []string
}

// NewWorkloadAnalyzer 创建新的工作负载分析器
func NewWorkloadAnalyzer(interval time.Duration, samples int) *WorkloadAnalyzer {
	return &WorkloadAnalyzer{
		interval: interval,
		samples:  samples,
	}
}

// Analyze 执行工作负载分析
func (a *WorkloadAnalyzer) Analyze() WorkloadAnalysis {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 收集样本
	goroutineSamples := make([]int, a.samples)
	latencySamples := make([]time.Duration, a.samples)

	for i := 0; i < a.samples; i++ {
		goroutineSamples[i] = runtime.NumGoroutine()
		latencySamples[i] = measureSchedulingLatency()
		time.Sleep(a.interval)
	}

	// 分析 Goroutine 增长
	growth := float64(goroutineSamples[a.samples-1]-goroutineSamples[0]) /
		(a.interval.Seconds() * float64(a.samples))

	// 计算 Goroutine 变化频率
	var churn float64
	for i := 1; i < a.samples; i++ {
		diff := goroutineSamples[i] - goroutineSamples[i-1]
		if diff < 0 {
			diff = -diff
		}
		churn += float64(diff)
	}
	churn /= float64(a.samples - 1)

	// 计算平均调度延迟
	var totalLatency time.Duration
	for _, l := range latencySamples {
		totalLatency += l
	}
	avgLatency := totalLatency / time.Duration(a.samples)

	// 判断负载类型
	isCPUBound := avgLatency > 100*time.Microsecond
	isIOBound := churn > 10 && avgLatency < 50*time.Microsecond
	isConcurrency := goroutineSamples[a.samples-1] > runtime.GOMAXPROCS(0)*10

	// 生成建议
	recommendations := a.generateRecommendations(
		goroutineSamples[a.samples-1],
		growth,
		avgLatency,
		isCPUBound,
		isIOBound,
		isConcurrency,
	)

	a.results = WorkloadAnalysis{
		GoroutineCount:    goroutineSamples[a.samples-1],
		GoroutineGrowth:   growth,
		GoroutineChurn:    churn,
		SchedulingLatency: avgLatency,
		IsCPUBound:        isCPUBound,
		IsIOBound:         isIOBound,
		IsConcurrency:     isConcurrency,
		Recommendations:   recommendations,
	}

	return a.results
}

func (a *WorkloadAnalyzer) generateRecommendations(
	count int,
	growth float64,
	latency time.Duration,
	cpuBound, ioBound, highConcurrency bool,
) []string {
	var recommendations []string

	if growth > 10 {
		recommendations = append(recommendations,
			"Goroutine 增长过快，检查是否有 Goroutine 泄漏")
	}

	if cpuBound {
		recommendations = append(recommendations,
			"检测到 CPU 密集型负载：",
			"  - 考虑增加 GOMAXPROCS",
			"  - 使用 worker pool 限制并发",
			"  - 检查是否有不必要的计算")
	}

	if ioBound {
		recommendations = append(recommendations,
			"检测到 I/O 密集型负载：",
			"  - 可以使用更多 Goroutine",
			"  - 考虑使用连接池",
			"  - 检查 I/O 超时设置")
	}

	if highConcurrency {
		recommendations = append(recommendations,
			"高并发负载：",
			"  - 使用 sync.Pool 减少分配",
			"  - 考虑使用 worker pool",
			"  - 监控调度延迟")
	}

	if latency > time.Millisecond {
		recommendations = append(recommendations,
			"调度延迟过高：",
			"  - 检查是否有长时间运行的 Goroutine",
			"  - 考虑增加 GOMAXPROCS",
			"  - 使用 runtime.Gosched() 主动让出")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "工作负载正常，无需调优")
	}

	return recommendations
}

// ==================
// 6. 便捷函数
// ==================

// GetSchedulerInfo 获取调度器信息
func GetSchedulerInfo() SchedulerInfo {
	return SchedulerInfo{
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		NumGoroutine: runtime.NumGoroutine(),
		NumCgoCall:   runtime.NumCgoCall(),
	}
}

// SchedulerInfo 调度器信息
type SchedulerInfo struct {
	NumCPU       int
	GOMAXPROCS   int
	NumGoroutine int
	NumCgoCall   int64
}

// String 格式化输出
func (s SchedulerInfo) String() string {
	return fmt.Sprintf(
		"调度器信息:\n"+
			"  CPU 核心数: %d\n"+
			"  GOMAXPROCS: %d\n"+
			"  Goroutine 数: %d\n"+
			"  CGO 调用数: %d",
		s.NumCPU, s.GOMAXPROCS, s.NumGoroutine, s.NumCgoCall,
	)
}

// PrintSchedulerInfo 打印调度器信息
func PrintSchedulerInfo() {
	info := GetSchedulerInfo()
	fmt.Println(info.String())
}

// WatchGoroutines 监控 Goroutine 数量变化
func WatchGoroutines(interval time.Duration, stopCh <-chan struct{}) {
	var lastCount int

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			current := runtime.NumGoroutine()
			diff := current - lastCount

			if diff != 0 {
				sign := "+"
				if diff < 0 {
					sign = ""
				}
				fmt.Printf("[Goroutine] 当前: %d (%s%d)\n", current, sign, diff)
			}

			lastCount = current

		case <-stopCh:
			return
		}
	}
}

// MeasureLatency 测量调度延迟（便捷函数）
func MeasureLatency(iterations int) LatencyResults {
	measurer := NewSchedulingLatencyMeasurer()
	measurer.Measure(iterations)
	return measurer.GetResults()
}

// DetectGoroutineLeak 检测 Goroutine 泄漏（便捷函数）
func DetectGoroutineLeak(baseline int, threshold float64) (leaked bool, current int) {
	current = runtime.NumGoroutine()
	if baseline <= 0 {
		return false, current
	}
	increase := float64(current-baseline) / float64(baseline)
	leaked = increase > threshold
	return
}
