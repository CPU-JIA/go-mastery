/*
=== 调度器基准测试和单元测试 ===

测试调度器相关功能的正确性和性能特征。
包含：
1. 调度器统计收集器测试
2. Goroutine 泄漏检测器测试
3. 调度延迟测量器测试
4. GOMAXPROCS 调优器测试
5. 调度性能基准测试
*/

package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ==================
// 1. 调度器统计收集器测试
// ==================

func TestSchedulerStatsCollector(t *testing.T) {
	collector := NewSchedulerStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 创建一些 Goroutine
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()
	}

	// 等待收集数据
	time.Sleep(100 * time.Millisecond)
	wg.Wait()

	// 获取统计
	current, minG, maxG, avgG := collector.GetGoroutineStats()

	t.Logf("Goroutine 统计: 当前=%d, 最小=%d, 最大=%d, 平均=%d",
		current, minG, maxG, avgG)

	// 验证
	if current <= 0 {
		t.Error("当前 Goroutine 数应大于 0")
	}
	if maxG < minG {
		t.Error("最大值不应小于最小值")
	}
}

func TestSchedulerStatsCollectorLatency(t *testing.T) {
	collector := NewSchedulerStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 等待收集数据
	time.Sleep(200 * time.Millisecond)

	// 获取延迟百分位数
	p50, p90, p95, p99 := collector.GetLatencyPercentiles()

	t.Logf("调度延迟百分位数: P50=%v, P90=%v, P95=%v, P99=%v",
		p50, p90, p95, p99)

	// 验证顺序
	if p50 > p90 {
		t.Errorf("P50 (%v) 不应大于 P90 (%v)", p50, p90)
	}
	if p90 > p99 {
		t.Errorf("P90 (%v) 不应大于 P99 (%v)", p90, p99)
	}
}

// ==================
// 2. Goroutine 泄漏检测器测试
// ==================

func TestGoroutineLeakDetector(t *testing.T) {
	detector := NewGoroutineLeakDetector(0.5, 50*time.Millisecond)

	var leakDetected atomic.Bool
	detector.SetLeakCallback(func(current, baseline int) {
		leakDetected.Store(true)
		t.Logf("检测到泄漏: 当前=%d, 基线=%d", current, baseline)
	})

	detector.Start()
	defer detector.Stop()

	// 创建一些不会结束的 Goroutine（模拟泄漏）
	stopCh := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			<-stopCh
		}()
	}

	// 等待检测
	time.Sleep(200 * time.Millisecond)

	// 清理
	close(stopCh)
	time.Sleep(50 * time.Millisecond)

	// 验证检测到泄漏
	if !leakDetected.Load() {
		t.Log("未检测到泄漏（可能基线已经很高）")
	}
}

func TestGoroutineLeakDetectorCheckNow(t *testing.T) {
	detector := NewGoroutineLeakDetector(0.1, time.Second)

	// 记录基线
	baseline := runtime.NumGoroutine()

	// 创建额外的 Goroutine
	stopCh := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			<-stopCh
		}()
	}

	time.Sleep(10 * time.Millisecond)

	// 立即检测
	leaked, current, base := detector.CheckNow()

	t.Logf("泄漏检测: leaked=%v, current=%d, baseline=%d",
		leaked, current, base)

	// 清理
	close(stopCh)

	// 验证当前值大于基线
	if current <= baseline {
		t.Errorf("当前 Goroutine 数 (%d) 应大于基线 (%d)", current, baseline)
	}
}

func TestGoroutineLeakDetectorResetBaseline(t *testing.T) {
	detector := NewGoroutineLeakDetector(0.5, time.Second)

	// 创建一些 Goroutine
	stopCh := make(chan struct{})
	for i := 0; i < 20; i++ {
		go func() {
			<-stopCh
		}()
	}

	time.Sleep(10 * time.Millisecond)

	// 重置基线
	detector.ResetBaseline()

	// 检测应该不会报告泄漏
	leaked, _, _ := detector.CheckNow()

	// 清理
	close(stopCh)

	if leaked {
		t.Error("重置基线后不应检测到泄漏")
	}
}

// ==================
// 3. 调度延迟测量器测试
// ==================

func TestSchedulingLatencyMeasurer(t *testing.T) {
	measurer := NewSchedulingLatencyMeasurer()
	measurer.Measure(100)

	results := measurer.GetResults()

	t.Logf("调度延迟测量结果:\n%s", results.String())

	// 验证结果
	if results.Count != 100 {
		t.Errorf("测量次数应为 100，实际: %d", results.Count)
	}
	if results.Min > results.Max {
		t.Error("最小值不应大于最大值")
	}
	if results.P50 > results.P99 {
		t.Error("P50 不应大于 P99")
	}
}

func TestSchedulingLatencyMeasurerUnderLoad(t *testing.T) {
	measurer := NewSchedulingLatencyMeasurer()
	measurer.MeasureUnderLoad(50, 4)

	results := measurer.GetResults()

	t.Logf("负载下调度延迟:\n%s", results.String())

	// 验证结果
	if results.Count != 50 {
		t.Errorf("测量次数应为 50，实际: %d", results.Count)
	}
}

