package main

import "fmt"

/*
=== Go语言第二课：变量和基本类型 ===

学习目标：
1. 掌握Go的基本数据类型
2. 学会变量声明的多种方式
3. 理解零值概念
4. 掌握类型转换

Go基本类型：
- 布尔型: bool
- 字符串: string
- 整型: int, int8, int16, int32, int64
- 无符号整型: uint, uint8, uint16, uint32, uint64, uintptr
- 浮点型: float32, float64
- 复数: complex64, complex128
- 字节: byte (等价于uint8)
- Unicode: rune (等价于int32)
*/

func main() {
	fmt.Println("=== Go语言变量和类型学习 ===")

	// 1. 变量声明方式
	demonstrateVariableDeclaration()

	// 2. 基本类型展示
	demonstrateBasicTypes()

	// 3. 零值演示
	demonstrateZeroValues()

	// 4. 类型转换
	demonstrateTypeConversion()

	// 5. 变量作用域
	demonstrateScope()
}

// 变量声明的多种方式
func demonstrateVariableDeclaration() {
	fmt.Println("1. 变量声明方式:")

	// 方式1: var 变量名 类型 = 值
	var name string = "Go语言"
	fmt.Printf("方式1 - var name string = \"Go语言\": %s\n", name)

	// 方式2: var 变量名 = 值 (类型推导)
	var version = 1.21
	fmt.Printf("方式2 - var version = 1.21: %.2f\n", version)

	// 方式3: 变量名 := 值 (短变量声明)
	language := "Go"
	fmt.Printf("方式3 - language := \"Go\": %s\n", language)

	// 方式4: 批量声明
	var (
		isActive bool    = true
		count    int     = 100
		ratio    float64 = 3.14
	)
	fmt.Printf("方式4 - 批量声明: %t, %d, %.2f\n", isActive, count, ratio)

	// 方式5: 多重赋值
	x, y, z := 1, 2, 3
	fmt.Printf("方式5 - 多重赋值: x=%d, y=%d, z=%d\n", x, y, z)

	fmt.Println()
}

// 基本类型演示
func demonstrateBasicTypes() {
	fmt.Println("2. Go基本类型:")

	// 布尔型
	var isLearning bool = true
	fmt.Printf("布尔型 bool: %t\n", isLearning)

	// 字符串
	var greeting string = "你好，世界！"
	fmt.Printf("字符串 string: %s\n", greeting)

	// 整型
	var age int = 25
	var smallNum int8 = 127
	var bigNum int64 = 9223372036854775807
	fmt.Printf("整型 int: %d, int8: %d, int64: %d\n", age, smallNum, bigNum)

	// 无符号整型
	var population uint = 1000000
	var byteValue uint8 = 255
	fmt.Printf("无符号整型 uint: %d, uint8: %d\n", population, byteValue)

	// 浮点型
	var height float32 = 175.5
	var weight float64 = 70.123456789
	fmt.Printf("浮点型 float32: %.2f, float64: %.6f\n", height, weight)

	// 字节和符文
	var letter byte = 'A'
	var chinese rune = '中'
	fmt.Printf("字节 byte: %c (%d), 符文 rune: %c (%d)\n", letter, letter, chinese, chinese)

	// 复数
	var complex1 complex64 = 1 + 2i
	var complex2 complex128 = 3 + 4i
	fmt.Printf("复数 complex64: %v, complex128: %v\n", complex1, complex2)

	fmt.Println()
}

// 零值演示
func demonstrateZeroValues() {
	fmt.Println("3. Go类型的零值:")

	// Go中所有类型都有零值，声明但未初始化的变量会被设为零值
	var boolZero bool
	var intZero int
	var floatZero float64
	var stringZero string
	var sliceZero []int
	var mapZero map[string]int
	var funcZero func()
	var interfaceZero interface{}
	var pointerZero *int

	fmt.Printf("bool零值: %t\n", boolZero)
	fmt.Printf("int零值: %d\n", intZero)
	fmt.Printf("float64零值: %f\n", floatZero)
	fmt.Printf("string零值: \"%s\" (空字符串)\n", stringZero)
	fmt.Printf("slice零值: %v (nil)\n", sliceZero)
	fmt.Printf("map零值: %v (nil)\n", mapZero)
	fmt.Printf("func零值: %T (nil)\n", funcZero)
	fmt.Printf("interface{}零值: %v (nil)\n", interfaceZero)
	fmt.Printf("*int零值: %v (nil)\n", pointerZero)

	fmt.Println()
}

// 类型转换演示
func demonstrateTypeConversion() {
	fmt.Println("4. 类型转换:")

	// Go需要显式类型转换
	var intValue int = 42
	var floatValue float64 = 3.14
	var stringValue string = "123"

	// 数值类型转换
	convertedFloat := float64(intValue)
	convertedInt := int(floatValue)
	fmt.Printf("int转float64: %d -> %.2f\n", intValue, convertedFloat)
	fmt.Printf("float64转int: %.2f -> %d (截断)\n", floatValue, convertedInt)

	// 字符串和数值转换需要用strconv包
	// 这里只演示概念，后续课程详细讲解
	fmt.Printf("字符串转换: %s (需要strconv包)\n", stringValue)

	// 字符和数值转换
	var char rune = 'A'
	var ascii int = int(char)
	fmt.Printf("字符转ASCII: %c -> %d\n", char, ascii)

	var asciiCode int = 65
	var charFromAscii rune = rune(asciiCode)
	fmt.Printf("ASCII转字符: %d -> %c\n", asciiCode, charFromAscii)

	fmt.Println()
}

// 变量作用域演示
func demonstrateScope() {
	fmt.Println("5. 变量作用域:")

	// 包级别变量 (在函数外声明)
	fmt.Printf("包级别变量 packageVar: %s\n", packageVar)

	// 函数级别变量
	functionVar := "函数级别变量"
	fmt.Printf("函数级别变量: %s\n", functionVar)

	// 块级别变量
	if true {
		blockVar := "块级别变量"
		fmt.Printf("块级别变量: %s\n", blockVar)

		// 可以访问外层变量
		fmt.Printf("在块内访问函数变量: %s\n", functionVar)
	}

	// blockVar在这里不可访问，会编译错误
	// fmt.Println(blockVar) // 取消注释会报错

	fmt.Println()
}

// 包级别变量
var packageVar string = "我是包级别变量"

/*
=== 练习题 ===

1. 声明不同类型的变量并输出它们的零值
2. 尝试各种类型转换并观察结果
3. 创建一个程序计算圆的面积和周长（使用float64）
4. 实验变量作用域，在不同的块中声明同名变量
5. 使用fmt.Printf的不同格式化动词输出变量

运行命令：
go run main.go

扩展练习：
1. 研究Go中整型溢出的行为
2. 了解Unicode和UTF-8编码
3. 探索complex128的数学运算
*/
