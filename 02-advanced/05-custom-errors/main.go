package main

import (
	"fmt"
	"strconv"
	"time"
)

// =============================================================================
// 1. 基础自定义错误类型
// =============================================================================

// MyError 实现了 error 接口的自定义错误类型
type MyError struct {
	Code    int
	Message string
	Time    time.Time
}

// Error 实现 error 接口
func (e *MyError) Error() string {
	return fmt.Sprintf("[%d] %s (发生时间: %v)", e.Code, e.Message, e.Time.Format("2006-01-02 15:04:05"))
}

// NewMyError 创建自定义错误的构造函数
func NewMyError(code int, message string) *MyError {
	return &MyError{
		Code:    code,
		Message: message,
		Time:    time.Now(),
	}
}

// =============================================================================
// 2. 包装错误和错误链
// =============================================================================

// ValidationError 验证错误类型
type ValidationError struct {
	Field string
	Value interface{}
	Rule  string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("验证失败: 字段'%s'的值'%v'不符合规则'%s'", e.Field, e.Value, e.Rule)
}

// DatabaseError 数据库错误类型
type DatabaseError struct {
	Operation string
	Table     string
	Cause     error // 原始错误
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("数据库操作失败: %s操作在表'%s'上失败", e.Operation, e.Table)
}

// Unwrap 实现错误解包，支持 errors.Is 和 errors.As
func (e *DatabaseError) Unwrap() error {
	return e.Cause
}

// =============================================================================
// 3. 错误类型断言和类型检查
// =============================================================================

// ProcessingError 处理错误的基础类型
type ProcessingError interface {
	error
	IsRetryable() bool
	GetSeverity() string
}

// NetworkError 网络错误
type NetworkError struct {
	URL     string
	Timeout bool
	Code    int
}

func (e *NetworkError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("网络超时: %s", e.URL)
	}
	return fmt.Sprintf("网络错误 %d: %s", e.Code, e.URL)
}

func (e *NetworkError) IsRetryable() bool {
	return e.Timeout || (e.Code >= 500 && e.Code < 600)
}

func (e *NetworkError) GetSeverity() string {
	if e.Code >= 500 {
		return "严重"
	}
	return "警告"
}

// BusinessError 业务逻辑错误
type BusinessError struct {
	Code        string
	Description string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("业务错误 %s: %s", e.Code, e.Description)
}

func (e *BusinessError) IsRetryable() bool {
	return false // 业务错误通常不可重试
}

func (e *BusinessError) GetSeverity() string {
	return "业务"
}

// =============================================================================
// 4. 错误聚合和多错误处理
// =============================================================================

// MultiError 多错误聚合类型
type MultiError struct {
	Errors []error
}

func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "无错误"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}

	result := fmt.Sprintf("发生了 %d 个错误:\n", len(m.Errors))
	for i, err := range m.Errors {
		result += fmt.Sprintf("  %d. %s\n", i+1, err.Error())
	}
	return result
}

// Add 添加错误到聚合中
func (m *MultiError) Add(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// HasErrors 检查是否有错误
func (m *MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

// ToError 转换为标准 error 接口
func (m *MultiError) ToError() error {
	if !m.HasErrors() {
		return nil
	}
	return m
}

// =============================================================================
// 5. 错误工厂和构建器模式
// =============================================================================

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	errorType string
	code      string
	message   string
	cause     error
	context   map[string]interface{}
}

// NewErrorBuilder 创建错误构建器
func NewErrorBuilder() *ErrorBuilder {
	return &ErrorBuilder{
		context: make(map[string]interface{}),
	}
}

// WithType 设置错误类型
func (b *ErrorBuilder) WithType(errorType string) *ErrorBuilder {
	b.errorType = errorType
	return b
}

// WithCode 设置错误代码
func (b *ErrorBuilder) WithCode(code string) *ErrorBuilder {
	b.code = code
	return b
}

// WithMessage 设置错误消息
func (b *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
	b.message = message
	return b
}

// WithCause 设置原因错误
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	b.cause = cause
	return b
}

// WithContext 添加上下文信息
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.context[key] = value
	return b
}

// Build 构建最终错误
func (b *ErrorBuilder) Build() error {
	if b.errorType == "business" {
		return &BusinessError{
			Code:        b.code,
			Description: b.message,
		}
	}

	if b.errorType == "validation" {
		if field, ok := b.context["field"]; ok {
			if value, ok := b.context["value"]; ok {
				return &ValidationError{
					Field: field.(string),
					Value: value,
					Rule:  b.message,
				}
			}
		}
	}

	// 默认返回基础自定义错误
	code := 0
	if b.code != "" {
		if c, err := strconv.Atoi(b.code); err == nil {
			code = c
		}
	}

	return NewMyError(code, b.message)
}

