// Package perfutil provides production-ready performance utilities for Go applications.
// It includes profiling, benchmarking, and metrics collection tools.
package perfutil

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

// ProfileType represents the type of profile to collect.
type ProfileType int

const (
	// ProfileCPU collects CPU profiling data.
	ProfileCPU ProfileType = iota
	// ProfileMemory collects memory allocation profiling data.
	ProfileMemory
	// ProfileGoroutine collects goroutine profiling data.
	ProfileGoroutine
	// ProfileBlock collects blocking profiling data.
	ProfileBlock
	// ProfileMutex collects mutex contention profiling data.
	ProfileMutex
)

// String returns the string representation of ProfileType.
func (p ProfileType) String() string {
	switch p {
	case ProfileCPU:
		return "cpu"
	case ProfileMemory:
		return "memory"
	case ProfileGoroutine:
		return "goroutine"
	case ProfileBlock:
		return "block"
	case ProfileMutex:
		return "mutex"
	default:
		return "unknown"
	}
}

// Profiler provides a unified interface for collecting various runtime profiles.
type Profiler struct {
	mu          sync.Mutex
	outputDir   string
	cpuFile     *os.File
	isRunning   atomic.Bool
	startTime   time.Time
	profileRate int
}

// ProfilerOption configures a Profiler.
type ProfilerOption func(*Profiler)

// WithOutputDir sets the output directory for profile files.
func WithOutputDir(dir string) ProfilerOption {
	return func(p *Profiler) {
		p.outputDir = dir
	}
}

// WithProfileRate sets the CPU profile sampling rate (samples per second).
func WithProfileRate(rate int) ProfilerOption {
	return func(p *Profiler) {
		p.profileRate = rate
	}
}

// NewProfiler creates a new Profiler with the given options.
func NewProfiler(opts ...ProfilerOption) *Profiler {
	p := &Profiler{
		outputDir:   ".",
		profileRate: 100, // Default: 100 samples per second
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// StartCPUProfile starts CPU profiling and writes to the specified writer.
func (p *Profiler) StartCPUProfile(w io.Writer) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning.Load() {
		return fmt.Errorf("profiler: CPU profiling already running")
	}

	runtime.SetCPUProfileRate(p.profileRate)
	if err := pprof.StartCPUProfile(w); err != nil {
		return fmt.Errorf("profiler: failed to start CPU profile: %w", err)
	}

	p.isRunning.Store(true)
	p.startTime = time.Now()
	return nil
}

// StartCPUProfileToFile starts CPU profiling and writes to a file.
func (p *Profiler) StartCPUProfileToFile(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning.Load() {
		return fmt.Errorf("profiler: CPU profiling already running")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("profiler: failed to create CPU profile file: %w", err)
	}

	runtime.SetCPUProfileRate(p.profileRate)
	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("profiler: failed to start CPU profile: %w", err)
	}

	p.cpuFile = f
	p.isRunning.Store(true)
	p.startTime = time.Now()
	return nil
}

// StopCPUProfile stops CPU profiling.
func (p *Profiler) StopCPUProfile() (time.Duration, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isRunning.Load() {
		return 0, fmt.Errorf("profiler: CPU profiling not running")
	}

	pprof.StopCPUProfile()
	duration := time.Since(p.startTime)

	if p.cpuFile != nil {
		if err := p.cpuFile.Close(); err != nil {
			return duration, fmt.Errorf("profiler: failed to close CPU profile file: %w", err)
		}
		p.cpuFile = nil
	}

	p.isRunning.Store(false)
	return duration, nil
}

// WriteMemoryProfile writes a memory profile to the specified writer.
func (p *Profiler) WriteMemoryProfile(w io.Writer) error {
	runtime.GC() // Get up-to-date statistics
	if err := pprof.WriteHeapProfile(w); err != nil {
		return fmt.Errorf("profiler: failed to write memory profile: %w", err)
	}
	return nil
}

// WriteMemoryProfileToFile writes a memory profile to a file.
func (p *Profiler) WriteMemoryProfileToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("profiler: failed to create memory profile file: %w", err)
	}
	defer f.Close()

	return p.WriteMemoryProfile(f)
}

