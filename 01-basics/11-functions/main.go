package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Person 结构体定义
type Person struct {
	Name string
	Age  int
}

/*
=== Go语言第十一课：函数(Functions) ===

学习目标：
1. 掌握函数的定义和调用
2. 理解参数传递机制
3. 学会多返回值的使用
4. 了解可变参数函数
5. 掌握函数作为值的使用

Go函数特点：
- 一等公民，可以作为值传递
- 支持多返回值
- 支持可变参数
- 支持命名返回值
- 支持defer语句
*/

func main() {
	fmt.Println("=== Go语言函数学习 ===")

	// 1. 基本函数定义和调用
	demonstrateBasicFunctions()

	// 2. 参数传递
	demonstrateParameterPassing()

	// 3. 多返回值
	demonstrateMultipleReturns()

	// 4. 可变参数
	demonstrateVariadicFunctions()

	// 5. 命名返回值
	demonstrateNamedReturns()

	// 6. 函数作为值
	demonstrateFunctionsAsValues()

	// 7. defer语句
	demonstrateDefer()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本函数定义和调用
func demonstrateBasicFunctions() {
	fmt.Println("1. 基本函数定义和调用:")

	// 调用无参数函数
	sayHello()

	// 调用有参数函数
	greet("张三")
	greetWithAge("李四", 25)

	// 调用有返回值函数
	sum := add(10, 20)
	fmt.Printf("10 + 20 = %d\n", sum)

	// 调用有多个参数和返回值的函数
	quotient, remainder := divide(17, 5)
	fmt.Printf("17 ÷ 5 = %d 余 %d\n", quotient, remainder)

	// 调用不同类型参数的函数
	area := calculateCircleArea(5.0)
	fmt.Printf("半径5的圆面积: %.2f\n", area)

	// 布尔返回值
	if isEven(42) {
		fmt.Println("42是偶数")
	}

	if !isEven(13) {
		fmt.Println("13是奇数")
	}

	// 字符串处理函数
	reversed := reverseString("Hello")
	fmt.Printf("'Hello'反转后: '%s'\n", reversed)

	fmt.Println()
}

// 参数传递
func demonstrateParameterPassing() {
	fmt.Println("2. 参数传递:")

	// 值传递（基本类型）
	x := 10
	fmt.Printf("调用前x: %d\n", x)
	modifyValue(x)
	fmt.Printf("调用后x: %d\n", x) // 不会改变

	// 值传递（复合类型-数组）
	arr := [3]int{1, 2, 3}
	fmt.Printf("调用前数组: %v\n", arr)
	modifyArray(arr)
	fmt.Printf("调用后数组: %v\n", arr) // 不会改变

	// 引用传递（指针）
	y := 20
	fmt.Printf("调用前y: %d\n", y)
	modifyPointer(&y)
	fmt.Printf("调用后y: %d\n", y) // 会改变

	// 引用传递（切片）
	slice := []int{1, 2, 3}
	fmt.Printf("调用前切片: %v\n", slice)
	modifySlice(slice)
	fmt.Printf("调用后切片: %v\n", slice) // 元素会改变

	// 引用传递（映射）
	m := map[string]int{"a": 1, "b": 2}
	fmt.Printf("调用前映射: %v\n", m)
	modifyMap(m)
	fmt.Printf("调用后映射: %v\n", m) // 会改变

	// 结构体传递
	person := Person{"Alice", 25}
	fmt.Printf("调用前结构体: %+v\n", person)
	modifyStruct(person)
	fmt.Printf("调用后结构体: %+v\n", person) // 不会改变（值传递）

	modifyStructPointer(&person)
	fmt.Printf("指针修改后结构体: %+v\n", person) // 会改变

	fmt.Println()
}

// 多返回值
func demonstrateMultipleReturns() {
	fmt.Println("3. 多返回值:")

	// 基本多返回值
	min, max := findMinMax([]int{3, 7, 1, 9, 4})
	fmt.Printf("数组[3,7,1,9,4]的最小值: %d, 最大值: %d\n", min, max)

	// 错误处理模式
	result, err := safeDivide(10, 2)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("10 ÷ 2 = %.2f\n", result)
	}

	result, err = safeDivide(10, 0)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %.2f\n", result)
	}

	// 忽略某些返回值
	_, max2 := findMinMax([]int{5, 2, 8, 1})
	fmt.Printf("只关心最大值: %d\n", max2)

	// ok模式
	value, ok := lookup(map[string]int{"apple": 5, "banana": 3}, "apple")
	if ok {
		fmt.Printf("找到apple: %d\n", value)
	}

	value, ok = lookup(map[string]int{"apple": 5, "banana": 3}, "orange")
	if !ok {
		fmt.Printf("未找到orange\n")
	}

	// 多种类型的返回值
	name, age, isStudent := getPersonInfo()
	fmt.Printf("个人信息: %s, %d岁, 学生: %t\n", name, age, isStudent)

	// 字符串处理多返回值
	words, count := splitAndCount("hello world go programming")
	fmt.Printf("分词结果: %v, 单词数: %d\n", words, count)

	fmt.Println()
}

