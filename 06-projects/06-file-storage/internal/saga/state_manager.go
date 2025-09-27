package saga

import (
	"context"
	"file-storage-service/internal/model"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// stateManager Saga状态管理器实现
type stateManager struct {
	db       *gorm.DB
	lockMgr  *lockManager
	eventBus EventBus
	mutex    sync.RWMutex
}

// NewSagaStateManager 创建状态管理器
func NewSagaStateManager(db *gorm.DB) SagaStateManager {
	return &stateManager{
		db:       db,
		lockMgr:  NewLockManager(),
		eventBus: NewEventBus(),
	}
}

// TransitionSagaStatus 转换Saga状态
func (s *stateManager) TransitionSagaStatus(ctx context.Context, sagaID uint, from, to model.SagaStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证状态转换是否合法
	if err := s.validateSagaStatusTransition(from, to); err != nil {
		return NewSagaErrorWithCause(ErrCodeInvalidStateTransition,
			fmt.Sprintf("invalid transition from %s to %s", from, to),
			sagaID, -1, err)
	}

	// 在事务中执行状态转换
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用乐观锁检查当前状态
		var execution model.SagaExecution
		if err := tx.Where("id = ? AND status = ?", sagaID, from).First(&execution).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NewSagaError(ErrCodeStateConflict,
					fmt.Sprintf("saga %d not in expected state %s", sagaID, from),
					sagaID, -1)
			}
			return err
		}

		// 更新状态和时间戳
		updates := map[string]interface{}{
			"status":     to,
			"updated_at": time.Now(),
		}

		// 如果转换到终态，设置完成时间
		if to.IsTerminal() {
			now := time.Now()
			updates["completed_at"] = &now
		}

		if err := tx.Model(&execution).Updates(updates).Error; err != nil {
			return err
		}

		// 记录状态转换事件
		event := &model.SagaEvent{
			SagaExecutionID: sagaID,
			EventType:       s.getSagaEventType(to),
			EventData:       fmt.Sprintf(`{"from": "%s", "to": "%s"}`, from, to),
			TraceID:         s.getTraceID(ctx),
			CreatedAt:       time.Now(),
		}

		return tx.Create(event).Error
	})

	if err != nil {
		return NewSagaErrorWithCause(ErrCodeStateConflict,
			"failed to transition saga status", sagaID, -1, err)
	}

	// 发布状态转换事件
	s.eventBus.Publish(SagaStatusChangedEvent{
		SagaID: sagaID,
		From:   from,
		To:     to,
	})

	return nil
}

// TransitionStepStatus 转换步骤状态
func (s *stateManager) TransitionStepStatus(ctx context.Context, stepID uint, from, to model.StepStatus) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证状态转换是否合法
	if err := s.validateStepStatusTransition(from, to); err != nil {
		return NewSagaErrorWithCause(ErrCodeInvalidStateTransition,
			fmt.Sprintf("invalid step transition from %s to %s", from, to),
			0, int(stepID), err)
	}

	// 在事务中执行状态转换
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 使用乐观锁检查当前状态
		var stepExecution model.SagaStepExecution
		if err := tx.Where("id = ? AND status = ?", stepID, from).First(&stepExecution).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return NewSagaError(ErrCodeStateConflict,
					fmt.Sprintf("step %d not in expected state %s", stepID, from),
					stepExecution.SagaExecutionID, int(stepID))
			}
			return err
		}

		// 更新状态和时间戳
		updates := map[string]interface{}{
			"status":     to,
			"updated_at": time.Now(),
		}

		// 根据状态设置相应的时间戳
		now := time.Now()
		switch to {
		case model.StepStatusRunning:
			updates["started_at"] = &now
		case model.StepStatusCompleted, model.StepStatusFailed, model.StepStatusSkipped:
			updates["completed_at"] = &now
		case model.StepStatusCompensated:
			updates["compensated_at"] = &now
		}

		if err := tx.Model(&stepExecution).Updates(updates).Error; err != nil {
			return err
		}

		// 记录步骤状态转换事件
		event := &model.SagaEvent{
			SagaExecutionID: stepExecution.SagaExecutionID,
			EventType:       s.getStepEventType(to),
			StepName:        stepExecution.StepName,
			StepIndex:       stepExecution.StepIndex,
			EventData:       fmt.Sprintf(`{"from": "%s", "to": "%s"}`, from, to),
			TraceID:         s.getTraceID(ctx),
			CreatedAt:       time.Now(),
		}

		return tx.Create(event).Error
	})

	if err != nil {
		return NewSagaErrorWithCause(ErrCodeStateConflict,
			"failed to transition step status", 0, int(stepID), err)
	}

	return nil
}

