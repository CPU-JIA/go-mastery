// Package testutil provides common testing utilities for the Go mastery project.
// This package includes test helpers, mock generators, assertion utilities,
// and performance testing tools to maintain high code quality standards.
package testutil

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestConfig provides configuration for test environments
type TestConfig struct {
	DatabaseURL string
	RedisURL    string
	LogLevel    string
	Timeout     time.Duration
}

// DefaultTestConfig returns a test configuration with sensible defaults
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		DatabaseURL: "sqlite::memory:",
		RedisURL:    "redis://localhost:6379/15", // Use test database
		LogLevel:    "debug",
		Timeout:     30 * time.Second,
	}
}

// HTTPTestHelper provides utilities for HTTP testing
type HTTPTestHelper struct {
	Server *httptest.Server
	Client *http.Client
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(handler http.Handler) *HTTPTestHelper {
	server := httptest.NewServer(handler)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &HTTPTestHelper{
		Server: server,
		Client: client,
	}
}

// Close cleans up the test helper resources
func (h *HTTPTestHelper) Close() {
	h.Server.Close()
}

// GET performs a GET request and returns the response
func (h *HTTPTestHelper) GET(path string) (*http.Response, error) {
	return h.Client.Get(h.Server.URL + path)
}

// POST performs a POST request with JSON body
func (h *HTTPTestHelper) POST(path string, body interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return h.Client.Post(
		h.Server.URL+path,
		"application/json",
		strings.NewReader(string(jsonBody)),
	)
}

// AssertJSON compares JSON responses for testing
func AssertJSON(t *testing.T, expected, actual interface{}) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Failed to marshal expected JSON: %v", err)
	}

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("Failed to marshal actual JSON: %v", err)
	}

	if !jsonEqual(expectedJSON, actualJSON) {
		t.Errorf("JSON mismatch:\nExpected: %s\nActual: %s", expectedJSON, actualJSON)
	}
}

// AssertHTTPStatus checks HTTP response status
func AssertHTTPStatus(t *testing.T, expected int, resp *http.Response) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status %d, got %d. Body: %s", expected, resp.StatusCode, body)
	}
}

// AssertNoError fails the test if error is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError fails the test if error is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

// AssertEqual performs deep equality comparison
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Values not equal:\nExpected: %+v\nActual: %+v", expected, actual)
	}
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("String %q does not contain %q", haystack, needle)
	}
}

// MockTimer provides controllable time for testing
type MockTimer struct {
	current time.Time
}

// NewMockTimer creates a new mock timer
func NewMockTimer(start time.Time) *MockTimer {
	return &MockTimer{current: start}
}

// Now returns the current mock time
func (m *MockTimer) Now() time.Time {
	return m.current
}

// Advance moves the mock time forward
func (m *MockTimer) Advance(duration time.Duration) {
	m.current = m.current.Add(duration)
}

// TestTimeout provides timeout context for tests
func TestTimeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

// BenchmarkHelper provides utilities for benchmark tests
type BenchmarkHelper struct {
	b *testing.B
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(b *testing.B) *BenchmarkHelper {
	return &BenchmarkHelper{b: b}
}

// MeasureMemory measures memory allocation during benchmark
func (h *BenchmarkHelper) MeasureMemory(fn func()) {
	h.b.ReportAllocs()
	h.b.ResetTimer()
	for i := 0; i < h.b.N; i++ {
		fn()
	}
}

// MeasureTime measures execution time with custom reporting
func (h *BenchmarkHelper) MeasureTime(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// jsonEqual compares two JSON byte slices for equality
func jsonEqual(a, b []byte) bool {
	var aInterface, bInterface interface{}

	if err := json.Unmarshal(a, &aInterface); err != nil {
		return false
	}

	if err := json.Unmarshal(b, &bInterface); err != nil {
		return false
	}

	return reflect.DeepEqual(aInterface, bInterface)
}

// RandomString generates a random string for testing
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)%len(b)] // Simplified for example
	}
	return string(b)
}

// SetupTestDB initializes a test database
func SetupTestDB(t *testing.T) string {
	t.Helper()
	// Return in-memory SQLite URL for testing
	return "sqlite::memory:"
}

// CleanupTestDB cleans up test database
func CleanupTestDB(t *testing.T, dbURL string) {
	t.Helper()
	// Cleanup logic would go here
}

// LogCapture captures log output for testing
type LogCapture struct {
	logs []string
}

// NewLogCapture creates a new log capture
func NewLogCapture() *LogCapture {
	return &LogCapture{
		logs: make([]string, 0),
	}
}

// Write implements io.Writer for capturing logs
func (lc *LogCapture) Write(p []byte) (n int, err error) {
	lc.logs = append(lc.logs, string(p))
	return len(p), nil
}

// GetLogs returns captured logs
func (lc *LogCapture) GetLogs() []string {
	return lc.logs
}

// ContainsLog checks if a log message was captured
func (lc *LogCapture) ContainsLog(message string) bool {
	for _, log := range lc.logs {
		if strings.Contains(log, message) {
			return true
		}
	}
	return false
}
