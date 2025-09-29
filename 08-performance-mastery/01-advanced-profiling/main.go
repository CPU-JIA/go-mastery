/*
=== Go性能掌控：高级性能分析技术 ===

本模块深入高级性能分析技术的实现和应用，探索：
1. 自定义采样器和热点分析
2. 实时性能监控和报警系统
3. 微基准测试的高级技巧
4. 性能回归检测机制
5. 调用图分析和火焰图生成
6. 代码覆盖率和性能相关性分析
7. 内存泄漏检测和分析
8. 锁竞争分析和优化
9. 网络I/O性能分析
10. 自动化性能优化建议

学习目标：
- 掌握企业级性能分析工具的构建
- 理解性能瓶颈的深度诊断方法
- 学会构建自动化性能监控体系
- 掌握性能优化的科学方法论
*/

package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"go-mastery/common/security"
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

// ==================
// 1. 高级采样器框架
// ==================

// SampleType 采样类型
type SampleType int

const (
	CPUSample SampleType = iota
	MemorySample
	GoroutineSample
	BlockSample
	MutexSample
	CustomSample
)

func (s SampleType) String() string {
	names := []string{"CPU", "Memory", "Goroutine", "Block", "Mutex", "Custom"}
	if int(s) < len(names) {
		return names[s]
	}
	return "Unknown"
}

// Sample 性能采样数据
type Sample struct {
	Type      SampleType
	Timestamp time.Time
	Value     float64
	Metadata  map[string]interface{}
	Location  StackTrace
}

// StackTrace 调用栈信息
type StackTrace struct {
	Frames []StackFrame
	Hash   uint64
}

// StackFrame 栈帧信息
type StackFrame struct {
	Function string
	File     string
	Line     int
	PC       uintptr
}

// AdvancedProfiler 高级性能分析器
type AdvancedProfiler struct {
	samples    []Sample
	sampleRate time.Duration
	running    bool
	stopCh     chan struct{}
	wg         sync.WaitGroup
	mutex      sync.RWMutex
	samplers   map[SampleType]Sampler
	hotspots   map[string]*Hotspot
	alerts     []PerformanceAlert
	thresholds map[string]float64
	collector  *MetricsCollector
}

// Sampler 采样器接口
type Sampler interface {
	Sample() (Sample, error)
	Configure(config map[string]interface{}) error
	Name() string
}

// Hotspot 性能热点
type Hotspot struct {
	Function     string
	File         string
	Line         int
	SampleCount  int64
	TotalValue   float64
	AverageValue float64
	LastSeen     time.Time
	Trend        TrendDirection
}

// TrendDirection 趋势方向
type TrendDirection int

const (
	TrendUp TrendDirection = iota
	TrendDown
	TrendStable
)

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	ID        string
	Type      string
	Message   string
	Severity  AlertSeverity
	Timestamp time.Time
	Value     float64
	Threshold float64
	Function  string
	Resolved  bool
	Metadata  map[string]interface{}
}

// AlertSeverity 告警严重程度
type AlertSeverity int

const (
	AlertInfo AlertSeverity = iota
	AlertWarning
	AlertCritical
)

func NewAdvancedProfiler() *AdvancedProfiler {
	profiler := &AdvancedProfiler{
		samples:    make([]Sample, 0),
		sampleRate: 100 * time.Millisecond,
		stopCh:     make(chan struct{}),
		samplers:   make(map[SampleType]Sampler),
		hotspots:   make(map[string]*Hotspot),
		alerts:     make([]PerformanceAlert, 0),
		thresholds: make(map[string]float64),
		collector:  NewMetricsCollector(),
	}

	// 注册默认采样器
	profiler.RegisterSampler(CPUSample, &CPUSampler{})
	profiler.RegisterSampler(MemorySample, &MemorySampler{})
	profiler.RegisterSampler(GoroutineSample, &GoroutineSampler{})

	// 设置默认阈值
	profiler.SetThreshold("cpu_usage", 80.0)
	profiler.SetThreshold("memory_usage", 85.0)
	profiler.SetThreshold("goroutine_count", 10000)
	profiler.SetThreshold("gc_pause", 10.0) // 10ms

	return profiler
}

func (p *AdvancedProfiler) RegisterSampler(sampleType SampleType, sampler Sampler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.samplers[sampleType] = sampler
}

