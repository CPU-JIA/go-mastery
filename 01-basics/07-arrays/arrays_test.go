package main

import (
	"reflect"
	"testing"
)

// ====================
// 1. sumArray 测试
// ====================

func TestSumArray(t *testing.T) {
	tests := []struct {
		name     string
		arr      [5]int
		expected int
	}{
		{
			name:     "正数数组",
			arr:      [5]int{1, 2, 3, 4, 5},
			expected: 15,
		},
		{
			name:     "零值数组",
			arr:      [5]int{0, 0, 0, 0, 0},
			expected: 0,
		},
		{
			name:     "负数数组",
			arr:      [5]int{-1, -2, -3, -4, -5},
			expected: -15,
		},
		{
			name:     "混合正负数",
			arr:      [5]int{10, -5, 3, -2, 4},
			expected: 10,
		},
		{
			name:     "包含零",
			arr:      [5]int{1, 0, 2, 0, 3},
			expected: 6,
		},
		{
			name:     "大数值",
			arr:      [5]int{100000, 200000, 300000, 400000, 500000},
			expected: 1500000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sumArray(tt.arr)
			if result != tt.expected {
				t.Errorf("sumArray(%v) = %d, want %d",
					tt.arr, result, tt.expected)
			}
		})
	}
}

// ====================
// 2. doubleArray 测试
// ====================

func TestDoubleArray(t *testing.T) {
	tests := []struct {
		name     string
		arr      [5]int
		expected [5]int
	}{
		{
			name:     "正数数组",
			arr:      [5]int{1, 2, 3, 4, 5},
			expected: [5]int{2, 4, 6, 8, 10},
		},
		{
			name:     "零值数组",
			arr:      [5]int{0, 0, 0, 0, 0},
			expected: [5]int{0, 0, 0, 0, 0},
		},
		{
			name:     "负数数组",
			arr:      [5]int{-1, -2, -3, -4, -5},
			expected: [5]int{-2, -4, -6, -8, -10},
		},
		{
			name:     "混合正负零",
			arr:      [5]int{-1, 0, 1, 0, 2},
			expected: [5]int{-2, 0, 2, 0, 4},
		},
		{
			name:     "单个非零元素",
			arr:      [5]int{0, 0, 5, 0, 0},
			expected: [5]int{0, 0, 10, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始数组
			original := tt.arr

			result := doubleArray(tt.arr)

			// 验证结果
			if result != tt.expected {
				t.Errorf("doubleArray(%v) = %v, want %v",
					tt.arr, result, tt.expected)
			}

			// 验证原数组未被修改（值传递）
			if tt.arr != original {
				t.Errorf("doubleArray modified original array: %v -> %v",
					original, tt.arr)
			}
		})
	}
}

// ====================
// 3. bubbleSort 测试
// ====================

