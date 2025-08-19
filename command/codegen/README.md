# Go API Code Generator

[English](README.md#english) | [中文](README_ZH.MD)

---
### Overview

The Go API Code Generator is a powerful tool that automatically generates Go model code from MySQL CREATE TABLE statements. It parses SQL files and creates corresponding Go structs with GORM tags, along with common database operations.

### Features

- **SQL Parsing**: Supports complex field type definitions (VARCHAR(255), DECIMAL(10,2), etc.)
- **Type Mapping**: Comprehensive mapping from MySQL types to Go types
- **GORM Integration**: Automatic generation of GORM tags (column, size, not null, default, etc.)
- **Comment Support**: Preserves SQL comments as Go struct field comments
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

**Generate code for all SQL files in directory:**
```bash
./codegen
```

**Use custom paths:**
```bash
./codegen -sql custom/sql/path -model custom/model/path -force
```

### SQL File Format

The generator expects standard MySQL CREATE TABLE statements. Here's an example:

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

### Generated Output

The generator creates Go structs with:

1. **Proper Go naming conventions** (snake_case → CamelCase)
2. **GORM tags** for database mapping
3. **JSON tags** for API serialization
4. **Field comments** from SQL comments
5. **Common database methods** (CRUD operations)

**Example Generated Code:**

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
func (a *App) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error { /* ... */ }
// ... more methods
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
├── app/repository/         # Generated repositories (future)
└── app/service/           # Generated services (future)
```

### Best Practices

1. **File Naming**: Use descriptive names for SQL files (e.g., `user_profile.sql`)
2. **Comments**: Add meaningful comments to SQL columns
3. **Constraints**: Use proper MySQL constraints (NOT NULL, DEFAULT, etc.)
4. **Indexes**: Define appropriate indexes in your SQL
5. **Backup**: Always backup existing files before using `-force`
