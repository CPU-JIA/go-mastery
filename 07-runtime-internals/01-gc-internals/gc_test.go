/*
=== GC 基准测试和单元测试 ===

测试 GC 相关功能的正确性和性能特征。
包含：
1. GC 统计收集器测试
2. GC 压力检测器测试
3. GC 调优助手测试
4. 内存分配模式基准测试
5. GC 暂停时间基准测试
*/

package main

import (
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"
)

// ==================
// 1. GC 统计收集器测试
// ==================

func TestGCStatsCollector(t *testing.T) {
	collector := NewGCStatsCollector(50*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 触发一些 GC
	for i := 0; i < 5; i++ {
		// 分配内存触发 GC
		_ = make([]byte, 10*1024*1024) // 10MB
		runtime.GC()
	}

	// 等待收集器收集数据
	time.Sleep(200 * time.Millisecond)

	// 验证百分位数计算
	p50, p90, p95, p99 := collector.GetPausePercentiles()

	// 暂停时间应该是正数（如果有 GC 发生）
	t.Logf("GC 暂停百分位数: P50=%v, P90=%v, P95=%v, P99=%v", p50, p90, p95, p99)

	// P50 应该小于等于 P90
	if p50 > p90 {
		t.Errorf("P50 (%v) 不应大于 P90 (%v)", p50, p90)
	}

	// P90 应该小于等于 P99
	if p90 > p99 {
		t.Errorf("P90 (%v) 不应大于 P99 (%v)", p90, p99)
	}
}

func TestGCStatsCollectorHeapGrowth(t *testing.T) {
	collector := NewGCStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 持续分配内存
	var data [][]byte
	for i := 0; i < 10; i++ {
		data = append(data, make([]byte, 1024*1024)) // 1MB
		time.Sleep(20 * time.Millisecond)
	}

	// 获取增长率
	rate := collector.GetHeapGrowthRate()
	t.Logf("堆增长率: %.2f bytes/sec", rate)

	// 保持引用防止 GC
	_ = data
}

// ==================
// 2. GC 压力检测器测试
// ==================

func TestGCPressureDetector(t *testing.T) {
	detector := NewGCPressureDetector(time.Second)

	// 第一次检测建立基线
	_ = detector.Detect()

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 第二次检测
	report := detector.Detect()

	t.Logf("GC 压力报告:")
	t.Logf("  级别: %s", report.Level)
	t.Logf("  频率: %.2f 次/秒", report.GCFrequency)
	t.Logf("  CPU 占用: %.2f%%", report.GCCPUPercent)
	t.Logf("  堆使用率: %.2f%%", report.HeapUsageRatio*100)
	t.Logf("  建议: %v", report.Recommendations)

	// 验证报告字段有效
	if report.HeapUsageRatio < 0 || report.HeapUsageRatio > 1 {
		t.Errorf("堆使用率应在 0-1 之间，实际: %f", report.HeapUsageRatio)
	}
}

func TestGCPressureLevelString(t *testing.T) {
	tests := []struct {
		level    GCPressureLevel
		expected string
	}{
		{GCPressureLow, "低"},
		{GCPressureMedium, "中等"},
		{GCPressureHigh, "高"},
		{GCPressureCritical, "临界"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("GCPressureLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

// ==================
// 3. GC 调优助手测试
// ==================

func TestGCTuningHelper(t *testing.T) {
	helper := NewGCTuningHelper()

	// 记录原始 GOGC
	originalGOGC := debug.SetGCPercent(-1)
	debug.SetGCPercent(originalGOGC)

	// 测试延迟优化
	helper.TuneForLatency()
	currentGOGC := debug.SetGCPercent(-1)
	debug.SetGCPercent(currentGOGC)

	if currentGOGC != 50 {
		t.Errorf("延迟优化后 GOGC 应为 50，实际: %d", currentGOGC)
	}

	// 恢复
	helper.Restore()
	restoredGOGC := debug.SetGCPercent(-1)
	debug.SetGCPercent(restoredGOGC)

	if restoredGOGC != originalGOGC {
		t.Errorf("恢复后 GOGC 应为 %d，实际: %d", originalGOGC, restoredGOGC)
	}
}

func TestGCTuningHelperThroughput(t *testing.T) {
	helper := NewGCTuningHelper()
	defer helper.Restore()

	// 测试吞吐量优化
	helper.TuneForThroughput()
	currentGOGC := debug.SetGCPercent(-1)
	debug.SetGCPercent(currentGOGC)

	if currentGOGC != 200 {
		t.Errorf("吞吐量优化后 GOGC 应为 200，实际: %d", currentGOGC)
	}
}

// ==================
// 4. 内存分配追踪器测试
// ==================

func TestAllocationTracker(t *testing.T) {
	tracker := NewAllocationTracker()

	// 进行一些分配
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 1024))
	}

	// 获取快照
	snapshot := tracker.TakeSnapshot()

	t.Logf("分配快照:")
	t.Logf("  堆对象: %d", snapshot.HeapObjects)
	t.Logf("  堆分配: %d bytes", snapshot.HeapAlloc)

	// 验证快照有效
	if snapshot.HeapObjects == 0 {
		t.Error("堆对象数不应为 0")
	}

	// 保持引用
	_ = data
}

func TestAllocationRate(t *testing.T) {
	tracker := NewAllocationTracker()

	// 进行分配
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 10*1024)) // 10KB
		time.Sleep(time.Millisecond)
	}

	rate := tracker.GetAllocationRate()
	t.Logf("分配速率: %.2f bytes/sec", rate)

	// 速率应该是正数
	if rate <= 0 {
		t.Error("分配速率应为正数")
	}

	_ = data
}

