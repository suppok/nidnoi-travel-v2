package usecase

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"backend/pkg/constants/appconfigcode"
	"backend/pkg/constants/emailoutboxstate"
	"backend/pkg/constants/linebookingevent"
	"backend/pkg/constants/linebookingstate"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
	"backend/pkg/modules/shared/services/appconfigload"
	"backend/pkg/modules/shared/services/emailoutbox"
	"backend/pkg/modules/shared/services/notification"
)

// wholesaleEmailDecision: whether a booking saved on the entry page emails its wholesale.
func wholesaleEmailDecision(w *domain.Wholesale) string {
	switch {
	case w == nil || !w.AutoEmail:
		return models.WholesaleEmailDisabled
	case len(w.Emails) == 0:
		return models.WholesaleEmailNoEmail
	default:
		return models.WholesaleEmailQueued
	}
}

// wholesaleEmailAddresses: From is MAIL_FROM, To the wholesale's emails, Reply-To and Cc the app
// config WHOLESALE_EMAIL. A missing or invalid config never blocks a booking: the email is queued
// without Reply-To/Cc and an error is logged (POST /app-configs/verify reports it too).
func (u *lineBookingUseCase) wholesaleEmailAddresses(ctx context.Context, appDB query.AppIDB, w *domain.Wholesale) (emailoutbox.Addresses, error) {
	addr := emailoutbox.Addresses{From: u.coreRegistry.Config.Mail.From, To: w.Emails, Cc: []string{}}
	code := appconfigcode.AppConfigWholesaleEmail
	cfg, err := appconfigload.GetAppConfigByOption(ctx, u.repositoryRegistry.AppConfigRepository.WithTx(appDB), code, false)
	if err != nil {
		return addr, err
	}
	switch {
	case cfg == nil:
		u.coreRegistry.Logger.Error("app config missing: wholesale email queued without Reply-To/Cc", zap.String("code", code.Code()))
	case cfg.Validate() != nil:
		u.coreRegistry.Logger.Error("app config invalid: wholesale email queued without Reply-To/Cc",
			zap.String("code", code.Code()), zap.Error(cfg.Validate()))
	default:
		addr.ReplyTo = cfg.ReplyTo
		addr.Cc = append(addr.Cc, cfg.Cc...)
	}
	return addr, nil
}

// contactCardID: the uploaded contact card (app config EMAIL_CONTACT_CARD), else the built-in one.
// A broken config never blocks a booking.
func (u *lineBookingUseCase) contactCardID(ctx context.Context, appDB query.AppIDB) string {
	code := appconfigcode.AppConfigEmailContactCard
	cfg, err := appconfigload.GetAppConfigByOption(ctx, u.repositoryRegistry.AppConfigRepository.WithTx(appDB), code, false)
	if err != nil {
		u.coreRegistry.Logger.Error("app config: read failed, using the built-in contact card", zap.String("code", code.Code()), zap.Error(err))
		return notification.DefaultContactCardID
	}
	if cfg == nil || cfg.ImageURL == "" {
		return notification.DefaultContactCardID
	}
	id, ok := notification.ContactCardID(cfg.ImageURL, u.shareRegistry.ClientRegistry.ObjectStorage.PublicBaseURL())
	if !ok {
		u.coreRegistry.Logger.Error("app config invalid: using the built-in contact card", zap.String("code", code.Code()))
	}
	return id
}

// renderWholesaleEmail renders the booking email to the wholesale and builds its outbox row. b must be
// loaded with its relations (BookingSource, SalesUser) and tour with Wholesale and Airline.
func (u *lineBookingUseCase) renderWholesaleEmail(ctx context.Context, appDB query.AppIDB, b *domain.LineBooking, tour *domain.TourProgram, w *domain.Wholesale, actorID string, now time.Time) (*domain.EmailOutbox, error) {
	email, err := notification.RenderWholesaleBooking(notification.NewWholesaleBooking(b, tour, u.contactCardID(ctx, appDB)))
	if err != nil {
		return nil, err
	}
	addr, err := u.wholesaleEmailAddresses(ctx, appDB, w)
	if err != nil {
		return nil, err
	}
	return emailoutbox.NewRow(addr, b.ID, emailoutboxstate.TypeWholesaleBooking, email, actorID, now), nil
}

// queueWholesaleEmail runs inside the create transaction: when the tour's wholesale has auto email on
// and an address, it queues the booking email in the outbox, so the booking and its email commit or
// roll back together. The sender worker delivers it; nothing is sent here.
func (u *lineBookingUseCase) queueWholesaleEmail(ctx context.Context, appDB query.AppIDB, bookingID, tourID, actorID string) (models.LineBookingWholesaleEmailResult, error) {
	result := models.LineBookingWholesaleEmailResult{To: []string{}}
	tour, err := u.repositoryRegistry.TourProgramRepository.WithTx(appDB).FindByID(ctx, tourID)
	if err != nil {
		return result, err
	}
	var w *domain.Wholesale
	if tour != nil {
		w = tour.Wholesale
	}
	if w != nil {
		result.WholesaleName = w.Name
	}
	result.Result = wholesaleEmailDecision(w)
	if result.Result != models.WholesaleEmailQueued {
		return result, nil
	}
	// Reload the booking with the source and sales owner the summary shows.
	b, err := u.repositoryRegistry.LineBookingRepository.WithTx(appDB).FindByID(ctx, bookingID)
	if err != nil {
		return result, err
	}
	row, err := u.renderWholesaleEmail(ctx, appDB, b, tour, w, actorID, time.Now())
	if err != nil {
		return result, err
	}
	// A duplicate (already queued for this booking) is a no-op, not an error.
	if _, err := u.repositoryRegistry.EmailOutboxRepository.WithTx(appDB).CreateIfAbsent(ctx, row); err != nil {
		return result, err
	}
	result.To = w.Emails
	return result, nil
}

