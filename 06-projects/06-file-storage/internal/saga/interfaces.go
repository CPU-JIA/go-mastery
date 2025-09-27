package saga

import (
	"context"
	"file-storage-service/internal/model"
	"time"
)

// SagaEngine 主要的Saga执行引擎接口
type SagaEngine interface {
	// StartSaga 启动新的Saga
	StartSaga(ctx context.Context, sagaType string, sagaContext interface{}) (*model.SagaExecution, error)

	// ResumeSaga 恢复中断的Saga（系统重启后）
	ResumeSaga(ctx context.Context, sagaID uint) error

	// CancelSaga 取消正在执行的Saga
	CancelSaga(ctx context.Context, sagaID uint) error

	// GetSagaStatus 获取Saga执行状态
	GetSagaStatus(ctx context.Context, sagaID uint) (*model.SagaExecution, error)

	// RegisterSagaDefinition 注册Saga定义（支持扩展）
	RegisterSagaDefinition(sagaType string, definition SagaDefinition) error

	// ListRunningSagas 列出正在运行的Saga
	ListRunningSagas(ctx context.Context, limit int) ([]*model.SagaExecution, error)
}

// SagaDefinition Saga类型定义接口
type SagaDefinition interface {
	// GetSagaType 获取Saga类型名称
	GetSagaType() string

	// GetSteps 获取步骤定义列表
	GetSteps() []SagaStepDefinition

	// ValidateContext 验证上下文数据
	ValidateContext(context interface{}) error

	// CalculateTimeout 计算超时时间
	CalculateTimeout(context interface{}) time.Duration

	// GetRetryPolicy 获取默认重试策略
	GetRetryPolicy() *model.RetryPolicy
}

// SagaStepDefinition Saga步骤定义
type SagaStepDefinition struct {
	Name          string                 // 步骤名称
	Execute       SagaStepExecuteFunc    // 执行函数
	Compensate    SagaStepCompensateFunc // 补偿函数
	RetryPolicy   *model.RetryPolicy     // 重试策略（可选，使用Saga默认）
	Timeout       time.Duration          // 步骤超时（可选，使用Saga默认）
	Prerequisites []string               // 前置条件步骤名称
	IsIdempotent  bool                   // 是否幂等操作
	IsCritical    bool                   // 是否关键步骤（失败必须补偿）
}

// SagaStepExecuteFunc 步骤执行函数类型
type SagaStepExecuteFunc func(ctx context.Context, sagaCtx interface{}, input interface{}) (output interface{}, err error)

// SagaStepCompensateFunc 步骤补偿函数类型
type SagaStepCompensateFunc func(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error

// SagaStateManager 状态管理器接口
type SagaStateManager interface {
	// 状态转换方法
	TransitionSagaStatus(ctx context.Context, sagaID uint, from, to model.SagaStatus) error
	TransitionStepStatus(ctx context.Context, stepID uint, from, to model.StepStatus) error

	// 原子性更新
	UpdateSagaProgress(ctx context.Context, sagaID uint, currentStep int, context string) error
	RecordStepResult(ctx context.Context, stepID uint, output string, error string) error

	// 锁定机制（防止并发执行同一Saga）
	AcquireSagaLock(ctx context.Context, sagaID uint, timeout time.Duration) (SagaLock, error)

	// 事件记录
	RecordSagaEvent(ctx context.Context, event *model.SagaEvent) error

	// 查询方法
	GetSagaExecution(ctx context.Context, sagaID uint) (*model.SagaExecution, error)
	GetSagaSteps(ctx context.Context, sagaID uint) ([]*model.SagaStepExecution, error)

	// 恢复相关
	GetRecoverableSagas(ctx context.Context) ([]*model.SagaExecution, error)
}

// SagaLock Saga锁接口
type SagaLock interface {
	// Unlock 释放锁
	Unlock() error

	// IsLocked 检查锁是否仍然有效
	IsLocked() bool

	// Extend 延长锁的有效时间
	Extend(duration time.Duration) error
}

// SagaExecutionEngine 执行引擎接口
type SagaExecutionEngine interface {
	// ExecuteSaga 执行Saga
	ExecuteSaga(ctx context.Context, execution *model.SagaExecution, definition SagaDefinition) error

	// ExecuteStep 执行单个步骤
	ExecuteStep(ctx context.Context, stepExecution *model.SagaStepExecution, stepDef SagaStepDefinition, sagaCtx interface{}) error

	// CompensateStep 补偿单个步骤
	CompensateStep(ctx context.Context, stepExecution *model.SagaStepExecution, stepDef SagaStepDefinition, sagaCtx interface{}) error

	// CompensateSaga 补偿整个Saga
	CompensateSaga(ctx context.Context, execution *model.SagaExecution, definition SagaDefinition) error

	// RetryStep 重试步骤
	RetryStep(ctx context.Context, stepExecution *model.SagaStepExecution, stepDef SagaStepDefinition, sagaCtx interface{}) error
}

// SagaOrchestrator Saga协调器接口
type SagaOrchestrator interface {
	// Orchestrate 协调整个Saga执行流程
	Orchestrate(ctx context.Context, sagaID uint) error

	// HandleTimeout 处理超时
	HandleTimeout(ctx context.Context, sagaID uint) error

	// HandleFailure 处理失败
	HandleFailure(ctx context.Context, sagaID uint, err error) error
}

// SagaRepository Saga数据访问接口
type SagaRepository interface {
	// CreateSagaExecution 创建Saga执行记录
	CreateSagaExecution(ctx context.Context, execution *model.SagaExecution) error

	// UpdateSagaExecution 更新Saga执行记录
	UpdateSagaExecution(ctx context.Context, execution *model.SagaExecution) error

	// GetSagaExecution 获取Saga执行记录
	GetSagaExecution(ctx context.Context, id uint) (*model.SagaExecution, error)

	// CreateSagaStepExecution 创建步骤执行记录
	CreateSagaStepExecution(ctx context.Context, stepExecution *model.SagaStepExecution) error

	// UpdateSagaStepExecution 更新步骤执行记录
	UpdateSagaStepExecution(ctx context.Context, stepExecution *model.SagaStepExecution) error

	// GetSagaStepExecution 获取步骤执行记录
	GetSagaStepExecution(ctx context.Context, id uint) (*model.SagaStepExecution, error)

	// ListSagaStepsByExecution 根据Saga ID获取所有步骤
	ListSagaStepsByExecution(ctx context.Context, sagaExecutionID uint) ([]*model.SagaStepExecution, error)

	// RecordSagaEvent 记录Saga事件
	RecordSagaEvent(ctx context.Context, event *model.SagaEvent) error

	// GetRunningSagas 获取运行中的Saga
	GetRunningSagas(ctx context.Context, limit int) ([]*model.SagaExecution, error)

	// GetTimeoutSagas 获取超时的Saga
	GetTimeoutSagas(ctx context.Context, before time.Time) ([]*model.SagaExecution, error)
}