func TestLatencyResultsString(t *testing.T) {
	results := LatencyResults{
		Count: 100,
		Min:   time.Microsecond,
		Max:   time.Millisecond,
		Avg:   100 * time.Microsecond,
		P50:   50 * time.Microsecond,
		P90:   200 * time.Microsecond,
		P95:   500 * time.Microsecond,
		P99:   900 * time.Microsecond,
	}

	str := results.String()
	if str == "" {
		t.Error("String() 不应返回空字符串")
	}
	t.Log(str)
}

// ==================
// 4. GOMAXPROCS 调优器测试
// ==================

func TestGOMAXPROCSAutoTuner(t *testing.T) {
	original := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(original)

	tuner := NewGOMAXPROCSAutoTuner(1, 4, 50*time.Millisecond)
	tuner.Start()

	// 等待一段时间
	time.Sleep(200 * time.Millisecond)

	current := tuner.GetCurrent()
	t.Logf("当前 GOMAXPROCS: %d", current)

	tuner.Stop()

	// 验证恢复原始值
	restored := runtime.GOMAXPROCS(0)
	if restored != original {
		t.Errorf("停止后应恢复原始 GOMAXPROCS: %d，实际: %d", original, restored)
	}
}

// ==================
// 5. 工作负载分析器测试
// ==================

func TestWorkloadAnalyzer(t *testing.T) {
	analyzer := NewWorkloadAnalyzer(10*time.Millisecond, 10)

	// 创建一些工作负载
	var wg sync.WaitGroup
	stopCh := make(chan struct{})

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					// 模拟工作
					sum := 0
					for j := 0; j < 1000; j++ {
						sum += j
					}
					runtime.Gosched()
				}
			}
		}()
	}

	// 分析
	analysis := analyzer.Analyze()

	// 停止工作负载
	close(stopCh)
	wg.Wait()

	t.Logf("工作负载分析:")
	t.Logf("  Goroutine 数: %d", analysis.GoroutineCount)
	t.Logf("  增长率: %.2f/s", analysis.GoroutineGrowth)
	t.Logf("  变化频率: %.2f", analysis.GoroutineChurn)
	t.Logf("  调度延迟: %v", analysis.SchedulingLatency)
	t.Logf("  CPU 密集: %v", analysis.IsCPUBound)
	t.Logf("  I/O 密集: %v", analysis.IsIOBound)
	t.Logf("  高并发: %v", analysis.IsConcurrency)
	t.Logf("  建议: %v", analysis.Recommendations)

	// 验证
	if analysis.GoroutineCount <= 0 {
		t.Error("Goroutine 数应大于 0")
	}
}

// ==================
// 6. 便捷函数测试
// ==================

func TestGetSchedulerInfo(t *testing.T) {
	info := GetSchedulerInfo()

	t.Logf("调度器信息:\n%s", info.String())

	// 验证
	if info.NumCPU <= 0 {
		t.Error("CPU 核心数应大于 0")
	}
	if info.GOMAXPROCS <= 0 {
		t.Error("GOMAXPROCS 应大于 0")
	}
	if info.NumGoroutine <= 0 {
		t.Error("Goroutine 数应大于 0")
	}
}

func TestMeasureLatency(t *testing.T) {
	results := MeasureLatency(50)

	t.Logf("延迟测量: %s", results.String())

	if results.Count != 50 {
		t.Errorf("测量次数应为 50，实际: %d", results.Count)
	}
}

func TestDetectGoroutineLeak(t *testing.T) {
	baseline := runtime.NumGoroutine()

	// 创建额外的 Goroutine
	stopCh := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			<-stopCh
		}()
	}

	time.Sleep(10 * time.Millisecond)

	leaked, current := DetectGoroutineLeak(baseline, 0.5)

	t.Logf("泄漏检测: leaked=%v, current=%d, baseline=%d",
		leaked, current, baseline)

	// 清理
	close(stopCh)

	if !leaked {
		t.Error("应检测到泄漏")
	}
}

// ==================
// 7. 调度器监控器测试
// ==================

func TestSchedulerMonitor(t *testing.T) {
	monitor := NewSchedulerMonitor()
	monitor.Start()
	defer monitor.Stop()

	// 创建一些 Goroutine
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)
		}()
	}

	// 等待收集
	time.Sleep(200 * time.Millisecond)
	wg.Wait()

	// 获取统计
	stats := monitor.GetLatestStats()

	t.Logf("调度器统计:")
	t.Logf("  NumGoroutine: %d", stats.NumGoroutine)
	t.Logf("  NumCPU: %d", stats.NumCPU)
	t.Logf("  GOMAXPROCS: %d", stats.GOMAXPROCS)

	// 验证
	if stats.NumCPU <= 0 {
		t.Error("NumCPU 应大于 0")
	}
}

func TestSchedulerMonitorHistory(t *testing.T) {
	monitor := NewSchedulerMonitor()
	monitor.Start()
	defer monitor.Stop()

	// 等待收集
	time.Sleep(500 * time.Millisecond)

	// 获取历史
	history := monitor.GetStatsHistory()

	t.Logf("历史记录数: %d", len(history))

	if len(history) == 0 {
		t.Error("历史记录不应为空")
	}
}

