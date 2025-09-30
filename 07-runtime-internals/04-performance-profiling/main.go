/*
=== Go运行时内核：性能分析与诊断 ===

本模块深入Go语言性能分析工具的实现原理，探索：
1. pprof工具的内核机制
2. CPU性能分析(CPU Profiling)
3. 内存分析(Memory Profiling)
4. 阻塞分析(Block Profiling)
5. Mutex争用分析(Mutex Profiling)
6. Goroutine分析(Goroutine Profiling)
7. 执行跟踪(Execution Tracing)
8. GC分析器深度使用
9. 自定义性能指标收集

学习目标：
- 掌握pprof的高级使用技巧
- 理解各种性能分析的原理
- 学会性能瓶颈的诊断方法
- 掌握自定义性能监控
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sort"
	"strings"
	"sync"
	"time"

	"go-mastery/common/security"
)

// 安全随机数生成函数
func secureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano() % int64(max)
		if fallback > int64(^uint(0)>>1) {
			fallback = fallback % int64(^uint(0)>>1)
		}
		return int(fallback)
	}
	// G115安全修复：检查int64到int的安全转换
	result := n.Int64()
	if result > int64(^uint(0)>>1) {
		result = result % int64(max)
	}
	return int(result)
}

func secureRandomFloat32() float32 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		// 安全fallback：使用时间戳
		return float32(time.Now().UnixNano()%1000) / 1000.0
	}
	return float32(n.Int64()) / float32(1<<24)
}

func secureRandomFloat64() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		// 安全fallback：使用时间戳
		return float64(time.Now().UnixNano()%1000) / 1000.0
	}
	return float64(n.Int64()) / float64(1<<53)
}

// ==================
// 1. 性能监控框架
// ==================

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	cpuProfile       *os.File
	memProfile       *os.File
	goroutineProfile *os.File
	blockProfile     *os.File
	mutexProfile     *os.File
	traceFile        *os.File

	// 自定义指标
	metrics    map[string]*Metric
	collectors []MetricCollector

	// 控制
	running bool
	stopCh  chan struct{}
	wg      sync.WaitGroup
	mutex   sync.RWMutex

	// 配置
	sampleRate time.Duration
	profileDir string
}

// Metric 性能指标
type Metric struct {
	Name      string
	Value     float64
	Type      string // counter, gauge, histogram
	Unit      string
	Timestamp time.Time
	Labels    map[string]string
	mutex     sync.RWMutex
}

// MetricCollector 指标收集器接口
type MetricCollector interface {
	Collect() []*Metric
	Name() string
}

func NewPerformanceMonitor(profileDir string) *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics:    make(map[string]*Metric),
		collectors: make([]MetricCollector, 0),
		stopCh:     make(chan struct{}),
		sampleRate: time.Second,
		profileDir: profileDir,
	}
}

func (pm *PerformanceMonitor) Start() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.running {
		return fmt.Errorf("monitor already running")
	}

	// 创建profile目录
	if err := os.MkdirAll(pm.profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %v", err)
	}

	// 启动HTTP pprof服务器
	go func() {
		log.Println("Starting pprof server on :6060")
		server := &http.Server{
			Addr:         "localhost:6060",
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Println(server.ListenAndServe())
	}()

	// 启动指标收集
	pm.running = true
	pm.wg.Add(1)
	go pm.collectMetrics()

	return nil
}

func (pm *PerformanceMonitor) Stop() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.running {
		return
	}

	pm.running = false
	close(pm.stopCh)
	pm.wg.Wait()

	// 关闭所有profile文件
	pm.closeProfiles()
}

func (pm *PerformanceMonitor) collectMetrics() {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.sampleRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.collectRuntimeMetrics()
			pm.collectCustomMetrics()
		case <-pm.stopCh:
			return
		}
	}
}

func (pm *PerformanceMonitor) collectRuntimeMetrics() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	timestamp := time.Now()

	// 内存指标
	pm.updateMetric("heap_alloc", float64(ms.HeapAlloc), "gauge", "bytes", timestamp)
	pm.updateMetric("heap_sys", float64(ms.HeapSys), "gauge", "bytes", timestamp)
	pm.updateMetric("heap_objects", float64(ms.HeapObjects), "gauge", "count", timestamp)
	pm.updateMetric("stack_inuse", float64(ms.StackInuse), "gauge", "bytes", timestamp)
	pm.updateMetric("next_gc", float64(ms.NextGC), "gauge", "bytes", timestamp)

	// GC指标
	pm.updateMetric("num_gc", float64(ms.NumGC), "counter", "count", timestamp)
	pm.updateMetric("gc_pause_total", float64(ms.PauseTotalNs), "counter", "nanoseconds", timestamp)

	// Goroutine指标
	pm.updateMetric("num_goroutine", float64(runtime.NumGoroutine()), "gauge", "count", timestamp)
	pm.updateMetric("num_cgo_call", float64(runtime.NumCgoCall()), "counter", "count", timestamp)

	// CPU指标
	pm.updateMetric("num_cpu", float64(runtime.NumCPU()), "gauge", "count", timestamp)
	pm.updateMetric("gomaxprocs", float64(runtime.GOMAXPROCS(0)), "gauge", "count", timestamp)
}

func (pm *PerformanceMonitor) collectCustomMetrics() {
	for _, collector := range pm.collectors {
		metrics := collector.Collect()
		for _, metric := range metrics {
			pm.metrics[metric.Name] = metric
		}
	}
}

func (pm *PerformanceMonitor) updateMetric(name string, value float64, metricType, unit string, timestamp time.Time) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.metrics[name] = &Metric{
		Name:      name,
		Value:     value,
		Type:      metricType,
		Unit:      unit,
		Timestamp: timestamp,
		Labels:    make(map[string]string),
	}
}

func (pm *PerformanceMonitor) RegisterCollector(collector MetricCollector) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.collectors = append(pm.collectors, collector)
}

func (pm *PerformanceMonitor) GetMetrics() map[string]*Metric {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	result := make(map[string]*Metric)
	for k, v := range pm.metrics {
		result[k] = v
	}
	return result
}

func (pm *PerformanceMonitor) closeProfiles() {
	if pm.cpuProfile != nil {
		if err := pm.cpuProfile.Close(); err != nil {
			log.Printf("Warning: failed to close CPU profile file: %v", err)
		}
	}
	if pm.memProfile != nil {
		if err := pm.memProfile.Close(); err != nil {
			log.Printf("Warning: failed to close memory profile file: %v", err)
		}
	}
	if pm.goroutineProfile != nil {
		if err := pm.goroutineProfile.Close(); err != nil {
			log.Printf("Warning: failed to close goroutine profile file: %v", err)
		}
	}
	if pm.blockProfile != nil {
		if err := pm.blockProfile.Close(); err != nil {
			log.Printf("Warning: failed to close block profile file: %v", err)
		}
	}
	if pm.mutexProfile != nil {
		if err := pm.mutexProfile.Close(); err != nil {
			log.Printf("Warning: failed to close mutex profile file: %v", err)
		}
	}
	if pm.traceFile != nil {
		if err := pm.traceFile.Close(); err != nil {
			log.Printf("Warning: failed to close trace file: %v", err)
		}
	}
}

// ==================
// 2. CPU性能分析
// ==================

// CPUProfiler CPU性能分析器
type CPUProfiler struct {
	duration    time.Duration
	sampleRate  int
	profileFile string
	running     bool
	samples     []CPUSample
	mutex       sync.Mutex
}

type CPUSample struct {
	Timestamp time.Time
	Function  string
	File      string
	Line      int
	CPU       float64
	Samples   int64
}

func NewCPUProfiler(profileFile string, duration time.Duration) *CPUProfiler {
	return &CPUProfiler{
		duration:    duration,
		sampleRate:  100, // 100 Hz
		profileFile: profileFile,
		samples:     make([]CPUSample, 0),
	}
}

func (cp *CPUProfiler) StartProfile() error {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if cp.running {
		return fmt.Errorf("CPU profiling already running")
	}

	// G301安全修复：使用安全的文件权限创建profile文件
	file, err := security.SecureCreateFile(cp.profileFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %v", err)
	}

	err = pprof.StartCPUProfile(file)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: failed to close CPU profile file: %v", closeErr)
		}
		return fmt.Errorf("failed to start CPU profile: %v", err)
	}

	cp.running = true

	// 设置自动停止
	time.AfterFunc(cp.duration, func() {
		cp.StopProfile()
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close CPU profile file: %v", err)
		}
	})

	if _, err := fmt.Printf("CPU profiling started, will run for %v\n", cp.duration); err != nil {
		log.Printf("Warning: failed to print CPU profiling start message: %v", err)
	}
	if _, err := fmt.Printf("Profile will be saved to: %s\n", cp.profileFile); err != nil {
		log.Printf("Warning: failed to print profile file path: %v", err)
	}

	return nil
}

func (cp *CPUProfiler) StopProfile() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if !cp.running {
		return
	}

	pprof.StopCPUProfile()
	cp.running = false

	if _, err := fmt.Printf("CPU profiling stopped. Analyze with: go tool pprof %s\n", cp.profileFile); err != nil {
		log.Printf("Warning: failed to print CPU profiling stop message: %v", err)
	}
}

func (cp *CPUProfiler) IsRunning() bool {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	return cp.running
}

// ==================
// 3. 内存分析
// ==================

// MemoryProfiler 内存性能分析器
type MemoryProfiler struct {
	profileFile string
	gcFirst     bool
	debug       int
}

func NewMemoryProfiler(profileFile string) *MemoryProfiler {
	return &MemoryProfiler{
		profileFile: profileFile,
		gcFirst:     true,
		debug:       0,
	}
}

func (mp *MemoryProfiler) WriteProfile() error {
	if mp.gcFirst {
		runtime.GC() // 强制GC以获得准确的内存使用情况
	}

	// G301安全修复：使用安全的文件权限创建profile文件
	file, err := security.SecureCreateFile(mp.profileFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close memory profile file: %v", err)
		}
	}()

	err = pprof.WriteHeapProfile(file)
	if err != nil {
		return fmt.Errorf("failed to write memory profile: %v", err)
	}

	if _, err := fmt.Printf("Memory profile saved to: %s\n", mp.profileFile); err != nil {
		log.Printf("Warning: failed to print memory profile path: %v", err)
	}
	if _, err := fmt.Printf("Analyze with: go tool pprof %s\n", mp.profileFile); err != nil {
		log.Printf("Warning: failed to print analysis command: %v", err)
	}

	return nil
}

func (mp *MemoryProfiler) WriteAllocProfile() error {
	// G301安全修复：使用安全的文件权限创建profile文件
	file, err := security.SecureCreateFile(mp.profileFile+".alloc", security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create alloc profile file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close alloc profile file: %v", err)
		}
	}()

	profile := pprof.Lookup("allocs")
	if profile == nil {
		return fmt.Errorf("allocs profile not available")
	}

	err = profile.WriteTo(file, mp.debug)
	if err != nil {
		return fmt.Errorf("failed to write alloc profile: %v", err)
	}

	if _, err := fmt.Printf("Allocation profile saved to: %s.alloc\n", mp.profileFile); err != nil {
		log.Printf("Warning: failed to print allocation profile path: %v", err)
	}

	return nil
}

// ==================
// 4. 阻塞分析
// ==================

// BlockProfiler 阻塞分析器
type BlockProfiler struct {
	rate        int
	profileFile string
	enabled     bool
}

func NewBlockProfiler(profileFile string) *BlockProfiler {
	return &BlockProfiler{
		rate:        1, // 记录每次阻塞
		profileFile: profileFile,
	}
}

func (bp *BlockProfiler) Enable() {
	runtime.SetBlockProfileRate(bp.rate)
	bp.enabled = true
	if _, err := fmt.Printf("Block profiling enabled with rate: %d\n", bp.rate); err != nil {
		log.Printf("Warning: failed to print block profiling rate: %v", err)
	}
}

func (bp *BlockProfiler) Disable() {
	runtime.SetBlockProfileRate(0)
	bp.enabled = false
	fmt.Println("Block profiling disabled")
}

func (bp *BlockProfiler) WriteProfile() error {
	if !bp.enabled {
		return fmt.Errorf("block profiling not enabled")
	}

	// G301安全修复：使用安全的文件权限创建profile文件
	file, err := security.SecureCreateFile(bp.profileFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create block profile file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close block profile file: %v", err)
		}
	}()

	profile := pprof.Lookup("block")
	if profile == nil {
		return fmt.Errorf("block profile not available")
	}

	err = profile.WriteTo(file, 0)
	if err != nil {
		return fmt.Errorf("failed to write block profile: %v", err)
	}

	if _, err := fmt.Printf("Block profile saved to: %s\n", bp.profileFile); err != nil {
		log.Printf("Warning: failed to print block profile path: %v", err)
	}
	if _, err := fmt.Printf("Analyze with: go tool pprof %s\n", bp.profileFile); err != nil {
		log.Printf("Warning: failed to print analysis command: %v", err)
	}

	return nil
}

// ==================
// 5. Mutex分析
// ==================

// MutexProfiler Mutex争用分析器
type MutexProfiler struct {
	rate        int
	profileFile string
	enabled     bool
}

func NewMutexProfiler(profileFile string) *MutexProfiler {
	return &MutexProfiler{
		rate:        1, // 记录每次mutex争用
		profileFile: profileFile,
	}
}

func (mp *MutexProfiler) Enable() {
	runtime.SetMutexProfileFraction(mp.rate)
	mp.enabled = true
	if _, err := fmt.Printf("Mutex profiling enabled with rate: %d\n", mp.rate); err != nil {
		log.Printf("Warning: failed to print mutex profiling rate: %v", err)
	}
}

func (mp *MutexProfiler) Disable() {
	runtime.SetMutexProfileFraction(0)
	mp.enabled = false
	fmt.Println("Mutex profiling disabled")
}

func (mp *MutexProfiler) WriteProfile() error {
	if !mp.enabled {
		return fmt.Errorf("mutex profiling not enabled")
	}

	// G301安全修复：使用安全的文件权限创建profile文件
	file, err := security.SecureCreateFile(mp.profileFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create mutex profile file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close mutex profile file: %v", err)
		}
	}()

	profile := pprof.Lookup("mutex")
	if profile == nil {
		return fmt.Errorf("mutex profile not available")
	}

	err = profile.WriteTo(file, 0)
	if err != nil {
		return fmt.Errorf("failed to write mutex profile: %v", err)
	}

	if _, err := fmt.Printf("Mutex profile saved to: %s\n", mp.profileFile); err != nil {
		log.Printf("Warning: failed to print mutex profile path: %v", err)
	}
	if _, err := fmt.Printf("Analyze with: go tool pprof %s\n", mp.profileFile); err != nil {
		log.Printf("Warning: failed to print analysis command: %v", err)
	}

	return nil
}

// ==================
// 6. Goroutine分析
// ==================

// GoroutineProfiler Goroutine分析器
type GoroutineProfiler struct {
	profileFile string
}

func NewGoroutineProfiler(profileFile string) *GoroutineProfiler {
	return &GoroutineProfiler{
		profileFile: profileFile,
	}
}

func (gp *GoroutineProfiler) WriteProfile() error {
	// G301安全修复：使用安全的文件权限创建profile文件
	file, err := security.SecureCreateFile(gp.profileFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Warning: failed to close goroutine profile file: %v", err)
		}
	}()

	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("goroutine profile not available")
	}

	err = profile.WriteTo(file, 2) // debug=2 for detailed stack traces
	if err != nil {
		return fmt.Errorf("failed to write goroutine profile: %v", err)
	}

	if _, err := fmt.Printf("Goroutine profile saved to: %s\n", gp.profileFile); err != nil {
		log.Printf("Warning: failed to print goroutine profile path: %v", err)
	}
	if _, err := fmt.Printf("Analyze with: go tool pprof %s\n", gp.profileFile); err != nil {
		log.Printf("Warning: failed to print analysis command: %v", err)
	}

	return nil
}

func (gp *GoroutineProfiler) PrintGoroutineStats() {
	if _, err := fmt.Printf("\n=== Goroutine统计 ===\n"); err != nil {
		log.Printf("Warning: failed to print goroutine statistics header: %v", err)
	}
	if _, err := fmt.Printf("当前Goroutine数量: %d\n", runtime.NumGoroutine()); err != nil {
		log.Printf("Warning: failed to print goroutine count: %v", err)
	}

	// 获取goroutine stack dump
	buf := make([]byte, 1<<20) // 1MB buffer
	stackSize := runtime.Stack(buf, true)

	// 分析stack dump
	scanner := bufio.NewScanner(bytes.NewReader(buf[:stackSize]))
	goroutineCount := 0
	stateCount := make(map[string]int)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "goroutine ") {
			goroutineCount++
			// 解析goroutine状态
			if idx := strings.Index(line, "["); idx != -1 {
				if end := strings.Index(line[idx:], "]"); end != -1 {
					state := line[idx+1 : idx+end]
					stateCount[state]++
				}
			}
		}
	}

	if _, err := fmt.Printf("Stack dump中的Goroutine数量: %d\n", goroutineCount); err != nil {
		log.Printf("Warning: failed to print stack dump goroutine count: %v", err)
	}
	if _, err := fmt.Printf("Goroutine状态分布:\n"); err != nil {
		log.Printf("Warning: failed to print goroutine states header: %v", err)
	}
	for state, count := range stateCount {
		if _, err := fmt.Printf("  %s: %d\n", state, count); err != nil {
			log.Printf("Warning: failed to print goroutine state info: %v", err)
		}
	}
}

// ==================
// 7. 执行跟踪
// ==================

// ExecutionTracer 执行跟踪器
type ExecutionTracer struct {
	traceFile string
	duration  time.Duration
	running   bool
	file      *os.File
	mutex     sync.Mutex
}

func NewExecutionTracer(traceFile string, duration time.Duration) *ExecutionTracer {
	return &ExecutionTracer{
		traceFile: traceFile,
		duration:  duration,
	}
}

func (et *ExecutionTracer) StartTrace() error {
	et.mutex.Lock()
	defer et.mutex.Unlock()

	if et.running {
		return fmt.Errorf("execution tracing already running")
	}

	// G301安全修复：使用安全的文件权限创建trace文件
	file, err := security.SecureCreateFile(et.traceFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("failed to create trace file: %v", err)
	}

	err = trace.Start(file)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Warning: failed to close trace file: %v", closeErr)
		}
		return fmt.Errorf("failed to start execution trace: %v", err)
	}

	et.file = file
	et.running = true

	// 设置自动停止
	time.AfterFunc(et.duration, func() {
		et.StopTrace()
	})

	if _, err := fmt.Printf("Execution tracing started, will run for %v\n", et.duration); err != nil {
		log.Printf("Warning: failed to print trace start message: %v", err)
	}
	if _, err := fmt.Printf("Trace will be saved to: %s\n", et.traceFile); err != nil {
		log.Printf("Warning: failed to print trace file path: %v", err)
	}

	return nil
}

func (et *ExecutionTracer) StopTrace() {
	et.mutex.Lock()
	defer et.mutex.Unlock()

	if !et.running {
		return
	}

	trace.Stop()
	if err := et.file.Close(); err != nil {
		log.Printf("Warning: failed to close trace file: %v", err)
	}
	et.running = false

	if _, err := fmt.Printf("Execution tracing stopped. Analyze with: go tool trace %s\n", et.traceFile); err != nil {
		log.Printf("Warning: failed to print trace stop message: %v", err)
	}
}

func (et *ExecutionTracer) IsRunning() bool {
	et.mutex.Lock()
	defer et.mutex.Unlock()
	return et.running
}

// ==================
// 8. 自定义指标收集器
// ==================

// ApplicationMetricsCollector 应用程序指标收集器
type ApplicationMetricsCollector struct {
	requestCount      int64
	errorCount        int64
	responseTimeSum   int64
	activeConnections int64
}

func (amc *ApplicationMetricsCollector) Collect() []*Metric {
	timestamp := time.Now()

	return []*Metric{
		{
			Name:      "app_requests_total",
			Value:     float64(amc.requestCount),
			Type:      "counter",
			Unit:      "count",
			Timestamp: timestamp,
			Labels:    map[string]string{"service": "api"},
		},
		{
			Name:      "app_errors_total",
			Value:     float64(amc.errorCount),
			Type:      "counter",
			Unit:      "count",
			Timestamp: timestamp,
			Labels:    map[string]string{"service": "api"},
		},
		{
			Name:      "app_response_time_sum",
			Value:     float64(amc.responseTimeSum),
			Type:      "counter",
			Unit:      "nanoseconds",
			Timestamp: timestamp,
			Labels:    map[string]string{"service": "api"},
		},
		{
			Name:      "app_active_connections",
			Value:     float64(amc.activeConnections),
			Type:      "gauge",
			Unit:      "count",
			Timestamp: timestamp,
			Labels:    map[string]string{"service": "api"},
		},
	}
}

func (amc *ApplicationMetricsCollector) Name() string {
	return "application_metrics"
}

func (amc *ApplicationMetricsCollector) IncrementRequests() {
	amc.requestCount++
}

func (amc *ApplicationMetricsCollector) IncrementErrors() {
	amc.errorCount++
}

func (amc *ApplicationMetricsCollector) AddResponseTime(duration time.Duration) {
	amc.responseTimeSum += duration.Nanoseconds()
}

func (amc *ApplicationMetricsCollector) SetActiveConnections(count int64) {
	amc.activeConnections = count
}

// ==================
// 9. 性能测试工作负载
// ==================

// WorkloadGenerator 工作负载生成器
type WorkloadGenerator struct {
	cpuIntensive    bool
	memoryIntensive bool
	ioIntensive     bool
	concurrency     int
	duration        time.Duration

	// 用于阻塞分析的mutex
	contentionMutex sync.Mutex

	// 用于内存分析的数据
	allocatedData [][]byte
	allocMutex    sync.Mutex
}

func NewWorkloadGenerator() *WorkloadGenerator {
	return &WorkloadGenerator{
		concurrency:   runtime.NumCPU(),
		duration:      30 * time.Second,
		allocatedData: make([][]byte, 0),
	}
}

func (wg *WorkloadGenerator) SetCPUIntensive(enabled bool) {
	wg.cpuIntensive = enabled
}

func (wg *WorkloadGenerator) SetMemoryIntensive(enabled bool) {
	wg.memoryIntensive = enabled
}

func (wg *WorkloadGenerator) SetIOIntensive(enabled bool) {
	wg.ioIntensive = enabled
}

func (wg *WorkloadGenerator) SetConcurrency(concurrency int) {
	wg.concurrency = concurrency
}

func (wg *WorkloadGenerator) SetDuration(duration time.Duration) {
	wg.duration = duration
}

func (wg *WorkloadGenerator) Run(ctx context.Context) {
	fmt.Printf("Starting workload generator with %d goroutines for %v\n",
		wg.concurrency, wg.duration)

	var workWG sync.WaitGroup

	// 启动工作goroutines
	for i := 0; i < wg.concurrency; i++ {
		workWG.Add(1)
		go func(workerID int) {
			defer workWG.Done()
			wg.runWorker(ctx, workerID)
		}(i)
	}

	// 启动mutex争用生成器
	if wg.cpuIntensive {
		for i := 0; i < wg.concurrency/2; i++ {
			workWG.Add(1)
			go func() {
				defer workWG.Done()
				wg.generateMutexContention(ctx)
			}()
		}
	}

	// 等待所有worker完成
	workWG.Wait()
	fmt.Println("Workload generation completed")
}

func (wg *WorkloadGenerator) runWorker(ctx context.Context, workerID int) {
	timeout := time.After(wg.duration)

	for {
		select {
		case <-timeout:
			return
		case <-ctx.Done():
			return
		default:
			if wg.cpuIntensive {
				wg.doCPUWork()
			}
			if wg.memoryIntensive {
				wg.doMemoryWork()
			}
			if wg.ioIntensive {
				wg.doIOWork()
			}

			// 短暂休息避免100% CPU
			time.Sleep(time.Microsecond * 100)
		}
	}
}

func (wg *WorkloadGenerator) doCPUWork() {
	// CPU密集型计算
	sum := 0
	for i := 0; i < 10000; i++ {
		sum += i * i
	}

	// 一些数学运算
	x := float64(sum)
	for i := 0; i < 100; i++ {
		x = x * 1.1
		if x > 1e10 {
			x = x / 1e9
		}
	}
}

func (wg *WorkloadGenerator) doMemoryWork() {
	// 分配各种大小的内存
	sizes := []int{64, 128, 256, 512, 1024, 2048, 4096}
	size := sizes[secureRandomInt(len(sizes))]

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(secureRandomInt(256))
	}

	wg.allocMutex.Lock()
	wg.allocatedData = append(wg.allocatedData, data)

	// 随机释放一些内存
	if len(wg.allocatedData) > 1000 {
		// 释放前一半
		wg.allocatedData = wg.allocatedData[500:]
	}
	wg.allocMutex.Unlock()
}

func (wg *WorkloadGenerator) doIOWork() {
	// 模拟I/O操作
	time.Sleep(time.Microsecond * time.Duration(secureRandomInt(1000)))
}

func (wg *WorkloadGenerator) generateMutexContention(ctx context.Context) {
	timeout := time.After(wg.duration)

	for {
		select {
		case <-timeout:
			return
		case <-ctx.Done():
			return
		default:
			// 生成mutex争用
			wg.contentionMutex.Lock()
			time.Sleep(time.Microsecond * time.Duration(secureRandomInt(100)))
			wg.contentionMutex.Unlock()

			time.Sleep(time.Microsecond * time.Duration(secureRandomInt(50)))
		}
	}
}

// ==================
// 10. 主演示函数
// ==================

func demonstratePerformanceProfiling() {
	fmt.Println("=== Go性能分析与诊断深度解析 ===")

	// 1. 启动性能监控
	monitor := NewPerformanceMonitor("./profiles")
	err := monitor.Start()
	if err != nil {
		log.Printf("Failed to start performance monitor: %v", err)
		return
	}
	defer monitor.Stop()

	// 注册自定义指标收集器
	appMetrics := &ApplicationMetricsCollector{}
	monitor.RegisterCollector(appMetrics)

	fmt.Println("\n1. 性能监控框架启动")
	fmt.Println("HTTP pprof服务器: http://localhost:6060/debug/pprof/")
	fmt.Println("可用的profile endpoints:")
	fmt.Println("  /debug/pprof/heap - 内存profile")
	fmt.Println("  /debug/pprof/goroutine - goroutine profile")
	fmt.Println("  /debug/pprof/profile - CPU profile (30秒)")
	fmt.Println("  /debug/pprof/block - 阻塞profile")
	fmt.Println("  /debug/pprof/mutex - mutex profile")
	fmt.Println("  /debug/pprof/trace - 执行跟踪")

	// 2. 设置各种profiler
	cpuProfiler := NewCPUProfiler("./profiles/cpu.prof", 10*time.Second)
	memProfiler := NewMemoryProfiler("./profiles/mem.prof")
	blockProfiler := NewBlockProfiler("./profiles/block.prof")
	mutexProfiler := NewMutexProfiler("./profiles/mutex.prof")
	goroutineProfiler := NewGoroutineProfiler("./profiles/goroutine.prof")
	tracer := NewExecutionTracer("./profiles/trace.out", 5*time.Second)

	// 3. 启用block和mutex profiling
	fmt.Println("\n2. 启用各种性能分析")
	blockProfiler.Enable()
	mutexProfiler.Enable()

	// 4. 创建工作负载
	fmt.Println("\n3. 生成性能分析工作负载")
	workload := NewWorkloadGenerator()
	workload.SetCPUIntensive(true)
	workload.SetMemoryIntensive(true)
	workload.SetIOIntensive(true)
	workload.SetConcurrency(8)
	workload.SetDuration(15 * time.Second)

	// 5. 启动CPU profiling和execution tracing
	fmt.Println("\n4. 启动CPU profiling和execution tracing")
	if err := cpuProfiler.StartProfile(); err != nil {
		log.Printf("Failed to start CPU profiling: %v", err)
	}
	if err := tracer.StartTrace(); err != nil {
		log.Printf("Failed to start execution tracing: %v", err)
	}

	// 6. 运行工作负载
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 模拟应用程序指标更新
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				appMetrics.IncrementRequests()
				if secureRandomFloat64() < 0.1 { // 10%错误率
					appMetrics.IncrementErrors()
				}
				appMetrics.AddResponseTime(time.Duration(secureRandomInt(100)) * time.Millisecond)
				appMetrics.SetActiveConnections(int64(secureRandomInt(100)))
			}
		}
	}()

	workload.Run(ctx)

	// 7. 等待profiling完成
	fmt.Println("\n5. 等待性能分析完成...")
	for cpuProfiler.IsRunning() || tracer.IsRunning() {
		time.Sleep(time.Second)
	}

	// 8. 生成各种profiles
	fmt.Println("\n6. 生成性能分析文件")

	if err := memProfiler.WriteProfile(); err != nil {
		log.Printf("Failed to write memory profile: %v", err)
	}

	if err := memProfiler.WriteAllocProfile(); err != nil {
		log.Printf("Failed to write allocation profile: %v", err)
	}

	if err := blockProfiler.WriteProfile(); err != nil {
		log.Printf("Failed to write block profile: %v", err)
	}

	if err := mutexProfiler.WriteProfile(); err != nil {
		log.Printf("Failed to write mutex profile: %v", err)
	}

	if err := goroutineProfiler.WriteProfile(); err != nil {
		log.Printf("Failed to write goroutine profile: %v", err)
	}

	// 9. 打印goroutine统计
	goroutineProfiler.PrintGoroutineStats()

	// 10. 打印收集的指标
	fmt.Println("\n7. 性能指标统计")
	metrics := monitor.GetMetrics()

	// 按名称排序指标
	var names []string
	for name := range metrics {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("收集到 %d 个性能指标:\n", len(metrics))
	for _, name := range names {
		metric := metrics[name]
		fmt.Printf("  %s: %.2f %s (%s)\n",
			metric.Name, metric.Value, metric.Unit, metric.Type)
	}

	// 11. 禁用profiling
	blockProfiler.Disable()
	mutexProfiler.Disable()

	fmt.Println("\n=== 性能分析文件生成完成 ===")
	fmt.Println("\n分析命令:")
	fmt.Println("  CPU分析:    go tool pprof ./profiles/cpu.prof")
	fmt.Println("  内存分析:   go tool pprof ./profiles/mem.prof")
	fmt.Println("  阻塞分析:   go tool pprof ./profiles/block.prof")
	fmt.Println("  Mutex分析:  go tool pprof ./profiles/mutex.prof")
	fmt.Println("  Goroutine:  go tool pprof ./profiles/goroutine.prof")
	fmt.Println("  执行跟踪:   go tool trace ./profiles/trace.out")

	fmt.Println("\n交互式分析:")
	fmt.Println("  pprof> top10        # 显示前10个函数")
	fmt.Println("  pprof> list func    # 显示函数源码")
	fmt.Println("  pprof> web          # 生成调用图")
	fmt.Println("  pprof> peek regex   # 查看匹配的函数")
}

func main() {
	// 确保profiles目录存在
	if err := os.MkdirAll("./profiles", 0755); err != nil {
		log.Printf("Warning: failed to create profiles directory: %v", err)
	}

	demonstratePerformanceProfiling()

	fmt.Println("\n=== Go性能分析与诊断深度解析完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. pprof提供多种性能分析类型：CPU、内存、阻塞、Mutex、Goroutine")
	fmt.Println("2. HTTP pprof端点可以实时获取性能数据")
	fmt.Println("3. 执行跟踪(trace)提供最详细的运行时行为分析")
	fmt.Println("4. 自定义指标收集器可以监控应用级别的性能")
	fmt.Println("5. 性能分析需要在生产负载下进行才有意义")
	fmt.Println("6. 不同类型的profiling适用于不同的性能问题")

	fmt.Println("\n高级特性:")
	fmt.Println("- CPU profiling使用统计采样，默认100Hz")
	fmt.Println("- 内存profiling跟踪分配而不是使用量")
	fmt.Println("- Block profiling识别goroutine阻塞瓶颈")
	fmt.Println("- Mutex profiling发现锁竞争问题")
	fmt.Println("- Execution trace显示调度器行为")
	fmt.Println("- 可以通过HTTP端点远程收集profile")
}

/*
=== 练习题 ===

1. 基础练习：
   - 分析一个简单程序的CPU热点
   - 识别内存泄漏问题
   - 测量goroutine阻塞时间
   - 创建自定义性能指标

2. 中级练习：
   - 分析Web服务器的性能瓶颈
   - 优化高并发程序的锁竞争
   - 使用execution trace分析调度问题
   - 实现性能回归检测

3. 高级练习：
   - 构建生产级性能监控系统
   - 分析大型微服务的性能
   - 实现自适应性能调优
   - 集成APM系统

4. 深度分析：
   - 分析Go runtime的性能特征
   - 研究不同GC策略的影响
   - 优化内存分配器性能
   - 分析NUMA系统的性能

5. 工具开发：
   - 开发自定义profiling工具
   - 实现性能数据可视化
   - 创建性能基准测试框架
   - 构建持续性能监控

运行命令：
go run main.go

然后在另一个终端中：
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap
go tool trace http://localhost:6060/debug/pprof/trace

重要概念：
- Sampling profiler: 统计采样式性能分析
- Call graph: 函数调用关系图
- Flame graph: 火焰图可视化
- Hot path: 性能热点路径
- Profile delta: 性能差异对比
- Continuous profiling: 持续性能分析
*/
