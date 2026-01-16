/*
=== GC 工具函数库 ===

提供可复用的 GC 监控、分析和调优工具函数。
这些函数可以在生产环境中使用，帮助诊断和优化 GC 性能。

主要功能：
1. GC 统计信息收集和格式化
2. GC 暂停时间分析
3. 内存压力检测
4. GC 调优建议生成
5. 实时 GC 监控
*/

package main

import (
	"fmt"
	"math"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// ==================
// 1. GC 统计工具
// ==================

// GCStatsCollector GC 统计收集器
// 用于收集和分析 GC 运行时统计信息
type GCStatsCollector struct {
	// 历史暂停时间记录
	pauseHistory []time.Duration
	// 历史堆大小记录
	heapHistory []uint64
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

// NewGCStatsCollector 创建新的 GC 统计收集器
// interval: 收集间隔
// maxHistory: 最大历史记录数
func NewGCStatsCollector(interval time.Duration, maxHistory int) *GCStatsCollector {
	return &GCStatsCollector{
		pauseHistory: make([]time.Duration, 0, maxHistory),
		heapHistory:  make([]uint64, 0, maxHistory),
		interval:     interval,
		maxHistory:   maxHistory,
		stopCh:       make(chan struct{}),
	}
}

// Start 启动统计收集
func (c *GCStatsCollector) Start() {
	if !c.running.CompareAndSwap(false, true) {
		return
	}

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		var lastNumGC uint32

		for {
			select {
			case <-ticker.C:
				var ms runtime.MemStats
				runtime.ReadMemStats(&ms)

				c.mu.Lock()

				// 记录新的 GC 暂停时间
				if ms.NumGC > lastNumGC {
					// 计算新增的 GC 次数
					newGCs := ms.NumGC - lastNumGC
					if newGCs > 256 {
						newGCs = 256 // PauseNs 数组最大长度
					}

					// 记录暂停时间
					for i := uint32(0); i < newGCs; i++ {
						idx := (ms.NumGC - 1 - i) % 256
						pause := time.Duration(ms.PauseNs[idx])
						c.pauseHistory = append(c.pauseHistory, pause)
					}

					lastNumGC = ms.NumGC
				}

				// 记录堆大小
				c.heapHistory = append(c.heapHistory, ms.HeapAlloc)

				// 限制历史记录数量
				if len(c.pauseHistory) > c.maxHistory {
					c.pauseHistory = c.pauseHistory[len(c.pauseHistory)-c.maxHistory:]
				}
				if len(c.heapHistory) > c.maxHistory {
					c.heapHistory = c.heapHistory[len(c.heapHistory)-c.maxHistory:]
				}

				c.mu.Unlock()

			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop 停止统计收集
func (c *GCStatsCollector) Stop() {
	if c.running.CompareAndSwap(true, false) {
		close(c.stopCh)
	}
}

// GetPausePercentiles 获取暂停时间百分位数
// 返回 P50, P90, P95, P99 百分位数
func (c *GCStatsCollector) GetPausePercentiles() (p50, p90, p95, p99 time.Duration) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.pauseHistory) == 0 {
		return 0, 0, 0, 0
	}

	// 复制并排序
	sorted := make([]time.Duration, len(c.pauseHistory))
	copy(sorted, c.pauseHistory)
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

// GetHeapGrowthRate 获取堆增长率
// 返回每秒的堆增长字节数
func (c *GCStatsCollector) GetHeapGrowthRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.heapHistory) < 2 {
		return 0
	}

	// 计算最近的增长率
	first := c.heapHistory[0]
	last := c.heapHistory[len(c.heapHistory)-1]
	samples := len(c.heapHistory)

	// 假设每个样本间隔为 c.interval
	totalTime := c.interval * time.Duration(samples-1)
	if totalTime == 0 {
		return 0
	}

	return float64(int64(last)-int64(first)) / totalTime.Seconds()
}

// ==================
// 2. GC 压力检测器
// ==================

// GCPressureLevel GC 压力级别
type GCPressureLevel int

const (
	// GCPressureLow 低压力 - GC 运行正常
	GCPressureLow GCPressureLevel = iota
	// GCPressureMedium 中等压力 - GC 频率较高
	GCPressureMedium
	// GCPressureHigh 高压力 - GC 频率过高，可能影响性能
	GCPressureHigh
	// GCPressureCritical 临界压力 - GC 占用大量 CPU 时间
	GCPressureCritical
)

func (l GCPressureLevel) String() string {
	switch l {
	case GCPressureLow:
		return "低"
	case GCPressureMedium:
		return "中等"
	case GCPressureHigh:
		return "高"
	case GCPressureCritical:
		return "临界"
	default:
		return "未知"
	}
}

// GCPressureDetector GC 压力检测器
// 用于检测当前 GC 压力级别并提供调优建议
type GCPressureDetector struct {
	// 检测窗口大小
	windowSize time.Duration
	// 上次检测时间
	lastCheck time.Time
	// 上次 GC 次数
	lastNumGC uint32
	// 上次 GC 总暂停时间
	lastPauseTotal time.Duration
	// 互斥锁
	mu sync.Mutex
}

// NewGCPressureDetector 创建新的 GC 压力检测器
func NewGCPressureDetector(windowSize time.Duration) *GCPressureDetector {
	return &GCPressureDetector{
		windowSize: windowSize,
		lastCheck:  time.Now(),
	}
}

// GCPressureReport GC 压力报告
type GCPressureReport struct {
	Level           GCPressureLevel // 压力级别
	GCFrequency     float64         // GC 频率 (次/秒)
	GCCPUPercent    float64         // GC CPU 占用百分比
	HeapUsageRatio  float64         // 堆使用率
	Recommendations []string        // 调优建议
}

// Detect 检测当前 GC 压力
func (d *GCPressureDetector) Detect() GCPressureReport {
	d.mu.Lock()
	defer d.mu.Unlock()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	now := time.Now()
	elapsed := now.Sub(d.lastCheck)

	// 计算 GC 频率
	gcCount := ms.NumGC - d.lastNumGC
	gcFrequency := float64(gcCount) / elapsed.Seconds()

	// 计算 GC CPU 占用
	pauseTotal := time.Duration(ms.PauseTotalNs)
	pauseDelta := pauseTotal - d.lastPauseTotal
	gcCPUPercent := float64(pauseDelta) / float64(elapsed) * 100

	// 计算堆使用率
	heapUsageRatio := float64(ms.HeapAlloc) / float64(ms.HeapSys)

	// 更新状态
	d.lastCheck = now
	d.lastNumGC = ms.NumGC
	d.lastPauseTotal = pauseTotal

	// 确定压力级别
	level := d.determineLevel(gcFrequency, gcCPUPercent, heapUsageRatio)

	// 生成建议
	recommendations := d.generateRecommendations(level, gcFrequency, gcCPUPercent, heapUsageRatio)

	return GCPressureReport{
		Level:           level,
		GCFrequency:     gcFrequency,
		GCCPUPercent:    gcCPUPercent,
		HeapUsageRatio:  heapUsageRatio,
		Recommendations: recommendations,
	}
}

func (d *GCPressureDetector) determineLevel(freq, cpuPercent, heapRatio float64) GCPressureLevel {
	// 基于多个指标综合判断
	// GC 频率阈值: 低 < 1/s, 中 < 5/s, 高 < 10/s, 临界 >= 10/s
	// CPU 占用阈值: 低 < 5%, 中 < 10%, 高 < 25%, 临界 >= 25%
	// 堆使用率阈值: 低 < 50%, 中 < 70%, 高 < 85%, 临界 >= 85%

	score := 0

	// 频率评分
	switch {
	case freq >= 10:
		score += 3
	case freq >= 5:
		score += 2
	case freq >= 1:
		score += 1
	}

	// CPU 占用评分
	switch {
	case cpuPercent >= 25:
		score += 3
	case cpuPercent >= 10:
		score += 2
	case cpuPercent >= 5:
		score += 1
	}

	// 堆使用率评分
	switch {
	case heapRatio >= 0.85:
		score += 3
	case heapRatio >= 0.70:
		score += 2
	case heapRatio >= 0.50:
		score += 1
	}

	// 综合评分
	switch {
	case score >= 7:
		return GCPressureCritical
	case score >= 5:
		return GCPressureHigh
	case score >= 3:
		return GCPressureMedium
	default:
		return GCPressureLow
	}
}

func (d *GCPressureDetector) generateRecommendations(level GCPressureLevel, freq, cpuPercent, heapRatio float64) []string {
	var recommendations []string

	if level == GCPressureLow {
		recommendations = append(recommendations, "GC 运行正常，无需调优")
		return recommendations
	}

	// 基于具体指标生成建议
	if freq >= 5 {
		recommendations = append(recommendations,
			"GC 频率过高，建议：",
			"  - 增加 GOGC 值（如 GOGC=200）减少 GC 频率",
			"  - 使用对象池减少分配",
			"  - 检查是否有不必要的临时对象分配",
		)
	}

	if cpuPercent >= 10 {
		recommendations = append(recommendations,
			"GC CPU 占用过高，建议：",
			"  - 减少堆上对象数量",
			"  - 使用 sync.Pool 复用对象",
			"  - 考虑使用栈分配（小对象、不逃逸）",
		)
	}

	if heapRatio >= 0.70 {
		recommendations = append(recommendations,
			"堆使用率过高，建议：",
			"  - 设置 GOMEMLIMIT 限制内存使用",
			"  - 检查是否有内存泄漏",
			"  - 优化数据结构减少内存占用",
		)
	}

	return recommendations
}

// ==================
// 3. GC 调优助手
// ==================

// GCTuningHelper GC 调优助手
// 提供自动化的 GC 调优功能
type GCTuningHelper struct {
	// 原始 GOGC 值
	originalGOGC int
	// 原始内存限制
	originalMemLimit int64
	// 目标延迟
	targetLatency time.Duration
	// 目标吞吐量
	targetThroughput float64
	// 是否已应用调优
	tuned bool
	// 互斥锁
	mu sync.Mutex
}

// NewGCTuningHelper 创建新的 GC 调优助手
func NewGCTuningHelper() *GCTuningHelper {
	return &GCTuningHelper{
		originalGOGC:     debug.SetGCPercent(-1),
		originalMemLimit: debug.SetMemoryLimit(-1),
		targetLatency:    5 * time.Millisecond,
		targetThroughput: 0.95, // 95% 吞吐量
	}
}

// SetTargetLatency 设置目标延迟
func (h *GCTuningHelper) SetTargetLatency(latency time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.targetLatency = latency
}

// SetTargetThroughput 设置目标吞吐量
func (h *GCTuningHelper) SetTargetThroughput(throughput float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.targetThroughput = throughput
}

// TuneForLatency 针对延迟优化
// 降低 GOGC 以减少单次 GC 暂停时间
func (h *GCTuningHelper) TuneForLatency() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 保存原始值
	if !h.tuned {
		h.originalGOGC = debug.SetGCPercent(-1)
		debug.SetGCPercent(h.originalGOGC)
	}

	// 降低 GOGC 以减少堆大小，从而减少 GC 暂停时间
	// 较小的堆意味着更少的对象需要扫描
	newGOGC := 50
	debug.SetGCPercent(newGOGC)

	h.tuned = true
	fmt.Printf("GC 延迟优化: GOGC 设置为 %d (原值: %d)\n", newGOGC, h.originalGOGC)
}

// TuneForThroughput 针对吞吐量优化
// 增加 GOGC 以减少 GC 频率
func (h *GCTuningHelper) TuneForThroughput() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 保存原始值
	if !h.tuned {
		h.originalGOGC = debug.SetGCPercent(-1)
		debug.SetGCPercent(h.originalGOGC)
	}

	// 增加 GOGC 以减少 GC 频率
	// 较大的堆意味着更少的 GC 次数
	newGOGC := 200
	debug.SetGCPercent(newGOGC)

	h.tuned = true
	fmt.Printf("GC 吞吐量优化: GOGC 设置为 %d (原值: %d)\n", newGOGC, h.originalGOGC)
}

