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
*/

package models

import (
	"encoding/json"
	"time"
)

// AssessmentFramework 评估框架定义
type AssessmentFramework struct {
	Version         string                           `json:"version"`          // 框架版本
	Name            string                           `json:"name"`             // 框架名称
	Description     string                           `json:"description"`      // 框架描述
	CreatedAt       time.Time                        `json:"created_at"`       // 创建时间
	UpdatedAt       time.Time                        `json:"updated_at"`       // 更新时间

	// 评估维度定义
	Dimensions      []AssessmentDimension            `json:"dimensions"`       // 评估维度
	WeightMatrix    WeightMatrix                     `json:"weight_matrix"`    // 权重矩阵
	ScoringRules    map[string]ScoringRule           `json:"scoring_rules"`    // 评分规则

	// 阈值和标准
	Thresholds      map[string]float64               `json:"thresholds"`       // 各种阈值设定
	Standards       map[string]QualityStandard       `json:"standards"`        // 质量标准
	LevelRequirements map[string]LevelRequirement     `json:"level_requirements"` // 等级要求

	// 配置选项
	AutoAssessment  bool                             `json:"auto_assessment"`  // 是否启用自动评估
	PeerReview      bool                             `json:"peer_review"`      // 是否启用同伴评审
	MentorReview    bool                             `json:"mentor_review"`    // 是否启用导师评审
}

// AssessmentDimension 评估维度定义
type AssessmentDimension struct {
	ID              string                           `json:"id"`               // 维度唯一标识
	Name            string                           `json:"name"`             // 维度名称
	Description     string                           `json:"description"`      // 维度描述
	Weight          float64                          `json:"weight"`           // 权重 (0.0-1.0)
	MaxScore        float64                          `json:"max_score"`        // 最高分
	Subdimensions   []AssessmentSubdimension         `json:"subdimensions"`    // 子维度
	Metrics         []PerformanceMetric              `json:"metrics"`          // 性能指标
	EvaluationMethod string                          `json:"evaluation_method"` // 评估方法
}

// AssessmentSubdimension 评估子维度
type AssessmentSubdimension struct {
	ID              string                           `json:"id"`               // 子维度标识
	Name            string                           `json:"name"`             // 子维度名称
	Description     string                           `json:"description"`      // 详细描述
	Weight          float64                          `json:"weight"`           // 在父维度中的权重
	Criteria        []EvaluationCriteria             `json:"criteria"`         // 评估标准
	AutomatedChecks []AutomatedCheck                 `json:"automated_checks"` // 自动化检查
}

// EvaluationCriteria 评估标准
type EvaluationCriteria struct {
	ID              string                           `json:"id"`               // 标准唯一标识
	Name            string                           `json:"name"`             // 标准名称
	Description     string                           `json:"description"`      // 详细描述
	Type            string                           `json:"type"`             // 标准类型: binary, scale, rubric
	Weight          float64                          `json:"weight"`           // 权重

	// 评分标准定义
	ScaleDefinition []ScalePoint                     `json:"scale_definition"` // 量表定义
	RubricLevels    []RubricLevel                    `json:"rubric_levels"`    // 评分等级
	Examples        []CriteriaExample                `json:"examples"`         // 示例

	// 自动化配置
	Automated       bool                             `json:"automated"`        // 是否可自动评估
	CheckFunction   string                           `json:"check_function"`   // 检查函数名
	Parameters      map[string]interface{}           `json:"parameters"`       // 检查参数
}

// ScalePoint 量表点定义
type ScalePoint struct {
	Value       float64 `json:"value"`       // 分值
	Label       string  `json:"label"`       // 标签
	Description string  `json:"description"` // 描述
}

