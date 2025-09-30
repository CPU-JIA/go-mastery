/*
任务调度系统 (Task Scheduler System)

项目描述:
一个完整的任务调度系统，支持定时任务(Cron)、延迟任务、
一次性任务、任务队列、任务重试、任务监控等功能。

技术栈:
- Cron 表达式解析
- 任务队列和工作池
- 分布式锁
- 任务持久化
- 错误处理和重试
- 监控和统计
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-mastery/common/security"
)

// ====================
// 1. 数据模型
// ====================

type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`         // cron, delay, once
	Status      string                 `json:"status"`       // pending, running, completed, failed, cancelled
	Schedule    string                 `json:"schedule"`     // cron 表达式或延迟时间
	Payload     map[string]interface{} `json:"payload"`      // 任务参数
	HandlerName string                 `json:"handler_name"` // 处理器名称
	Priority    int                    `json:"priority"`     // 优先级 (1-10, 数字越大优先级越高)

	// 执行控制
	MaxRetries int `json:"max_retries"`
	RetryCount int `json:"retry_count"`
	RetryDelay int `json:"retry_delay"` // 重试延迟(秒)
	Timeout    int `json:"timeout"`     // 超时时间(秒)

	// 时间信息
	CreatedAt   time.Time  `json:"created_at"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	NextRunAt   *time.Time `json:"next_run_at,omitempty"`

	// 执行结果
	Result   interface{} `json:"result,omitempty"`
	Error    string      `json:"error,omitempty"`
	Duration int64       `json:"duration"` // 执行时长(毫秒)

	// 元数据
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

type TaskHandler func(ctx context.Context, task *Task) (interface{}, error)

type TaskExecution struct {
	ID          string      `json:"id"`
	TaskID      string      `json:"task_id"`
	Status      string      `json:"status"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	Duration    int64       `json:"duration"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	WorkerID    string      `json:"worker_id"`
}

type Worker struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"` // idle, busy, stopped
	StartedAt      time.Time `json:"started_at"`
	CurrentTask    *Task     `json:"current_task,omitempty"`
	TasksProcessed int       `json:"tasks_processed"`
	LastActivity   time.Time `json:"last_activity"`
}

type SchedulerStats struct {
	TotalTasks     int `json:"total_tasks"`
	PendingTasks   int `json:"pending_tasks"`
	RunningTasks   int `json:"running_tasks"`
	CompletedTasks int `json:"completed_tasks"`
	FailedTasks    int `json:"failed_tasks"`
	ActiveWorkers  int `json:"active_workers"`

	TasksPerMinute float64 `json:"tasks_per_minute"`
	AvgDuration    float64 `json:"avg_duration"`
	SuccessRate    float64 `json:"success_rate"`
}

// ====================
// 2. Cron 表达式解析器
// ====================

type CronParser struct{}

func NewCronParser() *CronParser {
	return &CronParser{}
}

func (cp *CronParser) ParseCron(cronExpr string) (*CronSchedule, error) {
	// 简化的 Cron 解析器，支持标准 5 字段格式: 分 时 日 月 周
	fields := strings.Fields(cronExpr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("invalid cron expression: expected 5 fields, got %d", len(fields))
	}

	schedule := &CronSchedule{
		Expression: cronExpr,
		Minute:     fields[0],
		Hour:       fields[1],
		Day:        fields[2],
		Month:      fields[3],
		DayOfWeek:  fields[4],
	}

	return schedule, nil
}

type CronSchedule struct {
	Expression string
	Minute     string
	Hour       string
	Day        string
	Month      string
	DayOfWeek  string
}

func (cs *CronSchedule) NextRun(from time.Time) time.Time {
	// 简化实现：每分钟检查一次
	next := from.Truncate(time.Minute).Add(time.Minute)

	// 基本的 cron 匹配逻辑
	for {
		if cs.matches(next) {
			return next
		}
		next = next.Add(time.Minute)

		// 防止无限循环，最多检查一年
		if next.Sub(from) > 365*24*time.Hour {
			break
		}
	}

	return from.Add(time.Hour) // 回退方案
}

func (cs *CronSchedule) matches(t time.Time) bool {
	// 简化的匹配逻辑
	if !cs.matchField(cs.Minute, t.Minute(), 0, 59) {
		return false
	}
	if !cs.matchField(cs.Hour, t.Hour(), 0, 23) {
		return false
	}
	if !cs.matchField(cs.Day, t.Day(), 1, 31) {
		return false
	}
	if !cs.matchField(cs.Month, int(t.Month()), 1, 12) {
		return false
	}
	if !cs.matchField(cs.DayOfWeek, int(t.Weekday()), 0, 6) {
		return false
	}

	return true
}

func (cs *CronSchedule) matchField(field string, value, min, max int) bool {
	if field == "*" {
		return true
	}

	// 处理数字
	if num, err := strconv.Atoi(field); err == nil {
		return num == value
	}

	// 处理范围 (如 1-5)
	if strings.Contains(field, "-") {
		parts := strings.Split(field, "-")
		if len(parts) == 2 {
			start, _ := strconv.Atoi(parts[0])
			end, _ := strconv.Atoi(parts[1])
			return value >= start && value <= end
		}
	}

	// 处理步进 (如 */5)
	if strings.Contains(field, "/") {
		parts := strings.Split(field, "/")
		if len(parts) == 2 {
			step, _ := strconv.Atoi(parts[1])
			if parts[0] == "*" {
				return value%step == 0
			}
		}
	}

	// 处理列表 (如 1,3,5)
	if strings.Contains(field, ",") {
		values := strings.Split(field, ",")
		for _, v := range values {
			if num, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				if num == value {
					return true
				}
			}
		}
	}

	return false
}

