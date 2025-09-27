package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// 安全随机数生成函数
func secureRandomInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// G115安全修复：确保转换不会溢出
		fallback := time.Now().UnixNano() % int64(max)
		if fallback > int64(^uint(0)>>1) {
			fallback = fallback % int64(^uint(0)>>1)
		}
		return int(fallback)
	}
	// G115安全修复：检查int64到int的安全转换
	result := n.Int64()
	if result > int64(^uint(0)>>1) {
		result = result % int64(max)
	}
	return int(result)
}

func secureRandomFloat32() float32 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<24))
	if err != nil {
		// 安全fallback：使用时间戳
		return float32(time.Now().UnixNano()%1000) / 1000.0
	}
	return float32(n.Int64()) / float32(1<<24)
}

func secureRandomFloat64() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		// 安全fallback：使用时间戳
		return float64(time.Now().UnixNano()%1000) / 1000.0
	}
	return float64(n.Int64()) / float64(1<<53)
}

/*
微服务架构 - 监控与追踪练习

本练习涵盖微服务架构中的监控与追踪系统，包括：
1. 分布式链路追踪（Distributed Tracing）
2. 指标收集与监控（Metrics Collection）
3. 日志聚合与分析（Log Aggregation）
4. 健康检查系统（Health Check）
5. 性能监控（Performance Monitoring）
6. 告警系统（Alerting）
7. 可视化仪表板（Dashboard）
8. APM应用性能监控

主要概念：
- 链路追踪和Span
- 指标采集和时序数据库
- 日志结构化和聚合
- 监控告警规则
- SLI/SLO/SLA
- 可观测性三大支柱
*/

// === 分布式链路追踪 ===

// TraceContext 追踪上下文
type TraceContext struct {
	TraceID  string `json:"trace_id"`
	SpanID   string `json:"span_id"`
	ParentID string `json:"parent_id"`
}

// Span 追踪片段
type Span struct {
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	ParentSpanID  string                 `json:"parent_span_id"`
	OperationName string                 `json:"operation_name"`
	ServiceName   string                 `json:"service_name"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	Status        string                 `json:"status"` // success, error
	Tags          map[string]interface{} `json:"tags"`
	Logs          []SpanLog              `json:"logs"`
	Error         string                 `json:"error,omitempty"`
}

// SpanLog 追踪日志
type SpanLog struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields"`
}

// Tracer 追踪器接口
type Tracer interface {
	StartSpan(operationName string, parent *TraceContext) *Span
	FinishSpan(span *Span)
	InjectContext(span *Span) *TraceContext
	ExtractContext(traceID, spanID, parentID string) *TraceContext
}

// InMemoryTracer 内存追踪器实现
type InMemoryTracer struct {
	spans   map[string]*Span
	traces  map[string][]*Span
	mutex   sync.RWMutex
	sampler TraceSampler
}

// TraceSampler 采样器
type TraceSampler interface {
	ShouldSample(traceID string) bool
}

// SimpleSampler 简单采样器
type SimpleSampler struct {
	rate float64 // 采样率 0.0-1.0
}

func (s *SimpleSampler) ShouldSample(traceID string) bool {
	return secureRandomFloat64() < s.rate
}

func NewInMemoryTracer(sampleRate float64) *InMemoryTracer {
	return &InMemoryTracer{
		spans:   make(map[string]*Span),
		traces:  make(map[string][]*Span),
		sampler: &SimpleSampler{rate: sampleRate},
	}
}

func (t *InMemoryTracer) StartSpan(operationName string, parent *TraceContext) *Span {
	span := &Span{
		SpanID:        uuid.New().String(),
		OperationName: operationName,
		StartTime:     time.Now(),
		Tags:          make(map[string]interface{}),
		Logs:          make([]SpanLog, 0),
		Status:        "active",
	}

	if parent != nil {
		span.TraceID = parent.TraceID
		span.ParentSpanID = parent.SpanID
	} else {
		span.TraceID = uuid.New().String()
	}

	// 采样决策
	if !t.sampler.ShouldSample(span.TraceID) {
		span.Tags["sampled"] = false
		return span
	}

	span.Tags["sampled"] = true

	t.mutex.Lock()
	t.spans[span.SpanID] = span
	t.mutex.Unlock()

	return span
}

