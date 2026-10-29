package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"backend/pkg/core/api"
	"backend/pkg/core/config"
	cerrors "backend/pkg/core/errors"
)

// RateLimit limits requests per client IP (in memory — fine while the API runs as one instance).
func RateLimit(rule config.RateLimitRule, clientIPHeader string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          rule.Max,
		Expiration:   rule.Window,
		KeyGenerator: func(c *fiber.Ctx) string { return api.ClientIP(c, clientIPHeader) },
		LimitReached: func(*fiber.Ctx) error { return cerrors.RateLimited() },
	})
}
