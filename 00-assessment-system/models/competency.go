/*
=== Go语言学习评估系统 - 技能能力数据模型 ===

本文件定义了完整的技能能力体系和相关数据结构：
1. Go语言技能能力框架定义
2. 技能分类和等级体系
3. 学习进阶路径和依赖关系
4. 能力评估和认证标准
5. 个性化学习路径规划
6. 技能发展轨迹跟踪
7. 能力差距分析和改进建议
*/

package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// CompetencyFramework 技能能力框架
type CompetencyFramework struct {
	ID              string                           `json:"id"`               // 框架唯一标识
	Name            string                           `json:"name"`             // 框架名称
	Version         string                           `json:"version"`          // 版本号
	Description     string                           `json:"description"`      // 框架描述
	CreatedAt       time.Time                        `json:"created_at"`       // 创建时间
	UpdatedAt       time.Time                        `json:"updated_at"`       // 更新时间

	// 能力结构定义
	Categories      []CompetencyCategory             `json:"categories"`       // 能力类别
	SkillMap        SkillMap                         `json:"skill_map"`        // 技能地图
	Progressions    []LearningProgression            `json:"progressions"`     // 学习进阶路径
	Certifications  []CertificationPath              `json:"certifications"`   // 认证路径

	// 评估配置
	AssessmentRules map[string]AssessmentRule        `json:"assessment_rules"` // 评估规则
	LevelDefinitions map[string]LevelDefinition       `json:"level_definitions"` // 等级定义
	Prerequisites   map[string][]string              `json:"prerequisites"`    // 前置技能

	// 元数据
	Language        string                           `json:"language"`         // 适用语言
	Domain          string                           `json:"domain"`           // 应用领域
	Audience        string                           `json:"audience"`         // 目标受众
	Maintainers     []string                         `json:"maintainers"`      // 维护者
}

// CompetencyCategory 能力类别定义
type CompetencyCategory struct {
	ID              string                           `json:"id"`               // 类别唯一标识
	Name            string                           `json:"name"`             // 类别名称
	Description     string                           `json:"description"`      // 类别描述
	Weight          float64                          `json:"weight"`           // 在整体评估中的权重
	Color           string                           `json:"color"`            // 可视化颜色
	Icon            string                           `json:"icon"`             // 图标标识

	// 子技能定义
	Skills          []Skill                          `json:"skills"`           // 技能列表
	Subcategories   []CompetencyCategory             `json:"subcategories"`    // 子类别

	// 评估配置
	AssessmentMethods []AssessmentMethod              `json:"assessment_methods"` // 评估方法
	LevelRequirements map[int]LevelRequirement        `json:"level_requirements"` // 各等级要求
}

// Skill 技能定义
type Skill struct {
	ID              string                           `json:"id"`               // 技能唯一标识
	Name            string                           `json:"name"`             // 技能名称
	Description     string                           `json:"description"`      // 技能描述
	Category        string                           `json:"category"`         // 所属类别
	Difficulty      int                              `json:"difficulty"`       // 难度等级 (1-5)
	Importance      int                              `json:"importance"`       // 重要程度 (1-5)

	// 学习相关
	LearningObjectives []string                       `json:"learning_objectives"` // 学习目标
	Prerequisites   []string                         `json:"prerequisites"`    // 前置技能
	RelatedSkills   []string                         `json:"related_skills"`   // 相关技能
	Applications    []string                         `json:"applications"`     // 应用场景

	// 评估相关
	Assessments     []SkillAssessment                `json:"assessments"`      // 技能评估
	Examples        []SkillExample                   `json:"examples"`         // 技能示例
	Resources       []LearningResource               `json:"resources"`        // 学习资源

	// 进阶相关
	NextSkills      []string                         `json:"next_skills"`      // 后续技能
	MasteryIndicators []MasteryIndicator             `json:"mastery_indicators"` // 掌握指标
	CommonMistakes  []string                         `json:"common_mistakes"`  // 常见错误
	BestPractices   []string                         `json:"best_practices"`   // 最佳实践
}

// SkillAssessment 技能评估定义
type SkillAssessment struct {
	Type            string                           `json:"type"`             // 评估类型: knowledge, application, synthesis
	Method          string                           `json:"method"`           // 评估方法: test, project, code_review
	Description     string                           `json:"description"`      // 评估描述
	Criteria        []AssessmentCriterion            `json:"criteria"`         // 评估标准
	Weight          float64                          `json:"weight"`           // 权重
	TimeRequired    int                              `json:"time_required"`    // 所需时间(分钟)
}

