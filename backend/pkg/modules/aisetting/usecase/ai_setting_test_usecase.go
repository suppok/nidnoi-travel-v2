package usecase

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"backend/pkg/constants/aimodel"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/aisetting/models"
	"backend/pkg/modules/shared/services/aisettings"
)

func (u *aiSettingUseCase) TestValidate(request *models.AISettingTestRequest) error {
	request.APIKey = strings.TrimSpace(request.APIKey)
	return validation.ValidateStruct(request,
		validation.Field(&request.APIKey, validation.By(validAPIKey)),
		validation.Field(&request.Model, validation.In(aimodel.IDs()...)),
	)
}

// Test checks a key with GET /v1/models/{model} (no tokens billed): the key typed on the page when
// given, else the stored one. Nothing is saved.
func (u *aiSettingUseCase) Test(ctx context.Context, request *models.AISettingTestRequest) (*models.AISettingTestResponse, error) {
	if err := u.TestValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	key, model := request.APIKey, request.Model
	if key == "" || model == "" {
		s, err := u.shareRegistry.ServiceRegistry.AISettings.Load(ctx)
		if err != nil {
			return nil, err
		}
		if key == "" {
			if !s.HasAPIKey() {
				if s.KeyUnreadable {
					return nil, aisettings.NotConfigured("the stored API key can't be decrypted; enter it again")
				}
				return nil, cerrors.FieldError("apiKey", "enter an API key or save one first")
			}
			key = s.APIKey
		}
		if model == "" {
			model = s.Model
		}
	}
	if err := u.shareRegistry.ClientRegistry.Anthropic.Ping(ctx, key, model); err != nil {
		u.coreRegistry.Logger.Info("ai_settings_test_failed", zap.String("model", model), zap.Error(err))
		return nil, aisettings.MapClientError(err)
	}
	return &models.AISettingTestResponse{OK: true, Model: model}, nil
}
