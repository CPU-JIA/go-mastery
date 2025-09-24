package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// =============================================================================
// 1. 泛型基础概念
// =============================================================================

/*
泛型（Generics）在 Go 1.18 中引入，允许编写类型参数化的代码。

泛型的好处：
1. 类型安全：编译时检查类型
2. 代码重用：避免为不同类型重复写相同逻辑
3. 性能：避免运行时类型断言和反射
4. 表达能力：更精确地表达代码意图

泛型语法：
1. 类型参数：[T any] 或 [T comparable] 或 [T Constraint]
2. 类型约束：any, comparable, 自定义约束
3. 类型实例化：Func[int](value) 或 var x Func[string]

常用约束：
- any: interface{} 的别名，任意类型
- comparable: 可比较类型（支持 == 和 != 操作）
- 自定义约束：定义允许的类型集合
*/

// =============================================================================
// 2. 基础泛型函数
// =============================================================================

// Max 返回两个可比较值中的较大者
func Max[T comparable](a, b T) T {
	// 注意：这里需要类型断言或约束来支持比较
	// 我们使用反射来演示，实际应用中会使用更具体的约束
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if va.Int() > vb.Int() {
			return a
		}
		return b
	case reflect.Float32, reflect.Float64:
		if va.Float() > vb.Float() {
			return a
		}
		return b
	case reflect.String:
		if va.String() > vb.String() {
			return a
		}
		return b
	}
	return a
}

// Min 返回两个可比较值中的较小者
func Min[T comparable](a, b T) T {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if va.Int() < vb.Int() {
			return a
		}
		return b
	case reflect.Float32, reflect.Float64:
		if va.Float() < vb.Float() {
			return a
		}
		return b
	case reflect.String:
		if va.String() < vb.String() {
			return a
		}
		return b
	}
	return a
}

// Swap 交换两个值
func Swap[T any](a, b *T) {
	*a, *b = *b, *a
}

// Zero 返回类型的零值
func Zero[T any]() T {
	var zero T
	return zero
}

// IsZero 检查值是否为零值
func IsZero[T comparable](v T) bool {
	var zero T
	return v == zero
}

func demonstrateBasicGenerics() {
	fmt.Println("=== 1. 基础泛型函数 ===")

	// 整数比较
	fmt.Printf("Max(10, 20) = %d\n", Max(10, 20))
	fmt.Printf("Min(10, 20) = %d\n", Min(10, 20))

	// 浮点数比较
	fmt.Printf("Max(3.14, 2.71) = %.2f\n", Max(3.14, 2.71))
	fmt.Printf("Min(3.14, 2.71) = %.2f\n", Min(3.14, 2.71))

	// 字符串比较
	fmt.Printf("Max(\"apple\", \"banana\") = %s\n", Max("apple", "banana"))
	fmt.Printf("Min(\"apple\", \"banana\") = %s\n", Min("apple", "banana"))

	// 值交换
	x, y := 100, 200
	fmt.Printf("交换前: x=%d, y=%d\n", x, y)
	Swap(&x, &y)
	fmt.Printf("交换后: x=%d, y=%d\n", x, y)

	// 零值操作
	fmt.Printf("int 零值: %d\n", Zero[int]())
	fmt.Printf("string 零值: '%s'\n", Zero[string]())
	fmt.Printf("IsZero(0): %v\n", IsZero(0))
	fmt.Printf("IsZero(5): %v\n", IsZero(5))

	fmt.Println()
}

// =============================================================================
// 3. 泛型数据结构
// =============================================================================

// Stack 泛型栈
type Stack[T any] struct {
	items []T
}

// NewStack 创建新的栈
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0),
	}
}

// Push 压栈
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop 出栈
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}

	index := len(s.items) - 1
	item := s.items[index]
	s.items = s.items[:index]
	return item, true
}

// Peek 查看栈顶元素
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Size 栈大小
func (s *Stack[T]) Size() int {
	return len(s.items)
}

