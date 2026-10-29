package usecase

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/user"
	"backend/pkg/registry"
)

type userUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewUserUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) user.UserUseCase {
	return &userUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

// Shared field rules for create and update.

func emailRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.Length(3, 254), validation.Email}
}

func nameRules() []validation.Rule {
	return []validation.Rule{validation.Required, validation.RuneLength(1, 100)}
}

// rolesRules accepts the given roles only; create passes role.Assignable, update role.All (the
// super admin keeps super_admin when editing their own profile; granting it is rejected in Update).
func rolesRules(allowed []string) []validation.Rule {
	known := make([]any, len(allowed))
	for i, r := range allowed {
		known[i] = r
	}
	return []validation.Rule{validation.Required, validation.Each(validation.In(known...))}
}

// normalizeProfile trims, lowercases the email and de-duplicates roles in the fixed order.
func normalizeProfile(email, name *string, roles *[]string) {
	*email = strings.ToLower(strings.TrimSpace(*email))
	*name = strings.TrimSpace(*name)
	var out []string
	for _, r := range role.All {
		if slices.Contains(*roles, r) {
			out = append(out, r)
		}
	}
	// Unknown values are kept so validation reports them.
	for _, r := range *roles {
		if !role.Valid(r) {
			out = append(out, r)
		}
	}
	*roles = out
}

func sameRoles(a, b []string) bool {
	return len(a) == len(b) && !slices.ContainsFunc(a, func(r string) bool { return !slices.Contains(b, r) })
}

// ensureCanGrant fails with SUPER_ADMIN_REQUIRED when a non-super-admin assigns the admin role.
func ensureCanGrant(actor *coremodels.UserRequestInfo, roles []string) error {
	if slices.Contains(roles, role.Admin) && !slices.Contains(actor.Roles, role.SuperAdmin) {
		return cerrors.New(http.StatusForbidden, errorcode.SuperAdminRequired)
	}
	return nil
}

// ensureCanManage fails when actor may not change target: nobody else may change the super admin
// (SUPER_ADMIN_PROTECTED), and only the super admin may change an admin (SUPER_ADMIN_REQUIRED).
// Otherwise an admin could reset a fellow admin's password and sign in as them.
func ensureCanManage(actor *coremodels.UserRequestInfo, target *domain.AppUser) error {
	if slices.Contains(target.Roles, role.SuperAdmin) {
		return cerrors.Conflict(errorcode.SuperAdminProtected)
	}
	return ensureCanGrant(actor, target.Roles)
}

// findForUpdate loads the target user inside a transaction or fails with NOT_FOUND.
func (u *userUseCase) findForUpdate(ctx context.Context, appDB query.AppIDB, id string) (*domain.AppUser, error) {
	target, err := u.repositoryRegistry.AppUserRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, cerrors.NotFound()
	}
	return target, nil
}
