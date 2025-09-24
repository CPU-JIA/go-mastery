package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

// =============================================================================
// 1. Select 语句高级概念
// =============================================================================

/*
Select 语句是 Go 并发编程中的核心控制结构：

基本功能：
1. 多路复用：同时等待多个通道操作
2. 非阻塞操作：配合 default 分支实现非阻塞通信
3. 超时控制：结合 time.After 实现超时机制
4. 随机选择：当多个 case 同时就绪时随机选择一个

高级模式：
1. 多生产者-单消费者模式
2. 事件监听和分发模式
3. 心跳和健康检查模式
4. 优雅关闭模式
5. 请求合并模式
6. 限流模式
7. 任务调度模式

注意事项：
- 空的 select{} 会永远阻塞
- 没有 case 的 select 等价于空的 select
- select 不会按顺序执行，而是随机选择就绪的 case
- nil 通道上的操作会被忽略
*/

// =============================================================================
// 2. 基础 Select 模式复习
// =============================================================================

func demonstrateBasicSelect() {
	fmt.Println("=== 1. 基础 Select 模式复习 ===")

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)
	done := make(chan bool)

	// 生产者1
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "来自通道1的消息"
	}()

	// 生产者2
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch2 <- "来自通道2的消息"
	}()

	// 超时控制器
	go func() {
		time.Sleep(500 * time.Millisecond)
		done <- true
	}()

	messageCount := 0
	for {
		select {
		case msg1 := <-ch1:
			fmt.Printf("收到通道1消息: %s\n", msg1)
			messageCount++

		case msg2 := <-ch2:
			fmt.Printf("收到通道2消息: %s\n", msg2)
			messageCount++

		case <-done:
			fmt.Printf("超时结束，共收到 %d 条消息\n", messageCount)
			return

		case <-time.After(50 * time.Millisecond):
			fmt.Println("50ms 内没有收到任何消息")

		default:
			fmt.Println("没有消息可接收，执行其他工作...")
			time.Sleep(30 * time.Millisecond)
		}
	}
}

// =============================================================================
// 3. 多生产者-单消费者模式
// =============================================================================

// Producer 生产者接口
type Producer interface {
	Start(output chan<- string)
	Stop()
}

// FastProducer 快速生产者
type FastProducer struct {
	id   int
	quit chan bool
}

func NewFastProducer(id int) *FastProducer {
	return &FastProducer{
		id:   id,
		quit: make(chan bool),
	}
}

func (p *FastProducer) Start(output chan<- string) {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				msg := fmt.Sprintf("快速生产者%d: 消息-%d", p.id, time.Now().Unix())
				output <- msg
			case <-p.quit:
				fmt.Printf("快速生产者%d 停止\n", p.id)
				return
			}
		}
	}()
}

func (p *FastProducer) Stop() {
	close(p.quit)
}

// SlowProducer 慢速生产者
type SlowProducer struct {
	id   int
	quit chan bool
}

func NewSlowProducer(id int) *SlowProducer {
	return &SlowProducer{
		id:   id,
		quit: make(chan bool),
	}
}

func (p *SlowProducer) Start(output chan<- string) {
	go func() {
		ticker := time.NewTicker(800 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				msg := fmt.Sprintf("慢速生产者%d: 重要消息-%d", p.id, time.Now().Unix())
				output <- msg
			case <-p.quit:
				fmt.Printf("慢速生产者%d 停止\n", p.id)
				return
			}
		}
	}()
}

func (p *SlowProducer) Stop() {
	close(p.quit)
}

// Consumer 消费者
type Consumer struct {
	id       int
	messages chan string
	quit     chan bool
	stats    map[string]int
	mu       sync.Mutex
}

func NewConsumer(id int) *Consumer {
	return &Consumer{
		id:       id,
		messages: make(chan string, 10),
		quit:     make(chan bool),
		stats:    make(map[string]int),
	}
}

func (c *Consumer) Start() {
	go func() {
		for {
			select {
			case msg := <-c.messages:
				fmt.Printf("消费者%d 处理: %s\n", c.id, msg)

				// 更新统计信息
				c.mu.Lock()
				if msg[0:2] == "快速" {
					c.stats["fast"]++
				} else {
					c.stats["slow"]++
				}
				c.mu.Unlock()

				// 模拟处理时间
				time.Sleep(50 * time.Millisecond)

			case <-c.quit:
				fmt.Printf("消费者%d 停止\n", c.id)
				return
			}
		}
	}()
}

