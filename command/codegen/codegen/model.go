// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package codegen

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

const defaultModelOutPath = "app/model"

// Field represents a field in a database table.
type Field struct {
	Name            string // The name of the field in Go struct
	SQLType         string // The original SQL type definition
	Type            string // The Go type of the field
	JsonName        string // The JSON name of the field
	GormTag         string // The GORM tag for the field
	Comment         string // Comment associated with the field
	IsNullable      bool   // Whether the field can be null
	DefaultValue    string // Default value of the field
	Size            int    // Size for varchar, decimal precision, etc.
	Scale           int    // Scale for decimal types
	IsUnsigned      bool   // Whether the field is unsigned (for numeric types)
	IsAutoIncrement bool   // Whether the field is auto increment
	IsPrimaryKey    bool   // Whether the field is part of the primary key
}

// Model represents the structure of a database table.
type Model struct {
	PackageName    string              // The package name for the generated Go file
	Imports        map[string]struct{} // Imports required by the Go file
	StructName     string              // The name of the Go struct
	TableName      string              // The unqualified table name
	SchemaName     string              // Optional schema name
	TableFields    []Field             // The fields of the table
	UseGormModel   bool                // Whether the generated model can embed gorm.Model
	IDType         string              // Primary key type used by generated methods
	PrimaryKeyName string              // Primary key column name
	HasPrimaryKey  bool                // Whether the table defines a single-column primary key
	Dialect        string              // SQL dialect used by the parser
}

// NewModel creates a new instance of Model.
func NewModel() *Model {
	return NewModelWithDialect("mysql")
}

// NewModelWithDialect creates a new Model for the given SQL dialect.
func NewModelWithDialect(dialect string) *Model {
	if dialect == "" {
		dialect = "mysql"
	}

	return &Model{
		Imports: make(map[string]struct{}),
		Dialect: strings.ToLower(dialect),
	}
}

// getGoType maps SQL types to Go types and returns the Go type and any required import.
//
// Parameters:
//   - sqlType: A string representing the SQL type.
//   - isUnsigned: A boolean indicating if the type is unsigned.
//
// Returns:
//   - A string representing the Go type.
//   - A string representing the import path required for the Go type, if any.
func (m *Model) getGoType(sqlType string, isUnsigned, isNullable bool) (string, string) {
	normalizedType := strings.ToLower(strings.TrimSpace(sqlType))
	makeNullable := func(goType, importPath string) (string, string) {
		if !isNullable {
			return goType, importPath
		}
		switch goType {
		case "[]byte":
			return goType, importPath
		default:
			return "*" + goType, importPath
		}
	}

	if m.Dialect == "postgres" {
		switch {
		case strings.HasPrefix(normalizedType, "smallserial"):
			return makeNullable("int16", "")
		case strings.HasPrefix(normalizedType, "serial"):
			return makeNullable("int32", "")
		case strings.HasPrefix(normalizedType, "bigserial"):
			return makeNullable("int64", "")
		case strings.HasPrefix(normalizedType, "character varying"), strings.HasPrefix(normalizedType, "varchar"):
			return makeNullable("string", "")
		case strings.HasPrefix(normalizedType, "uuid"):
			return makeNullable("string", "")
		case strings.HasPrefix(normalizedType, "jsonb"):
			return makeNullable("datatypes.JSON", "gorm.io/datatypes")
		case strings.HasPrefix(normalizedType, "bytea"):
			return "[]byte", ""
		case strings.HasPrefix(normalizedType, "timestamp with time zone"), strings.HasPrefix(normalizedType, "timestamp without time zone"), strings.HasPrefix(normalizedType, "timestamptz"):
			return makeNullable("time.Time", "time")
		case strings.HasPrefix(normalizedType, "double precision"):
			return makeNullable("float64", "")
		case strings.HasPrefix(normalizedType, "integer"):
			return makeNullable("int32", "")
		}
	}

	// Extract base type and size information
	typeRegex := regexp.MustCompile(`^(\w+)(?:\((\d+)(?:,(\d+))?\))?`)
	matches := typeRegex.FindStringSubmatch(normalizedType)

	if len(matches) == 0 {
		if isNullable {
			return "*any", ""
		}
		return "any", ""
	}

	baseType := strings.ToLower(matches[1])

	switch baseType {
	case "tinyint":
		if isUnsigned {
			return makeNullable("uint8", "")
		}
		return makeNullable("int8", "")
	case "smallint":
		if isUnsigned {
			return makeNullable("uint16", "")
		}
		return makeNullable("int16", "")
	case "mediumint", "int":
		if isUnsigned {
			return makeNullable("uint32", "")
		}
		return makeNullable("int32", "")
	case "bigint":
		if isUnsigned {
			return makeNullable("uint64", "")
		}
		return makeNullable("int64", "")
	case "float":
		return makeNullable("float32", "")
	case "double", "real":
		return makeNullable("float64", "")
	case "decimal", "numeric":
		return makeNullable("decimal.Decimal", "github.com/shopspring/decimal")
	case "bit":
		return makeNullable("uint64", "")
	case "bool", "boolean":
		return makeNullable("bool", "")
	case "char", "varchar", "text", "tinytext", "mediumtext", "longtext", "enum", "set":
		return makeNullable("string", "")
	case "binary", "varbinary", "blob", "tinyblob", "mediumblob", "longblob":
		return "[]byte", ""
	case "date", "time", "datetime", "timestamp", "year":
		return makeNullable("time.Time", "time")
	case "json":
		return makeNullable("datatypes.JSON", "gorm.io/datatypes")
	default:
		return makeNullable("any", "")
	}
}