// RubricLevel 评分等级定义
type RubricLevel struct {
	Level       int     `json:"level"`       // 等级 (1-5)
	Name        string  `json:"name"`        // 等级名称
	Description string  `json:"description"` // 等级描述
	Score       float64 `json:"score"`       // 对应分数
	Indicators  []string `json:"indicators"` // 指标描述
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
	ID          string                           `json:"id"`           // 检查唯一标识
	Name        string                           `json:"name"`         // 检查名称
	Type        string                           `json:"type"`         // 检查类型: static, dynamic, test
	Tool        string                           `json:"tool"`         // 使用工具: golint, govet, gocyclo, etc.
	Command     string                           `json:"command"`      // 执行命令
	Parameters  map[string]interface{}           `json:"parameters"`   // 参数配置
	Weight      float64                          `json:"weight"`       // 在子维度中的权重

	// 结果处理
	ResultParser string                           `json:"result_parser"` // 结果解析器
	ScoreMapping map[string]float64               `json:"score_mapping"` // 分数映射
	Thresholds   map[string]float64               `json:"thresholds"`    // 阈值设定
}

// PerformanceMetric 性能指标定义
type PerformanceMetric struct {
	ID          string                           `json:"id"`           // 指标唯一标识
	Name        string                           `json:"name"`         // 指标名称
	Unit        string                           `json:"unit"`         // 计量单位
	Target      float64                          `json:"target"`       // 目标值
	Threshold   float64                          `json:"threshold"`    // 阈值
	Weight      float64                          `json:"weight"`       // 权重
	Aggregation string                           `json:"aggregation"`  // 聚合方式: avg, max, min, sum
}

// ScoringRule 评分规则定义
type ScoringRule struct {
	ID          string                           `json:"id"`           // 规则唯一标识
	Name        string                           `json:"name"`         // 规则名称
	Type        string                           `json:"type"`         // 规则类型: weighted_sum, formula, composite
	Formula     string                           `json:"formula"`      // 计算公式
	Parameters  map[string]interface{}           `json:"parameters"`   // 规则参数
	Conditions  []ScoringCondition               `json:"conditions"`   // 条件设定
}

// ScoringCondition 评分条件
type ScoringCondition struct {
	Field       string      `json:"field"`       // 字段名
	Operator    string      `json:"operator"`    // 操作符: >, <, ==, !=, >=, <=
	Value       interface{} `json:"value"`       // 比较值
	Score       float64     `json:"score"`       // 满足条件时的分数
	Modifier    float64     `json:"modifier"`    // 分数修饰符
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
	StageWeights        map[int]StageWeight `json:"stage_weights"` // 各阶段权重调整
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
	Category    string                           `json:"category"`     // 标准类别
	Name        string                           `json:"name"`         // 标准名称
	Description string                           `json:"description"`  // 标准描述
	Levels      map[string]QualityLevel          `json:"levels"`       // 各质量等级
	Metrics     []QualityMetric                  `json:"metrics"`      // 质量指标
}

// QualityLevel 质量等级定义
type QualityLevel struct {
	Level       string                           `json:"level"`        // 等级名称
	Score       float64                          `json:"score"`        // 等级分数
	Description string                           `json:"description"`  // 等级描述
	Requirements []string                         `json:"requirements"` // 要求列表
	Examples    []string                         `json:"examples"`     // 示例
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
	Level           string                           `json:"level"`            // 等级名称
	MinScore        float64                          `json:"min_score"`        // 最低分数要求
	RequiredStages  []int                            `json:"required_stages"`  // 必须完成的阶段
	ProjectRequirements []ProjectRequirement          `json:"project_requirements"` // 项目要求

	// 技能要求
	TechnicalSkills map[string]int                   `json:"technical_skills"` // 技术技能要求 (技能名 -> 等级)
	SoftSkills      map[string]int                   `json:"soft_skills"`      // 软技能要求

	// 考试要求
	ExamRequired    bool                             `json:"exam_required"`    // 是否需要考试
	ExamType        string                           `json:"exam_type"`        // 考试类型
	ExamDuration    int                              `json:"exam_duration"`    // 考试时长(分钟)
	PassingScore    float64                          `json:"passing_score"`    // 及格分数
}

