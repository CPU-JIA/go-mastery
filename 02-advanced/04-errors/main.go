package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
=== Go语言进阶特性第四课：错误处理(Error Handling) ===

学习目标：
1. 理解Go的错误处理哲学
2. 掌握error接口的使用
3. 学会错误的创建和处理
4. 了解错误传播和包装
5. 掌握最佳实践模式

Go错误处理特点：
- 显式错误处理，不使用异常
- error是内建接口类型
- 错误是值，可以传递和检查
- 支持错误包装和链式错误
- 鼓励早期错误检查
*/

func main() {
	fmt.Println("=== Go语言错误处理学习 ===")

	// 1. 基本错误处理
	demonstrateBasicErrorHandling()

	// 2. 错误创建方式
	demonstrateErrorCreation()

	// 3. 错误检查模式
	demonstrateErrorCheckingPatterns()

	// 4. 错误传播和包装
	demonstrateErrorPropagation()

	// 5. 错误类型和断言
	demonstrateErrorTypes()

	// 6. 错误处理最佳实践
	demonstrateBestPractices()

	// 7. 结构化错误处理
	demonstrateStructuredErrorHandling()

	// 8. 实际应用示例
	demonstratePracticalExamples()
}

// 基本错误处理
func demonstrateBasicErrorHandling() {
	fmt.Println("1. 基本错误处理:")

	// 简单的错误返回
	result, err := divide(10, 2)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	} else {
		fmt.Printf("结果: %.2f\n", result)
	}

	// 除零错误
	result, err = divide(10, 0)
	if err != nil {
		fmt.Printf("除零错误: %v\n", err)
	} else {
		fmt.Printf("结果: %.2f\n", result)
	}

	// 多个可能的错误
	fmt.Println("\n字符串转整数:")
	numbers := []string{"42", "3.14", "abc", "100"}

	for _, numStr := range numbers {
		if num, err := strconv.Atoi(numStr); err != nil {
			fmt.Printf("'%s' 转换失败: %v\n", numStr, err)
		} else {
			fmt.Printf("'%s' 转换成功: %d\n", numStr, num)
		}
	}

	// 文件操作错误
	fmt.Println("\n文件操作:")
	if content, err := readFile("existing.txt"); err != nil {
		fmt.Printf("读取文件错误: %v\n", err)
	} else {
		fmt.Printf("文件内容: %s\n", content)
	}

	if content, err := readFile("nonexistent.txt"); err != nil {
		fmt.Printf("文件不存在: %v\n", err)
	} else {
		fmt.Printf("文件内容: %s\n", content)
	}

	fmt.Println()
}

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为零")
	}
	return a / b, nil
}

func readFile(filename string) (string, error) {
	// 模拟文件读取
	if filename == "existing.txt" {
		return "文件内容示例", nil
	}
	return "", fmt.Errorf("文件 '%s' 不存在", filename)
}

// 错误创建方式
func demonstrateErrorCreation() {
	fmt.Println("2. 错误创建方式:")

	// 1. errors.New()
	err1 := errors.New("这是一个简单错误")
	fmt.Printf("errors.New: %v\n", err1)

	// 2. fmt.Errorf()
	user := "张三"
	err2 := fmt.Errorf("用户 %s 不存在", user)
	fmt.Printf("fmt.Errorf: %v\n", err2)

	// 3. 实现error接口
	err3 := &CustomError{
		Code:    404,
		Message: "资源未找到",
		Time:    time.Now(),
	}
	fmt.Printf("自定义错误: %v\n", err3)

	// 4. 错误常量
	fmt.Printf("预定义错误: %v\n", ErrInvalidInput)

	// 5. 错误变量
	fmt.Printf("错误变量: %v\n", ErrDatabaseConnection)

	// 6. 条件错误创建
	amount := -100.0
	if err := validateAmount(amount); err != nil {
		fmt.Printf("验证错误: %v\n", err)
	}

	amount = 50.0
	if err := validateAmount(amount); err != nil {
		fmt.Printf("验证错误: %v\n", err)
	} else {
		fmt.Printf("金额 %.2f 验证通过\n", amount)
	}

	fmt.Println()
}

