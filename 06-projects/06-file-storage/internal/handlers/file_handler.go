package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"file-storage-service/internal/models"
	"file-storage-service/internal/services"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService *services.FileService
}

// NewFileHandler 创建文件处理器
func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile 上传文件
func (fh *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 解析多部分表单
	if err := r.ParseMultipartForm(100 << 20); err != nil { // 100MB
		fh.sendError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 获取用户ID
	userID := fh.getUserID(r)

	// 获取上传选项
	options := services.UploadOptions{
		Visibility: r.FormValue("visibility"),
		Encrypt:    r.FormValue("encrypt") == "true",
		Tags:       strings.Split(r.FormValue("tags"), ","),
	}

	if options.Visibility == "" {
		options.Visibility = "private"
	}

	// 处理多个文件
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		fh.sendError(w, "No files provided", http.StatusBadRequest)
		return
	}

	var uploadedFiles []*models.File
	var errors []string

	for _, fileHeader := range files {
		file, err := fh.fileService.UploadFile(ctx, fileHeader, userID, options)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", fileHeader.Filename, err))
			continue
		}
		uploadedFiles = append(uploadedFiles, file)
	}

	// 构造响应
	response := &models.UploadResponse{
		Files: uploadedFiles,
		Count: len(uploadedFiles),
	}

	if len(errors) > 0 {
		if len(uploadedFiles) == 0 {
			fh.sendError(w, "All uploads failed: "+strings.Join(errors, "; "), http.StatusBadRequest)
			return
		}
		response.Message = fmt.Sprintf("Partially successful. Errors: %s", strings.Join(errors, "; "))
	} else {
		response.Message = "All files uploaded successfully"
	}

	fh.sendJSON(w, response)
}

// GetFile 获取文件信息
func (fh *FileHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	fileID := vars["id"]
	userID := fh.getUserID(r)

	file, err := fh.fileService.GetFile(ctx, fileID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fh.sendError(w, "File not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "permission denied") {
			fh.sendError(w, "Permission denied", http.StatusForbidden)
		} else {
			fh.sendError(w, "Failed to get file", http.StatusInternalServerError)
		}
		return
	}

	fh.sendJSON(w, map[string]interface{}{
		"file": file,
	})
}

// DownloadFile 下载文件
func (fh *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	fileID := vars["id"]
	userID := fh.getUserID(r)

	reader, file, err := fh.fileService.DownloadFile(ctx, fileID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fh.sendError(w, "File not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "permission denied") {
			fh.sendError(w, "Permission denied", http.StatusForbidden)
		} else {
			fh.sendError(w, "Failed to download file", http.StatusInternalServerError)
		}
		return
	}
	defer reader.Close()

	// 设置响应头
	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	// 检查是否为内联显示
	if r.URL.Query().Get("inline") == "true" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.OriginalName))
	} else {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	}

	// 复制文件内容
	http.ServeContent(w, r, file.OriginalName, file.CreatedAt, &readerAtSeeker{reader})
}

// DeleteFile 删除文件
func (fh *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	fileID := vars["id"]
	userID := fh.getUserID(r)

	err := fh.fileService.DeleteFile(ctx, fileID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			fh.sendError(w, "File not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "permission denied") {
			fh.sendError(w, "Permission denied", http.StatusForbidden)
		} else {
			fh.sendError(w, "Failed to delete file", http.StatusInternalServerError)
		}
		return
	}

	fh.sendJSON(w, map[string]string{
		"message": "File deleted successfully",
	})
}

