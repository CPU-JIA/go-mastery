/*
=== Go语言性能优化基础：从应用开发到系统编程的桥梁 ===

本模块是Web开发和系统编程之间的重要过渡，专注于Go语言性能优化基础，探索：
1. 性能分析理论基础 - CPU、内存、I/O性能概念
2. Go性能分析工具链 - pprof、trace、benchmark深度使用
3. 内存管理和垃圾收集入门 - GC基础原理和调优
4. 并发性能优化入门 - goroutine调优和锁优化
5. 性能监控和诊断实践 - 指标收集和问题定位
6. 常见性能问题解决 - 内存泄漏、CPU热点、I/O瓶颈
7. 性能测试和基准测试 - 自动化性能验证
8. 生产环境性能调优 - 实际案例和最佳实践
9. 性能优化工作流 - 系统化的性能改进方法
10. 为系统编程做准备 - 深层次性能概念预习

学习目标：
- 掌握Go语言性能分析的基本方法和工具
- 理解内存管理和垃圾收集的基础概念
- 学会识别和解决常见的性能问题
- 建立系统性的性能优化思维
- 为深入学习运行时内核打下坚实基础
*/

package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"sync/atomic"
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

// ==================
// 1. 性能分析理论基础
// ==================

// PerformanceTheory 性能理论基础
type PerformanceTheory struct {
	concepts map[string]PerformanceConcept
}

type PerformanceConcept struct {
	Name        string
	Description string
	Category    string
	Impact      int // 1-5, 影响程度
	Examples    []string
	Solutions   []string
}

func NewPerformanceTheory() *PerformanceTheory {
	pt := &PerformanceTheory{
		concepts: make(map[string]PerformanceConcept),
	}
	pt.initializeConcepts()
	return pt
}

func (pt *PerformanceTheory) initializeConcepts() {
	pt.concepts["cpu-bound"] = PerformanceConcept{
		Name:        "CPU密集型",
		Description: "程序的性能受CPU处理能力限制",
		Category:    "CPU",
		Impact:      5,
		Examples:    []string{"数学计算", "加密算法", "图像处理", "编译"},
		Solutions:   []string{"算法优化", "并行处理", "缓存计算结果", "硬件升级"},
	}

	pt.concepts["memory-bound"] = PerformanceConcept{
		Name:        "内存密集型",
		Description: "程序性能受内存访问模式和容量限制",
		Category:    "Memory",
		Impact:      4,
		Examples:    []string{"大数据处理", "缓存系统", "内存数据库"},
		Solutions:   []string{"优化数据结构", "减少内存分配", "使用内存池", "数据压缩"},
	}

	pt.concepts["io-bound"] = PerformanceConcept{
		Name:        "I/O密集型",
		Description: "程序性能受I/O操作（磁盘、网络）限制",
		Category:    "I/O",
		Impact:      5,
		Examples:    []string{"文件处理", "数据库查询", "网络请求", "日志写入"},
		Solutions:   []string{"异步I/O", "缓存机制", "连接池", "批量操作"},
	}

	pt.concepts["lock-contention"] = PerformanceConcept{
		Name:        "锁竞争",
		Description: "多个goroutine竞争同一锁资源导致的性能问题",
		Category:    "Concurrency",
		Impact:      4,
		Examples:    []string{"共享数据结构", "日志记录", "计数器更新"},
		Solutions:   []string{"减少锁粒度", "使用原子操作", "无锁算法", "分片设计"},
	}

	pt.concepts["gc-pressure"] = PerformanceConcept{
		Name:        "GC压力",
		Description: "频繁的内存分配导致垃圾收集器负担过重",
		Category:    "GC",
		Impact:      4,
		Examples:    []string{"频繁创建对象", "大量字符串拼接", "切片频繁扩容"},
		Solutions:   []string{"对象重用", "内存预分配", "减少指针使用", "调整GC参数"},
	}
}

func (pt *PerformanceTheory) ExplainConcept(name string) {
	if concept, exists := pt.concepts[name]; exists {
		fmt.Printf("=== %s ===\n", concept.Name)
		fmt.Printf("描述: %s\n", concept.Description)
		fmt.Printf("类别: %s\n", concept.Category)
		fmt.Printf("影响程度: %d/5\n", concept.Impact)
		fmt.Printf("常见示例:\n")
		for _, example := range concept.Examples {
			fmt.Printf("  - %s\n", example)
		}
		fmt.Printf("解决方案:\n")
		for _, solution := range concept.Solutions {
			fmt.Printf("  - %s\n", solution)
		}
	}
}

func (pt *PerformanceTheory) ListAllConcepts() {
	fmt.Printf("性能优化核心概念 (%d个):\n", len(pt.concepts))
	for _, concept := range pt.concepts {
		fmt.Printf("  %s (%s) - 影响程度: %d/5\n",
			concept.Name, concept.Category, concept.Impact)
	}
}

