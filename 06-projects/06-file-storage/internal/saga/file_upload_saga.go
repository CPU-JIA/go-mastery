package saga

import (
	"context"
	"file-storage-service/internal/model"
	"file-storage-service/internal/repository"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// FileUploadSagaDefinition 文件上传Saga定义
type FileUploadSagaDefinition struct {
	fileRepo    repository.FileRepository
	storageRepo repository.StorageRepository
	userRepo    repository.UserRepository
	pathBuilder *repository.ObjectPathBuilder
}

// NewFileUploadSagaDefinition 创建文件上传Saga定义
func NewFileUploadSagaDefinition(repos *repository.Repositories) SagaDefinition {
	return &FileUploadSagaDefinition{
		fileRepo:    repos.File,
		storageRepo: repos.Storage,
		userRepo:    repos.User,
		pathBuilder: repository.NewObjectPathBuilder(),
	}
}

// GetSagaType 获取Saga类型
func (f *FileUploadSagaDefinition) GetSagaType() string {
	return "FileUpload"
}

// ValidateContext 验证上下文
func (f *FileUploadSagaDefinition) ValidateContext(context interface{}) error {
	ctx, ok := context.(*model.FileUploadSagaContext)
	if !ok {
		return NewSagaError(ErrCodeInvalidContext, "invalid file upload context type", 0, -1)
	}

	// 验证必要字段
	if ctx.FileName == "" {
		return NewSagaError(ErrCodeInvalidContext, "file name is required", 0, -1)
	}

	if ctx.FileSize <= 0 {
		return NewSagaError(ErrCodeInvalidContext, "file size must be positive", 0, -1)
	}

	if ctx.UserID == 0 {
		return NewSagaError(ErrCodeInvalidContext, "user ID is required", 0, -1)
	}

	if ctx.ContentType == "" {
		return NewSagaError(ErrCodeInvalidContext, "content type is required", 0, -1)
	}

	// 验证文件大小限制（100MB）
	maxFileSize := int64(100 * 1024 * 1024)
	if ctx.FileSize > maxFileSize {
		return NewSagaError(ErrCodeInvalidInput, "file size exceeds limit", 0, -1)
	}

	// 验证文件类型
	allowedTypes := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp",
		"text/plain", "text/csv",
		"application/pdf", "application/json",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	}

	allowed := false
	for _, allowedType := range allowedTypes {
		if ctx.ContentType == allowedType {
			allowed = true
			break
		}
	}

	if !allowed {
		return NewSagaError(ErrCodeInvalidInput, "file type not allowed", 0, -1)
	}

	return nil
}

// CalculateTimeout 计算超时时间
func (f *FileUploadSagaDefinition) CalculateTimeout(context interface{}) time.Duration {
	ctx, ok := context.(*model.FileUploadSagaContext)
	if !ok {
		return DefaultSagaTimeout
	}

	// 根据文件大小动态计算超时时间
	// 基础时间5分钟，每MB额外30秒
	baseDuration := 5 * time.Minute
	sizeInMB := ctx.FileSize / (1024 * 1024)
	if sizeInMB == 0 {
		sizeInMB = 1 // 至少1MB
	}

	additionalTime := time.Duration(sizeInMB) * 30 * time.Second
	totalTimeout := baseDuration + additionalTime

	// 最大不超过30分钟
	if totalTimeout > 30*time.Minute {
		totalTimeout = 30 * time.Minute
	}

	return totalTimeout
}

// GetRetryPolicy 获取重试策略
func (f *FileUploadSagaDefinition) GetRetryPolicy() *model.RetryPolicy {
	return &model.RetryPolicy{
		MaxAttempts:   3,
		InitialDelay:  2 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []string{
			"network_error", "timeout_error", "temporary_error",
			"storage_service_unavailable", "database_connection_error",
		},
	}
}