// IsEmpty 检查栈是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Queue 泛型队列
type Queue[T any] struct {
	items []T
}

// NewQueue 创建新的队列
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items: make([]T, 0),
	}
}

// Enqueue 入队
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Dequeue 出队
func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}

	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// Front 查看队首元素
func (q *Queue[T]) Front() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	return q.items[0], true
}

// Size 队列大小
func (q *Queue[T]) Size() int {
	return len(q.items)
}

// IsEmpty 检查队列是否为空
func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

func demonstrateGenericDataStructures() {
	fmt.Println("=== 2. 泛型数据结构 ===")

	// 测试泛型栈
	fmt.Println("泛型栈测试:")
	intStack := NewStack[int]()
	intStack.Push(1)
	intStack.Push(2)
	intStack.Push(3)

	fmt.Printf("栈大小: %d\n", intStack.Size())
	if top, ok := intStack.Peek(); ok {
		fmt.Printf("栈顶元素: %d\n", top)
	}

	for !intStack.IsEmpty() {
		if item, ok := intStack.Pop(); ok {
			fmt.Printf("出栈: %d\n", item)
		}
	}

	// 测试字符串栈
	fmt.Println("字符串栈测试:")
	stringStack := NewStack[string]()
	stringStack.Push("first")
	stringStack.Push("second")
	stringStack.Push("third")

	for !stringStack.IsEmpty() {
		if item, ok := stringStack.Pop(); ok {
			fmt.Printf("出栈: %s\n", item)
		}
	}

	// 测试泛型队列
	fmt.Println("泛型队列测试:")
	intQueue := NewQueue[int]()
	intQueue.Enqueue(10)
	intQueue.Enqueue(20)
	intQueue.Enqueue(30)

	fmt.Printf("队列大小: %d\n", intQueue.Size())
	if front, ok := intQueue.Front(); ok {
		fmt.Printf("队首元素: %d\n", front)
	}

	for !intQueue.IsEmpty() {
		if item, ok := intQueue.Dequeue(); ok {
			fmt.Printf("出队: %d\n", item)
		}
	}

	fmt.Println()
}

// =============================================================================
// 4. 类型约束
// =============================================================================

// Numeric 数值类型约束
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Ordered 可排序类型约束
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// Stringer 字符串化约束
type Stringer interface {
	String() string
}

// Add 泛型加法
func Add[T Numeric](a, b T) T {
	return a + b
}

// Sum 计算切片元素总和
func Sum[T Numeric](slice []T) T {
	var sum T
	for _, v := range slice {
		sum += v
	}
	return sum
}

// Average 计算平均值
func Average[T Numeric](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	return Sum(slice) / T(len(slice))
}

// SortSlice 使用泛型排序切片
func SortSlice[T Ordered](slice []T) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}

// FindMax 查找切片中的最大值
func FindMax[T Ordered](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max, true
}

// FindMin 查找切片中的最小值
func FindMin[T Ordered](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min, true
}

func demonstrateTypeConstraints() {
	fmt.Println("=== 3. 类型约束 ===")

	// 数值运算
	fmt.Printf("Add(10, 20) = %d\n", Add(10, 20))
	fmt.Printf("Add(3.14, 2.86) = %.2f\n", Add(3.14, 2.86))

	// 切片操作
	intSlice := []int{1, 5, 3, 9, 2, 8}
	fmt.Printf("原始切片: %v\n", intSlice)
	fmt.Printf("总和: %d\n", Sum(intSlice))
	fmt.Printf("平均值: %d\n", Average(intSlice))

	if max, ok := FindMax(intSlice); ok {
		fmt.Printf("最大值: %d\n", max)
	}
	if min, ok := FindMin(intSlice); ok {
		fmt.Printf("最小值: %d\n", min)
	}

	// 排序
	SortSlice(intSlice)
	fmt.Printf("排序后: %v\n", intSlice)

	// 浮点数切片
	floatSlice := []float64{3.14, 2.71, 1.41, 1.73}
	fmt.Printf("浮点数切片: %v\n", floatSlice)
	fmt.Printf("总和: %.2f\n", Sum(floatSlice))
	fmt.Printf("平均值: %.2f\n", Average(floatSlice))

	// 字符串切片
	stringSlice := []string{"banana", "apple", "cherry", "date"}
	fmt.Printf("字符串切片: %v\n", stringSlice)
	SortSlice(stringSlice)
	fmt.Printf("排序后: %v\n", stringSlice)

	fmt.Println()
}

