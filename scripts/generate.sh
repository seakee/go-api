#!/usr/bin/env bash

# Generate a go-api project
# $1 is the new project name, e.g., "project-name". If not provided, the project name will be "go-api"
# $2 is the version number, e.g., "v1.0.0". If not provided, it will fetch the latest code from the main branch

# Set default values for project name and version
projectName="go-api"
projectVersion="main"

# If a project name is provided as the first argument, use it
if [ "$1" != "" ]; then
  projectName="$1"
fi

# If a version is provided as the second argument, use it
if [ "$2" != "" ]; then
  projectVersion="$2"
fi

# Set the project directory path
projectDir=$(pwd)/$projectName

# Remove the project directory if it already exists
rm -rf "$projectDir"

# Clone the go-api repository from GitHub
# Use the specified branch or tag (projectVersion)
git clone -b "$projectVersion" https://github.com/seakee/go-api.git "$projectDir"

# Remove the .git directory to detach from the original repository
rm -rf "$projectDir"/.git

# Replace all occurrences of 'github.com/seakee/go-api' with the new project name
# This updates import paths in the Go files
grep -rl 'github.com/seakee/go-api' "$projectDir" | xargs sed -i "" "s/github.com\/seakee\/go-api/$projectName/g"

# Replace all occurrences of 'go-api' with the new project name
# This updates any remaining references to the original project name
grep -rl 'go-api' "$projectDir" | xargs sed -i "" "s/go-api/$projectName/g"

# Print a success message
echo "Success！"