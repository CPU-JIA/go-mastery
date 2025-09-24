package main

import (
	"fmt"
	"strings"
	"time"
)

/*
=== Go语言第十三课：闭包(Closures) ===

学习目标：
1. 理解闭包的概念和原理
2. 掌握闭包的创建和使用
3. 学会闭包的变量捕获机制
4. 了解闭包的实际应用场景
5. 掌握高阶函数和函数式编程

Go闭包特点：
- 匿名函数可以访问外部变量
- 闭包会"捕获"外部变量的引用
- 闭包可以作为函数返回值
- 支持高阶函数和函数式编程
- 常用于回调、装饰器、工厂函数等
*/

func main() {
	fmt.Println("=== Go语言闭包学习 ===")

	// 1. 基本闭包概念
	demonstrateBasicClosures()

	// 2. 变量捕获机制
	demonstrateVariableCapture()

	// 3. 闭包作为返回值
	demonstrateClosureReturns()

	// 4. 闭包在循环中的陷阱
	demonstrateClosureInLoops()

	// 5. 高阶函数
	demonstrateHigherOrderFunctions()

	// 6. 函数式编程
	demonstrateFunctionalProgramming()

	// 7. 装饰器模式
	demonstrateDecoratorPattern()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本闭包概念
func demonstrateBasicClosures() {
	fmt.Println("1. 基本闭包概念:")

	// 简单的闭包
	x := 10
	closure := func() {
		fmt.Printf("闭包访问外部变量x: %d\n", x)
	}
	closure()

	// 闭包修改外部变量
	counter := 0
	increment := func() {
		counter++
		fmt.Printf("计数器: %d\n", counter)
	}

	increment()
	increment()
	increment()
	fmt.Printf("外部变量counter: %d\n", counter)

	// 多个闭包共享变量
	shared := 0
	inc1 := func() { shared++; fmt.Printf("inc1: %d\n", shared) }
	inc2 := func() { shared += 2; fmt.Printf("inc2: %d\n", shared) }

	inc1() // 1
	inc2() // 3
	inc1() // 4

	// 闭包作为参数
	processWithCallback(5, func(n int) {
		fmt.Printf("回调处理: %d的平方是%d\n", n, n*n)
	})

	// 闭包捕获多个变量
	name := "张三"
	age := 25
	greet := func() {
		fmt.Printf("你好，我是%s，今年%d岁\n", name, age)
	}
	greet()

	// 修改被捕获的变量
	name = "李四"
	age = 30
	greet() // 输出会使用新值

	fmt.Println()
}

// 变量捕获机制
func demonstrateVariableCapture() {
	fmt.Println("2. 变量捕获机制:")

	// 值捕获 vs 引用捕获
	fmt.Println("引用捕获演示:")
	count := 0
	functions := make([]func(), 3)

	for i := 0; i < 3; i++ {
		localI := i // 创建局部变量
		functions[i] = func() {
			count++
			fmt.Printf("  函数%d: i=%d, count=%d\n", localI, localI, count)
		}
	}

	fmt.Println("执行闭包:")
	for i, fn := range functions {
		fmt.Printf("调用函数%d: ", i)
		fn()
	}

	// 闭包的生命周期
	fmt.Println("\n闭包生命周期:")
	createCounter := func(start int) func() int {
		counter := start
		return func() int {
			counter++
			return counter
		}
	}

	counter1 := createCounter(0)
	counter2 := createCounter(100)

	fmt.Printf("counter1: %d, %d, %d\n", counter1(), counter1(), counter1())
	fmt.Printf("counter2: %d, %d, %d\n", counter2(), counter2(), counter2())

	// 闭包修改外部切片
	fmt.Println("\n修改外部切片:")
	numbers := []int{1, 2, 3, 4, 5}
	doubler := func() {
		for i := range numbers {
			numbers[i] *= 2
		}
	}

	fmt.Printf("修改前: %v\n", numbers)
	doubler()
	fmt.Printf("修改后: %v\n", numbers)

	fmt.Println()
}

// 闭包作为返回值
func demonstrateClosureReturns() {
	fmt.Println("3. 闭包作为返回值:")

	// 工厂函数
	addMaker := func(x int) func(int) int {
		return func(y int) int {
			return x + y
		}
	}

	add5 := addMaker(5)
	add10 := addMaker(10)

	fmt.Printf("add5(3) = %d\n", add5(3))
	fmt.Printf("add10(3) = %d\n", add10(3))

	// 配置函数
	createValidator := func(min, max int) func(int) bool {
		return func(value int) bool {
			return value >= min && value <= max
		}
	}

	ageValidator := createValidator(0, 120)
	scoreValidator := createValidator(0, 100)

	fmt.Printf("年龄25有效: %t\n", ageValidator(25))
	fmt.Printf("年龄150有效: %t\n", ageValidator(150))
	fmt.Printf("分数85有效: %t\n", scoreValidator(85))
	fmt.Printf("分数150有效: %t\n", scoreValidator(150))

	// 状态管理
	createToggle := func(initial bool) func() bool {
		state := initial
		return func() bool {
			state = !state
			return state
		}
	}

	toggle := createToggle(false)
	fmt.Printf("切换状态: %t, %t, %t, %t\n",
		toggle(), toggle(), toggle(), toggle())

	// 累加器
	createAccumulator := func() func(int) int {
		sum := 0
		return func(value int) int {
			sum += value
			return sum
		}
	}

	acc := createAccumulator()
	fmt.Printf("累加器: %d, %d, %d, %d\n",
		acc(1), acc(2), acc(3), acc(4)) // 1, 3, 6, 10

	// 记忆化函数
	memoize := func(fn func(int) int) func(int) int {
		cache := make(map[int]int)
		return func(n int) int {
			if result, exists := cache[n]; exists {
				fmt.Printf("  缓存命中: f(%d) = %d\n", n, result)
				return result
			}
			result := fn(n)
			cache[n] = result
			fmt.Printf("  计算结果: f(%d) = %d\n", n, result)
			return result
		}
	}

	// 斐波那契数列的记忆化版本
	var fib func(int) int
	fib = memoize(func(n int) int {
		if n <= 1 {
			return n
		}
		return fib(n-1) + fib(n-2)
	})

	fmt.Printf("斐波那契数列 fib(10) = %d\n", fib(10))

	fmt.Println()
}

// 闭包在循环中的陷阱
func demonstrateClosureInLoops() {
	fmt.Println("4. 闭包在循环中的陷阱:")

	// 错误的方式 - 所有闭包都引用同一个变量
	fmt.Println("错误示例:")
	var badFunctions []func()
	for i := 0; i < 3; i++ {
		badFunctions = append(badFunctions, func() {
			fmt.Printf("  错误: i = %d\n", i) // 都会输出3
		})
	}

	for j, fn := range badFunctions {
		fmt.Printf("执行函数%d: ", j)
		fn()
	}

	// 正确的方式1 - 使用局部变量
	fmt.Println("\n正确示例1 - 局部变量:")
	var goodFunctions1 []func()
	for i := 0; i < 3; i++ {
		localI := i // 创建局部变量
		goodFunctions1 = append(goodFunctions1, func() {
			fmt.Printf("  正确1: i = %d\n", localI)
		})
	}

	for j, fn := range goodFunctions1 {
		fmt.Printf("执行函数%d: ", j)
		fn()
	}

	// 正确的方式2 - 使用函数参数
	fmt.Println("\n正确示例2 - 函数参数:")
	var goodFunctions2 []func()
	for i := 0; i < 3; i++ {
		goodFunctions2 = append(goodFunctions2, func(index int) func() {
			return func() {
				fmt.Printf("  正确2: i = %d\n", index)
			}
		}(i))
	}

	for j, fn := range goodFunctions2 {
		fmt.Printf("执行函数%d: ", j)
		fn()
	}

	// 正确的方式3 - 立即执行函数
	fmt.Println("\n正确示例3 - 立即执行函数:")
	var goodFunctions3 []func()
	for i := 0; i < 3; i++ {
		func(index int) {
			goodFunctions3 = append(goodFunctions3, func() {
				fmt.Printf("  正确3: i = %d\n", index)
			})
		}(i)
	}

	for j, fn := range goodFunctions3 {
		fmt.Printf("执行函数%d: ", j)
		fn()
	}

	fmt.Println()
}

// 高阶函数
func demonstrateHigherOrderFunctions() {
	fmt.Println("5. 高阶函数:")

	// Map函数
	numbers := []int{1, 2, 3, 4, 5}
	squared := mapInts(numbers, func(n int) int {
		return n * n
	})
	fmt.Printf("原数组: %v\n", numbers)
	fmt.Printf("平方后: %v\n", squared)

	// Filter函数
	evens := filterInts(numbers, func(n int) bool {
		return n%2 == 0
	})
	fmt.Printf("偶数: %v\n", evens)

	// Reduce函数
	sum := reduceInts(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	fmt.Printf("求和: %d\n", sum)

	product := reduceInts(numbers, 1, func(acc, n int) int {
		return acc * n
	})
	fmt.Printf("求积: %d\n", product)

	// 组合函数
	isEvenAndPositive := combinePredicates(
		func(n int) bool { return n%2 == 0 },
		func(n int) bool { return n > 0 },
	)

	testNumbers := []int{-2, -1, 0, 1, 2, 3, 4}
	for _, n := range testNumbers {
		if isEvenAndPositive(n) {
			fmt.Printf("%d 是正偶数\n", n)
		}
	}

	// 函数管道
	pipeline := composeFunctions(
		func(n int) int { return n + 1 }, // 加1
		func(n int) int { return n * 2 }, // 乘2
		func(n int) int { return n - 3 }, // 减3
	)

	result := pipeline(5) // ((5+1)*2)-3 = 9
	fmt.Printf("管道处理5: %d\n", result)

	// 部分应用
	multiply := func(a, b int) int { return a * b }
	multiplyBy3 := partial(multiply, 3)

	fmt.Printf("3 × 4 = %d\n", multiplyBy3(4))
	fmt.Printf("3 × 7 = %d\n", multiplyBy3(7))

	fmt.Println()
}

// 函数式编程
func demonstrateFunctionalProgramming() {
	fmt.Println("6. 函数式编程:")

	// 函数式数据处理链
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	result := chain(data).
		Filter(func(n int) bool { return n%2 == 0 }).      // 过滤偶数
		Map(func(n int) int { return n * n }).             // 平方
		Filter(func(n int) bool { return n > 10 }).        // 过滤大于10的
		Reduce(0, func(acc, n int) int { return acc + n }) // 求和

	fmt.Printf("原数据: %v\n", data)
	fmt.Printf("处理结果: %d\n", result)

	// 柯里化
	curriedAdd := curry3(func(a, b, c int) int {
		return a + b + c
	})

	add5 := curriedAdd(5)
	add5And3 := add5(3)
	finalResult := add5And3(2) // 5 + 3 + 2 = 10

	fmt.Printf("柯里化结果: %d\n", finalResult)

	// Maybe类型模拟
	maybeValue := Some(42)
	result2 := maybeValue.
		Map(func(n int) int { return n * 2 }).
		Map(func(n int) int { return n + 1 }).
		GetOrElse(0)

	fmt.Printf("Maybe结果: %d\n", result2)

	// 函数组合器
	words := []string{"hello", "world", "go", "programming"}

	processWords := func(words []string) []string {
		return chainStrings(words).
			Filter(func(s string) bool { return len(s) > 2 }).
			Map(func(s string) string { return strings.ToUpper(s) }).
			Value()
	}

	processed := processWords(words)
	fmt.Printf("单词处理: %v -> %v\n", words, processed)

	fmt.Println()
}

// 装饰器模式
func demonstrateDecoratorPattern() {
	fmt.Println("7. 装饰器模式:")

	// 基础函数
	simpleAdd := func(a, b int) int {
		return a + b
	}

	// 添加日志装饰器
	loggedAdd := withLogging("ADD", simpleAdd)
	result1 := loggedAdd(3, 4)
	fmt.Printf("装饰器结果: %d\n", result1)

	// 添加计时装饰器
	timedAdd := withTiming(simpleAdd)
	result2 := timedAdd(5, 6)
	fmt.Printf("计时结果: %d\n", result2)

	// 组合多个装饰器
	decoratedAdd := withRetry(3, withLogging("RETRY_ADD", simpleAdd))
	result3 := decoratedAdd(7, 8)
	fmt.Printf("多重装饰器结果: %d\n", result3)

	// HTTP中间件模拟
	handler := func(req string) string {
		return fmt.Sprintf("处理请求: %s", req)
	}

	decoratedHandler := withAuth(withCORS(withHTTPLogging("HTTP", handler)))
	response := decoratedHandler("GET /api/users")
	fmt.Printf("HTTP响应: %s\n", response)

	// 缓存装饰器
	expensiveFunction := func(n int) int {
		fmt.Printf("  执行昂贵计算: %d\n", n)
		time.Sleep(time.Millisecond * 100) // 模拟耗时操作
		return n * n * n
	}

	cachedFunction := withCache(expensiveFunction)

	fmt.Println("缓存装饰器测试:")
	fmt.Printf("第一次调用: %d\n", cachedFunction(5))
	fmt.Printf("第二次调用: %d\n", cachedFunction(5)) // 使用缓存
	fmt.Printf("不同参数: %d\n", cachedFunction(3))

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 事件处理系统
	fmt.Println("事件处理系统:")
	eventBus := NewEventBus()

	// 注册事件处理器
	eventBus.Subscribe("user.login", func(data interface{}) {
		user := data.(map[string]string)
		fmt.Printf("  用户登录: %s\n", user["name"])
	})

	eventBus.Subscribe("user.logout", func(data interface{}) {
		user := data.(map[string]string)
		fmt.Printf("  用户登出: %s\n", user["name"])
	})

	// 发布事件
	eventBus.Publish("user.login", map[string]string{"name": "张三"})
	eventBus.Publish("user.logout", map[string]string{"name": "张三"})

	// 2. 状态机
	fmt.Println("\n状态机:")
	sm := NewStateMachine("idle")

	sm.AddTransition("idle", "start", "running", func() {
		fmt.Println("  开始运行")
	})

	sm.AddTransition("running", "pause", "paused", func() {
		fmt.Println("  暂停运行")
	})

	sm.AddTransition("paused", "resume", "running", func() {
		fmt.Println("  恢复运行")
	})

	sm.AddTransition("running", "stop", "idle", func() {
		fmt.Println("  停止运行")
	})

	fmt.Printf("初始状态: %s\n", sm.CurrentState())
	sm.Trigger("start")
	sm.Trigger("pause")
	sm.Trigger("resume")
	sm.Trigger("stop")

	// 3. 工作流引擎
	fmt.Println("\n工作流引擎:")
	workflow := NewWorkflow()

	workflow.AddStep("验证", func(data map[string]interface{}) error {
		fmt.Println("  执行验证步骤")
		return nil
	})

	workflow.AddStep("处理", func(data map[string]interface{}) error {
		fmt.Println("  执行处理步骤")
		data["processed"] = true
		return nil
	})

	workflow.AddStep("通知", func(data map[string]interface{}) error {
		fmt.Println("  执行通知步骤")
		return nil
	})

	workflowData := map[string]interface{}{"user": "张三"}
	if err := workflow.Execute(workflowData); err != nil {
		fmt.Printf("工作流执行失败: %v\n", err)
	} else {
		fmt.Println("工作流执行成功")
	}

	// 4. 配置系统
	fmt.Println("\n配置系统:")
	config := NewConfig()

	config.Set("database.host", "localhost")
	config.Set("database.port", 5432)
	config.Set("app.debug", true)

	// 使用闭包进行类型安全的配置访问
	getDBHost := config.String("database.host", "127.0.0.1")
	getDBPort := config.Int("database.port", 3306)
	getDebug := config.Bool("app.debug", false)

	fmt.Printf("数据库配置: %s:%d\n", getDBHost(), getDBPort())
	fmt.Printf("调试模式: %t\n", getDebug())

	// 5. 插件系统
	fmt.Println("\n插件系统:")
	pluginManager := NewPluginManager()

	// 注册插件
	pluginManager.Register("logger", func(data string) string {
		fmt.Printf("  [LOG] %s\n", data)
		return data
	})

	pluginManager.Register("upperCase", func(data string) string {
		result := strings.ToUpper(data)
		fmt.Printf("  [UPPER] %s -> %s\n", data, result)
		return result
	})

	pluginManager.Register("addPrefix", func(data string) string {
		result := ">>> " + data
		fmt.Printf("  [PREFIX] %s -> %s\n", data, result)
		return result
	})

	// 执行插件链
	result := pluginManager.Process("hello world")
	fmt.Printf("最终结果: %s\n", result)

	// 6. 重试机制
	fmt.Println("\n重试机制:")

	attempts := 0
	unstableOperation := func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("操作失败 (尝试 %d)", attempts)
		}
		return nil
	}

	retryFunc := createRetryFunction(3, time.Millisecond*100)
	if err := retryFunc(unstableOperation); err != nil {
		fmt.Printf("重试失败: %v\n", err)
	} else {
		fmt.Printf("重试成功 (总共尝试 %d 次)\n", attempts)
	}

	fmt.Println()
}

