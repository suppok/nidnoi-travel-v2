package usecase

import (
	"regexp"
	"testing"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/domain"
	"backend/pkg/modules/booking/models"
)

func validRequest() *models.BookingPublicCreateRequest {
	return &models.BookingPublicCreateRequest{
		TourCode:  "NIDN000001",
		PeriodID:  "6f1c2a52-0000-4000-8000-000000000001",
		AdultTwin: 2,
		FirstName: "สมชาย",
		Email:     "somchai@example.com",
		Phone:     "081-234-5678",
	}
}

func TestPublicCreateValidate(t *testing.T) {
	u := &bookingUseCase{}
	cases := []struct {
		name    string
		edit    func(r *models.BookingPublicCreateRequest)
		wantErr bool
	}{
		{"valid", func(*models.BookingPublicCreateRequest) {}, false},
		{"single adult only", func(r *models.BookingPublicCreateRequest) { r.AdultTwin, r.AdultSingle = 0, 1 }, false},
		{"no adult", func(r *models.BookingPublicCreateRequest) { r.AdultTwin, r.ChildNoBed = 0, 2 }, true},
		{"too many travellers", func(r *models.BookingPublicCreateRequest) { r.AdultTwin, r.ChildWithBed = 30, 21 }, true},
		{"negative count", func(r *models.BookingPublicCreateRequest) { r.ChildNoBed = -1 }, true},
		{"missing first name", func(r *models.BookingPublicCreateRequest) { r.FirstName = "" }, true},
		{"bad email", func(r *models.BookingPublicCreateRequest) { r.Email = "somchai@" }, true},
		{"missing phone", func(r *models.BookingPublicCreateRequest) { r.Phone = "" }, true},
		{"intl phone", func(r *models.BookingPublicCreateRequest) { r.Phone = "+66 81 234 5678" }, false},
		{"short phone", func(r *models.BookingPublicCreateRequest) { r.Phone = "1234" }, true},
		{"phone with letters", func(r *models.BookingPublicCreateRequest) { r.Phone = "081-CALL-ME" }, true},
		{"bad period id", func(r *models.BookingPublicCreateRequest) { r.PeriodID = "x" }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := validRequest()
			tc.edit(r)
			if err := u.PublicCreateValidate(r); (err != nil) != tc.wantErr {
				t.Fatalf("PublicCreateValidate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestNewBookingNo(t *testing.T) {
	// 2025-12-31 18:00 UTC is already 2026-01-01 in Thailand.
	no, err := newBookingNo(time.Date(2025, 12, 31, 18, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^26[2-9A-HJKMNP-Z]{6}$`).MatchString(no) {
		t.Fatalf("newBookingNo() = %q, want 26 + 6 characters without 0/O/1/I/L", no)
	}
}

func TestQuote(t *testing.T) {
	period := &domain.TourProgramPeriod{
		Price:             16919,
		SingleSupplement:  null.IntFrom(4900),
		ChildWithBedPrice: null.IntFrom(15919),
	}

	r := validRequest()
	r.AdultTwin, r.AdultSingle, r.ChildWithBed = 3, 1, 1
	b, err := quote(period, r)
	if err != nil {
		t.Fatalf("quote() error = %v", err)
	}
	if b.AdultSinglePrice.Int64 != 21819 {
		t.Errorf("single price = %d, want 21819", b.AdultSinglePrice.Int64)
	}
	if want := 3*16919 + 21819 + 15919; b.TotalPrice != want {
		t.Errorf("total = %d, want %d", b.TotalPrice, want)
	}
	if b.ChildNoBedPrice.Valid {
		t.Errorf("child-no-bed price = %v, want null", b.ChildNoBedPrice)
	}

	r = validRequest()
	r.ChildNoBed = 1 // the period has no child-no-bed price
	if _, err := quote(period, r); err == nil {
		t.Fatal("quote() with an unpriced row: want error")
	}

	period.SingleSupplement = null.Int{}
	r = validRequest()
	r.AdultSingle = 1
	if _, err := quote(period, r); err == nil {
		t.Fatal("quote() single room without a supplement: want error")
	}
}

func TestBookable(t *testing.T) {
	today := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	tour := &domain.TourProgram{ID: "t1", State: tourprogramstate.Selling}
	open := func() *domain.TourProgramPeriod {
		return &domain.TourProgramPeriod{TourProgramID: "t1", Status: true, StartDate: today.AddDate(0, 0, 10), SeatsAvailable: 5}
	}

	if !bookable(tour, open(), today) {
		t.Fatal("open period: want bookable")
	}
	cases := map[string]func(p *domain.TourProgramPeriod, tp *domain.TourProgram){
		"other tour":    func(p *domain.TourProgramPeriod, _ *domain.TourProgram) { p.TourProgramID = "t2" },
		"deleted":       func(p *domain.TourProgramPeriod, _ *domain.TourProgram) { p.Status = false },
		"departs today": func(p *domain.TourProgramPeriod, _ *domain.TourProgram) { p.StartDate = today },
		"sold out":      func(p *domain.TourProgramPeriod, _ *domain.TourProgram) { p.SeatsAvailable = 0 },
		"tour closed":   func(_ *domain.TourProgramPeriod, tp *domain.TourProgram) { tp.State = tourprogramstate.Closed },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			p, tp := open(), *tour
			edit(p, &tp)
			if bookable(&tp, p, today) {
				t.Fatal("want not bookable")
			}
		})
	}
	if bookable(tour, nil, today) {
		t.Fatal("missing period: want not bookable")
	}
}
