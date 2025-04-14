package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ResultSet represents a database query result
type ResultSet struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
	// Additional metadata about the query
	AffectedRows int64     `json:"affected_rows,omitempty"`
	ExecutionTime float64  `json:"execution_time,omitempty"` // in seconds
}

// ColumnInfo contains detailed information about a database column
type ColumnInfo struct {
	Name          string `json:"name"`
	DataType      string `json:"data_type"`
	IsNullable    bool   `json:"is_nullable"`
	DefaultValue  string `json:"default_value,omitempty"`
	IsPrimaryKey  bool   `json:"is_primary_key"`
	IsForeignKey  bool   `json:"is_foreign_key"`
	ReferencesTable string `json:"references_table,omitempty"`
	ReferencesColumn string `json:"references_column,omitempty"`
}

// TableInfo contains detailed information about a database table
type TableInfo struct {
	Schema      string      `json:"schema"`
	Name        string      `json:"name"`
	Columns     []ColumnInfo `json:"columns"`
	PrimaryKey  []string    `json:"primary_key,omitempty"`
	EstimatedRowCount int64   `json:"estimated_row_count"`
	CreateTime  time.Time   `json:"create_time"`
	Description string      `json:"description,omitempty"`
}

// ConnectionConfig holds configuration options for the PostgreSQL connector
type ConnectionConfig struct {
	MaxConnections int
	MinConnections int
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultConnectionConfig returns a default configuration for PostgreSQL connections
func DefaultConnectionConfig() ConnectionConfig {
	return ConnectionConfig{
		MaxConnections: 10,
		MinConnections: 2,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 15 * time.Minute,
	}
}

// PostgresConnector implements the Connector interface for PostgreSQL
type PostgresConnector struct {
	pool *pgxpool.Pool
	config ConnectionConfig
	connectionString string
}

// NewPostgresConnector creates a new PostgreSQL connector
func NewPostgresConnector() *PostgresConnector {
	return &PostgresConnector{
		config: DefaultConnectionConfig(),
	}
}

// NewPostgresConnectorWithConfig creates a new PostgreSQL connector with custom configuration
func NewPostgresConnectorWithConfig(config ConnectionConfig) *PostgresConnector {
	return &PostgresConnector{
		config: config,
	}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresConnector) Connect(ctx context.Context, connectionString string) error {
	p.connectionString = connectionString
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply custom configuration
	config.MaxConns = int32(p.config.MaxConnections)
	config.MinConns = int32(p.config.MinConnections)
	config.MaxConnLifetime = p.config.MaxConnLifetime
	config.MaxConnIdleTime = p.config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.pool = pool
	return nil
}

// Disconnect closes the PostgreSQL connection
func (p *PostgresConnector) Disconnect(ctx context.Context) error {
	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
	return nil
}

// Reconnect attempts to reconnect to the database using the stored connection string
func (p *PostgresConnector) Reconnect(ctx context.Context) error {
	if p.connectionString == "" {
		return fmt.Errorf("no previous connection string available")
	}
	
	if p.pool != nil {
		p.pool.Close()
		p.pool = nil
	}
	
	return p.Connect(ctx, p.connectionString)
}

// GetStatus returns the current status of the database connection
func (p *PostgresConnector) GetStatus(ctx context.Context) (map[string]interface{}, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}
	
	stats := p.pool.Stat()
	
	// Add database version information
	var version string
	err := p.pool.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		return nil, fmt.Errorf("failed to get database version: %w", err)
	}
	
	return map[string]interface{}{
		"acquired_connections": stats.AcquiredConns(),
		"idle_connections": stats.IdleConns(),
		"total_connections": stats.TotalConns(),
		"max_connections": p.config.MaxConnections,
		"database_version": version,
	}, nil
}

// ListSchemas returns a list of schemas in the database
func (p *PostgresConnector) ListSchemas(ctx context.Context) ([]string, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := p.pool.Query(ctx, `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema')
		ORDER BY schema_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, fmt.Errorf("failed to scan schema: %w", err)
		}
		schemas = append(schemas, schema)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over schemas: %w", rows.Err())
	}

	return schemas, nil
}

// ListTables returns a list of tables in the specified schema
func (p *PostgresConnector) ListTables(ctx context.Context, schema string) ([]string, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := p.pool.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("failed to scan table: %w", err)
		}
		tables = append(tables, table)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over tables: %w", rows.Err())
	}

	return tables, nil
}

// ListViews returns a list of views in the specified schema
func (p *PostgresConnector) ListViews(ctx context.Context, schema string) ([]string, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := p.pool.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 
		AND table_type = 'VIEW'
		ORDER BY table_name
	`, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query views: %w", err)
	}
	defer rows.Close()

	var views []string
	for rows.Next() {
		var view string
		if err := rows.Scan(&view); err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		views = append(views, view)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over views: %w", rows.Err())
	}

	return views, nil
}

// GetTableColumns returns information about the columns in a table
func (p *PostgresConnector) GetTableColumns(ctx context.Context, schema, table string) (map[string]string, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	rows, err := p.pool.Query(ctx, `
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_schema = $1 AND table_name = $2 
		ORDER BY ordinal_position
	`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]string)
	for rows.Next() {
		var column, dataType string
		if err := rows.Scan(&column, &dataType); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		columns[column] = dataType
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over columns: %w", rows.Err())
	}

	return columns, nil
}

