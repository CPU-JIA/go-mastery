package perfutil

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"
	"time"
)

// BenchmarkResult contains the results of a benchmark run.
type BenchmarkResult struct {
	Name         string
	Iterations   int
	TotalTime    time.Duration
	MinTime      time.Duration
	MaxTime      time.Duration
	MeanTime     time.Duration
	MedianTime   time.Duration
	StdDev       time.Duration
	P50          time.Duration
	P90          time.Duration
	P95          time.Duration
	P99          time.Duration
	OpsPerSecond float64
	AllocBytes   int64
	AllocObjects int64
}

// String returns a human-readable representation of the benchmark result.
func (r *BenchmarkResult) String() string {
	return fmt.Sprintf(
		"Benchmark: %s\n"+
			"  Iterations:    %d\n"+
			"  Total Time:    %v\n"+
			"  Mean:          %v\n"+
			"  Median:        %v\n"+
			"  Min:           %v\n"+
			"  Max:           %v\n"+
			"  StdDev:        %v\n"+
			"  P50:           %v\n"+
			"  P90:           %v\n"+
			"  P95:           %v\n"+
			"  P99:           %v\n"+
			"  Ops/sec:       %.2f\n"+
			"  Alloc Bytes:   %d\n"+
			"  Alloc Objects: %d",
		r.Name,
		r.Iterations,
		r.TotalTime,
		r.MeanTime,
		r.MedianTime,
		r.MinTime,
		r.MaxTime,
		r.StdDev,
		r.P50,
		r.P90,
		r.P95,
		r.P99,
		r.OpsPerSecond,
		r.AllocBytes,
		r.AllocObjects,
	)
}

// Benchmarker provides utilities for running and comparing benchmarks.
type Benchmarker struct {
	warmupIterations int
	minIterations    int
	maxIterations    int
	minDuration      time.Duration
	maxDuration      time.Duration
	gcBetweenRuns    bool
}

// BenchmarkerOption configures a Benchmarker.
type BenchmarkerOption func(*Benchmarker)

// WithWarmupIterations sets the number of warmup iterations.
func WithWarmupIterations(n int) BenchmarkerOption {
	return func(b *Benchmarker) {
		b.warmupIterations = n
	}
}

// WithMinIterations sets the minimum number of benchmark iterations.
func WithMinIterations(n int) BenchmarkerOption {
	return func(b *Benchmarker) {
		b.minIterations = n
	}
}

// WithMaxIterations sets the maximum number of benchmark iterations.
func WithMaxIterations(n int) BenchmarkerOption {
	return func(b *Benchmarker) {
		b.maxIterations = n
	}
}

// WithMinDuration sets the minimum benchmark duration.
func WithMinDuration(d time.Duration) BenchmarkerOption {
	return func(b *Benchmarker) {
		b.minDuration = d
	}
}

// WithMaxDuration sets the maximum benchmark duration.
func WithMaxDuration(d time.Duration) BenchmarkerOption {
	return func(b *Benchmarker) {
		b.maxDuration = d
	}
}

// WithGCBetweenRuns enables GC between benchmark runs.
func WithGCBetweenRuns(enabled bool) BenchmarkerOption {
	return func(b *Benchmarker) {
		b.gcBetweenRuns = enabled
	}
}

