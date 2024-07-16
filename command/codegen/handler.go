// Copyright 2024 Seakee. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/seakee/go-api/command/codegen/codegen"
	"log"
	"os"
	"path/filepath"
)

// main is the entry point of the program
// It defines and parses command line flags and calls the appropriate function to process SQL files based on the provided flags.
func main() {
	// Define command line flags
	force := flag.Bool("force", false, "force overwrite existing files")
	name := flag.String("name", "", "SQL file name (without .sql extension) in the bin/data/sql directory to generate code for")
	sqlPath := flag.String("sql", "bin/data/sql", "SQL directory")
	modelOutputPath := flag.String("model", "app/model", "Model directory")
	repoOutputPath := flag.String("repo", "app/repository", "Repository directory")
	serviceOutputPath := flag.String("service", "app/service", "Service directory")

	// Parse command line flags
	flag.Parse()

	if *name != "" {
		// If the name parameter is provided, process a single SQL file
		processSingleSQLFile(*force, *name, *sqlPath, *modelOutputPath, *repoOutputPath, *serviceOutputPath)
	} else {
		// Otherwise, process all SQL files in the sqlPath directory
		processSQLDirectory(*force, *sqlPath, *modelOutputPath, *repoOutputPath, *serviceOutputPath)
	}
}

// processSingleSQLFile processes a single SQL file and generates the corresponding code
//
// Parameters:
// - force: whether to force overwrite existing files
// - name: SQL file name (without .sql extension)
// - sqlPath: directory where the SQL file is located
// - modelOutputPath: directory to output the generated model code
// - repoOutputPath: directory to output the generated repository code
// - serviceOutputPath: directory to output the generated service code
func processSingleSQLFile(force bool, name, sqlPath, modelOutputPath, repoOutputPath, serviceOutputPath string) {
	// Create a new Model instance
	m := codegen.NewModel()

	// Construct the full path to the SQL file
	sqlFilePath := filepath.Join(sqlPath, name+".sql")
	// Generate the model code
	if err := m.Generate(force, sqlFilePath, modelOutputPath); err != nil {
		log.Fatalf("Failed to generate model from %s: %v", sqlFilePath, err)
	}

	// The following code generates repository and service code, currently commented out
	// repo := NewRepo(m)
	// if err := repo.Generate(m, repoOutputPath); err != nil {
	// 	log.Fatalf("Failed to generate repository from %s: %v", sqlFilePath, err)
	// }
	//
	// service := NewService(repo)
	// if err := service.Generate(repo, serviceOutputPath); err != nil {
	// 	log.Fatalf("Failed to generate service from %s: %v", sqlFilePath, err)
	// }
}

// processSQLDirectory processes all SQL files in a directory and generates the corresponding code
//
// Parameters:
// - force: whether to force overwrite existing files
// - sqlPath: directory where the SQL files are located
// - modelOutputPath: directory to output the generated model code
// - repoOutputPath: directory to output the generated repository code
// - serviceOutputPath: directory to output the generated service code
func processSQLDirectory(force bool, sqlPath, modelOutputPath, repoOutputPath, serviceOutputPath string) {
	// Walk through all files in the sqlPath directory
	err := filepath.Walk(sqlPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// If the file is not a directory and has a .sql extension, generate the corresponding code
		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			// Create a new Model instance
			m := codegen.NewModel()
			// Generate the model code
			if err = m.Generate(force, path, modelOutputPath); err != nil {
				return err
			}

			// The following code generates repository and service code, currently commented out
			// repo := NewRepo(m)
			// if err = repo.Generate(m, repoOutputPath); err != nil {
			// 	return err
			// }
			//
			// service := NewService(repo)
			// if err = service.Generate(repo, serviceOutputPath); err != nil {
			// 	return err
			// }
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to process SQL directory %s: %v", sqlPath, err)
	}
}
