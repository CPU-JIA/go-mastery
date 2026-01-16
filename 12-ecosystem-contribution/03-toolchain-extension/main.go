// Package main 演示 Go 工具链扩展开发
// 本模块涵盖 Go 工具链扩展的核心技术：
// - 自定义代码分析器 (go/analysis)
// - 代码生成工具 (go generate)
// - 自定义 go 命令插件
// - LSP 扩展开发
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// ============================================================================
// 代码分析器框架
// ============================================================================

// Analyzer 分析器接口
type Analyzer struct {
	Name     string
	Doc      string
	Run      func(*AnalysisPass) (interface{}, error)
	Requires []*Analyzer
}

// AnalysisPass 分析传递
type AnalysisPass struct {
	Analyzer    *Analyzer
	Fset        *token.FileSet
	Files       []*ast.File
	Report      func(pos token.Pos, message string)
	ResultOf    map[*Analyzer]interface{}
	Diagnostics []Diagnostic
}

// Diagnostic 诊断信息
type Diagnostic struct {
	Pos      token.Position
	Category string
	Message  string
	Severity DiagnosticSeverity
}

// DiagnosticSeverity 诊断严重程度
type DiagnosticSeverity int

const (
	SeverityHint DiagnosticSeverity = iota
	SeverityInfo
	SeverityWarning
	SeverityError
)

func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityHint:
		return "Hint"
	case SeverityInfo:
		return "Info"
	case SeverityWarning:
		return "Warning"
	case SeverityError:
		return "Error"
	default:
		return "Unknown"
	}
}

// ============================================================================
// 示例分析器：检测未使用的变量
// ============================================================================

// UnusedVarAnalyzer 未使用变量分析器
type UnusedVarAnalyzer struct {
	name string
	doc  string
}

// NewUnusedVarAnalyzer 创建未使用变量分析器
func NewUnusedVarAnalyzer() *UnusedVarAnalyzer {
	return &UnusedVarAnalyzer{
		name: "unusedvar",
		doc:  "检测声明但未使用的变量",
	}
}

// Analyze 执行分析
func (a *UnusedVarAnalyzer) Analyze(fset *token.FileSet, file *ast.File) []Diagnostic {
	var diagnostics []Diagnostic

	// 收集所有声明的变量
	declared := make(map[string]token.Pos)
	// 收集所有使用的变量
	used := make(map[string]bool)

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			// 短变量声明
			if node.Tok == token.DEFINE {
				for _, lhs := range node.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok {
						declared[ident.Name] = ident.Pos()
					}
				}
			}
		case *ast.ValueSpec:
			// var 声明
			for _, name := range node.Names {
				declared[name.Name] = name.Pos()
			}
		case *ast.Ident:
			// 变量使用
			used[node.Name] = true
		}
		return true
	})

	// 检查未使用的变量
	for name, pos := range declared {
		if !used[name] && name != "_" {
			diagnostics = append(diagnostics, Diagnostic{
				Pos:      fset.Position(pos),
				Category: "unusedvar",
				Message:  fmt.Sprintf("变量 '%s' 已声明但未使用", name),
				Severity: SeverityWarning,
			})
		}
	}

	return diagnostics
}

// ============================================================================
// 示例分析器：检测错误处理
// ============================================================================

// ErrorCheckAnalyzer 错误检查分析器
type ErrorCheckAnalyzer struct {
	name string
	doc  string
}

// NewErrorCheckAnalyzer 创建错误检查分析器
func NewErrorCheckAnalyzer() *ErrorCheckAnalyzer {
	return &ErrorCheckAnalyzer{
		name: "errcheck",
		doc:  "检测未处理的错误返回值",
	}
}

// Analyze 执行分析
func (a *ErrorCheckAnalyzer) Analyze(fset *token.FileSet, file *ast.File) []Diagnostic {
	var diagnostics []Diagnostic

	ast.Inspect(file, func(n ast.Node) bool {
		// 检查表达式语句（可能是忽略返回值的函数调用）
		if exprStmt, ok := n.(*ast.ExprStmt); ok {
			if call, ok := exprStmt.X.(*ast.CallExpr); ok {
				// 简化检查：如果函数名包含常见的错误返回函数
				if ident, ok := call.Fun.(*ast.Ident); ok {
					if isErrorReturningFunc(ident.Name) {
						diagnostics = append(diagnostics, Diagnostic{
							Pos:      fset.Position(call.Pos()),
							Category: "errcheck",
							Message:  fmt.Sprintf("函数 '%s' 的返回值未被检查", ident.Name),
							Severity: SeverityWarning,
						})
					}
				}
			}
		}
		return true
	})

	return diagnostics
}

