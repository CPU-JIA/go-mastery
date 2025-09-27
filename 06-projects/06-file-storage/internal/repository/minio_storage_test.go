package repository

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectPathBuilder(t *testing.T) {
	builder := NewObjectPathBuilder()

	t.Run("Should build valid object path", func(t *testing.T) {
		path := builder.BuildPath(123, "uuid-123", "test.jpg")

		// 验证路径格式：users/{userID}/{year}/{month}/{uuid}/{filename}
		parts := strings.Split(path, "/")
		assert.Len(t, parts, 6) // 修正：应该是6段
		assert.Equal(t, "users", parts[0])
		assert.Equal(t, "123", parts[1])
		assert.Len(t, parts[2], 4) // year
		assert.Len(t, parts[3], 2) // month
		assert.Equal(t, "uuid-123", parts[4])
		assert.Equal(t, "test.jpg", parts[5])

		// 验证不包含危险字符
		assert.NotContains(t, path, "../")
		assert.NotContains(t, path, "..\\")
	})

	t.Run("Should sanitize dangerous filenames", func(t *testing.T) {
		// 测试路径遍历攻击
		path1 := builder.BuildPath(123, "uuid-123", "../../../etc/passwd")
		assert.NotContains(t, path1, "../")

		// 测试Windows路径遍历
		path2 := builder.BuildPath(123, "uuid-123", "..\\..\\windows\\system32")
		assert.NotContains(t, path2, "..\\")
		assert.NotContains(t, path2, "\\")

		// 测试正常斜杠替换
		path3 := builder.BuildPath(123, "uuid-123", "folder/file.txt")
		assert.Contains(t, path3, "folder_file.txt")
	})

	t.Run("Should validate object keys", func(t *testing.T) {
		// 有效key
		err := builder.ValidateObjectKey("users/123/2024/09/uuid-123/test.jpg")
		assert.NoError(t, err)

		// 路径遍历攻击
		err = builder.ValidateObjectKey("users/../../../etc/passwd")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal detected")

		// 过长key
		longKey := strings.Repeat("a", 1025)
		err = builder.ValidateObjectKey(longKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})
}

func TestMinIOConfig(t *testing.T) {
	t.Run("Should create default config", func(t *testing.T) {
		config := DefaultMinIOConfig()

		assert.Equal(t, "localhost:9000", config.Endpoint)
		assert.Equal(t, "minioadmin", config.AccessKeyID)
		assert.Equal(t, "file-storage", config.BucketName)
		assert.Equal(t, int64(5*1024*1024), config.MultipartThreshold)
		assert.Equal(t, 3, config.MaxRetries)
		assert.False(t, config.UseSSL) // 测试环境不使用SSL
	})
}

// TestMinIOStorageRepository 需要运行MinIO服务器才能测试
// 这里我们创建一个Mock测试和集成测试分离的设计
func TestMinIOStorageRepositoryConfig(t *testing.T) {
	t.Run("Should validate config on creation", func(t *testing.T) {
		// 测试无效配置
		invalidConfig := &MinIOConfig{
			Endpoint:        "", // 空endpoint应该失败
			AccessKeyID:     "test",
			SecretAccessKey: "test",
			BucketName:      "test",
		}

		_, err := NewMinIOStorageRepository(invalidConfig)
		assert.Error(t, err)
	})

	t.Run("Should use default config when nil provided", func(t *testing.T) {
		// 由于没有MinIO服务器，这个测试会失败，但我们可以验证配置逻辑
		_, err := NewMinIOStorageRepository(nil)
		assert.Error(t, err)                            // 预期失败，因为没有MinIO服务器
		assert.Contains(t, err.Error(), "minio client") // 但错误应该是连接失败，不是配置问题
	})
}

// MockMinIOStorageRepository 用于单元测试的Mock实现
type MockMinIOStorageRepository struct {
	objects      map[string][]byte
	objectInfo   map[string]*ObjectInfo
	shouldFailOn string
	pathBuilder  *ObjectPathBuilder
}

func NewMockMinIOStorageRepository() *MockMinIOStorageRepository {
	return &MockMinIOStorageRepository{
		objects:     make(map[string][]byte),
		objectInfo:  make(map[string]*ObjectInfo),
		pathBuilder: NewObjectPathBuilder(),
	}
}

func (m *MockMinIOStorageRepository) SetFailure(operation string) {
	m.shouldFailOn = operation
}

func (m *MockMinIOStorageRepository) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	if m.shouldFailOn == "upload" {
		return fmt.Errorf("mock upload failure")
	}

	if err := m.pathBuilder.ValidateObjectKey(key); err != nil {
		return err
	}

	content, err := io.ReadAll(data)
	if err != nil {
		return err
	}

	m.objects[key] = content
	m.objectInfo[key] = &ObjectInfo{
		Key:          key,
		Size:         int64(len(content)),
		ContentType:  contentType,
		LastModified: time.Now(),
		ETag:         "mock-etag",
	}

	return nil
}

func (m *MockMinIOStorageRepository) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.shouldFailOn == "download" {
		return nil, fmt.Errorf("mock download failure")
	}

	content, exists := m.objects[key]
	if !exists {
		return nil, fmt.Errorf("object not found")
	}

	return io.NopCloser(strings.NewReader(string(content))), nil
}

func (m *MockMinIOStorageRepository) Exists(ctx context.Context, key string) (bool, error) {
	if m.shouldFailOn == "exists" {
		return false, fmt.Errorf("mock exists failure")
	}

	_, exists := m.objects[key]
	return exists, nil
}

