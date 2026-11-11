package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/bookingstate"
	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/linebookingpayment"
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
	maxAddressRunes          = 500
	maxRefRunes              = 200
	maxChargeNameRunes       = 100
	maxCharges               = 20
	maxUnitPrice             = 10_000_000
	maxCreditFeePercent      = 20.0
	minPhoneDigits           = 9
	maxPhoneDigits           = 15
	maxNoAttempts            = 3
	lineBookingNoUniqueIndex = "line_booking_booking_no_key"
)

var (
	publicCodePattern = regexp.MustCompile(`^[A-Za-z0-9]{1,32}$`)
	phonePattern      = regexp.MustCompile(`^\+?[0-9][0-9 -]*[0-9]$`)
	timePattern       = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)
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

func trimNull(v null.String) null.String {
	s := strings.TrimSpace(v.String)
	return null.NewString(s, v.Valid && s != "")
}

func normalizeCreate(r *models.LineBookingCreateRequest) {
	r.SalesUserID = strings.TrimSpace(r.SalesUserID)
	r.TourCode = strings.TrimSpace(r.TourCode)
	r.CustomerName = strings.TrimSpace(r.CustomerName)
	r.CustomerPhone = strings.TrimSpace(r.CustomerPhone)
	r.CustomerAddress = strings.TrimSpace(r.CustomerAddress)
	r.SlipImage = strings.TrimSpace(r.SlipImage)
	r.Note = strings.TrimSpace(r.Note)
	r.RefName = strings.TrimSpace(r.RefName)
	r.RefBookingNo = strings.TrimSpace(r.RefBookingNo)
	r.LineContactID = trimNull(r.LineContactID)
	r.BookingSourceID = trimNull(r.BookingSourceID)
	r.DepositDueDate = trimNull(r.DepositDueDate)
	r.RemainingDueDate = trimNull(r.RemainingDueDate)
	r.FullDueDate = trimNull(r.FullDueDate)
	r.DepositDueTime = strings.TrimSpace(r.DepositDueTime)
	r.FullDueTime = strings.TrimSpace(r.FullDueTime)
	if r.PaymentMethod == "" {
		r.PaymentMethod = linebookingpayment.Transfer
	}
	for i := range r.Charges {
		c := &r.Charges[i]
		c.Name = strings.TrimSpace(c.Name)
		if c.Kind != linebookingpayment.ChargeOption {
			c.Name = ""
		}
	}
}

