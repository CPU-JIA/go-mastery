package perfutil

import (
	"sync"
	"testing"
	"time"
)

func TestCounter(t *testing.T) {
	c := NewCounter("test_counter", map[string]string{"env": "test"})

	if c.Value() != 0 {
		t.Errorf("Expected initial value 0, got %d", c.Value())
	}

	c.Inc()
	if c.Value() != 1 {
		t.Errorf("Expected value 1 after Inc, got %d", c.Value())
	}

	c.Add(10)
	if c.Value() != 11 {
		t.Errorf("Expected value 11 after Add(10), got %d", c.Value())
	}

	// Negative values should be ignored
	c.Add(-5)
	if c.Value() != 11 {
		t.Errorf("Expected value 11 after Add(-5), got %d", c.Value())
	}

	if c.Name() != "test_counter" {
		t.Errorf("Expected name 'test_counter', got '%s'", c.Name())
	}

	labels := c.Labels()
	if labels["env"] != "test" {
		t.Errorf("Expected label env=test, got %v", labels)
	}

	c.Reset()
	if c.Value() != 0 {
		t.Errorf("Expected value 0 after Reset, got %d", c.Value())
	}
}

func TestCounter_Concurrent(t *testing.T) {
	c := NewCounter("concurrent_counter", nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc()
		}()
	}
	wg.Wait()

	if c.Value() != 100 {
		t.Errorf("Expected value 100, got %d", c.Value())
	}
}

func TestGauge(t *testing.T) {
	g := NewGauge("test_gauge", map[string]string{"env": "test"})

	if g.Value() != 0 {
		t.Errorf("Expected initial value 0, got %d", g.Value())
	}

	g.Set(50)
	if g.Value() != 50 {
		t.Errorf("Expected value 50 after Set, got %d", g.Value())
	}

	g.Inc()
	if g.Value() != 51 {
		t.Errorf("Expected value 51 after Inc, got %d", g.Value())
	}

	g.Dec()
	if g.Value() != 50 {
		t.Errorf("Expected value 50 after Dec, got %d", g.Value())
	}

	g.Add(10)
	if g.Value() != 60 {
		t.Errorf("Expected value 60 after Add(10), got %d", g.Value())
	}

	g.Sub(20)
	if g.Value() != 40 {
		t.Errorf("Expected value 40 after Sub(20), got %d", g.Value())
	}

	if g.Name() != "test_gauge" {
		t.Errorf("Expected name 'test_gauge', got '%s'", g.Name())
	}
}

func TestGauge_Concurrent(t *testing.T) {
	g := NewGauge("concurrent_gauge", nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			g.Inc()
		}()
		go func() {
			defer wg.Done()
			g.Dec()
		}()
	}
	wg.Wait()

	// After equal increments and decrements, should be 0
	if g.Value() != 0 {
		t.Errorf("Expected value 0, got %d", g.Value())
	}
}

func TestHistogram(t *testing.T) {
	h := NewHistogram("test_histogram", map[string]string{"env": "test"}, DefaultHistogramBuckets())

	// Observe some values
	h.Observe(0.001)
	h.Observe(0.01)
	h.Observe(0.1)
	h.Observe(1.0)
	h.Observe(5.0)

	if h.Count() != 5 {
		t.Errorf("Expected count 5, got %d", h.Count())
	}

	sum := h.Sum()
	expectedSum := 0.001 + 0.01 + 0.1 + 1.0 + 5.0
	if sum < expectedSum-0.001 || sum > expectedSum+0.001 {
		t.Errorf("Expected sum ~%f, got %f", expectedSum, sum)
	}

	mean := h.Mean()
	expectedMean := expectedSum / 5
	if mean < expectedMean-0.001 || mean > expectedMean+0.001 {
		t.Errorf("Expected mean ~%f, got %f", expectedMean, mean)
	}

	if h.Name() != "test_histogram" {
		t.Errorf("Expected name 'test_histogram', got '%s'", h.Name())
	}

	buckets, counts := h.Buckets()
	if len(buckets) == 0 {
		t.Error("Expected non-empty buckets")
	}
	if len(counts) == 0 {
		t.Error("Expected non-empty counts")
	}
}

