package saga

import (
	"context"
	"encoding/json"
	"file-storage-service/internal/model"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

// executionEngine Saga执行引擎实现
type executionEngine struct {
	stateManager   SagaStateManager
	retryManager   *RetryManager
	timeoutManager *TimeoutManager
	repository     SagaRepository
	eventBus       EventBus
	workerPool     *WorkerPool
	mutex          sync.RWMutex
}

// NewSagaExecutionEngine 创建Saga执行引擎
func NewSagaExecutionEngine(db *gorm.DB) SagaExecutionEngine {
	stateManager := NewSagaStateManager(db)
	repository := NewSagaRepository(db)
	eventBus := NewEventBus()

	engine := &executionEngine{
		stateManager:   stateManager,
		retryManager:   NewRetryManager(),
		timeoutManager: NewTimeoutManager(),
		repository:     repository,
		eventBus:       eventBus,
		workerPool:     NewWorkerPool(10), // 10个并发工作器
	}

	return engine
}

// ExecuteSaga 执行Saga
func (e *executionEngine) ExecuteSaga(ctx context.Context, execution *model.SagaExecution, definition SagaDefinition) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 获取Saga锁，防止重复执行
	lock, err := e.stateManager.AcquireSagaLock(ctx, execution.ID, DefaultLockTimeout)
	if err != nil {
		return NewSagaErrorWithCause(ErrCodeSagaLockFailed, "failed to acquire saga lock", execution.ID, -1, err)
	}
	defer lock.Unlock()

	// 验证Saga状态
	if execution.Status != model.SagaStatusPending && execution.Status != model.SagaStatusRunning {
		return NewSagaError(ErrCodeInvalidStateTransition,
			fmt.Sprintf("cannot execute saga in status %s", execution.Status), execution.ID, -1)
	}

	// 转换到运行状态
	if execution.Status == model.SagaStatusPending {
		if err := e.stateManager.TransitionSagaStatus(ctx, execution.ID,
			model.SagaStatusPending, model.SagaStatusRunning); err != nil {
			return err
		}
		execution.Status = model.SagaStatusRunning
	}

	// 创建超时上下文
	sagaTimeout := definition.CalculateTimeout(execution.Context)
	timeoutCtx, cancel := e.timeoutManager.CreateTimeoutContext(ctx, sagaTimeout)
	defer cancel()

	// 执行所有步骤
	steps := definition.GetSteps()
	for i := execution.CurrentStep; i < len(steps); i++ {
		select {
		case <-timeoutCtx.Done():
			// Saga超时
			if err := e.handleSagaTimeout(ctx, execution); err != nil {
				return err
			}
			return NewSagaError(ErrCodeSagaTimeout, "saga execution timeout", execution.ID, -1)
		default:
			// 执行当前步骤
			stepDef := steps[i]
			if err := e.executeStep(timeoutCtx, execution, stepDef, i, definition); err != nil {
				// 步骤执行失败，启动补偿
				return e.handleStepFailure(ctx, execution, definition, i, err)
			}

			// 更新进度
			execution.CurrentStep = i + 1
			if err := e.stateManager.UpdateSagaProgress(ctx, execution.ID, execution.CurrentStep, execution.Context); err != nil {
				return err
			}
		}
	}

	// 所有步骤执行成功，标记Saga完成
	return e.completeSaga(ctx, execution)
}

