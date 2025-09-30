/*
=== Go运行时内核：内存分配器深度解析 ===

本模块深入Go语言内存分配器的内核实现，探索：
1. TCMalloc启发的分配器架构
2. mspan/mcache/mcentral/mheap分层设计
3. Size Class和Object Size映射
4. 小对象、大对象、超大对象分配策略
5. 栈内存管理和栈增长
6. 内存对齐和指针安全
7. 内存分配器性能优化
8. 逃逸分析对分配的影响
9. 内存碎片化和整理

学习目标：
- 深入理解Go内存分配器架构
- 掌握不同大小对象的分配策略
- 学会内存分配性能分析和优化
- 理解栈和堆的管理机制
*/

package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// ==================
// 1. 内存统计和监控
// ==================

// MemoryStats 内存统计信息
type MemoryStats struct {
	// 堆内存统计
	HeapAlloc    uint64 // 堆上已分配的字节数
	HeapSys      uint64 // 从系统获得的堆内存
	HeapIdle     uint64 // 空闲的堆内存
	HeapInuse    uint64 // 正在使用的堆内存
	HeapReleased uint64 // 返回给系统的内存
	HeapObjects  uint64 // 堆上的对象数量

	// 栈内存统计
	StackInuse uint64 // 栈正在使用的内存
	StackSys   uint64 // 从系统获得的栈内存

	// 分配器统计
	MSpanInuse  uint64 // mspan结构使用的内存
	MSpanSys    uint64 // mspan系统内存
	MCacheInuse uint64 // mcache结构使用的内存
	MCacheSys   uint64 // mcache系统内存

	// 垃圾收集器统计
	NextGC       uint64 // 下次GC的目标堆大小
	LastGC       uint64 // 上次GC的时间戳
	PauseTotalNs uint64 // GC总暂停时间

	// 分配统计
	Mallocs    uint64 // 累计分配次数
	Frees      uint64 // 累计释放次数
	TotalAlloc uint64 // 累计分配字节数
	Lookups    uint64 // 累计指针查找次数

	// 系统统计
	Sys         uint64 // 从系统获得的总内存
	OtherSys    uint64 // 其他系统内存
	GCSys       uint64 // GC系统内存
	BuckHashSys uint64 // profiling bucket系统内存

	Timestamp time.Time // 时间戳
}

// MemoryMonitor 内存监控器
type MemoryMonitor struct {
	stats     []MemoryStats
	startTime time.Time
	mutex     sync.RWMutex
	running   bool
	stopCh    chan struct{}
}

func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		stats:     make([]MemoryStats, 0),
		startTime: time.Now(),
		stopCh:    make(chan struct{}),
	}
}

func (m *MemoryMonitor) Start() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return
	}

	m.running = true
	go m.monitor()
}

func (m *MemoryMonitor) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)
}

func (m *MemoryMonitor) monitor() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectStats()
		case <-m.stopCh:
			return
		}
	}
}

func (m *MemoryMonitor) collectStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	stats := MemoryStats{
		HeapAlloc:    ms.HeapAlloc,
		HeapSys:      ms.HeapSys,
		HeapIdle:     ms.HeapIdle,
		HeapInuse:    ms.HeapInuse,
		HeapReleased: ms.HeapReleased,
		HeapObjects:  ms.HeapObjects,
		StackInuse:   ms.StackInuse,
		StackSys:     ms.StackSys,
		MSpanInuse:   ms.MSpanInuse,
		MSpanSys:     ms.MSpanSys,
		MCacheInuse:  ms.MCacheInuse,
		MCacheSys:    ms.MCacheSys,
		NextGC:       ms.NextGC,
		LastGC:       ms.LastGC,
		PauseTotalNs: ms.PauseTotalNs,
		Mallocs:      ms.Mallocs,
		Frees:        ms.Frees,
		TotalAlloc:   ms.TotalAlloc,
		Lookups:      ms.Lookups,
		Sys:          ms.Sys,
		OtherSys:     ms.OtherSys,
		GCSys:        ms.GCSys,
		BuckHashSys:  ms.BuckHashSys,
		Timestamp:    time.Now(),
	}

	m.mutex.Lock()
	m.stats = append(m.stats, stats)
	// 保持最近1000个统计数据
	if len(m.stats) > 1000 {
		m.stats = m.stats[1:]
	}
	m.mutex.Unlock()
}

func (m *MemoryMonitor) GetLatestStats() MemoryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.stats) == 0 {
		return MemoryStats{}
	}

	return m.stats[len(m.stats)-1]
}

