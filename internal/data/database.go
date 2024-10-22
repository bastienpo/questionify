package data

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbInstance *sql.DB

func NewDatabase(logger *slog.Logger) (*sql.DB, error) {
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

	dbInstance = db
	return dbInstance, nil
}
