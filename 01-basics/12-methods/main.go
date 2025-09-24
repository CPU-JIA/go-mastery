package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

/*
=== Go语言第十二课：方法(Methods) ===

学习目标：
1. 理解方法与函数的区别
2. 掌握值接收者和指针接收者
3. 学会为自定义类型定义方法
4. 了解方法集和接口实现
5. 掌握方法的实际应用

Go方法特点：
- 方法是带有接收者的函数
- 可以为任何自定义类型定义方法
- 支持值接收者和指针接收者
- 方法集决定接口实现
- 可以链式调用
*/

func main() {
	fmt.Println("=== Go语言方法学习 ===")

	// 1. 基本方法定义
	demonstrateBasicMethods()

	// 2. 值接收者vs指针接收者
	demonstrateReceivers()

	// 3. 为不同类型定义方法
	demonstrateMethodsOnTypes()

	// 4. 方法集和接口
	demonstrateMethodSets()

	// 5. 嵌入类型的方法
	demonstrateEmbeddedMethods()

	// 6. 方法链式调用
	demonstrateMethodChaining()

	// 7. 方法作为值
	demonstrateMethodValues()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本结构体和方法定义
type Rectangle struct {
	Width, Height float64
}

// 值接收者方法
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// 指针接收者方法
func (r *Rectangle) Scale(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

func (r *Rectangle) SetDimensions(width, height float64) {
	r.Width = width
	r.Height = height
}

// String方法（实现fmt.Stringer接口）
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.1f×%.1f)", r.Width, r.Height)
}

// 基本方法定义
func demonstrateBasicMethods() {
	fmt.Println("1. 基本方法定义:")

	rect := Rectangle{Width: 10, Height: 5}
	fmt.Printf("矩形: %s\n", rect)

	// 调用值接收者方法
	area := rect.Area()
	perimeter := rect.Perimeter()

	fmt.Printf("面积: %.2f\n", area)
	fmt.Printf("周长: %.2f\n", perimeter)

	// 调用指针接收者方法
	fmt.Printf("缩放前: %s\n", rect)
	rect.Scale(2.0)
	fmt.Printf("缩放2倍后: %s\n", rect)

	rect.SetDimensions(6, 8)
	fmt.Printf("设置新尺寸后: %s\n", rect)

	fmt.Println()
}

// Circle类型
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Circumference() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) Perimeter() float64 {
	return c.Circumference() // 圆的周长就是它的边界长度
}

func (c *Circle) SetRadius(radius float64) {
	if radius > 0 {
		c.Radius = radius
	}
}

func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.1f)", c.Radius)
}

// 值接收者vs指针接收者
func demonstrateReceivers() {
	fmt.Println("2. 值接收者vs指针接收者:")

	// 值接收者 - 不会修改原值
	rect1 := Rectangle{Width: 3, Height: 4}
	fmt.Printf("原始矩形: %s\n", rect1)

	// 值接收者方法可以通过值或指针调用
	area1 := rect1.Area()
	area2 := (&rect1).Area()
	fmt.Printf("值调用面积: %.2f, 指针调用面积: %.2f\n", area1, area2)

	// 指针接收者 - 会修改原值
	fmt.Printf("修改前: %s\n", rect1)
	rect1.Scale(1.5) // Go自动取地址
	fmt.Printf("修改后: %s\n", rect1)

	// 指针接收者方法也可以通过值或指针调用
	rect2 := Rectangle{Width: 2, Height: 3}
	pRect2 := &rect2

	rect2.SetDimensions(5, 7)   // 值调用，Go自动取地址
	pRect2.SetDimensions(8, 10) // 指针调用
	fmt.Printf("最终矩形: %s\n", rect2)

	// 方法接收者的选择原则演示
	demonstrateReceiverGuidelines()

	fmt.Println()
}

// 接收者选择指导
func demonstrateReceiverGuidelines() {
	fmt.Println("\n接收者选择指导:")

	// 1. 需要修改接收者 - 使用指针接收者
	counter := &Counter{value: 0}
	fmt.Printf("计数器初始值: %d\n", counter.Value())
	counter.Increment()
	counter.Add(5)
	fmt.Printf("计数器最终值: %d\n", counter.Value())

	// 2. 大结构体 - 使用指针接收者避免复制
	large := &LargeStruct{}
	large.Process() // 避免复制大结构体

	// 3. 一致性 - 如果有指针接收者，其他方法也应该使用指针接收者
	person := &Person{Name: "张三", Age: 25}
	fmt.Printf("原始: %s\n", person.Info())
	person.Birthday()
	person.ChangeName("李四")
	fmt.Printf("修改后: %s\n", person.Info())
}

