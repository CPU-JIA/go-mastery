// Package main demonstrates switch statement usage in Go language.
// This module covers basic switch, multiple values, fallthrough,
// type switch, and practical examples.
package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	// 星期常量
	WeekdayMon = 1
	WeekdayTue = 2
	WeekdayWed = 3
	WeekdayThu = 4
	WeekdayFri = 5
	WeekdaySat = 6
	WeekdaySun = 7

	// 月份常量
	January   = 1
	February  = 2
	March     = 3
	April     = 4
	May       = 5
	June      = 6
	July      = 7
	August    = 8
	September = 9
	October   = 10
	November  = 11
	December  = 12

	// HTTP状态码
	StatusOK                  = 200
	StatusCreated             = 201
	StatusAccepted            = 202
	StatusNoContent           = 204
	StatusMovedPermanently    = 301
	StatusFound               = 302
	StatusNotModified         = 304
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusInternalServerError = 500
	StatusNotImplemented      = 501
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503

	// 分数等级
	GradeExcellent = 90
	GradeGood      = 80
	GradeFair      = 70
	GradePass      = 60

	// 温度阈值
	TempHot  = 35
	TempWarm = 25
	TempMild = 15
	TempCool = 5

	// 年龄阈值
	AdultAge = 18

	// 数值常量
	DefaultLevel = 3
	NumberFive   = 5
	NumberSeven  = 7
	NumberTen    = 10

	// HTTP状态码类别
	StatusClass2xx = 2
	StatusClass3xx = 3
	StatusClass4xx = 4
	StatusClass5xx = 5

	// 其他常量
	FloatHalf = 0.5
)

const (
	// 状态字符串常量
	StateIdle    = "idle"
	StateRunning = "running"
	StatePaused  = "paused"
	StateStopped = "stopped"
)

// 安全随机数生成函数
func secureRandomInt(maxValue int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxValue)))
	if err != nil {
		// 安全fallback：使用时间戳
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano() % int64(maxValue)
		// 检查是否在int范围内
		if fallback > int64(^uint(0)>>1) {
			fallback %= int64(^uint(0) >> 1)
		}
		return int(fallback)
	}
	// G115安全修复：检查int64到int的安全转换
	result := n.Int64()
	if result > int64(^uint(0)>>1) {
		result %= int64(maxValue)
	}
	return int(result)
}

/*
=== Go语言第六课：Switch语句 ===

学习目标：
1. 掌握switch语句的基本语法
2. 理解Go switch的特殊性（自动break）
3. 学会使用fallthrough关键字
4. 掌握无表达式switch和类型switch
5. 了解switch在Go中的高级用法

Go switch特点：
- 自动break，不需要手动添加
- 支持多个值匹配
- 可以没有表达式（等价于switch true）
- 支持类型switch
- case可以是表达式
*/

