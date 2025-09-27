# File Storage Service

A comprehensive file storage and management system built with Go, featuring secure file upload, download, processing, and management capabilities.

## 🚀 Features

### Core Features
- **Multi-file upload** with drag-and-drop support
- **Secure file storage** with encryption support
- **Image processing** and thumbnail generation
- **File versioning** and metadata management
- **Access control** and permission management
- **RESTful API** with comprehensive endpoints

### Advanced Features
- **Multiple storage backends** (Local, MinIO)
- **File compression** and batch operations
- **Upload tokens** for secure temporary access
- **Access logging** and audit trails
- **File search** and filtering
- **Statistics** and usage analytics

### Security Features
- **File validation** and type checking
- **Path traversal protection**
- **Encryption at rest** (AES-256)
- **Access control** and visibility settings
- **Audit logging** for all operations

## 📋 Prerequisites

- Go 1.24.6 or higher
- SQLite (default) or other supported databases
- MinIO (optional, for object storage)

## 🛠️ Installation

### 1. Clone the repository
```bash
git clone <repository-url>
cd file-storage-service
```

### 2. Install dependencies
```bash
go mod tidy
```

### 3. Configure environment
```bash
cp .env.example .env
# Edit .env with your configuration
```

### 4. Run the service
```bash
go run cmd/server/main.go
```

The service will start on `http://localhost:8080`

## 📖 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HOST` | Server host | `localhost` |
| `PORT` | Server port | `8080` |
| `DB_DRIVER` | Database driver | `sqlite` |
| `DB_DSN` | Database connection string | `file_storage.db` |
| `STORAGE_PROVIDER` | Storage provider (local/minio) | `local` |
| `STORAGE_LOCAL_PATH` | Local storage path | `./uploads` |
| `UPLOAD_MAX_SIZE` | Maximum file size in bytes | `104857600` (100MB) |
| `ENCRYPTION_KEY` | Encryption key (32 bytes) | `myverystrongpasswordo32bitlength` |

### Storage Providers

#### Local Storage
```env
STORAGE_PROVIDER=local
STORAGE_LOCAL_PATH=./uploads
```

#### MinIO Object Storage
```env
STORAGE_PROVIDER=minio
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=files
MINIO_USE_SSL=false
```

## 🔌 API Endpoints

### File Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/files` | Upload files |
| `GET` | `/api/v1/files` | List files |
| `GET` | `/api/v1/files/{id}` | Get file info |
| `GET` | `/api/v1/files/{id}/download` | Download file |
| `DELETE` | `/api/v1/files/{id}` | Delete file |

### Token Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/upload-token` | Generate upload token |

### Statistics

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/stats` | Get usage statistics |

## 📝 API Usage Examples

### Upload Files
```bash
curl -X POST \
  -F "files=@document.pdf" \
  -F "files=@image.jpg" \
  -F "visibility=public" \
  -F "encrypt=false" \
  http://localhost:8080/api/v1/files
```

### List Files
```bash
curl -X GET \
  -H "X-User-ID: user123" \
  "http://localhost:8080/api/v1/files?page=1&per_page=20"
```

### Download File
```bash
curl -X GET \
  -H "X-User-ID: user123" \
  -o downloaded_file.pdf \
  "http://localhost:8080/api/v1/files/{file-id}/download"
```

### Search Files
```bash
curl -X GET \
  -H "X-User-ID: user123" \
  "http://localhost:8080/api/v1/files?q=document&type=application/pdf"
```

### Generate Upload Token
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -d '{
    "expires_in": 3600,
    "max_size": 52428800,
    "allowed_types": ["image/*", "application/pdf"],
    "max_usage": 5
  }' \
  http://localhost:8080/api/v1/upload-token
```

## 🏗️ Architecture

```
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   ├── storage/         # Storage abstractions
│   └── utils/           # Utility functions
├── docs/                # API documentation
├── scripts/             # Deployment scripts
└── test/                # Test files
```

## 🧪 Testing

Run tests with:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

## 🐳 Docker Deployment

Build and run with Docker:
```bash
docker build -t file-storage-service .
docker run -p 8080:8080 file-storage-service
```

Or use Docker Compose:
```bash
docker-compose up -d
```

## 📊 Monitoring & Logging

The service provides:

- **Access logs** for all file operations
- **Error tracking** with detailed error messages
- **Performance metrics** via statistics endpoint
- **Health checks** for storage and database connectivity

## 🔒 Security Considerations

### Production Deployment
- Change default encryption keys
- Use environment variables for sensitive data
- Enable HTTPS/TLS
- Implement proper authentication
- Set up file size limits
- Configure rate limiting
- Regular security audits

### File Security
- All uploads are validated for type and size
- Path traversal attacks are prevented
- Files can be encrypted at rest
- Access controls are enforced
- Audit logs track all operations

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the documentation in the `/docs` folder
- Review the API examples above

## 🔄 Version History

- **v1.0.0** - Initial release with core functionality
  - File upload/download
  - Basic image processing
  - Local and MinIO storage
  - RESTful API
  - Web interface

## 🎯 Roadmap

- [ ] Advanced image processing (watermarks, filters)
- [ ] Video thumbnail generation
- [ ] Chunked upload for large files
- [ ] File deduplication
- [ ] Cloud storage integrations (AWS S3, Google Cloud)
- [ ] Advanced search with full-text indexing
- [ ] Real-time notifications
- [ ] Admin dashboard