// 可变参数
func demonstrateVariadicFunctions() {
	fmt.Println("4. 可变参数:")

	// 基本可变参数
	sum1 := sum(1, 2, 3)
	fmt.Printf("sum(1, 2, 3) = %d\n", sum1)

	sum2 := sum(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	fmt.Printf("sum(1...10) = %d\n", sum2)

	// 没有参数
	sum3 := sum()
	fmt.Printf("sum() = %d\n", sum3)

	// 传递切片
	numbers := []int{2, 4, 6, 8, 10}
	sum4 := sum(numbers...) // 展开切片
	fmt.Printf("sum(%v) = %d\n", numbers, sum4)

	// 混合参数（固定+可变）
	result := format("姓名: %s, 年龄: %d, 分数: %d, %d, %d", "张三", 20, 85, 92, 78)
	fmt.Printf("格式化结果: %s\n", result)

	// 字符串连接
	joined := join("-", "apple", "banana", "cherry")
	fmt.Printf("连接结果: %s\n", joined)

	// 空参数情况
	empty := join("-")
	fmt.Printf("空参数结果: '%s'\n", empty)

	// 数学函数
	maxVal := max(3, 7, 1, 9, 4, 2)
	fmt.Printf("最大值: %d\n", maxVal)

	minVal := min(3, 7, 1, 9, 4, 2)
	fmt.Printf("最小值: %d\n", minVal)

	// 不同类型的可变参数
	info := formatInfo("用户信息", "姓名", "张三", "年龄", 25, "城市", "北京")
	fmt.Printf("信息: %s\n", info)

	fmt.Println()
}

// 命名返回值
func demonstrateNamedReturns() {
	fmt.Println("5. 命名返回值:")

	// 基本命名返回值
	a, p := rectangleAreaAndPerimeter(5, 3)
	fmt.Printf("长方形(5x3) 面积: %d, 周长: %d\n", a, p)

	// 错误处理中的命名返回值
	content, err := readFile("config.txt")
	if err != nil {
		fmt.Printf("读取文件错误: %v\n", err)
	} else {
		fmt.Printf("文件内容: %s\n", content)
	}

	// 复杂计算的命名返回值
	mean, variance, stdDev := statistics([]float64{1, 2, 3, 4, 5})
	fmt.Printf("统计: 均值=%.2f, 方差=%.2f, 标准差=%.2f\n", mean, variance, stdDev)

	// 多个处理步骤的命名返回值
	cleaned, validated, processed := processData("  Hello World!  ")
	fmt.Printf("数据处理: 清理='%s', 验证=%t, 处理='%s'\n", cleaned, validated, processed)

	// 业务逻辑中的命名返回值
	discount, finalPrice, savings := calculatePrice(1000, "VIP")
	fmt.Printf("价格计算: 折扣=%.0f%%, 最终价格=%.2f, 节省=%.2f\n",
		discount*100, finalPrice, savings)

	fmt.Println()
}

// 函数作为值
func demonstrateFunctionsAsValues() {
	fmt.Println("6. 函数作为值:")

	// 函数变量
	var mathOp func(int, int) int

	mathOp = add
	fmt.Printf("使用add函数: 5 + 3 = %d\n", mathOp(5, 3))

	mathOp = multiply
	fmt.Printf("使用multiply函数: 5 * 3 = %d\n", mathOp(5, 3))

	// 函数作为参数
	numbers := []int{1, 2, 3, 4, 5}

	doubled := applyToSlice(numbers, double)
	fmt.Printf("数组翻倍: %v -> %v\n", numbers, doubled)

	squared := applyToSlice(numbers, square)
	fmt.Printf("数组平方: %v -> %v\n", numbers, squared)

	// 函数映射
	operations := map[string]func(int, int) int{
		"add":      add,
		"subtract": subtract,
		"multiply": multiply,
	}

	for name, op := range operations {
		result := op(10, 3)
		fmt.Printf("%s(10, 3) = %d\n", name, result)
	}

	// 函数切片
	filters := []func(int) bool{
		isEven,
		isPositive,
		func(n int) bool { return n > 10 }, // 匿名函数
	}

	testNumber := 12
	fmt.Printf("测试数字 %d:\n", testNumber)
	for i, filter := range filters {
		if filter(testNumber) {
			fmt.Printf("  通过过滤器 %d\n", i)
		}
	}

	// 高阶函数
	increment := makeAdder(1)
	addTen := makeAdder(10)

	fmt.Printf("increment(5) = %d\n", increment(5))
	fmt.Printf("addTen(5) = %d\n", addTen(5))

	// 函数组合
	addOne := func(x int) int { return x + 1 }
	multiplyTwo := func(x int) int { return x * 2 }

	composed := compose(multiplyTwo, addOne)
	fmt.Printf("compose(multiplyTwo, addOne)(5) = %d\n", composed(5)) // (5+1)*2 = 12

	fmt.Println()
}

// defer语句
func demonstrateDefer() {
	fmt.Println("7. defer语句:")

	// 基本defer
	fmt.Println("开始执行")
	defer fmt.Println("defer: 函数结束")
	fmt.Println("中间执行")

	// 多个defer（后进先出）
	deferOrder()

	// defer在循环中
	deferInLoop()

	// defer捕获变量
	deferVariableCapture()

	// defer用于资源清理
	processFile("data.txt")

	// defer用于错误恢复
	safeDivideWithRecover(10, 0)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 数据验证
	fmt.Println("数据验证:")
	users := []map[string]interface{}{
		{"name": "张三", "age": 25, "email": "zhang@example.com"},
		{"name": "", "age": 15, "email": "invalid-email"},
		{"name": "李四", "age": 30, "email": "li@example.com"},
	}

	for i, user := range users {
		if errors := validateUser(user); len(errors) > 0 {
			fmt.Printf("  用户%d验证失败: %v\n", i+1, errors)
		} else {
			fmt.Printf("  用户%d验证通过\n", i+1)
		}
	}

	// 2. 数据转换管道
	fmt.Println("\n数据转换管道:")
	rawData := []string{"1", "2", "3", "4", "5"}

	processed := pipeline(rawData,
		func(s []string) []string {
			fmt.Printf("  步骤1-原始数据: %v\n", s)
			return s
		},
		func(s []string) []string {
			result := make([]string, len(s))
			for i, v := range s {
				result[i] = "num_" + v
			}
			fmt.Printf("  步骤2-添加前缀: %v\n", result)
			return result
		},
		func(s []string) []string {
			result := make([]string, len(s))
			for i, v := range s {
				result[i] = strings.ToUpper(v)
			}
			fmt.Printf("  步骤3-转大写: %v\n", result)
			return result
		},
	)

	fmt.Printf("  最终结果: %v\n", processed)

	// 3. 重试机制
	fmt.Println("\n重试机制:")

	// 模拟不稳定的网络请求
	attempts := 0
	success := retry(3, func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("网络错误 (尝试 %d)", attempts)
		}
		return nil
	})

	if success {
		fmt.Printf("  请求成功 (总共尝试 %d 次)\n", attempts)
	} else {
		fmt.Printf("  请求失败 (总共尝试 %d 次)\n", attempts)
	}

	// 4. 缓存装饰器
	fmt.Println("\n缓存装饰器:")

	// 创建带缓存的斐波那契函数
	cachedFib := memoize(fibonacci)

	start := time.Now()
	result1 := cachedFib(40)
	duration1 := time.Since(start)

	start = time.Now()
	result2 := cachedFib(40) // 第二次调用，使用缓存
	duration2 := time.Since(start)

	fmt.Printf("  第一次计算fib(40): %d (耗时: %v)\n", result1, duration1)
	fmt.Printf("  第二次计算fib(40): %d (耗时: %v)\n", result2, duration2)

	// 5. 函数选项模式
	fmt.Println("\n函数选项模式:")

	client := NewHTTPClient(
		WithTimeout(30),
		WithRetries(3),
		WithUserAgent("MyApp/1.0"),
	)

	fmt.Printf("  HTTP客户端配置: %+v\n", client)

	// 6. 管道模式
	fmt.Println("\n管道模式:")

	// 数字处理管道
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	result := processNumbers(input,
		filterFunc(func(n int) bool { return n%2 == 0 }), // 过滤偶数
		mapFunc(func(n int) int { return n * n }),        // 平方
		mapFunc(func(n int) int { return n + 1 }),        // 加1
	)

	fmt.Printf("  输入: %v\n", input)
	fmt.Printf("  处理结果: %v\n", result)

	fmt.Println()
}

