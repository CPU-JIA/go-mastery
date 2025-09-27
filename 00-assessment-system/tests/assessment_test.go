package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"../evaluators"
	"../models"
)

func TestCodeQualityEvaluator(t *testing.T) {
	evaluator := evaluators.NewCodeQualityEvaluator()

	t.Run("EvaluateBasicCode", func(t *testing.T) {
		code := `
package main

import "fmt"

// HelloWorld prints hello world message
func HelloWorld() {
    fmt.Println("Hello, World!")
}

func main() {
    HelloWorld()
}
`

		result := evaluator.Evaluate(code, "main.go")

		assert.NotNil(t, result)
		assert.Greater(t, result.Score, 70.0) // Good quality code should score high
		assert.Contains(t, result.Feedback, "Good")
	})

	t.Run("EvaluatePoorCode", func(t *testing.T) {
		code := `
package main
import "fmt"
func main(){
fmt.Println("bad code")
x:=1
y:=2
z:=x+y
fmt.Println(z)
}
`

		result := evaluator.Evaluate(code, "bad.go")

		assert.NotNil(t, result)
		assert.Less(t, result.Score, 50.0) // Poor quality code should score low
		assert.Contains(t, result.Issues, "formatting")
	})

	t.Run("EvaluateWithComplexity", func(t *testing.T) {
		code := `
package main

func complexFunction() {
    for i := 0; i < 10; i++ {
        for j := 0; j < 10; j++ {
            for k := 0; k < 10; k++ {
                if i > 5 {
                    if j > 5 {
                        if k > 5 {
                            // Complex nested logic
                        }
                    }
                }
            }
        }
    }
}
`

		result := evaluator.Evaluate(code, "complex.go")

		assert.NotNil(t, result)
		assert.Less(t, result.Score, 60.0) // High complexity should reduce score
		assert.Contains(t, result.Issues, "complexity")
	})
}

func TestProjectEvaluator(t *testing.T) {
	evaluator := evaluators.NewProjectEvaluator()

	t.Run("EvaluateSimpleProject", func(t *testing.T) {
		files := map[string]string{
			"main.go": `package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}`,
			"README.md": "# Test Project\nA simple test project",
			"go.mod": "module test\ngo 1.24.6",
		}

		result := evaluator.EvaluateProject(files)

		assert.NotNil(t, result)
		assert.Greater(t, result.OverallScore, 0.0)
		assert.NotEmpty(t, result.Strengths)
		assert.NotEmpty(t, result.Recommendations)
	})

	t.Run("EvaluateProjectWithTests", func(t *testing.T) {
		files := map[string]string{
			"main.go": `package main
import "fmt"
func Add(a, b int) int { return a + b }
func main() { fmt.Println(Add(1, 2)) }`,
			"main_test.go": `package main
import "testing"
func TestAdd(t *testing.T) {
    if Add(1, 2) != 3 { t.Error("Add failed") }
}`,
			"README.md": "# Test Project",
			"go.mod": "module test\ngo 1.24.6",
		}

		result := evaluator.EvaluateProject(files)

		assert.NotNil(t, result)
		assert.Greater(t, result.OverallScore, 70.0) // Projects with tests should score higher
		assert.Contains(t, result.Strengths, "test")
	})

	t.Run("EvaluateIncompleteProject", func(t *testing.T) {
		files := map[string]string{
			"main.go": `package main
func main() {
    // TODO: implement
}`,
		}

		result := evaluator.EvaluateProject(files)

		assert.NotNil(t, result)
		assert.Less(t, result.OverallScore, 50.0) // Incomplete projects should score low
		assert.NotEmpty(t, result.Recommendations)
	})
}

func TestAssessmentModel(t *testing.T) {
	t.Run("CreateAssessment", func(t *testing.T) {
		assessment := &models.Assessment{
			ID:          "test-1",
			StudentID:   "student-123",
			ProjectName: "Hello World",
			CreatedAt:   time.Now(),
		}

		assert.Equal(t, "test-1", assessment.ID)
		assert.Equal(t, "student-123", assessment.StudentID)
		assert.Equal(t, "Hello World", assessment.ProjectName)
		assert.False(t, assessment.CreatedAt.IsZero())
	})

	t.Run("AddEvaluationResult", func(t *testing.T) {
		assessment := &models.Assessment{
			ID:        "test-1",
			StudentID: "student-123",
		}

		result := &models.EvaluationResult{
			Category: "code_quality",
			Score:    85.5,
			Feedback: "Good code quality",
		}

		assessment.Results = append(assessment.Results, *result)

		assert.Len(t, assessment.Results, 1)
		assert.Equal(t, "code_quality", assessment.Results[0].Category)
		assert.Equal(t, 85.5, assessment.Results[0].Score)
	})

	t.Run("CalculateOverallScore", func(t *testing.T) {
		assessment := &models.Assessment{
			Results: []models.EvaluationResult{
				{Category: "code_quality", Score: 80.0, Weight: 0.4},
				{Category: "functionality", Score: 90.0, Weight: 0.3},
				{Category: "design", Score: 70.0, Weight: 0.3},
			},
		}

		overallScore := assessment.CalculateOverallScore()

		// Expected: 80*0.4 + 90*0.3 + 70*0.3 = 32 + 27 + 21 = 80
		assert.InDelta(t, 80.0, overallScore, 0.1)
	})
}