func (c *Consumer) GetMessageChannel() chan<- string {
	return c.messages
}

func (c *Consumer) Stop() {
	close(c.quit)
}

func (c *Consumer) GetStats() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make(map[string]int)
	for k, v := range c.stats {
		result[k] = v
	}
	return result
}

func demonstrateMultiProducerSingleConsumer() {
	fmt.Println("=== 2. 多生产者-单消费者模式 ===")

	// 创建消费者
	consumer := NewConsumer(1)
	consumer.Start()

	// 创建多个生产者
	producers := []Producer{
		NewFastProducer(1),
		NewFastProducer(2),
		NewSlowProducer(1),
		NewSlowProducer(2),
	}

	// 启动生产者
	output := consumer.GetMessageChannel()
	for _, producer := range producers {
		producer.Start(output)
	}

	fmt.Println("多生产者-单消费者系统运行 3 秒...")
	time.Sleep(3 * time.Second)

	// 停止所有生产者
	for _, producer := range producers {
		producer.Stop()
	}

	// 等待剩余消息处理完
	time.Sleep(500 * time.Millisecond)
	consumer.Stop()

	// 显示统计信息
	stats := consumer.GetStats()
	fmt.Printf("处理统计: 快速消息 %d 条, 慢速消息 %d 条\n", stats["fast"], stats["slow"])

	fmt.Println()
}

// =============================================================================
// 4. 事件监听和分发模式
// =============================================================================

// Event 事件结构体
type Event struct {
	Type      string
	Data      interface{}
	Timestamp time.Time
}

// EventBus 事件总线
type EventBus struct {
	listeners map[string][]chan Event
	mu        sync.RWMutex
	quit      chan bool
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]chan Event),
		quit:      make(chan bool),
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType string, bufferSize int) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan Event, bufferSize)
	eb.listeners[eventType] = append(eb.listeners[eventType], ch)

	fmt.Printf("订阅事件类型: %s\n", eventType)
	return ch
}

// Publish 发布事件
func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	listeners := eb.listeners[event.Type]
	eb.mu.RUnlock()

	fmt.Printf("发布事件: %s - %v\n", event.Type, event.Data)

	for _, listener := range listeners {
		select {
		case listener <- event:
			// 成功发送
		default:
			fmt.Printf("警告: 事件 %s 的监听器缓冲区已满\n", event.Type)
		}
	}
}

// Start 启动事件总线
func (eb *EventBus) Start() {
	// 事件总线本身不需要特殊的启动逻辑
	fmt.Println("事件总线已启动")
}

// Stop 停止事件总线
func (eb *EventBus) Stop() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	close(eb.quit)

	// 关闭所有监听器通道
	for eventType, listeners := range eb.listeners {
		for _, listener := range listeners {
			close(listener)
		}
		fmt.Printf("关闭事件类型 %s 的所有监听器\n", eventType)
	}

	eb.listeners = make(map[string][]chan Event)
}

// EventListener 事件监听器
func EventListener(name string, eventCh <-chan Event, quit <-chan bool) {
	fmt.Printf("事件监听器 %s 开始工作\n", name)

	for {
		select {
		case event, ok := <-eventCh:
			if !ok {
				fmt.Printf("事件监听器 %s: 事件通道已关闭\n", name)
				return
			}
			fmt.Printf("监听器 %s 处理事件: %s - %v (时间: %v)\n",
				name, event.Type, event.Data, event.Timestamp.Format("15:04:05"))

			// 模拟事件处理时间
			time.Sleep(100 * time.Millisecond)

		case <-quit:
			fmt.Printf("事件监听器 %s 收到停止信号\n", name)
			return
		}
	}
}

