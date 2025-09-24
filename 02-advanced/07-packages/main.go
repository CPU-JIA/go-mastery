package main

import (
	"fmt"
	"go-mastery/02-advanced/07-packages/calculator"
	calc "go-mastery/02-advanced/07-packages/calculator"
	"go-mastery/02-advanced/07-packages/logger"
	_ "go-mastery/02-advanced/07-packages/logger"
	"go-mastery/02-advanced/07-packages/utils"
)

// =============================================================================
// 1. 包的基本概念和结构
// =============================================================================

/*
包（Package）是 Go 语言代码组织的基本单位。每个 Go 源文件都属于某个包。

包的作用：
1. 代码组织和模块化
2. 命名空间管理
3. 访问控制（公开/私有）
4. 代码重用

包的命名规则：
1. 包名应该简短、清晰、有意义
2. 通常使用小写字母
3. 避免使用下划线或混合大小写
4. 包名应该与其功能相关

目录结构示例：
go-mastery/
└── 02-advanced/
    └── 07-packages/
        ├── main.go           (当前文件)
        ├── calculator/       (计算器包)
        │   ├── basic.go      (基础运算)
        │   ├── advanced.go   (高级运算)
        │   └── constants.go  (常量定义)
        ├── logger/           (日志包)
        │   ├── logger.go     (日志实现)
        │   └── config.go     (配置)
        └── utils/            (工具包)
            ├── string.go     (字符串工具)
            ├── file.go       (文件工具)
            └── time.go       (时间工具)
*/

// =============================================================================
// 2. 包的导入和使用
// =============================================================================

func demonstratePackageImports() {
	fmt.Println("=== 1. 包的导入和使用 ===")

	// 标准库包
	fmt.Println("使用标准库 fmt 包进行格式化输出")

	// 本地自定义包
	// 注意：这些包需要先创建对应的文件
	result := calculator.Add(10, 20)
	fmt.Printf("计算器包：10 + 20 = %d\n", result)

	// 使用日志包
	logger.Info("这是一条信息日志")
	logger.Error("这是一条错误日志")

	// 使用工具包
	reversed := utils.Reverse("Hello, Go!")
	fmt.Printf("字符串反转：%s\n", reversed)

	fmt.Println()
}

// =============================================================================
// 3. 包的可见性和访问控制
// =============================================================================

/*
Go 语言的访问控制规则：

1. 首字母大写（导出/公开）：
   - 函数名、类型名、变量名、常量名首字母大写时，可以被其他包访问
   - 例如：Add(), UserType, MaxValue

2. 首字母小写（未导出/私有）：
   - 首字母小写的标识符只能在同一个包内访问
   - 例如：add(), userType, maxValue

3. 结构体字段的可见性：
   - 字段名首字母大写：可以被其他包访问
   - 字段名首字母小写：只能在同一包内访问
*/

// PublicStruct 公开的结构体
type PublicStruct struct {
	PublicField  string // 公开字段
	privateField int    // 私有字段
}

// PublicMethod 公开的方法
func (p *PublicStruct) PublicMethod() string {
	return fmt.Sprintf("Public: %s, Private: %d", p.PublicField, p.privateField)
}

// privateMethod 私有方法
func (p *PublicStruct) privateMethod() {
	fmt.Println("这是私有方法")
}

// PublicFunction 公开函数
func PublicFunction() {
	fmt.Println("这是公开函数")
}

// privateFunction 私有函数
func privateFunction() {
	fmt.Println("这是私有函数")
}

func demonstrateVisibility() {
	fmt.Println("=== 2. 包的可见性和访问控制 ===")

	// 在同一包内，可以访问所有标识符
	ps := &PublicStruct{
		PublicField:  "公开字段",
		privateField: 42, // 同一包内可以访问私有字段
	}

	fmt.Println("公开方法调用:", ps.PublicMethod())
	ps.privateMethod() // 同一包内可以调用私有方法

	PublicFunction()  // 调用公开函数
	privateFunction() // 同一包内可以调用私有函数

	fmt.Println()
}

// =============================================================================
// 4. 包的初始化和 init 函数
// =============================================================================

/*
包的初始化顺序：
1. 包级别变量按声明顺序初始化
2. init() 函数按文件中出现的顺序执行
3. main() 函数最后执行

init 函数特点：
1. 每个包可以有多个 init 函数
2. init 函数不能被直接调用
3. init 函数在包导入时自动执行
4. 用于包的初始化工作
*/

var packageLevelVar = initPackageVar()