func (u *lineBookingUseCase) CreateValidate(request *models.LineBookingCreateRequest) error {
	count := []validation.Rule{validation.Min(0), validation.Max(maxTravellers)}
	price := []validation.Rule{validation.Min(0), validation.Max(maxUnitPrice)}
	err := validation.ValidateStruct(request,
		validation.Field(&request.SalesUserID, validation.Required, validation.UUID),
		validation.Field(&request.TourCode, validation.Required),
		validation.Field(&request.PeriodID, validation.Required, validation.UUID),
		validation.Field(&request.AdultTwin, count...),
		validation.Field(&request.AdultSingle, count...),
		validation.Field(&request.ChildWithBed, count...),
		validation.Field(&request.ChildNoBed, count...),
		validation.Field(&request.Infant, count...),
		validation.Field(&request.AdultPrice, price...),
		validation.Field(&request.SingleSupplement, price...),
		validation.Field(&request.ChildWithBedPrice, price...),
		validation.Field(&request.ChildNoBedPrice, price...),
		validation.Field(&request.InfantPrice, price...),
		validation.Field(&request.AdultPriceBeforeDiscount, price...),
		validation.Field(&request.DiscountPerPerson, price...),
		validation.Field(&request.DepositPerPerson, price...),
		validation.Field(&request.Charges, validation.Length(0, maxCharges)),
		validation.Field(&request.PaymentMethod, validation.In(linebookingpayment.Methods...)),
		validation.Field(&request.CreditFeePercent, validation.Min(0.0), validation.Max(maxCreditFeePercent)),
		validation.Field(&request.DepositDueTime, validation.Match(timePattern)),
		validation.Field(&request.FullDueTime, validation.Match(timePattern)),
		validation.Field(&request.CustomerName, validation.Required, validation.RuneLength(1, maxNameRunes)),
		validation.Field(&request.CustomerPhone, validation.Required, validation.By(phoneRule)),
		validation.Field(&request.CustomerAddress, validation.RuneLength(0, maxAddressRunes)),
		validation.Field(&request.RefName, validation.RuneLength(0, maxRefRunes)),
		validation.Field(&request.RefBookingNo, validation.RuneLength(0, maxRefRunes)),
		validation.Field(&request.SlipImage, validation.Required),
		validation.Field(&request.Note, validation.RuneLength(0, maxNoteRunes)),
	)
	if err != nil {
		return err
	}
	for field, v := range map[string]null.String{"lineContactId": request.LineContactID, "bookingSourceId": request.BookingSourceID} {
		if v.Valid {
			if err := validation.Validate(v.String, validation.UUID); err != nil {
				return validation.Errors{field: err}
			}
		}
	}
	for field, v := range map[string]null.String{
		"depositDueDate":   request.DepositDueDate,
		"remainingDueDate": request.RemainingDueDate,
		"fullDueDate":      request.FullDueDate,
	} {
		if v.Valid {
			if err := validation.Validate(v.String, validation.Date(models.DateLayout)); err != nil {
				return validation.Errors{field: err}
			}
		}
	}
	for i := range request.Charges {
		c := &request.Charges[i]
		nameRules := []validation.Rule{validation.RuneLength(0, maxChargeNameRunes)}
		if c.Kind == linebookingpayment.ChargeOption {
			nameRules = append(nameRules, validation.Required)
		}
		if err := validation.ValidateStruct(c,
			validation.Field(&c.Kind, validation.Required, validation.In(linebookingpayment.ChargeKinds...)),
			validation.Field(&c.Name, nameRules...),
			validation.Field(&c.Count, count...),
			validation.Field(&c.Price, price...),
		); err != nil {
			return validation.Errors{fmt.Sprintf("charges.%d", i): err}
		}
	}
	if request.Infant > 0 && !request.InfantPrice.Valid {
		return validation.Errors{"infantPrice": errors.New("required when there are infants")}
	}
	if request.IsDeposit && request.DepositPerPerson <= 0 {
		return validation.Errors{"depositPerPerson": errors.New("required for a deposit")}
	}
	switch adults, total := request.AdultTwin+request.AdultSingle, request.AdultTwin+request.AdultSingle+request.ChildWithBed+request.ChildNoBed+request.Infant; {
	case adults < 1:
		return validation.Errors{"adultTwin": errors.New("at least one adult is required")}
	case total > maxTravellers:
		return validation.Errors{"adultTwin": fmt.Errorf("at most %d travellers per booking", maxTravellers)}
	}
	return nil
}

