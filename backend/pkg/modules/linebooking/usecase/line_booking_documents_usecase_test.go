package usecase

import (
	"testing"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

func TestDocumentsForPlan(t *testing.T) {
	const pdf = "https://media.test/bookings/documents/abc/x.pdf"
	deposit := &domain.LineBooking{IsDeposit: true}
	full := &domain.LineBooking{}

	tests := []struct {
		name string
		b    *domain.LineBooking
		r    models.LineBookingDocumentsRequest
		ok   bool
	}{
		{"deposit AI+IV", deposit, models.LineBookingDocumentsRequest{AIFile: pdf, IVFile: pdf}, true},
		{"deposit missing IV", deposit, models.LineBookingDocumentsRequest{AIFile: pdf}, false},
		{"deposit with RE", deposit, models.LineBookingDocumentsRequest{AIFile: pdf, IVFile: pdf, REFile: pdf}, false},
		{"full RE", full, models.LineBookingDocumentsRequest{REFile: pdf}, true},
		{"full missing RE", full, models.LineBookingDocumentsRequest{}, false},
		{"full with AI", full, models.LineBookingDocumentsRequest{AIFile: pdf, REFile: pdf}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := documentsForPlan(tt.b, &tt.r); (err == nil) != tt.ok {
				t.Fatalf("documentsForPlan err = %v, want ok %v", err, tt.ok)
			}
		})
	}
}

func TestSendsDocuments(t *testing.T) {
	const me, other = "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"
	sentToMe := &domain.LineBooking{SalesUserID: null.StringFrom(other), FinanceUserID: null.StringFrom(me)}
	sentToOther := &domain.LineBooking{SalesUserID: null.StringFrom(me), FinanceUserID: null.StringFrom(other)}

	tests := []struct {
		name  string
		roles []string
		b     *domain.LineBooking
		want  bool
	}{
		{"its finance user", []string{role.Finance}, sentToMe, true},
		{"another finance user", []string{role.Finance}, sentToOther, false},
		{"sales owner", []string{role.Sales}, sentToOther, false},
		{"admin", []string{role.Admin}, sentToOther, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &coremodels.UserRequestInfo{UserID: me, Roles: tt.roles}
			if got := sendsDocuments(tt.b, info); got != tt.want {
				t.Fatalf("sendsDocuments = %v, want %v", got, tt.want)
			}
		})
	}
}
