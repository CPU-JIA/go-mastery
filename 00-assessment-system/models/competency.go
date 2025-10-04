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
	"assessment-system/evaluators"
	"encoding/json"
	"fmt"
	"time"
)

// CompetencyFramework 技能能力框架
type CompetencyFramework struct {
	ID          string    `json:"id"`          // 框架唯一标识
	Name        string    `json:"name"`        // 框架名称
	Version     string    `json:"version"`     // 版本号
	Description string    `json:"description"` // 框架描述
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
	UpdatedAt   time.Time `json:"updated_at"`  // 更新时间

	// 能力结构定义
	Categories     []CompetencyCategory  `json:"categories"`     // 能力类别
	SkillMap       SkillMap              `json:"skill_map"`      // 技能地图
	Progressions   []LearningProgression `json:"progressions"`   // 学习进阶路径
	Certifications []CertificationPath   `json:"certifications"` // 认证路径

	// 评估配置
	AssessmentRules  map[string]AssessmentRule  `json:"assessment_rules"`  // 评估规则
	LevelDefinitions map[string]LevelDefinition `json:"level_definitions"` // 等级定义
	Prerequisites    map[string][]string        `json:"prerequisites"`     // 前置技能

	// 元数据
	Language    string   `json:"language"`    // 适用语言
	Domain      string   `json:"domain"`      // 应用领域
	Audience    string   `json:"audience"`    // 目标受众
	Maintainers []string `json:"maintainers"` // 维护者
}

// CompetencyCategory 能力类别定义
type CompetencyCategory struct {
	ID          string  `json:"id"`          // 类别唯一标识
	Name        string  `json:"name"`        // 类别名称
	Description string  `json:"description"` // 类别描述
	Weight      float64 `json:"weight"`      // 在整体评估中的权重
	Color       string  `json:"color"`       // 可视化颜色
	Icon        string  `json:"icon"`        // 图标标识

	// 子技能定义
	Skills        []Skill              `json:"skills"`        // 技能列表
	Subcategories []CompetencyCategory `json:"subcategories"` // 子类别

	// 评估配置
	AssessmentMethods []AssessmentMethod       `json:"assessment_methods"` // 评估方法
	LevelRequirements map[int]LevelRequirement `json:"level_requirements"` // 各等级要求
}

// Skill 技能定义
type Skill struct {
	ID          string `json:"id"`          // 技能唯一标识
	Name        string `json:"name"`        // 技能名称
	Description string `json:"description"` // 技能描述
	Category    string `json:"category"`    // 所属类别
	Difficulty  int    `json:"difficulty"`  // 难度等级 (1-5)
	Importance  int    `json:"importance"`  // 重要程度 (1-5)

	// 学习相关
	LearningObjectives []string `json:"learning_objectives"` // 学习目标
	Prerequisites      []string `json:"prerequisites"`       // 前置技能
	RelatedSkills      []string `json:"related_skills"`      // 相关技能
	Applications       []string `json:"applications"`        // 应用场景

	// 评估相关
	Assessments []SkillAssessment  `json:"assessments"` // 技能评估
	Examples    []SkillExample     `json:"examples"`    // 技能示例
	Resources   []LearningResource `json:"resources"`   // 学习资源

	// 进阶相关
	NextSkills        []string           `json:"next_skills"`        // 后续技能
	MasteryIndicators []MasteryIndicator `json:"mastery_indicators"` // 掌握指标
	CommonMistakes    []string           `json:"common_mistakes"`    // 常见错误
	BestPractices     []string           `json:"best_practices"`     // 最佳实践
}

// SkillAssessment 技能评估定义
type SkillAssessment struct {
	Type         string                `json:"type"`          // 评估类型: knowledge, application, synthesis
	Method       string                `json:"method"`        // 评估方法: test, project, code_review
	Description  string                `json:"description"`   // 评估描述
	Criteria     []AssessmentCriterion `json:"criteria"`      // 评估标准
	Weight       float64               `json:"weight"`        // 权重
	TimeRequired int                   `json:"time_required"` // 所需时间(分钟)
}

// AssessmentCriterion 评估标准
type AssessmentCriterion struct {
	Name        string           `json:"name"`        // 标准名称
	Description string           `json:"description"` // 标准描述
	Levels      []CriterionLevel `json:"levels"`      // 等级定义
	Weight      float64          `json:"weight"`      // 权重
	Measurable  bool             `json:"measurable"`  // 是否可量化
}

// CriterionLevel 标准等级
type CriterionLevel struct {
	Level       int      `json:"level"`       // 等级 (1-5)
	Name        string   `json:"name"`        // 等级名称
	Description string   `json:"description"` // 等级描述
	Indicators  []string `json:"indicators"`  // 指标
	Examples    []string `json:"examples"`    // 示例
}

// SkillExample 技能示例
type SkillExample struct {
	Title       string   `json:"title"`       // 示例标题
	Description string   `json:"description"` // 示例描述
	Code        string   `json:"code"`        // 代码示例
	Explanation string   `json:"explanation"` // 解释说明
	Level       int      `json:"level"`       // 适用等级
	Tags        []string `json:"tags"`        // 标签
}

// LearningResource 学习资源
type LearningResource struct {
	Type          string   `json:"type"`           // 资源类型: doc, video, book, course
	Title         string   `json:"title"`          // 资源标题
	URL           string   `json:"url"`            // 资源链接
	Description   string   `json:"description"`    // 资源描述
	Difficulty    int      `json:"difficulty"`     // 难度等级
	EstimatedTime int      `json:"estimated_time"` // 预估学习时间(小时)
	Language      string   `json:"language"`       // 语言
	Free          bool     `json:"free"`           // 是否免费
	Rating        float64  `json:"rating"`         // 评分 (1-5)
	Prerequisites []string `json:"prerequisites"`  // 前置要求
}

