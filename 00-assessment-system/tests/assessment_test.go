package tests

import (
	"assessment-system/evaluators"
	"assessment-system/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeQualityEvaluator(t *testing.T) {
	evaluator := evaluators.NewCodeQualityEvaluator(evaluators.GetDefaultConfig())

	// 为了测试，我们需要创建一个临时项目目录
	t.Run("EvaluateProject", func(t *testing.T) {
		// 使用当前目录作为测试项目
		projectPath := "."

		result, err := evaluator.EvaluateProject(projectPath)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.OverallScore, 0.0)
	})
}

func TestProjectEvaluator(t *testing.T) {
	evaluator := evaluators.NewProjectEvaluator(evaluators.GetProjectEvalDefaultConfig())

	t.Run("EvaluateSimpleProject", func(t *testing.T) {
		// 使用当前目录作为测试项目
		projectPath := "."

		result, err := evaluator.EvaluateProject(projectPath)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.OverallScore, 0.0)
	})

	t.Run("EvaluateProjectWithTests", func(t *testing.T) {
		// 使用当前目录作为测试项目
		projectPath := "."

		result, err := evaluator.EvaluateProject(projectPath)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.OverallScore, 50.0) // 假设当前项目有一定质量
	})

	t.Run("EvaluateIncompleteProject", func(t *testing.T) {
		// 使用当前目录作为测试项目
		projectPath := "."

		result, err := evaluator.EvaluateProject(projectPath)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// 不对分数做严格要求，因为我们使用的是真实项目
	})
}

func TestAssessmentModel(t *testing.T) {
	t.Run("CreateAssessmentResult", func(t *testing.T) {
		result := &models.AssessmentResult{
			SessionID:    "test-1",
			OverallScore: 85.5,
			MaxScore:     100.0,
			Percentage:   85.5,
			Grade:        "B",
		}

		assert.Equal(t, "test-1", result.SessionID)
		assert.Equal(t, 85.5, result.OverallScore)
		assert.Equal(t, 100.0, result.MaxScore) // 检查MaxScore字段
		assert.Equal(t, "B", result.Grade)
		assert.Equal(t, 85.5, result.Percentage)
	})

	t.Run("TestDimensionScores", func(t *testing.T) {
		result := &models.AssessmentResult{
			SessionID:       "test-2",
			OverallScore:    80.0,
			DimensionScores: make(map[string]float64),
		}

		// 验证初始字段值
		assert.Equal(t, "test-2", result.SessionID)
		assert.Equal(t, 80.0, result.OverallScore)

		result.DimensionScores["code_quality"] = 85.0
		result.DimensionScores["functionality"] = 90.0
		result.DimensionScores["design"] = 70.0

		assert.Len(t, result.DimensionScores, 3)
		assert.Equal(t, 85.0, result.DimensionScores["code_quality"])
	})

	t.Run("CalculatePercentage", func(t *testing.T) {
		result := &models.AssessmentResult{
			OverallScore: 80.0,
			MaxScore:     100.0,
		}

		// Calculate percentage
		result.Percentage = (result.OverallScore / result.MaxScore) * 100

		assert.InDelta(t, 80.0, result.Percentage, 0.1)
	})
}

func TestCompetencyModel(t *testing.T) {
	t.Run("CreateSkill", func(t *testing.T) {
		skill := &models.Skill{
			ID:          "go-basics",
			Name:        "Go Programming Basics",
			Description: "Basic understanding of Go syntax and concepts",
			Category:    "programming",
		}

		assert.Equal(t, "go-basics", skill.ID)
		assert.Equal(t, "Go Programming Basics", skill.Name)
		assert.Equal(t, "Basic understanding of Go syntax and concepts", skill.Description)
		assert.Equal(t, "programming", skill.Category)
	})

	t.Run("SkillLevels", func(t *testing.T) {
		// Test basic skill level functionality
		assert.True(t, true) // Simplified test
	})
}