// WriteGoroutineProfile writes a goroutine profile to the specified writer.
func (p *Profiler) WriteGoroutineProfile(w io.Writer, debug int) error {
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("profiler: goroutine profile not found")
	}
	if err := profile.WriteTo(w, debug); err != nil {
		return fmt.Errorf("profiler: failed to write goroutine profile: %w", err)
	}
	return nil
}

// WriteGoroutineProfileToFile writes a goroutine profile to a file.
func (p *Profiler) WriteGoroutineProfileToFile(filename string, debug int) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("profiler: failed to create goroutine profile file: %w", err)
	}
	defer f.Close()

	return p.WriteGoroutineProfile(f, debug)
}

// WriteBlockProfile writes a block profile to the specified writer.
func (p *Profiler) WriteBlockProfile(w io.Writer, debug int) error {
	profile := pprof.Lookup("block")
	if profile == nil {
		return fmt.Errorf("profiler: block profile not found")
	}
	if err := profile.WriteTo(w, debug); err != nil {
		return fmt.Errorf("profiler: failed to write block profile: %w", err)
	}
	return nil
}

// WriteMutexProfile writes a mutex profile to the specified writer.
func (p *Profiler) WriteMutexProfile(w io.Writer, debug int) error {
	profile := pprof.Lookup("mutex")
	if profile == nil {
		return fmt.Errorf("profiler: mutex profile not found")
	}
	if err := profile.WriteTo(w, debug); err != nil {
		return fmt.Errorf("profiler: failed to write mutex profile: %w", err)
	}
	return nil
}

// EnableBlockProfiling enables block profiling with the given rate.
// Rate controls the fraction of blocking events that are reported.
// A rate of 1 reports every blocking event.
func (p *Profiler) EnableBlockProfiling(rate int) {
	runtime.SetBlockProfileRate(rate)
}

// DisableBlockProfiling disables block profiling.
func (p *Profiler) DisableBlockProfiling() {
	runtime.SetBlockProfileRate(0)
}

// EnableMutexProfiling enables mutex profiling with the given rate.
// Rate controls the fraction of mutex contention events that are reported.
func (p *Profiler) EnableMutexProfiling(rate int) {
	runtime.SetMutexProfileFraction(rate)
}

// DisableMutexProfiling disables mutex profiling.
func (p *Profiler) DisableMutexProfiling() {
	runtime.SetMutexProfileFraction(0)
}

// MemStats returns current memory statistics.
func (p *Profiler) MemStats() *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &m
}

// MemStatsSnapshot represents a snapshot of memory statistics.
type MemStatsSnapshot struct {
	Timestamp     time.Time
	Alloc         uint64  // Bytes allocated and still in use
	TotalAlloc    uint64  // Bytes allocated (even if freed)
	Sys           uint64  // Bytes obtained from system
	NumGC         uint32  // Number of completed GC cycles
	HeapAlloc     uint64  // Bytes allocated on heap
	HeapSys       uint64  // Bytes obtained from system for heap
	HeapIdle      uint64  // Bytes in idle spans
	HeapInuse     uint64  // Bytes in non-idle spans
	HeapReleased  uint64  // Bytes released to OS
	HeapObjects   uint64  // Number of allocated objects
	StackInuse    uint64  // Bytes used by stack allocator
	StackSys      uint64  // Bytes obtained from system for stack
	MSpanInuse    uint64  // Bytes used by mspan structures
	MCacheInuse   uint64  // Bytes used by mcache structures
	GCCPUFraction float64 // Fraction of CPU used by GC
	NumGoroutine  int     // Number of goroutines
}

// TakeMemStatsSnapshot takes a snapshot of current memory statistics.
func (p *Profiler) TakeMemStatsSnapshot() *MemStatsSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemStatsSnapshot{
		Timestamp:     time.Now(),
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		NumGC:         m.NumGC,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MCacheInuse:   m.MCacheInuse,
		GCCPUFraction: m.GCCPUFraction,
		NumGoroutine:  runtime.NumGoroutine(),
	}
}

