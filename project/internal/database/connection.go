package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func ConnectBookingDB(cfg DBConfig) (*sql.DB, error) {
	return connectDB(cfg)
}

func ConnectHotelDB(cfg DBConfig) (*sql.DB, error) {
	return connectDB(cfg)
}

func connectDB(cfg DBConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