func initPackageVar() string {
	fmt.Println("包级别变量初始化")
	return "已初始化"
}

func init() {
	fmt.Println("第一个 init 函数执行")
}

func init() {
	fmt.Println("第二个 init 函数执行")
}

func demonstratePackageInit() {
	fmt.Println("=== 3. 包的初始化 ===")
	fmt.Printf("包级别变量的值：%s\n", packageLevelVar)
	fmt.Println()
}

// =============================================================================
// 5. 包的别名和空白导入
// =============================================================================

/*
包的导入方式在文件开头已定义：
- 包别名：calc "go-mastery/02-advanced/07-packages/calculator"
- 点导入：将包的导出标识符导入到当前包的命名空间
- 空白导入：_ "go-mastery/02-advanced/07-packages/logger"
*/

func demonstrateImportStyles() {
	fmt.Println("=== 4. 包的导入方式 ===")

	// 使用别名
	result := calc.Multiply(5, 6)
	fmt.Printf("使用别名调用：5 × 6 = %d\n", result)

	// 点导入示例（注释掉避免命名冲突）
	// reversed := Reverse("Hello") // 直接使用函数名，无需包前缀

	fmt.Println("空白导入的包已执行 init 函数")
	fmt.Println()
}

// =============================================================================
// 6. 内部包（Internal Packages）
// =============================================================================

/*
内部包是 Go 1.4 引入的特性，用于限制包的导入范围。

规则：
1. 路径中包含 "internal" 的包称为内部包
2. 内部包只能被其父目录或父目录的子目录中的包导入
3. 用于实现包的私有实现细节

示例结构：
myproject/
├── main.go
├── api/
│   ├── public.go
│   └── internal/
│       └── helper.go    (只能被 api 包或其子包导入)
└── utils/
    └── tools.go         (不能导入 api/internal/helper.go)
*/

func demonstrateInternalPackages() {
	fmt.Println("=== 5. 内部包概念 ===")
	fmt.Println("内部包用于限制包的访问范围，增强封装性")
	fmt.Println("包含 'internal' 目录的包只能被特定范围的包导入")
	fmt.Println()
}

// =============================================================================
// 7. 包的文档和注释
// =============================================================================

/*
包文档规范：

1. 包注释：
   - 在包声明前添加注释
   - 注释应该以包名开头
   - 说明包的用途和功能

2. 函数文档：
   - 在函数声明前添加注释
   - 注释应该以函数名开头
   - 说明函数的功能、参数和返回值

3. 类型文档：
   - 在类型声明前添加注释
   - 注释应该以类型名开头

4. 使用 godoc 工具生成文档：
   go doc <package>
   godoc -http=:6060
*/

// DocumentedFunction 这个函数演示了如何编写良好的文档注释。
// 参数 name 是要问候的名字，msg 是自定义消息。
// 返回格式化的问候语字符串。
//
// 示例：
//
//	greeting := DocumentedFunction("世界", "你好")
//	fmt.Println(greeting) // 输出: 你好, 世界!
func DocumentedFunction(name, msg string) string {
	return fmt.Sprintf("%s, %s!", msg, name)
}

func demonstrateDocumentation() {
	fmt.Println("=== 6. 包的文档 ===")

	greeting := DocumentedFunction("世界", "你好")
	fmt.Println("文档化函数调用:", greeting)

	fmt.Println("使用 'go doc main DocumentedFunction' 查看函数文档")
	fmt.Println("使用 'godoc -http=:6060' 启动本地文档服务器")
	fmt.Println()
}

// =============================================================================
// 8. 包的版本管理和模块
// =============================================================================

/*
Go Modules（Go 1.11+）：

1. go.mod 文件：
   - 定义模块名称和依赖
   - 管理版本信息

2. 语义版本控制：
   - v1.2.3 格式
   - 主版本.次版本.补丁版本

3. 常用命令：
   go mod init <module-name>  // 初始化模块
   go mod tidy               // 整理依赖
   go mod download           // 下载依赖
   go get <package>          // 添加依赖

4. 模块代理：
   - GOPROXY 环境变量
   - 加速依赖下载
*/

func demonstrateModules() {
	fmt.Println("=== 7. 模块和版本管理 ===")
	fmt.Println("Go Modules 是 Go 语言官方的依赖管理解决方案")
	fmt.Println("通过 go.mod 文件管理项目依赖和版本")
	fmt.Println("支持语义版本控制和模块代理")
	fmt.Println()
}

