package main

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

// ====================
// 1. 基本函数测试
// ====================

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 5, 3, 8},
		{"negative numbers", -5, -3, -8},
		{"mixed signs", 5, -3, 2},
		{"zero values", 0, 0, 0},
		{"large numbers", 1000000, 2000000, 3000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("add(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMultiply(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 5, 3, 15},
		{"negative numbers", -5, -3, 15},
		{"mixed signs", 5, -3, -15},
		{"zero multiplication", 5, 0, 0},
		{"one multiplication", 7, 1, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("multiply(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive result", 10, 3, 7},
		{"negative result", 3, 10, -7},
		{"zero result", 5, 5, 0},
		{"negative numbers", -5, -3, -2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subtract(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("subtract(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	tests := []struct {
		name              string
		a, b              int
		expectedQuotient  int
		expectedRemainder int
	}{
		{"basic division", 17, 5, 3, 2},
		{"exact division", 15, 3, 5, 0},
		{"zero remainder", 10, 2, 5, 0},
		{"dividend smaller", 3, 7, 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quotient, remainder := divide(tt.a, tt.b)
			if quotient != tt.expectedQuotient {
				t.Errorf("divide(%d, %d) quotient = %d, want %d", tt.a, tt.b, quotient, tt.expectedQuotient)
			}
			if remainder != tt.expectedRemainder {
				t.Errorf("divide(%d, %d) remainder = %d, want %d", tt.a, tt.b, remainder, tt.expectedRemainder)
			}
		})
	}
}

func TestCalculateCircleArea(t *testing.T) {
	tests := []struct {
		name     string
		radius   float64
		expected float64
		delta    float64
	}{
		{"unit circle", 1.0, math.Pi, 0.001},
		{"zero radius", 0.0, 0.0, 0.001},
		{"small radius", 0.5, math.Pi * 0.25, 0.001},
		{"large radius", 10.0, math.Pi * 100, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCircleArea(tt.radius)
			if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("calculateCircleArea(%f) = %f, want %f (±%f)", tt.radius, result, tt.expected, tt.delta)
			}
		})
	}
}

// ====================
// 2. 布尔函数测试
// ====================

func TestIsEven(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected bool
	}{
		{"positive even", 4, true},
		{"positive odd", 5, false},
		{"negative even", -4, true},
		{"negative odd", -5, false},
		{"zero", 0, true},
		{"large even", 1000000, true},
		{"large odd", 1000001, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEven(tt.number)
			if result != tt.expected {
				t.Errorf("isEven(%d) = %t, want %t", tt.number, result, tt.expected)
			}
		})
	}
}

func TestIsPositive(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected bool
	}{
		{"positive number", 5, true},
		{"negative number", -5, false},
		{"zero", 0, false},
		{"large positive", 1000000, true},
		{"large negative", -1000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPositive(tt.number)
			if result != tt.expected {
				t.Errorf("isPositive(%d) = %t, want %t", tt.number, result, tt.expected)
			}
		})
	}
}

// ====================
// 3. 字符串函数测试
// ====================

func TestReverseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "olleh"},
		{"empty string", "", ""},
		{"single character", "a", "a"},
		{"palindrome", "racecar", "racecar"},
		{"unicode string", "你好世界", "界世好你"},
		{"mixed content", "Hello123", "321olleH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseString(tt.input)
			if result != tt.expected {
				t.Errorf("reverseString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		parts     []string
		expected  string
	}{
		{"basic join", "-", []string{"apple", "banana", "cherry"}, "apple-banana-cherry"},
		{"empty separator", "", []string{"a", "b", "c"}, "abc"},
		{"empty parts", "-", []string{}, ""},
		{"single part", "-", []string{"apple"}, "apple"},
		{"empty strings", "-", []string{"", "", ""}, "--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := join(tt.separator, tt.parts...)
			if result != tt.expected {
				t.Errorf("join(%q, %v) = %q, want %q", tt.separator, tt.parts, result, tt.expected)
			}
		})
	}
}

func TestSplitAndCount(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedWords []string
		expectedCount int
	}{
		{"basic sentence", "hello world go programming", []string{"hello", "world", "go", "programming"}, 4},
		{"single word", "hello", []string{"hello"}, 1},
		{"empty string", "", []string{}, 0},
		{"multiple spaces", "hello    world", []string{"hello", "world"}, 2},
		{"leading/trailing spaces", "  hello world  ", []string{"hello", "world"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words, count := splitAndCount(tt.input)
			if !reflect.DeepEqual(words, tt.expectedWords) {
				t.Errorf("splitAndCount(%q) words = %v, want %v", tt.input, words, tt.expectedWords)
			}
			if count != tt.expectedCount {
				t.Errorf("splitAndCount(%q) count = %d, want %d", tt.input, count, tt.expectedCount)
			}
		})
	}
}

