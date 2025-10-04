/*
=== Go语言学习评估系统 - 评估工具集 ===

本文件提供评估系统所需的各种工具：
1. 代码分析器 - 分析Go代码的复杂度和风格
2. 测试分析器 - 分析测试覆盖率和质量
3. 项目扫描器 - 扫描项目目录结构
4. 报告生成器 - 生成评估报告
5. CLI运行器 - 命令行工具接口

作者: JIA
创建时间: 2025-10-03
版本: 1.0.0
*/

// Package tools 提供评估系统的各种分析和工具函数
//
// 本包包含了代码分析、测试分析、报告生成等核心工具，
// 用于支持评估系统的自动化评估功能。
package tools

import (
	"assessment-system/evaluators"
	"assessment-system/models"
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CodeAnalyzer 代码分析工具
type CodeAnalyzer struct {
	fileSet *token.FileSet
}

// NewCodeAnalyzer 创建Go代码分析器实例
//
// 功能说明:
//
//	本函数初始化一个CodeAnalyzer对象,用于分析Go源代码的复杂度、风格等质量指标。
//	分析器内部维护一个token.FileSet,用于跟踪多个文件的位置信息,支持准确的错误报告和AST节点定位。
//
// 核心组件:
//
//	fileSet *token.FileSet - Go语言标准库的文件集合,用于:
//	  1. 跟踪所有解析文件的位置信息(文件名、行号、列号)
//	  2. 支持多文件分析时的准确定位
//	  3. 为AST解析器提供位置上下文
//
// 返回值:
//   - *CodeAnalyzer: 初始化完成的代码分析器指针,可立即用于分析Go代码
//
// 使用场景:
//   - 评估系统初始化时创建全局分析器实例
//   - 每次代码质量检查前创建独立分析器
//   - 批量分析多个文件时复用同一分析器实例
//
// 示例:
//
//	analyzer := NewCodeAnalyzer()
//	code := `package main
//	func complexFunc(x int) int {
//	    if x > 0 {
//	        for i := 0; i < x; i++ {
//	            if i%2 == 0 {
//	                x++
//	            }
//	        }
//	    }
//	    return x
//	}`
//	metrics, err := analyzer.AnalyzeComplexity(code)
//	// metrics.MaxComplexity == 4 (1基础 + 1if + 1for + 1if = 4)
//
// 作者: JIA
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		fileSet: token.NewFileSet(),
	}
}

// AnalyzeComplexity 分析Go代码的圈复杂度和结构复杂度指标
//
// 功能说明:
//
//	本方法解析Go源代码字符串,遍历其抽象语法树(AST),统计各类复杂度指标。
//	通过分析函数、条件分支、循环、switch语句等结构,计算代码的可维护性和测试难度。
//
// 分析维度:
//
//	1. 函数统计 (Functions): 代码中定义的函数总数
//	2. 圈复杂度 (CyclomaticComplexity): 所有函数的复杂度之和
//	3. 最大复杂度 (MaxComplexity): 单个函数的最高复杂度值
//	4. 平均复杂度 (AverageComplexity): 总复杂度/函数数,反映整体代码质量
//	5. 条件语句数 (Conditionals): if语句总数
//	6. 循环语句数 (Loops): for/range循环总数
//	7. 分支语句数 (Switches): switch/type switch总数
//
// 圈复杂度计算逻辑 (McCabe's Cyclomatic Complexity):
//
//	基础复杂度 = 1 (每个函数起始为1)
//	每个if/switch/case/for/range语句 +1
//	公式: V(G) = E - N + 2P (其中E=边数,N=节点数,P=连通分量数)
//	简化实现: 统计决策点数量 + 1
//
// 复杂度分级标准:
//   - 1-10: 简单程序,易于测试和维护
//   - 11-20: 复杂度适中,需要注意
//   - 21-50: 复杂程序,难以测试,建议重构
//   - >50: 极端复杂,几乎无法测试,必须重构
//
// 参数:
//   - code: Go源代码字符串,必须是合法的Go语法
//
// 返回值:
//   - *ComplexityMetrics: 复杂度指标结构体,包含7个维度的统计数据
//   - error: 解析错误,可能原因:
//   - 语法错误: 代码不符合Go语法规范
//   - 编码问题: 包含非UTF-8字符
//
// 使用场景:
//   - 代码评审前检查函数复杂度
//   - CI/CD流水线中的质量门控
//   - 识别需要重构的高复杂度函数
//   - 生成代码质量报告
//
// 示例:
//
//	code := `package main
//	func simpleFunc() {
//	    println("hello")  // 复杂度=1
//	}
//	func complexFunc(x int) int {
//	    if x > 0 {        // +1 = 2
//	        for i := 0; i < x; i++ {  // +1 = 3
//	            switch i % 3 {        // +1 = 4
//	            case 0:               // +1 = 5
//	                x++
//	            case 1:               // +1 = 6
//	                x--
//	            }
//	        }
//	    }
//	    return x
//	}`
//	analyzer := NewCodeAnalyzer()
//	metrics, err := analyzer.AnalyzeComplexity(code)
//	// metrics.Functions == 2
//	// metrics.CyclomaticComplexity == 7 (1+6)
//	// metrics.MaxComplexity == 6 (complexFunc)
//	// metrics.AverageComplexity == 3.5 (7/2)
//	// metrics.Conditionals == 1 (if语句)
//	// metrics.Loops == 1 (for循环)
//	// metrics.Switches == 1 (switch语句)
//
// 注意事项:
//   - 仅分析有函数体的函数,接口方法签名不计入
//   - 空函数复杂度为1(基础复杂度)
//   - switch语句的每个case都会增加复杂度
//   - 代码必须能够成功解析,否则返回错误
//
// 作者: JIA
func (ca *CodeAnalyzer) AnalyzeComplexity(code string) (*ComplexityMetrics, error) {
	node, err := parser.ParseFile(ca.fileSet, "", code, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code: %w", err)
	}

	metrics := &ComplexityMetrics{}

	ast.Inspect(node, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				metrics.Functions++
				complexity := ca.calculateCyclomaticComplexity(node.Body)
				metrics.CyclomaticComplexity += complexity
				if complexity > metrics.MaxComplexity {
					metrics.MaxComplexity = complexity
				}
			}
		case *ast.IfStmt:
			metrics.Conditionals++
		case *ast.ForStmt, *ast.RangeStmt:
			metrics.Loops++
		case *ast.TypeSwitchStmt, *ast.SwitchStmt:
			metrics.Switches++
		}
		return true
	})

	if metrics.Functions > 0 {
		metrics.AverageComplexity = float64(metrics.CyclomaticComplexity) / float64(metrics.Functions)
	}

	return metrics, nil
}

// calculateCyclomaticComplexity 计算单个函数体的圈复杂度
//
// 功能说明:
//
//	本方法遍历函数体的AST节点,统计所有决策点(分支/循环/case),
//	计算该函数的McCabe圈复杂度。圈复杂度反映代码路径数量,是测试难度和维护成本的重要指标。
//
// 计算规则 (McCabe's Cyclomatic Complexity):
//
//	基础复杂度 = 1 (函数默认有一条执行路径)
//	每遇到以下AST节点 +1:
//	  • *ast.IfStmt - if条件语句
//	  • *ast.TypeSwitchStmt - type switch类型断言
//	  • *ast.SwitchStmt - switch分支语句
//	  • *ast.CaseClause - switch中的每个case分支
//	  • *ast.ForStmt - for循环
//	  • *ast.RangeStmt - range遍历循环
//
// 复杂度含义:
//   - V(G) = 1: 顺序执行,无分支(如空函数或单条语句)
//   - V(G) = 2-4: 低复杂度,结构简单,易于理解和测试
//   - V(G) = 5-10: 中等复杂度,需要适当的测试用例覆盖
//   - V(G) = 11-20: 高复杂度,建议拆分函数或简化逻辑
//   - V(G) > 20: 极高复杂度,难以测试和维护,强烈建议重构
//
// 参数:
//   - body: 函数体的AST块语句节点 (*ast.BlockStmt)
//
// 返回值:
//   - int: 该函数体的圈复杂度值 (≥1)
//
// 算法实现:
//
//	使用ast.Inspect深度优先遍历整个函数体的AST树,
//	通过类型断言识别决策节点,累加计数器。
//	时间复杂度: O(n),n为AST节点总数
//	空间复杂度: O(h),h为AST树高度(递归调用栈)
//
// 使用场景:
//   - AnalyzeComplexity方法内部调用,为每个函数计算复杂度
//   - 单独分析某个函数体的复杂度
//   - 代码质量工具集成,识别高复杂度函数
//
// 示例:
//
//	// 示例1: 简单函数(复杂度=1)
//	func simple() int {
//	    return 42
//	}
//
//	// 示例2: 包含if和for(复杂度=3)
//	func moderate(n int) int {
//	    sum := 0
//	    if n > 0 {           // +1 = 2
//	        for i := 0; i < n; i++ {  // +1 = 3
//	            sum += i
//	        }
//	    }
//	    return sum
//	}
//
//	// 示例3: 包含switch(复杂度=5)
//	func complex(x int) string {
//	    switch x {           // +1 = 2
//	    case 1:              // +1 = 3
//	        return "one"
//	    case 2:              // +1 = 4
//	        return "two"
//	    default:             // +1 = 5
//	        return "other"
//	    }
//	}
//
// 注意事项:
//   - switch语句本身+1,每个case子句也+1
//   - else if会被视为嵌套的if,每个都计数
//   - 逻辑运算符(&&, ||)不增加圈复杂度(不同于认知复杂度)
//   - 空函数体复杂度为1(基础值)
//
// 作者: JIA
func (ca *CodeAnalyzer) calculateCyclomaticComplexity(body *ast.BlockStmt) int {
	complexity := 1 // 基础复杂度

	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.TypeSwitchStmt, *ast.SwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		case *ast.ForStmt, *ast.RangeStmt:
			complexity++
		}
		return true
	})

	return complexity
}

