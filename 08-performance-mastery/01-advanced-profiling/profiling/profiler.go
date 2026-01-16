// Package profiling provides production-ready profiling helpers for Go applications.
// It includes HTTP handlers for pprof, automatic profiling triggers, and profiling middleware.
package profiling

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	rpprof "runtime/pprof"
	"sync"
	"sync/atomic"
	"time"
)

// Config holds configuration for the profiling system.
type Config struct {
	// EnableHTTP enables HTTP pprof endpoints
	EnableHTTP bool
	// HTTPAddr is the address for the pprof HTTP server (e.g., ":6060")
	HTTPAddr string
	// OutputDir is the directory for profile output files
	OutputDir string
	// CPUProfileDuration is the default duration for CPU profiles
	CPUProfileDuration time.Duration
	// MemProfileRate sets the memory profiling rate (0 = default)
	MemProfileRate int
	// BlockProfileRate sets the block profiling rate (0 = disabled)
	BlockProfileRate int
	// MutexProfileFraction sets the mutex profiling fraction (0 = disabled)
	MutexProfileFraction int
	// AutoProfile enables automatic profiling based on thresholds
	AutoProfile bool
	// AutoProfileInterval is the interval for checking auto-profile conditions
	AutoProfileInterval time.Duration
	// MemoryThreshold triggers memory profile when heap exceeds this (bytes)
	MemoryThreshold uint64
	// GoroutineThreshold triggers goroutine profile when count exceeds this
	GoroutineThreshold int
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		EnableHTTP:           true,
		HTTPAddr:             ":6060",
		OutputDir:            "./profiles",
		CPUProfileDuration:   30 * time.Second,
		MemProfileRate:       0,
		BlockProfileRate:     0,
		MutexProfileFraction: 0,
		AutoProfile:          false,
		AutoProfileInterval:  time.Minute,
		MemoryThreshold:      1 << 30, // 1GB
		GoroutineThreshold:   10000,
	}
}

// Profiler provides production-ready profiling capabilities.
type Profiler struct {
	config     *Config
	httpServer *http.Server
	isRunning  atomic.Bool
	cpuFile    *os.File
	cpuMu      sync.Mutex
	cancelFunc context.CancelFunc
	logger     Logger
}

// Logger interface for profiler logging.
type Logger interface {
	Printf(format string, v ...interface{})
}

// defaultLogger wraps the standard log package.
type defaultLogger struct{}

func (d *defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf("[profiler] "+format, v...)
}

// New creates a new Profiler with the given configuration.
func New(config *Config) *Profiler {
	if config == nil {
		config = DefaultConfig()
	}
	return &Profiler{
		config: config,
		logger: &defaultLogger{},
	}
}

// SetLogger sets a custom logger.
func (p *Profiler) SetLogger(logger Logger) {
	p.logger = logger
}

