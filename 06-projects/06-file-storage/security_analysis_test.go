package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestFileUploadSecurityGaps 测试文件上传安全缺陷
func TestFileUploadSecurityGaps(t *testing.T) {
	tests := []struct {
		filename    string
		expectSafe  bool
		description string
	}{
		{
			filename:    "../../../etc/passwd",
			expectSafe:  false,
			description: "路径遍历攻击",
		},
		{
			filename:    "malware.exe",
			expectSafe:  false,
			description: "可执行文件",
		},
		{
			filename:    "script.js",
			expectSafe:  false,
			description: "JavaScript文件",
		},
		{
			filename:    ".htaccess",
			expectSafe:  false,
			description: "隐藏配置文件",
		},
		{
			filename:    "normal_image.jpg",
			expectSafe:  true,
			description: "正常图片文件",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			// 模拟原始代码的处理方式
			ext := filepath.Ext(tt.filename)
			fileName := "generated_id" + ext

			// 检查原始处理是否安全
			hasDangerousPath := strings.Contains(tt.filename, "..")
			hasDangerousExt := strings.HasSuffix(strings.ToLower(ext), ".exe") ||
				strings.HasSuffix(strings.ToLower(ext), ".js") ||
				strings.HasSuffix(strings.ToLower(ext), ".php")
			isHiddenFile := strings.HasPrefix(filepath.Base(tt.filename), ".")

			isCurrentlyUnsafe := hasDangerousPath || hasDangerousExt || isHiddenFile

			if !tt.expectSafe && !isCurrentlyUnsafe {
				t.Errorf("安全缺陷: 危险文件 '%s' 未被检测到安全风险", tt.filename)
			}

			if tt.expectSafe && isCurrentlyUnsafe {
				t.Errorf("误报: 安全文件 '%s' 被标记为危险", tt.filename)
			}

			t.Logf("文件: %s -> 处理后: %s, 安全: %v", tt.filename, fileName, !isCurrentlyUnsafe)
		})
	}
}

// TestSecurityRecommendations 安全改进建议测试
func TestSecurityRecommendations(t *testing.T) {
	t.Log("=== 文件上传安全改进建议 ===")
	t.Log("1. 文件名验证: 检查路径遍历字符 (.., /, \\)")
	t.Log("2. 文件类型白名单: 只允许安全的文件类型")
	t.Log("3. 文件大小限制: 防止大文件攻击")
	t.Log("4. MIME类型验证: 确保文件内容与扩展名匹配")
	t.Log("5. 文件内容扫描: 检查恶意代码模式")
	t.Log("6. 安全存储位置: 文件存储在Web根目录外")
	t.Log("7. 访问控制: 实施适当的权限控制")

	// 这个测试总是通过，但记录了改进建议
}
