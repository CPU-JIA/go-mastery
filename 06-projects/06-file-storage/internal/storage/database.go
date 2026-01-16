package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"file-storage-service/internal/config"
	"file-storage-service/internal/models"
)

// GormStorage GORM数据库存储实现
type GormStorage struct {
	db *gorm.DB
}

// NewGormStorage 创建GORM存储实例
func NewGormStorage(config config.DatabaseConfig) (DatabaseStorage, error) {
	var db *gorm.DB
	var err error

	switch config.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.File{},
		&models.FileVersion{},
		&models.ImageInfo{},
		&models.Thumbnail{},
		&models.UploadToken{},
		&models.AccessLog{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &GormStorage{db: db}, nil
}

// CreateFile 创建文件记录
func (gs *GormStorage) CreateFile(ctx context.Context, file *models.File) error {
	return gs.db.WithContext(ctx).Create(file).Error
}

// GetFileByID 根据ID获取文件
func (gs *GormStorage) GetFileByID(ctx context.Context, id string) (*models.File, error) {
	var file models.File
	err := gs.db.WithContext(ctx).
		Preload("ImageInfo").
		Preload("ImageInfo.Thumbnails").
		Preload("Versions").
		First(&file, "id = ?", id).Error

	if err != nil {
		return nil, err
	}

	return &file, nil
}

// GetFilesByOwner 获取用户的文件列表
func (gs *GormStorage) GetFilesByOwner(ctx context.Context, ownerID string, page, perPage int) ([]*models.File, int64, error) {
	var files []*models.File
	var total int64

	query := gs.db.WithContext(ctx).Model(&models.File{})

	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * perPage
	err := query.
		Preload("ImageInfo").
		Preload("ImageInfo.Thumbnails").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&files).Error

	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// UpdateFile 更新文件
func (gs *GormStorage) UpdateFile(ctx context.Context, file *models.File) error {
	return gs.db.WithContext(ctx).Save(file).Error
}

// DeleteFile 删除文件
func (gs *GormStorage) DeleteFile(ctx context.Context, id string) error {
	return gs.db.WithContext(ctx).Delete(&models.File{}, "id = ?", id).Error
}

// SearchFiles 搜索文件
func (gs *GormStorage) SearchFiles(ctx context.Context, query string, filters map[string]interface{}, page, perPage int) ([]*models.File, int64, error) {
	var files []*models.File
	var total int64

	dbQuery := gs.db.WithContext(ctx).Model(&models.File{})

	// 文本搜索
	if query != "" {
		searchTerm := "%" + strings.ToLower(query) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(original_name) LIKE ?", searchTerm, searchTerm)
	}

	// 过滤器
	for key, value := range filters {
		switch key {
		case "mime_type":
			dbQuery = dbQuery.Where("mime_type LIKE ?", "%"+value.(string)+"%")
		case "owner_id":
			dbQuery = dbQuery.Where("owner_id = ?", value)
		case "visibility":
			dbQuery = dbQuery.Where("visibility = ?", value)
		case "min_size":
			dbQuery = dbQuery.Where("size >= ?", value)
		case "max_size":
			dbQuery = dbQuery.Where("size <= ?", value)
		}
	}

	// 获取总数
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * perPage
	err := dbQuery.
		Preload("ImageInfo").
		Preload("ImageInfo.Thumbnails").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&files).Error

	return files, total, err
}

// CreateFileVersion 创建文件版本
func (gs *GormStorage) CreateFileVersion(ctx context.Context, version *models.FileVersion) error {
	return gs.db.WithContext(ctx).Create(version).Error
}

// GetFileVersions 获取文件版本列表
func (gs *GormStorage) GetFileVersions(ctx context.Context, fileID string) ([]*models.FileVersion, error) {
	var versions []*models.FileVersion
	err := gs.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Order("version DESC").
		Find(&versions).Error

	return versions, err
}

// CreateImageInfo 创建图片信息
func (gs *GormStorage) CreateImageInfo(ctx context.Context, imageInfo *models.ImageInfo) error {
	return gs.db.WithContext(ctx).Create(imageInfo).Error
}

// UpdateImageInfo 更新图片信息
func (gs *GormStorage) UpdateImageInfo(ctx context.Context, imageInfo *models.ImageInfo) error {
	return gs.db.WithContext(ctx).Save(imageInfo).Error
}

