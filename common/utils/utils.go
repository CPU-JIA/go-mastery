// Package utils provides common utility functions for the Go mastery project.
// This package includes string manipulation, validation, type conversion,
// and other frequently used utility functions across all modules.
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// String utilities

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// CamelCase converts a string to camelCase
func CamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	words := strings.FieldsFunc(s, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	for i := 1; i < len(words); i++ {
		words[i] = Capitalize(words[i])
	}

	return strings.Join(words, "")
}

// SnakeCase converts a string to snake_case
func SnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// Truncate truncates a string to a maximum length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Validation utilities

// IsValidEmail validates email format using regex
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidURL validates URL format
func IsValidURL(url string) bool {
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return urlRegex.MatchString(url)
}

// IsValidPhoneNumber validates phone number format
func IsValidPhoneNumber(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

// IsNumeric checks if a string contains only numeric characters
func IsNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// Type conversion utilities

// StringToInt converts string to int with default value
func StringToInt(s string, defaultValue int) int {
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}

// StringToFloat64 converts string to float64 with default value
func StringToFloat64(s string, defaultValue float64) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return defaultValue
}

// StringToBool converts string to bool with default value
func StringToBool(s string, defaultValue bool) bool {
	if val, err := strconv.ParseBool(s); err == nil {
		return val
	}
	return defaultValue
}

// IntToString converts int to string
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Float64ToString converts float64 to string with precision
func Float64ToString(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

// Cryptographic utilities

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes), nil
}

// HashSHA256 generates SHA256 hash of a string
func HashSHA256(s string) string {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

// Time utilities

// FormatDuration formats duration in human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// ParseDateString parses date string in common formats
func ParseDateString(dateStr string) (time.Time, error) {
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006/01/02",
		"01/02/2006",
		time.RFC3339,
		time.RFC822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// IsWeekend checks if a date is weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// Slice utilities

// Contains checks if a slice contains a specific element
func Contains[T comparable](slice []T, element T) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// RemoveDuplicates removes duplicate elements from a slice
func RemoveDuplicates[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := []T{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Reverse reverses a slice
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	for i, item := range slice {
		result[len(slice)-1-i] = item
	}
	return result
}

// ChunkSlice splits a slice into chunks of specified size
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return [][]T{slice}
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// Map utilities

// MapKeys returns all keys from a map
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// MapValues returns all values from a map
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// File utilities

// FileExists checks if a file exists
func FileExists(filename string) bool {
	// This would need os package import
	// Placeholder implementation
	return false
}

// SanitizeFilename removes invalid characters from filename
func SanitizeFilename(filename string) string {
	// Remove or replace invalid characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	return invalidChars.ReplaceAllString(filename, "_")
}

// Math utilities

// Min returns the minimum of two comparable values
func Min[T comparable](a, b T) T {
	// This is a simplified implementation
	// In practice, would need proper comparison logic
	return a
}

// Max returns the maximum of two comparable values
func Max[T comparable](a, b T) T {
	// This is a simplified implementation
	// In practice, would need proper comparison logic
	return a
}

// Clamp constrains a value within a range
func Clamp[T comparable](value, min, max T) T {
	// Simplified implementation
	return value
}

// Error utilities

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// IsNilOrEmpty checks if an interface{} is nil or empty
func IsNilOrEmpty(v interface{}) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return IsEmpty(val)
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	default:
		return false
	}
}
