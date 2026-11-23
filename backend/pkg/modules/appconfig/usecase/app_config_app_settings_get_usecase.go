package usecase

import (
	"context"

	"backend/pkg/constants/appconfigcode"
	"backend/pkg/modules/appconfig/models"
	"backend/pkg/modules/shared/services/appconfigload"
)

// AppSettingsGet returns APP_SETTINGS, the example reader of a typed app config.
func (u *appConfigUseCase) AppSettingsGet(ctx context.Context, request *models.AppConfigAppSettingsGetRequest) (*models.AppConfigAppSettingsGetResponse, error) {
	cfg, err := appconfigload.GetAppConfig(ctx, u.repositoryRegistry.AppConfigRepository, appconfigcode.AppConfigAppSettings)
	if err != nil {
		return nil, err
	}
	return &models.AppConfigAppSettingsGetResponse{MaxFileSize: cfg.MaxFileSize}, nil
}