func isErrorReturningFunc(name string) bool {
	errorFuncs := []string{"Write", "Read", "Close", "Open", "Create"}
	for _, f := range errorFuncs {
		if strings.Contains(name, f) {
			return true
		}
	}
	return false
}

// ============================================================================
// 代码生成器框架
// ============================================================================

// CodeGenerator 代码生成器
type CodeGenerator struct {
	name        string
	description string
	templates   map[string]string
}

// NewCodeGenerator 创建代码生成器
func NewCodeGenerator(name, description string) *CodeGenerator {
	return &CodeGenerator{
		name:        name,
		description: description,
		templates:   make(map[string]string),
	}
}

// AddTemplate 添加模板
func (g *CodeGenerator) AddTemplate(name, template string) {
	g.templates[name] = template
}

// Generate 生成代码
func (g *CodeGenerator) Generate(templateName string, data map[string]interface{}) (string, error) {
	template, exists := g.templates[templateName]
	if !exists {
		return "", fmt.Errorf("模板不存在: %s", templateName)
	}

	// 简单的模板替换
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}

	return result, nil
}

// ============================================================================
// 示例：Stringer 生成器
// ============================================================================

// StringerGenerator Stringer 方法生成器
type StringerGenerator struct {
	*CodeGenerator
}

// NewStringerGenerator 创建 Stringer 生成器
func NewStringerGenerator() *StringerGenerator {
	gen := &StringerGenerator{
		CodeGenerator: NewCodeGenerator("stringer", "为枚举类型生成 String() 方法"),
	}

	gen.AddTemplate("stringer", `
// Code generated by stringer; DO NOT EDIT.

package {{.Package}}

import "strconv"

func _() {
	// 编译时检查，确保枚举值没有改变
	var x [1]struct{}
	{{range .Values}}_ = x[{{.Name}}-{{.Value}}]
	{{end}}
}

const _{{.TypeName}}_name = "{{.Names}}"

var _{{.TypeName}}_index = [...]uint8{{"{"}}{{.Indices}}{{"}"}}

func (i {{.TypeName}}) String() string {
	if i < 0 || i >= {{.TypeName}}(len(_{{.TypeName}}_index)-1) {
		return "{{.TypeName}}(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _{{.TypeName}}_name[_{{.TypeName}}_index[i]:_{{.TypeName}}_index[i+1]]
}
`)

	return gen
}

// GenerateStringer 生成 Stringer 代码
func (g *StringerGenerator) GenerateStringer(pkg, typeName string, values []EnumValue) (string, error) {
	// 构建名称字符串和索引
	var names strings.Builder
	var indices []string
	currentIndex := 0

	for _, v := range values {
		names.WriteString(v.Name)
		indices = append(indices, fmt.Sprintf("%d", currentIndex))
		currentIndex += len(v.Name)
	}
	indices = append(indices, fmt.Sprintf("%d", currentIndex))

	data := map[string]interface{}{
		"Package":  pkg,
		"TypeName": typeName,
		"Names":    names.String(),
		"Indices":  strings.Join(indices, ", "),
		"Values":   values,
	}

	return g.Generate("stringer", data)
}

// EnumValue 枚举值
type EnumValue struct {
	Name  string
	Value int
}

// ============================================================================
// 示例：Mock 生成器
// ============================================================================

// MockGenerator Mock 生成器
type MockGenerator struct {
	*CodeGenerator
}

// NewMockGenerator 创建 Mock 生成器
func NewMockGenerator() *MockGenerator {
	gen := &MockGenerator{
		CodeGenerator: NewCodeGenerator("mockgen", "为接口生成 Mock 实现"),
	}

	gen.AddTemplate("mock", `
// Code generated by mockgen; DO NOT EDIT.

package {{.Package}}

import (
	"sync"
)

// Mock{{.InterfaceName}} is a mock implementation of {{.InterfaceName}}
type Mock{{.InterfaceName}} struct {
	mu sync.Mutex
	{{range .Methods}}
	{{.Name}}Func func({{.Params}}) {{.Returns}}
	{{.Name}}Called int
	{{end}}
}

// NewMock{{.InterfaceName}} creates a new mock instance
func NewMock{{.InterfaceName}}() *Mock{{.InterfaceName}} {
	return &Mock{{.InterfaceName}}{}
}

{{range .Methods}}
// {{.Name}} implements {{$.InterfaceName}}.{{.Name}}
func (m *Mock{{$.InterfaceName}}) {{.Name}}({{.Params}}) {{.Returns}} {
	m.mu.Lock()
	m.{{.Name}}Called++
	m.mu.Unlock()

	if m.{{.Name}}Func != nil {
		return m.{{.Name}}Func({{.ParamNames}})
	}
	{{if .HasReturn}}return {{.ZeroReturn}}{{end}}
}
{{end}}
`)

	return gen
}

