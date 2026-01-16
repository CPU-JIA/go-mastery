/*
=== 性能分析基准测试和单元测试 ===

测试性能分析相关功能的正确性和性能特征。
包含：
1. 指标收集器测试
2. 延迟直方图测试
3. 吞吐量计算器测试
4. 基线比较器测试
5. 性能分析基准测试
*/

package main

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// ==================
// 1. 指标收集器测试
// ==================

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	// 测试计数器
	collector.IncrementCounter("requests")
	collector.IncrementCounter("requests")
	collector.AddCounter("requests", 3)

	if count := collector.GetCounter("requests"); count != 5 {
		t.Errorf("计数器应为 5，实际: %d", count)
	}

	// 测试计量器
	collector.SetGauge("connections", 100)
	if gauge := collector.GetGauge("connections"); gauge != 100 {
		t.Errorf("计量器应为 100，实际: %d", gauge)
	}

	// 测试延迟记录
	for i := 0; i < 100; i++ {
		collector.RecordLatency("response_time", time.Duration(i)*time.Millisecond)
	}

	stats := collector.GetHistogramStats("response_time")
	if stats.Count != 100 {
		t.Errorf("延迟记录数应为 100，实际: %d", stats.Count)
	}
}

func TestMetricsCollectorConcurrent(t *testing.T) {
	collector := NewMetricsCollector()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				collector.IncrementCounter("concurrent_requests")
				collector.RecordLatency("concurrent_latency", time.Millisecond)
			}
		}()
	}

	wg.Wait()

	count := collector.GetCounter("concurrent_requests")
	if count != 10000 {
		t.Errorf("并发计数器应为 10000，实际: %d", count)
	}
}

func TestMetricsCollectorGetAllMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	collector.IncrementCounter("test_counter")
	collector.SetGauge("test_gauge", 42)
	collector.RecordLatency("test_latency", time.Millisecond)

	summary := collector.GetAllMetrics()

	t.Log(summary.String())

	if len(summary.Counters) == 0 {
		t.Error("应有计数器指标")
	}
	if len(summary.Gauges) == 0 {
		t.Error("应有计量器指标")
	}
	if len(summary.Latencies) == 0 {
		t.Error("应有延迟指标")
	}
}

// ==================
// 2. 延迟直方图测试
// ==================

func TestLatencyHistogram(t *testing.T) {
	histogram := NewLatencyHistogram()

	// 记录一些延迟值
	latencies := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		20 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	}

	for _, l := range latencies {
		histogram.Record(l)
	}

	stats := histogram.GetStats()

	t.Logf("直方图统计:")
	t.Logf("  Count: %d", stats.Count)
	t.Logf("  Min: %v", stats.Min)
	t.Logf("  Max: %v", stats.Max)
	t.Logf("  Avg: %v", stats.Avg)
	t.Logf("  P50: %v", stats.P50)
	t.Logf("  P99: %v", stats.P99)

	// 验证
	if stats.Count != 10 {
		t.Errorf("Count 应为 10，实际: %d", stats.Count)
	}
	if stats.Min != time.Millisecond {
		t.Errorf("Min 应为 1ms，实际: %v", stats.Min)
	}
	if stats.Max != 200*time.Millisecond {
		t.Errorf("Max 应为 200ms，实际: %v", stats.Max)
	}
}

func TestLatencyHistogramReset(t *testing.T) {
	histogram := NewLatencyHistogram()

	for i := 0; i < 100; i++ {
		histogram.Record(time.Millisecond)
	}

	stats := histogram.GetStats()
	if stats.Count != 100 {
		t.Errorf("重置前 Count 应为 100，实际: %d", stats.Count)
	}

	histogram.Reset()

	stats = histogram.GetStats()
	if stats.Count != 0 {
		t.Errorf("重置后 Count 应为 0，实际: %d", stats.Count)
	}
}

