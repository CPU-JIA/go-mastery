package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
)

/*
微服务架构 - 消息队列与事件驱动练习

本练习涵盖微服务架构中的消息队列和事件驱动模式，包括：
1. 消息队列系统设计
2. 事件驱动架构
3. 发布-订阅模式
4. 消息路由和过滤
5. 死信队列处理
6. 消息持久化
7. 消息事务
8. 分布式事件溯源

主要概念：
- 异步消息传递
- 事件溯源模式
- CQRS模式
- Saga模式
- 事件总线
*/

// === 事件定义 ===

// Event 基础事件接口
type Event interface {
	GetEventID() string
	GetEventType() string
	GetAggregateID() string
	GetTimestamp() time.Time
	GetVersion() int
	GetPayload() interface{}
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	EventID     string      `json:"event_id"`
	EventType   string      `json:"event_type"`
	AggregateID string      `json:"aggregate_id"`
	Timestamp   time.Time   `json:"timestamp"`
	Version     int         `json:"version"`
	Payload     interface{} `json:"payload"`
}

func (e *BaseEvent) GetEventID() string      { return e.EventID }
func (e *BaseEvent) GetEventType() string    { return e.EventType }
func (e *BaseEvent) GetAggregateID() string  { return e.AggregateID }
func (e *BaseEvent) GetTimestamp() time.Time { return e.Timestamp }
func (e *BaseEvent) GetVersion() int         { return e.Version }
func (e *BaseEvent) GetPayload() interface{} { return e.Payload }

// 具体事件类型
type UserCreatedEvent struct {
	BaseEvent
	UserData UserData `json:"user_data"`
}

type UserUpdatedEvent struct {
	BaseEvent
	UserData UserData `json:"user_data"`
}

type OrderCreatedEvent struct {
	BaseEvent
	OrderData OrderData `json:"order_data"`
}

type OrderPaidEvent struct {
	BaseEvent
	PaymentData PaymentData `json:"payment_data"`
}

type UserData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

type OrderData struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Amount   float64   `json:"amount"`
	Status   string    `json:"status"`
	Products []Product `json:"products"`
}

type PaymentData struct {
	ID            string  `json:"id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
}

type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// === 消息队列接口 ===

type MessageQueue interface {
	Publish(topic string, message interface{}) error
	Subscribe(topic string, handler MessageHandler) error
	Close() error
}

type MessageHandler func(message *Message) error

type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Payload   interface{}            `json:"payload"`
	Headers   map[string]string      `json:"headers"`
	Timestamp time.Time              `json:"timestamp"`
	Retry     int                    `json:"retry"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// === 内存消息队列实现 ===

type InMemoryMessageQueue struct {
	topics     map[string][]MessageHandler
	deadLetter chan *Message
	mutex      sync.RWMutex
	maxRetries int
	retryDelay time.Duration
}

func NewInMemoryMessageQueue() *InMemoryMessageQueue {
	mq := &InMemoryMessageQueue{
		topics:     make(map[string][]MessageHandler),
		deadLetter: make(chan *Message, 1000),
		maxRetries: 3,
		retryDelay: time.Second * 2,
	}

	// 启动死信队列处理器
	go mq.processDeadLetters()

	return mq
}

func (mq *InMemoryMessageQueue) Publish(topic string, payload interface{}) error {
	message := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Payload:   payload,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
		Retry:     0,
		Metadata:  make(map[string]interface{}),
	}

	mq.mutex.RLock()
	handlers, exists := mq.topics[topic]
	mq.mutex.RUnlock()

	if !exists {
		log.Printf("主题 %s 没有订阅者", topic)
		return nil
	}

	// 异步处理消息
	go func() {
		for _, handler := range handlers {
			mq.processMessage(handler, message)
		}
	}()

	return nil
}

func (mq *InMemoryMessageQueue) Subscribe(topic string, handler MessageHandler) error {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()

	if mq.topics[topic] == nil {
		mq.topics[topic] = make([]MessageHandler, 0)
	}

	mq.topics[topic] = append(mq.topics[topic], handler)
	log.Printf("订阅主题: %s", topic)

	return nil
}

