package main

import (
	"fmt"
	"runtime"
	"time"
)

/*
=== Go语言进阶特性第六课：Panic和Recover ===

学习目标：
1. 理解panic和recover的机制
2. 掌握panic的正确使用场景
3. 学会使用recover进行错误恢复
4. 了解defer在panic中的作用
5. 掌握异常安全编程

Go panic/recover特点：
- panic用于不可恢复的错误
- recover只能在defer中调用
- 类似于其他语言的异常机制
- 应该谨慎使用，优先使用error
- 主要用于库的内部实现
*/

func main() {
	fmt.Println("=== Go语言Panic和Recover学习 ===")

	// 1. 基本panic和recover
	demonstrateBasicPanicRecover()

	// 2. panic的传播机制
	demonstratePanicPropagation()

	// 3. defer和panic的交互
	demonstrateDeferAndPanic()

	// 4. recover的使用场景
	demonstrateRecoverUseCases()

	// 5. 异常安全编程
	demonstrateExceptionSafety()

	// 6. panic的最佳实践
	demonstrateBestPractices()

	// 7. 错误恢复模式
	demonstrateRecoveryPatterns()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本panic和recover
func demonstrateBasicPanicRecover() {
	fmt.Println("1. 基本panic和recover:")

	// 简单的panic和recover
	fmt.Println("简单panic恢复:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("捕获panic: %v\n", r)
			}
		}()

		fmt.Println("  开始执行")
		panic("这是一个panic")
		// fmt.Println("  这行不会执行") // unreachable code after panic
	}()
	fmt.Println("  程序继续执行")

	// 不同类型的panic值
	fmt.Println("\n不同类型的panic值:")
	panicValues := []interface{}{
		"字符串panic",
		42,
		fmt.Errorf("错误类型的panic"),
		[]int{1, 2, 3},
		nil,
	}

	for i, value := range panicValues {
		func(val interface{}) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("  panic %d (%T): %v\n", i+1, r, r)
				}
			}()
			panic(val)
		}(value)
	}

	// 条件panic
	fmt.Println("\n条件panic:")
	testValues := []int{5, 0, -1, 10}
	for _, val := range testValues {
		result := safeDivide(10, val)
		fmt.Printf("  10 / %d = %.2f\n", val, result)
	}

	fmt.Println()
}

func safeDivide(a, b int) float64 {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    恢复除法panic: %v\n", r)
		}
	}()

	if b == 0 {
		panic("除数不能为零")
	}
	if b < 0 {
		panic(fmt.Sprintf("除数不能为负数: %d", b))
	}

	return float64(a) / float64(b)
}

// panic的传播机制
func demonstratePanicPropagation() {
	fmt.Println("2. panic的传播机制:")

	// 无recover的panic传播
	fmt.Println("panic向上传播:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  在顶层捕获: %v\n", r)
			}
		}()

		level1()
	}()

	// 中间层recover
	fmt.Println("\n中间层恢复:")
	func() {
		defer func() {
			fmt.Println("  顶层defer执行")
		}()

		level1WithRecover()
		fmt.Println("  顶层继续执行")
	}()

	// 重新panic
	fmt.Println("\n重新panic:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  最终捕获: %v\n", r)
			}
		}()

		level1WithRePanic()
	}()

	fmt.Println()
}

func level1() {
	defer func() {
		fmt.Println("    level1 defer")
	}()
	fmt.Println("    进入level1")
	level2()
	fmt.Println("    level1结束") // 不会执行
}

func level2() {
	defer func() {
		fmt.Println("    level2 defer")
	}()
	fmt.Println("    进入level2")
	level3()
	fmt.Println("    level2结束") // 不会执行
}

func level3() {
	defer func() {
		fmt.Println("    level3 defer")
	}()
	fmt.Println("    进入level3")
	panic("从level3发出的panic")
	// fmt.Println("    level3结束") // unreachable code after panic
}

func level1WithRecover() {
	defer func() {
		fmt.Println("    level1WithRecover defer")
	}()
	fmt.Println("    进入level1WithRecover")
	level2WithRecover()
	fmt.Println("    level1WithRecover结束")
}

