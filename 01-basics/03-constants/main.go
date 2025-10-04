package main

import "fmt"

/*
=== Go语言第三堂课：常量和iota ===

学习目标：
1. 理解常量的概念和用途
2. 掌握const关键字的使用
3. 学会使用iota生成枚举
4. 了解常量的类型推导和无类型常量

常量特点：
- 编译期确定值，运行时不可修改
- 只能是基本类型：布尔、数字、字符串
- 支持类型推导
- 可以是无类型的
*/

func main() {
	fmt.Println("=== Go语言常量学习 ===")

	// 1. 基本常量声明
	demonstrateBasicConstants()

	// 2. 常量组和批量声明
	demonstrateConstantGroups()

	// 3. iota枚举生成器
	demonstrateIota()

	// 4. 高级iota用法
	demonstrateAdvancedIota()

	// 5. 无类型常量
	demonstrateUntypedConstants()

	// 6. 常量的实际应用
	demonstratePracticalUsage()
}

// 基本常量声明
func demonstrateBasicConstants() {
	fmt.Println("1. 基本常量声明:")

	// 基本常量
	const message string = "这是一个字符串常量"
	const number int = 100
	const pi float64 = 3.14159
	const isValid bool = true

	fmt.Printf("字符串常量: %s\n", message)
	fmt.Printf("整数常量: %d\n", number)
	fmt.Printf("浮点常量: %.5f\n", pi)
	fmt.Printf("布尔常量: %t\n", isValid)

	// 类型推导
	const autoString = "自动推导类型"
	const autoNumber = 42
	const autoFloat = 2.71828

	fmt.Printf("自动推导: %s, %d, %.5f\n", autoString, autoNumber, autoFloat)

	fmt.Println()
}

// 常量组和批量声明
func demonstrateConstantGroups() {
	fmt.Println("2. 常量组声明:")

	// 常量组
	const (
		appName    = "Go学习系统"
		version    = "1.0.0"
		author     = "JIA总"
		maxRetries = 3
		timeout    = 30.0
	)

	fmt.Printf("应用名称: %s\n", appName)
	fmt.Printf("版本: %s\n", version)
	fmt.Printf("作者: %s\n", author)
	fmt.Printf("最大重试: %d\n", maxRetries)
	fmt.Printf("超时时间: %.1f秒\n", timeout)

	fmt.Println()
}

// iota枚举生成器
func demonstrateIota() {
	fmt.Println("3. iota枚举生成器:")

	// 基础iota用法
	const (
		Sunday    = iota // 0
		Monday           // 1
		Tuesday          // 2
		Wednesday        // 3
		Thursday         // 4
		Friday           // 5
		Saturday         // 6
	)

	fmt.Printf("星期枚举: 周日=%d, 周一=%d, 周五=%d\n", Sunday, Monday, Friday)

	// 跳过值的iota
	const (
		_        = iota             // 跳过0
		KB int64 = 1 << (10 * iota) // 1024
		MB                          // 1048576
		GB                          // 1073741824
		TB                          // 1099511627776
	)

	fmt.Printf("存储单位: KB=%d, MB=%d, GB=%d, TB=%d\n", KB, MB, GB, TB)

	// HTTP状态码
	const (
		StatusOK                  = 200
		StatusNotFound            = 404
		StatusInternalServerError = 500
	)

	// 使用iota生成位标志
	const (
		ReadPermission    = 1 << iota // 1
		WritePermission               // 2
		ExecutePermission             // 4
	)

	fmt.Printf("权限标志: 读=%d, 写=%d, 执行=%d\n",
		ReadPermission, WritePermission, ExecutePermission)

	fmt.Println()
}

// 高级iota用法
func demonstrateAdvancedIota() {
	fmt.Println("4. 高级iota应用:")

	// 复杂表达式中的iota
	const (
		_ = iota     // 0
		a = iota * 2 // 2
		b = iota * 3 // 6
		c = iota * 4 // 12
	)
	fmt.Printf("复杂表达式: a=%d, b=%d, c=%d\n", a, b, c)

	// 在同一行使用多个iota
	const (
		x, y = iota + 1, iota + 2 // x=1, y=2
		m, n                      // m=2, n=3
		p, q                      // p=3, q=4
	)
	fmt.Printf("多重iota: x=%d,y=%d; m=%d,n=%d; p=%d,q=%d\n", x, y, m, n, p, q)

	// 中断和重置iota
	const (
		Apple  = iota // 0
		Banana        // 1
		Orange = 100  // 100 (中断iota)
		Grape  = iota // 3 (继续iota)
		Mango         // 4
	)
	fmt.Printf("中断iota: Apple=%d, Banana=%d, Orange=%d, Grape=%d, Mango=%d\n",
		Apple, Banana, Orange, Grape, Mango)

	fmt.Println()
}

