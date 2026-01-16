// Package utils 时间工具函数测试
package utils

import (
	"testing"
	"time"
)

// =============================================================================
// FormatDuration 函数测试
// =============================================================================

// TestFormatDuration 测试时间间隔格式化
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		// 正向用例
		{"秒级别", 30 * time.Second, "30.0秒"},
		{"分钟级别", 5 * time.Minute, "5.0分钟"},
		{"小时级别", 3 * time.Hour, "3.0小时"},
		{"天级别", 48 * time.Hour, "2.0天"},
		// 边界条件
		{"刚好1分钟", time.Minute, "1.0分钟"},
		{"刚好1小时", time.Hour, "1.0小时"},
		{"刚好24小时", 24 * time.Hour, "1.0天"},
		{"59秒", 59 * time.Second, "59.0秒"},
		{"59分钟", 59 * time.Minute, "59.0分钟"},
		{"23小时", 23 * time.Hour, "23.0小时"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q; 期望 %q", tt.duration, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetChineseWeekday 函数测试
// =============================================================================

// TestGetChineseWeekday 测试获取中文星期
func TestGetChineseWeekday(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{"星期一", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "星期一"},
		{"星期二", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), "星期二"},
		{"星期三", time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), "星期三"},
		{"星期四", time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), "星期四"},
		{"星期五", time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), "星期五"},
		{"星期六", time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), "星期六"},
		{"星期日", time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), "星期日"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetChineseWeekday(tt.time)
			if result != tt.expected {
				t.Errorf("GetChineseWeekday(%v) = %q; 期望 %q", tt.time, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// IsWorkday 函数测试
// =============================================================================

// TestIsWorkday 测试工作日判断
func TestIsWorkday(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		// 工作日
		{"星期一是工作日", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true},
		{"星期二是工作日", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), true},
		{"星期三是工作日", time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), true},
		{"星期四是工作日", time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC), true},
		{"星期五是工作日", time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC), true},
		// 周末
		{"星期六不是工作日", time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC), false},
		{"星期日不是工作日", time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWorkday(tt.time)
			if result != tt.expected {
				t.Errorf("IsWorkday(%v) = %v; 期望 %v", tt.time, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetMonthStart 和 GetMonthEnd 函数测试
// =============================================================================

// TestGetMonthStart 测试获取月份开始时间
func TestGetMonthStart(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			"月中某天",
			time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC),
			time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"月初",
			time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			"月末",
			time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
			time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMonthStart(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("GetMonthStart(%v) = %v; 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetMonthEnd 测试获取月份结束时间
func TestGetMonthEnd(t *testing.T) {
	tests := []struct {
		name        string
		input       time.Time
		expectedDay int
	}{
		{"1月31天", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), 31},
		{"2月闰年29天", time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), 29},
		{"2月平年28天", time.Date(2023, 2, 15, 0, 0, 0, 0, time.UTC), 28},
		{"4月30天", time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC), 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMonthEnd(tt.input)
			if result.Day() != tt.expectedDay {
				t.Errorf("GetMonthEnd(%v).Day() = %d; 期望 %d", tt.input, result.Day(), tt.expectedDay)
			}
		})
	}
}

// =============================================================================
// GetWeekStart 和 GetWeekEnd 函数测试
// =============================================================================

// TestGetWeekStart 测试获取一周开始时间（星期一）
func TestGetWeekStart(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{
			"周三",
			time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC), // 周三
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // 周一
		},
		{
			"周一",
			time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), // 周一
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // 周一
		},
		{
			"周日",
			time.Date(2024, 1, 7, 10, 0, 0, 0, time.UTC), // 周日
			time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),  // 周一
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWeekStart(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("GetWeekStart(%v) = %v; 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// ParseChineseDate 函数测试
// =============================================================================

// TestParseChineseDate 测试解析中文日期格式
func TestParseChineseDate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedErr bool
		expectedDay int
	}{
		// 正向用例
		{"中文格式1", "2024年01月15日", false, 15},
		{"中文格式2", "2024年1月5日", false, 5},
		{"横线格式", "2024-01-15", false, 15},
		{"斜线格式", "2024/01/15", false, 15},
		// 异常路径
		{"无效格式", "15-01-2024", true, 0},
		{"空字符串", "", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseChineseDate(tt.input)
			if tt.expectedErr {
				if err == nil {
					t.Errorf("ParseChineseDate(%q) 应返回错误", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ParseChineseDate(%q) 返回错误: %v", tt.input, err)
				} else if result.Day() != tt.expectedDay {
					t.Errorf("ParseChineseDate(%q).Day() = %d; 期望 %d", tt.input, result.Day(), tt.expectedDay)
				}
			}
		})
	}
}

// =============================================================================
// FormatChineseDate 函数测试
// =============================================================================

// TestFormatChineseDate 测试格式化为中文日期
func TestFormatChineseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			"正常日期",
			time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			"2024年01月15日",
		},
		{
			"月份个位数",
			time.Date(2024, 3, 5, 0, 0, 0, 0, time.UTC),
			"2024年03月05日",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatChineseDate(tt.input)
			if result != tt.expected {
				t.Errorf("FormatChineseDate(%v) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// IsLeapYear 函数测试
// =============================================================================

// TestIsLeapYear 测试闰年判断
func TestIsLeapYear(t *testing.T) {
	tests := []struct {
		name     string
		year     int
		expected bool
	}{
		// 闰年
		{"能被4整除", 2024, true},
		{"能被400整除", 2000, true},
		// 非闰年
		{"能被100整除但不能被400整除", 1900, false},
		{"不能被4整除", 2023, false},
		{"普通年份", 2021, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLeapYear(tt.year)
			if result != tt.expected {
				t.Errorf("IsLeapYear(%d) = %v; 期望 %v", tt.year, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetQuarter 函数测试
// =============================================================================

// TestGetQuarter 测试获取季度
func TestGetQuarter(t *testing.T) {
	tests := []struct {
		name     string
		month    time.Month
		expected int
	}{
		{"1月第一季度", time.January, 1},
		{"3月第一季度", time.March, 1},
		{"4月第二季度", time.April, 2},
		{"6月第二季度", time.June, 2},
		{"7月第三季度", time.July, 3},
		{"9月第三季度", time.September, 3},
		{"10月第四季度", time.October, 4},
		{"12月第四季度", time.December, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, tt.month, 15, 0, 0, 0, 0, time.UTC)
			result := GetQuarter(testTime)
			if result != tt.expected {
				t.Errorf("GetQuarter(%v) = %d; 期望 %d", testTime, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkFormatDuration 时间格式化基准测试
func BenchmarkFormatDuration(b *testing.B) {
	d := 5 * time.Hour
	for i := 0; i < b.N; i++ {
		FormatDuration(d)
	}
}

// BenchmarkIsLeapYear 闰年判断基准测试
func BenchmarkIsLeapYear(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsLeapYear(2024)
	}
}

// BenchmarkGetQuarter 季度获取基准测试
func BenchmarkGetQuarter(b *testing.B) {
	t := time.Now()
	for i := 0; i < b.N; i++ {
		GetQuarter(t)
	}
}
