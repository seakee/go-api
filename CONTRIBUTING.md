# Contributing Guide | 贡献指南

[English](#english) | [中文](#中文)

---

## English

Thank you for your interest in contributing to the Go-API framework! This guide will help you get started with contributing to the project.

### Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Please be respectful and constructive in all interactions.

### How to Contribute

#### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Use the issue template** if available
3. **Provide detailed information** including:
   - Go version
   - Operating system
   - Steps to reproduce
   - Expected vs actual behavior
   - Error messages or logs

#### Suggesting Features

We welcome feature suggestions! Please:

1. **Check if the feature already exists** or is planned
2. **Create a detailed proposal** including:
   - Use case and motivation
   - Proposed implementation approach
   - Potential impact on existing code
   - Examples of usage

#### Pull Requests

1. **Fork the repository** and create a feature branch
2. **Follow coding standards** (see below)
3. **Write tests** for new functionality
4. **Update documentation** as needed
5. **Ensure all tests pass**
6. **Submit a pull request** with a clear description

### Development Setup

#### Prerequisites

- Go 1.24 or higher
- Docker and Docker Compose
- Make
- Git

#### Setup Steps

```bash
# Fork and clone the repository
git clone https://github.com/your-username/go-api.git
cd go-api

# Install dependencies
go mod download

# Setup development environment
cp bin/configs/local.json.default bin/configs/local.json

# Run tests
make test

# Run the application
make run
```

### Coding Standards

#### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and these additional guidelines:

1. **Formatting**
   ```bash
   # Format code before committing
   make fmt
   
   # Run linter
   make lint
   ```

2. **Naming Conventions**
   - Use `camelCase` for variables and functions
   - Use `PascalCase` for exported types and functions
   - Use descriptive names that explain intent
   - Avoid abbreviations unless they're well-known

3. **Package Structure**
   - Keep packages focused on a single responsibility
   - Use clear, descriptive package names
   - Avoid circular dependencies

#### Code Organization

```go
// Package declaration and imports
package controller

import (
    // Standard library imports first
    "context"
    "errors"
    "fmt"
    
    // Third-party imports
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // Local imports last
    "github.com/seakee/go-api/app/model/auth"
    "github.com/seakee/go-api/app/service/auth"
)
```

#### Documentation Standards

1. **Function Comments**
   ```go
   // CreateUser creates a new user with the provided parameters.
   //
   // Parameters:
   //   - ctx: context.Context for managing request lifecycle.
   //   - params: *CreateUserParams containing user information.
   //
   // Returns:
   //   - *CreateUserResult: containing the created user information.
   //   - error: error if the creation fails, otherwise nil.
   func (s *userService) CreateUser(ctx context.Context, params *CreateUserParams) (*CreateUserResult, error) {
       // Implementation
   }
   ```

2. **Type Comments**
   ```go
   // User represents a user in the system.
   type User struct {
       ID       uint   `json:"id"`       // Unique identifier
       Username string `json:"username"` // User's login name
       Email    string `json:"email"`    // User's email address
   }
   ```

3. **Package Comments**
   ```go
   // Package auth provides authentication and authorization functionality
   // for the go-api framework. It includes JWT token management,
   // user authentication, and permission checking.
   package auth
   ```

#### Error Handling

1. **Error Creation**
   ```go
   // Use fmt.Errorf for context
   return fmt.Errorf("failed to create user: %w", err)
   
   // Use errors.New for simple messages
   return errors.New("user already exists")
   ```

2. **Error Wrapping**
   ```go
   // Wrap errors to preserve context
   if err != nil {
       return fmt.Errorf("database operation failed: %w", err)
   }
   ```

#### Testing Standards

1. **Test File Naming**
   ```
   user.go -> user_test.go
   ```

2. **Test Function Naming**
   ```go
   func TestUserService_CreateUser(t *testing.T) {
       // Test implementation
   }
   
   func TestUserService_CreateUser_ValidationError(t *testing.T) {
       // Specific test case
   }
   ```

3. **Test Structure**
   ```go
   func TestUserService_CreateUser(t *testing.T) {
       tests := []struct {
           name    string
           input   CreateUserParams
           want    *CreateUserResult
           wantErr bool
       }{
           {
               name: "valid user creation",
               input: CreateUserParams{
                   Username: "testuser",
                   Email:    "test@example.com",
               },
               want: &CreateUserResult{
                   Username: "testuser",
                   Email:    "test@example.com",
               },
               wantErr: false,
           },
           // More test cases...
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Test implementation
           })
       }
   }
   ```

#### Database Migrations

When adding new database schemas:

1. **Create SQL files** in `bin/data/sql/`
2. **Use descriptive names** like `add_user_table.sql`
3. **Include comments** explaining the purpose
4. **Test with the code generator**

```sql
-- bin/data/sql/user_profiles.sql
CREATE TABLE `user_profiles` (
    `id` int NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `user_id` int NOT NULL COMMENT 'Reference to users table',
    `first_name` varchar(50) NOT NULL COMMENT 'User first name',
    `last_name` varchar(50) NOT NULL COMMENT 'User last name',
    `avatar_url` varchar(500) DEFAULT NULL COMMENT 'Profile picture URL',
    `bio` text COMMENT 'User biography',
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `unique_user_id` (`user_id`),
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User Profile Information';
```

### Git Workflow

#### Branch Naming

- `feature/feature-name` - New features
- `bugfix/issue-description` - Bug fixes
- `hotfix/critical-fix` - Critical production fixes
- `docs/documentation-update` - Documentation changes

#### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(auth): add JWT token refresh functionality

Implement automatic token refresh mechanism to improve user experience.
Tokens are refreshed 5 minutes before expiration.

Closes #123
```

```
fix(database): resolve connection pool exhaustion

Fixed connection pool not releasing connections properly in high-load scenarios.

Fixes #456
```

#### Pull Request Process

1. **Create a clear title** describing the change
2. **Fill out the PR template** completely
3. **Link related issues** using keywords like "Closes #123"
4. **Request reviews** from maintainers
5. **Address feedback** promptly
6. **Ensure CI passes** before requesting final review

#### PR Template Example

```markdown
## Description
Brief description of changes made.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests for new functionality
- [ ] Updated existing tests

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or properly documented)
```

### Release Process

#### Version Numbers

We follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH`
- `MAJOR`: Breaking changes
- `MINOR`: New features (backward compatible)
- `PATCH`: Bug fixes (backward compatible)

