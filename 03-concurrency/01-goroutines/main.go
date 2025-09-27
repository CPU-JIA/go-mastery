package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
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
// 1. Goroutine 基础概念
// =============================================================================

/*
Goroutine 是 Go 语言并发编程的核心概念：

1. 定义：
   - Goroutine 是由 Go 运行时管理的轻量级线程
   - 比操作系统线程更轻量，启动成本更低
   - 一个程序可以同时运行成千上万个 goroutine

2. 特点：
   - 栈空间初始只有几KB，可以动态增长
   - 由 Go 调度器管理，而非操作系统
   - 通过通道(channel)进行通信
   - 遵循"不要通过共享内存来通信，要通过通信来共享内存"的哲学

3. 创建方式：
   - 使用 go 关键字启动 goroutine
   - go function_name(parameters)
   - go func() { ... }()

4. 生命周期：
   - main 函数是主 goroutine
   - 当 main 函数结束时，所有 goroutine 都会被终止
   - 需要同步机制来等待 goroutine 完成

5. 调度器：
   - 使用 M:N 调度模型
   - M 个 goroutine 运行在 N 个操作系统线程上
   - 包含 G(Goroutine)、M(Machine/线程)、P(Processor/处理器) 三个核心概念
*/

// =============================================================================
// 2. 基础 Goroutine 示例
// =============================================================================

// sayHello 简单的打印函数
func sayHello(name string, count int) {
	for i := 1; i <= count; i++ {
		fmt.Printf("Hello from %s - %d\n", name, i)
		time.Sleep(100 * time.Millisecond) // 模拟一些工作
	}
}

// worker 工作函数
func worker(id int, jobs <-chan int, results chan<- int) {
	for job := range jobs {
		fmt.Printf("Worker %d 开始处理任务 %d\n", id, job)

		// 模拟工作
		time.Sleep(time.Duration(secureRandomInt(1000)) * time.Millisecond)

		// 计算结果（这里简单地将任务ID乘以2）
		result := job * 2

		fmt.Printf("Worker %d 完成任务 %d，结果：%d\n", id, job, result)
		results <- result
	}
}

