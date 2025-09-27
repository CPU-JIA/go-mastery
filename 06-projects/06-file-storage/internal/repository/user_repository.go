package repository

import (
	"context"
	"file-storage-service/internal/model"
	"time"

	"gorm.io/gorm"
)

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUUID(ctx context.Context, uuid string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *userRepository) UpdateStorageUsed(ctx context.Context, userID uint, sizeDelta int64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("storage_used", gorm.Expr("storage_used + ?", sizeDelta)).Error
}

func (r *userRepository) CheckStorageQuota(ctx context.Context, userID uint, requiredSize int64) (bool, error) {
	var user model.User
	err := r.db.WithContext(ctx).Select("storage_quota", "storage_used").First(&user, userID).Error
	if err != nil {
		return false, err
	}

	// 如果配额为0表示无限制
	if user.StorageQuota == 0 {
		return true, nil
	}

	return user.StorageUsed+requiredSize <= user.StorageQuota, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// 获取总数
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) UpdateLastLoginAt(ctx context.Context, userID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_login_at", &now).Error
}
