// Package main 演示代码审查最佳实践
package main

import (
	"fmt"
	"strings"
	"time"
)

// ReviewCategory 审查类别
type ReviewCategory int

const (
	CategoryCorrectness ReviewCategory = iota
	CategorySecurity
	CategoryPerformance
	CategoryMaintainability
	CategoryReadability
)

func (c ReviewCategory) String() string {
	names := []string{"正确性", "安全性", "性能", "可维护性", "可读性"}
	if int(c) < len(names) {
		return names[c]
	}
	return "未知"
}

// Severity 严重程度
type Severity int

const (
	SeverityNit Severity = iota
	SeverityWarning
	SeverityError
	SeverityBlocker
)

func (s Severity) String() string {
	names := []string{"Nit", "Warning", "Error", "Blocker"}
	if int(s) < len(names) {
		return names[s]
	}
	return "Unknown"
}

// ReviewComment 审查评论
type ReviewComment struct {
	File       string
	Line       int
	Category   ReviewCategory
	Severity   Severity
	Message    string
	Suggestion string
}

func (c *ReviewComment) String() string {
	return fmt.Sprintf("[%s][%s] %s:%d - %s", c.Severity, c.Category, c.File, c.Line, c.Message)
}

// ChecklistItem 清单项
type ChecklistItem struct {
	ID          string
	Category    ReviewCategory
	Description string
	Question    string
}

// ReviewChecklist 审查清单
type ReviewChecklist struct {
	items []ChecklistItem
}

// NewReviewChecklist 创建审查清单
func NewReviewChecklist() *ReviewChecklist {
	return &ReviewChecklist{
		items: []ChecklistItem{
			{"CORRECT-001", CategoryCorrectness, "边界条件", "是否处理了空值、零值、最大值？"},
			{"CORRECT-002", CategoryCorrectness, "错误处理", "所有错误是否都被正确处理？"},
			{"CORRECT-003", CategoryCorrectness, "并发安全", "共享资源访问是否安全？"},
			{"SECURITY-001", CategorySecurity, "输入验证", "外部输入是否经过验证？"},
			{"SECURITY-002", CategorySecurity, "敏感数据", "敏感数据是否得到保护？"},
			{"PERF-001", CategoryPerformance, "资源管理", "资源是否被正确释放？"},
			{"PERF-002", CategoryPerformance, "算法效率", "算法复杂度是否合理？"},
			{"MAINTAIN-001", CategoryMaintainability, "单一职责", "函数是否只做一件事？"},
			{"READ-001", CategoryReadability, "命名清晰", "命名是否清晰表达意图？"},
		},
	}
}

// Print 打印清单
func (c *ReviewChecklist) Print() {
	fmt.Println("\n代码审查清单:")
	for _, item := range c.items {
		fmt.Printf("  [ ] %s: %s\n      %s\n", item.ID, item.Description, item.Question)
	}
}

// CodeSmell 代码异味
type CodeSmell struct {
	Name    string
	Example string
	Fix     string
}

var commonSmells = []CodeSmell{
	{"忽略错误", `file, _ := os.Open(f)`, `file, err := os.Open(f); if err != nil { return err }`},
	{"过深嵌套", `if a { if b { if c { } } }`, `使用早返回减少嵌套`},
	{"魔法数字", `if len(s) < 8 {`, `const MinLen = 8; if len(s) < MinLen {`},
	{"资源泄漏", `file, _ := os.Open(f) // 没有 Close`, `defer file.Close()`},
}

// ReviewReport 审查报告
type ReviewReport struct {
	PRTitle   string
	Author    string
	Reviewer  string
	Comments  []ReviewComment
	Approved  bool
	Summary   string
	CreatedAt time.Time
}

// NewReviewReport 创建报告
func NewReviewReport(title, author, reviewer string) *ReviewReport {
	return &ReviewReport{
		PRTitle:   title,
		Author:    author,
		Reviewer:  reviewer,
		CreatedAt: time.Now(),
	}
}

// AddComment 添加评论
func (r *ReviewReport) AddComment(c ReviewComment) {
	r.Comments = append(r.Comments, c)
}

// Complete 完成审查
func (r *ReviewReport) Complete(approved bool, summary string) {
	r.Approved = approved
	r.Summary = summary
}

// String 报告字符串
func (r *ReviewReport) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("PR: %s\n", r.PRTitle))
	sb.WriteString(fmt.Sprintf("作者: %s | 审查者: %s\n", r.Author, r.Reviewer))
	status := "需要修改"
	if r.Approved {
		status = "已批准"
	}
	sb.WriteString(fmt.Sprintf("状态: %s\n", status))
	sb.WriteString(fmt.Sprintf("总结: %s\n", r.Summary))
	sb.WriteString(fmt.Sprintf("评论数: %d\n", len(r.Comments)))
	for i, c := range r.Comments {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, c.String()))
	}
	return sb.String()
}

func main() {
	fmt.Println("=== 代码审查最佳实践 ===")

	// 审查清单
	checklist := NewReviewChecklist()
	checklist.Print()

	// 常见代码异味
	fmt.Println("\n常见代码异味:")
	for i, smell := range commonSmells {
		fmt.Printf("  %d. %s\n     问题: %s\n     修复: %s\n", i+1, smell.Name, smell.Example, smell.Fix)
	}

	// 反馈原则
	fmt.Println("\n反馈原则:")
	principles := []string{
		"对事不对人 - 评论代码，不评论人",
		"具体明确 - 指出具体问题和解决方案",
		"解释原因 - 说明为什么这样更好",
		"区分优先级 - 标记必须修复 vs 建议",
	}
	for _, p := range principles {
		fmt.Printf("  - %s\n", p)
	}

	// 示例报告
	fmt.Println("\n=== 审查报告示例 ===")
	report := NewReviewReport("feat: 用户认证", "dev@example.com", "reviewer@example.com")
	report.AddComment(ReviewComment{
		File: "auth/handler.go", Line: 42, Category: CategorySecurity,
		Severity: SeverityError, Message: "密码明文日志", Suggestion: "移除或掩码",
	})
	report.AddComment(ReviewComment{
		File: "auth/service.go", Line: 78, Category: CategoryCorrectness,
		Severity: SeverityWarning, Message: "错误被忽略", Suggestion: "处理错误",
	})
	report.Complete(false, "有安全问题需修复")
	fmt.Println(report.String())

	fmt.Println("\n关键学习点:")
	fmt.Println("- 使用清单确保审查全面")
	fmt.Println("- 识别常见代码异味")
	fmt.Println("- 提供建设性反馈")
	fmt.Println("- 区分问题优先级")
}
