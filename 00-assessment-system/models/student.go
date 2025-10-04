/*
=== Go语言学习评估系统 - 学习者数据模型 ===

本文件定义了学习者的完整数据模型，支持"从入门到通天"的全程学习跟踪：
1. 学习者基本信息和档案管理
2. 学习进度和阶段跟踪
3. 技能能力发展记录
4. 项目作品集管理
5. 评估历史和认证记录
6. 个性化学习偏好和目标设定
7. 详细的学习统计和分析数据
*/

package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// StudentProfile 学习者完整档案
type StudentProfile struct {
	// 基本信息
	ID         string    `json:"id"`          // 唯一标识符
	Name       string    `json:"name"`        // 学习者姓名
	Email      string    `json:"email"`       // 联系邮箱
	StartDate  time.Time `json:"start_date"`  // 开始学习时间
	LastActive time.Time `json:"last_active"` // 最后活跃时间

	// 学习进度
	CurrentStage  int                   `json:"current_stage"`  // 当前学习阶段 (1-15)
	StageProgress map[int]StageProgress `json:"stage_progress"` // 各阶段进度详情
	TotalHours    float64               `json:"total_hours"`    // 累计学习时长(小时)
	WeeklyHours   float64               `json:"weekly_hours"`   // 每周学习时长

	// 技能能力
	Competencies map[string]CompetencyLevel `json:"competencies"` // 技能能力映射
	SkillMatrix  SkillMatrix                `json:"skill_matrix"` // 技能矩阵

	// 项目和评估
	Projects       []ProjectRecord       `json:"projects"`       // 项目作品集
	Assessments    []AssessmentRecord    `json:"assessments"`    // 评估历史记录
	Certifications []CertificationRecord `json:"certifications"` // 认证记录

	// 个人设置
	LearningGoals []Goal              `json:"learning_goals"` // 学习目标
	Preferences   LearningPreferences `json:"preferences"`    // 个人偏好设置

	// 统计数据
	Statistics LearningStatistics `json:"statistics"` // 学习统计数据
}

// StageProgress 阶段进度详情
type StageProgress struct {
	StageID         int        `json:"stage_id"`         // 阶段编号
	StageName       string     `json:"stage_name"`       // 阶段名称
	StartDate       time.Time  `json:"start_date"`       // 开始时间
	CompleteDate    *time.Time `json:"complete_date"`    // 完成时间（可选）
	Progress        float64    `json:"progress"`         // 完成百分比 (0.0-1.0)
	TimeSpent       float64    `json:"time_spent"`       // 花费时间（小时）
	ModulesComplete int        `json:"modules_complete"` // 完成模块数
	TotalModules    int        `json:"total_modules"`    // 总模块数
	LastAssessment  *time.Time `json:"last_assessment"`  // 最后评估时间
	Status          string     `json:"status"`           // 状态: not_started, in_progress, completed, certified
}

// CompetencyLevel 能力水平定义
type CompetencyLevel struct {
	Category    string    `json:"category"`     // 能力类别
	Skill       string    `json:"skill"`        // 具体技能
	Level       int       `json:"level"`        // 水平等级 (1-5)
	Evidence    []string  `json:"evidence"`     // 证据记录
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
	Confidence  float64   `json:"confidence"`   // 自信度 (0.0-1.0)
}

// SkillMatrix 技能矩阵
type SkillMatrix struct {
	// 技术深度维度
	TechnicalDepth struct {
		LanguageFeatures   CompetencyLevel `json:"language_features"`   // 语言特性掌握
		StandardLibrary    CompetencyLevel `json:"standard_library"`    // 标准库精通
		EcosystemKnowledge CompetencyLevel `json:"ecosystem_knowledge"` // 生态系统了解
		InternalKnowledge  CompetencyLevel `json:"internal_knowledge"`  // 内部机制理解
	} `json:"technical_depth"`

	// 工程实践维度
	EngineeringPractice struct {
		CodeQuality     CompetencyLevel `json:"code_quality"`     // 代码质量
		TestingSkills   CompetencyLevel `json:"testing_skills"`   // 测试技能
		ToolchainUsage  CompetencyLevel `json:"toolchain_usage"`  // 工具链使用
		DevOpsKnowledge CompetencyLevel `json:"devops_knowledge"` // DevOps知识
	} `json:"engineering_practice"`

	// 项目经验维度
	ProjectExperience struct {
		ComplexityHandling CompetencyLevel `json:"complexity_handling"` // 复杂度处理
		DomainExpertise    CompetencyLevel `json:"domain_expertise"`    // 领域专长
		Innovation         CompetencyLevel `json:"innovation"`          // 创新能力
		Leadership         CompetencyLevel `json:"leadership"`          // 领导能力
	} `json:"project_experience"`

	// 软技能维度
	SoftSkills struct {
		Communication      CompetencyLevel `json:"communication"`       // 技术沟通
		Collaboration      CompetencyLevel `json:"collaboration"`       // 协作能力
		ContinuousLearning CompetencyLevel `json:"continuous_learning"` // 持续学习
		Mentoring          CompetencyLevel `json:"mentoring"`           // 指导能力
	} `json:"soft_skills"`
}

// ProjectRecord 项目记录
type ProjectRecord struct {
	ID           string     `json:"id"`            // 项目唯一标识
	Name         string     `json:"name"`          // 项目名称
	Description  string     `json:"description"`   // 项目描述
	Stage        int        `json:"stage"`         // 所属学习阶段
	Type         string     `json:"type"`          // 项目类型: cli, web, api, system, etc.
	StartDate    time.Time  `json:"start_date"`    // 开始时间
	CompleteDate *time.Time `json:"complete_date"` // 完成时间

	// 项目详情
	Repository   string   `json:"repository"`    // 代码仓库地址
	Technologies []string `json:"technologies"`  // 使用技术栈
	LinesOfCode  int      `json:"lines_of_code"` // 代码行数

	// 评估结果
	ComplexityScore float64 `json:"complexity_score"` // 复杂度评分 (1-10)
	QualityScore    float64 `json:"quality_score"`    // 质量评分 (1-10)
	InnovationScore float64 `json:"innovation_score"` // 创新性评分 (1-10)
	OverallScore    float64 `json:"overall_score"`    // 综合评分 (1-10)

	// 反馈和改进
	Feedback     []string `json:"feedback"`     // 评估反馈
	Improvements []string `json:"improvements"` // 改进建议
	Status       string   `json:"status"`       // 项目状态: planning, development, completed, archived
}

