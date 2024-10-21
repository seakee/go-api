package config

import "time"

// Database defines database configuration options.
type Database struct {
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
