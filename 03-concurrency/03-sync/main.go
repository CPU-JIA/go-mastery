package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"runtime"
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

func secureRandomFloat32() float32 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		// 安全fallback：使用时间戳
		return float32(time.Now().UnixNano()%1000) / 1000.0
	}
	return float32(n.Int64()) / float32(1<<24)
}

// =============================================================================
// 1. Sync 包基础概念
// =============================================================================

/*
sync 包提供了基本的同步原语，用于协调 goroutine 之间的执行：

主要组件：
1. Mutex：互斥锁，用于保护共享资源
2. RWMutex：读写锁，允许多个读者或一个写者
3. WaitGroup：等待组，等待一组 goroutine 完成
4. Once：确保某个操作只执行一次
5. Cond：条件变量，用于等待或通知条件变化
6. Pool：对象池，用于重用对象以减少GC压力
7. Map：并发安全的映射

原子操作（atomic 包）：
- 提供低级别的原子内存操作
- 比锁更高效，但只适用于简单操作
- 支持各种数值类型的原子操作

使用原则：
- 优先使用 channel 进行通信
- 在需要保护共享状态时使用同步原语
- 选择合适的同步原语以避免性能问题
- 避免死锁和竞态条件
*/

// =============================================================================
// 2. Mutex 互斥锁
// =============================================================================

// Counter 计数器（非并发安全）
type UnsafeCounter struct {
	count int
}

func (c *UnsafeCounter) Increment() {
	c.count++
}

func (c *UnsafeCounter) GetCount() int {
	return c.count
}

// SafeCounter 并发安全的计数器
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounter) GetCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

func demonstrateMutex() {
	fmt.Println("=== 1. Mutex 互斥锁 ===")

	const numGoroutines = 100
	const numIncrements = 1000

	// 测试非并发安全的计数器
	fmt.Println("测试非并发安全的计数器:")
	unsafeCounter := &UnsafeCounter{}

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIncrements; j++ {
				unsafeCounter.Increment()
			}
		}()
	}

	wg.Wait()
	unsafeDuration := time.Since(start)

	expectedCount := numGoroutines * numIncrements
	actualUnsafeCount := unsafeCounter.GetCount()

	fmt.Printf("期望计数: %d\n", expectedCount)
	fmt.Printf("实际计数: %d\n", actualUnsafeCount)
	fmt.Printf("数据丢失: %d\n", expectedCount-actualUnsafeCount)
	fmt.Printf("执行时间: %v\n", unsafeDuration)

	// 测试并发安全的计数器
	fmt.Println("\n测试并发安全的计数器:")
	safeCounter := &SafeCounter{}

	start = time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIncrements; j++ {
				safeCounter.Increment()
			}
		}()
	}

	wg.Wait()
	safeDuration := time.Since(start)

	actualSafeCount := safeCounter.GetCount()

	fmt.Printf("期望计数: %d\n", expectedCount)
	fmt.Printf("实际计数: %d\n", actualSafeCount)
	fmt.Printf("数据准确性: %s\n", map[bool]string{true: "✓ 正确", false: "✗ 错误"}[actualSafeCount == expectedCount])
	fmt.Printf("执行时间: %v\n", safeDuration)
	fmt.Printf("性能差异: %.2fx\n", float64(safeDuration)/float64(unsafeDuration))

	fmt.Println()
}

// =============================================================================
// 3. RWMutex 读写锁
// =============================================================================

// DataStore 使用读写锁的数据存储
type DataStore struct {
	mu   sync.RWMutex
	data map[string]int
}

// NewDataStore 创建数据存储
func NewDataStore() *DataStore {
	return &DataStore{
		data: make(map[string]int),
	}
}

// Set 设置值（写操作）
func (ds *DataStore) Set(key string, value int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	fmt.Printf("写入: %s = %d\n", key, value)
	ds.data[key] = value

	// 模拟一些写入工作
	time.Sleep(10 * time.Millisecond)
}

// Get 获取值（读操作）
func (ds *DataStore) Get(key string) (int, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	value, exists := ds.data[key]
	if exists {
		fmt.Printf("读取: %s = %d\n", key, value)
	} else {
		fmt.Printf("读取: %s 不存在\n", key)
	}

	// 模拟一些读取工作
	time.Sleep(5 * time.Millisecond)

	return value, exists
}

