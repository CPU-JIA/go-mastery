package main

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

/*
=== Go语言进阶特性第二课：空接口(Empty Interface) ===

学习目标：
1. 理解空接口的概念和用途
2. 掌握空接口的使用场景
3. 学会类型断言和类型选择
4. 了解反射的基础应用
5. 掌握通用编程技巧

Go空接口特点：
- interface{}可以表示任何类型
- 是所有类型的超集
- 常用于通用函数和数据容器
- 需要通过类型断言获取具体类型
- 是反射的基础
*/

func main() {
	fmt.Println("=== Go语言空接口学习 ===")

	// 1. 空接口基础
	demonstrateEmptyInterfaceBasics()

	// 2. 类型断言和类型选择
	demonstrateTypeAssertions()

	// 3. 空接口在数据容器中的应用
	demonstrateDataContainers()

	// 4. 通用函数的实现
	demonstrateGenericFunctions()

	// 5. JSON和序列化
	demonstrateJSONHandling()

	// 6. 反射基础
	demonstrateReflectionBasics()

	// 7. 性能考虑
	demonstratePerformanceConsiderations()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 空接口基础
func demonstrateEmptyInterfaceBasics() {
	fmt.Println("1. 空接口基础:")

	// 空接口可以存储任何类型
	var empty interface{}

	empty = 42
	fmt.Printf("存储整数: %v (类型: %T)\n", empty, empty)

	empty = "Hello World"
	fmt.Printf("存储字符串: %v (类型: %T)\n", empty, empty)

	empty = []int{1, 2, 3}
	fmt.Printf("存储切片: %v (类型: %T)\n", empty, empty)

	empty = map[string]int{"apple": 5}
	fmt.Printf("存储映射: %v (类型: %T)\n", empty, empty)

	empty = time.Now()
	fmt.Printf("存储时间: %v (类型: %T)\n", empty, empty)

	// 空接口切片
	values := []interface{}{
		42,
		"hello",
		3.14,
		true,
		[]string{"a", "b"},
		map[string]int{"x": 1},
	}

	fmt.Println("\n空接口切片:")
	for i, v := range values {
		fmt.Printf("  [%d]: %v (类型: %T)\n", i, v, v)
	}

	// 空接口映射
	data := map[string]interface{}{
		"name":    "张三",
		"age":     25,
		"active":  true,
		"score":   95.5,
		"hobbies": []string{"读书", "游泳"},
	}

	fmt.Println("\n空接口映射:")
	for key, value := range data {
		fmt.Printf("  %s: %v (类型: %T)\n", key, value, value)
	}

	fmt.Println()
}

// 类型断言和类型选择
func demonstrateTypeAssertions() {
	fmt.Println("2. 类型断言和类型选择:")

	values := []interface{}{
		42,
		"hello world",
		3.14159,
		true,
		[]int{1, 2, 3},
		map[string]string{"key": "value"},
		Person{Name: "张三", Age: 25},
		&Person{Name: "李四", Age: 30},
		nil,
	}

	fmt.Println("安全的类型断言:")
	for i, value := range values {
		fmt.Printf("[%d] %v: ", i, value)

		// 使用逗号ok模式进行安全断言
		if str, ok := value.(string); ok {
			fmt.Printf("字符串，长度: %d\n", len(str))
		} else if num, ok := value.(int); ok {
			fmt.Printf("整数，平方: %d\n", num*num)
		} else if f, ok := value.(float64); ok {
			fmt.Printf("浮点数，四舍五入: %.0f\n", f)
		} else if b, ok := value.(bool); ok {
			fmt.Printf("布尔值，取反: %t\n", !b)
		} else if value == nil {
			fmt.Println("nil值")
		} else {
			fmt.Printf("其他类型: %T\n", value)
		}
	}

	fmt.Println("\n类型选择(Type Switch):")
	for i, value := range values {
		fmt.Printf("[%d] ", i)
		processValue(value)
	}

	// 类型断言的panic风险
	fmt.Println("\n类型断言的风险:")
	var empty interface{} = "hello"

	// 安全方式
	if str, ok := empty.(string); ok {
		fmt.Printf("安全转换: %s\n", str)
	}

	// 危险方式演示（已注释避免panic）
	// num := empty.(int) // 这会引发panic
	fmt.Println("危险转换已跳过（避免panic）")

	fmt.Println()
}

type Person struct {
	Name string
	Age  int
}

func (p Person) String() string {
	return fmt.Sprintf("Person{Name: %s, Age: %d}", p.Name, p.Age)
}

func processValue(value interface{}) {
	switch v := value.(type) {
	case nil:
		fmt.Println("处理nil值")
	case bool:
		fmt.Printf("处理布尔值: %t\n", v)
	case int:
		fmt.Printf("处理整数: %d (是否为偶数: %t)\n", v, v%2 == 0)
	case float64:
		fmt.Printf("处理浮点数: %.2f\n", v)
	case string:
		fmt.Printf("处理字符串: '%s' (长度: %d)\n", v, len(v))
	case []int:
		sum := 0
		for _, n := range v {
			sum += n
		}
		fmt.Printf("处理整数切片: %v (和: %d)\n", v, sum)
	case map[string]string:
		fmt.Printf("处理字符串映射: %v (键数: %d)\n", v, len(v))
	case Person:
		fmt.Printf("处理Person值: %s\n", v)
	case *Person:
		if v != nil {
			fmt.Printf("处理Person指针: %s\n", *v)
		} else {
			fmt.Println("处理nil Person指针")
		}
	default:
		fmt.Printf("处理未知类型: %T (%v)\n", v, v)
	}
}

// 空接口在数据容器中的应用
func demonstrateDataContainers() {
	fmt.Println("3. 空接口在数据容器中的应用:")

	// 通用栈
	fmt.Println("通用栈:")
	stack := NewStack()

	stack.Push(1)
	stack.Push("hello")
	stack.Push(3.14)
	stack.Push([]int{1, 2, 3})

	fmt.Printf("栈大小: %d\n", stack.Size())

	for !stack.IsEmpty() {
		value := stack.Pop()
		fmt.Printf("  弹出: %v (%T)\n", value, value)
	}

	// 通用队列
	fmt.Println("\n通用队列:")
	queue := NewQueue()

	queue.Enqueue("first")
	queue.Enqueue(42)
	queue.Enqueue(true)

	for !queue.IsEmpty() {
		value := queue.Dequeue()
		fmt.Printf("  出队: %v (%T)\n", value, value)
	}

	// 通用集合
	fmt.Println("\n通用集合:")
	set := NewSet()

	set.Add(1)
	set.Add("hello")
	set.Add(1) // 重复元素
	set.Add(3.14)
	set.Add("hello") // 重复元素

	fmt.Printf("集合大小: %d\n", set.Size())
	fmt.Printf("包含1: %t\n", set.Contains(1))
	fmt.Printf("包含'world': %t\n", set.Contains("world"))

	values := set.ToSlice()
	fmt.Printf("集合元素: %v\n", values)

	// 通用缓存
	fmt.Println("\n通用缓存:")
	cache := NewCache()

	cache.Set("user:123", map[string]interface{}{
		"name": "张三",
		"age":  25,
	})
	cache.Set("config:timeout", 30)
	cache.Set("flag:enabled", true)

	if value, found := cache.Get("user:123"); found {
		fmt.Printf("缓存命中: %v\n", value)
	}

	if value, found := cache.Get("nonexistent"); !found {
		fmt.Printf("缓存未命中: %v\n", value)
	}

	fmt.Println()
}

// 通用栈实现
type Stack struct {
	items []interface{}
}

func NewStack() *Stack {
	return &Stack{items: make([]interface{}, 0)}
}

func (s *Stack) Push(item interface{}) {
	s.items = append(s.items, item)
}

func (s *Stack) Pop() interface{} {
	if len(s.items) == 0 {
		return nil
	}
	index := len(s.items) - 1
	item := s.items[index]
	s.items = s.items[:index]
	return item
}

func (s *Stack) Size() int {
	return len(s.items)
}

func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

// 通用队列实现
type Queue struct {
	items []interface{}
}

func NewQueue() *Queue {
	return &Queue{items: make([]interface{}, 0)}
}

func (q *Queue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

func (q *Queue) Dequeue() interface{} {
	if len(q.items) == 0 {
		return nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

// 通用集合实现
type Set struct {
	items map[interface{}]bool
}

func NewSet() *Set {
	return &Set{items: make(map[interface{}]bool)}
}

func (s *Set) Add(item interface{}) {
	s.items[item] = true
}

func (s *Set) Contains(item interface{}) bool {
	return s.items[item]
}

func (s *Set) Size() int {
	return len(s.items)
}

func (s *Set) ToSlice() []interface{} {
	result := make([]interface{}, 0, len(s.items))
	for item := range s.items {
		result = append(result, item)
	}
	return result
}

// 通用缓存实现
type Cache struct {
	data map[string]interface{}
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]interface{})}
}

func (c *Cache) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *Cache) Get(key string) (interface{}, bool) {
	value, exists := c.data[key]
	return value, exists
}