func demonstratePerformanceTheory() {
	fmt.Println("=== 1. 性能分析理论基础 ===")

	theory := NewPerformanceTheory()
	theory.ListAllConcepts()

	fmt.Println("\n详细解释关键概念:")
	theory.ExplainConcept("cpu-bound")
	fmt.Println()
	theory.ExplainConcept("gc-pressure")
	fmt.Println()
}

// ==================
// 2. Go性能分析工具链
// ==================

// PerformanceProfiler Go性能分析器
type PerformanceProfiler struct {
	name       string
	startTime  time.Time
	cpuProfile *os.File
	memProfile *os.File
	traceFile  *os.File
	metrics    ProfilerMetrics
	config     ProfilerConfig
}

type ProfilerMetrics struct {
	CPUSamples      int64
	MemoryAllocated int64
	GoroutineCount  int
	GCCount         uint32
	GCPauseTotal    time.Duration
}

type ProfilerConfig struct {
	EnableCPUProfile    bool
	EnableMemoryProfile bool
	EnableTrace         bool
	SamplingRate        int
	ProfileDuration     time.Duration
}

func NewPerformanceProfiler(name string) *PerformanceProfiler {
	return &PerformanceProfiler{
		name:      name,
		startTime: time.Now(),
		config: ProfilerConfig{
			EnableCPUProfile:    true,
			EnableMemoryProfile: true,
			EnableTrace:         false,
			SamplingRate:        100,
			ProfileDuration:     30 * time.Second,
		},
	}
}

func (pp *PerformanceProfiler) StartProfiling() error {
	fmt.Printf("开始性能分析: %s\n", pp.name)

	// CPU Profile
	if pp.config.EnableCPUProfile {
		cpuFile, err := os.Create(fmt.Sprintf("cpu_%s.prof", pp.name))
		if err != nil {
			return err
		}
		pp.cpuProfile = cpuFile

		if err := pprof.StartCPUProfile(cpuFile); err != nil {
			return err
		}
		fmt.Println("  CPU profiling 已启动")
	}

	// Trace
	if pp.config.EnableTrace {
		traceFile, err := os.Create(fmt.Sprintf("trace_%s.out", pp.name))
		if err != nil {
			return err
		}
		pp.traceFile = traceFile

		if err := trace.Start(traceFile); err != nil {
			return err
		}
		fmt.Println("  Trace 已启动")
	}

	// 启动内存监控
	go pp.monitorMemory()

	return nil
}

func (pp *PerformanceProfiler) StopProfiling() error {
	fmt.Printf("停止性能分析: %s\n", pp.name)

	// 停止CPU Profile
	if pp.cpuProfile != nil {
		pprof.StopCPUProfile()
		pp.cpuProfile.Close()
		fmt.Println("  CPU profiling 已停止")
	}

	// 停止Trace
	if pp.traceFile != nil {
		trace.Stop()
		pp.traceFile.Close()
		fmt.Println("  Trace 已停止")
	}

	// 生成内存Profile
	if pp.config.EnableMemoryProfile {
		memFile, err := os.Create(fmt.Sprintf("mem_%s.prof", pp.name))
		if err != nil {
			return err
		}
		defer memFile.Close()

		runtime.GC() // 强制GC以获取准确的内存状态
		if err := pprof.WriteHeapProfile(memFile); err != nil {
			return err
		}
		fmt.Println("  Memory profile 已生成")
	}

	pp.printSummary()
	return nil
}

func (pp *PerformanceProfiler) monitorMemory() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			pp.metrics.MemoryAllocated = security.MustSafeUint64ToInt64(m.Alloc)
			pp.metrics.GoroutineCount = runtime.NumGoroutine()
			pp.metrics.GCCount = m.NumGC
			pp.metrics.GCPauseTotal = time.Duration(m.PauseTotalNs)

		case <-time.After(pp.config.ProfileDuration):
			return
		}
	}
}

