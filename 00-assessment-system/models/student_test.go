package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewStudentProfile(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		sname string
		email string
	}{
		{
			name:  "valid student profile",
			id:    "stu_001",
			sname: "Zhang San",
			email: "zhangsan@example.com",
		},
		{
			name:  "student with Chinese name",
			id:    "stu_002",
			sname: "李明",
			email: "liming@example.com",
		},
		{
			name:  "student with empty email",
			id:    "stu_003",
			sname: "Test User",
			email: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			student := NewStudentProfile(tt.id, tt.sname, tt.email)

			if student == nil {
				t.Fatal("NewStudentProfile returned nil")
			}

			if student.ID != tt.id {
				t.Errorf("ID = %q, want %q", student.ID, tt.id)
			}

			if student.Name != tt.sname {
				t.Errorf("Name = %q, want %q", student.Name, tt.sname)
			}

			if student.Email != tt.email {
				t.Errorf("Email = %q, want %q", student.Email, tt.email)
			}

			// Check default values
			if student.CurrentStage != 1 {
				t.Errorf("CurrentStage = %d, want 1", student.CurrentStage)
			}

			if student.TotalHours != 0 {
				t.Errorf("TotalHours = %f, want 0", student.TotalHours)
			}

			if student.StageProgress == nil {
				t.Error("StageProgress should be initialized")
			}

			if student.Competencies == nil {
				t.Error("Competencies should be initialized")
			}

			if student.Projects == nil {
				t.Error("Projects should be initialized")
			}

			// Check preferences defaults
			if student.Preferences.PreferredPace != "normal" {
				t.Errorf("PreferredPace = %q, want %q", student.Preferences.PreferredPace, "normal")
			}

			if student.Preferences.AvailableHours != DefaultAvailableHours {
				t.Errorf("AvailableHours = %f, want %f", student.Preferences.AvailableHours, DefaultAvailableHours)
			}

			if !student.Preferences.EmailNotifications {
				t.Error("EmailNotifications should be true by default")
			}
		})
	}
}

func TestStudentProfile_UpdateProgress(t *testing.T) {
	tests := []struct {
		name           string
		stage          int
		progress       float64
		hoursSpent     float64
		wantStage      int
		wantCompleted  bool
		wantTotalHours float64
	}{
		{
			name:           "partial progress",
			stage:          1,
			progress:       0.5,
			hoursSpent:     5.0,
			wantStage:      1,
			wantCompleted:  false,
			wantTotalHours: 5.0,
		},
		{
			name:           "complete stage",
			stage:          1,
			progress:       1.0,
			hoursSpent:     10.0,
			wantStage:      2,
			wantCompleted:  true,
			wantTotalHours: 10.0,
		},
		{
			name:           "progress exceeds 1.0",
			stage:          1,
			progress:       1.5,
			hoursSpent:     15.0,
			wantStage:      2,
			wantCompleted:  true,
			wantTotalHours: 15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			student := NewStudentProfile("test_001", "Test", "test@example.com")

			student.UpdateProgress(tt.stage, tt.progress, tt.hoursSpent)

			if student.TotalHours != tt.wantTotalHours {
				t.Errorf("TotalHours = %f, want %f", student.TotalHours, tt.wantTotalHours)
			}

			if student.CurrentStage != tt.wantStage {
				t.Errorf("CurrentStage = %d, want %d", student.CurrentStage, tt.wantStage)
			}

			stageProgress, exists := student.StageProgress[tt.stage]
			if !exists {
				t.Fatal("StageProgress not recorded")
			}

			if tt.wantCompleted {
				if stageProgress.Status != "completed" {
					t.Errorf("Status = %q, want %q", stageProgress.Status, "completed")
				}
				if stageProgress.CompleteDate == nil {
					t.Error("CompleteDate should be set when completed")
				}
			} else {
				if stageProgress.Status != "in_progress" {
					t.Errorf("Status = %q, want %q", stageProgress.Status, "in_progress")
				}
			}
		})
	}
}

