package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// TestXSSPrevention 测试XSS防护
func TestXSSPrevention(t *testing.T) {
	// 创建一个包含XSS payload但格式有效的表单数据
	form := url.Values{
		"name":    {"John<script>alert('XSS')</script>"},
		"email":   {"test@example.com"}, // 有效邮箱
		"message": {"Hello<img src=x onerror=alert('XSS')>World"},
	}

	// 创建POST请求
	req, err := http.NewRequest("POST", "/form", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 创建响应记录器
	rr := httptest.NewRecorder()

	// 调用处理器
	formHandler(rr, req)

	// 检查响应
	if rr.Code != http.StatusOK {
		t.Errorf("预期状态码 200，得到 %d", rr.Code)
	}

	body := rr.Body.String()

	// 确保脚本标签被转义
	if strings.Contains(body, "<script>") {
		t.Error("响应包含未转义的script标签，存在XSS风险")
	}

	// 确保HTML标签被转义
	if strings.Contains(body, "<img src=x") {
		t.Error("响应包含未转义的img标签，存在XSS风险")
	}

	// 确保转义后的内容存在
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Error("脚本标签未正确转义")
	}
}

// TestInputValidation 测试输入验证
func TestInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		formData    url.Values
		expectError bool
	}{
		{
			name: "有效输入",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"john@example.com"},
				"message": {"Hello World"},
			},
			expectError: false,
		},
		{
			name: "空名称",
			formData: url.Values{
				"name":    {""},
				"email":   {"john@example.com"},
				"message": {"Hello World"},
			},
			expectError: true,
		},
		{
			name: "无效邮箱",
			formData: url.Values{
				"name":    {"John Doe"},
				"email":   {"invalid-email"},
				"message": {"Hello World"},
			},
			expectError: true,
		},
		{
			name: "过长输入",
			formData: url.Values{
				"name":    {strings.Repeat("a", 1000)},
				"email":   {"john@example.com"},
				"message": {"Hello World"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/form", strings.NewReader(tt.formData.Encode()))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			rr := httptest.NewRecorder()
			formHandler(rr, req)

			if tt.expectError && rr.Code == http.StatusOK {
				t.Error("预期验证错误，但请求成功")
			}
			if !tt.expectError && rr.Code != http.StatusOK {
				t.Errorf("预期成功，但得到状态码 %d", rr.Code)
			}
		})
	}
}