// parseSQL parses the provided SQL schema string to extract table and field information.
// It updates the Model struct with the extracted information.
//
// The function processes the SQL schema line by line, identifying the table name and fields.
// It ignores certain fields like 'id', 'created_at', 'updated_at', and 'deleted_at'.
// For each relevant field, it determines the Go type and GORM tag, and adds these to the Model.
//
// Parameters:
//   - sql: A string containing the SQL schema.
//
// Returns:
//   - An error if there is an issue parsing the SQL schema, otherwise nil.
func (m *Model) parseSQL(sql string) error {
	m.TableFields = nil
	m.TableName = ""
	m.SchemaName = ""
	m.PrimaryKeyName = ""
	m.HasPrimaryKey = false
	m.IDType = ""
	m.UseGormModel = false

	// Split the SQL schema into lines for easier processing.
	lines := strings.Split(sql, "\n")

	// Process each line of the SQL schema.
	for _, line := range lines {
		// Trim whitespace and convert the line to lowercase for parsing.
		originalLine := strings.TrimSpace(line)
		line = strings.ToLower(originalLine)
		// Split the line into parts based on spaces.
		parts := strings.Fields(line)

		// Skip lines that do not have enough parts to be of interest.
		if len(parts) < 2 {
			continue
		}

		// Process different parts of the SQL schema based on the first word.
		switch parts[0] {
		case "create":
			// If the line starts with "create table", extract the table name.
			if len(parts) >= 3 && parts[1] == "table" {
				m.parseCreateTableLine(originalLine)
			}
		case "primary", "unique", "key", "index", "foreign", "engine", "default", "collate", "comment":
			// Ignore constraint, index, and table-level configuration lines.
			if primaryKeyName := m.extractPrimaryKeyName(originalLine); primaryKeyName != "" {
				m.markPrimaryKey(primaryKeyName)
			}
			continue
		case "constraint":
			if primaryKeyName := m.extractPrimaryKeyName(originalLine); primaryKeyName != "" {
				m.markPrimaryKey(primaryKeyName)
			}
			continue
		default:
			// Process lines that define fields.
			if strings.HasPrefix(line, ")") || strings.HasPrefix(line, "(") {
				continue
			}

			// Skip lines that contain table-level configurations.
			if strings.Contains(line, "engine") || strings.Contains(line, "charset") {
				continue
			}

			name := m.cleanIdentifier(parts[0])
			if len(parts) < 2 {
				continue
			}

			// Parse field attributes
			field := m.parseFieldDefinition(originalLine, name, parts)
			if field != nil {
				// Add the field to the Model's list of table fields.
				m.TableFields = append(m.TableFields, *field)
			}
		}
	}

	m.UseGormModel = m.canUseGormModel()
	if prioritizedField := m.prioritizedField(); prioritizedField != nil {
		m.IDType = m.generatedIDType(prioritizedField)
	} else if m.IDType == "" {
		m.IDType = "uint"
	}

	return nil
}

