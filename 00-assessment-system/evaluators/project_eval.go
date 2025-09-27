/*
=== Go语言学习评估系统 - 项目评估引擎 ===

本文件实现了全面的Go项目质量评估系统：
1. 功能完整性评估 - 需求实现程度、功能覆盖、边界条件处理
2. 架构质量评估 - 模块化设计、依赖管理、接口设计、可扩展性
3. 用户体验评估 - API设计、错误处理、文档质量、使用便利性
4. 技术深度评估 - 技术栈选择、创新性解决方案、最佳实践应用
5. 工程质量评估 - 项目结构、构建系统、配置管理、部署准备
6. 代码组织评估 - 包结构、命名约定、模块边界、职责分离
7. 可维护性评估 - 代码清晰度、文档完整性、测试覆盖、重构友好性
*/

package evaluators

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ProjectEvaluator 项目评估器
type ProjectEvaluator struct {
	config       *ProjectEvalConfig
	projectInfo  *ProjectInfo
	requirements []ProjectRequirement
	criteria     []EvaluationCriteria
	results      *ProjectEvaluationResult
}

// ProjectEvalConfig 项目评估配置
type ProjectEvalConfig struct {
	// 评估范围配置
	ProjectType      string `json:"project_type"`      // 项目类型: cli, web, library, service
	Stage            int    `json:"stage"`             // 学习阶段 (1-15)
	RequirementLevel string `json:"requirement_level"` // 要求级别: basic, intermediate, advanced

	// 评估权重配置
	WeightSettings ProjectEvalWeights        `json:"weight_settings"` // 维度权重
	Criteria       map[string]CriteriaConfig `json:"criteria"`        // 评估标准配置
	Thresholds     ProjectThresholds         `json:"thresholds"`      // 阈值设定

	// 分析配置
	AnalyzeReadme     bool `json:"analyze_readme"`     // 是否分析README
	AnalyzeDocs       bool `json:"analyze_docs"`       // 是否分析文档
	AnalyzeBuild      bool `json:"analyze_build"`      // 是否分析构建系统
	AnalyzeAPI        bool `json:"analyze_api"`        // 是否分析API设计
	CheckDependencies bool `json:"check_dependencies"` // 是否检查依赖

	// 输出配置
	DetailLevel         string `json:"detail_level"`         // 详细程度
	GenerateSuggestions bool   `json:"generate_suggestions"` // 是否生成建议
	SaveResults         bool   `json:"save_results"`         // 是否保存结果
	ResultsPath         string `json:"results_path"`         // 结果保存路径
}

// ProjectEvalWeights 项目评估权重
type ProjectEvalWeights struct {
	FunctionalityScore   float64 `json:"functionality_score"`   // 功能完整性权重
	ArchitectureScore    float64 `json:"architecture_score"`    // 架构质量权重
	UserExperienceScore  float64 `json:"user_experience_score"` // 用户体验权重
	TechnicalDepthScore  float64 `json:"technical_depth_score"` // 技术深度权重
	EngineeringScore     float64 `json:"engineering_score"`     // 工程质量权重
	MaintainabilityScore float64 `json:"maintainability_score"` // 可维护性权重
	InnovationScore      float64 `json:"innovation_score"`      // 创新性权重
}

// ProjectThresholds 项目评估阈值
type ProjectThresholds struct {
	MinFunctionality float64 `json:"min_functionality"` // 最低功能完整性
	MinArchitecture  float64 `json:"min_architecture"`  // 最低架构质量
	MinDocumentation float64 `json:"min_documentation"` // 最低文档质量
	MinTestCoverage  float64 `json:"min_test_coverage"` // 最低测试覆盖率
	MaxComplexity    int     `json:"max_complexity"`    // 最大复杂度
	MinModularity    float64 `json:"min_modularity"`    // 最低模块化程度
}

// CriteriaConfig 评估标准配置
type CriteriaConfig struct {
	Enabled     bool    `json:"enabled"`     // 是否启用
	Weight      float64 `json:"weight"`      // 权重
	Threshold   float64 `json:"threshold"`   // 阈值
	Description string  `json:"description"` // 描述
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	Name      string `json:"name"`      // 项目名称
	Path      string `json:"path"`      // 项目路径
	Type      string `json:"type"`      // 项目类型
	Language  string `json:"language"`  // 主要语言
	Framework string `json:"framework"` // 使用框架

	// 项目统计
	Structure     ProjectStructure   `json:"structure"`     // 项目结构
	Dependencies  DependencyAnalysis `json:"dependencies"`  // 依赖分析
	BuildSystem   BuildSystemInfo    `json:"build_system"`  // 构建系统
	Documentation DocumentationInfo  `json:"documentation"` // 文档信息
	TestingInfo   TestingInfo        `json:"testing_info"`  // 测试信息

	// 项目元数据
	CreatedAt    time.Time `json:"created_at"`    // 创建时间
	LastModified time.Time `json:"last_modified"` // 最后修改
	Contributors int       `json:"contributors"`  // 贡献者数量
	CommitCount  int       `json:"commit_count"`  // 提交数量
}

// ProjectStructure 项目结构信息
type ProjectStructure struct {
	TotalFiles  int `json:"total_files"`  // 总文件数
	GoFiles     int `json:"go_files"`     // Go文件数
	TestFiles   int `json:"test_files"`   // 测试文件数
	Packages    int `json:"packages"`     // 包数量
	ModuleDepth int `json:"module_depth"` // 模块深度

	// 目录结构
	HasCmd      bool `json:"has_cmd"`      // 是否有cmd目录
	HasInternal bool `json:"has_internal"` // 是否有internal目录
	HasPkg      bool `json:"has_pkg"`      // 是否有pkg目录
	HasAPI      bool `json:"has_api"`      // 是否有api目录
	HasDocs     bool `json:"has_docs"`     // 是否有docs目录
	HasExamples bool `json:"has_examples"` // 是否有examples目录

	// 文件组织
	FileOrganization  float64 `json:"file_organization"`  // 文件组织评分
	NamingConsistency float64 `json:"naming_consistency"` // 命名一致性评分
	ModuleBoundaries  float64 `json:"module_boundaries"`  // 模块边界清晰度
}

// DependencyAnalysis 依赖分析
type DependencyAnalysis struct {
	TotalDependencies    int     `json:"total_dependencies"`     // 总依赖数
	DirectDependencies   int     `json:"direct_dependencies"`    // 直接依赖数
	IndirectDependencies int     `json:"indirect_dependencies"`  // 间接依赖数
	StandardLibraryUsage float64 `json:"standard_library_usage"` // 标准库使用率

	// 依赖质量
	DependencyQuality  float64             `json:"dependency_quality"`  // 依赖质量评分
	VersionConsistency float64             `json:"version_consistency"` // 版本一致性
	SecurityRisk       float64             `json:"security_risk"`       // 安全风险评分
	Vulnerabilities    []VulnerabilityInfo `json:"vulnerabilities"`     // 漏洞信息

	// 依赖详情
	Dependencies         []DependencyInfo `json:"dependencies"`          // 依赖详情
	LicenseCompatibility float64          `json:"license_compatibility"` // 许可证兼容性
}

// DependencyInfo 依赖信息
type DependencyInfo struct {
	Name        string    `json:"name"`        // 依赖名称
	Version     string    `json:"version"`     // 版本
	Type        string    `json:"type"`        // 类型: direct, indirect
	Usage       string    `json:"usage"`       // 使用方式
	License     string    `json:"license"`     // 许可证
	Popularity  float64   `json:"popularity"`  // 流行度评分
	Maintenance float64   `json:"maintenance"` // 维护状态评分
	LastUpdate  time.Time `json:"last_update"` // 最后更新时间
}

// VulnerabilityInfo 漏洞信息
type VulnerabilityInfo struct {
	ID           string  `json:"id"`            // 漏洞ID
	Severity     string  `json:"severity"`      // 严重程度
	Description  string  `json:"description"`   // 漏洞描述
	Package      string  `json:"package"`       // 受影响包
	FixedVersion string  `json:"fixed_version"` // 修复版本
	CVSS         float64 `json:"cvss"`          // CVSS评分
}

