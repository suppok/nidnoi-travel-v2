package usecase

import (
	"testing"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
)

func TestVisibleTo(t *testing.T) {
	const me, other = "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"
	mine := &domain.LineBooking{SalesUserID: null.StringFrom(me)}
	theirs := &domain.LineBooking{SalesUserID: null.StringFrom(other)}
	noOwner := &domain.LineBooking{}
	sentToMe := &domain.LineBooking{SalesUserID: null.StringFrom(other), FinanceUserID: null.StringFrom(me)}
	sentToOther := &domain.LineBooking{SalesUserID: null.StringFrom(me), FinanceUserID: null.StringFrom(other)}

	tests := []struct {
		name  string
		roles []string
		b     *domain.LineBooking
		want  bool
	}{
		{"sales sees own", []string{role.Sales}, mine, true},
		{"sales hidden others", []string{role.Sales}, theirs, false},
		{"sales hidden ownerless", []string{role.Sales}, noOwner, false},
		{"admin sees all", []string{role.Admin}, theirs, true},
		{"super admin sees all", []string{role.SuperAdmin}, noOwner, true},
		{"finance sees sent to them", []string{role.Finance}, sentToMe, true},
		{"finance hidden others", []string{role.Finance}, theirs, false},
		{"finance hidden sent to others", []string{role.Finance}, &domain.LineBooking{FinanceUserID: null.StringFrom(other)}, false},
		{"sales sees own sent to other finance", []string{role.Sales}, sentToOther, true},
		{"sales and finance sees both", []string{role.Sales, role.Finance}, sentToMe, true},
		{"sales and admin sees all", []string{role.Sales, role.Admin}, theirs, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &coremodels.UserRequestInfo{UserID: me, Roles: tt.roles}
			if got := visibleTo(tt.b, info); got != tt.want {
				t.Fatalf("visibleTo = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOwnsAsSales(t *testing.T) {
	const me, other = "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"
	sentToMe := &domain.LineBooking{SalesUserID: null.StringFrom(other), FinanceUserID: null.StringFrom(me)}
	mine := &domain.LineBooking{SalesUserID: null.StringFrom(me)}

	tests := []struct {
		name  string
		roles []string
		b     *domain.LineBooking
		want  bool
	}{
		{"owner", []string{role.Sales}, mine, true},
		{"finance may not confirm", []string{role.Finance}, sentToMe, false},
		{"sales+finance not owner", []string{role.Sales, role.Finance}, sentToMe, false},
		{"admin", []string{role.Admin}, sentToMe, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &coremodels.UserRequestInfo{UserID: me, Roles: tt.roles}
			if got := ownsAsSales(tt.b, info); got != tt.want {
				t.Fatalf("ownsAsSales = %v, want %v", got, tt.want)
			}
		})
	}
}
