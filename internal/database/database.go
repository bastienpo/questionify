package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Database interface {
	Health() map[string]string
	Close() error
}

type database struct {
	db     *sql.DB
	logger *slog.Logger
}

var (
	databaseName = os.Getenv("DB_DATABASE")
	username     = os.Getenv("DB_USER")
	password     = os.Getenv("DB_PASSWORD")
	port         = os.Getenv("DB_PORT")
	host         = os.Getenv("DB_HOST")
	dbInstance   *database
)

func New(logger *slog.Logger) (Database, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}

	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s", host, port, username, password, databaseName)
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	dbInstance = &database{db: db, logger: logger}
	return dbInstance, nil
}

func (d *database) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats := make(map[string]string)

	err := d.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "unavailable"
		stats["message"] = fmt.Sprintf("database is unavailable: %v", err)
		d.logger.Error("database is unavailable", "error", err)

		return stats
	}

	dbStats := d.db.Stats()

	stats["status"] = "available"
	stats["status"] = "available"
	stats["message"] = "database is healthy"
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["idle_connections"] = strconv.Itoa(dbStats.Idle)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_open_connections"] = strconv.Itoa(dbStats.MaxOpenConnections)

	if dbStats.OpenConnections > 40 {
		stats["message"] = "database is under heavy load"
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "database is experiencing delays"
	}

	if dbStats.MaxIdleClosed > int64(dbStats.MaxOpenConnections/2) {
		stats["message"] = "Too many connections are being closed, consider increasing max open connections"
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.MaxOpenConnections/2) {
		stats["message"] = "Too many connections are being closed, consider increasing max lifetime"
	}

	return stats
}

func (d *database) Close() error {
	d.logger.Info("closing database connection")
	return d.db.Close()
}
