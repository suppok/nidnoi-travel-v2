package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/api"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/auth/jwt"
	cerrors "backend/pkg/core/errors"
)

// SessionResolver loads the current state of an authenticated user. It returns (nil, nil) when
// the user no longer exists or is deactivated.
type SessionResolver interface {
	ResolveSession(ctx context.Context, userID string) (*coremodels.UserRequestInfo, error)
}

// Auth verifies the bearer access token and loads the user.
type Auth interface {
	// Auth rejects users who still have to replace a temporary password (PASSWORD_CHANGE_REQUIRED).
	Auth(c *fiber.Ctx) error
	// AuthAllowPasswordChange is for GET /auth/me and PUT /auth/password only.
	AuthAllowPasswordChange(c *fiber.Ctx) error
}

type auth struct {
	jwt      jwt.Service
	resolver SessionResolver
}

func NewAuth(jwtService jwt.Service, resolver SessionResolver) Auth {
	return &auth{jwt: jwtService, resolver: resolver}
}

func (a *auth) Auth(c *fiber.Ctx) error {
	return a.authenticate(c, false)
}

func (a *auth) AuthAllowPasswordChange(c *fiber.Ctx) error {
	return a.authenticate(c, true)
}

func (a *auth) authenticate(c *fiber.Ctx, allowPasswordChange bool) error {
	header := c.Get(fiber.HeaderAuthorization)
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || token == "" {
		return cerrors.Unauthenticated()
	}

	userID, err := a.jwt.Parse(token)
	if err != nil {
		return cerrors.Unauthenticated()
	}

	info, err := a.resolver.ResolveSession(c.UserContext(), userID)
	if err != nil {
		return err
	}
	if info == nil {
		return cerrors.Unauthenticated()
	}
	if info.MustChangePassword && !allowPasswordChange {
		return cerrors.New(http.StatusForbidden, errorcode.PasswordChangeRequired)
	}

	api.SetUserRequestInfo(c, info)
	return c.Next()
}
