// Package main 演示 Go 标准库贡献指南
// 本模块涵盖向 Go 标准库贡献代码的完整流程：
// - 贡献准备工作
// - 代码规范与风格
// - 测试要求
// - 代码审查流程
package main

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// 贡献类型定义
// ============================================================================

// ContributionType 贡献类型
type ContributionType int

const (
	ContributionTypeBugFix ContributionType = iota
	ContributionTypeFeature
	ContributionTypeDocumentation
	ContributionTypeTest
	ContributionTypePerformance
	ContributionTypeRefactor
)

func (t ContributionType) String() string {
	switch t {
	case ContributionTypeBugFix:
		return "Bug Fix"
	case ContributionTypeFeature:
		return "New Feature"
	case ContributionTypeDocumentation:
		return "Documentation"
	case ContributionTypeTest:
		return "Test"
	case ContributionTypePerformance:
		return "Performance"
	case ContributionTypeRefactor:
		return "Refactor"
	default:
		return "Unknown"
	}
}

// CLStatus CL 状态
type CLStatus int

const (
	CLStatusDraft CLStatus = iota
	CLStatusPending
	CLStatusReviewing
	CLStatusApproved
	CLStatusMerged
	CLStatusAbandoned
)

func (s CLStatus) String() string {
	switch s {
	case CLStatusDraft:
		return "Draft"
	case CLStatusPending:
		return "Pending Review"
	case CLStatusReviewing:
		return "Under Review"
	case CLStatusApproved:
		return "Approved"
	case CLStatusMerged:
		return "Merged"
	case CLStatusAbandoned:
		return "Abandoned"
	default:
		return "Unknown"
	}
}

// ============================================================================
// 贡献核心结构
// ============================================================================

