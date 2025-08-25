# go-api.sh Usage Guide | go-api.sh 使用指南

[English](#english) | [中文](#中文)

---

## English

### Overview

The `go-api.sh` script provides a shell-based alternative to the Makefile for environments where Make is not available or preferred. It offers the same functionality as the Makefile but with more portable shell scripting.

### Prerequisites

- Bash or compatible shell (sh, zsh, etc.)
- Go 1.19+ installed
- Docker (for Docker-related commands)
- goimports tool: `go install golang.org/x/tools/cmd/goimports@latest`

### Script Location

```bash
# Make script executable
chmod +x scripts/go-api.sh

# Run from project root
./scripts/go-api.sh <command>
```

### Available Commands

#### Development Commands

##### `all`
Runs the complete development workflow: format and build.

```bash
./scripts/go-api.sh all
```

**What it does:**
- Calls `fmt` to format code
- Calls `build` to compile the application

##### `fmt`
Formats the source code using `gofmt` and organizes imports with `goimports`.

```bash
./scripts/go-api.sh fmt
```

**Features:**
- Formats all Go files in the current directory
- Organizes and cleans up import statements
- Ensures consistent code style across the project

##### `build`
Builds the application binary with optimizations.

```bash
./scripts/go-api.sh build
```

**What it does:**
- Creates `./bin` directory if it doesn't exist
- Builds optimized binary with `-ldflags="-s -w"` (strip symbols and debug info)
- Output: `./bin/go-api` (or value of APP_NAME environment variable)

##### `run`
Runs the compiled application.

```bash
./scripts/go-api.sh run
```

**Requirements:**
- Binary must be built first with `build` command
- Uses RUN_ENV environment variable for runtime configuration

#### Docker Commands

##### `docker-build`
Builds the Docker image for the application.

```bash
./scripts/go-api.sh docker-build
```

**Features:**
- Uses current Dockerfile
- Sets timezone using TZ environment variable
- Tags image with IMAGE_NAME environment variable

##### `docker-run`
Runs the application in a Docker container.

```bash
./scripts/go-api.sh docker-run
```

**What it does:**
- Calls `docker-clean` first to remove existing container
- Runs new container with:
  - Port mapping: 8080:8080
  - Volume mount for configs: `CONFIG_DIR:/bin/configs`
  - Environment variables: APP_NAME, RUN_ENV
  - Restart policy: always
  - Interactive terminal and detached mode

##### `docker-clean`
Stops and removes the existing Docker container.

```bash
./scripts/go-api.sh docker-clean
```

**Safe operation:**
- Gracefully stops container if running (ignores errors if not running)
- Forcefully removes container (ignores errors if doesn't exist)

#### Cleanup Commands

##### `clean`
Removes build artifacts.

```bash
./scripts/go-api.sh clean
```

**What it removes:**
- Application binary from `./bin/` directory
- Keeps configuration and other files intact

### Environment Variables

The script uses environment variables for configuration with sensible defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_NAME` | `go-api` | Application name and binary name |
| `IMAGE_NAME` | `$APP_NAME:latest` | Docker image name and tag |
| `CONFIG_DIR` | `$(pwd)/bin/configs` | Configuration directory path |
| `TZ` | `Asia/Shanghai` | Timezone for Docker container |
| `RUN_ENV` | `local` | Runtime environment |

### Usage Examples

#### Basic Development Workflow

```bash
# Complete development cycle
./scripts/go-api.sh all

# Individual steps
./scripts/go-api.sh fmt      # Format code
./scripts/go-api.sh build    # Build binary
./scripts/go-api.sh run      # Run application
```

#### Custom Environment Variables

```bash
# Build with custom app name
APP_NAME=my-api ./scripts/go-api.sh build

# Run Docker with production environment
RUN_ENV=prod ./scripts/go-api.sh docker-run

# Build Docker with custom image name
IMAGE_NAME=my-registry/go-api:v1.0.0 ./scripts/go-api.sh docker-build

# Use different timezone
TZ=UTC ./scripts/go-api.sh docker-build
```

#### Docker Workflow

```bash
# Complete Docker workflow
./scripts/go-api.sh docker-build
./scripts/go-api.sh docker-run

# Check if container is running
docker ps | grep go-api

# View container logs
docker logs go-api

# Stop and cleanup
./scripts/go-api.sh docker-clean
```

### Advanced Usage

#### Environment-specific Configuration

```bash
# Development environment
export RUN_ENV=dev
export APP_NAME=go-api-dev
./scripts/go-api.sh all
./scripts/go-api.sh run

# Production environment
export RUN_ENV=prod
export APP_NAME=go-api-prod
export CONFIG_DIR=/opt/go-api/configs
./scripts/go-api.sh docker-build
./scripts/go-api.sh docker-run
```

#### Custom Configuration Directory

```bash
# Use custom config directory
CONFIG_DIR=/path/to/custom/configs ./scripts/go-api.sh docker-run
```

#### Integration with Other Tools

```bash
# Use with process managers
supervisord start "cd /app && ./scripts/go-api.sh run"

# Use in systemd service
ExecStart=/app/scripts/go-api.sh run
```

### Troubleshooting

#### Script Issues

**Problem**: Permission denied
```bash
# Solution: Make script executable
chmod +x scripts/go-api.sh
```

**Problem**: Command not found
```bash
# Solution: Use full path or ensure script is executable
./scripts/go-api.sh all
# Or add to PATH
export PATH=$PATH:$(pwd)/scripts
go-api.sh all
```

#### Build Issues

**Problem**: `goimports: command not found`
```bash
# Solution: Install goimports
go install golang.org/x/tools/cmd/goimports@latest
```

**Problem**: Binary not found when running
```bash
# Solution: Build first
./scripts/go-api.sh build
./scripts/go-api.sh run
```

#### Docker Issues

**Problem**: Docker daemon not running
```bash
# Solution: Start Docker service
sudo systemctl start docker  # Linux
# Or start Docker Desktop on macOS/Windows
```

**Problem**: Port 8080 already in use
```bash
# Solution: Stop existing containers
./scripts/go-api.sh docker-clean
# Or check what's using the port
lsof -i :8080
```

**Problem**: Configuration not found in container
```bash
# Solution: Ensure config directory exists and contains files
ls -la ./bin/configs/
# Create default config if needed
cp bin/configs/local.json.default bin/configs/local.json
```

### Script Comparison with Makefile

| Feature | Makefile | go-api.sh |
|---------|----------|-----------|
| **Portability** | Requires Make | Works with any POSIX shell |
| **Dependencies** | Make utility | Standard shell |
| **Parallel execution** | Yes (make -j) | No |
| **Error handling** | Built-in | Manual |
| **Variable syntax** | Make variables | Environment variables |
| **IDE integration** | Better | Basic |
| **Complex workflows** | Better | Good enough |

### When to Use go-api.sh

- **Windows environments** where Make is not easily available
- **Minimal Docker containers** without Make installed
- **Shell scripting environments** where shell is preferred
- **CI/CD pipelines** that prefer shell scripts
- **Development environments** where Make is not installed

### Customization

You can extend the script for additional functionality:

```bash
# Add at the end of the script before the case statement
deploy() {
  docker_build
  echo "Deploying to registry..."
  docker push $IMAGE_NAME
}

test() {
  echo "Running tests..."
  go test -v ./...
}

# Add to the case statement
deploy)
  deploy
  ;;
test)
  test
  ;;
