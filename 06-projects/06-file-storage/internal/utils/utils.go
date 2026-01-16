package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

// GenerateFileID 生成文件ID
func GenerateFileID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%x", timestamp, randomBytes)))
	return hex.EncodeToString(hash[:])[:16]
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)

	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

// DetectMimeType 根据文件扩展名检测MIME类型
func DetectMimeType(extension string) string {
	mimeType := mime.TypeByExtension(extension)
	if mimeType == "" {
		// 默认类型
		return "application/octet-stream"
	}
	return mimeType
}

// SanitizeFilename 清理文件名
func SanitizeFilename(filename string) string {
	// 移除路径分隔符
	filename = filepath.Base(filename)

	// 替换危险字符
	dangerous := []string{"<", ">", ":", "\"", "|", "?", "*", "\\", "/"}
	for _, char := range dangerous {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// 限制长度
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		maxNameLength := 255 - len(ext)
		if maxNameLength > 0 {
			filename = nameWithoutExt[:maxNameLength] + ext
		}
	}

	return filename
}

// FormatFileSize 格式化文件大小
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	sizes := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), sizes[exp])
}

// IsImageType 检查是否为图片类型
func IsImageType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsVideoType 检查是否为视频类型
func IsVideoType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

// IsAudioType 检查是否为音频类型
func IsAudioType(mimeType string) bool {
	return strings.HasPrefix(mimeType, "audio/")
}

// GetFileTypeIcon 获取文件类型图标
func GetFileTypeIcon(mimeType string) string {
	switch {
	case IsImageType(mimeType):
		return "🖼️"
	case IsVideoType(mimeType):
		return "🎥"
	case IsAudioType(mimeType):
		return "🎵"
	case strings.Contains(mimeType, "pdf"):
		return "📄"
	case strings.Contains(mimeType, "text"):
		return "📝"
	case strings.Contains(mimeType, "zip") || strings.Contains(mimeType, "archive"):
		return "📦"
	case strings.Contains(mimeType, "word"):
		return "📘"
	case strings.Contains(mimeType, "excel") || strings.Contains(mimeType, "spreadsheet"):
		return "📊"
	case strings.Contains(mimeType, "powerpoint") || strings.Contains(mimeType, "presentation"):
		return "📽️"
	default:
		return "📁"
	}
}

// ValidateFileExtension 验证文件扩展名
func ValidateFileExtension(filename string, allowedExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, allowed := range allowedExtensions {
		if ext == strings.ToLower(allowed) {
			return true
		}
	}
	return false
}

// GetClientIP 获取客户端IP地址
func GetClientIP(remoteAddr, xForwardedFor, xRealIP string) string {
	// 优先使用 X-Forwarded-For
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// 其次使用 X-Real-IP
	if xRealIP != "" {
		return xRealIP
	}

	// 最后使用 RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		return remoteAddr[:idx]
	}

	return remoteAddr
}