// AssessmentCriterion 评估标准
type AssessmentCriterion struct {
	Name            string                           `json:"name"`             // 标准名称
	Description     string                           `json:"description"`      // 标准描述
	Levels          []CriterionLevel                 `json:"levels"`           // 等级定义
	Weight          float64                          `json:"weight"`           // 权重
	Measurable      bool                             `json:"measurable"`       // 是否可量化
}

// CriterionLevel 标准等级
type CriterionLevel struct {
	Level           int                              `json:"level"`            // 等级 (1-5)
	Name            string                           `json:"name"`             // 等级名称
	Description     string                           `json:"description"`      // 等级描述
	Indicators      []string                         `json:"indicators"`       // 指标
	Examples        []string                         `json:"examples"`         // 示例
}

// SkillExample 技能示例
type SkillExample struct {
	Title           string                           `json:"title"`            // 示例标题
	Description     string                           `json:"description"`      // 示例描述
	Code            string                           `json:"code"`             // 代码示例
	Explanation     string                           `json:"explanation"`      // 解释说明
	Level           int                              `json:"level"`            // 适用等级
	Tags            []string                         `json:"tags"`             // 标签
}

// LearningResource 学习资源
type LearningResource struct {
	Type            string                           `json:"type"`             // 资源类型: doc, video, book, course
	Title           string                           `json:"title"`            // 资源标题
	URL             string                           `json:"url"`              // 资源链接
	Description     string                           `json:"description"`      // 资源描述
	Difficulty      int                              `json:"difficulty"`       // 难度等级
	EstimatedTime   int                              `json:"estimated_time"`   // 预估学习时间(小时)
	Language        string                           `json:"language"`         // 语言
	Free            bool                             `json:"free"`             // 是否免费
	Rating          float64                          `json:"rating"`           // 评分 (1-5)
	Prerequisites   []string                         `json:"prerequisites"`    // 前置要求
}

// MasteryIndicator 掌握指标
type MasteryIndicator struct {
	Name            string                           `json:"name"`             // 指标名称
	Description     string                           `json:"description"`      // 指标描述
	Type            string                           `json:"type"`             // 指标类型: behavioral, performance, knowledge
	Threshold       float64                          `json:"threshold"`        // 掌握阈值
	MeasurementMethod string                         `json:"measurement_method"` // 测量方法
}

// SkillMap 技能地图
type SkillMap struct {
	Categories      map[string]SkillCategory         `json:"categories"`       // 技能分类
	Dependencies    map[string][]string              `json:"dependencies"`     // 技能依赖关系
	Progressions    map[string]SkillProgression      `json:"progressions"`     // 技能进阶路径
	Specializations []SpecializationTrack            `json:"specializations"`  // 专业化轨道
	CareerPaths     []CareerPath                     `json:"career_paths"`     // 职业发展路径
}

// SkillCategory 技能分类
type SkillCategory struct {
	ID              string                           `json:"id"`               // 分类标识
	Name            string                           `json:"name"`             // 分类名称
	Description     string                           `json:"description"`      // 分类描述
	Skills          []string                         `json:"skills"`           // 包含的技能ID
	CoreSkills      []string                         `json:"core_skills"`      // 核心技能
	AdvancedSkills  []string                         `json:"advanced_skills"`  // 高级技能
	EstimatedHours  int                              `json:"estimated_hours"`  // 预估学习时间
	Difficulty      float64                          `json:"difficulty"`       // 平均难度
}

// SkillProgression 技能进阶路径
type SkillProgression struct {
	SkillID         string                           `json:"skill_id"`         // 技能标识
	Levels          []ProgressionLevel               `json:"levels"`           // 进阶等级
	Milestones      []ProgressionMilestone           `json:"milestones"`       // 里程碑
	EstimatedPath   []string                         `json:"estimated_path"`   // 建议学习路径
	AlternativePaths [][]string                      `json:"alternative_paths"` // 替代路径
}