// parseFieldDefinition parses a single field definition line
func (m *Model) parseFieldDefinition(originalLine, fieldName string, parts []string) *Field {
	if len(parts) < 2 {
		return nil
	}

	field := &Field{
		Name:     strcase.ToCamel(fieldName),
		JsonName: fieldName,
	}

	// Parse field type and attributes
	typeStr := m.extractTypeDefinition(parts)
	field.SQLType = typeStr

	// Check for UNSIGNED
	field.IsUnsigned = strings.Contains(strings.ToUpper(originalLine), "UNSIGNED")

	// Check for NOT NULL
	field.IsNullable = !strings.Contains(strings.ToUpper(originalLine), "NOT NULL")
	if strings.Contains(strings.ToUpper(originalLine), "PRIMARY KEY") {
		field.IsNullable = false
	}

	// Check for AUTO_INCREMENT
	field.IsAutoIncrement = strings.Contains(strings.ToUpper(originalLine), "AUTO_INCREMENT")
	if m.Dialect == "postgres" {
		upperLine := strings.ToUpper(originalLine)
		if strings.Contains(upperLine, "GENERATED") && strings.Contains(upperLine, "AS IDENTITY") {
			field.IsAutoIncrement = true
			field.IsNullable = false
		}
		if strings.HasPrefix(strings.ToLower(typeStr), "serial") || strings.HasPrefix(strings.ToLower(typeStr), "bigserial") || strings.HasPrefix(strings.ToLower(typeStr), "smallserial") {
			field.IsAutoIncrement = true
			field.IsNullable = false
		}
	}

	if strings.Contains(strings.ToUpper(originalLine), "PRIMARY KEY") {
		field.IsPrimaryKey = true
		field.IsNullable = false
	}

	// Extract size and scale information
	field.Size, field.Scale = m.extractSizeAndScale(typeStr)

	// Extract default value
	field.DefaultValue = m.extractDefaultValue(originalLine)

	// Extract comment
	field.Comment = m.extractComment(originalLine)

	// Determine the Go type and any required import for the field.
	fieldType, importPath := m.getGoType(typeStr, field.IsUnsigned, field.IsNullable)
	field.Type = fieldType
	if field.IsPrimaryKey {
		m.HasPrimaryKey = true
		m.PrimaryKeyName = field.JsonName
		m.IDType = fieldType
	}

	if importPath != "" {
		m.Imports[importPath] = struct{}{}
	}

	// Generate GORM tag
	field.GormTag = m.generateGormTag(field, typeStr)

	return field
}

// extractSizeAndScale extracts size and scale from type definition like VARCHAR(255) or DECIMAL(10,2)
func (m *Model) extractSizeAndScale(typeStr string) (int, int) {
	typeRegex := regexp.MustCompile(`^(\w+)(?:\((\d+)(?:,(\d+))?\))?`)
	matches := typeRegex.FindStringSubmatch(typeStr)

	if len(matches) < 3 {
		return 0, 0
	}

	size := 0
	scale := 0

	if matches[2] != "" {
		if s, err := strconv.Atoi(matches[2]); err == nil {
			size = s
		}
	}

	if len(matches) > 3 && matches[3] != "" {
		if s, err := strconv.Atoi(matches[3]); err == nil {
			scale = s
		}
	}

	return size, scale
}

// extractDefaultValue extracts default value from field definition
func (m *Model) extractDefaultValue(line string) string {
	defaultRegex := regexp.MustCompile(`(?i)default\s+([^,\s]+)`)
	matches := defaultRegex.FindStringSubmatch(line)

	if len(matches) > 1 {
		defaultValue := strings.Trim(matches[1], "'\"")
		if strings.EqualFold(defaultValue, "NULL") {
			return ""
		}
		return defaultValue
	}

	return ""
}

// extractComment extracts comment from field definition
func (m *Model) extractComment(line string) string {
	// Support both COMMENT 'text' and COMMENT "text" formats
	commentRegex := regexp.MustCompile(`(?i)comment\s+['"]([^'"]*?)['"]`)
	matches := commentRegex.FindStringSubmatch(line)

	if len(matches) > 1 {
		return "// " + matches[1]
	}

	return ""
}

// generateGormTag generates appropriate GORM tag for the field
func (m *Model) generateGormTag(field *Field, typeStr string) string {
	var tags []string

	// Column name
	tags = append(tags, fmt.Sprintf("column:%s", field.JsonName))

	// Type specification for certain types
	baseType := strings.ToLower(strings.Split(typeStr, "(")[0])
	if m.Dialect == "mysql" {
		switch baseType {
		case "varchar", "char":
			if field.Size > 0 {
				tags = append(tags, fmt.Sprintf("type:varchar(%d)", field.Size))
			}
		case "decimal", "numeric":
			if field.Size > 0 && field.Scale > 0 {
				tags = append(tags, fmt.Sprintf("type:decimal(%d,%d)", field.Size, field.Scale))
			} else if field.Size > 0 {
				tags = append(tags, fmt.Sprintf("type:decimal(%d)", field.Size))
			}
		case "text", "tinytext", "mediumtext", "longtext":
			tags = append(tags, fmt.Sprintf("type:%s", baseType))
		}
	} else {
		switch baseType {
		case "numeric":
			if field.Size > 0 && field.Scale > 0 {
				tags = append(tags, fmt.Sprintf("type:numeric(%d,%d)", field.Size, field.Scale))
			}
		case "jsonb", "bytea", "uuid", "text", "date", "timestamp", "timestamptz":
			tags = append(tags, fmt.Sprintf("type:%s", baseType))
		case "varchar", "character varying":
			if field.Size > 0 {
				tags = append(tags, fmt.Sprintf("type:varchar(%d)", field.Size))
			}
		}
	}

	// NOT NULL constraint
	if !field.IsNullable {
		tags = append(tags, "not null")
	}

	// Default value
	if field.DefaultValue != "" {
		tags = append(tags, fmt.Sprintf("default:%s", field.DefaultValue))
	}

	if field.IsPrimaryKey {
		tags = append(tags, "primaryKey")
	}

	// Auto increment
	if field.IsAutoIncrement {
		tags = append(tags, "autoIncrement")
	}

	return strings.Join(tags, ";")
}