type Counter struct {
	value int
}

func (c *Counter) Increment() {
	c.value++
}

func (c *Counter) Add(n int) {
	c.value += n
}

func (c *Counter) Value() int {
	return c.value
}

type LargeStruct struct {
	data [1000]int
}

func (l *LargeStruct) Process() {
	// 处理大结构体，使用指针避免复制
	fmt.Println("处理大结构体（避免复制）")
}

type Person struct {
	Name string
	Age  int
}

func (p *Person) Birthday() {
	p.Age++
}

func (p *Person) ChangeName(name string) {
	p.Name = name
}

func (p *Person) Info() string {
	return fmt.Sprintf("%s (%d岁)", p.Name, p.Age)
}

// 为不同类型定义方法
func demonstrateMethodsOnTypes() {
	fmt.Println("3. 为不同类型定义方法:")

	// 为基本类型的别名定义方法
	var temp Temperature = 25.5
	fmt.Printf("温度: %.1f°C = %.1f°F = %.1fK\n",
		float64(temp), temp.ToFahrenheit(), temp.ToKelvin())

	// 为切片类型定义方法
	numbers := IntSlice{1, 5, 3, 9, 2, 7}
	fmt.Printf("原始切片: %v\n", numbers)
	fmt.Printf("总和: %d\n", numbers.Sum())
	fmt.Printf("平均值: %.2f\n", numbers.Average())
	fmt.Printf("最大值: %d\n", numbers.Max())
	fmt.Printf("最小值: %d\n", numbers.Min())

	numbers.Sort()
	fmt.Printf("排序后: %v\n", numbers)

	// 为映射类型定义方法
	scores := StudentScores{
		"张三": 85,
		"李四": 92,
		"王五": 78,
	}

	fmt.Printf("学生成绩: %v\n", map[string]int(scores))
	fmt.Printf("平均分: %.2f\n", scores.Average())
	fmt.Printf("最高分学生: %s\n", scores.TopStudent())
	scores.AddBonus(5)
	fmt.Printf("加分后: %v\n", map[string]int(scores))

	// 为函数类型定义方法
	var processor DataProcessor = func(data []int) []int {
		result := make([]int, len(data))
		for i, v := range data {
			result[i] = v * 2
		}
		return result
	}

	input := []int{1, 2, 3, 4, 5}
	output := processor.Process(input)
	fmt.Printf("数据处理: %v -> %v\n", input, output)

	validated := processor.ValidateAndProcess(input)
	fmt.Printf("验证并处理: %v -> %v\n", input, validated)

	fmt.Println()
}

// 自定义类型
type Temperature float64

func (t Temperature) ToFahrenheit() float64 {
	return float64(t)*9/5 + 32
}

func (t Temperature) ToKelvin() float64 {
	return float64(t) + 273.15
}

type IntSlice []int

func (s IntSlice) Sum() int {
	total := 0
	for _, v := range s {
		total += v
	}
	return total
}

func (s IntSlice) Average() float64 {
	if len(s) == 0 {
		return 0
	}
	return float64(s.Sum()) / float64(len(s))
}

func (s IntSlice) Max() int {
	if len(s) == 0 {
		return 0
	}
	max := s[0]
	for _, v := range s {
		if v > max {
			max = v
		}
	}
	return max
}

func (s IntSlice) Min() int {
	if len(s) == 0 {
		return 0
	}
	min := s[0]
	for _, v := range s {
		if v < min {
			min = v
		}
	}
	return min
}

func (s IntSlice) Sort() {
	// 简单冒泡排序
	n := len(s)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if s[j] > s[j+1] {
				s[j], s[j+1] = s[j+1], s[j]
			}
		}
	}
}

type StudentScores map[string]int

func (s StudentScores) Average() float64 {
	if len(s) == 0 {
		return 0
	}
	total := 0
	for _, score := range s {
		total += score
	}
	return float64(total) / float64(len(s))
}

func (s StudentScores) TopStudent() string {
	var topStudent string
	var topScore int

	for student, score := range s {
		if score > topScore {
			topScore = score
			topStudent = student
		}
	}
	return topStudent
}

func (s StudentScores) AddBonus(bonus int) {
	for student := range s {
		s[student] += bonus
	}
}

type DataProcessor func([]int) []int

