/*
=== Go编译器工具链：AST抽象语法树大师 ===

本模块深入探索Go语言AST（抽象语法树）操作的精髓，包括：
1. AST基础理论和Go语言AST结构
2. AST节点类型系统和层次结构
3. AST遍历算法和访问者模式
4. AST操作：增删改查和转换
5. 代码生成和重构技术
6. 静态分析和代码检查
7. 源码级别的程序变换
8. 性能优化和最佳实践
9. 实际应用：linter、formatter、refactoring tool
10. AST可视化和调试技术

学习目标：
- 深入理解Go语言AST结构和设计哲学
- 掌握AST操作的核心技术和算法
- 能够构建静态分析和代码转换工具
- 理解编译器前端的工作原理
- 具备开发Go语言工具链的能力
*/

package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// ==================
// 1. AST核心概念和工具
// ==================

// ASTAnalyzer AST分析器
type ASTAnalyzer struct {
	fileSet    *token.FileSet
	packages   map[string]*ast.Package
	files      map[string]*ast.File
	typeInfo   *types.Info
	config     AnalyzerConfig
	statistics AnalyzerStatistics
	visitors   map[string]ast.Visitor
	transforms []ASTTransform
	mutex      sync.RWMutex
}

// AnalyzerConfig AST分析器配置
type AnalyzerConfig struct {
	ParseComments    bool
	ParseTests       bool
	EnableTypeCheck  bool
	EnableDeps       bool
	MaxDepth         int
	EnableProfiling  bool
	EnableCaching    bool
	ParallelAnalysis bool
	OutputFormat     string
}

// AnalyzerStatistics AST分析统计
type AnalyzerStatistics struct {
	FilesProcessed int
	NodesAnalyzed  int64
	FunctionsFound int
	VariablesFound int
	ImportsFound   int
	ErrorsFound    int
	WarningsFound  int
	ProcessingTime int64 // 毫秒
	MemoryUsage    int64 // 字节
}

// ASTTransform AST转换接口
type ASTTransform interface {
	Transform(node ast.Node) ast.Node
	Name() string
	Priority() int
}

// NewASTAnalyzer 创建AST分析器
func NewASTAnalyzer(config AnalyzerConfig) *ASTAnalyzer {
	return &ASTAnalyzer{
		fileSet:  token.NewFileSet(),
		packages: make(map[string]*ast.Package),
		files:    make(map[string]*ast.File),
		typeInfo: &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		},
		config:     config,
		visitors:   make(map[string]ast.Visitor),
		transforms: make([]ASTTransform, 0),
	}
}

// ==================
// 2. AST节点结构解析
// ==================

// 演示Go语言AST节点类型和结构
func demonstrateASTStructure() {
	fmt.Println("=== 1. Go语言AST节点结构解析 ===")

	// 示例Go代码
	src := `package main

import (
	"fmt"
	"os"
)

// User 用户结构
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// SayHello 方法示例
func (u User) SayHello() string {
	return fmt.Sprintf("Hello, %s!", u.Name)
}

// 常量定义
const (
	MaxUsers = 100
	Version  = "1.0.0"
)

// 变量定义
var globalVar = "global"

func main() {
	user := User{ID: 1, Name: "Alice"}
	message := user.SayHello()
	fmt.Println(message)

	// 条件语句
	if len(os.Args) > 1 {
		fmt.Println("Arguments provided")
	}

	// 循环语句
	for i := 0; i < 3; i++ {
		fmt.Printf("Iteration %d\n", i)
	}
}`

	// 解析源码为AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("文件名: %s\n", file.Name.Name)
	fmt.Printf("包声明: %s\n", file.Name)
	fmt.Printf("导入数量: %d\n", len(file.Imports))
	fmt.Printf("声明数量: %d\n", len(file.Decls))

	// 分析导入
	fmt.Println("\n导入分析:")
	for i, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		}
		fmt.Printf("  %d. 路径: %s, 别名: %s\n", i+1, path, name)
	}

	// 分析声明
	fmt.Println("\n声明分析:")
	for i, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			analyzeGenDecl(d, i+1)
		case *ast.FuncDecl:
			analyzeFuncDecl(d, i+1)
		}
	}

	fmt.Println()
}