// ProgressionLevel 进阶等级
type ProgressionLevel struct {
	Level           int                              `json:"level"`            // 等级编号
	Name            string                           `json:"name"`             // 等级名称
	Description     string                           `json:"description"`      // 等级描述
	Requirements    []string                         `json:"requirements"`     // 达成要求
	Capabilities    []string                         `json:"capabilities"`     // 能力表现
	Projects        []string                         `json:"projects"`         // 建议项目
	EstimatedHours  int                              `json:"estimated_hours"`  // 预估时间
}

// ProgressionMilestone 进阶里程碑
type ProgressionMilestone struct {
	Name            string                           `json:"name"`             // 里程碑名称
	Description     string                           `json:"description"`      // 里程碑描述
	Requirements    []string                         `json:"requirements"`     // 达成要求
	Evidence        []string                         `json:"evidence"`         // 证明方式
	Rewards         []string                         `json:"rewards"`          // 奖励/认可
}

// SpecializationTrack 专业化轨道
type SpecializationTrack struct {
	ID              string                           `json:"id"`               // 轨道标识
	Name            string                           `json:"name"`             // 轨道名称
	Description     string                           `json:"description"`      // 轨道描述
	Domain          string                           `json:"domain"`           // 应用领域
	RequiredSkills  []string                         `json:"required_skills"`  // 必需技能
	ElectiveSkills  []string                         `json:"elective_skills"`  // 选修技能
	CoreProjects    []string                         `json:"core_projects"`    // 核心项目
	EstimatedDuration int                            `json:"estimated_duration"` // 预估时长(月)
	CareerOutcomes  []string                         `json:"career_outcomes"`  // 职业出路
}

// CareerPath 职业发展路径
type CareerPath struct {
	ID              string                           `json:"id"`               // 路径标识
	Title           string                           `json:"title"`            // 职位标题
	Description     string                           `json:"description"`      // 路径描述
	Industry        string                           `json:"industry"`         // 适用行业
	Roles           []CareerRole                     `json:"roles"`            // 职业角色进阶
	RequiredSkills  map[string]int                   `json:"required_skills"`  // 技能要求(技能ID->等级)
	Timeline        []CareerMilestone                `json:"timeline"`         // 发展时间线
	Salary          SalaryRange                      `json:"salary"`           // 薪资范围
}

// CareerRole 职业角色
type CareerRole struct {
	Level           string                           `json:"level"`            // 级别: junior, mid, senior, lead, architect
	Title           string                           `json:"title"`            // 职位名称
	Responsibilities []string                        `json:"responsibilities"` // 职责
	RequiredSkills  map[string]int                   `json:"required_skills"`  // 技能要求
	Experience      int                              `json:"experience"`       // 经验年限
	Salary          SalaryRange                      `json:"salary"`           // 薪资范围
}

// CareerMilestone 职业里程碑
type CareerMilestone struct {
	Stage           string                           `json:"stage"`            // 阶段名称
	TimeFrame       string                           `json:"time_frame"`       // 时间框架
	Objectives      []string                         `json:"objectives"`       // 目标
	KeySkills       []string                         `json:"key_skills"`       // 关键技能
	Achievements    []string                         `json:"achievements"`     // 成就指标
}

// SalaryRange 薪资范围
type SalaryRange struct {
	Currency        string                           `json:"currency"`         // 货币单位
	MinSalary       int                              `json:"min_salary"`       // 最低薪资
	MaxSalary       int                              `json:"max_salary"`       // 最高薪资
	MedianSalary    int                              `json:"median_salary"`    // 中位薪资
	Location        string                           `json:"location"`         // 地理位置
	LastUpdated     time.Time                        `json:"last_updated"`     // 最后更新时间
}

// LearningProgression 学习进阶路径
type LearningProgression struct {
	ID              string                           `json:"id"`               // 路径标识
	Name            string                           `json:"name"`             // 路径名称
	Description     string                           `json:"description"`      // 路径描述
	TargetAudience  string                           `json:"target_audience"`  // 目标受众
	Prerequisites   []string                         `json:"prerequisites"`    // 前置条件

	// 路径结构
	Phases          []LearningPhase                  `json:"phases"`           // 学习阶段
	TotalDuration   int                              `json:"total_duration"`   // 总时长(小时)
	Difficulty      string                           `json:"difficulty"`       // 难度等级

	// 学习目标
	LearningGoals   []string                         `json:"learning_goals"`   // 学习目标
	SkillsAcquired  []string                         `json:"skills_acquired"`  // 获得技能
	Competencies    []string                         `json:"competencies"`     // 能力获得

	// 评估和认证
	Assessments     []string                         `json:"assessments"`      // 评估项目
	Certifications  []string                         `json:"certifications"`   // 认证机会
	Portfolio       []string                         `json:"portfolio"`        // 作品集要求
}