func (p *AdvancedProfiler) SetThreshold(metric string, threshold float64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.thresholds[metric] = threshold
}

func (p *AdvancedProfiler) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return fmt.Errorf("profiler already running")
	}

	p.running = true

	// 启动采样器
	for sampleType, sampler := range p.samplers {
		p.wg.Add(1)
		go p.runSampler(sampleType, sampler)
	}

	// 启动热点分析
	p.wg.Add(1)
	go p.runHotspotAnalyzer()

	// 启动告警检测
	p.wg.Add(1)
	go p.runAlertDetector()

	fmt.Println("高级性能分析器已启动")
	return nil
}

func (p *AdvancedProfiler) Stop() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return
	}

	p.running = false
	close(p.stopCh)
	p.wg.Wait()

	fmt.Println("高级性能分析器已停止")
}

func (p *AdvancedProfiler) runSampler(sampleType SampleType, sampler Sampler) {
	defer p.wg.Done()

	ticker := time.NewTicker(p.sampleRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sample, err := sampler.Sample()
			if err != nil {
				log.Printf("采样器 %s 错误: %v", sampler.Name(), err)
				continue
			}

			p.mutex.Lock()
			p.samples = append(p.samples, sample)
			// 保持最近10000个样本
			if len(p.samples) > 10000 {
				p.samples = p.samples[1000:]
			}
			p.mutex.Unlock()

			// 更新热点数据
			p.updateHotspot(sample)

		case <-p.stopCh:
			return
		}
	}
}

func (p *AdvancedProfiler) updateHotspot(sample Sample) {
	if len(sample.Location.Frames) == 0 {
		return
	}

	frame := sample.Location.Frames[0]
	key := fmt.Sprintf("%s:%s:%d", frame.Function, frame.File, frame.Line)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	hotspot, exists := p.hotspots[key]
	if !exists {
		hotspot = &Hotspot{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		}
		p.hotspots[key] = hotspot
	}

	hotspot.SampleCount++
	hotspot.TotalValue += sample.Value
	hotspot.AverageValue = hotspot.TotalValue / float64(hotspot.SampleCount)
	hotspot.LastSeen = sample.Timestamp
}

func (p *AdvancedProfiler) runHotspotAnalyzer() {
	defer p.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.analyzeHotspots()

		case <-p.stopCh:
			return
		}
	}
}

func (p *AdvancedProfiler) analyzeHotspots() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 分析热点趋势
	for _, hotspot := range p.hotspots {
		// 简化的趋势分析
		if hotspot.SampleCount > 100 {
			if hotspot.AverageValue > hotspot.TotalValue/float64(hotspot.SampleCount)*1.1 {
				hotspot.Trend = TrendUp
			} else if hotspot.AverageValue < hotspot.TotalValue/float64(hotspot.SampleCount)*0.9 {
				hotspot.Trend = TrendDown
			} else {
				hotspot.Trend = TrendStable
			}
		}
	}
}

func (p *AdvancedProfiler) runAlertDetector() {
	defer p.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.checkAlerts()

		case <-p.stopCh:
			return
		}
	}
}

func (p *AdvancedProfiler) checkAlerts() {
	p.mutex.RLock()
	samples := make([]Sample, len(p.samples))
	copy(samples, p.samples)
	thresholds := make(map[string]float64)
	for k, v := range p.thresholds {
		thresholds[k] = v
	}
	p.mutex.RUnlock()

	// 检查最近的样本
	if len(samples) < 10 {
		return
	}

	recentSamples := samples[len(samples)-10:]

	// CPU使用率检查
	cpuSum := 0.0
	cpuCount := 0
	for _, sample := range recentSamples {
		if sample.Type == CPUSample {
			cpuSum += sample.Value
			cpuCount++
		}
	}

	if cpuCount > 0 {
		avgCPU := cpuSum / float64(cpuCount)
		if threshold, exists := thresholds["cpu_usage"]; exists && avgCPU > threshold {
			alert := PerformanceAlert{
				ID:        fmt.Sprintf("cpu_%d", time.Now().Unix()),
				Type:      "CPU_HIGH",
				Message:   fmt.Sprintf("CPU使用率过高: %.2f%%", avgCPU),
				Severity:  AlertWarning,
				Timestamp: time.Now(),
				Value:     avgCPU,
				Threshold: threshold,
			}
			p.addAlert(alert)
		}
	}
}

