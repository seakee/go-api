# Go API Code Generator

English | [中文](README_ZH.MD)

---
### Overview

The Go API Code Generator generates Go model and repository code from SQL `CREATE TABLE` statements. It currently supports MySQL and a minimal PostgreSQL subset. It parses SQL files and creates corresponding Go structs with GORM tags, along with repository interfaces and implementations for common database operations.

### Features

- **SQL Parsing**: Supports common MySQL and PostgreSQL field type definitions
- **Type Mapping**: Maps common SQL types to Go types
- **GORM Integration**: Automatic generation of GORM tags (column, size, not null, default, etc.)
- **Comment Support**: Preserves SQL comments as Go struct field comments
- **Repository Generation**: Automatic generation of repository interfaces and implementations
- **Smart Update Logic**: Generates update methods that only update non-zero fields
- **Explicit Field Updates**: Generates `UpdateFields` for safe zero-value updates
- **Flexible List Methods**: List methods that accept struct parameters for query conditions
- **Batch Processing**: Process single files or entire directories
- **Force Overwrite**: Option to overwrite existing files

### Installation

```bash
# Build the code generator
go build -o codegen ./command/codegen/handler.go
```

### Usage

#### Command Line Options

```bash
./codegen [options]
```

**Available Options:**

- `-force`: Force overwrite existing files (default: false)
- `-dialect`: SQL dialect, `mysql` or `postgres` (default: `mysql`)
- `-name`: SQL file name (without .sql extension) to generate code for
- `-sql`: SQL directory path (default: "bin/data/sql")
- `-model`: Model output directory (default: "app/model")
- `-repo`: Repository output directory (default: "app/repository")
- `-service`: Service output directory (default: "app/service")

#### Examples

**Generate code for a single SQL file:**
```bash
./codegen -name auth_app
```

**Generate code for a PostgreSQL SQL file:**
```bash
./codegen -dialect=postgres -sql bin/data/sql/postgres -name oauth_app
```

**Generate code for all SQL files in directory:**
```bash
./codegen
```

**Use custom paths:**
```bash
./codegen -sql custom/sql/path -model custom/model/path -force
```

### SQL File Format

#### MySQL Example

The generator supports standard MySQL `CREATE TABLE` statements. Example:

```sql
CREATE TABLE `auth_app`
(
    `id`           int                                                           NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `app_id`       varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci  NOT NULL COMMENT 'Application ID',
    `app_name`     varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci           DEFAULT NULL COMMENT 'Application Name',
    `app_secret`   varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'Application Secret',
    `status`       tinyint(1)                                                    NOT NULL DEFAULT '0' COMMENT 'Status: 0=Inactive, 1=Active, 2=Disabled',
    `created_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   timestamp                                                     NULL     DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`   timestamp                                                     NULL     DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `app_id` (`app_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'Application Information Table';
```

#### PostgreSQL Example

The generator also supports a minimal PostgreSQL subset for single-table DDL. Example:

```sql
CREATE TABLE IF NOT EXISTS "public"."oauth_app"
(
    "app_uuid"   uuid PRIMARY KEY,
    "app_name"   varchar(80)              NOT NULL,
    "payload"    jsonb                    DEFAULT NULL,
    "status"     integer                  NOT NULL DEFAULT 0,
    "created_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    "updated_at" timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);
```

### Generated Output

The generator creates:

**Model Files:**
1. **Proper Go naming conventions** (snake_case → CamelCase)
2. **GORM tags** for database mapping
3. **JSON tags** for API serialization
4. **Field comments** from SQL comments
5. **Common database methods** (CRUD operations)

**Repository Files:**
1. **Repository interfaces** with standard CRUD operations
2. **Repository implementations** with dependency injection
3. **Smart update methods** that only update non-zero fields
4. **Flexible list methods** that accept struct parameters for filtering

**Example Generated Model Code:**

```go
package auth

import (
    "context"
    "gorm.io/gorm"
)

type App struct {
    gorm.Model
    
    AppID       string `gorm:"column:app_id;size:30;not null" json:"app_id"`           // Application ID
    AppName     string `gorm:"column:app_name;size:50" json:"app_name"`                // Application Name
    AppSecret   string `gorm:"column:app_secret;size:256;not null" json:"app_secret"`  // Application Secret
    Status      int8   `gorm:"column:status;not null;default:0" json:"status"`         // Status: 0=Inactive, 1=Active, 2=Disabled
}

func (a *App) TableName() string {
    return "auth_app"
}

// Database operation methods...
func (a *App) First(ctx context.Context, db *gorm.DB) (*App, error) { /* ... */ }
func (a *App) Create(ctx context.Context, db *gorm.DB) (uint, error) { /* ... */ }
func (a *App) List(ctx context.Context, db *gorm.DB) ([]App, error) { /* ... */ }
// ... more methods
```

**Example Generated Repository Code:**

```go
package auth

import (
    "context"
    "github.com/seakee/go-api/app/model/auth"
    "github.com/sk-pkg/redis"
    "gorm.io/gorm"
)

// Repository interface
type AppRepo interface {
    GetApp(ctx context.Context, app *auth.App) (*auth.App, error)
    Create(ctx context.Context, app *auth.App) (uint, error)
    Update(ctx context.Context, id uint, app *auth.App) error
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, app *auth.App) ([]auth.App, error)
    GetByID(ctx context.Context, id uint) (*auth.App, error)
}

