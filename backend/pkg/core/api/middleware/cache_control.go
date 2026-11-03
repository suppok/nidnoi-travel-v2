package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CacheControl marks successful responses as publicly cacheable for maxAge (browsers and the
// Cloudflare edge; a Cache Rule makes the path eligible). Errors are left uncached so a transient
// failure is not pinned for maxAge.
func CacheControl(maxAge time.Duration) fiber.Handler {
	value := fmt.Sprintf("public, max-age=%d", int(maxAge.Seconds()))
	return func(c *fiber.Ctx) error {
		if err := c.Next(); err != nil {
			return err
		}
		if c.Response().StatusCode() == http.StatusOK {
			c.Set(fiber.HeaderCacheControl, value)
		} else {
			c.Set(fiber.HeaderCacheControl, "no-store")
		}
		return nil
	}
}
