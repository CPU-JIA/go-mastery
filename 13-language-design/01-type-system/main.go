// Package main 演示类型系统设计
// 本模块涵盖编程语言类型系统的核心概念：
// - 类型分类与层次
// - 类型推断算法
// - 泛型类型系统
// - 类型安全保证
package main

import (
	"fmt"
	"strings"
)

// ============================================================================
// 类型定义
// ============================================================================

// TypeKind 类型种类
type TypeKind int

const (
	TypeKindPrimitive TypeKind = iota // 原始类型
	TypeKindComposite                 // 复合类型
	TypeKindFunction                  // 函数类型
	TypeKindInterface                 // 接口类型
	TypeKindGeneric                   // 泛型类型
	TypeKindTypeParam                 // 类型参数
)

func (k TypeKind) String() string {
	switch k {
	case TypeKindPrimitive:
		return "Primitive"
	case TypeKindComposite:
		return "Composite"
	case TypeKindFunction:
		return "Function"
	case TypeKindInterface:
		return "Interface"
	case TypeKindGeneric:
		return "Generic"
	case TypeKindTypeParam:
		return "TypeParam"
	default:
		return "Unknown"
	}
}

// Type 类型接口
type Type interface {
	Kind() TypeKind
	Name() string
	String() string
	Equals(other Type) bool
	AssignableTo(other Type) bool
}

// ============================================================================
// 原始类型
// ============================================================================

// PrimitiveType 原始类型
type PrimitiveType struct {
	name string
	size int // 字节大小
}

func (p *PrimitiveType) Kind() TypeKind { return TypeKindPrimitive }
func (p *PrimitiveType) Name() string   { return p.name }
func (p *PrimitiveType) String() string { return p.name }

func (p *PrimitiveType) Equals(other Type) bool {
	if o, ok := other.(*PrimitiveType); ok {
		return p.name == o.name
	}
	return false
}

func (p *PrimitiveType) AssignableTo(other Type) bool {
	return p.Equals(other)
}

// 预定义原始类型
var (
	TypeInt     = &PrimitiveType{name: "int", size: 8}
	TypeInt8    = &PrimitiveType{name: "int8", size: 1}
	TypeInt16   = &PrimitiveType{name: "int16", size: 2}
	TypeInt32   = &PrimitiveType{name: "int32", size: 4}
	TypeInt64   = &PrimitiveType{name: "int64", size: 8}
	TypeFloat32 = &PrimitiveType{name: "float32", size: 4}
	TypeFloat64 = &PrimitiveType{name: "float64", size: 8}
	TypeBool    = &PrimitiveType{name: "bool", size: 1}
	TypeString  = &PrimitiveType{name: "string", size: 16}
)

// ============================================================================
// 复合类型
// ============================================================================

// StructType 结构体类型
type StructType struct {
	name   string
	fields []StructField
}

// StructField 结构体字段
type StructField struct {
	Name string
	Type Type
	Tag  string
}

func (s *StructType) Kind() TypeKind { return TypeKindComposite }
func (s *StructType) Name() string   { return s.name }

func (s *StructType) String() string {
	var fields []string
	for _, f := range s.fields {
		fields = append(fields, fmt.Sprintf("%s %s", f.Name, f.Type.String()))
	}
	return fmt.Sprintf("struct { %s }", strings.Join(fields, "; "))
}

func (s *StructType) Equals(other Type) bool {
	if o, ok := other.(*StructType); ok {
		if len(s.fields) != len(o.fields) {
			return false
		}
		for i, f := range s.fields {
			if f.Name != o.fields[i].Name || !f.Type.Equals(o.fields[i].Type) {
				return false
			}
		}
		return true
	}
	return false
}

func (s *StructType) AssignableTo(other Type) bool {
	return s.Equals(other)
}

// AddField 添加字段
func (s *StructType) AddField(name string, typ Type, tag string) {
	s.fields = append(s.fields, StructField{Name: name, Type: typ, Tag: tag})
}

// SliceType 切片类型
type SliceType struct {
	elem Type
}

