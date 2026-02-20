package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// MigrationRecord represents one row in hris_meta.schema_migrations.
type MigrationRecord struct {
	Version    string
	Name       string
	Checksum   string
	AppliedAt  time.Time
	ExecTimeMs int64
	Success    bool
	ErrorMsg   string
}

// IsFreshDB reports whether hris_meta.schema_migrations does not exist.
// conn must be the same connection that holds the advisory lock.
func (r *Repository) IsFreshDB(ctx context.Context, conn *sql.Conn) (bool, error) {
	var exists bool
	err := conn.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'hris_meta' AND table_name = 'schema_migrations'
		)
	`).Scan(&exists)
	if err != nil {
		return false, err
	}
	return !exists, nil
}

// EnsureMigrationsTable creates hris_meta and hris_meta.schema_migrations if not present.
func (r *Repository) EnsureMigrationsTable(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(ctx, `
		CREATE SCHEMA IF NOT EXISTS hris_meta;
		CREATE TABLE IF NOT EXISTS hris_meta.schema_migrations (
			version           TEXT PRIMARY KEY,
			name              TEXT NOT NULL,
			checksum          TEXT NOT NULL,
			applied_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			execution_time_ms BIGINT NOT NULL DEFAULT 0,
			success           BOOLEAN NOT NULL DEFAULT FALSE,
			error             TEXT
		);
	`)
	return err
}

// GetMigrationRecord returns the migration row for version, or nil if not found.
func (r *Repository) GetMigrationRecord(ctx context.Context, conn *sql.Conn, version string) (*MigrationRecord, error) {
	var rec MigrationRecord
	var errMsg sql.NullString
	var appliedAt time.Time
	err := conn.QueryRowContext(ctx, `
		SELECT version, name, checksum, applied_at, execution_time_ms, success, error
		FROM hris_meta.schema_migrations
		WHERE version = $1
	`, version).Scan(&rec.Version, &rec.Name, &rec.Checksum, &appliedAt, &rec.ExecTimeMs, &rec.Success, &errMsg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rec.AppliedAt = appliedAt
	if errMsg.Valid {
		rec.ErrorMsg = errMsg.String
	}
	return &rec, nil
}

// RecordMigration inserts or updates a row in hris_meta.schema_migrations.
func (r *Repository) RecordMigration(ctx context.Context, conn *sql.Conn, rec MigrationRecord) error {
	_, err := conn.ExecContext(ctx, `
		INSERT INTO hris_meta.schema_migrations
			(version, name, checksum, applied_at, execution_time_ms, success, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (version) DO UPDATE SET
			name              = EXCLUDED.name,
			checksum          = EXCLUDED.checksum,
			applied_at        = EXCLUDED.applied_at,
			execution_time_ms = EXCLUDED.execution_time_ms,
			success           = EXCLUDED.success,
			error             = EXCLUDED.error
	`, rec.Version, rec.Name, rec.Checksum, rec.AppliedAt, rec.ExecTimeMs, rec.Success, rec.ErrorMsg)
	return err
}

// ExecInTransaction runs query in a single transaction on conn.
func (r *Repository) ExecInTransaction(ctx context.Context, conn *sql.Conn, query string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	if _, err := tx.ExecContext(ctx, query); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