// ====================
// 4. 集合函数测试
// ====================

func TestFindMinMax(t *testing.T) {
	tests := []struct {
		name        string
		numbers     []int
		expectedMin int
		expectedMax int
	}{
		{"basic array", []int{3, 7, 1, 9, 4}, 1, 9},
		{"single element", []int{5}, 5, 5},
		{"all same", []int{5, 5, 5}, 5, 5},
		{"negative numbers", []int{-3, -7, -1, -9}, -9, -1},
		{"mixed signs", []int{-5, 0, 10, -2}, -5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := findMinMax(tt.numbers)
			if min != tt.expectedMin {
				t.Errorf("findMinMax(%v) min = %d, want %d", tt.numbers, min, tt.expectedMin)
			}
			if max != tt.expectedMax {
				t.Errorf("findMinMax(%v) max = %d, want %d", tt.numbers, max, tt.expectedMax)
			}
		})
	}
}

func TestFindMinMaxEmpty(t *testing.T) {
	min, max := findMinMax([]int{})
	if min != 0 || max != 0 {
		t.Errorf("findMinMax([]) = (%d, %d), want (0, 0)", min, max)
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		numbers  []int
		expected int
	}{
		{"multiple numbers", []int{1, 2, 3, 4, 5}, 15},
		{"empty array", []int{}, 0},
		{"single number", []int{10}, 10},
		{"negative numbers", []int{-1, -2, -3}, -6},
		{"mixed signs", []int{10, -5, 3}, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sum(tt.numbers...)
			if result != tt.expected {
				t.Errorf("sum(%v) = %d, want %d", tt.numbers, result, tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name     string
		numbers  []int
		expected int
	}{
		{"multiple numbers", []int{3, 7, 1, 9, 4}, 9},
		{"single number", []int{5}, 5},
		{"all same", []int{5, 5, 5}, 5},
		{"negative numbers", []int{-3, -7, -1}, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := max(tt.numbers...)
			if result != tt.expected {
				t.Errorf("max(%v) = %d, want %d", tt.numbers, result, tt.expected)
			}
		})
	}
}

func TestMaxEmpty(t *testing.T) {
	result := max()
	if result != 0 {
		t.Errorf("max() = %d, want 0", result)
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		numbers  []int
		expected int
	}{
		{"multiple numbers", []int{3, 7, 1, 9, 4}, 1},
		{"single number", []int{5}, 5},
		{"all same", []int{5, 5, 5}, 5},
		{"negative numbers", []int{-3, -7, -1}, -7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.numbers...)
			if result != tt.expected {
				t.Errorf("min(%v) = %d, want %d", tt.numbers, result, tt.expected)
			}
		})
	}
}

// ====================
// 5. 错误处理测试
// ====================

func TestSafeDivide(t *testing.T) {
	tests := []struct {
		name        string
		a, b        float64
		expected    float64
		expectError bool
	}{
		{"valid division", 10.0, 2.0, 5.0, false},
		{"division by zero", 10.0, 0.0, 0.0, true},
		{"negative numbers", -10.0, 2.0, -5.0, false},
		{"decimal result", 7.0, 2.0, 3.5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := safeDivide(tt.a, tt.b)

			if tt.expectError {
				if err == nil {
					t.Errorf("safeDivide(%f, %f) expected error, got none", tt.a, tt.b)
				}
			} else {
				if err != nil {
					t.Errorf("safeDivide(%f, %f) unexpected error: %v", tt.a, tt.b, err)
				}
				if math.Abs(result-tt.expected) > 0.001 {
					t.Errorf("safeDivide(%f, %f) = %f, want %f", tt.a, tt.b, result, tt.expected)
				}
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		expectedMsg string
	}{
		{"valid txt file", "config.txt", false, "文件内容示例"},
		{"empty filename", "", true, ""},
		{"unsupported type", "config.json", true, ""},
		{"another txt file", "data.txt", false, "文件内容示例"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := readFile(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("readFile(%q) expected error, got none", tt.filename)
				}
			} else {
				if err != nil {
					t.Errorf("readFile(%q) unexpected error: %v", tt.filename, err)
				}
				if content != tt.expectedMsg {
					t.Errorf("readFile(%q) = %q, want %q", tt.filename, content, tt.expectedMsg)
				}
			}
		})
	}
}

// ====================
// 6. 命名返回值测试
// ====================

