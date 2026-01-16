// Package main 演示分布式系统设计模式
// 本模块涵盖常见的分布式系统架构模式，包括：
// - Saga 模式：分布式事务管理
// - CQRS 模式：命令查询职责分离
// - Event Sourcing：事件溯源
// - Sidecar 模式：服务网格基础
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Saga 模式 - 分布式事务管理
// ============================================================================

// SagaStep 定义 Saga 中的一个步骤
type SagaStep struct {
	Name       string
	Execute    func(ctx context.Context, data interface{}) error
	Compensate func(ctx context.Context, data interface{}) error
}

// SagaOrchestrator 编排 Saga 事务
type SagaOrchestrator struct {
	steps          []SagaStep
	completedSteps []string
	mu             sync.Mutex
}

// NewSagaOrchestrator 创建新的 Saga 编排器
func NewSagaOrchestrator() *SagaOrchestrator {
	return &SagaOrchestrator{
		steps:          make([]SagaStep, 0),
		completedSteps: make([]string, 0),
	}
}

// AddStep 添加 Saga 步骤
func (s *SagaOrchestrator) AddStep(step SagaStep) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.steps = append(s.steps, step)
}

// Execute 执行 Saga 事务
func (s *SagaOrchestrator) Execute(ctx context.Context, data interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completedSteps = make([]string, 0)

	for _, step := range s.steps {
		fmt.Printf("  [Saga] 执行步骤: %s\n", step.Name)

		if err := step.Execute(ctx, data); err != nil {
			fmt.Printf("  [Saga] 步骤 %s 失败: %v\n", step.Name, err)
			s.compensate(ctx, data)
			return fmt.Errorf("saga 执行失败于步骤 %s: %w", step.Name, err)
		}

		s.completedSteps = append(s.completedSteps, step.Name)
		fmt.Printf("  [Saga] 步骤 %s 完成\n", step.Name)
	}

	return nil
}

// compensate 执行补偿操作（回滚）
func (s *SagaOrchestrator) compensate(ctx context.Context, data interface{}) {
	fmt.Println("  [Saga] 开始补偿操作...")

	// 逆序执行补偿
	for i := len(s.completedSteps) - 1; i >= 0; i-- {
		stepName := s.completedSteps[i]

		for _, step := range s.steps {
			if step.Name == stepName && step.Compensate != nil {
				fmt.Printf("  [Saga] 补偿步骤: %s\n", stepName)
				if err := step.Compensate(ctx, data); err != nil {
					fmt.Printf("  [Saga] 补偿失败: %v\n", err)
				}
				break
			}
		}
	}

	fmt.Println("  [Saga] 补偿操作完成")
}

// ============================================================================
// CQRS 模式 - 命令查询职责分离
// ============================================================================

// Command 表示一个命令
type Command interface {
	CommandName() string
}

// Query 表示一个查询
type Query interface {
	QueryName() string
}

// CommandHandler 命令处理器接口
type CommandHandler interface {
	Handle(ctx context.Context, cmd Command) error
}

// QueryHandler 查询处理器接口
type QueryHandler interface {
	Handle(ctx context.Context, query Query) (interface{}, error)
}

// CreateOrderCommand 创建订单命令
type CreateOrderCommand struct {
	OrderID    string
	CustomerID string
	Items      []OrderItem
	TotalPrice float64
}

func (c CreateOrderCommand) CommandName() string { return "CreateOrder" }

// OrderItem 订单项
type OrderItem struct {
	ProductID string
	Quantity  int
	Price     float64
}

// GetOrderQuery 获取订单查询
type GetOrderQuery struct {
	OrderID string
}

func (q GetOrderQuery) QueryName() string { return "GetOrder" }

// OrderReadModel 订单读模型
type OrderReadModel struct {
	OrderID     string      `json:"order_id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalPrice  float64     `json:"total_price"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	LastUpdated time.Time   `json:"last_updated"`
}