// Repository implementation
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

// Smart update method - only updates non-zero fields
func (r *appRepo) Update(ctx context.Context, id uint, app *auth.App) error {
    updates := make(map[string]interface{})
    
    if app.AppID != "" {
        updates["app_id"] = app.AppID
    }
    if app.AppName != "" {
        updates["app_name"] = app.AppName
    }
    // ... more field checks
    
    return r.db.WithContext(ctx).Model(&auth.App{}).Where("id = ?", id).Updates(updates).Error
}

// List method with struct parameter for filtering
func (r *appRepo) List(ctx context.Context, app *auth.App) ([]auth.App, error) {
    return app.List(ctx, r.db)
}
```

### Supported MySQL Types

| MySQL Type | Go Type | Notes |
|------------|---------|-------|
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

### Supported PostgreSQL Types

| PostgreSQL Type | Go Type | Notes |
|-----------------|---------|-------|
| smallserial | int16 | |
| serial | int32 | |
| bigserial | int64 | |
| integer | int32 | |
| bigint | int64 | |
| uuid | string | minimal support |
| varchar | string / *string | nullable columns become pointers |
| text | string / *string | |
| boolean | bool / *bool | |
| numeric | decimal.Decimal / *decimal.Decimal | |
| jsonb | datatypes.JSON / *datatypes.JSON | |
| bytea | []byte | |
| timestamp with time zone, timestamptz | time.Time / *time.Time | |
| timestamp without time zone | time.Time / *time.Time | |
| date | time.Time / *time.Time | |

### PostgreSQL Scope

Current PostgreSQL support is intentionally limited. It supports:

- single `CREATE TABLE` statements
- optional schema names
- quoted identifiers
- single-column primary keys
- common scalar PostgreSQL types listed above

It does not yet support:

- composite primary keys
- array types
- `COMMENT ON`
- split definitions via `ALTER TABLE`
- advanced PostgreSQL-specific features such as enums or domains

### TODO

- Add PostgreSQL generated output snippets to the README so users can quickly compare input and output.
- Extend PostgreSQL support in a future iteration, prioritizing `COMMENT ON` and array types.
| json | string | |
| blob | []byte | |

### Directory Structure

```
project/
├── bin/data/sql/           # SQL files (input)
│   └── auth_app.sql
├── app/model/              # Generated models (output)
│   └── auth/
│       └── app.go
├── app/repository/         # Generated repositories (output)
│   └── auth/
│       └── app.go
└── app/service/           # Generated services (future)
```

### Best Practices

1. **File Naming**: Use descriptive names for SQL files (e.g., `user_profile.sql`)
2. **Comments**: Add meaningful comments to SQL columns
3. **Constraints**: Use proper MySQL constraints (NOT NULL, DEFAULT, etc.)
4. **Indexes**: Define appropriate indexes in your SQL
5. **Backup**: Always backup existing files before using `-force`
