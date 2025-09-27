/*
=== Go性能掌控：深度优化技术 ===

本模块专注于Go语言深度性能优化技术的理论和实践，探索：
1. 算法复杂度分析和优化
2. 数据结构性能优化技术
3. 内存优化：零拷贝、对象池、内存布局
4. 并发优化：锁优化、无锁编程、协程调优
5. 编译器优化技术和内联优化
6. 汇编级性能调优
7. 缓存友好编程技术
8. SIMD向量化优化
9. 网络I/O性能优化
10. 数据库访问优化

学习目标：
- 掌握系统性的性能优化方法论
- 理解底层硬件对性能的影响
- 学会识别和消除性能瓶颈
- 掌握先进的优化技术和工具
*/

package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/bits"
	"net"
	"os"
	"runtime"
	"sort"
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

func secureRandomIntGeneric() int {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano()
		if fallback > int64(^uint(0)>>1) {
			fallback = fallback % int64(^uint(0)>>1)
		}
		return int(fallback)
	}
	// G115安全修复：检查int64到int的安全转换
	result := n.Int64()
	if result > int64(^uint(0)>>1) {
		result = result % int64(^uint(0)>>1)
	}
	return int(result)
}

// ==================
// 1. 算法优化技术
// ==================

// AlgorithmOptimizer 算法优化器
type AlgorithmOptimizer struct {
	benchmarkResults map[string]BenchmarkData
	mutex            sync.RWMutex
}

// BenchmarkData 基准测试数据
type BenchmarkData struct {
	Name        string
	Operations  int64
	NsPerOp     int64
	AllocsPerOp int64
	BytesPerOp  int64
	Complexity  string
	MemoryUsage int64
}

func NewAlgorithmOptimizer() *AlgorithmOptimizer {
	return &AlgorithmOptimizer{
		benchmarkResults: make(map[string]BenchmarkData),
	}
}

// ==================
// 1.1 排序算法优化
// ==================

// 演示不同排序算法的性能特征
func (ao *AlgorithmOptimizer) DemonstrateSortingOptimization() {
	fmt.Println("=== 排序算法优化演示 ===")

	sizes := []int{1000, 10000, 100000}

	for _, size := range sizes {
		fmt.Printf("\n数据大小: %d\n", size)

		// 生成测试数据
		data := generateRandomData(size)

		// 测试标准库排序
		ao.benchmarkSort("stdlib_sort", size, data, func(arr []int) {
			sort.Ints(arr)
		})

		// 测试快速排序
		ao.benchmarkSort("quicksort", size, data, func(arr []int) {
			quickSort(arr, 0, len(arr)-1)
		})

		// 测试归并排序
		ao.benchmarkSort("mergesort", size, data, func(arr []int) {
			mergeSort(arr)
		})

		// 测试基数排序（适用于整数）
		ao.benchmarkSort("radixsort", size, data, func(arr []int) {
			radixSort(arr)
		})

		// 测试并行排序
		ao.benchmarkSort("parallel_sort", size, data, func(arr []int) {
			parallelQuickSort(arr, 0, len(arr)-1, runtime.NumCPU())
		})
	}

	ao.printBenchmarkResults()
}

func (ao *AlgorithmOptimizer) benchmarkSort(name string, size int, originalData []int, sortFunc func([]int)) {
	// 复制数据避免影响其他测试
	data := make([]int, len(originalData))
	copy(data, originalData)

	// 预热
	for i := 0; i < 3; i++ {
		dataCopy := make([]int, len(data))
		copy(dataCopy, data)
		sortFunc(dataCopy)
	}

	// 测量内存使用
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// 执行基准测试
	iterations := 10
	start := time.Now()

	for i := 0; i < iterations; i++ {
		dataCopy := make([]int, len(data))
		copy(dataCopy, data)
		sortFunc(dataCopy)
	}

	elapsed := time.Since(start)
	runtime.ReadMemStats(&m2)

	// 记录结果
	result := BenchmarkData{
		Name:        fmt.Sprintf("%s_%d", name, size),
		Operations:  int64(iterations),
		NsPerOp:     elapsed.Nanoseconds() / int64(iterations),
		MemoryUsage: int64(m2.TotalAlloc - m1.TotalAlloc),
		Complexity:  getComplexity(name),
	}

	ao.mutex.Lock()
	ao.benchmarkResults[result.Name] = result
	ao.mutex.Unlock()

	fmt.Printf("  %s: %d ns/op, %d MB内存\n",
		name, result.NsPerOp, result.MemoryUsage/1024/1024)
}

func generateRandomData(size int) []int {
	data := make([]int, size)
	for i := range data {
		data[i] = secureRandomIntGeneric()
	}
	return data
}

func quickSort(arr []int, low, high int) {
	if low < high {
		pi := partition(arr, low, high)
		quickSort(arr, low, pi-1)
		quickSort(arr, pi+1, high)
	}
}

func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

func mergeSort(arr []int) {
	if len(arr) <= 1 {
		return
	}

	mid := len(arr) / 2
	left := make([]int, mid)
	right := make([]int, len(arr)-mid)

	copy(left, arr[:mid])
	copy(right, arr[mid:])

	mergeSort(left)
	mergeSort(right)

	merge(arr, left, right)
}