#### Release Steps

1. **Update version** in relevant files
2. **Update changelog** with new features and fixes
3. **Create release tag** with proper notes
4. **Build and test** release artifacts
5. **Publish release** with detailed notes

### Community Guidelines

#### Getting Help

- **Documentation**: Check the wiki first
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions
- **Email**: Contact maintainers for private matters

#### Communication

- **Be respectful** and constructive
- **Provide context** when asking questions
- **Help others** when you can
- **Share knowledge** and experience

#### Recognition

Contributors are recognized in:
- Release notes
- Contributors file
- Project README
- Annual reports

---

## 中文

感谢您对Go-API框架的贡献兴趣！本指南将帮助您开始为项目做贡献。

### 行为准则

通过参与此项目，您同意遵守我们的行为准则。请在所有互动中保持尊重和建设性。

### 如何贡献

#### 报告问题

在创建问题之前，请：

1. **搜索现有问题**以避免重复
2. **使用问题模板**（如果可用）
3. **提供详细信息**包括：
   - Go版本
   - 操作系统
   - 重现步骤
   - 期望与实际行为
   - 错误消息或日志

#### 建议功能

我们欢迎功能建议！请：

1. **检查功能是否已存在**或已计划
2. **创建详细提案**包括：
   - 用例和动机
   - 建议的实现方法
   - 对现有代码的潜在影响
   - 使用示例

#### 拉取请求

1. **Fork仓库**并创建功能分支
2. **遵循编码标准**（见下文）
3. **为新功能编写测试**
4. **根据需要更新文档**
5. **确保所有测试通过**
6. **提交拉取请求**并提供清晰描述