func (p *AdvancedProfiler) addAlert(alert PerformanceAlert) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.alerts = append(p.alerts, alert)

	// 保持最近100个告警
	if len(p.alerts) > 100 {
		p.alerts = p.alerts[10:]
	}

	fmt.Printf("🚨 性能告警: %s\n", alert.Message)
}

func (p *AdvancedProfiler) GetTopHotspots(limit int) []*Hotspot {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	hotspots := make([]*Hotspot, 0, len(p.hotspots))
	for _, hotspot := range p.hotspots {
		hotspots = append(hotspots, hotspot)
	}

	// 按样本数量排序
	sort.Slice(hotspots, func(i, j int) bool {
		return hotspots[i].SampleCount > hotspots[j].SampleCount
	})

	if limit > len(hotspots) {
		limit = len(hotspots)
	}

	return hotspots[:limit]
}

func (p *AdvancedProfiler) GetRecentAlerts(limit int) []PerformanceAlert {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if limit > len(p.alerts) {
		limit = len(p.alerts)
	}

	alerts := make([]PerformanceAlert, limit)
	copy(alerts, p.alerts[len(p.alerts)-limit:])

	return alerts
}

// ==================
// 2. 具体采样器实现
// ==================

// CPUSampler CPU采样器
type CPUSampler struct {
	lastCPUTime time.Duration
	lastSample  time.Time
}

func (s *CPUSampler) Name() string {
	return "CPU Sampler"
}

func (s *CPUSampler) Configure(config map[string]interface{}) error {
	return nil
}

func (s *CPUSampler) Sample() (Sample, error) {
	now := time.Now()

	// 获取当前CPU时间（简化版本）
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	cpuUsage := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10 // 简化计算

	sample := Sample{
		Type:      CPUSample,
		Timestamp: now,
		Value:     cpuUsage,
		Metadata:  map[string]interface{}{"goroutines": runtime.NumGoroutine()},
		Location:  s.getCurrentStackTrace(),
	}

	return sample, nil
}

func (s *CPUSampler) getCurrentStackTrace() StackTrace {
	pc := make([]uintptr, 32)
	n := runtime.Callers(2, pc)

	frames := make([]StackFrame, 0, n)
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pc[i])
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc[i])
		frame := StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
			PC:       pc[i],
		}
		frames = append(frames, frame)
	}

	return StackTrace{Frames: frames}
}

// MemorySampler 内存采样器
type MemorySampler struct{}

func (s *MemorySampler) Name() string {
	return "Memory Sampler"
}

func (s *MemorySampler) Configure(config map[string]interface{}) error {
	return nil
}

func (s *MemorySampler) Sample() (Sample, error) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	memUsagePercent := float64(ms.HeapInuse) / float64(ms.HeapSys) * 100

	sample := Sample{
		Type:      MemorySample,
		Timestamp: time.Now(),
		Value:     memUsagePercent,
		Metadata: map[string]interface{}{
			"heap_inuse":   ms.HeapInuse,
			"heap_sys":     ms.HeapSys,
			"heap_objects": ms.HeapObjects,
			"gc_runs":      ms.NumGC,
		},
		Location: StackTrace{}, // 内存采样通常不需要调用栈
	}

	return sample, nil
}

// GoroutineSampler Goroutine采样器
type GoroutineSampler struct{}

func (s *GoroutineSampler) Name() string {
	return "Goroutine Sampler"
}

func (s *GoroutineSampler) Configure(config map[string]interface{}) error {
	return nil
}

func (s *GoroutineSampler) Sample() (Sample, error) {
	numGoroutines := float64(runtime.NumGoroutine())

	sample := Sample{
		Type:      GoroutineSample,
		Timestamp: time.Now(),
		Value:     numGoroutines,
		Metadata: map[string]interface{}{
			"num_cpu":    runtime.NumCPU(),
			"gomaxprocs": runtime.GOMAXPROCS(0),
		},
		Location: StackTrace{},
	}

	return sample, nil
}

// ==================
// 3. 指标收集器
// ==================

// MetricsCollector 指标收集器
type MetricsCollector struct {
	metrics map[string]*Metric
	mutex   sync.RWMutex
}

// Metric 指标
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
	History   []MetricValue
}

// MetricType 指标类型
type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
	Summary
)

// MetricValue 指标值历史
type MetricValue struct {
	Value     float64
	Timestamp time.Time
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metric),
	}
}