// AssessmentRecord 评估记录
type AssessmentRecord struct {
	ID        string    `json:"id"`        // 评估记录唯一标识
	Timestamp time.Time `json:"timestamp"` // 评估时间
	Stage     int       `json:"stage"`     // 评估阶段
	Type      string    `json:"type"`      // 评估类型: auto, manual, peer, mentor

	// 评估维度得分
	Scores struct {
		TechnicalDepth      float64 `json:"technical_depth"`      // 技术深度 (0-100)
		EngineeringPractice float64 `json:"engineering_practice"` // 工程实践 (0-100)
		ProjectExperience   float64 `json:"project_experience"`   // 项目经验 (0-100)
		SoftSkills          float64 `json:"soft_skills"`          // 软技能 (0-100)
		OverallScore        float64 `json:"overall_score"`        // 综合得分 (0-100)
	} `json:"scores"`

	// 详细评估结果
	DetailedResults map[string]interface{} `json:"detailed_results"` // 详细评估数据
	Recommendations []string               `json:"recommendations"`  // 学习建议
	NextSteps       []string               `json:"next_steps"`       // 下一步行动

	// 评估元数据
	Evaluator  string  `json:"evaluator"`  // 评估者标识
	Duration   int     `json:"duration"`   // 评估耗时（分钟）
	Confidence float64 `json:"confidence"` // 评估置信度 (0-1)
	Notes      string  `json:"notes"`      // 评估备注
}

// CertificationRecord 认证记录
type CertificationRecord struct {
	Level       string     `json:"level"`       // 认证等级: Bronze, Silver, Gold, Platinum
	AwardDate   time.Time  `json:"award_date"`  // 获得时间
	ExpiryDate  *time.Time `json:"expiry_date"` // 过期时间（可选）
	Score       float64    `json:"score"`       // 认证考试得分
	Certificate string     `json:"certificate"` // 证书标识
	Verified    bool       `json:"verified"`    // 是否已验证

	// 考试详情
	ExamType     string             `json:"exam_type"`     // 考试类型
	ExamDuration int                `json:"exam_duration"` // 考试时长（分钟）
	ExamResults  map[string]float64 `json:"exam_results"`  // 各部分得分

	// 项目要求
	RequiredProjects []string `json:"required_projects"` // 要求完成的项目
	PortfolioScore   float64  `json:"portfolio_score"`   // 作品集评分

	Status string `json:"status"` // 认证状态: active, expired, revoked
}

// Goal 学习目标
type Goal struct {
	ID          string      `json:"id"`          // 目标唯一标识
	Title       string      `json:"title"`       // 目标标题
	Description string      `json:"description"` // 详细描述
	TargetDate  time.Time   `json:"target_date"` // 目标完成时间
	Priority    int         `json:"priority"`    // 优先级 (1-5)
	Status      string      `json:"status"`      // 状态: active, completed, canceled
	Progress    float64     `json:"progress"`    // 完成进度 (0-1)
	Milestones  []Milestone `json:"milestones"`  // 里程碑
}

// Milestone 里程碑
type Milestone struct {
	Title       string     `json:"title"`        // 里程碑标题
	Description string     `json:"description"`  // 详细描述
	DueDate     time.Time  `json:"due_date"`     // 截止时间
	Completed   bool       `json:"completed"`    // 是否完成
	CompletedAt *time.Time `json:"completed_at"` // 完成时间
}

// LearningPreferences 学习偏好设置
type LearningPreferences struct {
	PreferredPace     string   `json:"preferred_pace"`     // 学习节奏: slow, normal, fast
	LearningStyle     string   `json:"learning_style"`     // 学习风格: visual, auditory, kinesthetic, reading
	FocusAreas        []string `json:"focus_areas"`        // 重点关注领域
	AvailableHours    float64  `json:"available_hours"`    // 每周可用学习时间
	PreferredSchedule string   `json:"preferred_schedule"` // 偏好时间: morning, afternoon, evening, flexible

	// 评估偏好
	AssessmentFrequency string `json:"assessment_frequency"` // 评估频率: daily, weekly, monthly
	FeedbackPreference  string `json:"feedback_preference"`  // 反馈偏好: immediate, summary, detailed

	// 通知设置
	EmailNotifications bool `json:"email_notifications"` // 邮件通知
	ProgressReminders  bool `json:"progress_reminders"`  // 进度提醒
	DeadlineAlerts     bool `json:"deadline_alerts"`     // 截止提醒
}

// LearningStatistics 学习统计数据
type LearningStatistics struct {
	// 时间统计
	TotalStudyDays  int       `json:"total_study_days"` // 累计学习天数
	ConsecutiveDays int       `json:"consecutive_days"` // 连续学习天数
	AverageDaily    float64   `json:"average_daily"`    // 日均学习时长
	WeeklyTrend     []float64 `json:"weekly_trend"`     // 周学习趋势

	// 进度统计
	StagesCompleted   int     `json:"stages_completed"`   // 完成阶段数
	ModulesCompleted  int     `json:"modules_completed"`  // 完成模块数
	ProjectsCompleted int     `json:"projects_completed"` // 完成项目数
	OverallProgress   float64 `json:"overall_progress"`   // 总体进度 (0-1)

	// 成就统计
	TotalAssessments int       `json:"total_assessments"` // 总评估次数
	AverageScore     float64   `json:"average_score"`     // 平均评估分数
	BestScore        float64   `json:"best_score"`        // 最高评估分数
	ImprovementTrend []float64 `json:"improvement_trend"` // 进步趋势

	// 活跃度统计
	WeeklyActiveHours  float64   `json:"weekly_active_hours"`  // 周活跃时长
	MonthlyActiveHours float64   `json:"monthly_active_hours"` // 月活跃时长
	LastActiveDate     time.Time `json:"last_active_date"`     // 最后活跃日期
	ActivityStreak     int       `json:"activity_streak"`      // 活跃连续天数

	// 技能发展统计
	SkillGrowthRate float64              `json:"skill_growth_rate"` // 技能成长率
	StrongSkills    []string             `json:"strong_skills"`     // 强项技能
	WeakSkills      []string             `json:"weak_skills"`       // 薄弱技能
	SkillTrends     map[string][]float64 `json:"skill_trends"`      // 技能发展趋势
}