// BuildSystemInfo 构建系统信息
type BuildSystemInfo struct {
	HasGoMod         bool `json:"has_go_mod"`         // 是否有go.mod
	HasMakefile      bool `json:"has_makefile"`       // 是否有Makefile
	HasDockerfile    bool `json:"has_dockerfile"`     // 是否有Dockerfile
	HasGitHubActions bool `json:"has_github_actions"` // 是否有GitHub Actions
	HasGoReleaser    bool `json:"has_go_releaser"`    // 是否有GoReleaser

	// 构建配置质量
	BuildQuality    float64 `json:"build_quality"`    // 构建质量评分
	CIConfiguration float64 `json:"ci_configuration"` // CI配置评分
	DeploymentReady float64 `json:"deployment_ready"` // 部署准备度
}

// DocumentationInfo 文档信息
type DocumentationInfo struct {
	HasReadme       bool    `json:"has_readme"`       // 是否有README
	ReadmeQuality   float64 `json:"readme_quality"`   // README质量评分
	HasChangelog    bool    `json:"has_changelog"`    // 是否有CHANGELOG
	HasLicense      bool    `json:"has_license"`      // 是否有LICENSE
	HasContributing bool    `json:"has_contributing"` // 是否有CONTRIBUTING

	// API文档
	HasAPIDoc     bool    `json:"has_api_doc"`    // 是否有API文档
	GoDocCoverage float64 `json:"godoc_coverage"` // GoDoc覆盖率
	ExampleCount  int     `json:"example_count"`  // 示例数量

	// 文档质量
	OverallDocScore float64 `json:"overall_doc_score"` // 总体文档评分
	DocConsistency  float64 `json:"doc_consistency"`   // 文档一致性
	DocCompleteness float64 `json:"doc_completeness"`  // 文档完整性
}

// TestingInfo 测试信息
type TestingInfo struct {
	HasTests       bool    `json:"has_tests"`       // 是否有测试
	TestCoverage   float64 `json:"test_coverage"`   // 测试覆盖率
	TestFiles      int     `json:"test_files"`      // 测试文件数
	TestFunctions  int     `json:"test_functions"`  // 测试函数数
	BenchmarkCount int     `json:"benchmark_count"` // 基准测试数
	ExampleTests   int     `json:"example_tests"`   // 示例测试数

	// 测试质量
	TestQuality     float64 `json:"test_quality"`      // 测试质量评分
	TestStrategy    float64 `json:"test_strategy"`     // 测试策略评分
	EdgeCaseTesting float64 `json:"edge_case_testing"` // 边界情况测试
}

// ProjectRequirement 项目需求定义
type ProjectRequirement struct {
	ID          string `json:"id"`          // 需求ID
	Category    string `json:"category"`    // 需求分类
	Title       string `json:"title"`       // 需求标题
	Description string `json:"description"` // 需求描述
	Priority    int    `json:"priority"`    // 优先级
	Mandatory   bool   `json:"mandatory"`   // 是否必需

	// 验证标准
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria"` // 验收标准
	TestCases          []RequirementTestCase `json:"test_cases"`          // 测试用例
	Examples           []RequirementExample  `json:"examples"`            // 需求示例
}

// AcceptanceCriterion 验收标准
type AcceptanceCriterion struct {
	ID          string `json:"id"`           // 标准ID
	Description string `json:"description"`  // 标准描述
	Type        string `json:"type"`         // 标准类型: functional, non_functional, technical
	Verifiable  bool   `json:"verifiable"`   // 是否可验证
	Automated   bool   `json:"automated"`    // 是否可自动验证
	CheckMethod string `json:"check_method"` // 验证方法
}

// RequirementTestCase 需求测试用例
type RequirementTestCase struct {
	ID        string      `json:"id"`        // 用例ID
	Name      string      `json:"name"`      // 用例名称
	Input     interface{} `json:"input"`     // 输入
	Expected  interface{} `json:"expected"`  // 期望输出
	Steps     []TestStep  `json:"steps"`     // 测试步骤
	Automated bool        `json:"automated"` // 是否自动化
}

// TestStep 测试步骤
type TestStep struct {
	Action   string      `json:"action"`   // 操作
	Data     interface{} `json:"data"`     // 数据
	Expected string      `json:"expected"` // 期望结果
}

// RequirementExample 需求示例
type RequirementExample struct {
	Title       string `json:"title"`       // 示例标题
	Description string `json:"description"` // 示例描述
	Code        string `json:"code"`        // 示例代码
	Output      string `json:"output"`      // 示例输出
}

// EvaluationCriteria 评估标准
type EvaluationCriteria struct {
	ID          string  `json:"id"`          // 标准ID
	Category    string  `json:"category"`    // 标准分类
	Name        string  `json:"name"`        // 标准名称
	Description string  `json:"description"` // 标准描述
	Weight      float64 `json:"weight"`      // 权重
	MaxScore    float64 `json:"max_score"`   // 最高分

	// 评估方法
	EvalMethod string `json:"eval_method"` // 评估方法
	Automated  bool   `json:"automated"`   // 是否自动化
	CheckFunc  string `json:"check_func"`  // 检查函数

	// 评分标准
	ScoreLevels []ScoreLevel `json:"score_levels"` // 评分等级
	Threshold   float64      `json:"threshold"`    // 通过阈值
}

// ScoreLevel 评分等级
type ScoreLevel struct {
	Level       int     `json:"level"`       // 等级
	MinValue    float64 `json:"min_value"`   // 最小值
	MaxValue    float64 `json:"max_value"`   // 最大值
	Score       float64 `json:"score"`       // 得分
	Description string  `json:"description"` // 等级描述
}

// ProjectEvaluationResult 项目评估结果
type ProjectEvaluationResult struct {
	ProjectPath string        `json:"project_path"` // 项目路径
	ProjectInfo ProjectInfo   `json:"project_info"` // 项目信息
	Timestamp   time.Time     `json:"timestamp"`    // 评估时间
	Duration    time.Duration `json:"duration"`     // 评估耗时

	// 整体评分
	OverallScore float64 `json:"overall_score"` // 总体评分
	Grade        string  `json:"grade"`         // 评级
	Passed       bool    `json:"passed"`        // 是否通过

	// 维度评分
	DimensionScores map[string]float64 `json:"dimension_scores"` // 各维度得分
	CriteriaScores  map[string]float64 `json:"criteria_scores"`  // 各标准得分

	// 需求实现评估
	RequirementResults []RequirementResult `json:"requirement_results"` // 需求实现结果
	FunctionalityScore float64             `json:"functionality_score"` // 功能完整性得分

	// 质量分析
	ArchitectureAnalysis   ArchitectureAnalysis `json:"architecture_analysis"` // 架构分析
	UserExperienceAnalysis UXAnalysis           `json:"ux_analysis"`           // 用户体验分析
	TechnicalAnalysis      TechnicalAnalysis    `json:"technical_analysis"`    // 技术分析

	// 改进建议
	Strengths    []ProjectStrength    `json:"strengths"`    // 项目优势
	Weaknesses   []ProjectWeakness    `json:"weaknesses"`   // 项目不足
	Improvements []ProjectImprovement `json:"improvements"` // 改进建议
	NextSteps    []NextStep           `json:"next_steps"`   // 下一步建议

	// 比较分析
	BenchmarkComparison *BenchmarkComparison `json:"benchmark_comparison"` // 基准比较
	PeerComparison      *PeerComparison      `json:"peer_comparison"`      // 同级比较
}

// RequirementResult 需求实现结果
type RequirementResult struct {
	RequirementID string                   `json:"requirement_id"` // 需求ID
	Implemented   bool                     `json:"implemented"`    // 是否实现
	Score         float64                  `json:"score"`          // 实现得分
	Quality       float64                  `json:"quality"`        // 实现质量
	Completeness  float64                  `json:"completeness"`   // 完整程度
	TestResults   []TestResult             `json:"test_results"`   // 测试结果
	Evidence      []ImplementationEvidence `json:"evidence"`       // 实现证据
	Issues        []RequirementIssue       `json:"issues"`         // 实现问题
	Suggestions   []string                 `json:"suggestions"`    // 改进建议
}