func merge(arr, left, right []int) {
	i, j, k := 0, 0, 0

	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			arr[k] = left[i]
			i++
		} else {
			arr[k] = right[j]
			j++
		}
		k++
	}

	for i < len(left) {
		arr[k] = left[i]
		i++
		k++
	}

	for j < len(right) {
		arr[k] = right[j]
		j++
		k++
	}
}

func radixSort(arr []int) {
	if len(arr) <= 1 {
		return
	}

	// 找到最大值确定位数
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}

	// 对每一位进行计数排序
	for exp := 1; max/exp > 0; exp *= 10 {
		countingSort(arr, exp)
	}
}

func countingSort(arr []int, exp int) {
	n := len(arr)
	output := make([]int, n)
	count := make([]int, 10)

	// 计算每个数字的出现次数
	for i := 0; i < n; i++ {
		count[(arr[i]/exp)%10]++
	}

	// 计算累积计数
	for i := 1; i < 10; i++ {
		count[i] += count[i-1]
	}

	// 构建输出数组
	for i := n - 1; i >= 0; i-- {
		digit := (arr[i] / exp) % 10
		output[count[digit]-1] = arr[i]
		count[digit]--
	}

	// 复制回原数组
	copy(arr, output)
}

func parallelQuickSort(arr []int, low, high, maxGoroutines int) {
	if low < high {
		if maxGoroutines <= 1 || high-low < 1000 {
			// 串行处理小数组或goroutine限制
			quickSort(arr, low, high)
			return
		}

		pi := partition(arr, low, high)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			parallelQuickSort(arr, low, pi-1, maxGoroutines/2)
		}()

		go func() {
			defer wg.Done()
			parallelQuickSort(arr, pi+1, high, maxGoroutines/2)
		}()

		wg.Wait()
	}
}

func getComplexity(name string) string {
	switch name {
	case "quicksort", "stdlib_sort":
		return "O(n log n) avg, O(n²) worst"
	case "mergesort":
		return "O(n log n)"
	case "radixsort":
		return "O(d * (n + k))"
	case "parallel_sort":
		return "O(n log n / p)"
	default:
		return "Unknown"
	}
}

// ==================
// 2. 内存优化技术
// ==================

// MemoryOptimizer 内存优化器
type MemoryOptimizer struct {
	objectPools map[string]*ObjectPool
	bufferPools map[string]*BufferPool
	metrics     MemoryMetrics
	mutex       sync.RWMutex
}

// ObjectPool 对象池
type ObjectPool struct {
	pool     sync.Pool
	creates  int64
	gets     int64
	puts     int64
	maxSize  int64
	itemSize int64
}

// BufferPool 缓冲区池
type BufferPool struct {
	pool    sync.Pool
	size    int
	creates int64
	gets    int64
	puts    int64
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	TotalAllocations int64
	TotalFrees       int64
	PoolHits         int64
	PoolMisses       int64
	BytesSaved       int64
}

func NewMemoryOptimizer() *MemoryOptimizer {
	return &MemoryOptimizer{
		objectPools: make(map[string]*ObjectPool),
		bufferPools: make(map[string]*BufferPool),
	}
}

func (mo *MemoryOptimizer) CreateObjectPool(name string, itemSize int64, newFunc func() interface{}) {
	pool := &ObjectPool{
		itemSize: itemSize,
	}

	pool.pool = sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&pool.creates, 1)
			return newFunc()
		},
	}

	mo.mutex.Lock()
	mo.objectPools[name] = pool
	mo.mutex.Unlock()
}

func (mo *MemoryOptimizer) CreateBufferPool(name string, size int) {
	pool := &BufferPool{
		size: size,
	}

	pool.pool = sync.Pool{
		New: func() interface{} {
			atomic.AddInt64(&pool.creates, 1)
			return make([]byte, size)
		},
	}

	mo.mutex.Lock()
	mo.bufferPools[name] = pool
	mo.mutex.Unlock()
}

func (mo *MemoryOptimizer) GetObject(poolName string) interface{} {
	mo.mutex.RLock()
	pool, exists := mo.objectPools[poolName]
	mo.mutex.RUnlock()

	if !exists {
		return nil
	}

	atomic.AddInt64(&pool.gets, 1)
	atomic.AddInt64(&mo.metrics.PoolHits, 1)
	return pool.pool.Get()
}

func (mo *MemoryOptimizer) PutObject(poolName string, obj interface{}) {
	mo.mutex.RLock()
	pool, exists := mo.objectPools[poolName]
	mo.mutex.RUnlock()

	if !exists {
		return
	}

	atomic.AddInt64(&pool.puts, 1)
	atomic.AddInt64(&mo.metrics.BytesSaved, pool.itemSize)
	pool.pool.Put(obj)
}

func (mo *MemoryOptimizer) GetBuffer(poolName string) []byte {
	mo.mutex.RLock()
	pool, exists := mo.bufferPools[poolName]
	mo.mutex.RUnlock()

	if !exists {
		return nil
	}

	atomic.AddInt64(&pool.gets, 1)
	return pool.pool.Get().([]byte)
}

