package usecase

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain"
	"backend/pkg/modules/auth/models"
	"backend/pkg/modules/shared/client/line"
)

const lineUserIDUniqueKey = "app_user_line_user_id_key"

var lineLinkColumns = []string{"line_user_id", "line_display_name", "line_linked_at", "update_date", "update_by"}

func (u *authUseCase) LineLinkValidate(request *models.AuthLineLinkRequest) error {
	return u.validateLineCallback(request.Code, request.CodeVerifier, request.Nonce, request.RedirectURI)
}

// LineLink links the caller's own LINE account: the backoffice ran LINE Login ("เชื่อมบัญชี LINE")
// and sends the code back. A LINE account belongs to one user at most (LINE_ALREADY_LINKED); a new
// link replaces the caller's previous one.
func (u *authUseCase) LineLink(ctx context.Context, request *models.AuthLineLinkRequest) (*models.AuthLineLinkResponse, error) {
	info := request.UserRequestInfo
	if info == nil {
		return nil, cerrors.Unauthenticated()
	}
	if err := u.LineLinkValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	account, err := u.exchangeLineCode(ctx, request.Code, request.RedirectURI, request.CodeVerifier, request.Nonce)
	if err != nil {
		return nil, err
	}

	other, err := u.repositoryRegistry.AppUserRepository.FindByLineUserID(ctx, account.UserID)
	if err != nil {
		return nil, cerrors.Wrap(err, "find user by LINE account")
	}
	if other != nil && other.ID != info.UserID {
		return nil, cerrors.Conflict(errorcode.LINEAlreadyLinked)
	}

	now := time.Now()
	user := &domain.AppUser{
		ID:              info.UserID,
		LineUserID:      null.StringFrom(account.UserID),
		LineDisplayName: null.StringFrom(account.DisplayName),
		LineLinkedAt:    null.TimeFrom(now),
		UpdateDate:      now,
		UpdateBy:        info.UserID,
	}
	_, err = u.repositoryRegistry.AppUserRepository.Update(ctx, user, lineLinkColumns...)
	if query.UniqueViolationConstraint(err) == lineUserIDUniqueKey {
		return nil, cerrors.Conflict(errorcode.LINEAlreadyLinked)
	}
	if err != nil {
		return nil, cerrors.Wrap(err, "link line")
	}
	u.coreRegistry.Logger.Info("line account linked", zap.String("userId", info.UserID), zap.String("lineUserId", account.UserID))
	return &models.AuthLineLinkResponse{Line: models.AuthLineAccount{DisplayName: account.DisplayName, LinkedAt: now}}, nil
}

func (u *authUseCase) LineUnlink(ctx context.Context, request *models.AuthLineUnlinkRequest) (*models.AuthLineUnlinkResponse, error) {
	info := request.UserRequestInfo
	if info == nil {
		return nil, cerrors.Unauthenticated()
	}
	user := &domain.AppUser{ID: info.UserID, UpdateDate: time.Now(), UpdateBy: info.UserID}
	if _, err := u.repositoryRegistry.AppUserRepository.Update(ctx, user, lineLinkColumns...); err != nil {
		return nil, cerrors.Wrap(err, "unlink line")
	}
	return &models.AuthLineUnlinkResponse{}, nil
}

// exchangeLineCode redeems a LINE Login code from the backoffice callback.
func (u *authUseCase) exchangeLineCode(ctx context.Context, code, redirectURI, codeVerifier, nonce string) (*line.TokenInfo, error) {
	account, err := u.shareRegistry.ClientRegistry.LINE.ExchangeLoginCode(ctx, code, redirectURI, codeVerifier, nonce)
	switch {
	case errors.Is(err, line.ErrNotConfigured):
		return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
	case errors.Is(err, line.ErrLoginRejected):
		return nil, cerrors.New(http.StatusUnauthorized, errorcode.InvalidLINELogin)
	case err != nil:
		return nil, cerrors.Wrap(err, "LINE login")
	}
	return account, nil
}