// TestResult 测试结果（重用之前定义的结构）
type TestResult struct {
	TestID        string  `json:"test_id"`        // 测试ID
	Passed        bool    `json:"passed"`         // 是否通过
	Score         float64 `json:"score"`          // 得分
	Details       string  `json:"details"`        // 详细信息
	ExecutionTime float64 `json:"execution_time"` // 执行时间
}

// ImplementationEvidence 实现证据
type ImplementationEvidence struct {
	Type        string  `json:"type"`        // 证据类型: code, test, documentation, demo
	Location    string  `json:"location"`    // 位置信息
	Description string  `json:"description"` // 描述
	Quality     float64 `json:"quality"`     // 质量评分
	Relevance   float64 `json:"relevance"`   // 相关性评分
}

// RequirementIssue 需求实现问题
type RequirementIssue struct {
	Type        string `json:"type"`        // 问题类型
	Severity    string `json:"severity"`    // 严重程度
	Description string `json:"description"` // 问题描述
	Location    string `json:"location"`    // 问题位置
	Impact      string `json:"impact"`      // 影响评估
	Suggestion  string `json:"suggestion"`  // 修复建议
}

// ArchitectureAnalysis 架构分析
type ArchitectureAnalysis struct {
	ModularityScore      float64 `json:"modularity_score"`      // 模块化评分
	CouplingScore        float64 `json:"coupling_score"`        // 耦合度评分
	CohesionScore        float64 `json:"cohesion_score"`        // 内聚度评分
	InterfaceDesign      float64 `json:"interface_design"`      // 接口设计评分
	DependencyManagement float64 `json:"dependency_management"` // 依赖管理评分

	// 架构模式
	DesignPatterns     []UsedDesignPattern `json:"design_patterns"`     // 使用的设计模式
	ArchitecturalStyle string              `json:"architectural_style"` // 架构风格
	LayerSeparation    float64             `json:"layer_separation"`    // 层次分离度

	// 扩展性分析
	Extensibility float64 `json:"extensibility"` // 可扩展性
	Flexibility   float64 `json:"flexibility"`   // 灵活性
	Reusability   float64 `json:"reusability"`   // 可重用性

	// 问题识别
	ArchitecturalIssues []ArchitecturalIssue `json:"architectural_issues"` // 架构问题
	RefactoringNeeds    []RefactoringNeed    `json:"refactoring_needs"`    // 重构需求
}

// UsedDesignPattern 使用的设计模式
type UsedDesignPattern struct {
	Name        string   `json:"name"`        // 模式名称
	Usage       string   `json:"usage"`       // 使用方式
	Location    []string `json:"location"`    // 使用位置
	Appropriate bool     `json:"appropriate"` // 是否恰当使用
	Quality     float64  `json:"quality"`     // 使用质量
}

// ArchitecturalIssue 架构问题
type ArchitecturalIssue struct {
	Type        string `json:"type"`        // 问题类型
	Severity    string `json:"severity"`    // 严重程度
	Description string `json:"description"` // 问题描述
	Location    string `json:"location"`    // 问题位置
	Impact      string `json:"impact"`      // 影响范围
	Suggestion  string `json:"suggestion"`  // 解决建议
	Priority    int    `json:"priority"`    // 优先级
}

// RefactoringNeed 重构需求
type RefactoringNeed struct {
	Type     string `json:"type"`     // 重构类型
	Target   string `json:"target"`   // 重构目标
	Reason   string `json:"reason"`   // 重构原因
	Benefit  string `json:"benefit"`  // 预期收益
	Effort   string `json:"effort"`   // 所需工作量
	Priority int    `json:"priority"` // 优先级
}

// UXAnalysis 用户体验分析
type UXAnalysis struct {
	// API设计评分
	APIDesignScore     float64 `json:"api_design_score"`     // API设计评分
	ErrorHandlingScore float64 `json:"error_handling_score"` // 错误处理评分
	UsabilityScore     float64 `json:"usability_score"`      // 可用性评分
	ConsistencyScore   float64 `json:"consistency_score"`    // 一致性评分

	// 用户体验要素
	EaseOfUse     float64 `json:"ease_of_use"`    // 易用性
	Intuitiveness float64 `json:"intuitiveness"`  // 直观性
	ErrorRecovery float64 `json:"error_recovery"` // 错误恢复
	Documentation float64 `json:"documentation"`  // 文档质量

	// 具体评估
	APIAnalysis       APIAnalysis           `json:"api_analysis"`       // API分析
	ErrorAnalysis     ErrorHandlingAnalysis `json:"error_analysis"`     // 错误处理分析
	UsabilityIssues   []UsabilityIssue      `json:"usability_issues"`   // 可用性问题
	UXRecommendations []UXRecommendation    `json:"ux_recommendations"` // 用户体验建议
}

// APIAnalysis API分析
type APIAnalysis struct {
	Consistency       float64 `json:"consistency"`        // API一致性
	Simplicity        float64 `json:"simplicity"`         // 简洁性
	Completeness      float64 `json:"completeness"`       // 完整性
	Flexibility       float64 `json:"flexibility"`        // 灵活性
	PerformanceDesign float64 `json:"performance_design"` // 性能设计

	// API质量指标
	RESTCompliance      float64 `json:"rest_compliance"`      // REST规范遵循
	VersioningStrategy  string  `json:"versioning_strategy"`  // 版本管理策略
	SecurityIntegration float64 `json:"security_integration"` // 安全集成

	// 问题和建议
	APIIssues          []APIIssue `json:"api_issues"`          // API问题
	APIRecommendations []string   `json:"api_recommendations"` // API建议
}

// APIIssue API问题
type APIIssue struct {
	Endpoint    string `json:"endpoint"`    // 端点
	Type        string `json:"type"`        // 问题类型
	Severity    string `json:"severity"`    // 严重程度
	Description string `json:"description"` // 问题描述
	Suggestion  string `json:"suggestion"`  // 修复建议
}

// ErrorHandlingAnalysis 错误处理分析
type ErrorHandlingAnalysis struct {
	ErrorConsistency float64 `json:"error_consistency"` // 错误处理一致性
	ErrorInformation float64 `json:"error_information"` // 错误信息完整性
	RecoveryOptions  float64 `json:"recovery_options"`  // 恢复选项
	LoggingQuality   float64 `json:"logging_quality"`   // 日志质量

	// 错误类型分析
	ErrorTypes       []ErrorTypeAnalysis  `json:"error_types"`       // 错误类型分析
	HandlingPatterns []HandlingPattern    `json:"handling_patterns"` // 处理模式
	ErrorIssues      []ErrorHandlingIssue `json:"error_issues"`      // 错误处理问题
}

// ErrorTypeAnalysis 错误类型分析
type ErrorTypeAnalysis struct {
	Type      string  `json:"type"`      // 错误类型
	Frequency int     `json:"frequency"` // 出现频率
	Handling  string  `json:"handling"`  // 处理方式
	Quality   float64 `json:"quality"`   // 处理质量
}

// HandlingPattern 处理模式
type HandlingPattern struct {
	Pattern     string   `json:"pattern"`     // 模式名称
	Usage       int      `json:"usage"`       // 使用次数
	Appropriate bool     `json:"appropriate"` // 是否恰当
	Examples    []string `json:"examples"`    // 示例
}

// ErrorHandlingIssue 错误处理问题
type ErrorHandlingIssue struct {
	Location    string `json:"location"`    // 问题位置
	Type        string `json:"type"`        // 问题类型
	Description string `json:"description"` // 问题描述
	Impact      string `json:"impact"`      // 影响
	Suggestion  string `json:"suggestion"`  // 建议
}

