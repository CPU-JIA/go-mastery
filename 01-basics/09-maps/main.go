package main

import (
	"fmt"
	"sort"
)

/*
=== Go语言第九课：映射(Maps) ===

学习目标：
1. 理解映射的概念和特点
2. 掌握映射的创建和初始化
3. 学会映射的增删改查操作
4. 了解映射的遍历和排序
5. 掌握映射的高级用法

Go映射特点：
- 键值对的无序集合
- 引用类型，零值为nil
- 键必须是可比较的类型
- 值可以是任意类型
- 并发不安全
*/

func main() {
	fmt.Println("=== Go语言映射学习 ===")

	// 1. 映射创建和初始化
	demonstrateMapCreation()

	// 2. 映射基本操作
	demonstrateMapOperations()

	// 3. 映射遍历
	demonstrateMapIteration()

	// 4. 映射的值类型
	demonstrateMapValueTypes()

	// 5. 嵌套映射
	demonstrateNestedMaps()

	// 6. 映射作为集合
	demonstrateMapAsSet()

	// 7. 映射作为函数参数
	demonstrateMapAsParameter()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 映射创建和初始化
func demonstrateMapCreation() {
	fmt.Println("1. 映射创建和初始化:")

	// 1. nil映射
	var nilMap map[string]int
	fmt.Printf("nil映射: %v, 长度: %d, 是否为nil: %t\n",
		nilMap, len(nilMap), nilMap == nil)

	// 2. make函数创建
	makeMap := make(map[string]int)
	fmt.Printf("make映射: %v, 长度: %d, 是否为nil: %t\n",
		makeMap, len(makeMap), makeMap == nil)

	// 3. 字面量创建
	colors := map[string]string{
		"red":   "红色",
		"green": "绿色",
		"blue":  "蓝色",
	}
	fmt.Printf("字面量映射: %v, 长度: %d\n", colors, len(colors))

	// 4. 空映射字面量
	emptyMap := map[string]int{}
	fmt.Printf("空映射: %v, 长度: %d\n", emptyMap, len(emptyMap))

	// 5. 带容量提示的make（仅用于优化）
	capacityMap := make(map[string]int, 100)
	fmt.Printf("带容量提示: %v, 长度: %d\n", capacityMap, len(capacityMap))

	// 6. 不同类型的键
	intKeyMap := map[int]string{1: "one", 2: "two", 3: "three"}
	floatKeyMap := map[float64]bool{3.14: true, 2.71: false}
	boolKeyMap := map[bool]string{true: "是", false: "否"}

	fmt.Printf("整数键: %v\n", intKeyMap)
	fmt.Printf("浮点键: %v\n", floatKeyMap)
	fmt.Printf("布尔键: %v\n", boolKeyMap)

	// 7. 结构体作为键
	type Point struct {
		X, Y int
	}

	pointMap := map[Point]string{
		{0, 0}: "原点",
		{1, 1}: "对角点",
		{0, 1}: "Y轴点",
	}
	fmt.Printf("结构体键: %v\n", pointMap)

	fmt.Println()
}

// 映射基本操作
func demonstrateMapOperations() {
	fmt.Println("2. 映射基本操作:")

	// 创建映射
	students := make(map[string]int)
	fmt.Printf("初始映射: %v\n", students)

	// 添加元素
	students["Alice"] = 85
	students["Bob"] = 92
	students["Charlie"] = 78
	fmt.Printf("添加元素后: %v\n", students)

	// 访问元素
	aliceScore := students["Alice"]
	fmt.Printf("Alice的分数: %d\n", aliceScore)

	// 访问不存在的键（返回零值）
	davidScore := students["David"]
	fmt.Printf("David的分数: %d (不存在的键)\n", davidScore)

	// 检查键是否存在
	score, exists := students["Alice"]
	if exists {
		fmt.Printf("Alice存在，分数: %d\n", score)
	}

	score, exists = students["David"]
	if !exists {
		fmt.Printf("David不存在\n")
	}

	// 修改元素
	students["Alice"] = 90
	fmt.Printf("修改Alice分数后: %v\n", students)

	// 删除元素
	delete(students, "Bob")
	fmt.Printf("删除Bob后: %v\n", students)

	// 删除不存在的键（不会报错）
	delete(students, "NonExistent")
	fmt.Printf("删除不存在键后: %v\n", students)

	// 清空映射
	for key := range students {
		delete(students, key)
	}
	fmt.Printf("清空后: %v\n", students)

	fmt.Println()
}

// 映射遍历
func demonstrateMapIteration() {
	fmt.Println("3. 映射遍历:")

	countries := map[string]string{
		"CN": "中国",
		"US": "美国",
		"JP": "日本",
		"UK": "英国",
		"FR": "法国",
	}

	// 遍历键值对
	fmt.Println("遍历键值对:")
	for code, name := range countries {
		fmt.Printf("  %s: %s\n", code, name)
	}

	// 只遍历键
	fmt.Println("\n只遍历键:")
	for code := range countries {
		fmt.Printf("  %s\n", code)
	}

	// 只遍历值
	fmt.Println("\n只遍历值:")
	for _, name := range countries {
		fmt.Printf("  %s\n", name)
	}

	// 有序遍历（先排序键）
	fmt.Println("\n有序遍历:")
	var keys []string
	for code := range countries {
		keys = append(keys, code)
	}
	sort.Strings(keys)

	for _, code := range keys {
		fmt.Printf("  %s: %s\n", code, countries[code])
	}

	// 映射是无序的（多次运行可能顺序不同）
	fmt.Println("\n映射遍历的无序性演示:")
	for i := 0; i < 3; i++ {
		fmt.Printf("第%d次遍历: ", i+1)
		count := 0
		for code := range countries {
			if count < 3 { // 只显示前3个
				fmt.Printf("%s ", code)
				count++
			}
		}
		fmt.Println("...")
	}

	fmt.Println()
}

// 映射的值类型
func demonstrateMapValueTypes() {
	fmt.Println("4. 映射的值类型:")

	// 切片作为值
	studentGrades := map[string][]int{
		"Alice":   {85, 92, 78},
		"Bob":     {90, 85, 88},
		"Charlie": {75, 80, 85},
	}

	fmt.Println("切片作为值:")
	for student, grades := range studentGrades {
		total := 0
		for _, grade := range grades {
			total += grade
		}
		average := float64(total) / float64(len(grades))
		fmt.Printf("  %s: %v (平均: %.1f)\n", student, grades, average)
	}

	// 映射作为值
	cityInfo := map[string]map[string]interface{}{
		"北京": {
			"人口":  2154,
			"面积":  16410,
			"是首都": true,
		},
		"上海": {
			"人口":  2424,
			"面积":  6340,
			"是首都": false,
		},
	}

	fmt.Println("\n映射作为值:")
	for city, info := range cityInfo {
		fmt.Printf("  %s:\n", city)
		for key, value := range info {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	// 结构体作为值
	type Person struct {
		Name string
		Age  int
		City string
	}

	people := map[string]Person{
		"001": {"张三", 25, "北京"},
		"002": {"李四", 30, "上海"},
		"003": {"王五", 28, "广州"},
	}

	fmt.Println("\n结构体作为值:")
	for id, person := range people {
		fmt.Printf("  ID %s: %s, %d岁, 来自%s\n",
			id, person.Name, person.Age, person.City)
	}

	// 函数作为值
	operations := map[string]func(int, int) int{
		"add":      func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
		"multiply": func(a, b int) int { return a * b },
		"divide": func(a, b int) int {
			if b != 0 {
				return a / b
			}
			return 0
		},
	}

	fmt.Println("\n函数作为值:")
	a, b := 10, 3
	for name, op := range operations {
		result := op(a, b)
		fmt.Printf("  %s(%d, %d) = %d\n", name, a, b, result)
	}

	fmt.Println()
}

// 嵌套映射
func demonstrateNestedMaps() {
	fmt.Println("5. 嵌套映射:")

	// 三级嵌套：国家 -> 城市 -> 区域 -> 人口
	worldData := map[string]map[string]map[string]int{
		"中国": {
			"北京": {
				"朝阳区": 374,
				"海淀区": 313,
				"丰台区": 242,
			},
			"上海": {
				"浦东新区": 568,
				"黄浦区":  65,
				"徐汇区":  111,
			},
		},
		"美国": {
			"纽约": {
				"曼哈顿":  164,
				"布鲁克林": 259,
				"皇后区":  223,
			},
		},
	}

	fmt.Println("嵌套映射遍历:")
	for country, cities := range worldData {
		fmt.Printf("%s:\n", country)
		for city, districts := range cities {
			fmt.Printf("  %s:\n", city)
			for district, population := range districts {
				fmt.Printf("    %s: %d万人\n", district, population)
			}
		}
	}

	// 安全访问嵌套映射
	fmt.Println("\n安全访问嵌套映射:")

	// 错误的访问方式（可能panic）
	// population := worldData["中国"]["深圳"]["南山区"] // 如果深圳不存在会panic

	// 安全的访问方式
	if cities, countryExists := worldData["中国"]; countryExists {
		if districts, cityExists := cities["北京"]; cityExists {
			if population, districtExists := districts["朝阳区"]; districtExists {
				fmt.Printf("中国北京朝阳区人口: %d万人\n", population)
			}
		}
	}

	// 创建嵌套映射的安全方法
	fmt.Println("\n动态创建嵌套映射:")
	newData := make(map[string]map[string]int)

	// 确保中间层存在
	country := "日本"
	city := "东京"
	population := 1395

	if newData[country] == nil {
		newData[country] = make(map[string]int)
	}
	newData[country][city] = population

	fmt.Printf("添加数据: %v\n", newData)

	fmt.Println()
}

// 映射作为集合
func demonstrateMapAsSet() {
	fmt.Println("6. 映射作为集合:")

	// 使用map[T]bool实现集合
	fruits := map[string]bool{
		"苹果": true,
		"香蕉": true,
		"橙子": true,
	}

	fmt.Printf("水果集合: %v\n", fruits)

	// 添加元素
	fruits["葡萄"] = true
	fmt.Printf("添加葡萄后: %v\n", fruits)

	// 检查元素存在
	if fruits["苹果"] {
		fmt.Println("苹果在集合中")
	}

	if !fruits["西瓜"] {
		fmt.Println("西瓜不在集合中")
	}

	// 删除元素
	delete(fruits, "香蕉")
	fmt.Printf("删除香蕉后: %v\n", fruits)

	// 集合运算
	fmt.Println("\n集合运算:")

	set1 := map[int]bool{1: true, 2: true, 3: true, 4: true}
	set2 := map[int]bool{3: true, 4: true, 5: true, 6: true}

	fmt.Printf("集合1: %v\n", mapSetToSlice(set1))
	fmt.Printf("集合2: %v\n", mapSetToSlice(set2))

	// 交集
	intersection := make(map[int]bool)
	for elem := range set1 {
		if set2[elem] {
			intersection[elem] = true
		}
	}
	fmt.Printf("交集: %v\n", mapSetToSlice(intersection))

	// 并集
	union := make(map[int]bool)
	for elem := range set1 {
		union[elem] = true
	}
	for elem := range set2 {
		union[elem] = true
	}
	fmt.Printf("并集: %v\n", mapSetToSlice(union))

	// 差集（set1 - set2）
	difference := make(map[int]bool)
	for elem := range set1 {
		if !set2[elem] {
			difference[elem] = true
		}
	}
	fmt.Printf("差集(1-2): %v\n", mapSetToSlice(difference))

	// 使用map[T]struct{}节省内存（空结构体不占内存）
	efficientSet := map[string]struct{}{
		"a": {},
		"b": {},
		"c": {},
	}

	fmt.Printf("高效集合: %v\n", efficientSet)

	// 检查存在性
	if _, exists := efficientSet["a"]; exists {
		fmt.Println("'a'在高效集合中")
	}

	fmt.Println()
}

// 映射作为函数参数
func demonstrateMapAsParameter() {
	fmt.Println("7. 映射作为函数参数:")

	scores := map[string]int{
		"Alice":   85,
		"Bob":     92,
		"Charlie": 78,
	}

	fmt.Printf("原始映射: %v\n", scores)

	// 传递映射（引用传递）
	average := calculateAverage(scores)
	fmt.Printf("平均分: %.2f\n", average)

	// 在函数中修改映射
	addBonus(scores, 5)
	fmt.Printf("加分后: %v\n", scores)

	// 复制映射
	copied := copyMap(scores)
	fmt.Printf("复制的映射: %v\n", copied)

	// 修改复制的映射不影响原映射
	copied["Alice"] = 100
	fmt.Printf("修改复制后:\n")
	fmt.Printf("  原映射: %v\n", scores)
	fmt.Printf("  复制映射: %v\n", copied)

	// 过滤映射
	highScores := filterMap(scores, func(score int) bool {
		return score >= 90
	})
	fmt.Printf("高分学生: %v\n", highScores)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 单词计数
	fmt.Println("单词计数:")
	text := "the quick brown fox jumps over the lazy dog the fox is quick"
	wordCount := countWords(text)

	fmt.Printf("文本: %s\n", text)
	fmt.Println("单词频率:")
	for word, count := range wordCount {
		fmt.Printf("  %s: %d\n", word, count)
	}

	// 2. 缓存实现
	fmt.Println("\n缓存实现:")
	cache := make(map[string]interface{})

	// 存储数据
	cache["user:123"] = map[string]string{"name": "张三", "age": "25"}
	cache["config:timeout"] = 30
	cache["flag:enabled"] = true

	fmt.Println("缓存内容:")
	for key, value := range cache {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// 3. 配置管理
	fmt.Println("\n配置管理:")
	config := map[string]map[string]interface{}{
		"database": {
			"host":     "localhost",
			"port":     5432,
			"username": "admin",
			"password": "secret",
		},
		"server": {
			"port":    8080,
			"debug":   true,
			"timeout": 30,
		},
	}

	fmt.Println("应用配置:")
	for section, settings := range config {
		fmt.Printf("  [%s]\n", section)
		for key, value := range settings {
			fmt.Printf("    %s = %v\n", key, value)
		}
	}

	// 4. 路由表
	fmt.Println("\n路由表:")
	type Handler func(string) string

	routes := map[string]Handler{
		"/":          func(path string) string { return "欢迎页面" },
		"/about":     func(path string) string { return "关于我们" },
		"/contact":   func(path string) string { return "联系方式" },
		"/api/users": func(path string) string { return "用户API" },
	}

	testPaths := []string{"/", "/about", "/api/users", "/notfound"}

	for _, path := range testPaths {
		if handler, exists := routes[path]; exists {
			response := handler(path)
			fmt.Printf("  %s -> %s\n", path, response)
		} else {
			fmt.Printf("  %s -> 404 页面未找到\n", path)
		}
	}

	// 5. 数据索引
	fmt.Println("\n数据索引:")

	type User struct {
		ID   int
		Name string
		Age  int
		City string
	}

	users := []User{
		{1, "张三", 25, "北京"},
		{2, "李四", 30, "上海"},
		{3, "王五", 25, "北京"},
		{4, "赵六", 35, "广州"},
	}

	// 按年龄索引
	ageIndex := make(map[int][]User)
	for _, user := range users {
		ageIndex[user.Age] = append(ageIndex[user.Age], user)
	}

	// 按城市索引
	cityIndex := make(map[string][]User)
	for _, user := range users {
		cityIndex[user.City] = append(cityIndex[user.City], user)
	}

	fmt.Println("按年龄分组:")
	for age, userList := range ageIndex {
		fmt.Printf("  %d岁: ", age)
		for i, user := range userList {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(user.Name)
		}
		fmt.Println()
	}

	fmt.Println("按城市分组:")
	for city, userList := range cityIndex {
		fmt.Printf("  %s: ", city)
		for i, user := range userList {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(user.Name)
		}
		fmt.Println()
	}

	// 6. LRU缓存简单实现
	fmt.Println("\n简单LRU缓存:")
	lru := NewLRUCache(3)

	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3)
	fmt.Printf("初始: %v\n", lru.Keys())

	lru.Get("a") // 访问a，使其成为最新
	fmt.Printf("访问a后: %v\n", lru.Keys())

	lru.Put("d", 4) // 添加d，应该淘汰最旧的b
	fmt.Printf("添加d后: %v\n", lru.Keys())

	fmt.Println()
}

// 辅助函数

// 将map集合转换为切片（用于显示）
func mapSetToSlice(set map[int]bool) []int {
	var result []int
	for elem := range set {
		result = append(result, elem)
	}
	sort.Ints(result)
	return result
}

// 计算平均分
func calculateAverage(scores map[string]int) float64 {
	if len(scores) == 0 {
		return 0
	}

	total := 0
	for _, score := range scores {
		total += score
	}

	return float64(total) / float64(len(scores))
}

// 给所有分数加分
func addBonus(scores map[string]int, bonus int) {
	for name := range scores {
		scores[name] += bonus
	}
}

// 复制映射
func copyMap(original map[string]int) map[string]int {
	copied := make(map[string]int)
	for key, value := range original {
		copied[key] = value
	}
	return copied
}

// 过滤映射
func filterMap(scores map[string]int, predicate func(int) bool) map[string]int {
	result := make(map[string]int)
	for name, score := range scores {
		if predicate(score) {
			result[name] = score
		}
	}
	return result
}

// 单词计数
func countWords(text string) map[string]int {
	wordCount := make(map[string]int)

	// 简单的单词分割（实际应该使用正则表达式）
	word := ""
	for _, char := range text + " " {
		if char == ' ' {
			if word != "" {
				wordCount[word]++
				word = ""
			}
		} else {
			word += string(char)
		}
	}

	return wordCount
}

// 简单LRU缓存实现
type LRUCache struct {
	capacity int
	data     map[string]int
	order    []string
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		data:     make(map[string]int),
		order:    make([]string, 0),
	}
}

func (lru *LRUCache) Get(key string) (int, bool) {
	if value, exists := lru.data[key]; exists {
		// 移到最前面
		lru.moveToFront(key)
		return value, true
	}
	return 0, false
}

func (lru *LRUCache) Put(key string, value int) {
	if _, exists := lru.data[key]; exists {
		// 更新已存在的键
		lru.data[key] = value
		lru.moveToFront(key)
	} else {
		// 添加新键
		if len(lru.data) >= lru.capacity {
			// 移除最旧的
			oldest := lru.order[len(lru.order)-1]
			delete(lru.data, oldest)
			lru.order = lru.order[:len(lru.order)-1]
		}

		lru.data[key] = value
		lru.order = append([]string{key}, lru.order...)
	}
}

func (lru *LRUCache) moveToFront(key string) {
	// 从当前位置移除
	for i, k := range lru.order {
		if k == key {
			lru.order = append(lru.order[:i], lru.order[i+1:]...)
			break
		}
	}
	// 添加到最前面
	lru.order = append([]string{key}, lru.order...)
}

func (lru *LRUCache) Keys() []string {
	return lru.order
}

/*
=== 练习题 ===

1. 实现一个映射的反转函数（交换键和值）

2. 编写一个函数，合并多个映射

3. 实现一个线程安全的映射

4. 编写一个函数，比较两个映射是否相等

5. 实现一个支持过期时间的缓存

6. 创建一个基于映射的图数据结构

7. 实现一个简单的模板引擎（变量替换）

运行命令：
go run main.go

高级练习：
1. 实现一个一致性哈希算法
2. 编写一个分布式缓存系统
3. 实现一个基于映射的状态机
4. 创建一个数据库查询缓存
5. 实现一个简单的JSON解析器

重要概念：
- 映射是引用类型，零值为nil
- 键必须是可比较类型
- 映射不是并发安全的
- 遍历顺序是随机的
- 删除操作使用delete函数
- 可以通过逗号ok语法检查键是否存在
*/
