package main

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// =============================================================================
// 1. 单元测试基础
// =============================================================================

/*
Go 语言内置了测试框架，通过 testing 包提供测试功能。

测试文件命名规则：
- 测试文件名以 _test.go 结尾
- 测试函数名以 Test 开头
- 测试函数签名：func TestXxx(t *testing.T)

运行测试：
go test                    // 运行当前包的所有测试
go test -v                 // 详细输出
go test -run TestName      // 运行特定测试
go test ./...              // 运行所有子包测试
go test -cover             // 显示代码覆盖率

测试的好处：
1. 确保代码质量
2. 防止回归错误
3. 文档化代码行为
4. 重构时的安全网
5. 提高开发效率
*/

// Calculator 计算器结构体（被测试的代码）
type Calculator struct {
	memory float64
}

// Add 加法操作
func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.memory = result
	return result
}

// Subtract 减法操作
func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.memory = result
	return result
}

// Multiply 乘法操作
func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.memory = result
	return result
}

// Divide 除法操作
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	result := a / b
	c.memory = result
	return result, nil
}

// GetMemory 获取内存值
func (c *Calculator) GetMemory() float64 {
	return c.memory
}

// ClearMemory 清空内存
func (c *Calculator) ClearMemory() {
	c.memory = 0
}

// =============================================================================
// 2. 字符串处理工具（被测试的代码）
// =============================================================================

// StringProcessor 字符串处理器
type StringProcessor struct{}

// IsPalindrome 检查是否为回文
func (sp *StringProcessor) IsPalindrome(s string) bool {
	s = strings.ToLower(strings.ReplaceAll(s, " ", ""))
	for i := 0; i < len(s)/2; i++ {
		if s[i] != s[len(s)-1-i] {
			return false
		}
	}
	return true
}

// WordCount 统计单词数量
func (sp *StringProcessor) WordCount(s string) int {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	return len(strings.Fields(s))
}

// Reverse 反转字符串
func (sp *StringProcessor) Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ExtractNumbers 提取字符串中的数字
func (sp *StringProcessor) ExtractNumbers(s string) []int {
	var numbers []int
	var current strings.Builder

	for _, r := range s {
		if r >= '0' && r <= '9' {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				if num := parseNumber(current.String()); num != 0 {
					numbers = append(numbers, num)
				}
				current.Reset()
			}
		}
	}

	if current.Len() > 0 {
		if num := parseNumber(current.String()); num != 0 {
			numbers = append(numbers, num)
		}
	}

	return numbers
}

// parseNumber 辅助函数：解析数字
func parseNumber(s string) int {
	num := 0
	for _, r := range s {
		num = num*10 + int(r-'0')
	}
	return num
}

// =============================================================================
// 3. 用户管理系统（被测试的代码）
// =============================================================================

// User 用户结构体
type User struct {
	ID       int
	Username string
	Email    string
	Age      int
	IsActive bool
}

// UserManager 用户管理器
type UserManager struct {
	users  map[int]*User
	nextID int
}

// NewUserManager 创建新的用户管理器
func NewUserManager() *UserManager {
	return &UserManager{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// CreateUser 创建用户
func (um *UserManager) CreateUser(username, email string, age int) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if age < 0 || age > 150 {
		return nil, errors.New("age must be between 0 and 150")
	}

	// 检查用户名是否已存在
	for _, user := range um.users {
		if user.Username == username {
			return nil, errors.New("username already exists")
		}
	}

	user := &User{
		ID:       um.nextID,
		Username: username,
		Email:    email,
		Age:      age,
		IsActive: true,
	}

	um.users[um.nextID] = user
	um.nextID++

	return user, nil
}

// GetUser 获取用户
func (um *UserManager) GetUser(id int) (*User, error) {
	user, exists := um.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// UpdateUser 更新用户
func (um *UserManager) UpdateUser(id int, username, email string, age int) error {
	user, exists := um.users[id]
	if !exists {
		return errors.New("user not found")
	}

	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	if age >= 0 && age <= 150 {
		user.Age = age
	}

	return nil
}

// DeleteUser 删除用户
func (um *UserManager) DeleteUser(id int) error {
	if _, exists := um.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(um.users, id)
	return nil
}

// GetAllUsers 获取所有用户
func (um *UserManager) GetAllUsers() []*User {
	users := make([]*User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}
	return users
}

// GetActiveUsers 获取活跃用户
func (um *UserManager) GetActiveUsers() []*User {
	var activeUsers []*User
	for _, user := range um.users {
		if user.IsActive {
			activeUsers = append(activeUsers, user)
		}
	}
	return activeUsers
}

// DeactivateUser 停用用户
func (um *UserManager) DeactivateUser(id int) error {
	user, exists := um.users[id]
	if !exists {
		return errors.New("user not found")
	}
	user.IsActive = false
	return nil
}

// =============================================================================
// 4. 数学工具函数（被测试的代码）
// =============================================================================

// MathUtils 数学工具集合
type MathUtils struct{}

// IsPrime 检查是否为质数
func (mu *MathUtils) IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}

	sqrt := int(math.Sqrt(float64(n)))
	for i := 3; i <= sqrt; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// Factorial 计算阶乘
func (mu *MathUtils) Factorial(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("factorial not defined for negative numbers")
	}
	if n == 0 || n == 1 {
		return 1, nil
	}

	result := 1
	for i := 2; i <= n; i++ {
		// 检查溢出
		if result > math.MaxInt/i {
			return 0, errors.New("factorial overflow")
		}
		result *= i
	}
	return result, nil
}