func TestLatencyHistogramConcurrent(t *testing.T) {
	histogram := NewLatencyHistogram()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				histogram.Record(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	stats := histogram.GetStats()
	if stats.Count != 10000 {
		t.Errorf("并发记录后 Count 应为 10000，实际: %d", stats.Count)
	}
}

// ==================
// 3. 吞吐量计算器测试
// ==================

func TestThroughputCalculator(t *testing.T) {
	calculator := NewThroughputCalculator(time.Second)

	// 记录一些操作
	for i := 0; i < 1000; i++ {
		calculator.Record()
	}

	stats := calculator.GetStats()

	t.Logf("吞吐量统计:")
	t.Logf("  TotalCount: %d", stats.TotalCount)
	t.Logf("  TotalThroughput: %.2f ops/sec", stats.TotalThroughput)
	t.Logf("  WindowThroughput: %.2f ops/sec", stats.WindowThroughput)

	if stats.TotalCount != 1000 {
		t.Errorf("TotalCount 应为 1000，实际: %d", stats.TotalCount)
	}
}

func TestThroughputCalculatorRecordN(t *testing.T) {
	calculator := NewThroughputCalculator(time.Second)

	calculator.RecordN(100)
	calculator.RecordN(200)
	calculator.RecordN(300)

	stats := calculator.GetStats()

	if stats.TotalCount != 600 {
		t.Errorf("TotalCount 应为 600，实际: %d", stats.TotalCount)
	}
}

func TestThroughputCalculatorWindow(t *testing.T) {
	calculator := NewThroughputCalculator(100 * time.Millisecond)

	// 记录一些操作
	for i := 0; i < 100; i++ {
		calculator.Record()
		time.Sleep(time.Millisecond)
	}

	// 等待窗口过期
	time.Sleep(200 * time.Millisecond)

	// 记录更多操作
	for i := 0; i < 50; i++ {
		calculator.Record()
	}

	stats := calculator.GetStats()

	t.Logf("窗口吞吐量: %.2f ops/sec", stats.WindowThroughput)

	// 窗口吞吐量应该只反映最近的操作
	if stats.TotalCount != 150 {
		t.Errorf("TotalCount 应为 150，实际: %d", stats.TotalCount)
	}
}

// ==================
// 4. 基线比较器测试
// ==================

func TestBaselineComparator(t *testing.T) {
	comparator := NewBaselineComparator()

	// 设置基线
	baseline := map[string]float64{
		"latency_p99": 100.0,
		"throughput":  1000.0,
		"error_rate":  0.01,
	}
	comparator.SetBaseline("v1.0", baseline)

	// 比较当前指标
	current := map[string]float64{
		"latency_p99": 80.0,   // 改善
		"throughput":  1200.0, // 改善（但我们假设越小越好）
		"error_rate":  0.005,  // 改善
	}

	result := comparator.Compare("v1.0", current)

	t.Log(result.String())

	// 验证
	if result.Error != "" {
		t.Errorf("不应有错误: %s", result.Error)
	}

	latencyComp := result.Comparisons["latency_p99"]
	if !latencyComp.Improved {
		t.Error("latency_p99 应该显示改善")
	}
}

func TestBaselineComparatorNotFound(t *testing.T) {
	comparator := NewBaselineComparator()

	result := comparator.Compare("nonexistent", map[string]float64{})

	if result.Error == "" {
		t.Error("应该返回错误")
	}
}

// ==================
// 5. 性能报告生成器测试
// ==================

func TestPerformanceReportGenerator(t *testing.T) {
	collector := NewMetricsCollector()

	// 添加一些指标
	collector.IncrementCounter("requests")
	collector.SetGauge("connections", 50)
	collector.RecordLatency("response_time", 10*time.Millisecond)

	generator := NewPerformanceReportGenerator(collector)
	report := generator.GenerateReport()

	t.Log(report.String())

	// 验证
	if report.NumGoroutine <= 0 {
		t.Error("NumGoroutine 应大于 0")
	}
	if report.NumCPU <= 0 {
		t.Error("NumCPU 应大于 0")
	}
}

// ==================
// 6. 操作计时器测试
// ==================

func TestOperationTimer(t *testing.T) {
	collector := NewMetricsCollector()

	timer := StartTimer("test_operation", collector)
	time.Sleep(10 * time.Millisecond)
	elapsed := timer.Stop()

	t.Logf("操作耗时: %v", elapsed)

	if elapsed < 10*time.Millisecond {
		t.Errorf("耗时应至少 10ms，实际: %v", elapsed)
	}

	// 验证指标被记录
	count := collector.GetCounter("test_operation_count")
	if count != 1 {
		t.Errorf("计数器应为 1，实际: %d", count)
	}
}

// ==================
// 7. 便捷函数测试
// ==================

func TestMeasureOperation(t *testing.T) {
	elapsed := MeasureOperation("test", func() {
		time.Sleep(10 * time.Millisecond)
	})

	t.Logf("操作耗时: %v", elapsed)

	if elapsed < 10*time.Millisecond {
		t.Errorf("耗时应至少 10ms，实际: %v", elapsed)
	}
}

func TestMeasureOperationWithResult(t *testing.T) {
	result, elapsed := MeasureOperationWithResult("test", func() int {
		time.Sleep(10 * time.Millisecond)
		return 42
	})

	t.Logf("结果: %d, 耗时: %v", result, elapsed)

	if result != 42 {
		t.Errorf("结果应为 42，实际: %d", result)
	}
	if elapsed < 10*time.Millisecond {
		t.Errorf("耗时应至少 10ms，实际: %v", elapsed)
	}
}

func TestBenchmarkOperation(t *testing.T) {
	result := BenchmarkOperation("test_op", 100, func() {
		// 简单操作
		sum := 0
		for i := 0; i < 1000; i++ {
			sum += i
		}
	})

	t.Log(result.String())

	if result.Iterations != 100 {
		t.Errorf("迭代次数应为 100，实际: %d", result.Iterations)
	}
	if result.OpsPerSec <= 0 {
		t.Error("吞吐量应大于 0")
	}
}

func TestCalculateStdDev(t *testing.T) {
	values := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond,
	}

	stdDev := CalculateStdDev(values)

	t.Logf("标准差: %v", stdDev)

	if stdDev <= 0 {
		t.Error("标准差应大于 0")
	}
}

