package appconfig

import (
	"context"

	"backend/pkg/modules/appconfig/models"
)

type AppConfigUseCase interface {
	Verify(ctx context.Context, request *models.AppConfigVerifyRequest) (*models.AppConfigVerifyResponse, error)
	AppSettingsGet(ctx context.Context, request *models.AppConfigAppSettingsGetRequest) (*models.AppConfigAppSettingsGetResponse, error)
}
