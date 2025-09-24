package main

import "fmt"

/*
=== Go语言第四课：条件语句(if/else) ===

学习目标：
1. 掌握if/else语句的语法
2. 理解Go中条件语句的特点
3. 学会使用短变量声明在条件语句中
4. 掌握条件语句的嵌套和复合条件

Go条件语句特点：
- 条件表达式不需要括号
- 大括号必须有，且左大括号必须在同一行
- 支持在条件语句中进行短变量声明
- 条件必须是布尔类型
*/

func main() {
	fmt.Println("=== Go语言条件语句学习 ===")

	// 1. 基本if语句
	demonstrateBasicIf()

	// 2. if-else语句
	demonstrateIfElse()

	// 3. if-else if-else链
	demonstrateIfElseChain()

	// 4. 条件语句中的短变量声明
	demonstrateShortVarDeclaration()

	// 5. 复合条件和逻辑运算符
	demonstrateCompoundConditions()

	// 6. 嵌套条件语句
	demonstrateNestedConditions()

	// 7. 实际应用示例
	demonstratePracticalExamples()
}

// 基本if语句
func demonstrateBasicIf() {
	fmt.Println("1. 基本if语句:")

	age := 18
	score := 85
	isStudent := true

	// 基本if语句
	if age >= 18 {
		fmt.Println("✅ 已成年")
	}

	if score >= 60 {
		fmt.Println("✅ 考试及格")
	}

	if isStudent {
		fmt.Println("✅ 是学生")
	}

	// 注意：Go不支持三元运算符，必须使用if-else
	var result string
	if score >= 90 {
		result = "优秀"
	} else {
		result = "良好"
	}
	fmt.Printf("成绩评定: %s\n", result)

	fmt.Println()
}

// if-else语句
func demonstrateIfElse() {
	fmt.Println("2. if-else语句:")

	temperature := 25

	if temperature > 30 {
		fmt.Println("🌡️ 天气很热")
	} else {
		fmt.Println("🌡️ 天气适宜")
	}

	// 数值判断
	number := -5
	if number > 0 {
		fmt.Printf("%d 是正数\n", number)
	} else if number < 0 {
		fmt.Printf("%d 是负数\n", number)
	} else {
		fmt.Printf("%d 是零\n", number)
	}

	// 字符串判断
	username := "admin"
	if username == "admin" {
		fmt.Println("👨‍💼 管理员登录")
	} else {
		fmt.Println("👤 普通用户登录")
	}

	fmt.Println()
}

// if-else if-else链
func demonstrateIfElseChain() {
	fmt.Println("3. if-else if-else链:")

	// 成绩分级
	score := 78
	var grade string

	if score >= 90 {
		grade = "A"
	} else if score >= 80 {
		grade = "B"
	} else if score >= 70 {
		grade = "C"
	} else if score >= 60 {
		grade = "D"
	} else {
		grade = "F"
	}

	fmt.Printf("分数: %d, 等级: %s\n", score, grade)

	// 时间段判断
	hour := 14
	var timeOfDay string

	if hour >= 5 && hour < 12 {
		timeOfDay = "上午"
	} else if hour >= 12 && hour < 14 {
		timeOfDay = "中午"
	} else if hour >= 14 && hour < 18 {
		timeOfDay = "下午"
	} else if hour >= 18 && hour < 22 {
		timeOfDay = "晚上"
	} else {
		timeOfDay = "深夜"
	}

	fmt.Printf("时间: %d:00, 时段: %s\n", hour, timeOfDay)

	fmt.Println()
}

// 条件语句中的短变量声明
func demonstrateShortVarDeclaration() {
	fmt.Println("4. 短变量声明在条件语句中:")

	// 在if语句中声明变量
	if length := len("Hello, Go!"); length > 5 {
		fmt.Printf("字符串长度 %d 大于5\n", length)
	}
	// 注意：length变量只在if块内有效

	// 实际应用：错误处理模式
	if result, err := divideNumbers(10, 2); err != nil {
		fmt.Printf("❌ 计算错误: %v\n", err)
	} else {
		fmt.Printf("✅ 计算结果: %.2f\n", result)
	}

	// 模拟map查找
	userRoles := map[string]string{
		"alice":   "admin",
		"bob":     "user",
		"charlie": "guest",
	}

	if role, exists := userRoles["alice"]; exists {
		fmt.Printf("用户alice的角色: %s\n", role)
	} else {
		fmt.Println("用户不存在")
	}

	// 类型断言
	var value interface{} = "Hello"
	if str, ok := value.(string); ok {
		fmt.Printf("值是字符串: %s\n", str)
	} else {
		fmt.Println("值不是字符串")
	}

	fmt.Println()
}