func TestHistogram_ObserveDuration(t *testing.T) {
	h := NewHistogram("duration_histogram", nil, DefaultHistogramBuckets())

	h.ObserveDuration(100 * time.Millisecond)
	h.ObserveDuration(200 * time.Millisecond)

	if h.Count() != 2 {
		t.Errorf("Expected count 2, got %d", h.Count())
	}
}

func TestHistogram_Concurrent(t *testing.T) {
	h := NewHistogram("concurrent_histogram", nil, DefaultHistogramBuckets())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			h.Observe(val)
		}(float64(i) / 100)
	}
	wg.Wait()

	if h.Count() != 100 {
		t.Errorf("Expected count 100, got %d", h.Count())
	}
}

func TestSummary(t *testing.T) {
	s := NewSummary("test_summary", map[string]string{"env": "test"}, DefaultSummaryQuantiles(), time.Minute)

	// Observe some values
	for i := 1; i <= 100; i++ {
		s.Observe(float64(i))
	}

	if s.Count() != 100 {
		t.Errorf("Expected count 100, got %d", s.Count())
	}

	// Check quantiles
	p50 := s.Quantile(0.5)
	if p50 < 45 || p50 > 55 {
		t.Errorf("Expected P50 ~50, got %f", p50)
	}

	p99 := s.Quantile(0.99)
	if p99 < 95 {
		t.Errorf("Expected P99 >= 95, got %f", p99)
	}

	quantiles := s.Quantiles()
	if len(quantiles) != 4 {
		t.Errorf("Expected 4 quantiles, got %d", len(quantiles))
	}

	if s.Name() != "test_summary" {
		t.Errorf("Expected name 'test_summary', got '%s'", s.Name())
	}
}

func TestSummary_ObserveDuration(t *testing.T) {
	s := NewSummary("duration_summary", nil, DefaultSummaryQuantiles(), time.Minute)

	s.ObserveDuration(100 * time.Millisecond)
	s.ObserveDuration(200 * time.Millisecond)

	if s.Count() != 2 {
		t.Errorf("Expected count 2, got %d", s.Count())
	}
}

func TestSummary_Cleanup(t *testing.T) {
	s := NewSummary("cleanup_summary", nil, DefaultSummaryQuantiles(), 10*time.Millisecond)

	s.Observe(1.0)
	s.Observe(2.0)

	// Wait for values to expire
	time.Sleep(20 * time.Millisecond)

	s.Cleanup()

	// Quantile should return 0 for empty values
	p50 := s.Quantile(0.5)
	if p50 != 0 {
		t.Errorf("Expected P50 0 after cleanup, got %f", p50)
	}
}

func TestSummary_Empty(t *testing.T) {
	s := NewSummary("empty_summary", nil, DefaultSummaryQuantiles(), time.Minute)

	p50 := s.Quantile(0.5)
	if p50 != 0 {
		t.Errorf("Expected P50 0 for empty summary, got %f", p50)
	}
}

func TestTimer(t *testing.T) {
	h := NewHistogram("timer_histogram", nil, DefaultHistogramBuckets())

	timer := NewTimer(h)
	time.Sleep(10 * time.Millisecond)
	duration := timer.ObserveDuration()

	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", duration)
	}

	if h.Count() != 1 {
		t.Errorf("Expected histogram count 1, got %d", h.Count())
	}
}

func TestTimerWithSummary(t *testing.T) {
	s := NewSummary("timer_summary", nil, DefaultSummaryQuantiles(), time.Minute)

	timer := NewTimerWithSummary(s)
	time.Sleep(10 * time.Millisecond)
	duration := timer.ObserveDuration()

	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", duration)
	}

	if s.Count() != 1 {
		t.Errorf("Expected summary count 1, got %d", s.Count())
	}
}

