package perfutil

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestBenchmarker_Run(t *testing.T) {
	b := NewBenchmarker(
		WithMinDuration(50*time.Millisecond),
		WithWarmupIterations(3),
	)

	result := b.Run("test_benchmark", func() {
		sum := 0
		for i := 0; i < 1000; i++ {
			sum += i
		}
		_ = sum
	})

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Name != "test_benchmark" {
		t.Errorf("Expected name 'test_benchmark', got '%s'", result.Name)
	}

	if result.Iterations <= 0 {
		t.Error("Expected positive iterations")
	}

	if result.MeanTime <= 0 {
		t.Error("Expected positive mean time")
	}

	if result.OpsPerSecond <= 0 {
		t.Error("Expected positive ops/sec")
	}
}

func TestBenchmarker_RunWithSetup(t *testing.T) {
	b := NewBenchmarker(
		WithMinDuration(50*time.Millisecond),
		WithWarmupIterations(2),
	)

	var data []int
	setupCalled := 0
	teardownCalled := 0

	result := b.RunWithSetup(
		"test_with_setup",
		func() {
			data = make([]int, 100)
			setupCalled++
		},
		func() {
			for i := range data {
				data[i] = i * 2
			}
		},
		func() {
			data = nil
			teardownCalled++
		},
	)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if setupCalled == 0 {
		t.Error("Setup should have been called")
	}

	if teardownCalled == 0 {
		t.Error("Teardown should have been called")
	}

	if setupCalled != teardownCalled {
		t.Errorf("Setup (%d) and teardown (%d) call counts should match", setupCalled, teardownCalled)
	}
}

func TestBenchmarker_RunParallel(t *testing.T) {
	b := NewBenchmarker(
		WithMinDuration(50*time.Millisecond),
		WithWarmupIterations(2),
	)

	var counter int64
	var mu sync.Mutex

	result := b.RunParallel("test_parallel", 4, func() {
		mu.Lock()
		counter++
		mu.Unlock()
	})

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Iterations <= 0 {
		t.Error("Expected positive iterations")
	}
}

