package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	
	"github.com/royceleond/scansca/internal/db"
)

// PostgresConnector implements the db.Connector interface for PostgreSQL
type PostgresConnector struct {
	pool             *pgxpool.Pool
	Config           db.ConnectionPoolConfig
	connectionString string
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresConnector) Connect(ctx context.Context, connectionString string) error {
	p.connectionString = connectionString
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Apply custom configuration
	config.MaxConns = int32(p.Config.MaxConnections)
	config.MinConns = int32(p.Config.MinConnections)
	config.MaxConnLifetime = p.Config.MaxConnLifetime
	config.MaxConnIdleTime = p.Config.MaxConnIdleTime

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

// Ping checks if the database connection is alive
func (p *PostgresConnector) Ping(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("not connected to database")
	}
	return p.pool.Ping(ctx)
}

// GetStatus returns the current status of the database connection
func (p *PostgresConnector) GetStatus(ctx context.Context) (*db.DatabaseStatus, error) {
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
	
	return &db.DatabaseStatus{
		Connected:       true,
		AcquiredConns:   int(stats.AcquiredConns()),
		IdleConns:       int(stats.IdleConns()),
		TotalConns:      int(stats.TotalConns()),
		MaxConns:        p.Config.MaxConnections,
		DatabaseVersion: version,
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
func (p *PostgresConnector) GetTableInfo(ctx context.Context, schema, table string) (*db.TableInfo, error) {
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

	var columns []db.ColumnInfo
	var primaryKey []string

	for rows.Next() {
		var col db.ColumnInfo
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

	tableInfo := &db.TableInfo{
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
func (p *PostgresConnector) ExecuteQuery(ctx context.Context, query string, args ...interface{}) (db.ResultSet, error) {
	if p.pool == nil {
		return db.ResultSet{Error: "not connected to database"}, fmt.Errorf("not connected to database")
	}

	startTime := time.Now()
	
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return db.ResultSet{Error: err.Error()}, fmt.Errorf("failed to execute query: %w", err)
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
			return db.ResultSet{
				Columns: columns,
				Rows:    result,
				Error:   err.Error(),
				ExecutionTime: time.Since(startTime).Seconds(),
			}, fmt.Errorf("failed to read row values: %w", err)
		}
		result = append(result, values)
	}

	if rows.Err() != nil {
		return db.ResultSet{
			Columns: columns,
			Rows:    result,
			Error:   rows.Err().Error(),
			ExecutionTime: time.Since(startTime).Seconds(),
		}, fmt.Errorf("error iterating over results: %w", rows.Err())
	}

	// Calculate execution time
	executionTime := time.Since(startTime).Seconds()

	return db.ResultSet{
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

// GetConnectorType returns the type of the connector
func (p *PostgresConnector) GetConnectorType() string {
	return "postgresql"
}