func (m *MemoryMonitor) GetStatsHistory() []MemoryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]MemoryStats, len(m.stats))
	copy(history, m.stats)
	return history
}

// ==================
// 2. Size Class映射演示
// ==================

// SizeClassInfo Size Class信息
type SizeClassInfo struct {
	Size       int // 对象大小
	Pages      int // 页数
	Objects    int // 每个span的对象数
	WasteBytes int // 浪费的字节数
	MaxWaste   int // 最大浪费字节数
}

// SizeClassMapper Size Class映射器
type SizeClassMapper struct {
	classes []SizeClassInfo
}

func NewSizeClassMapper() *SizeClassMapper {
	// 模拟Go的Size Class映射表（简化版本）
	classes := []SizeClassInfo{
		{8, 1, 512, 0, 7},
		{16, 1, 256, 0, 15},
		{24, 1, 170, 8, 23},
		{32, 1, 128, 0, 31},
		{48, 1, 85, 8, 47},
		{64, 1, 64, 0, 63},
		{80, 1, 51, 16, 79},
		{96, 1, 42, 32, 95},
		{112, 1, 36, 32, 111},
		{128, 1, 32, 0, 127},
		{144, 1, 28, 32, 143},
		{160, 1, 25, 32, 159},
		{176, 1, 23, 16, 175},
		{192, 1, 21, 32, 191},
		{208, 1, 19, 48, 207},
		{224, 1, 18, 32, 223},
		{240, 1, 17, 16, 239},
		{256, 1, 16, 0, 255},
		{288, 1, 14, 32, 287},
		{320, 1, 12, 64, 319},
		{352, 1, 11, 96, 351},
		{384, 1, 10, 128, 383},
		{416, 1, 9, 160, 415},
		{448, 1, 9, 64, 447},
		{512, 1, 8, 0, 511},
		{576, 1, 7, 64, 575},
		{640, 1, 6, 128, 639},
		{704, 1, 5, 192, 703},
		{768, 1, 5, 128, 767},
		{896, 1, 4, 256, 895},
		{1024, 1, 4, 0, 1023},
		{1152, 1, 3, 384, 1151},
		{1280, 1, 3, 128, 1279},
		{1408, 2, 4, 512, 1407},
		{1536, 1, 2, 512, 1535},
		{1792, 2, 3, 256, 1791},
		{2048, 1, 2, 0, 2047},
		{2304, 2, 2, 512, 2303},
		{2688, 1, 1, 1344, 2687},
		{3072, 3, 2, 0, 3071},
		{3200, 2, 1, 1856, 3199},
		{3456, 3, 1, 1600, 3455},
		{4096, 1, 1, 0, 4095},
		{4864, 3, 1, 1120, 4863},
		{5376, 2, 1, 896, 5375},
		{6144, 3, 1, 512, 6143},
		{6528, 4, 1, 1536, 6527},
		{6784, 3, 1, 256, 6783},
		{6912, 4, 1, 512, 6911},
		{8192, 1, 1, 0, 8191},
		{9472, 2, 1, 1024, 9471},
		{9728, 3, 1, 512, 9727},
		{10240, 2, 1, 512, 10239},
		{10880, 2, 1, 128, 10879},
		{12288, 3, 1, 0, 12287},
		{13568, 2, 1, 256, 13567},
		{14336, 2, 1, 512, 14335},
		{16384, 1, 1, 0, 16383},
		{18432, 3, 1, 256, 18431},
		{19072, 2, 1, 1024, 19071},
		{20480, 4, 1, 0, 20479},
		{21760, 2, 1, 512, 21759},
		{24576, 3, 1, 0, 24575},
		{27264, 2, 1, 256, 27263},
		{28672, 4, 1, 0, 28671},
		{32768, 1, 1, 0, 32767},
	}

	return &SizeClassMapper{classes: classes}
}

func (scm *SizeClassMapper) FindSizeClass(size int) (int, SizeClassInfo) {
	for i, class := range scm.classes {
		if size <= class.Size {
			return i + 1, class // Size class从1开始
		}
	}
	// 大对象（>32KB）
	return 0, SizeClassInfo{Size: size, Pages: (size + 8191) / 8192, Objects: 1}
}

func (scm *SizeClassMapper) PrintSizeClasses() {
	fmt.Println("=== Go Size Class映射表 ===")
	fmt.Printf("%-5s %-8s %-6s %-8s %-10s %-10s\n",
		"Class", "Size", "Pages", "Objects", "WasteBytes", "MaxWaste")
	fmt.Println(strings.Repeat("-", 60))

	for i, class := range scm.classes {
		if i > 20 { // 只显示前20个
			fmt.Println("... (省略其余size class)")
			break
		}
		fmt.Printf("%-5d %-8d %-6d %-8d %-10d %-10d\n",
			i+1, class.Size, class.Pages, class.Objects, class.WasteBytes, class.MaxWaste)
	}
}

