package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// TestFileUploadSecurity 测试文件上传安全性
func TestFileUploadSecurity(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
		description string
	}{
		{
			name:        "路径遍历攻击1",
			filename:    "../../../etc/passwd",
			expectError: true,
			description: "不应允许路径遍历攻击",
		},
		{
			name:        "路径遍历攻击2",
			filename:    "..\\..\\windows\\system32\\hosts",
			expectError: true,
			description: "不应允许Windows风格的路径遍历",
		},
		{
			name:        "可执行文件",
			filename:    "malware.exe",
			expectError: true,
			description: "不应允许可执行文件上传",
		},
		{
			name:        "脚本文件",
			filename:    "malicious.php",
			expectError: true,
			description: "不应允许脚本文件上传",
		},
		{
			name:        "JS文件",
			filename:    "malicious.js",
			expectError: true,
			description: "不应允许JavaScript文件上传",
		},
		{
			name:        "正常图片",
			filename:    "photo.jpg",
			expectError: false,
			description: "应允许正常图片文件",
		},
		{
			name:        "正常文档",
			filename:    "document.pdf",
			expectError: false,
			description: "应允许PDF文档",
		},
		{
			name:        "隐藏文件",
			filename:    ".htaccess",
			expectError: true,
			description: "不应允许隐藏文件",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileName(tt.filename)
			if tt.expectError && err == nil {
				t.Errorf("预期错误但验证通过: %s", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("预期通过但验证失败: %s, 错误: %v", tt.description, err)
			}
		})
	}
}

// 安全的文件名验证函数
func validateFileName(filename string) error {
	// 清理文件名
	cleanName := filepath.Base(filename)

	// 检查是否包含路径遍历字符
	if strings.Contains(filename, "..") ||
		strings.Contains(filename, "/") ||
		strings.Contains(filename, "\\") {
		return fmt.Errorf("文件名包含非法字符")
	}

	// 检查隐藏文件
	if strings.HasPrefix(cleanName, ".") {
		return fmt.Errorf("不允许上传隐藏文件")
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(cleanName))

	// 白名单验证
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".txt":  true,
		".csv":  true,
		".zip":  true,
		".tar":  true,
		".gz":   true,
	}

	if !allowedExtensions[ext] {
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 检查文件名长度
	if len(cleanName) > 255 {
		return fmt.Errorf("文件名过长")
	}

	// 检查是否为空
	if cleanName == "" || cleanName == ext {
		return fmt.Errorf("无效的文件名")
	}

	return nil
}

// TestMimeTypeValidation 测试MIME类型验证
func TestMimeTypeValidation(t *testing.T) {
	tests := []struct {
		filename    string
		contentType string
		expectError bool
	}{
		{
			filename:    "image.jpg",
			contentType: "image/jpeg",
			expectError: false,
		},
		{
			filename:    "fake.jpg",
			contentType: "application/x-executable", // 伪装的可执行文件
			expectError: true,
		},
		{
			filename:    "document.pdf",
			contentType: "application/pdf",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			err := validateMimeType(tt.filename, tt.contentType)
			if tt.expectError && err == nil {
				t.Error("预期MIME类型验证失败但通过了")
			}
			if !tt.expectError && err != nil {
				t.Errorf("预期MIME类型验证通过但失败了: %v", err)
			}
		})
	}
}

// 安全的MIME类型验证函数
func validateMimeType(filename, contentType string) error {
	ext := strings.ToLower(filepath.Ext(filename))

	// 定义扩展名与MIME类型的映射
	expectedMimeTypes := map[string][]string{
		".jpg":  {"image/jpeg", "image/jpg"},
		".jpeg": {"image/jpeg"},
		".png":  {"image/png"},
		".gif":  {"image/gif"},
		".bmp":  {"image/bmp"},
		".webp": {"image/webp"},
		".pdf":  {"application/pdf"},
		".txt":  {"text/plain"},
		".csv":  {"text/csv", "application/csv"},
		".zip":  {"application/zip"},
	}

	allowedTypes, exists := expectedMimeTypes[ext]
	if !exists {
		return fmt.Errorf("不支持的文件扩展名: %s", ext)
	}

	// 检查MIME类型是否匹配
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			return nil
		}
	}

	return fmt.Errorf("MIME类型 %s 与文件扩展名 %s 不匹配", contentType, ext)
}
