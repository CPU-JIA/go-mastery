package model

import (
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
		&User{},
		&File{},
		&ImageInfo{},
		&ThumbnailInfo{},
		&AccessLog{},
		&UploadToken{},
		&FileShare{},
		&Folder{},
		&FileFolder{},
		&UserSettings{},
	)
	require.NoError(t, err)

	return db
}

func TestUserModel(t *testing.T) {
	db := setupTestDB(t)

	t.Run("Should create user with valid data", func(t *testing.T) {
		user := &User{
			UUID:      "user-123",
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "hashedpassword",
			FirstName: "Test",
			LastName:  "User",
			Role:      UserRoleCustomer,
			Status:    UserStatusActive,
		}

		result := db.Create(user)
		assert.NoError(t, result.Error)
		assert.Greater(t, user.ID, uint(0))
		assert.Equal(t, int64(1073741824), user.StorageQuota) // 默认1GB
		assert.True(t, user.CanUpload)
	})

	t.Run("Should enforce unique constraints", func(t *testing.T) {
		// 第一个用户
		user1 := &User{
			UUID:     "user-456",
			Username: "uniqueuser",
			Email:    "unique@example.com",
			Password: "password",
		}
		assert.NoError(t, db.Create(user1).Error)

		// 尝试创建具有相同username的用户
		user2 := &User{
			UUID:     "user-789",
			Username: "uniqueuser", // 重复username
			Email:    "different@example.com",
			Password: "password",
		}
		assert.Error(t, db.Create(user2).Error)

		// 尝试创建具有相同email的用户
		user3 := &User{
			UUID:     "user-101",
			Username: "differentuser",
			Email:    "unique@example.com", // 重复email
			Password: "password",
		}
		assert.Error(t, db.Create(user3).Error)
	})
}

func TestFileModel(t *testing.T) {
	db := setupTestDB(t)

	// 先创建用户
	user := &User{
		UUID:     "file-user-123",
		Username: "fileuser",
		Email:    "fileuser@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("Should create file with valid data", func(t *testing.T) {
		file := &File{
			UUID:         "file-123",
			OriginalName: "test.jpg",
			StorageName:  "storage_file_123.jpg",
			Size:         1024,
			MimeType:     "image/jpeg",
			IsEncrypted:  false,
			ChecksumMD5:  "d41d8cd98f00b204e9800998ecf8427e",
			UserID:       user.ID,
			Status:       FileStatusActive,
			UploadedAt:   time.Now(),
		}

		result := db.Create(file)
		assert.NoError(t, result.Error)
		assert.Greater(t, file.ID, uint(0))
	})

	t.Run("Should create file with image info", func(t *testing.T) {
		file := &File{
			UUID:         "image-file-456",
			OriginalName: "image.png",
			StorageName:  "storage_image_456.png",
			Size:         2048,
			MimeType:     "image/png",
			UserID:       user.ID,
			Status:       FileStatusActive,
			UploadedAt:   time.Now(),
		}
		require.NoError(t, db.Create(file).Error)

		imageInfo := &ImageInfo{
			FileID:    file.ID,
			Width:     800,
			Height:    600,
			Format:    "PNG",
			ColorMode: "RGBA",
			HasAlpha:  true,
			DPI:       72,
		}
		assert.NoError(t, db.Create(imageInfo).Error)

		// 验证关联
		var fileWithImage File
		err := db.Preload("ImageInfo").First(&fileWithImage, file.ID).Error
		assert.NoError(t, err)
		assert.NotNil(t, fileWithImage.ImageInfo)
		assert.Equal(t, 800, fileWithImage.ImageInfo.Width)
	})
}

