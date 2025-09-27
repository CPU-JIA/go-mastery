package saga

import (
	"context"
	"file-storage-service/internal/model"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	return db
}

func createTestSaga(t *testing.T, db *gorm.DB) *model.SagaExecution {
	saga := &model.SagaExecution{
		SagaType:    "test-saga",
		RequestID:   fmt.Sprintf("test-request-%d-%s", time.Now().UnixNano(), t.Name()),
		UserID:      1,
		Status:      model.SagaStatusPending,
		CurrentStep: 0,
		TotalSteps:  3,
		Context:     `{"test": "context"}`,
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := db.Create(saga).Error
	require.NoError(t, err)
	return saga
}

func createTestStep(t *testing.T, db *gorm.DB, sagaID uint, stepIndex int) *model.SagaStepExecution {
	step := &model.SagaStepExecution{
		SagaExecutionID: sagaID,
		StepName:        "test-step",
		StepIndex:       stepIndex,
		Status:          model.StepStatusPending,
		Input:           `{"input": "data"}`,
		AttemptCount:    0,
		MaxRetries:      3,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := db.Create(step).Error
	require.NoError(t, err)
	return step
}

func TestSagaStateManager(t *testing.T) {
	db := setupTestDB(t)
	stateManager := NewSagaStateManager(db)
	ctx := context.Background()

	t.Run("Should transition saga status successfully", func(t *testing.T) {
		saga := createTestSaga(t, db)

		// 测试合法的状态转换：pending -> running
		err := stateManager.TransitionSagaStatus(ctx, saga.ID, model.SagaStatusPending, model.SagaStatusRunning)
		assert.NoError(t, err)

		// 验证状态已更新
		var updatedSaga model.SagaExecution
		err = db.First(&updatedSaga, saga.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.SagaStatusRunning, updatedSaga.Status)

		// 验证事件已记录
		var events []model.SagaEvent
		err = db.Where("saga_execution_id = ?", saga.ID).Find(&events).Error
		require.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, EventSagaStarted, events[0].EventType)
	})

	t.Run("Should reject invalid saga status transitions", func(t *testing.T) {
		saga := createTestSaga(t, db)

		// 测试非法状态转换：pending -> completed (跳过running)
		err := stateManager.TransitionSagaStatus(ctx, saga.ID, model.SagaStatusPending, model.SagaStatusCompleted)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transition")

		// 验证状态未改变
		var unchangedSaga model.SagaExecution
		err = db.First(&unchangedSaga, saga.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.SagaStatusPending, unchangedSaga.Status)
	})

	t.Run("Should handle concurrent status transitions", func(t *testing.T) {
		saga := createTestSaga(t, db)

		// 模拟并发状态转换
		done := make(chan error, 2)

		go func() {
			err := stateManager.TransitionSagaStatus(ctx, saga.ID, model.SagaStatusPending, model.SagaStatusRunning)
			done <- err
		}()

		go func() {
			time.Sleep(10 * time.Millisecond) // 稍微延迟确保第一个先执行
			err := stateManager.TransitionSagaStatus(ctx, saga.ID, model.SagaStatusPending, model.SagaStatusRunning)
			done <- err
		}()

		// 收集结果
		results := make([]error, 2)
		for i := 0; i < 2; i++ {
			results[i] = <-done
		}

		// 应该有一个成功，一个因为状态冲突失败
		successCount := 0
		conflictCount := 0

		for _, err := range results {
			if err == nil {
				successCount++
			} else if sagaErr, ok := err.(*SagaError); ok && sagaErr.Code == ErrCodeStateConflict {
				conflictCount++
			}
		}

		assert.Equal(t, 1, successCount, "Exactly one transition should succeed")
		assert.Equal(t, 1, conflictCount, "Exactly one transition should fail with conflict")
	})

	t.Run("Should transition step status successfully", func(t *testing.T) {
		saga := createTestSaga(t, db)
		step := createTestStep(t, db, saga.ID, 0)

		// 转换步骤状态：pending -> running
		err := stateManager.TransitionStepStatus(ctx, step.ID, model.StepStatusPending, model.StepStatusRunning)
		assert.NoError(t, err)

		// 验证状态已更新并设置了开始时间
		var updatedStep model.SagaStepExecution
		err = db.First(&updatedStep, step.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StepStatusRunning, updatedStep.Status)
		assert.NotNil(t, updatedStep.StartedAt)

		// 转换到完成状态
		err = stateManager.TransitionStepStatus(ctx, step.ID, model.StepStatusRunning, model.StepStatusCompleted)
		assert.NoError(t, err)

		// 验证完成时间已设置
		err = db.First(&updatedStep, step.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StepStatusCompleted, updatedStep.Status)
		assert.NotNil(t, updatedStep.CompletedAt)
	})

	t.Run("Should update saga progress", func(t *testing.T) {
		saga := createTestSaga(t, db)
		newContext := `{"progress": 50, "currentFile": "test.jpg"}`

		err := stateManager.UpdateSagaProgress(ctx, saga.ID, 1, newContext)
		assert.NoError(t, err)

		// 验证进度已更新
		var updatedSaga model.SagaExecution
		err = db.First(&updatedSaga, saga.ID).Error
		require.NoError(t, err)
		assert.Equal(t, 1, updatedSaga.CurrentStep)
		assert.Equal(t, newContext, updatedSaga.Context)
	})

	t.Run("Should record step result", func(t *testing.T) {
		saga := createTestSaga(t, db)
		step := createTestStep(t, db, saga.ID, 0)

		output := `{"fileId": 123, "objectKey": "users/1/file.jpg"}`
		errorMsg := "network timeout"

		err := stateManager.RecordStepResult(ctx, step.ID, output, errorMsg)
		assert.NoError(t, err)

		// 验证结果已记录
		var updatedStep model.SagaStepExecution
		err = db.First(&updatedStep, step.ID).Error
		require.NoError(t, err)
		assert.Equal(t, output, updatedStep.Output)
		assert.Equal(t, errorMsg, updatedStep.Error)
	})

	t.Run("Should get saga execution with steps", func(t *testing.T) {
		saga := createTestSaga(t, db)
		step1 := createTestStep(t, db, saga.ID, 0)
		step2 := createTestStep(t, db, saga.ID, 1)

		execution, err := stateManager.GetSagaExecution(ctx, saga.ID)
		assert.NoError(t, err)
		assert.Equal(t, saga.ID, execution.ID)
		assert.Len(t, execution.Steps, 2)

		// 验证步骤顺序
		assert.Equal(t, step1.ID, execution.Steps[0].ID)
		assert.Equal(t, step2.ID, execution.Steps[1].ID)
	})

	t.Run("Should get saga steps in correct order", func(t *testing.T) {
		saga := createTestSaga(t, db)
		// 故意创建乱序的步骤
		createTestStep(t, db, saga.ID, 2)
		createTestStep(t, db, saga.ID, 0)
		createTestStep(t, db, saga.ID, 1)

		steps, err := stateManager.GetSagaSteps(ctx, saga.ID)
		assert.NoError(t, err)
		assert.Len(t, steps, 3)

		// 验证步骤按索引排序
		assert.Equal(t, 0, steps[0].StepIndex)
		assert.Equal(t, 1, steps[1].StepIndex)
		assert.Equal(t, 2, steps[2].StepIndex)
	})

	t.Run("Should get recoverable sagas", func(t *testing.T) {
		// 创建不同状态的Saga
		runningSaga := createTestSaga(t, db)
		runningSaga.Status = model.SagaStatusRunning
		db.Save(runningSaga)

		compensatingSaga := createTestSaga(t, db)
		compensatingSaga.Status = model.SagaStatusCompensating
		db.Save(compensatingSaga)

		completedSaga := createTestSaga(t, db)
		completedSaga.Status = model.SagaStatusCompleted
		db.Save(completedSaga)

		// 创建一个长时间pending的Saga（应该被恢复）
		stalePendingSaga := createTestSaga(t, db)
		stalePendingSaga.Status = model.SagaStatusPending
		stalePendingSaga.UpdatedAt = time.Now().Add(-10 * time.Minute)
		db.Save(stalePendingSaga)

		sagas, err := stateManager.GetRecoverableSagas(ctx)
		assert.NoError(t, err)

		// 应该返回running、compensating和stale pending的Saga
		assert.Len(t, sagas, 3)

		sagaIDs := make([]uint, len(sagas))
		for i, saga := range sagas {
			sagaIDs[i] = saga.ID
		}

		assert.Contains(t, sagaIDs, runningSaga.ID)
		assert.Contains(t, sagaIDs, compensatingSaga.ID)
		assert.Contains(t, sagaIDs, stalePendingSaga.ID)
		assert.NotContains(t, sagaIDs, completedSaga.ID)
	})

	t.Run("Should handle non-existent saga", func(t *testing.T) {
		_, err := stateManager.GetSagaExecution(ctx, 9999)
		assert.Error(t, err)

		sagaErr, ok := err.(*SagaError)
		assert.True(t, ok)
		assert.Equal(t, ErrCodeSagaNotFound, sagaErr.Code)
	})
}

func TestLockManager(t *testing.T) {
	lockMgr := NewLockManager()
	ctx := context.Background()

	t.Run("Should acquire and release lock successfully", func(t *testing.T) {
		lock, err := lockMgr.AcquireLock(ctx, "test-key", time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, lock)
		assert.True(t, lock.IsLocked())

		err = lock.Unlock()
		assert.NoError(t, err)
		assert.False(t, lock.IsLocked())
	})

	t.Run("Should prevent acquiring same lock twice", func(t *testing.T) {
		lock1, err := lockMgr.AcquireLock(ctx, "test-key-2", time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, lock1)

		// 尝试获取同一把锁应该失败
		lock2, err := lockMgr.AcquireLock(ctx, "test-key-2", time.Second)
		assert.Error(t, err)
		assert.Nil(t, lock2)
		assert.Contains(t, err.Error(), "already held")

		// 释放锁后应该可以重新获取
		err = lock1.Unlock()
		assert.NoError(t, err)

		lock3, err := lockMgr.AcquireLock(ctx, "test-key-2", time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, lock3)

		lock3.Unlock()
	})

	t.Run("Should extend lock timeout", func(t *testing.T) {
		lock, err := lockMgr.AcquireLock(ctx, "test-key-3", 100*time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, lock.IsLocked())

		// 延长锁时间
		err = lock.Extend(time.Second)
		assert.NoError(t, err)

		// 等待原始超时时间，锁应该仍然有效
		time.Sleep(200 * time.Millisecond)
		assert.True(t, lock.IsLocked())

		lock.Unlock()
	})

	t.Run("Should handle lock expiration", func(t *testing.T) {
		lock, err := lockMgr.AcquireLock(ctx, "test-key-4", 50*time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, lock.IsLocked())

		// 等待锁过期
		time.Sleep(100 * time.Millisecond)
		assert.False(t, lock.IsLocked())
	})
}

func TestEventBus(t *testing.T) {
	eventBus := NewEventBus()
	defer eventBus.Close()

	t.Run("Should subscribe and receive events", func(t *testing.T) {
		received := make(chan interface{}, 1)

		handler := func(event interface{}) {
			received <- event
		}

		eventBus.Subscribe("SagaStatusChanged", handler)

		testEvent := SagaStatusChangedEvent{
			SagaID: 123,
			From:   model.SagaStatusPending,
			To:     model.SagaStatusRunning,
		}

		eventBus.Publish(testEvent)

		select {
		case receivedEvent := <-received:
			assert.Equal(t, testEvent, receivedEvent)
		case <-time.After(time.Second):
			t.Fatal("Event was not received")
		}
	})

	t.Run("Should handle async events", func(t *testing.T) {
		received := make(chan interface{}, 1)

		handler := func(event interface{}) {
			received <- event
		}

		eventBus.Subscribe("SagaStatusChanged", handler)

		testEvent := SagaStatusChangedEvent{
			SagaID: 456,
			From:   model.SagaStatusRunning,
			To:     model.SagaStatusCompleted,
		}

		eventBus.PublishAsync(testEvent)

		select {
		case receivedEvent := <-received:
			assert.Equal(t, testEvent, receivedEvent)
		case <-time.After(2 * time.Second):
			t.Fatal("Async event was not received")
		}
	})
}

func TestRetryManager(t *testing.T) {
	retryMgr := NewRetryManager()

	t.Run("Should determine if error is retryable", func(t *testing.T) {
		policy := &model.RetryPolicy{
			MaxAttempts:     3,
			RetryableErrors: []string{"timeout", "network"},
		}

		// 可重试错误
		timeoutErr := NewSagaError(ErrCodeStepTimeout, "timeout", 123, 1)
		assert.True(t, retryMgr.ShouldRetry(timeoutErr, 1, policy))
		assert.True(t, retryMgr.ShouldRetry(timeoutErr, 2, policy))
		assert.False(t, retryMgr.ShouldRetry(timeoutErr, 3, policy)) // 达到最大重试次数

		// 不可重试错误
		validationErr := NewSagaError(ErrCodeInvalidContext, "validation failed", 123, 1)
		assert.False(t, retryMgr.ShouldRetry(validationErr, 1, policy))
	})

	t.Run("Should calculate retry delay with exponential backoff", func(t *testing.T) {
		policy := &model.RetryPolicy{
			InitialDelay:  time.Second,
			MaxDelay:      10 * time.Second,
			BackoffFactor: 2.0,
		}

		delay0 := retryMgr.CalculateDelay(0, policy)
		assert.Equal(t, time.Second, delay0)

		delay1 := retryMgr.CalculateDelay(1, policy)
		assert.Equal(t, 2*time.Second, delay1)

		delay2 := retryMgr.CalculateDelay(2, policy)
		assert.Equal(t, 4*time.Second, delay2)

		// 测试最大延迟限制
		delay10 := retryMgr.CalculateDelay(10, policy)
		assert.Equal(t, 10*time.Second, delay10)
	})
}

func TestTimeoutManager(t *testing.T) {
	timeoutMgr := NewTimeoutManager()
	ctx := context.Background()

	t.Run("Should create timeout context", func(t *testing.T) {
		timeoutCtx, cancel := timeoutMgr.CreateTimeoutContext(ctx, 100*time.Millisecond)
		defer cancel()

		// 等待超时
		time.Sleep(200 * time.Millisecond)

		// 上下文应该已超时
		assert.Error(t, timeoutCtx.Err())
		assert.True(t, timeoutMgr.IsTimeoutError(timeoutCtx.Err()))
	})

	t.Run("Should identify timeout errors", func(t *testing.T) {
		timeoutErr := NewSagaError(ErrCodeSagaTimeout, "saga timeout", 123, -1)
		assert.True(t, timeoutMgr.IsTimeoutError(timeoutErr))

		stepTimeoutErr := NewSagaError(ErrCodeStepTimeout, "step timeout", 123, 1)
		assert.True(t, timeoutMgr.IsTimeoutError(stepTimeoutErr))

		regularErr := NewSagaError(ErrCodeInvalidContext, "invalid context", 123, -1)
		assert.False(t, timeoutMgr.IsTimeoutError(regularErr))
	})
}