// Start starts the profiler with all configured features.
func (p *Profiler) Start() error {
	if p.isRunning.Load() {
		return fmt.Errorf("profiler already running")
	}

	// Create output directory
	if err := os.MkdirAll(p.config.OutputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Configure profiling rates
	if p.config.MemProfileRate > 0 {
		runtime.MemProfileRate = p.config.MemProfileRate
	}
	if p.config.BlockProfileRate > 0 {
		runtime.SetBlockProfileRate(p.config.BlockProfileRate)
	}
	if p.config.MutexProfileFraction > 0 {
		runtime.SetMutexProfileFraction(p.config.MutexProfileFraction)
	}

	// Start HTTP server if enabled
	if p.config.EnableHTTP {
		if err := p.startHTTPServer(); err != nil {
			return fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}

	// Start auto-profiling if enabled
	if p.config.AutoProfile {
		ctx, cancel := context.WithCancel(context.Background())
		p.cancelFunc = cancel
		go p.autoProfileLoop(ctx)
	}

	p.isRunning.Store(true)
	p.logger.Printf("Profiler started (HTTP: %s, AutoProfile: %v)",
		p.config.HTTPAddr, p.config.AutoProfile)

	return nil
}

// Stop stops the profiler and cleans up resources.
func (p *Profiler) Stop() error {
	if !p.isRunning.Load() {
		return nil
	}

	// Stop auto-profiling
	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	// Stop CPU profiling if running
	p.StopCPUProfile()

	// Stop HTTP server
	if p.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := p.httpServer.Shutdown(ctx); err != nil {
			p.logger.Printf("HTTP server shutdown error: %v", err)
		}
	}

	// Reset profiling rates
	runtime.SetBlockProfileRate(0)
	runtime.SetMutexProfileFraction(0)

	p.isRunning.Store(false)
	p.logger.Printf("Profiler stopped")

	return nil
}

// startHTTPServer starts the pprof HTTP server.
func (p *Profiler) startHTTPServer() error {
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Custom handlers
	mux.HandleFunc("/debug/pprof/heap", func(w http.ResponseWriter, r *http.Request) {
		pprof.Handler("heap").ServeHTTP(w, r)
	})
	mux.HandleFunc("/debug/pprof/goroutine", func(w http.ResponseWriter, r *http.Request) {
		pprof.Handler("goroutine").ServeHTTP(w, r)
	})
	mux.HandleFunc("/debug/pprof/block", func(w http.ResponseWriter, r *http.Request) {
		pprof.Handler("block").ServeHTTP(w, r)
	})
	mux.HandleFunc("/debug/pprof/mutex", func(w http.ResponseWriter, r *http.Request) {
		pprof.Handler("mutex").ServeHTTP(w, r)
	})

	// Status endpoint
	mux.HandleFunc("/debug/pprof/status", p.statusHandler)

	p.httpServer = &http.Server{
		Addr:              p.config.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := p.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			p.logger.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// statusHandler returns profiler status.
func (p *Profiler) statusHandler(w http.ResponseWriter, _ *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	status := fmt.Sprintf(`Profiler Status
===============
Running: %v
Goroutines: %d
Heap Alloc: %d MB
Heap Sys: %d MB
Heap Objects: %d
GC Cycles: %d
CPU Profile Active: %v
`,
		p.isRunning.Load(),
		runtime.NumGoroutine(),
		m.HeapAlloc/1024/1024,
		m.HeapSys/1024/1024,
		m.HeapObjects,
		m.NumGC,
		p.cpuFile != nil,
	)

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(status))
}

// autoProfileLoop runs the auto-profiling check loop.
func (p *Profiler) autoProfileLoop(ctx context.Context) {
	ticker := time.NewTicker(p.config.AutoProfileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkAutoProfile()
		}
	}
}

// checkAutoProfile checks conditions and triggers profiling if needed.
func (p *Profiler) checkAutoProfile() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Check memory threshold
	if m.HeapAlloc > p.config.MemoryThreshold {
		p.logger.Printf("Memory threshold exceeded (%d > %d), capturing heap profile",
			m.HeapAlloc, p.config.MemoryThreshold)
		if err := p.CaptureHeapProfile(); err != nil {
			p.logger.Printf("Failed to capture heap profile: %v", err)
		}
	}

	// Check goroutine threshold
	numGoroutines := runtime.NumGoroutine()
	if numGoroutines > p.config.GoroutineThreshold {
		p.logger.Printf("Goroutine threshold exceeded (%d > %d), capturing goroutine profile",
			numGoroutines, p.config.GoroutineThreshold)
		if err := p.CaptureGoroutineProfile(); err != nil {
			p.logger.Printf("Failed to capture goroutine profile: %v", err)
		}
	}
}

// StartCPUProfile starts CPU profiling.
func (p *Profiler) StartCPUProfile() error {
	p.cpuMu.Lock()
	defer p.cpuMu.Unlock()

	if p.cpuFile != nil {
		return fmt.Errorf("CPU profiling already running")
	}

	filename := p.generateFilename("cpu")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile file: %w", err)
	}

	if err := rpprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	p.cpuFile = f
	p.logger.Printf("CPU profiling started: %s", filename)

	return nil
}

// StopCPUProfile stops CPU profiling.
func (p *Profiler) StopCPUProfile() error {
	p.cpuMu.Lock()
	defer p.cpuMu.Unlock()

	if p.cpuFile == nil {
		return nil
	}

	rpprof.StopCPUProfile()
	if err := p.cpuFile.Close(); err != nil {
		return fmt.Errorf("failed to close CPU profile file: %w", err)
	}

	p.logger.Printf("CPU profiling stopped")
	p.cpuFile = nil

	return nil
}

// CaptureCPUProfile captures a CPU profile for the specified duration.
func (p *Profiler) CaptureCPUProfile(duration time.Duration) error {
	if err := p.StartCPUProfile(); err != nil {
		return err
	}

	time.Sleep(duration)

	return p.StopCPUProfile()
}

// CaptureHeapProfile captures a heap memory profile.
func (p *Profiler) CaptureHeapProfile() error {
	runtime.GC() // Get up-to-date statistics

	filename := p.generateFilename("heap")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create heap profile file: %w", err)
	}
	defer f.Close()

	if err := rpprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	p.logger.Printf("Heap profile captured: %s", filename)
	return nil
}