func TestAccessLogModel(t *testing.T) {
	db := setupTestDB(t)

	// 创建用户和文件
	user := &User{
		UUID:     "log-user-123",
		Username: "loguser",
		Email:    "loguser@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	file := &File{
		UUID:         "log-file-123",
		OriginalName: "log-test.txt",
		StorageName:  "storage_log_123.txt",
		Size:         512,
		MimeType:     "text/plain",
		UserID:       user.ID,
		Status:       FileStatusActive,
		UploadedAt:   time.Now(),
	}
	require.NoError(t, db.Create(file).Error)

	t.Run("Should create access log", func(t *testing.T) {
		log := &AccessLog{
			FileID:     file.ID,
			UserID:     user.ID,
			Action:     "download",
			IPAddress:  "192.168.1.1",
			UserAgent:  "TestAgent/1.0",
			AccessedAt: time.Now(),
		}

		result := db.Create(log)
		assert.NoError(t, result.Error)
		assert.Greater(t, log.ID, uint(0))
	})

	t.Run("Should preload associations", func(t *testing.T) {
		var logs []AccessLog
		err := db.Preload("File").Preload("User").Find(&logs).Error
		assert.NoError(t, err)
		if len(logs) > 0 {
			assert.Equal(t, file.OriginalName, logs[0].File.OriginalName)
			assert.Equal(t, user.Username, logs[0].User.Username)
		}
	})
}

func TestUploadTokenModel(t *testing.T) {
	db := setupTestDB(t)

	user := &User{
		UUID:     "token-user-123",
		Username: "tokenuser",
		Email:    "tokenuser@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("Should create upload token", func(t *testing.T) {
		token := &UploadToken{
			Token:          "upload-token-abc123",
			UserID:         user.ID,
			OriginalName:   "largefile.zip",
			TotalSize:      10485760, // 10MB
			ChunkSize:      1048576,  // 1MB
			TotalChunks:    10,
			UploadedChunks: []int{0, 1, 2},
			ExpiresAt:      time.Now().Add(24 * time.Hour),
		}

		result := db.Create(token)
		assert.NoError(t, result.Error)
		assert.Greater(t, token.ID, uint(0))

		// 验证JSON序列化
		var retrievedToken UploadToken
		err := db.First(&retrievedToken, token.ID).Error
		assert.NoError(t, err)
		assert.Len(t, retrievedToken.UploadedChunks, 3)
		assert.Equal(t, []int{0, 1, 2}, retrievedToken.UploadedChunks)
	})
}

func TestFileShareModel(t *testing.T) {
	db := setupTestDB(t)

	// 创建用户和文件
	user := &User{
		UUID:     "share-user-123",
		Username: "shareuser",
		Email:    "shareuser@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	file := &File{
		UUID:         "share-file-123",
		OriginalName: "shared-file.pdf",
		StorageName:  "storage_share_123.pdf",
		Size:         1024,
		MimeType:     "application/pdf",
		UserID:       user.ID,
		Status:       FileStatusActive,
		UploadedAt:   time.Now(),
	}
	require.NoError(t, db.Create(file).Error)

	t.Run("Should create file share", func(t *testing.T) {
		expiresAt := time.Now().Add(7 * 24 * time.Hour) // 7天
		share := &FileShare{
			FileID:       file.ID,
			ShareToken:   "share-token-xyz789",
			MaxDownloads: 10,
			ExpiresAt:    &expiresAt,
			IsActive:     true,
		}

		result := db.Create(share)
		assert.NoError(t, result.Error)
		assert.Greater(t, share.ID, uint(0))
	})
}

func TestFolderModel(t *testing.T) {
	db := setupTestDB(t)

	user := &User{
		UUID:     "folder-user-123",
		Username: "folderuser",
		Email:    "folderuser@example.com",
		Password: "password",
	}
	require.NoError(t, db.Create(user).Error)

	t.Run("Should create folder hierarchy", func(t *testing.T) {
		// 根文件夹
		rootFolder := &Folder{
			UUID:        "folder-root-123",
			Name:        "Documents",
			Path:        "/Documents",
			UserID:      user.ID,
			Description: "Root documents folder",
		}
		require.NoError(t, db.Create(rootFolder).Error)

		// 子文件夹
		subFolder := &Folder{
			UUID:     "folder-sub-456",
			Name:     "Images",
			Path:     "/Documents/Images",
			UserID:   user.ID,
			ParentID: &rootFolder.ID,
		}
		assert.NoError(t, db.Create(subFolder).Error)

		// 验证层级关系
		var parent Folder
		err := db.Preload("Children").First(&parent, rootFolder.ID).Error
		assert.NoError(t, err)
		assert.Len(t, parent.Children, 1)
		assert.Equal(t, "Images", parent.Children[0].Name)
	})
}