// 辅助函数定义

// 基本函数
func sayHello() {
	fmt.Println("Hello, World!")
}

func greet(name string) {
	fmt.Printf("你好, %s!\n", name)
}

func greetWithAge(name string, age int) {
	fmt.Printf("你好, %s! 你今年%d岁。\n", name, age)
}

func add(a, b int) int {
	return a + b
}

func divide(a, b int) (int, int) {
	return a / b, a % b
}

func calculateCircleArea(radius float64) float64 {
	return math.Pi * radius * radius
}

func isEven(n int) bool {
	return n%2 == 0
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// 参数传递相关函数
func modifyValue(x int) {
	x = 999
}

func modifyArray(arr [3]int) {
	arr[0] = 999
}

func modifyPointer(x *int) {
	*x = 999
}

func modifySlice(slice []int) {
	slice[0] = 999
}

func modifyMap(m map[string]int) {
	m["c"] = 999
}

func modifyStruct(p struct {
	Name string
	Age  int
}) {
	p.Name = "Modified"
	p.Age = 999
}

func modifyStructPointer(p *Person) {
	p.Name = "Modified"
	p.Age = 999
}

// 多返回值函数
func findMinMax(numbers []int) (int, int) {
	if len(numbers) == 0 {
		return 0, 0
	}

	min, max := numbers[0], numbers[0]
	for _, num := range numbers {
		if num < min {
			min = num
		}
		if num > max {
			max = num
		}
	}
	return min, max
}

func safeDivide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为零")
	}
	return a / b, nil
}

