package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/cache"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/database"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
	"github.com/spf13/cobra"
)

var (
	port           int
	allowedOrigins string
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	Long: `Start the LootStash Catalog API HTTP server.

The server exposes REST endpoints for querying game catalog data.

Endpoints:
  GET /health                    - Health check
  GET /api/v1/d2/items/search    - Search items by name
  GET /api/v1/d2/items/:type/:id - Get item details
  GET /api/v1/d2/runes           - List all runes

Examples:
  # Start with default settings (port 8080)
  lootstash-catalog serve

  # Start on custom port
  lootstash-catalog serve --port 3002

  # Allow specific origins
  lootstash-catalog serve --allowed-origins "http://localhost:3001"`,
	RunE: runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().IntVar(&port, "port", 8080, "Port to listen on")
	serveCmd.Flags().StringVar(&allowedOrigins, "allowed-origins", getEnvOrDefault("ALLOWED_ORIGIN", "*"), "Comma-separated list of allowed CORS origins (use * for all)")
}

func runServe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to database
	PrintInfo("Connecting to database...")
	db, err := database.NewConnection(ctx, GetDatabaseURL())
	if err != nil {
		PrintError(fmt.Sprintf("Failed to connect to database: %v", err))
		return err
	}
	defer db.Close()
	PrintSuccess("Connected to database")

	// Run catalog import on startup
	PrintInfo("Running catalog import...")
	var redisCache *cache.RedisCache
	redisCache, cacheErr := cache.NewRedisCache(ctx, GetRedisURL())
	if cacheErr != nil {
		PrintInfo(fmt.Sprintf("Redis not available: %v (continuing without cache)", cacheErr))
		redisCache = nil
	} else {
		defer redisCache.Close()
	}

	importer := d2.NewImporter(db, redisCache)
	if _, importErr := importer.Import(ctx, "catalogs/d2"); importErr != nil {
		PrintError(fmt.Sprintf("Catalog import failed: %v (server will start with existing data)", importErr))
	} else {
		PrintSuccess("Catalog import completed")
	}

	// Create repository
	repo := d2.NewRepository(db.Pool())

	// Create server config
	config := &api.Config{
		Port:           port,
		AllowedOrigins: allowedOrigins,
	}

	// Create and start server
	server := api.NewServer(repo, config)

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-shutdown
		PrintInfo("Shutting down server...")
		if err := server.Shutdown(); err != nil {
			PrintError(fmt.Sprintf("Error during shutdown: %v", err))
		}
	}()

	PrintSuccess(fmt.Sprintf("Starting server on port %d", port))
	PrintInfo(fmt.Sprintf("Allowed origins: %s", allowedOrigins))
	PrintInfo("Press Ctrl+C to stop")

	if err := server.Start(); err != nil {
		PrintError(fmt.Sprintf("Server error: %v", err))
		return err
	}

	return nil
}