// ==================
// 5. 便捷函数测试
// ==================

func TestGetGCStats(t *testing.T) {
	// 触发 GC
	runtime.GC()

	stats := GetGCStats()

	t.Logf("GC 统计:")
	t.Logf("  NumGC: %d", stats.NumGC)
	t.Logf("  HeapSize: %d", stats.HeapSize)
	t.Logf("  HeapUsed: %d", stats.HeapUsed)
	t.Logf("  GCPercent: %d", stats.GCPercent)

	// 验证基本字段
	if stats.NumGC == 0 {
		t.Error("GC 次数不应为 0（已触发 GC）")
	}

	if stats.HeapSize == 0 {
		t.Error("堆大小不应为 0")
	}
}

func TestForceGC(t *testing.T) {
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	// 分配一些内存
	data := make([]byte, 10*1024*1024) // 10MB
	_ = data
	data = nil

	// 强制 GC
	ForceGC()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	// GC 次数应该增加
	if after.NumGC <= before.NumGC {
		t.Error("ForceGC 后 GC 次数应该增加")
	}
}

// ==================
// 6. 基准测试
// ==================

// BenchmarkSmallAllocation 小对象分配基准测试
func BenchmarkSmallAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 64)
	}
}

// BenchmarkMediumAllocation 中等对象分配基准测试
func BenchmarkMediumAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

// BenchmarkLargeAllocation 大对象分配基准测试
func BenchmarkLargeAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 32*1024)
	}
}

// BenchmarkHugeAllocation 超大对象分配基准测试
func BenchmarkHugeAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024*1024)
	}
}

// BenchmarkGCPause GC 暂停时间基准测试
func BenchmarkGCPause(b *testing.B) {
	// 预分配一些对象
	var data [][]byte
	for i := 0; i < 10000; i++ {
		data = append(data, make([]byte, 1024))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.GC()
	}

	// 保持引用
	_ = data
}

// BenchmarkConcurrentAllocation 并发分配基准测试
func BenchmarkConcurrentAllocation(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = make([]byte, 1024)
		}
	})
}

// BenchmarkPoolVsAllocation 对象池 vs 直接分配对比
func BenchmarkPoolVsAllocation(b *testing.B) {
	pool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	b.Run("DirectAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data := make([]byte, 1024)
			_ = data
		}
	})

	b.Run("PoolAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data := pool.Get().([]byte)
			pool.Put(data)
		}
	})
}

// BenchmarkGOGCImpact 不同 GOGC 值的影响
func BenchmarkGOGCImpact(b *testing.B) {
	gogcValues := []int{50, 100, 200, 400}

	for _, gogc := range gogcValues {
		b.Run(
			"GOGC_"+string(rune('0'+gogc/100))+string(rune('0'+(gogc%100)/10))+string(rune('0'+gogc%10)),
			func(b *testing.B) {
				oldGOGC := debug.SetGCPercent(gogc)
				defer debug.SetGCPercent(oldGOGC)

				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					// 分配和释放
					data := make([]byte, 10*1024)
					_ = data
				}
			},
		)
	}
}

