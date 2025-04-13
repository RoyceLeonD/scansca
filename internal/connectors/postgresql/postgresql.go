package postgresql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ResultSet represents a database query result
type ResultSet struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
}

// PostgresConnector implements the Connector interface for PostgreSQL
type PostgresConnector struct {
	pool *pgxpool.Pool
}

// NewPostgresConnector creates a new PostgreSQL connector
func NewPostgresConnector() *PostgresConnector {
	return &PostgresConnector{}
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresConnector) Connect(ctx context.Context, connectionString string) error {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

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
	}
	return nil
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

// ExecuteQuery executes a query and returns the results
func (p *PostgresConnector) ExecuteQuery(ctx context.Context, query string) (ResultSet, error) {
	if p.pool == nil {
		return ResultSet{Error: "not connected to database"}, fmt.Errorf("not connected to database")
	}

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
			}, fmt.Errorf("failed to read row values: %w", err)
		}
		result = append(result, values)
	}

	if rows.Err() != nil {
		return ResultSet{
			Columns: columns,
			Rows:    result,
			Error:   rows.Err().Error(),
		}, fmt.Errorf("error iterating over results: %w", rows.Err())
	}

	return ResultSet{
		Columns: columns,
		Rows:    result,
	}, nil
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