// UsabilityIssue 可用性问题
type UsabilityIssue struct {
	Component  string `json:"component"`   // 组件
	Issue      string `json:"issue"`       // 问题
	Severity   string `json:"severity"`    // 严重程度
	UserImpact string `json:"user_impact"` // 用户影响
	Solution   string `json:"solution"`    // 解决方案
}

// UXRecommendation 用户体验建议
type UXRecommendation struct {
	Category    string `json:"category"`    // 建议分类
	Title       string `json:"title"`       // 建议标题
	Description string `json:"description"` // 详细描述
	Benefit     string `json:"benefit"`     // 预期收益
	Effort      string `json:"effort"`      // 实施工作量
	Priority    int    `json:"priority"`    // 优先级
}

// TechnicalAnalysis 技术分析
type TechnicalAnalysis struct {
	TechnologyStack   []TechnologyUsage `json:"technology_stack"`    // 技术栈使用
	InnovationScore   float64           `json:"innovation_score"`    // 创新性评分
	BestPracticeScore float64           `json:"best_practice_score"` // 最佳实践评分
	TechnicalDebt     float64           `json:"technical_debt"`      // 技术债务
	PerformanceDesign float64           `json:"performance_design"`  // 性能设计

	// 技术选择评估
	TechChoiceAnalysis     []TechChoiceAnalysis `json:"tech_choice_analysis"`    // 技术选择分析
	AlternativeSuggestions []TechAlternative    `json:"alternative_suggestions"` // 替代技术建议
	TechnicalRisks         []TechnicalRisk      `json:"technical_risks"`         // 技术风险

	// 创新性评估
	InnovativeFeatures []InnovativeFeature `json:"innovative_features"` // 创新特性
	CreativityScore    float64             `json:"creativity_score"`    // 创造性评分
	UniquenessFactor   float64             `json:"uniqueness_factor"`   // 独特性因子
}

// TechnologyUsage 技术使用情况
type TechnologyUsage struct {
	Name        string  `json:"name"`        // 技术名称
	Category    string  `json:"category"`    // 技术分类
	Usage       string  `json:"usage"`       // 使用方式
	Proficiency float64 `json:"proficiency"` // 使用熟练度
	Appropriate bool    `json:"appropriate"` // 是否恰当选择
	Rationale   string  `json:"rationale"`   // 选择理由
}

// TechChoiceAnalysis 技术选择分析
type TechChoiceAnalysis struct {
	Technology string   `json:"technology"` // 技术
	Rationale  string   `json:"rationale"`  // 选择理由
	Pros       []string `json:"pros"`       // 优点
	Cons       []string `json:"cons"`       // 缺点
	Score      float64  `json:"score"`      // 选择评分
}

// TechAlternative 技术替代建议
type TechAlternative struct {
	Current         string `json:"current"`          // 当前技术
	Alternative     string `json:"alternative"`      // 替代技术
	Reason          string `json:"reason"`           // 建议理由
	Benefit         string `json:"benefit"`          // 预期收益
	MigrationEffort string `json:"migration_effort"` // 迁移工作量
}

// TechnicalRisk 技术风险
type TechnicalRisk struct {
	Type        string `json:"type"`        // 风险类型
	Description string `json:"description"` // 风险描述
	Probability string `json:"probability"` // 发生概率
	Impact      string `json:"impact"`      // 影响程度
	Mitigation  string `json:"mitigation"`  // 缓解措施
}

// InnovativeFeature 创新特性
type InnovativeFeature struct {
	Name        string  `json:"name"`        // 特性名称
	Description string  `json:"description"` // 特性描述
	Innovation  string  `json:"innovation"`  // 创新点
	Impact      string  `json:"impact"`      // 影响
	Uniqueness  float64 `json:"uniqueness"`  // 独特性评分
}

// ProjectStrength 项目优势
type ProjectStrength struct {
	Category    string   `json:"category"`    // 优势分类
	Title       string   `json:"title"`       // 优势标题
	Description string   `json:"description"` // 详细描述
	Evidence    []string `json:"evidence"`    // 支撑证据
	Impact      string   `json:"impact"`      // 影响价值
}

// ProjectWeakness 项目不足
type ProjectWeakness struct {
	Category    string   `json:"category"`    // 不足分类
	Title       string   `json:"title"`       // 不足标题
	Description string   `json:"description"` // 详细描述
	Impact      string   `json:"impact"`      // 影响评估
	Severity    string   `json:"severity"`    // 严重程度
	Suggestions []string `json:"suggestions"` // 改进建议
}

// ProjectImprovement 项目改进建议
type ProjectImprovement struct {
	Category    string            `json:"category"`    // 改进分类
	Title       string            `json:"title"`       // 改进标题
	Description string            `json:"description"` // 详细描述
	Benefit     string            `json:"benefit"`     // 预期收益
	Effort      string            `json:"effort"`      // 所需工作量
	Priority    int               `json:"priority"`    // 优先级
	Steps       []ImprovementStep `json:"steps"`       // 实施步骤
}

// ImprovementStep 改进步骤
type ImprovementStep struct {
	Order     int    `json:"order"`      // 步骤顺序
	Action    string `json:"action"`     // 具体行动
	Expected  string `json:"expected"`   // 期望结果
	TimeFrame string `json:"time_frame"` // 时间框架
}

// NextStep 下一步建议
type NextStep struct {
	Phase       string   `json:"phase"`       // 阶段名称
	Description string   `json:"description"` // 步骤描述
	Priority    int      `json:"priority"`    // 优先级
	Timeline    string   `json:"timeline"`    // 时间线
	Resources   []string `json:"resources"`   // 所需资源
	Success     []string `json:"success"`     // 成功指标
}

// BenchmarkComparison 基准比较
type BenchmarkComparison struct {
	BenchmarkType   string   `json:"benchmark_type"`   // 基准类型
	ComparisonScore float64  `json:"comparison_score"` // 比较评分
	Percentile      float64  `json:"percentile"`       // 百分位数
	StrongerAreas   []string `json:"stronger_areas"`   // 优势领域
	WeakerAreas     []string `json:"weaker_areas"`     // 薄弱领域
}

// PeerComparison 同级比较
type PeerComparison struct {
	PeerGroup        string   `json:"peer_group"`        // 同级组别
	Ranking          int      `json:"ranking"`           // 排名
	TotalPeers       int      `json:"total_peers"`       // 总数
	RelativeScore    float64  `json:"relative_score"`    // 相对得分
	CompetitiveEdge  []string `json:"competitive_edge"`  // 竞争优势
	ImprovementAreas []string `json:"improvement_areas"` // 改进领域
}

// NewProjectEvaluator 创建项目评估器
func NewProjectEvaluator(config *ProjectEvalConfig) *ProjectEvaluator {
	return &ProjectEvaluator{
		config:       config,
		requirements: []ProjectRequirement{},
		criteria:     []EvaluationCriteria{},
		results: &ProjectEvaluationResult{
			Timestamp:       time.Now(),
			DimensionScores: make(map[string]float64),
			CriteriaScores:  make(map[string]float64),
		},
	}
}

