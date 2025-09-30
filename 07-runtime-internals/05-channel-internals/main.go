/*
=== Go运行时内核：Channel底层实现 ===

本模块深入Go语言Channel的内核实现，探索：
1. Hchan结构体详细分析
2. Channel的发送和接收机制
3. Select语句的实现原理
4. 缓冲Channel vs 无缓冲Channel
5. Channel的阻塞和唤醒机制
6. Goroutine队列管理
7. Channel的内存模型
8. Channel性能优化
9. Channel的垃圾回收处理

学习目标：
- 深入理解Channel的底层数据结构
- 掌握Channel通信的同步机制
- 学会Channel性能分析和优化
- 理解Select语句的调度原理
*/

package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
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
// 1. Channel结构模拟
// ==================

// MockHchan 模拟runtime.hchan结构
type MockHchan struct {
	// Channel基本信息
	qcount   uint           // 队列中的元素数量
	dataqsiz uint           // 缓冲区大小
	buf      unsafe.Pointer // #nosec G103 - 教学演示：模拟Go runtime的channel缓冲区指针，演示环形缓冲区实现
	elemsize uint16         // 元素大小
	closed   uint32         // 是否关闭
	elemtype unsafe.Pointer // #nosec G103 - 教学演示：模拟Go runtime的类型信息指针，用于反射和类型检查

	// 发送和接收索引
	sendx uint // 发送索引
	recvx uint // 接收索引

	// 等待队列
	recvq WaitQueue // 接收等待队列
	sendq WaitQueue // 发送等待队列

	// 互斥锁
	lock sync.Mutex
}

// MockSudog 模拟runtime.sudog结构
type MockSudog struct {
	// #nosec G103 - 教学演示：模拟Go runtime的sudog（blocking goroutine）结构
	// 这些unsafe.Pointer字段用于存储goroutine和数据的底层指针
	// 在真实的Go runtime中，sudog用于管理阻塞在channel/select/sync原语上的goroutine
	g           unsafe.Pointer // goroutine指针
	elem        unsafe.Pointer // 数据元素指针
	acquiretime int64          // 获取时间
	releasetime int64          // 释放时间
	ticket      uint32         // 票据
	isSelect    bool           // 是否来自select
	success     bool           // 是否成功
	parent      *MockSudog     // 父节点
	waitlink    *MockSudog     // 等待链表
	waittail    *MockSudog     // 等待尾部
	c           *MockHchan     // channel指针

	// 用于演示的字段
	goroutineID int
	data        interface{}
	done        chan struct{}
}

// WaitQueue 等待队列
type WaitQueue struct {
	first *MockSudog
	last  *MockSudog
	count int
}

func (wq *WaitQueue) enqueue(sg *MockSudog) {
	if wq.last == nil {
		wq.first = sg
		wq.last = sg
	} else {
		wq.last.waitlink = sg
		wq.last = sg
	}
	wq.count++
}

func (wq *WaitQueue) dequeue() *MockSudog {
	if wq.first == nil {
		return nil
	}

	sg := wq.first
	wq.first = sg.waitlink
	if wq.first == nil {
		wq.last = nil
	}
	wq.count--
	sg.waitlink = nil
	return sg
}

// ChannelSimulator Channel模拟器
type ChannelSimulator struct {
	channels    map[string]*MockHchan
	goroutines  map[int]*MockGoroutine
	selectCases []*SelectCase
	stats       ChannelStats
	mutex       sync.RWMutex
	nextID      int64
	buffers     map[string][]interface{} // 保持缓冲区引用避免GC
}

// MockGoroutine 模拟Goroutine
type MockGoroutine struct {
	ID       int
	State    string // "running", "waiting", "ready"
	WaitChan *MockHchan
	WaitType string // "send", "recv", "select"
	Data     interface{}
}

// SelectCase Select case结构
type SelectCase struct {
	Dir  string // "send", "recv", "default"
	Chan *MockHchan
	Data interface{}
}

// ChannelStats Channel统计信息
type ChannelStats struct {
	TotalChannels    int64
	BufferedChannels int64
	ClosedChannels   int64
	SendOperations   int64
	RecvOperations   int64
	BlockedSenders   int64
	BlockedReceivers int64
	SelectOperations int64
}