func (dp DataProcessor) Process(data []int) []int {
	return dp(data)
}

func (dp DataProcessor) ValidateAndProcess(data []int) []int {
	if len(data) == 0 {
		return []int{}
	}
	return dp(data)
}

// 接口定义
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Drawable interface {
	Draw()
}

type Scalable interface {
	Scale(factor float64)
}

// 方法集和接口
func demonstrateMethodSets() {
	fmt.Println("4. 方法集和接口:")

	// Rectangle实现了Shape接口
	rect := Rectangle{Width: 4, Height: 6}
	circle := Circle{Radius: 3}

	shapes := []Shape{rect, circle}

	fmt.Println("图形信息:")
	for i, shape := range shapes {
		fmt.Printf("  图形%d: 面积=%.2f, 周长=%.2f\n",
			i+1, shape.Area(), shape.Perimeter())
	}

	// 方法集差异演示
	demonstrateMethodSetDifferences()

	fmt.Println()
}

func demonstrateMethodSetDifferences() {
	fmt.Println("\n方法集差异:")

	// 值类型的方法集只包含值接收者方法
	rect := Rectangle{Width: 2, Height: 3}
	var shape Shape = rect // OK，Rectangle有值接收者的Area和Perimeter方法
	fmt.Printf("值类型实现接口: %v\n", shape)

	// 指针类型的方法集包含值接收者和指针接收者方法
	pRect := &Rectangle{Width: 4, Height: 5}
	var scalable Scalable = pRect // OK，*Rectangle有指针接收者的Scale方法
	scalable.Scale(1.5)
	fmt.Printf("指针类型实现接口: %v\n", pRect)

	// var scalable2 Scalable = rect // 编译错误！Rectangle没有指针接收者方法
}

// 嵌入类型的方法
func demonstrateEmbeddedMethods() {
	fmt.Println("5. 嵌入类型的方法:")

	// 嵌入提升方法
	box := Box{
		Rectangle: Rectangle{Width: 10, Height: 5},
		Height:    3,
	}

	// 可以直接调用嵌入类型的方法
	fmt.Printf("盒子底面积: %.2f\n", box.Area()) // 来自Rectangle

	// 也可以调用自己的方法
	fmt.Printf("盒子体积: %.2f\n", box.Volume())
	fmt.Printf("盒子表面积: %.2f\n", box.SurfaceArea())

	// 方法覆盖
	coloredRect := ColoredRectangle{
		Rectangle: Rectangle{Width: 6, Height: 4},
		Color:     "红色",
	}

	fmt.Printf("普通矩形描述: %s\n", coloredRect.Rectangle.String())
	fmt.Printf("彩色矩形描述: %s\n", coloredRect.String()) // 覆盖了Rectangle的String方法

	fmt.Println()
}

type Box struct {
	Rectangle         // 嵌入Rectangle
	Height    float64 // 盒子的高度
}

func (b Box) Volume() float64 {
	return b.Area() * b.Height // 使用嵌入的Area方法
}

func (b Box) SurfaceArea() float64 {
	baseArea := b.Area()
	sideArea := b.Perimeter() * b.Height
	return 2*baseArea + sideArea
}

type ColoredRectangle struct {
	Rectangle
	Color string
}

// 覆盖嵌入类型的方法
func (cr ColoredRectangle) String() string {
	return fmt.Sprintf("%s的%s", cr.Color, cr.Rectangle.String())
}

// 方法链式调用
func demonstrateMethodChaining() {
	fmt.Println("6. 方法链式调用:")

	// 构建器模式
	query := NewQueryBuilder().
		Select("name", "age", "email").
		From("users").
		Where("age > %v", 18).
		OrderBy("name").
		Limit(10).
		Build()

	fmt.Printf("SQL查询: %s\n", query)

	// 字符串处理链
	text := "  Hello World  "
	result := NewStringProcessor(text).
		Trim().
		ToLower().
		Replace("world", "go").
		AddPrefix(">>> ").
		AddSuffix(" <<<").
		String()

	fmt.Printf("字符串处理: '%s' -> '%s'\n", text, result)

	// 数学计算链
	calc := NewCalculator(10).
		Add(5).
		Multiply(2).
		Subtract(3).
		Divide(2)

	fmt.Printf("计算结果: %v = %.2f\n", calc.History(), calc.Value())

	fmt.Println()
}

// 查询构建器
type QueryBuilder struct {
	selectFields []string
	fromTable    string
	whereClause  string
	orderBy      string
	limit        int
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.selectFields = fields
	return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.fromTable = table
	return qb
}

