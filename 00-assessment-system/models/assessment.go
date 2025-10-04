/*
=== Go语言学习评估系统 - 评估框架数据模型 ===

本文件定义了完整的评估框架和相关数据结构：
1. 多维度评估框架定义
2. 评估任务和标准规范
3. 评分规则和权重体系
4. 评估结果分析模型
5. 认证考试框架
6. 自动化评估配置
7. 评估报告生成模型

作者: JIA
创建时间: 2025-10-03
版本: 1.0.0
*/

// Package models 提供Go语言学习评估系统的核心数据模型
//
// 本包定义了评估系统中所有关键的数据结构，包括评估结果、能力模型、
// 学习进度跟踪等。这些模型是整个评估系统的基础架构。
package models

import (
	"assessment-system/evaluators"
	"encoding/json"
	"time"
)

// AssessmentFramework 评估框架定义
type AssessmentFramework struct {
	Version     string    `json:"version"`     // 框架版本
	Name        string    `json:"name"`        // 框架名称
	Description string    `json:"description"` // 框架描述
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间

	// 评估维度定义
	Dimensions   []AssessmentDimension  `json:"dimensions"`    // 评估维度
	WeightMatrix WeightMatrix           `json:"weight_matrix"` // 权重矩阵
	ScoringRules map[string]ScoringRule `json:"scoring_rules"` // 评分规则

	// 阈值和标准
	Thresholds        map[string]float64          `json:"thresholds"`         // 各种阈值设定
	Standards         map[string]QualityStandard  `json:"standards"`          // 质量标准
	LevelRequirements map[string]LevelRequirement `json:"level_requirements"` // 等级要求

	// 配置选项
	AutoAssessment bool `json:"auto_assessment"` // 是否启用自动评估
	PeerReview     bool `json:"peer_review"`     // 是否启用同伴评审
	MentorReview   bool `json:"mentor_review"`   // 是否启用导师评审
}

// AssessmentDimension 评估维度定义
type AssessmentDimension struct {
	ID               string                   `json:"id"`                // 维度唯一标识
	Name             string                   `json:"name"`              // 维度名称
	Description      string                   `json:"description"`       // 维度描述
	Weight           float64                  `json:"weight"`            // 权重 (0.0-1.0)
	MaxScore         float64                  `json:"max_score"`         // 最高分
	Subdimensions    []AssessmentSubdimension `json:"subdimensions"`     // 子维度
	Metrics          []PerformanceMetric      `json:"metrics"`           // 性能指标
	EvaluationMethod string                   `json:"evaluation_method"` // 评估方法
}

// AssessmentSubdimension 评估子维度
type AssessmentSubdimension struct {
	ID              string               `json:"id"`               // 子维度标识
	Name            string               `json:"name"`             // 子维度名称
	Description     string               `json:"description"`      // 详细描述
	Weight          float64              `json:"weight"`           // 在父维度中的权重
	Criteria        []EvaluationCriteria `json:"criteria"`         // 评估标准
	AutomatedChecks []AutomatedCheck     `json:"automated_checks"` // 自动化检查
}

// EvaluationCriteria 评估标准
type EvaluationCriteria struct {
	ID          string  `json:"id"`          // 标准唯一标识
	Name        string  `json:"name"`        // 标准名称
	Description string  `json:"description"` // 详细描述
	Type        string  `json:"type"`        // 标准类型: binary, scale, rubric
	Weight      float64 `json:"weight"`      // 权重

	// 评分标准定义
	ScaleDefinition []ScalePoint      `json:"scale_definition"` // 量表定义
	RubricLevels    []RubricLevel     `json:"rubric_levels"`    // 评分等级
	Examples        []CriteriaExample `json:"examples"`         // 示例

	// 自动化配置
	Automated     bool                   `json:"automated"`      // 是否可自动评估
	CheckFunction string                 `json:"check_function"` // 检查函数名
	Parameters    map[string]interface{} `json:"parameters"`     // 检查参数
}

// ScalePoint 量表点定义
type ScalePoint struct {
	Value       float64 `json:"value"`       // 分值
	Label       string  `json:"label"`       // 标签
	Description string  `json:"description"` // 描述
}

// RubricLevel 评分等级定义
type RubricLevel struct {
	Level       int      `json:"level"`       // 等级 (1-5)
	Name        string   `json:"name"`        // 等级名称
	Description string   `json:"description"` // 等级描述
	Score       float64  `json:"score"`       // 对应分数
	Indicators  []string `json:"indicators"`  // 指标描述
}

// CriteriaExample 标准示例
type CriteriaExample struct {
	Type        string  `json:"type"`        // 示例类型: good, bad, average
	Code        string  `json:"code"`        // 代码示例
	Score       float64 `json:"score"`       // 示例评分
	Explanation string  `json:"explanation"` // 解释说明
}