// ChangeList 变更列表 (CL)
type ChangeList struct {
	ID           string
	Title        string
	Description  string
	Author       string
	Type         ContributionType
	Status       CLStatus
	Package      string
	Files        []FileChange
	Tests        []TestResult
	Reviewers    []string
	Comments     []ReviewComment
	TryBotStatus *TryBotStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FileChange 文件变更
type FileChange struct {
	Path      string
	Action    string // add, modify, delete
	Additions int
	Deletions int
}

// TestResult 测试结果
type TestResult struct {
	Package  string
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
}

// ReviewComment 审查评论
type ReviewComment struct {
	Author    string
	File      string
	Line      int
	Content   string
	Resolved  bool
	Timestamp time.Time
}

// TryBotStatus TryBot 状态
type TryBotStatus struct {
	Linux386   string
	LinuxAMD64 string
	LinuxARM64 string
	Darwin     string
	Windows    string
	FreeBSD    string
	AllPassed  bool
	LastRun    time.Time
}

// ============================================================================
// 贡献指南
// ============================================================================

// ContributionGuide 贡献指南
type ContributionGuide struct {
	prerequisites []string
	codeStyle     *CodeStyleGuide
	testingReqs   *TestingRequirements
	reviewProcess *ReviewProcess
}

// CodeStyleGuide 代码风格指南
type CodeStyleGuide struct {
	rules []StyleRule
}

// StyleRule 风格规则
type StyleRule struct {
	Name        string
	Description string
	Example     string
	BadExample  string
}

// TestingRequirements 测试要求
type TestingRequirements struct {
	requirements []TestRequirement
}

// TestRequirement 测试要求项
type TestRequirement struct {
	Name        string
	Description string
	Required    bool
}

// ReviewProcess 审查流程
type ReviewProcess struct {
	steps []ReviewStep
}

// ReviewStep 审查步骤
type ReviewStep struct {
	Order       int
	Name        string
	Description string
	Responsible string
}

// NewContributionGuide 创建贡献指南
func NewContributionGuide() *ContributionGuide {
	guide := &ContributionGuide{}
	guide.initPrerequisites()
	guide.initCodeStyle()
	guide.initTestingReqs()
	guide.initReviewProcess()
	return guide
}

func (g *ContributionGuide) initPrerequisites() {
	g.prerequisites = []string{
		"签署 Google CLA (Contributor License Agreement)",
		"安装 Go 开发环境 (建议使用最新稳定版)",
		"配置 Git 和 Gerrit 账户",
		"安装 git-codereview 工具: go install golang.org/x/review/git-codereview@latest",
		"克隆 Go 源码仓库: git clone https://go.googlesource.com/go",
		"阅读贡献指南: https://go.dev/doc/contribute",
	}
}

func (g *ContributionGuide) initCodeStyle() {
	g.codeStyle = &CodeStyleGuide{
		rules: []StyleRule{
			{
				Name:        "使用 gofmt",
				Description: "所有代码必须通过 gofmt 格式化",
				Example:     "func foo() {\n\treturn\n}",
				BadExample:  "func foo(){\nreturn\n}",
			},
			{
				Name:        "注释规范",
				Description: "导出的标识符必须有文档注释，以标识符名称开头",
				Example:     "// Reader reads data from the underlying source.\ntype Reader struct {}",
				BadExample:  "// This is a reader\ntype Reader struct {}",
			},
			{
				Name:        "错误处理",
				Description: "错误应该被处理或显式忽略，使用 _ 忽略时需要注释说明",
				Example:     "if err != nil {\n\treturn err\n}",
				BadExample:  "foo() // 忽略返回的错误",
			},
			{
				Name:        "命名规范",
				Description: "使用 MixedCaps 或 mixedCaps，避免下划线",
				Example:     "func readBuffer() {}",
				BadExample:  "func read_buffer() {}",
			},
			{
				Name:        "包注释",
				Description: "每个包应该有包注释，位于 doc.go 或主文件顶部",
				Example:     "// Package strings implements simple functions to manipulate UTF-8 encoded strings.",
				BadExample:  "// strings package",
			},
		},
	}
}

func (g *ContributionGuide) initTestingReqs() {
	g.testingReqs = &TestingRequirements{
		requirements: []TestRequirement{
			{
				Name:        "单元测试",
				Description: "所有新功能必须有对应的单元测试",
				Required:    true,
			},
			{
				Name:        "测试覆盖率",
				Description: "新代码应该有合理的测试覆盖率",
				Required:    true,
			},
			{
				Name:        "基准测试",
				Description: "性能敏感的代码应该有基准测试",
				Required:    false,
			},
			{
				Name:        "示例测试",
				Description: "公开 API 应该有示例测试用于文档",
				Required:    false,
			},
			{
				Name:        "回归测试",
				Description: "Bug 修复应该包含防止回归的测试",
				Required:    true,
			},
			{
				Name:        "跨平台测试",
				Description: "代码应该在所有支持的平台上通过测试",
				Required:    true,
			},
		},
	}
}

func (g *ContributionGuide) initReviewProcess() {
	g.reviewProcess = &ReviewProcess{
		steps: []ReviewStep{
			{
				Order:       1,
				Name:        "创建 Issue",
				Description: "在 GitHub 上创建 issue 描述问题或功能",
				Responsible: "贡献者",
			},
			{
				Order:       2,
				Name:        "讨论方案",
				Description: "与维护者讨论实现方案，获得初步认可",
				Responsible: "贡献者 + 维护者",
			},
			{
				Order:       3,
				Name:        "提交 CL",
				Description: "使用 git codereview 提交变更到 Gerrit",
				Responsible: "贡献者",
			},
			{
				Order:       4,
				Name:        "TryBot 测试",
				Description: "自动化测试在多个平台上运行",
				Responsible: "自动化系统",
			},
			{
				Order:       5,
				Name:        "代码审查",
				Description: "维护者审查代码，提出修改建议",
				Responsible: "维护者",
			},
			{
				Order:       6,
				Name:        "修改迭代",
				Description: "根据反馈修改代码，更新 CL",
				Responsible: "贡献者",
			},
			{
				Order:       7,
				Name:        "最终批准",
				Description: "获得 LGTM (Looks Good To Me) 批准",
				Responsible: "维护者",
			},
			{
				Order:       8,
				Name:        "合并",
				Description: "CL 被合并到主分支",
				Responsible: "维护者",
			},
		},
	}
}

// PrintPrerequisites 打印前置条件
func (g *ContributionGuide) PrintPrerequisites() {
	fmt.Println("\n贡献前置条件:")
	for i, prereq := range g.prerequisites {
		fmt.Printf("  %d. %s\n", i+1, prereq)
	}
}

// PrintCodeStyle 打印代码风格
func (g *ContributionGuide) PrintCodeStyle() {
	fmt.Println("\n代码风格规范:")
	for _, rule := range g.codeStyle.rules {
		fmt.Printf("\n  [%s]\n", rule.Name)
		fmt.Printf("  说明: %s\n", rule.Description)
		fmt.Printf("  正确示例:\n    %s\n", strings.ReplaceAll(rule.Example, "\n", "\n    "))
	}
}

// PrintTestingReqs 打印测试要求
func (g *ContributionGuide) PrintTestingReqs() {
	fmt.Println("\n测试要求:")
	for _, req := range g.testingReqs.requirements {
		required := ""
		if req.Required {
			required = "[必需]"
		} else {
			required = "[推荐]"
		}
		fmt.Printf("  %s %s\n", required, req.Name)
		fmt.Printf("      %s\n", req.Description)
	}
}

// PrintReviewProcess 打印审查流程
func (g *ContributionGuide) PrintReviewProcess() {
	fmt.Println("\n代码审查流程:")
	for _, step := range g.reviewProcess.steps {
		fmt.Printf("  %d. %s\n", step.Order, step.Name)
		fmt.Printf("     %s\n", step.Description)
		fmt.Printf("     负责人: %s\n", step.Responsible)
	}
}

// ============================================================================
// CL 模拟器
// ============================================================================

// CLSimulator CL 模拟器
type CLSimulator struct {
	cls map[string]*ChangeList
}

// NewCLSimulator 创建 CL 模拟器
func NewCLSimulator() *CLSimulator {
	return &CLSimulator{
		cls: make(map[string]*ChangeList),
	}
}

// CreateCL 创建 CL
func (s *CLSimulator) CreateCL(title, author, pkg string, clType ContributionType) *ChangeList {
	id := fmt.Sprintf("CL/%d", len(s.cls)+100001)

	cl := &ChangeList{
		ID:        id,
		Title:     title,
		Author:    author,
		Type:      clType,
		Status:    CLStatusDraft,
		Package:   pkg,
		Files:     make([]FileChange, 0),
		Tests:     make([]TestResult, 0),
		Comments:  make([]ReviewComment, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.cls[id] = cl
	return cl
}

// AddFileChange 添加文件变更
func (s *CLSimulator) AddFileChange(clID string, change FileChange) {
	if cl, exists := s.cls[clID]; exists {
		cl.Files = append(cl.Files, change)
		cl.UpdatedAt = time.Now()
	}
}

// SubmitForReview 提交审查
func (s *CLSimulator) SubmitForReview(clID string) error {
	cl, exists := s.cls[clID]
	if !exists {
		return fmt.Errorf("CL 不存在: %s", clID)
	}

	cl.Status = CLStatusPending
	cl.UpdatedAt = time.Now()

	fmt.Printf("  [Gerrit] CL %s 已提交审查\n", clID)
	return nil
}

// RunTryBot 运行 TryBot
func (s *CLSimulator) RunTryBot(clID string) error {
	cl, exists := s.cls[clID]
	if !exists {
		return fmt.Errorf("CL 不存在: %s", clID)
	}

	fmt.Printf("  [TryBot] 开始测试 CL %s...\n", clID)

	// 模拟 TryBot 测试
	cl.TryBotStatus = &TryBotStatus{
		Linux386:   "PASS",
		LinuxAMD64: "PASS",
		LinuxARM64: "PASS",
		Darwin:     "PASS",
		Windows:    "PASS",
		FreeBSD:    "PASS",
		AllPassed:  true,
		LastRun:    time.Now(),
	}

	fmt.Printf("  [TryBot] linux-386: %s\n", cl.TryBotStatus.Linux386)
	fmt.Printf("  [TryBot] linux-amd64: %s\n", cl.TryBotStatus.LinuxAMD64)
	fmt.Printf("  [TryBot] linux-arm64: %s\n", cl.TryBotStatus.LinuxARM64)
	fmt.Printf("  [TryBot] darwin-amd64: %s\n", cl.TryBotStatus.Darwin)
	fmt.Printf("  [TryBot] windows-amd64: %s\n", cl.TryBotStatus.Windows)
	fmt.Printf("  [TryBot] freebsd-amd64: %s\n", cl.TryBotStatus.FreeBSD)
	fmt.Printf("  [TryBot] 所有测试通过: %v\n", cl.TryBotStatus.AllPassed)

	return nil
}

// AddReviewComment 添加审查评论
func (s *CLSimulator) AddReviewComment(clID string, comment ReviewComment) {
	if cl, exists := s.cls[clID]; exists {
		comment.Timestamp = time.Now()
		cl.Comments = append(cl.Comments, comment)
		cl.Status = CLStatusReviewing
		cl.UpdatedAt = time.Now()

		fmt.Printf("  [Review] %s 在 %s:%d 评论:\n", comment.Author, comment.File, comment.Line)
		fmt.Printf("           %s\n", comment.Content)
	}
}

// ApproveCL 批准 CL
func (s *CLSimulator) ApproveCL(clID, reviewer string) error {
	cl, exists := s.cls[clID]
	if !exists {
		return fmt.Errorf("CL 不存在: %s", clID)
	}

	cl.Status = CLStatusApproved
	cl.Reviewers = append(cl.Reviewers, reviewer)
	cl.UpdatedAt = time.Now()

	fmt.Printf("  [Review] %s: LGTM (Looks Good To Me)\n", reviewer)
	fmt.Printf("  [Gerrit] CL %s 已批准\n", clID)

	return nil
}

// MergeCL 合并 CL
func (s *CLSimulator) MergeCL(clID string) error {
	cl, exists := s.cls[clID]
	if !exists {
		return fmt.Errorf("CL 不存在: %s", clID)
	}

	if cl.Status != CLStatusApproved {
		return fmt.Errorf("CL 必须先获得批准")
	}

	cl.Status = CLStatusMerged
	cl.UpdatedAt = time.Now()

	fmt.Printf("  [Gerrit] CL %s 已合并到主分支\n", clID)

	return nil
}

// PrintCLSummary 打印 CL 摘要
func (s *CLSimulator) PrintCLSummary(clID string) {
	cl, exists := s.cls[clID]
	if !exists {
		fmt.Printf("CL 不存在: %s\n", clID)
		return
	}

	fmt.Printf("\n  CL 摘要:\n")
	fmt.Printf("    ID: %s\n", cl.ID)
	fmt.Printf("    标题: %s\n", cl.Title)
	fmt.Printf("    作者: %s\n", cl.Author)
	fmt.Printf("    类型: %s\n", cl.Type)
	fmt.Printf("    包: %s\n", cl.Package)
	fmt.Printf("    状态: %s\n", cl.Status)
	fmt.Printf("    文件变更: %d\n", len(cl.Files))
	fmt.Printf("    评论数: %d\n", len(cl.Comments))
	fmt.Printf("    审查者: %v\n", cl.Reviewers)
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateContributionGuide() {
	fmt.Println("\n=== Go 标准库贡献指南 ===")

	guide := NewContributionGuide()

	guide.PrintPrerequisites()
	guide.PrintCodeStyle()
	guide.PrintTestingReqs()
	guide.PrintReviewProcess()
}

func demonstrateCLWorkflow() {
	fmt.Println("\n=== CL 工作流程模拟 ===")

	simulator := NewCLSimulator()

	// 1. 创建 CL
	fmt.Println("\n步骤 1: 创建 CL")
	cl := simulator.CreateCL(
		"strings: add Clone function",
		"gopher@example.com",
		"strings",
		ContributionTypeFeature,
	)
	fmt.Printf("  创建 CL: %s\n", cl.ID)
	fmt.Printf("  标题: %s\n", cl.Title)

	// 2. 添加文件变更
	fmt.Println("\n步骤 2: 添加文件变更")
	simulator.AddFileChange(cl.ID, FileChange{
		Path:      "src/strings/clone.go",
		Action:    "add",
		Additions: 25,
		Deletions: 0,
	})
	simulator.AddFileChange(cl.ID, FileChange{
		Path:      "src/strings/clone_test.go",
		Action:    "add",
		Additions: 50,
		Deletions: 0,
	})
	fmt.Printf("  添加文件: strings/clone.go (+25)\n")
	fmt.Printf("  添加文件: strings/clone_test.go (+50)\n")

	// 3. 提交审查
	fmt.Println("\n步骤 3: 提交审查")
	simulator.SubmitForReview(cl.ID)

	// 4. 运行 TryBot
	fmt.Println("\n步骤 4: 运行 TryBot")
	simulator.RunTryBot(cl.ID)

	// 5. 代码审查
	fmt.Println("\n步骤 5: 代码审查")
	simulator.AddReviewComment(cl.ID, ReviewComment{
		Author:  "reviewer@golang.org",
		File:    "src/strings/clone.go",
		Line:    15,
		Content: "Consider using copy() instead of manual loop for better performance",
	})

	// 6. 批准
	fmt.Println("\n步骤 6: 批准 CL")
	simulator.ApproveCL(cl.ID, "reviewer@golang.org")

	// 7. 合并
	fmt.Println("\n步骤 7: 合并 CL")
	simulator.MergeCL(cl.ID)

	// 8. 显示摘要
	simulator.PrintCLSummary(cl.ID)
}

func demonstrateCommitMessage() {
	fmt.Println("\n=== 提交信息规范 ===")
	fmt.Println(`Go 标准库的提交信息格式:

  <package>: <short description>

  <longer description if needed>

  Fixes #<issue number>

示例:

  strings: add Clone function

  Clone returns a fresh copy of s. It guarantees to make a copy of s
  into a new allocation, which can be important when retaining only
  a small substring of a much larger string.

  Fixes #45038

规则:
1. 第一行: 包名 + 冒号 + 简短描述 (不超过 72 字符)
2. 空行
3. 详细描述 (可选，每行不超过 72 字符)
4. 空行
5. 关联的 issue (Fixes #xxx 或 Updates #xxx)

常用前缀:
- <package>: 修改特定包
- all: 影响多个包的变更
- cmd/<tool>: 修改命令行工具
- runtime: 运行时相关
- go/types: 类型检查器相关`)
}

func demonstrateGerritCommands() {
	fmt.Println("\n=== Gerrit 常用命令 ===")
	fmt.Println(`git-codereview 工具命令:

1. 初始化设置
   git codereview hooks

2. 创建新分支
   git checkout -b my-feature

3. 提交变更
   git add .
   git codereview change

4. 发送到 Gerrit
   git codereview mail

5. 更新 CL (根据审查反馈修改后)
   git add .
   git codereview change
   git codereview mail

6. 同步上游变更
   git codereview sync

7. 查看待处理的 CL
   git codereview pending

8. 放弃 CL
   git codereview abandon

注意事项:
- 每个 CL 应该是一个独立的、可审查的变更
- 大的变更应该拆分成多个小的 CL
- 保持 CL 专注于单一目的`)
}

func main() {
	fmt.Println("=== Go 标准库贡献指南 ===")
	fmt.Println()
	fmt.Println("本模块演示向 Go 标准库贡献代码的完整流程:")
	fmt.Println("1. 贡献准备工作")
	fmt.Println("2. 代码规范与风格")
	fmt.Println("3. 测试要求")
	fmt.Println("4. 代码审查流程")

	demonstrateContributionGuide()
	demonstrateCommitMessage()
	demonstrateGerritCommands()
	demonstrateCLWorkflow()

	fmt.Println("\n=== 标准库贡献指南演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 贡献前需要签署 CLA 并配置开发环境")
	fmt.Println("- 代码必须遵循 Go 的风格规范")
	fmt.Println("- 测试是贡献的必要组成部分")
	fmt.Println("- 使用 Gerrit 进行代码审查")
	fmt.Println("- 提交信息需要遵循特定格式")
	fmt.Println("- 耐心等待审查，积极响应反馈")
}
