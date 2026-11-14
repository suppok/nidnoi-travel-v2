package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/core/config"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain"
	"backend/pkg/modules/auth/models"
	"backend/pkg/modules/shared/client/line"
	"backend/pkg/registry"
)

// fakeLINELogin answers ExchangeLoginCode from a code → LINE user ID map.
type fakeLINELogin struct {
	line.Client
	codes map[string]string
}

func (f *fakeLINELogin) LoginConfigured() bool  { return true }
func (f *fakeLINELogin) LoginChannelID() string { return "1234567890" }

func (f *fakeLINELogin) ExchangeLoginCode(_ context.Context, code, _, _, _ string) (*line.TokenInfo, error) {
	id, ok := f.codes[code]
	if !ok {
		return nil, line.ErrLoginRejected
	}
	return &line.TokenInfo{UserID: id}, nil
}

func TestLineLoginDenied(t *testing.T) {
	u := &authUseCase{
		coreRegistry: &registry.CoreRegistry{
			Logger: zap.NewNop(),
			Config: &config.Config{CORSAllowedOrigins: []string{"https://admin.example.com"}},
		},
		shareRegistry: &registry.ShareRegistry{ClientRegistry: registry.ClientRegistry{LINE: &fakeLINELogin{codes: map[string]string{"c-off": "U2", "c-stranger": "U9"}}}},
		repositoryRegistry: &registry.RepositoryRegistry{AppUserRepository: &fakeUsers{users: []domain.AppUser{
			{ID: "off", Roles: []string{"sales"}, Status: false, LineUserID: null.StringFrom("U2")},
		}}},
	}
	request := func(code, redirect string) *models.AuthLineLoginRequest {
		return &models.AuthLineLoginRequest{
			Code:         code,
			RedirectURI:  redirect,
			CodeVerifier: "v123456789012345678901234567890123456789012345",
			Nonce:        "n",
		}
	}
	const callback = "https://admin.example.com/login/line/callback"

	for name, tc := range map[string]struct {
		req  *models.AuthLineLoginRequest
		want string
	}{
		"not linked":       {request("c-stranger", callback), "LINE_NOT_LINKED"},
		"deactivated user": {request("c-off", callback), "LINE_NOT_LINKED"},
		"code refused":     {request("c-bad", callback), "INVALID_LINE_LOGIN"},
		"foreign origin":   {request("c-stranger", "https://evil.example.com/login/line/callback"), "VALIDATION"},
		"other path":       {request("c-stranger", "https://admin.example.com/somewhere"), "VALIDATION"},
		"missing verifier": {&models.AuthLineLoginRequest{Code: "c", RedirectURI: callback, Nonce: "n"}, "VALIDATION"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := u.LineLogin(context.Background(), tc.req)
			var appErr *cerrors.AppError
			if !errors.As(err, &appErr) || appErr.Code != tc.want {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}