func (pp *PerformanceProfiler) printSummary() {
	duration := time.Since(pp.startTime)
	fmt.Printf("\n性能分析总结 (%s):\n", pp.name)
	fmt.Printf("  运行时间: %v\n", duration)
	fmt.Printf("  内存使用: %s\n", formatBytes(pp.metrics.MemoryAllocated))
	fmt.Printf("  Goroutine数量: %d\n", pp.metrics.GoroutineCount)
	fmt.Printf("  GC次数: %d\n", pp.metrics.GCCount)
	fmt.Printf("  GC总暂停时间: %v\n", pp.metrics.GCPauseTotal)

	fmt.Printf("\n分析文件生成:\n")
	if pp.config.EnableCPUProfile {
		fmt.Printf("  CPU Profile: cpu_%s.prof\n", pp.name)
		fmt.Println("    使用方法: go tool pprof cpu_xxx.prof")
	}
	if pp.config.EnableMemoryProfile {
		fmt.Printf("  Memory Profile: mem_%s.prof\n", pp.name)
		fmt.Println("    使用方法: go tool pprof mem_xxx.prof")
	}
	if pp.config.EnableTrace {
		fmt.Printf("  Trace文件: trace_%s.out\n", pp.name)
		fmt.Println("    使用方法: go tool trace trace_xxx.out")
	}
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// CPU密集型任务示例
func cpuIntensiveTask() {
	fmt.Println("执行CPU密集型任务...")

	// 计算素数
	primes := make([]int, 0)
	for n := 2; n < 10000; n++ {
		isPrime := true
		for i := 2; i*i <= n; i++ {
			if n%i == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			primes = append(primes, n)
		}
	}
	fmt.Printf("找到素数 %d 个\n", len(primes))
}

// 内存密集型任务示例
func memoryIntensiveTask() {
	fmt.Println("执行内存密集型任务...")

	// 创建大量对象
	data := make([][]int, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = make([]int, 1000)
		for j := 0; j < 1000; j++ {
			data[i][j] = i * j
		}
	}

	// 处理数据
	sum := 0
	for _, row := range data {
		for _, val := range row {
			sum += val
		}
	}
	fmt.Printf("数据处理完成，总和: %d\n", sum)
}

func demonstrateProfilingTools() {
	fmt.Println("=== 2. Go性能分析工具链 ===")

	// 创建性能分析器
	profiler := NewPerformanceProfiler("demo")

	// 启动分析
	if err := profiler.StartProfiling(); err != nil {
		log.Printf("启动性能分析失败: %v", err)
		return
	}

	// 执行测试任务
	cpuIntensiveTask()
	memoryIntensiveTask()

	// 停止分析
	if err := profiler.StopProfiling(); err != nil {
		log.Printf("停止性能分析失败: %v", err)
	}

	fmt.Println("\n常用pprof命令:")
	fmt.Println("  go tool pprof cpu_demo.prof")
	fmt.Println("    (pprof) top10       # 显示CPU使用最多的10个函数")
	fmt.Println("    (pprof) list main   # 显示main函数的详细分析")
	fmt.Println("    (pprof) web         # 生成调用图")
	fmt.Println("    (pprof) png         # 生成PNG格式的调用图")

	fmt.Println("\n内存分析命令:")
	fmt.Println("  go tool pprof mem_demo.prof")
	fmt.Println("    (pprof) top         # 显示内存使用最多的函数")
	fmt.Println("    (pprof) list        # 显示详细的内存分配信息")

	fmt.Println()
}

// ==================
// 3. 基准测试和性能测试
// ==================

// BenchmarkSuite 基准测试套件
type BenchmarkSuite struct {
	name      string
	results   []BenchmarkResult
	baselines map[string]float64
}

type BenchmarkResult struct {
	Name         string
	Iterations   int64
	NsPerOp      int64
	AllocedBytes int64
	AllocsPerOp  int64
	MBPerSec     float64
}

func NewBenchmarkSuite(name string) *BenchmarkSuite {
	return &BenchmarkSuite{
		name:      name,
		results:   make([]BenchmarkResult, 0),
		baselines: make(map[string]float64),
	}
}

func (bs *BenchmarkSuite) RunBenchmark(name string, fn func()) BenchmarkResult {
	fmt.Printf("运行基准测试: %s\n", name)

	// 预热
	for i := 0; i < 1000; i++ {
		fn()
	}

	// 测量内存基准
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 执行基准测试
	iterations := int64(10000)
	start := time.Now()

	for i := int64(0); i < iterations; i++ {
		fn()
	}

	duration := time.Since(start)
	runtime.ReadMemStats(&m2)

	result := BenchmarkResult{
		Name:         name,
		Iterations:   iterations,
		NsPerOp:      duration.Nanoseconds() / iterations,
		AllocedBytes: security.MustSafeUint64ToInt64(m2.TotalAlloc - m1.TotalAlloc),
		AllocsPerOp:  security.MustSafeUint64ToInt64(m2.Mallocs-m1.Mallocs) / iterations,
	}

	bs.results = append(bs.results, result)
	return result
}

func (bs *BenchmarkSuite) CompareWithBaseline(name string, baseline float64) {
	bs.baselines[name] = baseline
}

func (bs *BenchmarkSuite) PrintResults() {
	fmt.Printf("\n基准测试结果 (%s):\n", bs.name)
	fmt.Printf("%-25s %12s %15s %15s %15s\n",
		"名称", "迭代次数", "纳秒/操作", "字节/操作", "分配/操作")
	fmt.Println(strings.Repeat("-", 85))

	for _, result := range bs.results {
		fmt.Printf("%-25s %12d %15d %15d %15d\n",
			result.Name,
			result.Iterations,
			result.NsPerOp,
			result.AllocedBytes/result.Iterations,
			result.AllocsPerOp)

		// 与基准比较
		if baseline, exists := bs.baselines[result.Name]; exists {
			improvement := (baseline - float64(result.NsPerOp)) / baseline * 100
			if improvement > 0 {
				fmt.Printf("  -> 比基准快 %.1f%%\n", improvement)
			} else {
				fmt.Printf("  -> 比基准慢 %.1f%%\n", -improvement)
			}
		}
	}
}

// 基准测试示例函数

// 字符串拼接基准测试
func benchmarkStringConcat() {
	var s string
	for i := 0; i < 100; i++ {
		s += "hello"
	}
}

func benchmarkStringBuilder() {
	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString("hello")
	}
	_ = builder.String()
}

