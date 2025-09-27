package repository

import (
	"blog-system/internal/model"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Tags").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Tags").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Tags").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *userRepository) List(params PaginationParams) ([]model.User, *PaginationResult, error) {
	var users []model.User
	var total int64

	// 计算总数
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询用户列表
	err := r.db.Preload("Tags").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return users, result, nil
}

func (r *userRepository) SearchByKeyword(keyword string, params PaginationParams) ([]model.User, *PaginationResult, error) {
	var users []model.User
	var total int64

	searchTerm := fmt.Sprintf("%%%s%%", keyword)
	query := r.db.Model(&model.User{}).Where(
		"username LIKE ? OR email LIKE ? OR full_name LIKE ? OR bio LIKE ?",
		searchTerm, searchTerm, searchTerm, searchTerm,
	)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询用户列表
	err := query.Preload("Tags").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&users).Error

	if err != nil {
		return nil, nil, err
	}

	// 计算总页数
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	result := &PaginationResult{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}

	return users, result, nil
}

func (r *userRepository) GetActiveUsers() ([]model.User, error) {
	var users []model.User
	err := r.db.Where("status = ?", model.StatusActive).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateLastLogin(id uint) error {
	now := time.Now()
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Update("last_login", &now).Error
}
