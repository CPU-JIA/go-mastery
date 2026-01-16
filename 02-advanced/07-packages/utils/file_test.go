// Package utils 文件工具函数测试
package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// FileExists 函数测试
// =============================================================================

// TestFileExists 测试文件存在检查
func TestFileExists(t *testing.T) {
	// 创建临时文件用于测试
	tmpFile, err := os.CreateTemp("", "test_file_*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"存在的文件", tmpFileName, true},
		{"不存在的文件", "/nonexistent/path/file.txt", false},
		{"空路径", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.filename)
			if result != tt.expected {
				t.Errorf("FileExists(%q) = %v; 期望 %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetFileExtension 函数测试
// =============================================================================

// TestGetFileExtension 测试获取文件扩展名
func TestGetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// 正向用例
		{"普通扩展名", "file.txt", ".txt"},
		{"多个点", "file.tar.gz", ".gz"},
		{"大写扩展名", "file.TXT", ".TXT"},
		// 边界条件
		{"无扩展名", "file", ""},
		{"隐藏文件", ".gitignore", ".gitignore"},
		{"路径中的文件", "/path/to/file.go", ".go"},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("GetFileExtension(%q) = %q; 期望 %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GetFilenameWithoutExtension 函数测试
// =============================================================================

// TestGetFilenameWithoutExtension 测试获取不带扩展名的文件名
func TestGetFilenameWithoutExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		// 正向用例
		{"普通文件", "file.txt", "file"},
		{"多个点", "file.tar.gz", "file.tar"},
		{"路径中的文件", "/path/to/file.go", "file"},
		// 边界条件
		{"无扩展名", "file", "file"},
		{"隐藏文件", ".gitignore", ""},
		{"空字符串", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilenameWithoutExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("GetFilenameWithoutExtension(%q) = %q; 期望 %q", tt.filename, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// ReadFileLines 和 WriteFileLines 函数测试
// =============================================================================

// TestReadWriteFileLines 测试文件行读写
func TestReadWriteFileLines(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_utils_")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "test.txt")

	// 测试写入
	lines := []string{"第一行", "第二行", "第三行"}
	err = WriteFileLines(tmpFile, lines)
	if err != nil {
		t.Fatalf("WriteFileLines 失败: %v", err)
	}

	// 测试读取
	readLines, err := ReadFileLines(tmpFile)
	if err != nil {
		t.Fatalf("ReadFileLines 失败: %v", err)
	}

	if len(readLines) != len(lines) {
		t.Errorf("读取行数 = %d; 期望 %d", len(readLines), len(lines))
	}

	for i, line := range readLines {
		if line != lines[i] {
			t.Errorf("第 %d 行 = %q; 期望 %q", i, line, lines[i])
		}
	}
}

// TestReadFileLines_NotExist 测试读取不存在的文件
func TestReadFileLines_NotExist(t *testing.T) {
	_, err := ReadFileLines("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadFileLines 应返回错误当文件不存在时")
	}
}

// =============================================================================
// CopyFile 函数测试
// =============================================================================

// TestCopyFile 测试文件复制
func TestCopyFile(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_copy_")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建源文件
	srcFile := filepath.Join(tmpDir, "source.txt")
	content := []byte("测试内容")
	err = os.WriteFile(srcFile, content, 0600)
	if err != nil {
		t.Fatalf("创建源文件失败: %v", err)
	}

	// 复制文件
	dstFile := filepath.Join(tmpDir, "dest.txt")
	err = CopyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("CopyFile 失败: %v", err)
	}

	// 验证目标文件存在
	if !FileExists(dstFile) {
		t.Error("目标文件不存在")
	}

	// 验证内容一致
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("读取目标文件失败: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("目标文件内容 = %q; 期望 %q", string(dstContent), string(content))
	}
}

// TestCopyFile_SourceNotExist 测试复制不存在的源文件
func TestCopyFile_SourceNotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_copy_err_")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = CopyFile("/nonexistent/source.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("CopyFile 应返回错误当源文件不存在时")
	}
}

// =============================================================================
// CreateDirectory 函数测试
// =============================================================================

// TestCreateDirectory 测试创建目录
func TestCreateDirectory(t *testing.T) {
	// 创建临时基础目录
	tmpDir, err := os.MkdirTemp("", "test_mkdir_")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试创建嵌套目录
	newDir := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = CreateDirectory(newDir)
	if err != nil {
		t.Fatalf("CreateDirectory 失败: %v", err)
	}

	// 验证目录存在
	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("目录不存在: %v", err)
	}

	if !info.IsDir() {
		t.Error("创建的不是目录")
	}
}

// =============================================================================
// GetFileSize 函数测试
// =============================================================================

// TestGetFileSize 测试获取文件大小
func TestGetFileSize(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_size_*.txt")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer os.Remove(tmpFileName)

	// 写入已知大小的内容
	content := "Hello, World!" // 13 bytes
	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
	tmpFile.Close()

	// 测试获取文件大小
	size, err := GetFileSize(tmpFileName)
	if err != nil {
		t.Fatalf("GetFileSize 失败: %v", err)
	}

	expectedSize := int64(len(content))
	if size != expectedSize {
		t.Errorf("GetFileSize(%q) = %d; 期望 %d", tmpFileName, size, expectedSize)
	}
}

// TestGetFileSize_NotExist 测试获取不存在文件的大小
func TestGetFileSize_NotExist(t *testing.T) {
	_, err := GetFileSize("/nonexistent/file.txt")
	if err == nil {
		t.Error("GetFileSize 应返回错误当文件不存在时")
	}
}

// =============================================================================
// ListFilesInDirectory 函数测试
// =============================================================================

// TestListFilesInDirectory 测试列出目录中的文件
func TestListFilesInDirectory(t *testing.T) {
	// 创建临时目录结构
	tmpDir, err := os.MkdirTemp("", "test_list_")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试文件
	testFiles := []string{"file1.txt", "file2.go", "file3.md"}
	for _, f := range testFiles {
		filePath := filepath.Join(tmpDir, f)
		err := os.WriteFile(filePath, []byte("test"), 0600)
		if err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}
	}

	// 创建子目录和文件
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0700)
	if err != nil {
		t.Fatalf("创建子目录失败: %v", err)
	}
	err = os.WriteFile(filepath.Join(subDir, "subfile.txt"), []byte("test"), 0600)
	if err != nil {
		t.Fatalf("创建子目录文件失败: %v", err)
	}

	// 测试列出文件
	files, err := ListFilesInDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ListFilesInDirectory 失败: %v", err)
	}

	// 应该有4个文件（3个根目录 + 1个子目录）
	expectedCount := 4
	if len(files) != expectedCount {
		t.Errorf("文件数量 = %d; 期望 %d", len(files), expectedCount)
	}
}

