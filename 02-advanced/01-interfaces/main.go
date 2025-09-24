package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

/*
=== Go语言进阶特性第一课：接口(Interfaces) ===

学习目标：
1. 理解接口的概念和作用
2. 掌握接口的定义和实现
3. 学会接口的组合和嵌入
4. 了解接口的最佳实践
5. 掌握多态编程

Go接口特点：
- 隐式实现，不需要显式声明
- 接口是类型，可以作为参数和返回值
- 空接口可以表示任何类型
- 支持接口组合
- 面向接口编程的核心
*/

func main() {
	fmt.Println("=== Go语言接口学习 ===")

	// 1. 基本接口定义和实现
	demonstrateBasicInterfaces()

	// 2. 接口组合
	demonstrateInterfaceComposition()

	// 3. 接口作为参数和返回值
	demonstrateInterfaceParameters()

	// 4. 类型断言和类型选择
	demonstrateTypeAssertions()

	// 5. 接口的零值和nil接口
	demonstrateNilInterfaces()

	// 6. 接口最佳实践
	demonstrateBestPractices()

	// 7. 标准库接口
	demonstrateStandardInterfaces()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本接口定义
type Shape interface {
	Area() float64
	Perimeter() float64
}

type Drawable interface {
	Draw()
}

type Resizable interface {
	Resize(factor float64)
}

// 实现类型
type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

func (r Rectangle) Draw() {
	fmt.Printf("绘制矩形: %.1f × %.1f\n", r.Width, r.Height)
}

func (r *Rectangle) Resize(factor float64) {
	r.Width *= factor
	r.Height *= factor
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) Draw() {
	fmt.Printf("绘制圆形: 半径 %.1f\n", c.Radius)
}

func (c *Circle) Resize(factor float64) {
	c.Radius *= factor
}

// 基本接口定义和实现
func demonstrateBasicInterfaces() {
	fmt.Println("1. 基本接口定义和实现:")

	// 创建实现了Shape接口的类型
	var s1 Shape = Rectangle{Width: 10, Height: 5}
	var s2 Shape = Circle{Radius: 3}

	shapes := []Shape{s1, s2}

	fmt.Println("图形信息:")
	for i, shape := range shapes {
		fmt.Printf("  图形%d: 面积=%.2f, 周长=%.2f\n",
			i+1, shape.Area(), shape.Perimeter())
	}

	// 接口的动态类型
	fmt.Println("\n接口的动态类型:")
	var shape Shape

	shape = Rectangle{Width: 4, Height: 6}
	fmt.Printf("当前类型: %T, 面积: %.2f\n", shape, shape.Area())

	shape = Circle{Radius: 2}
	fmt.Printf("当前类型: %T, 面积: %.2f\n", shape, shape.Area())

	// 接口值的比较
	fmt.Println("\n接口值的比较:")
	var s3, s4 Shape
	s3 = Rectangle{Width: 2, Height: 3}
	s4 = Rectangle{Width: 2, Height: 3}

	fmt.Printf("相同值的接口比较: %t\n", s3 == s4)

	// 空接口
	var empty interface{}
	empty = 42
	fmt.Printf("空接口存储整数: %v\n", empty)
	empty = "hello"
	fmt.Printf("空接口存储字符串: %v\n", empty)

	fmt.Println()
}

// 接口组合
type DrawableShape interface {
	Shape    // 嵌入Shape接口
	Drawable // 嵌入Drawable接口
}

type ResizableDrawableShape interface {
	DrawableShape // 嵌入组合接口
	Resizable     // 嵌入Resizable接口
}

func demonstrateInterfaceComposition() {
	fmt.Println("2. 接口组合:")

	// Rectangle实现了所有需要的方法
	rect := &Rectangle{Width: 8, Height: 4}

	// 作为组合接口使用
	var drawableShape DrawableShape = rect
	fmt.Printf("组合接口 - 面积: %.2f\n", drawableShape.Area())
	drawableShape.Draw()

	// 作为更复杂的组合接口使用
	var resizableShape ResizableDrawableShape = rect
	fmt.Printf("调整前尺寸: %.1f × %.1f\n", rect.Width, rect.Height)
	resizableShape.Resize(1.5)
	fmt.Printf("调整后尺寸: %.1f × %.1f\n", rect.Width, rect.Height)
	resizableShape.Draw()

	// 接口的选择性实现
	fmt.Println("\n接口的选择性实现:")
	triangle := Triangle{Base: 6, Height: 4}

	// Triangle只实现了Shape接口
	var shape Shape = triangle
	fmt.Printf("三角形面积: %.2f\n", shape.Area())

	// 但不能作为DrawableShape使用
	// var ds DrawableShape = triangle // 编译错误

	fmt.Println()
}

type Triangle struct {
	Base, Height float64
}

func (t Triangle) Area() float64 {
	return 0.5 * t.Base * t.Height
}

func (t Triangle) Perimeter() float64 {
	// 简化计算，假设是等腰三角形
	side := math.Sqrt((t.Base/2)*(t.Base/2) + t.Height*t.Height)
	return t.Base + 2*side
}

// 接口作为参数和返回值
func demonstrateInterfaceParameters() {
	fmt.Println("3. 接口作为参数和返回值:")

	shapes := []Shape{
		Rectangle{Width: 5, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 4, Height: 6},
	}

	// 接口作为参数
	totalArea := calculateTotalArea(shapes)
	fmt.Printf("总面积: %.2f\n", totalArea)

	largest := findLargestShape(shapes)
	fmt.Printf("最大图形面积: %.2f\n", largest.Area())

	// 接口作为返回值
	factory := getShapeFactory("rectangle")
	shape1 := factory()
	fmt.Printf("工厂创建的图形: %T, 面积: %.2f\n", shape1, shape1.Area())

	factory = getShapeFactory("circle")
	shape2 := factory()
	fmt.Printf("工厂创建的图形: %T, 面积: %.2f\n", shape2, shape2.Area())

	// 函数式接口
	processor := getShapeProcessor()
	processedShapes := processor(shapes, func(s Shape) bool {
		return s.Area() > 10
	})

	fmt.Printf("过滤后的图形数量: %d\n", len(processedShapes))

	fmt.Println()
}

func calculateTotalArea(shapes []Shape) float64 {
	total := 0.0
	for _, shape := range shapes {
		total += shape.Area()
	}
	return total
}

func findLargestShape(shapes []Shape) Shape {
	if len(shapes) == 0 {
		return nil
	}

	largest := shapes[0]
	for _, shape := range shapes[1:] {
		if shape.Area() > largest.Area() {
			largest = shape
		}
	}
	return largest
}

func getShapeFactory(shapeType string) func() Shape {
	switch shapeType {
	case "rectangle":
		return func() Shape {
			return &Rectangle{Width: 4, Height: 3}
		}
	case "circle":
		return func() Shape {
			return &Circle{Radius: 2}
		}
	default:
		return func() Shape {
			return &Rectangle{Width: 1, Height: 1}
		}
	}
}

func getShapeProcessor() func([]Shape, func(Shape) bool) []Shape {
	return func(shapes []Shape, filter func(Shape) bool) []Shape {
		var result []Shape
		for _, shape := range shapes {
			if filter(shape) {
				result = append(result, shape)
			}
		}
		return result
	}
}

// 类型断言和类型选择
func demonstrateTypeAssertions() {
	fmt.Println("4. 类型断言和类型选择:")

	shapes := []Shape{
		Rectangle{Width: 4, Height: 3},
		Circle{Radius: 2},
		Triangle{Base: 6, Height: 4},
	}

	fmt.Println("类型断言:")
	for i, shape := range shapes {
		// 安全的类型断言
		if rect, ok := shape.(Rectangle); ok {
			fmt.Printf("  图形%d是矩形: %.1f × %.1f\n", i+1, rect.Width, rect.Height)
		} else if circle, ok := shape.(Circle); ok {
			fmt.Printf("  图形%d是圆形: 半径 %.1f\n", i+1, circle.Radius)
		} else {
			fmt.Printf("  图形%d是其他类型: %T\n", i+1, shape)
		}
	}

	fmt.Println("\n类型选择(Type Switch):")
	for i, shape := range shapes {
		switch s := shape.(type) {
		case Rectangle:
			fmt.Printf("  图形%d: 矩形 %.1f×%.1f, 面积=%.2f\n",
				i+1, s.Width, s.Height, s.Area())
		case Circle:
			fmt.Printf("  图形%d: 圆形 r=%.1f, 面积=%.2f\n",
				i+1, s.Radius, s.Area())
		case Triangle:
			fmt.Printf("  图形%d: 三角形 底=%.1f 高=%.1f, 面积=%.2f\n",
				i+1, s.Base, s.Height, s.Area())
		default:
			fmt.Printf("  图形%d: 未知类型 %T\n", i+1, s)
		}
	}

	// 类型断言的panic风险
	fmt.Println("\n安全的类型断言:")
	var shape Shape = Circle{Radius: 3}

	// 安全方式
	if circle, ok := shape.(Circle); ok {
		fmt.Printf("安全转换成功: 半径 %.1f\n", circle.Radius)
	}

	// 危险方式（可能panic）
	// rect := shape.(Rectangle) // 如果shape不是Rectangle会panic

	fmt.Println()
}

// 接口的零值和nil接口
func demonstrateNilInterfaces() {
	fmt.Println("5. 接口的零值和nil接口:")

	// nil接口
	var shape Shape
	fmt.Printf("零值接口: %v, 是否为nil: %t\n", shape, shape == nil)

	// nil接口调用方法会panic
	// shape.Area() // panic: runtime error: invalid memory address

	// 接口存储nil指针
	var rectPtr *Rectangle
	shape = rectPtr
	fmt.Printf("接口存储nil指针: %v, 是否为nil: %t\n", shape, shape == nil)

	// 此时接口不为nil，但动态值为nil
	if shape != nil {
		fmt.Println("接口不为nil，但动态值为nil")
		// shape.Area() // panic: runtime error: invalid memory address
	}

	// 正确的nil检查
	fmt.Println("\n正确的nil检查:")
	checkAndCallShape(nil)
	checkAndCallShape(Rectangle{Width: 2, Height: 3})

	var nilPtr *Rectangle
	checkAndCallShape(nilPtr)

	fmt.Println()
}

func checkAndCallShape(s Shape) {
	if s == nil {
		fmt.Println("  形状为nil，跳过处理")
		return
	}

	// 检查动态值是否为nil
	switch shape := s.(type) {
	case *Rectangle:
		if shape == nil {
			fmt.Println("  矩形指针为nil，跳过处理")
			return
		}
	case *Circle:
		if shape == nil {
			fmt.Println("  圆形指针为nil，跳过处理")
			return
		}
	}

	fmt.Printf("  处理形状: %T, 面积: %.2f\n", s, s.Area())
}

// 接口最佳实践
func demonstrateBestPractices() {
	fmt.Println("6. 接口最佳实践:")

	// 1. 接口隔离原则 - 小而专一的接口
	fmt.Println("小而专一的接口:")
	writer := &FileWriter{filename: "test.txt"}
	closer := &FileCloser{filename: "test.txt"}

	writeData(writer, "Hello World")
	closeResource(closer)

	// 2. 接受接口，返回具体类型
	fmt.Println("\n接受接口，返回具体类型:")
	buffer := NewBuffer()
	copyData(buffer, "测试数据")
	result := buffer.String()
	fmt.Printf("缓冲区内容: %s\n", result)

	// 3. 组合小接口形成大接口
	fmt.Println("\n接口组合:")
	readWriter := &MemoryReadWriter{data: "初始数据"}
	processReadWriter(readWriter)

	// 4. 使用标准库接口
	fmt.Println("\n使用标准库接口:")
	data := []string{"banana", "apple", "cherry"}
	fmt.Printf("排序前: %v\n", data)
	sort.Strings(data)
	fmt.Printf("排序后: %v\n", data)

	// 自定义排序
	numbers := []int{3, 1, 4, 1, 5, 9, 2, 6}
	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] < numbers[j]
	})
	fmt.Printf("自定义排序: %v\n", numbers)

	fmt.Println()
}

