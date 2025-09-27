# Blog System API Documentation

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/logout` - User logout

### Articles
- `GET /api/v1/articles` - Get article list
- `GET /api/v1/articles/:id` - Get article by ID
- `POST /api/v1/articles` - Create new article
- `PUT /api/v1/articles/:id` - Update article
- `DELETE /api/v1/articles/:id` - Delete article

### Comments
- `GET /api/v1/articles/:id/comments` - Get article comments
- `POST /api/v1/articles/:id/comments` - Add comment
- `DELETE /api/v1/comments/:id` - Delete comment

### Categories
- `GET /api/v1/categories` - Get all categories
- `POST /api/v1/categories` - Create category

## Request/Response Examples

### Create Article
```json
POST /api/v1/articles
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Go语言性能优化实践",
  "content": "详细的性能优化内容...",
  "category_id": 1,
  "tags": ["golang", "performance"]
}
```

### Response
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": 123,
    "title": "Go语言性能优化实践",
    "slug": "go-performance-optimization",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

## Architecture

The blog system follows Clean Architecture principles:

```
├── cmd/server/           # Application entry point
├── internal/
│   ├── handler/         # HTTP handlers
│   ├── service/         # Business logic
│   ├── repository/      # Data access
│   └── model/          # Domain models
├── pkg/                # Reusable packages
└── docs/              # API documentation
```

## Database Schema

### Articles Table
```sql
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    user_id INTEGER REFERENCES users(id),
    category_id INTEGER REFERENCES categories(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

## Development

### Setup
```bash
cd 06-projects/01-blog-system
go mod tidy
go run cmd/server/main.go
```

### Testing
```bash
go test ./...
```

### Build
```bash
go build -o bin/blog-server cmd/server/main.go
```