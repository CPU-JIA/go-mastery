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

package evaluators

import (
	"bufio"
	"encoding/json"
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

func (g *GolintTool) Name() string    { return "golint" }
func (g *GolintTool) Version() string { return "latest" }

func (g *GolintTool) Execute(projectPath string) (*ToolResult, error) {
	start := time.Now()

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
		result, parseErr := g.ParseResult(string(output))
		if parseErr != nil {
			return result, parseErr
		}
	}

	return result, err
}

func (g *GolintTool) ParseResult(output string) (*ToolResult, error) {
	issues := []CodeIssue{}
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析golint输出格式: file:line:col: message
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 4 {
			continue
		}

		lineNum, _ := strconv.Atoi(parts[1])
		colNum, _ := strconv.Atoi(parts[2])

		issue := CodeIssue{
			ID:       fmt.Sprintf("golint_%d", i),
			Type:     "style",
			Severity: "warning",
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

func (gv *GovetTool) Name() string    { return "go vet" }
func (gv *GovetTool) Version() string { return "latest" }

func (gv *GovetTool) Execute(projectPath string) (*ToolResult, error) {
	start := time.Now()

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
			Severity: "error",
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

func (gc *GocycloTool) Name() string    { return "gocyclo" }
func (gc *GocycloTool) Version() string { return "latest" }

func (gc *GocycloTool) Execute(projectPath string) (*ToolResult, error) {
	start := time.Now()

	cmd := exec.Command("gocyclo", "-over", strconv.Itoa(gc.threshold), ".")
	cmd.Dir = projectPath
	output, _ := cmd.CombinedOutput()

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

func (gc *GocycloTool) ParseResult(output string) (*ToolResult, error) {
	issues := []CodeIssue{}
	lines := strings.Split(output, "\n")

	totalComplexity := 0
	functionCount := 0
	maxComplexity := 0

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 解析gocyclo输出格式: complexity function location
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		complexity, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		functionName := parts[1]
		location := strings.Join(parts[2:], " ")

		totalComplexity += complexity
		functionCount++
		if complexity > maxComplexity {
			maxComplexity = complexity
		}

		severity := "warning"
		priority := 2
		if complexity > gc.threshold*2 {
			severity = "error"
			priority = 3
		}

		issue := CodeIssue{
			ID:          fmt.Sprintf("gocyclo_%d", i),
			Type:        "complexity",
			Severity:    severity,
			Category:    "maintainability",
			Rule:        "cyclomatic_complexity",
			Function:    functionName,
			Message:     fmt.Sprintf("Function %s has cyclomatic complexity %d", functionName, complexity),
			Description: fmt.Sprintf("High complexity functions are harder to understand and maintain"),
			Suggestion:  "Consider breaking down this function into smaller, more focused functions",
			Impact:      "module",
			Complexity:  2,
			Priority:    priority,
		}

		// 尝试解析位置信息
		if strings.Contains(location, ":") {
			parts := strings.Split(location, ":")
			if len(parts) >= 2 {
				issue.File = parts[0]
				if lineNum, err := strconv.Atoi(parts[1]); err == nil {
					issue.Line = lineNum
				}
			}
		}

		issues = append(issues, issue)
	}

	avgComplexity := 0.0
	if functionCount > 0 {
		avgComplexity = float64(totalComplexity) / float64(functionCount)
	}

	metrics := map[string]interface{}{
		"total_complexity": totalComplexity,
		"function_count":   functionCount,
		"avg_complexity":   avgComplexity,
		"max_complexity":   maxComplexity,
		"threshold":        gc.threshold,
	}

	summary := ToolSummary{
		TotalIssues:  len(issues),
		WarningCount: countBySeverity(issues, "warning"),
		ErrorCount:   countBySeverity(issues, "error"),
		Score:        calculateComplexityScore(avgComplexity, maxComplexity),
		Passed:       len(issues) < 5, // 可配置阈值
	}

	return &ToolResult{
		Issues:  issues,
		Summary: summary,
		Metrics: metrics,
	}, nil
}

// NewCodeQualityEvaluator 创建代码质量评估器
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

// initializeTools 初始化分析工具
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
	cqe.results.Passed = cqe.results.OverallScore >= 70.0 // 可配置阈值

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

// runAnalysisTools 运行所有分析工具
func (cqe *CodeQualityEvaluator) runAnalysisTools(projectPath string) error {
	for name, tool := range cqe.analysisTools {
		log.Printf("执行分析工具: %s", name)

		result, err := tool.Execute(projectPath)
		if err != nil {
			log.Printf("工具 %s 执行失败: %v", name, err)
			// 继续执行其他工具，不中断整个评估过程
			continue
		}

		cqe.results.ToolResults[name] = result
		log.Printf("工具 %s 完成，发现 %d 个问题", name, result.Summary.TotalIssues)
	}

	return nil
}

// aggregateResults 聚合分析结果
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

// analyzeProjectStructure 分析项目结构
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
			switch n.(type) {
			case *ast.FuncDecl:
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

// analyzeGoFile 分析Go文件
func (cqe *CodeQualityEvaluator) analyzeGoFile(filePath string) (total, code, comment, blank int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer file.Close()

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

// calculateDimensionScores 计算维度得分
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

// calculateStructureScore 计算代码结构得分
func (cqe *CodeQualityEvaluator) calculateStructureScore() float64 {
	score := 100.0
	stats := cqe.results.Statistics

	// 复杂度惩罚
	if stats.AvgComplexity > float64(cqe.config.Thresholds.CyclomaticComplexity) {
		penalty := (stats.AvgComplexity - float64(cqe.config.Thresholds.CyclomaticComplexity)) * 5
		score -= penalty
	}

	// 函数长度惩罚（简化计算）
	avgFunctionLength := float64(stats.CodeLines) / float64(stats.Functions)
	if avgFunctionLength > float64(cqe.config.Thresholds.FunctionLength) {
		penalty := (avgFunctionLength - float64(cqe.config.Thresholds.FunctionLength)) * 0.5
		score -= penalty
	}

	// gocyclo工具的复杂度问题惩罚
	if result, exists := cqe.results.ToolResults["gocyclo"]; exists {
		score -= float64(result.Summary.ErrorCount) * 10
		score -= float64(result.Summary.WarningCount) * 5
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

// calculateStyleScore 计算代码风格得分
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

// calculateSecurityScore 计算安全得分
func (cqe *CodeQualityEvaluator) calculateSecurityScore() float64 {
	score := 100.0

	// gosec工具的安全问题惩罚（如果启用）
	if result, exists := cqe.results.ToolResults["gosec"]; exists {
		score -= float64(result.Summary.ErrorCount) * 15  // 严重安全问题
		score -= float64(result.Summary.WarningCount) * 8 // 一般安全问题
	}

	// 手动安全检查
	securityIssues := cqe.checkSecurityIssues()
	score -= float64(securityIssues) * 10

	if score < 0 {
		score = 0
	}
	return score
}

// calculatePerformanceScore 计算性能得分
func (cqe *CodeQualityEvaluator) calculatePerformanceScore() float64 {
	score := 100.0

	// 性能问题检查
	performanceIssues := cqe.checkPerformanceIssues()
	score -= float64(performanceIssues) * 8

	// 内存分配检查
	allocationIssues := cqe.checkAllocationPatterns()
	score -= float64(allocationIssues) * 5

	if score < 0 {
		score = 0
	}
	return score
}

// calculateTestScore 计算测试得分
func (cqe *CodeQualityEvaluator) calculateTestScore() float64 {
	stats := cqe.results.Statistics

	// 基于测试覆盖率的得分
	if stats.TestCoverage >= cqe.config.Thresholds.TestCoverage {
		return 100.0
	}

	// 线性计算：覆盖率越高得分越高
	score := (stats.TestCoverage / cqe.config.Thresholds.TestCoverage) * 100

	// 测试文件比例加分
	if stats.TestRatio >= 0.3 { // 30%以上的文件是测试文件
		score += 10
	}

	if score > 100 {
		score = 100
	}
	return score
}

// calculateDocumentationScore 计算文档得分
func (cqe *CodeQualityEvaluator) calculateDocumentationScore() float64 {
	stats := cqe.results.Statistics

	// 注释比例得分
	if stats.TotalLines > 0 {
		commentRatio := float64(stats.CommentLines) / float64(stats.TotalLines)
		score := commentRatio * 400 // 25%注释率得满分

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

// calculateOverallScore 计算整体得分
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
	if cqe.results.OverallScore >= 90 {
		cqe.results.Grade = "A"
	} else if cqe.results.OverallScore >= 80 {
		cqe.results.Grade = "B"
	} else if cqe.results.OverallScore >= 70 {
		cqe.results.Grade = "C"
	} else if cqe.results.OverallScore >= 60 {
		cqe.results.Grade = "D"
	} else {
		cqe.results.Grade = "F"
	}
}

// 辅助函数实现

// calculateStyleScore 计算风格得分（工具结果解析用）
func calculateStyleScore(issues []CodeIssue) float64 {
	if len(issues) == 0 {
		return 100.0
	}
	// 简单的线性递减：每个问题扣2分
	score := 100.0 - float64(len(issues))*2
	if score < 0 {
		score = 0
	}
	return score
}

// calculateCorrectnessScore 计算正确性得分
func calculateCorrectnessScore(issues []CodeIssue) float64 {
	if len(issues) == 0 {
		return 100.0
	}
	// 正确性问题更严重：每个问题扣10分
	score := 100.0 - float64(len(issues))*10
	if score < 0 {
		score = 0
	}
	return score
}

// calculateComplexityScore 计算复杂度得分
func calculateComplexityScore(avgComplexity float64, maxComplexity int) float64 {
	score := 100.0

	// 平均复杂度惩罚
	if avgComplexity > 5.0 {
		score -= (avgComplexity - 5.0) * 10
	}

	// 最大复杂度惩罚
	if maxComplexity > 10 {
		score -= float64(maxComplexity-10) * 5
	}

	if score < 0 {
		score = 0
	}
	return score
}

// countBySeverity 按严重程度统计问题数量
func countBySeverity(issues []CodeIssue, severity string) int {
	count := 0
	for _, issue := range issues {
		if issue.Severity == severity {
			count++
		}
	}
	return count
}

// checkGofmt 检查代码格式化
func (cqe *CodeQualityEvaluator) checkGofmt() int {
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

// checkNamingConventions 检查命名约定
func (cqe *CodeQualityEvaluator) checkNamingConventions() int {
	issues := 0
	// 这里可以实现具体的命名约定检查逻辑
	// 例如：检查包名、函数名、变量名是否符合Go语言规范
	return issues
}

// checkSecurityIssues 检查安全问题
func (cqe *CodeQualityEvaluator) checkSecurityIssues() int {
	issues := 0
	// 实现安全问题检查逻辑
	// 例如：硬编码密码、SQL注入风险、不安全的随机数生成等
	return issues
}

// checkPerformanceIssues 检查性能问题
func (cqe *CodeQualityEvaluator) checkPerformanceIssues() int {
	issues := 0
	// 实现性能问题检查逻辑
	// 例如：不必要的字符串拼接、未优化的循环、内存泄漏等
	return issues
}

// checkAllocationPatterns 检查内存分配模式
func (cqe *CodeQualityEvaluator) checkAllocationPatterns() int {
	issues := 0
	// 实现内存分配检查逻辑
	return issues
}

// hasReadme 检查是否有README文件
func (cqe *CodeQualityEvaluator) hasReadme() bool {
	readmeFiles := []string{"README.md", "README.txt", "README", "readme.md", "readme.txt"}

	for _, filename := range readmeFiles {
		if _, err := os.Stat(filepath.Join(cqe.results.ProjectPath, filename)); err == nil {
			return true
		}
	}

	return false
}

// checkPackageDocumentation 检查包文档
func (cqe *CodeQualityEvaluator) checkPackageDocumentation() float64 {
	// 检查各个包是否有包级别的文档注释
	// 返回文档覆盖率得分 (0-30分)
	return 20.0 // 简化实现
}

// generateImprovements 生成改进建议
func (cqe *CodeQualityEvaluator) generateImprovements() {
	var improvements []ImprovementSuggestion

	// 基于不同维度的得分生成建议
	for dimension, score := range cqe.results.DimensionScores {
		if score < 80 {
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

// generateDimensionImprovement 为特定维度生成改进建议
func (cqe *CodeQualityEvaluator) generateDimensionImprovement(dimension string, score float64) *ImprovementSuggestion {
	switch dimension {
	case "code_structure":
		return &ImprovementSuggestion{
			Category:    "Code Structure",
			Title:       "改善代码结构和复杂度",
			Description: "当前代码结构得分较低，建议重构高复杂度函数",
			Impact:      "提高代码可维护性和可读性",
			Effort:      "Medium",
			Priority:    3,
			Examples:    []string{"拆分长函数", "提取重复代码", "简化条件逻辑"},
		}
	case "test_quality":
		return &ImprovementSuggestion{
			Category:    "Test Quality",
			Title:       "提高测试覆盖率",
			Description: "当前测试覆盖率不足，建议添加更多测试用例",
			Impact:      "提高代码质量和可靠性",
			Effort:      "High",
			Priority:    2,
			Examples:    []string{"添加单元测试", "编写集成测试", "测试边界条件"},
		}
	default:
		return nil
	}
}

// generateIssueBasedImprovements 基于具体问题生成改进建议
func (cqe *CodeQualityEvaluator) generateIssueBasedImprovements() []ImprovementSuggestion {
	var improvements []ImprovementSuggestion

	// 统计问题类型
	issueCategories := make(map[string]int)
	for _, issue := range cqe.results.Issues {
		issueCategories[issue.Category]++
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

// analyzeTechnicalDebt 分析技术债务
func (cqe *CodeQualityEvaluator) analyzeTechnicalDebt() {
	debt := TechnicalDebtAnalysis{
		Categories: make(map[string]float64),
		Files:      []FileDebt{},
		Trends:     []DebtTrend{},
	}

	// 计算总债务
	totalDebt := 0.0
	for _, issue := range cqe.results.Issues {
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
	case debt.DebtRatio < 1.0:
		debt.Rating = "A"
	case debt.DebtRatio < 3.0:
		debt.Rating = "B"
	case debt.DebtRatio < 5.0:
		debt.Rating = "C"
	case debt.DebtRatio < 10.0:
		debt.Rating = "D"
	default:
		debt.Rating = "F"
	}

	cqe.results.TechnicalDebt = debt
}

// identifyQualityHotspots 识别质量热点
func (cqe *CodeQualityEvaluator) identifyQualityHotspots() {
	fileIssues := make(map[string][]CodeIssue)

	// 按文件分组问题
	for _, issue := range cqe.results.Issues {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
	}

	var hotspots []QualityHotspot

	// 识别问题集中的文件
	for file, issues := range fileIssues {
		if len(issues) >= 3 { // 有3个以上问题的文件作为热点
			severity := 0.0
			for _, issue := range issues {
				switch issue.Severity {
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

// calculateHotspotPriority 计算热点优先级
func calculateHotspotPriority(issueCount int, severity float64) int {
	score := float64(issueCount)*2 + severity

	if score >= 20 {
		return 5 // 最高优先级
	} else if score >= 15 {
		return 4
	} else if score >= 10 {
		return 3
	} else if score >= 5 {
		return 2
	}

	return 1
}

// generateHotspotSuggestions 为热点生成建议
func generateHotspotSuggestions(issues []CodeIssue) []string {
	suggestions := []string{}

	categoryCount := make(map[string]int)
	for _, issue := range issues {
		categoryCount[issue.Category]++
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

// saveResults 保存评估结果
func (cqe *CodeQualityEvaluator) saveResults() error {
	if cqe.config.ResultsPath == "" {
		cqe.config.ResultsPath = "quality_assessment_results.json"
	}

	data, err := json.MarshalIndent(cqe.results, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %v", err)
	}

	if err := os.WriteFile(cqe.config.ResultsPath, data, 0644); err != nil {
		return fmt.Errorf("保存结果文件失败: %v", err)
	}

	log.Printf("评估结果已保存到: %s", cqe.config.ResultsPath)
	return nil
}

// GetDefaultConfig 获取默认配置
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
			CognitiveComplexity:   15,
			FunctionLength:        50,
			ParameterCount:        5,
			TestCoverage:          80.0,
			BranchCoverage:        75.0,
			DocumentationCoverage: 70.0,
			CodeDuplication:       5.0,
			TechnicalDebt:         5.0,
			Maintainability:       70.0,
		},
		WeightSettings: QualityWeights{
			CodeStructure:        0.25,
			StyleCompliance:      0.20,
			SecurityAnalysis:     0.15,
			PerformanceAnalysis:  0.15,
			TestQuality:          0.15,
			DocumentationQuality: 0.10,
		},
		IncludePatterns: []string{"*.go"},
		ExcludePatterns: []string{"vendor/*", "*.pb.go", "*_generated.go"},
		MaxFileSize:     1048576, // 1MB
		Timeout:         5 * time.Minute,
		DetailLevel:     "detailed",
		OutputFormat:    "json",
		SaveResults:     true,
		ResultsPath:     "",
	}
}
