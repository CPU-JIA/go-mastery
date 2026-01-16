package perfutil

import (
	"bytes"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestProfiler_CPUProfile(t *testing.T) {
	p := NewProfiler(WithProfileRate(100))

	var buf bytes.Buffer
	err := p.StartCPUProfile(&buf)
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}

	// Do some work
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	_ = sum

	duration, err := p.StopCPUProfile()
	if err != nil {
		t.Fatalf("Failed to stop CPU profile: %v", err)
	}

	if duration <= 0 {
		t.Error("Expected positive duration")
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty CPU profile")
	}
}

func TestProfiler_DoubleStart(t *testing.T) {
	p := NewProfiler()

	var buf bytes.Buffer
	err := p.StartCPUProfile(&buf)
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}

	// Try to start again - should fail
	err = p.StartCPUProfile(&buf)
	if err == nil {
		t.Error("Expected error when starting CPU profile twice")
	}

	_, _ = p.StopCPUProfile()
}

func TestProfiler_StopWithoutStart(t *testing.T) {
	p := NewProfiler()

	_, err := p.StopCPUProfile()
	if err == nil {
		t.Error("Expected error when stopping without starting")
	}
}

func TestProfiler_MemoryProfile(t *testing.T) {
	p := NewProfiler()

	var buf bytes.Buffer
	err := p.WriteMemoryProfile(&buf)
	if err != nil {
		t.Fatalf("Failed to write memory profile: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty memory profile")
	}
}

func TestProfiler_GoroutineProfile(t *testing.T) {
	p := NewProfiler()

	var buf bytes.Buffer
	err := p.WriteGoroutineProfile(&buf, 1)
	if err != nil {
		t.Fatalf("Failed to write goroutine profile: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty goroutine profile")
	}
}

func TestProfiler_MemStats(t *testing.T) {
	p := NewProfiler()

	stats := p.MemStats()
	if stats == nil {
		t.Fatal("Expected non-nil MemStats")
	}

	if stats.Sys == 0 {
		t.Error("Expected non-zero Sys memory")
	}
}

func TestProfiler_MemStatsSnapshot(t *testing.T) {
	p := NewProfiler()

	snapshot := p.TakeMemStatsSnapshot()
	if snapshot == nil {
		t.Fatal("Expected non-nil snapshot")
	}

	if snapshot.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if snapshot.NumGoroutine <= 0 {
		t.Error("Expected positive goroutine count")
	}
}

func TestCompareMemStats(t *testing.T) {
	p := NewProfiler()

	before := p.TakeMemStatsSnapshot()

	// Allocate some memory and add a small delay to ensure measurable duration
	time.Sleep(time.Millisecond)
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	after := p.TakeMemStatsSnapshot()

	diff := CompareMemStats(before, after)
	if diff == nil {
		t.Fatal("Expected non-nil diff")
	}

	if diff.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	// TotalAlloc should increase
	if diff.TotalAllocDiff <= 0 {
		t.Error("Expected positive TotalAllocDiff after allocation")
	}
}

func TestProfiler_ProfileFunc(t *testing.T) {
	p := NewProfiler()

	profile, err := p.ProfileFunc("test_func", func() {
		time.Sleep(10 * time.Millisecond)
	})

	if err != nil {
		t.Fatalf("Failed to profile function: %v", err)
	}

	if profile.Name != "test_func" {
		t.Errorf("Expected name 'test_func', got '%s'", profile.Name)
	}

	if profile.Duration < 10*time.Millisecond {
		t.Errorf("Expected duration >= 10ms, got %v", profile.Duration)
	}
}

func TestGoroutineLeakDetector(t *testing.T) {
	detector := NewGoroutineLeakDetector()

	// Start some goroutines
	var wg sync.WaitGroup
	done := make(chan struct{})

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-done
		}()
	}

	// Give goroutines time to start
	time.Sleep(10 * time.Millisecond)

	diff := detector.Check()
	if diff < 5 {
		t.Errorf("Expected at least 5 new goroutines, got %d", diff)
	}

	// Clean up
	close(done)
	wg.Wait()

	// Allow goroutines to finish
	time.Sleep(10 * time.Millisecond)

	// Check again - should be back to baseline (or close)
	diff = detector.Check()
	if diff > 2 { // Allow some tolerance
		t.Errorf("Expected goroutines to return to baseline, diff: %d", diff)
	}
}

