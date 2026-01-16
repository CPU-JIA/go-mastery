package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"file-storage-service/internal/config"
)

// MinIOStorage MinIO对象存储实现
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
}

// NewMinIOStorage 创建MinIO存储实例
func NewMinIOStorage(config config.StorageConfig) (FileStorage, error) {
	// 创建MinIO客户端
	client, err := minio.New(config.MinIOConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.MinIOConfig.AccessKey, config.MinIOConfig.SecretKey, ""),
		Secure: config.MinIOConfig.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	storage := &MinIOStorage{
		client:     client,
		bucketName: config.MinIOConfig.BucketName,
	}

	// 确保存储桶存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, storage.bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, storage.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return storage, nil
}

// Save 保存文件到MinIO
func (ms *MinIOStorage) Save(ctx context.Context, path string, content io.Reader, size int64) error {
	_, err := ms.client.PutObject(ctx, ms.bucketName, path, content, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to save object to MinIO: %w", err)
	}
	return nil
}

// Get 从MinIO获取文件
func (ms *MinIOStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	object, err := ms.client.GetObject(ctx, ms.bucketName, path, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from MinIO: %w", err)
	}
	return object, nil
}

// Delete 从MinIO删除文件
func (ms *MinIOStorage) Delete(ctx context.Context, path string) error {
	err := ms.client.RemoveObject(ctx, ms.bucketName, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from MinIO: %w", err)
	}
	return nil
}

// Exists 检查MinIO中文件是否存在
func (ms *MinIOStorage) Exists(ctx context.Context, path string) (bool, error) {
	_, err := ms.client.StatObject(ctx, ms.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// Size 获取MinIO中文件大小
func (ms *MinIOStorage) Size(ctx context.Context, path string) (int64, error) {
	stat, err := ms.client.StatObject(ctx, ms.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get object size: %w", err)
	}
	return stat.Size, nil
}

// SaveMultiple 批量保存到MinIO
func (ms *MinIOStorage) SaveMultiple(ctx context.Context, files map[string]io.Reader) error {
	for path, content := range files {
		if err := ms.Save(ctx, path, content, -1); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMultiple 批量删除MinIO中的文件
func (ms *MinIOStorage) DeleteMultiple(ctx context.Context, paths []string) error {
	objectsCh := make(chan minio.ObjectInfo, len(paths))

	// 发送要删除的对象
	go func() {
		defer close(objectsCh)
		for _, path := range paths {
			objectsCh <- minio.ObjectInfo{Key: path}
		}
	}()

	// 批量删除
	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for rErr := range ms.client.RemoveObjects(ctx, ms.bucketName, objectsCh, opts) {
		if rErr.Err != nil {
			return fmt.Errorf("failed to delete object %s: %w", rErr.ObjectName, rErr.Err)
		}
	}

	return nil
}

// GetMetadata 获取MinIO对象元数据
func (ms *MinIOStorage) GetMetadata(ctx context.Context, path string) (map[string]string, error) {
	stat, err := ms.client.StatObject(ctx, ms.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	metadata := make(map[string]string)
	metadata["size"] = fmt.Sprintf("%d", stat.Size)
	metadata["last_modified"] = stat.LastModified.Format(time.RFC3339)
	metadata["etag"] = stat.ETag
	metadata["content_type"] = stat.ContentType

	// 添加用户定义的元数据
	for key, value := range stat.UserMetadata {
		metadata[key] = value
	}

	return metadata, nil
}

// SetMetadata 设置MinIO对象元数据
func (ms *MinIOStorage) SetMetadata(ctx context.Context, path string, metadata map[string]string) error {
	// MinIO不支持直接修改元数据，需要复制对象
	srcOpts := minio.CopySrcOptions{
		Bucket: ms.bucketName,
		Object: path,
	}

	dstOpts := minio.CopyDestOptions{
		Bucket:       ms.bucketName,
		Object:       path,
		UserMetadata: metadata,
	}

	_, err := ms.client.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to set object metadata: %w", err)
	}

	return nil
}

// GetPresignedUploadURL 获取预签名上传URL
func (ms *MinIOStorage) GetPresignedUploadURL(ctx context.Context, path string, expiry int64) (string, error) {
	expires := time.Duration(expiry) * time.Second
	url, err := ms.client.PresignedPutObject(ctx, ms.bucketName, path, expires)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}
	return url.String(), nil
}

// GetPresignedDownloadURL 获取预签名下载URL
func (ms *MinIOStorage) GetPresignedDownloadURL(ctx context.Context, path string, expiry int64) (string, error) {
	expires := time.Duration(expiry) * time.Second
	url, err := ms.client.PresignedGetObject(ctx, ms.bucketName, path, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned download URL: %w", err)
	}
	return url.String(), nil
}

// List 列出MinIO中的文件
func (ms *MinIOStorage) List(ctx context.Context, prefix string, maxKeys int) ([]FileInfo, error) {
	var files []FileInfo

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
		MaxKeys:   maxKeys,
	}

	for object := range ms.client.ListObjects(ctx, ms.bucketName, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}

		files = append(files, FileInfo{
			Path:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified.Unix(),
			ETag:         object.ETag,
		})
	}

	return files, nil
}

// HealthCheck MinIO健康检查
func (ms *MinIOStorage) HealthCheck(ctx context.Context) error {
	// 检查存储桶是否可访问
	exists, err := ms.client.BucketExists(ctx, ms.bucketName)
	if err != nil {
		return fmt.Errorf("MinIO health check failed: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", ms.bucketName)
	}

	return nil
}
