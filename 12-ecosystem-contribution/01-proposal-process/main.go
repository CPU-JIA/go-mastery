// Package main 演示 Go 语言提案流程
// 本模块涵盖 Go 语言演进的核心机制：
// - 提案生命周期
// - 设计文档编写
// - 社区讨论与反馈
// - 提案评审与决策
package main

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// 提案状态与类型定义
// ============================================================================

// ProposalStatus 提案状态
type ProposalStatus int

const (
	ProposalStatusDraft ProposalStatus = iota
	ProposalStatusProposed
	ProposalStatusUnderReview
	ProposalStatusAccepted
	ProposalStatusDeclined
	ProposalStatusWithdrawn
	ProposalStatusImplemented
)

func (s ProposalStatus) String() string {
	switch s {
	case ProposalStatusDraft:
		return "Draft"
	case ProposalStatusProposed:
		return "Proposed"
	case ProposalStatusUnderReview:
		return "Under Review"
	case ProposalStatusAccepted:
		return "Accepted"
	case ProposalStatusDeclined:
		return "Declined"
	case ProposalStatusWithdrawn:
		return "Withdrawn"
	case ProposalStatusImplemented:
		return "Implemented"
	default:
		return "Unknown"
	}
}

// ProposalType 提案类型
type ProposalType int

const (
	ProposalTypeLanguage ProposalType = iota // 语言特性
	ProposalTypeLibrary                      // 标准库
	ProposalTypeTooling                      // 工具链
	ProposalTypeProcess                      // 流程改进
)

func (t ProposalType) String() string {
	switch t {
	case ProposalTypeLanguage:
		return "Language Change"
	case ProposalTypeLibrary:
		return "Standard Library"
	case ProposalTypeTooling:
		return "Tooling"
	case ProposalTypeProcess:
		return "Process"
	default:
		return "Unknown"
	}
}

// ============================================================================
// 提案核心结构
// ============================================================================

// Proposal Go 语言提案
type Proposal struct {
	ID           string
	Title        string
	Author       string
	Type         ProposalType
	Status       ProposalStatus
	Summary      string
	Motivation   string
	Design       *DesignDocument
	Discussion   *Discussion
	Timeline     *ProposalTimeline
	Reviewers    []string
	Stakeholders []string
	RelatedIssue string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// DesignDocument 设计文档
type DesignDocument struct {
	Abstract       string
	Background     string
	Proposal       string
	Rationale      string
	Compatibility  string
	Implementation string
	OpenIssues     []string
	Alternatives   []Alternative
	Examples       []CodeExample
}

// Alternative 替代方案
type Alternative struct {
	Name        string
	Description string
	Pros        []string
	Cons        []string
	WhyRejected string
}

// CodeExample 代码示例
type CodeExample struct {
	Title       string
	Description string
	Before      string
	After       string
}

// Discussion 讨论记录
type Discussion struct {
	Comments    []Comment
	Concerns    []Concern
	Supporters  []string
	Opposers    []string
	NeutralVote int
}

// Comment 评论
type Comment struct {
	Author    string
	Content   string
	Timestamp time.Time
	Replies   []Comment
}

// Concern 关注点
type Concern struct {
	Title       string
	Description string
	RaisedBy    string
	Status      ConcernStatus
	Resolution  string
}

// ConcernStatus 关注点状态
type ConcernStatus int

const (
	ConcernStatusOpen ConcernStatus = iota
	ConcernStatusAddressed
	ConcernStatusWontFix
)

// ProposalTimeline 提案时间线
type ProposalTimeline struct {
	DraftCreated        time.Time
	Proposed            time.Time
	ReviewStarted       time.Time
	DecisionMade        time.Time
	ImplementationStart time.Time
	ImplementationEnd   time.Time
}

// ============================================================================
// 提案管理器
// ============================================================================

// ProposalManager 提案管理器
type ProposalManager struct {
	proposals map[string]*Proposal
	reviewers []string
}

// NewProposalManager 创建提案管理器
func NewProposalManager() *ProposalManager {
	return &ProposalManager{
		proposals: make(map[string]*Proposal),
		reviewers: []string{
			"rsc",            // Russ Cox
			"robpike",        // Rob Pike
			"griesemer",      // Robert Griesemer
			"ianlancetaylor", // Ian Lance Taylor
		},
	}
}

// CreateProposal 创建提案
func (pm *ProposalManager) CreateProposal(title, author string, propType ProposalType) *Proposal {
	id := fmt.Sprintf("proposal-%d", len(pm.proposals)+1)

	proposal := &Proposal{
		ID:        id,
		Title:     title,
		Author:    author,
		Type:      propType,
		Status:    ProposalStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Timeline: &ProposalTimeline{
			DraftCreated: time.Now(),
		},
		Discussion: &Discussion{
			Comments: make([]Comment, 0),
			Concerns: make([]Concern, 0),
		},
	}

	pm.proposals[id] = proposal
	return proposal
}

// SubmitProposal 提交提案
func (pm *ProposalManager) SubmitProposal(proposalID string) error {
	proposal, exists := pm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("提案不存在: %s", proposalID)
	}

	if proposal.Status != ProposalStatusDraft {
		return fmt.Errorf("只有草稿状态的提案可以提交")
	}

	// 验证提案完整性
	if err := pm.validateProposal(proposal); err != nil {
		return fmt.Errorf("提案验证失败: %w", err)
	}

	proposal.Status = ProposalStatusProposed
	proposal.Timeline.Proposed = time.Now()
	proposal.UpdatedAt = time.Now()

	fmt.Printf("  [ProposalManager] 提案已提交: %s\n", proposal.Title)
	return nil
}

