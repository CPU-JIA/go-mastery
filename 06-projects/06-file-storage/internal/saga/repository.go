package saga

import (
	"context"
	"file-storage-service/internal/model"
	"time"

	"gorm.io/gorm"
)

// sagaRepository Saga数据访问实现
type sagaRepository struct {
	db *gorm.DB
}

// NewSagaRepository 创建Saga仓储
func NewSagaRepository(db *gorm.DB) SagaRepository {
	return &sagaRepository{db: db}
}

// CreateSagaExecution 创建Saga执行记录
func (r *sagaRepository) CreateSagaExecution(ctx context.Context, execution *model.SagaExecution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

// UpdateSagaExecution 更新Saga执行记录
func (r *sagaRepository) UpdateSagaExecution(ctx context.Context, execution *model.SagaExecution) error {
	return r.db.WithContext(ctx).Save(execution).Error
}

// GetSagaExecution 获取Saga执行记录
func (r *sagaRepository) GetSagaExecution(ctx context.Context, id uint) (*model.SagaExecution, error) {
	var execution model.SagaExecution
	err := r.db.WithContext(ctx).Preload("Steps").First(&execution, id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// CreateSagaStepExecution 创建步骤执行记录
func (r *sagaRepository) CreateSagaStepExecution(ctx context.Context, stepExecution *model.SagaStepExecution) error {
	return r.db.WithContext(ctx).Create(stepExecution).Error
}

// UpdateSagaStepExecution 更新步骤执行记录
func (r *sagaRepository) UpdateSagaStepExecution(ctx context.Context, stepExecution *model.SagaStepExecution) error {
	return r.db.WithContext(ctx).Save(stepExecution).Error
}

// GetSagaStepExecution 获取步骤执行记录
func (r *sagaRepository) GetSagaStepExecution(ctx context.Context, id uint) (*model.SagaStepExecution, error) {
	var stepExecution model.SagaStepExecution
	err := r.db.WithContext(ctx).First(&stepExecution, id).Error
	if err != nil {
		return nil, err
	}
	return &stepExecution, nil
}

// ListSagaStepsByExecution 根据Saga ID获取所有步骤
func (r *sagaRepository) ListSagaStepsByExecution(ctx context.Context, sagaExecutionID uint) ([]*model.SagaStepExecution, error) {
	var steps []*model.SagaStepExecution
	err := r.db.WithContext(ctx).
		Where("saga_execution_id = ?", sagaExecutionID).
		Order("step_index ASC").
		Find(&steps).Error
	return steps, err
}

// RecordSagaEvent 记录Saga事件
func (r *sagaRepository) RecordSagaEvent(ctx context.Context, event *model.SagaEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// GetRunningSagas 获取运行中的Saga
func (r *sagaRepository) GetRunningSagas(ctx context.Context, limit int) ([]*model.SagaExecution, error) {
	var sagas []*model.SagaExecution
	query := r.db.WithContext(ctx).Where("status = ?", model.SagaStatusRunning)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at ASC").Find(&sagas).Error
	return sagas, err
}

// GetTimeoutSagas 获取超时的Saga
func (r *sagaRepository) GetTimeoutSagas(ctx context.Context, before time.Time) ([]*model.SagaExecution, error) {
	var sagas []*model.SagaExecution
	err := r.db.WithContext(ctx).
		Where("status IN ? AND timeout_at IS NOT NULL AND timeout_at < ?",
			[]model.SagaStatus{model.SagaStatusRunning, model.SagaStatusPending}, before).
		Find(&sagas).Error
	return sagas, err
}

// WorkerPool 工作池
type WorkerPool struct {
	workers    int
	jobChannel chan func()
	quit       chan bool
	active     bool
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount int) *WorkerPool {
	pool := &WorkerPool{
		workers:    workerCount,
		jobChannel: make(chan func(), workerCount*2), // 缓冲队列
		quit:       make(chan bool),
		active:     true,
	}

	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		go pool.worker(i)
	}

	return pool
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	for {
		select {
		case job := <-wp.jobChannel:
			if job != nil {
				// 执行任务，捕获panic
				func() {
					defer func() {
						if r := recover(); r != nil {
							// TODO: 记录panic日志
						}
					}()
					job()
				}()
			}
		case <-wp.quit:
			return
		}
	}
}

// Submit 提交任务
func (wp *WorkerPool) Submit(job func()) {
	if !wp.active {
		return
	}

	select {
	case wp.jobChannel <- job:
		// 任务已提交
	default:
		// 队列已满，可以选择阻塞或丢弃
		go job() // 异步执行避免阻塞
	}
}

// Close 关闭工作池
func (wp *WorkerPool) Close() {
	if !wp.active {
		return
	}

	wp.active = false
	close(wp.quit)
	close(wp.jobChannel)
}
