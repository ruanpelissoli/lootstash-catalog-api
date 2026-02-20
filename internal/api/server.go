package api

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberredis "github.com/gofiber/storage/redis/v3"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/handlers"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/middleware"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/api/services"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/cache"
	"github.com/ruanpelissoli/lootstash-catalog-api/internal/games/d2"
)

// Server represents the HTTP server
type Server struct {
	app     *fiber.App
	repo    *d2.Repository
	service *services.CatalogService
	config  *Config
}

// Config holds server configuration
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	AllowedOrigins  string
	JWTSecret       string // HMAC (HS256) - legacy/testing
	JWKSURL         string // ECDSA (ES256) - Supabase JWKS endpoint
	JWTAudience     string // Expected "aud" claim
	JWTIssuer       string // Expected "iss" claim
	AuthDebug       bool   // Debug logging for auth
	RedisURL        string
	RateLimitMax    int
	RateLimitWindow time.Duration
}

// DefaultConfig returns default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		AllowedOrigins:  "*",
		RateLimitMax:    100,
		RateLimitWindow: 1 * time.Minute,
	}
}

// NewServer creates a new HTTP server. redisCache may be nil to disable caching.
func NewServer(repo *d2.Repository, config *Config, redisCache *cache.RedisCache) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	app := fiber.New(fiber.Config{
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		AppName:      "LootStash Catalog API",
		ProxyHeader:  "Fly-Client-IP",
	})

	server := &Server{
		app:     app,
		repo:    repo,
		service: services.NewCatalogService(repo, redisCache),
		config:  config,
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

	// Rate limiter middleware
	if s.config.RedisURL != "" {
		storeCfg := parseRedisURL(s.config.RedisURL)

		store := fiberredis.New(storeCfg)

		s.app.Use(limiter.New(limiter.Config{
			Max:        s.config.RateLimitMax,
			Expiration: s.config.RateLimitWindow,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error":   "too_many_requests",
					"message": "Rate limit exceeded. Please try again later.",
				})
			},
			Storage: store,
		}))
	}

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
	d2Routes.Use(middleware.CacheHeaders())
	s.setupD2Routes(d2Routes)

	// Admin routes
	adminRoutes := v1.Group("/admin/d2")
	adminRoutes.Use(middleware.NoCacheHeaders())
	s.setupAdminRoutes(adminRoutes)
}

func (s *Server) setupD2Routes(router fiber.Router) {
	itemHandler := handlers.NewItemHandler(s.service)

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
	items.Get("/quest/:id", itemHandler.GetQuestItem)

	// Collection endpoints - list all items by type
	router.Get("/runes", itemHandler.GetAllRunes)
	router.Get("/gems", itemHandler.GetAllGems)
	router.Get("/bases", itemHandler.GetAllBases)
	router.Get("/uniques", itemHandler.GetAllUniques)
	router.Get("/sets", itemHandler.GetAllSets)
	router.Get("/runewords", itemHandler.GetAllRunewords)
	router.Get("/quests", itemHandler.GetAllQuestItems)
	router.Get("/classes", itemHandler.GetAllClasses)

	// Reference data endpoints - for marketplace filtering
	router.Get("/stats", itemHandler.GetAllStats)
	router.Get("/categories", itemHandler.GetAllCategories)
	router.Get("/rarities", itemHandler.GetAllRarities)
}

func (s *Server) setupAdminRoutes(router fiber.Router) {
	authConfig := middleware.AuthConfig{
		JWTSecret: s.config.JWTSecret,
		JWKSURL:   s.config.JWKSURL,
		Audience:  s.config.JWTAudience,
		Issuer:    s.config.JWTIssuer,
		Debug:     s.config.AuthDebug,
	}
	router.Use(middleware.NewAuthMiddleware(authConfig))
	router.Use(middleware.AdminMiddleware(s.repo))

	adminHandler := handlers.NewAdminHandler(s.repo)

	router.Post("/classes", adminHandler.CreateClass)
	router.Put("/classes/:classId", adminHandler.UpdateClass)

	items := router.Group("/items")
	items.Post("/:type", adminHandler.CreateItem)
	items.Put("/:type/:id", adminHandler.UpdateItem)
	items.Delete("/:type/:id", adminHandler.DeleteItem)
}

// parseRedisURL parses a Redis URL into a fiberredis.Config.
// Handles redis:// and rediss:// (TLS) schemes, as well as bare host:port.
func parseRedisURL(rawURL string) fiberredis.Config {
	cfg := fiberredis.Config{
		Host: "127.0.0.1",
		Port: 6379,
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		// Bare host:port (e.g. "localhost:6379") — split so fiberredis
		// doesn't produce "localhost:6379:6379".
		if host, portStr, splitErr := net.SplitHostPort(rawURL); splitErr == nil {
			cfg.Host = host
			if p, convErr := strconv.Atoi(portStr); convErr == nil {
				cfg.Port = p
			}
		} else {
			cfg.Host = rawURL
		}
		return cfg
	}

	hostname := parsed.Hostname()
	if hostname != "" {
		cfg.Host = hostname
	}

	if p := parsed.Port(); p != "" {
		if port, err := strconv.Atoi(p); err == nil {
			cfg.Port = port
		}
	}

	if parsed.User != nil {
		cfg.Username = parsed.User.Username()
		if pw, ok := parsed.User.Password(); ok {
			cfg.Password = pw
		}
	}

	// Parse database number from path (e.g. /0)
	if len(parsed.Path) > 1 {
		if db, err := strconv.Atoi(parsed.Path[1:]); err == nil {
			cfg.Database = db
		}
	}

	// Enable TLS for rediss:// scheme
	if parsed.Scheme == "rediss" {
		cfg.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	return cfg
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