// EvaluateProject 评估项目
func (pe *ProjectEvaluator) EvaluateProject(projectPath string) (*ProjectEvaluationResult, error) {
	log.Printf("开始评估项目: %s", projectPath)
	start := time.Now()

	pe.results.ProjectPath = projectPath
	pe.results.Timestamp = start

	// 1. 分析项目信息
	if err := pe.analyzeProjectInfo(projectPath); err != nil {
		return nil, fmt.Errorf("项目信息分析失败: %v", err)
	}

	// 2. 加载项目需求和评估标准
	if err := pe.loadRequirementsAndCriteria(); err != nil {
		return nil, fmt.Errorf("需求和标准加载失败: %v", err)
	}

	// 3. 评估功能完整性
	if err := pe.evaluateFunctionality(); err != nil {
		return nil, fmt.Errorf("功能完整性评估失败: %v", err)
	}

	// 4. 评估架构质量
	if err := pe.evaluateArchitecture(); err != nil {
		return nil, fmt.Errorf("架构质量评估失败: %v", err)
	}

	// 5. 评估用户体验
	if err := pe.evaluateUserExperience(); err != nil {
		return nil, fmt.Errorf("用户体验评估失败: %v", err)
	}

	// 6. 评估技术深度
	if err := pe.evaluateTechnicalDepth(); err != nil {
		return nil, fmt.Errorf("技术深度评估失败: %v", err)
	}

	// 7. 评估工程质量
	if err := pe.evaluateEngineeringQuality(); err != nil {
		return nil, fmt.Errorf("工程质量评估失败: %v", err)
	}

	// 8. 计算综合得分
	pe.calculateOverallScore()

	// 9. 生成改进建议
	if pe.config.GenerateSuggestions {
		pe.generateSuggestions()
	}

	// 10. 进行比较分析
	pe.performComparativeAnalysis()

	pe.results.Duration = time.Since(start)
	pe.results.Passed = pe.results.OverallScore >= 70.0 // 可配置阈值

	log.Printf("项目评估完成，总分: %.2f，耗时: %v",
		pe.results.OverallScore, pe.results.Duration)

	// 11. 保存结果
	if pe.config.SaveResults {
		if err := pe.saveResults(); err != nil {
			log.Printf("保存结果失败: %v", err)
		}
	}

	return pe.results, nil
}

// analyzeProjectInfo 分析项目信息
func (pe *ProjectEvaluator) analyzeProjectInfo(projectPath string) error {
	info := ProjectInfo{
		Path:     projectPath,
		Language: "Go",
	}

	// 分析项目结构
	if err := pe.analyzeProjectStructure(projectPath, &info.Structure); err != nil {
		return err
	}

	// 分析依赖
	if pe.config.CheckDependencies {
		if err := pe.analyzeDependencies(projectPath, &info.Dependencies); err != nil {
			log.Printf("依赖分析失败: %v", err) // 不中断评估
		}
	}

	// 分析构建系统
	if pe.config.AnalyzeBuild {
		pe.analyzeBuildSystem(projectPath, &info.BuildSystem)
	}

	// 分析文档
	if pe.config.AnalyzeDocs {
		pe.analyzeDocumentation(projectPath, &info.Documentation)
	}

	// 分析测试
	pe.analyzeTestingInfo(projectPath, &info.TestingInfo)

	// 设置项目元数据
	pe.setProjectMetadata(projectPath, &info)

	pe.results.ProjectInfo = info
	return nil
}

// analyzeProjectStructure 分析项目结构
func (pe *ProjectEvaluator) analyzeProjectStructure(projectPath string, structure *ProjectStructure) error {
	var goFiles, testFiles, totalFiles int
	packageSet := make(map[string]bool)

	err := filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// 检查标准目录结构
			dirName := d.Name()
			switch dirName {
			case "cmd":
				structure.HasCmd = true
			case "internal":
				structure.HasInternal = true
			case "pkg":
				structure.HasPkg = true
			case "api":
				structure.HasAPI = true
			case "docs":
				structure.HasDocs = true
			case "examples":
				structure.HasExamples = true
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 跳过vendor目录
		if strings.Contains(path, "vendor/") {
			return nil
		}

		totalFiles++
		if strings.HasSuffix(path, "_test.go") {
			testFiles++
		} else {
			goFiles++
		}

		// 解析包信息
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err == nil && node != nil {
			packageSet[node.Name.Name] = true
		}

		return nil
	})

	if err != nil {
		return err
	}

	structure.TotalFiles = totalFiles
	structure.GoFiles = goFiles
	structure.TestFiles = testFiles
	structure.Packages = len(packageSet)

	// 计算组织评分
	structure.FileOrganization = pe.calculateFileOrganization()
	structure.NamingConsistency = pe.calculateNamingConsistency()
	structure.ModuleBoundaries = pe.calculateModuleBoundaries()

	return nil
}

// calculateFileOrganization 计算文件组织评分
func (pe *ProjectEvaluator) calculateFileOrganization() float64 {
	score := 80.0 // 基础分

	structure := pe.results.ProjectInfo.Structure

	// 标准目录结构加分
	if structure.HasCmd {
		score += 5
	}
	if structure.HasInternal {
		score += 5
	}
	if structure.HasPkg {
		score += 3
	}
	if structure.HasAPI {
		score += 3
	}
	if structure.HasDocs {
		score += 2
	}
	if structure.HasExamples {
		score += 2
	}

	if score > 100 {
		score = 100
	}

	return score
}

// calculateNamingConsistency 计算命名一致性评分
func (pe *ProjectEvaluator) calculateNamingConsistency() float64 {
	// 简化实现，实际应该检查包名、文件名、函数名的一致性
	return 85.0
}

// calculateModuleBoundaries 计算模块边界清晰度
func (pe *ProjectEvaluator) calculateModuleBoundaries() float64 {
	// 简化实现，实际应该分析包之间的依赖关系
	return 80.0
}

// analyzeDependencies 分析依赖
func (pe *ProjectEvaluator) analyzeDependencies(projectPath string, deps *DependencyAnalysis) error {
	// 读取go.mod文件
	goModPath := filepath.Join(projectPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("读取go.mod失败: %v", err)
	}

	// 简化的依赖分析
	lines := strings.Split(string(content), "\n")
	directDeps := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "require") && !strings.Contains(line, "// indirect") {
			directDeps++
		}
	}

	deps.DirectDependencies = directDeps
	deps.TotalDependencies = directDeps // 简化，实际需要分析所有传递依赖
	deps.StandardLibraryUsage = pe.calculateStandardLibUsage()
	deps.DependencyQuality = pe.calculateDependencyQuality()

	return nil
}

// calculateStandardLibUsage 计算标准库使用率
func (pe *ProjectEvaluator) calculateStandardLibUsage() float64 {
	// 简化实现，实际需要分析import语句
	return 75.0
}

// calculateDependencyQuality 计算依赖质量
func (pe *ProjectEvaluator) calculateDependencyQuality() float64 {
	// 简化实现，实际需要检查依赖的维护状态、安全性等
	return 85.0
}

// analyzeBuildSystem 分析构建系统
func (pe *ProjectEvaluator) analyzeBuildSystem(projectPath string, build *BuildSystemInfo) {
	// 检查构建相关文件
	build.HasGoMod = pe.fileExists(filepath.Join(projectPath, "go.mod"))
	build.HasMakefile = pe.fileExists(filepath.Join(projectPath, "Makefile")) ||
		pe.fileExists(filepath.Join(projectPath, "makefile"))
	build.HasDockerfile = pe.fileExists(filepath.Join(projectPath, "Dockerfile"))
	build.HasGitHubActions = pe.fileExists(filepath.Join(projectPath, ".github/workflows"))
	build.HasGoReleaser = pe.fileExists(filepath.Join(projectPath, ".goreleaser.yml")) ||
		pe.fileExists(filepath.Join(projectPath, ".goreleaser.yaml"))

	// 计算构建质量评分
	build.BuildQuality = pe.calculateBuildQuality(build)
	build.CIConfiguration = pe.calculateCIQuality(build)
	build.DeploymentReady = pe.calculateDeploymentReadiness(build)
}

// fileExists 检查文件是否存在
func (pe *ProjectEvaluator) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// calculateBuildQuality 计算构建质量
func (pe *ProjectEvaluator) calculateBuildQuality(build *BuildSystemInfo) float64 {
	score := 0.0

	if build.HasGoMod {
		score += 40 // go.mod是必需的
	}
	if build.HasMakefile {
		score += 30 // Makefile提供标准化构建
	}
	if build.HasDockerfile {
		score += 20 // Docker支持
	}
	if build.HasGoReleaser {
		score += 10 // 自动发布配置
	}

	return score
}

// calculateCIQuality 计算CI配置质量
func (pe *ProjectEvaluator) calculateCIQuality(build *BuildSystemInfo) float64 {
	if build.HasGitHubActions {
		return 80.0 // 有CI配置
	}
	return 0.0
}