func TestBubbleSort(t *testing.T) {
	tests := []struct {
		name     string
		arr      [6]int
		expected [6]int
	}{
		{
			name:     "逆序数组",
			arr:      [6]int{6, 5, 4, 3, 2, 1},
			expected: [6]int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "已排序数组",
			arr:      [6]int{1, 2, 3, 4, 5, 6},
			expected: [6]int{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "随机顺序",
			arr:      [6]int{64, 34, 25, 12, 22, 11},
			expected: [6]int{11, 12, 22, 25, 34, 64},
		},
		{
			name:     "包含重复元素",
			arr:      [6]int{3, 1, 4, 1, 5, 9},
			expected: [6]int{1, 1, 3, 4, 5, 9},
		},
		{
			name:     "全部相同",
			arr:      [6]int{5, 5, 5, 5, 5, 5},
			expected: [6]int{5, 5, 5, 5, 5, 5},
		},
		{
			name:     "包含负数",
			arr:      [6]int{-3, 5, -1, 0, 2, -4},
			expected: [6]int{-4, -3, -1, 0, 2, 5},
		},
		{
			name:     "包含零",
			arr:      [6]int{0, 0, 1, 0, 2, 0},
			expected: [6]int{0, 0, 0, 0, 1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始数组
			original := tt.arr

			result := bubbleSort(tt.arr)

			// 验证结果
			if result != tt.expected {
				t.Errorf("bubbleSort(%v) = %v, want %v",
					tt.arr, result, tt.expected)
			}

			// 验证原数组未被修改（值传递）
			if tt.arr != original {
				t.Errorf("bubbleSort modified original array: %v -> %v",
					original, tt.arr)
			}
		})
	}
}

// ====================
// 4. linearSearch 测试
// ====================

func TestLinearSearch(t *testing.T) {
	tests := []struct {
		name     string
		arr      [8]int
		target   int
		expected int
	}{
		{
			name:     "找到目标-中间位置",
			arr:      [8]int{2, 7, 11, 15, 23, 31, 45, 67},
			target:   23,
			expected: 4,
		},
		{
			name:     "找到目标-第一个位置",
			arr:      [8]int{2, 7, 11, 15, 23, 31, 45, 67},
			target:   2,
			expected: 0,
		},
		{
			name:     "找到目标-最后位置",
			arr:      [8]int{2, 7, 11, 15, 23, 31, 45, 67},
			target:   67,
			expected: 7,
		},
		{
			name:     "未找到目标",
			arr:      [8]int{2, 7, 11, 15, 23, 31, 45, 67},
			target:   100,
			expected: -1,
		},
		{
			name:     "未找到-小于最小值",
			arr:      [8]int{2, 7, 11, 15, 23, 31, 45, 67},
			target:   1,
			expected: -1,
		},
		{
			name:     "查找零",
			arr:      [8]int{0, 1, 2, 3, 4, 5, 6, 7},
			target:   0,
			expected: 0,
		},
		{
			name:     "查找负数",
			arr:      [8]int{-5, -3, -1, 0, 1, 3, 5, 7},
			target:   -3,
			expected: 1,
		},
		{
			name:     "重复元素-返回第一个",
			arr:      [8]int{1, 2, 3, 3, 3, 4, 5, 6},
			target:   3,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := linearSearch(tt.arr, tt.target)
			if result != tt.expected {
				t.Errorf("linearSearch(%v, %d) = %d, want %d",
					tt.arr, tt.target, result, tt.expected)
			}
		})
	}
}

// ====================
// 5. printMatrix 测试（通过验证不panic）
// ====================

func TestPrintMatrix(t *testing.T) {
	tests := []struct {
		name   string
		matrix [3][3]int
	}{
		{
			name: "正常矩阵",
			matrix: [3][3]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
		},
		{
			name: "零矩阵",
			matrix: [3][3]int{
				{0, 0, 0},
				{0, 0, 0},
				{0, 0, 0},
			},
		},
		{
			name: "单位矩阵",
			matrix: [3][3]int{
				{1, 0, 0},
				{0, 1, 0},
				{0, 0, 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 只验证不会panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printMatrix panicked: %v", r)
				}
			}()
			printMatrix(tt.matrix)
		})
	}
}

// ====================
// 6. modifyArrayByPointer 测试
// ====================

func TestModifyArrayByPointer(t *testing.T) {
	tests := []struct {
		name     string
		arr      [5]int
		expected [5]int
	}{
		{
			name:     "修改第一个元素",
			arr:      [5]int{1, 2, 3, 4, 5},
			expected: [5]int{999, 2, 3, 4, 5},
		},
		{
			name:     "零值数组",
			arr:      [5]int{0, 0, 0, 0, 0},
			expected: [5]int{999, 0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arr := tt.arr
			modifyArrayByPointer(&arr)

			if arr != tt.expected {
				t.Errorf("modifyArrayByPointer result = %v, want %v",
					arr, tt.expected)
			}
		})
	}
}

// ====================
// 7. modifyArray 测试（验证值传递不修改原数组）
// ====================

func TestModifyArray(t *testing.T) {
	original := [5]int{1, 2, 3, 4, 5}
	arr := original

	modifyArray(arr)

	// 验证原数组未被修改
	if arr != original {
		t.Errorf("modifyArray should not modify original array: %v -> %v",
			original, arr)
	}
}

// ====================
// 8. 基准测试
// ====================

func BenchmarkSumArray(b *testing.B) {
	arr := [5]int{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sumArray(arr)
	}
}