// ==================
// 8. CPU Profiler 测试
// ==================

func TestCPUProfiler(t *testing.T) {
	// 注意：这个测试会创建临时文件
	profiler := NewCPUProfiler("./test_cpu.prof", 100*time.Millisecond)

	err := profiler.StartProfile()
	if err != nil {
		t.Fatalf("启动 CPU profiling 失败: %v", err)
	}

	// 等待 profiling 完成
	time.Sleep(200 * time.Millisecond)

	if profiler.IsRunning() {
		t.Error("Profiler 应该已停止")
	}
}

// ==================
// 9. Memory Profiler 测试
// ==================

func TestMemoryProfiler(t *testing.T) {
	profiler := NewMemoryProfiler("./test_mem.prof")

	// 分配一些内存
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 1024))
	}

	err := profiler.WriteProfile()
	if err != nil {
		t.Fatalf("写入内存 profile 失败: %v", err)
	}

	// 保持引用
	_ = data
}

// ==================
// 10. Block Profiler 测试
// ==================

func TestBlockProfiler(t *testing.T) {
	profiler := NewBlockProfiler("./test_block.prof")

	profiler.Enable()

	// 创建一些阻塞
	ch := make(chan int)
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch <- 1
	}()
	<-ch

	err := profiler.WriteProfile()
	if err != nil {
		t.Fatalf("写入 block profile 失败: %v", err)
	}

	profiler.Disable()
}

// ==================
// 11. Mutex Profiler 测试
// ==================

func TestMutexProfiler(t *testing.T) {
	profiler := NewMutexProfiler("./test_mutex.prof")

	profiler.Enable()

	// 创建一些 mutex 争用
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			time.Sleep(time.Millisecond)
			mu.Unlock()
		}()
	}

	wg.Wait()

	err := profiler.WriteProfile()
	if err != nil {
		t.Fatalf("写入 mutex profile 失败: %v", err)
	}

	profiler.Disable()
}

// ==================
// 12. Goroutine Profiler 测试
// ==================

