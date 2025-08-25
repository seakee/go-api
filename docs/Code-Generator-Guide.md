# Code Generator Guide | 代码生成器指南

[English](#english) | [中文](#中文)

---

## English

### Overview

The Go-API framework includes a powerful code generator that automatically creates Go models and repositories from MySQL CREATE TABLE statements. This tool significantly speeds up development by eliminating boilerplate code and ensuring consistency across your data layer.

### Features

- **SQL Parsing**: Supports complex field type definitions (VARCHAR(255), DECIMAL(10,2), etc.)
- **Type Mapping**: Comprehensive mapping from MySQL types to Go types
- **GORM Integration**: Automatic generation of GORM tags (column, size, not null, default, etc.)
- **Comment Support**: Preserves SQL comments as Go struct field comments
- **Repository Generation**: Automatic generation of repository interfaces and implementations
- **Smart Update Logic**: Generates update methods that only update non-zero fields
- **Flexible List Methods**: List methods that accept struct parameters for query conditions
- **Batch Processing**: Process single files or entire directories
- **Force Overwrite**: Option to overwrite existing files

### Installation and Usage

#### Building the Generator

```bash
# Build the code generator
go build -o codegen ./command/codegen/handler.go
```

#### Command Line Options

```bash
./codegen [options]
```

**Available Options:**

- `-force`: Force overwrite existing files (default: false)
- `-name`: SQL file name (without .sql extension) to generate code for
- `-sql`: SQL directory path (default: "bin/data/sql")
- `-model`: Model output directory (default: "app/model")
- `-repo`: Repository output directory (default: "app/repository")
- `-service`: Service output directory (default: "app/service")

#### Usage Examples

**Generate code for a single SQL file:**
```bash
./codegen -name auth_app
```

**Generate code for all SQL files in directory:**
```bash
./codegen
```

**Use custom paths:**
```bash
./codegen -sql custom/sql/path -model custom/model/path -force
```

**Generate with all options:**
```bash
./codegen -name users -sql bin/data/sql -model app/model -repo app/repository -service app/service -force
```

### SQL File Format

The generator expects standard MySQL CREATE TABLE statements. Here's a comprehensive example:

```sql
CREATE TABLE `auth_app`
(
    `id`           int                                                           NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `app_id`       varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL COMMENT 'Application ID',
    `app_name`     varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci           DEFAULT NULL COMMENT 'Application Name',
    `app_secret`   varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Application Secret',
    `redirect_uri` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci          DEFAULT NULL COMMENT 'Redirect URI',
    `description`  text COMMENT 'Description',
    `status`       tinyint(1)                                                    NOT NULL DEFAULT '0' COMMENT 'Status: 0=Inactive, 1=Active, 2=Disabled',
    `created_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp                                                     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_app_id` (`app_id`),
    KEY `idx_status` (`status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Application Information Table';
```

### Type Mapping

The generator maps MySQL types to appropriate Go types:

| MySQL Type | Go Type | GORM Tag Example |
|------------|---------|------------------|
| `int`, `integer` | `int` | `gorm:"column:id;not null"` |
| `tinyint` | `int8` | `gorm:"column:status;not null;default:0"` |
| `smallint` | `int16` | `gorm:"column:count;not null"` |
| `bigint` | `int64` | `gorm:"column:large_id;not null"` |
| `varchar(n)`, `char(n)` | `string` | `gorm:"column:name;size:50;not null"` |
| `text`, `longtext` | `string` | `gorm:"column:description;type:text"` |
| `decimal(m,n)`, `numeric(m,n)` | `float64` | `gorm:"column:price;type:decimal(10,2)"` |
| `float` | `float32` | `gorm:"column:rate;type:float"` |
| `double` | `float64` | `gorm:"column:amount;type:double"` |
| `timestamp`, `datetime` | `time.Time` | `gorm:"column:created_at"` |
| `date` | `time.Time` | `gorm:"column:birth_date;type:date"` |
| `json` | `string` | `gorm:"column:data;type:json"` |
| `blob` | `[]byte` | `gorm:"column:content;type:blob"` |

### Generated Output

#### Model Files

The generator creates comprehensive model files with:

1. **Proper Go naming conventions** (snake_case → CamelCase)
2. **GORM tags** for database mapping
3. **JSON tags** for API serialization
4. **Field comments** from SQL comments
5. **Common database methods** (CRUD operations)

**Example Generated Model:**

```go
package auth

import (
    "context"
    "errors"
    "fmt"
    "gorm.io/gorm"
)

type App struct {
    gorm.Model

    AppId       string `gorm:"column:app_id;size:30;not null" json:"app_id"`                  // Application ID
    AppName     string `gorm:"column:app_name;size:50" json:"app_name"`                       // Application Name
    AppSecret   string `gorm:"column:app_secret;size:256;not null" json:"app_secret"`         // Application Secret
    RedirectUri string `gorm:"column:redirect_uri;size:500" json:"redirect_uri"`              // Redirect URI
    Description string `gorm:"column:description;type:text" json:"description"`               // Description
    Status      int8   `gorm:"column:status;not null;default:0" json:"status"`                // Status: 0=Inactive, 1=Active, 2=Disabled

    // Query conditions for chaining methods
    queryCondition interface{}   `gorm:"-" json:"-"`
    queryArgs      []interface{} `gorm:"-" json:"-"`
}

// TableName specifies the table name for the App model.
func (a *App) TableName() string {
    return "auth_app"
}

// Where sets query conditions for chaining with other methods.
func (a *App) Where(query interface{}, args ...interface{}) *App {
    newApp := *a
    newApp.queryCondition = query
    newApp.queryArgs = args
    return &newApp
}

// First retrieves the first app matching the criteria from the database.
func (a *App) First(ctx context.Context, db *gorm.DB) (*App, error) {
    var app App

    query := db.WithContext(ctx)

    // Apply Where conditions if set
    if a.queryCondition != nil {
        query = query.Where(a.queryCondition, a.queryArgs...)
    } else {
        query = query.Where(a)
    }

    if err := query.First(&app).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil
        }
        return nil, fmt.Errorf("find first failed: %w", err)
    }

    return &app, nil
}