// MasteryIndicator 掌握指标
type MasteryIndicator struct {
	Name              string  `json:"name"`               // 指标名称
	Description       string  `json:"description"`        // 指标描述
	Type              string  `json:"type"`               // 指标类型: behavioral, performance, knowledge
	Threshold         float64 `json:"threshold"`          // 掌握阈值
	MeasurementMethod string  `json:"measurement_method"` // 测量方法
}

// SkillMap 技能地图
type SkillMap struct {
	Categories      map[string]SkillCategory    `json:"categories"`      // 技能分类
	Dependencies    map[string][]string         `json:"dependencies"`    // 技能依赖关系
	Progressions    map[string]SkillProgression `json:"progressions"`    // 技能进阶路径
	Specializations []SpecializationTrack       `json:"specializations"` // 专业化轨道
	CareerPaths     []CareerPath                `json:"career_paths"`    // 职业发展路径
}

// SkillCategory 技能分类
type SkillCategory struct {
	ID             string   `json:"id"`              // 分类标识
	Name           string   `json:"name"`            // 分类名称
	Description    string   `json:"description"`     // 分类描述
	Skills         []string `json:"skills"`          // 包含的技能ID
	CoreSkills     []string `json:"core_skills"`     // 核心技能
	AdvancedSkills []string `json:"advanced_skills"` // 高级技能
	EstimatedHours int      `json:"estimated_hours"` // 预估学习时间
	Difficulty     float64  `json:"difficulty"`      // 平均难度
}

// SkillProgression 技能进阶路径
type SkillProgression struct {
	SkillID          string                 `json:"skill_id"`          // 技能标识
	Levels           []ProgressionLevel     `json:"levels"`            // 进阶等级
	Milestones       []ProgressionMilestone `json:"milestones"`        // 里程碑
	EstimatedPath    []string               `json:"estimated_path"`    // 建议学习路径
	AlternativePaths [][]string             `json:"alternative_paths"` // 替代路径
}

// ProgressionLevel 进阶等级
type ProgressionLevel struct {
	Level          int      `json:"level"`           // 等级编号
	Name           string   `json:"name"`            // 等级名称
	Description    string   `json:"description"`     // 等级描述
	Requirements   []string `json:"requirements"`    // 达成要求
	Capabilities   []string `json:"capabilities"`    // 能力表现
	Projects       []string `json:"projects"`        // 建议项目
	EstimatedHours int      `json:"estimated_hours"` // 预估时间
}

// ProgressionMilestone 进阶里程碑
type ProgressionMilestone struct {
	Name         string   `json:"name"`         // 里程碑名称
	Description  string   `json:"description"`  // 里程碑描述
	Requirements []string `json:"requirements"` // 达成要求
	Evidence     []string `json:"evidence"`     // 证明方式
	Rewards      []string `json:"rewards"`      // 奖励/认可
}

// SpecializationTrack 专业化轨道
type SpecializationTrack struct {
	ID                string   `json:"id"`                 // 轨道标识
	Name              string   `json:"name"`               // 轨道名称
	Description       string   `json:"description"`        // 轨道描述
	Domain            string   `json:"domain"`             // 应用领域
	RequiredSkills    []string `json:"required_skills"`    // 必需技能
	ElectiveSkills    []string `json:"elective_skills"`    // 选修技能
	CoreProjects      []string `json:"core_projects"`      // 核心项目
	EstimatedDuration int      `json:"estimated_duration"` // 预估时长(月)
	CareerOutcomes    []string `json:"career_outcomes"`    // 职业出路
}

// CareerPath 职业发展路径
type CareerPath struct {
	ID             string            `json:"id"`              // 路径标识
	Title          string            `json:"title"`           // 职位标题
	Description    string            `json:"description"`     // 路径描述
	Industry       string            `json:"industry"`        // 适用行业
	Roles          []CareerRole      `json:"roles"`           // 职业角色进阶
	RequiredSkills map[string]int    `json:"required_skills"` // 技能要求(技能ID->等级)
	Timeline       []CareerMilestone `json:"timeline"`        // 发展时间线
	Salary         SalaryRange       `json:"salary"`          // 薪资范围
}

// CareerRole 职业角色
type CareerRole struct {
	Level            string         `json:"level"`            // 级别: junior, mid, senior, lead, architect
	Title            string         `json:"title"`            // 职位名称
	Responsibilities []string       `json:"responsibilities"` // 职责
	RequiredSkills   map[string]int `json:"required_skills"`  // 技能要求
	Experience       int            `json:"experience"`       // 经验年限
	Salary           SalaryRange    `json:"salary"`           // 薪资范围
}

// CareerMilestone 职业里程碑
type CareerMilestone struct {
	Stage        string   `json:"stage"`        // 阶段名称
	TimeFrame    string   `json:"time_frame"`   // 时间框架
	Objectives   []string `json:"objectives"`   // 目标
	KeySkills    []string `json:"key_skills"`   // 关键技能
	Achievements []string `json:"achievements"` // 成就指标
}

// SalaryRange 薪资范围
type SalaryRange struct {
	Currency     string    `json:"currency"`      // 货币单位
	MinSalary    int       `json:"min_salary"`    // 最低薪资
	MaxSalary    int       `json:"max_salary"`    // 最高薪资
	MedianSalary int       `json:"median_salary"` // 中位薪资
	Location     string    `json:"location"`      // 地理位置
	LastUpdated  time.Time `json:"last_updated"`  // 最后更新时间
}

