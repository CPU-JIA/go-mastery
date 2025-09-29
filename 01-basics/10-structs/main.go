// Package main demonstrates struct usage in Go language.
// This module covers struct definition, initialization, methods,
// embedding, tags, and practical examples.
package main

import (
	"fmt"
	"time"
)

const (
	// DefaultAge represents the default age value for examples.
	DefaultAge = 25
	// MiddleAge represents a middle-aged person's age.
	MiddleAge = 30
	// SampleAge represents a sample age value for demonstrations.
	SampleAge = 32

	// DefaultTimeout represents the default timeout value in seconds.
	DefaultTimeout = 30
	// ModifiedAge represents the modified age value for testing.
	ModifiedAge = 999
)

const (
	// ModifiedName represents the modified name value for testing.
	ModifiedName = "Modified"
)

/*
=== Go语言第十课：结构体(Structs) ===

学习目标：
1. 理解结构体的概念和用途
2. 掌握结构体的定义和初始化
3. 学会结构体的字段访问和修改
4. 了解结构体的嵌套和匿名字段
5. 掌握结构体的方法和指针接收者

Go结构体特点：
- 值类型，可以包含不同类型的字段
- 支持字段标签（tags）
- 支持匿名字段和嵌套
- 可以定义方法
- 零值是所有字段的零值
*/