// UpdateSagaProgress 更新Saga进度
func (s *stateManager) UpdateSagaProgress(ctx context.Context, sagaID uint, currentStep int, context string) error {
	updates := map[string]interface{}{
		"current_step": currentStep,
		"context":      context,
		"updated_at":   time.Now(),
	}

	if err := s.db.WithContext(ctx).Model(&model.SagaExecution{}).
		Where("id = ?", sagaID).Updates(updates).Error; err != nil {
		return NewSagaErrorWithCause(ErrCodeStateConflict,
			"failed to update saga progress", sagaID, -1, err)
	}

	return nil
}

// RecordStepResult 记录步骤执行结果
func (s *stateManager) RecordStepResult(ctx context.Context, stepID uint, output string, error string) error {
	updates := map[string]interface{}{
		"output":     output,
		"error":      error,
		"updated_at": time.Now(),
	}

	if err := s.db.WithContext(ctx).Model(&model.SagaStepExecution{}).
		Where("id = ?", stepID).Updates(updates).Error; err != nil {
		return NewSagaErrorWithCause(ErrCodeStateConflict,
			"failed to record step result", 0, int(stepID), err)
	}

	return nil
}

// AcquireSagaLock 获取Saga锁
func (s *stateManager) AcquireSagaLock(ctx context.Context, sagaID uint, timeout time.Duration) (SagaLock, error) {
	lockKey := fmt.Sprintf("saga:%d", sagaID)
	return s.lockMgr.AcquireLock(ctx, lockKey, timeout)
}

// RecordSagaEvent 记录Saga事件
func (s *stateManager) RecordSagaEvent(ctx context.Context, event *model.SagaEvent) error {
	if err := s.db.WithContext(ctx).Create(event).Error; err != nil {
		return NewSagaErrorWithCause(ErrCodeSerializationFailed,
			"failed to record saga event", event.SagaExecutionID, -1, err)
	}

	return nil
}

// GetSagaExecution 获取Saga执行记录
func (s *stateManager) GetSagaExecution(ctx context.Context, sagaID uint) (*model.SagaExecution, error) {
	var execution model.SagaExecution
	if err := s.db.WithContext(ctx).Preload("Steps").First(&execution, sagaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, NewSagaError(ErrCodeSagaNotFound,
				"saga execution not found", sagaID, -1)
		}
		return nil, NewSagaErrorWithCause(ErrCodeSerializationFailed,
			"failed to get saga execution", sagaID, -1, err)
	}

	return &execution, nil
}

// GetSagaSteps 获取Saga的所有步骤
func (s *stateManager) GetSagaSteps(ctx context.Context, sagaID uint) ([]*model.SagaStepExecution, error) {
	var steps []*model.SagaStepExecution
	if err := s.db.WithContext(ctx).Where("saga_execution_id = ?", sagaID).
		Order("step_index ASC").Find(&steps).Error; err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeSerializationFailed,
			"failed to get saga steps", sagaID, -1, err)
	}

	return steps, nil
}