// =============================================================================
// GetRelativePath 函数测试
// =============================================================================

// TestGetRelativePath 测试获取相对路径
func TestGetRelativePath(t *testing.T) {
	tests := []struct {
		name       string
		basePath   string
		targetPath string
		expected   string
	}{
		{
			"子目录",
			"/home/user",
			"/home/user/documents/file.txt",
			"documents/file.txt",
		},
		{
			"同级目录",
			"/home/user/dir1",
			"/home/user/dir2/file.txt",
			"../dir2/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRelativePath(tt.basePath, tt.targetPath)
			if err != nil {
				t.Fatalf("GetRelativePath 失败: %v", err)
			}
			// 使用 filepath.ToSlash 统一路径分隔符进行比较
			if filepath.ToSlash(result) != tt.expected {
				t.Errorf("GetRelativePath(%q, %q) = %q; 期望 %q",
					tt.basePath, tt.targetPath, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// EnsureDirectoryExists 函数测试
// =============================================================================

// TestEnsureDirectoryExists 测试确保目录存在
func TestEnsureDirectoryExists(t *testing.T) {
	// 创建临时基础目录
	tmpDir, err := os.MkdirTemp("", "test_ensure_")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 测试创建新目录
	newDir := filepath.Join(tmpDir, "new_dir")
	err = EnsureDirectoryExists(newDir)
	if err != nil {
		t.Fatalf("EnsureDirectoryExists 失败: %v", err)
	}

	if !FileExists(newDir) {
		t.Error("目录应该被创建")
	}

	// 测试已存在的目录（不应报错）
	err = EnsureDirectoryExists(newDir)
	if err != nil {
		t.Errorf("EnsureDirectoryExists 对已存在目录返回错误: %v", err)
	}
}

// =============================================================================
// 基准测试
// =============================================================================

// BenchmarkFileExists 文件存在检查基准测试
func BenchmarkFileExists(b *testing.B) {
	tmpFile, _ := os.CreateTemp("", "bench_*.txt")
	tmpFileName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpFileName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileExists(tmpFileName)
	}
}

// BenchmarkGetFileExtension 获取扩展名基准测试
func BenchmarkGetFileExtension(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetFileExtension("/path/to/file.txt")
	}
}