func main() {
	fmt.Println("=== Go语言Switch语句学习 ===")

	// 1. 基本switch语句
	demonstrateBasicSwitch()

	// 2. 多值匹配
	demonstrateMultipleValues()

	// 3. fallthrough关键字
	demonstrateFallthrough()

	// 4. 无表达式switch
	demonstrateExpressionlessSwitch()

	// 5. 类型switch
	demonstrateTypeSwitch()

	// 6. switch中的短变量声明
	demonstrateShortVarDeclaration()

	// 7. 高级switch用法
	demonstrateAdvancedSwitch()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本switch语句
func demonstrateBasicSwitch() {
	fmt.Println("1. 基本switch语句:")

	// 星期几
	day := WeekdayWed
	fmt.Printf("今天是星期%d，", day)

	switch day {
	case WeekdayMon:
		fmt.Println("星期一")
	case WeekdayTue:
		fmt.Println("星期二")
	case WeekdayWed:
		fmt.Println("星期三")
	case WeekdayThu:
		fmt.Println("星期四")
	case WeekdayFri:
		fmt.Println("星期五")
	case WeekdaySat:
		fmt.Println("星期六")
	case WeekdaySun:
		fmt.Println("星期日")
	default:
		fmt.Println("无效的日期")
	}

	// 等级评定
	grade := 'B'
	fmt.Printf("成绩等级 %c，", grade)

	switch grade {
	case 'A':
		fmt.Println("优秀!")
	case 'B':
		fmt.Println("良好!")
	case 'C':
		fmt.Println("中等")
	case 'D':
		fmt.Println("及格")
	case 'F':
		fmt.Println("不及格")
	default:
		fmt.Println("无效等级")
	}

	// 数字分类
	number := 42
	fmt.Printf("数字 %d ", number)

	switch {
	case number < 0:
		fmt.Println("是负数")
	case number == 0:
		fmt.Println("是零")
	case number > 0 && number <= NumberTen:
		fmt.Println("是1-10的正数")
	case number > NumberTen:
		fmt.Println("是大于10的正数")
	}

	fmt.Println()
}

// 多值匹配
func demonstrateMultipleValues() {
	fmt.Println("2. 多值匹配:")

	// 月份季节判断
	month := August
	fmt.Printf("第%d月是", month)

	switch month {
	case December, January, February:
		fmt.Println("冬季")
	case March, April, May:
		fmt.Println("春季")
	case June, July, August:
		fmt.Println("夏季")
	case September, October, November:
		fmt.Println("秋季")
	default:
		fmt.Println("无效月份")
	}

	// 字符类型判断
	char := 'a'
	fmt.Printf("字符 '%c' 是", char)

	switch char {
	case 'a', 'e', 'i', 'o', 'u':
		fmt.Println("元音字母")
	case 'A', 'E', 'I', 'O', 'U':
		fmt.Println("大写元音字母")
	default:
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			fmt.Println("辅音字母")
		} else {
			fmt.Println("非字母字符")
		}
	}

	// HTTP状态码分类
	statusCode := StatusNotFound
	fmt.Printf("HTTP状态码 %d: ", statusCode)

	switch statusCode {
	case StatusOK, StatusCreated, StatusAccepted, StatusNoContent:
		fmt.Println("成功响应")
	case StatusMovedPermanently, StatusFound, StatusNotModified:
		fmt.Println("重定向")
	case StatusBadRequest, StatusUnauthorized, StatusForbidden, StatusNotFound:
		fmt.Println("客户端错误")
	case StatusInternalServerError, StatusNotImplemented, StatusBadGateway, StatusServiceUnavailable:
		fmt.Println("服务器错误")
	default:
		fmt.Println("其他状态码")
	}

	fmt.Println()
}

// fallthrough关键字
func demonstrateFallthrough() {
	fmt.Println("3. fallthrough关键字:")

	score := 85
	fmt.Printf("分数 %d 的评价: ", score)

	switch {
	case score >= GradeExcellent:
		fmt.Print("优秀")
		fallthrough
	case score >= GradeGood:
		fmt.Print("良好")
		fallthrough
	case score >= GradeFair:
		fmt.Print("中等")
		fallthrough
	case score >= GradePass:
		fmt.Print("及格")
		fallthrough
	default:
		fmt.Println("(评价完成)")
	}

	// 权限检查示例
	userLevel := DefaultLevel
	fmt.Printf("\n用户等级 %d 拥有的权限: ", userLevel)

	switch userLevel {
	case 4:
		fmt.Print("系统管理 ")
		fallthrough
	case DefaultLevel:
		fmt.Print("用户管理 ")
		fallthrough
	case 2:
		fmt.Print("数据写入 ")
		fallthrough
	case 1:
		fmt.Print("数据读取 ")
		fallthrough
	default:
		fmt.Println("基础访问")
	}

	// 数字处理流水线
	num := 6
	fmt.Printf("\n数字 %d 的处理流程: ", num)

	switch {
	case num > NumberFive:
		fmt.Print("大数处理→")
		fallthrough
	case num > 0:
		fmt.Print("正数处理→")
		fallthrough
	default:
		fmt.Println("基础处理")
	}

	fmt.Println()
}

