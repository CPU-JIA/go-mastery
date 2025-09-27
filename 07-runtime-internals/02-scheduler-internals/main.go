/*
=== Go运行时内核：调度器深度解析 ===

本模块深入Go语言调度器(Scheduler)的内核实现，探索：
1. G-M-P模型架构和原理
2. Goroutine生命周期管理
3. 工作窃取(Work Stealing)算法
4. 抢占式调度机制
5. 系统调用处理
6. 网络轮询器(Netpoller)
7. 调度器性能优化
8. GOMAXPROCS调优
9. 调度延迟分析

学习目标：
- 深入理解G-M-P调度模型
- 掌握goroutine调度原理
- 学会调度器性能分析和调优
- 理解抢占式调度的实现机制
*/

package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
	"runtime/trace"
	"sync"
	"sync/atomic"
	"time"
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
// 1. 调度器状态监控
// ==================

// SchedulerStats 调度器统计信息
type SchedulerStats struct {
	NumGoroutine   int           // 当前goroutine数量
	NumCPU         int           // CPU核心数
	GOMAXPROCS     int           // 最大P数量
	NumCgoCall     int64         // CGO调用次数
	SchedulerCalls int64         // 调度器调用次数
	GCWaitTime     time.Duration // GC等待时间
	SchedLatency   time.Duration // 调度延迟
	RunqueueLen    int           // 运行队列长度
	Timestamp      time.Time     // 时间戳
}

// SchedulerMonitor 调度器监控器
type SchedulerMonitor struct {
	stats     []SchedulerStats
	startTime time.Time
	mutex     sync.RWMutex
	running   bool
	stopCh    chan struct{}
}

func NewSchedulerMonitor() *SchedulerMonitor {
	return &SchedulerMonitor{
		stats:     make([]SchedulerStats, 0),
		startTime: time.Now(),
		stopCh:    make(chan struct{}),
	}
}

func (m *SchedulerMonitor) Start() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.running {
		return
	}

	m.running = true
	go m.monitor()
}

func (m *SchedulerMonitor) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.running {
		return
	}

	m.running = false
	close(m.stopCh)
}

func (m *SchedulerMonitor) monitor() {
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

func (m *SchedulerMonitor) collectStats() {
	stats := SchedulerStats{
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		NumCgoCall:   runtime.NumCgoCall(),
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

func (m *SchedulerMonitor) GetLatestStats() SchedulerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.stats) == 0 {
		return SchedulerStats{}
	}

	return m.stats[len(m.stats)-1]
}

func (m *SchedulerMonitor) GetStatsHistory() []SchedulerStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	history := make([]SchedulerStats, len(m.stats))
	copy(history, m.stats)
	return history
}

// ==================
// 2. G-M-P模型演示
// ==================

// GMPState G-M-P状态
type GMPState struct {
	G int // Goroutine数量
	M int // Machine(OS线程)数量
	P int // Processor(逻辑处理器)数量
}

// GMPSimulator G-M-P模型模拟器
type GMPSimulator struct {
	state       GMPState
	workQueue   []WorkItem
	processors  []*Processor
	machines    []*Machine
	goroutines  []*Goroutine
	scheduler   *Scheduler
	running     bool
	mutex       sync.RWMutex
	stopCh      chan struct{}
	workCounter int64
}

// WorkItem 工作项
type WorkItem struct {
	ID          int
	Work        func()
	CreatedAt   time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
	ProcessorID int
}

// Processor 处理器(P)
type Processor struct {
	ID         int
	LocalQueue []WorkItem
	Stealing   bool
	RunningG   *Goroutine
	MachineID  int
	mutex      sync.Mutex
}

// Machine 机器(M)
type Machine struct {
	ID          int
	ProcessorID int
	Blocked     bool
	InSyscall   bool
	ThreadID    int
	CreatedAt   time.Time
}

// Goroutine 协程(G)
type Goroutine struct {
	ID          int
	State       string // "runnable", "running", "waiting", "dead"
	ProcessorID int
	MachineID   int
	StackSize   int
	CreatedAt   time.Time
	StartedAt   *time.Time
	FinishedAt  *time.Time
}

