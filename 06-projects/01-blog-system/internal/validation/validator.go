package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// CustomValidator 自定义验证器
type CustomValidator struct {
	validate *validator.Validate
}

// NewCustomValidator 创建新的验证器实例
func NewCustomValidator() *CustomValidator {
	validate := validator.New()
	cv := &CustomValidator{validate: validate}

	// 注册自定义验证规则
	cv.registerCustomValidations()

	return cv
}

// ValidateStruct 验证结构体
func (cv *CustomValidator) ValidateStruct(obj interface{}) error {
	return cv.validate.Struct(obj)
}

// registerCustomValidations 注册自定义验证规则
func (cv *CustomValidator) registerCustomValidations() {
	// 强密码验证
	cv.validate.RegisterValidation("strong_password", cv.validateStrongPassword)

	// 用户名验证（只允许字母、数字、下划线、连字符）
	cv.validate.RegisterValidation("safe_username", cv.validateSafeUsername)

	// 安全字符串验证（防止XSS）
	cv.validate.RegisterValidation("safe_string", cv.validateSafeString)

	// URL slug验证
	cv.validate.RegisterValidation("url_slug", cv.validateURLSlug)

	// 电话号码验证
	cv.validate.RegisterValidation("phone_number", cv.validatePhoneNumber)

	// 安全HTML内容验证
	cv.validate.RegisterValidation("safe_html", cv.validateSafeHTML)

	// 文件名验证
	cv.validate.RegisterValidation("safe_filename", cv.validateSafeFilename)

	// IP地址验证
	cv.validate.RegisterValidation("ip_address", cv.validateIPAddress)
}

// validateStrongPassword 强密码验证
func (cv *CustomValidator) validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// 最小长度检查
	if len(password) < 8 {
		return false
	}

	// 最大长度检查（防止DoS攻击）
	if len(password) > 128 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// 检查字符类型
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 至少包含三种字符类型
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// validateSafeUsername 安全用户名验证
func (cv *CustomValidator) validateSafeUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// 长度检查
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	// 只允许字母、数字、下划线、连字符
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return false
	}

	// 不能以特殊字符开头或结尾
	if strings.HasPrefix(username, "_") || strings.HasPrefix(username, "-") ||
		strings.HasSuffix(username, "_") || strings.HasSuffix(username, "-") {
		return false
	}

	// 禁止的用户名
	forbiddenUsernames := []string{
		"admin", "administrator", "root", "system", "user", "test", "guest",
		"null", "undefined", "api", "www", "mail", "email", "support",
		"help", "info", "contact", "about", "login", "register", "signup",
		"signin", "logout", "password", "security", "config", "settings",
	}

	for _, forbidden := range forbiddenUsernames {
		if strings.EqualFold(username, forbidden) {
			return false
		}
	}

	return true
}

// validateSafeString 安全字符串验证（防止XSS）
func (cv *CustomValidator) validateSafeString(fl validator.FieldLevel) bool {
	input := fl.Field().String()

	// 检查危险的HTML标签和JavaScript
	dangerousPatterns := []string{
		`<script[\s\S]*?>[\s\S]*?</script>`,
		`<iframe[\s\S]*?>`,
		`<object[\s\S]*?>`,
		`<embed[\s\S]*?>`,
		`<link[\s\S]*?>`,
		`<meta[\s\S]*?>`,
		`javascript:`,
		`vbscript:`,
		`on\w+\s*=`,
		`expression\s*\(`,
		`eval\s*\(`,
		`alert\s*\(`,
		`confirm\s*\(`,
		`prompt\s*\(`,
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(`(?i)`+pattern, input)
		if matched {
			return false
		}
	}

	return true
}

// validateURLSlug URL slug验证
func (cv *CustomValidator) validateURLSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()

	// 长度检查
	if len(slug) < 1 || len(slug) > 100 {
		return false
	}

	// 只允许小写字母、数字、连字符
	matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, slug)
	if !matched {
		return false
	}

	// 不能以连字符开头或结尾
	if strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}

	// 不能包含连续的连字符
	if strings.Contains(slug, "--") {
		return false
	}

	return true
}

// validatePhoneNumber 电话号码验证
func (cv *CustomValidator) validatePhoneNumber(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// 如果为空则允许（optional字段）
	if phone == "" {
		return true
	}

	// 基本的电话号码格式验证
	// 支持格式：+86-138-0013-8000, +1-555-123-4567, 13800138000等
	patterns := []string{
		`^\+\d{1,3}-?\d{1,14}$`, // 国际格式
		`^1[3-9]\d{9}$`,         // 中国手机号
		`^\d{3}-?\d{3}-?\d{4}$`, // 美国格式
		`^\d{10,15}$`,           // 纯数字格式
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, phone)
		if matched {
			return true
		}
	}

	return false
}

