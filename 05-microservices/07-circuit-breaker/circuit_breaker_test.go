package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCircuitBreakerState_String(t *testing.T) {
	tests := []struct {
		state CircuitBreakerState
		want  string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF-OPEN"},
		{CircuitBreakerState(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("CircuitBreakerState.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5", config.FailureThreshold)
	}

	if config.SuccessThreshold != 3 {
		t.Errorf("SuccessThreshold = %d, want 3", config.SuccessThreshold)
	}

	if config.Timeout != 2*time.Second {
		t.Errorf("Timeout = %v, want 2s", config.Timeout)
	}

	if config.RecoveryTimeout != 30*time.Second {
		t.Errorf("RecoveryTimeout = %v, want 30s", config.RecoveryTimeout)
	}

	if config.SlidingWindow != time.Minute {
		t.Errorf("SlidingWindow = %v, want 1m", config.SlidingWindow)
	}

	if config.MaxRequests != 10 {
		t.Errorf("MaxRequests = %d, want 10", config.MaxRequests)
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	if cb == nil {
		t.Fatal("NewCircuitBreaker returned nil")
	}

	if cb.state != StateClosed {
		t.Errorf("Initial state = %v, want CLOSED", cb.state)
	}

	if cb.failureCount != 0 {
		t.Errorf("Initial failureCount = %d, want 0", cb.failureCount)
	}

	if cb.successCount != 0 {
		t.Errorf("Initial successCount = %d, want 0", cb.successCount)
	}
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	stats := cb.GetStats()
	if stats.TotalRequests != 1 {
		t.Errorf("TotalRequests = %d, want 1", stats.TotalRequests)
	}

	if stats.SuccessRequests != 1 {
		t.Errorf("SuccessRequests = %d, want 1", stats.SuccessRequests)
	}

	if stats.FailureRequests != 0 {
		t.Errorf("FailureRequests = %d, want 0", stats.FailureRequests)
	}
}

func TestCircuitBreaker_Execute_Failure(t *testing.T) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	expectedErr := errors.New("test error")
	err := cb.Execute(func() error {
		return expectedErr
	})

	if err != expectedErr {
		t.Errorf("Execute() error = %v, want %v", err, expectedErr)
	}

	stats := cb.GetStats()
	if stats.FailureRequests != 1 {
		t.Errorf("FailureRequests = %d, want 1", stats.FailureRequests)
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
		RecoveryTimeout:  1 * time.Second,
		MaxRequests:      5,
	}
	cb := NewCircuitBreaker(config)

	// Cause failures to trigger open state
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Circuit should be open now
	if cb.GetState() != StateOpen {
		t.Errorf("State after 3 failures = %v, want OPEN", cb.GetState())
	}

	// Next request should be rejected
	err := cb.Execute(func() error {
		return nil
	})

	if err == nil || err.Error() != "circuit breaker is open" {
		t.Errorf("Execute() when open should return 'circuit breaker is open', got %v", err)
	}

	stats := cb.GetStats()
	if stats.RejectedRequests != 1 {
		t.Errorf("RejectedRequests = %d, want 1", stats.RejectedRequests)
	}
}

func TestCircuitBreaker_HalfOpenAfterRecoveryTimeout(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
		RecoveryTimeout:  200 * time.Millisecond,
		MaxRequests:      5,
	}
	cb := NewCircuitBreaker(config)

	// Trigger open state
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("State should be OPEN, got %v", cb.GetState())
	}

	// Wait for recovery timeout
	time.Sleep(250 * time.Millisecond)

	// Next request should transition to half-open
	_ = cb.Execute(func() error {
		return nil
	})

	state := cb.GetState()
	// State could be HALF-OPEN or CLOSED depending on success
	if state != StateHalfOpen && state != StateClosed {
		t.Errorf("State after recovery timeout = %v, want HALF-OPEN or CLOSED", state)
	}
}

func TestCircuitBreaker_RecoverFromHalfOpen(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
		RecoveryTimeout:  100 * time.Millisecond,
		MaxRequests:      10,
	}
	cb := NewCircuitBreaker(config)

	// Trigger open state
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Successful requests should recover the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return nil
		})
	}

	if cb.GetState() != StateClosed {
		t.Errorf("State after recovery = %v, want CLOSED", cb.GetState())
	}
}

func TestCircuitBreaker_ExecuteWithContext_Timeout(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 3,
		Timeout:          50 * time.Millisecond,
		RecoveryTimeout:  1 * time.Second,
		MaxRequests:      10,
	}
	cb := NewCircuitBreaker(config)

	err := cb.ExecuteWithContext(context.Background(), func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	if err == nil {
		t.Error("ExecuteWithContext() should timeout")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("ExecuteWithContext() error = %v, want context.DeadlineExceeded", err)
	}
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          100 * time.Millisecond,
		RecoveryTimeout:  100 * time.Millisecond,
		MaxRequests:      10,
	}
	cb := NewCircuitBreaker(config)

	var stateChanges []struct {
		from, to CircuitBreakerState
	}
	var mu sync.Mutex

	cb.OnStateChange(func(from, to CircuitBreakerState) {
		mu.Lock()
		stateChanges = append(stateChanges, struct{ from, to CircuitBreakerState }{from, to})
		mu.Unlock()
	})

	// Trigger failures to open circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errors.New("failure")
		})
	}

	// Wait for callback
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if len(stateChanges) == 0 {
		t.Error("OnStateChange callback was not called")
	} else {
		lastChange := stateChanges[len(stateChanges)-1]
		if lastChange.to != StateOpen {
			t.Errorf("Last state change to = %v, want OPEN", lastChange.to)
		}
	}
	mu.Unlock()
}