func BenchmarkDoubleArray(b *testing.B) {
	arr := [5]int{1, 2, 3, 4, 5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doubleArray(arr)
	}
}

func BenchmarkBubbleSort(b *testing.B) {
	arr := [6]int{64, 34, 25, 12, 22, 11}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bubbleSort(arr)
	}
}

func BenchmarkBubbleSortWorstCase(b *testing.B) {
	// 最坏情况：逆序数组
	arr := [6]int{6, 5, 4, 3, 2, 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bubbleSort(arr)
	}
}

func BenchmarkBubbleSortBestCase(b *testing.B) {
	// 最好情况：已排序数组
	arr := [6]int{1, 2, 3, 4, 5, 6}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bubbleSort(arr)
	}
}

func BenchmarkLinearSearch(b *testing.B) {
	arr := [8]int{2, 7, 11, 15, 23, 31, 45, 67}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		linearSearch(arr, 23)
	}
}

func BenchmarkLinearSearchWorstCase(b *testing.B) {
	// 最坏情况：目标在最后或不存在
	arr := [8]int{2, 7, 11, 15, 23, 31, 45, 67}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		linearSearch(arr, 100)
	}
}

// ====================
// 9. 边界条件测试
// ====================

func TestArrayEdgeCases(t *testing.T) {
	t.Run("排序稳定性", func(t *testing.T) {
		// 多次排序应该得到相同结果
		arr := [6]int{64, 34, 25, 12, 22, 11}
		result1 := bubbleSort(arr)
		result2 := bubbleSort(arr)

		if result1 != result2 {
			t.Errorf("bubbleSort not stable: %v != %v", result1, result2)
		}
	})

	t.Run("排序后再排序", func(t *testing.T) {
		arr := [6]int{64, 34, 25, 12, 22, 11}
		sorted := bubbleSort(arr)
		sortedAgain := bubbleSort(sorted)

		if sorted != sortedAgain {
			t.Errorf("sorting sorted array changed result: %v != %v",
				sorted, sortedAgain)
		}
	})

	t.Run("翻倍后求和", func(t *testing.T) {
		arr := [5]int{1, 2, 3, 4, 5}
		doubled := doubleArray(arr)

		originalSum := sumArray(arr)
		doubledSum := sumArray(doubled)

		if doubledSum != originalSum*2 {
			t.Errorf("sum of doubled array should be 2x original: %d != %d*2",
				doubledSum, originalSum)
		}
	})

	t.Run("查找排序后的元素", func(t *testing.T) {
		unsorted := [6]int{64, 34, 25, 12, 22, 11}
		sorted := bubbleSort(unsorted)

		// 转换为[8]int进行查找
		arr := [8]int{sorted[0], sorted[1], sorted[2], sorted[3], sorted[4], sorted[5], 0, 0}

		// 查找最小值（应该在索引0）
		idx := linearSearch(arr, sorted[0])
		if idx != 0 {
			t.Errorf("minimum element should be at index 0, got %d", idx)
		}
	})
}

// ====================
// 10. 属性测试
// ====================

