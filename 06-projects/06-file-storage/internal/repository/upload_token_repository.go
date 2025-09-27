package repository

import (
	"context"
	"file-storage-service/internal/model"
	"time"

	"gorm.io/gorm"
)

// uploadTokenRepository 上传令牌仓储实现
type uploadTokenRepository struct {
	db *gorm.DB
}

// NewUploadTokenRepository 创建上传令牌仓储
func NewUploadTokenRepository(db *gorm.DB) UploadTokenRepository {
	return &uploadTokenRepository{db: db}
}

func (r *uploadTokenRepository) Create(ctx context.Context, token *model.UploadToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *uploadTokenRepository) GetByToken(ctx context.Context, token string) (*model.UploadToken, error) {
	var uploadToken model.UploadToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&uploadToken).Error
	if err != nil {
		return nil, err
	}
	return &uploadToken, nil
}

func (r *uploadTokenRepository) Update(ctx context.Context, uploadToken *model.UploadToken) error {
	return r.db.WithContext(ctx).Save(uploadToken).Error
}

func (r *uploadTokenRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&model.UploadToken{}).Error
}

func (r *uploadTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&model.UploadToken{}).Error
}

func (r *uploadTokenRepository) UpdateUploadedChunks(ctx context.Context, token string, chunkIndex int) error {
	var uploadToken model.UploadToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&uploadToken).Error
	if err != nil {
		return err
	}

	// 检查chunk是否已存在
	for _, chunk := range uploadToken.UploadedChunks {
		if chunk == chunkIndex {
			return nil // chunk已存在，不需要更新
		}
	}

	// 添加新的chunk
	uploadToken.UploadedChunks = append(uploadToken.UploadedChunks, chunkIndex)

	return r.db.WithContext(ctx).Save(&uploadToken).Error
}
