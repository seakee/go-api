// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package app provides configuration management for the application.
// It includes structures for various configuration aspects such as system settings,
// logging, databases, caching, and external service integrations.
package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	envKey  = "RUN_ENV"  // Environment variable key for the running environment
	nameKey = "APP_NAME" // Environment variable key for the application name
)

var config *Config // Global configuration variable

// Config represents the entire application configuration.
type Config struct {
	System    SysConfig   `json:"system"`    // System-wide configuration
	Log       LogConfig   `json:"log"`       // Logging configuration
	Databases []Databases `json:"databases"` // Database configurations
	Cache     Cache       `json:"cache"`     // Caching configuration
	Redis     []Redis     `json:"redis"`     // Redis configurations
	Kafka     Kafka       `json:"kafka"`     // Kafka configuration
	Monitor   Monitor     `json:"monitor"`   // Monitoring configuration
	Notify    Notify      `json:"notify"`    // Notify configuration
}

// LogConfig defines logging configuration options.
type LogConfig struct {
	Driver  string `json:"driver"`   // Log driver: "stdout" or "file"
	Level   string `json:"level"`    // Log level: "debug", "info", "warn", "error", "fatal"
	LogPath string `json:"log_path"` // Log file path (only used when Driver is "file")
}

// SysConfig defines system-wide configuration options.
type SysConfig struct {
	Name         string        `json:"name"`          // Application name
	RunMode      string        `json:"run_mode"`      // Running mode
	HTTPPort     string        `json:"http_port"`     // HTTP server port
	ReadTimeout  time.Duration `json:"read_timeout"`  // Maximum request timeout
	WriteTimeout time.Duration `json:"write_timeout"` // Maximum response timeout
	Version      string        `json:"version"`       // Application version
	RootPath     string        `json:"root_path"`     // Root directory path
	DebugMode    bool          `json:"debug_mode"`    // Debug mode flag
	LangDir      string        `json:"lang_dir"`      // Language files directory
	DefaultLang  string        `json:"default_lang"`  // Default language
	EnvKey       string        `json:"env_key"`       // Environment key for reading runtime environment
	JwtSecret    string        `json:"jwt_secret"`    // JWT secret for authentication
	TokenExpire  time.Duration `json:"token_expire"`  // JWT token expiration time (in seconds)
	Env          string        `json:"env"`           // Runtime environment
}

// Databases defines database configuration options.
type Databases struct {
	Enable        bool          `json:"enable"`                     // Database enable flag
	DbType        string        `json:"db_type"`                    // Database type
	DbHost        string        `json:"db_host"`                    // Database host
	DbName        string        `json:"db_name"`                    // Database name
	DbUsername    string        `json:"db_username,omitempty"`      // Database username
	DbPassword    string        `json:"db_password,omitempty"`      // Database password
	DbMaxIdleConn int           `json:"db_max_idle_conn,omitempty"` // Maximum number of idle connections in the pool
	DbMaxOpenConn int           `json:"db_max_open_conn,omitempty"` // Maximum number of open connections to the database
	DbMaxLifetime time.Duration `json:"db_max_lifetime,omitempty"`  // Maximum amount of time a connection may be reused (in hours)
	AuthMechanism string        `json:"auth_mechanism"`             // Authentication mechanism (for MongoDB)
}

// Cache defines caching configuration options.
type Cache struct {
	Driver string `json:"driver"` // Cache driver
	Prefix string `json:"prefix"` // Cache key prefix
}

// Redis defines Redis configuration options.
type Redis struct {
	Name        string        `json:"name"`         // Redis connection name
	Enable      bool          `json:"enable"`       // Redis enable flag
	Host        string        `json:"host"`         // Redis host
	Auth        string        `json:"auth"`         // Redis authentication
	MaxIdle     int           `json:"max_idle"`     // Maximum number of idle connections in the pool
	MaxActive   int           `json:"max_active"`   // Maximum number of connections allocated by the pool at a given time
	IdleTimeout time.Duration `json:"idle_timeout"` // Close connections after remaining idle for this duration (in minutes)
	Prefix      string        `json:"prefix"`       // Redis key prefix
	DB          int           `json:"db"`           // Redis database number
}

// Kafka defines Kafka configuration options.
type Kafka struct {
	Brokers            []string `json:"brokers"`              // Kafka broker addresses
	MaxRetry           int      `json:"max_retry"`            // Maximum number of retries
	ClientID           string   `json:"client_id"`            // Kafka client ID
	ProducerEnable     bool     `json:"producer_enable"`      // Producer enable flag
	ConsumerEnable     bool     `json:"consumer_enable"`      // Consumer enable flag
	ConsumerGroup      string   `json:"consumer_group"`       // Consumer group name
	ConsumerTopics     []string `json:"consumer_topics"`      // Topics to consume
	ConsumerAutoSubmit bool     `json:"consumer_auto_submit"` // Auto-submit consumer offsets flag
}

// Monitor defines monitoring configuration options.
type Monitor struct {
	PanicRobot PanicRobot `json:"panic_robot"` // Panic robot configuration
}

// PanicRobot defines configuration for panic reporting.
type PanicRobot struct {
	Enable bool        `json:"enable"` // Panic robot enable flag
	Wechat robotConfig `json:"wechat"` // WeChat's configuration for panic reporting
	Feishu robotConfig `json:"feishu"` // Feishu configuration for panic reporting
}

// robotConfig defines configuration for messaging platforms.
type robotConfig struct {
	Enable  bool   `json:"enable"`   // Robot enable flag
	PushUrl string `json:"push_url"` // URL for pushing messages
}

// Notify defines notification configuration options.
type Notify struct {
	DefaultChannel string `json:"default_channel"`
	DefaultLevel   string `json:"default_level"`
	Lark           Lark   `json:"lark"`
}

// Lark defines Lark configuration options.
type Lark struct {
	Enable                 bool               `json:"enable"`
	DefaultSendChannelName string             `json:"default_send_channel_name"`
	ChannelSize            int                `json:"channel_size"`
	PoolSize               int                `json:"pool_size"`
	BotWebhooks            map[string]string  `json:"bot_webhooks"`
	Larks                  map[string]LarkApp `json:"larks"`
}

// LarkApp defines Lark application configuration options.
type LarkApp struct {
	AppType   string `json:"app_type"`
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
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
	checkConfig(config)

	return config, nil
}

// checkConfig performs validation checks on the loaded configuration.
// Currently, it only checks if the JwtSecret is set.
//
// Parameters:
//   - conf: *Config - A pointer to the configuration structure to check.
//
// The function will panic if the JwtSecret is empty.
func checkConfig(conf *Config) {
	if conf.System.JwtSecret == "" {
		log.Panicf("JwtSecret cannot be null")
	}
}

// GetConfig returns the global configuration object.
// This function should be called after LoadConfig has been executed.
//
// Returns:
//   - *Config: A pointer to the global configuration structure.
//
// Example usage:
//
//	cfg := GetConfig()
//	fmt.Printf("Application name: %s\n", cfg.System.Name)
func GetConfig() *Config {
	return config
}
