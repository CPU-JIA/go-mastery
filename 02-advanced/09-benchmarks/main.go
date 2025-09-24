package main

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// 1. 基准测试基础
// =============================================================================

/*
基准测试（Benchmark）用于测量代码的性能。

基准测试规则：
- 基准测试函数名以 Benchmark 开头
- 函数签名：func BenchmarkXxx(b *testing.B)
- 使用 b.N 作为循环次数
- Go 会自动调整 b.N 的值以获得可靠的测量结果

运行基准测试：
go test -bench=.                    // 运行所有基准测试
go test -bench=BenchmarkName        // 运行特定基准测试
go test -bench=. -benchmem          // 显示内存分配信息
go test -bench=. -count=5           // 运行5次取平均值
go test -bench=. -cpu=1,2,4         // 在不同CPU数量下测试
go test -bench=. -benchtime=10s     // 运行10秒

基准测试输出解释：
BenchmarkFunction-8   1000000   1234 ns/op   56 B/op   2 allocs/op
- Function: 测试函数名
- 8: GOMAXPROCS值
- 1000000: 运行次数
- 1234 ns/op: 每次操作的纳秒数
- 56 B/op: 每次操作分配的字节数
- 2 allocs/op: 每次操作的内存分配次数
*/

// =============================================================================
// 2. 字符串操作性能对比
// =============================================================================

// StringConcatenation 字符串拼接的不同方法
type StringConcatenation struct{}

// UsingPlus 使用 + 操作符拼接字符串
func (sc *StringConcatenation) UsingPlus(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s
	}
	return result
}

// UsingBuilder 使用 strings.Builder 拼接字符串
func (sc *StringConcatenation) UsingBuilder(strs []string) string {
	var builder strings.Builder
	for _, s := range strs {
		builder.WriteString(s)
	}
	return builder.String()
}

// UsingBuilderWithCap 使用预分配容量的 strings.Builder
func (sc *StringConcatenation) UsingBuilderWithCap(strs []string) string {
	totalLen := 0
	for _, s := range strs {
		totalLen += len(s)
	}

	var builder strings.Builder
	builder.Grow(totalLen)
	for _, s := range strs {
		builder.WriteString(s)
	}
	return builder.String()
}

// UsingJoin 使用 strings.Join 拼接字符串
func (sc *StringConcatenation) UsingJoin(strs []string) string {
	return strings.Join(strs, "")
}

// UsingByteSlice 使用字节切片拼接
func (sc *StringConcatenation) UsingByteSlice(strs []string) string {
	var result []byte
	for _, s := range strs {
		result = append(result, []byte(s)...)
	}
	return string(result)
}

// =============================================================================
// 3. 切片操作性能对比
// =============================================================================

// SliceOperations 切片操作的不同方法
type SliceOperations struct{}

// AppendOne 逐个添加元素
func (so *SliceOperations) AppendOne(n int) []int {
	var slice []int
	for i := 0; i < n; i++ {
		slice = append(slice, i)
	}
	return slice
}

// AppendWithCap 预分配容量后添加元素
func (so *SliceOperations) AppendWithCap(n int) []int {
	slice := make([]int, 0, n)
	for i := 0; i < n; i++ {
		slice = append(slice, i)
	}
	return slice
}

// MakeWithIndex 使用make预分配并通过索引赋值
func (so *SliceOperations) MakeWithIndex(n int) []int {
	slice := make([]int, n)
	for i := 0; i < n; i++ {
		slice[i] = i
	}
	return slice
}

