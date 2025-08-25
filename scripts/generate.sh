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

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

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
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        print_error "Please install them and try again."
        exit 1
    fi
    
    print_success "All dependencies are installed."
}

# Cross-platform sed replacement
sed_replace() {
    local pattern="$1"
    local replacement="$2"
    local file="$3"
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i "" "s|$pattern|$replacement|g" "$file"
    else
        # Linux and others
        sed -i "s|$pattern|$replacement|g" "$file"
    fi
}

# Parse command line arguments
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    show_usage
    exit 0
fi

# Set default values for project name and version
projectName="go-api"
projectVersion="main"

# If a project name is provided as the first argument, use it
if [[ -n "$1" ]]; then
    projectName="$1"
    validate_project_name "$projectName"
fi

# If a version is provided as the second argument, use it
if [[ -n "$2" ]]; then
    projectVersion="$2"
fi

# Check dependencies
check_dependencies

print_info "Creating project: $projectName"
print_info "Using version: $projectVersion"

# Set the project directory path
projectDir="$(pwd)/$projectName"

# Check if project directory already exists
if [[ -d "$projectDir" ]]; then
    print_warning "Directory '$projectName' already exists."
    read -p "Do you want to remove it and continue? (y/N): " -r
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled."
        exit 0
    fi
    print_info "Removing existing directory..."
    rm -rf "$projectDir"
fi

print_info "Cloning go-api repository..."

# Clone the go-api repository from GitHub
if ! git clone -b "$projectVersion" https://github.com/seakee/go-api.git "$projectDir"; then
    print_error "Failed to clone repository. Please check:"
    print_error "  - Internet connection"
    print_error "  - Branch/tag '$projectVersion' exists"
    print_error "  - Repository access permissions"
    exit 1
fi

print_success "Repository cloned successfully."

# Remove the .git directory to detach from the original repository
print_info "Removing original git history..."
rm -rf "$projectDir"/.git

# Only replace if project name is different from 'go-api'
if [[ "$projectName" != "go-api" ]]; then
    print_info "Updating import paths and project references..."
    
    # Find all Go files and relevant config files for replacement
    find "$projectDir" -type f \( -name "*.go" -o -name "*.mod" -o -name "*.md" -o -name "Dockerfile" -o -name "Makefile" \) -print0 | while IFS= read -r -d '' file; do
        # Replace import paths
        if grep -q 'github.com/seakee/go-api' "$file" 2>/dev/null; then
            sed_replace 'github.com/seakee/go-api' "$projectName" "$file"
        fi
        
        # Replace project name references (but be more careful)
        if grep -q '\bgo-api\b' "$file" 2>/dev/null; then
            sed_replace '\bgo-api\b' "$projectName" "$file"
        fi
    done
    
    print_success "Project references updated."
else
    print_info "Project name is 'go-api', skipping replacements."
fi

# Initialize new git repository
print_info "Initializing new git repository..."
cd "$projectDir"
git init
git add .
git commit -m "Initial commit: Created $projectName from go-api template"

# Initialize go module if project name changed
if [[ "$projectName" != "go-api" ]]; then
    print_info "Cleaning up Go module dependencies..."
    go mod tidy
    print_success "Go module dependencies updated."
fi

print_success "Project '$projectName' created successfully!"
print_info ""
print_info "Next steps:"
print_info "  1. cd $projectName"
print_info "  2. Review and update configuration files in bin/configs/"
print_info "  3. Update README.md with your project details"
print_info "  4. Set up your remote git repository:"
print_info "     git remote add origin <your-repository-url>"
print_info "     git push -u origin main"
print_info "  5. Run the project:"
print_info "     make run"
print_info ""
print_success "Happy coding! 🚀"