func demonstrateEventBusPattern() {
	fmt.Println("=== 3. 事件监听和分发模式 ===")

	// 创建事件总线
	eventBus := NewEventBus()
	eventBus.Start()

	// 订阅不同类型的事件
	userEventCh := eventBus.Subscribe("user", 5)
	orderEventCh := eventBus.Subscribe("order", 5)
	systemEventCh := eventBus.Subscribe("system", 5)

	quit := make(chan bool)

	// 启动事件监听器
	go EventListener("用户监听器", userEventCh, quit)
	go EventListener("订单监听器", orderEventCh, quit)
	go EventListener("系统监听器", systemEventCh, quit)

	// 启动事件生产者
	go func() {
		for i := 0; i < 10; i++ {
			// 随机产生不同类型的事件
			eventTypes := []string{"user", "order", "system"}
			eventType := eventTypes[rand.Intn(len(eventTypes))]

			var eventData interface{}
			switch eventType {
			case "user":
				eventData = fmt.Sprintf("用户登录: user-%d", i)
			case "order":
				eventData = fmt.Sprintf("新订单: order-%d", i)
			case "system":
				eventData = fmt.Sprintf("系统状态: status-%d", i)
			}

			event := Event{
				Type:      eventType,
				Data:      eventData,
				Timestamp: time.Now(),
			}

			eventBus.Publish(event)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// 运行 3 秒
	time.Sleep(3 * time.Second)

	// 停止所有组件
	close(quit)
	time.Sleep(200 * time.Millisecond)
	eventBus.Stop()

	fmt.Println()
}

// =============================================================================
// 5. 心跳和健康检查模式
// =============================================================================

// HealthChecker 健康检查器
type HealthChecker struct {
	name      string
	checkFunc func() bool
	interval  time.Duration
	timeout   time.Duration
	status    string
	lastCheck time.Time
	mu        sync.RWMutex
	statusCh  chan string
	quit      chan bool
}

func NewHealthChecker(name string, checkFunc func() bool, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		name:      name,
		checkFunc: checkFunc,
		interval:  interval,
		timeout:   timeout,
		status:    "unknown",
		statusCh:  make(chan string, 1),
		quit:      make(chan bool),
	}
}

func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				hc.performCheck()

			case <-hc.quit:
				fmt.Printf("健康检查器 %s 停止\n", hc.name)
				return
			}
		}
	}()
}

func (hc *HealthChecker) performCheck() {
	fmt.Printf("开始健康检查: %s\n", hc.name)

	// 使用通道实现超时检查
	resultCh := make(chan bool, 1)

	go func() {
		result := hc.checkFunc()
		select {
		case resultCh <- result:
		default:
			// 如果主goroutine已经超时，这里就不发送了
		}
	}()

	select {
	case result := <-resultCh:
		hc.updateStatus(result)
	case <-time.After(hc.timeout):
		hc.updateStatus(false)
		fmt.Printf("健康检查超时: %s\n", hc.name)
	}
}

func (hc *HealthChecker) updateStatus(isHealthy bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	oldStatus := hc.status
	if isHealthy {
		hc.status = "healthy"
	} else {
		hc.status = "unhealthy"
	}
	hc.lastCheck = time.Now()

	// 如果状态发生变化，发送通知
	if oldStatus != hc.status {
		select {
		case hc.statusCh <- hc.status:
		default:
			// 状态通道满了，丢弃旧状态
		}
		fmt.Printf("健康检查器 %s 状态变化: %s -> %s\n", hc.name, oldStatus, hc.status)
	}

	fmt.Printf("健康检查完成: %s - %s\n", hc.name, hc.status)
}

func (hc *HealthChecker) GetStatus() (string, time.Time) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status, hc.lastCheck
}

func (hc *HealthChecker) GetStatusChannel() <-chan string {
	return hc.statusCh
}

func (hc *HealthChecker) Stop() {
	close(hc.quit)
}

// ServiceMonitor 服务监控器
type ServiceMonitor struct {
	checkers []*HealthChecker
	quit     chan bool
}

func NewServiceMonitor() *ServiceMonitor {
	return &ServiceMonitor{
		quit: make(chan bool),
	}
}

func (sm *ServiceMonitor) AddChecker(checker *HealthChecker) {
	sm.checkers = append(sm.checkers, checker)
}

func (sm *ServiceMonitor) Start() {
	// 启动所有健康检查器
	for _, checker := range sm.checkers {
		checker.Start()
	}

	// 监控状态变化
	go sm.monitorStatusChanges()
}

func (sm *ServiceMonitor) monitorStatusChanges() {
	cases := make([]reflect.SelectCase, len(sm.checkers)+1)

	// 为每个健康检查器创建 select case
	for i, checker := range sm.checkers {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(checker.GetStatusChannel()),
		}
	}

	// 添加退出 case
	cases[len(sm.checkers)] = reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(sm.quit),
	}

	for {
		chosen, value, ok := reflect.Select(cases)

		if chosen == len(sm.checkers) {
			// 收到退出信号
			fmt.Println("服务监控器停止")
			return
		}

		if !ok {
			// 通道关闭
			continue
		}

		status := value.String()
		checkerName := sm.checkers[chosen].name
		fmt.Printf("监控器收到状态变化: %s - %s\n", checkerName, status)

		// 可以在这里添加告警逻辑
		if status == "unhealthy" {
			fmt.Printf("🚨 告警: 服务 %s 不健康!\n", checkerName)
		} else {
			fmt.Printf("✅ 恢复: 服务 %s 恢复健康\n", checkerName)
		}
	}
}