// ==================
// 3. 内存分配器模拟
// ==================

// MockSpan 模拟mspan结构
type MockSpan struct {
	StartAddr  uintptr // 起始地址
	NPPages    int     // 页数
	SizeClass  uint8   // Size class
	AllocCount uint16  // 已分配对象数
	ElemSize   uintptr // 对象大小
	AllocBits  []bool  // 分配位图
	FreeIndex  uint16  // 下一个空闲对象索引
	State      string  // 状态: "idle", "inuse", "manual", "dead"
	mutex      sync.Mutex

	// SAFETY: 保持底层内存存活，防止GC回收
	// 在真实的Go运行时中，内存由系统调用分配，不受GC管理
	backingMemory interface{} // 保持引用以防止GC
}

// MockCache 模拟mcache结构
type MockCache struct {
	TinyAllocs []unsafe.Pointer // tiny对象分配
	Alloc      []*MockSpan      // 每个size class的span
	mutex      sync.Mutex
}

// MockCentral 模拟mcentral结构
type MockCentral struct {
	SizeClass    uint8
	NonEmptyList []*MockSpan // 非空span列表
	EmptyList    []*MockSpan // 空span列表
	mutex        sync.Mutex
}

// MockHeap 模拟mheap结构
type MockHeap struct {
	Centrals   []*MockCentral // 每个size class的central
	FreeList   [][]*MockSpan  // 空闲span列表，按页数索引
	LargeSpans []*MockSpan    // 大对象span
	AllSpans   []*MockSpan    // 所有span
	SpansInUse int64          // 使用中的span数
	PagesInUse int64          // 使用中的页数
	PageSize   int            // 页大小
	ArenaSize  int64          // arena大小
	mutex      sync.RWMutex
}

// MemoryAllocator 内存分配器
type MemoryAllocator struct {
	heap       *MockHeap
	caches     map[int]*MockCache // 每个P的cache
	sizeMapper *SizeClassMapper
	nextSpanID int64
	allocCount int64
	freeCount  int64
	mutex      sync.Mutex
}

func NewMemoryAllocator() *MemoryAllocator {
	allocator := &MemoryAllocator{
		heap:       NewMockHeap(),
		caches:     make(map[int]*MockCache),
		sizeMapper: NewSizeClassMapper(),
		nextSpanID: 1,
	}

	// 为每个P创建cache
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		allocator.caches[i] = NewMockCache()
	}

	return allocator
}

func NewMockHeap() *MockHeap {
	numSizeClasses := 67 // Go 1.20的size class数量

	heap := &MockHeap{
		Centrals:   make([]*MockCentral, numSizeClasses),
		FreeList:   make([][]*MockSpan, 128), // 最多128页
		LargeSpans: make([]*MockSpan, 0),
		AllSpans:   make([]*MockSpan, 0),
		PageSize:   8192,     // 8KB页
		ArenaSize:  64 << 20, // 64MB arena
	}

	// 初始化centrals
	for i := 0; i < numSizeClasses; i++ {
		heap.Centrals[i] = &MockCentral{
			SizeClass:    uint8(i),
			NonEmptyList: make([]*MockSpan, 0),
			EmptyList:    make([]*MockSpan, 0),
		}
	}

	return heap
}

func NewMockCache() *MockCache {
	return &MockCache{
		TinyAllocs: make([]unsafe.Pointer, 0),
		Alloc:      make([]*MockSpan, 67), // 每个size class一个span
	}
}

// AllocateObject 分配对象
func (ma *MemoryAllocator) AllocateObject(size int) unsafe.Pointer {
	atomic.AddInt64(&ma.allocCount, 1)

	// 1. 确定size class
	sizeClass, classInfo := ma.sizeMapper.FindSizeClass(size)

	if sizeClass == 0 {
		// 大对象分配
		return ma.allocateLargeObject(size)
	}

	if size <= 16 && size <= 8 {
		// tiny对象分配
		return ma.allocateTinyObject(size)
	}

	// 2. 小对象分配
	return ma.allocateSmallObject(sizeClass, classInfo)
}

func (ma *MemoryAllocator) allocateTinyObject(size int) unsafe.Pointer {
	// 模拟tiny对象分配
	// 在实际Go中，tiny对象被打包到16字节的块中
	pid := runtime.GOMAXPROCS(0) % len(ma.caches) // 简化的P选择
	cache := ma.caches[pid]

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// 分配16字节的tiny块
	ptr := unsafe.Pointer(&[16]byte{})
	cache.TinyAllocs = append(cache.TinyAllocs, ptr)

	return ptr
}

