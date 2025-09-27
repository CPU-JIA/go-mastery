package repository

import (
	"context"
	"file-storage-service/internal/model"

	"gorm.io/gorm"
)

// NewRepositories 创建所有仓储的工厂函数
func NewRepositories(db *gorm.DB, minioConfig *MinIOConfig) (*Repositories, error) {
	// 创建MinIO存储仓储
	storageRepo, err := NewMinIOStorageRepository(minioConfig)
	if err != nil {
		return nil, err
	}

	return &Repositories{
		File:        NewFileRepository(db),
		User:        NewUserRepository(db),
		Storage:     storageRepo,
		AccessLog:   NewAccessLogRepository(db),
		UploadToken: NewUploadTokenRepository(db),
		FileShare:   NewFileShareRepository(db),
		Folder:      NewFolderRepository(db),
	}, nil
}

// NewFileShareRepository 创建文件分享仓储
func NewFileShareRepository(db *gorm.DB) FileShareRepository {
	return &fileShareRepository{db: db}
}

// NewFolderRepository 创建文件夹仓储
func NewFolderRepository(db *gorm.DB) FolderRepository {
	return &folderRepository{db: db}
}

// fileShareRepository 文件分享仓储实现（存根）
type fileShareRepository struct {
	db *gorm.DB
}

func (r *fileShareRepository) Create(ctx context.Context, share *model.FileShare) error {
	return r.db.WithContext(ctx).Create(share).Error
}

func (r *fileShareRepository) GetByShareToken(ctx context.Context, token string) (*model.FileShare, error) {
	var share model.FileShare
	err := r.db.WithContext(ctx).Where("share_token = ?", token).First(&share).Error
	if err != nil {
		return nil, err
	}
	return &share, nil
}

func (r *fileShareRepository) GetByFileID(ctx context.Context, fileID uint) ([]*model.FileShare, error) {
	var shares []*model.FileShare
	err := r.db.WithContext(ctx).Where("file_id = ?", fileID).Find(&shares).Error
	return shares, err
}

func (r *fileShareRepository) Update(ctx context.Context, share *model.FileShare) error {
	return r.db.WithContext(ctx).Save(share).Error
}

func (r *fileShareRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.FileShare{}, id).Error
}

func (r *fileShareRepository) IncrementDownloadCount(ctx context.Context, shareToken string) error {
	return r.db.WithContext(ctx).Model(&model.FileShare{}).
		Where("share_token = ?", shareToken).
		Update("download_count", gorm.Expr("download_count + 1")).Error
}

func (r *fileShareRepository) GetActiveShares(ctx context.Context, limit, offset int) ([]*model.FileShare, int64, error) {
	var shares []*model.FileShare
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.FileShare{}).
		Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Where("is_active = ?", true).
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&shares).Error

	return shares, total, err
}

// folderRepository 文件夹仓储实现（存根）
type folderRepository struct {
	db *gorm.DB
}

func (r *folderRepository) Create(ctx context.Context, folder *model.Folder) error {
	return r.db.WithContext(ctx).Create(folder).Error
}

func (r *folderRepository) GetByID(ctx context.Context, id uint) (*model.Folder, error) {
	var folder model.Folder
	err := r.db.WithContext(ctx).First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *folderRepository) GetByUserID(ctx context.Context, userID uint) ([]*model.Folder, error) {
	var folders []*model.Folder
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Find(&folders).Error
	return folders, err
}

func (r *folderRepository) GetByParentID(ctx context.Context, parentID *uint) ([]*model.Folder, error) {
	var folders []*model.Folder
	query := r.db.WithContext(ctx)

	if parentID == nil {
		query = query.Where("parent_id IS NULL")
	} else {
		query = query.Where("parent_id = ?", *parentID)
	}

	err := query.Order("created_at DESC").Find(&folders).Error
	return folders, err
}

func (r *folderRepository) Update(ctx context.Context, folder *model.Folder) error {
	return r.db.WithContext(ctx).Save(folder).Error
}

func (r *folderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Folder{}, id).Error
}

func (r *folderRepository) GetWithChildren(ctx context.Context, id uint) (*model.Folder, error) {
	var folder model.Folder
	err := r.db.WithContext(ctx).Preload("Children").First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *folderRepository) Move(ctx context.Context, folderID uint, newParentID *uint) error {
	updates := map[string]interface{}{"parent_id": newParentID}
	return r.db.WithContext(ctx).Model(&model.Folder{}).
		Where("id = ?", folderID).Updates(updates).Error
}
