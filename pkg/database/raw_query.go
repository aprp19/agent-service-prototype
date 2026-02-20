package database

import (
	"context"
	"database/sql"
	"fmt"
)

// RawQueryHelper provides utilities for executing raw SQL queries across different schemas
type RawQueryHelper struct {
	db *sql.DB
}

// NewRawQueryHelper creates a new RawQueryHelper instance from a database connection
func NewRawQueryHelper(db *sql.DB) *RawQueryHelper {
	return &RawQueryHelper{
		db: db,
	}
}

// NewRawQueryHelperFromDSN creates a new RawQueryHelper instance from a DSN string
func NewRawQueryHelperFromDSN(driverName, dataSourceName string) (*RawQueryHelper, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &RawQueryHelper{
		db: db,
	}, nil
}

// WithSchema returns a SchemaQueryExecutor for the specified schema
func (h *RawQueryHelper) WithSchema(schema string) *SchemaQueryExecutor {
	return &SchemaQueryExecutor{
		db:     h.db,
		schema: schema,
	}
}

// SchemaQueryExecutor executes queries within a specific schema
type SchemaQueryExecutor struct {
	db     *sql.DB
	schema string
}

// QueryRow executes a query that returns at most one row
func (e *SchemaQueryExecutor) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db := e.getDB()
	schemaQuery := e.setSearchPath(query)
	return db.QueryRowContext(ctx, schemaQuery, args...)
}

// Query executes a query that returns rows
func (e *SchemaQueryExecutor) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db := e.getDB()
	schemaQuery := e.setSearchPath(query)
	return db.QueryContext(ctx, schemaQuery, args...)
}

// Exec executes a query without returning any rows (INSERT, UPDATE, DELETE)
func (e *SchemaQueryExecutor) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db := e.getDB()
	schemaQuery := e.setSearchPath(query)
	return db.ExecContext(ctx, schemaQuery, args...)
}

// QueryRowDirect executes a query without setting search_path
func (e *SchemaQueryExecutor) QueryRowDirect(ctx context.Context, query string, args ...interface{}) *sql.Row {
	db := e.getDB()
	return db.QueryRowContext(ctx, query, args...)
}

// QueryDirect executes a query without setting search_path
func (e *SchemaQueryExecutor) QueryDirect(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db := e.getDB()
	return db.QueryContext(ctx, query, args...)
}

// ExecDirect executes a query without setting search_path
func (e *SchemaQueryExecutor) ExecDirect(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db := e.getDB()
	return db.ExecContext(ctx, query, args...)
}

// Transaction executes a function within a database transaction
func (e *SchemaQueryExecutor) Transaction(ctx context.Context, fn func(context.Context, *sql.Tx) error) error {
	db := e.getDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set search_path for the transaction
	if e.schema != "" {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", e.schema)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set search_path: %w", err)
		}
	}

	// Execute the transaction function
	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TransactionWithOptions executes a function within a database transaction with custom options
func (e *SchemaQueryExecutor) TransactionWithOptions(ctx context.Context, opts *sql.TxOptions, fn func(context.Context, *sql.Tx) error) error {
	db := e.getDB()

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set search_path for the transaction
	if e.schema != "" {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", e.schema)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to set search_path: %w", err)
		}
	}

	// Execute the transaction function
	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRawDB returns the underlying *sql.DB for advanced usage
func (e *SchemaQueryExecutor) GetRawDB() *sql.DB {
	return e.db
}

// getDB returns the underlying *sql.DB
func (e *SchemaQueryExecutor) getDB() *sql.DB {
	return e.db
}

// setSearchPath wraps the query with SET search_path if schema is specified
func (e *SchemaQueryExecutor) setSearchPath(query string) string {
	if e.schema == "" {
		return query
	}
	return fmt.Sprintf("SET search_path TO %s; %s", e.schema, query)
}

// BatchQuery represents a batch of queries to execute
type BatchQuery struct {
	Query string
	Args  []interface{}
}

// ExecuteBatch executes multiple queries in a single transaction
func (e *SchemaQueryExecutor) ExecuteBatch(ctx context.Context, queries []BatchQuery) error {
	return e.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		for _, q := range queries {
			if _, err := tx.ExecContext(ctx, q.Query, q.Args...); err != nil {
				return fmt.Errorf("failed to execute query '%s': %w", q.Query, err)
			}
		}
		return nil
	})
}