func (mc *MetricsCollector) RecordCounter(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.getMetricKey(name, labels)
	metric, exists := mc.metrics[key]
	if !exists {
		metric = &Metric{
			Name:    name,
			Type:    Counter,
			Labels:  labels,
			History: make([]MetricValue, 0),
		}
		mc.metrics[key] = metric
	}

	metric.Value += value
	metric.Timestamp = time.Now()
	metric.History = append(metric.History, MetricValue{
		Value:     metric.Value,
		Timestamp: metric.Timestamp,
	})

	// 保持最近1000个历史值
	if len(metric.History) > 1000 {
		metric.History = metric.History[100:]
	}
}

func (mc *MetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.getMetricKey(name, labels)
	metric, exists := mc.metrics[key]
	if !exists {
		metric = &Metric{
			Name:    name,
			Type:    Gauge,
			Labels:  labels,
			History: make([]MetricValue, 0),
		}
		mc.metrics[key] = metric
	}

	metric.Value = value
	metric.Timestamp = time.Now()
	metric.History = append(metric.History, MetricValue{
		Value:     value,
		Timestamp: metric.Timestamp,
	})

	if len(metric.History) > 1000 {
		metric.History = metric.History[100:]
	}
}

func (mc *MetricsCollector) getMetricKey(name string, labels map[string]string) string {
	var parts []string
	parts = append(parts, name)

	// 按键排序确保一致性
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	return strings.Join(parts, "|")
}

func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make(map[string]*Metric)
	for k, v := range mc.metrics {
		result[k] = v
	}

	return result
}

// ==================
// 4. 微基准测试框架
// ==================

// BenchmarkSuite 基准测试套件
type BenchmarkSuite struct {
	benchmarks []BenchmarkFunc
	results    []BenchmarkResult
	config     BenchmarkConfig
}

// BenchmarkFunc 基准测试函数
type BenchmarkFunc struct {
	Name     string
	Function func(*BenchmarkContext)
	Setup    func() interface{}
	Teardown func(interface{})
}

// BenchmarkConfig 基准测试配置
type BenchmarkConfig struct {
	Duration   time.Duration
	Iterations int
	WarmupRuns int
	CPUProfile bool
	MemProfile bool
	Parallel   bool
}

// BenchmarkContext 基准测试上下文
type BenchmarkContext struct {
	N         int
	startTime time.Time
	bytes     int64
	netAllocs int64
	netBytes  int64
}

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Name        string
	Iterations  int
	Duration    time.Duration
	NsPerOp     int64
	BytesPerOp  int64
	AllocsPerOp int64
	MBPerSec    float64
	MemAllocs   int64
	MemBytes    int64
	CPUSamples  []CPUProfileSample
	Confidence  float64
	Variability float64
}

// CPUProfileSample CPU采样数据
type CPUProfileSample struct {
	Function string
	File     string
	Line     int
	Samples  int64
	Percent  float64
}

func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		benchmarks: make([]BenchmarkFunc, 0),
		results:    make([]BenchmarkResult, 0),
		config: BenchmarkConfig{
			Duration:   time.Second,
			WarmupRuns: 3,
			Parallel:   false,
		},
	}
}

func (bs *BenchmarkSuite) AddBenchmark(name string, fn func(*BenchmarkContext)) {
	benchmark := BenchmarkFunc{
		Name:     name,
		Function: fn,
	}
	bs.benchmarks = append(bs.benchmarks, benchmark)
}

func (bs *BenchmarkSuite) SetConfig(config BenchmarkConfig) {
	bs.config = config
}

func (bs *BenchmarkSuite) Run() []BenchmarkResult {
	results := make([]BenchmarkResult, 0, len(bs.benchmarks))

	for _, benchmark := range bs.benchmarks {
		fmt.Printf("运行基准测试: %s\n", benchmark.Name)
		result := bs.runSingleBenchmark(benchmark)
		results = append(results, result)

		fmt.Printf("  %s: %d次迭代, %.2f ns/op, %.2f MB/s\n",
			benchmark.Name, result.Iterations, float64(result.NsPerOp), result.MBPerSec)
	}

	bs.results = results
	return results
}