// =============================================================================
// 5. 泛型接口和方法
// =============================================================================

// Container 泛型容器接口
type Container[T any] interface {
	Add(item T)
	Remove() (T, bool)
	Size() int
	IsEmpty() bool
	Clear()
}

// ArrayList 泛型数组列表
type ArrayList[T any] struct {
	items []T
}

// NewArrayList 创建新的数组列表
func NewArrayList[T any]() *ArrayList[T] {
	return &ArrayList[T]{
		items: make([]T, 0),
	}
}

// Add 添加元素
func (al *ArrayList[T]) Add(item T) {
	al.items = append(al.items, item)
}

// Remove 移除最后一个元素
func (al *ArrayList[T]) Remove() (T, bool) {
	if len(al.items) == 0 {
		var zero T
		return zero, false
	}

	index := len(al.items) - 1
	item := al.items[index]
	al.items = al.items[:index]
	return item, true
}

// Get 获取指定索引的元素
func (al *ArrayList[T]) Get(index int) (T, bool) {
	if index < 0 || index >= len(al.items) {
		var zero T
		return zero, false
	}
	return al.items[index], true
}

// Set 设置指定索引的元素
func (al *ArrayList[T]) Set(index int, item T) bool {
	if index < 0 || index >= len(al.items) {
		return false
	}
	al.items[index] = item
	return true
}

// Size 获取大小
func (al *ArrayList[T]) Size() int {
	return len(al.items)
}

// IsEmpty 检查是否为空
func (al *ArrayList[T]) IsEmpty() bool {
	return len(al.items) == 0
}

// Clear 清空
func (al *ArrayList[T]) Clear() {
	al.items = al.items[:0]
}

// ToSlice 转换为切片
func (al *ArrayList[T]) ToSlice() []T {
	result := make([]T, len(al.items))
	copy(result, al.items)
	return result
}

// LinkedNode 链表节点
type LinkedNode[T any] struct {
	Data T
	Next *LinkedNode[T]
}

// LinkedList 泛型链表
type LinkedList[T any] struct {
	head *LinkedNode[T]
	tail *LinkedNode[T]
	size int
}

// NewLinkedList 创建新的链表
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// Add 添加元素到链表尾部
func (ll *LinkedList[T]) Add(item T) {
	newNode := &LinkedNode[T]{Data: item}

	if ll.head == nil {
		ll.head = newNode
		ll.tail = newNode
	} else {
		ll.tail.Next = newNode
		ll.tail = newNode
	}
	ll.size++
}

// Remove 移除头部元素
func (ll *LinkedList[T]) Remove() (T, bool) {
	if ll.head == nil {
		var zero T
		return zero, false
	}

	data := ll.head.Data
	ll.head = ll.head.Next
	if ll.head == nil {
		ll.tail = nil
	}
	ll.size--
	return data, true
}

// Size 获取大小
func (ll *LinkedList[T]) Size() int {
	return ll.size
}

// IsEmpty 检查是否为空
func (ll *LinkedList[T]) IsEmpty() bool {
	return ll.size == 0
}

// Clear 清空链表
func (ll *LinkedList[T]) Clear() {
	ll.head = nil
	ll.tail = nil
	ll.size = 0
}

// ToSlice 转换为切片
func (ll *LinkedList[T]) ToSlice() []T {
	result := make([]T, 0, ll.size)
	current := ll.head
	for current != nil {
		result = append(result, current.Data)
		current = current.Next
	}
	return result
}

