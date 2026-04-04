package postgres

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	maxRetries = 10
	retryDelay = 2 * time.Second
)

func Open(dsn string, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open %w", err)
	}

	// Connection pool tuning
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = db.Ping()
		if err == nil {
			logger.Info("connected to database",
				slog.Int("max_open_conns", 25),
				slog.Int("max_idle_conns", 5),
			)
			return db, nil
		}

		logger.Warn("database not ready, retrying",
			slog.Int("attempt", attempt),
			slog.Int("max_retries", maxRetries),
			slog.String("error", err.Error()),
		)
		time.Sleep(retryDelay)
	}

	db.Close()
	return nil, fmt.Errorf("db: ping failed after %d attempts: %w", maxRetries, err)
}

func MigrateFS(db *sql.DB, migrationsFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationsFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
