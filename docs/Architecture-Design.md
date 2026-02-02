# Architecture Design

**Languages**: [English](Architecture-Design.md) | [中文](Architecture-Design-zh.md)

---

## Overview

The Go-API framework follows a clean, layered architecture pattern designed for enterprise-grade applications. It emphasizes separation of concerns, dependency injection, and maintainability.

### Architecture Layers

```
┌─────────────────────────────────────┐
│           Presentation Layer        │
│        (HTTP Controllers)           │
├─────────────────────────────────────┤
│           Business Layer            │
│           (Services)                │
├─────────────────────────────────────┤
│         Data Access Layer           │
│         (Repositories)              │
├─────────────────────────────────────┤
│            Data Layer               │
│      (Models & Database)            │
└─────────────────────────────────────┘
```

#### 1. Presentation Layer (Controllers)

**Purpose**: Handle HTTP requests and responses
**Location**: `app/http/controller/`

- Receives HTTP requests from clients
- Validates input parameters
- Calls appropriate service methods
- Formats and returns HTTP responses
- Handles authentication and authorization via middleware

**Example Structure**:
```go
type Handler interface {
    Create() gin.HandlerFunc
    GetByID() gin.HandlerFunc
    Update() gin.HandlerFunc
    Delete() gin.HandlerFunc
}

type handler struct {
    controller.BaseController
    service auth.AppService
}
```

#### 2. Business Layer (Services)

**Purpose**: Implement business logic and rules
**Location**: `app/service/`

- Contains all business logic
- Orchestrates operations across multiple repositories
- Implements complex business rules and validations
- Handles transactions and data consistency
- Independent of HTTP concerns

**Example Structure**:
```go
type AppService interface {
    CreateApp(ctx context.Context, params *CreateAppParams) (*CreateAppResult, error)
    ValidateCredentials(ctx context.Context, appID, appSecret string) (*App, error)
}

type appService struct {
    repo auth.AppRepo
}
```

#### 3. Data Access Layer (Repositories)

**Purpose**: Abstract data access operations
**Location**: `app/repository/`

- Provides abstraction over database operations
- Implements data access patterns
- Handles database connections and queries
- Converts between domain models and database entities
- Supports multiple database types (MySQL, MongoDB)

**Example Structure**:
```go
type AppRepo interface {
    Create(ctx context.Context, app *auth.App) (uint, error)
    GetByID(ctx context.Context, id uint) (*auth.App, error)
    Update(ctx context.Context, id uint, app *auth.App) error
    Delete(ctx context.Context, id uint) error
}
```

#### 4. Data Layer (Models)

**Purpose**: Define data structures and basic operations
**Location**: `app/model/`

- Defines data structures (structs)
- Contains basic CRUD operations
- Includes database schema definitions (GORM tags)
- Handles data validation and serialization
- Supports both SQL and NoSQL databases

### Dependency Injection Pattern

The framework uses constructor-based dependency injection:

```go
// Service layer receives only what it needs
func NewAppService(db *gorm.DB, redis *redis.Manager) AppService {
    repo := authRepo.NewAppRepo(db, redis)
    return &appService{repo: repo}
}

// Controller layer injects dependencies into service
func NewHandler(appCtx *http.Context) Handler {
    return &handler{
        service: authService.NewAppService(appCtx.SqlDB["go-api"], appCtx.Redis["go-api"]),
    }
}
```

### Middleware Architecture

```
Request → SetTraceID → CORS → RequestLogger → CheckAppAuth → Controller
```

Middleware components:
- **SetTraceID**: Generates unique trace IDs for request tracking
- **CORS**: Handles cross-origin requests
- **RequestLogger**: Logs request details for monitoring
- **CheckAppAuth**: Validates JWT tokens for authentication

### Configuration Management

Multi-environment configuration with hot-reload support:

```json
{
  "system": {
    "name": "go-api",
    "run_mode": "debug|release",
    "http_port": ":8080"
  },
  "databases": [...],
  "redis": [...],
  "kafka": {...}
}
```

Environment-specific files:
- `local.json` - Local development
- `dev.json` - Development environment
- `prod.json` - Production environment

### Error Handling Strategy

1. **Error Propagation**: Errors bubble up through layers
2. **Centralized Handling**: Controllers handle all error responses
3. **Internationalization**: Error messages support multiple languages
4. **Structured Logging**: All errors are logged with context

### Security Architecture

1. **JWT Authentication**: Token-based authentication system
2. **Middleware Protection**: Route-level authentication
3. **Input Validation**: Parameter validation at controller layer
4. **SQL Injection Prevention**: Parameterized queries via GORM
5. **CORS Protection**: Configurable cross-origin policies

### Performance Considerations

1. **Connection Pooling**: Database connection pools for efficiency
2. **Caching**: Redis integration for caching frequently accessed data
3. **Asynchronous Processing**: Background job processing via workers
4. **Structured Logging**: High-performance logging with Zap
5. **Graceful Shutdown**: Proper resource cleanup on shutdown