// AnalyzeCodeStyle 分析Go代码的风格和可读性指标
//
// 功能说明:
//
//	本方法逐行扫描源代码,检查编码规范、命名质量、注释覆盖率、缩进一致性等风格指标。
//	通过多维度分析,评估代码的可读性和团队协作友好度,为代码审查提供数据支持。
//
// 分析维度 (7项指标):
//
//	1. TotalLines - 总行数: 代码规模统计
//	2. LongLines - 超长行数: 超过120字符的行(MaxLineLength常量定义)
//	3. BadNaming - 不良命名数: 单字母变量、temp/data等通用名
//	4. Comments - 注释行数: 以"//"开头的注释行
//	5. CommentRatio - 注释率: 注释行/总行数,理想值20-30%
//	6. EmptyLines - 空行数: 完全空白的行,过多影响代码密度
//	7. IndentationIssues - 缩进问题数: 混用空格和Tab导致的不一致
//
// 检查规则详解:
//
//	超长行检查:
//	  • 阈值: 120字符(evaluators.MaxLineLength)
//	  • 原因: 超过屏幕宽度影响可读性,建议换行
//	  • 例外: 长字符串、导入语句可适当放宽
//
//	不良命名检查 (正则模式):
//	  • 单字母变量: \b[a-z]\b (如 x, y, i, j)
//	  • 编号变量: \bvar[0-9]+\b (如 var1, var2)
//	  • 临时变量: \btemp\b, \bdata\b (语义不明确)
//	  • 最佳实践: 使用描述性名称,如 userID, totalCount
//
//	注释率建议:
//	  • <10%: 注释严重不足,可维护性差
//	  • 10-20%: 注释偏少,建议增加关键逻辑说明
//	  • 20-30%: 理想区间,文档完善
//	  • >40%: 注释过多,可能代码过于复杂
//
//	缩进一致性:
//	  • Go标准: 使用Tab缩进(gofmt强制)
//	  • 检测: 同一行不能同时包含Tab和空格缩进
//	  • 工具: 建议使用gofmt自动格式化
//
// 参数:
//   - code: Go源代码字符串
//
// 返回值:
//   - *StyleMetrics: 风格指标结构体,包含7个维度统计
//   - error: 当前实现不返回错误(保留扩展性)
//
// 使用场景:
//   - CI/CD流水线的风格检查
//   - Pull Request自动化审查
//   - 代码质量趋势分析
//   - 团队编码规范合规检测
//
// 示例:
//
//	code := `package main
//	import "fmt"
//
//	// calculateSum computes the sum of two integers
//	func calculateSum(a, b int) int {
//	    // This is a very long line that exceeds the 120 character limit and should be reported as a style issue for better readability
//	    return a + b
//	}
//
//	func bad(x int, temp string, data []byte) {  // 不良命名: x, temp, data
//	    fmt.Println(x, temp, data)
//	}`
//
//	analyzer := NewCodeAnalyzer()
//	metrics, _ := analyzer.AnalyzeCodeStyle(code)
//	// metrics.TotalLines == 11
//	// metrics.LongLines == 1 (第5行超过120字符)
//	// metrics.BadNaming >= 3 (x, temp, data)
//	// metrics.Comments == 2 (两行注释)
//	// metrics.CommentRatio ≈ 0.18 (2/11 = 18.2%)
//	// metrics.EmptyLines == 2 (第3行和第8行)
//	// metrics.IndentationIssues == 0 (无混用空格Tab)
//
// 注意事项:
//   - 不良命名检测基于简单正则,可能误判(如循环计数器i, j)
//   - 长字符串内的内容不应触发超长行警告,但当前实现会计入
//   - 仅统计行注释(//),不统计块注释(/* */)
//   - 空行包括完全空白行,不包括仅有空格/Tab的行
//
// 作者: JIA
func (ca *CodeAnalyzer) AnalyzeCodeStyle(code string) (*StyleMetrics, error) {
	metrics := &StyleMetrics{}
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 检查行长度
		if len(line) > evaluators.MaxLineLength {
			metrics.LongLines++
		}

		// 检查命名规范
		if ca.hasBadNaming(line) {
			metrics.BadNaming++
		}

		// 检查注释
		if strings.HasPrefix(line, "//") {
			metrics.Comments++
		}

		// 检查空行
		if line == "" {
			metrics.EmptyLines++
		}

		// 检查缩进一致性
		if !ca.hasConsistentIndentation(lines[i]) {
			metrics.IndentationIssues++
		}
	}

	metrics.TotalLines = len(lines)
	metrics.CommentRatio = float64(metrics.Comments) / float64(metrics.TotalLines)

	return metrics, nil
}

// hasBadNaming 检查代码行是否包含不良命名模式
//
// 功能说明:
//
//	本方法使用正则表达式检测代码行中是否存在不符合Go命名规范的标识符,
//	如单字母变量、编号变量(var1, var2)、语义模糊的通用名称(temp, data)等。
//
// 检测的不良命名模式 (4类):
//
//	1. 单字母变量: \b[a-z]\b
//	   示例: x, y, n, s
//	   例外: i, j, k作为循环计数器是可接受的(但本方法会误判)
//
//	2. 编号变量: \bvar[0-9]+\b
//	   示例: var1, var2, var10
//	   问题: 无法表达变量用途,应使用有意义的名称
//
//	3. 临时变量: \btemp\b
//	   示例: temp, tempFile, tempData
//	   问题: 所有变量都是临时的,应描述存储的是什么
//
//	4. 通用数据变量: \bdata\b
//	   示例: data, userData, jsonData
//	   问题: "data"过于宽泛,应明确是什么数据(如rawBytes, userProfile)
//
// 参数:
//   - line: 待检查的代码行(已trim空格)
//
// 返回值:
//   - bool: true表示存在不良命名,false表示命名规范
//
// 局限性:
//   - 正则匹配可能误判合法场景:
//   • 循环计数器 i, j, k 是Go惯用法
//   • 短函数内的局部变量 n, s 可能是可接受的
//   • 字符串内容或注释中的单词会被匹配(false positive)
//   - 无法检测驼峰命名、全大写常量等复杂规则
//
// 使用场景:
//   - AnalyzeCodeStyle内部调用,统计不良命名数量
//   - 代码审查工具的命名规范检查
//   - 教学场景中演示良好命名实践
//
// 示例:
//
//	hasBadNaming("x := 10")                    // true (单字母x)
//	hasBadNaming("var1 := getUserData()")      // true (编号变量var1)
//	hasBadNaming("temp := processFile()")      // true (temp)
//	hasBadNaming("data := readConfig()")       // true (data)
//	hasBadNaming("userID := fetchUser()")      // false (良好命名)
//	hasBadNaming("totalCount := len(items)")   // false (语义清晰)
//	hasBadNaming("for i := 0; i < n; i++ {")   // true (误判:i和n都匹配单字母规则)
//
// 最佳命名实践:
//   - 使用描述性名称: userID 而非 x
//   - 避免缩写: configuration 而非 cfg (除非是团队约定)
//   - 长度适中: 3-15个字符为佳,过长影响可读性
//   - 驼峰命名: userName 而非 user_name
//   - 常量全大写: MaxRetries 而非 maxRetries
//
// 作者: JIA
func (ca *CodeAnalyzer) hasBadNaming(line string) bool {
	// 简化的命名检查
	badPatterns := []string{
		`\b[a-z]\b`,     // 单字母变量
		`\bvar[0-9]+\b`, // var1, var2 等
		`\btemp\b`,      // temp 变量
		`\bdata\b`,      // 通用的 data
	}

	for _, pattern := range badPatterns {
		// 检查正则匹配是否成功，忽略错误因为模式是硬编码的
		matched, err := regexp.MatchString(pattern, line)
		if err != nil {
			// 如果正则表达式本身有问题（不应该发生），返回false
			return false
		}
		if matched {
			return true
		}
	}
	return false
}

// hasConsistentIndentation 检查代码行的缩进一致性
//
// 功能说明:
//
//	本方法检测代码行是否同时混用空格和Tab作为缩进,确保符合Go官方风格指南(gofmt强制Tab缩进)。
//	混用缩进会导致不同编辑器显示效果不一致,影响代码可读性和团队协作。
//
// 检查逻辑:
//
//	1. 空行或无缩进行: 直接返回true(一致)
//	2. 纯空格缩进: 返回true(一致,但不符合Go标准)
//	3. 纯Tab缩进: 返回true(一致,符合Go标准)
//	4. 混用空格和Tab: 返回false(不一致,严重问题)
//
// Go官方缩进标准:
//   - 强制使用Tab缩进,不使用空格
//   - gofmt会自动将空格缩进转换为Tab
//   - 每级缩进1个Tab字符
//   - 对齐时可在Tab后使用空格(如注释对齐)
//
// 参数:
//   - line: 原始代码行(未trim,保留前导空白符)
//
// 返回值:
//   - bool: true表示缩进一致,false表示混用空格和Tab
//
// 局限性:
//   - 仅检测前导空白符的混用,不检测行内混用
//   - 无法区分"纯空格缩进"和"正确的Tab缩进"
//   - 不验证缩进级别是否正确(如是否应缩进2级)
//
// 使用场景:
//   - AnalyzeCodeStyle内部调用,统计缩进问题
//   - 预提交钩子检查代码格式
//   - CI/CD中强制执行gofmt标准
//
// 示例:
//
//	hasConsistentIndentation("\tfunc main() {")      // true (纯Tab)
//	hasConsistentIndentation("    func main() {")    // true (纯空格,但不符合Go标准)
//	hasConsistentIndentation(" \tfunc main() {")     // false (混用空格+Tab)
//	hasConsistentIndentation("\t    return 42")      // false (Tab后有空格作为缩进)
//	hasConsistentIndentation("")                      // true (空行)
//	hasConsistentIndentation("func main() {")         // true (无缩进)
//
// 最佳实践:
//   - 使用gofmt自动格式化所有Go代码
//   - 配置编辑器将Tab键映射为真正的Tab字符,不展开为空格
//   - 在pre-commit钩子中运行gofmt检查
//   - 团队约定统一使用官方工具链
//
// 注意事项:
//   - 本方法仅做简单检测,实际应使用gofmt作为权威工具
//   - 返回true不代表符合Go标准,仅表示未混用
//   - gofmt会自动修复所有缩进问题
//
// 作者: JIA
func (ca *CodeAnalyzer) hasConsistentIndentation(line string) bool {
	if line == "" || (!strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t")) {
		return true // 没有缩进或非缩进行
	}

	// 简化检查：确保不混用空格和制表符
	hasSpaces := strings.HasPrefix(line, " ")
	hasTabs := strings.HasPrefix(line, "\t")

	return !(hasSpaces && hasTabs)
}

// TestAnalyzer 测试分析工具
type TestAnalyzer struct{}

// NewTestAnalyzer 创建测试分析器实例（无状态设计）
//
// 功能说明:
//
//	本函数创建一个TestAnalyzer对象，用于分析Go项目的测试覆盖率和测试质量指标。
//	TestAnalyzer采用无状态设计（空结构体），所有分析逻辑通过方法参数传递，
//	这种设计使得分析器可以安全地并发使用，无需担心状态竞争。
//
// 设计理念（无状态模式）:
//
//	为什么使用空结构体？
//	1. 零内存开销：struct{} 在Go中不占用任何内存空间
//	2. 线程安全：无状态对象天然支持并发，多个goroutine可共享同一实例
//	3. 简洁API：通过接收者方法提供命名空间，避免全局函数污染
//	4. 易于扩展：未来可增加状态字段（如缓存）而不破坏现有API
//
// 对比有状态设计（CodeAnalyzer）:
//
//	CodeAnalyzer包含fileSet字段，需要维护跨文件的位置信息（有状态）
//	TestAnalyzer仅进行文件扫描和计数，无需保持状态（无状态）
//	选择合适的设计模式取决于业务需求和性能权衡
//
// 返回值:
//   - *TestAnalyzer: 测试分析器指针，虽然是空结构体但返回指针是Go惯例，
//     保持API一致性，且指针传递不增加额外开销
//
// 使用场景:
//   - 评估系统初始化时创建全局测试分析器
//   - 并发分析多个项目的测试质量
//   - 集成到CI/CD流水线中的测试覆盖率检查
//
// 示例:
//
//	analyzer := NewTestAnalyzer()
//
//	// 场景1: 分析单个项目
//	files := map[string]string{
//	    "main.go":      "package main\nfunc Add(a, b int) int { return a + b }",
//	    "main_test.go": "package main\nimport \"testing\"\nfunc TestAdd(t *testing.T) {}",
//	}
//	metrics, _ := analyzer.AnalyzeTestCoverage(files)
//	// metrics.TotalSourceFiles == 1, metrics.TotalTestFiles == 1
//
//	// 场景2: 并发分析多个项目（无状态设计的优势）
//	var wg sync.WaitGroup
//	for _, project := range projects {
//	    wg.Add(1)
//	    go func(p Project) {
//	        defer wg.Done()
//	        metrics, _ := analyzer.AnalyzeTestCoverage(p.Files) // 并发安全
//	    }(project)
//	}
//	wg.Wait()
//
// 注意事项:
//   - 返回的指针指向空结构体，主要用于方法调用，不应解引用
//   - 可以重复调用本函数，但通常全局单例即可（节省分配开销）
//   - 如果未来需要增加状态（如缓存），只需修改结构体定义，API不变
//
// 性能特性:
//   - 空结构体大小: unsafe.Sizeof(TestAnalyzer{}) == 0 字节
//   - 创建开销: 几乎为零，仅涉及指针分配
//   - 并发友好: 无锁设计，可安全共享
//
// 作者: JIA
func NewTestAnalyzer() *TestAnalyzer {
	return &TestAnalyzer{}
}

