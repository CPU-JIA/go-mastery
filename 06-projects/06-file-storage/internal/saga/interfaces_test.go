package saga

import (
	"context"
	"errors"
	"file-storage-service/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSagaError(t *testing.T) {
	t.Run("Should create saga error with all fields", func(t *testing.T) {
		sagaErr := NewSagaError("TEST_ERROR", "test message", 123, 2)

		assert.Equal(t, "TEST_ERROR", sagaErr.Code)
		assert.Equal(t, "test message", sagaErr.Message)
		assert.Equal(t, uint(123), sagaErr.SagaID)
		assert.Equal(t, 2, sagaErr.StepIndex)
		assert.NotNil(t, sagaErr.Metadata)
	})

	t.Run("Should format error message correctly", func(t *testing.T) {
		// Saga error with step
		sagaErr := NewSagaError("TEST_ERROR", "test message", 123, 2)
		expected := "saga error [TEST_ERROR] in saga 123 step 2: test message"
		assert.Equal(t, expected, sagaErr.Error())

		// Saga error without step
		sagaErr2 := NewSagaError("TEST_ERROR", "test message", 123, -1)
		expected2 := "saga error [TEST_ERROR] in saga 123: test message"
		assert.Equal(t, expected2, sagaErr2.Error())

		// General saga error
		sagaErr3 := NewSagaError("TEST_ERROR", "test message", 0, -1)
		expected3 := "saga error [TEST_ERROR]: test message"
		assert.Equal(t, expected3, sagaErr3.Error())
	})

	t.Run("Should create saga error with cause", func(t *testing.T) {
		originalErr := errors.New("original error")
		sagaErr := NewSagaErrorWithCause("TEST_ERROR", "test message", 123, 2, originalErr)

		assert.Equal(t, originalErr, sagaErr.Cause)
		assert.Equal(t, originalErr, sagaErr.Unwrap())
	})

	t.Run("Should add metadata", func(t *testing.T) {
		sagaErr := NewSagaError("TEST_ERROR", "test message", 123, 2)
		sagaErr.WithMetadata("key1", "value1")
		sagaErr.WithMetadata("key2", 42)

		assert.Equal(t, "value1", sagaErr.Metadata["key1"])
		assert.Equal(t, 42, sagaErr.Metadata["key2"])
	})

	t.Run("Should check error types correctly", func(t *testing.T) {
		sagaErr1 := NewSagaError("TEST_ERROR", "test message", 123, 2)
		sagaErr2 := NewSagaError("TEST_ERROR", "different message", 456, 3)
		sagaErr3 := NewSagaError("OTHER_ERROR", "test message", 123, 2)

		assert.True(t, sagaErr1.Is(sagaErr2))
		assert.False(t, sagaErr1.Is(sagaErr3))
	})
}

func TestIsRetryableError(t *testing.T) {
	t.Run("Should identify retryable saga errors", func(t *testing.T) {
		// Timeout errors should be retryable
		timeoutErr := NewSagaError(ErrCodeSagaTimeout, "timeout", 123, 1)
		assert.True(t, IsRetryableError(timeoutErr))

		stepTimeoutErr := NewSagaError(ErrCodeStepTimeout, "step timeout", 123, 1)
		assert.True(t, IsRetryableError(stepTimeoutErr))

		// Step execution failed with retryable cause
		networkErr := errors.New("network connection refused")
		stepErr := NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "step failed", 123, 1, networkErr)
		assert.True(t, IsRetryableError(stepErr))
	})

	t.Run("Should identify non-retryable saga errors", func(t *testing.T) {
		// Invalid context should not be retryable
		contextErr := NewSagaError(ErrCodeInvalidContext, "invalid context", 123, 1)
		assert.False(t, IsRetryableError(contextErr))

		// Compensation impossible should not be retryable
		compensationErr := NewSagaError(ErrCodeCompensationImpossible, "compensation impossible", 123, 1)
		assert.False(t, IsRetryableError(compensationErr))
	})

	t.Run("Should identify retryable temporary errors", func(t *testing.T) {
		temporaryErrors := []error{
			errors.New("connection timeout"),
			errors.New("network connection refused"),
			errors.New("temporary failure"),
			errors.New("rate limit exceeded"),
			errors.New("too many requests"),
		}

		for _, err := range temporaryErrors {
			assert.True(t, IsRetryableError(err), "Expected %v to be retryable", err)
		}
	})

	t.Run("Should identify non-retryable errors", func(t *testing.T) {
		nonRetryableErrors := []error{
			errors.New("validation failed"),
			errors.New("permission denied"),
			errors.New("resource not found"),
		}

		for _, err := range nonRetryableErrors {
			assert.False(t, IsRetryableError(err), "Expected %v to be non-retryable", err)
		}
	})
}