// TuneWithMemoryLimit 使用内存限制调优
// Go 1.19+ 推荐的调优方式
func (h *GCTuningHelper) TuneWithMemoryLimit(limit int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 保存原始值
	if !h.tuned {
		h.originalMemLimit = debug.SetMemoryLimit(-1)
	}

	// 设置内存限制
	debug.SetMemoryLimit(limit)

	// 配合使用较高的 GOGC
	// 内存限制会在接近限制时自动触发 GC
	debug.SetGCPercent(200)

	h.tuned = true
	fmt.Printf("GC 内存限制优化: 限制设置为 %d MB, GOGC=200\n", limit/(1024*1024))
}

// Restore 恢复原始设置
func (h *GCTuningHelper) Restore() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.tuned {
		return
	}

	debug.SetGCPercent(h.originalGOGC)
	if h.originalMemLimit > 0 {
		debug.SetMemoryLimit(h.originalMemLimit)
	} else {
		debug.SetMemoryLimit(math.MaxInt64)
	}

	h.tuned = false
	fmt.Printf("GC 设置已恢复: GOGC=%d\n", h.originalGOGC)
}

// ==================
// 4. 内存分配追踪器
// ==================

// AllocationTracker 内存分配追踪器
// 用于追踪和分析内存分配模式
type AllocationTracker struct {
	// 分配计数
	allocCount atomic.Int64
	// 分配字节数
	allocBytes atomic.Int64
	// 开始时间
	startTime time.Time
	// 快照历史
	snapshots []AllocationSnapshot
	// 互斥锁
	mu sync.RWMutex
}

