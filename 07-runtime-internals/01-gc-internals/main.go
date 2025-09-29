/*
=== Go运行时内核：垃圾收集器深度解析 ===

本模块深入Go语言垃圾收集器(GC)的内核实现，探索：
1. 三色标记算法原理和实现
2. 并发垃圾收集机制
3. 写屏障(Write Barrier)技术
4. GC触发机制和调优
5. 内存分配和回收策略
6. GC性能分析和优化
7. 堆栈扫描技术
8. 根对象识别
9. 弱引用和终结器

学习目标：
- 理解Go GC的三色标记并发算法
- 掌握GC调优参数和策略
- 学会GC性能分析和问题诊断
- 深入了解内存管理机制
*/

package main

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"go-mastery/common/security"
)

// ==================
// 1. GC统计和监控
// ==================

// 常量定义
const (
	MaxStatsHistory    = 1000
	DefaultMonitorTick = 100 * time.Millisecond
	DefaultAllocTick   = 10 * time.Millisecond
	SmallObjectSize    = 1024
	LargeObjectSize    = 1024 * 1024
	DefaultAllocBatch  = 100
	StackDemoDepth     = 5
	StackFrameSize     = 1024
)

// GCStats 包装runtime.MemStats提供更友好的GC统计信息
type GCStats struct {
	NumGC        uint32        // GC次数
	NumForcedGC  uint32        // 强制GC次数
	GCPauseTotal time.Duration // GC总暂停时间
	GCPauseMax   time.Duration // 最大GC暂停时间
	GCPauseMin   time.Duration // 最小GC暂停时间
	GCPauseAvg   time.Duration // 平均GC暂停时间
	HeapSize     uint64        // 堆大小
	HeapUsed     uint64        // 堆使用量
	HeapObjects  uint64        // 堆对象数
	StackSize    uint64        // 栈大小
	NextGC       uint64        // 下次GC触发阈值
	LastGCTime   time.Time     // 上次GC时间
	GCPercent    int           // GC目标百分比
}

// GCMonitor GC监控器
type GCMonitor struct {
	stats     []GCStats
	startTime time.Time
	mutex     sync.RWMutex
	running   atomic.Bool
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewGCMonitor() *GCMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &GCMonitor{
		stats:     make([]GCStats, 0, MaxStatsHistory),
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (m *GCMonitor) Start() {
	if m.running.CompareAndSwap(false, true) {
		go m.monitor()
	}
}

func (m *GCMonitor) Stop() {
	if m.running.CompareAndSwap(true, false) {
		m.cancel()
	}
}

func (m *GCMonitor) monitor() {
	ticker := time.NewTicker(DefaultMonitorTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.collectStats()
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *GCMonitor) collectStats() {
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
		GCPercent:    debug.SetGCPercent(-1), // 获取当前设置
	}

	// 计算GC暂停时间统计
	if ms.NumGC > 0 {
		pauseHistory := ms.PauseNs[:]

		var total, maxPause, minPause time.Duration
		minPause = time.Duration(math.MaxInt64)

		// 只统计最近的暂停时间
		recentPauses := min(int(ms.NumGC), len(pauseHistory))
		for i := range recentPauses {
			pause := time.Duration(pauseHistory[i])
			total += pause
			if pause > maxPause {
				maxPause = pause
			}
			if pause < minPause && pause > 0 {
				minPause = pause
			}
		}

		if recentPauses > 0 {
			stats.GCPauseMax = maxPause
			stats.GCPauseMin = minPause
			stats.GCPauseAvg = total / time.Duration(recentPauses)
		}
	}

	// 获取最后GC时间
	if ms.NumGC > 0 {
		stats.LastGCTime = time.Unix(0, security.MustSafeUint64ToInt64(ms.LastGC))
	}

	// 恢复GC百分比设置
	debug.SetGCPercent(stats.GCPercent)

	m.mutex.Lock()
	m.stats = append(m.stats, stats)
	// 保持最近MaxStatsHistory个统计数据
	if len(m.stats) > MaxStatsHistory {
		m.stats = slices.Delete(m.stats, 0, len(m.stats)-MaxStatsHistory)
	}
	m.mutex.Unlock()
}

func (m *GCMonitor) GetLatestStats() GCStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.stats) == 0 {
		return GCStats{}
	}

	return m.stats[len(m.stats)-1]
}

func (m *GCMonitor) GetStatsHistory() []GCStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]GCStats, len(m.stats))
	copy(history, m.stats)
	return history
}

