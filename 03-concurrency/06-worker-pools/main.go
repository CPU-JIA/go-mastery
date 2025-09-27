package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
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

// =============================================================================
// 1. Worker Pool 基础概念
// =============================================================================

/*
Worker Pool（工作池）是并发编程中的重要模式：

核心概念：
1. 固定数量的工作者（worker）
2. 任务队列（job queue）
3. 结果收集器（result collector）
4. 生命周期管理（启动、停止）

优势：
1. 控制并发数量，避免资源耗尽
2. 复用goroutine，减少创建/销毁开销
3. 提供背压（backpressure）机制
4. 便于监控和管理

常见模式：
1. 简单工作池
2. 带结果的工作池
3. 优先级工作池
4. 动态工作池
5. 链式工作池
6. 分阶段工作池

适用场景：
- CPU密集型任务
- I/O密集型任务
- 批处理作业
- 图片处理
- 数据处理管道
*/

// =============================================================================
// 2. 简单工作池
// =============================================================================

// Job 工作任务接口
type Job interface {
	Execute() interface{}
	GetID() string
}

// SimpleJob 简单任务实现
type SimpleJob struct {
	ID   string
	Data int
}

func (j *SimpleJob) Execute() interface{} {
	// 模拟工作负载
	result := j.Data * j.Data
	time.Sleep(time.Duration(secureRandomInt(100)) * time.Millisecond)
	return result
}

func (j *SimpleJob) GetID() string {
	return j.ID
}

// Result 工作结果
type Result struct {
	JobID  string
	Value  interface{}
	Error  error
	Worker int
}

// SimpleWorkerPool 简单工作池
type SimpleWorkerPool struct {
	workerCount int
	jobQueue    chan Job
	resultQueue chan Result
	quit        chan bool
	wg          sync.WaitGroup
}

// NewSimpleWorkerPool 创建简单工作池
func NewSimpleWorkerPool(workerCount, queueSize int) *SimpleWorkerPool {
	return &SimpleWorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize),
		resultQueue: make(chan Result, queueSize),
		quit:        make(chan bool),
	}
}

// Start 启动工作池
func (wp *SimpleWorkerPool) Start() {
	for i := 1; i <= wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	fmt.Printf("简单工作池启动，%d 个工作者\n", wp.workerCount)
}

// worker 工作者
func (wp *SimpleWorkerPool) worker(id int) {
	defer wp.wg.Done()

	fmt.Printf("工作者 %d 开始工作\n", id)

	for {
		select {
		case job := <-wp.jobQueue:
			fmt.Printf("工作者 %d 处理任务 %s\n", id, job.GetID())

			start := time.Now()
			result := job.Execute()
			duration := time.Since(start)

			wp.resultQueue <- Result{
				JobID:  job.GetID(),
				Value:  result,
				Worker: id,
			}

			fmt.Printf("工作者 %d 完成任务 %s，耗时 %v\n", id, job.GetID(), duration)

		case <-wp.quit:
			fmt.Printf("工作者 %d 收到停止信号\n", id)
			return
		}
	}
}

// Submit 提交任务
func (wp *SimpleWorkerPool) Submit(job Job) {
	wp.jobQueue <- job
}

// GetResult 获取结果
func (wp *SimpleWorkerPool) GetResult() Result {
	return <-wp.resultQueue
}

// Stop 停止工作池
func (wp *SimpleWorkerPool) Stop() {
	close(wp.jobQueue)

	// 发送停止信号给所有工作者
	for i := 0; i < wp.workerCount; i++ {
		wp.quit <- true
	}

	wp.wg.Wait()
	close(wp.resultQueue)
	close(wp.quit)

	fmt.Println("简单工作池已停止")
}

func demonstrateSimpleWorkerPool() {
	fmt.Println("=== 1. 简单工作池 ===")

	// 创建工作池
	pool := NewSimpleWorkerPool(3, 10)
	pool.Start()

	// 提交任务
	jobs := make([]*SimpleJob, 10)
	for i := 0; i < 10; i++ {
		jobs[i] = &SimpleJob{
			ID:   fmt.Sprintf("job-%d", i+1),
			Data: i + 1,
		}
		pool.Submit(jobs[i])
	}

	fmt.Printf("提交了 %d 个任务\n", len(jobs))

	// 收集结果
	results := make([]Result, len(jobs))
	for i := 0; i < len(jobs); i++ {
		result := pool.GetResult()
		results[i] = result
		fmt.Printf("收到结果: 任务 %s，值 %v，工作者 %d\n",
			result.JobID, result.Value, result.Worker)
	}

	pool.Stop()
	fmt.Println()
}

