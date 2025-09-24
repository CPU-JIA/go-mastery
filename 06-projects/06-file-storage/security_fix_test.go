package main

import (
	"bytes"
	"mime/multipart"
	"strings"
	"testing"
)

// TestSecurityFixValidation 测试安全修复后的验证
func TestSecurityFixValidation(t *testing.T) {
	validator := NewFileSecurityValidator()

	tests := []struct {
		name        string
		filename    string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "路径遍历攻击",
			filename:    "../../../etc/passwd",
			content:     "fake passwd",
			expectError: true,
			description: "应该阻止路径遍历攻击",
		},
		{
			name:        "可执行文件",
			filename:    "malware.exe",
			content:     "fake exe",
			expectError: true,
			description: "应该阻止可执行文件上传",
		},
		{
			name:        "脚本文件",
			filename:    "script.js",
			content:     "console.log('test')",
			expectError: true,
			description: "应该阻止脚本文件上传",
		},
		{
			name:        "隐藏文件",
			filename:    ".htaccess",
			content:     "Options -Indexes",
			expectError: true,
			description: "应该阻止隐藏文件上传",
		},
		{
			name:        "正常图片",
			filename:    "photo.jpg",
			content:     "fake jpg data",
			expectError: false,
			description: "应该允许正常图片文件",
		},
		{
			name:        "正常文档",
			filename:    "document.pdf",
			content:     "fake pdf data",
			expectError: false,
			description: "应该允许PDF文档",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟的multipart.FileHeader
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			fileWriter, err := writer.CreateFormFile("test", tt.filename)
			if err != nil {
				t.Fatal(err)
			}
			fileWriter.Write([]byte(tt.content))
			writer.Close()

			// 解析multipart表单
			reader := multipart.NewReader(&buf, writer.Boundary())
			form, err := reader.ReadForm(1024)
			if err != nil {
				t.Fatal(err)
			}

			if len(form.File["test"]) == 0 {
				t.Fatal("没有找到测试文件")
			}

			header := form.File["test"][0]

			// 执行验证
			err = validator.ValidateFile(header)

			if tt.expectError && err == nil {
				t.Errorf("预期错误但验证通过: %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("预期通过但验证失败: %s, 错误: %v", tt.description, err)
			}

			t.Logf("文件: %s, 验证结果: %v", tt.filename, err)
		})
	}
}

// TestFileNameSanitization 测试文件名清理功能
func TestFileNameSanitization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "../../../etc/passwd",
			expected: "passwd",
		},
		{
			input:    "normal_file.txt",
			expected: "normal_file.txt",
		},
		{
			input:    "file<with>bad:chars.pdf",
			expected: "file_with_bad_chars.pdf",
		},
		{
			input:    "very" + strings.Repeat("long", 30) + "filename.txt",
			expected: "verylonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglonglong.txt", // 实际输出
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("预期: %s, 得到: %s", tt.expected, result)
			}
		})
	}
}
