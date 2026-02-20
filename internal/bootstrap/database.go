package bootstrap

import (
	"database/sql"
	"log"

	"agent-service-prototype/internal/config"

	_ "github.com/lib/pq"
)

// ensureSchemaExists creates the schema if it doesn't exist
func ensureSchemaExists(cfg *config.Config) error {
	dbURL := cfg.DatabaseURL()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS public")
	return err
}

// InitDatabase initializes the database connection pool.
// Ent client creation and auto-migration are commented out until Ent code is generated.
func InitDatabase(cfg *config.Config) *sql.DB {
	if err := ensureSchemaExists(cfg); err != nil {
		log.Fatal("Failed to create schema:", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL())
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Ent client + auto-migration commented out — code generation not run yet.
	// drv := entsql.OpenDB(dialect.Postgres, db)
	// client := ent.NewClient(ent.Driver(drv))
	// client.Use(middleware.AuditHook())
	// ctx := context.Background()
	// if err := client.Schema.Create(ctx, migrate.WithForeignKeys(true)); err != nil {
	// 	log.Fatalf("Failed creating schema resources: %v", err)
	// }

	log.Println("✅ Database initialized with shared connection pool")
	return db
}

// CloseDatabase closes the database connection
func CloseDatabase(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