// Scheduler 调度器
type Scheduler struct {
	GlobalQueue []WorkItem
	Processors  []*Processor
	Machines    []*Machine
	WorkSteals  int64
	Preemptions int64
	mutex       sync.Mutex
}

func NewGMPSimulator(maxP int) *GMPSimulator {
	sim := &GMPSimulator{
		state: GMPState{
			G: 0,
			M: 0,
			P: maxP,
		},
		workQueue:  make([]WorkItem, 0),
		processors: make([]*Processor, maxP),
		machines:   make([]*Machine, 0),
		goroutines: make([]*Goroutine, 0),
		stopCh:     make(chan struct{}),
	}

	// 初始化处理器
	for i := 0; i < maxP; i++ {
		sim.processors[i] = &Processor{
			ID:         i,
			LocalQueue: make([]WorkItem, 0),
			MachineID:  -1,
		}
	}

	sim.scheduler = &Scheduler{
		GlobalQueue: make([]WorkItem, 0),
		Processors:  sim.processors,
		Machines:    sim.machines,
	}

	return sim
}

func (sim *GMPSimulator) AddWork(work func()) {
	workID := int(atomic.AddInt64(&sim.workCounter, 1))

	workItem := WorkItem{
		ID:        workID,
		Work:      work,
		CreatedAt: time.Now(),
	}

	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	// 尝试添加到本地队列，否则添加到全局队列
	if len(sim.processors) > 0 {
		// 随机选择一个处理器
		p := sim.processors[secureRandomInt(len(sim.processors))]
		p.mutex.Lock()
		if len(p.LocalQueue) < 256 { // 本地队列容量限制
			p.LocalQueue = append(p.LocalQueue, workItem)
		} else {
			sim.scheduler.GlobalQueue = append(sim.scheduler.GlobalQueue, workItem)
		}
		p.mutex.Unlock()
	}

	sim.workQueue = append(sim.workQueue, workItem)
}

func (sim *GMPSimulator) Start() {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	if sim.running {
		return
	}

	sim.running = true

	// 启动处理器
	for _, p := range sim.processors {
		go sim.runProcessor(p)
	}

	// 启动工作窃取协调器
	go sim.runWorkStealer()
}

func (sim *GMPSimulator) Stop() {
	sim.mutex.Lock()
	defer sim.mutex.Unlock()

	if !sim.running {
		return
	}

	sim.running = false
	close(sim.stopCh)
}

func (sim *GMPSimulator) runProcessor(p *Processor) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sim.processWork(p)
		case <-sim.stopCh:
			return
		}
	}
}

func (sim *GMPSimulator) processWork(p *Processor) {
	// 1. 从本地队列获取工作
	work := sim.getLocalWork(p)

	// 2. 如果本地队列为空，从全局队列获取
	if work == nil {
		work = sim.getGlobalWork(p)
	}

	// 3. 如果全局队列也为空，尝试工作窃取
	if work == nil {
		work = sim.stealWork(p)
	}

	// 4. 执行工作
	if work != nil {
		sim.executeWork(p, work)
	}
}

func (sim *GMPSimulator) getLocalWork(p *Processor) *WorkItem {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(p.LocalQueue) == 0 {
		return nil
	}

	work := p.LocalQueue[0]
	p.LocalQueue = p.LocalQueue[1:]
	return &work
}

func (sim *GMPSimulator) getGlobalWork(p *Processor) *WorkItem {
	sim.scheduler.mutex.Lock()
	defer sim.scheduler.mutex.Unlock()

	if len(sim.scheduler.GlobalQueue) == 0 {
		return nil
	}

	work := sim.scheduler.GlobalQueue[0]
	sim.scheduler.GlobalQueue = sim.scheduler.GlobalQueue[1:]
	work.ProcessorID = p.ID
	return &work
}