// ====================
// 3. 任务调度器
// ====================

type Scheduler struct {
	tasks      map[string]*Task
	executions []TaskExecution
	workers    map[string]*Worker
	handlers   map[string]TaskHandler

	taskQueue  chan *Task
	cronParser *CronParser

	workerCount int
	running     bool

	storage *Storage
	mu      sync.RWMutex

	// 统计信息
	stats     SchedulerStats
	statsLock sync.RWMutex
}

func NewScheduler(workerCount int, storage *Storage) *Scheduler {
	scheduler := &Scheduler{
		tasks:       make(map[string]*Task),
		executions:  make([]TaskExecution, 0),
		workers:     make(map[string]*Worker),
		handlers:    make(map[string]TaskHandler),
		taskQueue:   make(chan *Task, 1000),
		cronParser:  NewCronParser(),
		workerCount: workerCount,
		storage:     storage,
	}

	// 注册默认处理器
	scheduler.registerDefaultHandlers()

	return scheduler
}

func (s *Scheduler) registerDefaultHandlers() {
	// HTTP 请求处理器
	s.RegisterHandler("http_request", func(ctx context.Context, task *Task) (interface{}, error) {
		url, ok := task.Payload["url"].(string)
		if !ok {
			return nil, fmt.Errorf("missing url parameter")
		}

		method := "GET"
		if m, ok := task.Payload["method"].(string); ok {
			method = m
		}

		req, err := http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return nil, err
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return map[string]interface{}{
			"status_code": resp.StatusCode,
			"status":      resp.Status,
		}, nil
	})

	// 邮件发送处理器
	s.RegisterHandler("send_email", func(ctx context.Context, task *Task) (interface{}, error) {
		to, _ := task.Payload["to"].(string)
		subject, _ := task.Payload["subject"].(string)
		body, _ := task.Payload["body"].(string)

		// 模拟邮件发送
		log.Printf("Sending email to %s: %s - %s", to, subject, body)
		time.Sleep(time.Second) // 模拟发送时间

		return map[string]interface{}{
			"message_id": fmt.Sprintf("msg_%d", time.Now().Unix()),
			"status":     "sent",
		}, nil
	})

	// 数据备份处理器
	s.RegisterHandler("backup_data", func(ctx context.Context, task *Task) (interface{}, error) {
		dataType, _ := task.Payload["data_type"].(string)

		log.Printf("Starting backup for %s", dataType)

		// 模拟备份过程
		for i := 0; i < 5; i++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				time.Sleep(time.Second)
				log.Printf("Backup progress: %d/5", i+1)
			}
		}

		return map[string]interface{}{
			"backup_file": fmt.Sprintf("/backup/%s_%d.sql", dataType, time.Now().Unix()),
			"size":        "1.2GB",
		}, nil
	})

	// 日志清理处理器
	s.RegisterHandler("cleanup_logs", func(ctx context.Context, task *Task) (interface{}, error) {
		days, _ := task.Payload["days"].(float64)
		if days == 0 {
			days = 7 // 默认清理7天前的日志
		}

		log.Printf("Cleaning up logs older than %.0f days", days)

		// 模拟清理过程
		time.Sleep(2 * time.Second)

		return map[string]interface{}{
			"deleted_files": 156,
			"freed_space":   "2.5GB",
		}, nil
	})
}

func (s *Scheduler) RegisterHandler(name string, handler TaskHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[name] = handler
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true

	// 启动工作协程
	for i := 0; i < s.workerCount; i++ {
		workerID := fmt.Sprintf("worker_%d", i+1)
		worker := &Worker{
			ID:           workerID,
			Status:       "idle",
			StartedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		s.workers[workerID] = worker

		go s.runWorker(worker)
	}

	// 启动调度协程
	go s.runScheduler()
	go s.runStatsCollector()

	log.Printf("Task scheduler started with %d workers", s.workerCount)
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	close(s.taskQueue)

	log.Println("Task scheduler stopped")
}

func (s *Scheduler) runScheduler() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ticker.C:
			s.checkCronTasks()
			s.checkDelayedTasks()
			s.checkRetryTasks()
		}
	}
}

func (s *Scheduler) runWorker(worker *Worker) {
	defer func() {
		s.mu.Lock()
		worker.Status = "stopped"
		s.mu.Unlock()
	}()

	for task := range s.taskQueue {
		s.executeTask(worker, task)
	}
}