// 自定义错误类型
type CustomError struct {
	Code    int
	Message string
	Time    time.Time
}

func (ce *CustomError) Error() string {
	return fmt.Sprintf("[%d] %s (发生时间: %s)",
		ce.Code, ce.Message, ce.Time.Format("15:04:05"))
}

// 预定义错误
var (
	ErrInvalidInput       = errors.New("输入无效")
	ErrDatabaseConnection = errors.New("数据库连接失败")
	ErrPermissionDenied   = errors.New("权限不足")
	ErrResourceNotFound   = errors.New("资源未找到")
)

func validateAmount(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("金额不能为负数: %.2f", amount)
	}
	if amount > 10000 {
		return fmt.Errorf("金额超过限制: %.2f > 10000", amount)
	}
	return nil
}

// 错误检查模式
func demonstrateErrorCheckingPatterns() {
	fmt.Println("3. 错误检查模式:")

	// 1. 立即检查模式
	fmt.Println("立即检查模式:")
	if result, err := processData("valid_data"); err != nil {
		fmt.Printf("处理失败: %v\n", err)
	} else {
		fmt.Printf("处理成功: %s\n", result)
	}

	// 2. 延迟检查模式
	fmt.Println("\n延迟检查模式:")
	var errors []error
	results := make([]string, 0)

	if result, err := processData("data1"); err != nil {
		errors = append(errors, err)
	} else {
		results = append(results, result)
	}

	if result, err := processData(""); err != nil {
		errors = append(errors, err)
	} else {
		results = append(results, result)
	}

	if result, err := processData("data3"); err != nil {
		errors = append(errors, err)
	} else {
		results = append(results, result)
	}

	fmt.Printf("成功结果: %v\n", results)
	fmt.Printf("错误列表: %v\n", errors)

	// 3. 错误聚合
	fmt.Println("\n错误聚合:")
	if err := performBatchOperation(); err != nil {
		fmt.Printf("批量操作失败: %v\n", err)
	} else {
		fmt.Println("批量操作成功")
	}

	// 4. 条件错误检查
	fmt.Println("\n条件错误检查:")
	users := []string{"admin", "guest", "user", ""}
	for _, user := range users {
		if err := checkUserPermission(user); err != nil {
			fmt.Printf("用户 '%s': %v\n", user, err)
		} else {
			fmt.Printf("用户 '%s': 权限检查通过\n", user)
		}
	}

	fmt.Println()
}

func processData(data string) (string, error) {
	if data == "" {
		return "", errors.New("数据不能为空")
	}
	if strings.Contains(data, "invalid") {
		return "", fmt.Errorf("数据包含无效内容: %s", data)
	}
	return fmt.Sprintf("处理后的%s", data), nil
}

func performBatchOperation() error {
	var batchErrs []string

	// 模拟多个操作
	operations := []func() error{
		func() error { return nil }, // 成功
		func() error { return errors.New("操作2失败") },
		func() error { return nil }, // 成功
		func() error { return errors.New("操作4失败") },
	}

	for i, op := range operations {
		if err := op(); err != nil {
			batchErrs = append(batchErrs, fmt.Sprintf("操作%d: %v", i+1, err))
		}
	}

	if len(batchErrs) > 0 {
		return fmt.Errorf("批量操作中有 %d 个失败: %s",
			len(batchErrs), strings.Join(batchErrs, "; "))
	}

	return nil
}

func checkUserPermission(user string) error {
	if user == "" {
		return ErrInvalidInput
	}
	if user == "guest" {
		return ErrPermissionDenied
	}
	return nil
}