func (t *InMemoryTracer) FinishSpan(span *Span) {
	if span.Tags["sampled"] != true {
		return
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)

	if span.Status == "active" {
		span.Status = "success"
	}

	t.mutex.Lock()
	if t.traces[span.TraceID] == nil {
		t.traces[span.TraceID] = make([]*Span, 0)
	}
	t.traces[span.TraceID] = append(t.traces[span.TraceID], span)
	t.mutex.Unlock()

	log.Printf("完成Span: %s, 操作: %s, 耗时: %v", span.SpanID, span.OperationName, span.Duration)
}

func (t *InMemoryTracer) InjectContext(span *Span) *TraceContext {
	return &TraceContext{
		TraceID:  span.TraceID,
		SpanID:   span.SpanID,
		ParentID: span.ParentSpanID,
	}
}

func (t *InMemoryTracer) ExtractContext(traceID, spanID, parentID string) *TraceContext {
	return &TraceContext{
		TraceID:  traceID,
		SpanID:   spanID,
		ParentID: parentID,
	}
}

func (t *InMemoryTracer) GetTrace(traceID string) ([]*Span, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	spans, exists := t.traces[traceID]
	if !exists {
		return nil, fmt.Errorf("追踪不存在: %s", traceID)
	}

	return spans, nil
}

// === 指标收集系统 ===

// Metric 指标接口
type Metric interface {
	GetName() string
	GetType() string
	GetValue() interface{}
	GetLabels() map[string]string
	GetTimestamp() time.Time
}

// Counter 计数器指标
type Counter struct {
	Name      string            `json:"name"`
	Value     int64             `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
	mutex     sync.Mutex
}

func NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		Name:      name,
		Value:     0,
		Labels:    labels,
		Timestamp: time.Now(),
	}
}

func (c *Counter) Inc() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Value++
	c.Timestamp = time.Now()
}

func (c *Counter) Add(delta int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Value += delta
	c.Timestamp = time.Now()
}

func (c *Counter) GetName() string              { return c.Name }
func (c *Counter) GetType() string              { return "counter" }
func (c *Counter) GetValue() interface{}        { return c.Value }
func (c *Counter) GetLabels() map[string]string { return c.Labels }
func (c *Counter) GetTimestamp() time.Time      { return c.Timestamp }

// Gauge 仪表盘指标
type Gauge struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
	mutex     sync.Mutex
}

func NewGauge(name string, labels map[string]string) *Gauge {
	return &Gauge{
		Name:      name,
		Value:     0,
		Labels:    labels,
		Timestamp: time.Now(),
	}
}

func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Value = value
	g.Timestamp = time.Now()
}

func (g *Gauge) Inc() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Value++
	g.Timestamp = time.Now()
}

func (g *Gauge) Dec() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.Value--
	g.Timestamp = time.Now()
}

func (g *Gauge) GetName() string              { return g.Name }
func (g *Gauge) GetType() string              { return "gauge" }
func (g *Gauge) GetValue() interface{}        { return g.Value }
func (g *Gauge) GetLabels() map[string]string { return g.Labels }
func (g *Gauge) GetTimestamp() time.Time      { return g.Timestamp }

// Histogram 直方图指标
type Histogram struct {
	Name      string            `json:"name"`
	Buckets   map[float64]int64 `json:"buckets"`
	Count     int64             `json:"count"`
	Sum       float64           `json:"sum"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
	mutex     sync.Mutex
}

