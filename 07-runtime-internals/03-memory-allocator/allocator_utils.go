/*
=== 内存分配器工具函数库 ===

提供可复用的内存分配监控、分析和优化工具函数。
这些函数可以在生产环境中使用，帮助诊断和优化内存分配性能。

主要功能：
1. 内存分配统计收集
2. 对象池管理
3. 内存碎片分析
4. 分配热点检测
5. 内存使用报告生成
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
// 1. 内存分配统计工具
// ==================

// MemoryStatsCollector 内存统计收集器
// 用于收集和分析内存分配运行时统计信息
type MemoryStatsCollector struct {
	// 历史记录
	history []MemorySnapshot
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

// MemorySnapshot 内存快照
type MemorySnapshot struct {
	Timestamp    time.Time
	HeapAlloc    uint64 // 堆分配字节数
	HeapSys      uint64 // 堆系统内存
	HeapIdle     uint64 // 空闲堆内存
	HeapInuse    uint64 // 使用中堆内存
	HeapObjects  uint64 // 堆对象数
	StackInuse   uint64 // 栈使用内存
	Mallocs      uint64 // 累计分配次数
	Frees        uint64 // 累计释放次数
	TotalAlloc   uint64 // 累计分配字节数
	NumGC        uint32 // GC 次数
	GCPauseTotal uint64 // GC 总暂停时间
}

// NewMemoryStatsCollector 创建新的内存统计收集器
func NewMemoryStatsCollector(interval time.Duration, maxHistory int) *MemoryStatsCollector {
	return &MemoryStatsCollector{
		history:    make([]MemorySnapshot, 0, maxHistory),
		interval:   interval,
		maxHistory: maxHistory,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动统计收集
func (c *MemoryStatsCollector) Start() {
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
func (c *MemoryStatsCollector) Stop() {
	if c.running.CompareAndSwap(true, false) {
		close(c.stopCh)
	}
}

func (c *MemoryStatsCollector) collect() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	snapshot := MemorySnapshot{
		Timestamp:    time.Now(),
		HeapAlloc:    ms.HeapAlloc,
		HeapSys:      ms.HeapSys,
		HeapIdle:     ms.HeapIdle,
		HeapInuse:    ms.HeapInuse,
		HeapObjects:  ms.HeapObjects,
		StackInuse:   ms.StackInuse,
		Mallocs:      ms.Mallocs,
		Frees:        ms.Frees,
		TotalAlloc:   ms.TotalAlloc,
		NumGC:        ms.NumGC,
		GCPauseTotal: ms.PauseTotalNs,
	}

	c.mu.Lock()
	c.history = append(c.history, snapshot)
	if len(c.history) > c.maxHistory {
		c.history = c.history[1:]
	}
	c.mu.Unlock()
}

// GetLatest 获取最新快照
func (c *MemoryStatsCollector) GetLatest() MemorySnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.history) == 0 {
		return MemorySnapshot{}
	}
	return c.history[len(c.history)-1]
}

// GetHistory 获取历史记录
func (c *MemoryStatsCollector) GetHistory() []MemorySnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]MemorySnapshot, len(c.history))
	copy(result, c.history)
	return result
}

// GetAllocationRate 获取分配速率（字节/秒）
func (c *MemoryStatsCollector) GetAllocationRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.history) < 2 {
		return 0
	}

	first := c.history[0]
	last := c.history[len(c.history)-1]
	duration := last.Timestamp.Sub(first.Timestamp).Seconds()

	if duration == 0 {
		return 0
	}

	return float64(last.TotalAlloc-first.TotalAlloc) / duration
}

// GetObjectCreationRate 获取对象创建速率（对象/秒）
func (c *MemoryStatsCollector) GetObjectCreationRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.history) < 2 {
		return 0
	}

	first := c.history[0]
	last := c.history[len(c.history)-1]
	duration := last.Timestamp.Sub(first.Timestamp).Seconds()

	if duration == 0 {
		return 0
	}

	return float64(last.Mallocs-first.Mallocs) / duration
}

// ==================
// 2. 通用对象池
// ==================

// GenericPool 通用对象池
// 提供类型安全的对象池实现
type GenericPool[T any] struct {
	pool      sync.Pool
	newFunc   func() T
	resetFunc func(*T)
	stats     struct {
		Gets atomic.Int64
		Puts atomic.Int64
		News atomic.Int64
	}
}

// PoolStats 对象池统计
type PoolStats struct {
	Gets    int64
	Puts    int64
	News    int64
	HitRate float64
}

// NewGenericPool 创建新的通用对象池
// newFunc: 创建新对象的函数
// resetFunc: 重置对象的函数（可选）
func NewGenericPool[T any](newFunc func() T, resetFunc func(*T)) *GenericPool[T] {
	p := &GenericPool[T]{
		newFunc:   newFunc,
		resetFunc: resetFunc,
	}

	p.pool.New = func() interface{} {
		p.stats.News.Add(1)
		obj := newFunc()
		return &obj
	}

	return p
}

// Get 获取对象
func (p *GenericPool[T]) Get() *T {
	p.stats.Gets.Add(1)
	return p.pool.Get().(*T)
}

// Put 归还对象
func (p *GenericPool[T]) Put(obj *T) {
	if p.resetFunc != nil {
		p.resetFunc(obj)
	}
	p.stats.Puts.Add(1)
	p.pool.Put(obj)
}

// GetStats 获取统计信息
func (p *GenericPool[T]) GetStats() PoolStats {
	gets := p.stats.Gets.Load()
	news := p.stats.News.Load()
	puts := p.stats.Puts.Load()

	return PoolStats{
		Gets: gets,
		Puts: puts,
		News: news,
		HitRate: func() float64 {
			if gets > 0 {
				return float64(gets-news) / float64(gets)
			}
			return 0
		}(),
	}
}

// ==================
// 3. 字节缓冲池
// ==================

// ByteBufferPool 字节缓冲池
// 提供不同大小的字节缓冲区池
type ByteBufferPool struct {
	pools []*sync.Pool
	sizes []int
	stats ByteBufferPoolStats
}

// ByteBufferPoolStats 字节缓冲池统计
type ByteBufferPoolStats struct {
	Gets   atomic.Int64
	Puts   atomic.Int64
	Misses atomic.Int64 // 请求的大小超过最大池
}

// NewByteBufferPool 创建新的字节缓冲池
// sizes: 缓冲区大小列表，必须递增
func NewByteBufferPool(sizes []int) *ByteBufferPool {
	// 确保大小递增
	sort.Ints(sizes)

	pools := make([]*sync.Pool, len(sizes))
	for i, size := range sizes {
		s := size // 捕获变量
		pools[i] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, s)
			},
		}
	}

	return &ByteBufferPool{
		pools: pools,
		sizes: sizes,
	}
}

// Get 获取指定大小的缓冲区
func (p *ByteBufferPool) Get(size int) []byte {
	p.stats.Gets.Add(1)

	// 找到合适的池
	for i, s := range p.sizes {
		if size <= s {
			buf := p.pools[i].Get().([]byte)
			return buf[:size]
		}
	}

	// 请求的大小超过最大池，直接分配
	p.stats.Misses.Add(1)
	return make([]byte, size)
}

// Put 归还缓冲区
func (p *ByteBufferPool) Put(buf []byte) {
	p.stats.Puts.Add(1)

	cap := cap(buf)
	for i, s := range p.sizes {
		if cap == s {
			p.pools[i].Put(buf[:cap])
			return
		}
	}
	// 不是从池中获取的，不归还
}

// GetStats 获取统计信息
func (p *ByteBufferPool) GetStats() ByteBufferPoolStats {
	return ByteBufferPoolStats{
		Gets:   p.stats.Gets,
		Puts:   p.stats.Puts,
		Misses: p.stats.Misses,
	}
}

// ==================
// 4. 内存碎片分析器
// ==================

// FragmentationAnalyzer 内存碎片分析器
type FragmentationAnalyzer struct {
	samples []FragmentationSample
	mu      sync.Mutex
}

// FragmentationSample 碎片化样本
type FragmentationSample struct {
	Timestamp       time.Time
	HeapInuse       uint64
	HeapIdle        uint64
	HeapReleased    uint64
	FragmentRatio   float64 // 碎片率
	EfficiencyRatio float64 // 使用效率
}

// NewFragmentationAnalyzer 创建新的碎片分析器
func NewFragmentationAnalyzer() *FragmentationAnalyzer {
	return &FragmentationAnalyzer{
		samples: make([]FragmentationSample, 0),
	}
}

// Sample 采集碎片化样本
func (a *FragmentationAnalyzer) Sample() FragmentationSample {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// 计算碎片率：空闲内存 / 总堆内存
	fragmentRatio := float64(ms.HeapIdle) / float64(ms.HeapSys)

	// 计算使用效率：实际使用 / 已分配
	efficiencyRatio := float64(ms.HeapAlloc) / float64(ms.HeapInuse)

	sample := FragmentationSample{
		Timestamp:       time.Now(),
		HeapInuse:       ms.HeapInuse,
		HeapIdle:        ms.HeapIdle,
		HeapReleased:    ms.HeapReleased,
		FragmentRatio:   fragmentRatio,
		EfficiencyRatio: efficiencyRatio,
	}

	a.mu.Lock()
	a.samples = append(a.samples, sample)
	if len(a.samples) > 1000 {
		a.samples = a.samples[1:]
	}
	a.mu.Unlock()

	return sample
}

// GetAverageFragmentation 获取平均碎片率
func (a *FragmentationAnalyzer) GetAverageFragmentation() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.samples) == 0 {
		return 0
	}

	var total float64
	for _, s := range a.samples {
		total += s.FragmentRatio
	}
	return total / float64(len(a.samples))
}

// GetFragmentationReport 获取碎片化报告
func (a *FragmentationAnalyzer) GetFragmentationReport() FragmentationReport {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.samples) == 0 {
		return FragmentationReport{}
	}

	var totalFrag, totalEff float64
	minFrag, maxFrag := a.samples[0].FragmentRatio, a.samples[0].FragmentRatio

	for _, s := range a.samples {
		totalFrag += s.FragmentRatio
		totalEff += s.EfficiencyRatio
		if s.FragmentRatio < minFrag {
			minFrag = s.FragmentRatio
		}
		if s.FragmentRatio > maxFrag {
			maxFrag = s.FragmentRatio
		}
	}

	n := float64(len(a.samples))
	return FragmentationReport{
		SampleCount:        len(a.samples),
		AvgFragmentRatio:   totalFrag / n,
		MinFragmentRatio:   minFrag,
		MaxFragmentRatio:   maxFrag,
		AvgEfficiencyRatio: totalEff / n,
	}
}

// FragmentationReport 碎片化报告
type FragmentationReport struct {
	SampleCount        int
	AvgFragmentRatio   float64
	MinFragmentRatio   float64
	MaxFragmentRatio   float64
	AvgEfficiencyRatio float64
}

// String 格式化输出
func (r FragmentationReport) String() string {
	return fmt.Sprintf(
		"内存碎片化报告 (n=%d):\n"+
			"  平均碎片率: %.2f%%\n"+
			"  最小碎片率: %.2f%%\n"+
			"  最大碎片率: %.2f%%\n"+
			"  平均使用效率: %.2f%%",
		r.SampleCount,
		r.AvgFragmentRatio*100,
		r.MinFragmentRatio*100,
		r.MaxFragmentRatio*100,
		r.AvgEfficiencyRatio*100,
	)
}

// ==================
// 5. 分配热点检测器
// ==================

// AllocationHotspotDetector 分配热点检测器
// 检测频繁分配的代码位置
type AllocationHotspotDetector struct {
	baseline  runtime.MemStats
	current   runtime.MemStats
	interval  time.Duration
	threshold uint64 // 分配阈值
	onHotspot func(allocDelta, freeDelta uint64)
	running   atomic.Bool
	stopCh    chan struct{}
}

// NewAllocationHotspotDetector 创建新的分配热点检测器
func NewAllocationHotspotDetector(interval time.Duration, threshold uint64) *AllocationHotspotDetector {
	return &AllocationHotspotDetector{
		interval:  interval,
		threshold: threshold,
		stopCh:    make(chan struct{}),
	}
}

// SetHotspotCallback 设置热点回调
func (d *AllocationHotspotDetector) SetHotspotCallback(callback func(allocDelta, freeDelta uint64)) {
	d.onHotspot = callback
}

// Start 启动检测
func (d *AllocationHotspotDetector) Start() {
	if !d.running.CompareAndSwap(false, true) {
		return
	}

	runtime.ReadMemStats(&d.baseline)

	go func() {
		ticker := time.NewTicker(d.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				d.detect()
			case <-d.stopCh:
				return
			}
		}
	}()
}

// Stop 停止检测
func (d *AllocationHotspotDetector) Stop() {
	if d.running.CompareAndSwap(true, false) {
		close(d.stopCh)
	}
}

func (d *AllocationHotspotDetector) detect() {
	runtime.ReadMemStats(&d.current)

	allocDelta := d.current.Mallocs - d.baseline.Mallocs
	freeDelta := d.current.Frees - d.baseline.Frees

	if allocDelta > d.threshold && d.onHotspot != nil {
		d.onHotspot(allocDelta, freeDelta)
	}

	d.baseline = d.current
}

// ==================
// 6. 内存使用报告生成器
// ==================

// MemoryReportGenerator 内存报告生成器
type MemoryReportGenerator struct{}

// NewMemoryReportGenerator 创建新的报告生成器
func NewMemoryReportGenerator() *MemoryReportGenerator {
	return &MemoryReportGenerator{}
}

// GenerateReport 生成内存使用报告
func (g *MemoryReportGenerator) GenerateReport() MemoryReport {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return MemoryReport{
		Timestamp: time.Now(),

		// 堆内存
		HeapAlloc:    ms.HeapAlloc,
		HeapSys:      ms.HeapSys,
		HeapIdle:     ms.HeapIdle,
		HeapInuse:    ms.HeapInuse,
		HeapReleased: ms.HeapReleased,
		HeapObjects:  ms.HeapObjects,

		// 栈内存
		StackInuse: ms.StackInuse,
		StackSys:   ms.StackSys,

		// 分配器内存
		MSpanInuse:  ms.MSpanInuse,
		MSpanSys:    ms.MSpanSys,
		MCacheInuse: ms.MCacheInuse,
		MCacheSys:   ms.MCacheSys,

		// 分配统计
		Mallocs:    ms.Mallocs,
		Frees:      ms.Frees,
		TotalAlloc: ms.TotalAlloc,

		// GC 统计
		NumGC:        ms.NumGC,
		GCPauseTotal: ms.PauseTotalNs,
		NextGC:       ms.NextGC,

		// 系统内存
		Sys:      ms.Sys,
		OtherSys: ms.OtherSys,
	}
}

// MemoryReport 内存报告
type MemoryReport struct {
	Timestamp time.Time

	// 堆内存
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	HeapObjects  uint64

	// 栈内存
	StackInuse uint64
	StackSys   uint64

	// 分配器内存
	MSpanInuse  uint64
	MSpanSys    uint64
	MCacheInuse uint64
	MCacheSys   uint64

	// 分配统计
	Mallocs    uint64
	Frees      uint64
	TotalAlloc uint64

	// GC 统计
	NumGC        uint32
	GCPauseTotal uint64
	NextGC       uint64

	// 系统内存
	Sys      uint64
	OtherSys uint64
}

// String 格式化输出
func (r MemoryReport) String() string {
	return fmt.Sprintf(`
=== 内存使用报告 ===
时间: %s

堆内存:
  已分配: %d MB
  系统获取: %d MB
  空闲: %d MB
  使用中: %d MB
  已释放: %d MB
  对象数: %d

栈内存:
  使用中: %d KB
  系统获取: %d KB

分配器:
  MSpan 使用: %d KB
  MCache 使用: %d KB

分配统计:
  累计分配: %d 次
  累计释放: %d 次
  净分配: %d 次
  累计分配字节: %d MB

GC 统计:
  GC 次数: %d
  GC 总暂停: %d ms
  下次 GC: %d MB

系统内存:
  总计: %d MB
  其他: %d KB
`,
		r.Timestamp.Format("2006-01-02 15:04:05"),
		r.HeapAlloc/(1024*1024),
		r.HeapSys/(1024*1024),
		r.HeapIdle/(1024*1024),
		r.HeapInuse/(1024*1024),
		r.HeapReleased/(1024*1024),
		r.HeapObjects,
		r.StackInuse/1024,
		r.StackSys/1024,
		r.MSpanInuse/1024,
		r.MCacheInuse/1024,
		r.Mallocs,
		r.Frees,
		r.Mallocs-r.Frees,
		r.TotalAlloc/(1024*1024),
		r.NumGC,
		r.GCPauseTotal/1000000,
		r.NextGC/(1024*1024),
		r.Sys/(1024*1024),
		r.OtherSys/1024,
	)
}

// ==================
// 7. 便捷函数
// ==================

// GetMemoryStats 获取当前内存统计
func GetMemoryStats() MemorySnapshot {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return MemorySnapshot{
		Timestamp:    time.Now(),
		HeapAlloc:    ms.HeapAlloc,
		HeapSys:      ms.HeapSys,
		HeapIdle:     ms.HeapIdle,
		HeapInuse:    ms.HeapInuse,
		HeapObjects:  ms.HeapObjects,
		StackInuse:   ms.StackInuse,
		Mallocs:      ms.Mallocs,
		Frees:        ms.Frees,
		TotalAlloc:   ms.TotalAlloc,
		NumGC:        ms.NumGC,
		GCPauseTotal: ms.PauseTotalNs,
	}
}

// PrintMemoryStats 打印内存统计
func PrintMemoryStats() {
	stats := GetMemoryStats()

	fmt.Println("\n=== 内存统计 ===")
	fmt.Printf("堆分配: %d MB\n", stats.HeapAlloc/(1024*1024))
	fmt.Printf("堆使用: %d MB\n", stats.HeapInuse/(1024*1024))
	fmt.Printf("堆对象: %d\n", stats.HeapObjects)
	fmt.Printf("栈使用: %d KB\n", stats.StackInuse/1024)
	fmt.Printf("分配次数: %d\n", stats.Mallocs)
	fmt.Printf("释放次数: %d\n", stats.Frees)
	fmt.Printf("GC 次数: %d\n", stats.NumGC)
}

// WatchMemory 监控内存变化
func WatchMemory(interval time.Duration, stopCh <-chan struct{}) {
	var lastAlloc uint64

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var ms runtime.MemStats
			runtime.ReadMemStats(&ms)

			diff := int64(ms.HeapAlloc) - int64(lastAlloc)
			sign := "+"
			if diff < 0 {
				sign = ""
			}

			fmt.Printf("[Memory] 堆: %d MB (%s%d KB), 对象: %d, GC: %d\n",
				ms.HeapAlloc/(1024*1024),
				sign, diff/1024,
				ms.HeapObjects,
				ms.NumGC)

			lastAlloc = ms.HeapAlloc

		case <-stopCh:
			return
		}
	}
}

// CalculateFragmentation 计算当前碎片率
func CalculateFragmentation() float64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	if ms.HeapSys == 0 {
		return 0
	}
	return float64(ms.HeapIdle) / float64(ms.HeapSys)
}

// EstimateObjectSize 估算对象大小
// 通过分配前后的内存差异估算
func EstimateObjectSize(allocFunc func()) uint64 {
	runtime.GC()
	runtime.GC()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	allocFunc()

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	if after.HeapAlloc > before.HeapAlloc {
		return after.HeapAlloc - before.HeapAlloc
	}
	return 0
}
