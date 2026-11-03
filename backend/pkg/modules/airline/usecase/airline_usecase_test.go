package usecase

import (
	"context"
	"testing"

	"github.com/guregu/null/v5"

	"backend/pkg/modules/airline/models"
	"backend/pkg/modules/shared/client/objectstorage"
	"backend/pkg/registry"
)

type fakeStorage struct{ base string }

func (f fakeStorage) Put(context.Context, objectstorage.PutInput) error { return nil }
func (f fakeStorage) Delete(context.Context, string) error              { return nil }
func (f fakeStorage) PublicURL(key string) string                       { return f.base + "/" + key }
func (f fakeStorage) PublicBaseURL() string                             { return f.base }

func newTestUseCase(base string) *airlineUseCase {
	return &airlineUseCase{shareRegistry: &registry.ShareRegistry{
		ClientRegistry: registry.ClientRegistry{ObjectStorage: fakeStorage{base: base}},
	}}
}

func TestCreateValidate(t *testing.T) {
	const base = "https://media.example.com"
	u := newTestUseCase(base)

	cases := []struct {
		name    string
		code    string
		logo    null.String
		wantErr bool
	}{
		{"iata", "TG", null.String{}, false},
		{"iata with digit", "3K", null.String{}, false},
		{"icao", "THA", null.String{}, false},
		{"lowercase normalized", " fd ", null.String{}, false},
		{"too short", "T", null.String{}, true},
		{"too long", "THAI", null.String{}, true},
		{"hyphen", "T-G", null.String{}, true},
		{"uploaded logo", "TG", null.StringFrom(base + "/airlines/logos/abc123.png"), false},
		{"empty logo becomes null", "TG", null.StringFrom("  "), false},
		{"foreign logo", "TG", null.StringFrom("https://evil.example.com/airlines/logos/a.png"), true},
		{"wrong prefix", "TG", null.StringFrom(base + "/tours/banners/a.png"), true},
		{"nested path", "TG", null.StringFrom(base + "/airlines/logos/x/../a.png"), true},
		{"query string", "TG", null.StringFrom(base + "/airlines/logos/a.png?x=1"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &models.AirlineCreateRequest{Code: tc.code, Name: "Thai Airways", Logo: tc.logo}
			normalize(&req.Code, &req.Name, &req.Logo)
			err := u.CreateValidate(req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CreateValidate(code=%q, logo=%v) error = %v, wantErr %v", tc.code, tc.logo, err, tc.wantErr)
			}
		})
	}
}

func TestLogoRequiresStorage(t *testing.T) {
	u := newTestUseCase("")
	req := &models.AirlineCreateRequest{Code: "TG", Name: "Thai Airways", Logo: null.StringFrom("https://x/airlines/logos/a.png")}
	if err := u.CreateValidate(req); err == nil {
		t.Fatal("expected an error when R2_PUBLIC_URL is not set")
	}
}