// GetAll 获取所有值（读操作）
func (ds *DataStore) GetAll() map[string]int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	// 复制数据以避免外部修改
	result := make(map[string]int)
	for k, v := range ds.data {
		result[k] = v
	}

	fmt.Printf("读取所有数据: %d 个条目\n", len(result))
	return result
}

func demonstrateRWMutex() {
	fmt.Println("=== 2. RWMutex 读写锁 ===")

	store := NewDataStore()
	var wg sync.WaitGroup

	// 启动写入者
	numWriters := 2
	for i := 1; i <= numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()

			for j := 1; j <= 5; j++ {
				key := fmt.Sprintf("key_%d_%d", writerID, j)
				value := writerID*100 + j
				store.Set(key, value)
			}
		}(i)
	}

	// 启动读取者
	numReaders := 5
	for i := 1; i <= numReaders; i++ {
		wg.Add(1)
		go func(readerID int) {
			defer wg.Done()

			for j := 1; j <= 3; j++ {
				// 随机读取一些键
				key := fmt.Sprintf("key_%d_%d", secureRandomInt(numWriters)+1, secureRandomInt(5)+1)
				store.Get(key)

				// 偶尔获取所有数据
				if j%2 == 0 {
					store.GetAll()
				}

				time.Sleep(20 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// 最终状态
	fmt.Println("\n最终数据状态:")
	finalData := store.GetAll()
	for k, v := range finalData {
		fmt.Printf("  %s: %d\n", k, v)
	}

	fmt.Println()
}

// =============================================================================
// 4. Once 单次执行
// =============================================================================

// Config 配置单例
type Config struct {
	DatabaseURL string
	APIKey      string
	Debug       bool
}

var (
	configInstance *Config
	configOnce     sync.Once
)

// GetConfig 获取配置实例（单例模式）
func GetConfig() *Config {
	configOnce.Do(func() {
		fmt.Println("初始化配置（只执行一次）...")

		// 模拟从配置文件或环境变量加载配置
		time.Sleep(100 * time.Millisecond)

		configInstance = &Config{
			DatabaseURL: "postgres://localhost/mydb",
			APIKey:      "secret-api-key-12345",
			Debug:       true,
		}

		fmt.Println("配置初始化完成")
	})

	return configInstance
}

// ExpensiveResource 昂贵的资源初始化
type ExpensiveResource struct {
	data []int
}

var (
	resource     *ExpensiveResource
	resourceOnce sync.Once
)

// GetResource 获取资源（懒加载）
func GetResource() *ExpensiveResource {
	resourceOnce.Do(func() {
		fmt.Println("初始化昂贵资源（只执行一次）...")

		// 模拟昂贵的初始化过程
		time.Sleep(200 * time.Millisecond)

		resource = &ExpensiveResource{
			data: make([]int, 1000000), // 大数组
		}

		// 填充一些数据
		for i := range resource.data {
			resource.data[i] = i
		}

		fmt.Println("昂贵资源初始化完成")
	})

	return resource
}

func demonstrateOnce() {
	fmt.Println("=== 3. Once 单次执行 ===")

	var wg sync.WaitGroup

	// 多个 goroutine 同时尝试获取配置
	fmt.Println("多个 goroutine 同时获取配置:")
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fmt.Printf("Goroutine %d 请求配置\n", id)
			config := GetConfig()
			fmt.Printf("Goroutine %d 获得配置: %s\n", id, config.DatabaseURL)
		}(i)
	}

	wg.Wait()

	// 多个 goroutine 同时尝试获取资源
	fmt.Println("\n多个 goroutine 同时获取资源:")
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fmt.Printf("Goroutine %d 请求资源\n", id)
			res := GetResource()
			fmt.Printf("Goroutine %d 获得资源，数据长度: %d\n", id, len(res.data))
		}(i)
	}

	wg.Wait()
	fmt.Println()
}

// =============================================================================
// 5. Cond 条件变量
// =============================================================================

// Buffer 有界缓冲区
type Buffer struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	items    []int
	capacity int
}

