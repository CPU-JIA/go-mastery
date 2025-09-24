// Package utils 提供常用的工具函数
package utils

import (
	"strings"
	"unicode"
)

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// IsPalindrome 检查字符串是否为回文
func IsPalindrome(s string) bool {
	// 转换为小写并移除空格
	cleaned := strings.ToLower(strings.ReplaceAll(s, " ", ""))
	return cleaned == Reverse(cleaned)
}

// Capitalize 首字母大写
func Capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// CamelCase 转换为驼峰命名法
func CamelCase(s string) string {
	words := strings.Fields(s)
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += Capitalize(strings.ToLower(words[i]))
	}
	return result
}

// SnakeCase 转换为下划线命名法
func SnakeCase(s string) string {
	words := strings.Fields(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "_")
}

// KebabCase 转换为短横线命名法
func KebabCase(s string) string {
	words := strings.Fields(s)
	for i := range words {
		words[i] = strings.ToLower(words[i])
	}
	return strings.Join(words, "-")
}

// CountWords 统计单词数量
func CountWords(s string) int {
	return len(strings.Fields(s))
}

// Truncate 截断字符串到指定长度，添加省略号
func Truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// RemoveDuplicateSpaces 移除重复的空格
func RemoveDuplicateSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