func TestArrayProperties(t *testing.T) {
	t.Run("排序后数组有序", func(t *testing.T) {
		testCases := [][6]int{
			{6, 5, 4, 3, 2, 1},
			{1, 2, 3, 4, 5, 6},
			{3, 1, 4, 1, 5, 9},
			{-3, 5, -1, 0, 2, -4},
		}

		for _, tc := range testCases {
			sorted := bubbleSort(tc)
			for i := 0; i < len(sorted)-1; i++ {
				if sorted[i] > sorted[i+1] {
					t.Errorf("bubbleSort result not sorted: %v", sorted)
					break
				}
			}
		}
	})

	t.Run("排序保持元素", func(t *testing.T) {
		arr := [6]int{64, 34, 25, 12, 22, 11}
		sorted := bubbleSort(arr)

		// 计算排序前后的和（应该相等）
		sumBefore := 0
		sumAfter := 0
		for i := 0; i < 6; i++ {
			sumBefore += arr[i]
			sumAfter += sorted[i]
		}

		if sumBefore != sumAfter {
			t.Errorf("bubbleSort changed sum: %d -> %d", sumBefore, sumAfter)
		}
	})

	t.Run("翻倍每个元素是原来的两倍", func(t *testing.T) {
		arr := [5]int{1, 2, 3, 4, 5}
		doubled := doubleArray(arr)

		for i := 0; i < 5; i++ {
			if doubled[i] != arr[i]*2 {
				t.Errorf("doubled[%d] = %d, want %d",
					i, doubled[i], arr[i]*2)
			}
		}
	})

	t.Run("查找存在的元素返回有效索引", func(t *testing.T) {
		arr := [8]int{2, 7, 11, 15, 23, 31, 45, 67}

		for i, v := range arr {
			idx := linearSearch(arr, v)
			if idx < 0 || idx >= len(arr) {
				t.Errorf("linearSearch(%d) returned invalid index: %d", v, idx)
			}
			if arr[idx] != v {
				t.Errorf("linearSearch(%d) returned wrong index: arr[%d] = %d",
					v, idx, arr[idx])
			}
			// 对于非重复元素，索引应该匹配
			if idx != i {
				// 检查是否是重复元素
				found := false
				for j := 0; j < i; j++ {
					if arr[j] == v {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("linearSearch(%d) returned %d, want %d", v, idx, i)
				}
			}
		}
	})
}

// ====================
// 11. 数组比较测试
// ====================

func TestArrayComparison(t *testing.T) {
	t.Run("相等数组", func(t *testing.T) {
		arr1 := [5]int{1, 2, 3, 4, 5}
		arr2 := [5]int{1, 2, 3, 4, 5}

		if arr1 != arr2 {
			t.Errorf("equal arrays should be equal: %v != %v", arr1, arr2)
		}
	})

	t.Run("不相等数组", func(t *testing.T) {
		arr1 := [5]int{1, 2, 3, 4, 5}
		arr2 := [5]int{1, 2, 3, 4, 6}

		if arr1 == arr2 {
			t.Errorf("different arrays should not be equal: %v == %v", arr1, arr2)
		}
	})

	t.Run("数组复制独立性", func(t *testing.T) {
		arr1 := [5]int{1, 2, 3, 4, 5}
		arr2 := arr1 // 值复制

		arr2[0] = 999

		if arr1[0] == 999 {
			t.Error("array copy should be independent")
		}
	})
}

// ====================
// 12. 表驱动测试辅助
// ====================

func TestArrayOperationsIntegration(t *testing.T) {
	// 集成测试：组合多个操作
	t.Run("求和-翻倍-再求和", func(t *testing.T) {
		arr := [5]int{1, 2, 3, 4, 5}

		sum1 := sumArray(arr)
		doubled := doubleArray(arr)
		sum2 := sumArray(doubled)

		if sum2 != sum1*2 {
			t.Errorf("doubled sum should be 2x: %d != %d*2", sum2, sum1)
		}
	})

	t.Run("排序-查找", func(t *testing.T) {
		unsorted := [6]int{64, 34, 25, 12, 22, 11}
		sorted := bubbleSort(unsorted)

		// 最小值应该在第一个位置
		minVal := sorted[0]
		for _, v := range sorted {
			if v < minVal {
				t.Errorf("sorted array minimum not at index 0: %v", sorted)
				break
			}
		}
	})
}

// ====================
// 13. 使用 reflect 的深度比较测试
// ====================

func TestArrayDeepEqual(t *testing.T) {
	t.Run("使用reflect比较数组", func(t *testing.T) {
		arr1 := [5]int{1, 2, 3, 4, 5}
		arr2 := [5]int{1, 2, 3, 4, 5}

		if !reflect.DeepEqual(arr1, arr2) {
			t.Errorf("reflect.DeepEqual failed for equal arrays")
		}
	})

	t.Run("使用reflect比较不同数组", func(t *testing.T) {
		arr1 := [5]int{1, 2, 3, 4, 5}
		arr2 := [5]int{5, 4, 3, 2, 1}

		if reflect.DeepEqual(arr1, arr2) {
			t.Errorf("reflect.DeepEqual should return false for different arrays")
		}
	})
}
