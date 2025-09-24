package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileExists 检查文件是否存在
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// ReadFileLines 读取文件的所有行
func ReadFileLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// WriteFileLines 将字符串切片写入文件
func WriteFileLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetFilenameWithoutExtension 获取不带扩展名的文件名
func GetFilenameWithoutExtension(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// CreateDirectory 创建目录（如果不存在）
func CreateDirectory(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ListFilesInDirectory 列出目录中的所有文件
func ListFilesInDirectory(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// GetRelativePath 获取相对路径
func GetRelativePath(basePath, targetPath string) (string, error) {
	return filepath.Rel(basePath, targetPath)
}

// EnsureDirectoryExists 确保目录存在，如果不存在则创建
func EnsureDirectoryExists(path string) error {
	if !FileExists(path) {
		return CreateDirectory(path)
	}
	return nil
}

func init() {
	fmt.Println("utils file 模块初始化")
}
