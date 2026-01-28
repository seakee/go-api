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
	Admin        AdminConfig   `json:"admin"`         // Admin configuration
}

// AdminConfig defines admin configuration options.
type AdminConfig struct {
	SafeCodeExpireIn int         `json:"safe_code_expire_in"`
	TokenExpireIn    int64       `json:"token_expire_in"`
	JwtSecret        string      `json:"jwt_secret"`
	Oauth            OauthConfig `json:"oauth"`
}

// OauthConfig defines OAuth provider configuration options.
type OauthConfig struct {
	RedirectURL string `json:"redirect_url"`
	Feishu      Feishu `json:"feishu"`
	Wechat      Wechat `json:"wechat"`
}

// Feishu defines Feishu (Lark) OAuth configuration.
type Feishu struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	OauthURL     string `json:"oauth_url"`
}

// Wechat defines WeChat Work (Enterprise WeChat) OAuth configuration.
type Wechat struct {
	CorpID     string `json:"corp_id"`     // Corporation ID
	AgentID    string `json:"agent_id"`    // Application ID
	CorpSecret string `json:"corp_secret"` // Application secret
	OauthURL   string `json:"oauth_url"`   // URL for obtaining user authorization code
	ProxyURL   string `json:"proxy_url"`   // Proxy URL
}
