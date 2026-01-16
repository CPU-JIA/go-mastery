package perfutil

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType represents the type of metric.
type MetricType int

const (
	// MetricCounter is a monotonically increasing counter.
	MetricCounter MetricType = iota
	// MetricGauge is a value that can go up or down.
	MetricGauge
	// MetricHistogram tracks the distribution of values.
	MetricHistogram
	// MetricSummary tracks quantiles of values.
	MetricSummary
)

// String returns the string representation of MetricType.
func (m MetricType) String() string {
	switch m {
	case MetricCounter:
		return "counter"
	case MetricGauge:
		return "gauge"
	case MetricHistogram:
		return "histogram"
	case MetricSummary:
		return "summary"
	default:
		return "unknown"
	}
}

// Counter is a thread-safe monotonically increasing counter.
type Counter struct {
	name   string
	value  atomic.Int64
	labels map[string]string
}

// NewCounter creates a new counter with the given name.
func NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		name:   name,
		labels: labels,
	}
}

// Inc increments the counter by 1.
func (c *Counter) Inc() {
	c.value.Add(1)
}

// Add adds the given value to the counter.
func (c *Counter) Add(delta int64) {
	if delta < 0 {
		return // Counters can only increase
	}
	c.value.Add(delta)
}

// Value returns the current counter value.
func (c *Counter) Value() int64 {
	return c.value.Load()
}

// Name returns the counter name.
func (c *Counter) Name() string {
	return c.name
}

// Labels returns the counter labels.
func (c *Counter) Labels() map[string]string {
	return c.labels
}

// Reset resets the counter to zero.
func (c *Counter) Reset() {
	c.value.Store(0)
}

// Gauge is a thread-safe value that can go up or down.
type Gauge struct {
	name   string
	value  atomic.Int64
	labels map[string]string
}

// NewGauge creates a new gauge with the given name.
func NewGauge(name string, labels map[string]string) *Gauge {
	return &Gauge{
		name:   name,
		labels: labels,
	}
}

