// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package config provides configuration management for the application.
// It includes structures for various configuration aspects such as system settings,
// logging, databases, caching, and external service integrations.
package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	envKey  = "RUN_ENV"  // Environment variable key for the running environment
	nameKey = "APP_NAME" // Environment variable key for the application name
)

var config *Config // Global configuration variable

// Config represents the entire application configuration.
type Config struct {
	System    SysConfig  `json:"system"`    // System-wide configuration
	Log       LogConfig  `json:"log"`       // Logging configuration
	Databases []Database `json:"databases"` // Database configurations
	Cache     Cache      `json:"cache"`     // Caching configuration
	Redis     []Redis    `json:"redis"`     // Redis configurations
	Kafka     Kafka      `json:"kafka"`     // Kafka configuration
	Monitor   Monitor    `json:"monitor"`   // Monitoring configuration
	Notify    Notify     `json:"notify"`    // Notify configuration
}

// LoadConfig loads the application configuration from a JSON file.
// It determines the configuration file to load based on the runtime environment,
// unmarshal the JSON content into a Config struct, and performs some post-processing.
//
// The function uses environment variables to determine the runtime environment and application name.
// If these are not set, it falls back to default values.
//
// Returns:
//   - *Config: A pointer to the loaded configuration structure.
//   - error: An error if any occurred during the loading process.
func LoadConfig() (*Config, error) {
	var (
		runEnv     string
		appName    string
		rootPath   string
		cfgContent []byte
		err        error
	)

	// Get the runtime environment from environment variable, default to "local"
	runEnv = os.Getenv(envKey)
	if runEnv == "" {
		runEnv = "local"
	}

	// Get the current working directory
	rootPath, err = os.Getwd()
	if err != nil {
		log.Fatalf("Unable to get working directory: %v", err)
	}

	// Construct the configuration file path
	configFilePath := filepath.Join(rootPath, "bin", "configs", fmt.Sprintf("%s.json", runEnv))
	cfgContent, err = os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON content into the config struct
	err = json.Unmarshal(cfgContent, &config)
	if err != nil {
		return nil, err
	}

	// Override application name if set in environment variable
	appName = os.Getenv(nameKey)
	if appName != "" {
		config.System.Name = appName
	}

	// Set additional system configuration fields
	config.System.Env = runEnv
	config.System.RootPath = rootPath
	config.System.EnvKey = envKey
	config.System.LangDir = filepath.Join(rootPath, "bin", "lang")

	// Perform configuration checks
	err = checkConfig(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// checkConfig performs validation checks on the loaded configuration.
// Currently, it only checks if the JwtSecret is set.
//
// Parameters:
//   - conf: *Config - A pointer to the configuration structure to check.
//
// The function will panic if the JwtSecret is empty.
func checkConfig(conf *Config) error {
	if conf.System.JwtSecret == "" {
		return fmt.Errorf("jwtSecret cannot be null")
	}

	if conf.System.ReadTimeout <= 0 {
		return fmt.Errorf("readTimeout cannot be less than or equal to zero")
	}

	if conf.System.WriteTimeout <= 0 {
		return fmt.Errorf("writeTimeout cannot be less than or equal to zero")
	}

	if conf.System.HTTPPort == "" {
		return fmt.Errorf("httpPort cannot be null")
	}

	if conf.System.TokenExpire <= 0 {
		return fmt.Errorf("TokenExpire cannot be less than or equal to zero")
	}

	return nil
}

// Get returns the global configuration object.
// This function should be called after LoadConfig has been executed.
//
// Returns:
//   - *Config: A pointer to the global configuration structure.
//
// Example usage:
//
//	cfg := Get()
//	fmt.Printf("Application name: %s\n", cfg.System.Name)
func Get() *Config {
	return config
}