func TestMetricsRegistry(t *testing.T) {
	r := NewMetricsRegistry()

	// Test Counter
	c1 := r.Counter("requests", map[string]string{"method": "GET"})
	c1.Inc()

	c2 := r.Counter("requests", map[string]string{"method": "GET"})
	if c1 != c2 {
		t.Error("Expected same counter instance for same name and labels")
	}

	if c2.Value() != 1 {
		t.Errorf("Expected counter value 1, got %d", c2.Value())
	}

	// Test Gauge
	g1 := r.Gauge("connections", map[string]string{"type": "active"})
	g1.Set(10)

	g2 := r.Gauge("connections", map[string]string{"type": "active"})
	if g1 != g2 {
		t.Error("Expected same gauge instance for same name and labels")
	}

	// Test Histogram
	h1 := r.Histogram("latency", map[string]string{"endpoint": "/api"}, nil)
	h1.Observe(0.1)

	h2 := r.Histogram("latency", map[string]string{"endpoint": "/api"}, nil)
	if h1 != h2 {
		t.Error("Expected same histogram instance for same name and labels")
	}

	// Test Summary
	s1 := r.Summary("response_time", map[string]string{"service": "api"}, nil, 0)
	s1.Observe(0.1)

	s2 := r.Summary("response_time", map[string]string{"service": "api"}, nil, 0)
	if s1 != s2 {
		t.Error("Expected same summary instance for same name and labels")
	}
}

func TestMetricsRegistry_Snapshot(t *testing.T) {
	r := NewMetricsRegistry()

	c := r.Counter("test_counter", nil)
	c.Add(100)

	g := r.Gauge("test_gauge", nil)
	g.Set(50)

	snapshot := r.Snapshot()

	if snapshot.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if snapshot.Counters["test_counter"] != 100 {
		t.Errorf("Expected counter value 100, got %d", snapshot.Counters["test_counter"])
	}

	if snapshot.Gauges["test_gauge"] != 50 {
		t.Errorf("Expected gauge value 50, got %d", snapshot.Gauges["test_gauge"])
	}
}

func TestRateCalculator(t *testing.T) {
	rc := NewRateCalculator()

	// First call should return 0
	rate := rc.Rate("counter", 100)
	if rate != 0 {
		t.Errorf("Expected rate 0 for first call, got %f", rate)
	}

	// Wait a bit and calculate rate
	time.Sleep(100 * time.Millisecond)
	rate = rc.Rate("counter", 200)

	// Rate should be approximately 1000/sec (100 increase over 0.1 sec)
	if rate < 500 || rate > 1500 {
		t.Errorf("Expected rate ~1000/sec, got %f", rate)
	}
}

func TestMovingAverage(t *testing.T) {
	ma := NewMovingAverage(5)

	// Empty average should be 0
	if ma.Average() != 0 {
		t.Errorf("Expected average 0 for empty, got %f", ma.Average())
	}

	// Add values
	ma.Add(10)
	ma.Add(20)
	ma.Add(30)

	avg := ma.Average()
	expected := 20.0 // (10+20+30)/3
	if avg != expected {
		t.Errorf("Expected average %f, got %f", expected, avg)
	}

	// Fill window
	ma.Add(40)
	ma.Add(50)

	avg = ma.Average()
	expected = 30.0 // (10+20+30+40+50)/5
	if avg != expected {
		t.Errorf("Expected average %f, got %f", expected, avg)
	}

	// Add more - should slide window
	ma.Add(60)

	avg = ma.Average()
	expected = 40.0 // (20+30+40+50+60)/5
	if avg != expected {
		t.Errorf("Expected average %f, got %f", expected, avg)
	}
}

func TestExponentialMovingAverage(t *testing.T) {
	ema := NewExponentialMovingAverage(0.5)

	// First value sets the EMA
	ema.Add(100)
	if ema.Value() != 100 {
		t.Errorf("Expected EMA 100, got %f", ema.Value())
	}

	// Second value: EMA = 0.5*50 + 0.5*100 = 75
	ema.Add(50)
	if ema.Value() != 75 {
		t.Errorf("Expected EMA 75, got %f", ema.Value())
	}
}

func TestExponentialMovingAverage_InvalidAlpha(t *testing.T) {
	// Invalid alpha should default to 0.1
	ema := NewExponentialMovingAverage(0)
	ema.Add(100)
	ema.Add(50)

	// With alpha=0.1: EMA = 0.1*50 + 0.9*100 = 95
	if ema.Value() != 95 {
		t.Errorf("Expected EMA 95, got %f", ema.Value())
	}
}