// =============================================================================
// 3. 带上下文的工作池
// =============================================================================

// ContextJob 支持上下文的任务
type ContextJob struct {
	ID      string
	Data    interface{}
	Process func(ctx context.Context, data interface{}) (interface{}, error)
}

func (j *ContextJob) GetID() string {
	return j.ID
}

// ContextWorkerPool 支持上下文的工作池
type ContextWorkerPool struct {
	workerCount int
	jobQueue    chan *ContextJob
	resultQueue chan Result
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewContextWorkerPool 创建支持上下文的工作池
func NewContextWorkerPool(ctx context.Context, workerCount, queueSize int) *ContextWorkerPool {
	ctx, cancel := context.WithCancel(ctx)

	return &ContextWorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan *ContextJob, queueSize),
		resultQueue: make(chan Result, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动工作池
func (wp *ContextWorkerPool) Start() {
	for i := 1; i <= wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	fmt.Printf("上下文工作池启动，%d 个工作者\n", wp.workerCount)
}

// worker 工作者
func (wp *ContextWorkerPool) worker(id int) {
	defer wp.wg.Done()

	fmt.Printf("上下文工作者 %d 开始工作\n", id)

	for {
		select {
		case job := <-wp.jobQueue:
			fmt.Printf("上下文工作者 %d 处理任务 %s\n", id, job.GetID())

			// 为任务创建子上下文
			jobCtx, jobCancel := context.WithTimeout(wp.ctx, 2*time.Second)

			start := time.Now()
			result, err := job.Process(jobCtx, job.Data)
			duration := time.Since(start)

			jobCancel()

			wp.resultQueue <- Result{
				JobID:  job.GetID(),
				Value:  result,
				Error:  err,
				Worker: id,
			}

			if err != nil {
				fmt.Printf("上下文工作者 %d 任务 %s 失败: %v (耗时 %v)\n",
					id, job.GetID(), err, duration)
			} else {
				fmt.Printf("上下文工作者 %d 完成任务 %s，耗时 %v\n",
					id, job.GetID(), duration)
			}

		case <-wp.ctx.Done():
			fmt.Printf("上下文工作者 %d 收到取消信号: %v\n", id, wp.ctx.Err())
			return
		}
	}
}

// Submit 提交任务
func (wp *ContextWorkerPool) Submit(job *ContextJob) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	}
}

// GetResult 获取结果
func (wp *ContextWorkerPool) GetResult() (Result, error) {
	select {
	case result := <-wp.resultQueue:
		return result, nil
	case <-wp.ctx.Done():
		return Result{}, wp.ctx.Err()
	}
}

// Stop 停止工作池
func (wp *ContextWorkerPool) Stop() {
	wp.cancel()
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.resultQueue)
	fmt.Println("上下文工作池已停止")
}