// 切片操作基准测试
func benchmarkSliceAppend() {
	s := make([]int, 0)
	for i := 0; i < 100; i++ {
		s = append(s, i)
	}
}

func benchmarkSlicePrealloc() {
	s := make([]int, 0, 100) // 预分配容量
	for i := 0; i < 100; i++ {
		s = append(s, i)
	}
}

// Map操作基准测试
func benchmarkMapAccess() {
	m := make(map[int]string)
	for i := 0; i < 100; i++ {
		m[i] = fmt.Sprintf("value%d", i)
	}

	for i := 0; i < 100; i++ {
		_ = m[i]
	}
}

func benchmarkMapPrealloc() {
	m := make(map[int]string, 100) // 预分配容量
	for i := 0; i < 100; i++ {
		m[i] = fmt.Sprintf("value%d", i)
	}

	for i := 0; i < 100; i++ {
		_ = m[i]
	}
}

func demonstrateBenchmarking() {
	fmt.Println("=== 3. 基准测试和性能测试 ===")

	suite := NewBenchmarkSuite("性能对比测试")

	// 字符串拼接基准测试
	result1 := suite.RunBenchmark("字符串拼接", benchmarkStringConcat)
	suite.CompareWithBaseline("字符串拼接", float64(result1.NsPerOp))

	result2 := suite.RunBenchmark("StringBuilder", benchmarkStringBuilder)
	suite.CompareWithBaseline("StringBuilder", float64(result2.NsPerOp))

	// 切片操作基准测试
	result3 := suite.RunBenchmark("切片动态扩容", benchmarkSliceAppend)
	suite.CompareWithBaseline("切片动态扩容", float64(result3.NsPerOp))

	result4 := suite.RunBenchmark("切片预分配", benchmarkSlicePrealloc)
	suite.CompareWithBaseline("切片预分配", float64(result4.NsPerOp))

	// Map操作基准测试
	result5 := suite.RunBenchmark("Map动态扩容", benchmarkMapAccess)
	suite.CompareWithBaseline("Map动态扩容", float64(result5.NsPerOp))

	result6 := suite.RunBenchmark("Map预分配", benchmarkMapPrealloc)
	suite.CompareWithBaseline("Map预分配", float64(result6.NsPerOp))

	suite.PrintResults()

	fmt.Println("\n性能优化结论:")
	fmt.Println("  1. StringBuilder比字符串拼接快10-100倍")
	fmt.Println("  2. 预分配切片容量可避免重新分配")
	fmt.Println("  3. 预分配Map容量可减少哈希重建")
	fmt.Println("  4. 减少内存分配是性能优化的关键")

	fmt.Println()
}

// ==================
// 4. 内存管理和GC基础
// ==================

// MemoryManager 内存管理器
type MemoryManager struct {
	pools   map[string]*sync.Pool
	metrics MemoryMetrics
	config  MemoryConfig
}

type MemoryMetrics struct {
	TotalAllocated int64
	TotalFreed     int64
	ObjectsCreated int64
	ObjectsReused  int64
	GCCycles       int64
	GCPauseTime    time.Duration
}

type MemoryConfig struct {
	EnableObjectPooling bool
	PoolMaxSize         int
	GCTargetPercent     int
	EnableMemoryDebug   bool
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		pools: make(map[string]*sync.Pool),
		config: MemoryConfig{
			EnableObjectPooling: true,
			PoolMaxSize:         1000,
			GCTargetPercent:     100,
			EnableMemoryDebug:   true,
		},
	}
}

func (mm *MemoryManager) CreatePool(name string, newFunc func() interface{}) {
	mm.pools[name] = &sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&mm.metrics.ObjectsCreated, 1)
			return newFunc()
		},
	}
}