// AutomatedCheck 自动化检查定义
type AutomatedCheck struct {
	ID         string                 `json:"id"`         // 检查唯一标识
	Name       string                 `json:"name"`       // 检查名称
	Type       string                 `json:"type"`       // 检查类型: static, dynamic, test
	Tool       string                 `json:"tool"`       // 使用工具: golint, govet, gocyclo, etc.
	Command    string                 `json:"command"`    // 执行命令
	Parameters map[string]interface{} `json:"parameters"` // 参数配置
	Weight     float64                `json:"weight"`     // 在子维度中的权重

	// 结果处理
	ResultParser string             `json:"result_parser"` // 结果解析器
	ScoreMapping map[string]float64 `json:"score_mapping"` // 分数映射
	Thresholds   map[string]float64 `json:"thresholds"`    // 阈值设定
}

// PerformanceMetric 性能指标定义
type PerformanceMetric struct {
	ID          string  `json:"id"`          // 指标唯一标识
	Name        string  `json:"name"`        // 指标名称
	Unit        string  `json:"unit"`        // 计量单位
	Target      float64 `json:"target"`      // 目标值
	Threshold   float64 `json:"threshold"`   // 阈值
	Weight      float64 `json:"weight"`      // 权重
	Aggregation string  `json:"aggregation"` // 聚合方式: avg, max, min, sum
}

// ScoringRule 评分规则定义
type ScoringRule struct {
	ID         string                 `json:"id"`         // 规则唯一标识
	Name       string                 `json:"name"`       // 规则名称
	Type       string                 `json:"type"`       // 规则类型: weighted_sum, formula, composite
	Formula    string                 `json:"formula"`    // 计算公式
	Parameters map[string]interface{} `json:"parameters"` // 规则参数
	Conditions []ScoringCondition     `json:"conditions"` // 条件设定
}

// ScoringCondition 评分条件
type ScoringCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符: >, <, ==, !=, >=, <=
	Value    interface{} `json:"value"`    // 比较值
	Score    float64     `json:"score"`    // 满足条件时的分数
	Modifier float64     `json:"modifier"` // 分数修饰符
}

// WeightMatrix 权重矩阵
type WeightMatrix struct {
	// 主要维度权重
	TechnicalDepth      float64 `json:"technical_depth"`      // 技术深度权重
	EngineeringPractice float64 `json:"engineering_practice"` // 工程实践权重
	ProjectExperience   float64 `json:"project_experience"`   // 项目经验权重
	SoftSkills          float64 `json:"soft_skills"`          // 软技能权重

	// 评估方法权重
	AutomatedAssessment float64 `json:"automated_assessment"` // 自动化评估权重
	CodeReview          float64 `json:"code_review"`          // 代码审查权重
	ProjectEvaluation   float64 `json:"project_evaluation"`   // 项目评估权重
	PeerFeedback        float64 `json:"peer_feedback"`        // 同伴反馈权重
	MentorAssessment    float64 `json:"mentor_assessment"`    // 导师评估权重

	// 不同阶段权重调整
	StageWeights map[int]StageWeight `json:"stage_weights"` // 各阶段权重调整
}

// StageWeight 阶段权重调整
type StageWeight struct {
	Stage               int     `json:"stage"`                // 阶段编号
	TechnicalFocus      float64 `json:"technical_focus"`      // 技术重点权重调整
	PracticalFocus      float64 `json:"practical_focus"`      // 实践重点权重调整
	ProjectComplexity   float64 `json:"project_complexity"`   // 项目复杂度要求
	CommunicationWeight float64 `json:"communication_weight"` // 沟通技能权重
}

// QualityStandard 质量标准定义
type QualityStandard struct {
	Category    string                  `json:"category"`    // 标准类别
	Name        string                  `json:"name"`        // 标准名称
	Description string                  `json:"description"` // 标准描述
	Levels      map[string]QualityLevel `json:"levels"`      // 各质量等级
	Metrics     []QualityMetric         `json:"metrics"`     // 质量指标
}

// QualityLevel 质量等级定义
type QualityLevel struct {
	Level        string   `json:"level"`        // 等级名称
	Score        float64  `json:"score"`        // 等级分数
	Description  string   `json:"description"`  // 等级描述
	Requirements []string `json:"requirements"` // 要求列表
	Examples     []string `json:"examples"`     // 示例
}

// QualityMetric 质量指标
type QualityMetric struct {
	Name        string  `json:"name"`        // 指标名称
	Target      float64 `json:"target"`      // 目标值
	MinValue    float64 `json:"min_value"`   // 最小值
	MaxValue    float64 `json:"max_value"`   // 最大值
	Unit        string  `json:"unit"`        // 单位
	Description string  `json:"description"` // 描述
}

