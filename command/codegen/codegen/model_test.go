// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package codegen

import (
	"testing"
)

// TestModel_Generate tests the complete model generation process
func TestModel_Generate(t *testing.T) {
	m := NewModel()

	err := m.Generate(false, "/github.com/seakee/go-api/bin/data/sql/auth_app.sql", "")
	if err != nil {
		t.Fatal(err)
	}
}

// TestModel_parseSQL tests SQL parsing functionality
func TestModel_parseSQL(t *testing.T) {
	m := NewModel()

	// Test with actual auth_app.sql content
	sqlContent := `CREATE TABLE auth_app (
		id int NOT NULL AUTO_INCREMENT COMMENT 'id',
		app_id varchar(30) NOT NULL COMMENT '应用ID',
		app_name varchar(50) DEFAULT NULL COMMENT '应用名称',
		status tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态'
	);`

	err := m.parseSQL(sqlContent)
	if err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if m.TableName != "auth_app" {
		t.Errorf("TableName = %v, want auth_app", m.TableName)
	}

	// Debug: print all parsed fields
	t.Logf("Parsed %d fields:", len(m.TableFields))
	for i, field := range m.TableFields {
		t.Logf("Field %d: Name=%s, Type=%s, JsonName=%s, IsAutoIncrement=%v, IsNullable=%v",
			i, field.Name, field.Type, field.JsonName, field.IsAutoIncrement, field.IsNullable)
	}

	// Test that we have at least some fields
	if len(m.TableFields) == 0 {
		t.Error("No fields were parsed")
	}

	// Test that we have the expected fields (id is skipped by design)
	if len(m.TableFields) > 0 {
		// First field should be app_id
		appIdField := m.TableFields[0]
		if appIdField.Name != "AppId" {
			t.Errorf("First field name = %v, want AppId", appIdField.Name)
		}
		if appIdField.Type != "string" {
			t.Errorf("First field type = %v, want string", appIdField.Type)
		}
		if appIdField.IsNullable {
			t.Errorf("AppId field should not be nullable")
		}
	}

	// Test that we have at least 3 fields (app_id, app_name, status)
	if len(m.TableFields) < 3 {
		t.Errorf("Expected at least 3 fields, got %d", len(m.TableFields))
	}

	// Test status field properties
	if len(m.TableFields) >= 3 {
		statusField := m.TableFields[2] // status is the 3rd field
		if statusField.Name != "Status" {
			t.Errorf("Third field name = %v, want Status", statusField.Name)
		}
		if statusField.Type != "int8" {
			t.Errorf("Status field type = %v, want int8", statusField.Type)
		}
		if statusField.IsNullable {
			t.Errorf("Status field should not be nullable")
		}
		if statusField.DefaultValue != "0" {
			t.Errorf("Status field default value = %v, want 0", statusField.DefaultValue)
		}
	}
}