func analyzeGenDecl(decl *ast.GenDecl, index int) {
	fmt.Printf("  %d. 通用声明 - Token: %s\n", index, decl.Tok)
	for j, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.ImportSpec:
			fmt.Printf("     %d.%d 导入: %s\n", index, j+1, s.Path.Value)
		case *ast.TypeSpec:
			fmt.Printf("     %d.%d 类型定义: %s\n", index, j+1, s.Name.Name)
			if structType, ok := s.Type.(*ast.StructType); ok {
				for k, field := range structType.Fields.List {
					fieldName := "anonymous"
					if len(field.Names) > 0 {
						fieldName = field.Names[0].Name
					}
					fmt.Printf("         字段 %d: %s\n", k+1, fieldName)
				}
			}
		case *ast.ValueSpec:
			names := make([]string, len(s.Names))
			for k, name := range s.Names {
				names[k] = name.Name
			}
			fmt.Printf("     %d.%d 值声明: %s\n", index, j+1, strings.Join(names, ", "))
		}
	}
}

func analyzeFuncDecl(decl *ast.FuncDecl, index int) {
	funcName := decl.Name.Name
	fmt.Printf("  %d. 函数声明: %s\n", index, funcName)

	// 分析接收者
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		recv := decl.Recv.List[0]
		recvType := getTypeString(recv.Type)
		fmt.Printf("     接收者类型: %s\n", recvType)
	}

	// 分析参数
	if decl.Type.Params != nil {
		fmt.Printf("     参数数量: %d\n", len(decl.Type.Params.List))
	}

	// 分析返回值
	if decl.Type.Results != nil {
		fmt.Printf("     返回值数量: %d\n", len(decl.Type.Results.List))
	}

	// 分析函数体
	if decl.Body != nil {
		fmt.Printf("     语句数量: %d\n", len(decl.Body.List))
	}
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

// ==================
// 3. AST遍历和访问者模式
// ==================

// ASTVisitor 访问者接口
type ASTVisitor interface {
	Visit(node ast.Node) ast.Visitor
	Name() string
	GetResults() interface{}
}

// FunctionVisitor 函数访问者
type FunctionVisitor struct {
	functions []FunctionInfo
}

type FunctionInfo struct {
	Name       string
	Line       int
	Column     int
	ParamCount int
	IsMethod   bool
	Receiver   string
}

func NewFunctionVisitor() *FunctionVisitor {
	return &FunctionVisitor{
		functions: make([]FunctionInfo, 0),
	}
}

func (v *FunctionVisitor) Name() string {
	return "FunctionVisitor"
}

func (v *FunctionVisitor) GetResults() interface{} {
	return v.functions
}

func (v *FunctionVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		info := FunctionInfo{
			Name:       n.Name.Name,
			IsMethod:   n.Recv != nil,
			ParamCount: 0,
		}

		if n.Type.Params != nil {
			info.ParamCount = len(n.Type.Params.List)
		}

		if n.Recv != nil && len(n.Recv.List) > 0 {
			info.Receiver = getTypeString(n.Recv.List[0].Type)
		}

		v.functions = append(v.functions, info)
	}
	return v
}

// VariableVisitor 变量访问者
type VariableVisitor struct {
	variables []VariableInfo
}

type VariableInfo struct {
	Name  string
	Type  string
	Scope string
	Line  int
}

func NewVariableVisitor() *VariableVisitor {
	return &VariableVisitor{
		variables: make([]VariableInfo, 0),
	}
}

func (v *VariableVisitor) Name() string {
	return "VariableVisitor"
}

func (v *VariableVisitor) GetResults() interface{} {
	return v.variables
}