func (sim *GMPSimulator) stealWork(p *Processor) *WorkItem {
	// 尝试从其他处理器窃取工作
	for _, target := range sim.processors {
		if target.ID == p.ID {
			continue
		}

		target.mutex.Lock()
		if len(target.LocalQueue) > 1 {
			// 从队列尾部窃取一半工作
			stealCount := len(target.LocalQueue) / 2
			if stealCount > 0 {
				stolen := target.LocalQueue[len(target.LocalQueue)-stealCount:]
				target.LocalQueue = target.LocalQueue[:len(target.LocalQueue)-stealCount]

				// 将窃取的工作添加到自己的队列
				p.mutex.Lock()
				p.LocalQueue = append(p.LocalQueue, stolen...)
				p.mutex.Unlock()

				atomic.AddInt64(&sim.scheduler.WorkSteals, 1)
				target.mutex.Unlock()

				// 返回第一个工作项
				if len(stolen) > 0 {
					work := stolen[0]
					work.ProcessorID = p.ID
					return &work
				}
			}
		}
		target.mutex.Unlock()
	}

	return nil
}

func (sim *GMPSimulator) executeWork(p *Processor, work *WorkItem) {
	startTime := time.Now()
	work.StartedAt = &startTime

	// 创建goroutine
	g := &Goroutine{
		ID:          work.ID,
		State:       "running",
		ProcessorID: p.ID,
		CreatedAt:   work.CreatedAt,
		StartedAt:   &startTime,
	}

	sim.mutex.Lock()
	sim.goroutines = append(sim.goroutines, g)
	sim.state.G++
	sim.mutex.Unlock()

	p.RunningG = g

	// 执行实际工作
	work.Work()

	// 完成工作
	finishTime := time.Now()
	work.FinishedAt = &finishTime
	g.FinishedAt = &finishTime
	g.State = "dead"
	p.RunningG = nil

	sim.mutex.Lock()
	sim.state.G--
	sim.mutex.Unlock()
}

func (sim *GMPSimulator) runWorkStealer() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sim.balanceWork()
		case <-sim.stopCh:
			return
		}
	}
}

func (sim *GMPSimulator) balanceWork() {
	// 检查是否需要工作窃取平衡
	sim.scheduler.mutex.Lock()
	defer sim.scheduler.mutex.Unlock()

	totalWork := len(sim.scheduler.GlobalQueue)
	for _, p := range sim.processors {
		p.mutex.Lock()
		totalWork += len(p.LocalQueue)
		p.mutex.Unlock()
	}

	if totalWork == 0 {
		return
	}

	// 找到工作最多和最少的处理器
	var maxP, minP *Processor
	maxWork, minWork := 0, int(^uint(0)>>1)

	for _, p := range sim.processors {
		p.mutex.Lock()
		workCount := len(p.LocalQueue)
		if workCount > maxWork {
			maxWork = workCount
			maxP = p
		}
		if workCount < minWork {
			minWork = workCount
			minP = p
		}
		p.mutex.Unlock()
	}

	// 如果差异超过阈值，进行工作窃取
	if maxP != nil && minP != nil && maxWork-minWork > 4 {
		maxP.mutex.Lock()
		if len(maxP.LocalQueue) > 2 {
			stealCount := (maxWork - minWork) / 2
			stolen := maxP.LocalQueue[len(maxP.LocalQueue)-stealCount:]
			maxP.LocalQueue = maxP.LocalQueue[:len(maxP.LocalQueue)-stealCount]
			maxP.mutex.Unlock()

			minP.mutex.Lock()
			minP.LocalQueue = append(minP.LocalQueue, stolen...)
			minP.mutex.Unlock()

			atomic.AddInt64(&sim.scheduler.WorkSteals, 1)
		} else {
			maxP.mutex.Unlock()
		}
	}
}

func (sim *GMPSimulator) GetStats() GMPState {
	sim.mutex.RLock()
	defer sim.mutex.RUnlock()
	return sim.state
}