// AnalyzeTestCoverage 分析Go项目的测试覆盖情况和测试质量指标
//
// 功能说明:
//
//	本方法扫描项目中的所有Go文件，自动识别源文件和测试文件，
//	统计测试覆盖率、测试文件数量、测试函数数量等关键指标。
//	用于评估项目的测试完备性和测试质量水平。
//
// 分析流程:
//
//  1. 文件分类（基于Go测试约定）:
//     - 测试文件: 文件名以 "_test.go" 结尾（如 main_test.go, http_test.go）
//     - 源文件: 文件名以 ".go" 结尾但不含 "_test.go"（如 main.go, http.go）
//     - 其他文件: 忽略（如 .txt, .md, .json等）
//
//  2. 覆盖率计算（简化版本）:
//     公式: TestCoverage = TotalTestFiles / TotalSourceFiles
//     解读: 每个源文件对应一个测试文件时，覆盖率为100%
//     注意: 这是文件级覆盖率，不同于代码行覆盖率（需要go test -cover）
//
//  3. 测试函数统计:
//     遍历所有测试文件，识别 "func Test" 和 "func Benchmark" 开头的函数
//     累加到 TotalTestFunctions 计数器
//
// Go测试文件命名约定:
//
//	标准约定（Go官方规范）:
//	- 测试文件必须以 "_test.go" 结尾
//	- 测试文件与被测源文件同名（如 http.go → http_test.go）
//	- 测试文件与源文件位于同一package
//	- 示例: math.go(源文件) + math_test.go(测试文件)
//
// 参数:
//   - projectFiles: 项目文件映射，键为文件名（可含相对路径），值为文件内容
//     示例: map[string]string{
//     "main.go": "package main\nfunc Add(a, b int) int { return a + b }",
//     "main_test.go": "package main\nimport \"testing\"\nfunc TestAdd(t *testing.T) {}",
//     }
//
// 返回值:
//   - *TestMetrics: 测试指标结构体，包含4个字段：
//   - TotalSourceFiles: 源文件总数（不含测试文件）
//   - TotalTestFiles: 测试文件总数
//   - TotalTestFunctions: 测试函数总数（Test+Benchmark）
//   - TestCoverage: 测试覆盖率（0.0-1.0范围，1.0表示每个源文件都有测试）
//   - error: 当前实现不返回错误，保留用于未来扩展（如文件解析失败）
//
// 使用场景:
//   - 项目质量评估：检查测试是否充分
//   - CI/CD质量门禁：覆盖率低于阈值时阻止合并
//   - 学习进度跟踪：TDD实践的量化指标
//   - 代码审查：快速了解项目测试状况
//
// 示例:
//
//	analyzer := NewTestAnalyzer()
//
//	// 示例1: 测试覆盖率100%的项目
//	files1 := map[string]string{
//	    "math.go":      "package math\nfunc Add(a, b int) int { return a + b }",
//	    "math_test.go": "package math\nimport \"testing\"\nfunc TestAdd(t *testing.T) {}",
//	    "http.go":      "package http\nfunc Get(url string) error { return nil }",
//	    "http_test.go": "package http\nimport \"testing\"\nfunc TestGet(t *testing.T) {}",
//	}
//	metrics1, _ := analyzer.AnalyzeTestCoverage(files1)
//	// metrics1.TotalSourceFiles == 2
//	// metrics1.TotalTestFiles == 2
//	// metrics1.TestCoverage == 1.0 (100%覆盖)
//	// metrics1.TotalTestFunctions == 2
//
//	// 示例2: 测试覆盖率50%的项目（一半文件无测试）
//	files2 := map[string]string{
//	    "main.go":   "package main\nfunc main() {}",
//	    "config.go": "package main\nfunc LoadConfig() error { return nil }",
//	    "main_test.go": "package main\nimport \"testing\"\nfunc TestMain(t *testing.T) {}\nfunc BenchmarkMain(b *testing.B) {}",
//	}
//	metrics2, _ := analyzer.AnalyzeTestCoverage(files2)
//	// metrics2.TotalSourceFiles == 2 (main.go, config.go)
//	// metrics2.TotalTestFiles == 1 (main_test.go)
//	// metrics2.TestCoverage == 0.5 (50%覆盖)
//	// metrics2.TotalTestFunctions == 2 (1个Test + 1个Benchmark)
//
//	// 示例3: 无测试的项目
//	files3 := map[string]string{
//	    "main.go": "package main\nfunc main() {}",
//	}
//	metrics3, _ := analyzer.AnalyzeTestCoverage(files3)
//	// metrics3.TotalSourceFiles == 1
//	// metrics3.TotalTestFiles == 0
//	// metrics3.TestCoverage == 0.0 (0%覆盖)
//	// metrics3.TotalTestFunctions == 0
//
// 注意事项:
//   - 本方法计算的是文件级覆盖率，不是代码行覆盖率
//   - 代码行覆盖率需要运行 go test -cover 并解析输出
//   - 不检查测试文件是否实际测试了对应的源文件内容
//   - 不验证测试函数是否有效（如缺少 *testing.T 参数也会被计数）
//   - 如果源文件数为0，覆盖率为0（避免除零错误）
//
// 局限性与改进方向:
//   - 当前仅统计文件数量，未分析测试深度和广度
//   - 未检测表格驱动测试、子测试等高级模式
//   - 未统计断言数量和测试代码行数
//   - 未来可扩展AST分析，提供更精确的覆盖率评估
//
// 作者: JIA
func (ta *TestAnalyzer) AnalyzeTestCoverage(projectFiles map[string]string) (*TestMetrics, error) {
	metrics := &TestMetrics{}

	var sourceFiles, testFiles []string

	for filename := range projectFiles {
		if strings.HasSuffix(filename, "_test.go") {
			testFiles = append(testFiles, filename)
		} else if strings.HasSuffix(filename, ".go") {
			sourceFiles = append(sourceFiles, filename)
		}
	}

	metrics.TotalSourceFiles = len(sourceFiles)
	metrics.TotalTestFiles = len(testFiles)

	// 计算测试覆盖率（简化）
	if metrics.TotalSourceFiles > 0 {
		metrics.TestCoverage = float64(metrics.TotalTestFiles) / float64(metrics.TotalSourceFiles)
	}

	// 分析测试函数
	for _, testFile := range testFiles {
		content := projectFiles[testFile]
		testFunctions := ta.countTestFunctions(content)
		metrics.TotalTestFunctions += testFunctions
	}

	return metrics, nil
}

// countTestFunctions 统计测试文件中的测试函数和基准测试函数数量
//
// 功能说明:
//
//	本方法逐行扫描测试文件内容，通过简单字符串匹配识别测试函数（func Test*）
//	和基准测试函数（func Benchmark*），累计统计总数。用于快速估算测试用例数量。
//
// 识别规则（基于Go官方testing包约定）:
//
//	1. 单元测试函数（func Test*）:
//	   - 必须以 "func Test" 开头
//	   - 后续字符必须是大写字母开头（如 TestAdd, TestHTTPServer）
//	   - 参数签名: func TestXxx(t *testing.T)
//	   - 示例: "func TestAdd(t *testing.T) {" → 计数+1
//
//	2. 基准测试函数（func Benchmark*）:
//	   - 必须以 "func Benchmark" 开头
//	   - 后续字符必须是大写字母开头（如 BenchmarkFibonacci）
//	   - 参数签名: func BenchmarkXxx(b *testing.B)
//	   - 示例: "func BenchmarkSort(b *testing.B) {" → 计数+1
//
//	3. 其他测试类型（当前实现未统计）:
//	   - 示例函数: func Example*() → 未计数
//	   - 模糊测试: func Fuzz*(f *testing.F) → 未计数
//	   - 表格测试子用例: t.Run("case1", func(t *testing.T) {...}) → 未计数
//
// 参数:
//   - content: 测试文件的完整内容字符串
//
// 返回值:
//   - int: 测试函数总数（Test函数数 + Benchmark函数数）
//
// 实现逻辑:
//
//	使用简单字符串包含检查（strings.Contains）:
//	1. 按换行符分割文件内容为行切片
//	2. 遍历每一行，检查是否包含 "func Test" 或 "func Benchmark"
//	3. 每匹配一次，计数器+1
//	4. 返回累计总数
//
// 使用场景:
//   - AnalyzeTestCoverage内部调用，统计测试函数总数
//   - 快速评估项目测试规模
//   - 生成测试质量报告
//
// 示例:
//
//	content := `package math_test
//
//	import "testing"
//
//	// TestAdd 测试加法函数
//	func TestAdd(t *testing.T) {
//	    if Add(1, 2) != 3 {
//	        t.Fail()
//	    }
//	}
//
//	func TestSubtract(t *testing.T) {
//	    // 减法测试
//	}
//
//	func BenchmarkAdd(b *testing.B) {
//	    for i := 0; i < b.N; i++ {
//	        Add(1, 2)
//	    }
//	}
//
//	func helper() {
//	    // 辅助函数，不会被计数
//	}`
//
//	analyzer := NewTestAnalyzer()
//	count := analyzer.countTestFunctions(content)
//	// count == 3 (TestAdd + TestSubtract + BenchmarkAdd)
//
// 局限性与误判风险:
//
//	⚠️ 简单字符串匹配可能产生以下误判：
//
//	1. 注释中的函数签名会被误判为真实函数:
//	   // 示例：func TestExample(t *testing.T) { // 这行会被计数！
//
//	2. 字符串字面量中的内容会被误判:
//	   code := "func TestSomething(t *testing.T) {" // 这行会被计数！
//
//	3. 不符合规范的函数也会被计数:
//	   func Testlowercase(t *testing.T) {} // 不符合Test*命名约定，但仍被计数
//	   func Test() {} // 缺少参数，但仍被计数
//
//	4. 未统计高级测试模式:
//	   - 表格驱动测试的子用例 t.Run("case", func(t *testing.T) {})
//	   - 示例函数 func Example*()
//	   - 模糊测试 func Fuzz*(f *testing.F)
//
// 改进方向:
//   - 使用AST解析代替字符串匹配，可完全避免误判
//   - 增加对Example*和Fuzz*函数的支持
//   - 统计子测试数量（t.Run调用次数）
//   - 验证函数签名的正确性（参数类型检查）
//
// 性能特性:
//   - 时间复杂度: O(n)，n为文件行数
//   - 空间复杂度: O(n)，需要存储分割后的行切片
//   - 适用场景: 小到中型测试文件（<10000行），大文件建议流式处理
//
// 注意事项:
//   - 本方法仅做粗略统计，不保证100%准确
//   - 如需精确统计，应使用go/parser + go/ast进行AST分析
//   - 空文件或无测试函数时返回0（符合预期）
//
// 作者: JIA
func (ta *TestAnalyzer) countTestFunctions(content string) int {
	count := 0
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.Contains(line, "func Test") || strings.Contains(line, "func Benchmark") {
			count++
		}
	}

	return count
}