func NewChannelSimulator() *ChannelSimulator {
	return &ChannelSimulator{
		channels:    make(map[string]*MockHchan),
		goroutines:  make(map[int]*MockGoroutine),
		selectCases: make([]*SelectCase, 0),
		buffers:     make(map[string][]interface{}),
	}
}

// ==================
// 2. Channel操作模拟
// ==================

func (cs *ChannelSimulator) MakeChannel(name string, size uint) *MockHchan {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	ch := &MockHchan{
		dataqsiz: size,
		elemsize: 8, // 假设元素大小为8字节
		sendx:    0,
		recvx:    0,
	}

	if size > 0 {
		// 分配缓冲区
		buf := make([]interface{}, size)
		cs.buffers[name] = buf // 保持引用避免GC
		// #nosec G103 - 教学演示：模拟Go runtime的channel缓冲区分配
		// 在真实的Go runtime中，channel缓冲区也使用unsafe.Pointer存储
		// 这里演示了如何将Go slice转换为unsafe.Pointer来模拟底层实现
		ch.buf = unsafe.Pointer(&buf[0])
		atomic.AddInt64(&cs.stats.BufferedChannels, 1)
	}

	cs.channels[name] = ch
	atomic.AddInt64(&cs.stats.TotalChannels, 1)

	fmt.Printf("创建Channel '%s' (缓冲区大小: %d)\n", name, size)
	return ch
}

func (cs *ChannelSimulator) Send(ch *MockHchan, data interface{}, goroutineID int) bool {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	atomic.AddInt64(&cs.stats.SendOperations, 1)

	fmt.Printf("Goroutine %d 尝试发送数据: %v\n", goroutineID, data)

	// 检查channel是否关闭
	if ch.closed != 0 {
		fmt.Printf("Goroutine %d 发送失败: channel已关闭\n", goroutineID)
		return false
	}

	// 检查是否有等待的接收者
	if sg := ch.recvq.dequeue(); sg != nil {
		fmt.Printf("直接传递给等待的接收者 Goroutine %d\n", sg.goroutineID)
		sg.data = data
		sg.success = true
		close(sg.done) // 唤醒等待的goroutine
		return true
	}

	// 检查缓冲区是否有空间
	if ch.qcount < ch.dataqsiz {
		cs.enqueueData(ch, data)
		fmt.Printf("数据入队到缓冲区 (位置: %d)\n", ch.sendx)
		return true
	}

	// 需要阻塞
	fmt.Printf("Goroutine %d 因发送阻塞\n", goroutineID)
	atomic.AddInt64(&cs.stats.BlockedSenders, 1)

	// 创建sudog并加入发送队列
	sg := &MockSudog{
		goroutineID: goroutineID,
		data:        data,
		done:        make(chan struct{}),
		c:           ch,
	}

	ch.sendq.enqueue(sg)

	// 释放锁并等待
	ch.lock.Unlock()
	<-sg.done // 阻塞等待
	ch.lock.Lock()

	return sg.success
}

func (cs *ChannelSimulator) Recv(ch *MockHchan, goroutineID int) (interface{}, bool) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	atomic.AddInt64(&cs.stats.RecvOperations, 1)

	fmt.Printf("Goroutine %d 尝试接收数据\n", goroutineID)

	// 检查是否有等待的发送者
	if sg := ch.sendq.dequeue(); sg != nil {
		var data interface{}
		if ch.dataqsiz == 0 {
			// 无缓冲channel，直接传递
			data = sg.data
			fmt.Printf("从等待的发送者 Goroutine %d 直接接收: %v\n", sg.goroutineID, data)
		} else {
			// 有缓冲channel，从缓冲区接收，发送者数据入队
			data = cs.dequeueData(ch)
			cs.enqueueData(ch, sg.data)
			fmt.Printf("从缓冲区接收: %v，发送者数据入队\n", data)
		}

		sg.success = true
		close(sg.done) // 唤醒发送者
		return data, true
	}

	// 检查缓冲区是否有数据
	if ch.qcount > 0 {
		data := cs.dequeueData(ch)
		fmt.Printf("从缓冲区接收数据: %v (位置: %d)\n", data, ch.recvx)
		return data, true
	}

	// 检查channel是否关闭
	if ch.closed != 0 {
		fmt.Printf("Goroutine %d 从已关闭channel接收到零值\n", goroutineID)
		return nil, false
	}

	// 需要阻塞
	fmt.Printf("Goroutine %d 因接收阻塞\n", goroutineID)
	atomic.AddInt64(&cs.stats.BlockedReceivers, 1)

	// 创建sudog并加入接收队列
	sg := &MockSudog{
		goroutineID: goroutineID,
		done:        make(chan struct{}),
		c:           ch,
	}

	ch.recvq.enqueue(sg)

	// 释放锁并等待
	ch.lock.Unlock()
	<-sg.done // 阻塞等待
	ch.lock.Lock()

	return sg.data, sg.success
}

