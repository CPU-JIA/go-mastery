package model

import (
	"time"

	"gorm.io/gorm"
)

// SagaExecution Saga执行记录
type SagaExecution struct {
	ID                uint                `gorm:"primaryKey" json:"id"`
	SagaType          string              `gorm:"not null;index" json:"saga_type"`        // "FileUpload", "FileDelete"
	RequestID         string              `gorm:"uniqueIndex;not null" json:"request_id"` // 唯一请求标识
	UserID            uint                `gorm:"index" json:"user_id"`
	User              User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status            SagaStatus          `gorm:"default:pending;index" json:"status"`
	CurrentStep       int                 `gorm:"default:0" json:"current_step"`
	TotalSteps        int                 `gorm:"not null" json:"total_steps"`
	Context           string              `gorm:"type:text" json:"context"` // JSON序列化的上下文数据
	Steps             []SagaStepExecution `gorm:"foreignKey:SagaExecutionID" json:"steps,omitempty"`
	Error             string              `gorm:"type:text" json:"error,omitempty"`
	CompensationError string              `gorm:"type:text" json:"compensation_error,omitempty"`
	StartedAt         time.Time           `json:"started_at"`
	CompletedAt       *time.Time          `json:"completed_at,omitempty"`
	TimeoutAt         *time.Time          `json:"timeout_at,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	DeletedAt         gorm.DeletedAt      `gorm:"index" json:"-"`
}

// SagaStepExecution Saga步骤执行记录
type SagaStepExecution struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	SagaExecutionID uint          `gorm:"index;not null" json:"saga_execution_id"`
	SagaExecution   SagaExecution `gorm:"foreignKey:SagaExecutionID" json:"-"`
	StepName        string        `gorm:"not null" json:"step_name"`
	StepIndex       int           `gorm:"not null" json:"step_index"`
	Status          StepStatus    `gorm:"default:pending" json:"status"`
	Input           string        `gorm:"type:text" json:"input"`  // JSON序列化的输入数据
	Output          string        `gorm:"type:text" json:"output"` // JSON序列化的输出数据
	Error           string        `gorm:"type:text" json:"error,omitempty"`
	AttemptCount    int           `gorm:"default:0" json:"attempt_count"`
	MaxRetries      int           `gorm:"default:3" json:"max_retries"`
	StartedAt       *time.Time    `json:"started_at,omitempty"`
	CompletedAt     *time.Time    `json:"completed_at,omitempty"`
	CompensatedAt   *time.Time    `json:"compensated_at,omitempty"`
	NextRetryAt     *time.Time    `json:"next_retry_at,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// SagaStatus Saga执行状态枚举
type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"      // 等待执行
	SagaStatusRunning      SagaStatus = "running"      // 正在执行
	SagaStatusCompleted    SagaStatus = "completed"    // 执行完成
	SagaStatusCompensating SagaStatus = "compensating" // 正在补偿
	SagaStatusCompensated  SagaStatus = "compensated"  // 补偿完成
	SagaStatusFailed       SagaStatus = "failed"       // 执行失败
	SagaStatusTimeout      SagaStatus = "timeout"      // 执行超时
	SagaStatusCancelled    SagaStatus = "cancelled"    // 手动取消
)

// StepStatus 步骤执行状态枚举
type StepStatus string

const (
	StepStatusPending      StepStatus = "pending"      // 等待执行
	StepStatusRunning      StepStatus = "running"      // 正在执行
	StepStatusCompleted    StepStatus = "completed"    // 执行完成
	StepStatusFailed       StepStatus = "failed"       // 执行失败
	StepStatusCompensating StepStatus = "compensating" // 正在补偿
	StepStatusCompensated  StepStatus = "compensated"  // 补偿完成
	StepStatusSkipped      StepStatus = "skipped"      // 跳过执行
	StepStatusRetrying     StepStatus = "retrying"     // 正在重试
)