func (ma *MemoryAllocator) allocateSmallObject(sizeClass int, classInfo SizeClassInfo) unsafe.Pointer {
	pid := runtime.GOMAXPROCS(0) % len(ma.caches) // 简化的P选择
	cache := ma.caches[pid]

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// 从cache获取span
	span := cache.Alloc[sizeClass-1]
	if span == nil || span.FreeIndex >= uint16(len(span.AllocBits)) {
		// 从central获取新的span
		span = ma.getSpanFromCentral(sizeClass, classInfo)
		cache.Alloc[sizeClass-1] = span
	}

	if span == nil {
		return nil
	}

	// 从span分配对象
	return ma.allocateFromSpan(span)
}

func (ma *MemoryAllocator) allocateLargeObject(size int) unsafe.Pointer {
	// 大对象直接从heap分配
	pages := (size + ma.heap.PageSize - 1) / ma.heap.PageSize

	ma.heap.mutex.Lock()
	defer ma.heap.mutex.Unlock()

	// #nosec G103 - 教学演示：模拟Go内存分配器的大对象分配
	// 在真实的Go runtime中，大对象通过系统调用（如mmap）直接分配
	// 这里使用unsafe.Pointer获取Go slice的底层内存地址来模拟系统内存
	// 创建底层内存并保持引用
	backingMem := make([]byte, size)
	memPtr := unsafe.Pointer(&backingMem[0])

	// 创建新的大对象span
	span := &MockSpan{
		StartAddr:     uintptr(memPtr), // 从真实指针获取地址
		NPPages:       pages,
		SizeClass:     0, // 大对象的size class为0
		ElemSize:      uintptr(size),
		AllocCount:    1,
		AllocBits:     []bool{true},
		State:         "inuse",
		backingMemory: backingMem, // SAFETY: 保持内存引用活跃
	}

	ma.heap.LargeSpans = append(ma.heap.LargeSpans, span)
	ma.heap.AllSpans = append(ma.heap.AllSpans, span)
	ma.heap.PagesInUse += int64(pages)

	// SAFETY: 直接返回原始指针，避免uintptr->unsafe.Pointer转换
	// 在真实实现中，这里会是系统分配的内存
	return memPtr
}

func (ma *MemoryAllocator) getSpanFromCentral(sizeClass int, classInfo SizeClassInfo) *MockSpan {
	central := ma.heap.Centrals[sizeClass-1]

	central.mutex.Lock()
	defer central.mutex.Unlock()

	// 从非空列表获取span
	if len(central.NonEmptyList) > 0 {
		span := central.NonEmptyList[0]
		central.NonEmptyList = central.NonEmptyList[1:]
		return span
	}

	// 从空列表获取span并重新填充
	if len(central.EmptyList) > 0 {
		span := central.EmptyList[0]
		central.EmptyList = central.EmptyList[1:]
		ma.refillSpan(span, classInfo)
		return span
	}

	// 从heap获取新的span
	return ma.getSpanFromHeap(sizeClass, classInfo)
}

func (ma *MemoryAllocator) getSpanFromHeap(sizeClass int, classInfo SizeClassInfo) *MockSpan {
	ma.heap.mutex.Lock()
	defer ma.heap.mutex.Unlock()

	pages := classInfo.Pages
	if pages >= len(ma.heap.FreeList) {
		pages = len(ma.heap.FreeList) - 1
	}

	// 从free list获取span
	for p := pages; p < len(ma.heap.FreeList); p++ {
		if len(ma.heap.FreeList[p]) > 0 {
			span := ma.heap.FreeList[p][0]
			ma.heap.FreeList[p] = ma.heap.FreeList[p][1:]

			// 如果span太大，分割它
			if span.NPPages > pages {
				newSpan := ma.splitSpan(span, pages)
				return newSpan
			}

			ma.refillSpan(span, classInfo)
			return span
		}
	}

	// 创建新的span
	return ma.createNewSpan(sizeClass, classInfo)
}

func (ma *MemoryAllocator) splitSpan(span *MockSpan, pages int) *MockSpan {
	// 分割span
	newSpan := &MockSpan{
		StartAddr: span.StartAddr,
		NPPages:   pages,
		State:     "inuse",
	}

	// 更新原span
	span.StartAddr += uintptr(pages * ma.heap.PageSize)
	span.NPPages -= pages

	// 将剩余部分放回free list
	if span.NPPages > 0 {
		span.State = "idle"
		listIndex := span.NPPages
		if listIndex >= len(ma.heap.FreeList) {
			listIndex = len(ma.heap.FreeList) - 1
		}
		ma.heap.FreeList[listIndex] = append(ma.heap.FreeList[listIndex], span)
	}

	ma.heap.AllSpans = append(ma.heap.AllSpans, newSpan)
	return newSpan
}

