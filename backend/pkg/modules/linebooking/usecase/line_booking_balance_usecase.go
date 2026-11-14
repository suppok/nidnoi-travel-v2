package usecase

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	"backend/pkg/constants/linebookingstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
	"backend/pkg/repository"
)

const maxSlipURLRunes = 1000

// findEditable loads a booking reopened in LIFF by its booking no: NOT_FOUND when there is no such
// active booking, LINE_BOOKING_NOT_EDITABLE unless it is paid (deposit or in full).
// Any user who may create LINE bookings may reopen one (sales share booking numbers).
func findEditable(ctx context.Context, repo repository.LineBookingRepository, bookingNo string) (*domain.LineBooking, error) {
	no := strings.ToUpper(strings.TrimSpace(bookingNo))
	if !publicCodePattern.MatchString(no) {
		return nil, cerrors.NotFound()
	}
	b, err := repo.FindByBookingNo(ctx, no)
	if err != nil {
		return nil, err
	}
	if b == nil || !b.Status {
		return nil, cerrors.NotFound()
	}
	if !b.EditableFromLIFF() {
		return nil, cerrors.New(http.StatusConflict, errorcode.LineBookingNotEditable)
	}
	return b, nil
}

// LIFFGet returns a booking reopened in LIFF by its booking no.
func (u *lineBookingUseCase) LIFFGet(ctx context.Context, request *models.LineBookingLIFFGetRequest) (*models.LineBookingLIFFGetResponse, error) {
	b, err := findEditable(ctx, u.repositoryRegistry.LineBookingRepository, request.BookingNo)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking for LIFF")
	}
	return &models.LineBookingLIFFGetResponse{LineBookingItem: models.NewLineBookingItem(b)}, nil
}

// LIFFUpdate saves a reopened booking's deal (travellers, prices, charges, payment, customer) with
// the total quoted again, and records the balance slip when one is sent. The state and the
// finance documents don't change here; the balance moves on when the sales owner confirms it.
func (u *lineBookingUseCase) LIFFUpdate(ctx context.Context, request *models.LineBookingLIFFUpdateRequest) (*models.LineBookingLIFFUpdateResponse, error) {
	info := request.UserRequestInfo
	if info == nil || info.LineUserID == "" {
		return nil, cerrors.Unauthenticated()
	}
	request.BalanceSlipImage = strings.TrimSpace(request.BalanceSlipImage)

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		b, err := findEditable(ctx, repo, request.BookingNo)
		if err != nil {
			return err
		}

		// What stays fixed is taken from the booking, so the create rules validate the rest.
		r := &request.LineBookingCreateRequest
		r.TourCode, r.PeriodID, r.SlipImage = b.TourCode, b.TourProgramPeriodID, b.SlipImage
		r.SalesUserID = b.SalesUserID.String
		if !b.SalesUserID.Valid {
			r.SalesUserID = info.UserID // validation only; the owner isn't changed
		}
		r.IsDeposit = b.IsDeposit // the payment plan stays
		normalizeCreate(r)
		if err := u.CreateValidate(r); err != nil {
			return cerrors.Validation(err)
		}
		if err := validation.Validate(request.BalanceSlipImage, validation.RuneLength(0, maxSlipURLRunes)); err != nil {
			return cerrors.Validation(validation.Errors{"balanceSlipImage": err})
		}
		if request.BalanceSlipImage != "" && !b.AwaitsBalance() {
			return cerrors.FieldError("balanceSlipImage", "only a deposit booking whose deposit documents were sent, not yet paid in full, takes a balance slip")
		}

		period, err := u.repositoryRegistry.TourProgramPeriodRepository.WithTx(appDB).FindByID(ctx, b.TourProgramPeriodID)
		if err != nil {
			return err
		}
		if period == nil {
			return cerrors.Conflict(errorcode.PeriodUnavailable)
		}
		q, err := quoteLine(period, r)
		if err != nil {
			return cerrors.Validation(err)
		}

		lineContactID := null.String{}
		if r.LineContactID.Valid {
			contact, err := u.repositoryRegistry.LineContactRepository.WithTx(appDB).FindByID(ctx, r.LineContactID.String)
			if err != nil {
				return err
			}
			if contact == nil || !contact.Status {
				return cerrors.FieldError("lineContactId", "no such LINE contact")
			}
			lineContactID = r.LineContactID
		}
		sourceID := null.String{}
		if r.BookingSourceID.Valid {
			source, err := u.repositoryRegistry.BookingSourceRepository.WithTx(appDB).FindByID(ctx, r.BookingSourceID.String)
			if err != nil {
				return err
			}
			// A source deleted since the booking was made may stay on it.
			if source == nil || (!source.Status && source.ID != b.BookingSourceID.String) {
				return cerrors.FieldError("bookingSourceId", "no such booking source")
			}
			sourceID = r.BookingSourceID
		}

		now := time.Now()
		b.AdultTwinCount, b.AdultTwinPrice = q.AdultTwinCount, q.AdultTwinPrice
		b.AdultSingleCount, b.AdultSinglePrice = q.AdultSingleCount, q.AdultSinglePrice
		b.ChildWithBedCount, b.ChildWithBedPrice = q.ChildWithBedCount, q.ChildWithBedPrice
		b.ChildNoBedCount, b.ChildNoBedPrice = q.ChildNoBedCount, q.ChildNoBedPrice
		b.InfantCount, b.InfantPrice = q.InfantCount, q.InfantPrice
		b.AdultPriceBeforeDiscount, b.DiscountPerPerson = q.AdultPriceBeforeDiscount, q.DiscountPerPerson
		b.Charges = q.Charges
		b.PaymentMethod, b.CreditFeeBps, b.CardFee = q.PaymentMethod, q.CreditFeeBps, q.CardFee
		b.DepositPerPerson = q.DepositPerPerson
		b.TotalPrice = q.TotalPrice
		if b.IsDeposit {
			b.DepositDueDate = parseDate(r.DepositDueDate)
			b.DepositDueTime = r.DepositDueTime
			b.RemainingDueDate = parseDate(r.RemainingDueDate)
		} else {
			b.FullDueDate = parseDate(r.FullDueDate)
			b.FullDueTime = r.FullDueTime
		}
		b.CustomerName, b.CustomerPhone, b.CustomerAddress = r.CustomerName, r.CustomerPhone, r.CustomerAddress
		b.RefName, b.RefBookingNo = r.RefName, r.RefBookingNo
		b.BookingSourceID = sourceID
		b.LineContactID = lineContactID
		b.Note = r.Note
		b.UpdateDate = now
		b.UpdateBy = info.UserID
		columns := []string{
			"adult_twin_count", "adult_twin_price", "adult_single_count", "adult_single_price",
			"child_with_bed_count", "child_with_bed_price", "child_no_bed_count", "child_no_bed_price",
			"infant_count", "infant_price", "adult_price_before_discount", "discount_per_person", "charges",
			"payment_method", "credit_fee_bps", "card_fee", "deposit_per_person", "total_price",
			"deposit_due_date", "deposit_due_time", "remaining_due_date", "full_due_date", "full_due_time",
			"customer_name", "customer_phone", "customer_address", "ref_name", "ref_booking_no",
			"booking_source_id", "line_contact_id", "note", "update_date", "update_by",
		}
		if request.BalanceSlipImage != "" {
			b.BalanceSlipImage = request.BalanceSlipImage
			b.BalanceSubmittedAt = null.TimeFrom(now)
			b.BalanceSubmittedBy = null.StringFrom(info.UserID)
			columns = append(columns, "balance_slip_image", "balance_submitted_at", "balance_submitted_by")
		}
		if _, err := repo.Update(ctx, b, columns...); err != nil {
			return err
		}
		updated = b
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "update line booking from LIFF")
	}

	u.coreRegistry.Logger.Info("line booking edited from LIFF",
		zap.String("bookingNo", updated.BookingNo),
		zap.Bool("balance", request.BalanceSlipImage != ""),
		zap.String("userId", info.UserID))
	return &models.LineBookingLIFFUpdateResponse{BookingNo: updated.BookingNo, TotalPrice: updated.TotalPrice, BalancePending: updated.BalancePending()}, nil
}