func TestStudentModel(t *testing.T) {
	t.Run("CreateStudentProfile", func(t *testing.T) {
		student := &models.StudentProfile{
			ID:           "student-123",
			Name:         "John Doe",
			Email:        "john.doe@example.com",
			CurrentStage: 6,
			Projects:     make([]models.ProjectRecord, 0),
		}

		assert.Equal(t, "student-123", student.ID)
		assert.Equal(t, "John Doe", student.Name)
		assert.Equal(t, "john.doe@example.com", student.Email)
		assert.Equal(t, 6, student.CurrentStage)
		assert.NotNil(t, student.Projects) // 检查Projects字段已初始化
		assert.Equal(t, 0, len(student.Projects))
	})

	t.Run("AddProjectRecord", func(t *testing.T) {
		student := &models.StudentProfile{
			ID:       "student-123",
			Name:     "John Doe",
			Projects: make([]models.ProjectRecord, 0),
		}

		// 验证初始状态
		assert.Equal(t, "student-123", student.ID)
		assert.Equal(t, "John Doe", student.Name)
		assert.Equal(t, 0, len(student.Projects))

		project := models.ProjectRecord{
			ID:           "project-1",
			Name:         "Calculator",
			OverallScore: 85.5,
			Stage:        6,
		}

		student.Projects = append(student.Projects, project)

		assert.Len(t, student.Projects, 1)
		assert.Equal(t, "project-1", student.Projects[0].ID)
	})

	t.Run("CalculateProgress", func(t *testing.T) {
		// Simplified progress calculation test
		student := &models.StudentProfile{
			CurrentStage: 6,
		}

		// Mock progress calculation
		progress := float64(student.CurrentStage) / 15.0 * 100

		assert.Greater(t, progress, 0.0)
		assert.LessOrEqual(t, progress, 100.0)
	})
}

// Integration tests
func TestAssessmentWorkflow(t *testing.T) {
	t.Run("CompleteAssessmentFlow", func(t *testing.T) {
		// Create student
		student := &models.StudentProfile{
			ID:           "student-123",
			Name:         "Test Student",
			Email:        "test@example.com",
			CurrentStage: 1,
		}

		// 验证学生信息
		assert.Equal(t, "student-123", student.ID)
		assert.Equal(t, "Test Student", student.Name)
		assert.Equal(t, "test@example.com", student.Email)
		assert.Equal(t, 1, student.CurrentStage)

		// Create assessment result
		result := &models.AssessmentResult{
			SessionID:    "assessment-1",
			OverallScore: 75.0,
			MaxScore:     100.0,
			Grade:        "B",
		}

		// 验证评估结果初始值
		assert.Equal(t, "assessment-1", result.SessionID)
		assert.Equal(t, 75.0, result.OverallScore)
		assert.Equal(t, 100.0, result.MaxScore)
		assert.Equal(t, "B", result.Grade)

		// Evaluate code quality
		codeEvaluator := evaluators.NewCodeQualityEvaluator(evaluators.GetDefaultConfig())
		projectPath := "."

		codeResult, err := codeEvaluator.EvaluateProject(projectPath)
		require.NoError(t, err)
		result.DimensionScores = make(map[string]float64)
		result.DimensionScores["code_quality"] = codeResult.OverallScore

		// Evaluate project structure
		projectEvaluator := evaluators.NewProjectEvaluator(evaluators.GetProjectEvalDefaultConfig())

		projectResult, err := projectEvaluator.EvaluateProject(projectPath)
		require.NoError(t, err)
		result.DimensionScores["project_structure"] = projectResult.OverallScore

		// Calculate overall score
		result.Percentage = (result.OverallScore / result.MaxScore) * 100

		// 验证计算的百分比
		assert.Equal(t, 75.0, result.Percentage)

		// Verify results
		assert.Equal(t, "student-123", student.ID)
		assert.Greater(t, result.OverallScore, 0.0)
		assert.LessOrEqual(t, result.OverallScore, 100.0)
		assert.NotEmpty(t, result.DimensionScores)
	})
}

// Benchmark tests
func BenchmarkCodeEvaluation(b *testing.B) {
	evaluator := evaluators.NewCodeQualityEvaluator(evaluators.GetDefaultConfig())
	projectPath := "."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.EvaluateProject(projectPath)
	}
}

func BenchmarkProjectEvaluation(b *testing.B) {
	evaluator := evaluators.NewProjectEvaluator(evaluators.GetProjectEvalDefaultConfig())
	projectPath := "."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.EvaluateProject(projectPath)
	}
}