func level2WithRecover() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    在level2恢复panic: %v\n", r)
		}
		fmt.Println("    level2WithRecover defer")
	}()
	fmt.Println("    进入level2WithRecover")
	level3()
	fmt.Println("    level2WithRecover结束")
}

func level1WithRePanic() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    level1捕获并重新panic: %v\n", r)
			panic(fmt.Sprintf("重新包装: %v", r))
		}
	}()
	fmt.Println("    进入level1WithRePanic")
	level2()
	fmt.Println("    level1WithRePanic结束")
}

// defer和panic的交互
func demonstrateDeferAndPanic() {
	fmt.Println("3. defer和panic的交互:")

	// defer执行顺序
	fmt.Println("defer执行顺序:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  最终恢复: %v\n", r)
			}
		}()

		defer fmt.Println("  defer 3")
		defer fmt.Println("  defer 2")
		defer fmt.Println("  defer 1")

		panic("测试defer顺序")
	}()

	// defer中的panic
	fmt.Println("\ndefer中的panic:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  外层恢复: %v\n", r)
			}
		}()

		defer func() {
			fmt.Println("  defer函数开始")
			panic("defer中的panic")
			// fmt.Println("  defer函数结束") // unreachable code after panic
		}()

		fmt.Println("  主函数")
		panic("主函数的panic") // 会被defer中的panic覆盖
	}()

	// 多个defer中的panic
	fmt.Println("\n多个defer中的panic:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  最终panic: %v\n", r)
			}
		}()

		defer func() {
			panic("defer 1的panic")
		}()

		defer func() {
			panic("defer 2的panic") // 这个会被覆盖
		}()

		panic("主函数panic") // 这个也会被覆盖
	}()

	// defer中的资源清理
	fmt.Println("\ndefer资源清理:")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  处理panic并清理资源: %v\n", r)
			}
		}()

		resource := acquireResource()
		defer releaseResource(resource)

		processResource(resource)
	}()

	fmt.Println()
}

func acquireResource() string {
	fmt.Println("    获取资源")
	return "重要资源"
}

func releaseResource(resource string) {
	fmt.Printf("    释放资源: %s\n", resource)
}

func processResource(resource string) {
	fmt.Printf("    处理资源: %s\n", resource)
	panic("处理过程中出错")
}

// recover的使用场景
func demonstrateRecoverUseCases() {
	fmt.Println("4. recover的使用场景:")

	// 1. 网络服务器的请求处理
	fmt.Println("Web服务器请求处理:")
	server := &WebServer{}
	requests := []string{"/home", "/panic", "/error", "/api"}

	for _, path := range requests {
		server.HandleRequest(path)
	}

	// 2. 并发任务的错误隔离
	fmt.Println("\n并发任务错误隔离:")
	tasks := []func(){
		func() { fmt.Println("  任务1: 正常完成") },
		func() { panic("任务2: 发生panic") },
		func() { fmt.Println("  任务3: 正常完成") },
		func() { panic("任务4: 另一个panic") },
	}

	taskManager := &TaskManager{}
	taskManager.ExecuteTasks(tasks)

	// 3. 库函数的异常安全
	fmt.Println("\n库函数异常安全:")
	calculator := &SafeCalculator{}

	operations := []func() float64{
		func() float64 { return calculator.Divide(10, 2) },
		func() float64 { return calculator.Divide(10, 0) },
		func() float64 { return calculator.SquareRoot(-1) },
		func() float64 { return calculator.Logarithm(0) },
	}

	for i, op := range operations {
		result := calculator.SafeExecute(op)
		fmt.Printf("  操作%d结果: %.2f\n", i+1, result)
	}

	// 4. 递归函数的栈溢出保护
	fmt.Println("\n递归栈溢出保护:")
	result := safeFactorial(10)
	fmt.Printf("  factorial(10) = %d\n", result)

	result = safeFactorial(100000)
	fmt.Printf("  factorial(100000) = %d (可能溢出)\n", result)

	fmt.Println()
}

// Web服务器示例
type WebServer struct{}

