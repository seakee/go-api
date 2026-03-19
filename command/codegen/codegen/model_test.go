// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package codegen

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestModel_Generate tests the complete model generation process
func TestModel_Generate(t *testing.T) {
	m := NewModel()
	outputDir := t.TempDir()

	err := m.Generate(true, "../../../bin/data/sql/mysql/auth_app.sql", outputDir)
	if err != nil {
		t.Fatal(err)
	}

	generatedPath := filepath.Join(outputDir, "auth", "app.go")
	if _, err := os.Stat(generatedPath); err != nil {
		t.Fatalf("generated file missing: %v", err)
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
		status tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态',
		created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at timestamp NULL DEFAULT NULL
	);`

	err := m.parseSQL(sqlContent)
	if err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if m.TableName != "auth_app" {
		t.Errorf("TableName = %v, want auth_app", m.TableName)
	}

	if !m.UseGormModel {
		t.Error("expected auth_app to embed gorm.Model")
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

	// Test that we have the expected fields.
	if len(m.TableFields) > 0 {
		appIdField := m.TableFields[1]
		if appIdField.Name != "AppId" {
			t.Errorf("AppId field name = %v, want AppId", appIdField.Name)
		}
		if appIdField.Type != "string" {
			t.Errorf("AppId field type = %v, want string", appIdField.Type)
		}
		if appIdField.IsNullable {
			t.Errorf("AppId field should not be nullable")
		}
	}

	// Standard GORM fields are parsed, then filtered at template time.
	if len(m.TableFields) < 4 {
		t.Errorf("Expected parsed fields to include standard columns, got %d", len(m.TableFields))
	}

	// Test status field properties
	if len(m.TableFields) >= 4 {
		statusField := m.TableFields[3]
		if statusField.Name != "Status" {
			t.Errorf("Status field name = %v, want Status", statusField.Name)
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

func TestModel_parseSQLWithoutStandardFields(t *testing.T) {
	m := NewModel()

	sqlContent := `CREATE TABLE custom_job (
		id bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
		name varchar(64) DEFAULT NULL COMMENT 'name',
		run_at timestamp NULL DEFAULT NULL COMMENT 'run time'
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if m.UseGormModel {
		t.Fatal("custom_job should not embed gorm.Model")
	}

	fields := m.templateFields()
	if len(fields) != 3 {
		t.Fatalf("expected 3 template fields, got %d", len(fields))
	}

	if fields[1].Type != "*string" {
		t.Fatalf("nullable varchar should map to *string, got %s", fields[1].Type)
	}

	if fields[2].Type != "*time.Time" {
		t.Fatalf("nullable timestamp should map to *time.Time, got %s", fields[2].Type)
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
			name: "field with null default skips tag",
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
		{"null default", "name varchar(50) DEFAULT NULL", ""},
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
		isNullable bool
		expected   string
	}{
		{"int", "int", false, false, "int32"},
		{"int unsigned", "int", true, false, "uint32"},
		{"tinyint", "tinyint", false, false, "int8"},
		{"tinyint unsigned", "tinyint", true, false, "uint8"},
		{"smallint", "smallint", false, false, "int16"},
		{"smallint unsigned", "smallint", true, false, "uint16"},
		{"bigint", "bigint", false, false, "int64"},
		{"bigint unsigned", "bigint", true, false, "uint64"},
		{"nullable bigint", "bigint", false, true, "*int64"},
		{"varchar", "varchar", false, false, "string"},
		{"nullable varchar", "varchar", false, true, "*string"},
		{"text", "text", false, false, "string"},
		{"tinytext", "tinytext", false, false, "string"},
		{"mediumtext", "mediumtext", false, false, "string"},
		{"longtext", "longtext", false, false, "string"},
		{"decimal", "decimal", false, false, "decimal.Decimal"},
		{"nullable decimal", "decimal", false, true, "*decimal.Decimal"},
		{"float", "float", false, false, "float32"},
		{"double", "double", false, false, "float64"},
		{"timestamp", "timestamp", false, false, "time.Time"},
		{"nullable timestamp", "timestamp", false, true, "*time.Time"},
		{"datetime", "datetime", false, false, "time.Time"},
		{"date", "date", false, false, "time.Time"},
		{"time", "time", false, false, "time.Time"},
		{"blob", "blob", false, false, "[]byte"},
		{"tinyblob", "tinyblob", false, false, "[]byte"},
		{"mediumblob", "mediumblob", false, false, "[]byte"},
		{"longblob", "longblob", false, false, "[]byte"},
		{"json", "json", false, false, "datatypes.JSON"},
		{"nullable json", "json", false, true, "*datatypes.JSON"},
		{"unknown", "unknown", false, false, "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := m.getGoType(tt.sqlType, tt.isUnsigned, tt.isNullable)
			if result != tt.expected {
				t.Errorf("getGoType(%v, %v) = %v, want %v", tt.sqlType, tt.isUnsigned, result, tt.expected)
			}
		})
	}
}

