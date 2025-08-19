// Copyright 2025 Seakee.  All rights reserved.
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
)

const defaultRepositoryOutPath = "app/repository"

// Repository represents the structure for generating repository code.
type Repository struct {
	Model           *Model              // Reference to the associated model
	PackageName     string              // The package name for the generated Go file
	Imports         map[string]struct{} // Imports required by the Go file
	StructName      string              // The name of the repository struct
	ModelStructName string              // The name of the model struct
	TableName       string              // The name of the database table
}

// NewRepository creates a new instance of Repository.
func NewRepository(model *Model) *Repository {
	return &Repository{
		Model:           model,
		PackageName:     model.PackageName,
		StructName:      model.StructName + "Repository",
		ModelStructName: model.StructName,
		TableName:       model.TableName,
		Imports:         make(map[string]struct{}),
	}
}

// generateCode generates Go code for the repository based on the associated model.
//
// The function sets the package name and struct name based on the model,
// and uses a template to generate the Go code for the repository.
//
// Returns:
//   - A string containing the generated Go code.
//   - An error if there is an issue generating the code.
func (r *Repository) generateCode() (string, error) {
	// Set the package name and struct name based on the model.
	r.PackageName = r.Model.PackageName
	r.StructName = r.Model.StructName
	r.ModelStructName = r.Model.StructName
	r.TableName = r.Model.TableName

	// Add required imports
	r.Imports["errors"] = struct{}{}

	// Always add model import and use package prefix
	modelImportPath := fmt.Sprintf("github.com/seakee/go-api/app/model/%s", r.Model.PackageName)
	r.Imports[modelImportPath] = struct{}{}

	// Generate field checks for Update method
	fieldChecks := r.generateFieldChecks()

	// Generate query conditions for List method
	queryConditions := r.generateQueryConditions()

	// Create template data
	data := map[string]interface{}{
		"PackageName":           r.PackageName,
		"StructName":            r.StructName,
		"LowerStructName":       r.Model.PackageName, // Use model package name for type prefix
		"StructNameFirstLetter": strings.ToLower(r.StructName[:1]),
		"ModelStructName":       r.ModelStructName,
		"ModelStructNameLower":  strings.ToLower(r.ModelStructName),
		"TableName":             r.TableName,
		"ModelImport":           modelImportPath,
		"Imports":               r.Imports,
		"FieldChecks":           fieldChecks,
		"QueryConditions":       queryConditions,
	}

	// Parse and execute the template
	tmpl, err := template.New("repository").Parse(repositoryTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse repository template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute repository template: %w", err)
	}

	return buf.String(), nil
}

// generateFieldChecks generates field check code for the Update method
func (r *Repository) generateFieldChecks() string {
	var checks []string
	modelStructNameLower := strings.ToLower(r.ModelStructName)

	for _, field := range r.Model.TableFields {
		// Skip ID field as it's used for identification
		if strings.ToLower(field.Name) == "id" {
			continue
		}

		// Skip created_at and updated_at fields as they are managed by GORM
		fieldNameLower := strings.ToLower(field.Name)
		if fieldNameLower == "created_at" || fieldNameLower == "updated_at" || fieldNameLower == "createdat" || fieldNameLower == "updatedat" {
			continue
		}

		// Generate appropriate check based on field type
		var checkCondition string
		switch field.Type {
		case "string":
			checkCondition = fmt.Sprintf(`if %s.%s != "" {`, modelStructNameLower, field.Name)
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			checkCondition = fmt.Sprintf(`if %s.%s != 0 {`, modelStructNameLower, field.Name)
		case "float32", "float64":
			checkCondition = fmt.Sprintf(`if %s.%s != 0.0 {`, modelStructNameLower, field.Name)
		case "bool":
			// For bool, we always update since false is a valid value
			fieldCheck := fmt.Sprintf(`// Always update bool field%s	data["%s"] = %s.%s`, "\n\t", field.JsonName, modelStructNameLower, field.Name)
			checks = append(checks, fieldCheck)
			continue
		case "*string":
			checkCondition = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64":
			checkCondition = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "*float32", "*float64":
			checkCondition = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "*bool":
			checkCondition = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "time.Time":
			checkCondition = fmt.Sprintf(`if !%s.%s.IsZero() {`, modelStructNameLower, field.Name)
		case "*time.Time":
			checkCondition = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		default:
			// For other types, use nil check
			checkCondition = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		}

		// Add the data assignment
		if !strings.Contains(checkCondition, "data[") {
			fieldCheck := fmt.Sprintf(`%s%s		data["%s"] = %s.%s%s	}`, checkCondition, "\n", field.JsonName, modelStructNameLower, field.Name, "\n")
			checks = append(checks, fieldCheck)
		}
	}

	return strings.Join(checks, "\n\n\t")
}

