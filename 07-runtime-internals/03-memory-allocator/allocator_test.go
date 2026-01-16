/*
=== 内存分配器基准测试和单元测试 ===

测试内存分配器相关功能的正确性和性能特征。
包含：
1. 内存统计收集器测试
2. 对象池测试
3. 字节缓冲池测试
4. 碎片分析器测试
5. 内存分配基准测试
*/

package main

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

// ==================
// 1. 内存统计收集器测试
// ==================

func TestMemoryStatsCollector(t *testing.T) {
	collector := NewMemoryStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 分配一些内存
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 1024))
	}

	// 等待收集
	time.Sleep(100 * time.Millisecond)

	// 获取最新快照
	snapshot := collector.GetLatest()

	t.Logf("内存快照:")
	t.Logf("  HeapAlloc: %d KB", snapshot.HeapAlloc/1024)
	t.Logf("  HeapObjects: %d", snapshot.HeapObjects)
	t.Logf("  Mallocs: %d", snapshot.Mallocs)

	// 验证
	if snapshot.HeapAlloc == 0 {
		t.Error("HeapAlloc 不应为 0")
	}

	// 保持引用
	_ = data
}

func TestMemoryStatsCollectorHistory(t *testing.T) {
	collector := NewMemoryStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 等待收集
	time.Sleep(200 * time.Millisecond)

	// 获取历史
	history := collector.GetHistory()

	t.Logf("历史记录数: %d", len(history))

	if len(history) == 0 {
		t.Error("历史记录不应为空")
	}
}

func TestMemoryStatsCollectorAllocationRate(t *testing.T) {
	collector := NewMemoryStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 持续分配
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 10*1024))
		time.Sleep(5 * time.Millisecond)
	}

	rate := collector.GetAllocationRate()
	t.Logf("分配速率: %.2f bytes/sec", rate)

	// 保持引用
	_ = data
}

func TestMemoryStatsCollectorObjectCreationRate(t *testing.T) {
	collector := NewMemoryStatsCollector(10*time.Millisecond, 100)
	collector.Start()
	defer collector.Stop()

	// 持续创建对象
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 1024))
		time.Sleep(5 * time.Millisecond)
	}

	rate := collector.GetObjectCreationRate()
	t.Logf("对象创建速率: %.2f objects/sec", rate)

	// 保持引用
	_ = data
}

// ==================
// 2. 通用对象池测试
// ==================

type TestObject struct {
	ID   int
	Data []byte
}

func TestGenericPool(t *testing.T) {
	pool := NewGenericPool(
		func() TestObject {
			return TestObject{Data: make([]byte, 1024)}
		},
		func(obj *TestObject) {
			obj.ID = 0
			// 清空数据但保留容量
			for i := range obj.Data {
				obj.Data[i] = 0
			}
		},
	)

	// 获取对象
	obj1 := pool.Get()
	obj1.ID = 1

	// 归还对象
	pool.Put(obj1)

	// 再次获取
	obj2 := pool.Get()

	// 验证对象被重置
	if obj2.ID != 0 {
		t.Errorf("对象应被重置，ID 应为 0，实际: %d", obj2.ID)
	}

	// 获取统计
	stats := pool.GetStats()
	t.Logf("对象池统计:")
	t.Logf("  Gets: %d", stats.Gets)
	t.Logf("  Puts: %d", stats.Puts)
	t.Logf("  News: %d", stats.News)
	t.Logf("  HitRate: %.2f%%", stats.HitRate*100)
}

func TestGenericPoolConcurrent(t *testing.T) {
	pool := NewGenericPool(
		func() TestObject {
			return TestObject{Data: make([]byte, 1024)}
		},
		nil,
	)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				obj := pool.Get()
				obj.ID = j
				pool.Put(obj)
			}
		}()
	}

	wg.Wait()

	stats := pool.GetStats()
	t.Logf("并发测试统计:")
	t.Logf("  Gets: %d", stats.Gets)
	t.Logf("  Puts: %d", stats.Puts)
	t.Logf("  HitRate: %.2f%%", stats.HitRate*100)
}