// ==================
// 2. 三色标记算法演示
// ==================

// ObjectColor 对象颜色
type ObjectColor int

const (
	White ObjectColor = iota // 白色：未标记，可能被回收
	Gray                     // 灰色：已标记但子对象未扫描
	Black                    // 黑色：已标记且子对象已扫描
)

func (c ObjectColor) String() string {
	switch c {
	case White:
		return "White"
	case Gray:
		return "Gray"
	case Black:
		return "Black"
	default:
		return "Unknown"
	}
}

// MockObject 模拟对象用于演示三色标记
type MockObject struct {
	ID       int
	Color    ObjectColor
	Children []*MockObject
	Data     []byte // 模拟对象数据
}

// TricolorGC 三色标记垃圾收集器演示
type TricolorGC struct {
	objects   []*MockObject
	roots     []*MockObject
	grayQueue []*MockObject
	marked    map[*MockObject]bool
	phase     string
}

func NewTricolorGC() *TricolorGC {
	return &TricolorGC{
		objects: make([]*MockObject, 0),
		roots:   make([]*MockObject, 0),
		marked:  make(map[*MockObject]bool),
		phase:   "Idle",
	}
}

func (gc *TricolorGC) AddObject(id int, dataSize int) *MockObject {
	obj := &MockObject{
		ID:       id,
		Color:    White,
		Children: make([]*MockObject, 0),
		Data:     make([]byte, dataSize),
	}
	gc.objects = append(gc.objects, obj)
	return obj
}

func (gc *TricolorGC) AddRoot(obj *MockObject) {
	gc.roots = append(gc.roots, obj)
}

func (gc *TricolorGC) AddReference(parent, child *MockObject) {
	parent.Children = append(parent.Children, child)
}

// Mark 标记阶段
func (gc *TricolorGC) Mark() {
	fmt.Println("=== 三色标记算法演示 ===")
	gc.phase = "Mark"

	// 1. 初始化：所有对象为白色
	for _, obj := range gc.objects {
		obj.Color = White
	}
	gc.grayQueue = make([]*MockObject, 0)
	gc.marked = make(map[*MockObject]bool)

	fmt.Printf("初始状态：%d个对象全部为白色\n", len(gc.objects))
	gc.printState()

	// 2. 标记根对象为灰色
	fmt.Println("\n--- 标记根对象 ---")
	for _, root := range gc.roots {
		if root.Color == White {
			root.Color = Gray
			gc.grayQueue = append(gc.grayQueue, root)
			fmt.Printf("根对象 %d 标记为灰色\n", root.ID)
		}
	}
	gc.printState()

	// 3. 处理灰色队列
	fmt.Println("\n--- 处理灰色对象队列 ---")
	step := 1
	for len(gc.grayQueue) > 0 {
		fmt.Printf("\n第%d步：", step)
		obj := gc.grayQueue[0]
		gc.grayQueue = slices.Delete(gc.grayQueue, 0, 1)

		// 扫描子对象
		fmt.Printf("扫描对象 %d 的子对象\n", obj.ID)
		for _, child := range obj.Children {
			if child.Color == White {
				child.Color = Gray
				gc.grayQueue = append(gc.grayQueue, child)
				fmt.Printf("  子对象 %d 标记为灰色\n", child.ID)
			}
		}

		// 对象变为黑色
		obj.Color = Black
		gc.marked[obj] = true
		fmt.Printf("  对象 %d 标记为黑色\n", obj.ID)

		gc.printState()
		step++
	}

	fmt.Println("\n--- 标记阶段完成 ---")
	gc.printFinalState()
}

// Sweep 清除阶段
func (gc *TricolorGC) Sweep() {
	fmt.Println("\n=== 清除阶段 ===")
	gc.phase = "Sweep"

	freed := 0
	totalSize := 0

	for i := len(gc.objects) - 1; i >= 0; i-- {
		obj := gc.objects[i]
		if obj.Color == White {
			// 白色对象被回收
			fmt.Printf("回收对象 %d (大小: %d字节)\n", obj.ID, len(obj.Data))
			totalSize += len(obj.Data)
			freed++

			// 从列表中移除
			gc.objects = slices.Delete(gc.objects, i, i+1)
		} else {
			// 重置颜色为白色，准备下次GC
			obj.Color = White
		}
	}

	fmt.Printf("\n清除完成：回收 %d 个对象，释放 %d 字节内存\n", freed, totalSize)
	fmt.Printf("剩余对象：%d 个\n", len(gc.objects))

	gc.phase = "Idle"
}