// 复合条件和逻辑运算符
func demonstrateCompoundConditions() {
	fmt.Println("5. 复合条件和逻辑运算符:")

	age := 25
	hasLicense := true
	hasExperience := false
	salary := 50000

	// 逻辑与 (&&)
	if age >= 18 && hasLicense {
		fmt.Println("✅ 可以开车")
	}

	// 逻辑或 (||)
	if hasLicense || hasExperience {
		fmt.Println("✅ 符合驾驶条件之一")
	}

	// 逻辑非 (!)
	if !hasExperience {
		fmt.Println("⚠️ 缺乏经验")
	}

	// 复杂条件组合
	if (age >= 25 && salary > 40000) || (age >= 30 && salary > 30000) {
		fmt.Println("✅ 符合贷款条件")
	}

	// 范围检查
	score := 85
	if score >= 80 && score <= 90 {
		fmt.Println("✅ 分数在80-90区间")
	}

	// 多重条件
	username := "admin"
	password := "123456"
	isActive := true

	if username == "admin" && password == "123456" && isActive {
		fmt.Println("🎉 登录成功")
	} else {
		fmt.Println("❌ 登录失败")
	}

	fmt.Println()
}

// 嵌套条件语句
func demonstrateNestedConditions() {
	fmt.Println("6. 嵌套条件语句:")

	weather := "sunny"
	temperature := 25
	hasUmbrella := false

	if weather == "sunny" {
		fmt.Println("☀️ 今天晴天")
		if temperature > 30 {
			fmt.Println("   🌡️ 天气很热，记得防晒")
		} else if temperature > 20 {
			fmt.Println("   🌡️ 天气温和，适合出行")
		} else {
			fmt.Println("   🌡️ 天气较凉，多穿衣服")
		}
	} else if weather == "rainy" {
		fmt.Println("🌧️ 今天下雨")
		if hasUmbrella {
			fmt.Println("   ☂️ 有雨伞，可以出门")
		} else {
			fmt.Println("   ⚠️ 没有雨伞，建议待在室内")
		}
	} else {
		fmt.Println("🌫️ 天气状况未知")
	}

	// 用户权限检查
	userType := "admin"
	userLevel := 3

	if userType == "admin" {
		fmt.Println("👨‍💼 管理员用户")
		if userLevel >= 5 {
			fmt.Println("   🔓 超级管理员权限")
		} else if userLevel >= 3 {
			fmt.Println("   🔐 高级管理员权限")
		} else {
			fmt.Println("   🔒 基础管理员权限")
		}
	} else {
		fmt.Println("👤 普通用户")
	}

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("7. 实际应用示例:")

	// 银行账户操作
	balance := 1000.0
	withdrawAmount := 500.0

	fmt.Printf("账户余额: %.2f, 取款金额: %.2f\n", balance, withdrawAmount)

	if withdrawAmount <= 0 {
		fmt.Println("❌ 取款金额必须大于0")
	} else if withdrawAmount > balance {
		fmt.Println("❌ 余额不足")
	} else {
		balance -= withdrawAmount
		fmt.Printf("✅ 取款成功，余额: %.2f\n", balance)
	}

	// HTTP状态码处理
	statusCode := 200

	if statusCode >= 200 && statusCode < 300 {
		fmt.Println("✅ HTTP请求成功")
	} else if statusCode >= 400 && statusCode < 500 {
		fmt.Println("❌ 客户端错误")
	} else if statusCode >= 500 {
		fmt.Println("💥 服务器错误")
	} else {
		fmt.Println("ℹ️ 其他状态")
	}

	// 年龄分组
	age := 25

	if age < 13 {
		fmt.Println("👶 儿童")
	} else if age < 20 {
		fmt.Println("🧒 青少年")
	} else if age < 60 {
		fmt.Println("👨 成年人")
	} else {
		fmt.Println("👴 老年人")
	}

	// 文件扩展名检查
	filename := "document.pdf"

	if len(filename) > 4 {
		extension := filename[len(filename)-4:]
		if extension == ".txt" {
			fmt.Println("📄 文本文件")
		} else if extension == ".pdf" {
			fmt.Println("📑 PDF文件")
		} else if extension == ".jpg" || extension == ".png" {
			fmt.Println("🖼️ 图片文件")
		} else {
			fmt.Println("📁 未知文件类型")
		}
	} else {
		fmt.Println("❌ 文件名太短")
	}

	fmt.Println()
}

// 辅助函数：除法运算
func divideNumbers(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为零")
	}
	return a / b, nil
}

/*
=== 练习题 ===

1. 编写一个函数判断一个年份是否为闰年
   规则：能被4整除但不能被100整除，或者能被400整除

2. 创建一个简单的计算器，根据操作符进行不同的计算

3. 实现一个密码强度检查器：
   - 至少8位
   - 包含大小写字母
   - 包含数字
   - 包含特殊字符

4. 编写一个成绩管理系统：
   - 输入分数返回等级
   - 判断是否及格
   - 给出改进建议

5. 实现一个简单的用户认证系统

运行命令：
go run main.go

高级练习：
1. 实现一个复杂的条件路由系统
2. 创建一个多条件排序算法
3. 编写一个配置验证器
4. 实现一个状态机

注意事项：
- 避免过深的嵌套，考虑重构为多个函数
- 使用明确的变量名使条件更易读
- 考虑使用switch语句替代复杂的if-else链
- 注意短路求值的特性：&& 和 || 的求值顺序
*/