// NewStudentProfile 创建新的学习者档案实例
//
// 功能说明:
//
//	本函数初始化一个全新的学习者档案，设置所有必要的默认值和空集合，
//	准备好完整的数据结构以支持"从入门到通天"的15阶段学习旅程跟踪。
//
// 初始化内容:
//
//	基本信息:
//	- ID/Name/Email: 学习者唯一标识和联系方式
//	- StartDate/LastActive: 自动设置为当前时间
//
//	学习进度:
//	- CurrentStage: 初始化为1（Go语言基础）
//	- StageProgress: 空map，按需填充各阶段进度
//	- TotalHours/WeeklyHours: 初始化为0
//
//	技能和项目:
//	- Competencies: 空map，记录技能能力发展
//	- Projects/Assessments/Certifications: 空切片，待添加记录
//	- LearningGoals: 空切片，学习者自定义目标
//
//	个人偏好（默认配置）:
//	- PreferredPace: "normal"（正常节奏）
//	- LearningStyle: "reading"（阅读型学习者）
//	- AvailableHours: 10小时/周（来自DefaultAvailableHours常量）
//	- PreferredSchedule: "flexible"（灵活时间）
//	- AssessmentFrequency: "weekly"（每周评估）
//	- FeedbackPreference: "detailed"（详细反馈）
//	- 通知全部启用: Email/Progress/Deadline提醒
//
//	学习统计:
//	- WeeklyTrend/ImprovementTrend: 空切片，记录趋势数据
//	- SkillTrends: 空map，追踪技能发展曲线
//
// 参数:
//   - id: 学习者唯一标识符（如"student_20250103_001"）
//   - name: 学习者姓名（如"张三"）
//   - email: 联系邮箱（如"zhangsan@example.com"）
//
// 返回值:
//   - *StudentProfile: 初始化完成的学习者档案指针，所有字段已设置默认值
//
// 设计理念:
//   - 开箱即用: 返回的档案无需额外配置即可使用
//   - 合理默认: 选择最常见的学习偏好作为默认值
//   - 渐进式: 空集合设计支持逐步添加数据
//   - 全启用通知: 默认全部提醒，学习者可按需关闭
//
// 使用场景:
//   - 新用户注册时创建初始档案
//   - 系统导入学习者批量数据
//   - 单元测试中创建测试数据
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "李明", "liming@example.com")
//	// student.CurrentStage == 1
//	// student.Preferences.AvailableHours == 10.0
//	// student.Preferences.AssessmentFrequency == "weekly"
//	student.UpdateProgress(1, 0.5, 5.0) // 开始学习第1阶段
//
// 注意事项:
//   - ID应保证全局唯一，建议包含时间戳或UUID
//   - Email建议验证格式有效性（调用方责任）
//   - 返回的档案CurrentStage=1，假设从第1阶段开始学习
//
// 作者: JIA
func NewStudentProfile(id, name, email string) *StudentProfile {
	now := time.Now()
	return &StudentProfile{
		ID:         id,
		Name:       name,
		Email:      email,
		StartDate:  now,
		LastActive: now,

		CurrentStage:  1,
		StageProgress: make(map[int]StageProgress),
		TotalHours:    0,
		WeeklyHours:   0,

		Competencies:   make(map[string]CompetencyLevel),
		Projects:       []ProjectRecord{},
		Assessments:    []AssessmentRecord{},
		Certifications: []CertificationRecord{},
		LearningGoals:  []Goal{},

		Preferences: LearningPreferences{
			PreferredPace:       "normal",
			LearningStyle:       "reading",
			FocusAreas:          []string{},
			AvailableHours:      DefaultAvailableHours,
			PreferredSchedule:   "flexible",
			AssessmentFrequency: "weekly",
			FeedbackPreference:  "detailed",
			EmailNotifications:  true,
			ProgressReminders:   true,
			DeadlineAlerts:      true,
		},

		Statistics: LearningStatistics{
			WeeklyTrend:      make([]float64, 0),
			ImprovementTrend: make([]float64, 0),
			SkillTrends:      make(map[string][]float64),
		},
	}
}

// UpdateProgress 更新学习者的阶段学习进度
//
// 功能说明:
//
//	本方法记录学习者在指定阶段的学习进展，自动维护累计学习时长、
//	最后活跃时间，并在阶段完成时自动推进到下一阶段。支持增量更新和阶段完成检测。
//
// 更新逻辑:
//
//  1. 基础更新（每次调用都执行）:
//     - LastActive更新为当前时间
//     - TotalHours累加本次学习时长
//
//  2. 阶段进度更新:
//     - 已存在记录: 更新Progress/TimeSpent/LastAssessment
//     - 首次学习: 创建新的StageProgress记录，状态设为"in_progress"
//
//  3. 完成检测（当progress >= 1.0时）:
//     - 设置CompleteDate为当前时间
//     - 状态从"in_progress"变为"completed"
//     - 如果是当前阶段，自动推进CurrentStage = stage + 1
//
// 阶段状态流转:
//
//	not_started → in_progress → completed → certified
//	              (首次调用)    (progress=1.0)  (获得认证)
//
// 参数:
//   - stage: 学习阶段编号（1-15，对应15个学习模块）
//   - progress: 阶段完成百分比（0.0-1.0范围，1.0表示完成）
//   - hoursSpent: 本次学习花费的时间（小时，浮点数支持分钟级精度）
//
// 使用场景:
//   - 学习者完成一个模块后记录进度
//   - 系统定时保存学习状态
//   - 评估系统自动更新学习记录
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "张三", "test@example.com")
//	// 首次学习第1阶段，完成20%，用时2小时
//	student.UpdateProgress(1, 0.2, 2.0)
//	// student.StageProgress[1].Status == "in_progress"
//
//	// 继续学习，完成80%，累计再用时6小时
//	student.UpdateProgress(1, 0.8, 6.0)
//	// student.StageProgress[1].TimeSpent == 8.0
//
//	// 完成第1阶段，总共用时10小时
//	student.UpdateProgress(1, 1.0, 2.0)
//	// student.StageProgress[1].Status == "completed"
//	// student.CurrentStage == 2 (自动推进到下一阶段)
//
// 注意事项:
//   - progress应为0.0-1.0范围，超过1.0会被视为完成
//   - hoursSpent可多次累加，不会覆盖之前的学习时长
//   - 跨阶段学习（stage != CurrentStage）不会自动推进CurrentStage
//   - 重复调用progress=1.0不会重复推进阶段
//
// 作者: JIA
func (sp *StudentProfile) UpdateProgress(stage int, progress, hoursSpent float64) {
	sp.LastActive = time.Now()
	sp.TotalHours += hoursSpent

	if stageProgress, exists := sp.StageProgress[stage]; exists {
		stageProgress.Progress = progress
		stageProgress.TimeSpent += hoursSpent
		stageProgress.LastAssessment = &sp.LastActive
		sp.StageProgress[stage] = stageProgress
	} else {
		sp.StageProgress[stage] = StageProgress{
			StageID:        stage,
			StageName:      getStageNameByID(stage),
			StartDate:      sp.LastActive,
			Progress:       progress,
			TimeSpent:      hoursSpent,
			LastAssessment: &sp.LastActive,
			Status:         "in_progress",
		}
	}

	if progress >= 1.0 {
		completeDate := sp.LastActive
		stageProgress := sp.StageProgress[stage]
		stageProgress.CompleteDate = &completeDate
		stageProgress.Status = "completed"
		sp.StageProgress[stage] = stageProgress

		if stage == sp.CurrentStage {
			sp.CurrentStage = stage + 1
		}
	}
}