func TestStudentProfile_UpdateProgress_Incremental(t *testing.T) {
	student := NewStudentProfile("test_001", "Test", "test@example.com")

	// First update
	student.UpdateProgress(1, 0.3, 3.0)
	if student.TotalHours != 3.0 {
		t.Errorf("TotalHours after first update = %f, want 3.0", student.TotalHours)
	}

	// Second update - hours should accumulate
	student.UpdateProgress(1, 0.6, 4.0)
	if student.TotalHours != 7.0 {
		t.Errorf("TotalHours after second update = %f, want 7.0", student.TotalHours)
	}

	// Check stage progress time accumulation
	stageProgress := student.StageProgress[1]
	if stageProgress.TimeSpent != 7.0 {
		t.Errorf("StageProgress.TimeSpent = %f, want 7.0", stageProgress.TimeSpent)
	}

	if stageProgress.Progress != 0.6 {
		t.Errorf("StageProgress.Progress = %f, want 0.6", stageProgress.Progress)
	}
}

func TestStudentProfile_AddProject(t *testing.T) {
	student := NewStudentProfile("test_001", "Test", "test@example.com")

	project := &ProjectRecord{
		Name:         "CLI Calculator",
		Type:         "cli",
		Stage:        1,
		Technologies: []string{"Go", "Cobra"},
		OverallScore: 8.5,
	}

	student.AddProject(project)

	if len(student.Projects) != 1 {
		t.Fatalf("Projects count = %d, want 1", len(student.Projects))
	}

	addedProject := student.Projects[0]
	if addedProject.Name != "CLI Calculator" {
		t.Errorf("Project Name = %q, want %q", addedProject.Name, "CLI Calculator")
	}

	// Check auto-generated ID
	expectedID := "proj_1_CLI Calculator"
	if addedProject.ID != expectedID {
		t.Errorf("Project ID = %q, want %q", addedProject.ID, expectedID)
	}

	// Add second project
	project2 := &ProjectRecord{
		Name:         "Web API",
		Type:         "api",
		Stage:        4,
		OverallScore: 9.0,
	}
	student.AddProject(project2)

	if len(student.Projects) != 2 {
		t.Fatalf("Projects count = %d, want 2", len(student.Projects))
	}

	expectedID2 := "proj_2_Web API"
	if student.Projects[1].ID != expectedID2 {
		t.Errorf("Second Project ID = %q, want %q", student.Projects[1].ID, expectedID2)
	}
}

func TestStudentProfile_AddAssessment(t *testing.T) {
	student := NewStudentProfile("test_001", "Test", "test@example.com")

	assessment := &AssessmentRecord{
		Timestamp: time.Now(),
		Stage:     1,
		Type:      "auto",
		Scores: struct {
			TechnicalDepth      float64 `json:"technical_depth"`
			EngineeringPractice float64 `json:"engineering_practice"`
			ProjectExperience   float64 `json:"project_experience"`
			SoftSkills          float64 `json:"soft_skills"`
			OverallScore        float64 `json:"overall_score"`
		}{
			TechnicalDepth:      75.0,
			EngineeringPractice: 70.0,
			ProjectExperience:   65.0,
			SoftSkills:          60.0,
			OverallScore:        70.0,
		},
		Confidence: 0.85,
	}

	student.AddAssessment(assessment)

	if len(student.Assessments) != 1 {
		t.Fatalf("Assessments count = %d, want 1", len(student.Assessments))
	}

	addedAssessment := student.Assessments[0]
	expectedID := "assess_1_1"
	if addedAssessment.ID != expectedID {
		t.Errorf("Assessment ID = %q, want %q", addedAssessment.ID, expectedID)
	}

	if addedAssessment.Scores.OverallScore != 70.0 {
		t.Errorf("OverallScore = %f, want 70.0", addedAssessment.Scores.OverallScore)
	}
}

func TestStudentProfile_AddCertification(t *testing.T) {
	student := NewStudentProfile("test_001", "Test", "test@example.com")

	cert := &CertificationRecord{
		Level:       "Bronze",
		AwardDate:   time.Now(),
		Score:       78.5,
		Certificate: "CERT-BRONZE-001",
		Verified:    true,
		Status:      "active",
	}

	student.AddCertification(cert)

	if len(student.Certifications) != 1 {
		t.Fatalf("Certifications count = %d, want 1", len(student.Certifications))
	}

	addedCert := student.Certifications[0]
	if addedCert.Level != "Bronze" {
		t.Errorf("Certification Level = %q, want %q", addedCert.Level, "Bronze")
	}

	if addedCert.Score != 78.5 {
		t.Errorf("Certification Score = %f, want 78.5", addedCert.Score)
	}
}

