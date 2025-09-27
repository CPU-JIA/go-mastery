package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// 安全随机数生成函数
func secureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// 安全fallback：使用时间戳
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano() % int64(max)
		// 检查是否在int范围内
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

func secureRandomFloat32() float32 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		// 安全fallback：使用时间戳
		return float32(time.Now().UnixNano()%1000) / 1000.0
	}
	return float32(n.Int64()) / float32(1<<24)
}

// =============================================================================
// 1. Channel 基础概念
// =============================================================================

/*
Channel（通道）是 Go 语言并发编程的核心机制：

1. 定义：
   - Channel 是 goroutine 之间通信的管道
   - 实现了"不要通过共享内存来通信，要通过通信来共享内存"的哲学
   - 提供了同步和数据传递的功能

2. 类型：
   - 无缓冲通道：同步通道，发送和接收操作会阻塞直到另一方准备好
   - 有缓冲通道：异步通道，有一个缓冲区，满了才阻塞发送，空了才阻塞接收

3. 语法：
   - 声明：var ch chan int
   - 创建：ch := make(chan int) 或 ch := make(chan int, bufferSize)
   - 发送：ch <- value
   - 接收：value := <-ch 或 value, ok := <-ch
   - 关闭：close(ch)

4. 特性：
   - 发送到已关闭的通道会导致 panic
   - 从已关闭的通道接收会立即返回零值
   - 只有发送方应该关闭通道
   - 通道可以用于 select 语句

5. 应用场景：
   - 数据传递
   - 事件通知
   - 任务分发
   - 结果收集
   - 同步控制
*/

// =============================================================================
// 2. 基础 Channel 操作
// =============================================================================

func demonstrateBasicChannels() {
	fmt.Println("=== 1. 基础 Channel 操作 ===")

	// 创建无缓冲通道
	ch := make(chan string)

	// 启动一个 goroutine 发送数据
	go func() {
		fmt.Println("发送方：准备发送数据")
		ch <- "Hello, Channel!"
		fmt.Println("发送方：数据已发送")
	}()

	// 在主 goroutine 中接收数据
	fmt.Println("接收方：准备接收数据")
	message := <-ch
	fmt.Printf("接收方：收到消息 '%s'\n", message)

	// 演示双向通信
	response := make(chan string)

	go func() {
		msg := <-ch // 接收消息
		fmt.Printf("处理器：收到消息 '%s'\n", msg)

		// 处理并返回响应
		result := "处理完成：" + msg
		response <- result
	}()

	// 发送消息并等待响应
	ch <- "需要处理的数据"
	result := <-response
	fmt.Printf("主程序：收到响应 '%s'\n", result)

	fmt.Println()
}

// =============================================================================
// 3. 有缓冲 vs 无缓冲通道
// =============================================================================