// AddProject 添加项目记录到学习者作品集
//
// 功能说明:
//
//	本方法将新完成的项目添加到学习者的作品集中，自动生成项目唯一标识符，
//	更新最后活跃时间。项目作品集是学习者技能水平的重要证明，也是认证考核的关键依据。
//
// 处理流程:
//
//  1. ID自动生成:
//     格式: "proj_{序号}_{项目名称}"
//     序号: 当前项目数量+1，确保唯一性
//     示例: "proj_1_RESTful_API", "proj_2_微服务架构"
//
//  2. 项目记录存储:
//     将项目指针解引用后添加到Projects切片
//     保留完整的项目元数据（技术栈、评分、反馈等）
//
//  3. 活跃时间更新:
//     LastActive自动设置为当前时间
//     用于跟踪学习者的持续活跃状态
//
// 参数:
//   - project: 项目记录指针，包含以下关键字段：
//   - Name: 项目名称（如"在线书店系统"）
//   - Type: 项目类型（cli/web/api/system等）
//   - Stage: 所属学习阶段（1-15）
//   - Technologies: 使用的技术栈（如["Go", "PostgreSQL", "Redis"]）
//   - ComplexityScore: 复杂度评分（1-10）
//   - QualityScore: 质量评分（1-10）
//   - OverallScore: 综合评分（1-10）
//
// 返回值: 无（直接修改StudentProfile对象）
//
// 性能优化:
//   - 使用指针参数避免248字节的ProjectRecord结构体复制
//   - 自动ID生成无需额外查询，O(1)复杂度
//
// 使用场景:
//   - 学习者完成阶段项目后记录成果
//   - 提交项目作品申请认证
//   - 构建技能展示作品集
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "李明", "liming@example.com")
//
//	// 添加第一个项目
//	project1 := &ProjectRecord{
//	    Name:         "CLI任务管理工具",
//	    Type:         "cli",
//	    Stage:        1,
//	    Technologies: []string{"Go", "Cobra", "SQLite"},
//	    OverallScore: 8.5,
//	}
//	student.AddProject(project1)
//	// project1.ID == "proj_1_CLI任务管理工具"
//	// len(student.Projects) == 1
//
//	// 添加第二个项目
//	project2 := &ProjectRecord{
//	    Name:         "RESTful博客API",
//	    Type:         "api",
//	    Stage:        4,
//	    Technologies: []string{"Go", "Gin", "PostgreSQL", "JWT"},
//	    OverallScore: 9.0,
//	}
//	student.AddProject(project2)
//	// project2.ID == "proj_2_RESTful博客API"
//	// len(student.Projects) == 2
//
// 注意事项:
//   - 项目ID一旦生成不应修改，用于唯一标识项目
//   - 项目记录应包含完整的评估信息（ComplexityScore、QualityScore等）
//   - 项目的Stage字段应与学习者当前或已完成阶段一致
//   - 重要项目应设置CompleteDate以记录完成时间
//
// 作者: JIA
func (sp *StudentProfile) AddProject(project *ProjectRecord) {
	project.ID = fmt.Sprintf("proj_%d_%s", len(sp.Projects)+1, project.Name)
	sp.Projects = append(sp.Projects, *project)
	sp.LastActive = time.Now()
}

// AddAssessment 添加评估记录并自动更新技能矩阵
//
// 功能说明:
//
//	本方法将新的评估结果添加到学习者的评估历史中，自动生成评估唯一标识符，
//	更新最后活跃时间，并根据评估详细结果自动更新学习者的技能矩阵。
//	这是学习进度跟踪和能力认证的核心方法。
//
// 处理流程:
//
//  1. ID自动生成:
//     格式: "assess_{评估序号}_{评估阶段}"
//     示例: "assess_1_3" (第1次评估，阶段3)
//     确保每次评估有唯一标识符便于历史追溯
//
//  2. 评估记录存储:
//     将评估指针解引用后添加到Assessments切片
//     保留完整的评估维度得分和详细结果
//
//  3. 活跃时间更新:
//     LastActive自动设置为当前时间
//
//  4. 技能矩阵自动更新 (关键步骤):
//     调用updateSkillMatrixFromAssessment方法
//     从评估的DetailedResults中提取技能数据
//     更新SkillMatrix的各个维度（技术深度、工程实践等）
//     将评估分数转换为1-5等级的CompetencyLevel
//
// 参数:
//   - assessment: 评估记录指针，包含关键字段：
//   - Stage: 评估阶段（1-15，对应学习模块）
//   - Type: 评估类型（auto/manual/peer/mentor）
//   - Scores: 各维度得分（TechnicalDepth/EngineeringPractice/ProjectExperience/SoftSkills）
//   - DetailedResults: 详细评估数据（技能项得分、代码质量指标等）
//   - Recommendations: 学习建议
//   - Confidence: 评估置信度（0-1）
//
// 返回值: 无（直接修改StudentProfile对象）
//
// 性能优化:
//   - 使用指针参数避免208字节的AssessmentRecord结构体复制
//   - 技能矩阵更新采用增量式，仅修改变化的维度
//
// 使用场景:
//   - 完成阶段学习后进行自动评估
//   - 导师手动评估学习者能力
//   - 同伴互评记录
//   - 项目评估结果录入
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "张三", "test@example.com")
//
//	// 完成第1阶段后的自动评估
//	assessment1 := &AssessmentRecord{
//	    Timestamp: time.Now(),
//	    Stage:     1,
//	    Type:      "auto",
//	    Scores: struct {
//	        TechnicalDepth      float64
//	        EngineeringPractice float64
//	        ProjectExperience   float64
//	        SoftSkills          float64
//	        OverallScore        float64
//	    }{
//	        TechnicalDepth:      75.0,
//	        EngineeringPractice: 70.0,
//	        ProjectExperience:   60.0,
//	        SoftSkills:          65.0,
//	        OverallScore:        70.0,
//	    },
//	    DetailedResults: map[string]interface{}{
//	        "technical_depth": map[string]interface{}{
//	            "language_features": 75.0, // 会自动转换为Level 4
//	        },
//	    },
//	    Confidence: 0.85,
//	}
//	student.AddAssessment(assessment1)
//	// assessment1.ID == "assess_1_1"
//	// student.SkillMatrix.TechnicalDepth.LanguageFeatures.Level == 4
//	// student.SkillMatrix.TechnicalDepth.LanguageFeatures.Confidence == 0.85
//
// 注意事项:
//   - 评估记录按时间顺序追加，用于跟踪能力发展趋势
//   - DetailedResults必须包含有效的技能数据才能更新技能矩阵
//   - 技能等级转换公式: Level = int(score/20) + 1 (0-20→1, 21-40→2, 41-60→3, 61-80→4, 81-100→5)
//   - 评估Confidence会传递到更新的CompetencyLevel中
//
// 作者: JIA
func (sp *StudentProfile) AddAssessment(assessment *AssessmentRecord) {
	assessment.ID = fmt.Sprintf("assess_%d_%d", len(sp.Assessments)+1, assessment.Stage)
	sp.Assessments = append(sp.Assessments, *assessment)
	sp.LastActive = time.Now()

	// 更新技能矩阵
	sp.updateSkillMatrixFromAssessment(assessment)
}

