package repository

import (
	"context"
	"file-storage-service/internal/model"
	"time"

	"gorm.io/gorm"
)

// accessLogRepository 访问日志仓储实现
type accessLogRepository struct {
	db *gorm.DB
}

// NewAccessLogRepository 创建访问日志仓储
func NewAccessLogRepository(db *gorm.DB) AccessLogRepository {
	return &accessLogRepository{db: db}
}

func (r *accessLogRepository) Create(ctx context.Context, log *model.AccessLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *accessLogRepository) GetByFileID(ctx context.Context, fileID uint, limit, offset int) ([]*model.AccessLog, int64, error) {
	var logs []*model.AccessLog
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.AccessLog{}).Where("file_id = ?", fileID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Where("file_id = ?", fileID).
		Preload("User").
		Preload("File").
		Limit(limit).Offset(offset).
		Order("accessed_at DESC").
		Find(&logs).Error

	return logs, total, err
}

func (r *accessLogRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.AccessLog, int64, error) {
	var logs []*model.AccessLog
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.AccessLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Preload("User").
		Preload("File").
		Limit(limit).Offset(offset).
		Order("accessed_at DESC").
		Find(&logs).Error

	return logs, total, err
}

func (r *accessLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*model.AccessLog, int64, error) {
	var logs []*model.AccessLog
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.AccessLog{}).Where("action = ?", action).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Where("action = ?", action).
		Preload("User").
		Preload("File").
		Limit(limit).Offset(offset).
		Order("accessed_at DESC").
		Find(&logs).Error

	return logs, total, err
}

func (r *accessLogRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*model.AccessLog, int64, error) {
	var logs []*model.AccessLog
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.AccessLog{}).
		Where("accessed_at BETWEEN ? AND ?", startDate, endDate).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).
		Where("accessed_at BETWEEN ? AND ?", startDate, endDate).
		Preload("User").
		Preload("File").
		Limit(limit).Offset(offset).
		Order("accessed_at DESC").
		Find(&logs).Error

	return logs, total, err
}

func (r *accessLogRepository) DeleteOldLogs(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("accessed_at < ?", before).Delete(&model.AccessLog{}).Error
}