func (bs *BenchmarkSuite) runSingleBenchmark(benchmark BenchmarkFunc) BenchmarkResult {
	// 预热运行
	for i := 0; i < bs.config.WarmupRuns; i++ {
		ctx := &BenchmarkContext{N: 1000}
		benchmark.Function(ctx)
	}

	// 确定迭代次数
	iterations := bs.determineIterations(benchmark)

	// 执行基准测试
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	startTime := time.Now()
	ctx := &BenchmarkContext{
		N:         iterations,
		startTime: startTime,
	}

	benchmark.Function(ctx)

	duration := time.Since(startTime)
	runtime.ReadMemStats(&memAfter)

	// 计算结果
	result := BenchmarkResult{
		Name:       benchmark.Name,
		Iterations: iterations,
		Duration:   duration,
		NsPerOp:    duration.Nanoseconds() / int64(iterations),
		MemAllocs:  security.MustSafeUint64ToInt64(memAfter.Mallocs - memBefore.Mallocs),
		MemBytes:   security.MustSafeUint64ToInt64(memAfter.TotalAlloc - memBefore.TotalAlloc),
	}

	if ctx.bytes > 0 {
		result.BytesPerOp = ctx.bytes / int64(iterations)
		result.MBPerSec = float64(ctx.bytes) / duration.Seconds() / 1024 / 1024
	}

	if result.MemAllocs > 0 {
		result.AllocsPerOp = result.MemAllocs / int64(iterations)
	}

	return result
}

func (bs *BenchmarkSuite) determineIterations(benchmark BenchmarkFunc) int {
	// 自动确定合适的迭代次数
	targetDuration := bs.config.Duration

	// 先运行少量迭代估算性能
	testN := 100
	ctx := &BenchmarkContext{N: testN}

	start := time.Now()
	benchmark.Function(ctx)
	elapsed := time.Since(start)

	if elapsed == 0 {
		return 1000000 // 默认值
	}

	// 根据目标时间计算迭代次数
	iterations := int(targetDuration.Nanoseconds() / elapsed.Nanoseconds() * int64(testN))

	// 限制范围
	if iterations < 1 {
		iterations = 1
	}
	if iterations > 1000000 {
		iterations = 1000000
	}

	return iterations
}

func (ctx *BenchmarkContext) SetBytes(n int64) {
	ctx.bytes = n
}

func (ctx *BenchmarkContext) ResetTimer() {
	ctx.startTime = time.Now()
}

// ==================
// 5. 性能回归检测
// ==================

// RegressionDetector 性能回归检测器
type RegressionDetector struct {
	baselines map[string]Baseline
	mutex     sync.RWMutex
	threshold float64 // 回归阈值百分比
}

// Baseline 性能基线
type Baseline struct {
	Name         string
	Value        float64
	StandardDev  float64
	SampleCount  int
	LastUpdated  time.Time
	Measurements []float64
}

// RegressionReport 回归报告
type RegressionReport struct {
	TestName          string
	BaselineValue     float64
	CurrentValue      float64
	RegressionPercent float64
	IsRegression      bool
	Confidence        float64
	Timestamp         time.Time
}

func NewRegressionDetector(threshold float64) *RegressionDetector {
	return &RegressionDetector{
		baselines: make(map[string]Baseline),
		threshold: threshold,
	}
}

func (rd *RegressionDetector) UpdateBaseline(name string, value float64) {
	rd.mutex.Lock()
	defer rd.mutex.Unlock()

	baseline, exists := rd.baselines[name]
	if !exists {
		baseline = Baseline{
			Name:         name,
			Measurements: make([]float64, 0),
		}
	}

	baseline.Measurements = append(baseline.Measurements, value)

	// 保持最近100个测量值
	if len(baseline.Measurements) > 100 {
		baseline.Measurements = baseline.Measurements[1:]
	}

	// 计算统计值
	sum := 0.0
	for _, v := range baseline.Measurements {
		sum += v
	}
	baseline.Value = sum / float64(len(baseline.Measurements))
	baseline.SampleCount = len(baseline.Measurements)
	baseline.LastUpdated = time.Now()

	// 计算标准差
	if len(baseline.Measurements) > 1 {
		variance := 0.0
		for _, v := range baseline.Measurements {
			variance += math.Pow(v-baseline.Value, 2)
		}
		baseline.StandardDev = math.Sqrt(variance / float64(len(baseline.Measurements)-1))
	}

	rd.baselines[name] = baseline
}

