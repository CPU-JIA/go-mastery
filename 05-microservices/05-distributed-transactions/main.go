package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

/*
微服务架构 - 分布式事务处理练习

本练习涵盖微服务架构中的分布式事务处理，包括：
1. 两阶段提交（2PC）
2. 三阶段提交（3PC）
3. Saga模式
4. TCC模式（Try-Confirm-Cancel）
5. 分布式锁
6. 最终一致性
7. 补偿机制
8. 事务协调器

主要概念：
- ACID属性在分布式系统中的挑战
- CAP定理和BASE理论
- 分布式一致性算法
- 事务补偿模式
- 幂等性设计
*/

// === 分布式事务接口定义 ===

// TransactionManager 分布式事务管理器接口
type TransactionManager interface {
	BeginTransaction(ctx context.Context) (*Transaction, error)
	CommitTransaction(ctx context.Context, txID string) error
	AbortTransaction(ctx context.Context, txID string) error
	GetTransaction(txID string) (*Transaction, error)
}

// TransactionParticipant 事务参与者接口
type TransactionParticipant interface {
	Prepare(ctx context.Context, txID string) (bool, error)
	Commit(ctx context.Context, txID string) error
	Abort(ctx context.Context, txID string) error
	GetParticipantID() string
}

// Transaction 分布式事务
type Transaction struct {
	ID           string                   `json:"id"`
	Status       TransactionStatus        `json:"status"`
	Participants []TransactionParticipant `json:"-"`
	Operations   []TransactionOperation   `json:"operations"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
	Timeout      time.Duration            `json:"timeout"`
	Context      map[string]interface{}   `json:"context"`
}

// TransactionStatus 事务状态
type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusPreparing TransactionStatus = "preparing"
	StatusPrepared  TransactionStatus = "prepared"
	StatusCommitted TransactionStatus = "committed"
	StatusAborted   TransactionStatus = "aborted"
	StatusTimeout   TransactionStatus = "timeout"
)

// TransactionOperation 事务操作
type TransactionOperation struct {
	ID            string                 `json:"id"`
	ParticipantID string                 `json:"participant_id"`
	Type          string                 `json:"type"`
	Data          map[string]interface{} `json:"data"`
	Status        string                 `json:"status"`
	CreatedAt     time.Time              `json:"created_at"`
}

// === 两阶段提交（2PC）实现 ===

type TwoPhaseCommitManager struct {
	transactions map[string]*Transaction
	mutex        sync.RWMutex
	timeout      time.Duration
}

func NewTwoPhaseCommitManager(timeout time.Duration) *TwoPhaseCommitManager {
	return &TwoPhaseCommitManager{
		transactions: make(map[string]*Transaction),
		timeout:      timeout,
	}
}

func (tpc *TwoPhaseCommitManager) BeginTransaction(ctx context.Context) (*Transaction, error) {
	tx := &Transaction{
		ID:           uuid.New().String(),
		Status:       StatusPending,
		Participants: make([]TransactionParticipant, 0),
		Operations:   make([]TransactionOperation, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Timeout:      tpc.timeout,
		Context:      make(map[string]interface{}),
	}

	tpc.mutex.Lock()
	tpc.transactions[tx.ID] = tx
	tpc.mutex.Unlock()

	log.Printf("开始分布式事务: %s", tx.ID)

	// 启动超时处理
	go tpc.handleTimeout(tx.ID)

	return tx, nil
}

func (tpc *TwoPhaseCommitManager) CommitTransaction(ctx context.Context, txID string) error {
	tpc.mutex.Lock()
	tx, exists := tpc.transactions[txID]
	if !exists {
		tpc.mutex.Unlock()
		return fmt.Errorf("事务不存在: %s", txID)
	}

	if tx.Status != StatusPending {
		tpc.mutex.Unlock()
		return fmt.Errorf("事务状态无效: %s", tx.Status)
	}

	tx.Status = StatusPreparing
	tx.UpdatedAt = time.Now()
	tpc.mutex.Unlock()

	log.Printf("开始两阶段提交 - 阶段1（准备）: %s", txID)

	// 阶段1：准备阶段
	prepareChan := make(chan bool, len(tx.Participants))
	errorChan := make(chan error, len(tx.Participants))

	for _, participant := range tx.Participants {
		go func(p TransactionParticipant) {
			prepared, err := p.Prepare(ctx, txID)
			if err != nil {
				errorChan <- err
				return
			}
			prepareChan <- prepared
		}(participant)
	}

	preparedCount := 0
	allPrepared := true

	for i := 0; i < len(tx.Participants); i++ {
		select {
		case prepared := <-prepareChan:
			if prepared {
				preparedCount++
			} else {
				allPrepared = false
			}
		case err := <-errorChan:
			log.Printf("准备阶段失败: %v", err)
			allPrepared = false
		case <-time.After(30 * time.Second):
			log.Printf("准备阶段超时: %s", txID)
			allPrepared = false
		}
	}

	if !allPrepared || preparedCount != len(tx.Participants) {
		log.Printf("准备阶段失败，执行回滚: %s", txID)
		return tpc.AbortTransaction(ctx, txID)
	}

	tpc.mutex.Lock()
	tx.Status = StatusPrepared
	tx.UpdatedAt = time.Now()
	tpc.mutex.Unlock()

	log.Printf("开始两阶段提交 - 阶段2（提交）: %s", txID)

	// 阶段2：提交阶段
	commitChan := make(chan error, len(tx.Participants))

	for _, participant := range tx.Participants {
		go func(p TransactionParticipant) {
			err := p.Commit(ctx, txID)
			commitChan <- err
		}(participant)
	}

	commitErrors := make([]error, 0)
	for i := 0; i < len(tx.Participants); i++ {
		if err := <-commitChan; err != nil {
			commitErrors = append(commitErrors, err)
		}
	}

	if len(commitErrors) > 0 {
		tpc.mutex.Lock()
		tx.Status = StatusAborted
		tx.UpdatedAt = time.Now()
		tpc.mutex.Unlock()

		log.Printf("提交阶段失败: %v", commitErrors)
		return fmt.Errorf("提交阶段失败: %v", commitErrors)
	}

	tpc.mutex.Lock()
	tx.Status = StatusCommitted
	tx.UpdatedAt = time.Now()
	tpc.mutex.Unlock()

	log.Printf("分布式事务提交成功: %s", txID)
	return nil
}

func (tpc *TwoPhaseCommitManager) AbortTransaction(ctx context.Context, txID string) error {
	tpc.mutex.Lock()
	tx, exists := tpc.transactions[txID]
	if !exists {
		tpc.mutex.Unlock()
		return fmt.Errorf("事务不存在: %s", txID)
	}

	tx.Status = StatusAborted
	tx.UpdatedAt = time.Now()
	tpc.mutex.Unlock()

	log.Printf("开始回滚事务: %s", txID)

	// 并行执行回滚
	abortChan := make(chan error, len(tx.Participants))

	for _, participant := range tx.Participants {
		go func(p TransactionParticipant) {
			err := p.Abort(ctx, txID)
			abortChan <- err
		}(participant)
	}

	abortErrors := make([]error, 0)
	for i := 0; i < len(tx.Participants); i++ {
		if err := <-abortChan; err != nil {
			abortErrors = append(abortErrors, err)
		}
	}

	if len(abortErrors) > 0 {
		log.Printf("回滚过程中发生错误: %v", abortErrors)
	} else {
		log.Printf("事务回滚成功: %s", txID)
	}

	return nil
}

func (tpc *TwoPhaseCommitManager) GetTransaction(txID string) (*Transaction, error) {
	tpc.mutex.RLock()
	defer tpc.mutex.RUnlock()

	tx, exists := tpc.transactions[txID]
	if !exists {
		return nil, fmt.Errorf("事务不存在: %s", txID)
	}

	return tx, nil
}

func (tpc *TwoPhaseCommitManager) AddParticipant(txID string, participant TransactionParticipant) error {
	tpc.mutex.Lock()
	defer tpc.mutex.Unlock()

	tx, exists := tpc.transactions[txID]
	if !exists {
		return fmt.Errorf("事务不存在: %s", txID)
	}

	if tx.Status != StatusPending {
		return fmt.Errorf("无法向已开始的事务添加参与者")
	}

	tx.Participants = append(tx.Participants, participant)
	log.Printf("添加事务参与者: %s -> %s", txID, participant.GetParticipantID())

	return nil
}

func (tpc *TwoPhaseCommitManager) handleTimeout(txID string) {
	time.Sleep(tpc.timeout)

	tpc.mutex.Lock()
	tx, exists := tpc.transactions[txID]
	if !exists {
		tpc.mutex.Unlock()
		return
	}

	if tx.Status == StatusPending || tx.Status == StatusPreparing {
		tx.Status = StatusTimeout
		tx.UpdatedAt = time.Now()
		tpc.mutex.Unlock()

		log.Printf("事务超时，执行回滚: %s", txID)
		tpc.AbortTransaction(context.Background(), txID)
	} else {
		tpc.mutex.Unlock()
	}
}

// === TCC模式实现 ===

// TCCParticipant TCC参与者接口
type TCCParticipant interface {
	Try(ctx context.Context, txID string, params map[string]interface{}) error
	Confirm(ctx context.Context, txID string) error
	Cancel(ctx context.Context, txID string) error
	GetParticipantID() string
}

// TCCManager TCC事务管理器
type TCCManager struct {
	transactions map[string]*TCCTransaction
	mutex        sync.RWMutex
}

type TCCTransaction struct {
	ID           string                    `json:"id"`
	Status       TransactionStatus         `json:"status"`
	Participants map[string]TCCParticipant `json:"-"`
	Operations   []TCCOperation            `json:"operations"`
	CreatedAt    time.Time                 `json:"created_at"`
	UpdatedAt    time.Time                 `json:"updated_at"`
}

type TCCOperation struct {
	ID            string                 `json:"id"`
	ParticipantID string                 `json:"participant_id"`
	Phase         string                 `json:"phase"` // try, confirm, cancel
	Status        string                 `json:"status"`
	Params        map[string]interface{} `json:"params"`
	CreatedAt     time.Time              `json:"created_at"`
}

func NewTCCManager() *TCCManager {
	return &TCCManager{
		transactions: make(map[string]*TCCTransaction),
	}
}

func (tcc *TCCManager) BeginTCCTransaction(ctx context.Context) (*TCCTransaction, error) {
	tx := &TCCTransaction{
		ID:           uuid.New().String(),
		Status:       StatusPending,
		Participants: make(map[string]TCCParticipant),
		Operations:   make([]TCCOperation, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	tcc.mutex.Lock()
	tcc.transactions[tx.ID] = tx
	tcc.mutex.Unlock()

	log.Printf("开始TCC事务: %s", tx.ID)
	return tx, nil
}

func (tcc *TCCManager) AddTCCParticipant(txID string, participant TCCParticipant) error {
	tcc.mutex.Lock()
	defer tcc.mutex.Unlock()

	tx, exists := tcc.transactions[txID]
	if !exists {
		return fmt.Errorf("TCC事务不存在: %s", txID)
	}

	tx.Participants[participant.GetParticipantID()] = participant
	return nil
}

func (tcc *TCCManager) ExecuteTCCTransaction(ctx context.Context, txID string, operations map[string]map[string]interface{}) error {
	tcc.mutex.Lock()
	tx, exists := tcc.transactions[txID]
	if !exists {
		tcc.mutex.Unlock()
		return fmt.Errorf("TCC事务不存在: %s", txID)
	}

	tx.Status = StatusPreparing
	tx.UpdatedAt = time.Now()
	tcc.mutex.Unlock()

	log.Printf("开始TCC Try阶段: %s", txID)

	// Try阶段
	tryErrors := make(map[string]error)
	for participantID, params := range operations {
		participant, exists := tx.Participants[participantID]
		if !exists {
			tryErrors[participantID] = fmt.Errorf("参与者不存在: %s", participantID)
			continue
		}

		operation := TCCOperation{
			ID:            uuid.New().String(),
			ParticipantID: participantID,
			Phase:         "try",
			Status:        "executing",
			Params:        params,
			CreatedAt:     time.Now(),
		}

		tx.Operations = append(tx.Operations, operation)

		if err := participant.Try(ctx, txID, params); err != nil {
			tryErrors[participantID] = err
			operation.Status = "failed"
		} else {
			operation.Status = "success"
		}
	}

	if len(tryErrors) > 0 {
		log.Printf("TCC Try阶段失败，执行Cancel: %v", tryErrors)
		return tcc.cancelTCCTransaction(ctx, txID)
	}

	log.Printf("开始TCC Confirm阶段: %s", txID)

	// Confirm阶段
	confirmErrors := make(map[string]error)
	for participantID := range operations {
		participant := tx.Participants[participantID]

		operation := TCCOperation{
			ID:            uuid.New().String(),
			ParticipantID: participantID,
			Phase:         "confirm",
			Status:        "executing",
			CreatedAt:     time.Now(),
		}

		tx.Operations = append(tx.Operations, operation)

		if err := participant.Confirm(ctx, txID); err != nil {
			confirmErrors[participantID] = err
			operation.Status = "failed"
		} else {
			operation.Status = "success"
		}
	}

	if len(confirmErrors) > 0 {
		log.Printf("TCC Confirm阶段失败: %v", confirmErrors)
		tcc.mutex.Lock()
		tx.Status = StatusAborted
		tx.UpdatedAt = time.Now()
		tcc.mutex.Unlock()
		return fmt.Errorf("confirm阶段失败: %v", confirmErrors)
	}

	tcc.mutex.Lock()
	tx.Status = StatusCommitted
	tx.UpdatedAt = time.Now()
	tcc.mutex.Unlock()

	log.Printf("TCC事务提交成功: %s", txID)
	return nil
}

func (tcc *TCCManager) cancelTCCTransaction(ctx context.Context, txID string) error {
	tcc.mutex.Lock()
	tx, exists := tcc.transactions[txID]
	if !exists {
		tcc.mutex.Unlock()
		return fmt.Errorf("TCC事务不存在: %s", txID)
	}

	tx.Status = StatusAborted
	tx.UpdatedAt = time.Now()
	tcc.mutex.Unlock()

	log.Printf("开始TCC Cancel阶段: %s", txID)

	// Cancel阶段
	for participantID := range tx.Participants {
		participant := tx.Participants[participantID]

		operation := TCCOperation{
			ID:            uuid.New().String(),
			ParticipantID: participantID,
			Phase:         "cancel",
			Status:        "executing",
			CreatedAt:     time.Now(),
		}

		tx.Operations = append(tx.Operations, operation)

		if err := participant.Cancel(ctx, txID); err != nil {
			log.Printf("Cancel失败 %s: %v", participantID, err)
			operation.Status = "failed"
		} else {
			operation.Status = "success"
		}
	}

	log.Printf("TCC事务取消完成: %s", txID)
	return nil
}

// === 示例业务参与者 ===

// AccountService 账户服务
type AccountService struct {
	accounts map[string]*Account
	reserves map[string]float64 // 预留资金
	mutex    sync.RWMutex
}

type Account struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

func NewAccountService() *AccountService {
	service := &AccountService{
		accounts: make(map[string]*Account),
		reserves: make(map[string]float64),
	}

	// 初始化测试账户
	service.accounts["acc_1"] = &Account{ID: "acc_1", Balance: 1000.0, Status: "active"}
	service.accounts["acc_2"] = &Account{ID: "acc_2", Balance: 500.0, Status: "active"}

	return service
}

func (as *AccountService) GetParticipantID() string {
	return "account-service"
}

// 2PC参与者实现
func (as *AccountService) Prepare(ctx context.Context, txID string) (bool, error) {
	log.Printf("账户服务准备事务: %s", txID)
	// 模拟准备逻辑，比如锁定资源
	time.Sleep(100 * time.Millisecond)
	return true, nil
}

func (as *AccountService) Commit(ctx context.Context, txID string) error {
	log.Printf("账户服务提交事务: %s", txID)
	// 模拟提交逻辑
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (as *AccountService) Abort(ctx context.Context, txID string) error {
	log.Printf("账户服务回滚事务: %s", txID)
	// 模拟回滚逻辑
	time.Sleep(100 * time.Millisecond)
	return nil
}

// TCC参与者实现
func (as *AccountService) Try(ctx context.Context, txID string, params map[string]interface{}) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	accountID := params["account_id"].(string)
	amount := params["amount"].(float64)

	account, exists := as.accounts[accountID]
	if !exists {
		return fmt.Errorf("账户不存在: %s", accountID)
	}

	if account.Balance < amount {
		return fmt.Errorf("余额不足: %f < %f", account.Balance, amount)
	}

	// 预留资金
	reserveKey := fmt.Sprintf("%s_%s", txID, accountID)
	as.reserves[reserveKey] = amount
	account.Balance -= amount

	log.Printf("账户 %s 预留资金: %f", accountID, amount)
	return nil
}

func (as *AccountService) Confirm(ctx context.Context, txID string) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	// 确认预留的资金扣除
	for key := range as.reserves {
		if strings.HasPrefix(key, txID+"_") {
			delete(as.reserves, key)
		}
	}

	log.Printf("账户服务确认事务: %s", txID)
	return nil
}

func (as *AccountService) Cancel(ctx context.Context, txID string) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	// 恢复预留的资金
	for key, amount := range as.reserves {
		if strings.HasPrefix(key, txID+"_") {
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				accountID := parts[1]
				if account, exists := as.accounts[accountID]; exists {
					account.Balance += amount
				}
			}
			delete(as.reserves, key)
		}
	}

	log.Printf("账户服务取消事务: %s", txID)
	return nil
}

// 库存服务
type InventoryService struct {
	inventory map[string]*InventoryItem
	reserves  map[string]int // 预留库存
	mutex     sync.RWMutex
}

type InventoryItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Reserved  int    `json:"reserved"`
}

func NewInventoryService() *InventoryService {
	service := &InventoryService{
		inventory: make(map[string]*InventoryItem),
		reserves:  make(map[string]int),
	}

	// 初始化测试库存
	service.inventory["prod_1"] = &InventoryItem{ProductID: "prod_1", Quantity: 100, Reserved: 0}
	service.inventory["prod_2"] = &InventoryItem{ProductID: "prod_2", Quantity: 50, Reserved: 0}

	return service
}

func (is *InventoryService) GetParticipantID() string {
	return "inventory-service"
}

// 2PC参与者实现
func (is *InventoryService) Prepare(ctx context.Context, txID string) (bool, error) {
	log.Printf("库存服务准备事务: %s", txID)
	time.Sleep(100 * time.Millisecond)
	return true, nil
}

func (is *InventoryService) Commit(ctx context.Context, txID string) error {
	log.Printf("库存服务提交事务: %s", txID)
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (is *InventoryService) Abort(ctx context.Context, txID string) error {
	log.Printf("库存服务回滚事务: %s", txID)
	time.Sleep(100 * time.Millisecond)
	return nil
}

// TCC参与者实现
func (is *InventoryService) Try(ctx context.Context, txID string, params map[string]interface{}) error {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	productID := params["product_id"].(string)
	quantity := int(params["quantity"].(float64))

	item, exists := is.inventory[productID]
	if !exists {
		return fmt.Errorf("商品不存在: %s", productID)
	}

	if item.Quantity < quantity {
		return fmt.Errorf("库存不足: %d < %d", item.Quantity, quantity)
	}

	// 预留库存
	reserveKey := fmt.Sprintf("%s_%s", txID, productID)
	is.reserves[reserveKey] = quantity
	item.Quantity -= quantity
	item.Reserved += quantity

	log.Printf("商品 %s 预留库存: %d", productID, quantity)
	return nil
}

func (is *InventoryService) Confirm(ctx context.Context, txID string) error {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	// 确认库存扣减
	for key, quantity := range is.reserves {
		if strings.HasPrefix(key, txID+"_") {
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				productID := parts[1]
				if item, exists := is.inventory[productID]; exists {
					item.Reserved -= quantity
				}
			}
			delete(is.reserves, key)
		}
	}

	log.Printf("库存服务确认事务: %s", txID)
	return nil
}

func (is *InventoryService) Cancel(ctx context.Context, txID string) error {
	is.mutex.Lock()
	defer is.mutex.Unlock()

	// 恢复预留库存
	for key, quantity := range is.reserves {
		if strings.HasPrefix(key, txID+"_") {
			parts := strings.Split(key, "_")
			if len(parts) >= 2 {
				productID := parts[1]
				if item, exists := is.inventory[productID]; exists {
					item.Quantity += quantity
					item.Reserved -= quantity
				}
			}
			delete(is.reserves, key)
		}
	}

	log.Printf("库存服务取消事务: %s", txID)
	return nil
}

// === HTTP API ===

type TransactionAPI struct {
	tpcManager   *TwoPhaseCommitManager
	tccManager   *TCCManager
	accountSvc   *AccountService
	inventorySvc *InventoryService
}

func NewTransactionAPI() *TransactionAPI {
	return &TransactionAPI{
		tpcManager:   NewTwoPhaseCommitManager(5 * time.Minute),
		tccManager:   NewTCCManager(),
		accountSvc:   NewAccountService(),
		inventorySvc: NewInventoryService(),
	}
}

func (api *TransactionAPI) Create2PCTransaction(w http.ResponseWriter, r *http.Request) {
	tx, err := api.tpcManager.BeginTransaction(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 添加参与者
	api.tpcManager.AddParticipant(tx.ID, api.accountSvc)
	api.tpcManager.AddParticipant(tx.ID, api.inventorySvc)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (api *TransactionAPI) Commit2PCTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txID := vars["txId"]

	err := api.tpcManager.CommitTransaction(r.Context(), txID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "committed"})
}

func (api *TransactionAPI) CreateTCCTransaction(w http.ResponseWriter, r *http.Request) {
	tx, err := api.tccManager.BeginTCCTransaction(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 添加参与者
	api.tccManager.AddTCCParticipant(tx.ID, api.accountSvc)
	api.tccManager.AddTCCParticipant(tx.ID, api.inventorySvc)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (api *TransactionAPI) ExecuteTCCTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txID := vars["txId"]

	var req struct {
		Operations map[string]map[string]interface{} `json:"operations"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	err := api.tccManager.ExecuteTCCTransaction(r.Context(), txID, req.Operations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "executed"})
}