func TestStudentProfile_GetCurrentLevel(t *testing.T) {
	tests := []struct {
		name           string
		certifications []CertificationRecord
		wantLevel      string
	}{
		{
			name:           "no certifications",
			certifications: []CertificationRecord{},
			wantLevel:      "None",
		},
		{
			name: "single Bronze certification",
			certifications: []CertificationRecord{
				{Level: "Bronze", Status: "active"},
			},
			wantLevel: "Bronze",
		},
		{
			name: "multiple certifications - returns highest",
			certifications: []CertificationRecord{
				{Level: "Bronze", Status: "active"},
				{Level: "Silver", Status: "active"},
			},
			wantLevel: "Silver",
		},
		{
			name: "expired certification ignored",
			certifications: []CertificationRecord{
				{Level: "Bronze", Status: "active"},
				{Level: "Gold", Status: "expired"},
			},
			wantLevel: "Bronze",
		},
		{
			name: "all certifications expired",
			certifications: []CertificationRecord{
				{Level: "Silver", Status: "expired"},
				{Level: "Gold", Status: "revoked"},
			},
			wantLevel: "Bronze", // Returns default lowest level
		},
		{
			name: "Platinum certification",
			certifications: []CertificationRecord{
				{Level: "Bronze", Status: "active"},
				{Level: "Silver", Status: "active"},
				{Level: "Gold", Status: "active"},
				{Level: "Platinum", Status: "active"},
			},
			wantLevel: "Platinum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			student := NewStudentProfile("test_001", "Test", "test@example.com")
			student.Certifications = tt.certifications

			got := student.GetCurrentLevel()
			if got != tt.wantLevel {
				t.Errorf("GetCurrentLevel() = %q, want %q", got, tt.wantLevel)
			}
		})
	}
}

func TestStudentProfile_GetOverallScore(t *testing.T) {
	tests := []struct {
		name        string
		assessments []AssessmentRecord
		wantScore   float64
	}{
		{
			name:        "no assessments",
			assessments: []AssessmentRecord{},
			wantScore:   0.0,
		},
		{
			name: "single assessment",
			assessments: []AssessmentRecord{
				{Scores: struct {
					TechnicalDepth      float64 `json:"technical_depth"`
					EngineeringPractice float64 `json:"engineering_practice"`
					ProjectExperience   float64 `json:"project_experience"`
					SoftSkills          float64 `json:"soft_skills"`
					OverallScore        float64 `json:"overall_score"`
				}{OverallScore: 80.0}},
			},
			wantScore: 80.0,
		},
		{
			name: "multiple assessments - average",
			assessments: []AssessmentRecord{
				{Scores: struct {
					TechnicalDepth      float64 `json:"technical_depth"`
					EngineeringPractice float64 `json:"engineering_practice"`
					ProjectExperience   float64 `json:"project_experience"`
					SoftSkills          float64 `json:"soft_skills"`
					OverallScore        float64 `json:"overall_score"`
				}{OverallScore: 70.0}},
				{Scores: struct {
					TechnicalDepth      float64 `json:"technical_depth"`
					EngineeringPractice float64 `json:"engineering_practice"`
					ProjectExperience   float64 `json:"project_experience"`
					SoftSkills          float64 `json:"soft_skills"`
					OverallScore        float64 `json:"overall_score"`
				}{OverallScore: 80.0}},
				{Scores: struct {
					TechnicalDepth      float64 `json:"technical_depth"`
					EngineeringPractice float64 `json:"engineering_practice"`
					ProjectExperience   float64 `json:"project_experience"`
					SoftSkills          float64 `json:"soft_skills"`
					OverallScore        float64 `json:"overall_score"`
				}{OverallScore: 90.0}},
			},
			wantScore: 80.0, // (70 + 80 + 90) / 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			student := NewStudentProfile("test_001", "Test", "test@example.com")
			student.Assessments = tt.assessments

			got := student.GetOverallScore()
			if got != tt.wantScore {
				t.Errorf("GetOverallScore() = %f, want %f", got, tt.wantScore)
			}
		})
	}
}