func (ws *WebServer) HandleRequest(path string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  请求 %s 发生panic，已恢复: %v\n", path, r)
		}
	}()

	fmt.Printf("  处理请求: %s\n", path)

	switch path {
	case "/panic":
		panic("模拟服务器内部错误")
	case "/error":
		panic("数据库连接失败")
	default:
		fmt.Printf("    响应: 200 OK\n")
	}
}

// 任务管理器
type TaskManager struct{}

func (tm *TaskManager) ExecuteTasks(tasks []func()) {
	for i, task := range tasks {
		tm.executeTask(i+1, task)
	}
}

func (tm *TaskManager) executeTask(id int, task func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  任务%d发生panic，已隔离: %v\n", id, r)
		}
	}()

	task()
}

// 安全计算器
type SafeCalculator struct{}

func (sc *SafeCalculator) Divide(a, b float64) float64 {
	if b == 0 {
		panic("除数不能为零")
	}
	return a / b
}

func (sc *SafeCalculator) SquareRoot(x float64) float64 {
	if x < 0 {
		panic("负数不能开平方根")
	}
	// 简化实现
	return x // 实际应该计算平方根
}

func (sc *SafeCalculator) Logarithm(x float64) float64 {
	if x <= 0 {
		panic("对数的参数必须大于0")
	}
	// 简化实现
	return x // 实际应该计算对数
}

func (sc *SafeCalculator) SafeExecute(operation func() float64) float64 {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    计算错误已捕获: %v\n", r)
		}
	}()

	return operation()
}

// 安全递归
func safeFactorial(n int) int {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    递归overflow，返回-1: %v\n", r)
		}
	}()

	return factorial(n)
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

// 异常安全编程
func demonstrateExceptionSafety() {
	fmt.Println("5. 异常安全编程:")

	// 1. 资源管理
	fmt.Println("资源管理:")
	rm := &ResourceManager{}
	rm.ProcessWithResources([]string{"DB", "File", "Network"})

	// 2. 状态一致性
	fmt.Println("\n状态一致性:")
	counter := &SafeCounter{value: 10}
	fmt.Printf("  初始值: %d\n", counter.Get())

	counter.SafeUpdate(func(val int) int {
		if val > 15 {
			panic("值太大了")
		}
		return val + 5
	})
	fmt.Printf("  更新后: %d\n", counter.Get())

	counter.SafeUpdate(func(val int) int {
		return val + 10 // 这会触发panic
	})
	fmt.Printf("  panic后的值: %d\n", counter.Get())

	// 3. 事务性操作
	fmt.Println("\n事务性操作:")
	bank := &BankAccount{balance: 1000}
	fmt.Printf("  初始余额: %.2f\n", bank.GetBalance())

	bank.SafeTransfer(500)
	fmt.Printf("  转账后余额: %.2f\n", bank.GetBalance())

	bank.SafeTransfer(600) // 余额不足，会panic
	fmt.Printf("  失败转账后余额: %.2f\n", bank.GetBalance())

	fmt.Println()
}

// 资源管理器
type ResourceManager struct{}

func (rm *ResourceManager) ProcessWithResources(resources []string) {
	acquired := make([]string, 0)

	defer func() {
		// 清理所有已获取的资源
		for i := len(acquired) - 1; i >= 0; i-- {
			rm.releaseResource(acquired[i])
		}

		if r := recover(); r != nil {
			fmt.Printf("  处理完成，已清理所有资源。错误: %v\n", r)
		} else {
			fmt.Printf("  处理完成，已清理所有资源\n")
		}
	}()

	for _, resource := range resources {
		rm.acquireResource(resource)
		acquired = append(acquired, resource)

		if resource == "Network" {
			panic("网络连接失败")
		}
	}

	fmt.Printf("  所有资源处理完成: %v\n", acquired)
}

func (rm *ResourceManager) acquireResource(resource string) {
	fmt.Printf("    获取资源: %s\n", resource)
}

func (rm *ResourceManager) releaseResource(resource string) {
	fmt.Printf("    释放资源: %s\n", resource)
}

// 安全计数器
type SafeCounter struct {
	value int
}

func (sc *SafeCounter) Get() int {
	return sc.value
}