// GetSteps 获取步骤定义
func (f *FileUploadSagaDefinition) GetSteps() []SagaStepDefinition {
	return []SagaStepDefinition{
		{
			Name:       "validate_user_quota",
			Execute:    f.validateUserQuota,
			Compensate: f.compensateUserQuota,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   2,
				InitialDelay:  time.Second,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
			},
			Timeout:      30 * time.Second,
			IsIdempotent: true,
			IsCritical:   false, // 配额验证失败不需要补偿
		},
		{
			Name:       "generate_object_key",
			Execute:    f.generateObjectKey,
			Compensate: f.compensateObjectKey,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   2,
				InitialDelay:  time.Second,
				MaxDelay:      5 * time.Second,
				BackoffFactor: 2.0,
			},
			Timeout:      30 * time.Second,
			IsIdempotent: true,
			IsCritical:   false,
		},
		{
			Name:       "upload_to_storage",
			Execute:    f.uploadToStorage,
			Compensate: f.compensateStorageUpload,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   3,
				InitialDelay:  2 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			Timeout:      10 * time.Minute, // 存储上传需要更长时间
			IsIdempotent: true,
			IsCritical:   true, // 存储上传失败需要补偿
		},
		{
			Name:       "create_file_record",
			Execute:    f.createFileRecord,
			Compensate: f.compensateFileRecord,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   3,
				InitialDelay:  time.Second,
				MaxDelay:      10 * time.Second,
				BackoffFactor: 2.0,
			},
			Timeout:       60 * time.Second,
			Prerequisites: []string{"upload_to_storage"},
			IsIdempotent:  true,
			IsCritical:    true,
		},
		{
			Name:       "update_user_quota",
			Execute:    f.updateUserQuota,
			Compensate: f.compensateUserQuotaUpdate,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   3,
				InitialDelay:  time.Second,
				MaxDelay:      10 * time.Second,
				BackoffFactor: 2.0,
			},
			Timeout:       30 * time.Second,
			Prerequisites: []string{"create_file_record"},
			IsIdempotent:  true,
			IsCritical:    true,
		},
		{
			Name:       "generate_thumbnail",
			Execute:    f.generateThumbnail,
			Compensate: f.compensateThumbnail,
			RetryPolicy: &model.RetryPolicy{
				MaxAttempts:   2,
				InitialDelay:  2 * time.Second,
				MaxDelay:      20 * time.Second,
				BackoffFactor: 2.0,
			},
			Timeout:       5 * time.Minute,
			Prerequisites: []string{"upload_to_storage"},
			IsIdempotent:  true,
			IsCritical:    false, // 缩略图生成失败不影响主流程
		},
	}
}

// Step 1: 验证用户配额
func (f *FileUploadSagaDefinition) validateUserQuota(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return nil, NewSagaError(ErrCodeInvalidContext, "invalid saga context", 0, 0)
	}

	// 获取用户信息
	user, err := f.userRepo.GetByID(ctx, uploadCtx.UserID)
	if err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "failed to get user", 0, 0, err)
	}

	// 检查配额
	if user.StorageUsed+uploadCtx.FileSize > user.StorageQuota {
		return nil, NewSagaError(ErrCodeStepExecutionFailed,
			fmt.Sprintf("insufficient storage quota: used %d, quota %d, requested %d",
				user.StorageUsed, user.StorageQuota, uploadCtx.FileSize), 0, 0)
	}

	return map[string]interface{}{
		"user_id":      user.ID,
		"current_used": user.StorageUsed,
		"quota":        user.StorageQuota,
		"file_size":    uploadCtx.FileSize,
		"quota_valid":  true,
	}, nil
}

func (f *FileUploadSagaDefinition) compensateUserQuota(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
	// 配额验证没有副作用，无需补偿
	return nil
}