func TestModel_getGoTypePostgres(t *testing.T) {
	m := NewModelWithDialect("postgres")

	tests := []struct {
		name       string
		sqlType    string
		isNullable bool
		expected   string
	}{
		{"serial", "serial", false, "int32"},
		{"bigserial", "bigserial", false, "int64"},
		{"uuid", "uuid", false, "string"},
		{"jsonb", "jsonb", false, "datatypes.JSON"},
		{"nullable timestamptz", "timestamptz", true, "*time.Time"},
		{"double precision", "double precision", false, "float64"},
		{"bytea", "bytea", false, "[]byte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := m.getGoType(tt.sqlType, false, tt.isNullable)
			if result != tt.expected {
				t.Fatalf("getGoType(%q) = %s, want %s", tt.sqlType, result, tt.expected)
			}
		})
	}
}

func TestModel_generateCodeWithoutGormModel(t *testing.T) {
	m := NewModel()
	sqlContent := `CREATE TABLE job (
		id bigint NOT NULL AUTO_INCREMENT COMMENT 'id',
		name varchar(64) DEFAULT NULL COMMENT 'name'
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if strings.Contains(code, "gorm.Model") {
		t.Fatal("unexpected gorm.Model embed for partial standard table")
	}

	if !strings.Contains(code, "Id int64") {
		t.Fatal("expected explicit ID field in generated code")
	}
}

func TestModel_parseSQLPostgres(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE IF NOT EXISTS "public"."oauth_apps" (
		"app_uuid" uuid PRIMARY KEY,
		"name" varchar(80) NOT NULL,
		"payload" jsonb,
		"created_at" timestamp with time zone NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if m.SchemaName != "public" {
		t.Fatalf("SchemaName = %s, want public", m.SchemaName)
	}
	if m.TableName != "oauth_apps" {
		t.Fatalf("TableName = %s, want oauth_apps", m.TableName)
	}
	if m.PrimaryKeyName != "app_uuid" {
		t.Fatalf("PrimaryKeyName = %s, want app_uuid", m.PrimaryKeyName)
	}
	if m.IDType != "string" {
		t.Fatalf("IDType = %s, want string", m.IDType)
	}
	if m.UseGormModel {
		t.Fatal("postgres sample should not embed gorm.Model")
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if !strings.Contains(code, `return "public.oauth_apps"`) {
		t.Fatalf("expected qualified postgres table name, got:\n%s", code)
	}
	if !strings.Contains(code, "AppUuid string") {
		t.Fatalf("expected uuid primary key field in generated code, got:\n%s", code)
	}
	if !strings.Contains(code, "Payload *datatypes.JSON") {
		t.Fatalf("expected jsonb field mapping in generated code, got:\n%s", code)
	}
}

func TestModel_parseSQLPostgresTableLevelPrimaryKeyRecomputesType(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE oauth_apps (
		app_uuid uuid,
		name varchar(80) NOT NULL,
		PRIMARY KEY (app_uuid)
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if m.PrimaryKeyName != "app_uuid" {
		t.Fatalf("PrimaryKeyName = %s, want app_uuid", m.PrimaryKeyName)
	}
	if m.IDType != "string" {
		t.Fatalf("IDType = %s, want string", m.IDType)
	}

	var pkField *Field
	for i := range m.TableFields {
		if m.TableFields[i].JsonName == "app_uuid" {
			pkField = &m.TableFields[i]
			break
		}
	}
	if pkField == nil {
		t.Fatal("primary key field app_uuid not found")
	}
	if pkField.Type != "string" {
		t.Fatalf("pkField.Type = %s, want string", pkField.Type)
	}
	if pkField.IsNullable {
		t.Fatal("primary key field should be non-nullable")
	}
}

func TestModel_generateCodeCreateUsesTypedZeroValueOnError(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE oauth_apps (
		app_uuid uuid PRIMARY KEY,
		name varchar(80) NOT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if !strings.Contains(code, "return *new(string), fmt.Errorf(\"create failed: %w\", err)") {
		t.Fatalf("expected create error branch to return typed zero value, got:\n%s", code)
	}
}

func TestModel_generateCodeWithoutPrimaryKeyCreateNoInvalidField(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE logs (
		message text NOT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if strings.Contains(code, ".ID, nil") {
		t.Fatalf("create branch should not return missing ID field when table has no primary key, got:\n%s", code)
	}
	if !strings.Contains(code, "return *new(uint), nil") {
		t.Fatalf("create branch should return typed zero value when no primary key, got:\n%s", code)
	}
}

func TestModel_generateCodeCreateUsesImplicitIDField(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE logs (
		id bigint generated always as identity,
		message text NOT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if m.HasPrimaryKey {
		t.Fatal("implicit id should not be treated as explicit primary key metadata")
	}
	if m.IDType != "int64" {
		t.Fatalf("IDType = %s, want int64", m.IDType)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if !strings.Contains(code, "func (l *Log) Create(ctx context.Context, db *gorm.DB) (int64, error)") {
		t.Fatalf("expected create signature to use implicit id type, got:\n%s", code)
	}
	matched, err := regexp.MatchString(`return l\.(ID|Id), nil`, code)
	if err != nil {
		t.Fatalf("regexp.MatchString() error = %v", err)
	}
	if !matched {
		t.Fatalf("expected create branch to return implicit id field, got:\n%s", code)
	}
	if !strings.Contains(code, `query.Order("id desc").First(&log).Error`) {
		t.Fatalf("expected Last to order by implicit id field, got:\n%s", code)
	}
}

func TestModel_generateCodeWithoutPrimaryKeyOrIDLastReturnsError(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE logs (
		message text NOT NULL,
		created_at bigint NOT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if strings.Contains(code, "query.Last(&log).Error") {
		t.Fatalf("Last should not fall back to query.Last without an ordering field, got:\n%s", code)
	}
	if !strings.Contains(code, `return nil, fmt.Errorf("find last failed: no primary key or id field available for ordering")`) {
		t.Fatalf("expected Last to return a clear error when no ordering field exists, got:\n%s", code)
	}
}

func TestModel_generateCodeCreateUsesEmbeddedGormModelIDForImplicitID(t *testing.T) {
	m := NewModelWithDialect("mysql")
	sqlContent := `CREATE TABLE auth_app (
		id int NOT NULL AUTO_INCREMENT,
		app_id varchar(30) NOT NULL,
		created_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at timestamp NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at timestamp NULL DEFAULT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if !m.UseGormModel {
		t.Fatal("expected standard columns to embed gorm.Model")
	}
	if m.IDType != "uint" {
		t.Fatalf("IDType = %s, want uint", m.IDType)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if !strings.Contains(code, "func (a *App) Create(ctx context.Context, db *gorm.DB) (uint, error)") {
		t.Fatalf("expected create signature to use embedded gorm.Model ID type, got:\n%s", code)
	}
	if !strings.Contains(code, "return a.ID, nil") {
		t.Fatalf("expected create branch to return embedded gorm.Model ID, got:\n%s", code)
	}
}

func TestModel_generateCodeListByArgsUsesPrimaryKeyOrder(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE oauth_apps (
		app_uuid uuid PRIMARY KEY,
		name varchar(80) NOT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	code, err := m.generateCode()
	if err != nil {
		t.Fatalf("generateCode() error = %v", err)
	}

	if !strings.Contains(code, `queryBuilder = queryBuilder.Order("app_uuid desc")`) {
		t.Fatalf("expected ListByArgs to use parsed primary key for order, got:\n%s", code)
	}
	if strings.Contains(code, `Order("id desc")`) {
		t.Fatalf("expected ListByArgs not to hardcode id ordering, got:\n%s", code)
	}
}

func TestModel_generateCodeEmptyTableNameReturnsError(t *testing.T) {
	m := NewModel()

	if _, err := m.generateCode(); err == nil {
		t.Fatal("expected generateCode() to return error when table name is empty")
	}
}

func TestModel_getGoTypePostgresCharacterVarying(t *testing.T) {
	m := NewModelWithDialect("postgres")

	result, _ := m.getGoType("character varying(80)", false, false)
	if result != "string" {
		t.Fatalf("getGoType(character varying(80)) = %s, want string", result)
	}

	nullableResult, _ := m.getGoType("character varying(80)", false, true)
	if nullableResult != "*string" {
		t.Fatalf("getGoType(nullable character varying(80)) = %s, want *string", nullableResult)
	}
}

func TestModel_parseSQLPostgresFieldWithCollate(t *testing.T) {
	m := NewModelWithDialect("postgres")
	sqlContent := `CREATE TABLE "public"."users" (
		"name" character varying(80) COLLATE "en_US" NOT NULL
	);`

	if err := m.parseSQL(sqlContent); err != nil {
		t.Fatalf("parseSQL() error = %v", err)
	}

	if len(m.TableFields) != 1 {
		t.Fatalf("expected 1 parsed field, got %d", len(m.TableFields))
	}

	field := m.TableFields[0]
	if field.JsonName != "name" {
		t.Fatalf("field.JsonName = %s, want name", field.JsonName)
	}
	if field.Type != "string" {
		t.Fatalf("field.Type = %s, want string", field.Type)
	}
}