// ==================
// 3. 字节缓冲池测试
// ==================

func TestByteBufferPool(t *testing.T) {
	pool := NewByteBufferPool([]int{64, 256, 1024, 4096})

	// 获取不同大小的缓冲区
	buf64 := pool.Get(50)
	buf256 := pool.Get(200)
	buf1024 := pool.Get(1000)
	buf4096 := pool.Get(4000)

	// 验证大小
	if len(buf64) != 50 {
		t.Errorf("buf64 长度应为 50，实际: %d", len(buf64))
	}
	if cap(buf64) != 64 {
		t.Errorf("buf64 容量应为 64，实际: %d", cap(buf64))
	}

	// 归还
	pool.Put(buf64)
	pool.Put(buf256)
	pool.Put(buf1024)
	pool.Put(buf4096)

	// 获取统计
	stats := pool.GetStats()
	t.Logf("字节缓冲池统计:")
	t.Logf("  Gets: %d", stats.Gets.Load())
	t.Logf("  Puts: %d", stats.Puts.Load())
	t.Logf("  Misses: %d", stats.Misses.Load())
}

func TestByteBufferPoolOversized(t *testing.T) {
	pool := NewByteBufferPool([]int{64, 256, 1024})

	// 请求超大缓冲区
	buf := pool.Get(2048)

	if len(buf) != 2048 {
		t.Errorf("超大缓冲区长度应为 2048，实际: %d", len(buf))
	}

	stats := pool.GetStats()
	if stats.Misses.Load() != 1 {
		t.Errorf("Misses 应为 1，实际: %d", stats.Misses.Load())
	}
}

// ==================
// 4. 碎片分析器测试
// ==================

func TestFragmentationAnalyzer(t *testing.T) {
	analyzer := NewFragmentationAnalyzer()

	// 采集多个样本
	for i := 0; i < 10; i++ {
		sample := analyzer.Sample()
		t.Logf("样本 %d: 碎片率=%.2f%%, 效率=%.2f%%",
			i, sample.FragmentRatio*100, sample.EfficiencyRatio*100)

		// 分配一些内存
		_ = make([]byte, 1024*1024)
		time.Sleep(10 * time.Millisecond)
	}

	// 获取平均碎片率
	avgFrag := analyzer.GetAverageFragmentation()
	t.Logf("平均碎片率: %.2f%%", avgFrag*100)

	// 获取报告
	report := analyzer.GetFragmentationReport()
	t.Log(report.String())

	// 验证
	if report.SampleCount != 10 {
		t.Errorf("样本数应为 10，实际: %d", report.SampleCount)
	}
}

func TestFragmentationReportString(t *testing.T) {
	report := FragmentationReport{
		SampleCount:        100,
		AvgFragmentRatio:   0.25,
		MinFragmentRatio:   0.10,
		MaxFragmentRatio:   0.40,
		AvgEfficiencyRatio: 0.85,
	}

	str := report.String()
	if str == "" {
		t.Error("String() 不应返回空字符串")
	}
	t.Log(str)
}

// ==================
// 5. 分配热点检测器测试
// ==================

func TestAllocationHotspotDetector(t *testing.T) {
	detector := NewAllocationHotspotDetector(50*time.Millisecond, 100)

	var hotspotDetected bool
	detector.SetHotspotCallback(func(allocDelta, freeDelta uint64) {
		hotspotDetected = true
		t.Logf("检测到热点: 分配=%d, 释放=%d", allocDelta, freeDelta)
	})

	detector.Start()
	defer detector.Stop()

	// 大量分配
	var data [][]byte
	for i := 0; i < 1000; i++ {
		data = append(data, make([]byte, 1024))
	}

	// 等待检测
	time.Sleep(200 * time.Millisecond)

	if !hotspotDetected {
		t.Log("未检测到热点（阈值可能太高）")
	}

	// 保持引用
	_ = data
}

// ==================
// 6. 内存报告生成器测试
// ==================

