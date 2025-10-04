package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// PathValidationError 路径验证错误
type PathValidationError struct {
	Path   string
	Reason string
}

func (e *PathValidationError) Error() string {
	return fmt.Sprintf("unsafe path '%s': %s", e.Path, e.Reason)
}

// SecurePathOptions 安全路径选项
type SecurePathOptions struct {
	AllowAbsolute bool // 是否允许绝对路径
	AllowDotDot   bool // 是否允许..路径
	MaxDepth      int  // 最大路径深度（0表示不限制）
}

var (
	ErrPathTraversal     = errors.New("path contains directory traversal")
	ErrAbsolutePath      = errors.New("absolute paths not allowed")
	ErrEmptyPath         = errors.New("path cannot be empty")
	ErrInvalidCharacters = errors.New("path contains invalid characters")
	ErrPathTooDeep       = errors.New("path exceeds maximum depth")
)

// ValidateSecurePath 验证路径是否安全，防止目录遍历攻击
func ValidateSecurePath(path string, opts *SecurePathOptions) error {
	if opts == nil {
		opts = &SecurePathOptions{
			AllowAbsolute: false,
			AllowDotDot:   false,
			MaxDepth:      10, // 默认最大深度
		}
	}

	if path == "" {
		return &PathValidationError{Path: path, Reason: ErrEmptyPath.Error()}
	}

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查绝对路径
	if filepath.IsAbs(cleanPath) && !opts.AllowAbsolute {
		return &PathValidationError{Path: path, Reason: ErrAbsolutePath.Error()}
	}

	// 检查目录遍历
	if !opts.AllowDotDot && strings.Contains(cleanPath, "..") {
		return &PathValidationError{Path: path, Reason: ErrPathTraversal.Error()}
	}

	// 检查危险字符
	if err := validatePathCharacters(cleanPath); err != nil {
		return &PathValidationError{Path: path, Reason: err.Error()}
	}

	// 检查路径深度
	if opts.MaxDepth > 0 {
		depth := strings.Count(cleanPath, string(filepath.Separator))
		if depth > opts.MaxDepth {
			return &PathValidationError{Path: path, Reason: ErrPathTooDeep.Error()}
		}
	}

	return nil
}

// SecureJoinPath 安全地连接路径，防止目录遍历
func SecureJoinPath(base string, elem ...string) (string, error) {
	// 验证基础路径
	if base == "" {
		return "", &PathValidationError{Path: base, Reason: "base path cannot be empty"}
	}

	cleanBase := filepath.Clean(base)

	// 验证所有路径元素
	for i, e := range elem {
		if err := ValidateSecurePath(e, &SecurePathOptions{
			AllowAbsolute: false,
			AllowDotDot:   false,
			MaxDepth:      20,
		}); err != nil {
			return "", fmt.Errorf("invalid path element at index %d: %w", i, err)
		}
	}

	// 安全连接路径
	fullPath := filepath.Join(cleanBase, filepath.Join(elem...))
	cleanFullPath := filepath.Clean(fullPath)

	// 确保结果路径仍在基础路径下
	if !strings.HasPrefix(cleanFullPath, cleanBase) {
		return "", &PathValidationError{
			Path:   fullPath,
			Reason: "path escapes base directory",
		}
	}

	return cleanFullPath, nil
}

// ValidatePathWithinBase 验证路径是否在指定的基础目录内
func ValidatePathWithinBase(fullPath, basePath string) error {
	cleanFull := filepath.Clean(fullPath)
	cleanBase := filepath.Clean(basePath)

	// 确保基础路径以分隔符结尾
	if !strings.HasSuffix(cleanBase, string(filepath.Separator)) {
		cleanBase += string(filepath.Separator)
	}

	if !strings.HasPrefix(cleanFull+string(filepath.Separator), cleanBase) {
		return &PathValidationError{
			Path:   fullPath,
			Reason: fmt.Sprintf("path escapes base directory '%s'", basePath),
		}
	}

	return nil
}

// SanitizePath 清理并验证路径
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", &PathValidationError{Path: path, Reason: ErrEmptyPath.Error()}
	}

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 移除多余的分隔符
	for strings.Contains(cleanPath, "//") {
		cleanPath = strings.ReplaceAll(cleanPath, "//", "/")
	}

	// 验证清理后的路径
	if err := ValidateSecurePath(cleanPath, nil); err != nil {
		return "", err
	}

	return cleanPath, nil
}

// validatePathCharacters 验证路径字符是否安全
func validatePathCharacters(path string) error {
	// 检查空字节（NULL字节注入攻击）
	if strings.ContainsRune(path, '\x00') {
		return ErrInvalidCharacters
	}

	// 检查其他危险字符
	dangerousChars := []string{
		"\n", "\r", // 换行符
		"\t",          // 制表符
		"|", "&", ";", // 命令注入字符
		"<", ">", // 重定向字符
		"*", "?", // 通配符（在某些场景下危险）
	}

	for _, char := range dangerousChars {
		if strings.Contains(path, char) {
			return ErrInvalidCharacters
		}
	}

	return nil
}

// GetSafePath 获取安全的文件路径，包含完整的验证和清理
func GetSafePath(basePath, userPath string) (string, error) {
	// 验证基础路径
	if basePath == "" {
		return "", &PathValidationError{Path: basePath, Reason: "base path cannot be empty"}
	}

	// 验证用户提供的路径
	if err := ValidateSecurePath(userPath, &SecurePathOptions{
		AllowAbsolute: false,
		AllowDotDot:   false,
		MaxDepth:      15,
	}); err != nil {
		return "", fmt.Errorf("user path validation failed: %w", err)
	}

	// 安全连接路径
	fullPath, err := SecureJoinPath(basePath, userPath)
	if err != nil {
		return "", fmt.Errorf("path join failed: %w", err)
	}

	// 最终验证
	if err := ValidatePathWithinBase(fullPath, basePath); err != nil {
		return "", fmt.Errorf("final validation failed: %w", err)
	}

	return fullPath, nil
}

// IsPathSafe 快速检查路径是否安全（用于日志记录等场景）
func IsPathSafe(path string) bool {
	return ValidateSecurePath(path, nil) == nil
}