// 小接口示例
type Writer interface {
	Write(data string) error
}

type Closer interface {
	Close() error
}

type ReadWriter interface {
	Reader
	Writer
}

type Reader interface {
	Read() (string, error)
}

type FileWriter struct {
	filename string
}

func (fw *FileWriter) Write(data string) error {
	fmt.Printf("写入文件 %s: %s\n", fw.filename, data)
	return nil
}

type FileCloser struct {
	filename string
}

func (fc *FileCloser) Close() error {
	fmt.Printf("关闭文件 %s\n", fc.filename)
	return nil
}

func writeData(w Writer, data string) {
	w.Write(data)
}

func closeResource(c Closer) {
	c.Close()
}

// 缓冲区实现
type Buffer struct {
	data strings.Builder
}

func NewBuffer() *Buffer {
	return &Buffer{}
}

func (b *Buffer) Write(data string) error {
	b.data.WriteString(data)
	return nil
}

func (b *Buffer) String() string {
	return b.data.String()
}

func copyData(w Writer, data string) {
	w.Write(data)
}

// 内存读写器
type MemoryReadWriter struct {
	data string
}

func (mrw *MemoryReadWriter) Read() (string, error) {
	return mrw.data, nil
}

func (mrw *MemoryReadWriter) Write(data string) error {
	mrw.data = data
	return nil
}