func (gc *TricolorGC) printState() {
	white, gray, black := 0, 0, 0
	for _, obj := range gc.objects {
		switch obj.Color {
		case White:
			white++
		case Gray:
			gray++
		case Black:
			black++
		}
	}
	fmt.Printf("  当前状态 - 白色:%d, 灰色:%d, 黑色:%d, 灰色队列:%d\n",
		white, gray, black, len(gc.grayQueue))
}

func (gc *TricolorGC) printFinalState() {
	fmt.Println("对象标记结果：")
	for _, obj := range gc.objects {
		status := "存活"
		if obj.Color == White {
			status = "待回收"
		}
		fmt.Printf("  对象 %d: %s (%s)\n", obj.ID, obj.Color, status)
	}
}

// ==================
// 3. 并发GC机制演示
// ==================

// ConcurrentGCDemo 并发GC演示
type ConcurrentGCDemo struct {
	allocatedObjects [][]byte
	allocMutex       sync.Mutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

func NewConcurrentGCDemo() *ConcurrentGCDemo {
	ctx, cancel := context.WithCancel(context.Background())
	return &ConcurrentGCDemo{
		allocatedObjects: make([][]byte, 0),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (d *ConcurrentGCDemo) StartAllocation() {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		objSize := 1024 // 1KB per object

		for {
			select {
			case <-ticker.C:
				// 分配新对象
				obj := make([]byte, objSize)

				d.allocMutex.Lock()
				d.allocatedObjects = append(d.allocatedObjects, obj)

				// 随机释放一些对象
				if len(d.allocatedObjects) > 100 {
					// 释放前半部分对象
					d.allocatedObjects = slices.Delete(d.allocatedObjects, 0, 50)
				}
				d.allocMutex.Unlock()

			case <-d.ctx.Done():
				return
			}
		}
	}()
}

func (d *ConcurrentGCDemo) Stop() {
	d.cancel()
	d.wg.Wait()
}

// WriteBarrierDemo 写屏障演示
func WriteBarrierDemo() {
	fmt.Println("\n=== 写屏障机制演示 ===")

	// 创建对象图来演示写屏障
	parent := &MockObject{ID: 1, Color: Black} // 已扫描的黑色对象
	child := &MockObject{ID: 2, Color: White}  // 未扫描的白色对象

	fmt.Printf("初始状态：父对象%d (黑色), 子对象%d (白色)\n", parent.ID, child.ID)

	// 模拟并发修改：黑色对象指向白色对象
	fmt.Println("检测到黑色对象指向白色对象，触发写屏障...")

	// 写屏障的响应：将白色对象标记为灰色
	if parent.Color == Black && child.Color == White {
		child.Color = Gray
		fmt.Printf("写屏障响应：子对象%d 标记为灰色\n", child.ID)
	}

	fmt.Println("写屏障确保了三色不变性：黑色对象不直接指向白色对象")
}

// ==================
// 4. GC触发和调优
// ==================

// GCTuner GC调优器
type GCTuner struct {
	originalGCPercent int
	targetLatency     time.Duration
	maxPauseTime      time.Duration
}

func NewGCTuner() *GCTuner {
	return &GCTuner{
		originalGCPercent: debug.SetGCPercent(-1),
		targetLatency:     time.Millisecond * 2,  // 目标延迟2ms
		maxPauseTime:      time.Millisecond * 10, // 最大暂停10ms
	}
}

func (t *GCTuner) TuneForLatency() {
	fmt.Println("\n=== GC延迟调优 ===")
	fmt.Printf("原始GOGC: %d%%\n", t.originalGCPercent)

	// 降低GOGC以减少GC暂停时间
	newGCPercent := 50 // 更频繁的GC
	debug.SetGCPercent(newGCPercent)
	fmt.Printf("调整GOGC为: %d%% (更频繁GC，减少暂停时间)\n", newGCPercent)

	// 设置软内存限制
	if debug.SetMemoryLimit(-1) == math.MaxInt64 {
		// 设置500MB内存限制
		debug.SetMemoryLimit(500 * 1024 * 1024)
		fmt.Println("设置内存限制: 500MB")
	}
}

func (t *GCTuner) TuneForThroughput() {
	fmt.Println("\n=== GC吞吐量调优 ===")

	// 增加GOGC以提高吞吐量
	newGCPercent := 200 // 减少GC频率
	debug.SetGCPercent(newGCPercent)
	fmt.Printf("调整GOGC为: %d%% (减少GC频率，提高吞吐量)\n", newGCPercent)
}

func (t *GCTuner) Restore() {
	debug.SetGCPercent(t.originalGCPercent)
	debug.SetMemoryLimit(math.MaxInt64)
	fmt.Printf("恢复原始GC设置: GOGC=%d%%\n", t.originalGCPercent)
}

// ==================
// 5. 内存分配器演示
// ==================

// MemoryAllocatorDemo 内存分配器演示
type MemoryAllocatorDemo struct {
	smallObjects [][]byte // <32KB
	largeObjects [][]byte // >=32KB
	allocHistory []AllocationInfo
	mutex        sync.Mutex
}

type AllocationInfo struct {
	Size      int
	Timestamp time.Time
	Type      string // "small" or "large"
}

func NewMemoryAllocatorDemo() *MemoryAllocatorDemo {
	return &MemoryAllocatorDemo{
		smallObjects: make([][]byte, 0),
		largeObjects: make([][]byte, 0),
		allocHistory: make([]AllocationInfo, 0),
	}
}

func (d *MemoryAllocatorDemo) AllocateSmallObjects(count int) {
	fmt.Printf("\n=== 小对象分配演示 (每个 1KB) ===\n")

	for i := 0; i < count; i++ {
		obj := make([]byte, 1024) // 1KB 小对象

		d.mutex.Lock()
		d.smallObjects = append(d.smallObjects, obj)
		d.allocHistory = append(d.allocHistory, AllocationInfo{
			Size:      1024,
			Timestamp: time.Now(),
			Type:      "small",
		})
		d.mutex.Unlock()

		// 每100个对象输出一次状态
		if (i+1)%100 == 0 {
			d.printMemoryStats(fmt.Sprintf("已分配 %d 个小对象", i+1))
		}
	}
}

func (d *MemoryAllocatorDemo) AllocateLargeObjects(count int) {
	fmt.Printf("\n=== 大对象分配演示 (每个 1MB) ===\n")

	for i := 0; i < count; i++ {
		obj := make([]byte, 1024*1024) // 1MB 大对象

		d.mutex.Lock()
		d.largeObjects = append(d.largeObjects, obj)
		d.allocHistory = append(d.allocHistory, AllocationInfo{
			Size:      1024 * 1024,
			Timestamp: time.Now(),
			Type:      "large",
		})
		d.mutex.Unlock()

		d.printMemoryStats(fmt.Sprintf("已分配 %d 个大对象", i+1))
	}
}

func (d *MemoryAllocatorDemo) printMemoryStats(phase string) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	fmt.Printf("%s:\n", phase)
	fmt.Printf("  堆大小: %d KB\n", ms.HeapSys/1024)
	fmt.Printf("  堆使用: %d KB\n", ms.HeapInuse/1024)
	fmt.Printf("  堆对象: %d 个\n", ms.HeapObjects)
	fmt.Printf("  GC次数: %d\n", ms.NumGC)
	fmt.Printf("  栈大小: %d KB\n", ms.StackSys/1024)
	fmt.Println()
}

func (d *MemoryAllocatorDemo) ForceGC() {
	fmt.Println("=== 强制触发GC ===")
	before := d.getGCCount()

	runtime.GC()
	runtime.GC() // 确保完成

	after := d.getGCCount()
	fmt.Printf("GC执行完成，GC次数从 %d 增加到 %d\n", before, after)

	d.printMemoryStats("GC后状态")
}

func (d *MemoryAllocatorDemo) getGCCount() uint32 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.NumGC
}

