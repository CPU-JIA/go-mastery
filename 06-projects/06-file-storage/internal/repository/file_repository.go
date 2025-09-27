package repository

import (
	"context"
	"file-storage-service/internal/model"
	"strings"

	"gorm.io/gorm"
)

// fileRepository 文件仓储实现
type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建文件仓储
func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepository) GetByID(ctx context.Context, id uint) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetByUUID(ctx context.Context, uuid string) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.File{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&files).Error

	return files, total, err
}

func (r *fileRepository) Update(ctx context.Context, file *model.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.File{}, id).Error
}

func (r *fileRepository) GetWithDetails(ctx context.Context, id uint) (*model.File, error) {
	var file model.File
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("ImageInfo").
		Preload("ThumbnailInfo").
		First(&file, id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) Search(ctx context.Context, userID uint, query string, limit, offset int) ([]*model.File, int64, error) {
	var files []*model.File
	var total int64

	queryBuilder := r.db.WithContext(ctx).Model(&model.File{}).Where("user_id = ?", userID)

	if query != "" {
		searchQuery := "%" + strings.ToLower(query) + "%"
		queryBuilder = queryBuilder.Where("LOWER(original_name) LIKE ? OR LOWER(mime_type) LIKE ?", searchQuery, searchQuery)
	}

	// 获取总数
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := queryBuilder.Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&files).Error

	return files, total, err
}

func (r *fileRepository) GetByStatus(ctx context.Context, status model.FileStatus, limit, offset int) ([]*model.File, error) {
	var files []*model.File
	err := r.db.WithContext(ctx).Where("status = ?", status).
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

func (r *fileRepository) UpdateStatus(ctx context.Context, id uint, status model.FileStatus) error {
	return r.db.WithContext(ctx).Model(&model.File{}).Where("id = ?", id).Update("status", status).Error
}

func (r *fileRepository) GetUserStorageStats(ctx context.Context, userID uint) (*model.StorageStats, error) {
	var stats model.StorageStats
	stats.UserID = userID

	// 获取文件总数和总大小
	var totalSize int64
	err := r.db.WithContext(ctx).Model(&model.File{}).
		Where("user_id = ? AND status = ?", userID, model.FileStatusActive).
		Select("COUNT(*) as count, COALESCE(SUM(size), 0) as total_size").
		Row().Scan(&stats.TotalFiles, &totalSize)

	if err != nil {
		return nil, err
	}

	stats.TotalSize = totalSize

	// 计算用户配额使用率
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return nil, err
	}

	if user.StorageQuota > 0 {
		stats.UsedQuotaPercentage = float64(totalSize) / float64(user.StorageQuota) * 100
	}

	// 按文件类型统计
	stats.FilesByType = make(map[string]int64)
	rows, err := r.db.WithContext(ctx).Model(&model.File{}).
		Where("user_id = ? AND status = ?", userID, model.FileStatusActive).
		Select("mime_type, COUNT(*) as count").
		Group("mime_type").Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mimeType string
		var count int64
		if err := rows.Scan(&mimeType, &count); err != nil {
			continue
		}
		// 将mime type转为更友好的类型名
		fileType := getFileTypeFromMimeType(mimeType)
		stats.FilesByType[fileType] = count
	}

	// 获取最近上传的文件
	err = r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, model.FileStatusActive).
		Limit(5).Order("created_at DESC").Find(&stats.RecentFiles).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// getFileTypeFromMimeType 从MIME类型获取友好的文件类型名称
func getFileTypeFromMimeType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "images"
	case strings.HasPrefix(mimeType, "video/"):
		return "videos"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "text/"):
		return "documents"
	case strings.Contains(mimeType, "pdf"):
		return "documents"
	case strings.Contains(mimeType, "zip") || strings.Contains(mimeType, "archive"):
		return "archives"
	default:
		return "other"
	}
}