func (mo *MemoryOptimizer) PutBuffer(poolName string, buf []byte) {
	mo.mutex.RLock()
	pool, exists := mo.bufferPools[poolName]
	mo.mutex.RUnlock()

	if !exists {
		return
	}

	// 重置缓冲区
	if len(buf) == pool.size {
		for i := range buf {
			buf[i] = 0
		}
		atomic.AddInt64(&pool.puts, 1)
		pool.pool.Put(buf)
	}
}

// ==================
// 2.1 零拷贝技术演示
// ==================

// ZeroCopyDemo 零拷贝演示
func (mo *MemoryOptimizer) DemonstrateZeroCopy() {
	fmt.Println("\n=== 零拷贝技术演示 ===")

	// 创建测试数据
	sourceData := make([]byte, 1024*1024) // 1MB
	for i := range sourceData {
		sourceData[i] = byte(i % 256)
	}

	// 传统拷贝方式
	fmt.Println("1. 传统拷贝方式:")
	start := time.Now()
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		destData := make([]byte, len(sourceData))
		copy(destData, sourceData)
		_ = destData
	}

	runtime.ReadMemStats(&m2)
	elapsed1 := time.Since(start)
	allocBytes1 := m2.TotalAlloc - m1.TotalAlloc

	fmt.Printf("  时间: %v, 分配内存: %d MB\n",
		elapsed1, allocBytes1/1024/1024)

	// 使用字节切片的零拷贝方式
	fmt.Println("2. 切片零拷贝方式:")
	start = time.Now()
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		// 使用切片引用，避免内存拷贝
		destSlice := sourceData[:]
		_ = destSlice
	}

	runtime.ReadMemStats(&m2)
	elapsed2 := time.Since(start)
	allocBytes2 := m2.TotalAlloc - m1.TotalAlloc

	fmt.Printf("  时间: %v, 分配内存: %d KB\n",
		elapsed2, allocBytes2/1024)

	// 使用内存映射的零拷贝
	fmt.Println("3. 内存映射零拷贝:")
	mo.demonstrateMemoryMapping(sourceData)

	// 性能对比
	fmt.Printf("\n性能提升:\n")
	fmt.Printf("  时间提升: %.2fx\n", float64(elapsed1.Nanoseconds())/float64(elapsed2.Nanoseconds()))
	fmt.Printf("  内存节省: %.2fx\n", float64(allocBytes1)/float64(max(allocBytes2, 1)))
}

func (mo *MemoryOptimizer) demonstrateMemoryMapping(data []byte) {
	// 创建临时文件
	file, err := os.CreateTemp("", "mmap_test")
	if err != nil {
		fmt.Printf("  创建临时文件失败: %v\n", err)
		return
	}
	defer os.Remove(file.Name())
	defer file.Close()

	// 写入数据
	if _, err := file.Write(data); err != nil {
		fmt.Printf("  写入文件失败: %v\n", err)
		return
	}

	start := time.Now()

	// 模拟内存映射读取（简化版本）
	for i := 0; i < 100; i++ {
		if _, err := file.Seek(0, 0); err != nil {
			log.Printf("重置文件指针失败: %v", err)
			continue
		}

		// 读取文件头部分（模拟映射访问）
		buffer := make([]byte, 4096)
		if _, err := file.Read(buffer); err != nil {
			log.Printf("读取文件失败: %v", err)
			continue
		}
		_ = buffer
	}

	elapsed := time.Since(start)
	fmt.Printf("  时间: %v (内存映射模拟)\n", elapsed)
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// ==================
// 3. 并发优化技术
// ==================

// ConcurrencyOptimizer 并发优化器
type ConcurrencyOptimizer struct {
	lockContention   map[string]*LockMetrics
	goroutineMetrics *GoroutineMetrics
	mutex            sync.RWMutex
}

// LockMetrics 锁指标
type LockMetrics struct {
	Name            string
	AcquisitionTime time.Duration
	HoldTime        time.Duration
	Contentions     int64
	Acquisitions    int64
}

// GoroutineMetrics Goroutine指标
type GoroutineMetrics struct {
	Created   int64
	Destroyed int64
	MaxActive int64
	Current   int64
}

func NewConcurrencyOptimizer() *ConcurrencyOptimizer {
	return &ConcurrencyOptimizer{
		lockContention:   make(map[string]*LockMetrics),
		goroutineMetrics: &GoroutineMetrics{},
	}
}

// ==================
// 3.1 锁优化演示
// ==================

func (co *ConcurrencyOptimizer) DemonstrateLockOptimization() {
	fmt.Println("\n=== 锁优化技术演示 ===")

	dataSize := 1000000
	goroutines := runtime.NumCPU()

	// 1. 互斥锁基准测试
	fmt.Println("1. 互斥锁性能测试:")
	co.benchmarkMutexLock(dataSize, goroutines)

	// 2. 读写锁测试
	fmt.Println("2. 读写锁性能测试:")
	co.benchmarkRWLock(dataSize, goroutines)

	// 3. 原子操作测试
	fmt.Println("3. 原子操作性能测试:")
	co.benchmarkAtomicOperations(dataSize, goroutines)

	// 4. 无锁数据结构测试
	fmt.Println("4. 无锁队列性能测试:")
	co.benchmarkLockFreeQueue(dataSize, goroutines)

	// 5. 分片锁测试
	fmt.Println("5. 分片锁性能测试:")
	co.benchmarkShardedLock(dataSize, goroutines)
}

func (co *ConcurrencyOptimizer) benchmarkMutexLock(operations, goroutines int) {
	var mu sync.Mutex
	var counter int64

	start := time.Now()
	var wg sync.WaitGroup

	opsPerGoroutine := operations / goroutines

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  互斥锁: %v, %d ops/sec, 最终值: %d\n",
		elapsed, int64(float64(operations)/elapsed.Seconds()), counter)
}

