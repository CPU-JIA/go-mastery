package main

import (
	"fmt"
	"sort"
)

/*
=== Go语言第八课：切片(Slices) ===

学习目标：
1. 理解切片和数组的区别
2. 掌握切片的创建和初始化方法
3. 学会切片的扩容机制
4. 掌握切片的常用操作
5. 理解切片的底层原理

Go切片特点：
- 动态数组，长度可变
- 引用类型，指向底层数组
- 包含指针、长度、容量三个字段
- 零值为nil
- 自动扩容机制
*/

func main() {
	fmt.Println("=== Go语言切片学习 ===")

	// 1. 切片创建和初始化
	demonstrateSliceCreation()

	// 2. 切片操作
	demonstrateSliceOperations()

	// 3. 切片扩容机制
	demonstrateSliceExpansion()

	// 4. 切片的切片
	demonstrateSliceSlicing()

	// 5. 切片复制和比较
	demonstrateSliceCopyAndComparison()

	// 6. 多维切片
	demonstrateMultiDimensionalSlices()

	// 7. 切片作为函数参数
	demonstrateSliceAsParameter()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 切片创建和初始化
func demonstrateSliceCreation() {
	fmt.Println("1. 切片创建和初始化:")

	// 1. nil切片
	var nilSlice []int
	fmt.Printf("nil切片: %v, 长度: %d, 容量: %d, 是否为nil: %t\n",
		nilSlice, len(nilSlice), cap(nilSlice), nilSlice == nil)

	// 2. 空切片
	emptySlice := []int{}
	fmt.Printf("空切片: %v, 长度: %d, 容量: %d, 是否为nil: %t\n",
		emptySlice, len(emptySlice), cap(emptySlice), emptySlice == nil)

	// 3. 字面量创建
	fruits := []string{"苹果", "香蕉", "橙子"}
	fmt.Printf("字面量切片: %v, 长度: %d, 容量: %d\n",
		fruits, len(fruits), cap(fruits))

	// 4. make函数创建
	makeSlice := make([]int, 5) // 长度5，容量5
	fmt.Printf("make切片(5): %v, 长度: %d, 容量: %d\n",
		makeSlice, len(makeSlice), cap(makeSlice))

	makeSliceWithCap := make([]int, 3, 10) // 长度3，容量10
	fmt.Printf("make切片(3,10): %v, 长度: %d, 容量: %d\n",
		makeSliceWithCap, len(makeSliceWithCap), cap(makeSliceWithCap))

	// 5. 从数组创建切片
	array := [5]int{1, 2, 3, 4, 5}
	arraySlice := array[1:4] // 从数组创建切片
	fmt.Printf("从数组创建: %v, 长度: %d, 容量: %d\n",
		arraySlice, len(arraySlice), cap(arraySlice))

	// 6. 从切片创建切片
	subSlice := fruits[0:2]
	fmt.Printf("从切片创建: %v, 长度: %d, 容量: %d\n",
		subSlice, len(subSlice), cap(subSlice))

	fmt.Println()
}

// 切片操作
func demonstrateSliceOperations() {
	fmt.Println("2. 切片操作:")

	// 创建切片
	numbers := []int{1, 2, 3, 4, 5}
	fmt.Printf("原始切片: %v\n", numbers)

	// 访问元素
	fmt.Printf("第一个元素: %d\n", numbers[0])
	fmt.Printf("最后一个元素: %d\n", numbers[len(numbers)-1])

	// 修改元素
	numbers[2] = 30
	fmt.Printf("修改后: %v\n", numbers)

	// append操作
	numbers = append(numbers, 6)
	fmt.Printf("append(6): %v, 长度: %d, 容量: %d\n",
		numbers, len(numbers), cap(numbers))

	// append多个元素
	numbers = append(numbers, 7, 8, 9)
	fmt.Printf("append(7,8,9): %v, 长度: %d, 容量: %d\n",
		numbers, len(numbers), cap(numbers))

	// append另一个切片
	moreNumbers := []int{10, 11, 12}
	numbers = append(numbers, moreNumbers...)
	fmt.Printf("append切片: %v, 长度: %d, 容量: %d\n",
		numbers, len(numbers), cap(numbers))

	// 删除元素（通过切片重组）
	fmt.Println("\n删除操作:")
	original := []int{1, 2, 3, 4, 5}
	fmt.Printf("原始: %v\n", original)

	// 删除第一个元素
	withoutFirst := original[1:]
	fmt.Printf("删除第一个: %v\n", withoutFirst)

	// 删除最后一个元素
	withoutLast := original[:len(original)-1]
	fmt.Printf("删除最后一个: %v\n", withoutLast)

	// 删除中间元素（索引2）
	index := 2
	withoutMiddle := append(original[:index], original[index+1:]...)
	fmt.Printf("删除索引%d: %v\n", index, withoutMiddle)

	// 插入元素
	fmt.Println("\n插入操作:")
	data := []int{1, 2, 4, 5}
	insertIndex := 2
	insertValue := 3

	// 插入到中间
	data = append(data[:insertIndex], append([]int{insertValue}, data[insertIndex:]...)...)
	fmt.Printf("在索引%d插入%d: %v\n", insertIndex, insertValue, data)

	fmt.Println()
}

// 切片扩容机制
func demonstrateSliceExpansion() {
	fmt.Println("3. 切片扩容机制:")

	// 观察扩容过程
	slice := make([]int, 0, 1)
	fmt.Printf("初始: 长度=%d, 容量=%d\n", len(slice), cap(slice))

	for i := 0; i < 10; i++ {
		oldCap := cap(slice)
		slice = append(slice, i)
		newCap := cap(slice)

		fmt.Printf("添加%d: 长度=%d, 容量=%d", i, len(slice), newCap)
		if newCap != oldCap {
			fmt.Printf(" (扩容: %d -> %d)", oldCap, newCap)
		}
		fmt.Println()
	}

	// 大批量扩容
	fmt.Println("\n大批量扩容:")
	bigSlice := make([]int, 0)

	for i := 0; i < 5; i++ {
		oldCap := cap(bigSlice)
		bigSlice = append(bigSlice, make([]int, 1000)...)
		newCap := cap(bigSlice)

		fmt.Printf("添加1000个元素: 长度=%d, 容量=%d (扩容倍数: %.2f)\n",
			len(bigSlice), newCap, float64(newCap)/float64(oldCap))
	}

	// 预分配容量避免扩容
	fmt.Println("\n预分配容量:")
	preAllocated := make([]int, 0, 10)
	fmt.Printf("预分配容量10: 长度=%d, 容量=%d\n", len(preAllocated), cap(preAllocated))

	for i := 0; i < 10; i++ {
		oldCap := cap(preAllocated)
		preAllocated = append(preAllocated, i)
		newCap := cap(preAllocated)

		if newCap != oldCap {
			fmt.Printf("意外扩容在索引%d\n", i)
		}
	}
	fmt.Printf("添加10个元素后: 长度=%d, 容量=%d (无扩容)\n",
		len(preAllocated), cap(preAllocated))

	fmt.Println()
}

// 切片的切片
func demonstrateSliceSlicing() {
	fmt.Println("4. 切片的切片:")

	// 创建原始切片
	original := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	fmt.Printf("原始切片: %v\n", original)

	// 各种切片操作
	slice1 := original[2:5] // [2, 3, 4]
	slice2 := original[:3]  // [0, 1, 2]
	slice3 := original[7:]  // [7, 8, 9]
	slice4 := original[:]   // 完整复制

	fmt.Printf("original[2:5]: %v\n", slice1)
	fmt.Printf("original[:3]: %v\n", slice2)
	fmt.Printf("original[7:]: %v\n", slice3)
	fmt.Printf("original[:]: %v\n", slice4)

	// 三参数切片：[low:high:max]
	slice5 := original[2:5:6] // 长度3，容量4
	fmt.Printf("original[2:5:6]: %v, 长度: %d, 容量: %d\n",
		slice5, len(slice5), cap(slice5))

	// 共享底层数组
	fmt.Println("\n共享底层数组演示:")
	numbers := []int{1, 2, 3, 4, 5}
	sub1 := numbers[1:3] // [2, 3]
	sub2 := numbers[2:4] // [3, 4]

	fmt.Printf("原始: %v\n", numbers)
	fmt.Printf("sub1: %v\n", sub1)
	fmt.Printf("sub2: %v\n", sub2)

	// 修改sub1影响原数组和sub2
	sub1[1] = 30
	fmt.Printf("修改sub1[1]=30后:\n")
	fmt.Printf("原始: %v\n", numbers)
	fmt.Printf("sub1: %v\n", sub1)
	fmt.Printf("sub2: %v\n", sub2)

	// 切片扩容后不再共享
	fmt.Println("\n扩容后的独立性:")
	small := []int{1, 2}
	sub := small[0:1]
	fmt.Printf("small: %v, sub: %v\n", small, sub)

	sub = append(sub, 10, 20, 30) // 扩容
	fmt.Printf("sub扩容后: %v\n", sub)
	fmt.Printf("small不受影响: %v\n", small)

	fmt.Println()
}

// 切片复制和比较
func demonstrateSliceCopyAndComparison() {
	fmt.Println("5. 切片复制和比较:")

	// 切片不能直接比较（除了与nil比较）
	slice1 := []int{1, 2, 3}
	slice2 := []int{1, 2, 3}
	var nilSlice []int

	// fmt.Println(slice1 == slice2) // 编译错误！
	fmt.Printf("slice1: %v\n", slice1)
	fmt.Printf("slice2: %v\n", slice2)
	fmt.Printf("slice1 == nil: %t\n", slice1 == nil)
	fmt.Printf("nilSlice == nil: %t\n", nilSlice == nil)

	// 手动比较切片
	isEqual := equalSlices(slice1, slice2)
	fmt.Printf("slice1和slice2相等: %t\n", isEqual)

	// copy函数
	fmt.Println("\ncopy函数使用:")
	source := []int{1, 2, 3, 4, 5}
	dest := make([]int, 3)

	n := copy(dest, source)
	fmt.Printf("源切片: %v\n", source)
	fmt.Printf("目标切片: %v\n", dest)
	fmt.Printf("复制了%d个元素\n", n)

	// 复制到更大的切片
	bigDest := make([]int, 10)
	n = copy(bigDest, source)
	fmt.Printf("复制到大切片: %v, 复制了%d个元素\n", bigDest, n)

	// 重叠复制
	fmt.Println("\n重叠复制:")
	data := []int{1, 2, 3, 4, 5}
	copy(data[2:], data[0:3]) // 将前3个元素复制到索引2开始的位置
	fmt.Printf("重叠复制后: %v\n", data)

	// 深度复制
	fmt.Println("\n深度复制:")
	original := []int{1, 2, 3, 4, 5}
	deepCopy := make([]int, len(original))
	copy(deepCopy, original)

	fmt.Printf("原始: %v\n", original)
	fmt.Printf("深度复制: %v\n", deepCopy)

	// 修改不会相互影响
	deepCopy[0] = 100
	fmt.Printf("修改复制后:\n")
	fmt.Printf("原始: %v\n", original)
	fmt.Printf("深度复制: %v\n", deepCopy)

	fmt.Println()
}

// 多维切片
func demonstrateMultiDimensionalSlices() {
	fmt.Println("6. 多维切片:")

	// 创建二维切片
	matrix := [][]int{
		{1, 2, 3},
		{4, 5, 6},
		{7, 8, 9},
	}

	fmt.Println("二维切片:")
	printSliceMatrix(matrix)

	// 动态创建二维切片
	rows, cols := 3, 4
	dynamic := make([][]int, rows)
	for i := range dynamic {
		dynamic[i] = make([]int, cols)
		for j := range dynamic[i] {
			dynamic[i][j] = i*cols + j + 1
		}
	}

	fmt.Printf("\n动态创建的%dx%d矩阵:\n", rows, cols)
	printSliceMatrix(dynamic)

	// 不规则切片（锯齿数组）
	jagged := [][]int{
		{1},
		{2, 3},
		{4, 5, 6},
		{7, 8, 9, 10},
	}

	fmt.Println("\n锯齿数组:")
	for i, row := range jagged {
		fmt.Printf("行%d: %v\n", i, row)
	}

	// 三维切片
	cube := [][][]int{
		{
			{1, 2},
			{3, 4},
		},
		{
			{5, 6},
			{7, 8},
		},
	}

	fmt.Println("\n三维切片:")
	for i, layer := range cube {
		fmt.Printf("层%d:\n", i)
		for j, row := range layer {
			fmt.Printf("  行%d: %v\n", j, row)
		}
	}

	fmt.Println()
}

// 切片作为函数参数
func demonstrateSliceAsParameter() {
	fmt.Println("7. 切片作为函数参数:")

	numbers := []int{1, 2, 3, 4, 5}
	fmt.Printf("原始切片: %v\n", numbers)

	// 传递切片（传递的是切片结构体，但底层数组是共享的）
	sum := sumSlice(numbers)
	fmt.Printf("切片和: %d\n", sum)

	// 在函数中修改切片元素
	modifySliceElements(numbers)
	fmt.Printf("修改元素后: %v\n", numbers)

	// 在函数中追加元素（可能不影响原切片）
	fmt.Printf("append前: %v, 容量: %d\n", numbers, cap(numbers))
	appendToSlice(numbers)
	fmt.Printf("append后: %v, 容量: %d\n", numbers, cap(numbers))

	// 返回修改后的切片
	doubled := doubleSlice(numbers)
	fmt.Printf("翻倍后: %v\n", doubled)
	fmt.Printf("原切片: %v\n", numbers)

	// 切片指针
	modifySliceByPointer(&numbers)
	fmt.Printf("通过指针修改后: %v\n", numbers)

	fmt.Println()
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 动态数组
	fmt.Println("动态数组应用:")
	var dynamicArray []int

	// 模拟动态添加数据
	for i := 0; i < 5; i++ {
		dynamicArray = append(dynamicArray, i*i)
		fmt.Printf("  添加%d²: %v\n", i, dynamicArray)
	}

	// 2. 数据过滤
	fmt.Println("\n数据过滤:")
	allNumbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	evenNumbers := filterEven(allNumbers)
	fmt.Printf("  所有数字: %v\n", allNumbers)
	fmt.Printf("  偶数: %v\n", evenNumbers)

	// 3. 数据转换
	fmt.Println("\n数据转换:")
	strings := []string{"1", "2", "3", "4", "5"}
	integers := stringSliceToIntSlice(strings)
	fmt.Printf("  字符串: %v\n", strings)
	fmt.Printf("  整数: %v\n", integers)

	// 4. 切片排序
	fmt.Println("\n切片排序:")
	unsorted := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Printf("  排序前: %v\n", unsorted)

	// 使用标准库排序
	sorted := make([]int, len(unsorted))
	copy(sorted, unsorted)
	sort.Ints(sorted)
	fmt.Printf("  排序后: %v\n", sorted)

	// 5. 切片去重
	fmt.Println("\n切片去重:")
	withDuplicates := []int{1, 2, 2, 3, 3, 3, 4, 4, 5}
	unique := removeDuplicates(withDuplicates)
	fmt.Printf("  原始: %v\n", withDuplicates)
	fmt.Printf("  去重: %v\n", unique)

	// 6. 切片反转
	fmt.Println("\n切片反转:")
	original := []string{"a", "b", "c", "d", "e"}
	reversed := reverseSlice(original)
	fmt.Printf("  原始: %v\n", original)
	fmt.Printf("  反转: %v\n", reversed)

	// 7. 分页处理
	fmt.Println("\n分页处理:")
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	pageSize := 4

	for page := 0; page*pageSize < len(data); page++ {
		start := page * pageSize
		end := start + pageSize
		if end > len(data) {
			end = len(data)
		}
		pageData := data[start:end]
		fmt.Printf("  第%d页: %v\n", page+1, pageData)
	}

	fmt.Println()
}

// 辅助函数

// 打印切片矩阵
func printSliceMatrix(matrix [][]int) {
	for i, row := range matrix {
		fmt.Printf("行%d: ", i)
		for _, val := range row {
			fmt.Printf("%3d ", val)
		}
		fmt.Println()
	}
}

// 比较两个切片是否相等
func equalSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// 计算切片元素和
func sumSlice(slice []int) int {
	sum := 0
	for _, value := range slice {
		sum += value
	}
	return sum
}

// 修改切片元素
func modifySliceElements(slice []int) {
	for i := range slice {
		slice[i] *= 2
	}
}

// 尝试追加元素（可能不影响原切片）
func appendToSlice(slice []int) {
	slice = append(slice, 999)
	fmt.Printf("  函数内append后: %v\n", slice)
}

// 返回翻倍的切片
func doubleSlice(slice []int) []int {
	result := make([]int, len(slice))
	for i, value := range slice {
		result[i] = value * 2
	}
	return result
}

// 通过指针修改切片
func modifySliceByPointer(slice *[]int) {
	*slice = append(*slice, 1000)
}

// 过滤偶数
func filterEven(numbers []int) []int {
	var result []int
	for _, num := range numbers {
		if num%2 == 0 {
			result = append(result, num)
		}
	}
	return result
}

// 字符串切片转整数切片
func stringSliceToIntSlice(strings []string) []int {
	result := make([]int, len(strings))
	for i, s := range strings {
		// 简单转换，实际应该处理错误
		result[i] = int(s[0] - '0')
	}
	return result
}

// 去重
func removeDuplicates(slice []int) []int {
	keys := make(map[int]bool)
	var result []int

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// 反转切片
func reverseSlice(slice []string) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

/*
=== 练习题 ===

1. 实现一个函数，合并两个有序切片并保持有序

2. 编写一个函数，在切片中查找所有满足条件的元素索引

3. 实现一个通用的切片映射函数（类似map函数）

4. 编写一个函数，将一个切片分割成指定大小的子切片

5. 实现一个切片的洗牌算法

6. 编写一个函数，找出两个切片的交集和并集

7. 实现一个环形缓冲区

运行命令：
go run main.go

高级练习：
1. 实现一个基于切片的栈和队列
2. 编写一个切片的快速排序实现
3. 实现一个动态规划的最长公共子序列
4. 创建一个基于切片的最小堆
5. 实现一个切片的归并排序

重要概念：
- 切片是引用类型，共享底层数组
- append可能导致扩容和重新分配
- 切片的零值是nil
- 使用copy进行深度复制
- 三参数切片可以限制容量
*/