func (ma *MemoryAllocator) createNewSpan(sizeClass int, classInfo SizeClassInfo) *MockSpan {
	// #nosec G103 - 教学演示：模拟Go内存分配器的span创建过程
	// 在真实的Go runtime中，span内存通过arena分配器从操作系统获取
	// 这里演示了span如何管理原始内存地址
	// 为演示目的创建实际的backing memory
	// 在真实的Go运行时中，这里会调用系统分配函数
	memSize := classInfo.Pages * ma.heap.PageSize
	backingMem := make([]byte, memSize)
	memPtr := unsafe.Pointer(&backingMem[0])

	span := &MockSpan{
		StartAddr:     uintptr(memPtr), // 使用真实内存地址
		NPPages:       classInfo.Pages,
		SizeClass:     uint8(sizeClass),
		ElemSize:      uintptr(classInfo.Size),
		State:         "inuse",
		backingMemory: backingMem, // SAFETY: 保持内存引用活跃
	}

	ma.refillSpan(span, classInfo)
	ma.heap.AllSpans = append(ma.heap.AllSpans, span)
	ma.heap.SpansInUse++
	ma.heap.PagesInUse += int64(classInfo.Pages)

	return span
}

func (ma *MemoryAllocator) refillSpan(span *MockSpan, classInfo SizeClassInfo) {
	span.AllocBits = make([]bool, classInfo.Objects)
	span.FreeIndex = 0
	span.AllocCount = 0
	span.ElemSize = uintptr(classInfo.Size)
	span.SizeClass = uint8(classInfo.Size)
}

func (ma *MemoryAllocator) allocateFromSpan(span *MockSpan) unsafe.Pointer {
	span.mutex.Lock()
	defer span.mutex.Unlock()

	// 找到下一个空闲对象
	for span.FreeIndex < uint16(len(span.AllocBits)) {
		if !span.AllocBits[span.FreeIndex] {
			span.AllocBits[span.FreeIndex] = true
			span.AllocCount++

			// #nosec G103 - 教学演示：展示Go内存分配器如何从span中分配对象
			// 在真实的Go runtime中，分配器通过计算偏移量来定位空闲对象
			// 这里演示了安全的指针算术：先转换为slice索引，再获取指针
			// SAFETY: 在真实实现中需要确保指针算术的安全性
			// 这里为了演示目的，我们检查backingMemory是否存在
			if span.backingMemory != nil {
				// 如果有实际的backing memory，使用安全的指针算术
				if backingSlice, ok := span.backingMemory.([]byte); ok {
					offset := span.FreeIndex * uint16(span.ElemSize)
					if int(offset) < len(backingSlice) {
						span.FreeIndex++
						return unsafe.Pointer(&backingSlice[offset])
					}
				}
			}

			// 回退到模拟地址（仅用于演示，实际使用中有GC安全问题）
			// WARNING: 这种模式在生产代码中是不安全的
			// 在真实的内存分配器中，应该始终有backing memory

			// 为了避免unsafe.Pointer警告，在教学代码中我们选择返回nil
			// 而不是进行不安全的uintptr到unsafe.Pointer转换
			// 生产代码应该确保所有span都有有效的backingMemory

			// objAddr := span.StartAddr + uintptr(span.FreeIndex)*span.ElemSize
			// defer runtime.KeepAlive(span)
			// return unsafe.Pointer(objAddr) // UNSAFE: 不推荐的模式

			return nil // 安全的回退：拒绝分配而非执行不安全操作
		}
		span.FreeIndex++
	}

	return nil // span已满
}

// FreeObject 释放对象
func (ma *MemoryAllocator) FreeObject(ptr unsafe.Pointer, size int) {
	atomic.AddInt64(&ma.freeCount, 1)

	// 在实际Go中，free操作非常复杂
	// 这里只是一个简化的模拟
	fmt.Printf("释放对象: %p, 大小: %d\n", ptr, size)
}

func (ma *MemoryAllocator) GetStats() (int64, int64, int64, int64) {
	ma.heap.mutex.RLock()
	spansInUse := ma.heap.SpansInUse
	pagesInUse := ma.heap.PagesInUse
	ma.heap.mutex.RUnlock()

	allocCount := atomic.LoadInt64(&ma.allocCount)
	freeCount := atomic.LoadInt64(&ma.freeCount)

	return allocCount, freeCount, spansInUse, pagesInUse
}