func processReadWriter(rw ReadWriter) {
	// 读取
	data, _ := rw.Read()
	fmt.Printf("读取数据: %s\n", data)

	// 写入
	rw.Write("新数据")

	// 再次读取
	newData, _ := rw.Read()
	fmt.Printf("更新后数据: %s\n", newData)
}

// 标准库接口
func demonstrateStandardInterfaces() {
	fmt.Println("7. 标准库接口:")

	// fmt.Stringer接口
	fmt.Println("fmt.Stringer接口:")
	person := Person{Name: "张三", Age: 25}
	fmt.Printf("个人信息: %s\n", person) // 自动调用String()方法

	// error接口
	fmt.Println("\nerror接口:")
	if _, err := divide(10, 0); err != nil {
		fmt.Printf("错误: %s\n", err)
	}

	if result, err := divide(10, 2); err == nil {
		fmt.Printf("结果: %.2f\n", result)
	}

	// sort.Interface接口
	fmt.Println("\nsort.Interface接口:")
	people := People{
		{Name: "张三", Age: 25},
		{Name: "李四", Age: 20},
		{Name: "王五", Age: 30},
	}

	fmt.Printf("排序前: %v\n", people)
	sort.Sort(people)
	fmt.Printf("排序后: %v\n", people)

	fmt.Println()
}