// ProjectRequirement 项目要求
type ProjectRequirement struct {
	Type            string   `json:"type"`             // 项目类型
	MinComplexity   int      `json:"min_complexity"`   // 最低复杂度
	RequiredFeatures []string `json:"required_features"` // 必需功能
	TechStack       []string `json:"tech_stack"`       // 技术栈要求
	MinScore        float64  `json:"min_score"`        // 最低项目评分
}

// AssessmentTask 评估任务定义
type AssessmentTask struct {
	ID              string                           `json:"id"`               // 任务唯一标识
	Name            string                           `json:"name"`             // 任务名称
	Description     string                           `json:"description"`      // 任务描述
	Type            string                           `json:"type"`             // 任务类型: coding, design, analysis, presentation
	Stage           int                              `json:"stage"`            // 适用学习阶段
	Difficulty      int                              `json:"difficulty"`       // 难度等级 (1-5)
	EstimatedTime   int                              `json:"estimated_time"`   // 预估完成时间(分钟)

	// 任务内容
	Instructions    string                           `json:"instructions"`     // 详细说明
	Requirements    []string                         `json:"requirements"`     // 具体要求
	Resources       []TaskResource                   `json:"resources"`        // 参考资源
	StartingCode    string                           `json:"starting_code"`    // 起始代码
	TestCases       []TestCase                       `json:"test_cases"`       // 测试用例

	// 评估配置
	EvaluationCriteria []EvaluationCriteria           `json:"evaluation_criteria"` // 评估标准
	AutoGrading     bool                             `json:"auto_grading"`     // 自动评分
	TimeLimit       *int                             `json:"time_limit"`       // 时间限制
	MaxAttempts     *int                             `json:"max_attempts"`     // 最大尝试次数

	// 元数据
	CreatedBy       string                           `json:"created_by"`       // 创建者
	CreatedAt       time.Time                        `json:"created_at"`       // 创建时间
	UpdatedAt       time.Time                        `json:"updated_at"`       // 更新时间
	Tags            []string                         `json:"tags"`             // 标签
	Keywords        []string                         `json:"keywords"`         // 关键词
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
	ID              string                           `json:"id"`               // 会话唯一标识
	StudentID       string                           `json:"student_id"`       // 学习者标识
	TaskID          string                           `json:"task_id"`          // 任务标识
	Type            string                           `json:"type"`             // 评估类型
	Status          string                           `json:"status"`           // 状态: started, in_progress, completed, expired

	// 时间记录
	StartTime       time.Time                        `json:"start_time"`       // 开始时间
	EndTime         *time.Time                       `json:"end_time"`         // 结束时间
	Duration        int                              `json:"duration"`         // 实际用时(分钟)
	TimeRemaining   *int                             `json:"time_remaining"`   // 剩余时间

	// 提交内容
	Submissions     []TaskSubmission                 `json:"submissions"`      // 提交记录
	FinalSubmission *TaskSubmission                  `json:"final_submission"` // 最终提交

	// 评估结果
	Results         *AssessmentResult                `json:"results"`          // 评估结果
	Feedback        []AssessmentFeedback             `json:"feedback"`         // 反馈信息

	// 元数据
	AttemptNumber   int                              `json:"attempt_number"`   // 尝试次数
	IPAddress       string                           `json:"ip_address"`       // IP地址
	UserAgent       string                           `json:"user_agent"`       // 用户代理
	Environment     map[string]string                `json:"environment"`      // 环境信息
}

// TaskSubmission 任务提交
type TaskSubmission struct {
	ID          string                           `json:"id"`           // 提交标识
	Timestamp   time.Time                        `json:"timestamp"`    // 提交时间
	Code        string                           `json:"code"`         // 提交代码
	Files       map[string]string                `json:"files"`        // 提交文件
	Output      string                           `json:"output"`       // 运行输出
	TestResults []TestResult                     `json:"test_results"` // 测试结果
	Metadata    map[string]interface{}           `json:"metadata"`     // 元数据
}