```

---

## 中文

### 概述

`go-api.sh` 脚本为 Makefile 提供了基于 shell 的替代方案，适用于不支持 Make 或更偏好 shell 脚本的环境。它提供与 Makefile 相同的功能，但具有更好的可移植性。

### 前提条件

- Bash 或兼容的 shell（sh、zsh 等）
- 安装 Go 1.19+
- Docker（用于 Docker 相关命令）
- goimports 工具：`go install golang.org/x/tools/cmd/goimports@latest`

### 脚本位置

```bash
# 使脚本可执行
chmod +x scripts/go-api.sh

# 从项目根目录运行
./scripts/go-api.sh <命令>
```

### 可用命令

#### 开发命令

##### `all`
运行完整的开发工作流程：格式化和构建。

```bash
./scripts/go-api.sh all
```

**功能：**
- 调用 `fmt` 格式化代码
- 调用 `build` 编译应用程序

##### `fmt`
使用 `gofmt` 格式化源代码，并使用 `goimports` 整理导入。

```bash
./scripts/go-api.sh fmt
```

**特性：**
- 格式化当前目录中的所有 Go 文件
- 整理和清理导入语句
- 确保项目中一致的代码风格

##### `build`
构建优化的应用程序二进制文件。

```bash
./scripts/go-api.sh build
```

**功能：**
- 如果不存在则创建 `./bin` 目录
- 使用 `-ldflags="-s -w"` 构建优化的二进制文件（去除符号和调试信息）
- 输出：`./bin/go-api`（或 APP_NAME 环境变量的值）

##### `run`
运行编译的应用程序。

```bash
./scripts/go-api.sh run
```

**要求：**
- 必须先使用 `build` 命令构建二进制文件
- 使用 RUN_ENV 环境变量进行运行时配置

#### Docker 命令

##### `docker-build`
为应用程序构建 Docker 镜像。

```bash
./scripts/go-api.sh docker-build
```

**特性：**
- 使用当前 Dockerfile
- 使用 TZ 环境变量设置时区
- 使用 IMAGE_NAME 环境变量标记镜像

##### `docker-run`
在 Docker 容器中运行应用程序。

```bash
./scripts/go-api.sh docker-run
```

**功能：**
- 首先调用 `docker-clean` 移除现有容器
- 运行新容器，具有：
  - 端口映射：8080:8080
  - 配置卷挂载：`CONFIG_DIR:/bin/configs`
  - 环境变量：APP_NAME、RUN_ENV
  - 重启策略：always
  - 交互式终端和分离模式

##### `docker-clean`
停止并移除现有的 Docker 容器。

```bash
./scripts/go-api.sh docker-clean
```

**安全操作：**
- 如果容器正在运行则优雅停止（如果未运行则忽略错误）
- 强制移除容器（如果不存在则忽略错误）

#### 清理命令

##### `clean`
移除构建产物。

```bash
./scripts/go-api.sh clean
```

**移除内容：**
- `./bin/` 目录中的应用程序二进制文件
- 保持配置和其他文件不变

### 环境变量

脚本使用具有合理默认值的环境变量进行配置：

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `APP_NAME` | `go-api` | 应用程序名称和二进制文件名 |
| `IMAGE_NAME` | `$APP_NAME:latest` | Docker 镜像名称和标签 |
| `CONFIG_DIR` | `$(pwd)/bin/configs` | 配置目录路径 |
| `TZ` | `Asia/Shanghai` | Docker 容器的时区 |
| `RUN_ENV` | `local` | 运行时环境 |

### 使用示例

#### 基本开发工作流程

```bash
# 完整开发周期
./scripts/go-api.sh all