// NewBuffer 创建缓冲区
func NewBuffer(capacity int) *Buffer {
	b := &Buffer{
		items:    make([]int, 0, capacity),
		capacity: capacity,
	}
	b.notEmpty = sync.NewCond(&b.mu)
	b.notFull = sync.NewCond(&b.mu)
	return b
}

// Put 放入数据
func (b *Buffer) Put(item int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 等待缓冲区不满
	for len(b.items) == b.capacity {
		fmt.Printf("缓冲区满了，等待消费者消费...\n")
		b.notFull.Wait()
	}

	b.items = append(b.items, item)
	fmt.Printf("生产者放入: %d，缓冲区大小: %d\n", item, len(b.items))

	// 通知等待的消费者
	b.notEmpty.Signal()
}

// Get 获取数据
func (b *Buffer) Get() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 等待缓冲区不空
	for len(b.items) == 0 {
		fmt.Printf("缓冲区空了，等待生产者生产...\n")
		b.notEmpty.Wait()
	}

	item := b.items[0]
	b.items = b.items[1:]
	fmt.Printf("消费者取出: %d，缓冲区大小: %d\n", item, len(b.items))

	// 通知等待的生产者
	b.notFull.Signal()

	return item
}

func demonstrateCond() {
	fmt.Println("=== 4. Cond 条件变量 ===")

	buffer := NewBuffer(3)
	var wg sync.WaitGroup

	// 启动生产者
	wg.Add(1)
	go func() {
		defer wg.Done()

		for i := 1; i <= 10; i++ {
			buffer.Put(i)
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println("生产者完成")
	}()

	// 启动多个消费者
	for i := 1; i <= 2; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()

			for j := 0; j < 5; j++ {
				item := buffer.Get()
				fmt.Printf("消费者 %d 处理项目: %d\n", consumerID, item)
				time.Sleep(150 * time.Millisecond)
			}
			fmt.Printf("消费者 %d 完成\n", consumerID)
		}(i)
	}

	wg.Wait()
	fmt.Println()
}

// =============================================================================
// 6. Pool 对象池
// =============================================================================

// WorkItem 工作项
type WorkItem struct {
	ID   int
	Data []byte
}

// Reset 重置工作项
func (w *WorkItem) Reset() {
	w.ID = 0
	w.Data = w.Data[:0]
}

// 全局工作项池
var workItemPool = sync.Pool{
	New: func() interface{} {
		fmt.Println("创建新的 WorkItem")
		return &WorkItem{
			Data: make([]byte, 0, 1024), // 预分配容量
		}
	},
}

// getWorkItem 从池中获取工作项
func getWorkItem() *WorkItem {
	return workItemPool.Get().(*WorkItem)
}

// putWorkItem 将工作项放回池中
func putWorkItem(item *WorkItem) {
	item.Reset()
	workItemPool.Put(item)
}

// processWorkItem 处理工作项
func processWorkItem(id int) {
	// 从池中获取工作项
	item := getWorkItem()
	defer putWorkItem(item) // 确保归还到池中

	// 设置工作项数据
	item.ID = id
	item.Data = append(item.Data, fmt.Sprintf("WorkItem-%d", id)...)

	fmt.Printf("处理工作项 %d，数据: %s\n", item.ID, string(item.Data))

	// 模拟处理时间
	time.Sleep(50 * time.Millisecond)
}

func demonstratePool() {
	fmt.Println("=== 5. Pool 对象池 ===")

	var wg sync.WaitGroup

	// 启动多个 goroutine 处理工作项
	fmt.Println("使用对象池处理工作项:")
	start := time.Now()

	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processWorkItem(id)
		}(i)
	}

	wg.Wait()
	poolDuration := time.Since(start)

	fmt.Printf("使用对象池处理时间: %v\n", poolDuration)

	// 对比不使用对象池的情况
	fmt.Println("\n不使用对象池处理工作项:")
	start = time.Now()

	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 每次都创建新的工作项
			item := &WorkItem{
				ID:   id,
				Data: make([]byte, 0, 1024),
			}
			item.Data = append(item.Data, fmt.Sprintf("WorkItem-%d", id)...)

			fmt.Printf("处理工作项 %d，数据: %s\n", item.ID, string(item.Data))
			time.Sleep(50 * time.Millisecond)
		}(i)
	}

	wg.Wait()
	directDuration := time.Since(start)

	fmt.Printf("不使用对象池处理时间: %v\n", directDuration)
	fmt.Printf("性能差异: %.2fx\n", float64(directDuration)/float64(poolDuration))

	fmt.Println()
}