func (mm *MemoryManager) GetObject(poolName string) interface{} {
	if pool, exists := mm.pools[poolName]; exists {
		obj := pool.Get()
		atomic.AddInt64(&mm.metrics.ObjectsReused, 1)
		return obj
	}
	return nil
}

func (mm *MemoryManager) PutObject(poolName string, obj interface{}) {
	if pool, exists := mm.pools[poolName]; exists {
		pool.Put(obj)
	}
}

func (mm *MemoryManager) MonitorMemory(duration time.Duration) {
	fmt.Printf("开始内存监控 (持续 %v)\n", duration)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(duration)
	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			mm.printMemoryStats()
		case <-timeout:
			fmt.Printf("内存监控完成 (持续 %v)\n", time.Since(startTime))
			return
		}
	}
}

func (mm *MemoryManager) printMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("[内存] 堆大小: %s, 使用: %s, GC次数: %d, 暂停: %v\n",
		formatBytes(security.MustSafeUint64ToInt64(m.HeapSys)),
		formatBytes(security.MustSafeUint64ToInt64(m.HeapAlloc)),
		m.NumGC,
		time.Duration(m.PauseTotalNs))
}

func (mm *MemoryManager) TriggerGCTuning() {
	fmt.Println("演示GC调优...")

	// 保存原始设置
	originalPercent := debug.SetGCPercent(-1)
	fmt.Printf("原始GC目标百分比: %d%%\n", originalPercent)

	// 设置不同的GC目标
	testPercents := []int{50, 100, 200}

	for _, percent := range testPercents {
		fmt.Printf("\n设置GC目标为 %d%%:\n", percent)
		debug.SetGCPercent(percent)

		// 分配内存测试GC行为
		mm.allocateMemoryForGCTest()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("  堆大小: %s, GC次数: %d\n",
			formatBytes(security.MustSafeUint64ToInt64(m.HeapAlloc)), m.NumGC)
	}

	// 恢复原始设置
	debug.SetGCPercent(originalPercent)
	fmt.Printf("恢复GC目标为: %d%%\n", originalPercent)
}

func (mm *MemoryManager) allocateMemoryForGCTest() {
	// 分配一些内存来触发GC
	data := make([][]byte, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = make([]byte, 1024)
	}

	// 让GC有机会运行
	runtime.GC()

	// 释放引用
	data = nil
	runtime.GC()
}

// 内存泄漏检测示例
func (mm *MemoryManager) DetectMemoryLeak() {
	fmt.Println("内存泄漏检测演示...")

	// 记录开始时的内存状态
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	fmt.Printf("开始内存: %s\n", formatBytes(security.MustSafeUint64ToInt64(m1.HeapAlloc)))

	// 模拟潜在的内存泄漏
	leakyFunction()

	// 记录结束时的内存状态
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)
	fmt.Printf("结束内存: %s\n", formatBytes(security.MustSafeUint64ToInt64(m2.HeapAlloc)))

	// 检测泄漏
	leaked := security.MustSafeUint64ToInt64(m2.HeapAlloc) - security.MustSafeUint64ToInt64(m1.HeapAlloc)
	if leaked > 0 {
		fmt.Printf("⚠️ 可能的内存泄漏: %s\n", formatBytes(leaked))
	} else {
		fmt.Printf("✅ 未检测到内存泄漏\n")
	}
}

// 模拟内存泄漏的函数
var globalSlice [][]byte

func leakyFunction() {
	// 这会导致内存泄漏，因为数据保存在全局变量中
	for i := 0; i < 1000; i++ {
		data := make([]byte, 1024)
		globalSlice = append(globalSlice, data)
	}

	// 注意：在实际应用中，应该在适当的时候清理globalSlice
}

func demonstrateMemoryManagement() {
	fmt.Println("=== 4. 内存管理和GC基础 ===")

	mm := NewMemoryManager()

	// 创建对象池
	mm.CreatePool("buffer", func() interface{} {
		return make([]byte, 1024)
	})

	fmt.Println("对象池示例:")
	// 使用对象池
	buffer1 := mm.GetObject("buffer").([]byte)
	fmt.Printf("获取缓冲区: %T, 长度: %d\n", buffer1, len(buffer1))

	// 归还对象
	mm.PutObject("buffer", buffer1)
	fmt.Println("缓冲区已归还到对象池")

	// 再次获取（应该是同一个对象）
	buffer2 := mm.GetObject("buffer").([]byte)
	fmt.Printf("再次获取缓冲区: %T, 长度: %d\n", buffer2, len(buffer2))

	fmt.Printf("\n对象池统计:\n")
	fmt.Printf("  创建对象: %d\n", mm.metrics.ObjectsCreated)
	fmt.Printf("  重用对象: %d\n", mm.metrics.ObjectsReused)

	// GC调优演示
	mm.TriggerGCTuning()

	// 内存泄漏检测
	fmt.Println()
	mm.DetectMemoryLeak()

	// 启动内存监控
	fmt.Println()
	go mm.MonitorMemory(10 * time.Second)

	// 分配一些内存来观察监控效果
	for i := 0; i < 3; i++ {
		time.Sleep(3 * time.Second)
		data := make([][]byte, 1000)
		for j := 0; j < 1000; j++ {
			data[j] = make([]byte, 1024)
		}
		runtime.GC() // 手动触发GC
	}

	fmt.Println()
}