func lookup(m map[string]int, key string) (int, bool) {
	value, exists := m[key]
	return value, exists
}

func getPersonInfo() (string, int, bool) {
	return "张同学", 20, true
}

func splitAndCount(text string) ([]string, int) {
	words := strings.Fields(text)
	return words, len(words)
}

// 可变参数函数
func sum(numbers ...int) int {
	total := 0
	for _, num := range numbers {
		total += num
	}
	return total
}

func format(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}

func join(separator string, parts ...string) string {
	return strings.Join(parts, separator)
}

func max(numbers ...int) int {
	if len(numbers) == 0 {
		return 0
	}
	max := numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
	}
	return max
}

func min(numbers ...int) int {
	if len(numbers) == 0 {
		return 0
	}
	min := numbers[0]
	for _, num := range numbers {
		if num < min {
			min = num
		}
	}
	return min
}

func formatInfo(title string, keyValues ...interface{}) string {
	result := title + ": "
	for i := 0; i < len(keyValues); i += 2 {
		if i+1 < len(keyValues) {
			result += fmt.Sprintf("%v=%v ", keyValues[i], keyValues[i+1])
		}
	}
	return result
}

// 命名返回值函数
func rectangleAreaAndPerimeter(length, width int) (area, perimeter int) {
	area = length * width
	perimeter = 2 * (length + width)
	return // 裸返回
}

func readFile(filename string) (content string, err error) {
	// 模拟文件读取
	if filename == "" {
		err = errors.New("文件名不能为空")
		return
	}

	if strings.HasSuffix(filename, ".txt") {
		content = "文件内容示例"
	} else {
		err = errors.New("不支持的文件类型")
	}
	return
}

func statistics(data []float64) (mean, variance, stdDev float64) {
	if len(data) == 0 {
		return
	}

	// 计算均值
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	mean = sum / float64(len(data))

	// 计算方差
	sumSquares := 0.0
	for _, v := range data {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance = sumSquares / float64(len(data))

	// 计算标准差
	stdDev = math.Sqrt(variance)
	return
}

func processData(input string) (cleaned string, validated bool, processed string) {
	cleaned = strings.TrimSpace(input)
	validated = len(cleaned) > 0
	if validated {
		processed = strings.ToLower(cleaned)
	}
	return
}

func calculatePrice(originalPrice float64, customerType string) (discount, finalPrice, savings float64) {
	switch customerType {
	case "VIP":
		discount = 0.2
	case "Member":
		discount = 0.1
	default:
		discount = 0.0
	}

	savings = originalPrice * discount
	finalPrice = originalPrice - savings
	return
}

// 函数作为值相关
func multiply(a, b int) int {
	return a * b
}

func subtract(a, b int) int {
	return a - b
}

func double(n int) int {
	return n * 2
}

func square(n int) int {
	return n * n
}

func isPositive(n int) bool {
	return n > 0
}

func applyToSlice(slice []int, fn func(int) int) []int {
	result := make([]int, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func makeAdder(x int) func(int) int {
	return func(y int) int {
		return x + y
	}
}

func compose(f, g func(int) int) func(int) int {
	return func(x int) int {
		return f(g(x))
	}
}

// defer相关函数
func deferOrder() {
	fmt.Println("开始deferOrder")
	defer fmt.Println("defer 1")
	defer fmt.Println("defer 2")
	defer fmt.Println("defer 3")
	fmt.Println("结束deferOrder")
}

func deferInLoop() {
	fmt.Println("defer在循环中:")
	for i := 0; i < 3; i++ {
		defer fmt.Printf("  defer循环 %d\n", i)
	}
}

func deferVariableCapture() {
	fmt.Println("defer变量捕获:")
	x := 1
	defer func() { fmt.Printf("  defer中x: %d\n", x) }()
	x = 2
	defer func(val int) { fmt.Printf("  defer参数x: %d\n", val) }(x)
	x = 3
}

func processFile(filename string) {
	fmt.Printf("处理文件: %s\n", filename)
	defer fmt.Printf("  清理资源: %s\n", filename)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  恢复错误: %v\n", r)
		}
	}()

	// 模拟文件处理
	fmt.Printf("  读取文件: %s\n", filename)
}

func safeDivideWithRecover(a, b int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recover捕获错误: %v\n", r)
		}
	}()

	if b == 0 {
		panic("除数为零")
	}

	result := a / b
	fmt.Printf("安全除法结果: %d\n", result)
}