// CompareMemStats compares two memory snapshots and returns the difference.
func CompareMemStats(before, after *MemStatsSnapshot) *MemStatsDiff {
	return &MemStatsDiff{
		Duration:        after.Timestamp.Sub(before.Timestamp),
		AllocDiff:       int64(after.Alloc) - int64(before.Alloc),
		TotalAllocDiff:  int64(after.TotalAlloc) - int64(before.TotalAlloc),
		SysDiff:         int64(after.Sys) - int64(before.Sys),
		NumGCDiff:       int32(after.NumGC) - int32(before.NumGC),
		HeapAllocDiff:   int64(after.HeapAlloc) - int64(before.HeapAlloc),
		HeapObjectsDiff: int64(after.HeapObjects) - int64(before.HeapObjects),
		GoroutineDiff:   after.NumGoroutine - before.NumGoroutine,
	}
}

// MemStatsDiff represents the difference between two memory snapshots.
type MemStatsDiff struct {
	Duration        time.Duration
	AllocDiff       int64
	TotalAllocDiff  int64
	SysDiff         int64
	NumGCDiff       int32
	HeapAllocDiff   int64
	HeapObjectsDiff int64
	GoroutineDiff   int
}

// String returns a human-readable representation of the memory diff.
func (d *MemStatsDiff) String() string {
	return fmt.Sprintf(
		"Duration: %v\n"+
			"Alloc: %+d bytes\n"+
			"TotalAlloc: %+d bytes\n"+
			"Sys: %+d bytes\n"+
			"NumGC: %+d cycles\n"+
			"HeapAlloc: %+d bytes\n"+
			"HeapObjects: %+d objects\n"+
			"Goroutines: %+d",
		d.Duration,
		d.AllocDiff,
		d.TotalAllocDiff,
		d.SysDiff,
		d.NumGCDiff,
		d.HeapAllocDiff,
		d.HeapObjectsDiff,
		d.GoroutineDiff,
	)
}

// ProfileFunc profiles a function and returns profiling results.
func (p *Profiler) ProfileFunc(name string, fn func()) (*FuncProfile, error) {
	// Take before snapshot
	beforeMem := p.TakeMemStatsSnapshot()

	// Run function and measure time
	start := time.Now()
	fn()
	duration := time.Since(start)

	// Take after snapshot
	afterMem := p.TakeMemStatsSnapshot()

	return &FuncProfile{
		Name:     name,
		Duration: duration,
		MemDiff:  CompareMemStats(beforeMem, afterMem),
	}, nil
}

// FuncProfile contains profiling results for a function.
type FuncProfile struct {
	Name     string
	Duration time.Duration
	MemDiff  *MemStatsDiff
}

// String returns a human-readable representation of the function profile.
func (fp *FuncProfile) String() string {
	return fmt.Sprintf(
		"Function: %s\n"+
			"Duration: %v\n"+
			"Memory Changes:\n%s",
		fp.Name,
		fp.Duration,
		fp.MemDiff.String(),
	)
}

// GoroutineStats returns statistics about current goroutines.
type GoroutineStats struct {
	Count     int
	Timestamp time.Time
}

// GetGoroutineStats returns current goroutine statistics.
func (p *Profiler) GetGoroutineStats() *GoroutineStats {
	return &GoroutineStats{
		Count:     runtime.NumGoroutine(),
		Timestamp: time.Now(),
	}
}

// GoroutineLeakDetector helps detect goroutine leaks.
type GoroutineLeakDetector struct {
	baseline int
	mu       sync.Mutex
}

// NewGoroutineLeakDetector creates a new goroutine leak detector.
func NewGoroutineLeakDetector() *GoroutineLeakDetector {
	return &GoroutineLeakDetector{
		baseline: runtime.NumGoroutine(),
	}
}

// SetBaseline sets the baseline goroutine count.
func (d *GoroutineLeakDetector) SetBaseline() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.baseline = runtime.NumGoroutine()
}

// Check checks for goroutine leaks and returns the difference from baseline.
func (d *GoroutineLeakDetector) Check() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return runtime.NumGoroutine() - d.baseline
}

// CheckWithThreshold checks if goroutine count exceeds baseline by threshold.
func (d *GoroutineLeakDetector) CheckWithThreshold(threshold int) (leaked bool, diff int) {
	diff = d.Check()
	return diff > threshold, diff
}