// 通用函数的实现
func demonstrateGenericFunctions() {
	fmt.Println("4. 通用函数的实现:")

	// 通用打印函数
	fmt.Println("通用打印函数:")
	printValue(42)
	printValue("hello")
	printValue([]int{1, 2, 3})
	printValue(map[string]int{"a": 1})

	// 通用比较函数
	fmt.Println("\n通用比较函数:")
	fmt.Printf("42 == 42: %t\n", isEqual(42, 42))
	fmt.Printf("'hello' == 'world': %t\n", isEqual("hello", "world"))
	fmt.Printf("3.14 == 3.14: %t\n", isEqual(3.14, 3.14))

	// 通用转换函数
	fmt.Println("\n通用转换函数:")
	fmt.Printf("转换42: '%s'\n", toString(42))
	fmt.Printf("转换3.14: '%s'\n", toString(3.14))
	fmt.Printf("转换true: '%s'\n", toString(true))
	fmt.Printf("转换'hello': '%s'\n", toString("hello"))

	// 通用查找函数
	fmt.Println("\n通用查找函数:")
	slice1 := []interface{}{1, 2, 3, "hello", 4.5}
	fmt.Printf("查找3: 索引=%d\n", findIndex(slice1, 3))
	fmt.Printf("查找'hello': 索引=%d\n", findIndex(slice1, "hello"))
	fmt.Printf("查找'world': 索引=%d\n", findIndex(slice1, "world"))

	// 通用过滤函数
	fmt.Println("\n通用过滤函数:")
	numbers := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evens := filter(numbers, isEven)
	fmt.Printf("偶数: %v\n", evens)

	strings := []interface{}{"hello", "world", "go", "programming"}
	longStrings := filter(strings, isLongString)
	fmt.Printf("长字符串: %v\n", longStrings)

	// 通用映射函数
	fmt.Println("\n通用映射函数:")
	squares := mapFunc([]interface{}{1, 2, 3, 4, 5}, square)
	fmt.Printf("平方: %v\n", squares)

	upperCases := mapFunc([]interface{}{"hello", "world"}, toUpper)
	fmt.Printf("大写: %v\n", upperCases)

	fmt.Println()
}

