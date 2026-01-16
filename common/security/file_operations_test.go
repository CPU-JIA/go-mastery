package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func TestSecureWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	data := []byte("test data")

	if err := SecureWriteFile(testFile, data, nil); err != nil {
		t.Fatalf("SecureWriteFile failed: %v", err)
	}

	got, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(got) != string(data) {
		t.Errorf("file content mismatch: got %q want %q", string(got), string(data))
	}

	if err := ValidateFilePermissions(testFile, DefaultFileMode); err != nil {
		t.Errorf("permission validation failed: %v", err)
	}

	nestedFile := filepath.Join(tempDir, "nested", "dir", "test2.txt")
	opts := &SecureFileOptions{Mode: SecureFileMode_ReadOnlyUser, CreateDir: true}

	if err := SecureWriteFile(nestedFile, data, opts); err != nil {
		t.Fatalf("SecureWriteFile with CreateDir failed: %v", err)
	}

	if err := ValidateFilePermissions(nestedFile, SecureFileMode_ReadOnlyUser); err != nil {
		t.Errorf("custom permission validation failed: %v", err)
	}
}

func TestSecureCreateFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "created.txt")

	file, err := SecureCreateFile(testFile, DefaultFileMode)
	if err != nil {
		t.Fatalf("SecureCreateFile failed: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString("test content"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}

	if err := ValidateFilePermissions(testFile, DefaultFileMode); err != nil {
		t.Errorf("created file permission validation failed: %v", err)
	}
}

func TestSecureMkdirAll(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "secure", "nested", "dir")

	if err := SecureMkdirAll(testDir, DefaultDirMode); err != nil {
		t.Fatalf("SecureMkdirAll failed: %v", err)
	}

	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("created directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("created path is not a directory")
	}

	if runtime.GOOS != "windows" {
		expected := os.FileMode(DefaultDirMode)
		if info.Mode().Perm()&^expected != 0 {
			t.Errorf("directory permissions too loose: got %o expect max %o", info.Mode().Perm(), expected)
		}
	}
}

func TestGetRecommendedMode(t *testing.T) {
	cases := []struct {
		fileType string
		want     SecureFileMode
	}{
		{"config", DefaultConfigMode},
		{"configuration", DefaultConfigMode},
		{"log", DefaultLogMode},
		{"logs", DefaultLogMode},
		{"temp", DefaultTempMode},
		{"temporary", DefaultTempMode},
		{"executable", DefaultExecutableMode},
		{"binary", DefaultExecutableMode},
		{"data", DefaultFileMode},
		{"json", DefaultFileMode},
		{"unknown", DefaultFileMode},
	}

	for _, tc := range cases {
		t.Run(tc.fileType, func(t *testing.T) {
			if got := GetRecommendedMode(tc.fileType); got != tc.want {
				t.Errorf("GetRecommendedMode(%s) = %o want %o", tc.fileType, got, tc.want)
			}
		})
	}
}

func TestIsSecurePermission(t *testing.T) {
	cases := []struct {
		mode   os.FileMode
		secure bool
	}{
		{0o600, true},
		{0o644, false},
		{0o666, false},
		{0o777, false},
		{0o700, true},
		{0o755, false},
		{0o400, true},
		{0o444, false},
	}

	for _, tc := range cases {
		t.Run(tc.mode.String(), func(t *testing.T) {
			if got := IsSecurePermission(tc.mode); got != tc.secure {
				t.Errorf("IsSecurePermission(%o) = %v want %v", tc.mode, got, tc.secure)
			}
		})
	}
}

func TestValidateFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "perm_test.txt")

	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	file.Close()

	if err := ValidateFilePermissions(testFile, SecureFileMode_ReadWriteUser); err != nil {
		t.Errorf("validation should pass for correct permissions: %v", err)
	}

	if runtime.GOOS != "windows" {
		if err := ValidateFilePermissions(testFile, SecureFileMode_ReadOnlyUser); err == nil {
			t.Error("validation should fail when permissions exceed expected maximum")
		}
	}
}

func BenchmarkSecureWriteFile(b *testing.B) {
	tempDir := b.TempDir()
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "bench_"+strconv.Itoa(i)+".txt")
		_ = SecureWriteFile(testFile, data, nil)
	}
}

func BenchmarkRegularWriteFile(b *testing.B) {
	tempDir := b.TempDir()
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "bench_regular_"+strconv.Itoa(i)+".txt")
		_ = os.WriteFile(testFile, data, 0o644)
	}
}