// ListFiles 列出文件
func (fh *FileHandler) ListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := fh.getUserID(r)

	// 解析查询参数
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(query.Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	options := services.ListOptions{
		Page:    page,
		PerPage: perPage,
	}

	// 检查是否为搜索请求
	searchQuery := query.Get("q")
	if searchQuery != "" {
		filters := make(map[string]interface{})

		if mimeType := query.Get("type"); mimeType != "" {
			filters["mime_type"] = mimeType
		}

		if visibility := query.Get("visibility"); visibility != "" {
			filters["visibility"] = visibility
		}

		response, err := fh.fileService.SearchFiles(ctx, searchQuery, filters, page, perPage)
		if err != nil {
			fh.sendError(w, "Failed to search files", http.StatusInternalServerError)
			return
		}

		fh.sendJSON(w, response)
		return
	}

	// 普通列表请求
	response, err := fh.fileService.ListFiles(ctx, userID, options)
	if err != nil {
		fh.sendError(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	fh.sendJSON(w, response)
}

// GenerateUploadToken 生成上传令牌
func (fh *FileHandler) GenerateUploadToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := fh.getUserID(r)

	var req struct {
		ExpiresIn    int      `json:"expires_in"`
		MaxSize      int64    `json:"max_size"`
		AllowedTypes []string `json:"allowed_types"`
		MaxUsage     int      `json:"max_usage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fh.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if req.ExpiresIn == 0 {
		req.ExpiresIn = 3600 // 1小时
	}
	if req.MaxSize == 0 {
		req.MaxSize = 100 << 20 // 100MB
	}
	if req.MaxUsage == 0 {
		req.MaxUsage = 1
	}

	options := services.TokenOptions{
		ExpiresIn:    req.ExpiresIn,
		MaxSize:      req.MaxSize,
		AllowedTypes: req.AllowedTypes,
		MaxUsage:     req.MaxUsage,
	}

	response, err := fh.fileService.GenerateUploadToken(ctx, userID, options)
	if err != nil {
		fh.sendError(w, "Failed to generate upload token", http.StatusInternalServerError)
		return
	}

	fh.sendJSON(w, response)
}

// GetStats 获取统计信息
func (fh *FileHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := fh.fileService.GetFileStats(ctx)
	if err != nil {
		fh.sendError(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	fh.sendJSON(w, stats)
}

// ServeUploadPage 服务上传页面
func (fh *FileHandler) ServeUploadPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>File Storage System</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 2rem; }
        .header { text-align: center; margin-bottom: 2rem; }
        .upload-area { background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
        .drop-zone { border: 2px dashed #cbd5e0; border-radius: 8px; padding: 3rem; text-align: center; transition: all 0.3s; }
        .drop-zone.dragover { border-color: #4299e1; background: #ebf8ff; }
        .btn { padding: 0.75rem 1.5rem; border: none; border-radius: 6px; cursor: pointer; font-weight: 500; }
        .btn-primary { background: #4299e1; color: white; }
        .btn:hover { opacity: 0.9; }
        .file-list { margin-top: 2rem; }
        .file-item { display: flex; align-items: center; padding: 1rem; border-bottom: 1px solid #e2e8f0; }
        .file-icon { width: 40px; height: 40px; margin-right: 1rem; display: flex; align-items: center; justify-content: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 File Storage System</h1>
            <p>Secure, fast, and reliable file storage</p>
        </div>

        <div class="upload-area">
            <div class="drop-zone" id="dropZone">
                <h3>📤 Drop files here or click to upload</h3>
                <p>Supports multiple files, drag and drop</p>
                <input type="file" id="fileInput" multiple style="display: none;">
                <button class="btn btn-primary" onclick="document.getElementById('fileInput').click()">Choose Files</button>
            </div>

            <div class="file-list" id="fileList">
                <!-- Files will be listed here -->
            </div>
        </div>
    </div>

    <script>
        const dropZone = document.getElementById('dropZone');
        const fileInput = document.getElementById('fileInput');
        const fileList = document.getElementById('fileList');

        // Drag and drop handlers
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('dragover');
        });

        dropZone.addEventListener('dragleave', () => {
            dropZone.classList.remove('dragover');
        });

        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('dragover');
            handleFiles(e.dataTransfer.files);
        });

        fileInput.addEventListener('change', (e) => {
            handleFiles(e.target.files);
        });

        async function handleFiles(files) {
            const formData = new FormData();

            for (let file of files) {
                formData.append('files', file);
            }

            try {
                const response = await fetch('/api/v1/files', {
                    method: 'POST',
                    body: formData
                });

                const result = await response.json();

                if (response.ok) {
                    alert('Upload successful: ' + result.count + ' files uploaded');
                    displayFiles(result.files);
                } else {
                    alert('Upload failed: ' + result.error);
                }
            } catch (error) {
                alert('Upload failed: ' + error.message);
            }
        }

        function displayFiles(files) {
            fileList.innerHTML = '<h3>Recently Uploaded Files:</h3>';

            files.forEach(file => {
                const fileItem = document.createElement('div');
                fileItem.className = 'file-item';
                fileItem.innerHTML = ` + "`" + `
                    <div class="file-icon">${getFileIcon(file.mime_type)}</div>
                    <div>
                        <div><strong>${file.original_name}</strong></div>
                        <div><small>${formatFileSize(file.size)} • ${file.mime_type}</small></div>
                    </div>
                ` + "`" + `;
                fileList.appendChild(fileItem);
            });
        }

        function getFileIcon(mimeType) {
            if (mimeType.startsWith('image/')) return '🖼️';
            if (mimeType.startsWith('video/')) return '🎥';
            if (mimeType.startsWith('audio/')) return '🎵';
            if (mimeType.includes('pdf')) return '📄';
            if (mimeType.includes('text')) return '📝';
            return '📁';
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
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// 辅助方法

func (fh *FileHandler) getUserID(r *http.Request) string {
	// 简化的用户ID获取，实际项目中应该从JWT或session中获取
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}
	return userID
}

func (fh *FileHandler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (fh *FileHandler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(&models.ErrorResponse{
		Error: message,
		Code:  statusCode,
	})
}

// readerAtSeeker 包装器，用于http.ServeContent
type readerAtSeeker struct {
	reader interface{}
}

func (ras *readerAtSeeker) Read(p []byte) (n int, err error) {
	if r, ok := ras.reader.(interface{ Read([]byte) (int, error) }); ok {
		return r.Read(p)
	}
	return 0, fmt.Errorf("reader does not support Read")
}

func (ras *readerAtSeeker) Seek(offset int64, whence int) (int64, error) {
	if s, ok := ras.reader.(interface{ Seek(int64, int) (int64, error) }); ok {
		return s.Seek(offset, whence)
	}
	return 0, fmt.Errorf("reader does not support Seek")
}

func (ras *readerAtSeeker) ReadAt(p []byte, off int64) (n int, err error) {
	if ra, ok := ras.reader.(interface{ ReadAt([]byte, int64) (int, error) }); ok {
		return ra.ReadAt(p, off)
	}
	return 0, fmt.Errorf("reader does not support ReadAt")
}