// =============================================================================
// 6. 错误处理中间件和装饰器
// =============================================================================

// ErrorHandler 错误处理函数类型
type ErrorHandler func(error) error

// RecoveryHandler 恢复处理中间件
func RecoveryHandler(next func() error) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("从 panic 中恢复: %v\n", r)
		}
	}()

	return next()
}

// LoggingHandler 日志记录中间件
func LoggingHandler(next func() error) error {
	err := next()
	if err != nil {
		fmt.Printf("错误日志: %s [时间: %v]\n", err.Error(), time.Now())
	}
	return err
}

// RetryHandler 重试中间件
func RetryHandler(maxRetries int, next func() error) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			fmt.Printf("第 %d 次重试...\n", i)
			time.Sleep(time.Millisecond * 100) // 简单的退避策略
		}

		lastErr = next()
		if lastErr == nil {
			return nil
		}

		// 检查是否可重试
		if pe, ok := lastErr.(ProcessingError); ok && !pe.IsRetryable() {
			break
		}
	}

	return lastErr
}

// =============================================================================
// 7. 实际应用示例
// =============================================================================

// User 用户模型
type User struct {
	ID    int
	Name  string
	Email string
	Age   int
}

// UserService 用户服务
type UserService struct{}

// ValidateUser 验证用户数据
func (s *UserService) ValidateUser(user *User) error {
	multiErr := &MultiError{}

	if user.Name == "" {
		multiErr.Add(&ValidationError{
			Field: "name",
			Value: user.Name,
			Rule:  "不能为空",
		})
	}

	if user.Age < 0 || user.Age > 150 {
		multiErr.Add(&ValidationError{
			Field: "age",
			Value: user.Age,
			Rule:  "必须在 0-150 之间",
		})
	}

	if user.Email == "" {
		multiErr.Add(&ValidationError{
			Field: "email",
			Value: user.Email,
			Rule:  "不能为空",
		})
	}

	return multiErr.ToError()
}

// CreateUser 创建用户（模拟数据库操作）
func (s *UserService) CreateUser(user *User) error {
	// 首先验证
	if err := s.ValidateUser(user); err != nil {
		return err
	}

	// 模拟数据库操作失败
	if user.ID == 999 {
		return &DatabaseError{
			Operation: "INSERT",
			Table:     "users",
			Cause:     fmt.Errorf("主键冲突"),
		}
	}

	// 模拟业务逻辑错误
	if user.Name == "admin" {
		return &BusinessError{
			Code:        "USER_001",
			Description: "用户名 'admin' 是系统保留名称",
		}
	}

	fmt.Printf("用户创建成功: %+v\n", user)
	return nil
}

// =============================================================================
// 8. 演示函数
// =============================================================================

func demonstrateBasicCustomErrors() {
	fmt.Println("=== 1. 基础自定义错误 ===")

	err1 := NewMyError(404, "用户未找到")
	fmt.Println("自定义错误:", err1)

	// 类型断言获取详细信息
	var e error = err1
	if myErr, ok := e.(*MyError); ok {
		fmt.Printf("错误代码: %d, 消息: %s\n", myErr.Code, myErr.Message)
	}

	fmt.Println()
}

func demonstrateErrorWrapping() {
	fmt.Println("=== 2. 错误包装和链 ===")

	// 创建原始错误
	originalErr := fmt.Errorf("连接超时")

	// 包装错误
	dbErr := &DatabaseError{
		Operation: "SELECT",
		Table:     "users",
		Cause:     originalErr,
	}

	fmt.Println("包装错误:", dbErr)
	fmt.Println("原始错误:", dbErr.Unwrap())

	fmt.Println()
}

func demonstrateErrorTypeChecking() {
	fmt.Println("=== 3. 错误类型检查 ===")

	errors := []error{
		&NetworkError{URL: "https://api.example.com", Timeout: true},
		&NetworkError{URL: "https://api.example.com", Code: 500},
		&BusinessError{Code: "BIZ_001", Description: "余额不足"},
	}

	for i, err := range errors {
		fmt.Printf("错误 %d: %s\n", i+1, err)

		if pe, ok := err.(ProcessingError); ok {
			fmt.Printf("  可重试: %v, 严重程度: %s\n", pe.IsRetryable(), pe.GetSeverity())
		}
	}

	fmt.Println()
}