// =============================================================================
// 7. Map 并发安全的映射
// =============================================================================

func demonstrateSyncMap() {
	fmt.Println("=== 6. Map 并发安全的映射 ===")

	var syncMap sync.Map
	var regularMap = make(map[string]int)
	var regularMapMu sync.RWMutex

	var wg sync.WaitGroup

	// 测试 sync.Map
	fmt.Println("测试 sync.Map:")
	start := time.Now()

	// 写入操作
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)
			value := id * 10

			syncMap.Store(key, value)
			fmt.Printf("sync.Map 存储: %s = %d\n", key, value)
		}(i)
	}

	// 读取操作
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)
			if value, ok := syncMap.Load(key); ok {
				fmt.Printf("sync.Map 读取: %s = %d\n", key, value)
			}
		}(i)
	}

	wg.Wait()
	syncMapDuration := time.Since(start)

	// 统计 sync.Map 中的条目数
	syncMapCount := 0
	syncMap.Range(func(key, value interface{}) bool {
		syncMapCount++
		return true
	})

	fmt.Printf("sync.Map 执行时间: %v，条目数: %d\n", syncMapDuration, syncMapCount)

	// 测试带锁的普通 map
	fmt.Println("\n测试带锁的普通 map:")
	start = time.Now()

	// 写入操作
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)
			value := id * 10

			regularMapMu.Lock()
			regularMap[key] = value
			regularMapMu.Unlock()

			fmt.Printf("普通 map 存储: %s = %d\n", key, value)
		}(i)
	}

	// 读取操作
	for i := 1; i <= 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", id)

			regularMapMu.RLock()
			if value, ok := regularMap[key]; ok {
				regularMapMu.RUnlock()
				fmt.Printf("普通 map 读取: %s = %d\n", key, value)
			} else {
				regularMapMu.RUnlock()
			}
		}(i)
	}

	wg.Wait()
	regularMapDuration := time.Since(start)

	fmt.Printf("普通 map 执行时间: %v，条目数: %d\n", regularMapDuration, len(regularMap))
	fmt.Printf("性能比较: %.2fx\n", float64(regularMapDuration)/float64(syncMapDuration))

	fmt.Println()
}

// =============================================================================
// 8. Atomic 原子操作
// =============================================================================

// AtomicCounter 使用原子操作的计数器
type AtomicCounter struct {
	count int64
}

// Increment 原子递增
func (c *AtomicCounter) Increment() {
	atomic.AddInt64(&c.count, 1)
}

// Decrement 原子递减
func (c *AtomicCounter) Decrement() {
	atomic.AddInt64(&c.count, -1)
}

// GetCount 原子读取
func (c *AtomicCounter) GetCount() int64 {
	return atomic.LoadInt64(&c.count)
}

// SetCount 原子设置
func (c *AtomicCounter) SetCount(value int64) {
	atomic.StoreInt64(&c.count, value)
}

// CompareAndSwap 比较并交换
func (c *AtomicCounter) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&c.count, old, new)
}

