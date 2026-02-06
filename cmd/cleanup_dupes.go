package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/spf13/cobra"
)

var cleanupDryRun bool

var cleanupDupesCmd = &cobra.Command{
	Use:   "cleanup-dupes",
	Short: "Remove duplicate runes and gems from item_bases table",
	Long: `Removes runes and gems from the item_bases table since they have
their own dedicated tables (d2.runes and d2.gems).

This prevents duplicate search results.

Examples:
  # Preview what would be deleted
  lootstash-catalog cleanup-dupes --dry-run

  # Actually delete duplicates
  lootstash-catalog cleanup-dupes`,
	RunE: runCleanupDupes,
}

func init() {
	rootCmd.AddCommand(cleanupDupesCmd)
	cleanupDupesCmd.Flags().BoolVar(&cleanupDryRun, "dry-run", false, "Preview without deleting")
}

func runCleanupDupes(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Connect to database
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	pool := db.Pool()

	// Find runes in item_bases
	PrintInfo("Finding duplicate runes in item_bases...")
	runeRows, err := pool.Query(ctx, `
		SELECT ib.code, ib.name
		FROM d2.item_bases ib
		INNER JOIN d2.runes r ON LOWER(ib.name) = LOWER(r.name)
	`)
	if err != nil {
		return fmt.Errorf("failed to query rune duplicates: %w", err)
	}

	var runeCodes []string
	for runeRows.Next() {
		var code, name string
		if err := runeRows.Scan(&code, &name); err != nil {
			runeRows.Close()
			return err
		}
		runeCodes = append(runeCodes, code)
		fmt.Printf("  Found duplicate rune: %s (code: %s)\n", name, code)
	}
	runeRows.Close()

	// Find gems in item_bases
	PrintInfo("Finding duplicate gems in item_bases...")
	gemRows, err := pool.Query(ctx, `
		SELECT ib.code, ib.name
		FROM d2.item_bases ib
		INNER JOIN d2.gems g ON LOWER(ib.name) = LOWER(g.name)
	`)
	if err != nil {
		return fmt.Errorf("failed to query gem duplicates: %w", err)
	}

	var gemCodes []string
	for gemRows.Next() {
		var code, name string
		if err := gemRows.Scan(&code, &name); err != nil {
			gemRows.Close()
			return err
		}
		gemCodes = append(gemCodes, code)
		fmt.Printf("  Found duplicate gem: %s (code: %s)\n", name, code)
	}
	gemRows.Close()

	fmt.Printf("\nFound %d duplicate runes and %d duplicate gems\n", len(runeCodes), len(gemCodes))

	if cleanupDryRun {
		PrintInfo("Dry run - no changes made")
		return nil
	}

	// Delete duplicates
	if len(runeCodes) > 0 {
		PrintInfo("Deleting duplicate runes from item_bases...")
		for _, code := range runeCodes {
			_, err := pool.Exec(ctx, `DELETE FROM d2.item_bases WHERE code = $1`, code)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete rune %s: %v", code, err))
			}
		}
		PrintSuccess(fmt.Sprintf("Deleted %d duplicate runes", len(runeCodes)))
	}

	if len(gemCodes) > 0 {
		PrintInfo("Deleting duplicate gems from item_bases...")
		for _, code := range gemCodes {
			_, err := pool.Exec(ctx, `DELETE FROM d2.item_bases WHERE code = $1`, code)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete gem %s: %v", code, err))
			}
		}
		PrintSuccess(fmt.Sprintf("Deleted %d duplicate gems", len(gemCodes)))
	}

	PrintSuccess("Cleanup completed!")
	return nil
}
