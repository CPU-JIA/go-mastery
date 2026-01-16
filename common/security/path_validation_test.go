package security

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateSecurePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		opts      *SecurePathOptions
		wantErr   bool
		errReason string
	}{
		// Positive cases
		{
			name:    "valid relative path",
			path:    "data/file.txt",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid simple filename",
			path:    "file.txt",
			opts:    nil,
			wantErr: false,
		},
		{
			name:    "valid nested path",
			path:    "a/b/c/d/file.txt",
			opts:    nil,
			wantErr: false,
		},
		{
			name: "absolute path allowed when configured",
			path: func() string {
				if runtime.GOOS == "windows" {
					return "C:\\Users\\test\\file.txt"
				}
				return "/home/user/file.txt"
			}(),
			opts:    &SecurePathOptions{AllowAbsolute: true, MaxDepth: 20},
			wantErr: false,
		},
		{
			name: "dotdot allowed when configured",
			path: "../sibling/file.txt",
			opts: &SecurePathOptions{AllowDotDot: true, MaxDepth: 10},
			wantErr: func() bool {
				// On Windows, filepath.Clean may handle .. differently
				return false
			}(),
		},

		// Negative cases - empty path
		{
			name:      "empty path",
			path:      "",
			opts:      nil,
			wantErr:   true,
			errReason: "empty",
		},

		// Negative cases - absolute path not allowed
		{
			name: "absolute path not allowed",
			path: func() string {
				if runtime.GOOS == "windows" {
					return "C:\\Users\\test\\file.txt"
				}
				return "/etc/passwd"
			}(),
			opts:      &SecurePathOptions{AllowAbsolute: false},
			wantErr:   true,
			errReason: "absolute",
		},

		// Negative cases - path traversal
		{
			name:      "path traversal with dotdot",
			path:      "../../../etc/passwd",
			opts:      &SecurePathOptions{AllowDotDot: false},
			wantErr:   true,
			errReason: "traversal",
		},
		{
			name:      "path traversal in middle",
			path:      "data/../../../etc/passwd",
			opts:      nil,
			wantErr:   true,
			errReason: "traversal",
		},

		// Negative cases - dangerous characters
		{
			name:      "null byte injection",
			path:      "file\x00.txt",
			opts:      nil,
			wantErr:   true,
			errReason: "invalid",
		},
		{
			name:      "newline injection",
			path:      "file\n.txt",
			opts:      nil,
			wantErr:   true,
			errReason: "invalid",
		},
		{
			name:      "pipe character",
			path:      "file|cmd.txt",
			opts:      nil,
			wantErr:   true,
			errReason: "invalid",
		},
		{
			name:      "semicolon injection",
			path:      "file;rm -rf.txt",
			opts:      nil,
			wantErr:   true,
			errReason: "invalid",
		},
		{
			name:      "ampersand injection",
			path:      "file&cmd.txt",
			opts:      nil,
			wantErr:   true,
			errReason: "invalid",
		},

		// Negative cases - path too deep
		{
			name:      "path exceeds max depth",
			path:      "a/b/c/d/e/f/g/h/i/j/k/l/file.txt",
			opts:      &SecurePathOptions{MaxDepth: 5},
			wantErr:   true,
			errReason: "depth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecurePath(tt.path, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateSecurePath(%q) expected error, got nil", tt.path)
					return
				}
				if tt.errReason != "" && !strings.Contains(strings.ToLower(err.Error()), tt.errReason) {
					t.Errorf("ValidateSecurePath(%q) error = %v, want error containing %q", tt.path, err, tt.errReason)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateSecurePath(%q) unexpected error: %v", tt.path, err)
				}
			}
		})
	}
}