func printValue(value interface{}) {
	switch v := value.(type) {
	case nil:
		fmt.Println("  nil")
	case bool:
		fmt.Printf("  布尔值: %t\n", v)
	case int, int8, int16, int32, int64:
		fmt.Printf("  整数: %v\n", v)
	case uint, uint8, uint16, uint32, uint64:
		fmt.Printf("  无符号整数: %v\n", v)
	case float32, float64:
		fmt.Printf("  浮点数: %v\n", v)
	case string:
		fmt.Printf("  字符串: '%s'\n", v)
	case []int:
		fmt.Printf("  整数切片: %v\n", v)
	case map[string]int:
		fmt.Printf("  字符串到整数的映射: %v\n", v)
	default:
		fmt.Printf("  其他类型(%T): %v\n", v, v)
	}
}

func isEqual(a, b interface{}) bool {
	return a == b
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func findIndex(slice []interface{}, target interface{}) int {
	for i, value := range slice {
		if value == target {
			return i
		}
	}
	return -1
}

func filter(slice []interface{}, predicate func(interface{}) bool) []interface{} {
	var result []interface{}
	for _, value := range slice {
		if predicate(value) {
			result = append(result, value)
		}
	}
	return result
}

func isEven(value interface{}) bool {
	if num, ok := value.(int); ok {
		return num%2 == 0
	}
	return false
}

func isLongString(value interface{}) bool {
	if str, ok := value.(string); ok {
		return len(str) > 4
	}
	return false
}

func mapFunc(slice []interface{}, transform func(interface{}) interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, value := range slice {
		result[i] = transform(value)
	}
	return result
}

func square(value interface{}) interface{} {
	if num, ok := value.(int); ok {
		return num * num
	}
	return value
}

func toUpper(value interface{}) interface{} {
	if str, ok := value.(string); ok {
		return fmt.Sprintf(">>> %s <<<", str)
	}
	return value
}

// JSON和序列化处理
func demonstrateJSONHandling() {
	fmt.Println("5. JSON和序列化处理:")

	// 模拟JSON解析结果
	jsonData := map[string]interface{}{
		"name":    "张三",
		"age":     25,
		"active":  true,
		"score":   95.5,
		"hobbies": []interface{}{"读书", "游泳", "编程"},
		"address": map[string]interface{}{
			"city":    "北京",
			"zipcode": "100000",
		},
		"metadata": nil,
	}

	fmt.Println("解析JSON数据:")
	parseJSONData(jsonData)

	// 类型安全的访问
	fmt.Println("\n类型安全的JSON访问:")
	name := getStringField(jsonData, "name", "未知")
	age := getIntField(jsonData, "age", 0)
	active := getBoolField(jsonData, "active", false)
	score := getFloatField(jsonData, "score", 0.0)

	fmt.Printf("姓名: %s\n", name)
	fmt.Printf("年龄: %d\n", age)
	fmt.Printf("激活: %t\n", active)
	fmt.Printf("分数: %.2f\n", score)

	// 嵌套对象访问
	if address, ok := jsonData["address"].(map[string]interface{}); ok {
		city := getStringField(address, "city", "未知城市")
		zipcode := getStringField(address, "zipcode", "000000")
		fmt.Printf("地址: %s %s\n", city, zipcode)
	}

	// 数组访问
	if hobbies, ok := jsonData["hobbies"].([]interface{}); ok {
		fmt.Print("爱好: ")
		for i, hobby := range hobbies {
			if str, ok := hobby.(string); ok {
				if i > 0 {
					fmt.Print(", ")
				}
				fmt.Print(str)
			}
		}
		fmt.Println()
	}

	fmt.Println()
}

func parseJSONData(data map[string]interface{}) {
	for key, value := range data {
		fmt.Printf("  %s: ", key)
		switch v := value.(type) {
		case nil:
			fmt.Println("null")
		case bool:
			fmt.Printf("布尔值 %t\n", v)
		case float64:
			fmt.Printf("数字 %.2f\n", v)
		case string:
			fmt.Printf("字符串 \"%s\"\n", v)
		case []interface{}:
			fmt.Printf("数组 [长度=%d]\n", len(v))
		case map[string]interface{}:
			fmt.Printf("对象 [键数=%d]\n", len(v))
		default:
			fmt.Printf("未知类型 %T\n", v)
		}
	}
}

func getStringField(data map[string]interface{}, key, defaultValue string) string {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntField(data map[string]interface{}, key string, defaultValue int) int {
	if value, exists := data[key]; exists {
		if num, ok := value.(float64); ok { // JSON数字通常是float64
			return int(num)
		}
		if num, ok := value.(int); ok {
			return num
		}
	}
	return defaultValue
}

func getBoolField(data map[string]interface{}, key string, defaultValue bool) bool {
	if value, exists := data[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func getFloatField(data map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := data[key]; exists {
		if num, ok := value.(float64); ok {
			return num
		}
		if num, ok := value.(int); ok {
			return float64(num)
		}
	}
	return defaultValue
}

// 反射基础
func demonstrateReflectionBasics() {
	fmt.Println("6. 反射基础:")

	values := []interface{}{
		42,
		"hello",
		3.14159,
		true,
		[]int{1, 2, 3},
		map[string]int{"a": 1},
		Person{Name: "张三", Age: 25},
	}

	fmt.Println("反射信息:")
	for i, value := range values {
		t := reflect.TypeOf(value)
		v := reflect.ValueOf(value)

		fmt.Printf("  [%d] 值: %v\n", i, value)
		fmt.Printf("      类型: %v\n", t)
		fmt.Printf("      种类: %v\n", t.Kind())
		fmt.Printf("      可设置: %t\n", v.CanSet())

		// 特殊类型的额外信息
		switch t.Kind() {
		case reflect.Slice:
			fmt.Printf("      切片长度: %d\n", v.Len())
		case reflect.Map:
			fmt.Printf("      映射键数: %d\n", v.Len())
		case reflect.Struct:
			fmt.Printf("      字段数: %d\n", t.NumField())
		}
		fmt.Println()
	}

	// 反射修改值
	fmt.Println("反射修改值:")
	var x interface{} = 42
	v := reflect.ValueOf(&x).Elem() // 获取可修改的值
	if v.CanSet() {
		v.Set(reflect.ValueOf(100))
		fmt.Printf("修改后的值: %v\n", x)
	}

	fmt.Println()
}

// 性能考虑
func demonstratePerformanceConsiderations() {
	fmt.Println("7. 性能考虑:")

	// 类型断言 vs 反射性能对比
	var value interface{} = 42

	// 类型断言（快）
	start := time.Now()
	for i := 0; i < 1000000; i++ {
		if _, ok := value.(int); ok {
			// 类型断言
		}
	}
	assertionTime := time.Since(start)

	// 反射（慢）
	start = time.Now()
	for i := 0; i < 1000000; i++ {
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Int {
			// 反射检查
		}
	}
	reflectionTime := time.Since(start)

	fmt.Printf("类型断言耗时: %v\n", assertionTime)
	fmt.Printf("反射检查耗时: %v\n", reflectionTime)
	fmt.Printf("性能差异: %.2fx\n", float64(reflectionTime)/float64(assertionTime))

	// 装箱拆箱开销
	fmt.Println("\n装箱拆箱示例:")
	var numbers []interface{}
	start = time.Now()
	for i := 0; i < 100000; i++ {
		numbers = append(numbers, i) // 装箱
	}
	boxingTime := time.Since(start)

	start = time.Now()
	sum := 0
	for _, num := range numbers {
		if n, ok := num.(int); ok { // 拆箱
			sum += n
		}
	}
	unboxingTime := time.Since(start)

	fmt.Printf("装箱耗时: %v\n", boxingTime)
	fmt.Printf("拆箱耗时: %v\n", unboxingTime)
	fmt.Printf("计算结果: %d\n", sum)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 配置系统
	fmt.Println("配置系统:")
	config := NewConfigManager()

	config.Set("database.host", "localhost")
	config.Set("database.port", 5432)
	config.Set("app.debug", true)
	config.Set("app.version", "1.0.0")

	fmt.Printf("数据库主机: %s\n", config.GetString("database.host", "127.0.0.1"))
	fmt.Printf("数据库端口: %d\n", config.GetInt("database.port", 3306))
	fmt.Printf("调试模式: %t\n", config.GetBool("app.debug", false))

	// 2. 事件系统
	fmt.Println("\n事件系统:")
	eventBus := NewEventBus()

	// 注册事件处理器
	eventBus.Subscribe("user.login", func(data interface{}) {
		if userMap, ok := data.(map[string]interface{}); ok {
			name := getStringField(userMap, "name", "未知用户")
			fmt.Printf("  用户登录: %s\n", name)
		}
	})

	eventBus.Subscribe("user.logout", func(data interface{}) {
		if userMap, ok := data.(map[string]interface{}); ok {
			name := getStringField(userMap, "name", "未知用户")
			fmt.Printf("  用户登出: %s\n", name)
		}
	})

	// 发布事件
	eventBus.Publish("user.login", map[string]interface{}{
		"name": "张三",
		"id":   123,
	})

	eventBus.Publish("user.logout", map[string]interface{}{
		"name": "张三",
		"id":   123,
	})

	// 3. 数据转换工具
	fmt.Println("\n数据转换工具:")
	converter := NewDataConverter()

	// 字符串转换
	fmt.Printf("字符串转整数: %d\n", converter.ToInt("42", 0))
	fmt.Printf("字符串转浮点: %.2f\n", converter.ToFloat("3.14", 0.0))
	fmt.Printf("字符串转布尔: %t\n", converter.ToBool("true", false))

	// 数组转换
	mixed := []interface{}{1, "2", 3.0, "4", 5}
	ints := converter.ToIntSlice(mixed)
	fmt.Printf("混合数组转整数数组: %v\n", ints)

	// 4. 通用验证器
	fmt.Println("\n通用验证器:")
	validator := NewValidator()

	testData := map[string]interface{}{
		"name":  "张三",
		"age":   25,
		"email": "zhang@example.com",
		"score": 95.5,
	}

	validator.AddRule("name", func(value interface{}) bool {
		if str, ok := value.(string); ok && len(str) > 0 {
			return true
		}
		return false
	})

	validator.AddRule("age", func(value interface{}) bool {
		if age := getIntField(map[string]interface{}{"age": value}, "age", 0); age >= 18 && age <= 100 {
			return true
		}
		return false
	})

	if validator.Validate(testData) {
		fmt.Println("数据验证通过")
	} else {
		fmt.Println("数据验证失败")
	}

	// 5. 通用序列化器
	fmt.Println("\n通用序列化器:")
	serializer := NewSerializer()

	data := map[string]interface{}{
		"user":      "张三",
		"timestamp": time.Now().Unix(),
		"active":    true,
		"metadata":  map[string]string{"role": "admin"},
	}

	serialized := serializer.Serialize(data)
	fmt.Printf("序列化结果: %s\n", serialized)

	deserialized := serializer.Deserialize(serialized)
	fmt.Printf("反序列化结果: %v\n", deserialized)

	fmt.Println()
}

// 配置管理器
type ConfigManager struct {
	data map[string]interface{}
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{data: make(map[string]interface{})}
}

func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.data[key] = value
}

func (cm *ConfigManager) GetString(key, defaultValue string) string {
	if value, exists := cm.data[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (cm *ConfigManager) GetInt(key string, defaultValue int) int {
	if value, exists := cm.data[key]; exists {
		if num, ok := value.(int); ok {
			return num
		}
		if num, ok := value.(float64); ok {
			return int(num)
		}
	}
	return defaultValue
}

func (cm *ConfigManager) GetBool(key string, defaultValue bool) bool {
	if value, exists := cm.data[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// 事件总线
type EventBus struct {
	handlers map[string][]func(interface{})
}

func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[string][]func(interface{}))}
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

// 数据转换器
type DataConverter struct{}

func NewDataConverter() *DataConverter {
	return &DataConverter{}
}

func (dc *DataConverter) ToInt(value interface{}, defaultValue int) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if num, err := strconv.Atoi(v); err == nil {
			return num
		}
	}
	return defaultValue
}

func (dc *DataConverter) ToFloat(value interface{}, defaultValue float64) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num
		}
	}
	return defaultValue
}