// LearningPhase 学习阶段
type LearningPhase struct {
	ID              string                           `json:"id"`               // 阶段标识
	Name            string                           `json:"name"`             // 阶段名称
	Description     string                           `json:"description"`      // 阶段描述
	Order           int                              `json:"order"`            // 顺序
	Duration        int                              `json:"duration"`         // 持续时间(小时)

	// 学习内容
	Topics          []LearningTopic                  `json:"topics"`           // 学习主题
	Projects        []string                         `json:"projects"`         // 项目要求
	Exercises       []string                         `json:"exercises"`        // 练习题
	Resources       []LearningResource               `json:"resources"`        // 学习资源

	// 评估
	Assessments     []PhaseAssessment                `json:"assessments"`      // 阶段评估
	Prerequisites   []string                         `json:"prerequisites"`    // 前置要求
	ExitCriteria    []string                         `json:"exit_criteria"`    // 完成标准
}

// LearningTopic 学习主题
type LearningTopic struct {
	ID              string                           `json:"id"`               // 主题标识
	Name            string                           `json:"name"`             // 主题名称
	Description     string                           `json:"description"`      // 主题描述
	Duration        int                              `json:"duration"`         // 预估时间(小时)
	Difficulty      int                              `json:"difficulty"`       // 难度等级

	// 内容组织
	Subtopics       []string                         `json:"subtopics"`        // 子主题
	LearningObjectives []string                      `json:"learning_objectives"` // 学习目标
	KeyConcepts     []string                         `json:"key_concepts"`     // 关键概念
	PracticalSkills []string                         `json:"practical_skills"` // 实用技能

	// 学习材料
	Materials       []LearningMaterial               `json:"materials"`        // 学习材料
	Examples        []TopicExample                   `json:"examples"`         // 主题示例
	Exercises       []TopicExercise                  `json:"exercises"`        // 主题练习
}

// LearningMaterial 学习材料
type LearningMaterial struct {
	Type            string                           `json:"type"`             // 材料类型
	Title           string                           `json:"title"`            // 材料标题
	Content         string                           `json:"content"`          // 材料内容
	URL             string                           `json:"url"`              // 链接地址
	Duration        int                              `json:"duration"`         // 时长
	Interactive     bool                             `json:"interactive"`      // 是否交互式
}

// TopicExample 主题示例
type TopicExample struct {
	Name            string                           `json:"name"`             // 示例名称
	Description     string                           `json:"description"`      // 示例描述
	Code            string                           `json:"code"`             // 示例代码
	Explanation     string                           `json:"explanation"`      // 解释说明
	Difficulty      int                              `json:"difficulty"`       // 难度等级
}

// TopicExercise 主题练习
type TopicExercise struct {
	ID              string                           `json:"id"`               // 练习标识
	Name            string                           `json:"name"`             // 练习名称
	Description     string                           `json:"description"`      // 练习描述
	Instructions    string                           `json:"instructions"`     // 练习说明
	StarterCode     string                           `json:"starter_code"`     // 起始代码
	Solution        string                           `json:"solution"`         // 参考解答
	TestCases       []ExerciseTestCase               `json:"test_cases"`       // 测试用例
	Difficulty      int                              `json:"difficulty"`       // 难度等级
	EstimatedTime   int                              `json:"estimated_time"`   // 预估时间
}

// ExerciseTestCase 练习测试用例
type ExerciseTestCase struct {
	Input           string                           `json:"input"`            // 输入
	ExpectedOutput  string                           `json:"expected_output"`  // 期望输出
	Description     string                           `json:"description"`      // 用例描述
	Hidden          bool                             `json:"hidden"`           // 是否隐藏
}

// PhaseAssessment 阶段评估
type PhaseAssessment struct {
	Type            string                           `json:"type"`             // 评估类型
	Name            string                           `json:"name"`             // 评估名称
	Description     string                           `json:"description"`      // 评估描述
	Weight          float64                          `json:"weight"`           // 权重
	PassingScore    float64                          `json:"passing_score"`    // 及格分数
	TimeLimit       int                              `json:"time_limit"`       // 时间限制
	Attempts        int                              `json:"attempts"`         // 允许尝试次数
}

