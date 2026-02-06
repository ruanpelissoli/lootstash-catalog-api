package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/spf13/cobra"
)

var checkDupesCmd = &cobra.Command{
	Use:   "check-dupes",
	Short: "Show duplicate unique item names",
	RunE:  runCheckDupes,
}

func init() {
	rootCmd.AddCommand(checkDupesCmd)
}

func runCheckDupes(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Pool().Query(ctx, `
		SELECT name, COUNT(*) as cnt, array_agg(base_code) as bases, array_agg(index_id) as ids
		FROM d2.unique_items
		GROUP BY name
		HAVING COUNT(*) > 1
		ORDER BY cnt DESC, name
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Duplicate Unique Item Names:")
	fmt.Println("============================")

	count := 0
	for rows.Next() {
		var name string
		var cnt int
		var bases, ids []interface{}
		if err := rows.Scan(&name, &cnt, &bases, &ids); err != nil {
			return err
		}
		count++
		fmt.Printf("\n%d. '%s' (count: %d)\n", count, name, cnt)
		fmt.Printf("   Base codes: %v\n", bases)
		fmt.Printf("   Index IDs:  %v\n", ids)
	}

	if count == 0 {
		fmt.Println("No duplicates found!")
	} else {
		fmt.Printf("\nTotal: %d duplicate names\n", count)
	}

	return nil
}