// LearningProgression 学习进阶路径
type LearningProgression struct {
	ID             string   `json:"id"`              // 路径标识
	Name           string   `json:"name"`            // 路径名称
	Description    string   `json:"description"`     // 路径描述
	TargetAudience string   `json:"target_audience"` // 目标受众
	Prerequisites  []string `json:"prerequisites"`   // 前置条件

	// 路径结构
	Phases        []LearningPhase `json:"phases"`         // 学习阶段
	TotalDuration int             `json:"total_duration"` // 总时长(小时)
	Difficulty    string          `json:"difficulty"`     // 难度等级

	// 学习目标
	LearningGoals  []string `json:"learning_goals"`  // 学习目标
	SkillsAcquired []string `json:"skills_acquired"` // 获得技能
	Competencies   []string `json:"competencies"`    // 能力获得

	// 评估和认证
	Assessments    []string `json:"assessments"`    // 评估项目
	Certifications []string `json:"certifications"` // 认证机会
	Portfolio      []string `json:"portfolio"`      // 作品集要求
}

// LearningPhase 学习阶段
type LearningPhase struct {
	ID          string `json:"id"`          // 阶段标识
	Name        string `json:"name"`        // 阶段名称
	Description string `json:"description"` // 阶段描述
	Order       int    `json:"order"`       // 顺序
	Duration    int    `json:"duration"`    // 持续时间(小时)

	// 学习内容
	Topics    []LearningTopic    `json:"topics"`    // 学习主题
	Projects  []string           `json:"projects"`  // 项目要求
	Exercises []string           `json:"exercises"` // 练习题
	Resources []LearningResource `json:"resources"` // 学习资源

	// 评估
	Assessments   []PhaseAssessment `json:"assessments"`   // 阶段评估
	Prerequisites []string          `json:"prerequisites"` // 前置要求
	ExitCriteria  []string          `json:"exit_criteria"` // 完成标准
}

// LearningTopic 学习主题
type LearningTopic struct {
	ID          string `json:"id"`          // 主题标识
	Name        string `json:"name"`        // 主题名称
	Description string `json:"description"` // 主题描述
	Duration    int    `json:"duration"`    // 预估时间(小时)
	Difficulty  int    `json:"difficulty"`  // 难度等级

	// 内容组织
	Subtopics          []string `json:"subtopics"`           // 子主题
	LearningObjectives []string `json:"learning_objectives"` // 学习目标
	KeyConcepts        []string `json:"key_concepts"`        // 关键概念
	PracticalSkills    []string `json:"practical_skills"`    // 实用技能

	// 学习材料
	Materials []LearningMaterial `json:"materials"` // 学习材料
	Examples  []TopicExample     `json:"examples"`  // 主题示例
	Exercises []TopicExercise    `json:"exercises"` // 主题练习
}

// LearningMaterial 学习材料
type LearningMaterial struct {
	Type        string `json:"type"`        // 材料类型
	Title       string `json:"title"`       // 材料标题
	Content     string `json:"content"`     // 材料内容
	URL         string `json:"url"`         // 链接地址
	Duration    int    `json:"duration"`    // 时长
	Interactive bool   `json:"interactive"` // 是否交互式
}

// TopicExample 主题示例
type TopicExample struct {
	Name        string `json:"name"`        // 示例名称
	Description string `json:"description"` // 示例描述
	Code        string `json:"code"`        // 示例代码
	Explanation string `json:"explanation"` // 解释说明
	Difficulty  int    `json:"difficulty"`  // 难度等级
}

// TopicExercise 主题练习
type TopicExercise struct {
	ID            string             `json:"id"`             // 练习标识
	Name          string             `json:"name"`           // 练习名称
	Description   string             `json:"description"`    // 练习描述
	Instructions  string             `json:"instructions"`   // 练习说明
	StarterCode   string             `json:"starter_code"`   // 起始代码
	Solution      string             `json:"solution"`       // 参考解答
	TestCases     []ExerciseTestCase `json:"test_cases"`     // 测试用例
	Difficulty    int                `json:"difficulty"`     // 难度等级
	EstimatedTime int                `json:"estimated_time"` // 预估时间
}

// ExerciseTestCase 练习测试用例
type ExerciseTestCase struct {
	Input          string `json:"input"`           // 输入
	ExpectedOutput string `json:"expected_output"` // 期望输出
	Description    string `json:"description"`     // 用例描述
	Hidden         bool   `json:"hidden"`          // 是否隐藏
}

// PhaseAssessment 阶段评估
type PhaseAssessment struct {
	Type         string  `json:"type"`          // 评估类型
	Name         string  `json:"name"`          // 评估名称
	Description  string  `json:"description"`   // 评估描述
	Weight       float64 `json:"weight"`        // 权重
	PassingScore float64 `json:"passing_score"` // 及格分数
	TimeLimit    int     `json:"time_limit"`    // 时间限制
	Attempts     int     `json:"attempts"`      // 允许尝试次数
}

// CertificationPath 认证路径
type CertificationPath struct {
	ID          string `json:"id"`          // 认证标识
	Name        string `json:"name"`        // 认证名称
	Description string `json:"description"` // 认证描述
	Level       string `json:"level"`       // 认证等级
	Provider    string `json:"provider"`    // 认证机构

	// 认证要求
	Prerequisites    []string             `json:"prerequisites"`     // 前置条件
	RequiredSkills   map[string]int       `json:"required_skills"`   // 技能要求
	RequiredProjects []ProjectRequirement `json:"required_projects"` // 项目要求
	ExamRequirements ExamRequirements     `json:"exam_requirements"` // 考试要求

	// 认证价值
	Industry    []string `json:"industry"`    // 适用行业
	Roles       []string `json:"roles"`       // 适用职位
	Validity    int      `json:"validity"`    // 有效期(月)
	Recognition string   `json:"recognition"` // 认可度

	// 路径信息
	EstimatedTime   int     `json:"estimated_time"`   // 预估准备时间
	Cost            float64 `json:"cost"`             // 认证费用
	SuccessRate     float64 `json:"success_rate"`     // 通过率
	RenewalRequired bool    `json:"renewal_required"` // 是否需要续证
}