func (cs *ChannelSimulator) Close(ch *MockHchan) {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	if ch.closed != 0 {
		panic("close of closed channel")
	}

	ch.closed = 1
	atomic.AddInt64(&cs.stats.ClosedChannels, 1)

	fmt.Println("Channel已关闭，唤醒所有等待的接收者")

	// 唤醒所有等待的接收者
	for sg := ch.recvq.dequeue(); sg != nil; sg = ch.recvq.dequeue() {
		sg.data = nil
		sg.success = false
		close(sg.done)
	}

	// 发送者会在尝试发送时得到panic
}

func (cs *ChannelSimulator) enqueueData(ch *MockHchan, data interface{}) {
	// #nosec G103 - 教学演示：模拟Go runtime的channel缓冲区操作
	// 演示如何将unsafe.Pointer转换回原始类型以访问环形缓冲区
	// 在真实的Go runtime中，channel使用类似的unsafe操作来管理缓冲区
	// 模拟数据入队
	buf := (*[]interface{})(ch.buf)
	(*buf)[ch.sendx] = data
	ch.sendx++
	if ch.sendx == ch.dataqsiz {
		ch.sendx = 0
	}
	ch.qcount++
}

func (cs *ChannelSimulator) dequeueData(ch *MockHchan) interface{} {
	// #nosec G103 - 教学演示：模拟Go runtime的channel缓冲区操作
	// 演示如何从环形缓冲区中取出数据
	// 模拟数据出队
	buf := (*[]interface{})(ch.buf)
	data := (*buf)[ch.recvx]
	(*buf)[ch.recvx] = nil // 清空引用
	ch.recvx++
	if ch.recvx == ch.dataqsiz {
		ch.recvx = 0
	}
	ch.qcount--
	return data
}

// ==================
// 3. Select语句模拟
// ==================

func (cs *ChannelSimulator) Select(cases []SelectCase, goroutineID int) (int, interface{}, bool) {
	atomic.AddInt64(&cs.stats.SelectOperations, 1)

	fmt.Printf("Goroutine %d 执行select，检查 %d 个case\n", goroutineID, len(cases))

	// 第一阶段：尝试非阻塞操作
	for i, c := range cases {
		switch c.Dir {
		case "recv":
			if cs.tryRecv(c.Chan) {
				data, ok := cs.Recv(c.Chan, goroutineID)
				fmt.Printf("Select case %d (recv) 立即可用\n", i)
				return i, data, ok
			}
		case "send":
			if cs.trySend(c.Chan) {
				success := cs.Send(c.Chan, c.Data, goroutineID)
				fmt.Printf("Select case %d (send) 立即可用\n", i)
				return i, nil, success
			}
		case "default":
			fmt.Printf("Select执行default case\n")
			return i, nil, true
		}
	}

	// 第二阶段：没有可立即执行的case，准备阻塞
	fmt.Printf("所有case都不可用，准备阻塞\n")

	// 创建select context
	selectDone := make(chan int)
	var selectedCase int

	// 为每个case创建sudog并加入等待队列
	var sudogs []*MockSudog
	for i, c := range cases {
		if c.Dir == "default" {
			continue
		}

		sg := &MockSudog{
			goroutineID: goroutineID,
			data:        c.Data,
			done:        make(chan struct{}),
			c:           c.Chan,
			isSelect:    true,
		}

		sudogs = append(sudogs, sg)

		// 根据方向加入相应队列
		c.Chan.lock.Lock()
		if c.Dir == "send" {
			c.Chan.sendq.enqueue(sg)
		} else {
			c.Chan.recvq.enqueue(sg)
		}
		c.Chan.lock.Unlock()

		// 启动监听goroutine
		go func(caseIndex int, sudog *MockSudog) {
			<-sudog.done
			select {
			case selectDone <- caseIndex:
				// 第一个完成的case
			default:
				// 其他case，清理
			}
		}(i, sg)
	}

	// 等待任意case完成
	selectedCase = <-selectDone

	// 清理其他case的sudogs
	for i, _ := range sudogs {
		c := cases[i]
		if i != selectedCase {
			c.Chan.lock.Lock()
			// 从队列中移除（简化实现，实际Go运行时更复杂）
			c.Chan.lock.Unlock()
		}
	}

	fmt.Printf("Select选择了case %d\n", selectedCase)

	// 执行选中的case
	selectedSudog := sudogs[selectedCase]
	return selectedCase, selectedSudog.data, selectedSudog.success
}

