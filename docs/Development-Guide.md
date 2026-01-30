# Development Guide | 开发指南

[English](#english) | [中文](#中文)

---

## English

### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose (optional)
- MySQL 8.0+ or MongoDB 4.4+
- Redis 6.0+
- Make (for build automation)

### Project Setup

#### 1. Local Development Setup

```bash
# Clone the repository
git clone https://github.com/seakee/go-api.git
cd go-api

# Install dependencies
go mod download

# Copy configuration template
cp bin/configs/local.json.default bin/configs/local.json

# Edit configuration with your database settings
vim bin/configs/local.json

# Run the application
make run
```

#### 2. Using Project Generator

```bash
# Download and use the project generator
curl -O https://raw.githubusercontent.com/seakee/go-api/main/scripts/generate.sh
chmod +x generate.sh

# Create a new project
./generate.sh my-new-api v1.0.0
cd my-new-api

# Start development
go mod tidy
make run
```

### Configuration

#### Environment Configuration

Create environment-specific configuration files:

```json
{
  "system": {
    "name": "go-api",
    "run_mode": "debug",
    "http_port": ":8080",
    "read_timeout": 60,
    "write_timeout": 60,
    "jwt_secret": "your-secret-key",
    "default_lang": "zh-CN"
  },
  "log": {
    "driver": "stdout",
    "level": "debug",
    "path": "storage/logs/"
  },
  "databases": [
    {
      "enable": true,
      "db_type": "mysql",
      "db_name": "go-api",
      "db_host": "localhost",
      "db_port": 3306,
      "db_username": "root",
      "db_password": "password",
      "charset": "utf8mb4",
      "db_max_idle_conn": 10,
      "db_max_open_conn": 50,
      "conn_max_lifetime": 3,
      "conn_max_idle_time": 1
    }
  ],
  "redis": [
    {
      "enable": true,
      "name": "go-api",
      "host": "localhost:6379",
      "auth": "",
      "max_idle": 30,
      "max_active": 100,
      "idle_timeout": 30,
      "prefix": "go-api",
      "db": 0
    }
  ]
}
```

### Development Workflow

#### 1. Creating a New Feature

Follow these steps to add a new feature:

1. **Create Model** (if needed)
2. **Create Repository Interface and Implementation**
3. **Create Service Interface and Implementation**
4. **Create Controller**
5. **Register Routes**
6. **Add Tests**

#### 2. Adding a New Model

```bash
# Option 1: Use code generator
go run ./command/codegen/handler.go -name user_table

# Option 2: Manual creation
# Create app/model/user/user.go
```

Example model structure:
```go
package user

import (
    "context"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username string `gorm:"column:username;size:50;not null" json:"username"`
    Email    string `gorm:"column:email;size:100;not null" json:"email"`
    Status   int8   `gorm:"column:status;not null;default:1" json:"status"`
}

func (u *User) TableName() string {
    return "users"
}

// Add CRUD methods here...
```

#### 3. Creating Repository Layer

```go
// app/repository/user/user.go
package user

import (
    "context"
    "github.com/seakee/go-api/app/model/user"
    "github.com/sk-pkg/redis"
    "gorm.io/gorm"
)

type UserRepo interface {
    Create(ctx context.Context, user *user.User) (uint, error)
    GetByID(ctx context.Context, id uint) (*user.User, error)
    Update(ctx context.Context, id uint, user *user.User) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, user *user.User) ([]user.User, error)
}

type userRepo struct {
    db    *gorm.DB
    redis *redis.Manager
}

func NewUserRepo(db *gorm.DB, redis *redis.Manager) UserRepo {
    return &userRepo{db: db, redis: redis}
}

func (r *userRepo) Create(ctx context.Context, user *user.User) (uint, error) {
    return user.Create(ctx, r.db)
}

// Implement other methods...
```

#### 4. Creating Service Layer

```go
// app/service/user/user.go
package user

import (
    "context"
    "github.com/seakee/go-api/app/model/user"
    userRepo "github.com/seakee/go-api/app/repository/user"
    "github.com/sk-pkg/redis"
    "gorm.io/gorm"
)

type UserService interface {
    CreateUser(ctx context.Context, params *CreateUserParams) (*CreateUserResult, error)
    GetUserByID(ctx context.Context, id uint) (*user.User, error)
    UpdateUser(ctx context.Context, id uint, params *UpdateUserParams) error
    DeleteUser(ctx context.Context, id uint) error
    ListUsers(ctx context.Context, params *ListUserParams) ([]user.User, error)
}

type CreateUserParams struct {
    Username string `json:"username" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
}

type CreateUserResult struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

type userService struct {
    repo userRepo.UserRepo
}