// 无类型常量
func demonstrateUntypedConstants() {
	fmt.Println("5. 无类型常量:")

	// 无类型常量具有更高的精度
	const (
		BigNumber  = 1e100    // 无类型浮点常量
		SmallFloat = 1e-100   // 无类型浮点常量
		HugeInt    = 1 << 100 // 无类型整数常量
	)

	// 无类型常量可以赋值给兼容的类型
	var f32 float32 = SmallFloat
	var f64 float64 = SmallFloat
	var int64Val int64 = 1000000

	fmt.Printf("无类型常量转换: float32=%.2e, float64=%.2e\n", f32, f64)
	fmt.Printf("整数常量: %d\n", int64Val)

	// 无类型常量参与运算
	const Pi = 3.14159265358979323846
	radius := 5.0
	area := Pi * radius * radius
	fmt.Printf("圆面积计算: π=%.15f, 半径=%.1f, 面积=%.6f\n", Pi, radius, area)

	fmt.Println()
}

// 常量的实际应用
func demonstratePracticalUsage() {
	fmt.Println("6. 常量的实际应用:")

	// 配置常量
	const (
		DatabaseURL       = "localhost:5432"
		MaxConnections    = 100
		ConnectionTimeout = 30
		RetryAttempts     = 3
	)

	// 错误消息常量
	const (
		ErrInvalidInput   = "输入数据无效"
		ErrConnectionFail = "连接失败"
		ErrTimeout        = "请求超时"
	)

	// 数学常量
	const (
		E   = 2.71828182845904523536
		Pi  = 3.14159265358979323846
		Phi = 1.61803398874989484820 // 黄金比例
	)

	// 日志级别
	const (
		LogLevelDebug = iota
		LogLevelInfo
		LogLevelWarn
		LogLevelError
		LogLevelFatal
	)

	fmt.Printf("数据库配置: %s, 最大连接: %d\n", DatabaseURL, MaxConnections)
	fmt.Printf("数学常量: e=%.15f, π=%.15f, φ=%.15f\n", E, Pi, Phi)
	fmt.Printf("日志级别: Debug=%d, Info=%d, Error=%d\n",
		LogLevelDebug, LogLevelInfo, LogLevelError)

	// 使用常量进行计算
	circumference := 2 * Pi * 10
	fmt.Printf("半径10的圆周长: %.6f\n", circumference)

	fmt.Println()
}

// 类型化常量示例
type ByteSize int64

const (
	_               = iota
	KBSize ByteSize = 1 << (10 * iota)
	MBSize
	GBSize
	TBSize
)

func (b ByteSize) String() string {
	switch {
	case b >= TBSize:
		return fmt.Sprintf("%.2fTB", float64(b)/float64(TBSize))
	case b >= GBSize:
		return fmt.Sprintf("%.2fGB", float64(b)/float64(GBSize))
	case b >= MBSize:
		return fmt.Sprintf("%.2fMB", float64(b)/float64(MBSize))
	case b >= KBSize:
		return fmt.Sprintf("%.2fKB", float64(b)/float64(KBSize))
	default:
		return fmt.Sprintf("%dB", int64(b))
	}
}

func init() {
	fmt.Printf("自定义类型常量: 1GB = %s, 5GB = %s\n",
		GBSize, ByteSize(5*GBSize))
}

/*
=== 练习题 ===

1. 创建一个表示HTTP方法的常量组 (GET, POST, PUT, DELETE等)
2. 使用iota创建一个表示文件权限的位标志系统
3. 定义数学常量并计算圆、球的面积和体积
4. 创建一个表示颜色的枚举，包含RGB值
5. 实现一个自定义类型的常量，带有String()方法

运行命令：
go run main.go

高级练习：
1. 研究iota在不同const块中的行为
2. 探索无类型常量的精度限制
3. 实现一个时间单位转换系统
4. 创建一个配置系统，使用常量定义默认值

注意事项：
- 常量不能使用函数调用的结果
- 常量不能是slice、map、channel等复合类型
- iota只能在const声明中使用
- 常量的零值是其类型的零值
*/