// InterfaceMethod 接口方法
type InterfaceMethod struct {
	Name       string
	Params     string
	ParamNames string
	Returns    string
	HasReturn  bool
	ZeroReturn string
}

// GenerateMock 生成 Mock 代码
func (g *MockGenerator) GenerateMock(pkg, interfaceName string, methods []InterfaceMethod) (string, error) {
	data := map[string]interface{}{
		"Package":       pkg,
		"InterfaceName": interfaceName,
		"Methods":       methods,
	}

	return g.Generate("mock", data)
}

// ============================================================================
// 分析器运行器
// ============================================================================

// AnalyzerRunner 分析器运行器
type AnalyzerRunner struct {
	analyzers []interface{}
}

// NewAnalyzerRunner 创建分析器运行器
func NewAnalyzerRunner() *AnalyzerRunner {
	return &AnalyzerRunner{
		analyzers: make([]interface{}, 0),
	}
}

// AddAnalyzer 添加分析器
func (r *AnalyzerRunner) AddAnalyzer(analyzer interface{}) {
	r.analyzers = append(r.analyzers, analyzer)
}

// Run 运行所有分析器
func (r *AnalyzerRunner) Run(src string) []Diagnostic {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, parser.AllErrors)
	if err != nil {
		return []Diagnostic{{
			Category: "parser",
			Message:  fmt.Sprintf("解析错误: %v", err),
			Severity: SeverityError,
		}}
	}

	var allDiagnostics []Diagnostic

	for _, analyzer := range r.analyzers {
		switch a := analyzer.(type) {
		case *UnusedVarAnalyzer:
			diagnostics := a.Analyze(fset, file)
			allDiagnostics = append(allDiagnostics, diagnostics...)
		case *ErrorCheckAnalyzer:
			diagnostics := a.Analyze(fset, file)
			allDiagnostics = append(allDiagnostics, diagnostics...)
		}
	}

	return allDiagnostics
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateAnalyzer() {
	fmt.Println("\n=== 代码分析器演示 ===")
	fmt.Println("场景: 使用自定义分析器检查代码问题")

	// 示例代码
	src := `
package main

import "fmt"

func main() {
	x := 10
	y := 20
	fmt.Println(x)
	// y 未使用

	file.Close()  // 错误未检查
}
`

	fmt.Println("\n待分析代码:")
	fmt.Println(src)

	// 创建分析器运行器
	runner := NewAnalyzerRunner()
	runner.AddAnalyzer(NewUnusedVarAnalyzer())
	runner.AddAnalyzer(NewErrorCheckAnalyzer())

	// 运行分析
	fmt.Println("分析结果:")
	diagnostics := runner.Run(src)

	if len(diagnostics) == 0 {
		fmt.Println("  未发现问题")
	} else {
		for _, d := range diagnostics {
			fmt.Printf("  [%s] %s: %s\n", d.Severity, d.Category, d.Message)
		}
	}
}

func demonstrateCodeGenerator() {
	fmt.Println("\n=== 代码生成器演示 ===")
	fmt.Println("场景: 为枚举类型生成 String() 方法")

	// 模拟枚举定义
	fmt.Println("\n原始枚举定义:")
	fmt.Println(`type Color int

const (
	Red Color = iota
	Green
	Blue
)`)

	// 生成 Stringer
	gen := NewStringerGenerator()
	values := []EnumValue{
		{Name: "Red", Value: 0},
		{Name: "Green", Value: 1},
		{Name: "Blue", Value: 2},
	}

	fmt.Println("生成的 String() 方法 (简化版):")
	fmt.Println(`func (c Color) String() string {
	switch c {
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	default:
		return "Color(" + strconv.Itoa(int(c)) + ")"
	}
}`)

	// 显示生成器信息
	fmt.Printf("\n生成器: %s\n", gen.name)
	fmt.Printf("描述: %s\n", gen.description)
	fmt.Printf("枚举值: %v\n", values)
}

