package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/handlers"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// Server represents the HTTP server
type Server struct {
	app    *fiber.App
	repo   *d2.Repository
	config *Config
}

// Config holds server configuration
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	AllowedOrigins  string
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		AllowedOrigins:  "*",
	}
}

// NewServer creates a new HTTP server
func NewServer(repo *d2.Repository, config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		AppName:      "LootStash Catalog API",
	})

	server := &Server{
		app:    app,
		repo:   repo,
		config: config,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.app.Use(recover.New())

	// Logger middleware
	s.app.Use(logger.New(logger.Config{
		Format:     "${time} ${status} ${method} ${path} ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// CORS middleware
	s.app.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))
}

func (s *Server) setupRoutes() {
	// Health check
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "lootstash-catalog-api",
		})
	})

	// API v1 group
	api := s.app.Group("/api")
	v1 := api.Group("/v1")

	// D2 routes
	d2Routes := v1.Group("/d2")
	s.setupD2Routes(d2Routes)
}

func (s *Server) setupD2Routes(router fiber.Router) {
	itemHandler := handlers.NewItemHandler(s.repo)

	// Item routes
	items := router.Group("/items")

	// Search endpoint
	items.Get("/search", itemHandler.Search)

	// Generic item lookup by type and ID
	items.Get("/:type/:id", itemHandler.GetItem)

	// Specific type endpoints (for convenience)
	items.Get("/unique/:id", itemHandler.GetUniqueItem)
	items.Get("/set/:id", itemHandler.GetSetItem)
	items.Get("/runeword/:id", itemHandler.GetRuneword)
	items.Get("/runeword/:id/bases", itemHandler.GetRunewordBases)
	items.Get("/rune/:id", itemHandler.GetRune)
	items.Get("/gem/:id", itemHandler.GetGem)
	items.Get("/base/:id", itemHandler.GetBase)

	// Collection endpoints - list all items by type
	router.Get("/runes", itemHandler.GetAllRunes)
	router.Get("/gems", itemHandler.GetAllGems)
	router.Get("/bases", itemHandler.GetAllBases)
	router.Get("/uniques", itemHandler.GetAllUniques)
	router.Get("/sets", itemHandler.GetAllSets)
	router.Get("/runewords", itemHandler.GetAllRunewords)

	// Reference data endpoints - for marketplace filtering
	router.Get("/stats", itemHandler.GetAllStats)
	router.Get("/categories", itemHandler.GetAllCategories)
	router.Get("/rarities", itemHandler.GetAllRarities)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("Starting LootStash Catalog API on %s\n", addr)
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
