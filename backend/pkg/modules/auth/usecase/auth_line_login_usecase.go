package usecase

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/modules/auth/models"
	"backend/pkg/modules/shared/client/line"
)

// LineLoginCallbackPath is where LINE sends the browser back to the backoffice.
const LineLoginCallbackPath = "/login/line/callback"

// LineLoginConfig tells the login page whether to show "ล็อกอินด้วย LINE".
func (u *authUseCase) LineLoginConfig(_ context.Context) (*models.AuthLineLoginConfigResponse, error) {
	client := u.shareRegistry.ClientRegistry.LINE
	if !client.LoginConfigured() {
		return &models.AuthLineLoginConfigResponse{}, nil
	}
	return &models.AuthLineLoginConfigResponse{Enabled: true, ChannelID: client.LoginChannelID()}, nil
}

func (u *authUseCase) LineLoginValidate(request *models.AuthLineLoginRequest) error {
	fields := validation.Errors{}
	if err := validation.ValidateStruct(request,
		validation.Field(&request.Code, validation.Required, validation.RuneLength(1, 512)),
		validation.Field(&request.CodeVerifier, validation.Required, validation.RuneLength(43, 128)),
		validation.Field(&request.Nonce, validation.Required, validation.RuneLength(1, 255)),
		validation.Field(&request.RedirectURI, validation.Required),
	); err != nil {
		return err
	}
	// Only our own backoffice callback: an allowed origin + the callback path.
	origin, ok := strings.CutSuffix(request.RedirectURI, LineLoginCallbackPath)
	if !ok || !slices.Contains(u.coreRegistry.Config.CORSAllowedOrigins, origin) {
		fields["redirectUri"] = errors.New("must be the backoffice LINE callback")
	}
	return fields.Filter()
}

// LineLogin signs a staff user in with LINE: LINE vouches for the account, and the account must be
// the one a backoffice user linked ("เชื่อมบัญชี LINE"). An unlinked or deactivated account is
// denied with LINE_NOT_LINKED. The session is the same as a password login.
func (u *authUseCase) LineLogin(ctx context.Context, request *models.AuthLineLoginRequest) (*models.AuthLoginResponse, error) {
	if err := u.LineLoginValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	account, err := u.shareRegistry.ClientRegistry.LINE.ExchangeLoginCode(ctx, request.Code, request.RedirectURI, request.CodeVerifier, request.Nonce)
	switch {
	case errors.Is(err, line.ErrNotConfigured):
		return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
	case errors.Is(err, line.ErrLoginRejected):
		return nil, cerrors.New(http.StatusUnauthorized, errorcode.InvalidLINELogin)
	case err != nil:
		return nil, cerrors.Wrap(err, "LINE login")
	}

	user, err := u.repositoryRegistry.AppUserRepository.FindByLineUserID(ctx, account.UserID)
	if err != nil {
		return nil, cerrors.Wrap(err, "find user by LINE account")
	}
	if user == nil || !user.Status {
		u.coreRegistry.Logger.Info("LINE login denied: account not linked", zap.String("lineDisplayName", account.DisplayName))
		return nil, cerrors.New(http.StatusForbidden, errorcode.LINENotLinked)
	}
	return u.startSession(ctx, user, request.RememberMe, request.ClientIP, request.UserAgent)
}
