package middleware

import (
	"github.com/gofiber/fiber/v2"

	"backend/pkg/constants/apicode"
	"backend/pkg/core/api"
	"backend/pkg/core/auth/permission"
	cerrors "backend/pkg/core/errors"
)

// AuthPermissions checks Casbin policy for the authenticated user's roles. Always runs after Auth.
type AuthPermissions interface {
	RequiresPermissions(code apicode.Code) fiber.Handler
}

type authPermissions struct{ checker permission.Checker }

func NewAuthPermissions(checker permission.Checker) AuthPermissions {
	return &authPermissions{checker: checker}
}

func (p *authPermissions) RequiresPermissions(code apicode.Code) fiber.Handler {
	return func(c *fiber.Ctx) error {
		info := api.UserRequestInfoFrom(c)
		if info == nil {
			return cerrors.Unauthenticated()
		}
		if !p.checker.Allowed(info.Roles, code) {
			return cerrors.PermissionDenied()
		}
		return c.Next()
	}
}