// CaptureGoroutineProfile captures a goroutine profile.
func (p *Profiler) CaptureGoroutineProfile() error {
	filename := p.generateFilename("goroutine")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create goroutine profile file: %w", err)
	}
	defer f.Close()

	profile := rpprof.Lookup("goroutine")
	if profile == nil {
		return fmt.Errorf("goroutine profile not found")
	}

	if err := profile.WriteTo(f, 2); err != nil {
		return fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	p.logger.Printf("Goroutine profile captured: %s", filename)
	return nil
}

// CaptureBlockProfile captures a block profile.
func (p *Profiler) CaptureBlockProfile() error {
	filename := p.generateFilename("block")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create block profile file: %w", err)
	}
	defer f.Close()

	profile := rpprof.Lookup("block")
	if profile == nil {
		return fmt.Errorf("block profile not found")
	}

	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %w", err)
	}

	p.logger.Printf("Block profile captured: %s", filename)
	return nil
}

// CaptureMutexProfile captures a mutex contention profile.
func (p *Profiler) CaptureMutexProfile() error {
	filename := p.generateFilename("mutex")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create mutex profile file: %w", err)
	}
	defer f.Close()

	profile := rpprof.Lookup("mutex")
	if profile == nil {
		return fmt.Errorf("mutex profile not found")
	}

	if err := profile.WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}

	p.logger.Printf("Mutex profile captured: %s", filename)
	return nil
}

// CaptureAllProfiles captures all available profiles.
func (p *Profiler) CaptureAllProfiles(cpuDuration time.Duration) error {
	var errs []error

	// CPU profile
	if err := p.CaptureCPUProfile(cpuDuration); err != nil {
		errs = append(errs, fmt.Errorf("CPU: %w", err))
	}

	// Heap profile
	if err := p.CaptureHeapProfile(); err != nil {
		errs = append(errs, fmt.Errorf("heap: %w", err))
	}

	// Goroutine profile
	if err := p.CaptureGoroutineProfile(); err != nil {
		errs = append(errs, fmt.Errorf("goroutine: %w", err))
	}

	// Block profile (if enabled)
	if p.config.BlockProfileRate > 0 {
		if err := p.CaptureBlockProfile(); err != nil {
			errs = append(errs, fmt.Errorf("block: %w", err))
		}
	}

	// Mutex profile (if enabled)
	if p.config.MutexProfileFraction > 0 {
		if err := p.CaptureMutexProfile(); err != nil {
			errs = append(errs, fmt.Errorf("mutex: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("profile capture errors: %v", errs)
	}

	return nil
}

// WriteProfileTo writes a profile to the given writer.
func (p *Profiler) WriteProfileTo(profileName string, w io.Writer) error {
	profile := rpprof.Lookup(profileName)
	if profile == nil {
		return fmt.Errorf("profile %q not found", profileName)
	}

	return profile.WriteTo(w, 0)
}

// generateFilename generates a timestamped filename for profiles.
func (p *Profiler) generateFilename(profileType string) string {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.pprof", profileType, timestamp)
	return filepath.Join(p.config.OutputDir, filename)
}

// GetMemStats returns current memory statistics.
func (p *Profiler) GetMemStats() *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &m
}

// GetRuntimeStats returns runtime statistics.
func (p *Profiler) GetRuntimeStats() *RuntimeStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &RuntimeStats{
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		NumGC:        m.NumGC,
		GCCPUFrac:    m.GCCPUFraction,
	}
}

// RuntimeStats holds runtime statistics.
type RuntimeStats struct {
	NumGoroutine int
	NumCPU       int
	GOMAXPROCS   int
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	HeapObjects  uint64
	StackInuse   uint64
	StackSys     uint64
	NumGC        uint32
	GCCPUFrac    float64
}

// String returns a human-readable representation of RuntimeStats.
func (s *RuntimeStats) String() string {
	return fmt.Sprintf(`Runtime Stats:
  Goroutines: %d
  CPUs: %d (GOMAXPROCS: %d)
  Heap Alloc: %d MB
  Heap Sys: %d MB
  Heap Idle: %d MB
  Heap Inuse: %d MB
  Heap Released: %d MB
  Heap Objects: %d
  Stack Inuse: %d KB
  Stack Sys: %d KB
  GC Cycles: %d
  GC CPU Fraction: %.4f`,
		s.NumGoroutine,
		s.NumCPU, s.GOMAXPROCS,
		s.HeapAlloc/1024/1024,
		s.HeapSys/1024/1024,
		s.HeapIdle/1024/1024,
		s.HeapInuse/1024/1024,
		s.HeapReleased/1024/1024,
		s.HeapObjects,
		s.StackInuse/1024,
		s.StackSys/1024,
		s.NumGC,
		s.GCCPUFrac,
	)
}