func demonstrateBufferedChannels() {
	fmt.Println("=== 2. 有缓冲 vs 无缓冲通道 ===")

	// 无缓冲通道演示
	fmt.Println("无缓冲通道演示:")
	unbufferedCh := make(chan int)

	go func() {
		fmt.Println("  发送方：准备发送到无缓冲通道")
		unbufferedCh <- 1
		fmt.Println("  发送方：发送完成（说明接收方已准备好）")
	}()

	time.Sleep(100 * time.Millisecond) // 让发送方先启动
	fmt.Println("  接收方：准备接收")
	value := <-unbufferedCh
	fmt.Printf("  接收方：收到值 %d\n", value)

	// 有缓冲通道演示
	fmt.Println("\n有缓冲通道演示:")
	bufferedCh := make(chan int, 3) // 缓冲区大小为 3

	go func() {
		for i := 1; i <= 5; i++ {
			fmt.Printf("  发送方：发送 %d\n", i)
			bufferedCh <- i

			if i <= 3 {
				fmt.Printf("  发送方：%d 发送成功（缓冲区未满）\n", i)
			} else {
				fmt.Printf("  发送方：%d 发送完成（等待了接收方）\n", i)
			}
		}
		close(bufferedCh)
	}()

	time.Sleep(200 * time.Millisecond) // 让前几个值进入缓冲区

	for value := range bufferedCh {
		fmt.Printf("  接收方：收到 %d\n", value)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println()
}

// =============================================================================
// 4. Channel 方向性（只读/只写）
// =============================================================================

// sender 只能发送的通道参数
func sender(ch chan<- int, start, count int) {
	defer close(ch)

	for i := start; i < start+count; i++ {
		fmt.Printf("发送: %d\n", i)
		ch <- i
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("发送完成，通道已关闭")
}

// receiver 只能接收的通道参数
func receiver(ch <-chan int, name string) {
	fmt.Printf("%s 开始接收\n", name)

	for value := range ch {
		fmt.Printf("%s 收到: %d\n", name, value)
	}

	fmt.Printf("%s 接收完成\n", name)
}

// processor 既能接收又能发送
func processor(input <-chan int, output chan<- int) {
	defer close(output)

	for value := range input {
		processed := value * value // 计算平方
		fmt.Printf("处理: %d -> %d\n", value, processed)
		output <- processed
	}
	fmt.Println("处理完成")
}

func demonstrateChannelDirections() {
	fmt.Println("=== 3. Channel 方向性 ===")

	// 创建通道
	inputCh := make(chan int)
	outputCh := make(chan int)

	// 启动发送方
	go sender(inputCh, 1, 5)

	// 启动处理器
	go processor(inputCh, outputCh)

	// 启动接收方
	go receiver(outputCh, "接收器")

	time.Sleep(2 * time.Second)
	fmt.Println()
}

// =============================================================================
// 5. Select 语句
// =============================================================================

func demonstrateSelect() {
	fmt.Println("=== 4. Select 语句 ===")

	ch1 := make(chan string)
	ch2 := make(chan string)
	done := make(chan bool)

	// 启动两个发送方
	go func() {
		time.Sleep(200 * time.Millisecond)
		ch1 <- "来自通道1的消息"
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		ch2 <- "来自通道2的消息"
	}()

	// 使用 select 等待第一个可用的通道
	fmt.Println("等待消息...")

	select {
	case msg1 := <-ch1:
		fmt.Printf("收到通道1的消息: %s\n", msg1)
	case msg2 := <-ch2:
		fmt.Printf("收到通道2的消息: %s\n", msg2)
	case <-time.After(1 * time.Second):
		fmt.Println("超时：1秒内没有收到任何消息")
	}

	// 演示非阻塞接收
	fmt.Println("\n非阻塞接收演示:")

	for i := 0; i < 5; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("收到通道1: %s\n", msg)
		case msg := <-ch2:
			fmt.Printf("收到通道2: %s\n", msg)
		default:
			fmt.Printf("第%d次检查：没有消息可接收\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// 启动一个超时控制的 goroutine
	go func() {
		timer := time.NewTimer(1 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				fmt.Println("定时器：1秒到了，发送完成信号")
				done <- true
				return
			case msg := <-ch2: // ch2 可能还有延迟的消息
				fmt.Printf("延迟收到通道2: %s\n", msg)
				timer.Reset(1 * time.Second) // 重置定时器
			}
		}
	}()

	<-done
	fmt.Println()
}

// =============================================================================
// 6. 扇入扇出模式
// =============================================================================

// fanOut 扇出：将一个输入分发到多个输出
func fanOut(input <-chan int, output1, output2 chan<- int) {
	defer close(output1)
	defer close(output2)

	for value := range input {
		// 随机选择一个输出通道
		if secureRandomInt(2) == 0 {
			fmt.Printf("扇出到通道1: %d\n", value)
			output1 <- value
		} else {
			fmt.Printf("扇出到通道2: %d\n", value)
			output2 <- value
		}
	}
}

// fanIn 扇入：将多个输入合并到一个输出
func fanIn(input1, input2 <-chan int, output chan<- int) {
	defer close(output)

	var wg sync.WaitGroup
	wg.Add(2)

	// 处理第一个输入通道
	go func() {
		defer wg.Done()
		for value := range input1 {
			fmt.Printf("扇入来自通道1: %d\n", value)
			output <- value
		}
	}()

	// 处理第二个输入通道
	go func() {
		defer wg.Done()
		for value := range input2 {
			fmt.Printf("扇入来自通道2: %d\n", value)
			output <- value
		}
	}()

	wg.Wait()
}

func demonstrateFanInFanOut() {
	fmt.Println("=== 5. 扇入扇出模式 ===")

	// 创建通道
	source := make(chan int)
	fanOut1 := make(chan int)
	fanOut2 := make(chan int)
	result := make(chan int)

	// 启动扇出
	go fanOut(source, fanOut1, fanOut2)

	// 启动扇入
	go fanIn(fanOut1, fanOut2, result)

	// 发送数据
	go func() {
		defer close(source)
		for i := 1; i <= 10; i++ {
			fmt.Printf("发送到源: %d\n", i)
			source <- i
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// 接收最终结果
	fmt.Println("最终结果:")
	for value := range result {
		fmt.Printf("最终收到: %d\n", value)
	}

	fmt.Println()
}

// =============================================================================
// 7. 管道模式
// =============================================================================

// generateNumbers 数字生成器
func generateNumbers(max int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for i := 1; i <= max; i++ {
			out <- i
		}
	}()

	return out
}

// squareNumbers 计算平方
func squareNumbers(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()

	return out
}

// filterEven 过滤偶数
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()

	return out
}

// addPrefix 添加前缀
func addPrefix(in <-chan int, prefix string) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		for n := range in {
			out <- fmt.Sprintf("%s%d", prefix, n)
		}
	}()

	return out
}