// Create inserts a new app into the database and returns the ID.
func (a *App) Create(ctx context.Context, db *gorm.DB) (uint, error) {
    if err := db.WithContext(ctx).Create(a).Error; err != nil {
        return 0, fmt.Errorf("create failed: %w", err)
    }
    return a.ID, nil
}

// Updates applies the specified updates to the app in the database.
func (a *App) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
    query := db.WithContext(ctx).Model(&App{})

    if a.queryCondition != nil {
        query = query.Where(a.queryCondition, a.queryArgs...)
    } else if a.ID > 0 {
        query = query.Where("id = ?", a.ID)
    } else {
        query = query.Where(a)
    }

    return query.Updates(updates).Error
}

// Delete removes the app from the database.
func (a *App) Delete(ctx context.Context, db *gorm.DB) error {
    query := db.WithContext(ctx)

    if a.queryCondition != nil {
        query = query.Where(a.queryCondition, a.queryArgs...)
    } else {
        query = query.Where(a)
    }

    return query.Delete(&App{}).Error
}

// List retrieves all apps matching the criteria from the database.
func (a *App) List(ctx context.Context, db *gorm.DB) ([]App, error) {
    var apps []App

    query := db.WithContext(ctx)

    if a.queryCondition != nil {
        query = query.Where(a.queryCondition, a.queryArgs...)
    } else {
        query = query.Where(a)
    }

    if err := query.Find(&apps).Error; err != nil {
        return nil, fmt.Errorf("list failed: %w", err)
    }

    return apps, nil
}

// Additional utility methods...
```

#### Repository Files

The generator creates repository interfaces and implementations:

**Repository Interface:**

```go
package auth

import (
    "context"
    "github.com/seakee/go-api/app/model/auth"
    "github.com/sk-pkg/redis"
    "gorm.io/gorm"
)

