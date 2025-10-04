package middleware

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// SecurityHeaders 添加安全响应头
func SecurityHeaders() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Content Security Policy (CSP)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")

		// X-Frame-Options 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection 启用XSS过滤
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict-Transport-Security 强制HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Referrer-Policy 控制Referrer信息
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy 限制浏览器API访问
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		// X-Permitted-Cross-Domain-Policies 限制跨域策略
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Clear server header
		c.Header("Server", "")

		c.Next()
	})
}

// InputSanitizer 输入净化中间件
func InputSanitizer() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 为POST、PUT、PATCH请求净化请求体
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// 获取Content-Type
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// JSON数据的净化会在后续的绑定和验证阶段处理
				// 这里主要验证Content-Type和大小限制
				if c.Request.ContentLength > 1024*1024 { // 1MB limit
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{
						"error": "Request body too large",
						"code":  "REQUEST_TOO_LARGE",
					})
					c.Abort()
					return
				}
			}
		}

		// 净化URL参数
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				c.Request.URL.Query()[key][i] = SanitizeInput(value)
			}
		}

		c.Next()
	})
}

// RateLimiter 速率限制中间件
func RateLimiter(requestsPerSecond int, burst int) gin.HandlerFunc {
	// 使用IP作为键的限流器映射
	limiters := make(map[string]*rate.Limiter)

	return gin.HandlerFunc(func(c *gin.Context) {
		ip := c.ClientIP()

		// 获取或创建该IP的限流器
		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(requestsPerSecond)), burst)
			limiters[ip] = limiter
		}

		// 检查是否允许请求
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ValidateContentType 验证Content-Type
func ValidateContentType(allowedTypes ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")

			valid := false
			for _, allowedType := range allowedTypes {
				if strings.Contains(contentType, allowedType) {
					valid = true
					break
				}
			}

			if !valid {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error": "Unsupported content type",
					"code":  "UNSUPPORTED_MEDIA_TYPE",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// SanitizeInput 净化输入数据
func SanitizeInput(input string) string {
	// HTML转义
	input = html.EscapeString(input)

	// 移除潜在危险的字符和序列
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	input = strings.ReplaceAll(input, "/", "&#x2F;")

	// 移除SQL注入相关关键字（基础防护，主要依赖参数化查询）
	sqlPatterns := []string{
		`(?i)(union\s+select)`,
		`(?i)(insert\s+into)`,
		`(?i)(delete\s+from)`,
		`(?i)(drop\s+table)`,
		`(?i)(update\s+.*\s+set)`,
		`(?i)(exec\s*\()`,
		`(?i)(execute\s*\()`,
		`(?i)(sp_\w+)`,
		`(?i)(xp_\w+)`,
	}

	for _, pattern := range sqlPatterns {
		re := regexp.MustCompile(pattern)
		input = re.ReplaceAllString(input, "")
	}

	// 移除JavaScript相关内容
	jsPatterns := []string{
		`(?i)(javascript:)`,
		`(?i)(on\w+\s*=)`,
		`(?i)(<script[\s\S]*?</script>)`,
		`(?i)(eval\s*\()`,
		`(?i)(expression\s*\()`,
	}

	for _, pattern := range jsPatterns {
		re := regexp.MustCompile(pattern)
		input = re.ReplaceAllString(input, "")
	}

	// 限制长度
	if len(input) > 10000 { // 10KB limit
		input = input[:10000]
	}

	return strings.TrimSpace(input)
}

// PathTraversalProtection 路径遍历攻击保护
func PathTraversalProtection() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 检查路径遍历攻击模式
		dangerous := []string{
			"../",
			"..\\",
			"%2e%2e%2f",
			"%2e%2e\\",
			"..%2f",
			"..%5c",
		}

		for _, pattern := range dangerous {
			if strings.Contains(strings.ToLower(path), pattern) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid path",
					"code":  "INVALID_PATH",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	})
}

// RequestLogger 安全请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// 处理请求
		c.Next()

		// 记录请求信息（确保不记录敏感信息）
		latency := time.Since(start)
		status := c.Writer.Status()

		// 不记录包含密码的URL参数
		safeQuery := c.Request.URL.RawQuery
		if strings.Contains(strings.ToLower(safeQuery), "password") {
			safeQuery = "[REDACTED]"
		}

		logMsg := fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %s?%s",
			time.Now().Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			c.ClientIP(),
			c.Request.Method,
			c.Request.URL.Path,
			safeQuery,
		)

		// 根据状态码决定日志级别
		if status >= 400 {
			fmt.Printf("[SECURITY WARNING] %s\n", logMsg)
		} else {
			fmt.Printf("[INFO] %s\n", logMsg)
		}
	})
}
