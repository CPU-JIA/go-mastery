package repository

import (
	"context"
	"file-storage-service/internal/model"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移所有表
	err = db.AutoMigrate(
		&model.User{},
		&model.File{},
		&model.ImageInfo{},
		&model.ThumbnailInfo{},
		&model.AccessLog{},
		&model.UploadToken{},
		&model.FileShare{},
		&model.Folder{},
		&model.FileFolder{},
		&model.UserSettings{},
	)
	require.NoError(t, err)

	return db
}

func TestFileRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFileRepository(db)
	ctx := context.Background()

	// 创建测试用户
	user := &model.User{
		UUID:     "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Role:     model.UserRoleCustomer,
		Status:   model.UserStatusActive,
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("Should create file", func(t *testing.T) {
		file := &model.File{
			UUID:         "file-123",
			OriginalName: "test.jpg",
			StorageName:  "storage_test_123.jpg",
			Size:         1024,
			MimeType:     "image/jpeg",
			UserID:       user.ID,
			Status:       model.FileStatusActive,
			UploadedAt:   time.Now(),
		}

		err := repo.Create(ctx, file)
		assert.NoError(t, err)
		assert.Greater(t, file.ID, uint(0))
	})

	t.Run("Should get file by ID", func(t *testing.T) {
		file := &model.File{
			UUID:         "file-456",
			OriginalName: "test2.png",
			StorageName:  "storage_test_456.png",
			Size:         2048,
			MimeType:     "image/png",
			UserID:       user.ID,
			Status:       model.FileStatusActive,
			UploadedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, file))

		retrieved, err := repo.GetByID(ctx, file.ID)
		assert.NoError(t, err)
		assert.Equal(t, file.UUID, retrieved.UUID)
		assert.Equal(t, file.OriginalName, retrieved.OriginalName)
	})

	t.Run("Should get files by user ID with pagination", func(t *testing.T) {
		// 创建多个文件
		for i := 0; i < 5; i++ {
			file := &model.File{
				UUID:         fmt.Sprintf("file-user-%d", i),
				OriginalName: fmt.Sprintf("user-file-%d.txt", i),
				StorageName:  fmt.Sprintf("storage_user_%d.txt", i),
				Size:         int64(100 * (i + 1)),
				MimeType:     "text/plain",
				UserID:       user.ID,
				Status:       model.FileStatusActive,
				UploadedAt:   time.Now(),
			}
			require.NoError(t, repo.Create(ctx, file))
		}

		files, total, err := repo.GetByUserID(ctx, user.ID, 3, 0)
		assert.NoError(t, err)
		assert.Len(t, files, 3)
		assert.GreaterOrEqual(t, total, int64(5))
	})

	t.Run("Should update file status", func(t *testing.T) {
		file := &model.File{
			UUID:         "file-status-test",
			OriginalName: "status-test.pdf",
			StorageName:  "storage_status_test.pdf",
			Size:         4096,
			MimeType:     "application/pdf",
			UserID:       user.ID,
			Status:       model.FileStatusActive,
			UploadedAt:   time.Now(),
		}
		require.NoError(t, repo.Create(ctx, file))

		err := repo.UpdateStatus(ctx, file.ID, model.FileStatusArchived)
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, file.ID)
		assert.NoError(t, err)
		assert.Equal(t, model.FileStatusArchived, updated.Status)
	})
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("Should create user", func(t *testing.T) {
		user := &model.User{
			UUID:      "create-user-123",
			Username:  "createuser",
			Email:     "create@example.com",
			Password:  "password123",
			FirstName: "Create",
			LastName:  "User",
			Role:      model.UserRoleCustomer,
			Status:    model.UserStatusActive,
		}

		err := repo.Create(ctx, user)
		assert.NoError(t, err)
		assert.Greater(t, user.ID, uint(0))
	})

	t.Run("Should check storage quota", func(t *testing.T) {
		user := &model.User{
			UUID:         "quota-user-123",
			Username:     "quotauser",
			Email:        "quota@example.com",
			Password:     "password123",
			StorageQuota: 1024 * 1024, // 1MB
			StorageUsed:  512 * 1024,  // 512KB
		}
		require.NoError(t, repo.Create(ctx, user))

		// 应该允许上传300KB
		canUpload, err := repo.CheckStorageQuota(ctx, user.ID, 300*1024)
		assert.NoError(t, err)
		assert.True(t, canUpload)

		// 不应该允许上传600KB
		canUpload, err = repo.CheckStorageQuota(ctx, user.ID, 600*1024)
		assert.NoError(t, err)
		assert.False(t, canUpload)
	})

	t.Run("Should update storage used", func(t *testing.T) {
		user := &model.User{
			UUID:        "storage-user-123",
			Username:    "storageuser",
			Email:       "storage@example.com",
			Password:    "password123",
			StorageUsed: 1024,
		}
		require.NoError(t, repo.Create(ctx, user))

		err := repo.UpdateStorageUsed(ctx, user.ID, 2048)
		assert.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(3072), updated.StorageUsed) // 1024 + 2048
	})
}

func TestAccessLogRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAccessLogRepository(db)
	ctx := context.Background()

	// 创建测试数据
	user := &model.User{
		UUID:     "log-user-123",
		Username: "loguser",
		Email:    "log@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	file := &model.File{
		UUID:         "log-file-123",
		OriginalName: "log-test.txt",
		StorageName:  "storage_log_123.txt",
		Size:         512,
		MimeType:     "text/plain",
		UserID:       user.ID,
		Status:       model.FileStatusActive,
		UploadedAt:   time.Now(),
	}
	require.NoError(t, db.Create(file).Error)

	t.Run("Should create access log", func(t *testing.T) {
		log := &model.AccessLog{
			FileID:     file.ID,
			UserID:     user.ID,
			Action:     "download",
			IPAddress:  "192.168.1.1",
			UserAgent:  "TestAgent/1.0",
			AccessedAt: time.Now(),
		}

		err := repo.Create(ctx, log)
		assert.NoError(t, err)
		assert.Greater(t, log.ID, uint(0))
	})

	t.Run("Should get logs by file ID", func(t *testing.T) {
		// 创建多个日志
		for i := 0; i < 3; i++ {
			log := &model.AccessLog{
				FileID:     file.ID,
				UserID:     user.ID,
				Action:     fmt.Sprintf("action-%d", i),
				IPAddress:  "192.168.1.1",
				UserAgent:  "TestAgent/1.0",
				AccessedAt: time.Now(),
			}
			require.NoError(t, repo.Create(ctx, log))
		}

		logs, total, err := repo.GetByFileID(ctx, file.ID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 3)
		assert.GreaterOrEqual(t, total, int64(3))
	})
}

func TestUploadTokenRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUploadTokenRepository(db)
	ctx := context.Background()

	user := &model.User{
		UUID:     "token-user-123",
		Username: "tokenuser",
		Email:    "token@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("Should create upload token", func(t *testing.T) {
		token := &model.UploadToken{
			Token:          "upload-token-123",
			UserID:         user.ID,
			OriginalName:   "largefile.zip",
			TotalSize:      10485760, // 10MB
			ChunkSize:      1048576,  // 1MB
			TotalChunks:    10,
			UploadedChunks: []int{},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}

		err := repo.Create(ctx, token)
		assert.NoError(t, err)
		assert.Greater(t, token.ID, uint(0))
	})

	t.Run("Should get token by token string", func(t *testing.T) {
		tokenStr := "get-token-456"
		token := &model.UploadToken{
			Token:        tokenStr,
			UserID:       user.ID,
			OriginalName: "getfile.zip",
			TotalSize:    5242880, // 5MB
			ChunkSize:    1048576, // 1MB
			TotalChunks:  5,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, repo.Create(ctx, token))

		retrieved, err := repo.GetByToken(ctx, tokenStr)
		assert.NoError(t, err)
		assert.Equal(t, tokenStr, retrieved.Token)
		assert.Equal(t, "getfile.zip", retrieved.OriginalName)
	})

	t.Run("Should update uploaded chunks", func(t *testing.T) {
		tokenStr := "chunk-token-789"
		token := &model.UploadToken{
			Token:          tokenStr,
			UserID:         user.ID,
			OriginalName:   "chunkfile.zip",
			TotalSize:      3145728, // 3MB
			ChunkSize:      1048576, // 1MB
			TotalChunks:    3,
			UploadedChunks: []int{0},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}
		require.NoError(t, repo.Create(ctx, token))

		err := repo.UpdateUploadedChunks(ctx, tokenStr, 1)
		assert.NoError(t, err)

		updated, err := repo.GetByToken(ctx, tokenStr)
		assert.NoError(t, err)
		assert.Contains(t, updated.UploadedChunks, 1)
	})
}

// MockStorageRepository 存储仓储的Mock实现，用于测试
type MockStorageRepository struct {
	objects map[string][]byte
}

func NewMockStorageRepository() *MockStorageRepository {
	return &MockStorageRepository{
		objects: make(map[string][]byte),
	}
}

func (m *MockStorageRepository) Upload(ctx context.Context, key string, data io.Reader, size int64, contentType string) error {
	content, err := io.ReadAll(data)
	if err != nil {
		return err
	}
	m.objects[key] = content
	return nil
}

func (m *MockStorageRepository) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	content, exists := m.objects[key]
	if !exists {
		return nil, fmt.Errorf("object not found")
	}
	return io.NopCloser(strings.NewReader(string(content))), nil
}

func (m *MockStorageRepository) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.objects[key]
	return exists, nil
}

func (m *MockStorageRepository) Delete(ctx context.Context, key string) error {
	delete(m.objects, key)
	return nil
}

func (m *MockStorageRepository) GetObjectInfo(ctx context.Context, key string) (*ObjectInfo, error) {
	content, exists := m.objects[key]
	if !exists {
		return nil, fmt.Errorf("object not found")
	}
	return &ObjectInfo{
		Key:  key,
		Size: int64(len(content)),
	}, nil
}

func (m *MockStorageRepository) GeneratePresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("https://example.com/presigned/%s", key), nil
}

func (m *MockStorageRepository) UploadMultipart(ctx context.Context, key string, parts []io.Reader, contentType string) error {
	var allContent []byte
	for _, part := range parts {
		content, err := io.ReadAll(part)
		if err != nil {
			return err
		}
		allContent = append(allContent, content...)
	}
	m.objects[key] = allContent
	return nil
}

func (m *MockStorageRepository) ListObjects(ctx context.Context, prefix string, limit int) ([]string, error) {
	var keys []string
	count := 0
	for key := range m.objects {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
			count++
			if count >= limit {
				break
			}
		}
	}
	return keys, nil
}

func TestMockStorageRepository(t *testing.T) {
	repo := NewMockStorageRepository()
	ctx := context.Background()

	t.Run("Should upload and download object", func(t *testing.T) {
		key := "test/file.txt"
		content := "Hello, World!"

		err := repo.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
		assert.NoError(t, err)

		exists, err := repo.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		reader, err := repo.Download(ctx, key)
		assert.NoError(t, err)
		defer reader.Close()

		downloaded, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, content, string(downloaded))
	})
}