// quoteLine prices the request like the backoffice summary tool (summary.ts computeTotals). Unit
// prices default to the period's; total = price rows - discount per traveller + charges + card fee,
// where the card fee is on the deposit when there is one, otherwise on the whole amount.
func quoteLine(p *domain.TourProgramPeriod, r *models.LineBookingCreateRequest) (*domain.LineBooking, error) {
	or := func(v, fallback null.Int) null.Int {
		if v.Valid {
			return v
		}
		return fallback
	}
	adult := int(or(r.AdultPrice, null.IntFrom(int64(p.Price))).Int64)
	supplement := or(r.SingleSupplement, p.SingleSupplement)
	single := null.Int{}
	if supplement.Valid {
		single = null.IntFrom(int64(adult) + supplement.Int64)
	}
	childWithBed := or(r.ChildWithBedPrice, p.ChildWithBedPrice)
	childNoBed := or(r.ChildNoBedPrice, p.ChildNoBedPrice)
	infant := null.Int{}
	if r.Infant > 0 {
		infant = r.InfantPrice
	}

	fields := validation.Errors{}
	for name, row := range map[string]struct {
		count int
		price null.Int
	}{
		"adultSingle":  {r.AdultSingle, single},
		"childWithBed": {r.ChildWithBed, childWithBed},
		"childNoBed":   {r.ChildNoBed, childNoBed},
	} {
		if row.count > 0 && !row.price.Valid {
			fields[name] = errors.New("this period has no such price; enter one")
		}
	}
	if len(fields) > 0 {
		return nil, fields
	}

	b := &domain.LineBooking{
		AdultTwinCount:           r.AdultTwin,
		AdultTwinPrice:           adult,
		AdultSingleCount:         r.AdultSingle,
		AdultSinglePrice:         single,
		ChildWithBedCount:        r.ChildWithBed,
		ChildWithBedPrice:        childWithBed,
		ChildNoBedCount:          r.ChildNoBed,
		ChildNoBedPrice:          childNoBed,
		InfantCount:              r.Infant,
		InfantPrice:              infant,
		AdultPriceBeforeDiscount: r.AdultPriceBeforeDiscount,
		DiscountPerPerson:        r.DiscountPerPerson,
		Charges:                  make([]domain.LineBookingCharge, 0, len(r.Charges)),
		PaymentMethod:            r.PaymentMethod,
		IsDeposit:                r.IsDeposit,
	}
	travellers := b.Travellers()
	subtotal := r.AdultTwin*adult +
		r.AdultSingle*int(single.Int64) +
		r.ChildWithBed*int(childWithBed.Int64) +
		r.ChildNoBed*int(childNoBed.Int64) +
		r.Infant*int(infant.Int64) -
		r.DiscountPerPerson*travellers
	for _, c := range r.Charges {
		if c.Count == 0 || c.Price == 0 {
			continue // an unused line; the summary text leaves it out too
		}
		b.Charges = append(b.Charges, domain.LineBookingCharge{Kind: c.Kind, Name: c.Name, Count: c.Count, Price: c.Price})
		subtotal += c.Count * c.Price
	}
	if subtotal < 0 {
		return nil, validation.Errors{"discountPerPerson": errors.New("the discount is more than the price")}
	}

	deposit := 0
	if r.IsDeposit {
		b.DepositPerPerson = r.DepositPerPerson
		deposit = r.DepositPerPerson * travellers
		if deposit > subtotal {
			return nil, validation.Errors{"depositPerPerson": errors.New("the deposit is more than the total")}
		}
	}
	if r.PaymentMethod == linebookingpayment.Credit {
		b.CreditFeeBps = int(math.Round(r.CreditFeePercent * 100))
		base := subtotal
		if r.IsDeposit {
			base = deposit
		}
		b.CardFee = cardFee(base, b.CreditFeeBps)
	}
	b.TotalPrice = subtotal + b.CardFee
	return b, nil
}

// cardFee is amount x bps / 10000, rounded up to the next baht (same as the summary tool).
func cardFee(amount, bps int) int {
	return (amount*bps + 9999) / 10000
}

func parseDate(v null.String) *time.Time {
	if !v.Valid {
		return nil
	}
	t, err := time.Parse(models.DateLayout, v.String)
	if err != nil {
		return nil
	}
	return &t
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
		if request.BookingSourceID.Valid {
			source, err := u.repositoryRegistry.BookingSourceRepository.WithTx(appDB).FindByID(ctx, request.BookingSourceID.String)
			if err != nil {
				return err
			}
			if source == nil || !source.Status {
				return cerrors.FieldError("bookingSourceId", "no such booking source")
			}
			b.BookingSourceID = request.BookingSourceID
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
		b.CustomerAddress = request.CustomerAddress
		b.RefName = request.RefName
		b.RefBookingNo = request.RefBookingNo
		if b.IsDeposit {
			b.DepositDueDate = parseDate(request.DepositDueDate)
			b.DepositDueTime = request.DepositDueTime
			b.RemainingDueDate = parseDate(request.RemainingDueDate)
		} else {
			b.FullDueDate = parseDate(request.FullDueDate)
			b.FullDueTime = request.FullDueTime
		}
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
