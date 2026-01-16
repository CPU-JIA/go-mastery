package main

import (
	"reflect"
	"sort"
	"testing"
)

// ====================
// 1. mapSetToSlice 测试
// ====================

func TestMapSetToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[int]bool
		expected []int
	}{
		{
			name:     "基本集合",
			input:    map[int]bool{1: true, 2: true, 3: true},
			expected: []int{1, 2, 3},
		},
		{
			name:     "空集合",
			input:    map[int]bool{},
			expected: nil,
		},
		{
			name:     "单元素集合",
			input:    map[int]bool{42: true},
			expected: []int{42},
		},
		{
			name:     "包含负数",
			input:    map[int]bool{-3: true, -1: true, 0: true, 2: true},
			expected: []int{-3, -1, 0, 2},
		},
		{
			name:     "大数值",
			input:    map[int]bool{100: true, 200: true, 300: true},
			expected: []int{100, 200, 300},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapSetToSlice(tt.input)
			// 结果已排序，直接比较
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("mapSetToSlice(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ====================
// 2. calculateAverage 测试
// ====================

func TestCalculateAverage(t *testing.T) {
	tests := []struct {
		name     string
		scores   map[string]int
		expected float64
		delta    float64
	}{
		{
			name:     "正常分数",
			scores:   map[string]int{"Alice": 85, "Bob": 92, "Charlie": 78},
			expected: 85.0,
			delta:    0.001,
		},
		{
			name:     "空映射",
			scores:   map[string]int{},
			expected: 0.0,
			delta:    0.001,
		},
		{
			name:     "单个分数",
			scores:   map[string]int{"Alice": 100},
			expected: 100.0,
			delta:    0.001,
		},
		{
			name:     "相同分数",
			scores:   map[string]int{"A": 80, "B": 80, "C": 80},
			expected: 80.0,
			delta:    0.001,
		},
		{
			name:     "包含零分",
			scores:   map[string]int{"A": 100, "B": 0, "C": 50},
			expected: 50.0,
			delta:    0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAverage(tt.scores)
			if result < tt.expected-tt.delta || result > tt.expected+tt.delta {
				t.Errorf("calculateAverage(%v) = %f, want %f (delta: %f)",
					tt.scores, result, tt.expected, tt.delta)
			}
		})
	}
}

// ====================
// 3. addBonus 测试
// ====================

func TestAddBonus(t *testing.T) {
	tests := []struct {
		name     string
		scores   map[string]int
		bonus    int
		expected map[string]int
	}{
		{
			name:     "正常加分",
			scores:   map[string]int{"Alice": 85, "Bob": 90},
			bonus:    5,
			expected: map[string]int{"Alice": 90, "Bob": 95},
		},
		{
			name:     "零加分",
			scores:   map[string]int{"Alice": 85},
			bonus:    0,
			expected: map[string]int{"Alice": 85},
		},
		{
			name:     "负加分(扣分)",
			scores:   map[string]int{"Alice": 85, "Bob": 90},
			bonus:    -10,
			expected: map[string]int{"Alice": 75, "Bob": 80},
		},
		{
			name:     "空映射",
			scores:   map[string]int{},
			bonus:    10,
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 复制映射以避免修改测试数据
			scores := make(map[string]int)
			for k, v := range tt.scores {
				scores[k] = v
			}

			addBonus(scores, tt.bonus)

			if !reflect.DeepEqual(scores, tt.expected) {
				t.Errorf("addBonus(%v, %d) = %v, want %v",
					tt.scores, tt.bonus, scores, tt.expected)
			}
		})
	}
}

// ====================
// 4. copyMap 测试
// ====================

func TestCopyMap(t *testing.T) {
	tests := []struct {
		name     string
		original map[string]int
	}{
		{
			name:     "正常映射",
			original: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name:     "空映射",
			original: map[string]int{},
		},
		{
			name:     "单元素映射",
			original: map[string]int{"only": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copied := copyMap(tt.original)

			// 验证内容相等
			if !reflect.DeepEqual(copied, tt.original) {
				t.Errorf("copyMap(%v) = %v, want equal content", tt.original, copied)
			}

			// 验证是独立副本（修改不影响原映射）
			if len(tt.original) > 0 {
				for k := range copied {
					copied[k] = 9999
					break
				}
				// 原映射不应该被修改
				for k, v := range tt.original {
					if v == 9999 {
						t.Errorf("copyMap did not create independent copy, key %s was modified", k)
					}
				}
			}
		})
	}
}

// ====================
// 5. filterMap 测试
// ====================

func TestFilterMap(t *testing.T) {
	tests := []struct {
		name      string
		scores    map[string]int
		predicate func(int) bool
		expected  map[string]int
	}{
		{
			name:      "过滤高分(>=90)",
			scores:    map[string]int{"Alice": 85, "Bob": 92, "Charlie": 78, "David": 95},
			predicate: func(score int) bool { return score >= 90 },
			expected:  map[string]int{"Bob": 92, "David": 95},
		},
		{
			name:      "过滤及格(>=60)",
			scores:    map[string]int{"A": 55, "B": 60, "C": 75},
			predicate: func(score int) bool { return score >= 60 },
			expected:  map[string]int{"B": 60, "C": 75},
		},
		{
			name:      "全部通过",
			scores:    map[string]int{"A": 100, "B": 100},
			predicate: func(score int) bool { return score >= 60 },
			expected:  map[string]int{"A": 100, "B": 100},
		},
		{
			name:      "全部不通过",
			scores:    map[string]int{"A": 50, "B": 40},
			predicate: func(score int) bool { return score >= 60 },
			expected:  map[string]int{},
		},
		{
			name:      "空映射",
			scores:    map[string]int{},
			predicate: func(score int) bool { return true },
			expected:  map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMap(tt.scores, tt.predicate)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("filterMap(%v) = %v, want %v", tt.scores, result, tt.expected)
			}
		})
	}
}