func (rd *RegressionDetector) CheckRegression(name string, currentValue float64) RegressionReport {
	rd.mutex.RLock()
	baseline, exists := rd.baselines[name]
	rd.mutex.RUnlock()

	report := RegressionReport{
		TestName:     name,
		CurrentValue: currentValue,
		Timestamp:    time.Now(),
	}

	if !exists || baseline.SampleCount < 5 {
		report.IsRegression = false
		report.Confidence = 0.0
		return report
	}

	report.BaselineValue = baseline.Value
	report.RegressionPercent = (currentValue - baseline.Value) / baseline.Value * 100

	// 使用统计显著性检验
	if baseline.StandardDev > 0 {
		zScore := math.Abs(currentValue-baseline.Value) / baseline.StandardDev
		report.Confidence = rd.calculateConfidence(zScore)
	}

	// 判断是否为回归
	if report.RegressionPercent > rd.threshold && report.Confidence > 0.95 {
		report.IsRegression = true
	}

	return report
}

func (rd *RegressionDetector) calculateConfidence(zScore float64) float64 {
	// 简化的置信度计算（基于Z分数）
	if zScore < 1.0 {
		return 0.68
	} else if zScore < 2.0 {
		return 0.95
	} else if zScore < 3.0 {
		return 0.99
	}
	return 0.999
}

func (rd *RegressionDetector) GetBaselines() map[string]Baseline {
	rd.mutex.RLock()
	defer rd.mutex.RUnlock()

	result := make(map[string]Baseline)
	for k, v := range rd.baselines {
		result[k] = v
	}

	return result
}

// ==================
// 6. 内存泄漏检测
// ==================

// LeakDetector 内存泄漏检测器
type LeakDetector struct {
	snapshots    []MemorySnapshot
	allocHistory map[string]*AllocationHistory
	mutex        sync.RWMutex
}

// MemorySnapshot 内存快照
type MemorySnapshot struct {
	Timestamp   time.Time
	HeapInuse   uint64
	HeapObjects uint64
	StackInuse  uint64
	Allocs      uint64
	Frees       uint64
	GCRuns      uint32
	Goroutines  int
}

// AllocationHistory 分配历史
type AllocationHistory struct {
	Function      string
	AllocCount    int64
	TotalSize     int64
	PeakSize      int64
	LastAllocTime time.Time
	Trend         TrendDirection
}

// LeakReport 泄漏报告
type LeakReport struct {
	Detected        bool
	LeakRate        float64 // MB/hour
	SuspiciousFuncs []SuspiciousFunction
	Recommendations []string
	Timestamp       time.Time
}

// SuspiciousFunction 可疑函数
type SuspiciousFunction struct {
	Name       string
	AllocRate  float64
	Size       int64
	Confidence float64
}

func NewLeakDetector() *LeakDetector {
	return &LeakDetector{
		snapshots:    make([]MemorySnapshot, 0),
		allocHistory: make(map[string]*AllocationHistory),
	}
}

func (ld *LeakDetector) TakeSnapshot() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	snapshot := MemorySnapshot{
		Timestamp:   time.Now(),
		HeapInuse:   ms.HeapInuse,
		HeapObjects: ms.HeapObjects,
		StackInuse:  ms.StackInuse,
		Allocs:      ms.Mallocs,
		Frees:       ms.Frees,
		GCRuns:      ms.NumGC,
		Goroutines:  runtime.NumGoroutine(),
	}

	ld.mutex.Lock()
	ld.snapshots = append(ld.snapshots, snapshot)

	// 保持最近1000个快照
	if len(ld.snapshots) > 1000 {
		ld.snapshots = ld.snapshots[100:]
	}
	ld.mutex.Unlock()
}