func demonstrateContextWorkerPool() {
	fmt.Println("=== 2. 带上下文的工作池 ===")

	// 创建带5秒超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool := NewContextWorkerPool(ctx, 3, 10)
	pool.Start()

	// 提交不同类型的任务
	jobs := []*ContextJob{
		{
			ID:   "fast-job",
			Data: "快速任务",
			Process: func(ctx context.Context, data interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond)
				return fmt.Sprintf("处理完成: %s", data), nil
			},
		},
		{
			ID:   "slow-job",
			Data: "慢速任务",
			Process: func(ctx context.Context, data interface{}) (interface{}, error) {
				select {
				case <-time.After(3 * time.Second):
					return fmt.Sprintf("处理完成: %s", data), nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		},
		{
			ID:   "error-job",
			Data: "错误任务",
			Process: func(ctx context.Context, data interface{}) (interface{}, error) {
				time.Sleep(50 * time.Millisecond)
				return nil, fmt.Errorf("处理失败: %s", data)
			},
		},
	}

	// 提交任务
	for _, job := range jobs {
		if err := pool.Submit(job); err != nil {
			fmt.Printf("提交任务 %s 失败: %v\n", job.GetID(), err)
		}
	}

	// 收集结果
	for i := 0; i < len(jobs); i++ {
		result, err := pool.GetResult()
		if err != nil {
			fmt.Printf("获取结果失败: %v\n", err)
			break
		}

		if result.Error != nil {
			fmt.Printf("任务 %s 执行失败: %v\n", result.JobID, result.Error)
		} else {
			fmt.Printf("任务 %s 执行成功: %v\n", result.JobID, result.Value)
		}
	}

	pool.Stop()
	fmt.Println()
}

// =============================================================================
// 4. 优先级工作池
// =============================================================================

// PriorityJob 优先级任务
type PriorityJob struct {
	ID       string
	Priority int
	Data     interface{}
	Process  func(interface{}) interface{}
}

func (j *PriorityJob) GetID() string {
	return j.ID
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	jobs []*PriorityJob
	mu   sync.Mutex
	cond *sync.Cond
}

func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{}
	pq.cond = sync.NewCond(&pq.mu)
	return pq
}

// Push 添加任务（按优先级排序）
func (pq *PriorityQueue) Push(job *PriorityJob) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// 找到插入位置（优先级高的在前面）
	i := 0
	for i < len(pq.jobs) && pq.jobs[i].Priority >= job.Priority {
		i++
	}

	// 插入任务
	pq.jobs = append(pq.jobs, nil)
	copy(pq.jobs[i+1:], pq.jobs[i:])
	pq.jobs[i] = job

	pq.cond.Signal()
}

// Pop 获取最高优先级任务
func (pq *PriorityQueue) Pop() *PriorityJob {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for len(pq.jobs) == 0 {
		pq.cond.Wait()
	}

	job := pq.jobs[0]
	pq.jobs = pq.jobs[1:]
	return job
}

// Len 队列长度
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.jobs)
}

// Close 关闭队列
func (pq *PriorityQueue) Close() {
	pq.cond.Broadcast()
}

// PriorityWorkerPool 优先级工作池
type PriorityWorkerPool struct {
	workerCount int
	queue       *PriorityQueue
	resultQueue chan Result
	quit        chan bool
	wg          sync.WaitGroup
}

// NewPriorityWorkerPool 创建优先级工作池
func NewPriorityWorkerPool(workerCount int) *PriorityWorkerPool {
	return &PriorityWorkerPool{
		workerCount: workerCount,
		queue:       NewPriorityQueue(),
		resultQueue: make(chan Result, workerCount*2),
		quit:        make(chan bool),
	}
}

// Start 启动工作池
func (wp *PriorityWorkerPool) Start() {
	for i := 1; i <= wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	fmt.Printf("优先级工作池启动，%d 个工作者\n", wp.workerCount)
}

// worker 工作者
func (wp *PriorityWorkerPool) worker(id int) {
	defer wp.wg.Done()

	fmt.Printf("优先级工作者 %d 开始工作\n", id)

	for {
		select {
		case <-wp.quit:
			fmt.Printf("优先级工作者 %d 收到停止信号\n", id)
			return
		default:
			job := wp.queue.Pop()
			if job == nil {
				continue
			}

			fmt.Printf("优先级工作者 %d 处理任务 %s (优先级: %d)\n",
				id, job.GetID(), job.Priority)

			start := time.Now()
			result := job.Process(job.Data)
			duration := time.Since(start)

			wp.resultQueue <- Result{
				JobID:  job.GetID(),
				Value:  result,
				Worker: id,
			}

			fmt.Printf("优先级工作者 %d 完成任务 %s，耗时 %v\n",
				id, job.GetID(), duration)
		}
	}
}

// Submit 提交任务
func (wp *PriorityWorkerPool) Submit(job *PriorityJob) {
	wp.queue.Push(job)
}

// GetResult 获取结果
func (wp *PriorityWorkerPool) GetResult() Result {
	return <-wp.resultQueue
}