func TestCompetencyModel(t *testing.T) {
	t.Run("CreateCompetency", func(t *testing.T) {
		competency := &models.Competency{
			ID:          "go-basics",
			Name:        "Go Programming Basics",
			Description: "Basic understanding of Go syntax and concepts",
			Category:    "programming",
			Level:       models.BeginnerLevel,
		}

		assert.Equal(t, "go-basics", competency.ID)
		assert.Equal(t, "Go Programming Basics", competency.Name)
		assert.Equal(t, models.BeginnerLevel, competency.Level)
	})

	t.Run("CompetencyLevels", func(t *testing.T) {
		assert.Equal(t, "beginner", string(models.BeginnerLevel))
		assert.Equal(t, "intermediate", string(models.IntermediateLevel))
		assert.Equal(t, "advanced", string(models.AdvancedLevel))
		assert.Equal(t, "expert", string(models.ExpertLevel))
	})
}

func TestStudentModel(t *testing.T) {
	t.Run("CreateStudent", func(t *testing.T) {
		student := &models.Student{
			ID:       "student-123",
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Level:    models.IntermediateLevel,
			JoinedAt: time.Now(),
		}

		assert.Equal(t, "student-123", student.ID)
		assert.Equal(t, "John Doe", student.Name)
		assert.Equal(t, "john.doe@example.com", student.Email)
		assert.Equal(t, models.IntermediateLevel, student.Level)
	})

	t.Run("AddAssessment", func(t *testing.T) {
		student := &models.Student{
			ID:   "student-123",
			Name: "John Doe",
		}

		assessment := models.Assessment{
			ID:          "assessment-1",
			StudentID:   "student-123",
			ProjectName: "Calculator",
		}

		student.Assessments = append(student.Assessments, assessment)

		assert.Len(t, student.Assessments, 1)
		assert.Equal(t, "assessment-1", student.Assessments[0].ID)
	})

	t.Run("CalculateProgress", func(t *testing.T) {
		student := &models.Student{
			Assessments: []models.Assessment{
				{
					Results: []models.EvaluationResult{
						{Score: 70.0, Weight: 1.0},
					},
				},
				{
					Results: []models.EvaluationResult{
						{Score: 80.0, Weight: 1.0},
					},
				},
				{
					Results: []models.EvaluationResult{
						{Score: 90.0, Weight: 1.0},
					},
				},
			},
		}

		progress := student.CalculateProgress()
		assert.Greater(t, progress, 0.0)
		assert.LessOrEqual(t, progress, 100.0)
	})
}

// Integration tests
func TestAssessmentWorkflow(t *testing.T) {
	t.Run("CompleteAssessmentFlow", func(t *testing.T) {
		// Create student
		student := &models.Student{
			ID:    "student-123",
			Name:  "Test Student",
			Email: "test@example.com",
			Level: models.BeginnerLevel,
		}

		// Create assessment
		assessment := &models.Assessment{
			ID:          "assessment-1",
			StudentID:   student.ID,
			ProjectName: "First Go Program",
			CreatedAt:   time.Now(),
		}

		// Evaluate code quality
		codeEvaluator := evaluators.NewCodeQualityEvaluator()
		code := `package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}`

		codeResult := codeEvaluator.Evaluate(code, "main.go")
		assessment.Results = append(assessment.Results, models.EvaluationResult{
			Category: "code_quality",
			Score:    codeResult.Score,
			Feedback: codeResult.Feedback,
			Weight:   0.5,
		})

		// Evaluate project structure
		projectEvaluator := evaluators.NewProjectEvaluator()
		files := map[string]string{
			"main.go": code,
			"go.mod":  "module hello\ngo 1.24.6",
		}

		projectResult := projectEvaluator.EvaluateProject(files)
		assessment.Results = append(assessment.Results, models.EvaluationResult{
			Category: "project_structure",
			Score:    projectResult.OverallScore,
			Feedback: "Project structure evaluation",
			Weight:   0.3,
		})

		// Calculate overall score
		overallScore := assessment.CalculateOverallScore()
		assessment.OverallScore = overallScore

		// Add assessment to student
		student.Assessments = append(student.Assessments, *assessment)

		// Verify results
		assert.Len(t, student.Assessments, 1)
		assert.Greater(t, overallScore, 0.0)
		assert.LessOrEqual(t, overallScore, 100.0)
		assert.Len(t, assessment.Results, 2)
	})
}

// Benchmark tests
func BenchmarkCodeEvaluation(b *testing.B) {
	evaluator := evaluators.NewCodeQualityEvaluator()
	code := `package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.Evaluate(code, "main.go")
	}
}

func BenchmarkProjectEvaluation(b *testing.B) {
	evaluator := evaluators.NewProjectEvaluator()
	files := map[string]string{
		"main.go": `package main
import "fmt"
func main() {
    fmt.Println("Hello, World!")
}`,
		"go.mod": "module test\ngo 1.24.6",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.EvaluateProject(files)
	}
}