func demonstratePipeline() {
	fmt.Println("=== 6. 管道模式 ===")

	// 构建管道：生成数字 -> 计算平方 -> 过滤偶数 -> 添加前缀
	numbers := generateNumbers(10)
	squares := squareNumbers(numbers)
	evens := filterEven(squares)
	prefixed := addPrefix(evens, "偶数平方: ")

	// 接收最终结果
	fmt.Println("管道处理结果:")
	for result := range prefixed {
		fmt.Println(result)
	}

	fmt.Println()
}

// =============================================================================
// 8. 实际应用：任务队列
// =============================================================================

// Task 任务结构体
type Task struct {
	ID       int
	Data     string
	Priority int
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID   int
	Result   string
	Duration time.Duration
	Error    error
}

// TaskQueue 任务队列
type TaskQueue struct {
	tasks   chan Task
	results chan TaskResult
	workers int
	quit    chan bool
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(bufferSize, workers int) *TaskQueue {
	return &TaskQueue{
		tasks:   make(chan Task, bufferSize),
		results: make(chan TaskResult, bufferSize),
		workers: workers,
		quit:    make(chan bool),
	}
}

// Start 启动任务队列
func (tq *TaskQueue) Start() {
	fmt.Printf("启动 %d 个工作者\n", tq.workers)

	for i := 1; i <= tq.workers; i++ {
		go tq.worker(i)
	}
}

// worker 工作者
func (tq *TaskQueue) worker(id int) {
	fmt.Printf("工作者 %d 已启动\n", id)

	for {
		select {
		case task := <-tq.tasks:
			start := time.Now()
			fmt.Printf("工作者 %d 开始处理任务 %d: %s\n", id, task.ID, task.Data)

			// 模拟任务处理
			processingTime := time.Duration(secureRandomInt(1000)) * time.Millisecond
			time.Sleep(processingTime)

			// 模拟可能的错误
			var err error
			result := fmt.Sprintf("任务 %d 处理完成", task.ID)
			if secureRandomFloat32() < 0.1 { // 10% 错误率
				err = fmt.Errorf("任务 %d 处理失败", task.ID)
				result = ""
			}

			tq.results <- TaskResult{
				TaskID:   task.ID,
				Result:   result,
				Duration: time.Since(start),
				Error:    err,
			}

		case <-tq.quit:
			fmt.Printf("工作者 %d 收到退出信号\n", id)
			return
		}
	}
}

// Submit 提交任务
func (tq *TaskQueue) Submit(task Task) {
	tq.tasks <- task
}

// GetResult 获取结果
func (tq *TaskQueue) GetResult() TaskResult {
	return <-tq.results
}

// Stop 停止任务队列
func (tq *TaskQueue) Stop() {
	close(tq.tasks)
	for i := 0; i < tq.workers; i++ {
		tq.quit <- true
	}
	close(tq.results)
}

func demonstrateTaskQueue() {
	fmt.Println("=== 7. 实际应用：任务队列 ===")

	// 创建任务队列
	queue := NewTaskQueue(10, 3)
	queue.Start()

	// 提交任务
	tasks := []Task{
		{1, "处理用户数据", 1},
		{2, "发送邮件", 2},
		{3, "生成报告", 1},
		{4, "备份数据", 3},
		{5, "清理缓存", 2},
		{6, "更新索引", 1},
		{7, "同步数据", 2},
		{8, "压缩日志", 3},
	}

	fmt.Printf("提交 %d 个任务\n", len(tasks))
	for _, task := range tasks {
		queue.Submit(task)
	}

	// 收集结果
	fmt.Println("\n任务执行结果:")
	successCount := 0
	var totalDuration time.Duration

	for i := 0; i < len(tasks); i++ {
		result := queue.GetResult()

		if result.Error != nil {
			fmt.Printf("❌ 任务 %d 失败: %v (耗时: %v)\n",
				result.TaskID, result.Error, result.Duration)
		} else {
			fmt.Printf("✅ %s (耗时: %v)\n",
				result.Result, result.Duration)
			successCount++
		}

		totalDuration += result.Duration
	}

	// 统计信息
	fmt.Printf("\n统计信息:\n")
	fmt.Printf("成功: %d/%d\n", successCount, len(tasks))
	fmt.Printf("成功率: %.1f%%\n", float64(successCount)/float64(len(tasks))*100)
	fmt.Printf("平均处理时间: %v\n", totalDuration/time.Duration(len(tasks)))

	// 停止队列
	queue.Stop()
	fmt.Println("任务队列已停止")
	fmt.Println()
}

// =============================================================================
// 9. Channel 最佳实践
// =============================================================================

func demonstrateChannelBestPractices() {
	fmt.Println("=== 8. Channel 最佳实践 ===")

	fmt.Println("1. 通道所有权:")
	fmt.Println("   - 谁创建通道，谁负责关闭")
	fmt.Println("   - 通常是发送方关闭通道")
	fmt.Println("   - 接收方检查通道是否关闭")

	fmt.Println("\n2. 避免死锁:")
	fmt.Println("   - 确保有对应的接收方和发送方")
	fmt.Println("   - 使用缓冲通道避免阻塞")
	fmt.Println("   - 使用 select 和 default 实现非阻塞操作")

	fmt.Println("\n3. 性能考虑:")
	fmt.Println("   - 无缓冲通道有同步开销")
	fmt.Println("   - 缓冲通道减少阻塞但增加内存使用")
	fmt.Println("   - 选择合适的缓冲区大小")

	fmt.Println("\n4. 错误处理:")
	fmt.Println("   - 通过通道传递错误信息")
	fmt.Println("   - 使用结构体包装结果和错误")
	fmt.Println("   - 考虑使用 context 进行取消控制")

	fmt.Println("\n5. 常见模式:")
	fmt.Println("   - 生产者-消费者模式")
	fmt.Println("   - 扇入扇出模式")
	fmt.Println("   - 管道模式")
	fmt.Println("   - 工作池模式")

	// 演示正确的错误处理模式
	fmt.Println("\n错误处理示例:")

	type Result struct {
		Value string
		Error error
	}

	resultCh := make(chan Result, 1)

	go func() {
		defer close(resultCh)

		// 模拟可能失败的操作
		if secureRandomFloat32() < 0.5 {
			resultCh <- Result{Error: fmt.Errorf("操作失败")}
		} else {
			resultCh <- Result{Value: "操作成功"}
		}
	}()

	result := <-resultCh
	if result.Error != nil {
		fmt.Printf("错误: %v\n", result.Error)
	} else {
		fmt.Printf("成功: %s\n", result.Value)
	}

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - Channel 通道")
	fmt.Println("===========================")

	// 设置随机种子
	// 注意：crypto/rand不需要设置种子

	demonstrateBasicChannels()
	demonstrateBufferedChannels()
	demonstrateChannelDirections()
	demonstrateSelect()
	demonstrateFanInFanOut()
	demonstratePipeline()
	demonstrateTaskQueue()
	demonstrateChannelBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个支持优先级的任务队列")
	fmt.Println("2. 创建一个消息路由器，根据消息类型分发到不同处理器")
	fmt.Println("3. 实现一个限流器，控制请求频率")
	fmt.Println("4. 编写一个并发安全的事件总线")
	fmt.Println("5. 创建一个数据聚合器，收集多个源的数据")
	fmt.Println("6. 实现一个批处理器，累积数据到一定数量后批量处理")
	fmt.Println("\n请在此基础上练习更多 Channel 的使用模式！")
}