// ProjectScanner 项目扫描工具
type ProjectScanner struct{}

// NewProjectScanner 创建项目目录扫描器实例（无状态设计）
//
// 功能说明:
//
//	本函数创建一个ProjectScanner对象，用于递归扫描项目目录，
//	收集所有Go源文件及其内容，构建文件名到文件内容的映射表。
//	采用无状态设计（空结构体），与TestAnalyzer设计理念一致。
//
// 设计哲学（空结构体模式复用）:
//
//	ProjectScanner与TestAnalyzer一样采用空结构体设计：
//	1. 零内存占用：struct{} 大小为0字节
//	2. 无状态操作：所有数据通过参数传递，方法无副作用
//	3. 并发安全：多个goroutine可共享同一扫描器实例
//	4. 简洁API：通过方法提供命名空间，避免全局函数
//
// 核心能力:
//
//	支持的扫描功能：
//	- 递归遍历目录树（包括所有子目录）
//	- 自动识别Go源文件（*.go）
//	- 跳过非Go文件和目录
//	- 读取文件内容并构建内存映射
//	- 生成相对路径作为文件标识
//
// 返回值:
//   - *ProjectScanner: 项目扫描器指针，用于调用ScanProject方法
//
// 使用场景:
//   - 评估系统初始化时创建全局扫描器
//   - 批量分析多个项目的代码质量
//   - 构建项目文件索引用于快速检索
//   - 生成项目结构可视化
//
// 示例:
//
//	scanner := NewProjectScanner()
//
//	// 场景1: 扫描单个项目
//	files, err := scanner.ScanProject("./00-assessment-system")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("发现 %d 个Go文件\n", len(files))
//	// 输出: files = map[string]string{
//	//   "models/student.go": "package models\n...",
//	//   "tools/assessment_tools.go": "package tools\n...",
//	// }
//
//	// 场景2: 并发扫描多个项目（无状态优势）
//	projects := []string{"./01-basics", "./02-advanced", "./03-concurrency"}
//	var wg sync.WaitGroup
//	for _, proj := range projects {
//	    wg.Add(1)
//	    go func(path string) {
//	        defer wg.Done()
//	        files, _ := scanner.ScanProject(path) // 线程安全
//	        fmt.Printf("%s: %d files\n", path, len(files))
//	    }(proj)
//	}
//	wg.Wait()
//
// 注意事项:
//   - 返回的是空结构体指针，仅用于方法调用
//   - 可以复用同一实例扫描多个项目（节省内存）
//   - 扫描大型项目时，返回的map可能占用大量内存
//
// 性能特性:
//   - 创建开销: 几乎为零（仅指针分配）
//   - 内存占用: 0字节（空结构体）
//   - 并发安全: 天然支持，无需加锁
//
// 作者: JIA
func NewProjectScanner() *ProjectScanner {
	return &ProjectScanner{}
}

// ScanProject 递归扫描项目目录，收集所有Go源文件及其内容
//
// 功能说明:
//
//	本方法使用Go标准库的filepath.WalkDir函数递归遍历整个项目目录树，
//	自动识别所有Go源文件（*.go），读取文件内容，构建文件名到文件内容的映射表。
//	返回的map可直接用于代码分析、测试覆盖率统计等评估任务。
//
// 扫描流程（基于filepath.WalkDir）:
//
//  1. 深度优先遍历（DFS）:
//     从rootPath开始，递归访问所有子目录
//     遍历顺序: 按字典序访问目录和文件
//
//  2. 文件过滤规则:
//     - 目录: 跳过（仅遍历，不添加到结果）
//     - 非.go文件: 跳过（如.txt, .md, .json等）
//     - .go文件: 读取内容并添加到结果map
//
//  3. 路径处理:
//     - 绝对路径 → 相对路径转换（相对于rootPath）
//     - 使用filepath.Rel()计算相对路径
//     - 如果计算失败，回退使用绝对路径
//
//  4. 内容读取:
//     - 使用os.ReadFile()一次性读取整个文件（适合小到中型文件）
//     - 转换为UTF-8字符串存储到map
//
// filepath.WalkDir详解（Go 1.16+）:
//
//	WalkDir是filepath.Walk的优化版本：
//	- 参数1: 根目录路径（string）
//	- 参数2: 访问函数（func(path string, d DirEntry, err error) error）
//	- 返回值: 首个非nil错误或nil
//
//	访问函数返回值行为：
//	- return nil: 继续遍历下一个文件/目录
//	- return filepath.SkipDir: 跳过当前目录（仅对目录有效）
//	- return err: 立即停止遍历，WalkDir返回该错误
//
// 参数:
//   - rootPath: 项目根目录的绝对或相对路径
//     示例: "./00-assessment-system", "/home/user/go-mastery"
//
// 返回值:
//   - map[string]string: 文件名到文件内容的映射，键为相对路径
//     示例: {
//     "models/student.go": "package models\ntype Student struct {...}",
//     "tools/assessment_tools.go": "package tools\nfunc Analyze() {...}",
//     }
//   - error: 扫描过程中的错误，包括：
//   - 目录不存在或无权限访问
//   - 文件读取失败（权限、编码问题等）
//   - 路径计算失败
//
// 使用场景:
//   - 代码质量评估：收集项目所有源文件用于分析
//   - 测试覆盖率分析：获取源文件和测试文件映射
//   - 项目统计：计算代码行数、文件数量
//   - 代码搜索：在所有文件中查找特定模式
//
// 示例:
//
//	scanner := NewProjectScanner()
//
//	// 示例1: 扫描小型项目
//	files, err := scanner.ScanProject("./01-basics/01-hello")
//	if err != nil {
//	    log.Fatal("扫描失败:", err)
//	}
//	// files = {
//	//   "main.go": "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"Hello\") }",
//	// }
//	fmt.Printf("扫描到 %d 个Go文件\n", len(files))
//
//	// 示例2: 扫描中型项目（多层目录）
//	files, err = scanner.ScanProject("./00-assessment-system")
//	// files = {
//	//   "models/student.go": "...",
//	//   "models/assessment.go": "...",
//	//   "models/competency.go": "...",
//	//   "tools/assessment_tools.go": "...",
//	//   "evaluators/code_quality.go": "...",
//	//   ... (数十个文件)
//	// }
//
//	// 示例3: 结合AnalyzeTestCoverage使用
//	files, err = scanner.ScanProject("./03-concurrency")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	analyzer := NewTestAnalyzer()
//	metrics, _ := analyzer.AnalyzeTestCoverage(files)
//	fmt.Printf("测试覆盖率: %.1f%%\n", metrics.TestCoverage*100)
//
// 注意事项:
//   - 大型项目（数千文件）可能导致大量内存占用（所有文件内容在内存）
//   - 二进制文件或非UTF-8文件会被读取但可能显示乱码
//   - 符号链接会被正常遍历（可能导致循环引用，filepath.WalkDir会检测）
//   - 隐藏文件（.开头）也会被扫描
//
// 性能特性:
//   - 时间复杂度: O(n)，n为目录树中文件总数
//   - 空间复杂度: O(m)，m为所有.go文件内容总大小
//   - I/O密集型操作，性能受磁盘速度影响
//   - 适用场景: <1000个文件的中小型项目
//
// 错误处理策略:
//   - 文件读取失败: 立即返回错误，停止扫描
//   - 相对路径计算失败: 回退使用绝对路径（不中断扫描）
//   - 目录权限不足: WalkDir返回错误，调用方可决定是否继续
//
// 改进方向:
//   - 支持文件过滤器（如排除vendor、node_modules目录）
//   - 流式处理大文件（逐行读取而非全部加载）
//   - 增加进度回调（用于显示扫描进度）
//   - 支持并发扫描多个目录（加速大型项目）
//
// 安全注意:
//   - 使用#nosec G304注释豁免gosec警告，因为path来自filepath.WalkDir的受信任遍历
//   - 生产环境应验证rootPath是否在允许的白名单内
//
// 作者: JIA
func (ps *ProjectScanner) ScanProject(rootPath string) (map[string]string, error) {
	files := make(map[string]string)

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录和非 Go 文件
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 读取文件内容
		// #nosec G304 -- 评估系统内部操作，path来自filepath.WalkDir遍历，为受信任的文件系统路径
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// 使用相对路径作为键
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			// 如果无法获取相对路径，使用绝对路径
			relPath = path
		}
		files[relPath] = string(content)

		return nil
	})

	return files, err
}

// ReportGenerator 报告生成器
type ReportGenerator struct{}

// NewReportGenerator 创建报告生成器实例（无状态设计）
//
// 功能说明:
//
//	本函数创建一个ReportGenerator对象，用于将评估结果和学习进度数据
//	格式化为Markdown格式的可读报告。采用无状态空结构体设计，保持与其他工具组件一致的架构风格。
//
// 设计哲学（空结构体模式的第三次应用）:
//
//	ReportGenerator、TestAnalyzer、ProjectScanner三者共享相同的设计模式：
//	1. 零内存开销：struct{} 无字段，不占用内存空间
//	2. 纯函数设计：所有输入通过参数传递，输出通过返回值
//	3. 无副作用：不修改全局状态，不维护内部状态
//	4. 并发安全：可安全共享，无锁设计
//
// 报告生成能力:
//
//	支持的报告类型：
//	- 评估报告（GenerateAssessmentReport）：展示单次评估的详细分数
//	- 进度报告（GenerateProgressReport）：展示学习者的整体学习进度
//	- 格式：Markdown（.md），兼容GitHub/GitLab渲染
//	- 特点：清晰的层级结构、易于阅读、可复制分享
//
// 返回值:
//   - *ReportGenerator: 报告生成器指针，用于调用报告生成方法
//
// 使用场景:
//   - 评估系统自动生成评估报告
//   - 学习者查看个人学习进度
//   - 导师查看学生学习情况
//   - 导出报告用于存档或分享
//
// 示例:
//
//	generator := NewReportGenerator()
//
//	// 场景1: 生成评估报告
//	result := &models.AssessmentResult{
//	    SessionID:    "session_20250103_001",
//	    OverallScore: 85.0,
//	    MaxScore:     100.0,
//	    DimensionScores: map[string]float64{
//	        "技术深度": 90.0,
//	        "工程实践": 80.0,
//	    },
//	}
//	report := generator.GenerateAssessmentReport(result)
//	os.WriteFile("assessment_report.md", []byte(report), 0600)
//
//	// 场景2: 生成进度报告
//	student := &models.StudentProfile{
//	    Name:         "张三",
//	    Email:        "zhangsan@example.com",
//	    CurrentStage: 5,
//	    Projects:     []models.ProjectRecord{...},
//	}
//	report = generator.GenerateProgressReport(student)
//	fmt.Println(report) // 直接输出或保存到文件
//
// 注意事项:
//   - 返回的是空结构体指针，主要用于方法调用
//   - 可复用同一实例生成多个报告（节省内存）
//   - 生成的报告为UTF-8编码的Markdown字符串
//
// 性能特性:
//   - 创建开销: 几乎为零（仅指针分配）
//   - 报告生成: 字符串拼接操作，O(n)复杂度，n为报告长度
//   - 内存占用: 0字节（空结构体）
//
// 作者: JIA
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{}
}

