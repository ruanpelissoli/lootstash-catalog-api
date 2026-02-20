package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// CacheHeaders returns a middleware that sets Cache-Control headers on successful GET responses.
// - max-age=300: browsers cache for 5 minutes
// - s-maxage=3600: CDN/proxy caches for 1 hour
// - stale-while-revalidate=86400: serve stale while revalidating for up to 24h
func CacheHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if c.Method() == fiber.MethodGet && c.Response().StatusCode() >= 200 && c.Response().StatusCode() < 300 {
			c.Set("Cache-Control", "public, max-age=300, s-maxage=3600, stale-while-revalidate=86400")
		}

		return err
	}
}

// NoCacheHeaders returns a middleware that sets Cache-Control: no-store on all responses.
func NoCacheHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		c.Set("Cache-Control", "no-store")
		return err
	}
}
