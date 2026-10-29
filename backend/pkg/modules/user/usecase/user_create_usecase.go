package usecase

import (
	"context"
	"time"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/role"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/user/models"
)

func (u *userUseCase) CreateValidate(request *models.UserCreateRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Email, emailRules()...),
		validation.Field(&request.Name, nameRules()...),
		validation.Field(&request.Roles, rolesRules(role.Assignable)...),
		validation.Field(&request.TemporaryPassword, validation.Required, u.shareRegistry.ServiceRegistry.Password.Rule()),
	)
}

// Create adds a staff user with a temporary password; they must replace it at first login.
func (u *userUseCase) Create(ctx context.Context, request *models.UserCreateRequest) (*models.UserCreateResponse, error) {
	normalizeProfile(&request.Email, &request.Name, &request.Roles)
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	if err := ensureCanGrant(request.UserRequestInfo, request.Roles); err != nil {
		return nil, err
	}

	hash, err := u.shareRegistry.ServiceRegistry.Password.Hash(request.TemporaryPassword)
	if err != nil {
		return nil, cerrors.Wrap(err, "hash password")
	}

	var created *domain.AppUser
	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		users := u.repositoryRegistry.AppUserRepository.WithTx(appDB)

		existing, err := users.FindByEmail(ctx, request.Email)
		if err != nil {
			return err
		}
		if existing != nil {
			return cerrors.Conflict(errorcode.EmailTaken)
		}

		now := time.Now()
		created, err = users.Create(ctx, &domain.AppUser{
			ID:                 utils.NewID(),
			Email:              request.Email,
			Name:               request.Name,
			PasswordHash:       hash,
			Roles:              request.Roles,
			MustChangePassword: true,
			Status:             true,
			CreateDate:         now,
			UpdateDate:         now,
			CreateBy:           request.UserRequestInfo.UserID,
			UpdateBy:           request.UserRequestInfo.UserID,
		})
		if query.IsUniqueViolation(err) {
			return cerrors.Conflict(errorcode.EmailTaken)
		}
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "create user")
	}
	return &models.UserCreateResponse{UserItem: models.NewUserItem(created)}, nil
}