func NewHistogram(name string, buckets []float64, labels map[string]string) *Histogram {
	h := &Histogram{
		Name:      name,
		Buckets:   make(map[float64]int64),
		Count:     0,
		Sum:       0,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	for _, bucket := range buckets {
		h.Buckets[bucket] = 0
	}

	return h
}

func (h *Histogram) Observe(value float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.Count++
	h.Sum += value
	h.Timestamp = time.Now()

	for bucket := range h.Buckets {
		if value <= bucket {
			h.Buckets[bucket]++
		}
	}
}

func (h *Histogram) GetName() string              { return h.Name }
func (h *Histogram) GetType() string              { return "histogram" }
func (h *Histogram) GetValue() interface{}        { return h }
func (h *Histogram) GetLabels() map[string]string { return h.Labels }
func (h *Histogram) GetTimestamp() time.Time      { return h.Timestamp }

// MetricsRegistry 指标注册表
type MetricsRegistry struct {
	metrics map[string]Metric
	mutex   sync.RWMutex
}

func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics: make(map[string]Metric),
	}
}

func (mr *MetricsRegistry) Register(metric Metric) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	key := fmt.Sprintf("%s_%s", metric.GetName(), labelsToString(metric.GetLabels()))
	mr.metrics[key] = metric
}

func (mr *MetricsRegistry) GetMetric(name string, labels map[string]string) Metric {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	key := fmt.Sprintf("%s_%s", name, labelsToString(labels))
	return mr.metrics[key]
}

func (mr *MetricsRegistry) GetAllMetrics() map[string]Metric {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	result := make(map[string]Metric)
	for k, v := range mr.metrics {
		result[k] = v
	}

	return result
}

func labelsToString(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("{%s}", fmt.Sprint(parts))
}

// === 日志聚合系统 ===

// LogLevel 日志级别
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Fields    map[string]interface{} `json:"fields"`
	Source    string                 `json:"source"` // 日志来源
}

// Logger 结构化日志记录器
type Logger struct {
	service   string
	tracer    *InMemoryTracer
	collector LogCollector
}

type LogCollector interface {
	Collect(entry *LogEntry)
}

func NewLogger(service string, tracer *InMemoryTracer, collector LogCollector) *Logger {
	return &Logger{
		service:   service,
		tracer:    tracer,
		collector: collector,
	}
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, traceCtx *TraceContext) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Service:   l.service,
		Fields:    fields,
		Source:    "application",
	}

	if traceCtx != nil {
		entry.TraceID = traceCtx.TraceID
		entry.SpanID = traceCtx.SpanID
	}

	if l.collector != nil {
		l.collector.Collect(entry)
	}

	// 同时输出到标准输出
	log.Printf("[%s][%s] %s %v", l.service, level, message, fields)
}

func (l *Logger) Debug(message string, fields map[string]interface{}) {
	l.log(DEBUG, message, fields, nil)
}

func (l *Logger) Info(message string, fields map[string]interface{}) {
	l.log(INFO, message, fields, nil)
}

func (l *Logger) Warn(message string, fields map[string]interface{}) {
	l.log(WARN, message, fields, nil)
}

func (l *Logger) Error(message string, fields map[string]interface{}) {
	l.log(ERROR, message, fields, nil)
}

func (l *Logger) WithTrace(traceCtx *TraceContext) *TraceLogger {
	return &TraceLogger{
		logger:   l,
		traceCtx: traceCtx,
	}
}

// TraceLogger 带追踪上下文的日志记录器
type TraceLogger struct {
	logger   *Logger
	traceCtx *TraceContext
}

func (tl *TraceLogger) Debug(message string, fields map[string]interface{}) {
	tl.logger.log(DEBUG, message, fields, tl.traceCtx)
}

func (tl *TraceLogger) Info(message string, fields map[string]interface{}) {
	tl.logger.log(INFO, message, fields, tl.traceCtx)
}

func (tl *TraceLogger) Warn(message string, fields map[string]interface{}) {
	tl.logger.log(WARN, message, fields, tl.traceCtx)
}