// GetTableInfo returns detailed information about a table
func (p *PostgresConnector) GetTableInfo(ctx context.Context, schema, table string) (*TableInfo, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	// Get detailed column information
	rows, err := p.pool.Query(ctx, `
		WITH pk_columns AS (
			SELECT a.attname as column_name
			FROM pg_index i
			JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
			WHERE i.indrelid = ($1 || '.' || $2)::regclass
			AND i.indisprimary
		),
		fk_info AS (
			SELECT
				kcu.column_name,
				ccu.table_schema AS foreign_table_schema,
				ccu.table_name AS foreign_table_name,
				ccu.column_name AS foreign_column_name
			FROM information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu
				ON ccu.constraint_name = tc.constraint_name
			WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = $1
			AND tc.table_name = $2
		)
		SELECT 
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' as is_nullable,
			c.column_default,
			(c.column_name IN (SELECT column_name FROM pk_columns)) as is_primary_key,
			fk.column_name IS NOT NULL as is_foreign_key,
			fk.foreign_table_name as references_table,
			fk.foreign_column_name as references_column
		FROM information_schema.columns c
		LEFT JOIN fk_info fk ON c.column_name = fk.column_name
		WHERE c.table_schema = $1
		AND c.table_name = $2
		ORDER BY c.ordinal_position
	`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table information: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	var primaryKey []string

	for rows.Next() {
		var col ColumnInfo
		var defaultValue *string
		var referencesTable, referencesColumn *string

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&col.IsNullable,
			&defaultValue,
			&col.IsPrimaryKey,
			&col.IsForeignKey,
			&referencesTable,
			&referencesColumn,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		if defaultValue != nil {
			col.DefaultValue = *defaultValue
		}
		
		if referencesTable != nil {
			col.ReferencesTable = *referencesTable
		}
		
		if referencesColumn != nil {
			col.ReferencesColumn = *referencesColumn
		}

		columns = append(columns, col)

		if col.IsPrimaryKey {
			primaryKey = append(primaryKey, col.Name)
		}
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over column info: %w", rows.Err())
	}

	// Get table statistics (simplified version to avoid compatibility issues)
	var estimatedRowCount int64
	var description *string

	err = p.pool.QueryRow(ctx, `
		SELECT
			COALESCE(reltuples::bigint, 0) as estimated_row_count,
			COALESCE(obj_description(c.oid), '') as description
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1
		AND c.relname = $2
	`, schema, table).Scan(&estimatedRowCount, &description)
	if err != nil {
		return nil, fmt.Errorf("failed to get table statistics: %w", err)
	}

	tableInfo := &TableInfo{
		Schema:     schema,
		Name:       table,
		Columns:    columns,
		PrimaryKey: primaryKey,
		EstimatedRowCount: estimatedRowCount,
		CreateTime: time.Now(), // Use current time as we don't have a reliable way to get creation time
	}

	if description != nil {
		tableInfo.Description = *description
	}

	return tableInfo, nil
}

// ExecuteQuery executes a query and returns the results
func (p *PostgresConnector) ExecuteQuery(ctx context.Context, query string) (ResultSet, error) {
	if p.pool == nil {
		return ResultSet{Error: "not connected to database"}, fmt.Errorf("not connected to database")
	}

	startTime := time.Now()
	
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return ResultSet{Error: err.Error()}, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	// Read all rows
	var result [][]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return ResultSet{
				Columns: columns,
				Rows:    result,
				Error:   err.Error(),
				ExecutionTime: time.Since(startTime).Seconds(),
			}, fmt.Errorf("failed to read row values: %w", err)
		}
		result = append(result, values)
	}

	if rows.Err() != nil {
		return ResultSet{
			Columns: columns,
			Rows:    result,
			Error:   rows.Err().Error(),
			ExecutionTime: time.Since(startTime).Seconds(),
		}, fmt.Errorf("error iterating over results: %w", rows.Err())
	}

	// Calculate execution time
	executionTime := time.Since(startTime).Seconds()

	return ResultSet{
		Columns: columns,
		Rows:    result,
		AffectedRows: int64(len(result)),
		ExecutionTime: executionTime,
	}, nil
}

// ExecuteStatement executes a SQL statement (INSERT, UPDATE, DELETE) and returns affected rows
func (p *PostgresConnector) ExecuteStatement(ctx context.Context, statement string, args ...interface{}) (int64, error) {
	if p.pool == nil {
		return 0, fmt.Errorf("not connected to database")
	}

	result, err := p.pool.Exec(ctx, statement, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute statement: %w", err)
	}

	return result.RowsAffected(), nil
}

// ExecuteTransaction executes multiple statements in a transaction
func (p *PostgresConnector) ExecuteTransaction(ctx context.Context, stmts []string) error {
	if p.pool == nil {
		return fmt.Errorf("not connected to database")
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	for _, stmt := range stmts {
		_, err = tx.Exec(ctx, stmt)
		if err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteBatch executes multiple statements efficiently using a batch
func (p *PostgresConnector) ExecuteBatch(ctx context.Context, stmts []string) error {
	if p.pool == nil {
		return fmt.Errorf("not connected to database")
	}

	batch := &pgx.Batch{}
	for _, stmt := range stmts {
		batch.Queue(stmt)
	}

	results := p.pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("batch execution failed at statement %d: %w", i, err)
		}
	}

	return nil
}

// GetConnectorType returns the type of the connector
func (p *PostgresConnector) GetConnectorType() string {
	return "postgresql"
}

// Ping checks if the database connection is alive
func (p *PostgresConnector) Ping(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("not connected to database")
	}
	return p.pool.Ping(ctx)
}