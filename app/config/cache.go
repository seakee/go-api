package config

import "time"

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
