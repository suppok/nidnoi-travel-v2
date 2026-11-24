package aisetting

import (
	"context"

	"backend/pkg/modules/aisetting/models"
)

type AISettingUseCase interface {
	Get(ctx context.Context, request *models.AISettingGetRequest) (*models.AISettingResponse, error)
	Update(ctx context.Context, request *models.AISettingUpdateRequest) (*models.AISettingResponse, error)
	// Test checks an API key (the stored one or one being entered) with a free Anthropic call.
	Test(ctx context.Context, request *models.AISettingTestRequest) (*models.AISettingTestResponse, error)
}
