package saga

import (
	"context"
	"errors"
	"file-storage-service/internal/model"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestSagaEngine(t *testing.T) (SagaExecutionEngine, *gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(
		&model.SagaExecution{},
		&model.SagaStepExecution{},
		&model.SagaEvent{},
		&model.User{},
	)
	require.NoError(t, err)

	engine := NewSagaExecutionEngine(db)

	cleanup := func() {
		// 清理资源
		db.Exec("DELETE FROM saga_executions")
		db.Exec("DELETE FROM saga_step_executions")
		db.Exec("DELETE FROM saga_events")
	}

	return engine, db, cleanup
}

func createTestSagaExecution(t *testing.T, db *gorm.DB, status model.SagaStatus) *model.SagaExecution {
	execution := &model.SagaExecution{
		SagaType:    "test-saga",
		RequestID:   fmt.Sprintf("test-request-%d-%s", time.Now().UnixNano(), t.Name()),
		UserID:      1,
		Status:      status,
		CurrentStep: 0,
		TotalSteps:  3,
		Context:     `{"user_id": 1, "file_name": "test.jpg"}`,
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := db.Create(execution).Error
	require.NoError(t, err)
	return execution
}

// MockSagaStepDefinitions 创建测试用的步骤定义
func createMockStepDefinitions() []SagaStepDefinition {
	return []SagaStepDefinition{
		{
			Name: "validate_file",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return map[string]interface{}{"validated": true}, nil
			},
			Compensate: func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				return nil // 验证步骤不需要补偿
			},
			IsIdempotent: true,
		},
		{
			Name: "upload_to_storage",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return map[string]interface{}{"object_key": "users/1/test.jpg"}, nil
			},
			Compensate: func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				// 模拟删除上传的文件
				return nil
			},
			IsCritical:   true,
			IsIdempotent: true,
		},
		{
			Name: "update_database",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return map[string]interface{}{"file_id": 123}, nil
			},
			Compensate: func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				// 模拟删除数据库记录
				return nil
			},
			IsCritical:   true,
			IsIdempotent: true,
		},
	}
}

// MockFailingStepDefinitions 创建会失败的步骤定义
func createFailingStepDefinitions() []SagaStepDefinition {
	return []SagaStepDefinition{
		{
			Name: "validate_file",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return map[string]interface{}{"validated": true}, nil
			},
			Compensate: func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				return nil
			},
			IsIdempotent: true,
		},
		{
			Name: "upload_to_storage",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return nil, errors.New("storage service unavailable")
			},
			Compensate: func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				return nil
			},
			IsCritical:   true,
			IsIdempotent: true,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   2, // 只重试2次
				InitialDelay:  10 * time.Millisecond,
				MaxDelay:      100 * time.Millisecond,
				BackoffFactor: 2.0,
			},
		},
	}
}