// ==================
// 5. 并发性能优化
// ==================

// ConcurrencyOptimizer 并发性能优化器
type ConcurrencyOptimizer struct {
	workers     int
	concurrency int
	metrics     ConcurrencyMetrics
}

type ConcurrencyMetrics struct {
	TasksProcessed    int64
	ProcessingTime    time.Duration
	GoroutineCount    int
	ChannelOperations int64
	LockContention    int64
}

func NewConcurrencyOptimizer(workers int) *ConcurrencyOptimizer {
	return &ConcurrencyOptimizer{
		workers:     workers,
		concurrency: workers,
	}
}

// 原子操作 vs 锁的性能对比
func (co *ConcurrencyOptimizer) BenchmarkAtomicVsLock() {
	fmt.Println("原子操作 vs 锁性能对比:")

	const iterations = 1000000
	const goroutines = 10

	// 测试原子操作
	var atomicCounter int64
	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations/goroutines; j++ {
				atomic.AddInt64(&atomicCounter, 1)
			}
		}()
	}
	wg.Wait()
	atomicTime := time.Since(start)

	// 测试互斥锁
	var lockCounter int64
	var mutex sync.Mutex
	start = time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations/goroutines; j++ {
				mutex.Lock()
				lockCounter++
				mutex.Unlock()
			}
		}()
	}
	wg.Wait()
	lockTime := time.Since(start)

	fmt.Printf("  原子操作: %v (最终值: %d)\n", atomicTime, atomicCounter)
	fmt.Printf("  互斥锁:   %v (最终值: %d)\n", lockTime, lockCounter)
	fmt.Printf("  性能比:   %.2fx\n", float64(lockTime)/float64(atomicTime))
}

// Channel vs 共享内存性能对比
func (co *ConcurrencyOptimizer) BenchmarkChannelVsSharedMemory() {
	fmt.Println("\nChannel vs 共享内存性能对比:")

	const dataSize = 100000
	data := make([]int, dataSize)
	for i := range data {
		data[i] = secureRandomInt(1000)
	}

	// 使用Channel传递数据
	start := time.Now()
	co.processDataWithChannel(data)
	channelTime := time.Since(start)

	// 使用共享内存
	start = time.Now()
	co.processDataWithSharedMemory(data)
	sharedMemoryTime := time.Since(start)

	fmt.Printf("  Channel方式: %v\n", channelTime)
	fmt.Printf("  共享内存:    %v\n", sharedMemoryTime)
	fmt.Printf("  性能比:      %.2fx\n", float64(channelTime)/float64(sharedMemoryTime))
}

func (co *ConcurrencyOptimizer) processDataWithChannel(data []int) {
	inputChan := make(chan int, 100)
	outputChan := make(chan int, 100)

	// 启动工作者
	var wg sync.WaitGroup
	for i := 0; i < co.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for num := range inputChan {
				// 模拟计算
				result := num * num
				outputChan <- result
			}
		}()
	}

	// 发送数据
	go func() {
		for _, num := range data {
			inputChan <- num
		}
		close(inputChan)
	}()

	// 收集结果
	go func() {
		wg.Wait()
		close(outputChan)
	}()

	count := 0
	for range outputChan {
		count++
	}
}

func (co *ConcurrencyOptimizer) processDataWithSharedMemory(data []int) {
	results := make([]int, len(data))
	var wg sync.WaitGroup
	var index int64

	// 启动工作者
	for i := 0; i < co.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := atomic.AddInt64(&index, 1) - 1
				if i >= int64(len(data)) {
					break
				}
				// 模拟计算
				results[i] = data[i] * data[i]
			}
		}()
	}

	wg.Wait()
}

// 工作池模式优化
func (co *ConcurrencyOptimizer) DemonstrateWorkerPoolOptimization() {
	fmt.Println("\n工作池优化示例:")

	tasks := make([]int, 10000)
	for i := range tasks {
		tasks[i] = i
	}

	// 测试不同工作者数量的性能
	workerCounts := []int{1, 2, 4, 8, 16}

	for _, workers := range workerCounts {
		start := time.Now()
		co.processWithWorkerPool(tasks, workers)
		duration := time.Since(start)
		fmt.Printf("  %2d workers: %v\n", workers, duration)
	}
}