func (co *ConcurrencyOptimizer) benchmarkRWLock(operations, goroutines int) {
	var rwmu sync.RWMutex
	var counter int64

	start := time.Now()
	var wg sync.WaitGroup

	// 90% 读操作，10% 写操作
	readOps := operations * 9 / 10
	writeOps := operations / 10

	// 读goroutines
	readOpsPerGoroutine := readOps / goroutines
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < readOpsPerGoroutine; j++ {
				rwmu.RLock()
				_ = counter
				rwmu.RUnlock()
			}
		}()
	}

	// 写goroutines
	writeOpsPerGoroutine := writeOps / goroutines
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < writeOpsPerGoroutine; j++ {
				rwmu.Lock()
				counter++
				rwmu.Unlock()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  读写锁: %v, %d ops/sec, 最终值: %d\n",
		elapsed, int64(float64(operations)/elapsed.Seconds()), counter)
}

func (co *ConcurrencyOptimizer) benchmarkAtomicOperations(operations, goroutines int) {
	var counter int64

	start := time.Now()
	var wg sync.WaitGroup

	opsPerGoroutine := operations / goroutines

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  原子操作: %v, %d ops/sec, 最终值: %d\n",
		elapsed, int64(float64(operations)/elapsed.Seconds()), atomic.LoadInt64(&counter))
}

// 简化的无锁队列实现
type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

type node struct {
	data unsafe.Pointer
	next unsafe.Pointer
}

func newLockFreeQueue() *LockFreeQueue {
	n := &node{}
	q := &LockFreeQueue{
		head: unsafe.Pointer(n),
		tail: unsafe.Pointer(n),
	}
	return q
}

func (q *LockFreeQueue) enqueue(data interface{}) {
	n := &node{data: unsafe.Pointer(&data)}

	for {
		tail := (*node)(atomic.LoadPointer(&q.tail))
		next := (*node)(atomic.LoadPointer(&tail.next))

		if tail == (*node)(atomic.LoadPointer(&q.tail)) {
			if next == nil {
				if atomic.CompareAndSwapPointer(&tail.next, unsafe.Pointer(next), unsafe.Pointer(n)) {
					break
				}
			} else {
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			}
		}
	}
	atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer((*node)(atomic.LoadPointer(&q.tail))), unsafe.Pointer(n))
}

func (q *LockFreeQueue) dequeue() interface{} {
	for {
		head := (*node)(atomic.LoadPointer(&q.head))
		tail := (*node)(atomic.LoadPointer(&q.tail))
		next := (*node)(atomic.LoadPointer(&head.next))

		if head == (*node)(atomic.LoadPointer(&q.head)) {
			if head == tail {
				if next == nil {
					return nil
				}
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			} else {
				if next == nil {
					continue
				}
				data := atomic.LoadPointer(&next.data)
				if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
					return *(*interface{})(data)
				}
			}
		}
	}
}

func (co *ConcurrencyOptimizer) benchmarkLockFreeQueue(operations, goroutines int) {
	queue := newLockFreeQueue()

	start := time.Now()
	var wg sync.WaitGroup

	opsPerGoroutine := operations / goroutines

	// 生产者
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine*2; j++ {
				queue.enqueue(id*opsPerGoroutine + j)
			}
		}(i)
	}

	// 消费者
	consumed := int64(0)
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine*2; j++ {
				for queue.dequeue() == nil {
					runtime.Gosched()
				}
				atomic.AddInt64(&consumed, 1)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  无锁队列: %v, %d ops/sec, 消费数量: %d\n",
		elapsed, int64(float64(operations)/elapsed.Seconds()), consumed)
}

// 分片锁实现
type ShardedMap struct {
	shards []shard
	mask   uint64
}

type shard struct {
	mu   sync.RWMutex
	data map[string]int
}

func newShardedMap(shardCount int) *ShardedMap {
	// 确保是2的幂
	if bits.OnesCount(uint(shardCount)) != 1 {
		shardCount = 1 << bits.Len(uint(shardCount-1))
	}

	shards := make([]shard, shardCount)
	for i := range shards {
		shards[i].data = make(map[string]int)
	}

	return &ShardedMap{
		shards: shards,
		mask:   uint64(shardCount - 1),
	}
}

func (sm *ShardedMap) getShard(key string) *shard {
	hash := fnv1a(key)
	return &sm.shards[hash&sm.mask]
}

