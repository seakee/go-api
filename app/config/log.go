package config

// LogConfig defines logging configuration options.
type LogConfig struct {
	Driver  string `json:"driver"`   // Log driver: "stdout" or "file"
	Level   string `json:"level"`    // Log level: "debug", "info", "warn", "error", "fatal"
	LogPath string `json:"log_path"` // Log file path (only used when Driver is "file")
}