// ExamRequirements 考试要求
type ExamRequirements struct {
	Format       string        `json:"format"`        // 考试形式: online, proctored, hands-on
	Duration     int           `json:"duration"`      // 考试时长
	PassingScore float64       `json:"passing_score"` // 及格分数
	Sections     []ExamSection `json:"sections"`      // 考试部分
	Materials    []string      `json:"materials"`     // 允许材料
	Retake       RetakePolicy  `json:"retake"`        // 重考政策
}

// ExamSection 考试部分
type ExamSection struct {
	Name           string   `json:"name"`            // 部分名称
	Weight         float64  `json:"weight"`          // 权重
	Topics         []string `json:"topics"`          // 涵盖主题
	QuestionTypes  []string `json:"question_types"`  // 题目类型
	TimeAllocation int      `json:"time_allocation"` // 时间分配
}

// RetakePolicy 重考政策
type RetakePolicy struct {
	MaxAttempts   int      `json:"max_attempts"`   // 最大尝试次数
	WaitingPeriod int      `json:"waiting_period"` // 等待期(天)
	AdditionalFee float64  `json:"additional_fee"` // 额外费用
	Prerequisites []string `json:"prerequisites"`  // 重考前置条件
}

// AssessmentMethod 评估方法
type AssessmentMethod struct {
	ID          string                `json:"id"`          // 方法标识
	Name        string                `json:"name"`        // 方法名称
	Description string                `json:"description"` // 方法描述
	Type        string                `json:"type"`        // 方法类型
	Automated   bool                  `json:"automated"`   // 是否自动化
	Weight      float64               `json:"weight"`      // 权重
	Frequency   string                `json:"frequency"`   // 评估频率
	Tools       []string              `json:"tools"`       // 使用工具
	Criteria    []AssessmentCriterion `json:"criteria"`    // 评估标准
}

// LevelDefinition 等级定义
type LevelDefinition struct {
	Level       int      `json:"level"`       // 等级数值 (1-5)
	Name        string   `json:"name"`        // 等级名称
	Description string   `json:"description"` // 等级描述
	Evidence    []string `json:"evidence"`    // 所需证据
	Confidence  float64  `json:"confidence"`  // 置信度阈值
}

// AssessmentRule 评估规则
type AssessmentRule struct {
	ID          string                 `json:"id"`          // 规则标识
	Name        string                 `json:"name"`        // 规则名称
	Description string                 `json:"description"` // 规则描述
	Condition   string                 `json:"condition"`   // 触发条件
	Action      string                 `json:"action"`      // 执行动作
	Parameters  map[string]interface{} `json:"parameters"`  // 参数
	Priority    int                    `json:"priority"`    // 优先级
	Enabled     bool                   `json:"enabled"`     // 是否启用
}

// NewCompetencyFramework 创建Go语言技能能力评估框架实例
//
// 功能说明:
//
//	本函数创建并初始化一个完整的技能能力框架，用于定义Go语言学习的能力发展体系。
//	框架整合了技能分类、进阶路径、职业发展等多个维度，支持"从入门到通天"的完整学习旅程。
//
// 参数:
//   - name: 框架名称（如"Go语言能力体系"）
//   - version: 框架版本号（如"1.0.0"），用于版本管理和演进
//
// 返回值:
//   - *CompetencyFramework: 初始化完成的能力框架，包含：
//   - 自动生成的唯一ID（格式：framework_{name}_{version}）
//   - Go语言完整技能地图（4大类别、依赖关系、进阶路径）
//   - 默认5级能力定义（新手→进阶→胜任→精通→专家）
//   - 目标受众定位："从入门到通天"
//
// 框架核心组件:
//   - SkillMap: 包含语言基础、并发编程、Web开发、系统编程四大技能类别
//   - LevelDefinitions: 5级能力认证标准（置信度0.2/0.4/0.6/0.8/0.95）
//   - Prerequisites: 技能依赖关系图，指导学习路径规划
//
// 使用场景:
//   - 系统启动时加载默认能力框架
//   - 创建自定义的技能评估体系
//   - 为不同技术栈构建类似框架
//
// 示例:
//
//	framework := NewCompetencyFramework("Go高级能力框架", "2.0.0")
//	framework.Maintainers = []string{"JIA", "技术团队"}
//	// framework已包含完整的Go技能体系
//
// 作者: JIA
func NewCompetencyFramework(name, version string) *CompetencyFramework {
	now := time.Now()
	return &CompetencyFramework{
		ID:               fmt.Sprintf("framework_%s_%s", name, version),
		Name:             name,
		Version:          version,
		CreatedAt:        now,
		UpdatedAt:        now,
		Categories:       []CompetencyCategory{},
		SkillMap:         NewGoSkillMap(),
		Progressions:     []LearningProgression{},
		Certifications:   []CertificationPath{},
		AssessmentRules:  make(map[string]AssessmentRule),
		LevelDefinitions: getDefaultLevelDefinitions(),
		Prerequisites:    make(map[string][]string),
		Language:         "Go",
		Domain:           "Software Development",
		Audience:         "从入门到通天",
		Maintainers:      []string{},
	}
}