func (v *VariableVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.GenDecl:
		if n.Tok == token.VAR {
			for _, spec := range n.Specs {
				if vspec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range vspec.Names {
						info := VariableInfo{
							Name:  name.Name,
							Type:  getTypeString(vspec.Type),
							Scope: "package",
						}
						v.variables = append(v.variables, info)
					}
				}
			}
		}
	case *ast.AssignStmt:
		if n.Tok == token.DEFINE {
			for _, expr := range n.Lhs {
				if ident, ok := expr.(*ast.Ident); ok {
					info := VariableInfo{
						Name:  ident.Name,
						Type:  "inferred",
						Scope: "local",
					}
					v.variables = append(v.variables, info)
				}
			}
		}
	}
	return v
}

// 演示AST遍历和访问者模式
func demonstrateASTTraversal() {
	fmt.Println("=== 2. AST遍历和访问者模式 ===")

	src := `package main

import "fmt"

type Calculator struct {
	result float64
}

func (c *Calculator) Add(a, b float64) float64 {
	c.result = a + b
	return c.result
}

func NewCalculator() *Calculator {
	return &Calculator{result: 0}
}

var globalCounter int = 0

func main() {
	calc := NewCalculator()
	sum := calc.Add(10, 20)
	fmt.Printf("Result: %.2f\n", sum)

	localVar := "local"
	_ = localVar
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "calculator.go", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// 使用函数访问者
	funcVisitor := NewFunctionVisitor()
	ast.Walk(funcVisitor, file)

	functions := funcVisitor.GetResults().([]FunctionInfo)
	fmt.Printf("发现函数数量: %d\n", len(functions))
	for i, fn := range functions {
		fmt.Printf("  %d. %s (参数: %d, 方法: %t", i+1, fn.Name, fn.ParamCount, fn.IsMethod)
		if fn.IsMethod {
			fmt.Printf(", 接收者: %s", fn.Receiver)
		}
		fmt.Println(")")
	}

	// 使用变量访问者
	varVisitor := NewVariableVisitor()
	ast.Walk(varVisitor, file)

	variables := varVisitor.GetResults().([]VariableInfo)
	fmt.Printf("\n发现变量数量: %d\n", len(variables))
	for i, v := range variables {
		fmt.Printf("  %d. %s (类型: %s, 作用域: %s)\n", i+1, v.Name, v.Type, v.Scope)
	}

	fmt.Println()
}

// ==================
// 4. AST操作和转换
// ==================

// CodeTransformer 代码转换器
type CodeTransformer struct {
	fset       *token.FileSet
	transforms []ASTTransform
}

// MethodToFunctionTransform 方法转函数转换
type MethodToFunctionTransform struct {
	name string
}

func NewMethodToFunctionTransform() *MethodToFunctionTransform {
	return &MethodToFunctionTransform{name: "MethodToFunction"}
}

func (t *MethodToFunctionTransform) Name() string {
	return t.name
}

func (t *MethodToFunctionTransform) Priority() int {
	return 1
}

func (t *MethodToFunctionTransform) Transform(node ast.Node) ast.Node {
	if funcDecl, ok := node.(*ast.FuncDecl); ok {
		if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
			// 将方法转换为函数
			newFunc := &ast.FuncDecl{
				Name: ast.NewIdent(funcDecl.Name.Name + "Func"),
				Type: &ast.FuncType{
					Params: &ast.FieldList{
						List: make([]*ast.Field, 0),
					},
				},
				Body: funcDecl.Body,
			}

			// 添加接收者作为第一个参数
			recvField := funcDecl.Recv.List[0]
			newFunc.Type.Params.List = append(newFunc.Type.Params.List, recvField)

			// 添加原有参数
			if funcDecl.Type.Params != nil {
				newFunc.Type.Params.List = append(newFunc.Type.Params.List, funcDecl.Type.Params.List...)
			}

			// 保留返回类型
			newFunc.Type.Results = funcDecl.Type.Results

			return newFunc
		}
	}
	return node
}

// AddLogTransform 添加日志转换
type AddLogTransform struct {
	name string
}

func NewAddLogTransform() *AddLogTransform {
	return &AddLogTransform{name: "AddLog"}
}

func (t *AddLogTransform) Name() string {
	return t.name
}

func (t *AddLogTransform) Priority() int {
	return 2
}

func (t *AddLogTransform) Transform(node ast.Node) ast.Node {
	if funcDecl, ok := node.(*ast.FuncDecl); ok {
		if funcDecl.Body != nil && funcDecl.Name.Name != "init" {
			// 在函数开头添加日志语句
			logStmt := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   ast.NewIdent("fmt"),
						Sel: ast.NewIdent("Printf"),
					},
					Args: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: `"Entering function: ` + funcDecl.Name.Name + `\n"`,
						},
					},
				},
			}

			// 创建新的语句列表
			newStmts := make([]ast.Stmt, 0, len(funcDecl.Body.List)+1)
			newStmts = append(newStmts, logStmt)
			newStmts = append(newStmts, funcDecl.Body.List...)

			// 更新函数体
			funcDecl.Body.List = newStmts
		}
	}
	return node
}

// 演示AST操作和转换
func demonstrateASTTransformation() {
	fmt.Println("=== 3. AST操作和转换 ===")

	src := `package main

import "fmt"

type Counter struct {
	value int
}

func (c *Counter) Increment() {
	c.value++
}

func (c *Counter) GetValue() int {
	return c.value
}

func main() {
	counter := &Counter{}
	counter.Increment()
	fmt.Println(counter.GetValue())
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "counter.go", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("原始AST结构:")
	ast.Print(fset, file)

	// 应用转换
	transformer := &CodeTransformer{
		fset: fset,
		transforms: []ASTTransform{
			NewMethodToFunctionTransform(),
			NewAddLogTransform(),
		},
	}

	// 遍历并转换
	transformedFile := transformer.transformFile(file)

	fmt.Println("\n转换后的代码:")
	if err := format.Node(os.Stdout, fset, transformedFile); err != nil {
		log.Fatal(err)
	}

	fmt.Println()
}

func (ct *CodeTransformer) transformFile(file *ast.File) *ast.File {
	// 深度遍历并应用转换
	ast.Inspect(file, func(n ast.Node) bool {
		if n != nil {
			for _, transform := range ct.transforms {
				n = transform.Transform(n)
			}
		}
		return true
	})

	return file
}

// ==================
// 5. 静态分析工具
// ==================

// StaticAnalyzer 静态分析器
type StaticAnalyzer struct {
	rules   []AnalysisRule
	issues  []Issue
	metrics CodeMetrics
}

type AnalysisRule interface {
	Name() string
	Check(node ast.Node, fset *token.FileSet) []Issue
	Severity() Severity
}

type Issue struct {
	Rule       string
	Message    string
	Position   token.Pos
	Severity   Severity
	Suggestion string
	Category   string
}

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

type CodeMetrics struct {
	LinesOfCode          int
	CyclomaticComplexity int
	FunctionCount        int
	MaxFunctionLength    int
	DuplicationRatio     float64
}

// UnusedVariableRule 未使用变量检查规则
type UnusedVariableRule struct{}

func (r *UnusedVariableRule) Name() string {
	return "UnusedVariable"
}

func (r *UnusedVariableRule) Severity() Severity {
	return SeverityWarning
}

func (r *UnusedVariableRule) Check(node ast.Node, fset *token.FileSet) []Issue {
	issues := make([]Issue, 0)

	if assign, ok := node.(*ast.AssignStmt); ok {
		if assign.Tok == token.DEFINE {
			for _, lhs := range assign.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					if ident.Name != "_" && !isUsed(ident, node) {
						issues = append(issues, Issue{
							Rule:       r.Name(),
							Message:    fmt.Sprintf("Variable '%s' is defined but never used", ident.Name),
							Position:   ident.Pos(),
							Severity:   r.Severity(),
							Suggestion: fmt.Sprintf("Consider removing variable '%s' or using it", ident.Name),
							Category:   "CodeQuality",
						})
					}
				}
			}
		}
	}

	return issues
}

