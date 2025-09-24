package main

import "fmt"

/*
=== Go语言第七课：数组(Arrays) ===

学习目标：
1. 理解数组的概念和特点
2. 掌握数组的声明和初始化
3. 学会数组的访问和修改
4. 了解数组的长度和容量
5. 掌握多维数组的使用

Go数组特点：
- 长度固定，编译时确定
- 元素类型相同
- 值类型，赋值时会复制整个数组
- 长度是类型的一部分
- 零值是所有元素都为对应类型的零值
*/

func main() {
	fmt.Println("=== Go语言数组学习 ===")

	// 1. 数组声明和初始化
	demonstrateArrayDeclaration()

	// 2. 数组访问和修改
	demonstrateArrayAccess()

	// 3. 数组遍历
	demonstrateArrayIteration()

	// 4. 多维数组
	demonstrateMultiDimensionalArrays()

	// 5. 数组比较和复制
	demonstrateArrayComparison()

	// 6. 数组作为函数参数
	demonstrateArrayAsParameter()

	// 7. 实际应用示例
	demonstratePracticalExamples()
}

// 数组声明和初始化
func demonstrateArrayDeclaration() {
	fmt.Println("1. 数组声明和初始化:")

	// 声明数组（零值初始化）
	var numbers [5]int
	fmt.Printf("零值数组: %v\n", numbers)

	// 声明并初始化
	var fruits [3]string = [3]string{"苹果", "香蕉", "橙子"}
	fmt.Printf("初始化数组: %v\n", fruits)

	// 简短声明
	colors := [4]string{"红色", "绿色", "蓝色", "黄色"}
	fmt.Printf("简短声明: %v\n", colors)

	// 让编译器推断长度
	scores := [...]int{85, 92, 78, 95, 88}
	fmt.Printf("推断长度: %v (长度: %d)\n", scores, len(scores))

	// 指定索引初始化
	weekdays := [7]string{
		0: "星期日",
		1: "星期一",
		2: "星期二",
		6: "星期六", // 其他位置为零值
	}
	fmt.Printf("指定索引: %v\n", weekdays)

	// 部分初始化
	partial := [10]int{1, 2, 3} // 前3个元素，其余为0
	fmt.Printf("部分初始化: %v\n", partial)

	// 混合初始化
	mixed := [6]int{0: 10, 2: 20, 5: 50}
	fmt.Printf("混合初始化: %v\n", mixed)

	fmt.Println()
}

// 数组访问和修改
func demonstrateArrayAccess() {
	fmt.Println("2. 数组访问和修改:")

	// 创建数组
	numbers := [5]int{10, 20, 30, 40, 50}
	fmt.Printf("原始数组: %v\n", numbers)

	// 访问元素
	fmt.Printf("第一个元素: %d\n", numbers[0])
	fmt.Printf("最后一个元素: %d\n", numbers[len(numbers)-1])

	// 修改元素
	numbers[2] = 35
	fmt.Printf("修改后: %v\n", numbers)

	// 数组长度
	fmt.Printf("数组长度: %d\n", len(numbers))

	// 获取数组地址
	fmt.Printf("数组地址: %p\n", &numbers)
	fmt.Printf("第一个元素地址: %p\n", &numbers[0])
	fmt.Printf("第二个元素地址: %p\n", &numbers[1])

	// 数组切片（创建引用）
	slice := numbers[1:4]
	fmt.Printf("数组切片[1:4]: %v\n", slice)

	// 边界检查（运行时panic）
	// fmt.Println(numbers[10]) // 这会导致panic

	fmt.Println()
}

// 数组遍历
func demonstrateArrayIteration() {
	fmt.Println("3. 数组遍历:")

	languages := [4]string{"Go", "Python", "Java", "JavaScript"}

	// 传统for循环
	fmt.Println("传统for循环:")
	for i := 0; i < len(languages); i++ {
		fmt.Printf("  [%d]: %s\n", i, languages[i])
	}

	// range循环（索引和值）
	fmt.Println("range循环（索引和值）:")
	for index, language := range languages {
		fmt.Printf("  [%d]: %s\n", index, language)
	}

	// range循环（只要索引）
	fmt.Println("range循环（只要索引）:")
	for index := range languages {
		fmt.Printf("  索引 %d\n", index)
	}

	// range循环（只要值）
	fmt.Println("range循环（只要值）:")
	for _, language := range languages {
		fmt.Printf("  语言: %s\n", language)
	}

	// 查找元素
	target := "Java"
	found := false
	foundIndex := -1

	for i, lang := range languages {
		if lang == target {
			found = true
			foundIndex = i
			break
		}
	}

	if found {
		fmt.Printf("找到 %s 在索引 %d\n", target, foundIndex)
	} else {
		fmt.Printf("未找到 %s\n", target)
	}

	fmt.Println()
}

