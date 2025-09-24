package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestStructuredLogging tests structured logging with context and request ID
func TestStructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:      "DEBUG",
		Output:     "custom",
		Format:     "json",
		TimeFormat: time.RFC3339,
	}, &buf)

	ctx := context.WithValue(context.Background(), "request_id", "req-123")

	// Test with context
	logger.InfoContext(ctx, "test message", Fields{
		"user_id": 42,
		"action":  "login",
	})

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	// Validate structured log fields
	if logEntry["level"] != "INFO" {
		t.Errorf("Expected level INFO, got %v", logEntry["level"])
	}
	if logEntry["message"] != "test message" {
		t.Errorf("Expected message 'test message', got %v", logEntry["message"])
	}
	if logEntry["request_id"] != "req-123" {
		t.Errorf("Expected request_id 'req-123', got %v", logEntry["request_id"])
	}
	if logEntry["user_id"] != float64(42) {
		t.Errorf("Expected user_id 42, got %v", logEntry["user_id"])
	}
}

// TestLogLevels tests all log levels work correctly
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:      "DEBUG",
		Output:     "custom",
		Format:     "json",
		TimeFormat: time.RFC3339,
	}, &buf)

	logger.Debug("debug message", nil)
	logger.Info("info message", nil)
	logger.Warn("warn message", nil)
	logger.Error("error message", nil)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d", len(lines))
	}

	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	for i, line := range lines {
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Fatalf("Failed to parse JSON log line %d: %v", i, err)
		}

		if logEntry["level"] != expectedLevels[i] {
			t.Errorf("Line %d: Expected level %s, got %v", i, expectedLevels[i], logEntry["level"])
		}
	}
}

// TestRequestTracking tests request ID tracking functionality
func TestRequestTracking(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:  "INFO",
		Output: "custom",
		Format: "json",
	}, &buf)

	// Test with request ID in context
	ctx := WithRequestID(context.Background(), "req-456")
	logger.InfoContext(ctx, "processing request", Fields{
		"operation": "user_lookup",
	})

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["request_id"] != "req-456" {
		t.Errorf("Expected request_id 'req-456', got %v", logEntry["request_id"])
	}
}

// TestLogFiltering tests log level filtering
func TestLogFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:  "WARN", // Should only show WARN and ERROR
		Output: "custom",
		Format: "json",
	}, &buf)

	logger.Debug("debug message", nil)
	logger.Info("info message", nil)
	logger.Warn("warn message", nil)
	logger.Error("error message", nil)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should only have 2 lines (WARN and ERROR)
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines with WARN level, got %d", len(lines))
	}
}

// Benchmark tests for performance
func BenchmarkStructuredLogging(b *testing.B) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:  "INFO",
		Output: "custom",
		Format: "json",
	}, &buf)

	ctx := context.Background()
	fields := Fields{
		"user_id": 123,
		"action":  "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		logger.InfoContext(ctx, "benchmark message", fields)
	}
}

// TestMigrationFunctions tests fmt.Print migration support
func TestMigrationFunctions(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:  "INFO",
		Output: "custom",
		Format: "text",
	}, &buf)

	// Test PrintMigration functions
	logger.PrintMigration("hello", " ", "world")
	logger.PrintlnMigration("new line test")
	logger.PrintfMigration("formatted: %d %s", 42, "test")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Error("PrintMigration failed")
	}
	if !strings.Contains(output, "new line test") {
		t.Error("PrintlnMigration failed")
	}
	if !strings.Contains(output, "formatted: 42 test") {
		t.Error("PrintfMigration failed")
	}
}

// TestMigrationWithContext tests migration functions with context
func TestMigrationWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := NewStructured(&Config{
		Level:  "INFO",
		Output: "custom",
		Format: "json",
	}, &buf)

	ctx := WithRequestID(context.Background(), "req-789")
	logger.PrintfMigrationContext(ctx, "processing user %d", 123)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse JSON log: %v", err)
	}

	if logEntry["message"] != "processing user 123" {
		t.Errorf("Expected message 'processing user 123', got %v", logEntry["message"])
	}
	if logEntry["request_id"] != "req-789" {
		t.Errorf("Expected request_id 'req-789', got %v", logEntry["request_id"])
	}
}
