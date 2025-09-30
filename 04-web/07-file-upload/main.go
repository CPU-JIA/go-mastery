package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"go-mastery/common/security"
)

/*
文件上传和处理练习

本练习涵盖Go语言中的文件上传和处理，包括：
1. 单文件和多文件上传
2. 文件类型验证和安全检查
3. 图片处理和缩放
4. 文件存储策略
5. 上传进度跟踪
6. 大文件分块上传
7. 云存储集成
8. 文件元数据管理

主要概念：
- multipart/form-data处理
- 文件MIME类型检测
- 图像处理和缩放
- 文件安全验证
- 存储优化策略
*/

// === 文件信息结构体 ===

type FileInfo struct {
	ID           string                 `json:"id"`
	OriginalName string                 `json:"original_name"`
	Filename     string                 `json:"filename"`
	Size         int64                  `json:"size"`
	MimeType     string                 `json:"mime_type"`
	Extension    string                 `json:"extension"`
	Path         string                 `json:"path"`
	URL          string                 `json:"url"`
	Hash         string                 `json:"hash"`
	UploadedAt   time.Time              `json:"uploaded_at"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type ImageInfo struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format string `json:"format"`
}

// === 文件上传配置 ===

type UploadConfig struct {
	MaxFileSize           int64                // 最大文件大小（字节）
	MaxFiles              int                  // 最大文件数量
	AllowedTypes          []string             // 允许的MIME类型
	AllowedExts           []string             // 允许的文件扩展名
	UploadDir             string               // 上传目录
	EnableImageProcessing bool                 // 是否启用图片处理
	ImageSizes            map[string]ImageSize // 图片尺寸配置
}

type ImageSize struct {
	Width   uint
	Height  uint
	Quality int
}

// === 文件处理器 ===

type FileHandler struct {
	config    UploadConfig
	storage   FileStorage
	processor ImageProcessor
}

func NewFileHandler() *FileHandler {
	config := UploadConfig{
		MaxFileSize: 10 << 20, // 10MB
		MaxFiles:    5,
		AllowedTypes: []string{
			"image/jpeg", "image/png", "image/gif", "image/webp",
			"application/pdf", "text/plain", "application/json",
			"video/mp4", "audio/mp3",
		},
		AllowedExts: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp",
			".pdf", ".txt", ".json", ".mp4", ".mp3",
		},
		UploadDir:             "uploads",
		EnableImageProcessing: true,
		ImageSizes: map[string]ImageSize{
			"thumbnail": {Width: 150, Height: 150, Quality: 80},
			"medium":    {Width: 500, Height: 500, Quality: 85},
			"large":     {Width: 1200, Height: 1200, Quality: 90},
		},
	}

	// 确保上传目录存在
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		log.Printf("创建上传目录失败: %v", err)
	}
	for size := range config.ImageSizes {
		if err := os.MkdirAll(filepath.Join(config.UploadDir, size), 0755); err != nil {
			log.Printf("创建尺寸目录失败: %v", err)
		}
	}

	return &FileHandler{
		config:    config,
		storage:   NewLocalStorage(config.UploadDir),
		processor: NewImageProcessor(),
	}
}

// === 文件存储接口 ===

type FileStorage interface {
	Save(filename string, data io.Reader) error
	Delete(filename string) error
	GetURL(filename string) string
	Exists(filename string) bool
}

// 本地存储实现
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (ls *LocalStorage) Save(filename string, data io.Reader) error {
	fullPath := filepath.Join(ls.basePath, filename)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	// G301/G306安全修复：使用安全权限创建文件
	file, err := security.SecureCreateFile(fullPath, security.DefaultFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}

func (ls *LocalStorage) Delete(filename string) error {
	fullPath := filepath.Join(ls.basePath, filename)
	return os.Remove(fullPath)
}

func (ls *LocalStorage) GetURL(filename string) string {
	return "/uploads/" + filename
}

func (ls *LocalStorage) Exists(filename string) bool {
	fullPath := filepath.Join(ls.basePath, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// === 图片处理器 ===

type ImageProcessor interface {
	Process(src io.Reader, format string) (image.Image, error)
	Resize(img image.Image, width, height uint) image.Image
	Save(img image.Image, writer io.Writer, format string, quality int) error
	GetImageInfo(src io.Reader) (*ImageInfo, error)
}

type DefaultImageProcessor struct{}

func NewImageProcessor() *DefaultImageProcessor {
	return &DefaultImageProcessor{}
}

func (p *DefaultImageProcessor) Process(src io.Reader, format string) (image.Image, error) {
	switch format {
	case "image/jpeg":
		return jpeg.Decode(src)
	case "image/png":
		return png.Decode(src)
	default:
		img, _, err := image.Decode(src)
		return img, err
	}
}

func (p *DefaultImageProcessor) Resize(img image.Image, width, height uint) image.Image {
	return resize.Resize(width, height, img, resize.Lanczos3)
}

func (p *DefaultImageProcessor) Save(img image.Image, writer io.Writer, format string, quality int) error {
	switch format {
	case "image/jpeg":
		return jpeg.Encode(writer, img, &jpeg.Options{Quality: quality})
	case "image/png":
		return png.Encode(writer, img)
	default:
		return fmt.Errorf("不支持的图片格式: %s", format)
	}
}

func (p *DefaultImageProcessor) GetImageInfo(src io.Reader) (*ImageInfo, error) {
	config, format, err := image.DecodeConfig(src)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		Width:  config.Width,
		Height: config.Height,
		Format: format,
	}, nil
}

// === 文件验证和安全 ===

func (fh *FileHandler) validateFile(header *multipart.FileHeader) error {
	// 检查文件大小
	if header.Size > fh.config.MaxFileSize {
		return fmt.Errorf("文件大小超过限制: %d bytes > %d bytes", header.Size, fh.config.MaxFileSize)
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !contains(fh.config.AllowedExts, ext) {
		return fmt.Errorf("不允许的文件扩展名: %s", ext)
	}

	return nil
}

func (fh *FileHandler) detectMimeType(file multipart.File) (string, error) {
	// 读取文件头部用于MIME类型检测
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", err
	}
	// 如果读取的字节数不足，调整buffer大小
	if n < 512 {
		buffer = buffer[:n]
	}

	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("重置文件指针失败: %v", err)
	}

	// 检测MIME类型
	mimeType := http.DetectContentType(buffer)

	// 验证MIME类型
	if !contains(fh.config.AllowedTypes, mimeType) {
		return "", fmt.Errorf("不允许的文件类型: %s", mimeType)
	}

	return mimeType, nil
}

func (fh *FileHandler) calculateHash(file multipart.File) (string, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// === 文件上传处理 ===

// 单文件上传
func (fh *FileHandler) HandleSingleUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	err := r.ParseMultipartForm(fh.config.MaxFileSize)
	if err != nil {
		http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "获取文件失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件
	if err := fh.validateFile(header); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 检测MIME类型
	mimeType, err := fh.detectMimeType(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 计算文件哈希
	hash, err := fh.calculateHash(file)
	if err != nil {
		http.Error(w, "计算文件哈希失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hash[:8], ext)

	// 保存文件
	if err := fh.storage.Save(filename, file); err != nil {
		http.Error(w, "保存文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建文件信息
	fileInfo := &FileInfo{
		ID:           generateID(),
		OriginalName: header.Filename,
		Filename:     filename,
		Size:         header.Size,
		MimeType:     mimeType,
		Extension:    ext,
		Path:         filepath.Join(fh.config.UploadDir, filename),
		URL:          fh.storage.GetURL(filename),
		Hash:         hash,
		UploadedAt:   time.Now(),
	}

	// 如果是图片，进行图片处理
	if strings.HasPrefix(mimeType, "image/") && fh.config.EnableImageProcessing {
		if err := fh.processImage(file, filename, mimeType, fileInfo); err != nil {
			log.Printf("图片处理失败: %v", err)
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"file":    fileInfo,
	})
}

// 多文件上传
func (fh *FileHandler) HandleMultipleUpload(w http.ResponseWriter, r *http.Request) {
	// 解析multipart表单
	err := r.ParseMultipartForm(fh.config.MaxFileSize * int64(fh.config.MaxFiles))
	if err != nil {
		http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) > fh.config.MaxFiles {
		http.Error(w, fmt.Sprintf("文件数量超过限制: %d > %d", len(files), fh.config.MaxFiles), http.StatusBadRequest)
		return
	}

	var uploadedFiles []*FileInfo
	var errors []string

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: 打开文件失败", header.Filename))
			continue
		}

		// 验证文件
		if err := fh.validateFile(header); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", header.Filename, err.Error()))
			file.Close()
			continue
		}

		// 检测MIME类型
		mimeType, err := fh.detectMimeType(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", header.Filename, err.Error()))
			file.Close()
			continue
		}

		// 计算文件哈希
		hash, err := fh.calculateHash(file)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: 计算哈希失败", header.Filename))
			file.Close()
			continue
		}

		// 生成唯一文件名
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hash[:8], ext)

		// 保存文件
		if err := fh.storage.Save(filename, file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: 保存失败", header.Filename))
			file.Close()
			continue
		}

		// 创建文件信息
		fileInfo := &FileInfo{
			ID:           generateID(),
			OriginalName: header.Filename,
			Filename:     filename,
			Size:         header.Size,
			MimeType:     mimeType,
			Extension:    ext,
			Path:         filepath.Join(fh.config.UploadDir, filename),
			URL:          fh.storage.GetURL(filename),
			Hash:         hash,
			UploadedAt:   time.Now(),
		}

		// 如果是图片，进行图片处理
		if strings.HasPrefix(mimeType, "image/") && fh.config.EnableImageProcessing {
			if err := fh.processImage(file, filename, mimeType, fileInfo); err != nil {
				log.Printf("图片处理失败: %v", err)
			}
		}

		uploadedFiles = append(uploadedFiles, fileInfo)
		file.Close()
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": len(errors) == 0,
		"files":   uploadedFiles,
		"errors":  errors,
	}); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// 图片处理
func (fh *FileHandler) processImage(file multipart.File, filename, mimeType string, fileInfo *FileInfo) error {
	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// 获取图片信息
	imageInfo, err := fh.processor.GetImageInfo(file)
	if err != nil {
		return err
	}

	// 添加图片元数据
	fileInfo.Metadata = map[string]interface{}{
		"image":      imageInfo,
		"thumbnails": make(map[string]string),
	}

	// 重置文件指针
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// 解码图片
	img, err := fh.processor.Process(file, mimeType)
	if err != nil {
		return err
	}

	// 生成不同尺寸的图片
	thumbnails := make(map[string]string)
	for sizeName, sizeConfig := range fh.config.ImageSizes {
		// 调整图片大小
		resizedImg := fh.processor.Resize(img, sizeConfig.Width, sizeConfig.Height)

		// 生成缩略图文件名
		ext := filepath.Ext(filename)
		thumbFilename := fmt.Sprintf("%s_%s%s", strings.TrimSuffix(filename, ext), sizeName, ext)
		thumbPath := filepath.Join(sizeName, thumbFilename)

		// 保存缩略图
		// G301/G306安全修复：使用安全权限创建文件
		thumbFile, err := security.SecureCreateFile(filepath.Join(fh.config.UploadDir, thumbPath), security.DefaultFileMode)
		if err != nil {
			return err
		}

		err = fh.processor.Save(resizedImg, thumbFile, mimeType, sizeConfig.Quality)
		if closeErr := thumbFile.Close(); closeErr != nil {
			log.Printf("关闭缩略图文件失败: %v", closeErr)
		}
		if err != nil {
			return err
		}

		thumbnails[sizeName] = fh.storage.GetURL(thumbPath)
	}

	// 更新文件信息
	fileInfo.Metadata["thumbnails"] = thumbnails

	return nil
}

// === 大文件分块上传 ===

type ChunkUpload struct {
	ChunkID     string `json:"chunk_id"`
	Filename    string `json:"filename"`
	ChunkIndex  int    `json:"chunk_index"`
	TotalChunks int    `json:"total_chunks"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalSize   int64  `json:"total_size"`
}