// ====================
// 6. countWords 测试
// ====================

func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected map[string]int
	}{
		{
			name:     "基本句子",
			text:     "hello world",
			expected: map[string]int{"hello": 1, "world": 1},
		},
		{
			name:     "重复单词",
			text:     "the quick brown fox jumps over the lazy dog the fox",
			expected: map[string]int{"the": 3, "quick": 1, "brown": 1, "fox": 2, "jumps": 1, "over": 1, "lazy": 1, "dog": 1},
		},
		{
			name:     "单个单词",
			text:     "hello",
			expected: map[string]int{"hello": 1},
		},
		{
			name:     "空字符串",
			text:     "",
			expected: map[string]int{},
		},
		{
			name:     "多个空格",
			text:     "hello   world",
			expected: map[string]int{"hello": 1, "world": 1},
		},
		{
			name:     "全部相同单词",
			text:     "go go go",
			expected: map[string]int{"go": 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.text)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("countWords(%q) = %v, want %v", tt.text, result, tt.expected)
			}
		})
	}
}

// ====================
// 7. LRUCache 测试
// ====================

func TestNewLRUCache(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
	}{
		{"容量1", 1},
		{"容量3", 3},
		{"容量10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewLRUCache(tt.capacity)
			if cache == nil {
				t.Error("NewLRUCache returned nil")
			}
			if cache.capacity != tt.capacity {
				t.Errorf("capacity = %d, want %d", cache.capacity, tt.capacity)
			}
			if len(cache.data) != 0 {
				t.Errorf("data should be empty, got %v", cache.data)
			}
			if len(cache.order) != 0 {
				t.Errorf("order should be empty, got %v", cache.order)
			}
		})
	}
}

func TestLRUCache_Put(t *testing.T) {
	t.Run("基本Put操作", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("c", 3)

		if len(cache.data) != 3 {
			t.Errorf("cache size = %d, want 3", len(cache.data))
		}

		// 验证数据
		if cache.data["a"] != 1 || cache.data["b"] != 2 || cache.data["c"] != 3 {
			t.Errorf("cache data incorrect: %v", cache.data)
		}
	})

	t.Run("超出容量淘汰", func(t *testing.T) {
		cache := NewLRUCache(2)
		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("c", 3) // 应该淘汰最旧的

		if len(cache.data) != 2 {
			t.Errorf("cache size = %d, want 2", len(cache.data))
		}

		// 验证最旧的被淘汰
		if _, exists := cache.data["a"]; exists {
			t.Error("'a' should have been evicted")
		}
	})

	t.Run("更新已存在的键", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("a", 10) // 更新a

		if cache.data["a"] != 10 {
			t.Errorf("cache['a'] = %d, want 10", cache.data["a"])
		}

		// a应该移到最前面
		if cache.order[0] != "a" {
			t.Errorf("'a' should be at front, order = %v", cache.order)
		}
	})
}