// Stop 停止工作池
func (wp *PriorityWorkerPool) Stop() {
	// 关闭优先级队列
	wp.queue.Close()

	// 发送停止信号
	for i := 0; i < wp.workerCount; i++ {
		wp.quit <- true
	}

	wp.wg.Wait()
	close(wp.resultQueue)
	close(wp.quit)

	fmt.Println("优先级工作池已停止")
}

func demonstratePriorityWorkerPool() {
	fmt.Println("=== 3. 优先级工作池 ===")

	pool := NewPriorityWorkerPool(2)
	pool.Start()

	// 提交不同优先级的任务
	jobs := []*PriorityJob{
		{
			ID:       "low-1",
			Priority: 1,
			Data:     "低优先级任务1",
			Process: func(data interface{}) interface{} {
				time.Sleep(200 * time.Millisecond)
				return fmt.Sprintf("完成: %s", data)
			},
		},
		{
			ID:       "high-1",
			Priority: 10,
			Data:     "高优先级任务1",
			Process: func(data interface{}) interface{} {
				time.Sleep(100 * time.Millisecond)
				return fmt.Sprintf("完成: %s", data)
			},
		},
		{
			ID:       "medium-1",
			Priority: 5,
			Data:     "中优先级任务1",
			Process: func(data interface{}) interface{} {
				time.Sleep(150 * time.Millisecond)
				return fmt.Sprintf("完成: %s", data)
			},
		},
		{
			ID:       "high-2",
			Priority: 10,
			Data:     "高优先级任务2",
			Process: func(data interface{}) interface{} {
				time.Sleep(100 * time.Millisecond)
				return fmt.Sprintf("完成: %s", data)
			},
		},
		{
			ID:       "low-2",
			Priority: 1,
			Data:     "低优先级任务2",
			Process: func(data interface{}) interface{} {
				time.Sleep(200 * time.Millisecond)
				return fmt.Sprintf("完成: %s", data)
			},
		},
	}

	// 提交任务
	fmt.Println("提交任务（注意处理顺序应该按优先级排序）:")
	for _, job := range jobs {
		pool.Submit(job)
		fmt.Printf("提交任务 %s (优先级: %d)\n", job.GetID(), job.Priority)
	}

	// 收集结果
	fmt.Println("\n处理结果:")
	for i := 0; i < len(jobs); i++ {
		result := pool.GetResult()
		fmt.Printf("完成: %s -> %v\n", result.JobID, result.Value)
	}

	pool.Stop()
	fmt.Println()
}

// =============================================================================
// 5. 动态工作池
// =============================================================================

// DynamicWorkerPool 动态工作池
type DynamicWorkerPool struct {
	minWorkers     int
	maxWorkers     int
	currentWorkers int64
	jobQueue       chan Job
	resultQueue    chan Result
	scaleUpCh      chan bool
	scaleDownCh    chan bool
	quit           chan bool
	wg             sync.WaitGroup
	mu             sync.Mutex
}

// NewDynamicWorkerPool 创建动态工作池
func NewDynamicWorkerPool(minWorkers, maxWorkers, queueSize int) *DynamicWorkerPool {
	return &DynamicWorkerPool{
		minWorkers:     minWorkers,
		maxWorkers:     maxWorkers,
		currentWorkers: 0,
		jobQueue:       make(chan Job, queueSize),
		resultQueue:    make(chan Result, queueSize),
		scaleUpCh:      make(chan bool, maxWorkers),
		scaleDownCh:    make(chan bool, maxWorkers),
		quit:           make(chan bool),
	}
}

// Start 启动动态工作池
func (wp *DynamicWorkerPool) Start() {
	// 启动最小数量的工作者
	for i := 0; i < wp.minWorkers; i++ {
		wp.addWorker()
	}

	// 启动监控器
	go wp.monitor()

	fmt.Printf("动态工作池启动，初始工作者数量: %d\n", wp.minWorkers)
}

// addWorker 添加工作者
func (wp *DynamicWorkerPool) addWorker() {
	workerID := atomic.AddInt64(&wp.currentWorkers, 1)

	wp.wg.Add(1)
	go wp.worker(int(workerID))

	fmt.Printf("添加工作者 %d，当前工作者数量: %d\n", workerID, atomic.LoadInt64(&wp.currentWorkers))
}