// validateProposal 验证提案
func (pm *ProposalManager) validateProposal(proposal *Proposal) error {
	if proposal.Title == "" {
		return fmt.Errorf("标题不能为空")
	}
	if proposal.Summary == "" {
		return fmt.Errorf("摘要不能为空")
	}
	if proposal.Motivation == "" {
		return fmt.Errorf("动机说明不能为空")
	}
	if proposal.Design == nil {
		return fmt.Errorf("设计文档不能为空")
	}
	return nil
}

// StartReview 开始评审
func (pm *ProposalManager) StartReview(proposalID string) error {
	proposal, exists := pm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("提案不存在: %s", proposalID)
	}

	if proposal.Status != ProposalStatusProposed {
		return fmt.Errorf("只有已提交的提案可以开始评审")
	}

	proposal.Status = ProposalStatusUnderReview
	proposal.Timeline.ReviewStarted = time.Now()
	proposal.Reviewers = pm.reviewers
	proposal.UpdatedAt = time.Now()

	fmt.Printf("  [ProposalManager] 提案评审已开始: %s\n", proposal.Title)
	fmt.Printf("  [ProposalManager] 评审人: %v\n", proposal.Reviewers)
	return nil
}

// AddComment 添加评论
func (pm *ProposalManager) AddComment(proposalID, author, content string) error {
	proposal, exists := pm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("提案不存在: %s", proposalID)
	}

	comment := Comment{
		Author:    author,
		Content:   content,
		Timestamp: time.Now(),
	}

	proposal.Discussion.Comments = append(proposal.Discussion.Comments, comment)
	proposal.UpdatedAt = time.Now()

	fmt.Printf("  [Discussion] %s: %s\n", author, content)
	return nil
}

// RaiseConcern 提出关注点
func (pm *ProposalManager) RaiseConcern(proposalID string, concern Concern) error {
	proposal, exists := pm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("提案不存在: %s", proposalID)
	}

	concern.Status = ConcernStatusOpen
	proposal.Discussion.Concerns = append(proposal.Discussion.Concerns, concern)
	proposal.UpdatedAt = time.Now()

	fmt.Printf("  [Concern] %s 提出关注: %s\n", concern.RaisedBy, concern.Title)
	return nil
}

// MakeDecision 做出决定
func (pm *ProposalManager) MakeDecision(proposalID string, accepted bool, reason string) error {
	proposal, exists := pm.proposals[proposalID]
	if !exists {
		return fmt.Errorf("提案不存在: %s", proposalID)
	}

	if proposal.Status != ProposalStatusUnderReview {
		return fmt.Errorf("只有评审中的提案可以做出决定")
	}

	if accepted {
		proposal.Status = ProposalStatusAccepted
		fmt.Printf("  [Decision] 提案已接受: %s\n", proposal.Title)
	} else {
		proposal.Status = ProposalStatusDeclined
		fmt.Printf("  [Decision] 提案已拒绝: %s\n", proposal.Title)
	}

	fmt.Printf("  [Decision] 原因: %s\n", reason)

	proposal.Timeline.DecisionMade = time.Now()
	proposal.UpdatedAt = time.Now()

	return nil
}

