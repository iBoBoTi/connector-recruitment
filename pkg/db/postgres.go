package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/iBoBoTi/connector-service/config"
)

func NewPostgresDB(cfg config.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