// GCD 计算最大公约数
func (mu *MathUtils) GCD(a, b int) int {
	a, b = abs(a), abs(b)
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM 计算最小公倍数
func (mu *MathUtils) LCM(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return abs(a*b) / mu.GCD(a, b)
}

// abs 求绝对值
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// FibonacciSequence 生成斐波那契序列
func (mu *MathUtils) FibonacciSequence(n int) []int {
	if n <= 0 {
		return []int{}
	}
	if n == 1 {
		return []int{0}
	}
	if n == 2 {
		return []int{0, 1}
	}

	fib := make([]int, n)
	fib[0], fib[1] = 0, 1

	for i := 2; i < n; i++ {
		fib[i] = fib[i-1] + fib[i-2]
	}

	return fib
}

// =============================================================================
// 5. 性能测试和并发测试示例
// =============================================================================

// ConcurrentCounter 并发安全的计数器（需要测试并发安全性）
type ConcurrentCounter struct {
	count int
}

// Increment 递增（非并发安全）
func (cc *ConcurrentCounter) Increment() {
	cc.count++
}

// GetCount 获取计数
func (cc *ConcurrentCounter) GetCount() int {
	return cc.count
}

// Reset 重置计数器
func (cc *ConcurrentCounter) Reset() {
	cc.count = 0
}

// =============================================================================
// 6. HTTP 客户端和模拟测试示例
// =============================================================================

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Get(url string) (string, error)
	Post(url, data string) (string, error)
}

// RealHTTPClient 真实的HTTP客户端（生产环境使用）
type RealHTTPClient struct{}

// Get 发送GET请求
func (c *RealHTTPClient) Get(url string) (string, error) {
	// 这里应该是真实的HTTP请求
	// 为了演示，返回模拟响应
	time.Sleep(100 * time.Millisecond) // 模拟网络延迟
	return fmt.Sprintf("GET response from %s", url), nil
}

// Post 发送POST请求
func (c *RealHTTPClient) Post(url, data string) (string, error) {
	// 这里应该是真实的HTTP请求
	// 为了演示，返回模拟响应
	time.Sleep(100 * time.Millisecond) // 模拟网络延迟
	return fmt.Sprintf("POST response from %s with data: %s", url, data), nil
}

// APIService 使用HTTP客户端的服务
type APIService struct {
	client HTTPClient
}

// NewAPIService 创建API服务
func NewAPIService(client HTTPClient) *APIService {
	return &APIService{client: client}
}

// GetUserInfo 获取用户信息
func (api *APIService) GetUserInfo(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("user ID cannot be empty")
	}

	url := fmt.Sprintf("https://api.example.com/users/%s", userID)
	response, err := api.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	return response, nil
}

// CreateUser 创建用户
func (api *APIService) CreateUser(userData string) (string, error) {
	if userData == "" {
		return "", errors.New("user data cannot be empty")
	}

	url := "https://api.example.com/users"
	response, err := api.client.Post(url, userData)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return response, nil
}

// =============================================================================
// 7. 演示函数
// =============================================================================

func demonstrateTestingConcepts() {
	fmt.Println("=== 1. 测试概念和工具演示 ===")

	// 演示计算器
	calc := &Calculator{}
	result := calc.Add(10, 5)
	fmt.Printf("计算器测试：10 + 5 = %.2f\n", result)

	divResult, err := calc.Divide(10, 2)
	if err != nil {
		fmt.Printf("除法错误：%s\n", err)
	} else {
		fmt.Printf("计算器测试：10 ÷ 2 = %.2f\n", divResult)
	}

	// 演示字符串处理
	sp := &StringProcessor{}
	fmt.Printf("回文检测：'level' 是回文吗？%v\n", sp.IsPalindrome("level"))
	fmt.Printf("单词统计：'Hello world test' 有 %d 个单词\n", sp.WordCount("Hello world test"))
	fmt.Printf("字符串反转：'Go' 反转为 '%s'\n", sp.Reverse("Go"))

	fmt.Println()
}