// 无表达式switch
func demonstrateExpressionlessSwitch() {
	fmt.Println("4. 无表达式switch:")

	// 等价于 switch true
	temperature := TempWarm
	humidity := 60

	fmt.Printf("温度%d°C，湿度%d%%，天气状况: ", temperature, humidity)

	switch {
	case temperature > TempHot:
		fmt.Println("酷热")
	case temperature > TempWarm && humidity > 70:
		fmt.Println("闷热")
	case temperature > TempWarm:
		fmt.Println("温暖")
	case temperature > TempMild:
		fmt.Println("温和")
	case temperature > TempCool:
		fmt.Println("凉爽")
	default:
		fmt.Println("寒冷")
	}

	// 复杂条件判断
	age := 25
	income := 50000
	hasJob := true

	fmt.Printf("年龄%d，收入%d，有工作%t，信贷评级: ", age, income, hasJob)

	switch {
	case age >= 25 && income > 60000 && hasJob:
		fmt.Println("优秀")
	case age >= 21 && income > 40000 && hasJob:
		fmt.Println("良好")
	case age >= AdultAge && income > 20000:
		fmt.Println("一般")
	case age >= AdultAge:
		fmt.Println("较差")
	default:
		fmt.Println("不符合条件")
	}

	// 时间段判断
	hour := time.Now().Hour()
	fmt.Printf("当前时间%d点，", hour)

	switch {
	case hour >= 6 && hour < 12:
		fmt.Println("上午时光")
	case hour >= 12 && hour < 14:
		fmt.Println("午餐时间")
	case hour >= 14 && hour < 18:
		fmt.Println("下午时光")
	case hour >= 18 && hour < 22:
		fmt.Println("晚上时间")
	default:
		fmt.Println("深夜时分")
	}

	fmt.Println()
}

// 类型switch
func demonstrateTypeSwitch() {
	fmt.Println("5. 类型switch:")

	// 处理不同类型的接口值
	values := []interface{}{
		42,
		"hello",
		3.14,
		true,
		[]int{1, 2, 3},
		map[string]int{"key": 1},
		nil,
	}

	for i, value := range values {
		fmt.Printf("值%d: ", i+1)

		switch v := value.(type) {
		case nil:
			fmt.Println("nil值")
		case bool:
			if v {
				fmt.Println("布尔值: true")
			} else {
				fmt.Println("布尔值: false")
			}
		case int:
			fmt.Printf("整数: %d\n", v)
		case float64:
			fmt.Printf("浮点数: %.2f\n", v)
		case string:
			fmt.Printf("字符串: \"%s\" (长度: %d)\n", v, len(v))
		case []int:
			fmt.Printf("整数切片: %v (长度: %d)\n", v, len(v))
		case map[string]int:
			fmt.Printf("字符串到整数的映射: %v\n", v)
		default:
			fmt.Printf("未知类型: %T\n", v)
		}
	}

	// 类型断言与类型switch结合
	var data interface{} = "Go语言"

	switch value := data.(type) {
	case string:
		fmt.Printf("字符串数据: %s，转为大写: %s\n",
			value, value)
	case int:
		fmt.Printf("整数数据: %d，平方: %d\n", value, value*value)
	case float64:
		fmt.Printf("浮点数据: %.2f，开方: %.2f\n", value, value*0.5)
	}

	fmt.Println()
}

