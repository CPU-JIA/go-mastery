package calculator

import "math"

// Power 计算 a 的 b 次方
func Power(a, b float64) float64 {
	return math.Pow(a, b)
}

// SquareRoot 计算平方根
func SquareRoot(x float64) float64 {
	return math.Sqrt(x)
}

// Factorial 计算阶乘
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * Factorial(n-1)
}

// GCD 计算最大公约数
func GCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM 计算最小公倍数
func LCM(a, b int) int {
	return (a * b) / GCD(a, b)
}