func demonstrateGenericInterfaces() {
	fmt.Println("=== 4. 泛型接口和方法 ===")

	// 测试数组列表
	fmt.Println("数组列表测试:")
	arrayList := NewArrayList[string]()
	arrayList.Add("first")
	arrayList.Add("second")
	arrayList.Add("third")

	fmt.Printf("大小: %d\n", arrayList.Size())
	fmt.Printf("内容: %v\n", arrayList.ToSlice())

	if item, ok := arrayList.Get(1); ok {
		fmt.Printf("索引1的元素: %s\n", item)
	}

	arrayList.Set(1, "modified")
	fmt.Printf("修改后: %v\n", arrayList.ToSlice())

	// 测试链表
	fmt.Println("链表测试:")
	linkedList := NewLinkedList[int]()
	linkedList.Add(10)
	linkedList.Add(20)
	linkedList.Add(30)

	fmt.Printf("大小: %d\n", linkedList.Size())
	fmt.Printf("内容: %v\n", linkedList.ToSlice())

	for !linkedList.IsEmpty() {
		if item, ok := linkedList.Remove(); ok {
			fmt.Printf("移除: %d\n", item)
		}
	}

	fmt.Println()
}

// =============================================================================
// 6. 泛型算法
// =============================================================================

// Map 映射函数
func Map[T, R any](slice []T, fn func(T) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter 过滤函数
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce 归约函数
func Reduce[T, R any](slice []T, initial R, fn func(R, T) R) R {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// Contains 检查切片是否包含指定元素
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// IndexOf 查找元素在切片中的索引
func IndexOf[T comparable](slice []T, item T) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// Reverse 反转切片
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Unique 去重
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	var result []T

	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Chunk 分块
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func demonstrateGenericAlgorithms() {
	fmt.Println("=== 5. 泛型算法 ===")

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Printf("原始数组: %v\n", numbers)

	// Map: 每个数字乘以2
	doubled := Map(numbers, func(x int) int { return x * 2 })
	fmt.Printf("每个数乘以2: %v\n", doubled)

	// Map: 数字转字符串
	strings := Map(numbers, func(x int) string { return strconv.Itoa(x) })
	fmt.Printf("转换为字符串: %v\n", strings)

	// Filter: 过滤偶数
	evens := Filter(numbers, func(x int) bool { return x%2 == 0 })
	fmt.Printf("偶数: %v\n", evens)

	// Filter: 过滤大于5的数
	greaterThan5 := Filter(numbers, func(x int) bool { return x > 5 })
	fmt.Printf("大于5的数: %v\n", greaterThan5)

	// Reduce: 计算总和
	sum := Reduce(numbers, 0, func(acc, x int) int { return acc + x })
	fmt.Printf("总和: %d\n", sum)

	// Reduce: 计算乘积
	product := Reduce(numbers[:5], 1, func(acc, x int) int { return acc * x })
	fmt.Printf("前5个数的乘积: %d\n", product)

	// Contains
	fmt.Printf("包含5: %v\n", Contains(numbers, 5))
	fmt.Printf("包含15: %v\n", Contains(numbers, 15))

	// IndexOf
	fmt.Printf("5的索引: %d\n", IndexOf(numbers, 5))
	fmt.Printf("15的索引: %d\n", IndexOf(numbers, 15))

	// Reverse
	testSlice := []string{"a", "b", "c", "d", "e"}
	fmt.Printf("原始字符串数组: %v\n", testSlice)
	Reverse(testSlice)
	fmt.Printf("反转后: %v\n", testSlice)

	// Unique
	duplicates := []int{1, 2, 2, 3, 3, 3, 4, 4, 5}
	unique := Unique(duplicates)
	fmt.Printf("去重前: %v\n", duplicates)
	fmt.Printf("去重后: %v\n", unique)

	// Chunk
	chunks := Chunk(numbers, 3)
	fmt.Printf("分块(大小3): %v\n", chunks)

	fmt.Println()
}

// =============================================================================
// 7. 泛型与并发
// =============================================================================

// Channel 泛型通道包装器
type Channel[T any] struct {
	ch chan T
}

// NewChannel 创建新的泛型通道
func NewChannel[T any](buffer int) *Channel[T] {
	return &Channel[T]{
		ch: make(chan T, buffer),
	}
}

// Send 发送数据
func (c *Channel[T]) Send(data T) {
	c.ch <- data
}

// Receive 接收数据
func (c *Channel[T]) Receive() T {
	return <-c.ch
}

// TryReceive 尝试接收数据
func (c *Channel[T]) TryReceive() (T, bool) {
	select {
	case data := <-c.ch:
		return data, true
	default:
		var zero T
		return zero, false
	}
}

// Close 关闭通道
func (c *Channel[T]) Close() {
	close(c.ch)
}

// Worker 泛型工作函数类型
type Worker[T, R any] func(T) R

// ParallelMap 并行映射
func ParallelMap[T, R any](slice []T, worker Worker[T, R], workers int) []R {
	if len(slice) == 0 || workers <= 0 {
		return []R{}
	}

	// 输入通道
	input := make(chan T, len(slice))
	for _, item := range slice {
		input <- item
	}
	close(input)

	// 输出通道
	output := make(chan R, len(slice))

	// 启动工作者
	for i := 0; i < workers; i++ {
		go func() {
			for item := range input {
				output <- worker(item)
			}
		}()
	}

	// 收集结果
	results := make([]R, len(slice))
	for i := range results {
		results[i] = <-output
	}

	return results
}

func demonstrateGenericConcurrency() {
	fmt.Println("=== 6. 泛型与并发 ===")

	// 测试泛型通道
	fmt.Println("泛型通道测试:")
	ch := NewChannel[string](2)

	go func() {
		ch.Send("Hello")
		ch.Send("World")
		ch.Close()
	}()

	for i := 0; i < 2; i++ {
		if data, ok := ch.TryReceive(); ok {
			fmt.Printf("接收到: %s\n", data)
		}
	}

	// 并行处理示例
	fmt.Println("并行处理示例:")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// 定义工作函数：计算平方
	squareWorker := func(x int) int {
		return x * x
	}

	// 并行计算平方
	squares := ParallelMap(numbers, squareWorker, 3)
	fmt.Printf("原始数字: %v\n", numbers)
	fmt.Printf("平方结果: %v\n", squares)

	// 字符串处理示例
	words := []string{"hello", "world", "go", "generics", "concurrent"}

	// 定义工作函数：转换为大写
	upperWorker := func(s string) string {
		return strings.ToUpper(s)
	}

	upperWords := ParallelMap(words, upperWorker, 2)
	fmt.Printf("原始单词: %v\n", words)
	fmt.Printf("大写单词: %v\n", upperWords)

	fmt.Println()
}

// =============================================================================
// 8. 实际应用：泛型缓存
// =============================================================================

// Cache 泛型缓存接口
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Clear()
	Size() int
	Keys() []K
}

// MemoryCache 内存缓存实现
type MemoryCache[K comparable, V any] struct {
	data map[K]V
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache[K comparable, V any]() *MemoryCache[K, V] {
	return &MemoryCache[K, V]{
		data: make(map[K]V),
	}
}

// Get 获取缓存值
func (c *MemoryCache[K, V]) Get(key K) (V, bool) {
	value, exists := c.data[key]
	return value, exists
}

// Set 设置缓存值
func (c *MemoryCache[K, V]) Set(key K, value V) {
	c.data[key] = value
}

// Delete 删除缓存项
func (c *MemoryCache[K, V]) Delete(key K) {
	delete(c.data, key)
}

// Clear 清空缓存
func (c *MemoryCache[K, V]) Clear() {
	c.data = make(map[K]V)
}

// Size 缓存大小
func (c *MemoryCache[K, V]) Size() int {
	return len(c.data)
}

// Keys 获取所有键
func (c *MemoryCache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// User 用户结构体（用于缓存演示）
type User struct {
	ID   int
	Name string
	Age  int
}

func demonstrateGenericCache() {
	fmt.Println("=== 7. 实际应用：泛型缓存 ===")

	// 字符串->整数缓存
	intCache := NewMemoryCache[string, int]()
	intCache.Set("age", 25)
	intCache.Set("score", 95)
	intCache.Set("level", 10)

	fmt.Println("整数缓存:")
	for _, key := range intCache.Keys() {
		if value, ok := intCache.Get(key); ok {
			fmt.Printf("  %s: %d\n", key, value)
		}
	}

	// 整数->用户缓存
	userCache := NewMemoryCache[int, User]()
	userCache.Set(1, User{ID: 1, Name: "Alice", Age: 25})
	userCache.Set(2, User{ID: 2, Name: "Bob", Age: 30})
	userCache.Set(3, User{ID: 3, Name: "Carol", Age: 28})

	fmt.Println("用户缓存:")
	for _, key := range userCache.Keys() {
		if user, ok := userCache.Get(key); ok {
			fmt.Printf("  %d: {ID: %d, Name: %s, Age: %d}\n",
				key, user.ID, user.Name, user.Age)
		}
	}

	fmt.Printf("缓存大小: %d\n", userCache.Size())

	// 删除操作
	userCache.Delete(2)
	fmt.Printf("删除用户2后的大小: %d\n", userCache.Size())

	fmt.Println()
}

// =============================================================================
// 9. 泛型最佳实践和性能考虑
// =============================================================================

func demonstrateGenericsBestPractices() {
	fmt.Println("=== 8. 泛型最佳实践 ===")

	fmt.Println("1. 类型约束设计:")
	fmt.Println("   - 使用最小必要约束")
	fmt.Println("   - 优先使用内置约束 (any, comparable)")
	fmt.Println("   - 合理使用类型联合 (~int | ~string)")

	fmt.Println("\n2. 命名约定:")
	fmt.Println("   - 单个类型参数使用 T")
	fmt.Println("   - 多个类型参数使用 T, U, V 或有意义的名称")
	fmt.Println("   - 约束名称应该描述其用途")

	fmt.Println("\n3. 性能考虑:")
	fmt.Println("   - 泛型会产生代码膨胀")
	fmt.Println("   - 编译时类型检查，运行时无额外开销")
	fmt.Println("   - 避免过度泛型化")

	fmt.Println("\n4. 设计原则:")
	fmt.Println("   - 当类型参数提供明确价值时才使用泛型")
	fmt.Println("   - 避免为了泛型而泛型")
	fmt.Println("   - 保持接口简单和聚焦")

	fmt.Println("\n5. 常见用例:")
	fmt.Println("   - 数据结构 (栈、队列、树)")
	fmt.Println("   - 算法 (排序、搜索、映射)")
	fmt.Println("   - 工具函数 (转换、验证)")
	fmt.Println("   - 缓存和容器")

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 泛型编程 - 完整示例")
	fmt.Println("========================")

	demonstrateBasicGenerics()
	demonstrateGenericDataStructures()
	demonstrateTypeConstraints()
	demonstrateGenericInterfaces()
	demonstrateGenericAlgorithms()
	demonstrateGenericConcurrency()
	demonstrateGenericCache()
	demonstrateGenericsBestPractices()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 实现一个泛型二叉搜索树")
	fmt.Println("2. 创建一个支持过期时间的泛型缓存")
	fmt.Println("3. 实现泛型的生产者-消费者模式")
	fmt.Println("4. 编写泛型的函数式编程工具集")
	fmt.Println("5. 创建一个泛型的事件系统")
	fmt.Println("6. 实现泛型的配置管理器")
	fmt.Println("7. 设计泛型的数据验证框架")
	fmt.Println("\n在此文件中实现这些练习，掌握Go泛型的强大功能！")
}
