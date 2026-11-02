package usecase

import (
	"errors"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/tourprogramperiod"
	"backend/pkg/modules/tourprogramperiod/models"
	"backend/pkg/registry"
)

const (
	maxPrice     = 10_000_000
	maxSeats     = 999
	maxSpanDays  = 90
	maxNoteRunes = 500
)

type tourProgramPeriodUseCase struct {
	coreRegistry       *registry.CoreRegistry
	shareRegistry      *registry.ShareRegistry
	repositoryRegistry *registry.RepositoryRegistry
}

func NewTourProgramPeriodUseCase(
	coreRegistry *registry.CoreRegistry,
	shareRegistry *registry.ShareRegistry,
	repositoryRegistry *registry.RepositoryRegistry,
) tourprogramperiod.TourProgramPeriodUseCase {
	return &tourProgramPeriodUseCase{
		coreRegistry:       coreRegistry,
		shareRegistry:      shareRegistry,
		repositoryRegistry: repositoryRegistry,
	}
}

func optionalPrice(value any) error {
	if v, ok := value.(null.Int); ok && v.Valid && (v.Int64 < 0 || v.Int64 > maxPrice) {
		return errors.New("must be between 0 and 10,000,000")
	}
	return nil
}

func dateRule(value any) error {
	s, _ := value.(string)
	if _, err := time.Parse(models.DateLayout, s); err != nil {
		return errors.New("must be a date in YYYY-MM-DD format")
	}
	return nil
}

// fieldsValidate holds the field rules shared by create and update.
func fieldsValidate(f *models.TourProgramPeriodFields) error {
	err := validation.ValidateStruct(f,
		validation.Field(&f.StartDate, validation.Required, validation.By(dateRule)),
		validation.Field(&f.EndDate, validation.Required, validation.By(dateRule)),
		validation.Field(&f.Price, validation.Min(0), validation.Max(maxPrice)),
		validation.Field(&f.PriceBeforeDiscount, validation.By(optionalPrice), validation.By(func(value any) error {
			if v, ok := value.(null.Int); ok && v.Valid && v.Int64 <= int64(f.Price) {
				return errors.New("must be higher than the price")
			}
			return nil
		})),
		validation.Field(&f.ChildWithBedPrice, validation.By(optionalPrice)),
		validation.Field(&f.ChildNoBedPrice, validation.By(optionalPrice)),
		validation.Field(&f.SingleSupplement, validation.By(optionalPrice)),
		validation.Field(&f.Deposit, validation.By(optionalPrice)),
		validation.Field(&f.SeatsTotal, validation.Min(0), validation.Max(maxSeats)),
		validation.Field(&f.SeatsAvailable, validation.Min(0), validation.Max(f.SeatsTotal).Error("must not be more than total seats")),
		validation.Field(&f.Note, validation.RuneLength(0, maxNoteRunes)),
	)
	if err != nil {
		return err
	}

	start, _ := time.Parse(models.DateLayout, f.StartDate)
	end, _ := time.Parse(models.DateLayout, f.EndDate)
	if end.Before(start) {
		return validation.Errors{"endDate": errors.New("must be on or after the start date")}
	}
	if end.Sub(start) > maxSpanDays*24*time.Hour {
		return validation.Errors{"endDate": errors.New("a period can last at most 90 days")}
	}
	return nil
}

// applyFields copies validated fields onto the row.
func applyFields(p *domain.TourProgramPeriod, f *models.TourProgramPeriodFields) {
	p.StartDate, _ = time.Parse(models.DateLayout, f.StartDate)
	p.EndDate, _ = time.Parse(models.DateLayout, f.EndDate)
	p.Price = f.Price
	p.PriceBeforeDiscount = f.PriceBeforeDiscount
	p.ChildWithBedPrice = f.ChildWithBedPrice
	p.ChildNoBedPrice = f.ChildNoBedPrice
	p.SingleSupplement = f.SingleSupplement
	p.Deposit = f.Deposit
	p.SeatsTotal = f.SeatsTotal
	p.SeatsAvailable = f.SeatsAvailable
	p.Note = strings.TrimSpace(f.Note)
}

var editableColumns = []string{
	"start_date", "end_date", "price", "price_before_discount", "child_with_bed_price", "child_no_bed_price",
	"single_supplement", "deposit", "seats_total", "seats_available", "note", "update_date", "update_by",
}

// Customer-site rebuild rules (backend/CLAUDE.md "When to enqueue"), applied once the revalidator
// exists: create / delete / date change → yes; price change → only if the tour's starting price
// changes; seat change → never.