func (co *ConcurrencyOptimizer) processWithWorkerPool(tasks []int, workers int) {
	taskChan := make(chan int, len(tasks))
	var wg sync.WaitGroup

	// 启动工作者
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				// 模拟CPU密集型工作
				result := 0
				for j := 0; j < task%1000+100; j++ {
					result += j * j
				}
				_ = result
			}
		}()
	}

	// 发送任务
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	wg.Wait()
}

func demonstrateConcurrencyOptimization() {
	fmt.Println("=== 5. 并发性能优化 ===")

	optimizer := NewConcurrencyOptimizer(4)

	// 原子操作vs锁性能对比
	optimizer.BenchmarkAtomicVsLock()

	// Channel vs共享内存对比
	optimizer.BenchmarkChannelVsSharedMemory()

	// 工作池优化
	optimizer.DemonstrateWorkerPoolOptimization()

	fmt.Println("\n并发性能优化建议:")
	fmt.Println("  1. 优先使用原子操作而非锁")
	fmt.Println("  2. 根据任务类型选择Channel或共享内存")
	fmt.Println("  3. 工作者数量通常等于CPU核心数")
	fmt.Println("  4. 避免过度的context切换")
	fmt.Println("  5. 使用缓冲channel减少阻塞")

	fmt.Println()
}

// ==================
// 6. 实际性能优化案例
// ==================

// OptimizationCase 性能优化案例
type OptimizationCase struct {
	name        string
	description string
	before      func()
	after       func()
	improvement float64
}

// 创建性能优化案例集合
func createOptimizationCases() []OptimizationCase {
	return []OptimizationCase{
		{
			name:        "字符串构建优化",
			description: "使用strings.Builder替代字符串拼接",
			before:      stringConcatBefore,
			after:       stringConcatAfter,
		},
		{
			name:        "切片预分配优化",
			description: "预分配切片容量避免多次扩容",
			before:      sliceAppendBefore,
			after:       sliceAppendAfter,
		},
		{
			name:        "Map预分配优化",
			description: "预分配Map容量提高性能",
			before:      mapAccessBefore,
			after:       mapAccessAfter,
		},
		{
			name:        "接口断言优化",
			description: "使用类型switch优化多重断言",
			before:      typeAssertionBefore,
			after:       typeAssertionAfter,
		},
	}
}

// 优化案例实现

// 字符串构建优化
func stringConcatBefore() {
	var result string
	for i := 0; i < 1000; i++ {
		result += fmt.Sprintf("item_%d,", i)
	}
}

func stringConcatAfter() {
	var builder strings.Builder
	builder.Grow(10000) // 预分配容量
	for i := 0; i < 1000; i++ {
		builder.WriteString(fmt.Sprintf("item_%d,", i))
	}
	result := builder.String()
	_ = result
}

// 切片预分配优化
func sliceAppendBefore() {
	var slice []int
	for i := 0; i < 1000; i++ {
		slice = append(slice, i)
	}
}

func sliceAppendAfter() {
	slice := make([]int, 0, 1000) // 预分配容量
	for i := 0; i < 1000; i++ {
		slice = append(slice, i)
	}
}

// Map访问优化
func mapAccessBefore() {
	m := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		m[key] = i
	}
}

func mapAccessAfter() {
	m := make(map[string]int, 1000) // 预分配容量
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key_%d", i)
		m[key] = i
	}
}

// 接口断言优化
func typeAssertionBefore() {
	values := []interface{}{1, "hello", 3.14, true, 42}

	for i := 0; i < 1000; i++ {
		for _, v := range values {
			// 多重if断言
			if _, ok := v.(int); ok {
				// 处理int
			} else if _, ok := v.(string); ok {
				// 处理string
			} else if _, ok := v.(float64); ok {
				// 处理float64
			} else if _, ok := v.(bool); ok {
				// 处理bool
			}
		}
	}
}

func typeAssertionAfter() {
	values := []interface{}{1, "hello", 3.14, true, 42}

	for i := 0; i < 1000; i++ {
		for _, v := range values {
			// 使用type switch
			switch v.(type) {
			case int:
				// 处理int
			case string:
				// 处理string
			case float64:
				// 处理float64
			case bool:
				// 处理bool
			}
		}
	}
}