// Set sets the gauge to the given value.
func (g *Gauge) Set(value int64) {
	g.value.Store(value)
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc() {
	g.value.Add(1)
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec() {
	g.value.Add(-1)
}

// Add adds the given value to the gauge.
func (g *Gauge) Add(delta int64) {
	g.value.Add(delta)
}

// Sub subtracts the given value from the gauge.
func (g *Gauge) Sub(delta int64) {
	g.value.Add(-delta)
}

// Value returns the current gauge value.
func (g *Gauge) Value() int64 {
	return g.value.Load()
}

// Name returns the gauge name.
func (g *Gauge) Name() string {
	return g.name
}

// Labels returns the gauge labels.
func (g *Gauge) Labels() map[string]string {
	return g.labels
}

// Histogram tracks the distribution of values.
type Histogram struct {
	name    string
	labels  map[string]string
	buckets []float64
	counts  []atomic.Uint64
	sum     atomic.Uint64
	count   atomic.Uint64
	mu      sync.RWMutex
}

// NewHistogram creates a new histogram with the given buckets.
func NewHistogram(name string, labels map[string]string, buckets []float64) *Histogram {
	// Sort buckets
	sortedBuckets := make([]float64, len(buckets))
	copy(sortedBuckets, buckets)
	sort.Float64s(sortedBuckets)

	return &Histogram{
		name:    name,
		labels:  labels,
		buckets: sortedBuckets,
		counts:  make([]atomic.Uint64, len(sortedBuckets)+1), // +1 for +Inf bucket
	}
}

// DefaultHistogramBuckets returns default histogram buckets for latency measurements.
func DefaultHistogramBuckets() []float64 {
	return []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
}

// Observe records a value in the histogram.
func (h *Histogram) Observe(value float64) {
	// Find the bucket
	idx := sort.SearchFloat64s(h.buckets, value)
	h.counts[idx].Add(1)
	h.count.Add(1)

	// Add to sum (convert to uint64 for atomic operation)
	// Using fixed-point arithmetic: multiply by 1000000 for microsecond precision
	h.sum.Add(uint64(value * 1000000))
}

// ObserveDuration records a duration in the histogram.
func (h *Histogram) ObserveDuration(d time.Duration) {
	h.Observe(d.Seconds())
}

// Count returns the total number of observations.
func (h *Histogram) Count() uint64 {
	return h.count.Load()
}

// Sum returns the sum of all observations.
func (h *Histogram) Sum() float64 {
	return float64(h.sum.Load()) / 1000000
}

// Mean returns the mean of all observations.
func (h *Histogram) Mean() float64 {
	count := h.count.Load()
	if count == 0 {
		return 0
	}
	return h.Sum() / float64(count)
}

// Buckets returns the bucket boundaries and counts.
func (h *Histogram) Buckets() ([]float64, []uint64) {
	counts := make([]uint64, len(h.counts))
	for i := range h.counts {
		counts[i] = h.counts[i].Load()
	}
	return h.buckets, counts
}

// Name returns the histogram name.
func (h *Histogram) Name() string {
	return h.name
}

// Labels returns the histogram labels.
func (h *Histogram) Labels() map[string]string {
	return h.labels
}

// Summary tracks quantiles of values using a sliding window.
type Summary struct {
	name       string
	labels     map[string]string
	quantiles  []float64
	maxAge     time.Duration
	ageBuckets int
	mu         sync.RWMutex
	values     []timedValue
	count      atomic.Uint64
	sum        atomic.Uint64
}

type timedValue struct {
	value     float64
	timestamp time.Time
}

// NewSummary creates a new summary with the given quantiles.
func NewSummary(name string, labels map[string]string, quantiles []float64, maxAge time.Duration) *Summary {
	return &Summary{
		name:      name,
		labels:    labels,
		quantiles: quantiles,
		maxAge:    maxAge,
		values:    make([]timedValue, 0, 1000),
	}
}

// DefaultSummaryQuantiles returns default quantiles for summary metrics.
func DefaultSummaryQuantiles() []float64 {
	return []float64{0.5, 0.9, 0.95, 0.99}
}

// Observe records a value in the summary.
func (s *Summary) Observe(value float64) {
	s.mu.Lock()
	s.values = append(s.values, timedValue{value: value, timestamp: time.Now()})
	s.mu.Unlock()

	s.count.Add(1)
	s.sum.Add(uint64(value * 1000000))
}

// ObserveDuration records a duration in the summary.
func (s *Summary) ObserveDuration(d time.Duration) {
	s.Observe(d.Seconds())
}

// Quantile returns the value at the given quantile.
func (s *Summary) Quantile(q float64) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Filter out old values
	now := time.Now()
	var validValues []float64
	for _, v := range s.values {
		if now.Sub(v.timestamp) <= s.maxAge {
			validValues = append(validValues, v.value)
		}
	}

	if len(validValues) == 0 {
		return 0
	}

	sort.Float64s(validValues)
	idx := int(float64(len(validValues)) * q)
	if idx >= len(validValues) {
		idx = len(validValues) - 1
	}
	return validValues[idx]
}

// Quantiles returns all configured quantile values.
func (s *Summary) Quantiles() map[float64]float64 {
	result := make(map[float64]float64)
	for _, q := range s.quantiles {
		result[q] = s.Quantile(q)
	}
	return result
}

// Count returns the total number of observations.
func (s *Summary) Count() uint64 {
	return s.count.Load()
}

// Sum returns the sum of all observations.
func (s *Summary) Sum() float64 {
	return float64(s.sum.Load()) / 1000000
}

// Name returns the summary name.
func (s *Summary) Name() string {
	return s.name
}

// Labels returns the summary labels.
func (s *Summary) Labels() map[string]string {
	return s.labels
}

// Cleanup removes old values from the summary.
func (s *Summary) Cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	validValues := make([]timedValue, 0, len(s.values))
	for _, v := range s.values {
		if now.Sub(v.timestamp) <= s.maxAge {
			validValues = append(validValues, v)
		}
	}
	s.values = validValues
}

// Timer is a helper for timing operations.
type Timer struct {
	histogram *Histogram
	summary   *Summary
	start     time.Time
}

