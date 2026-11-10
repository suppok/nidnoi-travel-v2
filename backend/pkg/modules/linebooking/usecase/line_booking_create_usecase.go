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
	"backend/pkg/constants/linebookingstate"
	"backend/pkg/constants/tourprogramstate"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

const (
	maxTravellers            = 50
	maxNameRunes             = 200
	maxNoteRunes             = 2000
	minPhoneDigits           = 9
	maxPhoneDigits           = 15
	maxNoAttempts            = 3
	lineBookingNoUniqueIndex = "line_booking_booking_no_key"
)

var (
	publicCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{1,32}$`)
	phonePattern      = regexp.MustCompile(`^\+?[0-9][0-9 -]*[0-9]$`)
)

var isLineBookingNoTaken = func(err error) bool {
	return query.UniqueViolationConstraint(err) == lineBookingNoUniqueIndex
}

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

func normalizeCreate(r *models.LineBookingCreateRequest) {
	r.SalesUserID = strings.TrimSpace(r.SalesUserID)
	r.TourCode = strings.TrimSpace(r.TourCode)
	r.CustomerName = strings.TrimSpace(r.CustomerName)
	r.CustomerPhone = strings.TrimSpace(r.CustomerPhone)
	r.SlipImage = strings.TrimSpace(r.SlipImage)
	r.Note = strings.TrimSpace(r.Note)
	if r.LineContactID.Valid {
		id := strings.TrimSpace(r.LineContactID.String)
		r.LineContactID = null.NewString(id, id != "")
	}
}

func (u *lineBookingUseCase) CreateValidate(request *models.LineBookingCreateRequest) error {
	count := []validation.Rule{validation.Min(0), validation.Max(maxTravellers)}
	err := validation.ValidateStruct(request,
		validation.Field(&request.SalesUserID, validation.Required, validation.UUID),
		validation.Field(&request.TourCode, validation.Required),
		validation.Field(&request.PeriodID, validation.Required, validation.UUID),
		validation.Field(&request.AdultTwin, count...),
		validation.Field(&request.AdultSingle, count...),
		validation.Field(&request.ChildWithBed, count...),
		validation.Field(&request.ChildNoBed, count...),
		validation.Field(&request.CustomerName, validation.Required, validation.RuneLength(1, maxNameRunes)),
		validation.Field(&request.CustomerPhone, validation.Required, validation.By(phoneRule)),
		validation.Field(&request.SlipImage, validation.Required),
		validation.Field(&request.Note, validation.RuneLength(0, maxNoteRunes)),
	)
	if err != nil {
		return err
	}
	if request.LineContactID.Valid {
		if err := validation.Validate(request.LineContactID.String, validation.UUID); err != nil {
			return validation.Errors{"lineContactId": err}
		}
	}
	switch adults, total := request.AdultTwin+request.AdultSingle, request.AdultTwin+request.AdultSingle+request.ChildWithBed+request.ChildNoBed; {
	case adults < 1:
		return validation.Errors{"adultTwin": errors.New("at least one adult is required")}
	case total > maxTravellers:
		return validation.Errors{"adultTwin": fmt.Errorf("at most %d travellers per booking", maxTravellers)}
	}
	return nil
}

// quoteLine prices the request from the period (same logic as customer booking).
func quoteLine(p *domain.TourProgramPeriod, r *models.LineBookingCreateRequest) (*domain.LineBooking, error) {
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
			fields[name] = errors.New("this period has no such price")
		}
	}
	if len(fields) > 0 {
		return nil, fields
	}

	total := r.AdultTwin*p.Price +
		r.AdultSingle*int(single.Int64) +
		r.ChildWithBed*int(p.ChildWithBedPrice.Int64) +
		r.ChildNoBed*int(p.ChildNoBedPrice.Int64)
	return &domain.LineBooking{
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

// bookable checks the period is still open for booking.
func bookable(tour *domain.TourProgram, p *domain.TourProgramPeriod, today time.Time) bool {
	return p != nil &&
		p.TourProgramID == tour.ID &&
		tour.State == tourprogramstate.Selling &&
		p.Status &&
		p.StartDate.After(today) &&
		p.SeatsAvailable > 0
}

func newBookingNo(now time.Time) (string, error) {
	random, err := utils.RandomString(bookingstate.NoLength, bookingstate.NoAlphabet)
	if err != nil {
		return "", err
	}
	return now.In(utils.Thailand).Format("06") + random, nil
}

// Create records a LINE booking from the LIFF app. The LIFF middleware has already verified the
// token and resolved the linked staff user. It validates the tour/period, quotes prices and inserts
// the row. No seats are reserved.
func (u *lineBookingUseCase) Create(ctx context.Context, request *models.LineBookingCreateRequest) (*models.LineBookingCreateResponse, error) {
	normalizeCreate(request)
	if !publicCodePattern.MatchString(request.TourCode) {
		return nil, cerrors.NotFound()
	}
	if err := u.CreateValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}

	info := request.UserRequestInfo
	if info == nil || info.LineUserID == "" {
		return nil, cerrors.Unauthenticated()
	}

	var (
		created *domain.LineBooking
		txErr   error
	)
	for attempt := 1; ; attempt++ {
		created, txErr = u.createLineBooking(ctx, request, info)
		if attempt < maxNoAttempts && isLineBookingNoTaken(txErr) {
			continue
		}
		break
	}
	if txErr != nil {
		return nil, txErr
	}

	u.coreRegistry.Logger.Info("line booking received",
		zap.String("bookingNo", created.BookingNo),
		zap.String("tourCode", created.TourCode),
		zap.String("lineUserId", created.LineUserID),
		zap.String("userId", info.UserID))
	return &models.LineBookingCreateResponse{BookingNo: created.BookingNo, TotalPrice: created.TotalPrice}, nil
}

func (u *lineBookingUseCase) createLineBooking(ctx context.Context, request *models.LineBookingCreateRequest, info *coremodels.UserRequestInfo) (*domain.LineBooking, error) {
	no, err := newBookingNo(time.Now())
	if err != nil {
		return nil, err
	}

	var created *domain.LineBooking
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
		b, err := quoteLine(period, request)
		if err != nil {
			return cerrors.Validation(err)
		}
		if err := u.checkSalesUser(ctx, appDB, request.SalesUserID); err != nil {
			return err
		}
		b.SalesUserID = null.StringFrom(request.SalesUserID)
		if request.LineContactID.Valid {
			contact, err := u.repositoryRegistry.LineContactRepository.WithTx(appDB).FindByID(ctx, request.LineContactID.String)
			if err != nil {
				return err
			}
			if contact == nil || !contact.Status {
				return cerrors.FieldError("lineContactId", "no such LINE contact")
			}
			b.LineContactID = request.LineContactID
		}

		now := time.Now()
		b.ID = utils.NewID()
		b.BookingNo = no
		b.LineUserID = info.LineUserID
		b.LineDisplayName = info.LineDisplayName
		b.TourProgramID = tour.ID
		b.TourProgramPeriodID = period.ID
		b.TourCode = tour.Code
		b.TourName = tour.Name
		b.StartDate = period.StartDate
		b.EndDate = period.EndDate
		b.CustomerName = request.CustomerName
		b.CustomerPhone = request.CustomerPhone
		b.SlipImage = request.SlipImage
		b.Note = request.Note
		b.State = linebookingstate.New
		b.Status = true
		b.CreateDate = now
		b.UpdateDate = now
		b.CreateBy = info.UserID
		b.UpdateBy = info.UserID
		created, err = u.repositoryRegistry.LineBookingRepository.WithTx(appDB).Create(ctx, b)
		return err
	})
	return created, err
}