func (mq *InMemoryMessageQueue) processMessage(handler MessageHandler, message *Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("消息处理panic: %v", r)
			mq.retryMessage(handler, message)
		}
	}()

	err := handler(message)
	if err != nil {
		log.Printf("消息处理失败: %v", err)
		mq.retryMessage(handler, message)
	}
}

func (mq *InMemoryMessageQueue) retryMessage(handler MessageHandler, message *Message) {
	message.Retry++

	if message.Retry >= mq.maxRetries {
		log.Printf("消息重试次数超限，发送到死信队列: %s", message.ID)
		select {
		case mq.deadLetter <- message:
		default:
			log.Printf("死信队列已满，丢弃消息: %s", message.ID)
		}
		return
	}

	// 延迟重试
	go func() {
		time.Sleep(mq.retryDelay * time.Duration(message.Retry))
		mq.processMessage(handler, message)
	}()
}

func (mq *InMemoryMessageQueue) processDeadLetters() {
	for message := range mq.deadLetter {
		log.Printf("处理死信消息: %s, 主题: %s, 重试次数: %d",
			message.ID, message.Topic, message.Retry)

		// 这里可以实现死信消息的特殊处理逻辑
		// 比如发送告警、记录日志、人工干预等
	}
}

func (mq *InMemoryMessageQueue) Close() error {
	close(mq.deadLetter)
	return nil
}

// === Kafka消息队列实现 ===

type KafkaMessageQueue struct {
	brokers []string
	writers map[string]*kafka.Writer
	readers map[string]*kafka.Reader
	mutex   sync.RWMutex
}

func NewKafkaMessageQueue(brokers []string) *KafkaMessageQueue {
	return &KafkaMessageQueue{
		brokers: brokers,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}
}

func (kmq *KafkaMessageQueue) Publish(topic string, payload interface{}) error {
	writer := kmq.getWriter(topic)

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	message := kafka.Message{
		Key:   []byte(generateMessageID()),
		Value: data,
		Time:  time.Now(),
	}

	return writer.WriteMessages(context.Background(), message)
}

func (kmq *KafkaMessageQueue) Subscribe(topic string, handler MessageHandler) error {
	reader := kmq.getReader(topic)

	go func() {
		for {
			kafkaMsg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Printf("读取Kafka消息失败: %v", err)
				continue
			}

			message := &Message{
				ID:        string(kafkaMsg.Key),
				Topic:     kafkaMsg.Topic,
				Timestamp: kafkaMsg.Time,
				Headers:   make(map[string]string),
				Metadata:  make(map[string]interface{}),
			}

			// 反序列化payload
			if err := json.Unmarshal(kafkaMsg.Value, &message.Payload); err != nil {
				log.Printf("反序列化消息失败: %v", err)
				continue
			}

			if err := handler(message); err != nil {
				log.Printf("处理消息失败: %v", err)
			}
		}
	}()

	return nil
}

func (kmq *KafkaMessageQueue) getWriter(topic string) *kafka.Writer {
	kmq.mutex.Lock()
	defer kmq.mutex.Unlock()

	writer, exists := kmq.writers[topic]
	if !exists {
		writer = &kafka.Writer{
			Addr:     kafka.TCP(kmq.brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
		kmq.writers[topic] = writer
	}

	return writer
}

func (kmq *KafkaMessageQueue) getReader(topic string) *kafka.Reader {
	kmq.mutex.Lock()
	defer kmq.mutex.Unlock()

	reader, exists := kmq.readers[topic]
	if !exists {
		reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers: kmq.brokers,
			Topic:   topic,
			GroupID: "microservice-group",
		})
		kmq.readers[topic] = reader
	}

	return reader
}

func (kmq *KafkaMessageQueue) Close() error {
	kmq.mutex.Lock()
	defer kmq.mutex.Unlock()

	for _, writer := range kmq.writers {
		writer.Close()
	}

	for _, reader := range kmq.readers {
		reader.Close()
	}

	return nil
}