// CertificationPath 认证路径
type CertificationPath struct {
	ID              string                           `json:"id"`               // 认证标识
	Name            string                           `json:"name"`             // 认证名称
	Description     string                           `json:"description"`      // 认证描述
	Level           string                           `json:"level"`            // 认证等级
	Provider        string                           `json:"provider"`         // 认证机构

	// 认证要求
	Prerequisites   []string                         `json:"prerequisites"`    // 前置条件
	RequiredSkills  map[string]int                   `json:"required_skills"`  // 技能要求
	RequiredProjects []ProjectRequirement            `json:"required_projects"` // 项目要求
	ExamRequirements ExamRequirements                `json:"exam_requirements"` // 考试要求

	// 认证价值
	Industry        []string                         `json:"industry"`         // 适用行业
	Roles           []string                         `json:"roles"`            // 适用职位
	Validity        int                              `json:"validity"`         // 有效期(月)
	Recognition     string                           `json:"recognition"`      // 认可度

	// 路径信息
	EstimatedTime   int                              `json:"estimated_time"`   // 预估准备时间
	Cost            float64                          `json:"cost"`             // 认证费用
	SuccessRate     float64                          `json:"success_rate"`     // 通过率
	RenewalRequired bool                             `json:"renewal_required"` // 是否需要续证
}

// ExamRequirements 考试要求
type ExamRequirements struct {
	Format          string                           `json:"format"`           // 考试形式: online, proctored, hands-on
	Duration        int                              `json:"duration"`         // 考试时长
	PassingScore    float64                          `json:"passing_score"`    // 及格分数
	Sections        []ExamSection                    `json:"sections"`         // 考试部分
	Materials       []string                         `json:"materials"`        // 允许材料
	Retake          RetakePolicy                     `json:"retake"`           // 重考政策
}

// ExamSection 考试部分
type ExamSection struct {
	Name            string                           `json:"name"`             // 部分名称
	Weight          float64                          `json:"weight"`           // 权重
	Topics          []string                         `json:"topics"`           // 涵盖主题
	QuestionTypes   []string                         `json:"question_types"`   // 题目类型
	TimeAllocation  int                              `json:"time_allocation"`  // 时间分配
}

// RetakePolicy 重考政策
type RetakePolicy struct {
	MaxAttempts     int                              `json:"max_attempts"`     // 最大尝试次数
	WaitingPeriod   int                              `json:"waiting_period"`   // 等待期(天)
	AdditionalFee   float64                          `json:"additional_fee"`   // 额外费用
	Prerequisites   []string                         `json:"prerequisites"`    // 重考前置条件
}

// AssessmentMethod 评估方法
type AssessmentMethod struct {
	ID              string                           `json:"id"`               // 方法标识
	Name            string                           `json:"name"`             // 方法名称
	Description     string                           `json:"description"`      // 方法描述
	Type            string                           `json:"type"`             // 方法类型
	Automated       bool                             `json:"automated"`        // 是否自动化
	Weight          float64                          `json:"weight"`           // 权重
	Frequency       string                           `json:"frequency"`        // 评估频率
	Tools           []string                         `json:"tools"`            // 使用工具
	Criteria        []AssessmentCriterion            `json:"criteria"`         // 评估标准
}

// LevelDefinition 等级定义
type LevelDefinition struct {
	Level       int      `json:"level"`        // 等级数值 (1-5)
	Name        string   `json:"name"`         // 等级名称
	Description string   `json:"description"`  // 等级描述
	Evidence    []string `json:"evidence"`     // 所需证据
	Confidence  float64  `json:"confidence"`   // 置信度阈值
}

// AssessmentRule 评估规则
type AssessmentRule struct {
	ID              string                           `json:"id"`               // 规则标识
	Name            string                           `json:"name"`             // 规则名称
	Description     string                           `json:"description"`      // 规则描述
	Condition       string                           `json:"condition"`        // 触发条件
	Action          string                           `json:"action"`           // 执行动作
	Parameters      map[string]interface{}           `json:"parameters"`       // 参数
	Priority        int                              `json:"priority"`         // 优先级
	Enabled         bool                             `json:"enabled"`          // 是否启用
}