func (sc *SafeCounter) SafeUpdate(updater func(int) int) {
	oldValue := sc.value

	defer func() {
		if r := recover(); r != nil {
			sc.value = oldValue // 恢复原值
			fmt.Printf("    更新失败，恢复原值: %v\n", r)
		}
	}()

	newValue := updater(sc.value)
	sc.value = newValue
	fmt.Printf("    值更新: %d -> %d\n", oldValue, newValue)
}

// 银行账户
type BankAccount struct {
	balance float64
}

func (ba *BankAccount) GetBalance() float64 {
	return ba.balance
}

func (ba *BankAccount) SafeTransfer(amount float64) {
	originalBalance := ba.balance

	defer func() {
		if r := recover(); r != nil {
			ba.balance = originalBalance // 回滚
			fmt.Printf("    转账失败，余额回滚: %v\n", r)
		}
	}()

	if amount > ba.balance {
		panic(fmt.Sprintf("余额不足: %.2f > %.2f", amount, ba.balance))
	}

	ba.balance -= amount
	fmt.Printf("    转账成功: %.2f，剩余: %.2f\n", amount, ba.balance)
}

// panic的最佳实践
func demonstrateBestPractices() {
	fmt.Println("6. panic的最佳实践:")

	// 1. 库边界的panic转换
	fmt.Println("库边界的panic转换:")
	library := &MathLibrary{}

	// 正确使用：将内部panic转换为error
	if result, err := library.SafeDivide(10, 0); err != nil {
		fmt.Printf("  除法错误: %v\n", err)
	} else {
		fmt.Printf("  除法结果: %.2f\n", result)
	}

	// 2. 参数验证
	fmt.Println("\n参数验证:")
	validator := &ParameterValidator{}

	inputs := []interface{}{
		"valid_string",
		"",
		nil,
		42,
		[]int{1, 2, 3},
	}

	for i, input := range inputs {
		validator.SafeValidate(fmt.Sprintf("param%d", i+1), input)
	}

	// 3. 初始化检查
	fmt.Println("\n初始化检查:")
	service := &CriticalService{}
	service.SafeInitialize(map[string]string{
		"database_url": "localhost:5432",
		"api_key":      "",
	})

	// 4. 不可恢复错误的处理
	fmt.Println("\n不可恢复错误处理:")
	system := &SystemManager{}
	system.SafeStartup()

	fmt.Println()
}

// 数学库
type MathLibrary struct{}

func (ml *MathLibrary) divide(a, b float64) float64 {
	if b == 0 {
		panic("internal: division by zero")
	}
	return a / b
}

func (ml *MathLibrary) SafeDivide(a, b float64) (result float64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("数学运算错误: %v", r)
		}
	}()

	result = ml.divide(a, b)
	return result, nil
}

// 参数验证器
type ParameterValidator struct{}

func (pv *ParameterValidator) validate(name string, value interface{}) {
	if value == nil {
		panic(fmt.Sprintf("参数 %s 不能为nil", name))
	}

	if str, ok := value.(string); ok && str == "" {
		panic(fmt.Sprintf("字符串参数 %s 不能为空", name))
	}
}

func (pv *ParameterValidator) SafeValidate(name string, value interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  验证失败 %s: %v\n", name, r)
		} else {
			fmt.Printf("  验证成功 %s: %v\n", name, value)
		}
	}()

	pv.validate(name, value)
}

// 关键服务
type CriticalService struct{}

func (cs *CriticalService) initialize(config map[string]string) {
	if config["database_url"] == "" {
		panic("database_url配置缺失")
	}
	if config["api_key"] == "" {
		panic("api_key配置缺失")
	}
}

func (cs *CriticalService) SafeInitialize(config map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  服务初始化失败: %v\n", r)
		} else {
			fmt.Printf("  服务初始化成功\n")
		}
	}()

	cs.initialize(config)
}

// 系统管理器
type SystemManager struct{}

func (sm *SystemManager) checkCriticalResources() {
	// 模拟检查关键资源
	if time.Now().Second()%2 == 0 {
		panic("关键系统组件不可用")
	}
}