// 初始化分块上传
func (fh *FileHandler) HandleChunkUploadInit(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename  string `json:"filename"`
		FileSize  int64  `json:"file_size"`
		ChunkSize int64  `json:"chunk_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析请求失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 计算分块数量
	totalChunks := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)

	// 生成分块上传ID
	chunkID := generateID()

	// 创建临时目录
	tempDir := filepath.Join(fh.config.UploadDir, "chunks", chunkID)
	// G306安全修复：检查目录创建错误，使用安全权限
	if err := os.MkdirAll(tempDir, 0750); err != nil {
		http.Error(w, "创建临时目录失败", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"chunk_id":     chunkID,
		"total_chunks": totalChunks,
		"chunk_size":   req.ChunkSize,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// 上传文件块
func (fh *FileHandler) HandleChunkUpload(w http.ResponseWriter, r *http.Request) {
	// 解析表单
	err := r.ParseMultipartForm(fh.config.MaxFileSize)
	if err != nil {
		http.Error(w, "解析表单失败", http.StatusBadRequest)
		return
	}

	chunkID := r.FormValue("chunk_id")
	chunkIndexStr := r.FormValue("chunk_index")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		http.Error(w, "无效的块索引: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("chunk")
	if err != nil {
		http.Error(w, "获取文件块失败", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 保存文件块
	chunkPath := filepath.Join(fh.config.UploadDir, "chunks", chunkID, fmt.Sprintf("chunk_%d", chunkIndex))
	// G301/G306安全修复：使用安全权限创建文件
	chunkFile, err := security.SecureCreateFile(chunkPath, security.DefaultFileMode)
	if err != nil {
		http.Error(w, "创建文件块失败", http.StatusInternalServerError)
		return
	}
	defer chunkFile.Close()

	_, err = io.Copy(chunkFile, file)
	if err != nil {
		http.Error(w, "保存文件块失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"chunk_index": chunkIndex,
	}); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// 完成分块上传
func (fh *FileHandler) HandleChunkUploadComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChunkID     string `json:"chunk_id"`
		Filename    string `json:"filename"`
		TotalChunks int    `json:"total_chunks"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "解析请求失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 合并文件块
	chunksDir := filepath.Join(fh.config.UploadDir, "chunks", req.ChunkID)

	// 生成最终文件名
	ext := filepath.Ext(req.Filename)
	finalFilename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), req.ChunkID[:8], ext)
	finalPath := filepath.Join(fh.config.UploadDir, finalFilename)

	// 创建最终文件
	// G301/G306安全修复：使用安全权限创建文件
	finalFile, err := security.SecureCreateFile(finalPath, security.DefaultFileMode)
	if err != nil {
		http.Error(w, "创建最终文件失败", http.StatusInternalServerError)
		return
	}
	defer finalFile.Close()

	// 合并所有块
	for i := 0; i < req.TotalChunks; i++ {
		chunkPath := filepath.Join(chunksDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("打开文件块 %d 失败", i), http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(finalFile, chunkFile)
		if err := chunkFile.Close(); err != nil {
			log.Printf("关闭文件块失败: %v", err)
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("合并文件块 %d 失败", i), http.StatusInternalServerError)
			return
		}
	}

	// 清理临时文件
	if err := os.RemoveAll(chunksDir); err != nil {
		log.Printf("清理临时文件失败: %v", err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(finalPath)
	if err != nil {
		http.Error(w, "获取文件信息失败", http.StatusInternalServerError)
		return
	}

	// 创建文件信息响应
	result := &FileInfo{
		ID:           generateID(),
		OriginalName: req.Filename,
		Filename:     finalFilename,
		Size:         fileInfo.Size(),
		Extension:    ext,
		Path:         finalPath,
		URL:          fh.storage.GetURL(finalFilename),
		UploadedAt:   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"file":    result,
	}); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// === 文件管理API ===

// 获取文件列表
func (fh *FileHandler) HandleFileList(w http.ResponseWriter, r *http.Request) {
	// 遍历上传目录
	var files []FileInfo

	err := filepath.Walk(fh.config.UploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && !strings.Contains(path, "chunks") && !strings.Contains(path, "thumbnail") && !strings.Contains(path, "medium") && !strings.Contains(path, "large") {
			relPath, _ := filepath.Rel(fh.config.UploadDir, path)
			ext := filepath.Ext(info.Name())

			fileInfo := FileInfo{
				ID:           generateID(),
				OriginalName: info.Name(),
				Filename:     info.Name(),
				Size:         info.Size(),
				Extension:    ext,
				Path:         path,
				URL:          fh.storage.GetURL(relPath),
				UploadedAt:   info.ModTime(),
			}

			// 检测MIME类型
			if file, err := os.Open(path); err == nil {
				buffer := make([]byte, 512)
				if n, readErr := file.Read(buffer); readErr == nil {
					if n < 512 {
						buffer = buffer[:n]
					}
					fileInfo.MimeType = http.DetectContentType(buffer)
				}
				if closeErr := file.Close(); closeErr != nil {
					log.Printf("关闭文件失败: %v", closeErr)
				}
			}

			files = append(files, fileInfo)
		}

		return nil
	})

	if err != nil {
		http.Error(w, "读取文件列表失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"total": len(files),
	}); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// 删除文件
func (fh *FileHandler) HandleFileDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		http.Error(w, "文件名不能为空", http.StatusBadRequest)
		return
	}

	// 删除主文件
	if err := fh.storage.Delete(filename); err != nil {
		http.Error(w, "删除文件失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 删除缩略图
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	for sizeName := range fh.config.ImageSizes {
		thumbFilename := fmt.Sprintf("%s_%s%s", baseName, sizeName, ext)
		thumbPath := filepath.Join(sizeName, thumbFilename)
		if err := fh.storage.Delete(thumbPath); err != nil {
			log.Printf("删除缩略图失败: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件删除成功",
	}); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}

// === 辅助函数 ===

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// === 上传进度跟踪 ===

type ProgressTracker struct {
	uploads map[string]*UploadProgress
}

type UploadProgress struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	TotalSize    int64     `json:"total_size"`
	UploadedSize int64     `json:"uploaded_size"`
	Progress     float64   `json:"progress"`
	Status       string    `json:"status"` // uploading, completed, error
	StartTime    time.Time `json:"start_time"`
	UpdateTime   time.Time `json:"update_time"`
}

func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		uploads: make(map[string]*UploadProgress),
	}
}