type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("%s (%d岁)", p.Name, p.Age)
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零")
	}
	return a / b, nil
}

// 实现sort.Interface的People类型
type People []Person

func (p People) Len() int           { return len(p) }
func (p People) Less(i, j int) bool { return p[i].Age < p[j].Age }
func (p People) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 插件系统
	fmt.Println("插件系统:")
	pluginManager := &PluginManager{}

	pluginManager.Register(&LogPlugin{})
	pluginManager.Register(&EmailPlugin{})
	pluginManager.Register(&SMSPlugin{})

	pluginManager.ExecuteAll("系统启动")

	// 2. 数据存储抽象
	fmt.Println("\n数据存储抽象:")

	// 使用内存存储
	memStore := &MemoryStore{data: make(map[string]string)}
	testStorage(memStore, "内存存储")

	// 使用文件存储
	fileStore := &FileStore{filename: "data.txt"}
	testStorage(fileStore, "文件存储")

	// 3. 通知系统
	fmt.Println("\n通知系统:")
	notifier := &NotificationManager{}

	notifier.AddNotifier(&EmailNotifier{})
	notifier.AddNotifier(&SMSNotifier{})
	notifier.AddNotifier(&PushNotifier{})

	notifier.NotifyAll("系统维护通知", "系统将在今晚进行维护")

	// 4. 策略模式
	fmt.Println("\n策略模式:")
	calculator := &PriceCalculator{}

	// 普通用户价格策略
	calculator.SetStrategy(&RegularPriceStrategy{})
	fmt.Printf("普通用户价格: %.2f\n", calculator.Calculate(100))

	// VIP用户价格策略
	calculator.SetStrategy(&VIPPriceStrategy{})
	fmt.Printf("VIP用户价格: %.2f\n", calculator.Calculate(100))

	// 批发价格策略
	calculator.SetStrategy(&WholesalePriceStrategy{})
	fmt.Printf("批发价格: %.2f\n", calculator.Calculate(100))

	// 5. 中间件模式
	fmt.Println("\n中间件模式:")
	handler := &HTTPHandler{}

	// 添加中间件
	finalHandler := ChainMiddleware(handler,
		&LoggingMiddleware{},
		&AuthMiddleware{},
		&CORSMiddleware{},
	)

	finalHandler.Handle("GET /api/users")

	fmt.Println()
}

