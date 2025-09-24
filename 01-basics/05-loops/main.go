package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
=== Go语言第五课：循环语句(for) ===

学习目标：
1. 掌握for循环的三种形式
2. 理解range关键字的使用
3. 学会使用break和continue控制循环
4. 掌握无限循环和循环嵌套
5. 了解Go中没有while和do-while

Go循环特点：
- 只有for一种循环语句
- 可以模拟while和do-while循环
- range可以遍历数组、切片、字符串、映射、通道
- 支持标签和goto (不推荐使用goto)
*/

func main() {
	fmt.Println("=== Go语言循环语句学习 ===")

	// 1. 基本for循环
	demonstrateBasicFor()

	// 2. for作为while循环
	demonstrateWhileLoop()

	// 3. 无限循环
	demonstrateInfiniteLoop()

	// 4. range循环
	demonstrateRangeLoop()

	// 5. break和continue
	demonstrateBreakContinue()

	// 6. 嵌套循环
	demonstrateNestedLoops()

	// 7. 标签和跳转
	demonstrateLabels()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本for循环
func demonstrateBasicFor() {
	fmt.Println("1. 基本for循环:")

	// 标准三段式for循环
	fmt.Print("数字1-5: ")
	for i := 1; i <= 5; i++ {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 倒序循环
	fmt.Print("倒数5-1: ")
	for i := 5; i >= 1; i-- {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 步长为2的循环
	fmt.Print("奇数1-9: ")
	for i := 1; i <= 9; i += 2 {
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// 多个变量的循环
	fmt.Print("双变量循环: ")
	for i, j := 0, 10; i < 5; i, j = i+1, j-1 {
		fmt.Printf("(%d,%d) ", i, j)
	}
	fmt.Println()

	// 省略初始化语句
	k := 0
	fmt.Print("省略初始化: ")
	for ; k < 3; k++ {
		fmt.Printf("%d ", k)
	}
	fmt.Println()

	fmt.Println()
}

// for作为while循环
func demonstrateWhileLoop() {
	fmt.Println("2. for模拟while循环:")

	// 模拟while循环
	count := 0
	fmt.Print("while风格: ")
	for count < 5 {
		fmt.Printf("%d ", count)
		count++
	}
	fmt.Println()

	// 条件循环示例
	sum := 0
	i := 1
	fmt.Print("累加到和大于100: ")
	for sum <= 100 {
		sum += i
		fmt.Printf("%d ", i)
		i++
	}
	fmt.Printf("\n和: %d\n", sum)

	// 随机数示例
	rand.Seed(time.Now().UnixNano())
	fmt.Print("随机数直到得到6: ")
	for {
		num := rand.Intn(6) + 1
		fmt.Printf("%d ", num)
		if num == 6 {
			break
		}
	}
	fmt.Println()

	fmt.Println()
}

// 无限循环
func demonstrateInfiniteLoop() {
	fmt.Println("3. 无限循环示例:")

	// 计数器示例
	counter := 0
	fmt.Print("计数到5后退出: ")
	for {
		counter++
		fmt.Printf("%d ", counter)
		if counter >= 5 {
			break
		}
	}
	fmt.Println()

	// 菜单循环模拟
	fmt.Println("模拟菜单循环:")
	iteration := 0
	for {
		iteration++
		// 模拟用户选择
		if iteration == 1 {
			fmt.Println("  选择了选项1")
		} else if iteration == 2 {
			fmt.Println("  选择了选项2")
		} else {
			fmt.Println("  选择了退出")
			break
		}
	}

	fmt.Println()
}

// range循环
func demonstrateRangeLoop() {
	fmt.Println("4. range循环:")

	// 遍历数组
	numbers := [5]int{10, 20, 30, 40, 50}
	fmt.Println("遍历数组:")
	for index, value := range numbers {
		fmt.Printf("  索引%d: 值%d\n", index, value)
	}

	// 只要索引
	fmt.Print("只要索引: ")
	for index := range numbers {
		fmt.Printf("%d ", index)
	}
	fmt.Println()

	// 只要值
	fmt.Print("只要值: ")
	for _, value := range numbers {
		fmt.Printf("%d ", value)
	}
	fmt.Println()

	// 遍历切片
	fruits := []string{"苹果", "香蕉", "橙子", "葡萄"}
	fmt.Println("遍历切片:")
	for i, fruit := range fruits {
		fmt.Printf("  %d: %s\n", i, fruit)
	}

	// 遍历字符串
	text := "Hello"
	fmt.Println("遍历字符串:")
	for index, char := range text {
		fmt.Printf("  位置%d: 字符%c (Unicode: %d)\n", index, char, char)
	}

	// 遍历中文字符串
	chinese := "你好世界"
	fmt.Println("遍历中文字符串:")
	for index, char := range chinese {
		fmt.Printf("  位置%d: 字符%c (Unicode: %d)\n", index, char, char)
	}

	// 遍历map
	colors := map[string]string{
		"red":   "红色",
		"green": "绿色",
		"blue":  "蓝色",
	}
	fmt.Println("遍历映射:")
	for key, value := range colors {
		fmt.Printf("  %s: %s\n", key, value)
	}

	fmt.Println()
}

// break和continue
func demonstrateBreakContinue() {
	fmt.Println("5. break和continue:")

	// continue示例
	fmt.Print("跳过偶数(1-10): ")
	for i := 1; i <= 10; i++ {
		if i%2 == 0 {
			continue // 跳过当前迭代
		}
		fmt.Printf("%d ", i)
	}
	fmt.Println()

	// break示例
	fmt.Print("找到第一个大于15的数字: ")
	for i := 10; i <= 30; i++ {
		if i > 15 {
			fmt.Printf("%d", i)
			break // 退出循环
		}
	}
	fmt.Println()

	// 复杂条件控制
	fmt.Println("处理数据列表:")
	data := []int{1, 2, 0, 4, -1, 6, 7, 0, 9, 10}
	positiveSum := 0
	processedCount := 0

	for i, value := range data {
		if value == 0 {
			fmt.Printf("  跳过索引%d的零值\n", i)
			continue
		}

		if value < 0 {
			fmt.Printf("  遇到负数%d，停止处理\n", value)
			break
		}

		positiveSum += value
		processedCount++
		fmt.Printf("  处理%d，累计和:%d\n", value, positiveSum)
	}

	fmt.Printf("总共处理了%d个正数，和为%d\n", processedCount, positiveSum)

	fmt.Println()
}

// 嵌套循环
func demonstrateNestedLoops() {
	fmt.Println("6. 嵌套循环:")

	// 乘法表
	fmt.Println("九九乘法表:")
	for i := 1; i <= 9; i++ {
		for j := 1; j <= i; j++ {
			fmt.Printf("%d×%d=%2d  ", j, i, i*j)
		}
		fmt.Println()
	}

	// 矩阵遍历
	fmt.Println("3×3矩阵:")
	matrix := [3][3]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	for row := range matrix {
		for col := range matrix[row] {
			fmt.Printf("%d ", matrix[row][col])
		}
		fmt.Println()
	}

	// 查找模式
	fmt.Println("在矩阵中查找数字5:")
	found := false
	for i := 0; i < 3 && !found; i++ {
		for j := 0; j < 3; j++ {
			if matrix[i][j] == 5 {
				fmt.Printf("找到5在位置[%d][%d]\n", i, j)
				found = true
				break
			}
		}
	}

	fmt.Println()
}

// 标签和跳转
func demonstrateLabels() {
	fmt.Println("7. 标签和跳转:")

	// 使用标签跳出嵌套循环
	fmt.Println("跳出嵌套循环示例:")

OuterLoop:
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 3; j++ {
			fmt.Printf("i=%d, j=%d\n", i, j)
			if i == 2 && j == 2 {
				fmt.Println("条件满足，跳出所有循环")
				break OuterLoop
			}
		}
	}
	fmt.Println("跳出后继续执行")

	// 标签continue
	fmt.Println("\n标签continue示例:")

OuterLoop2:
	for i := 1; i <= 3; i++ {
		fmt.Printf("外层循环 i=%d\n", i)
		for j := 1; j <= 3; j++ {
			if j == 2 {
				fmt.Printf("  j=%d时跳到外层下一次迭代\n", j)
				continue OuterLoop2
			}
			fmt.Printf("  内层循环 j=%d\n", j)
		}
	}

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 素数检测
	fmt.Println("10以内的素数:")
	for num := 2; num <= 10; num++ {
		isPrime := true
		for i := 2; i < num; i++ {
			if num%i == 0 {
				isPrime = false
				break
			}
		}
		if isPrime {
			fmt.Printf("%d ", num)
		}
	}
	fmt.Println()

	// 2. 斐波那契数列
	fmt.Print("斐波那契数列前10项: ")
	a, b := 0, 1
	for i := 0; i < 10; i++ {
		fmt.Printf("%d ", a)
		a, b = b, a+b
	}
	fmt.Println()

	// 3. 数组求和与平均值
	scores := []float64{85.5, 92.0, 78.5, 95.0, 88.5}
	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	average := sum / float64(len(scores))
	fmt.Printf("成绩总和: %.1f, 平均分: %.2f\n", sum, average)

	// 4. 字符统计
	text := "Hello, Go World!"
	charCount := make(map[rune]int)
	for _, char := range text {
		charCount[char]++
	}
	fmt.Println("字符出现次数:")
	for char, count := range charCount {
		if char != ' ' && char != ',' && char != '!' {
			fmt.Printf("  '%c': %d次\n", char, count)
		}
	}

	// 5. 查找最大值和最小值
	numbers := []int{23, 45, 12, 67, 89, 34, 56}
	max, min := numbers[0], numbers[0]
	for _, num := range numbers {
		if num > max {
			max = num
		}
		if num < min {
			min = num
		}
	}
	fmt.Printf("数组中最大值: %d, 最小值: %d\n", max, min)

	// 6. 数据过滤
	fmt.Println("筛选大于50的数字:")
	for i, num := range numbers {
		if num > 50 {
			fmt.Printf("  索引%d: %d\n", i, num)
		}
	}

	fmt.Println()
}

/*
=== 练习题 ===

1. 编写程序计算1到100的所有偶数之和

2. 实现一个简单的猜数字游戏：
   - 生成1-100的随机数
   - 用户输入猜测(用固定数组模拟)
   - 给出高了/低了的提示

3. 编写一个函数打印指定大小的菱形图案

4. 实现冒泡排序算法

5. 编写程序统计一段文本中每个单词的出现频率

6. 实现一个简单的计算器解析器(处理简单的四则运算)

运行命令：
go run main.go

高级练习：
1. 实现快速排序算法
2. 编写一个简单的正则表达式匹配器
3. 实现一个基于循环的状态机
4. 创建一个数独求解器
5. 编写一个文本格式化器

注意事项：
- 避免无限循环导致程序崩溃
- 合理使用break和continue
- 嵌套循环要考虑时间复杂度
- range在循环过程中修改集合要小心
- 大量数据处理考虑使用并发
*/
