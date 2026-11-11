package usecase

import (
	"strings"
	"testing"

	"backend/pkg/modules/bookingsource/models"
)

func TestCreateValidate(t *testing.T) {
	u := &bookingSourceUseCase{}

	cases := []struct {
		name      string
		srcName   string
		shortName string
		sortOrder int
		wantErr   bool
	}{
		{"name only", "LINE OA", "", 0, false},
		{"with short name", "Agent", "agent", 10, false},
		{"blank name", "   ", "", 0, true},
		{"name too long", strings.Repeat("ก", 101), "", 0, true},
		{"short name too long", "Agent", strings.Repeat("a", 51), 0, true},
		{"negative order", "Agent", "", -1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &models.BookingSourceCreateRequest{Name: tc.srcName, ShortName: tc.shortName, SortOrder: tc.sortOrder}
			normalize(&req.Name, &req.ShortName)
			err := u.CreateValidate(req)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CreateValidate(%+v) error = %v, wantErr %v", req, err, tc.wantErr)
			}
		})
	}
}