// removeWorker 移除工作者
func (wp *DynamicWorkerPool) removeWorker() {
	current := atomic.LoadInt64(&wp.currentWorkers)
	if int(current) > wp.minWorkers {
		wp.scaleDownCh <- true
		fmt.Printf("请求移除工作者，当前工作者数量: %d\n", current)
	}
}

// worker 工作者
func (wp *DynamicWorkerPool) worker(id int) {
	defer wp.wg.Done()
	defer func() {
		atomic.AddInt64(&wp.currentWorkers, -1)
		fmt.Printf("工作者 %d 退出，当前工作者数量: %d\n", id, atomic.LoadInt64(&wp.currentWorkers))
	}()

	fmt.Printf("动态工作者 %d 开始工作\n", id)

	idleTimer := time.NewTimer(2 * time.Second)
	defer idleTimer.Stop()

	for {
		idleTimer.Reset(2 * time.Second)

		select {
		case job := <-wp.jobQueue:
			fmt.Printf("动态工作者 %d 处理任务 %s\n", id, job.GetID())

			start := time.Now()
			result := job.Execute()
			duration := time.Since(start)

			wp.resultQueue <- Result{
				JobID:  job.GetID(),
				Value:  result,
				Worker: id,
			}

			fmt.Printf("动态工作者 %d 完成任务 %s，耗时 %v\n", id, job.GetID(), duration)

		case <-wp.scaleDownCh:
			if int(atomic.LoadInt64(&wp.currentWorkers)) > wp.minWorkers {
				fmt.Printf("动态工作者 %d 收到缩容信号，准备退出\n", id)
				return
			}

		case <-idleTimer.C:
			// 空闲超时，如果工作者数量超过最小值则退出
			if int(atomic.LoadInt64(&wp.currentWorkers)) > wp.minWorkers {
				fmt.Printf("动态工作者 %d 空闲超时，自动退出\n", id)
				return
			}

		case <-wp.quit:
			fmt.Printf("动态工作者 %d 收到停止信号\n", id)
			return
		}
	}
}

// monitor 监控器
func (wp *DynamicWorkerPool) monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			queueLen := len(wp.jobQueue)
			currentWorkers := int(atomic.LoadInt64(&wp.currentWorkers))

			// 扩容逻辑：队列积压且工作者未达到最大值
			if queueLen > currentWorkers && currentWorkers < wp.maxWorkers {
				wp.addWorker()
			}

			// 缩容逻辑：队列空闲且工作者超过最小值
			if queueLen == 0 && currentWorkers > wp.minWorkers {
				wp.removeWorker()
			}

			fmt.Printf("监控: 队列长度=%d, 工作者数量=%d\n", queueLen, currentWorkers)

		case <-wp.quit:
			fmt.Println("动态工作池监控器停止")
			return
		}
	}
}

// Submit 提交任务
func (wp *DynamicWorkerPool) Submit(job Job) {
	wp.jobQueue <- job
}

// GetResult 获取结果
func (wp *DynamicWorkerPool) GetResult() Result {
	return <-wp.resultQueue
}

// GetWorkerCount 获取当前工作者数量
func (wp *DynamicWorkerPool) GetWorkerCount() int {
	return int(atomic.LoadInt64(&wp.currentWorkers))
}

// Stop 停止工作池
func (wp *DynamicWorkerPool) Stop() {
	close(wp.jobQueue)
	close(wp.quit)
	wp.wg.Wait()
	close(wp.resultQueue)
	fmt.Println("动态工作池已停止")
}

func demonstrateDynamicWorkerPool() {
	fmt.Println("=== 4. 动态工作池 ===")

	pool := NewDynamicWorkerPool(2, 6, 20)
	pool.Start()

	// 阶段1：少量任务
	fmt.Println("阶段1：提交少量任务")
	for i := 1; i <= 3; i++ {
		job := &SimpleJob{
			ID:   fmt.Sprintf("phase1-job-%d", i),
			Data: i,
		}
		pool.Submit(job)
	}

	time.Sleep(2 * time.Second)

	// 阶段2：大量任务
	fmt.Println("阶段2：提交大量任务（触发扩容）")
	for i := 1; i <= 15; i++ {
		job := &SimpleJob{
			ID:   fmt.Sprintf("phase2-job-%d", i),
			Data: i,
		}
		pool.Submit(job)
	}

	time.Sleep(3 * time.Second)

	// 阶段3：等待处理完成
	fmt.Println("阶段3：等待任务处理完成（观察缩容）")

	// 收集结果
	totalJobs := 18
	for i := 0; i < totalJobs; i++ {
		result := pool.GetResult()
		fmt.Printf("收到结果: %s -> %v (工作者 %d)\n",
			result.JobID, result.Value, result.Worker)
	}

	// 观察自动缩容
	time.Sleep(5 * time.Second)

	fmt.Printf("最终工作者数量: %d\n", pool.GetWorkerCount())
	pool.Stop()
	fmt.Println()
}