func (api *TransactionAPI) GetAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.accountSvc.accounts)
}

func (api *TransactionAPI) GetInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.inventorySvc.inventory)
}

func main() {
	api := NewTransactionAPI()

	router := mux.NewRouter()

	// 2PC API
	tpc := router.PathPrefix("/api/2pc").Subrouter()
	tpc.HandleFunc("/transactions", api.Create2PCTransaction).Methods("POST")
	tpc.HandleFunc("/transactions/{txId}/commit", api.Commit2PCTransaction).Methods("POST")

	// TCC API
	tcc := router.PathPrefix("/api/tcc").Subrouter()
	tcc.HandleFunc("/transactions", api.CreateTCCTransaction).Methods("POST")
	tcc.HandleFunc("/transactions/{txId}/execute", api.ExecuteTCCTransaction).Methods("POST")

	// 业务状态查询
	router.HandleFunc("/api/accounts", api.GetAccounts).Methods("GET")
	router.HandleFunc("/api/inventory", api.GetInventory).Methods("GET")

	// 演示分布式事务
	go func() {
		time.Sleep(3 * time.Second)

		log.Println("=== 演示TCC事务 ===")

		// 创建TCC事务
		tx, _ := api.tccManager.BeginTCCTransaction(context.Background())
		api.tccManager.AddTCCParticipant(tx.ID, api.accountSvc)
		api.tccManager.AddTCCParticipant(tx.ID, api.inventorySvc)

		// 执行业务操作
		operations := map[string]map[string]interface{}{
			"account-service": {
				"account_id": "acc_1",
				"amount":     100.0,
			},
			"inventory-service": {
				"product_id": "prod_1",
				"quantity":   5.0,
			},
		}

		err := api.tccManager.ExecuteTCCTransaction(context.Background(), tx.ID, operations)
		if err != nil {
			log.Printf("TCC事务执行失败: %v", err)
		} else {
			log.Printf("TCC事务执行成功")
		}
	}()

	fmt.Println("=== 分布式事务处理系统启动 ===")
	fmt.Println("服务端点:")
	fmt.Println("  事务API:    http://localhost:8080")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  POST /api/2pc/transactions          - 创建2PC事务")
	fmt.Println("  POST /api/2pc/transactions/{id}/commit - 提交2PC事务")
	fmt.Println("  POST /api/tcc/transactions          - 创建TCC事务")
	fmt.Println("  POST /api/tcc/transactions/{id}/execute - 执行TCC事务")
	fmt.Println("  GET  /api/accounts                  - 查看账户状态")
	fmt.Println("  GET  /api/inventory                 - 查看库存状态")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  # 创建TCC事务")
	fmt.Println("  curl -X POST http://localhost:8080/api/tcc/transactions")
	fmt.Println()
	fmt.Println("  # 执行TCC事务")
	fmt.Println("  curl -X POST http://localhost:8080/api/tcc/transactions/{tx-id}/execute \\")
	fmt.Println("    -H 'Content-Type: application/json' \\")
	fmt.Println("    -d '{\"operations\":{\"account-service\":{\"account_id\":\"acc_1\",\"amount\":50},\"inventory-service\":{\"product_id\":\"prod_1\",\"quantity\":3}}}'")

	log.Fatal(http.ListenAndServe(":8080", router))
}