// switch中的短变量声明
func demonstrateShortVarDeclaration() {
	fmt.Println("6. switch中的短变量声明:")

	// 随机数生成和处理
	// 注意：crypto/rand不需要设置种子

	switch num := secureRandomInt(10) + 1; {
	case num <= 3:
		fmt.Printf("小数字: %d\n", num)
	case num <= 7:
		fmt.Printf("中等数字: %d\n", num)
	default:
		fmt.Printf("大数字: %d\n", num)
	}

	// 字符串处理
	switch length := len("Go Programming"); {
	case length < 5:
		fmt.Printf("短字符串，长度: %d\n", length)
	case length < 10:
		fmt.Printf("中等字符串，长度: %d\n", length)
	default:
		fmt.Printf("长字符串，长度: %d\n", length)
	}

	// 函数调用结果处理
	switch result := divide(10, 3); {
	case result > 5:
		fmt.Printf("除法结果较大: %.2f\n", result)
	case result > 2:
		fmt.Printf("除法结果中等: %.2f\n", result)
	default:
		fmt.Printf("除法结果较小: %.2f\n", result)
	}

	fmt.Println()
}

// 高级switch用法
func demonstrateAdvancedSwitch() {
	fmt.Println("7. 高级switch用法:")

	// switch表达式中的函数调用
	switch getRandomDay() {
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		fmt.Println("工作日")
	case "Saturday", "Sunday":
		fmt.Println("周末")
	}

	// 嵌套switch
	userType := "admin"
	userLevel := 3

	fmt.Printf("用户类型: %s，等级: %d，权限: ", userType, userLevel)

	switch userType {
	case "admin":
		switch {
		case userLevel >= 5:
			fmt.Println("超级管理员")
		case userLevel >= 3:
			fmt.Println("高级管理员")
		default:
			fmt.Println("普通管理员")
		}
	case "user":
		switch {
		case userLevel >= 3:
			fmt.Println("VIP用户")
		default:
			fmt.Println("普通用户")
		}
	default:
		fmt.Println("访客")
	}

	// switch中的复杂表达式
	x, y := 5, 3

	switch {
	case x+y > 10:
		fmt.Printf("%d + %d = %d，和大于10\n", x, y, x+y)
	case x*y > 10:
		fmt.Printf("%d × %d = %d，积大于10\n", x, y, x*y)
	case x-y > 0:
		fmt.Printf("%d - %d = %d，差大于0\n", x, y, x-y)
	default:
		fmt.Printf("%d 和 %d 没有特殊关系\n", x, y)
	}

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 简单的状态机
	demonstrateStateMachine()

	// 2. 文件扩展名处理
	demonstrateFileTypeDetection()

	// 3. 错误码处理
	demonstrateErrorCodeHandling()

	// 4. 计算器操作
	demonstrateCalculatorOperations()

	fmt.Println()
}

// 状态机示例
func demonstrateStateMachine() {
	fmt.Println("简单状态机示例:")
	state := StateIdle

	for i := 0; i < NumberFive; i++ {
		fmt.Printf("  状态 %d: %s → ", i+1, state)

		switch state {
		case StateIdle:
			state = StateRunning
			fmt.Println("开始运行")
		case StateRunning:
			if i%2 == 0 {
				state = StatePaused
				fmt.Println("暂停")
			} else {
				state = StateStopped
				fmt.Println("停止")
			}
		case StatePaused:
			state = StateRunning
			fmt.Println("恢复运行")
		case StateStopped:
			state = StateIdle
			fmt.Println("重置为空闲")
		}
	}
}

// 文件类型检测示例
func demonstrateFileTypeDetection() {
	files := []string{"document.pdf", "image.jpg", "data.csv", "script.go", "unknown.xyz"}

	fmt.Println("\n文件类型识别:")
	for _, filename := range files {
		fmt.Printf("  %s: ", filename)

		// 获取扩展名
		ext := getFileExtension(filename)

		switch ext {
		case ".pdf":
			fmt.Println("PDF文档")
		case ".jpg", ".jpeg", ".png", ".gif":
			fmt.Println("图片文件")
		case ".csv", ".xlsx":
			fmt.Println("数据文件")
		case ".go", ".py", ".js", ".java":
			fmt.Println("程序代码")
		case ".txt", ".md":
			fmt.Println("文本文件")
		default:
			fmt.Println("未知类型")
		}
	}
}

