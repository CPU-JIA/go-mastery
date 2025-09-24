package main

import (
	"fmt"
	"math/rand"
	"time"
)

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
	day := 3
	fmt.Printf("今天是星期%d，", day)

	switch day {
	case 1:
		fmt.Println("星期一")
	case 2:
		fmt.Println("星期二")
	case 3:
		fmt.Println("星期三")
	case 4:
		fmt.Println("星期四")
	case 5:
		fmt.Println("星期五")
	case 6:
		fmt.Println("星期六")
	case 7:
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
	case number > 0 && number <= 10:
		fmt.Println("是1-10的正数")
	case number > 10:
		fmt.Println("是大于10的正数")
	}

	fmt.Println()
}

// 多值匹配
func demonstrateMultipleValues() {
	fmt.Println("2. 多值匹配:")

	// 月份季节判断
	month := 8
	fmt.Printf("第%d月是", month)

	switch month {
	case 12, 1, 2:
		fmt.Println("冬季")
	case 3, 4, 5:
		fmt.Println("春季")
	case 6, 7, 8:
		fmt.Println("夏季")
	case 9, 10, 11:
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
	statusCode := 404
	fmt.Printf("HTTP状态码 %d: ", statusCode)

	switch statusCode {
	case 200, 201, 202, 204:
		fmt.Println("成功响应")
	case 301, 302, 304:
		fmt.Println("重定向")
	case 400, 401, 403, 404:
		fmt.Println("客户端错误")
	case 500, 501, 502, 503:
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
	case score >= 90:
		fmt.Print("优秀")
		fallthrough
	case score >= 80:
		fmt.Print("良好")
		fallthrough
	case score >= 70:
		fmt.Print("中等")
		fallthrough
	case score >= 60:
		fmt.Print("及格")
		fallthrough
	default:
		fmt.Println("(评价完成)")
	}

	// 权限检查示例
	userLevel := 3
	fmt.Printf("\n用户等级 %d 拥有的权限: ", userLevel)

	switch userLevel {
	case 4:
		fmt.Print("系统管理 ")
		fallthrough
	case 3:
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
	case num > 5:
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
	temperature := 25
	humidity := 60

	fmt.Printf("温度%d°C，湿度%d%%，天气状况: ", temperature, humidity)

	switch {
	case temperature > 35:
		fmt.Println("酷热")
	case temperature > 25 && humidity > 70:
		fmt.Println("闷热")
	case temperature > 25:
		fmt.Println("温暖")
	case temperature > 15:
		fmt.Println("温和")
	case temperature > 5:
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
	case age >= 18 && income > 20000:
		fmt.Println("一般")
	case age >= 18:
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
			value, fmt.Sprintf("%s", value))
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
	rand.Seed(time.Now().UnixNano())

	switch num := rand.Intn(10) + 1; {
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
	fmt.Println("简单状态机示例:")
	state := "idle"

	for i := 0; i < 5; i++ {
		fmt.Printf("  状态 %d: %s → ", i+1, state)

		switch state {
		case "idle":
			state = "running"
			fmt.Println("开始运行")
		case "running":
			if i%2 == 0 {
				state = "paused"
				fmt.Println("暂停")
			} else {
				state = "stopped"
				fmt.Println("停止")
			}
		case "paused":
			state = "running"
			fmt.Println("恢复运行")
		case "stopped":
			state = "idle"
			fmt.Println("重置为空闲")
		}
	}

	// 2. 文件扩展名处理
	files := []string{"document.pdf", "image.jpg", "data.csv", "script.go", "unknown.xyz"}

	fmt.Println("\n文件类型识别:")
	for _, filename := range files {
		fmt.Printf("  %s: ", filename)

		// 获取扩展名
		ext := ""
		for i := len(filename) - 1; i >= 0; i-- {
			if filename[i] == '.' {
				ext = filename[i:]
				break
			}
		}

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

	// 3. 错误码处理
	fmt.Println("\n错误码处理:")
	errorCodes := []int{200, 404, 500, 403, 201}

	for _, code := range errorCodes {
		fmt.Printf("  状态码 %d: ", code)

		switch code / 100 { // 使用除法获取状态码类别
		case 2:
			switch code {
			case 200:
				fmt.Println("请求成功")
			case 201:
				fmt.Println("创建成功")
			case 204:
				fmt.Println("无内容")
			default:
				fmt.Println("成功响应")
			}
		case 3:
			fmt.Println("重定向")
		case 4:
			switch code {
			case 400:
				fmt.Println("请求错误")
			case 401:
				fmt.Println("未授权")
			case 403:
				fmt.Println("禁止访问")
			case 404:
				fmt.Println("未找到")
			default:
				fmt.Println("客户端错误")
			}
		case 5:
			fmt.Println("服务器错误")
		default:
			fmt.Println("未知状态码")
		}
	}

	// 4. 计算器操作
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

	fmt.Println()
}

// 辅助函数
func getRandomDay() string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	return days[rand.Intn(len(days))]
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