func isUsed(ident *ast.Ident, scope ast.Node) bool {
	// 简化的使用检查逻辑
	used := false
	ast.Inspect(scope, func(n ast.Node) bool {
		if other, ok := n.(*ast.Ident); ok {
			if other != ident && other.Name == ident.Name {
				used = true
				return false
			}
		}
		return true
	})
	return used
}

// LongFunctionRule 长函数检查规则
type LongFunctionRule struct {
	maxLength int
}

func NewLongFunctionRule(maxLength int) *LongFunctionRule {
	return &LongFunctionRule{maxLength: maxLength}
}

func (r *LongFunctionRule) Name() string {
	return "LongFunction"
}

func (r *LongFunctionRule) Severity() Severity {
	return SeverityWarning
}

func (r *LongFunctionRule) Check(node ast.Node, fset *token.FileSet) []Issue {
	issues := make([]Issue, 0)

	if funcDecl, ok := node.(*ast.FuncDecl); ok {
		if funcDecl.Body != nil {
			stmtCount := countStatements(funcDecl.Body)
			if stmtCount > r.maxLength {
				issues = append(issues, Issue{
					Rule:       r.Name(),
					Message:    fmt.Sprintf("Function '%s' is too long (%d statements, max %d)", funcDecl.Name.Name, stmtCount, r.maxLength),
					Position:   funcDecl.Pos(),
					Severity:   r.Severity(),
					Suggestion: "Consider breaking this function into smaller functions",
					Category:   "Maintainability",
				})
			}
		}
	}

	return issues
}