// TestModel_generateGormTag tests GORM tag generation
func TestModel_generateGormTag(t *testing.T) {
	m := NewModel()

	tests := []struct {
		name     string
		field    Field
		typeStr  string
		expected string
	}{
		{
			name: "varchar with size",
			field: Field{
				Name:       "AppName",
				JsonName:   "app_name",
				Size:       50,
				IsNullable: true,
			},
			typeStr:  "varchar(50)",
			expected: "column:app_name;type:varchar(50)",
		},
		{
			name: "int with auto increment",
			field: Field{
				Name:            "ID",
				JsonName:        "id",
				IsAutoIncrement: true,
				IsNullable:      false,
			},
			typeStr:  "int",
			expected: "column:id;not null;autoIncrement",
		},
		{
			name: "decimal with precision and scale",
			field: Field{
				Name:       "Price",
				JsonName:   "price",
				Size:       10,
				Scale:      2,
				IsNullable: false,
			},
			typeStr:  "decimal(10,2)",
			expected: "column:price;type:decimal(10,2);not null",
		},
		{
			name: "field with default value",
			field: Field{
				Name:         "Status",
				JsonName:     "status",
				DefaultValue: "0",
				IsNullable:   false,
			},
			typeStr:  "tinyint(1)",
			expected: "column:status;not null;default:0",
		},
		{
			name: "nullable field",
			field: Field{
				Name:       "Description",
				JsonName:   "description",
				IsNullable: true,
			},
			typeStr:  "text",
			expected: "column:description;type:text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.generateGormTag(&tt.field, tt.typeStr)
			if result != tt.expected {
				t.Errorf("generateGormTag() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestModel_generateNames tests package and struct name generation
func TestModel_generateNames(t *testing.T) {
	m := NewModel()

	tests := []struct {
		name            string
		tableName       string
		expectedPackage string
		expectedStruct  string
	}{
		{"simple table", "users", "users", "User"},
		{"table with underscore", "user_profiles", "user", "Profile"},
		{"complex table name", "auth_user_tokens", "auth", "UserToken"},
		{"single word", "products", "products", "Product"},
		{"already singular", "auth_app", "auth", "App"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packageName, structName := m.generateNames(tt.tableName)
			if packageName != tt.expectedPackage {
				t.Errorf("generateNames() packageName = %v, want %v", packageName, tt.expectedPackage)
			}
			if structName != tt.expectedStruct {
				t.Errorf("generateNames() structName = %v, want %v", structName, tt.expectedStruct)
			}
		})
	}
}

// TestModel_extractSizeAndScale tests size and scale extraction
func TestModel_extractSizeAndScale(t *testing.T) {
	m := NewModel()

	tests := []struct {
		name          string
		typeStr       string
		expectedSize  int
		expectedScale int
	}{
		{"varchar with size", "varchar(255)", 255, 0},
		{"decimal with precision and scale", "decimal(10,2)", 10, 2},
		{"int without size", "int", 0, 0},
		{"tinyint with size", "tinyint(1)", 1, 0},
		{"char with size", "char(10)", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, scale := m.extractSizeAndScale(tt.typeStr)
			if size != tt.expectedSize {
				t.Errorf("extractSizeAndScale() size = %v, want %v", size, tt.expectedSize)
			}
			if scale != tt.expectedScale {
				t.Errorf("extractSizeAndScale() scale = %v, want %v", scale, tt.expectedScale)
			}
		})
	}
}

// TestModel_extractComment tests comment extraction
func TestModel_extractComment(t *testing.T) {
	m := NewModel()

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"single quote comment", "id int NOT NULL COMMENT 'Primary key'", "// Primary key"},
		{"double quote comment", `name varchar(50) COMMENT "User name"`, "// User name"},
		{"chinese comment", "status tinyint(1) COMMENT '状态'", "// 状态"},
		{"no comment", "created_at timestamp", ""},
		{"comment with special chars", "desc text COMMENT 'Description with, special chars!'", "// Description with, special chars!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.extractComment(tt.line)
			if result != tt.expected {
				t.Errorf("extractComment() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestModel_extractDefaultValue tests default value extraction
func TestModel_extractDefaultValue(t *testing.T) {
	m := NewModel()

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"string default", "status tinyint(1) DEFAULT '0'", "0"},
		{"null default", "name varchar(50) DEFAULT NULL", "NULL"},
		{"current timestamp", "created_at timestamp DEFAULT CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"},
		{"no default", "id int NOT NULL", ""},
		{"numeric default", "count int DEFAULT 100", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.extractDefaultValue(tt.line)
			if result != tt.expected {
				t.Errorf("extractDefaultValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestModel_getGoType tests Go type mapping functionality
func TestModel_getGoType(t *testing.T) {
	m := NewModel()

	tests := []struct {
		name       string
		sqlType    string
		isUnsigned bool
		expected   string
	}{
		{"int", "int", false, "int32"},
		{"int unsigned", "int", true, "uint32"},
		{"tinyint", "tinyint", false, "int8"},
		{"tinyint unsigned", "tinyint", true, "uint8"},
		{"smallint", "smallint", false, "int16"},
		{"smallint unsigned", "smallint", true, "uint16"},
		{"bigint", "bigint", false, "int64"},
		{"bigint unsigned", "bigint", true, "uint64"},
		{"varchar", "varchar", false, "string"},
		{"text", "text", false, "string"},
		{"tinytext", "tinytext", false, "string"},
		{"mediumtext", "mediumtext", false, "string"},
		{"longtext", "longtext", false, "string"},
		{"decimal", "decimal", false, "decimal.Decimal"},
		{"float", "float", false, "float32"},
		{"double", "double", false, "float64"},
		{"timestamp", "timestamp", false, "time.Time"},
		{"datetime", "datetime", false, "time.Time"},
		{"date", "date", false, "time.Time"},
		{"time", "time", false, "time.Time"},
		{"blob", "blob", false, "[]byte"},
		{"tinyblob", "tinyblob", false, "[]byte"},
		{"mediumblob", "mediumblob", false, "[]byte"},
		{"longblob", "longblob", false, "[]byte"},
		{"json", "json", false, "datatypes.JSON"},
		{"unknown", "unknown", false, "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := m.getGoType(tt.sqlType, tt.isUnsigned)
			if result != tt.expected {
				t.Errorf("getGoType(%v, %v) = %v, want %v", tt.sqlType, tt.isUnsigned, result, tt.expected)
			}
		})
	}
}