// 多维数组
func demonstrateMultiDimensionalArrays() {
	fmt.Println("4. 多维数组:")

	// 二维数组声明
	var matrix [3][3]int
	fmt.Printf("零值二维数组:\n")
	printMatrix(matrix)

	// 二维数组初始化
	grid := [3][3]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}
	fmt.Printf("初始化二维数组:\n")
	printMatrix(grid)

	// 访问和修改二维数组
	fmt.Printf("元素[1][1]: %d\n", grid[1][1])
	grid[1][1] = 50
	fmt.Printf("修改后[1][1]: %d\n", grid[1][1])

	// 不规则二维数组初始化
	irregular := [3][4]int{
		{1, 2},       // 部分初始化
		{3, 4, 5},    // 部分初始化
		{6, 7, 8, 9}, // 完全初始化
	}
	fmt.Printf("不规则初始化:\n")
	for i := 0; i < 3; i++ {
		for j := 0; j < 4; j++ {
			fmt.Printf("%2d ", irregular[i][j])
		}
		fmt.Println()
	}

	// 三维数组
	cube := [2][2][2]int{
		{
			{1, 2},
			{3, 4},
		},
		{
			{5, 6},
			{7, 8},
		},
	}

	fmt.Println("三维数组:")
	for i := 0; i < 2; i++ {
		fmt.Printf("层 %d:\n", i)
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				fmt.Printf("  [%d][%d][%d]: %d\n", i, j, k, cube[i][j][k])
			}
		}
	}

	fmt.Println()
}

// 数组比较和复制
func demonstrateArrayComparison() {
	fmt.Println("5. 数组比较和复制:")

	// 数组比较（相同类型和长度才能比较）
	arr1 := [3]int{1, 2, 3}
	arr2 := [3]int{1, 2, 3}
	arr3 := [3]int{1, 2, 4}

	fmt.Printf("arr1: %v\n", arr1)
	fmt.Printf("arr2: %v\n", arr2)
	fmt.Printf("arr3: %v\n", arr3)

	fmt.Printf("arr1 == arr2: %t\n", arr1 == arr2)
	fmt.Printf("arr1 == arr3: %t\n", arr1 == arr3)
	fmt.Printf("arr1 != arr3: %t\n", arr1 != arr3)

	// 数组复制（值拷贝）
	original := [4]string{"a", "b", "c", "d"}
	copied := original // 完整复制

	fmt.Printf("原数组: %v\n", original)
	fmt.Printf("复制数组: %v\n", copied)

	// 修改复制的数组
	copied[0] = "modified"
	fmt.Printf("修改复制数组后:\n")
	fmt.Printf("  原数组: %v\n", original)
	fmt.Printf("  复制数组: %v\n", copied)

	// 数组地址比较
	fmt.Printf("原数组地址: %p\n", &original)
	fmt.Printf("复制数组地址: %p\n", &copied)

	fmt.Println()
}