// GetProposalSummary 获取提案摘要
func (pm *ProposalManager) GetProposalSummary(proposalID string) {
	proposal, exists := pm.proposals[proposalID]
	if !exists {
		fmt.Printf("提案不存在: %s\n", proposalID)
		return
	}

	fmt.Printf("\n  提案摘要:\n")
	fmt.Printf("    ID: %s\n", proposal.ID)
	fmt.Printf("    标题: %s\n", proposal.Title)
	fmt.Printf("    作者: %s\n", proposal.Author)
	fmt.Printf("    类型: %s\n", proposal.Type)
	fmt.Printf("    状态: %s\n", proposal.Status)
	fmt.Printf("    创建时间: %s\n", proposal.CreatedAt.Format("2006-01-02"))
	fmt.Printf("    评论数: %d\n", len(proposal.Discussion.Comments))
	fmt.Printf("    关注点: %d\n", len(proposal.Discussion.Concerns))
}

// ============================================================================
// 设计文档生成器
// ============================================================================

// DesignDocGenerator 设计文档生成器
type DesignDocGenerator struct{}

// NewDesignDocGenerator 创建设计文档生成器
func NewDesignDocGenerator() *DesignDocGenerator {
	return &DesignDocGenerator{}
}

// GenerateTemplate 生成设计文档模板
func (g *DesignDocGenerator) GenerateTemplate(title string) string {
	template := fmt.Sprintf(`# Proposal: %s

Author(s): [Your Name]

Last updated: %s

Discussion at https://golang.org/issue/NNNNN

## Abstract

[A short summary of the proposal.]

## Background

[An introduction of the necessary background and the problem being solved by the proposed change.]

## Proposal

[A precise statement of the proposed change.]

## Rationale

[A discussion of alternate approaches and the trade offs, advantages, and disadvantages of the specified approach.]

## Compatibility

[A discussion of the change with regard to the Go 1 compatibility guidelines.]

## Implementation

[A description of the steps in the implementation, who will do them, and when.]

## Open issues (if applicable)

[A discussion of issues relating to this proposal for which the author does not know the solution.]
`, title, time.Now().Format("2006-01-02"))

	return template
}

// GenerateExample 生成示例设计文档
func (g *DesignDocGenerator) GenerateExample() *DesignDocument {
	return &DesignDocument{
		Abstract: "本提案建议在 Go 语言中添加泛型支持，允许编写类型参数化的函数和类型。",
		Background: `Go 语言自诞生以来一直缺乏泛型支持。开发者经常需要：
1. 为不同类型编写重复代码
2. 使用 interface{} 并进行类型断言
3. 使用代码生成工具

这些方法都有明显的缺点，影响了代码的可读性、类型安全性和性能。`,
		Proposal: `引入类型参数语法：
- 函数可以声明类型参数：func F[T any](x T) T
- 类型可以声明类型参数：type List[T any] struct { ... }
- 类型约束使用接口定义：type Ordered interface { ... }`,
		Rationale: `选择方括号语法的原因：
1. 与现有语法兼容性好
2. 解析器实现相对简单
3. 与其他语言的泛型语法有一定相似性`,
		Compatibility: "本提案完全向后兼容。现有的 Go 代码无需修改即可继续工作。",
		Implementation: `实现分为以下阶段：
1. 编译器前端支持类型参数解析
2. 类型检查器支持泛型类型推断
3. 代码生成支持泛型实例化
4. 标准库添加泛型容器和算法`,
		OpenIssues: []string{
			"类型参数的零值如何处理？",
			"是否支持类型参数的方法？",
			"泛型代码的编译性能如何优化？",
		},
		Alternatives: []Alternative{
			{
				Name:        "使用尖括号语法",
				Description: "func F<T>(x T) T",
				Pros:        []string{"与 C++/Java/C# 一致"},
				Cons:        []string{"与比较运算符冲突", "解析复杂"},
				WhyRejected: "解析歧义问题难以解决",
			},
			{
				Name:        "使用圆括号语法",
				Description: "func F(type T)(x T) T",
				Pros:        []string{"无解析歧义"},
				Cons:        []string{"与函数调用混淆", "语法冗长"},
				WhyRejected: "可读性较差",
			},
		},
		Examples: []CodeExample{
			{
				Title:       "泛型切片反转",
				Description: "使用泛型实现通用的切片反转函数",
				Before: `func ReverseInts(s []int) []int {
    result := make([]int, len(s))
    for i, v := range s {
        result[len(s)-1-i] = v
    }
    return result
}

func ReverseStrings(s []string) []string {
    result := make([]string, len(s))
    for i, v := range s {
        result[len(s)-1-i] = v
    }
    return result
}`,
				After: `func Reverse[T any](s []T) []T {
    result := make([]T, len(s))
    for i, v := range s {
        result[len(s)-1-i] = v
    }
    return result
}

// 使用
ints := Reverse([]int{1, 2, 3})
strs := Reverse([]string{"a", "b", "c"})`,
			},
		},
	}
}