// NewBenchmarker creates a new Benchmarker with the given options.
func NewBenchmarker(opts ...BenchmarkerOption) *Benchmarker {
	b := &Benchmarker{
		warmupIterations: 10,
		minIterations:    100,
		maxIterations:    1000000,
		minDuration:      time.Second,
		maxDuration:      10 * time.Second,
		gcBetweenRuns:    true,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Run runs a benchmark and returns the result.
func (b *Benchmarker) Run(name string, fn func()) *BenchmarkResult {
	// Warmup phase
	for i := 0; i < b.warmupIterations; i++ {
		fn()
	}

	if b.gcBetweenRuns {
		runtime.GC()
	}

	// Determine iteration count
	iterations := b.determineIterations(fn)

	// Collect timing samples
	times := make([]time.Duration, iterations)

	// Memory stats before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	totalStart := time.Now()
	for i := 0; i < iterations; i++ {
		start := time.Now()
		fn()
		times[i] = time.Since(start)
	}
	totalTime := time.Since(totalStart)

	// Memory stats after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate statistics
	return b.calculateStats(name, times, totalTime, &memBefore, &memAfter)
}

// RunWithSetup runs a benchmark with setup and teardown functions.
func (b *Benchmarker) RunWithSetup(name string, setup func(), fn func(), teardown func()) *BenchmarkResult {
	// Warmup phase
	for i := 0; i < b.warmupIterations; i++ {
		if setup != nil {
			setup()
		}
		fn()
		if teardown != nil {
			teardown()
		}
	}

	if b.gcBetweenRuns {
		runtime.GC()
	}

	// Determine iteration count (without setup/teardown overhead)
	iterations := b.determineIterationsWithSetup(setup, fn, teardown)

	// Collect timing samples (only timing the actual function)
	times := make([]time.Duration, iterations)

	// Memory stats before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	totalStart := time.Now()
	for i := 0; i < iterations; i++ {
		if setup != nil {
			setup()
		}
		start := time.Now()
		fn()
		times[i] = time.Since(start)
		if teardown != nil {
			teardown()
		}
	}
	totalTime := time.Since(totalStart)

	// Memory stats after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	return b.calculateStats(name, times, totalTime, &memBefore, &memAfter)
}

// RunParallel runs a benchmark in parallel with the given number of goroutines.
func (b *Benchmarker) RunParallel(name string, goroutines int, fn func()) *BenchmarkResult {
	if goroutines <= 0 {
		goroutines = runtime.GOMAXPROCS(0)
	}

	// Warmup phase
	for i := 0; i < b.warmupIterations; i++ {
		fn()
	}

	if b.gcBetweenRuns {
		runtime.GC()
	}

	// Determine iterations per goroutine
	totalIterations := b.determineIterations(fn)
	iterationsPerGoroutine := totalIterations / goroutines
	if iterationsPerGoroutine < 1 {
		iterationsPerGoroutine = 1
	}

	// Collect timing samples from all goroutines
	allTimes := make([][]time.Duration, goroutines)
	var wg sync.WaitGroup

	// Memory stats before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	totalStart := time.Now()
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gIdx int) {
			defer wg.Done()
			times := make([]time.Duration, iterationsPerGoroutine)
			for i := 0; i < iterationsPerGoroutine; i++ {
				start := time.Now()
				fn()
				times[i] = time.Since(start)
			}
			allTimes[gIdx] = times
		}(g)
	}
	wg.Wait()
	totalTime := time.Since(totalStart)

	// Memory stats after
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Flatten all times
	var times []time.Duration
	for _, t := range allTimes {
		times = append(times, t...)
	}

	result := b.calculateStats(name+" (parallel)", times, totalTime, &memBefore, &memAfter)
	return result
}

func (b *Benchmarker) determineIterations(fn func()) int {
	// Start with a small number and increase until we reach minimum duration
	iterations := 1
	for iterations < b.maxIterations {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			fn()
		}
		elapsed := time.Since(start)

		if elapsed >= b.minDuration {
			return iterations
		}

		// Estimate iterations needed
		if elapsed > 0 {
			estimated := int(float64(iterations) * float64(b.minDuration) / float64(elapsed))
			if estimated > iterations*2 {
				iterations = estimated
			} else {
				iterations *= 2
			}
		} else {
			iterations *= 10
		}

		if iterations > b.maxIterations {
			iterations = b.maxIterations
		}
	}
	return iterations
}

func (b *Benchmarker) determineIterationsWithSetup(setup, fn, teardown func()) int {
	iterations := 1
	for iterations < b.maxIterations {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			if setup != nil {
				setup()
			}
			fn()
			if teardown != nil {
				teardown()
			}
		}
		elapsed := time.Since(start)

		if elapsed >= b.minDuration {
			return iterations
		}

		if elapsed > 0 {
			estimated := int(float64(iterations) * float64(b.minDuration) / float64(elapsed))
			if estimated > iterations*2 {
				iterations = estimated
			} else {
				iterations *= 2
			}
		} else {
			iterations *= 10
		}

		if iterations > b.maxIterations {
			iterations = b.maxIterations
		}
	}
	return iterations
}

func (b *Benchmarker) calculateStats(name string, times []time.Duration, totalTime time.Duration, memBefore, memAfter *runtime.MemStats) *BenchmarkResult {
	n := len(times)
	if n == 0 {
		return &BenchmarkResult{Name: name}
	}

	// Sort for percentile calculations
	sorted := make([]time.Duration, n)
	copy(sorted, times)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// Calculate mean
	var sum time.Duration
	for _, t := range times {
		sum += t
	}
	mean := sum / time.Duration(n)

	// Calculate standard deviation
	var sumSquares float64
	for _, t := range times {
		diff := float64(t - mean)
		sumSquares += diff * diff
	}
	stdDev := time.Duration(math.Sqrt(sumSquares / float64(n)))

	// Calculate percentiles
	p50 := sorted[int(float64(n)*0.50)]
	p90 := sorted[int(float64(n)*0.90)]
	p95 := sorted[int(float64(n)*0.95)]
	p99Idx := int(float64(n) * 0.99)
	if p99Idx >= n {
		p99Idx = n - 1
	}
	p99 := sorted[p99Idx]

	// Calculate ops per second
	opsPerSecond := float64(n) / totalTime.Seconds()

	// Calculate memory allocations
	allocBytes := int64(memAfter.TotalAlloc - memBefore.TotalAlloc)
	allocObjects := int64(memAfter.Mallocs - memBefore.Mallocs)

	return &BenchmarkResult{
		Name:         name,
		Iterations:   n,
		TotalTime:    totalTime,
		MinTime:      sorted[0],
		MaxTime:      sorted[n-1],
		MeanTime:     mean,
		MedianTime:   sorted[n/2],
		StdDev:       stdDev,
		P50:          p50,
		P90:          p90,
		P95:          p95,
		P99:          p99,
		OpsPerSecond: opsPerSecond,
		AllocBytes:   allocBytes,
		AllocObjects: allocObjects,
	}
}

