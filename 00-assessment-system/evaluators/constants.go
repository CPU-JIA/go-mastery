/*
=== Go语言学习评估系统 - 评估器常量定义 ===

本文件定义了评估系统中使用的所有常量，避免魔法数字的使用

作者: JIA
创建时间: 2025-10-03
版本: 1.0.0
*/

package evaluators

// 评分常量定义
const (
	// 基础分数常量
	MaxScore                = 100.0 // 最高分数
	PassingScore            = 70.0  // 及格分数
	GoodScore               = 80.0  // 良好分数
	ExcellentScore          = 90.0  // 优秀分数
	MinAcceptableScore      = 60.0  // 最低可接受分数
	DefaultBaseScore        = 100.0 // 默认基础分
	ZeroScore               = 0.0   // 零分
	MinCoverageRequirement  = 80.0  // 最低覆盖率要求
	MinDocumentationScore   = 70.0  // 最低文档分数
	MinPerformanceScore     = 70.0  // 最低性能分数
	OptimalTestCoverageRate = 80.0  // 最佳测试覆盖率

	// 复杂度相关常量
	MaxCyclomaticComplexity = 5.0  // 最大圈复杂度阈值
	MaxFunctionComplexity   = 10.0 // 最大函数复杂度
	RecommendedComplexity   = 5.0  // 推荐复杂度阈值

	// 扣分/加分权重
	ScorePerStyleIssue        = 2.0  // 每个风格问题扣分
	ScorePerComplexityIssue   = 5.0  // 每个复杂度问题扣分
	ScorePerSecurityHighIssue = 15.0 // 每个高危安全问题扣分
	ScorePerSecurityMedIssue  = 8.0  // 每个中危安全问题扣分
	ScorePerPerformanceIssue  = 8.0  // 每个性能问题扣分
	ScorePerAllocationIssue   = 5.0  // 每个内存分配问题扣分
	BonusForReadme            = 20.0 // README加分
	BonusForChangelog         = 15.0 // CHANGELOG加分
	BonusForLicense           = 10.0 // LICENSE加分
	BonusForContributing      = 5.0  // CONTRIBUTING加分

	// 百分比和比率常量
	ReadmeWeightInUsability    = 0.4 // README在可用性中的权重
	DocCoverageWeightInOverall = 0.3 // 文档覆盖率在总分中的权重
	CommentRatioForFullScore   = 400 // 注释率满分系数（25%注释率得满分）
	TestFileRatioThresholdHigh = 0.3 // 测试文件比例高门槛（30%）
	TestFileRatioThresholdLow  = 0.1 // 测试文件比例低门槛（10%）
	HalfComplexityPenalty      = 0.5 // 半复杂度惩罚系数

	// 优先级和分级
	LowPriorityIssueThreshold      = 5   // 低优先级问题阈值
	MediumPriorityIssueThreshold   = 15  // 中优先级问题阈值
	HighPriorityIssueThreshold     = 20  // 高优先级问题阈值
	MinIssueCountForHighPriority   = 3   // 高优先级问题最少数量
	DefaultIssueCountBeforeWarning = 10  // 默认问题数量警告阈值
	MinIssueCountForHotspot        = 3   // 质量热点最少问题数
	SeverityWeightError            = 3.0 // 错误严重程度权重
	SeverityWeightWarning          = 2.0 // 警告严重程度权重
	SeverityWeightInfo             = 1.0 // 信息严重程度权重

	// 数组索引和分隔符相关
	MinGolintFieldCount  = 4 // golint输出最少字段数
	GoModPathFileSize    = 4 // go.mod路径数组大小（用于strconv.Atoi）
	DefaultColumnIfError = 0 // strconv解析错误时的默认列号
	DefaultLineIfError   = 0 // strconv解析错误时的默认行号

	// 评级标准
	GradeAThreshold  = 90.0 // A等级阈值
	GradeBThreshold  = 80.0 // B等级阈值
	GradeCThreshold  = 70.0 // C等级阈值
	GradeDThreshold  = 60.0 // D等级阈值
	GradeFThreshold  = 0.0  // F等级阈值（低于60分）
	GradeAString     = "A"  // A等级字符串
	GradeBString     = "B"  // B等级字符串
	GradeCString     = "C"  // C等级字符串
	GradeDString     = "D"  // D等级字符串
	GradeFString     = "F"  // F等级字符串
	BronzeMinScore   = 70.0 // 铜牌最低分数
	SilverMinScore   = 80.0 // 银牌最低分数
	GoldMinScore     = 85.0 // 金牌最低分数
	PlatinumMinScore = 90.0 // 白金牌最低分数

	// 质量评级标准
	TechnicalDebtRatingA = 1.0  // 技术债务A级阈值（每千行代码债务小时数）
	TechnicalDebtRatingB = 3.0  // 技术债务B级阈值
	TechnicalDebtRatingC = 5.0  // 技术债务C级阈值
	TechnicalDebtRatingD = 10.0 // 技术债务D级阈值
)

// 工具和严重程度级别常量
const (
	// 严重程度级别字符串
	SeverityWarning = "warning" // 警告级别标识
	SeverityError   = "error"   // 错误级别标识
	SeverityInfo    = "info"    // 信息级别标识

	// 工具版本标识
	ToolVersionLatest = "latest" // 工具最新版本标识
)

