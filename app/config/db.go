package config

import "time"

// Database defines database configuration options.
type Database struct {
	Enable          bool          `json:"enable"`                        // Database enable flag
	DbType          string        `json:"db_type"`                       // Database type (mysql, postgres, sqlite, sqlserver, clickhouse, mongo)
	DbHost          string        `json:"db_host"`                       // Database host (for SQLite, this is the file path)
	DbPort          int           `json:"db_port,omitempty"`             // Database port
	DbName          string        `json:"db_name"`                       // Database name
	DbUsername      string        `json:"db_username,omitempty"`         // Database username
	DbPassword      string        `json:"db_password,omitempty"`         // Database password
	Charset         string        `json:"charset,omitempty"`             // Character set (for MySQL, default: utf8mb4)
	SSLMode         string        `json:"ssl_mode,omitempty"`            // SSL mode (for PostgreSQL)
	Timezone        string        `json:"timezone,omitempty"`            // Timezone (for PostgreSQL and ClickHouse)
	DbMaxIdleConn   int           `json:"db_max_idle_conn,omitempty"`    // Maximum number of idle connections in the pool
	DbMaxOpenConn   int           `json:"db_max_open_conn,omitempty"`    // Maximum number of open connections to the database
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime,omitempty"`   // Maximum connection lifetime (in hours)
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time,omitempty"`  // Maximum idle time for connections (in hours)
	AuthMechanism   string        `json:"auth_mechanism,omitempty"`      // Authentication mechanism (for MongoDB)
}
