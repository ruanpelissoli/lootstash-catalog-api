package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api"
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

	// Create repository
	repo := d2.NewRepository(db.Pool())

	// Parse rate limit config
	rateLimitMax := 100
	if v := getEnvOrDefault("RATE_LIMIT_MAX", ""); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			rateLimitMax = parsed
		}
	}
	rateLimitWindow := 1 * time.Minute
	if v := getEnvOrDefault("RATE_LIMIT_WINDOW", ""); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil && parsed > 0 {
			rateLimitWindow = parsed
		}
	}

	// Create server config
	supabaseURL := getEnvOrDefault("SUPABASE_URL", "")
	config := &api.Config{
		Port:            port,
		AllowedOrigins:  allowedOrigins,
		JWTSecret:       getEnvOrDefault("SUPABASE_JWT_SECRET", ""),
		JWKSURL:         supabaseURL + "/auth/v1/.well-known/jwks.json",
		JWTAudience:     "authenticated",
		JWTIssuer:       supabaseURL + "/auth/v1",
		AuthDebug:       getEnvOrDefault("AUTH_DEBUG", "") == "true",
		RedisURL:        GetRedisURL(),
		RateLimitMax:    rateLimitMax,
		RateLimitWindow: rateLimitWindow,
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
	if GetRedisURL() != "" {
		PrintInfo(fmt.Sprintf("Rate limit: %d requests per %s (Redis: %s)", rateLimitMax, rateLimitWindow, GetRedisURL()))
	} else {
		PrintInfo("Rate limiter: disabled (no REDIS_URL)")
	}
	PrintInfo("Press Ctrl+C to stop")

	if err := server.Start(); err != nil {
		PrintError(fmt.Sprintf("Server error: %v", err))
		return err
	}

	return nil
}