// generateQueryConditions generates query condition code for the List method
func (r *Repository) generateQueryConditions() string {
	var conditions []string
	modelStructNameLower := strings.ToLower(r.ModelStructName)

	for _, field := range r.Model.TableFields {
		// Generate appropriate condition based on field type
		var conditionCheck string
		switch field.Type {
		case "string":
			conditionCheck = fmt.Sprintf(`if %s.%s != "" {`, modelStructNameLower, field.Name)
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			conditionCheck = fmt.Sprintf(`if %s.%s != 0 {`, modelStructNameLower, field.Name)
		case "float32", "float64":
			conditionCheck = fmt.Sprintf(`if %s.%s != 0.0 {`, modelStructNameLower, field.Name)
		case "bool":
			// For bool, we always add condition since false is a valid value
			condition := fmt.Sprintf(`// Always add bool field condition%s	query = query.Where("%s = ?", %s.%s)`, "\n\t", field.JsonName, modelStructNameLower, field.Name)
			conditions = append(conditions, condition)
			continue
		case "*string":
			conditionCheck = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "*int", "*int8", "*int16", "*int32", "*int64", "*uint", "*uint8", "*uint16", "*uint32", "*uint64":
			conditionCheck = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "*float32", "*float64":
			conditionCheck = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "*bool":
			conditionCheck = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		case "time.Time":
			conditionCheck = fmt.Sprintf(`if !%s.%s.IsZero() {`, modelStructNameLower, field.Name)
		case "*time.Time":
			conditionCheck = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		default:
			// For other types, use nil check
			conditionCheck = fmt.Sprintf(`if %s.%s != nil {`, modelStructNameLower, field.Name)
		}

		// Add the where condition
		if !strings.Contains(conditionCheck, "query =") {
			condition := fmt.Sprintf(`%s%s		query = query.Where("%s = ?", %s.%s)%s	}`, conditionCheck, "\n", field.JsonName, modelStructNameLower, field.Name, "\n")
			conditions = append(conditions, condition)
		}
	}

	return strings.Join(conditions, "\n\n\t")
}