func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.whereClause = fmt.Sprintf(condition, args...)
	return qb
}

func (qb *QueryBuilder) OrderBy(field string) *QueryBuilder {
	qb.orderBy = field
	return qb
}

func (qb *QueryBuilder) Limit(n int) *QueryBuilder {
	qb.limit = n
	return qb
}

func (qb *QueryBuilder) Build() string {
	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(qb.selectFields, ", "), qb.fromTable)

	if qb.whereClause != "" {
		query += " WHERE " + qb.whereClause
	}

	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	return query
}

// 字符串处理器
type StringProcessor struct {
	value string
}

func NewStringProcessor(s string) *StringProcessor {
	return &StringProcessor{value: s}
}

func (sp *StringProcessor) Trim() *StringProcessor {
	sp.value = strings.TrimSpace(sp.value)
	return sp
}

func (sp *StringProcessor) ToLower() *StringProcessor {
	sp.value = strings.ToLower(sp.value)
	return sp
}

func (sp *StringProcessor) Replace(old, new string) *StringProcessor {
	sp.value = strings.Replace(sp.value, old, new, -1)
	return sp
}

func (sp *StringProcessor) AddPrefix(prefix string) *StringProcessor {
	sp.value = prefix + sp.value
	return sp
}

func (sp *StringProcessor) AddSuffix(suffix string) *StringProcessor {
	sp.value = sp.value + suffix
	return sp
}

func (sp *StringProcessor) String() string {
	return sp.value
}

// 计算器
type Calculator struct {
	value   float64
	history []string
}

func NewCalculator(initial float64) *Calculator {
	return &Calculator{
		value:   initial,
		history: []string{fmt.Sprintf("%.2f", initial)},
	}
}

func (c *Calculator) Add(n float64) *Calculator {
	c.value += n
	c.history = append(c.history, fmt.Sprintf("+%.2f", n))
	return c
}

func (c *Calculator) Subtract(n float64) *Calculator {
	c.value -= n
	c.history = append(c.history, fmt.Sprintf("-%.2f", n))
	return c
}

func (c *Calculator) Multiply(n float64) *Calculator {
	c.value *= n
	c.history = append(c.history, fmt.Sprintf("×%.2f", n))
	return c
}

func (c *Calculator) Divide(n float64) *Calculator {
	if n != 0 {
		c.value /= n
		c.history = append(c.history, fmt.Sprintf("÷%.2f", n))
	}
	return c
}

func (c *Calculator) Value() float64 {
	return c.value
}

func (c *Calculator) History() string {
	return strings.Join(c.history, " ")
}

