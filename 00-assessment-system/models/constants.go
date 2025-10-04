/*
=== Go语言学习评估系统 - 模型层常量定义 ===

本文件定义了评估系统模型层使用的所有常量，避免魔法数字的使用

作者: JIA
创建时间: 2025-10-03
版本: 1.0.0
*/

package models

// 权重常量定义
const (
	// DefaultWeights - 默认权重值
	WeightTechnicalDepth      = 0.40 // 技术深度权重
	WeightEngineeringPractice = 0.30 // 工程实践权重
	WeightProjectExperience   = 0.20 // 项目经验权重
	WeightSoftSkills          = 0.10 // 软技能权重

	WeightAutomatedAssessment = 0.50 // 自动化评估权重
	WeightCodeReview          = 0.30 // 代码审查权重
	WeightProjectEvaluation   = 0.15 // 项目评估权重
	WeightPeerFeedback        = 0.03 // 同伴反馈权重
	WeightMentorAssessment    = 0.02 // 导师评估权重
)

// 阈值常量定义
const (
	// DefaultThresholds - 默认阈值设定
	ThresholdPassingScore      = 70.0 // 及格分数
	ThresholdExcellentScore    = 90.0 // 优秀分数
	ThresholdMinCoverage       = 80.0 // 最低覆盖率
	ThresholdMaxComplexity     = 10.0 // 最大复杂度
	ThresholdMinDocumentation  = 85.0 // 最低文档分数
	ThresholdPerformanceTarget = 95.0 // 性能目标
)

// 评分等级常量定义
const (
	// GradeThresholds - 评分等级阈值
	GradeAPlusThreshold  = 95.0 // A+ 等级阈值
	GradeAThreshold      = 90.0 // A 等级阈值
	GradeAMinusThreshold = 85.0 // A- 等级阈值
	GradeBPlusThreshold  = 80.0 // B+ 等级阈值
	GradeBThreshold      = 75.0 // B 等级阈值
	GradeBMinusThreshold = 70.0 // B- 等级阈值
	GradeCPlusThreshold  = 65.0 // C+ 等级阈值
	GradeCThreshold      = 60.0 // C 等级阈值
)

// 技能等级计算常量
const (
	SkillLevelDivider = 20 // 技能等级计算除数 (分数/20 + 1 = 等级)
)

// 学习偏好默认值常量
const (
	DefaultAvailableHours = 10.0 // 默认每周可用学习小时数
)