func TestRectangleAreaAndPerimeter(t *testing.T) {
	tests := []struct {
		name              string
		length, width     int
		expectedArea      int
		expectedPerimeter int
	}{
		{"basic rectangle", 5, 3, 15, 16},
		{"square", 4, 4, 16, 16},
		{"zero width", 5, 0, 0, 10},
		{"zero length", 0, 3, 0, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area, perimeter := rectangleAreaAndPerimeter(tt.length, tt.width)
			if area != tt.expectedArea {
				t.Errorf("rectangleAreaAndPerimeter(%d, %d) area = %d, want %d", tt.length, tt.width, area, tt.expectedArea)
			}
			if perimeter != tt.expectedPerimeter {
				t.Errorf("rectangleAreaAndPerimeter(%d, %d) perimeter = %d, want %d", tt.length, tt.width, perimeter, tt.expectedPerimeter)
			}
		})
	}
}

func TestStatistics(t *testing.T) {
	tests := []struct {
		name         string
		data         []float64
		expectedMean float64
		delta        float64
	}{
		{"basic data", []float64{1, 2, 3, 4, 5}, 3.0, 0.001},
		{"empty data", []float64{}, 0.0, 0.001},
		{"single value", []float64{5.0}, 5.0, 0.001},
		{"negative values", []float64{-1, 0, 1}, 0.0, 0.001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean, variance, stdDev := statistics(tt.data)
			if math.Abs(mean-tt.expectedMean) > tt.delta {
				t.Errorf("statistics(%v) mean = %f, want %f", tt.data, mean, tt.expectedMean)
			}
			// 验证方差和标准差的关系
			if len(tt.data) > 0 && math.Abs(stdDev-math.Sqrt(variance)) > tt.delta {
				t.Errorf("statistics(%v) stdDev = %f, but sqrt(variance) = %f", tt.data, stdDev, math.Sqrt(variance))
			}
		})
	}
}

func TestCalculatePrice(t *testing.T) {
	tests := []struct {
		name             string
		originalPrice    float64
		customerType     string
		expectedDiscount float64
		expectedSavings  float64
	}{
		{"VIP customer", 1000.0, "VIP", 0.2, 200.0},
		{"Member customer", 1000.0, "Member", 0.1, 100.0},
		{"Regular customer", 1000.0, "Regular", 0.0, 0.0},
		{"Empty type", 1000.0, "", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discount, finalPrice, savings := calculatePrice(tt.originalPrice, tt.customerType)
			if discount != tt.expectedDiscount {
				t.Errorf("calculatePrice(%f, %q) discount = %f, want %f", tt.originalPrice, tt.customerType, discount, tt.expectedDiscount)
			}
			if savings != tt.expectedSavings {
				t.Errorf("calculatePrice(%f, %q) savings = %f, want %f", tt.originalPrice, tt.customerType, savings, tt.expectedSavings)
			}
			expectedFinalPrice := tt.originalPrice - tt.expectedSavings
			if finalPrice != expectedFinalPrice {
				t.Errorf("calculatePrice(%f, %q) finalPrice = %f, want %f", tt.originalPrice, tt.customerType, finalPrice, expectedFinalPrice)
			}
		})
	}
}

// ====================
// 7. 高阶函数测试
// ====================

func TestApplyToSlice(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}

	doubled := applyToSlice(input, double)
	expected := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(doubled, expected) {
		t.Errorf("applyToSlice with double: got %v, want %v", doubled, expected)
	}

	squared := applyToSlice(input, square)
	expected = []int{1, 4, 9, 16, 25}
	if !reflect.DeepEqual(squared, expected) {
		t.Errorf("applyToSlice with square: got %v, want %v", squared, expected)
	}
}

