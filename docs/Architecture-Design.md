# Architecture Design | 架构设计

[English](#english) | [中文](#中文)

---

## English

### Overview

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
        service: authService.NewAppService(appCtx.MysqlDB["go-api"], appCtx.Redis["go-api"]),
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

---

## 中文

### 概述

Go-API框架遵循干净的分层架构模式，专为企业级应用设计。它强调关注点分离、依赖注入和可维护性。

### 架构层次

```
┌─────────────────────────────────────┐
│           表现层                    │
│        (HTTP 控制器)                │
├─────────────────────────────────────┤
│           业务层                    │
│          (服务层)                   │
├─────────────────────────────────────┤
│         数据访问层                  │
│         (仓库层)                    │
├─────────────────────────────────────┤
│            数据层                   │
│      (模型和数据库)                 │
└─────────────────────────────────────┘
```

#### 1. 表现层（控制器）

**目的**: 处理HTTP请求和响应
**位置**: `app/http/controller/`

- 接收客户端的HTTP请求
- 验证输入参数
- 调用适当的服务方法
- 格式化并返回HTTP响应
- 通过中间件处理认证和授权

#### 2. 业务层（服务）

**目的**: 实现业务逻辑和规则
**位置**: `app/service/`

- 包含所有业务逻辑
- 协调多个仓库之间的操作
- 实现复杂的业务规则和验证
- 处理事务和数据一致性
- 独立于HTTP关注点

#### 3. 数据访问层（仓库）

**目的**: 抽象数据访问操作
**位置**: `app/repository/`

- 提供数据库操作的抽象
- 实现数据访问模式
- 处理数据库连接和查询
- 在领域模型和数据库实体之间转换
- 支持多种数据库类型（MySQL、MongoDB）

#### 4. 数据层（模型）

**目的**: 定义数据结构和基本操作
**位置**: `app/model/`

- 定义数据结构（结构体）
- 包含基本的CRUD操作
- 包括数据库模式定义（GORM标签）
- 处理数据验证和序列化
- 支持SQL和NoSQL数据库

### 依赖注入模式

框架使用基于构造函数的依赖注入：

```go
// 服务层只接收它需要的依赖
func NewAppService(db *gorm.DB, redis *redis.Manager) AppService {
    repo := authRepo.NewAppRepo(db, redis)
    return &appService{repo: repo}
}

// 控制器层将依赖注入到服务中
func NewHandler(appCtx *http.Context) Handler {
    return &handler{
        service: authService.NewAppService(appCtx.MysqlDB["go-api"], appCtx.Redis["go-api"]),
    }
}
```

### 中间件架构

```
请求 → SetTraceID → CORS → RequestLogger → CheckAppAuth → 控制器
```

中间件组件：
- **SetTraceID**: 为请求跟踪生成唯一的跟踪ID
- **CORS**: 处理跨域请求
- **RequestLogger**: 记录请求详情用于监控
- **CheckAppAuth**: 验证JWT令牌进行身份认证

### 配置管理

支持热重载的多环境配置：

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

环境特定文件：
- `local.json` - 本地开发
- `dev.json` - 开发环境
- `prod.json` - 生产环境

### 错误处理策略

1. **错误传播**: 错误在各层之间向上冒泡
2. **集中处理**: 控制器处理所有错误响应
3. **国际化**: 错误消息支持多种语言
4. **结构化日志**: 所有错误都记录上下文

### 安全架构

1. **JWT认证**: 基于令牌的认证系统
2. **中间件保护**: 路由级别的认证
3. **输入验证**: 控制器层的参数验证
4. **SQL注入防护**: 通过GORM使用参数化查询
5. **CORS保护**: 可配置的跨域策略

### 性能考虑

1. **连接池**: 数据库连接池提高效率
2. **缓存**: Redis集成用于缓存频繁访问的数据
3. **异步处理**: 通过工作者进行后台作业处理
4. **结构化日志**: 使用Zap进行高性能日志记录
5. **优雅关闭**: 关闭时正确清理资源