// 错误传播和包装
func demonstrateErrorPropagation() {
	fmt.Println("4. 错误传播和包装:")

	// 1. 错误传播
	fmt.Println("错误传播:")
	if err := highlevelOperation(); err != nil {
		fmt.Printf("高级操作失败: %v\n", err)
	}

	// 2. 错误包装
	fmt.Println("\n错误包装:")
	if err := wrappedOperation(); err != nil {
		fmt.Printf("包装错误: %v\n", err)
	}

	// 3. 错误链
	fmt.Println("\n错误链:")
	if err := chainedOperation(); err != nil {
		fmt.Printf("链式错误: %v\n", err)

		// 展开错误链
		fmt.Println("错误链展开:")
		current := err
		level := 0
		for current != nil {
			fmt.Printf("  层级%d: %v\n", level, current)
			if unwrapped, ok := current.(interface{ Unwrap() error }); ok {
				current = unwrapped.Unwrap()
			} else {
				break
			}
			level++
		}
	}

	// 4. 上下文错误
	fmt.Println("\n上下文错误:")
	if err := operationWithContext("user123", "action_delete"); err != nil {
		fmt.Printf("上下文错误: %v\n", err)
	}

	fmt.Println()
}

func highlevelOperation() error {
	if err := midlevelOperation(); err != nil {
		return err // 直接传播
	}
	return nil
}

func midlevelOperation() error {
	if err := lowlevelOperation(); err != nil {
		return err // 直接传播
	}
	return nil
}

func lowlevelOperation() error {
	return errors.New("底层操作失败")
}

func wrappedOperation() error {
	if err := lowlevelOperation(); err != nil {
		return fmt.Errorf("中层包装错误: %w", err)
	}
	return nil
}

// 可展开的错误类型
type WrappedError struct {
	Message string
	Cause   error
}

func (we *WrappedError) Error() string {
	if we.Cause != nil {
		return fmt.Sprintf("%s: %v", we.Message, we.Cause)
	}
	return we.Message
}

func (we *WrappedError) Unwrap() error {
	return we.Cause
}

func chainedOperation() error {
	if err := lowlevelOperation(); err != nil {
		wrapped1 := &WrappedError{
			Message: "第一层包装",
			Cause:   err,
		}
		wrapped2 := &WrappedError{
			Message: "第二层包装",
			Cause:   wrapped1,
		}
		return &WrappedError{
			Message: "顶层包装",
			Cause:   wrapped2,
		}
	}
	return nil
}

type ContextError struct {
	Operation string
	UserID    string
	Cause     error
}

func (ce *ContextError) Error() string {
	return fmt.Sprintf("操作 '%s' 失败 (用户: %s): %v",
		ce.Operation, ce.UserID, ce.Cause)
}

func operationWithContext(userID, operation string) error {
	if err := lowlevelOperation(); err != nil {
		return &ContextError{
			Operation: operation,
			UserID:    userID,
			Cause:     err,
		}
	}
	return nil
}

// 错误类型和断言
func demonstrateErrorTypes() {
	fmt.Println("5. 错误类型和断言:")

	errors := []error{
		&CustomError{Code: 404, Message: "未找到"},
		&ValidationError{Field: "email", Value: "invalid"},
		&NetworkError{Timeout: true, Retryable: true},
		fmt.Errorf("普通格式化错误"),
		ErrPermissionDenied,
	}

	fmt.Println("错误类型检查:")
	for i, err := range errors {
		fmt.Printf("错误%d: %v\n", i+1, err)

		// 类型断言
		switch e := err.(type) {
		case *CustomError:
			fmt.Printf("  自定义错误 - 代码: %d\n", e.Code)
		case *ValidationError:
			fmt.Printf("  验证错误 - 字段: %s, 值: %s\n", e.Field, e.Value)
		case *NetworkError:
			fmt.Printf("  网络错误 - 超时: %t, 可重试: %t\n", e.Timeout, e.Retryable)
		default:
			fmt.Printf("  其他错误类型: %T\n", e)
		}

		// 接口检查
		if retryable, ok := err.(RetryableError); ok {
			fmt.Printf("  可重试: %t\n", retryable.CanRetry())
		}

		if temp, ok := err.(TemporaryError); ok {
			fmt.Printf("  临时错误: %t\n", temp.Temporary())
		}

		fmt.Println()
	}

	fmt.Println()
}