func TestLRUCache_Get(t *testing.T) {
	t.Run("获取存在的键", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put("a", 1)
		cache.Put("b", 2)

		value, exists := cache.Get("a")
		if !exists {
			t.Error("'a' should exist")
		}
		if value != 1 {
			t.Errorf("value = %d, want 1", value)
		}
	})

	t.Run("获取不存在的键", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put("a", 1)

		value, exists := cache.Get("nonexistent")
		if exists {
			t.Error("'nonexistent' should not exist")
		}
		if value != 0 {
			t.Errorf("value = %d, want 0", value)
		}
	})

	t.Run("Get更新访问顺序", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put("a", 1)
		cache.Put("b", 2)
		cache.Put("c", 3)

		cache.Get("a") // 访问a，使其成为最新

		// a应该在最前面
		if cache.order[0] != "a" {
			t.Errorf("'a' should be at front after Get, order = %v", cache.order)
		}
	})
}

func TestLRUCache_Keys(t *testing.T) {
	cache := NewLRUCache(3)
	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)

	keys := cache.Keys()
	if len(keys) != 3 {
		t.Errorf("keys length = %d, want 3", len(keys))
	}

	// 最新的应该在前面
	if keys[0] != "c" {
		t.Errorf("newest key should be 'c', got %s", keys[0])
	}
}

func TestLRUCache_Integration(t *testing.T) {
	// 集成测试：模拟实际使用场景
	cache := NewLRUCache(3)

	// 添加三个元素
	cache.Put("a", 1)
	cache.Put("b", 2)
	cache.Put("c", 3)

	// 访问a，使其成为最新
	cache.Get("a")

	// 添加d，应该淘汰最旧的b
	cache.Put("d", 4)

	// 验证b被淘汰
	if _, exists := cache.Get("b"); exists {
		t.Error("'b' should have been evicted")
	}

	// 验证其他元素存在
	if _, exists := cache.Get("a"); !exists {
		t.Error("'a' should exist")
	}
	if _, exists := cache.Get("c"); !exists {
		t.Error("'c' should exist")
	}
	if _, exists := cache.Get("d"); !exists {
		t.Error("'d' should exist")
	}
}

// ====================
// 8. 基准测试
// ====================

func BenchmarkCalculateAverage(b *testing.B) {
	scores := map[string]int{
		"Alice": 85, "Bob": 92, "Charlie": 78,
		"David": 95, "Eve": 88, "Frank": 91,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateAverage(scores)
	}
}

func BenchmarkCountWords(b *testing.B) {
	text := "the quick brown fox jumps over the lazy dog the fox is quick"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countWords(text)
	}
}

func BenchmarkLRUCache_Put(b *testing.B) {
	cache := NewLRUCache(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune('a' + (i % 26)))
		cache.Put(key, i)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	cache := NewLRUCache(100)
	for i := 0; i < 100; i++ {
		cache.Put(string(rune('a'+i)), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune('a' + (i % 26)))
		cache.Get(key)
	}
}

func BenchmarkCopyMap(b *testing.B) {
	original := make(map[string]int)
	for i := 0; i < 100; i++ {
		original[string(rune('a'+i))] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copyMap(original)
	}
}

func BenchmarkFilterMap(b *testing.B) {
	scores := make(map[string]int)
	for i := 0; i < 100; i++ {
		scores[string(rune('a'+i))] = i * 10
	}
	predicate := func(score int) bool { return score >= 500 }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filterMap(scores, predicate)
	}
}

// ====================
// 9. 边界条件测试
// ====================

func TestEdgeCases(t *testing.T) {
	t.Run("nil映射处理", func(t *testing.T) {
		// calculateAverage 应该处理空映射
		result := calculateAverage(map[string]int{})
		if result != 0 {
			t.Errorf("calculateAverage(empty) = %f, want 0", result)
		}
	})

	t.Run("大数据量映射", func(t *testing.T) {
		largeMap := make(map[string]int)
		for i := 0; i < 10000; i++ {
			largeMap[string(rune(i))] = i
		}

		copied := copyMap(largeMap)
		if len(copied) != len(largeMap) {
			t.Errorf("copyMap large map: got %d elements, want %d",
				len(copied), len(largeMap))
		}
	})

	t.Run("特殊字符键", func(t *testing.T) {
		scores := map[string]int{
			"":      50,
			" ":     60,
			"中文":    70,
			"emoji": 80,
		}

		avg := calculateAverage(scores)
		expected := 65.0
		if avg != expected {
			t.Errorf("calculateAverage with special keys = %f, want %f", avg, expected)
		}
	})
}

// ====================
// 10. 辅助测试函数
// ====================

func TestMapSetToSlice_Sorted(t *testing.T) {
	// 验证结果是排序的
	input := map[int]bool{5: true, 2: true, 8: true, 1: true, 9: true}
	result := mapSetToSlice(input)

	// 检查是否已排序
	if !sort.IntsAreSorted(result) {
		t.Errorf("mapSetToSlice result should be sorted, got %v", result)
	}
}
