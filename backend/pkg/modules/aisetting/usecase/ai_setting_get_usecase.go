package usecase

import (
	"context"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/aimodel"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/aisetting/models"
	"backend/pkg/modules/shared/services/aisettings"
)

// Get returns the AI settings without the API key (admins).
func (u *aiSettingUseCase) Get(ctx context.Context, _ *models.AISettingGetRequest) (*models.AISettingResponse, error) {
	s, err := u.shareRegistry.ServiceRegistry.AISettings.Load(ctx)
	if err != nil {
		return nil, err
	}
	return u.response(ctx, s)
}

func (u *aiSettingUseCase) response(ctx context.Context, s *aisettings.Settings) (*models.AISettingResponse, error) {
	svc := u.shareRegistry.ServiceRegistry.AISettings
	res := &models.AISettingResponse{
		Enabled:              s.Enabled,
		Model:                s.Model,
		MaxPdfSizeMb:         s.MaxPdfSizeMb,
		HasAPIKey:            s.APIKeyLast4 != "",
		APIKeyLast4:          s.APIKeyLast4,
		KeyUnreadable:        s.KeyUnreadable,
		EncryptionConfigured: svc.Configured(),
		Ready:                svc.Configured() && s.Ready(),
		MaxPdfSizeMbLimit:    svc.MaxPdfSizeMbLimit(),
		Models:               aimodel.All,
	}
	if s.Saved {
		res.UpdateDate = null.TimeFrom(s.UpdateDate)
		user, err := u.repositoryRegistry.AppUserRepository.FindByID(ctx, s.UpdateBy)
		if err != nil {
			return nil, cerrors.Wrap(err, "get AI settings editor")
		}
		if user != nil {
			res.UpdateBy = user.Name
		}
	}
	return res, nil
}
