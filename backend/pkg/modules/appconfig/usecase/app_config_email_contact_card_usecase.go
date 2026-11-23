package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"backend/pkg/constants/appconfigcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/appconfig/models"
	"backend/pkg/modules/shared/services/notification"
)

// EmailContactCardGet returns the contact card used at the bottom of emails (admins).
func (u *appConfigUseCase) EmailContactCardGet(ctx context.Context, _ *models.AppConfigEmailContactCardGetRequest) (*models.AppConfigEmailContactCardResponse, error) {
	row, err := u.repositoryRegistry.AppConfigRepository.FindByCode(ctx, appconfigcode.AppConfigEmailContactCard.Code())
	if err != nil {
		return nil, cerrors.Wrap(err, "get email contact card")
	}
	return u.emailContactCardResponse(ctx, row)
}

func (u *appConfigUseCase) EmailContactCardUpdateValidate(request *models.AppConfigEmailContactCardUpdateRequest) error {
	base := u.shareRegistry.ClientRegistry.ObjectStorage.PublicBaseURL()
	return validation.ValidateStruct(request,
		validation.Field(&request.ImageURL, validation.By(func(any) error {
			if request.ImageURL == "" {
				return nil
			}
			if _, ok := notification.ContactCardID(request.ImageURL, base); !ok {
				return errors.New("must be an image uploaded with purpose email-contact-card")
			}
			return nil
		})),
	)
}

// EmailContactCardUpdate saves the contact card (admins). It applies to emails queued from now on;
// queued emails keep the card they were queued with.
func (u *appConfigUseCase) EmailContactCardUpdate(ctx context.Context, request *models.AppConfigEmailContactCardUpdateRequest) (*models.AppConfigEmailContactCardResponse, error) {
	if err := u.EmailContactCardUpdateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	code := appconfigcode.AppConfigEmailContactCard
	value, err := json.Marshal(struct {
		ImageURL string `json:"imageUrl"`
	}{request.ImageURL})
	if err != nil {
		return nil, cerrors.Wrap(err, "encode email contact card")
	}

	var saved *domain.AppConfig
	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.AppConfigRepository.WithTx(appDB)
		row, err := repo.FindByCode(ctx, code.Code())
		if err != nil {
			return err
		}
		now := time.Now()
		actor := request.UserRequestInfo.UserID
		if row == nil {
			saved, err = repo.Create(ctx, &domain.AppConfig{
				ID:          utils.NewID(),
				Code:        code.Code(),
				Value:       string(value),
				ModelType:   code.Code(),
				Description: "Contact card image at the bottom of emails",
				Status:      true,
				CreateDate:  now,
				UpdateDate:  now,
				CreateBy:    actor,
				UpdateBy:    actor,
			})
			return err
		}
		row.Value = string(value)
		row.Status = true
		row.UpdateDate = now
		row.UpdateBy = actor
		saved, err = repo.Update(ctx, row, "value", "status", "update_date", "update_by")
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "save email contact card")
	}
	return u.emailContactCardResponse(ctx, saved)
}

func (u *appConfigUseCase) emailContactCardResponse(ctx context.Context, row *domain.AppConfig) (*models.AppConfigEmailContactCardResponse, error) {
	response := &models.AppConfigEmailContactCardResponse{
		DefaultImage: "data:image/png;base64," + base64.StdEncoding.EncodeToString(notification.DefaultContactCard()),
	}
	if row == nil || !row.Status {
		return response, nil
	}
	if v, err := appconfigcode.AppConfigEmailContactCard.ToValue(row.Value); err == nil {
		response.ImageURL = v.ImageURL
	}
	response.UpdateDate = models.NullTime(row.UpdateDate)
	if user, err := u.repositoryRegistry.AppUserRepository.FindByID(ctx, row.UpdateBy); err != nil {
		return nil, cerrors.Wrap(err, "get email contact card editor")
	} else if user != nil {
		response.UpdateBy = user.Name
	}
	return response, nil
}