func (sm *SystemManager) SafeStartup() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  系统启动失败: %v\n", r)
			fmt.Printf("  系统将进入安全模式\n")
		} else {
			fmt.Printf("  系统启动成功\n")
		}
	}()

	sm.checkCriticalResources()
}

// 错误恢复模式
func demonstrateRecoveryPatterns() {
	fmt.Println("7. 错误恢复模式:")

	// 1. 重试模式
	fmt.Println("重试模式:")
	retrier := &RetryManager{}
	retrier.ExecuteWithRetry(func() {
		if time.Now().UnixNano()%3 == 0 {
			panic("随机失败")
		}
		fmt.Println("  操作成功")
	}, 3)

	// 2. 降级模式
	fmt.Println("\n降级模式:")
	service := &DegradableService{}
	result := service.GetDataWithFallback("important_data")
	fmt.Printf("  获取数据结果: %s\n", result)

	// 3. 断路器模式
	fmt.Println("\n断路器模式:")
	cb := &CircuitBreaker{threshold: 2}

	for i := 0; i < 5; i++ {
		result := cb.Execute(func() string {
			if i < 3 {
				panic("服务不可用")
			}
			return "服务正常"
		})
		fmt.Printf("  请求%d结果: %s\n", i+1, result)
	}

	fmt.Println()
}

// 重试管理器
type RetryManager struct{}

func (rm *RetryManager) ExecuteWithRetry(operation func(), maxRetries int) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if attempt < maxRetries {
						fmt.Printf("  尝试%d失败，重试: %v\n", attempt, r)
					} else {
						fmt.Printf("  尝试%d失败，放弃: %v\n", attempt, r)
					}
				}
			}()

			operation()
		}()

		// 如果执行到这里说明成功了
		if attempt <= maxRetries {
			return
		}
	}
}

// 可降级服务
type DegradableService struct{}

func (ds *DegradableService) getData(key string) string {
	if key == "important_data" {
		panic("主服务不可用")
	}
	return "主服务数据"
}

func (ds *DegradableService) getFallbackData(key string) string {
	return fmt.Sprintf("缓存数据: %s", key)
}

func (ds *DegradableService) GetDataWithFallback(key string) string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    主服务失败，使用降级: %v\n", r)
		}
	}()

	// 尝试主服务
	result := ds.getData(key)
	return result

	// 这部分不会执行，实际应该在recover中调用fallback
}

// 断路器
type CircuitBreaker struct {
	failures  int
	threshold int
	open      bool
}

func (cb *CircuitBreaker) Execute(operation func() string) string {
	if cb.open {
		return "断路器开启，服务不可用"
	}

	defer func() {
		if r := recover(); r != nil {
			cb.failures++
			if cb.failures >= cb.threshold {
				cb.open = true
				fmt.Printf("    断路器开启，失败次数: %d\n", cb.failures)
			}
		} else {
			cb.failures = 0 // 重置失败计数
		}
	}()

	return operation()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. HTTP服务器中间件
	fmt.Println("HTTP中间件:")
	middleware := &PanicRecoveryMiddleware{}

	handlers := []func(){
		func() { fmt.Println("  处理正常请求") },
		func() { panic("数据库连接失败") },
		func() { panic("内存不足") },
		func() { fmt.Println("  处理另一个正常请求") },
	}

	for i, handler := range handlers {
		middleware.HandleRequest(fmt.Sprintf("request-%d", i+1), handler)
	}

	// 2. 协程池
	fmt.Println("\n协程池:")
	pool := &WorkerPool{}

	jobs := []func(){
		func() { fmt.Println("  工作1完成") },
		func() { panic("工作2失败") },
		func() { time.Sleep(time.Millisecond * 100); fmt.Println("  工作3完成") },
		func() { panic("工作4失败") },
	}

	pool.ExecuteJobs(jobs)

	// 3. 插件系统
	fmt.Println("\n插件系统:")
	pluginManager := &PluginManager{}

	plugins := []Plugin{
		&GoodPlugin{name: "logger"},
		&BadPlugin{name: "crasher"},
		&GoodPlugin{name: "monitor"},
	}

	for _, plugin := range plugins {
		pluginManager.SafeExecutePlugin(plugin)
	}

	// 4. 数据处理管道
	fmt.Println("\n数据处理管道:")
	pipeline := &DataPipeline{}

	data := []interface{}{42, "hello", nil, 3.14, "world"}
	results := pipeline.Process(data)
	fmt.Printf("  处理结果: %v\n", results)

	fmt.Println()
}