func NewUserService(db *gorm.DB, redis *redis.Manager) UserService {
    repo := userRepo.NewUserRepo(db, redis)
    return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, params *CreateUserParams) (*CreateUserResult, error) {
    // Business logic implementation
    user := &user.User{
        Username: params.Username,
        Email:    params.Email,
        Status:   1,
    }
    
    id, err := s.repo.Create(ctx, user)
    if err != nil {
        return nil, err
    }
    
    return &CreateUserResult{
        ID:       id,
        Username: user.Username,
        Email:    user.Email,
    }, nil
}

// Implement other methods...
```

#### 5. Creating Controller Layer

```go
// app/http/controller/user/handler.go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/seakee/go-api/app/http"
    "github.com/seakee/go-api/app/http/controller"
    userService "github.com/seakee/go-api/app/service/user"
)

type Handler interface {
    Create() gin.HandlerFunc
    GetByID() gin.HandlerFunc
    Update() gin.HandlerFunc
    Delete() gin.HandlerFunc
    List() gin.HandlerFunc
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
        service: userService.NewUserService(appCtx.MysqlDB["go-api"], appCtx.Redis["go-api"]),
    }
}

func (h handler) Create() gin.HandlerFunc {
    return func(c *gin.Context) {
        var params userService.CreateUserParams
        
        if err := c.ShouldBindJSON(&params); err != nil {
            h.I18n.JSON(c, e.InvalidParams, nil, err)
            return
        }
        
        result, err := h.service.CreateUser(h.Context(c), &params)
        if err != nil {
            h.I18n.JSON(c, e.ServerError, nil, err)
            return
        }
        
        h.I18n.JSON(c, e.SUCCESS, result, nil)
    }
}

// Implement other methods...
```

#### 6. Registering Routes

```go
// app/http/router/external/service/user/user.go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/seakee/go-api/app/http"
    "github.com/seakee/go-api/app/http/controller/user"
)