func (m *MockMinIOStorageRepository) Delete(ctx context.Context, key string) error {
	if m.shouldFailOn == "delete" {
		return fmt.Errorf("mock delete failure")
	}

	delete(m.objects, key)
	delete(m.objectInfo, key)
	return nil
}

func (m *MockMinIOStorageRepository) GetObjectInfo(ctx context.Context, key string) (*ObjectInfo, error) {
	if m.shouldFailOn == "info" {
		return nil, fmt.Errorf("mock info failure")
	}

	info, exists := m.objectInfo[key]
	if !exists {
		return nil, fmt.Errorf("object not found")
	}

	return info, nil
}

func (m *MockMinIOStorageRepository) GeneratePresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if m.shouldFailOn == "presign" {
		return "", fmt.Errorf("mock presign failure")
	}

	// 检查对象是否存在
	if _, exists := m.objects[key]; !exists {
		return "", fmt.Errorf("object not found")
	}

	return fmt.Sprintf("https://mock-minio.example.com/%s?expires=%d", key, expires/time.Second), nil
}

func (m *MockMinIOStorageRepository) UploadMultipart(ctx context.Context, key string, parts []io.Reader, contentType string) error {
	if m.shouldFailOn == "multipart" {
		return fmt.Errorf("mock multipart failure")
	}

	// 合并所有分片
	var allContent []byte
	for _, part := range parts {
		content, err := io.ReadAll(part)
		if err != nil {
			return err
		}
		allContent = append(allContent, content...)
	}

	return m.Upload(ctx, key, strings.NewReader(string(allContent)), int64(len(allContent)), contentType)
}

func (m *MockMinIOStorageRepository) ListObjects(ctx context.Context, prefix string, limit int) ([]string, error) {
	if m.shouldFailOn == "list" {
		return nil, fmt.Errorf("mock list failure")
	}

	var keys []string
	count := 0

	for key := range m.objects {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}

	return keys, nil
}

func TestMockMinIOStorageRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Should upload and download file", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		key := "users/123/2024/09/uuid-123/test.txt"
		content := "Hello, MinIO!"

		// 上传
		err := repo.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
		assert.NoError(t, err)

		// 检查存在
		exists, err := repo.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// 下载
		reader, err := repo.Download(ctx, key)
		assert.NoError(t, err)
		defer reader.Close()

		downloaded, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, content, string(downloaded))
	})

	t.Run("Should get object info", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		key := "users/123/2024/09/uuid-456/info-test.txt"
		content := "Info test content"

		err := repo.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
		require.NoError(t, err)

		info, err := repo.GetObjectInfo(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, key, info.Key)
		assert.Equal(t, int64(len(content)), info.Size)
		assert.Equal(t, "text/plain", info.ContentType)
	})

	t.Run("Should generate presigned URL", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		key := "users/123/2024/09/uuid-789/presign-test.txt"
		content := "Presign test"

		err := repo.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
		require.NoError(t, err)

		url, err := repo.GeneratePresignedURL(ctx, key, time.Hour)
		assert.NoError(t, err)
		assert.Contains(t, url, key)
		assert.Contains(t, url, "expires=3600")
	})

	t.Run("Should handle multipart upload", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		key := "users/123/2024/09/uuid-multipart/large-file.txt"

		parts := []io.Reader{
			strings.NewReader("Part 1 content "),
			strings.NewReader("Part 2 content "),
			strings.NewReader("Part 3 content"),
		}

		err := repo.UploadMultipart(ctx, key, parts, "text/plain")
		assert.NoError(t, err)

		// 验证合并后的内容
		reader, err := repo.Download(ctx, key)
		assert.NoError(t, err)
		defer reader.Close()

		content, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, "Part 1 content Part 2 content Part 3 content", string(content))
	})

	t.Run("Should list objects with prefix", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		// 上传多个文件
		files := map[string]string{
			"users/123/2024/09/file1.txt": "content1",
			"users/123/2024/09/file2.txt": "content2",
			"users/456/2024/09/file3.txt": "content3",
		}

		for key, content := range files {
			err := repo.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
			require.NoError(t, err)
		}

		// 列出用户123的文件
		objects, err := repo.ListObjects(ctx, "users/123/", 10)
		assert.NoError(t, err)
		assert.Len(t, objects, 2)

		// 列出所有文件
		allObjects, err := repo.ListObjects(ctx, "users/", 10)
		assert.NoError(t, err)
		assert.Len(t, allObjects, 3)
	})

	t.Run("Should delete files", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		key := "users/123/2024/09/uuid-delete/delete-test.txt"
		content := "To be deleted"

		// 上传文件
		err := repo.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
		require.NoError(t, err)

		// 确认存在
		exists, err := repo.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// 删除文件
		err = repo.Delete(ctx, key)
		assert.NoError(t, err)

		// 确认已删除
		exists, err = repo.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Should validate object keys", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		// 测试路径遍历攻击
		err := repo.Upload(ctx, "../../etc/passwd", strings.NewReader("hack"), 4, "text/plain")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal detected")

		// 测试过长路径
		longKey := strings.Repeat("a", 1025)
		err = repo.Upload(ctx, longKey, strings.NewReader("test"), 4, "text/plain")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too long")
	})

	t.Run("Should handle errors gracefully", func(t *testing.T) {
		repo := NewMockMinIOStorageRepository()
		// 测试上传失败
		repo.SetFailure("upload")
		err := repo.Upload(ctx, "test-key", strings.NewReader("test"), 4, "text/plain")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock upload failure")

		// 重置失败状态
		repo.SetFailure("")

		// 测试下载不存在的文件
		_, err = repo.Download(ctx, "non-existent-key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}