// AllocationSnapshot 分配快照
type AllocationSnapshot struct {
	Timestamp   time.Time
	AllocCount  int64
	AllocBytes  int64
	HeapObjects uint64
	HeapAlloc   uint64
}

// NewAllocationTracker 创建新的分配追踪器
func NewAllocationTracker() *AllocationTracker {
	return &AllocationTracker{
		startTime: time.Now(),
		snapshots: make([]AllocationSnapshot, 0),
	}
}

// TakeSnapshot 获取当前快照
func (t *AllocationTracker) TakeSnapshot() AllocationSnapshot {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	snapshot := AllocationSnapshot{
		Timestamp:   time.Now(),
		AllocCount:  t.allocCount.Load(),
		AllocBytes:  t.allocBytes.Load(),
		HeapObjects: ms.HeapObjects,
		HeapAlloc:   ms.HeapAlloc,
	}

	t.mu.Lock()
	t.snapshots = append(t.snapshots, snapshot)
	// 限制快照数量
	if len(t.snapshots) > 1000 {
		t.snapshots = t.snapshots[len(t.snapshots)-1000:]
	}
	t.mu.Unlock()

	return snapshot
}

// GetAllocationRate 获取分配速率
// 返回每秒分配的字节数
func (t *AllocationTracker) GetAllocationRate() float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	elapsed := time.Since(t.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	return float64(ms.TotalAlloc) / elapsed
}