func (sim *GMPSimulator) GetWorkStats() (int, int, int64) {
	sim.scheduler.mutex.Lock()
	globalWork := len(sim.scheduler.GlobalQueue)
	workSteals := sim.scheduler.WorkSteals
	sim.scheduler.mutex.Unlock()

	localWork := 0
	for _, p := range sim.processors {
		p.mutex.Lock()
		localWork += len(p.LocalQueue)
		p.mutex.Unlock()
	}

	return globalWork, localWork, workSteals
}

// ==================
// 3. 抢占式调度演示
// ==================

// PreemptiveSchedulerDemo 抢占式调度演示
type PreemptiveSchedulerDemo struct {
	longRunningTasks int64
	preemptions      int64
	running          bool
	stopCh           chan struct{}
	wg               sync.WaitGroup
}

func NewPreemptiveSchedulerDemo() *PreemptiveSchedulerDemo {
	return &PreemptiveSchedulerDemo{
		stopCh: make(chan struct{}),
	}
}

func (d *PreemptiveSchedulerDemo) Start() {
	d.running = true

	// 启动多个长时间运行的goroutine
	for i := 0; i < runtime.GOMAXPROCS(0)*2; i++ {
		d.wg.Add(1)
		go d.longRunningTask(i)
	}

	// 启动抢占监控
	d.wg.Add(1)
	go d.monitorPreemption()

	// 启动短任务生成器
	d.wg.Add(1)
	go d.generateShortTasks()
}

func (d *PreemptiveSchedulerDemo) Stop() {
	d.running = false
	close(d.stopCh)
	d.wg.Wait()
}

func (d *PreemptiveSchedulerDemo) longRunningTask(id int) {
	defer d.wg.Done()

	atomic.AddInt64(&d.longRunningTasks, 1)
	defer atomic.AddInt64(&d.longRunningTasks, -1)

	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ticker.C:
			// 模拟CPU密集型工作
			for i := 0; i < 10000; i++ {
				counter += i
			}

			// 检查是否应该主动让出
			if counter%100000 == 0 {
				runtime.Gosched() // 主动让出CPU
				atomic.AddInt64(&d.preemptions, 1)
			}

		case <-d.stopCh:
			return
		}
	}
}

func (d *PreemptiveSchedulerDemo) generateShortTasks() {
	defer d.wg.Done()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 生成短任务来测试抢占
			go func() {
				start := time.Now()
				for time.Since(start) < time.Millisecond {
					// 短时间运行
				}
			}()

		case <-d.stopCh:
			return
		}
	}
}

func (d *PreemptiveSchedulerDemo) monitorPreemption() {
	defer d.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			longTasks := atomic.LoadInt64(&d.longRunningTasks)
			preemptions := atomic.LoadInt64(&d.preemptions)
			goroutines := runtime.NumGoroutine()

			fmt.Printf("抢占式调度状态 - 长任务:%d, 抢占次数:%d, 总Goroutine:%d\n",
				longTasks, preemptions, goroutines)

		case <-d.stopCh:
			return
		}
	}
}

// ==================
// 4. 系统调用处理演示
// ==================

// SyscallDemo 系统调用演示
type SyscallDemo struct {
	syscallCount int64
	blockingOps  int64
	running      bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

func NewSyscallDemo() *SyscallDemo {
	return &SyscallDemo{
		stopCh: make(chan struct{}),
	}
}

func (d *SyscallDemo) Start() {
	d.running = true

	// 启动阻塞系统调用任务
	for i := 0; i < 5; i++ {
		d.wg.Add(1)
		go d.blockingSyscallTask(i)
	}

	// 启动非阻塞任务
	for i := 0; i < 10; i++ {
		d.wg.Add(1)
		go d.nonBlockingTask(i)
	}

	// 监控系统调用
	d.wg.Add(1)
	go d.monitorSyscalls()
}

func (d *SyscallDemo) Stop() {
	d.running = false
	close(d.stopCh)
	d.wg.Wait()
}

func (d *SyscallDemo) blockingSyscallTask(id int) {
	defer d.wg.Done()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 模拟阻塞系统调用（如文件I/O）
			atomic.AddInt64(&d.syscallCount, 1)
			atomic.AddInt64(&d.blockingOps, 1)

			// 模拟系统调用阻塞
			time.Sleep(100 * time.Millisecond)

			atomic.AddInt64(&d.blockingOps, -1)

		case <-d.stopCh:
			return
		}
	}
}