// ==================
// 4. 栈内存管理演示
// ==================

// StackManager 栈管理器
type StackManager struct {
	stackSize        int
	maxStackSize     int
	stackGrowCount   int64
	stackShrinkCount int64
}

func NewStackManager() *StackManager {
	return &StackManager{
		stackSize:    2048,    // 2KB初始栈
		maxStackSize: 1 << 20, // 1MB最大栈
	}
}

func (sm *StackManager) demonstrateStackGrowth() {
	fmt.Println("\n=== 栈内存管理演示 ===")

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	// 触发栈增长
	sm.deepRecursion(20)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	fmt.Printf("栈增长前后对比:\n")
	fmt.Printf("  栈系统内存: %d KB -> %d KB\n",
		before.StackSys/1024, after.StackSys/1024)
	fmt.Printf("  栈使用内存: %d KB -> %d KB\n",
		before.StackInuse/1024, after.StackInuse/1024)
}

func (sm *StackManager) deepRecursion(depth int) int {
	if depth <= 0 {
		return 0
	}

	// 创建大的栈帧
	var largeArray [1024]int
	largeArray[0] = depth

	// 模拟栈增长检查
	if depth%5 == 0 {
		atomic.AddInt64(&sm.stackGrowCount, 1)
		fmt.Printf("栈增长检查 - 深度: %d\n", depth)
	}

	return largeArray[0] + sm.deepRecursion(depth-1)
}

// ==================
// 5. 内存对齐演示
// ==================

// AlignmentDemo 内存对齐演示
type AlignmentDemo struct{}

func (ad *AlignmentDemo) demonstrateAlignment() {
	fmt.Println("\n=== 内存对齐演示 ===")

	// 展示不同类型的对齐
	var (
		b   bool
		i8  int8
		i16 int16
		i32 int32
		i64 int64
		f32 float32
		f64 float64
		ptr unsafe.Pointer
	)

	// #nosec G103 - 教学演示：展示Go类型系统的内存对齐规则
	// unsafe.Sizeof和unsafe.Alignof用于查询类型的内存布局信息
	// 这些信息对于理解内存分配器的行为和优化结构体布局至关重要
	fmt.Printf("基本类型对齐:\n")
	fmt.Printf("  bool:    %d字节, 对齐: %d\n", unsafe.Sizeof(b), unsafe.Alignof(b))
	fmt.Printf("  int8:    %d字节, 对齐: %d\n", unsafe.Sizeof(i8), unsafe.Alignof(i8))
	fmt.Printf("  int16:   %d字节, 对齐: %d\n", unsafe.Sizeof(i16), unsafe.Alignof(i16))
	fmt.Printf("  int32:   %d字节, 对齐: %d\n", unsafe.Sizeof(i32), unsafe.Alignof(i32))
	fmt.Printf("  int64:   %d字节, 对齐: %d\n", unsafe.Sizeof(i64), unsafe.Alignof(i64))
	fmt.Printf("  float32: %d字节, 对齐: %d\n", unsafe.Sizeof(f32), unsafe.Alignof(f32))
	fmt.Printf("  float64: %d字节, 对齐: %d\n", unsafe.Sizeof(f64), unsafe.Alignof(f64))
	fmt.Printf("  pointer: %d字节, 对齐: %d\n", unsafe.Sizeof(ptr), unsafe.Alignof(ptr))

	// 展示结构体对齐
	type BadStruct struct {
		a bool  // 1字节
		b int64 // 8字节
		c bool  // 1字节
		d int32 // 4字节
	}

	type GoodStruct struct {
		b int64 // 8字节
		d int32 // 4字节
		a bool  // 1字节
		c bool  // 1字节
		// 自动填充2字节
	}

	var bad BadStruct
	var good GoodStruct

	// #nosec G103 - 教学演示：展示结构体字段对齐和内存布局优化
	// unsafe.Sizeof和unsafe.Offsetof用于分析结构体的内存布局
	// 这对于性能优化（减少缓存未命中）和内存节省非常重要
	fmt.Printf("\n结构体对齐:\n")
	fmt.Printf("  BadStruct:  %d字节 (内存浪费)\n", unsafe.Sizeof(bad))
	fmt.Printf("  GoodStruct: %d字节 (优化对齐)\n", unsafe.Sizeof(good))

	// 展示字段偏移
	fmt.Printf("\nBadStruct字段偏移:\n")
	fmt.Printf("  a: %d字节偏移\n", unsafe.Offsetof(bad.a))
	fmt.Printf("  b: %d字节偏移\n", unsafe.Offsetof(bad.b))
	fmt.Printf("  c: %d字节偏移\n", unsafe.Offsetof(bad.c))
	fmt.Printf("  d: %d字节偏移\n", unsafe.Offsetof(bad.d))

	fmt.Printf("\nGoodStruct字段偏移:\n")
	fmt.Printf("  b: %d字节偏移\n", unsafe.Offsetof(good.b))
	fmt.Printf("  d: %d字节偏移\n", unsafe.Offsetof(good.d))
	fmt.Printf("  a: %d字节偏移\n", unsafe.Offsetof(good.a))
	fmt.Printf("  c: %d字节偏移\n", unsafe.Offsetof(good.c))
}