func (tl *TraceLogger) Error(message string, fields map[string]interface{}) {
	tl.logger.log(ERROR, message, fields, tl.traceCtx)
}

// 内存日志收集器
type InMemoryLogCollector struct {
	logs    []LogEntry
	mutex   sync.RWMutex
	maxSize int
}

func NewInMemoryLogCollector(maxSize int) *InMemoryLogCollector {
	return &InMemoryLogCollector{
		logs:    make([]LogEntry, 0),
		maxSize: maxSize,
	}
}

func (ilc *InMemoryLogCollector) Collect(entry *LogEntry) {
	ilc.mutex.Lock()
	defer ilc.mutex.Unlock()

	ilc.logs = append(ilc.logs, *entry)

	// 保持最大数量限制
	if len(ilc.logs) > ilc.maxSize {
		ilc.logs = ilc.logs[1:]
	}
}

func (ilc *InMemoryLogCollector) GetLogs(limit int) []LogEntry {
	ilc.mutex.RLock()
	defer ilc.mutex.RUnlock()

	if limit <= 0 || limit > len(ilc.logs) {
		limit = len(ilc.logs)
	}

	return ilc.logs[len(ilc.logs)-limit:]
}

func (ilc *InMemoryLogCollector) QueryLogs(service, level, traceID string, limit int) []LogEntry {
	ilc.mutex.RLock()
	defer ilc.mutex.RUnlock()

	var result []LogEntry
	count := 0

	for i := len(ilc.logs) - 1; i >= 0 && count < limit; i-- {
		entry := ilc.logs[i]

		if service != "" && entry.Service != service {
			continue
		}

		if level != "" && string(entry.Level) != level {
			continue
		}

		if traceID != "" && entry.TraceID != traceID {
			continue
		}

		result = append([]LogEntry{entry}, result...)
		count++
	}

	return result
}

// === 健康检查系统 ===

// HealthCheck 健康检查接口
type HealthCheck interface {
	Name() string
	Check() HealthResult
}

// HealthResult 健康检查结果
type HealthResult struct {
	Status    string                 `json:"status"` // healthy, unhealthy, degraded
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	checks map[string]HealthCheck
	mutex  sync.RWMutex
}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]HealthCheck),
	}
}

func (hc *HealthChecker) Register(check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[check.Name()] = check
}

func (hc *HealthChecker) CheckAll() map[string]HealthResult {
	hc.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for k, v := range hc.checks {
		checks[k] = v
	}
	hc.mutex.RUnlock()

	results := make(map[string]HealthResult)
	var wg sync.WaitGroup

	for name, check := range checks {
		wg.Add(1)
		go func(n string, c HealthCheck) {
			defer wg.Done()
			start := time.Now()
			result := c.Check()
			result.Duration = time.Since(start)
			results[n] = result
		}(name, check)
	}

	wg.Wait()
	return results
}

// 数据库健康检查
type DatabaseHealthCheck struct {
	name string
}

func NewDatabaseHealthCheck(name string) *DatabaseHealthCheck {
	return &DatabaseHealthCheck{name: name}
}

func (dbhc *DatabaseHealthCheck) Name() string {
	return dbhc.name
}

func (dbhc *DatabaseHealthCheck) Check() HealthResult {
	// 模拟数据库健康检查
	time.Sleep(time.Duration(secureRandomInt(100)) * time.Millisecond)

	if secureRandomFloat32() < 0.1 { // 10%的概率不健康
		return HealthResult{
			Status:    "unhealthy",
			Message:   "数据库连接失败",
			Details:   map[string]interface{}{"error": "connection timeout"},
			Timestamp: time.Now(),
		}
	}

	return HealthResult{
		Status:    "healthy",
		Message:   "数据库连接正常",
		Details:   map[string]interface{}{"connections": 10, "max_connections": 100},
		Timestamp: time.Now(),
	}
}

