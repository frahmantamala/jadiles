package cmd

import (
	"context"
	"database/sql"
	"log"

	"github.com/frahmantamala/jadiles/internal/database"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

var (
	migrateCmd = &cobra.Command{
		RunE:  runMigration,
		Use:   "migrate",
		Short: "Run database migration files under migrations directory",
	}
	migrateRollback bool
	migrateDir      string
)

func init() {
	migrateCmd.Flags().BoolVarP(&migrateRollback, "rollback", "r", false, "Rollback the latest version of migration")
	migrateCmd.PersistentFlags().StringVarP(&migrateDir, "dir", "d", "db/migrations", "SQL migrations directory")
}

func runMigration(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	cfg, err := loadConfig(".")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	source := database.BuildPostgresSource(cfg.Database)

	db, err := sql.Open("postgres", source)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Set custom migration table name
	goose.SetTableName("schema_migrations")

	if migrateRollback {
		log.Println("Running migration rollback...")
		if err := goose.RunContext(ctx, "down", db, migrateDir); err != nil {
			log.Fatalf("Migration rollback failed: %v", err)
		}
		log.Println("Migration rollback completed successfully")
		return nil
	}

	log.Println("Running migrations...")
	if err := goose.RunContext(ctx, "up", db, migrateDir); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