func (sm *ShardedMap) Set(key string, value int) {
	shard := sm.getShard(key)
	shard.mu.Lock()
	shard.data[key] = value
	shard.mu.Unlock()
}

func (sm *ShardedMap) Get(key string) (int, bool) {
	shard := sm.getShard(key)
	shard.mu.RLock()
	value, exists := shard.data[key]
	shard.mu.RUnlock()
	return value, exists
}

func fnv1a(s string) uint64 {
	const prime = 1099511628211
	hash := uint64(14695981039346656037)
	for _, b := range []byte(s) {
		hash ^= uint64(b)
		hash *= prime
	}
	return hash
}

func (co *ConcurrencyOptimizer) benchmarkShardedLock(operations, goroutines int) {
	shardedMap := newShardedMap(32) // 32个分片

	start := time.Now()
	var wg sync.WaitGroup

	opsPerGoroutine := operations / goroutines

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				shardedMap.Set(key, j)
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Printf("  分片锁: %v, %d ops/sec\n",
		elapsed, int64(float64(operations)/elapsed.Seconds()))
}

// ==================
// 4. 缓存优化技术
// ==================

// CacheOptimizer 缓存优化器
type CacheOptimizer struct {
	cacheStats CacheStats
}

// CacheStats 缓存统计
type CacheStats struct {
	L1Misses     int64
	L2Misses     int64
	L3Misses     int64
	TLBMisses    int64
	BranchMisses int64
}

func NewCacheOptimizer() *CacheOptimizer {
	return &CacheOptimizer{}
}

// ==================
// 4.1 数据局部性优化
// ==================

func (co *CacheOptimizer) DemonstrateDataLocality() {
	fmt.Println("\n=== 数据局部性优化演示 ===")

	size := 1024 * 1024 // 1M元素

	// 1. 连续访问模式（缓存友好）
	fmt.Println("1. 连续访问模式:")
	co.benchmarkSequentialAccess(size)

	// 2. 随机访问模式（缓存不友好）
	fmt.Println("2. 随机访问模式:")
	co.benchmarkRandomAccess(size)

	// 3. 分块访问模式（优化的随机访问）
	fmt.Println("3. 分块访问模式:")
	co.benchmarkBlockedAccess(size)

	// 4. 数据结构布局优化
	fmt.Println("4. 数据结构布局优化:")
	co.benchmarkStructLayout()
}

func (co *CacheOptimizer) benchmarkSequentialAccess(size int) {
	data := make([]int64, size)

	start := time.Now()
	sum := int64(0)

	for i := 0; i < size; i++ {
		sum += data[i]
	}

	elapsed := time.Since(start)
	fmt.Printf("  连续访问: %v, sum=%d\n", elapsed, sum)
}

func (co *CacheOptimizer) benchmarkRandomAccess(size int) {
	data := make([]int64, size)
	indices := make([]int, size)

	// 生成随机索引
	for i := range indices {
		indices[i] = secureRandomInt(size)
	}

	start := time.Now()
	sum := int64(0)

	for _, idx := range indices {
		sum += data[idx]
	}

	elapsed := time.Since(start)
	fmt.Printf("  随机访问: %v, sum=%d\n", elapsed, sum)
}

