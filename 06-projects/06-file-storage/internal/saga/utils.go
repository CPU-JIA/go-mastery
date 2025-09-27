package saga

import (
	"context"
	"file-storage-service/internal/model"
	"fmt"
	"sync"
	"time"
)

// lockManager 分布式锁管理器
type lockManager struct {
	locks   map[string]*sagaLock
	mutex   sync.RWMutex
	timeout time.Duration
}

// NewLockManager 创建锁管理器
func NewLockManager() *lockManager {
	return &lockManager{
		locks:   make(map[string]*sagaLock),
		timeout: DefaultLockTimeout,
	}
}

// AcquireLock 获取锁
func (lm *lockManager) AcquireLock(ctx context.Context, key string, timeout time.Duration) (SagaLock, error) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	// 检查是否已经有锁存在
	if existingLock, exists := lm.locks[key]; exists {
		if existingLock.IsLocked() {
			return nil, NewSagaError(ErrCodeSagaLockFailed,
				fmt.Sprintf("lock %s is already held", key), 0, -1)
		}
		// 清理过期锁
		delete(lm.locks, key)
	}

	// 创建新锁
	lock := &sagaLock{
		key:        key,
		acquiredAt: time.Now(),
		expiresAt:  time.Now().Add(timeout),
		manager:    lm,
		locked:     true,
	}

	lm.locks[key] = lock

	// 启动自动清理协程
	go lm.autoCleanup(key, timeout)

	return lock, nil
}

// autoCleanup 自动清理过期锁
func (lm *lockManager) autoCleanup(key string, timeout time.Duration) {
	time.Sleep(timeout + time.Second) // 稍微延长一点时间确保锁已过期

	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if lock, exists := lm.locks[key]; exists {
		if !lock.IsLocked() {
			delete(lm.locks, key)
		}
	}
}

// releaseLock 释放锁（内部方法）
func (lm *lockManager) releaseLock(key string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if lock, exists := lm.locks[key]; exists {
		lock.locked = false
		delete(lm.locks, key)
	}
}

// sagaLock Saga锁实现
type sagaLock struct {
	key        string
	acquiredAt time.Time
	expiresAt  time.Time
	manager    *lockManager
	locked     bool
	mutex      sync.RWMutex
}

// Unlock 释放锁
func (sl *sagaLock) Unlock() error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	if !sl.locked {
		return NewSagaError(ErrCodeSagaLockFailed,
			fmt.Sprintf("lock %s is not held", sl.key), 0, -1)
	}

	sl.locked = false
	sl.manager.releaseLock(sl.key)
	return nil
}

// IsLocked 检查锁是否仍然有效
func (sl *sagaLock) IsLocked() bool {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	if !sl.locked {
		return false
	}

	// 检查是否已过期
	if time.Now().After(sl.expiresAt) {
		return false
	}

	return true
}

// Extend 延长锁的有效时间
func (sl *sagaLock) Extend(duration time.Duration) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	if !sl.IsLocked() {
		return NewSagaError(ErrCodeSagaLockFailed,
			fmt.Sprintf("cannot extend expired lock %s", sl.key), 0, -1)
	}

	sl.expiresAt = time.Now().Add(duration)
	return nil
}

// EventBus 事件总线接口
type EventBus interface {
	// Subscribe 订阅事件
	Subscribe(eventType string, handler EventHandler)

	// Unsubscribe 取消订阅
	Unsubscribe(eventType string, handler EventHandler)

	// Publish 发布事件
	Publish(event interface{})

	// PublishAsync 异步发布事件
	PublishAsync(event interface{})

	// Close 关闭事件总线
	Close()
}

// EventHandler 事件处理函数
type EventHandler func(event interface{})

// eventBus 事件总线实现
type eventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
	closed   bool
	workers  chan interface{}
}

// NewEventBus 创建事件总线
func NewEventBus() EventBus {
	bus := &eventBus{
		handlers: make(map[string][]EventHandler),
		workers:  make(chan interface{}, 100), // 100个缓冲区
	}

	// 启动异步处理协程
	go bus.processEvents()

	return bus
}

