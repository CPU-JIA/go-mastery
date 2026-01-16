package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"file-storage-service/internal/config"
	"go-mastery/common/security"
)

// LocalStorage 本地文件存储实现
type LocalStorage struct {
	basePath string
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(config config.StorageConfig) (FileStorage, error) {
	basePath := config.LocalPath
	if basePath == "" {
		basePath = "./uploads"
	}

	// 创建目录
	// #nosec G301 -- 本地存储基础目录，需要0755权限支持文件服务访问
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
	}, nil
}

// Save 保存文件
func (ls *LocalStorage) Save(ctx context.Context, path string, content io.Reader, size int64) error {
	fullPath := filepath.Join(ls.basePath, path)

	// 确保目录存在
	// #nosec G301 -- 文件保存时创建父目录，需要0755权限支持Web服务器访问
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	// G301/G306安全修复：使用安全权限创建文件
	file, err := security.SecureCreateFile(fullPath, security.DefaultFileMode)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 复制内容
	_, err = io.Copy(file, content)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Get 获取文件
func (ls *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	// G304安全修复：使用security.GetSafePath防止路径遍历攻击
	fullPath, err := security.GetSafePath(ls.basePath, path)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	// #nosec G304 -- 路径已通过security.GetSafePath()验证，确保在basePath范围内
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (ls *LocalStorage) Delete(ctx context.Context, path string) error {
	// G304安全修复：使用security.GetSafePath防止路径遍历攻击
	fullPath, err := security.GetSafePath(ls.basePath, path)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (ls *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	// G304安全修复：使用security.GetSafePath防止路径遍历攻击
	fullPath, err := security.GetSafePath(ls.basePath, path)
	if err != nil {
		return false, fmt.Errorf("invalid file path: %w", err)
	}

	_, err = os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Size 获取文件大小
func (ls *LocalStorage) Size(ctx context.Context, path string) (int64, error) {
	// G304安全修复：使用security.GetSafePath防止路径遍历攻击
	fullPath, err := security.GetSafePath(ls.basePath, path)
	if err != nil {
		return 0, fmt.Errorf("invalid file path: %w", err)
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return stat.Size(), nil
}

// SaveMultiple 批量保存文件
func (ls *LocalStorage) SaveMultiple(ctx context.Context, files map[string]io.Reader) error {
	for path, content := range files {
		if err := ls.Save(ctx, path, content, 0); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMultiple 批量删除文件
func (ls *LocalStorage) DeleteMultiple(ctx context.Context, paths []string) error {
	for _, path := range paths {
		if err := ls.Delete(ctx, path); err != nil {
			return err
		}
	}
	return nil
}

// GetMetadata 获取文件元数据
func (ls *LocalStorage) GetMetadata(ctx context.Context, path string) (map[string]string, error) {
	// G304安全修复：使用security.GetSafePath防止路径遍历攻击
	fullPath, err := security.GetSafePath(ls.basePath, path)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	stat, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	metadata := map[string]string{
		"size":          fmt.Sprintf("%d", stat.Size()),
		"last_modified": fmt.Sprintf("%d", stat.ModTime().Unix()),
		"mode":          stat.Mode().String(),
	}

	return metadata, nil
}

// SetMetadata 设置文件元数据（本地存储不支持）
func (ls *LocalStorage) SetMetadata(ctx context.Context, path string, metadata map[string]string) error {
	// 本地文件系统不支持自定义元数据
	return fmt.Errorf("local storage does not support custom metadata")
}

// GetPresignedUploadURL 获取预签名上传URL（本地存储不支持）
func (ls *LocalStorage) GetPresignedUploadURL(ctx context.Context, path string, expiry int64) (string, error) {
	return "", fmt.Errorf("local storage does not support presigned URLs")
}

// GetPresignedDownloadURL 获取预签名下载URL（本地存储不支持）
func (ls *LocalStorage) GetPresignedDownloadURL(ctx context.Context, path string, expiry int64) (string, error) {
	return "", fmt.Errorf("local storage does not support presigned URLs")
}

// List 列出文件
func (ls *LocalStorage) List(ctx context.Context, prefix string, maxKeys int) ([]FileInfo, error) {
	fullPrefix := filepath.Join(ls.basePath, prefix)

	var files []FileInfo
	count := 0

	err := filepath.Walk(fullPrefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if maxKeys > 0 && count >= maxKeys {
			return filepath.SkipDir
		}

		// 获取相对路径
		relPath, err := filepath.Rel(ls.basePath, path)
		if err != nil {
			return err
		}

		files = append(files, FileInfo{
			Path:         relPath,
			Size:         info.Size(),
			LastModified: info.ModTime().Unix(),
		})

		count++
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// HealthCheck 健康检查
func (ls *LocalStorage) HealthCheck(ctx context.Context) error {
	// 检查存储目录是否可访问
	_, err := os.Stat(ls.basePath)
	if err != nil {
		return fmt.Errorf("storage directory not accessible: %w", err)
	}

	// 尝试创建临时文件
	tempFile := filepath.Join(ls.basePath, ".health_check_"+fmt.Sprintf("%d", time.Now().UnixNano()))
	// G301/G306安全修复：使用安全权限创建临时文件
	file, err := security.SecureCreateFile(tempFile, security.GetRecommendedMode("temp"))
	if err != nil {
		return fmt.Errorf("cannot create files in storage directory: %w", err)
	}
	file.Close()
	os.Remove(tempFile)

	return nil
}
