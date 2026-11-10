package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain"
	"backend/pkg/modules/auth/commons"
	"backend/pkg/modules/auth/models"
	"backend/pkg/modules/shared/client/line"
)

const (
	// lineLinkCodeTTL is how long the backoffice QR code stays valid.
	lineLinkCodeTTL = 10 * time.Minute
	// liffURLBase is LINE's LIFF launcher; it opens the LIFF app inside LINE (or a browser on PC).
	liffURLBase         = "https://liff.line.me/"
	maxLineLinkCodeLen  = 100
	lineUserIDUniqueKey = "app_user_line_user_id_key"
)

var lineLinkColumns = []string{
	"line_user_id", "line_display_name", "line_linked_at",
	"line_link_code_hash", "line_link_code_expires_at", "update_date", "update_by",
}

// verifyLIFFToken checks a LIFF access token with LINE.
func (u *authUseCase) verifyLIFFToken(ctx context.Context, accessToken string) (*line.TokenInfo, error) {
	client := u.shareRegistry.ClientRegistry.LINE
	if !client.Configured() {
		return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
	}
	info, err := client.VerifyAccessToken(ctx, accessToken)
	if err != nil {
		if errors.Is(err, line.ErrNotConfigured) {
			return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
		}
		u.coreRegistry.Logger.Warn("LIFF token verification failed", zap.Error(err))
		return nil, cerrors.New(http.StatusUnauthorized, errorcode.InvalidLINEToken)
	}
	return info, nil
}

func (u *authUseCase) CreateLineLink(ctx context.Context, request *models.AuthLineLinkCreateRequest) (*models.AuthLineLinkCreateResponse, error) {
	info := request.UserRequestInfo
	if info == nil {
		return nil, cerrors.Unauthenticated()
	}
	liffID := u.coreRegistry.Config.LINE.LIFFID
	if liffID == "" || !u.shareRegistry.ClientRegistry.LINE.Configured() {
		return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
	}
	code, err := commons.NewRefreshToken()
	if err != nil {
		return nil, cerrors.Wrap(err, "line link code")
	}

	now := time.Now()
	expiresAt := now.Add(lineLinkCodeTTL)
	user := &domain.AppUser{
		ID:                    info.UserID,
		LineLinkCodeHash:      null.StringFrom(commons.HashRefreshToken(code)),
		LineLinkCodeExpiresAt: null.TimeFrom(expiresAt),
		UpdateDate:            now,
		UpdateBy:              info.UserID,
	}
	if _, err := u.repositoryRegistry.AppUserRepository.Update(ctx, user, "line_link_code_hash", "line_link_code_expires_at", "update_date", "update_by"); err != nil {
		return nil, cerrors.Wrap(err, "create line link")
	}
	return &models.AuthLineLinkCreateResponse{
		URL:       liffURLBase + url.PathEscape(liffID) + "?lineLink=" + url.QueryEscape(code),
		ExpiresAt: expiresAt,
	}, nil
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

func (u *authUseCase) LIFFLineLink(ctx context.Context, request *models.LIFFLineLinkRequest) (*models.LIFFLineLinkResponse, error) {
	request.Code = strings.TrimSpace(request.Code)
	if request.Code == "" || len(request.Code) > maxLineLinkCodeLen {
		return nil, cerrors.New(http.StatusBadRequest, errorcode.LINELinkCodeInvalid)
	}
	token, err := u.verifyLIFFToken(ctx, request.AccessToken)
	if err != nil {
		return nil, err
	}

	var linked *domain.AppUser
	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.AppUserRepository.WithTx(appDB)
		user, err := repo.FindByLineLinkCodeHash(ctx, commons.HashRefreshToken(request.Code))
		if err != nil {
			return err
		}
		now := time.Now()
		if user == nil || !user.Status || !user.LineLinkCodeExpiresAt.Valid || now.After(user.LineLinkCodeExpiresAt.Time) {
			return cerrors.New(http.StatusBadRequest, errorcode.LINELinkCodeInvalid)
		}
		other, err := repo.FindByLineUserID(ctx, token.UserID)
		if err != nil {
			return err
		}
		if other != nil && other.ID != user.ID {
			return cerrors.Conflict(errorcode.LINEAlreadyLinked)
		}

		user.LineUserID = null.StringFrom(token.UserID)
		user.LineDisplayName = null.StringFrom(token.DisplayName)
		user.LineLinkedAt = null.TimeFrom(now)
		user.LineLinkCodeHash = null.String{}
		user.LineLinkCodeExpiresAt = null.Time{}
		user.UpdateDate = now
		user.UpdateBy = user.ID
		linked, err = repo.Update(ctx, user, lineLinkColumns...)
		return err
	})
	if query.UniqueViolationConstraint(err) == lineUserIDUniqueKey {
		return nil, cerrors.Conflict(errorcode.LINEAlreadyLinked)
	}
	if err != nil {
		return nil, cerrors.Wrap(err, "link line")
	}
	u.coreRegistry.Logger.Info("line account linked", zap.String("userId", linked.ID), zap.String("lineUserId", token.UserID))
	return &models.LIFFLineLinkResponse{Name: linked.Name, Email: linked.Email}, nil
}

// ResolveLIFFSession verifies the LIFF token with LINE on every request, then loads the staff user
// linked to that LINE account, so deactivation and role changes apply immediately.
func (u *authUseCase) ResolveLIFFSession(ctx context.Context, accessToken string) (*coremodels.UserRequestInfo, error) {
	token, err := u.verifyLIFFToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	user, err := u.repositoryRegistry.AppUserRepository.FindByLineUserID(ctx, token.UserID)
	if err != nil {
		return nil, cerrors.Wrap(err, "resolve liff session")
	}
	if user == nil || !user.Status {
		return nil, cerrors.New(http.StatusForbidden, errorcode.LINENotLinked)
	}
	if user.MustChangePassword {
		return nil, cerrors.New(http.StatusForbidden, errorcode.PasswordChangeRequired)
	}
	return &coremodels.UserRequestInfo{
		UserID:          user.ID,
		Email:           user.Email,
		Name:            user.Name,
		Roles:           user.Roles,
		LineUserID:      token.UserID,
		LineDisplayName: token.DisplayName,
	}, nil
}

func (u *authUseCase) LIFFMe(_ context.Context, request *models.LIFFMeRequest) (*models.LIFFMeResponse, error) {
	info := request.UserRequestInfo
	if info == nil {
		return nil, cerrors.Unauthenticated()
	}
	return &models.LIFFMeResponse{
		Name:            info.Name,
		Email:           info.Email,
		Roles:           info.Roles,
		LineDisplayName: null.NewString(info.LineDisplayName, info.LineDisplayName != ""),
	}, nil
}
