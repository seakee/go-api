#!/usr/bin/env bash

# Generate a go-api project
# This script creates a new project based on the go-api template
#
# Usage:
#   ./generate.sh [project-name] [version] [module-name]
#   ./generate.sh my-project v1.0.0 github.com/myuser/my-project
#   ./generate.sh my-project main my-project
#   ./generate.sh my-project        # Uses latest main branch, module name = project name
#   ./generate.sh                   # Creates 'go-api' project from main branch
#
# Parameters:
#   $1 - Project name (optional, default: "go-api")
#   $2 - Version/branch (optional, default: "main")
#   $3 - Module name (optional, default: project-name)

set -e  # Exit on any error
set -u  # Exit on undefined variables
set -o pipefail  # Exit on pipe failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Global variables
projectDir=""
cleanup_required=false

# Print colored output with aligned formatting
print_info() { printf "${BLUE}%-9s${NC} %s\n" "[INFO]" "$1"; }
print_success() { printf "${GREEN}%-9s${NC} %s\n" "[SUCCESS]" "$1"; }
print_warning() { printf "${YELLOW}%-9s${NC} %s\n" "[WARNING]" "$1"; }
print_error() { printf "${RED}%-9s${NC} %s\n" "[ERROR]" "$1"; }

# Cleanup function for error handling
cleanup() {
    local exit_code=$?
    if [[ $cleanup_required == true && -n "$projectDir" && -d "$projectDir" ]]; then
        # Only cleanup if it's an incomplete project (no .git/config means incomplete)
        if [[ ! -f "$projectDir/.git/config" ]]; then
            print_warning "Cleaning up incomplete project directory: $projectDir"
            rm -rf "$projectDir"
        fi
    fi
    exit $exit_code
}

# Set up trap for cleanup
trap cleanup EXIT INT TERM

# Show usage information
show_usage() {
    echo "Usage: $0 [project-name] [version] [module-name]"
    echo ""
    echo "Parameters:"
    echo "  project-name  Name of the new project (default: go-api)"
    echo "  version       Git branch or tag to use (default: main)"
    echo "  module-name   Go module name (default: project-name)"
    echo ""
    echo "Examples:"
    echo "  $0                                              # Create 'go-api' from main"
    echo "  $0 my-awesome-api                               # Module: my-awesome-api"
    echo "  $0 my-api v1.2.0                                # Module: my-api"
    echo "  $0 my-api main github.com/myuser/my-api         # Custom module name"
    echo "  $0 my-api main my-api                           # Simple module name"
    echo ""
    echo "Module name guidelines:"
    echo "  - For local use only: use simple name (e.g., 'my-project')"
    echo "  - For GitHub: use 'github.com/username/project-name'"
    echo "  - For other Git hosts: use full repository URL"
    echo "  - Must be valid Go module name (lowercase, no spaces)"
    echo ""
    echo "Requirements:"
    echo "  - Git 2.0+"
    echo "  - Go 1.24+"
    echo "  - Write permission in current directory"
    echo "  - Internet connection for cloning repository"
}