// 实际应用函数
func validateUser(user map[string]interface{}) []string {
	var errors []string

	name, ok := user["name"].(string)
	if !ok || name == "" {
		errors = append(errors, "姓名不能为空")
	}

	age, ok := user["age"].(int)
	if !ok || age < 18 {
		errors = append(errors, "年龄必须大于等于18")
	}

	email, ok := user["email"].(string)
	if !ok || !strings.Contains(email, "@") {
		errors = append(errors, "邮箱格式无效")
	}

	return errors
}

func pipeline(data []string, steps ...func([]string) []string) []string {
	result := data
	for _, step := range steps {
		result = step(result)
	}
	return result
}

func retry(maxAttempts int, fn func() error) bool {
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return true
		} else {
			fmt.Printf("  尝试 %d 失败: %v\n", i+1, err)
		}
	}
	return false
}

func memoize(fn func(int) int) func(int) int {
	cache := make(map[int]int)
	return func(n int) int {
		if result, exists := cache[n]; exists {
			return result
		}
		result := fn(n)
		cache[n] = result
		return result
	}
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// 函数选项模式
type HTTPClient struct {
	Timeout   int
	Retries   int
	UserAgent string
}

type ClientOption func(*HTTPClient)

func WithTimeout(timeout int) ClientOption {
	return func(c *HTTPClient) {
		c.Timeout = timeout
	}
}

func WithRetries(retries int) ClientOption {
	return func(c *HTTPClient) {
		c.Retries = retries
	}
}

func WithUserAgent(userAgent string) ClientOption {
	return func(c *HTTPClient) {
		c.UserAgent = userAgent
	}
}

func NewHTTPClient(options ...ClientOption) *HTTPClient {
	client := &HTTPClient{
		Timeout:   10,
		Retries:   1,
		UserAgent: "DefaultAgent",
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// 管道模式
type NumberProcessor func([]int) []int

func filterFunc(predicate func(int) bool) NumberProcessor {
	return func(numbers []int) []int {
		var result []int
		for _, n := range numbers {
			if predicate(n) {
				result = append(result, n)
			}
		}
		return result
	}
}

func mapFunc(transform func(int) int) NumberProcessor {
	return func(numbers []int) []int {
		result := make([]int, len(numbers))
		for i, n := range numbers {
			result[i] = transform(n)
		}
		return result
	}
}

func processNumbers(numbers []int, processors ...NumberProcessor) []int {
	result := numbers
	for _, processor := range processors {
		result = processor(result)
	}
	return result
}

/*
=== 练习题 ===

1. 编写一个函数，计算任意数量数字的平均值

2. 实现一个通用的排序函数，接受比较函数作为参数

3. 编写一个递归函数，计算阶乘

4. 实现一个函数，返回字符串中每个字符的出现次数

5. 编写一个高阶函数，实现函数的柯里化

6. 实现一个简单的中间件系统

7. 编写一个函数，实现函数的组合链

运行命令：
go run main.go

高级练习：
1. 实现一个函数式编程的map/filter/reduce
2. 编写一个简单的表达式求值器
3. 实现一个函数缓存系统
4. 创建一个事件处理系统
5. 实现一个简单的状态机

重要概念：
- 函数是一等公民
- 支持多返回值
- defer延迟执行
- 可变参数使用...
- 命名返回值可以裸返回
- 函数可以作为值传递和存储
*/