// CreateThumbnail 创建缩略图
func (gs *GormStorage) CreateThumbnail(ctx context.Context, thumbnail *models.Thumbnail) error {
	return gs.db.WithContext(ctx).Create(thumbnail).Error
}

// CreateUploadToken 创建上传令牌
func (gs *GormStorage) CreateUploadToken(ctx context.Context, token *models.UploadToken) error {
	return gs.db.WithContext(ctx).Create(token).Error
}

// GetUploadToken 获取上传令牌
func (gs *GormStorage) GetUploadToken(ctx context.Context, token string) (*models.UploadToken, error) {
	var uploadToken models.UploadToken
	err := gs.db.WithContext(ctx).
		Where("token = ? AND expires_at > ?", token, time.Now()).
		First(&uploadToken).Error

	if err != nil {
		return nil, err
	}

	return &uploadToken, nil
}

// UpdateUploadToken 更新上传令牌
func (gs *GormStorage) UpdateUploadToken(ctx context.Context, token *models.UploadToken) error {
	return gs.db.WithContext(ctx).Save(token).Error
}

// DeleteUploadToken 删除上传令牌
func (gs *GormStorage) DeleteUploadToken(ctx context.Context, token string) error {
	return gs.db.WithContext(ctx).Delete(&models.UploadToken{}, "token = ?", token).Error
}

// CleanupExpiredTokens 清理过期令牌
func (gs *GormStorage) CleanupExpiredTokens(ctx context.Context) error {
	return gs.db.WithContext(ctx).
		Delete(&models.UploadToken{}, "expires_at < ?", time.Now()).Error
}

// CreateAccessLog 创建访问日志
func (gs *GormStorage) CreateAccessLog(ctx context.Context, log *models.AccessLog) error {
	return gs.db.WithContext(ctx).Create(log).Error
}

// GetAccessLogs 获取访问日志
func (gs *GormStorage) GetAccessLogs(ctx context.Context, fileID string, limit int) ([]*models.AccessLog, error) {
	var logs []*models.AccessLog
	query := gs.db.WithContext(ctx).Model(&models.AccessLog{})

	if fileID != "" {
		query = query.Where("file_id = ?", fileID)
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error

	return logs, err
}

// GetFileStats 获取文件统计
func (gs *GormStorage) GetFileStats(ctx context.Context) (*models.StatsResponse, error) {
	stats := &models.StatsResponse{
		FilesByType:  make(map[string]int64),
		FilesByOwner: make(map[string]int64),
	}

	// 总文件数和大小
	gs.db.WithContext(ctx).Model(&models.File{}).Count(&stats.TotalFiles)
	gs.db.WithContext(ctx).Model(&models.File{}).Select("COALESCE(SUM(size), 0)").Scan(&stats.TotalSize)

	// 按类型统计
	var typeStats []struct {
		MimeType string
		Count    int64
	}
	gs.db.WithContext(ctx).Model(&models.File{}).
		Select("SUBSTRING(mime_type, 1, POSITION('/' IN mime_type || '/') - 1) as mime_type, COUNT(*) as count").
		Group("SUBSTRING(mime_type, 1, POSITION('/' IN mime_type || '/') - 1)").
		Scan(&typeStats)

	for _, stat := range typeStats {
		if stat.MimeType != "" {
			stats.FilesByType[stat.MimeType] = stat.Count
		}
	}

	// 按所有者统计
	var ownerStats []struct {
		OwnerID string
		Count   int64
	}
	gs.db.WithContext(ctx).Model(&models.File{}).
		Select("owner_id, COUNT(*) as count").
		Group("owner_id").
		Scan(&ownerStats)

	for _, stat := range ownerStats {
		stats.FilesByOwner[stat.OwnerID] = stat.Count
	}

	// 最近上传数量（7天内）
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	gs.db.WithContext(ctx).Model(&models.File{}).
		Where("created_at > ?", oneWeekAgo).
		Count(&stats.RecentUploads)

	return stats, nil
}

// GetStorageUsage 获取存储使用量
func (gs *GormStorage) GetStorageUsage(ctx context.Context) (int64, error) {
	var totalSize int64
	err := gs.db.WithContext(ctx).Model(&models.File{}).
		Select("COALESCE(SUM(size), 0)").
		Scan(&totalSize).Error

	return totalSize, err
}

// HealthCheck 健康检查
func (gs *GormStorage) HealthCheck(ctx context.Context) error {
	sqlDB, err := gs.db.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}