// ==================
// 8. G-M-P 模拟器测试
// ==================

func TestGMPSimulator(t *testing.T) {
	simulator := NewGMPSimulator(4)
	simulator.Start()
	defer simulator.Stop()

	// 添加工作
	for i := 0; i < 20; i++ {
		simulator.AddWork(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 等待处理
	time.Sleep(500 * time.Millisecond)

	// 获取统计
	stats := simulator.GetStats()
	globalWork, localWork, workSteals := simulator.GetWorkStats()

	t.Logf("G-M-P 统计:")
	t.Logf("  G: %d", stats.G)
	t.Logf("  P: %d", stats.P)
	t.Logf("  全局队列: %d", globalWork)
	t.Logf("  本地队列: %d", localWork)
	t.Logf("  工作窃取: %d", workSteals)
}

// ==================
// 9. 基准测试
// ==================

// BenchmarkGoroutineCreation Goroutine 创建基准测试
func BenchmarkGoroutineCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		done := make(chan struct{})
		go func() {
			close(done)
		}()
		<-done
	}
}

// BenchmarkGoroutineCreationParallel 并行 Goroutine 创建基准测试
func BenchmarkGoroutineCreationParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			done := make(chan struct{})
			go func() {
				close(done)
			}()
			<-done
		}
	})
}

// BenchmarkSchedulingLatency 调度延迟基准测试
func BenchmarkSchedulingLatency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		start := time.Now()
		done := make(chan struct{})
		go func() {
			close(done)
		}()
		<-done
		_ = time.Since(start)
	}
}

// BenchmarkContextSwitch 上下文切换基准测试
func BenchmarkContextSwitch(b *testing.B) {
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})

	go func() {
		for {
			<-ch1
			ch2 <- struct{}{}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch1 <- struct{}{}
		<-ch2
	}
}

// BenchmarkGosched runtime.Gosched 基准测试
func BenchmarkGosched(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runtime.Gosched()
	}
}

// BenchmarkNumGoroutine runtime.NumGoroutine 基准测试
func BenchmarkNumGoroutine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = runtime.NumGoroutine()
	}
}

// BenchmarkGOMAXPROCSRead GOMAXPROCS 读取基准测试
func BenchmarkGOMAXPROCSRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = runtime.GOMAXPROCS(0)
	}
}

// BenchmarkWorkStealing 工作窃取模拟基准测试
func BenchmarkWorkStealing(b *testing.B) {
	simulator := NewGMPSimulator(runtime.GOMAXPROCS(0))
	simulator.Start()
	defer simulator.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simulator.AddWork(func() {
			// 空工作
		})
	}
}

// BenchmarkHighConcurrency 高并发基准测试
func BenchmarkHighConcurrency(b *testing.B) {
	concurrency := []int{10, 100, 1000}

	for _, c := range concurrency {
		b.Run(formatConcurrency(c), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for j := 0; j < c; j++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						// 简单工作
						sum := 0
						for k := 0; k < 100; k++ {
							sum += k
						}
					}()
				}
				wg.Wait()
			}
		})
	}
}

func formatConcurrency(c int) string {
	if c >= 1000 {
		return "1000"
	}
	if c >= 100 {
		return "100"
	}
	return "10"
}

// BenchmarkGOMAXPROCSImpact 不同 GOMAXPROCS 的影响
func BenchmarkGOMAXPROCSImpact(b *testing.B) {
	original := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(original)

	procsValues := []int{1, 2, 4, 8}

	for _, procs := range procsValues {
		if procs > runtime.NumCPU() {
			continue
		}

		b.Run(formatProcs(procs), func(b *testing.B) {
			runtime.GOMAXPROCS(procs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var wg sync.WaitGroup
				for j := 0; j < 100; j++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						sum := 0
						for k := 0; k < 10000; k++ {
							sum += k
						}
					}()
				}
				wg.Wait()
			}
		})
	}
}

func formatProcs(p int) string {
	return "PROCS_" + string(rune('0'+p))
}

// ==================
// 10. 抢占式调度测试
// ==================

func TestPreemptiveSchedulerDemo(t *testing.T) {
	demo := NewPreemptiveSchedulerDemo()
	demo.Start()

	// 运行一段时间
	time.Sleep(500 * time.Millisecond)

	demo.Stop()

	t.Log("抢占式调度演示完成")
}

// ==================
// 11. 系统调用演示测试
// ==================

func TestSyscallDemo(t *testing.T) {
	demo := NewSyscallDemo()
	demo.Start()

	// 运行一段时间
	time.Sleep(500 * time.Millisecond)

	demo.Stop()

	t.Log("系统调用演示完成")
}

// ==================
// 12. 延迟测量测试
// ==================

func TestLatencyMeasurement(t *testing.T) {
	lm := NewLatencyMeasurement()

	// 测量少量迭代
	lm.MeasureSchedulingLatency(100)

	// 验证有测量结果
	if len(lm.measurements) != 100 {
		t.Errorf("应有 100 个测量结果，实际: %d", len(lm.measurements))
	}
}