// LevelRequirement 等级要求定义
type LevelRequirement struct {
	Level               string               `json:"level"`                // 等级名称
	MinScore            float64              `json:"min_score"`            // 最低分数要求
	RequiredStages      []int                `json:"required_stages"`      // 必须完成的阶段
	ProjectRequirements []ProjectRequirement `json:"project_requirements"` // 项目要求

	// 技能要求
	TechnicalSkills map[string]int `json:"technical_skills"` // 技术技能要求 (技能名 -> 等级)
	SoftSkills      map[string]int `json:"soft_skills"`      // 软技能要求

	// 考试要求
	ExamRequired bool    `json:"exam_required"` // 是否需要考试
	ExamType     string  `json:"exam_type"`     // 考试类型
	ExamDuration int     `json:"exam_duration"` // 考试时长(分钟)
	PassingScore float64 `json:"passing_score"` // 及格分数
}

// ProjectRequirement 项目要求
type ProjectRequirement struct {
	Type             string   `json:"type"`              // 项目类型
	MinComplexity    int      `json:"min_complexity"`    // 最低复杂度
	RequiredFeatures []string `json:"required_features"` // 必需功能
	TechStack        []string `json:"tech_stack"`        // 技术栈要求
	MinScore         float64  `json:"min_score"`         // 最低项目评分
}

// AssessmentTask 评估任务定义
type AssessmentTask struct {
	ID            string `json:"id"`             // 任务唯一标识
	Name          string `json:"name"`           // 任务名称
	Description   string `json:"description"`    // 任务描述
	Type          string `json:"type"`           // 任务类型: coding, design, analysis, presentation
	Stage         int    `json:"stage"`          // 适用学习阶段
	Difficulty    int    `json:"difficulty"`     // 难度等级 (1-5)
	EstimatedTime int    `json:"estimated_time"` // 预估完成时间(分钟)

	// 任务内容
	Instructions string         `json:"instructions"`  // 详细说明
	Requirements []string       `json:"requirements"`  // 具体要求
	Resources    []TaskResource `json:"resources"`     // 参考资源
	StartingCode string         `json:"starting_code"` // 起始代码
	TestCases    []TestCase     `json:"test_cases"`    // 测试用例

	// 评估配置
	EvaluationCriteria []EvaluationCriteria `json:"evaluation_criteria"` // 评估标准
	AutoGrading        bool                 `json:"auto_grading"`        // 自动评分
	TimeLimit          *int                 `json:"time_limit"`          // 时间限制
	MaxAttempts        *int                 `json:"max_attempts"`        // 最大尝试次数

	// 元数据
	CreatedBy string    `json:"created_by"` // 创建者
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
	Tags      []string  `json:"tags"`       // 标签
	Keywords  []string  `json:"keywords"`   // 关键词
}

// TaskResource 任务资源
type TaskResource struct {
	Type        string `json:"type"`        // 资源类型: doc, video, link, code
	Title       string `json:"title"`       // 资源标题
	URL         string `json:"url"`         // 资源链接
	Description string `json:"description"` // 资源描述
	Essential   bool   `json:"essential"`   // 是否必需
}

// TestCase 测试用例
type TestCase struct {
	ID          string      `json:"id"`          // 测试用例标识
	Name        string      `json:"name"`        // 用例名称
	Input       interface{} `json:"input"`       // 输入数据
	Expected    interface{} `json:"expected"`    // 期望输出
	Description string      `json:"description"` // 用例描述
	Weight      float64     `json:"weight"`      // 权重
	Hidden      bool        `json:"hidden"`      // 是否隐藏
}

// AssessmentSession 评估会话
type AssessmentSession struct {
	ID        string `json:"id"`         // 会话唯一标识
	StudentID string `json:"student_id"` // 学习者标识
	TaskID    string `json:"task_id"`    // 任务标识
	Type      string `json:"type"`       // 评估类型
	Status    string `json:"status"`     // 状态: started, in_progress, completed, expired

	// 时间记录
	StartTime     time.Time  `json:"start_time"`     // 开始时间
	EndTime       *time.Time `json:"end_time"`       // 结束时间
	Duration      int        `json:"duration"`       // 实际用时(分钟)
	TimeRemaining *int       `json:"time_remaining"` // 剩余时间

	// 提交内容
	Submissions     []TaskSubmission `json:"submissions"`      // 提交记录
	FinalSubmission *TaskSubmission  `json:"final_submission"` // 最终提交

	// 评估结果
	Results  *AssessmentResult    `json:"results"`  // 评估结果
	Feedback []AssessmentFeedback `json:"feedback"` // 反馈信息

	// 元数据
	AttemptNumber int               `json:"attempt_number"` // 尝试次数
	IPAddress     string            `json:"ip_address"`     // IP地址
	UserAgent     string            `json:"user_agent"`     // 用户代理
	Environment   map[string]string `json:"environment"`    // 环境信息
}

