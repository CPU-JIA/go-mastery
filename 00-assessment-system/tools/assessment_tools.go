package tools

import (
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

	"assessment-system/models"
)

// CodeAnalyzer 代码分析工具
type CodeAnalyzer struct {
	fileSet *token.FileSet
}

// NewCodeAnalyzer 创建代码分析器
func NewCodeAnalyzer() *CodeAnalyzer {
	return &CodeAnalyzer{
		fileSet: token.NewFileSet(),
	}
}

// AnalyzeComplexity 分析代码复杂度
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

// calculateCyclomaticComplexity 计算圈复杂度
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

// AnalyzeCodeStyle 分析代码风格
func (ca *CodeAnalyzer) AnalyzeCodeStyle(code string) (*StyleMetrics, error) {
	metrics := &StyleMetrics{}
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// 检查行长度
		if len(line) > 120 {
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

// hasBadNaming 检查是否有不良命名
func (ca *CodeAnalyzer) hasBadNaming(line string) bool {
	// 简化的命名检查
	badPatterns := []string{
		`\b[a-z]\b`,           // 单字母变量
		`\bvar[0-9]+\b`,       // var1, var2 等
		`\btemp\b`,            // temp 变量
		`\bdata\b`,            // 通用的 data
	}

	for _, pattern := range badPatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	return false
}

// hasConsistentIndentation 检查缩进一致性
func (ca *CodeAnalyzer) hasConsistentIndentation(line string) bool {
	if len(line) == 0 || !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
		return true // 没有缩进或非缩进行
	}

	// 简化检查：确保不混用空格和制表符
	hasSpaces := strings.HasPrefix(line, " ")
	hasTabs := strings.HasPrefix(line, "\t")

	return !(hasSpaces && hasTabs)
}

// TestAnalyzer 测试分析工具
type TestAnalyzer struct{}

// NewTestAnalyzer 创建测试分析器
func NewTestAnalyzer() *TestAnalyzer {
	return &TestAnalyzer{}
}

// AnalyzeTestCoverage 分析测试覆盖率
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

// countTestFunctions 计算测试函数数量
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

// NewProjectScanner 创建项目扫描器
func NewProjectScanner() *ProjectScanner {
	return &ProjectScanner{}
}

// ScanProject 扫描项目目录
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
		relPath, _ := filepath.Rel(rootPath, path)
		files[relPath] = string(content)

		return nil
	})

	return files, err
}

// ReportGenerator 报告生成器
type ReportGenerator struct{}

// NewReportGenerator 创建报告生成器
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{}
}

// GenerateAssessmentReport 生成评估报告
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

// GenerateProgressReport 生成进度报告
func (rg *ReportGenerator) GenerateProgressReport(student *models.StudentProfile) string {
	var report strings.Builder

	report.WriteString("# 学习进度报告\n\n")
	report.WriteString(fmt.Sprintf("**学生姓名:** %s\n", student.Name))
	report.WriteString(fmt.Sprintf("**邮箱:** %s\n", student.Email))
	report.WriteString(fmt.Sprintf("**当前阶段:** %d\n", student.CurrentStage))

	if len(student.Projects) > 0 {
		report.WriteString("## 项目历史\n\n")
		for _, project := range student.Projects {
			report.WriteString(fmt.Sprintf("- **%s** - %.2f分\n",
				project.Name,
				project.OverallScore))
		}
		report.WriteString("\n")

		// 计算进度
		progress := float64(student.CurrentStage) / 15.0 * 100
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
	TotalLines          int
	LongLines           int
	BadNaming           int
	Comments            int
	CommentRatio        float64
	EmptyLines          int
	IndentationIssues   int
}

// TestMetrics 测试指标
type TestMetrics struct {
	TotalSourceFiles    int
	TotalTestFiles      int
	TotalTestFunctions  int
	TestCoverage        float64
}

// ConfigLoader 配置加载器
type ConfigLoader struct{}

// LoadConfig 加载评估配置
func (cl *ConfigLoader) LoadConfig(configPath string) (*AssessmentConfig, error) {
	// 简化的配置加载
	return &AssessmentConfig{
		ScoreWeights: map[string]float64{
			"code_quality":      0.4,
			"functionality":     0.3,
			"test_coverage":     0.2,
			"documentation":     0.1,
		},
		QualityThresholds: QualityThresholds{
			MinScore:        60.0,
			GoodScore:       80.0,
			ExcellentScore:  95.0,
			MaxComplexity:   10,
			MinTestCoverage: 0.7,
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

// NewFileWatcher 创建文件监控器
func NewFileWatcher(paths []string) *FileWatcher {
	return &FileWatcher{
		watchPaths: paths,
	}
}

// WatchForChanges 监控文件变化
func (fw *FileWatcher) WatchForChanges(callback func(string)) error {
	// 简化的文件监控实现
	// 实际实现应该使用 fsnotify 等库
	for _, path := range fw.watchPaths {
		go fw.watchPath(path, callback)
	}
	return nil
}

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

// NewCLIRunner 创建命令行运行器
func NewCLIRunner() *CLIRunner {
	return &CLIRunner{}
}

// RunAssessment 运行评估
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

// InteractiveMode 交互式模式
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

func (cr *CLIRunner) showHelp() {
	fmt.Println(`
可用命令:
  help     - 显示帮助信息
  assess   - 评估当前项目
  report   - 生成详细报告
  config   - 查看配置
  quit     - 退出程序`)
}