// === 事件总线 ===

type EventBus struct {
	messageQueue MessageQueue
	eventStore   EventStore
	handlers     map[string][]EventHandler
	mutex        sync.RWMutex
}

type EventHandler func(event Event) error

type EventStore interface {
	SaveEvent(event Event) error
	GetEvents(aggregateID string) ([]Event, error)
	GetEventsByType(eventType string) ([]Event, error)
}

func NewEventBus(messageQueue MessageQueue, eventStore EventStore) *EventBus {
	return &EventBus{
		messageQueue: messageQueue,
		eventStore:   eventStore,
		handlers:     make(map[string][]EventHandler),
	}
}

func (eb *EventBus) PublishEvent(event Event) error {
	// 保存事件到事件存储
	if err := eb.eventStore.SaveEvent(event); err != nil {
		return fmt.Errorf("保存事件失败: %w", err)
	}

	// 发布事件到消息队列
	topic := fmt.Sprintf("events.%s", event.GetEventType())
	return eb.messageQueue.Publish(topic, event)
}

func (eb *EventBus) SubscribeToEvent(eventType string, handler EventHandler) error {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	// 订阅消息队列
	topic := fmt.Sprintf("events.%s", eventType)
	return eb.messageQueue.Subscribe(topic, func(message *Message) error {
		// 将消息转换为事件
		event, err := eb.messageToEvent(message)
		if err != nil {
			return err
		}

		return handler(event)
	})
}

func (eb *EventBus) messageToEvent(message *Message) (Event, error) {
	// 简化实现，实际应该根据事件类型进行反序列化
	eventData, ok := message.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的事件数据")
	}

	baseEvent := &BaseEvent{
		EventID:     eventData["event_id"].(string),
		EventType:   eventData["event_type"].(string),
		AggregateID: eventData["aggregate_id"].(string),
		Timestamp:   time.Now(),
		Version:     int(eventData["version"].(float64)),
		Payload:     eventData["payload"],
	}

	return baseEvent, nil
}

// === 内存事件存储实现 ===

type InMemoryEventStore struct {
	events map[string][]Event // aggregateID -> events
	mutex  sync.RWMutex
}

func NewInMemoryEventStore() *InMemoryEventStore {
	return &InMemoryEventStore{
		events: make(map[string][]Event),
	}
}

func (es *InMemoryEventStore) SaveEvent(event Event) error {
	es.mutex.Lock()
	defer es.mutex.Unlock()

	aggregateID := event.GetAggregateID()
	if es.events[aggregateID] == nil {
		es.events[aggregateID] = make([]Event, 0)
	}

	es.events[aggregateID] = append(es.events[aggregateID], event)
	log.Printf("保存事件: %s, 聚合ID: %s", event.GetEventType(), aggregateID)

	return nil
}

func (es *InMemoryEventStore) GetEvents(aggregateID string) ([]Event, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	events := es.events[aggregateID]
	if events == nil {
		return []Event{}, nil
	}

	return events, nil
}

func (es *InMemoryEventStore) GetEventsByType(eventType string) ([]Event, error) {
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	var result []Event
	for _, events := range es.events {
		for _, event := range events {
			if event.GetEventType() == eventType {
				result = append(result, event)
			}
		}
	}

	return result, nil
}

// === Saga模式实现 ===

type SagaManager struct {
	eventBus  *EventBus
	sagas     map[string]Saga
	instances map[string]*SagaInstance
	mutex     sync.RWMutex
}

type Saga interface {
	GetSagaType() string
	GetSteps() []SagaStep
	CanStart(event Event) bool
}

type SagaStep struct {
	Name            string
	EventType       string
	Handler         SagaStepHandler
	CompensateEvent string
}

type SagaStepHandler func(instance *SagaInstance, event Event) error

