package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Database interface {
	Close() error
}

type database struct {
	db     *sql.DB
	logger *slog.Logger
}

var dbInstance *database

func New(logger *slog.Logger) (Database, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" || !strings.HasPrefix(dsn, "postgres://") {
		return nil, fmt.Errorf("invalid database dsn format")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	logger.Info("Connected to postgresql database")

	dbInstance = &database{db: db, logger: logger}
	return dbInstance, nil
}

func (d *database) Close() error {
	d.logger.Info("Closing database connection")
	return d.db.Close()
}
