package main

import (
	"reflect"
	"testing"
)

// ====================
// 1. equalSlices 测试
// ====================

func TestEqualSlices(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected bool
	}{
		{
			name:     "相等的切片",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "不相等的切片-不同值",
			a:        []int{1, 2, 3},
			b:        []int{1, 2, 4},
			expected: false,
		},
		{
			name:     "不相等的切片-不同长度",
			a:        []int{1, 2, 3},
			b:        []int{1, 2},
			expected: false,
		},
		{
			name:     "空切片相等",
			a:        []int{},
			b:        []int{},
			expected: true,
		},
		{
			name:     "nil切片与空切片",
			a:        nil,
			b:        []int{},
			expected: true,
		},
		{
			name:     "单元素相等",
			a:        []int{42},
			b:        []int{42},
			expected: true,
		},
		{
			name:     "单元素不相等",
			a:        []int{42},
			b:        []int{43},
			expected: false,
		},
		{
			name:     "包含负数",
			a:        []int{-1, -2, -3},
			b:        []int{-1, -2, -3},
			expected: true,
		},
		{
			name:     "包含零",
			a:        []int{0, 0, 0},
			b:        []int{0, 0, 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := equalSlices(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("equalSlices(%v, %v) = %t, want %t",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ====================
// 2. sumSlice 测试
// ====================

func TestSumSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		expected int
	}{
		{
			name:     "正数切片",
			slice:    []int{1, 2, 3, 4, 5},
			expected: 15,
		},
		{
			name:     "空切片",
			slice:    []int{},
			expected: 0,
		},
		{
			name:     "单元素",
			slice:    []int{42},
			expected: 42,
		},
		{
			name:     "负数切片",
			slice:    []int{-1, -2, -3},
			expected: -6,
		},
		{
			name:     "混合正负数",
			slice:    []int{10, -5, 3, -2},
			expected: 6,
		},
		{
			name:     "包含零",
			slice:    []int{0, 1, 0, 2, 0},
			expected: 3,
		},
		{
			name:     "大数值",
			slice:    []int{1000000, 2000000, 3000000},
			expected: 6000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sumSlice(tt.slice)
			if result != tt.expected {
				t.Errorf("sumSlice(%v) = %d, want %d",
					tt.slice, result, tt.expected)
			}
		})
	}
}

// ====================
// 3. modifySliceElements 测试
// ====================

func TestModifySliceElements(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正数翻倍",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{2, 4, 6, 8, 10},
		},
		{
			name:     "空切片",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素",
			input:    []int{7},
			expected: []int{14},
		},
		{
			name:     "包含零",
			input:    []int{0, 1, 0},
			expected: []int{0, 2, 0},
		},
		{
			name:     "负数翻倍",
			input:    []int{-1, -2, -3},
			expected: []int{-2, -4, -6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 复制输入以避免修改测试数据
			slice := make([]int, len(tt.input))
			copy(slice, tt.input)

			modifySliceElements(slice)

			if !reflect.DeepEqual(slice, tt.expected) {
				t.Errorf("modifySliceElements(%v) = %v, want %v",
					tt.input, slice, tt.expected)
			}
		})
	}
}

// ====================
// 4. doubleSlice 测试
// ====================

func TestDoubleSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "正数切片",
			input:    []int{1, 2, 3},
			expected: []int{2, 4, 6},
		},
		{
			name:     "空切片",
			input:    []int{},
			expected: []int{},
		},
		{
			name:     "单元素",
			input:    []int{5},
			expected: []int{10},
		},
		{
			name:     "负数切片",
			input:    []int{-1, -2},
			expected: []int{-2, -4},
		},
		{
			name:     "混合正负零",
			input:    []int{-1, 0, 1},
			expected: []int{-2, 0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始输入
			original := make([]int, len(tt.input))
			copy(original, tt.input)

			result := doubleSlice(tt.input)

			// 验证结果
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("doubleSlice(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}

			// 验证原切片未被修改
			if !reflect.DeepEqual(tt.input, original) {
				t.Errorf("doubleSlice modified original slice: %v -> %v",
					original, tt.input)
			}
		})
	}
}

// ====================
// 5. filterEven 测试
// ====================