func demonstrateAtomic() {
	fmt.Println("=== 7. Atomic 原子操作 ===")

	const numGoroutines = 100
	const numOperations = 1000

	// 测试原子操作性能
	fmt.Println("测试原子操作:")
	atomicCounter := &AtomicCounter{}

	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				atomicCounter.Increment()
			}
		}()
	}

	wg.Wait()
	atomicDuration := time.Since(start)

	expectedCount := int64(numGoroutines * numOperations)
	actualAtomicCount := atomicCounter.GetCount()

	fmt.Printf("期望计数: %d\n", expectedCount)
	fmt.Printf("实际计数: %d\n", actualAtomicCount)
	fmt.Printf("数据准确性: %s\n", map[bool]string{true: "✓ 正确", false: "✗ 错误"}[actualAtomicCount == expectedCount])
	fmt.Printf("执行时间: %v\n", atomicDuration)

	// 对比 Mutex 的性能
	fmt.Println("\n对比 Mutex 性能:")
	mutexCounter := &SafeCounter{}

	start = time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				mutexCounter.Increment()
			}
		}()
	}

	wg.Wait()
	mutexDuration := time.Since(start)

	actualMutexCount := mutexCounter.GetCount()

	fmt.Printf("期望计数: %d\n", expectedCount)
	fmt.Printf("实际计数: %d\n", actualMutexCount)
	fmt.Printf("数据准确性: %s\n", map[bool]string{true: "✓ 正确", false: "✗ 错误"}[int64(actualMutexCount) == expectedCount])
	fmt.Printf("执行时间: %v\n", mutexDuration)
	fmt.Printf("原子操作性能优势: %.2fx\n", float64(mutexDuration)/float64(atomicDuration))

	// 演示其他原子操作
	fmt.Println("\n其他原子操作:")
	atomicCounter.SetCount(100)
	fmt.Printf("设置计数为: %d\n", atomicCounter.GetCount())

	// 比较并交换
	if atomicCounter.CompareAndSwap(100, 200) {
		fmt.Printf("成功将 100 交换为 200，当前值: %d\n", atomicCounter.GetCount())
	}

	if !atomicCounter.CompareAndSwap(100, 300) {
		fmt.Printf("未能将 100 交换为 300（当前值不是100），当前值: %d\n", atomicCounter.GetCount())
	}

	fmt.Println()
}

// =============================================================================
// 9. 死锁检测和避免
// =============================================================================

// Account 银行账户
type Account struct {
	mu      sync.Mutex
	id      int
	balance int64
}

// Transfer 转账（可能导致死锁的版本）
func (from *Account) TransferDeadlockProne(to *Account, amount int64) error {
	from.mu.Lock()
	defer from.mu.Unlock()

	to.mu.Lock()
	defer to.mu.Unlock()

	if from.balance < amount {
		return fmt.Errorf("余额不足")
	}

	from.balance -= amount
	to.balance += amount

	fmt.Printf("转账成功: 账户%d -> 账户%d, 金额: %d\n", from.id, to.id, amount)
	return nil
}

// SafeTransfer 安全转账（避免死锁）
func SafeTransfer(from, to *Account, amount int64) error {
	// 按照ID顺序获取锁，避免死锁
	if from.id < to.id {
		from.mu.Lock()
		defer from.mu.Unlock()
		to.mu.Lock()
		defer to.mu.Unlock()
	} else {
		to.mu.Lock()
		defer to.mu.Unlock()
		from.mu.Lock()
		defer from.mu.Unlock()
	}

	if from.balance < amount {
		return fmt.Errorf("余额不足")
	}

	from.balance -= amount
	to.balance += amount

	fmt.Printf("安全转账成功: 账户%d -> 账户%d, 金额: %d\n", from.id, to.id, amount)
	return nil
}