func demonstrateBasicGoroutines() {
	fmt.Println("=== 1. 基础 Goroutine 示例 ===")

	// 获取当前 goroutine 数量
	fmt.Printf("程序开始时的 goroutine 数量: %d\n", runtime.NumGoroutine())

	// 启动多个 goroutine
	go sayHello("Goroutine-1", 3)
	go sayHello("Goroutine-2", 3)
	go sayHello("Goroutine-3", 3)

	// 查看启动后的 goroutine 数量
	fmt.Printf("启动 goroutine 后的数量: %d\n", runtime.NumGoroutine())

	// 等待一段时间让 goroutine 执行
	time.Sleep(2 * time.Second)

	// 使用匿名 goroutine
	go func() {
		fmt.Println("这是一个匿名 goroutine")
		for i := 1; i <= 3; i++ {
			fmt.Printf("匿名 goroutine - %d\n", i)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	// 等待匿名 goroutine 完成
	time.Sleep(1 * time.Second)

	fmt.Printf("程序结束时的 goroutine 数量: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// =============================================================================
// 3. WaitGroup 同步
// =============================================================================

func demonstrateWaitGroup() {
	fmt.Println("=== 2. WaitGroup 同步机制 ===")

	var wg sync.WaitGroup

	// 启动多个 goroutine 并使用 WaitGroup 等待
	for i := 1; i <= 5; i++ {
		wg.Add(1) // 增加等待计数

		go func(id int) {
			defer wg.Done() // 完成时减少计数

			fmt.Printf("Goroutine %d 开始工作\n", id)

			// 模拟不同的工作时间
			workTime := time.Duration(secureRandomInt(1000)) * time.Millisecond
			time.Sleep(workTime)

			fmt.Printf("Goroutine %d 完成工作，耗时 %v\n", id, workTime)
		}(i)
	}

	fmt.Println("等待所有 goroutine 完成...")
	wg.Wait() // 等待所有 goroutine 完成
	fmt.Println("所有 goroutine 已完成")

	fmt.Println()
}

// =============================================================================
// 4. 工作池模式
// =============================================================================

func demonstrateWorkerPool() {
	fmt.Println("=== 3. 工作池模式 ===")

	const numWorkers = 3
	const numJobs = 9

	// 创建任务和结果通道
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// 启动工作者
	fmt.Printf("启动 %d 个工作者\n", numWorkers)
	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	// 发送任务
	fmt.Printf("发送 %d 个任务\n", numJobs)
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs) // 关闭任务通道

	// 收集结果
	fmt.Println("收集结果:")
	for r := 1; r <= numJobs; r++ {
		result := <-results
		fmt.Printf("收到结果: %d\n", result)
	}

	fmt.Println()
}

// =============================================================================
// 5. Goroutine 泄露示例和避免方法
// =============================================================================

// leakyGoroutine 演示会导致泄露的 goroutine
func leakyGoroutine() {
	fmt.Println("=== 4. Goroutine 泄露示例 ===")

	fmt.Printf("开始前的 goroutine 数量: %d\n", runtime.NumGoroutine())

	// 创建一个永远不会被读取的通道
	ch := make(chan int)

	// 启动会泄露的 goroutine
	for i := 0; i < 5; i++ {
		go func(id int) {
			fmt.Printf("泄露的 Goroutine %d 尝试发送数据\n", id)
			ch <- id                                  // 这里会永远阻塞，因为没有接收者
			fmt.Printf("泄露的 Goroutine %d 发送完成\n", id) // 这行永远不会执行
		}(i)
	}

	time.Sleep(100 * time.Millisecond) // 让 goroutine 启动
	fmt.Printf("泄露后的 goroutine 数量: %d\n", runtime.NumGoroutine())

	// 这些 goroutine 将永远阻塞，造成内存泄露
	fmt.Println("注意：这些 goroutine 会一直阻塞，造成泄露")
	fmt.Println()
}

// properGoroutineUsage 演示正确的 goroutine 使用方法
func properGoroutineUsage() {
	fmt.Println("=== 5. 正确的 Goroutine 使用 ===")

	fmt.Printf("开始前的 goroutine 数量: %d\n", runtime.NumGoroutine())

	// 使用带缓冲的通道避免阻塞
	ch := make(chan int, 5)

	var wg sync.WaitGroup

	// 启动生产者 goroutine
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d 发送数据\n", id)
			ch <- id
		}(i)
	}

	// 启动消费者 goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			data := <-ch
			fmt.Printf("接收到数据: %d\n", data)
		}
	}()

	wg.Wait()
	close(ch)

	fmt.Printf("正确使用后的 goroutine 数量: %d\n", runtime.NumGoroutine())
	fmt.Println()
}

// =============================================================================
// 6. 上下文取消和超时
// =============================================================================

