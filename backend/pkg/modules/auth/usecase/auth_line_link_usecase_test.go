package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain"
	"backend/pkg/modules/shared/client/line"
	"backend/pkg/registry"
	"backend/pkg/repository"
)

type fakeLINE struct {
	line.Client
	tokens map[string]string // access token → LINE user ID
}

func (f *fakeLINE) Configured() bool { return true }

func (f *fakeLINE) VerifyAccessToken(_ context.Context, token string) (*line.TokenInfo, error) {
	id, ok := f.tokens[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return &line.TokenInfo{UserID: id, DisplayName: "LINE " + id}, nil
}

type fakeUsers struct {
	repository.AppUserRepository
	users []domain.AppUser
}

func (f *fakeUsers) FindByLineUserID(_ context.Context, lineUserID string) (*domain.AppUser, error) {
	for i := range f.users {
		if f.users[i].LineUserID.String == lineUserID {
			return &f.users[i], nil
		}
	}
	return nil, nil
}

func TestResolveLIFFSession(t *testing.T) {
	u := &authUseCase{
		coreRegistry:  &registry.CoreRegistry{Logger: zap.NewNop()},
		shareRegistry: &registry.ShareRegistry{ClientRegistry: registry.ClientRegistry{LINE: &fakeLINE{tokens: map[string]string{"t-sales": "U1", "t-off": "U2", "t-temp": "U3", "t-stranger": "U9"}}}},
		repositoryRegistry: &registry.RepositoryRegistry{AppUserRepository: &fakeUsers{users: []domain.AppUser{
			{ID: "sales", Name: "Sales", Roles: []string{"sales"}, Status: true, LineUserID: null.StringFrom("U1")},
			{ID: "off", Roles: []string{"sales"}, Status: false, LineUserID: null.StringFrom("U2")},
			{ID: "temp", Roles: []string{"sales"}, Status: true, MustChangePassword: true, LineUserID: null.StringFrom("U3")},
		}}},
	}

	info, err := u.ResolveLIFFSession(context.Background(), "t-sales")
	if err != nil {
		t.Fatalf("linked user: %v", err)
	}
	if info.UserID != "sales" || info.LineUserID != "U1" || info.LineDisplayName != "LINE U1" || len(info.Roles) != 1 {
		t.Fatalf("linked user: got %+v", info)
	}

	for token, want := range map[string]string{
		"t-bad":      "INVALID_LINE_TOKEN",
		"t-stranger": "LINE_NOT_LINKED",
		"t-off":      "LINE_NOT_LINKED",
		"t-temp":     "PASSWORD_CHANGE_REQUIRED",
	} {
		_, err := u.ResolveLIFFSession(context.Background(), token)
		var appErr *cerrors.AppError
		if !errors.As(err, &appErr) || appErr.Code != want {
			t.Errorf("%s: got %v, want %s", token, err, want)
		}
	}
}
