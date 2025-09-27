package repository

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint           string        `json:"endpoint"`
	AccessKeyID        string        `json:"access_key_id"`
	SecretAccessKey    string        `json:"secret_access_key"`
	UseSSL             bool          `json:"use_ssl"`
	BucketName         string        `json:"bucket_name"`
	Region             string        `json:"region"`
	MaxRetries         int           `json:"max_retries"`
	RetryDelay         time.Duration `json:"retry_delay"`
	MultipartThreshold int64         `json:"multipart_threshold"` // 5MB
	PartSize           int64         `json:"part_size"`           // 5MB
	Timeout            time.Duration `json:"timeout"`
}

// DefaultMinIOConfig 默认MinIO配置
func DefaultMinIOConfig() *MinIOConfig {
	return &MinIOConfig{
		Endpoint:           "localhost:9000",
		AccessKeyID:        "minioadmin",
		SecretAccessKey:    "minioadmin",
		UseSSL:             false,
		BucketName:         "file-storage",
		Region:             "us-east-1",
		MaxRetries:         3,
		RetryDelay:         time.Second,
		MultipartThreshold: 5 * 1024 * 1024, // 5MB
		PartSize:           5 * 1024 * 1024, // 5MB
		Timeout:            30 * time.Second,
	}
}

// ObjectPathBuilder 对象路径构建器
type ObjectPathBuilder struct{}

// NewObjectPathBuilder 创建对象路径构建器
func NewObjectPathBuilder() *ObjectPathBuilder {
	return &ObjectPathBuilder{}
}

// BuildPath 构建对象路径
// 格式: users/{userID}/{year}/{month}/{uuid}/{filename}
func (b *ObjectPathBuilder) BuildPath(userID uint, uuid, filename string) string {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// 清理文件名，防止路径遍历攻击
	cleanFilename := b.sanitizeFilename(filename)

	return fmt.Sprintf("users/%d/%s/%s/%s/%s", userID, year, month, uuid, cleanFilename)
}

// sanitizeFilename 清理文件名
func (b *ObjectPathBuilder) sanitizeFilename(filename string) string {
	// 移除危险字符
	cleaned := strings.ReplaceAll(filename, "../", "")
	cleaned = strings.ReplaceAll(cleaned, "..\\", "")
	cleaned = strings.ReplaceAll(cleaned, "/", "_")
	cleaned = strings.ReplaceAll(cleaned, "\\", "_")

	// 如果文件名为空，使用默认名称
	if cleaned == "" {
		cleaned = "unnamed_file"
	}

	return cleaned
}

// ValidateObjectKey 验证对象key
func (b *ObjectPathBuilder) ValidateObjectKey(key string) error {
	if strings.Contains(key, "../") || strings.Contains(key, "..\\") {
		return fmt.Errorf("invalid object key: path traversal detected")
	}
	if len(key) > 1024 {
		return fmt.Errorf("invalid object key: too long")
	}
	return nil
}

// minIOStorageRepository MinIO存储仓储实现
type minIOStorageRepository struct {
	client      *minio.Client
	config      *MinIOConfig
	pathBuilder *ObjectPathBuilder
}

// NewMinIOStorageRepository 创建MinIO存储仓储
func NewMinIOStorageRepository(config *MinIOConfig) (StorageRepository, error) {
	if config == nil {
		config = DefaultMinIOConfig()
	}

	// 创建MinIO客户端
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	repo := &minIOStorageRepository{
		client:      client,
		config:      config,
		pathBuilder: NewObjectPathBuilder(),
	}

	// 确保存储桶存在
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	if err := repo.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return repo, nil
}

// ensureBucket 确保存储桶存在
func (r *minIOStorageRepository) ensureBucket(ctx context.Context) error {
	exists, err := r.client.BucketExists(ctx, r.config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = r.client.MakeBucket(ctx, r.config.BucketName, minio.MakeBucketOptions{
			Region: r.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// Upload 上传文件
func (r *minIOStorageRepository) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return err
	}

	// 设置超时
	uploadCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	// 上传选项
	options := minio.PutObjectOptions{
		ContentType: contentType,
	}

	// 如果文件大小超过阈值，使用分片上传
	if size > r.config.MultipartThreshold {
		options.PartSize = uint64(r.config.PartSize)
	}

	_, err := r.client.PutObject(uploadCtx, r.config.BucketName, key, data, size, options)
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", key, err)
	}

	return nil
}

// Download 下载文件
func (r *minIOStorageRepository) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return nil, err
	}

	// 设置超时
	downloadCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	object, err := r.client.GetObject(downloadCtx, r.config.BucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %w", key, err)
	}

	return object, nil
}

// Exists 检查文件是否存在
func (r *minIOStorageRepository) Exists(ctx context.Context, key string) (bool, error) {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return false, err
	}

	// 设置超时
	statCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	_, err := r.client.StatObject(statCtx, r.config.BucketName, key, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence %s: %w", key, err)
	}

	return true, nil
}

// Delete 删除文件
func (r *minIOStorageRepository) Delete(ctx context.Context, key string) error {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return err
	}

	// 设置超时
	deleteCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	err := r.client.RemoveObject(deleteCtx, r.config.BucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (r *minIOStorageRepository) GetObjectInfo(ctx context.Context, key string) (*ObjectInfo, error) {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return nil, err
	}

	// 设置超时
	statCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	info, err := r.client.StatObject(statCtx, r.config.BucketName, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object info %s: %w", key, err)
	}

	return &ObjectInfo{
		Key:          key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		ETag:         info.ETag,
	}, nil
}

// GeneratePresignedURL 生成预签名URL
func (r *minIOStorageRepository) GeneratePresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return "", err
	}

	// 限制最大过期时间为7天
	if expires > 7*24*time.Hour {
		expires = 7 * 24 * time.Hour
	}

	url, err := r.client.PresignedGetObject(ctx, r.config.BucketName, key, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for %s: %w", key, err)
	}

	return url.String(), nil
}

// UploadMultipart 分片上传（这里简化实现，实际使用PutObject的分片功能）
func (r *minIOStorageRepository) UploadMultipart(ctx context.Context, key string, parts []io.Reader, contentType string) error {
	// 验证对象key
	if err := r.pathBuilder.ValidateObjectKey(key); err != nil {
		return err
	}

	// 将所有分片合并为一个Reader
	readers := make([]io.Reader, len(parts))
	copy(readers, parts)

	multiReader := io.MultiReader(readers...)

	// 使用普通上传，MinIO会自动处理分片
	return r.Upload(ctx, key, multiReader, -1, contentType)
}

// ListObjects 列出对象
func (r *minIOStorageRepository) ListObjects(ctx context.Context, prefix string, limit int) ([]string, error) {
	// 设置超时
	listCtx, cancel := context.WithTimeout(ctx, r.config.Timeout)
	defer cancel()

	var objects []string
	count := 0

	for object := range r.client.ListObjects(listCtx, r.config.BucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		objects = append(objects, object.Key)
		count++

		if limit > 0 && count >= limit {
			break
		}
	}

	return objects, nil
}