// TestResult 测试结果
type TestResult struct {
	TestCaseID  string      `json:"test_case_id"`  // 测试用例标识
	Passed      bool        `json:"passed"`        // 是否通过
	ActualOutput interface{} `json:"actual_output"` // 实际输出
	ExecutionTime float64    `json:"execution_time"` // 执行时间
	Error       *string     `json:"error"`         // 错误信息
	Score       float64     `json:"score"`         // 得分
}

// AssessmentResult 评估结果
type AssessmentResult struct {
	SessionID       string                           `json:"session_id"`       // 会话标识
	OverallScore    float64                          `json:"overall_score"`    // 总分
	MaxScore        float64                          `json:"max_score"`        // 满分
	Percentage      float64                          `json:"percentage"`       // 得分率
	Grade           string                           `json:"grade"`            // 等级

	// 维度得分
	DimensionScores map[string]float64               `json:"dimension_scores"` // 各维度得分
	CriteriaScores  map[string]float64               `json:"criteria_scores"`  // 各标准得分

	// 详细结果
	TestResults     []TestResult                     `json:"test_results"`     // 测试结果
	CodeAnalysis    *CodeAnalysisResult              `json:"code_analysis"`    // 代码分析
	PerformanceMetrics map[string]float64            `json:"performance_metrics"` // 性能指标

	// 统计信息
	CompletionTime  int                              `json:"completion_time"`  // 完成用时
	AttemptCount    int                              `json:"attempt_count"`    // 尝试次数
	HintUsed        int                              `json:"hint_used"`        // 使用提示次数
}

// CodeAnalysisResult 代码分析结果
type CodeAnalysisResult struct {
	LinesOfCode     int                              `json:"lines_of_code"`     // 代码行数
	CyclomaticComplexity int                         `json:"cyclomatic_complexity"` // 圈复杂度
	TestCoverage    float64                          `json:"test_coverage"`     // 测试覆盖率
	CodeQuality     float64                          `json:"code_quality"`      // 代码质量分
	SecurityScore   float64                          `json:"security_score"`    // 安全评分
	PerformanceScore float64                         `json:"performance_score"` // 性能评分

	// 详细分析
	Issues          []CodeIssue                      `json:"issues"`            // 代码问题
	Suggestions     []CodeSuggestion                 `json:"suggestions"`       // 改进建议
	Patterns        []DesignPattern                  `json:"patterns"`          // 设计模式使用
}

// CodeIssue 代码问题
type CodeIssue struct {
	Type        string `json:"type"`        // 问题类型
	Severity    string `json:"severity"`    // 严重程度
	Line        int    `json:"line"`        // 行号
	Column      int    `json:"column"`      // 列号
	Message     string `json:"message"`     // 问题描述
	Rule        string `json:"rule"`        // 规则名称
	Suggestion  string `json:"suggestion"`  // 修复建议
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
	Name        string `json:"name"`        // 模式名称
	Usage       string `json:"usage"`       // 使用情况
	Appropriate bool   `json:"appropriate"` // 是否恰当使用
	Score       float64 `json:"score"`      // 使用评分
}

// AssessmentFeedback 评估反馈
type AssessmentFeedback struct {
	Type        string    `json:"type"`        // 反馈类型: automated, peer, mentor
	Source      string    `json:"source"`      // 反馈来源
	Timestamp   time.Time `json:"timestamp"`   // 反馈时间
	Content     string    `json:"content"`     // 反馈内容
	Rating      *float64  `json:"rating"`      // 评分(可选)
	Helpful     *bool     `json:"helpful"`     // 是否有帮助(可选)
	Category    string    `json:"category"`    // 反馈分类
	Tags        []string  `json:"tags"`        // 标签
}