// executeStep 执行单个步骤
func (e *executionEngine) executeStep(ctx context.Context, execution *model.SagaExecution,
	stepDef SagaStepDefinition, stepIndex int, definition SagaDefinition) error {

	// 获取或创建步骤执行记录
	stepExecution, err := e.getOrCreateStepExecution(ctx, execution.ID, stepDef, stepIndex)
	if err != nil {
		return err
	}

	// 检查步骤是否已经完成
	if stepExecution.Status == model.StepStatusCompleted {
		return nil // 幂等性：步骤已完成，直接返回
	}

	// 检查前置条件
	if err := e.checkStepPrerequisites(ctx, execution.ID, stepDef.Prerequisites, stepIndex); err != nil {
		return NewSagaErrorWithCause(ErrCodeStepPrerequisiteFailed,
			"step prerequisites not satisfied", execution.ID, stepIndex, err)
	}

	// 反序列化Saga上下文
	var sagaContext interface{}
	if execution.Context != "" {
		if err := json.Unmarshal([]byte(execution.Context), &sagaContext); err != nil {
			return NewSagaErrorWithCause(ErrCodeDeserializationFailed,
				"failed to deserialize saga context", execution.ID, stepIndex, err)
		}
	}

	// 准备步骤输入
	var stepInput interface{}
	if stepExecution.Input != "" {
		if err := json.Unmarshal([]byte(stepExecution.Input), &stepInput); err != nil {
			return NewSagaErrorWithCause(ErrCodeDeserializationFailed,
				"failed to deserialize step input", execution.ID, stepIndex, err)
		}
	}

	// 创建步骤超时上下文
	stepTimeout := stepDef.Timeout
	if stepTimeout <= 0 {
		stepTimeout = DefaultStepTimeout
	}
	stepCtx, cancel := e.timeoutManager.CreateTimeoutContext(ctx, stepTimeout)
	defer cancel()

	// 转换步骤状态为运行中
	if err := e.stateManager.TransitionStepStatus(ctx, stepExecution.ID,
		stepExecution.Status, model.StepStatusRunning); err != nil {
		return err
	}

	// 执行步骤带重试
	return e.executeStepWithRetry(stepCtx, stepExecution, stepDef, sagaContext, stepInput)
}