func TestSagaExecutionEngine(t *testing.T) {
	engine, db, cleanup := setupTestSagaEngine(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Should execute saga successfully", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusPending)
		definition := NewMockSagaDefinition("test-saga")

		// 添加测试步骤
		steps := createMockStepDefinitions()
		for _, step := range steps {
			definition.AddStep(step.Name, step.Execute, step.Compensate)
		}

		err := engine.ExecuteSaga(ctx, execution, definition)
		assert.NoError(t, err)

		// 验证Saga状态
		var updatedExecution model.SagaExecution
		err = db.First(&updatedExecution, execution.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.SagaStatusCompleted, updatedExecution.Status)
		assert.Equal(t, 3, updatedExecution.CurrentStep)

		// 验证所有步骤都完成了
		var steps_result []*model.SagaStepExecution
		err = db.Where("saga_execution_id = ?", execution.ID).Find(&steps_result).Error
		require.NoError(t, err)
		assert.Len(t, steps_result, 3)

		for _, step := range steps_result {
			assert.Equal(t, model.StepStatusCompleted, step.Status)
			assert.NotEmpty(t, step.Output)
		}
	})

	t.Run("Should handle step failure with compensation", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusPending)
		definition := NewMockSagaDefinition("failing-saga")

		// 添加会失败的步骤
		steps := createFailingStepDefinitions()
		for _, step := range steps {
			definition.AddStep(step.Name, step.Execute, step.Compensate)
		}

		err := engine.ExecuteSaga(ctx, execution, definition)
		assert.Error(t, err) // 应该失败

		// 验证Saga状态
		var updatedExecution model.SagaExecution
		err = db.First(&updatedExecution, execution.ID).Error
		require.NoError(t, err)

		// 应该进入补偿状态或失败状态
		assert.True(t, updatedExecution.Status == model.SagaStatusCompensated ||
			updatedExecution.Status == model.SagaStatusFailed)
		assert.NotEmpty(t, updatedExecution.Error)

		// 验证第一步完成，第二步失败
		var steps_result []*model.SagaStepExecution
		err = db.Where("saga_execution_id = ?", execution.ID).
			Order("step_index ASC").Find(&steps_result).Error
		require.NoError(t, err)
		assert.Len(t, steps_result, 2)

		assert.Equal(t, model.StepStatusCompleted, steps_result[0].Status) // 第一步成功
		assert.True(t, steps_result[1].Status == model.StepStatusFailed ||
			steps_result[1].Status == model.StepStatusCompensated) // 第二步失败
	})

	t.Run("Should handle saga timeout", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusPending)

		// 创建会超时的定义
		definition := &timeoutSagaDefinition{
			sagaType: "timeout-saga",
			timeout:  50 * time.Millisecond, // 很短的超时时间
		}

		err := engine.ExecuteSaga(ctx, execution, definition)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")

		// 验证Saga状态
		var updatedExecution model.SagaExecution
		err = db.First(&updatedExecution, execution.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.SagaStatusTimeout, updatedExecution.Status)
	})

	t.Run("Should handle concurrent saga execution", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusPending)
		definition := NewMockSagaDefinition("concurrent-saga")

		// 添加简单步骤
		definition.AddStep("step1",
			func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				time.Sleep(100 * time.Millisecond) // 模拟长时间执行
				return "result", nil
			},
			func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				return nil
			})

		// 并发执行同一个Saga
		done := make(chan error, 2)

		go func() {
			err := engine.ExecuteSaga(ctx, execution, definition)
			done <- err
		}()

		go func() {
			time.Sleep(10 * time.Millisecond) // 稍微延迟
			err := engine.ExecuteSaga(ctx, execution, definition)
			done <- err
		}()

		// 收集结果
		results := make([]error, 2)
		for i := 0; i < 2; i++ {
			results[i] = <-done
		}

		// 应该有一个成功，一个因为锁失败
		successCount := 0
		lockFailureCount := 0

		for _, err := range results {
			if err == nil {
				successCount++
			} else if sagaErr, ok := err.(*SagaError); ok && sagaErr.Code == ErrCodeSagaLockFailed {
				lockFailureCount++
			}
		}

		assert.Equal(t, 1, successCount, "Exactly one execution should succeed")
		assert.Equal(t, 1, lockFailureCount, "Exactly one execution should fail with lock error")
	})

	t.Run("Should handle step prerequisite checking", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusPending)
		definition := &prerequisiteSagaDefinition{
			sagaType: "prerequisite-saga",
		}

		// 这个测试需要更复杂的设置来验证前置条件
		// 暂时验证不会panic
		assert.NotPanics(t, func() {
			engine.ExecuteSaga(ctx, execution, definition)
		})
	})
}

func TestSagaExecutionEngineStepExecution(t *testing.T) {
	engine, db, cleanup := setupTestSagaEngine(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Should execute single step successfully", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusRunning)

		// 创建步骤执行记录
		stepExecution := &model.SagaStepExecution{
			SagaExecutionID: execution.ID,
			StepName:        "test_step",
			StepIndex:       0,
			Status:          model.StepStatusRunning,
			Input:           `{"test": "input"}`,
			AttemptCount:    0,
			MaxRetries:      3,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := db.Create(stepExecution).Error
		require.NoError(t, err)

		// 定义步骤
		stepDef := SagaStepDefinition{
			Name: "test_step",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return map[string]interface{}{"success": true}, nil
			},
		}

		sagaContext := map[string]interface{}{"user_id": 1}

		err = engine.ExecuteStep(ctx, stepExecution, stepDef, sagaContext)
		assert.NoError(t, err)

		// 验证步骤状态
		var updatedStep model.SagaStepExecution
		err = db.First(&updatedStep, stepExecution.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StepStatusCompleted, updatedStep.Status)
		assert.NotEmpty(t, updatedStep.Output)
		assert.Contains(t, updatedStep.Output, "success")
	})

	t.Run("Should handle step failure", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusRunning)

		stepExecution := &model.SagaStepExecution{
			SagaExecutionID: execution.ID,
			StepName:        "failing_step",
			StepIndex:       0,
			Status:          model.StepStatusRunning,
			Input:           `{"test": "input"}`,
			AttemptCount:    0,
			MaxRetries:      3,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := db.Create(stepExecution).Error
		require.NoError(t, err)

		stepDef := SagaStepDefinition{
			Name: "failing_step",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return nil, errors.New("step execution failed")
			},
		}

		sagaContext := map[string]interface{}{"user_id": 1}

		err = engine.ExecuteStep(ctx, stepExecution, stepDef, sagaContext)
		assert.Error(t, err)

		// 验证步骤状态
		var updatedStep model.SagaStepExecution
		err = db.First(&updatedStep, stepExecution.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StepStatusFailed, updatedStep.Status)
		assert.NotEmpty(t, updatedStep.Error)
	})

	t.Run("Should compensate step successfully", func(t *testing.T) {
		execution := createTestSagaExecution(t, db, model.SagaStatusCompensating)

		stepExecution := &model.SagaStepExecution{
			SagaExecutionID: execution.ID,
			StepName:        "compensatable_step",
			StepIndex:       0,
			Status:          model.StepStatusCompleted,
			Input:           `{"test": "input"}`,
			Output:          `{"result": "success"}`,
			AttemptCount:    1,
			MaxRetries:      3,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := db.Create(stepExecution).Error
		require.NoError(t, err)

		stepDef := SagaStepDefinition{
			Name: "compensatable_step",
			Compensate: func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
				// 模拟补偿操作成功
				return nil
			},
			IsCritical: true,
		}

		sagaContext := map[string]interface{}{"user_id": 1}

		err = engine.CompensateStep(ctx, stepExecution, stepDef, sagaContext)
		assert.NoError(t, err)

		// 验证步骤状态
		var updatedStep model.SagaStepExecution
		err = db.First(&updatedStep, stepExecution.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StepStatusCompensated, updatedStep.Status)
	})
}