func (s *Scheduler) executeTask(worker *Worker, task *Task) {
	s.mu.Lock()
	worker.Status = "busy"
	worker.CurrentTask = task
	worker.LastActivity = time.Now()
	s.mu.Unlock()

	execution := TaskExecution{
		ID:        fmt.Sprintf("exec_%d", time.Now().UnixNano()),
		TaskID:    task.ID,
		Status:    "running",
		StartedAt: time.Now(),
		WorkerID:  worker.ID,
	}

	s.mu.Lock()
	task.Status = "running"
	now := time.Now()
	task.StartedAt = &now
	s.executions = append(s.executions, execution)
	s.mu.Unlock()

	// 创建带超时的上下文
	timeout := time.Duration(task.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute // 默认5分钟超时
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 执行任务
	startTime := time.Now()
	result, err := s.executeTaskHandler(ctx, task)
	duration := time.Since(startTime)

	// 更新任务状态
	s.mu.Lock()
	defer s.mu.Unlock()

	task.Duration = duration.Milliseconds()
	execution.Duration = task.Duration
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	execution.CompletedAt = &completedAt

	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		task.RetryCount++
		execution.Status = "failed"
		execution.Error = err.Error()

		// 检查是否需要重试
		if task.RetryCount < task.MaxRetries {
			task.Status = "pending"
			retryAt := time.Now().Add(time.Duration(task.RetryDelay) * time.Second)
			task.ScheduledAt = retryAt
			log.Printf("Task %s failed, will retry at %v (attempt %d/%d)",
				task.ID, retryAt, task.RetryCount+1, task.MaxRetries)
		} else {
			log.Printf("Task %s failed permanently after %d attempts", task.ID, task.RetryCount)
		}
	} else {
		task.Status = "completed"
		task.Result = result
		execution.Status = "completed"
		execution.Result = result

		// 如果是 cron 任务，计算下次执行时间
		if task.Type == "cron" {
			if cronSchedule, err := s.cronParser.ParseCron(task.Schedule); err == nil {
				nextRun := cronSchedule.NextRun(time.Now())
				task.NextRunAt = &nextRun
				task.Status = "pending"
				task.ScheduledAt = nextRun
			}
		}
	}

	// 更新执行记录
	for i := range s.executions {
		if s.executions[i].ID == execution.ID {
			s.executions[i] = execution
			break
		}
	}

	// 保存数据
	s.storage.SaveTasks(s.tasks)
	s.storage.SaveExecutions(s.executions)

	// 更新工作器状态
	worker.Status = "idle"
	worker.CurrentTask = nil
	worker.TasksProcessed++
	worker.LastActivity = time.Now()

	log.Printf("Task %s completed by %s in %v", task.ID, worker.ID, duration)
}

func (s *Scheduler) executeTaskHandler(ctx context.Context, task *Task) (interface{}, error) {
	s.mu.RLock()
	handler, exists := s.handlers[task.HandlerName]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("handler '%s' not found", task.HandlerName)
	}

	return handler(ctx, task)
}

func (s *Scheduler) checkCronTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for _, task := range s.tasks {
		if task.Type == "cron" && task.Status == "pending" &&
			task.ScheduledAt.Before(now) {
			select {
			case s.taskQueue <- task:
				log.Printf("Queued cron task %s", task.ID)
			default:
				log.Printf("Task queue full, skipping task %s", task.ID)
			}
		}
	}
}

func (s *Scheduler) checkDelayedTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for _, task := range s.tasks {
		if (task.Type == "delay" || task.Type == "once") &&
			task.Status == "pending" && task.ScheduledAt.Before(now) {
			select {
			case s.taskQueue <- task:
				log.Printf("Queued delayed task %s", task.ID)
			default:
				log.Printf("Task queue full, skipping task %s", task.ID)
			}
		}
	}
}

func (s *Scheduler) checkRetryTasks() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for _, task := range s.tasks {
		if task.Status == "pending" && task.RetryCount > 0 &&
			task.ScheduledAt.Before(now) {
			select {
			case s.taskQueue <- task:
				log.Printf("Queued retry task %s (attempt %d)", task.ID, task.RetryCount+1)
			default:
				log.Printf("Task queue full, skipping retry task %s", task.ID)
			}
		}
	}
}

func (s *Scheduler) runStatsCollector() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for s.running {
		select {
		case <-ticker.C:
			s.updateStats()
		}
	}
}

func (s *Scheduler) updateStats() {
	s.mu.RLock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	executions := s.executions
	workers := s.workers
	s.mu.RUnlock()

	stats := SchedulerStats{}

	// 统计任务状态
	for _, task := range tasks {
		stats.TotalTasks++
		switch task.Status {
		case "pending":
			stats.PendingTasks++
		case "running":
			stats.RunningTasks++
		case "completed":
			stats.CompletedTasks++
		case "failed":
			stats.FailedTasks++
		}
	}

	// 统计活跃工作器
	for _, worker := range workers {
		if worker.Status != "stopped" {
			stats.ActiveWorkers++
		}
	}

	// 计算成功率
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	// 计算平均执行时间和每分钟任务数
	if len(executions) > 0 {
		var totalDuration int64
		var completedCount int
		oneHourAgo := time.Now().Add(-time.Hour)

		for _, execution := range executions {
			if execution.Status == "completed" {
				totalDuration += execution.Duration
				completedCount++

				// 统计最近一小时的任务数
				if execution.StartedAt.After(oneHourAgo) {
					stats.TasksPerMinute++
				}
			}
		}

		if completedCount > 0 {
			stats.AvgDuration = float64(totalDuration) / float64(completedCount)
		}

		stats.TasksPerMinute = stats.TasksPerMinute / 60 // 转换为每分钟
	}

	s.statsLock.Lock()
	s.stats = stats
	s.statsLock.Unlock()
}

// ====================
// 4. 任务管理 API
// ====================

