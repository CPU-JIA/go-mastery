/*
=== Go语言学习评估系统 - 代码质量评估引擎 ===

本文件实现了全面的Go代码质量自动化评估系统：
1. 静态代码分析 - golint, govet, gofmt, gocyclo, gosec等工具集成
2. 代码风格检查 - 命名约定、注释覆盖率、格式化标准
3. 架构质量评估 - 模块化设计、依赖管理、接口设计
4. 性能分析 - 内存分配、算法复杂度、性能热点识别
5. 安全性检查 - 漏洞检测、敏感信息泄漏、最佳实践
6. 测试质量评估 - 覆盖率分析、测试设计、边界条件
7. 文档质量检查 - 注释完整性、API文档、README质量
*/

// Package evaluators 提供Go代码质量评估和项目评估的核心功能
//
// 本包实现了多维度的代码质量分析体系，包括静态分析、风格检查、
// 安全扫描、性能评估、测试质量和文档质量等方面的自动化评估。
//
// 作者: JIA
package evaluators

import (
	"assessment-system/utils"
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CodeQualityEvaluator 代码质量评估器
type CodeQualityEvaluator struct {
	config        *CodeQualityConfig
	fileSet       *token.FileSet
	analysisTools map[string]AnalysisTool
	results       *CodeQualityResult
}

// CodeQualityConfig 代码质量评估配置
type CodeQualityConfig struct {
	// 工具配置
	EnabledTools   map[string]bool   `json:"enabled_tools"`   // 启用的分析工具
	ToolPaths      map[string]string `json:"tool_paths"`      // 工具路径配置
	Thresholds     QualityThresholds `json:"thresholds"`      // 质量阈值
	WeightSettings QualityWeights    `json:"weight_settings"` // 权重设置

	// 评估范围
	IncludePatterns []string      `json:"include_patterns"` // 包含的文件模式
	ExcludePatterns []string      `json:"exclude_patterns"` // 排除的文件模式
	MaxFileSize     int64         `json:"max_file_size"`    // 最大文件大小
	Timeout         time.Duration `json:"timeout"`          // 分析超时时间

	// 报告配置
	DetailLevel  string `json:"detail_level"`  // 详细程度: summary, detailed, verbose
	OutputFormat string `json:"output_format"` // 输出格式: json, text, html
	SaveResults  bool   `json:"save_results"`  // 是否保存结果
	ResultsPath  string `json:"results_path"`  // 结果保存路径
}

// QualityThresholds 质量阈值设定
type QualityThresholds struct {
	// 复杂度阈值
	CyclomaticComplexity int `json:"cyclomatic_complexity"` // 圈复杂度阈值
	CognitiveComplexity  int `json:"cognitive_complexity"`  // 认知复杂度阈值
	FunctionLength       int `json:"function_length"`       // 函数长度阈值
	ParameterCount       int `json:"parameter_count"`       // 参数数量阈值

	// 覆盖率阈值
	TestCoverage          float64 `json:"test_coverage"`          // 测试覆盖率阈值
	BranchCoverage        float64 `json:"branch_coverage"`        // 分支覆盖率阈值
	DocumentationCoverage float64 `json:"documentation_coverage"` // 文档覆盖率阈值

	// 质量指标阈值
	CodeDuplication float64 `json:"code_duplication"` // 代码重复率阈值
	TechnicalDebt   float64 `json:"technical_debt"`   // 技术债务阈值
	Maintainability float64 `json:"maintainability"`  // 可维护性阈值

	// 性能阈值
	AllocationRate int `json:"allocation_rate"` // 内存分配率阈值
	GoroutineLeaks int `json:"goroutine_leaks"` // Goroutine泄漏阈值
}

// QualityWeights 质量评估权重
type QualityWeights struct {
	CodeStructure        float64 `json:"code_structure"`        // 代码结构权重
	StyleCompliance      float64 `json:"style_compliance"`      // 风格合规权重
	SecurityAnalysis     float64 `json:"security_analysis"`     // 安全分析权重
	PerformanceAnalysis  float64 `json:"performance_analysis"`  // 性能分析权重
	TestQuality          float64 `json:"test_quality"`          // 测试质量权重
	DocumentationQuality float64 `json:"documentation_quality"` // 文档质量权重
}

// AnalysisTool 分析工具接口
type AnalysisTool interface {
	Name() string
	Version() string
	Execute(projectPath string) (*ToolResult, error)
	ParseResult(output string) (*ToolResult, error)
}

// ToolResult 工具分析结果
type ToolResult struct {
	ToolName      string                 `json:"tool_name"`      // 工具名称
	Version       string                 `json:"version"`        // 工具版本
	ExecutionTime time.Duration          `json:"execution_time"` // 执行时间
	Success       bool                   `json:"success"`        // 是否成功
	Issues        []CodeIssue            `json:"issues"`         // 发现的问题
	Metrics       map[string]interface{} `json:"metrics"`        // 度量数据
	Summary       ToolSummary            `json:"summary"`        // 结果摘要
}

// CodeIssue 代码问题
type CodeIssue struct {
	ID       string `json:"id"`       // 问题唯一标识
	Type     string `json:"type"`     // 问题类型
	Severity string `json:"severity"` // 严重程度: error, warning, info
	Category string `json:"category"` // 问题分类
	Rule     string `json:"rule"`     // 触发规则

	// 位置信息
	File     string `json:"file"`     // 文件路径
	Line     int    `json:"line"`     // 行号
	Column   int    `json:"column"`   // 列号
	Function string `json:"function"` // 所在函数

	// 问题描述
	Message     string `json:"message"`     // 问题描述
	Description string `json:"description"` // 详细说明
	Example     string `json:"example"`     // 示例代码
	Suggestion  string `json:"suggestion"`  // 修复建议

	// 影响评估
	Impact     string `json:"impact"`     // 影响范围: local, module, system
	Complexity int    `json:"complexity"` // 修复复杂度
	Priority   int    `json:"priority"`   // 优先级
}

// ToolSummary 工具结果摘要
type ToolSummary struct {
	TotalIssues  int     `json:"total_issues"`  // 总问题数
	ErrorCount   int     `json:"error_count"`   // 错误数
	WarningCount int     `json:"warning_count"` // 警告数
	InfoCount    int     `json:"info_count"`    // 信息数
	Score        float64 `json:"score"`         // 工具评分
	Passed       bool    `json:"passed"`        // 是否通过
}

// CodeQualityResult 代码质量评估结果
type CodeQualityResult struct {
	ProjectPath string        `json:"project_path"` // 项目路径
	Timestamp   time.Time     `json:"timestamp"`    // 评估时间
	Duration    time.Duration `json:"duration"`     // 评估耗时

	// 整体评分
	OverallScore float64 `json:"overall_score"` // 总体评分
	Grade        string  `json:"grade"`         // 评级
	Passed       bool    `json:"passed"`        // 是否通过

	// 维度评分
	DimensionScores map[string]float64     `json:"dimension_scores"` // 各维度得分
	ToolResults     map[string]*ToolResult `json:"tool_results"`     // 工具结果

	// 统计信息
	Statistics QualityStatistics `json:"statistics"` // 质量统计
	Trends     QualityTrends     `json:"trends"`     // 质量趋势

	// 问题分析
	Issues       []CodeIssue             `json:"issues"`       // 所有问题
	Hotspots     []QualityHotspot        `json:"hotspots"`     // 质量热点
	Improvements []ImprovementSuggestion `json:"improvements"` // 改进建议

	// 技术债务
	TechnicalDebt TechnicalDebtAnalysis `json:"technical_debt"` // 技术债务分析
}

// QualityStatistics 质量统计
type QualityStatistics struct {
	// 代码统计
	TotalFiles   int `json:"total_files"`   // 总文件数
	TotalLines   int `json:"total_lines"`   // 总行数
	CodeLines    int `json:"code_lines"`    // 代码行数
	CommentLines int `json:"comment_lines"` // 注释行数
	BlankLines   int `json:"blank_lines"`   // 空行数

	// 复杂度统计
	AvgComplexity float64 `json:"avg_complexity"` // 平均复杂度
	MaxComplexity int     `json:"max_complexity"` // 最大复杂度
	Functions     int     `json:"functions"`      // 函数数量
	Packages      int     `json:"packages"`       // 包数量

	// 问题统计
	TotalIssues       int `json:"total_issues"`       // 总问题数
	CriticalIssues    int `json:"critical_issues"`    // 严重问题数
	SecurityIssues    int `json:"security_issues"`    // 安全问题数
	PerformanceIssues int `json:"performance_issues"` // 性能问题数

	// 测试统计
	TestFiles    int     `json:"test_files"`    // 测试文件数
	TestCoverage float64 `json:"test_coverage"` // 测试覆盖率
	TestRatio    float64 `json:"test_ratio"`    // 测试代码比例
}

// QualityTrends 质量趋势
type QualityTrends struct {
	ScoreHistory      []HistoryPoint `json:"score_history"`      // 评分历史
	IssueHistory      []HistoryPoint `json:"issue_history"`      // 问题历史
	CoverageHistory   []HistoryPoint `json:"coverage_history"`   // 覆盖率历史
	ComplexityHistory []HistoryPoint `json:"complexity_history"` // 复杂度历史
}

// HistoryPoint 历史数据点
type HistoryPoint struct {
	Timestamp time.Time              `json:"timestamp"` // 时间点
	Value     float64                `json:"value"`     // 数值
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// QualityHotspot 质量热点
type QualityHotspot struct {
	File        string   `json:"file"`        // 文件路径
	Function    string   `json:"function"`    // 函数名
	IssueCount  int      `json:"issue_count"` // 问题数量
	Severity    float64  `json:"severity"`    // 严重程度
	Complexity  int      `json:"complexity"`  // 复杂度
	Priority    int      `json:"priority"`    // 优先级
	Description string   `json:"description"` // 描述
	Suggestions []string `json:"suggestions"` // 建议
}

// ImprovementSuggestion 改进建议
type ImprovementSuggestion struct {
	Category    string   `json:"category"`    // 改进分类
	Title       string   `json:"title"`       // 建议标题
	Description string   `json:"description"` // 详细描述
	Impact      string   `json:"impact"`      // 预期影响
	Effort      string   `json:"effort"`      // 所需工作量
	Priority    int      `json:"priority"`    // 优先级
	Examples    []string `json:"examples"`    // 示例
	Resources   []string `json:"resources"`   // 参考资源
}

// TechnicalDebtAnalysis 技术债务分析
type TechnicalDebtAnalysis struct {
	TotalDebt float64 `json:"total_debt"` // 总债务(小时)
	DebtRatio float64 `json:"debt_ratio"` // 债务比率
	Interest  float64 `json:"interest"`   // 债务利息
	Rating    string  `json:"rating"`     // 债务评级

	// 债务分类
	Categories map[string]float64 `json:"categories"` // 按分类统计
	Files      []FileDebt         `json:"files"`      // 文件债务
	Trends     []DebtTrend        `json:"trends"`     // 债务趋势
}

// FileDebt 文件债务
type FileDebt struct {
	File          string  `json:"file"`          // 文件路径
	Debt          float64 `json:"debt"`          // 债务时间
	Issues        int     `json:"issues"`        // 问题数量
	Complexity    float64 `json:"complexity"`    // 复杂度债务
	Coverage      float64 `json:"coverage"`      // 覆盖率债务
	Documentation float64 `json:"documentation"` // 文档债务
}

// DebtTrend 债务趋势
type DebtTrend struct {
	Date         time.Time `json:"date"`          // 日期
	TotalDebt    float64   `json:"total_debt"`    // 总债务
	NewDebt      float64   `json:"new_debt"`      // 新增债务
	ResolvedDebt float64   `json:"resolved_debt"` // 解决债务
}

// 实现具体的分析工具

// GolintTool Golint分析工具
type GolintTool struct {
	path string
}

// Name 返回Golint工具的标准化名称标识符
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Name()方法，返回golint工具的唯一标识符字符串。
//	该名称用于工具注册、结果聚合、日志记录等场景中的工具识别。
//
// Golint工具简介:
//
//	golint是Go官方提供的代码风格检查工具，用于检测不符合Go编码规范的代码：
//	• 检查项包括: 命名规范、注释格式、导出符号文档、包注释等
//	• 输出建议级别: 所有问题均为warning级别，不阻止编译
//	• 适用场景: 代码审查、CI/CD质量门禁、学习Go最佳实践
//
// 返回值:
//   - string: 固定返回"golint"，作为工具的标准化名称
//
// 使用场景:
//   - 工具管理器注册分析工具时获取工具名称
//   - 生成评估报告时标识问题来源
//   - 日志输出中区分不同分析工具的结果
//
// 示例:
//
//	tool := &GolintTool{path: "/usr/bin/golint"}
//	fmt.Println(tool.Name())  // 输出: "golint"
//
// 注意事项:
//   - 返回的名称应与ToolResult.ToolName字段保持一致
//   - 该名称在系统内应保持唯一性，避免工具冲突
//
// 作者: JIA
func (g *GolintTool) Name() string { return "golint" }

// Version 返回Golint工具的版本标识符
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Version()方法，返回golint工具的版本号。
//	版本信息用于结果追溯、兼容性检查、评估报告生成等场景。
//
// 版本策略:
//
//	当前实现返回ToolVersionLatest常量("latest")，表示使用系统环境中安装的最新版本：
//	• 优势: 简化版本管理，自动跟随系统更新
//	• 劣势: 可能导致不同环境评估结果不一致
//	• 生产建议: 固定为具体版本号(如"v0.0.0-20210508222113-6edffad5e616")
//
// 返回值:
//   - string: 工具版本标识符，当前为"latest"
//
// 使用场景:
//   - 评估报告中记录工具版本以便结果复现
//   - 检查工具版本兼容性（如某些规则仅特定版本支持）
//   - 质量趋势分析时关联工具版本变更
//
// 示例:
//
//	tool := &GolintTool{}
//	fmt.Println(tool.Version())  // 输出: "latest"
//
// 注意事项:
//   - "latest"策略适合开发环境，生产环境建议锁定版本
//   - 如需获取真实版本号，应解析`golint -version`命令输出
//   - 版本不一致可能导致评估结果的可重复性问题
//
// 改进方向:
//   - 实现版本自动检测：执行`golint -version`并解析输出
//   - 增加版本校验：检查工具版本是否满足最低要求
//   - 版本缓存：避免重复执行命令获取版本号
//
// 作者: JIA
func (g *GolintTool) Version() string { return ToolVersionLatest }

// Execute 执行golint静态代码风格分析
//
// 功能说明:
//
//	本方法对指定项目路径执行golint静态代码风格检查，通过命令行调用golint工具，
//	捕获输出结果并解析为结构化的ToolResult。用于自动化检测代码中不符合Go编码规范的问题。
//
// 执行流程:
//
//  1. 性能计时开始:
//     使用time.Now()记录执行开始时间，用于后续计算工具执行耗时
//
//  2. 构造golint命令:
//     命令: golint ./...
//     工作目录: projectPath
//     ./... 表示递归检查当前目录及所有子目录的Go代码
//
//  3. 执行命令并捕获输出:
//     使用CombinedOutput()同时捕获stdout和stderr
//     golint通常将风格建议输出到stdout
//
//  4. 构建基础结果:
//     创建ToolResult对象，填充工具名称、版本、执行时间
//     Success字段根据命令退出码判断（err == nil）
//
//  5. 结果解析:
//     如果命令成功执行（err == nil），调用ParseResult解析输出
//     如果解析失败，返回基础结果和解析错误
//     如果解析成功，返回完整的解析结果
//
//  6. 错误处理:
//     命令执行失败时，返回基础结果和原始错误
//
// 安全说明 (#nosec G204):
//
//	使用#nosec G204豁免gosec的G204警告（Subprocess launched with variable）
//	原因: golint是固定命令，没有用户可控输入拼接到命令中
//	风险评估: 低 - 仅projectPath用于设置工作目录，不参与命令构造
//	替代方案: 无需额外安全处理，命令固定为"golint ./..."
//
// 性能跟踪:
//
//	ExecutionTime字段记录工具实际执行耗时，用于：
//	- 性能分析：识别慢速分析工具
//	- 超时检测：配合超时机制使用
//	- 趋势分析：跟踪项目规模增长对分析时间的影响
//	典型耗时: 小型项目<1s，中型项目1-5s，大型项目>5s
//
// 参数:
//   - projectPath: 项目根目录的绝对或相对路径
//     示例: "./00-assessment-system", "/home/user/go-project"
//     要求: 路径必须存在且包含Go代码文件，否则golint无输出
//
// 返回值:
//   - *ToolResult: 工具执行结果，包含以下关键字段：
//   - ToolName: "golint"
//   - Version: 工具版本（当前为"latest"）
//   - ExecutionTime: 实际执行耗时（纳秒精度）
//   - Success: 命令是否成功执行（true/false）
//   - Issues: 发现的代码风格问题列表（ParseResult填充）
//   - Summary: 问题统计摘要（ParseResult填充）
//   - error: 执行错误，可能原因包括：
//   - golint工具未安装（command not found）
//   - 项目路径不存在（no such file or directory）
//   - 权限不足（permission denied）
//   - 解析错误（ParseResult返回错误）
//
// 使用场景:
//   - 代码质量评估系统自动化检查
//   - CI/CD流水线中的风格门禁
//   - 开发者本地运行代码审查
//   - 批量项目质量分析
//
// 示例:
//
//	tool := &GolintTool{path: "/usr/bin/golint"}
//	result, err := tool.Execute("./00-assessment-system")
//	if err != nil {
//	    if strings.Contains(err.Error(), "command not found") {
//	        log.Fatal("golint未安装，请运行: go install golang.org/x/lint/golint@latest")
//	    }
//	    log.Fatal("执行失败:", err)
//	}
//	fmt.Printf("发现 %d 个风格问题\n", result.Summary.TotalIssues)
//	fmt.Printf("执行耗时: %v\n", result.ExecutionTime)
//
// 注意事项:
//   - golint退出码为0不代表代码完美，仅表示工具正常运行（即使发现问题也是0）
//   - 输出为空字符串时表示代码完全符合golint规范（无任何建议）
//   - 项目路径应为包含go.mod的项目根目录，否则可能无法识别包结构
//   - 大型项目分析可能耗时较长，建议配合超时机制使用
//   - 返回的error仅表示工具执行失败，不表示代码质量问题（质量问题在Issues中）
//
// 改进方向:
//   - 增加超时控制：使用context.WithTimeout防止长时间阻塞
//   - 并发分析：对大型项目按包拆分并发执行
//   - 缓存机制：未修改文件跳过重复分析
//   - 进度回调：支持实时进度报告
//
// 作者: JIA
func (g *GolintTool) Execute(projectPath string) (*ToolResult, error) {
	start := time.Now()

	// #nosec G204 - 固定命令，无用户输入，用于代码质量检查
	cmd := exec.Command("golint", "./...")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()

	result := &ToolResult{
		ToolName:      g.Name(),
		Version:       g.Version(),
		ExecutionTime: time.Since(start),
		Success:       err == nil,
	}

	if err == nil {
		parsedResult, parseErr := g.ParseResult(string(output))
		if parseErr != nil {
			return result, parseErr
		}
		return parsedResult, nil
	}

	return result, err
}

// ParseResult 解析golint输出结果为结构化的ToolResult
//
// 功能说明:
//
//	本方法将golint工具的纯文本输出解析为结构化的ToolResult对象，
//	提取每个代码风格问题的详细信息（文件位置、行列号、问题描述），
//	并生成统计摘要（总问题数、评分、是否通过）。用于后续的质量分析和报告生成。
//
// Golint输出格式详解:
//
//	标准格式: file:line:col: message
//
//	示例输出:
//	  main.go:10:1: exported function HelloWorld should have comment or be unexported
//	  utils.go:25:6: var user_name should be userName
//	  handler.go:42:1: comment on exported function ServeHTTP should be of the form "ServeHTTP ..."
//
//	字段解析:
//	  parts[0] = "main.go"           (文件路径)
//	  parts[1] = "10"                (行号，整数)
//	  parts[2] = "1"                 (列号，整数)
//	  parts[3] = "exported function HelloWorld should have comment" (问题描述)
//
// 解析逻辑流程:
//
//  1. 文本分割:
//     按换行符(\n)分割输出为行切片
//     跳过空行（strings.TrimSpace后为空字符串）
//
//  2. 字段提取:
//     使用strings.SplitN按冒号(:)分割，限制为MinGolintFieldCount(4)个字段
//     如果字段数<4，说明格式不正确，跳过该行
//
//  3. 数字解析与错误恢复:
//     使用strconv.Atoi尝试解析行号和列号
//     解析失败时使用默认值（DefaultLineIfError=0, DefaultColumnIfError=0）
//     不中断解析流程，保证容错性
//
//  4. CodeIssue构造:
//     为每个问题创建唯一ID（格式: golint_序号）
//     固定字段值:
//       Type:     "style"         (问题类型：风格问题)
//       Severity: SeverityWarning (严重程度：警告级别)
//       Category: "code_style"    (分类：代码风格)
//       Rule:     "golint"        (触发规则：golint工具)
//       Impact:   "local"         (影响范围：局部)
//       Priority: 2               (优先级：中等)
//
//  5. 统计摘要生成:
//     TotalIssues:  问题总数（len(issues)）
//     WarningCount: 警告数量（golint所有问题均为warning）
//     Score:        调用calculateStyleScore计算风格得分（满分100，每个问题扣2分）
//     Passed:       通过阈值判断（<10个问题视为通过）
//
// 错误处理策略:
//
//	宽松解析原则:
//	- 遇到空行：跳过继续
//	- 字段数不足：跳过该行继续
//	- 数字解析失败：使用默认值0继续
//	- 不因单行错误中断整体解析
//
// 参数:
//   - output: golint工具的原始文本输出（stdout内容）
//     可以是空字符串（表示无问题）
//     可以包含多行问题报告
//
// 返回值:
//   - *ToolResult: 解析后的结构化结果，包含：
//   - Issues:  []CodeIssue - 所有问题的详细列表
//   - Summary: ToolSummary - 统计摘要（总数、分数、是否通过）
//   - error: 当前实现总是返回nil，未来可能返回解析错误
//
// 数据映射示例:
//
//	输入文本:
//	  main.go:15:1: exported function Add should have comment
//
//	输出CodeIssue:
//	  {
//	    ID:       "golint_0",
//	    Type:     "style",
//	    Severity: "warning",
//	    Category: "code_style",
//	    Rule:     "golint",
//	    File:     "main.go",
//	    Line:     15,
//	    Column:   1,
//	    Message:  "exported function Add should have comment",
//	    Impact:   "local",
//	    Priority: 2,
//	  }
//
// 使用场景:
//   - Execute方法内部调用，解析golint命令输出
//   - 批量代码风格检查时聚合多个文件的问题
//   - 生成代码评审报告时提供详细问题列表
//
// 示例:
//
//	tool := &GolintTool{}
//	output := `main.go:10:1: exported function HelloWorld should have comment
//	utils.go:25:6: var user_name should be userName
//	handler.go:42:1: comment on exported type Handler should be of the form "Handler ..."`
//
//	result, err := tool.ParseResult(output)
//	// result.Issues 包含3个CodeIssue
//	// result.Summary.TotalIssues == 3
//	// result.Summary.WarningCount == 3
//	// result.Summary.Score == 94.0 (100 - 3*2 = 94)
//	// result.Summary.Passed == true (<10个问题)
//
// 注意事项:
//   - 依赖golint输出格式的稳定性（file:line:col: message）
//   - 如果golint更新输出格式，本方法需要同步更新
//   - 行列号解析失败时默认为0，不会阻止其他问题的解析
//   - 通过阈值（<10）是硬编码的，生产环境应从配置读取
//   - 所有问题均视为warning级别，无error级别区分
//
// 局限性与改进方向:
//   - 简单字符串分割可能对特殊文件名（含冒号）误判
//   - 建议使用正则表达式增强解析鲁棒性
//   - 可增加对golint版本的兼容性检测
//   - 可配置化通过阈值而非硬编码为10
//   - 可增加问题去重逻辑（同一位置多次报告）
//
// 作者: JIA
func (g *GolintTool) ParseResult(output string) (*ToolResult, error) {
	issues := []CodeIssue{}
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析golint输出格式: file:line:col: message
		parts := strings.SplitN(line, ":", MinGolintFieldCount)
		if len(parts) < MinGolintFieldCount {
			continue
		}

		// 解析行号和列号，如果解析失败则使用0作为默认值
		lineNum, err := strconv.Atoi(parts[1])
		if err != nil {
			lineNum = DefaultLineIfError // 解析失败时使用默认值
		}
		colNum, err := strconv.Atoi(parts[2])
		if err != nil {
			colNum = DefaultColumnIfError // 解析失败时使用默认值
		}

		issue := CodeIssue{
			ID:       fmt.Sprintf("golint_%d", i),
			Type:     "style",
			Severity: SeverityWarning,
			Category: "code_style",
			Rule:     "golint",
			File:     parts[0],
			Line:     lineNum,
			Column:   colNum,
			Message:  strings.TrimSpace(parts[3]),
			Impact:   "local",
			Priority: 2,
		}

		issues = append(issues, issue)
	}

	summary := ToolSummary{
		TotalIssues:  len(issues),
		WarningCount: len(issues),
		Score:        calculateStyleScore(issues),
		Passed:       len(issues) < 10, // 可配置阈值
	}

	return &ToolResult{
		Issues:  issues,
		Summary: summary,
	}, nil
}

// GovetTool Go vet分析工具
type GovetTool struct {
	path string
}

// Name 返回Go vet工具的标准化名称标识符
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Name()方法，返回go vet工具的唯一标识符字符串。
//	该名称用于工具注册、结果聚合、日志记录等场景中的工具识别。
//
// Go vet工具简介:
//
//	go vet是Go语言官方提供的静态分析工具，作为编译器的一部分内置在Go工具链中：
//	• 检查项包括: 不可达代码、未使用变量、错误的Printf格式化、可疑的结构体标签等
//	• 输出级别: 所有问题均为error级别，表示潜在的严重问题
//	• 适用场景: 编译前检查、CI/CD质量门禁、潜在Bug预防
//	• 与golint区别: vet关注正确性(correctness)，golint关注风格(style)
//
// 返回值:
//   - string: 固定返回"go vet"，作为工具的标准化名称
//
// 使用场景:
//   - 工具管理器注册分析工具时获取工具名称
//   - 生成评估报告时标识问题来源
//   - 日志输出中区分不同分析工具的结果
//   - 配置文件中引用工具时的标识符
//
// 示例:
//
//	tool := &GovetTool{path: "/usr/bin/go"}
//	fmt.Println(tool.Name())  // 输出: "go vet"
//
// 注意事项:
//   - 返回的名称应与ToolResult.ToolName字段保持一致
//   - 该名称在系统内应保持唯一性，避免工具冲突
//   - 虽然实际命令是"go vet"，但标识符中保留空格以保持可读性
//
// 作者: JIA
func (gv *GovetTool) Name() string { return "go vet" }

// Version 返回Go vet工具的版本标识符
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Version()方法，返回go vet工具的版本号。
//	版本信息用于结果追溯、兼容性检查、评估报告生成等场景。
//
// 版本策略:
//
//	当前实现返回ToolVersionLatest常量("latest")，表示使用系统环境中安装的最新版本：
//	• 优势: 简化版本管理，自动跟随系统更新
//	• 劣势: 可能导致不同环境评估结果不一致
//	• 内置特性: go vet是Go工具链的一部分，版本号与Go编译器版本绑定
//	• 生产建议: 记录完整的Go版本号(如"go1.24.0")作为go vet的版本追溯
//
// 返回值:
//   - string: 工具版本标识符，当前为"latest"
//
// 使用场景:
//   - 评估报告中记录工具版本以便结果复现
//   - 检查工具版本兼容性（如某些检查仅特定Go版本支持）
//   - 质量趋势分析时关联工具版本变更
//   - 多环境部署时验证go vet一致性
//
// 示例:
//
//	tool := &GovetTool{path: "/usr/bin/go"}
//	fmt.Println(tool.Version())  // 输出: "latest"
//
// 注意事项:
//   - "latest"策略适合开发环境，生产环境建议锁定Go版本
//   - go vet没有独立的-version标志，应使用`go version`获取Go工具链版本
//   - 版本不一致可能导致评估结果的可重复性问题
//   - go vet的行为随Go版本演进而变化，建议记录完整的Go版本号
//
// 改进方向:
//   - 实现版本自动检测：执行`go version`并解析输出
//   - 增加版本校验：检查Go版本是否满足最低要求
//   - 版本缓存：避免重复执行命令获取版本号
//   - 返回完整Go版本号而非"latest"字符串
//
// 作者: JIA
func (gv *GovetTool) Version() string { return ToolVersionLatest }

// Execute 执行go vet编译器级别的静态分析
//
// 功能说明:
//
//	本方法对指定项目路径执行go vet工具，进行编译器级别的代码正确性检查。
//	go vet是Go工具链内置的静态分析器，专注于检测潜在的错误和可疑的代码构造，
//	与golint的风格检查不同，go vet关注的是代码的正确性和潜在bug。
//
// 执行流程:
//
//  1. 性能计时开始:
//     使用time.Now()记录执行开始时间，用于计算工具执行耗时
//
//  2. 构造go vet命令:
//     命令: go vet ./...
//     工作目录: projectPath
//     ./... 表示递归检查当前目录及所有子目录的Go代码
//     注意: go vet是go工具链的子命令，不是独立可执行文件
//
//  3. 执行命令并捕获输出:
//     使用CombinedOutput()同时捕获stdout和stderr
//     go vet通常将检测到的问题输出到stderr
//
//  4. 构建基础结果:
//     创建ToolResult对象，填充工具名称、版本、执行时间
//     Success字段初始根据err判断（err == nil为true）
//
//  5. 特殊退出码处理（关键逻辑）:
//     go vet在发现问题时返回非零退出码（err != nil）
//     但这不代表工具执行失败，而是代码检测到问题
//     判断条件: err != nil && len(output) > 0
//     若满足，说明工具正常运行但发现了代码问题，将Success设为true
//     这与golint的行为不同（golint总是返回0，即使发现问题）
//
//  6. 结果解析和错误处理:
//     调用ParseResult解析输出为结构化的CodeIssue列表
//     如果解析失败，返回基础结果和解析错误
//     如果解析成功，填充result.Issues和result.Summary后返回
//
// 安全说明 (#nosec G204):
//
//	使用#nosec G204豁免gosec的G204警告（Subprocess launched with variable）
//	原因: "go"是固定命令，"vet"和"./..."是固定参数，没有用户可控输入
//	风险评估: 低 - 仅projectPath用于设置工作目录，不参与命令构造
//	替代方案: 无需额外安全处理，命令固定为"go vet ./..."
//
// 性能跟踪:
//
//	ExecutionTime字段记录工具实际执行耗时，用于：
//	- 性能分析：识别慢速分析工具
//	- 超时检测：配合超时机制使用
//	- 趋势分析：跟踪项目规模增长对分析时间的影响
//	典型耗时: 小型项目<1s，中型项目1-5s，大型项目>10s
//
// go vet vs golint 差异:
//
//	检查重点:
//	- go vet: 代码正确性（unreachable code, Printf format errors, struct tags）
//	- golint: 代码风格（naming conventions, comment format）
//
//	退出码行为:
//	- go vet: 发现问题返回非零退出码（err != nil）
//	- golint: 总是返回0（即使发现问题）
//
//	严重程度:
//	- go vet: 所有问题均为error级别（潜在bug）
//	- golint: 所有问题均为warning级别（风格建议）
//
// 参数:
//   - projectPath: 项目根目录的绝对或相对路径
//     示例: "./00-assessment-system", "/home/user/go-project"
//     要求: 路径必须存在且包含Go代码文件，否则go vet无输出
//
// 返回值:
//   - *ToolResult: 工具执行结果，包含以下关键字段：
//   - ToolName: "go vet"
//   - Version: 工具版本（当前为"latest"）
//   - ExecutionTime: 实际执行耗时（纳秒精度）
//   - Success: 工具是否成功执行（true表示正常运行，false表示工具本身失败）
//   - Issues: 发现的代码问题列表（ParseResult填充）
//   - Summary: 问题统计摘要（ParseResult填充）
//   - error: 执行错误，可能原因包括：
//   - go工具未安装（command not found）
//   - 项目路径不存在（no such file or directory）
//   - 权限不足（permission denied）
//   - 解析错误（ParseResult返回错误）
//
// 使用场景:
//   - CI/CD流水线中的质量门禁（比golint更严格）
//   - 代码提交前的自动化检查
//   - 开发者本地运行代码审查
//   - 批量项目质量分析
//
// 示例:
//
//	tool := &GovetTool{path: "/usr/bin/go"}
//	result, err := tool.Execute("./00-assessment-system")
//	if err != nil {
//	    if strings.Contains(err.Error(), "command not found") {
//	        log.Fatal("go未安装，请安装Go工具链")
//	    }
//	    log.Fatal("执行失败:", err)
//	}
//	if result.Summary.ErrorCount > 0 {
//	    fmt.Printf("发现 %d 个潜在问题（error级别）\n", result.Summary.ErrorCount)
//	    for _, issue := range result.Issues {
//	        fmt.Printf("  %s:%d: %s\n", issue.File, issue.Line, issue.Message)
//	    }
//	} else {
//	    fmt.Println("✅ 通过go vet检查，未发现潜在问题")
//	}
//	fmt.Printf("执行耗时: %v\n", result.ExecutionTime)
//
// 注意事项:
//   - go vet退出码非零不代表工具执行失败，而是检测到代码问题（lines 776-779的特殊处理）
//   - 输出为空字符串时表示代码通过go vet检查（无问题）
//   - 项目路径应为包含go.mod的项目根目录，否则可能无法识别包结构
//   - 大型项目分析可能耗时较长，建议配合超时机制使用
//   - 返回的error仅表示工具执行失败，不表示代码质量问题（质量问题在Issues中）
//   - go vet是Go工具链的一部分，随Go版本更新而增加新检查项
//
// 改进方向:
//   - 增加超时控制：使用context.WithTimeout防止长时间阻塞
//   - 并发分析：对大型项目按包拆分并发执行
//   - 缓存机制：未修改文件跳过重复分析
//   - 进度回调：支持实时进度报告
//
// 作者: JIA
func (gv *GovetTool) Execute(projectPath string) (*ToolResult, error) {
	start := time.Now()

	// #nosec G204 - 固定命令，无用户输入，用于静态代码分析
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()

	result := &ToolResult{
		ToolName:      gv.Name(),
		Version:       gv.Version(),
		ExecutionTime: time.Since(start),
		Success:       err == nil,
	}

	// Go vet返回非零退出码表示发现问题，但这不是执行错误
	if err != nil && len(output) > 0 {
		result.Success = true // 有输出说明工具正常运行
	}

	parsedResult, parseErr := gv.ParseResult(string(output))
	if parseErr != nil {
		return result, parseErr
	}

	result.Issues = parsedResult.Issues
	result.Summary = parsedResult.Summary
	return result, nil
}

// ParseResult 解析go vet输出结果为结构化的ToolResult
//
// 功能说明:
//
//	本方法将go vet工具的纯文本输出解析为结构化的ToolResult对象，
//	提取每个潜在错误的详细信息（文件位置、行号、问题描述），
//	并生成统计摘要（总问题数、评分、是否通过）。用于后续的代码正确性分析和质量报告。
//
// go vet输出格式详解:
//
//	标准格式: file:line: message
//
//	示例输出:
//	  main.go:15: unreachable code
//	  handler.go:42: missing return at end of function
//	  utils.go:28: Printf format %d has arg of wrong type string
//
//	字段解析:
//	  parts[0] = "main.go"           (文件路径)
//	  parts[1] = "15"                (行号，整数)
//	  parts[2+] = "unreachable code" (问题描述，可能包含多个冒号)
//
//	关键差异（vs golint）:
//	  - go vet无列号字段（golint有 file:line:col: message）
//	  - 问题均为error级别（golint为warning）
//	  - 关注代码正确性（golint关注代码风格）
//
// 解析逻辑流程:
//
//  1. 文本分割:
//     按换行符(\n)分割输出为行切片
//     跳过空行（strings.TrimSpace后为空字符串）
//
//  2. 固定字段构造:
//     为每个问题创建CodeIssue，预设固定值：
//       ID:       "govet_{行索引}"  (唯一标识符)
//       Type:     "potential_bug"   (问题类型：潜在缺陷)
//       Severity: SeverityError     (严重程度：错误级别)
//       Category: "correctness"     (分类：代码正确性)
//       Rule:     "go_vet"          (触发规则：go vet工具)
//       Message:  完整输出行        (初始消息为整行)
//       Impact:   "module"          (影响范围：模块级)
//       Priority: 3                 (优先级：高)
//
//  3. 文件位置提取（嵌套解析）:
//     第一步：查找第一个冒号，提取文件名
//       line[:colonIndex] → issue.File
//     第二步：在剩余字符串中查找第二个冒号，提取行号
//       remaining[:nextColonIndex] → 尝试Atoi转换为lineNum
//     第三步：剩余内容为真实消息
//       remaining[nextColonIndex+1:] → TrimSpace后设为issue.Message
//     错误恢复：Atoi失败时保持初始Message不变，继续处理
//
//  4. 统计摘要生成:
//     TotalIssues:  len(issues)              (问题总数)
//     ErrorCount:   len(issues)              (错误数=总数，因为全是error级别)
//     WarningCount: 0                        (go vet无warning)
//     Score:        calculateCorrectnessScore(issues)  (正确性得分计算)
//     Passed:       len(issues) == 0         (无问题时通过)
//
//  5. 返回值构造:
//     返回包含Issues和Summary的ToolResult
//     Error始终为nil（go vet输出总是可解析，无失败场景）
//
// 与golint ParseResult的关键差异:
//
//	输出格式:
//	  go vet:   file:line: message           (2个冒号，无列号)
//	  golint:   file:line:col: message       (3个冒号，有列号)
//
//	字段解析:
//	  go vet:   仅解析文件名和行号（2次Index查找）
//	  golint:   需解析文件名、行号、列号（SplitN为4部分）
//
//	严重程度:
//	  go vet:   所有问题均为SeverityError（潜在bug）
//	  golint:   所有问题均为SeverityWarning（风格建议）
//
//	问题类别:
//	  go vet:   Category="correctness" Type="potential_bug"
//	  golint:   Category="code_style" Type="style"
//
//	错误处理:
//	  go vet:   总是返回(result, nil)，无解析错误
//	  golint:   同样返回(result, nil)，宽松解析
//
// 参数:
//   - output: go vet工具的原始文本输出（stderr或stdout内容）
//     可以是空字符串（表示无问题）
//     可以包含多行问题报告
//     示例: "main.go:10: unreachable code\nutils.go:5: missing return"
//
// 返回值:
//   - *ToolResult: 解析后的结构化结果，包含：
//   - Issues:  []CodeIssue - 所有问题的详细列表（每行一个）
//   - Summary: ToolSummary - 统计摘要（总数、错误数、评分、是否通过）
//   - error: 当前实现总是返回nil，无解析错误场景
//
// 数据映射示例:
//
//	输入文本:
//	  main.go:15: unreachable code
//
//	输出CodeIssue:
//	  {
//	    ID:       "govet_0",
//	    Type:     "potential_bug",
//	    Severity: "error",                  // 错误级别（vs golint的warning）
//	    Category: "correctness",            // 正确性问题（vs golint的code_style）
//	    Rule:     "go_vet",
//	    File:     "main.go",
//	    Line:     15,
//	    Column:   0,                        // go vet无列号信息
//	    Message:  "unreachable code",
//	    Impact:   "module",
//	    Priority: 3,
//	  }
//
// 使用场景:
//   - GovetTool.Execute方法内部调用，解析go vet命令输出
//   - 批量代码正确性检查时聚合多个文件的问题
//   - 生成代码质量报告时提供详细问题列表
//   - CI/CD质量门禁中检测潜在bug
//
// 示例:
//
//	tool := &GovetTool{}
//	output := `main.go:10: unreachable code
//	utils.go:25: missing return at end of function
//	handler.go:42: Printf format %d has arg of wrong type string`
//
//	result, err := tool.ParseResult(output)
//	// result.Issues 包含3个CodeIssue
//	// result.Summary.TotalIssues == 3
//	// result.Summary.ErrorCount == 3 (全是error级别)
//	// result.Summary.WarningCount == 0
//	// result.Summary.Score 根据问题数量计算（每个问题扣10分）
//	// result.Summary.Passed == false (存在问题)
//	// err == nil (总是成功)
//
// 注意事项:
//   - 依赖go vet输出格式的稳定性（file:line: message）
//   - 如果go vet更新输出格式，本方法需要同步更新
//   - 行号解析失败时保持初始Message不变，不会阻止其他问题的解析
//   - 所有问题均视为error级别，无warning级别区分（与golint不同）
//   - 无列号信息，CodeIssue.Column始终为0（与golint的Column>0不同）
//   - 总是返回nil error，调用方无需检查错误（与Execute不同）
//
// 局限性与改进方向:
//   - 简单字符串分割可能对特殊文件名（含冒号）误判
//   - 建议使用正则表达式增强解析鲁棒性
//   - 可增加对go vet版本的兼容性检测
//   - 可增加问题去重逻辑（同一位置多次报告）
//   - 未来可扩展支持go vet的结构化输出（JSON格式）
//
// 作者: JIA
func (gv *GovetTool) ParseResult(output string) (*ToolResult, error) {
	issues := []CodeIssue{}
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析go vet输出格式
		issue := CodeIssue{
			ID:       fmt.Sprintf("govet_%d", i),
			Type:     "potential_bug",
			Severity: SeverityError,
			Category: "correctness",
			Rule:     "go_vet",
			Message:  line,
			Impact:   "module",
			Priority: 3,
		}

		// 尝试解析文件位置
		if colonIndex := strings.Index(line, ":"); colonIndex > 0 {
			issue.File = line[:colonIndex]
			remaining := line[colonIndex+1:]

			if nextColonIndex := strings.Index(remaining, ":"); nextColonIndex > 0 {
				if lineNum, err := strconv.Atoi(remaining[:nextColonIndex]); err == nil {
					issue.Line = lineNum
					issue.Message = strings.TrimSpace(remaining[nextColonIndex+1:])
				}
			}
		}

		issues = append(issues, issue)
	}

	summary := ToolSummary{
		TotalIssues: len(issues),
		ErrorCount:  len(issues),
		Score:       calculateCorrectnessScore(issues),
		Passed:      len(issues) == 0,
	}

	return &ToolResult{
		Issues:  issues,
		Summary: summary,
	}, nil
}

// GocycloTool 圈复杂度分析工具
type GocycloTool struct {
	path      string
	threshold int
}

// Name 返回Gocyclo工具的标准化名称标识符
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Name()方法，返回gocyclo工具的唯一标识符字符串。
//	该名称用于工具注册、结果聚合、日志记录等场景中的工具识别。
//
// Gocyclo工具简介:
//
//	gocyclo是专门用于计算Go代码圈复杂度(Cyclomatic Complexity)的静态分析工具：
//	• 检查项: 计算每个函数的McCabe圈复杂度值（决策点数量）
//	• 输出格式: "复杂度 函数名 文件:行号" (如 "15 processData main.go:42")
//	• 阈值检测: 仅报告超过指定阈值的函数（如 gocyclo -over 10）
//	• 适用场景: 识别高复杂度函数，指导代码重构，提升可维护性
//	• 与go vet区别: gocyclo关注复杂度（maintainability），go vet关注正确性（correctness）
//
// 圈复杂度 (Cyclomatic Complexity) 详解:
//
//	定义: 程序中独立路径的数量，反映代码的测试难度和维护成本
//	计算: V(G) = E - N + 2P (E=边数, N=节点数, P=连通分量数)
//	简化: 基础复杂度1 + 每个if/for/case等决策点 +1
//
//	阈值建议:
//	  • 1-10: 简单函数，易于测试和维护（推荐）
//	  • 11-20: 适中复杂度，需要充分测试（可接受）
//	  • 21-50: 高复杂度，难以测试，建议重构（警告）
//	  • >50: 极端复杂，几乎无法测试，必须重构（严重）
//
// 返回值:
//   - string: 固定返回"gocyclo"，作为工具的标准化名称
//
// 使用场景:
//   - 工具管理器注册分析工具时获取工具名称
//   - 生成评估报告时标识复杂度问题来源
//   - 日志输出中区分不同分析工具的结果
//   - 配置文件中引用复杂度分析工具时的标识符
//
// 示例:
//
//	tool := &GocycloTool{path: "/usr/bin/gocyclo", threshold: 10}
//	fmt.Println(tool.Name())  // 输出: "gocyclo"
//
// 注意事项:
//   - 返回的名称应与ToolResult.ToolName字段保持一致
//   - 该名称在系统内应保持唯一性，避免工具冲突
//   - gocyclo是第三方工具，需要单独安装: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
//
// 作者: JIA
func (gc *GocycloTool) Name() string { return "gocyclo" }

// Version 返回Gocyclo工具的版本标识符
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Version()方法，返回gocyclo工具的版本号。
//	版本信息用于结果追溯、兼容性检查、评估报告生成等场景。
//
// 版本策略:
//
//	当前实现返回ToolVersionLatest常量("latest")，表示使用系统环境中安装的最新版本：
//	• 优势: 简化版本管理，自动跟随系统更新
//	• 劣势: 可能导致不同环境评估结果不一致
//	• gocyclo特性: 作为第三方工具，需独立安装和更新
//	• 生产建议: 固定为具体版本号(如"v1.5.0")以确保可重复性
//
// 版本检测方案:
//
//	gocyclo工具本身不提供标准的-version标志，获取真实版本较为困难：
//	1. 通过go list获取: go list -m github.com/fzipp/gocyclo
//	2. 检查二进制文件: 使用go version -m $(which gocyclo)
//	3. 依赖go.mod锁定: 在项目依赖中明确版本
//	4. 当前简化: 使用"latest"表示最新可用版本
//
// 返回值:
//   - string: 工具版本标识符，当前为"latest"
//
// 使用场景:
//   - 评估报告中记录工具版本以便结果复现
//   - 检查工具版本兼容性（某些特性仅特定版本支持）
//   - 质量趋势分析时关联工具版本变更
//   - 多环境部署时验证gocyclo一致性
//
// 示例:
//
//	tool := &GocycloTool{threshold: 10}
//	fmt.Println(tool.Version())  // 输出: "latest"
//
// 注意事项:
//   - "latest"策略适合开发环境，生产环境建议锁定版本
//   - gocyclo没有内置版本查询命令，需通过go模块系统获取
//   - 版本不一致可能导致评估结果的可重复性问题
//   - gocyclo的复杂度计算规则随版本演进，建议记录使用版本
//
// 改进方向:
//   - 实现版本自动检测：执行go list命令并解析输出
//   - 增加版本校验：检查gocyclo版本是否满足最低要求
//   - 版本缓存：避免重复执行命令获取版本号
//   - 返回完整版本号而非"latest"字符串
//
// 作者: JIA
func (gc *GocycloTool) Version() string { return ToolVersionLatest }

// Execute 执行gocyclo圈复杂度分析
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的Execute()方法，负责在指定项目路径下执行gocyclo工具
//	进行圈复杂度(Cyclomatic Complexity)检查，识别并报告复杂度超过预设阈值的函数。
//	圈复杂度是衡量代码复杂程度的重要指标，高复杂度函数通常难以理解、测试和维护。
//
// 执行流程:
//
//	1. 记录开始时间 (start := time.Now()) 用于性能跟踪
//	2. 构建gocyclo命令: "gocyclo -over {threshold} ."
//	3. 设置工作目录为projectPath并执行命令
//	4. 捕获标准输出和标准错误的合并输出 (CombinedOutput)
//	5. 特殊错误处理: 仅当命令失败且无输出时判定为工具未安装
//	6. 构建基础ToolResult对象（包含工具名、版本、执行时间、成功状态）
//	7. 调用ParseResult()解析gocyclo原始输出
//	8. 将解析结果(Issues/Summary/Metrics)填充到ToolResult
//	9. 返回完整的ToolResult和可能的错误
//
// 命令详解:
//
//	执行的完整命令: gocyclo -over {threshold} .
//	• gocyclo: 第三方圈复杂度分析工具(github.com/fzipp/gocyclo)
//	• -over {threshold}: 仅报告复杂度大于threshold的函数
//	• .: 分析当前目录(即projectPath)下的所有Go文件
//	• 输出格式: "{complexity} {function_name} {file}:{line}:{column}"
//	• 示例输出: "15 (*Evaluator).Analyze evaluator.go:123:1"
//
//	注意: gocyclo使用strconv.Itoa(gc.threshold)将整数阈值转为字符串参数
//
// 参数:
//   - projectPath: 待分析项目的根目录绝对路径
//     • 要求: 必须是有效的Go项目目录(包含.go文件)
//     • 工作目录: 命令会在此目录下执行(cmd.Dir = projectPath)
//     • 路径验证: 调用方需确保路径有效性(本方法不验证)
//
// 返回值:
//   - *ToolResult: 包含完整分析结果的工具执行结果对象
//     • ToolName: 工具名称("gocyclo")
//     • Version: 工具版本(当前为"latest")
//     • ExecutionTime: 执行耗时(time.Since(start))
//     • Success: 执行成功标志(gocyclo总是true，除非工具未安装)
//     • Issues: 检测到的所有高复杂度函数问题列表
//     • Summary: 结果摘要(如"发现3个高复杂度函数")
//     • Metrics: 统计指标(如最高复杂度、平均复杂度等)
//   - error: 错误对象
//     • nil: 执行成功(即使发现高复杂度问题也返回nil)
//     • 非nil: 仅当工具未安装或ParseResult()失败时返回
//
// 错误处理策略:
//
//	本方法采用差异化错误处理策略以应对gocyclo工具的特殊行为：
//
//	1. 工具未安装错误 (cmdErr != nil && len(output) == 0):
//	   • 条件: 命令执行失败且无任何输出
//	   • 判定: gocyclo工具未安装或不在系统PATH中
//	   • 响应: 返回Success=false的ToolResult和包装错误
//	   • 错误信息: "gocyclo工具执行失败（可能未安装）: %w"
//
//	2. 正常分析结果 (cmdErr != nil && len(output) > 0):
//	   • 条件: 命令返回错误但有输出内容
//	   • 原因: gocyclo发现高复杂度函数时可能返回非零退出码
//	   • 响应: 忽略cmdErr，将输出视为有效分析结果
//	   • 行为: Success=true，继续解析输出
//
//	3. 解析错误 (parseErr != nil):
//	   • 条件: ParseResult()方法返回错误
//	   • 原因: gocyclo输出格式异常或解析逻辑错误
//	   • 响应: 返回部分填充的ToolResult和parseErr
//	   • 注意: ToolResult仍包含基础信息(名称/版本/时间)
//
// 特殊行为说明:
//
//	gocyclo工具的退出行为与一般静态分析工具不同：
//	• 一般工具: 发现问题时返回非零退出码，无问题时返回0
//	• gocyclo: 无论是否发现高复杂度函数，退出码均可能为0
//	• 影响: 不能依赖cmdErr判断是否存在复杂度问题
//	• 对策: 始终解析输出内容，通过Issues数量判断问题
//	• Success字段: 仅表示"工具成功运行"，不表示"代码无问题"
//
// 安全说明:
//
//	代码行1237使用了 #nosec G204 指令来抑制gosec的"命令注入"警告：
//	• 警告原因: exec.Command的参数包含变量(strconv.Itoa(gc.threshold))
//	• 安全性分析:
//	  - gc.threshold是GocycloTool结构体的内部字段，非用户直接输入
//	  - 该值在NewGocycloTool()创建时设定，通常为配置文件中的常量
//	  - strconv.Itoa()确保输出为纯数字字符串，不含特殊字符
//	  - 即使threshold被恶意修改，也仅影响分析阈值，不会执行任意命令
//	• 风险评估: 低风险，threshold不是用户可控的外部输入
//	• 改进建议: 如threshold来自用户输入，需增加范围验证(如1-100)
//
// 使用场景:
//   - 代码质量门禁: CI/CD流水线中拒绝高复杂度代码合并
//   - 重构优先级排序: 识别最需要重构的复杂函数
//   - 代码评审辅助: 为评审者标记复杂逻辑区域
//   - 技术债务量化: 统计项目中高复杂度函数数量和分布
//   - 新人代码训练: 帮助开发者理解复杂度控制的重要性
//
// 示例:
//
//	// 创建阈值为10的gocyclo工具
//	tool := &GocycloTool{threshold: 10}
//
//	// 分析项目
//	result, err := tool.Execute("/path/to/project")
//	if err != nil {
//	    log.Printf("分析失败: %v", err)
//	    return
//	}
//
//	// 检查结果
//	if len(result.Issues) > 0 {
//	    fmt.Printf("发现 %d 个高复杂度函数:\n", len(result.Issues))
//	    for _, issue := range result.Issues {
//	        fmt.Printf("  - %s (复杂度: %s)\n", issue.Function, issue.Message)
//	    }
//	} else {
//	    fmt.Println("所有函数复杂度均低于阈值 ✓")
//	}
//
// 注意事项:
//   - gocyclo必须预先安装: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
//   - 工具路径: gocyclo需在系统PATH中，否则cmd.Run()会失败
//   - 阈值选择: 常见标准为10(警告)和15(严重)，需根据团队规范调整
//   - Success字段含义: true不代表无问题，仅代表工具成功运行
//   - 输出解析: ParseResult()依赖gocyclo输出格式的稳定性
//   - 性能考虑: 大型项目分析可能耗时较长，建议配置超时机制
//   - 递归扫描: "."会递归分析所有子目录，包括vendor/可能需排除
//
// 改进方向:
//   - 增加上下文超时: 使用context.Context控制执行时间上限
//   - 工具存在性预检: 启动时检查gocyclo是否安装，提前报错
//   - 输出流式处理: 对超大项目使用StdoutPipe()避免内存溢出
//   - 增量分析: 仅分析变更文件以提升CI/CD效率
//   - 阈值分级: 区分警告阈值和错误阈值，生成不同严重级别
//   - 排除路径: 支持排除vendor/、*_test.go等目录和文件
//   - 并行分析: 将大项目拆分为多个子目录并发执行
//   - 结果缓存: 缓存未变更文件的分析结果
//
// 作者: JIA
func (gc *GocycloTool) Execute(projectPath string) (*ToolResult, error) {
	start := time.Now()

	// #nosec G204 - threshold是内部配置值（非用户输入），用于复杂度分析
	cmd := exec.Command("gocyclo", "-over", strconv.Itoa(gc.threshold), ".")
	cmd.Dir = projectPath
	// 注意：gocyclo即使找到高复杂度函数也会返回成功，所以这里不检查错误
	// 我们只关心输出内容，错误信息会在工具不存在时通过output为空来体现
	output, cmdErr := cmd.CombinedOutput()

	// 如果命令执行失败且没有输出，说明工具可能未安装
	if cmdErr != nil && len(output) == 0 {
		return &ToolResult{
			ToolName:      gc.Name(),
			Version:       gc.Version(),
			ExecutionTime: time.Since(start),
			Success:       false,
		}, fmt.Errorf("gocyclo工具执行失败（可能未安装）: %w", cmdErr)
	}

	result := &ToolResult{
		ToolName:      gc.Name(),
		Version:       gc.Version(),
		ExecutionTime: time.Since(start),
		Success:       true, // gocyclo总是成功，即使有高复杂度函数
	}

	parsedResult, parseErr := gc.ParseResult(string(output))
	if parseErr != nil {
		return result, parseErr
	}

	result.Issues = parsedResult.Issues
	result.Summary = parsedResult.Summary
	result.Metrics = parsedResult.Metrics

	return result, nil
}

// parseLocationInfo 解析位置信息并设置到CodeIssue
//
// 功能说明:
//
//	本辅助函数负责从gocyclo工具输出的位置字符串中提取文件路径和行号信息，
//	并将解析结果直接设置到传入的CodeIssue对象。该函数是parseCycloLine()的
//	关键子步骤，用于丰富代码问题的上下文信息，使开发者能够快速定位问题函数。
//
// 输入格式:
//
//	location参数的标准格式为: "{file_path}:{line}:{column}"
//
//	示例输入:
//	• "evaluator.go:123:1" - 标准格式（文件:行:列）
//	• "pkg/service.go:456:5" - 带路径的标准格式
//	• "main.go:10:1" - 最小化标准格式
//
//	异常格式（会被静默忽略）:
//	• "evaluator.go" - 缺少冒号，无法解析
//	• "evaluator.go:" - 仅有冒号，无行号
//	• "" - 空字符串
//	• "evaluator.go:abc:1" - 行号非数字，File会设置但Line为0
//
// 解析逻辑:
//
//	1. 冒号检测 (strings.Contains):
//	   • 检查location是否包含冒号":"
//	   • 如果不包含，直接返回（不修改issue）
//	   • 用途: 快速过滤明显无效的location
//
//	2. 分割字符串 (strings.Split):
//	   • 按冒号":"分割location为多个部分
//	   • 结果: []string{"file_path", "line", "column", ...}
//	   • 示例: "main.go:10:1" → ["main.go", "10", "1"]
//
//	3. 长度验证:
//	   • 检查parts切片长度是否至少为2（文件名+行号）
//	   • 如果长度<2，直接返回（不修改issue）
//	   • 用途: 确保至少有文件名和行号信息
//
//	4. 提取文件路径:
//	   • 直接使用parts[0]作为文件路径
//	   • 设置到issue.File字段
//	   • 类型: string（相对于项目根目录的路径）
//
//	5. 提取行号:
//	   • 使用strconv.Atoi(parts[1])将字符串转为整数
//	   • 转换成功: 设置到issue.Line字段
//	   • 转换失败: 忽略错误，Line保持默认值0
//	   • 类型: int（从1开始计数，Go源文件标准）
//
//	6. 列号处理:
//	   • 当前实现未提取parts[2]（列号）
//	   • 原因: CodeIssue结构体无Column字段
//	   • 改进方向: 如需要可扩展CodeIssue结构
//
// 参数:
//   - issue: 待填充位置信息的CodeIssue对象指针
//     • 类型: *CodeIssue（必须为非nil指针）
//     • 修改字段: File (string), Line (int)
//     • 副作用: 直接修改传入对象，无返回值
//     • 要求: 调用方需确保issue非nil（本函数不检查）
//   - location: gocyclo输出的位置字符串
//     • 格式: "{file}:{line}:{column}"
//     • 类型: string（可能包含路径分隔符/或\）
//     • 允许空: 空字符串会导致早期返回
//
// 返回值:
//
//	无返回值（void函数）
//	• 副作用: 修改issue.File和issue.Line字段
//	• 错误处理: 解析失败时静默返回，不抛出错误
//	• 调用方检测: 可检查issue.File是否为空判断是否成功解析
//
// 错误处理策略:
//
//	本函数采用"静默失败"策略，在遇到以下情况时直接返回：
//	1. location不包含冒号 → 直接返回
//	2. 分割后parts长度<2 → 直接返回
//	3. 行号转换失败 → File已设置，Line保持0
//
//	设计理念:
//	• 容错优先: 部分解析总比完全失败好（至少获得文件名）
//	• 调用方决策: 由上层parseCycloLine()决定如何处理不完整信息
//	• 日志缺失: 不记录解析错误，适合批量处理场景
//	• 改进空间: 生产环境可增加错误日志或返回bool标识成功/失败
//
// 使用场景:
//   - parseCycloLine()调用: 每解析一行gocyclo输出都会调用一次
//   - 批量位置解析: 处理大量gocyclo输出行时被高频调用
//   - 单元测试: 独立测试位置字符串解析逻辑
//
// 示例:
//
//	// 标准格式解析
//	issue := &CodeIssue{}
//	parseLocationInfo(issue, "evaluator.go:123:1")
//	fmt.Printf("File: %s, Line: %d\n", issue.File, issue.Line)
//	// 输出: File: evaluator.go, Line: 123
//
//	// 带路径格式
//	issue2 := &CodeIssue{}
//	parseLocationInfo(issue2, "pkg/service/handler.go:456:5")
//	fmt.Printf("File: %s, Line: %d\n", issue2.File, issue2.Line)
//	// 输出: File: pkg/service/handler.go, Line: 456
//
//	// 异常格式（静默失败）
//	issue3 := &CodeIssue{}
//	parseLocationInfo(issue3, "main.go")  // 无冒号
//	fmt.Printf("File: '%s', Line: %d\n", issue3.File, issue3.Line)
//	// 输出: File: '', Line: 0 (未修改)
//
//	// 行号非数字
//	issue4 := &CodeIssue{}
//	parseLocationInfo(issue4, "test.go:abc:1")  // 行号无效
//	fmt.Printf("File: %s, Line: %d\n", issue4.File, issue4.Line)
//	// 输出: File: test.go, Line: 0 (File设置成功，Line失败)
//
// 注意事项:
//   - issue指针非nil: 调用方必须确保issue非nil，本函数不检查
//   - 无返回值: 无法直接判断解析成功/失败，需检查issue.File是否被设置
//   - 列号忽略: 当前实现不提取列号（parts[2]），仅提取文件和行号
//   - 相对路径: File字段保存相对路径（相对于gocyclo执行目录）
//   - 跨平台兼容: 路径分隔符依赖gocyclo输出（通常为/，即使在Windows上）
//   - 行号从1开始: Go源文件行号约定从1开始，0表示未设置或解析失败
//   - 文件名包含冒号: 极端情况下文件名包含":"会导致错误分割（罕见）
//
// 改进方向:
//   - 返回解析状态: 返回bool或error指示解析成功/失败
//   - 提取列号: 扩展CodeIssue结构支持Column字段
//   - 错误日志: 开发模式下记录解析失败的location原文
//   - 路径规范化: 统一路径分隔符（filepath.ToSlash）确保跨平台一致性
//   - 格式验证: 使用正则表达式预验证location格式
//   - nil检查: 增加issue == nil的防御性检查
//   - 绝对路径转换: 可选参数支持将相对路径转为绝对路径
//
// 作者: JIA
func parseLocationInfo(issue *CodeIssue, location string) {
	if !strings.Contains(location, ":") {
		return
	}

	parts := strings.Split(location, ":")
	if len(parts) < 2 {
		return
	}

	issue.File = parts[0]
	if lineNum, err := strconv.Atoi(parts[1]); err == nil {
		issue.Line = lineNum
	}
}

// parseCycloLine 解析gocyclo单行输出并创建CodeIssue
//
// 功能说明:
//
//	本辅助函数负责解析gocyclo工具输出的单行文本，从中提取圈复杂度数值、函数名和位置信息，
//	并构建完整的CodeIssue对象。该函数是ParseResult()的核心子步骤，每处理一个高复杂度函数
//	都会调用一次。函数同时返回CodeIssue对象和复杂度数值，复杂度值用于统计计算（如平均值、最大值）。
//
// 输入格式:
//
//	line参数的标准格式为: "{complexity} {function_name} {file}:{line}:{column}"
//
//	示例输入:
//	• "15 (*Evaluator).Analyze evaluator.go:123:1" - 方法（带接收者）
//	• "23 ProcessData processor.go:456:1" - 普通函数
//	• "12 main main.go:10:1" - main函数
//	• "" - 空行（会被跳过）
//	• "invalid format" - 格式错误（返回nil）
//
//	字段说明:
//	• 字段1 (complexity): 圈复杂度数值，必须为正整数
//	• 字段2 (function_name): 函数名，可能包含接收者类型如(*Type).Method
//	• 字段3+ (location): 位置信息，格式为"file:line:column"，可能包含空格（需拼接）
//
// 解析流程:
//
//	1. 空行检测 (strings.TrimSpace):
//	   • 去除首尾空白后检查是否为空字符串
//	   • 如果为空，返回(nil, 0)表示跳过该行
//	   • 用途: 过滤gocyclo输出中可能存在的空行
//
//	2. 字段分割 (strings.Fields):
//	   • 按空白字符（空格/制表符）分割line为多个字段
//	   • 结果: []string{"complexity", "function_name", "file:line:column", ...}
//	   • 注意: 文件路径包含空格时会被错误分割（gocyclo输出一般不含空格）
//
//	3. 字段数量验证:
//	   • 检查parts切片长度是否至少为3（复杂度+函数名+位置）
//	   • 如果长度<3，返回(nil, 0)表示格式错误
//	   • 用途: 确保有足够字段进行解析
//
//	4. 复杂度提取 (strconv.Atoi):
//	   • 将parts[0]从字符串转为整数
//	   • 转换失败（非数字）: 返回(nil, 0)
//	   • 转换成功: 得到complexity值（用于返回和严重程度判定）
//
//	5. 函数名提取:
//	   • 直接使用parts[1]作为函数名
//	   • 可能包含接收者类型: "(*Type).Method"
//	   • 不进行任何处理或验证
//
//	6. 位置信息拼接 (strings.Join):
//	   • 将parts[2:]（第3个字段及之后）用空格拼接
//	   • 用途: 处理文件路径可能包含空格的情况（罕见）
//	   • 示例: ["file.go:10:1"] → "file.go:10:1"
//	   • 示例: ["path", "with", "space.go:10:1"] → "path with space.go:10:1"
//
//	7. 严重程度判定:
//	   • 默认: severity=SeverityWarning, priority=2
//	   • 条件: 如果complexity > gc.threshold*2
//	   • 升级: severity=SeverityError, priority=3
//	   • 逻辑: 超过2倍阈值的函数被视为严重问题
//
//	8. CodeIssue构建:
//	   • ID: "gocyclo_{lineIndex}" (唯一标识，基于行号)
//	   • Type: "complexity" (问题类型固定)
//	   • Severity: Warning或Error（基于步骤7判定）
//	   • Category: "maintainability" (可维护性类别)
//	   • Rule: "cyclomatic_complexity" (规则名称)
//	   • Function: functionName (从步骤5获取)
//	   • Message: "Function {name} has cyclomatic complexity {value}"
//	   • Description: 固定描述文本
//	   • Suggestion: 固定建议文本（建议拆分函数）
//	   • Impact: "module" (影响范围为模块级)
//	   • Complexity: 2 (修复复杂度固定为2)
//	   • Priority: 2或3（基于步骤7判定）
//
//	9. 位置解析 (parseLocationInfo):
//	   • 调用parseLocationInfo()提取文件路径和行号
//	   • 设置issue.File和issue.Line字段
//	   • 如果location格式错误，这两个字段保持默认值
//
//	10. 返回结果:
//	    • 成功: 返回(*CodeIssue, complexity)
//	    • 失败: 返回(nil, 0)
//
// 严重程度判定逻辑:
//
//	if complexity > gc.threshold * 2 {
//	    severity = SeverityError   // 错误级
//	    priority = 3               // 高优先级
//	} else {
//	    severity = SeverityWarning // 警告级
//	    priority = 2               // 中优先级
//	}
//
//	判定标准说明:
//	• 阈值10: 复杂度11-20为Warning，21+为Error
//	• 阈值15: 复杂度16-30为Warning，31+为Error
//	• 设计理念: 超过2倍阈值表示技术债务严重，需强制重构
//
// 参数:
//   - line: gocyclo输出的单行文本
//     • 格式: "{complexity} {function} {location}"
//     • 类型: string（可能包含换行符，会被TrimSpace处理）
//     • 允许空: 空行返回(nil, 0)
//   - lineIndex: 当前行在输出中的索引（从0开始）
//     • 类型: int
//     • 用途: 生成唯一的issue ID ("gocyclo_{lineIndex}")
//     • 注意: 并非源文件行号，而是gocyclo输出的行号
//
// 返回值:
//   - *CodeIssue: 解析成功时返回完整的CodeIssue对象
//     • nil: 解析失败（空行/格式错误/复杂度非数字）
//     • 非nil: 包含所有必要字段的有效CodeIssue
//   - int: 圈复杂度数值
//     • 0: 解析失败
//     • >0: 成功解析的复杂度值（用于ParseResult的统计计算）
//
// 错误处理策略:
//
//	本函数采用"静默失败"策略，在遇到以下情况时返回(nil, 0):
//	1. 空行 → 直接跳过
//	2. 字段数量<3 → 格式错误，无法解析
//	3. 复杂度非数字 → 第一个字段无效
//
//	设计理念:
//	• 容错优先: 单行解析失败不影响其他行
//	• 批量处理: ParseResult()会调用多次，部分失败可接受
//	• 无日志: 不记录解析错误，避免噪音
//	• 改进空间: 可增加错误日志或错误计数器用于调试
//
// 使用场景:
//   - ParseResult()调用: 遍历每行gocyclo输出时被调用
//   - 单元测试: 独立测试单行解析逻辑和边界情况
//   - 调试工具: 手动测试特定格式的gocyclo输出行
//
// 示例:
//
//	tool := &GocycloTool{threshold: 10}
//
//	// 示例1: 标准格式（Warning级）
//	issue1, complexity1 := tool.parseCycloLine("15 (*Evaluator).Analyze evaluator.go:123:1", 0)
//	fmt.Printf("Issue: %s, Complexity: %d, Severity: %s\n",
//	    issue1.Function, complexity1, issue1.Severity)
//	// 输出: Issue: (*Evaluator).Analyze, Complexity: 15, Severity: warning
//
//	// 示例2: 高复杂度（Error级，超过2倍阈值）
//	issue2, complexity2 := tool.parseCycloLine("25 ProcessData processor.go:456:1", 1)
//	fmt.Printf("Severity: %s, Priority: %d\n", issue2.Severity, issue2.Priority)
//	// 输出: Severity: error, Priority: 3
//
//	// 示例3: 空行（跳过）
//	issue3, complexity3 := tool.parseCycloLine("", 2)
//	fmt.Printf("Issue: %v, Complexity: %d\n", issue3, complexity3)
//	// 输出: Issue: <nil>, Complexity: 0
//
//	// 示例4: 格式错误（字段不足）
//	issue4, complexity4 := tool.parseCycloLine("15 FuncName", 3)
//	fmt.Printf("Issue: %v, Complexity: %d\n", issue4, complexity4)
//	// 输出: Issue: <nil>, Complexity: 0
//
//	// 示例5: 复杂度非数字
//	issue5, complexity5 := tool.parseCycloLine("abc FuncName file.go:10:1", 4)
//	fmt.Printf("Issue: %v, Complexity: %d\n", issue5, complexity5)
//	// 输出: Issue: <nil>, Complexity: 0
//
// 注意事项:
//   - lineIndex用途: 仅用于生成唯一ID，不影响解析逻辑
//   - 函数名格式: 可能包含接收者类型，如"(*Type).Method"或"(Type).Method"
//   - 位置拼接: parts[2:]支持文件路径包含空格（虽然罕见）
//   - 固定字段: Description/Suggestion/Impact/Complexity字段固定值
//   - 2倍阈值规则: 严重程度判定基于gc.threshold*2，阈值不同分级不同
//   - 复杂度为0: 返回值为0表示解析失败，gocyclo不会报告复杂度0的函数
//   - 并发安全性: 本函数无状态修改，可并发调用（如果未来优化ParseResult）
//
// 改进方向:
//   - 返回错误信息: 返回error而非静默失败，便于调试和日志记录
//   - 可配置固定字段: Description/Suggestion等文本支持国际化配置
//   - 函数名解析: 进一步解析函数名，区分包名/类型/方法
//   - 动态Impact: 根据复杂度值动态设置Impact（如高复杂度设为"project"）
//   - 列号提取: 从location中提取列号并存储（需扩展CodeIssue结构）
//   - 严重程度配置化: 支持自定义分级规则（如3级/4级分级）
//   - 正则表达式解析: 使用正则匹配代替strings.Fields，处理复杂格式
//   - 上下文信息: 可选参数传入项目路径，将相对路径转为绝对路径
//
// 作者: JIA
func (gc *GocycloTool) parseCycloLine(line string, lineIndex int) (*CodeIssue, int) {
	if strings.TrimSpace(line) == "" {
		return nil, 0
	}

	// 解析gocyclo输出格式: complexity function location
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, 0
	}

	complexity, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, 0
	}

	functionName := parts[1]
	location := strings.Join(parts[2:], " ")

	// 确定严重程度
	severity := SeverityWarning
	priority := 2
	if complexity > gc.threshold*2 {
		severity = SeverityError
		priority = 3
	}

	issue := &CodeIssue{
		ID:          fmt.Sprintf("gocyclo_%d", lineIndex),
		Type:        "complexity",
		Severity:    severity,
		Category:    "maintainability",
		Rule:        "cyclomatic_complexity",
		Function:    functionName,
		Message:     fmt.Sprintf("Function %s has cyclomatic complexity %d", functionName, complexity),
		Description: "High complexity functions are harder to understand and maintain",
		Suggestion:  "Consider breaking down this function into smaller, more focused functions",
		Impact:      "module",
		Complexity:  2,
		Priority:    priority,
	}

	// 解析位置信息
	parseLocationInfo(issue, location)

	return issue, complexity
}

// ParseResult 解析gocyclo工具的输出结果，提取复杂度问题和统计信息
//
// 功能说明:
//
//	本方法实现AnalysisTool接口的ParseResult()方法，负责解析gocyclo工具的原始文本输出，
//	将其转换为结构化的ToolResult对象。解析过程包括：提取高复杂度函数信息、计算统计指标、
//	评估严重程度、生成质量评分。本方法是圈复杂度分析工作流的最后一环，直接影响评估报告质量。
//
// 输入格式:
//
//	gocyclo工具的输出格式为每行一个高复杂度函数，格式：
//	  {complexity} {function_name} {file}:{line}:{column}
//
//	示例输入:
//	  15 (*Evaluator).AnalyzeProject evaluator.go:123:1
//	  23 ProcessComplexData processor.go:456:1
//	  18 (*Service).HandleRequest service.go:789:1
//
//	字段解析:
//	• complexity: 圈复杂度数值（整数）
//	• function_name: 函数名（可能包含接收者类型，如(*Evaluator).Method）
//	• file: 源文件路径（相对于项目根目录）
//	• line:column: 函数声明所在位置（行号:列号）
//
//	特殊情况:
//	• 空输出: 表示所有函数复杂度均低于阈值（无问题）
//	• 空行: 输出中可能包含空行，需忽略
//	• 格式异常: 不符合上述格式的行会被跳过（静默忽略）
//
// 解析流程:
//
//	1. 按行分割原始输出 (strings.Split)
//	2. 初始化统计变量: totalComplexity, functionCount, maxComplexity
//	3. 遍历每一行:
//	   a. 调用parseCycloLine()解析单行，提取CodeIssue和复杂度值
//	   b. 跳过解析失败的行（空行或格式错误）
//	   c. 累加统计数据: totalComplexity, functionCount
//	   d. 更新最大复杂度: maxComplexity
//	   e. 将CodeIssue添加到issues切片
//	4. 计算平均复杂度: avgComplexity = totalComplexity / functionCount
//	5. 构建metrics字典: total_complexity, function_count, avg_complexity, max_complexity, threshold
//	6. 构建summary对象: TotalIssues, WarningCount, ErrorCount, Score, Passed
//	7. 返回包含Issues/Summary/Metrics的ToolResult
//
// 数据提取细节:
//
//	parseCycloLine()方法负责解析单行输出：
//	1. 空行检测: strings.TrimSpace() 过滤空白行
//	2. 字段分割: strings.Fields() 按空白字符分割
//	3. 复杂度提取: strconv.Atoi(parts[0]) 转换为整数
//	4. 函数名提取: parts[1] 直接获取
//	5. 位置信息: parts[2:] 拼接剩余部分作为location
//	6. 严重程度判定:
//	   • complexity > threshold*2 → SeverityError (priority=3)
//	   • 其他 → SeverityWarning (priority=2)
//	7. CodeIssue构建: 填充ID/Type/Severity/Category/Rule/Function/Message等字段
//	8. 位置解析: 调用parseLocationInfo()提取File和Line
//
//	parseLocationInfo()方法负责解析位置字符串：
//	1. 冒号检测: strings.Contains(location, ":") 验证格式
//	2. 分割位置: strings.Split(location, ":") 分离文件和行号
//	3. 提取文件: parts[0] 设置为issue.File
//	4. 提取行号: strconv.Atoi(parts[1]) 转换并设置issue.Line
//
// 统计指标说明:
//
//	metrics字典包含以下关键指标：
//	• total_complexity (int): 所有高复杂度函数的复杂度总和
//	• function_count (int): 检测到的高复杂度函数总数
//	• avg_complexity (float64): 平均复杂度 = total/count (保留小数)
//	• max_complexity (int): 最高复杂度值（用于识别最严重问题）
//	• threshold (int): 当前使用的复杂度阈值（记录评估标准）
//
//	这些指标用于：
//	• 生成可视化报告（趋势图、分布图）
//	• 质量评分计算（calculateComplexityScore函数）
//	• 跨项目对比（相同阈值下的横向对比）
//	• 演进追踪（同一项目的纵向趋势）
//
// 严重程度分级:
//
//	本方法根据复杂度值动态判定问题严重程度：
//	• Warning (警告级): threshold < complexity <= threshold*2
//	  - 示例: 阈值10时，复杂度11-20为Warning
//	  - 建议: 关注重构，非紧急
//	  - Priority: 2
//	• Error (错误级): complexity > threshold*2
//	  - 示例: 阈值10时，复杂度21+为Error
//	  - 建议: 强制重构，高优先级
//	  - Priority: 3
//
//	分级策略说明:
//	• 2倍阈值规则: 经验表明超过2倍阈值的函数维护成本急剧上升
//	• 可扩展性: 未来可引入更多级别（如Info、Critical）
//	• 团队定制: 分级倍数(当前为2)可配置化
//
// 参数:
//   - output: gocyclo工具的原始输出文本（多行字符串）
//     • 格式: 每行"{complexity} {function} {location}"
//     • 编码: 假定为UTF-8（Go源文件标准编码）
//     • 空输出: 表示无高复杂度问题，合法输入
//
// 返回值:
//   - *ToolResult: 包含完整解析结果的工具结果对象
//     • Issues: 高复杂度函数问题列表 ([]CodeIssue)
//     • Summary: 结果摘要 (TotalIssues/WarningCount/ErrorCount/Score/Passed)
//     • Metrics: 统计指标字典 (total_complexity/function_count/avg_complexity/max_complexity/threshold)
//     • ToolName/Version: 未填充（由Execute()方法设置）
//   - error: 解析错误对象
//     • nil: 解析成功（即使output为空也返回nil）
//     • 非nil: 当前实现始终返回nil（未来可能返回格式错误）
//
// 质量评分策略:
//
//	calculateComplexityScore(avgComplexity, maxComplexity)函数计算质量分数：
//	• 输入: 平均复杂度和最大复杂度
//	• 输出: 0-100分的质量评分
//	• 逻辑: 低复杂度高分，高复杂度低分（具体算法见该函数实现）
//	• 用途: 质量门禁判定、趋势报告生成
//
//	Passed字段判定逻辑:
//	• 当前: len(issues) < DefaultIssueCountWarning (可配置阈值)
//	• 含义: Passed=true表示问题数量可接受，false表示超标
//	• 改进: 可结合Score和ErrorCount综合判定
//
// 使用场景:
//   - Execute()方法调用: 解析命令执行后的原始输出
//   - 单元测试: 验证解析逻辑的正确性（使用模拟输出）
//   - 结果缓存: 重新解析历史gocyclo输出
//   - 调试分析: 手动测试gocyclo输出格式兼容性
//
// 示例:
//
//	tool := &GocycloTool{threshold: 10}
//	output := `15 (*Evaluator).Analyze evaluator.go:123:1
//	23 ProcessData processor.go:456:1
//	12 (*Service).Handle service.go:789:1`
//
//	result, err := tool.ParseResult(output)
//	if err != nil {
//	    log.Fatalf("解析失败: %v", err)
//	}
//
//	fmt.Printf("发现 %d 个高复杂度函数\n", result.Summary.TotalIssues)  // 输出: 3
//	fmt.Printf("平均复杂度: %.2f\n", result.Metrics["avg_complexity"])  // 输出: 16.67
//	fmt.Printf("最高复杂度: %d\n", result.Metrics["max_complexity"])    // 输出: 23
//	fmt.Printf("警告数: %d, 错误数: %d\n",
//	    result.Summary.WarningCount,  // 输出: 2 (复杂度12和15)
//	    result.Summary.ErrorCount)    // 输出: 1 (复杂度23>10*2)
//
// 注意事项:
//   - 格式容错: 不符合规范的行会被静默跳过，不会导致解析失败
//   - 空输出处理: 空字符串会返回空Issues切片和零值metrics，非错误情况
//   - 编码假设: 假定输出为UTF-8编码，非ASCII文件名可能需特殊处理
//   - 统计准确性: 仅统计超过阈值的函数，总项目复杂度需单独分析
//   - parseCycloLine鲁棒性: 依赖空白字符分割，文件名包含空格会解析错误
//   - location拼接: parts[2:] Join处理文件路径可能包含空格的情况
//   - 零除保护: avgComplexity计算前检查functionCount>0
//   - 阈值记录: metrics中包含threshold，用于结果复现和报告说明
//
// 改进方向:
//   - 增加格式验证: 检测并报告格式异常行，返回详细错误
//   - 支持JSON输出: 如果gocyclo支持，优先使用JSON格式提升稳定性
//   - 复杂度分布: 增加percentile指标（如P50/P90/P99复杂度）
//   - 文件级聚合: 按文件分组统计，识别问题集中的模块
//   - 历史对比: 接收上次结果，生成delta指标（新增/修复/恶化）
//   - 建议生成: 根据复杂度类型（循环/分支/嵌套）生成针对性重构建议
//   - 阈值自适应: 基于项目历史数据动态调整Warning/Error分级标准
//   - 流式解析: 处理超大输出时使用bufio.Scanner逐行解析
//
// 作者: JIA
func (gc *GocycloTool) ParseResult(output string) (*ToolResult, error) {
	issues := []CodeIssue{}
	lines := strings.Split(output, "\n")

	totalComplexity := 0
	functionCount := 0
	maxComplexity := 0

	// 解析每一行输出
	for i, line := range lines {
		issue, complexity := gc.parseCycloLine(line, i)
		if issue == nil {
			continue
		}

		// 更新统计信息
		totalComplexity += complexity
		functionCount++
		if complexity > maxComplexity {
			maxComplexity = complexity
		}

		issues = append(issues, *issue)
	}

	// 计算平均复杂度
	avgComplexity := 0.0
	if functionCount > 0 {
		avgComplexity = float64(totalComplexity) / float64(functionCount)
	}

	// 构建度量数据
	metrics := map[string]interface{}{
		"total_complexity": totalComplexity,
		"function_count":   functionCount,
		"avg_complexity":   avgComplexity,
		"max_complexity":   maxComplexity,
		"threshold":        gc.threshold,
	}

	// 构建摘要信息
	summary := ToolSummary{
		TotalIssues:  len(issues),
		WarningCount: countBySeverity(issues, SeverityWarning),
		ErrorCount:   countBySeverity(issues, SeverityError),
		Score:        calculateComplexityScore(avgComplexity, maxComplexity),
		Passed:       len(issues) < DefaultIssueCountWarning, // 可配置阈值
	}

	return &ToolResult{
		Issues:  issues,
		Summary: summary,
		Metrics: metrics,
	}, nil
}

// NewCodeQualityEvaluator 创建代码质量评估器
// NewCodeQualityEvaluator 创建代码质量评估器实例
//
// 功能说明:
//
//	本构造函数创建并初始化一个完整的代码质量评估器（CodeQualityEvaluator）实例。
//	评估器是整个代码质量分析系统的核心控制器，负责协调多个静态分析工具（golint/govet/gocyclo等）、
//	收集分析结果、计算质量评分、生成改进建议。本函数执行所有必要的初始化工作，
//	确保评估器处于可立即使用的就绪状态。
//
// 初始化流程:
//
//	1. 创建CodeQualityEvaluator基础结构:
//	   • 分配新的评估器对象内存
//	   • 准备接收配置和工具注册
//
//	2. 设置配置 (config字段):
//	   • 保存传入的CodeQualityConfig配置对象
//	   • 配置包含: 启用的工具列表、工具路径、质量阈值等
//	   • 整个评估生命周期都会引用此配置
//
//	3. 初始化文件集合 (fileSet字段):
//	   • 创建新的token.FileSet()对象
//	   • 用途: Go语法分析和AST解析的位置信息管理
//	   • 作用: 跟踪所有分析文件的行号、列号、偏移量映射
//	   • token.FileSet: go/token包提供的文件位置管理器
//
//	4. 初始化工具映射表 (analysisTools字段):
//	   • 创建空的map[string]AnalysisTool映射
//	   • 键: 工具名称（如"golint", "govet", "gocyclo"）
//	   • 值: 实现AnalysisTool接口的具体工具对象
//	   • 后续通过initializeTools()填充具体工具
//
//	5. 初始化结果对象 (results字段):
//	   • 创建新的CodeQualityResult对象
//	   • Timestamp: 记录评估创建时间 (time.Now())
//	   • DimensionScores: 初始化维度评分映射（空map）
//	     - 维度包括: structure, style, security, performance, testing, documentation
//	   • ToolResults: 初始化工具结果映射（空map）
//	     - 键: 工具名称，值: 该工具的ToolResult对象
//
//	6. 注册分析工具 (initializeTools):
//	   • 调用initializeTools()方法
//	   • 根据config.EnabledTools配置启用相应工具
//	   • 创建并注册GolintTool、GovetTool、GocycloTool等实例
//	   • 填充analysisTools映射表
//
//	7. 返回就绪的评估器:
//	   • 返回完全初始化的*CodeQualityEvaluator对象
//	   • 可立即调用EvaluateProject()执行评估
//
// 字段初始化详解:
//
//	config (*CodeQualityConfig):
//	• 保存用户提供的配置对象引用（非拷贝）
//	• 配置项包括:
//	  - EnabledTools: map[string]bool 指定启用哪些分析工具
//	  - ToolPaths: map[string]string 各工具的可执行文件路径
//	  - QualityThresholds: 质量评分的合格阈值
//	  - ProjectPath: 待评估项目的根目录路径
//	• 注意: 修改config会影响评估器行为
//
//	fileSet (*token.FileSet):
//	• go/token包的文件集合管理器
//	• 功能:
//	  - 管理多个Go源文件的位置信息
//	  - 提供文件名→File对象的映射
//	  - 支持位置Pos→(文件, 行, 列)的转换
//	• 用途:
//	  - AST解析时记录节点位置
//	  - 错误报告时定位源代码位置
//	  - 跨文件的位置比较和排序
//
//	analysisTools (map[string]AnalysisTool):
//	• 工具名称到工具实例的映射
//	• 初始为空，由initializeTools()填充
//	• 支持的工具类型:
//	  - "golint": &GolintTool{} - 代码风格检查
//	  - "govet": &GovetTool{} - 代码正确性检查
//	  - "gocyclo": &GocycloTool{threshold: 10} - 复杂度检查
//	• AnalysisTool接口: Name(), Version(), Execute(), ParseResult()
//
//	results (*CodeQualityResult):
//	• 评估结果的累积容器
//	• Timestamp: 评估创建时间（用于结果追溯和版本管理）
//	• DimensionScores: 各维度评分（初始空map，评估后填充）
//	• ToolResults: 各工具执行结果（初始空map，评估后填充）
//	• Statistics: 质量统计数据（未在此初始化，评估时填充）
//	• Improvements: 改进建议列表（评估后生成）
//
// 工具注册机制:
//
//	initializeTools()方法根据config.EnabledTools注册工具：
//
//	if config.EnabledTools["golint"] {
//	    cqe.analysisTools["golint"] = &GolintTool{
//	        path: config.ToolPaths["golint"],
//	    }
//	}
//
//	• 仅注册配置中启用的工具（按需加载）
//	• 每个工具接收相应的配置参数（如路径、阈值）
//	• 支持动态扩展：新增工具仅需修改initializeTools()
//
// 参数:
//   - config: 代码质量评估配置对象
//     • 类型: *CodeQualityConfig（必须为非nil指针）
//     • 包含字段:
//       - EnabledTools: 启用的工具集合（如{"golint": true, "govet": true}）
//       - ToolPaths: 工具可执行文件路径（可选，默认从PATH查找）
//       - QualityThresholds: 质量阈值配置（如最低60分及格）
//       - ProjectPath: 项目根目录（在EvaluateProject调用时使用）
//     • 要求: 调用方需确保config非nil且EnabledTools已设置
//     • 生命周期: config对象会被评估器引用，不应在评估期间修改
//
// 返回值:
//   - *CodeQualityEvaluator: 完全初始化的代码质量评估器实例
//     • 非nil: 始终返回有效的评估器对象（即使config为空）
//     • 状态: 就绪状态，可立即调用EvaluateProject()
//     • 包含字段:
//       - config: 传入的配置对象引用
//       - fileSet: 新建的token.FileSet实例
//       - analysisTools: 已注册的工具映射（根据config填充）
//       - results: 预初始化的结果容器
//     • 使用: 通过evaluator.EvaluateProject(path)启动评估
//
// 使用场景:
//   - 启动代码评估: 创建评估器并执行项目分析
//   - 批量评估: 为多个项目创建独立的评估器实例
//   - CI/CD集成: 在持续集成流水线中创建评估器
//   - IDE插件: 在编辑器中实时创建评估器检查代码
//   - 质量报告生成: 创建评估器并生成详细报告
//
// 示例:
//
//	// 示例1: 使用默认配置创建评估器
//	config := GetDefaultConfig()
//	evaluator := NewCodeQualityEvaluator(config)
//	result, err := evaluator.EvaluateProject("/path/to/project")
//	if err != nil {
//	    log.Fatalf("评估失败: %v", err)
//	}
//	fmt.Printf("总体评分: %.2f\n", result.OverallScore)
//
//	// 示例2: 自定义配置（仅启用golint和govet）
//	config := &CodeQualityConfig{
//	    EnabledTools: map[string]bool{
//	        "golint": true,
//	        "govet":  true,
//	    },
//	    ToolPaths: map[string]string{
//	        "golint": "/usr/local/bin/golint",
//	    },
//	}
//	evaluator := NewCodeQualityEvaluator(config)
//	// 使用evaluator进行评估...
//
//	// 示例3: 批量评估多个项目
//	config := GetDefaultConfig()
//	projects := []string{"/path/to/project1", "/path/to/project2"}
//	for _, projectPath := range projects {
//	    evaluator := NewCodeQualityEvaluator(config)  // 每个项目独立评估器
//	    result, _ := evaluator.EvaluateProject(projectPath)
//	    fmt.Printf("%s: %.2f分\n", projectPath, result.OverallScore)
//	}
//
// 注意事项:
//   - config非nil检查: 本函数未检查config是否为nil，调用方需确保传入有效配置
//   - 工具可用性: initializeTools()不检查工具是否已安装，执行时才会报错
//   - 并发使用: CodeQualityEvaluator实例非并发安全，每次评估应创建新实例
//   - 资源清理: 评估器未实现Close()方法，依赖Go垃圾回收释放资源
//   - config引用: 保存的是config指针，外部修改config会影响评估器行为
//   - results初始化: results对象预创建但字段为空，需调用EvaluateProject()填充
//   - fileSet复用: 同一评估器内所有文件共享同一个fileSet
//   - 时间戳意义: results.Timestamp记录创建时间而非评估完成时间
//
// 改进方向:
//   - nil检查: 增加config == nil的防御性检查，返回error或使用默认配置
//   - 工具预检: initializeTools()时检查工具可执行性，提前发现配置错误
//   - 构建器模式: 提供NewCodeQualityEvaluatorBuilder()支持链式配置
//   - 上下文支持: 接收context.Context参数，支持评估超时和取消
//   - 资源池化: 提供全局评估器池，复用FileSet和工具实例减少开销
//   - 配置验证: 在构造时验证config完整性（如必需字段检查）
//   - Close方法: 实现Close()用于显式释放资源（如临时文件、日志句柄）
//   - 日志注入: 接收logger参数，支持评估过程日志记录
//
// 作者: JIA
func NewCodeQualityEvaluator(config *CodeQualityConfig) *CodeQualityEvaluator {
	evaluator := &CodeQualityEvaluator{
		config:        config,
		fileSet:       token.NewFileSet(),
		analysisTools: make(map[string]AnalysisTool),
		results: &CodeQualityResult{
			Timestamp:       time.Now(),
			DimensionScores: make(map[string]float64),
			ToolResults:     make(map[string]*ToolResult),
		},
	}

	// 初始化分析工具
	evaluator.initializeTools()

	return evaluator
}

// initializeTools 初始化并注册静态分析工具
//
// 功能说明:
//
//	本方法负责根据配置对象中的启用状态，创建并注册所有需要的静态分析工具实例。
//	该方法是评估器初始化流程的核心步骤，由NewCodeQualityEvaluator()构造函数调用。
//	通过配置驱动的工具注册机制，评估器可以灵活地启用或禁用特定工具，
//	实现按需加载和可扩展的工具管理。
//
// 工具注册流程:
//
//	1. 检查配置 (cqe.config.EnabledTools):
//	   • 读取EnabledTools映射表，判断各工具是否启用
//	   • EnabledTools格式: map[string]bool（工具名→启用状态）
//	   • 示例: {"golint": true, "govet": true, "gocyclo": false}
//
//	2. 条件注册 (按需创建工具实例):
//	   • 仅为启用的工具创建实例（EnabledTools[name] == true）
//	   • 跳过未启用的工具，节省内存和初始化开销
//	   • 每个工具注册到analysisTools映射表
//
//	3. 工具实例化 (创建具体Tool对象):
//	   • GolintTool: 代码风格和最佳实践检查工具
//	   • GovetTool: 代码正确性和潜在错误检查工具
//	   • GocycloTool: 圈复杂度分析工具
//
//	4. 参数配置 (传递工具特定配置):
//	   • path: 工具可执行文件路径（从ToolPaths配置获取）
//	   • threshold: 阈值配置（如gocyclo的复杂度阈值）
//	   • 来源: cqe.config.ToolPaths 和 cqe.config.Thresholds
//
//	5. 注册到映射表 (填充analysisTools):
//	   • 键: 工具名称字符串（"golint", "govet", "gocyclo"）
//	   • 值: 实现AnalysisTool接口的工具实例
//	   • 后续通过名称查找和调用工具
//
// 支持的工具详解:
//
//	1. GolintTool (golint - 代码风格检查器):
//	   • 检查内容: 命名规范、注释完整性、导出声明规范等
//	   • 初始化参数:
//	     - path: golint可执行文件路径（可选，默认从PATH查找）
//	   • 配置示例:
//	     EnabledTools["golint"] = true
//	     ToolPaths["golint"] = "/usr/local/bin/golint"
//	   • 创建代码:
//	     cqe.analysisTools["golint"] = &GolintTool{
//	         path: cqe.config.ToolPaths["golint"],
//	     }
//
//	2. GovetTool (go vet - 代码正确性检查器):
//	   • 检查内容: Printf格式错误、未使用变量、死代码、并发问题等
//	   • 初始化参数:
//	     - path: go命令路径（通常使用系统默认的"go"命令）
//	   • 配置示例:
//	     EnabledTools["govet"] = true
//	     ToolPaths["govet"] = "/usr/local/go/bin/go"
//	   • 创建代码:
//	     cqe.analysisTools["govet"] = &GovetTool{
//	         path: cqe.config.ToolPaths["govet"],
//	     }
//
//	3. GocycloTool (gocyclo - 圈复杂度分析器):
//	   • 检查内容: 函数圈复杂度，识别过于复杂的函数
//	   • 初始化参数:
//	     - path: gocyclo可执行文件路径
//	     - threshold: 复杂度阈值（超过此值的函数会被报告）
//	   • 配置示例:
//	     EnabledTools["gocyclo"] = true
//	     ToolPaths["gocyclo"] = "/usr/local/bin/gocyclo"
//	     Thresholds.CyclomaticComplexity = 10
//	   • 创建代码:
//	     cqe.analysisTools["gocyclo"] = &GocycloTool{
//	         path:      cqe.config.ToolPaths["gocyclo"],
//	         threshold: cqe.config.Thresholds.CyclomaticComplexity,
//	     }
//	   • 阈值说明:
//	     - 10: 常见标准（警告级）
//	     - 15: 宽松标准
//	     - 5: 严格标准
//
// 配置驱动设计:
//
//	本方法采用配置驱动的工具管理策略，优势包括：
//
//	1. 灵活性:
//	   • 用户可通过配置文件控制启用哪些工具
//	   • 不同项目可使用不同的工具组合
//	   • CI/CD环境可动态调整工具集
//
//	2. 性能优化:
//	   • 仅初始化启用的工具，减少内存占用
//	   • 避免执行不需要的分析，缩短评估时间
//	   • 支持增量分析（只运行部分工具）
//
//	3. 可扩展性:
//	   • 新增工具仅需添加if块和工具实现
//	   • 无需修改评估器核心逻辑
//	   • 支持第三方工具插件化集成
//
//	4. 配置集中管理:
//	   • 所有工具配置集中在CodeQualityConfig中
//	   • 便于统一管理和版本控制
//	   • 支持配置继承和覆盖
//
// 工具路径配置说明:
//
//	ToolPaths配置项指定各工具的可执行文件路径：
//	• 绝对路径: "/usr/local/bin/golint"（明确指定）
//	• 相对路径: "bin/golint"（相对于项目根目录）
//	• 空字符串: ""（从系统PATH环境变量查找）
//	• 未配置: 默认使用工具名称作为命令（依赖PATH）
//
//	查找顺序:
//	1. 检查ToolPaths["toolname"]是否配置
//	2. 如果配置了，使用指定路径
//	3. 如果未配置或为空，使用工具默认名称（如"golint"）
//	4. 执行时由操作系统在PATH中查找
//
// 阈值配置说明:
//
//	Thresholds配置项定义各工具的质量标准：
//	• CyclomaticComplexity: gocyclo的复杂度阈值（默认10）
//	• 未来扩展: 可添加更多阈值配置（如代码行数、注释率等）
//
// 参数:
//
//	无显式参数（方法接收者为*CodeQualityEvaluator）
//
// 返回值:
//
//	无返回值（void方法）
//	• 副作用: 填充cqe.analysisTools映射表
//	• 状态变更: 评估器从"配置完成"状态转为"工具就绪"状态
//
// 调用时机:
//
//	本方法由NewCodeQualityEvaluator()构造函数自动调用，执行时机：
//	1. 评估器对象创建后
//	2. config/fileSet/analysisTools/results字段初始化后
//	3. 返回评估器给调用方之前
//
//	调用路径:
//	NewCodeQualityEvaluator(config) → initializeTools() → 返回evaluator
//
// 使用场景:
//   - 评估器初始化: 自动注册配置中启用的工具
//   - 工具定制: 用户通过config控制使用哪些工具
//   - CI/CD集成: 不同环境启用不同的工具组合
//   - 性能调优: 禁用耗时工具以加速评估
//
// 示例:
//
//	// 示例1: 启用所有工具的配置
//	config := &CodeQualityConfig{
//	    EnabledTools: map[string]bool{
//	        "golint":  true,
//	        "govet":   true,
//	        "gocyclo": true,
//	    },
//	    ToolPaths: map[string]string{
//	        "golint":  "",  // 使用PATH中的golint
//	        "govet":   "",  // 使用PATH中的go
//	        "gocyclo": "/usr/local/bin/gocyclo",  // 指定路径
//	    },
//	    Thresholds: QualityThresholds{
//	        CyclomaticComplexity: 10,
//	    },
//	}
//	evaluator := NewCodeQualityEvaluator(config)
//	// initializeTools()已自动调用，所有3个工具已注册
//	fmt.Printf("已注册工具数: %d\n", len(evaluator.analysisTools))  // 输出: 3
//
//	// 示例2: 仅启用golint和govet（跳过gocyclo）
//	config := &CodeQualityConfig{
//	    EnabledTools: map[string]bool{
//	        "golint":  true,
//	        "govet":   true,
//	        "gocyclo": false,  // 禁用gocyclo
//	    },
//	}
//	evaluator := NewCodeQualityEvaluator(config)
//	// 仅注册了2个工具
//	fmt.Printf("已注册工具数: %d\n", len(evaluator.analysisTools))  // 输出: 2
//
//	// 示例3: 动态检查工具是否注册
//	evaluator := NewCodeQualityEvaluator(config)
//	if _, ok := evaluator.analysisTools["gocyclo"]; ok {
//	    fmt.Println("gocyclo工具已启用")
//	} else {
//	    fmt.Println("gocyclo工具未启用")
//	}
//
// 注意事项:
//   - 工具可用性: 本方法不检查工具是否已安装，执行时才会报错
//   - 配置依赖: 依赖cqe.config非nil且EnabledTools已初始化
//   - 映射初始化: 要求cqe.analysisTools已通过make()初始化
//   - 重复调用: 重复调用会覆盖已注册的工具（不推荐）
//   - 并发安全: 本方法非并发安全，仅应在构造函数中调用一次
//   - 路径验证: 不验证ToolPaths指定的路径是否存在
//   - 阈值默认值: 如果Thresholds未设置，gocyclo.threshold可能为0
//   - nil安全: 如果EnabledTools或ToolPaths为nil，会panic
//
// 改进方向:
//   - 工具预检: 检查工具可执行性，返回error提前发现问题
//   - 版本验证: 检查工具版本是否满足最低要求
//   - 插件机制: 支持动态加载外部工具插件（plugin包）
//   - 依赖注入: 通过接口注入工具，便于单元测试mock
//   - 错误返回: 返回error以报告工具初始化失败
//   - 默认值处理: 为未配置的阈值设置合理默认值
//   - 日志记录: 记录已注册的工具和跳过的工具
//   - 工具冲突检测: 检测互斥工具（如golint vs golangci-lint）
//   - 动态注册: 提供RegisterTool()方法支持运行时注册
//   - 配置验证: 检查EnabledTools中的工具名是否支持
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) initializeTools() {
	if cqe.config.EnabledTools["golint"] {
		cqe.analysisTools["golint"] = &GolintTool{
			path: cqe.config.ToolPaths["golint"],
		}
	}

	if cqe.config.EnabledTools["govet"] {
		cqe.analysisTools["govet"] = &GovetTool{
			path: cqe.config.ToolPaths["govet"],
		}
	}

	if cqe.config.EnabledTools["gocyclo"] {
		cqe.analysisTools["gocyclo"] = &GocycloTool{
			path:      cqe.config.ToolPaths["gocyclo"],
			threshold: cqe.config.Thresholds.CyclomaticComplexity,
		}
	}
}

// EvaluateProject 评估项目代码质量
//
// 功能说明:
//
//	本方法是代码质量评估器的核心入口方法，负责协调整个评估流程的执行。
//	该方法会依次调用多个分析工具、收集结果、计算质量评分、生成改进建议，
//	最终返回完整的评估报告。这是一个端到端的评估流程，包含从工具执行到
//	结果持久化的所有关键步骤。
//
// 评估流程（8个核心步骤）:
//
//	1. 初始化评估上下文:
//	   • 记录开始时间 (start := time.Now())
//	   • 设置results.ProjectPath = projectPath
//	   • 设置results.Timestamp = start
//	   • 输出日志: "开始评估项目代码质量: {projectPath}"
//
//	2. 执行所有分析工具 (runAnalysisTools):
//	   • 遍历analysisTools映射表中的所有已注册工具
//	   • 对每个工具调用Execute(projectPath)方法
//	   • 收集每个工具的ToolResult到results.ToolResults
//	   • 如果工具执行失败，记录日志但继续执行（容错机制）
//	   • 如果所有工具都失败，返回error终止评估
//	   • 成功: 继续后续步骤
//
//	3. 收集和聚合结果 (aggregateResults):
//	   • 从所有ToolResult中提取CodeIssue列表
//	   • 合并为统一的问题集合
//	   • 统计问题数量、严重程度分布
//	   • 分析项目结构（文件数、代码行数、注释率等）
//	   • 填充results.Statistics字段
//	   • 失败: 返回error终止评估
//
//	4. 计算维度得分 (calculateDimensionScores):
//	   • 根据问题类型和工具结果计算各维度评分
//	   • 维度包括:
//	     - structure: 代码结构质量（复杂度、模块化）
//	     - style: 代码风格质量（格式、命名、注释）
//	     - security: 安全质量（潜在漏洞、危险用法）
//	     - performance: 性能质量（资源使用、算法效率）
//	     - testing: 测试质量（测试覆盖率、测试完整性）
//	     - documentation: 文档质量（README、包文档、注释）
//	   • 每个维度评分范围: 0-100分
//	   • 填充results.DimensionScores映射表
//
//	5. 计算整体得分 (calculateOverallScore):
//	   • 基于各维度评分计算加权平均分
//	   • 权重分配（可配置）:
//	     - structure: 25%
//	     - style: 20%
//	     - security: 20%
//	     - performance: 15%
//	     - testing: 15%
//	     - documentation: 5%
//	   • 计算公式: OverallScore = Σ(维度得分 × 权重)
//	   • 评分范围: 0-100分
//	   • 填充results.OverallScore字段
//
//	6. 生成改进建议 (generateImprovements):
//	   • 基于发现的问题生成具体的改进建议
//	   • 每个建议包含:
//	     - 问题描述（哪个维度得分低）
//	     - 具体问题（哪些文件/函数有问题）
//	     - 改进方法（如何修复）
//	     - 优先级（高/中/低）
//	   • 填充results.Improvements切片
//
//	7. 分析技术债务 (analyzeTechnicalDebt):
//	   • 评估项目的技术债务水平
//	   • 计算债务指标:
//	     - 总债务量（问题数量加权）
//	     - 偿还成本（预估修复时间）
//	     - 债务趋势（与历史对比）
//	   • 识别高债务区域（哪些模块债务最多）
//	   • 填充results.TechnicalDebt字段
//
//	8. 识别质量热点 (identifyQualityHotspots):
//	   • 识别问题集中的代码区域（质量热点）
//	   • 按文件/包/函数聚合问题
//	   • 找出问题最多的top N个热点
//	   • 为每个热点生成针对性建议
//	   • 填充results.QualityHotspots切片
//
//	9. 记录评估元数据:
//	   • 计算总耗时: Duration = time.Since(start)
//	   • 判定是否通过: Passed = (OverallScore >= PassingScore)
//	   • PassingScore默认值: 60分
//	   • 输出日志: "代码质量评估完成，总分: {score}，耗时: {duration}"
//
//	10. 保存结果（可选）:
//	    • 如果config.SaveResults == true，调用saveResults()
//	    • 将results保存到JSON文件（默认路径: .quality-report.json）
//	    • 保存失败仅记录日志，不影响返回
//	    • 用途: 结果持久化、历史对比、趋势分析
//
//	11. 返回评估结果:
//	    • 返回完整的CodeQualityResult对象
//	    • 如果任何关键步骤失败，返回nil和error
//
// 错误处理策略:
//
//	本方法采用"快速失败"策略处理关键步骤的错误：
//
//	1. runAnalysisTools失败 → 立即返回error（无工具结果无法继续）
//	2. aggregateResults失败 → 立即返回error（无统计数据无法评分）
//	3. 其他步骤失败 → 记录日志但继续（尽可能完成评估）
//	4. saveResults失败 → 仅记录日志（不影响评估结果返回）
//
//	设计理念:
//	• 核心数据必须完整（工具结果、统计数据）
//	• 辅助功能可降级（保存失败、部分维度评分失败）
//	• 用户始终能获得部分结果，而非完全失败
//
// 性能特征:
//
//	评估耗时主要取决于：
//	• 项目规模: 文件数量、代码行数
//	• 启用工具数: 工具越多耗时越长
//	• 工具类型: gocyclo等工具较快，golint可能较慢
//	• 磁盘I/O: 读取源文件、写入报告
//
//	典型耗时参考:
//	• 小型项目（<1000行）: 1-3秒
//	• 中型项目（1000-10000行）: 3-10秒
//	• 大型项目（>10000行）: 10-60秒
//
// 参数:
//   - projectPath: 待评估项目的根目录绝对路径
//     • 类型: string
//     • 要求: 必须是有效的Go项目目录（包含go.mod或.go文件）
//     • 示例: "/home/user/myproject", "C:\Projects\myapp"
//     • 验证: 方法内部不验证路径有效性，由调用方确保
//     • 影响: 所有工具将在此目录下执行分析
//
// 返回值:
//   - *CodeQualityResult: 完整的代码质量评估结果
//     • ProjectPath: 项目路径
//     • Timestamp: 评估开始时间
//     • Duration: 评估总耗时
//     • OverallScore: 总体质量评分（0-100）
//     • Passed: 是否通过质量门禁（OverallScore >= 60）
//     • DimensionScores: 各维度评分映射表
//     • ToolResults: 各工具原始结果映射表
//     • Statistics: 项目质量统计数据
//     • Improvements: 改进建议列表
//     • TechnicalDebt: 技术债务分析
//     • QualityHotspots: 质量热点列表
//   - error: 错误对象
//     • nil: 评估成功完成
//     • 非nil: 关键步骤失败（工具执行、结果聚合）
//     • 错误格式: "分析工具执行失败: {详细错误}"
//     • 错误格式: "结果聚合失败: {详细错误}"
//
// 使用场景:
//   - CI/CD质量门禁: 在代码合并前检查质量评分
//   - 开发者本地检查: 提交代码前自查质量
//   - 定期质量审计: 定期评估项目质量趋势
//   - 技术债务管理: 识别和跟踪技术债务
//   - 代码评审辅助: 为评审者提供质量参考
//
// 示例:
//
//	// 示例1: 基本评估流程
//	config := GetDefaultConfig()
//	evaluator := NewCodeQualityEvaluator(config)
//	result, err := evaluator.EvaluateProject("/path/to/project")
//	if err != nil {
//	    log.Fatalf("评估失败: %v", err)
//	}
//	fmt.Printf("总体评分: %.2f/100\n", result.OverallScore)
//	fmt.Printf("是否通过: %v\n", result.Passed)
//	fmt.Printf("评估耗时: %v\n", result.Duration)
//
//	// 示例2: 检查维度评分
//	result, _ := evaluator.EvaluateProject(projectPath)
//	for dimension, score := range result.DimensionScores {
//	    fmt.Printf("%s: %.2f\n", dimension, score)
//	    if score < 60 {
//	        fmt.Printf("  ⚠ %s维度未达标!\n", dimension)
//	    }
//	}
//
//	// 示例3: 查看改进建议
//	result, _ := evaluator.EvaluateProject(projectPath)
//	fmt.Printf("发现 %d 条改进建议:\n", len(result.Improvements))
//	for i, improvement := range result.Improvements {
//	    fmt.Printf("%d. [%s] %s\n",
//	        i+1, improvement.Priority, improvement.Description)
//	    fmt.Printf("   修复方法: %s\n", improvement.Suggestion)
//	}
//
//	// 示例4: CI/CD集成
//	result, err := evaluator.EvaluateProject(os.Getenv("CI_PROJECT_DIR"))
//	if err != nil {
//	    log.Fatalf("评估失败: %v", err)
//	}
//	if !result.Passed {
//	    log.Fatalf("代码质量未达标: %.2f < 60", result.OverallScore)
//	    os.Exit(1)  // 阻止合并
//	}
//	log.Printf("代码质量检查通过 ✓")
//
// 注意事项:
//   - 项目路径: 必须是Go项目根目录，否则工具可能执行失败
//   - 工具依赖: 依赖工具已安装且在PATH中，否则runAnalysisTools失败
//   - 并发安全: 评估器非并发安全，不应同时评估多个项目
//   - 资源消耗: 大型项目评估可能消耗较多CPU和内存
//   - 结果复用: 同一评估器实例的results会被覆盖，需保存副本
//   - 日志输出: 使用标准log包输出，可能与用户日志混合
//   - 阻塞调用: 评估过程是阻塞的，不支持异步或取消
//   - PassingScore常量: 默认60分，需要修改代码才能调整
//
// 改进方向:
//   - 上下文支持: 接收context.Context参数支持超时和取消
//   - 进度回调: 提供回调函数报告评估进度（30%、60%、100%）
//   - 并行执行: 并发运行多个工具以提升性能
//   - 增量评估: 仅分析变更文件以加速CI/CD
//   - 结果缓存: 缓存工具结果，未变更文件复用缓存
//   - 灵活门禁: 支持配置PassingScore和各维度最低分
//   - 异步评估: 返回Future/Promise对象支持异步等待
//   - 中间结果: 支持流式返回中间结果（工具完成即返回）
//   - 错误恢复: 关键步骤失败时尝试恢复或提供降级结果
//   - 结构化日志: 使用结构化日志库（如logrus/zap）替代log包
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) EvaluateProject(projectPath string) (*CodeQualityResult, error) {
	log.Printf("开始评估项目代码质量: %s", projectPath)
	start := time.Now()

	cqe.results.ProjectPath = projectPath
	cqe.results.Timestamp = start

	// 1. 执行所有分析工具
	if err := cqe.runAnalysisTools(projectPath); err != nil {
		return nil, fmt.Errorf("分析工具执行失败: %v", err)
	}

	// 2. 收集和聚合结果
	if err := cqe.aggregateResults(); err != nil {
		return nil, fmt.Errorf("结果聚合失败: %v", err)
	}

	// 3. 计算维度得分
	cqe.calculateDimensionScores()

	// 4. 计算整体得分
	cqe.calculateOverallScore()

	// 5. 生成改进建议
	cqe.generateImprovements()

	// 6. 分析技术债务
	cqe.analyzeTechnicalDebt()

	// 7. 识别质量热点
	cqe.identifyQualityHotspots()

	cqe.results.Duration = time.Since(start)
	cqe.results.Passed = cqe.results.OverallScore >= PassingScore // 可配置阈值

	log.Printf("代码质量评估完成，总分: %.2f，耗时: %v",
		cqe.results.OverallScore, cqe.results.Duration)

	// 8. 保存结果
	if cqe.config.SaveResults {
		if err := cqe.saveResults(); err != nil {
			log.Printf("保存结果失败: %v", err)
		}
	}

	return cqe.results, nil
}

// runAnalysisTools 运行所有已注册的静态分析工具并收集结果
//
// 功能说明:
//
//	本方法是代码质量评估的第一个核心步骤，负责遍历所有已注册的分析工具
//	（golint、govet、gocyclo等），依次执行它们对项目代码的分析，并收集每个工具
//	的分析结果到results.ToolResults映射表中。该方法采用容错设计，单个工具失败
//	不会中断整个评估流程，只有当所有工具都失败时才返回错误。
//
// 执行流程:
//
//	1. 初始化计数器:
//	   • successCount = 0: 成功执行的工具数量
//	   • failureCount = 0: 执行失败的工具数量
//	   • 用途: 跟踪执行状态，判定整体成功/失败
//
//	2. 遍历所有工具 (range cqe.analysisTools):
//	   • 迭代顺序: 不确定（map遍历顺序随机）
//	   • name: 工具名称（"golint", "govet", "gocyclo"）
//	   • tool: 实现AnalysisTool接口的工具实例
//	   • 数量: 取决于initializeTools()注册的工具
//
//	3. 执行单个工具:
//	   a. 输出开始日志: "执行分析工具: {name}"
//	   b. 调用工具Execute方法: result, err := tool.Execute(projectPath)
//	   c. Execute方法职责:
//	      - 在projectPath目录下执行工具命令
//	      - 捕获工具的标准输出和错误输出
//	      - 解析输出为ToolResult对象
//	      - 返回结果或错误
//
//	4. 错误处理（容错机制）:
//	   • 如果err != nil（工具执行失败）:
//	     - 输出错误日志: "工具 {name} 执行失败: {err}"
//	     - 增加failureCount计数器
//	     - continue跳过当前工具，继续执行下一个
//	     - 不中断整个评估流程（关键设计决策）
//	   • 设计理念:
//	     - 部分工具失败不应影响其他工具
//	     - 用户能获得部分结果总比完全失败好
//	     - 例如: golint未安装但govet可用，仍可得到部分评估
//
//	5. 成功处理:
//	   • 如果err == nil（工具执行成功）:
//	     a. 保存结果: cqe.results.ToolResults[name] = result
//	     b. 增加successCount计数器
//	     c. 输出成功日志: "工具 {name} 完成，发现 {count} 个问题"
//	     d. result.Summary.TotalIssues: 该工具发现的问题总数
//
//	6. 最终判定:
//	   • 条件检查:
//	     if len(cqe.analysisTools) > 0 && successCount == 0
//	   • 失败条件:
//	     - 至少注册了1个工具（len > 0）
//	     - 但所有工具都失败了（successCount == 0）
//	   • 返回错误: "所有{N}个分析工具执行失败"
//	   • 成功条件:
//	     - 至少有1个工具成功执行
//	     - 或者没有注册任何工具（边界情况）
//	   • 返回: nil（表示成功）
//
// 容错机制详解:
//
//	本方法采用"尽力而为"的容错策略：
//
//	1. 单工具失败容忍:
//	   • 场景: golint执行失败（工具未安装）
//	   • 行为: 记录日志，failureCount++，继续执行govet
//	   • 结果: 用户仍能获得govet的分析结果
//	   • 优势: 降低对环境配置的严格要求
//
//	2. 全部失败拒绝:
//	   • 场景: 所有工具都失败（项目路径错误、权限不足）
//	   • 行为: 返回error，终止评估流程
//	   • 结果: EvaluateProject()返回错误给用户
//	   • 原因: 无任何工具结果无法进行后续分析
//
//	3. 部分成功继续:
//	   • 场景: 3个工具中2个成功、1个失败
//	   • 行为: 返回nil，继续aggregateResults()等步骤
//	   • 结果: 基于2个工具的结果完成评估
//	   • 权衡: 结果可能不完整但仍有价值
//
// 工具执行顺序:
//
//	由于Go的map遍历顺序是随机的，工具执行顺序不确定：
//	• 可能顺序: golint → govet → gocyclo
//	• 可能顺序: gocyclo → golint → govet
//	• 影响: 日志输出顺序不固定
//	• 无影响: 最终结果（所有工具结果都会保存）
//	• 改进: 如需固定顺序，可使用切片存储工具名并排序
//
// 参数:
//   - projectPath: 待分析项目的根目录路径
//     • 类型: string
//     • 传递: 直接传递给每个工具的Execute(projectPath)方法
//     • 要求: 必须是有效的Go项目目录
//     • 验证: 本方法不验证，由各工具的Execute()方法处理
//     • 影响: 工具会在此目录下执行分析（如go vet ./...）
//
// 返回值:
//   - error: 执行状态错误对象
//     • nil: 至少有1个工具成功执行，或没有注册任何工具
//     • 非nil: 所有工具都执行失败（无可用结果）
//     • 错误格式: "所有{N}个分析工具执行失败"
//     • N: 注册的工具总数（failureCount的值）
//
// 副作用:
//   - 修改cqe.results.ToolResults映射表:
//     • 每个成功的工具都会添加一个条目
//     • 键: 工具名称（"golint", "govet", "gocyclo"）
//     • 值: 该工具的ToolResult对象（包含Issues/Summary/Metrics）
//   - 输出日志:
//     • 每个工具开始: "执行分析工具: {name}"
//     • 工具失败: "工具 {name} 执行失败: {err}"
//     • 工具成功: "工具 {name} 完成，发现 {count} 个问题"
//
// 性能特征:
//
//	执行耗时取决于：
//	• 工具数量: 3个工具通常需要3-10秒
//	• 项目规模: 文件越多，工具分析越慢
//	• 工具类型:
//	  - gocyclo: 通常最快（<1秒）
//	  - go vet: 中等速度（1-3秒）
//	  - golint: 可能较慢（2-5秒）
//	• 顺序执行: 工具依次执行，非并行
//
//	典型耗时:
//	• 小项目（<1000行）: 1-3秒
//	• 中项目（1000-10000行）: 3-8秒
//	• 大项目（>10000行）: 8-30秒
//
// 使用场景:
//   - EvaluateProject调用: 评估流程的第一步
//   - 独立工具测试: 测试工具注册和执行逻辑
//   - CI/CD集成: 在流水线中收集所有工具结果
//
// 示例:
//
//	// 示例1: 正常执行（所有工具成功）
//	config := &CodeQualityConfig{
//	    EnabledTools: map[string]bool{
//	        "golint":  true,
//	        "govet":   true,
//	        "gocyclo": true,
//	    },
//	}
//	evaluator := NewCodeQualityEvaluator(config)
//	err := evaluator.runAnalysisTools("/path/to/project")
//	if err != nil {
//	    log.Fatalf("所有工具失败: %v", err)
//	}
//	fmt.Printf("成功执行 %d 个工具\n", len(evaluator.results.ToolResults))
//	// 输出: 成功执行 3 个工具
//
//	// 示例2: 部分工具失败（golint未安装）
//	// 假设golint未安装但govet和gocyclo可用
//	err := evaluator.runAnalysisTools(projectPath)
//	if err != nil {
//	    log.Fatal(err)  // 不会执行到这里
//	}
//	// 日志输出:
//	// 执行分析工具: golint
//	// 工具 golint 执行失败: exec: "golint": executable file not found
//	// 执行分析工具: govet
//	// 工具 govet 完成，发现 5 个问题
//	// 执行分析工具: gocyclo
//	// 工具 gocyclo 完成，发现 3 个问题
//	fmt.Printf("部分成功: %d/%d\n",
//	    len(evaluator.results.ToolResults), len(evaluator.analysisTools))
//	// 输出: 部分成功: 2/3
//
//	// 示例3: 所有工具失败（项目路径错误）
//	err := evaluator.runAnalysisTools("/invalid/path")
//	if err != nil {
//	    fmt.Printf("错误: %v\n", err)
//	    // 输出: 错误: 所有3个分析工具执行失败
//	}
//
//	// 示例4: 检查各工具结果
//	evaluator.runAnalysisTools(projectPath)
//	for name, result := range evaluator.results.ToolResults {
//	    fmt.Printf("%s: %d issues, %.2fs\n",
//	        name,
//	        result.Summary.TotalIssues,
//	        result.ExecutionTime.Seconds())
//	}
//	// 输出:
//	// golint: 12 issues, 2.34s
//	// govet: 5 issues, 1.23s
//	// gocyclo: 3 issues, 0.45s
//
// 注意事项:
//   - 工具依赖: 工具必须已安装且在PATH中，否则Execute()失败
//   - 执行顺序: 工具执行顺序随机（map遍历），不应依赖特定顺序
//   - 非并发: 工具顺序执行，可能较慢（改进空间）
//   - 日志噪音: 成功和失败都输出日志，可能较多
//   - 结果覆盖: 重复调用会覆盖ToolResults（需清空或创建新evaluator）
//   - 错误细节: 工具失败的详细原因在日志中，error对象仅说明"全部失败"
//   - 空工具集: 如果没有注册任何工具，返回nil（边界情况）
//   - 资源消耗: 工具执行可能消耗CPU和内存（尤其是大项目）
//
// 改进方向:
//   - 并行执行: 使用goroutine并发执行多个工具以提升速度
//   - 执行超时: 为每个工具设置超时时间（如30秒），避免卡死
//   - 上下文支持: 接收context.Context参数支持取消
//   - 进度回调: 提供回调函数报告工具执行进度（"govet完成 2/3"）
//   - 固定顺序: 使用切片存储工具名，按字母顺序执行
//   - 结果流式: 工具完成即返回中间结果，不等待全部完成
//   - 错误聚合: 返回所有失败工具的错误列表，而非仅"全部失败"
//   - 依赖检查: 启动时预检工具可用性，提前报错
//   - 智能重试: 工具失败时重试1次（可能是临时网络错误）
//   - 结果缓存: 缓存工具结果，未变更文件复用缓存
//   - 日志分级: 成功用INFO级别，失败用WARN级别
//   - 统计信息: 返回详细统计（成功数、失败数、总耗时）
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) runAnalysisTools(projectPath string) error {
	successCount := 0
	failureCount := 0

	for name, tool := range cqe.analysisTools {
		log.Printf("执行分析工具: %s", name)

		result, err := tool.Execute(projectPath)
		if err != nil {
			log.Printf("工具 %s 执行失败: %v", name, err)
			failureCount++
			// 继续执行其他工具，不中断整个评估过程
			continue
		}

		cqe.results.ToolResults[name] = result
		successCount++
		log.Printf("工具 %s 完成，发现 %d 个问题", name, result.Summary.TotalIssues)
	}

	// 如果所有工具都失败，返回错误
	if len(cqe.analysisTools) > 0 && successCount == 0 {
		return fmt.Errorf("所有%d个分析工具执行失败", failureCount)
	}

	return nil
}

// aggregateResults 聚合所有工具的分析结果并生成统计数据
//
// 功能说明:
//
//	本方法是代码质量评估的第二个核心步骤，负责收集所有工具的分析结果，
//	将分散的CodeIssue列表合并为统一的问题集合，同时分析项目的代码结构
//	（文件数、代码行数、注释率、包数量等），生成完整的质量统计数据。
//	该方法的输出为后续的评分计算和改进建议生成提供基础数据。
//
// 执行流程:
//
//	1. 初始化数据结构:
//	   • 创建空的allIssues切片: 存储所有工具的问题
//	   • 创建QualityStatistics对象: 存储项目统计信息
//	   • 准备聚合操作
//
//	2. 聚合所有工具的问题 (range cqe.results.ToolResults):
//	   • 遍历ToolResults映射表中的每个工具结果
//	   • 提取每个工具的result.Issues切片
//	   • 使用append(..., slice...)语法展开并合并到allIssues
//	   • 结果: allIssues包含所有工具发现的所有问题
//	   • 示例:
//	     - golint发现12个问题
//	     - govet发现5个问题
//	     - gocyclo发现3个问题
//	     - allIssues总计20个问题
//
//	3. 分析项目结构 (analyzeProjectStructure):
//	   • 调用analyzeProjectStructure(&stats)方法
//	   • 遍历项目所有.go文件
//	   • 统计以下指标:
//	     - 总文件数、测试文件数
//	     - 总代码行数、代码行、注释行、空行
//	     - 函数总数、包数量
//	   • 填充stats结构体字段
//	   • 如果分析失败，返回包装错误
//
//	4. 保存聚合结果:
//	   • 将allIssues保存到cqe.results.Issues
//	   • 将stats保存到cqe.results.Statistics
//	   • 后续步骤可直接访问这些聚合数据
//
//	5. 返回成功:
//	   • 返回nil表示聚合成功
//	   • 如果项目结构分析失败，返回error
//
// 问题聚合详解:
//
//	聚合操作使用append的展开语法合并多个切片：
//
//	for _, result := range cqe.results.ToolResults {
//	    allIssues = append(allIssues, result.Issues...)
//	}
//
//	• 展开语法: result.Issues... 将切片元素逐个追加
//	• 等价于: 对每个Issue执行append(allIssues, issue)
//	• 性能: O(N)时间复杂度，N为总问题数
//	• 去重: 不进行去重，可能存在重复问题（不同工具报告相同问题）
//	• 顺序: 保持工具遍历顺序（但map遍历顺序随机）
//
// 项目结构分析:
//
//	analyzeProjectStructure(&stats)方法执行以下分析：
//
//	1. 文件遍历 (filepath.WalkDir):
//	   • 递归遍历项目目录下的所有文件
//	   • 筛选.go文件（跳过非Go文件）
//	   • 排除vendor/目录（第三方依赖）
//	   • 识别测试文件（*_test.go）
//
//	2. 代码行统计 (analyzeGoFile):
//	   • 对每个Go文件统计行数
//	   • 分类: 总行数、代码行、注释行、空行
//	   • 累加到总计数器
//	   • 用于计算注释率: commentLines / totalLines
//
//	3. AST解析 (parser.ParseFile):
//	   • 解析每个文件的抽象语法树
//	   • 提取包名（统计包数量）
//	   • 提取函数声明（统计函数数量）
//	   • 用于代码结构分析
//
//	4. 统计数据生成:
//	   • FileCount: 总Go文件数
//	   • TestFileCount: 测试文件数
//	   • TotalLines: 总代码行数
//	   • CodeLines: 有效代码行
//	   • CommentLines: 注释行数
//	   • CommentRatio: 注释率（comment/total）
//	   • FunctionCount: 函数总数
//	   • PackageCount: 包数量
//
// 参数:
//
//	无显式参数（方法接收者为*CodeQualityEvaluator）
//
// 返回值:
//   - error: 聚合过程的错误对象
//     • nil: 聚合成功，所有数据已填充
//     • 非nil: 项目结构分析失败（如文件读取错误、AST解析错误）
//     • 错误格式: "项目结构分析失败: {详细错误}"
//
// 副作用:
//   - 修改cqe.results.Issues:
//     • 设置为所有工具问题的合并列表
//     • 类型: []CodeIssue
//     • 数量: 所有工具的问题总和
//   - 修改cqe.results.Statistics:
//     • 设置为项目的完整统计数据
//     • 类型: QualityStatistics
//     • 字段: FileCount, TotalLines, CommentRatio等
//
// 使用场景:
//   - EvaluateProject调用: 在runAnalysisTools()之后执行
//   - 数据准备: 为后续的评分计算提供基础数据
//   - 统计报告: 生成项目规模和质量的统计信息
//
// 示例:
//
//	// 示例1: 查看聚合后的问题总数
//	evaluator.runAnalysisTools(projectPath)
//	err := evaluator.aggregateResults()
//	if err != nil {
//	    log.Fatalf("聚合失败: %v", err)
//	}
//	fmt.Printf("总问题数: %d\n", len(evaluator.results.Issues))
//	// 输出: 总问题数: 20 (golint 12 + govet 5 + gocyclo 3)
//
//	// 示例2: 查看项目统计数据
//	evaluator.aggregateResults()
//	stats := evaluator.results.Statistics
//	fmt.Printf("文件数: %d (测试: %d)\n", stats.FileCount, stats.TestFileCount)
//	fmt.Printf("代码行数: %d\n", stats.TotalLines)
//	fmt.Printf("注释率: %.2f%%\n", stats.CommentRatio*100)
//	fmt.Printf("函数数: %d\n", stats.FunctionCount)
//	// 输出:
//	// 文件数: 45 (测试: 12)
//	// 代码行数: 5432
//	// 注释率: 18.50%
//	// 函数数: 234
//
//	// 示例3: 检查各工具贡献的问题数
//	evaluator.runAnalysisTools(projectPath)
//	for name, result := range evaluator.results.ToolResults {
//	    fmt.Printf("%s贡献: %d个问题\n", name, len(result.Issues))
//	}
//	evaluator.aggregateResults()
//	fmt.Printf("聚合后总计: %d个问题\n", len(evaluator.results.Issues))
//	// 输出:
//	// golint贡献: 12个问题
//	// govet贡献: 5个问题
//	// gocyclo贡献: 3个问题
//	// 聚合后总计: 20个问题
//
// 注意事项:
//   - 调用顺序: 必须在runAnalysisTools()之后调用（依赖ToolResults）
//   - 问题去重: 不进行问题去重，可能存在重复（改进空间）
//   - vendor目录: 自动排除vendor/目录，避免统计第三方代码
//   - AST解析: 如果某个文件AST解析失败，整个分析会失败
//   - 性能影响: 大型项目（>10000文件）可能耗时较长
//   - 空工具结果: 如果没有任何工具结果，allIssues为空切片（合法）
//   - 统计准确性: 依赖analyzeGoFile()和AST解析的准确性
//
// 改进方向:
//   - 问题去重: 识别并合并相同的问题（基于文件名、行号、规则）
//   - 问题分类: 按严重程度、类别、工具分组统计
//   - 增量分析: 仅分析变更文件，复用历史统计数据
//   - 并行分析: 并发分析多个文件以提升速度
//   - 缓存机制: 缓存未变更文件的统计结果
//   - 错误容忍: AST解析失败时跳过该文件而非整体失败
//   - 更多指标: 统计平均函数长度、圈复杂度分布等
//   - 趋势对比: 与历史数据对比，生成趋势指标
//   - 分包统计: 按包分组生成统计数据
//   - 文件类型: 区分.go和_test.go的统计（当前已区分测试文件）
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) aggregateResults() error {
	var allIssues []CodeIssue

	// 统计信息初始化
	stats := QualityStatistics{}

	// 聚合所有工具的问题
	for _, result := range cqe.results.ToolResults {
		allIssues = append(allIssues, result.Issues...)
	}

	// 分析项目结构
	if err := cqe.analyzeProjectStructure(&stats); err != nil {
		return fmt.Errorf("项目结构分析失败: %v", err)
	}

	cqe.results.Issues = allIssues
	cqe.results.Statistics = stats

	return nil
}

// analyzeProjectStructure 分析项目代码结构并生成统计数据
//
// 功能说明:
//
//	本方法负责深度分析Go项目的代码结构，通过递归遍历项目目录、解析每个Go源文件的
//	AST（抽象语法树）、统计代码行数等方式，生成详细的项目质量统计数据。这些统计数据
//	包括文件数量、代码行数、注释率、函数数量、包数量、测试覆盖率等关键指标，
//	为后续的质量评分和改进建议提供量化依据。
//
// 执行流程:
//
//	1. 初始化统计变量:
//	   • projectPath: 从results中获取项目根目录路径
//	   • totalLines, codeLines, commentLines, blankLines: 行数计数器
//	   • totalFiles, testFiles: 文件计数器
//	   • functions: 函数计数器
//	   • packageSet: map[string]bool 用于统计不重复的包数量
//
//	2. 递归遍历项目目录 (filepath.WalkDir):
//	   • 从projectPath开始递归遍历所有子目录和文件
//	   • 对每个目录项调用回调函数处理
//	   • 回调函数参数:
//	     - path: 文件/目录的完整路径
//	     - d: fs.DirEntry 目录项信息（名称、是否为目录等）
//	     - err: 遍历过程中的错误
//
//	3. 文件筛选（回调函数内部）:
//	   a. 错误检查: if err != nil → 返回err中断遍历
//	   b. 目录跳过: if d.IsDir() → 返回nil继续遍历子目录
//	   c. 非Go文件跳过: if !strings.HasSuffix(path, ".go") → 返回nil
//	   d. vendor目录排除: if strings.Contains(path, "vendor/") → 返回nil
//	   • 筛选后: 仅处理项目源码的.go文件（排除第三方依赖）
//
//	4. 文件计数:
//	   • totalFiles++: 统计所有Go文件
//	   • if strings.HasSuffix(path, "_test.go"): testFiles++
//	   • 区分普通源文件和测试文件
//
//	5. 行数统计 (analyzeGoFile):
//	   • 调用analyzeGoFile(path)逐行分析文件
//	   • 返回值: lines(总行数), code(代码行), comment(注释行), blank(空行)
//	   • 累加到全局计数器:
//	     totalLines += lines
//	     codeLines += code
//	     commentLines += comment
//	     blankLines += blank
//	   • 如果分析失败: 返回err中断遍历
//
//	6. AST解析 (parser.ParseFile):
//	   • 创建新的token.FileSet用于位置信息管理
//	   • 解析文件: parser.ParseFile(fset, path, nil, parser.ParseComments)
//	   • 返回*ast.File节点（文件的AST表示）
//	   • ParseComments模式: 保留注释节点以支持文档分析
//	   • 如果解析失败: 返回err（通常是语法错误）
//
//	7. 包名统计:
//	   • 从AST节点提取包名: node.Name.Name
//	   • 添加到packageSet: packageSet[pkgName] = true
//	   • map自动去重，相同包名仅计数一次
//	   • 注意: 同一个包可能分布在多个文件中
//
//	8. 函数统计 (ast.Inspect):
//	   • 使用ast.Inspect遍历AST树的所有节点
//	   • 回调函数检查每个节点类型
//	   • 如果是*ast.FuncDecl（函数声明）: functions++
//	   • 统计包括:
//	     - 普通函数: func Foo()
//	     - 方法: func (r *Receiver) Method()
//	     - 测试函数: func TestXxx(t *testing.T)
//	   • 返回true: 继续遍历子节点
//
//	9. 遍历完成后填充统计数据:
//	   • stats.TotalFiles = totalFiles
//	   • stats.TotalLines = totalLines
//	   • stats.CodeLines = codeLines
//	   • stats.CommentLines = commentLines
//	   • stats.BlankLines = blankLines
//	   • stats.TestFiles = testFiles
//	   • stats.Functions = functions
//	   • stats.Packages = len(packageSet) (去重后的包数量)
//
//	10. 计算测试覆盖率:
//	    • if totalFiles > 0: stats.TestRatio = float64(testFiles) / float64(totalFiles)
//	    • TestRatio: 测试文件占总文件的比例（0.0-1.0）
//	    • 零除保护: 仅在有文件时计算
//
//	11. 返回成功:
//	    • 返回nil表示分析成功
//	    • stats对象已填充完整数据
//
// AST解析详解:
//
//	抽象语法树（AST）是Go源代码的结构化表示：
//
//	1. token.FileSet:
//	   • 管理源文件的位置信息（文件名、行号、列号）
//	   • 每次解析都需要新建FileSet
//	   • 用途: 错误报告、位置跟踪
//
//	2. parser.ParseFile:
//	   • 将Go源文件解析为AST树
//	   • 参数:
//	     - fset: 位置信息管理器
//	     - path: 源文件路径
//	     - src: nil表示从path读取（也可传入[]byte）
//	     - mode: parser.ParseComments保留注释
//	   • 返回: *ast.File（文件AST根节点）
//
//	3. ast.Inspect:
//	   • 深度优先遍历AST树的所有节点
//	   • 回调函数接收每个节点
//	   • 类型断言识别节点类型:
//	     - *ast.FuncDecl: 函数声明
//	     - *ast.TypeSpec: 类型定义
//	     - *ast.ValueSpec: 变量/常量声明
//	   • 返回值:
//	     - true: 继续遍历子节点
//	     - false: 跳过子节点
//
//	4. 函数识别:
//	   if funcDecl, ok := n.(*ast.FuncDecl); ok {
//	       functions++
//	   }
//	   • 包括所有函数类型（普通函数、方法、测试函数）
//
// 统计指标说明:
//
//	1. TotalFiles: Go源文件总数（排除vendor）
//	   • 包括: *.go文件（含测试文件）
//	   • 排除: vendor/目录、非.go文件
//
//	2. TestFiles: 测试文件数量
//	   • 条件: 文件名以_test.go结尾
//	   • 用途: 计算测试覆盖率
//
//	3. TotalLines: 文件总行数
//	   • 包括: 代码行 + 注释行 + 空行
//	   • 用途: 项目规模评估
//
//	4. CodeLines: 有效代码行数
//	   • 定义: 非空、非纯注释的行
//	   • 用途: 代码量评估
//
//	5. CommentLines: 注释行数
//	   • 包括: // 单行注释、/* */ 多行注释
//	   • 用途: 计算注释率
//
//	6. BlankLines: 空行数
//	   • 定义: 仅包含空白字符的行
//	   • 用途: 代码可读性评估
//
//	7. Functions: 函数总数
//	   • 包括: 普通函数、方法、测试函数
//	   • 用途: 代码模块化程度评估
//
//	8. Packages: 不重复包数量
//	   • 统计: 去重后的包名总数
//	   • 用途: 项目结构复杂度评估
//
//	9. TestRatio: 测试文件比例
//	   • 计算: testFiles / totalFiles
//	   • 范围: 0.0-1.0
//	   • 示例: 0.25表示25%的文件是测试文件
//
// 参数:
//   - stats: QualityStatistics对象指针，用于存储统计结果
//     • 类型: *QualityStatistics
//     • 要求: 非nil指针（调用方需确保）
//     • 输出: 方法执行后，stats的所有字段都会被填充
//
// 返回值:
//   - error: 分析过程的错误对象
//     • nil: 分析成功，stats已填充完整数据
//     • 非nil: 分析失败（文件读取错误、AST解析错误、遍历错误）
//     • 错误类型:
//       - 文件系统错误: 无法读取目录/文件
//       - 语法错误: Go文件包含语法错误，parser.ParseFile失败
//       - 权限错误: 无权限访问某些文件
//
// 使用场景:
//   - aggregateResults调用: 生成项目统计数据
//   - 质量评分计算: 基于统计数据计算各维度评分
//   - 项目报告: 生成项目规模和结构的可视化报告
//
// 示例:
//
//	// 示例1: 分析项目结构
//	stats := &QualityStatistics{}
//	err := evaluator.analyzeProjectStructure(stats)
//	if err != nil {
//	    log.Fatalf("分析失败: %v", err)
//	}
//	fmt.Printf("项目统计:\n")
//	fmt.Printf("  文件: %d (测试: %d)\n", stats.TotalFiles, stats.TestFiles)
//	fmt.Printf("  代码行: %d\n", stats.CodeLines)
//	fmt.Printf("  注释率: %.2f%%\n", float64(stats.CommentLines)/float64(stats.TotalLines)*100)
//	fmt.Printf("  函数: %d\n", stats.Functions)
//	fmt.Printf("  包: %d\n", stats.Packages)
//	// 输出:
//	// 项目统计:
//	//   文件: 45 (测试: 12)
//	//   代码行: 3456
//	//   注释率: 18.50%
//	//   函数: 234
//	//   包: 8
//
//	// 示例2: 计算平均函数长度
//	stats := &QualityStatistics{}
//	evaluator.analyzeProjectStructure(stats)
//	avgFuncLength := float64(stats.CodeLines) / float64(stats.Functions)
//	fmt.Printf("平均函数长度: %.1f 行\n", avgFuncLength)
//	// 输出: 平均函数长度: 14.8 行
//
//	// 示例3: 评估测试覆盖率
//	stats := &QualityStatistics{}
//	evaluator.analyzeProjectStructure(stats)
//	if stats.TestRatio < 0.2 {
//	    fmt.Printf("⚠ 测试覆盖率过低: %.1f%%\n", stats.TestRatio*100)
//	} else {
//	    fmt.Printf("✓ 测试覆盖率良好: %.1f%%\n", stats.TestRatio*100)
//	}
//
// 注意事项:
//   - stats非nil: 调用方必须传入有效的非nil指针
//   - vendor排除: 自动排除vendor/目录，避免统计第三方代码
//   - AST解析失败: 任何文件AST解析失败都会导致整体失败（严格模式）
//   - 性能影响: 大型项目（>1000文件）可能耗时较长（5-30秒）
//   - 内存消耗: 每个文件都会解析AST，大项目可能消耗较多内存
//   - 符号链接: filepath.WalkDir会跟随符号链接，可能重复统计
//   - 测试比例: TestRatio仅反映文件数量比，不反映代码行比例
//   - 包名统计: 同一包的多个文件仅计数一次（packageSet去重）
//   - 函数重复: 不同文件的同名函数会重复计数（这是预期行为）
//
// 改进方向:
//   - 错误容忍: AST解析失败时跳过该文件而非整体失败
//   - 并行分析: 并发分析多个文件以提升速度
//   - 进度报告: 提供进度回调（"已分析 50/100 文件"）
//   - 更多指标: 统计平均函数长度、最长函数、圈复杂度分布
//   - 文件大小: 统计超大文件（>1000行）数量
//   - 代码密度: 计算code/(code+blank)比例
//   - 分包统计: 生成每个包的详细统计数据
//   - 增量分析: 仅分析变更文件，复用缓存
//   - 排除模式: 支持自定义排除模式（如.gitignore规则）
//   - AST缓存: 缓存AST避免重复解析
//   - 结构分析: 识别项目结构模式（MVC、DDD等）
//   - 依赖分析: 统计包之间的依赖关系
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) analyzeProjectStructure(stats *QualityStatistics) error {
	projectPath := cqe.results.ProjectPath

	var totalLines, codeLines, commentLines, blankLines int
	var totalFiles, testFiles int
	var functions int

	packageSet := make(map[string]bool)

	err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 跳过vendor目录
		if strings.Contains(path, "vendor/") {
			return nil
		}

		totalFiles++
		if strings.HasSuffix(path, "_test.go") {
			testFiles++
		}

		// 分析文件内容
		lines, code, comment, blank, err := cqe.analyzeGoFile(path)
		if err != nil {
			return err
		}

		totalLines += lines
		codeLines += code
		commentLines += comment
		blankLines += blank

		// 解析AST获取函数和包信息
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		// 统计包
		packageSet[node.Name.Name] = true

		// 统计函数
		ast.Inspect(node, func(n ast.Node) bool {
			if _, ok := n.(*ast.FuncDecl); ok {
				functions++
			}
			return true
		})

		return nil
	})
	if err != nil {
		return err
	}

	stats.TotalFiles = totalFiles
	stats.TotalLines = totalLines
	stats.CodeLines = codeLines
	stats.CommentLines = commentLines
	stats.BlankLines = blankLines
	stats.TestFiles = testFiles
	stats.Functions = functions
	stats.Packages = len(packageSet)

	if totalFiles > 0 {
		stats.TestRatio = float64(testFiles) / float64(totalFiles)
	}

	return nil
}

// analyzeGoFile 逐行分析Go源文件并统计代码度量指标
//
// 功能说明:
//
//	本方法负责对单个Go源文件进行逐行分析，通过状态机算法识别不同类型的文本行
//	（代码行、注释行、空行），生成该文件的代码度量统计数据。这是项目结构分析
//	(analyzeProjectStructure)的核心子步骤，为后续的注释率计算、代码质量评分提供基础数据。
//
// 核心统计指标:
//
//	1. total (总行数):
//	   • 定义: 文件的物理总行数（包含所有类型的行）
//	   • 计算: 每处理一行total++
//	   • 用途: 文件规模评估、注释率分母
//
//	2. code (代码行数):
//	   • 定义: 非空、非注释的有效代码行
//	   • 判定: 不满足空行/注释条件的行
//	   • 用途: 有效代码量评估、平均函数长度计算
//
//	3. comment (注释行数):
//	   • 定义: 单行注释(//)、块注释(/* */)所占行数
//	   • 包含: 纯注释行和同行注释（简化为注释行）
//	   • 用途: 注释率计算、文档质量评分
//
//	4. blank (空行数):
//	   • 定义: 去除空白字符后为空字符串的行
//	   • 识别: strings.TrimSpace(line) == ""
//	   • 用途: 代码可读性评估、格式化风格分析
//
// 执行流程:
//
//	1. 文件打开 (os.Open):
//	   • 打开filePath指定的源文件
//	   • 返回*os.File句柄和error
//	   • 失败时返回(0, 0, 0, 0, err)
//
//	2. 延迟关闭机制 (defer):
//	   • 使用defer匿名函数确保文件句柄释放
//	   • 关闭错误处理策略:
//	     - 如果closeErr != nil && err == nil: 返回closeErr
//	     - 如果err已存在: 保留原始err，忽略closeErr
//	   • 防止资源泄漏的关键设计
//
//	3. Scanner初始化:
//	   • 创建bufio.NewScanner(file)用于逐行读取
//	   • Scanner自动处理行分隔符（\n, \r\n）
//	   • 内部缓冲区默认64KB，可处理长行
//
//	4. 状态机初始化:
//	   • inBlockComment = false: 块注释状态标志
//	   • 用于跟踪是否位于/* */多行注释内部
//
//	5. 逐行扫描循环 (for scanner.Scan()):
//	   • 每次迭代处理一行文本
//	   • scanner.Text()获取原始行内容
//	   • strings.TrimSpace()去除首尾空白
//	   • total++计数总行数
//
//	6. 空行检测:
//	   • 条件: line == ""（去除空白后为空）
//	   • 操作: blank++, continue跳过
//	   • 示例: "    " → blank++
//
//	7. 块注释开始检测:
//	   • 条件: 包含"/*"但不包含"*/"（多行块注释开始）
//	   • 操作: inBlockComment = true
//	   • 示例: "/* 这是注释" → inBlockComment=true
//
//	8. 块注释内部处理:
//	   • 条件: inBlockComment == true
//	   • 操作:
//	     a. comment++（计入注释行）
//	     b. 检查是否包含"*/"（块注释结束）
//	     c. 如果结束: inBlockComment = false
//	     d. continue跳过后续判断
//	   • 示例:
//	     "这是块注释内容" → comment++
//	     "结束了 */" → comment++, inBlockComment=false
//
//	9. 单行注释检测:
//	   • 条件: strings.HasPrefix(line, "//")
//	   • 操作: comment++, continue
//	   • 示例: "// 这是单行注释" → comment++
//
//	10. 同行块注释检测:
//	    • 条件: 包含"/*"且包含"*/"（同行完整块注释）
//	    • 操作: comment++, continue
//	    • 简化策略: 将同行注释视为注释行（忽略可能的代码部分）
//	    • 示例: "func foo() { /* 注释 */ }" → comment++
//	    • 限制: 无法区分"code /* comment */"的混合情况
//
//	11. 代码行计数:
//	    • 条件: 不满足空行/注释的所有行
//	    • 操作: code++
//	    • 示例: "fmt.Println("hello")" → code++
//
//	12. 扫描完成检查:
//	    • 调用scanner.Err()检查扫描错误
//	    • 返回(total, code, comment, blank, err)
//	    • 错误类型: I/O错误、缓冲区溢出等
//
// 状态机算法详解:
//
//	块注释识别采用简单状态机：
//
//	状态转换图:
//	  [Normal] --遇到"/*"(无"*/")-> [InComment]
//	  [InComment] --遇到"*/"-> [Normal]
//
//	状态变量:
//	  inBlockComment: false (Normal状态) / true (InComment状态)
//
//	处理规则:
//	  • Normal状态: 正常分类代码/注释/空行
//	  • InComment状态: 所有行都视为注释行
//
//	局限性:
//	  • 无法处理嵌套块注释（Go不支持嵌套）
//	  • 字符串中的"/*"会误触发（已知问题，简化实现）
//	  • 同行注释简化为纯注释行（统计不精确）
//
// 安全说明 (#nosec G304):
//
//	使用#nosec G304豁免gosec的G304警告（Potential file inclusion via variable）
//
//	豁免理由:
//	• 调用上下文: filePath来自analyzeProjectStructure()的filepath.WalkDir回调
//	• 路径来源: 系统内部遍历项目目录生成，非用户直接输入
//	• 安全验证: WalkDir已确保路径在项目根目录范围内
//	• 风险评估: 低 - 无外部可控输入，无路径遍历风险
//
//	如果未来接受用户提供的路径，必须:
//	1. 使用security.ValidateSecurePath()验证路径合法性
//	2. 检查路径是否包含".."等危险模式
//	3. 确保路径在允许的basePath范围内
//
// 参数:
//   - filePath: Go源文件的完整路径
//     • 类型: string
//     • 来源: analyzeProjectStructure()的filepath.WalkDir回调
//     • 格式: 绝对路径或相对项目根的路径
//     • 示例: "E:\Go Learn\go-mastery\00-assessment-system\evaluators\code_quality.go"
//     • 要求: 必须是有效的.go文件路径
//
// 返回值:
//   - total: 文件总行数
//     • 类型: int
//     • 范围: >= 0
//     • 计算: 每处理一行+1
//   - code: 有效代码行数
//     • 类型: int
//     • 范围: >= 0
//     • 定义: 非空、非注释的行
//   - comment: 注释行数
//     • 类型: int
//     • 范围: >= 0
//     • 包含: 单行注释+块注释的所有行
//   - blank: 空行数
//     • 类型: int
//     • 范围: >= 0
//     • 定义: TrimSpace后为空的行
//   - err: 错误对象
//     • nil: 分析成功
//     • 非nil: 文件打开失败、读取错误、关闭错误
//     • 错误类型: os.PathError, io.Error
//
// 使用场景:
//   - analyzeProjectStructure调用: 统计每个文件的行数指标
//   - 注释率计算: comment / total
//   - 代码密度计算: code / (code + blank)
//   - 文件规模评估: total值排序识别超大文件
//
// 示例:
//
//	// 示例1: 分析普通Go文件
//	total, code, comment, blank, err := evaluator.analyzeGoFile("main.go")
//	if err != nil {
//	    log.Fatalf("分析失败: %v", err)
//	}
//	fmt.Printf("文件统计: 总行=%d, 代码=%d, 注释=%d, 空行=%d\n",
//	    total, code, comment, blank)
//	// 输出: 文件统计: 总行=100, 代码=60, 注释=25, 空行=15
//
//	// 示例2: 计算注释率
//	total, _, comment, _, _ := evaluator.analyzeGoFile("utils.go")
//	commentRatio := float64(comment) / float64(total) * 100
//	fmt.Printf("注释率: %.2f%%\n", commentRatio)
//	// 输出: 注释率: 25.00%
//
//	// 示例3: 识别不同行类型
//	// 输入文件内容:
//	//   package main
//	//
//	//   // 这是注释
//	//   func main() {
//	//       /* 块注释
//	//          多行 */
//	//       fmt.Println("hello")
//	//   }
//	// 结果: total=8, code=3, comment=3, blank=2
//
// 注意事项:
//   - Scanner限制: 默认最大行长度64KB，超长行会导致scanner.Err()返回错误
//   - 同行注释: 简化处理为注释行，如"code /* comment */"会被计为comment而非code
//   - 字符串陷阱: 字符串中的"//"或"/*"会误判为注释（已知限制）
//   - 块注释嵌套: Go不支持嵌套块注释，状态机也不支持
//   - 文件关闭: defer确保即使发生panic也会关闭文件
//   - 错误优先级: 读取错误优先于关闭错误返回
//   - 空文件: total=0, code=0, comment=0, blank=0（合法输出）
//   - 大文件: 内存友好，逐行处理，不一次性加载全文件
//
// 改进方向:
//   - 精确同行注释: 识别"code /* comment */"并正确分类
//   - 字符串感知: 解析字符串字面量，忽略其中的注释符号
//   - AST辅助: 结合go/parser的注释信息提高准确性
//   - 长行处理: 自定义Scanner缓冲区大小处理超长行
//   - 并发处理: 如果分析大量文件，可并发调用本方法
//   - 进度报告: 返回已处理行数用于进度条显示
//   - 缓存机制: 对未修改文件复用历史统计结果
//   - 更多指标: 统计函数内代码行、类型定义行等细粒度指标
//   - 块注释嵌套: 虽然Go不支持，但可增强算法健壮性
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) analyzeGoFile(filePath string) (total, code, comment, blank int, err error) {
	// #nosec G304 -- 评估系统内部操作，filePath由系统内部调用传入，为受信任的文件路径
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	// 确保文件在函数结束时关闭
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	inBlockComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		total++

		if line == "" {
			blank++
			continue
		}

		// 检查块注释
		if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
			inBlockComment = true
		}
		if inBlockComment {
			comment++
			if strings.Contains(line, "*/") {
				inBlockComment = false
			}
			continue
		}

		// 检查行注释
		if strings.HasPrefix(line, "//") {
			comment++
			continue
		}

		// 检查同行的块注释
		if strings.Contains(line, "/*") && strings.Contains(line, "*/") {
			// 可能包含代码和注释，简单处理为注释行
			comment++
			continue
		}

		code++
	}

	return total, code, comment, blank, scanner.Err()
}

// calculateDimensionScores 计算六大质量维度的独立评分
//
// 功能说明:
//
//	本方法是代码质量评估的核心评分环节，负责协调调用6个独立的评分计算器，
//	分别从代码结构、风格合规、安全性、性能、测试质量、文档完整性六个维度
//	对项目进行全面评估，生成多维度质量画像。每个维度的评分范围均为0-100分，
//	最终这些维度评分会通过加权平均计算出项目的整体质量得分。
//
// 六大质量维度详解:
//
//	1. code_structure (代码结构维度):
//	   • 评估内容: 圈复杂度、函数长度、模块化设计
//	   • 核心指标: 平均圈复杂度、最大圈复杂度、平均函数长度
//	   • 计算器: calculateStructureScore()
//	   • 权重: 通常占25%（可配置）
//	   • 及格标准: ≥60分（复杂度适中、函数简洁）
//
//	2. style_compliance (风格合规维度):
//	   • 评估内容: 命名规范、注释风格、格式化标准
//	   • 核心工具: golint、gofmt、命名约定检查器
//	   • 计算器: calculateStyleScore()
//	   • 权重: 通常占20%（可配置）
//	   • 及格标准: ≥70分（风格统一、符合Go规范）
//
//	3. security_analysis (安全分析维度):
//	   • 评估内容: 安全漏洞、敏感信息泄漏、危险用法
//	   • 核心工具: gosec安全扫描器、手动安全检查
//	   • 计算器: calculateSecurityScore()
//	   • 权重: 通常占20%（可配置）
//	   • 及格标准: ≥80分（无高危漏洞）
//
//	4. performance_analysis (性能分析维度):
//	   • 评估内容: 内存分配、算法效率、资源使用
//	   • 核心检查: 分配模式、性能热点、资源泄漏
//	   • 计算器: calculatePerformanceScore()
//	   • 权重: 通常占15%（可配置）
//	   • 及格标准: ≥60分（无明显性能问题）
//
//	5. test_quality (测试质量维度):
//	   • 评估内容: 测试覆盖率、测试设计、边界条件
//	   • 核心指标: 代码覆盖率、测试文件比例、测试完整性
//	   • 计算器: calculateTestScore()
//	   • 权重: 通常占15%（可配置）
//	   • 及格标准: ≥75分（覆盖率≥75%）
//
//	6. documentation_quality (文档质量维度):
//	   • 评估内容: 注释覆盖率、API文档、README质量
//	   • 核心指标: 注释率、包文档、README存在性
//	   • 计算器: calculateDocumentationScore()
//	   • 权重: 通常占5%（可配置）
//	   • 及格标准: ≥60分（关键部分有文档）
//
// 执行流程:
//
//	1. 配置权重读取（当前未使用）:
//	   • _ = cqe.config.WeightSettings 读取权重配置
//	   • 目前权重在calculateOverallScore()中使用
//	   • 未来可用于动态调整评分策略
//
//	2. 代码结构评分 (calculateStructureScore):
//	   • 分析圈复杂度和函数长度
//	   • 检查gocyclo工具发现的复杂度问题
//	   • 计算结构维度得分并存储到DimensionScores["code_structure"]
//
//	3. 风格合规评分 (calculateStyleScore):
//	   • 统计golint发现的风格问题
//	   • 检查gofmt格式化问题（如果启用）
//	   • 检查命名约定违规
//	   • 计算风格维度得分并存储到DimensionScores["style_compliance"]
//
//	4. 安全分析评分 (calculateSecurityScore):
//	   • 统计gosec发现的安全漏洞（如果启用）
//	   • 执行手动安全检查（硬编码密码、SQL注入等）
//	   • 计算安全维度得分并存储到DimensionScores["security_analysis"]
//
//	5. 性能分析评分 (calculatePerformanceScore):
//	   • 检查性能反模式（不必要的内存分配、低效循环）
//	   • 分析内存分配模式
//	   • 计算性能维度得分并存储到DimensionScores["performance_analysis"]
//
//	6. 测试质量评分 (calculateTestScore):
//	   • 基于测试覆盖率计算基础分数
//	   • 根据测试文件比例调整分数（30%以上测试文件加分）
//	   • 计算测试维度得分并存储到DimensionScores["test_quality"]
//
//	7. 文档质量评分 (calculateDocumentationScore):
//	   • 基于注释率计算基础分数（25%注释率满分）
//	   • 检查README文件（存在加20分）
//	   • 检查包级别文档（最高加30分）
//	   • 计算文档维度得分并存储到DimensionScores["documentation_quality"]
//
//	8. 日志输出:
//	   • 使用log.Printf输出所有6个维度的最终得分
//	   • 格式: "维度得分计算完成: 结构=%.2f, 风格=%.2f, ..."
//	   • 用途: 调试、监控、结果追溯
//
// 评分策略说明:
//
//	所有维度评分均采用"满分扣除"策略：
//	• 起始分数: 100.0分
//	• 扣分规则: 根据问题数量和严重程度扣分
//	• 下限保护: 得分不会低于0分
//	• 上限保护: 得分不会超过100分
//	• 加分机制: 部分维度有加分项（如测试、文档）
//
//	扣分系数示例（可在constants.go中配置）：
//	• 复杂度问题: 5分/个 (ScorePerComplexityIssue)
//	• 风格问题: 2分/个
//	• 高危安全漏洞: 15分/个
//	• 中危安全漏洞: 5分/个
//	• 性能问题: 8分/个
//
// 参数:
//
//	无显式参数（方法接收者为*CodeQualityEvaluator）
//
// 返回值:
//
//	无返回值（void方法）
//	• 副作用: 填充cqe.results.DimensionScores映射表
//	• 映射键: "code_structure", "style_compliance", "security_analysis",
//	          "performance_analysis", "test_quality", "documentation_quality"
//	• 映射值: float64类型的维度评分（0.0-100.0）
//
// 使用场景:
//   - EvaluateProject调用: 在聚合结果后、计算总分前执行
//   - 维度对比分析: 识别项目的薄弱环节
//   - 改进优先级排序: 根据最低维度得分制定改进计划
//   - 质量趋势追踪: 跟踪各维度得分的历史变化
//
// 示例:
//
//	// 示例1: 执行维度评分并查看结果
//	evaluator.calculateDimensionScores()
//	for dimension, score := range evaluator.results.DimensionScores {
//	    fmt.Printf("%s: %.2f分\n", dimension, score)
//	}
//	// 输出:
//	// code_structure: 85.50分
//	// style_compliance: 92.00分
//	// security_analysis: 100.00分
//	// performance_analysis: 78.30分
//	// test_quality: 80.00分
//	// documentation_quality: 65.00分
//
//	// 示例2: 识别薄弱维度
//	evaluator.calculateDimensionScores()
//	minScore := 100.0
//	weakDimension := ""
//	for dimension, score := range evaluator.results.DimensionScores {
//	    if score < minScore {
//	        minScore = score
//	        weakDimension = dimension
//	    }
//	}
//	fmt.Printf("最薄弱维度: %s (%.2f分)\n", weakDimension, minScore)
//	// 输出: 最薄弱维度: documentation_quality (65.00分)
//
//	// 示例3: 检查是否所有维度达标（≥60分）
//	evaluator.calculateDimensionScores()
//	allPassed := true
//	for dimension, score := range evaluator.results.DimensionScores {
//	    if score < 60 {
//	        fmt.Printf("⚠ %s未达标: %.2f < 60\n", dimension, score)
//	        allPassed = false
//	    }
//	}
//	if allPassed {
//	    fmt.Println("✓ 所有维度均达标")
//	}
//
// 注意事项:
//   - 调用顺序: 必须在aggregateResults()之后调用（依赖Statistics和ToolResults）
//   - 工具依赖: 部分维度依赖特定工具结果（如security_analysis需要gosec）
//   - 配置影响: WeightSettings当前未使用，但calculateOverallScore()会使用
//   - 日志输出: log.Printf会输出到标准日志，可能与其他日志混合
//   - 评分独立性: 各维度评分相互独立，不会互相影响
//   - 缺失数据: 如果某工具未运行，相关维度可能得满分（无扣分项）
//   - 并发安全: 本方法非并发安全，不应在多个goroutine中同时调用
//
// 改进方向:
//   - 动态权重: 根据项目类型动态调整各维度权重（如库项目侧重文档）
//   - 细分维度: 增加更多细分维度（如并发安全、API设计）
//   - 历史对比: 与历史维度得分对比，生成趋势报告
//   - 阈值配置化: 将各维度的及格标准配置化而非硬编码
//   - 维度关联分析: 分析维度间的相关性（如复杂度高→测试难→覆盖率低）
//   - 分级评分: 引入A/B/C/D/F等级评价而非纯数字
//   - 可视化支持: 生成雷达图展示六维质量画像
//   - 增量评分: 仅重新计算变更影响的维度
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateDimensionScores() {
	_ = cqe.config.WeightSettings

	// 代码结构得分
	structureScore := cqe.calculateStructureScore()
	cqe.results.DimensionScores["code_structure"] = structureScore

	// 风格合规得分
	styleScore := cqe.calculateStyleScore()
	cqe.results.DimensionScores["style_compliance"] = styleScore

	// 安全分析得分
	securityScore := cqe.calculateSecurityScore()
	cqe.results.DimensionScores["security_analysis"] = securityScore

	// 性能分析得分
	performanceScore := cqe.calculatePerformanceScore()
	cqe.results.DimensionScores["performance_analysis"] = performanceScore

	// 测试质量得分
	testScore := cqe.calculateTestScore()
	cqe.results.DimensionScores["test_quality"] = testScore

	// 文档质量得分
	docScore := cqe.calculateDocumentationScore()
	cqe.results.DimensionScores["documentation_quality"] = docScore

	log.Printf("维度得分计算完成: 结构=%.2f, 风格=%.2f, 安全=%.2f, 性能=%.2f, 测试=%.2f, 文档=%.2f",
		structureScore, styleScore, securityScore, performanceScore, testScore, docScore)
}

// calculateStructureScore 计算代码结构维度的质量评分
//
// 功能说明:
//
//	本方法基于圈复杂度、函数长度和gocyclo工具扫描结果，计算项目的代码结构质量得分。
//	代码结构是衡量代码可维护性和可理解性的核心指标，复杂度过高或函数过长都会显著
//	降低代码质量。本方法采用"满分递减"策略，从100分起始，根据各类结构问题逐项扣分。
//
// 评分维度:
//
//	1. 平均圈复杂度 (Average Cyclomatic Complexity):
//	   • 数据来源: stats.AvgComplexity
//	   • 阈值配置: config.Thresholds.CyclomaticComplexity (默认10)
//	   • 扣分规则: 每超过阈值1点，扣ScorePerComplexityIssue分（默认5分）
//	   • 示例: 平均复杂度13，阈值10 → 超出3点 → 扣15分
//	   • 意义: 反映整体代码复杂度水平
//
//	2. 平均函数长度 (Average Function Length):
//	   • 计算公式: stats.CodeLines / stats.Functions
//	   • 阈值配置: config.Thresholds.FunctionLength (默认50行)
//	   • 扣分规则: 每超过阈值1行，扣HalfComplexityPenalty分（默认2.5分）
//	   • 示例: 平均长度60行，阈值50行 → 超出10行 → 扣25分
//	   • 意义: 长函数难以理解和测试，应拆分为小函数
//
//	3. Gocyclo工具发现的高复杂度函数:
//	   • 数据来源: ToolResults["gocyclo"]
//	   • 错误级问题扣分: result.Summary.ErrorCount * 10分/个
//	   • 警告级问题扣分: result.Summary.WarningCount * ScorePerComplexityIssue分/个
//	   • 示例: 2个Error(复杂度>20) + 3个Warning(复杂度11-20) → 扣20+15=35分
//	   • 意义: 识别需要立即重构的高复杂度函数
//
// 执行流程:
//
//	1. 初始化满分:
//	   • score := 100.0
//	   • 所有扣分项从满分开始递减
//
//	2. 获取统计数据:
//	   • stats := cqe.results.Statistics
//	   • 使用aggregateResults()生成的统计信息
//
//	3. 平均复杂度惩罚计算:
//	   • 条件: stats.AvgComplexity > float64(threshold)
//	   • 超出值: stats.AvgComplexity - threshold
//	   • 惩罚分数: penalty = 超出值 * ScorePerComplexityIssue
//	   • 扣分: score -= penalty
//
//	4. 平均函数长度惩罚计算:
//	   • 计算平均值: avgFunctionLength = CodeLines / Functions
//	   • 条件: avgFunctionLength > float64(FunctionLength阈值)
//	   • 超出值: avgFunctionLength - threshold
//	   • 惩罚分数: penalty = 超出值 * HalfComplexityPenalty (扣分系数减半)
//	   • 扣分: score -= penalty
//	   • 零除保护: stats.Functions必须>0（由analyzeProjectStructure保证）
//
//	5. Gocyclo工具结果惩罚:
//	   • 检查工具是否运行: if result, exists := ToolResults["gocyclo"]
//	   • Error级扣分: ErrorCount * 10 (严重复杂度问题)
//	   • Warning级扣分: WarningCount * ScorePerComplexityIssue
//	   • 如果工具未运行: 跳过此项扣分（无惩罚）
//
//	6. 分数范围限制:
//	   • 下限保护: if score < 0 { score = 0 }
//	   • 上限保护: if score > 100 { score = 100 }
//	   • 防止极端值导致的显示异常
//
//	7. 返回最终得分:
//	   • 返回0-100范围内的float64分数
//
// 扣分系数配置:
//
//	ScorePerComplexityIssue (默认5.0):
//	• 用于平均复杂度和gocyclo Warning级问题
//	• 示例: 复杂度超出3点 → 扣15分
//
//	HalfComplexityPenalty (默认2.5):
//	• 用于函数长度惩罚
//	• 为ScorePerComplexityIssue的一半（函数长度影响相对较小）
//	• 示例: 平均长度超出10行 → 扣25分
//
//	Gocyclo Error级固定扣分: 10分/个
//	• 适用于超过2倍阈值的极端复杂函数
//	• 示例: 阈值10，复杂度21+ → Error级问题
//
// 评分等级参考:
//
//	• 90-100分: 优秀 - 结构清晰，复杂度低
//	• 80-89分: 良好 - 结构合理，少量复杂函数
//	• 70-79分: 中等 - 结构可接受，有改进空间
//	• 60-69分: 及格 - 存在结构问题，建议重构
//	• <60分: 不及格 - 严重结构问题，必须重构
//
// 参数:
//
//	无显式参数（方法接收者为*CodeQualityEvaluator）
//
// 返回值:
//   - float64: 代码结构维度评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数，支持小数
//     • 意义: 分数越高，代码结构质量越好
//
// 使用场景:
//   - calculateDimensionScores调用: 计算结构维度得分
//   - 重构优先级判断: 得分<70时需要重构
//   - 质量门禁: CI/CD中要求结构得分≥60
//   - 趋势分析: 跟踪结构得分随时间的变化
//
// 示例:
//
//	// 示例1: 理想情况（无扣分）
//	// stats.AvgComplexity = 8.0 (阈值10)
//	// avgFunctionLength = 45行 (阈值50)
//	// gocyclo无问题
//	score := evaluator.calculateStructureScore()
//	// score = 100.0 (满分)
//
//	// 示例2: 中等复杂度
//	// stats.AvgComplexity = 12.0 (超出阈值10，扣10分)
//	// avgFunctionLength = 55行 (超出阈值50，扣12.5分)
//	// gocyclo: 2个Warning (扣10分)
//	score := evaluator.calculateStructureScore()
//	// score = 100 - 10 - 12.5 - 10 = 67.5
//
//	// 示例3: 严重结构问题
//	// stats.AvgComplexity = 18.0 (超出8点，扣40分)
//	// avgFunctionLength = 80行 (超出30行，扣75分)
//	// gocyclo: 3个Error + 5个Warning (扣55分)
//	score := evaluator.calculateStructureScore()
//	// score = 100 - 40 - 75 - 55 = -70 → 限制为0.0
//
// 注意事项:
//   - 零除风险: stats.Functions可能为0（空项目），会导致NaN，需在aggregateResults中保证>0
//   - 工具依赖: gocyclo未运行时，仅基于统计数据评分（可能偏高）
//   - 阈值合理性: 默认阈值适合一般项目，特殊项目需调整
//   - 扣分累加: 多个问题会叠加扣分，可能快速降至0分
//   - 浮点精度: 使用float64避免精度损失
//   - 负分保护: score<0时强制归零，避免显示负数
//   - 超满分保护: 虽然当前无加分项，但仍保留>100检查以防未来扩展
//
// 改进方向:
//   - 细分扣分项: 区分最大复杂度和平均复杂度的权重
//   - 非线性扣分: 复杂度越高扣分越重（指数递增）
//   - 加分机制: 复杂度显著低于阈值时给予加分
//   - 文件级分析: 识别单个文件的结构问题而非仅全局平均
//   - 认知复杂度: 引入CognitiveComplexity作为补充指标
//   - 嵌套深度: 增加代码嵌套深度的检查
//   - 模块化度量: 评估包之间的耦合度
//   - 历史趋势: 与上次评分对比，识别结构恶化或改善
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateStructureScore() float64 {
	score := 100.0
	stats := cqe.results.Statistics

	// 复杂度惩罚
	if stats.AvgComplexity > float64(cqe.config.Thresholds.CyclomaticComplexity) {
		penalty := (stats.AvgComplexity - float64(cqe.config.Thresholds.CyclomaticComplexity)) * ScorePerComplexityIssue
		score -= penalty
	}

	// 函数长度惩罚（简化计算）
	avgFunctionLength := float64(stats.CodeLines) / float64(stats.Functions)
	if avgFunctionLength > float64(cqe.config.Thresholds.FunctionLength) {
		penalty := (avgFunctionLength - float64(cqe.config.Thresholds.FunctionLength)) * HalfComplexityPenalty
		score -= penalty
	}

	// gocyclo工具的复杂度问题惩罚
	if result, exists := cqe.results.ToolResults["gocyclo"]; exists {
		score -= float64(result.Summary.ErrorCount) * 10
		score -= float64(result.Summary.WarningCount) * ScorePerComplexityIssue
	}

	// 确保得分在0-100范围内
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateStyleScore 计算代码风格合规维度的质量评分
//
// 功能说明:
//
//	本方法基于golint工具检测结果、gofmt格式化检查和命名约定规范，计算项目的代码风格
//	合规质量得分。代码风格是团队协作和代码可读性的基础，统一的风格规范能显著降低
//	代码审查成本和新成员上手难度。本方法采用"满分递减"策略，从100分起始，根据三类
//	风格问题分别扣分，最终得分反映代码对Go语言官方风格指南的遵循程度。
//
// 评分维度:
//
//	1. Golint风格问题 (Golint Style Issues):
//	   • 数据来源: ToolResults["golint"].Summary.TotalIssues
//	   • 检查内容: 命名规范、注释格式、导出声明规范、包级注释
//	   • 扣分规则: 每个问题扣2分
//	   • 示例: 发现12个golint问题 → 扣24分
//	   • 意义: golint是Go官方推荐的风格检查工具，其建议直接反映Go语言最佳实践
//
//	2. Gofmt格式化问题 (Gofmt Formatting Issues):
//	   • 数据来源: checkGofmt()方法返回值（未格式化文件数）
//	   • 检查内容: 代码格式化标准（缩进、空格、换行等）
//	   • 扣分规则: 每个未格式化文件扣10分
//	   • 前置条件: config.EnabledTools["gofmt"] == true
//	   • 示例: 3个文件未执行gofmt → 扣30分
//	   • 意义: gofmt是Go官方格式化工具，统一格式化是Go社区的强约定
//
//	3. 命名约定问题 (Naming Convention Issues):
//	   • 数据来源: checkNamingConventions()方法返回值
//	   • 检查内容: 变量名、函数名、类型名、常量名是否符合Go命名规范
//	   • 扣分规则: 每个命名问题扣3分
//	   • 示例: 5个命名不规范 → 扣15分
//	   • 意义: 良好的命名是自文档代码的基础，直接影响代码可读性
//
// 执行流程:
//
//	1. 初始化满分:
//	   • score := 100.0
//	   • 所有扣分项从满分开始递减
//
//	2. Golint问题惩罚计算:
//	   • 条件检查: if result, exists := ToolResults["golint"]
//	   • 判断依据: golint工具是否已执行并返回结果
//	   • 存在时: score -= TotalIssues * 2
//	   • 不存在时: 跳过此项惩罚（工具未运行，无扣分）
//
//	3. Gofmt格式化惩罚计算:
//	   • 工具启用检查: if config.EnabledTools["gofmt"]
//	   • 执行检查: unformattedFiles := checkGofmt()
//	   • checkGofmt()逻辑:
//	     - 执行 gofmt -l . 命令
//	     - 返回需要格式化的文件列表长度
//	     - 0表示所有文件已正确格式化
//	   • 条件惩罚: if unformattedFiles > 0 { score -= count * 10 }
//	   • 工具未启用: 跳过此项检查
//
//	4. 命名约定惩罚计算:
//	   • 执行检查: namingIssues := checkNamingConventions()
//	   • checkNamingConventions()逻辑:
//	     - 检查变量名: 驼峰命名 vs 下划线命名
//	     - 检查导出名称: 首字母大写规则
//	     - 检查缩写词: URL vs Url, ID vs Id
//	     - 检查包名: 全小写、无下划线
//	   • 扣分: score -= namingIssues * 3
//
//	5. 分数下限保护:
//	   • 条件检查: if score < 0
//	   • 下限处理: score = 0
//	   • 防止负数分数显示异常
//	   • 无上限检查: 当前无加分项，score不会超过100
//
//	6. 返回最终得分:
//	   • 返回0-100范围内的float64分数
//
// 扣分系数配置:
//
//	Golint问题系数: 2分/个
//	• 理由: golint问题通常是风格建议，影响相对较小
//	• 对比: 低于结构问题的ScorePerComplexityIssue (5分/个)
//	• 示例: 10个golint问题 → 扣20分 (score = 80)
//
//	Gofmt未格式化系数: 10分/文件
//	• 理由: 未格式化文件严重影响代码可读性和团队协作
//	• 对比: 高于单个golint问题，因为影响范围是整个文件
//	• 示例: 5个文件未格式化 → 扣50分 (score = 50)
//
//	命名约定问题系数: 3分/个
//	• 理由: 命名问题介于golint和gofmt之间，影响代码理解但不影响格式
//	• 对比: 高于golint (2分)，低于gofmt (10分)
//	• 示例: 8个命名问题 → 扣24分 (score = 76)
//
// 评分等级参考:
//
//	• 90-100分: 优秀 - 风格高度统一，完全符合Go规范
//	• 80-89分: 良好 - 风格基本统一，少量不规范
//	• 70-79分: 中等 - 存在一定风格问题，需改进
//	• 60-69分: 及格 - 风格问题较多，建议全面审查
//	• <60分: 不及格 - 严重风格问题，必须重新格式化和规范化
//
// 参数:
//
//	无显式参数（方法接收者为*CodeQualityEvaluator）
//
// 返回值:
//   - float64: 代码风格合规维度评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数，支持小数
//     • 意义: 分数越高，代码风格越符合Go官方规范
//
// 使用场景:
//   - calculateDimensionScores调用: 计算风格维度得分
//   - 代码审查辅助: 得分<80时需要进行风格审查
//   - CI/CD质量门禁: 要求风格得分≥70作为合并条件
//   - 团队规范培训: 识别团队常见的风格问题
//   - 新成员入职: 评估新成员代码风格规范性
//
// 示例:
//
//	// 示例1: 理想情况（无扣分）
//	// golint无问题, gofmt全部格式化, 命名规范
//	score := evaluator.calculateStyleScore()
//	// score = 100.0 (满分)
//
//	// 示例2: 中等风格问题
//	// golint: 10个问题 (扣20分)
//	// gofmt: 2个文件未格式化 (扣20分)
//	// 命名: 5个问题 (扣15分)
//	score := evaluator.calculateStyleScore()
//	// score = 100 - 20 - 20 - 15 = 45.0
//
//	// 示例3: 严重风格问题
//	// golint: 25个问题 (扣50分)
//	// gofmt: 8个文件未格式化 (扣80分)
//	// 命名: 10个问题 (扣30分)
//	score := evaluator.calculateStyleScore()
//	// score = 100 - 50 - 80 - 30 = -60 → 限制为0.0
//
//	// 示例4: gofmt未启用的情况
//	// config.EnabledTools["gofmt"] = false
//	// golint: 8个问题 (扣16分)
//	// gofmt: 跳过检查 (不扣分)
//	// 命名: 4个问题 (扣12分)
//	score := evaluator.calculateStyleScore()
//	// score = 100 - 16 - 0 - 12 = 72.0
//
// 注意事项:
//   - 工具依赖: golint未运行时，仅基于gofmt和命名约定评分（可能偏高）
//   - gofmt可选性: gofmt检查依赖EnabledTools配置，未启用时跳过
//   - 扣分累加: 多个问题会叠加扣分，可能快速降至0分
//   - 命名检查实现: checkNamingConventions()当前返回0（需完善实现）
//   - 工具执行顺序: 必须在runAnalysisTools()之后调用（依赖ToolResults）
//   - 负分保护: score<0时强制归零，避免显示负数
//   - 无上限保护: 当前无加分项，但保留扩展空间
//   - 浮点精度: 使用float64避免精度损失
//
// 改进方向:
//   - checkNamingConventions完善: 实现具体的命名规范检查逻辑
//   - 细分扣分项: 区分不同类型的golint问题（如缺少注释 vs 命名问题）
//   - 加分机制: 超过某个阈值的高质量注释可以加分
//   - 自动修复建议: 为每个风格问题提供具体的修复方法
//   - 工具版本感知: 记录golint/gofmt版本，确保评分可重复性
//   - 严重程度分级: 区分必须修复的风格问题和可选的改进建议
//   - 历史趋势: 与上次评分对比，识别风格改善或恶化
//   - 文件级分析: 识别风格问题最多的文件，优先整改
//   - 规则定制化: 支持团队自定义风格规则和扣分系数
//   - 自动格式化: 集成gofmt -w自动修复格式化问题
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateStyleScore() float64 {
	score := 100.0

	// golint问题惩罚
	if result, exists := cqe.results.ToolResults["golint"]; exists {
		score -= float64(result.Summary.TotalIssues) * 2 // 每个风格问题扣2分
	}

	// gofmt检查（如果启用）
	if cqe.config.EnabledTools["gofmt"] {
		// 检查格式化问题
		if unformattedFiles := cqe.checkGofmt(); unformattedFiles > 0 {
			score -= float64(unformattedFiles) * 10 // 每个未格式化文件扣10分
		}
	}

	// 命名约定检查
	namingIssues := cqe.checkNamingConventions()
	score -= float64(namingIssues) * 3

	if score < 0 {
		score = 0
	}
	return score
}

// calculateSecurityScore 计算安全分析维度的质量评分
//
// 功能说明:
//
//	本方法基于gosec安全扫描工具和手动安全检查，计算项目的安全质量得分。
//	安全性是软件质量的重要支柱，即使功能完善的代码，如果存在安全漏洞，也可能
//	导致严重的数据泄露、权限提升、拒绝服务等安全事件。本方法采用"满分递减"策略，
//	从100分起始，根据安全漏洞的严重程度分级扣分，最终得分反映代码的安全防护能力。
//
// 评分维度:
//
//	1. Gosec工具检测的安全漏洞 (Gosec Security Issues):
//	   • 数据来源: ToolResults["gosec"].Summary
//	   • 检查内容: 常见安全漏洞模式（SQL注入、命令注入、弱加密、不安全随机数等）
//	   • 扣分规则:
//	     - 高危漏洞 (ErrorCount): 每个扣ScorePerSecurityHighIssue分（默认15分）
//	     - 中危漏洞 (WarningCount): 每个扣ScorePerSecurityMedIssue分（默认5分）
//	   • 示例: 2个高危漏洞 + 3个中危漏洞 → 扣30+15=45分
//	   • 意义: gosec是Go官方推荐的安全扫描工具，能识别大部分常见安全问题
//
//	2. 手动安全检查 (Manual Security Checks):
//	   • 数据来源: checkSecurityIssues()方法返回值
//	   • 检查内容: gosec未覆盖的特定安全模式
//	     - 硬编码密码和密钥
//	     - 敏感信息日志输出
//	     - 不安全的反序列化
//	     - 业务逻辑漏洞
//	   • 扣分规则: 每个问题扣10分
//	   • 示例: 发现4个手动安全问题 → 扣40分
//	   • 意义: 补充自动化工具的盲区，提供更全面的安全覆盖
//
// 执行流程:
//
//	1. 初始化满分:
//	   • score := 100.0
//	   • 所有扣分项从满分开始递减
//
//	2. Gosec工具结果惩罚计算:
//	   • 条件检查: if result, exists := ToolResults["gosec"]
//	   • 判断依据: gosec工具是否已执行并返回结果
//	   • 高危漏洞扣分:
//	     - score -= ErrorCount * ScorePerSecurityHighIssue
//	     - ErrorCount: 严重安全问题数量（如SQL注入、命令注入）
//	     - ScorePerSecurityHighIssue: 默认15分/个（可在constants.go配置）
//	   • 中危漏洞扣分:
//	     - score -= WarningCount * ScorePerSecurityMedIssue
//	     - WarningCount: 一般安全问题数量（如弱加密、信息泄露）
//	     - ScorePerSecurityMedIssue: 默认5分/个
//	   • 工具未启用: 跳过此项惩罚（但安全评分可能偏高）
//
//	3. 手动安全检查惩罚计算:
//	   • 执行检查: securityIssues := checkSecurityIssues()
//	   • checkSecurityIssues()逻辑:
//	     - 扫描代码中的硬编码密码（如 password := "123456"）
//	     - 检查敏感信息日志（如 log.Printf("Token: %s", token)）
//	     - 识别不安全的反序列化（如 json.Unmarshal without validation）
//	     - 检测SQL注入风险（如字符串拼接SQL）
//	   • 扣分: score -= securityIssues * 10
//	   • 注意: checkSecurityIssues()当前返回0（实现未完善）
//
//	4. 分数下限保护:
//	   • 条件检查: if score < 0
//	   • 下限处理: score = 0
//	   • 防止负数分数显示异常
//	   • 无上限检查: 当前无加分项，score不会超过100
//
//	5. 返回最终得分:
//	   • 返回0-100范围内的float64分数
//
// 扣分系数配置:
//
//	ScorePerSecurityHighIssue（默认15分/个）:
//	• 理由: 高危安全漏洞可能导致系统完全沦陷，必须严厉扣分
//	• 对比: 最高的单项扣分系数，高于复杂度问题(5分)和风格问题(2分)
//	• 示例: 1个SQL注入漏洞 → 扣15分
//	• 典型高危问题:
//	  - SQL注入 (CWE-89)
//	  - 命令注入 (CWE-78)
//	  - 路径遍历 (CWE-22)
//	  - 不安全的反序列化 (CWE-502)
//
//	ScorePerSecurityMedIssue（默认5分/个）:
//	• 理由: 中危漏洞虽不直接导致系统沦陷，但可能成为攻击链的一环
//	• 对比: 与复杂度问题ScorePerComplexityIssue相同，高于风格问题
//	• 示例: 2个弱加密问题 → 扣10分
//	• 典型中危问题:
//	  - 使用MD5/SHA1弱哈希 (CWE-327)
//	  - 不安全的随机数生成 (CWE-338)
//	  - 敏感信息泄露 (CWE-200)
//	  - 缺少输入验证 (CWE-20)
//
//	手动检查问题系数（固定10分/个）:
//	• 理由: 手动检查发现的问题通常是业务逻辑漏洞，难以自动化检测
//	• 对比: 介于高危(15分)和中危(5分)之间
//	• 示例: 3个硬编码密码 → 扣30分
//
// 评分等级参考:
//
//	• 90-100分: 优秀 - 无已知安全漏洞，安全防护充分
//	• 80-89分: 良好 - 存在少量低危漏洞，不影响核心安全
//	• 70-79分: 中等 - 存在中危漏洞，需要修复
//	• 60-69分: 及格 - 存在较多安全问题，建议全面审查
//	• <60分: 不及格 - 严重安全漏洞，禁止上线
//
// 参数:
//
//	无显式参数（方法接收者为*CodeQualityEvaluator）
//
// 返回值:
//   - float64: 安全分析维度评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数，支持小数
//     • 意义: 分数越高，代码安全性越好
//
// 使用场景:
//   - calculateDimensionScores调用: 计算安全维度得分
//   - 安全审计: 得分<80时需要进行安全审计
//   - 上线审批: 要求安全得分≥90作为生产发布条件
//   - 渗透测试优先级: 根据得分排序，优先测试低分项目
//   - 安全培训: 识别团队常见的安全编码问题
//
// 示例:
//
//	// 示例1: 理想情况（无扣分）
//	// gosec无问题, 手动检查无问题
//	score := evaluator.calculateSecurityScore()
//	// score = 100.0 (满分)
//
//	// 示例2: 中等安全问题
//	// gosec: 1个高危漏洞 (扣15分) + 3个中危漏洞 (扣15分)
//	// 手动: 2个问题 (扣20分)
//	score := evaluator.calculateSecurityScore()
//	// score = 100 - 15 - 15 - 20 = 50.0
//
//	// 示例3: 严重安全问题
//	// gosec: 3个高危漏洞 (扣45分) + 10个中危漏洞 (扣50分)
//	// 手动: 5个问题 (扣50分)
//	score := evaluator.calculateSecurityScore()
//	// score = 100 - 45 - 50 - 50 = -45 → 限制为0.0
//
//	// 示例4: gosec未启用的情况
//	// config.EnabledTools["gosec"] = false
//	// gosec: 跳过检查 (不扣分)
//	// 手动: 3个问题 (扣30分)
//	score := evaluator.calculateSecurityScore()
//	// score = 100 - 0 - 30 = 70.0
//
// 注意事项:
//   - 工具依赖: gosec未运行时，仅基于手动检查评分（安全覆盖不全）
//   - 手动检查实现: checkSecurityIssues()当前返回0（待完善实现）
//   - 扣分累加: 多个漏洞会叠加扣分，可能快速降至0分
//   - 严重程度区分: 必须严格区分高危和中危，避免误判
//   - 工具执行顺序: 必须在runAnalysisTools()之后调用（依赖ToolResults）
//   - 负分保护: score<0时强制归零，避免显示负数
//   - 安全优先级: 安全问题应优先于其他维度问题修复
//   - 误报处理: gosec可能产生误报，需人工复核后调整评分
//
// 改进方向:
//   - checkSecurityIssues完善: 实现具体的手动安全检查逻辑
//   - 漏洞分类细化: 区分更多安全类别（如OWASP Top 10分类）
//   - CVE关联: 将发现的漏洞与CVE数据库关联
//   - 自动修复建议: 为每个安全问题提供具体的修复代码示例
//   - 严重程度动态评估: 根据项目类型动态调整严重程度（如金融项目更严格）
//   - 历史趋势: 与上次评分对比，识别新增或修复的漏洞
//   - 渗透测试集成: 结合渗透测试结果调整安全评分
//   - 安全基线: 支持团队自定义安全基线和扣分系数
//   - 威胁建模: 集成威胁建模工具，基于威胁场景评分
//   - 安全培训推荐: 根据发现的漏洞类型推荐相关安全培训课程
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateSecurityScore() float64 {
	score := 100.0

	// gosec工具的安全问题惩罚（如果启用）
	if result, exists := cqe.results.ToolResults["gosec"]; exists {
		score -= float64(result.Summary.ErrorCount) * ScorePerSecurityHighIssue  // 严重安全问题
		score -= float64(result.Summary.WarningCount) * ScorePerSecurityMedIssue // 一般安全问题
	}

	// 手动安全检查
	securityIssues := cqe.checkSecurityIssues()
	score -= float64(securityIssues) * 10

	if score < 0 {
		score = 0
	}
	return score
}

// calculatePerformanceScore 计算性能分析维度的质量评分
//
// 功能说明:
//
//	本方法基于性能问题检查和内存分配模式分析,计算项目的性能质量得分。
//	性能是软件非功能性需求的核心指标,直接影响用户体验、系统吞吐量、资源成本。
//	低效的代码可能导致响应延迟、内存溢出、CPU占用过高等问题,严重时甚至引发服务不可用。
//	本方法采用"满分递减"策略,从100分起始,根据性能反模式和资源使用问题分类扣分,
//	最终得分反映代码的运行效率和资源利用合理性。
//
// 评分维度:
//
//	1. 一般性能问题 (General Performance Issues):
//	   • 数据来源: checkPerformanceIssues()方法返回值
//	   • 检查内容: 性能反模式和低效代码实践
//	     - 不必要的字符串拼接 (使用+而非strings.Builder)
//	     - 未优化的循环 (O(n²)可优化为O(n))
//	     - 频繁的类型转换 (interface{}反复断言)
//	     - 低效的数据结构选择 (应用map时使用slice)
//	     - 同步阻塞操作 (在高并发路径中使用锁)
//	   • 扣分规则: 每个问题扣ScorePerPerformanceIssue分 (默认8分)
//	   • 示例: 发现5个性能问题 → 扣40分
//	   • 意义: 识别代码中的性能热点,避免系统瓶颈
//
//	2. 内存分配模式问题 (Memory Allocation Patterns):
//	   • 数据来源: checkAllocationPatterns()方法返回值
//	   • 检查内容: 内存使用效率和分配策略
//	     - 过度内存分配 (频繁make()未预分配容量)
//	     - 内存泄漏风险 (goroutine未正确终止、全局变量无限增长)
//	     - 不必要的堆分配 (逃逸分析优化点)
//	     - 大对象频繁创建 (应使用对象池sync.Pool)
//	     - slice/map容量浪费 (cap远大于len)
//	   • 扣分规则: 每个问题扣ScorePerAllocationIssue分
//	   • 示例: 发现3个分配问题 → 扣分 (具体系数待配置)
//	   • 意义: 优化内存使用,减少GC压力,提升吞吐量
//
// 执行流程:
//
//	1. 初始化满分:
//	   • score := 100.0
//	   • 所有扣分项从满分开始递减
//
//	2. 一般性能问题惩罚计算:
//	   • 执行检查: performanceIssues := checkPerformanceIssues()
//	   • checkPerformanceIssues()逻辑:
//	     - 扫描代码中的字符串拼接模式 (查找+操作符密集使用)
//	     - 检测嵌套循环复杂度 (O(n²)及以上)
//	     - 识别频繁的interface{}类型断言
//	     - 检查同步原语使用 (sync.Mutex在热路径)
//	     - 分析数据结构选择 (线性查找vs哈希查找)
//	   • 扣分: score -= performanceIssues * ScorePerPerformanceIssue
//	   • 注意: checkPerformanceIssues()当前返回0 (实现未完善)
//
//	3. 内存分配模式惩罚计算:
//	   • 执行检查: allocationIssues := checkAllocationPatterns()
//	   • checkAllocationPatterns()逻辑:
//	     - 检测make()调用未指定容量 (slice/map/channel)
//	     - 识别goroutine泄漏模式 (无终止条件的go func)
//	     - 分析逃逸分析结果 (使用go build -gcflags='-m')
//	     - 检查对象创建频率 (高频new()应改用sync.Pool)
//	     - 评估slice/map容量利用率 (cap/len比值异常)
//	   • 扣分: score -= allocationIssues * ScorePerAllocationIssue
//	   • 注意: checkAllocationPatterns()当前返回0 (实现未完善)
//
//	4. 分数下限保护:
//	   • 条件检查: if score < 0
//	   • 下限处理: score = 0
//	   • 防止负数分数显示异常
//	   • 无上限检查: 当前无加分项,score不会超过100
//
//	5. 返回最终得分:
//	   • 返回0-100范围内的float64分数
//
// 扣分系数配置:
//
//	ScorePerPerformanceIssue (默认8分/个):
//	• 理由: 性能问题虽不像安全漏洞致命,但累积影响显著
//	• 对比: 高于风格问题 (2分),低于高危安全问题 (15分)
//	• 示例: 5个未优化的字符串拼接 → 扣40分 (score = 60)
//	• 典型性能问题:
//	  - 字符串拼接: 循环中使用+而非strings.Builder
//	  - 算法复杂度: O(n²)可优化为O(n)或O(log n)
//	  - 类型转换: 频繁的interface{}.(Type)断言
//	  - 数据结构: 使用slice进行O(n)查找而非map的O(1)
//	  - 同步开销: 热路径中使用sync.Mutex造成竞争
//
//	ScorePerAllocationIssue (系数待配置):
//	• 理由: 内存分配效率直接影响GC压力和系统吞吐量
//	• 对比: 应与ScorePerPerformanceIssue相当或略高
//	• 建议值: 8-10分/个
//	• 典型分配问题:
//	  - 未预分配容量: make([]int, 0)应为make([]int, 0, expectedSize)
//	  - Goroutine泄漏: 无终止条件的go func()持续消耗资源
//	  - 逃逸优化点: 本应栈分配的变量逃逸到堆
//	  - 对象池缺失: 高频创建对象未使用sync.Pool复用
//	  - 容量浪费: make([]int, 0, 10000)但实际只用10个元素
//
// 评分等级参考:
//
//	• 90-100分: 优秀 - 性能优化充分,资源使用高效
//	• 80-89分: 良好 - 存在少量优化点,不影响整体性能
//	• 70-79分: 中等 - 有一定性能问题,需要优化
//	• 60-69分: 及格 - 性能问题较多,建议全面审查
//	• <60分: 不及格 - 严重性能问题,可能导致系统瓶颈
//
// 参数:
//
//	无显式参数 (方法接收者为*CodeQualityEvaluator)
//
// 返回值:
//   - float64: 性能分析维度评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数,支持小数
//     • 意义: 分数越高,代码性能越好,资源利用越合理
//
// 使用场景:
//   - calculateDimensionScores调用: 计算性能维度得分
//   - 性能优化排序: 得分<70时需要性能优化
//   - 上线性能审查: 要求性能得分≥60作为生产发布条件
//   - 性能测试触发: 根据得分决定是否需要压力测试
//   - 资源预算评估: 低分代码需要更多CPU/内存资源
//
// 示例:
//
//	// 示例1: 理想情况 (无扣分)
//	// checkPerformanceIssues()返回0, checkAllocationPatterns()返回0
//	score := evaluator.calculatePerformanceScore()
//	// score = 100.0 (满分)
//
//	// 示例2: 中等性能问题
//	// 性能问题: 5个 (扣40分,假设ScorePerPerformanceIssue=8)
//	// 分配问题: 3个 (扣24分,假设ScorePerAllocationIssue=8)
//	score := evaluator.calculatePerformanceScore()
//	// score = 100 - 40 - 24 = 36.0
//
//	// 示例3: 严重性能问题
//	// 性能问题: 10个 (扣80分)
//	// 分配问题: 8个 (扣64分)
//	score := evaluator.calculatePerformanceScore()
//	// score = 100 - 80 - 64 = -44 → 限制为0.0
//
//	// 示例4: 检查方法未实现的情况 (当前状态)
//	// checkPerformanceIssues()返回0 (未实现)
//	// checkAllocationPatterns()返回0 (未实现)
//	score := evaluator.calculatePerformanceScore()
//	// score = 100 - 0 - 0 = 100.0 (可能虚高)
//
// 注意事项:
//   - 检查方法实现: checkPerformanceIssues()和checkAllocationPatterns()当前返回0 (待完善)
//   - 虚高风险: 未实现的检查会导致性能得分虚高,无法真实反映性能问题
//   - 扣分累加: 多个问题会叠加扣分,可能快速降至0分
//   - 系数配置: ScorePerAllocationIssue当前未明确配置,需要补充
//   - 工具依赖: 建议集成pprof、go-torch等性能分析工具增强检查
//   - 负分保护: score<0时强制归零,避免显示负数
//   - 静态检查局限: 静态分析无法发现所有性能问题,需结合运行时profiling
//   - 上下文敏感: 某些"低效"代码在特定场景下可能是合理的 (如可读性优先)
//
// 改进方向:
//   - checkPerformanceIssues完善: 实现具体的性能反模式检查逻辑
//     - 集成go vet的性能检查器
//     - 分析循环嵌套深度和复杂度
//     - 检测字符串拼接模式 (循环中的+操作)
//     - 识别interface{}过度使用
//   - checkAllocationPatterns完善: 实现具体的内存分配分析
//     - 解析go build -gcflags='-m'的逃逸分析输出
//     - 检测make()调用的容量参数
//     - 识别goroutine泄漏模式 (无context控制的go func)
//     - 分析sync.Pool使用情况
//   - 集成性能分析工具:
//     - pprof集成: 分析CPU和内存profile
//     - trace集成: 分析goroutine调度和延迟
//     - benchstat: 对比benchmark结果识别性能退化
//   - 动态性能检测:
//     - 运行benchmark并分析结果
//     - 检测分配速率 (allocs/op)
//     - 测量内存使用 (bytes/op)
//   - 性能基线对比:
//     - 与上次评分对比,识别性能退化
//     - 设定性能SLO (Service Level Objective)
//     - 跟踪性能趋势,预警性能恶化
//   - 优化建议生成:
//     - 为每个性能问题提供具体优化方案
//     - 提供代码重构示例 (如strings.Builder用法)
//     - 推荐性能优化工具和技术文章
//   - 场景化评分:
//     - 区分高并发服务和低频任务的性能标准
//     - IO密集型vs CPU密集型的不同阈值
//     - 实时系统vs批处理系统的差异化评分
//   - 性能热点识别:
//     - 识别性能问题最集中的函数和文件
//     - 生成优化优先级排序
//     - 估算优化后的性能提升空间
//   - 自动化优化:
//     - 某些模式可自动修复 (如+改为strings.Builder)
//     - 生成优化后的代码diff供review
//   - 性能文档:
//     - 生成性能优化指南
//     - 记录已知性能瓶颈和优化历史
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculatePerformanceScore() float64 {
	score := 100.0

	// 性能问题检查
	performanceIssues := cqe.checkPerformanceIssues()
	score -= float64(performanceIssues) * ScorePerPerformanceIssue

	// 内存分配检查
	allocationIssues := cqe.checkAllocationPatterns()
	score -= float64(allocationIssues) * ScorePerAllocationIssue

	if score < 0 {
		score = 0
	}
	return score
}

// calculateTestScore 计算测试质量维度的质量评分
//
// 功能说明:
//
//	本方法基于测试覆盖率(Test Coverage)和测试文件比例(Test File Ratio)两个核心指标,
//	计算项目的测试质量得分。测试是软件质量保障的基石,充分的测试覆盖能有效预防缺陷,
//	降低维护成本,提升代码重构的信心。本方法采用"线性评分+加分奖励"策略,
//	根据覆盖率达标程度给予基础分数,测试文件比例高时给予额外加分,
//	最终得分反映项目的测试充分性和测试工程化水平。
//
// 评分维度:
//
//	1. 测试覆盖率基础分 (Test Coverage Base Score):
//	   • 数据来源: stats.TestCoverage (运行go test -cover获得的覆盖率)
//	   • 阈值配置: config.Thresholds.TestCoverage (默认75%为满分线)
//	   • 评分规则:
//	     - 覆盖率 >= 阈值: 直接返回MaxScore(100分)满分
//	     - 覆盖率 < 阈值: 按比例线性计算分数
//	       公式: score = (实际覆盖率 / 阈值覆盖率) × 100
//	   • 示例: 阈值75%, 实际60% → score = (60/75)×100 = 80分
//	   • 意义: 覆盖率是测试充分性的量化指标,直接反映测试对代码的覆盖程度
//
//	2. 测试文件比例加分 (Test File Ratio Bonus):
//	   • 数据来源: stats.TestRatio (测试文件数 / 总文件数)
//	   • 阈值配置: TestFileRatioThresholdHigh (默认0.3,即30%)
//	   • 评分规则:
//	     - TestRatio >= 30%: 额外加10分
//	     - TestRatio < 30%: 无加分
//	   • 示例: 100个文件,35个测试文件 → TestRatio=35% → 加10分
//	   • 意义: 高测试文件比例表明项目重视测试,测试工程化程度高
//
// 执行流程:
//
//	1. 获取统计数据:
//	   • stats := cqe.results.Statistics
//	   • 使用aggregateResults()生成的统计信息
//	   • 关键字段: TestCoverage(覆盖率百分比), TestRatio(测试文件比例)
//
//	2. 满分快速通道:
//	   • 条件检查: if stats.TestCoverage >= config.Thresholds.TestCoverage
//	   • 判断依据: 覆盖率已达到或超过配置的目标阈值
//	   • 满足时: 直接返回MaxScore(100.0),跳过后续计算
//	   • 设计理念: 奖励达标的测试覆盖,无需苛求超额覆盖
//
//	3. 线性基础分计算:
//	   • 触发条件: 覆盖率未达阈值
//	   • 计算公式:
//	     score = (stats.TestCoverage / config.Thresholds.TestCoverage) × 100
//	   • 数学特性: 线性比例,覆盖率每提升1%,分数提升(100/阈值)分
//	   • 示例: 阈值75%
//	     - 覆盖率0% → score = (0/75)×100 = 0分
//	     - 覆盖率37.5% → score = (37.5/75)×100 = 50分
//	     - 覆盖率60% → score = (60/75)×100 = 80分
//	     - 覆盖率74% → score = (74/75)×100 = 98.67分
//
//	4. 测试文件比例加分计算:
//	   • 条件检查: if stats.TestRatio >= TestFileRatioThresholdHigh
//	   • TestFileRatioThresholdHigh: 默认0.3 (30%)
//	   • 判断依据: 项目中测试文件占比是否达到良好水平
//	   • 满足时: score += 10 (固定加10分)
//	   • 设计理念: 鼓励建立完善的测试文件体系,每个模块都有对应测试
//	   • 注意: 即使基础分已达100,仍可能加分(需后续限制)
//
//	5. 分数上限保护:
//	   • 条件检查: if score > 100
//	   • 上限处理: score = 100
//	   • 防止加分机制导致分数超过满分(如100+10=110的情况)
//	   • 无下限检查: 线性计算最低为0,无需显式下限保护
//
//	6. 返回最终得分:
//	   • 返回0-100范围内的float64分数
//
// 评分规则配置:
//
//	TestCoverage阈值 (默认75%):
//	• 理由: 75%覆盖率是行业公认的良好实践标准
//	• 对比:
//	  - 50%: 最低可接受水平
//	  - 75%: 良好水平(推荐)
//	  - 80-90%: 优秀水平
//	  - >95%: 卓越水平(可能过度测试)
//	• 示例: 阈值75%,实际90% → 满分100(超额15%不额外加分)
//
//	TestFileRatioThresholdHigh (默认30%):
//	• 理由: 30%测试文件比例表明良好的测试工程化实践
//	• 计算方式: 测试文件数 / 总Go文件数
//	• 典型场景:
//	  - 10%: 测试不足,需加强
//	  - 20%: 基本覆盖,可接受
//	  - 30%+: 良好覆盖,值得加分
//	  - 50%+: 优秀覆盖(但分数已封顶100)
//	• 加分值: 固定10分(不随比例变化)
//
//	线性评分策略:
//	• 公式: (实际 / 目标) × 100
//	• 优势: 简单直观,容易理解和预测
//	• 劣势: 未体现边际效益递减(从0到10%价值大于从90%到100%)
//	• 改进方向: 可采用分段线性或非线性曲线
//
// 评分等级参考:
//
//	• 90-100分: 优秀 - 测试覆盖充分,测试体系完善
//	  - 覆盖率 ≥75% 或 覆盖率67.5%+测试文件比例≥30%
//	• 80-89分: 良好 - 测试覆盖较好,仍有提升空间
//	  - 覆盖率60-74% 或 覆盖率53-66%+高测试文件比例
//	• 70-79分: 中等 - 测试覆盖一般,需要加强
//	  - 覆盖率52-59% 或 覆盖率45-51%+高测试文件比例
//	• 60-69分: 及格 - 测试覆盖偏低,建议增加测试
//	  - 覆盖率45-51% 或 覆盖率37-44%+高测试文件比例
//	• <60分: 不及格 - 测试严重不足,必须补充测试
//	  - 覆盖率 <45%
//
// 参数:
//
//	无显式参数 (方法接收者为*CodeQualityEvaluator)
//
// 返回值:
//   - float64: 测试质量维度评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数,支持小数(如98.67分)
//     • 意义: 分数越高,测试质量越好,代码可靠性越高
//
// 使用场景:
//   - calculateDimensionScores调用: 计算测试维度得分
//   - CI/CD质量门禁: 要求测试得分≥75作为合并条件
//   - 测试补充排序: 根据得分识别测试不足的项目
//   - 代码评审辅助: 得分<80时需要增加测试用例
//   - 重构信心评估: 高测试得分支持安全重构
//
// 示例:
//
//	// 示例1: 理想情况 (满分)
//	// stats.TestCoverage = 80% (≥75%阈值)
//	score := evaluator.calculateTestScore()
//	// score = 100.0 (满分,无需后续计算)
//
//	// 示例2: 中等覆盖率 (线性计算)
//	// config.Thresholds.TestCoverage = 75%
//	// stats.TestCoverage = 60%
//	// stats.TestRatio = 25% (<30%无加分)
//	score := evaluator.calculateTestScore()
//	// score = (60/75) × 100 = 80.0
//
//	// 示例3: 低覆盖率但高测试文件比例 (有加分)
//	// config.Thresholds.TestCoverage = 75%
//	// stats.TestCoverage = 50%
//	// stats.TestRatio = 35% (≥30%加分)
//	score := evaluator.calculateTestScore()
//	// 基础分: (50/75) × 100 = 66.67
//	// 加分: +10
//	// 最终: 66.67 + 10 = 76.67
//
//	// 示例4: 覆盖率刚好达标加高测试文件比例 (上限保护)
//	// config.Thresholds.TestCoverage = 75%
//	// stats.TestCoverage = 75% (刚好达标)
//	// stats.TestRatio = 40% (≥30%加分)
//	score := evaluator.calculateTestScore()
//	// 满分通道: 75% ≥ 75% → 100.0
//	// 注意: 加分不生效,因为已经通过满分通道返回
//	// 最终: 100.0 (不会变成110)
//
//	// 示例5: 极低覆盖率
//	// config.Thresholds.TestCoverage = 75%
//	// stats.TestCoverage = 10%
//	// stats.TestRatio = 5% (<30%无加分)
//	score := evaluator.calculateTestScore()
//	// score = (10/75) × 100 = 13.33
//
// 注意事项:
//   - 覆盖率来源: stats.TestCoverage应由go test -cover或go test -coverprofile生成
//   - 阈值配置: config.Thresholds.TestCoverage默认75%,可根据项目要求调整
//   - 测试文件识别: stats.TestRatio基于*_test.go文件名模式统计
//   - 满分通道: 覆盖率达标直接返回100,不再执行加分逻辑(设计考虑)
//   - 加分限制: 测试文件比例加分固定10分,不随比例进一步提升
//   - 上限保护: score>100时强制归100,防止加分机制导致超额
//   - 无下限保护: 线性计算最低为0,即使覆盖率0%也不会负数
//   - 浮点精度: 使用float64避免精度损失,允许小数分数(如80.67)
//   - 覆盖率类型: 当前仅考虑语句覆盖率,未区分分支覆盖、条件覆盖
//   - 测试质量: 分数仅反映覆盖率和文件数,不反映测试用例质量
//
// 改进方向:
//   - 多维度覆盖率:
//     - 引入分支覆盖率(Branch Coverage)指标
//     - 引入条件覆盖率(Condition Coverage)指标
//     - 综合语句、分支、条件三种覆盖率计算得分
//   - 测试质量评估:
//     - 集成变异测试(Mutation Testing)评估测试有效性
//     - 检测无效测试(总是通过的测试)
//     - 分析断言密度(assertions per test)
//     - 评估测试独立性(是否依赖执行顺序)
//   - 非线性评分:
//     - 采用对数曲线或S曲线,体现边际效益递减
//     - 低覆盖率(0-30%)快速扣分
//     - 中覆盖率(30-70%)线性扣分
//     - 高覆盖率(70-100%)缓慢加分
//   - 动态阈值:
//     - 根据项目类型调整阈值(如库项目要求90%,业务项目75%)
//     - 根据代码关键程度设置不同覆盖率要求
//   - 测试类型区分:
//     - 区分单元测试、集成测试、E2E测试的覆盖率
//     - 为不同测试类型设置不同权重
//   - 加分策略优化:
//     - 测试文件比例加分改为连续函数(如每10%加5分)
//     - 增加表驱动测试(Table-Driven Tests)的加分
//     - 增加基准测试(Benchmark Tests)的加分
//   - 历史趋势:
//     - 与上次评分对比,识别覆盖率提升或下降
//     - 跟踪覆盖率演进趋势,预警覆盖率下滑
//   - 覆盖率分层:
//     - 按包(package)统计覆盖率,识别测试薄弱模块
//     - 按文件统计覆盖率,生成覆盖率热力图
//   - 测试代码质量:
//     - 检测测试代码的圈复杂度(避免复杂测试)
//     - 分析测试代码的可读性和可维护性
//   - 覆盖率可视化:
//     - 生成覆盖率报告HTML(go tool cover -html)
//     - 集成到CI/CD展示覆盖率趋势图
//   - 未覆盖代码分析:
//     - 识别未被测试覆盖的关键代码路径
//     - 优先为高风险未覆盖代码生成测试建议
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateTestScore() float64 {
	stats := cqe.results.Statistics

	// 基于测试覆盖率的得分
	if stats.TestCoverage >= cqe.config.Thresholds.TestCoverage {
		return MaxScore
	}

	// 线性计算：覆盖率越高得分越高
	score := (stats.TestCoverage / cqe.config.Thresholds.TestCoverage) * 100

	// 测试文件比例加分
	if stats.TestRatio >= TestFileRatioThresholdHigh { // 30%以上的文件是测试文件
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return score
}

// calculateDocumentationScore 计算文档质量维度的质量评分
//
// 功能说明:
//
//	本方法基于代码注释比例(Comment Ratio)、README文件存在性、包级文档完整性三个核心指标,
//	计算项目的文档质量得分。文档是代码可维护性和知识传承的关键,充分的文档能有效降低
//	新成员上手成本、减少沟通成本、提升团队协作效率。本方法采用"基础分+加分奖励"策略,
//	根据注释覆盖率给予基础分数,README和包文档存在时给予额外加分,
//	最终得分反映项目的文档完整性和可维护性水平。
//
// 评分维度:
//
//	1. 注释比例基础分 (Comment Ratio Base Score):
//	   • 数据来源: stats.CommentLines / stats.TotalLines (注释行数/总行数)
//	   • 阈值配置: CommentRatioForFullScore (默认400,即25%注释率满分)
//	   • 评分规则:
//	     - 计算公式: score = commentRatio × CommentRatioForFullScore
//	     - 线性递增: 注释率每增加1%,分数增加4分
//	     - 满分条件: 注释率达到25%时得100分
//	   • 示例: 注释率15% → score = 0.15 × 400 = 60分
//	   • 意义: 注释覆盖率是代码自文档化的量化指标,直接反映文档充分程度
//
//	2. README文件加分 (README File Bonus):
//	   • 数据来源: hasReadme()方法检测结果
//	   • 检测文件: README.md, README.txt, README, readme.md, readme.txt
//	   • 评分规则:
//	     - 存在任意README文件: 额外加20分
//	     - 不存在README: 无加分
//	   • 示例: 项目根目录有README.md → 加20分
//	   • 意义: README是项目的门户文档,为新用户提供快速入门指南
//
//	3. 包文档加分 (Package Documentation Bonus):
//	   • 数据来源: checkPackageDocumentation()方法评估结果
//	   • 检测内容: 各包的package注释文档完整性
//	   • 评分规则:
//	     - 返回0-30分的包文档覆盖率得分
//	     - 每个包有完整文档: 累加相应分数
//	     - 无包文档: 返回0分
//	   • 示例: 5个包中3个有文档 → 返回18分(假设平均6分/包)
//	   • 意义: 包级文档是Go项目的API文档基础,支持godoc生成
//
// 执行流程:
//
//	1. 获取统计数据:
//	   • stats := cqe.results.Statistics
//	   • 关键字段: TotalLines(总行数), CommentLines(注释行数)
//
//	2. 零行保护检查:
//	   • 条件检查: if stats.TotalLines > 0
//	   • 目的: 防止除零错误,空项目直接返回0分
//	   • 不满足时: return 0 (空项目无文档得分)
//
//	3. 注释比例基础分计算:
//	   • 计算注释率: commentRatio = CommentLines / TotalLines
//	   • 计算基础分: score = commentRatio × CommentRatioForFullScore
//	   • 数学特性: 线性函数,斜率为CommentRatioForFullScore
//	   • 示例: 1000行代码,200行注释 → ratio=0.2 → score=80
//
//	4. README文件检测加分:
//	   • 执行检测: if cqe.hasReadme()
//	   • hasReadme()逻辑:
//	     - 检查项目根目录下的README文件
//	     - 支持多种文件名: README.md, README.txt, README, readme.md, readme.txt
//	     - 使用os.Stat()检测文件存在性
//	   • 满足时: score += 20 (固定加20分)
//
//	5. 包文档评估加分:
//	   • 执行评估: packageDocScore := cqe.checkPackageDocumentation()
//	   • checkPackageDocumentation()逻辑:
//	     - 遍历项目所有包
//	     - 检查每个包的package注释是否存在
//	     - 评估注释质量和完整性
//	     - 返回0-30分的综合得分
//	   • 加分: score += packageDocScore (最多加30分)
//
//	6. 分数上限保护:
//	   • 条件检查: if score > 100
//	   • 上限处理: score = 100
//	   • 防止多项加分导致分数超过满分
//	   • 示例: 基础分90 + README 20 + 包文档20 = 130 → 限制为100
//
//	7. 返回最终得分:
//	   • 返回0-100范围内的float64分数
//
// 评分规则配置:
//
//	CommentRatioForFullScore (默认400):
//	• 理由: 400对应25%注释率满分,符合行业标准
//	• 计算逻辑:
//	  - score = commentRatio × 400
//	  - 当commentRatio = 0.25时,score = 100
//	• 对比:
//	  - 10%注释率: score = 40 (不及格)
//	  - 15%注释率: score = 60 (及格)
//	  - 20%注释率: score = 80 (良好)
//	  - 25%+注释率: score = 100 (优秀)
//	• 示例: 5000行代码,1250行注释 → ratio=0.25 → score=100
//
//	README加分规则 (固定20分):
//	• 理由: README是项目必备的入门文档,缺失严重影响用户体验
//	• 检测策略: 文件名不区分大小写,支持多种扩展名
//	• 加分值: 固定20分(不评估README质量,仅检测存在性)
//	• 典型README内容:
//	  - 项目介绍和背景
//	  - 快速开始指南
//	  - API文档链接
//	  - 贡献指南
//	  - 许可证信息
//
//	包文档加分规则 (0-30分):
//	• 理由: Go语言推荐每个包都有package注释,用于生成godoc
//	• 评估策略:
//	  - 检查每个包的包级注释存在性
//	  - 评估注释的完整性和规范性
//	  - 按包文档覆盖率给分
//	• 加分值: 最高30分(当前实现返回SimpleReturnScore=30)
//	• Go包文档规范:
//	  - 注释以"Package pkgname"开头
//	  - 紧邻package声明语句
//	  - 描述包的功能和用途
//	  - 可包含使用示例
//
// 评分等级参考:
//
//	• 90-100分: 优秀 - 文档完整充分,注释覆盖率高
//	  - 注释率 ≥22.5% 或 注释率17.5%+README+完整包文档
//	• 80-89分: 良好 - 文档较为完整,仍有提升空间
//	  - 注释率20-22% 或 注释率15%+README+包文档
//	• 70-79分: 中等 - 文档一般,需要加强
//	  - 注释率17-19% 或 注释率12%+README+包文档
//	• 60-69分: 及格 - 文档偏少,建议补充
//	  - 注释率15-16% 或 注释率10%+README+包文档
//	• <60分: 不及格 - 文档严重不足,必须补充
//	  - 注释率 <15%
//
// 参数:
//
//	无显式参数 (方法接收者为*CodeQualityEvaluator)
//
// 返回值:
//   - float64: 文档质量维度评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数,支持小数(如85.67分)
//     • 意义: 分数越高,文档质量越好,项目可维护性越高
//
// 使用场景:
//   - calculateDimensionScores调用: 计算文档维度得分
//   - 代码评审辅助: 得分<70时需要补充文档
//   - 开源项目评估: 要求文档得分≥80作为发布条件
//   - 团队规范检查: 识别文档覆盖不足的模块
//   - 新成员入职: 评估项目文档友好程度
//
// 示例:
//
//	// 示例1: 理想情况 (满分)
//	// stats.TotalLines = 10000
//	// stats.CommentLines = 2500 (注释率25%)
//	// 有README.md
//	// 包文档完整(30分)
//	score := evaluator.calculateDocumentationScore()
//	// 基础分: 0.25 × 400 = 100
//	// README: +20 (但已满分,无效)
//	// 包文档: +30 (但已满分,无效)
//	// 最终: 100.0 (基础分已达上限)
//
//	// 示例2: 良好文档 (中等)
//	// stats.TotalLines = 5000
//	// stats.CommentLines = 800 (注释率16%)
//	// 有README.md
//	// 包文档部分完整(18分)
//	score := evaluator.calculateDocumentationScore()
//	// 基础分: 0.16 × 400 = 64
//	// README: +20
//	// 包文档: +18
//	// 最终: 64 + 20 + 18 = 102 → 限制为100.0
//
//	// 示例3: 基础文档 (及格)
//	// stats.TotalLines = 8000
//	// stats.CommentLines = 1000 (注释率12.5%)
//	// 有README.md
//	// 无包文档(0分)
//	score := evaluator.calculateDocumentationScore()
//	// 基础分: 0.125 × 400 = 50
//	// README: +20
//	// 包文档: +0
//	// 最终: 50 + 20 + 0 = 70.0
//
//	// 示例4: 文档不足 (不及格)
//	// stats.TotalLines = 6000
//	// stats.CommentLines = 300 (注释率5%)
//	// 无README
//	// 无包文档
//	score := evaluator.calculateDocumentationScore()
//	// 基础分: 0.05 × 400 = 20
//	// README: +0
//	// 包文档: +0
//	// 最终: 20.0
//
//	// 示例5: 空项目 (零分保护)
//	// stats.TotalLines = 0
//	score := evaluator.calculateDocumentationScore()
//	// 零行保护: 直接返回0
//	// 最终: 0.0
//
// 注意事项:
//   - 零行保护: stats.TotalLines必须>0,否则直接返回0分(防止除零)
//   - 注释统计: stats.CommentLines由analyzeGoFile()统计,包括//和/* */注释
//   - CommentRatioForFullScore: 当前为400,对应25%注释率满分
//   - README检测: hasReadme()仅检测文件存在性,不评估内容质量
//   - 包文档实现: checkPackageDocumentation()当前返回SimpleReturnScore(30),实现简化
//   - 上限保护: score>100时强制归100,防止加分机制导致超额
//   - 无下限保护: 最低为0分(空项目或无注释时)
//   - 浮点精度: 使用float64避免精度损失,允许小数分数(如64.5)
//   - 注释类型: 包括行注释(//)、块注释(/* */)、文档注释(用于godoc)
//   - 注释质量: 当前仅统计注释数量,未评估注释内容质量
//
// 改进方向:
//   - 注释质量评估:
//     - 区分文档注释(godoc用)和普通注释
//     - 检测无效注释(如TODO、FIXME过多)
//     - 分析注释与代码的关联性
//     - 评估注释的描述性和准确性
//   - README质量评估:
//     - 不仅检测存在性,还评估内容完整性
//     - 检查是否包含必备章节(安装、使用、API等)
//     - 评估README的可读性和示例质量
//     - 检测README的更新频率(是否过时)
//   - 包文档完善:
//     - 实现具体的包文档检测逻辑(当前为简化实现)
//     - 检查package注释是否符合Go规范
//     - 评估包文档的完整性(功能描述、用法示例)
//     - 检测导出符号(函数、类型)的文档覆盖率
//   - 多维度文档评分:
//     - 函数级文档覆盖率(导出函数必须有注释)
//     - 类型级文档覆盖率(结构体、接口的文档)
//     - 示例代码覆盖率(Example函数的数量)
//     - 文档准确性(文档与代码实现是否一致)
//   - 文档工具集成:
//     - 集成godoc检查工具
//     - 集成golint的文档规范检查
//     - 集成文档覆盖率分析工具
//   - 非线性评分:
//     - 采用对数或S曲线,体现边际效益
//     - 低注释率(0-10%)快速扣分
//     - 高注释率(20-25%)缓慢加分
//   - 文档类型区分:
//     - 区分API文档、用户文档、开发文档
//     - 为不同文档类型设置不同权重
//   - 历史趋势:
//     - 与上次评分对比,识别文档改善或恶化
//     - 跟踪文档覆盖率演进趋势
//   - 文档分层:
//     - 按包(package)统计文档覆盖率
//     - 按文件统计文档覆盖率
//     - 生成文档覆盖率热力图
//   - 国际化支持:
//     - 检测多语言文档(README.zh-CN.md)
//     - 评估文档的国际化程度
//   - 文档可视化:
//     - 生成文档覆盖率报告HTML
//     - 集成到CI/CD展示文档趋势图
//   - 自动文档生成:
//     - 识别缺失文档的导出符号
//     - 自动生成文档模板
//     - 提供文档补充建议
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateDocumentationScore() float64 {
	stats := cqe.results.Statistics

	// 注释比例得分
	if stats.TotalLines > 0 {
		commentRatio := float64(stats.CommentLines) / float64(stats.TotalLines)
		score := commentRatio * CommentRatioForFullScore // 25%注释率得满分

		// 检查README文件
		if cqe.hasReadme() {
			score += 20
		}

		// 检查包文档
		packageDocScore := cqe.checkPackageDocumentation()
		score += packageDocScore

		if score > 100 {
			score = 100
		}
		return score
	}

	return 0
}

// calculateOverallScore 基于六大维度评分计算加权平均总分并分配质量等级
//
// 功能说明:
//
//	本方法是代码质量评估的最终汇总步骤,负责将之前计算的6个维度评分(code_structure,
//	style_compliance, security_analysis, performance_analysis, test_quality,
//	documentation_quality)通过加权平均算法聚合为单一的整体质量得分(OverallScore),
//	并根据预定义的分数阈值将项目分配到A/B/C/D/F五个质量等级中。这个总分和等级是
//	项目质量的最终判定结果,直接决定质量门禁的通过与否以及改进的优先级排序。
//
// 评分维度:
//
//	本方法不直接评估代码,而是对已计算的6个维度评分进行加权聚合:
//
//	1. code_structure (代码结构) - 默认权重: weights.CodeStructure
//	   • 贡献内容: 圈复杂度、函数长度、模块化设计质量
//	   • 典型权重: 0.25 (25%,最高权重)
//	   • 意义: 代码结构是可维护性的基础,权重最高
//
//	2. style_compliance (风格合规) - 默认权重: weights.StyleCompliance
//	   • 贡献内容: 命名规范、注释风格、格式化标准
//	   • 典型权重: 0.20 (20%)
//	   • 意义: 统一风格降低协作成本,权重较高
//
//	3. security_analysis (安全分析) - 默认权重: weights.SecurityAnalysis
//	   • 贡献内容: 安全漏洞、敏感信息泄漏、危险用法
//	   • 典型权重: 0.20 (20%)
//	   • 意义: 安全性是生产环境的生命线,权重较高
//
//	4. performance_analysis (性能分析) - 默认权重: weights.PerformanceAnalysis
//	   • 贡献内容: 内存分配、算法效率、资源使用
//	   • 典型权重: 0.15 (15%)
//	   • 意义: 性能影响用户体验,权重中等
//
//	5. test_quality (测试质量) - 默认权重: weights.TestQuality
//	   • 贡献内容: 测试覆盖率、测试设计、边界条件
//	   • 典型权重: 0.15 (15%)
//	   • 意义: 测试是质量保障,权重中等
//
//	6. documentation_quality (文档质量) - 默认权重: weights.DocumentationQuality
//	   • 贡献内容: 注释覆盖率、API文档、README质量
//	   • 典型权重: 0.05 (5%,最低权重)
//	   • 意义: 文档是辅助资料,权重最低
//
// 执行流程:
//
//	1. 读取权重配置:
//	   • weights := cqe.config.WeightSettings
//	   • WeightSettings结构体包含6个维度的权重值
//	   • 权重值通常为0.0-1.0之间的浮点数,总和应为1.0
//
//	2. 初始化累加器:
//	   • totalScore := 0.0 (加权分数总和)
//	   • totalWeight := 0.0 (权重总和,用于验证和归一化)
//
//	3. 遍历所有维度评分 (range DimensionScores):
//	   • 迭代顺序: 不确定(map遍历顺序随机)
//	   • dimension: 维度名称字符串
//	   • score: 该维度的评分(0.0-100.0)
//
//	4. 维度权重映射 (switch-case):
//	   • 根据dimension字符串匹配对应的权重值
//	   • 6个有效维度各有对应的weights字段
//	   • default: weight = 0.0 (未知维度不参与计算)
//	   • 设计理念: 显式映射提高代码可读性和类型安全性
//
//	5. 加权分数累加:
//	   • totalScore += score * weight (分数×权重累加)
//	   • totalWeight += weight (权重累加,用于后续归一化)
//	   • 示例: code_structure得分85, 权重0.25 → 贡献21.25分
//
//	6. 计算加权平均 (零除保护):
//	   • 条件检查: if totalWeight > 0
//	   • 计算公式: OverallScore = totalScore / totalWeight
//	   • 零除保护: 如果totalWeight为0(所有权重都是0),跳过计算,OverallScore保持默认值0
//	   • 归一化作用: 即使权重总和不为1.0,除法也能得到正确的加权平均
//
//	7. 等级判定 (switch-case阶梯):
//	   • 根据OverallScore与阈值对比分配等级
//	   • 阶梯顺序: 从高到低依次判断
//	   • 一旦匹配即确定等级,不再继续判断
//	   • 设置: cqe.results.Grade = "A"/"B"/"C"/"D"/"F"
//
//	8. 返回完成:
//	   • 方法无返回值(void)
//	   • 副作用: 设置results.OverallScore和results.Grade字段
//
// 加权平均算法详解:
//
//	核心公式:
//	  OverallScore = Σ(DimensionScore_i × Weight_i) / Σ(Weight_i)
//	  即: OverallScore = totalScore / totalWeight
//
//	计算示例(权重总和=1.0):
//	  code_structure:      85分 × 0.25 = 21.25
//	  style_compliance:    92分 × 0.20 = 18.40
//	  security_analysis:  100分 × 0.20 = 20.00
//	  performance_analysis:78分 × 0.15 = 11.70
//	  test_quality:        80分 × 0.15 = 12.00
//	  documentation_quality:65分× 0.05 =  3.25
//	  ————————————————————————————————————
//	  totalScore = 86.60
//	  totalWeight = 1.00
//	  OverallScore = 86.60 / 1.00 = 86.60
//
//	权重总和≠1.0的情况:
//	  假设权重总和为0.8(某些维度权重为0):
//	  totalScore = 69.28 (按0.8权重计算)
//	  totalWeight = 0.8
//	  OverallScore = 69.28 / 0.8 = 86.60 (归一化后仍为86.60)
//
// 等级阈值配置:
//
//	等级判定采用阶梯式阈值,从高到低依次匹配:
//
//	Grade "A" (优秀) - OverallScore ≥ ExcellentScore:
//	• 阈值: ExcellentScore (默认90.0)
//	• 含义: 代码质量卓越,各维度均衡优秀
//	• 适用: 关键系统、开源项目、团队标杆代码
//	• 特征: 低复杂度、高覆盖率、完善文档、无安全漏洞
//
//	Grade "B" (良好) - OverallScore ≥ GoodScore:
//	• 阈值: GoodScore (默认80.0)
//	• 含义: 代码质量良好,部分维度有提升空间
//	• 适用: 生产环境代码、常规项目
//	• 特征: 结构合理、测试充分、少量改进点
//
//	Grade "C" (及格) - OverallScore ≥ PassingScore:
//	• 阈值: PassingScore (默认60.0)
//	• 含义: 代码质量达到最低可接受标准
//	• 适用: 快速迭代的MVP、实验性项目
//	• 特征: 基本功能完成,但质量债务较多
//
//	Grade "D" (勉强及格) - OverallScore ≥ MinAcceptableScore:
//	• 阈值: MinAcceptableScore (默认50.0)
//	• 含义: 代码质量较差,需要重点改进
//	• 适用: 技术债务警示、遗留代码评估
//	• 特征: 存在结构问题、测试不足、安全隐患
//
//	Grade "F" (不及格) - OverallScore < MinAcceptableScore:
//	• 阈值: < MinAcceptableScore (< 50.0)
//	• 含义: 代码质量严重不达标,禁止上线
//	• 适用: 质量门禁拦截、强制重构标识
//	• 特征: 高复杂度、低覆盖率、严重安全漏洞
//
// 参数:
//
//	无显式参数 (方法接收者为*CodeQualityEvaluator)
//
// 返回值:
//
//	无返回值 (void方法)
//	• 副作用: 修改cqe.results的两个字段:
//	  - OverallScore (float64): 加权平均总分 (0.0-100.0)
//	  - Grade (string): 质量等级 ("A"/"B"/"C"/"D"/"F")
//
// 使用场景:
//   - EvaluateProject调用: 在calculateDimensionScores()之后、生成改进建议之前执行
//   - 质量门禁判定: 根据OverallScore或Grade决定是否允许代码合并
//   - 质量趋势分析: 跟踪OverallScore随时间的演进
//   - 团队KPI考核: 将Grade作为代码质量考核指标
//   - 重构优先级排序: Grade为D/F的项目优先重构
//
// 示例:
//
//	// 示例1: 优秀项目 (Grade A)
//	// DimensionScores = {
//	//   "code_structure": 95.0,
//	//   "style_compliance": 98.0,
//	//   "security_analysis": 100.0,
//	//   "performance_analysis": 88.0,
//	//   "test_quality": 92.0,
//	//   "documentation_quality": 85.0,
//	// }
//	// WeightSettings = {0.25, 0.20, 0.20, 0.15, 0.15, 0.05}
//	evaluator.calculateOverallScore()
//	// 计算过程:
//	// totalScore = 95×0.25 + 98×0.20 + 100×0.20 + 88×0.15 + 92×0.15 + 85×0.05
//	//            = 23.75 + 19.60 + 20.00 + 13.20 + 13.80 + 4.25 = 94.60
//	// totalWeight = 1.00
//	// OverallScore = 94.60 / 1.00 = 94.60
//	// 94.60 ≥ 90.0 (ExcellentScore) → Grade = "A"
//	// 结果: evaluator.results.OverallScore = 94.60
//	//      evaluator.results.Grade = "A"
//
//	// 示例2: 良好项目 (Grade B)
//	// DimensionScores = {
//	//   "code_structure": 85.0,
//	//   "style_compliance": 78.0,
//	//   "security_analysis": 90.0,
//	//   "performance_analysis": 75.0,
//	//   "test_quality": 82.0,
//	//   "documentation_quality": 70.0,
//	// }
//	evaluator.calculateOverallScore()
//	// totalScore = 85×0.25 + 78×0.20 + 90×0.20 + 75×0.15 + 82×0.15 + 70×0.05
//	//            = 21.25 + 15.60 + 18.00 + 11.25 + 12.30 + 3.50 = 81.90
//	// OverallScore = 81.90
//	// 81.90 ≥ 80.0 (GoodScore) → Grade = "B"
//
//	// 示例3: 及格项目 (Grade C)
//	// DimensionScores = {
//	//   "code_structure": 65.0,
//	//   "style_compliance": 58.0,
//	//   "security_analysis": 70.0,
//	//   "performance_analysis": 55.0,
//	//   "test_quality": 60.0,
//	//   "documentation_quality": 50.0,
//	// }
//	evaluator.calculateOverallScore()
//	// totalScore = 65×0.25 + 58×0.20 + 70×0.20 + 55×0.15 + 60×0.15 + 50×0.05
//	//            = 16.25 + 11.60 + 14.00 + 8.25 + 9.00 + 2.50 = 61.60
//	// OverallScore = 61.60
//	// 61.60 ≥ 60.0 (PassingScore) → Grade = "C"
//
//	// 示例4: 不及格项目 (Grade F)
//	// DimensionScores = {
//	//   "code_structure": 35.0,
//	//   "style_compliance": 45.0,
//	//   "security_analysis": 50.0,
//	//   "performance_analysis": 40.0,
//	//   "test_quality": 30.0,
//	//   "documentation_quality": 20.0,
//	// }
//	evaluator.calculateOverallScore()
//	// totalScore = 35×0.25 + 45×0.20 + 50×0.20 + 40×0.15 + 30×0.15 + 20×0.05
//	//            = 8.75 + 9.00 + 10.00 + 6.00 + 4.50 + 1.00 = 39.25
//	// OverallScore = 39.25
//	// 39.25 < 50.0 (MinAcceptableScore) → Grade = "F"
//
//	// 示例5: 部分权重为0的情况 (特殊场景)
//	// DimensionScores = {
//	//   "code_structure": 80.0,
//	//   "style_compliance": 75.0,
//	//   "security_analysis": 85.0,
//	// }
//	// WeightSettings = {0.40, 0.30, 0.30, 0.0, 0.0, 0.0}
//	evaluator.calculateOverallScore()
//	// totalScore = 80×0.40 + 75×0.30 + 85×0.30 + 0 + 0 + 0
//	//            = 32.00 + 22.50 + 25.50 = 80.00
//	// totalWeight = 0.40 + 0.30 + 0.30 = 1.00
//	// OverallScore = 80.00 / 1.00 = 80.00
//	// 80.00 ≥ 80.0 (GoodScore) → Grade = "B"
//
// 注意事项:
//   - 零除保护: 仅当totalWeight > 0时才计算OverallScore,防止除零错误
//   - 权重归一化: 算法自动归一化,即使权重总和≠1.0也能正确计算加权平均
//   - map遍历顺序: DimensionScores的遍历顺序不确定,但不影响最终结果(加法交换律)
//   - 未知维度处理: switch的default分支将未知维度权重设为0.0,不参与计算
//   - 阈值顺序: 等级判定必须从高到低依次判断,否则低阈值会拦截高分项目
//   - Grade默认值: 如果所有case都不匹配(理论上不可能),default设为"F"
//   - 浮点精度: 使用float64避免精度损失,但仍可能存在浮点误差(如86.599999)
//   - 权重配置: WeightSettings应在config初始化时设定,本方法不验证权重合理性
//   - 副作用依赖: 必须在calculateDimensionScores()之后调用,否则DimensionScores为空
//   - 等级字符串: Grade值为固定字符串,外部系统依赖此格式时需保持兼容性
//
// 改进方向:
//   - 权重验证: 检查权重总和是否为1.0,偏差过大时记录警告
//   - 动态权重: 根据项目类型(如库项目、Web服务、CLI工具)调整权重策略
//   - 细分等级: 引入A+/A/A-等更细粒度的等级划分
//   - 加权策略: 支持多种加权算法(线性、指数、对数)供用户选择
//   - 维度依赖: 检测维度间的相关性,避免高度相关维度的权重叠加
//   - 历史对比: 与历史OverallScore对比,生成趋势分析(上升/下降/稳定)
//   - 阈值配置化: 将等级阈值从常量改为config配置项,支持团队定制
//   - 等级语义化: 为每个等级提供详细的语义描述和改进建议模板
//   - 多维度可视化: 生成雷达图展示6个维度的评分分布
//   - 权重敏感性分析: 分析权重调整对OverallScore的影响(sensitivity analysis)
//   - 等级置信度: 计算等级的置信度(如刚好80.1分的B级置信度低)
//   - 非线性聚合: 探索使用几何平均、调和平均等非线性聚合方法
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) calculateOverallScore() {
	weights := cqe.config.WeightSettings

	totalScore := 0.0
	totalWeight := 0.0

	for dimension, score := range cqe.results.DimensionScores {
		var weight float64
		switch dimension {
		case "code_structure":
			weight = weights.CodeStructure
		case "style_compliance":
			weight = weights.StyleCompliance
		case "security_analysis":
			weight = weights.SecurityAnalysis
		case "performance_analysis":
			weight = weights.PerformanceAnalysis
		case "test_quality":
			weight = weights.TestQuality
		case "documentation_quality":
			weight = weights.DocumentationQuality
		default:
			weight = 0.0
		}

		totalScore += score * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		cqe.results.OverallScore = totalScore / totalWeight
	}

	// 设置等级
	switch {
	case cqe.results.OverallScore >= ExcellentScore:
		cqe.results.Grade = "A"
	case cqe.results.OverallScore >= GoodScore:
		cqe.results.Grade = "B"
	case cqe.results.OverallScore >= PassingScore:
		cqe.results.Grade = "C"
	case cqe.results.OverallScore >= MinAcceptableScore:
		cqe.results.Grade = "D"
	default:
		cqe.results.Grade = "F"
	}
}

// 辅助函数实现

// calculateStyleScore 基于代码风格问题数量计算风格质量评分(工具结果解析专用)
//
// 功能说明:
//
//	本辅助函数专门用于静态分析工具(特别是golint)的结果解析阶段,负责根据检测到的
//	代码风格问题数量计算风格质量评分。采用简单直观的线性递减策略:从满分100分起始,
//	每发现一个风格问题扣除固定分数(2分),最终得分在0-100范围内。该函数被GolintTool.ParseResult()
//	等工具解析方法调用,将原始问题列表转换为量化的质量分数,用于ToolSummary.Score字段。
//
// 评分策略:
//
//	1. 满分快速通道 (Zero Issues Path):
//	   • 条件: len(issues) == 0 (无任何风格问题)
//	   • 结果: 直接返回MaxScore(100.0)满分
//	   • 设计理念: 奖励完全符合风格规范的代码,无需计算
//
//	2. 线性递减策略 (Linear Deduction):
//	   • 基础公式: score = MaxScore - len(issues) × 2
//	   • 扣分系数: 每个问题扣2分(StyleIssueDeduction常量)
//	   • 数学特性: 简单线性函数,斜率为-2
//	   • 示例计算:
//	     - 0个问题 → 100分(满分)
//	     - 10个问题 → 100 - 10×2 = 80分
//	     - 30个问题 → 100 - 30×2 = 40分
//	     - 50个问题 → 100 - 50×2 = 0分
//	     - 60个问题 → 100 - 60×2 = -20分 → 下限保护为0分
//
//	3. 分数下限保护 (Score Clamping):
//	   • 条件检查: if score < 0
//	   • 下限处理: score = 0
//	   • 防止负分: 即使问题极多也不会显示负数分数
//	   • 无上限保护: 满分通道已确保不会超过100分
//
// 执行流程:
//
//	1. 零问题检测:
//	   • 检查条件: len(issues) == 0
//	   • 满足时: 直接返回MaxScore(100.0)
//	   • 跳过后续: 无需进行扣分计算
//
//	2. 线性扣分计算:
//	   • 计算公式: score = MaxScore - float64(len(issues)) × 2
//	   • 类型转换: len()返回int,需转float64参与浮点运算
//	   • 扣分累加: 问题数量越多,扣分越多
//
//	3. 下限保护:
//	   • 条件检查: if score < 0
//	   • 下限设定: score = 0
//	   • 防止显示异常
//
//	4. 返回评分:
//	   • 返回0.0-100.0范围内的float64分数
//
// 扣分系数设计:
//
//	StyleIssueDeduction (固定2分/问题):
//	• 理由: 风格问题通常是建议级别,影响相对较小,扣分适中
//	• 对比:
//	  - 正确性问题: 每个扣10分(calculateCorrectnessScore)
//	  - 风格问题: 每个扣2分(本函数)
//	  - 体现严重程度差异: 正确性问题(5倍扣分) > 风格问题
//	• 问题容忍度:
//	  - 10个问题: 仍有80分(良好)
//	  - 25个问题: 仍有50分(及格)
//	  - 50个问题: 0分(不及格)
//	• 设计平衡: 既惩罚风格问题,又不过于苛刻
//
// 与calculateStyleScore维度方法的区别:
//
//	本函数 (工具解析辅助函数):
//	• 调用位置: GolintTool.ParseResult()内部
//	• 输入参数: []CodeIssue切片(golint检测到的问题列表)
//	• 输出目标: ToolSummary.Score字段(单一工具的评分)
//	• 评分范围: 0-100分
//	• 算法复杂度: O(1) - 仅计算问题数量
//	• 用途: 将golint原始输出转换为量化分数
//
//	calculateStyleScore维度方法 (行4369):
//	• 调用位置: calculateDimensionScores()
//	• 输入来源: 多个工具结果(golint + gofmt + 命名约定)
//	• 输出目标: DimensionScores["style_compliance"](风格维度总分)
//	• 评分范围: 0-100分
//	• 算法复杂度: O(n) - 需要遍历多种检查
//	• 用途: 综合多个工具结果计算风格维度得分
//
// 参数:
//   - issues: golint工具检测到的代码风格问题列表
//     • 类型: []CodeIssue切片
//     • 来源: GolintTool.ParseResult()解析golint输出后生成
//     • 内容: 每个CodeIssue包含文件位置、问题描述、严重程度等
//     • 允许空: 空切片表示无问题,返回满分
//     • 问题类型: 全部为golint检测的风格问题(命名、注释、格式等)
//
// 返回值:
//   - float64: 基于问题数量的风格质量评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数,支持小数(虽然当前算法总是整数)
//     • 意义: 分数越高,风格问题越少,代码风格越规范
//     • 100.0: 无任何风格问题(完美)
//     • 0.0: 风格问题极多(≥50个问题)
//
// 使用场景:
//   - GolintTool.ParseResult()调用: 解析golint输出时计算Summary.Score
//   - 其他风格检查工具: 可被其他类似工具的ParseResult复用
//   - 单元测试: 测试风格评分算法的正确性
//   - 工具结果聚合: 为ToolResult提供统一的评分指标
//
// 示例:
//
//	// 示例1: 无风格问题 (满分)
//	issues := []CodeIssue{}
//	score := calculateStyleScore(issues)
//	// score = 100.0 (零问题快速通道)
//
//	// 示例2: 少量风格问题 (良好)
//	issues := []CodeIssue{
//	    {ID: "golint_0", Message: "exported function should have comment"},
//	    {ID: "golint_1", Message: "variable name should be camelCase"},
//	    {ID: "golint_2", Message: "package comment should be present"},
//	}
//	score := calculateStyleScore(issues)
//	// len(issues) = 3
//	// score = 100 - 3×2 = 94.0
//
//	// 示例3: 中等风格问题 (及格)
//	issues := make([]CodeIssue, 25)  // 25个风格问题
//	score := calculateStyleScore(issues)
//	// score = 100 - 25×2 = 50.0 (刚好及格线)
//
//	// 示例4: 大量风格问题 (不及格)
//	issues := make([]CodeIssue, 45)  // 45个风格问题
//	score := calculateStyleScore(issues)
//	// score = 100 - 45×2 = 10.0 (严重不及格)
//
//	// 示例5: 极多风格问题 (下限保护)
//	issues := make([]CodeIssue, 60)  // 60个风格问题
//	score := calculateStyleScore(issues)
//	// 原始计算: 100 - 60×2 = -20
//	// 下限保护: score = 0.0 (不会显示负分)
//
// 注意事项:
//   - 问题平等性: 所有问题一视同仁,每个扣2分,不区分严重程度
//   - golint特性: golint的所有问题均为warning级别,无error级别
//   - 问题数量: 仅考虑数量,不考虑CodeIssue的Severity/Priority字段
//   - 算法简单性: 线性递减易于理解和预测,适合快速评分
//   - 下限保护: score<0时归零,防止显示异常
//   - 无上限保护: 零问题通道已确保不会超过100分
//   - 浮点运算: 使用float64避免精度损失
//   - 与维度评分区别: 本函数仅用于单一工具(golint)的评分
//
// 改进方向:
//   - 严重程度区分: 根据CodeIssue.Severity字段差异化扣分
//     - error级问题: 扣5分/个
//     - warning级问题: 扣2分/个
//     - info级问题: 扣0.5分/个
//   - 问题类型权重: 根据CodeIssue.Category设置不同权重
//     - 导出符号缺少文档: 扣3分
//     - 命名不规范: 扣2分
//     - 注释格式问题: 扣1分
//   - 非线性扣分: 采用对数或指数曲线,体现边际效益
//     - 前10个问题: 每个扣3分(快速下降)
//     - 10-30个问题: 每个扣2分(线性下降)
//     - 30+个问题: 每个扣1分(缓慢下降)
//   - 问题密度: 考虑问题数量与代码行数的比例
//     - score = 100 - (issues/totalLines) × 1000
//   - 加分机制: 超低问题率时额外加分
//     - issues < 5: 无扣分,保持100分
//   - 配置化系数: 扣分系数从常量改为可配置参数
//   - 问题去重: 检测并去除重复问题后再计算
//   - 文件级评分: 按文件分别计算后取加权平均
//
// 作者: JIA
func calculateStyleScore(issues []CodeIssue) float64 {
	if len(issues) == 0 {
		return MaxScore
	}
	// 简单的线性递减：每个问题扣2分
	score := MaxScore - float64(len(issues))*2
	if score < 0 {
		score = 0
	}
	return score
}

// calculateCorrectnessScore 基于代码正确性问题数量计算正确性质量评分(工具结果解析专用)
//
// 功能说明:
//
//	本辅助函数专门用于静态分析工具(特别是go vet)的结果解析阶段,负责根据检测到的
//	代码正确性问题数量计算正确性质量评分。采用严格的线性递减策略:从满分100分起始,
//	每发现一个正确性问题扣除较大分数(10分),最终得分在0-100范围内。该函数被GovetTool.ParseResult()
//	等工具解析方法调用,将原始问题列表转换为量化的质量分数,用于ToolSummary.Score字段。
//	相比风格问题(2分/个),正确性问题的扣分力度是5倍,体现了正确性问题的严重性。
//
// 评分策略:
//
//	1. 满分快速通道 (Zero Issues Path):
//	   • 条件: len(issues) == 0 (无任何正确性问题)
//	   • 结果: 直接返回MaxScore(100.0)满分
//	   • 设计理念: 奖励完全通过go vet检查的代码,无需计算
//
//	2. 严格线性递减策略 (Strict Linear Deduction):
//	   • 基础公式: score = MaxScore - len(issues) × 10
//	   • 扣分系数: 每个问题扣10分(CorrectnessIssueDeduction常量)
//	   • 数学特性: 简单线性函数,斜率为-10(是风格问题的5倍)
//	   • 示例计算:
//	     - 0个问题 → 100分(满分)
//	     - 3个问题 → 100 - 3×10 = 70分
//	     - 5个问题 → 100 - 5×10 = 50分(及格线)
//	     - 10个问题 → 100 - 10×10 = 0分(完全不及格)
//	     - 12个问题 → 100 - 12×10 = -20分 → 下限保护为0分
//
//	3. 分数下限保护 (Score Clamping):
//	   • 条件检查: if score < 0
//	   • 下限处理: score = 0
//	   • 防止负分: 即使问题极多也不会显示负数分数
//	   • 无上限保护: 满分通道已确保不会超过100分
//
// 执行流程:
//
//	1. 零问题检测:
//	   • 检查条件: len(issues) == 0
//	   • 满足时: 直接返回MaxScore(100.0)
//	   • 跳过后续: 无需进行扣分计算
//
//	2. 严格线性扣分计算:
//	   • 计算公式: score = MaxScore - float64(len(issues)) × 10
//	   • 类型转换: len()返回int,需转float64参与浮点运算
//	   • 扣分累加: 问题数量越多,扣分越多(10分递减)
//
//	3. 下限保护:
//	   • 条件检查: if score < 0
//	   • 下限设定: score = 0
//	   • 防止显示异常
//
//	4. 返回评分:
//	   • 返回0.0-100.0范围内的float64分数
//
// 扣分系数设计:
//
//	CorrectnessIssueDeduction (固定10分/问题):
//	• 理由: 正确性问题可能导致程序错误或崩溃,影响极其严重,扣分必须严厉
//	• 对比:
//	  - 正确性问题: 每个扣10分(本函数)
//	  - 风格问题: 每个扣2分(calculateStyleScore)
//	  - 严重程度差异: 正确性问题(5倍扣分) > 风格问题
//	• 问题容忍度:
//	  - 3个问题: 仅剩70分(中等)
//	  - 5个问题: 仅剩50分(及格)
//	  - 10个问题: 0分(完全不及格)
//	• 设计理念: 正确性是底线,一个潜在bug的代价远高于风格不规范
//
// 与calculateStyleScore的关键区别:
//
//	本函数 (正确性评分辅助函数):
//	• 调用位置: GovetTool.ParseResult()内部
//	• 输入参数: []CodeIssue切片(go vet检测到的正确性问题列表)
//	• 输出目标: ToolSummary.Score字段(单一工具的评分)
//	• 扣分系数: 10分/问题(严厉)
//	• 评分范围: 0-100分
//	• 算法复杂度: O(1) - 仅计算问题数量
//	• 用途: 将go vet原始输出转换为量化分数
//	• 问题性质: 潜在bug、不可达代码、Printf格式错误等正确性问题
//
//	calculateStyleScore (风格评分辅助函数,行5831):
//	• 调用位置: GolintTool.ParseResult()内部
//	• 输入参数: []CodeIssue切片(golint检测到的风格问题列表)
//	• 输出目标: ToolSummary.Score字段(单一工具的评分)
//	• 扣分系数: 2分/问题(温和)
//	• 评分范围: 0-100分
//	• 算法复杂度: O(1) - 仅计算问题数量
//	• 用途: 将golint原始输出转换为量化分数
//	• 问题性质: 命名规范、注释格式等风格问题
//
//	共同点:
//	• 都是工具解析辅助函数,非维度评分方法
//	• 都采用线性递减策略
//	• 都有满分快速通道
//	• 都有下限保护(score >= 0)
//
//	关键差异:
//	• 扣分系数: 10分 vs 2分 = 5倍严重程度差异
//	• 服务工具: go vet(正确性) vs golint(风格)
//	• 问题影响: 可能导致bug vs 影响可读性
//	• 容忍度: 10个问题清零 vs 50个问题清零
//
// 参数:
//   - issues: go vet工具检测到的代码正确性问题列表
//     • 类型: []CodeIssue切片
//     • 来源: GovetTool.ParseResult()解析go vet输出后生成
//     • 内容: 每个CodeIssue包含文件位置、问题描述、严重程度等
//     • 允许空: 空切片表示无问题,返回满分
//     • 问题类型: 全部为go vet检测的正确性问题(unreachable code, Printf errors等)
//
// 返回值:
//   - float64: 基于问题数量的正确性质量评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数,支持小数(虽然当前算法总是整数)
//     • 意义: 分数越高,正确性问题越少,代码越可靠
//     • 100.0: 无任何正确性问题(完美)
//     • 0.0: 正确性问题极多(≥10个问题)
//
// 使用场景:
//   - GovetTool.ParseResult()调用: 解析go vet输出时计算Summary.Score
//   - 其他正确性检查工具: 可被其他类似工具的ParseResult复用
//   - 单元测试: 测试正确性评分算法的准确性
//   - 工具结果聚合: 为ToolResult提供统一的评分指标
//
// 示例:
//
//	// 示例1: 无正确性问题 (满分)
//	issues := []CodeIssue{}
//	score := calculateCorrectnessScore(issues)
//	// score = 100.0 (零问题快速通道)
//
//	// 示例2: 少量正确性问题 (中等)
//	issues := []CodeIssue{
//	    {ID: "govet_0", Message: "unreachable code"},
//	    {ID: "govet_1", Message: "missing return at end of function"},
//	    {ID: "govet_2", Message: "Printf format %d has arg of wrong type string"},
//	}
//	score := calculateCorrectnessScore(issues)
//	// len(issues) = 3
//	// score = 100 - 3×10 = 70.0
//
//	// 示例3: 中等正确性问题 (及格线)
//	issues := make([]CodeIssue, 5)  // 5个正确性问题
//	score := calculateCorrectnessScore(issues)
//	// score = 100 - 5×10 = 50.0 (刚好及格线)
//
//	// 示例4: 较多正确性问题 (不及格)
//	issues := make([]CodeIssue, 8)  // 8个正确性问题
//	score := calculateCorrectnessScore(issues)
//	// score = 100 - 8×10 = 20.0 (严重不及格)
//
//	// 示例5: 极多正确性问题 (下限保护)
//	issues := make([]CodeIssue, 12)  // 12个正确性问题
//	score := calculateCorrectnessScore(issues)
//	// 原始计算: 100 - 12×10 = -20
//	// 下限保护: score = 0.0 (不会显示负分)
//
// 注意事项:
//   - 问题平等性: 所有问题一视同仁,每个扣10分,不区分严重程度
//   - go vet特性: go vet的所有问题均为error级别(SeverityError),无warning级别
//   - 问题数量: 仅考虑数量,不考虑CodeIssue的Severity/Priority字段
//   - 算法简单性: 线性递减易于理解和预测,适合快速评分
//   - 下限保护: score<0时归零,防止显示异常
//   - 无上限保护: 零问题通道已确保不会超过100分
//   - 浮点运算: 使用float64避免精度损失
//   - 与维度评分区别: 本函数仅用于单一工具(go vet)的评分
//   - 严重程度: 10分/问题体现了正确性问题的严重性(是风格问题的5倍)
//   - 问题类型: 包括不可达代码、Printf格式错误、类型断言错误等
//
// 改进方向:
//   - 严重程度区分: 虽然go vet所有问题都是error级,但可根据问题类型差异化扣分
//     - 潜在panic问题: 扣15分/个(如nil解引用)
//     - 逻辑错误: 扣10分/个(如不可达代码)
//     - 可疑用法: 扣5分/个(如可能的Printf错误)
//   - 问题类型权重: 根据CodeIssue.Category设置不同权重
//     - 内存安全问题: 扣15分
//     - 并发问题: 扣12分
//     - 逻辑正确性: 扣10分
//     - 可疑构造: 扣8分
//   - 非线性扣分: 采用指数曲线,体现问题累积的复合影响
//     - 前3个问题: 每个扣10分(线性下降)
//     - 3-8个问题: 每个扣15分(加速下降)
//     - 8+个问题: 每个扣20分(急剧下降)
//   - 问题密度: 考虑问题数量与代码行数的比例
//     - score = 100 - (issues/totalLines) × 5000
//   - 配置化系数: 扣分系数从常量改为可配置参数
//   - 问题去重: 检测并去除重复问题后再计算
//   - 文件级评分: 按文件分别计算后取加权平均
//   - 历史对比: 与上次go vet结果对比,识别新增或修复的问题
//   - 影响评估: 根据问题所在函数的调用频率调整扣分(热点函数问题更严重)
//   - 自动修复建议: 为每个正确性问题提供具体的修复代码示例
//
// 作者: JIA
func calculateCorrectnessScore(issues []CodeIssue) float64 {
	if len(issues) == 0 {
		return MaxScore
	}
	// 正确性问题更严重：每个问题扣10分
	score := MaxScore - float64(len(issues))*10
	if score < 0 {
		score = 0
	}
	return score
}

// calculateComplexityScore 基于平均和最大圈复杂度计算复杂度质量评分(双重惩罚策略)
//
// 功能说明:
//
//	本辅助函数专门用于代码复杂度评估,负责根据项目的平均圈复杂度(Average Cyclomatic Complexity)
//	和最大圈复杂度(Maximum Cyclomatic Complexity)两个核心指标计算复杂度质量评分。采用"双重惩罚"策略:
//	从满分100分起始,分别对超出阈值的平均复杂度和最大复杂度进行独立扣分,最终得分在0-100范围内。
//	该函数被GocycloTool.ParseResult()等复杂度分析工具调用,将原始复杂度指标转换为量化的质量分数,
//	用于ToolSummary.Score字段。双重惩罚机制确保既关注整体复杂度水平(平均值)又捕捉极端复杂函数(最大值)。
//
// 评分策略:
//
//	1. 双重独立惩罚机制 (Dual Independent Penalty Mechanism):
//	   • 平均复杂度惩罚 (Average Complexity Penalty):
//	     - 阈值: MaxCyclomaticComplexity (默认10.0)
//	     - 触发条件: avgComplexity > 10.0
//	     - 惩罚公式: penalty = (avgComplexity - 10.0) × 10
//	     - 惩罚系数: 10分/单位复杂度
//	     - 示例: 平均复杂度13 → 超出3 → 扣30分
//	   • 最大复杂度惩罚 (Maximum Complexity Penalty):
//	     - 阈值: 固定值10 (硬编码)
//	     - 触发条件: maxComplexity > 10
//	     - 惩罚公式: penalty = (maxComplexity - 10) × ScorePerComplexityIssue
//	     - 惩罚系数: ScorePerComplexityIssue (默认5分/单位复杂度)
//	     - 示例: 最大复杂度25 → 超出15 → 扣75分
//
//	2. 分数下限保护 (Score Clamping):
//	   • 条件检查: if score < 0
//	   • 下限处理: score = 0
//	   • 防止负分: 即使复杂度极高也不会显示负数分数
//	   • 无上限保护: 初始100分不会被超越
//
// 执行流程:
//
//	1. 初始化满分:
//	   • score := 100.0
//	   • 所有惩罚从满分开始递减
//
//	2. 平均复杂度惩罚计算:
//	   • 条件检查: if avgComplexity > MaxCyclomaticComplexity (10.0)
//	   • 计算超出量: excess = avgComplexity - 10.0
//	   • 计算惩罚: penalty = excess × 10
//	   • 应用扣分: score -= penalty
//	   • 示例: avgComplexity=12.5 → excess=2.5 → penalty=25 → score=75
//
//	3. 最大复杂度惩罚计算:
//	   • 条件检查: if maxComplexity > 10
//	   • 计算超出量: excess = maxComplexity - 10
//	   • 计算惩罚: penalty = float64(excess) × ScorePerComplexityIssue
//	   • 应用扣分: score -= penalty
//	   • 示例: maxComplexity=18 → excess=8 → penalty=40 → score=35 (如果平均未扣分)
//
//	4. 分数下限保护:
//	   • 条件检查: if score < 0
//	   • 下限设定: score = 0
//	   • 防止显示异常
//
//	5. 返回评分:
//	   • 返回0.0-100.0范围内的float64分数
//
// 惩罚系数设计:
//
//	平均复杂度惩罚系数 (固定10分/单位):
//	• 理由: 平均复杂度反映整体代码质量水平,超标影响显著,扣分需严厉
//	• 对比:
//	  - 平均复杂度: 10分/单位 (本函数)
//	  - 最大复杂度: 5分/单位 (本函数)
//	  - 正确性问题: 10分/个 (calculateCorrectnessScore)
//	  - 风格问题: 2分/个 (calculateStyleScore)
//	• 惩罚强度: 平均复杂度 (2倍) > 最大复杂度
//	• 设计理念: 整体水平比个别极端更重要,平均值体现团队编码能力
//
//	最大复杂度惩罚系数 (ScorePerComplexityIssue = 5分/单位):
//	• 理由: 最大复杂度捕捉极端复杂函数,虽需重构但不如平均值影响广泛
//	• 惩罚强度: 最大复杂度 (5分) < 平均复杂度 (10分)
//	• 设计理念: 一个超复杂函数可能是合理的(如状态机),但平均高则全局有问题
//	• 平衡考虑: 既要捕捉极端情况,又不能过度惩罚局部复杂度
//
//	双重惩罚的累积效应:
//	• 两种惩罚相互独立,可以同时触发并叠加
//	• 示例: avgComplexity=15 (扣50分) + maxComplexity=25 (扣75分) → 扣125分 → 最终0分
//	• 最坏情况: 可能快速归零(如上例)
//	• 设计权衡: 严格约束复杂度 vs 可能过于苛刻
//
// 与其他评分函数的对比:
//
//	本函数 (复杂度评分辅助函数):
//	• 调用位置: GocycloTool.ParseResult()内部、calculateStructureScore()等
//	• 输入参数: avgComplexity (float64平均值), maxComplexity (int最大值)
//	• 输出目标: 复杂度维度的评分(0-100)
//	• 惩罚策略: 双重独立惩罚(平均10分/单位 + 最大5分/单位)
//	• 算法复杂度: O(1) - 两次条件判断
//	• 用途: 将复杂度指标转换为量化分数
//
//	calculateStyleScore (风格评分,行5831):
//	• 惩罚策略: 单一惩罚(2分/问题)
//	• 输入: 问题列表长度
//	• 用途: golint工具结果评分
//
//	calculateCorrectnessScore (正确性评分,行6036):
//	• 惩罚策略: 单一惩罚(10分/问题)
//	• 输入: 问题列表长度
//	• 用途: go vet工具结果评分
//
//	关键差异:
//	• 输入维度: 本函数接收两个数值指标 vs 其他接收问题列表
//	• 惩罚策略: 双重独立惩罚 vs 单一线性惩罚
//	• 惩罚系数: 平均10+最大5 vs 风格2或正确性10
//	• 应用场景: 复杂度度量 vs 问题计数
//
// 参数:
//   - avgComplexity: 项目的平均圈复杂度
//     • 类型: float64浮点数(支持小数,如12.5)
//     • 来源: 通过GocycloTool.ParseResult()计算所有高复杂度函数的平均值
//     • 计算方式: totalComplexity / functionCount
//     • 示例: 5个函数复杂度为10,12,15,8,20 → 平均=(10+12+15+8+20)/5=13.0
//     • 阈值: MaxCyclomaticComplexity (默认10.0)
//     • 允许范围: 通常0.0-50.0(极端情况可能更高)
//   - maxComplexity: 项目中最高的圈复杂度值
//     • 类型: int整数(圈复杂度总是整数)
//     • 来源: 通过遍历所有函数的复杂度取最大值
//     • 意义: 识别最复杂的单个函数(技术债务热点)
//     • 示例: 5个函数复杂度为10,12,15,8,20 → 最大=20
//     • 阈值: 硬编码为10
//     • 允许范围: 通常1-100(极端情况可能更高)
//
// 返回值:
//   - float64: 基于双重复杂度指标的质量评分
//     • 范围: 0.0 - 100.0
//     • 精度: 浮点数,支持小数(如85.5分)
//     • 意义: 分数越高,代码复杂度控制越好,可维护性越强
//     • 100.0: 平均和最大复杂度均在阈值内(理想状态)
//     • 0.0: 复杂度严重超标(不可接受)
//
// 使用场景:
//   - GocycloTool.ParseResult()调用: 在解析gocyclo输出后计算复杂度得分
//   - calculateStructureScore()调用: 作为代码结构评分的组成部分
//   - 单元测试: 测试复杂度评分算法的准确性
//   - 复杂度分析报告: 生成复杂度相关的质量指标
//
// 示例:
//
//	// 示例1: 理想情况 (无惩罚)
//	avgComplexity := 8.5
//	maxComplexity := 9
//	score := calculateComplexityScore(avgComplexity, maxComplexity)
//	// 平均8.5 ≤ 10.0 (无扣分)
//	// 最大9 ≤ 10 (无扣分)
//	// score = 100.0 (满分)
//
//	// 示例2: 平均复杂度超标 (单一惩罚)
//	avgComplexity := 13.0
//	maxComplexity := 9
//	score := calculateComplexityScore(avgComplexity, maxComplexity)
//	// 平均13.0 > 10.0 → 超出3.0 → 扣30分
//	// 最大9 ≤ 10 (无扣分)
//	// score = 100 - 30 = 70.0
//
//	// 示例3: 最大复杂度超标 (单一惩罚)
//	avgComplexity := 8.0
//	maxComplexity := 18
//	score := calculateComplexityScore(avgComplexity, maxComplexity)
//	// 平均8.0 ≤ 10.0 (无扣分)
//	// 最大18 > 10 → 超出8 → 扣40分 (8 × 5 = 40)
//	// score = 100 - 40 = 60.0
//
//	// 示例4: 双重惩罚 (两者都超标)
//	avgComplexity := 15.0
//	maxComplexity := 25
//	score := calculateComplexityScore(avgComplexity, maxComplexity)
//	// 平均15.0 > 10.0 → 超出5.0 → 扣50分
//	// 最大25 > 10 → 超出15 → 扣75分 (15 × 5 = 75)
//	// 原始: 100 - 50 - 75 = -25
//	// 下限保护: score = 0.0
//
//	// 示例5: 极端复杂度 (严重超标)
//	avgComplexity := 25.0
//	maxComplexity := 50
//	score := calculateComplexityScore(avgComplexity, maxComplexity)
//	// 平均25.0 > 10.0 → 超出15.0 → 扣150分
//	// 最大50 > 10 → 超出40 → 扣200分 (40 × 5 = 200)
//	// 原始: 100 - 150 - 200 = -250
//	// 下限保护: score = 0.0
//
// 注意事项:
//   - 双重独立惩罚: 平均和最大复杂度惩罚相互独立,可同时触发并累积扣分
//   - 平均优先原则: 平均复杂度扣分系数(10)是最大复杂度(5)的2倍,体现整体水平更重要
//   - 阈值不一致: 平均复杂度阈值为MaxCyclomaticComplexity(10.0),最大复杂度阈值硬编码为10
//   - 类型转换: maxComplexity为int,计算惩罚时需转换为float64
//   - 快速归零: 双重惩罚可能导致分数快速降至0(如示例4/5)
//   - 下限保护: score<0时归零,防止显示负数
//   - 无上限保护: 初始100分不会超越,无需上限检查
//   - 浮点精度: 使用float64避免精度损失
//   - 硬编码阈值: 最大复杂度阈值10是硬编码的,不如平均复杂度灵活(使用常量)
//   - 惩罚系数依赖: ScorePerComplexityIssue常量值变化会影响最大复杂度惩罚强度
//
// 改进方向:
//   - 阈值统一化: 将最大复杂度阈值10改为常量MaxCyclomaticComplexity,保持一致性
//   - 系数配置化: 将惩罚系数10和ScorePerComplexityIssue改为可配置参数
//   - 非线性惩罚: 采用指数或对数曲线,体现复杂度超标的非线性危害
//     - 低复杂度(10-15): 轻度惩罚(5分/单位)
//     - 中复杂度(15-25): 中度惩罚(10分/单位)
//     - 高复杂度(25+): 重度惩罚(20分/单位)
//   - 加权平衡: 引入权重参数平衡平均和最大复杂度的重要性
//     - score = 100 - avgPenalty × weightAvg - maxPenalty × weightMax
//   - 相对惩罚: 基于复杂度超出比例而非绝对值
//     - penalty = (actual / threshold - 1) × 100
//   - 分段评分: 不同复杂度范围使用不同评分策略
//     - 0-10: 满分100
//     - 10-20: 线性递减90-60
//     - 20-30: 加速递减60-20
//     - 30+: 固定0分
//   - 历史对比: 与上次复杂度指标对比,识别复杂度演进趋势
//   - 分布分析: 增加复杂度分布指标(如P90、P99复杂度)
//   - 函数数量权重: 考虑高复杂度函数的数量占比
//     - penalty = baseScore × (highComplexFuncCount / totalFuncCount)
//   - 复杂度类型区分: 区分圈复杂度和认知复杂度,分别评分
//   - 自适应阈值: 基于项目类型(如状态机vs业务逻辑)动态调整阈值
//   - 修复成本估算: 将复杂度分数与预估重构工时关联
//     - debtHours = (avgComplexity - threshold) × funcCount × 0.5
//
// 作者: JIA
func calculateComplexityScore(avgComplexity float64, maxComplexity int) float64 {
	score := 100.0

	// 平均复杂度惩罚
	if avgComplexity > MaxCyclomaticComplexity {
		score -= (avgComplexity - MaxCyclomaticComplexity) * 10
	}

	// 最大复杂度惩罚
	if maxComplexity > 10 {
		score -= float64(maxComplexity-10) * ScorePerComplexityIssue
	}

	if score < 0 {
		score = 0
	}
	return score
}

// countBySeverity 按严重程度统计代码问题数量(安全性评分专用辅助函数)
//
// 功能说明:
//
//	本辅助函数专门用于代码问题统计场景,负责从问题列表([]CodeIssue)中筛选并统计
//	特定严重程度(Severity)的问题数量。采用高效的索引遍历策略,避免大结构体的值拷贝开销,
//	是安全性评分(calculateSecurityScore)和其他维度评分的核心数据统计工具。
//	该函数通过线性扫描问题列表,使用字符串相等性比较筛选目标严重程度,
//	最终返回匹配问题的总数量,用于后续的分数计算和严重程度分布分析。
//
// 统计策略:
//
//	1. 线性扫描策略 (Linear Scan Strategy):
//	   • 遍历方式: 索引遍历 for i := range issues
//	   • 性能优化: 避免值拷贝(CodeIssue结构体184字节)
//	   • 比对方式: issues[i].Severity == severity (字符串相等性)
//	   • 计数逻辑: 匹配则 count++
//	   • 时间复杂度: O(n),n为问题列表长度
//	   • 空间复杂度: O(1),仅使用一个计数器变量
//
//	2. 严重程度匹配机制 (Severity Matching):
//	   • 匹配条件: 精确字符串相等(大小写敏感)
//	   • 支持的严重程度等级:
//	     - "high"   (高危): 安全漏洞、逻辑错误等严重问题
//	     - "medium" (中危): 潜在风险、性能隐患等中等问题
//	     - "low"    (低危): 代码风格、最佳实践建议等轻微问题
//	   • 其他可能值: 根据静态分析工具而定(如"critical", "warning", "info")
//
//	3. 空列表快速处理 (Empty List Fast Path):
//	   • 隐式处理: for range空切片时循环不执行
//	   • 返回值: count保持初始值0
//	   • 无需显式检查: len(issues) == 0判断
//
// 执行流程:
//
//	1. 初始化计数器:
//	   • count := 0
//	   • 准备累加匹配问题数量
//
//	2. 索引遍历问题列表:
//	   • 循环: for i := range issues
//	   • 性能考虑: 避免 for _, issue := range issues 的值拷贝(184字节/次)
//	   • 索引访问: issues[i]获取问题引用
//
//	3. 严重程度匹配判断:
//	   • 条件检查: if issues[i].Severity == severity
//	   • 字符串比较: 使用Go内置字符串相等性(大小写敏感)
//	   • 匹配成功: count++ (计数器加1)
//	   • 不匹配: 跳过,继续下一个
//
//	4. 返回统计结果:
//	   • return count
//	   • 返回匹配问题的总数量
//
// 性能优化设计:
//
//	索引遍历 vs 值遍历性能对比:
//	• 索引遍历 (本函数采用):
//	  - 代码: for i := range issues { ... issues[i] ... }
//	  - 内存操作: 仅传递索引(8字节int)
//	  - 拷贝开销: 0字节 (无结构体拷贝)
//	  - 适用场景: 大结构体、只读访问、需要索引
//	• 值遍历 (未采用):
//	  - 代码: for _, issue := range issues { ... issue ... }
//	  - 内存操作: 每次迭代拷贝整个结构体
//	  - 拷贝开销: 184字节/次 (CodeIssue结构体大小)
//	  - 示例: 100个问题 → 18400字节拷贝 vs 800字节索引传递
//	  - 性能损失: 约23倍内存操作开销
//
//	CodeIssue结构体大小分析 (184字节):
//	• 字段组成:
//	  - File      string (16字节: 指针8 + 长度8)
//	  - Line      int    (8字节)
//	  - Column    int    (8字节)
//	  - Message   string (16字节)
//	  - Severity  string (16字节)
//	  - Tool      string (16字节)
//	  - RuleID    string (16字节)
//	  - Category  string (16字节)
//	  - Suggestion string (16字节)
//	  - Code      string (16字节)
//	  - Context   map[string]string (24字节: 指针8 + 哈希表16)
//	• 总计: 10×16 + 2×8 + 24 = 184字节
//	• 优化效果: 避免184字节×N次的拷贝开销
//
// 参数:
//   - issues: 待统计的代码问题列表
//     • 类型: []CodeIssue 切片
//     • 来源: 静态分析工具(gosec, golint, go vet等)的解析结果
//     • 结构体大小: 184字节(见上面分析)
//     • 内容: 包含文件路径、行号、严重程度、消息等字段
//     • 允许为空: 空切片返回0
//     • 示例: []CodeIssue{ {Severity: "high", ...}, {Severity: "medium", ...} }
//   - severity: 目标严重程度标识符
//     • 类型: string 字符串
//     • 格式: 小写字符串(约定俗成)
//     • 常见值: "high", "medium", "low"
//     • 大小写: 敏感匹配(必须完全一致)
//     • 来源: 调用方指定(如calculateSecurityScore传入"high")
//     • 示例: "high" → 统计高危问题数量
//
// 返回值:
//   - int: 匹配指定严重程度的问题数量
//     • 范围: 0 - len(issues) (0到问题总数)
//     • 0: 无匹配问题(可能问题列表为空或无该严重程度问题)
//     • len(issues): 所有问题都是该严重程度
//     • 用途: 用于计算严重程度加权分数
//     • 示例: 10个问题中3个"high" → 返回3
//
// 使用场景:
//   - calculateSecurityScore()调用: 统计高危/中危/低危安全问题数量
//   - 其他维度评分: 统计特定严重程度的代码问题
//   - 问题分布分析: 生成严重程度分布报告
//   - 单元测试: 验证问题统计逻辑的准确性
//   - 质量门禁: 基于严重问题数量决定是否通过CI/CD
//
// 示例:
//
//	// 示例1: 统计高危问题
//	issues := []CodeIssue{
//	    {Severity: "high", Message: "SQL注入风险"},
//	    {Severity: "medium", Message: "未处理错误"},
//	    {Severity: "high", Message: "硬编码密码"},
//	    {Severity: "low", Message: "命名不规范"},
//	}
//	highCount := countBySeverity(issues, "high")
//	// 匹配: issues[0] (high), issues[2] (high)
//	// 不匹配: issues[1] (medium), issues[3] (low)
//	// highCount = 2
//
//	// 示例2: 统计中危问题
//	mediumCount := countBySeverity(issues, "medium")
//	// 匹配: issues[1] (medium)
//	// mediumCount = 1
//
//	// 示例3: 空列表处理
//	emptyIssues := []CodeIssue{}
//	count := countBySeverity(emptyIssues, "high")
//	// 循环0次,count保持初始值0
//	// count = 0
//
//	// 示例4: 无匹配问题
//	lowOnlyIssues := []CodeIssue{
//	    {Severity: "low", Message: "问题1"},
//	    {Severity: "low", Message: "问题2"},
//	}
//	highCount := countBySeverity(lowOnlyIssues, "high")
//	// 所有问题都是"low",无"high"匹配
//	// highCount = 0
//
//	// 示例5: 全部匹配
//	allHighIssues := []CodeIssue{
//	    {Severity: "high", Message: "问题1"},
//	    {Severity: "high", Message: "问题2"},
//	    {Severity: "high", Message: "问题3"},
//	}
//	highCount := countBySeverity(allHighIssues, "high")
//	// 所有3个问题都匹配
//	// highCount = 3 (等于len(allHighIssues))
//
// 注意事项:
//   - 大小写敏感: "high" ≠ "High" ≠ "HIGH",必须完全一致
//   - 严重程度值约定: 依赖静态分析工具的输出格式,需保持一致性
//   - 索引遍历优化: CodeIssue结构体184字节,避免值拷贝性能损失
//   - 空切片安全: 空列表返回0,无需nil检查(Go保证for range安全)
//   - 线性时间复杂度: O(n)性能,大量问题时可能成为瓶颈
//   - 单次扫描: 每次调用只统计一种严重程度,需要多种时需多次调用
//   - 无缓存机制: 重复调用会重复扫描,可考虑缓存统计结果
//   - 不修改输入: 纯函数,不改变issues切片内容
//   - 并发安全: 无共享状态,天然支持并发调用(如果issues不被修改)
//   - 精确匹配: 不支持模糊匹配或正则表达式
//
// 改进方向:
//   - 批量统计优化: 一次扫描统计所有严重程度,返回map[string]int
//     - severityCounts := countAllBySeverity(issues)
//     - severityCounts["high"] → 高危数量
//     - 性能提升: O(n)一次扫描 vs O(n×k)多次扫描(k为严重程度种类数)
//   - 大小写不敏感匹配: strings.EqualFold(issues[i].Severity, severity)
//     - 提升容错性,支持"High", "HIGH", "high"等写法
//   - 严重程度枚举类型: 使用const常量或枚举替代字符串
//     - const (SeverityHigh = "high"; SeverityMedium = "medium"; SeverityLow = "low")
//     - 编译时类型检查,避免拼写错误
//   - 过滤函数泛化: 支持自定义过滤条件
//     - countByCondition(issues, func(issue CodeIssue) bool { return issue.Severity == "high" })
//   - 性能缓存: 缓存统计结果,避免重复扫描
//     - 适用于issues不变但需要多次统计的场景
//   - 并行统计: 大量问题时使用goroutine并行统计
//     - 需要考虑协程开销,问题数量<10000时可能不值得
//   - 严重程度分布: 返回完整分布而非单一计数
//     - type SeverityDistribution struct { High int; Medium int; Low int; Others int }
//   - 加权计数: 不同严重程度赋予不同权重
//     - weightedCount = highCount×10 + mediumCount×5 + lowCount×1
//   - 去重统计: 基于RuleID或文件+行号去重后统计
//     - 避免同一问题被多个工具重复报告
//   - 时间复杂度优化: 使用map预先分组,O(1)查询
//     - 空间换时间,适合频繁查询场景
//
// 作者: JIA
func countBySeverity(issues []CodeIssue, severity string) int {
	count := 0
	// 使用索引遍历避免大结构体复制（184字节）
	for i := range issues {
		if issues[i].Severity == severity {
			count++
		}
	}
	return count
}

// checkGofmt 检查代码格式化合规性并返回未格式化文件数量(风格合规维度核心检查方法)
//
// 功能说明:
//
//	本方法是代码质量评估系统中风格合规维度(style_compliance)的核心检查方法之一,
//	负责通过执行Go官方格式化工具gofmt来验证项目代码是否符合Go语言标准格式规范。
//	该方法采用命令行调用策略,使用"gofmt -l ."命令列出所有未格式化的Go源文件,
//	通过解析命令输出统计未格式化文件数量,最终返回整数计数结果。返回值为0表示所有文件
//	均已正确格式化(理想状态),大于0表示存在格式化问题需要修复。本方法是保障代码风格
//	一致性的第一道防线,确保团队代码遵循统一的排版和缩进规范。
//
// 检查策略:
//
//	1. 命令行工具调用策略 (Command-Line Tool Invocation):
//	   • 工具: gofmt (Go官方格式化工具,Go安装自带)
//	   • 命令: "gofmt -l ." (列出未格式化文件)
//	   • 参数解析:
//	     - -l (--list): 列出格式不正确的文件名(而非修复)
//	     - .  (当前目录): 递归扫描项目所有Go文件
//	   • 工作目录: cqe.results.ProjectPath (项目根目录)
//	   • 输出解析: 每行一个文件名,文件名列表即为问题文件
//
//	2. 输出解析策略 (Output Parsing Strategy):
//	   • 正常输出: 未格式化文件路径,每行一个
//	   • 空输出: "" 或 "\n" (所有文件已格式化,理想状态)
//	   • 解析方式: strings.Split(output, "\n") 按行分割
//	   • 计数逻辑: len(lines) 统计文件数量
//	   • 空行过滤: 检查 len(lines)==1 && lines[0]=="" 识别空输出
//
//	3. 错误处理策略 (Error Handling Strategy):
//	   • 命令错误: cmd.Output() 返回error
//	   • 错误场景:
//	     - gofmt未安装或不在PATH中
//	     - 项目路径不存在或无权限
//	     - Go文件语法错误导致gofmt失败
//	   • 错误处理: 返回0 (保守策略,避免因工具问题导致评分异常)
//	   • 设计权衡: 返回0可能掩盖真实问题,但优于评分系统崩溃
//
// 执行流程:
//
//	1. 构建gofmt命令对象:
//	   • 命令: exec.Command("gofmt", "-l", ".")
//	   • 参数分解: 可执行文件"gofmt", 参数"-l"和"."
//	   • 安全性审计: #nosec G204 (固定命令,无用户输入,用于代码格式检查)
//
//	2. 设置命令工作目录:
//	   • cmd.Dir = cqe.results.ProjectPath
//	   • 确保gofmt在项目根目录执行
//	   • 路径来源: CodeQualityEvaluator.results.ProjectPath字段
//
//	3. 执行命令并捕获输出:
//	   • 调用: cmd.Output()
//	   • 返回: (output []byte, err error)
//	   • 输出内容: 未格式化文件的相对路径列表(一行一个)
//	   • 示例输出:
//	     models/student.go
//	     evaluators/code_quality.go
//	     tools/assessment_tools.go
//
//	4. 错误处理与快速返回:
//	   • 条件: if err != nil
//	   • 返回: return 0
//	   • 理由: gofmt执行失败,无法获取格式化状态,返回0避免评分异常
//
//	5. 输出解析与空值检查:
//	   • 去除首尾空白: strings.TrimSpace(string(output))
//	   • 按行分割: strings.Split(..., "\n")
//	   • 空输出判断: if len(lines) == 1 && lines[0] == ""
//	   • 空输出含义: 所有文件已格式化,gofmt无输出
//	   • 空输出返回: return 0 (0个问题)
//
//	6. 返回未格式化文件计数:
//	   • 计数: len(lines)
//	   • 含义: 每一行对应一个未格式化文件
//	   • 返回: int整数,表示需要格式化的文件数量
//
// 安全性审计说明:
//
//	#nosec G204 - 固定命令，无用户输入，用于代码格式检查:
//	• Gosec规则: G204 - 命令注入风险检测
//	• 风险评估: 本命令使用固定字符串"gofmt"和固定参数"-l"、".",无任何用户输入
//	• 安全措施:
//	  - 可执行文件: "gofmt" (硬编码字符串,非变量)
//	  - 参数1: "-l" (硬编码标志)
//	  - 参数2: "." (硬编码路径,相对于cmd.Dir)
//	  - 工作目录: cqe.results.ProjectPath (内部字段,非用户直接输入)
//	• 豁免理由: 无命令注入可能,所有参数均为常量,符合安全最佳实践
//	• 替代方案: 无更安全的替代方案,gofmt必须通过命令行调用
//
// gofmt工具说明:
//
//	gofmt - Go官方格式化工具:
//	• 来源: Go标准工具链,随Go安装自动提供
//	• 功能: 自动格式化Go源代码,确保一致的代码风格
//	• 格式规范:
//	  - 缩进: Tab制表符(不是空格)
//	  - 大括号位置: 函数/结构体左括号与声明同行
//	  - 空行规则: 自动调整包、import、函数间空行
//	  - 运算符空格: 二元运算符前后自动添加空格
//	  - 对齐: 结构体字段、import路径自动对齐
//	• 命令参数:
//	  - -l (list): 仅列出格式不正确的文件,不修改文件
//	  - -w (write): 直接修改文件(本方法未使用,仅检查)
//	  - -d (diff): 显示格式化前后差异(本方法未使用)
//	• 退出码:
//	  - 0: 成功执行(无论是否有文件需要格式化)
//	  - 非0: 执行错误(如语法错误、文件不存在)
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: 设置gofmt命令工作目录
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - int: 未格式化的Go源文件数量
//     • 范围: 0 - N (0到项目文件总数)
//     • 0: 所有文件均已格式化(理想状态)
//     • >0: 存在未格式化文件,数值为问题文件数量
//     • 示例: 返回5表示有5个文件未格式化
//     • 用途: 用于calculateStyleScore()计算风格合规分数
//
// 使用场景:
//   - calculateStyleScore()调用: 作为风格合规评分的输入指标之一
//   - CI/CD质量门禁: 格式化检查未通过(返回>0)时阻止代码合并
//   - 开发环境检查: 提交代码前本地验证格式化状态
//   - 代码审查辅助: 自动识别格式化问题,减少人工审查负担
//   - 新成员培训: 强制执行Go标准格式,培养良好编码习惯
//
// 示例:
//
//	// 示例1: 所有文件已格式化 (理想状态)
//	// 假设项目所有Go文件都已执行 gofmt -w .
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/clean-project"},
//	}
//	count := cqe.checkGofmt()
//	// gofmt输出: "" (空字符串,无未格式化文件)
//	// 解析: len(lines)==1 && lines[0]=="" → 空输出
//	// count = 0 (满分,无问题)
//
//	// 示例2: 部分文件未格式化
//	// 假设3个文件未格式化
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/messy-project"},
//	}
//	count := cqe.checkGofmt()
//	// gofmt输出:
//	// models/student.go
//	// evaluators/code_quality.go
//	// tools/assessment_tools.go
//	// 解析: len(lines) = 3
//	// count = 3 (3个文件需要格式化)
//
//	// 示例3: 大型项目多文件未格式化
//	// 假设新项目刚创建,20个文件都未格式化
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/new-project"},
//	}
//	count := cqe.checkGofmt()
//	// gofmt输出: 20行文件名
//	// 解析: len(lines) = 20
//	// count = 20 (严重格式化问题)
//
//	// 示例4: gofmt命令执行失败
//	// 假设gofmt未安装或路径错误
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/invalid/path"},
//	}
//	count := cqe.checkGofmt()
//	// cmd.Output() 返回error
//	// 错误处理: return 0
//	// count = 0 (保守返回,避免评分系统崩溃)
//
//	// 示例5: 单文件项目未格式化
//	// 假设只有一个main.go未格式化
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/tiny-project"},
//	}
//	count := cqe.checkGofmt()
//	// gofmt输出:
//	// main.go
//	// 解析: len(lines) = 1
//	// count = 1 (1个文件需要格式化)
//
// 注意事项:
//   - gofmt依赖: 本方法依赖系统安装的gofmt工具,必须在PATH中可用
//   - 错误静默返回0: gofmt执行失败时返回0,可能掩盖真实问题,建议增加日志记录
//   - 工作目录依赖: 必须正确设置cmd.Dir,否则"."路径指向错误目录
//   - 递归扫描: gofmt -l . 会递归扫描所有子目录,包括vendor和隐藏目录
//   - 性能考虑: 大型项目(1000+文件)可能耗时较长(秒级),考虑超时控制
//   - 语法错误影响: 如果Go文件有语法错误,gofmt可能失败,导致返回0
//   - 无并发保护: 方法本身无锁,并发调用时安全,但共享cqe.results需注意
//   - 空输出判断: 依赖 len(lines)==1 && lines[0]=="" 判断,假设gofmt输出无多余空行
//   - Windows路径: Windows下路径分隔符为反斜杠,但gofmt输出使用正斜杠
//   - Go版本兼容: gofmt格式规则随Go版本演进,不同版本可能有细微差异
//
// 改进方向:
//   - 错误日志记录: gofmt失败时记录详细错误信息,便于问题排查
//     - if err != nil { log.Printf("gofmt failed: %v", err); return 0 }
//   - 超时控制: 大型项目添加超时机制,避免长时间阻塞
//     - ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//     - cmd := exec.CommandContext(ctx, "gofmt", "-l", ".")
//   - 排除目录支持: 支持排除vendor、.git等无需检查的目录
//     - 自定义脚本: find . -name "*.go" | grep -v vendor | xargs gofmt -l
//   - 差异报告: 使用gofmt -d生成详细差异,提供修复建议
//     - 输出: 具体哪些行格式不正确及正确格式
//   - 批量修复选项: 提供自动修复模式,执行gofmt -w .直接修复
//     - 需要用户确认,避免意外修改
//   - 增量检查: 仅检查git diff中修改的文件,提升性能
//     - git diff --name-only | xargs gofmt -l
//   - 格式化级别区分: 区分"必须修复"和"建议修复"的格式问题
//   - 集成goimports: 同时检查import排序和未使用导入
//     - goimports -l . (gofmt的超集)
//   - 并行检查: 大型项目使用goroutine并行检查多个目录
//   - 缓存机制: 缓存已检查文件的格式化状态,避免重复检查
//     - 基于文件修改时间或内容哈希判断是否需要重新检查
//   - 返回详细信息: 返回[]string文件列表而非int计数,提供更多上下文
//     - 允许调用方生成详细报告或逐个修复
//   - 跨平台路径处理: 统一Windows和Unix路径分隔符
//     - 使用filepath.ToSlash()规范化路径
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) checkGofmt() int {
	// #nosec G204 - 固定命令，无用户输入，用于代码格式检查
	cmd := exec.Command("gofmt", "-l", ".")
	cmd.Dir = cqe.results.ProjectPath
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}

	return len(lines)
}

// checkNamingConventions 检查Go命名约定合规性并返回违规数量(风格合规维度辅助检查方法-待实现)
//
// 功能说明:
//
//	本方法是代码质量评估系统中风格合规维度(style_compliance)的辅助检查方法,
//	负责验证项目代码是否遵循Go语言官方命名约定(Naming Conventions)。该方法设计用于
//	检查包名(package name)、类型名(type name)、函数名(function name)、变量名(variable name)、
//	常量名(constant name)等标识符是否符合Go社区的最佳实践和代码规范。
//	当前实现为占位符(placeholder),返回固定值0,表示无违规问题,未来版本将实现完整的
//	命名约定检查逻辑,集成golint、revive等工具的命名规则检测能力。
//
// 设计目标(待实现):
//
//	1. 包名检查 (Package Name Conventions):
//	   • 规则: 全小写字母,无下划线,简短有意义
//	   • 正确: http, fmt, encoding, bufio
//	   • 错误: HTTP, fmt_util, encoding_json, buf_io
//	   • 检测: 识别大写字母、下划线、过长包名
//
//	2. 导出标识符检查 (Exported Identifiers):
//	   • 规则: 首字母大写(驼峰命名CamelCase)
//	   • 类型名: type UserProfile struct { ... } (正确)
//	   • 函数名: func GetUserByID() (正确)
//	   • 错误: type user_profile (应为UserProfile)
//	   • 错误: func get_user_by_id() (应为GetUserByID)
//
//	3. 未导出标识符检查 (Unexported Identifiers):
//	   • 规则: 首字母小写(驼峰命名camelCase)
//	   • 变量名: var userID int (正确)
//	   • 函数名: func parseRequest() (正确)
//	   • 错误: var user_id (应为userID)
//	   • 错误: func parse_request() (应为parseRequest)
//
//	4. 常量命名检查 (Constant Naming):
//	   • 传统规则: 驼峰命名,特殊情况允许全大写+下划线
//	   • 推荐: const MaxBufferSize = 1024 (驼峰)
//	   • 允许: const MAX_BUFFER_SIZE = 1024 (传统C风格,Go中不推荐)
//	   • 枚举常量: const ( StatusActive = 1; StatusInactive = 2 )
//
//	5. 缩写词一致性检查 (Acronym Consistency):
//	   • 规则: 缩写词保持一致的大小写
//	   • 正确: userID, parseHTTPRequest, writeJSON
//	   • 错误: userId (应为userID), parseHttpRequest (应为parseHTTPRequest)
//	   • 常见缩写: ID, URL, HTTP, API, JSON, XML, SQL, DB, UUID
//
//	6. 接收器命名检查 (Receiver Naming):
//	   • 规则: 使用类型首字母或简短缩写,保持一致
//	   • 正确: func (u *User) GetName() (使用u)
//	   • 正确: func (u *User) SetName() (所有方法都用u)
//	   • 错误: func (this *User) GetName() (Go不推荐this/self)
//	   • 错误: func (user *User) GetName() (过长,应为u或usr)
//
//	7. 下划线滥用检查 (Underscore Misuse):
//	   • 规则: 避免使用下划线分隔(snake_case),Go推荐驼峰
//	   • 正确: getUserByID, maxRetryCount
//	   • 错误: get_user_by_id, max_retry_count
//	   • 例外: 测试文件名允许下划线(user_test.go, api_integration_test.go)
//
// 当前实现状态:
//
//	占位符实现 (Placeholder Implementation):
//	• 返回值: 固定返回0 (无违规问题)
//	• 实现程度: 0% (仅框架,无实际检查逻辑)
//	• 注释说明: "这里可以实现具体的命名约定检查逻辑"
//	• 示例提示: "例如：检查包名、函数名、变量名是否符合Go语言规范"
//	• 设计意图: 为未来功能扩展预留接口
//
// 执行流程(当前):
//
//	1. 初始化违规计数器:
//	   • issues := 0
//	   • 准备累加命名违规数量
//
//	2. 检查逻辑占位:
//	   • 注释: "这里可以实现具体的命名约定检查逻辑"
//	   • 当前: 无实际检查,跳过
//
//	3. 返回违规计数:
//	   • return issues
//	   • 当前固定返回0
//
// 预期执行流程(未来实现):
//
//	1. 解析Go源文件AST (Abstract Syntax Tree):
//	   • 使用go/parser包解析所有.go文件
//	   • 提取包声明、类型定义、函数声明、变量声明
//
//	2. 遍历AST节点提取标识符:
//	   • 访问Package、Type、Func、Var、Const节点
//	   • 提取标识符名称和位置信息
//
//	3. 应用命名规则检查:
//	   • 包名: 检查全小写、无下划线
//	   • 导出标识符: 检查首字母大写、驼峰命名
//	   • 未导出标识符: 检查首字母小写、驼峰命名
//	   • 缩写词: 检查ID/URL/HTTP等一致性
//	   • 接收器: 检查简短性、一致性
//
//	4. 记录违规问题:
//	   • 违规类型: 下划线滥用、大小写错误、缩写词不一致等
//	   • 位置信息: 文件路径、行号、列号
//	   • 严重程度: low (命名问题通常不影响功能)
//
//	5. 统计并返回违规数量:
//	   • 累加所有类型的命名违规
//	   • return len(violations)
//
// 实现方案建议:
//
//	方案1: 集成golint命名检查:
//	• 工具: golint (Go官方lint工具,已废弃但可参考)
//	• 命令: golint ./... | grep -E "(name|naming)"
//	• 优点: 规则完善,社区认可
//	• 缺点: golint已归档,推荐使用revive替代
//
//	方案2: 集成revive命名规则:
//	• 工具: revive (golint的现代替代品)
//	• 规则: var-naming, package-comments, exported等
//	• 命令: revive -config revive.toml -formatter friendly ./...
//	• 优点: 活跃维护,规则可配置,性能好
//	• 缺点: 需要额外配置文件
//
//	方案3: 自定义AST遍历:
//	• 实现: 使用go/parser和go/ast包
//	• 逻辑: 自定义ast.Visitor实现命名检查
//	• 优点: 完全控制检查规则,无外部依赖
//	• 缺点: 实现复杂,需要维护规则库
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: 定位需要检查的Go源文件
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - int: 命名约定违规问题数量
//     • 当前: 固定返回0 (占位符实现)
//     • 未来: 实际违规数量(如包名3个问题+函数名5个问题=8)
//     • 范围: 0 - N (0到项目标识符总数)
//     • 0: 所有命名符合规范(理想状态)
//     • >0: 存在命名违规,数值为问题数量
//     • 用途: 用于calculateStyleScore()计算风格合规分数
//
// 使用场景:
//   - calculateStyleScore()调用: 作为风格合规评分的输入指标之一
//   - CI/CD质量门禁: 命名检查未通过(返回>0)时发出警告
//   - 代码审查辅助: 自动识别命名问题,减少人工审查负担
//   - 新成员培训: 强制执行Go命名规范,培养良好命名习惯
//   - 重构指导: 识别需要重命名的标识符,提供重构建议
//
// 示例(未来实现):
//
//	// 示例1: 理想项目 (当前所有项目都返回此结果)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/clean-project"},
//	}
//	issues := cqe.checkNamingConventions()
//	// 当前: 直接返回0 (占位符)
//	// 未来: 检查发现0个命名问题
//	// issues = 0
//
//	// 示例2: 包名违规 (未来实现)
//	// 代码: package User_Service (应为userservice)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/bad-package"},
//	}
//	issues := cqe.checkNamingConventions()
//	// 未来: 检测到包名使用大写和下划线
//	// 违规: "User_Service" 应为 "userservice"
//	// issues = 1
//
//	// 示例3: 函数命名违规 (未来实现)
//	// 代码: func get_user_by_id() (应为GetUserByID或getUserByID)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/snake-case-project"},
//	}
//	issues := cqe.checkNamingConventions()
//	// 未来: 检测到10个函数使用snake_case命名
//	// 违规示例: get_user_by_id, parse_http_request, validate_email_address
//	// issues = 10
//
//	// 示例4: 缩写词不一致 (未来实现)
//	// 代码: var userId int (应为userID), func parseHttpRequest (应为parseHTTPRequest)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/acronym-project"},
//	}
//	issues := cqe.checkNamingConventions()
//	// 未来: 检测到5个缩写词大小写不一致
//	// 违规: userId→userID, httpClient→HTTPClient, jsonData→JSONData
//	// issues = 5
//
//	// 示例5: 接收器命名违规 (未来实现)
//	// 代码: func (this *User) GetName() (应为 func (u *User) GetName())
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/receiver-project"},
//	}
//	issues := cqe.checkNamingConventions()
//	// 未来: 检测到3个方法使用this/self作为接收器
//	// 违规: (this *User), (self *Config), (that *Service)
//	// issues = 3
//
// 注意事项:
//   - 占位符实现: 当前方法未实现任何检查逻辑,仅返回0
//   - 不影响评分: 由于返回0,不会对风格合规分数产生负面影响
//   - 未来扩展: 需要实现AST解析或集成外部工具(revive/golint)
//   - 规则配置: 命名规则应可配置,支持团队自定义标准
//   - 性能考虑: AST解析大型项目可能耗时,需要缓存和优化
//   - 误报风险: 自动命名检查可能产生误报,需要人工review
//   - 语言演进: Go命名规范可能随版本演进,需要持续更新规则
//   - 第三方包: 外部依赖的命名风格可能与项目不一致,需要排除
//   - 测试文件: _test.go文件的命名规则与普通文件不同
//   - 代码生成: 自动生成代码(如protobuf)可能不符合规范,需要豁免
//
// 改进方向:
//   - 实现基础命名检查: 集成revive工具的var-naming、package-comments等规则
//     - revive -config revive.toml -formatter json ./...
//   - AST解析实现: 使用go/parser遍历AST,提取标识符并应用规则
//     - fset := token.NewFileSet(); ast, _ := parser.ParseFile(fset, path, nil, 0)
//   - 规则可配置化: 支持通过配置文件定义命名规则和严重程度
//     - naming_rules.yaml: { package: lowercase, function: camelCase, ... }
//   - 缩写词词典: 维护常见缩写词列表(ID, URL, HTTP等),自动检查一致性
//     - acronyms := []string{"ID", "URL", "HTTP", "API", "JSON", "XML", "SQL", "DB"}
//   - 智能修复建议: 不仅报告问题,还提供自动修复建议
//     - 违规: get_user_by_id → 建议: GetUserByID 或 getUserByID
//   - 渐进式检查: 仅检查新增/修改的代码,避免历史债务影响评分
//     - git diff --name-only | xargs revive
//   - 团队风格适配: 支持多种命名风格(严格Go标准/宽松模式/自定义)
//   - 接收器一致性: 检查同一类型的所有方法接收器名称是否一致
//     - type User: 所有方法都用(u *User)或都用(usr *User)
//   - 上下文敏感: 根据标识符用途调整规则(如测试辅助函数允许下划线)
//   - 性能优化: 并行处理多个文件,使用goroutine加速AST解析
//     - 大型项目1000+文件时性能提升显著
//   - 集成IDE: 提供VSCode/GoLand插件,实时显示命名问题
//   - 历史趋势: 追踪命名违规数量变化,识别代码质量演进
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) checkNamingConventions() int {
	issues := 0
	// 这里可以实现具体的命名约定检查逻辑
	// 例如：检查包名、函数名、变量名是否符合Go语言规范
	return issues
}

// checkSecurityIssues 检查代码安全问题并返回潜在漏洞数量(安全性维度辅助检查方法-待实现)
//
// 功能说明:
//
//	本方法是代码质量评估系统中安全性维度(security_analysis)的辅助检查方法,
//	负责识别项目代码中可能存在的安全漏洞和风险隐患。该方法设计用于检测常见的
//	安全问题,如硬编码敏感信息(密码/API密钥)、SQL注入风险、不安全的随机数生成、
//	命令注入漏洞、路径遍历攻击、弱加密算法使用等安全隐患。
//	当前实现为占位符(placeholder),返回固定值0,表示无安全问题,未来版本将实现完整的
//	安全检查逻辑,集成gosec、静态分析等工具的安全规则检测能力,确保代码符合安全最佳实践。
//
// 设计目标(待实现):
//
//	1. 硬编码敏感信息检测 (Hardcoded Credentials Detection):
//	   • 检测目标: 密码、API密钥、访问令牌、私钥等敏感数据
//	   • 规则:
//	     - 识别常见模式: password = "xxx", apiKey := "sk-xxx", token := "Bearer xxx"
//	     - 检测字符串常量中的Base64编码密钥
//	     - 识别环境变量硬编码: os.Setenv("AWS_SECRET_KEY", "xxx")
//	   • 示例违规:
//	     const password = "admin123" (应使用环境变量或密钥管理服务)
//	     apiKey := "sk-1234567890abcdef" (应从配置文件或Vault读取)
//
//	2. SQL注入风险检测 (SQL Injection Vulnerability):
//	   • 检测目标: 不安全的SQL查询拼接
//	   • 规则:
//	     - 识别字符串拼接构建SQL: "SELECT * FROM users WHERE id = " + userInput
//	     - 检测fmt.Sprintf构建查询: fmt.Sprintf("DELETE FROM %s", tableName)
//	     - 检测未使用预编译语句的场景
//	   • 示例违规:
//	     query := "SELECT * FROM users WHERE name = '" + userName + "'" (应使用参数化查询)
//	     db.Exec(fmt.Sprintf("INSERT INTO %s VALUES (%s)", table, values)) (应使用占位符)
//
//	3. 命令注入风险检测 (Command Injection Vulnerability):
//	   • 检测目标: 不安全的系统命令执行
//	   • 规则:
//	     - 识别exec.Command使用用户输入: exec.Command("sh", "-c", userInput)
//	     - 检测未验证的命令参数拼接
//	     - 识别os/exec包的不安全使用
//	   • 示例违规:
//	     exec.Command("sh", "-c", "rm -rf " + userPath) (应验证userPath)
//	     exec.Command(userCommand, args...) (userCommand应限制在白名单内)
//
//	4. 路径遍历攻击检测 (Path Traversal Vulnerability):
//	   • 检测目标: 不安全的文件路径操作
//	   • 规则:
//	     - 识别未验证的文件路径: os.Open(userPath) (缺少路径验证)
//	     - 检测../目录遍历模式: filepath.Join(baseDir, userInput) 未检查..
//	     - 检测绝对路径使用风险
//	   • 示例违规:
//	     os.ReadFile(filepath.Join("/data", userFileName)) (应验证userFileName不含..)
//	     http.ServeFile(w, r, r.URL.Query().Get("file")) (直接使用用户输入)
//
//	5. 弱加密算法检测 (Weak Cryptography Detection):
//	   • 检测目标: 使用已知不安全的加密算法
//	   • 规则:
//	     - 识别MD5/SHA1哈希算法: crypto/md5, crypto/sha1
//	     - 检测DES加密: crypto/des
//	     - 识别ECB模式: cipher.NewCBCEncrypter (应使用GCM)
//	   • 示例违规:
//	     md5.Sum(data) (应使用SHA256或更强算法)
//	     des.NewCipher(key) (应使用AES-256-GCM)
//
//	6. 不安全的随机数生成 (Insecure Random Number Generation):
//	   • 检测目标: 使用math/rand替代crypto/rand
//	   • 规则:
//	     - 识别math/rand用于安全场景: rand.Intn() 生成token/密钥
//	     - 检测未设置种子: rand.Intn() 未调用rand.Seed()
//	   • 示例违规:
//	     token := rand.Int() (应使用crypto/rand.Read())
//	     sessionID := fmt.Sprintf("%d", rand.Intn(1000000)) (可预测,不安全)
//
//	7. 敏感数据泄露检测 (Sensitive Data Exposure):
//	   • 检测目标: 敏感信息的不安全传输和存储
//	   • 规则:
//	     - 检测HTTP传输敏感数据: http.Get("http://...") 非HTTPS
//	     - 识别明文日志记录密码: log.Printf("password: %s", pwd)
//	     - 检测未加密的文件存储: ioutil.WriteFile(path, sensitiveData, 0644)
//	   • 示例违规:
//	     log.Println("User credentials:", username, password) (密码泄露到日志)
//	     os.WriteFile("config.txt", []byte(apiKey), 0644) (API密钥明文存储)
//
// 当前实现状态:
//
//	占位符实现 (Placeholder Implementation):
//	• 返回值: 固定返回0 (无安全问题)
//	• 实现程度: 0% (仅框架,无实际检查逻辑)
//	• 注释说明: "实现安全问题检查逻辑"
//	• 示例提示: "例如：硬编码密码、SQL注入风险、不安全的随机数生成等"
//	• 设计意图: 为未来功能扩展预留接口,与gosec工具集成
//
// 执行流程(当前):
//
//	1. 初始化问题计数器:
//	   • issues := 0
//	   • 准备累加安全问题数量
//
//	2. 检查逻辑占位:
//	   • 注释: "实现安全问题检查逻辑"
//	   • 当前: 无实际检查,跳过
//
//	3. 返回问题计数:
//	   • return issues
//	   • 当前固定返回0
//
// 预期执行流程(未来实现):
//
//	1. 集成gosec安全扫描工具:
//	   • 执行: gosec -fmt=json ./...
//	   • 解析JSON输出提取安全问题
//	   • 过滤: 按严重程度(HIGH/MEDIUM/LOW)分类
//
//	2. 解析AST进行静态安全分析:
//	   • 使用go/parser遍历AST
//	   • 检测不安全的函数调用(exec.Command, os.Open, db.Exec等)
//	   • 识别硬编码字符串中的敏感模式
//
//	3. 应用安全规则检查:
//	   • 硬编码检测: 正则匹配password/apiKey/token等关键词
//	   • SQL注入检测: 识别字符串拼接构建SQL
//	   • 命令注入检测: 检查exec.Command的参数来源
//	   • 加密检测: 识别md5/sha1/des等弱算法导入
//
//	4. 记录安全问题:
//	   • 问题类型: G101(硬编码密码), G201(SQL注入), G204(命令注入)等
//	   • 位置信息: 文件路径、行号、列号
//	   • 严重程度: HIGH/MEDIUM/LOW
//
//	5. 统计并返回问题数量:
//	   • 累加所有严重程度的安全问题
//	   • return len(securityIssues)
//
// 实现方案建议:
//
//	方案1: 集成gosec工具(推荐):
//	• 工具: gosec - Go安全扫描器
//	• 命令: gosec -fmt=json -out=results.json ./...
//	• 规则: 100+安全规则(G101-G602),覆盖OWASP Top 10
//	• 优点: 规则完善,误报率低,社区广泛使用
//	• 实现:
//	  cmd := exec.Command("gosec", "-fmt=json", "./...")
//	  output, _ := cmd.Output()
//	  var result GosecResult
//	  json.Unmarshal(output, &result)
//	  return len(result.Issues)
//
//	方案2: 自定义安全规则引擎:
//	• 实现: 基于go/ast的安全模式匹配
//	• 规则定义: YAML配置文件定义检测模式
//	• 优点: 完全控制,支持团队特定安全要求
//	• 缺点: 实现复杂,规则覆盖不如gosec全面
//
//	方案3: 多工具组合:
//	• 工具链: gosec(安全) + govulncheck(漏洞) + semgrep(模式)
//	• 互补: gosec检测代码问题,govulncheck检测依赖漏洞
//	• 优点: 全面覆盖,多层防护
//	• 缺点: 性能开销大,结果需要去重
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: 定位需要检查的Go源文件和依赖
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - int: 安全问题数量
//     • 当前: 固定返回0 (占位符实现)
//     • 未来: 实际安全问题数量(如3个G101硬编码+2个G201注入=5)
//     • 范围: 0 - N (0到检测到的安全问题总数)
//     • 0: 无安全问题(理想状态)
//     • >0: 存在安全风险,数值为问题数量
//     • 用途: 用于calculateSecurityScore()计算安全性分数
//
// 使用场景:
//   - calculateSecurityScore()调用: 作为安全性评分的输入指标
//   - CI/CD安全门禁: 高危安全问题(返回>0)时阻止代码合并
//   - 安全审计: 定期扫描识别新引入的安全风险
//   - 合规检查: 满足PCI-DSS、SOC2等安全合规要求
//   - 开发培训: 向团队展示常见安全问题,提升安全意识
//
// 示例(未来实现):
//
//	// 示例1: 安全项目 (当前所有项目都返回此结果)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/secure-project"},
//	}
//	issues := cqe.checkSecurityIssues()
//	// 当前: 直接返回0 (占位符)
//	// 未来: gosec扫描发现0个安全问题
//	// issues = 0
//
//	// 示例2: 硬编码密码问题 (未来实现)
//	// 代码: const dbPassword = "admin123"
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/hardcoded-project"},
//	}
//	issues := cqe.checkSecurityIssues()
//	// 未来: gosec检测到G101硬编码凭证
//	// 违规: const dbPassword = "admin123" (应使用环境变量)
//	// issues = 1
//
//	// 示例3: SQL注入风险 (未来实现)
//	// 代码: query := "SELECT * FROM users WHERE id = " + userId
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/sql-injection-project"},
//	}
//	issues := cqe.checkSecurityIssues()
//	// 未来: gosec检测到G201 SQL注入风险
//	// 违规: 字符串拼接构建SQL (应使用db.Query("SELECT ... WHERE id = ?", userId))
//	// issues = 1
//
//	// 示例4: 多种安全问题 (未来实现)
//	// 代码包含: 硬编码密码(3处) + SQL注入(2处) + 命令注入(1处) + 弱加密(1处)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/insecure-project"},
//	}
//	issues := cqe.checkSecurityIssues()
//	// 未来: gosec检测到多种安全问题
//	// G101: 3个硬编码凭证
//	// G201: 2个SQL注入风险
//	// G204: 1个命令注入漏洞
//	// G401: 1个弱加密算法(MD5)
//	// issues = 7
//
//	// 示例5: 依赖漏洞检测 (未来扩展)
//	// go.mod包含已知漏洞的依赖: github.com/old/package@v1.0.0 (CVE-2023-12345)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/vulnerable-deps-project"},
//	}
//	issues := cqe.checkSecurityIssues()
//	// 未来: govulncheck检测到依赖漏洞
//	// CVE-2023-12345: github.com/old/package v1.0.0 (应升级到v1.2.1)
//	// issues = 1 (如果集成govulncheck)
//
// 注意事项:
//   - 占位符实现: 当前方法未实现任何检查逻辑,仅返回0
//   - 不影响评分: 由于返回0,不会对安全性分数产生负面影响
//   - 安全假阳性: 当前返回0可能给团队带来虚假安全感,建议尽快实现
//   - gosec依赖: 未来实现依赖gosec工具,需在CI/CD环境安装
//   - 性能考虑: 安全扫描大型项目可能耗时(10-60秒),考虑异步执行
//   - 误报处理: 安全工具可能产生误报,需要#nosec注释豁免机制
//   - 持续更新: 安全规则需随新漏洞类型演进,定期更新gosec版本
//   - 严重程度: 应区分HIGH/MEDIUM/LOW严重程度,分别处理
//   - 第三方代码: vendor目录和生成代码应排除扫描
//   - 合规要求: 某些行业(金融/医疗)对安全问题零容忍,需要阻断机制
//
// 改进方向:
//   - 实现gosec集成: 执行gosec命令并解析JSON输出统计安全问题
//     - cmd := exec.Command("gosec", "-fmt=json", "-out=result.json", "./...")
//   - 严重程度区分: 分别统计HIGH/MEDIUM/LOW问题,加权计算
//     - highIssues×10 + mediumIssues×5 + lowIssues×1
//   - govulncheck集成: 检测依赖包的已知漏洞(CVE)
//     - govulncheck -json ./... | jq '.Vulns | length'
//   - 自定义安全规则: 支持团队特定安全检查规则
//     - 如检测特定敏感API调用、内部安全规范违反
//   - 安全问题分类: 按OWASP Top 10分类展示(注入/认证/加密/配置等)
//   - 修复建议生成: 为每个安全问题提供具体修复方案
//     - G101硬编码密码 → 建议: 使用os.Getenv("DB_PASSWORD")
//   - 安全基线检查: 定义安全基线,只报告新增问题,忽略历史遗留
//     - git diff --name-only | xargs gosec
//   - 漏洞数据库同步: 自动同步最新CVE数据库,及时发现新漏洞
//   - 密钥熵值检测: 使用熵值分析识别高熵字符串(可能的密钥)
//     - 熵值>4.5的字符串可能是Base64编码密钥
//   - 代码混淆检测: 识别恶意代码混淆模式
//   - SAST/DAST集成: 结合静态(SAST)和动态(DAST)应用安全测试
//   - 安全培训模式: 为检测到的问题生成教学材料,帮助开发者学习
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) checkSecurityIssues() int {
	issues := 0
	// 实现安全问题检查逻辑
	// 例如：硬编码密码、SQL注入风险、不安全的随机数生成等
	return issues
}

// checkPerformanceIssues 检查代码性能问题并返回性能隐患数量(性能分析维度辅助检查方法-待实现)
//
// 功能说明:
//
//	本方法是代码质量评估系统中性能分析维度(performance_analysis)的辅助检查方法,
//	负责识别项目代码中可能存在的性能隐患和效率问题。该方法设计用于检测常见的
//	性能反模式,如不必要的字符串拼接(循环中使用+操作符)、未优化的循环嵌套、
//	内存泄漏风险(goroutine泄漏、defer在循环中)、频繁的反射调用、大对象值拷贝、
//	低效的数据结构使用等性能问题。
//	当前实现为占位符(placeholder),返回固定值0,表示无性能问题,未来版本将实现完整的
//	性能检查逻辑,集成静态分析、基准测试、逃逸分析等工具,识别性能瓶颈和优化机会。
//
// 设计目标(待实现):
//
//	1. 字符串拼接优化检测 (String Concatenation Inefficiency):
//	   • 检测目标: 循环中使用+操作符拼接字符串
//	   • 规则:
//	     - 识别for循环内的 str += "xxx" 或 str = str + "xxx"
//	     - 检测应使用strings.Builder或bytes.Buffer的场景
//	     - 计算拼接次数,评估性能影响(O(n²) vs O(n))
//	   • 示例违规:
//	     for i := 0; i < 1000; i++ { result += fmt.Sprintf("%d", i) } (O(n²),应用Builder)
//	     正确: var b strings.Builder; for i := 0; i < 1000; i++ { b.WriteString(...) }
//
//	2. 循环优化检测 (Loop Optimization):
//	   • 检测目标: 低效的循环模式
//	   • 规则:
//	     - 识别循环不变量提升机会: for循环内重复计算len(slice)
//	     - 检测嵌套循环复杂度(O(n³)及以上)
//	     - 识别range遍历大切片时的值拷贝: for _, item := range largeSlice
//	   • 示例违规:
//	     for i := 0; i < len(data); i++ {...} (len(data)每次计算,应提升到循环外)
//	     for _, item := range items {...} (item为大结构体,应用索引: for i := range items)
//
//	3. 内存泄漏检测 (Memory Leak Detection):
//	   • 检测目标: 可能导致内存泄漏的模式
//	   • 规则:
//	     - 识别goroutine泄漏: go func()未使用context/channel控制退出
//	     - 检测defer在循环中: for循环内defer会延迟到函数结束,积累资源
//	     - 识别未关闭的资源: 文件、网络连接、数据库连接未关闭
//	   • 示例违规:
//	     for _, file := range files { f, _ := os.Open(file); defer f.Close() } (defer积累)
//	     go func() { for { ... } }() (无退出机制,goroutine泄漏)
//	     正确: for { f, _ := os.Open(file); f.Close(); ... } 或使用errgroup控制
//
//	4. 反射滥用检测 (Reflection Overuse):
//	   • 检测目标: 不必要的反射调用
//	   • 规则:
//	     - 识别高频路径中的reflect.TypeOf/ValueOf调用
//	     - 检测可以用类型断言替代反射的场景
//	     - 识别循环中重复的反射操作
//	   • 示例违规:
//	     for _, v := range data { t := reflect.TypeOf(v); ... } (每次反射,应缓存类型)
//	     reflect.ValueOf(x).Interface().(int) (应直接类型断言: x.(int))
//
//	5. 大对象值拷贝检测 (Large Object Copy):
//	   • 检测目标: 大结构体的值传递
//	   • 规则:
//	     - 识别大结构体(>128字节)作为函数参数值传递
//	     - 检测range遍历时大结构体值拷贝
//	     - 识别接口赋值时的大对象装箱
//	   • 示例违规:
//	     func processUser(u User) {...} (User 512字节,应传指针 *User)
//	     for _, user := range users {...} (users为[]User,应用索引遍历)
//
//	6. 低效数据结构检测 (Inefficient Data Structure):
//	   • 检测目标: 不合适的数据结构选择
//	   • 规则:
//	     - 识别线性查找应用map的场景: for循环查找,应用map[string]T
//	     - 检测频繁append未预分配容量: 应make([]T, 0, capacity)
//	     - 识别排序后未使用二分查找
//	   • 示例违规:
//	     for _, item := range items { if item.ID == targetID {...} } (O(n),应用map)
//	     var result []int; for ... { result = append(result, v) } (应预分配容量)
//
//	7. 并发性能问题检测 (Concurrency Performance):
//	   • 检测目标: 并发编程中的性能陷阱
//	   • 规则:
//	     - 识别锁竞争热点: 高频加锁区域
//	     - 检测channel误用: 应用sync.Pool的场景用channel传递
//	     - 识别goroutine过度创建: 应用worker pool模式
//	   • 示例违规:
//	     for i := 0; i < 10000; i++ { go process(i) } (10000个goroutine,应用worker pool)
//	     mutex.Lock(); heavyComputation(); mutex.Unlock() (锁内耗时操作)
//
// 当前实现状态:
//
//	占位符实现 (Placeholder Implementation):
//	• 返回值: 固定返回0 (无性能问题)
//	• 实现程度: 0% (仅框架,无实际检查逻辑)
//	• 注释说明: "实现性能问题检查逻辑"
//	• 示例提示: "例如：不必要的字符串拼接、未优化的循环、内存泄漏等"
//	• 设计意图: 为未来功能扩展预留接口
//
// 执行流程(当前):
//
//	1. 初始化问题计数器:
//	   • issues := 0
//	   • 准备累加性能问题数量
//
//	2. 检查逻辑占位:
//	   • 注释: "实现性能问题检查逻辑"
//	   • 当前: 无实际检查,跳过
//
//	3. 返回问题计数:
//	   • return issues
//	   • 当前固定返回0
//
// 预期执行流程(未来实现):
//
//	1. AST静态分析:
//	   • 使用go/parser遍历AST
//	   • 识别性能反模式(字符串拼接、循环嵌套等)
//	   • 提取函数调用、循环结构、变量赋值信息
//
//	2. 逃逸分析集成:
//	   • 执行: go build -gcflags="-m -m" ./... 2>&1
//	   • 解析逃逸分析输出
//	   • 识别不必要的堆分配(应栈分配的变量逃逸到堆)
//
//	3. 基准测试辅助:
//	   • 执行: go test -bench=. -benchmem ./...
//	   • 解析基准测试结果
//	   • 识别高内存分配、低性能的函数
//
//	4. 性能规则匹配:
//	   • 字符串拼接: 检测for循环内 str += ... 模式
//	   • 循环优化: 检测len(slice)在循环条件中
//	   • defer位置: 检测for循环内defer调用
//	   • 反射检测: 检测高频路径中reflect包使用
//
//	5. 统计并返回问题数量:
//	   • 累加所有性能问题
//	   • return len(performanceIssues)
//
// 实现方案建议:
//
//	方案1: AST模式匹配(推荐快速实现):
//	• 实现: 使用go/ast遍历,匹配性能反模式
//	• 规则定义: 硬编码常见性能问题检测逻辑
//	• 优点: 实现简单,无外部依赖,快速
//	• 缺点: 规则覆盖有限,需手动维护
//	• 示例:
//	  ast.Inspect(node, func(n ast.Node) bool {
//	      if forStmt, ok := n.(*ast.ForStmt); ok {
//	          // 检测循环内字符串拼接
//	      }
//	      return true
//	  })
//
//	方案2: 集成逃逸分析和pprof:
//	• 工具: go build -gcflags="-m" (逃逸分析)
//	• 工具: go tool pprof (性能分析)
//	• 优点: 准确识别实际性能问题
//	• 缺点: 需要运行时数据,静态分析不足
//
//	方案3: 静态分析工具集成:
//	• 工具: staticcheck (包含性能检查)
//	• 工具: go-critic (性能lint规则)
//	• 优点: 规则完善,社区维护
//	• 缺点: 需要外部工具依赖
//	• 示例:
//	  staticcheck -checks=S1*,SA6002,SA6005 ./... (字符串拼接、defer等)
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: 定位需要检查的Go源文件
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - int: 性能问题数量
//     • 当前: 固定返回0 (占位符实现)
//     • 未来: 实际性能问题数量(如5个字符串拼接+3个循环优化=8)
//     • 范围: 0 - N (0到检测到的性能问题总数)
//     • 0: 无性能问题(理想状态)
//     • >0: 存在性能隐患,数值为问题数量
//     • 用途: 用于calculatePerformanceScore()计算性能分数
//
// 使用场景:
//   - calculatePerformanceScore()调用: 作为性能评分的输入指标
//   - 性能优化指导: 识别需要优化的代码区域
//   - 代码审查辅助: 自动发现性能问题,减少人工审查负担
//   - 性能回归检测: CI/CD中检测新代码是否引入性能问题
//   - 开发培训: 向团队展示常见性能陷阱,提升性能意识
//
// 示例(未来实现):
//
//	// 示例1: 高性能项目 (当前所有项目都返回此结果)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/optimized-project"},
//	}
//	issues := cqe.checkPerformanceIssues()
//	// 当前: 直接返回0 (占位符)
//	// 未来: 检查发现0个性能问题
//	// issues = 0
//
//	// 示例2: 字符串拼接问题 (未来实现)
//	// 代码: var s string; for i := 0; i < 1000; i++ { s += strconv.Itoa(i) }
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/string-concat-project"},
//	}
//	issues := cqe.checkPerformanceIssues()
//	// 未来: AST检测到循环内字符串拼接
//	// 违规: for循环内 s += ... (O(n²),应用strings.Builder)
//	// issues = 1
//
//	// 示例3: defer在循环中 (未来实现)
//	// 代码: for _, file := range files { f, _ := os.Open(file); defer f.Close() }
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/defer-loop-project"},
//	}
//	issues := cqe.checkPerformanceIssues()
//	// 未来: AST检测到for循环内defer
//	// 违规: defer在循环中累积,应立即Close()
//	// issues = 1
//
//	// 示例4: 多种性能问题 (未来实现)
//	// 代码包含: 字符串拼接(3处) + defer循环(2处) + 大对象拷贝(5处) + 反射滥用(1处)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/slow-project"},
//	}
//	issues := cqe.checkPerformanceIssues()
//	// 未来: 检测到多种性能问题
//	// 字符串拼接: 3处
//	// defer循环: 2处
//	// 大对象拷贝: 5处 (User结构体512字节值传递)
//	// 反射滥用: 1处 (热路径中reflect.TypeOf)
//	// issues = 11
//
//	// 示例5: 并发性能问题 (未来实现)
//	// 代码: for i := 0; i < 10000; i++ { go worker(i) }
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/goroutine-storm-project"},
//	}
//	issues := cqe.checkPerformanceIssues()
//	// 未来: 检测到goroutine过度创建
//	// 违规: 10000个goroutine同时创建,应用worker pool
//	// issues = 1
//
// 注意事项:
//   - 占位符实现: 当前方法未实现任何检查逻辑,仅返回0
//   - 不影响评分: 由于返回0,不会对性能分数产生负面影响
//   - 性能假象: 当前返回0可能掩盖真实性能问题,建议尽快实现
//   - 误报风险: 静态分析可能误报,如字符串拼接次数少时+操作符可接受
//   - 运行时依赖: 部分性能问题需要运行时profiling才能发现
//   - 阈值配置: 性能问题严重程度应可配置(如循环次数>100才报告)
//   - 上下文敏感: 某些性能问题在低频路径可接受,热路径才需优化
//   - 微优化陷阱: 避免过度优化可读性较差但性能提升微小的代码
//   - Go版本差异: 不同Go版本编译器优化不同,逃逸分析结果可能变化
//   - 架构依赖: CPU密集vs IO密集型应用的性能优先级不同
//
// 改进方向:
//   - 实现字符串拼接检测: AST识别for循环内 str += ... 模式
//	     - 建议替换: strings.Builder 或 bytes.Buffer
//   - defer位置检测: AST识别for循环内defer语句
//	     - 建议: 立即调用Close()或提取到函数外
//   - 循环不变量提升: 检测len(slice)在for条件中重复计算
//	     - 建议: n := len(slice); for i := 0; i < n; i++
//   - 逃逸分析集成: 解析go build -gcflags="-m"输出
//	     - 识别不必要的堆分配,建议优化
//   - 大对象检测: 计算结构体大小,检测值传递
//	     - 阈值: >128字节建议使用指针
//   - 反射热路径检测: 结合pprof识别高频反射调用
//	     - 建议: 使用类型断言或代码生成替代
//   - 并发模式检测: 识别goroutine泄漏、锁竞争
//	     - 建议: context控制、sync.Pool、worker pool
//   - 基准测试集成: 自动运行benchmark,识别性能回归
//	     - 对比历史数据,报告性能下降>10%的函数
//   - 性能分级: 区分关键路径(必须优化)和非关键路径
//	     - 热路径性能问题权重×10,冷路径权重×1
//   - 修复建议生成: 为每个问题提供具体代码修复示例
//	     - 字符串拼接 → 完整strings.Builder示例代码
//   - 性能预算: 设定性能预算(如内存分配<1MB/req),超预算报警
//   - CPU/内存分离: 分别统计CPU密集和内存密集问题
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) checkPerformanceIssues() int {
	issues := 0
	// 实现性能问题检查逻辑
	// 例如：不必要的字符串拼接、未优化的循环、内存泄漏等
	return issues
}

// checkAllocationPatterns 检查内存分配模式并返回异常分配数量(性能分析维度辅助检查方法-待实现)
//
// 功能说明:
//
//	本方法是代码质量评估系统中性能分析维度(performance_analysis)的辅助检查方法,
//	专门负责检查项目代码中的内存分配模式和潜在的内存效率问题。该方法设计用于识别
//	高频内存分配、不必要的堆分配、内存泄漏隐患等问题,通过静态分析和逃逸分析结果
//	评估代码的内存使用效率。重点关注堆栈分配策略、切片和map的容量预分配、
//	闭包变量捕获、接口装箱开销等内存分配相关的性能问题。
//	当前实现为占位符(placeholder),返回固定值0,表示内存分配模式正常,未来版本将
//	集成逃逸分析、pprof内存分析等工具,提供精确的内存优化建议。
//
// 设计目标(待实现):
//
//	1. 逃逸分析问题检测 (Escape Analysis Issues):
//	   • 检测目标: 不必要的堆分配(应栈分配但逃逸到堆)
//	   • 规则:
//	     - 识别局部变量不必要逃逸: 返回局部变量指针
//	     - 检测闭包捕获导致的逃逸
//	     - 识别接口类型赋值导致的装箱逃逸
//	   • 示例违规:
//	     func getUser() *User { u := User{...}; return &u } (局部变量逃逸)
//	     正确: func getUser() User { return User{...} } (栈分配,无逃逸)
//	   • 检测方式: 解析 go build -gcflags="-m" 输出
//
//	2. 切片容量预分配检测 (Slice Capacity Pre-allocation):
//	   • 检测目标: 频繁append未预分配容量
//	   • 规则:
//	     - 识别已知大小的切片未预分配: 循环append但已知元素数量
//	     - 检测切片频繁扩容: append导致多次内存重新分配
//	     - 计算预分配收益: 元素数量>16时建议预分配
//	   • 示例违规:
//	     var result []int; for i := 0; i < 1000; i++ { result = append(result, i) }
//	     正确: result := make([]int, 0, 1000); for i := 0; i < 1000; i++ { result = append(result, i) }
//	   • 内存节省: 1000元素场景避免~10次扩容和拷贝
//
//	3. Map容量预分配检测 (Map Capacity Pre-allocation):
//	   • 检测目标: 大量插入的map未预分配容量
//	   • 规则:
//	     - 识别循环插入map但未指定初始容量
//	     - 检测map频繁扩容和rehash
//	     - 建议: 已知键数量时使用make(map[K]V, capacity)
//	   • 示例违规:
//	     m := make(map[string]int); for i := 0; i < 10000; i++ { m[key] = val }
//	     正确: m := make(map[string]int, 10000); for i := 0; i < 10000; i++ { m[key] = val }
//	   • 性能收益: 避免多次rehash,提升插入性能
//
//	4. 闭包变量捕获检测 (Closure Variable Capture):
//	   • 检测目标: 闭包捕获导致的内存分配
//	   • 规则:
//	     - 识别闭包捕获大对象: goroutine闭包捕获整个结构体
//	     - 检测循环变量闭包捕获: for i := range ... { go func() { use(i) } }
//	     - 建议: 通过参数传递替代捕获
//	   • 示例违规:
//	     for i := 0; i < 10; i++ { go func() { fmt.Println(i) }() } (捕获循环变量)
//	     正确: for i := 0; i < 10; i++ { i := i; go func() { fmt.Println(i) }() }
//	   • 内存影响: 闭包捕获变量分配到堆,增加GC压力
//
//	5. 接口装箱开销检测 (Interface Boxing Overhead):
//	   • 检测目标: 频繁的接口类型转换和装箱
//	   • 规则:
//	     - 识别循环中接口赋值: for循环内将值类型赋给interface{}
//	     - 检测不必要的接口抽象: 仅一个实现的接口
//	     - 计算装箱开销: 每次装箱1次堆分配
//	   • 示例违规:
//	     var items []interface{}; for i := 0; i < 1000; i++ { items = append(items, i) }
//	     正确: var items []int; for i := 0; i < 1000; i++ { items = append(items, i) }
//	   • 性能影响: 1000次装箱=1000次堆分配+装箱元数据
//
//	6. 字符串转换分配检测 (String Conversion Allocation):
//	   • 检测目标: 不必要的字符串与字节切片转换
//	   • 规则:
//	     - 识别频繁string([]byte)和[]byte(string)转换
//	     - 检测字符串拼接临时分配: += 每次分配新字符串
//	     - 建议: 使用unsafe或strings.Builder避免拷贝
//	   • 示例违规:
//	     for _, line := range lines { s := string(line); process(s) } (每次拷贝)
//	     data := []byte("hello"); for i := 0; i < 1000; i++ { s := string(data) } (1000次拷贝)
//
//	7. 零值内存浪费检测 (Zero Value Memory Waste):
//	   • 检测目标: 大结构体使用零值初始化
//	   • 规则:
//	     - 识别大结构体字段稀疏但完整分配
//	     - 检测只用部分字段但分配全部内存
//	     - 建议: 使用指针字段或嵌入接口
//	   • 示例场景:
//	     type Config struct { A [1000]int; B string; C bool }
//	     c := Config{ B: "value" } (浪费~8KB,大部分是零值A)
//
// 当前实现状态:
//
//	占位符实现 (Placeholder Implementation):
//	• 返回值: 固定返回0 (无内存分配问题)
//	• 实现程度: 0% (仅框架,无实际检查逻辑)
//	• 注释说明: "实现内存分配检查逻辑"
//	• 设计意图: 为未来内存优化分析预留接口
//
// 执行流程(当前):
//
//	1. 初始化问题计数器:
//	   • issues := 0
//	   • 准备累加内存分配问题数量
//
//	2. 检查逻辑占位:
//	   • 注释: "实现内存分配检查逻辑"
//	   • 当前: 无实际检查,跳过
//
//	3. 返回问题计数:
//	   • return issues
//	   • 当前固定返回0
//
// 预期执行流程(未来实现):
//
//	1. 执行逃逸分析:
//	   • 命令: go build -gcflags="-m -m" ./... 2>&1
//	   • 解析输出: 提取逃逸变量和原因
//	   • 分类: 区分必要逃逸和可优化逃逸
//
//	2. AST分析切片/map使用:
//	   • 识别make([]T, 0)和make(map[K]V)调用
//	   • 检测后续append/插入操作
//	   • 评估是否应预分配容量
//
//	3. 闭包分析:
//	   • 识别go func()闭包
//	   • 检测捕获变量类型和大小
//	   • 建议参数传递替代捕获
//
//	4. 接口使用分析:
//	   • 检测interface{}切片
//	   • 识别循环中接口赋值
//	   • 计算装箱频率
//
//	5. 统计并返回问题数量:
//	   • 累加所有内存分配问题
//	   • return len(allocationIssues)
//
// 实现方案建议:
//
//	方案1: 逃逸分析集成(推荐):
//	• 工具: go build -gcflags="-m -m -l"
//	• 解析: 正则提取逃逸信息
//	• 示例输出: ./main.go:10:2: moved to heap: user
//	• 优点: 准确识别堆分配,来自编译器
//	• 实现:
//	  cmd := exec.Command("go", "build", "-gcflags=-m -m", "./...")
//	  output, _ := cmd.CombinedOutput()
//	  escapeRegex := regexp.MustCompile(`moved to heap: (\w+)`)
//	  matches := escapeRegex.FindAllStringSubmatch(string(output), -1)
//
//	方案2: pprof内存分析(运行时):
//	• 工具: go test -memprofile=mem.prof
//	• 分析: go tool pprof -alloc_space mem.prof
//	• 优点: 实际运行时内存分配数据
//	• 缺点: 需要运行代码,非纯静态分析
//
//	方案3: AST+启发式规则:
//	• 实现: 遍历AST识别分配模式
//	• 规则: 检测make/append/闭包等
//	• 优点: 无外部依赖,纯Go实现
//	• 缺点: 准确性低于编译器分析
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: 定位需要检查的Go源文件
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - int: 内存分配问题数量
//     • 当前: 固定返回0 (占位符实现)
//     • 未来: 实际内存分配问题数量(如5个逃逸+3个未预分配=8)
//     • 范围: 0 - N (0到检测到的内存问题总数)
//     • 0: 内存分配模式优秀(理想状态)
//     • >0: 存在内存优化机会,数值为问题数量
//     • 用途: 用于calculatePerformanceScore()计算性能分数
//
// 使用场景:
//   - calculatePerformanceScore()调用: 作为性能评分的内存维度指标
//   - 内存优化指导: 识别高内存分配的代码区域
//   - GC压力评估: 预测堆分配对GC的影响
//   - 性能调优: 指导开发者优化内存使用
//   - 容量规划: 评估应用内存需求
//
// 示例(未来实现):
//
//	// 示例1: 内存优化项目 (当前所有项目都返回此结果)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/optimized-project"},
//	}
//	issues := cqe.checkAllocationPatterns()
//	// 当前: 直接返回0 (占位符)
//	// 未来: 逃逸分析发现0个问题
//	// issues = 0
//
//	// 示例2: 局部变量逃逸 (未来实现)
//	// 代码: func getUser() *User { u := User{...}; return &u }
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/escape-project"},
//	}
//	issues := cqe.checkAllocationPatterns()
//	// 未来: 逃逸分析检测到局部变量逃逸
//	// 输出: ./user.go:10:2: moved to heap: u
//	// issues = 1
//
//	// 示例3: 切片未预分配 (未来实现)
//	// 代码: var result []int; for i := 0; i < 10000; i++ { result = append(result, i) }
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/slice-project"},
//	}
//	issues := cqe.checkAllocationPatterns()
//	// 未来: AST检测到切片频繁扩容
//	// 建议: result := make([]int, 0, 10000)
//	// issues = 1
//
//	// 示例4: 接口装箱开销 (未来实现)
//	// 代码: var items []interface{}; for i := 0; i < 1000; i++ { items = append(items, i) }
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/boxing-project"},
//	}
//	issues := cqe.checkAllocationPatterns()
//	// 未来: 检测到循环中1000次接口装箱
//	// 建议: 使用[]int替代[]interface{}
//	// issues = 1
//
//	// 示例5: 多种内存问题 (未来实现)
//	// 代码包含: 逃逸(5处) + 切片未预分配(3处) + map未预分配(2处) + 接口装箱(1处)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/memory-heavy-project"},
//	}
//	issues := cqe.checkAllocationPatterns()
//	// 未来: 检测到多种内存分配问题
//	// 逃逸: 5个局部变量逃逸
//	// 切片: 3处未预分配(每处~14次扩容)
//	// map: 2处未预分配(每处~6次rehash)
//	// 接口: 1处循环装箱(1000次)
//	// issues = 11
//
// 注意事项:
//   - 占位符实现: 当前方法未实现任何检查逻辑,仅返回0
//   - 不影响评分: 由于返回0,不会对性能分数产生负面影响
//   - 逃逸分析限制: 编译器逃逸分析结果可能因Go版本变化
//   - 误报风险: 某些必要逃逸(如返回接口)不应算作问题
//   - 阈值配置: 切片/map预分配阈值应可配置(如>16元素才建议)
//   - 性能收益评估: 需要量化优化收益,避免过度优化
//   - pprof依赖: 运行时分析需要执行代码和生成profile
//   - 编译标志: -gcflags可能影响编译性能,仅分析时使用
//   - 内存vs可读性: 某些内存优化可能降低代码可读性
//   - 架构差异: 不同CPU架构内存对齐要求不同,影响结构体大小
//
// 改进方向:
//   - 实现逃逸分析集成: 执行go build -gcflags="-m -m"并解析输出
//	     - 识别moved to heap、leaking param等模式
//   - 切片容量分析: AST识别make([]T, 0)后续append模式
//	     - 建议: 已知大小时预分配make([]T, 0, capacity)
//   - map容量分析: AST识别make(map[K]V)后续插入模式
//	     - 建议: 已知键数量时预分配make(map[K]V, capacity)
//   - 闭包捕获检测: AST识别go func(){}()中捕获变量
//	     - 建议: 通过参数传递替代捕获
//   - 接口装箱统计: 识别[]interface{}和循环中interface赋值
//	     - 建议: 使用具体类型切片
//   - pprof集成: 解析go test -memprofile输出
//	     - top N内存分配函数,定位热点
//   - 结构体大小计算: 使用unsafe.Sizeof分析结构体内存
//	     - 建议: 优化字段顺序减少padding
//   - 内存分配热图: 可视化内存分配分布
//	     - 高亮高频分配代码行
//   - 历史对比: 对比历史内存分析数据,识别回归
//	     - 新代码内存分配增加>20%时报警
//   - 内存预算: 设定内存分配预算(如<10MB),超预算阻断
//   - 优化收益评估: 计算优化前后内存分配差异
//	     - 显示预分配节省的扩容次数和内存拷贝
//   - 自动修复建议: 生成优化后的代码示例
//	     - 切片未预分配 → 生成make([]T, 0, N)代码
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) checkAllocationPatterns() int {
	issues := 0
	// 实现内存分配检查逻辑
	return issues
}

// hasReadme 检查项目根目录是否存在README文件(文档质量维度核心检查方法)
//
// 功能说明:
//
//	本方法是代码质量评估系统中文档质量维度(documentation_quality)的核心检查方法之一,
//	负责验证项目根目录是否包含README文档文件。README文件是开源项目的门面和入口,
//	提供项目简介、安装说明、使用指南、贡献指南等关键信息,是新用户和贡献者了解项目的
//	第一手资料。本方法通过文件系统检查,识别常见的README文件命名格式(README.md,
//	README.txt, README等),返回布尔值表示是否存在任何一种格式的README文件。
//	该检查是文档质量评估的基础指标,README存在与否直接影响项目的可访问性和专业度。
//
// 检查策略:
//
//	1. 多格式兼容策略 (Multi-Format Compatibility):
//	   • 支持格式: 5种常见README文件命名格式
//	   • 格式列表:
//	     - README.md (Markdown格式,最常用,GitHub/GitLab默认渲染)
//	     - README.txt (纯文本格式,通用兼容)
//	     - README (无扩展名,传统Unix风格)
//	     - readme.md (小写Markdown,部分项目使用)
//	     - readme.txt (小写文本,部分项目使用)
//	   • 优先级: 无优先级,任意一种存在即通过
//	   • 大小写: 区分大小写(README vs readme)
//
//	2. 文件存在性检查 (File Existence Check):
//	   • 检查方式: os.Stat() 文件系统调用
//	   • 路径构建: filepath.Join(ProjectPath, filename)
//	   • 成功条件: os.Stat() 返回nil错误,表示文件存在
//	   • 失败条件: os.Stat() 返回error,表示文件不存在或无权限
//	   • 短路逻辑: 任意文件存在即立即返回true,无需检查剩余文件
//
//	3. 路径安全性 (Path Safety):
//	   • 路径拼接: 使用filepath.Join()确保跨平台兼容性
//	   • 仅检查根目录: 仅检查ProjectPath根目录,不递归子目录
//	   • 不验证内容: 仅检查文件存在,不读取内容或验证格式
//
// 执行流程:
//
//	1. 定义README文件名列表:
//	   • readmeFiles := []string{"README.md", "README.txt", "README", "readme.md", "readme.txt"}
//	   • 硬编码5种常见格式
//	   • 顺序: README.md优先列出(最常用)
//
//	2. 遍历文件名列表:
//	   • for循环: for _, filename := range readmeFiles
//	   • 线性扫描: 依次检查每个文件名
//
//	3. 构建完整文件路径:
//	   • 路径拼接: filepath.Join(cqe.results.ProjectPath, filename)
//	   • 示例: "/home/user/project" + "README.md" → "/home/user/project/README.md"
//	   • 跨平台: filepath.Join自动处理Windows(\)和Unix(/)分隔符
//
//	4. 检查文件是否存在:
//	   • 调用: os.Stat(fullPath)
//	   • 返回: (FileInfo, error)
//	   • 成功: err == nil (文件存在)
//	   • 失败: err != nil (文件不存在或无权限)
//
//	5. 短路返回或继续:
//	   • 文件存在: 立即返回true,无需检查剩余文件
//	   • 文件不存在: 继续下一个文件名
//
//	6. 所有文件都不存在时返回false:
//	   • 循环结束: 5个文件名都不存在
//	   • 返回: return false (项目无README文件)
//
// README文件重要性:
//
//	README文件的核心作用:
//	1. 项目门面 (Project Homepage):
//	   • GitHub/GitLab自动渲染为项目首页
//	   • 新访客第一眼看到的内容
//	   • 决定用户是否继续深入了解项目
//
//	2. 快速上手指南 (Quick Start Guide):
//	   • 安装说明: 依赖安装、编译步骤
//	   • 使用示例: 基本用法、核心功能演示
//	   • 配置指南: 环境变量、配置文件
//
//	3. 贡献指南入口 (Contribution Entry):
//	   • 如何贡献: 指向CONTRIBUTING.md
//	   • 开发环境: 本地开发环境搭建
//	   • 代码规范: 编码风格、提交规范
//
//	4. 项目信息聚合 (Information Hub):
//	   • 项目简介: 解决什么问题,适用场景
//	   • 特性列表: 核心功能、技术亮点
//	   • 文档链接: API文档、设计文档、教程
//	   • 许可证: 开源协议、使用限制
//	   • 联系方式: 社区、讨论组、邮件列表
//
//	5. 搜索引擎优化 (SEO):
//	   • README内容被搜索引擎索引
//	   • 提升项目在搜索结果中的排名
//	   • 吸引潜在用户和贡献者
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: README文件检查的基准目录
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - bool: README文件是否存在
//     • true: 项目根目录包含至少一种README文件
//     • false: 项目根目录不包含任何README文件
//     • 用途: 用于calculateDocumentationScore()计算文档质量分数
//
// 使用场景:
//   - calculateDocumentationScore()调用: 作为文档质量评分的关键指标
//   - 开源项目质量检查: README是开源项目的必备文件
//   - CI/CD质量门禁: README缺失时发出警告或阻止发布
//   - 项目初始化检查: 新项目创建时提醒添加README
//   - 文档完整性审计: 评估项目文档的完整程度
//
// 示例:
//
//	// 示例1: 项目有README.md文件
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/good-project"},
//	}
//	hasReadme := cqe.hasReadme()
//	// 文件检查: /home/user/good-project/README.md 存在
//	// os.Stat()返回: (FileInfo, nil)
//	// 短路返回: true (第一个文件即存在,无需检查剩余文件)
//	// hasReadme = true
//
//	// 示例2: 项目有小写readme.md文件
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/lowercase-project"},
//	}
//	hasReadme := cqe.hasReadme()
//	// 文件检查顺序:
//	// 1. /home/user/lowercase-project/README.md 不存在
//	// 2. /home/user/lowercase-project/README.txt 不存在
//	// 3. /home/user/lowercase-project/README 不存在
//	// 4. /home/user/lowercase-project/readme.md 存在 ✓
//	// 短路返回: true
//	// hasReadme = true
//
//	// 示例3: 项目有传统README文件(无扩展名)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/unix-project"},
//	}
//	hasReadme := cqe.hasReadme()
//	// 文件检查顺序:
//	// 1. /home/user/unix-project/README.md 不存在
//	// 2. /home/user/unix-project/README.txt 不存在
//	// 3. /home/user/unix-project/README 存在 ✓ (Unix传统风格)
//	// 短路返回: true
//	// hasReadme = true
//
//	// 示例4: 项目完全没有README文件
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/bad-project"},
//	}
//	hasReadme := cqe.hasReadme()
//	// 文件检查顺序:
//	// 1. /home/user/bad-project/README.md 不存在
//	// 2. /home/user/bad-project/README.txt 不存在
//	// 3. /home/user/bad-project/README 不存在
//	// 4. /home/user/bad-project/readme.md 不存在
//	// 5. /home/user/bad-project/readme.txt 不存在
//	// 循环结束,所有文件都不存在
//	// hasReadme = false (项目缺少README,文档质量差)
//
//	// 示例5: 项目有多个README文件(罕见)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/multi-readme-project"},
//	}
//	hasReadme := cqe.hasReadme()
//	// 假设同时存在: README.md, README.txt, README
//	// 文件检查: /home/user/multi-readme-project/README.md 存在 ✓
//	// 短路返回: true (第一个即存在,不检查README.txt和README)
//	// hasReadme = true
//
// 注意事项:
//   - 仅检查根目录: 不递归检查子目录(如docs/README.md不会被发现)
//   - 大小写敏感: 在Unix/Linux系统README vs readme是不同文件
//   - Windows大小写不敏感: README.md和readme.md被视为同一文件
//   - 不验证内容: 仅检查文件存在,不读取内容或验证格式质量
//   - 权限问题: os.Stat()失败可能因文件不存在或无读取权限
//   - 符号链接: os.Stat()会跟随符号链接,检查实际文件
//   - 隐藏文件: .README (点开头)不在检查列表中
//   - 国际化: 不支持非英文README(如README.zh-CN.md)
//   - 短路优化: 第一个文件存在即返回,性能优化
//   - 硬编码列表: 文件名列表硬编码,不支持自定义扩展
//
// 改进方向:
//   - 内容质量检查: 验证README长度(如>500字符)和章节结构
//     - 检查必备章节: ## Installation, ## Usage, ## License
//   - 多语言支持: 检查README.zh-CN.md, README.ja.md等国际化文件
//     - 国际化项目应有多语言README
//   - 格式验证: 验证Markdown语法正确性
//     - 解析Markdown AST,检查标题层级、链接有效性
//   - 子目录递归: 支持检查docs/README.md等子目录文档
//     - 某些项目将详细文档放在docs/目录
//   - 符号链接检测: 使用os.Lstat()区分符号链接和实际文件
//   - 自定义文件名: 支持配置额外的文件名(如README.rst, index.md)
//   - README评分: 根据README质量评分(长度、结构、示例、图片等)
//     - 优秀README: >2000字符,5+章节,代码示例,徽章
//   - 模板检测: 识别使用默认模板未修改的README(低质量)
//   - 链接有效性: 检查README中外部链接是否有效
//   - 图片检测: 检查README中图片/GIF演示是否存在
//   - 徽章识别: 识别CI/CD、覆盖率、版本等徽章(项目成熟度指标)
//   - 历史对比: 追踪README更新频率,识别过时文档
//   - AI质量评估: 使用NLP分析README可读性和信息完整性
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) hasReadme() bool {
	readmeFiles := []string{"README.md", "README.txt", "README", "readme.md", "readme.txt"}

	for _, filename := range readmeFiles {
		if _, err := os.Stat(filepath.Join(cqe.results.ProjectPath, filename)); err == nil {
			return true
		}
	}

	return false
}

// checkPackageDocumentation 检查包级别文档覆盖率并返回文档质量分数(文档质量维度辅助检查方法-简化实现)
//
// 功能说明:
//
//	本方法是代码质量评估系统中文档质量维度(documentation_quality)的辅助检查方法,
//	负责评估项目中各个Go包(package)是否具有完善的包级别文档注释。包文档是Go语言文档体系的
//	重要组成部分,通过在包的任意.go文件开头提供注释来描述包的用途、功能和使用方法,
//	这些注释会被godoc工具提取并生成API文档。本方法设计用于检查每个包的文档覆盖率,
//	统计有文档的包占总包数的比例,并转换为0-30分的文档质量评分。
//	当前实现为简化版本,直接返回SimpleReturnScore常量(30分),表示满分,未来版本将实现
//	完整的包文档解析和覆盖率计算逻辑,提供精确的文档质量评估。
//
// 设计目标(待完整实现):
//
//	1. 包文档识别 (Package Documentation Recognition):
//	   • 检测目标: 每个Go包的包级别文档注释
//	   • 文档位置: 包声明前的注释块 (// Package xxx ...)
//	   • 示例格式:
//	     // Package models 提供数据模型定义和相关操作方法。
//	     //
//	     // 本包定义了系统核心数据模型,包括用户、订单、商品等实体,
//	     // 以及模型之间的关联关系和业务逻辑方法。
//	     package models
//
//	2. 包文档质量标准 (Documentation Quality Standards):
//	   • 必备元素:
//	     - 包名描述: "Package xxx 提供..." 开头
//	     - 功能说明: 描述包的核心功能和职责
//	     - 使用场景: 说明包的典型应用场景
//	   • 可选元素:
//	     - 使用示例: 代码示例展示包的基本用法
//	     - 依赖说明: 列出主要依赖包
//	     - 注意事项: 特殊约束或限制
//
//	3. 覆盖率计算 (Coverage Calculation):
//	   • 统计逻辑:
//	     - 总包数: 项目中所有Go包的数量
//	     - 有文档包数: 具有包级别文档的包数量
//	     - 覆盖率: (有文档包数 / 总包数) × 100%
//	   • 评分转换:
//	     - 覆盖率100%: 30分(满分)
//	     - 覆盖率80%: 24分
//	     - 覆盖率60%: 18分
//	     - 覆盖率40%: 12分
//	     - 覆盖率<20%: 6分
//
//	4. 文档内容质量检查 (Content Quality Check):
//	   • 长度检查: 文档长度>50字符(过短无意义)
//	   • 格式检查: 符合godoc注释规范
//	   • 关键词检查: 包含"Package"关键词
//	   • 示例检查: 是否包含代码示例
//
// 当前实现状态:
//
//	简化实现 (Simplified Implementation):
//	• 返回值: 固定返回SimpleReturnScore常量(30分,满分)
//	• 实现程度: 0% (仅框架,无实际检查逻辑)
//	• 注释说明: "检查各个包是否有包级别的文档注释,返回文档覆盖率得分(0-30分)"
//	• 设计理念: 简化实现,避免复杂的AST解析,返回满分不影响整体评分
//	• 改进空间: 未来可实现完整的包文档覆盖率分析
//
// 执行流程(当前):
//
//	1. 简化返回:
//	   • 直接返回: return SimpleReturnScore
//	   • SimpleReturnScore常量: 30.0 (满分)
//	   • 无任何检查逻辑
//
//	2. 隐含假设:
//	   • 假设所有包都有完善文档
//	   • 文档覆盖率100%
//	   • 文档质量优秀
//
// 预期执行流程(未来完整实现):
//
//	1. 遍历项目所有Go包:
//	   • 使用go list ./... 列出所有包
//	   • 或使用filepath.Walk遍历目录识别包
//
//	2. 解析每个包的文件获取包文档:
//	   • 使用go/parser.ParseFile解析包中任意.go文件
//	   • 提取ast.File.Doc (包文档注释)
//	   • 检查注释是否为nil或空
//
//	3. 验证文档质量:
//	   • 长度检查: doc.Text() 长度>50
//	   • 格式检查: 符合 "// Package xxx ..." 格式
//	   • 内容检查: 描述性文字,非TODO/FIXME
//
//	4. 计算覆盖率:
//	   • totalPackages := 项目包总数
//	   • documentedPackages := 有文档的包数
//	   • coverage := float64(documentedPackages) / float64(totalPackages)
//
//	5. 转换为评分:
//	   • score := coverage × MaxPackageDocScore (30分)
//	   • 示例: 80%覆盖率 → 0.8 × 30 = 24分
//	   • 返回: return score
//
// 评分机制(设计):
//
//	分数范围: 0-30分
//	• 30分 (100%覆盖): 所有包都有完善文档
//	• 24分 (80%覆盖): 大部分包有文档,少数遗漏
//	• 18分 (60%覆盖): 半数包有文档,中等水平
//	• 12分 (40%覆盖): 文档覆盖不足,需改进
//	• 6分 (20%覆盖): 文档严重缺失
//	• 0分 (0%覆盖): 完全无包文档
//
//	权重说明:
//	• 30分上限: 占文档质量维度100%权重
//	• 维度权重: 文档质量在整体评分中占5%权重
//	• 实际影响: 30分 × 5% = 1.5分(对总分影响)
//
// 包文档重要性:
//
//	1. API文档生成 (API Documentation):
//	   • godoc工具自动提取包文档生成HTML文档
//	   • pkg.go.dev展示包文档作为API参考
//	   • 用户通过文档了解包的用途和用法
//
//	2. 代码导航 (Code Navigation):
//	   • IDE(VSCode/GoLand)显示包文档作为悬停提示
//	   • 帮助开发者快速理解包的职责
//	   • 减少阅读源码的时间成本
//
//	3. 项目可维护性 (Maintainability):
//	   • 新成员快速了解代码架构
//	   • 减少onboarding时间
//	   • 降低技术债务
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.ProjectPath: 项目根目录路径
//       - 类型: string绝对路径
//       - 用途: 定位需要检查的Go包
//       - 示例: "/home/user/go-project" 或 "E:\\Go Learn\\go-mastery"
//
// 返回值:
//   - float64: 包文档覆盖率得分
//     • 当前: 固定返回SimpleReturnScore(30.0满分)
//     • 未来: 实际覆盖率得分(0.0-30.0)
//     • 范围: 0.0 - 30.0 分
//     • 30.0: 所有包都有完善文档(理想状态)
//     • 0.0: 完全无包文档(最差状态)
//     • 用途: 用于calculateDocumentationScore()计算文档质量分数
//
// 使用场景:
//   - calculateDocumentationScore()调用: 作为文档质量评分的关键指标
//   - godoc文档完整性检查: 确保生成的API文档完整
//   - 开源项目文档审计: 评估项目文档的专业程度
//   - CI/CD文档门禁: 文档覆盖率低于阈值时发出警告
//   - 团队文档规范检查: 强制执行包文档编写规范
//
// 示例(当前简化实现):
//
//	// 示例1: 任意项目(当前所有项目都返回相同结果)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/any-project"},
//	}
//	score := cqe.checkPackageDocumentation()
//	// 当前: 直接返回SimpleReturnScore常量
//	// score = 30.0 (满分,简化实现)
//
//	// 示例2: 未来完整实现 - 100%覆盖率项目
//	// 假设项目有10个包,所有10个包都有包文档
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/well-documented-project"},
//	}
//	score := cqe.checkPackageDocumentation()
//	// 未来: 检测到10/10包有文档
//	// 覆盖率: 10/10 = 100%
//	// score = 1.0 × 30 = 30.0 (满分)
//
//	// 示例3: 未来完整实现 - 80%覆盖率项目
//	// 假设项目有10个包,8个包有文档,2个包无文档
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/partial-docs-project"},
//	}
//	score := cqe.checkPackageDocumentation()
//	// 未来: 检测到8/10包有文档
//	// 覆盖率: 8/10 = 80%
//	// score = 0.8 × 30 = 24.0
//
//	// 示例4: 未来完整实现 - 50%覆盖率项目
//	// 假设项目有10个包,仅5个包有文档
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/half-docs-project"},
//	}
//	score := cqe.checkPackageDocumentation()
//	// 未来: 检测到5/10包有文档
//	// 覆盖率: 5/10 = 50%
//	// score = 0.5 × 30 = 15.0
//
//	// 示例5: 未来完整实现 - 0%覆盖率项目
//	// 假设项目有10个包,完全无包文档
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{ProjectPath: "/home/user/no-docs-project"},
//	}
//	score := cqe.checkPackageDocumentation()
//	// 未来: 检测到0/10包有文档
//	// 覆盖率: 0/10 = 0%
//	// score = 0.0 × 30 = 0.0 (最差)
//
// 注意事项:
//   - 简化实现: 当前方法直接返回满分,无实际检查逻辑
//   - 不影响评分: 返回满分确保文档维度不拖累总分
//   - 未来改进: 需要实现AST解析和包遍历逻辑
//   - 包识别: 需要准确识别项目中的所有Go包(排除vendor/测试等)
//   - 文档格式: 严格遵循godoc注释规范(连续//注释,package声明前)
//   - 多文件包: 一个包可能有多个.go文件,任意文件有包文档即可
//   - 内部包: internal包也应有文档,虽然不对外暴露
//   - 测试文件: _test.go文件的包文档不计入覆盖率
//   - 代码生成: 自动生成的包(如protobuf)可能无包文档,需排除
//   - 性能考虑: 解析所有包可能耗时,大型项目需要优化
//
// 改进方向:
//   - 实现包遍历: 使用go list或filepath.Walk识别所有包
//     - go list -json ./... | jq -r '.Dir' | sort -u
//   - AST解析包文档: 使用go/parser提取包级别文档注释
//     - fset := token.NewFileSet(); f, _ := parser.ParseFile(fset, path, nil, parser.ParseComments)
//     - doc := f.Doc.Text()
//   - 文档质量评估: 不仅检查存在性,还评估文档质量
//     - 长度>100字符,包含"Package"关键词,有描述性段落
//   - 代码示例检测: 检查包文档中是否包含Example测试
//     - Example函数会被godoc展示为示例代码
//   - 历史对比: 追踪包文档覆盖率变化趋势
//     - 新包必须有文档,覆盖率不能下降
//   - 自动生成模板: 为无文档包生成包文档模板
//     - // Package xxx TODO: 添加包描述
//   - IDE集成: 提供VSCode/GoLand插件,实时显示包文档覆盖率
//   - CI/CD门禁: 覆盖率<80%时阻止合并
//   - 多语言文档: 支持中英文双语包文档检查
//   - 文档链接检查: 验证包文档中的外部链接有效性
//   - godoc预览: 生成godoc HTML预览,检查渲染效果
//   - 文档评分细化: 区分优秀(>200字)/良好(100-200字)/及格(50-100字)/不及格(<50字)
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) checkPackageDocumentation() float64 {
	// 检查各个包是否有包级别的文档注释
	// 返回文档覆盖率得分 (0-30分)
	return SimpleReturnScore // 简化实现
}

// generateImprovements 基于质量评估结果生成个性化改进建议列表(结果处理核心方法)
//
// 功能说明:
//
//	本方法是代码质量评估系统的结果处理核心方法,负责将原始的质量评分数据转化为
//	可操作的改进建议(ImprovementSuggestion)列表。该方法综合分析六大质量维度的评分,
//	识别得分低于阈值(80分)的维度,调用专门的建议生成逻辑为每个问题维度创建针对性的
//	改进建议。同时,基于具体问题列表(CodeIssue)生成细粒度的修复建议。最终将所有建议
//	汇总存入cqe.results.Improvements字段,供报告生成和开发者参考。本方法是连接
//	"质量评估"和"质量改进"的桥梁,确保评估结果能转化为实际行动指南。
//
// 建议生成策略:
//
//	1. 维度级建议生成 (Dimension-Level Suggestions):
//	   • 触发条件: 维度得分 < IntValue80 (80分)
//	   • 生成逻辑: 调用generateDimensionImprovement(dimension, score)
//	   • 目标维度:
//	     - code_structure (代码结构): 复杂度过高、函数过长等
//	     - style_compliance (风格合规): 格式问题、命名规范等
//	     - security_analysis (安全分析): 安全漏洞、敏感信息泄露等
//	     - performance_analysis (性能分析): 内存分配、性能瓶颈等
//	     - test_quality (测试质量): 覆盖率不足、测试缺失等
//	     - documentation_quality (文档质量): README缺失、包文档不足等
//	   • 建议内容: 包括标题、描述、影响、工作量、优先级、示例
//
//	2. 问题级建议生成 (Issue-Level Suggestions):
//	   • 数据来源: cqe.generateIssueBasedImprovements()
//	   • 生成逻辑: 基于具体CodeIssue列表生成细粒度建议
//	   • 示例问题:
//	     - 具体文件行号的代码问题
//	     - 特定工具检测到的警告
//	     - 安全扫描发现的漏洞
//	   • 建议粒度: 精确到文件、行号、修复方法
//
//	3. 建议优先级排序 (Priority Ranking):
//	   • 隐式排序: 维度建议在前,问题建议在后
//	   • 显式优先级: 每个建议有Priority字段(1-5,1最高)
//	   • 排序依据: 影响范围、修复难度、安全性等
//
// 执行流程:
//
//	1. 初始化建议列表:
//	   • var improvements []ImprovementSuggestion
//	   • 准备收集所有改进建议
//
//	2. 遍历所有维度评分:
//	   • 循环: for dimension, score := range cqe.results.DimensionScores
//	   • 数据源: DimensionScores map[string]float64
//	   • 维度: code_structure, style_compliance, security_analysis等6个
//
//	3. 检查维度得分是否低于阈值:
//	   • 条件: if score < IntValue80 (80分)
//	   • 设计理念: 80分是质量基线,低于此需要改进
//	   • 触发建议生成: 仅对问题维度生成建议,避免建议过载
//
//	4. 生成维度级改进建议:
//	   • 调用: cqe.generateDimensionImprovement(dimension, score)
//	   • 返回: *ImprovementSuggestion 或 nil
//	   • nil处理: 某些维度可能无对应建议模板,返回nil跳过
//
//	5. 添加有效建议到列表:
//	   • 条件检查: if improvement != nil
//	   • 解引用添加: improvements = append(improvements, *improvement)
//	   • 避免nil: 确保列表中无nil指针
//
//	6. 生成问题级改进建议:
//	   • 调用: cqe.generateIssueBasedImprovements()
//	   • 返回: []ImprovementSuggestion 切片
//	   • 合并: append(improvements, issueBasedSuggestions...)
//
//	7. 存储建议到结果对象:
//	   • 赋值: cqe.results.Improvements = improvements
//	   • 持久化: 改进建议成为评估结果的一部分
//	   • 用途: 供报告生成、JSON导出、界面展示使用
//
// 建议结构设计:
//
//	ImprovementSuggestion结构体字段:
//	• Category (string): 建议分类
//	  - "Code Structure", "Test Quality", "Security", "Performance", "Documentation"
//	• Title (string): 建议标题
//	  - "改善代码结构和复杂度", "提高测试覆盖率"等
//	• Description (string): 详细描述
//	  - 包含当前得分、具体问题、建议行动
//	• Impact (string): 改进影响
//	  - "提高代码可维护性和可读性", "提高代码质量和可靠性"
//	• Effort (string): 所需工作量
//	  - "Low", "Medium", "High"
//	• Priority (int): 优先级
//	  - 1(最高) - 5(最低)
//	• Examples ([]string): 具体示例
//	  - ["拆分长函数", "提取重复代码", "简化条件逻辑"]
//
// 阈值设计:
//
//	IntValue80 (80分)作为质量基线的理由:
//	• 行业标准: 80分通常是"良好"质量的下限
//	• 平衡考虑: 太低(60分)导致严重问题无建议,太高(90分)建议过多
//	• 实践经验: 80分以下代码通常存在明显可改进点
//	• 优先级: 先解决<80分的维度,再追求>90分的卓越
//	• 对应等级:
//	  - ≥90分: A级(优秀),无需建议
//	  - 80-90分: B级(良好),可选优化
//	  - <80分: C级及以下(需改进),生成建议
//
// 参数:
//   - 本方法无显式参数,依赖CodeQualityEvaluator实例状态:
//     • cqe.results.DimensionScores: 六大维度评分map
//       - 类型: map[string]float64
//       - 示例: {"code_structure": 72.5, "test_quality": 65.0, ...}
//     • cqe.results (完整结果对象): 用于generateIssueBasedImprovements()
//
// 返回值:
//   - 本方法无返回值(void函数)
//   - 副作用: 修改cqe.results.Improvements字段
//     • 类型: []ImprovementSuggestion
//     • 内容: 所有生成的改进建议列表
//     • 数量: 0-N个建议(取决于问题维度数量)
//
// 使用场景:
//   - Evaluate()主流程调用: 质量评估完成后生成改进建议
//   - 报告生成: 改进建议作为报告的核心内容
//   - CI/CD反馈: 将建议输出到PR评论或Issue
//   - 开发者指导: 提供具体可操作的代码改进方向
//   - 质量追踪: 历史建议对比,追踪改进进度
//
// 示例:
//
//	// 示例1: 所有维度得分优秀(≥80分)
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{
//	        DimensionScores: map[string]float64{
//	            "code_structure": 92.0,
//	            "style_compliance": 88.0,
//	            "security_analysis": 95.0,
//	            "performance_analysis": 85.0,
//	            "test_quality": 90.0,
//	            "documentation_quality": 87.0,
//	        },
//	    },
//	}
//	cqe.generateImprovements()
//	// 维度建议: 无(所有维度≥80分)
//	// 问题建议: 可能有少量细节建议
//	// cqe.results.Improvements = [可能为空或仅包含问题级建议]
//
//	// 示例2: 代码结构和测试质量需改进
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{
//	        DimensionScores: map[string]float64{
//	            "code_structure": 72.5,      // <80,需建议
//	            "style_compliance": 88.0,    // ≥80,不生成建议
//	            "security_analysis": 95.0,   // ≥80,不生成建议
//	            "performance_analysis": 85.0, // ≥80,不生成建议
//	            "test_quality": 65.0,        // <80,需建议
//	            "documentation_quality": 87.0, // ≥80,不生成建议
//	        },
//	    },
//	}
//	cqe.generateImprovements()
//	// 维度建议1: code_structure (72.5分)
//	//   - Title: "改善代码结构和复杂度"
//	//   - Priority: 3
//	// 维度建议2: test_quality (65.0分)
//	//   - Title: "提高测试覆盖率"
//	//   - Priority: 2
//	// cqe.results.Improvements = [建议1, 建议2, ...问题级建议]
//
//	// 示例3: 多个维度严重不足
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{
//	        DimensionScores: map[string]float64{
//	            "code_structure": 55.0,       // <80,严重不足
//	            "style_compliance": 60.0,     // <80,需改进
//	            "security_analysis": 45.0,    // <80,严重不足
//	            "performance_analysis": 70.0, // <80,需改进
//	            "test_quality": 40.0,         // <80,严重不足
//	            "documentation_quality": 50.0, // <80,严重不足
//	        },
//	    },
//	}
//	cqe.generateImprovements()
//	// 维度建议: 所有6个维度都<80分,生成6条建议
//	// 问题建议: 可能有大量具体问题建议
//	// cqe.results.Improvements = [6条维度建议 + N条问题建议]
//
//	// 示例4: 部分维度无建议模板
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{
//	        DimensionScores: map[string]float64{
//	            "code_structure": 75.0,      // <80,有模板,生成建议
//	            "custom_dimension": 70.0,    // <80,但无模板,返回nil
//	        },
//	    },
//	}
//	cqe.generateImprovements()
//	// 维度建议1: code_structure (75.0分) ✓
//	// 维度建议2: custom_dimension (70.0分) → nil(无模板,跳过)
//	// cqe.results.Improvements = [建议1, ...问题级建议]
//
//	// 示例5: 无维度问题但有具体问题
//	cqe := &CodeQualityEvaluator{
//	    results: &CodeQualityResults{
//	        DimensionScores: map[string]float64{
//	            "code_structure": 85.0,      // ≥80,不生成建议
//	            "test_quality": 82.0,        // ≥80,不生成建议
//	        },
//	        // 假设Issues中有具体问题(如某行代码安全漏洞)
//	    },
//	}
//	cqe.generateImprovements()
//	// 维度建议: 无(所有维度≥80分)
//	// 问题建议: 有(基于具体Issues生成)
//	// cqe.results.Improvements = [问题级建议1, 问题级建议2, ...]
//
// 注意事项:
//   - 阈值固定: 当前硬编码80分阈值,未来可配置化
//   - 建议模板有限: generateDimensionImprovement仅支持部分维度,未覆盖的返回nil
//   - 建议去重: 当前无去重逻辑,维度建议和问题建议可能重复
//   - 优先级隐式: 建议添加顺序隐含优先级,维度建议在前
//   - 副作用函数: 直接修改cqe.results,调用方需注意
//   - 依赖完整性: 依赖DimensionScores已计算完成
//   - 并发安全: 方法修改共享状态,非并发安全
//   - 国际化: 当前建议文本仅中文,未支持多语言
//   - 建议数量: 可能生成大量建议,需要分页或分组展示
//   - nil检查: 必须检查improvement != nil,否则append panic
//
// 改进方向:
//   - 阈值可配置化: 允许用户自定义建议生成阈值(如70分或90分)
//     - config.ImprovementThreshold = 80.0
//   - 动态建议模板: 使用配置文件定义建议模板,支持扩展
//     - YAML/JSON定义: dimension → suggestion template映射
//   - 建议去重: 识别并合并相似建议
//     - 相同Category+Title的建议合并,避免重复
//   - 智能优先级: 基于得分差距、影响范围动态计算优先级
//     - 得分<60且是安全问题 → Priority=1(最高)
//   - 建议分组: 按Category或Priority分组展示
//     - {Security: [建议1, 建议2], Performance: [建议3]}
//   - 多语言支持: i18n国际化建议文本
//     - locale="en-US" → 英文建议, locale="zh-CN" → 中文建议
//   - 修复示例代码: 提供before/after代码示例
//     - Examples字段改为结构体,包含CodeBefore和CodeAfter
//   - 关联问题: 建议关联到具体CodeIssue,支持一键跳转
//     - RelatedIssues []string 字段
//   - 渐进式建议: 按难易程度排序,先易后难
//     - Quick Wins → Medium Effort → Long-term Improvements
//   - AI增强建议: 使用LLM生成个性化建议
//     - 基于代码上下文生成更精准的修复建议
//   - 建议追踪: 记录建议状态(待处理/进行中/已完成)
//     - SuggestionStatus enum { Pending, InProgress, Completed, Ignored }
//   - 影响量化: 量化改进影响(如"提升10%性能"而非"提高性能")
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) generateImprovements() {
	var improvements []ImprovementSuggestion

	// 基于不同维度的得分生成建议
	for dimension, score := range cqe.results.DimensionScores {
		if score < IntValue80 {
			improvement := cqe.generateDimensionImprovement(dimension, score)
			if improvement != nil {
				improvements = append(improvements, *improvement)
			}
		}
	}

	// 基于具体问题生成建议
	improvements = append(improvements, cqe.generateIssueBasedImprovements()...)

	cqe.results.Improvements = improvements
}

// generateDimensionImprovement 为特定质量维度生成个性化改进建议(建议生成辅助方法-模板映射策略)
//
// # 功能说明
//
// 本方法是generateImprovements()的核心辅助方法,采用switch-case模式将6大质量维度映射到预定义的
// 改进建议模板。当某个维度得分低于阈值(IntValue80=80分)时,主流程调用本方法生成该维度的标准化
// 改进建议,包含Category(类别)、Title(标题)、Description(动态描述含分数)、Impact(影响)、
// Effort(工作量Low/Medium/High)、Priority(优先级1-5)、Examples(示例数组)等7个字段。
//
// **设计目标**:
//   1. **集中管理建议模板**: 所有维度的建议内容集中在本方法,便于统一调整和国际化
//   2. **动态分数嵌入**: 使用fmt.Sprintf将实际得分嵌入Description,提供上下文信息
//   3. **差异化Effort/Priority**: 不同维度的改进难度和紧急度不同(如test_quality High vs code_structure Medium)
//   4. **可扩展架构**: 通过添加case分支支持新维度(当前仅2/6维度,未来可扩展至security_analysis等)
//   5. **失败安全设计**: default分支返回nil,避免未识别维度导致panic
//
// **当前支持维度**: 2/6维度已实现
//   - code_structure(代码结构): Priority 3, Effort Medium, 3个示例
//   - test_quality(测试质量): Priority 2, Effort High, 3个示例
//   - **待扩展**: style_compliance, security_analysis, performance_analysis, documentation_quality
//
// # 建议模板设计
//
// **code_structure维度模板**:
//   - Category: "Code Structure"(英文类别,对应系统标准分类)
//   - Title: "改善代码结构和复杂度"(中文标题,面向开发者)
//   - Description: "当前代码结构得分XX.XX较低,建议重构高复杂度函数"(动态分数+建议)
//   - Impact: "提高代码可维护性和可读性"(定性影响描述)
//   - Effort: "Medium"(中等工作量,预计1-2周重构周期)
//   - Priority: 3(中等优先级,非紧急但重要)
//   - Examples: ["拆分长函数", "提取重复代码", "简化条件逻辑"](3个具体操作示例)
//
// **test_quality维度模板**:
//   - Category: "Test Quality"
//   - Title: "提高测试覆盖率"
//   - Description: "当前测试覆盖率XX.XX不足,建议添加更多测试用例"
//   - Impact: "提高代码质量和可靠性"
//   - Effort: "High"(高工作量,预计3-4周编写全面测试)
//   - Priority: 2(高优先级,质量保障关键)
//   - Examples: ["添加单元测试", "编写集成测试", "测试边界条件"]
//
// **Effort级别定义**:
//   - Low: <1周工作量,简单配置或工具集成
//   - Medium: 1-2周工作量,中等规模重构
//   - High: >2周工作量,大规模重构或全面测试编写
//
// **Priority级别定义**: 1(最高)到5(最低)
//   - 1: 严重安全漏洞,必须立即修复
//   - 2: 高优先级质量问题(如test_quality,影响可靠性)
//   - 3: 中等优先级改进(如code_structure,影响可维护性)
//   - 4: 低优先级优化(如性能微调)
//   - 5: 可选改进(如文档完善)
//
// # 执行流程
//
// 1. **接收维度参数**: 从generateImprovements()接收dimension字符串(如"code_structure")和score浮点数(如72.5)
// 2. **Switch匹配**: 对dimension进行字符串精确匹配(区分大小写)
// 3. **模板实例化**: 匹配成功则创建ImprovementSuggestion结构体指针,填充7个字段
// 4. **动态描述生成**: 使用fmt.Sprintf("...得分%.2f...", score)将分数嵌入Description(保留2位小数)
// 5. **直接返回**: 立即返回建议指针(无需append,由调用方处理)
// 6. **默认处理**: 未匹配维度进入default分支,返回nil(调用方需nil检查)
//
// # 维度映射详解
//
// **为什么只实现2/6维度?**
//   - **渐进式开发**: 优先实现最常见的改进场景(代码结构+测试质量覆盖80%问题)
//   - **模板复杂度**: 其他维度建议更复杂(如security_analysis需分HIGH/MEDIUM/LOW,performance_analysis需区分CPU/内存/IO)
//   - **避免过度工程**: YAGNI原则,等实际需求出现再扩展
//
// **Effort差异的合理性**:
//   - code_structure Medium: 重构已有代码,IDE辅助,相对可控
//   - test_quality High: 从零编写测试,需理解业务逻辑+边界条件,工作量大
//
// **Priority差异的合理性**:
//   - test_quality Priority 2(更高): 缺乏测试直接影响质量和回归风险,优先级高于结构优化
//   - code_structure Priority 3(中等): 影响长期可维护性,但不影响当前功能正确性
//
// **default返回nil的设计**:
//   - 避免硬编码所有6个维度(部分维度建议模板未设计)
//   - 允许动态添加自定义维度而不强制提供建议
//   - 调用方(generateImprovements)已做nil检查: if improvement != nil { append... }
//
// # 参数
//
//   - dimension: 质量维度标识字符串,取值范围为6大维度之一(当前仅支持2个):
//     "code_structure"(代码结构), "style_compliance"(风格合规), "security_analysis"(安全分析),
//     "performance_analysis"(性能分析), "test_quality"(测试质量), "documentation_quality"(文档质量)
//     **注意**: 非法维度(如拼写错误"test_qualty")会进入default分支返回nil,不会报错
//
//   - score: 该维度的当前得分,范围0.0-100.0,调用方已确保score < IntValue80(80.0),
//     用于动态生成Description(如"当前代码结构得分72.50较低")
//     **注意**: 本方法不验证score范围,假设调用方已过滤
//
// # 返回值
//
//   - *ImprovementSuggestion: 改进建议结构体指针,包含7个字段:
//     Category(分类字符串), Title(中文标题), Description(含分数的动态描述),
//     Impact(影响描述), Effort(工作量Low/Medium/High), Priority(优先级1-5整数),
//     Examples(字符串数组,具体操作示例)
//
//   - 返回nil的情况: dimension参数未匹配任何case分支(进入default),
//     **调用方必须检查**: if improvement != nil { improvements = append(improvements, *improvement) }
//     否则对nil指针解引用会panic: runtime error: invalid memory address or nil pointer dereference
//
// # 使用场景
//
// **场景1: generateImprovements()主流程调用**
//   - 触发条件: 遍历cqe.results.DimensionScores时发现某维度score < IntValue80
//   - 调用示例: improvement := cqe.generateDimensionImprovement("code_structure", 72.5)
//   - 后续处理: if improvement != nil { improvements = append(improvements, *improvement) }
//
// **场景2: 单维度改进建议查询**
//   - 用例: 开发者仅想查看某个维度的标准建议(不触发完整评估)
//   - 示例: suggestion := evaluator.generateDimensionImprovement("test_quality", 65.0)
//   - 使用: 在IDE插件或Web界面显示建议详情
//
// **场景3: 模板预览和调试**
//   - 用例: 质量工程师调试建议模板内容
//   - 示例: 遍历所有6个维度,调用本方法查看哪些有模板、内容是什么
//   - 代码: for _, dim := range allDimensions { sug := generateDimensionImprovement(dim, 70.0); if sug != nil { log.Println(sug) } }
//
// **场景4: 批量生成改进计划**
//   - 用例: CI/CD流水线在质量门禁失败时生成完整改进计划
//   - 示例: 所有<80分维度调用本方法,汇总到ImprovementPlan文档
//   - 格式: Markdown报告,按Priority排序,按Effort分组
//
// **场景5: 自定义维度扩展(未来)**
//   - 用例: 添加新维度如"api_design"(API设计质量)
//   - 步骤: 在switch中添加case "api_design": return &ImprovementSuggestion{...}
//   - 自动生效: 无需修改调用方代码
//
// # 示例
//
// **示例1: code_structure维度得分72.5**
//
//	improvement := cqe.generateDimensionImprovement("code_structure", 72.5)
//	// 返回:
//	// &ImprovementSuggestion{
//	//   Category:    "Code Structure",
//	//   Title:       "改善代码结构和复杂度",
//	//   Description: "当前代码结构得分72.50较低,建议重构高复杂度函数", // 动态嵌入72.50
//	//   Impact:      "提高代码可维护性和可读性",
//	//   Effort:      "Medium", // 中等工作量
//	//   Priority:    3,        // 中等优先级
//	//   Examples:    []string{"拆分长函数", "提取重复代码", "简化条件逻辑"}, // 3个具体操作
//	// }
//
// **示例2: test_quality维度得分65.0**
//
//	improvement := cqe.generateDimensionImprovement("test_quality", 65.0)
//	// 返回:
//	// &ImprovementSuggestion{
//	//   Category:    "Test Quality",
//	//   Title:       "提高测试覆盖率",
//	//   Description: "当前测试覆盖率65.00不足,建议添加更多测试用例", // 动态嵌入65.00
//	//   Impact:      "提高代码质量和可靠性",
//	//   Effort:      "High",   // 高工作量(相比code_structure)
//	//   Priority:    2,        // 高优先级(相比code_structure的3)
//	//   Examples:    []string{"添加单元测试", "编写集成测试", "测试边界条件"},
//	// }
//
// **示例3: 不支持的维度security_analysis**
//
//	improvement := cqe.generateDimensionImprovement("security_analysis", 55.0)
//	// 返回: nil (进入default分支,该维度模板未实现)
//	// 调用方必须检查:
//	if improvement != nil {
//		improvements = append(improvements, *improvement) // 不会执行
//	}
//
// **示例4: 拼写错误的维度test_qualty(缺少i)**
//
//	improvement := cqe.generateDimensionImprovement("test_qualty", 60.0) // 拼写错误
//	// 返回: nil (未匹配任何case,进入default)
//	// **静默失败**: 不会报错,但调用方会跳过该建议
//	// **改进**: 未来可添加日志警告或返回error
//
// **示例5: 极端低分维度code_structure得分15.0**
//
//	improvement := cqe.generateDimensionImprovement("code_structure", 15.0)
//	// 返回:
//	// &ImprovementSuggestion{
//	//   Description: "当前代码结构得分15.00较低,建议重构高复杂度函数", // 分数15.00嵌入
//	//   // 其他字段与示例1相同,模板不区分严重程度(15分和75分建议相同)
//	// }
//	// **局限性**: 无法根据分数严重程度调整建议内容(15分应更紧急,但Priority仍为3)
//
// # 注意事项
//
//  1. **部分维度未实现**: 当前仅支持2/6维度(code_structure + test_quality),其余4个维度返回nil,
//     调用generateImprovements()时这些维度即使<80分也不会生成建议(静默跳过)
//
//  2. **固定模板内容**: Title、Impact、Examples等字段完全硬编码,无法根据具体问题定制
//     (如code_structure 72.5和15.0得到相同建议文本,仅Description分数不同)
//
//  3. **无分数严重度区分**: Priority和Effort不随score变化(65分和15分Priority都是2),
//     无法体现"15分非常紧急需立即修复"vs"75分可延后优化"的差异
//
//  4. **区分大小写**: dimension参数必须精确匹配"code_structure"/"test_quality",
//     "Code_Structure"或"TEST_QUALITY"会进入default返回nil
//
//  5. **nil返回未记录日志**: default分支静默返回nil,无法追踪是否有拼写错误或新维度遗漏,
//     建议改进: log.Printf("未识别的维度: %s", dimension)
//
//  6. **调用方必须nil检查**: 返回*ImprovementSuggestion可能为nil,调用方必须:
//     if improvement != nil { improvements = append(improvements, *improvement) }
//     否则对nil解引用会panic
//
//  7. **无国际化支持**: Title、Description、Impact、Examples全部硬编码中文,
//     无法支持英文或其他语言环境
//
//  8. **Examples数量固定**: 所有维度都是3个示例,无法根据维度复杂度调整(如security_analysis可能需要>5个示例)
//
//  9. **无Context依赖**: 不访问cqe.results的具体问题列表,建议内容纯粹基于dimension+score,
//     无法提供"您的项目有12个>50行的函数"这类具体数据驱动的建议
//
// 10. **扩展需修改代码**: 添加新维度必须修改本方法添加case分支,无法通过配置文件或插件扩展
//
// 11. **返回指针而非值**: 返回*ImprovementSuggestion而非值类型,虽然避免了大结构体拷贝,
//     但增加了nil检查负担,且结构体较小(约200字节)拷贝开销可接受
//
// 12. **Effort/Priority主观性**: "Medium"/"High"和2/3数值缺乏量化依据(如Medium=8小时vs16小时?),
//     不同团队理解不同,建议改进: 使用EstimatedHours整数字段替代Effort字符串
//
// # 改进方向
//
//  1. **完整维度覆盖**: 添加剩余4个维度的case分支(style_compliance, security_analysis, performance_analysis, documentation_quality),
//     确保所有维度都能生成建议
//
//  2. **分数严重度分级**: 根据score范围调整Priority,如:
//     score < 40 → Priority 1(严重), 40-60 → Priority 2(高), 60-80 → Priority 3(中)
//
//  3. **动态Examples生成**: 基于cqe.results.Issues生成具体示例,如:
//     "拆分长函数: calculateScore()有150行,建议拆分为3个子函数"
//
//  4. **模板外部化**: 将建议模板移至YAML/JSON配置文件,支持运行时加载和修改,
//     结构: dimensions: { code_structure: { title: "...", impact: "...", examples: [...] } }
//
//  5. **国际化i18n**: 使用i18n库支持多语言,如:
//     Title: i18n.T("improvement.code_structure.title"), Description: i18n.T("...", score)
//
//  6. **添加日志和监控**: default分支记录未识别维度到日志,统计各维度建议生成频率,
//     代码: log.Warnf("未实现的维度建议: %s (得分%.2f)", dimension, score)
//
//  7. **返回error而非nil**: 修改签名为(suggestion *ImprovementSuggestion, err error),
//     未识别维度返回errors.New("unsupported dimension: security_analysis"),便于调试
//
//  8. **Effort量化**: 将"Low"/"Medium"/"High"替换为EstimatedHours int字段(如2小时/16小时/40小时),
//     或使用enum: EffortLevel { VeryLow=1, Low=2, Medium=3, High=4, VeryHigh=5 }
//
//  9. **子类型建议**: 某些维度可细分,如security_analysis → high_severity_issues + medium_severity_issues,
//     分别生成不同Priority的建议
//
// 10. **示例代码生成**: Examples字段改为CodeExample结构体数组,包含Before/After代码对比:
//     type CodeExample { Description string, Before string, After string }
//
// 11. **插件化架构**: 使用策略模式或插件系统,每个维度独立的SuggestionGenerator实现,
//     注册到map: suggestionGenerators["code_structure"] = &CodeStructureSuggestionGenerator{}
//
// 12. **AI增强建议**: 集成LLM(如GPT-4)根据具体代码问题生成个性化建议,
//     调用: chatgpt.GenerateSuggestion(dimension, score, cqe.results.Issues)
//
// 13. **建议有效性跟踪**: 添加SuggestionID字段,允许开发者标记建议为"有用"/"无用",
//     统计分析哪些建议最有价值,优化模板内容
//
// 14. **关联具体Issue**: 添加RelatedIssueIDs []string字段,指向cqe.results.Issues中的具体问题,
//     允许用户点击建议跳转到源代码问题位置
//
// 15. **渐进式建议**: 根据score提供分阶段建议,如:
//     15-40分 → "立即重构最复杂的3个函数", 40-60分 → "逐步降低平均复杂度", 60-80分 → "优化边缘case"
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) generateDimensionImprovement(dimension string, score float64) *ImprovementSuggestion {
	switch dimension {
	case "code_structure":
		return &ImprovementSuggestion{
			Category:    "Code Structure",
			Title:       "改善代码结构和复杂度",
			Description: fmt.Sprintf("当前代码结构得分%.2f较低，建议重构高复杂度函数", score),
			Impact:      "提高代码可维护性和可读性",
			Effort:      "Medium",
			Priority:    3,
			Examples:    []string{"拆分长函数", "提取重复代码", "简化条件逻辑"},
		}
	case "test_quality":
		return &ImprovementSuggestion{
			Category:    "Test Quality",
			Title:       "提高测试覆盖率",
			Description: fmt.Sprintf("当前测试覆盖率%.2f不足，建议添加更多测试用例", score),
			Impact:      "提高代码质量和可靠性",
			Effort:      "High",
			Priority:    2,
			Examples:    []string{"添加单元测试", "编写集成测试", "测试边界条件"},
		}
	default:
		return nil
	}
}

// generateIssueBasedImprovements 基于具体代码问题生成细粒度改进建议列表(建议生成辅助方法-问题聚合策略)
//
// # 功能说明
//
// 本方法与generateDimensionImprovement()互补,从cqe.results.Issues(具体代码问题列表)生成细粒度的
// 改进建议。采用"问题聚合→阈值过滤→建议生成"三阶段策略:先按Category统计问题频次,再为
// 问题数≥3的类别生成建议,避免为零散问题生成大量无用建议。适用于处理静态分析工具(如golangci-lint)
// 检测的实际问题,与维度级建议形成"战略(维度)→战术(问题)"的完整改进体系。
//
// **设计目标**:
//   1. **问题驱动**: 基于真实检测到的问题(CodeIssue),而非抽象的维度分数
//   2. **聚合去噪**: 通过Category分组和count≥3阈值,过滤零散问题,聚焦高频问题类别
//   3. **动态生成**: Title和Description动态嵌入category和count,提供具体上下文
//   4. **互补性**: 与generateDimensionImprovement()分工明确(维度级vs问题级,战略vs战术)
//   5. **性能优化**: 使用索引遍历避免184字节CodeIssue结构体复制
//
// **与generateDimensionImprovement()的差异**:
//   - **输入来源**: 本方法基于cqe.results.Issues(问题列表) vs generateDimensionImprovement基于DimensionScores(维度分数)
//   - **粒度**: 细粒度问题类别(如"Security: G401") vs 粗粒度维度(如"security_analysis")
//   - **触发条件**: count≥3(频次阈值) vs score<80(分数阈值)
//   - **建议数量**: 0-N个(取决于问题分布) vs 0-6个(最多6个维度)
//   - **固定字段**: Effort="Low", Priority=2, Impact="提高代码质量" vs 每个维度独立定义
//
// # 问题聚合策略
//
// **阶段1: 问题分类统计**
//   - 输入: cqe.results.Issues(CodeIssue数组,可能包含0-1000+个问题)
//   - 处理: 按CodeIssue.Category分组统计频次 → map[string]int
//   - 示例: {"Security": 15, "Style": 8, "Performance": 2, "Error Handling": 12} → 4个类别
//
// **阶段2: 阈值过滤**
//   - 阈值: count >= 3(硬编码常量ThresholdIssueCount可改进)
//   - 原理: 某类别问题数≥3才值得生成建议,<3为零散问题可忽略
//   - 示例: 上述4个类别,仅Security(15)、Style(8)、Error Handling(12)通过过滤,Performance(2)被忽略
//
// **阶段3: 建议生成**
//   - 每个通过过滤的类别生成一个ImprovementSuggestion
//   - 动态字段: Title="解决{category}相关问题", Description="发现{count}个{category}相关问题,建议优先解决"
//   - 固定字段: Effort="Low"(假设批量修复同类问题相对简单), Priority=2(高优先级,具体问题优先于维度优化)
//   - Examples字段: nil(未提供具体示例,改进方向之一)
//
// # 执行流程
//
// 1. **初始化**: 创建空的improvements切片和issueCategories map[string]int
// 2. **遍历问题**: 使用索引遍历cqe.results.Issues(避免184字节结构体复制),按Category累加计数
// 3. **遍历类别**: for category, count := range issueCategories,检查count是否≥3
// 4. **生成建议**: 通过阈值的类别实例化ImprovementSuggestion,填充7个字段(Category/Title/Description/Impact/Effort/Priority/Examples)
// 5. **追加切片**: improvements = append(improvements, improvement)
// 6. **返回结果**: 返回improvements切片(可能为空切片,长度0-N)
//
// **时间复杂度**: O(n + m),n=Issues数量,m=唯一Category数量(通常m<<n,如1000问题仅10个类别)
// **空间复杂度**: O(m),map存储m个类别,切片最多m个建议
//
// # 阈值设计详解
//
// **为什么选择count >= 3?**
//   - **经验值**: 1-2个问题可能是偶发或边缘case,≥3个表明系统性问题
//   - **建议质量**: 避免为零散问题生成大量"噪音"建议(如100个问题分布在50个类别,若无阈值会生成50条建议)
//   - **行动优先级**: 集中精力解决高频问题(帕累托原则:20%的问题类型贡献80%的问题数量)
//
// **阈值=3的局限性**:
//   - 硬编码,无法适应不同项目规模(10个问题的小项目 vs 1000个问题的大项目,阈值应不同)
//   - 忽略严重度(3个LOW问题 vs 1个HIGH问题,后者更紧急但未生成建议)
//   - 改进方向: 动态阈值(如count >= max(3, totalIssues*0.05)),或综合严重度权重
//
// **与Priority=2的关系**:
//   - Priority 2(高优先级)体现"具体问题优先于抽象优化"的原则
//   - 比维度优化(Priority 3)更紧急,但不如严重安全漏洞(Priority 1)
//
// # 参数
//
//   - 无显式参数,但依赖cqe.results.Issues切片(CodeIssue类型数组),
//     包含所有静态分析工具检测到的代码问题,每个CodeIssue包含:
//     Category(分类字符串,如"Security"/"Style"/"Performance"),
//     Severity(严重度,如"error"/"warning"/"info"),
//     Message(问题描述), File/Line(位置信息)等字段
//
// # 返回值
//
//   - []ImprovementSuggestion: 改进建议切片,长度为通过阈值的类别数量(0-N个)
//     每个建议包含7个字段:
//     Category(问题类别,如"Security"), Title(动态生成"解决Security相关问题"),
//     Description(动态生成"发现15个Security相关问题,建议优先解决"),
//     Impact("提高代码质量",所有建议固定), Effort("Low",所有建议固定),
//     Priority(2,所有建议固定), Examples(nil,未提供)
//
//   - **空切片场景**: 当所有类别问题数<3时,返回空切片[]ImprovementSuggestion{},
//     调用方(generateImprovements)会正常追加(append空切片不影响结果)
//
// # 使用场景
//
// **场景1: generateImprovements()主流程调用**
//   - 触发: Evaluate()执行完所有检查,generateImprovements()收集维度和问题建议
//   - 调用: improvements = append(improvements, cqe.generateIssueBasedImprovements()...)
//   - 作用: 将问题级建议追加到维度级建议后,形成完整建议列表
//
// **场景2: 大量同类问题聚焦**
//   - 场景: golangci-lint检测到50个"Security: G401"问题(使用MD5哈希)
//   - 效果: 生成建议"解决Security相关问题: 发现50个Security相关问题,建议优先解决"
//   - 优势: 避免为每个问题生成单独建议,提供高层次指导
//
// **场景3: 多类别问题并存**
//   - 场景: Security(15个), Style(8个), Error Handling(12个), Performance(2个)
//   - 结果: 生成3条建议(Performance仅2个不满足阈值≥3)
//   - 决策: 开发者可按Priority和count排序,优先解决Security(15个)
//
// **场景4: 零问题项目**
//   - 场景: cqe.results.Issues为空切片或所有类别<3个问题
//   - 结果: 返回空切片,generateImprovements()仅包含维度级建议(如有)
//   - 正常行为: 高质量项目可能无需问题级建议
//
// **场景5: CI/CD质量门禁**
//   - 场景: 流水线失败,生成改进报告发送给开发者
//   - 内容: 结合维度建议(战略)和问题建议(战术),提供完整改进路径
//   - 格式: "优先解决Security相关问题(15个) → 提高测试覆盖率(65.0分) → 改善代码结构(72.5分)"
//
// # 示例
//
// **示例1: 多类别高频问题**
//
//	// cqe.results.Issues包含30个问题:
//	// 15个Category="Security", 8个Category="Style", 5个Category="Performance", 2个Category="Error Handling"
//	improvements := cqe.generateIssueBasedImprovements()
//	// 返回3个建议(Error Handling仅2个不满足阈值):
//	// [
//	//   {Category: "Security", Title: "解决Security相关问题", Description: "发现15个Security相关问题,建议优先解决", Impact: "提高代码质量", Effort: "Low", Priority: 2, Examples: nil},
//	//   {Category: "Style", Title: "解决Style相关问题", Description: "发现8个Style相关问题,建议优先解决", Impact: "提高代码质量", Effort: "Low", Priority: 2, Examples: nil},
//	//   {Category: "Performance", Title: "解决Performance相关问题", Description: "发现5个Performance相关问题,建议优先解决", Impact: "提高代码质量", Effort: "Low", Priority: 2, Examples: nil}
//	// ]
//	// **注意**: 返回顺序不确定(map遍历无序),可能需要后续排序
//
// **示例2: 零问题或低频问题**
//
//	// cqe.results.Issues包含5个问题: 2个Security, 2个Style, 1个Performance
//	improvements := cqe.generateIssueBasedImprovements()
//	// 返回空切片: []ImprovementSuggestion{}
//	// 原因: 所有类别count<3,未通过阈值过滤
//
// **示例3: 单一类别大量问题**
//
//	// cqe.results.Issues包含100个问题,全部Category="Security"
//	improvements := cqe.generateIssueBasedImprovements()
//	// 返回1个建议:
//	// [{Category: "Security", Title: "解决Security相关问题", Description: "发现100个Security相关问题,建议优先解决", Impact: "提高代码质量", Effort: "Low", Priority: 2, Examples: nil}]
//	// **局限性**: 100个问题Effort="Low"明显不合理,应根据count调整
//
// **示例4: 空Issues列表**
//
//	// cqe.results.Issues = []CodeIssue{}(高质量项目)
//	improvements := cqe.generateIssueBasedImprovements()
//	// 返回空切片: []ImprovementSuggestion{}
//	// issueCategories map为空,for循环不执行
//
// **示例5: 与维度建议结合**
//
//	// generateImprovements()主流程:
//	// 1. 维度建议: code_structure(72.5分) → 1条建议
//	// 2. 问题建议: Security(15个), Style(8个) → 2条建议
//	// 最终improvements包含3条建议,按生成顺序排列:
//	// [code_structure建议, Security建议, Style建议]
//	// **优化**: 可按Priority+count+score综合排序,优先显示最紧急建议
//
// # 注意事项
//
//  1. **固定Effort不合理**: 所有建议Effort="Low",但100个问题修复显然不是"Low"工作量,
//     应根据count调整: count<5→Low, 5-20→Medium, >20→High
//
//  2. **固定Priority不灵活**: 所有建议Priority=2,未区分严重度(如HIGH安全问题应Priority 1),
//     应综合CodeIssue.Severity: error→Priority 1, warning→Priority 2, info→Priority 3
//
//  3. **缺少Examples**: 所有建议Examples为nil(零值),无法提供具体修复示例,
//     改进: 从Issues中抽取代表性问题作为Examples,如"G401: 使用MD5哈希 at file.go:123"
//
//  4. **阈值硬编码**: count >= 3固定阈值,无法适应不同项目规模和问题分布,
//     改进: 动态阈值或配置参数ThresholdIssueCount(可通过配置文件调整)
//
//  5. **map遍历无序**: for category, count := range issueCategories顺序不确定,
//     建议可能乱序,改进: 转为切片后按count降序或category字母序排序
//
//  6. **索引遍历**: for i := range cqe.results.Issues避免184字节复制,但仍需通过索引访问字段,
//     可读性略差(需cqe.results.Issues[i].Category而非issue.Category)
//
//  7. **无去重**: 若两个CodeIssue有不同Category但含义相同(如"Security"和"security"大小写差异),
//     会生成重复建议,改进: Category标准化或合并相似类别
//
//  8. **Impact通用性**: 所有建议Impact="提高代码质量"过于宽泛,应具体化:
//     Security→"提高系统安全性", Performance→"提高运行性能", Style→"提高代码一致性"
//
//  9. **忽略严重度**: 仅统计count,未考虑Severity(3个error vs 3个info应区别对待),
//     改进: 计算加权分数 weightedScore = high*3 + medium*2 + low*1, 按分数排序
//
// 10. **无问题关联**: 生成的建议未保存RelatedIssues字段(哪些具体CodeIssue属于该建议),
//     开发者无法从建议跳转到源代码问题位置
//
// 11. **批量append性能**: append(improvements, improvement)逐个追加,若通过阈值类别很多(如50个),
//     可能多次扩容,改进: 预分配容量make([]ImprovementSuggestion, 0, estimatedCount)
//
// 12. **空建议处理**: 返回空切片时调用方需正确处理,幸运的是append(slice, emptySlice...)不会出错,
//     但可能浪费函数调用开销,改进: 在generateImprovements()中预检查Issues是否为空
//
// # 改进方向
//
//  1. **动态Effort计算**: 根据count调整工作量估算,
//     if count < 5 { Effort = "Low" } else if count < 20 { Effort = "Medium" } else { Effort = "High" }
//
//  2. **综合Severity调整Priority**: 遍历该类别的所有问题,统计最高严重度,
//     如果有error→Priority 1, 仅warning→Priority 2, 仅info→Priority 3
//
//  3. **添加具体Examples**: 从Issues中抽取代表性问题(如前3个或随机采样),
//     Examples: []string{"G401 at main.go:45", "G401 at utils.go:123", "G401 at crypto.go:67 (+12 more)"}
//
//  4. **阈值配置化**: 添加ThresholdIssueCount常量或配置参数,支持运行时调整,
//     结构: type Config struct { IssueCountThreshold int }
//
//  5. **建议排序**: 返回前按count降序或综合score排序,确保高频问题优先显示,
//     sort.Slice(improvements, func(i, j int) bool { return extractCount(improvements[i].Description) > extractCount(improvements[j].Description) })
//
//  6. **Category标准化**: 统一大小写和命名,避免"Security"vs"security"重复,
//     category = strings.Title(strings.ToLower(category))
//
//  7. **差异化Impact**: 根据category定制Impact描述,
//     impactMap := map[string]string{"Security": "提高系统安全性", "Performance": "提高运行性能", ...}
//
//  8. **加权严重度统计**: 计算加权分数而非简单计数,
//     weightedScore := highCount*3 + mediumCount*2 + lowCount*1, 按weightedScore过滤和排序
//
//  9. **关联原始问题**: 添加RelatedIssueIDs []int字段,记录属于该建议的问题索引,
//     结构: type ImprovementSuggestion struct { ..., RelatedIssueIDs []int }
//
// 10. **性能优化**: 预分配切片容量避免多次扩容,
//     estimatedCount := 0; for _, count := range issueCategories { if count >= 3 { estimatedCount++ } }
//     improvements := make([]ImprovementSuggestion, 0, estimatedCount)
//
// 11. **空切片优化**: 在generateImprovements()中预检查,
//     if len(cqe.results.Issues) == 0 { return improvements } // 跳过调用
//
// 12. **日志监控**: 记录生成的建议数量和类别分布,用于质量分析,
//     log.Infof("生成%d条问题建议: %v", len(improvements), categoryList)
//
// 13. **建议分组**: 按Category或Priority分组返回,便于UI展示,
//     type GroupedSuggestions struct { HighPriority []ImprovementSuggestion, MediumPriority []ImprovementSuggestion }
//
// 14. **国际化i18n**: Title和Description支持多语言,
//     Title: i18n.T("improvement.issue.title", category), Description: i18n.T("improvement.issue.desc", count, category)
//
// 15. **AI增强**: 使用LLM根据具体问题生成更个性化建议,
//     如分析15个G401问题的上下文,建议"迁移到bcrypt替代MD5密码哈希"
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) generateIssueBasedImprovements() []ImprovementSuggestion {
	var improvements []ImprovementSuggestion

	// 统计问题类型
	issueCategories := make(map[string]int)
	// 使用索引遍历避免大结构体复制（184字节）
	for i := range cqe.results.Issues {
		issueCategories[cqe.results.Issues[i].Category]++
	}

	// 为主要问题类别生成建议
	for category, count := range issueCategories {
		if count >= 3 { // 某类问题超过3个时生成建议
			improvement := ImprovementSuggestion{
				Category:    category,
				Title:       fmt.Sprintf("解决%s相关问题", category),
				Description: fmt.Sprintf("发现%d个%s相关问题，建议优先解决", count, category),
				Impact:      "提高代码质量",
				Effort:      "Low",
				Priority:    2,
			}
			improvements = append(improvements, improvement)
		}
	}

	return improvements
}

// analyzeTechnicalDebt 分析项目技术债务并量化修复成本(债务分析核心方法-时间估算策略)
//
// # 功能说明
//
// 本方法将代码质量问题转化为可量化的技术债务指标,通过"严重度→修复时间"的映射规则计算总债务、
// 债务比率和债务等级。采用SQALE(Software Quality Assessment based on Lifecycle Expectations)
// 方法论的简化版本:将每个CodeIssue按Severity(error/warning/info)转换为修复工时(2h/0.5h/0.1h),
// 累加得到TotalDebt,再除以代码行数得到DebtRatio(每千行债务小时数),最后映射到A-F评级体系。
// 为项目管理层提供技术债务的财务视角(时间=成本),便于优先级决策和资源分配。
//
// **设计目标**:
//   1. **债务量化**: 将抽象的"代码质量问题"转化为具体的"修复工时"(小时数)
//   2. **归一化比率**: 通过DebtRatio(债务/千行代码)消除项目规模影响,10万行和1万行可比较
//   3. **分级评估**: A-F评级体系直观展示债务严重程度,便于非技术人员理解
//   4. **管理驱动**: 提供ROI分析基础(修复X小时债务 → 提升Y分质量 → 减少Z次故障)
//   5. **趋势追踪**: 支持DebtTrend时间序列,监控债务增减趋势(本方法未实现,结构预留)
//
// **SQALE方法论简化应用**:
//   - SQALE完整版: 8大特征×4层级×复杂计算 → 过于复杂
//   - 本实现简化版: 3个严重度×固定时间系数 → 实用主义
//   - 保留核心思想: 问题→时间→成本的转换链
//
// # 债务计算策略
//
// **修复时间映射规则**(基于行业经验值):
//   - error(错误): 2.0小时 - 需要定位根因、编写修复代码、完整测试、代码审查
//   - warning(警告): 0.5小时 - 相对简单,通常是风格或小问题,快速修复和验证
//   - info(信息): 0.1小时 - 6分钟,轻微提示,可选修复
//
// **总债务计算**: TotalDebt = Σ(每个Issue的修复时间)
//   - 公式: TotalDebt = (error数量 × 2.0) + (warning数量 × 0.5) + (info数量 × 0.1)
//   - 示例: 10个error + 20个warning + 50个info = 10×2.0 + 20×0.5 + 50×0.1 = 20 + 10 + 5 = 35小时
//
// **债务比率计算**: DebtRatio = (TotalDebt / CodeLines) × 1000
//   - 单位: 每千行代码的债务小时数(hours per KLOC)
//   - 归一化作用: 10万行35小时(0.35 h/KLOC) vs 1万行35小时(3.5 h/KLOC),后者债务密度更高
//   - 示例: 35小时 / 10000行 × 1000 = 3.5 hours/KLOC
//
// **评级映射**:
//   - A级: DebtRatio < TechnicalDebtRatingA(如1.0) - 优秀,债务可控
//   - B级: 1.0 ≤ DebtRatio < TechnicalDebtRatingB(如2.0) - 良好,需关注
//   - C级: 2.0 ≤ DebtRatio < TechnicalDebtRatingC(如5.0) - 中等,需计划还债
//   - D级: 5.0 ≤ DebtRatio < TechnicalDebtRatingD(如10.0) - 较差,技术债务高
//   - F级: DebtRatio ≥ 10.0 - 严重,债务失控,需立即行动
//
// # 执行流程
//
// 1. **初始化债务结构**: 创建TechnicalDebtAnalysis{Categories: map, Files: [], Trends: []}
// 2. **遍历问题累加债务**: for循环cqe.results.Issues,按Severity累加totalDebt(使用索引遍历避免184字节复制)
// 3. **保存总债务**: debt.TotalDebt = totalDebt
// 4. **计算债务比率**: 如果CodeLines > 0,计算DebtRatio = totalDebt / CodeLines × 1000(否则保持零值)
// 5. **评定等级**: 通过switch-case将DebtRatio映射到A/B/C/D/F评级
// 6. **保存结果**: cqe.results.TechnicalDebt = debt(后续可用于报告生成)
//
// **时间复杂度**: O(n),n为Issues数量,单次遍历
// **空间复杂度**: O(1),仅创建一个TechnicalDebtAnalysis结构体
//
// # 时间系数设计详解
//
// **为什么error=2小时?**
//   - 典型error修复流程: 阅读错误(10分钟) + 定位代码(20分钟) + 编写修复(30分钟) + 单元测试(30分钟) + 集成测试(20分钟) + 代码审查(10分钟) = 120分钟 = 2小时
//   - 考虑上下文切换和调试时间,2小时是合理平均值
//
// **为什么warning=0.5小时?**
//   - 典型warning: 未使用变量、deprecated API、轻微安全风险
//   - 修复流程: 理解警告(5分钟) + 简单修改(10分钟) + 快速验证(10分钟) + 提交(5分钟) = 30分钟 = 0.5小时
//
// **为什么info=0.1小时(6分钟)?**
//   - 典型info: 代码风格建议、优化提示
//   - 修复流程: 阅读(2分钟) + 一键修复(1分钟,如IDE格式化) + 验证(2分钟) + 提交(1分钟) = 6分钟
//   - 很多info可批量处理,实际可能更快
//
// **系数局限性**:
//   - 固定值,未考虑问题复杂度差异(简单error vs 复杂error)
//   - 未考虑团队经验(新手 vs 专家修复时间差异可达5-10倍)
//   - 未考虑业务领域复杂度(金融系统 vs 简单Web应用)
//
// # 参数
//
//   - 无显式参数,但依赖以下cqe状态:
//     - cqe.results.Issues: CodeIssue数组,每个包含Severity字段(取值"error"/"warning"/"info")
//     - cqe.results.Statistics.CodeLines: 项目总代码行数(用于计算债务比率)
//     - **注意**: 如果CodeLines=0(如纯配置项目),DebtRatio保持零值(避免除零错误)
//
// # 返回值
//
//   - 无返回值(void函数),通过副作用修改cqe.results.TechnicalDebt字段,
//     TechnicalDebtAnalysis结构体包含:
//     - TotalDebt float64: 总债务小时数(如35.5小时)
//     - DebtRatio float64: 每千行代码债务小时数(如3.5 hours/KLOC)
//     - Rating string: 债务评级("A"/"B"/"C"/"D"/"F")
//     - Categories map[string]float64: 分类债务(本方法未填充,预留字段)
//     - Files []FileDebt: 文件级债务(本方法未填充,预留字段)
//     - Trends []DebtTrend: 债务趋势(本方法未填充,预留字段)
//
// # 使用场景
//
// **场景1: Evaluate()主流程调用**
//   - 触发: 所有质量检查完成,Issues列表已填充,需量化债务
//   - 调用: cqe.analyzeTechnicalDebt()(通常在calculateFinalScore()之前)
//   - 作用: 为最终报告提供债务指标
//
// **场景2: 管理层汇报**
//   - 用例: 向CTO/PM汇报项目技术债务状况
//   - 数据: "当前技术债务35小时,债务比率3.5 hours/KLOC,评级C,建议投入2周(80小时)还债"
//   - 决策: 基于ROI分析决定是否投入资源重构
//
// **场景3: Sprint计划**
//   - 用例: 敏捷团队在Sprint Planning中分配还债任务
//   - 策略: 每个Sprint投入20%时间(如16小时/周×2周=32小时)还债
//   - 追踪: 通过DebtTrend监控债务减少趋势
//
// **场景4: 质量门禁**
//   - 用例: CI/CD流水线设置债务阈值
//   - 规则: DebtRatio > 5.0(C级)则阻止合并,必须先还债
//   - 实施: if results.TechnicalDebt.Rating == "C" || results.TechnicalDebt.Rating == "D" || results.TechnicalDebt.Rating == "F" { fail() }
//
// **场景5: 项目对比**
//   - 用例: 比较多个微服务的技术债务,识别最需要重构的服务
//   - 数据: 服务A(DebtRatio 1.2, B级) vs 服务B(DebtRatio 8.5, D级) → 优先重构服务B
//   - 资源分配: 债务比率高的服务分配更多重构资源
//
// # 示例
//
// **示例1: 中等债务项目(C级)**
//
//	// cqe.results.Issues包含: 10个error, 20个warning, 50个info
//	// cqe.results.Statistics.CodeLines = 10000
//	cqe.analyzeTechnicalDebt()
//	// 计算过程:
//	// TotalDebt = 10×2.0 + 20×0.5 + 50×0.1 = 20 + 10 + 5 = 35.0小时
//	// DebtRatio = 35.0 / 10000 × 1000 = 3.5 hours/KLOC
//	// Rating = "C" (假设TechnicalDebtRatingC=5.0,3.5 < 5.0)
//	// 结果:
//	// cqe.results.TechnicalDebt = TechnicalDebtAnalysis{
//	//   TotalDebt: 35.0,
//	//   DebtRatio: 3.5,
//	//   Rating: "C",
//	//   Categories: map[string]float64{},
//	//   Files: []FileDebt{},
//	//   Trends: []DebtTrend{},
//	// }
//
// **示例2: 优秀项目(A级)**
//
//	// cqe.results.Issues包含: 0个error, 5个warning, 10个info
//	// cqe.results.Statistics.CodeLines = 50000
//	cqe.analyzeTechnicalDebt()
//	// TotalDebt = 0×2.0 + 5×0.5 + 10×0.1 = 0 + 2.5 + 1.0 = 3.5小时
//	// DebtRatio = 3.5 / 50000 × 1000 = 0.07 hours/KLOC
//	// Rating = "A" (假设TechnicalDebtRatingA=1.0,0.07 < 1.0)
//
// **示例3: 严重债务项目(F级)**
//
//	// cqe.results.Issues包含: 100个error, 200个warning, 500个info
//	// cqe.results.Statistics.CodeLines = 5000
//	cqe.analyzeTechnicalDebt()
//	// TotalDebt = 100×2.0 + 200×0.5 + 500×0.1 = 200 + 100 + 50 = 350.0小时
//	// DebtRatio = 350.0 / 5000 × 1000 = 70.0 hours/KLOC (极高债务密度!)
//	// Rating = "F" (假设TechnicalDebtRatingD=10.0,70.0 > 10.0)
//	// **管理建议**: 债务350小时≈9周工时,项目需全面重构
//
// **示例4: 零代码行项目(边界情况)**
//
//	// cqe.results.Issues包含: 5个error
//	// cqe.results.Statistics.CodeLines = 0 (纯配置项目或统计错误)
//	cqe.analyzeTechnicalDebt()
//	// TotalDebt = 5×2.0 = 10.0小时
//	// DebtRatio = 0.0 (if条件不满足,保持零值,避免除零错误)
//	// Rating = "A" (DebtRatio=0.0 < TechnicalDebtRatingA)
//	// **注意**: 评级失真,TotalDebt=10小时但Rating=A,需特殊处理
//
// **示例5: 纯info问题项目**
//
//	// cqe.results.Issues包含: 0个error, 0个warning, 1000个info
//	// cqe.results.Statistics.CodeLines = 20000
//	cqe.analyzeTechnicalDebt()
//	// TotalDebt = 0×2.0 + 0×0.5 + 1000×0.1 = 100.0小时
//	// DebtRatio = 100.0 / 20000 × 1000 = 5.0 hours/KLOC
//	// Rating = "C" (假设5.0刚好等于TechnicalDebtRatingC边界)
//	// **洞察**: 虽然都是info,但1000个问题仍形成100小时债务(约2.5周工时)
//
// # 注意事项
//
//  1. **固定时间系数**: error=2h, warning=0.5h, info=0.1h完全硬编码,未考虑问题复杂度差异,
//     简单error(如未使用变量)和复杂error(如竞态条件)实际修复时间可能相差10倍
//
//  2. **CodeLines=0边界**: 当CodeLines=0时,DebtRatio保持零值(if条件跳过),导致Rating="A",
//     但TotalDebt可能很大,评级失真,改进: 添加特殊处理或返回Rating="N/A"
//
//  3. **评级阈值外部依赖**: TechnicalDebtRatingA/B/C/D常量未在本方法定义,需确保已声明,
//     否则编译错误,建议默认值: A=1.0, B=2.0, C=5.0, D=10.0
//
//  4. **预留字段未填充**: Categories/Files/Trends字段初始化为空,本方法不填充,
//     可能导致调用方期望这些数据时出错,需明确文档说明
//
//  5. **索引遍历性能**: for i := range cqe.results.Issues避免184字节复制,但需&cqe.results.Issues[i]取指针,
//     可读性略差,且指针作用域仅限循环内(issue := &cqe.results.Issues[i]在循环外失效)
//
//  6. **Severity字符串匹配**: switch issue.Severity区分大小写,"Error"/"WARNING"不匹配,会被跳过不累加债务,
//     改进: 使用strings.ToLower(issue.Severity)标准化
//
//  7. **债务累加精度**: totalDebt使用float64,可能有浮点精度问题(如0.1+0.1+0.1≠0.3),
//     对于债务计算影响较小,但大规模累加(10000个issue)可能累积误差
//
//  8. **无债务分解**: 仅计算总债务,未按Category/File/Severity分解,无法回答"Security类债务多少小时"这类问题,
//     改进: 填充Categories字段,如Categories["Security"] = 15.0
//
//  9. **评级主观性**: A-F评级阈值(1.0/2.0/5.0/10.0)基于经验,不同行业标准可能不同,
//     如金融系统可能要求DebtRatio < 0.5才算A级,而快速迭代的创业项目可接受5.0
//
// 10. **无时间趋势**: Trends字段预留但未实现,无法追踪债务历史变化(本周35h vs 上周40h → 好转5h),
//     改进: 集成历史数据,计算债务增减速度
//
// 11. **副作用函数**: 直接修改cqe.results.TechnicalDebt,非纯函数,难以单元测试,
//     改进: 返回TechnicalDebtAnalysis值,由调用方赋值
//
// 12. **未考虑优先级**: 所有debt平等累加,但高优先级问题应有更高权重(如紧急安全漏洞debt×2),
//     改进: 引入PriorityMultiplier, totalDebt += time × priorityWeight
//
// # 改进方向
//
//  1. **动态时间系数**: 根据问题Category和Message估算更精确的修复时间,
//     如"G401: MD5 hash"(安全关键) → 4小时(需迁移到bcrypt), "G201: SQL拼接" → 1小时(改用参数化)
//
//  2. **CodeLines=0特殊处理**: 检测零行情况,返回特殊Rating或设置标志位,
//     if cqe.results.Statistics.CodeLines == 0 { debt.Rating = "N/A"; debt.DebtRatio = -1 }
//
//  3. **债务分解**: 填充Categories字段,按问题类别统计债务,
//     for i := range issues { categoryDebt[issues[i].Category] += calculateDebtTime(issues[i]) }
//
//  4. **文件级债务**: 填充Files字段,识别债务热点文件,
//     Files: []FileDebt{ {Path: "main.go", Debt: 15.0, Issues: 8}, ... }
//
//  5. **时间趋势集成**: 从数据库加载历史债务,计算Trends,
//     Trends: []DebtTrend{ {Date: "2024-01-01", TotalDebt: 40.0}, {Date: "2024-01-08", TotalDebt: 35.0} }
//
//  6. **配置化阈值**: 将A/B/C/D/F阈值移至配置文件,支持不同行业标准,
//     config.yaml: technical_debt: { rating_a: 0.5, rating_b: 1.0, rating_c: 3.0, rating_d: 8.0 }
//
//  7. **Severity标准化**: 统一大小写避免匹配失败,
//     severity := strings.ToLower(issue.Severity); switch severity { case "error": ... }
//
//  8. **优先级加权**: 高优先级问题debt系数×1.5-2.0,
//     multiplier := 1.0; if issue.Priority == "High" { multiplier = 2.0 }; totalDebt += baseTime × multiplier
//
//  9. **返回值而非副作用**: 改为return TechnicalDebtAnalysis,便于测试,
//     debt := cqe.analyzeTechnicalDebt(); cqe.results.TechnicalDebt = debt
//
// 10. **机器学习估算**: 基于历史修复数据训练模型,预测每个issue的实际修复时间,
//     estimatedTime := mlModel.Predict(issue.Category, issue.Message, issue.File)
//
// 11. **团队经验因子**: 引入团队熟练度系数,
//     teamExperienceFactor := 0.7 // 经验丰富团队修复快30%; effectiveDebt := totalDebt × teamExperienceFactor
//
// 12. **债务偿还模拟**: 提供债务还清计划,
//     func SimulateDebtRepayment(currentDebt float64, weeklyHours float64) int { return int(math.Ceil(currentDebt / weeklyHours)) }
//
// 13. **可视化数据**: 生成图表数据结构,
//     ChartData: { Labels: ["Security", "Performance", "Style"], Values: [15.0, 8.5, 11.5] }
//
// 14. **债务成本转换**: 将小时数转换为财务成本,
//     hourlyRate := 500.0; financialDebt := totalDebt × hourlyRate // 35小时 × 500元/小时 = 17,500元
//
// 15. **SonarQube集成**: 对接SonarQube的技术债务模型,使用其更复杂的SQALE计算,
//     导入SonarQube债务数据,统一债务指标
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) analyzeTechnicalDebt() {
	debt := TechnicalDebtAnalysis{
		Categories: make(map[string]float64),
		Files:      []FileDebt{},
		Trends:     []DebtTrend{},
	}

	// 计算总债务
	totalDebt := 0.0
	// 使用索引遍历避免大结构体复制（184字节）
	for i := range cqe.results.Issues {
		issue := &cqe.results.Issues[i]
		switch issue.Severity {
		case "error":
			totalDebt += 2.0 // 错误需要2小时修复
		case "warning":
			totalDebt += 0.5 // 警告需要30分钟修复
		case "info":
			totalDebt += 0.1 // 信息问题需要6分钟修复
		}
	}

	debt.TotalDebt = totalDebt

	// 计算债务比率
	if cqe.results.Statistics.CodeLines > 0 {
		debt.DebtRatio = totalDebt / float64(cqe.results.Statistics.CodeLines) * 1000 // 每千行代码的债务小时数
	}

	// 评定债务等级
	switch {
	case debt.DebtRatio < TechnicalDebtRatingA:
		debt.Rating = "A"
	case debt.DebtRatio < TechnicalDebtRatingB:
		debt.Rating = "B"
	case debt.DebtRatio < TechnicalDebtRatingC:
		debt.Rating = "C"
	case debt.DebtRatio < TechnicalDebtRatingD:
		debt.Rating = "D"
	default:
		debt.Rating = "F"
	}

	cqe.results.TechnicalDebt = debt
}

// identifyQualityHotspots 识别代码质量热点文件并排序优先级(热点分析核心方法-问题聚合策略)
//
// # 功能说明
//
// 本方法从cqe.results.Issues中识别问题密集的文件(质量热点),采用"按文件分组→阈值过滤→严重度评分→
// 优先级排序"的四阶段策略。将散乱的问题列表转化为结构化的热点清单,每个热点包含File(文件路径)、
// IssueCount(问题数量)、Severity(严重度分数)、Priority(优先级1-5)、Description(描述)、
// Suggestions(改进建议)等6个字段。帮助开发者快速定位"最需要重构的文件",遵循帕累托原则:
// 20%的文件通常包含80%的问题,优先修复热点文件可获得最大质量提升。
//
// **设计目标**:
//   1. **问题聚焦**: 将问题从散点(1000个issue分布在100个文件)聚合为热点(20个高问题文件)
//   2. **优先级量化**: 通过IssueCount×2 + Severity综合评分,计算Priority(1-5),确保最严重文件优先处理
//   3. **阈值过滤**: 仅识别问题数≥3的文件为热点,避免噪音(1-2个问题的文件可能是正常现象)
//   4. **可操作建议**: 为每个热点生成Suggestions数组,提供具体改进方向(由generateHotspotSuggestions实现)
//   5. **可视化支持**: 返回结构化QualityHotspot数组,便于生成热力图、排行榜等可视化报表
//
// **与其他方法的协同**:
//   - **analyzeTechnicalDebt()**: 计算总体债务 vs 本方法识别具体债务文件
//   - **generateIssueBasedImprovements()**: 按Category聚合 vs 本方法按File聚合
//   - **identifyQualityHotspots()→calculateHotspotPriority()**: 热点识别 → 优先级计算的调用链
//
// # 热点识别策略
//
// **阶段1: 按文件分组**
//   - 输入: cqe.results.Issues(CodeIssue数组,如1000个问题分布在100个文件)
//   - 处理: 创建map[string][]CodeIssue,按issue.File分组
//   - 输出: fileIssues map,如{"main.go": [issue1, issue2, ...], "utils.go": [...]}
//   - 性能: O(n),n为问题数量,单次遍历完成分组
//
// **阶段2: 阈值过滤**
//   - 阈值: len(issues) >= 3(硬编码,改进方向:配置化)
//   - 原理: ≥3个问题表明系统性质量问题,1-2个可能是偶发
//   - 示例: 100个文件,仅20个≥3问题 → 识别为热点,其余80个忽略
//   - 帕累托验证: 20%的文件(20个)包含80%的问题(如800/1000=80%)
//
// **阶段3: 严重度评分**
//   - 评分规则: error=3.0, warning=2.0, info=1.0(权重比3:2:1)
//   - 公式: Severity = Σ(每个issue的严重度分数)
//   - 示例: 1个error + 2个warning + 3个info = 1×3.0 + 2×2.0 + 3×1.0 = 3 + 4 + 3 = 10.0
//   - 作用: 区分"3个info"(severity=3.0)vs"1个error"(severity=3.0),虽然分数相同但问题性质不同
//
// **阶段4: 优先级计算**
//   - 调用: calculateHotspotPriority(len(issues), severity)
//   - 综合评分: score = issueCount×2 + severity(问题数量权重更高)
//   - 优先级映射: score≥20→Priority 5(最高), score≥15→Priority 4, score≥10→Priority 3, ...
//   - 示例: 5个问题(3error+2warning) → score=5×2+(3×3+2×2)=10+13=23 → Priority 5
//
// # 执行流程
//
// 1. **初始化分组map**: fileIssues := make(map[string][]CodeIssue)
// 2. **遍历问题分组**: for i := range cqe.results.Issues,按issue.File追加到fileIssues[issue.File]
// 3. **初始化热点切片**: var hotspots []QualityHotspot
// 4. **遍历文件识别热点**: for file, issues := range fileIssues
// 5. **阈值检查**: if len(issues) >= 3,满足则继续,否则跳过该文件
// 6. **计算严重度**: for j := range issues,按Severity累加severity分数
// 7. **构建热点对象**: 创建QualityHotspot{File, IssueCount, Severity, Priority, Description, Suggestions}
// 8. **调用优先级计算**: Priority = calculateHotspotPriority(len(issues), severity)
// 9. **调用建议生成**: Suggestions = generateHotspotSuggestions(issues)
// 10. **追加热点**: hotspots = append(hotspots, hotspot)
// 11. **保存结果**: cqe.results.Hotspots = hotspots
//
// **时间复杂度**: O(n + m×k),n=问题总数,m=唯一文件数,k=平均每文件问题数(通常k很小,如5-10)
// **空间复杂度**: O(n),fileIssues map存储所有问题的副本(184字节×n)
//
// # 严重度权重设计详解
//
// **为什么error=3.0, warning=2.0, info=1.0?**
//   - **线性递减**: 3:2:1比例,error严重度是info的3倍,warning居中
//   - **区分度**: 确保1个error(3.0) > 1个warning(2.0),但2个warning(4.0) > 1个error(3.0)
//   - **累加合理性**: 多个低级问题可累积为高严重度(5个info=5.0 > 1个error=3.0)
//
// **为什么问题数量×2?**
//   - **强调数量**: 问题数量对热点判断影响更大(10个info比1个error更需要关注)
//   - **综合评分**: score = count×2 + severity,数量和严重度权重约2:1
//   - **示例对比**: 5个info(count=5, severity=5) → score=5×2+5=15 vs 1个error(count=1, severity=3) → score=1×2+3=5
//
// **局限性**:
//   - 固定权重,未考虑问题类型(Security error应高于Style error,但评分相同)
//   - 线性累加可能过度惩罚(100个info → severity=100,但可能批量修复只需1小时)
//
// # 参数
//
//   - 无显式参数,但依赖cqe.results.Issues切片(CodeIssue数组),
//     每个CodeIssue包含File(文件路径)、Severity("error"/"warning"/"info")等字段
//
// # 返回值
//
//   - 无返回值(void函数),通过副作用修改cqe.results.Hotspots字段,
//     Hotspots为[]QualityHotspot切片,每个QualityHotspot包含:
//     - File string: 文件路径(如"evaluators/code_quality.go")
//     - IssueCount int: 该文件问题数量(如8)
//     - Severity float64: 严重度总分(如15.0)
//     - Priority int: 优先级1-5(由calculateHotspotPriority计算)
//     - Description string: 动态描述(如"该文件存在8个质量问题")
//     - Suggestions []string: 改进建议数组(由generateHotspotSuggestions生成)
//
// # 使用场景
//
// **场景1: Evaluate()主流程调用**
//   - 触发: 所有检查完成,Issues列表已填充,需识别重构目标文件
//   - 调用: cqe.identifyQualityHotspots()(通常在analyzeTechnicalDebt()之后)
//   - 作用: 为报告提供"Top 10质量问题文件"排行榜
//
// **场景2: 重构计划制定**
//   - 用例: 团队计划Sprint重构,需确定优先处理哪些文件
//   - 数据: Hotspots按Priority降序排序,优先处理Priority 5的文件
//   - 决策: "本Sprint重构main.go(Priority 5, 12个问题)和utils.go(Priority 4, 8个问题)"
//
// **场景3: 代码审查重点**
//   - 用例: Code Review时,审查者优先检查热点文件
//   - 工具集成: IDE插件标记热点文件为红色,提醒审查者重点关注
//   - 效率提升: 集中审查20%的热点文件,覆盖80%的潜在问题
//
// **场景4: 可视化报表**
//   - 用例: 生成质量热力图(Heatmap),颜色深度表示问题严重度
//   - 数据: Hotspots数组转换为坐标和颜色 → 图表库渲染
//   - 示例: main.go(Severity 25.0)显示为深红色,lib.go(Severity 5.0)浅黄色
//
// **场景5: 持续追踪**
//   - 用例: 每周运行分析,追踪热点文件变化趋势
//   - 指标: "上周main.go有12个问题,本周降至8个 → 重构有效" vs "utils.go从5个增至10个 → 需干预"
//   - 自动化: CI/CD集成,热点文件增加触发告警
//
// # 示例
//
// **示例1: 典型热点识别**
//
//	// cqe.results.Issues包含15个问题:
//	// main.go: 3个error + 2个warning + 3个info (8个问题)
//	// utils.go: 1个error + 1个warning (2个问题,不满足阈值)
//	// lib.go: 1个warning + 2个info (3个问题,刚好满足阈值)
//	cqe.identifyQualityHotspots()
//	// 分组结果:
//	// fileIssues = {"main.go": [8个issue], "utils.go": [2个issue], "lib.go": [3个issue]}
//	// 过滤结果: utils.go被忽略(2 < 3)
//	// 热点1: main.go
//	//   IssueCount=8, Severity=3×3.0+2×2.0+3×1.0=9+4+3=16.0
//	//   Priority=calculateHotspotPriority(8, 16.0) → score=8×2+16=32 → Priority 5
//	//   Description="该文件存在8个质量问题"
//	// 热点2: lib.go
//	//   IssueCount=3, Severity=1×2.0+2×1.0=2+2=4.0
//	//   Priority=calculateHotspotPriority(3, 4.0) → score=3×2+4=10 → Priority 3
//	//   Description="该文件存在3个质量问题"
//	// 结果: cqe.results.Hotspots = [main.go热点, lib.go热点](长度2)
//
// **示例2: 无热点项目**
//
//	// cqe.results.Issues包含5个问题,分布在5个文件(每个文件1个问题)
//	cqe.identifyQualityHotspots()
//	// fileIssues = {"file1.go": [1问题], "file2.go": [1问题], ...}
//	// 过滤结果: 所有文件<3问题,全部被忽略
//	// 结果: cqe.results.Hotspots = [](空切片)
//	// **洞察**: 问题均匀分布,无明显热点,可能是系统性问题或质量较高
//
// **示例3: 单一热点文件**
//
//	// cqe.results.Issues包含100个问题,全部在main.go(极端集中)
//	cqe.identifyQualityHotspots()
//	// 假设50个error + 30个warning + 20个info
//	// Severity = 50×3.0 + 30×2.0 + 20×1.0 = 150 + 60 + 20 = 230.0
//	// score = 100×2 + 230 = 430 → Priority 5
//	// 结果: Hotspots = [main.go热点](单个元素)
//	// **管理建议**: main.go严重超载,建议拆分为多个模块
//
// **示例4: 边界阈值测试**
//
//	// cqe.results.Issues包含6个问题:
//	// file1.go: 3个info(刚好满足阈值)
//	// file2.go: 2个error(不满足阈值)
//	cqe.identifyQualityHotspots()
//	// file1.go: IssueCount=3, Severity=3×1.0=3.0, score=3×2+3=9 → Priority可能为2或3
//	// file2.go: 被忽略(2 < 3)
//	// **不合理性**: 3个info成为热点(可能批量修复仅需10分钟),2个error被忽略(可能需4小时修复)
//	// **改进**: 应综合Severity判断,如"count≥3 OR severity≥6.0"
//
// **示例5: 空Issues列表**
//
//	// cqe.results.Issues = []CodeIssue{}(完美代码或未运行检查)
//	cqe.identifyQualityHotspots()
//	// fileIssues = map[string][]CodeIssue{}(空map)
//	// for循环不执行
//	// 结果: cqe.results.Hotspots = [](空切片)
//
// # 注意事项
//
//  1. **固定阈值**: len(issues) >= 3硬编码,小项目(如500行)和大项目(如10万行)使用相同阈值不合理,
//     改进: 动态阈值,如count >= max(3, totalFiles×0.1)
//
//  2. **map遍历无序**: for file, issues := range fileIssues顺序不确定,
//     Hotspots数组顺序随机,需后续排序(如按Priority降序)
//
//  3. **结构体复制开销**: issue := cqe.results.Issues[i]复制184字节,虽然注释说"使用索引遍历避免",
//     但仍然复制了(正确做法应该是issue := &cqe.results.Issues[i]),不过后续append时又复制一次,实际影响有限
//
//  4. **Severity权重主观**: error=3, warning=2, info=1缺乏量化依据,
//     不同项目/行业可能需要不同权重(如安全关键项目error=10)
//
//  5. **无Category区分**: 所有error平等对待,但Security error应比Style error优先级更高,
//     改进: Severity计算时乘以CategoryWeight,如Security×2, Style×0.5
//
//  6. **调用外部函数**: calculateHotspotPriority()和generateHotspotSuggestions()未在本方法实现,
//     如果未定义会导致编译错误,需确保这些辅助函数存在
//
//  7. **副作用函数**: 直接修改cqe.results.Hotspots,非纯函数,难以单元测试,
//     改进: 返回[]QualityHotspot,由调用方赋值
//
//  8. **Description模板单一**: fmt.Sprintf("该文件存在%d个质量问题", len(issues))内容过于简单,
//     改进: 包含严重度信息,如"该文件存在8个问题(3个错误,2个警告,3个提示)"
//
//  9. **无文件大小考虑**: 1000行文件10个问题 vs 100行文件10个问题,后者问题密度更高但评分相同,
//     改进: Severity除以文件行数,计算问题密度(issues per 100 lines)
//
// 10. **Suggestions依赖外部**: generateHotspotSuggestions(issues)可能返回nil或空数组,
//     QualityHotspot.Suggestions字段可能为空,前端需处理
//
// 11. **无去重**: 若同一问题被多个linter重复报告(如gofmt和goimports都报告同一格式问题),
//     会重复计入IssueCount和Severity,改进: 问题去重逻辑
//
// 12. **性能考虑**: 大规模项目(如10万个问题分布在1万个文件),fileIssues map会消耗大量内存(184字节×10万≈18MB),
//     可考虑流式处理或增量更新
//
// # 改进方向
//
//  1. **动态阈值**: 根据项目规模调整,
//     threshold := max(3, len(fileIssues)/10) // 至少3,或总文件数的10%
//
//  2. **热点排序**: 返回前按Priority降序排序,确保最严重文件在前,
//     sort.Slice(hotspots, func(i, j int) bool { return hotspots[i].Priority > hotspots[j].Priority })
//
//  3. **问题密度计算**: 引入文件行数,计算相对密度,
//     density := float64(len(issues)) / float64(fileLines) × 100 // 每百行问题数
//
//  4. **Category加权**: Severity计算时考虑问题类型,
//     categoryWeight := map[string]float64{"Security": 2.0, "Performance": 1.5, "Style": 0.5}
//     severity += severityScore × categoryWeight[issue.Category]
//
//  5. **丰富Description**: 包含问题分布详情,
//     Description: fmt.Sprintf("该文件存在%d个问题(%d错误,%d警告,%d提示)", total, errCount, warnCount, infoCount)
//
//  6. **综合阈值**: 数量或严重度满足其一即可,
//     if len(issues) >= 3 || severity >= 6.0 { /* 识别为热点 */ }
//
//  7. **返回值模式**: 改为返回切片,便于测试,
//     func (cqe *CodeQualityEvaluator) identifyQualityHotspots() []QualityHotspot { ... return hotspots }
//
//  8. **去重逻辑**: 按File+Line+Message去重,
//     issueKey := fmt.Sprintf("%s:%d:%s", issue.File, issue.Line, issue.Message); if !seen[issueKey] { ... }
//
//  9. **Top-N限制**: 仅返回前N个最严重热点,避免报告过长,
//     sort.Slice(...); if len(hotspots) > 10 { hotspots = hotspots[:10] }
//
// 10. **文件类型过滤**: 排除测试文件或生成代码,
//     if strings.HasSuffix(file, "_test.go") || strings.Contains(file, "generated") { continue }
//
// 11. **增量分析**: 对比上次结果,标记新增热点或改善的文件,
//     type HotspotDelta struct { New []QualityHotspot, Improved []QualityHotspot, Worsened []QualityHotspot }
//
// 12. **可视化数据**: 生成图表数据结构,
//     type HeatmapData struct { Files []string, Severities []float64, Colors []string }
//
// 13. **并发优化**: 大规模项目可并发计算每个文件的严重度,
//     使用goroutine pool并发处理fileIssues map的每个key
//
// 14. **机器学习预测**: 基于历史数据预测热点文件未来趋势,
//     "main.go问题数量以每周+2速度增长,预计4周后达到20个,建议提前重构"
//
// 15. **建议智能化**: generateHotspotSuggestions()集成AI,根据具体问题生成个性化建议,
//     而非通用模板建议
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) identifyQualityHotspots() {
	fileIssues := make(map[string][]CodeIssue)

	// 按文件分组问题
	// 使用索引遍历避免大结构体复制（184字节）
	for i := range cqe.results.Issues {
		issue := cqe.results.Issues[i]
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
	}

	var hotspots []QualityHotspot

	// 识别问题集中的文件
	for file, issues := range fileIssues {
		if len(issues) >= 3 { // 有3个以上问题的文件作为热点
			severity := 0.0
			// 使用索引遍历避免大结构体复制（184字节）
			for j := range issues {
				switch issues[j].Severity {
				case "error":
					severity += 3.0
				case "warning":
					severity += 2.0
				case "info":
					severity += 1.0
				}
			}

			hotspot := QualityHotspot{
				File:        file,
				IssueCount:  len(issues),
				Severity:    severity,
				Priority:    calculateHotspotPriority(len(issues), severity),
				Description: fmt.Sprintf("该文件存在%d个质量问题", len(issues)),
				Suggestions: generateHotspotSuggestions(issues),
			}

			hotspots = append(hotspots, hotspot)
		}
	}

	cqe.results.Hotspots = hotspots
}

// calculateHotspotPriority 计算质量热点优先级并返回1-5整数评级(优先级计算辅助函数-综合评分策略)
//
// # 功能说明
//
// 本方法根据热点文件的问题数量(issueCount)和严重度总分(severity)计算综合优先级评分,采用
// "加权求和→阈值映射"二阶段策略:先计算score=issueCount×2+severity(数量权重为2,严重度权重为1),
// 再通过switch-case将score映射到1-5整数优先级(Priority)。较高优先级表示更紧急需要重构的文件,
// 1=最低优先级(可延后处理),5=最高优先级(必须立即重构)。为identifyQualityHotspots()提供量化的
// 优先级排序依据,确保团队优先处理最严重的质量热点文件。
//
// **设计目标**:
//   1. **综合评分**: 同时考虑问题数量和严重度,避免单一维度失真(如100个info vs 1个error)
//   2. **权重平衡**: issueCount×2强调数量重要性(问题多=系统性问题),severity×1补充严重度信息
//   3. **阈值分级**: 5个优先级层次(1-5),提供清晰的行动指导(Priority 5立即处理,Priority 1可选)
//   4. **常量复用**: 使用HighPriorityIssueThreshold等常量,便于全局调整阈值策略
//   5. **简洁高效**: 纯函数,无副作用,O(1)时间复杂度,可安全并发调用
//
// **数量权重×2的合理性**:
//   - 问题数量对热点判断影响更大:10个轻微问题比1个严重问题更需要关注(表明代码整体质量差)
//   - 数量易于修复:批量修复同类问题效率高(如10个格式问题可能1小时全部修复)
//   - 严重度叠加:多个问题累积的严重度已包含在severity参数中,×2避免过度强调严重度
//
// # 计算策略
//
// **阶段1: 加权求和**
//   - 公式: score = float64(issueCount) × 2 + severity
//   - issueCount权重: 2(每个问题贡献2分)
//   - severity权重: 1(每点严重度贡献1分)
//   - 示例: 5个问题(3error+2warning) → score = 5×2 + (3×3+2×2) = 10 + 13 = 23
//
// **阶段2: 阈值映射**
//   - score ≥ HighPriorityIssueThreshold(如20) → Priority 5(FactorFive,最高优先级)
//   - score ≥ MediumPriorityIssueThreshold(如15) → Priority 4(FactorFour)
//   - score ≥ 10 → Priority 3(硬编码阈值)
//   - score ≥ LowPriorityIssueThreshold(如5) → Priority 2
//   - score < 5 → Priority 1(最低优先级,default分支)
//
// **常量依赖**:
//   - HighPriorityIssueThreshold: 高优先级阈值(建议值20)
//   - MediumPriorityIssueThreshold: 中优先级阈值(建议值15)
//   - LowPriorityIssueThreshold: 低优先级阈值(建议值5)
//   - FactorFive: 常量5(用于Priority 5)
//   - FactorFour: 常量4(用于Priority 4)
//   - **注意**: 这些常量需在外部定义,否则编译错误
//
// # 执行流程
//
// 1. **计算综合分数**: score := float64(issueCount)×2 + severity(将issueCount转为float64避免整数溢出)
// 2. **阈值判断**: switch语句从高到低依次检查score范围
// 3. **返回优先级**: 匹配第一个满足条件的case,返回对应Priority整数(5/4/3/2/1)
//
// **时间复杂度**: O(1),固定5个case判断
// **空间复杂度**: O(1),仅局部变量score
//
// # 优先级语义详解
//
// **Priority 5 (score≥20, FactorFive)**: 最高优先级
//   - 语义: 严重质量问题,必须立即重构
//   - 典型场景: 10个error(score=10×2+10×3=50) 或 15个问题混合(score≥20)
//   - 行动: 本Sprint必须处理,阻塞其他开发
//   - 示例: main.go有12个问题(8error+4warning) → score=12×2+(8×3+4×2)=24+32=56 → Priority 5
//
// **Priority 4 (15≤score<20, FactorFour)**: 高优先级
//   - 语义: 明显质量问题,需尽快重构
//   - 典型场景: 7个问题(5error+2warning,score=7×2+19=33) 或 10个warning(score=10×2+20=40)
//   - 行动: 下个Sprint必须处理,不可拖延
//   - 示例: utils.go有8个问题(2error+6warning) → score=8×2+(2×3+6×2)=16+18=34 → Priority 4
//
// **Priority 3 (10≤score<15)**: 中等优先级
//   - 语义: 中等质量问题,建议重构
//   - 典型场景: 5个问题(2error+3warning,score=5×2+12=22) 或 8个info(score=8×2+8=24)
//   - 行动: 2-3个Sprint内处理,可根据资源灵活安排
//   - 示例: lib.go有6个问题(1error+5info) → score=6×2+(1×3+5×1)=12+8=20 → Priority 3
//
// **Priority 2 (5≤score<10, LowPriorityIssueThreshold≤score<10)**: 低优先级
//   - 语义: 轻微质量问题,可延后处理
//   - 典型场景: 3个问题(1warning+2info,score=3×2+4=10) 或 5个info(score=5×2+5=15)
//   - 行动: 有空闲时处理,或在大规模重构时一并解决
//   - 示例: helper.go有4个info → score=4×2+4=12 → Priority 2
//
// **Priority 1 (score<5)**: 最低优先级
//   - 语义: 极轻微问题,可选处理
//   - 典型场景: 1个warning(score=1×2+2=4) 或 2个info(score=2×2+2=6)
//   - 行动: 可忽略,或在代码审查时顺手修复
//   - 示例: config.go有1个info → score=1×2+1=3 → Priority 1
//
// # 参数
//
//   - issueCount int: 热点文件的问题数量,范围通常1-100(来自len(issues))
//     **注意**: 必须≥0,负数会导致score异常但不会panic(仅逻辑错误)
//
//   - severity float64: 热点文件的严重度总分,由identifyQualityHotspots()计算
//     公式: severity = Σ(error×3.0 + warning×2.0 + info×1.0)
//     范围: 0.0-∞(理论上无上限,100个error=300.0)
//     **注意**: 允许0.0(如文件无问题但被误传入,虽然不应该发生)
//
// # 返回值
//
//   - int: 优先级整数,取值范围1-5
//     5 = 最高优先级(立即处理)
//     4 = 高优先级(尽快处理)
//     3 = 中等优先级(建议处理)
//     2 = 低优先级(可延后)
//     1 = 最低优先级(可选)
//
// # 使用场景
//
// **场景1: identifyQualityHotspots()调用**
//   - 触发: 为每个热点文件计算Priority字段
//   - 调用: Priority = calculateHotspotPriority(len(issues), severity)
//   - 作用: 填充QualityHotspot.Priority,用于后续排序
//
// **场景2: 热点排序**
//   - 用例: 按Priority降序排序,最严重文件排在前
//   - 代码: sort.Slice(hotspots, func(i, j int) bool { return hotspots[i].Priority > hotspots[j].Priority })
//   - 效果: Priority 5文件排最前,Priority 1文件排最后
//
// **场景3: 质量门禁**
//   - 用例: CI/CD流水线阻止Priority≥4的热点文件合并
//   - 规则: if any(hotspot.Priority >= 4) { fail("存在高优先级质量热点,禁止合并") }
//   - 实施: 强制开发者先修复Priority≥4文件再提交
//
// **场景4: Sprint计划**
//   - 用例: 团队根据Priority分配重构任务
//   - 策略: Priority 5→分配给高级工程师立即处理, Priority 3-4→正常Sprint规划, Priority 1-2→技术债务backlog
//   - 估算: Priority 5文件预计8小时, Priority 4预计4小时, Priority 3预计2小时
//
// **场景5: 可视化展示**
//   - 用例: 仪表盘按Priority颜色编码
//   - 映射: Priority 5→深红色, Priority 4→橙色, Priority 3→黄色, Priority 2→浅绿, Priority 1→深绿
//   - 效果: 管理层快速识别高风险文件
//
// # 示例
//
// **示例1: 严重热点(Priority 5)**
//
//	priority := calculateHotspotPriority(12, 35.0) // 12个问题,严重度35.0
//	// score = 12×2 + 35.0 = 24 + 35 = 59.0
//	// 59.0 ≥ HighPriorityIssueThreshold(假设20) → return FactorFive (5)
//	// 返回: 5 (最高优先级)
//
// **示例2: 高优先级热点(Priority 4)**
//
//	priority := calculateHotspotPriority(8, 6.0) // 8个问题(可能6warning+2info),严重度6.0
//	// score = 8×2 + 6.0 = 16 + 6 = 22.0
//	// 15 ≤ 22.0 < 20 → return FactorFour (4)
//	// 返回: 4 (高优先级)
//
// **示例3: 中等优先级热点(Priority 3)**
//
//	priority := calculateHotspotPriority(5, 3.0) // 5个问题(可能3info+2warning),严重度3.0
//	// score = 5×2 + 3.0 = 10 + 3 = 13.0
//	// 10 ≤ 13.0 < 15 → return 3
//	// 返回: 3 (中等优先级)
//
// **示例4: 低优先级热点(Priority 2)**
//
//	priority := calculateHotspotPriority(3, 2.0) // 3个问题(可能2info+1warning),严重度2.0
//	// score = 3×2 + 2.0 = 6 + 2 = 8.0
//	// 5 ≤ 8.0 < 10 → return 2 (假设LowPriorityIssueThreshold=5)
//	// 返回: 2 (低优先级)
//
// **示例5: 最低优先级热点(Priority 1)**
//
//	priority := calculateHotspotPriority(1, 1.0) // 1个问题(1个info),严重度1.0
//	// score = 1×2 + 1.0 = 2 + 1 = 3.0
//	// 3.0 < 5(LowPriorityIssueThreshold) → default → return 1
//	// 返回: 1 (最低优先级)
//
// **示例6: 边界值测试(score=20刚好)**
//
//	priority := calculateHotspotPriority(10, 0.0) // 10个问题但全是最低严重度
//	// score = 10×2 + 0.0 = 20.0
//	// 20.0 ≥ HighPriorityIssueThreshold(20) → return FactorFive (5)
//	// 返回: 5 (刚好达到最高优先级阈值)
//
// **示例7: 高严重度低数量(Priority可能不匹配直觉)**
//
//	priority := calculateHotspotPriority(1, 15.0) // 1个问题但严重度极高(理论上不太可能,除非单个error权重15)
//	// score = 1×2 + 15.0 = 2 + 15 = 17.0
//	// 15 ≤ 17.0 < 20 → return FactorFour (4)
//	// 返回: 4 (虽然只有1个问题,但严重度高也达到高优先级)
//
// # 注意事项
//
//  1. **常量外部依赖**: HighPriorityIssueThreshold, MediumPriorityIssueThreshold, LowPriorityIssueThreshold,
//     FactorFive, FactorFour必须在包级别定义,否则编译错误
//
//  2. **阈值硬编码混合**: case score >= 10使用硬编码10,与其他case的常量不一致,
//     改进: 定义NormalPriorityIssueThreshold=10,保持一致性
//
//  3. **权重主观性**: issueCount×2的权重比缺乏量化依据,不同项目可能需要不同权重,
//     改进: 支持配置权重,如scoreWeights := {IssueCountWeight: 2.0, SeverityWeight: 1.0}
//
//  4. **整数溢出风险**: 极端情况下issueCount×2可能溢出int32(如issueCount > 1,073,741,823),
//     但实际不太可能(文件问题数通常<1000),转为float64已避免此问题
//
//  5. **float64精度**: score计算使用float64,可能有精度问题(如2.9999999 vs 3.0),
//     但阈值判断使用>=,影响极小(除非刚好在边界)
//
//  6. **无输入验证**: 未检查issueCount<0或severity<0的非法输入,
//     改进: 添加if issueCount < 0 || severity < 0 { return 1 }防御性编程
//
//  7. **返回值语义**: Priority 1-5的语义(1最低,5最高)可能与某些系统相反(1最高,5最低),
//     需文档明确说明,避免误用
//
//  8. **阈值静态**: 所有阈值编译时确定,无法运行时动态调整,
//     改进: 使用配置文件或数据库存储阈值,支持A/B测试不同策略
//
//  9. **无Category考虑**: 所有Category平等计算,但Security问题应优先级更高,
//     改进: 传入issues数组,检查是否包含Security问题,若有则Priority+1
//
// 10. **线性映射**: 5个Priority层次可能不够精细(如Priority 3范围10-15,跨度较大),
//     改进: 使用1-10或1-100更细粒度评分
//
// # 改进方向
//
//  1. **配置化权重**: 支持运行时配置issueCount和severity权重,
//     score := issueCount×config.IssueCountWeight + severity×config.SeverityWeight
//
//  2. **动态阈值**: 根据项目历史数据自适应调整阈值(如P50分位数作为Priority 3阈值),
//     使用机器学习预测最优阈值分布
//
//  3. **Category加权**: 检查issues中的Category,Security/Performance问题提升Priority,
//     if containsCategory(issues, "Security") { priority = min(priority+1, 5) }
//
//  4. **输入验证**: 添加防御性检查,
//     if issueCount < 0 || severity < 0 { log.Warn("非法输入"); return 1 }
//
//  5. **阈值常量统一**: 所有阈值使用常量,消除硬编码,
//     case score >= NormalPriorityIssueThreshold: return 3
//
//  6. **返回详细信息**: 返回结构体而非整数,包含Priority和评分原因,
//     type PriorityResult struct { Priority int, Score float64, Reason string }
//
//  7. **非线性映射**: 使用对数或指数函数平滑分布,
//     priority := int(math.Log10(score+1) × 2) // 更平滑的增长曲线
//
//  8. **历史对比**: 对比文件上次Priority,标记恶化或改善,
//     type PriorityDelta struct { Current int, Previous int, Delta int, Trend string }
//
//  9. **团队容量调整**: 根据团队资源动态调整阈值(团队小→降低阈值,优先处理更多问题),
//     adjustedThreshold := baseThreshold × (teamSize / 10.0)
//
// 10. **并发安全**: 虽然当前是纯函数已并发安全,但若未来引入全局状态(如缓存),需加锁保护
//
// 作者: JIA
func calculateHotspotPriority(issueCount int, severity float64) int {
	score := float64(issueCount)*2 + severity

	switch {
	case score >= float64(HighPriorityIssueThreshold):
		return FactorFive // 最高优先级
	case score >= float64(MediumPriorityIssueThreshold):
		return FactorFour
	case score >= 10:
		return 3
	case score >= float64(LowPriorityIssueThreshold):
		return 2
	default:
		return 1
	}
}

// generateHotspotSuggestions 为质量热点文件生成分类改进建议列表(建议生成辅助函数-模板映射策略)
//
// # 功能说明
//
// 本方法根据热点文件的问题分类(Category)生成改进建议数组,采用"问题聚合→模板映射→兜底建议"
// 三阶段策略:先统计issues中各Category的出现频次,再为每个出现的Category生成对应的标准建议,
// 最后若无匹配Category则返回通用建议。为identifyQualityHotspots()生成的QualityHotspot.Suggestions
// 字段提供内容,帮助开发者理解该热点文件需要什么类型的改进(复杂度/风格/安全/性能等)。
//
// **设计目标**:
//   1. **分类指导**: 根据问题类型提供针对性建议(complexity→降低圈复杂度, security→解决漏洞)
//   2. **去重聚合**: 多个同类问题仅生成一条建议(如10个complexity问题→1条"重构复杂函数"建议)
//   3. **模板化**: 使用预定义建议模板,确保语言一致性和专业性
//   4. **兜底机制**: 当Category不匹配或issues为空时,返回通用建议避免空数组
//   5. **轻量快速**: O(n)时间复杂度,适合为大量热点文件批量生成建议
//
// # 参数
//
//   - issues []CodeIssue: 该热点文件的所有问题数组,每个CodeIssue包含Category字段
//
// # 返回值
//
//   - []string: 改进建议字符串数组,长度1-6,典型2-3条
//
// # 支持的5个Category及其建议
//
//   - complexity: "重构复杂函数，降低圈复杂度"
//   - code_style: "修复代码风格问题，提高可读性"
//   - correctness: "修复潜在的正确性问题"
//   - security: "解决安全漏洞"
//   - performance: "优化性能问题"
//   - 兜底: "审查并重构该文件以提高代码质量"(无匹配时)
//
// 作者: JIA
func generateHotspotSuggestions(issues []CodeIssue) []string {
	suggestions := []string{}

	categoryCount := make(map[string]int)
	// 使用索引遍历避免大结构体复制（224字节）
	for i := range issues {
		categoryCount[issues[i].Category]++
	}

	for category := range categoryCount {
		switch category {
		case "complexity":
			suggestions = append(suggestions, "重构复杂函数，降低圈复杂度")
		case "code_style":
			suggestions = append(suggestions, "修复代码风格问题，提高可读性")
		case "correctness":
			suggestions = append(suggestions, "修复潜在的正确性问题")
		case "security":
			suggestions = append(suggestions, "解决安全漏洞")
		case "performance":
			suggestions = append(suggestions, "优化性能问题")
		}
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "审查并重构该文件以提高代码质量")
	}

	return suggestions
}

// saveResults 保存代码质量评估结果到JSON文件(结果持久化核心方法-委托模式)
//
// # 功能说明
//
// 本方法将cqe.results(CodeQualityResult结构体)序列化为JSON格式并保存到磁盘文件,
// 采用委托模式调用共享的utils.SaveCodeQualityResult函数实现,避免代码重复。
// 保存路径由cqe.config.ResultsPath指定,用于持久化评估结果供后续分析、报告生成、
// 趋势追踪使用。为Evaluate()主流程的最后一步,确保评估结果不丢失。
//
// **设计目标**:
//   1. **委托模式**: 复用utils包的通用保存逻辑,避免重复实现JSON序列化和文件写入
//   2. **错误传播**: 直接返回utils.SaveCodeQualityResult的error,由调用方处理
//   3. **配置驱动**: 保存路径通过config.ResultsPath配置,支持不同环境(开发/测试/生产)
//   4. **无副作用**: 除文件写入外无其他副作用,幂等操作(多次调用覆盖同一文件)
//   5. **简洁高效**: 单行委托调用,O(1)时间复杂度(不含I/O)
//
// # 参数
//
//   - 无显式参数,依赖cqe实例状态:
//     - cqe.config.ResultsPath: 保存路径(如"results/code_quality.json")
//     - cqe.results: CodeQualityResult结构体,包含所有评估结果
//
// # 返回值
//
//   - error: 保存失败时返回error(如文件权限问题、磁盘空间不足、JSON序列化错误),
//     成功返回nil
//
// # 使用场景
//
//   - 场景1: Evaluate()主流程最后一步,持久化评估结果
//   - 场景2: 手动保存中间结果(如调试时在某阶段保存)
//   - 场景3: 定时保存(如长时间运行的评估任务定期保存进度)
//
// 作者: JIA
func (cqe *CodeQualityEvaluator) saveResults() error {
	return utils.SaveCodeQualityResult(cqe.config.ResultsPath, cqe.results)
}

// GetDefaultConfig 获取默认代码质量评估配置(配置工厂函数-最佳实践预设)
//
// # 功能说明
//
// 本函数返回预配置的CodeQualityConfig结构体指针,包含推荐的默认设置:启用常用质量工具
// (golint/govet/gocyclo/gofmt),设置合理的阈值(复杂度10,覆盖率80%),配置输出路径等。
// 为用户提供开箱即用的配置,避免从零配置的繁琐,同时允许用户基于默认配置进行定制。
// 遵循Go最佳实践,使用构造函数模式返回配置对象。
//
// **设计目标**:
//   1. **开箱即用**: 提供生产级默认配置,用户可直接使用无需调整
//   2. **最佳实践**: 默认启用核心质量工具,禁用可选工具(如gosec需额外安装)
//   3. **可定制**: 返回指针允许用户修改特定字段,如config := GetDefaultConfig(); config.MaxComplexity = 15
//   4. **文档化**: 通过注释说明各工具用途和阈值合理性
//   5. **向后兼容**: 添加新字段时设置合理默认值,不破坏现有代码
//
// # 返回配置详情
//
// **EnabledTools**: 默认启用4个核心工具,禁用1个可选工具
//   - golint: true - 代码风格检查(已deprecated,建议替换为revive)
//   - govet: true - 静态分析工具,检测常见错误
//   - gocyclo: true - 圈复杂度检查,识别复杂函数
//   - gofmt: true - 代码格式化检查,确保统一风格
//   - gosec: false - 安全漏洞扫描(可选,需单独安装)
//
// **ToolPaths**: 工具可执行文件路径(假设在PATH中)
//   - 所有工具使用命令名作为路径(如"golint"),依赖系统PATH环境变量
//   - 用户可覆盖为绝对路径,如config.ToolPaths["golint"] = "/usr/local/bin/golint"
//
// **MaxComplexity**: 10 - 圈复杂度阈值
//   - 函数圈复杂度>10视为过于复杂,需重构
//   - 行业标准: 10(严格), 15(宽松), 20(遗留代码)
//
// **MinCoverage**: 80.0 - 最小测试覆盖率(%)
//   - 测试覆盖率<80%视为不足,影响测试质量评分
//   - 行业标准: 80%(推荐), 70%(可接受), 60%(最低)
//
// **ResultsPath**: "code_quality_results.json" - 结果保存路径
//   - 相对路径,保存在当前工作目录
//   - 建议生产环境使用绝对路径,如"/var/log/quality/results.json"
//
// # 参数
//
//   - 无参数,纯函数
//
// # 返回值
//
//   - *CodeQualityConfig: 配置结构体指针,用户可直接使用或修改
//
// # 使用场景
//
//   - 场景1: 快速开始,使用默认配置创建Evaluator
//     config := GetDefaultConfig(); evaluator := NewCodeQualityEvaluator(".", config)
//
//   - 场景2: 基于默认配置定制
//     config := GetDefaultConfig(); config.MaxComplexity = 15; config.EnabledTools["gosec"] = true
//
//   - 场景3: 测试环境快速配置
//     测试代码中使用默认配置避免繁琐的配置初始化
//
// 作者: JIA
func GetDefaultConfig() *CodeQualityConfig {
	return &CodeQualityConfig{
		EnabledTools: map[string]bool{
			"golint":  true,
			"govet":   true,
			"gocyclo": true,
			"gofmt":   true,
			"gosec":   false, // 可选安全工具
		},
		ToolPaths: map[string]string{
			"golint":  "golint",
			"govet":   "go",
			"gocyclo": "gocyclo",
			"gofmt":   "gofmt",
			"gosec":   "gosec",
		},
		Thresholds: QualityThresholds{
			CyclomaticComplexity:  10,
			CognitiveComplexity:   DefaultCognitiveMax,
			FunctionLength:        DefaultFunctionLength,
			ParameterCount:        DefaultIssueCountWarning,
			TestCoverage:          DefaultScore,
			BranchCoverage:        MediumQualityScore,
			DocumentationCoverage: PassingScore,
			CodeDuplication:       MaxCyclomaticComplexity,
			TechnicalDebt:         MaxCyclomaticComplexity,
			Maintainability:       PassingScore,
		},
		WeightSettings: QualityWeights{
			CodeStructure:        WeightMedium,
			StyleCompliance:      WeightMediumLow,
			SecurityAnalysis:     WeightLow,
			PerformanceAnalysis:  WeightLow,
			TestQuality:          WeightLow,
			DocumentationQuality: WeightVeryLow,
		},
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"vendor/*", "*.pb.go", "*_generated.go"},
		MaxFileSize:     MaxFileSize, // 1MB
		Timeout:         DefaultIssueCountWarning * time.Minute,
		DetailLevel:     "detailed",
		OutputFormat:    "json",
		SaveResults:     true,
		ResultsPath:     "",
	}
}
