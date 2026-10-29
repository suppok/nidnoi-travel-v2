package usecase

import (
	"context"
	"slices"
	"time"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/role"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/user/models"
)

func (u *userUseCase) UpdateValidate(request *models.UserUpdateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Email, emailRules()...),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.Roles, rolesRules(role.All)...),
	)
}

// Update changes email, name and roles. Roles take effect on the user's next request. Only the
// super admin may edit the super admin (name and email; their roles are fixed), and only the
// super admin may edit an admin or grant the admin role.
func (u *userUseCase) Update(ctx context.Context, request *models.UserUpdateRequest) (*models.UserUpdateResponse, error) {
	normalizeProfile(&request.Email, &request.Name, &request.Roles)
	if err := u.UpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.AppUser
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		users := u.repositoryRegistry.AppUserRepository.WithTx(appDB)

		target, err := u.findForUpdate(ctx, appDB, request.ID)
		if err != nil {
			return err
		}

		isSelf := target.ID == request.UserRequestInfo.UserID
		if !isSelf {
			if err := ensureCanManage(request.UserRequestInfo, target); err != nil {
				return err
			}
			if err := ensureCanGrant(request.UserRequestInfo, request.Roles); err != nil {
				return err
			}
		}
		if !sameRoles(target.Roles, request.Roles) {
			if isSelf {
				return cerrors.Conflict(errorcode.CannotModifySelf)
			}
			if slices.Contains(request.Roles, role.SuperAdmin) {
				return cerrors.FieldError("roles", "super_admin cannot be assigned")
			}
		}

		if request.Email != target.Email {
			existing, err := users.FindByEmail(ctx, request.Email)
			if err != nil {
				return err
			}
			if existing != nil {
				return cerrors.Conflict(errorcode.EmailTaken)
			}
		}

		target.Email = request.Email
		target.Name = request.Name
		target.Roles = request.Roles
		target.UpdateDate = time.Now()
		target.UpdateBy = request.UserRequestInfo.UserID
		updated, err = users.Update(ctx, target, "email", "name", "roles", "update_date", "update_by")
		if query.IsUniqueViolation(err) {
			return cerrors.Conflict(errorcode.EmailTaken)
		}
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update user")
	}
	return &models.UserUpdateResponse{UserItem: models.NewUserItem(updated)}, nil
}