func main() {
	fmt.Println("=== Go语言结构体学习 ===")

	// 1. 结构体定义和初始化
	demonstrateStructDefinition()

	// 2. 结构体操作
	demonstrateStructOperations()

	// 3. 结构体指针
	demonstrateStructPointers()

	// 4. 匿名结构体
	demonstrateAnonymousStructs()

	// 5. 嵌套结构体
	demonstrateNestedStructs()

	// 6. 匿名字段和嵌入
	demonstrateEmbeddedStructs()

	// 7. 结构体标签
	demonstrateStructTags()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// Person represents a basic person with name, age and city information.
type Person struct {
	Name string
	Age  int
	City string
}

// Point represents a 2D coordinate point with X and Y values.
type Point struct {
	X, Y float64
}

// Rectangle represents a rectangle with width and height dimensions.
type Rectangle struct {
	Width, Height float64
}

// CacheItem represents a cached item with value and timestamp information.
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Cache represents an in-memory cache with TTL support.
type Cache struct {
	items map[string]CacheItem
	ttl   time.Duration
}

// 结构体定义和初始化
func demonstrateStructDefinition() {
	fmt.Println("1. 结构体定义和初始化:")

	// 1. 零值初始化
	var p1 Person
	fmt.Printf("零值结构体: %+v\n", p1) // %+v 显示字段名

	// 2. 字面量初始化（按字段顺序）
	p2 := Person{"张三", DefaultAge, "北京"}
	fmt.Printf("按顺序初始化: %+v\n", p2)

	// 3. 字段名初始化（推荐）
	p3 := Person{
		Name: "李四",
		Age:  MiddleAge,
		City: "上海",
	}
	fmt.Printf("按字段名初始化: %+v\n", p3)

	// 4. 部分字段初始化
	p4 := Person{
		Name: "王五",
		City: "广州", // Age使用零值
	}
	fmt.Printf("部分初始化: %+v\n", p4)

	// 5. new函数创建
	p5 := new(Person)
	p5.Name = "赵六"
	p5.Age = 35
	fmt.Printf("new创建: %+v\n", *p5)

	// 6. 取地址创建
	p6 := &Person{
		Name: "孙七",
		Age:  28,
		City: "深圳",
	}
	fmt.Printf("取地址创建: %+v\n", *p6)

	// 7. 不同类型的字段
	type Employee struct {
		ID       int
		Name     string
		Salary   float64
		IsActive bool
		Skills   []string
		JoinDate time.Time
	}

	emp := Employee{
		ID:       1001,
		Name:     "张工程师",
		Salary:   15000.50,
		IsActive: true,
		Skills:   []string{"Go", "Python", "Docker"},
		JoinDate: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	fmt.Printf("复杂结构体: %+v\n", emp)

	fmt.Println()
}

// 结构体操作
func demonstrateStructOperations() {
	fmt.Println("2. 结构体操作:")

	// 创建结构体
	person := Person{
		Name: "Alice",
		Age:  25,
		City: "北京",
	}

	fmt.Printf("原始结构体: %+v\n", person)

	// 访问字段
	fmt.Printf("姓名: %s\n", person.Name)
	fmt.Printf("年龄: %d\n", person.Age)
	fmt.Printf("城市: %s\n", person.City)

	// 修改字段
	person.Age = 26
	person.City = "上海"
	fmt.Printf("修改后: %+v\n", person)

	// 结构体比较（相同类型且所有字段可比较）
	person1 := Person{"Bob", 30, "广州"}
	person2 := Person{"Bob", 30, "广州"}
	person3 := Person{"Charlie", 30, "广州"}

	fmt.Printf("person1 == person2: %t\n", person1 == person2)
	fmt.Printf("person1 == person3: %t\n", person1 == person3)

	// 结构体复制（值复制）
	originalPoint := Point{X: 1.0, Y: 2.0}
	copiedPoint := originalPoint

	fmt.Printf("原始点: %+v\n", originalPoint)
	fmt.Printf("复制点: %+v\n", copiedPoint)

	// 修改复制的结构体不影响原始结构体
	copiedPoint.X = 10.0
	fmt.Printf("修改复制后:\n")
	fmt.Printf("  原始点: %+v\n", originalPoint)
	fmt.Printf("  复制点: %+v\n", copiedPoint)

	// 结构体作为map的键
	pointMap := make(map[Point]string)
	pointMap[Point{0, 0}] = "原点"
	pointMap[Point{1, 1}] = "对角点"
	pointMap[Point{0, 1}] = "Y轴点"

	fmt.Printf("结构体作为键: %v\n", pointMap)

	// 查找
	if desc, exists := pointMap[Point{0, 0}]; exists {
		fmt.Printf("点(0,0): %s\n", desc)
	}

	fmt.Println()
}

// 结构体指针
func demonstrateStructPointers() {
	fmt.Println("3. 结构体指针:")

	// 创建结构体指针
	person := &Person{
		Name: "David",
		Age:  35,
		City: "深圳",
	}

	fmt.Printf("结构体指针: %p\n", person)
	fmt.Printf("指针指向的值: %+v\n", *person)

	// 通过指针访问字段（Go自动解引用）
	fmt.Printf("姓名: %s\n", person.Name) // 等价于 (*person).Name
	fmt.Printf("年龄: %d\n", person.Age)

	// 通过指针修改字段
	person.Age = 36
	person.City = "杭州"
	fmt.Printf("修改后: %+v\n", *person)

	// 指针传递vs值传递
	fmt.Println("\n指针传递vs值传递:")

	original := Person{"Eve", 28, "成都"}
	fmt.Printf("函数调用前: %+v\n", original)

	// 值传递（不会修改原结构体）
	modifyPersonByValue(original)
	fmt.Printf("值传递后: %+v\n", original)

	// 指针传递（会修改原结构体）
	modifyPersonByPointer(&original)
	fmt.Printf("指针传递后: %+v\n", original)

	// 返回结构体指针
	newPerson := createPerson("Frank", SampleAge, "重庆")
	fmt.Printf("创建的新人: %+v\n", *newPerson)

	// 指针数组
	people := []*Person{
		{"Alice", 25, "北京"},
		{"Bob", 30, "上海"},
		{"Charlie", 35, "广州"},
	}

	fmt.Println("\n指针数组:")
	for i, p := range people {
		fmt.Printf("  %d: %+v\n", i, *p)
	}

	// 修改指针数组中的元素
	people[0].Age = 26
	fmt.Printf("修改后第一个人: %+v\n", *people[0])

	fmt.Println()
}

// 匿名结构体
func demonstrateAnonymousStructs() {
	fmt.Println("4. 匿名结构体:")

	// 临时使用的结构体
	config := struct {
		Host     string
		Port     int
		Database string
		SSL      bool
	}{
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		SSL:      true,
	}

	fmt.Printf("配置信息: %+v\n", config)

	// 匿名结构体切片
	tasks := []struct {
		ID          int
		Description string
		Completed   bool
	}{
		{1, "学习Go语言", false},
		{2, "编写代码", true},
		{3, "测试程序", false},
	}

	fmt.Println("任务列表:")
	for _, task := range tasks {
		status := "未完成"
		if task.Completed {
			status = "已完成"
		}
		fmt.Printf("  %d: %s [%s]\n", task.ID, task.Description, status)
	}

	// 匿名结构体作为函数参数
	processData(struct {
		Name  string
		Value int
	}{
		Name:  "测试数据",
		Value: 42,
	})

	// 匿名结构体作为map值
	serverStatus := map[string]struct {
		Online   bool
		Load     float64
		LastPing time.Time
	}{
		"web-1": {true, 0.85, time.Now()},
		"web-2": {false, 0.0, time.Now().Add(-time.Hour)},
		"db-1":  {true, 0.45, time.Now()},
	}

	fmt.Println("\n服务器状态:")
	for server, status := range serverStatus {
		onlineStatus := "离线"
		if status.Online {
			onlineStatus = "在线"
		}
		fmt.Printf("  %s: %s (负载: %.2f)\n", server, onlineStatus, status.Load)
	}

	fmt.Println()
}

// 嵌套结构体
func demonstrateNestedStructs() {
	fmt.Println("5. 嵌套结构体:")

	// 定义嵌套结构体
	type Address struct {
		Street   string
		City     string
		Province string
		ZipCode  string
	}

	type Contact struct {
		Email string
		Phone string
	}

	type Student struct {
		ID      int
		Name    string
		Age     int
		Address Address // 嵌套结构体
		Contact Contact // 嵌套结构体
		Scores  map[string]int
	}

	// 创建嵌套结构体
	student := Student{
		ID:   2023001,
		Name: "张同学",
		Age:  20,
		Address: Address{
			Street:   "清华大学路1号",
			City:     "北京",
			Province: "北京",
			ZipCode:  "100084",
		},
		Contact: Contact{
			Email: "zhang@example.com",
			Phone: "13800138000",
		},
		Scores: map[string]int{
			"数学": 95,
			"英语": 88,
			"物理": 92,
		},
	}

	fmt.Printf("学生信息: %+v\n", student)

	// 访问嵌套字段
	fmt.Printf("姓名: %s\n", student.Name)
	fmt.Printf("地址: %s, %s\n", student.Address.Street, student.Address.City)
	fmt.Printf("邮箱: %s\n", student.Contact.Email)
	fmt.Printf("数学成绩: %d\n", student.Scores["数学"])

	// 修改嵌套字段
	student.Address.City = "上海"
	student.Contact.Phone = "13900139000"
	fmt.Printf("修改后地址: %s\n", student.Address.City)
	fmt.Printf("修改后电话: %s\n", student.Contact.Phone)

	// 部分初始化嵌套结构体
	student2 := Student{
		Name: "李同学",
		Address: Address{
			City: "广州",
		},
		Scores: make(map[string]int),
	}

	fmt.Printf("部分初始化: %+v\n", student2)

	fmt.Println()
}

// 匿名字段和嵌入
func demonstrateEmbeddedStructs() {
	fmt.Println("6. 匿名字段和嵌入:")

	// 定义基础结构体
	type Animal struct {
		Name string
		Age  int
	}

	type Mammal struct {
		Animal   // 匿名字段（嵌入）
		FurColor string
	}

	type Dog struct {
		Mammal // 嵌入Mammal
		Breed  string
	}

	// 创建嵌入结构体
	dog := Dog{
		Mammal: Mammal{
			Animal: Animal{
				Name: "小白",
				Age:  3,
			},
			FurColor: "白色",
		},
		Breed: "金毛",
	}

	fmt.Printf("狗的信息: %+v\n", dog)

	// 直接访问嵌入的字段（字段提升）
	fmt.Printf("名字: %s\n", dog.Name)     // 等价于 dog.Animal.Name
	fmt.Printf("年龄: %d\n", dog.Age)      // 等价于 dog.Animal.Age
	fmt.Printf("毛色: %s\n", dog.FurColor) // 等价于 dog.Mammal.FurColor
	fmt.Printf("品种: %s\n", dog.Breed)

	// 也可以通过完整路径访问
	fmt.Printf("完整路径访问名字: %s\n", dog.Mammal.Animal.Name)

	// 修改嵌入字段
	dog.Name = "小黑" // 等价于 dog.Animal.Name = "小黑"
	dog.Age = 4
	fmt.Printf("修改后: 名字=%s, 年龄=%d\n", dog.Name, dog.Age)

	// 字段名冲突的处理
	type Base struct {
		ID   int
		Name string
	}

	type Extended struct {
		Base
		ID   string // 与Base.ID冲突
		Type string
	}

	ext := Extended{
		Base: Base{
			ID:   123,
			Name: "基础名称",
		},
		ID:   "EXT001", // 这个ID会隐藏Base.ID
		Type: "扩展类型",
	}

	fmt.Printf("扩展结构体: %+v\n", ext)
	fmt.Printf("外层ID: %s\n", ext.ID)      // Extended.ID
	fmt.Printf("内层ID: %d\n", ext.Base.ID) // Base.ID
	fmt.Printf("名称: %s\n", ext.Name)      // 提升的Base.Name

	// 多重嵌入
	type Timestamp struct {
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	type User struct {
		Person    // 嵌入Person
		Timestamp // 嵌入Timestamp
		Email     string
	}

	user := User{
		Person: Person{
			Name: "用户",
			Age:  25,
			City: "北京",
		},
		Timestamp: Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Email: "user@example.com",
	}

	fmt.Printf("用户信息: %+v\n", user)
	fmt.Printf("创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	fmt.Println()
}

// 结构体标签
func demonstrateStructTags() {
	fmt.Println("7. 结构体标签:")

	// 定义带标签的结构体
	type Product struct {
		ID          int      `json:"id" xml:"product_id" db:"product_id"`
		Name        string   `json:"name" xml:"product_name" db:"name"`
		Price       float64  `json:"price" xml:"price" db:"price"`
		Description string   `json:"description,omitempty" xml:"desc,omitempty"`
		IsAvailable bool     `json:"is_available" xml:"available" db:"is_available"`
		Tags        []string `json:"tags,omitempty" xml:"tags>tag"`
	}

	product := Product{
		ID:          1,
		Name:        "智能手机",
		Price:       2999.99,
		Description: "最新款智能手机",
		IsAvailable: true,
		Tags:        []string{"电子产品", "通讯设备"},
	}

	fmt.Printf("产品信息: %+v\n", product)

	// 反射获取标签信息
	showStructTags(product)

	// 验证标签
	type User struct {
		Name     string `validate:"required,min=2,max=50"`
		Email    string `validate:"required,email"`
		Age      int    `validate:"min=18,max=120"`
		Password string `validate:"required,min=8"`
	}

	user := User{
		Name:     "张三",
		Email:    "zhang@example.com",
		Age:      25,
		Password: "password123",
	}

	fmt.Printf("\n用户信息: %+v\n", user)
	validateStruct(user)

	// 数据库标签示例
	type Employee struct {
		ID        int       `db:"id,primary_key,auto_increment"`
		FirstName string    `db:"first_name,not_null"`
		LastName  string    `db:"last_name,not_null"`
		Email     string    `db:"email,unique"`
		Salary    float64   `db:"salary"`
		HiredAt   time.Time `db:"hired_at,default=NOW()"`
	}

	emp := Employee{
		ID:        1,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@company.com",
		Salary:    75000.00,
		HiredAt:   time.Now(),
	}

	fmt.Printf("\n员工信息: %+v\n", emp)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 配置结构体
	type DatabaseConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		SSL      bool   `json:"ssl"`
	}

	type ServerConfig struct {
		Port         int            `json:"port"`
		Host         string         `json:"host"`
		Debug        bool           `json:"debug"`
		Database     DatabaseConfig `json:"database"`
		AllowedHosts []string       `json:"allowed_hosts"`
	}

	config := ServerConfig{
		Port:  8080,
		Host:  "0.0.0.0",
		Debug: true,
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "app",
			Password: "secret",
			Database: "myapp",
			SSL:      false,
		},
		AllowedHosts: []string{"localhost", "127.0.0.1"},
	}

	fmt.Printf("服务器配置: %+v\n", config)

	// 2. 业务模型
	type OrderItem struct {
		ProductID string  `json:"product_id"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
		Subtotal  float64 `json:"subtotal"`
	}

	type Order struct {
		ID          string      `json:"id"`
		CustomerID  string      `json:"customer_id"`
		Items       []OrderItem `json:"items"`
		TotalAmount float64     `json:"total_amount"`
		Status      string      `json:"status"`
		CreatedAt   time.Time   `json:"created_at"`
		UpdatedAt   time.Time   `json:"updated_at"`
	}

	order := Order{
		ID:         "ORD-2023-001",
		CustomerID: "CUST-001",
		Items: []OrderItem{
			{
				ProductID: "PROD-001",
				Quantity:  2,
				Price:     999.99,
				Subtotal:  1999.98,
			},
			{
				ProductID: "PROD-002",
				Quantity:  1,
				Price:     299.99,
				Subtotal:  299.99,
			},
		},
		TotalAmount: 2299.97,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	fmt.Printf("\n订单信息: %+v\n", order)

	// 计算订单总价
	calculatedTotal := 0.0
	for _, item := range order.Items {
		calculatedTotal += item.Subtotal
	}
	fmt.Printf("计算的总价: %.2f\n", calculatedTotal)

	// 3. 状态机
	type State struct {
		Name        string
		OnEnter     func()
		OnExit      func()
		Transitions map[string]string
	}

	type StateMachine struct {
		CurrentState string
		States       map[string]State
	}

	// 创建简单状态机
	machine := StateMachine{
		CurrentState: "idle",
		States: map[string]State{
			"idle": {
				Name:    "空闲",
				OnEnter: func() { fmt.Println("进入空闲状态") },
				OnExit:  func() { fmt.Println("离开空闲状态") },
				Transitions: map[string]string{
					"start": "running",
				},
			},
			"running": {
				Name:    "运行中",
				OnEnter: func() { fmt.Println("进入运行状态") },
				OnExit:  func() { fmt.Println("离开运行状态") },
				Transitions: map[string]string{
					"pause": "paused",
					"stop":  "idle",
				},
			},
			"paused": {
				Name:    "暂停",
				OnEnter: func() { fmt.Println("进入暂停状态") },
				OnExit:  func() { fmt.Println("离开暂停状态") },
				Transitions: map[string]string{
					"resume": "running",
					"stop":   "idle",
				},
			},
		},
	}

	fmt.Println("\n状态机演示:")
	fmt.Printf("当前状态: %s\n", machine.CurrentState)

	// 状态转换
	transitions := []string{"start", "pause", "resume", "stop"}
	for _, event := range transitions {
		if newState, exists := machine.States[machine.CurrentState].Transitions[event]; exists {
			machine.States[machine.CurrentState].OnExit()
			machine.CurrentState = newState
			machine.States[machine.CurrentState].OnEnter()
			fmt.Printf("事件: %s -> 新状态: %s\n", event, machine.CurrentState)
		} else {
			fmt.Printf("事件 %s 在状态 %s 中无效\n", event, machine.CurrentState)
		}
	}

	// 4. 缓存系统
	cache := &Cache{
		items: make(map[string]CacheItem),
		ttl:   time.Minute * 5,
	}

	// 设置缓存
	cache.Set("user:123", map[string]string{"name": "张三", "email": "zhang@example.com"})
	cache.Set("config:timeout", DefaultTimeout)

	fmt.Println("\n缓存演示:")

	// 获取缓存
	if value, found := cache.Get("user:123"); found {
		fmt.Printf("缓存命中: %v\n", value)
	}

	if value, found := cache.Get("nonexistent"); !found {
		fmt.Printf("缓存未命中: %v\n", value)
	}

	fmt.Println()
}

// 辅助函数

// 值传递修改（不会影响原结构体）
func modifyPersonByValue(p Person) {
	p.Age = ModifiedAge
	p.Name = ModifiedName
}

// 指针传递修改（会影响原结构体）
func modifyPersonByPointer(p *Person) {
	p.Age = ModifiedAge
	p.Name = "Modified"
}

// 创建并返回Person指针
func createPerson(name string, age int, city string) *Person {
	return &Person{
		Name: name,
		Age:  age,
		City: city,
	}
}

// 处理匿名结构体
func processData(data struct {
	Name  string
	Value int
}) {
	fmt.Printf("处理数据: %s = %d\n", data.Name, data.Value)
}

// 显示结构体标签（使用反射）
func showStructTags(_ interface{}) {
	fmt.Println("结构体标签信息:")
	// 这里应该使用reflect包，为简化示例，只是打印说明
	fmt.Println("  ID字段: json:\"id\", xml:\"product_id\", db:\"product_id\"")
	fmt.Println("  Name字段: json:\"name\", xml:\"product_name\", db:\"name\"")
	fmt.Println("  Description字段: json:\"description,omitempty\"")
}

// 简单的验证函数
func validateStruct(_ interface{}) {
	fmt.Println("结构体验证:")
	fmt.Println("  ✓ 所有必需字段都已填写")
	fmt.Println("  ✓ 邮箱格式正确")
	fmt.Println("  ✓ 年龄在有效范围内")
	fmt.Println("  ✓ 密码长度符合要求")
}

// Set stores a value in the cache with the configured TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.items[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
		CreatedAt: time.Now(),
	}
}

// Get retrieves a value from the cache, returning the value and whether it was found.
func (c *Cache) Get(key string) (interface{}, bool) {
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		delete(c.items, key)
		return nil, false
	}

	return item.Value, true
}

/*
=== 练习题 ===

1. 设计一个图书管理系统的结构体（Book, Author, Library等）

2. 实现一个员工管理系统，包含部门嵌套

3. 创建一个电商系统的产品结构体，支持变体和属性

4. 设计一个博客系统的数据结构（Post, Comment, User等）

5. 实现一个游戏角色系统，使用嵌入实现不同类型的角色

6. 创建一个配置管理系统，支持不同环境的配置

7. 设计一个任务队列系统的数据结构

运行命令：
go run main.go

高级练习：
1. 实现一个ORM映射系统
2. 创建一个事件溯源系统
3. 设计一个插件系统架构
4. 实现一个简单的序列化系统
5. 创建一个领域驱动设计的聚合根

重要概念：
- 结构体是值类型
- 支持嵌入和组合
- 标签用于元数据
- 方法可以定义在结构体上
- 零值是所有字段的零值
- 可以作为map的键（如果所有字段都可比较）
*/
