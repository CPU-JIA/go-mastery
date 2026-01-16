package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewAssessmentFramework(t *testing.T) {
	tests := []struct {
		name    string
		version string
		fwName  string
	}{
		{
			name:    "standard framework",
			version: "1.0.0",
			fwName:  "Go Assessment Framework",
		},
		{
			name:    "Chinese name framework",
			version: "2.0.0",
			fwName:  "Go语言能力评估框架",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fw := NewAssessmentFramework(tt.version, tt.fwName)

			if fw == nil {
				t.Fatal("NewAssessmentFramework returned nil")
			}

			if fw.Version != tt.version {
				t.Errorf("Version = %q, want %q", fw.Version, tt.version)
			}

			if fw.Name != tt.fwName {
				t.Errorf("Name = %q, want %q", fw.Name, tt.fwName)
			}

			// Check default values
			if fw.Dimensions == nil {
				t.Error("Dimensions should be initialized")
			}

			if fw.ScoringRules == nil {
				t.Error("ScoringRules should be initialized")
			}

			if fw.Standards == nil {
				t.Error("Standards should be initialized")
			}

			if fw.LevelRequirements == nil {
				t.Error("LevelRequirements should be initialized")
			}

			// Check default settings
			if !fw.AutoAssessment {
				t.Error("AutoAssessment should be true by default")
			}

			if fw.PeerReview {
				t.Error("PeerReview should be false by default")
			}

			if !fw.MentorReview {
				t.Error("MentorReview should be true by default")
			}

			// Check timestamps
			if fw.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}

			if fw.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestGetDefaultWeightMatrix(t *testing.T) {
	wm := getDefaultWeightMatrix()

	// Check dimension weights sum to 1.0 (with floating point tolerance)
	dimensionSum := wm.TechnicalDepth + wm.EngineeringPractice + wm.ProjectExperience + wm.SoftSkills
	if diff := dimensionSum - 1.0; diff < -0.001 || diff > 0.001 {
		t.Errorf("Dimension weights sum = %f, want 1.0", dimensionSum)
	}

	// Check assessment method weights sum to 1.0 (with floating point tolerance)
	methodSum := wm.AutomatedAssessment + wm.CodeReview + wm.ProjectEvaluation + wm.PeerFeedback + wm.MentorAssessment
	if diff := methodSum - 1.0; diff < -0.001 || diff > 0.001 {
		t.Errorf("Assessment method weights sum = %f, want 1.0", methodSum)
	}

	// Check individual weights
	if wm.TechnicalDepth != WeightTechnicalDepth {
		t.Errorf("TechnicalDepth = %f, want %f", wm.TechnicalDepth, WeightTechnicalDepth)
	}

	if wm.EngineeringPractice != WeightEngineeringPractice {
		t.Errorf("EngineeringPractice = %f, want %f", wm.EngineeringPractice, WeightEngineeringPractice)
	}

	// Check StageWeights is initialized
	if wm.StageWeights == nil {
		t.Error("StageWeights should be initialized")
	}
}

func TestGetDefaultThresholds(t *testing.T) {
	thresholds := getDefaultThresholds()

	expectedThresholds := map[string]float64{
		"passing_score":      ThresholdPassingScore,
		"excellent_score":    ThresholdExcellentScore,
		"min_coverage":       ThresholdMinCoverage,
		"max_complexity":     ThresholdMaxComplexity,
		"min_documentation":  ThresholdMinDocumentation,
		"performance_target": ThresholdPerformanceTarget,
	}

	for key, expected := range expectedThresholds {
		if got, exists := thresholds[key]; !exists {
			t.Errorf("Threshold %q not found", key)
		} else if got != expected {
			t.Errorf("Threshold[%q] = %f, want %f", key, got, expected)
		}
	}
}

func TestGetDefaultLevelRequirements(t *testing.T) {
	requirements := getDefaultLevelRequirements()

	expectedLevels := []string{"Bronze", "Silver", "Gold", "Platinum"}

	for _, level := range expectedLevels {
		req, exists := requirements[level]
		if !exists {
			t.Errorf("Level %q not found in requirements", level)
			continue
		}

		if req.Level != level {
			t.Errorf("Level = %q, want %q", req.Level, level)
		}

		if !req.ExamRequired {
			t.Errorf("Level %q should require exam", level)
		}

		if req.ExamDuration <= 0 {
			t.Errorf("Level %q ExamDuration should be positive", level)
		}

		if len(req.RequiredStages) == 0 {
			t.Errorf("Level %q should have required stages", level)
		}
	}

	// Check progression of requirements
	if requirements["Bronze"].MinScore >= requirements["Silver"].MinScore {
		t.Error("Silver should require higher score than Bronze")
	}

	if requirements["Silver"].MinScore >= requirements["Gold"].MinScore {
		t.Error("Gold should require higher score than Silver")
	}

	if requirements["Gold"].MinScore >= requirements["Platinum"].MinScore {
		t.Error("Platinum should require higher score than Gold")
	}
}

func TestCalculateGrade(t *testing.T) {
	tests := []struct {
		percentage float64
		wantGrade  string
	}{
		{100.0, "A+"},
		{95.0, "A+"},
		{94.9, "A"},
		{90.0, "A"},
		{89.9, "A-"},
		{85.0, "A-"},
		{84.9, "B+"},
		{80.0, "B+"},
		{79.9, "B"},
		{75.0, "B"},
		{74.9, "B-"},
		{70.0, "B-"},
		{69.9, "C+"},
		{65.0, "C+"},
		{64.9, "C"},
		{60.0, "C"},
		{59.9, "F"},
		{50.0, "F"},
		{0.0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.wantGrade, func(t *testing.T) {
			got := calculateGrade(tt.percentage)
			if got != tt.wantGrade {
				t.Errorf("calculateGrade(%f) = %q, want %q", tt.percentage, got, tt.wantGrade)
			}
		})
	}
}

func TestAssessmentResult_CalculateOverallScore(t *testing.T) {
	framework := NewAssessmentFramework("1.0.0", "Test Framework")
	framework.Dimensions = []AssessmentDimension{
		{ID: "technical", Weight: 0.4, MaxScore: 100},
		{ID: "practice", Weight: 0.3, MaxScore: 100},
		{ID: "project", Weight: 0.2, MaxScore: 100},
		{ID: "soft", Weight: 0.1, MaxScore: 100},
	}

	tests := []struct {
		name            string
		dimensionScores map[string]float64
		maxScore        float64
		wantScore       float64
		wantGrade       string
	}{
		{
			name: "all perfect scores",
			dimensionScores: map[string]float64{
				"technical": 100,
				"practice":  100,
				"project":   100,
				"soft":      100,
			},
			maxScore:  100,
			wantScore: 100,
			wantGrade: "A+",
		},
		{
			name: "mixed scores",
			dimensionScores: map[string]float64{
				"technical": 80,
				"practice":  70,
				"project":   60,
				"soft":      50,
			},
			maxScore:  100,
			wantScore: 70, // 80*0.4 + 70*0.3 + 60*0.2 + 50*0.1 = 32 + 21 + 12 + 5 = 70
			wantGrade: "B-",
		},
		{
			name: "partial dimensions",
			dimensionScores: map[string]float64{
				"technical": 90,
				"practice":  80,
			},
			maxScore:  100,
			wantScore: 85.71, // (90*0.4 + 80*0.3) / (0.4 + 0.3) = 60 / 0.7 ≈ 85.71
			wantGrade: "A-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &AssessmentResult{
				DimensionScores: tt.dimensionScores,
				MaxScore:        tt.maxScore,
			}

			got := result.CalculateOverallScore(framework)

			// Allow small floating point difference
			diff := got - tt.wantScore
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.1 {
				t.Errorf("CalculateOverallScore() = %f, want %f", got, tt.wantScore)
			}

			if result.Grade != tt.wantGrade {
				t.Errorf("Grade = %q, want %q", result.Grade, tt.wantGrade)
			}
		})
	}
}