// 辅助函数定义

func processWithCallback(n int, callback func(int)) {
	callback(n)
}

// 高阶函数实现
func mapInts(slice []int, fn func(int) int) []int {
	result := make([]int, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func filterInts(slice []int, predicate func(int) bool) []int {
	var result []int
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

func reduceInts(slice []int, initial int, fn func(int, int) int) int {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

func combinePredicates(predicates ...func(int) bool) func(int) bool {
	return func(n int) bool {
		for _, predicate := range predicates {
			if !predicate(n) {
				return false
			}
		}
		return true
	}
}

func composeFunctions(functions ...func(int) int) func(int) int {
	return func(n int) int {
		result := n
		for _, fn := range functions {
			result = fn(result)
		}
		return result
	}
}

func partial(fn func(int, int) int, a int) func(int) int {
	return func(b int) int {
		return fn(a, b)
	}
}

// 函数式编程工具

// 链式处理器
type IntChain struct {
	data []int
}

func chain(data []int) *IntChain {
	return &IntChain{data: data}
}

func (c *IntChain) Filter(predicate func(int) bool) *IntChain {
	c.data = filterInts(c.data, predicate)
	return c
}

func (c *IntChain) Map(fn func(int) int) *IntChain {
	c.data = mapInts(c.data, fn)
	return c
}

func (c *IntChain) Reduce(initial int, fn func(int, int) int) int {
	return reduceInts(c.data, initial, fn)
}

// 柯里化
func curry3(fn func(int, int, int) int) func(int) func(int) func(int) int {
	return func(a int) func(int) func(int) int {
		return func(b int) func(int) int {
			return func(c int) int {
				return fn(a, b, c)
			}
		}
	}
}

// Maybe类型
type Maybe struct {
	value    interface{}
	hasValue bool
}

func Some(value interface{}) *Maybe {
	return &Maybe{value: value, hasValue: true}
}

func None() *Maybe {
	return &Maybe{hasValue: false}
}

func (m *Maybe) Map(fn func(int) int) *Maybe {
	if !m.hasValue {
		return m
	}
	if v, ok := m.value.(int); ok {
		return Some(fn(v))
	}
	return None()
}

func (m *Maybe) GetOrElse(defaultValue int) int {
	if !m.hasValue {
		return defaultValue
	}
	if v, ok := m.value.(int); ok {
		return v
	}
	return defaultValue
}

// 字符串链式处理
type StringChain struct {
	data []string
}

func chainStrings(data []string) *StringChain {
	return &StringChain{data: data}
}

func (c *StringChain) Filter(predicate func(string) bool) *StringChain {
	var result []string
	for _, v := range c.data {
		if predicate(v) {
			result = append(result, v)
		}
	}
	c.data = result
	return c
}

func (c *StringChain) Map(fn func(string) string) *StringChain {
	result := make([]string, len(c.data))
	for i, v := range c.data {
		result[i] = fn(v)
	}
	c.data = result
	return c
}

func (c *StringChain) Value() []string {
	return c.data
}

// 装饰器函数
func withLogging(operation string, fn func(int, int) int) func(int, int) int {
	return func(a, b int) int {
		fmt.Printf("  [%s] 执行: %d, %d\n", operation, a, b)
		result := fn(a, b)
		fmt.Printf("  [%s] 结果: %d\n", operation, result)
		return result
	}
}

func withTiming(fn func(int, int) int) func(int, int) int {
	return func(a, b int) int {
		start := time.Now()
		result := fn(a, b)
		duration := time.Since(start)
		fmt.Printf("  [TIMING] 耗时: %v\n", duration)
		return result
	}
}

func withRetry(maxRetries int, fn func(int, int) int) func(int, int) int {
	return func(a, b int) int {
		for i := 0; i < maxRetries; i++ {
			result := fn(a, b)
			if result != 0 { // 简单的成功条件
				return result
			}
			fmt.Printf("  [RETRY] 尝试 %d/%d\n", i+1, maxRetries)
		}
		return 0
	}
}

// HTTP中间件模拟
func withHTTPLogging(operation string, handler func(string) string) func(string) string {
	return func(req string) string {
		fmt.Printf("  [%s] 处理请求: %s\n", operation, req)
		result := handler(req)
		fmt.Printf("  [%s] 响应: %s\n", operation, result)
		return result
	}
}

func withAuth(handler func(string) string) func(string) string {
	return func(req string) string {
		fmt.Printf("  [AUTH] 验证请求: %s\n", req)
		return handler(req)
	}
}

func withCORS(handler func(string) string) func(string) string {
	return func(req string) string {
		fmt.Printf("  [CORS] 添加CORS头\n")
		return handler(req)
	}
}

func withCache(fn func(int) int) func(int) int {
	cache := make(map[int]int)
	return func(n int) int {
		if result, exists := cache[n]; exists {
			fmt.Printf("  [CACHE] 缓存命中: %d\n", n)
			return result
		}
		result := fn(n)
		cache[n] = result
		return result
	}
}

// 事件总线
type EventBus struct {
	handlers map[string][]func(interface{})
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]func(interface{})),
	}
}

func (eb *EventBus) Subscribe(event string, handler func(interface{})) {
	eb.handlers[event] = append(eb.handlers[event], handler)
}

func (eb *EventBus) Publish(event string, data interface{}) {
	if handlers, exists := eb.handlers[event]; exists {
		for _, handler := range handlers {
			handler(data)
		}
	}
}

// 状态机
type StateMachine struct {
	currentState string
	transitions  map[string]map[string]struct {
		nextState string
		action    func()
	}
}

func NewStateMachine(initialState string) *StateMachine {
	return &StateMachine{
		currentState: initialState,
		transitions: make(map[string]map[string]struct {
			nextState string
			action    func()
		}),
	}
}

func (sm *StateMachine) AddTransition(from, event, to string, action func()) {
	if sm.transitions[from] == nil {
		sm.transitions[from] = make(map[string]struct {
			nextState string
			action    func()
		})
	}
	sm.transitions[from][event] = struct {
		nextState string
		action    func()
	}{to, action}
}

func (sm *StateMachine) Trigger(event string) bool {
	if transitions, exists := sm.transitions[sm.currentState]; exists {
		if transition, exists := transitions[event]; exists {
			if transition.action != nil {
				transition.action()
			}
			sm.currentState = transition.nextState
			return true
		}
	}
	return false
}

func (sm *StateMachine) CurrentState() string {
	return sm.currentState
}

// 工作流引擎
type Workflow struct {
	steps []struct {
		name string
		fn   func(map[string]interface{}) error
	}
}

func NewWorkflow() *Workflow {
	return &Workflow{}
}

func (w *Workflow) AddStep(name string, fn func(map[string]interface{}) error) {
	w.steps = append(w.steps, struct {
		name string
		fn   func(map[string]interface{}) error
	}{name, fn})
}

func (w *Workflow) Execute(data map[string]interface{}) error {
	for _, step := range w.steps {
		if err := step.fn(data); err != nil {
			return fmt.Errorf("步骤 '%s' 失败: %v", step.name, err)
		}
	}
	return nil
}

// 配置系统
type Config struct {
	data map[string]interface{}
}

func NewConfig() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}