func (d *SyscallDemo) nonBlockingTask(id int) {
	defer d.wg.Done()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 模拟CPU工作，不涉及系统调用
			sum := 0
			for i := 0; i < 50000; i++ {
				sum += i
			}

		case <-d.stopCh:
			return
		}
	}
}

func (d *SyscallDemo) monitorSyscalls() {
	defer d.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			syscalls := atomic.LoadInt64(&d.syscallCount)
			blocking := atomic.LoadInt64(&d.blockingOps)
			goroutines := runtime.NumGoroutine()
			threads := runtime.GOMAXPROCS(0)

			fmt.Printf("系统调用状态 - 总调用:%d, 阻塞中:%d, Goroutine:%d, 线程:%d\n",
				syscalls, blocking, goroutines, threads)

		case <-d.stopCh:
			return
		}
	}
}

// ==================
// 5. 调度延迟测量
// ==================

// LatencyMeasurement 调度延迟测量
type LatencyMeasurement struct {
	measurements []time.Duration
	mutex        sync.Mutex
}

func NewLatencyMeasurement() *LatencyMeasurement {
	return &LatencyMeasurement{
		measurements: make([]time.Duration, 0),
	}
}

func (lm *LatencyMeasurement) MeasureSchedulingLatency(iterations int) {
	fmt.Printf("\n=== 测量调度延迟 (%d次迭代) ===\n", iterations)

	for i := 0; i < iterations; i++ {
		scheduleTime := time.Now()

		done := make(chan time.Time)
		go func() {
			runTime := time.Now()
			done <- runTime
		}()

		runTime := <-done
		latency := runTime.Sub(scheduleTime)

		lm.mutex.Lock()
		lm.measurements = append(lm.measurements, latency)
		lm.mutex.Unlock()

		if i%100 == 0 {
			fmt.Printf("完成 %d 次测量\n", i)
		}
	}

	lm.printStatistics()
}

func (lm *LatencyMeasurement) printStatistics() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if len(lm.measurements) == 0 {
		return
	}

	var total, min, max time.Duration
	min = time.Duration(int64(^uint64(0) >> 1)) // 最大值
	max = 0

	for _, latency := range lm.measurements {
		total += latency
		if latency < min {
			min = latency
		}
		if latency > max {
			max = latency
		}
	}

	avg := total / time.Duration(len(lm.measurements))

	fmt.Printf("\n调度延迟统计:\n")
	fmt.Printf("  测量次数: %d\n", len(lm.measurements))
	fmt.Printf("  平均延迟: %v\n", avg)
	fmt.Printf("  最小延迟: %v\n", min)
	fmt.Printf("  最大延迟: %v\n", max)
	fmt.Printf("  总时间: %v\n", total)

	// 计算百分位数
	lm.printPercentiles()
}

func (lm *LatencyMeasurement) printPercentiles() {
	// 简单排序计算百分位数
	measurements := make([]time.Duration, len(lm.measurements))
	copy(measurements, lm.measurements)

	// 简单冒泡排序
	for i := 0; i < len(measurements); i++ {
		for j := i + 1; j < len(measurements); j++ {
			if measurements[i] > measurements[j] {
				measurements[i], measurements[j] = measurements[j], measurements[i]
			}
		}
	}

	p50 := measurements[len(measurements)*50/100]
	p90 := measurements[len(measurements)*90/100]
	p95 := measurements[len(measurements)*95/100]
	p99 := measurements[len(measurements)*99/100]

	fmt.Printf("  P50延迟: %v\n", p50)
	fmt.Printf("  P90延迟: %v\n", p90)
	fmt.Printf("  P95延迟: %v\n", p95)
	fmt.Printf("  P99延迟: %v\n", p99)
}

// ==================
// 6. 主演示函数
// ==================