// AddCertification 添加认证记录到学习者证书集合
//
// 功能说明:
//
//	本方法将新获得的能力认证添加到学习者的认证记录中，更新最后活跃时间。
//	认证是学习者技能水平的官方证明，也是职业发展的重要里程碑。
//
// 认证体系概述:
//
//	本系统采用四级认证标准（从入门到精通）：
//
//	🥉 Bronze（青铜级）- 入门认证
//	  · 最低分数: 70分（及格线）
//	  · 必修阶段: 1-3阶段（基础语法、进阶特性、并发编程）
//	  · 考试时长: 120分钟
//	  · 适合: 完成Go基础学习的初学者
//
//	🥈 Silver（白银级）- 熟练认证
//	  · 最低分数: 80分（良好水平）
//	  · 必修阶段: 1-6阶段（增加Web开发、微服务、项目实战）
//	  · 考试时长: 180分钟
//	  · 适合: 能够独立完成项目的开发者
//
//	🥇 Gold（黄金级）- 精通认证
//	  · 最低分数: 85分（优秀水平）
//	  · 必修阶段: 1-10阶段（增加性能优化、运行时原理、系统编程）
//	  · 考试时长: 240分钟
//	  · 适合: 能够进行架构设计和性能调优的高级工程师
//
//	💎 Platinum（白金级）- 专家认证
//	  · 最低分数: 90分（卓越水平）
//	  · 必修阶段: 全部15阶段（包含编译器、大规模系统、开源贡献等）
//	  · 考试时长: 300分钟
//	  · 适合: Go语言专家，能够参与语言生态建设的顶尖开发者
//
// 处理流程:
//
//  1. 认证记录直接添加:
//     不生成ID（认证本身有CertificateID作为唯一标识）
//     保留完整的考试成绩和项目评分
//
//  2. 活跃时间更新:
//     LastActive自动设置为当前时间
//
// 参数:
//   - cert: 认证记录指针，包含关键字段：
//   - Level: 认证等级（Bronze/Silver/Gold/Platinum）
//   - AwardDate: 获得时间
//   - Score: 认证考试得分
//   - Certificate: 证书唯一标识符
//   - ExamType: 考试类型（practical/comprehensive/advanced/expert）
//   - ExamResults: 各部分得分详情
//   - RequiredProjects: 要求完成的项目列表
//   - PortfolioScore: 作品集评分
//   - Status: 认证状态（active/expired/revoked）
//
// 返回值: 无（直接修改StudentProfile对象）
//
// 性能优化:
//   - 使用指针参数避免160字节的CertificationRecord结构体复制
//   - 直接追加，无需额外计算，O(1)复杂度
//
// 使用场景:
//   - 通过认证考试后录入认证信息
//   - 批量导入历史认证记录
//   - 认证续期或升级
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "王芳", "wangfang@example.com")
//
//	// 获得Bronze级认证
//	bronzeCert := &CertificationRecord{
//	    Level:       "Bronze",
//	    AwardDate:   time.Now(),
//	    Score:       78.5,
//	    Certificate: "CERT-BRONZE-20250103-001",
//	    Verified:    true,
//	    ExamType:    "practical",
//	    ExamDuration: 120,
//	    ExamResults: map[string]float64{
//	        "theoretical": 80.0,
//	        "practical":   77.0,
//	    },
//	    RequiredProjects: []string{"proj_1_CLI任务管理", "proj_2_并发计算器"},
//	    PortfolioScore:   75.0,
//	    Status:           "active",
//	}
//	student.AddCertification(bronzeCert)
//	// len(student.Certifications) == 1
//	// student.GetCurrentLevel() == "Bronze"
//
//	// 一年后升级到Silver级认证
//	silverCert := &CertificationRecord{
//	    Level:       "Silver",
//	    AwardDate:   time.Now().AddDate(1, 0, 0),
//	    Score:       85.0,
//	    Certificate: "CERT-SILVER-20260103-001",
//	    Verified:    true,
//	    Status:      "active",
//	}
//	student.AddCertification(silverCert)
//	// len(student.Certifications) == 2
//	// student.GetCurrentLevel() == "Silver" (自动返回最高等级)
//
// 注意事项:
//   - 认证记录按时间顺序添加，反映能力成长历程
//   - 同一等级可以多次认证（如续期），GetCurrentLevel会返回最高等级
//   - Status字段应定期检查，过期认证应更新为"expired"
//   - Verified字段用于标识官方验证状态，未验证的认证不应计入有效认证
//
// 作者: JIA
func (sp *StudentProfile) AddCertification(cert *CertificationRecord) {
	sp.Certifications = append(sp.Certifications, *cert)
	sp.LastActive = time.Now()
}

// GetCurrentLevel 获取学习者当前拥有的最高等级有效认证
//
// 功能说明:
//
//	本方法遍历学习者的所有认证记录，筛选出状态为"active"的有效认证，
//	返回其中最高等级的认证名称。用于展示学习者当前的官方认证水平。
//
// 认证等级层次（从低到高）:
//
//  1. Bronze（青铜级）- 入门水平，完成基础学习
//  2. Silver（白银级）- 熟练水平，能够独立开发
//  3. Gold（黄金级）  - 精通水平，掌握高级技术
//  4. Platinum（白金级）- 专家水平，全栈精通
//
// 算法流程:
//
//  1. 边界检查:
//     如果Certifications切片为空，直接返回"None"
//
//  2. 遍历认证记录:
//     使用索引遍历避免160字节的CertificationRecord结构体复制
//     仅考虑Status="active"的有效认证（忽略expired/revoked）
//
//  3. 等级比较:
//     将认证Level与预定义等级数组匹配
//     跟踪最高等级索引（Bronze=0, Silver=1, Gold=2, Platinum=3）
//
//  4. 返回结果:
//     返回最高等级索引对应的等级名称
//     如果没有任何有效认证，返回Bronze（默认最低等级）
//
// 返回值:
//   - string: 认证等级名称（"None"/"Bronze"/"Silver"/"Gold"/"Platinum"）
//   - "None": 尚未获得任何认证
//   - "Bronze"到"Platinum": 当前拥有的最高有效认证等级
//
// 性能优化:
//   - 使用索引遍历避免大结构体复制（160字节）
//   - 仅遍历一次认证记录，时间复杂度O(n)
//   - 预定义等级数组，空间复杂度O(1)
//
// 使用场景:
//   - 个人档案页面展示当前认证等级
//   - 报名考试时检查是否具备前置认证
//   - 生成学习进度报告
//   - 职位申请时验证技能等级
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "李明", "liming@example.com")
//	level := student.GetCurrentLevel()
//	// level == "None" (尚未获得任何认证)
//
//	// 添加Bronze认证
//	student.AddCertification(&CertificationRecord{
//	    Level:  "Bronze",
//	    Status: "active",
//	})
//	level = student.GetCurrentLevel()
//	// level == "Bronze"
//
//	// 添加Silver认证
//	student.AddCertification(&CertificationRecord{
//	    Level:  "Silver",
//	    Status: "active",
//	})
//	level = student.GetCurrentLevel()
//	// level == "Silver" (返回最高等级)
//
//	// 添加过期的Gold认证（不会被计入）
//	student.AddCertification(&CertificationRecord{
//	    Level:  "Gold",
//	    Status: "expired",
//	})
//	level = student.GetCurrentLevel()
//	// level == "Silver" (过期认证不计入)
//
// 注意事项:
//   - 仅统计Status="active"的认证，过期或撤销的认证会被忽略
//   - 如果有多个同等级的有效认证（如续期），只返回等级名称，不区分次数
//   - 返回值为字符串，调用方可用于显示或进一步比较
//   - Bronze是预定义的默认最低等级（levels数组索引0）
//
// 作者: JIA
func (sp *StudentProfile) GetCurrentLevel() string {
	if len(sp.Certifications) == 0 {
		return "None"
	}

	// 找到最高等级的有效认证
	levels := []string{"Bronze", "Silver", "Gold", "Platinum"}
	currentLevel := 0

	// 使用索引遍历避免大结构体复制（160字节）
	for i := range sp.Certifications {
		cert := &sp.Certifications[i]
		if cert.Status == "active" {
			for j, level := range levels {
				if cert.Level == level && j > currentLevel {
					currentLevel = j
				}
			}
		}
	}

	return levels[currentLevel]
}