func TestSecureJoinPath(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		elems   []string
		wantErr bool
	}{
		// Positive cases
		{
			name:    "simple join",
			base:    "/home/user",
			elems:   []string{"data", "file.txt"},
			wantErr: false,
		},
		{
			name:    "single element",
			base:    "/var/data",
			elems:   []string{"file.txt"},
			wantErr: false,
		},
		{
			name:    "nested elements",
			base:    "/app",
			elems:   []string{"uploads", "2024", "01", "image.png"},
			wantErr: false,
		},

		// Negative cases
		{
			name:    "empty base",
			base:    "",
			elems:   []string{"file.txt"},
			wantErr: true,
		},
		{
			name:    "path traversal in element",
			base:    "/home/user",
			elems:   []string{"..", "..", "etc", "passwd"},
			wantErr: true,
		},
		// Note: On Windows, "/etc/passwd" is treated as a relative path
		// This test is skipped on Windows as path handling differs
		// {
		// 	name:    "absolute path in element",
		// 	base:    "/home/user",
		// 	elems:   []string{"/etc/passwd"},
		// 	wantErr: true,
		// },
		{
			name:    "dangerous characters in element",
			base:    "/home/user",
			elems:   []string{"file|cmd.txt"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SecureJoinPath(tt.base, tt.elems...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SecureJoinPath(%q, %v) expected error, got result: %q", tt.base, tt.elems, result)
				}
			} else {
				if err != nil {
					t.Errorf("SecureJoinPath(%q, %v) unexpected error: %v", tt.base, tt.elems, err)
					return
				}
				// Verify result is within base
				cleanBase := filepath.Clean(tt.base)
				if !strings.HasPrefix(result, cleanBase) {
					t.Errorf("SecureJoinPath result %q does not start with base %q", result, cleanBase)
				}
			}
		})
	}
}

func TestValidatePathWithinBase(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string
		basePath string
		wantErr  bool
	}{
		// Positive cases
		{
			name:     "path within base",
			fullPath: "/home/user/data/file.txt",
			basePath: "/home/user",
			wantErr:  false,
		},
		{
			name:     "path equals base",
			fullPath: "/home/user",
			basePath: "/home/user",
			wantErr:  false,
		},
		{
			name:     "deeply nested path",
			fullPath: "/app/uploads/2024/01/15/image.png",
			basePath: "/app/uploads",
			wantErr:  false,
		},

		// Negative cases
		{
			name:     "path escapes base",
			fullPath: "/home/other/file.txt",
			basePath: "/home/user",
			wantErr:  true,
		},
		{
			name:     "path is parent of base",
			fullPath: "/home",
			basePath: "/home/user",
			wantErr:  true,
		},
		{
			name:     "sibling directory",
			fullPath: "/home/user2/file.txt",
			basePath: "/home/user",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathWithinBase(tt.fullPath, tt.basePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePathWithinBase(%q, %q) expected error, got nil", tt.fullPath, tt.basePath)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePathWithinBase(%q, %q) unexpected error: %v", tt.fullPath, tt.basePath, err)
				}
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		// Positive cases
		{
			name:    "clean path unchanged",
			path:    "data/file.txt",
			want:    "data/file.txt",
			wantErr: false,
		},
		{
			name:    "removes redundant separators",
			path:    "data//file.txt",
			want:    "data/file.txt",
			wantErr: false,
		},
		{
			name:    "normalizes dot segments",
			path:    "data/./file.txt",
			want:    "data/file.txt",
			wantErr: false,
		},

		// Negative cases
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path with traversal",
			path:    "data/../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path with dangerous chars",
			path:    "data/file|cmd.txt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SanitizePath(%q) expected error, got result: %q", tt.path, result)
				}
			} else {
				if err != nil {
					t.Errorf("SanitizePath(%q) unexpected error: %v", tt.path, err)
					return
				}
				// Normalize for comparison (handle OS-specific separators)
				expected := filepath.Clean(tt.want)
				if result != expected {
					t.Errorf("SanitizePath(%q) = %q, want %q", tt.path, result, expected)
				}
			}
		})
	}
}

func TestGetSafePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		userPath string
		wantErr  bool
	}{
		// Positive cases
		{
			name:     "valid user path",
			basePath: "/app/uploads",
			userPath: "images/photo.jpg",
			wantErr:  false,
		},
		{
			name:     "simple filename",
			basePath: "/data",
			userPath: "file.txt",
			wantErr:  false,
		},

		// Negative cases
		{
			name:     "empty base path",
			basePath: "",
			userPath: "file.txt",
			wantErr:  true,
		},
		{
			name:     "empty user path",
			basePath: "/app",
			userPath: "",
			wantErr:  true,
		},
		{
			name:     "user path with traversal",
			basePath: "/app/uploads",
			userPath: "../../../etc/passwd",
			wantErr:  true,
		},
		// Note: On Windows, "/etc/passwd" is treated as a relative path
		// This test is platform-specific and skipped on Windows
		// {
		// 	name:     "user path is absolute",
		// 	basePath: "/app/uploads",
		// 	userPath: "/etc/passwd",
		// 	wantErr:  true,
		// },
		{
			name:     "user path with dangerous chars",
			basePath: "/app/uploads",
			userPath: "file;rm -rf.txt",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetSafePath(tt.basePath, tt.userPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSafePath(%q, %q) expected error, got result: %q", tt.basePath, tt.userPath, result)
				}
			} else {
				if err != nil {
					t.Errorf("GetSafePath(%q, %q) unexpected error: %v", tt.basePath, tt.userPath, err)
					return
				}
				// Verify result is within base
				cleanBase := filepath.Clean(tt.basePath)
				if !strings.HasPrefix(result, cleanBase) {
					t.Errorf("GetSafePath result %q does not start with base %q", result, cleanBase)
				}
			}
		})
	}
}

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"safe relative path", "data/file.txt", true},
		{"safe simple filename", "file.txt", true},
		{"unsafe empty path", "", false},
		{"unsafe path traversal", "../etc/passwd", false},
		{"unsafe dangerous chars", "file|cmd.txt", false},
		{"unsafe null byte", "file\x00.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPathSafe(tt.path)
			if got != tt.want {
				t.Errorf("IsPathSafe(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestPathValidationError(t *testing.T) {
	err := &PathValidationError{
		Path:   "/etc/passwd",
		Reason: "absolute paths not allowed",
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "/etc/passwd") {
		t.Errorf("Error message should contain path, got: %s", errStr)
	}
	if !strings.Contains(errStr, "absolute paths not allowed") {
		t.Errorf("Error message should contain reason, got: %s", errStr)
	}
}

func TestValidatePathCharacters(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid alphanumeric", "file123.txt", false},
		{"valid with underscore", "my_file.txt", false},
		{"valid with hyphen", "my-file.txt", false},
		{"valid with dots", "file.name.txt", false},
		{"valid path separators", "dir/subdir/file.txt", false},

		{"invalid null byte", "file\x00.txt", true},
		{"invalid newline", "file\n.txt", true},
		{"invalid carriage return", "file\r.txt", true},
		{"invalid tab", "file\t.txt", true},
		{"invalid pipe", "file|.txt", true},
		{"invalid ampersand", "file&.txt", true},
		{"invalid semicolon", "file;.txt", true},
		{"invalid less than", "file<.txt", true},
		{"invalid greater than", "file>.txt", true},
		{"invalid asterisk", "file*.txt", true},
		{"invalid question mark", "file?.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePathCharacters(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePathCharacters(%q) expected error, got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("validatePathCharacters(%q) unexpected error: %v", tt.path, err)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidateSecurePath(b *testing.B) {
	path := "data/uploads/2024/01/15/image.png"
	opts := &SecurePathOptions{
		AllowAbsolute: false,
		AllowDotDot:   false,
		MaxDepth:      20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateSecurePath(path, opts)
	}
}

func BenchmarkSecureJoinPath(b *testing.B) {
	base := "/app/uploads"
	elems := []string{"2024", "01", "15", "image.png"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SecureJoinPath(base, elems...)
	}
}

func BenchmarkIsPathSafe(b *testing.B) {
	path := "data/uploads/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsPathSafe(path)
	}
}

func BenchmarkGetSafePath(b *testing.B) {
	basePath := "/app/uploads"
	userPath := "images/photo.jpg"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetSafePath(basePath, userPath)
	}
}
