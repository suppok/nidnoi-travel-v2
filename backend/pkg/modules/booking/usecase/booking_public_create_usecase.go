package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/bookingstate"
	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/tourprogramstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/booking/models"
)

const (
	maxTravellers   = 50
	maxNameRunes    = 100
	maxEmailRunes   = 254
	maxNoteRunes    = 1000
	minPhoneDigits  = 9
	maxPhoneDigits  = 15
	unavailableNote = "this period has no such price"

	// Booking number collisions: retries, and the unique index that reports one.
	maxNoAttempts        = 3
	bookingNoUniqueIndex = "booking_booking_no_key"
)

// isBookingNoTaken reports a collision on the booking number's unique index (a variable so tests
// can simulate one without Postgres).
var isBookingNoTaken = func(err error) bool {
	return query.UniqueViolationConstraint(err) == bookingNoUniqueIndex
}

var (
	publicCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{1,32}$`)
	// Digits with optional leading +, spaces and hyphens: "081-234-5678", "+66 81 234 5678".
	phonePattern = regexp.MustCompile(`^\+?[0-9][0-9 -]*[0-9]$`)
)

func phoneRule(value any) error {
	s, _ := value.(string)
	if s == "" {
		return nil
	}
	digits := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			digits++
		}
	}
	if !phonePattern.MatchString(s) || digits < minPhoneDigits || digits > maxPhoneDigits {
		return fmt.Errorf("must be a phone number of %d-%d digits", minPhoneDigits, maxPhoneDigits)
	}
	return nil
}

func normalizePublicCreate(r *models.BookingPublicCreateRequest) {
	r.TourCode = strings.TrimSpace(r.TourCode)
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.Phone = strings.TrimSpace(r.Phone)
	r.Note = strings.TrimSpace(r.Note)
}

func (u *bookingUseCase) PublicCreateValidate(request *models.BookingPublicCreateRequest) error {
	count := []validation.Rule{validation.Min(0), validation.Max(maxTravellers)}
	err := validation.ValidateStruct(request,
		validation.Field(&request.PeriodID, validation.Required, validation.UUID),
		validation.Field(&request.AdultTwin, count...),
		validation.Field(&request.AdultSingle, count...),
		validation.Field(&request.ChildWithBed, count...),
		validation.Field(&request.ChildNoBed, count...),
		validation.Field(&request.FirstName, validation.Required, validation.RuneLength(1, maxNameRunes)),
		validation.Field(&request.LastName, validation.RuneLength(0, maxNameRunes)),
		validation.Field(&request.Email, validation.Required, validation.RuneLength(1, maxEmailRunes), validation.Email),
		validation.Field(&request.Phone, validation.Required, validation.By(phoneRule)),
		validation.Field(&request.Note, validation.RuneLength(0, maxNoteRunes)),
	)
	if err != nil {
		return err
	}
	switch adults, total := request.AdultTwin+request.AdultSingle, request.AdultTwin+request.AdultSingle+request.ChildWithBed+request.ChildNoBed; {
	case adults < 1:
		return validation.Errors{"adultTwin": errors.New("at least one adult is required")}
	case total > maxTravellers:
		return validation.Errors{"adultTwin": fmt.Errorf("at most %d travellers per booking", maxTravellers)}
	}
	return nil
}

// quote prices the request from the period. A count for a price row the period doesn't offer is a
// field error.
func quote(p *domain.TourProgramPeriod, r *models.BookingPublicCreateRequest) (*domain.Booking, error) {
	single := null.Int{}
	if p.SingleSupplement.Valid {
		single = null.IntFrom(int64(p.Price) + p.SingleSupplement.Int64)
	}
	fields := validation.Errors{}
	for name, row := range map[string]struct {
		count int
		price null.Int
	}{
		"adultSingle":  {r.AdultSingle, single},
		"childWithBed": {r.ChildWithBed, p.ChildWithBedPrice},
		"childNoBed":   {r.ChildNoBed, p.ChildNoBedPrice},
	} {
		if row.count > 0 && !row.price.Valid {
			fields[name] = errors.New(unavailableNote)
		}
	}
	if len(fields) > 0 {
		return nil, fields
	}

	total := r.AdultTwin*p.Price +
		r.AdultSingle*int(single.Int64) +
		r.ChildWithBed*int(p.ChildWithBedPrice.Int64) +
		r.ChildNoBed*int(p.ChildNoBedPrice.Int64)
	return &domain.Booking{
		AdultTwinCount:    r.AdultTwin,
		AdultTwinPrice:    p.Price,
		AdultSingleCount:  r.AdultSingle,
		AdultSinglePrice:  single,
		ChildWithBedCount: r.ChildWithBed,
		ChildWithBedPrice: p.ChildWithBedPrice,
		ChildNoBedCount:   r.ChildNoBed,
		ChildNoBedPrice:   p.ChildNoBedPrice,
		TotalPrice:        total,
	}, nil
}

// bookable reports whether customers may still send a booking for the period: the tour is selling,
// and the period is active, upcoming and not sold out.
func bookable(tour *domain.TourProgram, p *domain.TourProgramPeriod, today time.Time) bool {
	return p != nil &&
		p.TourProgramID == tour.ID &&
		tour.State == tourprogramstate.Selling &&
		p.Status &&
		p.StartDate.After(today) &&
		p.SeatsAvailable > 0
}

// PublicCreate records a booking request from the customer site. It reserves no seats: sales
// contacts the customer to confirm.
func (u *bookingUseCase) PublicCreate(ctx context.Context, request *models.BookingPublicCreateRequest) (*models.BookingPublicCreateResponse, error) {
	normalizePublicCreate(request)
	if request.Website != "" {
		// Honeypot filled in: answer like a success (with a number that is never stored) so the bot
		// moves on.
		u.coreRegistry.Logger.Info("booking honeypot triggered", zap.String("tourCode", request.TourCode))
		fake, _ := newBookingNo(time.Now())
		return &models.BookingPublicCreateResponse{BookingNo: fake}, nil
	}
	if !publicCodePattern.MatchString(request.TourCode) {
		return nil, cerrors.NotFound()
	}
	if err := u.PublicCreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	// A random booking number can collide with an existing one (unlikely: ~887 million values). The
	// unique violation aborts the transaction, so retry the whole transaction with a new number.
	var (
		created *domain.Booking
		err     error
	)
	for attempt := 1; ; attempt++ {
		created, err = u.createBooking(ctx, request)
		if attempt < maxNoAttempts && isBookingNoTaken(err) {
			continue
		}
		break
	}
	if err != nil {
		return nil, cerrors.Wrap(err, "create booking")
	}

	// No personal data in logs (PDPA).
	u.coreRegistry.Logger.Info("booking received", zap.String("bookingNo", created.BookingNo), zap.String("tourCode", created.TourCode))
	return &models.BookingPublicCreateResponse{BookingNo: created.BookingNo, TotalPrice: created.TotalPrice}, nil
}

// newBookingNo returns the 2-digit year in Thailand plus random characters, e.g. 26X8K9AQ.
func newBookingNo(now time.Time) (string, error) {
	random, err := utils.RandomString(bookingstate.NoLength, bookingstate.NoAlphabet)
	if err != nil {
		return "", err
	}
	return now.In(utils.Thailand).Format("06") + random, nil
}

// createBooking checks the period, quotes it and inserts the booking in one transaction.
func (u *bookingUseCase) createBooking(ctx context.Context, request *models.BookingPublicCreateRequest) (*domain.Booking, error) {
	no, err := newBookingNo(time.Now())
	if err != nil {
		return nil, err
	}

	var created *domain.Booking
	err = u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		tour, err := u.repositoryRegistry.TourProgramRepository.WithTx(appDB).FindPublicStateByCode(ctx, request.TourCode)
		if err != nil {
			return err
		}
		if tour == nil {
			return cerrors.NotFound()
		}
		period, err := u.repositoryRegistry.TourProgramPeriodRepository.WithTx(appDB).FindByID(ctx, request.PeriodID)
		if err != nil {
			return err
		}
		if !bookable(tour, period, utils.TodayInThailand(time.Now())) {
			return cerrors.Conflict(errorcode.PeriodUnavailable)
		}
		b, err := quote(period, request)
		if err != nil {
			return cerrors.Validation(err)
		}

		now := time.Now()
		b.ID = utils.NewID()
		b.BookingNo = no
		b.TourProgramID = tour.ID
		b.TourProgramPeriodID = period.ID
		b.TourCode = tour.Code
		b.TourName = tour.Name
		b.StartDate = period.StartDate
		b.EndDate = period.EndDate
		b.FirstName = request.FirstName
		b.LastName = request.LastName
		b.Email = request.Email
		b.Phone = request.Phone
		b.CustomerNote = request.Note
		b.State = bookingstate.New
		b.Status = true
		b.CreateDate = now
		b.UpdateDate = now
		b.CreateBy = bookingstate.SystemUserID
		b.UpdateBy = bookingstate.SystemUserID
		created, err = u.repositoryRegistry.BookingRepository.WithTx(appDB).Create(ctx, b)
		return err
	})
	return created, err
}