# Validate project name
validate_project_name() {
    local name="$1"

    # Check if name contains invalid characters
    if [[ ! "$name" =~ ^[a-zA-Z0-9_-]+$ ]]; then
        print_error "Invalid project name: '$name'. Only letters, numbers, hyphens, and underscores are allowed."
        exit 1
    fi

    # Check if name is too short
    if [[ ${#name} -lt 2 ]]; then
        print_error "Project name must be at least 2 characters long."
        exit 1
    fi

    # Check if name is too long (filesystem limitation)
    if [[ ${#name} -gt 100 ]]; then
        print_error "Project name is too long (max 100 characters)."
        exit 1
    fi

    # Check if name starts with a dot (hidden directory)
    if [[ "$name" =~ ^\..*$ ]]; then
        print_warning "Project name starts with '.', this will create a hidden directory."
        read -p "Continue? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Operation cancelled."
            exit 0
        fi
    fi
}

# Validate Go module name
validate_module_name() {
    local module_name="$1"

    # Check if module name is empty
    if [[ -z "$module_name" ]]; then
        print_error "Module name cannot be empty"
        exit 1
    fi

    # Check if module name contains invalid characters for Go modules
    if [[ "$module_name" =~ [A-Z[:space:]] ]]; then
        print_error "Invalid module name: '$module_name'"
        print_error "Go module names should be lowercase and contain no spaces"
        print_error "Valid examples: my-project, github.com/user/project"
        exit 1
    fi

    # Warn about local vs remote module names
    if [[ "$module_name" != *"."* ]]; then
        print_info "Using local module name: '$module_name'"
        print_info "This is suitable for local development or internal use"
    else
        print_info "Using remote module name: '$module_name'"
        print_info "Make sure this matches your intended repository location"
    fi
}

# Check file system permissions
check_file_permissions() {
    print_info "Checking file system permissions..."

    local current_dir
    current_dir="$(pwd)"

    if [[ ! -w "$current_dir" ]]; then
        print_error "No write permission in current directory: $current_dir"
        print_error "Please run the script from a directory where you have write permissions."
        exit 1
    fi

    print_success "File system permissions OK."
}

# Check Go version compatibility
check_go_version() {
    local go_version_output
    local go_version
    local min_version="1.24"

    go_version_output=$(go version 2>/dev/null || echo "")
    if [[ -z "$go_version_output" ]]; then
        print_error "Unable to determine Go version"
        exit 1
    fi

    # Extract version number (e.g., "1.21.0" from "go version go1.21.0 darwin/amd64")
    go_version=$(echo "$go_version_output" | grep -oE 'go[0-9]+\.[0-9]+(\.[0-9]+)?' | sed 's/go//' | head -1)

    if [[ -z "$go_version" ]]; then
        print_error "Unable to parse Go version from: $go_version_output"
        exit 1
    fi

    # Simple version comparison (works for major.minor format)
    if ! printf '%s\n%s\n' "$min_version" "$go_version" | sort -V -C; then
        print_error "Go version $go_version is too old. Minimum required: $min_version"
        print_error "Please update Go and try again."
        exit 1
    fi

    print_success "Go version $go_version is compatible."
}

# Check Git version
check_git_version() {
    local git_version_output
    local git_version
    local min_version="2.0"

    git_version_output=$(git --version 2>/dev/null || echo "")
    if [[ -z "$git_version_output" ]]; then
        print_error "Unable to determine Git version"
        exit 1
    fi

    # Extract version number
    git_version=$(echo "$git_version_output" | grep -oE '[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1)

    if [[ -z "$git_version" ]]; then
        print_error "Unable to parse Git version from: $git_version_output"
        exit 1
    fi

    # Simple version comparison
    if ! printf '%s\n%s\n' "$min_version" "$git_version" | sort -V -C; then
        print_error "Git version $git_version is too old. Minimum required: $min_version"
        print_error "Please update Git and try again."
        exit 1
    fi

    print_success "Git version $git_version is compatible."
}

# Check if required tools are installed
check_dependencies() {
    print_info "Checking dependencies..."

    local missing_tools=()

    if ! command -v git &> /dev/null; then
        missing_tools+=("git")
    fi

    if ! command -v go &> /dev/null; then
        missing_tools+=("go")
    fi

    # Check for additional useful tools
    if ! command -v make &> /dev/null; then
        print_warning "Make is not installed. Some project commands may not work."
    fi

    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_error "Please install them and try again."
        exit 1
    fi

    # Check versions
    check_git_version
    check_go_version

    print_success "All dependencies are installed and compatible."
}

# Check network connectivity
check_network() {
    print_info "Checking network connectivity..."

    if ! ping -c 1 github.com &> /dev/null; then
        print_error "Cannot reach github.com. Please check your internet connection."
        exit 1
    fi

    print_success "Network connectivity OK."
}

# Cross-platform sed replacement - Fixed version
sed_replace() {
    local pattern="$1"
    local replacement="$2"
    local file="$3"

    # Check file exists and is readable
    if [[ ! -f "$file" || ! -r "$file" ]]; then
        print_warning "Skipping file: $file (not found or not readable)"
        return 0
    fi

    # Check file is writable
    if [[ ! -w "$file" ]]; then
        print_warning "Skipping file: $file (not writable)"
        return 0
    fi

    # Check if file contains the pattern (without regex escaping for grep)
    local grep_pattern="${pattern//\\./.}"  # Convert \. to . for grep
    if ! grep -q "$grep_pattern" "$file" 2>/dev/null; then
        return 0
    fi

    print_info "Updating file: $file"

    # Create backup
    local backup_file="${file}.bak.$"
    if ! cp "$file" "$backup_file"; then
        print_error "Failed to create backup for: $file"
        return 1
    fi

    # Perform replacement with better error handling
    local sed_result=0
    local temp_file="${file}.tmp.$"

    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS - use temporary file approach to avoid sed -i issues
        if sed "s|${pattern}|${replacement}|g" "$file" > "$temp_file"; then
            if mv "$temp_file" "$file"; then
                sed_result=0
            else
                sed_result=1
            fi
        else
            sed_result=1
        fi
    else
        # Linux and others - use in-place editing
        sed -i "s|${pattern}|${replacement}|g" "$file" || sed_result=$?
    fi

    # Clean up temp file if it exists
    [[ -f "$temp_file" ]] && rm -f "$temp_file"

    if [[ $sed_result -ne 0 ]]; then
        print_error "Failed to update file: $file"
        # Restore from backup
        if ! mv "$backup_file" "$file"; then
            print_error "Failed to restore backup for: $file"
        fi
        return 1
    fi

    # Verify the replacement worked
    local verification_pattern="${replacement//\//\\/}"  # Escape forward slashes for grep
    if ! grep -q "$verification_pattern" "$file" 2>/dev/null; then
        print_warning "Replacement may not have worked correctly in: $file"
    fi

    # Remove backup on success
    rm -f "$backup_file"
    return 0
}

# Batch replace content in files
replace_in_files() {
    local old_pattern="$1"
    local new_replacement="$2"
    local project_dir="$3"

    print_info "Searching for files to update..."

    # Use array to store found files
    local files_to_update=()
    local file_count=0

    # Find all relevant files with better error handling
    while IFS= read -r -d '' file; do
        if [[ -f "$file" && -r "$file" ]]; then
            files_to_update+=("$file")
            ((file_count++))
        fi
    done < <(find "$project_dir" -type f \( -name "*.go" -o -name "*.mod" -o -name "*.md" -o -name "Dockerfile" -o -name "Makefile" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" \) -print0 2>/dev/null || true)

    if [[ $file_count -eq 0 ]]; then
        print_warning "No files found to update"
        return 0
    fi

    print_info "Found $file_count files to check"

    local updated_count=0
    local failed_count=0

    # Update each file
    for file in "${files_to_update[@]}"; do
        if sed_replace "$old_pattern" "$new_replacement" "$file"; then
            ((updated_count++))
        else
            ((failed_count++))
        fi
    done

    print_success "Updated $updated_count files successfully"
    if [[ $failed_count -gt 0 ]]; then
        print_warning "$failed_count files failed to update"
    fi
}

# Enhanced git clone with timeout and retry
clone_repository() {
    local repo_url="$1"
    local branch="$2"
    local target_dir="$3"
    local timeout=300
    local max_retries=3
    local retry_count=0

    while [[ $retry_count -lt $max_retries ]]; do
        print_info "Cloning repository (attempt $((retry_count + 1))/$max_retries)..."

        if timeout "$timeout" git clone -b "$branch" --depth 1 "$repo_url" "$target_dir" 2>/dev/null; then
            return 0
        fi

        ((retry_count++))
        if [[ $retry_count -lt $max_retries ]]; then
            print_warning "Clone attempt failed, retrying in 5 seconds..."
            sleep 5
            # Clean up partial clone
            [[ -d "$target_dir" ]] && rm -rf "$target_dir"
        fi
    done

    return 1
}

# Cross-platform file replacement using grep and sed
replace_in_files_cross_platform() {
    local old_pattern="$1"
    local new_replacement="$2"
    local project_dir="$3"

    print_info "Searching for files containing: $old_pattern"

    # Find files containing the pattern
    local files_to_update=()
    while IFS= read -r -d '' file; do
        files_to_update+=("$file")
    done < <(find "$project_dir" -type f \( -name "*.go" -o -name "*.mod" -o -name "*.md" -o -name "Dockerfile" -o -name "Makefile" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" \) -exec grep -l "$old_pattern" {} + -print0 2>/dev/null || true)

    local file_count=${#files_to_update[@]}

    if [[ $file_count -eq 0 ]]; then
        print_warning "No files found containing: $old_pattern"
        return 0
    fi

    print_info "Found $file_count files to update"

    local updated_count=0
    local failed_count=0

    # Update each file
    for file in "${files_to_update[@]}"; do
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS - use temporary file approach to avoid issues
            local temp_file="${file}.tmp.$$"
            if sed "s|$old_pattern|$new_replacement|g" "$file" > "$temp_file" && mv "$temp_file" "$file"; then
                ((updated_count++))
            else
                ((failed_count++))
                print_error "Failed to update: $file"
                [[ -f "$temp_file" ]] && rm -f "$temp_file"
            fi
        else
            # Linux and others
            if sed -i "s|$old_pattern|$new_replacement|g" "$file"; then
                ((updated_count++))
            else
                ((failed_count++))
                print_error "Failed to update: $file"
            fi
        fi
    done

    print_success "Updated $updated_count files successfully"
    if [[ $failed_count -gt 0 ]]; then
        print_warning "$failed_count files failed to update"
    fi
}

# Parse command line arguments
if [[ $# -gt 0 && ("$1" == "-h" || "$1" == "--help") ]]; then
    show_usage
    exit 0
fi

# Set default values
projectName="go-api"
projectVersion="main"

# Parse arguments with better validation
if [[ $# -gt 3 ]]; then
    print_error "Too many arguments. Use --help for usage information."
    exit 1
fi

if [[ -n "${1:-}" ]]; then
    projectName="$1"
    validate_project_name "$projectName"
fi

if [[ -n "${2:-}" ]]; then
    projectVersion="$2"
fi

# Determine module name
if [[ -n "${3:-}" ]]; then
    moduleName="$3"
    validate_module_name "$moduleName"
else
    moduleName="$projectName"
    validate_module_name "$moduleName"
fi

# Pre-flight checks
print_info "Starting pre-flight checks..."
check_file_permissions
check_dependencies
check_network

print_info "Creating project: $projectName"
print_info "Using version: $projectVersion"
print_info "Module name: $moduleName"

# Set the project directory path
projectDir="$(pwd)/$projectName"
cleanup_required=true

# Enhanced directory existence check
if [[ -d "$projectDir" ]]; then
    if [[ ! -w "$projectDir" ]]; then
        print_error "Directory '$projectName' exists but is not writable"
        exit 1
    fi

    # Check if it's already a git repository
    if [[ -d "$projectDir/.git" ]]; then
        print_warning "Directory '$projectName' already contains a git repository."
    else
        print_warning "Directory '$projectName' already exists."
    fi

    read -p "Do you want to remove it and continue? (y/N): " -r
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled."
        cleanup_required=false
        exit 0
    fi
    print_info "Removing existing directory..."
    rm -rf "$projectDir"
fi

print_info "Cloning go-api repository..."

# Clone with enhanced error handling
if ! clone_repository "https://github.com/seakee/go-api.git" "$projectVersion" "$projectDir"; then
    print_error "Failed to clone repository after multiple attempts."
    print_error "Please check:"
    print_error "  - Internet connection"
    print_error "  - Branch/tag '$projectVersion' exists"
    print_error "  - Repository access permissions"
    print_error "  - Firewall settings"
    exit 1
fi

print_success "Repository cloned successfully."

# Verify clone integrity
if [[ ! -d "$projectDir" || ! -f "$projectDir/go.mod" ]]; then
    print_error "Cloned repository appears to be incomplete"
    exit 1
fi

# Clean up git-related files and references
cleanup_git_references() {
    local project_dir="$1"
    local module_name="$2"

    print_info "Cleaning up git-related files and references..."

    # Remove git-related files
    local git_files=(
        ".git"
        ".gitignore.bak"
        ".github"
        "CONTRIBUTING.md"
        "CHANGELOG.md"
    )

    for git_file in "${git_files[@]}"; do
        if [[ -e "$project_dir/$git_file" ]]; then
            print_info "Removing: $git_file"
            rm -rf "$project_dir/$git_file"
        fi
    done

    # Clean up any remaining git references in files - use the module name
    print_info "Cleaning remaining git references in files"
    replace_in_files_cross_platform "github.com/seakee/go-api" "$module_name" "$project_dir"

    # Clean up author and contributor information
    local readme_file="$project_dir/README.md"
    if [[ -f "$readme_file" ]]; then
        print_info "Cleaning README.md author information"
        # Create a clean README template
        cat > "$readme_file" << EOF
# $projectName

A high-performance Go API project based on the go-api framework. Built for rapid development of scalable backend services with enterprise-grade features.

## Description

This project provides a robust foundation for building RESTful APIs with Go, featuring:
- Clean architecture with layered design (Model-Repository-Service-Controller)
- Built-in dependency injection and configuration management
- Multi-database support (MySQL, MongoDB)
- JWT authentication and middleware system
- Internationalization (i18n) support
- High-performance logging with structured output
- Docker containerization support

## Features

- **🚀 High Performance**: Built on Gin framework for optimal performance
- **🏗️ Clean Architecture**: Follows MVC + Repository pattern with proper separation of concerns
- **🔧 Configuration Management**: Environment-based configuration with JSON files
- **🔐 Authentication**: JWT-based authentication with middleware support
- **🗄️ Multi-Database**: Support for MySQL and MongoDB with GORM and qmgo
- **📝 Logging**: Structured logging with Zap for high performance
- **🌍 Internationalization**: Built-in i18n support for multiple languages
- **⏰ Task Scheduling**: Built-in job scheduler for background tasks
- **📡 Message Queue**: Kafka consumer support for event-driven architecture
- **🐳 Docker Ready**: Complete Docker setup for development and production
- **🛠️ Code Generation**: SQL-based code generation tools for rapid development

## Installation

### Prerequisites

- Go 1.24 or higher
- Git
- Make (optional, but recommended)
- Docker (optional, for containerized development)

### Setup

\`\`\`bash
git clone <your-repository-url>
cd $projectName
go mod download
\`\`\`

## Usage

### Development

\`\`\`bash
# Start development server
make run

# Or run directly
go run main.go
\`\`\`

### Build

\`\`\`bash
# Build for current platform
make build

# Build Docker image
make docker-build
\`\`\`

### Test

\`\`\`bash
# Run tests
make test

# Run tests with coverage
go test -cover ./...
\`\`\`

### Docker

\`\`\`bash
# Run with Docker
make docker-run

# Or manually
docker run -p 8080:8080 -v \$(pwd)/bin/configs:/bin/configs $projectName
\`\`\`

## Configuration

Update configuration files in \`bin/configs/\` directory according to your environment:

- \`dev.json\` - Development environment
- \`prod.json\` - Production environment

Example configuration structure:

\`\`\`json
{
  "system": {
    "name": "$projectName",
    "port": 8080,
    "debug": true
  },
  "database": {
    "mysql": {
      "host": "localhost",
      "port": 3306,
      "database": "your_db",
      "username": "your_user",
      "password": "your_password"
    }
  }
}
\`\`\`

## Project Structure

\`\`\`
$projectName/
├── app/
│   ├── config/          # Configuration management
│   ├── http/           # HTTP layer (controllers, middleware, routes)
│   ├── model/          # Data models
│   ├── repository/     # Data access layer
│   ├── service/        # Business logic layer
│   ├── job/            # Background jobs
│   └── pkg/            # Shared packages
├── bootstrap/          # Application bootstrap
├── bin/
│   ├── configs/        # Configuration files
│   ├── data/          # SQL files and data
│   └── lang/          # Language files for i18n
├── command/           # CLI commands and code generation
├── scripts/           # Utility scripts
├── Dockerfile         # Docker configuration
├── Makefile          # Build automation
└── main.go           # Application entry point
\`\`\`

## API Documentation

The API follows RESTful conventions. Key endpoints:

- \`GET /health\` - Health check
- \`POST /auth/login\` - User authentication
- \`GET /api/v1/*\` - API endpoints (requires authentication)

For detailed API documentation, run the server and visit the documentation endpoint or refer to the API specification files.

## Development Guide

### Adding New Features

1. **Model**: Define data structures in \`app/model/\`
2. **Repository**: Implement data access in \`app/repository/\`
3. **Service**: Add business logic in \`app/service/\`
4. **Controller**: Create HTTP handlers in \`app/http/controller/\`
5. **Routes**: Register routes in \`app/http/router/\`

### Code Generation

Use the built-in code generator to scaffold new modules:

\`\`\`bash
# Generate model and repository from SQL
go run command/codegen/handler.go -sql=your_table.sql
\`\`\`

### Testing

Write tests following Go conventions:

\`\`\`bash
# Create test files
touch app/service/your_service_test.go

# Run specific tests
go test ./app/service/
\`\`\`

## Contributing

We welcome contributions! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch: \`git checkout -b feature/amazing-feature\`
3. **Follow** the coding standards:
   - Use \`gofmt\` and \`goimports\` for formatting
   - Write comprehensive tests
   - Follow the established architecture patterns
   - Add proper documentation and comments
4. **Commit** your changes: \`git commit -m 'feat: add amazing feature'\`
5. **Push** to the branch: \`git push origin feature/amazing-feature\`
6. **Submit** a pull request

### Code Standards

- Follow Go best practices and conventions
- Maintain the layered architecture (Model → Repository → Service → Controller)
- Use dependency injection patterns
- Write comprehensive unit tests
- Document all public functions and methods
- Use conventional commit messages

### Development Setup

\`\`\`bash
# Install development dependencies
make dev-setup

# Run code quality checks
make lint
make fmt

# Run all tests
make test-all
\`\`\`

## License

This project is licensed under the [MIT License](LICENSE) - see the LICENSE file for details.

## Support

If you encounter any issues or have questions:

1. Check the [documentation](docs/)
2. Search existing [issues](../../issues)
3. Create a new issue with detailed information
4. Join our community discussions

## Acknowledgments

- Built with [go-api](https://github.com/seakee/go-api) framework
- Powered by [Gin](https://gin-gonic.com/) web framework
- Database integration with [GORM](https://gorm.io/)
- Logging with [Zap](https://go.uber.org/zap)

---

**Happy coding! 🚀**
EOF
    fi

    # Clean up any license files that reference the original project
    local license_file="$project_dir/LICENSE"
    if [[ -f "$license_file" ]]; then
        print_info "Removing original LICENSE file"
        rm -f "$license_file"
        print_info "Please add your own LICENSE file"
    fi

    print_success "Git references cleaned up"
}

# Clean up git references before processing
cleanup_git_references "$projectDir" "$moduleName"

# Only replace if project name is different from 'go-api'
if [[ "$projectName" != "go-api" ]]; then
    print_info "Updating import paths and project references..."

    # Use the simplified approach - replace 'go-api' with new project name
    replace_in_files_cross_platform "go-api" "$projectName" "$projectDir"

    print_success "Project references updated."
else
    print_info "Project name is 'go-api', skipping replacements."
fi

# Initialize new git repository with better error handling
print_info "Initializing new git repository..."
cd "$projectDir" || {
    print_error "Failed to change directory to $projectDir"
    exit 1
}

if ! git init; then
    print_error "Failed to initialize git repository"
    exit 1
fi

# Set initial branch name explicitly
if ! git checkout -b main 2>/dev/null; then
    print_warning "Could not create 'main' branch, using default branch"
fi

# Add files with verification
if ! git add .; then
    print_error "Failed to add files to git repository"
    exit 1
fi

if ! git commit -m "Initial commit: Created $projectName from go-api template"; then
    print_error "Failed to create initial commit"
    exit 1
fi

print_success "Git repository initialized successfully."

# Handle Go module operations
if [[ "$projectName" != "go-api" || "$moduleName" != "go-api" ]]; then
    print_info "Cleaning up Go module dependencies..."

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        print_warning "go.mod not found, initializing new module..."
        if ! go mod init "$moduleName"; then
            print_error "Failed to initialize Go module"
            exit 1
        fi
    else
        # Update existing go.mod file
        print_info "Updating go.mod module name..."
        temp_file="go.mod.tmp.$$"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            if sed "s|module github.com/seakee/go-api|module $moduleName|g" go.mod > "$temp_file" && mv "$temp_file" go.mod; then
                print_success "go.mod updated successfully"
            else
                print_error "Failed to update go.mod"
                [[ -f "$temp_file" ]] && rm -f "$temp_file"
                exit 1
            fi
        else
            if sed -i "s|module github.com/seakee/go-api|module $moduleName|g" go.mod; then
                print_success "go.mod updated successfully"
            else
                print_error "Failed to update go.mod"
                exit 1
            fi
        fi
    fi

    # Clean up dependencies with timeout
    if timeout 120 go mod tidy; then
        print_success "Go module dependencies updated successfully."
    else
        print_warning "Go module tidy failed or timed out."
        print_info "This might be due to network issues or missing dependencies."
        print_info "You can run 'go mod tidy' manually later."
    fi

    # Verify the module
    if ! go mod verify; then
        print_warning "Go module verification failed. You may need to review dependencies."
    fi
else
    print_info "Using default module configuration"
fi

# Final verification
print_info "Performing final verification..."
if [[ ! -f "go.mod" || ! -d ".git" ]]; then
    print_error "Project setup appears to be incomplete"
    exit 1
fi

cleanup_required=false  # Disable cleanup since project is complete

print_success "Project '$projectName' created successfully!"
print_info ""
print_info "Project location: $projectDir"
print_info ""
print_info "Clean project created with:"
print_info "  ✓ Fresh git repository (no original history)"
print_info "  ✓ Updated import paths and references"
print_info "  ✓ Go module name: $moduleName"
print_info "  ✓ Clean README.md template"
print_info "  ✓ Original LICENSE removed (add your own)"
print_info ""
print_info "Project structure:"
if command -v tree &> /dev/null; then
    tree -L 2 -a "$projectDir" | head -20
else
    ls -la "$projectDir" | head -10
fi

print_info ""
print_info "Next steps:"
print_info "  1. cd $projectName"
print_info "  2. Add your LICENSE file"
print_info "  3. Update README.md with your project details"
print_info "  4. Review and update configuration files in bin/configs/"
print_info "  5. Set up your remote git repository:"
print_info "     git remote add origin <your-repository-url>"
print_info "     git push -u origin main"
print_info "  6. Build and run the project:"
print_info "     go mod download  # Download dependencies"
print_info "     make run         # Or: go run main.go"
print_info ""
print_info "Important:"
print_info "  📝 Update README.md with your project information"
print_info "  📄 Add appropriate LICENSE file"
print_info "  ⚙️  Configure bin/configs/ files for your environment"
print_info "  🔗 Set up your own git remote repository"
print_info ""
print_info "Useful commands:"
print_info "  make help        # Show available make targets"
print_info "  go mod tidy      # Clean up dependencies"
print_info "  go test ./...    # Run tests"
print_info ""
print_success "Happy coding! 🚀"