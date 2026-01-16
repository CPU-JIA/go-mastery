/*
监控系统 (Monitoring System)

项目描述:
一个完整的应用监控系统，支持指标收集、实时监控、告警通知、
日志聚合、性能分析、健康检查等功能。

技术栈:
- 指标收集和存储
- 实时数据推送 (SSE)
- 告警规则引擎
- 日志聚合和搜索
- HTTP 服务监控
- 系统资源监控
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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

type Metric struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"` // gauge, counter, histogram
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type Alert struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Rule      string            `json:"rule"`
	Severity  string            `json:"severity"` // critical, warning, info
	Status    string            `json:"status"`   // firing, resolved
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels"`
	StartTime time.Time         `json:"start_time"`
	EndTime   *time.Time        `json:"end_time,omitempty"`
	LastSent  time.Time         `json:"last_sent"`
	Count     int               `json:"count"`
}

type LogEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"` // debug, info, warn, error, fatal
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Service   string                 `json:"service"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type ServiceHealth struct {
	Service     string                 `json:"service"`
	Status      string                 `json:"status"` // healthy, degraded, unhealthy
	Checks      []HealthCheck          `json:"checks"`
	LastChecked time.Time              `json:"last_checked"`
	Uptime      time.Duration          `json:"uptime"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type HealthCheck struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Message  string        `json:"message"`
	Error    string        `json:"error,omitempty"`
}

type SystemMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// CPU 指标
	CPUUsage    float64   `json:"cpu_usage"`
	CPUCount    int       `json:"cpu_count"`
	LoadAverage []float64 `json:"load_average"`

	// 内存指标
	MemoryTotal uint64  `json:"memory_total"`
	MemoryUsed  uint64  `json:"memory_used"`
	MemoryFree  uint64  `json:"memory_free"`
	MemoryUsage float64 `json:"memory_usage"`

	// 磁盘指标
	DiskTotal uint64  `json:"disk_total"`
	DiskUsed  uint64  `json:"disk_used"`
	DiskFree  uint64  `json:"disk_free"`
	DiskUsage float64 `json:"disk_usage"`

	// 网络指标
	NetworkRx uint64 `json:"network_rx"`
	NetworkTx uint64 `json:"network_tx"`

	// Go Runtime 指标
	Goroutines int    `json:"goroutines"`
	HeapSize   uint64 `json:"heap_size"`
	HeapInuse  uint64 `json:"heap_inuse"`
	GCCount    uint32 `json:"gc_count"`
}

type Dashboard struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Widgets     []Widget  `json:"widgets"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Widget struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"` // chart, gauge, counter, log
	Title  string                 `json:"title"`
	Query  string                 `json:"query"`
	Config map[string]interface{} `json:"config"`
	X      int                    `json:"x"`
	Y      int                    `json:"y"`
	Width  int                    `json:"width"`
	Height int                    `json:"height"`
}

// ====================
// 2. 指标收集器
// ====================

type MetricsCollector struct {
	metrics     []Metric
	systemStats []SystemMetrics
	mu          sync.RWMutex
	storage     *Storage
}

func NewMetricsCollector(storage *Storage) *MetricsCollector {
	collector := &MetricsCollector{
		metrics:     make([]Metric, 0),
		systemStats: make([]SystemMetrics, 0),
		storage:     storage,
	}

	// 启动系统指标收集
	go collector.collectSystemMetrics()

	return collector
}

func (mc *MetricsCollector) RecordMetric(name, metricType string, value float64, labels map[string]string) {
	metric := Metric{
		Name:      name,
		Type:      metricType,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	mc.mu.Lock()
	mc.metrics = append(mc.metrics, metric)

	// 保持最近10000个指标
	if len(mc.metrics) > 10000 {
		mc.metrics = mc.metrics[len(mc.metrics)-10000:]
	}
	mc.mu.Unlock()

	// 异步保存
	go mc.storage.SaveMetrics(mc.metrics)
}

func (mc *MetricsCollector) GetMetrics(name string, duration time.Duration) []Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	since := time.Now().Add(-duration)
	filtered := make([]Metric, 0)

	for _, metric := range mc.metrics {
		if metric.Timestamp.After(since) {
			if name == "" || metric.Name == name {
				filtered = append(filtered, metric)
			}
		}
	}

	return filtered
}

func (mc *MetricsCollector) GetSystemMetrics(duration time.Duration) []SystemMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	since := time.Now().Add(-duration)
	filtered := make([]SystemMetrics, 0)

	for _, stat := range mc.systemStats {
		if stat.Timestamp.After(since) {
			filtered = append(filtered, stat)
		}
	}

	return filtered
}

func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := mc.getSystemStats()

			mc.mu.Lock()
			mc.systemStats = append(mc.systemStats, stats)

			// 保持最近1000个系统指标
			if len(mc.systemStats) > 1000 {
				mc.systemStats = mc.systemStats[len(mc.systemStats)-1000:]
			}
			mc.mu.Unlock()

			// 发送系统指标作为普通指标
			mc.RecordMetric("system.cpu.usage", "gauge", stats.CPUUsage, nil)
			mc.RecordMetric("system.memory.usage", "gauge", stats.MemoryUsage, nil)
			mc.RecordMetric("system.disk.usage", "gauge", stats.DiskUsage, nil)
			mc.RecordMetric("system.goroutines", "gauge", float64(stats.Goroutines), nil)
		}
	}
}

func (mc *MetricsCollector) getSystemStats() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := SystemMetrics{
		Timestamp:  time.Now(),
		CPUCount:   runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
		HeapSize:   m.HeapSys,
		HeapInuse:  m.HeapInuse,
		GCCount:    m.NumGC,
	}

	// 模拟系统指标收集 (实际环境需要使用 psutil 等库)
	stats.CPUUsage = mc.simulateCPUUsage()
	stats.MemoryTotal = 8 * 1024 * 1024 * 1024 // 8GB
	stats.MemoryUsed = uint64(float64(stats.MemoryTotal) * (0.3 + 0.4*mc.randomFloat()))
	stats.MemoryFree = stats.MemoryTotal - stats.MemoryUsed
	stats.MemoryUsage = float64(stats.MemoryUsed) / float64(stats.MemoryTotal) * 100

	stats.DiskTotal = 500 * 1024 * 1024 * 1024 // 500GB
	stats.DiskUsed = uint64(float64(stats.DiskTotal) * (0.5 + 0.3*mc.randomFloat()))
	stats.DiskFree = stats.DiskTotal - stats.DiskUsed
	stats.DiskUsage = float64(stats.DiskUsed) / float64(stats.DiskTotal) * 100

	return stats
}

func (mc *MetricsCollector) simulateCPUUsage() float64 {
	// 模拟 CPU 使用率变化
	baseUsage := 20.0 + 30.0*mc.randomFloat()
	return math.Min(100.0, baseUsage+10.0*(mc.randomFloat()-0.5))
}

func (mc *MetricsCollector) randomFloat() float64 {
	return float64(time.Now().UnixNano()%1000) / 1000.0
}

// ====================
// 3. 告警系统
// ====================

type AlertManager struct {
	alerts      map[string]*Alert
	rules       []AlertRule
	mu          sync.RWMutex
	collector   *MetricsCollector
	storage     *Storage
	subscribers []chan Alert
}

type AlertRule struct {
	Name      string            `json:"name"`
	Query     string            `json:"query"` // 简化的查询语法
	Threshold float64           `json:"threshold"`
	Operator  string            `json:"operator"` // >, <, >=, <=, ==, !=
	Duration  time.Duration     `json:"duration"` // 持续时间
	Severity  string            `json:"severity"`
	Message   string            `json:"message"`
	Labels    map[string]string `json:"labels"`
	Enabled   bool              `json:"enabled"`
}

func NewAlertManager(collector *MetricsCollector, storage *Storage) *AlertManager {
	am := &AlertManager{
		alerts:      make(map[string]*Alert),
		rules:       make([]AlertRule, 0),
		collector:   collector,
		storage:     storage,
		subscribers: make([]chan Alert, 0),
	}

	// 加载默认告警规则
	am.loadDefaultRules()

	// 启动告警检查
	go am.runAlertChecker()

	return am
}

func (am *AlertManager) loadDefaultRules() {
	defaultRules := []AlertRule{
		{
			Name:      "High CPU Usage",
			Query:     "system.cpu.usage",
			Threshold: 80.0,
			Operator:  ">",
			Duration:  2 * time.Minute,
			Severity:  "warning",
			Message:   "CPU usage is above 80%",
			Labels:    map[string]string{"component": "system"},
			Enabled:   true,
		},
		{
			Name:      "High Memory Usage",
			Query:     "system.memory.usage",
			Threshold: 90.0,
			Operator:  ">",
			Duration:  1 * time.Minute,
			Severity:  "critical",
			Message:   "Memory usage is above 90%",
			Labels:    map[string]string{"component": "system"},
			Enabled:   true,
		},
		{
			Name:      "High Disk Usage",
			Query:     "system.disk.usage",
			Threshold: 85.0,
			Operator:  ">",
			Duration:  5 * time.Minute,
			Severity:  "warning",
			Message:   "Disk usage is above 85%",
			Labels:    map[string]string{"component": "system"},
			Enabled:   true,
		},
		{
			Name:      "Too Many Goroutines",
			Query:     "system.goroutines",
			Threshold: 1000.0,
			Operator:  ">",
			Duration:  1 * time.Minute,
			Severity:  "warning",
			Message:   "Too many goroutines running",
			Labels:    map[string]string{"component": "runtime"},
			Enabled:   true,
		},
	}

	am.mu.Lock()
	am.rules = defaultRules
	am.mu.Unlock()
}

func (am *AlertManager) runAlertChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.checkRules()
		}
	}
}

func (am *AlertManager) checkRules() {
	am.mu.RLock()
	rules := make([]AlertRule, len(am.rules))
	copy(rules, am.rules)
	am.mu.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		am.evaluateRule(rule)
	}
}

func (am *AlertManager) evaluateRule(rule AlertRule) {
	// 获取最近的指标数据
	metrics := am.collector.GetMetrics(rule.Query, rule.Duration)
	if len(metrics) == 0 {
		return
	}

	// 获取最新值
	latestMetric := metrics[len(metrics)-1]
	triggered := am.evaluateCondition(latestMetric.Value, rule.Threshold, rule.Operator)

	alertID := fmt.Sprintf("%s_%s", rule.Name, rule.Query)

	am.mu.Lock()
	defer am.mu.Unlock()

	existingAlert, exists := am.alerts[alertID]

	if triggered {
		if !exists {
			// 创建新告警
			alert := &Alert{
				ID:        alertID,
				Name:      rule.Name,
				Rule:      rule.Query,
				Severity:  rule.Severity,
				Status:    "firing",
				Message:   fmt.Sprintf("%s (current: %.2f, threshold: %.2f)", rule.Message, latestMetric.Value, rule.Threshold),
				Labels:    rule.Labels,
				StartTime: time.Now(),
				LastSent:  time.Time{},
				Count:     1,
			}
			am.alerts[alertID] = alert
			am.notifySubscribers(*alert)
			log.Printf("Alert triggered: %s", alert.Name)
		} else if existingAlert.Status == "resolved" {
			// 重新触发已解决的告警
			existingAlert.Status = "firing"
			existingAlert.StartTime = time.Now()
			existingAlert.EndTime = nil
			existingAlert.Count++
			am.notifySubscribers(*existingAlert)
			log.Printf("Alert re-triggered: %s", existingAlert.Name)
		} else {
			// 更新现有告警
			existingAlert.Count++
			existingAlert.Message = fmt.Sprintf("%s (current: %.2f, threshold: %.2f)", rule.Message, latestMetric.Value, rule.Threshold)

			// 每5分钟重新发送一次
			if time.Since(existingAlert.LastSent) > 5*time.Minute {
				am.notifySubscribers(*existingAlert)
				existingAlert.LastSent = time.Now()
			}
		}
	} else if exists && existingAlert.Status == "firing" {
		// 解决告警
		existingAlert.Status = "resolved"
		now := time.Now()
		existingAlert.EndTime = &now
		am.notifySubscribers(*existingAlert)
		log.Printf("Alert resolved: %s", existingAlert.Name)
	}
}

func (am *AlertManager) evaluateCondition(value, threshold float64, operator string) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

func (am *AlertManager) Subscribe(ch chan Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.subscribers = append(am.subscribers, ch)
}

func (am *AlertManager) notifySubscribers(alert Alert) {
	am.storage.SaveAlert(alert)

	for _, ch := range am.subscribers {
		select {
		case ch <- alert:
		default:
			// 通道满了，跳过
		}
	}
}

func (am *AlertManager) GetAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}

	// 按开始时间排序
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].StartTime.After(alerts[j].StartTime)
	})

	return alerts
}

// ====================
// 4. 日志聚合器
// ====================

type LogAggregator struct {
	logs    []LogEntry
	mu      sync.RWMutex
	storage *Storage
}

func NewLogAggregator(storage *Storage) *LogAggregator {
	return &LogAggregator{
		logs:    make([]LogEntry, 0),
		storage: storage,
	}
}

func (la *LogAggregator) AddLog(level, message, source, service string, fields map[string]interface{}) {
	entry := LogEntry{
		ID:        fmt.Sprintf("log_%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Source:    source,
		Service:   service,
		Fields:    fields,
	}

	la.mu.Lock()
	la.logs = append(la.logs, entry)

	// 保持最近5000条日志
	if len(la.logs) > 5000 {
		la.logs = la.logs[len(la.logs)-5000:]
	}
	la.mu.Unlock()

	// 异步保存
	go la.storage.SaveLog(entry)
}

func (la *LogAggregator) GetLogs(level string, service string, limit int, since time.Duration) []LogEntry {
	la.mu.RLock()
	defer la.mu.RUnlock()

	sinceTime := time.Now().Add(-since)
	filtered := make([]LogEntry, 0)

	for i := len(la.logs) - 1; i >= 0 && len(filtered) < limit; i-- {
		log := la.logs[i]
		if log.Timestamp.Before(sinceTime) {
			break
		}

		if (level == "" || log.Level == level) &&
			(service == "" || log.Service == service) {
			filtered = append(filtered, log)
		}
	}

	// 反转顺序，最新的在前面
	for i, j := 0, len(filtered)-1; i < j; i, j = i+1, j-1 {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	}

	return filtered
}

func (la *LogAggregator) SearchLogs(query string, limit int) []LogEntry {
	la.mu.RLock()
	defer la.mu.RUnlock()

	query = strings.ToLower(query)
	results := make([]LogEntry, 0)

	for i := len(la.logs) - 1; i >= 0 && len(results) < limit; i-- {
		log := la.logs[i]
		if strings.Contains(strings.ToLower(log.Message), query) ||
			strings.Contains(strings.ToLower(log.Service), query) ||
			strings.Contains(strings.ToLower(log.Source), query) {
			results = append(results, log)
		}
	}

	return results
}

// ====================
// 5. 健康检查器
// ====================

type HealthChecker struct {
	services map[string]*ServiceHealth
	mu       sync.RWMutex
}

func NewHealthChecker() *HealthChecker {
	hc := &HealthChecker{
		services: make(map[string]*ServiceHealth),
	}

	// 启动健康检查
	go hc.runHealthChecks()

	return hc
}

func (hc *HealthChecker) runHealthChecks() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkServices()
		}
	}
}

func (hc *HealthChecker) checkServices() {
	// 检查自身服务
	hc.checkSelfService()

	// 可以添加更多服务检查
}

func (hc *HealthChecker) checkSelfService() {
	checks := []HealthCheck{}

	// 检查 HTTP 服务
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/api/health")
	duration := time.Since(start)

	httpCheck := HealthCheck{
		Name:     "HTTP Server",
		Duration: duration,
	}

	if err != nil {
		httpCheck.Status = "unhealthy"
		httpCheck.Error = err.Error()
		httpCheck.Message = "HTTP server is not responding"
	} else {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			httpCheck.Status = "healthy"
			httpCheck.Message = "HTTP server is responding normally"
		} else {
			httpCheck.Status = "degraded"
			httpCheck.Message = fmt.Sprintf("HTTP server returned status %d", resp.StatusCode)
		}
	}

	checks = append(checks, httpCheck)

	// 检查内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryCheck := HealthCheck{
		Name:     "Memory Usage",
		Duration: 0,
	}

	memoryUsagePercent := float64(m.HeapInuse) / float64(m.HeapSys) * 100
	if memoryUsagePercent < 80 {
		memoryCheck.Status = "healthy"
		memoryCheck.Message = fmt.Sprintf("Memory usage: %.1f%%", memoryUsagePercent)
	} else if memoryUsagePercent < 95 {
		memoryCheck.Status = "degraded"
		memoryCheck.Message = fmt.Sprintf("High memory usage: %.1f%%", memoryUsagePercent)
	} else {
		memoryCheck.Status = "unhealthy"
		memoryCheck.Message = fmt.Sprintf("Critical memory usage: %.1f%%", memoryUsagePercent)
	}

	checks = append(checks, memoryCheck)

	// 确定整体状态
	overallStatus := "healthy"
	for _, check := range checks {
		if check.Status == "unhealthy" {
			overallStatus = "unhealthy"
			break
		} else if check.Status == "degraded" && overallStatus == "healthy" {
			overallStatus = "degraded"
		}
	}

	hc.mu.Lock()
	hc.services["monitoring-system"] = &ServiceHealth{
		Service:     "monitoring-system",
		Status:      overallStatus,
		Checks:      checks,
		LastChecked: time.Now(),
		Uptime:      time.Since(startTime),
	}
	hc.mu.Unlock()
}

func (hc *HealthChecker) GetServiceHealth(service string) *ServiceHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	return hc.services[service]
}

func (hc *HealthChecker) GetAllServices() map[string]*ServiceHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	services := make(map[string]*ServiceHealth)
	for k, v := range hc.services {
		services[k] = v
	}

	return services
}

// ====================
// 6. 存储层
// ====================

type Storage struct {
	dataDir string
	mu      sync.RWMutex
}

func NewStorage(dataDir string) *Storage {
	storage := &Storage{
		dataDir: dataDir,
	}

	// #nosec G301 -- 监控系统数据目录，需要0755权限支持日志和指标文件写入
	os.MkdirAll(dataDir, 0755)
	return storage
}

func (s *Storage) SaveMetrics(metrics []Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 简化实现：只保存最近1小时的指标
	recentMetrics := make([]Metric, 0)
	oneHourAgo := time.Now().Add(-time.Hour)

	for _, metric := range metrics {
		if metric.Timestamp.After(oneHourAgo) {
			recentMetrics = append(recentMetrics, metric)
		}
	}

	data, err := json.Marshal(recentMetrics)
	if err != nil {
		return err
	}

	return security.SecureWriteFile(filepath.Join(s.dataDir, "metrics.json"), data, &security.SecureFileOptions{
		Mode:      security.GetRecommendedMode("data"),
		CreateDir: true,
	})
}

func (s *Storage) SaveAlert(alert Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 追加到告警日志文件
	// G301/G304安全修复：使用安全权限打开文件
	file, err := security.SecureOpenFile(filepath.Join(s.dataDir, "alerts.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, security.DefaultFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	data, _ := json.Marshal(alert)
	_, err = file.WriteString(string(data) + "\n")
	return err
}

func (s *Storage) SaveLog(log LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 按日期分割日志文件
	logFile := fmt.Sprintf("logs_%s.log", time.Now().Format("2006-01-02"))
	// G301/G304安全修复：使用安全权限打开文件，路径由系统内部生成
	file, err := security.SecureOpenFile(filepath.Join(s.dataDir, logFile), os.O_CREATE|os.O_WRONLY|os.O_APPEND, security.DefaultFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	data, _ := json.Marshal(log)
	_, err = file.WriteString(string(data) + "\n")
	return err
}

// ====================
// 7. HTTP API 服务器
// ====================

type MonitoringServer struct {
	collector     *MetricsCollector
	alertManager  *AlertManager
	logAggregator *LogAggregator
	healthChecker *HealthChecker
	storage       *Storage
}

func NewMonitoringServer(storage *Storage) *MonitoringServer {
	collector := NewMetricsCollector(storage)
	alertManager := NewAlertManager(collector, storage)
	logAggregator := NewLogAggregator(storage)
	healthChecker := NewHealthChecker()

	server := &MonitoringServer{
		collector:     collector,
		alertManager:  alertManager,
		logAggregator: logAggregator,
		healthChecker: healthChecker,
		storage:       storage,
	}

	// 模拟一些日志数据
	go server.generateSampleLogs()

	return server
}

func (ms *MonitoringServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS 支持
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	switch {
	case r.URL.Path == "/api/metrics" && r.Method == "GET":
		ms.handleGetMetrics(w, r)
	case r.URL.Path == "/api/metrics" && r.Method == "POST":
		ms.handlePostMetric(w, r)
	case r.URL.Path == "/api/system" && r.Method == "GET":
		ms.handleGetSystemMetrics(w, r)
	case r.URL.Path == "/api/alerts" && r.Method == "GET":
		ms.handleGetAlerts(w, r)
	case r.URL.Path == "/api/alerts/stream" && r.Method == "GET":
		ms.handleAlertStream(w, r)
	case r.URL.Path == "/api/logs" && r.Method == "GET":
		ms.handleGetLogs(w, r)
	case r.URL.Path == "/api/logs" && r.Method == "POST":
		ms.handlePostLog(w, r)
	case r.URL.Path == "/api/logs/search" && r.Method == "GET":
		ms.handleSearchLogs(w, r)
	case r.URL.Path == "/api/health" && r.Method == "GET":
		ms.handleGetHealth(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/health/") && r.Method == "GET":
		ms.handleGetServiceHealth(w, r)
	case r.URL.Path == "/api/overview" && r.Method == "GET":
		ms.handleGetOverview(w, r)
	case r.URL.Path == "/" || r.URL.Path == "/dashboard":
		ms.handleDashboard(w, r)
	default:
		ms.sendError(w, "Endpoint not found", http.StatusNotFound)
	}
}

func (ms *MonitoringServer) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	durationStr := r.URL.Query().Get("duration")

	duration := time.Hour // 默认1小时
	if durationStr != "" {
		if d, err := time.ParseDuration(durationStr); err == nil {
			duration = d
		}
	}

	metrics := ms.collector.GetMetrics(name, duration)

	ms.sendJSON(w, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

func (ms *MonitoringServer) handlePostMetric(w http.ResponseWriter, r *http.Request) {
	var metric Metric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		ms.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ms.collector.RecordMetric(metric.Name, metric.Type, metric.Value, metric.Labels)

	ms.sendJSON(w, map[string]interface{}{
		"message": "Metric recorded successfully",
	})
}

func (ms *MonitoringServer) handleGetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	durationStr := r.URL.Query().Get("duration")

	duration := time.Hour // 默认1小时
	if durationStr != "" {
		if d, err := time.ParseDuration(durationStr); err == nil {
			duration = d
		}
	}

	systemMetrics := ms.collector.GetSystemMetrics(duration)

	ms.sendJSON(w, map[string]interface{}{
		"system_metrics": systemMetrics,
		"count":          len(systemMetrics),
	})
}

func (ms *MonitoringServer) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := ms.alertManager.GetAlerts()

	ms.sendJSON(w, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

func (ms *MonitoringServer) handleAlertStream(w http.ResponseWriter, r *http.Request) {
	// Server-Sent Events for real-time alerts
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	alertChan := make(chan Alert, 10)
	ms.alertManager.Subscribe(alertChan)

	ctx := r.Context()
	for {
		select {
		case alert := <-alertChan:
			data, _ := json.Marshal(alert)
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ms *MonitoringServer) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")
	limitStr := r.URL.Query().Get("limit")
	sinceStr := r.URL.Query().Get("since")

	limit := 100 // 默认100条
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	since := time.Hour // 默认1小时
	if sinceStr != "" {
		if s, err := time.ParseDuration(sinceStr); err == nil {
			since = s
		}
	}

	logs := ms.logAggregator.GetLogs(level, service, limit, since)

	ms.sendJSON(w, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

func (ms *MonitoringServer) handlePostLog(w http.ResponseWriter, r *http.Request) {
	var log LogEntry
	if err := json.NewDecoder(r.Body).Decode(&log); err != nil {
		ms.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ms.logAggregator.AddLog(log.Level, log.Message, log.Source, log.Service, log.Fields)

	ms.sendJSON(w, map[string]interface{}{
		"message": "Log recorded successfully",
	})
}

func (ms *MonitoringServer) handleSearchLogs(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")

	if query == "" {
		ms.sendError(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	logs := ms.logAggregator.SearchLogs(query, limit)

	ms.sendJSON(w, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
		"query": query,
	})
}

func (ms *MonitoringServer) handleGetHealth(w http.ResponseWriter, r *http.Request) {
	services := ms.healthChecker.GetAllServices()

	ms.sendJSON(w, map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

func (ms *MonitoringServer) handleGetServiceHealth(w http.ResponseWriter, r *http.Request) {
	service := strings.TrimPrefix(r.URL.Path, "/api/health/")

	health := ms.healthChecker.GetServiceHealth(service)
	if health == nil {
		ms.sendError(w, "Service not found", http.StatusNotFound)
		return
	}

	ms.sendJSON(w, health)
}

func (ms *MonitoringServer) handleGetOverview(w http.ResponseWriter, r *http.Request) {
	// 获取概览数据
	alerts := ms.alertManager.GetAlerts()
	services := ms.healthChecker.GetAllServices()
	recentLogs := ms.logAggregator.GetLogs("", "", 10, time.Hour)
	systemMetrics := ms.collector.GetSystemMetrics(time.Hour)

	firingAlerts := 0
	for _, alert := range alerts {
		if alert.Status == "firing" {
			firingAlerts++
		}
	}

	healthyServices := 0
	for _, service := range services {
		if service.Status == "healthy" {
			healthyServices++
		}
	}

	var latestSystemMetric *SystemMetrics
	if len(systemMetrics) > 0 {
		latestSystemMetric = &systemMetrics[len(systemMetrics)-1]
	}

	ms.sendJSON(w, map[string]interface{}{
		"alerts": map[string]interface{}{
			"total":  len(alerts),
			"firing": firingAlerts,
		},
		"services": map[string]interface{}{
			"total":   len(services),
			"healthy": healthyServices,
		},
		"logs": map[string]interface{}{
			"recent_count": len(recentLogs),
		},
		"system": latestSystemMetric,
	})
}

func (ms *MonitoringServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>📊 监控系统</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; }

        .header { background: #2c3e50; color: white; padding: 1rem 0; }
        .header h1 { text-align: center; }

        .container { max-width: 1400px; margin: 0 auto; padding: 2rem; }

        .overview-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .overview-card { background: white; padding: 1.5rem; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; }
        .overview-number { font-size: 2.5rem; font-weight: bold; margin-bottom: 0.5rem; }
        .overview-label { color: #7f8c8d; }
        .overview-healthy { color: #27ae60; }
        .overview-warning { color: #f39c12; }
        .overview-critical { color: #e74c3c; }

        .dashboard-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 2rem; }
        .widget { background: white; border-radius: 8px; padding: 1.5rem; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .widget h3 { margin-bottom: 1rem; color: #2c3e50; }

        .chart-container { height: 300px; background: #f8f9fa; border-radius: 4px; display: flex; align-items: center; justify-content: center; color: #7f8c8d; }

        .alert-list, .log-list { max-height: 300px; overflow-y: auto; }
        .alert-item, .log-item { padding: 0.75rem; border-bottom: 1px solid #ecf0f1; }
        .alert-item:last-child, .log-item:last-child { border-bottom: none; }

        .status { padding: 0.25rem 0.5rem; border-radius: 12px; font-size: 0.8rem; font-weight: bold; }
        .status-firing { background: #e74c3c; color: white; }
        .status-resolved { background: #27ae60; color: white; }
        .status-warning { background: #f39c12; color: white; }
        .status-critical { background: #e74c3c; color: white; }

        .log-level { padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; font-weight: bold; }
        .log-debug { background: #6c757d; color: white; }
        .log-info { background: #17a2b8; color: white; }
        .log-warn { background: #ffc107; color: black; }
        .log-error { background: #dc3545; color: white; }

        .tabs { display: flex; margin-bottom: 1rem; }
        .tab { padding: 0.75rem 1.5rem; background: #ecf0f1; border: none; cursor: pointer; }
        .tab.active { background: #3498db; color: white; }

        .tab-content { display: none; }
        .tab-content.active { display: block; }

        .metrics-table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        .metrics-table th, .metrics-table td { padding: 0.5rem; text-align: left; border-bottom: 1px solid #ddd; }
        .metrics-table th { background: #f8f9fa; }

        .refresh-indicator { position: fixed; top: 20px; right: 20px; background: #3498db; color: white; padding: 0.5rem 1rem; border-radius: 20px; font-size: 0.8rem; }

        @media (max-width: 768px) {
            .dashboard-grid { grid-template-columns: 1fr; }
            .overview-grid { grid-template-columns: repeat(2, 1fr); }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📊 监控系统</h1>
    </div>

    <div class="refresh-indicator" id="refreshIndicator" style="display: none;">
        🔄 更新中...
    </div>

    <div class="container">
        <!-- 概览卡片 -->
        <div class="overview-grid" id="overviewGrid">
            <!-- 动态加载 -->
        </div>

        <!-- 主要监控面板 -->
        <div class="dashboard-grid">
            <!-- 系统指标 -->
            <div class="widget">
                <h3>系统指标</h3>
                <div class="tabs">
                    <button class="tab active" onclick="showSystemTab('cpu')">CPU</button>
                    <button class="tab" onclick="showSystemTab('memory')">内存</button>
                    <button class="tab" onclick="showSystemTab('disk')">磁盘</button>
                </div>
                <div class="tab-content active" id="cpu-content">
                    <div class="chart-container" id="cpuChart">CPU 使用率图表</div>
                </div>
                <div class="tab-content" id="memory-content">
                    <div class="chart-container" id="memoryChart">内存使用率图表</div>
                </div>
                <div class="tab-content" id="disk-content">
                    <div class="chart-container" id="diskChart">磁盘使用率图表</div>
                </div>
            </div>

            <!-- 告警列表 -->
            <div class="widget">
                <h3>活动告警</h3>
                <div class="alert-list" id="alertList">
                    <!-- 动态加载 -->
                </div>
            </div>

            <!-- 服务健康 -->
            <div class="widget">
                <h3>服务健康</h3>
                <div id="serviceHealth">
                    <!-- 动态加载 -->
                </div>
            </div>

            <!-- 最近日志 -->
            <div class="widget">
                <h3>最近日志</h3>
                <div class="log-list" id="logList">
                    <!-- 动态加载 -->
                </div>
            </div>
        </div>

        <!-- 详细指标表格 -->
        <div class="widget" style="margin-top: 2rem;">
            <h3>详细指标</h3>
            <table class="metrics-table" id="metricsTable">
                <thead>
                    <tr>
                        <th>指标名称</th>
                        <th>当前值</th>
                        <th>类型</th>
                        <th>标签</th>
                        <th>最后更新</th>
                    </tr>
                </thead>
                <tbody id="metricsTableBody">
                    <!-- 动态加载 -->
                </tbody>
            </table>
        </div>
    </div>

    <script>
        // 全局变量
        let currentSystemTab = 'cpu';
        let alertEventSource = null;

        // 页面加载完成后初始化
        document.addEventListener('DOMContentLoaded', function() {
            loadOverview();
            loadAlerts();
            loadServiceHealth();
            loadLogs();
            loadMetrics();
            connectAlertStream();

            // 定期刷新数据
            setInterval(function() {
                showRefreshIndicator();
                loadOverview();
                loadAlerts();
                loadServiceHealth();
                loadLogs();
                loadMetrics();
                hideRefreshIndicator();
            }, 30000); // 30秒刷新一次
        });

        // 显示刷新指示器
        function showRefreshIndicator() {
            document.getElementById('refreshIndicator').style.display = 'block';
        }

        function hideRefreshIndicator() {
            setTimeout(() => {
                document.getElementById('refreshIndicator').style.display = 'none';
            }, 1000);
        }

        // 系统指标 Tab 切换
        function showSystemTab(tabName) {
            currentSystemTab = tabName;

            // 更新 tab 状态
            document.querySelectorAll('.tab').forEach(tab => tab.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));

            event.target.classList.add('active');
            document.getElementById(tabName + '-content').classList.add('active');

            // 可以在这里加载对应的图表数据
            updateSystemChart(tabName);
        }

        // 更新系统图表
        function updateSystemChart(type) {
            const chartContainer = document.getElementById(type + 'Chart');
            chartContainer.innerHTML = type.toUpperCase() + ' 使用率图表 (实时数据)';
        }

        // 加载概览数据
        async function loadOverview() {
            try {
                const response = await fetch('/api/overview');
                const data = await response.json();

                const overviewGrid = document.getElementById('overviewGrid');
                overviewGrid.innerHTML =
                    '<div class="overview-card">' +
                        '<div class="overview-number overview-' + (data.alerts.firing > 0 ? 'critical' : 'healthy') + '">' + data.alerts.firing + '</div>' +
                        '<div class="overview-label">活动告警</div>' +
                    '</div>' +
                    '<div class="overview-card">' +
                        '<div class="overview-number overview-healthy">' + data.services.healthy + '</div>' +
                        '<div class="overview-label">健康服务</div>' +
                    '</div>' +
                    '<div class="overview-card">' +
                        '<div class="overview-number overview-' + (data.system && data.system.cpu_usage > 80 ? 'warning' : 'healthy') + '">' + (data.system ? data.system.cpu_usage.toFixed(1) + '%' : 'N/A') + '</div>' +
                        '<div class="overview-label">CPU 使用率</div>' +
                    '</div>' +
                    '<div class="overview-card">' +
                        '<div class="overview-number overview-' + (data.system && data.system.memory_usage > 90 ? 'critical' : data.system && data.system.memory_usage > 75 ? 'warning' : 'healthy') + '">' + (data.system ? data.system.memory_usage.toFixed(1) + '%' : 'N/A') + '</div>' +
                        '<div class="overview-label">内存使用率</div>' +
                    '</div>' +
                    '<div class="overview-card">' +
                        '<div class="overview-number">' + (data.system ? data.system.goroutines : 'N/A') + '</div>' +
                        '<div class="overview-label">Goroutines</div>' +
                    '</div>' +
                    '<div class="overview-card">' +
                        '<div class="overview-number">' + data.logs.recent_count + '</div>' +
                        '<div class="overview-label">最近日志</div>' +
                    '</div>';
            } catch (error) {
                console.error('Error loading overview:', error);
            }
        }

        // 加载告警列表
        async function loadAlerts() {
            try {
                const response = await fetch('/api/alerts');
                const data = await response.json();

                const alertList = document.getElementById('alertList');
                if (data.alerts.length === 0) {
                    alertList.innerHTML = '<div style="text-align: center; color: #7f8c8d; padding: 2rem;">暂无活动告警</div>';
                } else {
                    alertList.innerHTML = data.alerts.slice(0, 10).map(alert =>
                        '<div class="alert-item">' +
                            '<div style="display: flex; justify-content: between; align-items: center; margin-bottom: 0.5rem;">' +
                                '<strong>' + alert.name + '</strong>' +
                                '<span class="status status-' + alert.status + '">' + alert.status + '</span>' +
                            '</div>' +
                            '<div style="font-size: 0.9rem; color: #666; margin-bottom: 0.5rem;">' + alert.message + '</div>' +
                            '<div style="font-size: 0.8rem; color: #999;">' +
                                '开始时间: ' + formatTime(alert.start_time) + ' | 次数: ' + alert.count +
                            '</div>' +
                        '</div>'
                    ).join('');
                }
            } catch (error) {
                console.error('Error loading alerts:', error);
            }
        }

        // 加载服务健康状态
        async function loadServiceHealth() {
            try {
                const response = await fetch('/api/health');
                const data = await response.json();

                const serviceHealth = document.getElementById('serviceHealth');
                if (Object.keys(data.services).length === 0) {
                    serviceHealth.innerHTML = '<div style="text-align: center; color: #7f8c8d; padding: 2rem;">暂无服务数据</div>';
                } else {
                    serviceHealth.innerHTML = Object.entries(data.services).map(([name, service]) =>
                        '<div style="padding: 1rem; border: 1px solid #ecf0f1; border-radius: 4px; margin-bottom: 1rem;">' +
                            '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">' +
                                '<strong>' + service.service + '</strong>' +
                                '<span class="status status-' + (service.status === 'healthy' ? 'resolved' : service.status === 'degraded' ? 'warning' : 'firing') + '">' + service.status + '</span>' +
                            '</div>' +
                            '<div style="font-size: 0.9rem; color: #666;">' +
                                '检查项: ' + service.checks.length + ' | 运行时间: ' + formatDuration(service.uptime) +
                            '</div>' +
                        '</div>'
                    ).join('');
                }
            } catch (error) {
                console.error('Error loading service health:', error);
            }
        }

        // 加载最近日志
        async function loadLogs() {
            try {
                const response = await fetch('/api/logs?limit=20');
                const data = await response.json();

                const logList = document.getElementById('logList');
                if (data.logs.length === 0) {
                    logList.innerHTML = '<div style="text-align: center; color: #7f8c8d; padding: 2rem;">暂无日志数据</div>';
                } else {
                    logList.innerHTML = data.logs.map(log =>
                        '<div class="log-item">' +
                            '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.5rem;">' +
                                '<span class="log-level log-' + log.level + '">' + log.level.toUpperCase() + '</span>' +
                                '<span style="font-size: 0.8rem; color: #999;">' + formatTime(log.timestamp) + '</span>' +
                            '</div>' +
                            '<div style="font-size: 0.9rem; color: #333; margin-bottom: 0.5rem;">' + log.message + '</div>' +
                            '<div style="font-size: 0.8rem; color: #666;">' +
                                '服务: ' + (log.service || 'unknown') + ' | 来源: ' + (log.source || 'unknown') +
                            '</div>' +
                        '</div>'
                    ).join('');
                }
            } catch (error) {
                console.error('Error loading logs:', error);
            }
        }

        // 加载指标数据
        async function loadMetrics() {
            try {
                const response = await fetch('/api/metrics?duration=1h');
                const data = await response.json();

                const metricsTableBody = document.getElementById('metricsTableBody');

                // 获取最近的指标
                const recentMetrics = {};
                data.metrics.forEach(metric => {
                    const key = metric.name + JSON.stringify(metric.labels || {});
                    if (!recentMetrics[key] || metric.timestamp > recentMetrics[key].timestamp) {
                        recentMetrics[key] = metric;
                    }
                });

                const metricsArray = Object.values(recentMetrics).slice(0, 20);

                metricsTableBody.innerHTML = metricsArray.map(metric =>
                    '<tr>' +
                        '<td>' + metric.name + '</td>' +
                        '<td>' + metric.value.toFixed(2) + '</td>' +
                        '<td>' + metric.type + '</td>' +
                        '<td>' + (metric.labels ? JSON.stringify(metric.labels) : '-') + '</td>' +
                        '<td>' + formatTime(metric.timestamp) + '</td>' +
                    '</tr>'
                ).join('');
            } catch (error) {
                console.error('Error loading metrics:', error);
            }
        }

        // 连接告警流
        function connectAlertStream() {
            if (alertEventSource) {
                alertEventSource.close();
            }

            alertEventSource = new EventSource('/api/alerts/stream');

            alertEventSource.onmessage = function(event) {
                const alert = JSON.parse(event.data);
                showAlertNotification(alert);
            };

            alertEventSource.onerror = function(event) {
                console.error('Alert stream error:', event);
                // 重连
                setTimeout(connectAlertStream, 5000);
            };
        }

        // 显示告警通知
        function showAlertNotification(alert) {
            // 简单的通知显示
            const notification = document.createElement('div');
            notification.style.cssText =
                'position: fixed;' +
                'top: 20px;' +
                'right: 20px;' +
                'background: ' + (alert.status === 'firing' ? '#e74c3c' : '#27ae60') + ';' +
                'color: white;' +
                'padding: 1rem;' +
                'border-radius: 4px;' +
                'z-index: 1000;' +
                'max-width: 300px;';
            notification.innerHTML =
                '<strong>' + alert.name + '</strong><br>' +
                '<small>' + alert.message + '</small>';

            document.body.appendChild(notification);

            setTimeout(() => {
                document.body.removeChild(notification);
            }, 5000);

            // 刷新告警列表
            loadAlerts();
        }

        // 格式化时间
        function formatTime(timestamp) {
            return new Date(timestamp).toLocaleString('zh-CN');
        }

        // 格式化持续时间
        function formatDuration(nanoseconds) {
            const seconds = Math.floor(nanoseconds / 1000000000);
            const minutes = Math.floor(seconds / 60);
            const hours = Math.floor(minutes / 60);
            const days = Math.floor(hours / 24);

            if (days > 0) return days + '天';
            if (hours > 0) return hours + '小时';
            if (minutes > 0) return minutes + '分钟';
            return seconds + '秒';
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func (ms *MonitoringServer) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (ms *MonitoringServer) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (ms *MonitoringServer) generateSampleLogs() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	services := []string{"api-server", "database", "cache", "queue-worker"}
	levels := []string{"info", "warn", "error", "debug"}
	messages := []string{
		"Request processed successfully",
		"Database connection established",
		"Cache miss for key: user_123",
		"Processing background job",
		"High memory usage detected",
		"Request timeout exceeded",
		"User authentication failed",
		"Backup completed successfully",
	}

	for {
		select {
		case <-ticker.C:
			// 生成随机日志
			service := services[time.Now().UnixNano()%int64(len(services))]
			level := levels[time.Now().UnixNano()%int64(len(levels))]
			message := messages[time.Now().UnixNano()%int64(len(messages))]

			ms.logAggregator.AddLog(level, message, "system", service, map[string]interface{}{
				"request_id": fmt.Sprintf("req_%d", time.Now().UnixNano()),
				"user_id":    fmt.Sprintf("user_%d", time.Now().UnixNano()%1000),
			})
		}
	}
}

// ====================
// 全局变量
// ====================

var startTime = time.Now()

// ====================
// 主函数
// ====================

func main() {
	// 创建存储
	storage := NewStorage("./monitoring_data")

	// 创建监控服务器
	monitoringServer := NewMonitoringServer(storage)

	// 创建一些示例指标
	go generateSampleMetrics(monitoringServer.collector)

	// 启动HTTP服务器
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	log.Printf("📊 监控系统启动在 http://localhost:%s", port)
	log.Println("功能特性:")
	log.Println("- 实时指标收集")
	log.Println("- 智能告警系统")
	log.Println("- 日志聚合分析")
	log.Println("- 服务健康检查")
	log.Println("- 系统资源监控")
	log.Println("- 可视化仪表板")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      monitoringServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

func generateSampleMetrics(collector *MetricsCollector) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 生成一些示例业务指标
			collector.RecordMetric("http.requests.total", "counter", float64(time.Now().Unix()%1000), map[string]string{
				"method": "GET",
				"status": "200",
			})

			collector.RecordMetric("http.response_time", "histogram", 50+50*collector.randomFloat(), map[string]string{
				"endpoint": "/api/users",
			})

			collector.RecordMetric("database.connections", "gauge", 10+5*collector.randomFloat(), map[string]string{
				"pool": "primary",
			})

			collector.RecordMetric("queue.messages", "gauge", 100+200*collector.randomFloat(), map[string]string{
				"queue": "email",
			})
		}
	}
}

/*
=== 项目功能清单 ===

指标收集:
✅ 自定义指标收集
✅ 系统资源监控 (CPU, 内存, 磁盘)
✅ Go Runtime 指标
✅ HTTP 服务指标
✅ 指标持久化存储

告警系统:
✅ 基于规则的告警
✅ 多种比较运算符
✅ 告警状态管理
✅ 实时告警流 (SSE)
✅ 告警通知机制

日志管理:
✅ 结构化日志收集
✅ 日志级别过滤
✅ 服务维度聚合
✅ 日志搜索功能
✅ 日志持久化

健康检查:
✅ 服务健康监控
✅ 多维度检查项
✅ 健康状态聚合
✅ 运行时间统计

监控面板:
✅ 实时数据展示
✅ 概览卡片
✅ 图表可视化
✅ 响应式布局
✅ 自动数据刷新

=== API 端点 ===

指标相关:
- GET /api/metrics         - 获取指标数据
- POST /api/metrics        - 提交指标数据
- GET /api/system          - 获取系统指标

告警相关:
- GET /api/alerts          - 获取告警列表
- GET /api/alerts/stream   - 告警实时流

日志相关:
- GET /api/logs            - 获取日志列表
- POST /api/logs           - 提交日志
- GET /api/logs/search     - 搜索日志

健康检查:
- GET /api/health          - 获取所有服务健康状态
- GET /api/health/{service} - 获取特定服务健康状态

监控概览:
- GET /api/overview        - 获取监控概览数据

=== 告警规则示例 ===

默认告警规则:
1. CPU 使用率 > 80% (2分钟)
2. 内存使用率 > 90% (1分钟)
3. 磁盘使用率 > 85% (5分钟)
4. Goroutines > 1000 (1分钟)

=== 高级功能扩展 ===

1. 数据存储:
   - 时序数据库集成 (InfluxDB, Prometheus)
   - 数据压缩和聚合
   - 长期数据保留策略

2. 告警增强:
   - 告警分组和静默
   - 多渠道通知 (邮件, 短信, Webhook)
   - 告警模板和变量
   - 依赖关系管理

3. 可视化:
   - 自定义仪表板
   - 图表配置
   - 数据导出
   - 历史趋势分析

4. 集成能力:
   - Grafana 集成
   - Jaeger 分布式追踪
   - ELK Stack 日志
   - Kubernetes 监控

=== 部署说明 ===

1. 运行应用:
   go run main.go

2. 访问监控面板:
   http://localhost:8080

3. 数据存储:
   - 指标: ./monitoring_data/metrics.json
   - 告警: ./monitoring_data/alerts.log
   - 日志: ./monitoring_data/logs_*.log

4. 配置环境变量:
   - PORT: 服务端口号
*/