// NewCompetencyFramework 创建新的能力框架
func NewCompetencyFramework(name, version string) *CompetencyFramework {
	now := time.Now()
	return &CompetencyFramework{
		ID:              fmt.Sprintf("framework_%s_%s", name, version),
		Name:            name,
		Version:         version,
		CreatedAt:       now,
		UpdatedAt:       now,
		Categories:      []CompetencyCategory{},
		SkillMap:        NewGoSkillMap(),
		Progressions:    []LearningProgression{},
		Certifications:  []CertificationPath{},
		AssessmentRules: make(map[string]AssessmentRule),
		LevelDefinitions: getDefaultLevelDefinitions(),
		Prerequisites:   make(map[string][]string),
		Language:        "Go",
		Domain:          "Software Development",
		Audience:        "从入门到通天",
		Maintainers:     []string{},
	}
}

// NewGoSkillMap 创建Go语言技能地图
func NewGoSkillMap() SkillMap {
	return SkillMap{
		Categories:      getGoSkillCategories(),
		Dependencies:    getSkillDependencies(),
		Progressions:    getSkillProgressions(),
		Specializations: getSpecializationTracks(),
		CareerPaths:     getCareerPaths(),
	}
}

// getGoSkillCategories 获取Go语言技能分类
func getGoSkillCategories() map[string]SkillCategory {
	return map[string]SkillCategory{
		"language_fundamentals": {
			ID:          "language_fundamentals",
			Name:        "语言基础",
			Description: "Go语言核心语法和概念",
			Skills:      []string{"syntax", "types", "functions", "packages"},
			CoreSkills:  []string{"syntax", "types", "functions"},
			AdvancedSkills: []string{"reflection", "unsafe"},
			EstimatedHours: 40,
			Difficulty:  2.0,
		},
		"concurrency": {
			ID:          "concurrency",
			Name:        "并发编程",
			Description: "Go语言并发模型和模式",
			Skills:      []string{"goroutines", "channels", "select", "sync"},
			CoreSkills:  []string{"goroutines", "channels"},
			AdvancedSkills: []string{"advanced_patterns", "performance_tuning"},
			EstimatedHours: 60,
			Difficulty:  4.0,
		},
		"web_development": {
			ID:          "web_development",
			Name:        "Web开发",
			Description: "使用Go进行Web应用开发",
			Skills:      []string{"http", "routing", "middleware", "templates"},
			CoreSkills:  []string{"http", "routing"},
			AdvancedSkills: []string{"microservices", "graphql"},
			EstimatedHours: 80,
			Difficulty:  3.0,
		},
		"system_programming": {
			ID:          "system_programming",
			Name:        "系统编程",
			Description: "系统级编程和工具开发",
			Skills:      []string{"cli", "file_io", "networking", "databases"},
			CoreSkills:  []string{"cli", "file_io"},
			AdvancedSkills: []string{"low_level", "performance"},
			EstimatedHours: 100,
			Difficulty:  4.5,
		},
	}
}

// getSkillDependencies 获取技能依赖关系
func getSkillDependencies() map[string][]string {
	return map[string][]string{
		"functions":    {"syntax", "types"},
		"interfaces":   {"types", "functions"},
		"goroutines":   {"functions"},
		"channels":     {"goroutines"},
		"http":         {"functions", "interfaces"},
		"microservices": {"http", "concurrency"},
	}
}

// getSkillProgressions 获取技能进阶路径
func getSkillProgressions() map[string]SkillProgression {
	return map[string]SkillProgression{
		"concurrency": {
			SkillID: "concurrency",
			Levels: []ProgressionLevel{
				{
					Level:       1,
					Name:        "基础并发",
					Description: "理解goroutines和channels基本概念",
					Requirements: []string{"创建goroutines", "使用unbuffered channels"},
					Capabilities: []string{"简单并发任务", "基础通信"},
					Projects:    []string{"并发计算器", "生产者-消费者模式"},
					EstimatedHours: 20,
				},
				{
					Level:       2,
					Name:        "并发模式",
					Description: "掌握常见并发模式",
					Requirements: []string{"select语句", "buffered channels", "worker pools"},
					Capabilities: []string{"复杂并发控制", "资源池管理"},
					Projects:    []string{"Web爬虫", "任务调度器"},
					EstimatedHours: 25,
				},
				{
					Level:       3,
					Name:        "高级并发",
					Description: "性能优化和高级模式",
					Requirements: []string{"context使用", "sync包", "原子操作"},
					Capabilities: []string{"并发安全设计", "性能调优"},
					Projects:    []string{"高性能服务器", "分布式系统组件"},
					EstimatedHours: 35,
				},
			},
		},
	}
}

