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

func assertGeneratedRepositoryCodeFormats(t *testing.T, repo *Repository, code string) {
	t.Helper()

	if _, err := repo.formatGoCode(code); err != nil {
		t.Fatalf("generated repository code is not valid Go: %v\n%s", err, code)
	}
}

func mustParseRepositoryTestModel(t *testing.T, dialect, sqlContent string) *Model {
	t.Helper()

	model := NewModelWithDialect(dialect)
	if err := model.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	return model
}

func TestRepository_generateCode(t *testing.T) {
	// Create a test model
	model := &Model{
		PackageName: "auth",
		StructName:  "App",
		TableName:   "auth_app",
		IDType:      "uint",
		TableFields: []Field{
			{
				Name:     "ID",
				JsonName: "id",
				Type:     "uint",
				Comment:  "// Primary key",
				GormTag:  `gorm:"primaryKey;autoIncrement"`,
			},
			{
				Name:     "AppName",
				JsonName: "app_name",
				Type:     "string",
				Comment:  "// Application name",
				GormTag:  `gorm:"type:varchar(255);not null"`,
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
		"type AppRepo interface",
		"type appRepo struct",
		"NewAppRepo",
		"func (r *appRepo) Create",
		"func (r *appRepo) GetByID",
		"func (r *appRepo) Update",
		"func (r *appRepo) UpdateFields",
		"func (r *appRepo) Delete",
		"func (r *appRepo) List",
		"gorm.io/gorm",
		"context",
		"github.com/sk-pkg/redis",
	}

	for _, element := range expectedElements {
		if !strings.Contains(code, element) {
			t.Errorf("Generated code does not contain expected element: %s", element)
		}
	}
	assertGeneratedRepositoryCodeFormats(t, repo, code)
}

func TestRepository_WriteRepositoryFile(t *testing.T) {
	// Create a test model
	model := &Model{
		PackageName: "test",
		StructName:  "TestModel",
		TableName:   "test_model",
		IDType:      "uint",
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
		IDType:      "uint",
		TableFields: []Field{
			{
				Name:     "ID",
				JsonName: "id",
				Type:     "uint",
				Comment:  "// Primary key",
				GormTag:  `gorm:"primaryKey;autoIncrement"`,
			},
			{
				Name:     "AppName",
				JsonName: "app_name",
				Type:     "string",
				Comment:  "// Application name",
				GormTag:  `gorm:"type:varchar(255);not null"`,
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
		"NewAppRepo",
		"Create",
		"GetByID",
		"Update",
		"Delete",
		"List",
		"UpdateFields",
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
		IDType:      "uint",
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

func TestRepository_GenerateCodeUsesCustomIDType(t *testing.T) {
	model := &Model{
		PackageName:    "job",
		StructName:     "Task",
		TableName:      "job_task",
		IDType:         "int64",
		PrimaryKeyName: "id",
		TableFields: []Field{
			{Name: "ID", JsonName: "id", Type: "int64"},
			{Name: "Name", JsonName: "name", Type: "string"},
		},
	}

	repo := NewRepository(model)
	code, err := repo.generateCode()
	if err != nil {
		t.Fatalf("Failed to generate repository code: %v", err)
	}

	if !strings.Contains(code, "GetByID(ctx context.Context, id int64)") {
		t.Fatalf("expected generated repository to use int64 IDs, got:\n%s", code)
	}
	assertGeneratedRepositoryCodeFormats(t, repo, code)
}

func TestRepository_GenerateCodeUsesPostgresPrimaryKeyName(t *testing.T) {
	model := &Model{
		PackageName:    "oauth",
		StructName:     "App",
		TableName:      "oauth_apps",
		IDType:         "string",
		PrimaryKeyName: "app_uuid",
		TableFields: []Field{
			{Name: "AppUuid", JsonName: "app_uuid", Type: "string", IsPrimaryKey: true},
			{Name: "Name", JsonName: "name", Type: "string"},
		},
	}

	repo := NewRepository(model)
	code, err := repo.generateCode()
	if err != nil {
		t.Fatalf("Failed to generate repository code: %v", err)
	}

	if !strings.Contains(code, `Where("app_uuid = ?", id)`) {
		t.Fatalf("expected generated repository to query by postgres primary key, got:\n%s", code)
	}
	if !strings.Contains(code, "updateModel.AppUuid = id") {
		t.Fatalf("expected generated repository to assign custom primary key field, got:\n%s", code)
	}
	assertGeneratedRepositoryCodeFormats(t, repo, code)
}

func TestRepository_GenerateCodeUsesImplicitIDField(t *testing.T) {
	model := mustParseRepositoryTestModel(t, "postgres", `CREATE TABLE logs (
		id bigint generated always as identity,
		message text NOT NULL
	);`)

	repo := NewRepository(model)
	code, err := repo.generateCode()
	if err != nil {
		t.Fatalf("Failed to generate repository code: %v", err)
	}

	if !strings.Contains(code, `Where("id = ?", id)`) {
		t.Fatalf("expected generated repository to query by implicit id field, got:\n%s", code)
	}
	if !strings.Contains(code, "updateModel.Id = id") {
		t.Fatalf("expected generated repository to assign implicit id field, got:\n%s", code)
	}
	assertGeneratedRepositoryCodeFormats(t, repo, code)
}

func TestRepository_GenerateCodeRejectsCompositePrimaryKey(t *testing.T) {
	model := mustParseRepositoryTestModel(t, "postgres", `CREATE TABLE logs (
		tenant_id uuid NOT NULL,
		id bigint NOT NULL,
		message text NOT NULL,
		PRIMARY KEY (tenant_id, id)
	);`)

	repo := NewRepository(model)
	_, err := repo.generateCode()
	if err == nil {
		t.Fatal("expected repository generation to reject composite primary keys")
	}
	if !strings.Contains(err.Error(), "does not support composite primary keys") {
		t.Fatalf("expected composite primary key error, got %v", err)
	}
}

func TestRepository_GenerateCodeRejectsMissingIdentifier(t *testing.T) {
	model := mustParseRepositoryTestModel(t, "postgres", `CREATE TABLE logs (
		message text NOT NULL
	);`)

	repo := NewRepository(model)
	_, err := repo.generateCode()
	if err == nil {
		t.Fatal("expected repository generation to reject tables without a single identifier field")
	}
	if !strings.Contains(err.Error(), "requires a single-column primary key or id field") {
		t.Fatalf("expected missing identifier error, got %v", err)
	}
}