func TestCircuitBreaker_GetStats(t *testing.T) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	// Execute some requests
	for i := 0; i < 5; i++ {
		_ = cb.Execute(func() error {
			return nil
		})
	}

	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return errors.New("error")
		})
	}

	stats := cb.GetStats()

	if stats.TotalRequests != 8 {
		t.Errorf("TotalRequests = %d, want 8", stats.TotalRequests)
	}

	if stats.SuccessRequests != 5 {
		t.Errorf("SuccessRequests = %d, want 5", stats.SuccessRequests)
	}

	if stats.FailureRequests != 3 {
		t.Errorf("FailureRequests = %d, want 3", stats.FailureRequests)
	}

	if stats.State != "CLOSED" {
		t.Errorf("State = %q, want CLOSED", stats.State)
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold: 100,
		SuccessThreshold: 10,
		Timeout:          1 * time.Second,
		RecoveryTimeout:  1 * time.Second,
		MaxRequests:      100,
	}
	cb := NewCircuitBreaker(config)

	var wg sync.WaitGroup
	numGoroutines := 100
	requestsPerGoroutine := 10

	var successCount int64
	var failureCount int64

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				err := cb.Execute(func() error {
					if id%2 == 0 {
						return nil
					}
					return errors.New("error")
				})
				if err == nil {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&failureCount, 1)
				}
			}
		}(i)
	}

	wg.Wait()

	stats := cb.GetStats()
	expectedTotal := int64(numGoroutines * requestsPerGoroutine)

	if stats.TotalRequests != expectedTotal {
		t.Errorf("TotalRequests = %d, want %d", stats.TotalRequests, expectedTotal)
	}

	// Verify counts are consistent
	if stats.SuccessRequests+stats.FailureRequests+stats.RejectedRequests != expectedTotal {
		t.Errorf("Sum of requests (%d+%d+%d) != TotalRequests (%d)",
			stats.SuccessRequests, stats.FailureRequests, stats.RejectedRequests, expectedTotal)
	}
}

func TestUnstableService(t *testing.T) {
	// Test with 0% failure rate
	service := NewUnstableService(0.0, 0)
	for i := 0; i < 10; i++ {
		err := service.Call()
		if err != nil {
			t.Errorf("Service with 0%% failure rate returned error: %v", err)
		}
	}

	// Test with 100% failure rate
	service = NewUnstableService(1.0, 0)
	for i := 0; i < 10; i++ {
		err := service.Call()
		if err == nil {
			t.Error("Service with 100% failure rate should always fail")
		}
	}
}

func TestUnstableService_Delay(t *testing.T) {
	delay := 50 * time.Millisecond
	service := NewUnstableService(0.0, delay)

	start := time.Now()
	_ = service.Call()
	elapsed := time.Since(start)

	if elapsed < delay {
		t.Errorf("Service call took %v, expected at least %v", elapsed, delay)
	}
}

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	if collector == nil {
		t.Fatal("NewMetricsCollector returned nil")
	}

	// Register circuit breakers
	cb1 := NewCircuitBreaker(DefaultConfig())
	cb2 := NewCircuitBreaker(DefaultConfig())

	collector.Register("service-1", cb1)
	collector.Register("service-2", cb2)

	// Execute some requests
	_ = cb1.Execute(func() error { return nil })
	_ = cb2.Execute(func() error { return errors.New("error") })

	// Get all stats
	allStats := collector.GetAllStats()

	if len(allStats) != 2 {
		t.Errorf("GetAllStats() returned %d entries, want 2", len(allStats))
	}

	if _, exists := allStats["service-1"]; !exists {
		t.Error("service-1 not found in stats")
	}

	if _, exists := allStats["service-2"]; !exists {
		t.Error("service-2 not found in stats")
	}

	if allStats["service-1"].SuccessRequests != 1 {
		t.Errorf("service-1 SuccessRequests = %d, want 1", allStats["service-1"].SuccessRequests)
	}

	if allStats["service-2"].FailureRequests != 1 {
		t.Errorf("service-2 FailureRequests = %d, want 1", allStats["service-2"].FailureRequests)
	}
}

func TestSecureRandomFloat64(t *testing.T) {
	// Test that function returns values in [0, 1) range
	for i := 0; i < 100; i++ {
		val := secureRandomFloat64()
		if val < 0 || val >= 1 {
			t.Errorf("secureRandomFloat64() = %f, want value in [0, 1)", val)
		}
	}

	// Test that function produces different values
	values := make(map[float64]bool)
	for i := 0; i < 100; i++ {
		values[secureRandomFloat64()] = true
	}

	// Should have at least some variety
	if len(values) < 50 {
		t.Errorf("secureRandomFloat64() produced only %d unique values out of 100", len(values))
	}
}

// Benchmark tests
func BenchmarkCircuitBreaker_Execute_Success(b *testing.B) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(func() error {
			return nil
		})
	}
}

func BenchmarkCircuitBreaker_Execute_Failure(b *testing.B) {
	config := CircuitBreakerConfig{
		FailureThreshold: 1000000, // High threshold to prevent opening
		SuccessThreshold: 3,
		Timeout:          1 * time.Second,
		RecoveryTimeout:  1 * time.Second,
		MaxRequests:      10,
	}
	cb := NewCircuitBreaker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Execute(func() error {
			return errors.New("error")
		})
	}
}

func BenchmarkCircuitBreaker_GetStats(b *testing.B) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	// Pre-populate some stats
	for i := 0; i < 100; i++ {
		_ = cb.Execute(func() error { return nil })
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.GetStats()
	}
}

func BenchmarkCircuitBreaker_Concurrent(b *testing.B) {
	config := DefaultConfig()
	cb := NewCircuitBreaker(config)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cb.Execute(func() error {
				return nil
			})
		}
	})
}
