package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/cache"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [game]",
	Short: "Import catalog data for a specific game",
	Long: `Import parses game data files and stores them in the catalog database.

Available games:
  d2 - Diablo II: Resurrected`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	game := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Initialize database connection
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Initialize Redis cache
	PrintInfo("Connecting to Redis...")
	redisCache, err := cache.NewRedisCache(ctx, GetRedisURL())
	if err != nil {
		PrintError(fmt.Sprintf("Failed to connect to Redis: %v (continuing without cache)", err))
		redisCache = nil
	} else {
		PrintSuccess("Connected to Redis")
		defer redisCache.Close()
	}

	// Run game-specific import
	switch game {
	case "d2":
		return importD2(ctx, db, redisCache)
	default:
		return fmt.Errorf("unknown game: %s. Available games: d2", game)
	}
}

func importD2(ctx context.Context, db *database.DB, redisCache *cache.RedisCache) error {
	PrintInfo("Starting Diablo II catalog import...")

	importer := d2.NewImporter(db, redisCache)

	stats, err := importer.Import(ctx, "catalogs/d2")
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	PrintSuccess("Diablo II catalog import completed!")
	fmt.Println("\nImport Statistics:")
	fmt.Printf("  Item Types:       %d imported, %d skipped\n", stats.ItemTypes.Imported, stats.ItemTypes.Skipped)
	fmt.Printf("  Item Bases:       %d imported, %d skipped\n", stats.ItemBases.Imported, stats.ItemBases.Skipped)
	fmt.Printf("  Unique Items:     %d imported, %d skipped\n", stats.UniqueItems.Imported, stats.UniqueItems.Skipped)
	fmt.Printf("  Set Bonuses:      %d imported, %d skipped\n", stats.SetBonuses.Imported, stats.SetBonuses.Skipped)
	fmt.Printf("  Set Items:        %d imported, %d skipped\n", stats.SetItems.Imported, stats.SetItems.Skipped)
	fmt.Printf("  Runewords:        %d imported, %d skipped\n", stats.Runewords.Imported, stats.Runewords.Skipped)
	fmt.Printf("  Runes:            %d imported, %d skipped\n", stats.Runes.Imported, stats.Runes.Skipped)
	fmt.Printf("  Gems:             %d imported, %d skipped\n", stats.Gems.Imported, stats.Gems.Skipped)
	fmt.Printf("  Properties:       %d imported, %d skipped\n", stats.Properties.Imported, stats.Properties.Skipped)
	fmt.Printf("  Affixes:          %d imported, %d skipped\n", stats.Affixes.Imported, stats.Affixes.Skipped)
	fmt.Printf("  Treasure Classes: %d imported, %d skipped\n", stats.TreasureClasses.Imported, stats.TreasureClasses.Skipped)
	fmt.Printf("  Item Ratios:      %d imported, %d skipped\n", stats.ItemRatios.Imported, stats.ItemRatios.Skipped)
	fmt.Printf("  Runeword Bases:   %d imported, %d skipped\n", stats.RunewordBases.Imported, stats.RunewordBases.Skipped)

	if len(stats.MissingStatCodes) > 0 {
		fmt.Println("\n⚠ Missing stat codes (not in FilterableStats):")
		for _, code := range stats.MissingStatCodes {
			fmt.Printf("  - %s\n", code)
		}
		fmt.Println("  Add these to internal/games/d2/statcodes.go to enable filtering/item creation.")
	}

	return nil
}