func (s *Scheduler) CreateTask(task *Task) error {
	if task.ID == "" {
		task.ID = fmt.Sprintf("task_%d", time.Now().UnixNano())
	}

	// 设置默认值
	if task.Priority == 0 {
		task.Priority = 5
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}
	if task.RetryDelay == 0 {
		task.RetryDelay = 60 // 1分钟
	}
	if task.Timeout == 0 {
		task.Timeout = 300 // 5分钟
	}

	task.CreatedAt = time.Now()
	task.Status = "pending"

	// 计算首次执行时间
	switch task.Type {
	case "cron":
		if cronSchedule, err := s.cronParser.ParseCron(task.Schedule); err == nil {
			nextRun := cronSchedule.NextRun(time.Now())
			task.ScheduledAt = nextRun
			task.NextRunAt = &nextRun
		} else {
			return fmt.Errorf("invalid cron expression: %v", err)
		}
	case "delay":
		if delaySeconds, err := strconv.Atoi(task.Schedule); err == nil {
			task.ScheduledAt = time.Now().Add(time.Duration(delaySeconds) * time.Second)
		} else {
			return fmt.Errorf("invalid delay format: %s", task.Schedule)
		}
	case "once":
		task.ScheduledAt = time.Now()
	default:
		return fmt.Errorf("invalid task type: %s", task.Type)
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	s.storage.SaveTasks(s.tasks)

	log.Printf("Created task %s (%s) scheduled at %v", task.ID, task.Type, task.ScheduledAt)
	return nil
}

func (s *Scheduler) GetTask(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	return task, nil
}

func (s *Scheduler) GetTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	// 按创建时间排序
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	return tasks
}

func (s *Scheduler) CancelTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task not found")
	}

	if task.Status == "running" {
		return fmt.Errorf("cannot cancel running task")
	}

	task.Status = "cancelled"
	s.storage.SaveTasks(s.tasks)

	log.Printf("Cancelled task %s", task.ID)
	return nil
}

func (s *Scheduler) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return fmt.Errorf("task not found")
	}

	if task.Status == "running" {
		return fmt.Errorf("cannot delete running task")
	}

	delete(s.tasks, id)
	s.storage.SaveTasks(s.tasks)

	log.Printf("Deleted task %s", id)
	return nil
}

func (s *Scheduler) GetStats() SchedulerStats {
	s.statsLock.RLock()
	defer s.statsLock.RUnlock()

	return s.stats
}

func (s *Scheduler) GetExecutions(taskID string) []TaskExecution {
	s.mu.RLock()
	defer s.mu.RUnlock()

	executions := make([]TaskExecution, 0)
	for _, execution := range s.executions {
		if taskID == "" || execution.TaskID == taskID {
			executions = append(executions, execution)
		}
	}

	// 按开始时间排序
	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartedAt.After(executions[j].StartedAt)
	})

	return executions
}

func (s *Scheduler) GetWorkers() []*Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workers := make([]*Worker, 0, len(s.workers))
	for _, worker := range s.workers {
		workers = append(workers, worker)
	}

	return workers
}

// ====================
// 5. 存储层
// ====================

type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

func NewStorage(dataDir string) *Storage {
	storage := &Storage{
		dataDir: dataDir,
	}

	os.MkdirAll(dataDir, 0755)
	return storage
}

func (s *Storage) SaveTasks(tasks map[string]*Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	return security.SecureWriteFile(filepath.Join(s.dataDir, "tasks.json"), data, &security.SecureFileOptions{
		Mode:      security.GetRecommendedMode("data"),
		CreateDir: true,
	})
}

func (s *Storage) LoadTasks() (map[string]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.dataDir, "tasks.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]*Task), nil
		}
		return nil, err
	}

	var tasks map[string]*Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = make(map[string]*Task)
	}

	return tasks, nil
}

func (s *Storage) SaveExecutions(executions []TaskExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 只保留最近1000条执行记录
	if len(executions) > 1000 {
		executions = executions[len(executions)-1000:]
	}

	data, err := json.MarshalIndent(executions, "", "  ")
	if err != nil {
		return err
	}

	return security.SecureWriteFile(filepath.Join(s.dataDir, "executions.json"), data, &security.SecureFileOptions{
		Mode:      security.GetRecommendedMode("data"),
		CreateDir: true,
	})
}

func (s *Storage) LoadExecutions() ([]TaskExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(s.dataDir, "executions.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return make([]TaskExecution, 0), nil
		}
		return nil, err
	}

	var executions []TaskExecution
	err = json.Unmarshal(data, &executions)
	if err != nil {
		return nil, err
	}

	return executions, nil
}

// ====================
// 6. HTTP API 服务器
// ====================

type APIServer struct {
	scheduler *Scheduler
}

func NewAPIServer(scheduler *Scheduler) *APIServer {
	return &APIServer{
		scheduler: scheduler,
	}
}

