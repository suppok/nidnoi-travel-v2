package usecase

import (
	"testing"

	"github.com/guregu/null/v5"

	"backend/pkg/modules/linebooking/models"
)

func TestUpdateLineContactValidate(t *testing.T) {
	const id = "0f8fad5b-d9cb-469f-a165-70867728950e"
	u := &lineBookingUseCase{}
	cases := []struct {
		name    string
		request models.LineBookingUpdateLineContactRequest
		ok      bool
	}{
		{"link", models.LineBookingUpdateLineContactRequest{ID: id, LineContactID: null.StringFrom(id)}, true},
		{"unlink", models.LineBookingUpdateLineContactRequest{ID: id}, true},
		{"bad booking id", models.LineBookingUpdateLineContactRequest{ID: "x", LineContactID: null.StringFrom(id)}, false},
		{"bad contact id", models.LineBookingUpdateLineContactRequest{ID: id, LineContactID: null.StringFrom("x")}, false},
		{"empty contact id", models.LineBookingUpdateLineContactRequest{ID: id, LineContactID: null.StringFrom("")}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := u.UpdateLineContactValidate(&tc.request); (err == nil) != tc.ok {
				t.Fatalf("UpdateLineContactValidate() = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}