// HTTP中间件
type PanicRecoveryMiddleware struct{}

func (prm *PanicRecoveryMiddleware) HandleRequest(requestID string, handler func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  请求 %s panic已恢复: %v\n", requestID, r)
			// 记录错误日志
			fmt.Printf("    错误已记录到日志系统\n")
			// 返回500错误给客户端
			fmt.Printf("    返回HTTP 500给客户端\n")
		} else {
			fmt.Printf("  请求 %s 处理成功\n", requestID)
		}
	}()

	handler()
}

// 工作池
type WorkerPool struct{}

func (wp *WorkerPool) ExecuteJobs(jobs []func()) {
	for i, job := range jobs {
		wp.executeJob(i+1, job)
	}
}

func (wp *WorkerPool) executeJob(id int, job func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  工作%d发生panic，工作池继续运行: %v\n", id, r)
		}
	}()

	job()
}

// 插件接口
type Plugin interface {
	Execute()
	Name() string
}

type GoodPlugin struct {
	name string
}

func (gp *GoodPlugin) Execute() {
	fmt.Printf("    插件 %s 执行成功\n", gp.name)
}

func (gp *GoodPlugin) Name() string {
	return gp.name
}

type BadPlugin struct {
	name string
}

func (bp *BadPlugin) Execute() {
	panic(fmt.Sprintf("插件 %s 执行失败", bp.name))
}

func (bp *BadPlugin) Name() string {
	return bp.name
}

// 插件管理器
type PluginManager struct{}

func (pm *PluginManager) SafeExecutePlugin(plugin Plugin) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  插件 %s 执行失败: %v\n", plugin.Name(), r)
		} else {
			fmt.Printf("  插件 %s 执行成功\n", plugin.Name())
		}
	}()

	plugin.Execute()
}

// 数据处理管道
type DataPipeline struct{}

func (dp *DataPipeline) Process(data []interface{}) []interface{} {
	var results []interface{}

	for i, item := range data {
		result := dp.processItem(i+1, item)
		if result != nil {
			results = append(results, result)
		}
	}

	return results
}

func (dp *DataPipeline) processItem(id int, item interface{}) interface{} {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("    项目%d处理失败: %v\n", id, r)
		}
	}()

	if item == nil {
		panic("不能处理nil数据")
	}

	// 处理不同类型的数据
	switch v := item.(type) {
	case int:
		return v * 2
	case string:
		return fmt.Sprintf("processed_%s", v)
	case float64:
		return v + 1.0
	default:
		panic(fmt.Sprintf("不支持的数据类型: %T", v))
	}
}

/*
=== 练习题 ===

1. 实现一个异常安全的并发队列

2. 创建一个支持panic恢复的RPC服务器

3. 设计一个数据库事务管理器，支持panic回滚

4. 实现一个Web爬虫，具有错误隔离机制

5. 创建一个任务调度器，支持任务失败恢复

6. 设计一个插件化系统的异常处理机制

7. 实现一个分布式系统的故障隔离组件

运行命令：
go run main.go

高级练习：
1. 实现基于panic的轻量级异常系统
2. 创建panic的性能监控和分析工具
3. 设计panic的分布式传播机制
4. 实现基于panic的错误恢复策略
5. 创建panic安全的内存管理系统

重要概念：
- panic用于不可恢复的错误
- recover只能在defer中使用
- panic会终止当前函数执行
- defer函数总是会执行
- 异常安全编程的重要性
- 合理使用panic和error的边界
*/

func init() {
	// 设置全局panic处理
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("全局panic恢复: %v\n", r)
			fmt.Printf("调用栈:\n")

			// 打印调用栈
			buf := make([]byte, 1024)
			for {
				n := runtime.Stack(buf, false)
				if n < len(buf) {
					fmt.Printf("%s", buf[:n])
					break
				}
				buf = make([]byte, 2*len(buf))
			}
		}
	}()
}
