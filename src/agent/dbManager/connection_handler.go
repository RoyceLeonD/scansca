package dbmanager

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewDBManager(config DBConfig) *DBManager {
	return &DBManager{Config: config}
}

func (dm *DBManager) Connect() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dm.Config.Host, dm.Config.Port, dm.Config.User, dm.Config.Password, dm.Config.DBName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error opening database connection: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %v", err)
	}

	dm.DB = db
	return nil
}

func (dm *DBManager) Close() error {
	if dm.DB != nil {
		return dm.DB.Close()
	}
	return nil
}

func (dm *DBManager) ExecuteQuery(query string) (*sql.Rows, error) {
	if dm.DB == nil {
		return nil, fmt.Errorf("database connection not established")
	}
	return dm.DB.Query(query)
}