func TestFilterEven(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "混合奇偶数",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expected: []int{2, 4, 6, 8, 10},
		},
		{
			name:     "全部偶数",
			input:    []int{2, 4, 6, 8},
			expected: []int{2, 4, 6, 8},
		},
		{
			name:     "全部奇数",
			input:    []int{1, 3, 5, 7},
			expected: nil,
		},
		{
			name:     "空切片",
			input:    []int{},
			expected: nil,
		},
		{
			name:     "包含零",
			input:    []int{0, 1, 2},
			expected: []int{0, 2},
		},
		{
			name:     "负偶数",
			input:    []int{-4, -3, -2, -1, 0},
			expected: []int{-4, -2, 0},
		},
		{
			name:     "单个偶数",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "单个奇数",
			input:    []int{41},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterEven(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("filterEven(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

// ====================
// 6. stringSliceToIntSlice 测试
// ====================

func TestStringSliceToIntSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []int
	}{
		{
			name:     "单数字字符串",
			input:    []string{"1", "2", "3", "4", "5"},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "空切片",
			input:    []string{},
			expected: []int{},
		},
		{
			name:     "单元素",
			input:    []string{"9"},
			expected: []int{9},
		},
		{
			name:     "包含零",
			input:    []string{"0", "1", "0"},
			expected: []int{0, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSliceToIntSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("stringSliceToIntSlice(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

// ====================
// 7. removeDuplicates 测试
// ====================

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "有重复元素",
			input:    []int{1, 2, 2, 3, 3, 3, 4, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "无重复元素",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "全部相同",
			input:    []int{5, 5, 5, 5, 5},
			expected: []int{5},
		},
		{
			name:     "空切片",
			input:    []int{},
			expected: nil,
		},
		{
			name:     "单元素",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "两个相同元素",
			input:    []int{1, 1},
			expected: []int{1},
		},
		{
			name:     "包含负数",
			input:    []int{-1, -1, 0, 0, 1, 1},
			expected: []int{-1, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("removeDuplicates(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

// ====================
// 8. reverseSlice 测试
// ====================

func TestReverseSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "基本反转",
			input:    []string{"a", "b", "c", "d", "e"},
			expected: []string{"e", "d", "c", "b", "a"},
		},
		{
			name:     "空切片",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "单元素",
			input:    []string{"only"},
			expected: []string{"only"},
		},
		{
			name:     "两元素",
			input:    []string{"first", "second"},
			expected: []string{"second", "first"},
		},
		{
			name:     "回文",
			input:    []string{"a", "b", "a"},
			expected: []string{"a", "b", "a"},
		},
		{
			name:     "中文字符",
			input:    []string{"你", "好", "世", "界"},
			expected: []string{"界", "世", "好", "你"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始输入
			original := make([]string, len(tt.input))
			copy(original, tt.input)

			result := reverseSlice(tt.input)

			// 验证结果
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("reverseSlice(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}

			// 验证原切片未被修改
			if !reflect.DeepEqual(tt.input, original) {
				t.Errorf("reverseSlice modified original slice: %v -> %v",
					original, tt.input)
			}
		})
	}
}

// ====================
// 9. 基准测试
// ====================

func BenchmarkSumSlice(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumSlice(slice)
	}
}

func BenchmarkFilterEven(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterEven(slice)
	}
}

func BenchmarkRemoveDuplicates(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i % 100 // 创建重复元素
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		removeDuplicates(slice)
	}
}

func BenchmarkReverseSlice(b *testing.B) {
	slice := make([]string, 1000)
	for i := range slice {
		slice[i] = string(rune('a' + (i % 26)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reverseSlice(slice)
	}
}

func BenchmarkEqualSlices(b *testing.B) {
	slice1 := make([]int, 1000)
	slice2 := make([]int, 1000)
	for i := range slice1 {
		slice1[i] = i
		slice2[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		equalSlices(slice1, slice2)
	}
}

func BenchmarkDoubleSlice(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doubleSlice(slice)
	}
}

// ====================
// 10. 边界条件测试
// ====================

func TestSliceEdgeCases(t *testing.T) {
	t.Run("大切片求和", func(t *testing.T) {
		largeSlice := make([]int, 10000)
		expectedSum := 0
		for i := range largeSlice {
			largeSlice[i] = i
			expectedSum += i
		}

		result := sumSlice(largeSlice)
		if result != expectedSum {
			t.Errorf("sumSlice(large) = %d, want %d", result, expectedSum)
		}
	})

	t.Run("大切片去重", func(t *testing.T) {
		largeSlice := make([]int, 10000)
		for i := range largeSlice {
			largeSlice[i] = i % 100 // 只有100个唯一值
		}

		result := removeDuplicates(largeSlice)
		if len(result) != 100 {
			t.Errorf("removeDuplicates(large) length = %d, want 100", len(result))
		}
	})

	t.Run("双重反转恢复原值", func(t *testing.T) {
		original := []string{"a", "b", "c", "d", "e"}
		reversed := reverseSlice(original)
		doubleReversed := reverseSlice(reversed)

		if !reflect.DeepEqual(doubleReversed, original) {
			t.Errorf("double reverse should equal original: %v != %v",
				doubleReversed, original)
		}
	})

	t.Run("过滤后求和", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		evens := filterEven(numbers)
		sum := sumSlice(evens)

		// 2+4+6+8+10 = 30
		if sum != 30 {
			t.Errorf("sum of evens = %d, want 30", sum)
		}
	})
}

// ====================
// 11. 并发安全测试
// ====================

func TestSliceConcurrency(t *testing.T) {
	t.Run("并发读取切片", func(t *testing.T) {
		slice := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				_ = sumSlice(slice)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// ====================
// 12. 属性测试
// ====================

func TestSliceProperties(t *testing.T) {
	t.Run("去重后长度不大于原长度", func(t *testing.T) {
		testCases := [][]int{
			{1, 2, 3, 4, 5},
			{1, 1, 1, 1, 1},
			{1, 2, 2, 3, 3, 3},
			{},
		}

		for _, tc := range testCases {
			result := removeDuplicates(tc)
			if len(result) > len(tc) {
				t.Errorf("removeDuplicates result length %d > original length %d",
					len(result), len(tc))
			}
		}
	})

	t.Run("翻倍后每个元素是原来的两倍", func(t *testing.T) {
		original := []int{1, 2, 3, 4, 5}
		doubled := doubleSlice(original)

		for i := range original {
			if doubled[i] != original[i]*2 {
				t.Errorf("doubled[%d] = %d, want %d",
					i, doubled[i], original[i]*2)
			}
		}
	})

	t.Run("过滤偶数后全部是偶数", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		evens := filterEven(numbers)

		for _, n := range evens {
			if n%2 != 0 {
				t.Errorf("filterEven result contains odd number: %d", n)
			}
		}
	})

	t.Run("反转保持长度不变", func(t *testing.T) {
		original := []string{"a", "b", "c", "d", "e"}
		reversed := reverseSlice(original)

		if len(reversed) != len(original) {
			t.Errorf("reverseSlice changed length: %d -> %d",
				len(original), len(reversed))
		}
	})
}