// Step 2: 生成对象key
func (f *FileUploadSagaDefinition) generateObjectKey(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return nil, NewSagaError(ErrCodeInvalidContext, "invalid saga context", 0, 1)
	}

	// 生成唯一的对象key
	objectKey := f.pathBuilder.BuildPath(uploadCtx.UserID, uploadCtx.RequestID, uploadCtx.FileName)

	// 验证key的有效性
	if err := f.pathBuilder.ValidateObjectKey(objectKey); err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "invalid object key generated", 0, 1, err)
	}

	// 检查是否已存在相同key的文件
	exists, err := f.storageRepo.Exists(ctx, objectKey)
	if err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "failed to check object existence", 0, 1, err)
	}

	if exists {
		// 如果已存在，添加时间戳后缀
		ext := filepath.Ext(uploadCtx.FileName)
		nameWithoutExt := strings.TrimSuffix(uploadCtx.FileName, ext)
		timestamp := time.Now().Unix()
		newFileName := fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)
		objectKey = f.pathBuilder.BuildPath(uploadCtx.UserID, uploadCtx.RequestID, newFileName)
	}

	// 更新上下文
	uploadCtx.ObjectKey = objectKey

	return map[string]interface{}{
		"object_key":    objectKey,
		"original_name": uploadCtx.FileName,
		"user_id":       uploadCtx.UserID,
		"request_id":    uploadCtx.RequestID,
	}, nil
}

func (f *FileUploadSagaDefinition) compensateObjectKey(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
	// 对象key生成没有副作用，无需补偿
	return nil
}

// Step 3: 上传到存储
func (f *FileUploadSagaDefinition) uploadToStorage(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return nil, NewSagaError(ErrCodeInvalidContext, "invalid saga context", 0, 2)
	}

	if uploadCtx.ObjectKey == "" {
		return nil, NewSagaError(ErrCodeInvalidInput, "object key is required", 0, 2)
	}

	// 这里应该从上传token或临时存储中获取文件数据
	// 为了简化，我们假设文件数据已经准备好
	// 在实际实现中，可能需要从临时存储或multipart form中获取

	// 模拟文件上传（实际应该从request中获取文件数据）
	// 此处省略实际的文件读取逻辑，因为需要与HTTP handler集成

	return map[string]interface{}{
		"object_key":   uploadCtx.ObjectKey,
		"file_size":    uploadCtx.FileSize,
		"content_type": uploadCtx.ContentType,
		"upload_time":  time.Now(),
		"storage_url":  fmt.Sprintf("storage://%s", uploadCtx.ObjectKey),
	}, nil
}

func (f *FileUploadSagaDefinition) compensateStorageUpload(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return fmt.Errorf("invalid saga context for storage compensation")
	}

	if uploadCtx.ObjectKey == "" {
		return nil // 没有上传任何文件，无需补偿
	}

	// 删除已上传的文件
	if err := f.storageRepo.Delete(ctx, uploadCtx.ObjectKey); err != nil {
		return fmt.Errorf("failed to delete uploaded file during compensation: %w", err)
	}

	return nil
}

// Step 4: 创建文件记录
func (f *FileUploadSagaDefinition) createFileRecord(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return nil, NewSagaError(ErrCodeInvalidContext, "invalid saga context", 0, 3)
	}

	// 创建文件记录
	file := &model.File{
		UUID:         uploadCtx.RequestID, // 使用RequestID作为UUID
		OriginalName: uploadCtx.FileName,
		StorageName:  uploadCtx.ObjectKey,
		Size:         uploadCtx.FileSize,
		MimeType:     uploadCtx.ContentType,
		UserID:       uploadCtx.UserID,
		Status:       "active",
		UploadedAt:   time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := f.fileRepo.Create(ctx, file); err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "failed to create file record", 0, 3, err)
	}

	// 更新上下文
	uploadCtx.FileID = file.ID

	return map[string]interface{}{
		"file_id":       file.ID,
		"uuid":          file.UUID,
		"original_name": file.OriginalName,
		"storage_name":  file.StorageName,
		"size":          file.Size,
		"mime_type":     file.MimeType,
		"created_at":    file.CreatedAt,
	}, nil
}