// 方法作为值
func demonstrateMethodValues() {
	fmt.Println("7. 方法作为值:")

	// 方法表达式
	rect := Rectangle{Width: 8, Height: 6}

	// 方法值 - 绑定到具体实例
	areaMethod := rect.Area
	perimeterMethod := rect.Perimeter

	fmt.Printf("矩形: %s\n", rect)
	fmt.Printf("通过方法值计算面积: %.2f\n", areaMethod())
	fmt.Printf("通过方法值计算周长: %.2f\n", perimeterMethod())

	// 方法表达式 - 需要传入接收者
	areaExpr := Rectangle.Area
	perimeterExpr := Rectangle.Perimeter

	fmt.Printf("通过方法表达式计算面积: %.2f\n", areaExpr(rect))
	fmt.Printf("通过方法表达式计算周长: %.2f\n", perimeterExpr(rect))

	// 方法集合
	operations := map[string]func() float64{
		"area":      rect.Area,
		"perimeter": rect.Perimeter,
	}

	for name, op := range operations {
		fmt.Printf("%s: %.2f\n", name, op())
	}

	// 指针接收者的方法值
	scaleMethod := rect.Scale
	scaleMethod(1.5)
	fmt.Printf("缩放后: %s\n", rect)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 银行账户系统
	fmt.Println("银行账户系统:")
	account := NewBankAccount("张三", 1000)

	fmt.Printf("初始状态: %s\n", account.String())

	account.Deposit(500)
	fmt.Printf("存款后: %s\n", account.String())

	if account.Withdraw(200) {
		fmt.Printf("取款后: %s\n", account.String())
	}

	if !account.Withdraw(2000) {
		fmt.Printf("余额不足，取款失败\n")
	}

	account.AddInterest(0.05)
	fmt.Printf("加息后: %s\n", account.String())

	// 2. HTTP请求构建器
	fmt.Println("\nHTTP请求构建器:")
	request := NewHTTPRequest("https://api.example.com").
		Method("POST").
		Header("Content-Type", "application/json").
		Header("Authorization", "Bearer token123").
		Body(`{"name":"张三","age":25}`).
		Timeout(30 * time.Second)

	fmt.Printf("HTTP请求: %s\n", request.String())

	// 3. 日志系统
	fmt.Println("\n日志系统:")
	logger := NewLogger("MyApp").
		SetLevel("INFO").
		AddField("version", "1.0").
		AddField("env", "production")

	logger.Info("应用启动")
	logger.Warning("配置文件使用默认值")
	logger.Error("数据库连接失败")

	// 4. 数据验证器
	fmt.Println("\n数据验证器:")
	validator := NewValidator(map[string]interface{}{
		"name":  "张三",
		"age":   25,
		"email": "zhang@example.com",
	})

	isValid := validator.
		Required("name").
		MinLength("name", 2).
		Range("age", 18, 100).
		Email("email").
		IsValid()

	if isValid {
		fmt.Println("数据验证通过")
	} else {
		fmt.Printf("数据验证失败: %v\n", validator.Errors())
	}

	// 5. 缓存系统
	fmt.Println("\n缓存系统:")
	cache := NewCache().
		SetTTL(time.Minute * 5).
		SetMaxSize(100)

	cache.Set("user:123", map[string]string{"name": "张三"})
	cache.Set("config:timeout", 30)

	if value, found := cache.Get("user:123"); found {
		fmt.Printf("缓存命中: %v\n", value)
	}

	cache.SetWithTTL("temp:data", "临时数据", time.Second*10)
	fmt.Printf("缓存状态: %s\n", cache.Stats())

	fmt.Println()
}

// 银行账户
type BankAccount struct {
	owner   string
	balance float64
	history []string
}

func NewBankAccount(owner string, initialBalance float64) *BankAccount {
	account := &BankAccount{
		owner:   owner,
		balance: initialBalance,
		history: []string{},
	}
	account.addHistory(fmt.Sprintf("开户，初始余额: %.2f", initialBalance))
	return account
}

func (ba *BankAccount) Deposit(amount float64) {
	if amount > 0 {
		ba.balance += amount
		ba.addHistory(fmt.Sprintf("存款: %.2f", amount))
	}
}

func (ba *BankAccount) Withdraw(amount float64) bool {
	if amount > 0 && amount <= ba.balance {
		ba.balance -= amount
		ba.addHistory(fmt.Sprintf("取款: %.2f", amount))
		return true
	}
	return false
}

func (ba *BankAccount) AddInterest(rate float64) {
	interest := ba.balance * rate
	ba.balance += interest
	ba.addHistory(fmt.Sprintf("利息: %.2f (利率: %.2f%%)", interest, rate*100))
}

func (ba *BankAccount) Balance() float64 {
	return ba.balance
}

func (ba *BankAccount) addHistory(record string) {
	timestamp := time.Now().Format("15:04:05")
	ba.history = append(ba.history, fmt.Sprintf("[%s] %s", timestamp, record))
}

func (ba *BankAccount) String() string {
	return fmt.Sprintf("%s的账户余额: %.2f", ba.owner, ba.balance)
}

// HTTP请求构建器
type HTTPRequest struct {
	url     string
	method  string
	headers map[string]string
	body    string
	timeout time.Duration
}

func NewHTTPRequest(url string) *HTTPRequest {
	return &HTTPRequest{
		url:     url,
		method:  "GET",
		headers: make(map[string]string),
		timeout: time.Second * 30,
	}
}

func (hr *HTTPRequest) Method(method string) *HTTPRequest {
	hr.method = method
	return hr
}

func (hr *HTTPRequest) Header(key, value string) *HTTPRequest {
	hr.headers[key] = value
	return hr
}

func (hr *HTTPRequest) Body(body string) *HTTPRequest {
	hr.body = body
	return hr
}

func (hr *HTTPRequest) Timeout(timeout time.Duration) *HTTPRequest {
	hr.timeout = timeout
	return hr
}

func (hr *HTTPRequest) String() string {
	return fmt.Sprintf("%s %s (headers: %d, timeout: %v)",
		hr.method, hr.url, len(hr.headers), hr.timeout)
}

// 日志系统
type Logger struct {
	appName string
	level   string
	fields  map[string]interface{}
}

