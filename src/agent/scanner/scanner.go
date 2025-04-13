package scanner

import (
	"fmt"

	dbmanager "github.com/royceleond/antska/src/agent/dbManager"
)

type Scanner struct {
	DBManager *dbmanager.DBManager
}

func NewScanner(dbm *dbmanager.DBManager) *Scanner {
	return &Scanner{DBManager: dbm}
}

func (s *Scanner) ScanDatabase() (*DatabaseInfo, error) {
	if s.DBManager.DB == nil {
		return nil, fmt.Errorf("database connection not established")
	}

	dbName, err := s.getDatabaseName()
	if err != nil {
		return nil, err
	}

	schemas, err := s.getSchemas()
	if err != nil {
		return nil, err
	}

	totalTables := 0
	for _, schema := range schemas {
		totalTables += len(schema.Tables)
	}

	return &DatabaseInfo{
		Name:         dbName,
		Schemas:      schemas,
		TotalSchemas: len(schemas),
		TotalTables:  totalTables,
	}, nil
}

func (s *Scanner) getDatabaseName() (string, error) {
	var dbName string
	err := s.DBManager.DB.QueryRow("SELECT current_database()").Scan(&dbName)
	if err != nil {
		return "", fmt.Errorf("error getting database name: %v", err)
	}
	return dbName, nil
}

func (s *Scanner) getSchemas() ([]SchemaInfo, error) {
	rows, err := s.DBManager.DB.Query(`
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog')
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying schemas: %v", err)
	}
	defer rows.Close()

	var schemas []SchemaInfo
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("error scanning schema name: %v", err)
		}

		tables, err := s.getTables(schemaName)
		if err != nil {
			return nil, err
		}

		schemas = append(schemas, SchemaInfo{
			Name:   schemaName,
			Tables: tables,
		})
	}

	return schemas, nil
}

func (s *Scanner) getTables(schemaName string) ([]TableInfo, error) {
	rows, err := s.DBManager.DB.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1
	`, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error querying tables for schema %s: %v", schemaName, err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("error scanning table name: %v", err)
		}

		columnCount, err := s.getColumnCount(schemaName, tableName)
		if err != nil {
			return nil, err
		}

		rowCount, err := s.getRowCount(schemaName, tableName)
		if err != nil {
			return nil, err
		}

		tables = append(tables, TableInfo{
			Schema:      schemaName,
			Name:        tableName,
			ColumnCount: columnCount,
			RowCount:    rowCount,
		})
	}

	return tables, nil
}

func (s *Scanner) getColumnCount(schemaName, tableName string) (int, error) {
	var count int
	err := s.DBManager.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_schema = $1 AND table_name = $2
	`, schemaName, tableName).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting column count for table %s.%s: %v", schemaName, tableName, err)
	}
	return count, nil
}

func (s *Scanner) getRowCount(schemaName, tableName string) (int64, error) {
	var count int64
	err := s.DBManager.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", schemaName, tableName)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting row count for table %s.%s: %v", schemaName, tableName, err)
	}
	return count, nil
}

func (s *Scanner) GetColumnDetails(schemaName, tableName string) ([]ColumnInfo, error) {
	rows, err := s.DBManager.DB.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
	`, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("error querying column details: %v", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var isNullable string
		if err := rows.Scan(&col.Name, &col.DataType, &isNullable); err != nil {
			return nil, fmt.Errorf("error scanning column details: %v", err)
		}
		col.IsNullable = (isNullable == "YES")
		columns = append(columns, col)
	}

	return columns, nil
}