func countStatements(block *ast.BlockStmt) int {
	count := 0
	ast.Inspect(block, func(n ast.Node) bool {
		switch n.(type) {
		case ast.Stmt:
			count++
		}
		return true
	})
	return count
}

// 演示静态分析工具
func demonstrateStaticAnalysis() {
	fmt.Println("=== 4. 静态分析工具 ===")

	src := `package main

import "fmt"

func longFunction(x int) int {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
	g := 7
	h := 8
	i := 9
	j := 10
	unused := "this is not used"
	result := a + b + c + d + e + f + g + h + i + j
	return result + x
}

func shortFunction() {
	fmt.Println("Short and sweet")
}

func main() {
	result := longFunction(42)
	fmt.Println(result)
	shortFunction()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "analysis_example.go", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// 创建静态分析器
	analyzer := &StaticAnalyzer{
		rules: []AnalysisRule{
			&UnusedVariableRule{},
			NewLongFunctionRule(5), // 最大5个语句
		},
		issues: make([]Issue, 0),
	}

	// 运行分析
	analyzer.analyzeFile(file, fset)

	fmt.Printf("分析完成，发现 %d 个问题:\n", len(analyzer.issues))
	for i, issue := range analyzer.issues {
		pos := fset.Position(issue.Position)
		severityStr := []string{"INFO", "WARNING", "ERROR", "CRITICAL"}[issue.Severity]
		fmt.Printf("  %d. [%s] %s:%d:%d - %s\n",
			i+1, severityStr, pos.Filename, pos.Line, pos.Column, issue.Message)
		fmt.Printf("     建议: %s\n", issue.Suggestion)
	}

	fmt.Println()
}

func (sa *StaticAnalyzer) analyzeFile(file *ast.File, fset *token.FileSet) {
	ast.Inspect(file, func(n ast.Node) bool {
		if n != nil {
			for _, rule := range sa.rules {
				issues := rule.Check(n, fset)
				sa.issues = append(sa.issues, issues...)
			}
		}
		return true
	})
}

// ==================
// 6. 代码生成工具
// ==================

// CodeGenerator 代码生成器
type CodeGenerator struct {
	templates map[string]string
	output    strings.Builder
}

func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		templates: make(map[string]string),
	}
	cg.initTemplates()
	return cg
}

func (cg *CodeGenerator) initTemplates() {
	cg.templates["getter"] = `
func ({{.ReceiverName}} *{{.TypeName}}) Get{{.FieldName}}() {{.FieldType}} {
	return {{.ReceiverName}}.{{.FieldName}}
}`

	cg.templates["setter"] = `
func ({{.ReceiverName}} *{{.TypeName}}) Set{{.FieldName}}(value {{.FieldType}}) {
	{{.ReceiverName}}.{{.FieldName}} = value
}`

	cg.templates["constructor"] = `
func New{{.TypeName}}({{.Parameters}}) *{{.TypeName}} {
	return &{{.TypeName}}{
		{{.Assignments}}
	}
}`
}

// 演示代码生成
func demonstrateCodeGeneration() {
	fmt.Println("=== 5. 代码生成工具 ===")

	src := `package main

type Person struct {
	ID   int
	Name string
	Age  int
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "person.go", src, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	generator := NewCodeGenerator()

	// 分析结构体并生成代码
	ast.Inspect(file, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				generator.generateStructMethods(typeSpec.Name.Name, structType)
			}
		}
		return true
	})

	fmt.Println("生成的代码:")
	fmt.Println(generator.output.String())
}