func TestAssessmentResult_CalculateOverallScore_EmptyDimensions(t *testing.T) {
	framework := NewAssessmentFramework("1.0.0", "Test Framework")
	framework.Dimensions = []AssessmentDimension{
		{ID: "technical", Weight: 0.5, MaxScore: 100},
	}

	result := &AssessmentResult{
		DimensionScores: map[string]float64{},
		MaxScore:        100,
	}

	got := result.CalculateOverallScore(framework)

	// With no matching dimensions, score should be 0
	if got != 0 {
		t.Errorf("CalculateOverallScore() with empty dimensions = %f, want 0", got)
	}
}

func TestAssessmentFramework_ToJSON(t *testing.T) {
	fw := NewAssessmentFramework("1.0.0", "Test Framework")

	jsonData, err := fw.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("ToJSON() returned empty data")
	}

	// Verify JSON is valid
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Errorf("ToJSON() produced invalid JSON: %v", err)
	}

	// Check key fields
	if parsed["version"] != "1.0.0" {
		t.Errorf("JSON version = %v, want %q", parsed["version"], "1.0.0")
	}

	if parsed["name"] != "Test Framework" {
		t.Errorf("JSON name = %v, want %q", parsed["name"], "Test Framework")
	}
}

func TestAssessmentFramework_FromJSON(t *testing.T) {
	jsonData := `{
		"version": "2.0.0",
		"name": "Loaded Framework",
		"auto_assessment": false,
		"peer_review": true
	}`

	fw := &AssessmentFramework{}
	err := fw.FromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if fw.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", fw.Version, "2.0.0")
	}

	if fw.Name != "Loaded Framework" {
		t.Errorf("Name = %q, want %q", fw.Name, "Loaded Framework")
	}

	if fw.AutoAssessment {
		t.Error("AutoAssessment should be false")
	}

	if !fw.PeerReview {
		t.Error("PeerReview should be true")
	}
}