func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
    userHandler := user.NewHandler(ctx)
    {
        api.POST("user", ctx.Middleware.CheckAppAuth(), userHandler.Create())
        api.GET("user/:id", userHandler.GetByID())
        api.PUT("user/:id", ctx.Middleware.CheckAppAuth(), userHandler.Update())
        api.DELETE("user/:id", ctx.Middleware.CheckAppAuth(), userHandler.Delete())
        api.GET("users", userHandler.List())
    }
}
```

Then register in the main router:
```go
// app/http/router/external/handler.go
func RegisterRoutes(api *gin.RouterGroup, ctx *http.Context) {
    // ... existing routes

    // Add user routes
    userGroup := api.Group("user")
    user.RegisterRoutes(userGroup, ctx)
}
```

### Middleware Development

#### Creating Custom Middleware

1. **Define Interface Method**:
```go
// app/http/middleware/handler.go
type Middleware interface {
    CheckAppAuth() gin.HandlerFunc
    // Add your new middleware
    RateLimiter() gin.HandlerFunc
}
```

2. **Implement Middleware**:
```go
// app/http/middleware/rate_limiter.go
func (m middleware) RateLimiter() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Rate limiting logic
        clientIP := c.ClientIP()
        
        // Check rate limit using Redis
        key := fmt.Sprintf("rate_limit:%s", clientIP)
        count, err := m.redis["go-api"].Incr(key)
        if err != nil {
            c.Next()
            return
        }
        
        if count == 1 {
            m.redis["go-api"].Expire(key, 60) // 1 minute window
        }
        
        if count > 100 { // 100 requests per minute
            m.i18n.JSON(c, e.TooManyRequests, nil, nil)
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Code Generation

#### Using the Built-in Generator

The framework includes a powerful code generator that creates models and repositories from SQL files.

1. **Create SQL File**:
```sql
-- bin/data/sql/users.sql
CREATE TABLE `users` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `username` varchar(50) NOT NULL COMMENT 'Username',
    `email` varchar(100) NOT NULL COMMENT 'Email Address',
    `password_hash` varchar(255) NOT NULL COMMENT 'Password Hash',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT 'Status: 0=Inactive, 1=Active',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_username` (`username`),
    UNIQUE KEY `unique_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User Information Table';
```

2. **Generate Code**:
```bash
# Generate from specific SQL file
go run ./command/codegen/handler.go -name users

# Generate from all SQL files
go run ./command/codegen/handler.go

# Generate with custom paths
go run ./command/codegen/handler.go -sql custom/sql -model custom/model -repo custom/repo
```

### Testing

#### Unit Testing

Create tests for each layer:

```go
// app/service/user/user_test.go
package user

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockUserRepo struct {
    mock.Mock
}

func (m *MockUserRepo) Create(ctx context.Context, user *user.User) (uint, error) {
    args := m.Called(ctx, user)
    return args.Get(0).(uint), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
    // Test implementation
    mockRepo := new(MockUserRepo)
    service := &userService{repo: mockRepo}
    
    params := &CreateUserParams{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(uint(1), nil)
    
    result, err := service.CreateUser(context.Background(), params)
    
    assert.NoError(t, err)
    assert.Equal(t, uint(1), result.ID)
    assert.Equal(t, "testuser", result.Username)
    mockRepo.AssertExpectations(t)
}
```

#### Integration Testing

```go
// app/http/controller/user/user_test.go
package user

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    // Setup test environment
    router := gin.New()
    // Add your handler setup here
    
    params := map[string]interface{}{
        "username": "testuser",
        "email":    "test@example.com",
    }
    
    jsonData, _ := json.Marshal(params)
    req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Build and Deployment

#### Build Commands

```bash
# Development
make run          # Run application locally
make test         # Run all tests
make fmt          # Format code
make lint         # Run linter
make all          # fmt + test + build

# Production
make build        # Build binary
make docker-build # Build Docker image
make docker-run   # Run Docker container
```

#### Docker Deployment

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
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

---

## 中文

### 前提条件

- Go 1.22 或更高版本
- Docker 和 Docker Compose（可选）
- MySQL 8.0+ 或 MongoDB 4.4+
- Redis 6.0+
- Make（用于构建自动化）

### 项目设置

#### 1. 本地开发设置

```bash
# 克隆仓库
git clone https://github.com/seakee/go-api.git
cd go-api

# 安装依赖
go mod download

# 复制配置模板
cp bin/configs/local.json.default bin/configs/local.json

# 编辑配置文件，设置数据库信息
vim bin/configs/local.json

# 运行应用
make run
```

#### 2. 使用项目生成器

```bash
# 下载并使用项目生成器
curl -O https://raw.githubusercontent.com/seakee/go-api/main/scripts/generate.sh
chmod +x generate.sh

# 创建新项目
./generate.sh my-new-api v1.0.0
cd my-new-api

# 开始开发
go mod tidy
make run
```

### 开发工作流程

#### 1. 创建新功能

按以下步骤添加新功能：

1. **创建模型**（如果需要）
2. **创建仓库接口和实现**
3. **创建服务接口和实现**
4. **创建控制器**
5. **注册路由**
6. **添加测试**

#### 2. 添加新模型

```bash
# 选项1：使用代码生成器
go run ./command/codegen/handler.go -name user_table

# 选项2：手动创建
# 创建 app/model/user/user.go
```

#### 3. 中间件开发

创建自定义中间件：

1. **定义接口方法**:
```go
// app/http/middleware/handler.go
type Middleware interface {
    CheckAppAuth() gin.HandlerFunc
    // 添加新中间件
    RateLimiter() gin.HandlerFunc
}
```

2. **实现中间件**:
```go
// app/http/middleware/rate_limiter.go
func (m middleware) RateLimiter() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 限流逻辑
        clientIP := c.ClientIP()
        
        // 使用Redis检查限流
        key := fmt.Sprintf("rate_limit:%s", clientIP)
        count, err := m.redis["go-api"].Incr(key)
        if err != nil {
            c.Next()
            return
        }
        
        if count == 1 {
            m.redis["go-api"].Expire(key, 60) // 1分钟窗口
        }
        
        if count > 100 { // 每分钟100个请求
            m.i18n.JSON(c, e.TooManyRequests, nil, nil)
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 代码生成

#### 使用内置生成器

框架包含一个强大的代码生成器，可以从SQL文件创建模型和仓库。

1. **创建SQL文件**:
```sql
-- bin/data/sql/users.sql
CREATE TABLE `users` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT '主键',
    `username` varchar(50) NOT NULL COMMENT '用户名',
    `email` varchar(100) NOT NULL COMMENT '邮箱地址',
    `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
    `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态: 0=禁用, 1=启用',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_username` (`username`),
    UNIQUE KEY `unique_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';
```

2. **生成代码**:
```bash
# 从特定SQL文件生成
go run ./command/codegen/handler.go -name users

# 从所有SQL文件生成
go run ./command/codegen/handler.go

# 使用自定义路径生成
go run ./command/codegen/handler.go -sql custom/sql -model custom/model -repo custom/repo
```

### 测试

#### 单元测试

为每层创建测试：

```go
// app/service/user/user_test.go
func TestUserService_CreateUser(t *testing.T) {
    // 测试实现
    mockRepo := new(MockUserRepo)
    service := &userService{repo: mockRepo}
    
    params := &CreateUserParams{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(uint(1), nil)
    
    result, err := service.CreateUser(context.Background(), params)
    
    assert.NoError(t, err)
    assert.Equal(t, uint(1), result.ID)
    assert.Equal(t, "testuser", result.Username)
    mockRepo.AssertExpectations(t)
}
```

### 构建和部署

#### 构建命令

```bash
# 开发
make run          # 本地运行应用
make test         # 运行所有测试
make fmt          # 格式化代码
make lint         # 运行代码检查
make all          # fmt + test + build

# 生产
make build        # 构建二进制文件
make docker-build # 构建Docker镜像
make docker-run   # 运行Docker容器
```