func (sm *ServiceMonitor) Stop() {
	for _, checker := range sm.checkers {
		checker.Stop()
	}
	close(sm.quit)
}

func (sm *ServiceMonitor) GetOverallStatus() map[string]string {
	status := make(map[string]string)
	for _, checker := range sm.checkers {
		st, _ := checker.GetStatus()
		status[checker.name] = st
	}
	return status
}

func demonstrateHealthCheckPattern() {
	fmt.Println("=== 4. 心跳和健康检查模式 ===")

	monitor := NewServiceMonitor()

	// 模拟数据库健康检查
	dbChecker := NewHealthChecker("数据库", func() bool {
		// 模拟数据库连接检查
		return rand.Float32() > 0.3 // 70% 成功率
	}, 1*time.Second, 500*time.Millisecond)

	// 模拟API健康检查
	apiChecker := NewHealthChecker("API服务", func() bool {
		// 模拟API响应检查
		return rand.Float32() > 0.2 // 80% 成功率
	}, 1*time.Second, 800*time.Millisecond)

	// 模拟缓存健康检查
	cacheChecker := NewHealthChecker("缓存服务", func() bool {
		// 模拟缓存连接检查
		return rand.Float32() > 0.1 // 90% 成功率
	}, 2*time.Second, 300*time.Millisecond)

	monitor.AddChecker(dbChecker)
	monitor.AddChecker(apiChecker)
	monitor.AddChecker(cacheChecker)

	monitor.Start()

	// 运行 8 秒观察健康检查
	time.Sleep(8 * time.Second)

	// 显示最终状态
	fmt.Println("\n最终健康状态:")
	overallStatus := monitor.GetOverallStatus()
	for service, status := range overallStatus {
		fmt.Printf("  %s: %s\n", service, status)
	}

	monitor.Stop()
	fmt.Println()
}

// =============================================================================
// 6. 优雅关闭模式
// =============================================================================

// GracefulServer 支持优雅关闭的服务器
type GracefulServer struct {
	name     string
	shutdown chan bool
	done     chan bool
	wg       sync.WaitGroup
}

func NewGracefulServer(name string) *GracefulServer {
	return &GracefulServer{
		name:     name,
		shutdown: make(chan bool),
		done:     make(chan bool),
	}
}

func (gs *GracefulServer) Start() {
	fmt.Printf("服务器 %s 启动\n", gs.name)

	// 启动多个工作goroutine
	for i := 1; i <= 3; i++ {
		gs.wg.Add(1)
		go gs.worker(i)
	}

	// 监听关闭信号
	go gs.gracefulShutdown()
}

func (gs *GracefulServer) worker(id int) {
	defer gs.wg.Done()

	workerName := fmt.Sprintf("%s-Worker%d", gs.name, id)
	fmt.Printf("%s 开始工作\n", workerName)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 模拟处理请求
			fmt.Printf("%s 处理请求\n", workerName)

		case <-gs.shutdown:
			fmt.Printf("%s 收到关闭信号，开始清理...\n", workerName)

			// 模拟清理工作
			time.Sleep(200 * time.Millisecond)

			fmt.Printf("%s 清理完成，退出\n", workerName)
			return
		}
	}
}

func (gs *GracefulServer) gracefulShutdown() {
	<-gs.shutdown

	fmt.Printf("服务器 %s 开始优雅关闭\n", gs.name)

	// 等待所有工作goroutine完成
	gs.wg.Wait()

	fmt.Printf("服务器 %s 优雅关闭完成\n", gs.name)
	close(gs.done)
}

func (gs *GracefulServer) Shutdown() {
	close(gs.shutdown)
	<-gs.done
}

func demonstrateGracefulShutdown() {
	fmt.Println("=== 5. 优雅关闭模式 ===")

	// 创建多个服务器
	servers := []*GracefulServer{
		NewGracefulServer("WebServer"),
		NewGracefulServer("APIServer"),
		NewGracefulServer("TaskServer"),
	}

	// 启动所有服务器
	for _, server := range servers {
		server.Start()
	}

	fmt.Println("所有服务器运行 3 秒...")
	time.Sleep(3 * time.Second)

	// 优雅关闭所有服务器
	fmt.Println("开始优雅关闭所有服务器...")

	var shutdownWg sync.WaitGroup
	for _, server := range servers {
		shutdownWg.Add(1)
		go func(s *GracefulServer) {
			defer shutdownWg.Done()
			s.Shutdown()
		}(server)
	}

	shutdownWg.Wait()
	fmt.Println("所有服务器已优雅关闭")
	fmt.Println()
}