// cancelableWorker 支持取消的工作者
func cancelableWorker(id int, done <-chan bool) {
	for {
		select {
		case <-done:
			fmt.Printf("Worker %d 收到取消信号，正在退出\n", id)
			return
		default:
			fmt.Printf("Worker %d 正在工作...\n", id)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func demonstrateCancellation() {
	fmt.Println("=== 6. Goroutine 取消机制 ===")

	done := make(chan bool)

	// 启动多个可取消的工作者
	for i := 1; i <= 3; i++ {
		go cancelableWorker(i, done)
	}

	// 让工作者运行一段时间
	fmt.Println("让工作者运行 2 秒...")
	time.Sleep(2 * time.Second)

	// 发送取消信号
	fmt.Println("发送取消信号...")
	close(done) // 关闭通道会向所有监听者发送零值

	// 等待一点时间让 goroutine 清理
	time.Sleep(1 * time.Second)
	fmt.Println("所有工作者已停止")
	fmt.Println()
}

// =============================================================================
// 7. Goroutine 性能测试
// =============================================================================

// heavyTask 模拟重任务
func heavyTask(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	// 模拟CPU密集型任务
	count := 0
	for i := 0; i < 1000000; i++ {
		count += i
	}

	fmt.Printf("任务 %d 完成，计算结果: %d\n", id, count)
}

func demonstratePerformance() {
	fmt.Println("=== 7. Goroutine 性能测试 ===")

	// 测试串行执行
	fmt.Println("串行执行 10 个重任务:")
	start := time.Now()

	var wg sync.WaitGroup
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		heavyTask(i, &wg) // 串行执行
	}
	wg.Wait()

	serialTime := time.Since(start)
	fmt.Printf("串行执行时间: %v\n", serialTime)

	// 测试并行执行
	fmt.Println("\n并行执行 10 个重任务:")
	start = time.Now()

	var wg2 sync.WaitGroup
	for i := 1; i <= 10; i++ {
		wg2.Add(1)
		go heavyTask(i, &wg2) // 并行执行
	}
	wg2.Wait()

	parallelTime := time.Since(start)
	fmt.Printf("并行执行时间: %v\n", parallelTime)

	if serialTime > parallelTime {
		fmt.Printf("性能提升: %.2fx\n", float64(serialTime)/float64(parallelTime))
	}

	fmt.Println()
}

// =============================================================================
// 8. 实际应用：并发下载器
// =============================================================================

// DownloadTask 下载任务
type DownloadTask struct {
	ID  int
	URL string
}

// DownloadResult 下载结果
type DownloadResult struct {
	TaskID   int
	Success  bool
	Duration time.Duration
	Error    error
}

// mockDownload 模拟下载函数
func mockDownload(task DownloadTask) DownloadResult {
	start := time.Now()

	// 模拟网络延迟
	delay := time.Duration(secureRandomInt(2000)) * time.Millisecond
	time.Sleep(delay)

	// 模拟随机失败
	success := secureRandomFloat32() > 0.2 // 80% 成功率

	result := DownloadResult{
		TaskID:   task.ID,
		Success:  success,
		Duration: time.Since(start),
	}

	if !success {
		result.Error = fmt.Errorf("下载失败: %s", task.URL)
	}

	return result
}

// ConcurrentDownloader 并发下载器
type ConcurrentDownloader struct {
	maxWorkers int
}

// NewConcurrentDownloader 创建并发下载器
func NewConcurrentDownloader(maxWorkers int) *ConcurrentDownloader {
	return &ConcurrentDownloader{maxWorkers: maxWorkers}
}

// Download 并发下载
func (cd *ConcurrentDownloader) Download(tasks []DownloadTask) []DownloadResult {
	taskCh := make(chan DownloadTask, len(tasks))
	resultCh := make(chan DownloadResult, len(tasks))

	// 启动工作者
	var wg sync.WaitGroup
	for i := 0; i < cd.maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range taskCh {
				fmt.Printf("Worker %d 开始下载: %s\n", workerID, task.URL)
				result := mockDownload(task)
				resultCh <- result
			}
		}(i)
	}

	// 发送任务
	go func() {
		for _, task := range tasks {
			taskCh <- task
		}
		close(taskCh)
	}()

	// 等待所有工作者完成
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 收集结果
	var results []DownloadResult
	for result := range resultCh {
		results = append(results, result)
	}

	return results
}