// WholesaleEmailGet returns the state of the booking's email to its wholesale (same visibility as get).
func (u *lineBookingUseCase) WholesaleEmailGet(ctx context.Context, request *models.LineBookingWholesaleEmailGetRequest) (*models.LineBookingWholesaleEmailGetResponse, error) {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return nil, cerrors.Validation(validation.Errors{"id": err})
	}
	b, err := u.repositoryRegistry.LineBookingRepository.FindByID(ctx, request.ID)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking for wholesale email")
	}
	if b == nil || !visibleTo(b, request.UserRequestInfo) {
		return nil, cerrors.NotFound()
	}
	row, err := u.repositoryRegistry.EmailOutboxRepository.FindByBooking(ctx, b.ID, emailoutboxstate.TypeWholesaleBooking)
	if err != nil {
		return nil, cerrors.Wrap(err, "get wholesale email")
	}
	response := &models.LineBookingWholesaleEmailGetResponse{}
	if row != nil {
		item := models.NewLineBookingWholesaleEmailItem(row)
		response.Email = &item
	}
	return response, nil
}

func (u *lineBookingUseCase) WholesaleEmailResendValidate(request *models.LineBookingWholesaleEmailResendRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
	)
}

// WholesaleEmailResend sends the booking email to the wholesale again (sales owner or admin). It
// re-reads the wholesale's current addresses, MAIL_FROM and WHOLESALE_EMAIL, re-renders the email
// and starts a new send round (pending, attempts 0). It also sends a booking whose wholesale had auto
// email off when it was saved: a manual send ignores auto email.
func (u *lineBookingUseCase) WholesaleEmailResend(ctx context.Context, request *models.LineBookingWholesaleEmailResendRequest) (*models.LineBookingWholesaleEmailResendResponse, error) {
	if err := u.WholesaleEmailResendValidate(request); err != nil {
		return nil, cerrors.Validation(err)
	}
	info := request.UserRequestInfo

	var saved *domain.EmailOutbox
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		b, err := u.repositoryRegistry.LineBookingRepository.WithTx(appDB).FindByID(ctx, request.ID)
		if err != nil {
			return err
		}
		if b == nil || !b.Status || !visibleTo(b, info) {
			return cerrors.NotFound()
		}
		if !ownsAsSales(b, info) {
			return cerrors.PermissionDenied()
		}
		if b.State == linebookingstate.Cancelled {
			return cerrors.FieldError("id", "the booking is cancelled")
		}
		tour, err := u.repositoryRegistry.TourProgramRepository.WithTx(appDB).FindByID(ctx, b.TourProgramID)
		if err != nil {
			return err
		}
		if tour == nil {
			return cerrors.NotFound()
		}
		w := tour.Wholesale
		if w == nil || len(w.Emails) == 0 {
			return cerrors.FieldError("id", "the wholesale has no email address")
		}

		now := time.Now()
		fresh, err := u.renderWholesaleEmail(ctx, appDB, b, tour, w, info.UserID, now)
		if err != nil {
			return err
		}
		repo := u.repositoryRegistry.EmailOutboxRepository.WithTx(appDB)
		row, err := repo.FindByBooking(ctx, b.ID, emailoutboxstate.TypeWholesaleBooking)
		if err != nil {
			return err
		}
		if row == nil {
			if _, err := repo.CreateIfAbsent(ctx, fresh); err != nil {
				return err
			}
			saved = fresh
		} else {
			row.FromEmail, row.ToEmails, row.CcEmails, row.ReplyTo = fresh.FromEmail, fresh.ToEmails, fresh.CcEmails, fresh.ReplyTo
			row.Subject, row.BodyHTML, row.BodyText = fresh.Subject, fresh.BodyHTML, fresh.BodyText
			row.State = emailoutboxstate.Pending
			row.Attempts = 0
			row.SendRound++
			row.NextAttemptAt = now
			row.LastError = ""
			row.UpdateDate = now
			row.UpdateBy = info.UserID
			if saved, err = repo.Update(ctx, row, "from_email", "to_emails", "cc_emails", "reply_to", "subject", "body_html",
				"body_text", "state", "attempts", "send_round", "next_attempt_at", "last_error", "update_date", "update_by"); err != nil {
				return err
			}
		}
		return u.logEvent(ctx, appDB, b.ID, linebookingevent.WholesaleEmailResent, info.UserID,
			domain.LineBookingEventDetail{Target: strings.Join(w.Emails, ", ")})
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "resend wholesale email")
	}
	return &models.LineBookingWholesaleEmailResendResponse{Email: models.NewLineBookingWholesaleEmailItem(saved)}, nil
}