// 辅助测试结构

// timeoutSagaDefinition 会超时的Saga定义
type timeoutSagaDefinition struct {
	sagaType string
	timeout  time.Duration
}

func (t *timeoutSagaDefinition) GetSagaType() string {
	return t.sagaType
}

func (t *timeoutSagaDefinition) GetSteps() []SagaStepDefinition {
	return []SagaStepDefinition{
		{
			Name: "slow_step",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				time.Sleep(200 * time.Millisecond) // 比超时时间长
				return "result", nil
			},
		},
	}
}

func (t *timeoutSagaDefinition) ValidateContext(context interface{}) error {
	return nil
}

func (t *timeoutSagaDefinition) CalculateTimeout(context interface{}) time.Duration {
	return t.timeout
}

func (t *timeoutSagaDefinition) GetRetryPolicy() *model.RetryPolicy {
	return model.DefaultRetryPolicy()
}

// prerequisiteSagaDefinition 有前置条件的Saga定义
type prerequisiteSagaDefinition struct {
	sagaType string
}

func (p *prerequisiteSagaDefinition) GetSagaType() string {
	return p.sagaType
}

func (p *prerequisiteSagaDefinition) ValidateContext(context interface{}) error {
	return nil
}

func (p *prerequisiteSagaDefinition) CalculateTimeout(context interface{}) time.Duration {
	return DefaultSagaTimeout
}

func (p *prerequisiteSagaDefinition) GetRetryPolicy() *model.RetryPolicy {
	return model.DefaultRetryPolicy()
}

func (p *prerequisiteSagaDefinition) GetSteps() []SagaStepDefinition {
	return []SagaStepDefinition{
		{
			Name: "step1",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return "step1_result", nil
			},
		},
		{
			Name: "step2",
			Execute: func(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
				return "step2_result", nil
			},
			Prerequisites: []string{"step1"}, // step2依赖step1
		},
	}
}

func TestWorkerPool(t *testing.T) {
	t.Run("Should execute jobs concurrently", func(t *testing.T) {
		pool := NewWorkerPool(3)
		defer pool.Close()

		counter := 0
		done := make(chan bool, 5)

		// 提交多个任务
		for i := 0; i < 5; i++ {
			pool.Submit(func() {
				counter++
				time.Sleep(10 * time.Millisecond)
				done <- true
			})
		}

		// 等待所有任务完成
		for i := 0; i < 5; i++ {
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("Task execution timeout")
			}
		}

		assert.Equal(t, 5, counter)
	})

	t.Run("Should handle panic in jobs", func(t *testing.T) {
		pool := NewWorkerPool(1)
		defer pool.Close()

		done := make(chan bool, 2)

		// 提交一个会panic的任务
		pool.Submit(func() {
			panic("test panic")
		})

		// 提交一个正常任务
		pool.Submit(func() {
			done <- true
		})

		// 正常任务应该仍能执行
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Normal task should execute even after panic")
		}
	})
}

func TestSagaRepository(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.SagaExecution{}, &model.SagaStepExecution{}, &model.SagaEvent{})
	require.NoError(t, err)

	repo := NewSagaRepository(db)
	ctx := context.Background()

	t.Run("Should create and retrieve saga execution", func(t *testing.T) {
		execution := &model.SagaExecution{
			SagaType:    "test-saga",
			RequestID:   "test-request-repo",
			UserID:      1,
			Status:      model.SagaStatusPending,
			CurrentStep: 0,
			TotalSteps:  2,
			Context:     `{"test": true}`,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := repo.CreateSagaExecution(ctx, execution)
		assert.NoError(t, err)
		assert.NotZero(t, execution.ID)

		retrieved, err := repo.GetSagaExecution(ctx, execution.ID)
		assert.NoError(t, err)
		assert.Equal(t, execution.SagaType, retrieved.SagaType)
		assert.Equal(t, execution.RequestID, retrieved.RequestID)
	})

	t.Run("Should get running sagas", func(t *testing.T) {
		// 创建运行中的Saga
		execution1 := &model.SagaExecution{
			SagaType:  "running-saga-1",
			RequestID: "running-request-1",
			Status:    model.SagaStatusRunning,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		execution2 := &model.SagaExecution{
			SagaType:  "running-saga-2",
			RequestID: "running-request-2",
			Status:    model.SagaStatusRunning,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.CreateSagaExecution(ctx, execution1)
		require.NoError(t, err)
		err = repo.CreateSagaExecution(ctx, execution2)
		require.NoError(t, err)

		runningSagas, err := repo.GetRunningSagas(ctx, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(runningSagas), 2) // 至少有我们创建的2个
	})
}
