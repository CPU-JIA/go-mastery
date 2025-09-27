package saga

import (
	"errors"
	"fmt"
	"time"
)

// 预定义的错误类型
var (
	// 系统级错误
	ErrSagaEngineNotInitialized = errors.New("saga engine not initialized")
	ErrSagaDefinitionNotFound   = errors.New("saga definition not found")
	ErrSagaDefinitionExists     = errors.New("saga definition already exists")

	// 执行错误
	ErrSagaNotFound            = errors.New("saga execution not found")
	ErrSagaAlreadyRunning      = errors.New("saga is already running")
	ErrSagaAlreadyCompleted    = errors.New("saga is already completed")
	ErrSagaLockAcquisitionFail = errors.New("failed to acquire saga lock")
	ErrSagaTimeout             = errors.New("saga execution timeout")

	// 步骤执行错误
	ErrStepNotFound           = errors.New("step not found")
	ErrStepExecutionFailed    = errors.New("step execution failed")
	ErrStepCompensationFailed = errors.New("step compensation failed")
	ErrStepTimeout            = errors.New("step execution timeout")
	ErrStepPrerequisiteFailed = errors.New("step prerequisite not satisfied")

	// 状态转换错误
	ErrInvalidStateTransition = errors.New("invalid state transition")
	ErrStateConflict          = errors.New("state conflict detected")

	// 重试和补偿错误
	ErrMaxRetriesExceeded     = errors.New("maximum retries exceeded")
	ErrCompensationRequired   = errors.New("compensation required")
	ErrCompensationImpossible = errors.New("compensation is not possible")

	// 上下文和数据错误
	ErrInvalidSagaContext    = errors.New("invalid saga context")
	ErrInvalidStepInput      = errors.New("invalid step input")
	ErrSerializationFailed   = errors.New("serialization failed")
	ErrDeserializationFailed = errors.New("deserialization failed")
)

// SagaError 统一的Saga错误类型
type SagaError struct {
	Code      string                 // 错误代码
	Message   string                 // 错误消息
	SagaID    uint                   // 关联的Saga ID
	StepIndex int                    // 关联的步骤索引（-1表示不是步骤错误）
	Cause     error                  // 原始错误
	Metadata  map[string]interface{} // 额外的错误元数据
}