// ==================
// 6. 主演示函数
// ==================

func demonstrateMemoryAllocator() {
	fmt.Println("=== Go内存分配器深度解析 ===")

	// 1. 启动内存监控
	monitor := NewMemoryMonitor()
	monitor.Start()
	defer monitor.Stop()

	fmt.Println("\n1. 内存分配器基本信息")
	stats := monitor.GetLatestStats()
	fmt.Printf("堆大小: %d KB\n", stats.HeapSys/1024)
	fmt.Printf("堆使用: %d KB\n", stats.HeapInuse/1024)
	fmt.Printf("堆对象: %d 个\n", stats.HeapObjects)
	fmt.Printf("栈大小: %d KB\n", stats.StackSys/1024)

	// 2. Size Class映射演示
	fmt.Println("\n2. Size Class映射演示")
	sizeMapper := NewSizeClassMapper()
	sizeMapper.PrintSizeClasses()

	// 测试不同大小的对象映射
	testSizes := []int{8, 17, 64, 128, 1024, 4096, 32768, 65536}
	fmt.Printf("\n对象大小到Size Class的映射:\n")
	for _, size := range testSizes {
		class, info := sizeMapper.FindSizeClass(size)
		if class == 0 {
			fmt.Printf("  %5d 字节 -> 大对象 (需要 %d 页)\n", size, info.Pages)
		} else {
			fmt.Printf("  %5d 字节 -> Class %2d (实际分配 %d 字节)\n",
				size, class, info.Size)
		}
	}

	// 3. 内存分配器模拟
	fmt.Println("\n3. 内存分配器模拟")
	allocator := NewMemoryAllocator()

	// 分配不同大小的对象
	fmt.Println("分配各种大小的对象...")

	// 小对象
	for i := 0; i < 100; i++ {
		size := 8 + (i%16)*8 // 8, 16, 24, ..., 128
		ptr := allocator.AllocateObject(size)
		if ptr == nil {
			fmt.Printf("分配失败: %d 字节\n", size)
		}
	}

	// 中等对象
	for i := 0; i < 50; i++ {
		size := 1024 + (i%8)*1024 // 1KB, 2KB, ..., 8KB
		ptr := allocator.AllocateObject(size)
		if ptr == nil {
			fmt.Printf("分配失败: %d 字节\n", size)
		}
	}

	// 大对象
	for i := 0; i < 10; i++ {
		size := 32768 + i*8192 // 32KB, 40KB, 48KB, ...
		ptr := allocator.AllocateObject(size)
		if ptr == nil {
			fmt.Printf("分配失败: %d 字节\n", size)
		}
	}

	allocCount, freeCount, spansInUse, pagesInUse := allocator.GetStats()
	fmt.Printf("\n分配统计:\n")
	fmt.Printf("  分配次数: %d\n", allocCount)
	fmt.Printf("  释放次数: %d\n", freeCount)
	fmt.Printf("  使用中span: %d\n", spansInUse)
	fmt.Printf("  使用中页数: %d\n", pagesInUse)

	// 4. 栈内存管理演示
	stackManager := NewStackManager()
	stackManager.demonstrateStackGrowth()

	// 5. 内存对齐演示
	alignmentDemo := &AlignmentDemo{}
	alignmentDemo.demonstrateAlignment()

	// 6. 真实内存分配测试
	fmt.Println("\n4. 真实内存分配测试")
	demonstrateRealAllocation()

	// 7. 最终统计
	fmt.Println("\n=== 最终内存统计 ===")
	finalStats := monitor.GetLatestStats()
	fmt.Printf("最终堆大小: %d KB\n", finalStats.HeapSys/1024)
	fmt.Printf("最终堆使用: %d KB\n", finalStats.HeapInuse/1024)
	fmt.Printf("最终堆对象: %d 个\n", finalStats.HeapObjects)
	fmt.Printf("累计分配: %d 次\n", finalStats.Mallocs)
	fmt.Printf("累计释放: %d 次\n", finalStats.Frees)
	fmt.Printf("累计分配字节: %d KB\n", finalStats.TotalAlloc/1024)
}

