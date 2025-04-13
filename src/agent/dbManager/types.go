package dbmanager

import "database/sql"

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type DBManager struct {
	Config DBConfig
	DB     *sql.DB
}