/*
练习任务：

1. 基础练习：
   - 实现三阶段提交（3PC）
   - 添加分布式锁机制
   - 实现事务日志记录
   - 添加事务恢复机制

2. 中级练习：
   - 实现Saga模式的完整实现
   - 添加补偿事务链
   - 实现事务超时和重试
   - 添加事务监控

3. 高级练习：
   - 实现基于Raft的分布式一致性
   - 添加事务隔离级别
   - 实现读写分离的事务处理
   - 集成分布式锁服务

4. 性能优化：
   - 实现事务并发控制
   - 添加事务批处理
   - 优化网络通信
   - 实现事务缓存

5. 容错和恢复：
   - 实现故障检测和自动恢复
   - 添加数据一致性检查
   - 实现分区容错
   - 添加灾难恢复机制

事务模式对比：

2PC（两阶段提交）：
- 优点：保证强一致性
- 缺点：阻塞协议，单点故障

3PC（三阶段提交）：
- 优点：减少阻塞时间
- 缺点：实现复杂，网络分区问题

Saga模式：
- 优点：长事务支持，无阻塞
- 缺点：最终一致性，补偿复杂

TCC模式：
- 优点：业务无侵入，一致性保证
- 缺点：需要实现三个接口

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/google/uuid

2. 运行程序：go run main.go

分布式事务架构：
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ 事务协调器   │────│  参与者A     │    │  参与者B     │
│ (TM)        │    │ 账户服务     │    │ 库存服务     │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
   ┌─────────┐         ┌─────────┐        ┌─────────┐
   │事务日志  │         │ 本地资源 │        │ 本地资源 │
   └─────────┘         └─────────┘        └─────────┘

扩展建议：
- 集成消息队列进行异步处理
- 实现基于时间戳的多版本并发控制
- 添加分布式事务监控和可视化
- 集成微服务框架（如Spring Cloud）
*/