// =============================================================================
// 9. 包的设计原则和最佳实践
// =============================================================================

/*
包设计原则：

1. 单一职责原则：
   - 每个包应该有明确的单一职责
   - 避免包功能过于复杂

2. 最小暴露原则：
   - 只暴露必要的接口
   - 隐藏实现细节

3. 稳定性原则：
   - 公开接口应该保持稳定
   - 避免频繁的破坏性变更

4. 命名一致性：
   - 包名和目录名应该一致
   - 使用清晰、一致的命名规范

5. 依赖管理：
   - 避免循环依赖
   - 减少不必要的依赖

最佳实践：

1. 包结构组织：
   cmd/        // 主要应用程序
   pkg/        // 库代码
   internal/   // 私有应用程序和库代码
   api/        // API 定义
   web/        // Web 应用程序组件
   configs/    // 配置文件模板
   deployments/ // 部署配置
   test/       // 测试数据

2. 错误处理：
   - 包应该定义自己的错误类型
   - 提供有意义的错误信息

3. 接口设计：
   - 保持接口小而专注
   - 优先使用接口而非具体类型
*/

func demonstrateBestPractices() {
	fmt.Println("=== 8. 包设计原则和最佳实践 ===")
	fmt.Println("1. 单一职责：每个包专注一个功能领域")
	fmt.Println("2. 最小暴露：只暴露必要的公开接口")
	fmt.Println("3. 稳定性：保持公开 API 的稳定性")
	fmt.Println("4. 命名一致：使用清晰、一致的命名")
	fmt.Println("5. 避免循环依赖：合理设计包的依赖关系")
	fmt.Println()
}

// =============================================================================
// 10. 实际项目中的包组织示例
// =============================================================================

// UserService 用户服务接口
type UserService interface {
	CreateUser(user *User) error
	GetUser(id int) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(id int) error
}

// User 用户模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// userService 用户服务实现
type userService struct {
	// 依赖注入
	logger logger.Logger
	// repository UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService(lg logger.Logger) UserService {
	return &userService{
		logger: lg,
	}
}

func (s *userService) CreateUser(user *User) error {
	s.logger.Info(fmt.Sprintf("创建用户: %s", user.Name))
	// 实现创建用户逻辑
	return nil
}

func (s *userService) GetUser(id int) (*User, error) {
	s.logger.Info(fmt.Sprintf("获取用户: ID=%d", id))
	// 实现获取用户逻辑
	return &User{ID: id, Name: "示例用户", Email: "user@example.com"}, nil
}

func (s *userService) UpdateUser(user *User) error {
	s.logger.Info(fmt.Sprintf("更新用户: %s", user.Name))
	// 实现更新用户逻辑
	return nil
}

func (s *userService) DeleteUser(id int) error {
	s.logger.Info(fmt.Sprintf("删除用户: ID=%d", id))
	// 实现删除用户逻辑
	return nil
}

func demonstrateProjectStructure() {
	fmt.Println("=== 9. 实际项目包组织 ===")

	// 创建服务实例
	loggerInstance := logger.New("INFO")
	userSvc := NewUserService(loggerInstance)

	// 使用服务
	user := &User{
		ID:    1,
		Name:  "张三",
		Email: "zhangsan@example.com",
	}

	userSvc.CreateUser(user)
	retrievedUser, _ := userSvc.GetUser(1)
	fmt.Printf("获取到的用户: %+v\n", retrievedUser)

	fmt.Println()
}

// =============================================================================
// 主函数
// =============================================================================

func main() {
	fmt.Println("Go 包管理和组织 - 完整示例")
	fmt.Println("==============================")

	demonstratePackageImports()
	demonstrateVisibility()
	demonstratePackageInit()
	demonstrateImportStyles()
	demonstrateInternalPackages()
	demonstrateDocumentation()
	demonstrateModules()
	demonstrateBestPractices()
	demonstrateProjectStructure()

	fmt.Println("=== 练习任务 ===")
	fmt.Println("1. 创建一个 'config' 包，实现配置文件的读取和管理")
	fmt.Println("2. 设计一个 'cache' 包，提供内存缓存功能")
	fmt.Println("3. 实现一个 'middleware' 包，为 HTTP 服务提供中间件")
	fmt.Println("4. 创建一个 'validation' 包，提供数据验证功能")
	fmt.Println("5. 设计合理的项目结构，演示包之间的依赖关系")
	fmt.Println("\n在对应的子目录中创建这些包，并在主程序中使用它们！")
}
