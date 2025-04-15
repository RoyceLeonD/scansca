package db

import (
	"time"
)

// ResultSet represents a database query result
type ResultSet struct {
	Columns       []string        `json:"columns"`
	Rows          [][]interface{} `json:"rows"`
	Error         string          `json:"error,omitempty"`
	AffectedRows  int64           `json:"affected_rows,omitempty"`
	ExecutionTime float64         `json:"execution_time,omitempty"` // in seconds
}

// ColumnInfo contains detailed information about a database column
type ColumnInfo struct {
	Name             string `json:"name"`
	DataType         string `json:"data_type"`
	IsNullable       bool   `json:"is_nullable"`
	DefaultValue     string `json:"default_value,omitempty"`
	IsPrimaryKey     bool   `json:"is_primary_key"`
	IsForeignKey     bool   `json:"is_foreign_key"`
	ReferencesTable  string `json:"references_table,omitempty"`
	ReferencesColumn string `json:"references_column,omitempty"`
}

// TableInfo contains detailed information about a database table
type TableInfo struct {
	Schema           string       `json:"schema"`
	Name             string       `json:"name"`
	Columns          []ColumnInfo `json:"columns"`
	PrimaryKey       []string     `json:"primary_key,omitempty"`
	EstimatedRowCount int64       `json:"estimated_row_count"`
	CreateTime       time.Time    `json:"create_time"`
	Description      string       `json:"description,omitempty"`
}

// DatabaseConfig holds configuration for a database connection
type DatabaseConfig struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	ConnectionString string                 `json:"connection_string"`
	Options          map[string]interface{} `json:"options,omitempty"`
}

// ConnectionPoolConfig holds common connection pool configuration
type ConnectionPoolConfig struct {
	MaxConnections  int           `json:"max_connections"`
	MinConnections  int           `json:"min_connections"`
	MaxConnLifetime time.Duration `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `json:"max_conn_idle_time"`
}

// DefaultConnectionPoolConfig returns a default connection pool configuration
func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxConnections:  10,
		MinConnections:  2,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 15 * time.Minute,
	}
}

// DatabaseStatus represents the current status of a database connection
type DatabaseStatus struct {
	Connected         bool                   `json:"connected"`
	AcquiredConns     int                    `json:"acquired_connections"`
	IdleConns         int                    `json:"idle_connections"`
	TotalConns        int                    `json:"total_connections"`
	MaxConns          int                    `json:"max_connections"`
	DatabaseVersion   string                 `json:"database_version,omitempty"`
	AdditionalDetails map[string]interface{} `json:"additional_details,omitempty"`
}