// 外部服务健康检查
type ExternalServiceHealthCheck struct {
	name string
	url  string
}

func NewExternalServiceHealthCheck(name, url string) *ExternalServiceHealthCheck {
	return &ExternalServiceHealthCheck{name: name, url: url}
}

func (eshc *ExternalServiceHealthCheck) Name() string {
	return eshc.name
}

func (eshc *ExternalServiceHealthCheck) Check() HealthResult {
	// 模拟外部服务健康检查
	time.Sleep(time.Duration(secureRandomInt(200)) * time.Millisecond)

	if secureRandomFloat32() < 0.05 { // 5%的概率不健康
		return HealthResult{
			Status:    "unhealthy",
			Message:   "外部服务不可用",
			Details:   map[string]interface{}{"url": eshc.url, "error": "service unavailable"},
			Timestamp: time.Now(),
		}
	}

	return HealthResult{
		Status:    "healthy",
		Message:   "外部服务正常",
		Details:   map[string]interface{}{"url": eshc.url, "response_time": "50ms"},
		Timestamp: time.Now(),
	}
}

// === 告警系统 ===

// AlertRule 告警规则
type AlertRule struct {
	Name        string                 `json:"name"`
	MetricName  string                 `json:"metric_name"`
	Condition   string                 `json:"condition"` // gt, lt, eq, gte, lte
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]interface{} `json:"annotations"`
	Enabled     bool                   `json:"enabled"`
}

// Alert 告警
type Alert struct {
	Rule        *AlertRule `json:"rule"`
	Value       float64    `json:"value"`
	Status      string     `json:"status"` // firing, resolved
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      time.Time  `json:"ends_at,omitempty"`
	Fingerprint string     `json:"fingerprint"`
}

// AlertManager 告警管理器
type AlertManager struct {
	rules     []*AlertRule
	alerts    map[string]*Alert
	registry  *MetricsRegistry
	mutex     sync.RWMutex
	notifiers []AlertNotifier
}

type AlertNotifier interface {
	Notify(alert *Alert) error
}

func NewAlertManager(registry *MetricsRegistry) *AlertManager {
	am := &AlertManager{
		rules:     make([]*AlertRule, 0),
		alerts:    make(map[string]*Alert),
		registry:  registry,
		notifiers: make([]AlertNotifier, 0),
	}

	// 启动告警评估器
	go am.startEvaluator()

	return am
}

func (am *AlertManager) AddRule(rule *AlertRule) {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	am.rules = append(am.rules, rule)
}

func (am *AlertManager) AddNotifier(notifier AlertNotifier) {
	am.notifiers = append(am.notifiers, notifier)
}

func (am *AlertManager) startEvaluator() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		am.evaluateRules()
	}
}

func (am *AlertManager) evaluateRules() {
	am.mutex.RLock()
	rules := make([]*AlertRule, len(am.rules))
	copy(rules, am.rules)
	am.mutex.RUnlock()

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		metric := am.registry.GetMetric(rule.MetricName, rule.Labels)
		if metric == nil {
			continue
		}

		value := am.extractNumericValue(metric.GetValue())
		shouldFire := am.evaluateCondition(value, rule.Condition, rule.Threshold)

		fingerprint := am.generateFingerprint(rule)

		am.mutex.Lock()
		existingAlert, exists := am.alerts[fingerprint]

		if shouldFire {
			if !exists {
				// 新告警
				alert := &Alert{
					Rule:        rule,
					Value:       value,
					Status:      "firing",
					StartsAt:    time.Now(),
					Fingerprint: fingerprint,
				}

				am.alerts[fingerprint] = alert

				// 发送通知
				for _, notifier := range am.notifiers {
					go notifier.Notify(alert)
				}

				log.Printf("触发告警: %s, 值: %f, 阈值: %f", rule.Name, value, rule.Threshold)
			} else if existingAlert.Status == "resolved" {
				// 重新触发
				existingAlert.Status = "firing"
				existingAlert.StartsAt = time.Now()
				existingAlert.EndsAt = time.Time{}

				for _, notifier := range am.notifiers {
					go notifier.Notify(existingAlert)
				}
			}
		} else if exists && existingAlert.Status == "firing" {
			// 告警恢复
			existingAlert.Status = "resolved"
			existingAlert.EndsAt = time.Now()

			for _, notifier := range am.notifiers {
				go notifier.Notify(existingAlert)
			}

			log.Printf("告警恢复: %s", rule.Name)
		}

		am.mutex.Unlock()
	}
}