// CopySlice 复制切片的不同方法
func (so *SliceOperations) CopySlice(src []int) []int {
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

// AppendSlice 使用append复制切片
func (so *SliceOperations) AppendSlice(src []int) []int {
	return append([]int(nil), src...)
}

// =============================================================================
// 4. 映射操作性能对比
// =============================================================================

// MapOperations 映射操作的不同方法
type MapOperations struct{}

// CreateMap 创建并填充映射
func (mo *MapOperations) CreateMap(n int) map[int]string {
	m := make(map[int]string)
	for i := 0; i < n; i++ {
		m[i] = strconv.Itoa(i)
	}
	return m
}

// CreateMapWithCap 预分配容量创建映射
func (mo *MapOperations) CreateMapWithCap(n int) map[int]string {
	m := make(map[int]string, n)
	for i := 0; i < n; i++ {
		m[i] = strconv.Itoa(i)
	}
	return m
}

// AccessMap 访问映射元素
func (mo *MapOperations) AccessMap(m map[int]string, keys []int) []string {
	var results []string
	for _, key := range keys {
		if value, ok := m[key]; ok {
			results = append(results, value)
		}
	}
	return results
}

// AccessMapDirect 直接访问映射元素（不检查存在性）
func (mo *MapOperations) AccessMapDirect(m map[int]string, keys []int) []string {
	var results []string
	for _, key := range keys {
		results = append(results, m[key])
	}
	return results
}

// =============================================================================
// 5. 排序算法性能对比
// =============================================================================

// SortingAlgorithms 不同排序算法的实现
type SortingAlgorithms struct{}

// BubbleSort 冒泡排序
func (sa *SortingAlgorithms) BubbleSort(arr []int) {
	n := len(arr)
	for i := 0; i < n; i++ {
		for j := 0; j < n-1-i; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

// QuickSort 快速排序
func (sa *SortingAlgorithms) QuickSort(arr []int) {
	sa.quickSortHelper(arr, 0, len(arr)-1)
}

func (sa *SortingAlgorithms) quickSortHelper(arr []int, low, high int) {
	if low < high {
		pi := sa.partition(arr, low, high)
		sa.quickSortHelper(arr, low, pi-1)
		sa.quickSortHelper(arr, pi+1, high)
	}
}

func (sa *SortingAlgorithms) partition(arr []int, low, high int) int {
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

// BuiltinSort 使用内置排序
func (sa *SortingAlgorithms) BuiltinSort(arr []int) {
	sort.Ints(arr)
}

// InsertionSort 插入排序
func (sa *SortingAlgorithms) InsertionSort(arr []int) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

// =============================================================================
// 6. 哈希算法性能对比
// =============================================================================

// HashingAlgorithms 不同哈希算法的性能测试
type HashingAlgorithms struct{}

// MD5Hash 使用MD5哈希
func (ha *HashingAlgorithms) MD5Hash(data []byte) []byte {
	hasher := md5.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// SHA256Hash 使用SHA256哈希
func (ha *HashingAlgorithms) SHA256Hash(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// SimpleHash 简单哈希函数
func (ha *HashingAlgorithms) SimpleHash(data []byte) uint32 {
	var hash uint32
	for _, b := range data {
		hash = hash*31 + uint32(b)
	}
	return hash
}

// HashWithInterface 使用接口的哈希
func (ha *HashingAlgorithms) HashWithInterface(data []byte, hasher hash.Hash) []byte {
	hasher.Reset()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// =============================================================================
// 7. 并发性能测试
// =============================================================================

// ConcurrencyBenchmarks 并发性能测试
type ConcurrencyBenchmarks struct{}

// SequentialWork 顺序执行工作
func (cb *ConcurrencyBenchmarks) SequentialWork(tasks []func()) {
	for _, task := range tasks {
		task()
	}
}

// ConcurrentWork 并发执行工作
func (cb *ConcurrencyBenchmarks) ConcurrentWork(tasks []func()) {
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(t func()) {
			defer wg.Done()
			t()
		}(task)
	}
	wg.Wait()
}

// ConcurrentWorkWithLimit 限制并发数的并发执行
func (cb *ConcurrencyBenchmarks) ConcurrentWorkWithLimit(tasks []func(), limit int) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, limit)

	for _, task := range tasks {
		wg.Add(1)
		go func(t func()) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			t()
		}(task)
	}
	wg.Wait()
}

// =============================================================================
// 8. 内存分配优化
// =============================================================================

// MemoryOptimization 内存分配优化示例
type MemoryOptimization struct{}

// CreateObjectsWithoutPool 不使用对象池
func (mo *MemoryOptimization) CreateObjectsWithoutPool(n int) []*Object {
	var objects []*Object
	for i := 0; i < n; i++ {
		obj := &Object{
			ID:   i,
			Name: fmt.Sprintf("Object_%d", i),
			Data: make([]byte, 1024),
		}
		objects = append(objects, obj)
	}
	return objects
}

// CreateObjectsWithPool 使用对象池
func (mo *MemoryOptimization) CreateObjectsWithPool(n int, pool *ObjectPool) []*Object {
	var objects []*Object
	for i := 0; i < n; i++ {
		obj := pool.Get()
		obj.ID = i
		obj.Name = fmt.Sprintf("Object_%d", i)
		objects = append(objects, obj)
	}
	return objects
}

// Object 测试对象
type Object struct {
	ID   int
	Name string
	Data []byte
}

// Reset 重置对象
func (o *Object) Reset() {
	o.ID = 0
	o.Name = ""
	for i := range o.Data {
		o.Data[i] = 0
	}
}

// ObjectPool 对象池
type ObjectPool struct {
	pool sync.Pool
}

// NewObjectPool 创建对象池
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{
			New: func() interface{} {
				return &Object{
					Data: make([]byte, 1024),
				}
			},
		},
	}
}

// Get 从池中获取对象
func (p *ObjectPool) Get() *Object {
	return p.pool.Get().(*Object)
}

// Put 将对象放回池中
func (p *ObjectPool) Put(obj *Object) {
	obj.Reset()
	p.pool.Put(obj)
}

// =============================================================================
// 9. JSON序列化性能对比
// =============================================================================

// SerializationBenchmarks 序列化性能测试
type SerializationBenchmarks struct{}

// Person 测试用的结构体
type Person struct {
	Name     string            `json:"name"`
	Age      int               `json:"age"`
	Email    string            `json:"email"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
}

// CreateLargePerson 创建包含大量数据的Person
func (sb *SerializationBenchmarks) CreateLargePerson() *Person {
	metadata := make(map[string]string)
	for i := 0; i < 100; i++ {
		metadata[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	return &Person{
		Name:     "张三",
		Age:      30,
		Email:    "zhangsan@example.com",
		Address:  "北京市朝阳区某某街道某某号",
		Metadata: metadata,
	}
}

// =============================================================================
// 10. 缓存性能测试
// =============================================================================

// CacheImplementations 不同缓存实现的性能测试
type CacheImplementations struct{}

// SimpleMapCache 简单的映射缓存
type SimpleMapCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSimpleMapCache 创建简单映射缓存
func NewSimpleMapCache() *SimpleMapCache {
	return &SimpleMapCache{
		data: make(map[string]interface{}),
	}
}

// Get 获取缓存值
func (c *SimpleMapCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

// Set 设置缓存值
func (c *SimpleMapCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// SyncMapCache 使用sync.Map的缓存
type SyncMapCache struct {
	data sync.Map
}

// NewSyncMapCache 创建sync.Map缓存
func NewSyncMapCache() *SyncMapCache {
	return &SyncMapCache{}
}

// Get 获取缓存值
func (c *SyncMapCache) Get(key string) (interface{}, bool) {
	return c.data.Load(key)
}

// Set 设置缓存值
func (c *SyncMapCache) Set(key string, value interface{}) {
	c.data.Store(key, value)
}

// =============================================================================
// 11. 演示函数
// =============================================================================

func demonstrateStringConcatenation() {
	fmt.Println("=== 1. 字符串拼接性能对比 ===")

	sc := &StringConcatenation{}
	testStrings := []string{"Hello", " ", "World", "!", " ", "Go", " ", "Benchmark"}

	start := time.Now()
	result1 := sc.UsingPlus(testStrings)
	fmt.Printf("+ 操作符: %s (耗时: %v)\n", result1, time.Since(start))

	start = time.Now()
	result2 := sc.UsingBuilder(testStrings)
	fmt.Printf("Builder: %s (耗时: %v)\n", result2, time.Since(start))

	start = time.Now()
	result3 := sc.UsingJoin(testStrings)
	fmt.Printf("Join: %s (耗时: %v)\n", result3, time.Since(start))

	fmt.Println()
}

func demonstrateSliceOperations() {
	fmt.Println("=== 2. 切片操作性能对比 ===")

	so := &SliceOperations{}
	n := 10000

	start := time.Now()
	slice1 := so.AppendOne(n)
	fmt.Printf("逐个添加: 长度=%d (耗时: %v)\n", len(slice1), time.Since(start))

	start = time.Now()
	slice2 := so.AppendWithCap(n)
	fmt.Printf("预分配容量: 长度=%d (耗时: %v)\n", len(slice2), time.Since(start))

	start = time.Now()
	slice3 := so.MakeWithIndex(n)
	fmt.Printf("make预分配: 长度=%d (耗时: %v)\n", len(slice3), time.Since(start))

	fmt.Println()
}

func demonstrateSortingPerformance() {
	fmt.Println("=== 3. 排序算法性能对比 ===")

	sa := &SortingAlgorithms{}

	// 生成测试数据
	generateRandomSlice := func(size int) []int {
		rand.Seed(time.Now().UnixNano())
		slice := make([]int, size)
		for i := range slice {
			slice[i] = rand.Intn(1000)
		}
		return slice
	}

	size := 1000

	// 测试插入排序
	data := generateRandomSlice(size)
	start := time.Now()
	sa.InsertionSort(data)
	fmt.Printf("插入排序 (n=%d): %v\n", size, time.Since(start))

	// 测试快速排序
	data = generateRandomSlice(size)
	start = time.Now()
	sa.QuickSort(data)
	fmt.Printf("快速排序 (n=%d): %v\n", size, time.Since(start))

	// 测试内置排序
	data = generateRandomSlice(size)
	start = time.Now()
	sa.BuiltinSort(data)
	fmt.Printf("内置排序 (n=%d): %v\n", size, time.Since(start))

	fmt.Println()
}

func demonstrateHashingPerformance() {
	fmt.Println("=== 4. 哈希算法性能对比 ===")

	ha := &HashingAlgorithms{}
	data := []byte("这是一个用于测试哈希性能的示例数据，包含中文和English混合内容")

	start := time.Now()
	md5Result := ha.MD5Hash(data)
	fmt.Printf("MD5哈希: %x (耗时: %v)\n", md5Result, time.Since(start))

	start = time.Now()
	sha256Result := ha.SHA256Hash(data)
	fmt.Printf("SHA256哈希: %x (耗时: %v)\n", sha256Result[:8], time.Since(start))

	start = time.Now()
	simpleResult := ha.SimpleHash(data)
	fmt.Printf("简单哈希: %x (耗时: %v)\n", simpleResult, time.Since(start))

	fmt.Println()
}

func demonstrateConcurrencyPerformance() {
	fmt.Println("=== 5. 并发性能对比 ===")

	cb := &ConcurrencyBenchmarks{}

	// 创建一些模拟任务
	createTask := func(id int) func() {
		return func() {
			// 模拟一些工作
			time.Sleep(time.Millisecond)
		}
	}

	tasks := make([]func(), 100)
	for i := range tasks {
		tasks[i] = createTask(i)
	}

	start := time.Now()
	cb.SequentialWork(tasks)
	fmt.Printf("顺序执行: %v\n", time.Since(start))

	start = time.Now()
	cb.ConcurrentWork(tasks)
	fmt.Printf("并发执行: %v\n", time.Since(start))

	start = time.Now()
	cb.ConcurrentWorkWithLimit(tasks, 10)
	fmt.Printf("限制并发数(10): %v\n", time.Since(start))

	fmt.Println()
}

func demonstrateMemoryOptimization() {
	fmt.Println("=== 6. 内存分配优化 ===")

	mo := &MemoryOptimization{}
	pool := NewObjectPool()
	n := 1000

	start := time.Now()
	objects1 := mo.CreateObjectsWithoutPool(n)
	fmt.Printf("不使用对象池: 创建%d个对象 (耗时: %v)\n", len(objects1), time.Since(start))

	start = time.Now()
	objects2 := mo.CreateObjectsWithPool(n, pool)
	fmt.Printf("使用对象池: 创建%d个对象 (耗时: %v)\n", len(objects2), time.Since(start))

	// 将对象放回池中
	for _, obj := range objects2 {
		pool.Put(obj)
	}

	fmt.Println()
}

func demonstrateBenchmarkingBestPractices() {
	fmt.Println("=== 7. 基准测试最佳实践 ===")
	fmt.Println("1. 避免在基准测试中进行不必要的工作")
	fmt.Println("2. 使用 b.ResetTimer() 重置计时器")
	fmt.Println("3. 使用 b.StopTimer() 和 b.StartTimer() 控制计时")
	fmt.Println("4. 避免编译器优化掉被测试的代码")
	fmt.Println("5. 为不同的输入大小创建子基准测试")
	fmt.Println("6. 使用 testing.B.RunParallel() 测试并发性能")
	fmt.Println("7. 分析内存分配使用 -benchmem 标志")
	fmt.Println("8. 使用 benchmark 包进行更复杂的性能分析")

	fmt.Println("\n=== 性能分析工具 ===")
	fmt.Println("1. go test -cpuprofile=cpu.prof -bench=.")
	fmt.Println("2. go test -memprofile=mem.prof -bench=.")
	fmt.Println("3. go tool pprof cpu.prof")
	fmt.Println("4. go tool pprof mem.prof")
	fmt.Println("5. go test -trace=trace.out -bench=.")
	fmt.Println("6. go tool trace trace.out")

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 基准测试和性能优化 - 完整示例")
	fmt.Println("===================================")

	demonstrateStringConcatenation()
	demonstrateSliceOperations()
	demonstrateSortingPerformance()
	demonstrateHashingPerformance()
	demonstrateConcurrencyPerformance()
	demonstrateMemoryOptimization()
	demonstrateBenchmarkingBestPractices()

	fmt.Println("=== 基准测试文件示例 ===")
	fmt.Println("创建对应的 *_bench_test.go 文件来进行基准测试：")
	fmt.Println("1. string_concat_bench_test.go - 字符串拼接基准测试")
	fmt.Println("2. slice_operations_bench_test.go - 切片操作基准测试")
	fmt.Println("3. sorting_bench_test.go - 排序算法基准测试")
	fmt.Println("4. hashing_bench_test.go - 哈希算法基准测试")
	fmt.Println("5. concurrency_bench_test.go - 并发性能基准测试")

	fmt.Println("\n=== 运行基准测试命令 ===")
	fmt.Println("go test -bench=.                    // 运行所有基准测试")
	fmt.Println("go test -bench=. -benchmem          // 显示内存分配")
	fmt.Println("go test -bench=. -count=3           // 运行3次")
	fmt.Println("go test -bench=. -cpuprofile=cpu.prof")
	fmt.Println("go test -bench=. -memprofile=mem.prof")

	fmt.Println("\n=== 练习任务 ===")
	fmt.Println("1. 为字符串拼接创建详细的基准测试")
	fmt.Println("2. 比较不同容器类型的性能差异")
	fmt.Println("3. 测试不同并发模式的性能表现")
	fmt.Println("4. 分析内存分配模式和优化策略")
	fmt.Println("5. 创建自定义的性能分析工具")
	fmt.Println("6. 实现一个简单的性能监控系统")
	fmt.Println("\n创建基准测试文件并运行性能分析！")
}