// ==================
// 6. 堆栈扫描演示
// ==================

// StackScanDemo 堆栈扫描演示
func StackScanDemo() {
	fmt.Println("\n=== 堆栈扫描机制演示 ===")

	// 创建栈上变量
	localVar := make([]byte, 1024)
	fmt.Printf("栈上变量地址: %p\n", &localVar)

	// 获取当前goroutine的栈信息
	buf := make([]byte, 1024)
	stackSize := runtime.Stack(buf, false)

	fmt.Printf("当前goroutine栈大小: %d 字节\n", stackSize)
	fmt.Printf("栈内容片段:\n%s\n", string(buf[:min(200, stackSize)]))

	// 演示栈增长
	demonstrateStackGrowth(5)
}

func demonstrateStackGrowth(depth int) {
	if depth <= 0 {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		fmt.Printf("递归深度0，当前栈系统内存: %d KB\n", ms.StackSys/1024)
		return
	}

	// 创建大的栈帧
	largeArray := [1024]int{}
	largeArray[0] = depth

	if depth%2 == 0 {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		fmt.Printf("递归深度%d，栈系统内存: %d KB\n", depth, ms.StackSys/1024)
	}

	demonstrateStackGrowth(depth - 1)
}

// ==================
// 7. 终结器(Finalizer)演示
// ==================

