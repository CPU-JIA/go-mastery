// Package calculator 提供基本和高级数学运算功能的测试
package calculator

import (
	"math"
	"testing"
)

// =============================================================================
// 基础运算测试 (basic.go)
// =============================================================================

// TestAdd 测试加法运算
func TestAdd(t *testing.T) {
	// 表驱动测试用例
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		// 正向用例
		{"正数相加", 2, 3, 5},
		{"零值相加", 0, 0, 0},
		{"负数相加", -2, -3, -5},
		{"正负相加", 5, -3, 2},
		// 边界条件
		{"大数相加", 1000000, 2000000, 3000000},
		{"最小值边界", -1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d; 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestSubtract 测试减法运算
func TestSubtract(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		// 正向用例
		{"正数相减", 5, 3, 2},
		{"零值相减", 0, 0, 0},
		{"负数相减", -5, -3, -2},
		{"正减负", 5, -3, 8},
		// 边界条件
		{"结果为负", 3, 5, -2},
		{"大数相减", 1000000, 500000, 500000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Subtract(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Subtract(%d, %d) = %d; 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestMultiply 测试乘法运算
func TestMultiply(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		// 正向用例
		{"正数相乘", 4, 5, 20},
		{"零值相乘", 5, 0, 0},
		{"负数相乘", -4, -5, 20},
		{"正负相乘", 4, -5, -20},
		// 边界条件
		{"乘以1", 100, 1, 100},
		{"乘以-1", 100, -1, -100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%d, %d) = %d; 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestDivide 测试除法运算
func TestDivide(t *testing.T) {
	tests := []struct {
		name              string
		a, b              int
		expectedQuotient  int
		expectedRemainder int
	}{
		// 正向用例
		{"整除", 10, 2, 5, 0},
		{"有余数", 10, 3, 3, 1},
		{"负数除法", -10, 3, -3, -1},
		// 边界条件
		{"除以1", 100, 1, 100, 0},
		{"被除数为0", 0, 5, 0, 0},
		// 异常路径 - 除以0
		{"除以0", 10, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotient, remainder := Divide(tt.a, tt.b)
			if quotient != tt.expectedQuotient || remainder != tt.expectedRemainder {
				t.Errorf("Divide(%d, %d) = (%d, %d); 期望 (%d, %d)",
					tt.a, tt.b, quotient, remainder, tt.expectedQuotient, tt.expectedRemainder)
			}
		})
	}
}

// =============================================================================
// 高级运算测试 (advanced.go)
// =============================================================================

// TestPower 测试幂运算
func TestPower(t *testing.T) {
	tests := []struct {
		name     string
		a, b     float64
		expected float64
	}{
		// 正向用例
		{"2的3次方", 2, 3, 8},
		{"10的2次方", 10, 2, 100},
		{"负数的偶次方", -2, 2, 4},
		// 边界条件
		{"任何数的0次方", 5, 0, 1},
		{"0的正次方", 0, 5, 0},
		{"1的任意次方", 1, 100, 1},
		{"负指数", 2, -1, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Power(tt.a, tt.b)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Power(%f, %f) = %f; 期望 %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestSquareRoot 测试平方根运算
func TestSquareRoot(t *testing.T) {
	tests := []struct {
		name     string
		x        float64
		expected float64
	}{
		// 正向用例
		{"完全平方数", 16, 4},
		{"非完全平方数", 2, math.Sqrt(2)},
		// 边界条件
		{"0的平方根", 0, 0},
		{"1的平方根", 1, 1},
		{"大数平方根", 10000, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SquareRoot(tt.x)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("SquareRoot(%f) = %f; 期望 %f", tt.x, result, tt.expected)
			}
		})
	}
}

// TestSquareRoot_Negative 测试负数平方根（应返回NaN）
func TestSquareRoot_Negative(t *testing.T) {
	result := SquareRoot(-1)
	if !math.IsNaN(result) {
		t.Errorf("SquareRoot(-1) 应返回 NaN，实际返回 %f", result)
	}
}

// TestFactorial 测试阶乘运算
func TestFactorial(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		expected int
	}{
		// 正向用例
		{"5的阶乘", 5, 120},
		{"3的阶乘", 3, 6},
		{"10的阶乘", 10, 3628800},
		// 边界条件
		{"0的阶乘", 0, 1},
		{"1的阶乘", 1, 1},
		{"负数阶乘", -5, 1}, // 根据实现，负数返回1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Factorial(tt.n)
			if result != tt.expected {
				t.Errorf("Factorial(%d) = %d; 期望 %d", tt.n, result, tt.expected)
			}
		})
	}
}

// TestGCD 测试最大公约数
func TestGCD(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		// 正向用例
		{"互质数", 7, 11, 1},
		{"有公约数", 12, 18, 6},
		{"相同数", 15, 15, 15},
		// 边界条件
		{"其中一个为0", 10, 0, 10},
		{"大数GCD", 48, 180, 12},
		{"倍数关系", 10, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GCD(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("GCD(%d, %d) = %d; 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestLCM 测试最小公倍数
func TestLCM(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		// 正向用例
		{"互质数", 3, 5, 15},
		{"有公约数", 4, 6, 12},
		{"相同数", 7, 7, 7},
		// 边界条件
		{"倍数关系", 3, 9, 9},
		{"大数LCM", 12, 18, 36},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LCM(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("LCM(%d, %d) = %d; 期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkAdd 加法基准测试
func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(100, 200)
	}
}

// BenchmarkFactorial 阶乘基准测试
func BenchmarkFactorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Factorial(10)
	}
}

// BenchmarkGCD 最大公约数基准测试
func BenchmarkGCD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GCD(48, 180)
	}
}