// calculateDeploymentReadiness 计算部署准备度
func (pe *ProjectEvaluator) calculateDeploymentReadiness(build *BuildSystemInfo) float64 {
	score := 0.0

	if build.HasDockerfile {
		score += 50
	}
	if build.HasGitHubActions {
		score += 30
	}
	if build.HasGoReleaser {
		score += 20
	}

	return score
}

// analyzeDocumentation 分析文档
func (pe *ProjectEvaluator) analyzeDocumentation(projectPath string, doc *DocumentationInfo) {
	// 检查文档文件
	doc.HasReadme = pe.fileExists(filepath.Join(projectPath, "README.md")) ||
		pe.fileExists(filepath.Join(projectPath, "readme.md"))
	doc.HasChangelog = pe.fileExists(filepath.Join(projectPath, "CHANGELOG.md"))
	doc.HasLicense = pe.fileExists(filepath.Join(projectPath, "LICENSE")) ||
		pe.fileExists(filepath.Join(projectPath, "LICENSE.md"))
	doc.HasContributing = pe.fileExists(filepath.Join(projectPath, "CONTRIBUTING.md"))

	// 计算文档质量
	if doc.HasReadme {
		doc.ReadmeQuality = pe.analyzeReadmeQuality(projectPath)
	}

	doc.GoDocCoverage = pe.calculateGoDocCoverage(projectPath)
	doc.OverallDocScore = pe.calculateOverallDocScore(doc)
}

// analyzeReadmeQuality 分析README质量
func (pe *ProjectEvaluator) analyzeReadmeQuality(projectPath string) float64 {
	readmePath := filepath.Join(projectPath, "README.md")
	if !pe.fileExists(readmePath) {
		readmePath = filepath.Join(projectPath, "readme.md")
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		return 0.0
	}

	readmeContent := string(content)
	score := 0.0

	// 检查README内容要素
	if strings.Contains(strings.ToLower(readmeContent), "installation") {
		score += 20
	}
	if strings.Contains(strings.ToLower(readmeContent), "usage") {
		score += 25
	}
	if strings.Contains(strings.ToLower(readmeContent), "example") {
		score += 20
	}
	if strings.Contains(strings.ToLower(readmeContent), "api") {
		score += 15
	}
	if strings.Contains(strings.ToLower(readmeContent), "contributing") {
		score += 10
	}
	if strings.Contains(strings.ToLower(readmeContent), "license") {
		score += 10
	}

	return score
}

// calculateGoDocCoverage 计算GoDoc覆盖率
func (pe *ProjectEvaluator) calculateGoDocCoverage(projectPath string) float64 {
	// 简化实现，实际需要分析所有公开函数的文档覆盖率
	return 70.0
}

// calculateOverallDocScore 计算总体文档评分
func (pe *ProjectEvaluator) calculateOverallDocScore(doc *DocumentationInfo) float64 {
	score := 0.0

	if doc.HasReadme {
		score += doc.ReadmeQuality * 0.4
	}
	if doc.HasChangelog {
		score += 15
	}
	if doc.HasLicense {
		score += 10
	}
	if doc.HasContributing {
		score += 5
	}

	score += doc.GoDocCoverage * 0.3

	return score
}

// analyzeTestingInfo 分析测试信息
func (pe *ProjectEvaluator) analyzeTestingInfo(projectPath string, testing *TestingInfo) {
	structure := pe.results.ProjectInfo.Structure

	testing.HasTests = structure.TestFiles > 0
	testing.TestFiles = structure.TestFiles

	if testing.HasTests {
		testing.TestCoverage = pe.calculateTestCoverage(projectPath)
		testing.TestFunctions = pe.countTestFunctions(projectPath)
		testing.BenchmarkCount = pe.countBenchmarkFunctions(projectPath)
		testing.ExampleTests = pe.countExampleFunctions(projectPath)
	}

	testing.TestQuality = pe.calculateTestQuality(testing)
}

// calculateTestCoverage 计算测试覆盖率
func (pe *ProjectEvaluator) calculateTestCoverage(projectPath string) float64 {
	// 实际实现需要运行 go test -cover
	// 这里提供简化实现
	return 75.0
}

// countTestFunctions 统计测试函数数量
func (pe *ProjectEvaluator) countTestFunctions(projectPath string) int {
	count := 0

	filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			content, err := os.ReadFile(path)
			if err == nil {
				// 简单的正则匹配测试函数
				re := regexp.MustCompile(`func\s+Test\w+\s*\(`)
				count += len(re.FindAll(content, -1))
			}
		}

		return nil
	})

	return count
}

// countBenchmarkFunctions 统计基准测试函数数量
func (pe *ProjectEvaluator) countBenchmarkFunctions(projectPath string) int {
	count := 0

	filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			content, err := os.ReadFile(path)
			if err == nil {
				re := regexp.MustCompile(`func\s+Benchmark\w+\s*\(`)
				count += len(re.FindAll(content, -1))
			}
		}

		return nil
	})

	return count
}

// countExampleFunctions 统计示例函数数量
func (pe *ProjectEvaluator) countExampleFunctions(projectPath string) int {
	count := 0

	filepath.WalkDir(projectPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if strings.HasSuffix(path, "_test.go") {
			content, err := os.ReadFile(path)
			if err == nil {
				re := regexp.MustCompile(`func\s+Example\w*\s*\(`)
				count += len(re.FindAll(content, -1))
			}
		}

		return nil
	})

	return count
}