// Subscribe 订阅事件
func (eb *eventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.closed {
		return
	}

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

// Unsubscribe 取消订阅
func (eb *eventBus) Unsubscribe(eventType string, handler EventHandler) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.closed {
		return
	}

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		// 使用函数指针比较（在实际实现中可能需要更复杂的比较逻辑）
		if &h == &handler {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish 发布事件
func (eb *eventBus) Publish(event interface{}) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	if eb.closed {
		return
	}

	eventType := eb.getEventType(event)
	if handlers, exists := eb.handlers[eventType]; exists {
		for _, handler := range handlers {
			// 同步处理事件，如果发生panic不影响其他处理器
			func() {
				defer func() {
					if r := recover(); r != nil {
						// TODO: 记录错误日志
					}
				}()
				handler(event)
			}()
		}
	}
}

// PublishAsync 异步发布事件
func (eb *eventBus) PublishAsync(event interface{}) {
	if eb.closed {
		return
	}

	select {
	case eb.workers <- event:
		// 事件已加入队列
	default:
		// 队列已满，丢弃事件（或者可以选择阻塞）
	}
}

// processEvents 处理异步事件
func (eb *eventBus) processEvents() {
	for event := range eb.workers {
		if eb.closed {
			break
		}
		eb.Publish(event)
	}
}

// Close 关闭事件总线
func (eb *eventBus) Close() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.closed {
		return
	}

	eb.closed = true
	close(eb.workers)
}

// getEventType 获取事件类型
func (eb *eventBus) getEventType(event interface{}) string {
	switch event.(type) {
	case SagaStatusChangedEvent:
		return "SagaStatusChanged"
	default:
		return fmt.Sprintf("%T", event)
	}
}

// RetryManager 重试管理器
type RetryManager struct {
	defaultPolicy *model.RetryPolicy
}

// NewRetryManager 创建重试管理器
func NewRetryManager() *RetryManager {
	return &RetryManager{
		defaultPolicy: model.DefaultRetryPolicy(),
	}
}

// ShouldRetry 判断是否应该重试
func (rm *RetryManager) ShouldRetry(err error, attemptCount int, policy *model.RetryPolicy) bool {
	if policy == nil {
		policy = rm.defaultPolicy
	}

	// 检查是否达到最大重试次数
	if attemptCount >= policy.MaxAttempts {
		return false
	}

	// 检查是否为可重试的错误
	return IsRetryableError(err)
}

// CalculateDelay 计算重试延迟
func (rm *RetryManager) CalculateDelay(attemptCount int, policy *model.RetryPolicy) time.Duration {
	if policy == nil {
		policy = rm.defaultPolicy
	}

	// 指数退避算法
	delay := policy.InitialDelay
	for i := 0; i < attemptCount; i++ {
		delay = time.Duration(float64(delay) * policy.BackoffFactor)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
			break
		}
	}

	return delay
}

// IsRetryableErrorType 检查错误类型是否可重试
func (rm *RetryManager) IsRetryableErrorType(err error, policy *model.RetryPolicy) bool {
	if policy == nil || len(policy.RetryableErrors) == 0 {
		return IsRetryableError(err)
	}

	errorMsg := err.Error()
	for _, retryableError := range policy.RetryableErrors {
		if contains(errorMsg, retryableError) {
			return true
		}
	}

	return false
}

// TimeoutManager 超时管理器
type TimeoutManager struct {
	defaultTimeout time.Duration
}

// NewTimeoutManager 创建超时管理器
func NewTimeoutManager() *TimeoutManager {
	return &TimeoutManager{
		defaultTimeout: DefaultSagaTimeout,
	}
}

// CreateTimeoutContext 创建带超时的上下文
func (tm *TimeoutManager) CreateTimeoutContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = tm.defaultTimeout
	}

	return context.WithTimeout(ctx, timeout)
}

// IsTimeoutError 检查是否为超时错误
func (tm *TimeoutManager) IsTimeoutError(err error) bool {
	if err == context.DeadlineExceeded {
		return true
	}

	if sagaErr, ok := err.(*SagaError); ok {
		return sagaErr.Code == ErrCodeSagaTimeout || sagaErr.Code == ErrCodeStepTimeout
	}

	return contains(err.Error(), "timeout") || contains(err.Error(), "deadline exceeded")
}