func TestSagaMetrics(t *testing.T) {
	t.Run("Should calculate success rate correctly", func(t *testing.T) {
		metrics := &SagaMetrics{
			TotalExecutions:       100,
			SuccessfulExecutions:  85,
			FailedExecutions:      10,
			CompensatedExecutions: 5,
		}

		successRate := metrics.CalculateSuccessRate()
		assert.Equal(t, 85.0, successRate)
	})

	t.Run("Should handle zero executions", func(t *testing.T) {
		metrics := &SagaMetrics{
			TotalExecutions:      0,
			SuccessfulExecutions: 0,
		}

		successRate := metrics.CalculateSuccessRate()
		assert.Equal(t, 0.0, successRate)
	})
}

// Mock implementations for testing interface compliance

type MockSagaEngine struct {
	definitions map[string]SagaDefinition
	executions  map[uint]*model.SagaExecution
}

func NewMockSagaEngine() *MockSagaEngine {
	return &MockSagaEngine{
		definitions: make(map[string]SagaDefinition),
		executions:  make(map[uint]*model.SagaExecution),
	}
}

func (m *MockSagaEngine) StartSaga(ctx context.Context, sagaType string, sagaContext interface{}) (*model.SagaExecution, error) {
	if _, exists := m.definitions[sagaType]; !exists {
		return nil, NewSagaError(ErrCodeDefinitionNotFound, "definition not found", 0, -1)
	}

	execution := &model.SagaExecution{
		ID:       uint(len(m.executions) + 1),
		SagaType: sagaType,
		Status:   model.SagaStatusPending,
	}

	m.executions[execution.ID] = execution
	return execution, nil
}

func (m *MockSagaEngine) ResumeSaga(ctx context.Context, sagaID uint) error {
	if _, exists := m.executions[sagaID]; !exists {
		return NewSagaError(ErrCodeSagaNotFound, "saga not found", sagaID, -1)
	}
	return nil
}

func (m *MockSagaEngine) CancelSaga(ctx context.Context, sagaID uint) error {
	execution, exists := m.executions[sagaID]
	if !exists {
		return NewSagaError(ErrCodeSagaNotFound, "saga not found", sagaID, -1)
	}

	if execution.Status.IsTerminal() {
		return NewSagaError(ErrCodeSagaAlreadyCompleted, "saga already completed", sagaID, -1)
	}

	execution.Status = model.SagaStatusCancelled
	return nil
}

func (m *MockSagaEngine) GetSagaStatus(ctx context.Context, sagaID uint) (*model.SagaExecution, error) {
	execution, exists := m.executions[sagaID]
	if !exists {
		return nil, NewSagaError(ErrCodeSagaNotFound, "saga not found", sagaID, -1)
	}
	return execution, nil
}

func (m *MockSagaEngine) RegisterSagaDefinition(sagaType string, definition SagaDefinition) error {
	if _, exists := m.definitions[sagaType]; exists {
		return NewSagaError(ErrCodeDefinitionExists, "definition already exists", 0, -1)
	}
	m.definitions[sagaType] = definition
	return nil
}

