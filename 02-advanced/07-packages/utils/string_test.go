// Package utils 字符串工具函数测试
package utils

import (
	"testing"
)

// =============================================================================
// Reverse 函数测试
// =============================================================================

// TestReverse 测试字符串反转
func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正向用例
		{"英文字符串", "hello", "olleh"},
		{"中文字符串", "你好世界", "界世好你"},
		{"混合字符串", "hello你好", "好你olleh"},
		// 边界条件
		{"空字符串", "", ""},
		{"单字符", "a", "a"},
		{"回文字符串", "aba", "aba"},
		// 特殊字符
		{"带空格", "a b c", "c b a"},
		{"带数字", "abc123", "321cba"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.input)
			if result != tt.expected {
				t.Errorf("Reverse(%q) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// IsPalindrome 函数测试
// =============================================================================

// TestIsPalindrome 测试回文检测
func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// 正向用例 - 是回文
		{"简单回文", "aba", true},
		{"带空格回文", "a b a", true},
		{"大小写混合回文", "Aba", true},
		{"长回文", "level", true},
		// 负向用例 - 不是回文
		{"非回文", "hello", false},
		{"几乎回文", "abca", false},
		// 边界条件
		{"空字符串", "", true},
		{"单字符", "a", true},
		{"两个相同字符", "aa", true},
		{"两个不同字符", "ab", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPalindrome(tt.input)
			if result != tt.expected {
				t.Errorf("IsPalindrome(%q) = %v; 期望 %v", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Capitalize 函数测试
// =============================================================================

// TestCapitalize 测试首字母大写
func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正向用例
		{"小写开头", "hello", "Hello"},
		{"已大写", "Hello", "Hello"},
		{"全小写", "world", "World"},
		// 边界条件
		{"空字符串", "", ""},
		{"单字符小写", "a", "A"},
		{"单字符大写", "A", "A"},
		{"数字开头", "123abc", "123abc"},
		// 特殊情况
		{"中文开头", "你好world", "你好world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Capitalize(tt.input)
			if result != tt.expected {
				t.Errorf("Capitalize(%q) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// CamelCase 函数测试
// =============================================================================

// TestCamelCase 测试驼峰命名转换
func TestCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正向用例
		{"两个单词", "hello world", "helloWorld"},
		{"三个单词", "get user name", "getUserName"},
		{"大写单词", "HELLO WORLD", "helloWorld"},
		// 边界条件
		{"单个单词", "hello", "hello"},
		{"多空格分隔", "hello   world", "helloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("CamelCase(%q) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// SnakeCase 函数测试
// =============================================================================

// TestSnakeCase 测试下划线命名转换
func TestSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正向用例
		{"两个单词", "hello world", "hello_world"},
		{"三个单词", "get user name", "get_user_name"},
		{"大写单词", "HELLO WORLD", "hello_world"},
		// 边界条件
		{"单个单词", "hello", "hello"},
		{"多空格分隔", "hello   world", "hello_world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("SnakeCase(%q) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// KebabCase 函数测试
// =============================================================================

// TestKebabCase 测试短横线命名转换
func TestKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正向用例
		{"两个单词", "hello world", "hello-world"},
		{"三个单词", "get user name", "get-user-name"},
		{"大写单词", "HELLO WORLD", "hello-world"},
		// 边界条件
		{"单个单词", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := KebabCase(tt.input)
			if result != tt.expected {
				t.Errorf("KebabCase(%q) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// CountWords 函数测试
// =============================================================================

// TestCountWords 测试单词计数
func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// 正向用例
		{"三个单词", "hello world go", 3},
		{"单个单词", "hello", 1},
		{"多空格分隔", "hello   world", 2},
		// 边界条件
		{"空字符串", "", 0},
		{"只有空格", "   ", 0},
		{"前后有空格", "  hello world  ", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountWords(tt.input)
			if result != tt.expected {
				t.Errorf("CountWords(%q) = %d; 期望 %d", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Truncate 函数测试
// =============================================================================

// TestTruncate 测试字符串截断
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		length   int
		expected string
	}{
		// 正向用例
		{"正常截断", "hello world", 8, "hello..."},
		{"不需要截断", "hello", 10, "hello"},
		{"刚好长度", "hello", 5, "hello"},
		// 边界条件
		{"长度为0", "hello", 0, ""},
		{"长度为1", "hello", 1, "h"},
		{"长度为3", "hello", 3, "hel"},
		{"长度为4", "hello", 4, "h..."},
		{"空字符串", "", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q; 期望 %q", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// RemoveDuplicateSpaces 函数测试
// =============================================================================

// TestRemoveDuplicateSpaces 测试移除重复空格
func TestRemoveDuplicateSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正向用例
		{"多个空格", "hello   world", "hello world"},
		{"前后空格", "  hello world  ", "hello world"},
		{"混合空格", "  hello   world  go  ", "hello world go"},
		// 边界条件
		{"无重复空格", "hello world", "hello world"},
		{"空字符串", "", ""},
		{"只有空格", "     ", ""},
		{"单个单词", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveDuplicateSpaces(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveDuplicateSpaces(%q) = %q; 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkReverse 字符串反转基准测试
func BenchmarkReverse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Reverse("hello world")
	}
}

// BenchmarkIsPalindrome 回文检测基准测试
func BenchmarkIsPalindrome(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsPalindrome("a man a plan a canal panama")
	}
}

// BenchmarkCamelCase 驼峰转换基准测试
func BenchmarkCamelCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CamelCase("get user name by id")
	}
}