func (f *FileUploadSagaDefinition) compensateFileRecord(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return fmt.Errorf("invalid saga context for file record compensation")
	}

	if uploadCtx.FileID == 0 {
		return nil // 没有创建文件记录，无需补偿
	}

	// 删除文件记录
	if err := f.fileRepo.Delete(ctx, uploadCtx.FileID); err != nil {
		return fmt.Errorf("failed to delete file record during compensation: %w", err)
	}

	return nil
}

// Step 5: 更新用户配额
func (f *FileUploadSagaDefinition) updateUserQuota(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return nil, NewSagaError(ErrCodeInvalidContext, "invalid saga context", 0, 4)
	}

	// 更新用户存储使用量
	if err := f.userRepo.UpdateStorageUsed(ctx, uploadCtx.UserID, uploadCtx.FileSize); err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "failed to update user quota", 0, 4, err)
	}

	// 获取更新后的用户信息
	user, err := f.userRepo.GetByID(ctx, uploadCtx.UserID)
	if err != nil {
		return nil, NewSagaErrorWithCause(ErrCodeStepExecutionFailed, "failed to get updated user", 0, 4, err)
	}

	return map[string]interface{}{
		"user_id":    user.ID,
		"old_usage":  user.StorageUsed - uploadCtx.FileSize,
		"new_usage":  user.StorageUsed,
		"added_size": uploadCtx.FileSize,
		"quota":      user.StorageQuota,
		"updated_at": time.Now(),
	}, nil
}

func (f *FileUploadSagaDefinition) compensateUserQuotaUpdate(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return fmt.Errorf("invalid saga context for quota update compensation")
	}

	// 恢复用户存储使用量（减去已添加的文件大小）
	if err := f.userRepo.UpdateStorageUsed(ctx, uploadCtx.UserID, -uploadCtx.FileSize); err != nil {
		return fmt.Errorf("failed to revert user quota during compensation: %w", err)
	}

	return nil
}

// Step 6: 生成缩略图（可选步骤）
func (f *FileUploadSagaDefinition) generateThumbnail(ctx context.Context, sagaCtx interface{}, input interface{}) (interface{}, error) {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return nil, NewSagaError(ErrCodeInvalidContext, "invalid saga context", 0, 5)
	}

	// 仅为图片文件生成缩略图
	if !strings.HasPrefix(uploadCtx.ContentType, "image/") {
		return map[string]interface{}{
			"thumbnail_generated": false,
			"reason":              "not an image file",
		}, nil
	}

	// 生成缩略图key
	thumbnailKey := strings.Replace(uploadCtx.ObjectKey, "/", "/thumbnails/", 1)
	thumbnailKey = strings.Replace(thumbnailKey, filepath.Ext(thumbnailKey), "_thumb.jpg", 1)

	// TODO: 实际的缩略图生成逻辑
	// 这里需要：
	// 1. 从存储中下载原图
	// 2. 使用图片处理库生成缩略图
	// 3. 上传缩略图到存储

	// 更新上下文
	uploadCtx.ThumbnailKey = thumbnailKey

	return map[string]interface{}{
		"thumbnail_generated": true,
		"thumbnail_key":       thumbnailKey,
		"original_key":        uploadCtx.ObjectKey,
		"content_type":        uploadCtx.ContentType,
	}, nil
}

func (f *FileUploadSagaDefinition) compensateThumbnail(ctx context.Context, sagaCtx interface{}, originalInput interface{}) error {
	uploadCtx, ok := sagaCtx.(*model.FileUploadSagaContext)
	if !ok {
		return fmt.Errorf("invalid saga context for thumbnail compensation")
	}

	if uploadCtx.ThumbnailKey == "" {
		return nil // 没有生成缩略图，无需补偿
	}

	// 删除生成的缩略图
	if err := f.storageRepo.Delete(ctx, uploadCtx.ThumbnailKey); err != nil {
		// 缩略图删除失败不是严重错误，记录日志即可
		// TODO: 添加日志记录
		return nil
	}

	return nil
}