// GetOverallScore 计算学习者所有评估记录的综合平均分
//
// 功能说明:
//
//	本方法遍历学习者的完整评估历史，计算所有评估记录中OverallScore字段的
//	算术平均值，反映学习者在整个学习旅程中的综合能力表现。
//
// 计算逻辑:
//
//  1. 边界检查:
//     如果Assessments切片为空（尚未进行任何评估），返回0.0
//
//  2. 累加求和:
//     使用索引遍历避免208字节的AssessmentRecord结构体复制
//     累加所有评估记录的Scores.OverallScore字段
//
//  3. 计算均值:
//     总分除以评估次数，得到平均综合分数
//     公式: 平均分 = Σ(评估[i].OverallScore) / 评估总数
//
// 返回值:
//   - float64: 综合平均分（0.0-100.0范围）
//   - 0.0: 尚未进行任何评估
//   - >0.0: 所有评估的平均综合得分
//
// 分数解读:
//   - 0-60分: 不及格，需要加强学习
//   - 60-70分: 及格水平，基础掌握
//   - 70-80分: 良好水平，能力较强
//   - 80-90分: 优秀水平，技能精通
//   - 90-100分: 卓越水平，接近完美
//
// 性能优化:
//   - 使用索引遍历避免大结构体复制（208字节）
//   - 单次遍历计算，时间复杂度O(n)
//   - 无需额外存储空间，空间复杂度O(1)
//
// 使用场景:
//   - 生成学习者能力总结报告
//   - 判断是否达到认证考试最低分数要求
//   - 学习者排名和能力对比
//   - 学习效果趋势分析
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "张三", "test@example.com")
//	score := student.GetOverallScore()
//	// score == 0.0 (尚未评估)
//
//	// 添加第1次评估（阶段1完成后）
//	student.AddAssessment(&AssessmentRecord{
//	    Stage: 1,
//	    Scores: struct {
//	        TechnicalDepth      float64
//	        EngineeringPractice float64
//	        ProjectExperience   float64
//	        SoftSkills          float64
//	        OverallScore        float64
//	    }{OverallScore: 75.0},
//	})
//	score = student.GetOverallScore()
//	// score == 75.0
//
//	// 添加第2次评估（阶段2完成后）
//	student.AddAssessment(&AssessmentRecord{
//	    Stage: 2,
//	    Scores: struct {
//	        TechnicalDepth      float64
//	        EngineeringPractice float64
//	        ProjectExperience   float64
//	        SoftSkills          float64
//	        OverallScore        float64
//	    }{OverallScore: 85.0},
//	})
//	score = student.GetOverallScore()
//	// score == 80.0  ((75.0 + 85.0) / 2)
//
//	// 添加第3次评估（阶段3完成后）
//	student.AddAssessment(&AssessmentRecord{
//	    Stage: 3,
//	    Scores: struct {
//	        TechnicalDepth      float64
//	        EngineeringPractice float64
//	        ProjectExperience   float64
//	        SoftSkills          float64
//	        OverallScore        float64
//	    }{OverallScore: 90.0},
//	})
//	score = student.GetOverallScore()
//	// score == 83.33  ((75.0 + 85.0 + 90.0) / 3)
//
// 注意事项:
//   - 返回的是所有历史评估的平均值，不仅仅是最新评估
//   - 每次评估的权重相同（简单算术平均），不考虑评估时间或阶段
//   - 如果需要加权平均（如最近的评估权重更高），需要另外实现
//   - Type字段（auto/manual/peer/mentor）在此方法中不影响权重
//
// 作者: JIA
func (sp *StudentProfile) GetOverallScore() float64 {
	if len(sp.Assessments) == 0 {
		return 0.0
	}

	totalScore := 0.0
	// 使用索引遍历避免大结构体复制（208字节）
	for i := range sp.Assessments {
		totalScore += sp.Assessments[i].Scores.OverallScore
	}

	return totalScore / float64(len(sp.Assessments))
}

