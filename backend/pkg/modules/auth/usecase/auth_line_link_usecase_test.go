package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/config"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/domain"
	"backend/pkg/modules/auth/models"
	"backend/pkg/registry"
	"backend/pkg/repository"
)

type fakeUsers struct {
	repository.AppUserRepository
	users   []domain.AppUser
	updated []*domain.AppUser
}

func (f *fakeUsers) FindByLineUserID(_ context.Context, lineUserID string) (*domain.AppUser, error) {
	for i := range f.users {
		if f.users[i].LineUserID.String == lineUserID {
			return &f.users[i], nil
		}
	}
	return nil, nil
}

func (f *fakeUsers) Update(_ context.Context, obj *domain.AppUser, _ ...string) (*domain.AppUser, error) {
	f.updated = append(f.updated, obj)
	return obj, nil
}

func TestLineLink(t *testing.T) {
	users := &fakeUsers{users: []domain.AppUser{
		{ID: "me", Status: true, LineUserID: null.StringFrom("U-mine")},
		{ID: "other", Status: true, LineUserID: null.StringFrom("U-taken")},
	}}
	u := &authUseCase{
		coreRegistry: &registry.CoreRegistry{
			Logger: zap.NewNop(),
			Config: &config.Config{CORSAllowedOrigins: []string{"https://admin.example.com"}},
		},
		shareRegistry: &registry.ShareRegistry{ClientRegistry: registry.ClientRegistry{LINE: &fakeLINELogin{codes: map[string]string{
			"c-new": "U-new", "c-mine": "U-mine", "c-taken": "U-taken",
		}}}},
		repositoryRegistry: &registry.RepositoryRegistry{AppUserRepository: users},
	}
	const callback = "https://admin.example.com/login/line/callback"
	request := func(code, redirect string) *models.AuthLineLinkRequest {
		return &models.AuthLineLinkRequest{
			Code:            code,
			RedirectURI:     redirect,
			CodeVerifier:    "v123456789012345678901234567890123456789012345",
			Nonce:           "n",
			UserRequestInfo: &coremodels.UserRequestInfo{UserID: "me"},
		}
	}

	for _, code := range []string{"c-new", "c-mine"} {
		res, err := u.LineLink(context.Background(), request(code, callback))
		if err != nil {
			t.Fatalf("%s: %v", code, err)
		}
		last := users.updated[len(users.updated)-1]
		if last.ID != "me" || !last.LineUserID.Valid || res.Line.LinkedAt.IsZero() {
			t.Fatalf("%s: updated %+v, response %+v", code, last, res)
		}
	}

	for name, tc := range map[string]struct {
		req  *models.AuthLineLinkRequest
		want string
	}{
		"linked to another user": {request("c-taken", callback), "LINE_ALREADY_LINKED"},
		"code refused":           {request("c-bad", callback), "INVALID_LINE_LOGIN"},
		"foreign origin":         {request("c-new", "https://evil.example.com/login/line/callback"), "VALIDATION"},
		"not signed in":          {&models.AuthLineLinkRequest{Code: "c-new", RedirectURI: callback}, "UNAUTHENTICATED"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := u.LineLink(context.Background(), tc.req)
			var appErr *cerrors.AppError
			if !errors.As(err, &appErr) || appErr.Code != tc.want {
				t.Fatalf("got %v, want %s", err, tc.want)
			}
		})
	}
}