// TaskSubmission 任务提交
type TaskSubmission struct {
	ID          string                 `json:"id"`           // 提交标识
	Timestamp   time.Time              `json:"timestamp"`    // 提交时间
	Code        string                 `json:"code"`         // 提交代码
	Files       map[string]string      `json:"files"`        // 提交文件
	Output      string                 `json:"output"`       // 运行输出
	TestResults []TestResult           `json:"test_results"` // 测试结果
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
}

// TestResult 测试结果
type TestResult struct {
	TestCaseID    string      `json:"test_case_id"`   // 测试用例标识
	Passed        bool        `json:"passed"`         // 是否通过
	ActualOutput  interface{} `json:"actual_output"`  // 实际输出
	ExecutionTime float64     `json:"execution_time"` // 执行时间
	Error         *string     `json:"error"`          // 错误信息
	Score         float64     `json:"score"`          // 得分
}

// AssessmentResult 评估结果
type AssessmentResult struct {
	SessionID    string  `json:"session_id"`    // 会话标识
	OverallScore float64 `json:"overall_score"` // 总分
	MaxScore     float64 `json:"max_score"`     // 满分
	Percentage   float64 `json:"percentage"`    // 得分率
	Grade        string  `json:"grade"`         // 等级

	// 维度得分
	DimensionScores map[string]float64 `json:"dimension_scores"` // 各维度得分
	CriteriaScores  map[string]float64 `json:"criteria_scores"`  // 各标准得分

	// 详细结果
	TestResults        []TestResult        `json:"test_results"`        // 测试结果
	CodeAnalysis       *CodeAnalysisResult `json:"code_analysis"`       // 代码分析
	PerformanceMetrics map[string]float64  `json:"performance_metrics"` // 性能指标

	// 统计信息
	CompletionTime int `json:"completion_time"` // 完成用时
	AttemptCount   int `json:"attempt_count"`   // 尝试次数
	HintUsed       int `json:"hint_used"`       // 使用提示次数
}

// CodeAnalysisResult 代码分析结果
type CodeAnalysisResult struct {
	LinesOfCode          int     `json:"lines_of_code"`         // 代码行数
	CyclomaticComplexity int     `json:"cyclomatic_complexity"` // 圈复杂度
	TestCoverage         float64 `json:"test_coverage"`         // 测试覆盖率
	CodeQuality          float64 `json:"code_quality"`          // 代码质量分
	SecurityScore        float64 `json:"security_score"`        // 安全评分
	PerformanceScore     float64 `json:"performance_score"`     // 性能评分

	// 详细分析
	Issues      []CodeIssue      `json:"issues"`      // 代码问题
	Suggestions []CodeSuggestion `json:"suggestions"` // 改进建议
	Patterns    []DesignPattern  `json:"patterns"`    // 设计模式使用
}

// CodeIssue 代码问题
type CodeIssue struct {
	Type       string `json:"type"`       // 问题类型
	Severity   string `json:"severity"`   // 严重程度
	Line       int    `json:"line"`       // 行号
	Column     int    `json:"column"`     // 列号
	Message    string `json:"message"`    // 问题描述
	Rule       string `json:"rule"`       // 规则名称
	Suggestion string `json:"suggestion"` // 修复建议
}

// CodeSuggestion 代码建议
type CodeSuggestion struct {
	Type        string `json:"type"`        // 建议类型
	Priority    string `json:"priority"`    // 优先级
	Description string `json:"description"` // 建议描述
	Example     string `json:"example"`     // 示例代码
	Impact      string `json:"impact"`      // 预期影响
}

// DesignPattern 设计模式
type DesignPattern struct {
	Name        string  `json:"name"`        // 模式名称
	Usage       string  `json:"usage"`       // 使用情况
	Appropriate bool    `json:"appropriate"` // 是否恰当使用
	Score       float64 `json:"score"`       // 使用评分
}

// AssessmentFeedback 评估反馈
type AssessmentFeedback struct {
	Type      string    `json:"type"`      // 反馈类型: automated, peer, mentor
	Source    string    `json:"source"`    // 反馈来源
	Timestamp time.Time `json:"timestamp"` // 反馈时间
	Content   string    `json:"content"`   // 反馈内容
	Rating    *float64  `json:"rating"`    // 评分(可选)
	Helpful   *bool     `json:"helpful"`   // 是否有帮助(可选)
	Category  string    `json:"category"`  // 反馈分类
	Tags      []string  `json:"tags"`      // 标签
}