func (s *SliceType) Kind() TypeKind { return TypeKindComposite }
func (s *SliceType) Name() string   { return "[]" + s.elem.Name() }
func (s *SliceType) String() string { return "[]" + s.elem.String() }

func (s *SliceType) Equals(other Type) bool {
	if o, ok := other.(*SliceType); ok {
		return s.elem.Equals(o.elem)
	}
	return false
}

func (s *SliceType) AssignableTo(other Type) bool {
	return s.Equals(other)
}

// MapType 映射类型
type MapType struct {
	key   Type
	value Type
}

func (m *MapType) Kind() TypeKind { return TypeKindComposite }
func (m *MapType) Name() string   { return fmt.Sprintf("map[%s]%s", m.key.Name(), m.value.Name()) }
func (m *MapType) String() string { return fmt.Sprintf("map[%s]%s", m.key.String(), m.value.String()) }

func (m *MapType) Equals(other Type) bool {
	if o, ok := other.(*MapType); ok {
		return m.key.Equals(o.key) && m.value.Equals(o.value)
	}
	return false
}

func (m *MapType) AssignableTo(other Type) bool {
	return m.Equals(other)
}

// ============================================================================
// 函数类型
// ============================================================================

// FunctionType 函数类型
type FunctionType struct {
	params  []Type
	results []Type
}

func (f *FunctionType) Kind() TypeKind { return TypeKindFunction }
func (f *FunctionType) Name() string   { return "func" }

func (f *FunctionType) String() string {
	var params, results []string
	for _, p := range f.params {
		params = append(params, p.String())
	}
	for _, r := range f.results {
		results = append(results, r.String())
	}

	resultStr := ""
	if len(results) == 1 {
		resultStr = " " + results[0]
	} else if len(results) > 1 {
		resultStr = " (" + strings.Join(results, ", ") + ")"
	}

	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), resultStr)
}