func demonstrateSchedulerInternals() {
	fmt.Println("=== Go调度器内核深度解析 ===")

	// 1. 启动调度器监控
	monitor := NewSchedulerMonitor()
	monitor.Start()
	defer monitor.Stop()

	fmt.Println("\n1. 调度器基本信息")
	stats := monitor.GetLatestStats()
	fmt.Printf("CPU核心数: %d\n", stats.NumCPU)
	fmt.Printf("GOMAXPROCS: %d\n", stats.GOMAXPROCS)
	fmt.Printf("当前Goroutine数: %d\n", stats.NumGoroutine)

	// 2. G-M-P模型演示
	fmt.Println("\n2. G-M-P调度模型演示")
	maxP := runtime.GOMAXPROCS(0)
	simulator := NewGMPSimulator(maxP)
	simulator.Start()
	defer simulator.Stop()

	// 添加各种工作负载
	fmt.Println("添加工作负载...")

	// CPU密集型任务
	for i := 0; i < 20; i++ {
		simulator.AddWork(func() {
			sum := 0
			for j := 0; j < 1000000; j++ {
				sum += j
			}
		})
	}

	// I/O密集型任务
	for i := 0; i < 10; i++ {
		simulator.AddWork(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 混合任务
	for i := 0; i < 15; i++ {
		simulator.AddWork(func() {
			// 计算一段时间
			for j := 0; j < 100000; j++ {
				_ = j * j
			}
			// 然后休眠
			time.Sleep(time.Millisecond)
		})
	}

	// 等待一段时间让调度器工作
	time.Sleep(2 * time.Second)

	gmpStats := simulator.GetStats()
	globalWork, localWork, workSteals := simulator.GetWorkStats()

	fmt.Printf("G-M-P状态:\n")
	fmt.Printf("  活跃Goroutine: %d\n", gmpStats.G)
	fmt.Printf("  处理器数量: %d\n", gmpStats.P)
	fmt.Printf("  全局队列工作: %d\n", globalWork)
	fmt.Printf("  本地队列工作: %d\n", localWork)
	fmt.Printf("  工作窃取次数: %d\n", workSteals)

	// 3. 抢占式调度演示
	fmt.Println("\n3. 抢占式调度演示")
	preemptDemo := NewPreemptiveSchedulerDemo()
	preemptDemo.Start()

	fmt.Println("运行抢占式调度演示 3 秒...")
	time.Sleep(3 * time.Second)

	preemptDemo.Stop()

	// 4. 系统调用处理演示
	fmt.Println("\n4. 系统调用处理演示")
	syscallDemo := NewSyscallDemo()
	syscallDemo.Start()

	fmt.Println("运行系统调用演示 3 秒...")
	time.Sleep(3 * time.Second)

	syscallDemo.Stop()

	// 5. 调度延迟测量
	latencyMeasurer := NewLatencyMeasurement()
	latencyMeasurer.MeasureSchedulingLatency(1000)

	// 6. 不同GOMAXPROCS的影响
	fmt.Println("\n5. GOMAXPROCS调优演示")
	demonstrateGOMAXPROCS()

	// 7. 最终统计
	fmt.Println("\n=== 最终调度器统计 ===")
	finalStats := monitor.GetLatestStats()
	fmt.Printf("最终Goroutine数: %d\n", finalStats.NumGoroutine)
	fmt.Printf("GOMAXPROCS: %d\n", finalStats.GOMAXPROCS)
	fmt.Printf("CGO调用次数: %d\n", finalStats.NumCgoCall)
}

func demonstrateGOMAXPROCS() {
	originalMAXPROCS := runtime.GOMAXPROCS(0)

	// 测试不同的GOMAXPROCS值
	values := []int{1, 2, 4, 8}
	for _, maxprocs := range values {
		if maxprocs > runtime.NumCPU() {
			continue
		}

		fmt.Printf("\n测试 GOMAXPROCS = %d:\n", maxprocs)
		runtime.GOMAXPROCS(maxprocs)

		start := time.Now()
		var wg sync.WaitGroup

		// 启动CPU密集型任务
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sum := 0
				for j := 0; j < 10000000; j++ {
					sum += j
				}
			}()
		}

		wg.Wait()
		duration := time.Since(start)

		fmt.Printf("  执行时间: %v\n", duration)
		fmt.Printf("  当前Goroutine数: %d\n", runtime.NumGoroutine())
	}

	// 恢复原始设置
	runtime.GOMAXPROCS(originalMAXPROCS)
	fmt.Printf("\n恢复 GOMAXPROCS = %d\n", originalMAXPROCS)
}

