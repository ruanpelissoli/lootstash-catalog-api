package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify [game]",
	Short: "Verify catalog data integrity for a specific game",
	Long:  `Verify checks for duplicates and data integrity issues in the catalog.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runVerify,
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) error {
	game := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	switch game {
	case "d2":
		return verifyD2(ctx, db)
	default:
		return fmt.Errorf("unknown game: %s", game)
	}
}

func verifyD2(ctx context.Context, db *database.DB) error {
	PrintInfo("Verifying Diablo II catalog data...")

	pool := db.Pool()

	// Check counts
	tables := []struct {
		name  string
		query string
	}{
		{"Item Types", "SELECT COUNT(*) FROM d2.item_types"},
		{"Item Bases", "SELECT COUNT(*) FROM d2.item_bases"},
		{"Unique Items", "SELECT COUNT(*) FROM d2.unique_items"},
		{"Set Bonuses", "SELECT COUNT(*) FROM d2.set_bonuses"},
		{"Set Items", "SELECT COUNT(*) FROM d2.set_items"},
		{"Runewords", "SELECT COUNT(*) FROM d2.runewords"},
		{"Runes", "SELECT COUNT(*) FROM d2.runes"},
		{"Gems", "SELECT COUNT(*) FROM d2.gems"},
		{"Properties", "SELECT COUNT(*) FROM d2.properties"},
		{"Affixes", "SELECT COUNT(*) FROM d2.affixes"},
	}

	fmt.Println("\n=== Record Counts ===")
	for _, t := range tables {
		var count int
		if err := pool.QueryRow(ctx, t.query).Scan(&count); err != nil {
			return fmt.Errorf("failed to count %s: %w", t.name, err)
		}
		fmt.Printf("  %-15s %d\n", t.name+":", count)
	}

	// Check for duplicates
	fmt.Println("\n=== Duplicate Checks ===")

	duplicateChecks := []struct {
		name  string
		query string
	}{
		{"Item Types (code)", `SELECT code, COUNT(*) as cnt FROM d2.item_types GROUP BY code HAVING COUNT(*) > 1`},
		{"Item Bases (code)", `SELECT code, COUNT(*) as cnt FROM d2.item_bases GROUP BY code HAVING COUNT(*) > 1`},
		{"Unique Items (index_id)", `SELECT index_id, COUNT(*) as cnt FROM d2.unique_items GROUP BY index_id HAVING COUNT(*) > 1`},
		{"Unique Items (name)", `SELECT name, COUNT(*) as cnt FROM d2.unique_items GROUP BY name HAVING COUNT(*) > 1`},
		{"Set Bonuses (name)", `SELECT name, COUNT(*) as cnt FROM d2.set_bonuses GROUP BY name HAVING COUNT(*) > 1`},
		{"Set Items (index_id)", `SELECT index_id, COUNT(*) as cnt FROM d2.set_items GROUP BY index_id HAVING COUNT(*) > 1`},
		{"Runewords (name)", `SELECT name, COUNT(*) as cnt FROM d2.runewords GROUP BY name HAVING COUNT(*) > 1`},
		{"Runes (code)", `SELECT code, COUNT(*) as cnt FROM d2.runes GROUP BY code HAVING COUNT(*) > 1`},
		{"Gems (code)", `SELECT code, COUNT(*) as cnt FROM d2.gems GROUP BY code HAVING COUNT(*) > 1`},
		{"Properties (code)", `SELECT code, COUNT(*) as cnt FROM d2.properties GROUP BY code HAVING COUNT(*) > 1`},
		{"Affixes (name+type)", `SELECT name, affix_type, COUNT(*) as cnt FROM d2.affixes GROUP BY name, affix_type HAVING COUNT(*) > 1`},
	}

	allGood := true
	for _, check := range duplicateChecks {
		rows, err := pool.Query(ctx, check.query)
		if err != nil {
			return fmt.Errorf("failed to check %s: %w", check.name, err)
		}

		hasDupes := false
		for rows.Next() {
			hasDupes = true
			allGood = false
		}
		rows.Close()

		if hasDupes {
			PrintError(fmt.Sprintf("%s: DUPLICATES FOUND", check.name))
		} else {
			PrintSuccess(fmt.Sprintf("%s: OK", check.name))
		}
	}

	// Sample data check
	fmt.Println("\n=== Sample Data ===")

	// Sample unique items
	var uniqueName, uniqueBase string
	var uniqueProps int
	err := pool.QueryRow(ctx, `
		SELECT name, base_code, jsonb_array_length(properties)
		FROM d2.unique_items
		WHERE name = 'The Gnasher'`).Scan(&uniqueName, &uniqueBase, &uniqueProps)
	if err != nil {
		PrintError("The Gnasher not found")
	} else {
		fmt.Printf("  Unique 'The Gnasher': base=%s, props=%d\n", uniqueBase, uniqueProps)
	}

	// Sample runeword
	var rwName string
	var rwRunes, rwProps int
	err = pool.QueryRow(ctx, `
		SELECT display_name, jsonb_array_length(runes), jsonb_array_length(properties)
		FROM d2.runewords
		WHERE display_name = 'Enigma'`).Scan(&rwName, &rwRunes, &rwProps)
	if err != nil {
		PrintError("Enigma runeword not found")
	} else {
		fmt.Printf("  Runeword 'Enigma': runes=%d, props=%d\n", rwRunes, rwProps)
	}

	// Sample set
	var setName string
	var setPartial, setFull int
	err = pool.QueryRow(ctx, `
		SELECT name, jsonb_array_length(partial_bonuses), jsonb_array_length(full_bonuses)
		FROM d2.set_bonuses
		WHERE name = 'Tal Rasha''s Wrappings'`).Scan(&setName, &setPartial, &setFull)
	if err != nil {
		PrintError("Tal Rasha's Wrappings set not found")
	} else {
		fmt.Printf("  Set 'Tal Rasha's Wrappings': partial=%d, full=%d bonuses\n", setPartial, setFull)
	}

	if allGood {
		fmt.Println("\n" + "=== VERIFICATION PASSED ===")
	} else {
		fmt.Println("\n" + "=== VERIFICATION FAILED - Duplicates found ===")
	}

	return nil
}