// NewAssessmentFramework 创建新的评估框架实例
//
// 功能说明:
//
//	本函数创建并初始化一个完整的评估框架对象，用于定义Go语言学习的
//	多维度评估标准体系。框架包含维度定义、权重矩阵、评分规则等核心要素。
//
// 参数:
//   - version: 框架版本号（如"1.0.0"），用于框架升级和兼容性管理
//   - name: 框架名称（如"Go语言能力评估框架"），用于标识和展示
//
// 返回值:
//   - *AssessmentFramework: 初始化完成的评估框架指针，包含以下默认配置：
//   - 默认权重矩阵（技术深度40%、工程实践30%等）
//   - 默认阈值（及格70分、优秀90分等）
//   - 四级认证要求（Bronze、Silver、Gold、Platinum）
//   - 启用自动评估和导师评审，关闭同伴评审
//
// 使用场景:
//   - 系统初始化时创建评估标准
//   - 为不同学习阶段定制评估框架
//   - 实现多版本评估标准的并行管理
//
// 示例:
//
//	framework := NewAssessmentFramework("1.0.0", "Go语言能力评估框架")
//	framework.AutoAssessment = true  // 启用自动评估
//	framework.Dimensions = append(framework.Dimensions, customDimension)
//
// 注意事项:
//   - 创建后的框架可以通过修改字段进行定制
//   - 默认配置基于教学经验，适合大多数学习场景
//   - 时间戳自动设置为当前时间
//
// 作者: JIA
func NewAssessmentFramework(version, name string) *AssessmentFramework {
	now := time.Now()
	return &AssessmentFramework{
		Version:           version,
		Name:              name,
		CreatedAt:         now,
		UpdatedAt:         now,
		Dimensions:        []AssessmentDimension{},
		WeightMatrix:      getDefaultWeightMatrix(),
		ScoringRules:      make(map[string]ScoringRule),
		Thresholds:        getDefaultThresholds(),
		Standards:         make(map[string]QualityStandard),
		LevelRequirements: getDefaultLevelRequirements(),
		AutoAssessment:    true,
		PeerReview:        false,
		MentorReview:      true,
	}
}

// getDefaultWeightMatrix 获取默认权重矩阵配置
//
// 功能说明:
//
//	本函数返回评估系统的默认权重分配矩阵，定义了各评估维度和评估方法的
//	相对重要性。权重矩阵是评估框架的核心，直接影响最终得分的计算。
//
// 权重分配逻辑:
//
//	主要维度权重（总和=1.0）:
//	  - 技术深度: 40% (TechnicalDepth) - 考察Go语言核心知识掌握程度
//	  - 工程实践: 30% (EngineeringPractice) - 考察工程化开发能力
//	  - 项目经验: 20% (ProjectExperience) - 考察实战项目完成质量
//	  - 软技能: 10% (SoftSkills) - 考察沟通、文档等能力
//
//	评估方法权重（总和=1.0）:
//	  - 自动化评估: 50% (AutomatedAssessment) - 代码质量工具分析
//	  - 代码审查: 30% (CodeReview) - 人工代码审查
//	  - 项目评估: 15% (ProjectEvaluation) - 项目完整性评估
//	  - 同伴反馈: 3% (PeerFeedback) - 学习者互评
//	  - 导师评估: 2% (MentorAssessment) - 导师专业评估
//
// 返回值:
//   - WeightMatrix: 包含所有默认权重配置的权重矩阵结构体，
//     StageWeights初始化为空map，可后续按阶段定制
//
// 设计理念:
//   - 重视技术深度和自动化评估，确保客观性
//   - 平衡理论知识与实践能力
//   - 为不同学习阶段保留权重调整空间
//
// 使用场景:
//   - NewAssessmentFramework中初始化默认权重
//   - 创建标准评估框架时提供基准配置
//   - 作为自定义权重配置的参考模板
//
// 作者: JIA
func getDefaultWeightMatrix() WeightMatrix {
	return WeightMatrix{
		TechnicalDepth:      WeightTechnicalDepth,
		EngineeringPractice: WeightEngineeringPractice,
		ProjectExperience:   WeightProjectExperience,
		SoftSkills:          WeightSoftSkills,

		AutomatedAssessment: WeightAutomatedAssessment,
		CodeReview:          WeightCodeReview,
		ProjectEvaluation:   WeightProjectEvaluation,
		PeerFeedback:        WeightPeerFeedback,
		MentorAssessment:    WeightMentorAssessment,

		StageWeights: make(map[int]StageWeight),
	}
}