func demonstrateOptimizationCases() {
	fmt.Println("=== 6. 实际性能优化案例 ===")

	cases := createOptimizationCases()

	for _, optimizationCase := range cases {
		fmt.Printf("\n案例: %s\n", optimizationCase.name)
		fmt.Printf("描述: %s\n", optimizationCase.description)

		// 测试优化前性能
		start := time.Now()
		for i := 0; i < 100; i++ {
			optimizationCase.before()
		}
		beforeTime := time.Since(start)

		// 测试优化后性能
		start = time.Now()
		for i := 0; i < 100; i++ {
			optimizationCase.after()
		}
		afterTime := time.Since(start)

		improvement := float64(beforeTime-afterTime) / float64(beforeTime) * 100
		fmt.Printf("优化前: %v\n", beforeTime)
		fmt.Printf("优化后: %v\n", afterTime)
		if improvement > 0 {
			fmt.Printf("性能提升: %.1f%%\n", improvement)
		} else {
			fmt.Printf("性能下降: %.1f%%\n", -improvement)
		}
	}

	fmt.Println("\n性能优化总结:")
	fmt.Println("✅ 预分配内存可显著提升性能")
	fmt.Println("✅ 选择合适的数据结构和算法")
	fmt.Println("✅ 避免不必要的内存分配")
	fmt.Println("✅ 使用性能分析工具验证优化效果")
	fmt.Println("✅ 在实际场景中测试优化结果")

	fmt.Println()
}

// ==================
// 主函数和综合演示
// ==================

func init() {
	// 启用pprof HTTP服务器用于实时性能分析
	go func() {
		log.Println("pprof server started at :6060")
		log.Println("访问 http://localhost:6060/debug/pprof/ 查看实时性能数据")
		server := &http.Server{
			Addr:         "localhost:6060",
			Handler:      nil,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		log.Println(server.ListenAndServe())
	}()
}

func main() {
	fmt.Println("🚀 Go语言性能优化基础：从应用开发到系统编程的桥梁")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("CPU核心数: %d\n", runtime.NumCPU())
	fmt.Printf("最大并发数: %d\n", runtime.GOMAXPROCS(0))
	fmt.Println()

	// 1. 性能分析理论基础
	demonstratePerformanceTheory()

	// 2. Go性能分析工具链
	demonstrateProfilingTools()

	// 3. 基准测试和性能测试
	demonstrateBenchmarking()

	// 4. 内存管理和GC基础
	demonstrateMemoryManagement()

	// 5. 并发性能优化
	demonstrateConcurrencyOptimization()

	// 6. 实际性能优化案例
	demonstrateOptimizationCases()

	fmt.Println("🎯 性能优化基础课程完成！")
	fmt.Println("你现在已经掌握了:")
	fmt.Println("✅ 性能分析的理论基础和核心概念")
	fmt.Println("✅ Go语言性能分析工具的使用方法")
	fmt.Println("✅ 基准测试和性能测试的实施技巧")
	fmt.Println("✅ 内存管理和垃圾收集的基础知识")
	fmt.Println("✅ 并发性能优化的策略和方法")
	fmt.Println("✅ 实际项目中的性能优化经验")
	fmt.Println()
	fmt.Println("🚀 现在你已经准备好深入学习系统级编程了！")
	fmt.Println("接下来的运行时内核模块将建立在这些基础之上。")
	fmt.Println()
	fmt.Println("💡 记住：性能优化是一个持续的过程")
	fmt.Println("   - 先测量，再优化")
	fmt.Println("   - 专注于瓶颈")
	fmt.Println("   - 验证优化效果")
	fmt.Println("   - 平衡可读性和性能")
}

/*
=== 练习题 ===

1. **性能分析实践**
   - 为你的一个项目添加性能分析功能
   - 使用pprof工具分析CPU和内存使用情况
   - 生成性能报告并识别瓶颈

2. **基准测试编写**
   - 为核心函数编写基准测试
   - 对比不同实现方案的性能差异
   - 建立性能回归检测机制

3. **内存优化项目**
   - 实现一个高效的对象池
   - 优化一个内存密集型应用
   - 设计内存泄漏检测工具

4. **并发优化实践**
   - 优化一个高并发应用的锁竞争
   - 实现不同并发模式的性能对比
   - 设计高效的工作池模式

5. **综合优化案例**
   - 选择一个实际项目进行全面性能优化
   - 建立性能监控和告警系统
   - 编写性能优化最佳实践文档

运行命令：
go run main.go

性能分析命令：
go tool pprof http://localhost:6060/debug/pprof/profile
go tool pprof http://localhost:6060/debug/pprof/heap

学习目标验证：
- 能够使用Go性能分析工具定位问题
- 掌握常见的性能优化技巧
- 理解内存管理和GC的基本原理
- 具备系统级编程的预备知识
- 能够设计高性能的Go应用程序

下一步学习方向：
- 07-runtime-internals: 深入Go运行时内核
- 08-performance-mastery: 高级性能调优技术
- 09-system-programming: 系统级编程技术
*/