func (cs *ChannelSimulator) tryRecv(ch *MockHchan) bool {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// 有等待的发送者或缓冲区有数据或channel已关闭
	return ch.sendq.count > 0 || ch.qcount > 0 || ch.closed != 0
}

func (cs *ChannelSimulator) trySend(ch *MockHchan) bool {
	ch.lock.Lock()
	defer ch.lock.Unlock()

	// channel未关闭且(有等待的接收者或缓冲区有空间)
	return ch.closed == 0 && (ch.recvq.count > 0 || ch.qcount < ch.dataqsiz)
}

// ==================
// 4. Channel性能测试
// ==================

// ChannelBenchmark Channel性能测试
type ChannelBenchmark struct {
	bufferedCh   chan int
	unbufferedCh chan int
	sendCount    int64
	recvCount    int64
	startTime    time.Time
}

func NewChannelBenchmark(bufferSize int) *ChannelBenchmark {
	return &ChannelBenchmark{
		bufferedCh:   make(chan int, bufferSize),
		unbufferedCh: make(chan int),
		startTime:    time.Now(),
	}
}

func (cb *ChannelBenchmark) BenchmarkBufferedChannel(goroutines, messages int) {
	fmt.Printf("\n=== 缓冲Channel性能测试 ===\n")
	fmt.Printf("Goroutines: %d, Messages: %d\n", goroutines, messages)

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动发送者
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(senderID int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				cb.bufferedCh <- senderID*1000 + j
				atomic.AddInt64(&cb.sendCount, 1)
			}
		}(i)
	}

	// 启动接收者
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(receiverID int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				<-cb.bufferedCh
				atomic.AddInt64(&cb.recvCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalOps := atomic.LoadInt64(&cb.sendCount) + atomic.LoadInt64(&cb.recvCount)
	throughput := float64(totalOps) / duration.Seconds()

	fmt.Printf("完成时间: %v\n", duration)
	fmt.Printf("总操作数: %d\n", totalOps)
	fmt.Printf("吞吐量: %.2f ops/sec\n", throughput)
}

func (cb *ChannelBenchmark) BenchmarkUnbufferedChannel(goroutines, messages int) {
	fmt.Printf("\n=== 无缓冲Channel性能测试 ===\n")
	fmt.Printf("Goroutines: %d, Messages: %d\n", goroutines, messages)

	// 重置计数器
	atomic.StoreInt64(&cb.sendCount, 0)
	atomic.StoreInt64(&cb.recvCount, 0)

	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动发送者
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(senderID int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				cb.unbufferedCh <- senderID*1000 + j
				atomic.AddInt64(&cb.sendCount, 1)
			}
		}(i)
	}

	// 启动接收者
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(receiverID int) {
			defer wg.Done()
			for j := 0; j < messages; j++ {
				<-cb.unbufferedCh
				atomic.AddInt64(&cb.recvCount, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	totalOps := atomic.LoadInt64(&cb.sendCount) + atomic.LoadInt64(&cb.recvCount)
	throughput := float64(totalOps) / duration.Seconds()

	fmt.Printf("完成时间: %v\n", duration)
	fmt.Printf("总操作数: %d\n", totalOps)
	fmt.Printf("吞吐量: %.2f ops/sec\n", throughput)
}

func (cb *ChannelBenchmark) BenchmarkSelectPerformance(cases, iterations int) {
	fmt.Printf("\n=== Select性能测试 ===\n")
	fmt.Printf("Cases: %d, Iterations: %d\n", cases, iterations)

	// 创建多个channel
	channels := make([]chan int, cases)
	for i := range channels {
		channels[i] = make(chan int, 1)
	}

	startTime := time.Now()

	var wg sync.WaitGroup

	// 启动数据发送者
	for i, ch := range channels {
		wg.Add(1)
		go func(ch chan int, id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				select {
				case ch <- id*1000 + j:
				default:
				}
				time.Sleep(time.Microsecond) // 模拟工作
			}
		}(ch, i)
	}

	// 启动select接收者
	wg.Add(1)
	go func() {
		defer wg.Done()
		received := 0
		for received < cases*iterations {
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
					received++
				case <-channels[1]:
					received++
				}
			case 4:
				select {
				case <-channels[0]:
					received++
				case <-channels[1]:
					received++
				case <-channels[2]:
					received++
				case <-channels[3]:
					received++
				}
			default:
				// 动态select（实际实现会更复杂）
				for i, ch := range channels {
					select {
					case <-ch:
						received++
					default:
						if i == len(channels)-1 {
							time.Sleep(time.Microsecond)
						}
					}
				}
			}
		}
	}()

	wg.Wait()
	duration := time.Since(startTime)

	selectOps := cases * iterations
	throughput := float64(selectOps) / duration.Seconds()

	fmt.Printf("完成时间: %v\n", duration)
	fmt.Printf("Select操作数: %d\n", selectOps)
	fmt.Printf("吞吐量: %.2f selects/sec\n", throughput)
}