// NewGoSkillMap 创建Go语言完整技能地图
//
// 功能说明:
//
//	构建Go语言学习的完整技能地图，包含技能分类、依赖关系、进阶路径、
//	专业化方向和职业发展等五大维度，为学习者提供清晰的成长路径。
//
// 技能地图结构:
//   - Categories: 4大技能类别（语言基础、并发、Web、系统编程）
//   - Dependencies: 技能依赖图（如goroutines依赖functions）
//   - Progressions: 进阶路径（如并发从基础→模式→高级）
//   - Specializations: 专业化轨道（Web后端、系统工具）
//   - CareerPaths: 职业路径（后端工程师：初级→高级）
//
// 返回值:
//   - SkillMap: 完整的Go技能地图结构
//
// 作者: JIA
func NewGoSkillMap() SkillMap {
	return SkillMap{
		Categories:      getGoSkillCategories(),
		Dependencies:    getSkillDependencies(),
		Progressions:    getSkillProgressions(),
		Specializations: getSpecializationTracks(),
		CareerPaths:     getCareerPaths(),
	}
}

// getGoSkillCategories 获取Go语言技能分类体系映射
//
// 功能说明:
//
//	本函数返回Go语言学习路径的四大核心技能类别定义，每个类别包含技能列表、
//	学习时长、难度等级等完整信息。这些类别构成了"从入门到通天"的技能框架基础。
//
// 四大技能类别详解:
//
//	📚 language_fundamentals（语言基础）- 难度: 简单(2.0)
//	  核心技能: syntax, types, functions
//	  高级技能: reflection, unsafe
//	  预估学习: 40小时
//	  适用: 入门阶段，掌握Go语言核心语法和编程概念
//
//	🔀 concurrency（并发编程）- 难度: 困难(4.0)
//	  核心技能: goroutines, channels
//	  高级技能: advanced_patterns, performance_tuning
//	  预估学习: 60小时
//	  适用: 进阶阶段，理解Go独特的并发模型和并发安全
//
//	🌐 web_development（Web开发）- 难度: 中等(3.0)
//	  核心技能: http, routing
//	  高级技能: microservices, graphql
//	  预估学习: 80小时
//	  适用: 实战阶段，构建Web应用和API服务
//
//	⚙️ system_programming（系统编程）- 难度: 非常困难(4.5)
//	  核心技能: cli, file_io
//	  高级技能: low_level, performance
//	  预估学习: 100小时
//	  适用: 高级阶段，开发系统工具和底层程序
//
// 返回值:
//   - map[string]SkillCategory: 技能类别映射，键为类别ID，值为类别详细定义
//
// 设计理念:
//   - 渐进式难度：从简单(2.0)到非常困难(4.5)
//   - 时长递增：从40小时到100小时反映学习深度
//   - 双层技能：核心技能打基础，高级技能促提升
//
// 使用场景:
//   - 初始化技能地图时加载类别定义
//   - 生成学习路径时计算总时长
//   - 评估学习者能力时匹配对应类别
//
// 作者: JIA
func getGoSkillCategories() map[string]SkillCategory {
	return map[string]SkillCategory{
		"language_fundamentals": {
			ID:             "language_fundamentals",
			Name:           "语言基础",
			Description:    "Go语言核心语法和概念",
			Skills:         []string{"syntax", "types", "functions", "packages"},
			CoreSkills:     []string{"syntax", "types", "functions"},
			AdvancedSkills: []string{"reflection", "unsafe"},
			EstimatedHours: evaluators.LearningHours40,
			Difficulty:     evaluators.DifficultyEasy,
		},
		"concurrency": {
			ID:             "concurrency",
			Name:           "并发编程",
			Description:    "Go语言并发模型和模式",
			Skills:         []string{"goroutines", "channels", "select", "sync"},
			CoreSkills:     []string{"goroutines", "channels"},
			AdvancedSkills: []string{"advanced_patterns", "performance_tuning"},
			EstimatedHours: evaluators.LearningHours60,
			Difficulty:     evaluators.DifficultyHard,
		},
		"web_development": {
			ID:             "web_development",
			Name:           "Web开发",
			Description:    "使用Go进行Web应用开发",
			Skills:         []string{"http", "routing", "middleware", "templates"},
			CoreSkills:     []string{"http", "routing"},
			AdvancedSkills: []string{"microservices", "graphql"},
			EstimatedHours: evaluators.LearningHours80,
			Difficulty:     evaluators.DifficultyMedium,
		},
		"system_programming": {
			ID:             "system_programming",
			Name:           "系统编程",
			Description:    "系统级编程和工具开发",
			Skills:         []string{"cli", "file_io", "networking", "databases"},
			CoreSkills:     []string{"cli", "file_io"},
			AdvancedSkills: []string{"low_level", "performance"},
			EstimatedHours: evaluators.LearningHours100,
			Difficulty:     evaluators.DifficultyVeryHard,
		},
	}
}

// getSkillDependencies 获取技能依赖关系图谱
//
// 功能说明:
//
//	本函数定义Go语言技能之间的依赖关系，形成有向无环图(DAG)结构，
//	指导学习者按照正确的前置关系顺序掌握技能，避免"跳级学习"导致基础不牢。
//
// 依赖关系详解:
//
//	每个键代表目标技能，对应的值数组列出所有前置技能（必须先掌握的技能）
//
//	核心依赖链:
//	- functions → [syntax, types]
//	  说明: 函数编程需要先理解语法和类型系统
//
//	- interfaces → [types, functions]
//	  说明: 接口是高级类型抽象，需要类型和函数基础
//
//	- goroutines → [functions]
//	  说明: 并发编程需要先掌握函数定义和调用
//
//	- channels → [goroutines]
//	  说明: 通道通信需要先理解goroutine概念
//
//	- http → [functions, interfaces]
//	  说明: Web开发需要函数和接口知识（Handler接口）
//
//	- microservices → [http, concurrency]
//	  说明: 微服务架构需要Web基础和并发能力
//
// 返回值:
//   - map[string][]string: 依赖映射，键为技能ID，值为前置技能ID列表
//
// 设计理念:
//   - 最小依赖集：只列出直接前置，不列传递依赖（如microservices不列functions）
//   - DAG结构：确保无循环依赖，可拓扑排序生成学习路径
//   - 渐进式：从基础(syntax/types)到高级(microservices)
//
// 使用场景:
//   - 生成个性化学习路径时检查前置条件
//   - 学习者尝试跳级时给出警告提示
//   - 可视化技能树时绘制依赖箭头
//
// 示例:
//
//	想学microservices → 检查依赖 → 需要http和concurrency →
//	递归检查 → http需要functions和interfaces →
//	最终路径: syntax → types → functions → interfaces → http → goroutines → channels → concurrency → microservices
//
// 作者: JIA
func getSkillDependencies() map[string][]string {
	return map[string][]string{
		"functions":     {"syntax", "types"},
		"interfaces":    {"types", "functions"},
		"goroutines":    {"functions"},
		"channels":      {"goroutines"},
		"http":          {"functions", "interfaces"},
		"microservices": {"http", "concurrency"},
	}
}