// getDefaultThresholds 获取默认评估阈值配置
//
// 功能说明:
//
//	本函数返回评估系统使用的所有关键阈值配置。这些阈值定义了及格标准、
//	优秀标准、最低质量要求等核心评判基准，是评估决策的重要依据。
//
// 阈值配置说明:
//
//   - passing_score (70.0): 及格分数线
//     最低通过标准，低于此分数视为未达标
//
//   - excellent_score (90.0): 优秀分数线
//     优秀水平标准，超过此分数可获得优秀认证
//
//   - min_coverage (80.0): 最低测试覆盖率要求
//     代码测试覆盖率需达到80%以上，确保质量
//
//   - max_complexity (10.0): 最大圈复杂度允许值
//     单函数圈复杂度超过10视为过于复杂，需要重构
//
//   - min_documentation (70.0): 最低文档完整度分数
//     文档评分需达到70分以上，确保可维护性
//
//   - performance_target (95.0): 性能目标分数
//     性能优化目标，接近此分数表示性能优秀
//
// 返回值:
//   - map[string]float64: 键值对映射，键为阈值名称，值为阈值数值
//
// 设计考量:
//   - 及格线70分符合教育学惯例
//   - 测试覆盖率80%平衡了质量与开发效率
//   - 复杂度10是业界公认的可维护性临界点
//
// 使用场景:
//   - 初始化评估框架时设置默认阈值
//   - 自动化评估工具进行质量判定
//   - 生成评估报告时提供参考标准
//
// 作者: JIA
func getDefaultThresholds() map[string]float64 {
	return map[string]float64{
		"passing_score":      ThresholdPassingScore,
		"excellent_score":    ThresholdExcellentScore,
		"min_coverage":       ThresholdMinCoverage,
		"max_complexity":     ThresholdMaxComplexity,
		"min_documentation":  ThresholdMinDocumentation,
		"performance_target": ThresholdPerformanceTarget,
	}
}

// getDefaultLevelRequirements 获取默认认证等级要求配置
//
// 功能说明:
//
//	本函数返回"Go从入门到通天"学习路径的四级认证体系要求。每个等级定义了
//	完整的准入标准，包括分数要求、必修阶段、考试配置等，形成递进式能力认证。
//
// 四级认证体系详解:
//
//	🥉 Bronze（青铜级）- 入门认证
//	  · 最低分数: 70分（及格线）
//	  · 必修阶段: 1-3阶段（基础语法、进阶特性、并发编程）
//	  · 考试类型: practical（实践考试）
//	  · 考试时长: 120分钟
//	  · 及格分数: 80分
//	  适合：完成Go基础学习，具备基本开发能力的初学者
//
//	🥈 Silver（白银级）- 熟练认证
//	  · 最低分数: 80分（良好水平）
//	  · 必修阶段: 1-6阶段（增加Web开发、微服务、项目实战）
//	  · 考试类型: comprehensive（综合考试）
//	  · 考试时长: 180分钟
//	  · 及格分数: 85分
//	  适合：掌握Go核心技术，能够独立完成项目的开发者
//
//	🥇 Gold（黄金级）- 精通认证
//	  · 最低分数: 85分（优秀水平）
//	  · 必修阶段: 1-10阶段（增加性能优化、运行时原理、系统编程）
//	  · 考试类型: advanced（高级考试）
//	  · 考试时长: 240分钟
//	  · 及格分数: 90分
//	  适合：深度掌握Go技术栈，能够进行架构设计和性能调优的高级工程师
//
//	💎 Platinum（白金级）- 专家认证
//	  · 最低分数: 90分（卓越水平）
//	  · 必修阶段: 全部15阶段（包含编译器、大规模系统、开源贡献等）
//	  · 考试类型: expert（专家考试）
//	  · 考试时长: 300分钟
//	  · 及格分数: 95分
//	  适合：Go语言专家，能够参与语言生态建设和技术领导的顶尖开发者
//
// 返回值:
//   - map[string]LevelRequirement: 以等级名称为键的认证要求映射
//
// 设计理念:
//   - 渐进式难度提升，确保学习者稳步成长
//   - 考试时长随复杂度增加，反映真实能力要求
//   - 每个等级的及格分数高于最低准入分数，保证质量
//   - 阶段覆盖从基础到高级，形成完整知识体系
//
// 使用场景:
//   - 评估系统初始化时加载认证标准
//   - 学习者查看认证要求和学习路径
//   - 自动判定学习者当前可申请的认证等级
//
// 作者: JIA
func getDefaultLevelRequirements() map[string]LevelRequirement {
	return map[string]LevelRequirement{
		"Bronze": {
			Level:          "Bronze",
			MinScore:       evaluators.PassingScore,
			RequiredStages: []int{1, 2, 3},
			ExamRequired:   true,
			ExamType:       "practical",
			ExamDuration:   evaluators.ExamDurationBronze,
			PassingScore:   evaluators.DefaultScore,
		},
		"Silver": {
			Level:          "Silver",
			MinScore:       evaluators.DefaultScore,
			RequiredStages: []int{1, 2, 3, 4, 5, 6},
			ExamRequired:   true,
			ExamType:       "comprehensive",
			ExamDuration:   evaluators.ExamDurationSilver,
			PassingScore:   evaluators.HighQualityScore,
		},
		"Gold": {
			Level:          "Gold",
			MinScore:       evaluators.HighQualityScore,
			RequiredStages: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			ExamRequired:   true,
			ExamType:       "advanced",
			ExamDuration:   evaluators.ExamDurationGold,
			PassingScore:   evaluators.ExcellentScore,
		},
		"Platinum": {
			Level:          "Platinum",
			MinScore:       evaluators.ExcellentScore,
			RequiredStages: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			ExamRequired:   true,
			ExamType:       "expert",
			ExamDuration:   evaluators.ExamDurationPlatinum,
			PassingScore:   evaluators.Score95,
		},
	}
}