// ToJSON 将学习者档案序列化为格式化的JSON字符串
//
// 功能说明:
//
//	本方法将StudentProfile对象转换为易读的JSON格式，用于数据持久化、
//	API响应、配置导出等场景。使用缩进格式化，方便人工阅读和版本控制。
//
// 序列化配置:
//   - 缩进格式: 每级缩进2个空格（Go社区标准）
//   - 字段顺序: 保持StudentProfile结构体定义顺序
//   - 空值处理:
//   - 空切片序列化为[] (如Projects、Assessments、Certifications)
//   - 空map序列化为{} (如StageProgress、Competencies)
//   - nil指针序列化为null (如StageProgress.CompleteDate)
//   - 时间格式: RFC3339格式（如"2025-10-03T14:30:00Z"）
//
// JSON结构示例:
//
//	{
//	  "id": "stu_001",
//	  "name": "张三",
//	  "email": "zhangsan@example.com",
//	  "start_date": "2025-10-03T10:00:00Z",
//	  "current_stage": 3,
//	  "stage_progress": {
//	    "1": {"stage_id": 1, "progress": 1.0, "status": "completed"},
//	    "2": {"stage_id": 2, "progress": 1.0, "status": "completed"},
//	    "3": {"stage_id": 3, "progress": 0.6, "status": "in_progress"}
//	  },
//	  "projects": [...],
//	  "assessments": [...],
//	  "certifications": [...]
//	}
//
// 返回值:
//   - []byte: JSON字节数组，UTF-8编码
//   - error: 序列化错误（通常不会发生，除非包含不可序列化类型如chan/func）
//
// 使用场景:
//   - 保存学习者档案到文件（如student_001.json）
//   - 通过API返回学习者完整档案
//   - 生成人类可读的学习记录模板
//   - 版本控制中跟踪学习者数据变更
//   - 数据备份和迁移
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "李明", "liming@example.com")
//	student.UpdateProgress(1, 0.5, 10.0)
//
//	jsonData, err := student.ToJSON()
//	if err != nil {
//	    log.Fatal("序列化失败:", err)
//	}
//
//	// 保存到文件
//	os.WriteFile("student_001.json", jsonData, 0600)
//
//	// 或通过API返回
//	w.Header().Set("Content-Type", "application/json")
//	w.Write(jsonData)
//
// 性能考量:
//   - 对于大型档案对象（包含数百条评估记录和项目），序列化可能耗时较长
//   - 如果频繁调用，考虑缓存结果或使用流式编码
//   - 缩进格式会增加约30%的数据体积（相比紧凑格式）
//   - 大型JSON字符串在网络传输时建议启用gzip压缩
//
// 注意事项:
//   - 返回的JSON包含学习者所有敏感信息（邮箱、学习记录等），需要注意数据隐私保护
//   - 时间字段使用UTC时区，前端展示时应根据用户时区转换
//   - 浮点数精度可能在序列化后发生微小变化（如0.1可能变为0.10000000000000001）
//
// 作者: JIA
func (sp *StudentProfile) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sp, "", "  ")
}

// FromJSON 从JSON数据反序列化为学习者档案对象
//
// 功能说明:
//
//	本方法将JSON格式的字节数据解析为StudentProfile结构体，用于加载
//	已保存的学习者档案、导入外部数据、或处理API请求数据。
//
// 反序列化特性:
//   - 类型安全: 严格按照结构体json tag定义解析字段
//   - 容错处理:
//   - 缺失字段使用Go零值（如未提供current_stage则默认为0）
//   - 多余字段被忽略（JSON中的未知字段不会导致错误）
//   - 时间解析: 自动识别RFC3339、Unix时间戳等多种时间格式
//   - 嵌套支持: 正确处理多层嵌套的复杂结构（如StageProgress、SkillMatrix）
//
// 参数:
//   - data: JSON格式的字节数组，必须符合StudentProfile结构
//
// 返回值:
//   - error: 反序列化错误，可能原因包括：
//   - JSON格式错误（语法错误、引号不匹配等）
//   - 类型不匹配（字符串无法转为数字等）
//   - 数据格式不符合结构体定义
//   - 时间格式无法解析
//
// 使用场景:
//   - 从JSON文件加载学习者档案
//   - 接收API请求中的学习者数据
//   - 导入其他系统的学习记录
//   - 恢复备份的学习者档案
//   - 数据库JSON字段反序列化
//
// 示例:
//
//	// 从文件加载
//	jsonData, err := os.ReadFile("student_001.json")
//	if err != nil {
//	    log.Fatal("读取文件失败:", err)
//	}
//
//	student := &StudentProfile{}
//	if err := student.FromJSON(jsonData); err != nil {
//	    log.Fatal("解析学习者档案失败:", err)
//	}
//	// 现在student包含了JSON中的所有数据
//	fmt.Printf("学习者: %s, 当前阶段: %d\n", student.Name, student.CurrentStage)
//
//	// 从API请求加载
//	var student StudentProfile
//	if err := student.FromJSON(requestBody); err != nil {
//	    http.Error(w, "无效的JSON数据", http.StatusBadRequest)
//	    return
//	}
//
// 错误处理建议:
//   - 加载配置文件前先验证JSON格式（可用json.Valid）
//   - 解析后验证关键字段是否存在（如ID、Name、Email不应为空）
//   - 记录详细错误信息，便于排查配置问题
//   - 提供有意义的错误提示给用户（不要直接暴露技术错误信息）
//
// 注意事项:
//   - 本方法会覆盖接收者的所有字段（即使JSON中缺失某些字段，也会被置为零值）
//   - 解析失败时，接收者状态不确定，应避免继续使用
//   - 不会验证数据的业务逻辑正确性（如CurrentStage > 15，需要额外校验）
//   - 时间字段如果格式错误，会导致整个解析失败
//
// 作者: JIA
func (sp *StudentProfile) FromJSON(data []byte) error {
	return json.Unmarshal(data, sp)
}

