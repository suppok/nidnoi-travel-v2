package appconfig

import (
	"context"

	"backend/pkg/modules/appconfig/models"
)

type AppConfigUseCase interface {
	Verify(ctx context.Context, request *models.AppConfigVerifyRequest) (*models.AppConfigVerifyResponse, error)
	AppSettingsGet(ctx context.Context, request *models.AppConfigAppSettingsGetRequest) (*models.AppConfigAppSettingsGetResponse, error)
	// EmailContactCardGet / EmailContactCardUpdate: the contact card image at the bottom of emails.
	EmailContactCardGet(ctx context.Context, request *models.AppConfigEmailContactCardGetRequest) (*models.AppConfigEmailContactCardResponse, error)
	EmailContactCardUpdate(ctx context.Context, request *models.AppConfigEmailContactCardUpdateRequest) (*models.AppConfigEmailContactCardResponse, error)
}
