package profiling

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if !config.EnableHTTP {
		t.Error("Expected EnableHTTP to be true")
	}

	if config.HTTPAddr != ":6060" {
		t.Errorf("Expected HTTPAddr ':6060', got '%s'", config.HTTPAddr)
	}

	if config.CPUProfileDuration != 30*time.Second {
		t.Errorf("Expected CPUProfileDuration 30s, got %v", config.CPUProfileDuration)
	}
}

func TestNew(t *testing.T) {
	// Test with nil config
	p := New(nil)
	if p == nil {
		t.Fatal("New(nil) returned nil")
	}
	if p.config == nil {
		t.Error("Expected default config to be set")
	}

	// Test with custom config
	config := &Config{
		EnableHTTP: false,
		OutputDir:  "/tmp/test-profiles",
	}
	p = New(config)
	if p.config.EnableHTTP {
		t.Error("Expected EnableHTTP to be false")
	}
}

func TestProfiler_StartStop(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false, // Disable HTTP for testing
		OutputDir:  tmpDir,
	}

	p := New(config)

	// Start profiler
	err := p.Start()
	if err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}

	if !p.isRunning.Load() {
		t.Error("Expected profiler to be running")
	}

	// Try to start again - should fail
	err = p.Start()
	if err == nil {
		t.Error("Expected error when starting profiler twice")
	}

	// Stop profiler
	err = p.Stop()
	if err != nil {
		t.Fatalf("Failed to stop profiler: %v", err)
	}

	if p.isRunning.Load() {
		t.Error("Expected profiler to be stopped")
	}

	// Stop again - should be no-op
	err = p.Stop()
	if err != nil {
		t.Fatalf("Stop should be idempotent: %v", err)
	}
}

func TestProfiler_CPUProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false,
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Start CPU profiling
	err := p.StartCPUProfile()
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}

	// Do some work
	sum := 0
	for i := 0; i < 1000000; i++ {
		sum += i
	}
	_ = sum

	// Stop CPU profiling
	err = p.StopCPUProfile()
	if err != nil {
		t.Fatalf("Failed to stop CPU profile: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "cpu-*.pprof"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected CPU profile file to be created")
	}
}

func TestProfiler_CPUProfile_DoubleStart(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false,
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Start CPU profiling
	err := p.StartCPUProfile()
	if err != nil {
		t.Fatalf("Failed to start CPU profile: %v", err)
	}
	defer p.StopCPUProfile()

	// Try to start again - should fail
	err = p.StartCPUProfile()
	if err == nil {
		t.Error("Expected error when starting CPU profile twice")
	}
}

func TestProfiler_HeapProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false,
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Capture heap profile
	err := p.CaptureHeapProfile()
	if err != nil {
		t.Fatalf("Failed to capture heap profile: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "heap-*.pprof"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected heap profile file to be created")
	}
}

func TestProfiler_GoroutineProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false,
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Capture goroutine profile
	err := p.CaptureGoroutineProfile()
	if err != nil {
		t.Fatalf("Failed to capture goroutine profile: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "goroutine-*.pprof"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected goroutine profile file to be created")
	}
}

func TestProfiler_BlockProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP:       false,
		OutputDir:        tmpDir,
		BlockProfileRate: 1,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Capture block profile
	err := p.CaptureBlockProfile()
	if err != nil {
		t.Fatalf("Failed to capture block profile: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "block-*.pprof"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected block profile file to be created")
	}
}

func TestProfiler_MutexProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP:           false,
		OutputDir:            tmpDir,
		MutexProfileFraction: 1,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Capture mutex profile
	err := p.CaptureMutexProfile()
	if err != nil {
		t.Fatalf("Failed to capture mutex profile: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "mutex-*.pprof"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected mutex profile file to be created")
	}
}

func TestProfiler_WriteProfileTo(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false,
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	var buf bytes.Buffer
	err := p.WriteProfileTo("goroutine", &buf)
	if err != nil {
		t.Fatalf("Failed to write profile: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected non-empty profile data")
	}
}

func TestProfiler_WriteProfileTo_InvalidProfile(t *testing.T) {
	p := New(nil)

	var buf bytes.Buffer
	err := p.WriteProfileTo("nonexistent", &buf)
	if err == nil {
		t.Error("Expected error for nonexistent profile")
	}
}

