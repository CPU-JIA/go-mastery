# 🛍️ Modern E-commerce Backend API

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/doc/go1.24)
[![Gin Framework](https://img.shields.io/badge/Gin-v1.10+-00D4AA?style=for-the-badge)](https://gin-gonic.com/)
[![GORM](https://img.shields.io/badge/GORM-v1.25+-FF6B6B?style=for-the-badge)](https://gorm.io/)

A modern, scalable e-commerce backend API built with Go, featuring Clean Architecture, comprehensive product management, shopping cart, order processing, and payment integration.

## ✨ Key Features

### 🏗️ Architecture Excellence
- **Clean Architecture** - Separation of concerns with clear layer boundaries
- **Repository Pattern** - Data access abstraction for maintainability
- **Dependency Injection** - Loosely coupled components
- **RESTful API Design** - Standard HTTP methods and status codes

### 🔒 Security & Authentication
- **JWT Authentication** - Secure token-based authentication
- **Password Encryption** - bcrypt secure password hashing
- **Role-Based Access Control** - Admin, seller, and customer roles
- **Input Validation** - Comprehensive data validation and sanitization
- **Rate Limiting** - API abuse protection
- **CORS Support** - Cross-origin resource sharing

### 🛒 E-commerce Core Features
- **User Management** - Registration, authentication, profile management
- **Product Catalog** - Advanced product management with categories, tags, and images
- **Shopping Cart** - Persistent cart with real-time updates
- **Order Processing** - Complete order lifecycle management
- **Payment Integration** - Multiple payment gateway support
- **Inventory Management** - Stock tracking with low-stock alerts
- **Coupon System** - Flexible discount and promotion system
- **Product Reviews** - Customer feedback and rating system
- **Wishlist** - Save products for later functionality

### 📊 Data Management
- **GORM Integration** - Modern Go ORM with advanced features
- **Database Flexibility** - SQLite for development, PostgreSQL for production
- **Auto Migration** - Database schema synchronization
- **Soft Delete** - Safe data removal with recovery options
- **Transaction Support** - ACID compliance for data integrity
- **Connection Pooling** - Optimized database performance

### 🚀 Enterprise Features
- **Configuration Management** - Viper-based flexible configuration
- **Structured Logging** - Comprehensive request/response logging
- **Health Checks** - Service monitoring endpoints
- **Graceful Shutdown** - Clean service termination
- **Docker Support** - Containerized deployment
- **Hot Reload** - Development-friendly auto-restart
- **API Documentation** - Comprehensive endpoint documentation

## 📋 System Requirements

- **Go 1.24+**
- **SQLite 3.x** (Development)
- **PostgreSQL 12+** (Production recommended)
- **Redis 6.0+** (Optional, for caching)

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd 02-ecommerce-backend
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Configure Environment

Copy and edit configuration files:

```bash
cp configs/config.yaml.example configs/config.yaml
cp .env.example .env
```

### 4. Run the Application

```bash
# Development mode
go run cmd/server/main.go

# Or with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Production build
go build -o ecommerce-server cmd/server/main.go
./ecommerce-server
```

### 5. Verify Installation

Access these endpoints to verify the service:

- **Health Check**: http://localhost:8080/health
- **API Root**: http://localhost:8080/api/v1
- **Service Info**: http://localhost:8080/

## 📖 API Documentation

### 🔑 Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/register` | User registration | No |
| POST | `/api/v1/auth/login` | User login | No |
| POST | `/api/v1/auth/refresh` | Refresh JWT token | No |

### 🛍️ Product Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/products` | Get product list | Optional |
| GET | `/api/v1/products/{id}` | Get product details | Optional |
| GET | `/api/v1/products/slug/{slug}` | Get product by slug | Optional |
| GET | `/api/v1/products/search` | Search products | Optional |
| GET | `/api/v1/products/featured` | Get featured products | Optional |
| GET | `/api/v1/products/{id}/related` | Get related products | Optional |

### 👤 User Management

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| PUT | `/api/v1/user/password` | Change password | Yes |

### 🛒 Shopping Cart (Planned)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/cart` | Get cart contents | Yes |
| POST | `/api/v1/cart` | Add item to cart | Yes |
| PUT | `/api/v1/cart/items/{product_id}` | Update item quantity | Yes |
| DELETE | `/api/v1/cart/items/{product_id}` | Remove item from cart | Yes |

### 📦 Order Management (Planned)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/orders` | Get user orders | Yes |
| POST | `/api/v1/orders` | Create new order | Yes |
| GET | `/api/v1/orders/{id}` | Get order details | Yes |
| PUT | `/api/v1/orders/{id}/cancel` | Cancel order | Yes |

### 💳 Payment Processing (Planned)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/payments/orders/{order_id}` | Process payment | Yes |
| GET | `/api/v1/payments/{id}` | Get payment details | Yes |

## 🏗️ Project Structure

```
02-ecommerce-backend/
├── cmd/server/              # Application entry point
│   └── main.go
├── internal/                # Private application code
│   ├── config/             # Configuration management
│   ├── handler/            # HTTP request handlers
│   ├── middleware/         # HTTP middleware
│   ├── model/             # Data models and schemas
│   ├── repository/        # Data access layer
│   └── service/           # Business logic layer
├── configs/               # Configuration files
│   └── config.yaml
├── scripts/              # Database and deployment scripts
├── test/                # Test files and utilities
├── logs/                # Application logs
├── uploads/             # File uploads directory
├── Dockerfile           # Container configuration
├── docker-compose.yml   # Multi-service deployment
├── .env.example        # Environment variables template
├── go.mod              # Go module definition
└── README.md           # Project documentation
```

## 🔧 Configuration Options

### Server Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"          # debug, release, test
  read_timeout: 60s
  write_timeout: 60s
```

### Database Configuration

```yaml
database:
  driver: "sqlite"       # sqlite, postgres
  sqlite:
    path: "ecommerce.db"
  postgres:
    host: "localhost"
    port: 5432
    user: "ecommerce_user"
    password: "ecommerce_password"
    dbname: "ecommerce_system"
```

### JWT Configuration

```yaml
jwt:
  secret: "your-super-secret-jwt-key"
  expires_in: 24h
  refresh_expires_in: 168h
```

## 🐳 Docker Deployment

### Build and Run with Docker

```bash
# Build image
docker build -t ecommerce-backend:latest .

# Run container
docker run -d -p 8080:8080 \
  --name ecommerce-backend \
  -e ECOMMERCE_DATABASE_DRIVER=sqlite \
  ecommerce-backend:latest
```

### Full Stack with Docker Compose

```bash
# Start all services
docker-compose up -d

# Start with development tools
docker-compose --profile dev up -d

# Start with Nginx proxy
docker-compose --profile with-nginx up -d
```

Services included:
- **API Server**: Port 8080
- **PostgreSQL**: Port 5432
- **Redis**: Port 6379
- **Adminer** (dev profile): Port 8081
- **Redis Commander** (dev profile): Port 8082
- **Nginx** (with-nginx profile): Port 80/443

## 📝 Development Guide

### Adding New Features

1. **Data Models** - Define in `internal/model/`
2. **Repository Layer** - Implement in `internal/repository/`
3. **Business Logic** - Add to `internal/service/`
4. **HTTP Handlers** - Create in `internal/handler/`
5. **Route Registration** - Update `cmd/server/main.go`

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/service/

# Generate test coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Linting (install golangci-lint first)
golangci-lint run
```

## 🔍 API Usage Examples

### User Registration

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### User Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "login_id": "johndoe",
    "password": "securepassword123"
  }'
```

### Get Products

```bash
curl -X GET "http://localhost:8080/api/v1/products?page=1&limit=20&category_id=1"
```

### Search Products

```bash
curl -X GET "http://localhost:8080/api/v1/products/search?q=electronics&page=1&limit=10"
```

## 🎯 Default Accounts

The system creates default accounts for testing:

- **Admin**: `admin` / `admin123`
- **Customer**: `customer` / `customer123`

## 🔍 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check database configuration in config.yaml or environment variables
   - Verify database service is running
   - Ensure proper permissions and credentials

2. **JWT Authentication Failed**
   - Verify JWT secret configuration
   - Check token format in Authorization header: `Bearer <token>`
   - Ensure token hasn't expired

3. **Permission Denied**
   - Check user role and permissions
   - Verify authentication middleware is applied correctly

4. **Rate Limit Exceeded**
   - Adjust rate limiting settings in configuration
   - Implement user-specific rate limiting if needed

### Log Analysis

```bash
# View real-time logs
tail -f logs/app.log

# Search for errors
grep "ERROR" logs/app.log

# View startup logs
grep "Starting\|Health\|Database" logs/app.log
```

## 🛡️ Security Best Practices

- Always use environment variables for sensitive configuration
- Enable HTTPS in production
- Regularly update dependencies
- Implement proper input validation
- Use strong JWT secrets
- Enable rate limiting
- Implement proper CORS policies
- Regular security audits

## 🤝 Contributing

1. Fork the project
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- [Gin](https://gin-gonic.com/) - HTTP web framework
- [GORM](https://gorm.io/) - Go ORM library
- [Viper](https://github.com/spf13/viper) - Configuration management
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT implementation
- [Shopspring Decimal](https://github.com/shopspring/decimal) - Decimal number handling

---

**🎉 Success!** Your modern e-commerce backend API is ready for business!

For detailed API documentation and advanced usage examples, check the inline code documentation and API endpoint descriptions above.