// 错误接口
type RetryableError interface {
	CanRetry() bool
}

type TemporaryError interface {
	Temporary() bool
}

// 验证错误
type ValidationError struct {
	Field string
	Value string
}

func (ve *ValidationError) Error() string {
	return fmt.Sprintf("字段 '%s' 验证失败: 值 '%s' 无效", ve.Field, ve.Value)
}

func (ve *ValidationError) CanRetry() bool {
	return false // 验证错误通常不可重试
}

// 网络错误
type NetworkError struct {
	Timeout   bool
	Retryable bool
}

func (ne *NetworkError) Error() string {
	if ne.Timeout {
		return "网络超时"
	}
	return "网络连接失败"
}

func (ne *NetworkError) CanRetry() bool {
	return ne.Retryable
}

func (ne *NetworkError) Temporary() bool {
	return ne.Timeout || ne.Retryable
}

// 错误处理最佳实践
func demonstrateBestPractices() {
	fmt.Println("6. 错误处理最佳实践:")

	// 1. 错误信息应该清晰明确
	fmt.Println("清晰的错误信息:")
	if err := processUser("", ""); err != nil {
		fmt.Printf("用户处理错误: %v\n", err)
	}

	// 2. 添加上下文信息
	fmt.Println("\n上下文信息:")
	if err := saveUserData("user123", nil); err != nil {
		fmt.Printf("保存错误: %v\n", err)
	}

	// 3. 错误分类
	fmt.Println("\n错误分类:")
	testCases := []struct {
		input string
		desc  string
	}{
		{"", "空输入"},
		{"abc", "无效格式"},
		{"user@invalid", "无效邮箱"},
		{"user@example.com", "有效邮箱"},
	}

	for _, tc := range testCases {
		if err := validateEmail(tc.input); err != nil {
			fmt.Printf("%s: %v\n", tc.desc, err)

			// 根据错误类型决定处理方式
			switch err.(type) {
			case *ValidationError:
				fmt.Println("  -> 重新输入")
			default:
				fmt.Println("  -> 系统错误，联系管理员")
			}
		} else {
			fmt.Printf("%s: 验证通过\n", tc.desc)
		}
	}

	// 4. 错误恢复
	fmt.Println("\n错误恢复:")
	if result, err := operationWithFallback("primary"); err != nil {
		fmt.Printf("主操作失败，使用备用方案: %s\n", result)
	} else {
		fmt.Printf("主操作成功: %s\n", result)
	}

	// 5. 批量操作的错误处理
	fmt.Println("\n批量操作错误处理:")
	items := []string{"item1", "", "item3", "invalid_item", "item5"}
	results, errors := processBatch(items)

	fmt.Printf("成功处理: %d 项\n", len(results))
	fmt.Printf("失败: %d 项\n", len(errors))

	for i, err := range errors {
		fmt.Printf("  错误%d: %v\n", i+1, err)
	}

	fmt.Println()
}

func processUser(name, email string) error {
	if name == "" {
		return fmt.Errorf("processUser: 用户名不能为空")
	}
	if email == "" {
		return fmt.Errorf("processUser: 邮箱不能为空 (用户: %s)", name)
	}
	return nil
}

func saveUserData(userID string, data map[string]interface{}) error {
	if data == nil {
		return fmt.Errorf("saveUserData: 用户 %s 的数据为空", userID)
	}
	// 模拟数据库错误
	return fmt.Errorf("saveUserData: 保存用户 %s 数据失败: %w", userID, ErrDatabaseConnection)
}

func validateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Value: "(空)"}
	}
	if !strings.Contains(email, "@") {
		return &ValidationError{Field: "email", Value: email}
	}
	if strings.Contains(email, "invalid") {
		return &ValidationError{Field: "email", Value: email}
	}
	return nil
}