### 开发环境设置

#### 前提条件

- Go 1.24或更高版本
- Docker和Docker Compose
- Make
- Git

#### 设置步骤

```bash
# Fork并克隆仓库
git clone https://github.com/your-username/go-api.git
cd go-api

# 安装依赖
go mod download

# 设置开发环境
cp bin/configs/local.json.default bin/configs/local.json

# 运行测试
make test

# 运行应用
make run
```

### 编码标准

#### Go风格指南

遵循官方[Go代码审查评论](https://github.com/golang/go/wiki/CodeReviewComments)和这些额外指导原则：

1. **格式化**
   ```bash
   # 提交前格式化代码
   make fmt
   
   # 运行代码检查
   make lint
   ```

2. **命名约定**
   - 变量和函数使用`camelCase`
   - 导出的类型和函数使用`PascalCase`
   - 使用说明意图的描述性名称
   - 避免缩写，除非是众所周知的

3. **包结构**
   - 保持包专注于单一职责
   - 使用清晰、描述性的包名
   - 避免循环依赖

#### 代码组织

```go
// 包声明和导入
package controller

import (
    // 标准库导入优先
    "context"
    "errors"
    "fmt"
    
    // 第三方导入
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    
    // 本地导入最后
    "github.com/seakee/go-api/app/model/auth"
    "github.com/seakee/go-api/app/service/auth"
)
```

#### 文档标准

1. **函数注释**
   ```go
   // CreateUser 使用提供的参数创建新用户。
   //
   // 参数:
   //   - ctx: context.Context 用于管理请求生命周期。
   //   - params: *CreateUserParams 包含用户信息。
   //
   // 返回值:
   //   - *CreateUserResult: 包含创建的用户信息。
   //   - error: 创建失败时返回错误，否则为nil。
   func (s *userService) CreateUser(ctx context.Context, params *CreateUserParams) (*CreateUserResult, error) {
       // 实现
   }
   ```

#### 错误处理

1. **错误创建**
   ```go
   // 使用fmt.Errorf添加上下文
   return fmt.Errorf("创建用户失败: %w", err)
   
   // 使用errors.New创建简单消息
   return errors.New("用户已存在")
   ```

2. **错误包装**
   ```go
   // 包装错误以保留上下文
   if err != nil {
       return fmt.Errorf("数据库操作失败: %w", err)
   }
   ```

#### 测试标准

1. **测试文件命名**
   ```
   user.go -> user_test.go
   ```

2. **测试函数命名**
   ```go
   func TestUserService_CreateUser(t *testing.T) {
       // 测试实现
   }
   
   func TestUserService_CreateUser_ValidationError(t *testing.T) {
       // 特定测试用例
   }
   ```

### Git工作流程

#### 分支命名

- `feature/feature-name` - 新功能
- `bugfix/issue-description` - 错误修复
- `hotfix/critical-fix` - 关键生产修复
- `docs/documentation-update` - 文档更改

#### 提交消息

遵循[约定式提交](https://www.conventionalcommits.org/zh-hans/)：

```
type(scope): description

[可选的正文]

[可选的脚注]
```

**类型：**
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更改
- `style`: 代码风格更改
- `refactor`: 代码重构
- `test`: 添加或更新测试
- `chore`: 维护任务

**示例：**
```
feat(auth): 添加JWT令牌刷新功能

实现自动令牌刷新机制以改善用户体验。
令牌在过期前5分钟自动刷新。

Closes #123
```

### 社区指导原则

#### 获取帮助

- **文档**: 首先查看wiki
- **问题**: 创建新问题前搜索现有问题
- **讨论**: 使用GitHub讨论进行提问
- **邮箱**: 私人事务联系维护者

#### 沟通

- **保持尊重**和建设性
- **提问时提供上下文**
- **在力所能及时帮助他人**
- **分享知识**和经验

#### 认可

贡献者将在以下地方得到认可：
- 发布说明
- 贡献者文件
- 项目README
- 年度报告