// ==================
// 5. Channel内存模型演示
// ==================

// MemoryModelDemo Channel内存模型演示
type MemoryModelDemo struct {
	ch     chan int
	shared int
	mutex  sync.Mutex
}

func NewMemoryModelDemo() *MemoryModelDemo {
	return &MemoryModelDemo{
		ch: make(chan int),
	}
}

func (mmd *MemoryModelDemo) DemonstrateHappensBefore() {
	fmt.Printf("\n=== Channel内存模型演示 ===\n")

	var wg sync.WaitGroup

	// 演示发送happens-before接收
	fmt.Println("1. 发送happens-before接收")
	wg.Add(2)

	go func() {
		defer wg.Done()
		mmd.shared = 42 // 写操作
		mmd.ch <- 1     // 发送操作
		fmt.Println("发送者：设置shared=42并发送信号")
	}()

	go func() {
		defer wg.Done()
		<-mmd.ch                                       // 接收操作
		fmt.Printf("接收者：收到信号，shared=%d\n", mmd.shared) // 读操作
	}()

	wg.Wait()

	// 演示关闭happens-before接收
	fmt.Println("\n2. 关闭happens-before接收")
	ch2 := make(chan int)

	wg.Add(2)

	go func() {
		defer wg.Done()
		mmd.shared = 100 // 写操作
		close(ch2)       // 关闭操作
		fmt.Println("发送者：设置shared=100并关闭channel")
	}()

	go func() {
		defer wg.Done()
		<-ch2                                            // 接收操作（接收到零值）
		fmt.Printf("接收者：收到关闭信号，shared=%d\n", mmd.shared) // 读操作
	}()

	wg.Wait()

	// 演示无缓冲channel的同步性质
	fmt.Println("\n3. 无缓冲channel的同步性质")
	unbuffered := make(chan int)

	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("发送者：准备发送")
		unbuffered <- 1
		fmt.Println("发送者：发送完成")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // 确保发送者先阻塞
		fmt.Println("接收者：准备接收")
		<-unbuffered
		fmt.Println("接收者：接收完成")
	}()

	wg.Wait()
}

// ==================
// 6. 主演示函数
// ==================