func (am *AlertManager) evaluateCondition(value float64, condition string, threshold float64) bool {
	switch condition {
	case "gt":
		return value > threshold
	case "gte":
		return value >= threshold
	case "lt":
		return value < threshold
	case "lte":
		return value <= threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}

func (am *AlertManager) extractNumericValue(value interface{}) float64 {
	switch v := value.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case int:
		return float64(v)
	case float32:
		return float64(v)
	default:
		return 0
	}
}

func (am *AlertManager) generateFingerprint(rule *AlertRule) string {
	return fmt.Sprintf("%s_%s_%s", rule.Name, rule.MetricName, labelsToString(rule.Labels))
}

func (am *AlertManager) GetAlerts() map[string]*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make(map[string]*Alert)
	for k, v := range am.alerts {
		result[k] = v
	}

	return result
}

// 控制台告警通知器
type ConsoleNotifier struct{}

func (cn *ConsoleNotifier) Notify(alert *Alert) error {
	status := "🔥 FIRING"
	if alert.Status == "resolved" {
		status = "✅ RESOLVED"
	}

	log.Printf("[ALERT %s] %s - %s: 当前值 %.2f, 阈值 %.2f",
		status, alert.Rule.Name, alert.Rule.MetricName, alert.Value, alert.Rule.Threshold)
	return nil
}

// === 监控仪表板 ===

type MonitoringDashboard struct {
	tracer        *InMemoryTracer
	registry      *MetricsRegistry
	logCollector  *InMemoryLogCollector
	healthChecker *HealthChecker
	alertManager  *AlertManager
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]bool
	mutex         sync.Mutex
}

func NewMonitoringDashboard(tracer *InMemoryTracer, registry *MetricsRegistry,
	logCollector *InMemoryLogCollector, healthChecker *HealthChecker,
	alertManager *AlertManager) *MonitoringDashboard {

	return &MonitoringDashboard{
		tracer:        tracer,
		registry:      registry,
		logCollector:  logCollector,
		healthChecker: healthChecker,
		alertManager:  alertManager,
		upgrader:      websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		clients:       make(map[*websocket.Conn]bool),
	}
}

// 获取系统指标
func (md *MonitoringDashboard) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc":        m.Alloc,
			"total_alloc":  m.TotalAlloc,
			"sys":          m.Sys,
			"heap_alloc":   m.HeapAlloc,
			"heap_sys":     m.HeapSys,
			"heap_objects": m.HeapObjects,
		},
		"runtime": map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"gc_cycles":  m.NumGC,
			"cpu_count":  runtime.NumCPU(),
		},
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// 获取应用指标
func (md *MonitoringDashboard) GetApplicationMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := md.registry.GetAllMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// 获取追踪信息
func (md *MonitoringDashboard) GetTrace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceID := vars["traceId"]

	spans, err := md.tracer.GetTrace(traceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spans)
}

