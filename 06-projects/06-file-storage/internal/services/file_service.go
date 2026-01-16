package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"file-storage-service/internal/config"
	"file-storage-service/internal/models"
	"file-storage-service/internal/storage"
	"file-storage-service/internal/utils"
)

// FileService 文件服务
type FileService struct {
	fileStorage storage.FileStorage
	dbStorage   storage.DatabaseStorage
	config      *config.Config
}

// NewFileService 创建文件服务实例
func NewFileService(fileStorage storage.FileStorage, config *config.Config) (*FileService, error) {
	// 创建数据库存储
	dbStorage, err := storage.NewDatabaseStorage(config.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to create database storage: %w", err)
	}

	return &FileService{
		fileStorage: fileStorage,
		dbStorage:   dbStorage,
		config:      config,
	}, nil
}

// UploadFile 上传文件
func (fs *FileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, userID string, options UploadOptions) (*models.File, error) {
	// 验证文件
	if err := fs.validateFile(fileHeader); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// 打开文件
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// 生成文件ID和路径
	fileID := utils.GenerateFileID()
	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileName := fileID + extension
	storagePath := fmt.Sprintf("files/%s/%s", time.Now().Format("2006/01/02"), fileName)

	// 计算文件哈希
	hasher := sha256.New()
	teeReader := io.TeeReader(src, hasher)

	// 保存到存储
	if err := fs.fileStorage.Save(ctx, storagePath, teeReader, fileHeader.Size); err != nil {
		return nil, fmt.Errorf("failed to save file to storage: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))

	// 创建文件记录
	file := &models.File{
		ID:           fileID,
		Name:         fileName,
		OriginalName: fileHeader.Filename,
		Path:         storagePath,
		Size:         fileHeader.Size,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		Extension:    extension,
		Checksum:     checksum,
		OwnerID:      userID,
		Visibility:   options.Visibility,
		Status:       "processing",
		Encrypted:    options.Encrypt,
		Metadata:     make(map[string]interface{}),
		Tags:         options.Tags,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 如果MIME类型为空，尝试检测
	if file.MimeType == "" {
		file.MimeType = utils.DetectMimeType(extension)
	}

	// 保存到数据库
	if err := fs.dbStorage.CreateFile(ctx, file); err != nil {
		// 如果数据库保存失败，清理已上传的文件
		fs.fileStorage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	// 后台处理图片（如果是图片）
	if file.IsImage() {
		go fs.processImageAsync(context.Background(), file)
	}

	// 记录访问日志
	fs.logAccess(ctx, fileID, userID, "upload", true, "")

	file.Status = "ready"
	file.UpdatedAt = time.Now()
	fs.dbStorage.UpdateFile(ctx, file)

	return file, nil
}

// GetFile 获取文件信息
func (fs *FileService) GetFile(ctx context.Context, fileID, userID string) (*models.File, error) {
	file, err := fs.dbStorage.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// 检查访问权限
	if !fs.canAccess(file, userID) {
		return nil, fmt.Errorf("permission denied")
	}

	// 记录访问日志
	fs.logAccess(ctx, fileID, userID, "view", true, "")

	// 更新访问时间
	now := time.Now()
	file.AccessedAt = &now
	fs.dbStorage.UpdateFile(ctx, file)

	return file, nil
}

// DownloadFile 下载文件
func (fs *FileService) DownloadFile(ctx context.Context, fileID, userID string) (io.ReadCloser, *models.File, error) {
	file, err := fs.GetFile(ctx, fileID, userID)
	if err != nil {
		return nil, nil, err
	}

	// 从存储获取文件
	reader, err := fs.fileStorage.Get(ctx, file.Path)
	if err != nil {
		fs.logAccess(ctx, fileID, userID, "download", false, err.Error())
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 记录下载日志
	fs.logAccess(ctx, fileID, userID, "download", true, "")

	return reader, file, nil
}

// DeleteFile 删除文件
func (fs *FileService) DeleteFile(ctx context.Context, fileID, userID string) error {
	file, err := fs.dbStorage.GetFileByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// 检查删除权限
	if !fs.canDelete(file, userID) {
		return fmt.Errorf("permission denied")
	}

	// 从存储删除文件
	if err := fs.fileStorage.Delete(ctx, file.Path); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// 删除缩略图
	if file.ImageInfo != nil {
		for _, thumbnail := range file.ImageInfo.Thumbnails {
			fs.fileStorage.Delete(ctx, thumbnail.Path)
		}
	}

	// 从数据库删除记录
	if err := fs.dbStorage.DeleteFile(ctx, fileID); err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	// 记录删除日志
	fs.logAccess(ctx, fileID, userID, "delete", true, "")

	return nil
}

// ListFiles 列出文件
func (fs *FileService) ListFiles(ctx context.Context, userID string, options ListOptions) (*models.ListFilesResponse, error) {
	files, total, err := fs.dbStorage.GetFilesByOwner(ctx, userID, options.Page, options.PerPage)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// 设置文件URL
	baseURL := fs.getBaseURL()
	for _, file := range files {
		file.URL = file.GetURL(baseURL)
	}

	totalPages := int((total + int64(options.PerPage) - 1) / int64(options.PerPage))

	return &models.ListFilesResponse{
		Files:      files,
		TotalCount: total,
		Page:       options.Page,
		PerPage:    options.PerPage,
		TotalPages: totalPages,
	}, nil
}

// SearchFiles 搜索文件
func (fs *FileService) SearchFiles(ctx context.Context, query string, filters map[string]interface{}, page, perPage int) (*models.ListFilesResponse, error) {
	files, total, err := fs.dbStorage.SearchFiles(ctx, query, filters, page, perPage)
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	// 设置文件URL
	baseURL := fs.getBaseURL()
	for _, file := range files {
		file.URL = file.GetURL(baseURL)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &models.ListFilesResponse{
		Files:      files,
		TotalCount: total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// GenerateUploadToken 生成上传令牌
func (fs *FileService) GenerateUploadToken(ctx context.Context, userID string, options TokenOptions) (*models.TokenResponse, error) {
	token := utils.GenerateRandomString(32)
	expiresAt := time.Now().Add(time.Duration(options.ExpiresIn) * time.Second)

	uploadToken := &models.UploadToken{
		Token:        token,
		OwnerID:      userID,
		MaxSize:      options.MaxSize,
		AllowedTypes: options.AllowedTypes,
		MaxUsage:     options.MaxUsage,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}

	if err := fs.dbStorage.CreateUploadToken(ctx, uploadToken); err != nil {
		return nil, fmt.Errorf("failed to create upload token: %w", err)
	}

	return &models.TokenResponse{
		Token:        token,
		ExpiresIn:    options.ExpiresIn,
		MaxSize:      options.MaxSize,
		AllowedTypes: options.AllowedTypes,
		MaxUsage:     options.MaxUsage,
	}, nil
}

// ValidateUploadToken 验证上传令牌
func (fs *FileService) ValidateUploadToken(ctx context.Context, token string) (*models.UploadToken, error) {
	uploadToken, err := fs.dbStorage.GetUploadToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}

	// 检查使用次数
	if uploadToken.MaxUsage > 0 && uploadToken.UsageCount >= uploadToken.MaxUsage {
		return nil, fmt.Errorf("token usage limit exceeded")
	}

	return uploadToken, nil
}

// GetFileStats 获取文件统计
func (fs *FileService) GetFileStats(ctx context.Context) (*models.StatsResponse, error) {
	return fs.dbStorage.GetFileStats(ctx)
}

// 辅助方法

func (fs *FileService) validateFile(fileHeader *multipart.FileHeader) error {
	// 检查文件大小
	if fileHeader.Size > fs.config.Upload.MaxSize {
		return fmt.Errorf("file size exceeds limit: %d bytes", fs.config.Upload.MaxSize)
	}

	// 检查文件类型
	mimeType := fileHeader.Header.Get("Content-Type")
	if !fs.isAllowedType(mimeType) {
		return fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// 检查文件名
	if fileHeader.Filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	return nil
}

func (fs *FileService) isAllowedType(mimeType string) bool {
	for _, allowedType := range fs.config.Upload.AllowedTypes {
		if strings.Contains(allowedType, "*") {
			prefix := strings.TrimSuffix(allowedType, "*")
			if strings.HasPrefix(mimeType, prefix) {
				return true
			}
		} else if mimeType == allowedType {
			return true
		}
	}
	return false
}

func (fs *FileService) canAccess(file *models.File, userID string) bool {
	if file.Visibility == "public" {
		return true
	}
	return file.OwnerID == userID
}

func (fs *FileService) canDelete(file *models.File, userID string) bool {
	return file.OwnerID == userID || userID == "admin"
}

func (fs *FileService) getBaseURL() string {
	return fmt.Sprintf("http://%s:%s", fs.config.Server.Host, fs.config.Server.Port)
}

func (fs *FileService) logAccess(ctx context.Context, fileID, userID, action string, success bool, errorMsg string) {
	log := &models.AccessLog{
		FileID:    fileID,
		UserID:    userID,
		Action:    action,
		Success:   success,
		ErrorMsg:  errorMsg,
		CreatedAt: time.Now(),
	}
	fs.dbStorage.CreateAccessLog(ctx, log)
}

func (fs *FileService) processImageAsync(ctx context.Context, file *models.File) {
	// 简化的图片处理实现
	// 实际实现中应该使用专业的图片处理库
	// 这里只是创建图片信息记录
	imageInfo := &models.ImageInfo{
		FileID:    file.ID,
		Width:     1024, // 示例值
		Height:    768,  // 示例值
		Format:    strings.TrimPrefix(file.Extension, "."),
		CreatedAt: time.Now(),
	}

	fs.dbStorage.CreateImageInfo(ctx, imageInfo)
}

// UploadOptions 上传选项
type UploadOptions struct {
	Visibility string   `json:"visibility"`
	Encrypt    bool     `json:"encrypt"`
	Tags       []string `json:"tags"`
}

// ListOptions 列表选项
type ListOptions struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// TokenOptions 令牌选项
type TokenOptions struct {
	ExpiresIn    int      `json:"expires_in"`
	MaxSize      int64    `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
	MaxUsage     int      `json:"max_usage"`
}