// FinalizableObject 可终结的对象
type FinalizableObject struct {
	ID       int
	Resource *ExternalResource
}

// ExternalResource 外部资源
type ExternalResource struct {
	Handle uintptr
	Name   string
}

func NewFinalizableObject(id int, resourceName string) *FinalizableObject {
	obj := &FinalizableObject{
		ID: id,
		Resource: &ExternalResource{
			Handle: uintptr(unsafe.Pointer(&id)), // 模拟外部句柄
			Name:   resourceName,
		},
	}

	// 设置终结器
	runtime.SetFinalizer(obj, (*FinalizableObject).Finalize)

	fmt.Printf("创建可终结对象 %d，资源: %s\n", id, resourceName)
	return obj
}

func (obj *FinalizableObject) Finalize() {
	fmt.Printf("终结器被调用：清理对象 %d 的资源 %s\n", obj.ID, obj.Resource.Name)

	// 模拟资源清理
	obj.Resource.Handle = 0
	obj.Resource.Name = ""
}

func (obj *FinalizableObject) Close() {
	fmt.Printf("手动关闭对象 %d\n", obj.ID)

	// 清理终结器
	runtime.SetFinalizer(obj, nil)

	// 手动清理资源
	obj.Finalize()
}

// ==================
// 8. 主演示函数
// ==================

func demonstrateGCInternals() {
	fmt.Println("=== Go垃圾收集器内核深度解析 ===")

	// 1. 启动GC监控
	monitor := NewGCMonitor()
	monitor.Start()
	defer monitor.Stop()

	fmt.Println("\n1. GC监控和统计")
	time.Sleep(100 * time.Millisecond)
	stats := monitor.GetLatestStats()
	fmt.Printf("当前GC统计:\n")
	fmt.Printf("  GC次数: %d\n", stats.NumGC)
	fmt.Printf("  堆大小: %d KB\n", stats.HeapSize/1024)
	fmt.Printf("  堆使用: %d KB\n", stats.HeapUsed/1024)
	fmt.Printf("  对象数: %d\n", stats.HeapObjects)
	fmt.Printf("  下次GC阈值: %d KB\n", stats.NextGC/1024)

	// 2. 三色标记算法演示
	fmt.Println("\n2. 三色标记算法演示")
	gcDemo := NewTricolorGC()

	// 创建对象图: 1->2->3, 1->4, 5->6 (5,6不可达)
	obj1 := gcDemo.AddObject(1, 100)
	obj2 := gcDemo.AddObject(2, 200)
	obj3 := gcDemo.AddObject(3, 150)
	obj4 := gcDemo.AddObject(4, 300)
	obj5 := gcDemo.AddObject(5, 400) // 不可达
	obj6 := gcDemo.AddObject(6, 250) // 不可达

	gcDemo.AddReference(obj1, obj2)
	gcDemo.AddReference(obj2, obj3)
	gcDemo.AddReference(obj1, obj4)
	gcDemo.AddReference(obj5, obj6) // 不可达的引用

	gcDemo.AddRoot(obj1) // 只有obj1是根对象

	gcDemo.Mark()
	gcDemo.Sweep()

	// 3. 写屏障演示
	WriteBarrierDemo()

	// 4. 并发GC演示
	fmt.Println("\n3. 并发GC机制演示")
	concurrentDemo := NewConcurrentGCDemo()
	concurrentDemo.StartAllocation()

	// 运行3秒，观察并发GC
	fmt.Println("开始并发分配对象...")
	startTime := time.Now()
	initialGC := stats.NumGC

	time.Sleep(3 * time.Second)
	concurrentDemo.Stop()

	newStats := monitor.GetLatestStats()
	fmt.Printf("3秒内触发GC次数: %d\n", newStats.NumGC-initialGC)
	fmt.Printf("平均GC暂停时间: %v\n", newStats.GCPauseAvg)
	fmt.Printf("运行时间: %v\n", time.Since(startTime))

	// 5. GC调优演示
	fmt.Println("\n4. GC调优演示")
	tuner := NewGCTuner()
	tuner.TuneForLatency()

	// 分配一些内存观察调优效果
	allocDemo := NewMemoryAllocatorDemo()
	allocDemo.AllocateSmallObjects(500)

	tuner.TuneForThroughput()
	allocDemo.AllocateLargeObjects(10)

	tuner.Restore()

	// 6. 强制GC
	fmt.Println("\n5. 强制GC演示")
	allocDemo.ForceGC()

	// 7. 堆栈扫描演示
	StackScanDemo()

	// 8. 终结器演示
	fmt.Println("\n6. 终结器机制演示")
	finalObj1 := NewFinalizableObject(1, "database_connection")
	finalObj2 := NewFinalizableObject(2, "file_handle")
	finalObj3 := NewFinalizableObject(3, "network_socket")

	// 显式关闭一个对象
	finalObj2.Close()

	// 让对象变为不可达
	fmt.Printf("准备释放对象 %d 和 %d\n", finalObj1.ID, finalObj3.ID)
	finalObj1 = nil
	finalObj3 = nil

	// 强制GC触发终结器
	runtime.GC()
	runtime.GC()
	time.Sleep(100 * time.Millisecond) // 等待终结器执行

	// 9. 最终统计
	fmt.Println("\n=== 最终GC统计 ===")
	finalStats := monitor.GetLatestStats()
	fmt.Printf("总GC次数: %d\n", finalStats.NumGC)
	fmt.Printf("总GC暂停时间: %v\n", finalStats.GCPauseTotal)
	fmt.Printf("最大GC暂停: %v\n", finalStats.GCPauseMax)
	fmt.Printf("平均GC暂停: %v\n", finalStats.GCPauseAvg)
	fmt.Printf("最终堆大小: %d KB\n", finalStats.HeapSize/1024)
	fmt.Printf("最终堆使用: %d KB\n", finalStats.HeapUsed/1024)
}