// ExecuteStep 执行单个步骤（公开接口）
func (e *executionEngine) ExecuteStep(ctx context.Context, stepExecution *model.SagaStepExecution,
	stepDef SagaStepDefinition, sagaCtx interface{}) error {

	// 反序列化步骤输入
	var stepInput interface{}
	if stepExecution.Input != "" {
		if err := json.Unmarshal([]byte(stepExecution.Input), &stepInput); err != nil {
			return NewSagaErrorWithCause(ErrCodeDeserializationFailed,
				"failed to deserialize step input", stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
		}
	}

	// 执行步骤
	output, err := stepDef.Execute(ctx, sagaCtx, stepInput)
	if err != nil {
		// 记录错误
		e.stateManager.RecordStepResult(ctx, stepExecution.ID, "", err.Error())
		e.stateManager.TransitionStepStatus(ctx, stepExecution.ID, model.StepStatusRunning, model.StepStatusFailed)
		return err
	}

	// 序列化输出
	var outputStr string
	if output != nil {
		outputBytes, err := json.Marshal(output)
		if err != nil {
			return NewSagaErrorWithCause(ErrCodeSerializationFailed,
				"failed to serialize step output", stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
		}
		outputStr = string(outputBytes)
	}

	// 记录成功结果
	if err := e.stateManager.RecordStepResult(ctx, stepExecution.ID, outputStr, ""); err != nil {
		return err
	}

	// 转换步骤状态为完成
	return e.stateManager.TransitionStepStatus(ctx, stepExecution.ID, model.StepStatusRunning, model.StepStatusCompleted)
}

// executeStepWithRetry 带重试的步骤执行
func (e *executionEngine) executeStepWithRetry(ctx context.Context, stepExecution *model.SagaStepExecution,
	stepDef SagaStepDefinition, sagaCtx interface{}, stepInput interface{}) error {

	var lastErr error
	retryPolicy := stepDef.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = model.DefaultRetryPolicy()
	}

	for attempt := stepExecution.AttemptCount; attempt < retryPolicy.MaxAttempts; attempt++ {
		// 更新尝试次数
		stepExecution.AttemptCount = attempt + 1
		if err := e.repository.UpdateSagaStepExecution(ctx, stepExecution); err != nil {
			return err
		}

		// 执行步骤
		output, err := stepDef.Execute(ctx, sagaCtx, stepInput)
		if err != nil {
			lastErr = err

			// 检查是否可以重试
			if !e.retryManager.ShouldRetry(err, stepExecution.AttemptCount, retryPolicy) {
				// 不可重试，直接失败
				e.stateManager.RecordStepResult(ctx, stepExecution.ID, "", err.Error())
				e.stateManager.TransitionStepStatus(ctx, stepExecution.ID, model.StepStatusRunning, model.StepStatusFailed)
				return NewSagaErrorWithCause(ErrCodeStepExecutionFailed,
					fmt.Sprintf("step failed after %d attempts", stepExecution.AttemptCount),
					stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
			}

			// 计算重试延迟
			delay := e.retryManager.CalculateDelay(stepExecution.AttemptCount-1, retryPolicy)

			// 设置下次重试时间
			nextRetryAt := time.Now().Add(delay)
			stepExecution.NextRetryAt = &nextRetryAt

			// 转换到重试状态
			e.stateManager.TransitionStepStatus(ctx, stepExecution.ID, model.StepStatusRunning, model.StepStatusRetrying)

			// 记录重试事件
			e.recordRetryEvent(ctx, stepExecution, attempt+1, err)

			// 等待重试延迟
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// 继续重试
				e.stateManager.TransitionStepStatus(ctx, stepExecution.ID, model.StepStatusRetrying, model.StepStatusRunning)
				continue
			}
		}

		// 步骤执行成功
		var outputStr string
		if output != nil {
			outputBytes, err := json.Marshal(output)
			if err != nil {
				return NewSagaErrorWithCause(ErrCodeSerializationFailed,
					"failed to serialize step output", stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
			}
			outputStr = string(outputBytes)
		}

		// 记录成功结果
		if err := e.stateManager.RecordStepResult(ctx, stepExecution.ID, outputStr, ""); err != nil {
			return err
		}

		// 转换步骤状态为完成
		if err := e.stateManager.TransitionStepStatus(ctx, stepExecution.ID,
			model.StepStatusRunning, model.StepStatusCompleted); err != nil {
			return err
		}

		return nil
	}

	// 所有重试都失败了
	e.stateManager.RecordStepResult(ctx, stepExecution.ID, "", lastErr.Error())
	e.stateManager.TransitionStepStatus(ctx, stepExecution.ID, model.StepStatusRunning, model.StepStatusFailed)

	return NewSagaErrorWithCause(ErrCodeMaxRetriesExceeded,
		fmt.Sprintf("step failed after %d attempts", retryPolicy.MaxAttempts),
		stepExecution.SagaExecutionID, stepExecution.StepIndex, lastErr)
}

// CompensateStep 补偿单个步骤
func (e *executionEngine) CompensateStep(ctx context.Context, stepExecution *model.SagaStepExecution,
	stepDef SagaStepDefinition, sagaCtx interface{}) error {

	// 只有成功完成的步骤才能被补偿
	if stepExecution.Status != model.StepStatusCompleted {
		return nil // 已经失败的步骤不需要补偿
	}

	// 检查是否有补偿函数
	if stepDef.Compensate == nil {
		if stepDef.IsCritical {
			return NewSagaError(ErrCodeCompensationImpossible,
				"critical step has no compensation function",
				stepExecution.SagaExecutionID, stepExecution.StepIndex)
		}
		return nil // 非关键步骤可以跳过补偿
	}

	// 转换步骤状态为补偿中
	if err := e.stateManager.TransitionStepStatus(ctx, stepExecution.ID,
		model.StepStatusCompleted, model.StepStatusCompensating); err != nil {
		return err
	}

	// 反序列化原始输入用于补偿
	var originalInput interface{}
	if stepExecution.Input != "" {
		if err := json.Unmarshal([]byte(stepExecution.Input), &originalInput); err != nil {
			return NewSagaErrorWithCause(ErrCodeDeserializationFailed,
				"failed to deserialize original input for compensation",
				stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
		}
	}

	// 执行补偿
	if err := stepDef.Compensate(ctx, sagaCtx, originalInput); err != nil {
		// 补偿失败
		e.stateManager.RecordStepResult(ctx, stepExecution.ID, "",
			fmt.Sprintf("compensation failed: %s", err.Error()))
		return NewSagaErrorWithCause(ErrCodeStepCompensationFailed,
			"step compensation failed", stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
	}

	// 转换步骤状态为已补偿
	return e.stateManager.TransitionStepStatus(ctx, stepExecution.ID,
		model.StepStatusCompensating, model.StepStatusCompensated)
}

// CompensateSaga 补偿整个Saga
func (e *executionEngine) CompensateSaga(ctx context.Context, execution *model.SagaExecution, definition SagaDefinition) error {
	// 转换Saga状态为补偿中
	if err := e.stateManager.TransitionSagaStatus(ctx, execution.ID,
		execution.Status, model.SagaStatusCompensating); err != nil {
		return err
	}

	// 获取所有步骤
	steps, err := e.stateManager.GetSagaSteps(ctx, execution.ID)
	if err != nil {
		return err
	}

	// 反序列化Saga上下文
	var sagaContext interface{}
	if execution.Context != "" {
		if err := json.Unmarshal([]byte(execution.Context), &sagaContext); err != nil {
			return NewSagaErrorWithCause(ErrCodeDeserializationFailed,
				"failed to deserialize saga context for compensation", execution.ID, -1, err)
		}
	}

	// 按相反顺序补偿已完成的步骤
	stepDefs := definition.GetSteps()
	compensationFailed := false

	for i := len(steps) - 1; i >= 0; i-- {
		step := steps[i]
		stepDef := stepDefs[step.StepIndex]

		if err := e.CompensateStep(ctx, step, stepDef, sagaContext); err != nil {
			// 记录补偿错误但继续尝试补偿其他步骤
			execution.CompensationError = err.Error()
			e.repository.UpdateSagaExecution(ctx, execution)
			compensationFailed = true
		}
	}

	// 根据补偿结果设置最终状态
	if compensationFailed {
		return e.stateManager.TransitionSagaStatus(ctx, execution.ID,
			model.SagaStatusCompensating, model.SagaStatusFailed)
	}

	return e.stateManager.TransitionSagaStatus(ctx, execution.ID,
		model.SagaStatusCompensating, model.SagaStatusCompensated)
}

// RetryStep 重试步骤
func (e *executionEngine) RetryStep(ctx context.Context, stepExecution *model.SagaStepExecution,
	stepDef SagaStepDefinition, sagaCtx interface{}) error {

	// 只有失败的步骤才能重试
	if stepExecution.Status != model.StepStatusFailed && stepExecution.Status != model.StepStatusRetrying {
		return NewSagaError(ErrCodeInvalidStateTransition,
			fmt.Sprintf("cannot retry step in status %s", stepExecution.Status),
			stepExecution.SagaExecutionID, stepExecution.StepIndex)
	}

	// 检查是否还能重试
	retryPolicy := stepDef.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = model.DefaultRetryPolicy()
	}

	if stepExecution.AttemptCount >= retryPolicy.MaxAttempts {
		return NewSagaError(ErrCodeMaxRetriesExceeded,
			"maximum retry attempts exceeded",
			stepExecution.SagaExecutionID, stepExecution.StepIndex)
	}

	// 转换状态为运行中
	if err := e.stateManager.TransitionStepStatus(ctx, stepExecution.ID,
		stepExecution.Status, model.StepStatusRunning); err != nil {
		return err
	}

	// 准备步骤输入
	var stepInput interface{}
	if stepExecution.Input != "" {
		if err := json.Unmarshal([]byte(stepExecution.Input), &stepInput); err != nil {
			return NewSagaErrorWithCause(ErrCodeDeserializationFailed,
				"failed to deserialize step input for retry",
				stepExecution.SagaExecutionID, stepExecution.StepIndex, err)
		}
	}

	// 执行步骤
	return e.executeStepWithRetry(ctx, stepExecution, stepDef, sagaCtx, stepInput)
}

// getOrCreateStepExecution 获取或创建步骤执行记录
func (e *executionEngine) getOrCreateStepExecution(ctx context.Context, sagaID uint,
	stepDef SagaStepDefinition, stepIndex int) (*model.SagaStepExecution, error) {

	// 尝试获取现有的步骤执行记录
	steps, err := e.stateManager.GetSagaSteps(ctx, sagaID)
	if err != nil {
		return nil, err
	}

	// 查找对应的步骤
	for _, step := range steps {
		if step.StepIndex == stepIndex && step.StepName == stepDef.Name {
			return step, nil
		}
	}

	// 创建新的步骤执行记录
	stepExecution := &model.SagaStepExecution{
		SagaExecutionID: sagaID,
		StepName:        stepDef.Name,
		StepIndex:       stepIndex,
		Status:          model.StepStatusPending,
		Input:           "", // 可以从Saga上下文中填充
		AttemptCount:    0,
		MaxRetries:      3,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := e.repository.CreateSagaStepExecution(ctx, stepExecution); err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeSerializationFailed,
			"failed to create step execution", sagaID, stepIndex, err)
	}

	return stepExecution, nil
}

// checkStepPrerequisites 检查步骤前置条件
func (e *executionEngine) checkStepPrerequisites(ctx context.Context, sagaID uint,
	prerequisites []string, currentStepIndex int) error {

	if len(prerequisites) == 0 {
		return nil
	}

	// 获取所有步骤
	steps, err := e.stateManager.GetSagaSteps(ctx, sagaID)
	if err != nil {
		return err
	}

	// 检查每个前置条件
	for _, prereq := range prerequisites {
		found := false
		for _, step := range steps {
			if step.StepName == prereq && step.StepIndex < currentStepIndex {
				if step.Status != model.StepStatusCompleted {
					return fmt.Errorf("prerequisite step '%s' is not completed (status: %s)", prereq, step.Status)
				}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("prerequisite step '%s' not found or not executed yet", prereq)
		}
	}

	return nil
}

// handleStepFailure 处理步骤失败
func (e *executionEngine) handleStepFailure(ctx context.Context, execution *model.SagaExecution,
	definition SagaDefinition, failedStepIndex int, stepErr error) error {

	// 记录步骤失败
	execution.Error = stepErr.Error()
	if err := e.repository.UpdateSagaExecution(ctx, execution); err != nil {
		return err
	}

	// 检查是否需要补偿
	if e.shouldCompensate(definition.GetSteps(), failedStepIndex) {
		// 启动补偿流程
		return e.CompensateSaga(ctx, execution, definition)
	}

	// 直接标记为失败
	return e.stateManager.TransitionSagaStatus(ctx, execution.ID,
		model.SagaStatusRunning, model.SagaStatusFailed)
}

// shouldCompensate 判断是否需要补偿
func (e *executionEngine) shouldCompensate(steps []SagaStepDefinition, failedStepIndex int) bool {
	// 检查之前的步骤中是否有需要补偿的关键步骤
	for i := 0; i < failedStepIndex; i++ {
		if steps[i].IsCritical && steps[i].Compensate != nil {
			return true
		}
	}
	return false
}

// handleSagaTimeout 处理Saga超时
func (e *executionEngine) handleSagaTimeout(ctx context.Context, execution *model.SagaExecution) error {
	execution.Error = "saga execution timeout"
	if err := e.repository.UpdateSagaExecution(ctx, execution); err != nil {
		return err
	}

	return e.stateManager.TransitionSagaStatus(ctx, execution.ID,
		model.SagaStatusRunning, model.SagaStatusTimeout)
}

// completeSaga 完成Saga
func (e *executionEngine) completeSaga(ctx context.Context, execution *model.SagaExecution) error {
	return e.stateManager.TransitionSagaStatus(ctx, execution.ID,
		model.SagaStatusRunning, model.SagaStatusCompleted)
}

// recordRetryEvent 记录重试事件
func (e *executionEngine) recordRetryEvent(ctx context.Context, stepExecution *model.SagaStepExecution,
	attemptCount int, err error) {

	event := &model.SagaEvent{
		SagaExecutionID: stepExecution.SagaExecutionID,
		EventType:       EventStepRetrying,
		StepName:        stepExecution.StepName,
		StepIndex:       stepExecution.StepIndex,
		Error:           err.Error(),
		EventData:       fmt.Sprintf(`{"attempt": %d, "error": "%s"}`, attemptCount, err.Error()),
		CreatedAt:       time.Now(),
	}

	e.stateManager.RecordSagaEvent(ctx, event)
}