func (co *CacheOptimizer) benchmarkBlockedAccess(size int) {
	data := make([]int64, size)
	blockSize := 64 // 缓存行大小通常是64字节

	start := time.Now()
	sum := int64(0)

	// 分块访问，提高缓存局部性
	for block := 0; block < size; block += blockSize {
		end := block + blockSize
		if end > size {
			end = size
		}
		for i := block; i < end; i++ {
			sum += data[i]
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("  分块访问: %v, sum=%d\n", elapsed, sum)
}

// 演示结构体布局对性能的影响
type BadStructLayout struct {
	a bool  // 1字节
	b int64 // 8字节 - 需要7字节填充
	c bool  // 1字节
	d int32 // 4字节 - 需要3字节填充
	e bool  // 1字节
	f int64 // 8字节 - 需要7字节填充
}

type GoodStructLayout struct {
	b int64 // 8字节
	f int64 // 8字节
	d int32 // 4字节
	a bool  // 1字节
	c bool  // 1字节
	e bool  // 1字节
	// 1字节填充到对齐边界
}

func (co *CacheOptimizer) benchmarkStructLayout() {
	count := 1000000

	// 测试布局差的结构体
	badStructs := make([]BadStructLayout, count)
	start := time.Now()

	for i := range badStructs {
		badStructs[i].a = true
		badStructs[i].b = int64(i)
		badStructs[i].c = true
		badStructs[i].d = int32(i)
		badStructs[i].e = true
		badStructs[i].f = int64(i * 2)
	}

	badElapsed := time.Since(start)
	badSize := unsafe.Sizeof(BadStructLayout{})

	// 测试布局好的结构体
	goodStructs := make([]GoodStructLayout, count)
	start = time.Now()

	for i := range goodStructs {
		goodStructs[i].a = true
		goodStructs[i].b = int64(i)
		goodStructs[i].c = true
		goodStructs[i].d = int32(i)
		goodStructs[i].e = true
		goodStructs[i].f = int64(i * 2)
	}

	goodElapsed := time.Since(start)
	goodSize := unsafe.Sizeof(GoodStructLayout{})

	fmt.Printf("  差的布局: %v, 大小=%d字节\n", badElapsed, badSize)
	fmt.Printf("  好的布局: %v, 大小=%d字节\n", goodElapsed, goodSize)
	fmt.Printf("  性能提升: %.2fx, 内存节省: %.2fx\n",
		float64(badElapsed.Nanoseconds())/float64(goodElapsed.Nanoseconds()),
		float64(badSize)/float64(goodSize))
}

// ==================
// 5. I/O优化技术
// ==================

// IOOptimizer I/O优化器
type IOOptimizer struct {
	bufferPools map[int]*sync.Pool
	metrics     IOMetrics
}

// IOMetrics I/O指标
type IOMetrics struct {
	TotalReads   int64
	TotalWrites  int64
	TotalBytes   int64
	BufferHits   int64
	BufferMisses int64
}

func NewIOOptimizer() *IOOptimizer {
	return &IOOptimizer{
		bufferPools: make(map[int]*sync.Pool),
	}
}

// ==================
// 5.1 批量I/O优化
// ==================

func (io *IOOptimizer) DemonstrateIOOptimization() {
	fmt.Println("\n=== I/O优化技术演示 ===")

	// 创建测试文件
	testFile := "io_test.dat"
	defer os.Remove(testFile)

	dataSize := 1024 * 1024 // 1MB
	data := make([]byte, dataSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// 1. 单次写入vs批量写入
	fmt.Println("1. 写入性能对比:")
	io.benchmarkWrite(testFile, data, false) // 单次写入
	io.benchmarkWrite(testFile, data, true)  // 批量写入

	// 2. 单次读取vs批量读取
	fmt.Println("2. 读取性能对比:")
	io.benchmarkRead(testFile, dataSize, false) // 单次读取
	io.benchmarkRead(testFile, dataSize, true)  // 批量读取

	// 3. 缓冲区优化
	fmt.Println("3. 缓冲区优化:")
	io.benchmarkBufferedIO(testFile, data)

	// 4. 内存映射I/O
	fmt.Println("4. 内存映射I/O:")
	io.benchmarkMemoryMappedIO(testFile, dataSize)
}

func (io *IOOptimizer) benchmarkWrite(filename string, data []byte, batch bool) {
	start := time.Now()

	if batch {
		// 批量写入
		file, err := os.Create(filename + ".batch")
		if err != nil {
			fmt.Printf("  批量写入失败: %v\n", err)
			return
		}
		defer file.Close()
		defer os.Remove(filename + ".batch")

		writer := bufio.NewWriterSize(file, 64*1024) // 64KB缓冲区
		if _, err := writer.Write(data); err != nil {
			fmt.Printf("  批量写入失败: %v\n", err)
			return
		}
		if err := writer.Flush(); err != nil {
			fmt.Printf("  刷新缓冲区失败: %v\n", err)
			return
		}

		elapsed := time.Since(start)
		throughput := float64(len(data)) / elapsed.Seconds() / 1024 / 1024
		fmt.Printf("  批量写入: %v, %.2f MB/s\n", elapsed, throughput)
	} else {
		// 单次写入
		file, err := os.Create(filename + ".single")
		if err != nil {
			fmt.Printf("  单次写入失败: %v\n", err)
			return
		}
		defer file.Close()
		defer os.Remove(filename + ".single")

		// 模拟小块写入
		chunkSize := 1024
		for i := 0; i < len(data); i += chunkSize {
			end := i + chunkSize
			if end > len(data) {
				end = len(data)
			}
			if _, err := file.Write(data[i:end]); err != nil {
				fmt.Printf("  单次写入失败: %v\n", err)
				return
			}
		}

		elapsed := time.Since(start)
		throughput := float64(len(data)) / elapsed.Seconds() / 1024 / 1024
		fmt.Printf("  单次写入: %v, %.2f MB/s\n", elapsed, throughput)
	}
}

func (io *IOOptimizer) benchmarkRead(filename string, size int, batch bool) {
	// 首先创建测试文件
	file, err := os.Create(filename)
	if err != nil {
		return
	}

	data := make([]byte, size)
	if _, err := file.Write(data); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
	}
	if err := file.Close(); err != nil {
		fmt.Printf("关闭文件失败: %v\n", err)
	}
	defer os.Remove(filename)

	start := time.Now()

	if batch {
		// 批量读取
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("  批量读取失败: %v\n", err)
			return
		}
		defer file.Close()

		reader := bufio.NewReaderSize(file, 64*1024) // 64KB缓冲区
		buffer := make([]byte, size)
		if _, err := reader.Read(buffer); err != nil {
			fmt.Printf("  批量读取失败: %v\n", err)
			return
		}

		elapsed := time.Since(start)
		throughput := float64(size) / elapsed.Seconds() / 1024 / 1024
		fmt.Printf("  批量读取: %v, %.2f MB/s\n", elapsed, throughput)
	} else {
		// 单次读取
		file, err := os.Open(filename)
		if err != nil {
			fmt.Printf("  单次读取失败: %v\n", err)
			return
		}
		defer file.Close()

		// 模拟小块读取
		chunkSize := 1024
		buffer := make([]byte, chunkSize)
		for i := 0; i < size; i += chunkSize {
			if _, err := file.Read(buffer); err != nil {
				fmt.Printf("  单次读取失败: %v\n", err)
				return
			}
		}

		elapsed := time.Since(start)
		throughput := float64(size) / elapsed.Seconds() / 1024 / 1024
		fmt.Printf("  单次读取: %v, %.2f MB/s\n", elapsed, throughput)
	}
}

func (io *IOOptimizer) benchmarkBufferedIO(filename string, data []byte) {
	// 测试不同缓冲区大小的影响
	bufferSizes := []int{4096, 8192, 16384, 32768, 65536}

	for _, bufSize := range bufferSizes {
		start := time.Now()

		file, err := os.Create(filename + fmt.Sprintf(".buf_%d", bufSize))
		if err != nil {
			continue
		}

		writer := bufio.NewWriterSize(file, bufSize)
		if _, err := writer.Write(data); err != nil {
			fmt.Printf("写入缓冲区失败: %v\n", err)
			continue
		}
		if err := writer.Flush(); err != nil {
			fmt.Printf("刷新缓冲区失败: %v\n", err)
			continue
		}
		if err := file.Close(); err != nil {
			fmt.Printf("关闭文件失败: %v\n", err)
		}

		elapsed := time.Since(start)
		throughput := float64(len(data)) / elapsed.Seconds() / 1024 / 1024

		fmt.Printf("  缓冲区%dKB: %v, %.2f MB/s\n",
			bufSize/1024, elapsed, throughput)

		os.Remove(filename + fmt.Sprintf(".buf_%d", bufSize))
	}
}

func (io *IOOptimizer) benchmarkMemoryMappedIO(filename string, size int) {
	// 简化的内存映射模拟
	start := time.Now()

	file, err := os.Create(filename + ".mmap")
	if err != nil {
		fmt.Printf("  内存映射失败: %v\n", err)
		return
	}
	defer file.Close()
	defer os.Remove(filename + ".mmap")

	// 预分配文件大小
	if err := file.Truncate(int64(size)); err != nil {
		fmt.Printf("  预分配文件失败: %v\n", err)
		return
	}

	// 模拟内存映射访问（实际需要系统调用）
	data := make([]byte, size)
	if _, err := file.WriteAt(data, 0); err != nil {
		fmt.Printf("  写入文件失败: %v\n", err)
		return
	}

	elapsed := time.Since(start)
	throughput := float64(size) / elapsed.Seconds() / 1024 / 1024
	fmt.Printf("  内存映射: %v, %.2f MB/s\n", elapsed, throughput)
}

// ==================
// 6. 网络优化技术
// ==================

// NetworkOptimizer 网络优化器
type NetworkOptimizer struct {
	connectionPools map[string]*ConnectionPool
	metrics         NetworkMetrics
}

// ConnectionPool 连接池
type ConnectionPool struct {
	pool       chan net.Conn
	factory    func() (net.Conn, error)
	maxConn    int
	activeConn int32
	totalConn  int64
	reuseCount int64
}

// NetworkMetrics 网络指标
type NetworkMetrics struct {
	TotalConnections int64
	PoolHits         int64
	PoolMisses       int64
	BytesSent        int64
	BytesReceived    int64
}

func NewNetworkOptimizer() *NetworkOptimizer {
	return &NetworkOptimizer{
		connectionPools: make(map[string]*ConnectionPool),
	}
}

func (no *NetworkOptimizer) CreateConnectionPool(name string, factory func() (net.Conn, error), maxConn int) {
	pool := &ConnectionPool{
		pool:    make(chan net.Conn, maxConn),
		factory: factory,
		maxConn: maxConn,
	}

	no.connectionPools[name] = pool
}

func (pool *ConnectionPool) Get() (net.Conn, error) {
	select {
	case conn := <-pool.pool:
		atomic.AddInt64(&pool.reuseCount, 1)
		return conn, nil
	default:
		atomic.AddInt64(&pool.totalConn, 1)
		return pool.factory()
	}
}

func (pool *ConnectionPool) Put(conn net.Conn) {
	select {
	case pool.pool <- conn:
		// 连接放回池中
	default:
		// 池已满，关闭连接
		if err := conn.Close(); err != nil {
			log.Printf("关闭连接失败: %v", err)
		}
	}
}

// ==================
// 7. 主演示函数
// ==================

func demonstrateOptimizationTechniques() {
	fmt.Println("=== Go深度优化技术演示 ===")

	// 1. 算法优化
	fmt.Println("\n1. 算法优化演示")
	algoOpt := NewAlgorithmOptimizer()
	algoOpt.DemonstrateSortingOptimization()

	// 2. 内存优化
	fmt.Println("\n2. 内存优化演示")
	memOpt := NewMemoryOptimizer()

	// 创建对象池
	memOpt.CreateObjectPool("large_struct", 1024, func() interface{} {
		return make([]byte, 1024)
	})

	memOpt.CreateBufferPool("4k_buffer", 4096)

	// 演示对象池使用
	start := time.Now()
	for i := 0; i < 10000; i++ {
		obj := memOpt.GetObject("large_struct")
		memOpt.PutObject("large_struct", obj)
	}
	elapsed := time.Since(start)
	fmt.Printf("对象池操作: %v\n", elapsed)

	// 零拷贝演示
	memOpt.DemonstrateZeroCopy()

	// 3. 并发优化
	fmt.Println("\n3. 并发优化演示")
	concOpt := NewConcurrencyOptimizer()
	concOpt.DemonstrateLockOptimization()

	// 4. 缓存优化
	fmt.Println("\n4. 缓存优化演示")
	cacheOpt := NewCacheOptimizer()
	cacheOpt.DemonstrateDataLocality()

	// 5. I/O优化
	fmt.Println("\n5. I/O优化演示")
	ioOpt := NewIOOptimizer()
	ioOpt.DemonstrateIOOptimization()

	// 6. 编译器优化提示
	fmt.Println("\n6. 编译器优化技巧")
	demonstrateCompilerOptimizations()
}

func demonstrateCompilerOptimizations() {
	fmt.Println("编译器优化技巧:")
	fmt.Println("1. 使用内联函数（go:noinline 标记避免内联）")
	fmt.Println("2. 循环展开和向量化")
	fmt.Println("3. 边界检查消除")
	fmt.Println("4. 函数内联和常量折叠")
	fmt.Println("5. 使用 -gcflags='-m' 查看内联决策")

	// 演示边界检查消除
	demonstrateBoundsCheckElimination()
}

func demonstrateBoundsCheckElimination() {
	size := 1000000
	data := make([]int, size)

	// 有边界检查的版本
	start := time.Now()
	sum1 := 0
	for i := 0; i < len(data); i++ {
		sum1 += data[i]
	}
	elapsed1 := time.Since(start)

	// 优化的版本（编译器可以消除边界检查）
	start = time.Now()
	sum2 := 0
	for i := range data {
		sum2 += data[i]
	}
	elapsed2 := time.Since(start)

	fmt.Printf("边界检查版本: %v, sum=%d\n", elapsed1, sum1)
	fmt.Printf("优化版本: %v, sum=%d\n", elapsed2, sum2)

	if elapsed1 > elapsed2 {
		fmt.Printf("性能提升: %.2fx\n", float64(elapsed1.Nanoseconds())/float64(elapsed2.Nanoseconds()))
	}
}

func (ao *AlgorithmOptimizer) printBenchmarkResults() {
	ao.mutex.RLock()
	defer ao.mutex.RUnlock()

	fmt.Println("\n基准测试结果汇总:")
	for name, result := range ao.benchmarkResults {
		fmt.Printf("  %s: %d ns/op, %s\n",
			name, result.NsPerOp, result.Complexity)
	}
}

func main() {
	demonstrateOptimizationTechniques()

	fmt.Println("\n=== Go深度优化技术演示完成 ===")
	fmt.Println("\n学习要点总结:")
	fmt.Println("1. 算法优化：选择合适的算法和数据结构")
	fmt.Println("2. 内存优化：零拷贝、对象池、内存布局优化")
	fmt.Println("3. 并发优化：锁优化、无锁编程、分片技术")
	fmt.Println("4. 缓存优化：数据局部性、缓存友好编程")
	fmt.Println("5. I/O优化：批量操作、缓冲区调优、内存映射")
	fmt.Println("6. 编译器优化：内联、边界检查消除、常量折叠")

	fmt.Println("\n高级优化技术:")
	fmt.Println("- SIMD向量化计算")
	fmt.Println("- CPU缓存行对齐")
	fmt.Println("- 分支预测优化")
	fmt.Println("- 热点代码路径优化")
	fmt.Println("- 硬件特性利用")
	fmt.Println("- 性能敏感的汇编优化")
}

/*
=== 练习题 ===

1. 算法优化：
   - 实现并行归并排序
   - 优化字符串匹配算法
   - 实现高效的哈希表
   - 分析时间空间复杂度权衡

2. 内存优化：
   - 实现自定义内存分配器
   - 优化结构体内存布局
   - 实现内存池管理系统
   - 分析内存碎片化问题

3. 并发优化：
   - 实现高性能无锁数据结构
   - 优化锁粒度和持有时间
   - 实现工作窃取调度器
   - 分析false sharing问题

4. 系统优化：
   - 实现零拷贝网络库
   - 优化系统调用开销
   - 实现高性能序列化
   - 分析CPU缓存友好编程

编译优化：
go build -gcflags='-m -l -N' # 查看优化信息
go build -ldflags='-s -w'    # 减小二进制大小
GOGC=off go run main.go      # 关闭GC测试

重要概念：
- 算法复杂度：时间和空间复杂度分析
- 内存层次：L1/L2/L3缓存、主内存、虚拟内存
- 并发控制：锁、原子操作、无锁编程
- 编译优化：内联、循环展开、常量传播
- 硬件特性：SIMD、分支预测、流水线
*/