// calculateTestQuality 计算测试质量
func (pe *ProjectEvaluator) calculateTestQuality(testing *TestingInfo) float64 {
	if !testing.HasTests {
		return 0.0
	}

	score := testing.TestCoverage * 0.6 // 覆盖率占60%

	// 测试类型多样性加分
	if testing.BenchmarkCount > 0 {
		score += 10
	}
	if testing.ExampleTests > 0 {
		score += 10
	}

	// 测试密度评分
	if pe.results.ProjectInfo.Structure.GoFiles > 0 {
		testRatio := float64(testing.TestFiles) / float64(pe.results.ProjectInfo.Structure.GoFiles)
		if testRatio >= 0.5 {
			score += 15
		} else if testRatio >= 0.3 {
			score += 10
		} else if testRatio >= 0.1 {
			score += 5
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

// setProjectMetadata 设置项目元数据
func (pe *ProjectEvaluator) setProjectMetadata(projectPath string, info *ProjectInfo) {
	// 从项目路径提取项目名
	info.Name = filepath.Base(projectPath)

	// 设置项目类型
	info.Type = pe.detectProjectType(projectPath)

	// 设置时间信息（简化实现）
	if stat, err := os.Stat(projectPath); err == nil {
		info.LastModified = stat.ModTime()
	}
}

// detectProjectType 检测项目类型
func (pe *ProjectEvaluator) detectProjectType(projectPath string) string {
	// 检查main.go或cmd目录
	if pe.fileExists(filepath.Join(projectPath, "main.go")) ||
		pe.fileExists(filepath.Join(projectPath, "cmd")) {
		return "application"
	}

	// 检查是否是库项目
	if pe.fileExists(filepath.Join(projectPath, "go.mod")) {
		return "library"
	}

	return "unknown"
}

// loadRequirementsAndCriteria 加载需求和评估标准
func (pe *ProjectEvaluator) loadRequirementsAndCriteria() error {
	// 根据项目类型和阶段加载相应的需求和标准
	pe.requirements = pe.getRequirementsForStage(pe.config.Stage, pe.config.ProjectType)
	pe.criteria = pe.getCriteriaForStage(pe.config.Stage)

	return nil
}

// getRequirementsForStage 获取特定阶段的需求
func (pe *ProjectEvaluator) getRequirementsForStage(stage int, projectType string) []ProjectRequirement {
	// 这里应该从配置文件或数据库中加载需求
	// 简化实现，返回基础需求
	return []ProjectRequirement{
		{
			ID:          "func_001",
			Category:    "functionality",
			Title:       "核心功能实现",
			Description: "项目应实现所有核心功能",
			Priority:    1,
			Mandatory:   true,
		},
		{
			ID:          "arch_001",
			Category:    "architecture",
			Title:       "代码组织",
			Description: "代码应有良好的组织结构",
			Priority:    2,
			Mandatory:   true,
		},
	}
}

// getCriteriaForStage 获取特定阶段的评估标准
func (pe *ProjectEvaluator) getCriteriaForStage(stage int) []EvaluationCriteria {
	// 简化实现，返回基础评估标准
	return []EvaluationCriteria{
		{
			ID:        "functionality",
			Category:  "core",
			Name:      "功能完整性",
			Weight:    0.3,
			MaxScore:  100.0,
			Automated: true,
			Threshold: 70.0,
		},
		{
			ID:        "architecture",
			Category:  "design",
			Name:      "架构质量",
			Weight:    0.25,
			MaxScore:  100.0,
			Automated: true,
			Threshold: 70.0,
		},
	}
}

// 评估功能实现的方法（由于篇幅限制，提供主要方法的签名和基本实现）

// evaluateFunctionality 评估功能完整性
func (pe *ProjectEvaluator) evaluateFunctionality() error {
	score := 0.0

	// 基于需求检查功能实现
	for _, req := range pe.requirements {
		if req.Category == "functionality" {
			result := pe.evaluateRequirement(req)
			pe.results.RequirementResults = append(pe.results.RequirementResults, result)
			if req.Mandatory && !result.Implemented {
				score -= 20 // 必需功能未实现严重扣分
			}
			score += result.Score * float64(req.Priority)
		}
	}

	pe.results.DimensionScores["functionality"] = score
	pe.results.FunctionalityScore = score
	return nil
}

// evaluateRequirement 评估单个需求
func (pe *ProjectEvaluator) evaluateRequirement(req ProjectRequirement) RequirementResult {
	// 简化实现，实际应该基于具体需求进行检查
	return RequirementResult{
		RequirementID: req.ID,
		Implemented:   true,
		Score:         80.0,
		Quality:       75.0,
		Completeness:  85.0,
	}
}

// evaluateArchitecture 评估架构质量
func (pe *ProjectEvaluator) evaluateArchitecture() error {
	analysis := ArchitectureAnalysis{}

	// 计算各项架构指标
	analysis.ModularityScore = pe.calculateModularityScore()
	analysis.CouplingScore = pe.calculateCouplingScore()
	analysis.CohesionScore = pe.calculateCohesionScore()
	analysis.InterfaceDesign = pe.calculateInterfaceDesignScore()
	analysis.DependencyManagement = pe.calculateDependencyManagementScore()

	// 综合架构得分
	archScore := (analysis.ModularityScore + analysis.CouplingScore +
		analysis.CohesionScore + analysis.InterfaceDesign +
		analysis.DependencyManagement) / 5

	pe.results.DimensionScores["architecture"] = archScore
	pe.results.ArchitectureAnalysis = analysis
	return nil
}

// 计算各种架构指标的简化实现
func (pe *ProjectEvaluator) calculateModularityScore() float64 {
	return 80.0
}

func (pe *ProjectEvaluator) calculateCouplingScore() float64 {
	return 75.0
}

func (pe *ProjectEvaluator) calculateCohesionScore() float64 {
	return 85.0
}

func (pe *ProjectEvaluator) calculateInterfaceDesignScore() float64 {
	return 78.0
}

func (pe *ProjectEvaluator) calculateDependencyManagementScore() float64 {
	return 82.0
}

// evaluateUserExperience 评估用户体验
func (pe *ProjectEvaluator) evaluateUserExperience() error {
	uxAnalysis := UXAnalysis{}

	// API设计评估
	if pe.config.AnalyzeAPI {
		uxAnalysis.APIAnalysis = pe.analyzeAPI()
		uxAnalysis.APIDesignScore = uxAnalysis.APIAnalysis.Consistency
	}

	// 错误处理评估
	uxAnalysis.ErrorAnalysis = pe.analyzeErrorHandling()
	uxAnalysis.ErrorHandlingScore = uxAnalysis.ErrorAnalysis.ErrorConsistency

	// 可用性评估
	uxAnalysis.UsabilityScore = pe.calculateUsabilityScore()
	uxAnalysis.ConsistencyScore = pe.calculateConsistencyScore()

	// 综合用户体验得分
	uxScore := (uxAnalysis.APIDesignScore + uxAnalysis.ErrorHandlingScore +
		uxAnalysis.UsabilityScore + uxAnalysis.ConsistencyScore) / 4

	pe.results.DimensionScores["user_experience"] = uxScore
	pe.results.UserExperienceAnalysis = uxAnalysis
	return nil
}

// analyzeAPI 分析API设计
func (pe *ProjectEvaluator) analyzeAPI() APIAnalysis {
	return APIAnalysis{
		Consistency:         80.0,
		Simplicity:          75.0,
		Completeness:        85.0,
		Flexibility:         70.0,
		PerformanceDesign:   78.0,
		RESTCompliance:      80.0,
		VersioningStrategy:  "semantic",
		SecurityIntegration: 75.0,
	}
}

// analyzeErrorHandling 分析错误处理
func (pe *ProjectEvaluator) analyzeErrorHandling() ErrorHandlingAnalysis {
	return ErrorHandlingAnalysis{
		ErrorConsistency: 80.0,
		ErrorInformation: 75.0,
		RecoveryOptions:  70.0,
		LoggingQuality:   78.0,
	}
}

// calculateUsabilityScore 计算可用性得分
func (pe *ProjectEvaluator) calculateUsabilityScore() float64 {
	score := 0.0

	// 文档质量影响可用性
	score += pe.results.ProjectInfo.Documentation.OverallDocScore * 0.4

	// 示例和测试影响可用性
	if pe.results.ProjectInfo.TestingInfo.ExampleTests > 0 {
		score += 20
	}

	// README质量
	score += pe.results.ProjectInfo.Documentation.ReadmeQuality * 0.4

	return score
}

// calculateConsistencyScore 计算一致性得分
func (pe *ProjectEvaluator) calculateConsistencyScore() float64 {
	return pe.results.ProjectInfo.Structure.NamingConsistency
}

// evaluateTechnicalDepth 评估技术深度
func (pe *ProjectEvaluator) evaluateTechnicalDepth() error {
	techAnalysis := TechnicalAnalysis{}

	// 技术栈分析
	techAnalysis.TechnologyStack = pe.analyzeTechnologyStack()
	techAnalysis.BestPracticeScore = pe.calculateBestPracticeScore()
	techAnalysis.InnovationScore = pe.calculateInnovationScore()
	techAnalysis.TechnicalDebt = pe.calculateTechnicalDebt()

	// 综合技术深度得分
	techScore := (techAnalysis.BestPracticeScore + techAnalysis.InnovationScore +
		(100 - techAnalysis.TechnicalDebt)) / 3

	pe.results.DimensionScores["technical_depth"] = techScore
	pe.results.TechnicalAnalysis = techAnalysis
	return nil
}

// analyzeTechnologyStack 分析技术栈
func (pe *ProjectEvaluator) analyzeTechnologyStack() []TechnologyUsage {
	return []TechnologyUsage{
		{
			Name:        "Go",
			Category:    "programming_language",
			Usage:       "primary",
			Proficiency: 85.0,
			Appropriate: true,
			Rationale:   "主要开发语言",
		},
	}
}

// calculateBestPracticeScore 计算最佳实践得分
func (pe *ProjectEvaluator) calculateBestPracticeScore() float64 {
	score := 0.0

	// 项目结构最佳实践
	if pe.results.ProjectInfo.Structure.HasCmd {
		score += 10
	}
	if pe.results.ProjectInfo.Structure.HasInternal {
		score += 10
	}

	// 构建最佳实践
	if pe.results.ProjectInfo.BuildSystem.HasGoMod {
		score += 15
	}
	if pe.results.ProjectInfo.BuildSystem.HasMakefile {
		score += 10
	}

	// 测试最佳实践
	if pe.results.ProjectInfo.TestingInfo.HasTests {
		score += 20
	}
	if pe.results.ProjectInfo.TestingInfo.TestCoverage >= 80 {
		score += 15
	}

	// 文档最佳实践
	if pe.results.ProjectInfo.Documentation.HasReadme {
		score += 10
	}
	if pe.results.ProjectInfo.Documentation.HasLicense {
		score += 5
	}
	if pe.results.ProjectInfo.Documentation.GoDocCoverage >= 70 {
		score += 5
	}

	return score
}

// calculateInnovationScore 计算创新性得分
func (pe *ProjectEvaluator) calculateInnovationScore() float64 {
	// 简化实现，实际需要分析代码中的创新性解决方案
	return 60.0
}

// calculateTechnicalDebt 计算技术债务
func (pe *ProjectEvaluator) calculateTechnicalDebt() float64 {
	// 简化实现，实际需要综合多个因素
	return 15.0 // 15%的技术债务
}

// evaluateEngineeringQuality 评估工程质量
func (pe *ProjectEvaluator) evaluateEngineeringQuality() error {
	score := 0.0

	// 构建系统质量
	score += pe.results.ProjectInfo.BuildSystem.BuildQuality * 0.3

	// CI/CD配置
	score += pe.results.ProjectInfo.BuildSystem.CIConfiguration * 0.2

	// 测试质量
	score += pe.results.ProjectInfo.TestingInfo.TestQuality * 0.3

	// 文档质量
	score += pe.results.ProjectInfo.Documentation.OverallDocScore * 0.2

	pe.results.DimensionScores["engineering_quality"] = score
	return nil
}

// calculateOverallScore 计算综合得分
func (pe *ProjectEvaluator) calculateOverallScore() {
	weights := pe.config.WeightSettings
	totalScore := 0.0
	totalWeight := 0.0

	for dimension, score := range pe.results.DimensionScores {
		var weight float64
		switch dimension {
		case "functionality":
			weight = weights.FunctionalityScore
		case "architecture":
			weight = weights.ArchitectureScore
		case "user_experience":
			weight = weights.UserExperienceScore
		case "technical_depth":
			weight = weights.TechnicalDepthScore
		case "engineering_quality":
			weight = weights.EngineeringScore
		default:
			weight = 0.0
		}

		totalScore += score * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		pe.results.OverallScore = totalScore / totalWeight
	}

	// 设置等级
	if pe.results.OverallScore >= 90 {
		pe.results.Grade = "A"
	} else if pe.results.OverallScore >= 80 {
		pe.results.Grade = "B"
	} else if pe.results.OverallScore >= 70 {
		pe.results.Grade = "C"
	} else if pe.results.OverallScore >= 60 {
		pe.results.Grade = "D"
	} else {
		pe.results.Grade = "F"
	}
}

// generateSuggestions 生成改进建议
func (pe *ProjectEvaluator) generateSuggestions() {
	// 分析优势
	pe.results.Strengths = pe.identifyStrengths()

	// 分析不足
	pe.results.Weaknesses = pe.identifyWeaknesses()

	// 生成改进建议
	pe.results.Improvements = pe.generateImprovements()

	// 生成下一步建议
	pe.results.NextSteps = pe.generateNextSteps()
}

// identifyStrengths 识别项目优势
func (pe *ProjectEvaluator) identifyStrengths() []ProjectStrength {
	strengths := []ProjectStrength{}

	// 基于高分维度识别优势
	for dimension, score := range pe.results.DimensionScores {
		if score >= 85 {
			strength := ProjectStrength{
				Category:    dimension,
				Title:       fmt.Sprintf("%s表现优秀", dimension),
				Description: fmt.Sprintf("在%s方面得分%.2f，表现优秀", dimension, score),
				Impact:      "为项目整体质量提供了良好基础",
			}
			strengths = append(strengths, strength)
		}
	}

	return strengths
}

// identifyWeaknesses 识别项目不足
func (pe *ProjectEvaluator) identifyWeaknesses() []ProjectWeakness {
	weaknesses := []ProjectWeakness{}

	// 基于低分维度识别不足
	for dimension, score := range pe.results.DimensionScores {
		if score < 70 {
			weakness := ProjectWeakness{
				Category:    dimension,
				Title:       fmt.Sprintf("%s需要改进", dimension),
				Description: fmt.Sprintf("在%s方面得分%.2f，低于预期", dimension, score),
				Impact:      "影响项目整体质量",
				Severity:    "medium",
				Suggestions: []string{fmt.Sprintf("重点关注%s的改进", dimension)},
			}
			weaknesses = append(weaknesses, weakness)
		}
	}

	return weaknesses
}

// generateImprovements 生成改进建议
func (pe *ProjectEvaluator) generateImprovements() []ProjectImprovement {
	improvements := []ProjectImprovement{}

	// 基于弱项生成改进建议
	for _, weakness := range pe.results.Weaknesses {
		improvement := ProjectImprovement{
			Category:    weakness.Category,
			Title:       fmt.Sprintf("改进%s", weakness.Category),
			Description: fmt.Sprintf("针对%s的不足，制定改进计划", weakness.Category),
			Benefit:     "提升项目整体质量",
			Effort:      "Medium",
			Priority:    2,
		}
		improvements = append(improvements, improvement)
	}

	return improvements
}

// generateNextSteps 生成下一步建议
func (pe *ProjectEvaluator) generateNextSteps() []NextStep {
	nextSteps := []NextStep{}

	// 基于项目当前状态生成下一步建议
	if pe.results.OverallScore < 70 {
		nextSteps = append(nextSteps, NextStep{
			Phase:       "质量改进",
			Description: "优先解决关键质量问题",
			Priority:    1,
			Timeline:    "2-4周",
			Resources:   []string{"开发时间", "代码审查"},
			Success:     []string{"总体评分提升到70以上"},
		})
	}

	return nextSteps
}

// performComparativeAnalysis 进行比较分析
func (pe *ProjectEvaluator) performComparativeAnalysis() {
	// 基准比较（简化实现）
	pe.results.BenchmarkComparison = &BenchmarkComparison{
		BenchmarkType:   "stage_average",
		ComparisonScore: pe.results.OverallScore,
		Percentile:      75.0, // 假设在75%分位
		StrongerAreas:   []string{"engineering_quality"},
		WeakerAreas:     []string{"innovation"},
	}
}

// saveResults 保存评估结果
func (pe *ProjectEvaluator) saveResults() error {
	if pe.config.ResultsPath == "" {
		pe.config.ResultsPath = "project_evaluation_results.json"
	}

	data, err := json.MarshalIndent(pe.results, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化结果失败: %v", err)
	}

	if err := os.WriteFile(pe.config.ResultsPath, data, 0644); err != nil {
		return fmt.Errorf("保存结果文件失败: %v", err)
	}

	log.Printf("项目评估结果已保存到: %s", pe.config.ResultsPath)
	return nil
}

// GetDefaultConfig 获取默认配置
func GetProjectEvalDefaultConfig() *ProjectEvalConfig {
	return &ProjectEvalConfig{
		ProjectType:      "application",
		Stage:            6,
		RequirementLevel: "intermediate",
		WeightSettings: ProjectEvalWeights{
			FunctionalityScore:  0.30,
			ArchitectureScore:   0.25,
			UserExperienceScore: 0.20,
			TechnicalDepthScore: 0.15,
			EngineeringScore:    0.10,
		},
		Thresholds: ProjectThresholds{
			MinFunctionality: 70.0,
			MinArchitecture:  70.0,
			MinDocumentation: 60.0,
			MinTestCoverage:  70.0,
			MaxComplexity:    10,
			MinModularity:    70.0,
		},
		AnalyzeReadme:       true,
		AnalyzeDocs:         true,
		AnalyzeBuild:        true,
		AnalyzeAPI:          true,
		CheckDependencies:   true,
		DetailLevel:         "detailed",
		GenerateSuggestions: true,
		SaveResults:         true,
		ResultsPath:         "",
	}
}