// validateSafeHTML 安全HTML内容验证
func (cv *CustomValidator) validateSafeHTML(fl validator.FieldLevel) bool {
	html := fl.Field().String()

	// 危险的HTML标签
	dangerousTags := []string{
		"script", "iframe", "object", "embed", "form", "input", "button",
		"select", "textarea", "link", "meta", "style", "base", "frame",
		"frameset", "applet", "canvas", "svg", "math",
	}

	// 检查危险标签
	for _, tag := range dangerousTags {
		pattern := fmt.Sprintf(`(?i)<%s[\s\S]*?>`, tag)
		matched, _ := regexp.MatchString(pattern, html)
		if matched {
			return false
		}
	}

	// 检查JavaScript事件处理器
	eventHandlers := []string{
		"onclick", "onload", "onmouseover", "onmouseout", "onfocus", "onblur",
		"onchange", "onsubmit", "onreset", "onselect", "onkeydown", "onkeyup",
		"onkeypress", "onmousedown", "onmouseup", "onmousemove",
	}

	for _, handler := range eventHandlers {
		pattern := fmt.Sprintf(`(?i)%s\s*=`, handler)
		matched, _ := regexp.MatchString(pattern, html)
		if matched {
			return false
		}
	}

	return true
}

// validateSafeFilename 安全文件名验证
func (cv *CustomValidator) validateSafeFilename(fl validator.FieldLevel) bool {
	filename := fl.Field().String()

	// 长度检查
	if len(filename) < 1 || len(filename) > 255 {
		return false
	}

	// 禁止的字符
	forbiddenChars := []string{
		"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t",
	}

	for _, char := range forbiddenChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// 禁止的文件名
	forbiddenNames := []string{
		".", "..", "CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	for _, name := range forbiddenNames {
		if strings.EqualFold(filename, name) {
			return false
		}
	}

	// 不能以点或空格开头/结尾
	if strings.HasPrefix(filename, ".") || strings.HasPrefix(filename, " ") ||
		strings.HasSuffix(filename, ".") || strings.HasSuffix(filename, " ") {
		return false
	}

	return true
}

// validateIPAddress IP地址验证
func (cv *CustomValidator) validateIPAddress(fl validator.FieldLevel) bool {
	ip := fl.Field().String()

	// IPv4验证
	ipv4Pattern := `^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$`
	matched, _ := regexp.MatchString(ipv4Pattern, ip)
	if matched {
		// 验证每个段是否在0-255范围内
		parts := strings.Split(ip, ".")
		for _, part := range parts {
			if len(part) > 1 && part[0] == '0' {
				return false // 禁止前导零
			}
			var num int
			fmt.Sscanf(part, "%d", &num)
			if num < 0 || num > 255 {
				return false
			}
		}
		return true
	}

	// IPv6验证（简化版）
	ipv6Pattern := `^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$`
	matched, _ = regexp.MatchString(ipv6Pattern, ip)
	return matched
}

// InitCustomValidator 初始化自定义验证器到Gin
func InitCustomValidator() {
	cv := NewCustomValidator()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证规则到Gin的验证器
		v.RegisterValidation("strong_password", cv.validateStrongPassword)
		v.RegisterValidation("safe_username", cv.validateSafeUsername)
		v.RegisterValidation("safe_string", cv.validateSafeString)
		v.RegisterValidation("url_slug", cv.validateURLSlug)
		v.RegisterValidation("phone_number", cv.validatePhoneNumber)
		v.RegisterValidation("safe_html", cv.validateSafeHTML)
		v.RegisterValidation("safe_filename", cv.validateSafeFilename)
		v.RegisterValidation("ip_address", cv.validateIPAddress)
	}
}

// ValidateEmail 增强的邮箱验证
func ValidateEmail(email string) error {
	// 基本格式验证
	if len(email) < 5 || len(email) > 254 {
		return fmt.Errorf("邮箱长度必须在5-254字符之间")
	}

	// 使用标准库验证
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("邮箱格式无效")
	}

	// 检查危险字符
	dangerousChars := []string{"<", ">", "\"", "'", "&", "script", "javascript"}
	lowerEmail := strings.ToLower(email)
	for _, char := range dangerousChars {
		if strings.Contains(lowerEmail, char) {
			return fmt.Errorf("邮箱包含非法字符")
		}
	}

	// 检查域名黑名单（可选）
	blockedDomains := []string{
		"tempmail.org", "10minutemail.com", "guerrillamail.com",
		"mailinator.com", "yopmail.com",
	}

	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		domain := strings.ToLower(parts[1])
		for _, blocked := range blockedDomains {
			if domain == blocked {
				return fmt.Errorf("不支持该邮箱服务商")
			}
		}
	}

	return nil
}
