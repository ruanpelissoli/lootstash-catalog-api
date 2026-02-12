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
	Short: "Remove duplicate items from the database",
	Long: `Removes duplicates from the database:
- Runes and gems that appear in both item_bases and their dedicated tables
- Unique items, set items, and base items with duplicate names (keeps lowest index_id)

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

	// Find duplicate unique items (same name, different index_id — keep lowest)
	PrintInfo("Finding duplicate unique items...")
	dupeUniqueRows, err := pool.Query(ctx, `
		SELECT u.index_id, u.name
		FROM d2.unique_items u
		WHERE u.index_id NOT IN (
			SELECT MIN(index_id) FROM d2.unique_items GROUP BY LOWER(name)
		)
		ORDER BY u.name
	`)
	if err != nil {
		return fmt.Errorf("failed to query unique duplicates: %w", err)
	}

	type dupeItem struct {
		id   int
		name string
	}
	var dupeUniques []dupeItem
	for dupeUniqueRows.Next() {
		var d dupeItem
		if err := dupeUniqueRows.Scan(&d.id, &d.name); err != nil {
			dupeUniqueRows.Close()
			return err
		}
		dupeUniques = append(dupeUniques, d)
		fmt.Printf("  Found duplicate unique: %s (index_id: %d)\n", d.name, d.id)
	}
	dupeUniqueRows.Close()

	// Find duplicate set items (same name, different index_id — keep lowest)
	PrintInfo("Finding duplicate set items...")
	dupeSetRows, err := pool.Query(ctx, `
		SELECT s.index_id, s.name
		FROM d2.set_items s
		WHERE s.index_id NOT IN (
			SELECT MIN(index_id) FROM d2.set_items GROUP BY LOWER(name)
		)
		ORDER BY s.name
	`)
	if err != nil {
		return fmt.Errorf("failed to query set item duplicates: %w", err)
	}

	var dupeSetItems []dupeItem
	for dupeSetRows.Next() {
		var d dupeItem
		if err := dupeSetRows.Scan(&d.id, &d.name); err != nil {
			dupeSetRows.Close()
			return err
		}
		dupeSetItems = append(dupeSetItems, d)
		fmt.Printf("  Found duplicate set item: %s (index_id: %d)\n", d.name, d.id)
	}
	dupeSetRows.Close()

	// Find duplicate base items (same name, different id — keep lowest)
	PrintInfo("Finding duplicate base items...")
	dupeBaseRows, err := pool.Query(ctx, `
		SELECT b.id, b.name
		FROM d2.item_bases b
		WHERE b.id NOT IN (
			SELECT MIN(id) FROM d2.item_bases GROUP BY LOWER(name)
		)
		ORDER BY b.name
	`)
	if err != nil {
		return fmt.Errorf("failed to query base item duplicates: %w", err)
	}

	var dupeBases []dupeItem
	for dupeBaseRows.Next() {
		var d dupeItem
		if err := dupeBaseRows.Scan(&d.id, &d.name); err != nil {
			dupeBaseRows.Close()
			return err
		}
		dupeBases = append(dupeBases, d)
		fmt.Printf("  Found duplicate base: %s (id: %d)\n", d.name, d.id)
	}
	dupeBaseRows.Close()

	fmt.Printf("\nDuplicates found: %d runes, %d gems, %d uniques, %d set items, %d bases\n",
		len(runeCodes), len(gemCodes), len(dupeUniques), len(dupeSetItems), len(dupeBases))

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

	if len(dupeUniques) > 0 {
		PrintInfo("Deleting duplicate unique items...")
		for _, d := range dupeUniques {
			_, err := pool.Exec(ctx, `DELETE FROM d2.unique_items WHERE index_id = $1`, d.id)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete unique %s (id %d): %v", d.name, d.id, err))
			}
		}
		PrintSuccess(fmt.Sprintf("Deleted %d duplicate uniques", len(dupeUniques)))
	}

	if len(dupeSetItems) > 0 {
		PrintInfo("Deleting duplicate set items...")
		for _, d := range dupeSetItems {
			_, err := pool.Exec(ctx, `DELETE FROM d2.set_items WHERE index_id = $1`, d.id)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete set item %s (id %d): %v", d.name, d.id, err))
			}
		}
		PrintSuccess(fmt.Sprintf("Deleted %d duplicate set items", len(dupeSetItems)))
	}

	if len(dupeBases) > 0 {
		PrintInfo("Deleting duplicate base items...")
		for _, d := range dupeBases {
			_, err := pool.Exec(ctx, `DELETE FROM d2.item_bases WHERE id = $1`, d.id)
			if err != nil {
				PrintError(fmt.Sprintf("Failed to delete base %s (id %d): %v", d.name, d.id, err))
			}
		}
		PrintSuccess(fmt.Sprintf("Deleted %d duplicate bases", len(dupeBases)))
	}

	PrintSuccess("Cleanup completed!")
	return nil
}