func (c *Config) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *Config) String(key, defaultValue string) func() string {
	return func() string {
		if value, exists := c.data[key]; exists {
			if str, ok := value.(string); ok {
				return str
			}
		}
		return defaultValue
	}
}

func (c *Config) Int(key string, defaultValue int) func() int {
	return func() int {
		if value, exists := c.data[key]; exists {
			if i, ok := value.(int); ok {
				return i
			}
		}
		return defaultValue
	}
}

func (c *Config) Bool(key string, defaultValue bool) func() bool {
	return func() bool {
		if value, exists := c.data[key]; exists {
			if b, ok := value.(bool); ok {
				return b
			}
		}
		return defaultValue
	}
}

// 插件管理器
type PluginManager struct {
	plugins []struct {
		name string
		fn   func(string) string
	}
}

func NewPluginManager() *PluginManager {
	return &PluginManager{}
}

func (pm *PluginManager) Register(name string, fn func(string) string) {
	pm.plugins = append(pm.plugins, struct {
		name string
		fn   func(string) string
	}{name, fn})
}

func (pm *PluginManager) Process(data string) string {
	result := data
	for _, plugin := range pm.plugins {
		result = plugin.fn(result)
	}
	return result
}

// 重试机制
func createRetryFunction(maxRetries int, delay time.Duration) func(func() error) error {
	return func(operation func() error) error {
		var lastErr error
		for i := 0; i < maxRetries; i++ {
			if err := operation(); err != nil {
				lastErr = err
				fmt.Printf("  重试 %d/%d: %v\n", i+1, maxRetries, err)
				if i < maxRetries-1 {
					time.Sleep(delay)
				}
			} else {
				return nil
			}
		}
		return fmt.Errorf("重试 %d 次后仍失败: %v", maxRetries, lastErr)
	}
}

/*
=== 练习题 ===

1. 实现一个通用的缓存系统，支持TTL和LRU策略

2. 创建一个函数式的数据处理库，包含map、filter、reduce等

3. 实现一个简单的模板引擎，使用闭包处理变量替换

4. 设计一个中间件系统，支持HTTP请求处理

5. 创建一个异步任务调度器，使用闭包管理任务

6. 实现一个简单的依赖注入容器

7. 设计一个表达式求值器，支持变量绑定

运行命令：
go run main.go

高级练习：
1. 实现一个函数式的Promise/Future库
2. 创建一个响应式编程框架
3. 设计一个规则引擎系统
4. 实现一个简单的编译器前端
5. 创建一个分布式锁系统

重要概念：
- 闭包捕获外部变量的引用
- 闭包可以作为返回值实现工厂模式
- 注意循环中闭包的变量捕获问题
- 闭包是实现函数式编程的基础
- 装饰器模式的核心实现机制
- 支持高阶函数和函数组合
*/
