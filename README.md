# Go-API Framework

**Languages**: [English](README.md) | [中文](README-zh.md)

---

### Overview

`go-api` is a powerful, high-performance Go framework designed for building enterprise-grade web APIs. It provides a complete solution with layered architecture, dependency injection, comprehensive middleware support, and automatic code generation capabilities.

### Key Features

- 🚀 **High Performance**: Built on Gin framework with optimized logging and database connections
- 🏗️ **Layered Architecture**: Strict Model → Repository → Service → Controller pattern
- 🔧 **Dependency Injection**: Clean architecture with proper separation of concerns
- ⚙️ **Configuration Management**: Multi-environment support with JSON-based configuration
- 📝 **Advanced Logging**: Structured logging with Zap for high performance
- 🗄️ **Multi-Database Support**: MySQL, PostgreSQL, SQLite, SQL Server, ClickHouse (via xdb/GORM) and MongoDB (qmgo)
- 🔐 **JWT Authentication**: Built-in app authentication with JWT tokens
- 🌐 **Internationalization**: Multi-language support (zh-CN, en-US)
- 📊 **Middleware System**: CORS, authentication, request logging, and custom middleware
- ⚡ **Code Generation**: Automatic model and repository generation from SQL files
- 🔄 **Task Scheduling**: Built-in job scheduling system
- 📨 **Message Queue**: Kafka producer/consumer support
- 🚨 **Monitoring**: Panic recovery with notification integration
- 🐳 **Docker Ready**: Complete Docker support with optimized images

### Quick Start

#### Method 1: Using Project Generator Script

```bash
# Download the project generator
curl -O https://raw.githubusercontent.com/seakee/go-api/main/scripts/generate.sh
chmod +x generate.sh

# Generate a new project
./generate.sh my-api-project v1.0.0
cd my-api-project

# Install dependencies and run
go mod tidy
make run
```

#### Method 2: Clone and Customize

```bash
# Clone the repository
git clone https://github.com/seakee/go-api.git
cd go-api

# Install dependencies
go mod download

# Copy and configure local settings
cp bin/configs/local.json.default bin/configs/local.json
# Edit bin/configs/local.json with your database settings

# Run the application
make run
```

### Architecture Overview