// CQRSBus CQRS 总线
type CQRSBus struct {
	commandHandlers map[string]CommandHandler
	queryHandlers   map[string]QueryHandler
	mu              sync.RWMutex
}

// NewCQRSBus 创建 CQRS 总线
func NewCQRSBus() *CQRSBus {
	return &CQRSBus{
		commandHandlers: make(map[string]CommandHandler),
		queryHandlers:   make(map[string]QueryHandler),
	}
}

// RegisterCommandHandler 注册命令处理器
func (b *CQRSBus) RegisterCommandHandler(cmdName string, handler CommandHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.commandHandlers[cmdName] = handler
}

// RegisterQueryHandler 注册查询处理器
func (b *CQRSBus) RegisterQueryHandler(queryName string, handler QueryHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queryHandlers[queryName] = handler
}

// SendCommand 发送命令
func (b *CQRSBus) SendCommand(ctx context.Context, cmd Command) error {
	b.mu.RLock()
	handler, exists := b.commandHandlers[cmd.CommandName()]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("未找到命令处理器: %s", cmd.CommandName())
	}

	return handler.Handle(ctx, cmd)
}

// ExecuteQuery 执行查询
func (b *CQRSBus) ExecuteQuery(ctx context.Context, query Query) (interface{}, error) {
	b.mu.RLock()
	handler, exists := b.queryHandlers[query.QueryName()]
	b.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("未找到查询处理器: %s", query.QueryName())
	}

	return handler.Handle(ctx, query)
}

// ============================================================================
// Event Sourcing - 事件溯源
// ============================================================================

// Event 事件接口
type Event interface {
	EventType() string
	AggregateID() string
	Timestamp() time.Time
	Version() int
}

// BaseEvent 基础事件
type BaseEvent struct {
	Type        string    `json:"type"`
	AggregateId string    `json:"aggregate_id"`
	Time        time.Time `json:"timestamp"`
	Ver         int       `json:"version"`
}

func (e BaseEvent) EventType() string    { return e.Type }
func (e BaseEvent) AggregateID() string  { return e.AggregateId }
func (e BaseEvent) Timestamp() time.Time { return e.Time }
func (e BaseEvent) Version() int         { return e.Ver }

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	BaseEvent
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	TotalPrice float64     `json:"total_price"`
}