func operationWithFallback(operation string) (string, error) {
	// 尝试主操作
	if operation == "primary" {
		// 模拟主操作失败
		if result, err := primaryOperation(); err != nil {
			// 尝试备用操作
			if fallbackResult, fallbackErr := fallbackOperation(); fallbackErr != nil {
				return "", fmt.Errorf("主操作和备用操作都失败: 主错误: %w", err)
			} else {
				return fallbackResult, fmt.Errorf("主操作失败，使用备用操作: %w", err)
			}
		} else {
			return result, nil
		}
	}
	return "", errors.New("未知操作")
}

func primaryOperation() (string, error) {
	return "", errors.New("主操作失败")
}

func fallbackOperation() (string, error) {
	return "备用操作结果", nil
}

func processBatch(items []string) ([]string, []error) {
	var results []string
	var errors []error

	for i, item := range items {
		if result, err := processItem(item); err != nil {
			errors = append(errors, fmt.Errorf("项目%d (%s): %w", i+1, item, err))
		} else {
			results = append(results, result)
		}
	}

	return results, errors
}

func processItem(item string) (string, error) {
	if item == "" {
		return "", errors.New("项目不能为空")
	}
	if strings.Contains(item, "invalid") {
		return "", errors.New("项目包含无效内容")
	}
	return fmt.Sprintf("processed_%s", item), nil
}

// 结构化错误处理
func demonstrateStructuredErrorHandling() {
	fmt.Println("7. 结构化错误处理:")

	// 1. 错误记录器
	fmt.Println("错误记录器:")
	logger := &ErrorLogger{}

	operations := []func() error{
		func() error { return nil },
		func() error { return errors.New("操作1失败") },
		func() error { return &NetworkError{Timeout: true} },
		func() error { return &ValidationError{Field: "name", Value: ""} },
	}

	for i, op := range operations {
		if err := op(); err != nil {
			logger.LogError(fmt.Sprintf("操作%d", i+1), err)
		} else {
			fmt.Printf("操作%d 成功\n", i+1)
		}
	}

	// 2. 错误收集器
	fmt.Println("\n错误收集器:")
	collector := &ErrorCollector{}

	collector.Add("步骤1", nil)
	collector.Add("步骤2", errors.New("步骤2失败"))
	collector.Add("步骤3", &ValidationError{Field: "email", Value: "invalid"})
	collector.Add("步骤4", nil)

	if collector.HasErrors() {
		fmt.Printf("收集到 %d 个错误:\n", collector.ErrorCount())
		for _, err := range collector.GetErrors() {
			fmt.Printf("  %v\n", err)
		}
	}

	// 3. 错误恢复机制
	fmt.Println("\n错误恢复机制:")
	recoverer := &ErrorRecoverer{}

	if result, err := recoverer.ExecuteWithRecovery(func() (interface{}, error) {
		return nil, &NetworkError{Timeout: true, Retryable: true}
	}); err != nil {
		fmt.Printf("最终失败: %v\n", err)
	} else {
		fmt.Printf("恢复成功: %v\n", result)
	}

	fmt.Println()
}

// 错误记录器
type ErrorLogger struct{}

func (el *ErrorLogger) LogError(operation string, err error) {
	timestamp := time.Now().Format("15:04:05")

	switch e := err.(type) {
	case *NetworkError:
		fmt.Printf("[%s] 网络错误 在 %s: %v (可重试: %t)\n",
			timestamp, operation, e, e.CanRetry())
	case *ValidationError:
		fmt.Printf("[%s] 验证错误 在 %s: %v\n",
			timestamp, operation, e)
	default:
		fmt.Printf("[%s] 一般错误 在 %s: %v\n",
			timestamp, operation, e)
	}
}

// 错误收集器
type ErrorCollector struct {
	errors []error
}

func (ec *ErrorCollector) Add(operation string, err error) {
	if err != nil {
		ec.errors = append(ec.errors, fmt.Errorf("%s: %w", operation, err))
	}
}

func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

func (ec *ErrorCollector) ErrorCount() int {
	return len(ec.errors)
}