func demonstrateUserManagement() {
	fmt.Println("=== 2. 用户管理系统演示 ===")

	um := NewUserManager()

	// 创建用户
	user1, err := um.CreateUser("张三", "zhangsan@example.com", 25)
	if err != nil {
		fmt.Printf("创建用户失败：%s\n", err)
	} else {
		fmt.Printf("创建用户成功：%+v\n", user1)
	}

	// 获取用户
	retrievedUser, err := um.GetUser(user1.ID)
	if err != nil {
		fmt.Printf("获取用户失败：%s\n", err)
	} else {
		fmt.Printf("获取用户成功：%+v\n", retrievedUser)
	}

	// 更新用户
	err = um.UpdateUser(user1.ID, "李四", "", 30)
	if err != nil {
		fmt.Printf("更新用户失败：%s\n", err)
	} else {
		fmt.Println("用户更新成功")
	}

	fmt.Println()
}

func demonstrateMathUtils() {
	fmt.Println("=== 3. 数学工具演示 ===")

	mu := &MathUtils{}

	// 质数检测
	numbers := []int{2, 3, 4, 5, 17, 18, 29, 30}
	for _, n := range numbers {
		fmt.Printf("%d 是质数吗？%v\n", n, mu.IsPrime(n))
	}

	// 阶乘计算
	fact, err := mu.Factorial(5)
	if err != nil {
		fmt.Printf("阶乘计算错误：%s\n", err)
	} else {
		fmt.Printf("5! = %d\n", fact)
	}

	// 最大公约数和最小公倍数
	fmt.Printf("GCD(12, 18) = %d\n", mu.GCD(12, 18))
	fmt.Printf("LCM(12, 18) = %d\n", mu.LCM(12, 18))

	// 斐波那契序列
	fib := mu.FibonacciSequence(10)
	fmt.Printf("前10项斐波那契数列：%v\n", fib)

	fmt.Println()
}

func demonstrateAPIService() {
	fmt.Println("=== 4. API 服务演示 ===")

	// 使用真实的HTTP客户端
	realClient := &RealHTTPClient{}
	apiService := NewAPIService(realClient)

	// 获取用户信息
	userInfo, err := apiService.GetUserInfo("123")
	if err != nil {
		fmt.Printf("获取用户信息失败：%s\n", err)
	} else {
		fmt.Printf("用户信息：%s\n", userInfo)
	}

	// 创建用户
	userData := `{"name": "测试用户", "email": "test@example.com"}`
	createResponse, err := apiService.CreateUser(userData)
	if err != nil {
		fmt.Printf("创建用户失败：%s\n", err)
	} else {
		fmt.Printf("创建用户响应：%s\n", createResponse)
	}

	fmt.Println()
}

func demonstrateTestingBestPractices() {
	fmt.Println("=== 5. 测试最佳实践 ===")
	fmt.Println("1. 测试命名：使用描述性的测试名称")
	fmt.Println("2. AAA 模式：Arrange（准备）、Act（执行）、Assert（断言）")
	fmt.Println("3. 测试独立性：每个测试都应该独立运行")
	fmt.Println("4. 测试覆盖率：追求高覆盖率，但不要盲目追求100%")
	fmt.Println("5. 模拟和桩：使用模拟对象测试外部依赖")
	fmt.Println("6. 表驱动测试：使用表格驱动测试多种情况")
	fmt.Println("7. 基准测试：测试性能关键代码")
	fmt.Println("8. 集成测试：测试组件间的交互")
	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 测试和调试 - 完整示例")
	fmt.Println("===========================")

	demonstrateTestingConcepts()
	demonstrateUserManagement()
	demonstrateMathUtils()
	demonstrateAPIService()
	demonstrateTestingBestPractices()

	fmt.Println("=== 测试文件示例 ===")
	fmt.Println("创建对应的 *_test.go 文件来测试以上功能：")
	fmt.Println("1. calculator_test.go - 测试计算器功能")
	fmt.Println("2. string_processor_test.go - 测试字符串处理")
	fmt.Println("3. user_manager_test.go - 测试用户管理")
	fmt.Println("4. math_utils_test.go - 测试数学工具")
	fmt.Println("5. api_service_test.go - 测试API服务（使用模拟）")

	fmt.Println("\n=== 运行测试命令 ===")
	fmt.Println("go test                    // 运行所有测试")
	fmt.Println("go test -v                 // 详细输出")
	fmt.Println("go test -cover             // 显示覆盖率")
	fmt.Println("go test -bench=.           // 运行基准测试")
	fmt.Println("go test -race              // 检测竞态条件")

	fmt.Println("\n=== 练习任务 ===")
	fmt.Println("1. 为计算器类创建完整的测试套件")
	fmt.Println("2. 编写表驱动测试来测试字符串处理函数")
	fmt.Println("3. 创建用户管理器的集成测试")
	fmt.Println("4. 实现API服务的模拟测试")
	fmt.Println("5. 编写并发计数器的竞态条件测试")
	fmt.Println("6. 创建数学工具的基准测试")
	fmt.Println("\n在对应的 *_test.go 文件中实现这些测试！")
}