// PrintSummary 打印分配摘要
func (t *AllocationTracker) PrintSummary() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	elapsed := time.Since(t.startTime)

	fmt.Println("\n=== 内存分配摘要 ===")
	fmt.Printf("运行时间: %v\n", elapsed)
	fmt.Printf("总分配次数: %d\n", ms.Mallocs)
	fmt.Printf("总释放次数: %d\n", ms.Frees)
	fmt.Printf("净分配次数: %d\n", ms.Mallocs-ms.Frees)
	fmt.Printf("总分配字节: %d MB\n", ms.TotalAlloc/(1024*1024))
	fmt.Printf("当前堆使用: %d MB\n", ms.HeapAlloc/(1024*1024))
	fmt.Printf("分配速率: %.2f MB/s\n", float64(ms.TotalAlloc)/(1024*1024)/elapsed.Seconds())
}

// ==================
// 5. 便捷函数
// ==================

// ForceGC 强制执行 GC 并等待完成
func ForceGC() {
	runtime.GC()
	runtime.GC() // 执行两次确保完成
}

// GetGCStats 获取当前 GC 统计信息
func GetGCStats() GCStats {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	stats := GCStats{
		NumGC:        ms.NumGC,
		NumForcedGC:  ms.NumForcedGC,
		GCPauseTotal: time.Duration(ms.PauseTotalNs),
		HeapSize:     ms.HeapSys,
		HeapUsed:     ms.HeapInuse,
		HeapObjects:  ms.HeapObjects,
		StackSize:    ms.StackSys,
		NextGC:       ms.NextGC,
		GCPercent:    debug.SetGCPercent(-1),
	}

	// 恢复 GOGC 设置
	debug.SetGCPercent(stats.GCPercent)

	// 计算暂停时间统计
	if ms.NumGC > 0 {
		var total, maxPause time.Duration
		minPause := time.Duration(math.MaxInt64)

		recentPauses := min(int(ms.NumGC), 256)
		for i := 0; i < recentPauses; i++ {
			pause := time.Duration(ms.PauseNs[i])
			total += pause
			if pause > maxPause {
				maxPause = pause
			}
			if pause < minPause && pause > 0 {
				minPause = pause
			}
		}

		stats.GCPauseMax = maxPause
		stats.GCPauseMin = minPause
		if recentPauses > 0 {
			stats.GCPauseAvg = total / time.Duration(recentPauses)
		}
	}

	return stats
}