func (ld *LeakDetector) AnalyzeLeaks() LeakReport {
	ld.mutex.RLock()
	snapshots := make([]MemorySnapshot, len(ld.snapshots))
	copy(snapshots, ld.snapshots)
	ld.mutex.RUnlock()

	report := LeakReport{
		Timestamp:       time.Now(),
		SuspiciousFuncs: make([]SuspiciousFunction, 0),
		Recommendations: make([]string, 0),
	}

	if len(snapshots) < 10 {
		return report
	}

	// 分析内存增长趋势
	timeSpan := snapshots[len(snapshots)-1].Timestamp.Sub(snapshots[0].Timestamp)
	memoryGrowth := security.MustSafeUint64ToInt64(snapshots[len(snapshots)-1].HeapInuse) - security.MustSafeUint64ToInt64(snapshots[0].HeapInuse)

	if timeSpan.Hours() > 0 {
		report.LeakRate = float64(memoryGrowth) / timeSpan.Hours() / 1024 / 1024 // MB/hour
	}

	// 检测是否存在泄漏
	if report.LeakRate > 10.0 { // 超过10MB/hour认为可能有泄漏
		report.Detected = true
		report.Recommendations = append(report.Recommendations,
			"检测到持续的内存增长，建议进行详细的内存分析",
			"使用go tool pprof分析heap profile",
			"检查是否有goroutine泄漏",
			"验证defer语句和资源清理代码",
		)
	}

	// 分析goroutine数量趋势
	goroutineGrowth := snapshots[len(snapshots)-1].Goroutines - snapshots[0].Goroutines
	if goroutineGrowth > 100 {
		report.Detected = true
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("检测到goroutine数量增长 %d，可能存在goroutine泄漏", goroutineGrowth),
		)
	}

	return report
}

// ==================
// 7. 主演示函数
// ==================

func demonstrateAdvancedProfiling() {
	fmt.Println("=== Go高级性能分析技术演示 ===")

	// 1. 启动高级性能分析器
	fmt.Println("\n1. 启动高级性能分析器")
	profiler := NewAdvancedProfiler()
	profiler.Start()
	defer profiler.Stop()

	// 2. 启动HTTP pprof服务器
	go func() {
		log.Println("pprof server starting on :6060")
		server := &http.Server{
			Addr:         "localhost:6060",
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Println(server.ListenAndServe())
	}()

	// 3. 创建一些性能负载
	fmt.Println("\n2. 生成性能测试负载")
	generatePerformanceLoad()

	// 等待采集数据
	time.Sleep(3 * time.Second)

	// 4. 分析性能热点
	fmt.Println("\n3. 性能热点分析")
	hotspots := profiler.GetTopHotspots(5)
	for i, hotspot := range hotspots {
		fmt.Printf("热点 %d: %s (%s:%d)\n", i+1, hotspot.Function, hotspot.File, hotspot.Line)
		fmt.Printf("  样本数: %d, 平均值: %.2f, 趋势: %v\n",
			hotspot.SampleCount, hotspot.AverageValue, hotspot.Trend)
	}

	// 5. 检查性能告警
	fmt.Println("\n4. 性能告警检查")
	alerts := profiler.GetRecentAlerts(5)
	for _, alert := range alerts {
		fmt.Printf("告警: %s (严重程度: %d)\n", alert.Message, alert.Severity)
	}

	// 6. 微基准测试
	fmt.Println("\n5. 微基准测试演示")
	suite := NewBenchmarkSuite()

	// 添加测试用例
	suite.AddBenchmark("StringConcat", benchmarkStringConcat)
	suite.AddBenchmark("MapAccess", benchmarkMapAccess)
	suite.AddBenchmark("SliceAlloc", benchmarkSliceAlloc)

	results := suite.Run()

	// 7. 性能回归检测
	fmt.Println("\n6. 性能回归检测")
	detector := NewRegressionDetector(10.0) // 10%阈值

	for _, result := range results {
		// 更新基线（模拟历史数据）
		for i := 0; i < 10; i++ {
			baselineValue := float64(result.NsPerOp) * (0.9 + secureRandomFloat64()*0.2)
			detector.UpdateBaseline(result.Name, baselineValue)
		}

		// 检查当前结果是否回归
		report := detector.CheckRegression(result.Name, float64(result.NsPerOp))
		if report.IsRegression {
			fmt.Printf("🚨 性能回归: %s, 下降 %.2f%%, 置信度 %.2f\n",
				report.TestName, report.RegressionPercent, report.Confidence)
		} else {
			fmt.Printf("✅ 性能正常: %s\n", report.TestName)
		}
	}

	// 8. 内存泄漏检测
	fmt.Println("\n7. 内存泄漏检测")
	leakDetector := NewLeakDetector()

	// 模拟内存分配和泄漏检测
	for i := 0; i < 10; i++ {
		leakDetector.TakeSnapshot()
		simulateMemoryAllocation()
		time.Sleep(100 * time.Millisecond)
	}

	leakReport := leakDetector.AnalyzeLeaks()
	if leakReport.Detected {
		fmt.Printf("🚨 检测到潜在内存泄漏: %.2f MB/hour\n", leakReport.LeakRate)
		for _, rec := range leakReport.Recommendations {
			fmt.Printf("  建议: %s\n", rec)
		}
	} else {
		fmt.Printf("✅ 未检测到内存泄漏\n")
	}

	// 9. 指标收集演示
	fmt.Println("\n8. 指标收集演示")
	collector := profiler.collector

	// 记录一些示例指标
	collector.RecordCounter("http_requests_total", 100, map[string]string{"method": "GET"})
	collector.RecordGauge("memory_usage_bytes", 1024*1024*256, map[string]string{"type": "heap"})

	metrics := collector.GetMetrics()
	fmt.Printf("收集到 %d 个指标\n", len(metrics))
	for _, metric := range metrics {
		fmt.Printf("  %s: %.2f (%v)\n", metric.Name, metric.Value, metric.Type)
	}
}

func generatePerformanceLoad() {
	// CPU密集型任务
	go func() {
		for i := 0; i < 1000000; i++ {
			sum := 0
			for j := 0; j < 1000; j++ {
				sum += j * j
			}
		}
	}()

	// 内存分配任务
	go func() {
		data := make([][]byte, 0)
		for i := 0; i < 1000; i++ {
			chunk := make([]byte, 1024*i)
			data = append(data, chunk)
		}
	}()

	// Goroutine密集型任务
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * 10)
		}()
	}
	wg.Wait()
}