// getSkillProgressions 获取技能进阶路径映射
//
// 功能说明:
//
//	本函数定义关键技能的三级进阶路径，每个等级明确学习目标、能力表现、
//	实战项目和时长要求。当前版本聚焦于Go最核心的并发编程(concurrency)技能。
//
// 三级进阶体系（以concurrency为例）:
//
//	🌱 Level 1: 基础并发（20小时）
//	  学习目标:
//	  - 理解goroutines和channels基本概念
//	  - 掌握创建goroutines和使用unbuffered channels
//	  能力表现:
//	  - 简单并发任务、基础通信
//	  实战项目:
//	  - 并发计算器：多goroutine并行计算
//	  - 生产者-消费者模式：理解channel阻塞特性
//
//	🌿 Level 2: 并发模式（25小时）
//	  学习目标:
//	  - 掌握select语句、buffered channels、worker pools
//	  能力表现:
//	  - 复杂并发控制、资源池管理
//	  实战项目:
//	  - Web爬虫：并发抓取网页
//	  - 任务调度器：worker pool模式实现
//
//	🌳 Level 3: 高级并发（35小时）
//	  学习目标:
//	  - 掌握context使用、sync包、原子操作
//	  能力表现:
//	  - 并发安全设计、性能调优
//	  实战项目:
//	  - 高性能服务器：处理高并发请求
//	  - 分布式系统组件：实现并发安全的共享状态
//
// 总学习时长: 20 + 25 + 35 = 80小时（与并发类别预估一致）
//
// 返回值:
//   - map[string]SkillProgression: 技能进阶映射，键为技能ID，值为进阶路径定义
//
// 设计理念:
//   - 递进式难度：从简单概念到复杂模式再到性能优化
//   - 实战导向：每个等级都有对应实战项目验证能力
//   - 时长递增：Level 1最短(20h)，Level 3最长(35h)，反映难度增长
//
// 扩展性:
//
//	当前仅定义concurrency进阶路径，未来可扩展其他技能：
//	- "web_development": Level 1(基础HTTP) → Level 2(RESTful API) → Level 3(微服务)
//	- "system_programming": Level 1(CLI工具) → Level 2(网络编程) → Level 3(底层优化)
//
// 使用场景:
//   - 学习者查看技能成长路径和里程碑
//   - 系统生成阶段性学习计划
//   - 评估学习者当前等级并推荐下一步项目
//
// 作者: JIA
func getSkillProgressions() map[string]SkillProgression {
	return map[string]SkillProgression{
		"concurrency": {
			SkillID: "concurrency",
			Levels: []ProgressionLevel{
				{
					Level:          1,
					Name:           "基础并发",
					Description:    "理解goroutines和channels基本概念",
					Requirements:   []string{"创建goroutines", "使用unbuffered channels"},
					Capabilities:   []string{"简单并发任务", "基础通信"},
					Projects:       []string{"并发计算器", "生产者-消费者模式"},
					EstimatedHours: evaluators.LearningHours20,
				},
				{
					Level:          2,
					Name:           "并发模式",
					Description:    "掌握常见并发模式",
					Requirements:   []string{"select语句", "buffered channels", "worker pools"},
					Capabilities:   []string{"复杂并发控制", "资源池管理"},
					Projects:       []string{"Web爬虫", "任务调度器"},
					EstimatedHours: evaluators.LearningHours25,
				},
				{
					Level:          3,
					Name:           "高级并发",
					Description:    "性能优化和高级模式",
					Requirements:   []string{"context使用", "sync包", "原子操作"},
					Capabilities:   []string{"并发安全设计", "性能调优"},
					Projects:       []string{"高性能服务器", "分布式系统组件"},
					EstimatedHours: evaluators.LearningHours35,
				},
			},
		},
	}
}