func main() {
	// 设置GC调试环境变量
	debug.SetGCPercent(100) // 确保正常的GC行为

	demonstrateGCInternals()

	fmt.Println("\n=== Go GC内核深度解析完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 三色标记算法是Go GC的核心，确保并发安全")
	fmt.Println("2. 写屏障保证三色不变性，防止对象丢失")
	fmt.Println("3. GC调优需要平衡延迟和吞吐量")
	fmt.Println("4. 小对象和大对象有不同的分配策略")
	fmt.Println("5. 终结器提供资源清理，但不应依赖其时机")
	fmt.Println("6. 栈扫描是GC根集合识别的重要组成")
	fmt.Println("7. 并发GC允许程序在GC期间继续执行")

	fmt.Println("\n高级特性:")
	fmt.Println("- GOGC环境变量控制GC触发频率")
	fmt.Println("- SetMemoryLimit()可设置软内存限制")
	fmt.Println("- runtime.GC()可强制触发GC")
	fmt.Println("- SetFinalizer()设置对象终结器")
	fmt.Println("- 三色不变性：黑色对象不直接指向白色对象")
}

/*
=== 练习题 ===

1. 基础练习：
   - 实现自己的GC统计收集器
   - 编写三色标记算法的详细实现
   - 分析不同GOGC值对程序性能的影响
   - 实现对象引用图的可视化

2. 中级练习：
   - 实现写屏障的模拟机制
   - 编写内存分配器的性能基准测试
   - 分析栈扫描对GC性能的影响
   - 实现自定义的对象池

3. 高级练习：
   - 分析Go GC的源码实现
   - 实现增量式GC算法
   - 优化大对象分配的GC影响
   - 实现分代GC的概念验证

4. 性能优化：
   - 使用pprof分析GC性能瓶颈
   - 优化高并发下的GC表现
   - 实现zero-GC的数据结构
   - 分析Go 1.19+的GC改进

5. 深度研究：
   - 研究其他语言的GC算法差异
   - 实现并发标记清除算法
   - 分析NUMA架构对GC的影响
   - 研究实时GC的可能性

运行命令：
go run main.go

环境变量：
export GOGC=50          # 设置GC目标百分比
export GODEBUG=gctrace=1 # 启用GC跟踪
export GOMEMLIMIT=500MiB # 设置内存限制

重要概念：
- 三色标记：White(未标记)、Gray(标记中)、Black(已标记)
- 写屏障：保证并发标记的正确性
- 栈扫描：识别栈上的根对象
- 终结器：对象被GC前的清理机制
- 分配器：小对象(<32KB)和大对象(>=32KB)分别处理
*/