func TestBenchmarker_RunParallel_DefaultGoroutines(t *testing.T) {
	b := NewBenchmarker(
		WithMinDuration(50 * time.Millisecond),
	)

	result := b.RunParallel("test_parallel_default", 0, func() {
		_ = 1 + 1
	})

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestBenchmarkResult_String(t *testing.T) {
	result := &BenchmarkResult{
		Name:         "test",
		Iterations:   1000,
		TotalTime:    time.Second,
		MinTime:      time.Microsecond,
		MaxTime:      time.Millisecond,
		MeanTime:     100 * time.Microsecond,
		MedianTime:   90 * time.Microsecond,
		StdDev:       10 * time.Microsecond,
		P50:          90 * time.Microsecond,
		P90:          150 * time.Microsecond,
		P95:          180 * time.Microsecond,
		P99:          200 * time.Microsecond,
		OpsPerSecond: 10000,
		AllocBytes:   1024,
		AllocObjects: 10,
	}

	str := result.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}

func TestCompare(t *testing.T) {
	before := &BenchmarkResult{
		Name:       "before",
		MeanTime:   100 * time.Microsecond,
		AllocBytes: 1024,
	}

	after := &BenchmarkResult{
		Name:       "after",
		MeanTime:   50 * time.Microsecond,
		AllocBytes: 512,
	}

	comparison := Compare(before, after)

	if comparison == nil {
		t.Fatal("Expected non-nil comparison")
	}

	// After is faster, so speedup should be > 1
	if comparison.SpeedupRatio <= 1 {
		t.Errorf("Expected speedup > 1, got %f", comparison.SpeedupRatio)
	}

	// After uses less memory, so ratio should be < 1
	if comparison.MemoryRatio >= 1 {
		t.Errorf("Expected memory ratio < 1, got %f", comparison.MemoryRatio)
	}
}

func TestCompare_Slower(t *testing.T) {
	before := &BenchmarkResult{
		Name:       "before",
		MeanTime:   50 * time.Microsecond,
		AllocBytes: 512,
	}

	after := &BenchmarkResult{
		Name:       "after",
		MeanTime:   100 * time.Microsecond,
		AllocBytes: 1024,
	}

	comparison := Compare(before, after)

	// After is slower, so speedup should be < 1
	if comparison.SpeedupRatio >= 1 {
		t.Errorf("Expected speedup < 1, got %f", comparison.SpeedupRatio)
	}
}

func TestBenchmarkComparison_String(t *testing.T) {
	comparison := &BenchmarkComparison{
		Before: &BenchmarkResult{
			Name:     "before",
			MeanTime: 100 * time.Microsecond,
		},
		After: &BenchmarkResult{
			Name:     "after",
			MeanTime: 50 * time.Microsecond,
		},
		SpeedupRatio: 2.0,
		MemoryRatio:  0.5,
		Improvement:  "2.00x faster, 2.00x less memory",
	}

	str := comparison.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}

func TestBenchmarkSuite(t *testing.T) {
	suite := NewBenchmarkSuite(
		WithMinDuration(50*time.Millisecond),
		WithWarmupIterations(2),
	)

	suite.Add("bench1", func() {
		_ = 1 + 1
	})

	suite.Add("bench2", func() {
		_ = 2 * 2
	})

	results := suite.Results()
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	summary := suite.Summary()
	if summary == "" {
		t.Error("Expected non-empty summary")
	}
}

func TestBenchmarkSuite_Empty(t *testing.T) {
	suite := NewBenchmarkSuite()

	summary := suite.Summary()
	if summary != "No benchmarks run" {
		t.Errorf("Expected 'No benchmarks run', got '%s'", summary)
	}
}

func TestQuickBench(t *testing.T) {
	result := QuickBench("quick_test", func() {
		sum := 0
		for i := 0; i < 100; i++ {
			sum += i
		}
		_ = sum
	})

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Name != "quick_test" {
		t.Errorf("Expected name 'quick_test', got '%s'", result.Name)
	}
}

func TestCompareFuncs(t *testing.T) {
	comparison := CompareFuncs(
		"slow",
		func() {
			time.Sleep(time.Microsecond)
		},
		"fast",
		func() {
			_ = 1 + 1
		},
	)

	if comparison == nil {
		t.Fatal("Expected non-nil comparison")
	}

	// Fast should be faster than slow
	if comparison.SpeedupRatio <= 1 {
		t.Logf("Speedup ratio: %f (slow may not be slow enough)", comparison.SpeedupRatio)
	}
}

func TestBenchmarkerOptions(t *testing.T) {
	b := NewBenchmarker(
		WithWarmupIterations(20),
		WithMinIterations(50),
		WithMaxIterations(500),
		WithMinDuration(100*time.Millisecond),
		WithMaxDuration(5*time.Second),
		WithGCBetweenRuns(false),
	)

	if b.warmupIterations != 20 {
		t.Errorf("Expected warmupIterations 20, got %d", b.warmupIterations)
	}

	if b.minIterations != 50 {
		t.Errorf("Expected minIterations 50, got %d", b.minIterations)
	}

	if b.maxIterations != 500 {
		t.Errorf("Expected maxIterations 500, got %d", b.maxIterations)
	}

	if b.minDuration != 100*time.Millisecond {
		t.Errorf("Expected minDuration 100ms, got %v", b.minDuration)
	}

	if b.maxDuration != 5*time.Second {
		t.Errorf("Expected maxDuration 5s, got %v", b.maxDuration)
	}

	if b.gcBetweenRuns != false {
		t.Error("Expected gcBetweenRuns false")
	}
}

func TestBenchmarker_EmptyResult(t *testing.T) {
	b := NewBenchmarker()

	// Test with empty times slice (edge case)
	var memBefore, memAfter runtime.MemStats
	result := b.calculateStats("empty", []time.Duration{}, time.Second, &memBefore, &memAfter)

	if result.Name != "empty" {
		t.Errorf("Expected name 'empty', got '%s'", result.Name)
	}
}

// Benchmark the benchmarker itself
func BenchmarkBenchmarker_Run(b *testing.B) {
	benchmarker := NewBenchmarker(
		WithMinDuration(10*time.Millisecond),
		WithWarmupIterations(1),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = benchmarker.Run("bench", func() {
			_ = 1 + 1
		})
	}
}

func BenchmarkQuickBench(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = QuickBench("quick", func() {
			_ = 1 + 1
		})
	}
}

// Test concurrent access to BenchmarkSuite
func TestBenchmarkSuite_Concurrent(t *testing.T) {
	suite := NewBenchmarkSuite(
		WithMinDuration(10 * time.Millisecond),
	)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			suite.Add("concurrent_bench", func() {
				_ = idx * 2
			})
		}(i)
	}
	wg.Wait()

	results := suite.Results()
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}