// 评分标准补充常量
const (
	// 细分评分阈值
	Score60  = 60.0  // 基础及格分
	Score75  = 75.0  // 中等分数
	Score78  = 78.0  // 良好偏下分数
	Score82  = 82.0  // 良好偏上分数
	Score85  = 85.0  // 优秀入门分数
	Score95  = 95.0  // 卓越分数
	Score100 = 100.0 // 满分

	// 特定评估阈值
	DefaultScore       = 80.0 // 默认良好分数
	MediumQualityScore = 75.0 // 中等质量分数
	HighQualityScore   = 85.0 // 高质量分数
	SimpleReturnScore  = 20.0 // 简化返回分数
	PartialScore       = 15.0 // 部分得分
	MinorBonus         = 10.0 // 小额加分
	PenaltyPerIssue    = 2.0  // 每个问题扣分
	MajorPenalty       = 10.0 // 重大扣分
	CriticalPenalty    = 15.0 // 严重扣分
	TechnicalDebtRate  = 15.0 // 技术债务率(%)
)

// 权重系数常量
const (
	// 评分权重系数
	WeightVeryLow    = 0.10 // 10%权重
	WeightLow        = 0.15 // 15%权重
	WeightMediumLow  = 0.20 // 20%权重
	WeightMedium     = 0.25 // 25%权重
	WeightMediumHigh = 0.30 // 30%权重
	WeightHigh       = 0.40 // 40%权重
	WeightVeryHigh   = 0.60 // 60%权重
	WeightCritical   = 0.80 // 80%权重
	WeightAlmostFull = 0.95 // 95%权重

	// 特定维度权重
	ReadmeWeightInDoc     = 0.40 // README在文档中的权重
	DocCoverageWeight     = 0.30 // 文档覆盖率权重
	TestCoverageWeight    = 0.60 // 测试覆盖率权重
	BuildQualityWeight    = 0.30 // 构建质量权重
	CIConfigWeight        = 0.20 // CI配置权重
	TestQualityWeight     = 0.30 // 测试质量权重
	DocOverallScoreWeight = 0.20 // 文档总分权重
)

// 计算因子常量
const (
	// 数值计算因子
	FactorTwo        = 2  // 因子2
	FactorThree      = 3  // 因子3
	FactorFour       = 4  // 因子4
	FactorFive       = 5  // 因子5
	FactorSix        = 6  // 因子6
	FactorTen        = 10 // 因子10
	FactorFifteen    = 15 // 因子15
	FactorTwenty     = 20 // 因子20
	FactorTwentyFive = 25 // 因子25

	// 浮点因子
	FactorTwoFloat   = 2.0  // 浮点因子2.0
	FactorThreeFloat = 3.0  // 浮点因子3.0
	FactorFourFloat  = 4.0  // 浮点因子4.0
	FactorFiveFloat  = 5.0  // 浮点因子5.0
	FactorTenFloat   = 10.0 // 浮点因子10.0
)

// 配置和限制常量
const (
	// 工具和系统限制
	MaxFileSize              = 1048576 // 最大文件大小 (1MB)
	DefaultComplexityLimit   = 10      // 默认复杂度限制
	DefaultFunctionLength    = 50      // 默认函数长度
	DefaultIssueCountWarning = 5       // 默认问题数量警告阈值
	MaxFileSizeBytes         = 1048576 // 文件大小限制字节数

	// 时间和配置值
	DefaultTimeout      = 300 // 默认超时(秒)
	DefaultCacheSeconds = 120 // 默认缓存时间(秒)
	LowPriorityValue    = 6   // 低优先级值
	DefaultCognitiveMax = 15  // 默认认知复杂度上限
	IntValue80          = 80  // 整型80（用于特定配置）

	// 考试时长常量（分钟）
	ExamDurationBronze   = 120 // 铜牌考试时长120分钟
	ExamDurationSilver   = 180 // 银牌考试时长180分钟
	ExamDurationGold     = 240 // 金牌考试时长240分钟
	ExamDurationPlatinum = 300 // 白金牌考试时长300分钟

	// 学习时长常量（小时）
	LearningHours20  = 20  // 学习时长20小时
	LearningHours25  = 25  // 学习时长25小时
	LearningHours35  = 35  // 学习时长35小时
	LearningHours40  = 40  // 学习时长40小时
	LearningHours60  = 60  // 学习时长60小时
	LearningHours80  = 80  // 学习时长80小时
	LearningHours100 = 100 // 学习时长100小时

	// 专业轨道持续时间（月）
	TrackDuration4 = 4 // 专业轨道4个月
	TrackDuration6 = 6 // 专业轨道6个月

	// 工作经验年限
	ExperienceYears1 = 1 // 1年工作经验
	ExperienceYears5 = 5 // 5年工作经验

	// 技能等级常量
	SkillLevel2 = 2 // 技能等级2
	SkillLevel3 = 3 // 技能等级3
	SkillLevel4 = 4 // 技能等级4
	SkillLevel5 = 5 // 技能等级5

	// 薪资常量（美元）
	SalaryJuniorMin    = 60000  // 初级最低薪资
	SalaryJuniorMedian = 70000  // 初级中位薪资
	SalaryJuniorMax    = 80000  // 初级最高薪资
	SalarySeniorMin    = 120000 // 高级最低薪资
	SalarySeniorMedian = 140000 // 高级中位薪资
	SalarySeniorMax    = 160000 // 高级最高薪资
)

// 难度系数常量（浮点数）
const (
	DifficultyEasy     = 2.0 // 简单难度2.0
	DifficultyMedium   = 3.0 // 中等难度3.0
	DifficultyHard     = 4.0 // 困难难度4.0
	DifficultyVeryHard = 4.5 // 非常困难4.5
)

// 代码质量和评估工具常量
const (
	MaxLineLength          = 120  // 最大行长度限制
	TotalLearningStages    = 15.0 // 总学习阶段数（浮点）
	MinTestCoverageRatio   = 0.70 // 最小测试覆盖率要求（70%）
	MinTestCoveragePercent = 70.0 // 最小测试覆盖率百分比
)