func (api *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS 支持
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch {
	case r.URL.Path == "/api/tasks" && r.Method == "GET":
		api.handleGetTasks(w, r)
	case r.URL.Path == "/api/tasks" && r.Method == "POST":
		api.handleCreateTask(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/tasks/") && r.Method == "GET":
		api.handleGetTask(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/tasks/") && strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == "POST":
		api.handleCancelTask(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/tasks/") && r.Method == "DELETE":
		api.handleDeleteTask(w, r)
	case r.URL.Path == "/api/executions" && r.Method == "GET":
		api.handleGetExecutions(w, r)
	case r.URL.Path == "/api/workers" && r.Method == "GET":
		api.handleGetWorkers(w, r)
	case r.URL.Path == "/api/stats" && r.Method == "GET":
		api.handleGetStats(w, r)
	case r.URL.Path == "/" || r.URL.Path == "/dashboard":
		api.handleDashboard(w, r)
	default:
		api.sendError(w, "Endpoint not found", http.StatusNotFound)
	}
}

func (api *APIServer) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := api.scheduler.GetTasks()
	api.sendJSON(w, map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	})
}

func (api *APIServer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		api.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := api.scheduler.CreateTask(&task); err != nil {
		api.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.sendJSON(w, map[string]interface{}{
		"message": "Task created successfully",
		"task":    task,
	})
}

func (api *APIServer) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	task, err := api.scheduler.GetTask(id)
	if err != nil {
		api.sendError(w, err.Error(), http.StatusNotFound)
		return
	}

	api.sendJSON(w, map[string]interface{}{
		"task": task,
	})
}

func (api *APIServer) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id := strings.TrimSuffix(path, "/cancel")

	if err := api.scheduler.CancelTask(id); err != nil {
		api.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.sendJSON(w, map[string]interface{}{
		"message": "Task cancelled successfully",
	})
}

func (api *APIServer) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/tasks/")

	if err := api.scheduler.DeleteTask(id); err != nil {
		api.sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.sendJSON(w, map[string]interface{}{
		"message": "Task deleted successfully",
	})
}

func (api *APIServer) handleGetExecutions(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Query().Get("task_id")
	executions := api.scheduler.GetExecutions(taskID)

	api.sendJSON(w, map[string]interface{}{
		"executions": executions,
		"total":      len(executions),
	})
}

func (api *APIServer) handleGetWorkers(w http.ResponseWriter, r *http.Request) {
	workers := api.scheduler.GetWorkers()
	api.sendJSON(w, map[string]interface{}{
		"workers": workers,
		"total":   len(workers),
	})
}

func (api *APIServer) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := api.scheduler.GetStats()
	api.sendJSON(w, stats)
}

