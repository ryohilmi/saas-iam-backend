package providers

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Don't forget the driver!
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func NewDatabase(config DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("postgresql://postgres:%s@%s:%s/%s?sslmode=disable", config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return db, nil
}