// 获取日志
func (md *MonitoringDashboard) GetLogs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	level := r.URL.Query().Get("level")
	traceID := r.URL.Query().Get("trace_id")
	limit := 100

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := fmt.Sscanf(l, "%d", &limit); parsed != 1 || err != nil {
			limit = 100
		}
	}

	logs := md.logCollector.QueryLogs(service, level, traceID, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// 获取健康检查结果
func (md *MonitoringDashboard) GetHealth(w http.ResponseWriter, r *http.Request) {
	results := md.healthChecker.CheckAll()

	overall := "healthy"
	for _, result := range results {
		if result.Status != "healthy" {
			overall = "unhealthy"
			break
		}
	}

	response := map[string]interface{}{
		"overall": overall,
		"checks":  results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 获取告警
func (md *MonitoringDashboard) GetAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := md.alertManager.GetAlerts()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// WebSocket实时数据推送
func (md *MonitoringDashboard) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := md.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	md.mutex.Lock()
	md.clients[conn] = true
	md.mutex.Unlock()

	defer func() {
		md.mutex.Lock()
		delete(md.clients, conn)
		md.mutex.Unlock()
	}()

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (md *MonitoringDashboard) startDataBroadcaster() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		md.broadcastData()
	}
}

func (md *MonitoringDashboard) broadcastData() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	data := map[string]interface{}{
		"type": "metrics",
		"data": map[string]interface{}{
			"memory": map[string]interface{}{
				"alloc": m.Alloc,
				"sys":   m.Sys,
			},
			"runtime": map[string]interface{}{
				"goroutines": runtime.NumGoroutine(),
			},
			"timestamp": time.Now(),
		},
	}

	message, _ := json.Marshal(data)

	md.mutex.Lock()
	defer md.mutex.Unlock()

	for conn := range md.clients {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			conn.Close()
			delete(md.clients, conn)
		}
	}
}

