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
	Status      string      `json:"status"`      // 状态: active, completed, cancelled
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

// NewStudentProfile 创建新的学习者档案
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
			AvailableHours:      10.0,
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

// UpdateProgress 更新学习进度
func (sp *StudentProfile) UpdateProgress(stage int, progress float64, hoursSpent float64) {
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

// AddProject 添加项目记录
func (sp *StudentProfile) AddProject(project ProjectRecord) {
	project.ID = fmt.Sprintf("proj_%d_%s", len(sp.Projects)+1, project.Name)
	sp.Projects = append(sp.Projects, project)
	sp.LastActive = time.Now()
}

// AddAssessment 添加评估记录
func (sp *StudentProfile) AddAssessment(assessment AssessmentRecord) {
	assessment.ID = fmt.Sprintf("assess_%d_%d", len(sp.Assessments)+1, assessment.Stage)
	sp.Assessments = append(sp.Assessments, assessment)
	sp.LastActive = time.Now()

	// 更新技能矩阵
	sp.updateSkillMatrixFromAssessment(assessment)
}

// AddCertification 添加认证记录
func (sp *StudentProfile) AddCertification(cert CertificationRecord) {
	sp.Certifications = append(sp.Certifications, cert)
	sp.LastActive = time.Now()
}

// GetCurrentLevel 获取当前认证等级
func (sp *StudentProfile) GetCurrentLevel() string {
	if len(sp.Certifications) == 0 {
		return "None"
	}

	// 找到最高等级的有效认证
	levels := []string{"Bronze", "Silver", "Gold", "Platinum"}
	currentLevel := 0

	for _, cert := range sp.Certifications {
		if cert.Status == "active" {
			for i, level := range levels {
				if cert.Level == level && i > currentLevel {
					currentLevel = i
				}
			}
		}
	}

	return levels[currentLevel]
}

// GetOverallScore 计算综合评分
func (sp *StudentProfile) GetOverallScore() float64 {
	if len(sp.Assessments) == 0 {
		return 0.0
	}

	totalScore := 0.0
	for _, assessment := range sp.Assessments {
		totalScore += assessment.Scores.OverallScore
	}

	return totalScore / float64(len(sp.Assessments))
}

// ToJSON 序列化为JSON
func (sp *StudentProfile) ToJSON() ([]byte, error) {
	return json.MarshalIndent(sp, "", "  ")
}

// FromJSON 从JSON反序列化
func (sp *StudentProfile) FromJSON(data []byte) error {
	return json.Unmarshal(data, sp)
}

// updateSkillMatrixFromAssessment 从评估结果更新技能矩阵
func (sp *StudentProfile) updateSkillMatrixFromAssessment(assessment AssessmentRecord) {
	now := time.Now()

	// 更新技能矩阵的各个维度
	if technical, ok := assessment.DetailedResults["technical_depth"].(map[string]interface{}); ok {
		if languageFeatures, ok := technical["language_features"].(float64); ok {
			sp.SkillMatrix.TechnicalDepth.LanguageFeatures = CompetencyLevel{
				Category:    "technical_depth",
				Skill:       "language_features",
				Level:       int(languageFeatures/20) + 1, // 转换为1-5等级
				LastUpdated: now,
				Confidence:  assessment.Confidence,
			}
		}
	}

	// 类似地更新其他技能维度...
}

// getStageNameByID 根据阶段ID获取阶段名称
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