// getSpecializationTracks 获取专业化职业轨道定义
//
// 功能说明:
//
//	本函数定义Go语言学习者可选择的两大专业化发展方向，每个轨道包含
//	必修技能、选修技能、核心项目、预估时长和职业出路，帮助学习者规划职业路径。
//
// 两大专业化轨道详解:
//
//	🌐 Track 1: Web后端开发（6个月）
//	  领域: Web Development
//	  必修技能(4项):
//	  - http: HTTP协议和服务器编程
//	  - databases: 数据库操作（SQL/NoSQL）
//	  - apis: RESTful/GraphQL API设计
//	  - authentication: 用户认证和授权
//	  选修技能(3项):
//	  - caching: Redis等缓存技术
//	  - message_queues: Kafka/RabbitMQ消息队列
//	  - monitoring: Prometheus/Grafana监控
//	  核心项目:
//	  - RESTful API: 完整的CRUD接口
//	  - 微服务架构: 服务拆分和通信
//	  - 实时系统: WebSocket实时通信
//	  职业出路:
//	  - 后端工程师、API开发者、微服务架构师
//
//	⚙️ Track 2: 系统工具开发（4个月）
//	  领域: System Programming
//	  必修技能(4项):
//	  - cli: 命令行工具开发
//	  - file_io: 文件和目录操作
//	  - system_calls: 系统调用和OS交互
//	  - cross_platform: 跨平台兼容性
//	  选修技能(2项):
//	  - performance_tuning: 性能分析和优化
//	  - memory_management: 内存管理和GC调优
//	  核心项目:
//	  - 命令行工具: 如代码生成器、部署工具
//	  - 系统监控: CPU/内存/磁盘监控
//	  - 自动化脚本: CI/CD流程自动化
//	  职业出路:
//	  - DevOps工程师、系统工程师、工具开发者
//
// 返回值:
//   - []SpecializationTrack: 专业化轨道切片，包含所有可选轨道定义
//
// 设计理念:
//   - 双轨分化：Web方向（6个月）vs 系统方向（4个月），时长反映复杂度
//   - 必选+可选：必修技能保证核心能力，选修技能提升竞争力
//   - 实战为王：每个轨道都有3个核心项目作为能力证明
//   - 职业导向：明确职业出路，帮助学习者做职业规划
//
// 使用场景:
//   - 完成基础学习后，学习者选择专业化方向
//   - 生成个性化学习路径时加载对应轨道技能
//   - 求职时根据轨道匹配目标岗位
//
// 扩展性:
//
//	未来可添加更多轨道：
//	- 云原生开发: Kubernetes, Docker, Serverless
//	- 数据工程: 数据处理、ETL、大数据
//	- 区块链开发: 智能合约、分布式账本
//
// 作者: JIA
func getSpecializationTracks() []SpecializationTrack {
	return []SpecializationTrack{
		{
			ID:                "web_backend",
			Name:              "Web后端开发",
			Description:       "专注于Web后端服务开发",
			Domain:            "Web Development",
			RequiredSkills:    []string{"http", "databases", "apis", "authentication"},
			ElectiveSkills:    []string{"caching", "message_queues", "monitoring"},
			CoreProjects:      []string{"RESTful API", "微服务架构", "实时系统"},
			EstimatedDuration: evaluators.TrackDuration6,
			CareerOutcomes:    []string{"后端工程师", "API开发者", "微服务架构师"},
		},
		{
			ID:                "system_tools",
			Name:              "系统工具开发",
			Description:       "专注于系统工具和CLI应用开发",
			Domain:            "System Programming",
			RequiredSkills:    []string{"cli", "file_io", "system_calls", "cross_platform"},
			ElectiveSkills:    []string{"performance_tuning", "memory_management"},
			CoreProjects:      []string{"命令行工具", "系统监控", "自动化脚本"},
			EstimatedDuration: evaluators.TrackDuration4,
			CareerOutcomes:    []string{"DevOps工程师", "系统工程师", "工具开发者"},
		},
	}
}

// getCareerPaths 获取职业发展路径定义
//
// 功能说明:
//
//	本函数定义完整的职业发展路径，包含从初级到高级的角色晋升梯度，
//	每个角色明确技能要求、工作职责、经验年限和薪资范围，为学习者提供清晰的职业规划蓝图。
//
// 后端工程师职业路径详解（完整晋升体系）:
//
//	🌱 Junior（初级后端工程师）- 1年经验
//	  技能要求:
//	  - http: Level 3（精通）- 熟练使用Go标准库开发HTTP服务
//	  - databases: Level 2（中级）- 掌握SQL/NoSQL基础操作
//	  - testing: Level 3（精通）- 编写完整的单元测试和集成测试
//	  工作职责:
//	  - 实现基础功能: 完成需求文档中的CRUD接口
//	  - 编写单元测试: 保证代码质量和覆盖率
//	  - 参与代码审查: 学习最佳实践，提升代码质量
//	  薪资范围（美元）:
//	  - 最低: $60,000/年
//	  - 中位: $70,000/年
//	  - 最高: $80,000/年
//	  职业特点: 在指导下工作，聚焦代码实现和测试
//
//	🏆 Senior（高级后端工程师）- 5年经验
//	  技能要求:
//	  - microservices: Level 4（专家）- 设计和实现微服务架构
//	  - performance: Level 4（专家）- 性能分析、优化和调优
//	  - leadership: Level 3（精通）- 技术指导和团队协作
//	  工作职责:
//	  - 架构设计: 制定技术方案，解决复杂技术问题
//	  - 技术选型: 评估和选择合适的技术栈和工具
//	  - 团队指导: 指导初级工程师，进行代码审查和技术培训
//	  薪资范围（美元）:
//	  - 最低: $120,000/年
//	  - 中位: $140,000/年
//	  - 最高: $160,000/年
//	  职业特点: 独立负责模块，技术决策影响团队
//
// 薪资增长分析:
//
//	从初级到高级，薪资增长约100%（$70k → $140k中位数）
//	反映技能深度、责任范围和业务影响力的大幅提升
//
// 晋升关键要素:
//  1. 技能深度: 从Level 2-3提升到Level 3-4
//  2. 技能广度: 从单一技能扩展到架构、性能、领导力
//  3. 经验积累: 从1年到5年，实战经验是核心
//  4. 职责升级: 从执行到设计，从个人到团队
//
// 返回值:
//   - []CareerPath: 职业路径切片，包含完整的角色晋升体系
//
// 设计理念:
//   - 双级模型：初级和高级代表典型晋升阶梯（可扩展为5级）
//   - 技能量化：用Level 1-5明确表示每个技能的要求等级
//   - 薪资透明：提供市场化薪资范围，帮助学习者做职业决策
//   - 职责清晰：明确每个级别的工作重点和成长方向
//
// 使用场景:
//   - 学习者规划职业发展路径
//   - 评估当前能力与目标岗位的差距
//   - 求职时参考薪资范围进行谈判
//   - HR制定技术岗位JD和薪资标准
//
// 扩展性:
//
//	当前仅定义后端工程师路径，未来可扩展：
//	- 架构师路径: Senior → Lead → Principal → Chief Architect
//	- 全栈工程师路径: 前端+后端技能组合
//	- DevOps工程师路径: 运维+开发技能融合
//
// 作者: JIA
func getCareerPaths() []CareerPath {
	return []CareerPath{
		{
			ID:          "backend_engineer",
			Title:       "后端工程师",
			Description: "专注于服务端开发和架构设计",
			Industry:    "Software Development",
			Roles: []CareerRole{
				{
					Level:            "junior",
					Title:            "初级后端工程师",
					Responsibilities: []string{"实现基础功能", "编写单元测试", "参与代码审查"},
					RequiredSkills: map[string]int{
						"http":      evaluators.SkillLevel3,
						"databases": evaluators.SkillLevel2,
						"testing":   evaluators.SkillLevel3,
					},
					Experience: evaluators.ExperienceYears1,
					Salary: SalaryRange{
						Currency:     "USD",
						MinSalary:    evaluators.SalaryJuniorMin,
						MaxSalary:    evaluators.SalaryJuniorMax,
						MedianSalary: evaluators.SalaryJuniorMedian,
					},
				},
				{
					Level:            "senior",
					Title:            "高级后端工程师",
					Responsibilities: []string{"架构设计", "技术选型", "团队指导"},
					RequiredSkills: map[string]int{
						"microservices": evaluators.SkillLevel4,
						"performance":   evaluators.SkillLevel4,
						"leadership":    evaluators.SkillLevel3,
					},
					Experience: evaluators.ExperienceYears5,
					Salary: SalaryRange{
						Currency:     "USD",
						MinSalary:    evaluators.SalarySeniorMin,
						MaxSalary:    evaluators.SalarySeniorMax,
						MedianSalary: evaluators.SalarySeniorMedian,
					},
				},
			},
		},
	}
}

