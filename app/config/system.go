package config

import (
	"time"
)

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