func TestGoroutineLeakDetector_Threshold(t *testing.T) {
	detector := NewGoroutineLeakDetector()

	leaked, diff := detector.CheckWithThreshold(10)
	if leaked {
		t.Errorf("Should not detect leak with threshold 10, diff: %d", diff)
	}
}

func TestProfileType_String(t *testing.T) {
	tests := []struct {
		pt       ProfileType
		expected string
	}{
		{ProfileCPU, "cpu"},
		{ProfileMemory, "memory"},
		{ProfileGoroutine, "goroutine"},
		{ProfileBlock, "block"},
		{ProfileMutex, "mutex"},
		{ProfileType(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.pt.String(); got != tt.expected {
			t.Errorf("ProfileType(%d).String() = %s, want %s", tt.pt, got, tt.expected)
		}
	}
}

func TestProfiler_BlockProfiling(t *testing.T) {
	p := NewProfiler()

	p.EnableBlockProfiling(1)
	defer p.DisableBlockProfiling()

	// Do some blocking operations
	ch := make(chan int)
	go func() {
		time.Sleep(time.Millisecond)
		ch <- 1
	}()
	<-ch

	var buf bytes.Buffer
	err := p.WriteBlockProfile(&buf, 1)
	if err != nil {
		t.Fatalf("Failed to write block profile: %v", err)
	}
}

func TestProfiler_MutexProfiling(t *testing.T) {
	p := NewProfiler()

	p.EnableMutexProfiling(1)
	defer p.DisableMutexProfiling()

	// Do some mutex operations
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			time.Sleep(time.Microsecond)
			mu.Unlock()
		}()
	}
	wg.Wait()

	var buf bytes.Buffer
	err := p.WriteMutexProfile(&buf, 1)
	if err != nil {
		t.Fatalf("Failed to write mutex profile: %v", err)
	}
}

func TestMemStatsDiff_String(t *testing.T) {
	diff := &MemStatsDiff{
		Duration:        time.Second,
		AllocDiff:       1024,
		TotalAllocDiff:  2048,
		SysDiff:         4096,
		NumGCDiff:       1,
		HeapAllocDiff:   1024,
		HeapObjectsDiff: 10,
		GoroutineDiff:   5,
	}

	str := diff.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}

func TestFuncProfile_String(t *testing.T) {
	fp := &FuncProfile{
		Name:     "test",
		Duration: time.Second,
		MemDiff: &MemStatsDiff{
			Duration: time.Second,
		},
	}

	str := fp.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}
}

func TestProfiler_GetGoroutineStats(t *testing.T) {
	p := NewProfiler()

	stats := p.GetGoroutineStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if stats.Count <= 0 {
		t.Error("Expected positive goroutine count")
	}

	if stats.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func BenchmarkProfiler_TakeMemStatsSnapshot(b *testing.B) {
	p := NewProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.TakeMemStatsSnapshot()
	}
}

func BenchmarkProfiler_GetGoroutineStats(b *testing.B) {
	p := NewProfiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.GetGoroutineStats()
	}
}

func BenchmarkGoroutineLeakDetector_Check(b *testing.B) {
	detector := NewGoroutineLeakDetector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detector.Check()
	}
}

func BenchmarkCompareMemStats(b *testing.B) {
	p := NewProfiler()
	before := p.TakeMemStatsSnapshot()
	after := p.TakeMemStatsSnapshot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CompareMemStats(before, after)
	}
}

// Test that profiler options work correctly
func TestProfilerOptions(t *testing.T) {
	p := NewProfiler(
		WithOutputDir("/tmp"),
		WithProfileRate(200),
	)

	if p.outputDir != "/tmp" {
		t.Errorf("Expected outputDir '/tmp', got '%s'", p.outputDir)
	}

	if p.profileRate != 200 {
		t.Errorf("Expected profileRate 200, got %d", p.profileRate)
	}
}

// Test concurrent access to profiler
func TestProfiler_ConcurrentMemStats(t *testing.T) {
	p := NewProfiler()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.TakeMemStatsSnapshot()
		}()
	}
	wg.Wait()
}

// Test that GoroutineLeakDetector is thread-safe
func TestGoroutineLeakDetector_Concurrent(t *testing.T) {
	detector := NewGoroutineLeakDetector()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			detector.SetBaseline()
			_ = detector.Check()
		}()
	}
	wg.Wait()
}

// Ensure runtime package is used
var _ = runtime.NumGoroutine
