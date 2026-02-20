package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// This file demonstrates how to use the RawQueryHelper

// Example1_SimpleQuery demonstrates how to execute a simple SELECT query
func Example1_SimpleQuery(db *sql.DB) error {
	ctx := context.Background()
	helper := NewRawQueryHelper(db)

	// Query from a specific schema
	var id int
	var name string
	row := helper.WithSchema("public").QueryRow(ctx, "SELECT id, name FROM users WHERE id = $1", 1)
	if err := row.Scan(&id, &name); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	fmt.Printf("User: ID=%d, Name=%s\n", id, name)
	return nil
}

// Example2_Transaction demonstrates transaction usage
func Example2_Transaction(db *sql.DB) error {
	ctx := context.Background()
	helper := NewRawQueryHelper(db)

	// Execute multiple operations in a transaction
	err := helper.WithSchema("public").Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Insert user
		result, err := tx.ExecContext(ctx,
			"INSERT INTO users (name, email) VALUES ($1, $2)",
			"Jane Doe", "jane@example.com")
		if err != nil {
			return err
		}

		lastID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		// Insert user profile
		_, err = tx.ExecContext(ctx,
			"INSERT INTO user_profiles (user_id, bio) VALUES ($1, $2)",
			lastID, "Software Engineer")
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	fmt.Println("Transaction completed successfully")
	return nil
}

// Example3_CrossSchemaQuery demonstrates querying across multiple schemas
func Example3_CrossSchemaQuery(db *sql.DB) error {
	ctx := context.Background()
	helper := NewRawQueryHelper(db)

	// Query data from different schemas in the same query
	query := `
		SELECT 
			o.id as org_id,
			o.name as org_name,
			COUNT(a.id) as analytics_count
		FROM public.organizations o
		LEFT JOIN analytics.organization_analytics a ON o.id = a.organization_id
		WHERE o.status = $1
		GROUP BY o.id, o.name
	`

	rows, err := helper.WithSchema("").QueryDirect(ctx, query, "active")
	if err != nil {
		return fmt.Errorf("failed to execute cross-schema query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var orgID int
		var orgName string
		var analyticsCount int
		if err := rows.Scan(&orgID, &orgName, &analyticsCount); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		fmt.Printf("Org: %s (ID: %d) - Analytics: %d\n", orgName, orgID, analyticsCount)
	}

	return nil
}

// Example4_BatchExecute demonstrates batch query execution
func Example4_BatchExecute(db *sql.DB) error {
	ctx := context.Background()
	helper := NewRawQueryHelper(db)

	// Execute multiple queries in a single transaction
	queries := []BatchQuery{
		{
			Query: "INSERT INTO users (name, email) VALUES ($1, $2)",
			Args:  []interface{}{"User1", "user1@example.com"},
		},
		{
			Query: "INSERT INTO users (name, email) VALUES ($1, $2)",
			Args:  []interface{}{"User2", "user2@example.com"},
		},
		{
			Query: "INSERT INTO users (name, email) VALUES ($1, $2)",
			Args:  []interface{}{"User3", "user3@example.com"},
		},
	}

	if err := helper.WithSchema("public").ExecuteBatch(ctx, queries); err != nil {
		return fmt.Errorf("batch execution failed: %w", err)
	}

	fmt.Println("Batch execution completed successfully")
	return nil
}

// Example5_GetRawDB demonstrates getting the raw database connection
func Example5_GetRawDB(db *sql.DB) error {
	helper := NewRawQueryHelper(db)

	// Get the raw *sql.DB for advanced operations
	rawDB := helper.WithSchema("public").GetRawDB()

	// Set connection pool settings
	rawDB.SetMaxOpenConns(25)
	rawDB.SetMaxIdleConns(5)
	rawDB.SetConnMaxLifetime(5 * time.Minute)

	// Ping database
	if err := rawDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	fmt.Println("Database connection is healthy")
	return nil
}