// OrderPaidEvent 订单支付事件
type OrderPaidEvent struct {
	BaseEvent
	PaymentID     string    `json:"payment_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaidAt        time.Time `json:"paid_at"`
}

// OrderShippedEvent 订单发货事件
type OrderShippedEvent struct {
	BaseEvent
	ShipmentID     string    `json:"shipment_id"`
	Carrier        string    `json:"carrier"`
	TrackingNumber string    `json:"tracking_number"`
	ShippedAt      time.Time `json:"shipped_at"`
}

// EventStore 事件存储
type EventStore struct {
	events map[string][]Event
	mu     sync.RWMutex
}

// NewEventStore 创建事件存储
func NewEventStore() *EventStore {
	return &EventStore{
		events: make(map[string][]Event),
	}
}

// Append 追加事件
func (es *EventStore) Append(event Event) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	aggregateID := event.AggregateID()
	es.events[aggregateID] = append(es.events[aggregateID], event)

	fmt.Printf("  [EventStore] 存储事件: %s (聚合ID: %s, 版本: %d)\n",
		event.EventType(), aggregateID, event.Version())

	return nil
}

// GetEvents 获取聚合的所有事件
func (es *EventStore) GetEvents(aggregateID string) []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	return es.events[aggregateID]
}

// GetEventsSince 获取指定版本之后的事件
func (es *EventStore) GetEventsSince(aggregateID string, version int) []Event {
	es.mu.RLock()
	defer es.mu.RUnlock()

	events := es.events[aggregateID]
	result := make([]Event, 0)

	for _, event := range events {
		if event.Version() > version {
			result = append(result, event)
		}
	}

	return result
}

// OrderAggregate 订单聚合根
type OrderAggregate struct {
	ID         string
	CustomerID string
	Items      []OrderItem
	TotalPrice float64
	Status     string
	Version    int
	events     []Event
}

// NewOrderAggregate 创建订单聚合
func NewOrderAggregate(id string) *OrderAggregate {
	return &OrderAggregate{
		ID:     id,
		Status: "pending",
		events: make([]Event, 0),
	}
}

// Apply 应用事件到聚合
func (o *OrderAggregate) Apply(event Event) {
	switch e := event.(type) {
	case *OrderCreatedEvent:
		o.CustomerID = e.CustomerID
		o.Items = e.Items
		o.TotalPrice = e.TotalPrice
		o.Status = "created"
	case *OrderPaidEvent:
		o.Status = "paid"
	case *OrderShippedEvent:
		o.Status = "shipped"
	}
	o.Version = event.Version()
}

// Rebuild 从事件重建聚合状态
func (o *OrderAggregate) Rebuild(events []Event) {
	for _, event := range events {
		o.Apply(event)
	}
}

// ============================================================================
// Sidecar 模式 - 服务网格基础
// ============================================================================

// Sidecar 边车代理
type Sidecar struct {
	serviceName string
	config      SidecarConfig
	metrics     *SidecarMetrics
	mu          sync.RWMutex
}

// SidecarConfig 边车配置
type SidecarConfig struct {
	RetryAttempts   int
	RetryDelay      time.Duration
	CircuitBreaker  bool
	RateLimitRPS    int
	TracingEnabled  bool
	MetricsEnabled  bool
	HealthCheckPath string
}

// SidecarMetrics 边车指标
type SidecarMetrics struct {
	RequestCount    int64
	SuccessCount    int64
	FailureCount    int64
	TotalLatency    time.Duration
	CircuitOpen     bool
	LastRequestTime time.Time
}

// NewSidecar 创建边车代理
func NewSidecar(serviceName string, config SidecarConfig) *Sidecar {
	return &Sidecar{
		serviceName: serviceName,
		config:      config,
		metrics:     &SidecarMetrics{},
	}
}

// Intercept 拦截请求
func (s *Sidecar) Intercept(ctx context.Context, request func() error) error {
	s.mu.Lock()
	s.metrics.RequestCount++
	s.metrics.LastRequestTime = time.Now()
	s.mu.Unlock()

	startTime := time.Now()

	// 熔断检查
	if s.config.CircuitBreaker && s.metrics.CircuitOpen {
		fmt.Printf("  [Sidecar:%s] 熔断器开启，拒绝请求\n", s.serviceName)
		return fmt.Errorf("circuit breaker open")
	}

	// 重试逻辑
	var lastErr error
	for attempt := 0; attempt <= s.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			fmt.Printf("  [Sidecar:%s] 重试第 %d 次\n", s.serviceName, attempt)
			time.Sleep(s.config.RetryDelay)
		}

		lastErr = request()
		if lastErr == nil {
			s.mu.Lock()
			s.metrics.SuccessCount++
			s.metrics.TotalLatency += time.Since(startTime)
			s.mu.Unlock()

			fmt.Printf("  [Sidecar:%s] 请求成功 (耗时: %v)\n",
				s.serviceName, time.Since(startTime))
			return nil
		}
	}

	s.mu.Lock()
	s.metrics.FailureCount++
	s.metrics.TotalLatency += time.Since(startTime)

	// 检查是否需要开启熔断
	failureRate := float64(s.metrics.FailureCount) / float64(s.metrics.RequestCount)
	if failureRate > 0.5 && s.metrics.RequestCount > 10 {
		s.metrics.CircuitOpen = true
		fmt.Printf("  [Sidecar:%s] 熔断器开启 (失败率: %.2f%%)\n",
			s.serviceName, failureRate*100)
	}
	s.mu.Unlock()

	return lastErr
}

// GetMetrics 获取指标
func (s *Sidecar) GetMetrics() SidecarMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.metrics
}

// ============================================================================
// 演示函数
// ============================================================================

func demonstrateSaga() {
	fmt.Println("\n=== Saga 模式演示 ===")
	fmt.Println("场景: 电商订单创建流程（库存扣减 -> 支付 -> 发货）")

	saga := NewSagaOrchestrator()

	// 添加库存扣减步骤
	saga.AddStep(SagaStep{
		Name: "扣减库存",
		Execute: func(ctx context.Context, data interface{}) error {
			fmt.Println("    -> 正在扣减库存...")
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		Compensate: func(ctx context.Context, data interface{}) error {
			fmt.Println("    <- 恢复库存...")
			return nil
		},
	})

	// 添加支付步骤
	saga.AddStep(SagaStep{
		Name: "处理支付",
		Execute: func(ctx context.Context, data interface{}) error {
			fmt.Println("    -> 正在处理支付...")
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		Compensate: func(ctx context.Context, data interface{}) error {
			fmt.Println("    <- 退款处理...")
			return nil
		},
	})

	// 添加发货步骤
	saga.AddStep(SagaStep{
		Name: "创建发货单",
		Execute: func(ctx context.Context, data interface{}) error {
			fmt.Println("    -> 正在创建发货单...")
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		Compensate: func(ctx context.Context, data interface{}) error {
			fmt.Println("    <- 取消发货单...")
			return nil
		},
	})

	ctx := context.Background()
	orderData := map[string]interface{}{
		"order_id":    "ORD-001",
		"customer_id": "CUST-001",
		"amount":      299.99,
	}

	fmt.Println("\n执行成功场景:")
	if err := saga.Execute(ctx, orderData); err != nil {
		fmt.Printf("Saga 执行失败: %v\n", err)
	} else {
		fmt.Println("Saga 执行成功!")
	}

	// 演示失败场景
	fmt.Println("\n执行失败场景 (模拟支付失败):")
	sagaWithFailure := NewSagaOrchestrator()

	sagaWithFailure.AddStep(SagaStep{
		Name: "扣减库存",
		Execute: func(ctx context.Context, data interface{}) error {
			fmt.Println("    -> 正在扣减库存...")
			return nil
		},
		Compensate: func(ctx context.Context, data interface{}) error {
			fmt.Println("    <- 恢复库存...")
			return nil
		},
	})

	sagaWithFailure.AddStep(SagaStep{
		Name: "处理支付",
		Execute: func(ctx context.Context, data interface{}) error {
			fmt.Println("    -> 正在处理支付...")
			return fmt.Errorf("支付网关超时")
		},
		Compensate: func(ctx context.Context, data interface{}) error {
			fmt.Println("    <- 退款处理...")
			return nil
		},
	})

	if err := sagaWithFailure.Execute(ctx, orderData); err != nil {
		fmt.Printf("Saga 执行失败 (预期): %v\n", err)
	}
}

func demonstrateCQRS() {
	fmt.Println("\n=== CQRS 模式演示 ===")
	fmt.Println("场景: 订单系统的命令和查询分离")

	bus := NewCQRSBus()

	// 模拟写模型存储
	writeStore := make(map[string]interface{})
	// 模拟读模型存储
	readStore := make(map[string]*OrderReadModel)

	// 注册命令处理器
	bus.RegisterCommandHandler("CreateOrder", &createOrderHandler{
		writeStore: writeStore,
		readStore:  readStore,
	})

	// 注册查询处理器
	bus.RegisterQueryHandler("GetOrder", &getOrderHandler{
		readStore: readStore,
	})

	ctx := context.Background()

	// 发送创建订单命令
	fmt.Println("\n发送创建订单命令:")
	cmd := CreateOrderCommand{
		OrderID:    "ORD-002",
		CustomerID: "CUST-002",
		Items: []OrderItem{
			{ProductID: "PROD-001", Quantity: 2, Price: 99.99},
			{ProductID: "PROD-002", Quantity: 1, Price: 149.99},
		},
		TotalPrice: 349.97,
	}

	if err := bus.SendCommand(ctx, cmd); err != nil {
		fmt.Printf("命令执行失败: %v\n", err)
		return
	}

	// 执行查询
	fmt.Println("\n执行订单查询:")
	query := GetOrderQuery{OrderID: "ORD-002"}
	result, err := bus.ExecuteQuery(ctx, query)
	if err != nil {
		fmt.Printf("查询执行失败: %v\n", err)
		return
	}

	if order, ok := result.(*OrderReadModel); ok {
		orderJSON, _ := json.MarshalIndent(order, "  ", "  ")
		fmt.Printf("  查询结果:\n  %s\n", string(orderJSON))
	}
}

// createOrderHandler 创建订单命令处理器
type createOrderHandler struct {
	writeStore map[string]interface{}
	readStore  map[string]*OrderReadModel
}

func (h *createOrderHandler) Handle(ctx context.Context, cmd Command) error {
	createCmd := cmd.(CreateOrderCommand)

	fmt.Printf("  [CommandHandler] 处理创建订单命令: %s\n", createCmd.OrderID)

	// 写入写模型
	h.writeStore[createCmd.OrderID] = createCmd

	// 同步到读模型（实际场景中可能是异步的）
	h.readStore[createCmd.OrderID] = &OrderReadModel{
		OrderID:     createCmd.OrderID,
		CustomerID:  createCmd.CustomerID,
		Items:       createCmd.Items,
		TotalPrice:  createCmd.TotalPrice,
		Status:      "created",
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}

	fmt.Printf("  [CommandHandler] 订单创建成功，已同步到读模型\n")
	return nil
}

// getOrderHandler 获取订单查询处理器
type getOrderHandler struct {
	readStore map[string]*OrderReadModel
}

func (h *getOrderHandler) Handle(ctx context.Context, query Query) (interface{}, error) {
	getQuery := query.(GetOrderQuery)

	fmt.Printf("  [QueryHandler] 处理订单查询: %s\n", getQuery.OrderID)

	order, exists := h.readStore[getQuery.OrderID]
	if !exists {
		return nil, fmt.Errorf("订单不存在: %s", getQuery.OrderID)
	}

	return order, nil
}

func demonstrateEventSourcing() {
	fmt.Println("\n=== Event Sourcing 模式演示 ===")
	fmt.Println("场景: 通过事件重建订单状态")

	eventStore := NewEventStore()
	orderID := "ORD-003"

	// 创建并存储事件
	fmt.Println("\n存储事件序列:")

	// 事件1: 订单创建
	event1 := &OrderCreatedEvent{
		BaseEvent: BaseEvent{
			Type:        "OrderCreated",
			AggregateId: orderID,
			Time:        time.Now(),
			Ver:         1,
		},
		CustomerID: "CUST-003",
		Items: []OrderItem{
			{ProductID: "PROD-001", Quantity: 1, Price: 199.99},
		},
		TotalPrice: 199.99,
	}
	eventStore.Append(event1)

	// 事件2: 订单支付
	event2 := &OrderPaidEvent{
		BaseEvent: BaseEvent{
			Type:        "OrderPaid",
			AggregateId: orderID,
			Time:        time.Now().Add(time.Minute),
			Ver:         2,
		},
		PaymentID:     "PAY-001",
		Amount:        199.99,
		PaymentMethod: "credit_card",
		PaidAt:        time.Now().Add(time.Minute),
	}
	eventStore.Append(event2)

	// 事件3: 订单发货
	event3 := &OrderShippedEvent{
		BaseEvent: BaseEvent{
			Type:        "OrderShipped",
			AggregateId: orderID,
			Time:        time.Now().Add(2 * time.Hour),
			Ver:         3,
		},
		ShipmentID:     "SHIP-001",
		Carrier:        "顺丰快递",
		TrackingNumber: "SF1234567890",
		ShippedAt:      time.Now().Add(2 * time.Hour),
	}
	eventStore.Append(event3)

	// 从事件重建聚合状态
	fmt.Println("\n从事件重建订单状态:")
	order := NewOrderAggregate(orderID)
	events := eventStore.GetEvents(orderID)

	fmt.Printf("  共有 %d 个事件\n", len(events))

	for _, event := range events {
		fmt.Printf("  应用事件: %s (版本: %d)\n", event.EventType(), event.Version())
		order.Apply(event)
	}

	fmt.Printf("\n  重建后的订单状态:\n")
	fmt.Printf("    订单ID: %s\n", order.ID)
	fmt.Printf("    客户ID: %s\n", order.CustomerID)
	fmt.Printf("    总金额: %.2f\n", order.TotalPrice)
	fmt.Printf("    状态: %s\n", order.Status)
	fmt.Printf("    版本: %d\n", order.Version)

	// 演示部分重建
	fmt.Println("\n从版本1开始重建 (只应用后续事件):")
	partialEvents := eventStore.GetEventsSince(orderID, 1)
	fmt.Printf("  需要应用 %d 个事件\n", len(partialEvents))
}

func demonstrateSidecar() {
	fmt.Println("\n=== Sidecar 模式演示 ===")
	fmt.Println("场景: 边车代理处理服务间通信")

	config := SidecarConfig{
		RetryAttempts:   3,
		RetryDelay:      100 * time.Millisecond,
		CircuitBreaker:  true,
		RateLimitRPS:    100,
		TracingEnabled:  true,
		MetricsEnabled:  true,
		HealthCheckPath: "/health",
	}

	sidecar := NewSidecar("user-service", config)

	ctx := context.Background()

	// 成功请求
	fmt.Println("\n模拟成功请求:")
	err := sidecar.Intercept(ctx, func() error {
		fmt.Println("    -> 执行实际服务调用...")
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
	}

	// 需要重试的请求
	fmt.Println("\n模拟需要重试的请求:")
	retryCount := 0
	err = sidecar.Intercept(ctx, func() error {
		retryCount++
		if retryCount < 3 {
			return fmt.Errorf("临时错误")
		}
		fmt.Println("    -> 第三次尝试成功!")
		return nil
	})
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
	}

	// 显示指标
	metrics := sidecar.GetMetrics()
	fmt.Printf("\n边车指标统计:\n")
	fmt.Printf("  总请求数: %d\n", metrics.RequestCount)
	fmt.Printf("  成功数: %d\n", metrics.SuccessCount)
	fmt.Printf("  失败数: %d\n", metrics.FailureCount)
	fmt.Printf("  熔断器状态: %v\n", metrics.CircuitOpen)
	if metrics.RequestCount > 0 {
		avgLatency := metrics.TotalLatency / time.Duration(metrics.RequestCount)
		fmt.Printf("  平均延迟: %v\n", avgLatency)
	}
}

func main() {
	fmt.Println("=== 分布式系统设计模式 ===")
	fmt.Println()
	fmt.Println("本模块演示四种核心分布式系统设计模式:")
	fmt.Println("1. Saga 模式 - 分布式事务管理")
	fmt.Println("2. CQRS 模式 - 命令查询职责分离")
	fmt.Println("3. Event Sourcing - 事件溯源")
	fmt.Println("4. Sidecar 模式 - 服务网格基础")

	demonstrateSaga()
	demonstrateCQRS()
	demonstrateEventSourcing()
	demonstrateSidecar()

	fmt.Println("\n=== 分布式模式演示完成 ===")
	fmt.Println()
	fmt.Println("关键学习点:")
	fmt.Println("- Saga: 通过补偿操作实现分布式事务的最终一致性")
	fmt.Println("- CQRS: 分离读写模型，优化查询性能和扩展性")
	fmt.Println("- Event Sourcing: 通过事件序列重建状态，支持审计和时间旅行")
	fmt.Println("- Sidecar: 将横切关注点（重试、熔断、监控）从业务逻辑中分离")
}
