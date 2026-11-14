package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/guregu/null/v5"

	"backend/pkg/constants/linebookingpayment"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

// logEvent appends an action to a booking's history inside the caller's transaction, so the change
// and its history row commit or roll back together.
func (u *lineBookingUseCase) logEvent(ctx context.Context, appDB query.AppIDB, bookingID, action, userID string, detail domain.LineBookingEventDetail) error {
	_, err := u.repositoryRegistry.LineBookingEventRepository.WithTx(appDB).Create(ctx, &domain.LineBookingEvent{
		ID:            utils.NewID(),
		LineBookingID: bookingID,
		Action:        action,
		Detail:        detail,
		CreateDate:    time.Now(),
		CreateBy:      userID,
	})
	return err
}

// documentCodes is the documents of the current round, e.g. "AI + IV".
func documentCodes(b *domain.LineBooking) string {
	if b.NeedsReceipt() {
		return "RE"
	}
	return "AI + IV"
}

func userName(user *domain.AppUser) string {
	if user == nil {
		return ""
	}
	return user.Name
}

func contactName(c *domain.LineContact) string {
	if c == nil {
		return ""
	}
	return c.DisplayName
}

var chargeLabels = map[string]string{
	linebookingpayment.ChargeVisa:            "วีซ่า",
	linebookingpayment.ChargeForeignPassport: "พาสปอร์ตต่างชาติ",
	linebookingpayment.ChargeJoinLand:        "Join land",
	linebookingpayment.ChargeFuelTax:         "ภาษีน้ำมัน",
	linebookingpayment.ChargeAirportTax:      "ภาษีสนามบิน",
}

func intText(v int) string { return strconv.Itoa(v) }

func nullIntText(v null.Int) string {
	if !v.Valid {
		return ""
	}
	return strconv.FormatInt(v.Int64, 10)
}

func dateText(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(models.DateLayout)
}

// chargesText is the charge lines as one readable string: "วีซ่า 2×1500, ประกัน 1×500".
func chargesText(charges []domain.LineBookingCharge) string {
	parts := make([]string, len(charges))
	for i, c := range charges {
		name := chargeLabels[c.Kind]
		if c.Kind == linebookingpayment.ChargeOption {
			name = c.Name
		}
		parts[i] = fmt.Sprintf("%s %d×%d", name, c.Count, c.Price)
	}
	return strings.Join(parts, ", ")
}

// dealSnapshot is the part of a booking a LIFF edit may change, as display strings in a fixed order.
// Booking source and LINE contact are read from their relations.
func dealSnapshot(b *domain.LineBooking) []domain.LineBookingEventChange {
	field := func(name, value string) domain.LineBookingEventChange {
		return domain.LineBookingEventChange{Field: name, To: value}
	}
	source := ""
	if b.BookingSource != nil {
		source = b.BookingSource.Name
	}
	return []domain.LineBookingEventChange{
		field("adultTwin", intText(b.AdultTwinCount)),
		field("adultSingle", intText(b.AdultSingleCount)),
		field("childWithBed", intText(b.ChildWithBedCount)),
		field("childNoBed", intText(b.ChildNoBedCount)),
		field("infant", intText(b.InfantCount)),
		field("adultPrice", intText(b.AdultTwinPrice)),
		field("adultSinglePrice", nullIntText(b.AdultSinglePrice)),
		field("childWithBedPrice", nullIntText(b.ChildWithBedPrice)),
		field("childNoBedPrice", nullIntText(b.ChildNoBedPrice)),
		field("infantPrice", nullIntText(b.InfantPrice)),
		field("adultPriceBeforeDiscount", nullIntText(b.AdultPriceBeforeDiscount)),
		field("discountPerPerson", intText(b.DiscountPerPerson)),
		field("charges", chargesText(b.Charges)),
		field("paymentMethod", b.PaymentMethod),
		field("creditFeePercent", strconv.FormatFloat(float64(b.CreditFeeBps)/100, 'f', -1, 64)),
		field("depositPerPerson", intText(b.DepositPerPerson)),
		field("depositDueDate", dateText(b.DepositDueDate)),
		field("depositDueTime", b.DepositDueTime),
		field("remainingDueDate", dateText(b.RemainingDueDate)),
		field("fullDueDate", dateText(b.FullDueDate)),
		field("fullDueTime", b.FullDueTime),
		field("totalPrice", intText(b.TotalPrice)),
		field("customerName", b.CustomerName),
		field("customerPhone", b.CustomerPhone),
		field("customerAddress", b.CustomerAddress),
		field("bookingSource", source),
		field("refName", b.RefName),
		field("refBookingNo", b.RefBookingNo),
		field("lineContact", contactName(b.LineContact)),
		field("note", b.Note),
	}
}

// diffSnapshots lists the fields whose value changed between two dealSnapshot results.
func diffSnapshots(before, after []domain.LineBookingEventChange) []domain.LineBookingEventChange {
	changes := []domain.LineBookingEventChange{}
	for i := range after {
		if before[i].To != after[i].To {
			changes = append(changes, domain.LineBookingEventChange{Field: after[i].Field, From: before[i].To, To: after[i].To})
		}
	}
	return changes
}

// Events returns a booking's history, oldest first. Same visibility as GET /line-bookings/:id.
func (u *lineBookingUseCase) Events(ctx context.Context, request *models.LineBookingEventListRequest) (*models.LineBookingEventListResponse, error) {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return nil, cerrors.Validation(validation.Errors{"id": err})
	}
	b, err := u.repositoryRegistry.LineBookingRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking for events")
	}
	if b == nil || !visibleTo(b, request.UserRequestInfo) {
		return nil, cerrors.NotFound()
	}
	return u.eventList(ctx, b.ID)
}

// LIFFEvents returns the history of a booking reopened in LIFF (same rule as LIFFGet).
func (u *lineBookingUseCase) LIFFEvents(ctx context.Context, request *models.LineBookingLIFFGetRequest) (*models.LineBookingEventListResponse, error) {
	b, err := findEditable(ctx, u.repositoryRegistry.LineBookingRepository, request.BookingNo)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking for LIFF events")
	}
	return u.eventList(ctx, b.ID)
}

func (u *lineBookingUseCase) eventList(ctx context.Context, bookingID string) (*models.LineBookingEventListResponse, error) {
	events, err := u.repositoryRegistry.LineBookingEventRepository.ListByBooking(ctx, bookingID)
	if err != nil {
		return nil, cerrors.Wrap(err, "list line booking events")
	}
	items := make([]models.LineBookingEventItem, len(events))
	for i, e := range events {
		items[i] = models.LineBookingEventItem{
			ID:         e.ID,
			Action:     e.Action,
			Changes:    e.Detail.Changes,
			Target:     e.Detail.Target,
			Note:       e.Detail.Note,
			CreateDate: e.CreateDate,
		}
		if items[i].Changes == nil {
			items[i].Changes = []domain.LineBookingEventChange{}
		}
		if user := e.CreateUser; user != nil {
			items[i].User = &models.LineBookingUser{ID: user.ID, Name: user.Name}
		}
	}
	return &models.LineBookingEventListResponse{Items: items}, nil
}
