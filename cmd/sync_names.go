package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/spf13/cobra"
)

var syncNamesDryRun bool

var syncNamesCmd = &cobra.Command{
	Use:   "sync-names",
	Short: "Update item names from HTML files (community names)",
	Long: `Updates item names in the database using the display names from diablo2.io HTML files.

The game data files use internal names like "Fathom" but the community uses
names like "Death's Fathom". This command syncs the display names.

Examples:
  # Preview what would be updated
  lootstash-catalog sync-names --dry-run

  # Actually update names
  lootstash-catalog sync-names`,
	RunE: runSyncNames,
}

func init() {
	rootCmd.AddCommand(syncNamesCmd)
	syncNamesCmd.Flags().BoolVar(&syncNamesDryRun, "dry-run", false, "Preview without updating")
}

func runSyncNames(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
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
	parser := d2.NewHTMLParser()
	pagesPath := "catalogs/d2/pages"

	// Parse HTML files
	PrintInfo("Parsing HTML files...")

	// Build HTML name lookup (normalized -> display name)
	htmlNames := make(map[string]string)

	// Parse uniques
	if items, err := parser.ParseFile(pagesPath + "/uniques.html"); err == nil {
		for _, item := range items {
			key := normalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d unique items from HTML\n", len(items))
	}

	// Parse sets
	if items, err := parser.ParseFile(pagesPath + "/sets.html"); err == nil {
		for _, item := range items {
			key := normalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d set items from HTML\n", len(items))
	}

	// Parse base items
	if items, err := parser.ParseFile(pagesPath + "/base.html"); err == nil {
		for _, item := range items {
			key := normalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d base items from HTML\n", len(items))
	}

	// Parse misc (runes, gems)
	if items, err := parser.ParseFile(pagesPath + "/misc.html"); err == nil {
		for _, item := range items {
			key := normalizeForLookup(item.Name)
			htmlNames[key] = item.Name
		}
		fmt.Printf("  Parsed %d misc items from HTML\n", len(items))
	}

	var updated, notFound int

	// Update unique items
	PrintInfo("\nChecking unique items...")
	rows, err := pool.Query(ctx, `SELECT id, name FROM d2.unique_items ORDER BY id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var dbName string
		rows.Scan(&id, &dbName)

		htmlName := findBestMatch(dbName, htmlNames)
		if htmlName != "" && htmlName != dbName {
			if syncNamesDryRun {
				fmt.Printf("  Would update: '%s' -> '%s'\n", dbName, htmlName)
			} else {
				_, err := pool.Exec(ctx, `UPDATE d2.unique_items SET name = $1 WHERE id = $2`, htmlName, id)
				if err != nil {
					PrintError(fmt.Sprintf("Failed to update %s: %v", dbName, err))
				} else {
					fmt.Printf("  Updated: '%s' -> '%s'\n", dbName, htmlName)
				}
			}
			updated++
		} else if htmlName == "" {
			notFound++
		}
	}
	rows.Close()

	// Update set items
	PrintInfo("\nChecking set items...")
	rows, err = pool.Query(ctx, `SELECT id, name FROM d2.set_items ORDER BY id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var dbName string
		rows.Scan(&id, &dbName)

		htmlName := findBestMatch(dbName, htmlNames)
		if htmlName != "" && htmlName != dbName {
			if syncNamesDryRun {
				fmt.Printf("  Would update: '%s' -> '%s'\n", dbName, htmlName)
			} else {
				_, err := pool.Exec(ctx, `UPDATE d2.set_items SET name = $1 WHERE id = $2`, htmlName, id)
				if err != nil {
					PrintError(fmt.Sprintf("Failed to update %s: %v", dbName, err))
				} else {
					fmt.Printf("  Updated: '%s' -> '%s'\n", dbName, htmlName)
				}
			}
			updated++
		} else if htmlName == "" {
			notFound++
		}
	}
	rows.Close()

	// Update runes
	PrintInfo("\nChecking runes...")
	rows, err = pool.Query(ctx, `SELECT id, name FROM d2.runes ORDER BY id`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int
		var dbName string
		rows.Scan(&id, &dbName)

		// For runes, try both the name and name without "Rune" suffix
		htmlName := findBestMatch(dbName, htmlNames)
		if htmlName == "" {
			// Try just the rune name without " Rune" suffix
			shortName := strings.TrimSuffix(dbName, " Rune")
			if found, ok := htmlNames[normalizeForLookup(shortName)]; ok {
				// Update to add " Rune" suffix if HTML just has "Ber" etc.
				htmlName = found + " Rune"
			}
		}

		if htmlName != "" && htmlName != dbName {
			if syncNamesDryRun {
				fmt.Printf("  Would update: '%s' -> '%s'\n", dbName, htmlName)
			} else {
				_, err := pool.Exec(ctx, `UPDATE d2.runes SET name = $1 WHERE id = $2`, htmlName, id)
				if err != nil {
					PrintError(fmt.Sprintf("Failed to update %s: %v", dbName, err))
				} else {
					fmt.Printf("  Updated: '%s' -> '%s'\n", dbName, htmlName)
				}
			}
			updated++
		} else if htmlName == "" {
			notFound++
		}
	}
	rows.Close()

	fmt.Printf("\nSummary: %d items would be updated, %d items not found in HTML\n", updated, notFound)

	if syncNamesDryRun {
		PrintInfo("Dry run - no changes made")
	} else {
		PrintSuccess("Sync completed!")
	}

	return nil
}

func normalizeForLookup(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

func findBestMatch(dbName string, htmlNames map[string]string) string {
	// Try exact match
	key := normalizeForLookup(dbName)
	if name, ok := htmlNames[key]; ok {
		return name
	}

	// Try partial match - if DB name is contained in HTML name
	for htmlKey, htmlName := range htmlNames {
		if strings.Contains(htmlKey, key) || strings.Contains(key, htmlKey) {
			if len(key) > 3 && len(htmlKey) > 3 {
				return htmlName
			}
		}
	}

	return ""
}