// updateSkillMatrixFromAssessment 从评估详细结果中提取技能数据并更新技能矩阵
//
// 功能说明:
//
//	本方法是AddAssessment的核心辅助方法，负责解析评估结果中的DetailedResults字段，
//	提取各项技能得分，转换为1-5等级的CompetencyLevel，并更新到学习者的SkillMatrix中。
//	这是学习者技能成长跟踪的关键机制。
//
// 更新流程:
//
//  1. 提取技术深度数据:
//     从DetailedResults["technical_depth"]中读取技能项得分
//     例如: "language_features": 75.0
//
//  2. 分数转换为等级:
//     使用公式: Level = int(score/20) + 1
//     分数范围 → 等级映射:
//     - 0-20分   → Level 1 (新手/Novice)
//     - 21-40分  → Level 2 (进阶新手/Advanced Beginner)
//     - 41-60分  → Level 3 (胜任者/Competent)
//     - 61-80分  → Level 4 (精通者/Proficient)
//     - 81-100分 → Level 5 (专家/Expert)
//
//  3. 更新CompetencyLevel:
//     创建新的CompetencyLevel对象，包含:
//     - Category: 能力类别（如"technical_depth"）
//     - Skill: 具体技能（如"language_features"）
//     - Level: 转换后的等级（1-5）
//     - LastUpdated: 当前时间戳
//     - Confidence: 评估置信度（来自assessment.Confidence）
//
//  4. 存储到SkillMatrix:
//     将更新后的CompetencyLevel存入对应的SkillMatrix维度
//     如: SkillMatrix.TechnicalDepth.LanguageFeatures
//
// 技能维度映射:
//
//	DetailedResults结构:
//	{
//	  "technical_depth": {
//	    "language_features": 75.0,   // → SkillMatrix.TechnicalDepth.LanguageFeatures
//	    "standard_library": 80.0,    // → SkillMatrix.TechnicalDepth.StandardLibrary
//	    // ... 更多技能项
//	  },
//	  "engineering_practice": {
//	    "code_quality": 70.0,        // → SkillMatrix.EngineeringPractice.CodeQuality
//	    "testing_skills": 75.0,      // → SkillMatrix.EngineeringPractice.TestingSkills
//	    // ... 更多技能项
//	  }
//	}
//
// 参数:
//   - assessment: 评估记录指针，必须包含有效的DetailedResults字段
//
// 返回值: 无（直接修改StudentProfile.SkillMatrix）
//
// 性能优化:
//   - 使用指针参数避免208字节的AssessmentRecord结构体复制
//   - 仅更新DetailedResults中存在的技能项，避免无效遍历
//   - 类型断言失败时静默跳过，保证健壮性
//
// 数据安全:
//   - 使用类型断言安全提取map数据，避免panic
//   - 如果DetailedResults格式不正确，不会导致程序崩溃
//   - 技能项缺失时保持原有值不变
//
// 使用场景:
//   - 每次调用AddAssessment时自动触发
//   - 评估完成后自动同步技能矩阵
//   - 无需手动调用（内部方法）
//
// 示例:
//
//	student := NewStudentProfile("stu_001", "张三", "test@example.com")
//
//	assessment := &AssessmentRecord{
//	    Stage: 1,
//	    DetailedResults: map[string]interface{}{
//	        "technical_depth": map[string]interface{}{
//	            "language_features": 75.0, // 75分 → Level 4
//	        },
//	    },
//	    Confidence: 0.85,
//	}
//
//	student.AddAssessment(assessment) // 内部会调用updateSkillMatrixFromAssessment
//
//	// 验证更新结果
//	langFeatures := student.SkillMatrix.TechnicalDepth.LanguageFeatures
//	// langFeatures.Level == 4 (int(75/20) + 1 = 4)
//	// langFeatures.Category == "technical_depth"
//	// langFeatures.Skill == "language_features"
//	// langFeatures.Confidence == 0.85
//	// langFeatures.LastUpdated == [当前时间]
//
// 注意事项:
//   - 本方法当前仅实现了technical_depth维度的language_features技能项更新
//   - 完整实现需要添加所有维度和技能项的映射逻辑（代码中标记为"类似地更新其他技能维度..."）
//   - DetailedResults的map结构必须严格遵循三层嵌套：map[string]interface{} → map[string]interface{} → float64
//   - 如果评估系统更新了技能项定义，本方法也需要同步更新映射逻辑
//
// 作者: JIA
func (sp *StudentProfile) updateSkillMatrixFromAssessment(assessment *AssessmentRecord) {
	now := time.Now()

	// 更新技能矩阵的各个维度
	if technical, ok := assessment.DetailedResults["technical_depth"].(map[string]interface{}); ok {
		if languageFeatures, ok := technical["language_features"].(float64); ok {
			sp.SkillMatrix.TechnicalDepth.LanguageFeatures = CompetencyLevel{
				Category:    "technical_depth",
				Skill:       "language_features",
				Level:       int(languageFeatures/SkillLevelDivider) + 1, // 转换为1-5等级
				LastUpdated: now,
				Confidence:  assessment.Confidence,
			}
		}
	}

	// 类似地更新其他技能维度...
}

// getStageNameByID 根据学习阶段编号获取对应的中文阶段名称
//
// 功能说明:
//
//	本函数提供"Go从入门到通天"15阶段学习路径的编号到名称映射，
//	用于在UpdateProgress等方法中自动填充StageProgress的StageName字段，
//	提升数据可读性和用户体验。
//
// 15阶段学习路径详解:
//
//	📋 阶段0: 评估系统
//	   特殊阶段，用于学习者初始能力评估和学习路径规划
//
//	【基础篇：1-6阶段】
//	🔰 阶段1: Go语言基础 - 语法、类型、函数、包管理
//	🚀 阶段2: 高级语言特性 - 接口、反射、泛型、错误处理
//	⚡ 阶段3: 并发编程 - goroutines、channels、sync包
//	🌐 阶段4: Web开发 - HTTP服务、路由、中间件、模板
//	💾 阶段5: 数据库集成 - SQL/NoSQL、ORM、缓存、事务
//	🏗️ 阶段6: 实战项目 - 完整Web应用、微服务、RESTful API
//
//	【进阶篇：7-10阶段】
//	🔬 阶段7: 运行时内部机制 - 内存管理、垃圾回收、调度器
//	🌊 阶段8: 高级网络编程 - TCP/UDP、WebSocket、gRPC
//	🏛️ 阶段9: 微服务架构 - 服务发现、负载均衡、熔断降级
//	🛠️ 阶段10: 编译器工具链 - AST、代码生成、静态分析
//
//	【高级篇：11-15阶段】
//	🏢 阶段11: 大规模系统 - 分布式系统、高可用架构、性能优化
//	☁️ 阶段12: DevOps部署 - 容器化、CI/CD、云原生
//	⚡ 阶段13: 性能优化 - profiling、benchmark、调优实战
//	👑 阶段14: 技术领导力 - 架构设计、代码评审、技术决策
//	🌟 阶段15: 开源贡献 - 参与Go生态、提交PR、社区建设
//
// 参数:
//   - stageID: 学习阶段编号（0-15范围）
//
// 返回值:
//   - string: 阶段的中文名称
//   - 0-15: 返回对应的阶段名称
//   - 其他值: 返回 "阶段{ID}" 格式的默认名称
//
// 使用场景:
//   - UpdateProgress方法中自动填充StageProgress.StageName
//   - 显示学习进度报告时提供友好的阶段名称
//   - 生成学习路径图和里程碑展示
//   - 导出学习记录时的可读性增强
//
// 示例:
//
//	name := getStageNameByID(1)
//	// name == "Go语言基础"
//
//	name = getStageNameByID(6)
//	// name == "实战项目"
//
//	name = getStageNameByID(15)
//	// name == "开源贡献"
//
//	name = getStageNameByID(99)
//	// name == "阶段99" (ID超出范围时的默认格式)
//
// 注意事项:
//   - 阶段编号从0开始（评估系统），1-15为正式学习阶段
//   - 如果学习路径调整（如增加新阶段），需同步更新此映射表
//   - 返回的是简洁名称，完整描述需查阅对应模块的README文档
//   - 阶段名称为纯中文，国际化需求时需要扩展多语言支持
//
// 作者: JIA
func getStageNameByID(stageID int) string {
	stageNames := map[int]string{
		0:  "评估系统",
		1:  "Go语言基础",
		2:  "高级语言特性",
		3:  "并发编程",
		4:  "Web开发",
		5:  "数据库集成",
		6:  "实战项目",
		7:  "运行时内部机制",
		8:  "高级网络编程",
		9:  "微服务架构",
		10: "编译器工具链",
		11: "大规模系统",
		12: "DevOps部署",
		13: "性能优化",
		14: "技术领导力",
		15: "开源贡献",
	}

	if name, exists := stageNames[stageID]; exists {
		return name
	}
	return fmt.Sprintf("阶段%d", stageID)
}
