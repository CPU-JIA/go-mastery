package demo

import (
	"bytes"
	"strings"
	"testing"
)

func TestStructuredLoggingMode(t *testing.T) {
	// Test that structured logging can be enabled/disabled
	originalState := enableStructuredLogging
	defer func() { enableStructuredLogging = originalState }()

	EnableStructuredLogging(true)
	if !enableStructuredLogging {
		t.Error("EnableStructuredLogging(true) should enable structured logging")
	}

	EnableStructuredLogging(false)
	if enableStructuredLogging {
		t.Error("EnableStructuredLogging(false) should disable structured logging")
	}
}

func TestPrintWithStructuredLogging(t *testing.T) {
	// Capture output using custom output writer
	var buf bytes.Buffer
	originalOutput := customOutput
	originalStructured := enableStructuredLogging
	defer func() {
		customOutput = originalOutput
		enableStructuredLogging = originalStructured
	}()

	SetOutput(&buf)
	EnableStructuredLogging(true)
	Print("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}

	// Verify it's JSON structured format
	if !strings.Contains(output, "level") {
		t.Errorf("Expected structured log format, got: %s", output)
	}
}

func TestFormatCompatibilityFunctions(t *testing.T) {
	// Test that we can handle various fmt.Print patterns
	var buf bytes.Buffer
	originalOutput := customOutput
	defer func() { customOutput = originalOutput }()

	SetOutput(&buf)

	// Test PrintMigration works without panic
	PrintMigration("value1", " ", "value2")
	output := buf.String()
	if !strings.Contains(output, "value1") {
		t.Errorf("Expected output to contain 'value1', got: %s", output)
	}
}

func TestMigrationFunctions(t *testing.T) {
	// Test all migration functions
	var buf bytes.Buffer
	originalOutput := customOutput
	originalStructured := enableStructuredLogging
	defer func() {
		customOutput = originalOutput
		enableStructuredLogging = originalStructured
	}()

	SetOutput(&buf)
	EnableStructuredLogging(false) // Test non-structured mode

	PrintMigration("hello")
	PrintlnMigration("world")
	PrintfMigration("format: %s", "test")

	output := buf.String()
	if !strings.Contains(output, "hello") {
		t.Error("PrintMigration failed")
	}
	if !strings.Contains(output, "world") {
		t.Error("PrintlnMigration failed")
	}
	if !strings.Contains(output, "format: test") {
		t.Error("PrintfMigration failed")
	}
}