// getDefaultLevelDefinitions 获取默认五级能力认证标准
//
// 功能说明:
//
//	返回基于德雷福斯技能习得模型的五级能力定义，从新手到专家的完整发展阶梯。
//	每个等级明确定义了能力特征、所需证据和置信度阈值。
//
// 五级能力体系（Dreyfus Model）:
//
//	🌱 Novice（新手，Level 1）- 置信度20%
//	  完成基础教程，实现简单功能，需要明确指导
//
//	🌿 Advanced Beginner（进阶新手，Level 2）- 置信度40%
//	  能在指导下完成任务，解决基础问题
//
//	🌳 Competent（胜任者，Level 3）- 置信度60%
//	  独立项目开发，参与代码审查，大多数任务熟练
//
//	🏆 Proficient（精通者，Level 4）- 置信度80%
//	  性能优化，架构设计，技术指导，深度理解
//
//	⭐ Expert（专家，Level 5）- 置信度95%
//	  技术创新，社区贡献，标准制定，领域专家
//
// 返回值:
//   - map[string]LevelDefinition: 以等级名称为键的能力定义映射
//
// 作者: JIA
func getDefaultLevelDefinitions() map[string]LevelDefinition {
	return map[string]LevelDefinition{
		"novice": {
			Level:       1,
			Name:        "新手",
			Description: "初学者，具备基础理解",
			Evidence:    []string{"完成基础教程", "实现简单功能"},
			Confidence:  evaluators.WeightMediumLow,
		},
		"advanced_beginner": {
			Level:       2,
			Name:        "进阶新手",
			Description: "能够在指导下完成任务",
			Evidence:    []string{"完成指导项目", "解决基础问题"},
			Confidence:  evaluators.WeightHigh,
		},
		"competent": {
			Level:       3,
			Name:        "胜任者",
			Description: "能够独立完成大多数任务",
			Evidence:    []string{"独立项目开发", "代码审查参与"},
			Confidence:  evaluators.WeightVeryHigh,
		},
		"proficient": {
			Level:       evaluators.SkillLevel4,
			Name:        "精通者",
			Description: "深度理解，能够优化和改进",
			Evidence:    []string{"性能优化", "架构设计", "技术指导"},
			Confidence:  evaluators.WeightCritical,
		},
		"expert": {
			Level:       evaluators.SkillLevel5,
			Name:        "专家",
			Description: "领域专家，能够创新和引领",
			Evidence:    []string{"技术创新", "社区贡献", "标准制定"},
			Confidence:  evaluators.WeightAlmostFull,
		},
	}
}

// ToJSON 将能力框架序列化为格式化的JSON字符串
//
// 功能说明:
//
//	序列化CompetencyFramework为易读的JSON格式（缩进2空格），
//	用于配置导出、数据持久化、API响应等场景。
//
// 返回值:
//   - []byte: JSON字节数组，UTF-8编码
//   - error: 序列化错误（极少发生）
//
// 作者: JIA
func (cf *CompetencyFramework) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cf, "", "  ")
}

// FromJSON 从JSON数据反序列化为能力框架对象
//
// 功能说明:
//
//	将JSON字节数据解析为CompetencyFramework结构体，
//	用于加载配置文件、导入数据、处理API请求等场景。
//
// 参数:
//   - data: JSON格式的字节数组
//
// 返回值:
//   - error: 解析错误（JSON格式错误或类型不匹配）
//
// 作者: JIA
func (cf *CompetencyFramework) FromJSON(data []byte) error {
	return json.Unmarshal(data, cf)
}