func (cg *CodeGenerator) generateStructMethods(typeName string, structType *ast.StructType) {
	receiverName := strings.ToLower(typeName[:1])

	// 生成构造函数
	var params, assignments []string
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fieldType := getTypeString(field.Type)
			paramName := strings.ToLower(name.Name)

			params = append(params, fmt.Sprintf("%s %s", paramName, fieldType))
			assignments = append(assignments, fmt.Sprintf("%s: %s,", name.Name, paramName))
		}
	}

	cg.output.WriteString(fmt.Sprintf("\nfunc New%s(%s) *%s {\n",
		typeName, strings.Join(params, ", "), typeName))
	cg.output.WriteString(fmt.Sprintf("\treturn &%s{\n", typeName))
	for _, assignment := range assignments {
		cg.output.WriteString(fmt.Sprintf("\t\t%s\n", assignment))
	}
	cg.output.WriteString("\t}\n}\n")

	// 生成getter和setter方法
	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			fieldType := getTypeString(field.Type)

			// Getter
			cg.output.WriteString(fmt.Sprintf("\nfunc (%s *%s) Get%s() %s {\n",
				receiverName, typeName, name.Name, fieldType))
			cg.output.WriteString(fmt.Sprintf("\treturn %s.%s\n", receiverName, name.Name))
			cg.output.WriteString("}\n")

			// Setter
			cg.output.WriteString(fmt.Sprintf("\nfunc (%s *%s) Set%s(value %s) {\n",
				receiverName, typeName, name.Name, fieldType))
			cg.output.WriteString(fmt.Sprintf("\t%s.%s = value\n", receiverName, name.Name))
			cg.output.WriteString("}\n")
		}
	}
}

// ==================
// 7. AST可视化工具
// ==================

// ASTVisualizer AST可视化器
type ASTVisualizer struct {
	output  strings.Builder
	depth   int
	options VisualizerOptions
}

type VisualizerOptions struct {
	ShowPositions bool
	ShowTypes     bool
	MaxDepth      int
	CompactMode   bool
}

func NewASTVisualizer(options VisualizerOptions) *ASTVisualizer {
	return &ASTVisualizer{
		options: options,
	}
}

func (v *ASTVisualizer) Visualize(node ast.Node) string {
	v.output.Reset()
	v.depth = 0
	v.visualizeNode(node)
	return v.output.String()
}