// CalculateOverallScore 计算评估综合得分（加权平均法）
//
// 功能说明:
//
//	本方法根据评估框架定义的维度权重，对各维度得分进行加权平均计算，
//	得出最终的综合评分。同时自动计算得分率、判定等级，并更新结果对象。
//
// 计算逻辑:
//  1. 加权求和: 遍历所有维度得分，乘以对应权重后累加
//     公式: totalScore = Σ(维度得分 × 维度权重)
//  2. 归一化: 除以总权重得到加权平均分
//     公式: overallScore = totalScore / totalWeight
//  3. 得分率: 综合分除以满分，转换为百分比
//     公式: percentage = (overallScore / maxScore) × 100
//  4. 等级判定: 根据得分率映射到字母等级（A+/A/A-/B+/B/B-/C+/C/F）
//
// 参数:
//   - framework: 评估框架指针，包含维度定义和权重配置
//
// 返回值:
//   - float64: 计算得出的综合得分（0.0-100.0范围）
//
// 副作用:
//
//	本方法会修改接收者AssessmentResult的以下字段：
//	- OverallScore: 更新为计算后的综合得分
//	- Percentage: 更新为得分率百分比
//	- Grade: 更新为对应的字母等级
//
// 性能优化:
//   - 使用索引遍历避免大结构体复制（AssessmentDimension为128字节）
//   - 提前break减少不必要的遍历
//   - 仅在totalWeight>0时进行除法运算，避免除零
//
// 使用示例:
//
//	result := &AssessmentResult{
//	    DimensionScores: map[string]float64{
//	        "technical": 85.0,
//	        "practice":  78.0,
//	    },
//	    MaxScore: 100.0,
//	}
//	finalScore := result.CalculateOverallScore(framework)
//	// finalScore ≈ 82.2 (假设technical权重0.6, practice权重0.4)
//	// result.Grade = "B+" (82.2%在B+范围)
//
// 注意事项:
//   - 确保framework.Dimensions包含与DimensionScores匹配的维度ID
//   - 如果某维度分数缺失，该维度不参与计算（权重相应减少）
//   - MaxScore应提前设置，否则Percentage计算会错误
//
// 作者: JIA
func (ar *AssessmentResult) CalculateOverallScore(framework *AssessmentFramework) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for dimensionID, score := range ar.DimensionScores {
		// 使用索引遍历避免大结构体复制（128字节）
		for i := range framework.Dimensions {
			dimension := &framework.Dimensions[i]
			if dimension.ID == dimensionID {
				totalScore += score * dimension.Weight
				totalWeight += dimension.Weight
				break
			}
		}
	}

	if totalWeight > 0 {
		ar.OverallScore = totalScore / totalWeight
	}

	ar.Percentage = (ar.OverallScore / ar.MaxScore) * 100
	ar.Grade = calculateGrade(ar.Percentage)

	return ar.OverallScore
}