// WriteRepositoryFile writes the generated repository code to a file.
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
func (r *Repository) WriteRepositoryFile(force bool, outputPath, content string) error {
	// Use the default output path if none is provided.
	if outputPath == "" {
		workPath, err := os.Getwd()
		if err != nil {
			return err
		}

		outputPath = filepath.Join(workPath, defaultRepositoryOutPath)
	}

	// Determine the output file path based on the table name.
	// Use the same logic as Model generator for consistency
	parts := strings.Split(r.TableName, "_")
	if len(parts) > 1 {
		// For tables like "auth_app", create "auth/app.go"
		dirPath := strings.Join(parts[:len(parts)-1], string(os.PathSeparator))
		fileName := parts[len(parts)-1] + ".go"
		outputPath = filepath.Join(outputPath, dirPath, fileName)
	} else {
		// For single word tables like "user", create "user/user.go"
		outputPath = filepath.Join(outputPath, r.TableName, r.TableName+".go")
	}

	log.Printf("Starting to write Repository file: %s\n", outputPath)

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

// Generate orchestrates the repository generation process.
//
// It generates the Go code for the repository, formats the code,
// and writes the formatted code to the output file.
//
// Parameters:
//   - force: A boolean indicating whether to overwrite existing files.
//   - outputPath: A string representing the output path for the generated code.
//
// Returns:
//   - An error if there is an issue during the generation process.
func (r *Repository) Generate(force bool, outputPath string) error {
	log.Printf("Starting to generate Repository for %s...\n", r.Model.StructName)

	// Generate the Go code for the repository.
	code, err := r.generateCode()
	if err != nil {
		return fmt.Errorf("error generating repository code: %w", err)
	}

	log.Printf("Successfully generated %s Repository\n", r.StructName)
	log.Printf("Starting to format %s Repository...\n", r.StructName)

	// Format the generated Go code.
	formattedContent, err := r.formatGoCode(code)
	if err != nil {
		return err
	}

	log.Printf("Successfully formatted %s Repository\n", r.StructName)

	// Write the formatted code to the output file.
	if err = r.WriteRepositoryFile(force, outputPath, formattedContent); err != nil {
		return fmt.Errorf("error writing repository file: %w", err)
	}

	log.Printf("%s Repository has been successfully generated\n", r.StructName)

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
func (r *Repository) formatGoCode(code string) (string, error) {
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

const repositoryTemplate = `// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package {{.PackageName}}

import (
	"context"
{{if .ModelImport}}	"{{.ModelImport}}"
{{end}}	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// {{.StructName}}Repo defines the interface for {{.ModelStructNameLower}}-related database operations.
type {{.StructName}}Repo interface {
	// Get{{.ModelStructName}} retrieves a {{.ModelStructNameLower}} by its properties.
	Get{{.ModelStructName}}(ctx context.Context, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) (*{{.LowerStructName}}.{{.ModelStructName}}, error)

	// Create inserts a new {{.ModelStructNameLower}} into the database.
	Create(ctx context.Context, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) (uint, error)

	// Update updates an existing {{.ModelStructNameLower}} in the database.
	Update(ctx context.Context, id uint, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) error

	// Delete deletes a {{.ModelStructNameLower}} by its ID.
	Delete(ctx context.Context, id uint) error

	// List retrieves {{.ModelStructNameLower}} records based on query conditions.
	List(ctx context.Context, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) ([]{{.LowerStructName}}.{{.ModelStructName}}, error)

	// GetByID retrieves a {{.ModelStructNameLower}} by its ID.
	GetByID(ctx context.Context, id uint) (*{{.LowerStructName}}.{{.ModelStructName}}, error)
}

// {{.ModelStructNameLower}}Repo implements the {{.StructName}}Repo interface.
type {{.ModelStructNameLower}}Repo struct {
	redis *redis.Manager
	db    *gorm.DB
}

// New{{.StructName}}Repo creates a new instance of the {{.ModelStructNameLower}} repository.
//
// Parameters:
//   - db: A pointer to the gorm.DB instance for database operations.
//   - redis: A pointer to the redis.Manager for caching operations.
//
// Returns:
//   - {{.StructName}}Repo: An implementation of the {{.StructName}}Repo interface.
//
// Example:
//
//		db := // initialize gorm.DB
//		redisManager := // initialize redis.Manager
//		{{.ModelStructNameLower}}Repo := New{{.StructName}}Repo(db, redisManager)
func New{{.StructName}}Repo(db *gorm.DB, redis *redis.Manager) {{.StructName}}Repo {
	return &{{.ModelStructNameLower}}Repo{redis: redis, db: db}
}

// Get{{.ModelStructName}} retrieves a {{.ModelStructNameLower}} by its properties using the model's First method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - {{.ModelStructNameLower}}: *{{.LowerStructName}}.{{.ModelStructName}} the {{.ModelStructNameLower}} with search criteria.
//
// Returns:
//   - *{{.LowerStructName}}.{{.ModelStructName}}: pointer to the retrieved {{.ModelStructNameLower}}, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (r *{{.ModelStructNameLower}}Repo) Get{{.ModelStructName}}(ctx context.Context, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) (*{{.LowerStructName}}.{{.ModelStructName}}, error) {
	return {{.ModelStructNameLower}}.First(ctx, r.db)
}

// Create creates a new {{.ModelStructNameLower}} record using the model's Create method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - {{.ModelStructNameLower}}: *{{.LowerStructName}}.{{.ModelStructName}} the {{.ModelStructNameLower}} to create.
//
// Returns:
//   - uint: the ID of the created {{.ModelStructNameLower}}.
//   - error: error if the creation fails, otherwise nil.
func (r *{{.ModelStructNameLower}}Repo) Create(ctx context.Context, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) (uint, error) {
	return {{.ModelStructNameLower}}.Create(ctx, r.db)
}

// GetByID retrieves a {{.ModelStructNameLower}} by its ID using the model's Where and First methods.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the {{.ModelStructNameLower}} to retrieve.
//
// Returns:
//   - *{{.LowerStructName}}.{{.ModelStructName}}: pointer to the retrieved {{.ModelStructNameLower}}, or nil if not found.
//   - error: error if the query fails, otherwise nil.
func (r *{{.ModelStructNameLower}}Repo) GetByID(ctx context.Context, id uint) (*{{.LowerStructName}}.{{.ModelStructName}}, error) {
	{{.ModelStructNameLower}} := &{{.LowerStructName}}.{{.ModelStructName}}{}
	return {{.ModelStructNameLower}}.Where("id = ?", id).First(ctx, r.db)
}

// Update updates an existing {{.ModelStructNameLower}} record using the model's Updates method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the {{.ModelStructNameLower}} to update.
//   - {{.ModelStructNameLower}}: *{{.LowerStructName}}.{{.ModelStructName}} the {{.ModelStructNameLower}} with updated fields.
//
// Returns:
//   - error: error if the update fails, otherwise nil.
//
// Note: This method will only update non-zero value fields. You need to manually check
// each field and add it to the data map if it's not a zero value.
func (r *{{.ModelStructNameLower}}Repo) Update(ctx context.Context, id uint, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) error {
	data := make(map[string]interface{})
	
{{.FieldChecks}}
	
	if len(data) == 0 {
		return nil // No fields to update
	}

	updateModel := &{{.LowerStructName}}.{{.ModelStructName}}{}
	updateModel.ID = id

	return updateModel.Updates(ctx, r.db, data)
}

// Delete deletes a {{.ModelStructNameLower}} record using the model's Where and Delete methods.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - id: uint the ID of the {{.ModelStructNameLower}} to delete.
//
// Returns:
//   - error: error if the deletion fails, otherwise nil.
func (r *{{.ModelStructNameLower}}Repo) Delete(ctx context.Context, id uint) error {
	{{.ModelStructNameLower}} := &{{.LowerStructName}}.{{.ModelStructName}}{}
	return {{.ModelStructNameLower}}.Where("id = ?", id).Delete(ctx, r.db)
}

// List retrieves {{.ModelStructNameLower}} records based on query conditions using the model's List method.
//
// Parameters:
//   - ctx: context.Context for managing request-scoped values, cancellation signals, and deadlines.
//   - {{.ModelStructNameLower}}: *{{.LowerStructName}}.{{.ModelStructName}} the {{.ModelStructNameLower}} with query conditions.
//
// Returns:
//   - []{{.LowerStructName}}.{{.ModelStructName}}: slice of {{.ModelStructNameLower}} records.
//   - error: error if the query fails, otherwise nil.
func (r *{{.ModelStructNameLower}}Repo) List(ctx context.Context, {{.ModelStructNameLower}} *{{.LowerStructName}}.{{.ModelStructName}}) ([]{{.LowerStructName}}.{{.ModelStructName}}, error) {
	return {{.ModelStructNameLower}}.List(ctx, r.db)
}`
