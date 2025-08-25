#!/usr/bin/env bash

# Generate a go-api project
# This script creates a new project based on the go-api template
#
# Usage:
#   ./generate.sh [project-name] [version]
#   ./generate.sh my-project v1.0.0
#   ./generate.sh my-project        # Uses latest main branch
#   ./generate.sh                   # Creates 'go-api' project from main branch
#
# Parameters:
#   $1 - Project name (optional, default: "go-api")
#   $2 - Version/branch (optional, default: "main")

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

# Print colored output
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

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
    echo "Usage: $0 [project-name] [version]"
    echo ""
    echo "Parameters:"
    echo "  project-name  Name of the new project (default: go-api)"
    echo "  version       Git branch or tag to use (default: main)"
    echo ""
    echo "Examples:"
    echo "  $0                           # Create 'go-api' from main branch"
    echo "  $0 my-awesome-api            # Create 'my-awesome-api' from main branch"
    echo "  $0 my-api v1.2.0             # Create 'my-api' from tag v1.2.0"
    echo "  $0 my-api feature/new-auth   # Create 'my-api' from feature branch"
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

# Parse command line arguments
if [[ "$1" == "-h" || "$1" == "--help" ]] 2>/dev/null; then
    show_usage
    exit 0
fi

# Set default values
projectName="go-api"
projectVersion="main"

# Parse arguments with better validation
if [[ $# -gt 2 ]]; then
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

# Pre-flight checks
print_info "Starting pre-flight checks..."
check_file_permissions
check_dependencies
check_network

print_info "Creating project: $projectName"
print_info "Using version: $projectVersion"

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

# Remove the .git directory to detach from the original repository
print_info "Removing original git history..."
rm -rf "$projectDir"/.git

# Only replace if project name is different from 'go-api'
if [[ "$projectName" != "go-api" ]]; then
    print_info "Updating import paths and project references..."

    # Use the simplified approach - replace 'go-api' with new project name
    replace_in_files "go-api" "$projectName" "$projectDir"

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
if [[ "$projectName" != "go-api" ]]; then
    print_info "Cleaning up Go module dependencies..."

    # Check if go.mod exists
    if [[ ! -f "go.mod" ]]; then
        print_warning "go.mod not found, initializing new module..."
        if ! go mod init "$projectName"; then
            print_error "Failed to initialize Go module"
            exit 1
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
print_info "Project structure:"
if command -v tree &> /dev/null; then
    tree -L 2 -a "$projectDir" | head -20
else
    ls -la "$projectDir" | head -10
fi

print_info ""
print_info "Next steps:"
print_info "  1. cd $projectName"
print_info "  2. Review and update configuration files in bin/configs/"
print_info "  3. Update README.md with your project details"
print_info "  4. Set up your remote git repository:"
print_info "     git remote add origin <your-repository-url>"
print_info "     git push -u origin main"
print_info "  5. Build and run the project:"
print_info "     go mod download  # Download dependencies"
print_info "     make run         # Or: go run main.go"
print_info ""
print_info "Useful commands:"
print_info "  make help        # Show available make targets"
print_info "  go mod tidy      # Clean up dependencies"
print_info "  go test ./...    # Run tests"
print_info ""
print_success "Happy coding! 🚀"