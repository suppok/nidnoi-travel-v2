package usecase

import (
	"testing"

	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

func TestWholesaleEmailDecision(t *testing.T) {
	cases := []struct {
		name string
		w    *domain.Wholesale
		want string
	}{
		{"no wholesale", nil, models.WholesaleEmailDisabled},
		{"auto email off", &domain.Wholesale{Emails: []string{"a@x.com"}}, models.WholesaleEmailDisabled},
		{"auto email on, no address", &domain.Wholesale{AutoEmail: true, Emails: []string{}}, models.WholesaleEmailNoEmail},
		{"auto email on with address", &domain.Wholesale{AutoEmail: true, Emails: []string{"a@x.com"}}, models.WholesaleEmailQueued},
	}
	for _, c := range cases {
		if got := wholesaleEmailDecision(c.w); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}