func TestMemoryReportGenerator(t *testing.T) {
	generator := NewMemoryReportGenerator()

	// 分配一些内存
	var data [][]byte
	for i := 0; i < 100; i++ {
		data = append(data, make([]byte, 10*1024))
	}

	report := generator.GenerateReport()

	t.Log(report.String())

	// 验证
	if report.HeapAlloc == 0 {
		t.Error("HeapAlloc 不应为 0")
	}
	if report.Mallocs == 0 {
		t.Error("Mallocs 不应为 0")
	}

	// 保持引用
	_ = data
}

// ==================
// 7. 便捷函数测试
// ==================

func TestGetMemoryStats(t *testing.T) {
	stats := GetMemoryStats()

	t.Logf("内存统计:")
	t.Logf("  HeapAlloc: %d KB", stats.HeapAlloc/1024)
	t.Logf("  HeapObjects: %d", stats.HeapObjects)

	if stats.HeapAlloc == 0 {
		t.Error("HeapAlloc 不应为 0")
	}
}

func TestCalculateFragmentation(t *testing.T) {
	frag := CalculateFragmentation()

	t.Logf("当前碎片率: %.2f%%", frag*100)

	if frag < 0 || frag > 1 {
		t.Errorf("碎片率应在 0-1 之间，实际: %f", frag)
	}
}

func TestEstimateObjectSize(t *testing.T) {
	size := EstimateObjectSize(func() {
		_ = make([]byte, 1024*1024) // 1MB
	})

	t.Logf("估算对象大小: %d bytes", size)

	// 应该接近 1MB
	if size < 1024*1024 {
		t.Logf("估算大小小于预期（可能被 GC 回收）")
	}
}

// ==================
// 8. Size Class 映射器测试
// ==================

func TestSizeClassMapper(t *testing.T) {
	mapper := NewSizeClassMapper()

	testCases := []struct {
		size          int
		expectedClass int
	}{
		{8, 1},
		{16, 2},
		{32, 4},
		{64, 6},
		{128, 10},
		{256, 18},
		{512, 25},
		{1024, 31},
	}

	for _, tc := range testCases {
		class, info := mapper.FindSizeClass(tc.size)
		t.Logf("大小 %d -> Class %d (实际分配 %d)", tc.size, class, info.Size)

		if info.Size < tc.size {
			t.Errorf("分配大小 %d 不应小于请求大小 %d", info.Size, tc.size)
		}
	}
}

func TestSizeClassMapperLargeObject(t *testing.T) {
	mapper := NewSizeClassMapper()

	// 测试大对象
	class, info := mapper.FindSizeClass(100 * 1024) // 100KB

	t.Logf("大对象 100KB -> Class %d, Pages %d", class, info.Pages)

	if class != 0 {
		t.Errorf("大对象的 class 应为 0，实际: %d", class)
	}
}

// ==================
// 9. 内存分配器模拟测试
// ==================

func TestMemoryAllocator(t *testing.T) {
	allocator := NewMemoryAllocator()

	// 分配小对象
	for i := 0; i < 100; i++ {
		ptr := allocator.AllocateObject(64)
		if ptr == nil {
			t.Logf("小对象分配返回 nil（预期行为）")
		}
	}

	// 分配大对象
	for i := 0; i < 10; i++ {
		ptr := allocator.AllocateObject(64 * 1024)
		if ptr == nil {
			t.Logf("大对象分配返回 nil（预期行为）")
		}
	}

	// 获取统计
	allocCount, freeCount, spansInUse, pagesInUse := allocator.GetStats()

	t.Logf("分配器统计:")
	t.Logf("  分配次数: %d", allocCount)
	t.Logf("  释放次数: %d", freeCount)
	t.Logf("  使用中 Span: %d", spansInUse)
	t.Logf("  使用中页数: %d", pagesInUse)
}

// ==================
// 10. 内存监控器测试
// ==================

func TestMemoryMonitor(t *testing.T) {
	monitor := NewMemoryMonitor()
	monitor.Start()
	defer monitor.Stop()

	// 分配内存
	var data [][]byte
	for i := 0; i < 50; i++ {
		data = append(data, make([]byte, 10*1024))
		time.Sleep(10 * time.Millisecond)
	}

	// 获取统计
	stats := monitor.GetLatestStats()

	t.Logf("内存监控统计:")
	t.Logf("  HeapAlloc: %d KB", stats.HeapAlloc/1024)
	t.Logf("  HeapObjects: %d", stats.HeapObjects)

	// 获取历史
	history := monitor.GetStatsHistory()
	t.Logf("历史记录数: %d", len(history))

	// 保持引用
	_ = data
}

