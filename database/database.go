package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// FIXME: move this to env var
type DB struct {
	RO *sql.DB
	RW *sql.DB
}

// New creates a new DB connection (read-write and read-only)
func New() (*DB, error) {
	dsn := viper.GetString("DSN")
	if dsn == "" {
		return nil, errors.New("dsn cannot be empty")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db connection: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// TODO: use read-replica connection
	return &DB{
		RO: db,
		RW: db,
	}, nil
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create the postgres driver: %v", err)
	}

	// migrationsDir, err := filepath.Abs("./migrations")
	// if err != nil {
	// 	return fmt.Errorf("could not get absolute path to migrations directory: %v", err)
	// }

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("could not create the migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run the migrations: %v", err)
	}

	return nil
}