// NewAssessmentFramework 创建新的评估框架
func NewAssessmentFramework(version, name string) *AssessmentFramework {
	now := time.Now()
	return &AssessmentFramework{
		Version:           version,
		Name:             name,
		CreatedAt:        now,
		UpdatedAt:        now,
		Dimensions:       []AssessmentDimension{},
		WeightMatrix:     getDefaultWeightMatrix(),
		ScoringRules:     make(map[string]ScoringRule),
		Thresholds:       getDefaultThresholds(),
		Standards:        make(map[string]QualityStandard),
		LevelRequirements: getDefaultLevelRequirements(),
		AutoAssessment:   true,
		PeerReview:       false,
		MentorReview:     true,
	}
}

// getDefaultWeightMatrix 获取默认权重矩阵
func getDefaultWeightMatrix() WeightMatrix {
	return WeightMatrix{
		TechnicalDepth:      0.40,
		EngineeringPractice: 0.30,
		ProjectExperience:   0.20,
		SoftSkills:         0.10,

		AutomatedAssessment: 0.50,
		CodeReview:         0.30,
		ProjectEvaluation:  0.15,
		PeerFeedback:       0.03,
		MentorAssessment:   0.02,

		StageWeights:       make(map[int]StageWeight),
	}
}

// getDefaultThresholds 获取默认阈值设定
func getDefaultThresholds() map[string]float64 {
	return map[string]float64{
		"passing_score":      70.0,
		"excellent_score":    90.0,
		"min_coverage":       80.0,
		"max_complexity":     10.0,
		"min_documentation":  85.0,
		"performance_target": 95.0,
	}
}

// getDefaultLevelRequirements 获取默认等级要求
func getDefaultLevelRequirements() map[string]LevelRequirement {
	return map[string]LevelRequirement{
		"Bronze": {
			Level:           "Bronze",
			MinScore:        70.0,
			RequiredStages:  []int{1, 2, 3},
			ExamRequired:    true,
			ExamType:        "practical",
			ExamDuration:    120,
			PassingScore:    80.0,
		},
		"Silver": {
			Level:           "Silver",
			MinScore:        80.0,
			RequiredStages:  []int{1, 2, 3, 4, 5, 6},
			ExamRequired:    true,
			ExamType:        "comprehensive",
			ExamDuration:    180,
			PassingScore:    85.0,
		},
		"Gold": {
			Level:           "Gold",
			MinScore:        85.0,
			RequiredStages:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			ExamRequired:    true,
			ExamType:        "advanced",
			ExamDuration:    240,
			PassingScore:    90.0,
		},
		"Platinum": {
			Level:           "Platinum",
			MinScore:        90.0,
			RequiredStages:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			ExamRequired:    true,
			ExamType:        "expert",
			ExamDuration:    300,
			PassingScore:    95.0,
		},
	}
}

// CalculateOverallScore 计算综合得分
func (ar *AssessmentResult) CalculateOverallScore(framework *AssessmentFramework) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for dimensionID, score := range ar.DimensionScores {
		for _, dimension := range framework.Dimensions {
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

// calculateGrade 根据得分率计算等级
func calculateGrade(percentage float64) string {
	switch {
	case percentage >= 95:
		return "A+"
	case percentage >= 90:
		return "A"
	case percentage >= 85:
		return "A-"
	case percentage >= 80:
		return "B+"
	case percentage >= 75:
		return "B"
	case percentage >= 70:
		return "B-"
	case percentage >= 65:
		return "C+"
	case percentage >= 60:
		return "C"
	default:
		return "F"
	}
}

// ToJSON 序列化为JSON
func (af *AssessmentFramework) ToJSON() ([]byte, error) {
	return json.MarshalIndent(af, "", "  ")
}

// FromJSON 从JSON反序列化
func (af *AssessmentFramework) FromJSON(data []byte) error {
	return json.Unmarshal(data, af)
}