func (f *FunctionType) Equals(other Type) bool {
	if o, ok := other.(*FunctionType); ok {
		if len(f.params) != len(o.params) || len(f.results) != len(o.results) {
			return false
		}
		for i, p := range f.params {
			if !p.Equals(o.params[i]) {
				return false
			}
		}
		for i, r := range f.results {
			if !r.Equals(o.results[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (f *FunctionType) AssignableTo(other Type) bool {
	return f.Equals(other)
}

// ============================================================================
// 接口类型
// ============================================================================

// InterfaceType 接口类型
type InterfaceType struct {
	name    string
	methods []MethodSignature
}

// MethodSignature 方法签名
type MethodSignature struct {
	Name    string
	Params  []Type
	Results []Type
}

func (i *InterfaceType) Kind() TypeKind { return TypeKindInterface }
func (i *InterfaceType) Name() string   { return i.name }

func (i *InterfaceType) String() string {
	if i.name != "" {
		return i.name
	}
	var methods []string
	for _, m := range i.methods {
		methods = append(methods, m.Name+"()")
	}
	return fmt.Sprintf("interface { %s }", strings.Join(methods, "; "))
}

func (i *InterfaceType) Equals(other Type) bool {
	if o, ok := other.(*InterfaceType); ok {
		if len(i.methods) != len(o.methods) {
			return false
		}
		// 简化比较：只比较方法名
		for idx, m := range i.methods {
			if m.Name != o.methods[idx].Name {
				return false
			}
		}
		return true
	}
	return false
}

func (i *InterfaceType) AssignableTo(other Type) bool {
	if o, ok := other.(*InterfaceType); ok {
		// 检查是否实现了所有方法
		for _, om := range o.methods {
			found := false
			for _, im := range i.methods {
				if im.Name == om.Name {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}
	return false
}

// AddMethod 添加方法
func (i *InterfaceType) AddMethod(name string, params, results []Type) {
	i.methods = append(i.methods, MethodSignature{
		Name:    name,
		Params:  params,
		Results: results,
	})
}

// ============================================================================
// 泛型类型
// ============================================================================

// TypeParameter 类型参数
type TypeParameter struct {
	name       string
	constraint Type // 约束（通常是接口）
}

func (t *TypeParameter) Kind() TypeKind { return TypeKindTypeParam }
func (t *TypeParameter) Name() string   { return t.name }
func (t *TypeParameter) String() string {
	if t.constraint != nil {
		return fmt.Sprintf("%s %s", t.name, t.constraint.String())
	}
	return t.name
}

func (t *TypeParameter) Equals(other Type) bool {
	if o, ok := other.(*TypeParameter); ok {
		return t.name == o.name
	}
	return false
}

func (t *TypeParameter) AssignableTo(other Type) bool {
	if t.constraint != nil {
		return t.constraint.AssignableTo(other)
	}
	return true // any
}

// GenericType 泛型类型
type GenericType struct {
	name       string
	typeParams []*TypeParameter
	underlying Type
}

func (g *GenericType) Kind() TypeKind { return TypeKindGeneric }
func (g *GenericType) Name() string   { return g.name }

func (g *GenericType) String() string {
	var params []string
	for _, p := range g.typeParams {
		params = append(params, p.String())
	}
	return fmt.Sprintf("%s[%s]", g.name, strings.Join(params, ", "))
}

func (g *GenericType) Equals(other Type) bool {
	if o, ok := other.(*GenericType); ok {
		return g.name == o.name && len(g.typeParams) == len(o.typeParams)
	}
	return false
}

func (g *GenericType) AssignableTo(other Type) bool {
	return g.Equals(other)
}

// Instantiate 实例化泛型类型
func (g *GenericType) Instantiate(typeArgs []Type) (Type, error) {
	if len(typeArgs) != len(g.typeParams) {
		return nil, fmt.Errorf("类型参数数量不匹配: 期望 %d, 得到 %d",
			len(g.typeParams), len(typeArgs))
	}

	// 检查约束
	for i, arg := range typeArgs {
		param := g.typeParams[i]
		if param.constraint != nil {
			if !arg.AssignableTo(param.constraint) {
				return nil, fmt.Errorf("类型 %s 不满足约束 %s",
					arg.String(), param.constraint.String())
			}
		}
	}

	fmt.Printf("  实例化 %s 为 %s[", g.name, g.name)
	for i, arg := range typeArgs {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(arg.String())
	}
	fmt.Println("]")

	return g.underlying, nil
}

// ============================================================================
// 类型推断器
// ============================================================================

// TypeInferrer 类型推断器
type TypeInferrer struct {
	env map[string]Type
}

// NewTypeInferrer 创建类型推断器
func NewTypeInferrer() *TypeInferrer {
	return &TypeInferrer{
		env: make(map[string]Type),
	}
}

// Infer 推断表达式类型
func (ti *TypeInferrer) Infer(expr Expression) (Type, error) {
	switch e := expr.(type) {
	case *IntLiteral:
		return TypeInt, nil
	case *FloatLiteral:
		return TypeFloat64, nil
	case *StringLiteral:
		return TypeString, nil
	case *BoolLiteral:
		return TypeBool, nil
	case *Variable:
		if t, ok := ti.env[e.Name]; ok {
			return t, nil
		}
		return nil, fmt.Errorf("未定义的变量: %s", e.Name)
	case *BinaryExpr:
		return ti.inferBinary(e)
	case *CallExpr:
		return ti.inferCall(e)
	default:
		return nil, fmt.Errorf("无法推断类型")
	}
}

func (ti *TypeInferrer) inferBinary(e *BinaryExpr) (Type, error) {
	leftType, err := ti.Infer(e.Left)
	if err != nil {
		return nil, err
	}

	rightType, err := ti.Infer(e.Right)
	if err != nil {
		return nil, err
	}

	// 简化的类型推断规则
	switch e.Op {
	case "+", "-", "*", "/":
		if leftType.Equals(TypeInt) && rightType.Equals(TypeInt) {
			return TypeInt, nil
		}
		if leftType.Equals(TypeFloat64) || rightType.Equals(TypeFloat64) {
			return TypeFloat64, nil
		}
		if leftType.Equals(TypeString) && rightType.Equals(TypeString) && e.Op == "+" {
			return TypeString, nil
		}
	case "==", "!=", "<", ">", "<=", ">=":
		return TypeBool, nil
	case "&&", "||":
		if leftType.Equals(TypeBool) && rightType.Equals(TypeBool) {
			return TypeBool, nil
		}
	}

	return nil, fmt.Errorf("无法对 %s 和 %s 执行 %s 操作",
		leftType.String(), rightType.String(), e.Op)
}

func (ti *TypeInferrer) inferCall(e *CallExpr) (Type, error) {
	funcType, err := ti.Infer(e.Func)
	if err != nil {
		return nil, err
	}

	if ft, ok := funcType.(*FunctionType); ok {
		if len(ft.results) > 0 {
			return ft.results[0], nil
		}
		return nil, nil
	}

	return nil, fmt.Errorf("%s 不是函数类型", funcType.String())
}

// Define 定义变量类型
func (ti *TypeInferrer) Define(name string, typ Type) {
	ti.env[name] = typ
}

// ============================================================================
// 表达式定义
// ============================================================================

// Expression 表达式接口
type Expression interface {
	exprNode()
}

type IntLiteral struct{ Value int64 }
type FloatLiteral struct{ Value float64 }
type StringLiteral struct{ Value string }
type BoolLiteral struct{ Value bool }
type Variable struct{ Name string }
type BinaryExpr struct {
	Left  Expression
	Op    string
	Right Expression
}
type CallExpr struct {
	Func Expression
	Args []Expression
}

func (*IntLiteral) exprNode()    {}
func (*FloatLiteral) exprNode()  {}
func (*StringLiteral) exprNode() {}
func (*BoolLiteral) exprNode()   {}
func (*Variable) exprNode()      {}
func (*BinaryExpr) exprNode()    {}
func (*CallExpr) exprNode()      {}

// ============================================================================
// 演示函数
// ============================================================================

func demonstratePrimitiveTypes() {
	fmt.Println("\n=== 原始类型 ===")

	types := []Type{TypeInt, TypeInt8, TypeInt16, TypeInt32, TypeInt64,
		TypeFloat32, TypeFloat64, TypeBool, TypeString}

	fmt.Println("Go 原始类型:")
	for _, t := range types {
		pt := t.(*PrimitiveType)
		fmt.Printf("  %s: %d 字节\n", pt.Name(), pt.size)
	}
}

func demonstrateCompositeTypes() {
	fmt.Println("\n=== 复合类型 ===")

	// 结构体类型
	personType := &StructType{name: "Person"}
	personType.AddField("Name", TypeString, `json:"name"`)
	personType.AddField("Age", TypeInt, `json:"age"`)

	fmt.Printf("结构体类型: %s\n", personType.String())

	// 切片类型
	intSlice := &SliceType{elem: TypeInt}
	fmt.Printf("切片类型: %s\n", intSlice.String())

	// 映射类型
	stringIntMap := &MapType{key: TypeString, value: TypeInt}
	fmt.Printf("映射类型: %s\n", stringIntMap.String())
}

func demonstrateFunctionTypes() {
	fmt.Println("\n=== 函数类型 ===")

	// func(int, int) int
	addFunc := &FunctionType{
		params:  []Type{TypeInt, TypeInt},
		results: []Type{TypeInt},
	}
	fmt.Printf("加法函数: %s\n", addFunc.String())

	// func(string) (int, error)
	parseFunc := &FunctionType{
		params:  []Type{TypeString},
		results: []Type{TypeInt, &InterfaceType{name: "error"}},
	}
	fmt.Printf("解析函数: %s\n", parseFunc.String())

	// func()
	voidFunc := &FunctionType{}
	fmt.Printf("无参无返回: %s\n", voidFunc.String())
}

func demonstrateInterfaceTypes() {
	fmt.Println("\n=== 接口类型 ===")

	// Reader 接口
	reader := &InterfaceType{name: "Reader"}
	reader.AddMethod("Read", []Type{&SliceType{elem: TypeInt8}}, []Type{TypeInt, &InterfaceType{name: "error"}})

	fmt.Printf("Reader 接口: %s\n", reader.String())

	// Writer 接口
	writer := &InterfaceType{name: "Writer"}
	writer.AddMethod("Write", []Type{&SliceType{elem: TypeInt8}}, []Type{TypeInt, &InterfaceType{name: "error"}})

	fmt.Printf("Writer 接口: %s\n", writer.String())

	// 空接口
	anyType := &InterfaceType{name: "any"}
	fmt.Printf("空接口: %s\n", anyType.String())
}

func demonstrateGenericTypes() {
	fmt.Println("\n=== 泛型类型 ===")

	// 定义 comparable 约束
	comparable := &InterfaceType{name: "comparable"}

	// 定义泛型 List[T any]
	listType := &GenericType{
		name: "List",
		typeParams: []*TypeParameter{
			{name: "T", constraint: nil}, // any
		},
		underlying: &StructType{name: "List"},
	}
	fmt.Printf("泛型列表: %s\n", listType.String())

	// 定义泛型 Map[K comparable, V any]
	mapType := &GenericType{
		name: "Map",
		typeParams: []*TypeParameter{
			{name: "K", constraint: comparable},
			{name: "V", constraint: nil},
		},
		underlying: &StructType{name: "Map"},
	}
	fmt.Printf("泛型映射: %s\n", mapType.String())

	// 实例化泛型
	fmt.Println("\n实例化泛型类型:")
	listType.Instantiate([]Type{TypeInt})
	listType.Instantiate([]Type{TypeString})
}

func demonstrateTypeInference() {
	fmt.Println("\n=== 类型推断 ===")

	inferrer := NewTypeInferrer()

	// 定义一些变量
	inferrer.Define("x", TypeInt)
	inferrer.Define("y", TypeFloat64)
	inferrer.Define("s", TypeString)
	inferrer.Define("b", TypeBool)

	// 测试表达式
	testCases := []struct {
		desc string
		expr Expression
	}{
		{"整数字面量 42", &IntLiteral{Value: 42}},
		{"浮点字面量 3.14", &FloatLiteral{Value: 3.14}},
		{"字符串字面量", &StringLiteral{Value: "hello"}},
		{"变量 x", &Variable{Name: "x"}},
		{"x + 1", &BinaryExpr{Left: &Variable{Name: "x"}, Op: "+", Right: &IntLiteral{Value: 1}}},
		{"x > 0", &BinaryExpr{Left: &Variable{Name: "x"}, Op: ">", Right: &IntLiteral{Value: 0}}},
		{"b && true", &BinaryExpr{Left: &Variable{Name: "b"}, Op: "&&", Right: &BoolLiteral{Value: true}}},
	}

	for _, tc := range testCases {
		typ, err := inferrer.Infer(tc.expr)
		if err != nil {
			fmt.Printf("  %s => 错误: %v\n", tc.desc, err)
		} else {
			fmt.Printf("  %s => %s\n", tc.desc, typ.String())
		}
	}
}

func main() {
	fmt.Println("=== 类型系统设计 ===")
	fmt.Println()
	fmt.Println("本模块演示编程语言类型系统的核心概念:")
	fmt.Println("1. 原始类型")
	fmt.Println("2. 复合类型")
	fmt.Println("3. 函数类型")
	fmt.Println("4. 接口类型")
	fmt.Println("5. 泛型类型")
	fmt.Println("6. 类型推断")

	demonstratePrimitiveTypes()
	demonstrateCompositeTypes()
	demonstrateFunctionTypes()
	demonstrateInterfaceTypes()
	demonstrateGenericTypes()
	demonstrateTypeInference()

	fmt.Println("\n=== 类型系统演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 类型系统是编程语言的核心组件")
	fmt.Println("- Go 使用结构化类型系统（鸭子类型）")
	fmt.Println("- 泛型通过类型参数实现代码复用")
	fmt.Println("- 类型推断减少显式类型声明")
	fmt.Println("- 类型安全在编译时捕获错误")
}
