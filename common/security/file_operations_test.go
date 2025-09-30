package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecureWriteFile(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testData := []byte("test data")

	// 测试默认权限写入
	err := SecureWriteFile(testFile, testData, nil)
	if err != nil {
		t.Fatalf("SecureWriteFile failed: %v", err)
	}

	// 验证文件内容
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("File content mismatch: got %s, expected %s", string(data), string(testData))
	}

	// 验证文件权限
	err = ValidateFilePermissions(testFile, DefaultFileMode)
	if err != nil {
		t.Errorf("File permissions validation failed: %v", err)
	}

	// 测试自定义权限
	testFile2 := filepath.Join(tempDir, "test2.txt")
	opts := &SecureFileOptions{Mode: SecureFileMode_ReadOnlyUser}

	err = SecureWriteFile(testFile2, testData, opts)
	if err != nil {
		t.Fatalf("SecureWriteFile with custom mode failed: %v", err)
	}

	err = ValidateFilePermissions(testFile2, SecureFileMode_ReadOnlyUser)
	if err != nil {
		t.Errorf("Custom file permissions validation failed: %v", err)
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

	// 写入数据
	_, err = file.WriteString("test content")
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	// 验证权限
	err = ValidateFilePermissions(testFile, DefaultFileMode)
	if err != nil {
		t.Errorf("Created file permissions validation failed: %v", err)
	}
}

func TestSecureMkdirAll(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "secure", "nested", "dir")

	err := SecureMkdirAll(testDir, DefaultDirMode)
	if err != nil {
		t.Fatalf("SecureMkdirAll failed: %v", err)
	}

	// 验证目录存在
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Created directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}

	// 验证权限
	actualMode := info.Mode().Perm()
	expectedMode := os.FileMode(DefaultDirMode)

	if actualMode != expectedMode {
		t.Errorf("Directory permissions mismatch: got %o, expected %o", actualMode, expectedMode)
	}
}

func TestGetRecommendedMode(t *testing.T) {
	tests := []struct {
		fileType     string
		expectedMode SecureFileMode
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

	for _, tt := range tests {
		t.Run(tt.fileType, func(t *testing.T) {
			mode := GetRecommendedMode(tt.fileType)
			if mode != tt.expectedMode {
				t.Errorf("GetRecommendedMode(%s) = %o, expected %o", tt.fileType, mode, tt.expectedMode)
			}
		})
	}
}

func TestIsSecurePermission(t *testing.T) {
	tests := []struct {
		mode   os.FileMode
		secure bool
	}{
		{0600, true},  // 仅所有者读写
		{0644, false}, // 其他用户可读，不够安全
		{0666, false}, // 所有用户读写，不安全
		{0777, false}, // 所有用户读写执行，极不安全
		{0700, true},  // 仅所有者读写执行
		{0755, false}, // 其他用户可读执行，可能不够安全
		{0400, true},  // 仅所有者只读
		{0444, false}, // 所有用户只读，可能泄露信息
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			result := IsSecurePermission(tt.mode)
			if result != tt.secure {
				t.Errorf("IsSecurePermission(%o) = %t, expected %t", tt.mode, result, tt.secure)
			}
		})
	}
}

func TestValidateFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "perm_test.txt")

	// 创建文件with特定权限
	file, err := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// 验证正确的权限
	err = ValidateFilePermissions(testFile, SecureFileMode_ReadWriteUser)
	if err != nil {
		t.Errorf("Validation should pass for correct permissions: %v", err)
	}

	// 验证错误的权限
	err = ValidateFilePermissions(testFile, SecureFileMode_ReadOnlyUser)
	if err == nil {
		t.Error("Validation should fail for incorrect permissions")
	}
}

// 基准测试
func BenchmarkSecureWriteFile(b *testing.B) {
	tempDir := b.TempDir()
	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "bench_"+string(rune(i))+".txt")
		_ = SecureWriteFile(testFile, testData, nil)
	}
}

func BenchmarkRegularWriteFile(b *testing.B) {
	tempDir := b.TempDir()
	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, "bench_regular_"+string(rune(i))+".txt")
		_ = os.WriteFile(testFile, testData, 0644)
	}
}