// 插件系统
type Plugin interface {
	Execute(data string)
	Name() string
}

type PluginManager struct {
	plugins []Plugin
}

func (pm *PluginManager) Register(plugin Plugin) {
	pm.plugins = append(pm.plugins, plugin)
}

func (pm *PluginManager) ExecuteAll(data string) {
	for _, plugin := range pm.plugins {
		fmt.Printf("  执行插件 %s: %s\n", plugin.Name(), data)
		plugin.Execute(data)
	}
}

type LogPlugin struct{}

func (lp *LogPlugin) Execute(data string) {
	fmt.Printf("    [LOG] %s\n", data)
}

func (lp *LogPlugin) Name() string {
	return "日志插件"
}

type EmailPlugin struct{}

func (ep *EmailPlugin) Execute(data string) {
	fmt.Printf("    [EMAIL] 发送邮件: %s\n", data)
}

func (ep *EmailPlugin) Name() string {
	return "邮件插件"
}

type SMSPlugin struct{}

func (sp *SMSPlugin) Execute(data string) {
	fmt.Printf("    [SMS] 发送短信: %s\n", data)
}

func (sp *SMSPlugin) Name() string {
	return "短信插件"
}

// 数据存储抽象
type Storage interface {
	Save(key, value string) error
	Load(key string) (string, error)
	Delete(key string) error
}

type MemoryStore struct {
	data map[string]string
}

func (ms *MemoryStore) Save(key, value string) error {
	ms.data[key] = value
	return nil
}

func (ms *MemoryStore) Load(key string) (string, error) {
	if value, exists := ms.data[key]; exists {
		return value, nil
	}
	return "", fmt.Errorf("键 %s 不存在", key)
}

func (ms *MemoryStore) Delete(key string) error {
	delete(ms.data, key)
	return nil
}

type FileStore struct {
	filename string
}

func (fs *FileStore) Save(key, value string) error {
	fmt.Printf("    保存到文件 %s: %s=%s\n", fs.filename, key, value)
	return nil
}

func (fs *FileStore) Load(key string) (string, error) {
	fmt.Printf("    从文件 %s 加载: %s\n", fs.filename, key)
	return "模拟数据", nil
}

func (fs *FileStore) Delete(key string) error {
	fmt.Printf("    从文件 %s 删除: %s\n", fs.filename, key)
	return nil
}

func testStorage(store Storage, name string) {
	fmt.Printf("  测试 %s:\n", name)
	store.Save("user1", "张三")
	value, _ := store.Load("user1")
	fmt.Printf("    加载结果: %s\n", value)
	store.Delete("user1")
}

// 通知系统
type Notifier interface {
	Notify(title, message string) error
}

type NotificationManager struct {
	notifiers []Notifier
}

func (nm *NotificationManager) AddNotifier(notifier Notifier) {
	nm.notifiers = append(nm.notifiers, notifier)
}