```
go-api/
├── app/                             # Application layer
│   ├── config/                     # Configuration management
│   │   └── config.go              # Config loader and structures
│   ├── http/                       # HTTP layer
│   │   ├── controller/             # HTTP controllers
│   │   │   ├── auth/               # Authentication controllers
│   │   │   │   ├── app.go          # App CRUD operations
│   │   │   │   ├── handler.go      # Auth handler interface
│   │   │   │   └── jwt.go          # JWT token operations
│   │   │   └── base.go             # Base controller
│   │   ├── middleware/             # HTTP middleware
│   │   │   ├── check_app_auth.go   # JWT authentication
│   │   │   ├── cors.go             # CORS handling
│   │   │   ├── handler.go          # Middleware interface
│   │   │   ├── request_logger.go   # Request logging
│   │   │   └── set_trace_id.go     # Trace ID injection
│   │   ├── router/                 # Route definitions
│   │   │   ├── external/           # External API routes
│   │   │   │   └── service/        # External service routes
│   │   │   │       └── auth/       # Auth endpoints
│   │   │   ├── internal/           # Internal API routes
│   │   │   │   └── service/        # Internal service routes
│   │   │   │       └── auth/       # Auth endpoints
│   │   │   └── handler.go          # Main router
│   │   └── context.go              # HTTP context wrapper
│   ├── model/                      # Data models
│   │   └── auth/                   # Authentication models
│   │       ├── app.go              # App model (MySQL)
│   │       └── app_mgo.go          # App model (MongoDB)
│   ├── pkg/                        # Utility packages
│   │   ├── e/                      # Error codes
│   │   │   └── code.go             # Error code definitions
│   │   ├── jwt/                    # JWT utilities
│   │   │   └── jwt.go              # JWT generation/parsing
│   │   ├── schedule/               # Task scheduling
│   │   │   └── schedule.go         # Job scheduler
│   │   └── trace/                  # Distributed tracing
│   │       └── trace.go            # Trace ID generation
│   ├── repository/                 # Data access layer
│   │   └── auth/                   # Auth repository
│   │       └── app.go              # App repository
│   ├── service/                    # Business logic layer
│   │   └── auth/                   # Auth services
│   │       └── app.go              # App service
│   └── worker/                     # Background workers
│       └── handler.go              # Worker handler
├── bin/                            # Runtime resources
│   ├── configs/                    # Configuration files
│   │   ├── dev.json                # Development config
│   │   ├── local.json              # Local config
│   │   └── prod.json               # Production config
│   ├── data/                       # Data files
│   │   └── sql/                    # SQL scripts
│   │       └── auth_app.sql        # App table schema
│   └── lang/                       # Language files
│       ├── en-US.json              # English messages
│       └── zh-CN.json              # Chinese messages
├── bootstrap/                      # Application bootstrap
│   ├── app.go                      # Main app initialization
│   ├── database.go                 # Database setup
│   ├── http.go                     # HTTP server setup
│   ├── kafka.go                    # Kafka setup
│   └── schedule.go                 # Scheduler setup
├── command/                        # CLI commands
│   └── codegen/                    # Code generator
│       ├── codegen/                # Generator logic
│       ├── handler.go              # CLI handler
│       └── README.md               # Generator docs
├── scripts/                        # Utility scripts
│   └── generate.sh                 # Project generator
├── docs/                           # Project documentation
│   ├── Home.md                     # Wiki homepage (English)
│   ├── Home-zh.md                  # Wiki homepage (Chinese)
│   ├── Architecture-Design.md      # Architecture documentation
│   ├── Development-Guide.md        # Development workflow guide
│   ├── API-Documentation.md        # Complete API reference
│   ├── Code-Generator-Guide.md     # Code generation tool guide
│   └── Deployment-Guide.md         # Production deployment guide
├── Dockerfile                      # Docker configuration
├── Makefile                        # Build automation
├── docker-compose.yml              # Docker Compose
├── go.mod                          # Go module
├── go.sum                          # Dependencies
└── main.go                         # Application entry point
```

### Core Components

#### 1. Layered Architecture

The framework follows a strict 4-layer architecture:

- **Model Layer**: Data structures and database operations
- **Repository Layer**: Data access abstraction with interfaces
- **Service Layer**: Business logic implementation
- **Controller Layer**: HTTP request handling and response formatting

#### 2. Configuration Management

Supports multiple environments with JSON-based configuration:

```json
{
  "system": {
    "name": "go-api",
    "run_mode": "debug",
    "http_port": ":8080",
    "jwt_secret": "your-secret-key"
  },
  "databases": [
    {
      "enable": true,
      "db_type": "mysql",
      "db_name": "go-api",
      "db_host": "localhost",
      "db_port": 3306,
      "charset": "utf8mb4",
      "conn_max_lifetime": 3
    }
  ]
}
```

#### 3. Middleware System

Built-in middleware for common functionality:

- **Authentication**: JWT-based app authentication
- **CORS**: Cross-origin resource sharing
- **Logging**: Structured request/response logging
- **Trace ID**: Distributed tracing support
- **Panic Recovery**: Automatic panic recovery with notifications

#### 4. Authentication System

Complete JWT-based authentication:

```bash
# Get JWT token
curl -X POST http://localhost:8080/go-api/external/service/auth/token \
  -d "app_id=your_app_id&app_secret=your_app_secret"

# Use token in requests
curl -H "Authorization: your_jwt_token" \
  http://localhost:8080/go-api/external/service/auth/app
```

### Development Guide

#### Adding a New Controller

1. Create controller structure:

```go
// app/http/controller/user/handler.go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/seakee/go-api/app/http"
)

type Handler interface {
    Create() gin.HandlerFunc
    GetByID() gin.HandlerFunc
}

type handler struct {
    controller.BaseController
    service userService.UserService
}

func NewHandler(appCtx *http.Context) Handler {
    return &handler{
        BaseController: controller.BaseController{
            AppCtx: appCtx,
            Logger: appCtx.Logger,
            Redis:  appCtx.Redis["go-api"],
            I18n:   appCtx.I18n,
        },
        service: userService.NewUserService(appCtx.SqlDB["go-api"], appCtx.Redis["go-api"]),
    }
}
```

2. Register routes:

```go
// app/http/router/external/service/user/user.go
func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
    userHandler := user.NewHandler(ctx)
    {
        api.POST("user", ctx.Middleware.CheckAppAuth(), userHandler.Create())
        api.GET("user/:id", userHandler.GetByID())
    }
}
```

#### Adding Middleware

1. Define in interface:

```go
// app/http/middleware/handler.go
type Middleware interface {
    CheckAppAuth() gin.HandlerFunc
    YourNewMiddleware() gin.HandlerFunc  // Add this
}
```

2. Implement middleware:

```go
// app/http/middleware/your_middleware.go
func (m middleware) YourNewMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Your middleware logic
        c.Next()
    }
}
```

#### Code Generation

Generate models and repositories from SQL files:

```bash
# Generate from SQL file
go run ./command/codegen/handler.go -name user_table

# Generate all SQL files
go run ./command/codegen/handler.go

# Custom paths
go run ./command/codegen/handler.go -sql custom/sql -model custom/model
```

SQL file format:
```sql
CREATE TABLE `users` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `username` varchar(50) NOT NULL COMMENT 'Username',
    `email` varchar(100) NOT NULL COMMENT 'Email Address',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Status',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User Information';
```

### API Endpoints

#### External APIs (Public)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/go-api/external/service/auth/token` | Get JWT token | No |
| POST | `/go-api/external/service/auth/app` | Create app | Yes |
| GET | `/go-api/external/service/ping` | Health check | No |

#### Internal APIs (Private)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/go-api/internal/service/auth/token` | Get JWT token | No |
| POST | `/go-api/internal/service/auth/app` | Create app | Yes |
| GET | `/go-api/internal/service/ping` | Health check | No |

### Docker Deployment

#### Using Docker Compose

```yaml
# docker-compose.yml
version: '3.8'
services:
  go-api:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./bin/configs:/bin/configs
      - ./bin/logs:/bin/logs
    environment:
      - RUN_ENV=prod
      - APP_NAME=go-api
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: go-api
    ports:
      - "3306:3306"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

#### Build and Run

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
docker-compose up -d

# Run single container
make docker-run
```

### Build Commands

```bash
# Development
make run          # Run application
make test         # Run tests
make fmt          # Format code
make all          # fmt + test + build

# Production
make build        # Build binary
make docker-build # Build Docker image
make docker-run   # Run Docker container
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|----------|
| `RUN_ENV` | Runtime environment | `local` |
| `APP_NAME` | Application name | `go-api` |
| `CONFIG_DIR` | Configuration directory | `./bin/configs` |

### Documentation

Complete project documentation is available in the `docs/` directory:

- **[📚 GitHub Wiki](https://github.com/seakee/go-api/wiki)** - Complete wiki with all documentation
- **[Wiki Home](docs/Home.md)** - Documentation index and quick navigation
- **[Architecture Design](docs/Architecture-Design.md)** - System architecture and design patterns
- **[Development Guide](docs/Development-Guide.md)** - Detailed development workflow
- **[API Documentation](docs/API-Documentation.md)** - Complete API reference
- **[Code Generator](docs/Code-Generator-Guide.md)** - Code generation tool guide
- **[Deployment Guide](docs/Deployment-Guide.md)** - Production deployment guide
- **[Makefile Usage](docs/Makefile-Usage.md)** - Build automation and development tools
- **[go-api.sh Usage](docs/go-api.sh-Usage.md)** - Shell script alternative to Makefile

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/new-feature`
3. Commit changes: `git commit -am 'Add new feature'`
4. Push to branch: `git push origin feature/new-feature`
5. Submit a Pull Request

See [Contributing Guide](CONTRIBUTING.md) for more details.

### License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