// AppRepo defines the interface for app-related database operations.
type AppRepo interface {
    // GetApp retrieves a app by its properties.
    GetApp(ctx context.Context, app *auth.App) (*auth.App, error)

    // Create inserts a new app into the database.
    Create(ctx context.Context, app *auth.App) (uint, error)

    // Update updates an existing app in the database.
    Update(ctx context.Context, id uint, app *auth.App) error

    // Delete deletes a app by its ID.
    Delete(ctx context.Context, id uint) error

    // List retrieves app records based on query conditions.
    List(ctx context.Context, app *auth.App) ([]auth.App, error)

    // GetByID retrieves a app by its ID.
    GetByID(ctx context.Context, id uint) (*auth.App, error)
}
```

**Repository Implementation:**

```go
// appRepo implements the AppRepo interface.
type appRepo struct {
    redis *redis.Manager
    db    *gorm.DB
}

// NewAppRepo creates a new instance of the app repository.
func NewAppRepo(db *gorm.DB, redis *redis.Manager) AppRepo {
    return &appRepo{redis: redis, db: db}
}

// Create creates a new app record using the model's Create method.
func (r *appRepo) Create(ctx context.Context, app *auth.App) (uint, error) {
    return app.Create(ctx, r.db)
}

// Update updates an existing app record with smart field detection.
func (r *appRepo) Update(ctx context.Context, id uint, app *auth.App) error {
    data := make(map[string]interface{})

    if app.AppId != "" {
        data["app_id"] = app.AppId
    }
    if app.AppName != "" {
        data["app_name"] = app.AppName
    }
    if app.AppSecret != "" {
        data["app_secret"] = app.AppSecret
    }
    if app.RedirectUri != "" {
        data["redirect_uri"] = app.RedirectUri
    }
    if app.Description != "" {
        data["description"] = app.Description
    }
    if app.Status != 0 {
        data["status"] = app.Status
    }

    if len(data) == 0 {
        return nil // No fields to update
    }

    updateModel := &auth.App{}
    updateModel.ID = id

    return updateModel.Updates(ctx, r.db, data)
}

// GetByID retrieves a app by its ID.
func (r *appRepo) GetByID(ctx context.Context, id uint) (*auth.App, error) {
    app := &auth.App{}
    return app.Where("id = ?", id).First(ctx, r.db)
}

// List retrieves app records based on query conditions.
func (r *appRepo) List(ctx context.Context, app *auth.App) ([]auth.App, error) {
    return app.List(ctx, r.db)
}

// Additional methods...
```

### Advanced Features

#### 1. Custom Templates

You can customize the generated code by modifying the templates in `command/codegen/codegen/`:

- `model.go`: Contains model generation templates
- `repository.go`: Contains repository generation templates

#### 2. Naming Conventions

The generator follows these naming conventions:

- **Snake case to Camel case**: `user_name` → `UserName`
- **Table names**: Preserved as-is for `TableName()` method
- **Package names**: Derived from directory structure
- **File names**: Follow Go conventions (lowercase with underscores)

#### 3. GORM Tag Generation

The generator automatically creates appropriate GORM tags:

```go
// VARCHAR(50) NOT NULL → 
`gorm:"column:username;size:50;not null"`

// INT DEFAULT 1 →
`gorm:"column:status;not null;default:1"`

// TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP →
`gorm:"column:created_at"`
```

#### 4. Comment Preservation

SQL comments are converted to Go comments:

```sql
`username` varchar(50) NOT NULL COMMENT 'User login name'
```

Becomes:

```go
Username string `gorm:"column:username;size:50;not null" json:"username"` // User login name
```

### Integration with Framework

#### 1. Using Generated Models

```go
// Create a new app
app := &auth.App{
    AppId:     "app123",
    AppName:   "My Application",
    AppSecret: "secret123",
    Status:    1,
}

id, err := app.Create(ctx, db)
if err != nil {
    log.Fatal(err)
}

// Query apps
searchApp := &auth.App{Status: 1}
apps, err := searchApp.List(ctx, db)
if err != nil {
    log.Fatal(err)
}

// Update with Where condition
updateApp := &auth.App{}
updates := map[string]interface{}{
    "app_name": "Updated Name",
    "status":   2,
}
err = updateApp.Where("app_id = ?", "app123").Updates(ctx, db, updates)
```

#### 2. Using Generated Repositories

```go
// Initialize repository
repo := auth.NewAppRepo(db, redisManager)