func (dc *DataConverter) ToBool(value interface{}, defaultValue bool) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

func (dc *DataConverter) ToIntSlice(values []interface{}) []int {
	var result []int
	for _, value := range values {
		result = append(result, dc.ToInt(value, 0))
	}
	return result
}

// 通用验证器
type Validator struct {
	rules map[string]func(interface{}) bool
}

func NewValidator() *Validator {
	return &Validator{rules: make(map[string]func(interface{}) bool)}
}

func (v *Validator) AddRule(field string, rule func(interface{}) bool) {
	v.rules[field] = rule
}

func (v *Validator) Validate(data map[string]interface{}) bool {
	for field, rule := range v.rules {
		if value, exists := data[field]; exists {
			if !rule(value) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// 通用序列化器
type Serializer struct{}

func NewSerializer() *Serializer {
	return &Serializer{}
}

func (s *Serializer) Serialize(data interface{}) string {
	return fmt.Sprintf("%v", data)
}

func (s *Serializer) Deserialize(data string) map[string]interface{} {
	// 简化实现，实际应该解析字符串
	return map[string]interface{}{"serialized": data}
}

/*
=== 练习题 ===

1. 实现一个通用的深拷贝函数，支持任意类型

2. 创建一个类型安全的JSON解析器，减少类型断言

3. 实现一个通用的数据验证框架

4. 设计一个支持多种数据源的配置系统

5. 创建一个通用的对象池，支持任意类型

6. 实现一个类型安全的事件系统

7. 设计一个通用的序列化/反序列化框架

运行命令：
go run main.go

高级练习：
1. 实现一个反射式的ORM框架
2. 创建一个通用的RPC系统
3. 设计一个插件化的数据处理管道
4. 实现一个类型安全的模板引擎
5. 创建一个通用的缓存代理系统

重要概念：
- 空接口是所有类型的超集
- 类型断言的安全使用
- 反射的基础应用
- 性能权衡考虑
- 通用编程的实践
- JSON数据的处理模式
*/