// SagaContext Saga执行上下文
type SagaContext struct {
	RequestID     string                 `json:"request_id"`
	UserID        uint                   `json:"user_id"`
	CorrelationID string                 `json:"correlation_id"`
	TraceID       string                 `json:"trace_id"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]string      `json:"metadata"`
}

// FileUploadSagaContext 文件上传Saga上下文
type FileUploadSagaContext struct {
	SagaContext
	FileName     string `json:"file_name"`
	FileSize     int64  `json:"file_size"`
	ContentType  string `json:"content_type"`
	ObjectKey    string `json:"object_key"`
	UploadToken  string `json:"upload_token"`
	FileID       uint   `json:"file_id,omitempty"`
	ThumbnailKey string `json:"thumbnail_key,omitempty"`
}

// FileDeleteSagaContext 文件删除Saga上下文
type FileDeleteSagaContext struct {
	SagaContext
	FileID       uint   `json:"file_id"`
	ObjectKey    string `json:"object_key"`
	ThumbnailKey string `json:"thumbnail_key,omitempty"`
	FileSize     int64  `json:"file_size"`
}

// SagaEvent Saga事件记录
type SagaEvent struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	SagaExecutionID uint          `gorm:"index;not null" json:"saga_execution_id"`
	SagaExecution   SagaExecution `gorm:"foreignKey:SagaExecutionID" json:"-"`
	EventType       string        `gorm:"not null" json:"event_type"` // started, step_started, step_completed, compensated, failed
	StepName        string        `json:"step_name,omitempty"`
	StepIndex       int           `json:"step_index,omitempty"`
	Duration        int64         `json:"duration"` // 执行时长(毫秒)
	Error           string        `gorm:"type:text" json:"error,omitempty"`
	EventData       string        `gorm:"type:text" json:"event_data,omitempty"` // JSON格式的事件数据
	TraceID         string        `json:"trace_id,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryPolicy 默认重试策略
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:   3,
		InitialDelay:  time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"network_error",
			"timeout_error",
			"temporary_error",
			"rate_limit_error",
		},
	}
}

// SagaMetrics Saga度量统计
type SagaMetrics struct {
	SagaType              string                   `json:"saga_type"`
	TotalExecutions       int64                    `json:"total_executions"`
	SuccessfulExecutions  int64                    `json:"successful_executions"`
	FailedExecutions      int64                    `json:"failed_executions"`
	CompensatedExecutions int64                    `json:"compensated_executions"`
	SuccessRate           float64                  `json:"success_rate"`
	AverageExecutionTime  time.Duration            `json:"average_execution_time"`
	AverageStepTime       map[string]time.Duration `json:"average_step_time"`
	StepFailureRate       map[string]float64       `json:"step_failure_rate"`
	LastUpdated           time.Time                `json:"last_updated"`
}

// IsTerminal 判断Saga状态是否为终态
func (s SagaStatus) IsTerminal() bool {
	switch s {
	case SagaStatusCompleted, SagaStatusCompensated, SagaStatusFailed, SagaStatusTimeout, SagaStatusCancelled:
		return true
	default:
		return false
	}
}

// IsSuccessful 判断Saga是否成功完成
func (s SagaStatus) IsSuccessful() bool {
	return s == SagaStatusCompleted
}

// NeedsCompensation 判断是否需要补偿
func (s SagaStatus) NeedsCompensation() bool {
	switch s {
	case SagaStatusFailed, SagaStatusTimeout:
		return true
	default:
		return false
	}
}

// IsTerminal 判断步骤状态是否为终态
func (s StepStatus) IsTerminal() bool {
	switch s {
	case StepStatusCompleted, StepStatusFailed, StepStatusCompensated, StepStatusSkipped:
		return true
	default:
		return false
	}
}

// IsSuccessful 判断步骤是否成功完成
func (s StepStatus) IsSuccessful() bool {
	return s == StepStatusCompleted
}

// CanRetry 判断步骤是否可以重试
func (s StepStatus) CanRetry() bool {
	return s == StepStatusFailed
}

// TableName 指定表名
func (SagaExecution) TableName() string {
	return "saga_executions"
}

func (SagaStepExecution) TableName() string {
	return "saga_step_executions"
}

func (SagaEvent) TableName() string {
	return "saga_events"
}
