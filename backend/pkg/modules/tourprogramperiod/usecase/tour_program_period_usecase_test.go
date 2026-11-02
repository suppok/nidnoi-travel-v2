package usecase

import (
	"errors"
	"testing"

	"github.com/guregu/null/v5"

	"backend/pkg/core/validation"
	"backend/pkg/modules/tourprogramperiod/models"
)

func valid() models.TourProgramPeriodFields {
	return models.TourProgramPeriodFields{
		StartDate: "2026-12-01", EndDate: "2026-12-05",
		Price: 29999, SeatsTotal: 30, SeatsAvailable: 12,
	}
}

func fieldErr(t *testing.T, err error) map[string]bool {
	t.Helper()
	var verrs validation.Errors
	if !errors.As(err, &verrs) {
		t.Fatalf("want validation.Errors, got %v", err)
	}
	out := map[string]bool{}
	for k := range verrs {
		out[k] = true
	}
	return out
}

func TestPeriodFieldsValidate(t *testing.T) {
	f := valid()
	if err := fieldsValidate(&f); err != nil {
		t.Fatalf("valid period rejected: %v", err)
	}

	cases := map[string]func(f *models.TourProgramPeriodFields){
		"startDate":           func(f *models.TourProgramPeriodFields) { f.StartDate = "01/12/2026" },
		"endDate":             func(f *models.TourProgramPeriodFields) { f.EndDate = "2026-11-30" },
		"seatsAvailable":      func(f *models.TourProgramPeriodFields) { f.SeatsAvailable = 31 },
		"price":               func(f *models.TourProgramPeriodFields) { f.Price = -1 },
		"priceBeforeDiscount": func(f *models.TourProgramPeriodFields) { f.PriceBeforeDiscount = null.IntFrom(29999) },
		"deposit":             func(f *models.TourProgramPeriodFields) { f.Deposit = null.IntFrom(-5) },
	}
	for field, mutate := range cases {
		f := valid()
		mutate(&f)
		if got := fieldErr(t, fieldsValidate(&f)); !got[field] {
			t.Errorf("%s: want error on %s, got %v", field, field, got)
		}
	}

	long := valid()
	long.EndDate = "2027-04-01"
	if got := fieldErr(t, fieldsValidate(&long)); !got["endDate"] {
		t.Errorf("span over 90 days accepted")
	}
}