// Error 实现error接口
func (e *SagaError) Error() string {
	if e.SagaID > 0 {
		if e.StepIndex >= 0 {
			return fmt.Sprintf("saga error [%s] in saga %d step %d: %s", e.Code, e.SagaID, e.StepIndex, e.Message)
		}
		return fmt.Sprintf("saga error [%s] in saga %d: %s", e.Code, e.SagaID, e.Message)
	}
	return fmt.Sprintf("saga error [%s]: %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *SagaError) Unwrap() error {
	return e.Cause
}

// Is 检查错误类型
func (e *SagaError) Is(target error) bool {
	if sagaErr, ok := target.(*SagaError); ok {
		return e.Code == sagaErr.Code
	}
	return false
}

// NewSagaError 创建新的Saga错误
func NewSagaError(code, message string, sagaID uint, stepIndex int) *SagaError {
	return &SagaError{
		Code:      code,
		Message:   message,
		SagaID:    sagaID,
		StepIndex: stepIndex,
		Metadata:  make(map[string]interface{}),
	}
}

// NewSagaErrorWithCause 创建带原因的Saga错误
func NewSagaErrorWithCause(code, message string, sagaID uint, stepIndex int, cause error) *SagaError {
	return &SagaError{
		Code:      code,
		Message:   message,
		SagaID:    sagaID,
		StepIndex: stepIndex,
		Cause:     cause,
		Metadata:  make(map[string]interface{}),
	}
}

// WithMetadata 添加元数据
func (e *SagaError) WithMetadata(key string, value interface{}) *SagaError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// 错误代码常量
const (
	// 系统错误代码
	ErrCodeEngineNotInitialized = "ENGINE_NOT_INITIALIZED"
	ErrCodeDefinitionNotFound   = "DEFINITION_NOT_FOUND"
	ErrCodeDefinitionExists     = "DEFINITION_EXISTS"

	// 执行错误代码
	ErrCodeSagaNotFound         = "SAGA_NOT_FOUND"
	ErrCodeSagaAlreadyRunning   = "SAGA_ALREADY_RUNNING"
	ErrCodeSagaAlreadyCompleted = "SAGA_ALREADY_COMPLETED"
	ErrCodeSagaLockFailed       = "SAGA_LOCK_FAILED"
	ErrCodeSagaTimeout          = "SAGA_TIMEOUT"

	// 步骤错误代码
	ErrCodeStepNotFound           = "STEP_NOT_FOUND"
	ErrCodeStepExecutionFailed    = "STEP_EXECUTION_FAILED"
	ErrCodeStepCompensationFailed = "STEP_COMPENSATION_FAILED"
	ErrCodeStepTimeout            = "STEP_TIMEOUT"
	ErrCodeStepPrerequisiteFailed = "STEP_PREREQUISITE_FAILED"

	// 状态错误代码
	ErrCodeInvalidStateTransition = "INVALID_STATE_TRANSITION"
	ErrCodeStateConflict          = "STATE_CONFLICT"

	// 重试和补偿错误代码
	ErrCodeMaxRetriesExceeded     = "MAX_RETRIES_EXCEEDED"
	ErrCodeCompensationRequired   = "COMPENSATION_REQUIRED"
	ErrCodeCompensationImpossible = "COMPENSATION_IMPOSSIBLE"

	// 数据错误代码
	ErrCodeInvalidContext        = "INVALID_CONTEXT"
	ErrCodeInvalidInput          = "INVALID_INPUT"
	ErrCodeSerializationFailed   = "SERIALIZATION_FAILED"
	ErrCodeDeserializationFailed = "DESERIALIZATION_FAILED"
)

// 常量定义
const (
	// 默认超时配置
	DefaultSagaTimeout = 30 * time.Minute
	DefaultStepTimeout = 5 * time.Minute
	DefaultLockTimeout = 10 * time.Minute

	// 默认重试配置
	DefaultMaxRetries    = 3
	DefaultInitialDelay  = time.Second
	DefaultMaxDelay      = 30 * time.Second
	DefaultBackoffFactor = 2.0

	// 系统限制
	MaxSagaSteps       = 50
	MaxSagaContextSize = 10 * 1024 * 1024 // 10MB
	MaxStepInputSize   = 1 * 1024 * 1024  // 1MB
	MaxConcurrentSagas = 1000

	// 事件类型
	EventSagaStarted      = "saga_started"
	EventSagaCompleted    = "saga_completed"
	EventSagaFailed       = "saga_failed"
	EventSagaCancelled    = "saga_cancelled"
	EventSagaTimeout      = "saga_timeout"
	EventStepStarted      = "step_started"
	EventStepCompleted    = "step_completed"
	EventStepFailed       = "step_failed"
	EventStepRetrying     = "step_retrying"
	EventStepCompensating = "step_compensating"
	EventStepCompensated  = "step_compensated"
)

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	if sagaErr, ok := err.(*SagaError); ok {
		switch sagaErr.Code {
		case ErrCodeSagaTimeout, ErrCodeStepTimeout:
			return true
		case ErrCodeStepExecutionFailed:
			// 检查原因是否为网络错误等临时错误
			return isTemporaryError(sagaErr.Cause)
		}
	}

	// 检查常见的临时错误
	return isTemporaryError(err)
}

// isTemporaryError 检查是否为临时错误
func isTemporaryError(err error) bool {
	if err == nil {
		return false
	}

	// 检查错误消息中的临时错误指示符
	message := err.Error()
	temporaryIndicators := []string{
		"timeout", "connection refused", "connection reset",
		"network", "temporary", "rate limit", "too many requests",
	}

	for _, indicator := range temporaryIndicators {
		if contains(message, indicator) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0))
}

// indexOf 查找子串位置（简化实现）
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// SagaMetrics Saga度量数据
type SagaMetrics struct {
	TotalExecutions       int64         `json:"total_executions"`
	SuccessfulExecutions  int64         `json:"successful_executions"`
	FailedExecutions      int64         `json:"failed_executions"`
	CompensatedExecutions int64         `json:"compensated_executions"`
	AverageExecutionTime  time.Duration `json:"average_execution_time"`
	SuccessRate           float64       `json:"success_rate"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// CalculateSuccessRate 计算成功率
func (m *SagaMetrics) CalculateSuccessRate() float64 {
	if m.TotalExecutions == 0 {
		return 0
	}
	return float64(m.SuccessfulExecutions) / float64(m.TotalExecutions) * 100
}