func main() {
	// 创建监控组件
	tracer := NewInMemoryTracer(1.0) // 100%采样
	registry := NewMetricsRegistry()
	logCollector := NewInMemoryLogCollector(10000)
	healthChecker := NewHealthChecker()
	alertManager := NewAlertManager(registry)

	// 创建示例指标
	requestCounter := NewCounter("http_requests_total", map[string]string{"method": "GET", "endpoint": "/api"})
	responseTime := NewHistogram("http_request_duration_seconds", []float64{0.1, 0.5, 1.0, 2.0, 5.0}, map[string]string{"method": "GET"})
	activeConnections := NewGauge("active_connections", map[string]string{"service": "api-gateway"})

	registry.Register(requestCounter)
	registry.Register(responseTime)
	registry.Register(activeConnections)

	// 注册健康检查
	healthChecker.Register(NewDatabaseHealthCheck("main-database"))
	healthChecker.Register(NewExternalServiceHealthCheck("payment-service", "http://payment.example.com"))

	// 添加告警规则
	alertManager.AddRule(&AlertRule{
		Name:       "HighMemoryUsage",
		MetricName: "active_connections",
		Condition:  "gt",
		Threshold:  50,
		Duration:   time.Minute,
		Labels:     map[string]string{"service": "api-gateway"},
		Enabled:    true,
	})

	alertManager.AddNotifier(&ConsoleNotifier{})

	// 创建日志记录器
	logger := NewLogger("monitoring-service", tracer, logCollector)

	// 创建监控仪表板
	dashboard := NewMonitoringDashboard(tracer, registry, logCollector, healthChecker, alertManager)

	// 启动数据广播
	go dashboard.startDataBroadcaster()

	// 创建路由器
	router := mux.NewRouter()

	// 监控API
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/metrics/system", dashboard.GetSystemMetrics).Methods("GET")
	api.HandleFunc("/metrics/application", dashboard.GetApplicationMetrics).Methods("GET")
	api.HandleFunc("/traces/{traceId}", dashboard.GetTrace).Methods("GET")
	api.HandleFunc("/logs", dashboard.GetLogs).Methods("GET")
	api.HandleFunc("/health", dashboard.GetHealth).Methods("GET")
	api.HandleFunc("/alerts", dashboard.GetAlerts).Methods("GET")

	// WebSocket
	router.HandleFunc("/ws", dashboard.HandleWebSocket)

	// 模拟业务流量和监控数据
	go func() {
		for {
			time.Sleep(time.Duration(secureRandomInt(2000)) * time.Millisecond)

			// 创建追踪
			span := tracer.StartSpan("api_request", nil)
			span.ServiceName = "api-service"
			span.Tags["http.method"] = "GET"
			span.Tags["http.url"] = "/api/users"

			logger.WithTrace(tracer.InjectContext(span)).Info("处理API请求", map[string]interface{}{
				"method": "GET",
				"path":   "/api/users",
				"user":   "alice",
			})

			// 模拟处理时间
			processingTime := time.Duration(secureRandomInt(1000)) * time.Millisecond
			time.Sleep(processingTime)

			// 更新指标
			requestCounter.Inc()
			responseTime.Observe(processingTime.Seconds())
			activeConnections.Set(float64(secureRandomInt(100)))

			// 偶尔产生错误
			if secureRandomFloat32() < 0.1 {
				span.Status = "error"
				span.Error = "模拟错误"
				logger.WithTrace(tracer.InjectContext(span)).Error("API请求失败", map[string]interface{}{
					"error": "database connection failed",
				})
			}

			tracer.FinishSpan(span)
		}
	}()

	fmt.Println("=== 监控与追踪系统启动 ===")
	fmt.Println("服务端点:")
	fmt.Println("  监控API:    http://localhost:8080")
	fmt.Println("  实时数据:   ws://localhost:8080/ws")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  GET  /api/metrics/system      - 系统指标")
	fmt.Println("  GET  /api/metrics/application - 应用指标")
	fmt.Println("  GET  /api/traces/{traceId}    - 链路追踪")
	fmt.Println("  GET  /api/logs                - 日志查询")
	fmt.Println("  GET  /api/health              - 健康检查")
	fmt.Println("  GET  /api/alerts              - 告警信息")
	fmt.Println()
	fmt.Println("示例请求:")
	fmt.Println("  # 查看系统指标")
	fmt.Println("  curl http://localhost:8080/api/metrics/system")
	fmt.Println()
	fmt.Println("  # 查看应用指标")
	fmt.Println("  curl http://localhost:8080/api/metrics/application")
	fmt.Println()
	fmt.Println("  # 查询日志")
	fmt.Println("  curl 'http://localhost:8080/api/logs?service=api-service&level=INFO&limit=10'")
	fmt.Println()
	fmt.Println("  # 健康检查")
	fmt.Println("  curl http://localhost:8080/api/health")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

/*
练习任务：

1. 基础练习：
   - 实现更多指标类型（Summary、Rate等）
   - 添加采样策略优化
   - 实现日志轮转和压缩
   - 添加配置热重载

2. 中级练习：
   - 集成Prometheus和Grafana
   - 实现Jaeger追踪集成
   - 添加ELK日志栈集成
   - 实现自定义指标导出器

3. 高级练习：
   - 实现分布式追踪采样策略
   - 添加机器学习异常检测
   - 实现智能告警降噪
   - 集成APM工具

4. 性能优化：
   - 实现指标聚合和预计算
   - 添加数据压缩和批处理
   - 优化内存使用和GC
   - 实现异步数据处理

5. 可视化和分析：
   - 实现实时仪表板
   - 添加链路图可视化
   - 实现服务依赖图
   - 添加性能分析报告

监控架构图：
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   应用服务   │────│  监控代理    │────│  监控中心    │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  指标收集    │    │  日志聚合    │    │  告警管理    │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  链路追踪    │    │  健康检查    │    │  可视化     │
└─────────────┘    └─────────────┘    └─────────────┘

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/gorilla/websocket
   go get github.com/google/uuid

2. 运行程序：go run main.go

扩展建议：
- 集成时序数据库（InfluxDB、Prometheus）
- 实现多租户监控隔离
- 添加成本监控和优化
- 实现监控即代码（MaC）
*/
