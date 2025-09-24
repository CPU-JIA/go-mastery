package main

import "fmt"

/*
=== Go语言第一课：Hello World ===

学习目标：
1. 了解Go程序的基本结构
2. 掌握package和import的使用
3. 理解main函数的作用
4. 学会使用fmt包进行输出

知识点：
- package main: 定义可执行程序的包
- import "fmt": 导入格式化I/O包
- func main(): 程序入口点
- fmt.Println(): 输出并换行
*/

func main() {
	// 基础输出
	fmt.Println("🚀 欢迎来到Go语言世界！")
	fmt.Println("Hello, World!")

	// 格式化输出
	name := "JIA总"
	language := "Go"
	fmt.Printf("您好 %s，欢迎学习 %s 语言！\n", name, language)

	// 多种输出方式
	fmt.Print("这是 Print: 不换行输出")
	fmt.Print(" 继续输出\n")

	fmt.Println("这是 Println: 自动换行输出")

	// Sprintf: 格式化为字符串
	message := fmt.Sprintf("学习进度: %d%%", 1)
	fmt.Println(message)
}

/*
=== 练习题 ===

1. 修改程序，输出您的姓名和今天的日期
2. 使用fmt.Printf输出一个格式化的表格
3. 尝试不同的格式化动词：%s, %d, %f, %t, %v
4. 创建一个输出多行ASCII艺术字的程序

运行命令：
go run main.go

预期输出：
🚀 欢迎来到Go语言世界！
Hello, World!
您好 JIA总，欢迎学习 Go 语言！
这是 Print: 不换行输出 继续输出
这是 Println: 自动换行输出
学习进度: 1%
*/
