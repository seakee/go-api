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
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

const defaultModelOutPath = "app/model"

// Field represents a field in a database table.
type Field struct {
	Name     string // The name of the field in Go struct
	Type     string // The Go type of the field
	JsonName string // The JSON name of the field
	GormTag  string // The GORM tag for the field
	Comment  string // Comment associated with the field
}

// Model represents the structure of a database table.
type Model struct {
	PackageName string              // The package name for the generated Go file
	Imports     map[string]struct{} // Imports required by the Go file
	StructName  string              // The name of the Go struct
	TableName   string              // The name of the database table
	TableFields []Field             // The fields of the table
}

// NewModel creates a new instance of Model.
func NewModel() *Model {
	return &Model{
		Imports: make(map[string]struct{}),
	}
}

// getGoType maps SQL types to Go types and returns the Go type and any required import.
//
// Parameters:
//   - sqlType: A string representing the SQL type.
//
// Returns:
//   - A string representing the Go type.
//   - A string representing the import path required for the Go type, if any.
func (m *Model) getGoType(sqlType string) (string, string) {
	switch {
	case strings.HasPrefix(sqlType, "int"):
		return "int", ""
	case strings.HasPrefix(sqlType, "tinyint"):
		return "int8", ""
	case strings.HasPrefix(sqlType, "smallint"):
		return "int16", ""
	case strings.HasPrefix(sqlType, "mediumint"):
		return "int32", ""
	case strings.HasPrefix(sqlType, "bigint"):
		return "int64", ""
	case strings.HasPrefix(sqlType, "float"):
		return "float32", ""
	case strings.HasPrefix(sqlType, "double"), strings.HasPrefix(sqlType, "real"):
		return "float64", ""
	case strings.HasPrefix(sqlType, "decimal"), strings.HasPrefix(sqlType, "numeric"):
		return "decimal.Decimal", "github.com/shopspring/decimal"
	case strings.HasPrefix(sqlType, "bit"):
		return "uint64", ""
	case strings.HasPrefix(sqlType, "bool"), strings.HasPrefix(sqlType, "boolean"):
		return "bool", ""
	case strings.HasPrefix(sqlType, "char"), strings.HasPrefix(sqlType, "varchar"), strings.HasPrefix(sqlType, "text"), strings.HasPrefix(sqlType, "enum"), strings.HasPrefix(sqlType, "set"):
		return "string", ""
	case strings.HasPrefix(sqlType, "binary"), strings.HasPrefix(sqlType, "varbinary"), strings.HasPrefix(sqlType, "blob"):
		return "[]byte", ""
	case strings.HasPrefix(sqlType, "date"), strings.HasPrefix(sqlType, "time"), strings.HasPrefix(sqlType, "datetime"), strings.HasPrefix(sqlType, "timestamp"), strings.HasPrefix(sqlType, "year"):
		return "time.Time", "time"
	case strings.HasPrefix(sqlType, "json"):
		return "datatypes.JSON", "gorm.io/datatypes"
	default:
		return "any", ""
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
	// Split the SQL schema into lines for easier processing.
	lines := strings.Split(sql, "\n")

	// Process each line of the SQL schema.
	for _, line := range lines {
		// Trim whitespace and convert the line to lowercase.
		line = strings.TrimSpace(strings.ToLower(line))
		// Split the line into parts based on spaces.
		parts := strings.Fields(line)

		// Skip lines that do not have enough parts to be of interest.
		if len(parts) < 3 {
			continue
		}

		// Process different parts of the SQL schema based on the first word.
		switch parts[0] {
		case "create":
			// If the line starts with "create table", extract the table name.
			if parts[1] == "table" {
				m.TableName = strings.Trim(parts[2], "`")
			}
		case "primary", "unique", "key":
			// Ignore lines starting with "primary", "unique", or "key".
			goType, _ := m.getGoType(parts[1])
			if goType == "any" {
				return nil
			}
		default:
			// Process lines that define fields.
			name := strings.Trim(parts[0], "`")
			// Skip certain predefined field names.
			if name == "id" || name == "created_at" || name == "updated_at" || name == "deleted_at" {
				continue
			}

			// Extract the comment if it exists.
			comment := ""
			if strings.Trim(parts[len(parts)-2], "") == "comment" {
				comment = strings.Trim(parts[len(parts)-1], "'`,")
				comment = "// " + comment
			}

			// Determine the Go type and any required import for the field.
			fieldType, importPath := m.getGoType(parts[1])
			if importPath != "" {
				m.Imports[importPath] = struct{}{}
			}

			// Create a Field struct with the extracted information.
			field := Field{
				Name:     strcase.ToCamel(name),          // Convert the field name to CamelCase.
				Type:     fieldType,                      // Set the field type.
				JsonName: name,                           // Set the JSON name for the field.
				GormTag:  fmt.Sprintf("column:%s", name), // Set the GORM tag.
				Comment:  comment,                        // Set the associated comment.
			}

			// Add the field to the Model's list of table fields.
			m.TableFields = append(m.TableFields, field)
		}
	}

	return nil
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
	parts := strings.Split(m.TableName, "_")
	if len(parts) > 1 {
		m.PackageName = parts[len(parts)-2]
		m.StructName = strcase.ToCamel(parts[len(parts)-1])
	} else {
		m.PackageName = m.TableName
		m.StructName = strcase.ToCamel(m.TableName)
	}

	// Parse the model template.
	tmpl := template.Must(template.New("model").Parse(modelTemplate))
	var result strings.Builder
	// Execute the template with the model data.
	err := tmpl.Execute(&result, map[string]interface{}{
		"Package":               m.PackageName,
		"StructName":            m.StructName,
		"StructNameFirstLetter": strings.ToLower(m.StructName[0:1]),
		"StructNameLower":       strings.ToLower(m.StructName),
		"TableName":             m.TableName,
		"TableFields":           m.TableFields,
		"Imports":               m.Imports,
	})
	if err != nil {
		return "", err
	}

	return result.String(), nil
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
	gorm.Model
	{{"\n"}}
	{{- range .TableFields}}
	{{.Name}} {{.Type}} ` + "`gorm:\"{{.GormTag}}\" json:\"{{.JsonName}}\"`" + ` {{.Comment}}
	{{- end}}
}

// TableName specifies the table name for the {{.StructName}} model.
func ({{.StructNameFirstLetter}} *{{.StructName}}) TableName() string {
	return "{{.TableName}}"
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

    // Perform the database query with context.
	if err := db.WithContext(ctx).Where({{.StructNameFirstLetter}}).First(&{{.StructNameLower}}).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &{{.StructNameLower}}, nil
}

// Last retrieves the last {{.StructNameLower}} matching the criteria from the database, ordered by ID in descending order.
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

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where({{.StructNameFirstLetter}}).Order("id desc").First(&{{.StructNameLower}}).Error; err != nil {
		// If no record is found, return nil without an error.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// Return the error if the query fails.
		return nil, fmt.Errorf("find last failed: %w", err)
	}

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
func ({{.StructNameFirstLetter}} *{{.StructName}}) Create(ctx context.Context, db *gorm.DB) (uint, error) {
	// Perform the database insert operation with context.
	if err := db.WithContext(ctx).Create({{.StructNameFirstLetter}}).Error; err != nil {
		return 0, fmt.Errorf("create failed: %w", err)
	}

	return {{.StructNameFirstLetter}}.ID, nil
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
	// Perform the database delete operation with context.
	return db.WithContext(ctx).Where({{.StructNameFirstLetter}}).Delete({{.StructNameFirstLetter}}).Error
}

// Remove removes the {{.StructNameLower}} from the database permanently.
//
// Parameters:
// 	- ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
// 	- db: *gorm.DB database connection.
//
// Returns:
// 	- error: error if the remove operation fails, otherwise nil.
func ({{.StructNameFirstLetter}} *{{.StructName}}) Remove(ctx context.Context, db *gorm.DB) error {
	// Perform the database remove operation with context.
	return db.WithContext(ctx).Unscoped().Delete({{.StructNameFirstLetter}}).Error
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
	// Perform the database update operation with context.
	return db.WithContext(ctx).Model({{.StructNameFirstLetter}}).Updates(updates).Error
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
	
	// Perform the database query with context.
	if err := db.WithContext(ctx).Where({{.StructNameFirstLetter}}).Find(&{{.StructNameLower}}s).Error; err != nil {
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

	// Perform the database query with context.
	if err := db.WithContext(ctx).Where(query, args...).Order("id desc").Find(&{{.StructNameLower}}s).Error; err != nil {
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

	// Perform the database count operation with context.
	if err := db.WithContext(ctx).Model(&{{.StructName}}{}).Where({{.StructNameFirstLetter}}).Count(&count).Error; err != nil {
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

	// Perform the database query with context, applying offset and limit for pagination.
	if err := db.WithContext(ctx).Where({{.StructNameFirstLetter}}).Offset((page - 1) * size).Limit(size).Find(&{{.StructNameLower}}s).Error; err != nil {
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

	// Perform the database query with context, applying the specified sort order.
	if err := db.WithContext(ctx).Where({{.StructNameFirstLetter}}).Order(sort).Find(&{{.StructNameLower}}s).Error; err != nil {
		// Return the error if the query fails.
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return {{.StructNameLower}}s, nil
}
`
