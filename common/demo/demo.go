// Package demo provides logging utilities for educational Go examples
// It maintains the simplicity of fmt.Print while adding structured logging capabilities
package demo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// For demo purposes, we maintain the simple output style but can enable structured logging
var (
	// Demo output maintains fmt.Print behavior for educational clarity
	enableStructuredLogging = false
	// Custom output writer for testing and redirection
	customOutput io.Writer = os.Stdout
	// Log level for structured logging
	currentLogLevel = "INFO"
)

// Print replaces fmt.Print with consistent demo output
func Print(args ...interface{}) {
	if enableStructuredLogging {
		logStructured("INFO", fmt.Sprint(args...), nil)
	} else {
		fmt.Print(args...)
	}
}

// Println replaces fmt.Println with consistent demo output
func Println(args ...interface{}) {
	if enableStructuredLogging {
		logStructured("INFO", fmt.Sprint(args...), nil)
	} else {
		fmt.Println(args...)
	}
}

// Printf replaces fmt.Printf with consistent demo output
func Printf(format string, args ...interface{}) {
	if enableStructuredLogging {
		logStructured("INFO", fmt.Sprintf(format, args...), nil)
	} else {
		fmt.Printf(format, args...)
	}
}

// Section prints a formatted section header for educational clarity
func Section(title string) {
	fmt.Printf("\n=== %s ===\n", title)
}

// Example prints an example description
func Example(description string) {
	fmt.Printf("Example: %s\n", description)
}

// Result prints a result description
func Result(description string, value interface{}) {
	fmt.Printf("Result: %s -> %v\n", description, value)
}

// Error prints an error for demo purposes
func Error(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}

// Debug prints debug information for demos
func Debug(msg string) {
	fmt.Printf("Debug: %s\n", msg)
}

// EnableStructuredLogging enables structured output for production use
func EnableStructuredLogging(enabled bool) {
	enableStructuredLogging = enabled
}

// SetOutput sets custom output writer for testing and redirection
func SetOutput(w io.Writer) {
	customOutput = w
}

// SetLogLevel sets the log level for structured logging
func SetLogLevel(level string) {
	currentLogLevel = level
}

// Migration helper functions for fmt.Print compatibility

// PrintMigration provides fmt.Print-compatible function for migration
func PrintMigration(args ...interface{}) {
	if enableStructuredLogging {
		logStructured("INFO", fmt.Sprint(args...), nil)
	} else {
		fmt.Fprint(customOutput, args...)
	}
}

// PrintlnMigration provides fmt.Println-compatible function for migration
func PrintlnMigration(args ...interface{}) {
	if enableStructuredLogging {
		logStructured("INFO", fmt.Sprint(args...), nil)
	} else {
		fmt.Fprintln(customOutput, args...)
	}
}

// PrintfMigration provides fmt.Printf-compatible function for migration
func PrintfMigration(format string, args ...interface{}) {
	if enableStructuredLogging {
		logStructured("INFO", fmt.Sprintf(format, args...), nil)
	} else {
		fmt.Fprintf(customOutput, format, args...)
	}
}

// logStructured handles structured logging output
func logStructured(level, message string, fields map[string]interface{}) {
	logEntry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level,
		"message":   message,
		"context":   "demo",
	}

	if fields != nil {
		for k, v := range fields {
			logEntry[k] = v
		}
	}

	// For demo purposes, we use a simplified JSON output
	output, err := json.Marshal(logEntry)
	if err != nil {
		// Fallback to simple format
		fmt.Fprintf(customOutput, "[%s] [%s] %s\n", time.Now().Format("15:04:05"), level, message)
		return
	}

	fmt.Fprintln(customOutput, string(output))
}