func TestStudentProfile_ToJSON(t *testing.T) {
	student := NewStudentProfile("test_001", "Test User", "test@example.com")
	student.CurrentStage = 3
	student.TotalHours = 50.5

	jsonData, err := student.ToJSON()
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
	if parsed["id"] != "test_001" {
		t.Errorf("JSON id = %v, want %q", parsed["id"], "test_001")
	}

	if parsed["name"] != "Test User" {
		t.Errorf("JSON name = %v, want %q", parsed["name"], "Test User")
	}
}

func TestStudentProfile_FromJSON(t *testing.T) {
	jsonData := `{
		"id": "test_002",
		"name": "JSON User",
		"email": "json@example.com",
		"current_stage": 5,
		"total_hours": 100.5
	}`

	student := &StudentProfile{}
	err := student.FromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if student.ID != "test_002" {
		t.Errorf("ID = %q, want %q", student.ID, "test_002")
	}

	if student.Name != "JSON User" {
		t.Errorf("Name = %q, want %q", student.Name, "JSON User")
	}

	if student.CurrentStage != 5 {
		t.Errorf("CurrentStage = %d, want 5", student.CurrentStage)
	}

	if student.TotalHours != 100.5 {
		t.Errorf("TotalHours = %f, want 100.5", student.TotalHours)
	}
}

func TestStudentProfile_FromJSON_Invalid(t *testing.T) {
	invalidJSON := `{"id": "test", "current_stage": "not_a_number"}`

	student := &StudentProfile{}
	err := student.FromJSON([]byte(invalidJSON))
	if err == nil {
		t.Error("FromJSON() should return error for invalid JSON")
	}
}

func TestGetStageNameByID(t *testing.T) {
	tests := []struct {
		stageID  int
		wantName string
	}{
		{0, "评估系统"},
		{1, "Go语言基础"},
		{2, "高级语言特性"},
		{3, "并发编程"},
		{4, "Web开发"},
		{5, "数据库集成"},
		{6, "实战项目"},
		{7, "运行时内部机制"},
		{8, "高级网络编程"},
		{9, "微服务架构"},
		{10, "编译器工具链"},
		{11, "大规模系统"},
		{12, "DevOps部署"},
		{13, "性能优化"},
		{14, "技术领导力"},
		{15, "开源贡献"},
		{99, "阶段99"}, // Unknown stage
		{-1, "阶段-1"}, // Negative stage
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			got := getStageNameByID(tt.stageID)
			if got != tt.wantName {
				t.Errorf("getStageNameByID(%d) = %q, want %q", tt.stageID, got, tt.wantName)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewStudentProfile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewStudentProfile("test_001", "Test User", "test@example.com")
	}
}

func BenchmarkStudentProfile_UpdateProgress(b *testing.B) {
	student := NewStudentProfile("test_001", "Test User", "test@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		student.UpdateProgress(1, 0.5, 1.0)
	}
}

func BenchmarkStudentProfile_GetOverallScore(b *testing.B) {
	student := NewStudentProfile("test_001", "Test User", "test@example.com")
	// Add some assessments
	for i := 0; i < 10; i++ {
		student.Assessments = append(student.Assessments, AssessmentRecord{
			Scores: struct {
				TechnicalDepth      float64 `json:"technical_depth"`
				EngineeringPractice float64 `json:"engineering_practice"`
				ProjectExperience   float64 `json:"project_experience"`
				SoftSkills          float64 `json:"soft_skills"`
				OverallScore        float64 `json:"overall_score"`
			}{OverallScore: float64(70 + i)},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = student.GetOverallScore()
	}
}

func BenchmarkStudentProfile_ToJSON(b *testing.B) {
	student := NewStudentProfile("test_001", "Test User", "test@example.com")
	student.UpdateProgress(1, 0.5, 10.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = student.ToJSON()
	}
}