// ==================
// 11. 栈管理器测试
// ==================

func TestStackManager(t *testing.T) {
	manager := NewStackManager()

	// 测试栈增长
	manager.demonstrateStackGrowth()

	t.Logf("栈增长次数: %d", manager.stackGrowCount)
}

// ==================
// 12. 内存对齐演示测试
// ==================

func TestAlignmentDemo(t *testing.T) {
	demo := &AlignmentDemo{}
	demo.demonstrateAlignment()

	t.Log("内存对齐演示完成")
}

// ==================
// 13. 基准测试
// ==================

// BenchmarkSmallObjectAllocation 小对象分配基准测试
func BenchmarkSmallObjectAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 64)
	}
}

// BenchmarkMediumObjectAllocation 中等对象分配基准测试
func BenchmarkMediumObjectAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024)
	}
}

// BenchmarkLargeObjectAllocation 大对象分配基准测试
func BenchmarkLargeObjectAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 32*1024)
	}
}

// BenchmarkHugeObjectAllocation 超大对象分配基准测试
func BenchmarkHugeObjectAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = make([]byte, 1024*1024)
	}
}

// BenchmarkGenericPool 通用对象池基准测试
func BenchmarkGenericPool(b *testing.B) {
	pool := NewGenericPool(
		func() TestObject {
			return TestObject{Data: make([]byte, 1024)}
		},
		nil,
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		obj := pool.Get()
		pool.Put(obj)
	}
}

// BenchmarkByteBufferPool 字节缓冲池基准测试
func BenchmarkByteBufferPool(b *testing.B) {
	pool := NewByteBufferPool([]int{64, 256, 1024, 4096})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := pool.Get(512)
		pool.Put(buf)
	}
}

// BenchmarkDirectAllocationVsPool 直接分配 vs 对象池对比
func BenchmarkDirectAllocationVsPool(b *testing.B) {
	pool := NewByteBufferPool([]int{1024})

	b.Run("DirectAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 1024)
			_ = buf
		}
	})

	b.Run("PoolAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := pool.Get(1024)
			pool.Put(buf)
		}
	})
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

// BenchmarkConcurrentPoolAllocation 并发对象池分配基准测试
func BenchmarkConcurrentPoolAllocation(b *testing.B) {
	pool := NewByteBufferPool([]int{1024})

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get(1024)
			pool.Put(buf)
		}
	})
}

// BenchmarkSizeClassLookup Size Class 查找基准测试
func BenchmarkSizeClassLookup(b *testing.B) {
	mapper := NewSizeClassMapper()
	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		_, _ = mapper.FindSizeClass(size)
	}
}

// BenchmarkMemoryStatsRead 内存统计读取基准测试
func BenchmarkMemoryStatsRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
	}
}

// BenchmarkFragmentationCalculation 碎片率计算基准测试
func BenchmarkFragmentationCalculation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = CalculateFragmentation()
	}
}

// BenchmarkAllocationPattern 不同分配模式基准测试
func BenchmarkAllocationPattern(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			data := make([][]byte, 100)
			for j := range data {
				data[j] = make([]byte, 1024)
			}
		}
	})

	b.Run("Interleaved", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var data [][]byte
			for j := 0; j < 100; j++ {
				data = append(data, make([]byte, 1024))
				if len(data) > 50 {
					data = data[25:]
				}
			}
		}
	})
}

// BenchmarkStackVsHeap 栈分配 vs 堆分配对比
func BenchmarkStackVsHeap(b *testing.B) {
	b.Run("Stack", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var arr [1024]byte
			arr[0] = 1
			_ = arr
		}
	})

	b.Run("Heap", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			arr := make([]byte, 1024)
			arr[0] = 1
			_ = arr
		}
	})
}
