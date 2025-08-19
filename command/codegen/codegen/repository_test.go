// Copyright 2025 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepository_generateCode(t *testing.T) {
	// Create a test model
	model := &Model{
		PackageName: "auth",
		StructName:  "App",
		TableName:   "auth_app",
		TableFields: []Field{
			{
				Name:    "ID",
				Type:    "uint",
				Comment: "// Primary key",
				GormTag: `gorm:"primaryKey;autoIncrement"`,
			},
			{
				Name:    "AppName",
				Type:    "string",
				Comment: "// Application name",
				GormTag: `gorm:"type:varchar(255);not null"`,
			},
		},
	}

	// Create repository instance
	repo := NewRepository(model)

	// Generate code
	code, err := repo.generateCode()
	if err != nil {
		t.Fatalf("Failed to generate repository code: %v", err)
	}

	// Verify the generated code contains expected elements
	expectedElements := []string{
		"package auth",
		"type AppRepository struct",
		"NewAppRepository",
		"func (a *AppRepository) Create",
		"func (a *AppRepository) GetByID",
		"func (a *AppRepository) Update",
		"func (a *AppRepository) Delete",
		"func (a *AppRepository) List",
		"func (a *AppRepository) Count",
		"gorm.io/gorm",
		"context",
		"fmt",
		"errors",
	}

	for _, element := range expectedElements {
		if !strings.Contains(code, element) {
			t.Errorf("Generated code does not contain expected element: %s", element)
		}
	}
}

func TestRepository_WriteRepositoryFile(t *testing.T) {
	// Create a test model
	model := &Model{
		PackageName: "test",
		StructName:  "TestModel",
		TableName:   "test_model",
	}

	// Create repository instance
	repo := NewRepository(model)

	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Test content
	testContent := `package test

type TestModelRepository struct {
	db *gorm.DB
}
`

	// Write repository file
	err := repo.WriteRepositoryFile(true, tempDir, testContent)
	if err != nil {
		t.Fatalf("Failed to write repository file: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tempDir, "test", "model.go")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Repository file was not created at expected path: %s", expectedPath)
	}

	// Read and verify file content
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read repository file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("File content does not match expected content")
	}
}

func TestRepository_Generate(t *testing.T) {
	// Create a test model with complete data
	model := &Model{
		PackageName: "auth",
		StructName:  "App",
		TableName:   "auth_app",
		TableFields: []Field{
			{
				Name:    "ID",
				Type:    "uint",
				Comment: "// Primary key",
				GormTag: `gorm:"primaryKey;autoIncrement"`,
			},
			{
				Name:    "AppName",
				Type:    "string",
				Comment: "// Application name",
				GormTag: `gorm:"type:varchar(255);not null"`,
			},
		},
	}

	// Create repository instance
	repo := NewRepository(model)

	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Generate repository
	err := repo.Generate(true, tempDir)
	if err != nil {
		t.Fatalf("Failed to generate repository: %v", err)
	}

	// Verify file was created
	expectedPath := filepath.Join(tempDir, "auth", "app.go")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Repository file was not created at expected path: %s", expectedPath)
	}

	// Read and verify file content contains expected methods
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read repository file: %v", err)
	}

	contentStr := string(content)
	expectedMethods := []string{
		"NewAppRepository",
		"Create",
		"GetByID",
		"Update",
		"UpdateFields",
		"Delete",
		"List",
		"ListWithPagination",
		"Count",
		"FindByCondition",
		"FirstByCondition",
		"BatchCreate",
		"BatchDelete",
		"Exists",
		"Transaction",
	}

	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, method) {
			t.Errorf("Generated repository does not contain expected method: %s", method)
		}
	}
}

func TestNewRepository(t *testing.T) {
	// Create a test model
	model := &Model{
		PackageName: "test",
		StructName:  "TestModel",
		TableName:   "test_model",
	}

	// Create repository instance
	repo := NewRepository(model)

	// Verify repository was created correctly
	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}

	if repo.Model != model {
		t.Error("Repository model reference is incorrect")
	}

	if repo.Imports == nil {
		t.Error("Repository imports map is nil")
	}
}