func NewLogger(appName string) *Logger {
	return &Logger{
		appName: appName,
		level:   "INFO",
		fields:  make(map[string]interface{}),
	}
}

func (l *Logger) SetLevel(level string) *Logger {
	l.level = level
	return l
}

func (l *Logger) AddField(key string, value interface{}) *Logger {
	l.fields[key] = value
	return l
}

func (l *Logger) Info(message string) {
	l.log("INFO", message)
}

func (l *Logger) Warning(message string) {
	l.log("WARNING", message)
}

func (l *Logger) Error(message string) {
	l.log("ERROR", message)
}

func (l *Logger) log(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s [%s] %s: %s\n", timestamp, level, l.appName, l.formatFields(), message)
}

func (l *Logger) formatFields() string {
	if len(l.fields) == 0 {
		return ""
	}
	var parts []string
	for k, v := range l.fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return strings.Join(parts, " ")
}

// 数据验证器
type Validator struct {
	data   map[string]interface{}
	errors []string
}

func NewValidator(data map[string]interface{}) *Validator {
	return &Validator{
		data:   data,
		errors: []string{},
	}
}

func (v *Validator) Required(field string) *Validator {
	if value, exists := v.data[field]; !exists || value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s是必需的", field))
	}
	return v
}

func (v *Validator) MinLength(field string, minLen int) *Validator {
	if value, exists := v.data[field]; exists {
		if str, ok := value.(string); ok && len(str) < minLen {
			v.errors = append(v.errors, fmt.Sprintf("%s长度不能少于%d", field, minLen))
		}
	}
	return v
}

func (v *Validator) Range(field string, min, max int) *Validator {
	if value, exists := v.data[field]; exists {
		if num, ok := value.(int); ok && (num < min || num > max) {
			v.errors = append(v.errors, fmt.Sprintf("%s必须在%d-%d之间", field, min, max))
		}
	}
	return v
}

func (v *Validator) Email(field string) *Validator {
	if value, exists := v.data[field]; exists {
		if email, ok := value.(string); ok && !strings.Contains(email, "@") {
			v.errors = append(v.errors, fmt.Sprintf("%s格式无效", field))
		}
	}
	return v
}

func (v *Validator) IsValid() bool {
	return len(v.errors) == 0
}

func (v *Validator) Errors() []string {
	return v.errors
}

// 缓存系统
type Cache struct {
	data    map[string]CacheItem
	ttl     time.Duration
	maxSize int
}

type CacheItem struct {
	value     interface{}
	expiresAt time.Time
}

func NewCache() *Cache {
	return &Cache{
		data:    make(map[string]CacheItem),
		ttl:     time.Minute * 10,
		maxSize: 1000,
	}
}

func (c *Cache) SetTTL(ttl time.Duration) *Cache {
	c.ttl = ttl
	return c
}

func (c *Cache) SetMaxSize(size int) *Cache {
	c.maxSize = size
	return c
}

func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if len(c.data) >= c.maxSize {
		c.evictExpired()
	}

	c.data[key] = CacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	item, exists := c.data[key]
	if !exists || time.Now().After(item.expiresAt) {
		delete(c.data, key)
		return nil, false
	}
	return item.value, true
}

func (c *Cache) evictExpired() {
	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiresAt) {
			delete(c.data, key)
		}
	}
}

func (c *Cache) Stats() string {
	return fmt.Sprintf("缓存项: %d/%d", len(c.data), c.maxSize)
}

/*
=== 练习题 ===

1. 为自定义的Vector类型实现数学运算方法（加法、减法、点积等）

2. 创建一个Stack类型并实现Push、Pop、Peek等方法

3. 设计一个图书管理系统，为Book类型实现各种管理方法

4. 实现一个简单的状态机，使用方法来处理状态转换

5. 创建一个链表数据结构并实现相关方法

6. 设计一个购物车系统，使用方法链进行操作

7. 实现一个配置管理器，支持链式设置和验证

运行命令：
go run main.go

高级练习：
1. 实现一个通用的数据结构库（树、图等）
2. 创建一个事件系统，支持方法作为事件处理器
3. 设计一个ORM框架的基础结构
4. 实现一个函数式编程风格的数据处理库
5. 创建一个插件系统架构

重要概念：
- 方法是带接收者的函数
- 值接收者 vs 指针接收者的选择
- 方法集决定接口实现
- 嵌入类型的方法提升
- 方法可以作为值传递
- 支持链式调用模式
*/