type SagaInstance struct {
	ID             string                 `json:"id"`
	SagaType       string                 `json:"saga_type"`
	Status         string                 `json:"status"` // started, completed, failed, compensating
	CurrentStep    int                    `json:"current_step"`
	Data           map[string]interface{} `json:"data"`
	CompletedSteps []string               `json:"completed_steps"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func NewSagaManager(eventBus *EventBus) *SagaManager {
	return &SagaManager{
		eventBus:  eventBus,
		sagas:     make(map[string]Saga),
		instances: make(map[string]*SagaInstance),
	}
}

func (sm *SagaManager) RegisterSaga(saga Saga) {
	sm.sagas[saga.GetSagaType()] = saga

	// 订阅相关事件
	for _, step := range saga.GetSteps() {
		sm.eventBus.SubscribeToEvent(step.EventType, func(event Event) error {
			return sm.handleEvent(event)
		})
	}
}

func (sm *SagaManager) handleEvent(event Event) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 检查是否需要启动新的Saga
	for _, saga := range sm.sagas {
		if saga.CanStart(event) {
			instance := &SagaInstance{
				ID:             generateMessageID(),
				SagaType:       saga.GetSagaType(),
				Status:         "started",
				CurrentStep:    0,
				Data:           make(map[string]interface{}),
				CompletedSteps: make([]string, 0),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			sm.instances[instance.ID] = instance
			log.Printf("启动Saga: %s, 实例ID: %s", saga.GetSagaType(), instance.ID)
		}
	}

	// 处理现有Saga实例
	for _, instance := range sm.instances {
		saga := sm.sagas[instance.SagaType]
		steps := saga.GetSteps()

		if instance.CurrentStep < len(steps) {
			step := steps[instance.CurrentStep]
			if step.EventType == event.GetEventType() {
				if err := step.Handler(instance, event); err != nil {
					log.Printf("Saga步骤失败: %v", err)
					instance.Status = "failed"
					// 这里应该触发补偿流程
				} else {
					instance.CompletedSteps = append(instance.CompletedSteps, step.Name)
					instance.CurrentStep++
					instance.UpdatedAt = time.Now()

					if instance.CurrentStep >= len(steps) {
						instance.Status = "completed"
						log.Printf("Saga完成: %s", instance.ID)
					}
				}
			}
		}
	}

	return nil
}

// === 示例：订单处理Saga ===

type OrderProcessingSaga struct{}

func (ops *OrderProcessingSaga) GetSagaType() string {
	return "OrderProcessing"
}

func (ops *OrderProcessingSaga) GetSteps() []SagaStep {
	return []SagaStep{
		{
			Name:      "ValidateOrder",
			EventType: "OrderCreated",
			Handler:   ops.validateOrder,
		},
		{
			Name:      "ReserveInventory",
			EventType: "OrderValidated",
			Handler:   ops.reserveInventory,
		},
		{
			Name:      "ProcessPayment",
			EventType: "InventoryReserved",
			Handler:   ops.processPayment,
		},
		{
			Name:      "CompleteOrder",
			EventType: "PaymentProcessed",
			Handler:   ops.completeOrder,
		},
	}
}

func (ops *OrderProcessingSaga) CanStart(event Event) bool {
	return event.GetEventType() == "OrderCreated"
}

func (ops *OrderProcessingSaga) validateOrder(instance *SagaInstance, event Event) error {
	log.Printf("验证订单: %s", event.GetAggregateID())
	// 模拟订单验证逻辑
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ops *OrderProcessingSaga) reserveInventory(instance *SagaInstance, event Event) error {
	log.Printf("预留库存: %s", event.GetAggregateID())
	// 模拟库存预留逻辑
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ops *OrderProcessingSaga) processPayment(instance *SagaInstance, event Event) error {
	log.Printf("处理支付: %s", event.GetAggregateID())
	// 模拟支付处理逻辑
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (ops *OrderProcessingSaga) completeOrder(instance *SagaInstance, event Event) error {
	log.Printf("完成订单: %s", event.GetAggregateID())
	return nil
}

// === WebSocket事件流 ===

type EventStreamServer struct {
	eventBus *EventBus
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mutex    sync.Mutex
}

func NewEventStreamServer(eventBus *EventBus) *EventStreamServer {
	return &EventStreamServer{
		eventBus: eventBus,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (ess *EventStreamServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ess.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	ess.mutex.Lock()
	ess.clients[conn] = true
	ess.mutex.Unlock()

	defer func() {
		ess.mutex.Lock()
		delete(ess.clients, conn)
		ess.mutex.Unlock()
	}()

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (ess *EventStreamServer) BroadcastEvent(event Event) {
	message, _ := json.Marshal(map[string]interface{}{
		"type":  "event",
		"event": event,
	})

	ess.mutex.Lock()
	defer ess.mutex.Unlock()

	for conn := range ess.clients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			conn.Close()
			delete(ess.clients, conn)
		}
	}
}

// === 辅助函数 ===

func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

// === HTTP API ===

type EventAPI struct {
	eventBus *EventBus
}

func NewEventAPI(eventBus *EventBus) *EventAPI {
	return &EventAPI{eventBus: eventBus}
}

func (api *EventAPI) PublishUserCreated(w http.ResponseWriter, r *http.Request) {
	var userData UserData
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, "无效的用户数据", http.StatusBadRequest)
		return
	}

	event := &UserCreatedEvent{
		BaseEvent: BaseEvent{
			EventID:     generateEventID(),
			EventType:   "UserCreated",
			AggregateID: userData.ID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		UserData: userData,
	}

	if err := api.eventBus.PublishEvent(event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"event_id": event.EventID})
}

func (api *EventAPI) PublishOrderCreated(w http.ResponseWriter, r *http.Request) {
	var orderData OrderData
	if err := json.NewDecoder(r.Body).Decode(&orderData); err != nil {
		http.Error(w, "无效的订单数据", http.StatusBadRequest)
		return
	}

	event := &OrderCreatedEvent{
		BaseEvent: BaseEvent{
			EventID:     generateEventID(),
			EventType:   "OrderCreated",
			AggregateID: orderData.ID,
			Timestamp:   time.Now(),
			Version:     1,
		},
		OrderData: orderData,
	}

	if err := api.eventBus.PublishEvent(event); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"event_id": event.EventID})
}

func main() {
	// 创建消息队列
	messageQueue := NewInMemoryMessageQueue()
	defer messageQueue.Close()

	// 创建事件存储
	eventStore := NewInMemoryEventStore()

	// 创建事件总线
	eventBus := NewEventBus(messageQueue, eventStore)

	// 创建Saga管理器
	sagaManager := NewSagaManager(eventBus)

	// 注册Saga
	orderSaga := &OrderProcessingSaga{}
	sagaManager.RegisterSaga(orderSaga)

	// 创建事件流服务器
	eventStreamServer := NewEventStreamServer(eventBus)

	// 订阅事件处理器
	eventBus.SubscribeToEvent("UserCreated", func(event Event) error {
		log.Printf("处理用户创建事件: %s", event.GetAggregateID())
		eventStreamServer.BroadcastEvent(event)
		return nil
	})

	eventBus.SubscribeToEvent("OrderCreated", func(event Event) error {
		log.Printf("处理订单创建事件: %s", event.GetAggregateID())
		eventStreamServer.BroadcastEvent(event)
		return nil
	})

	// 创建API
	eventAPI := NewEventAPI(eventBus)

	// 创建路由器
	router := mux.NewRouter()

	// 事件发布API
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/events/user-created", eventAPI.PublishUserCreated).Methods("POST")
	api.HandleFunc("/events/order-created", eventAPI.PublishOrderCreated).Methods("POST")

	// WebSocket事件流
	router.HandleFunc("/events/stream", eventStreamServer.HandleWebSocket)

	// 事件查询API
	router.HandleFunc("/api/events/{aggregateId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		aggregateID := vars["aggregateId"]

		events, err := eventStore.GetEvents(aggregateID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}).Methods("GET")

	// 演示事件发布
	go func() {
		time.Sleep(3 * time.Second)

		// 发布用户创建事件
		userData := UserData{
			ID:       "user_1",
			Username: "alice",
			Email:    "alice@example.com",
			Status:   "active",
		}

		userEvent := &UserCreatedEvent{
			BaseEvent: BaseEvent{
				EventID:     generateEventID(),
				EventType:   "UserCreated",
				AggregateID: userData.ID,
				Timestamp:   time.Now(),
				Version:     1,
			},
			UserData: userData,
		}

		eventBus.PublishEvent(userEvent)

		time.Sleep(2 * time.Second)

		// 发布订单创建事件
		orderData := OrderData{
			ID:     "order_1",
			UserID: "user_1",
			Amount: 99.99,
			Status: "pending",
			Products: []Product{
				{ID: "prod_1", Name: "Laptop", Price: 99.99, Quantity: 1},
			},
		}

		orderEvent := &OrderCreatedEvent{
			BaseEvent: BaseEvent{
				EventID:     generateEventID(),
				EventType:   "OrderCreated",
				AggregateID: orderData.ID,
				Timestamp:   time.Now(),
				Version:     1,
			},
			OrderData: orderData,
		}

		eventBus.PublishEvent(orderEvent)
	}()

	fmt.Println("=== 消息队列与事件驱动系统启动 ===")
	fmt.Println("服务端点:")
	fmt.Println("  事件API:    http://localhost:8080")
	fmt.Println("  事件流:     ws://localhost:8080/events/stream")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  POST /api/events/user-created   - 发布用户创建事件")
	fmt.Println("  POST /api/events/order-created  - 发布订单创建事件")
	fmt.Println("  GET  /api/events/{aggregateId}  - 查询聚合事件")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  # 发布用户创建事件")
	fmt.Println("  curl -X POST http://localhost:8080/api/events/user-created \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"id\":\"user_2\",\"username\":\"bob\",\"email\":\"bob@example.com\",\"status\":\"active\"}'")
	fmt.Println()
	fmt.Println("  # 发布订单创建事件")
	fmt.Println("  curl -X POST http://localhost:8080/api/events/order-created \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"id\":\"order_2\",\"user_id\":\"user_2\",\"amount\":149.99,\"status\":\"pending\",\"products\":[{\"id\":\"prod_2\",\"name\":\"Phone\",\"price\":149.99,\"quantity\":1}]}'")

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
练习任务：

1. 基础练习：
   - 实现更多事件类型和处理器
   - 添加事件版本控制
   - 实现事件快照机制
   - 添加事件重放功能

2. 中级练习：
   - 集成RabbitMQ或Apache Kafka
   - 实现事件序列化/反序列化
   - 添加消息压缩和批处理
   - 实现事件流分片

3. 高级练习：
   - 实现分布式事件溯源
   - 添加事件存储压缩
   - 实现跨服务事件传播
   - 集成分布式锁机制

4. Saga模式练习：
   - 实现更复杂的业务流程
   - 添加补偿机制
   - 实现Saga状态持久化
   - 添加并行步骤支持

5. 监控和运维：
   - 实现事件流监控
   - 添加消息积压告警
   - 实现性能指标收集
   - 添加事件审计功能

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/gorilla/websocket
   go get github.com/segmentio/kafka-go

2. 可选：启动Kafka
   docker run -d --name kafka -p 9092:9092 confluentinc/cp-kafka

3. 运行程序：go run main.go

事件驱动架构图：
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   服务A     │────│  事件总线    │────│   服务B     │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       │            ┌─────────────┐            │
       │────────────│  消息队列    │────────────│
       │            └─────────────┘            │
       │                   │                   │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  事件存储    │    │ Saga管理器   │    │   服务C     │
└─────────────┘    └─────────────┘    └─────────────┘

扩展建议：
- 实现事件驱动微服务网格
- 集成分布式追踪系统
- 添加事件治理和策略
- 实现事件驱动的CQRS模式
*/