func (ec *ErrorCollector) GetErrors() []error {
	return ec.errors
}

// 错误恢复器
type ErrorRecoverer struct{}

func (er *ErrorRecoverer) ExecuteWithRecovery(operation func() (interface{}, error)) (interface{}, error) {
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}

		// 检查是否可以重试
		if retryable, ok := err.(RetryableError); ok && retryable.CanRetry() {
			fmt.Printf("尝试%d失败，重试中: %v\n", attempt, err)
			time.Sleep(time.Millisecond * 100) // 简短等待
			continue
		}

		// 不可重试的错误
		return nil, fmt.Errorf("不可重试的错误 (尝试%d): %w", attempt, err)
	}

	return nil, fmt.Errorf("重试%d次后仍然失败", maxRetries)
}

// 实际应用示例
func demonstratePracticalExamples() {
	fmt.Println("8. 实际应用示例:")

	// 1. 文件处理系统
	fmt.Println("文件处理系统:")
	processor := &FileProcessor{}

	files := []string{"config.json", "data.csv", "", "nonexistent.txt"}
	for _, filename := range files {
		if err := processor.ProcessFile(filename); err != nil {
			fmt.Printf("文件 '%s' 处理失败: %v\n", filename, err)
		} else {
			fmt.Printf("文件 '%s' 处理成功\n", filename)
		}
	}

	// 2. 用户服务
	fmt.Println("\n用户服务:")
	userService := &UserService{}

	users := []User{
		{ID: "1", Name: "张三", Email: "zhang@example.com"},
		{ID: "", Name: "李四", Email: "li@example.com"},
		{ID: "3", Name: "", Email: "wang@example.com"},
		{ID: "4", Name: "赵六", Email: "invalid-email"},
	}

	for _, user := range users {
		if err := userService.CreateUser(user); err != nil {
			fmt.Printf("创建用户失败: %v\n", err)
		} else {
			fmt.Printf("用户 %s 创建成功\n", user.Name)
		}
	}

	// 3. 支付系统
	fmt.Println("\n支付系统:")
	paymentSystem := &PaymentSystem{}

	transactions := []Transaction{
		{ID: "1", Amount: 100.0, Currency: "USD"},
		{ID: "2", Amount: -50.0, Currency: "USD"},
		{ID: "3", Amount: 200.0, Currency: "INVALID"},
		{ID: "4", Amount: 15000.0, Currency: "USD"},
	}

	for _, tx := range transactions {
		if err := paymentSystem.ProcessPayment(tx); err != nil {
			fmt.Printf("支付 %s 失败: %v\n", tx.ID, err)
		} else {
			fmt.Printf("支付 %s 成功\n", tx.ID)
		}
	}

	// 4. API客户端
	fmt.Println("\nAPI客户端:")
	client := &APIClient{BaseURL: "https://api.example.com"}

	endpoints := []string{"/users", "/posts", "/invalid", "/timeout"}
	for _, endpoint := range endpoints {
		if data, err := client.Get(endpoint); err != nil {
			fmt.Printf("API请求 %s 失败: %v\n", endpoint, err)
		} else {
			fmt.Printf("API请求 %s 成功: %s\n", endpoint, data)
		}
	}

	fmt.Println()
}

// 文件处理器
type FileProcessor struct{}

func (fp *FileProcessor) ProcessFile(filename string) error {
	if filename == "" {
		return &ValidationError{Field: "filename", Value: "(空)"}
	}

	// 模拟文件读取
	if filename == "nonexistent.txt" {
		return &FileError{
			Filename:  filename,
			Operation: "read",
			Cause:     os.ErrNotExist,
		}
	}

	// 模拟处理
	if filename == "config.json" {
		return nil
	}

	return nil
}

type FileError struct {
	Filename  string
	Operation string
	Cause     error
}

func (fe *FileError) Error() string {
	return fmt.Sprintf("文件操作失败: %s '%s': %v",
		fe.Operation, fe.Filename, fe.Cause)
}