func demonstrateDeadlockPrevention() {
	fmt.Println("=== 8. 死锁检测和避免 ===")

	account1 := &Account{id: 1, balance: 1000}
	account2 := &Account{id: 2, balance: 1000}

	var wg sync.WaitGroup

	fmt.Println("演示安全转账（避免死锁）:")

	// 启动多个并发转账操作
	for i := 0; i < 5; i++ {
		wg.Add(2)

		go func(round int) {
			defer wg.Done()
			err := SafeTransfer(account1, account2, 50)
			if err != nil {
				fmt.Printf("转账失败（轮次%d）: %v\n", round, err)
			}
		}(i)

		go func(round int) {
			defer wg.Done()
			err := SafeTransfer(account2, account1, 30)
			if err != nil {
				fmt.Printf("转账失败（轮次%d）: %v\n", round, err)
			}
		}(i)

		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	fmt.Printf("最终余额 - 账户1: %d, 账户2: %d\n", account1.balance, account2.balance)

	fmt.Println("\n死锁预防最佳实践:")
	fmt.Println("1. 固定的锁获取顺序")
	fmt.Println("2. 使用超时机制")
	fmt.Println("3. 避免嵌套锁")
	fmt.Println("4. 使用 defer 确保锁被释放")
	fmt.Println("5. 优先使用 channel 而不是共享状态")

	fmt.Println()
}

// =============================================================================
// 10. 性能对比和最佳实践
// =============================================================================

func demonstratePerformanceComparison() {
	fmt.Println("=== 9. 性能对比和最佳实践 ===")

	const iterations = 100000

	// 测试不同同步机制的性能
	fmt.Println("性能对比测试:")

	// 1. 原子操作
	var atomicValue int64
	start := time.Now()
	for i := 0; i < iterations; i++ {
		atomic.AddInt64(&atomicValue, 1)
	}
	atomicTime := time.Since(start)
	fmt.Printf("原子操作: %v\n", atomicTime)

	// 2. Mutex
	var mutexValue int64
	var mu sync.Mutex
	start = time.Now()
	for i := 0; i < iterations; i++ {
		mu.Lock()
		mutexValue++
		mu.Unlock()
	}
	mutexTime := time.Since(start)
	fmt.Printf("Mutex: %v\n", mutexTime)

	// 3. Channel
	ch := make(chan int64, 1)
	ch <- 0
	start = time.Now()
	for i := 0; i < iterations; i++ {
		val := <-ch
		ch <- val + 1
	}
	channelValue := <-ch
	channelTime := time.Since(start)
	fmt.Printf("Channel: %v\n", channelTime)

	fmt.Printf("\n性能排序（最快到最慢）:\n")
	fmt.Printf("1. 原子操作: %.2fx 基准\n", 1.0)
	fmt.Printf("2. Mutex: %.2fx 原子操作\n", float64(mutexTime)/float64(atomicTime))
	fmt.Printf("3. Channel: %.2fx 原子操作\n", float64(channelTime)/float64(atomicTime))

	fmt.Printf("\n最终值验证:\n")
	fmt.Printf("原子操作结果: %d\n", atomicValue)
	fmt.Printf("Mutex 结果: %d\n", mutexValue)
	fmt.Printf("Channel 结果: %d\n", channelValue)

	fmt.Println("\n最佳实践指南:")
	fmt.Println("1. 选择合适的同步机制:")
	fmt.Println("   - 简单计数：使用 atomic")
	fmt.Println("   - 复杂共享状态：使用 Mutex")
	fmt.Println("   - 数据传递：使用 Channel")

	fmt.Println("\n2. 性能考虑:")
	fmt.Println("   - 原子操作 > Mutex > Channel（在简单操作上）")
	fmt.Println("   - 读多写少的场景考虑 RWMutex")
	fmt.Println("   - 避免过细粒度的锁")

	fmt.Println("\n3. 避免常见问题:")
	fmt.Println("   - 死锁：固定锁顺序，使用超时")
	fmt.Println("   - 活锁：增加随机延迟")
	fmt.Println("   - 饥饿：公平锁，避免长时间持锁")

	fmt.Printf("\n当前运行时信息:\n")
	fmt.Printf("Goroutine 数量: %d\n", runtime.NumGoroutine())
	fmt.Printf("CPU 数量: %d\n", runtime.NumCPU())

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 并发编程 - Sync 同步原语")
	fmt.Println("============================")

	// 设置随机种子
	// 注意：crypto/rand不需要设置种子

	demonstrateMutex()
	demonstrateRWMutex()
	demonstrateOnce()
	demonstrateCond()
	demonstratePool()
	demonstrateSyncMap()
	demonstrateAtomic()
	demonstrateDeadlockPrevention()
	demonstratePerformanceComparison()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个读写分离的缓存系统")
	fmt.Println("2. 创建一个支持超时的分布式锁")
	fmt.Println("3. 实现一个线程安全的优先级队列")
	fmt.Println("4. 编写一个连接池管理器")
	fmt.Println("5. 创建一个支持限流的API网关")
	fmt.Println("6. 实现一个多生产者多消费者的消息队列")
	fmt.Println("\n请在此基础上练习更多同步机制的使用场景！")
}