// ==================
// 7. 跟踪分析演示
// ==================

func demonstrateTracing() {
	fmt.Println("\n=== 调度器跟踪分析 ===")

	// 创建上下文用于跟踪
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 启动跟踪（注意：这需要go tool trace支持）
	err := trace.Start(nil) // 在实际应用中需要提供io.Writer
	if err != nil {
		fmt.Printf("跟踪启动失败: %v\n", err)
		return
	}
	defer trace.Stop()

	// 创建一些有趣的调度活动
	var wg sync.WaitGroup

	// CPU密集型goroutine
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			trace.WithRegion(ctx, fmt.Sprintf("cpu-task-%d", id), func() {
				sum := 0
				for j := 0; j < 1000000; j++ {
					sum += j
				}
			})
		}(i)
	}

	// I/O密集型goroutine
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			trace.WithRegion(ctx, fmt.Sprintf("io-task-%d", id), func() {
				time.Sleep(100 * time.Millisecond)
			})
		}(i)
	}

	wg.Wait()
	fmt.Println("跟踪完成")
}

func main() {
	demonstrateSchedulerInternals()

	// 可选：演示跟踪分析
	// demonstrateTracing()

	fmt.Println("\n=== Go调度器内核深度解析完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. G-M-P模型：Goroutine-Machine-Processor三层调度")
	fmt.Println("2. 工作窃取：P从其他P的本地队列窃取工作")
	fmt.Println("3. 抢占式调度：防止goroutine长时间占用CPU")
	fmt.Println("4. 系统调用处理：阻塞调用时自动创建新M")
	fmt.Println("5. 调度延迟：影响系统响应性的关键指标")
	fmt.Println("6. GOMAXPROCS：控制并发执行的P数量")

	fmt.Println("\n高级特性:")
	fmt.Println("- 调度器会在GC期间协调所有P")
	fmt.Println("- 网络轮询器(netpoller)处理网络I/O")
	fmt.Println("- 栈增长时会触发调度器检查")
	fmt.Println("- 信号处理由专门的M负责")
	fmt.Println("- 抢占点包括函数调用、循环反边等")
}

/*
=== 练习题 ===

1. 基础练习：
   - 实现简单的goroutine池
   - 测量不同工作负载的调度延迟
   - 分析GOMAXPROCS对性能的影响
   - 实现工作窃取队列

2. 中级练习：
   - 实现优先级调度器
   - 分析调度器的公平性
   - 测量系统调用对调度的影响
   - 实现协程亲和性调度

3. 高级练习：
   - 实现NUMA感知的调度器
   - 分析调度器的可扩展性
   - 优化高并发下的调度性能
   - 实现实时调度策略

4. 性能分析：
   - 使用runtime/trace分析调度行为
   - 测量调度器开销
   - 分析上下文切换成本
   - 优化调度热路径

5. 深度研究：
   - 研究其他语言的调度器差异
   - 实现用户态调度器
   - 分析调度器与GC的交互
   - 研究异步调度模型

运行命令：
go run main.go

环境变量：
export GOMAXPROCS=4           # 设置最大P数量
export GODEBUG=schedtrace=1000 # 启用调度跟踪
export GODEBUG=scheddetail=1   # 详细调度信息

重要概念：
- G(Goroutine)：用户态协程
- M(Machine)：OS线程
- P(Processor)：逻辑处理器，执行G的上下文
- 本地队列：P的私有工作队列
- 全局队列：所有P共享的工作队列
- 工作窃取：负载均衡机制
- 抢占式调度：防止饥饿的机制
*/