func (m *MockSagaEngine) ListRunningSagas(ctx context.Context, limit int) ([]*model.SagaExecution, error) {
	var result []*model.SagaExecution
	count := 0

	for _, execution := range m.executions {
		if execution.Status == model.SagaStatusRunning {
			result = append(result, execution)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return result, nil
}

type MockSagaDefinition struct {
	sagaType    string
	steps       []SagaStepDefinition
	timeout     time.Duration
	retryPolicy *model.RetryPolicy
}

func NewMockSagaDefinition(sagaType string) *MockSagaDefinition {
	return &MockSagaDefinition{
		sagaType:    sagaType,
		steps:       []SagaStepDefinition{},
		timeout:     DefaultSagaTimeout,
		retryPolicy: model.DefaultRetryPolicy(),
	}
}

func (m *MockSagaDefinition) GetSagaType() string {
	return m.sagaType
}

func (m *MockSagaDefinition) GetSteps() []SagaStepDefinition {
	return m.steps
}

func (m *MockSagaDefinition) ValidateContext(context interface{}) error {
	if context == nil {
		return NewSagaError(ErrCodeInvalidContext, "context cannot be nil", 0, -1)
	}
	return nil
}

func (m *MockSagaDefinition) CalculateTimeout(context interface{}) time.Duration {
	return m.timeout
}

func (m *MockSagaDefinition) GetRetryPolicy() *model.RetryPolicy {
	return m.retryPolicy
}

func (m *MockSagaDefinition) AddStep(name string, executeFunc SagaStepExecuteFunc, compensateFunc SagaStepCompensateFunc) {
	step := SagaStepDefinition{
		Name:       name,
		Execute:    executeFunc,
		Compensate: compensateFunc,
	}
	m.steps = append(m.steps, step)
}

func TestMockSagaEngine(t *testing.T) {
	t.Run("Should register and start saga", func(t *testing.T) {
		engine := NewMockSagaEngine()
		definition := NewMockSagaDefinition("test-saga")

		// Register definition
		err := engine.RegisterSagaDefinition("test-saga", definition)
		require.NoError(t, err)

		// Try to register again (should fail)
		err = engine.RegisterSagaDefinition("test-saga", definition)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		// Start saga
		ctx := context.Background()
		execution, err := engine.StartSaga(ctx, "test-saga", map[string]string{"key": "value"})
		require.NoError(t, err)
		assert.Equal(t, "test-saga", execution.SagaType)
		assert.Equal(t, model.SagaStatusPending, execution.Status)

		// Get saga status
		status, err := engine.GetSagaStatus(ctx, execution.ID)
		require.NoError(t, err)
		assert.Equal(t, execution.ID, status.ID)
	})

	t.Run("Should handle saga not found", func(t *testing.T) {
		engine := NewMockSagaEngine()
		ctx := context.Background()

		// Try to start saga without definition
		_, err := engine.StartSaga(ctx, "non-existent", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "definition not found")

		// Try to get non-existent saga
		_, err = engine.GetSagaStatus(ctx, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "saga not found")

		// Try to cancel non-existent saga
		err = engine.CancelSaga(ctx, 999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "saga not found")
	})

	t.Run("Should cancel running saga", func(t *testing.T) {
		engine := NewMockSagaEngine()
		definition := NewMockSagaDefinition("test-saga")

		err := engine.RegisterSagaDefinition("test-saga", definition)
		require.NoError(t, err)

		ctx := context.Background()
		execution, err := engine.StartSaga(ctx, "test-saga", nil)
		require.NoError(t, err)

		// Cancel saga
		err = engine.CancelSaga(ctx, execution.ID)
		require.NoError(t, err)

		// Check status
		status, err := engine.GetSagaStatus(ctx, execution.ID)
		require.NoError(t, err)
		assert.Equal(t, model.SagaStatusCancelled, status.Status)

		// Try to cancel again (should fail)
		err = engine.CancelSaga(ctx, execution.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already completed")
	})
}

func TestMockSagaDefinition(t *testing.T) {
	t.Run("Should create definition with steps", func(t *testing.T) {
		definition := NewMockSagaDefinition("test-saga")

		executeFunc := func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
			return "success", nil
		}

		compensateFunc := func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
			return nil
		}

		definition.AddStep("step1", executeFunc, compensateFunc)
		definition.AddStep("step2", executeFunc, compensateFunc)

		assert.Equal(t, "test-saga", definition.GetSagaType())
		assert.Len(t, definition.GetSteps(), 2)
		assert.Equal(t, "step1", definition.GetSteps()[0].Name)
		assert.Equal(t, "step2", definition.GetSteps()[1].Name)
	})

	t.Run("Should validate context", func(t *testing.T) {
		definition := NewMockSagaDefinition("test-saga")

		// Valid context
		err := definition.ValidateContext(map[string]string{"key": "value"})
		assert.NoError(t, err)

		// Invalid context (nil)
		err = definition.ValidateContext(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("Should return timeout and retry policy", func(t *testing.T) {
		definition := NewMockSagaDefinition("test-saga")

		timeout := definition.CalculateTimeout(nil)
		assert.Equal(t, DefaultSagaTimeout, timeout)

		policy := definition.GetRetryPolicy()
		assert.NotNil(t, policy)
		assert.Equal(t, 3, policy.MaxAttempts)
	})
}