func demonstrateConcurrentDownloader() {
	fmt.Println("=== 8. 实际应用：并发下载器 ===")

	// 创建下载任务
	tasks := []DownloadTask{
		{1, "https://example.com/file1.jpg"},
		{2, "https://example.com/file2.jpg"},
		{3, "https://example.com/file3.jpg"},
		{4, "https://example.com/file4.jpg"},
		{5, "https://example.com/file5.jpg"},
		{6, "https://example.com/file6.jpg"},
		{7, "https://example.com/file7.jpg"},
		{8, "https://example.com/file8.jpg"},
	}

	// 创建并发下载器
	downloader := NewConcurrentDownloader(3)

	fmt.Printf("开始并发下载 %d 个文件，使用 %d 个工作者\n", len(tasks), 3)
	start := time.Now()

	results := downloader.Download(tasks)

	duration := time.Since(start)
	fmt.Printf("\n下载完成，总耗时: %v\n", duration)

	// 统计结果
	successCount := 0
	var totalDuration time.Duration

	fmt.Println("\n下载结果:")
	for _, result := range results {
		if result.Success {
			successCount++
			fmt.Printf("✓ 任务 %d 成功，耗时: %v\n", result.TaskID, result.Duration)
		} else {
			fmt.Printf("✗ 任务 %d 失败: %v\n", result.TaskID, result.Error)
		}
		totalDuration += result.Duration
	}

	fmt.Printf("\n统计信息:\n")
	fmt.Printf("成功: %d/%d\n", successCount, len(tasks))
	fmt.Printf("成功率: %.1f%%\n", float64(successCount)/float64(len(tasks))*100)
	fmt.Printf("平均下载时间: %v\n", totalDuration/time.Duration(len(tasks)))

	fmt.Println()
}

// =============================================================================
// 9. 最佳实践和常见陷阱
// =============================================================================

func demonstrateBestPractices() {
	fmt.Println("=== 9. Goroutine 最佳实践 ===")

	fmt.Println("1. 避免 Goroutine 泄露:")
	fmt.Println("   - 确保每个 goroutine 都有退出条件")
	fmt.Println("   - 使用 context 或 done channel 进行取消")
	fmt.Println("   - 避免在不会被读取的通道上发送数据")

	fmt.Println("\n2. 合理控制并发数:")
	fmt.Println("   - 使用工作池模式限制同时运行的 goroutine 数量")
	fmt.Println("   - 避免创建过多的 goroutine（通常不超过 CPU 核心数的几倍）")
	fmt.Println("   - 考虑使用 semaphore 或 worker pool")

	fmt.Println("\n3. 正确使用同步原语:")
	fmt.Println("   - 优先使用 channel 进行通信")
	fmt.Println("   - 在需要共享状态时才使用 mutex")
	fmt.Println("   - 正确使用 WaitGroup 等待 goroutine 完成")

	fmt.Println("\n4. 错误处理:")
	fmt.Println("   - Goroutine 中的 panic 不会被主程序捕获")
	fmt.Println("   - 使用 recover 在 goroutine 内部处理 panic")
	fmt.Println("   - 通过 channel 传递错误信息")

	fmt.Println("\n5. 性能考虑:")
	fmt.Println("   - Goroutine 切换有成本，不要过度创建")
	fmt.Println("   - CPU 密集型任务的并发数不应超过 CPU 核心数")
	fmt.Println("   - I/O 密集型任务可以有更多的并发数")

	fmt.Printf("\n当前系统信息:\n")
	fmt.Printf("CPU 核心数: %d\n", runtime.NumCPU())
	fmt.Printf("当前 Goroutine 数: %d\n", runtime.NumGoroutine())
	fmt.Printf("当前 GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - Goroutine 基础")
	fmt.Println("=============================")

	// 设置随机种子
	// 注意：crypto/rand不需要设置种子

	demonstrateBasicGoroutines()
	demonstrateWaitGroup()
	demonstrateWorkerPool()
	leakyGoroutine()
	properGoroutineUsage()
	demonstrateCancellation()
	demonstratePerformance()
	demonstrateConcurrentDownloader()
	demonstrateBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个并发的网页爬虫")
	fmt.Println("2. 创建一个支持限流的并发任务处理器")
	fmt.Println("3. 实现一个 goroutine 池来重用 goroutine")
	fmt.Println("4. 编写一个并发的文件处理程序")
	fmt.Println("5. 创建一个实时数据处理管道")
	fmt.Println("6. 实现一个支持优雅关闭的服务器")
	fmt.Println("\n请在此基础上练习更多 goroutine 的使用场景！")
}