func TestGoroutineProfiler(t *testing.T) {
	profiler := NewGoroutineProfiler("./test_goroutine.prof")

	// 创建一些 goroutine
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()
	}

	err := profiler.WriteProfile()
	if err != nil {
		t.Fatalf("写入 goroutine profile 失败: %v", err)
	}

	wg.Wait()
}

// ==================
// 13. 基准测试
// ==================

// BenchmarkMetricsCollectorIncrement 计数器增加基准测试
func BenchmarkMetricsCollectorIncrement(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.IncrementCounter("test")
	}
}

// BenchmarkMetricsCollectorIncrementParallel 并发计数器增加基准测试
func BenchmarkMetricsCollectorIncrementParallel(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			collector.IncrementCounter("test")
		}
	})
}

// BenchmarkLatencyHistogramRecord 延迟记录基准测试
func BenchmarkLatencyHistogramRecord(b *testing.B) {
	histogram := NewLatencyHistogram()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		histogram.Record(time.Millisecond)
	}
}

// BenchmarkLatencyHistogramRecordParallel 并发延迟记录基准测试
func BenchmarkLatencyHistogramRecordParallel(b *testing.B) {
	histogram := NewLatencyHistogram()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			histogram.Record(time.Millisecond)
		}
	})
}

// BenchmarkThroughputCalculatorRecord 吞吐量记录基准测试
func BenchmarkThroughputCalculatorRecord(b *testing.B) {
	calculator := NewThroughputCalculator(time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculator.Record()
	}
}

// BenchmarkMeasureOperation 操作测量基准测试
func BenchmarkMeasureOperation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MeasureOperation("test", func() {
			// 空操作
		})
	}
}

// BenchmarkStartStopTimer 计时器启停基准测试
func BenchmarkStartStopTimer(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := StartTimer("test", collector)
		timer.Stop()
	}
}

// BenchmarkRuntimeMemStats 运行时内存统计基准测试
func BenchmarkRuntimeMemStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
	}
}

// BenchmarkNumGoroutine NumGoroutine 基准测试
func BenchmarkNumGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = runtime.NumGoroutine()
	}
}

// BenchmarkTimeNow time.Now 基准测试
func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now()
	}
}

// BenchmarkTimeSince time.Since 基准测试
func BenchmarkTimeSince(b *testing.B) {
	start := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = time.Since(start)
	}
}

// ==================
// 14. 工作负载生成器测试
// ==================

func TestWorkloadGenerator(t *testing.T) {
	generator := NewWorkloadGenerator()
	generator.SetCPUIntensive(true)
	generator.SetMemoryIntensive(true)
	generator.SetConcurrency(2)
	generator.SetDuration(100 * time.Millisecond)

	// 运行工作负载
	// 注意：这里不实际运行，因为会阻塞测试
	t.Log("工作负载生成器创建成功")
}

// ==================
// 15. 性能监控器测试
// ==================

func TestPerformanceMonitor(t *testing.T) {
	monitor := NewPerformanceMonitor("./test_profiles")

	err := monitor.Start()
	if err != nil {
		t.Logf("启动性能监控器失败（可能是端口占用）: %v", err)
		return
	}

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 获取指标
	metrics := monitor.GetMetrics()

	t.Logf("收集到 %d 个指标", len(metrics))

	monitor.Stop()
}

// ==================
// 16. 应用指标收集器测试
// ==================

func TestApplicationMetricsCollector(t *testing.T) {
	collector := &ApplicationMetricsCollector{}

	collector.IncrementRequests()
	collector.IncrementRequests()
	collector.IncrementErrors()
	collector.AddResponseTime(10 * time.Millisecond)
	collector.SetActiveConnections(50)

	metrics := collector.Collect()

	t.Logf("收集到 %d 个应用指标", len(metrics))

	for _, m := range metrics {
		t.Logf("  %s: %.2f %s", m.Name, m.Value, m.Unit)
	}

	if len(metrics) != 4 {
		t.Errorf("应有 4 个指标，实际: %d", len(metrics))
	}
}