// NewTimer creates a new timer that records to a histogram.
func NewTimer(h *Histogram) *Timer {
	return &Timer{
		histogram: h,
		start:     time.Now(),
	}
}

// NewTimerWithSummary creates a new timer that records to a summary.
func NewTimerWithSummary(s *Summary) *Timer {
	return &Timer{
		summary: s,
		start:   time.Now(),
	}
}

// ObserveDuration records the elapsed time since the timer was created.
func (t *Timer) ObserveDuration() time.Duration {
	d := time.Since(t.start)
	if t.histogram != nil {
		t.histogram.ObserveDuration(d)
	}
	if t.summary != nil {
		t.summary.ObserveDuration(d)
	}
	return d
}

// MetricsRegistry is a registry for all metrics.
type MetricsRegistry struct {
	mu         sync.RWMutex
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	summaries  map[string]*Summary
}

// NewMetricsRegistry creates a new metrics registry.
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		summaries:  make(map[string]*Summary),
	}
}

// Counter returns or creates a counter with the given name.
func (r *MetricsRegistry) Counter(name string, labels map[string]string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if c, ok := r.counters[key]; ok {
		return c
	}

	c := NewCounter(name, labels)
	r.counters[key] = c
	return c
}

// Gauge returns or creates a gauge with the given name.
func (r *MetricsRegistry) Gauge(name string, labels map[string]string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if g, ok := r.gauges[key]; ok {
		return g
	}

	g := NewGauge(name, labels)
	r.gauges[key] = g
	return g
}

// Histogram returns or creates a histogram with the given name.
func (r *MetricsRegistry) Histogram(name string, labels map[string]string, buckets []float64) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if h, ok := r.histograms[key]; ok {
		return h
	}

	if buckets == nil {
		buckets = DefaultHistogramBuckets()
	}
	h := NewHistogram(name, labels, buckets)
	r.histograms[key] = h
	return h
}

// Summary returns or creates a summary with the given name.
func (r *MetricsRegistry) Summary(name string, labels map[string]string, quantiles []float64, maxAge time.Duration) *Summary {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.makeKey(name, labels)
	if s, ok := r.summaries[key]; ok {
		return s
	}

	if quantiles == nil {
		quantiles = DefaultSummaryQuantiles()
	}
	if maxAge == 0 {
		maxAge = 10 * time.Minute
	}
	s := NewSummary(name, labels, quantiles, maxAge)
	r.summaries[key] = s
	return s
}

func (r *MetricsRegistry) makeKey(name string, labels map[string]string) string {
	key := name
	if len(labels) > 0 {
		keys := make([]string, 0, len(labels))
		for k := range labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			key += fmt.Sprintf(",%s=%s", k, labels[k])
		}
	}
	return key
}

// Snapshot returns a snapshot of all metrics.
func (r *MetricsRegistry) Snapshot() *MetricsSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := &MetricsSnapshot{
		Timestamp: time.Now(),
		Counters:  make(map[string]int64),
		Gauges:    make(map[string]int64),
	}

	for key, c := range r.counters {
		snapshot.Counters[key] = c.Value()
	}

	for key, g := range r.gauges {
		snapshot.Gauges[key] = g.Value()
	}

	return snapshot
}

// MetricsSnapshot represents a point-in-time snapshot of all metrics.
type MetricsSnapshot struct {
	Timestamp time.Time
	Counters  map[string]int64
	Gauges    map[string]int64
}

// RateCalculator calculates rates from counter values.
type RateCalculator struct {
	mu       sync.Mutex
	previous map[string]ratePoint
}

type ratePoint struct {
	value     int64
	timestamp time.Time
}

// NewRateCalculator creates a new rate calculator.
func NewRateCalculator() *RateCalculator {
	return &RateCalculator{
		previous: make(map[string]ratePoint),
	}
}

// Rate calculates the rate of change for a counter.
func (rc *RateCalculator) Rate(name string, currentValue int64) float64 {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	now := time.Now()
	prev, ok := rc.previous[name]
	rc.previous[name] = ratePoint{value: currentValue, timestamp: now}

	if !ok {
		return 0
	}

	duration := now.Sub(prev.timestamp).Seconds()
	if duration == 0 {
		return 0
	}

	return float64(currentValue-prev.value) / duration
}