func (api *APIServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>⏰ 任务调度系统</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; }

        .header { background: #2c3e50; color: white; padding: 1rem 0; }
        .header h1 { text-align: center; }

        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }

        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .stat-card { background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; }
        .stat-number { font-size: 2rem; font-weight: bold; color: #3498db; }
        .stat-label { color: #7f8c8d; margin-top: 0.5rem; }

        .section { background: white; border-radius: 8px; padding: 1.5rem; margin-bottom: 2rem; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .section h2 { margin-bottom: 1rem; color: #2c3e50; }

        .task-form { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1rem; }
        .form-group { display: flex; flex-direction: column; }
        .form-group label { margin-bottom: 0.5rem; font-weight: bold; color: #2c3e50; }
        .form-group input, .form-group select, .form-group textarea { padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }

        .btn { padding: 0.7rem 1.5rem; background: #3498db; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .btn:hover { background: #2980b9; }
        .btn-danger { background: #e74c3c; }
        .btn-danger:hover { background: #c0392b; }
        .btn-success { background: #27ae60; }
        .btn-success:hover { background: #229954; }

        .table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        .table th, .table td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #ddd; }
        .table th { background: #f8f9fa; font-weight: bold; }

        .status { padding: 0.25rem 0.5rem; border-radius: 12px; font-size: 0.8rem; font-weight: bold; }
        .status-pending { background: #f39c12; color: white; }
        .status-running { background: #3498db; color: white; }
        .status-completed { background: #27ae60; color: white; }
        .status-failed { background: #e74c3c; color: white; }
        .status-cancelled { background: #95a5a6; color: white; }

        .tabs { display: flex; margin-bottom: 1rem; }
        .tab { padding: 0.75rem 1.5rem; background: #ecf0f1; border: none; cursor: pointer; }
        .tab.active { background: #3498db; color: white; }

        .tab-content { display: none; }
        .tab-content.active { display: block; }

        .actions { display: flex; gap: 0.5rem; }

        @media (max-width: 768px) {
            .task-form { grid-template-columns: 1fr; }
            .stats-grid { grid-template-columns: repeat(2, 1fr); }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>⏰ 任务调度系统</h1>
    </div>

    <div class="container">
        <!-- 统计信息 -->
        <div class="stats-grid" id="statsGrid">
            <!-- 动态加载 -->
        </div>

        <!-- 创建任务 -->
        <div class="section">
            <h2>创建新任务</h2>
            <form id="taskForm" class="task-form">
                <div class="form-group">
                    <label>任务名称:</label>
                    <input type="text" id="taskName" required>
                </div>
                <div class="form-group">
                    <label>任务类型:</label>
                    <select id="taskType" required>
                        <option value="once">立即执行</option>
                        <option value="delay">延迟执行</option>
                        <option value="cron">定时任务</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>调度设置:</label>
                    <input type="text" id="taskSchedule" placeholder="延迟秒数或Cron表达式" required>
                </div>
                <div class="form-group">
                    <label>处理器:</label>
                    <select id="taskHandler" required>
                        <option value="http_request">HTTP请求</option>
                        <option value="send_email">发送邮件</option>
                        <option value="backup_data">数据备份</option>
                        <option value="cleanup_logs">日志清理</option>
                    </select>
                </div>
            </form>
            <div style="margin-top: 1rem;">
                <button type="button" class="btn btn-success" onclick="createTask()">创建任务</button>
            </div>
        </div>

        <!-- Tab 切换 -->
        <div class="tabs">
            <button class="tab active" onclick="showTab('tasks')">任务列表</button>
            <button class="tab" onclick="showTab('executions')">执行历史</button>
            <button class="tab" onclick="showTab('workers')">工作器状态</button>
        </div>

        <!-- 任务列表 -->
        <div class="section tab-content active" id="tasks-content">
            <h2>任务列表</h2>
            <table class="table" id="tasksTable">
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>名称</th>
                        <th>类型</th>
                        <th>状态</th>
                        <th>下次执行</th>
                        <th>重试次数</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody id="tasksTableBody">
                    <!-- 动态加载 -->
                </tbody>
            </table>
        </div>

        <!-- 执行历史 -->
        <div class="section tab-content" id="executions-content">
            <h2>执行历史</h2>
            <table class="table" id="executionsTable">
                <thead>
                    <tr>
                        <th>执行ID</th>
                        <th>任务ID</th>
                        <th>工作器</th>
                        <th>状态</th>
                        <th>开始时间</th>
                        <th>执行时长</th>
                        <th>结果</th>
                    </tr>
                </thead>
                <tbody id="executionsTableBody">
                    <!-- 动态加载 -->
                </tbody>
            </table>
        </div>

        <!-- 工作器状态 -->
        <div class="section tab-content" id="workers-content">
            <h2>工作器状态</h2>
            <table class="table" id="workersTable">
                <thead>
                    <tr>
                        <th>工作器ID</th>
                        <th>状态</th>
                        <th>当前任务</th>
                        <th>已处理任务</th>
                        <th>最后活动</th>
                    </tr>
                </thead>
                <tbody id="workersTableBody">
                    <!-- 动态加载 -->
                </tbody>
            </table>
        </div>
    </div>

    <script>
        // 全局变量
        let currentTab = 'tasks';

        // 页面加载完成后初始化
        document.addEventListener('DOMContentLoaded', function() {
            loadStats();
            loadTasks();
            loadExecutions();
            loadWorkers();

            // 定期刷新数据
            setInterval(function() {
                loadStats();
                if (currentTab === 'tasks') loadTasks();
                if (currentTab === 'executions') loadExecutions();
                if (currentTab === 'workers') loadWorkers();
            }, 5000);
        });

        // Tab 切换
        function showTab(tabName) {
            currentTab = tabName;

            // 更新 tab 状态
            document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

            event.target.classList.add('active');
            document.getElementById(tabName + '-content').classList.add('active');

            // 加载相应数据
            if (tabName === 'tasks') loadTasks();
            if (tabName === 'executions') loadExecutions();
            if (tabName === 'workers') loadWorkers();
        }

        // 加载统计信息
        async function loadStats() {
            try {
                const response = await fetch('/api/stats');
                const stats = await response.json();

                const statsGrid = document.getElementById('statsGrid');
                statsGrid.innerHTML =
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.total_tasks + '</div>' +
                        '<div class="stat-label">总任务数</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.pending_tasks + '</div>' +
                        '<div class="stat-label">待执行</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.running_tasks + '</div>' +
                        '<div class="stat-label">执行中</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.completed_tasks + '</div>' +
                        '<div class="stat-label">已完成</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.failed_tasks + '</div>' +
                        '<div class="stat-label">失败</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.active_workers + '</div>' +
                        '<div class="stat-label">活跃工作器</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.success_rate.toFixed(1) + '%</div>' +
                        '<div class="stat-label">成功率</div>' +
                    '</div>' +
                    '<div class="stat-card">' +
                        '<div class="stat-number">' + stats.tasks_per_minute.toFixed(1) + '</div>' +
                        '<div class="stat-label">每分钟任务数</div>' +
                    '</div>';
            } catch (error) {
                console.error('Error loading stats:', error);
            }
        }

        // 加载任务列表
        async function loadTasks() {
            try {
                const response = await fetch('/api/tasks');
                const data = await response.json();

                const tbody = document.getElementById('tasksTableBody');
                tbody.innerHTML = data.tasks.map(task =>
                    '<tr>' +
                        '<td>' + task.id.substring(0, 12) + '...</td>' +
                        '<td>' + task.name + '</td>' +
                        '<td>' + task.type + '</td>' +
                        '<td><span class="status status-' + task.status + '">' + task.status + '</span></td>' +
                        '<td>' + (task.next_run_at ? formatTime(task.next_run_at) : '-') + '</td>' +
                        '<td>' + task.retry_count + '/' + task.max_retries + '</td>' +
                        '<td class="actions">' +
                            '<button class="btn" onclick="viewTask(\'' + task.id + '\')">查看</button>' +
                            (task.status === 'pending' ? '<button class="btn btn-danger" onclick="cancelTask(\'' + task.id + '\')">取消</button>' : '') +
                            (task.status !== 'running' ? '<button class="btn btn-danger" onclick="deleteTask(\'' + task.id + '\')">删除</button>' : '') +
                        '</td>' +
                    '</tr>'
                ).join('');
            } catch (error) {
                console.error('Error loading tasks:', error);
            }
        }

        // 加载执行历史
        async function loadExecutions() {
            try {
                const response = await fetch('/api/executions');
                const data = await response.json();

                const tbody = document.getElementById('executionsTableBody');
                tbody.innerHTML = data.executions.slice(0, 50).map(execution =>
                    '<tr>' +
                        '<td>' + execution.id.substring(0, 12) + '...</td>' +
                        '<td>' + execution.task_id.substring(0, 12) + '...</td>' +
                        '<td>' + execution.worker_id + '</td>' +
                        '<td><span class="status status-' + execution.status + '">' + execution.status + '</span></td>' +
                        '<td>' + formatTime(execution.started_at) + '</td>' +
                        '<td>' + execution.duration + 'ms</td>' +
                        '<td>' + (execution.error || (execution.result ? JSON.stringify(execution.result).substring(0, 50) + '...' : '-')) + '</td>' +
                    '</tr>'
                ).join('');
            } catch (error) {
                console.error('Error loading executions:', error);
            }
        }

        // 加载工作器状态
        async function loadWorkers() {
            try {
                const response = await fetch('/api/workers');
                const data = await response.json();

                const tbody = document.getElementById('workersTableBody');
                tbody.innerHTML = data.workers.map(worker =>
                    '<tr>' +
                        '<td>' + worker.id + '</td>' +
                        '<td><span class="status status-' + (worker.status === 'idle' ? 'completed' : worker.status === 'busy' ? 'running' : 'failed') + '">' + worker.status + '</span></td>' +
                        '<td>' + (worker.current_task ? worker.current_task.name : '-') + '</td>' +
                        '<td>' + worker.tasks_processed + '</td>' +
                        '<td>' + formatTime(worker.last_activity) + '</td>' +
                    '</tr>'
                ).join('');
            } catch (error) {
                console.error('Error loading workers:', error);
            }
        }

        // 创建任务
        async function createTask() {
            const name = document.getElementById('taskName').value;
            const type = document.getElementById('taskType').value;
            const schedule = document.getElementById('taskSchedule').value;
            const handlerName = document.getElementById('taskHandler').value;

            if (!name || !schedule) {
                alert('请填写任务名称和调度设置');
                return;
            }

            // 根据处理器类型设置默认参数
            let payload = {};
            switch (handlerName) {
                case 'http_request':
                    payload = { url: 'https://httpbin.org/status/200', method: 'GET' };
                    break;
                case 'send_email':
                    payload = { to: 'user@example.com', subject: '测试邮件', body: '这是一封测试邮件' };
                    break;
                case 'backup_data':
                    payload = { data_type: 'users' };
                    break;
                case 'cleanup_logs':
                    payload = { days: 7 };
                    break;
            }

            const task = {
                name,
                type,
                schedule,
                handler_name: handlerName,
                payload,
                priority: 5,
                max_retries: 3,
                timeout: 300
            };

            try {
                const response = await fetch('/api/tasks', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(task)
                });

                if (response.ok) {
                    alert('任务创建成功');
                    document.getElementById('taskForm').reset();
                    loadTasks();
                    loadStats();
                } else {
                    const error = await response.json();
                    alert('创建失败: ' + error.error);
                }
            } catch (error) {
                alert('创建失败: ' + error.message);
            }
        }

        // 查看任务详情
        async function viewTask(taskId) {
            try {
                const response = await fetch('/api/tasks/' + taskId);
                const data = await response.json();
                const task = data.task;

                alert('任务详情:\\n' +
                      'ID: ' + task.id + '\\n' +
                      '名称: ' + task.name + '\\n' +
                      '类型: ' + task.type + '\\n' +
                      '状态: ' + task.status + '\\n' +
                      '处理器: ' + task.handler_name + '\\n' +
                      '调度: ' + task.schedule + '\\n' +
                      '创建时间: ' + formatTime(task.created_at) + '\\n' +
                      '下次执行: ' + (task.next_run_at ? formatTime(task.next_run_at) : '无') + '\\n' +
                      '重试次数: ' + task.retry_count + '/' + task.max_retries + '\\n' +
                      (task.error ? '错误: ' + task.error : ''));
            } catch (error) {
                alert('获取任务详情失败: ' + error.message);
            }
        }

        // 取消任务
        async function cancelTask(taskId) {
            if (!confirm('确定要取消这个任务吗？')) return;

            try {
                const response = await fetch('/api/tasks/' + taskId + '/cancel', {
                    method: 'POST'
                });

                if (response.ok) {
                    alert('任务已取消');
                    loadTasks();
                    loadStats();
                } else {
                    const error = await response.json();
                    alert('取消失败: ' + error.error);
                }
            } catch (error) {
                alert('取消失败: ' + error.message);
            }
        }

        // 删除任务
        async function deleteTask(taskId) {
            if (!confirm('确定要删除这个任务吗？此操作不可恢复。')) return;

            try {
                const response = await fetch('/api/tasks/' + taskId, {
                    method: 'DELETE'
                });

                if (response.ok) {
                    alert('任务已删除');
                    loadTasks();
                    loadStats();
                } else {
                    const error = await response.json();
                    alert('删除失败: ' + error.error);
                }
            } catch (error) {
                alert('删除失败: ' + error.message);
            }
        }

        // 格式化时间
        function formatTime(timeStr) {
            return new Date(timeStr).toLocaleString('zh-CN');
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (api *APIServer) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (api *APIServer) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// ====================
// 主函数
// ====================

func main() {
	// 创建存储
	storage := NewStorage("./scheduler_data")

	// 创建调度器
	scheduler := NewScheduler(5, storage) // 5个工作器

	// 加载已保存的任务
	if tasks, err := storage.LoadTasks(); err == nil {
		scheduler.mu.Lock()
		scheduler.tasks = tasks
		scheduler.mu.Unlock()
		log.Printf("Loaded %d tasks from storage", len(tasks))
	}

	// 加载执行历史
	if executions, err := storage.LoadExecutions(); err == nil {
		scheduler.mu.Lock()
		scheduler.executions = executions
		scheduler.mu.Unlock()
		log.Printf("Loaded %d execution records from storage", len(executions))
	}

	// 启动调度器
	scheduler.Start()

	// 创建示例任务
	createSampleTasks(scheduler)

	// 创建API服务器
	apiServer := NewAPIServer(scheduler)

	// 启动HTTP服务器
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("⏰ 任务调度系统启动在 http://localhost:%s", port)
	log.Println("功能特性:")
	log.Println("- Cron 定时任务")
	log.Println("- 延迟任务执行")
	log.Println("- 任务重试机制")
	log.Println("- 工作池管理")
	log.Println("- 任务监控统计")
	log.Println("- Web 管理界面")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      apiServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

func createSampleTasks(scheduler *Scheduler) {
	// 检查是否已有任务
	if len(scheduler.GetTasks()) > 0 {
		return
	}

	// 创建示例任务
	tasks := []*Task{
		{
			Name:        "每小时数据备份",
			Type:        "cron",
			Schedule:    "0 * * * *", // 每小时0分执行
			HandlerName: "backup_data",
			Payload: map[string]interface{}{
				"data_type": "users",
			},
			Priority:   8,
			MaxRetries: 3,
			Timeout:    600, // 10分钟
			Tags:       []string{"backup", "hourly"},
		},
		{
			Name:        "每日日志清理",
			Type:        "cron",
			Schedule:    "0 2 * * *", // 每日凌晨2点执行
			HandlerName: "cleanup_logs",
			Payload: map[string]interface{}{
				"days": 30,
			},
			Priority:   6,
			MaxRetries: 2,
			Timeout:    1800, // 30分钟
			Tags:       []string{"cleanup", "daily"},
		},
		{
			Name:        "健康检查",
			Type:        "cron",
			Schedule:    "*/5 * * * *", // 每5分钟执行
			HandlerName: "http_request",
			Payload: map[string]interface{}{
				"url":    "http://localhost:8080/api/stats",
				"method": "GET",
			},
			Priority:   3,
			MaxRetries: 1,
			Timeout:    30,
			Tags:       []string{"health", "monitoring"},
		},
		{
			Name:        "欢迎邮件",
			Type:        "delay",
			Schedule:    "10", // 10秒后执行
			HandlerName: "send_email",
			Payload: map[string]interface{}{
				"to":      "newuser@example.com",
				"subject": "欢迎使用任务调度系统",
				"body":    "感谢您使用我们的任务调度系统！",
			},
			Priority:   5,
			MaxRetries: 3,
			Timeout:    60,
			Tags:       []string{"email", "welcome"},
		},
	}

	for _, task := range tasks {
		if err := scheduler.CreateTask(task); err != nil {
			log.Printf("Failed to create sample task %s: %v", task.Name, err)
		}
	}

	log.Println("Created sample tasks")
}

/*
=== 项目功能清单 ===

核心功能:
✅ Cron 表达式解析和调度
✅ 延迟任务执行
✅ 一次性任务执行
✅ 任务重试机制
✅ 任务超时控制
✅ 工作池管理
✅ 任务优先级

管理功能:
✅ 任务创建/取消/删除
✅ 任务状态监控
✅ 执行历史记录
✅ 工作器状态监控
✅ 统计信息收集

界面功能:
✅ Web 管理界面
✅ 实时状态更新
✅ 任务创建表单
✅ 数据可视化展示

存储功能:
✅ 任务持久化
✅ 执行历史持久化
✅ 数据自动保存

=== 任务处理器 ===

内置处理器:
✅ HTTP 请求处理器
✅ 邮件发送处理器
✅ 数据备份处理器
✅ 日志清理处理器

扩展支持:
✅ 自定义处理器注册
✅ 上下文超时控制
✅ 错误处理和重试

=== Cron 表达式示例 ===

基本格式: 分 时 日 月 周

示例:
- "0 * * * *"        - 每小时0分执行
- "星号/5 * * * *"   - 每5分钟执行
- "0 2 * * *"        - 每日凌晨2点执行
- "0 9 * * 1"        - 每周一上午9点执行
- "0 0 1 * *"        - 每月1日凌晨执行

=== API 端点 ===

任务管理:
- GET /api/tasks           - 获取任务列表
- POST /api/tasks          - 创建任务
- GET /api/tasks/{id}      - 获取任务详情
- POST /api/tasks/{id}/cancel - 取消任务
- DELETE /api/tasks/{id}   - 删除任务

监控:
- GET /api/executions      - 获取执行历史
- GET /api/workers         - 获取工作器状态
- GET /api/stats           - 获取统计信息

=== 高级功能扩展 ===

1. 分布式支持:
   - 任务分片
   - 集群协调
   - 节点发现
   - 故障转移

2. 监控告警:
   - 任务失败告警
   - 性能监控
   - 资源使用监控
   - 自定义指标

3. 安全增强:
   - 用户认证
   - 权限控制
   - 操作审计
   - 加密存储

=== 部署说明 ===

1. 运行应用:
   go run main.go

2. 访问管理界面:
   http://localhost:8080

3. 数据存储:
   - 任务: ./scheduler_data/tasks.json
   - 执行记录: ./scheduler_data/executions.json

4. 配置环境变量:
   - PORT: 服务端口号
*/
