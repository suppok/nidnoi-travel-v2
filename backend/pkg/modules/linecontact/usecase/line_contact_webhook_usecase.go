package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/guregu/null/v5"
	"go.uber.org/zap"

	"backend/pkg/constants/errorcode"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/domain"
	"backend/pkg/modules/linecontact/models"
	"backend/pkg/modules/shared/client/line"
)

// systemUserID is create_by/update_by for rows written by webhooks.
const systemUserID = "00000000-0000-0000-0000-000000000000"

// profileMaxAge is how long a stored display name and picture are trusted before a message refreshes them.
const profileMaxAge = 24 * time.Hour

type contactEventKind int

const (
	contactFollow contactEventKind = iota + 1
	contactUnfollow
	contactMessage
)

type contactEvent struct {
	UserID string
	Kind   contactEventKind
	At     time.Time
}

// collectContactEvents keeps the events that tell us about a 1:1 customer: follow, unfollow and
// messages from a user. Group/room events and other types are ignored.
func collectContactEvents(events []models.WebhookEvent) []contactEvent {
	var out []contactEvent
	for _, e := range events {
		if e.Source.Type != "user" || e.Source.UserID == "" {
			continue
		}
		var kind contactEventKind
		switch e.Type {
		case "follow":
			kind = contactFollow
		case "unfollow":
			kind = contactUnfollow
		case "message":
			kind = contactMessage
		default:
			continue
		}
		out = append(out, contactEvent{UserID: e.Source.UserID, Kind: kind, At: time.UnixMilli(e.Timestamp)})
	}
	return out
}

func (u *lineContactUseCase) HandleWebhook(ctx context.Context, request *models.LineWebhookRequest) (*models.LineWebhookResponse, error) {
	lineClient := u.shareRegistry.ClientRegistry.LINE
	if !lineClient.WebhookConfigured() {
		return nil, cerrors.New(http.StatusServiceUnavailable, errorcode.LINENotConfigured)
	}
	if !lineClient.VerifySignature(request.Body, request.Signature) {
		return nil, cerrors.New(http.StatusUnauthorized, errorcode.InvalidLINESignature)
	}

	var payload models.WebhookPayload
	if err := json.Unmarshal(request.Body, &payload); err != nil {
		return nil, cerrors.New(http.StatusBadRequest, errorcode.Validation)
	}

	// The console's "Verify" button sends an empty event list; it still gets a 200.
	for _, e := range collectContactEvents(payload.Events) {
		if err := u.applyContactEvent(ctx, e); err != nil {
			return nil, cerrors.Wrap(err, "record line contact")
		}
	}
	return &models.LineWebhookResponse{}, nil
}

// applyContactEvent creates the contact on first sight, then records the event. Profile lookups
// that fail are logged and never fail the webhook.
func (u *lineContactUseCase) applyContactEvent(ctx context.Context, e contactEvent) error {
	repo := u.repositoryRegistry.LineContactRepository
	now := time.Now()
	if err := repo.CreateIfAbsent(ctx, &domain.LineContact{
		ID:         utils.NewID(),
		LineUserID: e.UserID,
		Status:     true,
		CreateDate: now,
		UpdateDate: now,
		CreateBy:   systemUserID,
		UpdateBy:   systemUserID,
	}); err != nil {
		return err
	}
	contact, err := repo.FindByLineUserID(ctx, e.UserID)
	if err != nil {
		return err
	}
	if contact == nil {
		return errors.New("line contact missing after insert")
	}

	refreshProfile := false
	switch e.Kind {
	case contactFollow:
		contact.Following = true
		refreshProfile = true
	case contactUnfollow:
		contact.Following = false
	case contactMessage:
		if !contact.LastMessageAt.Valid || e.At.After(contact.LastMessageAt.Time) {
			contact.LastMessageAt = null.TimeFrom(e.At)
		}
		refreshProfile = !contact.ProfileUpdatedAt.Valid || now.Sub(contact.ProfileUpdatedAt.Time) > profileMaxAge
	}

	if refreshProfile {
		u.refreshProfile(ctx, contact, now)
	}

	contact.UpdateDate = now
	contact.UpdateBy = systemUserID
	_, err = repo.Update(ctx, contact,
		"display_name", "picture_url", "following", "last_message_at", "profile_updated_at", "update_date", "update_by")
	return err
}

func (u *lineContactUseCase) refreshProfile(ctx context.Context, contact *domain.LineContact, now time.Time) {
	profile, err := u.shareRegistry.ClientRegistry.LINE.GetProfile(ctx, contact.LineUserID)
	switch {
	case err == nil:
		contact.DisplayName = profile.DisplayName
		contact.PictureURL = null.NewString(profile.PictureURL, profile.PictureURL != "")
		contact.Following = true
		contact.ProfileUpdatedAt = null.TimeFrom(now)
	case errors.Is(err, line.ErrProfileNotFound):
		// Not a friend (never added the OA, or blocked it): pushes would fail.
		contact.Following = false
		contact.ProfileUpdatedAt = null.TimeFrom(now)
	case errors.Is(err, line.ErrNotConfigured):
		// No channel access token yet: keep the contact without a profile.
	default:
		u.coreRegistry.Logger.Warn("line profile lookup failed", zap.Error(err))
	}
}
