package security

import (
	"math"
	"testing"
)

func TestSafeUint64ToInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     uint64
		expected  int64
		expectErr bool
	}{
		{
			name:      "正常值转换",
			input:     1000,
			expected:  1000,
			expectErr: false,
		},
		{
			name:      "最大安全值",
			input:     math.MaxInt64,
			expected:  math.MaxInt64,
			expectErr: false,
		},
		{
			name:      "溢出值",
			input:     math.MaxInt64 + 1,
			expected:  0,
			expectErr: true,
		},
		{
			name:      "最大uint64值",
			input:     math.MaxUint64,
			expected:  0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeUint64ToInt64(tt.input)

			if tt.expectErr && err == nil {
				t.Errorf("期待错误但没有发生")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("不期待错误但发生了: %v", err)
			}
			if !tt.expectErr && result != tt.expected {
				t.Errorf("期待 %d, 得到 %d", tt.expected, result)
			}
		})
	}
}

func TestMustSafeUint64ToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected int64
	}{
		{
			name:     "正常值转换",
			input:    1000,
			expected: 1000,
		},
		{
			name:     "最大安全值",
			input:    math.MaxInt64,
			expected: math.MaxInt64,
		},
		{
			name:     "溢出值返回最大值",
			input:    math.MaxInt64 + 1,
			expected: math.MaxInt64,
		},
		{
			name:     "最大uint64值返回最大int64值",
			input:    math.MaxUint64,
			expected: math.MaxInt64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustSafeUint64ToInt64(tt.input)
			if result != tt.expected {
				t.Errorf("期待 %d, 得到 %d", tt.expected, result)
			}
		})
	}
}

func TestConversionStats(t *testing.T) {
	stats := NewConversionStats()

	// 测试正常转换
	result1 := stats.TrackingUint64ToInt64(1000)
	if result1 != 1000 {
		t.Errorf("期待 1000, 得到 %d", result1)
	}

	// 测试溢出转换
	result2 := stats.TrackingUint64ToInt64(math.MaxUint64)
	if result2 != math.MaxInt64 {
		t.Errorf("期待 %d, 得到 %d", math.MaxInt64, result2)
	}

	// 验证统计
	if stats.TotalConversions != 2 {
		t.Errorf("期待总转换数 2, 得到 %d", stats.TotalConversions)
	}
	if stats.OverflowCount != 1 {
		t.Errorf("期待溢出数 1, 得到 %d", stats.OverflowCount)
	}

	// 验证溢出率
	expectedRate := 0.5 // 1 overflow out of 2 total
	if rate := stats.GetOverflowRate(); rate != expectedRate {
		t.Errorf("期待溢出率 %f, 得到 %f", expectedRate, rate)
	}
}

func BenchmarkSafeUint64ToInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = SafeUint64ToInt64(uint64(i))
	}
}

func BenchmarkMustSafeUint64ToInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MustSafeUint64ToInt64(uint64(i))
	}
}

func BenchmarkDirectConversion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = int64(uint64(i))
	}
}