// =============================================================================
// 7. Select 最佳实践和性能考虑
// =============================================================================

func demonstrateSelectBestPractices() {
	fmt.Println("=== 6. Select 最佳实践和性能考虑 ===")

	fmt.Println("1. Select 语句最佳实践:")
	fmt.Println("   ✓ 使用 default 分支实现非阻塞操作")
	fmt.Println("   ✓ 合理使用 time.After 进行超时控制")
	fmt.Println("   ✓ 避免在循环中创建 time.After")
	fmt.Println("   ✓ 使用 time.NewTimer 和 timer.Reset() 重用定时器")

	fmt.Println("\n2. 性能考虑:")

	// 演示错误的做法：在循环中使用 time.After
	fmt.Println("❌ 错误做法 - 在循环中使用 time.After:")
	badExample := func() {
		ch := make(chan int, 1)
		ch <- 1

		start := time.Now()
		for i := 0; i < 1000; i++ {
			select {
			case <-ch:
				// 处理消息
			case <-time.After(time.Millisecond): // 每次都创建新的定时器
				// 超时处理
			}
		}
		fmt.Printf("  耗时: %v (创建了1000个定时器)\n", time.Since(start))
	}

	// 演示正确的做法：重用定时器
	fmt.Println("✓ 正确做法 - 重用定时器:")
	goodExample := func() {
		ch := make(chan int, 1)
		ch <- 1

		start := time.Now()
		timer := time.NewTimer(time.Millisecond)
		defer timer.Stop()

		for i := 0; i < 1000; i++ {
			select {
			case <-ch:
				// 处理消息
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(time.Millisecond)
			case <-timer.C:
				// 超时处理
				timer.Reset(time.Millisecond)
			}
		}
		fmt.Printf("  耗时: %v (重用了1个定时器)\n", time.Since(start))
	}

	badExample()
	goodExample()

	fmt.Println("\n3. 处理 nil 通道:")
	nilChannelExample := func() {
		var ch1 chan int
		ch2 := make(chan int, 1)
		ch2 <- 42

		select {
		case <-ch1: // nil 通道，永远不会被选中
			fmt.Println("从 nil 通道接收")
		case val := <-ch2:
			fmt.Printf("从正常通道接收: %d\n", val)
		default:
			fmt.Println("默认分支")
		}
	}
	nilChannelExample()

	fmt.Println("\n4. 随机性演示:")
	randomExample := func() {
		ch1 := make(chan int, 1)
		ch2 := make(chan int, 1)

		// 同时准备两个通道
		ch1 <- 1
		ch2 <- 2

		results := make(map[string]int)

		for i := 0; i < 10; i++ {
			select {
			case <-ch1:
				results["ch1"]++
				ch1 <- 1 // 重新准备
			case <-ch2:
				results["ch2"]++
				ch2 <- 2 // 重新准备
			}
		}

		fmt.Printf("随机选择结果: ch1=%d, ch2=%d\n", results["ch1"], results["ch2"])
	}
	randomExample()

	fmt.Println("\n5. 常见陷阱:")
	fmt.Println("   - 忘记处理通道关闭")
	fmt.Println("   - 在热路径中使用 time.After")
	fmt.Println("   - select 中的 case 顺序不影响执行")
	fmt.Println("   - default 分支使 select 变为非阻塞")

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - 高级 Select 模式")
	fmt.Println("==============================")

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	demonstrateBasicSelect()
	demonstrateMultiProducerSingleConsumer()
	demonstrateEventBusPattern()
	demonstrateHealthCheckPattern()
	demonstrateGracefulShutdown()
	demonstrateSelectBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个支持负载均衡的请求分发器")
	fmt.Println("2. 创建一个事件驱动的状态机")
	fmt.Println("3. 实现一个支持优先级的消息队列")
	fmt.Println("4. 编写一个分布式任务协调器")
	fmt.Println("5. 创建一个实时数据聚合系统")
	fmt.Println("6. 实现一个支持动态配置的服务发现系统")
	fmt.Println("\n请在此基础上练习更多高级 Select 模式的使用！")
}