func (fe *FileError) Unwrap() error {
	return fe.Cause
}

// 用户服务
type User struct {
	ID    string
	Name  string
	Email string
}

type UserService struct{}

func (us *UserService) CreateUser(user User) error {
	// 验证用户数据
	if user.ID == "" {
		return &ValidationError{Field: "id", Value: "(空)"}
	}
	if user.Name == "" {
		return &ValidationError{Field: "name", Value: "(空)"}
	}
	if err := validateEmail(user.Email); err != nil {
		return fmt.Errorf("用户创建失败: %w", err)
	}

	// 模拟创建用户
	return nil
}

// 支付系统
type Transaction struct {
	ID       string
	Amount   float64
	Currency string
}

type PaymentSystem struct{}

func (ps *PaymentSystem) ProcessPayment(tx Transaction) error {
	// 验证交易
	if tx.Amount <= 0 {
		return &PaymentError{
			Type:    "validation",
			Message: fmt.Sprintf("金额必须大于0: %.2f", tx.Amount),
			TxID:    tx.ID,
		}
	}

	if tx.Amount > 10000 {
		return &PaymentError{
			Type:    "limit",
			Message: fmt.Sprintf("金额超过限制: %.2f > 10000", tx.Amount),
			TxID:    tx.ID,
		}
	}

	if tx.Currency != "USD" && tx.Currency != "EUR" {
		return &PaymentError{
			Type:    "currency",
			Message: fmt.Sprintf("不支持的货币: %s", tx.Currency),
			TxID:    tx.ID,
		}
	}

	// 模拟支付处理
	return nil
}

type PaymentError struct {
	Type    string
	Message string
	TxID    string
}

func (pe *PaymentError) Error() string {
	return fmt.Sprintf("支付错误 [%s] (交易: %s): %s",
		pe.Type, pe.TxID, pe.Message)
}

func (pe *PaymentError) CanRetry() bool {
	return pe.Type != "validation" && pe.Type != "limit"
}

// API客户端
type APIClient struct {
	BaseURL string
}

func (ac *APIClient) Get(endpoint string) (string, error) {
	url := ac.BaseURL + endpoint

	// 模拟API请求
	switch endpoint {
	case "/users":
		return `{"users": []}`, nil
	case "/posts":
		return `{"posts": []}`, nil
	case "/invalid":
		return "", &APIError{
			StatusCode: 404,
			Message:    "Not Found",
			URL:        url,
		}
	case "/timeout":
		return "", &NetworkError{Timeout: true, Retryable: true}
	default:
		return "", &APIError{
			StatusCode: 500,
			Message:    "Internal Server Error",
			URL:        url,
		}
	}
}

type APIError struct {
	StatusCode int
	Message    string
	URL        string
}

func (ae *APIError) Error() string {
	return fmt.Sprintf("API错误 %d: %s (URL: %s)",
		ae.StatusCode, ae.Message, ae.URL)
}

func (ae *APIError) CanRetry() bool {
	return ae.StatusCode >= 500 // 服务器错误可以重试
}

func (ae *APIError) Temporary() bool {
	return ae.StatusCode == 503 || ae.StatusCode == 504
}

/*
=== 练习题 ===

1. 实现一个完整的错误处理中间件系统

2. 创建一个支持错误重试和熔断的HTTP客户端

3. 设计一个分布式系统的错误传播机制

4. 实现一个错误监控和报警系统

5. 创建一个支持错误恢复的任务队列

6. 设计一个数据库操作的错误处理框架

7. 实现一个多语言的错误消息系统

运行命令：
go run main.go

高级练习：
1. 实现错误的结构化日志记录
2. 创建错误的指标收集和分析
3. 设计错误的自动化修复机制
4. 实现错误的链路追踪
5. 创建错误处理的最佳实践检查器

重要概念：
- 错误是值，不是异常
- 显式错误检查
- 错误包装和传播
- 错误分类和处理策略
- 结构化错误信息
- 错误恢复机制
*/
