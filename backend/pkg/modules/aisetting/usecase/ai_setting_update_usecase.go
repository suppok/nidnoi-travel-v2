package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/aimodel"
	"backend/pkg/constants/appconfigcode"
	appconfigmodels "backend/pkg/constants/appconfigcode/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/aisetting/models"
	"backend/pkg/modules/shared/services/aisettings"
)

// apiKeyPattern: Anthropic API keys are "sk-ant-" followed by URL-safe characters.
var apiKeyPattern = regexp.MustCompile(`^sk-ant-[A-Za-z0-9_\-]{10,300}$`)

func validAPIKey(value any) error {
	if s, _ := value.(string); s != "" && !apiKeyPattern.MatchString(s) {
		return errors.New("must be an Anthropic API key (sk-ant-…)")
	}
	return nil
}

func (u *aiSettingUseCase) UpdateValidate(request *models.AISettingUpdateRequest, hasStoredKey bool) error {
	request.APIKey = null.StringFrom(strings.TrimSpace(request.APIKey.String))
	newKey := request.APIKey.String
	limit := u.shareRegistry.ServiceRegistry.AISettings.MaxPdfSizeMbLimit()
	return validation.ValidateStruct(request,
		validation.Field(&request.Model, validation.Required, validation.In(aimodel.IDs()...)),
		validation.Field(&request.MaxPdfSizeMb, validation.Required, validation.Min(1), validation.Max(limit)),
		validation.Field(&request.APIKey, validation.By(func(any) error {
			if err := validAPIKey(newKey); err != nil {
				return err
			}
			if newKey != "" && request.ClearAPIKey {
				return errors.New("cannot set and clear the key at once")
			}
			return nil
		})),
		validation.Field(&request.Enabled, validation.By(func(any) error {
			keyAfter := newKey != "" || (hasStoredKey && !request.ClearAPIKey)
			if request.Enabled && !keyAfter {
				return errors.New("an API key is required to turn AI on")
			}
			return nil
		})),
	)
}

// Update saves the AI settings (admins). A new apiKey is encrypted before it is stored; without one
// the stored key is kept unless clearApiKey is set.
func (u *aiSettingUseCase) Update(ctx context.Context, request *models.AISettingUpdateRequest) (*models.AISettingResponse, error) {
	svc := u.shareRegistry.ServiceRegistry.AISettings
	if !svc.Configured() {
		return nil, aisettings.NotConfigured("AI_SETTINGS_ENCRYPTION_KEY is not set on the server")
	}
	current, err := svc.Load(ctx)
	if err != nil {
		return nil, err
	}
	if err := u.UpdateValidate(request, current.HasAPIKey()); err != nil {
		return nil, cerrors.Validation(err)
	}

	code := appconfigcode.AppConfigAISettings
	var stored appconfigmodels.AppConfigAISettings
	if row, err := u.repositoryRegistry.AppConfigRepository.FindByCode(ctx, code.Code()); err != nil {
		return nil, cerrors.Wrap(err, "find AI settings")
	} else if row != nil {
		if v, err := code.ToValue(row.Value); err == nil {
			stored = v
		}
	}
	value := appconfigmodels.AppConfigAISettings{
		Enabled:          request.Enabled,
		Model:            request.Model,
		MaxPdfSizeMb:     request.MaxPdfSizeMb,
		APIKeyCiphertext: stored.APIKeyCiphertext,
		APIKeyLast4:      stored.APIKeyLast4,
	}
	if current.KeyUnreadable && request.APIKey.String == "" {
		// An unreadable key is useless; keep it only until a new one is entered or it is cleared.
		value.Enabled = false
	}
	switch {
	case request.APIKey.String != "":
		sealed, err := svc.SealAPIKey(request.APIKey.String)
		if err != nil {
			return nil, cerrors.Wrap(err, "encrypt API key")
		}
		value.APIKeyCiphertext = sealed
		value.APIKeyLast4 = request.APIKey.String[len(request.APIKey.String)-4:]
	case request.ClearAPIKey:
		value.APIKeyCiphertext, value.APIKeyLast4 = "", ""
	}
	if err := code.Validate(mustJSON(value)); err != nil {
		return nil, cerrors.Wrap(err, "AI settings value")
	}

	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.AppConfigRepository.WithTx(appDB)
		row, err := repo.FindByCode(ctx, code.Code())
		if err != nil {
			return err
		}
		now := time.Now()
		actor := request.UserRequestInfo.UserID
		if row == nil {
			_, err = repo.Create(ctx, &domain.AppConfig{
				ID:          utils.NewID(),
				Code:        code.Code(),
				Value:       mustJSON(value),
				ModelType:   code.Code(),
				Description: "AI settings: Claude API key (encrypted), model and on/off",
				Status:      true,
				CreateDate:  now,
				UpdateDate:  now,
				CreateBy:    actor,
				UpdateBy:    actor,
			})
			return err
		}
		row.Value = mustJSON(value)
		row.Status = true
		row.UpdateDate = now
		row.UpdateBy = actor
		_, err = repo.Update(ctx, row, "value", "status", "update_date", "update_by")
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "save AI settings")
	}
	u.coreRegistry.Logger.Info("ai_settings_updated",
		zap.String("userId", request.UserRequestInfo.UserID),
		zap.Bool("enabled", value.Enabled),
		zap.String("model", value.Model),
		zap.Bool("keyChanged", request.APIKey.String != "" || request.ClearAPIKey))

	saved, err := svc.Load(ctx)
	if err != nil {
		return nil, err
	}
	return u.response(ctx, saved)
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err) // plain struct of strings, ints and bools
	}
	return string(b)
}