func TestProfiler_GetMemStats(t *testing.T) {
	p := New(nil)

	stats := p.GetMemStats()
	if stats == nil {
		t.Fatal("Expected non-nil MemStats")
	}

	if stats.Sys == 0 {
		t.Error("Expected non-zero Sys memory")
	}
}

func TestProfiler_GetRuntimeStats(t *testing.T) {
	p := New(nil)

	stats := p.GetRuntimeStats()
	if stats == nil {
		t.Fatal("Expected non-nil RuntimeStats")
	}

	if stats.NumGoroutine <= 0 {
		t.Error("Expected positive goroutine count")
	}

	if stats.NumCPU <= 0 {
		t.Error("Expected positive CPU count")
	}

	if stats.GOMAXPROCS <= 0 {
		t.Error("Expected positive GOMAXPROCS")
	}
}

func TestRuntimeStats_String(t *testing.T) {
	stats := &RuntimeStats{
		NumGoroutine: 10,
		NumCPU:       4,
		GOMAXPROCS:   4,
		HeapAlloc:    1024 * 1024,
		HeapSys:      2048 * 1024,
		NumGC:        5,
	}

	str := stats.String()
	if str == "" {
		t.Error("Expected non-empty string")
	}

	if !strings.Contains(str, "Goroutines: 10") {
		t.Error("Expected string to contain goroutine count")
	}
}

func TestProfiler_HTTPServer(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: true,
		HTTPAddr:   ":0", // Use random port
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Note: We can't easily test the HTTP endpoints without knowing the port
	// In a real test, we'd use httptest.Server or extract the port
}

func TestProfiler_AutoProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP:          false,
		OutputDir:           tmpDir,
		AutoProfile:         true,
		AutoProfileInterval: 100 * time.Millisecond,
		MemoryThreshold:     1, // Very low threshold to trigger
		GoroutineThreshold:  1, // Very low threshold to trigger
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}

	// Wait for auto-profile to trigger
	time.Sleep(200 * time.Millisecond)

	p.Stop()

	// Check if profiles were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected auto-profile to create files")
	}
}

func TestProfiler_SetLogger(t *testing.T) {
	p := New(nil)

	// Create a custom logger
	var logBuf bytes.Buffer
	customLogger := &testLogger{w: &logBuf}

	p.SetLogger(customLogger)

	// Trigger some logging
	tmpDir := t.TempDir()
	p.config.OutputDir = tmpDir
	p.config.EnableHTTP = false

	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	p.Stop()

	if logBuf.Len() == 0 {
		t.Error("Expected custom logger to receive messages")
	}
}

type testLogger struct {
	w io.Writer
}

func (l *testLogger) Printf(format string, v ...interface{}) {
	fmt.Fprintf(l.w, format+"\n", v...)
}

func TestProfiler_CaptureAllProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP:           false,
		OutputDir:            tmpDir,
		BlockProfileRate:     1,
		MutexProfileFraction: 1,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Capture all profiles with short CPU duration
	err := p.CaptureAllProfiles(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to capture all profiles: %v", err)
	}

	// Verify files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read dir: %v", err)
	}

	expectedTypes := []string{"cpu", "heap", "goroutine", "block", "mutex"}
	for _, expectedType := range expectedTypes {
		found := false
		for _, f := range files {
			if strings.HasPrefix(f.Name(), expectedType+"-") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s profile file to be created", expectedType)
		}
	}
}

func TestProfiler_CaptureCPUProfile(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		EnableHTTP: false,
		OutputDir:  tmpDir,
	}

	p := New(config)
	if err := p.Start(); err != nil {
		t.Fatalf("Failed to start profiler: %v", err)
	}
	defer p.Stop()

	// Capture CPU profile for short duration
	err := p.CaptureCPUProfile(50 * time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to capture CPU profile: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tmpDir, "cpu-*.pprof"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}
	if len(files) == 0 {
		t.Error("Expected CPU profile file to be created")
	}
}

// Benchmarks

func BenchmarkProfiler_GetMemStats(b *testing.B) {
	p := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.GetMemStats()
	}
}

func BenchmarkProfiler_GetRuntimeStats(b *testing.B) {
	p := New(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.GetRuntimeStats()
	}
}

func BenchmarkRuntimeStats_String(b *testing.B) {
	stats := &RuntimeStats{
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		HeapAlloc:    1024 * 1024,
		HeapSys:      2048 * 1024,
		NumGC:        5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stats.String()
	}
}
