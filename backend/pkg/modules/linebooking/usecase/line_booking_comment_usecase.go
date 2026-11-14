package usecase

import (
	"context"
	"slices"
	"strings"
	"time"

	"backend/pkg/constants/linebookingevent"
	"backend/pkg/constants/role"
	coremodels "backend/pkg/core/api/models"
	"backend/pkg/core/database/query"
	cerrors "backend/pkg/core/errors"
	"backend/pkg/core/utils"
	"backend/pkg/core/validation"
	"backend/pkg/domain"
	"backend/pkg/modules/linebooking/models"
)

const (
	maxCommentRunes   = 2000
	maxMentions       = 20
	maxNotificationPg = 100
)

// mentionable is who may be @mentioned on b: the people who can see it — its active sales owner,
// its active finance user and every active admin / super admin. Sorted by name, no duplicates.
func (u *lineBookingUseCase) mentionable(ctx context.Context, appDB query.AppIDB, b *domain.LineBooking) ([]models.LineBookingUser, error) {
	users := map[string]string{}
	for _, user := range []*domain.AppUser{b.SalesUser, b.FinanceUser} {
		if user != nil && user.Status {
			users[user.ID] = user.Name
		}
	}
	repo := u.repositoryRegistry.AppUserRepository.WithTx(appDB)
	for _, r := range []string{role.SuperAdmin, role.Admin} {
		admins, err := repo.ListActiveByRole(ctx, r)
		if err != nil {
			return nil, err
		}
		for _, a := range admins {
			users[a.ID] = a.Name
		}
	}
	items := make([]models.LineBookingUser, 0, len(users))
	for id, name := range users {
		items = append(items, models.LineBookingUser{ID: id, Name: name})
	}
	slices.SortFunc(items, func(a, b models.LineBookingUser) int { return strings.Compare(a.Name, b.Name) })
	return items, nil
}

// findVisible loads a booking the caller can see (NOT_FOUND otherwise).
func (u *lineBookingUseCase) findVisible(ctx context.Context, appDB query.AppIDB, id string, info *coremodels.UserRequestInfo) (*domain.LineBooking, error) {
	b, err := u.repositoryRegistry.LineBookingRepository.WithTx(appDB).FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil || !visibleTo(b, info) {
		return nil, cerrors.NotFound()
	}
	return b, nil
}

// Mentionable lists who may be @mentioned on a booking.
func (u *lineBookingUseCase) Mentionable(ctx context.Context, request *models.LineBookingMentionableRequest) (*models.LineBookingMentionableResponse, error) {
	if err := validation.Validate(request.ID, validation.Required, validation.UUID); err != nil {
		return nil, cerrors.Validation(validation.Errors{"id": err})
	}
	appDB := u.coreRegistry.AppDB
	b, err := u.findVisible(ctx, appDB, request.ID, request.UserRequestInfo)
	if err != nil {
		return nil, cerrors.Wrap(err, "get line booking for mentions")
	}
	items, err := u.mentionable(ctx, appDB, b)
	if err != nil {
		return nil, cerrors.Wrap(err, "list mentionable users")
	}
	return &models.LineBookingMentionableResponse{Items: items}, nil
}