func TestPercentileCalculator(t *testing.T) {
	pc := NewPercentileCalculator()

	// Empty should return 0
	if pc.Percentile(50) != 0 {
		t.Errorf("Expected P50 0 for empty, got %f", pc.Percentile(50))
	}

	// Add values 1-100
	for i := 1; i <= 100; i++ {
		pc.Add(float64(i))
	}

	if pc.Count() != 100 {
		t.Errorf("Expected count 100, got %d", pc.Count())
	}

	p50 := pc.Percentile(50)
	if p50 < 49 || p50 > 51 {
		t.Errorf("Expected P50 ~50, got %f", p50)
	}

	p99 := pc.Percentile(99)
	if p99 < 98 {
		t.Errorf("Expected P99 >= 98, got %f", p99)
	}

	pc.Reset()
	if pc.Count() != 0 {
		t.Errorf("Expected count 0 after reset, got %d", pc.Count())
	}
}

func TestDefaultRegistry(t *testing.T) {
	// Test global functions
	c := GetCounter("global_counter", nil)
	c.Inc()

	g := GetGauge("global_gauge", nil)
	g.Set(100)

	h := GetHistogram("global_histogram", nil, nil)
	h.Observe(0.1)

	s := GetSummary("global_summary", nil, nil, 0)
	s.Observe(0.1)

	// Verify they're in the default registry
	snapshot := DefaultRegistry.Snapshot()
	if snapshot.Counters["global_counter"] != 1 {
		t.Error("Expected global counter in default registry")
	}
}

func TestMetricType_String(t *testing.T) {
	tests := []struct {
		mt       MetricType
		expected string
	}{
		{MetricCounter, "counter"},
		{MetricGauge, "gauge"},
		{MetricHistogram, "histogram"},
		{MetricSummary, "summary"},
		{MetricType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.mt.String(); got != tt.expected {
			t.Errorf("MetricType(%d).String() = %s, want %s", tt.mt, got, tt.expected)
		}
	}
}

func TestDefaultHistogramBuckets(t *testing.T) {
	buckets := DefaultHistogramBuckets()
	if len(buckets) == 0 {
		t.Error("Expected non-empty default buckets")
	}

	// Verify buckets are sorted
	for i := 1; i < len(buckets); i++ {
		if buckets[i] <= buckets[i-1] {
			t.Error("Expected sorted buckets")
		}
	}
}

func TestDefaultSummaryQuantiles(t *testing.T) {
	quantiles := DefaultSummaryQuantiles()
	if len(quantiles) == 0 {
		t.Error("Expected non-empty default quantiles")
	}

	// Verify quantiles are in valid range
	for _, q := range quantiles {
		if q < 0 || q > 1 {
			t.Errorf("Invalid quantile: %f", q)
		}
	}
}

// Benchmarks
func BenchmarkCounter_Inc(b *testing.B) {
	c := NewCounter("bench_counter", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Inc()
	}
}

func BenchmarkGauge_Set(b *testing.B) {
	g := NewGauge("bench_gauge", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Set(int64(i))
	}
}

func BenchmarkHistogram_Observe(b *testing.B) {
	h := NewHistogram("bench_histogram", nil, DefaultHistogramBuckets())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Observe(float64(i) / float64(b.N))
	}
}

func BenchmarkSummary_Observe(b *testing.B) {
	s := NewSummary("bench_summary", nil, DefaultSummaryQuantiles(), time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Observe(float64(i))
	}
}

func BenchmarkMovingAverage_Add(b *testing.B) {
	ma := NewMovingAverage(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ma.Add(float64(i))
	}
}

func BenchmarkExponentialMovingAverage_Add(b *testing.B) {
	ema := NewExponentialMovingAverage(0.1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ema.Add(float64(i))
	}
}

func BenchmarkPercentileCalculator_Add(b *testing.B) {
	pc := NewPercentileCalculator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Add(float64(i))
	}
}

func BenchmarkMetricsRegistry_Counter(b *testing.B) {
	r := NewMetricsRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := r.Counter("bench", nil)
		c.Inc()
	}
}

// Concurrent benchmarks
func BenchmarkCounter_Inc_Parallel(b *testing.B) {
	c := NewCounter("bench_counter_parallel", nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Inc()
		}
	})
}

func BenchmarkHistogram_Observe_Parallel(b *testing.B) {
	h := NewHistogram("bench_histogram_parallel", nil, DefaultHistogramBuckets())

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			h.Observe(float64(i) / 1000)
			i++
		}
	})
}
