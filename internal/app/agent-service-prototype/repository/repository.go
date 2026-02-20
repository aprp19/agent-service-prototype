package repository

import (
	"database/sql"
)

// Repository provides data access (raw DB and, when available, Ent).
// Migration-related methods are in migration.go.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a repository with the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// DB returns the underlying *sql.DB for connection management (e.g. Conn).
func (r *Repository) DB() *sql.DB {
	return r.db
}

// ---------------------------------------------------------------------------
// Ent-based code (commented until code generation is run)
// ---------------------------------------------------------------------------
//
// type Repository struct {
// 	client    *ent.Client
// 	rawHelper *database.RawQueryHelper
// }
//
// func NewRepository(client *ent.Client, db *sql.DB) *Repository { ... }
// func (r *Repository) ExampleGetByID(...) (*ent.ExampleEntity, error) { ... }
// func (r *Repository) ExampleGetByIDRaw(...) (*ent.ExampleEntity, error) { ... }