// 数组作为函数参数
func demonstrateArrayAsParameter() {
	fmt.Println("6. 数组作为函数参数:")

	numbers := [5]int{1, 2, 3, 4, 5}
	fmt.Printf("原数组: %v\n", numbers)

	// 传递数组到函数（值传递）
	sum := sumArray(numbers)
	fmt.Printf("数组元素和: %d\n", sum)
	fmt.Printf("函数调用后原数组: %v\n", numbers)

	// 尝试修改数组（无效，因为是值传递）
	modifyArray(numbers)
	fmt.Printf("尝试修改后原数组: %v\n", numbers)

	// 通过指针修改数组
	modifyArrayByPointer(&numbers)
	fmt.Printf("通过指针修改后: %v\n", numbers)

	// 返回数组
	doubled := doubleArray(numbers)
	fmt.Printf("翻倍数组: %v\n", doubled)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("7. 实际应用示例:")

	// 1. 成绩统计
	fmt.Println("成绩统计:")
	scores := [10]float64{85.5, 92.0, 78.5, 95.0, 88.5, 91.0, 76.5, 89.0, 93.5, 87.0}

	total := 0.0
	max := scores[0]
	min := scores[0]

	for _, score := range scores {
		total += score
		if score > max {
			max = score
		}
		if score < min {
			min = score
		}
	}

	average := total / float64(len(scores))
	fmt.Printf("  平均分: %.2f\n", average)
	fmt.Printf("  最高分: %.1f\n", max)
	fmt.Printf("  最低分: %.1f\n", min)

	// 2. 频率统计
	fmt.Println("\n字符频率统计:")
	text := "hello world"
	var frequency [26]int // a-z的频率

	for _, char := range text {
		if char >= 'a' && char <= 'z' {
			frequency[char-'a']++
		}
	}

	for i, count := range frequency {
		if count > 0 {
			fmt.Printf("  '%c': %d次\n", 'a'+i, count)
		}
	}

	// 3. 简单排序（冒泡排序）
	fmt.Println("\n冒泡排序:")
	unsorted := [6]int{64, 34, 25, 12, 22, 11}
	fmt.Printf("  排序前: %v\n", unsorted)

	sorted := bubbleSort(unsorted)
	fmt.Printf("  排序后: %v\n", sorted)

	// 4. 矩阵运算
	fmt.Println("\n矩阵转置:")
	matrix := [3][2]int{
		{1, 2},
		{3, 4},
		{5, 6},
	}

	fmt.Println("  原矩阵:")
	for i := 0; i < 3; i++ {
		fmt.Printf("    ")
		for j := 0; j < 2; j++ {
			fmt.Printf("%d ", matrix[i][j])
		}
		fmt.Println()
	}

	// 转置
	transposed := [2][3]int{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			transposed[j][i] = matrix[i][j]
		}
	}

	fmt.Println("  转置后:")
	for i := 0; i < 2; i++ {
		fmt.Printf("    ")
		for j := 0; j < 3; j++ {
			fmt.Printf("%d ", transposed[i][j])
		}
		fmt.Println()
	}

	// 5. 数据查找
	fmt.Println("\n线性查找:")
	data := [8]int{2, 7, 11, 15, 23, 31, 45, 67}
	target := 23

	index := linearSearch(data, target)
	if index != -1 {
		fmt.Printf("  找到 %d 在索引 %d\n", target, index)
	} else {
		fmt.Printf("  未找到 %d\n", target)
	}

	fmt.Println()
}

// 辅助函数

// 打印3x3矩阵
func printMatrix(matrix [3][3]int) {
	for i := 0; i < 3; i++ {
		fmt.Print("  ")
		for j := 0; j < 3; j++ {
			fmt.Printf("%2d ", matrix[i][j])
		}
		fmt.Println()
	}
}

// 计算数组元素和
func sumArray(arr [5]int) int {
	sum := 0
	for _, value := range arr {
		sum += value
	}
	return sum
}

// 尝试修改数组（无效）
func modifyArray(arr [5]int) {
	arr[0] = 999 // 只修改副本
}

// 通过指针修改数组
func modifyArrayByPointer(arr *[5]int) {
	arr[0] = 999 // 修改原数组
}

// 返回翻倍的数组
func doubleArray(arr [5]int) [5]int {
	var result [5]int
	for i, value := range arr {
		result[i] = value * 2
	}
	return result
}

// 冒泡排序
func bubbleSort(arr [6]int) [6]int {
	result := arr // 复制数组
	n := len(result)

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

// 线性查找
func linearSearch(arr [8]int, target int) int {
	for i, value := range arr {
		if value == target {
			return i
		}
	}
	return -1
}

/*
=== 练习题 ===

1. 编写一个函数，找出数组中第二大的元素

2. 实现选择排序算法对数组进行排序

3. 编写一个函数，检查两个数组是否包含相同的元素（不考虑顺序）

4. 创建一个3x3的井字棋游戏板，实现胜利条件检查

5. 实现矩阵乘法（两个2x2矩阵相乘）

6. 编写一个函数，将数组向左或向右旋转n个位置

7. 实现一个简单的图像滤镜（在3x3像素矩阵上应用滤镜）

运行命令：
go run main.go

高级练习：
1. 实现快速排序的非递归版本
2. 编写一个稀疏矩阵的压缩表示
3. 实现一个简单的迷宫求解算法
4. 创建一个俄罗斯方块的游戏板检查器
5. 实现一个数独验证器

注意事项：
- 数组长度固定，需要在编译时确定
- 数组是值类型，传递时会复制整个数组
- 访问数组时注意边界检查
- 多维数组在内存中是连续存储的
- 考虑使用切片(slice)替代数组以获得更多灵活性
*/
