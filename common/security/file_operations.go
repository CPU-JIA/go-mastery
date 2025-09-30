package security

import (
	"fmt"
	"io/fs"
	"os"
)

// SecureFileMode 定义安全的文件权限模式
type SecureFileMode fs.FileMode

const (
	// 安全的文件权限模式，遵循最小权限原则
	SecureFileMode_ReadOnlyUser    SecureFileMode = 0400 // 仅所有者可读
	SecureFileMode_ReadWriteUser   SecureFileMode = 0600 // 仅所有者可读写
	SecureFileMode_ReadOnlyAll     SecureFileMode = 0444 // 所有用户只读
	SecureFileMode_ReadWriteAll    SecureFileMode = 0666 // 所有用户读写（不推荐）
	SecureFileMode_ExecutableUser  SecureFileMode = 0700 // 仅所有者可执行
	SecureFileMode_ExecutableAll   SecureFileMode = 0755 // 所有者可执行，其他用户可读执行

	// 默认推荐的安全权限
	DefaultFileMode      = SecureFileMode_ReadWriteUser   // 0600 - 文件默认权限
	DefaultDirMode       = SecureFileMode_ExecutableUser  // 0700 - 目录默认权限
	DefaultConfigMode    = SecureFileMode_ReadOnlyUser    // 0400 - 配置文件权限
	DefaultLogMode       = SecureFileMode_ReadWriteUser   // 0600 - 日志文件权限
	DefaultTempMode      = SecureFileMode_ReadWriteUser   // 0600 - 临时文件权限
	DefaultExecutableMode = SecureFileMode_ExecutableUser // 0700 - 可执行文件权限
)

// SecureFileOptions 安全文件操作选项
type SecureFileOptions struct {
	Mode      SecureFileMode
	CreateDir bool // 是否创建父目录
}

// SecureWriteFile 安全写入文件，使用适当的权限
func SecureWriteFile(filename string, data []byte, opts *SecureFileOptions) error {
	if opts == nil {
		opts = &SecureFileOptions{Mode: DefaultFileMode}
	}

	// 如果需要，创建父目录
	if opts.CreateDir {
		dir := filename[:len(filename)-len(filename[len(filename)-1:])]
		if err := os.MkdirAll(dir, fs.FileMode(DefaultDirMode)); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return os.WriteFile(filename, data, fs.FileMode(opts.Mode))
}

// SecureCreateFile 安全创建文件，使用适当的权限
func SecureCreateFile(filename string, mode SecureFileMode) (*os.File, error) {
	// 使用OpenFile而不是Create，以确保指定正确的权限
	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(mode))
}

// SecureOpenFile 安全打开文件，验证权限
func SecureOpenFile(filename string, flag int, mode SecureFileMode) (*os.File, error) {
	return os.OpenFile(filename, flag, fs.FileMode(mode))
}

// SecureMkdirAll 安全创建目录，使用适当的权限
func SecureMkdirAll(path string, mode SecureFileMode) error {
	return os.MkdirAll(path, fs.FileMode(mode))
}

// ValidateFilePermissions 验证文件权限是否安全
func ValidateFilePermissions(filename string, expectedMode SecureFileMode) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	actualMode := info.Mode().Perm()
	expected := fs.FileMode(expectedMode)

	if actualMode != expected {
		return fmt.Errorf("file permissions are not secure: got %o, expected %o", actualMode, expected)
	}

	return nil
}

// GetRecommendedMode 根据文件类型返回推荐的权限模式
func GetRecommendedMode(fileType string) SecureFileMode {
	switch fileType {
	case "config", "configuration":
		return DefaultConfigMode
	case "log", "logs":
		return DefaultLogMode
	case "temp", "temporary":
		return DefaultTempMode
	case "executable", "binary":
		return DefaultExecutableMode
	case "data", "json", "yaml", "xml":
		return DefaultFileMode
	default:
		return DefaultFileMode
	}
}

// IsSecurePermission 检查文件权限是否安全
func IsSecurePermission(mode fs.FileMode) bool {
	perm := mode.Perm()

	// 检查是否给其他用户过多权限
	if perm&0007 == 0007 { // 其他用户有读写执行权限
		return false
	}

	if perm&0002 != 0 { // 其他用户有写权限
		return false
	}

	// 检查组权限是否过于宽松
	if perm&0070 == 0070 { // 组用户有读写执行权限
		return false
	}

	return true
}