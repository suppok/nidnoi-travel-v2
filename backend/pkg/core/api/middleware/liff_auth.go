package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/core/api"
	coremodels "backend/pkg/core/api/models"
	cerrors "backend/pkg/core/errors"
)

// LIFFSessionResolver turns a LIFF access token into the staff user who linked that LINE account.
// It returns an *AppError (INVALID_LINE_TOKEN, LINE_NOT_LINKED, …) when there is none.
type LIFFSessionResolver interface {
	ResolveLIFFSession(ctx context.Context, accessToken string) (*coremodels.UserRequestInfo, error)
}

// LIFFAuth authenticates the LIFF app: "Authorization: Bearer <LIFF access token>". The token is
// verified with LINE on every request, then mapped to the linked staff user, so routes can use
// AuthPermissions exactly like JWT routes.
type LIFFAuth interface {
	Auth(c *fiber.Ctx) error
}

type liffAuth struct{ resolver LIFFSessionResolver }

func NewLIFFAuth(resolver LIFFSessionResolver) LIFFAuth {
	return &liffAuth{resolver: resolver}
}

// BearerToken returns the bearer token of the request, or "".
func BearerToken(c *fiber.Ctx) string {
	token, _ := strings.CutPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
	return strings.TrimSpace(token)
}

func (a *liffAuth) Auth(c *fiber.Ctx) error {
	token := BearerToken(c)
	if token == "" {
		return cerrors.Unauthenticated()
	}
	info, err := a.resolver.ResolveLIFFSession(c.UserContext(), token)
	if err != nil {
		return err
	}
	api.SetUserRequestInfo(c, info)
	return c.Next()
}