// calculateGrade 根据得分率计算字母等级（九级评分制）
//
// 功能说明:
//
//	本函数实现标准化的九级字母评分映射，将0-100的百分制得分率转换为
//	直观的字母等级。等级划分参考国际教育评分标准，适合学术和职业认证场景。
//
// 等级划分标准:
//
//	A+ : 95分及以上  - 卓越水平，接近完美
//	A  : 90-94分     - 优秀水平，全面掌握
//	A- : 85-89分     - 优秀偏下，熟练应用
//	B+ : 80-84分     - 良好水平，较好掌握
//	B  : 75-79分     - 良好偏下，基本熟练
//	B- : 70-74分     - 中等偏上，达到要求
//	C+ : 65-69分     - 中等水平，部分掌握
//	C  : 60-64分     - 及格水平，最低达标
//	F  : 60分以下    - 不及格，需要重修
//
// 参数:
//   - percentage: 得分率百分比（0.0-100.0范围）
//
// 返回值:
//   - string: 对应的字母等级（"A+", "A", "A-", ..., "F"）
//
// 设计理念:
//   - 九级划分提供更细致的能力区分度
//   - A+设置在95分体现卓越标准的高要求
//   - 60分及格线符合教育惯例
//   - switch-case结构清晰，易于理解和维护
//
// 使用场景:
//   - CalculateOverallScore中自动判定等级
//   - 生成评估报告时展示等级
//   - 认证系统判定是否达到等级要求
//
// 示例:
//
//	calculateGrade(97.5)  // 返回 "A+"
//	calculateGrade(82.0)  // 返回 "B+"
//	calculateGrade(59.9)  // 返回 "F"
//
// 注意事项:
//   - 边界值采用左闭右开原则（例如90分算A，不算A+）
//   - 负数或超过100的输入未做防护，调用方需保证输入合法
//
// 作者: JIA
func calculateGrade(percentage float64) string {
	switch {
	case percentage >= GradeAPlusThreshold:
		return "A+"
	case percentage >= GradeAThreshold:
		return "A"
	case percentage >= GradeAMinusThreshold:
		return "A-"
	case percentage >= GradeBPlusThreshold:
		return "B+"
	case percentage >= GradeBThreshold:
		return "B"
	case percentage >= GradeBMinusThreshold:
		return "B-"
	case percentage >= GradeCPlusThreshold:
		return "C+"
	case percentage >= GradeCThreshold:
		return "C"
	default:
		return "F"
	}
}

// ToJSON 将评估框架序列化为格式化的JSON字符串
//
// 功能说明:
//
//	本方法将AssessmentFramework对象转换为易读的JSON格式，用于数据持久化、
//	API响应、配置导出等场景。使用缩进格式化，方便人工阅读和版本控制。
//
// 序列化配置:
//   - 缩进格式: 每级缩进2个空格（Go社区标准）
//   - 字段顺序: 保持结构体定义顺序
//   - 空值处理: 空切片序列化为[]，空map序列化为{}
//   - 时间格式: RFC3339格式（如"2025-10-03T14:30:00Z"）
//
// 返回值:
//   - []byte: JSON字节数组，UTF-8编码
//   - error: 序列化错误（通常不会发生，除非包含不可序列化类型）
//
// 使用场景:
//   - 保存评估框架配置到文件
//   - 通过API返回框架定义
//   - 生成人类可读的配置模板
//   - 版本控制中跟踪框架变更
//
// 示例:
//
//	framework := NewAssessmentFramework("1.0.0", "Go评估框架")
//	jsonData, err := framework.ToJSON()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("framework.json", jsonData, 0644)
//
// 性能考量:
//   - 对于大型框架对象，序列化可能耗时较长
//   - 如果频繁调用，考虑缓存结果或使用流式编码
//   - 缩进格式会增加约30%的数据体积
//
// 作者: JIA
func (af *AssessmentFramework) ToJSON() ([]byte, error) {
	return json.MarshalIndent(af, "", "  ")
}

// FromJSON 从JSON数据反序列化为评估框架对象
//
// 功能说明:
//
//	本方法将JSON格式的字节数据解析为AssessmentFramework结构体，用于加载
//	已保存的框架配置、导入外部配置、或处理API请求数据。
//
// 反序列化特性:
//   - 类型安全: 严格按照结构体tag定义解析JSON字段
//   - 容错处理: 缺失字段使用零值，多余字段被忽略
//   - 时间解析: 自动识别RFC3339、Unix时间戳等多种时间格式
//   - 嵌套支持: 正确处理多层嵌套的复杂结构
//
// 参数:
//   - data: JSON格式的字节数组，必须符合AssessmentFramework结构
//
// 返回值:
//   - error: 反序列化错误，可能原因包括：
//   - JSON格式错误（语法错误、引号不匹配等）
//   - 类型不匹配（字符串无法转为数字等）
//   - 数据格式不符合结构体定义
//
// 使用场景:
//   - 从配置文件加载评估框架
//   - 接收API请求中的框架定义
//   - 导入其他系统的评估标准
//   - 恢复备份的框架配置
//
// 示例:
//
//	jsonData, err := os.ReadFile("framework.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	framework := &AssessmentFramework{}
//	if err := framework.FromJSON(jsonData); err != nil {
//	    log.Fatal("解析框架配置失败:", err)
//	}
//	// 现在framework包含了JSON中的所有数据
//
// 错误处理建议:
//   - 加载配置文件前先验证JSON格式（可用json.Valid）
//   - 解析后验证关键字段是否存在（如Version、Dimensions）
//   - 记录详细错误信息，便于排查配置问题
//
// 注意事项:
//   - 本方法会覆盖接收者的所有字段
//   - 解析失败时，接收者状态不确定，应避免使用
//   - 不会验证数据的业务逻辑正确性（如权重总和是否为1）
//
// 作者: JIA
func (af *AssessmentFramework) FromJSON(data []byte) error {
	return json.Unmarshal(data, af)
}