func (pt *ProgressTracker) Start(id, filename string, totalSize int64) {
	pt.uploads[id] = &UploadProgress{
		ID:         id,
		Filename:   filename,
		TotalSize:  totalSize,
		Status:     "uploading",
		StartTime:  time.Now(),
		UpdateTime: time.Now(),
	}
}

func (pt *ProgressTracker) Update(id string, uploadedSize int64) {
	if progress, exists := pt.uploads[id]; exists {
		progress.UploadedSize = uploadedSize
		progress.Progress = float64(uploadedSize) / float64(progress.TotalSize) * 100
		progress.UpdateTime = time.Now()
	}
}

func (pt *ProgressTracker) Complete(id string) {
	if progress, exists := pt.uploads[id]; exists {
		progress.Status = "completed"
		progress.Progress = 100
		progress.UpdateTime = time.Now()
	}
}

func (pt *ProgressTracker) Error(id string) {
	if progress, exists := pt.uploads[id]; exists {
		progress.Status = "error"
		progress.UpdateTime = time.Now()
	}
}

func (pt *ProgressTracker) Get(id string) (*UploadProgress, bool) {
	progress, exists := pt.uploads[id]
	return progress, exists
}

// === 示例：上传页面HTML ===

const uploadHTML = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文件上传示例</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .upload-area { border: 2px dashed #ccc; padding: 40px; text-align: center; margin: 20px 0; }
        .upload-area.dragover { border-color: #007bff; background-color: #f8f9fa; }
        .file-list { margin: 20px 0; }
        .file-item { border: 1px solid #ddd; padding: 10px; margin: 5px 0; border-radius: 5px; }
        .progress { width: 100%; height: 20px; background-color: #f0f0f0; border-radius: 10px; overflow: hidden; }
        .progress-bar { height: 100%; background-color: #007bff; transition: width 0.3s; }
        .thumbnail { max-width: 100px; max-height: 100px; margin: 10px; }
    </style>
</head>
<body>
    <h1>文件上传示例</h1>

    <div class="upload-area" id="uploadArea">
        <p>拖拽文件到此处或点击选择文件</p>
        <input type="file" id="fileInput" multiple style="display: none;">
        <button onclick="document.getElementById('fileInput').click()">选择文件</button>
    </div>

    <div class="file-list" id="fileList"></div>

    <script>
        const uploadArea = document.getElementById('uploadArea');
        const fileInput = document.getElementById('fileInput');
        const fileList = document.getElementById('fileList');

        // 拖拽上传
        uploadArea.addEventListener('dragover', (e) => {
            e.preventDefault();
            uploadArea.classList.add('dragover');
        });

        uploadArea.addEventListener('dragleave', () => {
            uploadArea.classList.remove('dragover');
        });

        uploadArea.addEventListener('drop', (e) => {
            e.preventDefault();
            uploadArea.classList.remove('dragover');
            const files = e.dataTransfer.files;
            uploadFiles(files);
        });

        // 文件选择上传
        fileInput.addEventListener('change', (e) => {
            uploadFiles(e.target.files);
        });

        function uploadFiles(files) {
            const formData = new FormData();
            for (let file of files) {
                formData.append('files', file);
            }

            const fileItem = createFileItem('批量上传', files.length + ' 个文件');

            fetch('/api/upload/multiple', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    fileItem.innerHTML = '<h4>上传成功</h4>';
                    data.files.forEach(file => {
                        const item = createFileItem(file.original_name, formatFileSize(file.size));
                        if (file.mime_type.startsWith('image/')) {
                            const img = document.createElement('img');
                            img.src = file.url;
                            img.className = 'thumbnail';
                            item.appendChild(img);
                        }
                    });
                } else {
                    fileItem.innerHTML = '<h4>上传失败</h4><p>' + (data.errors || []).join('<br>') + '</p>';
                }
            })
            .catch(error => {
                fileItem.innerHTML = '<h4>上传错误</h4><p>' + error.message + '</p>';
            });
        }

        function createFileItem(name, info) {
            const item = document.createElement('div');
            item.className = 'file-item';
            item.innerHTML = '<h4>' + name + '</h4><p>' + info + '</p>';
            fileList.appendChild(item);
            return item;
        }

        function formatFileSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }
    </script>
</body>
</html>
`

func handleUploadPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(uploadHTML)); err != nil {
		log.Printf("写入响应失败: %v", err)
	}
}

func main() {
	// 创建文件处理器
	fileHandler := NewFileHandler()

	// 创建路由器
	router := mux.NewRouter()

	// 上传页面
	router.HandleFunc("/", handleUploadPage).Methods("GET")

	// 文件上传API
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/upload/single", fileHandler.HandleSingleUpload).Methods("POST")
	api.HandleFunc("/upload/multiple", fileHandler.HandleMultipleUpload).Methods("POST")

	// 分块上传API
	api.HandleFunc("/upload/chunk/init", fileHandler.HandleChunkUploadInit).Methods("POST")
	api.HandleFunc("/upload/chunk", fileHandler.HandleChunkUpload).Methods("POST")
	api.HandleFunc("/upload/chunk/complete", fileHandler.HandleChunkUploadComplete).Methods("POST")

	// 文件管理API
	api.HandleFunc("/files", fileHandler.HandleFileList).Methods("GET")
	api.HandleFunc("/files/{filename}", fileHandler.HandleFileDelete).Methods("DELETE")

	// 静态文件服务
	router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads/"))))

	fmt.Println("=== 文件上传服务器启动 ===")
	fmt.Println("页面端点:")
	fmt.Println("  GET  /                    - 文件上传页面")
	fmt.Println()
	fmt.Println("API端点:")
	fmt.Println("  POST /api/upload/single   - 单文件上传")
	fmt.Println("  POST /api/upload/multiple - 多文件上传")
	fmt.Println("  POST /api/upload/chunk/init - 初始化分块上传")
	fmt.Println("  POST /api/upload/chunk    - 上传文件块")
	fmt.Println("  POST /api/upload/chunk/complete - 完成分块上传")
	fmt.Println("  GET  /api/files           - 获取文件列表")
	fmt.Println("  DELETE /api/files/{filename} - 删除文件")
	fmt.Println()
	fmt.Println("文件访问:")
	fmt.Println("  GET  /uploads/{filename}  - 访问上传的文件")
	fmt.Println()
	fmt.Println("服务器运行在 http://localhost:8080")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

/*
练习任务：

1. 基础练习：
   - 添加更多文件类型支持
   - 实现文件重命名功能
   - 添加文件下载API
   - 实现文件夹组织功能

2. 中级练习：
   - 实现断点续传功能
   - 添加文件版本控制
   - 实现文件搜索和过滤
   - 添加文件分享链接功能

3. 高级练习：
   - 集成云存储（AWS S3、阿里云OSS）
   - 实现CDN加速
   - 添加视频转码功能
   - 实现文件同步机制

4. 安全练习：
   - 实现文件病毒扫描
   - 添加访问权限控制
   - 实现文件加密存储
   - 添加水印功能

5. 性能优化：
   - 实现并发上传
   - 添加文件压缩
   - 优化大文件处理
   - 实现缓存策略

运行前准备：
1. 安装依赖：
   go get github.com/gorilla/mux
   go get github.com/nfnt/resize

2. 创建上传目录：
   mkdir -p uploads

3. 运行程序：go run main.go

4. 访问应用：http://localhost:8080

目录结构：
project/
├── main.go
└── uploads/
    ├── (上传的文件)
    ├── thumbnail/
    │   └── (缩略图)
    ├── medium/
    │   └── (中等尺寸图片)
    ├── large/
    │   └── (大尺寸图片)
    └── chunks/
        └── (临时文件块)

扩展建议：
- 实现图片自动旋转和优化
- 添加文件元数据提取
- 集成图像识别和标签
- 实现文件去重机制
*/