func TestAssessmentFramework_FromJSON_Invalid(t *testing.T) {
	invalidJSON := `{"version": 123}` // version should be string

	fw := &AssessmentFramework{}
	err := fw.FromJSON([]byte(invalidJSON))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON")
	}
}

func TestAssessmentDimension(t *testing.T) {
	dimension := AssessmentDimension{
		ID:               "technical_depth",
		Name:             "Technical Depth",
		Description:      "Measures technical knowledge depth",
		Weight:           0.4,
		MaxScore:         100,
		EvaluationMethod: "automated",
	}

	if dimension.ID != "technical_depth" {
		t.Errorf("ID = %q, want %q", dimension.ID, "technical_depth")
	}

	if dimension.Weight != 0.4 {
		t.Errorf("Weight = %f, want 0.4", dimension.Weight)
	}

	if dimension.MaxScore != 100 {
		t.Errorf("MaxScore = %f, want 100", dimension.MaxScore)
	}
}

func TestAssessmentTask(t *testing.T) {
	now := time.Now()
	task := AssessmentTask{
		ID:            "task_001",
		Name:          "Implement Calculator",
		Description:   "Build a CLI calculator",
		Type:          "coding",
		Stage:         1,
		Difficulty:    3,
		EstimatedTime: 60,
		AutoGrading:   true,
		CreatedAt:     now,
		Tags:          []string{"cli", "basics"},
	}

	if task.ID != "task_001" {
		t.Errorf("ID = %q, want %q", task.ID, "task_001")
	}

	if task.Difficulty != 3 {
		t.Errorf("Difficulty = %d, want 3", task.Difficulty)
	}

	if !task.AutoGrading {
		t.Error("AutoGrading should be true")
	}

	if len(task.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(task.Tags))
	}
}

func TestAssessmentSession(t *testing.T) {
	now := time.Now()
	session := AssessmentSession{
		ID:        "session_001",
		StudentID: "student_001",
		TaskID:    "task_001",
		Type:      "coding",
		Status:    "in_progress",
		StartTime: now,
	}

	if session.ID != "session_001" {
		t.Errorf("ID = %q, want %q", session.ID, "session_001")
	}

	if session.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", session.Status, "in_progress")
	}

	if session.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
}

func TestCodeAnalysisResult(t *testing.T) {
	result := CodeAnalysisResult{
		LinesOfCode:          500,
		CyclomaticComplexity: 8,
		TestCoverage:         85.5,
		CodeQuality:          90.0,
		SecurityScore:        95.0,
		PerformanceScore:     88.0,
	}

	if result.LinesOfCode != 500 {
		t.Errorf("LinesOfCode = %d, want 500", result.LinesOfCode)
	}

	if result.CyclomaticComplexity != 8 {
		t.Errorf("CyclomaticComplexity = %d, want 8", result.CyclomaticComplexity)
	}

	if result.TestCoverage != 85.5 {
		t.Errorf("TestCoverage = %f, want 85.5", result.TestCoverage)
	}
}

// Benchmark tests
func BenchmarkNewAssessmentFramework(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewAssessmentFramework("1.0.0", "Benchmark Framework")
	}
}

func BenchmarkCalculateGrade(b *testing.B) {
	percentages := []float64{95.0, 85.0, 75.0, 65.0, 55.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateGrade(percentages[i%len(percentages)])
	}
}

func BenchmarkAssessmentResult_CalculateOverallScore(b *testing.B) {
	framework := NewAssessmentFramework("1.0.0", "Benchmark Framework")
	framework.Dimensions = []AssessmentDimension{
		{ID: "technical", Weight: 0.4, MaxScore: 100},
		{ID: "practice", Weight: 0.3, MaxScore: 100},
		{ID: "project", Weight: 0.2, MaxScore: 100},
		{ID: "soft", Weight: 0.1, MaxScore: 100},
	}

	result := &AssessmentResult{
		DimensionScores: map[string]float64{
			"technical": 85,
			"practice":  80,
			"project":   75,
			"soft":      70,
		},
		MaxScore: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = result.CalculateOverallScore(framework)
	}
}

func BenchmarkAssessmentFramework_ToJSON(b *testing.B) {
	fw := NewAssessmentFramework("1.0.0", "Benchmark Framework")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fw.ToJSON()
	}
}