// Comment adds a comment to a booking's history and notifies, besides the author, the booking's
// sales owner and finance user ("comment") and every @mentioned user ("mention"). Anyone who can
// see the booking may comment.
func (u *lineBookingUseCase) Comment(ctx context.Context, request *models.LineBookingCommentRequest) (*models.LineBookingCommentResponse, error) {
	request.Text = strings.TrimSpace(request.Text)
	if err := validation.ValidateStruct(request,
		validation.Field(&request.ID, validation.Required, validation.UUID),
		validation.Field(&request.Text, validation.Required, validation.RuneLength(1, maxCommentRunes)),
		validation.Field(&request.MentionUserIDs, validation.Length(0, maxMentions), validation.Each(validation.UUID)),
	); err != nil {
		return nil, cerrors.Validation(err)
	}
	info := request.UserRequestInfo

	var item models.LineBookingEventItem
	err := u.coreRegistry.DBTx.WithTx(ctx, func(ctx context.Context, appDB query.AppIDB) error {
		b, err := u.findVisible(ctx, appDB, request.ID, info)
		if err != nil {
			return err
		}
		allowed, err := u.mentionable(ctx, appDB, b)
		if err != nil {
			return err
		}
		mentions := []domain.LineBookingEventMention{}
		for _, id := range request.MentionUserIDs {
			i := slices.IndexFunc(allowed, func(m models.LineBookingUser) bool { return m.ID == id })
			if i < 0 {
				return cerrors.FieldError("mentionUserIds", "can only mention users who can see this booking")
			}
			if !slices.ContainsFunc(mentions, func(m domain.LineBookingEventMention) bool { return m.ID == id }) {
				mentions = append(mentions, domain.LineBookingEventMention{ID: id, Name: allowed[i].Name})
			}
		}

		now := time.Now()
		event, err := u.repositoryRegistry.LineBookingEventRepository.WithTx(appDB).Create(ctx, &domain.LineBookingEvent{
			ID:            utils.NewID(),
			LineBookingID: b.ID,
			Action:        linebookingevent.Comment,
			Detail:        domain.LineBookingEventDetail{Note: request.Text, Mentions: mentions},
			CreateDate:    now,
			CreateBy:      info.UserID,
		})
		if err != nil {
			return err
		}

		// Recipients: mentioned users first (a mention wins over a plain comment), then the owner and
		// the finance user; never the author.
		kinds := map[string]string{}
		order := []string{}
		add := func(userID, kind string) {
			if userID == "" || userID == info.UserID {
				return
			}
			if _, ok := kinds[userID]; !ok {
				kinds[userID] = kind
				order = append(order, userID)
			}
		}
		for _, m := range mentions {
			add(m.ID, linebookingevent.NotifyMention)
		}
		add(b.SalesUserID.String, linebookingevent.NotifyComment)
		add(b.FinanceUserID.String, linebookingevent.NotifyComment)
		notifications := u.repositoryRegistry.LineBookingNotificationRepository.WithTx(appDB)
		for _, userID := range order {
			if _, err := notifications.Create(ctx, &domain.LineBookingNotification{
				ID:            utils.NewID(),
				UserID:        userID,
				LineBookingID: b.ID,
				EventID:       event.ID,
				Kind:          kinds[userID],
				CreateDate:    now,
			}); err != nil {
				return err
			}
		}

		item = models.LineBookingEventItem{
			ID:         event.ID,
			Action:     event.Action,
			Changes:    []domain.LineBookingEventChange{},
			Note:       request.Text,
			Mentions:   mentionItems(mentions),
			CreateDate: event.CreateDate,
			User:       &models.LineBookingUser{ID: info.UserID, Name: info.Name},
		}
		return nil
	})
	if err != nil {
		return nil, cerrors.Wrap(err, "comment on line booking")
	}
	return &models.LineBookingCommentResponse{LineBookingEventItem: item}, nil
}

func mentionItems(mentions []domain.LineBookingEventMention) []models.LineBookingUser {
	items := make([]models.LineBookingUser, len(mentions))
	for i, m := range mentions {
		items[i] = models.LineBookingUser{ID: m.ID, Name: m.Name}
	}
	return items
}

// Notifications lists the caller's comment notifications, newest first.
func (u *lineBookingUseCase) Notifications(ctx context.Context, request *models.LineBookingNotificationListRequest) (*models.LineBookingNotificationListResponse, error) {
	if err := validation.ValidateStruct(request,
		validation.Field(&request.Page, validation.Min(0)),
		validation.Field(&request.PageSize, validation.Min(0), validation.Max(maxNotificationPg)),
	); err != nil {
		return nil, cerrors.Validation(err)
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = defaultPageSize
	}
	rows, total, err := u.repositoryRegistry.LineBookingNotificationRepository.List(ctx, request.UserRequestInfo.UserID, request.Unread, request.Page, request.PageSize)
	if err != nil {
		return nil, cerrors.Wrap(err, "list line booking notifications")
	}
	items := make([]models.LineBookingNotificationItem, len(rows))
	for i, n := range rows {
		item := models.LineBookingNotificationItem{ID: n.ID, Kind: n.Kind, LineBookingID: n.LineBookingID, CreateDate: n.CreateDate}
		if n.ReadAt.Valid {
			at := n.ReadAt.Time
			item.ReadAt = &at
		}
		if n.LineBooking != nil {
			item.BookingNo = n.LineBooking.BookingNo
		}
		if e := n.Event; e != nil {
			item.Text = e.Detail.Note
			if user := e.CreateUser; user != nil {
				item.Author = &models.LineBookingUser{ID: user.ID, Name: user.Name}
			}
		}
		items[i] = item
	}
	return &models.LineBookingNotificationListResponse{Items: items, Total: total, Page: request.Page, PageSize: request.PageSize}, nil
}

// NotificationsRead marks the caller's notifications on one booking read (they opened it).
func (u *lineBookingUseCase) NotificationsRead(ctx context.Context, request *models.LineBookingNotificationReadRequest) error {
	if err := validation.Validate(request.LineBookingID, validation.Required, validation.UUID); err != nil {
		return cerrors.Validation(validation.Errors{"lineBookingId": err})
	}
	if err := u.repositoryRegistry.LineBookingNotificationRepository.MarkRead(ctx, request.UserRequestInfo.UserID, request.LineBookingID, time.Now()); err != nil {
		return cerrors.Wrap(err, "mark line booking notifications read")
	}
	return nil
}