func benchmarkStringConcat(ctx *BenchmarkContext) {
	str := ""
	for i := 0; i < ctx.N; i++ {
		str += "test"
	}
}

func benchmarkMapAccess(ctx *BenchmarkContext) {
	m := make(map[int]string)
	for i := 0; i < 1000; i++ {
		m[i] = fmt.Sprintf("value%d", i)
	}

	for i := 0; i < ctx.N; i++ {
		_ = m[i%1000]
	}
}

func benchmarkSliceAlloc(ctx *BenchmarkContext) {
	for i := 0; i < ctx.N; i++ {
		slice := make([]int, 1000)
		for j := range slice {
			slice[j] = j
		}
	}
}

func simulateMemoryAllocation() {
	// 模拟一些内存分配
	data := make([][]byte, 100)
	for i := range data {
		data[i] = make([]byte, 1024)
	}
}

func main() {
	demonstrateAdvancedProfiling()

	fmt.Println("\n=== Go高级性能分析技术演示完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 自定义采样器：构建专业的性能监控体系")
	fmt.Println("2. 热点分析：自动识别性能瓶颈和趋势")
	fmt.Println("3. 实时告警：基于阈值的智能性能监控")
	fmt.Println("4. 微基准测试：精确的性能测量和比较")
	fmt.Println("5. 回归检测：统计学方法检测性能退化")
	fmt.Println("6. 泄漏检测：自动化内存和goroutine泄漏分析")
	fmt.Println("7. 指标收集：结构化的性能数据管理")

	fmt.Println("\n高级特性:")
	fmt.Println("- 多维度性能采样和分析")
	fmt.Println("- 统计学基础的性能评估")
	fmt.Println("- 自适应阈值和智能告警")
	fmt.Println("- 企业级性能监控架构")
	fmt.Println("- 生产环境性能诊断工具")
}

/*
=== 练习题 ===

1. 高级性能分析：
   - 实现更精确的CPU使用率计算
   - 添加网络I/O性能监控
   - 实现分布式性能数据聚合
   - 创建性能数据可视化界面

2. 微基准测试优化：
   - 实现统计显著性检验
   - 添加多维度性能比较
   - 创建基准测试套件管理
   - 实现持续性能监控

3. 内存分析深化：
   - 实现详细的内存分配追踪
   - 添加内存使用模式分析
   - 创建内存优化建议系统
   - 实现内存使用预测

4. 企业级应用：
   - 集成Prometheus/Grafana监控
   - 实现性能数据持久化
   - 创建性能报告生成系统
   - 建立性能SLA监控体系

运行命令：
go run main.go

分析命令：
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap
go tool trace http://localhost:6060/debug/pprof/trace

重要概念：
- 性能采样：系统化的性能数据收集
- 热点分析：基于统计的性能瓶颈识别
- 回归检测：自动化的性能退化发现
- 泄漏检测：内存和资源泄漏的早期发现
- 基准测试：科学的性能测量方法
*/