func demonstrateRealAllocation() {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	// 分配大量小对象
	objects := make([][]byte, 1000)
	for i := range objects {
		objects[i] = make([]byte, 128) // 128字节对象
	}

	runtime.ReadMemStats(&after)
	fmt.Printf("分配1000个128字节对象:\n")
	fmt.Printf("  堆增长: %d KB\n", (after.HeapInuse-before.HeapInuse)/1024)
	fmt.Printf("  对象增加: %d 个\n", after.HeapObjects-before.HeapObjects)
	fmt.Printf("  分配次数增加: %d\n", after.Mallocs-before.Mallocs)

	// 分配大对象
	runtime.ReadMemStats(&before)
	largeObjects := make([][]byte, 10)
	for i := range largeObjects {
		largeObjects[i] = make([]byte, 1024*1024) // 1MB对象
	}

	runtime.ReadMemStats(&after)
	fmt.Printf("\n分配10个1MB大对象:\n")
	fmt.Printf("  堆增长: %d KB\n", (after.HeapInuse-before.HeapInuse)/1024)
	fmt.Printf("  对象增加: %d 个\n", after.HeapObjects-before.HeapObjects)
	fmt.Printf("  分配次数增加: %d\n", after.Mallocs-before.Mallocs)

	// 触发GC观察内存回收
	runtime.GC()
	runtime.ReadMemStats(&after)
	fmt.Printf("\nGC后堆使用: %d KB\n", after.HeapInuse/1024)

	// 清除引用让对象可被回收
	objects = nil
	largeObjects = nil
	runtime.GC()
	runtime.GC() // 确保完成

	var final runtime.MemStats
	runtime.ReadMemStats(&final)
	fmt.Printf("清除引用并GC后堆使用: %d KB\n", final.HeapInuse/1024)
}

func main() {
	demonstrateMemoryAllocator()

	fmt.Println("\n=== Go内存分配器深度解析完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. TCMalloc架构：mheap->mcentral->mcache->mspan层次结构")
	fmt.Println("2. Size Class映射：优化内存分配效率和减少碎片")
	fmt.Println("3. 分层分配策略：tiny/小/大对象采用不同分配路径")
	fmt.Println("4. 栈内存管理：自动增长和收缩机制")
	fmt.Println("5. 内存对齐：确保性能和正确性")
	fmt.Println("6. 无锁算法：减少分配器锁竞争")

	fmt.Println("\n高级特性:")
	fmt.Println("- mspan位图管理对象分配状态")
	fmt.Println("- mcache提供无锁的P本地分配")
	fmt.Println("- mcentral协调多个P之间的span共享")
	fmt.Println("- mheap管理页级别的内存分配")
	fmt.Println("- 大对象直接从heap分配，绕过cache")
	fmt.Println("- 栈分割和栈拷贝优化栈内存使用")
}

/*
=== 练习题 ===

1. 基础练习：
   - 实现简单的对象池
   - 分析不同大小对象的分配性能
   - 测量内存碎片化程度
   - 实现自定义的size class映射

2. 中级练习：
   - 实现内存分配器的性能基准测试
   - 分析栈溢出和栈收缩的条件
   - 优化结构体内存布局
   - 实现内存使用情况的可视化

3. 高级练习：
   - 实现NUMA感知的内存分配器
   - 分析内存分配器的可扩展性
   - 优化高并发下的分配性能
   - 实现内存压缩和整理

4. 性能分析：
   - 使用pprof分析内存分配热点
   - 测量内存分配器的延迟
   - 分析内存分配对GC的影响
   - 优化内存分配密集型应用

5. 深度研究：
   - 研究其他语言的内存分配器差异
   - 实现用户态内存管理器
   - 分析虚拟内存管理
   - 研究持久化内存分配器

运行命令：
go run main.go

环境变量：
export GODEBUG=allocfreetrace=1 # 跟踪内存分配
export GODEBUG=madvdontneed=1   # 内存归还策略

重要概念：
- mspan: 管理特定size class的内存页
- mcache: P本地的span缓存
- mcentral: 特定size class的span中心分配器
- mheap: 页级别的堆管理器
- Size Class: 预定义的对象大小类别
- Page: 8KB的内存页（Go 1.20+）
- Arena: 64MB的内存区域（64位系统）
*/