// ConfirmBalance is the sales owner's confirm of the balance recorded from LIFF: the booking
// becomes paid in full and goes back to its finance user for the RE (the deposit's AI + IV stay).
func (u *lineBookingUseCase) ConfirmBalance(ctx context.Context, request *models.LineBookingConfirmBalanceRequest) (*models.LineBookingConfirmBalanceResponse, error) {
	if err := validation.ValidateStruct(request, validation.Field(&request.ID, validation.Required, validation.UUID)); err != nil {
		return nil, cerrors.Validation(err)
	}

	var updated *domain.LineBooking
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		repo := u.repositoryRegistry.LineBookingRepository.WithTx(appDB)
		info := request.UserRequestInfo
		target, err := repo.FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if target == nil || !target.Status || !visibleTo(target, info) {
			return cerrors.NotFound()
		}
		if !ownsAsSales(target, info) {
			return cerrors.PermissionDenied()
		}
		if !target.BalancePending() {
			return cerrors.New(http.StatusConflict, errorcode.LineBookingBalanceNotPending)
		}

		now := time.Now()
		target.State = linebookingstate.PaidInFull
		target.BalanceConfirmedAt = null.TimeFrom(now)
		// A new documents round: finance uploads the RE, the owner reviews and sends it.
		target.REFile = ""
		target.DocumentsAt, target.DocumentsBy = null.Time{}, null.String{}
		target.DocumentsApprovedAt, target.DocumentsApprovedBy = null.Time{}, null.String{}
		target.CustomerSentAt, target.CustomerSentContactID = null.Time{}, null.String{}
		target.FixNote, target.FixRequestedAt = "", null.Time{}
		target.UpdateDate = now
		target.UpdateBy = info.UserID
		if _, err = repo.Update(ctx, target, "state", "balance_confirmed_at", "re_file",
			"documents_at", "documents_by", "documents_approved_at", "documents_approved_by",
			"customer_sent_at", "customer_sent_contact_id", "fix_note", "fix_requested_at",
			"update_date", "update_by"); err != nil {
			return err
		}
		updated, err = repo.FindByID(ctx, target.ID)
		return err
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "confirm line booking balance")
	}
	return &models.LineBookingConfirmBalanceResponse{LineBookingItem: models.NewLineBookingItem(updated)}, nil
}
