// Package calculator 提供基本的数学运算功能
package calculator

// Add 执行两个整数的加法运算
func Add(a, b int) int {
	return a + b
}

// Subtract 执行两个整数的减法运算
func Subtract(a, b int) int {
	return a - b
}

// Multiply 执行两个整数的乘法运算
func Multiply(a, b int) int {
	return a * b
}

// Divide 执行两个整数的除法运算，返回商和余数
func Divide(a, b int) (int, int) {
	if b == 0 {
		return 0, 0
	}
	return a / b, a % b
}