func (nm *NotificationManager) NotifyAll(title, message string) {
	for _, notifier := range nm.notifiers {
		notifier.Notify(title, message)
	}
}

type EmailNotifier struct{}

func (en *EmailNotifier) Notify(title, message string) error {
	fmt.Printf("  [邮件] %s: %s\n", title, message)
	return nil
}

type SMSNotifier struct{}

func (sn *SMSNotifier) Notify(title, message string) error {
	fmt.Printf("  [短信] %s: %s\n", title, message)
	return nil
}

type PushNotifier struct{}

func (pn *PushNotifier) Notify(title, message string) error {
	fmt.Printf("  [推送] %s: %s\n", title, message)
	return nil
}

// 策略模式
type PriceStrategy interface {
	Calculate(basePrice float64) float64
}

type PriceCalculator struct {
	strategy PriceStrategy
}

func (pc *PriceCalculator) SetStrategy(strategy PriceStrategy) {
	pc.strategy = strategy
}

func (pc *PriceCalculator) Calculate(basePrice float64) float64 {
	if pc.strategy == nil {
		return basePrice
	}
	return pc.strategy.Calculate(basePrice)
}

type RegularPriceStrategy struct{}

func (rps *RegularPriceStrategy) Calculate(basePrice float64) float64 {
	return basePrice
}

type VIPPriceStrategy struct{}

func (vps *VIPPriceStrategy) Calculate(basePrice float64) float64 {
	return basePrice * 0.8 // 8折
}

type WholesalePriceStrategy struct{}

func (wps *WholesalePriceStrategy) Calculate(basePrice float64) float64 {
	return basePrice * 0.6 // 6折
}

// 中间件模式
type Handler interface {
	Handle(request string)
}

type Middleware interface {
	Process(request string, next Handler)
}

type HTTPHandler struct{}

func (h *HTTPHandler) Handle(request string) {
	fmt.Printf("    处理请求: %s\n", request)
}

type LoggingMiddleware struct{}

func (lm *LoggingMiddleware) Process(request string, next Handler) {
	fmt.Printf("    [中间件] 记录请求: %s\n", request)
	next.Handle(request)
	fmt.Printf("    [中间件] 记录响应\n")
}

type AuthMiddleware struct{}

func (am *AuthMiddleware) Process(request string, next Handler) {
	fmt.Printf("    [中间件] 身份验证\n")
	next.Handle(request)
}

type CORSMiddleware struct{}

func (cm *CORSMiddleware) Process(request string, next Handler) {
	fmt.Printf("    [中间件] 添加CORS头\n")
	next.Handle(request)
}

// 中间件链
type MiddlewareHandler struct {
	middleware Middleware
	next       Handler
}

func (mh *MiddlewareHandler) Handle(request string) {
	mh.middleware.Process(request, mh.next)
}

func ChainMiddleware(handler Handler, middlewares ...Middleware) Handler {
	result := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		result = &MiddlewareHandler{
			middleware: middlewares[i],
			next:       result,
		}
	}
	return result
}

/*
=== 练习题 ===

1. 设计一个日志系统，支持不同的日志输出方式（文件、控制台、网络）

2. 实现一个缓存系统，支持不同的缓存策略（LRU、LFU、TTL）

3. 创建一个数据库抽象层，支持不同的数据库类型

4. 设计一个消息队列系统，支持不同的消息传递机制

5. 实现一个HTTP客户端库，支持不同的传输协议

6. 创建一个配置管理系统，支持多种配置源

7. 设计一个任务调度器，支持不同的调度策略

运行命令：
go run main.go

高级练习：
1. 实现一个服务发现系统
2. 创建一个分布式锁接口
3. 设计一个流处理框架
4. 实现一个插件化的Web框架
5. 创建一个通用的序列化接口

重要概念：
- 接口是Go多态的基础
- 隐式实现，无需显式声明
- 接口组合实现复杂功能
- 空接口可以表示任何类型
- 类型断言用于类型转换
- 面向接口编程的设计原则
*/