// Use repository methods
app := &auth.App{
    AppName: "New App",
    Status:  1,
}

id, err := repo.Create(ctx, app)
if err != nil {
    log.Fatal(err)
}

// Get by ID
foundApp, err := repo.GetByID(ctx, id)
if err != nil {
    log.Fatal(err)
}
```

### Best Practices

#### 1. SQL File Organization

Organize your SQL files by feature or module:

```
bin/data/sql/
├── auth/
│   ├── auth_app.sql
│   └── auth_user.sql
├── user/
│   ├── user_profile.sql
│   └── user_setting.sql
└── order/
    ├── order.sql
    └── order_item.sql
```

#### 2. Regeneration Strategy

- Use version control to track generated code changes
- Run generation in CI/CD pipeline
- Use `-force` flag carefully in production

#### 3. Customization

- Extend generated models with additional methods
- Use composition for complex business logic
- Keep generated code separate from custom code

### Troubleshooting

#### Common Issues

1. **Parse Error**: Check SQL syntax, especially CREATE TABLE statement
2. **Type Mapping**: Verify supported MySQL types
3. **File Permissions**: Ensure write permissions to output directories
4. **Import Conflicts**: Check for naming conflicts with existing packages

#### Debug Mode

Run generator with verbose output:

```bash
./codegen -name table_name -v
```

#### Manual Fixes

If automatic generation fails, you can:

1. Fix SQL syntax issues
2. Add missing table comments
3. Ensure proper field types
4. Check character encoding

---

## 中文

### 概述

Go-API框架包含一个强大的代码生成器，可以从MySQL CREATE TABLE语句自动创建Go模型和仓库。这个工具通过消除样板代码并确保数据层的一致性，显著加快了开发速度。

### 特性

- **SQL解析**: 支持复杂的字段类型定义（VARCHAR(255)、DECIMAL(10,2)等）
- **类型映射**: MySQL类型到Go类型的全面映射
- **GORM集成**: 自动生成GORM标签（column、size、not null、default等）
- **注释支持**: 将SQL注释保存为Go结构体字段注释
- **仓库生成**: 自动生成仓库接口和实现
- **智能更新逻辑**: 生成只更新非零值字段的更新方法
- **灵活的列表方法**: 接受结构体参数进行查询条件的列表方法
- **批处理**: 处理单个文件或整个目录
- **强制覆盖**: 覆盖现有文件的选项

### 安装和使用

#### 构建生成器

```bash
# 构建代码生成器
go build -o codegen ./command/codegen/handler.go
```

#### 命令行选项

```bash
./codegen [选项]
```

**可用选项：**

- `-force`: 强制覆盖现有文件（默认：false）
- `-name`: 要生成代码的SQL文件名（不含.sql扩展名）
- `-sql`: SQL目录路径（默认："bin/data/sql"）
- `-model`: 模型输出目录（默认："app/model"）
- `-repo`: 仓库输出目录（默认："app/repository"）
- `-service`: 服务输出目录（默认："app/service"）

#### 使用示例

**为单个SQL文件生成代码：**
```bash
./codegen -name auth_app
```

**为目录中的所有SQL文件生成代码：**
```bash
./codegen
```

**使用自定义路径：**
```bash
./codegen -sql custom/sql/path -model custom/model/path -force
```

### SQL文件格式

生成器期望标准的MySQL CREATE TABLE语句。这是一个综合示例：

```sql
CREATE TABLE `auth_app`
(
    `id`           int                                                           NOT NULL AUTO_INCREMENT COMMENT '主键',
    `app_id`       varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL COMMENT '应用ID',
    `app_name`     varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci           DEFAULT NULL COMMENT '应用名称',
    `app_secret`   varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '应用密钥',
    `redirect_uri` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci          DEFAULT NULL COMMENT '重定向URI',
    `description`  text COMMENT '描述',
    `status`       tinyint(1)                                                    NOT NULL DEFAULT '0' COMMENT '状态：0=未激活，1=激活，2=禁用',
    `created_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp                                                     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_app_id` (`app_id`),
    KEY `idx_status` (`status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '应用信息表';
```

### 类型映射

生成器将MySQL类型映射到适当的Go类型：

| MySQL类型 | Go类型 | GORM标签示例 |
|-----------|--------|-------------|
| `int`, `integer` | `int` | `gorm:"column:id;not null"` |
| `tinyint` | `int8` | `gorm:"column:status;not null;default:0"` |
| `smallint` | `int16` | `gorm:"column:count;not null"` |
| `bigint` | `int64` | `gorm:"column:large_id;not null"` |
| `varchar(n)`, `char(n)` | `string` | `gorm:"column:name;size:50;not null"` |
| `text`, `longtext` | `string` | `gorm:"column:description;type:text"` |
| `decimal(m,n)`, `numeric(m,n)` | `float64` | `gorm:"column:price;type:decimal(10,2)"` |
| `float` | `float32` | `gorm:"column:rate;type:float"` |
| `double` | `float64` | `gorm:"column:amount;type:double"` |
| `timestamp`, `datetime` | `time.Time` | `gorm:"column:created_at"` |
| `date` | `time.Time` | `gorm:"column:birth_date;type:date"` |
| `json` | `string` | `gorm:"column:data;type:json"` |
| `blob` | `[]byte` | `gorm:"column:content;type:blob"` |

### 生成的输出

生成器创建包含以下内容的综合模型文件：

1. **正确的Go命名约定**（snake_case → CamelCase）
2. **GORM标签**用于数据库映射
3. **JSON标签**用于API序列化
4. **来自SQL注释的字段注释**
5. **常见的数据库方法**（CRUD操作）

### 高级特性

#### 1. 自定义模板

您可以通过修改`command/codegen/codegen/`中的模板来自定义生成的代码：

- `model.go`: 包含模型生成模板
- `repository.go`: 包含仓库生成模板

#### 2. 命名约定

生成器遵循这些命名约定：

- **下划线转驼峰**: `user_name` → `UserName`
- **表名**: 在`TableName()`方法中保持原样
- **包名**: 从目录结构派生
- **文件名**: 遵循Go约定（小写带下划线）

#### 3. GORM标签生成

生成器自动创建适当的GORM标签：

```go
// VARCHAR(50) NOT NULL → 
`gorm:"column:username;size:50;not null"`

// INT DEFAULT 1 →
`gorm:"column:status;not null;default:1"`

// TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP →
`gorm:"column:created_at"`
```

### 与框架集成

#### 1. 使用生成的模型

```go
// 创建新应用
app := &auth.App{
    AppId:     "app123",
    AppName:   "我的应用",
    AppSecret: "secret123",
    Status:    1,
}

id, err := app.Create(ctx, db)
if err != nil {
    log.Fatal(err)
}

// 查询应用
searchApp := &auth.App{Status: 1}
apps, err := searchApp.List(ctx, db)
if err != nil {
    log.Fatal(err)
}
```

#### 2. 使用生成的仓库

```go
// 初始化仓库
repo := auth.NewAppRepo(db, redisManager)

// 使用仓库方法
app := &auth.App{
    AppName: "新应用",
    Status:  1,
}

id, err := repo.Create(ctx, app)
if err != nil {
    log.Fatal(err)
}
```

### 最佳实践

#### 1. SQL文件组织

按功能或模块组织SQL文件：

```
bin/data/sql/
├── auth/
│   ├── auth_app.sql
│   └── auth_user.sql
├── user/
│   ├── user_profile.sql
│   └── user_setting.sql
└── order/
    ├── order.sql
    └── order_item.sql
```

#### 2. 重新生成策略

- 使用版本控制跟踪生成的代码更改
- 在CI/CD管道中运行生成
- 在生产环境中谨慎使用`-force`标志

### 故障排除

#### 常见问题

1. **解析错误**: 检查SQL语法，特别是CREATE TABLE语句
2. **类型映射**: 验证支持的MySQL类型
3. **文件权限**: 确保对输出目录有写权限
4. **导入冲突**: 检查与现有包的命名冲突

#### 调试模式

使用详细输出运行生成器：

```bash
./codegen -name table_name -v
```