// BenchmarkHeapPressure 堆压力测试
func BenchmarkHeapPressure(b *testing.B) {
	sizes := []int{1024, 10 * 1024, 100 * 1024}

	for _, size := range sizes {
		b.Run(
			"Size_"+formatSize(size),
			func(b *testing.B) {
				b.ReportAllocs()
				var data [][]byte
				for i := 0; i < b.N; i++ {
					data = append(data, make([]byte, size))
					if len(data) > 1000 {
						data = data[500:]
					}
				}
			},
		)
	}
}

func formatSize(size int) string {
	if size >= 1024*1024 {
		return string(rune('0'+size/(1024*1024))) + "MB"
	}
	if size >= 1024 {
		return string(rune('0'+size/1024)) + "KB"
	}
	return string(rune('0'+size)) + "B"
}

// ==================
// 7. 三色标记算法测试
// ==================

func TestTricolorGC(t *testing.T) {
	gc := NewTricolorGC()

	// 创建对象图
	obj1 := gc.AddObject(1, 100)
	obj2 := gc.AddObject(2, 200)
	obj3 := gc.AddObject(3, 150)
	obj4 := gc.AddObject(4, 300) // 不可达

	gc.AddReference(obj1, obj2)
	gc.AddReference(obj2, obj3)
	gc.AddRoot(obj1)

	// 执行标记
	gc.Mark()

	// 验证标记结果
	if obj1.Color != Black {
		t.Errorf("obj1 应为黑色，实际: %s", obj1.Color)
	}
	if obj2.Color != Black {
		t.Errorf("obj2 应为黑色，实际: %s", obj2.Color)
	}
	if obj3.Color != Black {
		t.Errorf("obj3 应为黑色，实际: %s", obj3.Color)
	}
	if obj4.Color != White {
		t.Errorf("obj4 应为白色（不可达），实际: %s", obj4.Color)
	}

	// 执行清除
	gc.Sweep()

	// 验证清除结果
	if len(gc.objects) != 3 {
		t.Errorf("清除后应剩余 3 个对象，实际: %d", len(gc.objects))
	}
}

func TestObjectColorString(t *testing.T) {
	tests := []struct {
		color    ObjectColor
		expected string
	}{
		{White, "White"},
		{Gray, "Gray"},
		{Black, "Black"},
	}

	for _, tt := range tests {
		if got := tt.color.String(); got != tt.expected {
			t.Errorf("ObjectColor(%d).String() = %s, want %s", tt.color, got, tt.expected)
		}
	}
}

// ==================
// 8. GC 监控器测试
// ==================

func TestGCMonitor(t *testing.T) {
	monitor := NewGCMonitor()
	monitor.Start()
	defer monitor.Stop()

	// 触发一些 GC
	for i := 0; i < 3; i++ {
		_ = make([]byte, 5*1024*1024)
		runtime.GC()
	}

	// 等待收集
	time.Sleep(200 * time.Millisecond)

	// 获取统计
	stats := monitor.GetLatestStats()

	t.Logf("GC 监控统计:")
	t.Logf("  NumGC: %d", stats.NumGC)
	t.Logf("  HeapSize: %d KB", stats.HeapSize/1024)
	t.Logf("  HeapUsed: %d KB", stats.HeapUsed/1024)

	// 验证
	if stats.NumGC == 0 {
		t.Error("GC 次数不应为 0")
	}

	// 获取历史
	history := monitor.GetStatsHistory()
	if len(history) == 0 {
		t.Error("历史记录不应为空")
	}
}

// ==================
// 9. 并发 GC 演示测试
// ==================

func TestConcurrentGCDemo(t *testing.T) {
	demo := NewConcurrentGCDemo()
	demo.StartAllocation()

	// 运行一段时间
	time.Sleep(500 * time.Millisecond)

	demo.Stop()

	// 验证没有 panic
	t.Log("并发 GC 演示完成")
}

// ==================
// 10. GC 调优器测试
// ==================

func TestGCTuner(t *testing.T) {
	tuner := NewGCTuner()

	// 测试延迟调优
	tuner.TuneForLatency()

	currentGOGC := debug.SetGCPercent(-1)
	debug.SetGCPercent(currentGOGC)

	if currentGOGC != 50 {
		t.Errorf("延迟调优后 GOGC 应为 50，实际: %d", currentGOGC)
	}

	// 恢复
	tuner.Restore()
}
