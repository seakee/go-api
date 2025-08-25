# Makefile Usage Guide | Makefile 使用指南

[English](#english) | [中文](#中文)

---

## English

### Overview

The Makefile provides convenient build automation for the Go-API project. It includes targets for development, testing, building, and deployment tasks.

### Prerequisites

- Go 1.19+ installed
- Docker (for Docker-related targets)
- goimports tool: `go install golang.org/x/tools/cmd/goimports@latest`

### Available Targets

#### Development Targets

##### `make all` (Default)
Runs the complete development workflow: format, test, and build.

```bash
make all
# Equivalent to: make fmt test build
```

##### `make fmt`
Formats the source code using `gofmt` and organizes imports with `goimports`.

```bash
make fmt
```

**What it does:**
- Formats all Go files in the current directory
- Organizes and cleans up import statements
- Ensures consistent code style

##### `make test`
Runs all tests with verbose output.

```bash
make test
```

**Features:**
- Runs tests in all subdirectories
- Provides verbose output for debugging
- Exits with error code if tests fail

##### `make build`
Builds the application binary with optimizations.

```bash
make build
```

**What it does:**
- Runs `make fmt` first to ensure code is formatted
- Creates `./bin` directory if it doesn't exist
- Builds optimized binary with `-s -w` flags (strip symbols and debug info)
- Output: `./bin/go-api` (or custom APP_NAME)

##### `make run`
Runs the compiled application.

```bash
make run
```

**Requirements:**
- Binary must be built first with `make build`
- Configuration files should be in `bin/configs/`

#### Docker Targets

##### `make docker-build`
Builds the Docker image for the application.

```bash
make docker-build
```

**Features:**
- Uses optimized Dockerfile
- Sets timezone using TZ environment variable
- Tags image as `go-api:latest` (or custom IMAGE_NAME)

##### `make docker-run`
Runs the application in a Docker container.

```bash
make docker-run
```

**What it does:**
- Stops and removes existing container (via `docker-clean`)
- Runs new container with:
  - Port mapping: 8080:8080
  - Volume mount for configs: `./bin/configs:/bin/configs`
  - Environment variables: APP_NAME, RUN_ENV
  - Restart policy: always
  - Detached mode

##### `make docker-clean`
Stops and removes the existing Docker container.

```bash
make docker-clean
```

**Safe operation:**
- Gracefully stops container if running
- Removes container forcefully if needed
- Ignores errors if container doesn't exist

#### Cleanup Targets

##### `make clean`
Removes build artifacts.

```bash
make clean
```

**What it removes:**
- Application binary from `./bin/` directory
- Keeps configuration and other files intact

### Environment Variables

You can customize the build process using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_NAME` | `go-api` | Application name and binary name |
| `TZ` | `Asia/Shanghai` | Timezone for Docker container |
| `IMAGE_NAME` | `$(APP_NAME):latest` | Docker image name and tag |
| `CONFIG_DIR` | `$(pwd)/bin/configs` | Configuration directory path |
| `RUN_ENV` | `local` | Runtime environment |

### Usage Examples

#### Basic Development Workflow

```bash
# Complete development cycle
make all

# Individual steps
make fmt      # Format code
make test     # Run tests
make build    # Build binary
make run      # Run application
```

#### Custom App Name

```bash
# Build with custom name
APP_NAME=my-api make build

# Run Docker with custom name
APP_NAME=my-api make docker-run
```

#### Different Environment

```bash
# Build for production
RUN_ENV=prod make build

# Run Docker in production mode
RUN_ENV=prod make docker-run
```

#### Custom Docker Image

```bash
# Build with custom image name
IMAGE_NAME=my-registry/go-api:v1.0.0 make docker-build

# Run with custom image
IMAGE_NAME=my-registry/go-api:v1.0.0 make docker-run
```

#### Different Timezone

```bash
# Build Docker with different timezone
TZ=UTC make docker-build
```

### Troubleshooting

#### Build Issues

**Problem**: `goimports: command not found`
```bash
# Solution: Install goimports
go install golang.org/x/tools/cmd/goimports@latest
```

**Problem**: Binary not found when running `make run`
```bash
# Solution: Build first
make build
make run
```

#### Docker Issues

**Problem**: Permission denied accessing Docker
```bash
# Solution: Add user to docker group or use sudo
sudo make docker-build
sudo make docker-run
```

**Problem**: Port 8080 already in use
```bash
# Solution: Stop existing containers or change port
make docker-clean
# Or modify Makefile to use different port
```

**Problem**: Configuration files not found in container
```bash
# Solution: Ensure configs exist locally
ls -la ./bin/configs/
# Create default config if needed
cp bin/configs/local.json.default bin/configs/local.json
```

### Advanced Usage

#### Parallel Execution

```bash
# Run format and tests in parallel (if supported)
make -j2 fmt test
```

#### Verbose Output

```bash
# See actual commands being executed
make -n build

# Verbose make output
make -d build
```

#### Integration with CI/CD

```yaml
# GitHub Actions example
- name: Build and Test
  run: |
    make all
    make docker-build
```

### Customization

You can extend the Makefile for your specific needs:

```makefile
# Add custom target
deploy: docker-build
	@echo "Deploying application..."
	@docker push $(IMAGE_NAME)
	
# Add linting
lint:
	@golangci-lint run ./...
	
# Add coverage
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
```

---

## 中文

### 概述

Makefile 为 Go-API 项目提供了便捷的构建自动化。它包括用于开发、测试、构建和部署任务的目标。

### 前提条件

- 安装 Go 1.19+
- Docker（用于 Docker 相关目标）
- goimports 工具：`go install golang.org/x/tools/cmd/goimports@latest`

### 可用目标

#### 开发目标

##### `make all`（默认）
运行完整的开发工作流程：格式化、测试和构建。

```bash
make all
# 等同于：make fmt test build
```

##### `make fmt`
使用 `gofmt` 格式化源代码，并使用 `goimports` 整理导入。

```bash
make fmt
```

**功能：**
- 格式化当前目录中的所有 Go 文件
- 整理和清理导入语句
- 确保一致的代码风格

##### `make test`
运行所有测试并提供详细输出。

```bash
make test
```

**特性：**
- 在所有子目录中运行测试
- 提供详细输出用于调试
- 如果测试失败则退出并返回错误代码

##### `make build`
构建优化的应用程序二进制文件。

```bash
make build
```

**功能：**
- 首先运行 `make fmt` 确保代码格式化
- 如果不存在则创建 `./bin` 目录
- 使用 `-s -w` 标志构建优化的二进制文件（去除符号和调试信息）
- 输出：`./bin/go-api`（或自定义 APP_NAME）

##### `make run`
运行编译的应用程序。

```bash
make run
```

**要求：**
- 必须先使用 `make build` 构建二进制文件
- 配置文件应该在 `bin/configs/` 目录中

#### Docker 目标

##### `make docker-build`
为应用程序构建 Docker 镜像。

```bash
make docker-build
```

**特性：**
- 使用优化的 Dockerfile
- 使用 TZ 环境变量设置时区
- 标记镜像为 `go-api:latest`（或自定义 IMAGE_NAME）

##### `make docker-run`
在 Docker 容器中运行应用程序。

```bash
make docker-run
```

**功能：**
- 停止并移除现有容器（通过 `docker-clean`）
- 运行新容器，具有：
  - 端口映射：8080:8080
  - 配置卷挂载：`./bin/configs:/bin/configs`
  - 环境变量：APP_NAME、RUN_ENV
  - 重启策略：always
  - 分离模式

##### `make docker-clean`
停止并移除现有的 Docker 容器。

```bash
make docker-clean
```

**安全操作：**
- 如果容器正在运行则优雅停止
- 如果需要则强制移除容器
- 如果容器不存在则忽略错误

#### 清理目标

##### `make clean`
移除构建产物。

```bash
make clean
```

**移除内容：**
- `./bin/` 目录中的应用程序二进制文件
- 保持配置和其他文件不变

### 环境变量

您可以使用环境变量自定义构建过程：

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `APP_NAME` | `go-api` | 应用程序名称和二进制文件名 |
| `TZ` | `Asia/Shanghai` | Docker 容器的时区 |
| `IMAGE_NAME` | `$(APP_NAME):latest` | Docker 镜像名称和标签 |
| `CONFIG_DIR` | `$(pwd)/bin/configs` | 配置目录路径 |
| `RUN_ENV` | `local` | 运行时环境 |

### 使用示例

#### 基本开发工作流程

```bash
# 完整开发周期
make all

# 单独步骤
make fmt      # 格式化代码
make test     # 运行测试
make build    # 构建二进制文件
make run      # 运行应用程序
```

#### 自定义应用名称

```bash
# 使用自定义名称构建
APP_NAME=my-api make build

# 使用自定义名称运行 Docker
APP_NAME=my-api make docker-run
```

#### 不同环境

```bash
# 为生产环境构建
RUN_ENV=prod make build

# 在生产模式下运行 Docker
RUN_ENV=prod make docker-run
```

### 故障排除

#### 构建问题

**问题**：`goimports: command not found`
```bash
# 解决方案：安装 goimports
go install golang.org/x/tools/cmd/goimports@latest
```

**问题**：运行 `make run` 时找不到二进制文件
```bash
# 解决方案：先构建
make build
make run
```

#### Docker 问题

**问题**：访问 Docker 权限被拒绝
```bash
# 解决方案：将用户添加到 docker 组或使用 sudo
sudo make docker-build
sudo make docker-run
```

**问题**：端口 8080 已被使用
```bash
# 解决方案：停止现有容器或更改端口
make docker-clean
# 或修改 Makefile 使用不同端口
```

### 自定义

您可以根据特定需求扩展 Makefile：

```makefile
# 添加自定义目标
deploy: docker-build
	@echo "部署应用程序..."
	@docker push $(IMAGE_NAME)
	
# 添加代码检查
lint:
	@golangci-lint run ./...
	
# 添加覆盖率检查
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
```