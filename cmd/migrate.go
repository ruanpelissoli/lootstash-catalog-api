package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [game]",
	Short: "Run database migrations for a specific game's catalog (legacy - use supabase db reset)",
	Long: `DEPRECATED: Migrations are now managed by Supabase CLI.

Use 'supabase db reset' to apply all migrations from supabase/migrations/

Migration files are organized by game schema:
  - 20240109000000_create_d2_catalog.sql  (Diablo II schema)
  - Future games will have their own migration files

This command exists for backwards compatibility or direct PostgreSQL connections.

Available games:
  d2 - Diablo II: Resurrected`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	game := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize database connection
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Run game-specific migrations
	switch game {
	case "d2":
		return migrateD2(ctx, db)
	default:
		return fmt.Errorf("unknown game: %s. Available games: d2", game)
	}
}

func migrateD2(ctx context.Context, db *database.DB) error {
	PrintInfo("Running Diablo II catalog migrations...")

	if err := db.MigrateD2(ctx); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	PrintSuccess("Diablo II catalog migrations completed!")
	return nil
}
