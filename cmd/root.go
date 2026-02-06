package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	databaseURL string
	redisURL    string
)

var rootCmd = &cobra.Command{
	Use:   "lootstash-catalog",
	Short: "LootStash Catalog API - Game catalog data management",
	Long: `LootStash Catalog API manages static game data for various ARPG games.
It parses game data files and stores them in PostgreSQL with Redis caching.

Supported games:
  - d2: Diablo II: Resurrected`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&databaseURL, "database-url", getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:54322/postgres"), "PostgreSQL connection string")
	rootCmd.PersistentFlags().StringVar(&redisURL, "redis-url", getEnvOrDefault("REDIS_URL", "localhost:6379"), "Redis connection string")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetDatabaseURL() string {
	return databaseURL
}

func GetRedisURL() string {
	return redisURL
}

func PrintSuccess(msg string) {
	fmt.Printf("✓ %s\n", msg)
}

func PrintError(msg string) {
	fmt.Printf("✗ %s\n", msg)
}

func PrintInfo(msg string) {
	fmt.Printf("→ %s\n", msg)
}
