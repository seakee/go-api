# Go API 代码生成器

[English](README.md) | 中文

---
### 概述

Go API 代码生成器是一个强大的工具，可以从 MySQL CREATE TABLE 语句自动生成 Go 模型和仓库代码。它解析 SQL 文件并创建相应的带有 GORM 标签的 Go 结构体，以及仓库接口和实现，用于常用的数据库操作。

### 功能特性

- **SQL 解析**: 支持复杂的字段类型定义（VARCHAR(255), DECIMAL(10,2) 等）
- **类型映射**: 从 MySQL 类型到 Go 类型的全面映射
- **GORM 集成**: 自动生成 GORM 标签（column, size, not null, default 等）
- **注释支持**: 保留 SQL 注释作为 Go 结构体字段注释
- **仓库生成**: 自动生成仓库接口和实现
- **智能更新逻辑**: 生成只更新非零值字段的更新方法
- **灵活的列表方法**: 接受结构体参数作为查询条件的列表方法
- **批量处理**: 处理单个文件或整个目录
- **强制覆盖**: 可选择覆盖现有文件

### 安装

```bash
# 构建代码生成器
go build -o codegen ./command/codegen/handler.go
```

### 使用方法

#### 命令行选项

```bash
./codegen [选项]
```

**可用选项：**

- `-force`: 强制覆盖现有文件（默认：false）
- `-name`: 要生成代码的 SQL 文件名（不含 .sql 扩展名）
- `-sql`: SQL 目录路径（默认："bin/data/sql"）
- `-model`: 模型输出目录（默认："app/model"）
- `-repo`: 仓库输出目录（默认："app/repository"）
- `-service`: 服务输出目录（默认："app/service"）

#### 使用示例

**为单个 SQL 文件生成代码：**
```bash
./codegen -name auth_app
```

**为目录中所有 SQL 文件生成代码：**
```bash
./codegen
```

**使用自定义路径：**
```bash
./codegen -sql custom/sql/path -model custom/model/path -force
```

### SQL 文件格式

生成器期望标准的 MySQL CREATE TABLE 语句。示例如下：

```sql
CREATE TABLE `auth_app`
(
    `id`           int                                                           NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `app_id`       varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL COMMENT '应用ID',
    `app_name`     varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci           DEFAULT NULL COMMENT '应用名称',
    `app_secret`   varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '应用密钥',
    `status`       tinyint(1)                                                    NOT NULL DEFAULT '0' COMMENT '状态：0=未激活，1=正常，2=禁用',
    `created_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp                                                     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `app_id` (`app_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = '应用信息表';
```

### 生成的输出

生成器创建：

**模型文件：**
1. **正确的 Go 命名规范**（snake_case → CamelCase）
2. **GORM 标签**用于数据库映射
3. **JSON 标签**用于 API 序列化
4. **字段注释**来自 SQL 注释
5. **常用数据库方法**（CRUD 操作）

**仓库文件：**
1. **仓库接口**包含标准 CRUD 操作
2. **仓库实现**支持依赖注入
3. **智能更新方法**只更新非零值字段
4. **灵活的列表方法**接受结构体参数进行过滤

**生成的模型代码示例：**

```go
package auth

import (
    "context"
    "gorm.io/gorm"
)

type App struct {
    gorm.Model
    
    AppID       string `gorm:"column:app_id;size:30;not null" json:"app_id"`           // 应用ID
    AppName     string `gorm:"column:app_name;size:50" json:"app_name"`                // 应用名称
    AppSecret   string `gorm:"column:app_secret;size:256;not null" json:"app_secret"`  // 应用密钥
    Status      int8   `gorm:"column:status;not null;default:0" json:"status"`         // 状态：0=未激活，1=正常，2=禁用
}

func (a *App) TableName() string {
    return "auth_app"
}

// 数据库操作方法...
func (a *App) First(ctx context.Context, db *gorm.DB) (*App, error) { /* ... */ }
func (a *App) Create(ctx context.Context, db *gorm.DB) (uint, error) { /* ... */ }
func (a *App) List(ctx context.Context, db *gorm.DB) ([]App, error) { /* ... */ }
// ... 更多方法
```

**生成的仓库代码示例：**

```go
package auth

import (
    "context"
    "github.com/seakee/go-api/app/model/auth"
    "github.com/sk-pkg/redis"
    "gorm.io/gorm"
)

// 仓库接口
type AppRepo interface {
    GetApp(ctx context.Context, app *auth.App) (*auth.App, error)
    Create(ctx context.Context, app *auth.App) (uint, error)
    Update(ctx context.Context, id uint, app *auth.App) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, app *auth.App) ([]auth.App, error)
    GetByID(ctx context.Context, id uint) (*auth.App, error)
}

// 仓库实现
type appRepo struct {
    redis *redis.Manager
    db    *gorm.DB
}

func NewAppRepo(db *gorm.DB, redis *redis.Manager) AppRepo {
    return &appRepo{
        redis: redis,
        db:    db,
    }
}

// 智能更新方法 - 只更新非零值字段
func (r *appRepo) Update(ctx context.Context, id uint, app *auth.App) error {
    updates := make(map[string]interface{})
    
    if app.AppID != "" {
        updates["app_id"] = app.AppID
    }
    if app.AppName != "" {
        updates["app_name"] = app.AppName
    }
    // ... 更多字段检查
    
    return r.db.WithContext(ctx).Model(&auth.App{}).Where("id = ?", id).Updates(updates).Error
}

// 带结构体参数的列表方法用于过滤
func (r *appRepo) List(ctx context.Context, app *auth.App) ([]auth.App, error) {
    return app.List(ctx, r.db)
}
```

### 支持的 MySQL 类型

| MySQL 类型 | Go 类型 | 备注 |
|------------|---------|------|
| int, integer | int | |
| tinyint | int8 | |
| smallint | int16 | |
| bigint | int64 | |
| varchar, char | string | |
| text, longtext | string | |
| decimal, numeric | float64 | |
| float | float32 | |
| double | float64 | |
| timestamp, datetime | time.Time | |
| date | time.Time | |
| json | string | |
| blob | []byte | |

### 目录结构

```
project/
├── bin/data/sql/           # SQL 文件（输入）
│   └── auth_app.sql
├── app/model/              # 生成的模型（输出）
│   └── auth/
│       └── app.go
├── app/repository/         # 生成的仓库（输出）
│   └── auth/
│       └── app.go
└── app/service/           # 生成的服务（未来功能）
```

### 最佳实践

1. **文件命名**: 为 SQL 文件使用描述性名称（如 `user_profile.sql`）
2. **注释**: 为 SQL 列添加有意义的注释
3. **约束**: 使用适当的 MySQL 约束（NOT NULL, DEFAULT 等）
4. **索引**: 在 SQL 中定义适当的索引
5. **备份**: 使用 `-force` 前始终备份现有文件

### 故障排除

#### 常见问题

1. **文件不存在错误**
    - 检查 SQL 文件路径是否正确
    - 确保文件具有 `.sql` 扩展名

2. **权限错误**
    - 确保对输出目录有写权限
    - 使用 `-force` 覆盖只读文件

3. **SQL 解析错误**
    - 检查 SQL 语法是否正确
    - 确保使用标准的 CREATE TABLE 语句

4. **生成的代码编译错误**
    - 检查 Go 模块路径是否正确
    - 确保导入的包可用

#### 调试技巧

- 使用单个文件测试：`./codegen -name your_table`
- 检查生成的文件内容
- 验证 SQL 文件格式