func demonstrateChannelInternals() {
	fmt.Println("=== Go Channel底层实现深度解析 ===")

	// 1. Channel结构和基本操作
	fmt.Println("\n1. Channel结构和基本操作演示")
	simulator := NewChannelSimulator()

	// 创建不同类型的channel
	unbufferedCh := simulator.MakeChannel("unbuffered", 0)
	bufferedCh := simulator.MakeChannel("buffered", 3)

	// 模拟无缓冲channel操作
	fmt.Println("\n--- 无缓冲Channel操作 ---")
	go func() {
		time.Sleep(100 * time.Millisecond)
		simulator.Send(unbufferedCh, "hello", 1)
	}()
	data, ok := simulator.Recv(unbufferedCh, 2)
	fmt.Printf("接收到: %v, ok: %t\n", data, ok)

	// 模拟缓冲channel操作
	fmt.Println("\n--- 缓冲Channel操作 ---")
	simulator.Send(bufferedCh, "msg1", 3)
	simulator.Send(bufferedCh, "msg2", 3)
	simulator.Send(bufferedCh, "msg3", 3)

	data, ok = simulator.Recv(bufferedCh, 4)
	fmt.Printf("接收到: %v, ok: %t\n", data, ok)

	// 2. Select语句演示
	fmt.Println("\n2. Select语句实现演示")
	ch1 := simulator.MakeChannel("select1", 1)
	ch2 := simulator.MakeChannel("select2", 1)

	simulator.Send(ch1, "from ch1", 5)

	cases := []SelectCase{
		{Dir: "recv", Chan: ch1},
		{Dir: "recv", Chan: ch2},
		{Dir: "default"},
	}

	caseIndex, data, ok := simulator.Select(cases, 6)
	fmt.Printf("Select结果: case %d, data: %v, ok: %t\n", caseIndex, data, ok)

	// 3. Channel性能测试
	fmt.Println("\n3. Channel性能测试")
	benchmark := NewChannelBenchmark(100)

	// 测试不同配置
	benchmark.BenchmarkBufferedChannel(4, 1000)
	benchmark.BenchmarkUnbufferedChannel(4, 1000)
	benchmark.BenchmarkSelectPerformance(2, 1000)
	benchmark.BenchmarkSelectPerformance(4, 1000)

	// 4. 内存模型演示
	memoryDemo := NewMemoryModelDemo()
	memoryDemo.DemonstrateHappensBefore()

	// 5. 真实Channel操作测试
	fmt.Println("\n4. 真实Channel操作测试")
	demonstrateRealChannelBehavior()

	// 6. Channel垃圾回收演示
	demonstrateChannelGC()

	// 7. 统计信息
	fmt.Println("\n=== Channel统计信息 ===")
	stats := simulator.stats
	fmt.Printf("总Channel数: %d\n", stats.TotalChannels)
	fmt.Printf("缓冲Channel数: %d\n", stats.BufferedChannels)
	fmt.Printf("已关闭Channel数: %d\n", stats.ClosedChannels)
	fmt.Printf("发送操作数: %d\n", stats.SendOperations)
	fmt.Printf("接收操作数: %d\n", stats.RecvOperations)
	fmt.Printf("阻塞的发送者: %d\n", stats.BlockedSenders)
	fmt.Printf("阻塞的接收者: %d\n", stats.BlockedReceivers)
	fmt.Printf("Select操作数: %d\n", stats.SelectOperations)
}

func demonstrateRealChannelBehavior() {
	fmt.Printf("\n=== 真实Channel行为测试 ===\n")

	// 测试channel的方向性
	fmt.Println("1. Channel方向性测试")
	ch := make(chan int, 2)

	// 只发送channel
	sendOnlyCh := func(ch chan<- int) {
		ch <- 1
		ch <- 2
		close(ch)
	}

	// 只接收channel
	recvOnlyCh := func(ch <-chan int) {
		for v := range ch {
			fmt.Printf("接收: %d\n", v)
		}
	}

	go sendOnlyCh(ch)
	recvOnlyCh(ch)

	// 测试select的随机性
	fmt.Println("\n2. Select随机性测试")
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)

	ch1 <- 1
	ch2 <- 2

	count1, count2 := 0, 0
	for i := 0; i < 100; i++ {
		select {
		case <-ch1:
			count1++
			ch1 <- 1 // 重新填充
		case <-ch2:
			count2++
			ch2 <- 2 // 重新填充
		}
	}

	fmt.Printf("ch1被选择次数: %d, ch2被选择次数: %d\n", count1, count2)

	// 测试nil channel的行为
	fmt.Println("\n3. Nil Channel行为测试")
	var nilCh chan int

	select {
	case nilCh <- 1:
		fmt.Println("这不会执行")
	case <-nilCh:
		fmt.Println("这也不会执行")
	default:
		fmt.Println("Nil channel永远阻塞，执行default")
	}

	// 测试channel的容量和长度
	fmt.Println("\n4. Channel容量和长度测试")
	buffered := make(chan int, 5)
	fmt.Printf("空channel - len: %d, cap: %d\n", len(buffered), cap(buffered))

	buffered <- 1
	buffered <- 2
	fmt.Printf("有2个元素 - len: %d, cap: %d\n", len(buffered), cap(buffered))

	<-buffered
	fmt.Printf("取出1个元素 - len: %d, cap: %d\n", len(buffered), cap(buffered))
}