func TestMakeAdder(t *testing.T) {
	add5 := makeAdder(5)
	add10 := makeAdder(10)

	tests := []struct {
		name     string
		fn       func(int) int
		input    int
		expected int
	}{
		{"add 5 to 3", add5, 3, 8},
		{"add 5 to 0", add5, 0, 5},
		{"add 10 to 5", add10, 5, 15},
		{"add 10 to -3", add10, -3, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.input)
			if result != tt.expected {
				t.Errorf("function(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompose(t *testing.T) {
	addOne := func(x int) int { return x + 1 }
	multiplyTwo := func(x int) int { return x * 2 }

	composed := compose(multiplyTwo, addOne)

	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"5 -> 6 -> 12", 5, 12},
		{"0 -> 1 -> 2", 0, 2},
		{"-1 -> 0 -> 0", -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := composed(tt.input)
			if result != tt.expected {
				t.Errorf("composed(%d) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ====================
// 8. 数据验证测试
// ====================

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name     string
		user     map[string]interface{}
		hasError bool
		errors   []string
	}{
		{
			name: "valid user",
			user: map[string]interface{}{
				"name":  "张三",
				"age":   25,
				"email": "zhang@example.com",
			},
			hasError: false,
		},
		{
			name: "empty name",
			user: map[string]interface{}{
				"name":  "",
				"age":   25,
				"email": "zhang@example.com",
			},
			hasError: true,
			errors:   []string{"姓名不能为空"},
		},
		{
			name: "underage",
			user: map[string]interface{}{
				"name":  "张三",
				"age":   15,
				"email": "zhang@example.com",
			},
			hasError: true,
			errors:   []string{"年龄必须大于等于18"},
		},
		{
			name: "invalid email",
			user: map[string]interface{}{
				"name":  "张三",
				"age":   25,
				"email": "invalid-email",
			},
			hasError: true,
			errors:   []string{"邮箱格式无效"},
		},
		{
			name: "multiple errors",
			user: map[string]interface{}{
				"name":  "",
				"age":   15,
				"email": "invalid",
			},
			hasError: true,
			errors:   []string{"姓名不能为空", "年龄必须大于等于18", "邮箱格式无效"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateUser(tt.user)
			if tt.hasError {
				if len(errors) == 0 {
					t.Errorf("validateUser(%v) expected errors, got none", tt.user)
				}
				for _, expectedErr := range tt.errors {
					found := false
					for _, actualErr := range errors {
						if strings.Contains(actualErr, expectedErr) || actualErr == expectedErr {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("validateUser(%v) expected error containing %q, got %v", tt.user, expectedErr, errors)
					}
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("validateUser(%v) expected no errors, got %v", tt.user, errors)
				}
			}
		})
	}
}

// ====================
// 9. 性能基准测试
// ====================

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		add(123, 456)
	}
}

func BenchmarkReverseString(b *testing.B) {
	str := "Hello, World! This is a test string for benchmarking."
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reverseString(str)
	}
}

func BenchmarkFindMinMax(b *testing.B) {
	numbers := []int{3, 7, 1, 9, 4, 2, 8, 5, 6, 10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findMinMax(numbers)
	}
}

func BenchmarkSum(b *testing.B) {
	numbers := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		numbers[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum(numbers...)
	}
}

func BenchmarkMemoizedFibonacci(b *testing.B) {
	fib := memoize(fibonacci)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fib(20)
	}
}

// ====================
// 10. 辅助测试函数
// ====================

func TestDouble(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 10},
		{0, 0},
		{-3, -6},
	}

	for _, tt := range tests {
		result := double(tt.input)
		if result != tt.expected {
			t.Errorf("double(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestSquare(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{5, 25},
		{0, 0},
		{-3, 9},
	}

	for _, tt := range tests {
		result := square(tt.input)
		if result != tt.expected {
			t.Errorf("square(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestLookup(t *testing.T) {
	m := map[string]int{
		"apple":  5,
		"banana": 3,
		"cherry": 8,
	}

	tests := []struct {
		name        string
		key         string
		expectedVal int
		expectedOk  bool
	}{
		{"existing key", "apple", 5, true},
		{"missing key", "orange", 0, false},
		{"another existing key", "banana", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := lookup(m, tt.key)
			if val != tt.expectedVal {
				t.Errorf("lookup(m, %q) value = %d, want %d", tt.key, val, tt.expectedVal)
			}
			if ok != tt.expectedOk {
				t.Errorf("lookup(m, %q) ok = %t, want %t", tt.key, ok, tt.expectedOk)
			}
		})
	}
}

// ====================
// 11. 错误边界测试
// ====================

func TestEdgeCases(t *testing.T) {
	// 测试空切片情况
	t.Run("empty slice operations", func(t *testing.T) {
		result := sum()
		if result != 0 {
			t.Errorf("sum() with no args = %d, want 0", result)
		}

		maxResult := max()
		if maxResult != 0 {
			t.Errorf("max() with no args = %d, want 0", maxResult)
		}
	})

	// 测试大数值
	t.Run("large numbers", func(t *testing.T) {
		large1 := 1000000000
		large2 := 2000000000
		expected := 3000000000
		result := add(large1, large2)
		if result != expected {
			t.Errorf("add(%d, %d) = %d, want %d", large1, large2, result, expected)
		}
	})

	// 测试 Unicode 字符串
	t.Run("unicode handling", func(t *testing.T) {
		unicode := "你好世界🌍"
		reversed := reverseString(unicode)
		doubleReversed := reverseString(reversed)
		if doubleReversed != unicode {
			t.Errorf("double reverse of %q should equal original, got %q", unicode, doubleReversed)
		}
	})
}