// generateCode generates Go code for the model based on the parsed SQL schema.
//
// The function sets the package name and struct name based on the table name,
// and uses a template to generate the Go code for the model.
//
// Returns:
//   - A string containing the generated Go code.
//   - An error if there is an issue generating the code.
func (m *Model) generateCode() (string, error) {
	// Determine the package and struct names based on the table name.
	m.PackageName, m.StructName = m.generateNames(m.TableName)
	if m.TableName == "" {
		return "", fmt.Errorf("table name is empty")
	}
	if m.PackageName == "" || m.StructName == "" {
		return "", fmt.Errorf("invalid table name: %s", m.TableName)
	}

	displayFields := m.templateFields()
	primaryKeyFieldName := ""
	primaryKeyName := m.PrimaryKeyName
	if primaryKeyName == "" {
		primaryKeyName = "id"
	}
	if primaryKeyField := m.primaryKeyField(); primaryKeyField != nil {
		primaryKeyFieldName = primaryKeyField.Name
	}
	prioritizedFieldName := ""
	hasPrioritizedField := false
	prioritizedField := m.prioritizedField()
	if prioritizedField != nil {
		hasPrioritizedField = true
		prioritizedFieldName = m.generatedFieldName(prioritizedField)
	}

	structNameFirstLetter := strings.ToLower(m.StructName[0:1])
	defaultOrderExpr := ""
	if prioritizedField != nil {
		defaultOrderExpr = fmt.Sprintf("%s desc", prioritizedField.JsonName)
	}

	// Parse the model template.
	tmpl := template.Must(template.New("model").Parse(modelTemplate))
	var result strings.Builder
	// Execute the template with the model data.
	err := tmpl.Execute(&result, map[string]interface{}{
		"Package":               m.PackageName,
		"StructName":            m.StructName,
		"StructNameFirstLetter": structNameFirstLetter,
		"StructNameLower":       strings.ToLower(m.StructName),
		"TableName":             m.qualifiedTableName(),
		"TableFields":           displayFields,
		"Imports":               m.Imports,
		"UseGormModel":          m.UseGormModel,
		"IDType":                m.IDType,
		"PrimaryKeyName":        primaryKeyName,
		"PrimaryKeyFieldName":   primaryKeyFieldName,
		"PrioritizedFieldName":  prioritizedFieldName,
		"HasPrioritizedField":   hasPrioritizedField,
		"DefaultOrderExpr":      defaultOrderExpr,
		"HasPrimaryKey":         m.HasPrimaryKey,
	})
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// generateNames generates package name and struct name from table name
// Supports various naming conventions:
// - user_profiles -> package: user, struct: Profile
// - auth_user_tokens -> package: auth, struct: UserToken
// - products -> package: products, struct: Product
// - user_role_permissions -> package: user, struct: RolePermission
func (m *Model) generateNames(tableName string) (packageName, structName string) {
	tableName = m.cleanIdentifier(tableName)
	parts := strings.Split(tableName, "_")

	switch len(parts) {
	case 1:
		// Single word table name: products -> package: products, struct: Product
		packageName = tableName
		structName = strcase.ToCamel(tableName)
		// Handle plural to singular conversion for common cases
		if strings.HasSuffix(structName, "s") && len(structName) > 1 {
			structName = structName[:len(structName)-1]
		}
	case 2:
		// Two parts: user_profiles -> package: user, struct: Profile
		packageName = parts[0]
		structName = strcase.ToCamel(parts[1])
		// Handle plural to singular conversion
		if strings.HasSuffix(structName, "s") && len(structName) > 1 {
			structName = structName[:len(structName)-1]
		}
	default:
		// Multiple parts: auth_user_tokens -> package: auth, struct: UserToken
		packageName = parts[0]
		// Combine remaining parts for struct name
		remainingParts := parts[1:]
		var structParts []string
		for _, part := range remainingParts {
			// Handle plural to singular for the last part
			if part == remainingParts[len(remainingParts)-1] && strings.HasSuffix(part, "s") && len(part) > 1 {
				part = part[:len(part)-1]
			}
			structParts = append(structParts, strcase.ToCamel(part))
		}
		structName = strings.Join(structParts, "")
	}

	// Ensure package name is valid Go identifier
	packageName = strings.ToLower(packageName)

	// Ensure struct name starts with uppercase
	if len(structName) > 0 {
		structName = strings.ToUpper(structName[0:1]) + structName[1:]
	}

	return packageName, structName
}

func (m *Model) parseCreateTableLine(line string) {
	re := regexp.MustCompile(`(?i)^create\s+table\s+(?:if\s+not\s+exists\s+)?(.+?)(?:\s*\(|$)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(line))
	if len(matches) < 2 {
		return
	}

	schemaName, tableName := m.parseQualifiedName(matches[1])
	m.SchemaName = schemaName
	m.TableName = tableName
}

func (m *Model) parseQualifiedName(identifier string) (string, string) {
	raw := strings.TrimSpace(identifier)
	raw = strings.TrimSuffix(raw, "(")
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, ",")

	parts := strings.Split(raw, ".")
	if len(parts) == 1 {
		return "", m.cleanIdentifier(parts[0])
	}

	schemaName := m.cleanIdentifier(parts[len(parts)-2])
	tableName := m.cleanIdentifier(parts[len(parts)-1])
	return schemaName, tableName
}

func (m *Model) cleanIdentifier(identifier string) string {
	cleaned := strings.TrimSpace(identifier)
	cleaned = strings.TrimSuffix(cleaned, ",")
	cleaned = strings.Trim(cleaned, "`\"")
	return cleaned
}

func (m *Model) qualifiedTableName() string {
	if m.SchemaName == "" {
		return m.TableName
	}
	return m.SchemaName + "." + m.TableName
}

func (m *Model) extractTypeDefinition(parts []string) string {
	if len(parts) < 2 {
		return ""
	}

	stopWords := map[string]struct{}{
		"not":        {},
		"null":       {},
		"default":    {},
		"comment":    {},
		"constraint": {},
		"primary":    {},
		"unique":     {},
		"references": {},
		"check":      {},
		"generated":  {},
		"collate":    {},
	}

	typeParts := make([]string, 0, 3)
	for _, part := range parts[1:] {
		normalized := strings.ToLower(strings.Trim(part, ","))
		if _, isStopWord := stopWords[normalized]; isStopWord {
			break
		}
		typeParts = append(typeParts, normalized)
		if normalized == "zone" || normalized == "precision" {
			break
		}
	}

	return strings.Join(typeParts, " ")
}

func (m *Model) extractPrimaryKeyName(line string) string {
	re := regexp.MustCompile(`(?i)primary\s+key\s*\(([^)]+)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return ""
	}

	columns := strings.Split(matches[1], ",")
	if len(columns) != 1 {
		return ""
	}

	return m.cleanIdentifier(columns[0])
}

func (m *Model) markPrimaryKey(columnName string) {
	for i := range m.TableFields {
		if m.TableFields[i].JsonName != columnName {
			continue
		}
		m.TableFields[i].IsPrimaryKey = true
		m.TableFields[i].IsNullable = false
		fieldType, importPath := m.getGoType(m.TableFields[i].SQLType, m.TableFields[i].IsUnsigned, m.TableFields[i].IsNullable)
		m.TableFields[i].Type = fieldType
		if importPath != "" {
			m.Imports[importPath] = struct{}{}
		}
		m.HasPrimaryKey = true
		m.PrimaryKeyName = columnName
		m.IDType = m.TableFields[i].Type
		m.TableFields[i].GormTag = m.generateGormTag(&m.TableFields[i], m.TableFields[i].SQLType)
		return
	}
}

func (m *Model) canUseGormModel() bool {
	requiredFields := map[string]bool{
		"id":         false,
		"created_at": false,
		"updated_at": false,
		"deleted_at": false,
	}

	for _, field := range m.TableFields {
		if _, ok := requiredFields[field.JsonName]; ok {
			requiredFields[field.JsonName] = true
		}
	}

	for _, exists := range requiredFields {
		if !exists {
			return false
		}
	}

	return true
}

func (m *Model) templateFields() []Field {
	if !m.UseGormModel {
		return m.TableFields
	}

	fields := make([]Field, 0, len(m.TableFields))
	for _, field := range m.TableFields {
		switch field.JsonName {
		case "id", "created_at", "updated_at", "deleted_at":
			continue
		default:
			fields = append(fields, field)
		}
	}

	return fields
}

func (m *Model) primaryKeyField() *Field {
	for i := range m.TableFields {
		if m.TableFields[i].IsPrimaryKey {
			return &m.TableFields[i]
		}
	}
	return nil
}

func (m *Model) prioritizedField() *Field {
	if primaryKeyField := m.primaryKeyField(); primaryKeyField != nil {
		return primaryKeyField
	}

	for i := range m.TableFields {
		if strings.EqualFold(m.TableFields[i].JsonName, "id") {
			return &m.TableFields[i]
		}
	}

	return nil
}

func (m *Model) generatedIDType(field *Field) string {
	if field == nil {
		return "uint"
	}
	if m.UseGormModel && strings.EqualFold(field.JsonName, "id") {
		return "uint"
	}
	return field.Type
}

func (m *Model) generatedFieldName(field *Field) string {
	if field == nil {
		return ""
	}
	if m.UseGormModel && strings.EqualFold(field.JsonName, "id") {
		return "ID"
	}
	return field.Name
}

// readSQLFile reads the content of the specified SQL file.
//
// Parameters:
//   - filePath: A string representing the path to the SQL file.
//
// Returns:
//   - A string containing the content of the SQL file.
//   - An error if there is an issue reading the file.
func (m *Model) readSQLFile(filePath string) (string, error) {
	// Check if the file exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read the file content.
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return string(content), nil
}

// WriteModelFile writes the generated model code to a file.
//
// This function ensures that the output directory exists, and writes
// the generated code to the specified output path.
//
// Parameters:
//   - force: A boolean indicating whether to overwrite existing files.
//   - outputPath: A string representing the output path for the file.
//   - content: A string containing the generated code to be written.
//
// Returns:
//   - An error if there is an issue writing the file.
func (m *Model) WriteModelFile(force bool, outputPath, content string) error {
	// Use the default output path if none is provided.
	if outputPath == "" {
		workPath, err := os.Getwd()
		if err != nil {
			return err
		}

		outputPath = filepath.Join(workPath, defaultModelOutPath)
	}

	// Determine the output file path based on the table name.
	parts := strings.Split(m.TableName, "_")
	if len(parts) > 1 {
		outputPath = filepath.Join(outputPath, strings.ReplaceAll(m.TableName, "_", string(os.PathSeparator))+".go")
	} else {
		outputPath = filepath.Join(outputPath, m.TableName, m.TableName+".go")
	}

	log.Printf("Starting to write Model file: %s\n", outputPath)

	// Check if the file already exists and handle overwriting based on the force flag.
	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			log.Printf("%s already exists, not overwriting\n", outputPath)
			return nil
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	// Ensure the output directory exists.
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Write the generated code to the file.
	return os.WriteFile(outputPath, []byte(content), 0644)
}

// Generate orchestrates the model generation process.
//
// It reads the SQL schema file, parses the schema, generates the Go code,
// formats the code, and writes the formatted code to the output file.
//
// Parameters:
//   - force: A boolean indicating whether to overwrite existing files.
//   - sqlPath: A string representing the path to the SQL schema file.
//   - outputPath: A string representing the output path for the generated code.
//
// Returns:
//   - An error if there is an issue during the generation process.
func (m *Model) Generate(force bool, sqlPath, outputPath string) error {
	log.Printf("-------codegen-------\n")
	log.Printf("Starting to read %s\n", sqlPath)

	// Read the SQL schema file.
	sql, err := m.readSQLFile(sqlPath)
	if err != nil {
		return fmt.Errorf("error reading SQL file: %w", err)
	}

	log.Printf("Successfully read %s\n", sqlPath)
	log.Printf("Starting to parse %s \n", sqlPath)

	// Parse the SQL schema.
	err = m.parseSQL(sql)
	if err != nil {
		return fmt.Errorf("error parsing SQL: %w", err)
	}

	log.Printf("Successfully parsed %s\n", sqlPath)
	log.Printf("Starting to generate Model...\n")

	// Generate the Go code for the model.
	code, err := m.generateCode()
	if err != nil {
		return fmt.Errorf("error generating code: %w", err)
	}

	log.Printf("Successfully generated %s Model\n", m.StructName)
	log.Printf("Starting to format %s Model...\n", m.StructName)

	// Format the generated Go code.
	formattedContent, err := m.formatGoCode(code)
	if err != nil {
		return err
	}

	log.Printf("Successfully formatted %s Model\n", m.StructName)

	// Write the formatted code to the output file.
	if err = m.WriteModelFile(force, outputPath, formattedContent); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	log.Printf("%s Model has been successfully generated\n", m.StructName)

	return nil
}

// formatGoCode formats the generated Go code using 'gofmt'.
//
// Parameters:
//   - code: A string containing the generated Go code.
//
// Returns:
//   - A string containing the formatted Go code.
//   - An error if there is an issue formatting the code.
func (m *Model) formatGoCode(code string) (string, error) {
	// Execute the gofmt command to format the code.
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(code)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}

const modelTemplate = `// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package {{.Package}}

import (
	"context"
	"errors"
	"fmt"
	
	{{- range $import, $value := .Imports}}
	"{{$import}}"
	{{- end}}

	"gorm.io/gorm"
)

type {{.StructName}} struct {
	{{- if .UseGormModel}}
	gorm.Model
	{{- end}}
	{{"\n"}}
	{{- range .TableFields}}
	{{.Name}} {{.Type}} ` + "`gorm:\"{{.GormTag}}\" json:\"{{.JsonName}}\"`" + ` {{.Comment}}
	{{- end}}

	// Query conditions for chaining methods
	queryCondition interface{}   ` + "`gorm:\"-\" json:\"-\"`" + `
	queryArgs      []interface{} ` + "`gorm:\"-\" json:\"-\"`" + `
}

// TableName specifies the table name for the {{.StructName}} model.
func ({{.StructNameFirstLetter}} *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
}

// Where sets query conditions for chaining with other methods.
//
// Parameters:
// 	- query: SQL query string or condition.
// 	- args: variadic arguments for the SQL query.
//
// Returns:
// 	- *{{.StructName}}: pointer to the {{.StructName}} instance for method chaining.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Where(query interface{}, args ...interface{}) *{{.StructName}} {
	new{{.StructName}} := *{{.StructNameFirstLetter}}
	new{{.StructName}}.queryCondition = query
	new{{.StructName}}.queryArgs = args
	return &new{{.StructName}}
}

// First retrieves the first {{.StructNameLower}} matching the criteria from the database.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- *{{.StructName}}: pointer to the retrieved {{.StructNameLower}}, or nil if not found.
// 	- error: error if the query fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) First(ctx context.Context, db *gorm.DB) (*{{.StructName}}, error) {
	var {{.StructNameLower}} {{.StructName}}

	query := db.WithContext(ctx)
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}

	// Perform the database query with context.
	if err := query.First(&{{.StructNameLower}}).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &{{.StructNameLower}}, nil
}

// Last retrieves the last {{.StructNameLower}} matching the criteria from the database, ordered by primary key or id in descending order when available.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- *{{.StructName}}: pointer to the retrieved {{.StructNameLower}}, or nil if not found.
// 	- error: error if the query fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Last(ctx context.Context, db *gorm.DB) (*{{.StructName}}, error) {
	var {{.StructNameLower}} {{.StructName}}

	query := db.WithContext(ctx)
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}

	// Perform the database query with context.
	{{- if .HasPrioritizedField}}
	if err := query.Order("{{.DefaultOrderExpr}}").First(&{{.StructNameLower}}).Error; err != nil {
	{{- else}}
	return nil, fmt.Errorf("find last failed: no primary key or id field available for ordering")
	{{- end}}
	{{- if .HasPrioritizedField}}
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}
	{{- end}}

	return &{{.StructNameLower}}, nil
}

// Create inserts a new {{.StructNameLower}} into the database and returns the ID of the created {{.StructName}}.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- uint: ID of the created {{.StructNameLower}}.
// 	- error: error if the insert operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Create(ctx context.Context, db *gorm.DB) ({{.IDType}}, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create({{.StructNameFirstLetter}}).Error; err != nil {
		return *new({{.IDType}}), fmt.Errorf("create failed: %w", err)
	}

	{{- if .HasPrioritizedField}}
	return {{.StructNameFirstLetter}}.{{.PrioritizedFieldName}}, nil
	{{- else}}
	return *new({{.IDType}}), nil
	{{- end}}
}

// Delete removes the {{.StructNameLower}} from the database.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- error: error if the delete operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Delete(ctx context.Context, db *gorm.DB) error {
	query := db.WithContext(ctx)
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}

	// Perform the database delete operation with context.
	return query.Delete(&{{.StructName}}{}).Error
}

// Updates applies the specified updates to the {{.StructNameLower}} in the database.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
// 	- updates: map[string]interface{} containing the updates to apply.
//
// Returns:
// 	- error: error if the update operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	// Build query with context and explicitly set the model
	query := db.WithContext(ctx).Model(&{{.StructName}}{})
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		{{- if .HasPrimaryKey}}
		if {{.StructNameFirstLetter}}.{{.PrimaryKeyFieldName}} != *new({{.IDType}}) {
			// Use primary key for updates if available.
			query = query.Where("{{.PrimaryKeyName}} = ?", {{.StructNameFirstLetter}}.{{.PrimaryKeyFieldName}})
		} else {
			// Use struct fields as condition
			query = query.Where({{.StructNameFirstLetter}})
		}
		{{- else}}
		// Use struct fields as condition
		query = query.Where({{.StructNameFirstLetter}})
		{{- end}}
	}

	// Perform the database update operation with context.
	return query.Updates(updates).Error
}

// List retrieves all {{.StructNameLower}}s matching the criteria from the database.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- []{{.StructName}}: slice of retrieved {{.StructNameLower}}s.
// 	- error: error if the query fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) List(ctx context.Context, db *gorm.DB) ([]{{.StructName}}, error) {
	var {{.StructNameLower}}s []{{.StructName}}

	query := db.WithContext(ctx)
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}
	
	// Perform the database query with context.
	if err := query.Find(&{{.StructNameLower}}s).Error; err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}

	return {{.StructNameLower}}s, nil
}

// ListByArgs retrieves {{.StructNameLower}}s matching the specified query and arguments from the database, ordered by ID in descending order.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
// 	- query: SQL query string.
// 	- args: variadic arguments for the SQL query.
//
// Returns:
// 	- []{{.StructName}}: slice of retrieved {{.StructNameLower}}s.
// 	- error: error if the query fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) ListByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) ([]{{.StructName}}, error) {
	var {{.StructNameLower}}s []{{.StructName}}

	queryBuilder := db.WithContext(ctx).Model(&{{.StructName}}{}).Where(query, args...)
	{{- if .HasPrimaryKey}}
	queryBuilder = queryBuilder.Order("{{.DefaultOrderExpr}}")
	{{- end}}

	// Perform the database query with context.
	if err := queryBuilder.Find(&{{.StructNameLower}}s).Error; err != nil {
		return nil, fmt.Errorf("list by args failed: %w", err)
	}

	return {{.StructNameLower}}s, nil
}

// CountByArgs counts the number of {{.StructNameLower}}s matching the specified query and arguments in the database.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
// 	- query: SQL query string.
// 	- args: variadic arguments for the SQL query.
//
// Returns:
// 	- int64: count of matching {{.StructNameLower}}s.
// 	- error: error if the count operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) CountByArgs(ctx context.Context, db *gorm.DB, query interface{}, args ...interface{}) (int64, error) {
	var count int64

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&{{.StructName}}{}).Where(query, args...).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by args failed: %w", err)
	}

	return count, nil
}

// Count counts the number of {{.StructNameLower}}s matching the criteria in the database.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- int64: count of matching {{.StructNameLower}}s.
// 	- error: error if the count operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Count(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64

	query := db.WithContext(ctx).Model(&{{.StructName}}{})
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}

	// Perform the database count operation with context.
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

// BatchInsert inserts multiple {{.StructNameLower}}s into the database in a single batch operation.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
// 	- {{.StructNameLower}}s: slice of {{.StructName}} instances to be inserted.
//
// Returns:
// 	- error: error if the batch insert operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) BatchInsert(ctx context.Context, db *gorm.DB, {{.StructNameLower}}s []{{.StructName}}) error {
	// Perform the database batch insert operation with context.
	return db.WithContext(ctx).Create(&{{.StructNameLower}}s).Error
}

// FindWithPagination retrieves {{.StructNameLower}}s matching the criteria from the database with pagination support.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
// 	- page: page number for pagination (1-based).
// 	- size: number of {{.StructNameLower}}s per page.
//
// Returns:
// 	- []{{.StructName}}: slice of retrieved {{.StructNameLower}}s.
// 	- error: error if the query fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) FindWithPagination(ctx context.Context, db *gorm.DB, page, size int) ([]{{.StructName}}, error) {
	var {{.StructNameLower}}s []{{.StructName}}

	query := db.WithContext(ctx)
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}

	// Perform the database query with context, applying offset and limit for pagination.
	if err := query.Offset((page - 1) * size).Limit(size).Find(&{{.StructNameLower}}s).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return {{.StructNameLower}}s, nil
}

// FindWithSort retrieves {{.StructNameLower}}s matching the criteria from the database with sorting support.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
// 	- sort: sorting criteria (e.g., "id desc").
//
// Returns:
// 	- []{{.StructName}}: slice of retrieved {{.StructNameLower}}s.
// 	- error: error if the query fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) FindWithSort(ctx context.Context, db *gorm.DB, sort string) ([]{{.StructName}}, error) {
	var {{.StructNameLower}}s []{{.StructName}}

	query := db.WithContext(ctx)
	
	// Apply Where conditions if set
	if {{.StructNameFirstLetter}}.queryCondition != nil {
		query = query.Where({{.StructNameFirstLetter}}.queryCondition, {{.StructNameFirstLetter}}.queryArgs...)
	} else {
		query = query.Where({{.StructNameFirstLetter}})
	}

	// Perform the database query with context, applying the specified sort order.
	if err := query.Order(sort).Find(&{{.StructNameLower}}s).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return {{.StructNameLower}}s, nil
}
`