func demonstrateMockGenerator() {
	fmt.Println("\n=== Mock 生成器演示 ===")
	fmt.Println("场景: 为接口生成 Mock 实现")

	// 模拟接口定义
	fmt.Println("\n原始接口定义:")
	fmt.Println(`type UserRepository interface {
	GetByID(id int) (*User, error)
	Save(user *User) error
	Delete(id int) error
}`)

	// 生成 Mock
	gen := NewMockGenerator()
	methods := []InterfaceMethod{
		{
			Name:       "GetByID",
			Params:     "id int",
			ParamNames: "id",
			Returns:    "(*User, error)",
			HasReturn:  true,
			ZeroReturn: "nil, nil",
		},
		{
			Name:       "Save",
			Params:     "user *User",
			ParamNames: "user",
			Returns:    "error",
			HasReturn:  true,
			ZeroReturn: "nil",
		},
		{
			Name:       "Delete",
			Params:     "id int",
			ParamNames: "id",
			Returns:    "error",
			HasReturn:  true,
			ZeroReturn: "nil",
		},
	}

	fmt.Println("生成的 Mock 实现 (简化版):")
	fmt.Println(`type MockUserRepository struct {
	GetByIDFunc func(id int) (*User, error)
	SaveFunc    func(user *User) error
	DeleteFunc  func(id int) error
}

func (m *MockUserRepository) GetByID(id int) (*User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

// ... 其他方法类似`)

	fmt.Printf("\n生成器: %s\n", gen.name)
	fmt.Printf("描述: %s\n", gen.description)
	fmt.Printf("方法数: %d\n", len(methods))
}

func demonstrateGoGenerate() {
	fmt.Println("\n=== go generate 使用指南 ===")
	fmt.Println(`go generate 是 Go 的代码生成工具，通过特殊注释触发。

使用方式:

1. 在源文件中添加 generate 指令:
   //go:generate stringer -type=Color
   //go:generate mockgen -source=repository.go -destination=mock_repository.go

2. 运行 go generate:
   go generate ./...

常用生成器:

1. stringer (golang.org/x/tools/cmd/stringer)
   - 为枚举类型生成 String() 方法
   - 用法: //go:generate stringer -type=MyEnum

2. mockgen (github.com/golang/mock/mockgen)
   - 为接口生成 Mock 实现
   - 用法: //go:generate mockgen -source=interface.go

3. protoc-gen-go (google.golang.org/protobuf/cmd/protoc-gen-go)
   - 从 .proto 文件生成 Go 代码
   - 用法: //go:generate protoc --go_out=. *.proto

4. go-bindata (github.com/go-bindata/go-bindata)
   - 将静态文件嵌入 Go 代码
   - 用法: //go:generate go-bindata -o bindata.go assets/

5. enumer (github.com/dmarkham/enumer)
   - stringer 的增强版，支持更多功能
   - 用法: //go:generate enumer -type=Status -json

最佳实践:

1. 将生成的文件命名为 *_gen.go 或 *_string.go
2. 在生成的文件开头添加 "Code generated ... DO NOT EDIT."
3. 将 go generate 命令添加到 Makefile 或 CI 流程
4. 提交生成的代码到版本控制`)
}

func demonstrateLSPExtension() {
	fmt.Println("\n=== LSP 扩展开发 ===")
	fmt.Println(`Language Server Protocol (LSP) 扩展开发指南:

gopls 是 Go 的官方语言服务器，支持以下功能:

1. 代码补全 (Completion)
   - 智能补全变量、函数、类型
   - 支持导入包自动补全

2. 跳转定义 (Go to Definition)
   - 跳转到函数、类型、变量定义
   - 支持跨包跳转

3. 查找引用 (Find References)
   - 查找所有使用位置
   - 支持重命名重构

4. 悬停信息 (Hover)
   - 显示类型信息
   - 显示文档注释

5. 代码诊断 (Diagnostics)
   - 语法错误
   - 类型错误
   - 静态分析警告

扩展 gopls 的方式:

1. 自定义分析器
   - 实现 go/analysis.Analyzer 接口
   - 通过 gopls 配置启用

2. 代码操作 (Code Actions)
   - 快速修复
   - 重构建议

3. 代码片段 (Snippets)
   - 自定义代码模板
   - 参数占位符

配置示例 (settings.json):
{
  "gopls": {
    "analyses": {
      "unusedparams": true,
      "shadow": true
    },
    "staticcheck": true,
    "gofumpt": true
  }
}`)
}

func main() {
	fmt.Println("=== Go 工具链扩展开发 ===")
	fmt.Println()
	fmt.Println("本模块演示 Go 工具链扩展的核心技术:")
	fmt.Println("1. 自定义代码分析器")
	fmt.Println("2. 代码生成工具")
	fmt.Println("3. go generate 使用")
	fmt.Println("4. LSP 扩展开发")

	demonstrateAnalyzer()
	demonstrateCodeGenerator()
	demonstrateMockGenerator()
	demonstrateGoGenerate()
	demonstrateLSPExtension()

	fmt.Println("\n=== 工具链扩展演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- go/analysis 包提供了标准的分析器框架")
	fmt.Println("- go/ast 和 go/parser 用于解析和操作 Go 代码")
	fmt.Println("- go generate 是代码生成的标准方式")
	fmt.Println("- gopls 是可扩展的语言服务器")
	fmt.Println("- 好的工具应该集成到现有工作流中")
}