// 获取文件扩展名
func getFileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}

// 错误码处理示例
func demonstrateErrorCodeHandling() {
	fmt.Println("\n错误码处理:")
	errorCodes := []int{StatusOK, StatusNotFound, StatusInternalServerError, StatusForbidden, StatusCreated}

	for _, code := range errorCodes {
		fmt.Printf("  状态码 %d: ", code)

		switch code / 100 { // 使用除法获取状态码类别
		case StatusClass2xx:
			handleSuccessStatusCode(code)
		case StatusClass3xx:
			fmt.Println("重定向")
		case StatusClass4xx:
			handleClientErrorStatusCode(code)
		case StatusClass5xx:
			fmt.Println("服务器错误")
		default:
			fmt.Println("未知状态码")
		}
	}
}

// 处理成功状态码
func handleSuccessStatusCode(code int) {
	switch code {
	case StatusOK:
		fmt.Println("请求成功")
	case StatusCreated:
		fmt.Println("创建成功")
	case StatusNoContent:
		fmt.Println("无内容")
	default:
		fmt.Println("成功响应")
	}
}

// 处理客户端错误状态码
func handleClientErrorStatusCode(code int) {
	switch code {
	case StatusBadRequest:
		fmt.Println("请求错误")
	case StatusUnauthorized:
		fmt.Println("未授权")
	case StatusForbidden:
		fmt.Println("禁止访问")
	case StatusNotFound:
		fmt.Println("未找到")
	default:
		fmt.Println("客户端错误")
	}
}

// 计算器操作示例
func demonstrateCalculatorOperations() {
	fmt.Println("\n简单计算器:")
	operations := []struct {
		a, b float64
		op   string
	}{
		{10, 5, "+"},
		{10, 5, "-"},
		{10, 5, "*"},
		{10, 5, "/"},
		{10, 0, "/"},
		{10, 5, "%"},
	}

	for _, calc := range operations {
		fmt.Printf("  %.1f %s %.1f = ", calc.a, calc.op, calc.b)

		switch calc.op {
		case "+":
			fmt.Printf("%.1f\n", calc.a+calc.b)
		case "-":
			fmt.Printf("%.1f\n", calc.a-calc.b)
		case "*":
			fmt.Printf("%.1f\n", calc.a*calc.b)
		case "/":
			if calc.b != 0 {
				fmt.Printf("%.2f\n", calc.a/calc.b)
			} else {
				fmt.Println("错误：除数不能为零")
			}
		default:
			fmt.Println("不支持的操作")
		}
	}
}

// 辅助函数
func getRandomDay() string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	return days[secureRandomInt(len(days))]
}

func divide(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

/*
=== 练习题 ===

1. 编写一个月份天数计算器，考虑闰年情况

2. 实现一个简单的自动贩卖机程序：
   - 不同商品有不同价格
   - 根据投入金额返回商品或找零

3. 创建一个学生成绩管理系统：
   - 根据分数给出等级
   - 根据等级给出奖学金等级

4. 编写一个颜色分类器：
   - 输入RGB值
   - 输出主要颜色名称

5. 实现一个简单的命令行解析器

6. 创建一个工作日程安排器：
   - 根据时间推荐活动
   - 处理特殊日期

运行命令：
go run main.go

高级练习：
1. 实现一个状态机框架
2. 编写一个表达式求值器
3. 创建一个路由分发器
4. 实现一个配置文件解析器
5. 编写一个简单的编译器前端

性能提示：
- switch比多个if-else效率更高
- 将最常见的case放在前面
- 使用类型switch处理interface{}
- 避免在case中进行复杂计算
*/