// ============================================================================
// 提案流程模拟器
// ============================================================================

// ProposalSimulator 提案流程模拟器
type ProposalSimulator struct {
	manager *ProposalManager
	docGen  *DesignDocGenerator
}

// NewProposalSimulator 创建提案流程模拟器
func NewProposalSimulator() *ProposalSimulator {
	return &ProposalSimulator{
		manager: NewProposalManager(),
		docGen:  NewDesignDocGenerator(),
	}
}

// SimulateFullProcess 模拟完整提案流程
func (s *ProposalSimulator) SimulateFullProcess() {
	fmt.Println("\n=== 模拟 Go 提案完整流程 ===")

	// 1. 创建提案
	fmt.Println("\n步骤 1: 创建提案草稿")
	proposal := s.manager.CreateProposal(
		"添加迭代器支持",
		"gopher",
		ProposalTypeLanguage,
	)
	fmt.Printf("  创建提案: %s\n", proposal.Title)

	// 2. 编写设计文档
	fmt.Println("\n步骤 2: 编写设计文档")
	proposal.Summary = "本提案建议在 Go 语言中添加原生迭代器支持"
	proposal.Motivation = "简化集合遍历，支持惰性求值，提高代码可读性"
	proposal.Design = &DesignDocument{
		Abstract:       "添加 iter 包和迭代器语法支持",
		Background:     "当前 Go 的 range 只支持内置类型",
		Proposal:       "引入 iter.Seq[T] 类型和 for range 扩展",
		Rationale:      "保持 Go 的简洁性同时提供更强大的迭代能力",
		Compatibility:  "完全向后兼容",
		Implementation: "分阶段实现，预计 Go 1.23 发布",
	}
	fmt.Printf("  设计文档已完成\n")

	// 3. 提交提案
	fmt.Println("\n步骤 3: 提交提案")
	if err := s.manager.SubmitProposal(proposal.ID); err != nil {
		fmt.Printf("  提交失败: %v\n", err)
		return
	}

	// 4. 社区讨论
	fmt.Println("\n步骤 4: 社区讨论")
	s.manager.AddComment(proposal.ID, "developer1", "这个提案很有价值，可以大大简化代码")
	s.manager.AddComment(proposal.ID, "developer2", "需要考虑与现有 range 的兼容性")
	s.manager.AddComment(proposal.ID, "developer3", "性能影响如何？需要基准测试")

	// 5. 提出关注点
	fmt.Println("\n步骤 5: 提出关注点")
	s.manager.RaiseConcern(proposal.ID, Concern{
		Title:       "性能开销",
		Description: "迭代器可能引入额外的函数调用开销",
		RaisedBy:    "performance_expert",
	})
	s.manager.RaiseConcern(proposal.ID, Concern{
		Title:       "错误处理",
		Description: "迭代过程中的错误如何传播？",
		RaisedBy:    "error_handling_advocate",
	})

	// 6. 开始正式评审
	fmt.Println("\n步骤 6: 开始正式评审")
	if err := s.manager.StartReview(proposal.ID); err != nil {
		fmt.Printf("  评审启动失败: %v\n", err)
		return
	}

	// 7. 评审讨论
	fmt.Println("\n步骤 7: 评审讨论")
	s.manager.AddComment(proposal.ID, "rsc", "设计合理，与 Go 的设计哲学一致")
	s.manager.AddComment(proposal.ID, "robpike", "语法简洁，易于理解")
	s.manager.AddComment(proposal.ID, "ianlancetaylor", "实现方案可行，性能可接受")

	// 8. 做出决定
	fmt.Println("\n步骤 8: 做出决定")
	s.manager.MakeDecision(proposal.ID, true,
		"提案设计合理，社区反馈积极，实现方案可行")

	// 9. 显示最终状态
	s.manager.GetProposalSummary(proposal.ID)
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateProposalLifecycle() {
	fmt.Println("\n=== Go 提案生命周期 ===")
	fmt.Println(`提案状态流转:

  Draft (草稿)
    |
    v
  Proposed (已提交)
    |
    v
  Under Review (评审中)
    |
    +---> Accepted (已接受) ---> Implemented (已实现)
    |
    +---> Declined (已拒绝)
    |
    +---> Withdrawn (已撤回)

关键阶段:
1. 草稿阶段: 作者编写设计文档，收集初步反馈
2. 提交阶段: 正式提交到 golang/go 仓库
3. 评审阶段: Go 团队和社区进行讨论
4. 决策阶段: Go 团队做出最终决定
5. 实现阶段: 如果接受，开始实现工作`)
}

func demonstrateDesignDocument() {
	fmt.Println("\n=== 设计文档结构 ===")

	generator := NewDesignDocGenerator()
	doc := generator.GenerateExample()

	fmt.Println("\n设计文档核心部分:")
	fmt.Printf("\n1. 摘要 (Abstract):\n   %s\n", doc.Abstract)
	fmt.Printf("\n2. 背景 (Background):\n   %s\n",
		strings.ReplaceAll(doc.Background, "\n", "\n   "))
	fmt.Printf("\n3. 提案 (Proposal):\n   %s\n",
		strings.ReplaceAll(doc.Proposal, "\n", "\n   "))
	fmt.Printf("\n4. 理由 (Rationale):\n   %s\n",
		strings.ReplaceAll(doc.Rationale, "\n", "\n   "))
	fmt.Printf("\n5. 兼容性 (Compatibility):\n   %s\n", doc.Compatibility)

	fmt.Println("\n6. 待解决问题:")
	for i, issue := range doc.OpenIssues {
		fmt.Printf("   %d. %s\n", i+1, issue)
	}

	fmt.Println("\n7. 替代方案:")
	for _, alt := range doc.Alternatives {
		fmt.Printf("   - %s: %s\n", alt.Name, alt.WhyRejected)
	}
}

func demonstrateCommunityDiscussion() {
	fmt.Println("\n=== 社区讨论最佳实践 ===")
	fmt.Println(`有效参与提案讨论的建议:

1. 提供建设性反馈
   - 说明具体的使用场景
   - 提供代码示例
   - 指出潜在问题并建议解决方案

2. 尊重他人观点
   - 保持专业和礼貌
   - 承认不同观点的价值
   - 避免人身攻击

3. 关注技术细节
   - 讨论实现可行性
   - 考虑性能影响
   - 评估兼容性风险

4. 提供实际经验
   - 分享类似功能的使用经验
   - 提供基准测试数据
   - 展示真实世界的用例

5. 跟踪讨论进展
   - 定期查看更新
   - 回应他人的问题
   - 更新自己的观点`)
}

func demonstrateProposalTypes() {
	fmt.Println("\n=== 提案类型说明 ===")
	fmt.Println(`Go 提案主要分为以下类型:

1. 语言变更 (Language Change)
   - 新语法特性
   - 类型系统扩展
   - 运行时行为变更
   示例: 泛型、迭代器、错误处理改进

2. 标准库 (Standard Library)
   - 新包添加
   - 现有包扩展
   - API 变更
   示例: slices 包、maps 包、slog 包

3. 工具链 (Tooling)
   - 编译器改进
   - go 命令增强
   - 调试工具
   示例: go mod、go work、go vet 规则

4. 流程改进 (Process)
   - 发布流程
   - 贡献指南
   - 治理结构
   示例: 提案流程本身的改进`)
}

func main() {
	fmt.Println("=== Go 语言提案流程 ===")
	fmt.Println()
	fmt.Println("本模块演示 Go 语言演进的核心机制:")
	fmt.Println("1. 提案生命周期")
	fmt.Println("2. 设计文档编写")
	fmt.Println("3. 社区讨论与反馈")
	fmt.Println("4. 提案评审与决策")

	demonstrateProposalLifecycle()
	demonstrateProposalTypes()
	demonstrateDesignDocument()
	demonstrateCommunityDiscussion()

	// 模拟完整流程
	simulator := NewProposalSimulator()
	simulator.SimulateFullProcess()

	fmt.Println("\n=== 提案流程演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- 提案需要清晰的动机和设计文档")
	fmt.Println("- 社区讨论是提案成功的关键")
	fmt.Println("- Go 团队重视向后兼容性")
	fmt.Println("- 好的提案需要考虑替代方案")
	fmt.Println("- 实现计划应该具体可行")
}