// GetRecoverableSagas 获取可恢复的Saga
func (s *stateManager) GetRecoverableSagas(ctx context.Context) ([]*model.SagaExecution, error) {
	var sagas []*model.SagaExecution

	// 查找处于非终态的Saga
	if err := s.db.WithContext(ctx).
		Where("status IN ?", []model.SagaStatus{
			model.SagaStatusRunning,
			model.SagaStatusCompensating,
		}).
		Or("status = ? AND updated_at < ?", model.SagaStatusPending, time.Now().Add(-5*time.Minute)).
		Find(&sagas).Error; err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeSerializationFailed,
			"failed to get recoverable sagas", 0, -1, err)
	}

	return sagas, nil
}

// validateSagaStatusTransition 验证Saga状态转换
func (s *stateManager) validateSagaStatusTransition(from, to model.SagaStatus) error {
	validTransitions := map[model.SagaStatus][]model.SagaStatus{
		model.SagaStatusPending: {
			model.SagaStatusRunning,
			model.SagaStatusCancelled,
		},
		model.SagaStatusRunning: {
			model.SagaStatusCompleted,
			model.SagaStatusCompensating,
			model.SagaStatusFailed,
			model.SagaStatusTimeout,
			model.SagaStatusCancelled,
		},
		model.SagaStatusCompensating: {
			model.SagaStatusCompensated,
			model.SagaStatusFailed,
		},
	}

	allowedStates, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("no valid transitions from state %s", from)
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return nil
		}
	}

	return fmt.Errorf("transition from %s to %s is not allowed", from, to)
}

// validateStepStatusTransition 验证步骤状态转换
func (s *stateManager) validateStepStatusTransition(from, to model.StepStatus) error {
	validTransitions := map[model.StepStatus][]model.StepStatus{
		model.StepStatusPending: {
			model.StepStatusRunning,
			model.StepStatusSkipped,
		},
		model.StepStatusRunning: {
			model.StepStatusCompleted,
			model.StepStatusFailed,
		},
		model.StepStatusFailed: {
			model.StepStatusRetrying,
			model.StepStatusCompensating,
		},
		model.StepStatusRetrying: {
			model.StepStatusRunning,
			model.StepStatusCompensating,
		},
		model.StepStatusCompensating: {
			model.StepStatusCompensated,
		},
		model.StepStatusCompleted: {
			model.StepStatusCompensating, // 需要补偿时
		},
	}

	allowedStates, exists := validTransitions[from]
	if !exists {
		return fmt.Errorf("no valid transitions from state %s", from)
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return nil
		}
	}

	return fmt.Errorf("transition from %s to %s is not allowed", from, to)
}

// getSagaEventType 根据Saga状态获取事件类型
func (s *stateManager) getSagaEventType(status model.SagaStatus) string {
	switch status {
	case model.SagaStatusRunning:
		return EventSagaStarted
	case model.SagaStatusCompleted:
		return EventSagaCompleted
	case model.SagaStatusFailed:
		return EventSagaFailed
	case model.SagaStatusCancelled:
		return EventSagaCancelled
	case model.SagaStatusTimeout:
		return EventSagaTimeout
	default:
		return "saga_status_changed"
	}
}

// getStepEventType 根据步骤状态获取事件类型
func (s *stateManager) getStepEventType(status model.StepStatus) string {
	switch status {
	case model.StepStatusRunning:
		return EventStepStarted
	case model.StepStatusCompleted:
		return EventStepCompleted
	case model.StepStatusFailed:
		return EventStepFailed
	case model.StepStatusRetrying:
		return EventStepRetrying
	case model.StepStatusCompensating:
		return EventStepCompensating
	case model.StepStatusCompensated:
		return EventStepCompensated
	default:
		return "step_status_changed"
	}
}

// getTraceID 从上下文中获取跟踪ID
func (s *stateManager) getTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value("traceID").(string); ok {
		return traceID
	}
	return ""
}

// SagaStatusChangedEvent Saga状态变化事件
type SagaStatusChangedEvent struct {
	SagaID uint
	From   model.SagaStatus
	To     model.SagaStatus
}