func demonstrateChannelGC() {
	fmt.Printf("\n=== Channel垃圾回收演示 ===\n")

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	// 创建大量channel
	channels := make([]chan int, 10000)
	for i := range channels {
		channels[i] = make(chan int, secureRandomInt(10)+1)
	}

	runtime.ReadMemStats(&after)
	fmt.Printf("创建10000个channel后堆增长: %d KB\n",
		(after.HeapInuse-before.HeapInuse)/1024)

	// 清除引用
	channels = nil
	runtime.GC()
	runtime.GC() // 确保完成

	var final runtime.MemStats
	runtime.ReadMemStats(&final)
	fmt.Printf("GC后堆使用: %d KB (减少: %d KB)\n",
		final.HeapInuse/1024, (after.HeapInuse-final.HeapInuse)/1024)

	// 演示goroutine泄漏检测
	fmt.Println("\nGoroutine泄漏检测:")
	before_goroutines := runtime.NumGoroutine()

	// 创建一些会泄漏的goroutine
	ch := make(chan int)
	for i := 0; i < 5; i++ {
		go func() {
			<-ch // 永远阻塞
		}()
	}

	time.Sleep(100 * time.Millisecond)
	after_goroutines := runtime.NumGoroutine()
	fmt.Printf("泄漏的goroutine数: %d\n", after_goroutines-before_goroutines)

	// 关闭channel清理goroutine
	close(ch)
	time.Sleep(100 * time.Millisecond)
	final_goroutines := runtime.NumGoroutine()
	fmt.Printf("清理后goroutine数: %d\n", final_goroutines)
}

func main() {
	demonstrateChannelInternals()

	fmt.Println("\n=== Go Channel底层实现深度解析完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. Hchan结构：qcount、dataqsiz、buf、sendx、recvx、sendq、recvq")
	fmt.Println("2. 通信机制：直接传递、缓冲区、阻塞队列三种路径")
	fmt.Println("3. Select实现：轮询、随机选择、阻塞等待的多阶段算法")
	fmt.Println("4. 内存模型：发送happens-before接收的同步保证")
	fmt.Println("5. 性能特征：缓冲vs无缓冲、select开销、goroutine切换成本")
	fmt.Println("6. Sudog结构：管理阻塞goroutine的双向链表")

	fmt.Println("\n高级特性:")
	fmt.Println("- Channel是Go并发原语的核心抽象")
	fmt.Println("- 无缓冲channel提供同步语义")
	fmt.Println("- 缓冲channel提供异步语义")
	fmt.Println("- Select提供多路复用和非阻塞操作")
	fmt.Println("- Channel关闭是广播机制")
	fmt.Println("- Nil channel永远阻塞")
	fmt.Println("- Channel操作是内存屏障")
}

/*
=== 练习题 ===

1. 基础练习：
   - 实现channel的超时机制
   - 分析不同缓冲区大小的性能影响
   - 实现优先级channel
   - 测量channel操作的延迟

2. 中级练习：
   - 实现fan-in和fan-out模式
   - 分析select语句的公平性
   - 实现channel池管理
   - 优化高并发下的channel性能

3. 高级练习：
   - 实现带优先级的select
   - 分析channel的内存分配模式
   - 实现channel的监控和调试工具
   - 优化channel的GC影响

4. 深度分析：
   - 研究Go运行时的channel实现源码
   - 分析channel与调度器的交互
   - 实现无锁的channel变体
   - 分析channel的NUMA影响

5. 实战应用：
   - 实现基于channel的流处理框架
   - 构建channel池化系统
   - 实现channel的可观测性
   - 设计channel的最佳实践指南

运行命令：
go run main.go

环境变量：
export GODEBUG=schedtrace=1000 # 查看调度器行为
export GOMAXPROCS=1           # 测试单核性能

重要概念：
- Hchan: runtime层的channel结构体
- Sudog: 阻塞goroutine的等待节点
- G队列: 发送和接收等待队列
- 内存屏障: channel操作的同步语义
- Happens-before: 内存模型的顺序关系
- 循环缓冲区: 高效的FIFO实现
- Select轮询: 多channel的选择算法
*/