# 单独步骤
./scripts/go-api.sh fmt      # 格式化代码
./scripts/go-api.sh build    # 构建二进制文件
./scripts/go-api.sh run      # 运行应用程序
```

#### 自定义环境变量

```bash
# 使用自定义应用名称构建
APP_NAME=my-api ./scripts/go-api.sh build

# 使用生产环境运行 Docker
RUN_ENV=prod ./scripts/go-api.sh docker-run

# 使用自定义镜像名称构建 Docker
IMAGE_NAME=my-registry/go-api:v1.0.0 ./scripts/go-api.sh docker-build

# 使用不同时区
TZ=UTC ./scripts/go-api.sh docker-build
```

### 故障排除

#### 脚本问题

**问题**：权限被拒绝
```bash
# 解决方案：使脚本可执行
chmod +x scripts/go-api.sh
```

**问题**：找不到命令
```bash
# 解决方案：使用完整路径或确保脚本可执行
./scripts/go-api.sh all
# 或添加到 PATH
export PATH=$PATH:$(pwd)/scripts
go-api.sh all
```

#### 构建问题

**问题**：`goimports: command not found`
```bash
# 解决方案：安装 goimports
go install golang.org/x/tools/cmd/goimports@latest
```

#### Docker 问题

**问题**：Docker 守护进程未运行
```bash
# 解决方案：启动 Docker 服务
sudo systemctl start docker  # Linux
# 或在 macOS/Windows 上启动 Docker Desktop
```

### 脚本与 Makefile 的比较

| 特性 | Makefile | go-api.sh |
|------|----------|-----------|
| **可移植性** | 需要 Make | 适用于任何 POSIX shell |
| **依赖** | Make 工具 | 标准 shell |
| **并行执行** | 是（make -j） | 否 |
| **错误处理** | 内置 | 手动 |
| **变量语法** | Make 变量 | 环境变量 |
| **IDE 集成** | 更好 | 基本 |
| **复杂工作流** | 更好 | 足够好 |

### 何时使用 go-api.sh

- **Windows 环境**，Make 不易获得
- **最小 Docker 容器**，未安装 Make
- **Shell 脚本环境**，更偏好 shell
- **CI/CD 管道**，偏好 shell 脚本
- **开发环境**，未安装 Make

### 自定义

您可以扩展脚本以获得额外功能：

```bash
# 在 case 语句之前的脚本末尾添加
deploy() {
  docker_build
  echo "部署到注册表..."
  docker push $IMAGE_NAME
}

test() {
  echo "运行测试..."
  go test -v ./...
}

# 添加到 case 语句
deploy)
  deploy
  ;;
test)
  test
  ;;
```