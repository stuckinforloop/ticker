package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// FIXME: move this to env var
const DSN = "postgres://tsdbadmin:xugvkycqyl4k0puk@q4yexnqrsd.bxguec8z4z.tsdb.cloud.timescale.com:37104/tsdb?sslmode=require"

type DB struct {
	RO *sql.DB
	RW *sql.DB
}

// New creates a new DB connection (read-write and read-only)
func New() (*DB, error) {
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, fmt.Errorf("open db connection: %w", err)
	}

	// TODO: run db migrations

	return &DB{
		RO: db,
		RW: db,
	}, nil
}