// MovingAverage calculates a moving average.
type MovingAverage struct {
	mu     sync.Mutex
	window int
	values []float64
	sum    float64
}

// NewMovingAverage creates a new moving average calculator.
func NewMovingAverage(window int) *MovingAverage {
	return &MovingAverage{
		window: window,
		values: make([]float64, 0, window),
	}
}

// Add adds a value to the moving average.
func (ma *MovingAverage) Add(value float64) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	if len(ma.values) >= ma.window {
		ma.sum -= ma.values[0]
		ma.values = ma.values[1:]
	}
	ma.values = append(ma.values, value)
	ma.sum += value
}

// Average returns the current moving average.
func (ma *MovingAverage) Average() float64 {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	if len(ma.values) == 0 {
		return 0
	}
	return ma.sum / float64(len(ma.values))
}

// ExponentialMovingAverage calculates an exponential moving average.
type ExponentialMovingAverage struct {
	mu      sync.Mutex
	alpha   float64
	value   float64
	hasData bool
}

// NewExponentialMovingAverage creates a new EMA calculator.
// Alpha is the smoothing factor (0 < alpha <= 1). Higher alpha = more weight on recent values.
func NewExponentialMovingAverage(alpha float64) *ExponentialMovingAverage {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.1 // Default
	}
	return &ExponentialMovingAverage{
		alpha: alpha,
	}
}

// Add adds a value to the EMA.
func (ema *ExponentialMovingAverage) Add(value float64) {
	ema.mu.Lock()
	defer ema.mu.Unlock()

	if !ema.hasData {
		ema.value = value
		ema.hasData = true
		return
	}
	ema.value = ema.alpha*value + (1-ema.alpha)*ema.value
}

// Value returns the current EMA value.
func (ema *ExponentialMovingAverage) Value() float64 {
	ema.mu.Lock()
	defer ema.mu.Unlock()
	return ema.value
}

// PercentileCalculator calculates percentiles from a stream of values.
type PercentileCalculator struct {
	mu     sync.Mutex
	values []float64
	sorted bool
}

// NewPercentileCalculator creates a new percentile calculator.
func NewPercentileCalculator() *PercentileCalculator {
	return &PercentileCalculator{
		values: make([]float64, 0, 1000),
	}
}

// Add adds a value to the calculator.
func (pc *PercentileCalculator) Add(value float64) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.values = append(pc.values, value)
	pc.sorted = false
}

// Percentile returns the value at the given percentile (0-100).
func (pc *PercentileCalculator) Percentile(p float64) float64 {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if len(pc.values) == 0 {
		return 0
	}

	if !pc.sorted {
		sort.Float64s(pc.values)
		pc.sorted = true
	}

	idx := int(math.Ceil(float64(len(pc.values)) * p / 100))
	if idx >= len(pc.values) {
		idx = len(pc.values) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return pc.values[idx]
}

// Reset clears all values.
func (pc *PercentileCalculator) Reset() {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.values = pc.values[:0]
	pc.sorted = false
}

// Count returns the number of values.
func (pc *PercentileCalculator) Count() int {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return len(pc.values)
}

// DefaultRegistry is the default metrics registry.
var DefaultRegistry = NewMetricsRegistry()

// GetCounter returns a counter from the default registry.
func GetCounter(name string, labels map[string]string) *Counter {
	return DefaultRegistry.Counter(name, labels)
}

// GetGauge returns a gauge from the default registry.
func GetGauge(name string, labels map[string]string) *Gauge {
	return DefaultRegistry.Gauge(name, labels)
}

// GetHistogram returns a histogram from the default registry.
func GetHistogram(name string, labels map[string]string, buckets []float64) *Histogram {
	return DefaultRegistry.Histogram(name, labels, buckets)
}

// GetSummary returns a summary from the default registry.
func GetSummary(name string, labels map[string]string, quantiles []float64, maxAge time.Duration) *Summary {
	return DefaultRegistry.Summary(name, labels, quantiles, maxAge)
}