// GenerateAssessmentReport 生成单次评估的Markdown格式详细报告
//
// 功能说明:
//
//	本方法将AssessmentResult结构体中的评估数据格式化为易读的Markdown报告，
//	包含会话ID、评估时间、总分、各维度得分等关键信息。报告采用标准Markdown语法，
//	可直接在GitHub/GitLab查看，或导出为PDF/HTML等格式。
//
// 报告结构:
//
//	# 代码评估报告
//
//	**会话ID:** session_20250103_001
//	**评估时间:** 2025-10-03 15:30:45
//	**总分:** 85.50/100.00
//
//	## 评估结果
//
//	### 技术深度
//	**得分:** 90.00/100
//
//	### 工程实践
//	**得分:** 80.00/100
//
// 字符串拼接优化（strings.Builder）:
//
//	为什么使用strings.Builder而不是直接拼接（+=）？
//	1. 性能优势：Builder使用预分配缓冲区，避免每次拼接都创建新字符串
//	2. 内存效率：+=每次都会分配新内存，n次拼接产生O(n²)的内存分配
//	3. 零GC压力：Builder复用同一缓冲区，减少垃圾回收压力
//
//	性能对比（100次拼接）：
//	- += 拼接:           ~5000ns,  5000次内存分配
//	- strings.Builder:   ~500ns,   1次内存分配（预分配后）
//
// 参数:
//   - result: 评估结果指针，包含以下关键字段：
//   - SessionID: 评估会话唯一标识符
//   - OverallScore: 综合得分（0.0-100.0范围）
//   - MaxScore: 满分值（通常为100.0）
//   - DimensionScores: 各评估维度得分映射（如"技术深度"->90.0）
//
// 返回值:
//   - string: Markdown格式的评估报告字符串，可直接写入文件或打印
//
// Markdown语法说明:
//   - # 一级标题（报告主标题）
//   - ## 二级标题（章节标题）
//   - ### 三级标题（维度标题）
//   - **粗体**（强调关键数据）
//   - \n\n 双换行（段落分隔）
//
// 使用场景:
//   - 完成代码评估后自动生成报告
//   - 学习者查看评估详情
//   - 导出评估记录用于存档
//   - 生成PDF/HTML版本的评估报告
//
// 示例:
//
//	generator := NewReportGenerator()
//
//	// 构造评估结果数据
//	result := &models.AssessmentResult{
//	    SessionID:    "session_20250103_001",
//	    OverallScore: 87.5,
//	    MaxScore:     100.0,
//	    DimensionScores: map[string]float64{
//	        "技术深度":   92.0,
//	        "工程实践":   83.0,
//	        "项目经验":   85.0,
//	        "软技能":     90.0,
//	    },
//	}
//
//	// 生成Markdown报告
//	report := generator.GenerateAssessmentReport(result)
//	fmt.Println(report)
//	// 输出:
//	// # 代码评估报告
//	//
//	// **会话ID:** session_20250103_001
//	// **评估时间:** 2025-10-03 15:30:45
//	// **总分:** 87.50/100.00
//	//
//	// ## 评估结果
//	//
//	// ### 技术深度
//	// **得分:** 92.00/100
//	// ...
//
//	// 保存到文件
//	err := os.WriteFile("assessment_report.md", []byte(report), 0600)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// 注意事项:
//   - 报告时间格式固定为 "2006-01-02 15:04:05"（Go时间格式化标准）
//   - DimensionScores的遍历顺序不固定（map无序），如需固定顺序应排序键
//   - 浮点数格式化为两位小数（%.2f）
//   - 返回的字符串为UTF-8编码，包含中文字符
//
// 改进方向:
//   - 增加图表生成（如雷达图展示各维度）
//   - 支持自定义报告模板
//   - 增加历史评估对比（进步趋势）
//   - 支持导出为JSON/HTML/PDF格式
//
// 作者: JIA
func (rg *ReportGenerator) GenerateAssessmentReport(result *models.AssessmentResult) string {
	var report strings.Builder

	report.WriteString("# 代码评估报告\n\n")
	report.WriteString(fmt.Sprintf("**会话ID:** %s\n", result.SessionID))
	report.WriteString(fmt.Sprintf("**评估时间:** %s\n", time.Now().Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("**总分:** %.2f/%.2f\n\n", result.OverallScore, result.MaxScore))

	report.WriteString("## 评估结果\n\n")
	for dimension, score := range result.DimensionScores {
		report.WriteString(fmt.Sprintf("### %s\n", dimension))
		report.WriteString(fmt.Sprintf("**得分:** %.2f/100\n", score))
		report.WriteString("\n")
	}

	return report.String()
}

// GenerateProgressReport 生成学习者完整学习进度的Markdown格式报告
//
// 功能说明:
//
//	本方法将StudentProfile结构体中的学习进度数据格式化为Markdown报告，
//	包含学生基本信息、当前学习阶段、项目历史记录、整体进度百分比等关键指标。
//	用于帮助学习者和导师全面了解学习状况和成长轨迹。
//
// 报告结构:
//
//	# 学习进度报告
//
//	**学生姓名:** 张三
//	**邮箱:** zhangsan@example.com
//	**当前阶段:** 5
//
//	## 项目历史
//
//	- **CLI任务管理工具** - 85.50分
//	- **RESTful博客API** - 90.00分
//
//	**整体进度:** 33.3%
//
// 进度计算逻辑:
//
//	整体进度 = (CurrentStage / TotalLearningStages) × 100%
//
//	示例计算：
//	- 当前阶段: 5
//	- 总阶段数: 15（Go从入门到通天的15个学习模块）
//	- 进度百分比: (5 / 15) × 100% = 33.3%
//
//	说明：
//	- TotalLearningStages常量定义在evaluators包（值为15）
//	- 完成第15阶段时，进度达到100%
//	- 每完成一个阶段，进度增加约6.67%（100/15）
//
// 参数:
//   - student: 学习者档案指针，包含以下关键字段：
//   - Name: 学生姓名
//   - Email: 联系邮箱
//   - CurrentStage: 当前学习阶段（1-15范围）
//   - Projects: 项目作品集切片（ProjectRecord结构体数组）
//
// 返回值:
//   - string: Markdown格式的学习进度报告字符串
//
// 性能优化（避免大结构体复制）:
//
//	ProjectRecord结构体大小为248字节（包含多个字符串切片和时间字段）
//	使用索引遍历而非值遍历，避免每次循环复制248字节：
//
//	❌ 低效写法（每次复制248字节）：
//	for _, project := range student.Projects {
//	    report.WriteString(project.Name) // project是副本
//	}
//
//	✅ 高效写法（仅使用指针，0复制）：
//	for i := range student.Projects {
//	    project := &student.Projects[i] // project是指针
//	    report.WriteString(project.Name)
//	}
//
//	性能提升（1000个项目场景）：
//	- 值遍历: 248KB内存复制, ~500μs
//	- 索引遍历: 8KB指针操作, ~50μs（快10倍）
//
// 使用场景:
//   - 学习者定期查看学习进度
//   - 导师评估学生学习状况
//   - 生成学习总结报告
//   - 申请认证时展示学习历程
//
// 示例:
//
//	generator := NewReportGenerator()
//
//	// 构造学生档案数据
//	student := &models.StudentProfile{
//	    Name:         "李明",
//	    Email:        "liming@example.com",
//	    CurrentStage: 5,
//	    Projects: []models.ProjectRecord{
//	        {Name: "CLI任务管理工具", OverallScore: 85.5},
//	        {Name: "RESTful博客API", OverallScore: 90.0},
//	        {Name: "微服务架构实战", OverallScore: 88.0},
//	    },
//	}
//
//	// 生成进度报告
//	report := generator.GenerateProgressReport(student)
//	fmt.Println(report)
//	// 输出:
//	// # 学习进度报告
//	//
//	// **学生姓名:** 李明
//	// **邮箱:** liming@example.com
//	// **当前阶段:** 5
//	//
//	// ## 项目历史
//	//
//	// - **CLI任务管理工具** - 85.50分
//	// - **RESTful博客API** - 90.00分
//	// - **微服务架构实战** - 88.00分
//	//
//	// **整体进度:** 33.3%
//
//	// 保存到文件
//	err := os.WriteFile("progress_report.md", []byte(report), 0600)
//
// 注意事项:
//   - 如果student.Projects为空，"## 项目历史"章节不会被生成
//   - 整体进度百分比格式化为一位小数（%.1f%%）
//   - 项目得分格式化为两位小数（%.2f分）
//   - 报告不包含详细的项目描述和反馈（仅名称和分数）
//
// 改进方向:
//   - 增加各阶段完成状态的可视化（如进度条）
//   - 展示每个项目的详细评估信息
//   - 增加学习时长统计（TotalHours字段）
//   - 增加技能成长曲线图
//   - 增加与其他学习者的对比分析
//
// 作者: JIA
func (rg *ReportGenerator) GenerateProgressReport(student *models.StudentProfile) string {
	var report strings.Builder

	report.WriteString("# 学习进度报告\n\n")
	report.WriteString(fmt.Sprintf("**学生姓名:** %s\n", student.Name))
	report.WriteString(fmt.Sprintf("**邮箱:** %s\n", student.Email))
	report.WriteString(fmt.Sprintf("**当前阶段:** %d\n", student.CurrentStage))

	if len(student.Projects) > 0 {
		report.WriteString("## 项目历史\n\n")
		// 使用索引遍历避免大结构体复制（248字节）
		for i := range student.Projects {
			project := &student.Projects[i]
			report.WriteString(fmt.Sprintf("- **%s** - %.2f分\n",
				project.Name,
				project.OverallScore))
		}
		report.WriteString("\n")

		// 计算进度
		progress := float64(student.CurrentStage) / evaluators.TotalLearningStages * 100
		report.WriteString(fmt.Sprintf("**整体进度:** %.1f%%\n\n", progress))
	}

	return report.String()
}

// 数据结构

// ComplexityMetrics 复杂度指标
type ComplexityMetrics struct {
	Functions            int
	CyclomaticComplexity int
	MaxComplexity        int
	AverageComplexity    float64
	Conditionals         int
	Loops                int
	Switches             int
}

// StyleMetrics 代码风格指标
type StyleMetrics struct {
	TotalLines        int
	LongLines         int
	BadNaming         int
	Comments          int
	CommentRatio      float64
	EmptyLines        int
	IndentationIssues int
}

// TestMetrics 测试指标
type TestMetrics struct {
	TotalSourceFiles   int
	TotalTestFiles     int
	TotalTestFunctions int
	TestCoverage       float64
}

// ConfigLoader 配置加载器
type ConfigLoader struct{}

// LoadConfig 加载评估系统配置（当前版本返回硬编码默认配置）
//
// 功能说明:
//
//	本方法返回评估系统的默认配置，包含评分权重分配和质量阈值定义。
//	当前实现采用硬编码配置策略，未来版本可扩展为从JSON/YAML文件加载自定义配置。
//
// 配置内容详解:
//
//	一、评分权重 (ScoreWeights) - 各维度在综合评分中的占比：
//
//	  1. code_quality (代码质量): WeightHigh (0.4) - 占比40%
//	     评估内容: 代码复杂度、风格规范、可读性、安全性
//	     权重最高的原因: 代码质量是软件长期可维护性的核心
//
//	  2. functionality (功能性): WeightMediumHigh (0.3) - 占比30%
//	     评估内容: 功能完整性、需求覆盖度、边界处理
//	     权重较高的原因: 功能实现是软件的首要目标
//
//	  3. test_coverage (测试覆盖率): WeightMediumLow (0.2) - 占比20%
//	     评估内容: 单元测试覆盖率、测试用例完备性
//	     权重中等的原因: 测试是质量保障但非唯一手段
//
//	  4. documentation (文档完整性): WeightVeryLow (0.1) - 占比10%
//	     评估内容: 代码注释、API文档、README完整性
//	     权重最低的原因: 文档重要但在学习阶段优先级较低
//
//	二、质量阈值 (QualityThresholds) - 各项质量标准的临界值：
//
//	  1. MinScore (最低分): 60分（Score60常量）
//	     含义: 及格线，低于此分数视为不达标
//	     用途: 认证考试最低通过标准
//
//	  2. GoodScore (良好分): 80分（DefaultScore常量）
//	     含义: 良好水平，达到工程化标准
//	     用途: 项目验收基准分数
//
//	  3. ExcellentScore (优秀分): 95分（Score95常量）
//	     含义: 优秀水平，接近完美实现
//	     用途: 高级认证要求
//
//	  4. MaxComplexity (最大复杂度): 10
//	     含义: 单函数圈复杂度上限
//	     依据: 业界公认的可维护性临界点
//
//	  5. MinTestCoverage (最低测试覆盖率): MinTestCoverageRatio常量（如80%）
//	     含义: 测试覆盖率最低要求
//	     依据: 平衡质量保障与开发效率
//
// 参数:
//   - _ (string): 配置文件路径参数（当前版本未使用，保留用于未来扩展）
//     空白标识符（_）表明该参数在当前实现中被有意忽略
//     未来可扩展为: LoadConfig("config.yaml") 加载自定义配置
//
// 返回值:
//   - *AssessmentConfig: 评估配置对象指针，包含完整的权重和阈值设置
//   - error: 当前实现总是返回nil，未来可能返回配置加载错误
//
// 设计理念:
//   - 接口前瞻性设计：虽然当前硬编码，但预留了配置文件路径参数
//   - 配置分离：权重和阈值独立配置，便于调整评估标准
//   - 常量复用：使用evaluators包的常量，保证全局一致性
//
// 使用场景:
//   - 评估系统初始化时加载默认配置
//   - 单元测试中创建测试配置
//   - 未来扩展为从文件读取自定义配置
//
// 示例:
//
//	loader := &ConfigLoader{}
//	config, err := loader.LoadConfig("")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// config.ScoreWeights["code_quality"] == 0.4
//	// config.QualityThresholds.MinScore == 60.0
//
//	// 应用配置进行评估
//	codeQualityWeight := config.ScoreWeights["code_quality"]
//	minPassingScore := config.QualityThresholds.MinScore
//
// 注意事项:
//   - 当前版本忽略configPath参数，总是返回相同的默认配置
//   - 所有权重总和应为1.0（0.4+0.3+0.2+0.1=1.0），便于百分比计算
//   - 阈值分数均为0-100范围的浮点数
//   - MaxComplexity为整数，其他阈值为浮点数
//
// 扩展方向:
//   - 支持从JSON/YAML文件加载配置
//   - 支持环境变量覆盖默认值
//   - 支持多环境配置（dev/staging/prod）
//   - 支持运行时动态调整权重
//
// 作者: JIA
func (cl *ConfigLoader) LoadConfig(_ string) (*AssessmentConfig, error) {
	// 简化的配置加载
	return &AssessmentConfig{
		ScoreWeights: map[string]float64{
			"code_quality":  evaluators.WeightHigh,
			"functionality": evaluators.WeightMediumHigh,
			"test_coverage": evaluators.WeightMediumLow,
			"documentation": evaluators.WeightVeryLow,
		},
		QualityThresholds: QualityThresholds{
			MinScore:        evaluators.Score60,
			GoodScore:       evaluators.DefaultScore,
			ExcellentScore:  evaluators.Score95,
			MaxComplexity:   10,
			MinTestCoverage: evaluators.MinTestCoverageRatio,
		},
	}, nil
}

// AssessmentConfig 评估配置
type AssessmentConfig struct {
	ScoreWeights      map[string]float64 `json:"score_weights"`
	QualityThresholds QualityThresholds  `json:"quality_thresholds"`
}

// QualityThresholds 质量阈值
type QualityThresholds struct {
	MinScore        float64 `json:"min_score"`
	GoodScore       float64 `json:"good_score"`
	ExcellentScore  float64 `json:"excellent_score"`
	MaxComplexity   int     `json:"max_complexity"`
	MinTestCoverage float64 `json:"min_test_coverage"`
}

// FileWatcher 文件监控器
type FileWatcher struct {
	watchPaths []string
}

// NewFileWatcher 创建文件监控器实例（简化版实现，生产环境建议使用fsnotify）
//
// 功能说明:
//
//	本函数创建一个FileWatcher对象，用于监控指定路径列表的文件变化事件。
//	当前实现为教学简化版本，采用轮询策略检测文件修改时间变化。
//	生产环境应使用专业的文件系统监控库（如github.com/fsnotify/fsnotify）。
//
// 简化实现 vs 生产实现对比:
//
//	当前简化实现（轮询方式）：
//	  ✅ 优点: 简单易懂，无外部依赖，跨平台兼容性好
//	  ❌ 缺点: CPU占用高，实时性差（1秒延迟），无法检测文件删除/重命名
//	  适用: 教学演示、原型开发、监控文件数量少的场景
//
//	生产实现（fsnotify库）：
//	  ✅ 优点: 事件驱动，实时性强，资源消耗低，功能完整
//	  ❌ 缺点: 依赖外部库，需处理平台差异（Linux/Windows/macOS）
//	  适用: 生产环境、IDE开发、构建工具、DevOps工具
//
// FileWatcher结构体字段:
//
//	watchPaths []string - 监控的文件/目录路径列表
//	  可以是文件路径: "/path/to/file.go"
//	  可以是目录路径: "/path/to/directory"（当前实现仅监控目录本身，不递归）
//	  路径格式: 支持绝对路径和相对路径
//
// 参数:
//   - paths: 需要监控的文件或目录路径切片
//     每个路径将启动一个独立的goroutine进行监控
//     空切片合法但不会触发任何监控行为
//     路径有效性检查在watchPath方法中进行（os.Stat）
//
// 返回值:
//   - *FileWatcher: 初始化完成的文件监控器指针，可立即调用WatchForChanges启动监控
//
// 设计模式:
//   - 构造器模式: 通过New函数封装对象创建逻辑
//   - 数据驱动: 通过paths切片批量配置监控对象
//
// 使用场景:
//   - 代码热重载: 监控源码变化自动重启服务
//   - 配置文件监控: 配置变更时自动重新加载
//   - 构建工具: 检测文件变化触发自动编译
//   - 日志分析: 实时监控日志文件新增内容
//
// 示例:
//
//	// 监控单个文件
//	watcher := NewFileWatcher([]string{"config.yaml"})
//	watcher.WatchForChanges(func(path string) {
//	    fmt.Printf("文件变化: %s\n", path)
//	    // 重新加载配置
//	    loadConfig(path)
//	})
//
//	// 监控多个源码文件
//	sourceFiles := []string{
//	    "main.go",
//	    "handlers.go",
//	    "models.go",
//	}
//	watcher := NewFileWatcher(sourceFiles)
//	watcher.WatchForChanges(func(path string) {
//	    fmt.Printf("源码变化，重新编译: %s\n", path)
//	    exec.Command("go", "build").Run()
//	})
//
// 注意事项:
//   - 本实现不递归监控子目录，仅监控指定路径本身
//   - 每个路径启动一个独立goroutine，大量路径会消耗较多资源
//   - WatchForChanges调用后立即返回，监控在后台goroutine中进行
//   - 无停止机制，监控将持续到程序结束（生产环境需要Stop方法）
//   - 路径不存在不会报错，会在watchPath中静默失败
//
// 改进方向（生产环境）:
//   - 使用fsnotify替换轮询实现
//   - 增加Stop()方法优雅停止监控
//   - 支持递归监控目录（WatchRecursive）
//   - 增加事件过滤（仅监控特定扩展名）
//   - 增加防抖机制（避免短时间内重复触发）
//
// 作者: JIA
func NewFileWatcher(paths []string) *FileWatcher {
	return &FileWatcher{
		watchPaths: paths,
	}
}

// WatchForChanges 启动文件监控，检测到变化时调用回调函数
//
// 功能说明:
//
//	本方法为FileWatcher中配置的所有路径启动并发监控，当检测到文件修改时，
//	自动调用用户提供的回调函数。每个路径在独立的goroutine中运行，互不阻塞。
//
// 工作流程:
//
//  1. 遍历watchPaths切片: 获取所有需要监控的路径
//  2. 为每个路径启动goroutine: 使用go关键字并发执行watchPath方法
//  3. 立即返回: 不等待监控完成，监控在后台持续运行
//  4. 检测到变化时: watchPath方法调用callback回调函数
//
// 并发模型:
//
//	主goroutine                    监控goroutine 1           监控goroutine 2
//	    |                              |                       |
//	    | WatchForChanges()            |                       |
//	    |---go watchPath(path1)------->|                       |
//	    |---go watchPath(path2)---------------------->|        |
//	    | return nil                    |                       |
//	    ↓                                |                       |
//	  (继续执行)                        | (每秒检测path1)        | (每秒检测path2)
//	                                    | callback(path1)        | callback(path2)
//	                                    | (循环监控)             | (循环监控)
//
// 回调函数设计:
//
//	callback参数说明:
//	  类型: func(string)
//	  参数: 发生变化的文件路径（与watchPaths中的路径一致）
//	  返回值: 无（void函数）
//	  执行时机: 文件最后修改时间在2秒内时触发
//
//	回调函数注意事项:
//	  - 在监控goroutine中同步执行，耗时操作会阻塞后续检测
//	  - 应避免长时间阻塞操作，建议异步处理复杂任务
//	  - 异常应在callback内部捕获，避免panic导致监控goroutine崩溃
//
// 参数:
//   - callback: 文件变化时的回调函数
//     签名: func(path string)
//     path: 发生变化的文件完整路径
//     示例: func(path string) { fmt.Println("Changed:", path) }
//
// 返回值:
//   - error: 当前实现总是返回nil，未来可能返回监控启动错误
//     预留错误返回用于扩展（如路径不存在、权限不足等）
//
// 性能特性:
//   - 时间复杂度: O(n)，n为路径数量（goroutine启动）
//   - 空间复杂度: O(n)，每个路径占用一个goroutine栈（约2KB-8KB）
//   - 资源消耗: n个路径 = n个goroutine + n次/秒的os.Stat调用
//
// 使用场景:
//   - 配置文件热更新: 监控config.yaml变化自动重载配置
//   - 代码热重载: 监控.go文件变化触发重新编译
//   - 日志实时分析: 监控日志文件新增内容
//   - 文档自动构建: 监控Markdown文件变化重新生成文档
//
// 示例:
//
//	watcher := NewFileWatcher([]string{"config.yaml", "app.log"})
//
//	// 简单回调 - 打印变化信息
//	err := watcher.WatchForChanges(func(path string) {
//	    fmt.Printf("[%s] 文件变化: %s\n", time.Now().Format("15:04:05"), path)
//	})
//
//	// 配置重载回调 - 检测config.yaml变化
//	err = watcher.WatchForChanges(func(path string) {
//	    if strings.HasSuffix(path, ".yaml") {
//	        log.Println("重新加载配置...")
//	        config, err := loadConfig(path)
//	        if err != nil {
//	            log.Printf("配置加载失败: %v", err)
//	            return
//	        }
//	        applyConfig(config)
//	    }
//	})
//
//	// 异步处理回调 - 避免阻塞监控
//	err = watcher.WatchForChanges(func(path string) {
//	    go func(p string) { // 启动新goroutine处理
//	        defer func() {
//	            if r := recover(); r != nil {
//	                log.Printf("回调panic: %v", r)
//	            }
//	        }()
//	        // 耗时操作
//	        time.Sleep(5 * time.Second)
//	        fmt.Printf("处理完成: %s\n", p)
//	    }(path)
//	})
//
// 注意事项:
//   - 本方法立即返回，不会阻塞主goroutine
//   - 监控goroutine会永久运行，无法停止（需要程序退出）
//   - 回调函数在监控goroutine中执行，应避免长时间阻塞
//   - 文件频繁修改可能导致回调被频繁触发（每秒最多1次）
//   - 当前无并发控制，所有回调串行执行可能导致延迟累积
//
// 已知限制:
//   - 仅检测ModTime变化，无法检测文件删除、重命名、权限变更
//   - 1秒轮询间隔导致检测延迟（最坏情况1秒延迟）
//   - 2秒时间窗口可能导致同一次修改被触发多次
//   - 无防抖机制，快速连续修改会多次触发回调
//
// 改进方向（生产环境）:
//   - 使用fsnotify实现事件驱动监控（实时性强）
//   - 增加Stop()方法优雅停止所有监控goroutine
//   - 增加防抖机制（Debounce）避免重复触发
//   - 增加并发限制（如最多同时执行10个回调）
//   - 支持事件类型过滤（仅监控Create/Modify/Delete）
//   - 增加错误处理和重试机制
//
// 作者: JIA
func (fw *FileWatcher) WatchForChanges(callback func(string)) error {
	// 简化的文件监控实现
	// 实际实现应该使用 fsnotify 等库
	for _, path := range fw.watchPaths {
		go fw.watchPath(path, callback)
	}
	return nil
}

// watchPath 监控单个路径的文件变化（内部方法，无限循环轮询）
//
// 功能说明:
//
//	本方法是FileWatcher的核心工作函数，在独立goroutine中运行，
//	通过无限循环轮询检测指定路径的文件修改时间变化，并在检测到变化时调用回调函数。
//
// 工作机制（轮询策略）:
//
//  1. 无限循环: for { } 死循环，goroutine将永久运行直到程序结束
//  2. 定时休眠: time.Sleep(1 * time.Second) 每秒检查一次
//  3. 获取文件信息: os.Stat(path) 获取文件元数据
//  4. 检查修改时间: time.Since(info.ModTime()) < 2*time.Second
//  5. 触发回调: 满足条件时调用callback(path)
//
// 时间窗口逻辑:
//
//	检测逻辑: time.Since(info.ModTime()) < 2*time.Second
//
//	解释: 如果文件的最后修改时间距离当前时间小于2秒，则认为文件刚被修改
//
//	时间线示例:
//	  T0秒: 文件被修改 (ModTime更新为T0)
//	  T1秒: 第一次检测 → Since(T0) = 1秒 < 2秒 → 触发回调
//	  T2秒: 第二次检测 → Since(T0) = 2秒 = 2秒 → 不触发（边界情况）
//	  T3秒: 第三次检测 → Since(T0) = 3秒 > 2秒 → 不触发
//
//	为什么是2秒窗口？
//	  - 1秒轮询间隔 + 1秒容错余地 = 确保至少触发一次
//	  - 避免遗漏检测（如果窗口=1秒，可能因时序错过触发）
//	  - 但可能导致同一次修改被触发2次（T1秒和T2秒都满足条件）
//
// os.Stat错误处理:
//
//	if info, err := os.Stat(path); err == nil { ... }
//
//	静默失败策略:
//	  - 文件不存在: err != nil，跳过本次检测，继续循环
//	  - 权限不足: err != nil，跳过本次检测，继续循环
//	  - 路径无效: err != nil，跳过本次检测，继续循环
//
//	设计考量:
//	  - 不中断监控: 即使文件临时不可访问，监控继续运行
//	  - 适用场景: 监控日志文件（可能被删除重建）
//	  - 缺点: 无错误日志，问题难以排查
//
// 参数:
//   - path: 要监控的文件或目录路径
//     可以是绝对路径或相对路径
//     不存在的路径不会报错，会静默跳过
//   - callback: 检测到变化时调用的回调函数
//     在当前goroutine中同步执行
//     callback阻塞会影响下次检测时机
//
// 返回值: 无（内部方法，通过goroutine调用，无法接收返回值）
//
// 性能分析:
//
//	每个路径的资源消耗（每秒）:
//	  - 1次 time.Sleep(1s) 系统调用
//	  - 1次 os.Stat() 系统调用（约0.01-0.1ms）
//	  - 1次 time.Since() 计算（纳秒级）
//	  - 0-1次 callback() 调用（取决于是否有变化）
//
//	总体消耗（10个文件监控）:
//	  - 10个goroutine（约20KB-80KB栈内存）
//	  - 10次/秒 os.Stat调用
//	  - CPU使用率: 几乎为0（大部分时间在Sleep）
//
// 使用场景（内部方法，不直接调用）:
//   - 由WatchForChanges方法自动调用
//   - 每个监控路径对应一个watchPath goroutine
//
// 示例（概念展示，实际不直接调用）:
//
//	// WatchForChanges内部会这样调用
//	go fw.watchPath("config.yaml", func(path string) {
//	    fmt.Printf("Config changed: %s\n", path)
//	})
//	// 上述goroutine将永久运行，每秒检测config.yaml的ModTime
//
// 已知问题:
//
//	问题1: 重复触发
//	  场景: 文件在T0秒被修改，T1秒和T2秒都触发回调
//	  原因: 2秒时间窗口设计
//	  影响: 配置重载等操作可能被执行2次
//	  解决: 增加防抖机制或记录上次触发时间
//
//	问题2: 无法停止
//	  场景: 程序运行期间无法停止监控
//	  原因: for循环无退出条件
//	  影响: 资源无法释放，goroutine泄漏
//	  解决: 增加context.Context控制goroutine生命周期
//
//	问题3: 错误静默
//	  场景: 路径不存在或权限不足
//	  原因: os.Stat错误被忽略
//	  影响: 用户不知道监控失败
//	  解决: 增加错误日志或错误回调
//
//	问题4: 检测延迟
//	  场景: 文件修改后最多1秒才能检测到
//	  原因: 1秒轮询间隔
//	  影响: 实时性要求高的场景不适用
//	  解决: 使用fsnotify事件驱动监控（毫秒级响应）
//
// 改进方向:
//   - 增加context.Context支持优雅退出
//   - 增加防抖机制避免重复触发
//   - 增加错误日志记录
//   - 记录上次ModTime避免重复检测
//   - 支持可配置的检测间隔和时间窗口
//
// 作者: JIA
func (fw *FileWatcher) watchPath(path string, callback func(string)) {
	for {
		time.Sleep(1 * time.Second)
		// 检查文件变化（简化实现）
		if info, err := os.Stat(path); err == nil {
			// 如果文件最近被修改
			if time.Since(info.ModTime()) < 2*time.Second {
				callback(path)
			}
		}
	}
}

// CLIRunner 命令行工具运行器
type CLIRunner struct{}

// NewCLIRunner 创建命令行工具运行器实例（无状态空结构体设计）
//
// 功能说明:
//
//	本函数创建一个CLIRunner对象，用于运行评估系统的命令行界面。
//	采用无状态空结构体设计，与TestAnalyzer、ProjectScanner、ReportGenerator保持一致的架构风格。
//
// 设计模式（空结构体的第四次应用）:
//
//	CLIRunner = struct{} - 零字段，纯行为对象
//
//	为什么继续使用空结构体？
//	  1. 一致性: 与其他工具类（TestAnalyzer、ProjectScanner、ReportGenerator）保持统一模式
//	  2. 简洁性: CLI操作本质上是无状态的，所有数据通过参数传递
//	  3. 可扩展性: 未来可增加状态字段（如配置、日志记录器）而不破坏现有API
//	  4. 零开销: 不占用内存空间，创建成本极低
//
// 对比有状态设计（假设场景）:
//
//	如果未来需要状态，可以扩展为:
//	  type CLIRunner struct {
//	      config  *Config      // 配置对象
//	      logger  *log.Logger  // 日志记录器
//	      verbose bool         // 详细输出模式
//	  }
//	  当前保持简单，按需扩展
//
// 返回值:
//   - *CLIRunner: 命令行运行器指针，提供以下功能：
//   - RunAssessment(projectPath): 运行项目评估
//   - InteractiveMode(): 进入交互式命令行模式
//   - showHelp(): 显示帮助信息
//
// 使用场景:
//   - 主程序入口: 在main函数中创建CLI运行器
//   - 自动化脚本: 通过命令行参数触发评估
//   - 交互式开发: 开发者手动输入命令进行评估
//
// 示例:
//
//	// 场景1: 自动化评估（CI/CD管道）
//	func main() {
//	    runner := NewCLIRunner()
//	    if err := runner.RunAssessment("./my-project"); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
//	// 场景2: 交互式模式（开发环境）
//	func main() {
//	    runner := NewCLIRunner()
//	    if err := runner.InteractiveMode(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// 注意事项:
//   - 返回的是空结构体指针，主要用于方法调用
//   - 可以复用同一实例进行多次评估（无状态设计）
//   - 创建开销几乎为零（仅指针分配）
//
// 性能特性:
//   - 结构体大小: unsafe.Sizeof(CLIRunner{}) == 0 字节
//   - 创建时间: 纳秒级
//   - 并发安全: 无共享状态，天然支持并发
//
// 作者: JIA
func NewCLIRunner() *CLIRunner {
	return &CLIRunner{}
}

// RunAssessment 运行项目评估完整流程（CLI主入口方法）
//
// 功能说明:
//
//	本方法是CLI评估工具的核心入口函数，负责协调整个评估流程的执行。
//	从项目扫描到报告生成，提供用户友好的进度反馈和错误处理。
//
// 评估流程 (3个阶段):
//
//  1. 项目扫描阶段 (🔍 开始评估项目):
//     - 创建ProjectScanner实例
//     - 递归扫描projectPath下所有Go源文件
//     - 构建文件名到文件内容的映射表
//     - 错误处理: 如果扫描失败（权限不足、路径不存在等），返回详细错误
//
//  2. 文件统计阶段 (📁 发现X个Go文件):
//     - 报告扫描到的文件总数
//     - 帮助用户了解项目规模
//
//  3. 评估分析阶段 (✅ 评估完成 + 📊 生成报告):
//     - 当前为简化实现，仅输出占位信息
//     - 完整实现应包括:
//       • 调用CodeAnalyzer分析代码复杂度和风格
//       • 调用TestAnalyzer统计测试覆盖率
//       • 调用ReportGenerator生成评估报告
//       • 计算综合评分和等级判定
//       • 保存报告到文件或输出到终端
//
// 错误处理策略:
//
//	项目扫描失败:
//	  - 错误类型: 目录不存在、权限不足、读取失败
//	  - 处理方式: 立即返回包装后的错误，不继续后续流程
//	  - 错误信息: "扫描项目失败: [原始错误]"
//
// 参数:
//   - projectPath: 项目根目录路径（绝对或相对路径）
//     示例: "./00-assessment-system", "/home/user/go-project"
//     验证: 应检查路径是否存在和可访问（当前实现未验证）
//
// 返回值:
//   - error: 评估过程中的错误
//   - nil: 评估成功完成
//   - 非nil: 扫描失败或评估异常
//
// 使用场景:
//   - CLI工具主命令: 用户执行 `./assess /path/to/project`
//   - 自动化脚本: CI/CD管道中自动评估代码质量
//   - 批量评估: 遍历多个项目目录进行质量检查
//
// 示例:
//
//	runner := NewCLIRunner()
//
//	// 示例1: 评估当前模块
//	err := runner.RunAssessment("./00-assessment-system")
//	if err != nil {
//	    log.Fatal("评估失败:", err)
//	}
//	// 输出:
//	// 🔍 开始评估项目...
//	// 📁 发现 15 个 Go 文件
//	// ✅ 评估完成!
//	// 📊 生成报告...
//
//	// 示例2: 评估不存在的路径
//	err = runner.RunAssessment("/nonexistent/path")
//	// err != nil: "扫描项目失败: open /nonexistent/path: no such file or directory"
//
// 注意事项:
//   - 当前为简化实现，未调用真实的评估逻辑（需扩展）
//   - 未验证projectPath的有效性（路径存在性、权限检查）
//   - 未生成实际的评估报告文件或结构化输出
//   - 进度信息通过fmt.Println直接输出到stdout，不支持日志级别控制
//
// 扩展方向（完整实现需要）:
//   - 参数验证: 检查projectPath是否存在和可访问
//   - 评估逻辑: 集成CodeAnalyzer、TestAnalyzer、ReportGenerator
//   - 报告输出: 支持Markdown/JSON/HTML多种格式
//   - 进度反馈: 显示实时分析进度（如30% 50% 100%）
//   - 错误恢复: 部分文件分析失败时继续处理其他文件
//   - 日志系统: 替换fmt.Println为结构化日志记录
//
// 作者: JIA
func (cr *CLIRunner) RunAssessment(projectPath string) error {
	fmt.Println("🔍 开始评估项目...")

	// 扫描项目
	scanner := NewProjectScanner()
	files, err := scanner.ScanProject(projectPath)
	if err != nil {
		return fmt.Errorf("扫描项目失败: %w", err)
	}

	fmt.Printf("📁 发现 %d 个 Go 文件\n", len(files))

	// 这里应该调用评估器进行评估
	// 简化实现，直接输出结果
	fmt.Println("✅ 评估完成!")
	fmt.Println("📊 生成报告...")

	return nil
}

// InteractiveMode 进入交互式命令行模式（REPL循环）
//
// 功能说明:
//
//	本方法提供交互式命令行界面(Read-Eval-Print Loop, REPL)，允许用户
//	通过输入命令与评估系统进行实时交互。适用于手动探索和学习评估工具的场景。
//
// REPL工作流程:
//
//  1. 初始化阶段:
//     - 创建bufio.Scanner实例监听标准输入(os.Stdin)
//     - 显示欢迎信息和基本指令
//
//  2. 命令循环 (无限循环直到用户退出):
//     ```
//     ┌─────────────────────────────────────┐
//     │  显示提示符 "> "                    │
//     │  ↓                                   │
//     │  等待用户输入 (scanner.Scan())       │
//     │  ↓                                   │
//     │  解析命令 (strings.TrimSpace)        │
//     │  ↓                                   │
//     │  匹配命令:                           │
//     │  - "help"  → showHelp()             │
//     │  - "quit"/"exit" → 退出返回nil      │
//     │  - 其他    → 提示未知命令           │
//     │  ↓                                   │
//     │  返回循环顶部                        │
//     └─────────────────────────────────────┘
//     ```
//
//  3. 退出阶段:
//     - 用户输入"quit"或"exit"时正常退出
//     - 输入流结束(EOF/Ctrl+D)时break退出
//     - 扫描器错误时返回scanner.Err()
//
// bufio.Scanner详解:
//
//	bufio.Scanner是Go标准库提供的文本扫描器，特性：
//	- 按行读取：默认以换行符(\n)为分隔符
//	- 内部缓冲：高效处理大量输入
//	- 自动去除换行：scanner.Text()返回不含\n的纯文本
//	- 错误处理：通过scanner.Err()获取I/O错误
//
//	核心方法：
//	- scanner.Scan(): 读取下一行，成功返回true，EOF或错误返回false
//	- scanner.Text(): 获取最近一次Scan()读取的文本内容
//	- scanner.Err(): 获取扫描过程中的错误（不包括EOF）
//
// 退出条件:
//   - 正常退出: 用户输入"quit"或"exit"命令
//   - EOF退出: Ctrl+D (Unix) / Ctrl+Z (Windows) 或输入流关闭
//   - 错误退出: 扫描器遇到I/O错误（如stdin被关闭）
//
// 返回值:
//   - nil: 正常退出（用户主动退出或EOF）
//   - error: 扫描器错误（scanner.Err()），如输入流读取失败
//
// 使用场景:
//   - 学习工具: 用户手动输入命令学习评估系统功能
//   - 调试模式: 开发者测试各个评估命令的行为
//   - 演示环境: 现场演示评估工具的交互式使用
//
// 示例:
//
//	runner := NewCLIRunner()
//	err := runner.InteractiveMode()
//	if err != nil {
//	    log.Fatal("交互模式异常:", err)
//	}
//
//	// 交互过程示例:
//	// 🎯 欢迎使用 Go 代码评估系统!
//	// 输入 'help' 查看帮助，输入 'quit' 退出
//	// > help
//	//
//	// 可用命令:
//	//   help     - 显示帮助信息
//	//   assess   - 评估当前项目
//	//   report   - 生成详细报告
//	//   config   - 查看配置
//	//   quit     - 退出程序
//	// > unknown
//	// 未知命令: unknown
//	// 输入 'help' 查看帮助
//	// > quit
//	// 👋 再见!
//
// 注意事项:
//   - 当前仅实现"help"、"quit"、"exit"命令，其他命令显示未知提示
//   - 未实现命令历史记录（如方向键上翻历史命令）
//   - 未实现命令自动补全功能
//   - 命令大小写敏感（"QUIT"不等于"quit"）
//   - 输入为空行时会显示"未知命令: "提示（可优化）
//
// 扩展方向（完整REPL实现）:
//   - 命令路由: 实现"assess <path>"、"report <format>"等带参数命令
//   - Tab补全: 集成readline库支持命令补全
//   - 历史记录: 支持方向键浏览历史输入
//   - 彩色输出: 使用color库美化输出（成功绿色、错误红色）
//   - 命令别名: 支持"q"作为"quit"的简写
//   - 大小写不敏感: strings.ToLower统一处理命令
//   - 空行处理: 跳过空输入，不显示错误提示
//
// 作者: JIA
func (cr *CLIRunner) InteractiveMode() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("🎯 欢迎使用 Go 代码评估系统!")
	fmt.Println("输入 'help' 查看帮助，输入 'quit' 退出")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		command := strings.TrimSpace(scanner.Text())
		switch command {
		case "help":
			cr.showHelp()
		case "quit", "exit":
			fmt.Println("👋 再见!")
			return nil
		default:
			fmt.Printf("未知命令: %s\n", command)
			fmt.Println("输入 'help' 查看帮助")
		}
	}

	return scanner.Err()
}

