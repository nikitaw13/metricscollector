package repository

import (
	"context"
	"database/sql"
)

// PostgresStorage wraps a sql.DB connection for PostgreSQL operations.
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage creates a new PostgresStorage instance
func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db: db,
	}
}

// PingContext verifies the database connection is alive using the provided context.
func (ps *PostgresStorage) PingContext(ctx context.Context) error {
	return ps.db.PingContext(ctx)
}