// Compare compares two benchmark results and returns a comparison.
type BenchmarkComparison struct {
	Before       *BenchmarkResult
	After        *BenchmarkResult
	SpeedupRatio float64 // > 1 means faster, < 1 means slower
	MemoryRatio  float64 // > 1 means more memory, < 1 means less memory
	Improvement  string  // Human-readable improvement description
}

// Compare compares two benchmark results.
func Compare(before, after *BenchmarkResult) *BenchmarkComparison {
	speedup := float64(before.MeanTime) / float64(after.MeanTime)

	var memoryRatio float64
	if before.AllocBytes > 0 {
		memoryRatio = float64(after.AllocBytes) / float64(before.AllocBytes)
	}

	var improvement string
	if speedup > 1 {
		improvement = fmt.Sprintf("%.2fx faster", speedup)
	} else if speedup < 1 {
		improvement = fmt.Sprintf("%.2fx slower", 1/speedup)
	} else {
		improvement = "no change"
	}

	if memoryRatio < 1 {
		improvement += fmt.Sprintf(", %.2fx less memory", 1/memoryRatio)
	} else if memoryRatio > 1 {
		improvement += fmt.Sprintf(", %.2fx more memory", memoryRatio)
	}

	return &BenchmarkComparison{
		Before:       before,
		After:        after,
		SpeedupRatio: speedup,
		MemoryRatio:  memoryRatio,
		Improvement:  improvement,
	}
}

// String returns a human-readable comparison.
func (c *BenchmarkComparison) String() string {
	return fmt.Sprintf(
		"Comparison: %s vs %s\n"+
			"  Before Mean:   %v\n"+
			"  After Mean:    %v\n"+
			"  Speedup:       %.2fx\n"+
			"  Memory Ratio:  %.2fx\n"+
			"  Summary:       %s",
		c.Before.Name,
		c.After.Name,
		c.Before.MeanTime,
		c.After.MeanTime,
		c.SpeedupRatio,
		c.MemoryRatio,
		c.Improvement,
	)
}

// BenchmarkSuite runs multiple benchmarks and collects results.
type BenchmarkSuite struct {
	benchmarker *Benchmarker
	results     []*BenchmarkResult
	mu          sync.Mutex
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(opts ...BenchmarkerOption) *BenchmarkSuite {
	return &BenchmarkSuite{
		benchmarker: NewBenchmarker(opts...),
		results:     make([]*BenchmarkResult, 0),
	}
}

// Add adds a benchmark to the suite and runs it.
func (s *BenchmarkSuite) Add(name string, fn func()) *BenchmarkResult {
	result := s.benchmarker.Run(name, fn)
	s.mu.Lock()
	s.results = append(s.results, result)
	s.mu.Unlock()
	return result
}

// Results returns all benchmark results.
func (s *BenchmarkSuite) Results() []*BenchmarkResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	results := make([]*BenchmarkResult, len(s.results))
	copy(results, s.results)
	return results
}

// Summary returns a summary of all benchmarks.
func (s *BenchmarkSuite) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.results) == 0 {
		return "No benchmarks run"
	}

	var summary string
	summary += "Benchmark Suite Summary\n"
	summary += "=======================\n\n"

	for _, r := range s.results {
		summary += fmt.Sprintf("%-30s  Mean: %-12v  Ops/sec: %-12.2f  Alloc: %d bytes\n",
			r.Name, r.MeanTime, r.OpsPerSecond, r.AllocBytes)
	}

	return summary
}

// QuickBench is a convenience function for quick benchmarking.
func QuickBench(name string, fn func()) *BenchmarkResult {
	b := NewBenchmarker(
		WithMinDuration(100*time.Millisecond),
		WithWarmupIterations(5),
	)
	return b.Run(name, fn)
}

// CompareFuncs compares two functions and returns the comparison.
func CompareFuncs(name1 string, fn1 func(), name2 string, fn2 func()) *BenchmarkComparison {
	b := NewBenchmarker()
	result1 := b.Run(name1, fn1)
	result2 := b.Run(name2, fn2)
	return Compare(result1, result2)
}