// getSpecializationTracks 获取专业化轨道
func getSpecializationTracks() []SpecializationTrack {
	return []SpecializationTrack{
		{
			ID:          "web_backend",
			Name:        "Web后端开发",
			Description: "专注于Web后端服务开发",
			Domain:      "Web Development",
			RequiredSkills: []string{"http", "databases", "apis", "authentication"},
			ElectiveSkills: []string{"caching", "message_queues", "monitoring"},
			CoreProjects: []string{"RESTful API", "微服务架构", "实时系统"},
			EstimatedDuration: 6,
			CareerOutcomes: []string{"后端工程师", "API开发者", "微服务架构师"},
		},
		{
			ID:          "system_tools",
			Name:        "系统工具开发",
			Description: "专注于系统工具和CLI应用开发",
			Domain:      "System Programming",
			RequiredSkills: []string{"cli", "file_io", "system_calls", "cross_platform"},
			ElectiveSkills: []string{"performance_tuning", "memory_management"},
			CoreProjects: []string{"命令行工具", "系统监控", "自动化脚本"},
			EstimatedDuration: 4,
			CareerOutcomes: []string{"DevOps工程师", "系统工程师", "工具开发者"},
		},
	}
}

// getCareerPaths 获取职业发展路径
func getCareerPaths() []CareerPath {
	return []CareerPath{
		{
			ID:          "backend_engineer",
			Title:       "后端工程师",
			Description: "专注于服务端开发和架构设计",
			Industry:    "Software Development",
			Roles: []CareerRole{
				{
					Level:       "junior",
					Title:       "初级后端工程师",
					Responsibilities: []string{"实现基础功能", "编写单元测试", "参与代码审查"},
					RequiredSkills: map[string]int{"http": 3, "databases": 2, "testing": 3},
					Experience:  1,
					Salary:     SalaryRange{Currency: "USD", MinSalary: 60000, MaxSalary: 80000, MedianSalary: 70000},
				},
				{
					Level:       "senior",
					Title:       "高级后端工程师",
					Responsibilities: []string{"架构设计", "技术选型", "团队指导"},
					RequiredSkills: map[string]int{"microservices": 4, "performance": 4, "leadership": 3},
					Experience:  5,
					Salary:     SalaryRange{Currency: "USD", MinSalary: 120000, MaxSalary: 160000, MedianSalary: 140000},
				},
			},
		},
	}
}

// getDefaultLevelDefinitions 获取默认等级定义
func getDefaultLevelDefinitions() map[string]LevelDefinition {
	return map[string]LevelDefinition{
		"novice": {
			Level:       1,
			Name:        "新手",
			Description: "初学者，具备基础理解",
			Evidence:    []string{"完成基础教程", "实现简单功能"},
			Confidence:  0.2,
		},
		"advanced_beginner": {
			Level:       2,
			Name:        "进阶新手",
			Description: "能够在指导下完成任务",
			Evidence:    []string{"完成指导项目", "解决基础问题"},
			Confidence:  0.4,
		},
		"competent": {
			Level:       3,
			Name:        "胜任者",
			Description: "能够独立完成大多数任务",
			Evidence:    []string{"独立项目开发", "代码审查参与"},
			Confidence:  0.6,
		},
		"proficient": {
			Level:       4,
			Name:        "精通者",
			Description: "深度理解，能够优化和改进",
			Evidence:    []string{"性能优化", "架构设计", "技术指导"},
			Confidence:  0.8,
		},
		"expert": {
			Level:       5,
			Name:        "专家",
			Description: "领域专家，能够创新和引领",
			Evidence:    []string{"技术创新", "社区贡献", "标准制定"},
			Confidence:  0.95,
		},
	}
}

// ToJSON 序列化为JSON
func (cf *CompetencyFramework) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cf, "", "  ")
}

// FromJSON 从JSON反序列化
func (cf *CompetencyFramework) FromJSON(data []byte) error {
	return json.Unmarshal(data, cf)
}