func demonstrateMultiErrors() {
	fmt.Println("=== 4. 多错误聚合 ===")

	service := &UserService{}

	// 测试验证错误
	invalidUser := &User{
		ID:    1,
		Name:  "", // 空名称
		Email: "", // 空邮箱
		Age:   -5, // 无效年龄
	}

	if err := service.CreateUser(invalidUser); err != nil {
		fmt.Println("验证失败:")
		fmt.Println(err)
	}

	fmt.Println()
}

func demonstrateErrorBuilder() {
	fmt.Println("=== 5. 错误构建器 ===")

	// 使用构建器创建业务错误
	err1 := NewErrorBuilder().
		WithType("business").
		WithCode("PAY_001").
		WithMessage("支付金额超过限制").
		Build()

	fmt.Println("业务错误:", err1)

	// 使用构建器创建验证错误
	err2 := NewErrorBuilder().
		WithType("validation").
		WithMessage("必须是正数").
		WithContext("field", "amount").
		WithContext("value", -100).
		Build()

	fmt.Println("验证错误:", err2)

	fmt.Println()
}

func demonstrateErrorMiddleware() {
	fmt.Println("=== 6. 错误处理中间件 ===")

	// 可能失败的操作
	riskyOperation := func() error {
		return &NetworkError{
			URL:     "https://api.example.com",
			Timeout: true,
		}
	}

	// 应用中间件
	fmt.Println("使用重试和日志中间件:")
	err := LoggingHandler(func() error {
		return RetryHandler(2, riskyOperation)
	})

	if err != nil {
		fmt.Println("最终错误:", err)
	}

	fmt.Println()
}

func demonstrateRealWorldScenario() {
	fmt.Println("=== 7. 实际应用场景 ===")

	service := &UserService{}

	// 测试用例
	testCases := []*User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", Age: 25},
		{ID: 999, Name: "李四", Email: "lisi@example.com", Age: 30},   // 会触发数据库错误
		{ID: 2, Name: "admin", Email: "admin@example.com", Age: 35}, // 会触发业务错误
	}

	for i, user := range testCases {
		fmt.Printf("测试用户 %d: %s\n", i+1, user.Name)

		if err := service.CreateUser(user); err != nil {
			// 根据错误类型进行不同处理
			switch e := err.(type) {
			case *MultiError:
				fmt.Println("  验证错误:")
				for _, validationErr := range e.Errors {
					fmt.Printf("    - %s\n", validationErr)
				}
			case *DatabaseError:
				fmt.Printf("  数据库错误: %s (原因: %v)\n", e.Error(), e.Unwrap())
			case *BusinessError:
				fmt.Printf("  业务错误: %s\n", e.Error())
			default:
				fmt.Printf("  未知错误: %s\n", err)
			}
		}
		fmt.Println()
	}
}

func demonstrateErrorRecovery() {
	fmt.Println("=== 8. 错误恢复模式 ===")

	// 模拟可能panic的操作
	panicOperation := func() error {
		panic("模拟系统崩溃")
	}

	// 使用恢复中间件
	fmt.Println("使用恢复中间件:")
	err := RecoveryHandler(func() error {
		return LoggingHandler(panicOperation)
	})

	if err != nil {
		fmt.Println("操作失败:", err)
	} else {
		fmt.Println("操作成功完成")
	}

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 自定义错误处理 - 完整示例")
	fmt.Println("=================================")

	demonstrateBasicCustomErrors()
	demonstrateErrorWrapping()
	demonstrateErrorTypeChecking()
	demonstrateMultiErrors()
	demonstrateErrorBuilder()
	demonstrateErrorMiddleware()
	demonstrateRealWorldScenario()
	demonstrateErrorRecovery()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 创建一个自定义错误类型 `ConfigError`，包含配置文件路径和错误详情")
	fmt.Println("2. 实现一个错误聚合器，可以按错误类型分组")
	fmt.Println("3. 创建一个支持错误重试的装饰器，包含指数退避策略")
	fmt.Println("4. 实现一个错误转换器，将第三方库错误转换为自定义错误")
	fmt.Println("5. 设计一个错误报告系统，可以收集和分析错误统计信息")
	fmt.Println("\n请在此文件中实现这些练习，加深对 Go 自定义错误处理的理解！")
}
