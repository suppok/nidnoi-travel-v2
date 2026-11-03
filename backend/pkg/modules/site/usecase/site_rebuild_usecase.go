package usecase

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/modules/shared/services/revalidation"
	"backend/pkg/modules/site/models"
)

// Rebuild triggers a customer-site build now, bypassing the debounce (backoffice "rebuild" button).
func (u *siteUseCase) Rebuild(ctx context.Context, request *models.SiteRebuildRequest) (*models.SiteRebuildResponse, error) {
	err := u.shareRegistry.ServiceRegistry.Revalidator.RebuildNow(ctx)
	if revalidation.IsNotConfigured(err) {
		return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.SiteRebuildUnavailable)
	}
	if err != nil {
		u.coreRegistry.Logger.Error("manual site rebuild failed", zap.Error(err))
		return nil, cerrors.New(http.StatusBadGateway, errorcode.SiteRebuildFailed)
	}
	u.coreRegistry.Logger.Info("manual site rebuild", zap.String("userId", request.UserRequestInfo.UserID))
	return &models.SiteRebuildResponse{Triggered: true}, nil
}