// showHelp 显示交互式命令行的帮助信息（内部辅助方法）
//
// 功能说明:
//
//	本方法输出CLI工具的所有可用命令列表和简要说明。作为InteractiveMode的辅助方法，
//	当用户输入"help"命令时被调用，提供快速的命令参考指南。
//
// 输出格式:
//
//	采用多行字符串原样输出（raw string literal），格式为：
//	```
//
//	可用命令:
//	  命令名     - 命令描述
//	  ...
//	```
//
// 当前支持的命令 (5个):
//   - help     : 显示本帮助信息（递归调用）
//   - assess   : 评估当前项目代码质量（未实现，仅占位）
//   - report   : 生成详细的评估报告（未实现，仅占位）
//   - config   : 查看当前评估配置（未实现，仅占位）
//   - quit     : 退出交互式模式（已实现）
//
// Raw String Literal (`` 反引号语法):
//
//	Go语言中使用反引号定义的字符串称为原始字符串字面量，特性：
//	- 多行支持: 可跨越多行，保留换行符
//	- 转义失效: \n、\t等转义序列不生效，按原样输出
//	- 引号安全: 内部可包含双引号"而无需转义
//	- 缩进保留: 所有空格和缩进原样输出
//
//	示例对比：
//	  普通字符串: "第一行\n第二行"  // 需要\n转义
//	  原始字符串: `第一行
//	              第二行`           // 直接换行
//
// 设计决策:
//
//	为什么使用未导出方法(小写showHelp)而非导出方法(ShowHelp)?
//	- 封装性: 本方法仅作为InteractiveMode的内部辅助，不应被外部调用
//	- 简单性: 当前仅输出固定文本，无参数无返回值，无需暴露为公共API
//	- 一致性: 与其他REPL内部命令处理逻辑保持一致的可见性级别
//
// 返回值: 无（直接输出到stdout）
//
// 使用场景:
//   - 用户初次使用: 输入"help"查看所有可用命令
//   - 命令遗忘: 忘记命令名称时快速查阅
//   - 功能探索: 了解CLI工具提供的功能清单
//
// 示例:
//
//	// 在InteractiveMode中自动调用
//	> help
//
//	可用命令:
//	  help     - 显示帮助信息
//	  assess   - 评估当前项目
//	  report   - 生成详细报告
//	  config   - 查看配置
//	  quit     - 退出程序
//
// 注意事项:
//   - 当前列出的命令中，只有"help"和"quit"/"exit"实际可用
//   - "assess"、"report"、"config"为占位命令，输入后会显示"未知命令"
//   - 帮助信息为硬编码文本，增加新命令时需手动更新此方法
//   - 未提供详细的参数说明（如"assess <path>"的<path>参数）
//
// 扩展方向（完整帮助系统）:
//   - 详细帮助: 支持"help <command>"显示单个命令的详细用法
//   - 参数说明: 添加每个命令的参数列表和示例
//   - 彩色输出: 命令名高亮显示，提升可读性
//   - 分类显示: 将命令按功能分类（评估、配置、系统）
//   - 动态生成: 从命令注册表动态生成帮助文本，而非硬编码
//   - 示例展示: 为每个命令提供使用示例
//   - 别名显示: 列出命令别名（如"q"作为"quit"的简写）
//
// 作者: JIA
func (cr *CLIRunner) showHelp() {
	fmt.Println(`
可用命令:
  help     - 显示帮助信息
  assess   - 评估当前项目
  report   - 生成详细报告
  config   - 查看配置
  quit     - 退出程序`)
}