func (v *ASTVisualizer) visualizeNode(node ast.Node) {
	if node == nil || (v.options.MaxDepth > 0 && v.depth > v.options.MaxDepth) {
		return
	}

	indent := strings.Repeat("  ", v.depth)
	nodeType := reflect.TypeOf(node).Elem().Name()

	v.output.WriteString(fmt.Sprintf("%s%s", indent, nodeType))

	// 添加节点特定信息
	switch n := node.(type) {
	case *ast.Ident:
		v.output.WriteString(fmt.Sprintf(" [%s]", n.Name))
	case *ast.BasicLit:
		v.output.WriteString(fmt.Sprintf(" [%s: %s]", n.Kind, n.Value))
	case *ast.FuncDecl:
		v.output.WriteString(fmt.Sprintf(" [%s]", n.Name.Name))
	case *ast.TypeSpec:
		v.output.WriteString(fmt.Sprintf(" [%s]", n.Name.Name))
	}

	v.output.WriteString("\n")

	// 递归可视化子节点
	v.depth++
	ast.Inspect(node, func(child ast.Node) bool {
		if child != node && child != nil {
			v.visualizeNode(child)
			return false // 不继续深入，我们手动控制
		}
		return child == node
	})
	v.depth--
}

// 演示AST可视化
func demonstrateASTVisualization() {
	fmt.Println("=== 6. AST可视化工具 ===")

	src := `package main

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fibonacci.go", src, 0)
	if err != nil {
		log.Fatal(err)
	}

	visualizer := NewASTVisualizer(VisualizerOptions{
		ShowPositions: true,
		ShowTypes:     true,
		MaxDepth:      10,
		CompactMode:   false,
	})

	visualization := visualizer.Visualize(file)
	fmt.Printf("AST可视化结果:\n%s\n", visualization)
}

// ==================
// 8. 性能优化技术
// ==================

// ASTOptimizer AST优化器
type ASTOptimizer struct {
	optimizations []Optimization
	statistics    OptimizationStats
}

type Optimization interface {
	Name() string
	Apply(node ast.Node) (ast.Node, bool)
	Description() string
}

type OptimizationStats struct {
	OptimizationsApplied int
	NodesOptimized       int
	EstimatedSpeedup     float64
}

// ConstantFoldingOptimization 常量折叠优化
type ConstantFoldingOptimization struct{}

func (o *ConstantFoldingOptimization) Name() string {
	return "ConstantFolding"
}

func (o *ConstantFoldingOptimization) Description() string {
	return "Fold constant expressions at compile time"
}

func (o *ConstantFoldingOptimization) Apply(node ast.Node) (ast.Node, bool) {
	if binExpr, ok := node.(*ast.BinaryExpr); ok {
		// 检查是否为常量表达式
		if lval := getConstantValue(binExpr.X); lval != nil {
			if rval := getConstantValue(binExpr.Y); rval != nil {
				// 执行常量折叠
				if result := evaluateConstantExpression(lval, binExpr.Op, rval); result != nil {
					return result, true
				}
			}
		}
	}
	return node, false
}

func getConstantValue(expr ast.Expr) *ast.BasicLit {
	if lit, ok := expr.(*ast.BasicLit); ok &&
		(lit.Kind == token.INT || lit.Kind == token.FLOAT) {
		return lit
	}
	return nil
}

func evaluateConstantExpression(left *ast.BasicLit, op token.Token, right *ast.BasicLit) *ast.BasicLit {
	if left.Kind == token.INT && right.Kind == token.INT {
		lval, _ := strconv.Atoi(left.Value)
		rval, _ := strconv.Atoi(right.Value)

		var result int
		switch op {
		case token.ADD:
			result = lval + rval
		case token.SUB:
			result = lval - rval
		case token.MUL:
			result = lval * rval
		case token.QUO:
			if rval != 0 {
				result = lval / rval
			} else {
				return nil
			}
		default:
			return nil
		}

		return &ast.BasicLit{
			Kind:  token.INT,
			Value: strconv.Itoa(result),
		}
	}
	return nil
}

// 演示性能优化
func demonstrateASTOptimization() {
	fmt.Println("=== 7. AST性能优化技术 ===")

	src := `package main

func calculate() int {
	a := 10 + 20
	b := 5 * 6
	c := 100 / 10
	return a + b + c
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "calculate.go", src, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("优化前的代码:")
	format.Node(os.Stdout, fset, file)

	optimizer := &ASTOptimizer{
		optimizations: []Optimization{
			&ConstantFoldingOptimization{},
		},
	}

	// 应用优化
	optimizedFile := optimizer.optimizeFile(file)

	fmt.Println("\n优化后的代码:")
	format.Node(os.Stdout, fset, optimizedFile)

	fmt.Printf("\n优化统计:\n")
	fmt.Printf("应用的优化数量: %d\n", optimizer.statistics.OptimizationsApplied)
	fmt.Printf("优化的节点数量: %d\n", optimizer.statistics.NodesOptimized)

	fmt.Println()
}

func (opt *ASTOptimizer) optimizeFile(file *ast.File) *ast.File {
	optimized := false

	ast.Inspect(file, func(n ast.Node) bool {
		for _, optimization := range opt.optimizations {
			if newNode, applied := optimization.Apply(n); applied {
				// 这里简化了实际的节点替换逻辑
				opt.statistics.OptimizationsApplied++
				opt.statistics.NodesOptimized++
				optimized = true
				_ = newNode // 实际应用中需要替换节点
			}
		}
		return true
	})

	_ = optimized
	return file
}

// ==================
// 主函数和综合演示
// ==================

func main() {
	fmt.Println("🚀 Go编译器工具链：AST抽象语法树大师")
	fmt.Println(strings.Repeat("=", 50))

	// 1. AST结构解析
	demonstrateASTStructure()

	// 2. AST遍历和访问者模式
	demonstrateASTTraversal()

	// 3. AST操作和转换
	demonstrateASTTransformation()

	// 4. 静态分析工具
	demonstrateStaticAnalysis()

	// 5. 代码生成工具
	demonstrateCodeGeneration()

	// 6. AST可视化工具
	demonstrateASTVisualization()

	// 7. 性能优化技术
	demonstrateASTOptimization()

	fmt.Println("🎯 AST操作大师课程完成！")
	fmt.Println("你现在已经掌握了:")
	fmt.Println("✅ Go语言AST结构和操作")
	fmt.Println("✅ 访问者模式和遍历技术")
	fmt.Println("✅ 代码转换和重构技术")
	fmt.Println("✅ 静态分析和代码检查")
	fmt.Println("✅ 自动化代码生成")
	fmt.Println("✅ AST可视化和调试")
	fmt.Println("✅ 编译器优化技术")
	fmt.Println()
	fmt.Println("🌟 继续探索编译器工具链的其他模块！")
}

/*
=== 练习题 ===

1. **AST分析器增强**
   - 实现一个复杂度计算器，计算函数的圈复杂度
   - 添加代码重复检测功能
   - 实现依赖关系分析器

2. **代码转换工具**
   - 实现接口到struct的自动转换
   - 创建错误处理模式转换器
   - 开发并发安全性检查和修复工具

3. **静态分析规则**
   - 实现更多代码质量规则（命名约定、注释覆盖率等）
   - 添加安全漏洞检测规则
   - 创建性能反模式检测器

4. **代码生成器**
   - 实现ORM代码生成器
   - 创建API文档生成工具
   - 开发测试用例生成器

5. **高级应用**
   - 构建完整的代码重构工具
   - 实现跨文件的依赖分析
   - 开发IDE插件集成

运行命令：
go run main.go

学习目标验证：
- 能够解析和操作Go语言AST
- 掌握访问者模式的实际应用
- 具备构建静态分析工具的能力
- 理解编译器前端的工作原理
*/