// =============================================================================
// 6. 工作池最佳实践
// =============================================================================

func demonstrateWorkerPoolBestPractices() {
	fmt.Println("=== 5. 工作池最佳实践 ===")

	fmt.Println("1. 工作池设计原则:")
	fmt.Println("   ✓ 根据任务类型选择合适的工作者数量")
	fmt.Println("   ✓ CPU密集型：工作者数 ≈ CPU核心数")
	fmt.Println("   ✓ I/O密集型：工作者数可以更多")
	fmt.Println("   ✓ 提供合理的队列缓冲大小")

	fmt.Println("\n2. 资源管理:")
	fmt.Println("   ✓ 确保正确关闭所有goroutine")
	fmt.Println("   ✓ 使用context进行取消控制")
	fmt.Println("   ✓ 实现优雅关闭机制")
	fmt.Println("   ✓ 监控队列长度和处理延迟")

	fmt.Println("\n3. 错误处理:")
	fmt.Println("   ✓ 任务级别的错误处理")
	fmt.Println("   ✓ 工作者崩溃的恢复机制")
	fmt.Println("   ✓ 超时处理")
	fmt.Println("   ✓ 重试策略")

	fmt.Println("\n4. 性能优化:")
	fmt.Println("   ✓ 避免频繁的内存分配")
	fmt.Println("   ✓ 使用对象池重用资源")
	fmt.Println("   ✓ 批处理相关任务")
	fmt.Println("   ✓ 监控和调优")

	fmt.Println("\n5. 常见陷阱:")
	fmt.Println("   ✗ 工作者数量过多导致上下文切换开销")
	fmt.Println("   ✗ 队列大小设置不当")
	fmt.Println("   ✗ 忘记处理工作者崩溃")
	fmt.Println("   ✗ 没有实现优雅关闭")

	// 性能对比示例
	fmt.Println("\n6. 性能对比示例:")

	const taskCount = 100

	// 串行处理
	start := time.Now()
	for i := 0; i < taskCount; i++ {
		// 模拟任务
		time.Sleep(10 * time.Millisecond)
	}
	serialTime := time.Since(start)

	// 工作池处理
	start = time.Now()
	pool := NewSimpleWorkerPool(4, 20)
	pool.Start()

	for i := 0; i < taskCount; i++ {
		job := &SimpleJob{
			ID:   fmt.Sprintf("perf-job-%d", i),
			Data: i,
		}
		pool.Submit(job)
	}

	for i := 0; i < taskCount; i++ {
		pool.GetResult()
	}

	pool.Stop()
	parallelTime := time.Since(start)

	fmt.Printf("串行处理时间: %v\n", serialTime)
	fmt.Printf("并行处理时间: %v\n", parallelTime)
	fmt.Printf("性能提升: %.2fx\n", float64(serialTime)/float64(parallelTime))

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - Worker Pool 工作池模式")
	fmt.Println("===================================")

	demonstrateSimpleWorkerPool()
	demonstrateContextWorkerPool()
	demonstratePriorityWorkerPool()
	demonstrateDynamicWorkerPool()
	demonstrateWorkerPoolBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个支持任务重试的工作池")
	fmt.Println("2. 创建一个多阶段处理的工作池")
	fmt.Println("3. 实现一个支持任务依赖的工作池")
	fmt.Println("4. 编写一个分布式工作池")
	fmt.Println("5. 创建一个支持限流的工作池")
	fmt.Println("6. 实现一个自适应的工作池")
	fmt.Println("\n请在此基础上练习更多工作池模式的使用！")
}