// PrintGCStats 打印 GC 统计信息
func PrintGCStats() {
	stats := GetGCStats()

	fmt.Println("\n=== GC 统计信息 ===")
	fmt.Printf("GC 次数: %d (强制: %d)\n", stats.NumGC, stats.NumForcedGC)
	fmt.Printf("GC 总暂停: %v\n", stats.GCPauseTotal)
	fmt.Printf("GC 暂停 - 平均: %v, 最大: %v, 最小: %v\n",
		stats.GCPauseAvg, stats.GCPauseMax, stats.GCPauseMin)
	fmt.Printf("堆大小: %d MB\n", stats.HeapSize/(1024*1024))
	fmt.Printf("堆使用: %d MB\n", stats.HeapUsed/(1024*1024))
	fmt.Printf("堆对象: %d\n", stats.HeapObjects)
	fmt.Printf("下次 GC: %d MB\n", stats.NextGC/(1024*1024))
	fmt.Printf("GOGC: %d%%\n", stats.GCPercent)
}

// WatchGC 监控 GC 活动
// 在后台打印 GC 事件
func WatchGC(interval time.Duration, stopCh <-chan struct{}) {
	var lastNumGC uint32

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)

			if ms.NumGC > lastNumGC {
				gcCount := ms.NumGC - lastNumGC
				lastPause := time.Duration(ms.PauseNs[(ms.NumGC-1)%256])

				fmt.Printf("[GC] 新增 %d 次 GC, 最近暂停: %v, 堆: %d MB\n",
					gcCount, lastPause, ms.HeapAlloc/(1024*1024))

				lastNumGC = ms.NumGC
			}

		case <-stopCh:
			return
		}
	}
}
